package pdf

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/gin-gonic/gin"
)

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
	s = strings.ReplaceAll(s, "(", "\\(")
	return strings.ReplaceAll(s, ")", "\\)")
}

// --- new watermark drawer (diagonal bottom-left to top-right) ---
func drawWatermark(contentStream *bytes.Buffer, text string, pageDims PageDimensions) {
	if strings.TrimSpace(text) == "" {
		return
	}
	// Proportional font size (fallback minimum)
	fontSize := int(pageDims.Width / 8)
	if fontSize < 40 {
		fontSize = 40
	}
	// Position roughly centered
	x := pageDims.Width * 0.20
	y := pageDims.Height * 0.30

	// 45 degree rotation matrix components
	c := 0.7071
	s := 0.7071

	contentStream.WriteString("q\n")
	// Light gray fill/stroke
	contentStream.WriteString("0.85 0.85 0.85 rg 0.85 0.85 0.85 RG\n")
	contentStream.WriteString("BT\n")
	contentStream.WriteString(fmt.Sprintf("/F1 %d Tf\n", fontSize))
	contentStream.WriteString(fmt.Sprintf("%.4f %.4f %.4f %.4f %.2f %.2f Tm\n", c, s, -s, c, x, y))
	contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", escapeText(text)))
	contentStream.WriteString("ET\nQ\n")
}

// --- new page initializer (border + watermark) ---
func initializePage(contentStream *bytes.Buffer, borderConfig, watermark string, pageDims PageDimensions) {
	drawPageBorder(contentStream, borderConfig, pageDims)
	if watermark != "" {
		drawWatermark(contentStream, watermark, pageDims)
	}
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

// PageManager handles multi-page document generation
type PageManager struct {
	Pages            []int   // List of page object IDs
	CurrentPageIndex int     // Current page index (0-based)
	CurrentYPos      float64 // Current Y position on page
	PageDimensions   PageDimensions
	ContentStreams   []bytes.Buffer // Content for each page
}

// NewPageManager creates a new page manager with initial page
func NewPageManager(pageDims PageDimensions) *PageManager {
	pm := &PageManager{
		Pages:            []int{3}, // First page starts at object 3
		CurrentPageIndex: 0,        // Start with first page
		CurrentYPos:      pageDims.Height - margin,
		PageDimensions:   pageDims,
		ContentStreams:   make([]bytes.Buffer, 1),
	}
	return pm
}

// AddNewPage creates a new page when current page is full
func (pm *PageManager) AddNewPage() {
	// Calculate next page object ID
	nextPageID := 3 + len(pm.Pages) // Sequential page IDs starting from 3
	pm.Pages = append(pm.Pages, nextPageID)
	pm.CurrentPageIndex = len(pm.Pages) - 1 // Move to new page
	pm.CurrentYPos = pm.PageDimensions.Height - margin
	pm.ContentStreams = append(pm.ContentStreams, bytes.Buffer{})
}

// CheckPageBreak determines if a new page is needed based on required height
func (pm *PageManager) CheckPageBreak(requiredHeight float64) bool {
	return pm.CurrentYPos-requiredHeight < margin
}

// GetCurrentContentStream returns the current page's content stream
func (pm *PageManager) GetCurrentContentStream() *bytes.Buffer {
	return &pm.ContentStreams[pm.CurrentPageIndex]
}

// GetCurrentPageID returns the current page object ID
func (pm *PageManager) GetCurrentPageID() int {
	return pm.Pages[pm.CurrentPageIndex]
}

// GenerateTemplatePDF generates a PDF document with multi-page support based on a template
func GenerateTemplatePDF(c *gin.Context, template models.PDFTemplate) {
	var pdfBuffer bytes.Buffer
	xrefOffsets := make(map[int]int)

	// Get page dimensions from config
	pageConfig := template.Config
	pageDims := getPageDimensions(pageConfig.Page, pageConfig.PageAlignment)

	// Initialize page manager
	pageManager := NewPageManager(pageDims)

	// PDF Header
	pdfBuffer.WriteString("%PDF-1.7\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Generate all content first to know how many pages we need
	generateAllContent(template, pageManager)

	// Object 2: Pages (will be updated after we know total page count)
	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n")
	pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>\n",
		formatPageKids(pageManager.Pages), len(pageManager.Pages)))
	pdfBuffer.WriteString("endobj\n")

	// Calculate object IDs
	totalPages := len(pageManager.Pages)
	contentObjectStart := totalPages + 3               // Content objects start after pages
	fontObjectStart := contentObjectStart + totalPages // Fonts start after content

	// Generate page objects
	for i, pageID := range pageManager.Pages {
		xrefOffsets[pageID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", pageID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] ",
			pageDims.Width, pageDims.Height))
		pdfBuffer.WriteString(fmt.Sprintf("/Contents %d 0 R ", contentObjectStart+i))
		pdfBuffer.WriteString(fmt.Sprintf("/Resources << /Font << /F1 %d 0 R /F2 %d 0 R /F3 %d 0 R /F4 %d 0 R >> >> >>\n",
			fontObjectStart, fontObjectStart+1, fontObjectStart+2, fontObjectStart+3))
		pdfBuffer.WriteString("endobj\n")
	}

	// Generate content stream objects
	for i, contentStream := range pageManager.ContentStreams {
		objectID := contentObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Length %d >>\n", contentStream.Len()))
		pdfBuffer.WriteString("stream\n")
		pdfBuffer.Write(contentStream.Bytes())
		pdfBuffer.WriteString("\nendstream\nendobj\n")
	}

	// Generate font objects
	fontNames := []string{"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique"}
	fontRefs := []string{"/F1", "/F2", "/F3", "/F4"}

	for i, fontName := range fontNames {
		objectID := fontObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Font /Subtype /Type1 /Name %s /BaseFont /%s >>\n",
			fontRefs[i], fontName))
		pdfBuffer.WriteString("endobj\n")
	}

	// Cross-reference table
	totalObjects := fontObjectStart + 4
	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", totalObjects))
	for i := 1; i < totalObjects; i++ {
		pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", xrefOffsets[i]))
	}

	// Trailer
	pdfBuffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", totalObjects))
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	// HTTP Response
	filename := fmt.Sprintf("template-pdf-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}

// generateAllContent processes the template and generates content across multiple pages
func generateAllContent(template models.PDFTemplate, pageManager *PageManager) {
	// Initialize first page
	initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)

	// Title - Check if it fits on current page
	titleProps := parseProps(template.Title.Props)
	titleHeight := float64(titleProps.FontSize + 50) // Title + spacing

	if pageManager.CheckPageBreak(titleHeight) {
		pageManager.AddNewPage()
		initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)
	}

	drawTitle(pageManager.GetCurrentContentStream(), template.Title, titleProps, pageManager)

	// Tables - Process each table with automatic page breaks
	for _, table := range template.Table {
		drawTable(table, pageManager, template.Config.PageBorder, template.Config.Watermark)
	}

	// Draw footer and page numbers on every page (footer first to avoid overlap)
	totalPages := len(pageManager.Pages)
	for i := 0; i < totalPages; i++ {
		// Draw footer on this page if footer text provided
		if template.Footer.Text != "" {
			drawFooter(&pageManager.ContentStreams[i], template.Footer, pageManager.PageDimensions)
		}
		// Draw page number on this page
		drawPageNumber(&pageManager.ContentStreams[i], i+1, totalPages, pageManager.PageDimensions)
	}
}

