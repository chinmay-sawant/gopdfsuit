package pdfobj

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"strings"
	"testing"
)

func minimalPDF(objects string) []byte {
	return []byte("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n" + objects +
		"trailer\n<< /Size 50 /Root 1 0 R >>\nstartxref\n0\n%%EOF\n")
}

const twoObjs = `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] >>
endobj
`

// TestParseObjectMapTable proves the seam builds identical object maps for
// the xref-table variant and for files carrying extra whitespace, generation
// bumps, and benign stream content.
func TestParseObjectMapTable(t *testing.T) {
	cases := map[string]struct {
		pdf      []byte
		wantObjs []int
		wantRoot int
	}{
		"xref-table": {
			pdf:      minimalPDF(twoObjs),
			wantObjs: []int{1, 2, 3},
			wantRoot: 1,
		},
		"extra-whitespace": {
			pdf:      minimalPDF("\n\n" + twoObjs + "\n"),
			wantObjs: []int{1, 2, 3},
			wantRoot: 1,
		},
		"bumped-generation": {
			pdf: minimalPDF(`1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 1 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] >>
endobj
`),
			wantObjs: []int{1, 2, 3},
			wantRoot: 1,
		},
		"stream-with-endobj-text": {
			pdf: minimalPDF(twoObjs + `4 0 obj
<< /Length 44 >>
stream
BT (endobj 1 0 R xref) Tj ET
endstream
endobj
`),
			wantObjs: []int{1, 2, 3, 4},
			wantRoot: 1,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := Parse(tc.pdf)
			for _, num := range tc.wantObjs {
				body, ok := p.Body(num)
				if !ok || len(body) == 0 {
					t.Fatalf("object %d missing or empty", num)
				}
			}
			if !p.HasRoot || p.RootNum != tc.wantRoot {
				t.Fatalf("root = %d,%v, want %d,true", p.RootNum, p.HasRoot, tc.wantRoot)
			}
			if p.Version != "1.4" {
				t.Fatalf("version = %q, want 1.4", p.Version)
			}
			if len(p.Objects) != len(tc.wantObjs) {
				t.Fatalf("object count = %d, want %d", len(p.Objects), len(tc.wantObjs))
			}
		})
	}
}

// TestParseXRefStreamVariant proves xref-stream augmentation recovers an
// object that the plain boundary scan cannot see: the xref stream's type-1
// entry points at an offset where a real object lives.
func TestParseXRefStreamVariant(t *testing.T) {
	hidden := `9 0 obj
<< /FT /Tx /T (Hidden) /V (found) >>
endobj
`
	// Layout: header, then the xref-stream object recording offset of obj 9.
	prefix := "%PDF-1.7\n"
	xrefDict := "8 0 obj\n<< /Type /XRef /W[1 1 2] /Index[0 1] /Length 0 >>\nstream\n"
	// One entry: f1=1 (offset), f2=0, f3=offset-of-obj-9 (2 bytes).
	hiddenOff := len(prefix) + len(xrefDict)
	_ = hiddenOff
	suffix := "\nendstream\nendobj\n" + hidden + "trailer\n<< /Size 10 /Root 1 0 R >>\nstartxref\n0\n%%EOF\n"

	// Build with a placeholder, then patch the real offset into the entry.
	var buf bytes.Buffer
	buf.WriteString(prefix)
	entryPos := len(prefix) + len("8 0 obj\n<< /Type /XRef /W[1 1 2] /Index[0 1] /Length 0 >>\nstream\n")
	buf.WriteString("8 0 obj\n<< /Type /XRef /W[1 1 2] /Index[0 1] /Length 0 >>\nstream\n")
	buf.WriteString("\x00\x00\x00\x00") // placeholder entry
	buf.WriteString(suffix)
	raw := buf.Bytes()
	realOff := bytes.Index(raw, []byte("9 0 obj"))
	if realOff < 0 {
		t.Fatal("fixture broken: hidden object not found")
	}
	raw[entryPos] = 1
	raw[entryPos+1] = 0
	raw[entryPos+2] = byte(realOff >> 8)
	raw[entryPos+3] = byte(realOff)

	p := Parse(raw)
	body, ok := p.Body(9)
	if !ok {
		t.Fatalf("xref-stream augmentation missed object 9 (have %v)", p.Objects)
	}
	if !bytes.Contains(body, []byte("(Hidden)")) {
		t.Fatalf("object 9 body wrong: %q", body)
	}
}

