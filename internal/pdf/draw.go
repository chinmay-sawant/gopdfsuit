package pdf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// fmtNum formats a float with 2 decimal places (standard PDF precision)
func fmtNum(f float64) string {
	return fmt.Sprintf("%.2f", f)
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

	// Track characters for font subsetting
	registry := GetFontRegistry()
	if registry.HasFont("Helvetica") {
		registry.MarkCharsUsed("Helvetica", text)
	}

	// 45 degree rotation matrix components
	c := 0.7071
	s := 0.7071

	// Use props for proper font encoding
	watermarkProps := models.Props{FontName: "Helvetica", FontSize: fontSize}

	contentStream.WriteString("q\n")
	// Light gray fill/stroke
	contentStream.WriteString("0.85 0.85 0.85 rg 0.85 0.85 0.85 RG\n")
	contentStream.WriteString("BT\n")
	// Use getFontReference to handle PDF/A font substitution (Helvetica -> Liberation)
	fontRef := getFontReference(watermarkProps)
	contentStream.WriteString(fmt.Sprintf("%s %d Tf\n", fontRef, fontSize))
	contentStream.WriteString(fmt.Sprintf("%s %s %s %s %s %s Tm\n", fmtNum(c), fmtNum(s), fmtNum(-s), fmtNum(c), fmtNum(x), fmtNum(y)))
	contentStream.WriteString(fmt.Sprintf("%s Tj\n", formatTextForPDF(watermarkProps, text)))
	contentStream.WriteString("ET\nQ\n")
}

// --- new page initializer (border + watermark) ---
func initializePage(contentStream *bytes.Buffer, borderConfig, watermark string, pageDims PageDimensions) {
	drawPageBorder(contentStream, borderConfig, pageDims)
	if watermark != "" {
		drawWatermark(contentStream, watermark, pageDims)
	}
}

// drawPageBorder draws the page border
func drawPageBorder(contentStream *bytes.Buffer, borderConfig string, pageDims PageDimensions) {
	pageBorders := parseBorders(borderConfig)
	if pageBorders[0] > 0 || pageBorders[1] > 0 || pageBorders[2] > 0 || pageBorders[3] > 0 {
		contentStream.WriteString("q\n")
		if pageBorders[0] > 0 { // left border
			contentStream.WriteString(fmt.Sprintf("%d w\n", pageBorders[0]))
			contentStream.WriteString(fmt.Sprintf("%d %d m %d %s l S\n",
				margin, margin, margin, fmtNum(pageDims.Height-margin)))
		}
		if pageBorders[1] > 0 { // right border
			contentStream.WriteString(fmt.Sprintf("%d w\n", pageBorders[1]))
			contentStream.WriteString(fmt.Sprintf("%s %d m %s %s l S\n",
				fmtNum(pageDims.Width-margin), margin, fmtNum(pageDims.Width-margin), fmtNum(pageDims.Height-margin)))
		}
		if pageBorders[2] > 0 { // top border
			contentStream.WriteString(fmt.Sprintf("%d w\n", pageBorders[2]))
			contentStream.WriteString(fmt.Sprintf("%d %s m %s %s l S\n",
				margin, fmtNum(pageDims.Height-margin), fmtNum(pageDims.Width-margin), fmtNum(pageDims.Height-margin)))
		}
		if pageBorders[3] > 0 { // bottom border
			contentStream.WriteString(fmt.Sprintf("%d w\n", pageBorders[3]))
			contentStream.WriteString(fmt.Sprintf("%d %d m %s %d l S\n",
				margin, margin, fmtNum(pageDims.Width-margin), margin))
		}
		contentStream.WriteString("Q\n")
	}
}

// drawTitle renders the document title (either simple text or embedded table)
func drawTitle(contentStream *bytes.Buffer, title models.Title, titleProps models.Props, pageManager *PageManager, cellImageObjectIDs map[string]int) {
	// Check if title has an embedded table
	if title.Table != nil && len(title.Table.Rows) > 0 {
		drawTitleTable(contentStream, title.Table, pageManager, cellImageObjectIDs, title.BgColor, title.TextColor, titleProps)
		return
	}

	// Simple text title
	// Draw background color if specified
	if r, g, b, _, valid := parseHexColor(title.BgColor); valid {
		rectX := float64(margin)
		rectY := pageManager.CurrentYPos - float64(titleProps.FontSize)
		rectW := pageManager.PageDimensions.Width - 2*float64(margin)
		rectH := float64(titleProps.FontSize)

		contentStream.WriteString("q\n")
		contentStream.WriteString(fmt.Sprintf("%s %s %s rg\n", fmtNum(r), fmtNum(g), fmtNum(b)))
		contentStream.WriteString(fmt.Sprintf("%s %s %s %s re f\n",
			fmtNum(rectX), fmtNum(rectY), fmtNum(rectW), fmtNum(rectH)))
		contentStream.WriteString("Q\n")
	}

	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(titleProps))
	contentStream.WriteString(" ")
	contentStream.WriteString(strconv.Itoa(titleProps.FontSize))
	contentStream.WriteString(" Tf\n")

	// Set text color
	if r, g, b, _, valid := parseHexColor(title.TextColor); valid {
		contentStream.WriteString(fmt.Sprintf("%s %s %s rg\n", fmtNum(r), fmtNum(g), fmtNum(b)))
	} else {
		contentStream.WriteString("0 0 0 rg\n")
	}

	// Calculate approximate text width
	textWidth := EstimateTextWidth(titleProps.FontName, title.Text, float64(titleProps.FontSize))

	// Calculate available width (page width minus both margins)
	availableWidth := pageManager.PageDimensions.Width - 2*margin

	var titleX float64
	switch titleProps.Alignment {
	case "center":
		// Center the text within the available area (between margins)
		titleX = margin + (availableWidth-textWidth)/2
	case "right":
		// Right align: position text so it ends at the right margin
		titleX = pageManager.PageDimensions.Width - margin - textWidth
	default:
		titleX = margin
	}

	pageManager.CurrentYPos -= float64(titleProps.FontSize)
	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(titleX), fmtNum(pageManager.CurrentYPos)))
	contentStream.WriteString(fmt.Sprintf("%s Tj\n", formatTextForPDF(titleProps, title.Text)))
	contentStream.WriteString("ET\n")

	// Add Link Annotation if provided
	if title.Link != "" {
		// Calculate approximate bounding box for the text
		// BBox: [titleX, titleY, titleX+textWidth, titleY+fontSize]
		// Use Y pos (baseline) + descent (approx)
		rectX := titleX
		rectY := pageManager.CurrentYPos - float64(titleProps.FontSize)*0.2 // Slightly below baseline
		rectW := textWidth
		rectH := float64(titleProps.FontSize) * 1.2
		pageManager.AddLinkAnnotation(rectX, rectY, rectW, rectH, title.Link)
	}
}

