// Package handlers provides HTTP handlers for the application.
package handlers

import (
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/middleware"
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
		v1.POST("/redact/page-info", handleRedactPageInfo)
		v1.POST("/redact/text-positions", handleRedactTextPositions)
		v1.POST("/redact/capabilities", handleRedactCapabilities)
		v1.POST("/redact/apply", handleRedactApply)
		v1.POST("/redact/search", handleRedactSearch)
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
