package pdf

import "strings"

// Page size constants in points (1 inch = 72 points)
var pageSizes = map[string][2]float64{
	"A4":     {595, 842},  // A4: 8.27 × 11.69 inches
	"LETTER": {612, 792},  // Letter: 8.5 × 11 inches
	"LEGAL":  {612, 1008}, // Legal: 8.5 × 14 inches
	"A3":     {842, 1191}, // A3: 11.69 × 16.54 inches
	"A5":     {420, 595},  // A5: 5.83 × 8.27 inches
}

const margin = 72 // Standard 1 inch margin

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
