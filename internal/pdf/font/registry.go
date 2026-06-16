package font

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// CustomFontRegistry manages loaded custom fonts for PDF generation
type CustomFontRegistry struct {
	mu     sync.RWMutex
	fonts  map[string]*RegisteredFont
	noLock bool
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
	CachedRef     string            // Cached PDF reference string (e.g., "/CF1")
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
		return errors.New("failed to load font from " + path + ": " + err.Error())
	}

	return r.RegisterFont(name, font)
}

// RegisterFontFromData loads and registers a TTF/OTF font from raw bytes
func (r *CustomFontRegistry) RegisterFontFromData(name string, data []byte) error {
	font, err := LoadTTFFromData(data)
	if err != nil {
		return errors.New("failed to parse font data: " + err.Error())
	}

	return r.RegisterFont(name, font)
}

// RegisterFontFromBase64 loads and registers a TTF/OTF font from base64-encoded data
func (r *CustomFontRegistry) RegisterFontFromBase64(name string, base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return errors.New("failed to decode base64 font data: " + err.Error())
	}

	return r.RegisterFontFromData(name, data)
}

// RegisterFont registers a parsed TTFFont with a given name
func (r *CustomFontRegistry) RegisterFont(name string, font *TTFFont) error {
	if !r.noLock {
		r.mu.Lock()
	}

	r.fonts[name] = &RegisteredFont{
		Name:      name,
		Font:      font,
		UsedChars: make(map[rune]bool),
	}

	if !r.noLock {
		r.mu.Unlock()
	}
	return nil
}

// GetFont returns a registered font by name
func (r *CustomFontRegistry) GetFont(name string) (*RegisteredFont, bool) {
	if !r.noLock {
		r.mu.RLock()
	}

	font, ok := r.fonts[name]
	if !r.noLock {
		r.mu.RUnlock()
	}
	return font, ok
}

// HasFont checks if a font is registered
func (r *CustomFontRegistry) HasFont(name string) bool {
	if !r.noLock {
		r.mu.RLock()
	}
	_, ok := r.fonts[name]
	if !r.noLock {
		r.mu.RUnlock()
	}
	return ok
}

// MarkCharsUsed marks characters as used for a font (for subsetting)
func (r *CustomFontRegistry) MarkCharsUsed(name string, text string) {
	if !r.noLock {
		r.mu.Lock()
	}

	if font, ok := r.fonts[name]; ok {
		for _, char := range text {
			font.UsedChars[char] = true
		}
	}

	if !r.noLock {
		r.mu.Unlock()
	}
}

// GenerateSubsets generates subset fonts for all registered fonts with used characters
func (r *CustomFontRegistry) GenerateSubsets() error {
	r.mu.Lock()

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

		// C6: reuse cached subset for identical glyph sets (common in benchmark/workload templates).
		if cached, ok := lookupCachedSubset(font.Font, usedGlyphs); ok {
			font.SubsetData = cached.data
			font.OldToNewGlyph = make(map[uint16]uint16, len(cached.oldToNew))
			for k, v := range cached.oldToNew {
				font.OldToNewGlyph[k] = v
			}
			continue
		}

		subsetData, oldToNew, err := SubsetTTF(font.Font, usedGlyphs)
		if err != nil {
			r.mu.Unlock()
			return fmt.Errorf("failed to subset font %s: %w", name, err)
		}

		font.SubsetData = subsetData
		font.OldToNewGlyph = oldToNew
		storeCachedSubset(font.Font, usedGlyphs, subsetData, oldToNew)
	}

	r.mu.Unlock()
	return nil
}

// GetAllFonts returns all registered fonts
func (r *CustomFontRegistry) GetAllFonts() []*RegisteredFont {
	r.mu.RLock()
	fonts := make([]*RegisteredFont, 0, len(r.fonts))
	for _, font := range r.fonts {
		fonts = append(fonts, font)
	}
	r.mu.RUnlock()
	return fonts
}

// GetUsedFonts returns fonts that have characters marked as used
func (r *CustomFontRegistry) GetUsedFonts() []*RegisteredFont {
	r.mu.RLock()
	fonts := make([]*RegisteredFont, 0)
	for _, font := range r.fonts {
		if len(font.UsedChars) > 0 {
			fonts = append(fonts, font)
		}
	}
	r.mu.RUnlock()
	return fonts
}

// Clear removes all registered fonts
func (r *CustomFontRegistry) Clear() {
	r.mu.Lock()
	r.fonts = make(map[string]*RegisteredFont)
	r.mu.Unlock()
}

// ResetUsage clears usage data for all fonts (call before generating a new PDF)
func (r *CustomFontRegistry) ResetUsage() {
	r.mu.Lock()
	for _, font := range r.fonts {
		font.UsedChars = make(map[rune]bool, 256)
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
	r.mu.Unlock()
}

var registryClonePool sync.Pool

// CloneForGeneration creates a shallow clone of the registry with reset usage data.
// This allows concurrent PDF generation without race conditions on UsedChars.
func (r *CustomFontRegistry) CloneForGeneration() *CustomFontRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone, _ := registryClonePool.Get().(*CustomFontRegistry)
	if clone == nil {
		clone = &CustomFontRegistry{
			fonts:  make(map[string]*RegisteredFont, len(r.fonts)),
			noLock: true,
		}
	}
	if len(clone.fonts) < len(r.fonts) {
		clone.fonts = make(map[string]*RegisteredFont, len(r.fonts))
	} else {
		for name := range clone.fonts {
			delete(clone.fonts, name)
		}
	}
	for name, font := range r.fonts {
		clone.fonts[name] = &RegisteredFont{
			Name:      font.Name,
			Font:      font.Font,
			UsedChars: make(map[rune]bool, 256),
		}
	}
	return clone
}

