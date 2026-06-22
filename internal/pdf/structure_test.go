package pdf

import (
	"bytes"
	"testing"
)

func TestReserveMCIDs_batchAllocation(t *testing.T) {
	sm := NewStructureManager(true)
	start := sm.ReserveMCIDs(0, 7)
	if start != 0 {
		t.Fatalf("expected start MCID 0, got %d", start)
	}
	if sm.NextMCID[0] != 7 {
		t.Fatalf("expected NextMCID 7, got %d", sm.NextMCID[0])
	}
	next := sm.GetNextMCID(0)
	if next != 7 {
		t.Fatalf("expected next MCID 7, got %d", next)
	}
}

func TestBeginMarkedContentBufWithMCID_usesReservedID(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTR)

	var buf bytes.Buffer
	base := sm.ReserveMCIDs(0, 2)
	sm.BeginMarkedContentBufWithMCID(&buf, 0, StructTD, nil, base)
	sm.EndMarkedContentBuf(&buf)
	sm.BeginMarkedContentBufWithMCID(&buf, 0, StructTD, nil, base+1)
	sm.EndMarkedContentBuf(&buf)

	if sm.NextMCID[0] != 2 {
		t.Fatalf("expected NextMCID 2 after reserved pair, got %d", sm.NextMCID[0])
	}
	if len(sm.ParentTree[0]) != 2 {
		t.Fatalf("expected 2 parent tree entries, got %d", len(sm.ParentTree[0]))
	}
}

func TestBeginMarkedContentBufWithMCID_createsTDUnderTR(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent

	sm.BeginStructureElementCap(StructTR, 3)
	tr := sm.CurrentParent

	var buf bytes.Buffer
	base := sm.ReserveMCIDs(0, 3)
	for i := range 3 {
		sm.BeginMarkedContentBufWithMCID(&buf, 0, StructTD, nil, base+i)
		sm.EndMarkedContentBuf(&buf)
	}
	sm.EndStructureElement()

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
			t.Fatalf("kid %d: expected MCID on TD, got kids=%+v inline=(%d,%v)", i, kid.Elem.Kids, kid.Elem.MCID, kid.Elem.HasMCID)
		}
	}
}

func TestBeginTableRowWithTDMCIDs_parentTreeReferencesTD(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	sm.BeginTableRowWithTDMCIDs(0, 10, 3)
	tr := sm.CurrentParent

	if len(sm.ParentTree[0]) != 3 {
		t.Fatalf("expected 3 parent tree entries, got %d", len(sm.ParentTree[0]))
	}
	for i, ref := range sm.ParentTree[0] {
		if ref.Type != StructTD {
			t.Fatalf("ParentTree[%d] type=%s, want TD", i, ref.Type)
		}
		if mcid, ok := ref.LeafMCID(); !ok || mcid != 10+i {
			t.Fatalf("ParentTree[%d] MCID=%d ok=%v, want %d", i, mcid, ok, 10+i)
		}
		if ref.Parent != tr {
			t.Fatalf("ParentTree[%d] parent=%p, want TR %p", i, ref.Parent, tr)
		}
	}
}

func TestBeginTableRowWithTDMCIDs_trPageID(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	sm.BeginTableRowWithTDMCIDs(2, 0, 4)
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

func TestBeginTableRowWithTDMCIDs_createsTDUnderTR(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent

	sm.BeginTableRowWithTDMCIDs(0, 10, 3)
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
		if mcid, ok := kid.Elem.LeafMCID(); !ok || mcid != 10+i {
			t.Fatalf("kid %d: expected MCID %d, got %d ok=%v", i, 10+i, mcid, ok)
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

// TestBeginTableRowWithTDMCIDs_arenaAllocates asserts that the TD StructElems
// allocated by BeginTableRowWithTDMCIDs set up a TR → TD hierarchy and that
// ReleaseStructElemsToPool walks all elems back to the global pool. P1
// (2026-06-20 checklist) originally specified a per-document arena; we
// instead rely on the global sync.Pool + selective field clear (P2) to
// avoid the per-elem memclr that the pool was paying. This test pins the
// observable behaviour: the TD/TR shape is correct and release leaves the
// Elements slice at root-only.
func TestBeginTableRowWithTDMCIDs_arenaAllocates(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	sm.ReserveElementCapacity(arenaActivationThreshold)

	const rows = 4
	const cols = 7
	for r := 0; r < rows; r++ {
		sm.BeginTableRowWithTDMCIDs(0, r*cols, cols)
		sm.EndStructureElement()
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
	sm.BeginTableRowWithTDMCIDs(0, 0, 3)
	sm.EndStructureElement()
	sm.BeginTableRowWithTDMCIDs(0, 3, 3)
	sm.EndStructureElement()

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
// uses one manager per PDF.
func TestReleaseStructElemsToPool_canRunTwice(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	for r := 0; r < 3; r++ {
		sm.BeginTableRowWithTDMCIDs(0, r*4, 4)
		sm.EndStructureElement()
	}
	sm.ReleaseStructElemsToPool()

	sm.BeginStructureElement(StructTable)
	sm.BeginTableRowWithTDMCIDs(0, 0, 5)
	tr := sm.CurrentParent
	if len(tr.Kids) != 5 {
		t.Fatalf("expected 5 TDs after reuse, got %d", len(tr.Kids))
	}
	sm.EndStructureElement()
	sm.ReleaseStructElemsToPool()
}