// drawPageBorder draws the page border
func drawPageBorder(contentStream *bytes.Buffer, borderConfig string, pageDims PageDimensions) {
	pageBorders := parseBorders(borderConfig)
	if pageBorders[0] > 0 || pageBorders[1] > 0 || pageBorders[2] > 0 || pageBorders[3] > 0 {
		contentStream.WriteString("q\n")
		if pageBorders[0] > 0 { // left border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[0])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(margin), float64(margin), float64(margin), pageDims.Height-margin))
		}
		if pageBorders[1] > 0 { // right border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[1])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				pageDims.Width-margin, float64(margin), pageDims.Width-margin, pageDims.Height-margin))
		}
		if pageBorders[2] > 0 { // top border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[2])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(margin), pageDims.Height-margin, pageDims.Width-margin, pageDims.Height-margin))
		}
		if pageBorders[3] > 0 { // bottom border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[3])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(margin), float64(margin), pageDims.Width-margin, float64(margin)))
		}
		contentStream.WriteString("Q\n")
	}
}

// drawTitle renders the document title
func drawTitle(contentStream *bytes.Buffer, title models.Title, titleProps models.Props, pageManager *PageManager) {
	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(titleProps))
	contentStream.WriteString(" ")
	contentStream.WriteString(strconv.Itoa(titleProps.FontSize))
	contentStream.WriteString(" Tf\n")

	var titleX float64
	switch titleProps.Alignment {
	case "center":
		titleX = pageManager.PageDimensions.Width / 2
	case "right":
		titleX = pageManager.PageDimensions.Width - margin
	default:
		titleX = margin
	}

	pageManager.CurrentYPos -= float64(titleProps.FontSize + 20)
	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", titleX, pageManager.CurrentYPos))
	contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", title.Text))
	contentStream.WriteString("ET\n")

	pageManager.CurrentYPos -= 30
}

