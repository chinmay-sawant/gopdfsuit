package compress

import (
	"bytes"
	"compress/zlib"
	"testing"
)

func TestRejectsHighObjectNumber(t *testing.T) {
	pdf := []byte("%PDF-1.4\n50001 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer << /Root 50001 0 R >>\n%%EOF\n")
	_, err := CompressPDF(pdf, Options{})
	if err == nil {
		t.Fatal("expected error for object number above MaxObjects")
	}
}

func TestInflateBombCapped(t *testing.T) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	zeros := make([]byte, MaxInflateBytes+1024)
	if _, err := w.Write(zeros); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := decompressFlate(buf.Bytes())
	if err == nil {
		t.Fatal("expected inflate cap")
	}
}

func TestImagePixelBudget(t *testing.T) {
	if imagePixelBudgetOK(0, 10) {
		t.Fatal("zero width should fail")
	}
	if imagePixelBudgetOK(100000, 100000) {
		t.Fatal("100k² should fail")
	}
	if !imagePixelBudgetOK(100, 100) {
		t.Fatal("100x100 should pass")
	}
}

func TestCapsMaxImageDim(t *testing.T) {
	o := Options{MaxImageDim: 99_999_999}.withDefaults()
	if o.MaxImageDim != maxAllowedDim {
		t.Fatalf("MaxImageDim=%d want %d", o.MaxImageDim, maxAllowedDim)
	}
}
