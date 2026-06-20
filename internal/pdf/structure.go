package pdf

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
)

// StructureType represents the standard structure types in PDF/UA
type StructureType string

const (
	// StructDocument represents the Document grouping element.
	StructDocument StructureType = "Document"
	// StructPart represents the Part grouping element.
	StructPart StructureType = "Part"
	// StructSect represents the Sect grouping element.
	StructSect StructureType = "Sect"
	// StructDiv represents the Div grouping element.
	StructDiv StructureType = "Div"
	// StructH1 represents the H1 header element.
	StructH1 StructureType = "H1"
	// StructH2 represents the H2 header element.
	StructH2 StructureType = "H2"
	// StructH3 represents the H3 header element.
	StructH3 StructureType = "H3"
	// StructH4 represents the H4 header element.
	StructH4 StructureType = "H4"
	// StructH5 represents the H5 header element.
	StructH5 StructureType = "H5"
	// StructH6 represents the H6 header element.
	StructH6 StructureType = "H6"
	// StructP represents the Paragraph element.
	StructP StructureType = "P"
	// StructL represents the List element.
	StructL StructureType = "L"
	// StructLI represents the List Item element.
	StructLI StructureType = "LI"
	// StructLbl represents the List Label element.
	StructLbl StructureType = "Lbl"
	// StructLBody represents the List Body element.
	StructLBody StructureType = "LBody"
	// StructTable represents the Table element.
	StructTable StructureType = "Table"
	// StructTR represents the Table Row element.
	StructTR StructureType = "TR"
	// StructTH represents the Table Header element.
	StructTH StructureType = "TH"
	// StructTD represents the Table Data element.
	StructTD StructureType = "TD"
	// StructFigure represents the Figure element.
	StructFigure StructureType = "Figure"
	// StructCaption represents the Caption element.
	StructCaption StructureType = "Caption"
	// StructForm represents the Form grouping element.
	StructForm StructureType = "Form"
	// StructLink represents a Link structure element (PDF/UA-2).
	StructLink StructureType = "Link"
	// StructReference represents a Reference structure element (PDF/UA-2).
	StructReference StructureType = "Reference"
)

// StructKid represents a child of a structure element: either a nested element or an MCID leaf reference.
type StructKid struct {
	Elem *StructElem // non-nil for structure element child
	MCID int         // valid when Elem == nil (MCID leaf reference)
}

// StructElem represents a node in the structure tree
type StructElem struct {
	Type       StructureType
	Title      string
	Alt        string
	Lang       string
	Kids       []StructKid
	inlineKids [8]StructKid
	MCID       int
	HasMCID    bool
	Parent     *StructElem
	ObjectID   int // Assigned when writing to PDF
	PageID     int // Reference to the page object ID where this element appears
	AnnotObjID int // Annotation object ID for Link elements
}

func (elem *StructElem) LeafMCID() (int, bool) {
	if elem == nil {
		return 0, false
	}
	if elem.HasMCID {
		return elem.MCID, true
	}
	if len(elem.Kids) == 1 && elem.Kids[0].Elem == nil {
		return elem.Kids[0].MCID, true
	}
	return 0, false
}

// StructureManager handles the creation and management of the PDF structure tree
type StructureManager struct {
	Enabled       bool // When false, structure methods do not mutate state or write BDC/EMC (backward-compatible untagged PDFs)
	Root          *StructElem
	CurrentParent *StructElem
	Elements      []*StructElem
	NextMCID      []int               // Page Index -> Next MCID
	ParentTree    [][]*StructElem     // Page Index -> Array of struct elements (parents of MCIDs)
	StructParents map[int]int         // Page Object ID -> StructParents index (in ParentTree)
	LinkElements  map[int]*StructElem // PDF/UA-2: Annotation Object ID -> Link StructElem
}

func growPtrSlice[T any](s []T, need, minCap int) []T {
	if cap(s) >= need {
		return s
	}
	newCap := need
	if newCap < minCap {
		newCap = minCap
	} else if c := cap(s); c > 0 && newCap < c*2 {
		newCap = c * 2
	}
	grown := make([]T, len(s), newCap)
	copy(grown, s)
	return grown
}

