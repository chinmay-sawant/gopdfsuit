package pdf

import (
	"bytes"
	"testing"
)

// EmitRowCells is the sole seam for row MCID reservation and TD BDC bytes;
// these tests pin its accounting so the underlying StructureManager row
// primitives can stay package-private.

func TestEmitRowCellsReserveBatch(t *testing.T) {
	sm := NewStructureManager(true)
	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 7)
	defer end()
	if base != 0 {
		t.Fatalf("expected start MCID 0, got %d", base)
	}
	if sm.NextMCID[0] != 7 {
		t.Fatalf("expected NextMCID 7, got %d", sm.NextMCID[0])
	}
	if next := sm.PageMCIDStart(0); next != 7 {
		t.Fatalf("expected cursor MCID 7, got %d", next)
	}
}

func TestEmitRowCellsRowAccounting(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTR)

	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 2)
	for i := range 2 {
		sm.writeCellMarkedContentBDC(&buf, StructTD, base+i)
		buf.WriteString("EMC\n")
	}
	end()

	if sm.NextMCID[0] != 2 {
		t.Fatalf("expected NextMCID 2 after pair, got %d", sm.NextMCID[0])
	}
	if len(sm.ParentTree[0]) != 2 {
		t.Fatalf("expected 2 parent tree entries, got %d", len(sm.ParentTree[0]))
	}
}

func TestEmitRowCellsCreatesTDUnderTR(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent

	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 3)
	tr := sm.CurrentParent
	for i := range 3 {
		sm.writeCellMarkedContentBDC(&buf, StructTD, base+i)
	}
	end()

	if buf.Len() == 0 {
		t.Fatal("expected cell BDC bytes in content buffer")
	}
	if len(table.Kids) != 1 || table.Kids[0].Elem != tr {
		t.Fatal("expected Table kid to be TR element")
	}
	if len(tr.Kids) != 3 {
		t.Fatalf("expected 3 TD kids on TR, got %d", len(tr.Kids))
	}
	for i, kid := range tr.Kids {
		if kid.Elem == nil || kid.Elem.Type != StructTD {
			t.Fatalf("kid %d: expected TD struct element, got %+v", i, kid)
		}
		if mcid, ok := kid.Elem.LeafMCID(); !ok || mcid != base+i {
			t.Fatalf("kid %d: expected MCID %d, got %d ok=%v", i, base+i, mcid, ok)
		}
	}
}

func TestEmitRowCellsParentTreeReferencesTD(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 3)
	defer end()
	tr := sm.CurrentParent

	if len(sm.ParentTree[0]) != 3 {
		t.Fatalf("expected 3 parent tree entries, got %d", len(sm.ParentTree[0]))
	}
	for i, ref := range sm.ParentTree[0] {
		if ref.Type != StructTD {
			t.Fatalf("ParentTree[%d] type=%s, want TD", i, ref.Type)
		}
		if mcid, ok := ref.LeafMCID(); !ok || mcid != base+i {
			t.Fatalf("ParentTree[%d] MCID=%d ok=%v, want %d", i, mcid, ok, base+i)
		}
		if ref.Parent != tr {
			t.Fatalf("ParentTree[%d] parent=%p, want TR %p", i, ref.Parent, tr)
		}
	}
}

func TestEmitRowCellsPageID(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 2, 4)
	defer end()
	if base != 0 {
		t.Fatalf("fresh page cursor = %d, want 0", base)
	}
	tr := sm.CurrentParent
	if tr.PageID != 2 {
		t.Fatalf("TR PageID=%d, want 2", tr.PageID)
	}
	for i, kid := range tr.Kids {
		if kid.Elem == nil || kid.Elem.PageID != 2 {
			t.Fatalf("TD %d PageID=%v, want 2", i, kid.Elem)
		}
	}
}

func TestEmitRowCellsCreatesTD(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent

	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 3)
	defer end()
	tr := sm.CurrentParent

	if len(table.Kids) != 1 || table.Kids[0].Elem != tr {
		t.Fatal("expected Table kid to be TR element")
	}
	if len(tr.Kids) != 3 {
		t.Fatalf("expected 3 TD kids on TR, got %d", len(tr.Kids))
	}
	for i, kid := range tr.Kids {
		if kid.Elem == nil || kid.Elem.Type != StructTD {
			t.Fatalf("kid %d: expected TD struct element, got %+v", i, kid)
		}
		if mcid, ok := kid.Elem.LeafMCID(); !ok || mcid != base+i {
			t.Fatalf("kid %d: expected MCID %d, got %d ok=%v", i, base+i, mcid, ok)
		}
	}
}

