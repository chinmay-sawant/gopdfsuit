package handlers

import (
	"archive/zip"
	"bytes"
	"log"
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
	"github.com/gin-gonic/gin"
)

// handleMergePDFs accepts multiple 'pdf' form files, merges them into a single PDF,
// and returns the merged PDF as application/pdf
func handleMergePDFs(c *gin.Context) {
	// Parse multipart form (let Gin handle it) - use Request.MultipartReader via FormFile in a loop
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("handleMergePDFs: invalid multipart form: %v", err)
		abortError(c, http.StatusBadRequest, "invalid request")
		return
	}

	files := form.File["pdf"]
	if len(files) == 0 {
		abortError(c, http.StatusBadRequest, "No pdf files provided; use field name 'pdf' multiple times")
		return
	}

	var pdfBytesList [][]byte
	// Process files in the exact order they appear in the form to maintain selection sequence
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			log.Printf("handleMergePDFs: open upload failed: %v", err)
			abortErrorAndStop(c, http.StatusInternalServerError, "failed to process upload")
			return
		}
		buf, ok, err := pdfService.ReadUpload(f, UploadKindPDF)
		_ = f.Close()
		if err != nil {
			log.Printf("handleMergePDFs: read upload failed: %v", err)
			abortErrorAndStop(c, http.StatusBadRequest, "invalid request")
			return
		}
		if !ok {
			abortErrorAndStop(c, http.StatusRequestEntityTooLarge, "pdf exceeds maximum size")
			return
		}
		pdfBytesList = append(pdfBytesList, buf)
	}

	merged, err := pdfService.MergePDFs(pdfBytesList)
	if err != nil {
		abortPDFError(c, err)
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=merged.pdf")
	c.Data(http.StatusOK, mimeTypePDF, merged)
}

// handlerSplitPDF accepts a 'pdf' file and splits it according to optional 'pages' and 'max_per_file' form fields,
// and returns the resulting PDFs in a zip file as application/zip
func handlerSplitPDF(c *gin.Context) {
	// Read uploaded PDF file
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
		log.Printf("handlerSplitPDF: read upload failed: %v", err)
		abortError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if !ok {
		abortError(c, http.StatusRequestEntityTooLarge, "pdf exceeds maximum size")
		return
	}

	// Optional page spec string and max per file
	pagesSpec := c.PostForm("pages") // e.g. "1-3,5"
	maxPerFile := 0
	if v := c.PostForm("max_per_file"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxPerFile = n
		}
	}

	// Parse pages into []int
	pages, err := merge.ParsePageSpec(pagesSpec, 0)
	if err != nil {
		abortError(c, http.StatusBadRequest, "Invalid pages spec: "+err.Error())
		return
	}

	spec := merge.SplitSpec{
		Pages:      pages,
		MaxPerFile: maxPerFile,
	}

	outs, err := pdfService.SplitPDF(pdfBytes, spec)
	if err != nil {
		abortPDFError(c, err)
		return
	}

	// If single output, return directly as PDF
	if len(outs) == 1 {
		c.Header("Content-Type", mimeTypePDF)
		c.Header("Content-Disposition", "attachment; filename=split.pdf")
		c.Data(http.StatusOK, mimeTypePDF, outs[0])
		return
	}

	// Multiple outputs: return a zip archive
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	var numBuf [20]byte
	var zipErr string
	for i, b := range outs {
		name := "originalfile-part" + string(strconv.AppendInt(numBuf[:0], int64(i+1), 10)) + ".pdf"
		fw, err := zw.Create(name)
		if err != nil {
			zipErr = "zip create failed: " + err.Error()
			break
		}
		if _, err := fw.Write(b); err != nil {
			zipErr = "zip write failed: " + err.Error()
			break
		}
	}
	_ = zw.Close()
	if zipErr != "" {
		abortError(c, http.StatusInternalServerError, zipErr)
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=splits.zip")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}
