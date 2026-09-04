package font

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func provisionServer(t *testing.T, body []byte, status int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestFetchToTempOK(t *testing.T) {
	url := provisionServer(t, []byte("font-bytes"), http.StatusOK)
	path, err := FetchToTemp(context.Background(), url, t.TempDir(), "font-*.ttf", 1024, "", 5*time.Second)
	if err != nil {
		t.Fatalf("FetchToTemp: %v", err)
	}
	defer func() { _ = os.Remove(path) }()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp: %v", err)
	}
	if string(got) != "font-bytes" {
		t.Fatalf("unexpected body %q", got)
	}
}

func TestFetchToTempSHA(t *testing.T) {
	body := []byte("pinned-bytes")
	sum := sha256.Sum256(body)
	want := hex.EncodeToString(sum[:])
	url := provisionServer(t, body, http.StatusOK)

	if _, err := FetchToTemp(context.Background(), url, t.TempDir(), "f-*.ttf", 1024, want, 5*time.Second); err != nil {
		t.Fatalf("matching digest must pass: %v", err)
	}
	if _, err := FetchToTemp(context.Background(), url, t.TempDir(), "f-*.ttf", 1024, "00"+want[2:], 5*time.Second); err == nil {
		t.Fatal("mismatched digest must fail")
	}
}

func TestFetchToTempSizeCap(t *testing.T) {
	url := provisionServer(t, []byte("0123456789abcdef"), http.StatusOK)
	dir := t.TempDir()
	if _, err := FetchToTemp(context.Background(), url, dir, "f-*.ttf", 4, "", 5*time.Second); err == nil {
		t.Fatal("oversize body must fail")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		t.Fatalf("failed download left file %s behind", e.Name())
	}
}

func TestFetchToTempHTTPError(t *testing.T) {
	url := provisionServer(t, []byte("boom"), http.StatusInternalServerError)
	if _, err := FetchToTemp(context.Background(), url, t.TempDir(), "f-*.ttf", 1024, "", 5*time.Second); err == nil {
		t.Fatal("HTTP 500 must fail")
	}
}

func TestFetchToTempBadArgs(t *testing.T) {
	if _, err := FetchToTemp(context.Background(), "", t.TempDir(), "f-*.ttf", 1024, "", 5*time.Second); err == nil {
		t.Fatal("empty URL must fail")
	}
	url := provisionServer(t, []byte("x"), http.StatusOK)
	if _, err := FetchToTemp(context.Background(), url, t.TempDir(), "f-*.ttf", 0, "", 5*time.Second); err == nil {
		t.Fatal("non-positive cap must fail")
	}
}
