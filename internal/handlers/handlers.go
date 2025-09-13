package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/internal/pdf"
	"github.com/gin-gonic/gin"
)

// getProjectRoot returns the base directory where the `web` folder lives.
// Resolution strategy:
// 1. If environment variable GOPDFSUIT_ROOT is set, use it.
// 2. Otherwise, use the current working directory.
// This keeps behavior predictable when the binary is run from cmd/gopdfsuit or
// from the repo root. If you need a different behavior, set GOPDFSUIT_ROOT.
func getProjectRoot() string {
	// 1. Env override
	if v := os.Getenv("GOPDFSUIT_ROOT"); v != "" {
		return v
	}

	// searchUp looks for a directory that contains `web` starting from start
	// and walking up at most maxDepth levels.
	searchUp := func(start string, maxDepth int) string {
		cur := start
		for i := 0; i <= maxDepth; i++ {
			if cur == "" || cur == string(filepath.Separator) {
				break
			}
			// if a web directory exists here, assume this is the project root
			if info, err := os.Stat(filepath.Join(cur, "web")); err == nil && info.IsDir() {
				return cur
			}
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			cur = parent
		}
		return ""
	}

	// 2. Try current working directory and walk up
	if wd, err := os.Getwd(); err == nil {
		if p := searchUp(wd, 6); p != "" {
			return p
		}
	}

	// 3. Try executable directory (useful when running the compiled binary)
	if exe, err := os.Executable(); err == nil {
		if p := searchUp(filepath.Dir(exe), 6); p != "" {
			return p
		}
	}

	// 4. Fallback: assume repo root is two levels above the cwd (common layout
	// when running from cmd/gopdfsuit), but only if that path exists.
	if wd, err := os.Getwd(); err == nil {
		twoUp := filepath.Clean(filepath.Join(wd, "..", ".."))
		if info, err := os.Stat(filepath.Join(twoUp, "web")); err == nil && info.IsDir() {
			return twoUp
		}
		return wd
	}

	return "."
}

// RegisterRoutes wires up API routes onto the provided Gin router.
func RegisterRoutes(router *gin.Engine) {
	// Resolve project base directory so paths work whether binary is run from
	// the repo root or from inside cmd/gopdfsuit (where the exe often lives).
	base := getProjectRoot()

	// Serve static files
	router.Static("/static", filepath.Join(base, "web", "static"))

	// Load HTML templates (use absolute path)
	router.LoadHTMLGlob(filepath.Join(base, "web", "templates", "*"))

	// Root endpoint for PDF viewer
	router.GET("/", handlePDFViewer)
	// PDF filler UI
	router.GET("/filler", handlePDFFiller)

	// HTML to PDF/Image UI pages (powered by gochromedp)
	router.GET("/htmltopdf", handlehtmlToPDFPage)
	router.GET("/htmltoimage", handlehtmlToImagePage)

	// API endpoints
	v1 := router.Group("/api/v1")
	v1.POST("/generate/template-pdf", handleGenerateTemplatePDF)
	v1.POST("/fill", handleFillPDF)
	v1.POST("/merge", handleMergePDFs)
	v1.GET("/template-data", handleGetTemplateData)

	// HTML to PDF/Image endpoints (powered by gochromedp)
	v1.POST("/htmltopdf", handlehtmlToPDF)
	v1.POST("/htmltoimage", handlehtmlToImage)
	v1.GET("/htmltopdf", handlehtmlToPDF)
	v1.GET("/htmltoimage", handlehtmlToImage)

	// Template editor endpoint
	router.GET("/editor", PDFEditorHandler)
	// PDF merge UI
	router.GET("/merge", PDFMergeHandler)
}

// handlePDFViewer serves the PDF viewer HTML page
func handlePDFViewer(c *gin.Context) {
	c.HTML(http.StatusOK, "pdf_viewer.html", gin.H{
		"title": "GoPdfSuit - PDF Viewer",
	})
}

// handlePDFFiller serves the PDF filler UI page
func handlePDFFiller(c *gin.Context) {
	c.HTML(http.StatusOK, "pdf_filler.html", gin.H{
		"title": "GoPdfSuit - PDF Filler",
	})
}

// handleGetTemplateData serves JSON template data based on file query parameter
func handleGetTemplateData(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'file' query parameter"})
		return
	}

	// Security: only allow specific file extensions and prevent path traversal
	if filepath.Ext(filename) != ".json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JSON files are allowed"})
		return
	}

	// Clean the filename to prevent path traversal and resolve against
	// project root so files at repository root are found when running the
	// server from cmd/gopdfsuit.
	filename = filepath.Base(filename)
	filePath := filepath.Join(getProjectRoot(), filename)

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template file not found: " + filename})
		return
	}

	// Validate JSON structure
	var template models.PDFTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON template format: " + err.Error()})
		return
	}

	// Return the JSON data
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}

func handleGenerateTemplatePDF(c *gin.Context) {
	var template models.PDFTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template data: " + err.Error()})
		return
	}
	pdf.GenerateTemplatePDF(c, template)
}

