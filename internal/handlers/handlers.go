// Package handlers provides HTTP handlers for the application.
package handlers

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/middleware"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/merge"
	"github.com/gin-gonic/gin"
)

const mimeTypePDF = "application/pdf"

var templatePDFPool = sync.Pool{
	New: func() any {
		return new(models.PDFTemplate)
	},
}

var templateDataCache sync.Map

var pprofForbiddenResp = gin.H{"error": "Forbidden: Pprof is only accessible from localhost"}

// resetTemplate clears a pooled PDFTemplate before unmarshal (and before Put) so
// omitted JSON fields do not leak from prior requests while still retaining
// hot backing arrays for the next pooled decode.
func resetTemplate(t *models.PDFTemplate) {
	if t == nil {
		return
	}
	t.ResetForReuse()
}

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
	// Add cache headers for static assets
	staticWithCache := func(relativePath, root string) {
		handler := http.FileServer(http.Dir(root))
		router.GET(relativePath+"/*filepath", func(c *gin.Context) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			handler.ServeHTTP(c.Writer, c.Request)
		})
	}
	staticWithCache("/gopdfsuit/assets", filepath.Join(base, "docs", "assets"))
	staticWithCache("/assets", filepath.Join(base, "docs", "assets")) // Fallback for backward compatibility

	// Benchmark fast path: skip CORS/auth middleware on template-pdf (GIN_FAST_API=1).
	fastAPI := os.Getenv("GIN_FAST_API") == "1"
	if fastAPI {
		router.POST("/api/v1/generate/template-pdf", handleGenerateTemplatePDF)
	}

	// API endpoints - protected with Google OAuth when running on Cloud Run
	v1 := router.Group("/api/v1")
	v1.Use(middleware.CORSMiddleware())       // Add CORS middleware
	v1.Use(middleware.GoogleAuthMiddleware()) // Only enforces auth on Cloud Run
	{
		// Handle all OPTIONS requests for CORS
		v1.OPTIONS("/*path", func(c *gin.Context) { //nolint:revive
			// Handled by CORSMiddleware
		})

		if !fastAPI {
			v1.POST("/generate/template-pdf", handleGenerateTemplatePDF)
		}
		v1.POST("/fill", handleFillPDF)
		v1.POST("/merge", handleMergePDFs)
		v1.POST("/split", handlerSplitPDF)
		v1.GET("/template-data", handleGetTemplateData)
		v1.GET("/fonts", handleGetFonts)
		v1.POST("/fonts", handleUploadFont)

		// HTML to PDF/Image endpoints (powered by gochromedp)
		v1.POST("/htmltopdf", handleHTMLToPDF)
		v1.POST("/htmltoimage", handleHTMLToImage)

		// Redaction endpoints
		v1.POST("/redact/page-info", HandleRedactPageInfo)
		v1.POST("/redact/text-positions", HandleRedactTextPositions)
		v1.POST("/redact/capabilities", HandleRedactCapabilities)
		v1.POST("/redact/apply", HandleRedactApply)
		v1.POST("/redact/search", HandleRedactSearch)
	}

	// Add pprof routes for profiling
	pprofGroup := router.Group("/debug/pprof")
	// Restrict pprof access to localhost only
	pprofGroup.Use(func(c *gin.Context) {
		clientIP := c.ClientIP()
		if clientIP != "127.0.0.1" && clientIP != "::1" {
			c.AbortWithStatusJSON(http.StatusForbidden, pprofForbiddenResp)
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
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
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
	filePath := filepath.Clean(filepath.Join(getProjectRoot(), filename))
	if !strings.HasPrefix(filePath, filepath.Clean(getProjectRoot())) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}

	if cached, ok := templateDataCache.Load(filePath); ok {
		c.Header("Content-Type", "application/json")
		c.Data(http.StatusOK, "application/json", cached.([]byte))
		return
	}

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template file not found: " + filename})
		return
	}

	// Validate JSON structure using sonic for performance
	var template models.PDFTemplate
	if err := sonic.Unmarshal(data, &template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON template format: " + err.Error()})
		return
	}

	templateDataCache.Store(filePath, data)

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

	var data []byte
	if file.Size > 0 {
		buf := bytes.NewBuffer(make([]byte, 0, file.Size))
		if _, err := io.Copy(buf, f); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
			return
		}
		data = buf.Bytes()
	} else {
		buf := bytes.NewBuffer(make([]byte, 0, 128<<10))
		if _, err := io.Copy(buf, f); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
			return
		}
		data = buf.Bytes()
	}

	// Register font
	fontName := file.Filename[:len(file.Filename)-len(ext)]
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
	template := templatePDFPool.Get().(*models.PDFTemplate)
	resetTemplate(template)
	defer func() {
		resetTemplate(template)
		templatePDFPool.Put(template)
	}()

	tier := c.GetHeader("X-Payload-Tier")
	if cl := c.Request.ContentLength; cl > 0 || tier != "" {
		template.PreallocForDecode(int(cl), tier)
	}

	if err := decodeTemplateJSON(c.Request.Body, int(c.Request.ContentLength), tier, template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template data: " + err.Error()})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=generated.pdf")

	if _, ok := pdfService.(defaultPDFService); ok {
		doc, err := pdf.GenerateTemplatePDFBorrowed(*template)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + err.Error()})
			return
		}
		defer doc.Release()
		c.Status(http.StatusOK)
		if _, err := c.Writer.Write(doc.Bytes()); err != nil {
			return
		}
		return
	}

	pdfBytes, err := pdfService.GenerateTemplatePDF(*template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + err.Error()})
		return
	}
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
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
			pdfBytes = unsafe.Slice(unsafe.StringData(b), len(b))
		}
	}
	if len(xfdfBytes) == 0 {
		if b := c.PostForm("xfdf_bytes"); b != "" {
			xfdfBytes = unsafe.Slice(unsafe.StringData(b), len(b))
		}
	}

	if len(pdfBytes) == 0 || len(xfdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing pdf or xfdf data"})
		return
	}

	out, err := pdfService.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=filled.pdf")
	c.Data(http.StatusOK, mimeTypePDF, out)
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
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file: " + err.Error()})
			return
		}
		buf, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file: " + err.Error()})
			return
		}
		pdfBytesList = append(pdfBytesList, buf)
	}

	merged, err := pdfService.MergePDFs(pdfBytesList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=merged.pdf")
	c.Data(http.StatusOK, mimeTypePDF, merged)
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

	outs, err := pdfService.SplitPDF(pdfBytes, spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": zipErr})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=splits.zip")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())

}

// handleHTMLToPDF handles HTML to PDF conversion using htmltopdf
func handleHTMLToPDF(c *gin.Context) {

	var req models.HTMLToPDFRequest
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request data: " + err.Error()})
		return
	}

	if err := sonic.Unmarshal(data, &req); err != nil {
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
		req.MarginTop = "10mm" //nolint:goconst
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

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=converted.pdf")
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}

// handleHTMLToImage handles HTML to image conversion using htmltoimage
func handleHTMLToImage(c *gin.Context) {

	var req models.HTMLToImageRequest
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request data: " + err.Error()})
		return
	}

	if err := sonic.Unmarshal(data, &req); err != nil {
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
	c.Header("Content-Disposition", "attachment; filename=converted."+req.Format)
	c.Data(http.StatusOK, contentType, imageBytes)
}