// drawTitleTable renders an embedded table within the title section (no borders by default)
func drawTitleTable(contentStream *bytes.Buffer, table *models.TitleTable, pageManager *PageManager, cellImageObjectIDs map[string]int, defaultBgColor, defaultTextColor string, defaultProps models.Props) {
	availableWidth := (pageManager.PageDimensions.Width - 2*margin)
	baseRowHeight := float64(25) // Standard row height

	// Compute column widths in points using weights if provided
	colWidths := make([]float64, table.MaxColumns)
	if len(table.ColumnWidths) == table.MaxColumns {
		// Normalize weights to sum 1
		var sum float64
		for _, w := range table.ColumnWidths {
			if w > 0 {
				sum += w
			}
		}
		if sum <= 0 {
			for i := range colWidths {
				colWidths[i] = availableWidth / float64(table.MaxColumns)
			}
		} else {
			for i, w := range table.ColumnWidths {
				if w <= 0 {
					w = 0
				}
				colWidths[i] = (w / sum) * availableWidth
			}
		}
	} else {
		for i := range colWidths {
			colWidths[i] = availableWidth / float64(table.MaxColumns)
		}
	}

	for rowIdx, row := range table.Rows {
		// Determine this row's height
		rowHeight := baseRowHeight
		for _, cell := range row.Row {
			if cell.Height != nil && *cell.Height > rowHeight {
				rowHeight = *cell.Height
			}
		}

		// Draw row cells - Pass 1: Backgrounds
		// We draw all backgrounds first so that if text from one cell overflows into another,
		// it isn't covered by the next cell's background.
		bgX := float64(margin)
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			// Use cell-specific width if provided, otherwise use column width
			cellWidth := colWidths[colIdx]
			if cell.Width != nil && *cell.Width > 0 {
				cellWidth = *cell.Width
			}

			// Use cell-specific height if provided, otherwise use row height
			cellHeight := rowHeight
			if cell.Height != nil && *cell.Height > 0 {
				cellHeight = *cell.Height
			}

			// Draw cell background color
			// Use cell-specific color if available, otherwise use default (title-level) color
			bgColor := cell.BgColor
			if bgColor == "" {
				bgColor = defaultBgColor
			}
			if r, g, b, _, valid := parseHexColor(bgColor); valid {
				contentStream.WriteString("q\n")
				contentStream.WriteString(fmt.Sprintf("%s %s %s rg\n", fmtNum(r), fmtNum(g), fmtNum(b)))
				contentStream.WriteString(fmt.Sprintf("%s %s %s %s re f\n",
					fmtNum(bgX), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellWidth), fmtNum(cellHeight)))
				contentStream.WriteString("Q\n")
			}

			bgX += cellWidth
		}

		// Draw row cells - Pass 2: Content and Borders
		currentX := float64(margin)
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			// Capture cell coordinates for link
			// Capture cell coordinates for link

			var cellProps models.Props
			if cell.Props == "" {
				cellProps = defaultProps
			} else {
				cellProps = parseProps(cell.Props)
			}
			cellX := currentX

			// Use cell-specific width if provided, otherwise use column width
			cellWidth := colWidths[colIdx]
			if cell.Width != nil && *cell.Width > 0 {
				cellWidth = *cell.Width
			}

			// Use cell-specific height if provided, otherwise use row height
			cellHeight := rowHeight
			if cell.Height != nil && *cell.Height > 0 {
				cellHeight = *cell.Height
			}

			// Update X position for next cell
			currentX += cellWidth

			// Draw image first (so borders are drawn on top)
			if cell.Image != nil && cell.Image.ImageData != "" {
				// Check if we have an XObject for this title cell image
				cellKey := fmt.Sprintf("title:%d:%d", rowIdx, colIdx)
				if _, exists := cellImageObjectIDs[cellKey]; exists {
					// Render actual image using XObject - fit inside cell with small padding for border
					borderPadding := 1.0 // Small padding to keep image inside borders
					imgWidth := cellWidth - 2*borderPadding
					imgHeight := cellHeight - 2*borderPadding

					imgX := cellX + borderPadding
					imgY := pageManager.CurrentYPos - cellHeight + borderPadding

					// Draw actual image using XObject with clipping to prevent overflow
					contentStream.WriteString("q\n")
					// Set up clipping rectangle to confine image within cell bounds (with padding) - using 're' operator
					shortKey := strings.ReplaceAll(cellKey, ":", "_")
					contentStream.WriteString(fmt.Sprintf("%s %s %s %s re W n\n",
						fmtNum(imgX), fmtNum(imgY), fmtNum(imgWidth), fmtNum(imgHeight)))
					contentStream.WriteString(fmt.Sprintf("%s 0 0 %s %s %s cm\n",
						fmtNum(imgWidth), fmtNum(imgHeight), fmtNum(imgX), fmtNum(imgY)))
					contentStream.WriteString(fmt.Sprintf("/C%s Do\n", shortKey))
					contentStream.WriteString("Q\n")
				} else {
					// Fall back to placeholder
					imgWidth := cellWidth
					imgHeight := cellHeight
					imgX := cellX
					imgY := pageManager.CurrentYPos - cellHeight

					// Draw placeholder border using 're' operator
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					contentStream.WriteString("0.7 0.7 0.7 RG\n")
					contentStream.WriteString(fmt.Sprintf("%s %s %s %s re S\n",
						fmtNum(imgX), fmtNum(imgY), fmtNum(imgWidth), fmtNum(imgHeight)))
					contentStream.WriteString("Q\n")

					// Draw image name
					if cell.Image.ImageName != "" && len(cell.Image.ImageName) < 20 {
						contentStream.WriteString("BT\n")
						fontRef := getFontReference(models.Props{FontName: "Helvetica"})
						contentStream.WriteString(fmt.Sprintf("%s 8 Tf\n", fontRef))
						contentStream.WriteString("0.5 0.5 0.5 rg\n")
						textX := imgX + imgWidth/2 - float64(len(cell.Image.ImageName)*2)
						textY := imgY + imgHeight/2
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))
						contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", escapeText(cell.Image.ImageName)))
						contentStream.WriteString("ET\n")
					}
				}
			} else if cell.Text != "" {
				// Draw text with font styling
				contentStream.WriteString("BT\n")
				contentStream.WriteString(getFontReference(cellProps))
				contentStream.WriteString(" ")
				contentStream.WriteString(strconv.Itoa(cellProps.FontSize))
				contentStream.WriteString(" Tf\n")

				// Set text color - always explicitly set to avoid state leakage, default to black
				// Use cell-specific color if available, otherwise use default (title-level) color
				textColor := cell.TextColor
				if textColor == "" {
					textColor = defaultTextColor
				}
				if r, g, b, _, valid := parseHexColor(textColor); valid {
					contentStream.WriteString(fmt.Sprintf("%s %s %s rg\n", fmtNum(r), fmtNum(g), fmtNum(b)))
				} else {
					// Default to black if no valid color specified
					contentStream.WriteString("0 0 0 rg\n")
				}

				// Calculate approximate text width
				textWidth := EstimateTextWidth(cellProps.FontName, cell.Text, float64(cellProps.FontSize))

				var textX float64
				switch cellProps.Alignment {
				case "center":
					textX = cellX + (cellWidth-textWidth)/2
				case "right":
					textX = cellX + cellWidth - textWidth - 5
				default:
					textX = cellX + 5
				}

				textY := pageManager.CurrentYPos - cellHeight/2 - float64(cellProps.FontSize)/2

				contentStream.WriteString("1 0 0 1 0 0 Tm\n")
				contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))

				// Add underline support
				if cellProps.Underline {
					contentStream.WriteString("ET\n")
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					underlineY := textY - 2
					textWidth := float64(len(cell.Text) * cellProps.FontSize / 2)
					contentStream.WriteString(fmt.Sprintf("%s %s m %s %s l S\n",
						fmtNum(textX), fmtNum(underlineY), fmtNum(textX+textWidth), fmtNum(underlineY)))
					contentStream.WriteString("Q\n")
					contentStream.WriteString("BT\n")
					contentStream.WriteString(getFontReference(cellProps))
					contentStream.WriteString(" ")
					contentStream.WriteString(strconv.Itoa(cellProps.FontSize))
					contentStream.WriteString(" Tf\n")
					contentStream.WriteString("1 0 0 1 0 0 Tm\n")
					contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))
				}

				contentStream.WriteString(fmt.Sprintf("%s Tj\n", formatTextForPDF(cellProps, cell.Text)))
				contentStream.WriteString("ET\n")
			}

			// Draw cell borders AFTER content (so they appear on top of images)
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[0], fmtNum(cellX), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellX), fmtNum(pageManager.CurrentYPos)))
				}
				if cellProps.Borders[1] > 0 { // right
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[1], fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos)))
				}
				if cellProps.Borders[2] > 0 { // top
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[2], fmtNum(cellX), fmtNum(pageManager.CurrentYPos), fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos)))
				}
				if cellProps.Borders[3] > 0 { // bottom
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[3], fmtNum(cellX), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos-cellHeight)))
				}
				contentStream.WriteString("Q\n")
			}

			// Add Link Annotation if provided
			if cell.Link != "" {
				// Use captured cellStartX
				linkY := pageManager.CurrentYPos - cellHeight
				pageManager.AddLinkAnnotation(cellX, linkY, cellWidth, cellHeight, cell.Link)
			}
		}

		pageManager.CurrentYPos -= rowHeight
	}
}

