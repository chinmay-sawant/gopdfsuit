package pdf

import (
	"bytes"
	"encoding/base64"
	"regexp"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// tinyPNG is a 1x1 red PNG used to exercise image XObject emission.
const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

var objHeaderRe = regexp.MustCompile(`(?m)^(\d+) 0 obj`)

func collectObjectIDs(t *testing.T, pdfBytes []byte) []int {
	t.Helper()
	matches := objHeaderRe.FindAllSubmatch(pdfBytes, -1)
	if len(matches) == 0 {
		t.Fatal("no object headers found")
	}
	ids := make([]int, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, atoiBytes(m[1]))
	}
	return ids
}

func atoiBytes(b []byte) int {
	n := 0
	for _, c := range b {
		n = n*10 + int(c-'0')
	}
	return n
}

func imageTemplate(pages int) models.PDFTemplate {
	raw, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		panic(err)
	}
	_ = raw
	tpl := models.PDFTemplate{
		Config: models.Config{Page: "A4", PageAlignment: 1},
		Title: models.Title{
			Props: "Helvetica:14:100:center:0:0:0:0",
			Text:  "Image ID test",
		},
	}
	for range pages {
		tpl.Elements = append(tpl.Elements, models.Element{
			Type:  "image",
			Image: &models.Image{ImageName: "dot.png", ImageData: tinyPNGBase64, Width: 10, Height: 10},
		})
	}
	return tpl
}

// TestLayoutContentFontIDs covers the dense/shifted/guard branches of the
// object-ID layout: small docs keep the legacy dense IDs, very large docs
// shift content/fonts above the extras region, and absurd page counts fail.
func TestLayoutContentFontIDs(t *testing.T) {
	// Legacy: 3 pages, extras at [2000, 2010).
	cs, fs, err := layoutContentFontIDs(3, 2000, 2010)
	if err != nil {
		t.Fatalf("legacy layout: %v", err)
	}
	if cs != 6 || fs != 9 {
		t.Fatalf("legacy layout = (%d, %d), want (6, 9)", cs, fs)
	}
	// Shifted: 1500 pages would run content [1503, 3003) into extras.
	cs, fs, err = layoutContentFontIDs(1500, 2000, 2010)
	if err != nil {
		t.Fatalf("shifted layout: %v", err)
	}
	if cs != 2010 || fs != 3510 {
		t.Fatalf("shifted layout = (%d, %d), want (2010, 3510)", cs, fs)
	}
	// Guard: page IDs themselves reach the extras region.
	if _, _, err = layoutContentFontIDs(1998, 2000, 2500); err == nil {
		t.Fatal("expected layout-range error for 1998 pages, got nil")
	}
}

// TestShiftedLayoutLargeDoc generates a ~1100-page doc with an image, which
// forces the content/font block above the extras region, and asserts every
// emitted object ID is unique (no page/content/font/image collision).
func TestShiftedLayoutLargeDoc(t *testing.T) {
	tpl := models.PDFTemplate{
		Config: models.Config{Page: "A4", PageAlignment: 1},
		Title:  models.Title{Props: "Helvetica:14:100:center:0:0:0:0", Text: "shift"},
		Image:  []models.Image{{ImageName: "dot.png", ImageData: tinyPNGBase64, Width: 10, Height: 10}},
	}
	rows := make([]models.Row, 0, 35000)
	for i := range 35000 {
		rows = append(rows, models.Row{Row: []models.Cell{{Props: "Helvetica:10:100:left:0:0:0:0", Text: "row content"}}})
		_ = i
	}
	tpl.Elements = append(tpl.Elements, models.Element{Type: "table", Table: &models.Table{MaxColumns: 1, Rows: rows}})
	out, err := GenerateTemplatePDF(tpl)
	if err != nil {
		t.Fatalf("GenerateTemplatePDF: %v", err)
	}
	pages := bytes.Count(out, []byte("/Type /Page "))
	t.Logf("pages=%d bytes=%d", pages, len(out))
	if pages < 965 || pages >= 1997 {
		t.Fatalf("fixture must land in shifted-layout range [965, 1997): %d pages", pages)
	}
	ids := collectObjectIDs(t, out)
	seen := make(map[int]int, len(ids))
	for _, id := range ids {
		seen[id]++
		if seen[id] > 1 {
			t.Fatalf("object ID %d emitted %d times", id, seen[id])
		}
	}
	if !bytes.Contains(out, []byte("/Subtype /Image")) {
		t.Fatal("expected an image XObject")
	}
}

// TestImageObjectIDsFromAllocator asserts image XObjects take IDs from the
// page-manager allocator (>= 2000), never the old fixed base 1000 that
// collided with page IDs once documents grew past ~997 pages.
func TestImageObjectIDsFromAllocator(t *testing.T) {
	out, err := GenerateTemplatePDF(imageTemplate(3))
	if err != nil {
		t.Fatalf("GenerateTemplatePDF: %v", err)
	}
	ids := collectObjectIDs(t, out)
	seen := make(map[int]int, len(ids))
	for _, id := range ids {
		seen[id]++
	}
	for id, n := range seen {
		if n > 1 {
			t.Fatalf("object ID %d emitted %d times", id, n)
		}
	}
	idx := bytes.Index(out, []byte("/Subtype /Image"))
	if idx < 0 {
		t.Fatal("expected an image XObject")
	}
	headers := objHeaderRe.FindAllSubmatchIndex(out, -1)
	imgID := -1
	for _, h := range headers {
		if h[0] < idx {
			imgID = atoiBytes(out[h[2]:h[3]])
		} else {
			break
		}
	}
	if imgID < 2000 {
		t.Fatalf("image object ID %d not from allocator range (>= 2000)", imgID)
	}
	for _, pageID := range []int{3, 4, 5} {
		if imgID == pageID {
			t.Fatalf("image ID %d collides with page ID", imgID)
		}
	}
}
