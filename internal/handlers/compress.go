package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/gin-gonic/gin"
)

// handleCompressPDF accepts a single 'pdf' form file and optional level /
// quality / max_image_dim fields, then returns a compressed PDF.
func handleCompressPDF(c *gin.Context) {
	pdfFile, _, err := c.Request.FormFile("pdf")
	if err != nil {
		abortError(c, http.StatusBadRequest, "Missing pdf file: "+err.Error())
		return
	}
	defer func() {
		_ = pdfFile.Close()
	}()
	pdfBytes, ok, err := pdfService.ReadUpload(pdfFile, UploadKindPDF)
	if err != nil {
		log.Printf("handleCompressPDF: read upload failed: %v", err)
		abortError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if !ok {
		abortError(c, http.StatusRequestEntityTooLarge, "pdf exceeds maximum size")
		return
	}
	if len(pdfBytes) == 0 {
		abortError(c, http.StatusBadRequest, "pdf file is empty")
		return
	}

	var opts compress.Options
	if v := c.PostForm("level"); v != "" {
		opts.Level = compress.Level(v)
	}
	if v := c.PostForm("quality"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.JPEGQuality = n
		}
	}
	if v := c.PostForm("max_image_dim"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.MaxImageDim = n
		}
	}

	out, err := pdfService.CompressPDF(pdfBytes, opts)
	if err != nil {
		abortPDFError(c, err)
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=compressed.pdf")
	c.Data(http.StatusOK, mimeTypePDF, out)
}
