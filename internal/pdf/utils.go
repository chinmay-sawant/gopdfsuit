package pdf

import (
	"strconv"
	"strings"
	"sync"

	"unicode/utf8"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// hexNibble maps ASCII byte to hex value (0-15). 0xFF = invalid.
var hexNibble [256]byte

func init() {
	for i := range hexNibble {
		hexNibble[i] = 0xFF
	}
	for i := byte('0'); i <= '9'; i++ {
		hexNibble[i] = i - '0'
	}
	for i := byte('a'); i <= 'f'; i++ {
		hexNibble[i] = i - 'a' + 10
	}
	for i := byte('A'); i <= 'F'; i++ {
		hexNibble[i] = i - 'A' + 10
	}
}

// parseHexColor parses a hexadecimal color string and returns RGB values (0.0-1.0)
// Supports formats: "#RRGGBB", "#RRGGBBAA", "RRGGBB", "RRGGBBAA"
// Returns r, g, b, a (alpha) and a boolean indicating if the color is valid and non-transparent
// Uses inline hex nibble lookup instead of strconv.ParseInt for speed.
func parseHexColor(hexColor string) (r, g, b, a float64, valid bool) {
	if hexColor == "" {
		return 0, 0, 0, 0, false
	}

	// Remove # prefix if present
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Check length
	if len(hexColor) != 6 && len(hexColor) != 8 {
		return 0, 0, 0, 0, false
	}

	// Inline hex decode using lookup table (avoids strconv.ParseInt overhead)
	h0, h1 := hexNibble[hexColor[0]], hexNibble[hexColor[1]]
	h2, h3 := hexNibble[hexColor[2]], hexNibble[hexColor[3]]
	h4, h5 := hexNibble[hexColor[4]], hexNibble[hexColor[5]]
	if h0|h1|h2|h3|h4|h5 == 0xFF {
		return 0, 0, 0, 0, false
	}

	r = float64(h0<<4|h1) / 255.0
	g = float64(h2<<4|h3) / 255.0
	b = float64(h4<<4|h5) / 255.0
	a = 1.0

	// Parse alpha if present
	if len(hexColor) == 8 {
		h6, h7 := hexNibble[hexColor[6]], hexNibble[hexColor[7]]
		if h6|h7 == 0xFF {
			return 0, 0, 0, 0, false
		}
		a = float64(h6<<4|h7) / 255.0
	}

	// Consider fully transparent colors as "not valid" for rendering
	if a == 0 {
		return 0, 0, 0, 0, false
	}

	return r, g, b, a, true
}

// propsCache memoizes parseProps results since the same prop strings are parsed repeatedly.
var propsCache sync.Map // string -> models.Props

func parseProps(props string) models.Props {
	// Fast path: check cache
	if cached, ok := propsCache.Load(props); ok {
		return cached.(models.Props)
	}

	parts := strings.Split(props, ":")

	// Default values
	fontName := "Helvetica"
	fontSize := 12
	styleCode := "000"
	alignment := "left"
	borders := [4]int{0, 0, 0, 0}

	// Helper to safe index
	getPart := func(idx int) string {
		if idx < len(parts) {
			return parts[idx]
		}
		return ""
	}

	// 1. Font Name
	if name := strings.TrimSpace(getPart(0)); name != "" {
		fontName = name
	}

	// 2. Font Size
	if sizeStr := getPart(1); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			fontSize = s
		}
	}

	// 3. Style Code
	if sc := getPart(2); len(sc) == 3 {
		styleCode = sc
	}

	bold := styleCode[0] == '1'
	italic := styleCode[1] == '1'
	underline := styleCode[2] == '1'

	// 4. Alignment
	if align := getPart(3); align != "" {
		alignment = align
	}

	// 5-8. Borders
	if len(parts) >= 8 {
		for i := 4; i < 8; i++ {
			if val, err := strconv.Atoi(parts[i]); err == nil {
				borders[i-4] = val
			}
		}
	} else if len(parts) >= 5 {
		// Try to parse as many borders as available starting from index 4
		for i := 4; i < len(parts) && i < 8; i++ {
			if val, err := strconv.Atoi(parts[i]); err == nil {
				borders[i-4] = val
			}
		}
	}

	result := models.Props{
		FontName:  fontName,
		FontSize:  fontSize,
		StyleCode: styleCode,
		Bold:      bold,
		Italic:    italic,
		Underline: underline,
		Alignment: alignment,
		Borders:   borders,
	}

	// Cache for future calls
	propsCache.Store(props, result)
	return result
}

