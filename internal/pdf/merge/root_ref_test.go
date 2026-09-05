package merge

import (
	"bytes"
	"fmt"
	"testing"
)

// craftedRootTrapPDF returns a minimal 1-page PDF whose content stream draws
// the literal text "(/Root 99 0 R)" BEFORE the trailer's "/Root 1 0 R".
// A first-match /Root scan honors the decoy (object 99 does not exist) and
// silently extracts zero pages for the file.
func craftedRootTrapPDF() []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.7\n")
	offsets := map[int]int{}
	emit := func(num int, body string) {
		offsets[num] = buf.Len()
		buf.WriteString(body)
	}
	emit(1, "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	emit(2, "2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	emit(3, "3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n")
	emit(4, "4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")
	emit(5, "5 0 obj\n<< /Length 44 >>\nstream\nBT /F1 12 Tf 72 720 Td (/Root 99 0 R) Tj ET\nendstream\nendobj\n")
	xrefPos := buf.Len()
	buf.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n")
	buf.WriteString(fmt.Sprintf("%d\n%%%%EOF\n", xrefPos))
	return buf.Bytes()
}

// countOutputPages counts /Type /Page occurrences that are not /Type /Pages.
func countOutputPages(pdf []byte) int {
	const needle = "/Type /Page"
	n := 0
	for i := 0; i+len(needle) <= len(pdf); {
		j := bytes.Index(pdf[i:], []byte(needle))
		if j < 0 {
			break
		}
		i += j + len(needle)
		if i < len(pdf) && pdf[i] == 's' {
			continue
		}
		n++
	}
	return n
}

func TestMergeKeepsPagesWhenDecoyRootPrecedesTrailer(t *testing.T) {
	doc := craftedRootTrapPDF()
	fc, err := parseFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	if got := findRootRef(doc, fc.Objects); got != "1 0" {
		t.Fatalf("findRootRef picked decoy: %q, want %q", got, "1 0")
	}
	merged, err := MergePDFs([][]byte{doc, doc})
	if err != nil {
		t.Fatal(err)
	}
	if n := countOutputPages(merged); n != 2 {
		t.Fatalf("merged page count = %d, want 2 (a file lost its pages)", n)
	}
}
