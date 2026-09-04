package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleFillPDF accepts multipart form data with fields 'pdf' and 'xfdf' (files or raw bytes)
// and returns the filled PDF bytes as application/pdf
func handleFillPDF(c *gin.Context) {
	// Multipart file uploads read through the shared upload policy; when a
	// field is absent the raw body-field fallback below still applies.
	var pdfBytes []byte
	if _, err := c.FormFile("pdf"); err == nil {
		pdfBytes = readSingleUpload(c, "pdf", UploadKindPDF)
		if pdfBytes == nil {
			return
		}
	}

	var xfdfBytes []byte
	if _, err := c.FormFile("xfdf"); err == nil {
		xfdfBytes = readSingleUpload(c, "xfdf", UploadKindXFDF)
		if xfdfBytes == nil {
			return
		}
	}

	// If files not provided, try to read raw body fields
	if len(pdfBytes) == 0 {
		if b := c.PostForm("pdf_bytes"); b != "" {
			if int64(len(b)) > pdfService.UploadLimit(UploadKindPDF) {
				abortError(c, http.StatusRequestEntityTooLarge, "pdf exceeds maximum size")
				return
			}
			pdfBytes = []byte(b)
		}
	}
	if len(xfdfBytes) == 0 {
		if b := c.PostForm("xfdf_bytes"); b != "" {
			if int64(len(b)) > pdfService.UploadLimit(UploadKindXFDF) {
				abortError(c, http.StatusRequestEntityTooLarge, "xfdf exceeds maximum size")
				return
			}
			xfdfBytes = []byte(b)
		}
	}

	if len(pdfBytes) == 0 || len(xfdfBytes) == 0 {
		abortError(c, http.StatusBadRequest, "Missing pdf or xfdf data")
		return
	}

	out, err := pdfService.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		abortPDFError(c, err, "PDF processing failed")
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=filled.pdf")
	c.Data(http.StatusOK, mimeTypePDF, out)
}
