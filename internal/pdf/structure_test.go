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
