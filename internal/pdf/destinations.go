package pdf

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// DestinationStore owns named destinations, outline ID allocation, link
// annotations, and link structure elements (Phase 5 D4). Callers use
// Define to register a destination, Resolve/Lookup to read it back, and
// Emit to write the Dests name tree. OutlineBuilder and links.go delegate
// here so the DestKey invariant lives in one home.
type DestinationStore struct {
	pm        *PageManager
	encryptor ObjectEncryptor
}

// NewDestinationStore creates a destination store bound to a page manager.
func NewDestinationStore(pm *PageManager, encryptor ObjectEncryptor) *DestinationStore {
	return &DestinationStore{pm: pm, encryptor: encryptor}
}

// Define registers a named destination for internal linking.
func (d *DestinationStore) Define(name string, pageIndex int, y float64) {
	d.pm.NamedDests[name] = NamedDest{
		PageIndex: pageIndex,
		Y:         y,
	}
}

// DefineFull registers a fully populated named destination.
func (d *DestinationStore) DefineFull(name string, dest NamedDest) {
	d.pm.NamedDests[name] = dest
}

// Lookup returns a named destination by name.
func (d *DestinationStore) Lookup(name string) (NamedDest, bool) {
	dest, ok := d.pm.NamedDests[name]
	return dest, ok
}

// UpdateStructElemID attaches a structure element ID to an existing
// named destination (PDF/UA-2 structure destinations for bookmarks).
func (d *DestinationStore) UpdateStructElemID(name string, structElemID int) {
	if existing, ok := d.pm.NamedDests[name]; ok {
		existing.StructElemID = structElemID
		d.pm.NamedDests[name] = existing
	}
}

// AllocID reserves one object ID for outline or name-tree objects.
func (d *DestinationStore) AllocID() int {
	return d.pm.AllocObjectID()
}

// Commit stores an outline or name-tree object body under a reserved ID.
func (d *DestinationStore) Commit(id int, content []byte) {
	d.pm.CommitExtraObject(id, content)
}

// ResolvePageObjectID maps a named destination to its page object ID,
// falling back to the first page when the index is out of range.
func (d *DestinationStore) ResolvePageObjectID(dest NamedDest) int {
	if dest.PageIndex < len(d.pm.Pages) {
		return d.pm.Pages[dest.PageIndex]
	}
	if len(d.pm.Pages) > 0 {
		return d.pm.Pages[0]
	}
	return 0
}

// EmitNameTree builds the Dests name tree plus the wrapping Names
// dictionary, commits both as extra objects, and returns the Names
// object ID. It returns false when no named destinations exist.
func (d *DestinationStore) EmitNameTree() (int, bool) {
	if len(d.pm.NamedDests) == 0 {
		return 0, false
	}

	// Build Names array for Dests name tree
	var namesArray bytes.Buffer
	namesArray.WriteString("[")

	// Sort names for binary search tree compliance
	names := make([]string, 0, len(d.pm.NamedDests))
	for name := range d.pm.NamedDests {
		names = append(names, name)
	}
	sort.Strings(names)

	// Create Dests name tree object ID upfront for encryption key generation
	destsTreeID := d.pm.AllocObjectID()

	var numBuf [20]byte
	for i, name := range names {
		dest := d.pm.NamedDests[name]
		pageObjID := d.ResolvePageObjectID(dest)
		if i > 0 {
			namesArray.WriteString(" ")
		}

		// Handle Name encryption
		nameStr := ""
		if d.encryptor != nil {
			// Names in name tree are strings and must be encrypted.
			// Copy into a fresh slice first: EncryptString may pad in
			// place, which must never mutate immutable string bytes.
			nameBytes := append([]byte(nil), name...)
			encrypted := d.encryptor.EncryptString(nameBytes, destsTreeID, 0)
			nameStr = "<" + hex.EncodeToString(encrypted) + ">"
		} else {
			nameStr = "(" + escapeText(name) + ")"
		}

		// PDF/UA-2: Output as dictionary with both /D and /SD keys
		// /D is the page-based destination (for compatibility)
		// /SD is the structure destination (required for PDF/UA-2)
		if dest.StructElemID > 0 {
			namesArray.WriteString(nameStr)
			namesArray.WriteString(" << /D [")
			namesArray.Write(strconv.AppendInt(numBuf[:0], int64(pageObjID), 10))
			namesArray.WriteString(" 0 R /XYZ null ")
			namesArray.Write(appendFmtNum(numBuf[:0], dest.Y))
			namesArray.WriteString(" null] /SD [")
			namesArray.Write(strconv.AppendInt(numBuf[:0], int64(dest.StructElemID), 10))
			namesArray.WriteString(" 0 R /XYZ null ")
			namesArray.Write(appendFmtNum(numBuf[:0], dest.Y))
			namesArray.WriteString(" null] >>")
		} else {
			namesArray.WriteString(nameStr)
			namesArray.WriteString(" [")
			namesArray.Write(strconv.AppendInt(numBuf[:0], int64(pageObjID), 10))
			namesArray.WriteString(" 0 R /XYZ null ")
			namesArray.Write(appendFmtNum(numBuf[:0], dest.Y))
			namesArray.WriteString(" null]")
		}
	}
	namesArray.WriteString("]")

	destsTreeContent := "<< /Names " + namesArray.String() + " >>"
	d.pm.CommitExtraObject(destsTreeID, append([]byte(nil), destsTreeContent...))

	// Create Names dictionary object
	namesID := d.pm.AllocObjectID()

	namesContent := fmt.Appendf(nil, "<< /Dests %d 0 R >>", destsTreeID)
	d.pm.CommitExtraObject(namesID, namesContent)

	return namesID, true
}