// drawTable renders a table with automatic page breaks
func drawTable(table models.Table, pageManager *PageManager, borderConfig, watermark string) {
	cellWidth := (pageManager.PageDimensions.Width - 2*margin) / float64(table.MaxColumns)
	rowHeight := float64(25) // Standard row height

	for _, row := range table.Rows {
		// Check if row fits on current page
		if pageManager.CheckPageBreak(rowHeight) {
			// Create new page and initialize it
			pageManager.AddNewPage()
			initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions)
		}

		// Get current content stream for this page
		contentStream := pageManager.GetCurrentContentStream()

		// Draw row cells
		for colIndex, cell := range row.Row {
			if colIndex >= table.MaxColumns {
				break
			}

			cellProps := parseProps(cell.Props)
			cellX := float64(margin) + float64(colIndex)*cellWidth

			// Draw cell borders
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[0]), cellX, pageManager.CurrentYPos-rowHeight, cellX, pageManager.CurrentYPos))
				}
				if cellProps.Borders[1] > 0 { // right
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[1]), cellX+cellWidth, pageManager.CurrentYPos-rowHeight, cellX+cellWidth, pageManager.CurrentYPos))
				}
				if cellProps.Borders[2] > 0 { // top
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[2]), cellX, pageManager.CurrentYPos, cellX+cellWidth, pageManager.CurrentYPos))
				}
				if cellProps.Borders[3] > 0 { // bottom
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[3]), cellX, pageManager.CurrentYPos-rowHeight, cellX+cellWidth, pageManager.CurrentYPos-rowHeight))
				}
				contentStream.WriteString("Q\n")
			}

			// Draw text or checkbox
			if cell.Checkbox != nil {
				// Draw checkbox
				checkboxSize := 10.0
				checkboxX := cellX + (cellWidth-checkboxSize)/2
				checkboxY := pageManager.CurrentYPos - (rowHeight+checkboxSize)/2

				contentStream.WriteString("q\n")
				contentStream.WriteString("1 w\n")
				contentStream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re S\n",
					checkboxX, checkboxY, checkboxSize, checkboxSize))

				if *cell.Checkbox {
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l %.2f %.2f m %.2f %.2f l S\n",
						checkboxX+2, checkboxY+2, checkboxX+checkboxSize-2, checkboxY+checkboxSize-2,
						checkboxX+checkboxSize-2, checkboxY+2, checkboxX+2, checkboxY+checkboxSize-2))
				}
				contentStream.WriteString("Q\n")
			} else if cell.Text != "" {
				// Draw text with font styling
				contentStream.WriteString("BT\n")
				contentStream.WriteString(getFontReference(cellProps))
				contentStream.WriteString(" ")
				contentStream.WriteString(strconv.Itoa(cellProps.FontSize))
				contentStream.WriteString(" Tf\n")

				var textX float64
				switch cellProps.Alignment {
				case "center":
					textX = cellX + cellWidth/2
				case "right":
					textX = cellX + cellWidth - 5
				default:
					textX = cellX + 5
				}

				textY := pageManager.CurrentYPos - rowHeight/2 - float64(cellProps.FontSize)/2

				// Reset text matrix and position absolutely
				contentStream.WriteString("1 0 0 1 0 0 Tm\n")
				contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))

				// Add underline support
				if cellProps.Underline {
					// End text object before drawing underline
					contentStream.WriteString("ET\n")
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					underlineY := textY - 2
					textWidth := float64(len(cell.Text) * cellProps.FontSize / 2)
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
						textX, underlineY, textX+textWidth, underlineY))
					contentStream.WriteString("Q\n")
					// Start text object again
					contentStream.WriteString("BT\n")
					contentStream.WriteString(getFontReference(cellProps))
					contentStream.WriteString(" ")
					contentStream.WriteString(strconv.Itoa(cellProps.FontSize))
					contentStream.WriteString(" Tf\n")
					contentStream.WriteString("1 0 0 1 0 0 Tm\n")
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))
				}

				contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", cell.Text))
				contentStream.WriteString("ET\n")
			}
		}

		pageManager.CurrentYPos -= rowHeight
	}

	pageManager.CurrentYPos -= 20 // Space between tables
}

// drawFooter renders the document footer
func drawFooter(contentStream *bytes.Buffer, footer models.Footer, pageDims PageDimensions) {
	footerProps := parseProps(footer.Font)
	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(footerProps))
	contentStream.WriteString(" ")
	contentStream.WriteString(strconv.Itoa(footerProps.FontSize))
	contentStream.WriteString(" Tf\n")

	// Position footer outside the page border on the left side
	footerX := float64(20) // 20pt from left edge (outside margin)
	footerY := float64(20) // 20pt from bottom edge (outside margin)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", footerX, footerY))
	contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", footer.Text))
	contentStream.WriteString("ET\n")
}

// drawPageNumber renders page number in bottom right corner
func drawPageNumber(contentStream *bytes.Buffer, currentPage, totalPages int, pageDims PageDimensions) {
	pageText := fmt.Sprintf("Page %d of %d", currentPage, totalPages)

	contentStream.WriteString("BT\n")
	contentStream.WriteString("/F1 10 Tf\n") // Use Helvetica, 10pt

	// Calculate text width for proper right alignment
	textWidth := float64(len(pageText)) * 6 // Approximate character width for 10pt font

	// Position outside the page border on the right side
	pageNumberX := pageDims.Width - textWidth - 20 // 20pt from right edge (outside margin)
	pageNumberY := float64(20)                     // 20pt from bottom edge (outside margin)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", pageNumberX, pageNumberY))
	contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", pageText))
	contentStream.WriteString("ET\n")
}

// formatPageKids formats the page object IDs for the Pages object
func formatPageKids(pageIDs []int) string {
	var kids []string
	for _, id := range pageIDs {
		kids = append(kids, fmt.Sprintf("%d 0 R", id))
	}
	return strings.Join(kids, " ")
}
