package pdf

import (
	"bytes"
	"strings"
	"testing"
)

// TableTagger reserves consecutive MCIDs and emits matching BDC bytes.
func TestTableTaggerMCIDAccounting(t *testing.T) {
	sm := NewStructureManager(true)
	tagger := NewTableTagger(sm)
	before := tagger.MCIDCursor(0)

	var buf bytes.Buffer
	base := tagger.Reserve(0, 3)
	if base != before {
		t.Fatalf("Reserve base = %d, want cursor %d", base, before)
	}
	if got := tagger.MCIDCursor(0); got != before+3 {
		t.Fatalf("cursor after reserve = %d, want %d", got, before+3)
	}
	rowParent := sm.CurrentParent
	rowKidsBefore := len(rowParent.Kids)
	tagger.BeginRowWithBase(0, base, 3)
	for i := range 3 {
		tagger.WriteCell(&buf, base+i)
		buf.WriteString("EMC\n")
	}
	tagger.EndRow()
	if sm.CurrentParent != rowParent {
		t.Fatal("EndRow did not pop back to the pre-row parent")
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

// Disabled tagger is a no-op: no reservation, no bytes, no structure churn.
func TestTableTaggerDisabledNoop(t *testing.T) {
	sm := NewStructureManager(false)
	tagger := NewTableTagger(sm)
	var buf bytes.Buffer
	if base := tagger.BeginRow(0, 4); base != 0 {
		t.Fatalf("disabled BeginRow = %d, want 0", base)
	}
	tagger.WriteCell(&buf, 7)
	tagger.EndRow()
	if buf.Len() != 0 {
		t.Fatalf("disabled tagger wrote %d bytes", buf.Len())
	}
	if len(sm.Elements) != 1 {
		t.Fatalf("disabled tagger created %d elements", len(sm.Elements)-1)
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