// drawTable renders a table with automatic page breaks
func drawTable(table models.Table, tableIdx int, pageManager *PageManager, borderConfig, watermark string, cellImageObjectIDs map[string]int) {
	availableWidth := (pageManager.PageDimensions.Width - 2*margin)
	baseRowHeight := float64(25) // Standard row height

	// Compute column widths in points using weights if provided
	colWidths := make([]float64, table.MaxColumns)
	if len(table.ColumnWidths) == table.MaxColumns {
		// Normalize weights to sum 1
		var sum float64
		for _, w := range table.ColumnWidths {
			if w > 0 {
				sum += w
			}
		}
		if sum <= 0 {
			for i := range colWidths {
				colWidths[i] = availableWidth / float64(table.MaxColumns)
			}
		} else {
			for i, w := range table.ColumnWidths {
				if w <= 0 {
					w = 0
				}
				colWidths[i] = (w / sum) * availableWidth
			}
		}
	} else {
		for i := range colWidths {
			colWidths[i] = availableWidth / float64(table.MaxColumns)
		}
	}

	for rowIdx, row := range table.Rows {
		// Determine this row's height - check if any cell in row has custom height
		rowHeight := baseRowHeight
		if rowIdx < len(table.RowHeights) && table.RowHeights[rowIdx] > 0 {
			rowHeight = baseRowHeight * table.RowHeights[rowIdx]
		}
		// Override with max cell height if any cell specifies it
		for _, cell := range row.Row {
			if cell.Height != nil && *cell.Height > rowHeight {
				rowHeight = *cell.Height
			}
		}

		// Check if row fits on current page
		if pageManager.CheckPageBreak(rowHeight) {
			// Create new page and initialize it
			pageManager.AddNewPage()
			initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions)
		}

		// Get current content stream for this page
		contentStream := pageManager.GetCurrentContentStream()

		// Draw row cells
		currentX := float64(margin)
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			cellProps := parseProps(cell.Props)
			cellX := currentX

			// Use cell-specific width if provided, otherwise use column width
			cellWidth := colWidths[colIdx]
			if cell.Width != nil && *cell.Width > 0 {
				cellWidth = *cell.Width
			}

			// Use cell-specific height if provided, otherwise use row height
			cellHeight := rowHeight
			if cell.Height != nil && *cell.Height > 0 {
				cellHeight = *cell.Height
			}

			// Update X position for next cell
			currentX += cellWidth

			// Draw cell background color FIRST (before any content)
			// Cell-specific bgcolor takes precedence over table-level bgcolor
			bgColor := cell.BgColor
			if bgColor == "" {
				bgColor = table.BgColor
			}
			if r, g, b, _, valid := parseHexColor(bgColor); valid {
				contentStream.WriteString("q\n")
				contentStream.WriteString(fmt.Sprintf("%s %s %s rg\n", fmtNum(r), fmtNum(g), fmtNum(b)))
				contentStream.WriteString(fmt.Sprintf("%s %s %s %s re f\n",
					fmtNum(cellX), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellWidth), fmtNum(cellHeight)))
				contentStream.WriteString("Q\n")
			}

			// Draw content (so borders are drawn on top of images)
			if cell.Image != nil {
				// Check if we have an XObject for this cell image
				cellKey := fmt.Sprintf("%d:%d:%d", tableIdx, rowIdx, colIdx)
				if _, exists := cellImageObjectIDs[cellKey]; exists && cell.Image.ImageData != "" {
					// Render actual image using XObject - fit inside cell with small padding for border
					borderPadding := 1.0 // Small padding to keep image inside borders
					imgWidth := cellWidth - 2*borderPadding
					imgHeight := cellHeight - 2*borderPadding

					// Position at cell's top-left corner with padding
					imgX := cellX + borderPadding
					imgY := pageManager.CurrentYPos - cellHeight + borderPadding

					// Draw actual image using XObject with clipping to prevent overflow - using short names
					shortKey := strings.ReplaceAll(cellKey, ":", "_")
					contentStream.WriteString("q\n")
					// Set up clipping rectangle to confine image within cell bounds (with padding)
					contentStream.WriteString(fmt.Sprintf("%s %s %s %s re W n\n",
						fmtNum(imgX), fmtNum(imgY), fmtNum(imgWidth), fmtNum(imgHeight)))
					contentStream.WriteString(fmt.Sprintf("%s 0 0 %s %s %s cm\n",
						fmtNum(imgWidth), fmtNum(imgHeight), fmtNum(imgX), fmtNum(imgY)))
					contentStream.WriteString(fmt.Sprintf("/C%s Do\n", shortKey))
					contentStream.WriteString("Q\n")
				} else {
					// Fall back to placeholder if no XObject - fit 100% to cell
					imgWidth := cellWidth
					imgHeight := cellHeight

					imgX := cellX
					imgY := pageManager.CurrentYPos - cellHeight

					// Draw placeholder border using 're' operator
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					contentStream.WriteString("0.7 0.7 0.7 RG\n")
					contentStream.WriteString(fmt.Sprintf("%s %s %s %s re S\n",
						fmtNum(imgX), fmtNum(imgY), fmtNum(imgWidth), fmtNum(imgHeight)))
					contentStream.WriteString("Q\n")

					// Draw image name
					if cell.Image.ImageName != "" && len(cell.Image.ImageName) < 20 {
						contentStream.WriteString("BT\n")
						fontRef := getFontReference(models.Props{FontName: "Helvetica"})
						contentStream.WriteString(fmt.Sprintf("%s 8 Tf\n", fontRef))
						contentStream.WriteString("0.5 0.5 0.5 rg\n")
						textX := imgX + imgWidth/2 - float64(len(cell.Image.ImageName)*2)
						textY := imgY + imgHeight/2
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))
						contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", escapeText(cell.Image.ImageName)))
						contentStream.WriteString("ET\n")
					}
				}
			} else if cell.FormField != nil {
				// Draw form field widget
				fieldWidth := 12.0
				fieldHeight := 12.0

				if cell.FormField.Type == "text" {
					fieldWidth = cellWidth - 4
					fieldHeight = cellHeight - 4
				}

				fieldX := cellX + (cellWidth-fieldWidth)/2
				fieldY := pageManager.CurrentYPos - (cellHeight+fieldHeight)/2

				drawWidget(cell, fieldX, fieldY, fieldWidth, fieldHeight, pageManager)

			} else if cell.Checkbox != nil {
				// Draw checkbox using 're' operator
				checkboxSize := 10.0
				checkboxX := cellX + (cellWidth-checkboxSize)/2
				checkboxY := pageManager.CurrentYPos - (cellHeight+checkboxSize)/2

				contentStream.WriteString("q\n")
				contentStream.WriteString("1 w\n")
				contentStream.WriteString(fmt.Sprintf("%s %s %s %s re S\n",
					fmtNum(checkboxX), fmtNum(checkboxY), fmtNum(checkboxSize), fmtNum(checkboxSize)))

				if *cell.Checkbox {
					contentStream.WriteString(fmt.Sprintf("%s %s m %s %s l %s %s m %s %s l S\n",
						fmtNum(checkboxX+2), fmtNum(checkboxY+2), fmtNum(checkboxX+checkboxSize-2), fmtNum(checkboxY+checkboxSize-2),
						fmtNum(checkboxX+checkboxSize-2), fmtNum(checkboxY+2), fmtNum(checkboxX+2), fmtNum(checkboxY+checkboxSize-2)))
				}
				contentStream.WriteString("Q\n")
			} else if cell.Text != "" {
				// Draw text with font styling
				contentStream.WriteString("BT\n")
				contentStream.WriteString(getFontReference(cellProps))
				contentStream.WriteString(" ")
				contentStream.WriteString(strconv.Itoa(cellProps.FontSize))
				contentStream.WriteString(" Tf\n")

				// Set text color (cell-level takes precedence over table-level, default to black)
				// Always explicitly set the color to avoid state leakage from previous tables
				textColor := cell.TextColor
				if textColor == "" {
					textColor = table.TextColor
				}
				if r, g, b, _, valid := parseHexColor(textColor); valid {
					contentStream.WriteString(fmt.Sprintf("%s %s %s rg\n", fmtNum(r), fmtNum(g), fmtNum(b)))
				} else {
					// Default to black if no valid color specified
					contentStream.WriteString("0 0 0 rg\n")
				}

				// Calculate approximate text width
				textWidth := EstimateTextWidth(cellProps.FontName, cell.Text, float64(cellProps.FontSize))

				var textX float64
				switch cellProps.Alignment {
				case "center":
					// Center the text within the cell
					textX = cellX + (cellWidth-textWidth)/2
				case "right":
					// Right align: position text so it ends near the right edge of cell
					textX = cellX + cellWidth - textWidth - 5
				default:
					textX = cellX + 5
				}

				textY := pageManager.CurrentYPos - cellHeight/2 - float64(cellProps.FontSize)/2

				// Reset text matrix and position absolutely
				contentStream.WriteString("1 0 0 1 0 0 Tm\n")
				contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))

				// Add underline support
				if cellProps.Underline {
					// End text object before drawing underline
					contentStream.WriteString("ET\n")
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					underlineY := textY - 2
					textWidth := float64(len(cell.Text) * cellProps.FontSize / 2)
					contentStream.WriteString(fmt.Sprintf("%s %s m %s %s l S\n",
						fmtNum(textX), fmtNum(underlineY), fmtNum(textX+textWidth), fmtNum(underlineY)))
					contentStream.WriteString("Q\n")
					// Start text object again
					contentStream.WriteString("BT\n")
					contentStream.WriteString(getFontReference(cellProps))
					contentStream.WriteString(" ")
					contentStream.WriteString(strconv.Itoa(cellProps.FontSize))
					contentStream.WriteString(" Tf\n")
					contentStream.WriteString("1 0 0 1 0 0 Tm\n")
					contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))
				}

				contentStream.WriteString(fmt.Sprintf("%s Tj\n", formatTextForPDF(cellProps, cell.Text)))
				contentStream.WriteString("ET\n")
			}

			// Draw cell borders AFTER content (so they appear on top of images)
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[0], fmtNum(cellX), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellX), fmtNum(pageManager.CurrentYPos)))
				}
				if cellProps.Borders[1] > 0 { // right
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[1], fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos)))
				}
				if cellProps.Borders[2] > 0 { // top
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[2], fmtNum(cellX), fmtNum(pageManager.CurrentYPos), fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos)))
				}
				if cellProps.Borders[3] > 0 { // bottom
					contentStream.WriteString(fmt.Sprintf("%d w %s %s m %s %s l S\n",
						cellProps.Borders[3], fmtNum(cellX), fmtNum(pageManager.CurrentYPos-cellHeight), fmtNum(cellX+cellWidth), fmtNum(pageManager.CurrentYPos-cellHeight)))
				}
				contentStream.WriteString("Q\n")
			}

			// Create link annotation if cell has a link
			if cell.Link != "" {
				DrawCellLink(cell.Link, cellX, pageManager.CurrentYPos-cellHeight, cellWidth, cellHeight, pageManager)
			}
		}

		pageManager.CurrentYPos -= rowHeight
	}
}

