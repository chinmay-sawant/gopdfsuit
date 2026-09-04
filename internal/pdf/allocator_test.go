package pdf

import "testing"

// Allocator hands out unique IDs and records offsets and extras.
// The allocator is bound-only: the test binds a real PageManager instead of
// using a standalone counter.
func TestAllocatorUniqueness(t *testing.T) {
	reg := NewFontRegistry()
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{Top: 36, Bottom: 36, Left: 36, Right: 36}, false, reg, true, 32*1024)
	defer pm.ReleaseContentStreams()
	a := pm.ObjectAllocator(nil)
	// Start the counter at 2000 like the old standalone allocator.
	a.SeekTo(2000)
	ids := make(map[int]bool)
	for range 64 {
		id := a.Alloc()
		if ids[id] {
			t.Fatalf("duplicate object ID %d", id)
		}
		ids[id] = true
	}
	base := a.AllocN(8)
	for i := range 8 {
		if ids[base+i] {
			t.Fatalf("AllocN overlapped ID %d", base+i)
		}
	}
	if got := a.Next(); got != base+8 {
		t.Fatalf("Next = %d, want %d", got, base+8)
	}
	a.CommitString(base, "<< /Type /Test >>")
	if body, ok := a.Lookup(base); !ok || string(body) != "<< /Type /Test >>" {
		t.Fatalf("Lookup = %q, %v", body, ok)
	}
	a.SetOffset(base, 12345)
	if off, ok := a.Offset(base); !ok || off != 12345 {
		t.Fatalf("Offset = %d, %v", off, ok)
	}
}

// Bound allocator shares the PageManager counter: no ID may repeat and
// layout math must see the same next ID as the manager field.
func TestAllocatorBoundToPageManager(t *testing.T) {
	reg := NewFontRegistry()
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{Top: 36, Bottom: 36, Left: 36, Right: 36}, false, reg, true, 32*1024)
	defer pm.ReleaseContentStreams()
	var off []int
	alloc := pm.ObjectAllocator(&off)
	first := alloc.Alloc()
	second := pm.AllocObjectID()
	if first == second {
		t.Fatalf("bound allocator and facade returned same ID %d", first)
	}
	alloc.CommitString(first, "<< /Type /Bound >>")
	if _, ok := pm.ExtraObjects[first]; !ok {
		t.Fatalf("bound commit missing from ExtraObjects")
	}
	alloc.SetOffset(first, 77)
	if got, ok := xrefOffsetAt(off, first); !ok || got != 77 {
		t.Fatalf("bound offset = %d, %v", got, ok)
	}
	cs, fs, err := alloc.LayoutContentFontIDs(1, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if cs != 4 || fs != 5 {
		t.Fatalf("layout = %d,%d want 4,5", cs, fs)
	}
}
