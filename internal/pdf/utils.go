package pdf

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

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

// getFontReference returns the appropriate font reference based on style properties
func getFontReference(props models.Props) string {
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

// escapePDFString escapes parentheses and backslashes for PDF literal strings
func escapePDFString(s string) string {
	// Produce a PDF-literal-safe string where:
	// - existing escapes for parentheses (i.e. "\(" or "\)") are preserved
	// - other backslashes are escaped to "\\"
	// - unescaped parentheses are prefixed with a backslash
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\\' {
			// If backslash is followed by a parenthesis, keep as an escape for that paren
			if i+1 < len(s) && (s[i+1] == '(' || s[i+1] == ')') {
				b.WriteByte('\\')
				b.WriteByte(s[i+1])
				i++ // skip next
				continue
			}
			// Otherwise escape the backslash itself
			b.WriteString("\\\\")
			continue
		}
		if ch == '(' || ch == ')' {
			// not already escaped (previous char wasn't backslash) -> escape it
			b.WriteByte('\\')
			b.WriteByte(ch)
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}