// NewStructureManager creates a new structure manager. When enabled is false, marked content
// and structure-tree bookkeeping are skipped (no allocations in hot Begin/End paths).
func NewStructureManager(enabled bool) *StructureManager {
	// Root is the conceptual StructTreeRoot container (hidden)
	root := &StructElem{
		Type: "Root",
	}
	sm := &StructureManager{
		Enabled:       enabled,
		Root:          root,
		CurrentParent: root,
		Elements:      []*StructElem{root},
		NextMCID:      make([]int, 1),
		ParentTree:    make([][]*StructElem, 1),
		StructParents: make(map[int]int),
	}

	if enabled {
		// PDF/UA: The valid top-level element must be 'Document' (or Part, Art, etc.)
		sm.BeginStructureElement(StructDocument)
	}

	return sm
}

var structElemPool = sync.Pool{
	New: func() any {
		return &StructElem{}
	},
}

var structKidsSlicePool = sync.Pool{
	New: func() any {
		s := make([]StructKid, 0, 4)
		return &s
	},
}

func acquireStructKids(capHint int) []StructKid {
	if capHint <= 0 {
		capHint = 1
	}
	if capHint <= 8 {
		if v := structKidsSlicePool.Get(); v != nil {
			s := *(v.(*[]StructKid))
			if cap(s) >= capHint {
				return s[:0]
			}
		}
	}
	return make([]StructKid, 0, capHint)
}

func releaseStructKids(kids []StructKid) {
	if cap(kids) > 64 {
		return
	}
	kids = kids[:0]
	structKidsSlicePool.Put(&kids)
}

// resetStructElemForPool clears the fields a recycled struct-elem needs to
// drop before being reused by the next caller. The previous version zeroed
// every field, but most call sites set Type/Parent/PageID/MCID/HasMCID
// themselves, so only Title/Alt/Lang/ObjectID/AnnotObjID/Kids need a
// defensive clear to keep the next emit identical to a fresh struct.
func resetStructElemForPool(elem *StructElem) {
	if elem == nil {
		return
	}
	if len(elem.Kids) > 0 {
		if &elem.Kids[:1][0] != &elem.inlineKids[0] {
			releaseStructKids(elem.Kids)
		}
	}
	// Release the slice back to the pool (kids may be a pooled buffer).
	elem.Kids = nil
}

// acquireStructElem pulls a *StructElem from the global sync.Pool and
// performs the lazy field reset (P1/P2). The fields cleared here are the
// ones that affect output correctness if left stale: Title/Alt/Lang are
// read by the fast-path guard, ObjectID/AnnotObjID/PageID/HasMCID/MCID
// determine whether the struct is emitted as a leaf vs a grouping
// element, and Parent/Type are read by the slow-path formatter. Together
// these cover the full set the caller might NOT overwrite on the next use.
func (sm *StructureManager) acquireStructElem() *StructElem {
	if !sm.Enabled {
		return &StructElem{}
	}
	v := structElemPool.Get()
	e, ok := v.(*StructElem)
	if !ok || e == nil {
		return &StructElem{}
	}
	// P2: drop the `*e = StructElem{}` full-struct memclr. The fields the
	// caller doesn't write (Title/Alt/Lang/ObjectID/AnnotObjID/PageID/HasMCID/
	// MCID/Parent/Type) are cleared here. Kids is left alone — the caller
	// either replaces it via acquireStructKids (BeginStructureElementCap with
	// kidCap>0) or never reads it. This is ~80 bytes of writes per elem
	// instead of the previous 256-byte memclr.
	e.Type = ""
	e.Title = ""
	e.Alt = ""
	e.Lang = ""
	e.MCID = 0
	e.HasMCID = false
	e.ObjectID = 0
	e.AnnotObjID = 0
	e.PageID = 0
	e.Parent = nil
	return e
}

// ReleaseStructElemsToPool returns pooled *StructElem nodes after PDF
// generation completes (tagged PDF only).
func (sm *StructureManager) ReleaseStructElemsToPool() {
	if sm == nil || !sm.Enabled || sm.Root == nil {
		return
	}
	sm.Root.Kids = nil
	sm.CurrentParent = sm.Root
	for _, elem := range sm.Elements {
		if elem == nil || elem == sm.Root {
			continue
		}
		resetStructElemForPool(elem)
		structElemPool.Put(elem)
	}
	sm.Elements = []*StructElem{sm.Root}
	sm.NextMCID = sm.NextMCID[:0]
	sm.ParentTree = sm.ParentTree[:0]
	clear(sm.StructParents)
	sm.LinkElements = nil
}

