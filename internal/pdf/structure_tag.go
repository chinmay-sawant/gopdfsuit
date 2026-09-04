package pdf

import "bytes"

// TableTagger owns per-row MCID reservation, TD BDC bytes, and ParentTree
// appends for tagged tables (Phase 5 D4). draw.go table paths call
// BeginRow/WriteCell/EndRow instead of touching StructureManager MCID
// internals directly, so the BDC/MCID accounting lives in one home.
type TableTagger struct {
	sm *StructureManager
}

// NewTableTagger creates a tagger bound to a structure manager.
func NewTableTagger(sm *StructureManager) *TableTagger {
	return &TableTagger{sm: sm}
}

// BeginRow reserves cellCount consecutive MCIDs, creates the TR grouping
// element with one TD leaf per column, and returns the base MCID.
// It is a no-op returning 0 when tagging is disabled.
func (t *TableTagger) BeginRow(pageIndex, cellCount int) int {
	if t == nil || t.sm == nil || !t.sm.Enabled || cellCount <= 0 {
		return 0
	}
	base := t.sm.ReserveMCIDsLite(pageIndex, cellCount)
	t.sm.beginTableRowWithTDMCIDs(pageIndex, base, cellCount)
	return base
}

// Reserve allocates cellCount consecutive MCIDs on a page and returns the
// base ID without creating structure elements yet.
func (t *TableTagger) Reserve(pageIndex, cellCount int) int {
	if t == nil || t.sm == nil || !t.sm.Enabled || cellCount <= 0 {
		return 0
	}
	return t.sm.ReserveMCIDsLite(pageIndex, cellCount)
}

// BeginRowWithBase creates the TR/TD structure elements for a
// previously reserved MCID base (e.g. a cached shared-row replay).
func (t *TableTagger) BeginRowWithBase(pageIndex, base, cellCount int) {
	if t == nil || t.sm == nil || !t.sm.Enabled || cellCount <= 0 {
		return
	}
	t.sm.beginTableRowWithTDMCIDs(pageIndex, base, cellCount)
}

// WriteCell emits the TD BDC operator for one cell MCID.
func (t *TableTagger) WriteCell(buf *bytes.Buffer, mcid int) {
	if t == nil || t.sm == nil || !t.sm.Enabled {
		return
	}
	t.sm.writeCellMarkedContentBDC(buf, StructTD, mcid)
}

// WriteCellWithTag emits the BDC operator for one cell MCID with an
// explicit structure tag (TH for header cells, TD otherwise).
func (t *TableTagger) WriteCellWithTag(buf *bytes.Buffer, tag StructureType, mcid int) {
	if t == nil || t.sm == nil || !t.sm.Enabled {
		return
	}
	t.sm.writeCellMarkedContentBDC(buf, tag, mcid)
}

// EndRow closes the current TR grouping element.
func (t *TableTagger) EndRow() {
	if t == nil || t.sm == nil || !t.sm.Enabled {
		return
	}
	t.sm.EndStructureElement()
}

// MCIDCursor returns the next unassigned MCID for a page, for accounting
// checks without reserving.
func (t *TableTagger) MCIDCursor(pageIndex int) int {
	if t == nil || t.sm == nil || !t.sm.Enabled {
		return 0
	}
	return t.sm.PageMCIDStart(pageIndex)
}
