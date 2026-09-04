// Package handlers provides HTTP handlers for the application.
package handlers

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/middleware"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
	"github.com/gin-gonic/gin"
)

const mimeTypePDF = "application/pdf"

const (
	// maxPDFBytes caps single PDF uploads (mirrors compress.MaxInputBytes).
	maxPDFBytes = int64(compress.MaxInputBytes)
	// maxXFDFBytes caps XFDF form-data uploads.
	maxXFDFBytes = 8 << 20
	// maxFontBytes caps custom font uploads.
	maxFontBytes = 10 << 20
	// maxHTMLBodyBytes caps HTML-to-PDF/Image JSON request bodies.
	maxHTMLBodyBytes = 2 << 20
)

// errFetchURLBlocked is returned when a requested fetch URL targets a
// non-public address (SSRF guard).
var errFetchURLBlocked = errors.New("url target is not allowed")

// readBounded reads r capped at limit+1 bytes. ok=false means the input
// exceeded limit (caller should reject, e.g. with 413).
func readBounded(r io.Reader, limit int64) (data []byte, ok bool, err error) {
	data, err = io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > limit {
		return nil, false, nil
	}
	return data, true, nil
}

// pdfErrorStatus maps backend failures to HTTP codes: malformed client input
// yields 422, engine failures yield 500.
func pdfErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	msg := strings.ToLower(err.Error())
	for _, s := range []string{
		"invalid", "malformed", "corrupt", "parse", "password", "encrypt",
		"spec", "page", "range", "empty", "not a pdf", "unsupported",
		"too small", "trailer", "xref", "header", "crypt",
	} {
		if strings.Contains(msg, s) {
			return http.StatusUnprocessableEntity
		}
	}
	return http.StatusInternalServerError
}

// abortPDFError logs backend detail server-side and replies with a generic
// client message plus the mapped status code.
func abortPDFError(c *gin.Context, err error) {
	log.Printf("pdf handler %s failed: %v", c.Request.URL.Path, err)
	if pdfErrorStatus(err) == http.StatusUnprocessableEntity {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid PDF input"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF processing failed"})
}

// validateFetchURL allows only http/https URLs whose host does not resolve to
// loopback, private, link-local (incl. cloud metadata), or multicast targets.
func validateFetchURL(raw string) error {
	if raw == "" {
		return nil // HTML-content path: no fetch performed.
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url scheme must be http or https")
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("invalid url")
	}
	lower := strings.ToLower(strings.TrimSuffix(host, "."))
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return errFetchURLBlocked
	}
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedFetchIP(ip) {
			return errFetchURLBlocked
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return errors.New("invalid url")
	}
	for _, ip := range ips {
		if isBlockedFetchIP(ip) {
			return errFetchURLBlocked
		}
	}
	return nil
}

func isBlockedFetchIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}
	if ip.IsPrivate() || !ip.IsGlobalUnicast() {
		return true
	}
	return false
}

var templatePDFPool = sync.Pool{
	New: func() any {
		return new(models.PDFTemplate)
	},
}

var templateDataCache sync.Map

var pprofForbiddenResp = gin.H{"error": "Forbidden: Pprof is only accessible from localhost"}

