package pdf

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// CustomFontRegistry manages loaded custom fonts for PDF generation
type CustomFontRegistry struct {
	mu    sync.RWMutex
	fonts map[string]*RegisteredFont
}

// RegisteredFont represents a custom font loaded and ready for PDF embedding
type RegisteredFont struct {
	Name          string            // User-friendly reference name (used in templates)
	Font          *TTFFont          // Parsed font data
	UsedChars     map[rune]bool     // Characters used in the current document
	SubsetData    []byte            // Subset font data (generated when needed)
	OldToNewGlyph map[uint16]uint16 // Glyph ID mapping after subsetting
	ObjectID      int               // PDF object ID for font dictionary
	DescriptorID  int               // PDF object ID for font descriptor
	ToUnicodeID   int               // PDF object ID for ToUnicode CMap
	CIDFontID     int               // PDF object ID for CIDFont
	CIDToGIDMapID int               // PDF object ID for CIDToGIDMap
	FontFileID    int               // PDF object ID for embedded font file
	WidthsID      int               // PDF object ID for widths array
}

// Global font registry instance
var globalFontRegistry = &CustomFontRegistry{
	fonts: make(map[string]*RegisteredFont),
}

// GetFontRegistry returns the global font registry
func GetFontRegistry() *CustomFontRegistry {
	return globalFontRegistry
}

// NewFontRegistry creates a new font registry (for isolated use cases)
func NewFontRegistry() *CustomFontRegistry {
	return &CustomFontRegistry{
		fonts: make(map[string]*RegisteredFont),
	}
}

// RegisterFontFromFile loads and registers a TTF/OTF font from a file path
func (r *CustomFontRegistry) RegisterFontFromFile(name string, path string) error {
	font, err := LoadTTFFromFile(path)
	if err != nil {
		return fmt.Errorf("failed to load font from %s: %w", path, err)
	}

	return r.RegisterFont(name, font)
}

// RegisterFontFromData loads and registers a TTF/OTF font from raw bytes
func (r *CustomFontRegistry) RegisterFontFromData(name string, data []byte) error {
	font, err := LoadTTFFromData(data)
	if err != nil {
		return fmt.Errorf("failed to parse font data: %w", err)
	}

	return r.RegisterFont(name, font)
}

// RegisterFontFromBase64 loads and registers a TTF/OTF font from base64-encoded data
func (r *CustomFontRegistry) RegisterFontFromBase64(name string, base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 font data: %w", err)
	}

	return r.RegisterFontFromData(name, data)
}

// RegisterFont registers a parsed TTFFont with a given name
func (r *CustomFontRegistry) RegisterFont(name string, font *TTFFont) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fonts[name] = &RegisteredFont{
		Name:      name,
		Font:      font,
		UsedChars: make(map[rune]bool),
	}

	return nil
}

// GetFont returns a registered font by name
func (r *CustomFontRegistry) GetFont(name string) (*RegisteredFont, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	font, ok := r.fonts[name]
	return font, ok
}

// HasFont checks if a font is registered
func (r *CustomFontRegistry) HasFont(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.fonts[name]
	return ok
}

// MarkCharsUsed marks characters as used for a font (for subsetting)
func (r *CustomFontRegistry) MarkCharsUsed(name string, text string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if font, ok := r.fonts[name]; ok {
		for _, char := range text {
			font.UsedChars[char] = true
		}
	}
}

// GenerateSubsets generates subset fonts for all registered fonts with used characters
func (r *CustomFontRegistry) GenerateSubsets() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, font := range r.fonts {
		if len(font.UsedChars) == 0 {
			continue
		}

		// Collect used glyphs
		usedGlyphs := make([]uint16, 0, len(font.UsedChars))
		for char := range font.UsedChars {
			if glyphID, ok := font.Font.CharToGlyph[char]; ok {
				usedGlyphs = append(usedGlyphs, glyphID)
			}
		}

		// Generate subset
		subsetData, oldToNew, err := SubsetTTF(font.Font, usedGlyphs)
		if err != nil {
			return fmt.Errorf("failed to subset font %s: %w", name, err)
		}

		font.SubsetData = subsetData
		font.OldToNewGlyph = oldToNew
	}

	return nil
}

// GetAllFonts returns all registered fonts
func (r *CustomFontRegistry) GetAllFonts() []*RegisteredFont {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fonts := make([]*RegisteredFont, 0, len(r.fonts))
	for _, font := range r.fonts {
		fonts = append(fonts, font)
	}
	return fonts
}

// GetUsedFonts returns fonts that have characters marked as used
func (r *CustomFontRegistry) GetUsedFonts() []*RegisteredFont {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fonts := make([]*RegisteredFont, 0)
	for _, font := range r.fonts {
		if len(font.UsedChars) > 0 {
			fonts = append(fonts, font)
		}
	}
	return fonts
}

// Clear removes all registered fonts
func (r *CustomFontRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fonts = make(map[string]*RegisteredFont)
}

// ResetUsage clears usage data for all fonts (call before generating a new PDF)
func (r *CustomFontRegistry) ResetUsage() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, font := range r.fonts {
		font.UsedChars = make(map[rune]bool)
		font.SubsetData = nil
		font.OldToNewGlyph = nil
		font.ObjectID = 0
		font.DescriptorID = 0
		font.ToUnicodeID = 0
		font.CIDFontID = 0
		font.CIDToGIDMapID = 0
		font.FontFileID = 0
		font.WidthsID = 0
	}
}