// drawSpacer adds vertical space in the document
func drawSpacer(spacer models.Spacer, pageManager *PageManager) {
	height := spacer.Height
	if height <= 0 {
		height = 20 // Default spacer height
	}
	pageManager.CurrentYPos -= height
}

// drawFooter renders the document footer
func drawFooter(contentStream *bytes.Buffer, footer models.Footer, pageManager *PageManager) {
	footerProps := parseProps(footer.Font)
	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(footerProps))
	contentStream.WriteString(" ")
	contentStream.WriteString(strconv.Itoa(footerProps.FontSize))
	contentStream.WriteString(" Tf\n")

	// Position footer outside the page border on the left side
	footerX := 20 // 20pt from left edge (outside margin)
	footerY := 20 // 20pt from bottom edge (outside margin)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%d %d Td\n", footerX, footerY))
	contentStream.WriteString(fmt.Sprintf("%s Tj\n", formatTextForPDF(footerProps, footer.Text)))
	contentStream.WriteString("ET\n")

	// Add Link Annotation if provided
	if footer.Link != "" {
		// Calculate approximate text width
		// Using standard estimation for width since we don't have exact calc here easily without refactoring
		// But footer is likely simple text.
		textWidth := EstimateTextWidth(footerProps.FontName, footer.Text, float64(footerProps.FontSize))

		rectX := float64(footerX)
		rectY := float64(footerY) - float64(footerProps.FontSize)*0.2
		rectW := textWidth
		rectH := float64(footerProps.FontSize) * 1.2
		pageManager.AddLinkAnnotation(rectX, rectY, rectW, rectH, footer.Link)
	}
}