// isLoopbackPeer reports whether req arrived over a loopback connection,
// based on the direct peer address. Unlike Gin's ClientIP it ignores
// X-Forwarded-For, which any client can forge when trusted proxies include
// public ranges (Gin's default).
func isLoopbackPeer(req *http.Request) bool {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

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

	// Benchmark fast path (GIN_FAST_API=1): skip extra non-auth middleware such
	// as CORS, but NEVER skip authentication. The template-pdf route always
	// lives inside the v1 auth group so GoogleAuthMiddleware still runs.
	fastAPI := os.Getenv("GIN_FAST_API") == "1"

	// API endpoints - protected with Google OAuth when running on Cloud Run
	v1 := router.Group("/api/v1")
	if !fastAPI {
		v1.Use(middleware.CORSMiddleware()) // Add CORS middleware
	}
	v1.Use(middleware.GoogleAuthMiddleware()) // Only enforces auth on Cloud Run (or REQUIRE_AUTH=1)
	{
		// Handle all OPTIONS requests for CORS
		v1.OPTIONS("/*path", func(c *gin.Context) { //nolint:revive
			// Handled by CORSMiddleware
		})

		v1.POST("/generate/template-pdf", handleGenerateTemplatePDF)
		v1.POST("/fill", handleFillPDF)
		v1.POST("/merge", handleMergePDFs)
		v1.POST("/split", handlerSplitPDF)
		v1.POST("/compress", handleCompressPDF)
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
	// Restrict pprof access to localhost only. The check uses the direct
	// peer address (RemoteAddr), not ClientIP: Gin trusts X-Forwarded-For
	// by default, so ClientIP is spoofable from anywhere behind a proxy.
	pprofGroup.Use(func(c *gin.Context) {
		if !isLoopbackPeer(c.Request) {
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
		log.Printf("handleGetTemplateData: invalid template %s: %v", filename, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template data"})
		return
	}

	templateDataCache.Store(filePath, data)

	// Return the JSON data
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}

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

	data, ok, err := readBounded(f, maxFontBytes)
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

	// Bound the JSON body before streaming decode.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTemplateJSONBody)
	if err := decodeTemplateJSON(c.Request.Body, int(c.Request.ContentLength), tier, template); err != nil {
		if isBodyTooLargeErr(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "template too large"})
			return
		}
		log.Printf("handleGenerateTemplatePDF: invalid template: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template data"})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=generated.pdf")

	if _, ok := pdfService.(defaultPDFService); ok {
		doc, err := pdf.GenerateTemplatePDFBorrowed(*template)
		if err != nil {
			log.Printf("handleGenerateTemplatePDF: generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
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
		log.Printf("handleGenerateTemplatePDF: generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
		return
	}
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}

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
		data, ok, err := readBounded(pdfFile, maxPDFBytes)
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
		data, ok, err := readBounded(xfdfFile, maxXFDFBytes)
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
			if int64(len(b)) > maxPDFBytes {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "pdf exceeds maximum size"})
				return
			}
			pdfBytes = []byte(b)
		}
	}
	if len(xfdfBytes) == 0 {
		if b := c.PostForm("xfdf_bytes"); b != "" {
			if int64(len(b)) > maxXFDFBytes {
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

// handleMergePDFs accepts multiple 'pdf' form files, merges them into a single PDF,
// and returns the merged PDF as application/pdf
func handleMergePDFs(c *gin.Context) {
	// Parse multipart form (let Gin handle it) - use Request.MultipartReader via FormFile in a loop
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("handleMergePDFs: invalid multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
			log.Printf("handleMergePDFs: open upload failed: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to process upload"})
			return
		}
		buf, ok, err := readBounded(f, maxPDFBytes)
		_ = f.Close()
		if err != nil {
			log.Printf("handleMergePDFs: read upload failed: %v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "pdf exceeds maximum size"})
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
	pdfBytes, ok, err := readBounded(pdfFile, maxPDFBytes)
	if err != nil {
		log.Printf("handlerSplitPDF: read upload failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if !ok {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "pdf exceeds maximum size"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": zipErr})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=splits.zip")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())

}

// handleCompressPDF accepts a single 'pdf' form file and optional level /
// quality / max_image_dim fields, then returns a compressed PDF.
func handleCompressPDF(c *gin.Context) {
	pdfFile, _, err := c.Request.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing pdf file: " + err.Error()})
		return
	}
	defer func() {
		_ = pdfFile.Close()
	}()
	pdfBytes, err := io.ReadAll(io.LimitReader(pdfFile, int64(compress.MaxInputBytes)+1))
	if err != nil {
		log.Printf("handleCompressPDF: read upload failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(pdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is empty"})
		return
	}
	if len(pdfBytes) > compress.MaxInputBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "pdf exceeds maximum size"})
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

// handleHTMLToPDF handles HTML to PDF conversion using htmltopdf
func handleHTMLToPDF(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxHTMLBodyBytes)

	var req models.HTMLToPDFRequest
	data, err := c.GetRawData()
	if err != nil {
		if isBodyTooLargeErr(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		log.Printf("handleHTMLToPDF: read body failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := sonic.Unmarshal(data, &req); err != nil {
		log.Printf("handleHTMLToPDF: invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.HTML == "" && req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either html or url is required"})
		return
	}
	if err := validateFetchURL(req.URL); err != nil {
		if errors.Is(err, errFetchURLBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "url target is not allowed"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
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

	pdfBytes, err := pdfService.HTMLToPDF(req)
	if err != nil {
		log.Printf("handleHTMLToPDF: conversion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF conversion failed"})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=converted.pdf")
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}

// handleHTMLToImage handles HTML to image conversion using htmltoimage
func handleHTMLToImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxHTMLBodyBytes)

	var req models.HTMLToImageRequest
	data, err := c.GetRawData()
	if err != nil {
		if isBodyTooLargeErr(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		log.Printf("handleHTMLToImage: read body failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := sonic.Unmarshal(data, &req); err != nil {
		log.Printf("handleHTMLToImage: invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.HTML == "" && req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either html or url is required"})
		return
	}
	if err := validateFetchURL(req.URL); err != nil {
		if errors.Is(err, errFetchURLBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "url target is not allowed"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
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

	imageBytes, err := pdfService.HTMLToImage(req)
	if err != nil {
		log.Printf("handleHTMLToImage: conversion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image conversion failed"})
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
