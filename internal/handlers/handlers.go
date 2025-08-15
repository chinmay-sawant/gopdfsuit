package handlers

import (
	"encoding/json"
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

	// API endpoints
	v1 := router.Group("/api/v1")
	v1.POST("/generate/template-pdf", handleGenerateTemplatePDF)
	v1.GET("/template-data", handleGetTemplateData)

	// Template editor endpoint
	router.GET("/editor", PDFEditorHandler)
}

// handlePDFViewer serves the PDF viewer HTML page
func handlePDFViewer(c *gin.Context) {
	c.HTML(http.StatusOK, "pdf_viewer.html", gin.H{
		"title": "GoPdfSuit - PDF Viewer",
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

// PDFEditorHandler serves the template editor interface
func PDFEditorHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "pdf_editor.html", gin.H{
		"title": "GoPdfSuit - Template Editor",
	})
}
