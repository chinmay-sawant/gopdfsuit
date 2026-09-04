package pdf

import "fmt"

// Allocator is the single seam for object-ID reservation, ExtraObjects
// commit, and xref-offset recording during generation (Phase 5 D1).
//
// It is bound-only: it wraps a live PageManager plus the generation xref
// offsets slice, so every reservation flows through one counter. There is no
// standalone mode; unit tests bind a real PageManager (see allocator_test.go
// and metadata_test.go). When offsets is nil (tests, lazy binds), offsets
// are recorded in a private slice instead of the generation table.
type Allocator struct {
	pm      *PageManager
	offsets *[]int
	own     []int
}

// BindPageManager binds the allocator to live generation state.
func (a *Allocator) BindPageManager(pm *PageManager, offsets *[]int) *Allocator {
	a.pm = pm
	a.offsets = offsets
	return a
}

// Next peeks at the next object ID without reserving it.
func (a *Allocator) Next() int {
	return a.pm.NextObjectID
}

// Alloc reserves one object ID.
func (a *Allocator) Alloc() int {
	return a.AllocN(1)
}

// AllocN reserves n consecutive object IDs and returns the first.
func (a *Allocator) AllocN(n int) int {
	if n <= 0 {
		n = 1
	}
	id := a.pm.NextObjectID
	a.pm.NextObjectID += n
	return id
}

// SeekTo sets the next object ID to id. Generation fixups use it to align
// the counter with externally assigned ID blocks (image deduper, font
// registry) instead of writing the counter field directly.
func (a *Allocator) SeekTo(id int) {
	a.pm.NextObjectID = id
}

// EnsureBeyond advances the counter to id when it lags behind a
// caller-computed high-water mark (e.g. the dense font block). It never
// moves the counter backwards.
func (a *Allocator) EnsureBeyond(id int) {
	if a.pm.NextObjectID < id {
		a.pm.NextObjectID = id
	}
}

// Commit stores an extra object body under a reserved ID.
func (a *Allocator) Commit(id int, content []byte) {
	a.pm.ExtraObjects[id] = content
}

// CommitString stores an extra object body from a string.
func (a *Allocator) CommitString(id int, content string) {
	a.Commit(id, []byte(content))
}

// Lookup returns a committed extra object body.
func (a *Allocator) Lookup(id int) ([]byte, bool) {
	b, ok := a.pm.ExtraObjects[id]
	return b, ok
}

// SetOffset records the byte offset of an object for the xref table.
func (a *Allocator) SetOffset(id, offset int) {
	if a.offsets != nil {
		setXrefOffset(a.offsets, id, offset)
		return
	}
	growXrefOffsets(&a.own, id)
	a.own[id] = offset
}

// Offset returns a recorded xref offset.
func (a *Allocator) Offset(id int) (int, bool) {
	if a.offsets != nil {
		return xrefOffsetAt(*a.offsets, id)
	}
	return xrefOffsetAt(a.own, id)
}

// LayoutContentFontIDs assigns the content-stream and std-font object ID
// blocks for a document with totalPages pages.
//
// Pages are dense from object 3 while allocator-backed extras (images,
// custom fonts, outlines, annotations) occupy [extraRegionBase, nextID).
// The dense layout [totalPages+3, ...) only fits while it ends before that
// region; for very large documents the whole content/font block is shifted
// above the extras instead of overlapping them and corrupting the xref.
// Fail closed: once page IDs themselves reach the extras region no valid
// layout exists, so an error is returned instead of a corrupt PDF.
func (a *Allocator) LayoutContentFontIDs(totalPages, extraRegionBase int) (contentStart, fontStart int, err error) {
	return layoutContentFontIDs(totalPages, extraRegionBase, a.Next())
}

// layoutContentFontIDs is the pure function behind Allocator.LayoutContentFontIDs.
// It stays package-visible for existing layout tests.
func layoutContentFontIDs(totalPages, extraRegionBase, nextID int) (contentStart, fontStart int, err error) {
	if totalPages+3 > extraRegionBase {
		return 0, 0, fmt.Errorf("document has %d pages, exceeding the supported layout range", totalPages)
	}
	contentStart = totalPages + 3         // Content objects start after pages
	fontStart = contentStart + totalPages // Fonts start after content
	if fontStart+maxLowRegionFontObjects > extraRegionBase {
		contentStart = nextID
		fontStart = contentStart + totalPages
	}
	return contentStart, fontStart, nil
}
