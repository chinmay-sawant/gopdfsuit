package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleFillPDF accepts multipart form data with fields 'pdf' and 'xfdf' (files or raw bytes)
// and returns the filled PDF bytes as application/pdf
func handleFillPDF(c *gin.Context) {
	// Try multipart form file upload (bounded reads; oversized yields 413)
	pdfFile, _, _ := c.Request.FormFile("pdf")
	var pdfBytes []byte
	if pdfFile != nil {
		defer func() {
			_ = pdfFile.Close()
		}()
		data, ok, err := pdfService.ReadUpload(pdfFile, UploadKindPDF)
		if err != nil {
			log.Printf("handleFillPDF: read pdf failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if !ok {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "pdf exceeds maximum size"})
			return
		}
		pdfBytes = data
	}

	xfdfFile, _, _ := c.Request.FormFile("xfdf")
	var xfdfBytes []byte
	if xfdfFile != nil {
		defer func() {
			_ = xfdfFile.Close()
		}()
		data, ok, err := pdfService.ReadUpload(xfdfFile, UploadKindXFDF)
		if err != nil {
			log.Printf("handleFillPDF: read xfdf failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if !ok {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "xfdf exceeds maximum size"})
			return
		}
		xfdfBytes = data
	}

	// If files not provided, try to read raw body fields
	if len(pdfBytes) == 0 {
		if b := c.PostForm("pdf_bytes"); b != "" {
			if int64(len(b)) > pdfService.UploadLimit(UploadKindPDF) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "pdf exceeds maximum size"})
				return
			}
			pdfBytes = []byte(b)
		}
	}
	if len(xfdfBytes) == 0 {
		if b := c.PostForm("xfdf_bytes"); b != "" {
			if int64(len(b)) > pdfService.UploadLimit(UploadKindXFDF) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "xfdf exceeds maximum size"})
				return
			}
			xfdfBytes = []byte(b)
		}
	}

	if len(pdfBytes) == 0 || len(xfdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing pdf or xfdf data"})
		return
	}

	out, err := pdfService.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		abortPDFError(c, err)
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=filled.pdf")
	c.Data(http.StatusOK, mimeTypePDF, out)
}
