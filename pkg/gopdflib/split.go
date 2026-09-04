// Package gopdflib provides PDF splitting functionality.
package gopdflib

import (
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

// SplitPDF splits a PDF into multiple parts based on the specification.
// Returns a slice of PDF byte slices, one for each output part.
//
// Example - Split specific pages:
//
//	pdfBytes, _ := os.ReadFile("document.pdf")
//	spec := gopdflib.SplitSpec{Pages: []int{1, 3, 5}}
//	parts, err := gopdflib.SplitPDF(pdfBytes, spec)
//
// Example - Split every N pages:
//
//	spec := gopdflib.SplitSpec{MaxPerFile: 5}
//	parts, err := gopdflib.SplitPDF(pdfBytes, spec)
func SplitPDF(file []byte, spec SplitSpec) ([][]byte, error) {
	const op = "gopdflib: SplitPDF"
	if len(file) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	parts, err := merge.SplitPDF(file, toInternalSplitSpec(spec))
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return parts, nil
}

// ParsePageSpec parses a page specification string like "1-3,5,7-9" into a sorted
// slice of 1-based page numbers. If totalPages is provided (>0), it validates
// that pages don't exceed the total. An empty (or whitespace-only) spec
// selects no pages and returns an empty slice with no error.
//
// Example:
//
//	pages, err := gopdflib.ParsePageSpec("1-3,5,7-9", 10)
//	// pages = [1, 2, 3, 5, 7, 8, 9]
func ParsePageSpec(spec string, totalPages int) ([]int, error) {
	if strings.TrimSpace(spec) == "" {
		return nil, nil
	}
	pages, err := merge.ParsePageSpec(spec, totalPages)
	if err != nil {
		return nil, wrapEngineError("gopdflib: ParsePageSpec", err)
	}
	return pages, nil
}
