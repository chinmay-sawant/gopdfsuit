package pdf

import (
	"bytes"
	"testing"
)

func TestFormatTRStructElemObjectToMatchesSlowPath(t *testing.T) {
	sm := NewStructureManager(true)
	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	table.ObjectID = 50

	sm.ReserveElementCapacity(600)
	sm.BeginTableRowWithTDMCIDs(0, 10, 3)
	tr := sm.CurrentParent
	tr.ObjectID = 60

	for i, kid := range tr.Kids {
		kid.Elem.ObjectID = 200 + i
	}

	ctx := structElemFormatCtx{
		namespaceID:      99,
		structTreeRootID: 5,
		root:             sm.Root,
		pages:            []int{3},
	}

	var fast, slow bytes.Buffer
	if !formatTRStructElemObjectTo(&fast, tr, ctx) {
		t.Fatal("expected TR fast path to match arena TR")
	}

	tr.groupEmitFast = false
	formatStructElemObjectTo(&slow, tr, ctx)

	if !bytes.Equal(fast.Bytes(), slow.Bytes()) {
		t.Fatalf("TR fast/slow output mismatch\nfast:\n%s\nslow:\n%s", fast.Bytes(), slow.Bytes())
	}
}
