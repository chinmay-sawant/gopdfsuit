package font

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DefaultFetchTimeout bounds a single font-asset download so a network hang
// cannot block PDF generation forever.
const DefaultFetchTimeout = 60 * time.Second

// HTTPStatusError reports a non-200 response. Callers that pin legacy
// message formats can errors.As-match it to recover the status code.
type HTTPStatusError struct {
	URL    string
	Status int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("download %s: HTTP %d", e.URL, e.Status)
}

// FetchToTemp downloads url into a temp file inside dir (os.TempDir when
// empty) and returns its path. The body is capped at maxBytes: a source
// larger than the cap fails instead of caching a truncated asset. When
// wantSHA256 is non-empty the bytes must match that hex digest before the
// path is returned. The caller owns the file: rename it into place or
// remove it.
func FetchToTemp(ctx context.Context, url, dir, pattern string, maxBytes int64, wantSHA256 string, timeout time.Duration) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty download URL")
	}
	if maxBytes <= 0 {
		return "", fmt.Errorf("maxBytes must be positive")
	}
	if timeout <= 0 {
		timeout = DefaultFetchTimeout
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("download request: %w", err)
	}
	resp, err := client.Do(req) //nolint:gosec // URLs are pinned constants or test fixtures, not user input
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", &HTTPStatusError{URL: url, Status: resp.StatusCode}
	}

	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Read one byte past the cap so truncation is detectable: a full
	// maxBytes+1 means the source was larger than allowed.
	n, err := io.Copy(tmpFile, io.LimitReader(resp.Body, maxBytes+1))
	if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("write download file: %w", err)
	}
	if n > maxBytes {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("download %s exceeds maximum size (%d bytes)", url, maxBytes)
	}

	if wantSHA256 != "" {
		if err := VerifyFileSHA256(tmpPath, wantSHA256); err != nil {
			_ = os.Remove(tmpPath)
			return "", err
		}
	}

	return tmpPath, nil
}

// VerifyFileSHA256 rejects the file at path unless its hex digest equals
// want. A mismatch means upstream re-released the asset or the transfer was
// tampered with; the error names both digests so the pin can be refreshed.
func VerifyFileSHA256(path, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("checksum %s: %w", path, err)
	}
	defer func() { _ = f.Close }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("checksum %s: %w", path, err)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != want {
		return fmt.Errorf("checksum mismatch for %s (got %s, want %s): refresh the SHA256 pin", path, got, want)
	}
	return nil
}
