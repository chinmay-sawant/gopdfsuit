package gopdflib_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// TestWrapperBoundaryValidation asserts every public wrapper rejects
// obviously-invalid input at the boundary without reaching the engine.
func TestWrapperBoundaryValidation(t *testing.T) {
	t.Run("MergePDFs", func(t *testing.T) {
		cases := []struct {
			name  string
			files [][]byte
		}{
			{"nil", nil},
			{"empty", [][]byte{}},
			{"nil entry", [][]byte{nil}},
			{"empty entry", [][]byte{[]byte("%PDF-1.4"), {}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := gopdflib.MergePDFs(tc.files); err == nil {
					t.Fatal("expected error")
				}
			})
		}
	})

	t.Run("SplitPDF", func(t *testing.T) {
		for _, data := range [][]byte{nil, {}} {
			if _, err := gopdflib.SplitPDF(data, gopdflib.SplitSpec{}); err == nil {
				t.Fatal("expected error for empty PDF")
			}
		}
	})

	t.Run("ParsePageSpec", func(t *testing.T) {
		// Empty spec is the established "select no pages" contract
		// (merge.ParsePageSpec returns nil, nil; python test_empty_spec).
		for _, spec := range []string{"", "   "} {
			pages, err := gopdflib.ParsePageSpec(spec, 10)
			if err != nil {
				t.Fatalf("empty spec %q must not error: %v", spec, err)
			}
			if len(pages) != 0 {
				t.Fatalf("empty spec %q must select no pages, got %v", spec, pages)
			}
		}
		if _, err := gopdflib.ParsePageSpec("1-2", 10); err != nil {
			t.Fatalf("valid spec must pass boundary: %v", err)
		}
		if _, err := gopdflib.ParsePageSpec("0", 10); err == nil {
			t.Fatal("expected error for page 0")
		}
	})

	t.Run("CompressPDF", func(t *testing.T) {
		for _, data := range [][]byte{nil, {}} {
			if _, err := gopdflib.CompressPDF(data, gopdflib.CompressOptions{}); err == nil {
				t.Fatal("expected error for empty PDF")
			}
		}
	})

	t.Run("FillPDFWithXFDF", func(t *testing.T) {
		if _, err := gopdflib.FillPDFWithXFDF(nil, []byte("<xfdf/>")); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, err := gopdflib.FillPDFWithXFDF([]byte("%PDF-1.4"), nil); err == nil {
			t.Fatal("expected error for empty XFDF")
		}
		if _, err := gopdflib.FillPDFWithXFDF([]byte("%PDF-1.4"), []byte{}); err == nil {
			t.Fatal("expected error for empty XFDF")
		}
	})

	t.Run("Redact", func(t *testing.T) {
		if _, err := gopdflib.GetPageInfo(nil); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, err := gopdflib.ExtractTextPositions([]byte{}, 1); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, err := gopdflib.FindTextOccurrences([]byte("%PDF-1.4"), ""); err == nil {
			t.Fatal("expected error for empty searchText")
		}
		if _, err := gopdflib.FindTextOccurrences(nil, "x"); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, err := gopdflib.ApplyRedactions(nil, nil); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, err := gopdflib.ApplyRedactionsAdvanced(nil, gopdflib.ApplyRedactionOptions{}); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, _, err := gopdflib.ApplyRedactionsAdvancedWithReport(nil, gopdflib.ApplyRedactionOptions{}); err == nil {
			t.Fatal("expected error for empty PDF")
		}
		if _, err := gopdflib.AnalyzePageCapabilities(nil); err == nil {
			t.Fatal("expected error for empty PDF")
		}
	})

	t.Run("HTML", func(t *testing.T) {
		if _, err := gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{}); err == nil {
			t.Fatal("expected error for empty HTML+URL")
		} else if !strings.Contains(err.Error(), "HTML content or a URL") {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := gopdflib.ConvertHTMLToImage(gopdflib.HTMLToImageRequest{}); err == nil {
			t.Fatal("expected error for empty HTML+URL")
		}
	})
}