// CreateLinkAnnotation creates a PDF link annotation object with PDF/UA-2 structure
// For external links (URLs), it creates a /URI action
// For internal links (bookmarks), it creates a /GoTo action with named destination
// Returns the annotation object ID
func (d *DestinationStore) CreateLinkAnnotation(annot LinkAnnotation) int {
	pm := d.pm
	var annotDict strings.Builder
	var structParentIdx int

	annotDict.WriteString("<< /Type /Annot /Subtype /Link")
	annotDict.WriteString(fmt.Sprintf(" /Rect [%s %s %s %s]",
		fmtNum(annot.Rect[0]), fmtNum(annot.Rect[1]),
		fmtNum(annot.Rect[2]), fmtNum(annot.Rect[3])))

	// Border style - no visible border (0 0 0 means no border)
	annotDict.WriteString(" /Border [0 0 0]")

	// Highlight mode - invert when clicked
	annotDict.WriteString(" /H /I")

	// PDF/A-4 compliance: F key is required
	// Flag 4 = Print. This ensures the annotation is considered printable,
	// satisfying the requirement that all non-popup annotations must have an F key.
	annotDict.WriteString(" /F 4")

	// PDF/UA-2: StructParent links annotation to structure tree
	if pm.Structure.Enabled {
		structParentIdx = pm.GetNextAnnotStructParent()
		annotDict.WriteString(fmt.Sprintf(" /StructParent %d", structParentIdx))
	}

	// Add action based on link type
	switch {
	case annot.URI != "":
		// External URL - use URI action
		annotDict.WriteString(fmt.Sprintf(" /A << /Type /Action /S /URI /URI (%s) >>",
			escapeText(annot.URI)))
	case annot.Dest != "":
		// Internal link - use named destination
		// PDF/UA-2: Use /Dest (name) - the named destination contains both /D and /SD
		annotDict.WriteString(fmt.Sprintf(" /Dest (%s)", escapeText(annot.Dest)))
	case annot.PageIndex >= 0:
		// Internal link with explicit page destination
		// Format: [pageRef /XYZ left top zoom]
		// XYZ = position at (left, top) with zoom factor.
		// Page IDs live in pageManager.Pages; fall back to the first page
		// when the index is out of range instead of assuming 3+index.
		pageIdx := annot.PageIndex
		if pageIdx < 0 || pageIdx >= len(pm.Pages) {
			pageIdx = 0
		}
		pageObjID := 3 // First page starts at object 3
		if len(pm.Pages) > 0 {
			pageObjID = pm.Pages[pageIdx]
		}
		annotDict.WriteString(fmt.Sprintf(" /Dest [%d 0 R /XYZ null %s null]",
			pageObjID, fmtNum(annot.DestY)))
	}

	annotDict.WriteString(" >>")

	objID := pm.AddExtraObject(annotDict.String())

	if pm.Structure.Enabled {
		// PDF/UA-2: Link structure element that references this annotation
		pm.AddLinkStructureElement(objID, structParentIdx)
	}

	return objID
}
