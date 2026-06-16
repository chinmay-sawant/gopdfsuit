package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/svg"
)

// fmtNumImg formats a float with 2 decimal places for image dimensions
func fmtNumImg(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

// rgbDataPool recycles byte slices for RGB conversion
var rgbDataPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 1024*1024) // Start with 1MB
		return &buf
	},
}

const maxImageCacheEntries = 256

// imageCache stores decoded images keyed by a hash of their base64 data.
// A single-slot most-recently-used cache short-circuits the FNV-1a hash
// whenever the same base64 string is decoded back-to-back, which is the
// common case for image rows in a table.
type imageCache struct {
	mu    sync.RWMutex
	cache map[uint64]*ImageObject // FNV-1a hash -> decoded image

	// Single-slot MRU: if the incoming base64 string shares the same
	// (data pointer, length) as the last one, we know the hash matches
	// and can return the cached object without hashing at all.
	lastDataPtr *byte
	lastDataLen int
	lastHash    uint64
	lastObj     *ImageObject
}

var imgCache = &imageCache{
	cache: make(map[uint64]*ImageObject),
}

// clear drops all cached image entries and the MRU slot.
func (c *imageCache) clear() {
	c.cache = make(map[uint64]*ImageObject)
	c.lastDataPtr = nil
	c.lastDataLen = 0
	c.lastHash = 0
	c.lastObj = nil
}

// fnv1aHash computes FNV-1a hash for quick image deduplication
func fnv1aHash(data string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	hash := uint64(offset64)
	for i := 0; i < len(data); i++ {
		hash ^= uint64(data[i])
		hash *= prime64
	}
	return hash
}

// ResetImageCache clears the image cache (call between PDF generations if needed)
func ResetImageCache() {
	imgCache.mu.Lock()
	imgCache.clear()
	imgCache.mu.Unlock()
}

// copyCachedImageObject returns a fresh ImageObject that shares the heavy
// fields (decoded pixel data) with the cached entry.
func copyCachedImageObject(cached *ImageObject) *ImageObject {
	return &ImageObject{
		Width:        cached.Width,
		Height:       cached.Height,
		ColorSpace:   cached.ColorSpace,
		BitsPerComp:  cached.BitsPerComp,
		Filter:       cached.Filter,
		ImageData:    cached.ImageData,
		ImageDataLen: cached.ImageDataLen,
		CacheKey:     cached.CacheKey,
		SourceLen:    cached.SourceLen,
	}
}

// getRGBDataBuffer returns a buffer with at least the requested length
func getRGBDataBuffer(length int) []byte {
	bufPtr := rgbDataPool.Get().(*[]byte)
	buf := *bufPtr
	if cap(buf) < length {
		// If capacity is insufficient, allocate a new one (old one is discarded from pool)
		return make([]byte, length)
	}
	return buf[:length]
}

// putRGBDataBuffer returns a buffer to the pool
func putRGBDataBuffer(buf []byte) {
	rgbDataPool.Put(&buf)
}

// ImageObject represents a PDF image XObject
type ImageObject struct {
	ImageData    []byte
	ColorSpace   string
	Filter       string
	ObjectID     int
	Width        int
	Height       int
	BitsPerComp  int
	ImageDataLen int
	IsForm       bool
	CacheKey     uint64
	SourceLen    int
}

