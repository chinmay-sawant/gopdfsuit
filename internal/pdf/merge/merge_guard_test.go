package merge

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func minimalPDF(objects string) []byte {
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	b.WriteString(objects)
	b.WriteString("trailer\n<< /Size 50 /Root 1 0 R >>\nstartxref\n0\n%%EOF\n")
	return b.Bytes()
}

const singlePageObjs = `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 40 >>
stream
BT (see 99 0 R and 3 0 R) Tj ET
endstream
endobj
`

// TestMergeRejectsSparseObjects proves a sparse huge object number is
// rejected fast with an error instead of blowing time/memory.
func TestMergeRejectsSparseObjects(t *testing.T) {
	pdf := minimalPDF(singlePageObjs + "999999 0 obj\n<< /Type /Example >>\nendobj\n")
	start := time.Now()
	_, err := MergePDFs([][]byte{pdf})
	if err == nil || !strings.Contains(err.Error(), "objects") {
		t.Fatalf("expected object-cap error, got %v", err)
	}
	if time.Since(start) > 10*time.Second {
		t.Fatalf("sparse-object rejection took too long")
	}
}

// TestReplaceRefsSkipsLiterals proves text like "(see 1 0 R)" inside a
// literal string survives remapping while real refs are still remapped.
func TestReplaceRefsSkipsLiterals(t *testing.T) {
	body := []byte("<< /Title (see 1 0 R) /Parent 5 0 R /Hex <31203052> >>")
	out := ReplaceRefsOutsideStreams(body, 10)
	if !bytes.Contains(out, []byte("(see 1 0 R)")) {
		t.Fatalf("literal string was rewritten: %q", out)
	}
	if !bytes.Contains(out, []byte("15 0 R")) {
		t.Fatalf("real ref was not remapped: %q", out)
	}
	if !bytes.Contains(out, []byte("<31203052>")) {
		t.Fatalf("hex string was rewritten: %q", out)
	}
}

// TestHasEncryptIgnoresStreamText proves "/Encrypt" text inside a content
// stream no longer triggers a false encrypted-PDF rejection.
func TestHasEncryptIgnoresStreamText(t *testing.T) {
	pdf := minimalPDF(singlePageObjs)
	if hasEncrypt(pdf) {
		t.Fatalf("plain PDF falsely detected as encrypted")
	}
	withText := minimalPDF(strings.Replace(singlePageObjs,
		"(see 99 0 R and 3 0 R)", "(see /Encrypt here)", 1))
	if hasEncrypt(withText) {
		t.Fatalf("/Encrypt inside a content stream falsely detected as encryption")
	}
	encrypted := minimalPDF(singlePageObjs)
	encrypted = bytes.Replace(encrypted,
		[]byte("/Root 1 0 R"),
		[]byte("/Root 1 0 R /Encrypt 9 0 R"), 1)
	if !hasEncrypt(encrypted) {
		t.Fatalf("real /Encrypt trailer entry was not detected")
	}
}

const twoPageObjs = `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R 6 0 R] /Count 2 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 40 >>
stream
BT (see 6 0 R and 7 0 R) Tj ET
endstream
endobj
6 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] /Contents 7 0 R /PieceInfo (page2marker) >>
endobj
7 0 obj
<< /Length 10 >>
stream
page two
endstream
endobj
`

