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
func drawTable(table models.Table, tableIdx int, pageManager *PageManager, borderConfig, watermark string, cellImageObjectIDs map[string]int) {
	cellWidth := (pageManager.PageDimensions.Width - 2*margin) / float64(table.MaxColumns)
	rowHeight := float64(25) // Standard row height

	for rowIdx, row := range table.Rows {
		// Check if row fits on current page
		if pageManager.CheckPageBreak(rowHeight) {
			// Create new page and initialize it
			pageManager.AddNewPage()
			initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions)
		}

		// Get current content stream for this page
		contentStream := pageManager.GetCurrentContentStream()

		// Draw row cells
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			cellProps := parseProps(cell.Props)
			cellX := float64(margin) + float64(colIdx)*cellWidth

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

			// Draw text, checkbox, or image
			if cell.Image != nil {
				// Check if we have an XObject for this cell image
				cellKey := fmt.Sprintf("%d:%d:%d", tableIdx, rowIdx, colIdx)
				if _, exists := cellImageObjectIDs[cellKey]; exists && cell.Image.ImageData != "" {
					// Render actual image using XObject
					imgWidth := cell.Image.Width
					imgHeight := cell.Image.Height
					if imgWidth > cellWidth-10 {
						imgWidth = cellWidth - 10
					}
					if imgHeight > rowHeight-10 {
						imgHeight = rowHeight - 10
					}

					imgX := cellX + (cellWidth-imgWidth)/2
					imgY := pageManager.CurrentYPos - (rowHeight+imgHeight)/2

					// Draw actual image using XObject
					contentStream.WriteString("q\n")
					contentStream.WriteString(fmt.Sprintf("%.2f 0 0 %.2f %.2f %.2f cm\n",
						imgWidth, imgHeight, imgX, imgY))
					contentStream.WriteString(fmt.Sprintf("/CellImg_%s Do\n", cellKey))
					contentStream.WriteString("Q\n")
				} else {
					// Fall back to placeholder if no XObject
					imgWidth := cell.Image.Width
					imgHeight := cell.Image.Height
					if imgWidth > cellWidth-10 {
						imgWidth = cellWidth - 10
					}
					if imgHeight > rowHeight-10 {
						imgHeight = rowHeight - 10
					}

					imgX := cellX + (cellWidth-imgWidth)/2
					imgY := pageManager.CurrentYPos - (rowHeight+imgHeight)/2

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
			} else if cell.Checkbox != nil {
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
func drawImageWithXObjectInternal(image models.Image, imageXObjectRef string, pageManager *PageManager, borderConfig, watermark string) {
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

	// Draw the image using XObject
	drawImageWithXObject(contentStream, image, imageXObjectRef, pageManager)
}
