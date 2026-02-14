package pdf

import (
	"bytes"
	"strconv"
	"strings"
)

// StructureType represents the standard structure types in PDF/UA
type StructureType string

const (
	StructDocument  StructureType = "Document"
	StructPart      StructureType = "Part"
	StructSect      StructureType = "Sect"
	StructDiv       StructureType = "Div"
	StructH1        StructureType = "H1"
	StructH2        StructureType = "H2"
	StructH3        StructureType = "H3"
	StructH4        StructureType = "H4"
	StructH5        StructureType = "H5"
	StructH6        StructureType = "H6"
	StructP         StructureType = "P"
	StructL         StructureType = "L"     // List
	StructLI        StructureType = "LI"    // List Item
	StructLbl       StructureType = "Lbl"   // List Label
	StructLBody     StructureType = "LBody" // List Body
	StructTable     StructureType = "Table"
	StructTR        StructureType = "TR"
	StructTH        StructureType = "TH"
	StructTD        StructureType = "TD"
	StructFigure    StructureType = "Figure"
	StructCaption   StructureType = "Caption"
	StructForm      StructureType = "Form"
	StructLink      StructureType = "Link"      // PDF/UA-2: Link annotation wrapper
	StructReference StructureType = "Reference" // PDF/UA-2: Reference structure
)

// StructElem represents a node in the structure tree
type StructElem struct {
	Type     StructureType
	Title    string
	Alt      string
	Lang     string
	Kids     []interface{} // Can be *StructElem or int (MCID reference object)
	Parent   *StructElem
	ObjectID int // Assigned when writing to PDF
	PageID   int // Reference to the page object ID where this element appears
}

// StructureManager handles the creation and management of the PDF structure tree
type StructureManager struct {
	Root          *StructElem
	CurrentParent *StructElem
	Elements      []*StructElem
	NextMCID      map[int]int           // Page Index -> Next MCID
	ParentTree    map[int][]*StructElem // Page Index -> Array of struct elements (parents of MCIDs)
	StructParents map[int]int           // Page Object ID -> StructParents index (in ParentTree)
	LinkElements  map[int]*StructElem   // PDF/UA-2: Annotation Object ID -> Link StructElem
}

// NewStructureManager creates a new structure manager
func NewStructureManager() *StructureManager {
	// Root is the conceptual StructTreeRoot container (hidden)
	root := &StructElem{
		Type: "Root",
	}
	sm := &StructureManager{
		Root:          root,
		CurrentParent: root,
		Elements:      []*StructElem{root},
		NextMCID:      make(map[int]int),
		ParentTree:    make(map[int][]*StructElem),
		StructParents: make(map[int]int),
	}

	// PDF/UA: The valid top-level element must be 'Document' (or Part, Art, etc.)
	// We automatically start the 'Document' element here.
	// All subsequent content will be children of this Document element.
	sm.BeginStructureElement(StructDocument)

	return sm
}

// GetNextMCID returns the next available MCID for a page
func (sm *StructureManager) GetNextMCID(pageIndex int) int {
	mcid := sm.NextMCID[pageIndex]
	sm.NextMCID[pageIndex]++
	return mcid
}

// BeginMarkedContent starts a new structure element and returns the tag, properties strings and MCID
func (sm *StructureManager) BeginMarkedContent(streamBuilder *strings.Builder, pageIndex int, tag StructureType, props map[string]string) int {
	// 1. Create structure element
	elem := &StructElem{
		Type:   tag,
		Parent: sm.CurrentParent,
		PageID: pageIndex,
	}

	if val, ok := props["Title"]; ok {
		elem.Title = val
	}
	if val, ok := props["Alt"]; ok {
		elem.Alt = val
	}

	// 2. Add as kid to current parent
	sm.CurrentParent.Kids = append(sm.CurrentParent.Kids, elem)
	sm.Elements = append(sm.Elements, elem)

	// 3. Set current parent to this new element
	sm.CurrentParent = elem

	// 4. Generate MCID for content stream
	mcid := sm.GetNextMCID(pageIndex)

	// Track in ParentTree for PDF/UA logic
	// The element we just created is the parent of this marked content
	if sm.ParentTree[pageIndex] == nil {
		sm.ParentTree[pageIndex] = make([]*StructElem, 0)
	}
	sm.ParentTree[pageIndex] = append(sm.ParentTree[pageIndex], elem)

	// 5. Add KID for MCID (Leaf node)
	// For leaf nodes in structure tree that point to content, we need a special marking
	// The Kid is an integer MCID, but it also needs to reference the page
	// In the actual PDF structure, this is represented slightly differently,
	// but for our internal representation:
	elem.Kids = append(elem.Kids, mcid)

	// Write BMC/BDC operator — direct writes, no intermediate allocation
	var intBuf [12]byte
	streamBuilder.WriteByte('/')
	streamBuilder.WriteString(string(tag))
	streamBuilder.WriteString(" <</MCID ")
	streamBuilder.Write(strconv.AppendInt(intBuf[:0], int64(mcid), 10))
	if alt, ok := props["Alt"]; ok {
		streamBuilder.WriteString(" /Alt (")
		streamBuilder.WriteString(escapeText(alt))
		streamBuilder.WriteByte(')')
	}
	streamBuilder.WriteString(">> BDC\n")

	return mcid
}

