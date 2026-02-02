package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/middleware"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/merge"
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
			if info, err := os.Stat(filepath.Join(cur, "docs")); err == nil && info.IsDir() {
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
		if info, err := os.Stat(filepath.Join(twoUp, "docs")); err == nil && info.IsDir() {
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

	// Serve static assets from Vite build (matching the base path in vite.config.js)
	router.Static("/gopdfsuit/assets", filepath.Join(base, "docs", "assets"))
	router.Static("/assets", filepath.Join(base, "docs", "assets")) // Fallback for backward compatibility

	// API endpoints - protected with Google OAuth when running on Cloud Run
	v1 := router.Group("/api/v1")
	v1.Use(middleware.CORSMiddleware())       // Add CORS middleware
	v1.Use(middleware.GoogleAuthMiddleware()) // Only enforces auth on Cloud Run
	{
		// Handle all OPTIONS requests for CORS
		v1.OPTIONS("/*path", func(c *gin.Context) {
			// Handled by CORSMiddleware
		})

		v1.POST("/generate/template-pdf", handleGenerateTemplatePDF)
		v1.POST("/fill", handleFillPDF)
		v1.POST("/merge", handleMergePDFs)
		v1.POST("/split", handlerSplitPDF)
		v1.GET("/template-data", handleGetTemplateData)
		v1.GET("/fonts", handleGetFonts)
		v1.POST("/fonts", handleUploadFont)

		// HTML to PDF/Image endpoints (powered by gochromedp)
		v1.POST("/htmltopdf", handlehtmlToPDF)
		v1.POST("/htmltoimage", handlehtmlToImage)
	}

	// Add pprof routes for profiling
	pprofGroup := router.Group("/debug/pprof")
	// Restrict pprof access to localhost only
	pprofGroup.Use(func(c *gin.Context) {
		clientIP := c.ClientIP()
		if clientIP != "127.0.0.1" && clientIP != "::1" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Pprof is only accessible from localhost"})
			return
		}
		c.Next()
	})
	{
		pprofGroup.GET("/", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/cmdline", gin.WrapF(http.HandlerFunc(pprof.Cmdline)))
		pprofGroup.GET("/profile", gin.WrapF(http.HandlerFunc(pprof.Profile)))
		pprofGroup.GET("/symbol", gin.WrapF(http.HandlerFunc(pprof.Symbol)))
		pprofGroup.POST("/symbol", gin.WrapF(http.HandlerFunc(pprof.Symbol)))
		pprofGroup.GET("/trace", gin.WrapF(http.HandlerFunc(pprof.Trace)))
		pprofGroup.GET("/heap", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/goroutine", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/allocs", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/block", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/mutex", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/threadcreate", gin.WrapF(http.HandlerFunc(pprof.Index)))
	}

	// Redirect root path to /gopdfsuit
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/gopdfsuit")
	})

	// Serve React app for all frontend routes (SPA fallback)
	router.NoRoute(handleSPA)
}

// handleSPA serves the React SPA for all frontend routes
func handleSPA(c *gin.Context) {
	base := getProjectRoot()
	indexPath := filepath.Join(base, "docs", "index.html")

	// Check if the file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Frontend not built. Please run 'npm run build' in the frontend directory.",
		})
		return
	}

	c.File(indexPath)
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

// handleGetFonts returns the list of available fonts for PDF generation
func handleGetFonts(c *gin.Context) {
	// Get available fonts from the pdf package
	fonts := pdf.GetAvailableFonts()

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

	// Read file content
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file: " + err.Error()})
		return
	}
	defer func() {
		_ = f.Close()
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
		return
	}

	// Register font
	fontName := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	err = pdf.GetFontRegistry().RegisterFontFromData(fontName, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to register font: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Font uploaded successfully",
		"name":    fontName,
	})
}

