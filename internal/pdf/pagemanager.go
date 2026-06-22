package pdf

import (
	"bytes"
	"strconv"
	"sync"
)

// P5/P34 (2026-06-20/22 checklists): per-page content stream buckets are sized
// from measured Zerodha output (retail 8.7 KiB/page, active ~18 KiB/page,
// HFT stripe ~28–35 KiB/page). Buffers above maxPooledPageContentStreamCap are
// dropped instead of being returned to the pool.
const (
	maxPooledPageContentStreamCap = 128 * 1024
	pageStreamRetailCap             = 32 * 1024  // measured 8,755 B
	pageStreamActiveCap             = 48 * 1024  // measured ~18 KiB/page
	pageStreamHFTRowBytes           = 120        // compliant BDC/EMC + text per cell
	pageStreamStripeSlack           = 4096       // border/watermark/page-init headroom
	pageStreamMinCap                = 32 * 1024
)

var pageContentStreamPools = [...]struct {
	capacity int
	pool     sync.Pool
}{
	{capacity: 32 * 1024},
	{capacity: 48 * 1024},
	{capacity: 64 * 1024},
	{capacity: 96 * 1024},
	{capacity: 128 * 1024},
}

func alignPageStreamCap(want int) int {
	if want < pageStreamMinCap {
		return pageStreamMinCap
	}
	bucket := pageContentStreamBucket(want)
	return pageContentStreamPools[bucket].capacity
}

func estimateSharedRowStripeCap(rows, cols int) int {
	if rows <= 0 || cols <= 0 {
		return pageStreamMinCap
	}
	est := rows*cols*pageStreamHFTRowBytes + pageStreamStripeSlack
	return alignPageStreamCap(est)
}

func getPageContentStreamBuffer(wantCap int) *bytes.Buffer {
	wantCap = alignPageStreamCap(wantCap)
	bucket := pageContentStreamBucket(wantCap)
	pooled := pageContentStreamPools[bucket].pool.Get()
	if pooled == nil {
		pooled = new(bytes.Buffer)
	}
	buf := pooled.(*bytes.Buffer)
	buf.Reset()
	if buf.Cap() < wantCap {
		buf.Grow(wantCap - buf.Cap())
	}
	return buf
}

func pageContentStreamBucket(initialCap int) int {
	if initialCap < 32*1024 {
		initialCap = 32 * 1024
	}
	for i := range pageContentStreamPools {
		if initialCap <= pageContentStreamPools[i].capacity {
			return i
		}
	}
	return len(pageContentStreamPools) - 1
}

func putPageContentStreamBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	// P5: discard oversized buffers instead of returning them to a too-small
	// bucket. This keeps the pool bounded even when a single HFT page grows
	// to >256KB during emission.
	if buf.Cap() > maxPooledPageContentStreamCap {
		return
	}
	bucket := pageContentStreamBucket(buf.Cap())
	buf.Reset()
	pageContentStreamPools[bucket].pool.Put(buf)
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
	sharedRowBytes          int                  // profiled shared-layout row emit size (P40)
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
	if initialStreamCap < pageStreamMinCap {
		initialStreamCap = pageStreamMinCap
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
	rowsPerPage := max(int((pm.CurrentYPos-pm.Margins.Bottom)/rowHeight), 1)
	est := estimateSharedRowStripeCap(rowsPerPage, cols)
	if est <= pm.InitialStreamCap {
		return
	}
	pm.InitialStreamCap = est
}

// NoteSharedRowBytes records the measured shared-layout row emit size for later stripes.
func (pm *PageManager) NoteSharedRowBytes(n int) {
	if n <= 0 {
		return
	}
	if pm.sharedRowBytes == 0 || n > pm.sharedRowBytes {
		pm.sharedRowBytes = n
	}
}

// GrowCurrentStreamForStripe pre-sizes the active page stream once per stripe.
func (pm *PageManager) GrowCurrentStreamForStripe(rows, cols int) {
	if rows <= 0 || cols <= 0 {
		return
	}
	perRow := cols * pageStreamHFTRowBytes
	if pm.sharedRowBytes > 0 {
		perRow = pm.sharedRowBytes
	}
	need := pm.GetCurrentContentStream().Len() + rows*perRow + pageStreamStripeSlack
	stream := pm.GetCurrentContentStream()
	if stream.Cap() < need {
		stream.Grow(need - stream.Cap())
	}
}

// PageStreamProfile returns max per-page stream len/cap after generation.
func (pm *PageManager) PageStreamProfile() (maxLen, maxCap, totalCap int) {
	for _, stream := range pm.ContentStreams {
		if stream == nil {
			continue
		}
		if stream.Len() > maxLen {
			maxLen = stream.Len()
		}
		if stream.Cap() > maxCap {
			maxCap = stream.Cap()
		}
		totalCap += stream.Cap()
	}
	return maxLen, maxCap, totalCap
}

// ContentWidth returns the available width for content on the current page.
func (pm *PageManager) ContentWidth() float64 {
	return pm.PageDimensions.Width - pm.Margins.Left - pm.Margins.Right
}

// GetCurrentContentStream returns the current page's content stream
func (pm *PageManager) GetCurrentContentStream() *bytes.Buffer {
	return pm.ContentStreams[pm.CurrentPageIndex]
}

// appendContentStream appends bytes to a page stream with a single upfront grow.
func appendContentStream(stream *bytes.Buffer, b []byte) {
	if len(b) == 0 {
		return
	}
	need := stream.Len() + len(b)
	if need > stream.Cap() {
		stream.Grow(need - stream.Cap())
	}
	_, _ = stream.Write(b)
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
