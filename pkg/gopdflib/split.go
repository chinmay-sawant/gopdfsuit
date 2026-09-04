// Package gopdflib provides PDF splitting functionality.
package gopdflib

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
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

// splitSpecJSON is the WASM/JSON shape for a split request: a JS object
// serialized with JSON.stringify, e.g. {pages: [1,3,5], maxPerFile: 2} or
// {pages: "1-3,5", max_per_file: 2}. pages accepts either an array of
// 1-based page numbers or a page-spec string parsed by ParsePageSpec.
type splitSpecJSON struct {
	Pages         any      `json:"pages"`
	Ranges        [][2]int `json:"ranges"`
	MaxPerFile    *int     `json:"maxPerFile"`
	MaxPerFileAlt *int     `json:"max_per_file"`
}

// ParseSplitSpecJSON parses a split-spec object for the WASM split binding:
// the JS caller passes {pages, maxPerFile} (pages as an array of 1-based
// page numbers or a "1-3,5" string, maxPerFile as a positive int) and the
// shim forwards the serialized JSON here. A nil or empty document selects
// all pages in a single file. max_per_file is accepted as an alias; when
// both keys are present maxPerFile wins. Unknown fields are ignored.
func ParseSplitSpecJSON(data []byte) (SplitSpec, error) {
	const op = "gopdflib: ParseSplitSpecJSON"
	if len(data) == 0 {
		return SplitSpec{}, nil
	}
	var raw splitSpecJSON
	if err := sonic.Unmarshal(data, &raw); err != nil {
		return SplitSpec{}, fmt.Errorf("%w: %s: %w", ErrInvalidInput, op, err)
	}
	var spec SplitSpec
	pages, err := parseSplitPagesValue(raw.Pages)
	if err != nil {
		return SplitSpec{}, fmt.Errorf("%w: %s: %w", ErrInvalidInput, op, err)
	}
	spec.Pages = pages
	spec.Ranges = raw.Ranges
	for _, r := range spec.Ranges {
		if r[0] < 1 || r[1] < r[0] {
			return SplitSpec{}, invalidInputError(op, fmt.Sprintf("invalid range: [%d,%d]", r[0], r[1]))
		}
	}
	maxPerFile := 0
	switch {
	case raw.MaxPerFile != nil:
		maxPerFile = *raw.MaxPerFile
	case raw.MaxPerFileAlt != nil:
		maxPerFile = *raw.MaxPerFileAlt
	}
	if maxPerFile < 0 {
		return SplitSpec{}, invalidInputError(op, "maxPerFile must not be negative")
	}
	spec.MaxPerFile = maxPerFile
	return spec, nil
}

// parseSplitPagesValue normalizes the pages member of a split-spec object:
// nil selects no explicit pages, a string is parsed as a page-spec, a JSON
// array holds 1-based page numbers, and a single JSON number selects one page.
func parseSplitPagesValue(v any) ([]int, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) == "" {
			return nil, nil
		}
		return merge.ParsePageSpec(t, 0)
	case float64:
		n := int(t)
		if t != float64(n) || n < 1 {
			return nil, fmt.Errorf("invalid page: %v", t)
		}
		return []int{n}, nil
	case []any:
		pages := make([]int, 0, len(t))
		for _, e := range t {
			f, ok := e.(float64)
			if !ok {
				return nil, fmt.Errorf("invalid page: %v", e)
			}
			n := int(f)
			if f != float64(n) || n < 1 {
				return nil, fmt.Errorf("invalid page: %v", e)
			}
			pages = append(pages, n)
		}
		return pages, nil
	default:
		return nil, fmt.Errorf("invalid pages: expected an array of page numbers or a spec string, got %T", v)
	}
}

// SplitPDFWithSpecJSON splits a PDF using a serialized split-spec object
// (see ParseSplitSpecJSON) instead of a SplitSpec struct. It is the WASM
// split binding: (pdfBytes, specObject) in, one [][]byte out. Multi-file
// results are returned as a slice of PDFs; the JS shim converts each part
// to a Uint8Array array. Zipping is done in JS, never Go-side, so
// archive/zip stays out of the WASM closure.
func SplitPDFWithSpecJSON(file []byte, specJSON []byte) ([][]byte, error) {
	spec, err := ParseSplitSpecJSON(specJSON)
	if err != nil {
		return nil, err
	}
	return SplitPDF(file, spec)
}
