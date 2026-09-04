package handlers

import (
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
	"github.com/gin-gonic/gin"
)

// handleCompressPDF accepts a single 'pdf' form file and optional level /
// quality / max_image_dim fields, then returns a compressed PDF.
func handleCompressPDF(c *gin.Context) {
	pdfBytes := readSingleUpload(c, "pdf", UploadKindPDF)
	if pdfBytes == nil {
		return
	}

	// Level names route through the gopdflib tier vocabulary so handler
	// input and the Go/CGO/WASM entry points share one policy: empty or
	// unknown selects Medium.
	var opts compress.Options
	if v := c.PostForm("level"); v != "" {
		opts.Level = compress.Level(gopdflib.ParseCompressLevel(v))
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
		abortPDFError(c, err, "PDF processing failed")
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=compressed.pdf")
	c.Data(http.StatusOK, mimeTypePDF, out)
}