// drawPageNumber renders page number in bottom right corner
func drawPageNumber(contentStream *bytes.Buffer, currentPage, totalPages int, pageDims PageDimensions) {
	pageText := fmt.Sprintf("Page %d of %d", currentPage, totalPages)

	// Track characters for font subsetting
	registry := GetFontRegistry()
	if registry.HasFont("Helvetica") {
		registry.MarkCharsUsed("Helvetica", pageText)
	}

	// Use props for proper font encoding
	pageProps := models.Props{FontName: "Helvetica", FontSize: 10}

	contentStream.WriteString("BT\n")
	fontRef := getFontReference(pageProps)
	contentStream.WriteString(fmt.Sprintf("%s 10 Tf\n", fontRef)) // Use Helvetica, 10pt

	// Calculate text width for proper right alignment
	textWidth := float64(len(pageText)) * 6 // Approximate character width for 10pt font

	// Position outside the page border on the right side
	pageNumberX := pageDims.Width - textWidth - 20 // 20pt from right edge (outside margin)
	pageNumberY := 20                              // 20pt from bottom edge (outside margin)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%s %d Td\n", fmtNum(pageNumberX), pageNumberY))
	contentStream.WriteString(fmt.Sprintf("%s Tj\n", formatTextForPDF(pageProps, pageText)))
	contentStream.WriteString("ET\n")
}

