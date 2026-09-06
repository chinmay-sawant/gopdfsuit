// Package handlers provides HTTP handlers for the application.
package handlers

import (
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/middleware"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
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

// RegisterRoutes wires up API routes onto the provided Gin router with the
// startup-resolved RoutePolicy (see ResolveServerConfig).
func RegisterRoutes(router *gin.Engine) {
	RegisterRoutesWithPolicy(router, ResolveServerConfig().Policy)
}

// RegisterRoutesWithPolicy wires up API routes with an explicit policy.
// Auth approach is env-based, not code removal: GoogleAuthMiddleware stays
// registered on the v1 group and self-gates via authEnforced() (public
// unless REQUIRE_AUTH=1 or K_SERVICE/K_REVISION is set). To enforce auth
// locally or in staging, set REQUIRE_AUTH=1 rather than editing routes.
// Benchmark fast path (!EnableCORS, i.e. GIN_FAST_API=1): skip extra
// non-auth middleware such as CORS, but NEVER skip authentication. The
// template-pdf route always lives inside the v1 auth group so
// GoogleAuthMiddleware still runs.
func RegisterRoutesWithPolicy(router *gin.Engine, policy RoutePolicy) {
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

	// API endpoints - protected with Google OAuth when running on Cloud Run.
	// policy.EnableCORS is false only on the GIN_FAST_API=1 benchmark path;
	// authentication is never skipped (see RegisterRoutesWithPolicy).
	v1 := router.Group("/api/v1")
	policy = normalizeRoutePolicy(policy)
	if policy.EnableCORS {
		v1.Use(middleware.CORSMiddleware()) // Add CORS middleware
	}
	v1.Use(routeAuthMiddleware(policy.RequireAuth))
	v1.Use(routeBodyLimitMiddleware(policy))
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

func routeBodyLimitMiddleware(policy RoutePolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		limit := int64(0)
		switch c.Request.URL.Path {
		case "/api/v1/generate/template-pdf":
			limit = policy.MaxTemplateBodyBytes
		case "/api/v1/htmltopdf", "/api/v1/htmltoimage":
			limit = policy.MaxHTMLBodyBytes
		case "/api/v1/merge":
			limit = policy.MaxMergeBodyBytes
			c.Set(mergeFileLimitContextKey, policy.MaxMergeFiles)
		default:
			limit = policy.MaxMultipartBodyBytes
		}
		if !applyBodyLimit(c, limit) {
			return
		}
		c.Next()
	}
}

func routeAuthMiddleware(requireAuth bool) gin.HandlerFunc {
	if !requireAuth {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format. Expected: Bearer <token>"})
			c.Abort()
			return
		}
		payload, err := idtoken.Validate(c.Request.Context(), parts[1], resolveAuthAudience())
		if err != nil {
			log.Printf("route authentication failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
			c.Abort()
			return
		}
		c.Set("user_email", payload.Claims["email"])
		c.Set("user_name", payload.Claims["name"])
		c.Set("user_picture", payload.Claims["picture"])
		c.Set("user_sub", payload.Subject)
		c.Next()
	}
}

func resolveAuthAudience() string {
	if audience := os.Getenv("GOOGLE_OAUTH_AUDIENCE"); audience != "" {
		return audience
	}
	if audience := os.Getenv("GOOGLE_CLIENT_ID"); audience != "" {
		return audience
	}
	return os.Getenv("CLOUD_RUN_SERVICE_URL")
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
