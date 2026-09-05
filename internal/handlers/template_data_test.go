package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func writeTemplateFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func getWithFile(t *testing.T, file string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/template-data?file="+file, nil)
	handleGetTemplateData(c)
	return w
}

func TestTemplateDataRereadsOnModify(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_ROOT", dir)
	clearTemplateDataCache()

	writeTemplateFile(t, dir, "a.json", `{"pages":[]}`)
	w1 := getWithFile(t, "a.json")
	if w1.Code != http.StatusOK {
		t.Fatalf("first read code = %d", w1.Code)
	}

	// Ensure mtime advances, then rewrite with different payload.
	writeTemplateFile(t, dir, "a.json", `{"pages":[{"elements":[]}]}`)
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(dir, "a.json"), future, future); err != nil {
		t.Fatal(err)
	}
	w2 := getWithFile(t, "a.json")
	if w2.Code != http.StatusOK {
		t.Fatalf("second read code = %d", w2.Code)
	}
	if w1.Body.String() == w2.Body.String() {
		t.Fatal("expected re-read after modify, got identical cached body")
	}
}

func TestTemplateDataBoundedEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_ROOT", dir)
	clearTemplateDataCache()
	before := TemplateDataEvictionCount()

	for i := 0; i < maxTemplateDataEntries+5; i++ {
		name := filepath.Base(filepath.Join(dir, string(rune('a'+i%26))+string(rune('0'+i/26))+".json"))
		writeTemplateFile(t, dir, name, `{"pages":[]}`)
		w := getWithFile(t, name)
		if w.Code != http.StatusOK {
			t.Fatalf("read %s code = %d", name, w.Code)
		}
	}

	templateDataMu.Lock()
	n := len(templateDataEntries)
	templateDataMu.Unlock()
	if n > maxTemplateDataEntries {
		t.Fatalf("entries = %d, max = %d", n, maxTemplateDataEntries)
	}
	if TemplateDataEvictionCount() <= before {
		t.Fatal("expected eviction counter to advance on overflow")
	}
}

func TestTemplateDataConcurrentReadStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_ROOT", dir)
	clearTemplateDataCache()
	writeTemplateFile(t, dir, "c.json", `{"pages":[]}`)

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				w := getWithFile(t, "c.json")
				if w.Code != http.StatusOK {
					t.Errorf("concurrent read code = %d", w.Code)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestTemplateDataStoreSkipsOversizedEntry(t *testing.T) {
	clearTemplateDataCache()
	path := filepath.Join(t.TempDir(), "large.json")
	if err := os.WriteFile(path, []byte(`{"pages":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	templateDataStore(path, make([]byte, maxTemplateDataBytes+1), fi)
	templateDataMu.Lock()
	defer templateDataMu.Unlock()
	if len(templateDataEntries) != 0 || templateDataBytes != 0 {
		t.Fatalf("oversized entry changed cache: entries=%d bytes=%d", len(templateDataEntries), templateDataBytes)
	}
}