// drawImage renders an image in the PDF with automatic page breaks
func drawImage(image models.Image, pageManager *PageManager, borderConfig, watermark string) {
	// Skip if no image data
	if image.ImageData == "" {
		return
	}

	imageHeight := image.Height
	if imageHeight == 0 {
		imageHeight = 200 // Default height
	}

	// Add some spacing before image
	spacing := float64(20)

	// Check if image fits on current page
	if pageManager.CheckPageBreak(imageHeight + spacing) {
		// Create new page and initialize it
		pageManager.AddNewPage()
		initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions)
	}

	// Get current content stream for this page
	contentStream := pageManager.GetCurrentContentStream()

	// For now, we'll draw a placeholder rectangle for the image
	// Full PDF image embedding would require creating XObject image streams
	// which is complex. This is a simplified version that shows where the image would go.

	imageWidth := image.Width
	if imageWidth == 0 {
		imageWidth = 300 // Default width
	}

	// Center the image horizontally
	imageX := (pageManager.PageDimensions.Width - imageWidth) / 2
	imageY := pageManager.CurrentYPos - imageHeight

	// Draw a border around the image area using 're' operator
	contentStream.WriteString("q\n")
	contentStream.WriteString("0.5 w\n")
	contentStream.WriteString("0.8 0.8 0.8 RG\n") // Light gray border
	contentStream.WriteString(fmt.Sprintf("%s %s %s %s re S\n",
		fmtNum(imageX), fmtNum(imageY), fmtNum(imageWidth), fmtNum(imageHeight)))
	contentStream.WriteString("Q\n")

	// Add image name text in the center
	if image.ImageName != "" {
		contentStream.WriteString("BT\n")
		fontRef := getFontReference(models.Props{FontName: "Helvetica"})
		contentStream.WriteString(fmt.Sprintf("%s 10 Tf\n", fontRef))
		contentStream.WriteString("0.6 0.6 0.6 rg\n") // Gray text

		// Center the text
		textX := imageX + imageWidth/2
		textY := imageY + imageHeight/2

		contentStream.WriteString("1 0 0 1 0 0 Tm\n")
		contentStream.WriteString(fmt.Sprintf("%s %s Td\n", fmtNum(textX), fmtNum(textY)))
		contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", escapeText(image.ImageName)))
		contentStream.WriteString("ET\n")
	}

	// Add Link Annotation if provided
	if image.Link != "" {
		pageManager.AddLinkAnnotation(imageX, imageY, imageWidth, imageHeight, image.Link)
	}

	pageManager.CurrentYPos -= (imageHeight + spacing)
}

