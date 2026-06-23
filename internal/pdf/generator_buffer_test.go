package pdf

import (
	"os"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

func TestMeasuredZerodhaOutputFitsBufferCaps(t *testing.T) {
	const (
		measuredRetail = 61_301
		measuredActive = 73_835
		measuredHFT    = 2_291_942
	)
	if retailPDFBufferCap < measuredRetail {
		t.Fatalf("retail cap %d < measured %d", retailPDFBufferCap, measuredRetail)
	}
	if activePDFBufferCap < measuredActive {
		t.Fatalf("active cap %d < measured %d", activePDFBufferCap, measuredActive)
	}
	if hftPDFBufferCap < measuredHFT {
		t.Fatalf("hft cap %d < measured %d", hftPDFBufferCap, measuredHFT)
	}
}

func TestTemplateCapacityTier(t *testing.T) {
	retail := models.PDFTemplate{
		Elements: []models.Element{{Type: "table", Table: &models.Table{Rows: make([]models.Row, 5)}}},
	}
	if got := templateCapacityTier(retail); got != "retail" {
		t.Fatalf("retail tier: got %q", got)
	}

	active := models.PDFTemplate{
		Elements: []models.Element{{Type: "table", Table: &models.Table{Rows: make([]models.Row, 41)}}},
	}
	if got := templateCapacityTier(active); got != "active" {
		t.Fatalf("active tier: got %q", got)
	}

	hft := models.PDFTemplate{
		Elements: []models.Element{{Type: "table", Table: &models.Table{Rows: make([]models.Row, 2000)}}},
	}
	if got := templateCapacityTier(hft); got != "hft" {
		t.Fatalf("hft tier: got %q", got)
	}
}

func TestPageStreamCapsAlignToMeasuredZerodha(t *testing.T) {
	if got := alignPageStreamCap(8755); got != pageStreamRetailCap {
		t.Fatalf("retail page cap: got %d want %d", got, pageStreamRetailCap)
	}
	if got := alignPageStreamCap(18805); got != 32*1024 {
		t.Fatalf("active measured page len cap bucket: got %d want %d", got, 32*1024)
	}
	if got := estimateInitialContentStreamCap(models.PDFTemplate{
		Elements: []models.Element{{Type: "table", Table: &models.Table{Rows: make([]models.Row, 41)}}},
	}); got != pageStreamActiveCap {
		t.Fatalf("active template initial stream cap: got %d want %d", got, pageStreamActiveCap)
	}
	if got := estimateSharedRowStripeCap(40, 7); got != 48*1024 {
		t.Fatalf("hft stripe cap: got %d want %d", got, 48*1024)
	}
}

func TestArenaCapTiers(t *testing.T) {
	if got := arenaCapForNeed(100); got != 0 {
		t.Fatalf("below threshold: got %d", got)
	}
	if got := arenaCapForNeed(600); got != arenaTierSmall {
		t.Fatalf("small tier: got %d want %d", got, arenaTierSmall)
	}
	if got := arenaCapForNeed(5000); got != arenaTierMedium {
		t.Fatalf("medium tier: got %d want %d", got, arenaTierMedium)
	}
	if got := arenaCapForNeed(20000); got != arenaTierLarge {
		t.Fatalf("large tier: got %d want %d", got, arenaTierLarge)
	}
}

func TestWarmedHFTBufferDoesNotGrowDuringEmit(t *testing.T) {
	t.Setenv("BENCH_DEBUG_CAPS", "1")
	ResetPDFCapacityHighWater()

	template := compliantLargeTableTemplate(1010)
	for i := 0; i < 2; i++ {
		doc, err := GenerateTemplatePDFBorrowed(template)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		doc.Release()
	}

	hw := SnapshotPDFCapacityHighWater()
	if hw.HFTLen == 0 {
		t.Fatal("expected HFT high-water len to be recorded")
	}
	if hw.PostCapGrow > 0 {
		t.Fatalf("expected zero post-content cap grows on warmed HFT buffer, got %d", hw.PostCapGrow)
	}
}

func compliantLargeTableTemplate(rows int) models.PDFTemplate {
	tableRows := make([]models.Row, rows)
	for i := range rows {
		tableRows[i] = models.Row{Row: []models.Cell{
			{Props: "Helvetica:8:000:center:1:0:0:1", Text: "RELIANCE"},
			{Props: "Helvetica:8:000:center:0:0:0:1", Text: "BUY"},
			{Props: "Helvetica:8:000:center:0:0:0:1", Text: "10"},
			{Props: "Helvetica:8:000:right:0:0:0:1", Text: "100.00"},
			{Props: "Helvetica:8:000:right:0:1:0:1", Text: "1000.00"},
		}}
	}
	embed := true
	return models.PDFTemplate{
		Config: models.Config{
			Page:                "A4",
			PageAlignment:       1,
			PDFACompliant:       true,
			TaggedPDF:           true,
			ArlingtonCompatible: true,
			EmbedFonts:          &embed,
		},
		Title: models.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "HFT buffer test",
		},
		Elements: []models.Element{{
			Type: "table",
			Table: &models.Table{
				MaxColumns:           5,
				ColumnWidths:         []float64{2, 1, 1, 1, 1},
				Rows:                 tableRows,
				SharedRowLayout:      true,
				SharedRowTemplateRow: 1,
			},
		}},
	}
}

func TestMain(m *testing.M) {
	WarmRuntimePools()
	os.Exit(m.Run())
}
