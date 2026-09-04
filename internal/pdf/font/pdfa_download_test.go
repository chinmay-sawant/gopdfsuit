package font

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// The Liberation tarball URL is a hardcoded, unauthenticated download whose
// bytes are verified against liberationFontsArchiveSHA256 in downloadFonts.
// Pin the URL here so any change is a deliberate, reviewed diff.
func TestLiberationFontsURLPinned(t *testing.T) {
	const want = "https://github.com/liberationfonts/liberation-fonts/files/7261482/liberation-fonts-ttf-2.1.5.tar.gz"
	if defaultLiberationFontsArchiveURL != want {
		t.Fatalf("liberation fonts URL changed to %q; review for supply-chain safety and update the pin", defaultLiberationFontsArchiveURL)
	}
	const wantSum = "7191c669bf38899f73a2094ed00f7b800553364f90e2637010a69c0e268f25d0"
	if liberationFontsArchiveSHA256 != wantSum {
		t.Fatalf("liberation fonts digest changed to %q; review for supply-chain safety", liberationFontsArchiveSHA256)
	}
}

// TestVerifyLiberationDigestMismatch points the verifier at the default URL
// with garbage bytes and asserts a checksum error (offline: no download).
func TestVerifyLiberationDigestMismatch(t *testing.T) {
	old := liberationFontsArchiveURL
	liberationFontsArchiveURL = defaultLiberationFontsArchiveURL
	defer func() { liberationFontsArchiveURL = old }()

	f, err := os.CreateTemp(t.TempDir(), "archive-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write([]byte("not the pinned release")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := verifyLiberationDigest(f); err == nil {
		t.Fatal("expected checksum mismatch, got nil")
	}
}

// TestVerifyLiberationDigestSkipsOverride asserts httptest overrides skip
// verification (their bytes are fixtures, not the release).
func TestVerifyLiberationDigestSkipsOverride(t *testing.T) {
	old := liberationFontsArchiveURL
	liberationFontsArchiveURL = "http://127.0.0.1/fixture.tar.gz"
	defer func() { liberationFontsArchiveURL = old }()

	f, err := os.CreateTemp(t.TempDir(), "archive-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write([]byte("fixture bytes")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := verifyLiberationDigest(f); err != nil {
		t.Fatalf("override must skip verification: %v", err)
	}
}

func newIsolatedManager(t *testing.T) *PDFAFontManager {
	t.Helper()
	m := &PDFAFontManager{loadedFonts: make(map[string]*TTFFont)}
	if err := m.Initialize(PDFAFontConfig{FontsDirectory: t.TempDir(), AutoDownload: true}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return m
}

func withArchiveServer(t *testing.T, status int, body []byte) func() {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	old := liberationFontsArchiveURL
	liberationFontsArchiveURL = srv.URL
	t.Cleanup(func() { liberationFontsArchiveURL = old })
	return func() {}
}

func TestDownloadFontsServerError(t *testing.T) {
	withArchiveServer(t, http.StatusInternalServerError, []byte("boom"))
	m := newIsolatedManager(t)
	if err := m.downloadFonts(); err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}

func TestDownloadFontsMalformedArchive(t *testing.T) {
	withArchiveServer(t, http.StatusOK, []byte("not a gzip archive at all"))
	m := newIsolatedManager(t)
	if err := m.downloadFonts(); err == nil {
		t.Fatal("expected error for malformed tar.gz, got nil")
	}
}

func buildTinyFontsTar(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	payload := []byte("fake-ttf-bytes")
	if err := tw.WriteHeader(&tar.Header{Name: "liberation-fonts-ttf-2.1.5/LiberationSans-Regular.ttf", Mode: 0o644, Size: int64(len(payload))}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.WriteHeader(&tar.Header{Name: "README.txt", Mode: 0o644, Size: 3}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write([]byte("hi\n")); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func TestDownloadFontsSuccessExtractsKnownTTFOnly(t *testing.T) {
	withArchiveServer(t, http.StatusOK, buildTinyFontsTar(t))
	m := newIsolatedManager(t)
	if err := m.downloadFonts(); err != nil {
		t.Fatalf("downloadFonts: %v", err)
	}
	if _, err := os.Stat(filepath.Join(m.config.FontsDirectory, "LiberationSans-Regular.ttf")); err != nil {
		t.Fatalf("expected extracted font: %v", err)
	}
	if _, err := os.Stat(filepath.Join(m.config.FontsDirectory, "README.txt")); !os.IsNotExist(err) {
		t.Fatalf("non-TTF file must not be extracted, stat err: %v", err)
	}
}
