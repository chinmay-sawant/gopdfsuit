// Package merge provides PDF merging functionality without external dependencies.
// It properly handles form fields, annotations, appearance streams, and various PDF versions.
package merge

import (
	"bytes"
)

// ObjectBoundary represents the position of a PDF object in the file
type ObjectBoundary struct {
	ObjNum    int // Object number
	GenNum    int // Generation number
	Start     int // Position of "N G obj"
	BodyStart int // Position right after "obj"
	End       int // Position right after "endobj"
}

// PDFObject represents a parsed PDF object
type PDFObject struct {
	Number     int
	Generation int
	Body       []byte
	HasStream  bool
}

// MergeContext holds the state during a merge operation
type MergeContext struct {
	// Output buffer
	Output bytes.Buffer

	// Object tracking
	Offsets    map[int]int // Object ID -> byte offset in output
	CurrentMax int         // Highest object number assigned

	// Page tracking
	MergedPages []int // Ordered list of page object numbers

	// Form field tracking
	MergedFields   []int        // Form field object numbers
	FieldSet       map[int]bool // Quick lookup to avoid duplicates
	WidgetToFields map[int]int  // Widget object -> Field object mapping

	// Annotation tracking
	AnnotationDeps map[int][]int // Widget/Annotation -> dependent objects (AP streams, etc.)

	// Version tracking
	HighestVersion string
}

// NewMergeContext creates a new merge context
func NewMergeContext() *MergeContext {
	return &MergeContext{
		Offsets:        make(map[int]int),
		CurrentMax:     2, // Reserve 1 for Catalog, 2 for Pages
		FieldSet:       make(map[int]bool),
		WidgetToFields: make(map[int]int),
		AnnotationDeps: make(map[int][]int),
		HighestVersion: "1.4",
	}
}

// FileContext holds parsed data for a single input PDF
type FileContext struct {
	Data       []byte
	Objects    map[int][]byte // Object number -> body
	MaxObj     int            // Maximum object number in this file
	Pages      []int          // Page object numbers (in order)
	FormFields []int          // Form field object numbers
	Annots     map[int][]int  // Page object -> annotation object numbers
	APDeps     map[int][]int  // Widget -> appearance stream dependencies
}

// NewFileContext creates a new file context
func NewFileContext(data []byte) *FileContext {
	return &FileContext{
		Data:    data,
		Objects: make(map[int][]byte),
		Annots:  make(map[int][]int),
		APDeps:  make(map[int][]int),
	}
}
