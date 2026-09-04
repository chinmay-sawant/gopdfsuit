// Package gopdflib provides PDF merging functionality.
package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

// MergePDFs combines multiple PDF files into a single PDF document.
// Files should be provided as byte slices in the desired order.
// At least one non-empty PDF is required.
//
// Example:
//
//	pdf1, _ := os.ReadFile("doc1.pdf")
//	pdf2, _ := os.ReadFile("doc2.pdf")
//	merged, err := gopdflib.MergePDFs([][]byte{pdf1, pdf2})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("merged.pdf", merged, 0644)
func MergePDFs(files [][]byte) ([]byte, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("gopdflib: MergePDFs needs at least 1 PDF file")
	}
	for i, f := range files {
		if len(f) == 0 {
			return nil, fmt.Errorf("gopdflib: MergePDFs file at index %d is empty", i)
		}
	}
	return merge.MergePDFs(files)
}