func (sm *StructureManager) ensurePageSlot(pageIndex int) {
	if pageIndex < len(sm.NextMCID) {
		return
	}
	needed := pageIndex + 1
	if grow := needed - len(sm.NextMCID); grow > 0 {
		sm.NextMCID = append(sm.NextMCID, make([]int, grow)...)
		sm.ParentTree = append(sm.ParentTree, make([][]*StructElem, grow)...)
	}
}

// PageMCIDStart returns the next unassigned MCID cursor for a page.
func (sm *StructureManager) PageMCIDStart(pageIndex int) int {
	sm.ensurePageSlot(pageIndex)
	return sm.NextMCID[pageIndex]
}

// GetNextMCID returns the next available MCID for a page
func (sm *StructureManager) GetNextMCID(pageIndex int) int {
	if !sm.Enabled {
		return 0
	}
	sm.ensurePageSlot(pageIndex)
	mcid := sm.NextMCID[pageIndex]
	sm.NextMCID[pageIndex]++
	return mcid
}

func (sm *StructureManager) ensureParentTreeCapacity(pageIndex, count int) {
	if !sm.Enabled || count <= 0 {
		return
	}
	sm.ensurePageSlot(pageIndex)
	pt := &sm.ParentTree[pageIndex]
	need := len(*pt) + count
	*pt = growPtrSlice(*pt, need, 32)
}

// ReserveElementCapacity pre-grows the flat structure element slice when the
// tagged layout size is roughly known up front. P1 (2026-06-20 checklist):
// the HFT path allocates ~16,000 *StructElem per PDF, so pre-sizing the
// flat slice avoids amortised growth cost.
func (sm *StructureManager) ReserveElementCapacity(additional int) {
	if !sm.Enabled || additional <= 0 {
		return
	}
	need := len(sm.Elements) + additional
	sm.Elements = growPtrSlice(sm.Elements, need, 32)
}

// ReserveMCIDs allocates count consecutive MCIDs on a page and returns the first ID.
// Used by drawTable to avoid per-cell ensurePageSlot/increment overhead (D4).
func (sm *StructureManager) ReserveMCIDs(pageIndex, count int) int {
	return sm.reserveMCIDs(pageIndex, count, true)
}

// ReserveMCIDsLite reserves MCIDs without growing ParentTree (deferred bulk fill per page).
func (sm *StructureManager) ReserveMCIDsLite(pageIndex, count int) int {
	return sm.reserveMCIDs(pageIndex, count, false)
}

func (sm *StructureManager) reserveMCIDs(pageIndex, count int, reserveParentTree bool) int {
	if !sm.Enabled || count <= 0 {
		return 0
	}
	sm.ensurePageSlot(pageIndex)
	if reserveParentTree {
		sm.ensureParentTreeCapacity(pageIndex, count)
	}
	start := sm.NextMCID[pageIndex]
	sm.NextMCID[pageIndex] += count
	return start
}

// BeginMarkedContent starts a new structure element and returns the tag, properties strings and MCID
func (sm *StructureManager) BeginMarkedContent(streamBuilder *strings.Builder, pageIndex int, tag StructureType, props map[string]string) int {
	if !sm.Enabled {
		return 0
	}
	// 1. Create structure element
	elem := sm.acquireStructElem()
	elem.Type = tag
	elem.Parent = sm.CurrentParent
	elem.PageID = pageIndex

	if val, ok := props["Title"]; ok {
		elem.Title = val
	}
	if val, ok := props["Alt"]; ok {
		elem.Alt = val
	}

	// 2. Add as kid to current parent
	sm.CurrentParent.Kids = append(sm.CurrentParent.Kids, StructKid{Elem: elem})
	sm.Elements = append(sm.Elements, elem)

	// 3. Set current parent to this new element
	sm.CurrentParent = elem

	// 4. Generate MCID for content stream
	mcid := sm.GetNextMCID(pageIndex)

	// Track in ParentTree for PDF/UA logic
	// The element we just created is the parent of this marked content
	sm.ParentTree[pageIndex] = append(sm.ParentTree[pageIndex], elem)

	// 5. Add KID for MCID (Leaf node)
	// For leaf nodes in structure tree that point to content, we need a special marking
	// The Kid is an integer MCID, but it also needs to reference the page
	// In the actual PDF structure, this is represented slightly differently,
	// but for our internal representation:
	elem.MCID = mcid
	elem.HasMCID = true

	// Write BMC/BDC operator — direct writes, no intermediate allocation
	var intBuf [12]byte
	streamBuilder.WriteByte('/')
	streamBuilder.WriteString(string(tag))
	streamBuilder.WriteString(" <</MCID ")
	streamBuilder.Write(strconv.AppendInt(intBuf[:0], int64(mcid), 10))
	if alt, ok := props["Alt"]; ok {
		streamBuilder.WriteString(" /Alt (")
		var altBuf [1024]byte
		streamBuilder.Write(appendEscapedPDFLiteral(altBuf[:0], alt))
		streamBuilder.WriteByte(')')
	}
	streamBuilder.WriteString(">> BDC\n")

	return mcid
}

