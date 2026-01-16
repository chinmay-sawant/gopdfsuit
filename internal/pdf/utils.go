package pdf

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// parseHexColor parses a hexadecimal color string and returns RGB values (0.0-1.0)
// Supports formats: "#RRGGBB", "#RRGGBBAA", "RRGGBB", "RRGGBBAA"
// Returns r, g, b, a (alpha) and a boolean indicating if the color is valid and non-transparent
func parseHexColor(hexColor string) (r, g, b, a float64, valid bool) {
	if hexColor == "" {
		return 0, 0, 0, 0, false
	}

	// Remove # prefix if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Check length
	if len(hexColor) != 6 && len(hexColor) != 8 {
		return 0, 0, 0, 0, false
	}

	// Parse RGB values
	rVal, err := strconv.ParseInt(hexColor[0:2], 16, 64)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	gVal, err := strconv.ParseInt(hexColor[2:4], 16, 64)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	bVal, err := strconv.ParseInt(hexColor[4:6], 16, 64)
	if err != nil {
		return 0, 0, 0, 0, false
	}

	r = float64(rVal) / 255.0
	g = float64(gVal) / 255.0
	b = float64(bVal) / 255.0
	a = 1.0

	// Parse alpha if present
	if len(hexColor) == 8 {
		aVal, err := strconv.ParseInt(hexColor[6:8], 16, 64)
		if err != nil {
			return 0, 0, 0, 0, false
		}
		a = float64(aVal) / 255.0
	}

	// Consider fully transparent colors as "not valid" for rendering
	if a == 0 {
		return 0, 0, 0, 0, false
	}

	return r, g, b, a, true
}

func parseProps(props string) models.Props {
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

	return models.Props{
		FontName:  fontName,
		FontSize:  fontSize,
		StyleCode: styleCode,
		Bold:      bold,
		Italic:    italic,
		Underline: underline,
		Alignment: alignment,
		Borders:   borders,
	}
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
func resolveFontName(props models.Props) string {
	registry := GetFontRegistry()

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
func getFontReference(props models.Props) string {
	registry := GetFontRegistry()

	// Resolve usage to actual font name (handling fallbacks)
	actualFontName := resolveFontName(props)

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
func getWidgetFontReference() string {
	registry := GetFontRegistry()
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
func getWidgetFontName() string {
	registry := GetFontRegistry()
	// If Helvetica is a custom font, we use page-level font resources
	if registry.HasFont("Helvetica") {
		return "" // Signal to use page-level resources
	}
	return "Helvetica" // Use inline Helvetica definition
}

// getWidgetFontObjectID returns the PDF object ID for the widget font.
// In PDF/A mode, this returns the object ID of the Liberation font that replaces Helvetica.
// Returns 0 if no custom font is registered (standard mode).
func getWidgetFontObjectID() int {
	registry := GetFontRegistry()
	if registry.HasFont("Helvetica") {
		if font, ok := registry.GetFont("Helvetica"); ok {
			return font.ObjectID
		}
	}
	return 0
}

// formatPageKids formats the page object IDs for the Pages object
func formatPageKids(pageIDs []int) string {
	var kids []string
	for _, id := range pageIDs {
		kids = append(kids, strconv.Itoa(id)+" 0 R")
	}
	return strings.Join(kids, " ")
}

// escapePDFString escapes characters as required for PDF literal strings.
func escapePDFString(s string) string {
	var sb strings.Builder
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

// isCustomFont checks if the font name refers to a registered custom font
func isCustomFont(fontName string) bool {
	registry := GetFontRegistry()
	return registry.HasFont(fontName)
}

// markFontUsage marks characters as used for font subsetting
// props is used to resolve the actual font (handling fallbacks)
func markFontUsage(props models.Props, text string) {
	resolvedName := resolveFontName(props)
	if isCustomFont(resolvedName) {
		registry := GetFontRegistry()
		registry.MarkCharsUsed(resolvedName, text)
	}
}

// EstimateTextWidth estimates the width of text in points for a given font and size
// Uses actual glyph widths for custom fonts, approximation for standard fonts
func EstimateTextWidth(fontName string, text string, fontSize float64) float64 {
	// For width estimation, we create a dummy props with just the name
	// This might be slightly inaccurate for bold/italic if falling back, but sufficient for layout
	props := models.Props{FontName: fontName, FontSize: int(fontSize)}
	resolvedName := resolveFontName(props)

	registry := GetFontRegistry()
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

	return float64(len([]rune(text))) * fontSize * avgCharWidth
}

// formatTextForPDF formats text for use in a PDF content stream
// For custom fonts, returns hex-encoded string; for standard fonts, returns escaped literal
func formatTextForPDF(props models.Props, text string) string {
	resolvedName := resolveFontName(props)

	if isCustomFont(resolvedName) {
		return EncodeTextForCustomFont(resolvedName, text)
	}
	return "(" + escapePDFString(text) + ")"
}
