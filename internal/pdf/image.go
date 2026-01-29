package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"strconv"
	"strings"

	"sync"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// fmtNumImg formats a float with 2 decimal places for image dimensions
func fmtNumImg(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

// rgbDataPool recycles byte slices for RGB conversion
var rgbDataPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 1024*1024) // Start with 1MB
		return &buf
	},
}

// zlibWriterPool recycles zlib writers to avoid allocation overhead
// Each zlib.NewWriter allocates ~256KB for compression tables
var zlibWriterPool = sync.Pool{
	New: func() interface{} {
		// Create writer that will be reset with actual buffer later
		w, _ := zlib.NewWriterLevel(io.Discard, zlib.BestSpeed)
		return w
	},
}

// compressBufPool recycles bytes.Buffer for compression output
var compressBufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getZlibWriter returns a pooled zlib writer reset to write to the given buffer
func getZlibWriter(buf *bytes.Buffer) *zlib.Writer {
	w := zlibWriterPool.Get().(*zlib.Writer)
	w.Reset(buf)
	return w
}

// putZlibWriter returns a zlib writer to the pool
func putZlibWriter(w *zlib.Writer) {
	zlibWriterPool.Put(w)
}

// getCompressBuffer returns a pooled compression buffer
func getCompressBuffer() *bytes.Buffer {
	buf := compressBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// imageCache stores decoded images keyed by a hash of their base64 data
// This avoids re-decoding the same image when it appears multiple times
type imageCache struct {
	mu    sync.RWMutex
	cache map[uint64]*ImageObject // FNV-1a hash -> decoded image
}

var imgCache = &imageCache{
	cache: make(map[uint64]*ImageObject),
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
	imgCache.cache = make(map[uint64]*ImageObject)
	imgCache.mu.Unlock()
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
	ObjectID     int
	Width        int
	Height       int
	ColorSpace   string
	BitsPerComp  int
	Filter       string
	ImageData    []byte
	ImageDataLen int
	IsForm       bool
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

	// Check cache first (fast path for duplicate images)
	hash := fnv1aHash(cleanData)
	imgCache.mu.RLock()
	if cached, ok := imgCache.cache[hash]; ok {
		imgCache.mu.RUnlock()
		// Return a copy with a new ObjectID (will be set by caller)
		return &ImageObject{
			Width:        cached.Width,
			Height:       cached.Height,
			ColorSpace:   cached.ColorSpace,
			BitsPerComp:  cached.BitsPerComp,
			Filter:       cached.Filter,
			ImageData:    cached.ImageData,
			ImageDataLen: cached.ImageDataLen,
		}, nil
	}
	imgCache.mu.RUnlock()

	// Decode base64 to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(cleanData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	// Check if data is SVG (simple check)
	if bytes.Contains(imageBytes, []byte("<svg")) || bytes.Contains(imageBytes, []byte("<SVG")) {
		pdfCmds, w, h, err := ConvertSVGToPDFCommands(imageBytes)
		if err == nil {
			// Successfully converted SVG to PDF commands
			// We return a Form XObject structure
			return &ImageObject{
				Width:        w,
				Height:       h,
				ImageData:    pdfCmds,
				ImageDataLen: len(pdfCmds),
				IsForm:       true,
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
		compressBufPool.Put(compressedBuf)

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
		compressBufPool.Put(compressedBuf)
	}

	// Store in cache for future lookups of same image data
	imgCache.mu.Lock()
	imgCache.cache[hash] = imgObj
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

	// Fast path for NRGBA (common for PNGs)
	if nrgba, ok := img.(*image.NRGBA); ok {
		for y := 0; y < height; y++ {
			// Calculate starting offset for this row in the source image
			rowStart := (y + bounds.Min.Y - nrgba.Rect.Min.Y) * nrgba.Stride
			for x := 0; x < width; x++ {
				pixOffset := rowStart + (x+bounds.Min.X-nrgba.Rect.Min.X)*4
				// Just take R, G, B, ignore Alpha
				rgbData[idx] = nrgba.Pix[pixOffset]
				rgbData[idx+1] = nrgba.Pix[pixOffset+1]
				rgbData[idx+2] = nrgba.Pix[pixOffset+2]
				idx += 3
			}
		}
		return nil
	}

	// Fast path for RGBA
	if rgba, ok := img.(*image.RGBA); ok {
		for y := 0; y < height; y++ {
			rowStart := (y + bounds.Min.Y - rgba.Rect.Min.Y) * rgba.Stride
			for x := 0; x < width; x++ {
				pixOffset := rowStart + (x+bounds.Min.X-rgba.Rect.Min.X)*4
				// RGBA uses premultiplied alpha
				rgbData[idx] = rgba.Pix[pixOffset]
				rgbData[idx+1] = rgba.Pix[pixOffset+1]
				rgbData[idx+2] = rgba.Pix[pixOffset+2]
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

	// Optimize for NRGBA (common for PNG) which has straight alpha
	if nrgba, ok := img.(*image.NRGBA); ok {
		pix := nrgba.Pix
		stride := nrgba.Stride
		minX := bounds.Min.X - nrgba.Rect.Min.X
		minY := bounds.Min.Y - nrgba.Rect.Min.Y
		for y := 0; y < height; y++ {
			rowStart := (y + minY) * stride
			for x := 0; x < width; x++ {
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
					// Blend with white: Result = C*alpha + 255*(255-alpha)
					// Fast divide by 255: (x * 257 + 256) >> 16 ≈ x / 255
					// For better accuracy: (x + 127) / 255 ≈ (x * 0x8081) >> 23
					invA := 255 - a
					white := 255 * invA
					// Using: ((n * 0x8081) >> 23) approximates n/255 with high accuracy
					rgbData[idx] = byte(((r*a + white) * 0x8081) >> 23)
					rgbData[idx+1] = byte(((g*a + white) * 0x8081) >> 23)
					rgbData[idx+2] = byte(((b*a + white) * 0x8081) >> 23)
				}
				idx += 3
			}
		}
		return nil
	}

	// Optimize for RGBA which has pre-multiplied alpha
	if rgba, ok := img.(*image.RGBA); ok {
		pix := rgba.Pix
		stride := rgba.Stride
		minX := bounds.Min.X - rgba.Rect.Min.X
		minY := bounds.Min.Y - rgba.Rect.Min.Y
		for y := range height {
			rowStart := (y + minY) * stride
			for x := range width {
				pixOffset := rowStart + (x+minX)*4
				// Values are already premultiplied by alpha: C_pre = C_straight * alpha
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
					// Blend with white: Result = C_pre + 255*(1-alpha/255)
					// Fast divide by 255
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

// CreateImageXObject creates a PDF XObject for an image
func CreateImageXObject(imgObj *ImageObject, objectID int) string {
	var buf bytes.Buffer

	// Handle Form XObject (Vectors/SVG)
	if imgObj.IsForm {
		b := make([]byte, 0, 256)
		b = strconv.AppendInt(b, int64(objectID), 10)
		b = append(b, " 0 obj\n<< /Type /XObject\n   /Subtype /Form\n   /BBox [0 0 1 1]\n"...)
		b = append(b, "   /Resources << /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>\n"...) // Basic resources
		b = append(b, "   /Length "...)
		b = strconv.AppendInt(b, int64(imgObj.ImageDataLen), 10)
		b = append(b, "\n>>\nstream\n"...)

		buf.Write(b)
		buf.Write(imgObj.ImageData)
		buf.WriteString("\nendstream\nendobj\n")
		return buf.String()
	}

	// Pre-allocate buffer with capacity for typical image XObject header
	b := make([]byte, 0, 256)

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

	// Write header and image data in two operations
	buf.Write(b)
	buf.Write(imgObj.ImageData)
	buf.WriteString("\nendstream\nendobj\n")

	return buf.String()
}

// ImageEncryptor interface for encrypting image data
type ImageEncryptor interface {
	EncryptStream(data []byte, objNum, genNum int) []byte
}

// CreateEncryptedImageXObject creates an encrypted PDF XObject for an image
func CreateEncryptedImageXObject(imgObj *ImageObject, objectID int, encryptor ImageEncryptor) string {
	var buf bytes.Buffer

	// Encrypt the image data (or form stream commands)
	encryptedData := encryptor.EncryptStream(imgObj.ImageData, objectID, 0)

	// Handle Form XObject (Vectors/SVG)
	if imgObj.IsForm {
		b := make([]byte, 0, 256)
		b = strconv.AppendInt(b, int64(objectID), 10)
		b = append(b, " 0 obj\n<< /Type /XObject\n   /Subtype /Form\n   /BBox [0 0 1 1]\n"...)
		b = append(b, "   /Resources << /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>\n"...)
		b = append(b, "   /Length "...)
		b = strconv.AppendInt(b, int64(len(encryptedData)), 10)
		b = append(b, "\n>>\nstream\n"...)

		buf.Write(b)
		buf.Write(encryptedData)
		buf.WriteString("\nendstream\nendobj\n")
		return buf.String()
	}

	// Pre-allocate buffer with capacity for typical image XObject header
	b := make([]byte, 0, 256)
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

	// Write header and encrypted data in two operations
	buf.Write(b)
	buf.Write(encryptedData)
	buf.WriteString("\nendstream\n")
	buf.WriteString("endobj\n")

	return buf.String()
}

// drawImageWithXObject renders an image using XObject reference
// For standalone images, it fits the image to the full usable width (between margins)
func drawImageWithXObject(contentStream *bytes.Buffer, image models.Image, imageXObjectRef string, pageManager *PageManager, originalImgWidth, originalImgHeight int) {
	// Calculate the usable width (page width minus margins on both sides)
	usableWidth := pageManager.PageDimensions.Width - 2*margin

	// Use the full usable width for the image
	imageWidth := usableWidth

	// Calculate height to maintain aspect ratio
	var imageHeight float64
	if originalImgWidth > 0 && originalImgHeight > 0 {
		// Maintain aspect ratio based on original image dimensions
		aspectRatio := float64(originalImgHeight) / float64(originalImgWidth)
		imageHeight = imageWidth * aspectRatio
	} else if image.Height > 0 && image.Width > 0 {
		// Use provided dimensions to calculate aspect ratio
		aspectRatio := image.Height / image.Width
		imageHeight = imageWidth * aspectRatio
	} else {
		// Default height if no dimensions available
		imageHeight = 200
	}

	// Position image at the left margin
	imageX := float64(margin)
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
