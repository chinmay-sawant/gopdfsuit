package pdf

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
)

func TestResolveFontName_appliesHelveticaStyleFlags(t *testing.T) {
	reg := NewFontRegistry()
	props := parseProps("Helvetica:18:100:center")
	if props.Bold != true {
		t.Fatal("expected bold from style code 100")
	}
	got := resolveFontName(props, reg)
	if got != "Helvetica-Bold" {
		t.Fatalf("resolveFontName = %q, want Helvetica-Bold", got)
	}
}

func TestEstimateTextWidth_activeTraderTitle(t *testing.T) {
	reg := NewFontRegistry()
	text := "ACTIVE TRADER CONTRACT NOTE"
	props := parseProps("Helvetica:18:100:center")
	resolved := resolveFontName(props, reg)

	got := EstimateTextWidth(resolved, text, 18, reg)
	want := font.StandardTextWidth("Helvetica-Bold", text, 18)
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("EstimateTextWidth = %.3f, want %.3f", got, want)
	}
	// Coarse 0.5em guess was 252pt; accurate Helvetica-Bold is ~307pt at 18pt.
	if math.Abs(got-252) < 1 {
		t.Fatalf("width %.3f still matches old 0.5em approximation", got)
	}
}

func TestCellTextX_clampsCenterWhenOverestimated(t *testing.T) {
	cellX := 72.0
	cellWidth := 169.125
	textWidth := 243.0
	got := cellTextX(cellX, cellWidth, textWidth, alignCenter)
	if got < cellX {
		t.Fatalf("cellTextX = %.3f, must be >= cellX %.3f", got, cellX)
	}
}

func TestCellTextX_centerKnownTitle(t *testing.T) {
	reg := NewFontRegistry()
	text := "ACTIVE TRADER CONTRACT NOTE"
	props := models.Props{FontName: "Helvetica", FontSize: 18, Bold: true, Alignment: alignCenter}
	resolved := resolveFontName(props, reg)
	width := EstimateTextWidth(resolved, text, 18, reg)
	got := cellTextX(72, 169.125, width, alignCenter)
	if got < 72 {
		t.Fatalf("centered title X %.3f starts before cell", got)
	}
	if got > 120 {
		t.Fatalf("centered title X %.3f unexpectedly far right", got)
	}
}