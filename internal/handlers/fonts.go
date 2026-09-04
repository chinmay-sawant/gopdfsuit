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
	file, err := c.FormFile("font")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No font file provided"})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".ttf" && ext != ".otf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .ttf and .otf files are supported"})
		return
	}

	// Read file content, capped to reject oversized uploads with 413
	f, err := file.Open()
	if err != nil {
		log.Printf("handleUploadFont: open failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process font upload"})
		return
	}
	defer func() {
		_ = f.Close()
	}()

	data, ok, err := pdfService.ReadUpload(f, UploadKindFont)
	if err != nil {
		log.Printf("handleUploadFont: read failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if !ok {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "font file exceeds maximum size"})
		return
	}

	// Register font
	fontName := file.Filename[:len(file.Filename)-len(ext)]
	err = pdfService.RegisterFont(fontName, data)
	if err != nil {
		log.Printf("handleUploadFont: register %q failed: %v", fontName, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid font data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Font uploaded successfully",
		"name":    fontName,
	})
}
