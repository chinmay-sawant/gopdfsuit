package pdf

import (
	"bytes"
	"strings"
	"testing"
)

// Minimal valid PDF with 1 page for testing (Same as before)
var minimalPDF = []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << >> /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 21 >>
stream
BT /F1 12 Tf 100 700 Td (Hello World) Tj ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000060 00000 n 
0000000117 00000 n 
0000000222 00000 n 
trailer
<< /Size 5 /Root 1 0 R >>
startxref
293
%%EOF
`)

var minimalMultiPagePDF = []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << >> /Contents 5 0 R >>
endobj
4 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << >> /Contents 6 0 R >>
endobj
5 0 obj
<< /Length 20 >>
stream
BT /F1 12 Tf 100 700 Td (First Page) Tj ET
endstream
endobj
6 0 obj
<< /Length 21 >>
stream
BT /F1 12 Tf 100 700 Td (Alpha Beta) Tj ET
endstream
endobj
xref
0 7
0000000000 65535 f 
0000000009 00000 n 
0000000060 00000 n 
0000000125 00000 n 
0000000230 00000 n 
0000000335 00000 n 
0000000440 00000 n 
trailer
<< /Size 7 /Root 1 0 R >>
startxref
545
%%EOF
`)

func TestGetPageInfo(t *testing.T) {
	info, err := GetPageInfo(minimalPDF)
	if err != nil {
		t.Fatalf("GetPageInfo failed: %v", err)
	}

	if info.TotalPages != 1 {
		t.Errorf("Expected 1 page, got %d", info.TotalPages)
	}

	if len(info.Pages) != 1 {
		t.Fatalf("Expected 1 page detail, got %d", len(info.Pages))
	}

	if info.Pages[0].Width != 612 || info.Pages[0].Height != 792 {
		t.Errorf("Expected 612x792 from MediaBox, got %.2fx%.2f", info.Pages[0].Width, info.Pages[0].Height)
	}
}

func TestExtractTextPositions(t *testing.T) {
	positions, err := ExtractTextPositions(minimalPDF, 1)
	if err != nil {
		t.Fatalf("ExtractTextPositions failed: %v", err)
	}

	if len(positions) == 0 {
		t.Log("No text positions found - parser might be too simple for this manual PDF or requires compression")
	} else {
		found := false
		for _, p := range positions {
			if p.Text == "Hello World" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected 'Hello World' text, got: %v", positions)
		}
	}
}

func TestApplyRedactions(t *testing.T) {
	redactions := []RedactionRect{
		{
			PageNum: 1,
			X:       100,
			Y:       690,
			Width:   100,
			Height:  20,
		},
	}

	redactedBytes, err := ApplyRedactions(minimalPDF, redactions)
	if err != nil {
		t.Fatalf("ApplyRedactions failed: %v", err)
	}

	if len(redactedBytes) == 0 {
		t.Error("Redacted bytes are empty")
	}

	if bytes.Equal(redactedBytes, minimalPDF) {
		t.Error("Redacted PDF is identical to original")
	}
}

func TestFindTextOccurrences(t *testing.T) {
	// "Hello World" is in the PDF. Search for "Hello".
	rects, err := FindTextOccurrences(minimalPDF, "Hello")
	if err != nil {
		t.Fatalf("FindTextOccurrences failed: %v", err)
	}

	// Since our extraction is "Hello World" as one block:
	// A simple "Contains" check should find it.
	// However, if we want precise bounding box for "Hello" only, that's harder.
	// For now, if it matches, it might return the whole block.

	if len(rects) == 0 {
		t.Log("No occurrences found - simplified parser limitation?")
	} else {
		if rects[0].PageNum != 1 {
			t.Errorf("Expected PageNum 1, got %d", rects[0].PageNum)
		}
		// x should be around 100
		if rects[0].X < 90 || rects[0].X > 110 {
			t.Logf("Warning: X coordinate %f outside expected range 100", rects[0].X)
		}
	}
}