func parseBorders(borderStr string) [4]int {
	parts := strings.Split(borderStr, ":")
	borders := [4]int{0, 0, 0, 0}
	for i, part := range parts {
		if i < 4 {
			borders[i], _ = strconv.Atoi(part)
		}
	}
	return borders
}

// --- new helper to escape parentheses in text ---
func escapeText(s string) string {
	// Use the same PDF-safe escaping as escapePDFString to avoid duplicating logic.
	return escapePDFString(s)
}

// resolveFontName resolves the actual font name to use, handling fallbacks
func resolveFontName(props models.Props, registry *CustomFontRegistry) string {
	// 1. Check if the requested font is registered as a custom font
	if registry.HasFont(props.FontName) {
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

// getFontReference returns the appropriate font reference based on the font name
// and style properties. If a specific font name is provided (e.g., "Helvetica-Bold"),
// it takes precedence. Otherwise, falls back to using bold/italic style flags.
// For custom fonts, checks the font registry and returns the custom font reference.
func getFontReference(props models.Props, registry *CustomFontRegistry) string {
	// Resolve usage to actual font name (handling fallbacks)
	actualFontName := resolveFontName(props, registry)

	// If resolved font is custom (including PDF/A substitution), use it
	if registry.HasFont(actualFontName) {
		ref := registry.GetFontReference(actualFontName)
		if ref != "" {
			return ref
		}
	}

	// Otherwise, return standard font reference code
	switch actualFontName {
	// Helvetica family (F1-F4)
	case "Helvetica":
		return "/F1"
	case "Helvetica-Bold":
		return "/F2"
	case "Helvetica-Oblique":
		return "/F3"
	case "Helvetica-BoldOblique":
		return "/F4"
	// Times family (F5-F8)
	case "Times-Roman":
		return "/F5"
	case "Times-Bold":
		return "/F6"
	case "Times-Italic":
		return "/F7"
	case "Times-BoldItalic":
		return "/F8"
	// Courier family (F9-F12)
	case "Courier":
		return "/F9"
	case "Courier-Bold":
		return "/F10"
	case "Courier-Oblique":
		return "/F11"
	case "Courier-BoldOblique":
		return "/F12"
	// Symbol fonts (F13-F14)
	case "Symbol":
		return "/F13"
	case "ZapfDingbats":
		return "/F14"
	}

	return "/F1" // Ultimate fallback
}

// getWidgetFontReference returns the appropriate font reference for form field widgets.
// In PDF/A mode (when Helvetica is registered as a custom Liberation font), this returns
// the custom font reference. Otherwise, it returns /F1 for standard Helvetica.
// This should be used in widget DA strings and appearance streams instead of hardcoded /Helv.
func getWidgetFontReference(registry *CustomFontRegistry) string {
	// Check if Helvetica is registered as custom font (PDF/A mode with Liberation)
	if registry.HasFont("Helvetica") {
		ref := registry.GetFontReference("Helvetica")
		if ref != "" {
			return ref
		}
	}
	return "/F1" // Standard Helvetica reference
}

// getWidgetFontName returns the font name to use in widget resource dictionaries.
// In PDF/A mode, widgets should not embed their own font definitions - they should
// reference fonts from the page resources. Returns empty string if using page fonts.
func getWidgetFontName(registry *CustomFontRegistry) string {
	// If Helvetica is a custom font, we use page-level font resources
	if registry.HasFont("Helvetica") {
		return "" // Signal to use page-level resources
	}
	return "Helvetica" // Use inline Helvetica definition
}

// getWidgetFontObjectID returns the PDF object ID for the widget font.
// In PDF/A mode, this returns the object ID of the Liberation font that replaces Helvetica.
// Returns 0 if no custom font is registered (standard mode).
func getWidgetFontObjectID(registry *CustomFontRegistry) int {
	if registry.HasFont("Helvetica") {
		if font, ok := registry.GetFont("Helvetica"); ok {
			return font.ObjectID
		}
	}
	return 0
}

// formatPageKids formats the page object IDs for the Pages object
func formatPageKids(pageIDs []int) string {
	var buf []byte
	for i, id := range pageIDs {
		if i > 0 {
			buf = append(buf, ' ')
		}
		buf = strconv.AppendInt(buf, int64(id), 10)
		buf = append(buf, " 0 R"...)
	}
	return string(buf)
}

// escapePDFString escapes characters as required for PDF literal strings.
func escapePDFString(s string) string {
	// Fast path: most text has no special characters
	if !strings.ContainsAny(s, `()\`) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s) + 4)
	for _, r := range s {
		switch r {
		case '(', ')', '\\':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// isCustomFontCheck checks if the font name refers to a registered custom font
func isCustomFontCheck(fontName string, registry *CustomFontRegistry) bool {
	return registry.HasFont(fontName)
}

// EstimateTextWidth estimates the width of text in points for a given font and size
// Uses actual glyph widths for custom fonts, approximation for standard fonts
// Takes resolvedFontName to avoid repeated lookups
func EstimateTextWidth(resolvedName string, text string, fontSize float64, registry *CustomFontRegistry) float64 {
	if registry.HasFont(resolvedName) {
		return registry.GetScaledTextWidth(resolvedName, text, fontSize)
	}

	// Approximation for standard fonts (average character width ~0.5-0.6 em)
	avgCharWidth := 0.5
	switch resolvedName {
	case "Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique":
		avgCharWidth = 0.6 // Monospace is wider
	case "Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic":
		avgCharWidth = 0.45 // Times is slightly narrower
	}

	// Use utf8.RuneCountInString to avoid allocating rune slice
	return float64(utf8.RuneCountInString(text)) * fontSize * avgCharWidth
}

// formatTextForPDF formats text for use in a PDF content stream
// For custom fonts, returns hex-encoded string; for standard fonts, returns escaped literal
// Accepts a pre-resolved font name to avoid redundant resolveFontName calls.
func formatTextForPDF(resolvedName string, text string, registry *CustomFontRegistry) string {
	if isCustomFontCheck(resolvedName, registry) {
		return EncodeTextForCustomFont(resolvedName, text, registry)
	}
	return "(" + escapePDFString(text) + ")"
}

// WrapText splits text into multiple lines that fit within the specified maxWidth.
// It wraps on word boundaries when possible, and handles long words that exceed maxWidth.
// Returns a slice of strings, each representing one line of text.
func WrapText(text string, resolvedFontName string, fontSize float64, maxWidth float64, registry *CustomFontRegistry) []string {
	if text == "" {
		return []string{""}
	}

	// Handle edge case where maxWidth is too small
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		// Check if word alone exceeds maxWidth (need to break it up)
		wordWidth := EstimateTextWidth(resolvedFontName, word, fontSize, registry)
		if wordWidth > maxWidth {
			// Flush current line first
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			// Break long word into chunks that fit
			lines = append(lines, wrapLongWord(word, resolvedFontName, fontSize, maxWidth, registry)...)
			continue
		}

		// Try adding word to current line
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		testWidth := EstimateTextWidth(resolvedFontName, testLine, fontSize, registry)
		if testWidth <= maxWidth {
			// Fits - add to current line
			currentLine = testLine
		} else {
			// Doesn't fit - start new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	// Don't forget the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Ensure at least one line is returned
	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// wrapLongWord breaks a single word that's too long into multiple lines
func wrapLongWord(word string, resolvedFontName string, fontSize float64, maxWidth float64, registry *CustomFontRegistry) []string {
	var lines []string
	runes := []rune(word)
	start := 0

	for start < len(runes) {
		// Binary search for the maximum number of characters that fit
		end := start + 1
		for end <= len(runes) {
			substr := string(runes[start:end])
			if EstimateTextWidth(resolvedFontName, substr, fontSize, registry) > maxWidth {
				break
			}
			end++
		}

		// Back up one character (the one that caused overflow)
		if end > start+1 {
			end--
		}

		// Ensure we make progress (at least one character per line)
		if end == start {
			end = start + 1
		}

		lines = append(lines, string(runes[start:end]))
		start = end
	}

	return lines
}

// CalculateWrappedTextHeight calculates the total height needed for wrapped text
// lineCount: number of lines of text
// fontSize: font size in points
// lineSpacing: multiplier for line height (e.g., 1.2 for 120% line height)
// Returns the total height in points
func CalculateWrappedTextHeight(lineCount int, fontSize float64, lineSpacing float64) float64 {
	if lineCount <= 0 {
		return 0
	}
	// Height = (number of lines * font size * line spacing)
	// We use (lineCount) because each line takes fontSize * lineSpacing vertical space
	return float64(lineCount) * fontSize * lineSpacing
}
