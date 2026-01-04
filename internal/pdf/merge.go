package pdf

import (
	"github.com/chinmay-sawant/gopdfsuit/internal/pdf/merge"
)

// MergePDFs merges multiple PDF files into one.
// This is the main entry point that delegates to the merge package.
// It properly handles form fields, widgets, appearance streams, and various PDF versions.
func MergePDFs(files [][]byte) ([]byte, error) {
	return merge.MergePDFs(files)
}
