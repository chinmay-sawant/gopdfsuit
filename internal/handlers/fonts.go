package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// handleGetFonts returns the list of available fonts for PDF generation
func handleGetFonts(c *gin.Context) {
	// Get available fonts via the PDF service (mockable in tests)
	fonts := pdfService.GetFonts()

	c.JSON(http.StatusOK, gin.H{
		"fonts": fonts,
	})
}

// handleUploadFont handles the upload of custom font files
func handleUploadFont(c *gin.Context) {
	if !ensureMultipartBodyLimit(c) {
		return
	}
	file, err := c.FormFile("font")
	if err != nil {
		if isBodyTooLargeErr(err) {
			abortError(c, http.StatusRequestEntityTooLarge, overLimitMessage(UploadKindFont))
			return
		}
		abortError(c, http.StatusBadRequest, "font file is required")
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".ttf" && ext != ".otf" {
		abortError(c, http.StatusBadRequest, "Only .ttf and .otf files are supported")
		return
	}

	// Read file content through the shared upload policy (413 on over-limit).
	data := readUploadData(c, file, UploadKindFont)
	if data == nil {
		return
	}

	// Register font
	fontName := file.Filename[:len(file.Filename)-len(ext)]
	err = pdfService.RegisterFont(fontName, data)
	if err != nil {
		log.Printf("handleUploadFont: register %q failed: %v", fontName, err)
		abortError(c, http.StatusBadRequest, "invalid font data")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Font uploaded successfully",
		"name":    fontName,
	})
}
