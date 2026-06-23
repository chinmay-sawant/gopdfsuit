package models

import "testing"

func TestPreallocForDecodeHFTRowsRemainZeroLength(t *testing.T) {
	var tmpl PDFTemplate
	tmpl.PreallocForDecode(0, "hft")

	if len(tmpl.Elements) < 2 {
		t.Fatalf("elements len = %d, want at least 2", len(tmpl.Elements))
	}

	first := tmpl.Elements[0].Table
	if first == nil {
		t.Fatal("expected first preallocated table")
	}
	if got, want := len(first.Rows), 0; got != want {
		t.Fatalf("first rows len = %d, want %d", got, want)
	}
	if got, want := cap(first.Rows), 1; got < want {
		t.Fatalf("first rows cap = %d, want >= %d", got, want)
	}
	if got, want := cap(first.Rows[:cap(first.Rows)][0].Row), 4; got < want {
		t.Fatalf("first row cell cap = %d, want >= %d", got, want)
	}

	second := tmpl.Elements[1].Table
	if second == nil {
		t.Fatal("expected second preallocated table")
	}
	if got, want := len(second.Rows), 0; got != want {
		t.Fatalf("second rows len = %d, want %d", got, want)
	}
	if got, want := cap(second.Rows), 2001; got < want {
		t.Fatalf("second rows cap = %d, want >= %d", got, want)
	}
	if got, want := cap(second.Rows[:cap(second.Rows)][0].Row), 7; got < want {
		t.Fatalf("second row cell cap = %d, want >= %d", got, want)
	}
}

func TestResetForReusePreservesHFTBackingArrays(t *testing.T) {
	var tmpl PDFTemplate
	tmpl.PreallocForDecode(0, "hft")

	firstTable := tmpl.Elements[0].Table
	secondTable := tmpl.Elements[1].Table
	if firstTable == nil || secondTable == nil {
		t.Fatal("expected preallocated inline tables")
	}

	firstRowsPtr := &firstTable.Rows[:cap(firstTable.Rows)][0]
	secondRowsPtr := &secondTable.Rows[:cap(secondTable.Rows)][0]
	firstCellPtr := &firstTable.Rows[:cap(firstTable.Rows)][0].Row[:cap(firstTable.Rows[:cap(firstTable.Rows)][0].Row)][0]
	secondCellPtr := &secondTable.Rows[:cap(secondTable.Rows)][0].Row[:cap(secondTable.Rows[:cap(secondTable.Rows)][0].Row)][0]

	tmpl.Elements[0].Type = "table"
	tmpl.Elements[0].Table.Rows = tmpl.Elements[0].Table.Rows[:1]
	tmpl.Elements[0].Table.Rows[0].Row = append(tmpl.Elements[0].Table.Rows[0].Row, Cell{Text: "stale"})
	tmpl.Elements[1].Type = "table"
	tmpl.Elements[1].Table.Rows = tmpl.Elements[1].Table.Rows[:2]
	tmpl.Elements[1].Table.Rows[0].Row = append(tmpl.Elements[1].Table.Rows[0].Row, Cell{Text: "stale"})

	tmpl.ResetForReuse()
	tmpl.PreallocForDecode(0, "hft")

	if got, want := len(tmpl.Elements), 2; got != want {
		t.Fatalf("elements len = %d, want %d", got, want)
	}
	if tmpl.Elements[0].Table == nil || tmpl.Elements[1].Table == nil {
		t.Fatal("expected inline tables after reuse")
	}
	if &tmpl.Elements[0].Table.Rows[:cap(tmpl.Elements[0].Table.Rows)][0] != firstRowsPtr {
		t.Fatal("first table rows backing array was not reused")
	}
	if &tmpl.Elements[1].Table.Rows[:cap(tmpl.Elements[1].Table.Rows)][0] != secondRowsPtr {
		t.Fatal("second table rows backing array was not reused")
	}
	if &tmpl.Elements[0].Table.Rows[:cap(tmpl.Elements[0].Table.Rows)][0].Row[:cap(tmpl.Elements[0].Table.Rows[:cap(tmpl.Elements[0].Table.Rows)][0].Row)][0] != firstCellPtr {
		t.Fatal("first table cell backing array was not reused")
	}
	if &tmpl.Elements[1].Table.Rows[:cap(tmpl.Elements[1].Table.Rows)][0].Row[:cap(tmpl.Elements[1].Table.Rows[:cap(tmpl.Elements[1].Table.Rows)][0].Row)][0] != secondCellPtr {
		t.Fatal("second table cell backing array was not reused")
	}
	if got := len(tmpl.Elements[0].Table.Rows); got != 0 {
		t.Fatalf("first table rows len = %d, want 0", got)
	}
	if got := len(tmpl.Elements[1].Table.Rows); got != 0 {
		t.Fatalf("second table rows len = %d, want 0", got)
	}
}

func TestSetPrecomputedStandardFontsPreservedUntilReset(t *testing.T) {
	var tmpl PDFTemplate
	tmpl.SetPrecomputedStandardFonts("Helvetica", "Times-Roman")

	got := tmpl.PrecomputedStandardFonts()
	if len(got) != 2 || got[0] != "Helvetica" || got[1] != "Times-Roman" {
		t.Fatalf("precomputed fonts = %#v, want []string{\"Helvetica\", \"Times-Roman\"}", got)
	}

	tmpl.ResetForReuse()
	if got := tmpl.PrecomputedStandardFonts(); len(got) != 0 {
		t.Fatalf("precomputed fonts after reset = %#v, want empty", got)
	}
}
