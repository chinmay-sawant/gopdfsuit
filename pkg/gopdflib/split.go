// Package gopdflib provides PDF splitting functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/internal/pdf/merge"
)

// SplitSpec defines split criteria for splitting PDFs.
type SplitSpec = merge.SplitSpec

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
	return merge.SplitPDF(file, spec)
}

// ParsePageSpec parses a page specification string like "1-3,5,7-9" into a sorted
// slice of 1-based page numbers. If totalPages is provided (>0), it validates
// that pages don't exceed the total.
//
// Example:
//
//	pages, err := gopdflib.ParsePageSpec("1-3,5,7-9", 10)
//	// pages = [1, 2, 3, 5, 7, 8, 9]
func ParsePageSpec(spec string, totalPages int) ([]int, error) {
	return merge.ParsePageSpec(spec, totalPages)
}