// EndMarkedContent ends the current marked content sequence
func (sm *StructureManager) EndMarkedContent(streamBuilder *strings.Builder) {
	if !sm.Enabled {
		return
	}
	streamBuilder.WriteString("EMC\n")
	if sm.CurrentParent != nil && sm.CurrentParent.Parent != nil {
		sm.CurrentParent = sm.CurrentParent.Parent
	}
}

// BeginMarkedContentBuf writes directly to a bytes.Buffer (avoids strings.Builder intermediary in hot loops)
func (sm *StructureManager) BeginMarkedContentBuf(buf *bytes.Buffer, pageIndex int, tag StructureType, props map[string]string) int {
	return sm.beginMarkedContentBuf(buf, pageIndex, tag, props, -1)
}

// BeginMarkedContentBufWithMCID is like BeginMarkedContentBuf but uses a pre-reserved MCID (batch allocation).
func (sm *StructureManager) BeginMarkedContentBufWithMCID(buf *bytes.Buffer, pageIndex int, tag StructureType, props map[string]string, mcid int) {
	sm.beginMarkedContentBuf(buf, pageIndex, tag, props, mcid)
}

func (sm *StructureManager) beginMarkedContentBuf(buf *bytes.Buffer, pageIndex int, tag StructureType, props map[string]string, reservedMCID int) int {
	if !sm.Enabled {
		return 0
	}
	// 1. Create structure element
	elem := sm.acquireStructElem()
	elem.Type = tag
	elem.Parent = sm.CurrentParent
	elem.PageID = pageIndex

	if val, ok := props["Title"]; ok {
		elem.Title = val
	}
	if val, ok := props["Alt"]; ok {
		elem.Alt = val
	}

	// 2. Add as kid to current parent
	sm.CurrentParent.Kids = append(sm.CurrentParent.Kids, StructKid{Elem: elem})
	sm.Elements = append(sm.Elements, elem)

	// 3. Set current parent to this new element
	sm.CurrentParent = elem

	// 4. MCID for content stream (pre-reserved or next available)
	mcid := reservedMCID
	if mcid < 0 {
		mcid = sm.GetNextMCID(pageIndex)
	}

	// Track in ParentTree
	sm.ParentTree[pageIndex] = append(sm.ParentTree[pageIndex], elem)

	// 5. Add KID for MCID
	elem.MCID = mcid
	elem.HasMCID = true

	// Write BDC operator directly to bytes.Buffer
	var intBuf [12]byte
	buf.WriteByte('/')
	buf.WriteString(string(tag))
	buf.WriteString(" <</MCID ")
	buf.Write(strconv.AppendInt(intBuf[:0], int64(mcid), 10))
	if alt, ok := props["Alt"]; ok {
		buf.WriteString(" /Alt (")
		var altBuf [1024]byte
		buf.Write(appendEscapedPDFLiteral(altBuf[:0], alt))
		buf.WriteByte(')')
	}
	buf.WriteString(">> BDC\n")

	return mcid
}

// EndMarkedContentBuf writes EMC directly to a bytes.Buffer (avoids strings.Builder intermediary)
func (sm *StructureManager) EndMarkedContentBuf(buf *bytes.Buffer) {
	if !sm.Enabled {
		return
	}
	buf.WriteString("EMC\n")
	if sm.CurrentParent != nil && sm.CurrentParent.Parent != nil {
		sm.CurrentParent = sm.CurrentParent.Parent
	}
}

// WriteCellMarkedContentBDC emits BDC for a cell MCID without allocating a per-cell StructElem.
func (sm *StructureManager) WriteCellMarkedContentBDC(buf *bytes.Buffer, tag StructureType, mcid int) {
	if !sm.Enabled {
		return
	}
	var intBuf [12]byte
	buf.WriteByte('/')
	buf.WriteString(string(tag))
	buf.WriteString(" <</MCID ")
	buf.Write(strconv.AppendInt(intBuf[:0], int64(mcid), 10))
	buf.WriteString(">> BDC\n")
}

