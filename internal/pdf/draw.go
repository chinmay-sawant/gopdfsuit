package pdf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

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

// drawTitle renders the document title (either simple text or embedded table)
func drawTitle(contentStream *bytes.Buffer, title models.Title, titleProps models.Props, pageManager *PageManager, cellImageObjectIDs map[string]int) {
	// Check if title has an embedded table
	if title.Table != nil && len(title.Table.Rows) > 0 {
		drawTitleTable(contentStream, title.Table, pageManager, cellImageObjectIDs)
		return
	}

	// Simple text title
	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(titleProps))
	contentStream.WriteString(" ")
	contentStream.WriteString(strconv.Itoa(titleProps.FontSize))
	contentStream.WriteString(" Tf\n")

	// Calculate approximate text width (using average character width ratio of 0.5 for Helvetica)
	textWidth := float64(len(title.Text)) * float64(titleProps.FontSize) * 0.5

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

	pageManager.CurrentYPos -= float64(titleProps.FontSize + 20)
	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", titleX, pageManager.CurrentYPos))
	contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", title.Text))
	contentStream.WriteString("ET\n")

	pageManager.CurrentYPos -= 30
}

// drawTitleTable renders an embedded table within the title section (no borders by default)
func drawTitleTable(contentStream *bytes.Buffer, table *models.TitleTable, pageManager *PageManager, cellImageObjectIDs map[string]int) {
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

			// Draw cell borders (title table cells have borders if specified in props)
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[0]), cellX, pageManager.CurrentYPos-cellHeight, cellX, pageManager.CurrentYPos))
				}
				if cellProps.Borders[1] > 0 { // right
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[1]), cellX+cellWidth, pageManager.CurrentYPos-cellHeight, cellX+cellWidth, pageManager.CurrentYPos))
				}
				if cellProps.Borders[2] > 0 { // top
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[2]), cellX, pageManager.CurrentYPos, cellX+cellWidth, pageManager.CurrentYPos))
				}
				if cellProps.Borders[3] > 0 { // bottom
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[3]), cellX, pageManager.CurrentYPos-cellHeight, cellX+cellWidth, pageManager.CurrentYPos-cellHeight))
				}
				contentStream.WriteString("Q\n")
			}

			// Draw image or text
			if cell.Image != nil && cell.Image.ImageData != "" {
				// Check if we have an XObject for this title cell image
				cellKey := fmt.Sprintf("title:%d:%d", rowIdx, colIdx)
				if _, exists := cellImageObjectIDs[cellKey]; exists {
					// Render actual image using XObject - fit 100% to cell
					imgWidth := cellWidth
					imgHeight := cellHeight

					imgX := cellX
					imgY := pageManager.CurrentYPos - cellHeight

					// Draw actual image using XObject
					contentStream.WriteString("q\n")
					contentStream.WriteString(fmt.Sprintf("%.2f 0 0 %.2f %.2f %.2f cm\n",
						imgWidth, imgHeight, imgX, imgY))
					contentStream.WriteString(fmt.Sprintf("/CellImg_%s Do\n", cellKey))
					contentStream.WriteString("Q\n")
				} else {
					// Fall back to placeholder
					imgWidth := cellWidth
					imgHeight := cellHeight
					imgX := cellX
					imgY := pageManager.CurrentYPos - cellHeight

					// Draw placeholder border
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					contentStream.WriteString("0.7 0.7 0.7 RG\n")
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re S\n",
						imgX, imgY, imgWidth, imgHeight))
					contentStream.WriteString("Q\n")

					// Draw image name
					if cell.Image.ImageName != "" && len(cell.Image.ImageName) < 20 {
						contentStream.WriteString("BT\n")
						contentStream.WriteString("/F1 8 Tf\n")
						contentStream.WriteString("0.5 0.5 0.5 rg\n")
						textX := imgX + imgWidth/2 - float64(len(cell.Image.ImageName)*2)
						textY := imgY + imgHeight/2
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))
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

				// Calculate approximate text width
				textWidth := float64(len(cell.Text)) * float64(cellProps.FontSize) * 0.5

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
				contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))

				// Add underline support
				if cellProps.Underline {
					contentStream.WriteString("ET\n")
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					underlineY := textY - 2
					textWidth := float64(len(cell.Text) * cellProps.FontSize / 2)
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
						textX, underlineY, textX+textWidth, underlineY))
					contentStream.WriteString("Q\n")
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

	// Add some spacing after the title table
	pageManager.CurrentYPos -= 10
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

			// Draw cell borders
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[0]), cellX, pageManager.CurrentYPos-cellHeight, cellX, pageManager.CurrentYPos))
				}
				if cellProps.Borders[1] > 0 { // right
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[1]), cellX+cellWidth, pageManager.CurrentYPos-cellHeight, cellX+cellWidth, pageManager.CurrentYPos))
				}
				if cellProps.Borders[2] > 0 { // top
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[2]), cellX, pageManager.CurrentYPos, cellX+cellWidth, pageManager.CurrentYPos))
				}
				if cellProps.Borders[3] > 0 { // bottom
					contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
						float64(cellProps.Borders[3]), cellX, pageManager.CurrentYPos-cellHeight, cellX+cellWidth, pageManager.CurrentYPos-cellHeight))
				}
				contentStream.WriteString("Q\n")
			}

			// Draw text, checkbox, or image
			if cell.Image != nil {
				// Check if we have an XObject for this cell image
				cellKey := fmt.Sprintf("%d:%d:%d", tableIdx, rowIdx, colIdx)
				if _, exists := cellImageObjectIDs[cellKey]; exists && cell.Image.ImageData != "" {
					// Render actual image using XObject - fit 100% to cell
					// Use full cell dimensions (no padding)
					imgWidth := cellWidth
					imgHeight := cellHeight

					// Position at cell's top-left corner
					imgX := cellX
					imgY := pageManager.CurrentYPos - cellHeight

					// Draw actual image using XObject
					contentStream.WriteString("q\n")
					contentStream.WriteString(fmt.Sprintf("%.2f 0 0 %.2f %.2f %.2f cm\n",
						imgWidth, imgHeight, imgX, imgY))
					contentStream.WriteString(fmt.Sprintf("/CellImg_%s Do\n", cellKey))
					contentStream.WriteString("Q\n")
				} else {
					// Fall back to placeholder if no XObject - fit 100% to cell
					imgWidth := cellWidth
					imgHeight := cellHeight

					imgX := cellX
					imgY := pageManager.CurrentYPos - cellHeight

					// Draw placeholder border
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					contentStream.WriteString("0.7 0.7 0.7 RG\n")
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re S\n",
						imgX, imgY, imgWidth, imgHeight))
					contentStream.WriteString("Q\n")

					// Draw image name
					if cell.Image.ImageName != "" && len(cell.Image.ImageName) < 20 {
						contentStream.WriteString("BT\n")
						contentStream.WriteString("/F1 8 Tf\n")
						contentStream.WriteString("0.5 0.5 0.5 rg\n")
						textX := imgX + imgWidth/2 - float64(len(cell.Image.ImageName)*2)
						textY := imgY + imgHeight/2
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))
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
				// Draw checkbox
				checkboxSize := 10.0
				checkboxX := cellX + (cellWidth-checkboxSize)/2
				checkboxY := pageManager.CurrentYPos - (cellHeight+checkboxSize)/2

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

				// Calculate approximate text width (using average character width ratio of 0.5 for Helvetica)
				textWidth := float64(len(cell.Text)) * float64(cellProps.FontSize) * 0.5

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
func drawFooter(contentStream *bytes.Buffer, footer models.Footer) {
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

	// Draw a border around the image area
	contentStream.WriteString("q\n")
	contentStream.WriteString("0.5 w\n")
	contentStream.WriteString("0.8 0.8 0.8 RG\n") // Light gray border
	contentStream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re S\n",
		imageX, imageY, imageWidth, imageHeight))
	contentStream.WriteString("Q\n")

	// Add image name text in the center
	if image.ImageName != "" {
		contentStream.WriteString("BT\n")
		contentStream.WriteString("/F1 10 Tf\n")
		contentStream.WriteString("0.6 0.6 0.6 rg\n") // Gray text

		// Center the text
		textX := imageX + imageWidth/2
		textY := imageY + imageHeight/2

		contentStream.WriteString("1 0 0 1 0 0 Tm\n")
		contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))
		contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", escapeText(image.ImageName)))
		contentStream.WriteString("ET\n")
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
	// Calculate rect
	rect := fmt.Sprintf("[%.2f %.2f %.2f %.2f]", x, y, x+w, y+h)

	var widgetDict strings.Builder
	widgetDict.WriteString("<< /Type /Annot /Subtype /Widget")
	widgetDict.WriteString(fmt.Sprintf(" /Rect %s", rect))
	widgetDict.WriteString(fmt.Sprintf(" /T (%s)", escapeText(field.Name)))
	widgetDict.WriteString(" /F 4") // Print flag

	if field.Type == "checkbox" {
		widgetDict.WriteString(" /FT /Btn")

		onState := "/Yes"
		offState := "/Off"

		val := offState
		if field.Checked {
			val = onState
		}

		widgetDict.WriteString(fmt.Sprintf(" /V %s /AS %s", val, val))

		// Checkbox Appearance Streams
		// On Appearance (Box with X)
		onAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %.2f %.2f re S 2 2 m %.2f %.2f l 2 %.2f m %.2f 2 l S Q", w, h, w-2, h-2, h-2, w-2)
		onAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(onAP), onAP))

		// Off Appearance (Empty Box)
		offAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %.2f %.2f re S Q", w, h)
		offAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(offAP), offAP))

		widgetDict.WriteString(fmt.Sprintf(" /AP << /N << /Yes %d 0 R /Off %d 0 R >> >>", onAPID, offAPID))

	} else if field.Type == "radio" {
		widgetDict.WriteString(" /FT /Btn /Ff 49152") // Radio button flag

		onState := "/" + field.Value
		offState := "/Off"

		val := offState
		if field.Checked {
			val = onState
		}

		widgetDict.WriteString(fmt.Sprintf(" /V %s /AS %s", val, val))

		if field.Shape == "square" {
			// Radio Appearance Streams (Square with dot)
			onAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %.2f %.2f re S 3 3 %.2f %.2f re f Q", w, h, w-6, h-6)
			onAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(onAP), onAP))

			offAP := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %.2f %.2f re S Q", w, h)
			offAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(offAP), offAP))

			widgetDict.WriteString(fmt.Sprintf(" /AP << /N << /%s %d 0 R /Off %d 0 R >> >>", field.Value, onAPID, offAPID))
		} else {
			// Default to Round (Circle)
			// Add /MK dictionary with appearance characteristics for circle radio button
			// /BC = border color (black), /BG = background color (light gray), /CA = caption (l = bullet in ZapfDingbats)
			widgetDict.WriteString(" /MK << /BC [0 0 0] /BG [0.9 0.9 0.9] /CA (l) >>")

			// Center point and radius calculations
			cx := w / 2
			cy := h / 2
			outerR := cx - 0.5      // Outer circle radius (slightly smaller than half width)
			innerR := outerR * 0.45 // Inner dot radius (about 45% of outer)

			// Bézier curve control point factor
			k := 0.5523 // approximation of 4*(sqrt(2)-1)/3 for circle

			// Build outer circle path using Bézier curves (centered at origin, will use cm to translate)
			// This draws a circle from the right side, going counter-clockwise
			// Each Bézier curve segment: x1 y1 x2 y2 x3 y3 c
			outerCirclePath := fmt.Sprintf("%.4f 0 m %.4f %.4f %.4f %.4f 0 %.4f c %.4f %.4f %.4f %.4f %.4f 0 c %.4f %.4f %.4f %.4f 0 %.4f c %.4f %.4f %.4f %.4f %.4f 0 c h",
				outerR,
				outerR, outerR*k, outerR*k, outerR, outerR,
				-outerR*k, outerR, -outerR, outerR*k, -outerR,
				-outerR, -outerR*k, -outerR*k, -outerR, -outerR,
				outerR*k, -outerR, outerR, -outerR*k, outerR)

			// Build inner dot circle path
			innerCirclePath := fmt.Sprintf("%.4f 0 m %.4f %.4f %.4f %.4f 0 %.4f c %.4f %.4f %.4f %.4f %.4f 0 c %.4f %.4f %.4f %.4f 0 %.4f c %.4f %.4f %.4f %.4f %.4f 0 c h",
				innerR,
				innerR, innerR*k, innerR*k, innerR, innerR,
				-innerR*k, innerR, -innerR, innerR*k, -innerR,
				-innerR, -innerR*k, -innerR*k, -innerR, -innerR,
				innerR*k, -innerR, innerR, -innerR*k, innerR)

			// ON appearance: Light background fill + dark stroke + dark inner dot
			// Using 'B' (fill and stroke) for combined operation
			onAP := fmt.Sprintf("q\n0.9 0.9 0.9 rg 0 0 0 RG 1 w\n1 0 0 1 %.2f %.2f cm\n%s\nB\nQ\nq\n0 0 0 rg\n1 0 0 1 %.2f %.2f cm\n%s\nf\nQ",
				cx, cy, outerCirclePath,
				cx, cy, innerCirclePath)
			onAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(onAP), onAP))

			// OFF appearance: Light background fill + dark stroke (no inner dot)
			offAP := fmt.Sprintf("q\n0.9 0.9 0.9 rg 0 0 0 RG 1 w\n1 0 0 1 %.2f %.2f cm\n%s\nB\nQ",
				cx, cy, outerCirclePath)
			offAPID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(offAP), offAP))

			widgetDict.WriteString(fmt.Sprintf(" /AP << /N << /%s %d 0 R /Off %d 0 R >> >>", field.Value, onAPID, offAPID))
		}
	} else if field.Type == "text" {
		widgetDict.WriteString(" /FT /Tx") // Text field
		widgetDict.WriteString(fmt.Sprintf(" /V (%s)", escapeText(field.Value)))

		// Appearance Stream for Text Field (Simple Box)
		apBox := fmt.Sprintf("q 1 w 0 0 0 RG 0 0 %.2f %.2f re S Q", w, h)
		apID := pageManager.AddExtraObject(fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %.2f %.2f] /Resources << >> /Length %d >> stream\n%s\nendstream", w, h, len(apBox), apBox))

		widgetDict.WriteString(fmt.Sprintf(" /AP << /N %d 0 R >>", apID))
	}

	widgetDict.WriteString(" >>")

	objID := pageManager.AddExtraObject(widgetDict.String())
	pageManager.AddAnnotation(objID)
}
