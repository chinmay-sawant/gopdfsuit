package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/gin-gonic/gin"
)

var templateDataCache sync.Map

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
