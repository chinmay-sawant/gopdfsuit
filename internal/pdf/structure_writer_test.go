package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestFormatStructElemTDLeaf_StableOutput is a golden-bytes test that pins
// the output of formatSingleMCIDTableCellStructElem + formatStructElemObjectTo
// to a checked-in reference. Any change to the TD-leaf or grouping-elem
// emission must update the reference, but the bytes must remain PDF/A-4 and
// PDF/UA-2 compliant (see guides/optimizations/20260620_zerodha_x10_pprof_optimization_checklist.md
// P3 / P8). When the reference file is absent the test writes it on first
// run so subsequent runs are checked.
//
// Regenerate the reference with:
//
//	UPDATE_STRUCT_ELEM_GOLDEN=1 go test ./internal/pdf -run TestFormatStructElemTDLeaf_StableOutput
func TestFormatStructElemTDLeaf_StableOutput(t *testing.T) {
	// Build a small but representative structure:
	//   Root
	//   └── Document (ObjectID=100, slow path)
	//       └── Table (ObjectID=101, no body emission here)
	//           └── TR (ObjectID=102)
	//               ├── TD (ObjectID=200, MCID=0, fast path)
	//               ├── TD (ObjectID=201, MCID=1, fast path)
	//               └── TD (ObjectID=202, MCID=2, fast path)
	sm := NewStructureManager(true)

	// Set up an explicit grouping element chain so we can stamp known
	// ObjectIDs and exercise both the slow (Document) and fast (TD) paths.
	root := sm.Root
	sm.BeginStructureElement(StructDocument)
	doc := sm.CurrentParent
	doc.ObjectID = 100
	doc.Title = "Example document"
	doc.Alt = "An example document"

	sm.BeginStructureElement(StructTable)
	table := sm.CurrentParent
	table.ObjectID = 101

	sm.BeginTableRowWithTDMCIDs(0, 0, 3)
	tr := sm.CurrentParent
	tr.ObjectID = 102

	tdRefs := make([]*StructElem, 0, 3)
	for _, kid := range tr.Kids {
		tdRefs = append(tdRefs, kid.Elem)
	}
	if len(tdRefs) != 3 {
		t.Fatalf("expected 3 TDs, got %d", len(tdRefs))
	}
	for i, td := range tdRefs {
		td.ObjectID = 200 + i
		td.PageID = 0
	}

	pages := []int{0: 3}

	var out bytes.Buffer
	ctx := structElemFormatCtx{
		namespaceID:      99,
		structTreeRootID: 5,
		root:             root,
		pages:            pages,
	}

	// Emit the Document (slow path)
	formatStructElemObjectTo(&out, doc, ctx)
	// Emit each TD via the fast path
	for _, td := range tdRefs {
		formatStructElemObjectTo(&out, td, ctx)
	}

	got := out.Bytes()

	refPath := filepath.Join("testdata", "struct_elem_td_leaf.golden")
	if os.Getenv("UPDATE_STRUCT_ELEM_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(refPath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(refPath, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("Wrote golden file: %s (%d bytes)", refPath, len(got))
		return
	}

	want, err := os.ReadFile(refPath)
	if err != nil {
		if os.IsNotExist(err) {
			// First run: write the golden and pass.
			if err := os.MkdirAll(filepath.Dir(refPath), 0o755); err != nil {
				t.Fatalf("mkdir testdata: %v", err)
			}
			if err := os.WriteFile(refPath, got, 0o644); err != nil {
				t.Fatalf("write golden: %v", err)
			}
			t.Logf("Created golden file: %s (%d bytes) - re-run to verify", refPath, len(got))
			return
		}
		t.Fatalf("read golden: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("TD-leaf output drifted from golden\n--- got ---\n%s\n--- want ---\n%s\nRe-run with UPDATE_STRUCT_ELEM_GOLDEN=1 to refresh.", got, want)
	}
}