// DecodeImageData decodes base64 image data and returns image information
// Uses caching to avoid re-decoding duplicate images
func DecodeImageData(base64Data string) (*ImageObject, error) {
	// Remove any data URL prefix if present
	cleanData := base64Data
	if strings.Contains(cleanData, ",") {
		parts := strings.Split(cleanData, ",")
		if len(parts) > 1 {
			cleanData = parts[1]
		}
	}

	// Fast path: same image as the most recent decode. Two pointer-sized
	// comparisons are enough because Go strings are immutable - matching
	// (data pointer, length) implies matching contents. This avoids the
	// full FNV-1a hash on the hot path of repeated image rows.
	dataPtr := unsafe.StringData(cleanData)
	dataLen := len(cleanData)
	imgCache.mu.RLock()
	if imgCache.lastObj != nil && imgCache.lastDataPtr == dataPtr && imgCache.lastDataLen == dataLen {
		cached := imgCache.lastObj
		imgCache.mu.RUnlock()
		return copyCachedImageObject(cached), nil
	}
	imgCache.mu.RUnlock()

	// Check cache first (fast path for duplicate images)
	hash := fnv1aHash(cleanData)
	imgCache.mu.RLock()
	if cached, ok := imgCache.cache[hash]; ok {
		// Promote the matched entry to the MRU slot so subsequent calls
		// can skip the hash entirely.
		imgCache.lastDataPtr = dataPtr
		imgCache.lastDataLen = dataLen
		imgCache.lastHash = hash
		imgCache.lastObj = cached
		imgCache.mu.RUnlock()
		return copyCachedImageObject(cached), nil
	}
	imgCache.mu.RUnlock()

	// Decode base64 to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(cleanData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	// Check if data is SVG (simple check)
	if bytes.Contains(imageBytes, []byte("<svg")) || bytes.Contains(imageBytes, []byte("<SVG")) {
		pdfCmds, w, h, err := svg.ConvertSVGToPDFCommands(imageBytes)
		if err == nil {
			// Successfully converted SVG to PDF commands
			// We return a Form XObject structure
			return &ImageObject{
				Width:        w,
				Height:       h,
				ImageData:    pdfCmds,
				ImageDataLen: len(pdfCmds),
				IsForm:       true,
				CacheKey:     hash,
				SourceLen:    len(cleanData),
			}, nil
		}
		// If SVG conversion fails, try to proceed as raster image (might fail too) or return error?
		// Let's try to proceed, maybe it's not really SVG
	}

	// Try to decode as PNG/JPEG
	img, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	imgObj := &ImageObject{
		Width:       width,
		Height:      height,
		ColorSpace:  "/DeviceRGB",
		BitsPerComp: 8,
		CacheKey:    hash,
		SourceLen:   len(cleanData),
	}

	// Convert image to raw RGB data for PDF
	switch format {
	case "png":
		// For PNG, convert to RGB with proper alpha handling
		// Check if image has transparency
		hasAlpha := false
		switch img.(type) {
		case *image.NRGBA, *image.RGBA, *image.RGBA64, *image.NRGBA64:
			hasAlpha = true
		}

		rgbSize := width * height * 3
		rawRGB := getRGBDataBuffer(rgbSize)
		defer putRGBDataBuffer(rawRGB)

		if hasAlpha {
			// For images with transparency, convert to RGBA with white background
			err = convertToRGBWithAlpha(img, rawRGB)
		} else {
			// For opaque images, convert to RGB
			err = convertToRGB(img, rawRGB)
		}
		if err != nil {
			return nil, err
		}

		// Compress with pooled zlib writer (avoids 256KB allocation per compression)
		compressedBuf := getCompressBuffer()
		zlibWriter := getZlibWriter(compressedBuf)
		if _, err := zlibWriter.Write(rawRGB); err != nil {
			_ = zlibWriter.Close()
			putZlibWriter(zlibWriter)
			return nil, err
		}
		_ = zlibWriter.Close()
		putZlibWriter(zlibWriter)

		imgObj.Filter = "/FlateDecode"
		// Copy compressed data since buffer will be reused
		imgObj.ImageData = append([]byte(nil), compressedBuf.Bytes()...)
		imgObj.ImageDataLen = len(imgObj.ImageData)
		putCompressBuffer(compressedBuf)

	case "jpeg", "jpg":
		// For JPEG, use original bytes directly to preserve quality
		// No re-encoding needed - this prevents quality loss and distortion
		imgObj.Filter = "/DCTDecode"
		imgObj.ImageData = imageBytes // Use original JPEG data
		imgObj.ImageDataLen = len(imageBytes)

	default:
		// For other formats, convert to RGB and compress with zlib
		rgbSize := width * height * 3
		rawRGB := getRGBDataBuffer(rgbSize)
		defer putRGBDataBuffer(rawRGB)

		err = convertToRGB(img, rawRGB)
		if err != nil {
			return nil, err
		}
		// Compress with pooled zlib writer
		compressedBuf := getCompressBuffer()
		zlibWriter := getZlibWriter(compressedBuf)
		if _, err := zlibWriter.Write(rawRGB); err != nil {
			_ = zlibWriter.Close()
			putZlibWriter(zlibWriter)
			return nil, err
		}
		_ = zlibWriter.Close()
		putZlibWriter(zlibWriter)

		imgObj.Filter = "/FlateDecode"
		imgObj.ImageData = append([]byte(nil), compressedBuf.Bytes()...)
		imgObj.ImageDataLen = len(imgObj.ImageData)
		putCompressBuffer(compressedBuf)
	}

	// Store in cache for future lookups of same image data, with bounded size.
	imgCache.mu.Lock()
	if len(imgCache.cache) >= maxImageCacheEntries {
		imgCache.clear()
	}
	imgCache.cache[hash] = imgObj
	imgCache.lastDataPtr = dataPtr
	imgCache.lastDataLen = dataLen
	imgCache.lastHash = hash
	imgCache.lastObj = imgObj
	imgCache.mu.Unlock()

	return imgObj, nil
}

