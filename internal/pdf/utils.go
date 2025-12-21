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
	if len(parts) < 5 {
		return models.Props{
			FontName:  "Helvetica",
			FontSize:  12,
			StyleCode: "000",
			Bold:      false,
			Italic:    false,
			Underline: false,
			Alignment: "left",
			Borders:   [4]int{0, 0, 0, 0},
		}
	}

	fontSize, _ := strconv.Atoi(parts[1])
	if fontSize == 0 {
		fontSize = 12
	}

	// Parse 3-digit style code (bold:italic:underline)
	styleCode := parts[2]
	if len(styleCode) != 3 {
		styleCode = "000" // Default to normal text
	}

	bold := styleCode[0] == '1'
	italic := styleCode[1] == '1'
	underline := styleCode[2] == '1'

	// Parse alignment (now at index 3)
	alignment := "left"
	if len(parts) > 3 {
		alignment = parts[3]
	}

	// Parse borders (now starting at index 4)
	borders := [4]int{0, 0, 0, 0}
	if len(parts) >= 8 {
		for i := 4; i < 8 && i < len(parts); i++ {
			borders[i-4], _ = strconv.Atoi(parts[i])
		}
	}

	return models.Props{
		FontName:  parts[0],
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

// getFontReference returns the appropriate font reference based on the font name
// and style properties. If a specific font name is provided (e.g., "Helvetica-Bold"),
// it takes precedence. Otherwise, falls back to using bold/italic style flags.
func getFontReference(props models.Props) string {
	// Check if FontName directly specifies a font variant
	switch props.FontName {
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

	// Fallback: use bold/italic flags (for legacy "font1", "font2" style IDs)
	if props.Bold && props.Italic {
		return "/F4" // Helvetica-BoldOblique
	} else if props.Bold {
		return "/F2" // Helvetica-Bold
	} else if props.Italic {
		return "/F3" // Helvetica-Oblique
	}
	return "/F1" // Helvetica (normal)
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