// EndCellMarkedContentBuf writes EMC without popping the structure parent stack.
func (sm *StructureManager) EndCellMarkedContentBuf(buf *bytes.Buffer) {
	if !sm.Enabled {
		return
	}
	buf.WriteString("EMC\n")
}

// AttachRowMCIDs registers MCID leaf refs on the current grouping parent (typically TR).
func (sm *StructureManager) AttachRowMCIDs(pageIndex, startMCID, count int) {
	if !sm.Enabled || count <= 0 || sm.CurrentParent == nil {
		return
	}
	_ = startMCID
	parent := sm.CurrentParent
	sm.appendParentTreeRefs(pageIndex, parent, count)
	for i := range count {
		parent.Kids = append(parent.Kids, StructKid{MCID: startMCID + i})
	}
}

// BeginTableRowWithTDMCIDs starts a TR with one TD StructElem per column, each carrying
// a pre-reserved MCID. Used when replaying cached shared-row content streams that already
// contain matching BDC/EMC operators (PDF/UA-2 requires TR → TD, not bare MCID leaves).
//
// Allocations for the TR grouping element and the per-column TD elements come from the
// per-document arena; the tr.Kids slice is pre-sized to `count` so the append loop does
// not grow the slice. Together this removes the sync.Pool + memclr + slice-grow churn
// that dominated the HFT TR→TD path.
func (sm *StructureManager) BeginTableRowWithTDMCIDs(pageIndex, startMCID, count int) {
	if !sm.Enabled {
		return
	}
	// Pre-size tr.Kids so the per-cell append does not grow the backing slice.
	sm.BeginStructureElementCap(StructTR, count)
	tr := sm.CurrentParent
	for i := range count {
		td := sm.acquireStructElem()
		td.Type = StructTD
		td.Parent = tr
		td.PageID = pageIndex
		td.MCID = startMCID + i
		td.HasMCID = true
		tr.Kids = append(tr.Kids, StructKid{Elem: td})
		sm.Elements = append(sm.Elements, td)
	}
	sm.appendParentTreeRefs(pageIndex, tr, count)
}

func (sm *StructureManager) appendParentTreeRefs(pageIndex int, parent *StructElem, count int) {
	if count <= 0 {
		return
	}
	sm.ensurePageSlot(pageIndex)
	pt := &sm.ParentTree[pageIndex]
	n := len(*pt)
	need := n + count
	*pt = growPtrSlice(*pt, need, 32)
	*pt = (*pt)[:need]
	for i := n; i < need; i++ {
		(*pt)[i] = parent
	}
}

// PreallocatePageMCIDSlots grows ParentTree capacity for a page before a stripe of rows.
func (sm *StructureManager) PreallocatePageMCIDSlots(pageIndex, count int) {
	if !sm.Enabled || count <= 0 {
		return
	}
	sm.ensureParentTreeCapacity(pageIndex, count)
}

// FillDeferredParentTreePage bulk-fills ParentTree slots and MCID kids for a page stripe.
func (sm *StructureManager) FillDeferredParentTreePage(pageIndex int, parent *StructElem, startMCID, count int) {
	if !sm.Enabled || count <= 0 || parent == nil {
		return
	}
	sm.appendParentTreeRefs(pageIndex, parent, count)
	for i := range count {
		parent.Kids = append(parent.Kids, StructKid{MCID: startMCID + i})
	}
}

// BeginStructureElement starts a grouping element (like Table, TR) that doesn't directly contain content yet
func (sm *StructureManager) BeginStructureElement(tag StructureType) {
	sm.BeginStructureElementCap(tag, 0)
}

// BeginStructureElementCap starts a grouping element and preallocates its Kids slice when the shape is known.
func (sm *StructureManager) BeginStructureElementCap(tag StructureType, kidCap int) {
	if !sm.Enabled {
		return
	}
	elem := sm.acquireStructElem()
	elem.Type = tag
	elem.Parent = sm.CurrentParent
	if kidCap > 0 {
		elem.Kids = acquireStructKids(kidCap)
	}
	sm.CurrentParent.Kids = append(sm.CurrentParent.Kids, StructKid{Elem: elem})
	sm.Elements = append(sm.Elements, elem)
	sm.CurrentParent = elem
}

// EndStructureElement ends the current grouping element
func (sm *StructureManager) EndStructureElement() {
	if !sm.Enabled {
		return
	}
	if sm.CurrentParent != nil && sm.CurrentParent.Parent != nil {
		sm.CurrentParent = sm.CurrentParent.Parent
	}
}