// convertToRGB converts an image to raw RGB bytes
func convertToRGB(img image.Image, rgbData []byte) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	expectedLen := width * height * 3

	if len(rgbData) < expectedLen {
		return fmt.Errorf("rgbData buffer too small: got %d, want %d", len(rgbData), expectedLen)
	}

	idx := 0

	// Fast paths for common types (avoids repeated type assertions)
	switch v := img.(type) {
	case *image.NRGBA:
		for y := range height {
			rowStart := (y + bounds.Min.Y - v.Rect.Min.Y) * v.Stride
			for x := range width {
				pixOffset := rowStart + (x+bounds.Min.X-v.Rect.Min.X)*4
				rgbData[idx] = v.Pix[pixOffset]
				rgbData[idx+1] = v.Pix[pixOffset+1]
				rgbData[idx+2] = v.Pix[pixOffset+2]
				idx += 3
			}
		}
		return nil
	case *image.RGBA:
		for y := range height {
			rowStart := (y + bounds.Min.Y - v.Rect.Min.Y) * v.Stride
			for x := range width {
				pixOffset := rowStart + (x+bounds.Min.X-v.Rect.Min.X)*4
				rgbData[idx] = v.Pix[pixOffset]
				rgbData[idx+1] = v.Pix[pixOffset+1]
				rgbData[idx+2] = v.Pix[pixOffset+2]
				idx += 3
			}
		}
		return nil
	}

	// Read image top-to-bottom (normal order) - Slow path
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit
			rgbData[idx] = byte(r >> 8)
			rgbData[idx+1] = byte(g >> 8)
			rgbData[idx+2] = byte(b >> 8)
			idx += 3
		}
	}

	return nil
}

// convertToRGBWithAlpha converts an image with alpha channel to RGB
// Blends transparent pixels with white background
// Optimized with fast integer division approximation
func convertToRGBWithAlpha(img image.Image, rgbData []byte) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	expectedLen := width * height * 3

	if len(rgbData) < expectedLen {
		return fmt.Errorf("rgbData buffer too small: got %d, want %d", len(rgbData), expectedLen)
	}

	idx := 0

	// Optimize for NRGBA (common for PNG) and RGBA using type switch
	switch v := img.(type) {
	case *image.NRGBA:
		pix := v.Pix
		stride := v.Stride
		minX := bounds.Min.X - v.Rect.Min.X
		minY := bounds.Min.Y - v.Rect.Min.Y
		for y := range height {
			rowStart := (y + minY) * stride
			for x := range width {
				pixOffset := rowStart + (x+minX)*4
				r := uint32(pix[pixOffset])
				g := uint32(pix[pixOffset+1])
				b := uint32(pix[pixOffset+2])
				a := uint32(pix[pixOffset+3])

				switch a {
				case 255:
					rgbData[idx] = byte(r)
					rgbData[idx+1] = byte(g)
					rgbData[idx+2] = byte(b)
				case 0:
					rgbData[idx] = 255
					rgbData[idx+1] = 255
					rgbData[idx+2] = 255
				default:
					invA := 255 - a
					white := 255 * invA
					rgbData[idx] = byte(((r*a + white) * 0x8081) >> 23)
					rgbData[idx+1] = byte(((g*a + white) * 0x8081) >> 23)
					rgbData[idx+2] = byte(((b*a + white) * 0x8081) >> 23)
				}
				idx += 3
			}
		}
		return nil
	case *image.RGBA:
		pix := v.Pix
		stride := v.Stride
		minX := bounds.Min.X - v.Rect.Min.X
		minY := bounds.Min.Y - v.Rect.Min.Y
		for y := range height {
			rowStart := (y + minY) * stride
			for x := range width {
				pixOffset := rowStart + (x+minX)*4
				rPre := uint32(pix[pixOffset])
				gPre := uint32(pix[pixOffset+1])
				bPre := uint32(pix[pixOffset+2])
				a := uint32(pix[pixOffset+3])

				switch a {
				case 255:
					rgbData[idx] = byte(rPre)
					rgbData[idx+1] = byte(gPre)
					rgbData[idx+2] = byte(bPre)
				case 0:
					rgbData[idx] = 255
					rgbData[idx+1] = 255
					rgbData[idx+2] = 255
				default:
					bgPart := ((255 * (255 - a)) * 0x8081) >> 23
					rgbData[idx] = byte(rPre + bgPart)
					rgbData[idx+1] = byte(gPre + bgPart)
					rgbData[idx+2] = byte(bPre + bgPart)
				}
				idx += 3
			}
		}
		return nil
	}

	// Slow path for other image types
	// Read image top-to-bottom (normal order)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Convert from 16-bit to 8-bit
			r8 := uint32(r >> 8)
			g8 := uint32(g >> 8)
			b8 := uint32(b >> 8)
			a8 := uint32(a >> 8)

			switch a8 {
			case 255:
				rgbData[idx] = byte(r8)
				rgbData[idx+1] = byte(g8)
				rgbData[idx+2] = byte(b8)
			case 0:
				rgbData[idx] = 255
				rgbData[idx+1] = 255
				rgbData[idx+2] = 255
			default:
				// Blend with white background using fast integer math
				invA := 255 - a8
				white := 255 * invA
				rgbData[idx] = byte(((r8*a8 + white) * 0x8081) >> 23)
				rgbData[idx+1] = byte(((g8*a8 + white) * 0x8081) >> 23)
				rgbData[idx+2] = byte(((b8*a8 + white) * 0x8081) >> 23)
			}
			idx += 3
		}
	}

	return nil
}

