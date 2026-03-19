package font

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestFontCacheFail(t *testing.T) {
	tmpDir := t.TempDir()
	var hits atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()

	oldURL := liberationFontsArchiveURL
	liberationFontsArchiveURL = server.URL + "/liberation-fonts.tar.gz"
	t.Cleanup(func() {
		liberationFontsArchiveURL = oldURL
	})

	manager := &PDFAFontManager{
		loadedFonts: make(map[string]*TTFFont),
	}

	err := manager.Initialize(PDFAFontConfig{
		FontsDirectory:         filepath.Join(tmpDir, "fonts"),
		FallbackFontsDirectory: filepath.Join(tmpDir, "missing-system-fonts"),
		AutoDownload:           true,
	})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = manager.EnsureFontsAvailable()
	if err == nil {
		t.Fatal("expected first EnsureFontsAvailable call to fail")
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("expected one download attempt after first call, got %d", got)
	}

	err = manager.EnsureFontsAvailable()
	if err == nil {
		t.Fatal("expected second EnsureFontsAvailable call to return cached failure")
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("expected cached failure to avoid re-download, got %d attempts", got)
	}

	if manager.lastEnsureErr == nil {
		t.Fatal("expected cached ensure error to be stored")
	}
	if want := "failed to download fonts: HTTP 404"; manager.lastEnsureErr.Error() != want {
		t.Fatalf("expected cached error %q, got %q", want, manager.lastEnsureErr.Error())
	}
}

func TestFontCacheRetry(t *testing.T) {
	tmpDir := t.TempDir()
	fontsDir := filepath.Join(tmpDir, "fonts")

	manager := &PDFAFontManager{
		loadedFonts: make(map[string]*TTFFont),
	}

	err := manager.Initialize(PDFAFontConfig{
		FontsDirectory:         fontsDir,
		FallbackFontsDirectory: filepath.Join(tmpDir, "missing-system-fonts"),
		AutoDownload:           false,
	})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if err := manager.EnsureFontsAvailable(); err == nil {
		t.Fatal("expected missing fonts to fail")
	}

	if err := os.MkdirAll(fontsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	fontPath := filepath.Join(fontsDir, "LiberationSans-Regular.ttf")
	if err := os.WriteFile(fontPath, []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := manager.EnsureFontsAvailable(); err != nil {
		t.Fatalf("expected EnsureFontsAvailable to succeed after font appears, got %v", err)
	}
	if manager.lastEnsureErr != nil {
		t.Fatalf("expected cached error to clear after success, got %v", manager.lastEnsureErr)
	}
	if !manager.ensureAttempted {
		t.Fatal("expected ensureAttempted to remain true after the initial failure")
	}
}
