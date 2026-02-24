package pdf

import (
	"math"
	"testing"
)

// TestLiberationSansCompliance checks if LiberationSans font metrics and subsetting
// behave as expected for PDF/A compliance (specifically Glyph Widths).
func TestLiberationSansCompliance(t *testing.T) {
	// Initialize font manager (uses default or fallback)
	manager := GetPDFAFontManager()

	// Ensure fonts available (might download)
	if err := manager.EnsureFontsAvailable(); err != nil {
		t.Logf("Skipping test: Liberation fonts not found/downloadable: %v", err)
		return
	}

	// Load LiberationSans-Regular
	fontName := "LiberationSans-Regular"
	font, err := manager.GetLiberationFont("Helvetica")
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}

	// 1. Check Hyphen Width
	char := '-'
	glyphID := font.CharToGlyph[char]
	rawWidth := font.GetGlyphWidth(glyphID)

	scale := 1000.0 / float64(font.UnitsPerEm)
	scaledWidth := int(math.Round(float64(rawWidth) * scale))

	// The validator expects consistency with PDF /W array.
	// Common Arial/Helvetica hyphen width is 333 (scaled).
	if scaledWidth != 333 {
		t.Errorf("Hyphen width mismatch: expected 333, got %d (raw %d, UPM %d)", scaledWidth, rawWidth, font.UnitsPerEm)
	}

	// 2. Test Subsetting Integrity
	// Create registry
	registry := NewFontRegistry()
	if err := registry.RegisterFont(fontName, font); err != nil {
		t.Fatalf("RegisterFont failed: %v", err)
	}

	// Mark Hyphen used
	registry.MarkCharsUsed(fontName, "-")

	// Generate Subsets
	if err := registry.GenerateSubsets(); err != nil {
		t.Fatalf("GenerateSubsets failed: %v", err)
	}

	// Verify mapping
	regFont, _ := registry.GetFont(fontName)
	if newID, ok := regFont.OldToNewGlyph[glyphID]; ok {
		if newID == 0 {
			t.Errorf("Hyphen mapped to GID 0 (.notdef) in subset! Must be > 0.")
		}

		// Parse subset
		subsetFont, err := LoadTTFFromData(regFont.SubsetData)
		if err != nil {
			t.Fatalf("Failed to parse generated subset: %v", err)
		}

		subsetW := subsetFont.GetGlyphWidth(newID)
		if subsetW != rawWidth {
			t.Errorf("Subset glyph width mismatch: original %d, subset %d", rawWidth, subsetW)
		}
	} else {
		t.Errorf("Hyphen glyph (GID %d) missing from OldToNew map!", glyphID)
	}

	// 3. Verify CIDToGIDMap Generation
	cidMapStr := generateCIDToGIDMap(regFont, nil)
	if len(cidMapStr) < 50 {
		t.Errorf("CIDToGIDMap stream too short: %d", len(cidMapStr))
	}
}