// CreateImageXObject creates a PDF XObject for an image.
func CreateImageXObject(imgObj *ImageObject, objectID int) []byte {
	// Handle Form XObject (Vectors/SVG)
	if imgObj.IsForm {
		b := make([]byte, 0, 256+imgObj.ImageDataLen+24)
		b = strconv.AppendInt(b, int64(objectID), 10)
		b = append(b, " 0 obj\n<< /Type /XObject\n   /Subtype /Form\n   /BBox [0 0 1 1]\n"...)
		b = append(b, "   /Resources << /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>\n"...)
		b = append(b, "   /Length "...)
		b = strconv.AppendInt(b, int64(imgObj.ImageDataLen), 10)
		b = append(b, "\n>>\nstream\n"...)
		b = append(b, imgObj.ImageData...)
		b = append(b, "\nendstream\nendobj\n"...)
		return b
	}

	// Pre-allocate buffer with capacity for the header and image payload.
	b := make([]byte, 0, 256+imgObj.ImageDataLen+24)

	b = strconv.AppendInt(b, int64(objectID), 10)
	b = append(b, " 0 obj\n<< /Type /XObject\n   /Subtype /Image\n   /Width "...)
	b = strconv.AppendInt(b, int64(imgObj.Width), 10)
	b = append(b, "\n   /Height "...)
	b = strconv.AppendInt(b, int64(imgObj.Height), 10)
	b = append(b, "\n   /ColorSpace "...)
	b = append(b, imgObj.ColorSpace...)
	b = append(b, "\n   /BitsPerComponent "...)
	b = strconv.AppendInt(b, int64(imgObj.BitsPerComp), 10)
	b = append(b, "\n"...)

	if imgObj.Filter != "" {
		b = append(b, "   /Filter "...)
		b = append(b, imgObj.Filter...)
		b = append(b, "\n"...)
	}

	b = append(b, "   /Length "...)
	b = strconv.AppendInt(b, int64(imgObj.ImageDataLen), 10)
	b = append(b, "\n>>\nstream\n"...)

	b = append(b, imgObj.ImageData...)
	b = append(b, "\nendstream\nendobj\n"...)

	return b
}

// ImageEncryptor interface for encrypting image data
type ImageEncryptor interface {
	EncryptStream(data []byte, objNum, genNum int) []byte
}

