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

const (
	pageWidth  = 595
	pageHeight = 842
	margin     = 72
)

func parseProps(props string) models.Props {
	parts := strings.Split(props, ":")
	if len(parts) < 4 {
		return models.Props{FontName: "Helvetica", FontSize: 12, Alignment: "left", Borders: [4]int{0, 0, 0, 0}}
	}

	fontSize, _ := strconv.Atoi(parts[1])
	if fontSize == 0 {
		fontSize = 12
	}

	borders := [4]int{0, 0, 0, 0}
	if len(parts) >= 7 {
		for i := 3; i < 7 && i < len(parts); i++ {
			borders[i-3], _ = strconv.Atoi(parts[i])
		}
	}

	return models.Props{
		FontName:  parts[0],
		FontSize:  fontSize,
		Alignment: parts[2],
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

// GenerateTemplatePDF generates a PDF document based on a template and sends it to the client.
func GenerateTemplatePDF(c *gin.Context, template models.PDFTemplate) {
	var pdfBuffer bytes.Buffer
	xrefOffsets := make(map[int]int)

	// PDF Header
	pdfBuffer.WriteString("%PDF-1.7\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Object 3: Page
	xrefOffsets[3] = pdfBuffer.Len()
	pdfBuffer.WriteString("3 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R ")
	pdfBuffer.WriteString("/Resources << /Font << /F1 5 0 R /F2 6 0 R >> >> >>\n")
	pdfBuffer.WriteString("endobj\n")

	// Generate content stream
	var contentStream bytes.Buffer
	yPosition := float64(pageHeight - margin)

	// Page border
	pageBorders := parseBorders(template.Config.PageBorder)
	if pageBorders[0] > 0 || pageBorders[1] > 0 || pageBorders[2] > 0 || pageBorders[3] > 0 {
		contentStream.WriteString("q\n")
		if pageBorders[0] > 0 { // left border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[0])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(margin), float64(margin), float64(margin), float64(pageHeight-margin)))
		}
		if pageBorders[1] > 0 { // right border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[1])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(pageWidth-margin), float64(margin), float64(pageWidth-margin), float64(pageHeight-margin)))
		}
		if pageBorders[2] > 0 { // top border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[2])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(margin), float64(pageHeight-margin), float64(pageWidth-margin), float64(pageHeight-margin)))
		}
		if pageBorders[3] > 0 { // bottom border
			contentStream.WriteString(fmt.Sprintf("%.2f w\n", float64(pageBorders[3])))
			contentStream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n",
				float64(margin), float64(margin), float64(pageWidth-margin), float64(margin)))
		}
		contentStream.WriteString("Q\n")
	}

	// Title
	titleProps := parseProps(template.Title.Props)
	contentStream.WriteString("BT\n")
	contentStream.WriteString("/F1 ")
	contentStream.WriteString(strconv.Itoa(titleProps.FontSize))
	contentStream.WriteString(" Tf\n")

	var titleX float64
	switch titleProps.Alignment {
	case "center":
		titleX = pageWidth / 2
	case "right":
		titleX = pageWidth - margin
	default:
		titleX = margin
	}

	yPosition -= float64(titleProps.FontSize + 20)
	contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", titleX, yPosition))
	contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", template.Title.Text))
	contentStream.WriteString("ET\n")

	yPosition -= 30

	// Tables
	for _, table := range template.Table {
		cellWidth := float64(pageWidth-2*margin) / float64(table.MaxColumns)

		for _, row := range table.Rows {
			rowHeight := float64(20)

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
							float64(cellProps.Borders[0]), cellX, yPosition-rowHeight, cellX, yPosition))
					}
					if cellProps.Borders[1] > 0 { // right
						contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
							float64(cellProps.Borders[1]), cellX+cellWidth, yPosition-rowHeight, cellX+cellWidth, yPosition))
					}
					if cellProps.Borders[2] > 0 { // top
						contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
							float64(cellProps.Borders[2]), cellX, yPosition, cellX+cellWidth, yPosition))
					}
					if cellProps.Borders[3] > 0 { // bottom
						contentStream.WriteString(fmt.Sprintf("%.2f w %.2f %.2f m %.2f %.2f l S\n",
							float64(cellProps.Borders[3]), cellX, yPosition-rowHeight, cellX+cellWidth, yPosition-rowHeight))
					}
					contentStream.WriteString("Q\n")
				}

				// Draw text or checkbox
				if cell.Checkbox != nil {
					// Draw checkbox
					checkboxSize := 10.0
					checkboxX := cellX + (cellWidth-checkboxSize)/2
					checkboxY := yPosition - (rowHeight+checkboxSize)/2

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
					// Draw text
					contentStream.WriteString("BT\n")
					contentStream.WriteString("/F1 ")
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

					textY := yPosition - rowHeight/2 - float64(cellProps.FontSize)/2
					contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", textX, textY))
					contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", cell.Text))
					contentStream.WriteString("ET\n")
				}
			}

			yPosition -= rowHeight
		}

		yPosition -= 20 // Space between tables
	}

	// Footer
	if template.Footer.Text != "" {
		footerProps := parseProps(template.Footer.Font)
		contentStream.WriteString("BT\n")
		contentStream.WriteString("/F1 ")
		contentStream.WriteString(strconv.Itoa(footerProps.FontSize))
		contentStream.WriteString(" Tf\n")

		var footerX float64
		switch footerProps.Alignment {
		case "center":
			footerX = pageWidth / 2
		case "right":
			footerX = pageWidth - margin
		default:
			footerX = margin
		}

		contentStream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", footerX, float64(margin+10)))
		contentStream.WriteString(fmt.Sprintf("(%s) Tj\n", template.Footer.Text))
		contentStream.WriteString("ET\n")
	}

	// Object 4: Content Stream
	xrefOffsets[4] = pdfBuffer.Len()
	pdfBuffer.WriteString("4 0 obj\n")
	pdfBuffer.WriteString(fmt.Sprintf("<< /Length %d >>\n", contentStream.Len()))
	pdfBuffer.WriteString("stream\n")
	pdfBuffer.Write(contentStream.Bytes())
	pdfBuffer.WriteString("\nendstream\nendobj\n")

	// Object 5: Font Helvetica
	xrefOffsets[5] = pdfBuffer.Len()
	pdfBuffer.WriteString("5 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Font /Subtype /Type1 /Name /F1 /BaseFont /Helvetica >>\n")
	pdfBuffer.WriteString("endobj\n")

	// Object 6: Font Helvetica-Bold
	xrefOffsets[6] = pdfBuffer.Len()
	pdfBuffer.WriteString("6 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Font /Subtype /Type1 /Name /F2 /BaseFont /Helvetica-Bold >>\n")
	pdfBuffer.WriteString("endobj\n")

	// Cross-reference table
	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString("xref\n0 7\n0000000000 65535 f \n")
	for i := 1; i <= 6; i++ {
		pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", xrefOffsets[i]))
	}

	// Trailer
	pdfBuffer.WriteString("trailer\n<< /Size 7 /Root 1 0 R >>\n")
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	// HTTP Response
	filename := fmt.Sprintf("template-pdf-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}
