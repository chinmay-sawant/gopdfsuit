package pdf

import (
	"bytes"
	"fmt"
)

// PageManager handles multi-page document generation
type PageManager struct {
	Pages                 []int   // List of page object IDs
	CurrentPageIndex      int     // Current page index (0-based)
	CurrentYPos           float64 // Current Y position on page
	PageDimensions        PageDimensions
	ContentStreams        []bytes.Buffer       // Content for each page
	PageAnnots            [][]int              // Annotation Object IDs per page
	ExtraObjects          map[int]string       // Object ID -> Object Content
	NextObjectID          int                  // Counter for new objects
	ArlingtonCompatible   bool                 // Whether to use Arlington Model compliant fonts
	Structure             *StructureManager    // PDF/UA Structure Tree Manager
	NextAnnotStructParent int                  // PDF/UA-2: Counter for annotation StructParent values
	AnnotStructElems      []AnnotStructElem    // PDF/UA-2: Annotation to structure element mapping
	NamedDests            map[string]NamedDest // Map of named destinations for internal linking
	FontRegistry          *CustomFontRegistry  // Per-generation font registry for thread-safe font access
}

// AnnotStructElem tracks the relationship between an annotation and its structure element
type AnnotStructElem struct {
	AnnotObjID      int // Annotation object ID
	StructParentIdx int // StructParent index for ParentTree
	PageIndex       int // Page where the annotation appears
}

// NamedDest represents a named destination in the PDF
type NamedDest struct {
	PageIndex    int     // 0-based page index
	Y            float64 // Y position on page
	StructElemID int     // PDF/UA-2: Structure Element ID for /SD
}

// NewPageManager creates a new page manager with initial page
func NewPageManager(pageDims PageDimensions, arlingtonCompatible bool, fontRegistry *CustomFontRegistry) *PageManager {
	pm := &PageManager{
		Pages:                 []int{3}, // First page starts at object 3
		CurrentPageIndex:      0,        // Start with first page
		CurrentYPos:           pageDims.Height - margin,
		PageDimensions:        pageDims,
		ContentStreams:        make([]bytes.Buffer, 1),
		PageAnnots:            make([][]int, 1),
		ExtraObjects:          make(map[int]string),
		NextObjectID:          2000, // Start extra objects at 2000 to avoid conflicts
		ArlingtonCompatible:   arlingtonCompatible,
		Structure:             NewStructureManager(),
		NextAnnotStructParent: 1000, // Start annotation StructParents at 1000 to avoid conflicts with page StructParents
		AnnotStructElems:      make([]AnnotStructElem, 0),
		NamedDests:            make(map[string]NamedDest),
		FontRegistry:          fontRegistry,
	}
	return pm
}

// AddNewPage creates a new page when current page is full
func (pm *PageManager) AddNewPage() {
	// Calculate next page object ID
	nextPageID := 3 + len(pm.Pages) // Sequential page IDs starting from 3
	pm.Pages = append(pm.Pages, nextPageID)
	pm.CurrentPageIndex = len(pm.Pages) - 1 // Move to new page
	pm.CurrentYPos = pm.PageDimensions.Height - margin
	pm.ContentStreams = append(pm.ContentStreams, bytes.Buffer{})
	pm.PageAnnots = append(pm.PageAnnots, []int{})
}

// AddAnnotation adds an annotation object ID to the current page
func (pm *PageManager) AddAnnotation(objID int) {
	pm.PageAnnots[pm.CurrentPageIndex] = append(pm.PageAnnots[pm.CurrentPageIndex], objID)
}

// AddExtraObject adds an extra object (like a widget) to the manager
func (pm *PageManager) AddExtraObject(content string) int {
	id := pm.NextObjectID
	pm.NextObjectID++
	pm.ExtraObjects[id] = content
	return id
}

// AddLinkAnnotation adds a link annotation to the current page
func (pm *PageManager) AddLinkAnnotation(x, y, w, h float64, url string) {
	if url == "" {
		return
	}

	// Create annotation object
	annotID := pm.NextObjectID
	pm.NextObjectID++

	// PDF/UA-2: Get StructParent index for this annotation
	structParentIdx := pm.GetNextAnnotStructParent()

	// PDF Rectangle: [LLx LLy URx URy]
	rect := fmt.Sprintf("[%s %s %s %s]", fmtNum(x), fmtNum(y), fmtNum(x+w), fmtNum(y+h))

	validUrl := escapePDFString(url)

	// PDF/UA-2: Include StructParent entry
	content := fmt.Sprintf("<< /Type /Annot /Subtype /Link /Rect %s /Border [0 0 0] /F 4 /StructParent %d /A << /Type /Action /S /URI /URI (%s) >> >>",
		rect, structParentIdx, validUrl)

	pm.ExtraObjects[annotID] = content
	pm.AddAnnotation(annotID)

	// PDF/UA-2: Create Link structure element for this annotation
	pm.AddLinkStructureElement(annotID, structParentIdx)
}

// CheckPageBreak determines if a new page is needed based on required height
func (pm *PageManager) CheckPageBreak(requiredHeight float64) bool {
	return pm.CurrentYPos-requiredHeight < margin
}

// GetCurrentContentStream returns the current page's content stream
func (pm *PageManager) GetCurrentContentStream() *bytes.Buffer {
	return &pm.ContentStreams[pm.CurrentPageIndex]
}

// GetCurrentPageID returns the current page object ID
func (pm *PageManager) GetCurrentPageID() int {
	return pm.Pages[pm.CurrentPageIndex]
}

// GetNextAnnotStructParent returns and increments the StructParent counter for annotations
// PDF/UA-2: Each annotation needs a unique StructParent value for ParentTree lookup
func (pm *PageManager) GetNextAnnotStructParent() int {
	idx := pm.NextAnnotStructParent
	pm.NextAnnotStructParent++
	return idx
}

// AddLinkStructureElement creates a Link structure element for an annotation
// PDF/UA-2 requires link annotations to be wrapped in Link structure elements
func (pm *PageManager) AddLinkStructureElement(annotObjID int, structParentIdx int) {
	// Track the annotation-to-structure relationship
	pm.AnnotStructElems = append(pm.AnnotStructElems, AnnotStructElem{
		AnnotObjID:      annotObjID,
		StructParentIdx: structParentIdx,
		PageIndex:       pm.CurrentPageIndex,
	})

	// Create a Link structure element in the structure tree
	// This will be properly connected during PDF generation
	pm.Structure.AddLinkElement(annotObjID, pm.CurrentPageIndex)
}