// drawImageWithXObjectInternal handles image drawing with XObject, including page breaks
func drawImageWithXObjectInternal(image models.Image, imageXObjectRef string, pageManager *PageManager, borderConfig, watermark string, originalImgWidth, originalImgHeight int) {
	// Calculate usable width to estimate height for page break check
	usableWidth := pageManager.PageDimensions.Width - 2*margin

	// Calculate height based on aspect ratio
	var imageHeight float64
	if originalImgWidth > 0 && originalImgHeight > 0 {
		aspectRatio := float64(originalImgHeight) / float64(originalImgWidth)
		imageHeight = usableWidth * aspectRatio
	} else if image.Height > 0 && image.Width > 0 {
		aspectRatio := image.Height / image.Width
		imageHeight = usableWidth * aspectRatio
	} else {
		imageHeight = 200 // Default height
	}

	// Check if image fits on current page (no extra spacing)
	if pageManager.CheckPageBreak(imageHeight) {
		// Create new page and initialize it
		pageManager.AddNewPage()
		initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions)
	}

	// Get current content stream for this page
	contentStream := pageManager.GetCurrentContentStream()

	// Draw the image using XObject
	drawImageWithXObject(contentStream, image, imageXObjectRef, pageManager, originalImgWidth, originalImgHeight)
}

