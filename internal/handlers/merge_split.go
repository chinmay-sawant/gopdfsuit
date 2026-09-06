package handlers

import (
	"archive/zip"
	"bytes"
	"log"
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/merge"
	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
	"github.com/gin-gonic/gin"
)

// handleMergePDFs accepts multiple 'pdf' form files, merges them into a single PDF,
// and returns the merged PDF as application/pdf
func handleMergePDFs(c *gin.Context) {
	if !applyBodyLimit(c, requestBodyLimit(c, maxMergeBodyBytes)) {
		return
	}
	// Parse multipart form (let Gin handle it) - use Request.MultipartReader via FormFile in a loop
	form, err := c.MultipartForm()
	if err != nil {
		if isBodyTooLargeErr(err) {
			abortError(c, http.StatusRequestEntityTooLarge, "merge request body too large")
			return
		}
		log.Printf("handleMergePDFs: invalid multipart form: %v", err)
		abortError(c, http.StatusBadRequest, "invalid request")
		return
	}

	files := form.File["pdf"]
	if len(files) == 0 {
		abortError(c, http.StatusBadRequest, "No pdf files provided; use field name 'pdf' multiple times")
		return
	}
	if len(files) > mergeFileCountLimit(c) {
		abortError(c, http.StatusRequestEntityTooLarge, "too many PDF files")
		return
	}

	pdfBytesList := make([][]byte, 0, len(files))
	var totalBytes int64
	// Process files in the exact order they appear in the form to maintain selection sequence
	for _, fh := range files {
		if fh.Size < 0 || fh.Size > mergeBodyLimit(c) || totalBytes > mergeBodyLimit(c)-fh.Size {
			abortError(c, http.StatusRequestEntityTooLarge, "merge input exceeds maximum size")
			return
		}
		totalBytes += fh.Size
		buf := readUploadData(c, fh, UploadKindPDF)
		if buf == nil {
			// Rejection already written by readUploadData.
			return
		}
		pdfBytesList = append(pdfBytesList, buf)
	}

	merged, err := pdfService.MergePDFs(pdfBytesList)
	if err != nil {
		abortPDFError(c, err, "PDF processing failed")
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=merged.pdf")
	c.Data(http.StatusOK, mimeTypePDF, merged)
}

func mergeBodyLimit(c *gin.Context) int64 {
	return requestBodyLimit(c, maxMergeBodyBytes)
}

// handlerSplitPDF accepts a 'pdf' file and splits it according to optional 'pages' and 'max_per_file' form fields,
// and returns the resulting PDFs in a zip file as application/zip
func handlerSplitPDF(c *gin.Context) {
	// Read uploaded PDF file
	pdfBytes := readSingleUpload(c, "pdf", UploadKindPDF)
	if pdfBytes == nil {
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

	// Page specs route through the gopdflib constructor so handler input
	// shares validation with the Go/CGO/WASM split entry points.
	pages, err := gopdflib.ParsePageSpec(pagesSpec, 0)
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
		abortPDFError(c, err, "PDF processing failed")
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
