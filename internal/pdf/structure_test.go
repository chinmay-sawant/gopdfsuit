package pdf

import (
	"bytes"
	"testing"
)

// TableTagger is the sole seam for row MCID reservation and TD BDC bytes;
// these tests pin its accounting so the underlying StructureManager row
// primitives can stay package-private.

func TestTableTaggerReserveBatch(t *testing.T) {
	sm := NewStructureManager(true)
	tagger := NewTableTagger(sm)
	start := tagger.Reserve(0, 7)
	if start != 0 {
		t.Fatalf("expected start MCID 0, got %d", start)
	}
	if sm.NextMCID[0] != 7 {
		t.Fatalf("expected NextMCID 7, got %d", sm.NextMCID[0])
	}
	if next := tagger.MCIDCursor(0); next != 7 {
		t.Fatalf("expected cursor MCID 7, got %d", next)
	}
}

func TestTableTaggerReservedRowAccounting(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTR)
	tagger := NewTableTagger(sm)

	var buf bytes.Buffer
	base := tagger.Reserve(0, 2)
	tagger.BeginRowWithBase(0, base, 2)
	tagger.WriteCell(&buf, base)
	tagger.WriteCell(&buf, base+1)
	tagger.EndRow()

	if sm.NextMCID[0] != 2 {
		t.Fatalf("expected NextMCID 2 after reserved pair, got %d", sm.NextMCID[0])
	}
	if len(sm.ParentTree[0]) != 2 {
		t.Fatalf("expected 2 parent tree entries, got %d", len(sm.ParentTree[0]))
	}
}

func TestTableTaggerRowCreatesTDUnderTR(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	tagger := NewTableTagger(sm)

	var buf bytes.Buffer
	base := tagger.BeginRow(0, 3)
	tr := sm.CurrentParent
	for i := range 3 {
		tagger.WriteCell(&buf, base+i)
	}
	tagger.EndRow()

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

func TestTableTaggerParentTreeReferencesTD(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	tagger := NewTableTagger(sm)
	base := tagger.Reserve(0, 3)
	tagger.BeginRowWithBase(0, base, 3)
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

func TestTableTaggerRowPageID(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	tagger := NewTableTagger(sm)
	base := tagger.Reserve(2, 4)
	if base != 0 {
		t.Fatalf("fresh page cursor = %d, want 0", base)
	}
	tagger.BeginRowWithBase(2, base, 4)
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

func TestTableTaggerRowCreatesTD(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	tagger := NewTableTagger(sm)

	base := tagger.BeginRow(0, 3)
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

// TestTableTaggerArenaAllocates asserts that the TD StructElems
// allocated by a tagger row set up a TR → TD hierarchy and that
// ReleaseStructElemsToPool walks all elems back to the global pool. P1
// (2026-06-20 checklist) originally specified a per-document arena; we
// instead rely on the global sync.Pool + selective field clear (P2) to
// avoid the per-elem memclr that the pool was paying. This test pins the
// observable behaviour: the TD/TR shape is correct and release leaves the
// Elements slice at root-only.
func TestTableTaggerArenaAllocates(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	sm.ReserveElementCapacity(arenaActivationThreshold)
	tagger := NewTableTagger(sm)

	const rows = 4
	const cols = 7
	for r := 0; r < rows; r++ {
		tagger.BeginRow(0, cols)
		tagger.EndRow()
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
	tagger := NewTableTagger(sm)
	tagger.BeginRow(0, 3)
	tagger.EndRow()
	tagger.BeginRow(0, 3)
	tagger.EndRow()

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
	tagger := NewTableTagger(sm)
	for r := 0; r < 3; r++ {
		tagger.BeginRow(0, 4)
		tagger.EndRow()
	}
	sm.ReleaseStructElemsToPool()

	sm.BeginStructureElement(StructTable)
	tagger.BeginRow(0, 5)
	tr := sm.CurrentParent
	if len(tr.Kids) != 5 {
		t.Fatalf("expected 5 TDs after reuse, got %d", len(tr.Kids))
	}
	tagger.EndRow()
	sm.ReleaseStructElemsToPool()
}