// drawWidget creates a widget annotation for a form field
func drawWidget(cell models.Cell, x, y, w, h float64, pageManager *PageManager) {
	if cell.FormField == nil {
		return
	}

	field := cell.FormField
	// Calculate rect with optimized precision
	rect := fmt.Sprintf("[%s %s %s %s]", fmtNum(x), fmtNum(y), fmtNum(x+w), fmtNum(y+h))

	var widgetDict strings.Builder
	widgetDict.WriteString("<< /Type /Annot /Subtype /Widget")
	widgetDict.WriteString(fmt.Sprintf(" /Rect %s", rect))
	widgetDict.WriteString(fmt.Sprintf(" /T (%s)", escapeText(field.Name)))
	widgetDict.WriteString(" /F 4") // Print flag

	switch field.Type {
	case "checkbox":
		widgetDict.WriteString(" /FT /Btn")

		onState := "/Yes"
		offState := "/Off"

		val := offState
		if field.Checked {
			val = onState
		}

		widgetDict.WriteString(fmt.Sprintf(" /V %s /AS %s", val, val))

		// Checkbox Appearance Streams using 're' operator
		// On Appearance (Box with X)
		onAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %s %s re S 2 2 m %s %s l 2 %s m %s 2 l S Q", fmtNum(w), fmtNum(h), fmtNum(w-2), fmtNum(h-2), fmtNum(h-2), fmtNum(w-2))
		onAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /ProcSet [/PDF] >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), len(onAP), onAP))

		// Off Appearance (Empty Box)
		offAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %s %s re S Q", fmtNum(w), fmtNum(h))
		offAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /ProcSet [/PDF] >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), len(offAP), offAP))

		widgetDict.WriteString(fmt.Sprintf(" /AP << /N << /Yes %d 0 R /Off %d 0 R >> >>", onAPID, offAPID))

	case "radio":
		widgetDict.WriteString(" /FT /Btn /Ff 49152") // Radio button flag

		onState := "/" + field.Value
		offState := "/Off"

		val := offState
		if field.Checked {
			val = onState
		}

		widgetDict.WriteString(fmt.Sprintf(" /V %s /AS %s", val, val))

		if field.Shape == "square" {
			// Radio Appearance Streams (Square with dot) using 're' operator
			onAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %s %s re S 3 3 %s %s re f Q", fmtNum(w), fmtNum(h), fmtNum(w-6), fmtNum(h-6))
			onAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /ProcSet [/PDF] >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), len(onAP), onAP))

			offAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %s %s re S Q", fmtNum(w), fmtNum(h))
			offAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /ProcSet [/PDF] >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), len(offAP), offAP))

			widgetDict.WriteString(fmt.Sprintf(" /AP << /N << /%s %d 0 R /Off %d 0 R >> >>", field.Value, onAPID, offAPID))
		} else {
			// Default to Round (Circle)
			// Add /MK dictionary with appearance characteristics for circle radio button
			widgetDict.WriteString(" /MK << /BC [0 0 0] /BG [0.9 0.9 0.9] /CA (l) >>")

			// Center point and radius calculations
			cx := w / 2
			cy := h / 2
			outerR := cx - 0.5      // Outer circle radius
			innerR := outerR * 0.45 // Inner dot radius

			// Bézier curve control point factor
			k := 0.5523 // approximation of 4*(sqrt(2)-1)/3 for circle

			// Build outer circle path using Bézier curves with reduced precision
			outerCirclePath := fmt.Sprintf("%s 0 m %s %s %s %s 0 %s c %s %s %s %s %s 0 c %s %s %s %s 0 %s c %s %s %s %s %s 0 c h",
				fmtNum(outerR),
				fmtNum(outerR), fmtNum(outerR*k), fmtNum(outerR*k), fmtNum(outerR), fmtNum(outerR),
				fmtNum(-outerR*k), fmtNum(outerR), fmtNum(-outerR), fmtNum(outerR*k), fmtNum(-outerR),
				fmtNum(-outerR), fmtNum(-outerR*k), fmtNum(-outerR*k), fmtNum(-outerR), fmtNum(-outerR),
				fmtNum(outerR*k), fmtNum(-outerR), fmtNum(outerR), fmtNum(-outerR*k), fmtNum(outerR))

			// Build inner dot circle path
			innerCirclePath := fmt.Sprintf("%s 0 m %s %s %s %s 0 %s c %s %s %s %s %s 0 c %s %s %s %s 0 %s c %s %s %s %s %s 0 c h",
				fmtNum(innerR),
				fmtNum(innerR), fmtNum(innerR*k), fmtNum(innerR*k), fmtNum(innerR), fmtNum(innerR),
				fmtNum(-innerR*k), fmtNum(innerR), fmtNum(-innerR), fmtNum(innerR*k), fmtNum(-innerR),
				fmtNum(-innerR), fmtNum(-innerR*k), fmtNum(-innerR*k), fmtNum(-innerR), fmtNum(-innerR),
				fmtNum(innerR*k), fmtNum(-innerR), fmtNum(innerR), fmtNum(-innerR*k), fmtNum(innerR))

			// ON appearance: Light background fill + dark stroke + dark inner dot
			onAP := fmt.Sprintf("q\n0.9 0.9 0.9 rg 0 0 0 RG 1 w\n1 0 0 1 %s %s cm\n%s\nB\nQ\nq\n0 0 0 rg\n1 0 0 1 %s %s cm\n%s\nf\nQ",
				fmtNum(cx), fmtNum(cy), outerCirclePath,
				fmtNum(cx), fmtNum(cy), innerCirclePath)
			onAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /ProcSet [/PDF] >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), len(onAP), onAP))

			// OFF appearance: Light background fill + dark stroke (no inner dot)
			offAP := fmt.Sprintf("q\n0.9 0.9 0.9 rg 0 0 0 RG 1 w\n1 0 0 1 %s %s cm\n%s\nB\nQ",
				fmtNum(cx), fmtNum(cy), outerCirclePath)
			offAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /ProcSet [/PDF] >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), len(offAP), offAP))

			widgetDict.WriteString(fmt.Sprintf(" /AP << /N << /%s %d 0 R /Off %d 0 R >> >>", field.Value, onAPID, offAPID))
		}
	case "text":
		widgetDict.WriteString(" /FT /Tx") // Text field
		widgetDict.WriteString(fmt.Sprintf(" /V (%s)", escapeText(field.Value)))

		// Calculate font size based on field height
		fontSize := 10.0
		if h < 14 {
			fontSize = h - 4
		}
		if fontSize < 6 {
			fontSize = 6
		}

		// Mark field value for font subsetting (critical for PDF/A compliance)
		// Form field text is rendered in appearance streams using custom fonts
		if field.Value != "" {
			registry := GetFontRegistry()
			if registry.HasFont("Helvetica") {
				// PDF/A mode: Liberation font registered as Helvetica
				registry.MarkCharsUsed("Helvetica", field.Value)
			}
		}

		// Get the appropriate font reference for widgets (handles PDF/A mode)
		widgetFontRef := getWidgetFontReference()

		// Default Appearance string - used by viewer to render text
		// Use proper font reference instead of hardcoded /Helv
		widgetDict.WriteString(fmt.Sprintf(" /DA (%s %s Tf 0 g)", widgetFontRef, fmtNum(fontSize)))

		// Build appearance stream: border + text properly structured
		// Use /Tx BMC ... EMC to mark text content area (viewer replaces this when editing)
		var apStream strings.Builder
		// Draw border first
		apStream.WriteString(fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %s %s re S Q ", fmtNum(w), fmtNum(h)))
		// Text content marked with /Tx BMC ... EMC (marked content)
		// This tells PDF viewer this is the text area it should manage
		apStream.WriteString("/Tx BMC ")
		if field.Value != "" {
			textY := (h - fontSize) / 2
			textX := 2.0
			// Use proper encoding for field value
			fieldProps := models.Props{FontName: "Helvetica", FontSize: int(fontSize)}
			encodedValue := formatTextForPDF(fieldProps, field.Value)
			// Use proper font reference in appearance stream
			apStream.WriteString(fmt.Sprintf("q BT %s %s Tf 0 g %s %s Td %s Tj ET Q ", widgetFontRef, fmtNum(fontSize), fmtNum(textX), fmtNum(textY), encodedValue))
		}
		apStream.WriteString("EMC")
		apContent := apStream.String()

		// Create appearance XObject
		// IMPORTANT: Form XObjects must declare all resources they use in their own Resources dictionary
		// This is required for PDF/A-4 compliance - resources cannot be inherited from page level
		var apObjContent string
		if getWidgetFontName() == "" {
			// PDF/A mode: Get the font object ID from the font registry
			// The widgetFontRef (e.g., /CF2000) references a custom font that must be in Resources
			fontObjID := getWidgetFontObjectID()
			if fontObjID > 0 {
				// Include the font reference in the XObject's Resources dictionary
				apObjContent = fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /Font << %s %d 0 R >> >> /Length %d >> stream\n%s\nendstream",
					fmtNum(w), fmtNum(h), widgetFontRef, fontObjID, len(apContent), apContent)
			} else {
				// Fallback: empty resources (should not happen in PDF/A mode)
				apObjContent = fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << >> /Length %d >> stream\n%s\nendstream",
					fmtNum(w), fmtNum(h), len(apContent), apContent)
			}
		} else {
			// Standard mode: Embed Helvetica definition in XObject resources
			var helveticaFont string
			if pageManager.ArlingtonCompatible {
				helveticaFont = GetHelveticaFontResourceString()
			} else {
				helveticaFont = GetSimpleHelveticaFontResourceString()
			}
			apObjContent = fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources << /Font << /F1 %s >> >> /Length %d >> stream\n%s\nendstream", fmtNum(w), fmtNum(h), helveticaFont, len(apContent), apContent)
		}
		apID := pageManager.AddExtraObject(apObjContent)

		widgetDict.WriteString(fmt.Sprintf(" /AP << /N %d 0 R >>", apID))
	}

	widgetDict.WriteString(" >>")

	objID := pageManager.AddExtraObject(widgetDict.String())
	pageManager.AddAnnotation(objID)
}