// CloneForGeneration creates a shallow clone of the registry with reset usage data.
// This allows concurrent PDF generation without race conditions on UsedChars.
func (r *CustomFontRegistry) CloneForGeneration() *CustomFontRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := &CustomFontRegistry{
		fonts: make(map[string]*RegisteredFont, len(r.fonts)),
	}

	for name, font := range r.fonts {
		// Create a new RegisteredFont instance sharing the same static TTFFont data
		// but with fresh usage maps and specific PDF object IDs
		// Pre-size UsedChars to 256 to avoid map rehashing during font scanning
		clone.fonts[name] = &RegisteredFont{
			Name:      font.Name,
			Font:      font.Font,
			UsedChars: make(map[rune]bool, 256),
			// Other fields default to zero/nil
		}
	}

	return clone
}

// AssignObjectIDs assigns PDF object IDs to font objects
// Returns the next available object ID
func (r *CustomFontRegistry) AssignObjectIDs(startID int) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentID := startID
	for _, font := range r.fonts {
		if len(font.UsedChars) == 0 {
			continue
		}

		// Each custom font needs 7 objects:
		// 1. Type 0 font dictionary
		// 2. CIDFont dictionary
		// 3. FontDescriptor
		// 4. Widths array
		// 5. CIDToGIDMap stream
		// 6. ToUnicode CMap
		// 7. FontFile2 stream
		font.ObjectID = currentID
		currentID++
		font.CIDFontID = currentID
		currentID++
		font.DescriptorID = currentID
		currentID++
		font.WidthsID = currentID
		currentID++
		font.CIDToGIDMapID = currentID
		currentID++
		font.ToUnicodeID = currentID
		currentID++
		font.FontFileID = currentID
		currentID++
	}

	return currentID
}

// GetFontReference returns the PDF reference string for a custom font (e.g., "/CF1")
func (r *CustomFontRegistry) GetFontReference(name string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if font, ok := r.fonts[name]; ok && font.ObjectID > 0 {
		return fmt.Sprintf("/CF%d", font.ObjectID)
	}
	return ""
}

// IsCustomFont checks if a font name refers to a custom font (not a standard PDF font)
func IsCustomFont(fontName string) bool {
	standardFonts := map[string]bool{
		"Helvetica": true, "Helvetica-Bold": true, "Helvetica-Oblique": true, "Helvetica-BoldOblique": true,
		"Times-Roman": true, "Times-Bold": true, "Times-Italic": true, "Times-BoldItalic": true,
		"Courier": true, "Courier-Bold": true, "Courier-Oblique": true, "Courier-BoldOblique": true,
		"Symbol": true, "ZapfDingbats": true,
		"font1": true, "font2": true, // Legacy font references
	}

	return !standardFonts[fontName]
}

// LoadFontsFromDirectory loads all TTF/OTF fonts from a directory
func (r *CustomFontRegistry) LoadFontsFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read font directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".ttf" && ext != ".otf" {
			continue
		}

		fontName := strings.TrimSuffix(name, ext)
		fontPath := filepath.Join(dir, name)

		if err := r.RegisterFontFromFile(fontName, fontPath); err != nil {
			// Log error but continue with other fonts
			fmt.Printf("Warning: failed to load font %s: %v\n", fontPath, err)
		}
	}

	return nil
}

// GetTextWidth calculates the width of text in a custom font (in PDF units at 1pt)
func (r *CustomFontRegistry) GetTextWidth(fontName string, text string) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	font, ok := r.fonts[fontName]
	if !ok {
		return 0
	}

	var totalWidth float64
	for _, char := range text {
		width := font.Font.GetCharWidthScaled(char)
		totalWidth += float64(width) / 1000.0
	}

	return totalWidth
}

// GetScaledTextWidth calculates the width of text at a specific font size
func (r *CustomFontRegistry) GetScaledTextWidth(fontName string, text string, fontSize float64) float64 {
	return r.GetTextWidth(fontName, text) * fontSize
}

// GeneratePDFFontResources generates the PDF font resource dictionary entries for custom fonts
func (r *CustomFontRegistry) GeneratePDFFontResources() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var resources strings.Builder
	for _, font := range r.fonts {
		if font.ObjectID > 0 {
			resources.WriteString(fmt.Sprintf(" /CF%d %d 0 R", font.ObjectID, font.ObjectID))
		}
	}

	return resources.String()
}

// IsCustomFont checks if the font name refers to a registered custom font
func (r *CustomFontRegistry) IsCustomFont(fontName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.fonts[fontName]
	return ok
}

// ResolveFontName resolves the actual font name to use, handling fallbacks
func (r *CustomFontRegistry) ResolveFontName(props models.Props) string {
	// 1. Check if the requested font is registered as a custom font
	if r.IsCustomFont(props.FontName) {
		return props.FontName
	}

	// 2. Check if it's a known standard font name
	switch props.FontName {
	case "Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique",
		"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic",
		"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique",
		"Symbol", "ZapfDingbats":
		return props.FontName
	}

	// 3. Fallback logic: map unknown fonts to Helvetica family
	var fallbackName string
	if props.Bold && props.Italic {
		fallbackName = "Helvetica-BoldOblique"
	} else if props.Bold {
		fallbackName = "Helvetica-Bold"
	} else if props.Italic {
		fallbackName = "Helvetica-Oblique"
	} else {
		fallbackName = "Helvetica"
	}

	return fallbackName
}
