package merge

import (
	"bytes"
	"compress/flate"
	"strings"
	"testing"
)

func TestInflateCappedRejectsBomb(t *testing.T) {
	// 100 bytes compress to a tiny flate stream that expands past the cap.
	var bomb bytes.Buffer
	w, _ := flate.NewWriter(&bomb, flate.BestCompression)
	chunk := bytes.Repeat([]byte("A"), 1<<20)
	for i := 0; i < MaxInflateBytes/(1<<20)+2; i++ {
		if _, err := w.Write(chunk); err != nil {
			t.Fatalf("write bomb: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close bomb: %v", err)
	}
	t.Logf("bomb: %d compressed bytes", bomb.Len())

	if _, err := InflateCapped(bomb.Bytes(), true); err == nil {
		t.Fatal("expected bomb rejection, got nil error")
	} else if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInflateCappedRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.BestCompression)
	if _, err := w.Write([]byte("hello flate")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := InflateCapped(buf.Bytes(), true)
	if err != nil {
		t.Fatalf("valid flate rejected: %v", err)
	}
	if string(out) != "hello flate" {
		t.Fatalf("round trip mismatch: %q", out)
	}
}
