package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
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
}

// DecodeImageData decodes base64 image data and returns image information
func DecodeImageData(base64Data string) (*ImageObject, error) {
	// Remove any data URL prefix if present
	if strings.Contains(base64Data, ",") {
		parts := strings.Split(base64Data, ",")
		if len(parts) > 1 {
			base64Data = parts[1]
		}
	}

	// Decode base64 to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	// Try to decode as PNG first
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

		// Compress with zlib (PDF FlateDecode expects zlib format)
		var compressedBuf bytes.Buffer
		zlibWriter, err := zlib.NewWriterLevel(&compressedBuf, zlib.BestSpeed)
		if err != nil {
			return nil, err
		}
		if _, err := zlibWriter.Write(rawRGB); err != nil {
			_ = zlibWriter.Close()
			return nil, err
		}
		_ = zlibWriter.Close()
		imgObj.Filter = "/FlateDecode"
		imgObj.ImageData = compressedBuf.Bytes()
		imgObj.ImageDataLen = len(imgObj.ImageData)

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
		// Compress with zlib (PDF FlateDecode expects zlib format)
		var compressedBuf bytes.Buffer
		zlibWriter, err := zlib.NewWriterLevel(&compressedBuf, zlib.BestSpeed)
		if err != nil {
			return nil, err
		}
		if _, err := zlibWriter.Write(rawRGB); err != nil {
			_ = zlibWriter.Close()
			return nil, err
		}
		_ = zlibWriter.Close()
		imgObj.Filter = "/FlateDecode"
		imgObj.ImageData = compressedBuf.Bytes()
		imgObj.ImageDataLen = len(imgObj.ImageData)
	}

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
		for y := range height {
			rowStart := (y + bounds.Min.Y - nrgba.Rect.Min.Y) * nrgba.Stride
			for x := range width {
				pixOffset := rowStart + (x+bounds.Min.X-nrgba.Rect.Min.X)*4
				r := int(nrgba.Pix[pixOffset])
				g := int(nrgba.Pix[pixOffset+1])
				b := int(nrgba.Pix[pixOffset+2])
				a := int(nrgba.Pix[pixOffset+3])

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
					// Blend with white background: Result = C*alpha + 255*(1-alpha)
					// using integer math: (C*a + 255*(255-a))/255
					invA := 255 - a
					rgbData[idx] = byte((r*a + 255*invA) / 255)
					rgbData[idx+1] = byte((g*a + 255*invA) / 255)
					rgbData[idx+2] = byte((b*a + 255*invA) / 255)
				}
				idx += 3
			}
		}
		return nil
	}

	// Optimize for RGBA which has pre-multiplied alpha
	if rgba, ok := img.(*image.RGBA); ok {
		for y := range height {
			rowStart := (y + bounds.Min.Y - rgba.Rect.Min.Y) * rgba.Stride
			for x := range width {
				pixOffset := rowStart + (x+bounds.Min.X-rgba.Rect.Min.X)*4
				// Values are already premultiplied by alpha: C_pre = C_straight * alpha
				rPre := int(rgba.Pix[pixOffset])
				gPre := int(rgba.Pix[pixOffset+1])
				bPre := int(rgba.Pix[pixOffset+2])
				a := int(rgba.Pix[pixOffset+3])

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
					// Blend with white: Result = C_pre + 255*(1-alpha)
					// (Assuming 0-255 range for all)
					bgPart := (255 * (255 - a)) / 255
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
			r8 := byte(r >> 8)
			g8 := byte(g >> 8)
			b8 := byte(b >> 8)
			a8 := float64(a) / 65535.0

			// Blend with white background (255, 255, 255)
			// Formula: result = foreground * alpha + background * (1 - alpha)
			rgbData[idx] = byte(float64(r8)*a8 + 255*(1-a8))
			rgbData[idx+1] = byte(float64(g8)*a8 + 255*(1-a8))
			rgbData[idx+2] = byte(float64(b8)*a8 + 255*(1-a8))
			idx += 3
		}
	}

	return nil
}

// CreateImageXObject creates a PDF XObject for an image
func CreateImageXObject(imgObj *ImageObject, objectID int) string {
	var buf bytes.Buffer
	var b []byte

	b = strconv.AppendInt(b, int64(objectID), 10)
	b = append(b, " 0 obj\n"...)
	buf.Write(b)
	buf.WriteString("<< /Type /XObject\n")
	buf.WriteString("   /Subtype /Image\n")

	var intBuf []byte
	intBuf = append(intBuf, "   /Width "...)
	intBuf = strconv.AppendInt(intBuf, int64(imgObj.Width), 10)
	intBuf = append(intBuf, "\n"...)
	buf.Write(intBuf)

	intBuf = intBuf[:0]
	intBuf = append(intBuf, "   /Height "...)
	intBuf = strconv.AppendInt(intBuf, int64(imgObj.Height), 10)
	intBuf = append(intBuf, "\n"...)
	buf.Write(intBuf)

	buf.WriteString("   /ColorSpace ")
	buf.WriteString(imgObj.ColorSpace)
	buf.WriteString("\n")

	intBuf = intBuf[:0]
	intBuf = append(intBuf, "   /BitsPerComponent "...)
	intBuf = strconv.AppendInt(intBuf, int64(imgObj.BitsPerComp), 10)
	intBuf = append(intBuf, "\n"...)
	buf.Write(intBuf)

	if imgObj.Filter != "" {
		buf.WriteString("   /Filter ")
		buf.WriteString(imgObj.Filter)
		buf.WriteString("\n")
	}

	intBuf = intBuf[:0]
	intBuf = append(intBuf, "   /Length "...)
	intBuf = strconv.AppendInt(intBuf, int64(imgObj.ImageDataLen), 10)
	intBuf = append(intBuf, "\n"...)
	buf.Write(intBuf)
	buf.WriteString(">>\n")
	buf.WriteString("stream\n")
	buf.Write(imgObj.ImageData)
	buf.WriteString("\nendstream\n")
	buf.WriteString("endobj\n")

	return buf.String()
}

// ImageEncryptor interface for encrypting image data
type ImageEncryptor interface {
	EncryptStream(data []byte, objNum, genNum int) []byte
}

// CreateEncryptedImageXObject creates an encrypted PDF XObject for an image
func CreateEncryptedImageXObject(imgObj *ImageObject, objectID int, encryptor ImageEncryptor) string {
	var buf bytes.Buffer

	// Encrypt the image data
	encryptedData := encryptor.EncryptStream(imgObj.ImageData, objectID, 0)

	var b []byte
	b = strconv.AppendInt(b, int64(objectID), 10)
	b = append(b, " 0 obj\n"...)
	buf.Write(b)
	buf.WriteString("<< /Type /XObject\n")
	buf.WriteString("   /Subtype /Image\n")
	fmt.Fprintf(&buf, "   /Width %d\n", imgObj.Width)
	fmt.Fprintf(&buf, "   /Height %d\n", imgObj.Height)
	fmt.Fprintf(&buf, "   /ColorSpace %s\n", imgObj.ColorSpace)
	fmt.Fprintf(&buf, "   /BitsPerComponent %d\n", imgObj.BitsPerComp)

	if imgObj.Filter != "" {
		fmt.Fprintf(&buf, "   /Filter %s\n", imgObj.Filter)
	}

	fmt.Fprintf(&buf, "   /Length %d\n", len(encryptedData))
	buf.WriteString(">>\n")
	buf.WriteString("stream\n")
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