// CreateEncryptedImageXObject creates an encrypted PDF XObject for an image.
func CreateEncryptedImageXObject(imgObj *ImageObject, objectID int, encryptor ImageEncryptor) []byte {
	// Encrypt the image data (or form stream commands)
	encryptedData := encryptor.EncryptStream(imgObj.ImageData, objectID, 0)

	// Handle Form XObject (Vectors/SVG)
	if imgObj.IsForm {
		b := make([]byte, 0, 256+len(encryptedData)+24)
		b = strconv.AppendInt(b, int64(objectID), 10)
		b = append(b, " 0 obj\n<< /Type /XObject\n   /Subtype /Form\n   /BBox [0 0 1 1]\n"...)
		b = append(b, "   /Resources << /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>\n"...)
		b = append(b, "   /Length "...)
		b = strconv.AppendInt(b, int64(len(encryptedData)), 10)
		b = append(b, "\n>>\nstream\n"...)
		b = append(b, encryptedData...)
		b = append(b, "\nendstream\nendobj\n"...)
		return b
	}

	// Pre-allocate buffer with capacity for the header and encrypted payload.
	b := make([]byte, 0, 256+len(encryptedData)+24)
	b = strconv.AppendInt(b, int64(objectID), 10)
	b = append(b, " 0 obj\n<< /Type /XObject\n   /Subtype /Image\n   /Width "...)
	b = strconv.AppendInt(b, int64(imgObj.Width), 10)
	b = append(b, "\n   /Height "...)
	b = strconv.AppendInt(b, int64(imgObj.Height), 10)
	b = append(b, "\n   /ColorSpace "...)
	b = append(b, imgObj.ColorSpace...)
	b = append(b, "\n   /BitsPerComponent "...)
	b = strconv.AppendInt(b, int64(imgObj.BitsPerComp), 10)
	b = append(b, "\n"...)

	if imgObj.Filter != "" {
		b = append(b, "   /Filter "...)
		b = append(b, imgObj.Filter...)
		b = append(b, "\n"...)
	}

	b = append(b, "   /Length "...)
	b = strconv.AppendInt(b, int64(len(encryptedData)), 10)
	b = append(b, "\n>>\nstream\n"...)

	b = append(b, encryptedData...)
	b = append(b, "\nendstream\nendobj\n"...)

	return b
}

// drawImageWithXObject renders an image using XObject reference
// For standalone images, it fits the image to the full usable width (between margins)
func drawImageWithXObject(contentStream *bytes.Buffer, image models.Image, imageXObjectRef string, pageManager *PageManager, originalImgWidth, originalImgHeight int) {
	// Calculate the usable width (page width minus margins on both sides)
	usableWidth := pageManager.ContentWidth()

	// Use the full usable width for the image
	imageWidth := usableWidth

	// Calculate height to maintain aspect ratio
	var imageHeight float64
	switch {
	case originalImgWidth > 0 && originalImgHeight > 0:
		// Maintain aspect ratio based on original image dimensions
		aspectRatio := float64(originalImgHeight) / float64(originalImgWidth)
		imageHeight = imageWidth * aspectRatio
	case image.Height > 0 && image.Width > 0:
		// Use provided dimensions to calculate aspect ratio
		aspectRatio := image.Height / image.Width
		imageHeight = imageWidth * aspectRatio
	default:
		// Default height if no dimensions available
		imageHeight = 200
	}

	// Position image at the left margin
	imageX := pageManager.Margins.Left
	imageY := pageManager.CurrentYPos - imageHeight

	// Save graphics state
	contentStream.WriteString("q\n")

	// Set up transformation matrix to position and scale the image
	// PDF images are drawn in a 1x1 unit square by default
	// We need to scale and translate to our desired size and position
	var imgBuf []byte
	imgBuf = append(imgBuf, fmtNumImg(imageWidth)...)
	imgBuf = append(imgBuf, " 0 0 "...)
	imgBuf = append(imgBuf, fmtNumImg(imageHeight)...)
	imgBuf = append(imgBuf, ' ')
	imgBuf = append(imgBuf, fmtNumImg(imageX)...)
	imgBuf = append(imgBuf, ' ')
	imgBuf = append(imgBuf, fmtNumImg(imageY)...)
	imgBuf = append(imgBuf, " cm\n"...)
	contentStream.Write(imgBuf)

	// Draw the image using the XObject reference
	imgBuf = imgBuf[:0]
	imgBuf = append(imgBuf, imageXObjectRef...)
	imgBuf = append(imgBuf, " Do\n"...)
	contentStream.Write(imgBuf)

	// Restore graphics state
	contentStream.WriteString("Q\n")

	pageManager.CurrentYPos -= imageHeight // No extra spacing
}
