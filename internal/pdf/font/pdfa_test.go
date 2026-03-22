package font

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// newCachingTestManager creates a PDFAFontManager pointed at a temp directory,
// using an httptest server that always returns the given HTTP status code.
// Returns the manager, the hit counter, and a cleanup function.
func newCachingTestManager(t *testing.T, statusCode int) (*PDFAFontManager, *atomic.Int32) {
	t.Helper()
	tmpDir := t.TempDir()
	var hits atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, http.StatusText(statusCode), statusCode)
	}))
	t.Cleanup(server.Close)

	m := NewPDFAFontManager()
	if err := m.Initialize(PDFAFontConfig{
		FontsDirectory:         filepath.Join(tmpDir, "fonts"),
		FallbackFontsDirectory: filepath.Join(tmpDir, "fallback"),
		AutoDownload:           true,
		ArchiveURL:             server.URL + "/liberation-fonts.tar.gz",
	}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return m, &hits
}

// TestCacheFailError verifies that a 404 from the download server
// causes EnsureFontsAvailable to return a non-nil error.
func TestCacheFailError(t *testing.T) {
	m, _ := newCachingTestManager(t, http.StatusNotFound)

	err := m.EnsureFontsAvailable()
	if err == nil {
		t.Fatal("expected error when download returns 404, got nil")
	}
	if msg := err.Error(); !strings.Contains(strings.ToLower(msg), "font") &&
		!strings.Contains(msg, "404") && !strings.Contains(msg, "download") {
		t.Errorf("error message should reference fonts or download failure; got: %q", msg)
	}
}

// TestCacheFailNoRetry verifies that a failed download is cached so
// subsequent calls do not trigger additional HTTP requests.
func TestCacheFailNoRetry(t *testing.T) {
	m, hits := newCachingTestManager(t, http.StatusNotFound)

	// Prime the cache with a failed attempt.
	_ = m.EnsureFontsAvailable()
	hitsAfterFirst := hits.Load()
	if hitsAfterFirst != 1 {
		t.Fatalf("expected exactly 1 HTTP hit after first call, got %d", hitsAfterFirst)
	}

	// A second call must reuse the cached error and not hit the server again.
	_ = m.EnsureFontsAvailable()
	if got := hits.Load(); got != hitsAfterFirst {
		t.Errorf("download cache broken: HTTP hit count grew from %d to %d on second call", hitsAfterFirst, got)
	}
}

// TestCacheFailStableMsg verifies that successive failures return the
// same error message, confirming the result is served from cache.
func TestCacheFailStableMsg(t *testing.T) {
	m, _ := newCachingTestManager(t, http.StatusNotFound)

	firstErr := m.EnsureFontsAvailable()
	secondErr := m.EnsureFontsAvailable()

	if firstErr == nil || secondErr == nil {
		t.Fatalf("both calls must fail; got first=%v second=%v", firstErr, secondErr)
	}
	if firstErr.Error() != secondErr.Error() {
		t.Errorf("cached error message changed: first=%q second=%q", firstErr.Error(), secondErr.Error())
	}
}

// newOfflineManager creates a PDFAFontManager with AutoDownload disabled,
// using the supplied fontsDir as the primary fonts directory.
func newOfflineManager(t *testing.T, fontsDir string) *PDFAFontManager {
	t.Helper()
	tmpDir := t.TempDir()
	m := NewPDFAFontManager()
	if err := m.Initialize(PDFAFontConfig{
		FontsDirectory:         fontsDir,
		FallbackFontsDirectory: filepath.Join(tmpDir, "fallback"),
		AutoDownload:           false,
	}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return m
}

// TestRetryErrWhenAbsent verifies that EnsureFontsAvailable returns
// an error (mentioning fonts) when the fonts directory does not exist.
func TestRetryErrWhenAbsent(t *testing.T) {
	fontsDir := filepath.Join(t.TempDir(), "fonts") // does not yet exist
	m := newOfflineManager(t, fontsDir)

	err := m.EnsureFontsAvailable()
	if err == nil {
		t.Fatal("expected error when fonts directory is absent, got nil")
	}
	if msg := err.Error(); !strings.Contains(strings.ToLower(msg), "font") {
		t.Errorf("error should mention fonts; got: %q", msg)
	}
}

// TestRetryOKWithFonts verifies that EnsureFontsAvailable
// succeeds once the required font file is written to the configured directory.
func TestRetryOKWithFonts(t *testing.T) {
	fontsDir := filepath.Join(t.TempDir(), "fonts")
	m := newOfflineManager(t, fontsDir)

	// Ensure initial failure so we know the recovery path is exercised.
	if err := m.EnsureFontsAvailable(); err == nil {
		t.Fatal("expected initial failure before fonts are placed")
	}

	// Write the font file that the manager checks for.
	if err := os.MkdirAll(fontsDir, 0o750); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	fontPath := filepath.Join(fontsDir, "LiberationSans-Regular.ttf")
	fontData := []byte("\x00\x01\x00\x00") // minimal SFNT/TTF magic bytes
	if err := os.WriteFile(fontPath, fontData, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := m.EnsureFontsAvailable(); err != nil {
		t.Fatalf("EnsureFontsAvailable should succeed after fonts are placed, got: %v", err)
	}
}