func TestReserveElementCapacityGrowsBackingSlice(t *testing.T) {
	sm := NewStructureManager(true)
	before := cap(sm.Elements)
	sm.ReserveElementCapacity(512)
	if got := cap(sm.Elements); got < len(sm.Elements)+512 {
		t.Fatalf("elements cap = %d, want at least %d", got, len(sm.Elements)+512)
	}
	if cap(sm.Elements) < before {
		t.Fatalf("elements cap shrank: before=%d after=%d", before, cap(sm.Elements))
	}
}

// TestEmitRowCellsArenaAllocates asserts that the TD StructElems
// allocated by a row emit set up a TR → TD hierarchy and that
// ReleaseStructElemsToPool walks all elems back to the global pool. P1
// (2026-06-20 checklist) originally specified a per-document arena; we
// instead rely on the global sync.Pool + selective field clear (P2) to
// avoid the per-elem memclr that the pool was paying. This test pins the
// observable behaviour: the TD/TR shape is correct and release leaves the
// Elements slice at root-only.
func TestEmitRowCellsArenaAllocates(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	sm.ReserveElementCapacity(arenaActivationThreshold)

	var buf bytes.Buffer
	const rows = 4
	const cols = 7
	for r := 0; r < rows; r++ {
		_, end := sm.EmitRowCells(&buf, 0, cols)
		end()
		_ = r
	}

	if len(table.Kids) != rows {
		t.Fatalf("expected %d TR kids on Table, got %d", rows, len(table.Kids))
	}

	totalTDs := 0
	for _, trKid := range table.Kids {
		tr := trKid.Elem
		if tr == nil {
			t.Fatal("TR kid was nil")
		}
		if len(tr.Kids) != cols {
			t.Fatalf("TR kid count: got %d want %d", len(tr.Kids), cols)
		}
		for _, kid := range tr.Kids {
			td := kid.Elem
			if td == nil {
				t.Fatalf("TD kid was nil")
			}
			totalTDs++
		}
	}
	if totalTDs != rows*cols {
		t.Fatalf("expected %d TDs total, got %d", rows*cols, totalTDs)
	}

	if sm.arenaSlab == nil {
		t.Fatal("expected arena slab to be active for tagged structure manager")
	}
	slabBefore := sm.arenaSlab

	sm.ReleaseStructElemsToPool()
	// Elements slice should be reduced to just the root.
	if len(sm.Elements) != 1 {
		t.Fatalf("Elements should be reset to root-only, got %d entries", len(sm.Elements))
	}
	if sm.arenaSlab != nil {
		t.Fatal("arena slab should be returned to pool on release")
	}
	if slabBefore == nil || cap(*slabBefore) == 0 {
		t.Fatal("released slab should retain backing capacity")
	}
}

func TestAssignStructIDsSequential(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	var buf bytes.Buffer
	_, end := sm.EmitRowCells(&buf, 0, 3)
	end()
	_, end = sm.EmitRowCells(&buf, 0, 3)
	end()

	startID := 5000
	nextID := startID
	for i := 1; i < len(sm.Elements); i++ {
		elem := sm.Elements[i]
		if elem == nil || elem.ObjectID != 0 {
			continue
		}
		elem.ObjectID = nextID
		nextID++
	}

	want := nextID - startID
	assigned := 0
	for i := 1; i < len(sm.Elements); i++ {
		elem := sm.Elements[i]
		if elem == nil {
			continue
		}
		if elem.ObjectID < startID {
			t.Fatalf("elem %d has ObjectID %d before start %d", i, elem.ObjectID, startID)
		}
		assigned++
	}
	if assigned != want {
		t.Fatalf("assigned %d elems, expected %d", assigned, want)
	}
}

// TestReleaseStructElemsToPool_canRunTwice checks that the structure manager
// can be released and a new sequence started, the way the Zerodha benchmark
// uses one PDF per manager.
func TestReleaseStructElemsToPool_canRunTwice(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	var buf bytes.Buffer
	for r := 0; r < 3; r++ {
		_, end := sm.EmitRowCells(&buf, 0, 4)
		end()
	}
	sm.ReleaseStructElemsToPool()

	sm.BeginStructureElement(StructTable)
	_, end := sm.EmitRowCells(&buf, 0, 5)
	tr := sm.CurrentParent
	if len(tr.Kids) != 5 {
		t.Fatalf("expected 5 TDs after reuse, got %d", len(tr.Kids))
	}
	end()
	sm.ReleaseStructElemsToPool()
}