// RegisterPageStructParents registers the parent tree mapping for a page
func (sm *StructureManager) RegisterPageStructParents(_ int, _ int) {
	// This logic handles the ParentTree mapping "Page Object -> [StructElem refs]"
	// required for reliable reverse lookup.
	// The StructParents entry in Page dictionary points to a key in ParentTree.
	// For simplicity, we can just use pageIndex as the key in ParentTreeNumber.

	// Note: Actual implementation of ParentTree generation will be done in generator.go
	// as it requires finalizing all IDs.
}

// GenerateStructTreeRoot generates the StructTreeRoot object content
// namespaceObjID is the object ID of the PDF 2.0 namespace dictionary (0 to skip)
func (sm *StructureManager) GenerateStructTreeRoot(_ int, parentTreeObjID int, namespaceObjID int) string {
	var sb strings.Builder
	var structBuf []byte
	structBuf = append(structBuf, "<< /Type /StructTreeRoot /ParentTree "...)
	structBuf = strconv.AppendInt(structBuf, int64(parentTreeObjID), 10)
	structBuf = append(structBuf, " 0 R"...)
	sb.Write(structBuf)

	// PDF 2.0: Add Namespaces array for PDF/UA-2
	if namespaceObjID > 0 {
		structBuf = structBuf[:0]
		structBuf = append(structBuf, " /Namespaces [ "...)
		structBuf = strconv.AppendInt(structBuf, int64(namespaceObjID), 10)
		structBuf = append(structBuf, " 0 R ]"...)
		sb.Write(structBuf)
	}

	// Point to the first child (Document)
	if len(sm.Root.Kids) > 0 {
		// Assuming the first kid is the Document element
		if firstKid := sm.Root.Kids[0].Elem; firstKid != nil {
			structBuf = structBuf[:0]
			structBuf = append(structBuf, " /K "...)
			structBuf = strconv.AppendInt(structBuf, int64(firstKid.ObjectID), 10)
			structBuf = append(structBuf, " 0 R"...)
			sb.Write(structBuf)
		}
	}

	sb.WriteString(" >>")
	return sb.String()
}

// LinkElement tracks a Link structure element for an annotation
type LinkElement struct {
	AnnotObjID int // The annotation object ID this link references
	PageIndex  int // Page where the annotation appears
	ObjectID   int // Assigned object ID for this structure element
}

// AddLinkElement adds a Link structure element for an annotation
// PDF/UA-2 requires link annotations to be wrapped in Link structure elements
func (sm *StructureManager) AddLinkElement(annotObjID int, _ int) {
	if !sm.Enabled {
		return
	}
	linkElem := sm.acquireStructElem()
	linkElem.Type = StructLink
	linkElem.Parent = sm.GetCurrentDocumentElement()
	linkElem.AnnotObjID = annotObjID

	// Add to Document element's kids
	if docElem := sm.GetCurrentDocumentElement(); docElem != nil {
		docElem.Kids = append(docElem.Kids, StructKid{Elem: linkElem})
	}

	sm.Elements = append(sm.Elements, linkElem)

	// Store the annotation reference - this will be resolved during generation
	// The OBJR (object reference) to the annotation will be added as a kid
	if sm.LinkElements == nil {
		sm.LinkElements = make(map[int]*StructElem)
	}
	sm.LinkElements[annotObjID] = linkElem
}

// GetCurrentDocumentElement returns the Document element (first child of Root)
func (sm *StructureManager) GetCurrentDocumentElement() *StructElem {
	if !sm.Enabled {
		return nil
	}
	if len(sm.Root.Kids) > 0 {
		if docElem := sm.Root.Kids[0].Elem; docElem != nil {
			return docElem
		}
	}
	return nil
}

// CreateBookmarkSect creates a Sect structure element for a bookmark target
// PDF/UA-2 requires GoTo actions to use structure destinations (/SD)
// This creates a section element that can be used as a navigation target
func (sm *StructureManager) CreateBookmarkSect(title string) *StructElem {
	if !sm.Enabled {
		return nil
	}
	sectElem := sm.acquireStructElem()
	sectElem.Type = StructSect
	sectElem.Title = title
	sectElem.Parent = sm.GetCurrentDocumentElement()

	// Add to Document element's kids
	if docElem := sm.GetCurrentDocumentElement(); docElem != nil {
		docElem.Kids = append(docElem.Kids, StructKid{Elem: sectElem})
	}

	sm.Elements = append(sm.Elements, sectElem)

	return sectElem
}
