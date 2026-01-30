// Package gopdflib provides PDF merging functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/internal/pdf/merge"
)

// MergePDFs combines multiple PDF files into a single PDF document.
// Files should be provided as byte slices in the desired order.
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
	return merge.MergePDFs(files)
}