// handleFillPDF accepts multipart form data with fields 'pdf' and 'xfdf' (files or raw bytes)
// and returns the filled PDF bytes as application/pdf
func handleFillPDF(c *gin.Context) {
	// Try multipart form file upload
	pdfFile, pdfHeader, _ := c.Request.FormFile("pdf")
	var pdfBytes []byte
	if pdfFile != nil {
		defer pdfFile.Close()
		buf := make([]byte, pdfHeader.Size)
		_, err := pdfFile.Read(buf)
		if err == nil {
			pdfBytes = buf
		}
	}

	xfdfFile, xfdfHeader, _ := c.Request.FormFile("xfdf")
	var xfdfBytes []byte
	if xfdfFile != nil {
		defer xfdfFile.Close()
		buf := make([]byte, xfdfHeader.Size)
		_, err := xfdfFile.Read(buf)
		if err == nil {
			xfdfBytes = buf
		}
	}

	// If files not provided, try to read raw body fields
	if len(pdfBytes) == 0 {
		if b := c.PostForm("pdf_bytes"); b != "" {
			pdfBytes = []byte(b)
		}
	}
	if len(xfdfBytes) == 0 {
		if b := c.PostForm("xfdf_bytes"); b != "" {
			xfdfBytes = []byte(b)
		}
	}

	if len(pdfBytes) == 0 || len(xfdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing pdf or xfdf data"})
		return
	}

	out, err := pdf.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=filled.pdf")
	c.Data(http.StatusOK, "application/pdf", out)
}

// handleMergePDFs accepts multiple 'pdf' form files, merges them into a single PDF,
// and returns the merged PDF as application/pdf
func handleMergePDFs(c *gin.Context) {
	// Parse multipart form (let Gin handle it) - use Request.MultipartReader via FormFile in a loop
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form: " + err.Error()})
		return
	}

	files := form.File["pdf"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pdf files provided; use field name 'pdf' multiple times"})
		return
	}

	var pdfBytesList [][]byte
	// Process files in the exact order they appear in the form to maintain selection sequence
	for i, fh := range files {
		fmt.Printf("Processing file %d: %s\n", i+1, fh.Filename)
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file: " + err.Error()})
			return
		}
		buf, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file: " + err.Error()})
			return
		}
		pdfBytesList = append(pdfBytesList, buf)
	}

	merged, err := pdf.MergePDFs(pdfBytesList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=merged.pdf")
	c.Data(http.StatusOK, "application/pdf", merged)
}

// PDFEditorHandler serves the template editor interface
func PDFEditorHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "pdf_editor.html", gin.H{
		"title": "GoPdfSuit - Template Editor",
	})
}

// PDFMergeHandler serves the merge UI
func PDFMergeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "pdf_merge.html", gin.H{
		"title": "GoPdfSuit - PDF Merge",
	})
}

// handlehtmlToPDFPage serves the htmltopdf converter UI
func handlehtmlToPDFPage(c *gin.Context) {
	c.HTML(http.StatusOK, "htmltopdf.html", gin.H{
		"title": "GoPdfSuit - HTML to PDF Converter",
	})
}

// handlehtmlToImagePage serves the htmltoimage converter UI
func handlehtmlToImagePage(c *gin.Context) {
	c.HTML(http.StatusOK, "htmltoimage.html", gin.H{
		"title": "GoPdfSuit - HTML to Image Converter",
	})
}

// handlehtmlToPDF handles HTML to PDF conversion using htmltopdf
func handlehtmlToPDF(c *gin.Context) {
	var req models.HtmlToPDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Set defaults
	if req.PageSize == "" {
		req.PageSize = "A4"
	}
	if req.Orientation == "" {
		req.Orientation = "Portrait"
	}
	if req.MarginTop == "" {
		req.MarginTop = "10mm"
	}
	if req.MarginRight == "" {
		req.MarginRight = "10mm"
	}
	if req.MarginBottom == "" {
		req.MarginBottom = "10mm"
	}
	if req.MarginLeft == "" {
		req.MarginLeft = "10mm"
	}
	if req.DPI == 0 {
		req.DPI = 300
	}

	pdfBytes, err := pdf.ConvertHTMLToPDF(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF conversion failed: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=converted.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// handlehtmlToImage handles HTML to image conversion using htmltoimage
func handlehtmlToImage(c *gin.Context) {
	var req models.HtmlToImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Set defaults
	if req.Format == "" {
		req.Format = "png"
	}
	if req.Quality == 0 {
		req.Quality = 94
	}
	if req.Zoom == 0 {
		req.Zoom = 1.0
	}

	imageBytes, err := pdf.ConvertHTMLToImage(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Image conversion failed: " + err.Error()})
		return
	}

	contentType := "image/png"
	switch req.Format {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "svg":
		contentType = "image/svg+xml"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=converted.%s", req.Format))
	c.Data(http.StatusOK, contentType, imageBytes)
}