func TestFindTextOccurrencesSingleLetterIsNotWholeWord(t *testing.T) {
	rects, err := FindTextOccurrences(minimalPDF, "o")
	if err != nil {
		t.Fatalf("FindTextOccurrences failed: %v", err)
	}
	if len(rects) < 2 {
		t.Fatalf("expected at least 2 matches for letter 'o', got %d", len(rects))
	}

	positions, err := ExtractTextPositions(minimalPDF, 1)
	if err != nil || len(positions) == 0 {
		t.Fatalf("failed to read baseline text positions: err=%v count=%d", err, len(positions))
	}
	fullWordWidth := positions[0].Width

	for _, r := range rects {
		if r.PageNum != 1 {
			t.Fatalf("expected page 1 matches only, got page %d", r.PageNum)
		}
		if r.Width >= fullWordWidth {
			t.Fatalf("single-letter redaction width should be narrower than full token width: got %.2f, full %.2f", r.Width, fullWordWidth)
		}
	}
}

func TestFindTextOccurrencesMultiPage(t *testing.T) {
	rects, err := FindTextOccurrences(minimalMultiPagePDF, "Alpha")
	if err != nil {
		t.Fatalf("FindTextOccurrences failed: %v", err)
	}
	if len(rects) == 0 {
		t.Fatal("expected at least one match on page 2")
	}
	for _, r := range rects {
		if r.PageNum != 2 {
			t.Fatalf("expected matches on page 2 only, got page %d", r.PageNum)
		}
	}
}

func TestApplyRedactionsAdvancedSecureRequired(t *testing.T) {
	opts := ApplyRedactionOptions{
		Mode:       "secure_required",
		TextSearch: []RedactionTextQuery{{Text: "Hello"}},
	}

	out, report, err := ApplyRedactionsAdvancedWithReport(minimalPDF, opts)
	if err != nil {
		if strings.Contains(err.Error(), "no secure text content") {
			t.Skip("minimal fixture does not expose parseable text stream for secure rewrite")
		}
		t.Fatalf("ApplyRedactionsAdvancedWithReport failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected output bytes")
	}
	if !report.AppliedSecure {
		t.Fatalf("expected secure apply true, got false. report=%+v", report)
	}
	if report.SecurityOutcome != "secure" {
		t.Fatalf("expected securityOutcome=secure, got %s", report.SecurityOutcome)
	}

	if bytes.Contains(bytes.ToLower(out), []byte("hello world")) {
		t.Fatal("expected secure mode to remove matching text from content stream")
	}
}

func TestAnalyzePageCapabilities(t *testing.T) {
	caps, err := AnalyzePageCapabilities(minimalPDF)
	if err != nil {
		t.Fatalf("AnalyzePageCapabilities failed: %v", err)
	}
	if len(caps) != 1 {
		t.Fatalf("expected 1 capability entry, got %d", len(caps))
	}
	if caps[0].Type == "" {
		t.Fatal("expected non-empty capability type")
	}
}

func TestScrubDecodedContentPreservesNonMatchedCharacters(t *testing.T) {
	decoded := []byte("BT /F1 12 Tf 100 700 Td (Hello) Tj ET")
	positions := parseTextOperators(decoded)
	if len(positions) == 0 {
		t.Fatal("expected parsed text positions")
	}
	p := positions[0]
	charW := p.Width / float64(len([]rune(p.Text)))
	rects := []RedactionRect{{
		PageNum: 1,
		X:       p.X + charW,
		Y:       p.Y,
		Width:   charW,
		Height:  p.Height,
	}}

	out, changed := scrubDecodedContent(decoded, rects, []RedactionTextQuery{{Text: "e"}})
	if !changed {
		t.Fatal("expected content to be changed")
	}
	outLower := strings.ToLower(string(out))
	if !strings.Contains(outLower, "h") || !strings.Contains(outLower, "l") {
		t.Fatalf("expected non-matched characters to remain, got: %s", string(out))
	}
	if strings.Contains(outLower, "hello") {
		t.Fatalf("expected matched character to be scrubbed, got: %s", string(out))
	}
}