func handleGenerateTemplatePDF(c *gin.Context) {
	var template models.PDFTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template data: " + err.Error()})
		return
	}

	pdfBytes, err := pdf.GenerateTemplatePDF(template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=generated.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// handleFillPDF accepts multipart form data with fields 'pdf' and 'xfdf' (files or raw bytes)
// and returns the filled PDF bytes as application/pdf
func handleFillPDF(c *gin.Context) {
	// Try multipart form file upload
	pdfFile, pdfHeader, _ := c.Request.FormFile("pdf")
	var pdfBytes []byte
	if pdfFile != nil {
		defer func() {
			_ = pdfFile.Close()
		}()
		buf := make([]byte, pdfHeader.Size)
		_, err := pdfFile.Read(buf)
		if err == nil {
			pdfBytes = buf
		}
	}

	xfdfFile, xfdfHeader, _ := c.Request.FormFile("xfdf")
	var xfdfBytes []byte
	if xfdfFile != nil {
		defer func() {
			_ = xfdfFile.Close()
		}()
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
		_ = f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file: " + err.Error()})
			return
		}
		pdfBytesList = append(pdfBytesList, buf)
	}

	merged, err := merge.MergePDFs(pdfBytesList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=merged.pdf")
	c.Data(http.StatusOK, "application/pdf", merged)
}

// handleSplitPDF accepts a 'pdf' file and splits it according to optional 'pages' and 'max_per_file' form fields,
// and returns the resulting PDFs in a zip file as application/zip
func handlerSplitPDF(c *gin.Context) {
	// Read uploaded PDF file
	pdfFile, _, err := c.Request.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing pdf file: " + err.Error()})
		return
	}
	defer func() {
		_ = pdfFile.Close()
	}()
	pdfBytes, err := io.ReadAll(pdfFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read pdf: " + err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pages spec: " + err.Error()})
		return
	}

	spec := merge.SplitSpec{
		Pages:      pages,
		MaxPerFile: maxPerFile,
	}

	outs, err := merge.SplitPDF(pdfBytes, spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If single output, return directly as PDF
	if len(outs) == 1 {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=split.pdf")
		c.Data(http.StatusOK, "application/pdf", outs[0])
		return
	}

	// Multiple outputs: return a zip archive
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for i, b := range outs {
		name := fmt.Sprintf("originalfile-part%d.pdf", i+1)
		fw, err := zw.Create(name)
		if err != nil {
			_ = zw.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "zip create failed: " + err.Error()})
			return
		}
		if _, err := fw.Write(b); err != nil {
			_ = zw.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "zip write failed: " + err.Error()})
			return
		}
	}
	_ = zw.Close()

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=splits.zip")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())

}

// handlehtmlToPDF handles HTML to PDF conversion using htmltopdf
func handlehtmlToPDF(c *gin.Context) {
	log.Printf("Starting HTML to PDF conversion request")

	var req models.HtmlToPDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	log.Printf("Request parsed successfully. HTML length: %d, URL: %s", len(req.HTML), req.URL)

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

	log.Printf("Calling pdf.ConvertHTMLToPDF with options: PageSize=%s, Orientation=%s, DPI=%d", req.PageSize, req.Orientation, req.DPI)

	pdfBytes, err := pdf.ConvertHTMLToPDF(req)
	if err != nil {
		log.Printf("PDF conversion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF conversion failed: " + err.Error()})
		return
	}

	log.Printf("PDF conversion successful. PDF size: %d bytes", len(pdfBytes))

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=converted.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// handlehtmlToImage handles HTML to image conversion using htmltoimage
func handlehtmlToImage(c *gin.Context) {
	log.Printf("Starting HTML to image conversion request")

	var req models.HtmlToImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	log.Printf("Request parsed successfully. HTML length: %d, URL: %s, Format: %s", len(req.HTML), req.URL, req.Format)

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

	log.Printf("Calling pdf.ConvertHTMLToImage with options: Format=%s, Quality=%d, Zoom=%.2f", req.Format, req.Quality, req.Zoom)

	imageBytes, err := pdf.ConvertHTMLToImage(req)
	if err != nil {
		log.Printf("Image conversion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Image conversion failed: " + err.Error()})
		return
	}

	log.Printf("Image conversion successful. Image size: %d bytes", len(imageBytes))

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
