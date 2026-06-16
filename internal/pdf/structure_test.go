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
		if len(kid.Elem.Kids) != 1 || kid.Elem.Kids[0].MCID != base+i {
			t.Fatalf("kid %d: expected MCID on TD, got %+v", i, kid.Elem.Kids)
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