// EndMarkedContent ends the current marked content sequence
func (sm *StructureManager) EndMarkedContent(streamBuilder *strings.Builder) {
	streamBuilder.WriteString("EMC\n")
	if sm.CurrentParent != nil && sm.CurrentParent.Parent != nil {
		sm.CurrentParent = sm.CurrentParent.Parent
	}
}

// BeginMarkedContentBuf writes directly to a bytes.Buffer (avoids strings.Builder intermediary in hot loops)
func (sm *StructureManager) BeginMarkedContentBuf(buf *bytes.Buffer, pageIndex int, tag StructureType, props map[string]string) int {
	// 1. Create structure element
	elem := &StructElem{
		Type:   tag,
		Parent: sm.CurrentParent,
		PageID: pageIndex,
	}

	if val, ok := props["Title"]; ok {
		elem.Title = val
	}
	if val, ok := props["Alt"]; ok {
		elem.Alt = val
	}

	// 2. Add as kid to current parent
	sm.CurrentParent.Kids = append(sm.CurrentParent.Kids, elem)
	sm.Elements = append(sm.Elements, elem)

	// 3. Set current parent to this new element
	sm.CurrentParent = elem

	// 4. Generate MCID for content stream
	mcid := sm.GetNextMCID(pageIndex)

	// Track in ParentTree
	if sm.ParentTree[pageIndex] == nil {
		sm.ParentTree[pageIndex] = make([]*StructElem, 0)
	}
	sm.ParentTree[pageIndex] = append(sm.ParentTree[pageIndex], elem)

	// 5. Add KID for MCID
	elem.Kids = append(elem.Kids, mcid)

	// Write BDC operator directly to bytes.Buffer
	var intBuf [12]byte
	buf.WriteByte('/')
	buf.WriteString(string(tag))
	buf.WriteString(" <</MCID ")
	buf.Write(strconv.AppendInt(intBuf[:0], int64(mcid), 10))
	if alt, ok := props["Alt"]; ok {
		buf.WriteString(" /Alt (")
		buf.WriteString(escapeText(alt))
		buf.WriteByte(')')
	}
	buf.WriteString(">> BDC\n")

	return mcid
}

// EndMarkedContentBuf writes EMC directly to a bytes.Buffer (avoids strings.Builder intermediary)
func (sm *StructureManager) EndMarkedContentBuf(buf *bytes.Buffer) {
	buf.WriteString("EMC\n")
	if sm.CurrentParent != nil && sm.CurrentParent.Parent != nil {
		sm.CurrentParent = sm.CurrentParent.Parent
	}
}

// BeginStructureElement starts a grouping element (like Table, TR) that doesn't directly contain content yet
func (sm *StructureManager) BeginStructureElement(tag StructureType) {
	elem := &StructElem{
		Type:   tag,
		Parent: sm.CurrentParent,
	}
	sm.CurrentParent.Kids = append(sm.CurrentParent.Kids, elem)
	sm.Elements = append(sm.Elements, elem)
	sm.CurrentParent = elem
}

// EndStructureElement ends the current grouping element
func (sm *StructureManager) EndStructureElement() {
	if sm.CurrentParent != nil && sm.CurrentParent.Parent != nil {
		sm.CurrentParent = sm.CurrentParent.Parent
	}
}

// RegisterPageStructParents registers the parent tree mapping for a page
func (sm *StructureManager) RegisterPageStructParents(pageObjectID int, pageIndex int) {
	// This logic handles the ParentTree mapping "Page Object -> [StructElem refs]"
	// required for reliable reverse lookup.
	// The StructParents entry in Page dictionary points to a key in ParentTree.
	// For simplicity, we can just use pageIndex as the key in ParentTreeNumber.

	// Note: Actual implementation of ParentTree generation will be done in generator.go
	// as it requires finalizing all IDs.
}

// GenerateStructTreeRoot generates the StructTreeRoot object content
// namespaceObjID is the object ID of the PDF 2.0 namespace dictionary (0 to skip)
func (sm *StructureManager) GenerateStructTreeRoot(rootObjID int, parentTreeObjID int, namespaceObjID int) string {
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
		if firstKid, ok := sm.Root.Kids[0].(*StructElem); ok {
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
func (sm *StructureManager) AddLinkElement(annotObjID int, pageIndex int) {
	// Create a Link structure element
	linkElem := &StructElem{
		Type:   StructLink,
		Parent: sm.GetCurrentDocumentElement(),
	}

	// Add to Document element's kids
	if docElem := sm.GetCurrentDocumentElement(); docElem != nil {
		docElem.Kids = append(docElem.Kids, linkElem)
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
	if len(sm.Root.Kids) > 0 {
		if docElem, ok := sm.Root.Kids[0].(*StructElem); ok {
			return docElem
		}
	}
	return nil
}

// CreateBookmarkSect creates a Sect structure element for a bookmark target
// PDF/UA-2 requires GoTo actions to use structure destinations (/SD)
// This creates a section element that can be used as a navigation target
func (sm *StructureManager) CreateBookmarkSect(title string) *StructElem {
	sectElem := &StructElem{
		Type:   StructSect,
		Title:  title,
		Parent: sm.GetCurrentDocumentElement(),
	}

	// Add to Document element's kids
	if docElem := sm.GetCurrentDocumentElement(); docElem != nil {
		docElem.Kids = append(docElem.Kids, sectElem)
	}

	sm.Elements = append(sm.Elements, sectElem)

	return sectElem
}
