package models

import "testing"

func TestPreallocForDecodeDoesNotCreateElements(t *testing.T) {
	for _, tier := range []string{"retail", "active", "hft"} {
		t.Run(tier, func(t *testing.T) {
			tmpl := PDFTemplate{}
			tmpl.PreallocForDecode(0, tier)
			if len(tmpl.Elements) != 0 {
				t.Fatalf("elements len = %d, want 0", len(tmpl.Elements))
			}
			if cap(tmpl.Elements) == 0 {
				t.Fatal("expected element capacity to be reserved")
			}
		})
	}
}

func TestResetForReusePreservesElementBackingArray(t *testing.T) {
	var tmpl PDFTemplate
	tmpl.PreallocForDecode(0, "hft")
	if cap(tmpl.Elements) == 0 {
		t.Fatal("expected element capacity")
	}
	elements := tmpl.Elements[:cap(tmpl.Elements)]
	elements[0].Type = "stale"
	elementsPtr := &elements[0]

	tmpl.ResetForReuse()
	tmpl.PreallocForDecode(0, "hft")

	if len(tmpl.Elements) != 0 {
		t.Fatalf("elements len = %d, want 0", len(tmpl.Elements))
	}
	if &tmpl.Elements[:cap(tmpl.Elements)][0] != elementsPtr {
		t.Fatal("element backing array was not reused")
	}
}

func TestResetForReuseClearsTopLevelImageReferences(t *testing.T) {
	var tmpl PDFTemplate
	tmpl.Image = make([]Image, 1, 2)
	tmpl.Image[0].ImageData = "sensitive base64 payload"
	backing := tmpl.Image[:cap(tmpl.Image)]
	tmpl.ResetForReuse()
	if len(tmpl.Image) != 0 {
		t.Fatalf("image len = %d, want 0", len(tmpl.Image))
	}
	if backing[0].ImageData != "" {
		t.Fatal("reset retained top-level image payload")
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
