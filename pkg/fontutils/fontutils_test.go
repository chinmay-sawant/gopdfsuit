package fontutils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testFont(name, url string) MathFontInfo {
	return MathFontInfo{Name: name, FileName: name + ".ttf", DownloadURL: url}
}

// TestDownloadFontRejectsDigestMismatch serves valid-sized bytes that do not
// match the pinned SHA256 and asserts an error plus no cached file.
func TestDownloadFontRejectsDigestMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("valid-size but wrong content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_FONTS_DIR", dir)

	font := testFont("Pinned", srv.URL)
	font.SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
	err := downloadFontContext(context.Background(), font)
	if err == nil {
		t.Fatal("expected checksum error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "Pinned.ttf")); !os.IsNotExist(statErr) {
		t.Fatal("mismatched font must not be cached")
	}
}

// TestDownloadFontAcceptsMatchingDigest serves bytes whose digest equals the
// pin and asserts the file lands in the fonts dir.
func TestDownloadFontAcceptsMatchingDigest(t *testing.T) {
	body := []byte("exactly these bytes")
	sum := sha256.Sum256(body)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_FONTS_DIR", dir)

	font := testFont("Pinned", srv.URL)
	font.SHA256 = hex.EncodeToString(sum[:])
	if err := downloadFontContext(context.Background(), font); err != nil {
		t.Fatalf("matching digest must download: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Pinned.ttf")); err != nil {
		t.Fatalf("expected cached font: %v", err)
	}
}

// TestDownloadFontRejectsTruncation serves more than the cap and asserts an
// error plus no cached file and no leftover temp file.
func TestDownloadFontRejectsTruncation(t *testing.T) {
	oldMax := maxFontDownloadSize
	maxFontDownloadSize = 1024
	defer func() { maxFontDownloadSize = oldMax }()

	big := make([]byte, 2048)
	for i := range big {
		big[i] = byte(i)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(big)
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_FONTS_DIR", dir)

	err := downloadFontContext(context.Background(), testFont("Big", srv.URL))
	if err == nil {
		t.Fatal("expected truncation error")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "Big.ttf")); !os.IsNotExist(statErr) {
		t.Fatal("truncated font must not be cached")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("temp file leaked: %v", entries)
	}
}

// TestEnsureMathFontsContextSurfacesErrors asserts download failures (HTTP
// 500 here) are collected and returned instead of only logged.
func TestEnsureMathFontsContextSurfacesErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_FONTS_DIR", dir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errs := ensureMathFonts(ctx, []MathFontInfo{testFont("Missing", srv.URL)})
	if len(errs) != 1 {
		t.Fatalf("expected 1 download error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "HTTP 500") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

// TestEnsureMathFontsContextRespectsCancel asserts a cancelled context aborts
// the download instead of hanging.
func TestEnsureMathFontsContextRespectsCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Setenv("GOPDFSUIT_FONTS_DIR", dir)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	errs := ensureMathFonts(ctx, []MathFontInfo{testFont("Slow", srv.URL)})
	if len(errs) != 1 {
		t.Fatalf("expected 1 context error, got %d", len(errs))
	}
}
