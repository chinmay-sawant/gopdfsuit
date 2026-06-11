package pdf

import (
	"bytes"
	"strconv"
	"sync"
)

var pageContentStreamPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func getPageContentStreamBuffer(initialCap int) *bytes.Buffer {
	buf := pageContentStreamPool.Get().(*bytes.Buffer)
	buf.Reset()
	if initialCap < 64*1024 {
		initialCap = 64 * 1024
	}
	if cap(buf.Bytes()) < initialCap {
		buf.Grow(initialCap)
	}
	return buf
}

func putPageContentStreamBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	pageContentStreamPool.Put(buf)
}

// PageManager handles multi-page document generation
type PageManager struct {
	Pages                   []int   // List of page object IDs
	CurrentPageIndex        int     // Current page index (0-based)
	CurrentYPos             float64 // Current Y position on page
	PageDimensions          PageDimensions
	Margins                 PageMargins
	ContentStreams          []*bytes.Buffer      // Content for each page
	PageAnnots              [][]int              // Annotation Object IDs per page
	ExtraObjects            map[int][]byte       // Object ID -> Object Content
	NextObjectID            int                  // Counter for new objects
	ArlingtonCompatible     bool                 // Whether to use Arlington Model compliant fonts
	Structure               *StructureManager    // PDF/UA Structure Tree Manager
	NextAnnotStructParent   int                  // PDF/UA-2: Counter for annotation StructParent values
	AnnotStructElems        []AnnotStructElem    // PDF/UA-2: Annotation to structure element mapping
	NamedDests              map[string]NamedDest // Map of named destinations for internal linking
	FontRegistry            *CustomFontRegistry  // Per-generation font registry for thread-safe font access
	InitialStreamCap        int                  // Initial capacity for pooled page content streams
	cachedPageInit          []byte               // Reused border/watermark init bytes for continuation pages (C3)
	cachedPageInitBorder    string
	cachedPageInitWatermark string
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

// NewPageManager creates a new page manager with initial page.
// When taggedPDF is false, marked content and structure-tree bookkeeping are disabled.
func NewPageManager(pageDims PageDimensions, margins PageMargins, arlingtonCompatible bool, fontRegistry *CustomFontRegistry, taggedPDF bool, initialStreamCap int) *PageManager {
	if initialStreamCap < 64*1024 {
		initialStreamCap = 64 * 1024
	}
	firstStream := getPageContentStreamBuffer(initialStreamCap)
	pm := &PageManager{
		Pages:                 []int{3}, // First page starts at object 3
		CurrentPageIndex:      0,        // Start with first page
		CurrentYPos:           pageDims.Height - margins.Top,
		PageDimensions:        pageDims,
		Margins:               margins,
		ContentStreams:        []*bytes.Buffer{firstStream},
		PageAnnots:            make([][]int, 1),
		ExtraObjects:          make(map[int][]byte),
		NextObjectID:          2000, // Start extra objects at 2000 to avoid conflicts
		ArlingtonCompatible:   arlingtonCompatible,
		Structure:             NewStructureManager(taggedPDF),
		NextAnnotStructParent: 1000, // Start annotation StructParents at 1000 to avoid conflicts with page StructParents
		AnnotStructElems:      make([]AnnotStructElem, 0),
		NamedDests:            make(map[string]NamedDest),
		FontRegistry:          fontRegistry,
		InitialStreamCap:      initialStreamCap,
	}
	return pm
}

// AddNewPage creates a new page when current page is full
func (pm *PageManager) AddNewPage() {
	// Calculate next page object ID
	nextPageID := 3 + len(pm.Pages) // Sequential page IDs starting from 3
	pm.Pages = append(pm.Pages, nextPageID)
	pm.CurrentPageIndex = len(pm.Pages) - 1 // Move to new page
	pm.CurrentYPos = pm.PageDimensions.Height - pm.Margins.Top
	nb := getPageContentStreamBuffer(pm.InitialStreamCap)
	pm.ContentStreams = append(pm.ContentStreams, nb)
	pm.PageAnnots = append(pm.PageAnnots, []int{})
}

// ReleaseContentStreams returns pooled page content buffers after PDF generation completes.
func (pm *PageManager) ReleaseContentStreams() {
	for i, stream := range pm.ContentStreams {
		putPageContentStreamBuffer(stream)
		pm.ContentStreams[i] = nil
	}
	pm.ContentStreams = nil
}

// AddAnnotation adds an annotation object ID to the current page
func (pm *PageManager) AddAnnotation(objID int) {
	pm.PageAnnots[pm.CurrentPageIndex] = append(pm.PageAnnots[pm.CurrentPageIndex], objID)
}

// AddExtraObject adds an extra object (like a widget) to the manager
func (pm *PageManager) AddExtraObject(content string) int {
	id := pm.NextObjectID
	pm.NextObjectID++
	pm.ExtraObjects[id] = []byte(content)
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

	// PDF Rectangle: [LLx LLy URx URy]
	validURL := escapePDFString(url)
	var content []byte
	if pm.Structure.Enabled {
		structParentIdx := pm.GetNextAnnotStructParent()
		content = append(content, "<< /Type /Annot /Subtype /Link /Rect ["...)
		content = appendRect(content, x, y, w, h)
		content = append(content, "] /Border [0 0 0] /F 4 /StructParent "...)
		content = strconv.AppendInt(content, int64(structParentIdx), 10)
		content = append(content, " /A << /Type /Action /S /URI /URI ("...)
		content = append(content, validURL...)
		content = append(content, ") >> >>"...)
		pm.ExtraObjects[annotID] = content
		pm.AddAnnotation(annotID)
		pm.AddLinkStructureElement(annotID, structParentIdx)
		return
	}

	content = append(content, "<< /Type /Annot /Subtype /Link /Rect ["...)
	content = appendRect(content, x, y, w, h)
	content = append(content, "] /Border [0 0 0] /F 4 /A << /Type /Action /S /URI /URI ("...)
	content = append(content, validURL...)
	content = append(content, ") >> >>"...)

	pm.ExtraObjects[annotID] = content
	pm.AddAnnotation(annotID)
}

func appendRect(dst []byte, x, y, w, h float64) []byte {
	dst = appendFmtNum(dst, x)
	dst = append(dst, ' ')
	dst = appendFmtNum(dst, y)
	dst = append(dst, ' ')
	dst = appendFmtNum(dst, x+w)
	dst = append(dst, ' ')
	dst = appendFmtNum(dst, y+h)
	return dst
}

// CheckPageBreak determines if a new page is needed based on required height
func (pm *PageManager) CheckPageBreak(requiredHeight float64) bool {
	return pm.CurrentYPos-requiredHeight < pm.Margins.Bottom
}

// RowsFitOnCurrentPage estimates how many rows of rowHeight fit above the bottom margin.
func (pm *PageManager) RowsFitOnCurrentPage(rowHeight float64) int {
	if rowHeight <= 0 {
		return 1
	}
	available := pm.CurrentYPos - pm.Margins.Bottom
	n := int(available / rowHeight)
	if n < 1 {
		return 1
	}
	return n
}

// PrepareLargeTableStripe sizes per-page stream capacity from row geometry for large tables.
func (pm *PageManager) PrepareLargeTableStripe(rowHeight float64, cols int) {
	if rowHeight <= 0 || cols <= 0 {
		return
	}
	rowsPerPage := int((pm.CurrentYPos - pm.Margins.Bottom) / rowHeight)
	if rowsPerPage < 1 {
		rowsPerPage = 1
	}
	est := rowsPerPage * cols * 512
	if est <= pm.InitialStreamCap {
		return
	}
	const maxPageStreamCap = 256 * 1024
	if est > maxPageStreamCap {
		est = maxPageStreamCap
	}
	pm.InitialStreamCap = est
}

// ContentWidth returns the available width for content on the current page.
func (pm *PageManager) ContentWidth() float64 {
	return pm.PageDimensions.Width - pm.Margins.Left - pm.Margins.Right
}

// GetCurrentContentStream returns the current page's content stream
func (pm *PageManager) GetCurrentContentStream() *bytes.Buffer {
	return pm.ContentStreams[pm.CurrentPageIndex]
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
