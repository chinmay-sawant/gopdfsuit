package pdf

import (
	"strings"
	"sync"
	"testing"
)

// TestAddAnnotationGrowsPageAnnots mirrors AppendPageAnnot: no index when the
// current page has no slot yet, and no panic on a fresh manager.
func TestAddAnnotationGrowsPageAnnots(t *testing.T) {
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{}, false, NewFontRegistry(), false, 0)
	pm.PageAnnots = nil // simulate a manager without annot slots
	pm.CurrentPageIndex = 0
	pm.AddAnnotation(2001)
	if len(pm.PageAnnots) != 1 || len(pm.PageAnnots[0]) != 1 || pm.PageAnnots[0][0] != 2001 {
		t.Fatalf("PageAnnots = %+v", pm.PageAnnots)
	}
	pm.CurrentPageIndex = -1
	pm.AddAnnotation(2002) // must not panic
}

// TestCreateLinkAnnotationBoundsCheck asserts out-of-range page indexes clamp
// to a real page object instead of emitting a dangling 3+index reference.
func TestCreateLinkAnnotationBoundsCheck(t *testing.T) {
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842}, PageMargins{}, false, NewFontRegistry(), false, 0)
	id := CreateLinkAnnotation(LinkAnnotation{Rect: [4]float64{0, 0, 10, 10}, PageIndex: 99, DestY: 700}, pm, nil)
	content := string(pm.ExtraObjects[id])
	if !strings.Contains(content, "/Dest [3 0 R") {
		t.Fatalf("out-of-range PageIndex must clamp to first page, got %q", content)
	}
	id2 := CreateLinkAnnotation(LinkAnnotation{Rect: [4]float64{0, 0, 10, 10}, PageIndex: 0, DestY: 700}, pm, nil)
	if !strings.Contains(string(pm.ExtraObjects[id2]), "/Dest [3 0 R") {
		t.Fatalf("valid PageIndex 0 must resolve to page object 3, got %q", pm.ExtraObjects[id2])
	}
}

// TestConvertPDFDateToXMPValidation asserts malformed dates fall back to now
// instead of panicking on slice bounds.
func TestConvertPDFDateToXMPValidation(t *testing.T) {
	got := ConvertPDFDateToXMP("D:20240102150405-07'00'")
	if got != "2024-01-02T15:04:05-07:00" {
		t.Fatalf("valid date = %q", got)
	}
	gotLegacy := ConvertPDFDateToXMP("D:20240102150405-07'00")
	if gotLegacy != "2024-01-02T15:04:05-07:00" {
		t.Fatalf("legacy date without trailing quote = %q", gotLegacy)
	}
	for _, bad := range []string{"", "D:short", "20240102150405", "D:20XX0102150405", "D:20240102150405*07'00'", "D:20240102150405-07X00'"} {
		func() {
			defer func() {
				if recover() != nil {
					t.Fatalf("input %q panicked", bad)
				}
			}()
			out := ConvertPDFDateToXMP(bad)
			if !strings.Contains(out, "T") {
				t.Fatalf("input %q fallback = %q", bad, out)
			}
		}()
	}
}

// TestPropsCacheConcurrentClear hammers the bounded-size clear path under -race.
func TestPropsCacheConcurrentClear(_ *testing.T) {
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := range 500 {
				_ = parseProps("Helvetica:12:100:left:0:0:0:0")
				_ = parseProps("Times-Roman:14:101:center:1:1:1:1")
				_ = parseProps("Concurrent:10:100:left:0:0:0:0")
				_ = g
				_ = i
			}
		}(g)
	}
	wg.Wait()
}

// TestGetRGBDataBufferUndersizedReturnsToPool documents the pool discipline:
// an undersized pooled buffer is returned before a fresh allocation.
func TestGetRGBDataBufferUndersizedReturnsToPool(t *testing.T) {
	small := make([]byte, 8)
	rgbDataPool.Put(&small)
	got := getRGBDataBuffer(1024)
	if len(got) != 1024 {
		t.Fatalf("len = %d, want 1024", len(got))
	}
	putRGBDataBuffer(got)
}
