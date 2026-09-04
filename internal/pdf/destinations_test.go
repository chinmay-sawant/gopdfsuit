package pdf

import (
	"strings"
	"testing"
)

// Destinations Define/Lookup round-trips and Resolve maps to page IDs.
func TestDestinationsDefineResolve(t *testing.T) {
	reg := NewFontRegistry()
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{Top: 36, Bottom: 36, Left: 36, Right: 36}, false, reg, true, 32*1024)
	defer pm.ReleaseContentStreams()
	d := NewDestinationStore(pm, nil)

	d.Define("ch1", 0, 700)
	dest, ok := d.Lookup("ch1")
	if !ok || dest.PageIndex != 0 || dest.Y != 700 {
		t.Fatalf("Lookup = %+v, %v", dest, ok)
	}
	if _, ok := d.Lookup("missing"); ok {
		t.Fatal("Lookup of missing dest succeeded")
	}
	if got := d.ResolvePageObjectID(dest); got != pm.Pages[0] {
		t.Fatalf("Resolve = %d, want %d", got, pm.Pages[0])
	}
	d.UpdateStructElemID("ch1", 4242)
	if dest, _ := d.Lookup("ch1"); dest.StructElemID != 4242 {
		t.Fatalf("StructElemID = %d, want 4242", dest.StructElemID)
	}
}

// Emit writes a sorted name tree plus wrapper and resolves struct dests.
func TestDestinationsEmitNameTree(t *testing.T) {
	reg := NewFontRegistry()
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{Top: 36, Bottom: 36, Left: 36, Right: 36}, false, reg, true, 32*1024)
	defer pm.ReleaseContentStreams()
	d := NewDestinationStore(pm, nil)

	if _, ok := d.EmitNameTree(); ok {
		t.Fatal("Emit with no dests succeeded")
	}
	d.DefineFull("b-dest", NamedDest{PageIndex: 0, Y: 100, StructElemID: 9001})
	d.Define("a-dest", 0, 200)

	namesID, ok := d.EmitNameTree()
	if !ok || namesID == 0 {
		t.Fatalf("Emit = %d, %v", namesID, ok)
	}
	namesBody, ok := pm.ExtraObjects[namesID]
	if !ok || !strings.Contains(string(namesBody), "/Dests") {
		t.Fatalf("Names object missing /Dests: %q", namesBody)
	}
	found := false
	for _, body := range pm.ExtraObjects {
		s := string(body)
		if strings.Contains(s, "/Names") && strings.Contains(s, "(a-dest)") && strings.Contains(s, "(b-dest)") {
			found = true
			if strings.Index(s, "(a-dest)") > strings.Index(s, "(b-dest)") {
				t.Fatalf("name tree not sorted: %q", s)
			}
			if !strings.Contains(s, "/SD [9001 0 R") {
				t.Fatalf("struct dest missing /SD: %q", s)
			}
		}
	}
	if !found {
		t.Fatal("Dests name tree object not found")
	}
}

// Link annotations allocate unique IDs and gain structure elements when tagged.
func TestDestinationsLinkAnnotation(t *testing.T) {
	reg := NewFontRegistry()
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{Top: 36, Bottom: 36, Left: 36, Right: 36}, false, reg, true, 32*1024)
	defer pm.ReleaseContentStreams()
	d := NewDestinationStore(pm, nil)

	external := d.CreateLinkAnnotation(LinkAnnotation{Rect: [4]float64{0, 0, 10, 10}, URI: "https://example.com"})
	internal := d.CreateLinkAnnotation(LinkAnnotation{Rect: [4]float64{0, 0, 10, 10}, Dest: "a-dest"})
	if external == 0 || internal == 0 || external == internal {
		t.Fatalf("link IDs = %d, %d", external, internal)
	}
	if body := string(pm.ExtraObjects[external]); !strings.Contains(body, "/URI") {
		t.Fatalf("external annot missing /URI: %q", body)
	}
	if body := string(pm.ExtraObjects[internal]); !strings.Contains(body, "/Dest (a-dest)") {
		t.Fatalf("internal annot missing /Dest: %q", body)
	}
	if len(pm.AnnotStructElems) != 2 {
		t.Fatalf("AnnotStructElems = %d, want 2", len(pm.AnnotStructElems))
	}
}
