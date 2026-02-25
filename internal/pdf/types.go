package pdf

import (
	"strconv"
	"strings"
)

// Page size constants in points (1 inch = 72 points)
var pageSizes = map[string][2]float64{
	"A4":     {595, 842},  // A4: 8.27 × 11.69 inches
	"LETTER": {612, 792},  // Letter: 8.5 × 11 inches
	"LEGAL":  {612, 1008}, // Legal: 8.5 × 14 inches
	"A3":     {842, 1191}, // A3: 11.69 × 16.54 inches
	"A5":     {420, 595},  // A5: 5.83 × 8.27 inches
}

const defaultMargin = 72.0 // Standard 1 inch margin

// PageMargins represents the margins (left, right, top, bottom) for a PDF page in points.
type PageMargins struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

// DefaultPageMargins returns the standard 1-inch margins (72 points) for all sides.
func DefaultPageMargins() PageMargins {
	return PageMargins{Left: defaultMargin, Right: defaultMargin, Top: defaultMargin, Bottom: defaultMargin}
}

// ParsePageMargins parses margins in "left:right:top:bottom" points format.
// Missing or invalid values gracefully fall back to defaults.
func ParsePageMargins(margins string) PageMargins {
	parsed := DefaultPageMargins()
	if strings.TrimSpace(margins) == "" {
		return parsed
	}

	parts := strings.Split(margins, ":")
	if len(parts) > 0 {
		if value, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil && value >= 0 {
			parsed.Left = value
		}
	}
	if len(parts) > 1 {
		if value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil && value >= 0 {
			parsed.Right = value
		}
	}
	if len(parts) > 2 {
		if value, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64); err == nil && value >= 0 {
			parsed.Top = value
		}
	}
	if len(parts) > 3 {
		if value, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64); err == nil && value >= 0 {
			parsed.Bottom = value
		}
	}

	return parsed
}

// PageDimensions holds the current page dimensions and orientation
type PageDimensions struct {
	Width  float64
	Height float64
}

// getPageDimensions calculates page dimensions based on size and orientation
func getPageDimensions(pageSize string, orientation int) PageDimensions {
	// Default to A4 if page size not found
	size, exists := pageSizes[strings.ToUpper(pageSize)]
	if !exists {
		size = pageSizes["A4"]
	}

	width, height := size[0], size[1]

	// Orientation: 1 = Portrait (vertical), 2 = Landscape (horizontal)
	if orientation == 2 {
		// Swap width and height for landscape
		width, height = height, width
	}

	return PageDimensions{Width: width, Height: height}
}

// ObjectEncryptor defines the interface for encrypting PDF objects
type ObjectEncryptor interface {
	EncryptString(data []byte, objNum, genNum int) []byte
	EncryptStream(data []byte, objNum, genNum int) []byte
	GetEncryptDictionary(encryptObjID int) string
}
