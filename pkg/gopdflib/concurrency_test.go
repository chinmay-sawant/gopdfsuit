package gopdflib_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

var errNotPDF = errors.New("output is not a PDF")

func seedTemplate(title string) gopdflib.PDFTemplate {
	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{Page: "A4", PageAlignment: 1},
		Title:  gopdflib.Title{Props: "Helvetica:14:100:center:0:0:0:0", Text: title},
	}
}

// TestConcurrentGenerateSharedFonts runs concurrent GeneratePDF calls sharing
// the global font registry. Run with -race: no data race and all outputs
// must be valid PDFs.
func TestConcurrentGenerateSharedFonts(t *testing.T) {
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out, err := gopdflib.GeneratePDF(seedTemplate("Race Title"))
			if err != nil {
				errs[i] = err
				return
			}
			if !bytes.HasPrefix(out, []byte("%PDF-")) {
				errs[i] = errNotPDF
			}
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
}

func TestConcurrentPreparedTemplateGeneration(t *testing.T) {
	prepared, err := gopdflib.PrepareTemplate(seedTemplate("Prepared Race"))
	if err != nil {
		t.Fatalf("PrepareTemplate: %v", err)
	}
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doc, err := prepared.GeneratePDFBorrowed()
			if err != nil {
				errs[i] = err
				return
			}
			if !bytes.HasPrefix(doc.Bytes(), []byte("%PDF-")) {
				errs[i] = errNotPDF
			}
			doc.Release()
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
}

func TestConcurrentPreparedPDFATitleGeneration(t *testing.T) {
	tmpl := seedTemplate("PDF/A Race")
	tmpl.Config.PdfTitle = "PDF/A Race"
	tmpl.Config.PDFA = &gopdflib.PDFAConfig{Enabled: true, Conformance: "4"}
	prepared, err := gopdflib.PrepareTemplate(tmpl)
	if err != nil {
		t.Fatalf("PrepareTemplate: %v", err)
	}

	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doc, err := prepared.GeneratePDFBorrowed()
			if err != nil {
				errs[i] = err
				return
			}
			if !bytes.HasPrefix(doc.Bytes(), []byte("%PDF-")) {
				errs[i] = errNotPDF
			}
			doc.Release()
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
}

// TestBorrowedCopyBytes verifies the copying helper survives Release.
func TestBorrowedCopyBytes(t *testing.T) {
	doc, err := gopdflib.GeneratePDFBorrowed(seedTemplate("Borrowed"))
	if err != nil {
		t.Fatalf("GeneratePDFBorrowed: %v", err)
	}
	borrowed := doc.Bytes()
	if len(borrowed) == 0 {
		t.Fatal("expected non-empty borrowed bytes")
	}
	copied := doc.CopyBytes()
	if !bytes.Equal(borrowed, copied) {
		t.Fatal("CopyBytes must equal Bytes before Release")
	}
	doc.Release()
	if !bytes.HasPrefix(copied, []byte("%PDF-")) {
		t.Fatal("copied bytes must remain valid after Release")
	}
	var nilDoc *gopdflib.BorrowedPDF
	if nilDoc.CopyBytes() != nil {
		t.Fatal("nil CopyBytes must be nil")
	}
	nilDoc.Release() // must not panic
}

// TestWarmRuntimePoolsIdempotent asserts repeated warm calls plus generation work.
func TestWarmRuntimePoolsIdempotent(t *testing.T) {
	for range 5 {
		gopdflib.WarmRuntimePools()
	}
	out, err := gopdflib.GeneratePDF(seedTemplate("Warm"))
	if err != nil {
		t.Fatalf("GeneratePDF after repeated warm: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("expected a PDF")
	}
}