// ReleaseGenerationClone returns a per-generation clone to the pool.
func (r *CustomFontRegistry) ReleaseGenerationClone(clone *CustomFontRegistry) {
	if clone == nil || !clone.noLock {
		return
	}
	registryClonePool.Put(clone)
}

// AssignObjectIDs assigns PDF object IDs to font objects
// Returns the next available object ID
func (r *CustomFontRegistry) AssignObjectIDs(startID int) int {
	r.mu.Lock()

	currentID := startID
	var refBuf [8]byte
	for _, font := range r.fonts {
		// Assign IDs to ALL fonts, even if unused (so GetFontReference works during generation)
		// We will filter out unused fonts when generating resources and embedding.

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
		// Cache the reference string
		font.CachedRef = "/CF" + string(strconv.AppendInt(refBuf[:0], int64(font.ObjectID), 10))
	}

	r.mu.Unlock()
	return currentID
}

// GetFontReference returns the PDF reference string for a custom font (e.g., "/CF1")
func (r *CustomFontRegistry) GetFontReference(name string) string {
	if !r.noLock {
		r.mu.RLock()
	}

	var ref string
	if font, ok := r.fonts[name]; ok {
		// Use cached reference if available (populated by AssignObjectIDs or JIT)
		if font.CachedRef != "" {
			ref = font.CachedRef
		} else if font.ObjectID > 0 {
			// Fallback (shouldn't happen if AssignObjectIDs called)
			ref = fmt.Sprintf("/CF%d", font.ObjectID)
		}
	}

	if !r.noLock {
		r.mu.RUnlock()
	}
	return ref
}

const (
	fontHelvetica            = "Helvetica"
	fontHelveticaBold        = "Helvetica-Bold"
	fontHelveticaOblique     = "Helvetica-Oblique"
	fontHelveticaBoldOblique = "Helvetica-BoldOblique"
)

// standardFontSet is pre-allocated once to avoid per-call map allocation in IsCustomFont.
var standardFontSet = map[string]bool{
	fontHelvetica: true, fontHelveticaBold: true, fontHelveticaOblique: true, fontHelveticaBoldOblique: true,
	"Times-Roman": true, "Times-Bold": true, "Times-Italic": true, "Times-BoldItalic": true,
	"Courier": true, "Courier-Bold": true, "Courier-Oblique": true, "Courier-BoldOblique": true,
	"Symbol": true, "ZapfDingbats": true,
	"font1": true, "font2": true, // Legacy font references
}

// IsCustomFont checks if a font name refers to a custom font (not a standard PDF font)
func IsCustomFont(fontName string) bool {
	return !standardFontSet[fontName]
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

		fontName := name
		if len(name) > len(ext) && name[len(name)-len(ext):] == ext {
			fontName = name[:len(name)-len(ext)]
		}
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

	font, ok := r.fonts[fontName]
	if !ok {
		r.mu.RUnlock()
		return 0
	}

	var totalWidth float64
	for _, char := range text {
		width := font.Font.GetCharWidthScaled(char)
		totalWidth += float64(width) / 1000.0
	}

	r.mu.RUnlock()
	return totalWidth
}

// GetScaledTextWidth calculates the width of text at a specific font size
func (r *CustomFontRegistry) GetScaledTextWidth(fontName string, text string, fontSize float64) float64 {
	return r.GetTextWidth(fontName, text) * fontSize
}

// GeneratePDFFontResources generates the PDF font resource dictionary entries for custom fonts
func (r *CustomFontRegistry) GeneratePDFFontResources() string {
	r.mu.RLock()

	var resources strings.Builder
	var idBuf [8]byte
	for _, font := range r.fonts {
		// Only output resources for fonts that were actually used
		if font.ObjectID > 0 && len(font.UsedChars) > 0 {
			idStr := string(strconv.AppendInt(idBuf[:0], int64(font.ObjectID), 10))
			if font.CachedRef != "" {
				resources.WriteString(" ")
				resources.WriteString(font.CachedRef)
				resources.WriteString(" ")
				resources.WriteString(idStr)
				resources.WriteString(" 0 R")
			} else {
				resources.WriteString(" /CF")
				resources.WriteString(idStr)
				resources.WriteString(" ")
				resources.WriteString(idStr)
				resources.WriteString(" 0 R")
			}
		}
	}

	r.mu.RUnlock()
	return resources.String()
}

// IsCustomFont checks if the font name refers to a registered custom font
func (r *CustomFontRegistry) IsCustomFont(fontName string) bool {
	r.mu.RLock()
	_, ok := r.fonts[fontName]
	r.mu.RUnlock()
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
	case fontHelvetica, fontHelveticaBold, fontHelveticaOblique, fontHelveticaBoldOblique,
		"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic",
		"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique",
		"Symbol", "ZapfDingbats":
		return props.FontName
	}

	// 3. Fallback logic: map unknown fonts to Helvetica family
	var fallbackName string
	switch {
	case props.Bold && props.Italic:
		fallbackName = fontHelveticaBoldOblique
	case props.Bold:
		fallbackName = fontHelveticaBold
	case props.Italic:
		fallbackName = fontHelveticaOblique
	default:
		fallbackName = fontHelvetica
	}

	return fallbackName
}
