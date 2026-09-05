package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/gin-gonic/gin"
)

const (
	maxTemplateDataEntries = 64
	maxTemplateDataBytes   = 16 * 1024 * 1024
)

type templateDataEntry struct {
	data    []byte
	modTime time.Time
	size    int64
}

var (
	templateDataMu      sync.Mutex
	templateDataEntries = make(map[string]*templateDataEntry)
	templateDataBytes   int64
	// TemplateDataEvictions counts bounded-cache evictions (clear-all on
	// overflow). Exported via accessor for B4 metrics wiring.
	templateDataEvictions atomic.Int64
)

// TemplateDataEvictionCount reports bounded-cache evictions.
func TemplateDataEvictionCount() int64 {
	return templateDataEvictions.Load()
}

// templateDataLookup returns a copy of the cached payload when the on-disk
// file is unchanged (same mtime and size). Caller must hold no lock;
// stat is done before acquiring the mutex to keep the hot path short.
func templateDataLookup(filePath string, fi os.FileInfo) ([]byte, bool) {
	templateDataMu.Lock()
	defer templateDataMu.Unlock()
	e, ok := templateDataEntries[filePath]
	if !ok {
		return nil, false
	}
	if !e.modTime.Equal(fi.ModTime()) || e.size != fi.Size() {
		delete(templateDataEntries, filePath)
		templateDataBytes -= int64(len(e.data))
		return nil, false
	}
	return e.data, true
}

func templateDataStore(filePath string, data []byte, fi os.FileInfo) {
	if len(data) > maxTemplateDataBytes {
		return
	}
	templateDataMu.Lock()
	defer templateDataMu.Unlock()
	if _, ok := templateDataEntries[filePath]; ok {
		return
	}
	if len(templateDataEntries) >= maxTemplateDataEntries ||
		templateDataBytes+int64(len(data)) > maxTemplateDataBytes {
		clear(templateDataEntries)
		templateDataBytes = 0
		templateDataEvictions.Add(1)
	}
	templateDataEntries[filePath] = &templateDataEntry{
		data:    data,
		modTime: fi.ModTime(),
		size:    fi.Size(),
	}
	templateDataBytes += int64(len(data))
}

// clearTemplateDataCache drops all entries (tests / memory pressure).
func clearTemplateDataCache() {
	templateDataMu.Lock()
	defer templateDataMu.Unlock()
	clear(templateDataEntries)
	templateDataBytes = 0
}

// handleGetTemplateData serves JSON template data based on file query parameter
func handleGetTemplateData(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		abortError(c, http.StatusBadRequest, "Missing 'file' query parameter")
		return
	}

	// Security: only allow specific file extensions and prevent path traversal
	if filepath.Ext(filename) != ".json" {
		abortError(c, http.StatusBadRequest, "Only JSON files are allowed")
		return
	}

	// Clean the filename to prevent path traversal and resolve against
	// project root so files at repository root are found when running the
	// server from cmd/gopdfsuit.
	filename = filepath.Base(filename)
	filePath := filepath.Clean(filepath.Join(getProjectRoot(), filename))
	if !strings.HasPrefix(filePath, filepath.Clean(getProjectRoot())) {
		abortError(c, http.StatusBadRequest, "Invalid file path")
		return
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		abortError(c, http.StatusNotFound, "Template file not found: "+filename)
		return
	}

	if cached, ok := templateDataLookup(filePath, fi); ok {
		c.Header("Content-Type", "application/json")
		c.Data(http.StatusOK, "application/json", cached)
		return
	}

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		abortError(c, http.StatusNotFound, "Template file not found: "+filename)
		return
	}

	// Validate JSON structure using sonic for performance
	var template models.PDFTemplate
	if err := sonic.Unmarshal(data, &template); err != nil {
		log.Printf("handleGetTemplateData: invalid template %s: %v", filename, err)
		abortError(c, http.StatusBadRequest, "invalid template data")
		return
	}

	if fi2, err := os.Stat(filePath); err == nil {
		fi = fi2
	}
	templateDataStore(filePath, data, fi)

	// Return the JSON data
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}
