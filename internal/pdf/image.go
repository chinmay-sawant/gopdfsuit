package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// fmtNumImg formats a float with 2 decimal places for image dimensions
func fmtNumImg(f float64) string {
	return fmt.Sprintf("%.2f", f)
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

		var rawRGB []byte
		if hasAlpha {
			// For images with transparency, convert to RGBA with white background
			rawRGB, err = convertToRGBWithAlpha(img)
			if err != nil {
				return nil, err
			}
		} else {
			// For opaque images, convert to RGB
			rawRGB, err = convertToRGB(img)
			if err != nil {
				return nil, err
			}
		}
		// Compress with zlib (PDF FlateDecode expects zlib format)
		var compressedBuf bytes.Buffer
		zlibWriter := zlib.NewWriter(&compressedBuf)
		zlibWriter.Write(rawRGB)
		zlibWriter.Close()
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
		rawRGB, err := convertToRGB(img)
		if err != nil {
			return nil, err
		}
		// Compress with zlib (PDF FlateDecode expects zlib format)
		var compressedBuf bytes.Buffer
		zlibWriter := zlib.NewWriter(&compressedBuf)
		zlibWriter.Write(rawRGB)
		zlibWriter.Close()
		imgObj.Filter = "/FlateDecode"
		imgObj.ImageData = compressedBuf.Bytes()
		imgObj.ImageDataLen = len(imgObj.ImageData)
	}

	return imgObj, nil
}

// convertToRGB converts an image to raw RGB bytes
func convertToRGB(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create RGB buffer
	rgbData := make([]byte, width*height*3)

	// Read image top-to-bottom (normal order)
	idx := 0
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

	return rgbData, nil
}

// convertToRGBWithAlpha converts an image with alpha channel to RGB
// Blends transparent pixels with white background
func convertToRGBWithAlpha(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create RGB buffer
	rgbData := make([]byte, width*height*3)

	// Read image top-to-bottom (normal order)
	idx := 0
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

	return rgbData, nil
}

// CreateImageXObject creates a PDF XObject for an image
func CreateImageXObject(imgObj *ImageObject, objectID int) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))
	buf.WriteString("<< /Type /XObject\n")
	buf.WriteString("   /Subtype /Image\n")
	buf.WriteString(fmt.Sprintf("   /Width %d\n", imgObj.Width))
	buf.WriteString(fmt.Sprintf("   /Height %d\n", imgObj.Height))
	buf.WriteString(fmt.Sprintf("   /ColorSpace %s\n", imgObj.ColorSpace))
	buf.WriteString(fmt.Sprintf("   /BitsPerComponent %d\n", imgObj.BitsPerComp))

	if imgObj.Filter != "" {
		buf.WriteString(fmt.Sprintf("   /Filter %s\n", imgObj.Filter))
	}

	buf.WriteString(fmt.Sprintf("   /Length %d\n", imgObj.ImageDataLen))
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

	buf.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))
	buf.WriteString("<< /Type /XObject\n")
	buf.WriteString("   /Subtype /Image\n")
	buf.WriteString(fmt.Sprintf("   /Width %d\n", imgObj.Width))
	buf.WriteString(fmt.Sprintf("   /Height %d\n", imgObj.Height))
	buf.WriteString(fmt.Sprintf("   /ColorSpace %s\n", imgObj.ColorSpace))
	buf.WriteString(fmt.Sprintf("   /BitsPerComponent %d\n", imgObj.BitsPerComp))

	if imgObj.Filter != "" {
		buf.WriteString(fmt.Sprintf("   /Filter %s\n", imgObj.Filter))
	}

	buf.WriteString(fmt.Sprintf("   /Length %d\n", len(encryptedData)))
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
	contentStream.WriteString(fmt.Sprintf("%s 0 0 %s %s %s cm\n",
		fmtNumImg(imageWidth), fmtNumImg(imageHeight), fmtNumImg(imageX), fmtNumImg(imageY)))

	// Draw the image using the XObject reference
	contentStream.WriteString(fmt.Sprintf("%s Do\n", imageXObjectRef))

	// Restore graphics state
	contentStream.WriteString("Q\n")

	pageManager.CurrentYPos -= imageHeight // No extra spacing
}
