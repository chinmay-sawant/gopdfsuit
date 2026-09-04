// Package gopdflib provides PDF merging functionality.
package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

// MaxMergeInputBytes caps a single input file accepted by MergePDFs
// (32 MiB, shared with the engine). The WASM shim enforces the same cap
// per Uint8Array before copying the files into the module.
const MaxMergeInputBytes = merge.MaxMergeInputBytes

// MaxMergeFileCount caps the number of input files accepted by MergePDFs.
// The CGO MergePDFs entry point enforces the same cap on its count argument
// before copying parts, so all callers share one policy sourced here.
const MaxMergeFileCount = 1 << 16

// MergePDFs combines multiple PDF files into a single PDF document.
// Files should be provided as byte slices in the desired order.
// At least one non-empty PDF is required.
//
// WASM mapping: the JS shim passes an Array of Uint8Array, copying each
// element with js.CopyBytesToGo into a [][]byte in order and calling this
// function directly. No other translation is needed.
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
	const op = "gopdflib: MergePDFs"
	if len(files) == 0 {
		return nil, invalidInputError(op, "needs at least 1 PDF file")
	}
	if len(files) >= MaxMergeFileCount {
		return nil, invalidInputError(op, "too many PDF files")
	}
	for i, f := range files {
		if len(f) == 0 {
			return nil, invalidInputError(op, fmt.Sprintf("file at index %d is empty", i))
		}
	}
	out, err := merge.MergePDFs(files)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}