// TestSplitIgnoresStreamRefs proves ref-like text inside stream bytes does
// not pull phantom objects (here: page 2) into a page-1-only split, and the
// literal itself survives untouched.
func TestSplitIgnoresStreamRefs(t *testing.T) {
	pdf := minimalPDF(twoPageObjs)
	parts, err := SplitPDF(pdf, SplitSpec{Pages: []int{1}})
	if err != nil {
		t.Fatalf("SplitPDF failed: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	if bytes.Contains(parts[0], []byte("page2marker")) {
		t.Fatalf("page-2 object pulled into page-1 split via stream text")
	}
	if !bytes.Contains(parts[0], []byte("(see 6 0 R and 7 0 R)")) {
		t.Fatalf("stream literal was altered by split")
	}
}

// TestDictPartScansDictOnly is a focused unit test for the split scanner.
func TestDictPartScansDictOnly(t *testing.T) {
	body := []byte("<< /Contents 4 0 R >>\nstream\n99 0 R\nendstream")
	if got := string(dictPart(body)); strings.Contains(got, "99 0 R") {
		t.Fatalf("dictPart leaked stream bytes: %q", got)
	}
	if !strings.Contains(string(dictPart(body)), "4 0 R") {
		t.Fatalf("dictPart dropped dict refs")
	}
}

func TestMergeAndSplitPreservePageTreeOrderAndInheritance(t *testing.T) {
	const objects = `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [6 0 R 3 0 R] /Count 2 /MediaBox [0 0 200 300] /Resources 9 0 R >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /Contents 4 0 R /PieceInfo (first-in-serialization) >>
endobj
4 0 obj
<< /Length 0 >>
stream
endstream
endobj
6 0 obj
<< /Type /Page /Parent 2 0 R /Contents 7 0 R /PieceInfo (first-in-tree) >>
endobj
7 0 obj
<< /Length 0 >>
stream
endstream
endobj
9 0 obj
<< /ProcSet [/PDF] >>
endobj
`
	pdf := minimalPDF(objects)

	merged, err := MergePDFs([][]byte{pdf})
	if err != nil {
		t.Fatalf("MergePDFs failed: %v", err)
	}
	assertOrderedPagesAndInheritance(t, merged)

	parts, err := SplitPDF(pdf, SplitSpec{})
	if err != nil {
		t.Fatalf("SplitPDF failed: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("expected one split output, got %d", len(parts))
	}
	assertOrderedPagesAndInheritance(t, parts[0])
}

func assertOrderedPagesAndInheritance(t *testing.T, pdf []byte) {
	t.Helper()
	if !bytes.Contains(pdf, []byte("/Kids [8 0 R 5 0 R]")) {
		t.Fatalf("page tree order changed: %s", pdf)
	}
	if !bytes.Contains(pdf, []byte("/MediaBox [0 0 200 300]")) {
		t.Fatalf("inherited MediaBox was not materialized: %s", pdf)
	}
	if !bytes.Contains(pdf, []byte("/Resources 11 0 R")) {
		t.Fatalf("inherited Resources were not materialized: %s", pdf)
	}
}

func TestMergeAndSplitRejectPageTreeCycles(t *testing.T) {
	tests := []struct {
		name    string
		objects string
	}{
		{
			name: "self cycle",
			objects: `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [2 0 R] /Count 1 >>
endobj
`,
		},
		{
			name: "two object cycle",
			objects: `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Pages /Kids [2 0 R] /Count 1 >>
endobj
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdf := minimalPDF(tt.objects)
			if _, err := MergePDFs([][]byte{pdf}); err == nil || !strings.Contains(err.Error(), "cycle") {
				t.Fatalf("MergePDFs error = %v, want cycle error", err)
			}
			if _, err := SplitPDF(pdf, SplitSpec{}); err == nil || !strings.Contains(err.Error(), "cycle") {
				t.Fatalf("SplitPDF error = %v, want cycle error", err)
			}
		})
	}
}

func TestMergeRejectsExcessivePageTreeDepth(t *testing.T) {
	const depth = 1100
	var objects strings.Builder
	objects.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	for num := 2; num < depth+2; num++ {
		fmt.Fprintf(&objects, "%d 0 obj\n<< /Type /Pages /Kids [%d 0 R] /Count 1 >>\nendobj\n", num, num+1)
	}
	fmt.Fprintf(&objects, "%d 0 obj\n<< /Type /Page /MediaBox [0 0 10 10] >>\nendobj\n", depth+2)

	_, err := MergePDFs([][]byte{minimalPDF(objects.String())})
	if err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("MergePDFs error = %v, want depth error", err)
	}
}
