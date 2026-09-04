package pdf

import (
	"bytes"
	"strings"
	"testing"
)

// EmitRowCells reserves consecutive MCIDs, builds TR + TD elems, and emits
// matching BDC bytes through writeCellMarkedContentBDC.
func TestEmitRowCellsMCIDAccounting(t *testing.T) {
	sm := NewStructureManager(true)
	before := sm.PageMCIDStart(0)

	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 3)
	if base != before {
		t.Fatalf("EmitRowCells base = %d, want cursor %d", base, before)
	}
	if got := sm.PageMCIDStart(0); got != before+3 {
		t.Fatalf("cursor after emit = %d, want %d", got, before+3)
	}
	rowParent := sm.CurrentParent.Parent
	rowKidsBefore := len(rowParent.Kids) - 1 // TR already attached by EmitRowCells
	for i := range 3 {
		sm.writeCellMarkedContentBDC(&buf, StructTD, base+i)
		buf.WriteString("EMC\n")
	}
	end()
	if sm.CurrentParent != rowParent {
		t.Fatal("end did not pop back to the pre-row parent")
	}
	if len(rowParent.Kids) != rowKidsBefore+1 {
		t.Fatalf("TR not attached: kids %d -> %d", rowKidsBefore, len(rowParent.Kids))
	}
	out := buf.String()
	for i := range 3 {
		want := "/TD <</MCID " + itoa(base+i) + ">> BDC"
		if !strings.Contains(out, want) {
			t.Fatalf("missing BDC %q in %q", want, out)
		}
	}
	if len(sm.ParentTree[0]) < 3 {
		t.Fatalf("ParentTree has %d entries, want >= 3", len(sm.ParentTree[0]))
	}
}

// Disabled manager is a no-op: no reservation, no bytes, no structure churn.
// Nil manager behaves the same.
func TestEmitRowCellsDisabledNoop(t *testing.T) {
	sm := NewStructureManager(false)
	var buf bytes.Buffer
	base, end := sm.EmitRowCells(&buf, 0, 4)
	end()
	if base != 0 {
		t.Fatalf("disabled EmitRowCells = %d, want 0", base)
	}
	sm.writeCellMarkedContentBDC(&buf, StructTD, 7)
	if buf.Len() != 0 {
		t.Fatalf("disabled manager wrote %d bytes", buf.Len())
	}
	if len(sm.Elements) != 1 {
		t.Fatalf("disabled emit created %d elements", len(sm.Elements)-1)
	}

	var nilSM *StructureManager
	base, end = nilSM.EmitRowCells(&buf, 0, 4)
	end()
	if base != 0 {
		t.Fatalf("nil EmitRowCells = %d, want 0", base)
	}

	for _, n := range []int{0, -1} {
		func() {
			defer func() {
				if recover() != nil {
					t.Fatalf("count %d EmitRowCells panicked", n)
				}
			}()
			_, end := sm.EmitRowCells(&buf, 0, n)
			end()
		}()
	}
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