// TestAugmentRejectsBadWidthsTable is the hang/panic regression test at the
// seam level: hostile /W arrays must be skipped, never slice out of range.
func TestAugmentRejectsBadWidthsTable(_ *testing.T) {
	for _, w := range []string{"/W[0 0 0]", "/W[1 -2 2]", "/W[99 99 99]"} {
		data := []byte("1 0 obj\n<< " + w + " /Index[0 1] /Length 4 >>\nstream\nabcd\nendstream\nendobj\n")
		m := map[int][]byte{}
		g := map[int]int{}
		AugmentRawIntMap(data, m, g) // must return, not hang or panic
		sm := map[string][]byte{}
		AugmentStrMap(data, sm)
	}
}

// TestInflateCappedTable covers the capped-inflation seam: zlib and raw
// flate round-trips, invalid input errors, and zip-bomb rejection.
func TestInflateCappedTable(t *testing.T) {
	flateOf := func(b []byte) []byte {
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flate.BestCompression)
		if _, err := w.Write(b); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	zlibOf := func(b []byte) []byte {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		if _, err := w.Write(b); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	t.Run("zlib-round-trip", func(t *testing.T) {
		out, err := InflateCapped(zlibOf([]byte("hello seam")), false)
		if err != nil || string(out) != "hello seam" {
			t.Fatalf("out=%q err=%v", out, err)
		}
	})
	t.Run("flate-round-trip", func(t *testing.T) {
		out, err := InflateCapped(flateOf([]byte("hello flate")), true)
		if err != nil || string(out) != "hello flate" {
			t.Fatalf("out=%q err=%v", out, err)
		}
	})
	t.Run("invalid-errors", func(t *testing.T) {
		if _, err := InflateCapped([]byte("not compressed"), false); err == nil {
			t.Fatal("expected error for invalid zlib")
		}
	})
	t.Run("bomb-rejected", func(t *testing.T) {
		var bomb bytes.Buffer
		w, _ := flate.NewWriter(&bomb, flate.BestCompression)
		chunk := bytes.Repeat([]byte("A"), 1<<20)
		for i := 0; i < MaxInflateBytes/(1<<20)+2; i++ {
			if _, err := w.Write(chunk); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := InflateCapped(bomb.Bytes(), true); err == nil {
			t.Fatal("expected bomb rejection")
		} else if !strings.Contains(err.Error(), "exceeds") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("decompress-any-fallback", func(t *testing.T) {
		out, err := DecompressAny(flateOf([]byte("fallback ok")))
		if err != nil || string(out) != "fallback ok" {
			t.Fatalf("out=%q err=%v", out, err)
		}
	})
}

// TestHasEncryptEntryIgnoresStreamText proves keyword scans skip stream
// bytes at the seam level.
func TestHasEncryptEntryIgnoresStreamText(t *testing.T) {
	plain := minimalPDF(twoObjs)
	if HasEncryptEntry(plain) {
		t.Fatal("plain PDF detected as encrypted")
	}
	streamPDF := minimalPDF(twoObjs + `4 0 obj
<< /Length 30 >>
stream
BT (/Encrypt here) Tj ET
endstream
endobj
`)
	if HasEncryptEntry(streamPDF) {
		t.Fatal("/Encrypt inside a stream falsely detected")
	}
	enc := minimalPDF(twoObjs)
	enc = bytes.Replace(enc, []byte("/Root 1 0 R"), []byte("/Root 1 0 R /Encrypt 9 0 R"), 1)
	if !HasEncryptEntry(enc) {
		t.Fatal("real /Encrypt entry missed")
	}
}
