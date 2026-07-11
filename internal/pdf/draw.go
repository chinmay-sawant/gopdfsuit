package pdf

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v5/typstsyntax"
)

// fmtNum formats a float with 2 decimal places (standard PDF precision)
func fmtNum(f float64) string {
	var buf [24]byte
	b := appendFmtNum(buf[:0], f)
	return string(b)
}

// formXObjectStream builds a Form XObject object body without fmt (PERF-35).
func formXObjectStream(fw, fh, stream string) string {
	var b strings.Builder
	var nbuf [20]byte
	b.Grow(96 + len(stream))
	b.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
	b.WriteString(fw)
	b.WriteByte(' ')
	b.WriteString(fh)
	b.WriteString("] /Resources << /ProcSet [/PDF] >> /Length ")
	b.Write(strconv.AppendInt(nbuf[:0], int64(len(stream)), 10))
	b.WriteString(" >> stream\n")
	b.WriteString(stream)
	b.WriteString("\nendstream")
	return b.String()
}

// bezierCirclePath builds a 4-curve circle path centered at origin with radius r.
func bezierCirclePath(r float64) string {
	k := 0.5523
	rk, nr, nrk := r*k, -r, -r*k
	var b strings.Builder
	b.Grow(160)
	// r 0 m  r rk  rk r  0 r c  -rk r  -r rk  -r 0 c  -r -rk  -rk -r  0 -r c  rk -r  r -rk  r 0 c h
	parts := [...]float64{r, r, rk, rk, r, r, nrk, r, nr, rk, nr, nr, nrk, nrk, nr, nr, rk, nr, r, nrk, r}
	// Layout matches previous fmt.Sprintf circle path
	write := func(v float64) { b.WriteString(fmtNum(v)); b.WriteByte(' ') }
	write(parts[0])
	b.WriteString("0 m ")
	write(parts[1])
	write(parts[2])
	write(parts[3])
	write(parts[4])
	b.WriteString("0 ")
	write(parts[5])
	b.WriteString("c ")
	write(parts[6])
	write(parts[7])
	write(parts[8])
	write(parts[9])
	write(parts[10])
	b.WriteString("0 c ")
	write(parts[11])
	write(parts[12])
	write(parts[13])
	write(parts[14])
	b.WriteString("0 ")
	write(parts[15])
	b.WriteString("c ")
	write(parts[16])
	write(parts[17])
	write(parts[18])
	write(parts[19])
	write(parts[20])
	b.WriteString("0 c h")
	return b.String()
}

// appendFmtNum appends a float formatted to 2 decimal places directly to dst.
// Uses integer math instead of strconv.AppendFloat to avoid the expensive
// bigFtoa/genericFtoa code path (~10% CPU in profiling).
func appendFmtNum(dst []byte, f float64) []byte {
	if f < 0 {
		dst = append(dst, '-')
		f = -f
	}
	// Round to 2 decimal places using integer math
	scaled := int64(f*100 + 0.5)
	intPart := scaled / 100
	fracPart := scaled % 100
	dst = strconv.AppendInt(dst, intPart, 10)
	if fracPart > 0 {
		// Combine '.' and leading zero in one append when possible (PERF-119)
		if fracPart < 10 {
			dst = append(dst, '.', '0')
			dst = strconv.AppendInt(dst, fracPart, 10)
		} else if fracPart%10 == 0 {
			// Trim trailing zero (e.g. 0.50 -> 0.5)
			dst = append(dst, '.')
			dst = strconv.AppendInt(dst, fracPart/10, 10)
		} else {
			dst = append(dst, '.')
			dst = strconv.AppendInt(dst, fracPart, 10)
		}
	}
	return dst
}

// --- new watermark drawer (diagonal bottom-left to top-right) ---
func drawWatermark(contentStream *bytes.Buffer, text string, pageDims PageDimensions, registry *CustomFontRegistry) {
	if trimSpace(text) == "" {
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
	if registry.HasFont("Helvetica") {
		registry.MarkCharsUsed("Helvetica", text)
	}

	// 45 degree rotation matrix components
	c := 0.7071
	s := 0.7071

	// Use props for proper font encoding
	watermarkProps := models.Props{FontName: "Helvetica", FontSize: fontSize}

	// Begin Artifact mark
	contentStream.WriteString("/Artifact <</Attached [/Top] /Type /Pagination >> BDC\n")

	contentStream.WriteString("q\n")
	// Light gray fill/stroke
	contentStream.WriteString("0.85 0.85 0.85 rg 0.85 0.85 0.85 RG\n")
	contentStream.WriteString("BT\n")
	// Use getFontReference to handle PDF/A font substitution (Helvetica -> Liberation)
	fontRef := getFontReference(watermarkProps, registry)

	// Pre-allocate buffer and build complete watermark command sequence
	resolvedName := resolveFontName(watermarkProps, registry)
	textCmd := formatTextForPDF(resolvedName, text, registry)
	var sizeBuf [12]byte
	sizeNum := strconv.AppendInt(sizeBuf[:0], int64(fontSize), 10)
	var wb strings.Builder
	wb.Grow(256 + len(fontRef) + 1 + len(sizeNum) + 4 + len(textCmd))
	wb.WriteString(fontRef)
	wb.WriteByte(' ')
	wb.Write(sizeNum)
	wb.WriteString(" Tf\n")
	tmp := make([]byte, 0, 48)
	tmp = appendFmtNum(tmp, c)
	wb.Write(tmp)
	wb.WriteByte(' ')
	tmp = tmp[:0]
	tmp = appendFmtNum(tmp, s)
	wb.Write(tmp)
	wb.WriteByte(' ')
	tmp = tmp[:0]
	tmp = appendFmtNum(tmp, -s)
	wb.Write(tmp)
	wb.WriteByte(' ')
	tmp = tmp[:0]
	tmp = appendFmtNum(tmp, c)
	wb.Write(tmp)
	wb.WriteByte(' ')
	tmp = tmp[:0]
	tmp = appendFmtNum(tmp, x)
	wb.Write(tmp)
	wb.WriteByte(' ')
	tmp = tmp[:0]
	tmp = appendFmtNum(tmp, y)
	wb.Write(tmp)
	wb.WriteString(" Tm\n")
	wb.WriteString(string(textCmd))
	wb.WriteString(" Tj\n")

	// Single write for entire watermark command sequence
	contentStream.WriteString(wb.String())
	contentStream.WriteString("ET\nQ\n")

	// End Artifact mark
	contentStream.WriteString("EMC\n")
}

// --- new page initializer (border + watermark) ---
func initializePage(contentStream *bytes.Buffer, borderConfig, watermark string, pageDims PageDimensions, margins PageMargins, registry *CustomFontRegistry) {
	drawPageBorder(contentStream, borderConfig, pageDims, margins)
	if watermark != "" {
		drawWatermark(contentStream, watermark, pageDims, registry)
	}
}

// drawPageBorder draws the page border
func drawPageBorder(contentStream *bytes.Buffer, borderConfig string, pageDims PageDimensions, margins PageMargins) {
	pageBorders := parseBorders(borderConfig)
	if pageBorders[0] > 0 || pageBorders[1] > 0 || pageBorders[2] > 0 || pageBorders[3] > 0 {
		// Begin Artifact mark
		contentStream.WriteString("/Artifact <</Attached [/Top] /Type /Pagination >> BDC\n")

		contentStream.WriteString("q\n")
		// Pre-allocate buffer for border drawing commands
		borderBuf := make([]byte, 0, 128)

		if pageBorders[0] > 0 { // left border
			borderBuf = borderBuf[:0]
			borderBuf = strconv.AppendInt(borderBuf, int64(pageBorders[0]), 10)
			borderBuf = append(borderBuf, " w\n"...)
			borderBuf = appendFmtNum(borderBuf, margins.Left)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, margins.Bottom)
			borderBuf = append(borderBuf, " m "...)
			borderBuf = appendFmtNum(borderBuf, margins.Left)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, pageDims.Height-margins.Top)
			borderBuf = append(borderBuf, " l S\n"...)
			contentStream.Write(borderBuf)
		}
		if pageBorders[1] > 0 { // right border
			borderBuf = borderBuf[:0]
			borderBuf = strconv.AppendInt(borderBuf, int64(pageBorders[1]), 10)
			borderBuf = append(borderBuf, " w\n"...)
			borderBuf = appendFmtNum(borderBuf, pageDims.Width-margins.Right)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, margins.Bottom)
			borderBuf = append(borderBuf, " m "...)
			borderBuf = appendFmtNum(borderBuf, pageDims.Width-margins.Right)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, pageDims.Height-margins.Top)
			borderBuf = append(borderBuf, " l S\n"...)
			contentStream.Write(borderBuf)
		}
		if pageBorders[2] > 0 { // top border
			borderBuf = borderBuf[:0]
			borderBuf = strconv.AppendInt(borderBuf, int64(pageBorders[2]), 10)
			borderBuf = append(borderBuf, " w\n"...)
			borderBuf = appendFmtNum(borderBuf, margins.Left)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, pageDims.Height-margins.Top)
			borderBuf = append(borderBuf, " m "...)
			borderBuf = appendFmtNum(borderBuf, pageDims.Width-margins.Right)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, pageDims.Height-margins.Top)
			borderBuf = append(borderBuf, " l S\n"...)
			contentStream.Write(borderBuf)
		}
		if pageBorders[3] > 0 { // bottom border
			borderBuf = borderBuf[:0]
			borderBuf = strconv.AppendInt(borderBuf, int64(pageBorders[3]), 10)
			borderBuf = append(borderBuf, " w\n"...)
			borderBuf = appendFmtNum(borderBuf, margins.Left)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, margins.Bottom)
			borderBuf = append(borderBuf, " m "...)
			borderBuf = appendFmtNum(borderBuf, pageDims.Width-margins.Right)
			borderBuf = append(borderBuf, ' ')
			borderBuf = appendFmtNum(borderBuf, margins.Bottom)
			borderBuf = append(borderBuf, " l S\n"...)
			contentStream.Write(borderBuf)
		}
		contentStream.WriteString("Q\n")

		// End Artifact mark
		contentStream.WriteString("EMC\n")
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
	// PDF/UA: Start Heading Structure Element wrapping EVERYTHING (Background + Text)
	var sb strings.Builder
	pageManager.Structure.BeginMarkedContent(&sb, pageManager.CurrentPageIndex, StructH1, map[string]string{"Title": title.Text})
	contentStream.WriteString(sb.String())

	// Draw background color if specified
	if r, g, b, _, valid := parseHexColor(title.BgColor); valid {
		rectX := pageManager.Margins.Left
		rectY := pageManager.CurrentYPos - float64(titleProps.FontSize)
		rectW := pageManager.ContentWidth()
		rectH := float64(titleProps.FontSize)

		contentStream.WriteString("q\n")
		// PERF-119: Builder for color+rect ops
		var cb strings.Builder
		cb.Grow(96)
		tmp := make([]byte, 0, 32)
		tmp = appendFmtNum(tmp, r)
		cb.Write(tmp)
		cb.WriteByte(' ')
		tmp = tmp[:0]
		tmp = appendFmtNum(tmp, g)
		cb.Write(tmp)
		cb.WriteByte(' ')
		tmp = tmp[:0]
		tmp = appendFmtNum(tmp, b)
		cb.Write(tmp)
		cb.WriteString(" rg\n")
		tmp = tmp[:0]
		tmp = appendFmtNum(tmp, rectX)
		cb.Write(tmp)
		cb.WriteByte(' ')
		tmp = tmp[:0]
		tmp = appendFmtNum(tmp, rectY)
		cb.Write(tmp)
		cb.WriteByte(' ')
		tmp = tmp[:0]
		tmp = appendFmtNum(tmp, rectW)
		cb.Write(tmp)
		cb.WriteByte(' ')
		tmp = tmp[:0]
		tmp = appendFmtNum(tmp, rectH)
		cb.Write(tmp)
		cb.WriteString(" re f\n")
		contentStream.WriteString(cb.String())
		contentStream.WriteString("Q\n")
	}

	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(titleProps, pageManager.FontRegistry))
	contentStream.WriteString(" ")
	var titleFsBuf [12]byte
	contentStream.Write(strconv.AppendInt(titleFsBuf[:0], int64(titleProps.FontSize), 10))
	contentStream.WriteString(" Tf\n")

	// Set text color
	if r, g, b, _, valid := parseHexColor(title.TextColor); valid {
		colorBuf := make([]byte, 0, 32)
		colorBuf = append(appendFmtNum(colorBuf, r), ' ')
		colorBuf = append(appendFmtNum(colorBuf, g), ' ')
		colorBuf = append(appendFmtNum(colorBuf, b), " rg\n"...)
		contentStream.Write(colorBuf)
	} else {
		contentStream.WriteString("0 0 0 rg\n")
	}

	// Mark chars used for subsetting calculation
	pageManager.FontRegistry.MarkCharsUsed(titleProps.FontName, title.Text)

	// Calculate approximate text width
	resolvedName := resolveFontName(titleProps, pageManager.FontRegistry)
	textWidth := EstimateTextWidth(resolvedName, title.Text, float64(titleProps.FontSize), pageManager.FontRegistry)

	// Calculate available width (page width minus both margins)
	availableWidth := pageManager.ContentWidth()

	var titleX float64
	switch titleProps.Alignment {
	case "center": //nolint:goconst
		// Center the text within the available area (between margins)
		titleX = pageManager.Margins.Left + (availableWidth-textWidth)/2
	case "right": //nolint:goconst
		// Right align: position text so it ends at the right margin
		titleX = pageManager.PageDimensions.Width - pageManager.Margins.Right - textWidth
	default:
		titleX = pageManager.Margins.Left
	}

	pageManager.CurrentYPos -= float64(titleProps.FontSize)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	var titleBuf []byte
	titleBuf = appendFmtNum(titleBuf, titleX)
	titleBuf = append(titleBuf, ' ')
	titleBuf = appendFmtNum(titleBuf, pageManager.CurrentYPos)
	titleBuf = append(titleBuf, " Td\n"...)
	contentStream.Write(titleBuf)

	titleBuf = titleBuf[:0]
	titleBuf = append(titleBuf, formatTextForPDF(resolvedName, title.Text, pageManager.FontRegistry)...)
	titleBuf = append(titleBuf, " Tj\n"...)
	contentStream.Write(titleBuf)
	contentStream.WriteString("ET\n")

	// PDF/UA: End Structure Element
	pageManager.Structure.EndMarkedContentBuf(contentStream)

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
//
//nolint:gocyclo
func drawTitleTable(contentStream *bytes.Buffer, table *models.TitleTable, pageManager *PageManager, cellImageObjectIDs map[string]int, defaultBgColor, defaultTextColor string, defaultProps models.Props) {
	availableWidth := pageManager.ContentWidth()
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

	// PDF/UA: Start Table Structure Element (Logical grouping)
	pageManager.Structure.BeginStructureElement(StructTable)

	// Precompute stable values to avoid re-evaluation in loops (PERF-230)
	helveticaFontRef := getFontReference(models.Props{FontName: "Helvetica"}, pageManager.FontRegistry)
	defaultFontRef := getFontReference(defaultProps, pageManager.FontRegistry)
	defaultResolvedName := resolveFontName(defaultProps, pageManager.FontRegistry)
	var defaultTextR, defaultTextG, defaultTextB float64
	var defaultTextHasColor bool
	if dr, dg, db, _, dvalid := parseHexColor(defaultTextColor); dvalid {
		defaultTextR, defaultTextG, defaultTextB, defaultTextHasColor = dr, dg, db, true
	}
	var defaultBgR, defaultBgG, defaultBgB float64
	var defaultBgHasColor bool
	if dbr, dbg, dbb, _, dbvalid := parseHexColor(defaultBgColor); dbvalid {
		defaultBgR, defaultBgG, defaultBgB, defaultBgHasColor = dbr, dbg, dbb, true
	}
	rowEst := len(table.Rows)
	propsCache := make(map[string]models.Props, rowEst)
	fontRefCache := make(map[string]string, rowEst)
	resolvedFontCache := make(map[string]string, rowEst)
	colorCache := make(map[string]struct {
		r, g, b float64
		valid   bool
	}, rowEst*2)

	for rowIdx, row := range table.Rows {
		// Determine this row's height
		rowHeight := baseRowHeight
		for _, cell := range row.Row {
			if cell.Height != nil && *cell.Height > rowHeight {
				rowHeight = *cell.Height
			}
		}

		// Draw row cells - Pass 1: Backgrounds
		// PDF/UA: Mark backgrounds as Artifacts
		contentStream.WriteString("/Artifact BMC\n")
		bgX := pageManager.Margins.Left
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
			bgColor := cell.BgColor
			if bgColor == "" {
				bgColor = defaultBgColor
			}
			if bgColor == defaultBgColor && defaultBgHasColor {
				contentStream.WriteString("q\n")
				bgBuf := make([]byte, 0, 64)
				bgBuf = appendFmtNum(bgBuf, defaultBgR)
				bgBuf = append(bgBuf, ' ')
				bgBuf = appendFmtNum(bgBuf, defaultBgG)
				bgBuf = append(bgBuf, ' ')
				bgBuf = appendFmtNum(bgBuf, defaultBgB)
				bgBuf = append(bgBuf, " rg\n"...)
				contentStream.Write(bgBuf)
				bgBuf = bgBuf[:0]
				bgBuf = appendFmtNum(bgBuf, bgX)
				bgBuf = append(bgBuf, ' ')
				bgBuf = appendFmtNum(bgBuf, pageManager.CurrentYPos-cellHeight)
				bgBuf = append(bgBuf, ' ')
				bgBuf = appendFmtNum(bgBuf, cellWidth)
				bgBuf = append(bgBuf, ' ')
				bgBuf = appendFmtNum(bgBuf, cellHeight)
				bgBuf = append(bgBuf, " re f\n"...)
				contentStream.Write(bgBuf)
				contentStream.WriteString("Q\n")
			} else {
				cc, ok := colorCache[bgColor]
				if !ok {
					r, g, b, _, valid := parseHexColor(bgColor)
					cc.r, cc.g, cc.b, cc.valid = r, g, b, valid
					colorCache[bgColor] = cc
				}
				if cc.valid {
					contentStream.WriteString("q\n")
					bgBuf := make([]byte, 0, 64)
					bgBuf = appendFmtNum(bgBuf, cc.r)
					bgBuf = append(bgBuf, ' ')
					bgBuf = appendFmtNum(bgBuf, cc.g)
					bgBuf = append(bgBuf, ' ')
					bgBuf = appendFmtNum(bgBuf, cc.b)
					bgBuf = append(bgBuf, " rg\n"...)
					contentStream.Write(bgBuf)

					bgBuf = bgBuf[:0]
					bgBuf = appendFmtNum(bgBuf, bgX)
					bgBuf = append(bgBuf, ' ')
					bgBuf = appendFmtNum(bgBuf, pageManager.CurrentYPos-cellHeight)
					bgBuf = append(bgBuf, ' ')
					bgBuf = appendFmtNum(bgBuf, cellWidth)
					bgBuf = append(bgBuf, ' ')
					bgBuf = appendFmtNum(bgBuf, cellHeight)
					bgBuf = append(bgBuf, " re f\n"...)
					contentStream.Write(bgBuf)
					contentStream.WriteString("Q\n")
				}
			}

			bgX += cellWidth
		}
		contentStream.WriteString("EMC\n")

		// Draw row cells - Pass 2: Content and Borders
		// PDF/UA: Start TR Structure Element
		pageManager.Structure.BeginStructureElement(StructTR)

		currentX := pageManager.Margins.Left
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			// PDF/UA: Start TD Structure Element
			pageManager.Structure.BeginMarkedContentBuf(contentStream, pageManager.CurrentPageIndex, StructTD, nil)

			// Capture cell coordinates for link
			// Capture cell coordinates for link

			var cellProps models.Props
			if cell.Props == "" {
				cellProps = defaultProps
			} else {
				cp, ok := propsCache[cell.Props]
				if !ok {
					cp = parseProps(cell.Props)
					propsCache[cell.Props] = cp
				}
				cellProps = cp
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
				cellKey := buildCellKey2("title", rowIdx, colIdx)
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
					var imgBuf []byte
					imgBuf = appendFmtNum(imgBuf, imgX)
					imgBuf = append(imgBuf, ' ')
					imgBuf = appendFmtNum(imgBuf, imgY)
					imgBuf = append(imgBuf, ' ')
					imgBuf = appendFmtNum(imgBuf, imgWidth)
					imgBuf = append(imgBuf, ' ')
					imgBuf = appendFmtNum(imgBuf, imgHeight)
					imgBuf = append(imgBuf, " re W n\n"...)
					contentStream.Write(imgBuf)

					imgBuf = imgBuf[:0]
					imgBuf = appendFmtNum(imgBuf, imgWidth)
					imgBuf = append(imgBuf, " 0 0 "...)
					imgBuf = appendFmtNum(imgBuf, imgHeight)
					imgBuf = append(imgBuf, ' ')
					imgBuf = appendFmtNum(imgBuf, imgX)
					imgBuf = append(imgBuf, ' ')
					imgBuf = appendFmtNum(imgBuf, imgY)
					imgBuf = append(imgBuf, " cm\n"...)
					contentStream.Write(imgBuf)

					imgBuf = imgBuf[:0]
					imgBuf = append(imgBuf, "/C"...)
					imgBuf = append(imgBuf, shortKey...)
					imgBuf = append(imgBuf, " Do\n"...)
					contentStream.Write(imgBuf)
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
					var placeholderBuf []byte
					placeholderBuf = appendFmtNum(placeholderBuf, imgX)
					placeholderBuf = append(placeholderBuf, ' ')
					placeholderBuf = appendFmtNum(placeholderBuf, imgY)
					placeholderBuf = append(placeholderBuf, ' ')
					placeholderBuf = appendFmtNum(placeholderBuf, imgWidth)
					placeholderBuf = append(placeholderBuf, ' ')
					placeholderBuf = appendFmtNum(placeholderBuf, imgHeight)
					placeholderBuf = append(placeholderBuf, " re S\n"...)
					contentStream.Write(placeholderBuf)
					contentStream.WriteString("Q\n")

					// Draw image name
					if cell.Image.ImageName != "" && len(cell.Image.ImageName) < 20 {
						contentStream.WriteString("BT\n")
						var imgNameBuf []byte
						imgNameBuf = append(imgNameBuf, helveticaFontRef...)
						imgNameBuf = append(imgNameBuf, " 8 Tf\n"...)
						contentStream.Write(imgNameBuf)
						contentStream.WriteString("0.5 0.5 0.5 rg\n")
						textX := imgX + imgWidth/2 - float64(len(cell.Image.ImageName)*2)
						textY := imgY + imgHeight/2
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						imgNameBuf = imgNameBuf[:0]
						imgNameBuf = appendFmtNum(imgNameBuf, textX)
						imgNameBuf = append(imgNameBuf, ' ')
						imgNameBuf = appendFmtNum(imgNameBuf, textY)
						imgNameBuf = append(imgNameBuf, " Td\n"...)
						contentStream.Write(imgNameBuf)
						imgNameBuf = imgNameBuf[:0]
						imgNameBuf = append(imgNameBuf, '(')
						imgNameBuf = append(imgNameBuf, escapeText(cell.Image.ImageName)...)
						imgNameBuf = append(imgNameBuf, ") Tj\n"...)
						contentStream.Write(imgNameBuf)
						contentStream.WriteString("ET\n")
					}
				}
			} else if cell.Text != "" {
				// Draw text with font styling
				contentStream.WriteString("BT\n")
				if cell.Props == "" {
					contentStream.WriteString(defaultFontRef)
				} else {
					fr, ok := fontRefCache[cell.Props]
					if !ok {
						fr = getFontReference(cellProps, pageManager.FontRegistry)
						fontRefCache[cell.Props] = fr
					}
					contentStream.WriteString(fr)
				}
				contentStream.WriteString(" ")
				var cellFsBuf [12]byte
				contentStream.Write(strconv.AppendInt(cellFsBuf[:0], int64(cellProps.FontSize), 10))
				contentStream.WriteString(" Tf\n")

				// Set text color - always explicitly set to avoid state leakage, default to black
				// Use cell-specific color if available, otherwise use default (title-level) color
				textColor := cell.TextColor
				if textColor == "" {
					textColor = defaultTextColor
				}
			var colorBuf []byte
			if textColor == defaultTextColor && defaultTextHasColor {
				colorBuf = colorBuf[:0]
				colorBuf = append(appendFmtNum(colorBuf, defaultTextR), ' ')
				colorBuf = append(appendFmtNum(colorBuf, defaultTextG), ' ')
				colorBuf = append(appendFmtNum(colorBuf, defaultTextB), " rg\n"...)
				contentStream.Write(colorBuf)
			} else if tc, ok := colorCache[textColor]; ok && tc.valid {
				colorBuf = colorBuf[:0]
				colorBuf = append(appendFmtNum(colorBuf, tc.r), ' ')
				colorBuf = append(appendFmtNum(colorBuf, tc.g), ' ')
				colorBuf = append(appendFmtNum(colorBuf, tc.b), " rg\n"...)
				contentStream.Write(colorBuf)
			} else if r, g, b, _, valid := parseHexColor(textColor); valid {
				colorBuf = colorBuf[:0]
				colorBuf = append(appendFmtNum(colorBuf, r), ' ')
				colorBuf = append(appendFmtNum(colorBuf, g), ' ')
				colorBuf = append(appendFmtNum(colorBuf, b), " rg\n"...)
				contentStream.Write(colorBuf)
				colorCache[textColor] = struct {
					r, g, b float64
					valid   bool
				}{r, g, b, true}
			} else {
				// Default to black if no valid color specified
				contentStream.WriteString("0 0 0 rg\n")
			}

				// Calculate approximate text width
				var resolvedName string
				if cell.Props == "" {
					resolvedName = defaultResolvedName
				} else {
					rn, ok := resolvedFontCache[cell.Props]
					if !ok {
						rn = resolveFontName(cellProps, pageManager.FontRegistry)
						resolvedFontCache[cell.Props] = rn
					}
					resolvedName = rn
				}
				textWidth := EstimateTextWidth(resolvedName, cell.Text, float64(cellProps.FontSize), pageManager.FontRegistry)

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
				var textPosBuf []byte
				textPosBuf = appendFmtNum(textPosBuf, textX)
				textPosBuf = append(textPosBuf, ' ')
				textPosBuf = appendFmtNum(textPosBuf, textY)
				textPosBuf = append(textPosBuf, " Td\n"...)
				contentStream.Write(textPosBuf)

				// Add underline support
				if cellProps.Underline {
					contentStream.WriteString("ET\n")
					contentStream.WriteString("q\n")
					contentStream.WriteString("0.5 w\n")
					underlineY := textY - 2
					textWidth := float64(len(cell.Text) * cellProps.FontSize / 2)
					var underlineBuf []byte
					underlineBuf = appendFmtNum(underlineBuf, textX)
					underlineBuf = append(underlineBuf, ' ')
					underlineBuf = appendFmtNum(underlineBuf, underlineY)
					underlineBuf = append(underlineBuf, " m "...)
					underlineBuf = appendFmtNum(underlineBuf, textX+textWidth)
					underlineBuf = append(underlineBuf, ' ')
					underlineBuf = appendFmtNum(underlineBuf, underlineY)
					underlineBuf = append(underlineBuf, " l S\n"...)
					contentStream.Write(underlineBuf)
					contentStream.WriteString("Q\n")
					contentStream.WriteString("BT\n")
					if cell.Props == "" {
						contentStream.WriteString(defaultFontRef)
					} else {
						fr, ok := fontRefCache[cell.Props]
						if !ok {
							fr = getFontReference(cellProps, pageManager.FontRegistry)
							fontRefCache[cell.Props] = fr
						}
						contentStream.WriteString(fr)
					}
					contentStream.WriteString(" ")
					contentStream.Write(strconv.AppendInt(cellFsBuf[:0], int64(cellProps.FontSize), 10))
					contentStream.WriteString(" Tf\n")
					contentStream.WriteString("1 0 0 1 0 0 Tm\n")
					textPosBuf = textPosBuf[:0]
					textPosBuf = appendFmtNum(textPosBuf, textX)
					textPosBuf = append(textPosBuf, ' ')
					textPosBuf = appendFmtNum(textPosBuf, textY)
					textPosBuf = append(textPosBuf, " Td\n"...)
					contentStream.Write(textPosBuf)
				}

				// Mark chars used for subsetting
				pageManager.FontRegistry.MarkCharsUsed(cellProps.FontName, cell.Text)

				textPosBuf = textPosBuf[:0]
				textPosBuf = append(textPosBuf, formatTextForPDF(resolvedName, cell.Text, pageManager.FontRegistry)...)
				textPosBuf = append(textPosBuf, " Tj\n"...)
				contentStream.Write(textPosBuf)
				contentStream.WriteString("ET\n")
			}

			// Draw cell borders AFTER content (so they appear on top of images)
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					var borderBuf []byte
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[0]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				if cellProps.Borders[1] > 0 { // right
					var borderBuf []byte
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[1]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				if cellProps.Borders[2] > 0 { // top
					var borderBuf []byte
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[2]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				if cellProps.Borders[3] > 0 { // bottom
					var borderBuf []byte
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[3]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				contentStream.WriteString("Q\n")
			}

			// Add Link Annotation if provided
			if cell.Link != "" {
				// Use captured cellStartX
				linkY := pageManager.CurrentYPos - cellHeight
				pageManager.AddLinkAnnotation(cellX, linkY, cellWidth, cellHeight, cell.Link)
			}
			// Register named destination anchor if provided
			if cell.Dest != "" {
				pageManager.NamedDests[cell.Dest] = NamedDest{
					PageIndex: pageManager.CurrentPageIndex,
					Y:         pageManager.CurrentYPos,
				}
			}
			// PDF/UA: End TD Structure Element
			pageManager.Structure.EndMarkedContentBuf(contentStream)
		}

		// PDF/UA: End TR Structure Element
		pageManager.Structure.EndStructureElement()

		pageManager.CurrentYPos -= rowHeight
	}

	// PDF/UA: End Table Structure Element
	pageManager.Structure.EndStructureElement()
}

// drawTable renders a table with automatic page breaks
//
//nolint:gocyclo
func drawTable(table models.Table, imageKeyPrefix string, pageManager *PageManager, borderConfig, watermark string, cellImageObjectIDs map[string]int) {
	availableWidth := pageManager.ContentWidth()
	baseRowHeight := float64(25) // Standard row height

	// PDF/UA: Start Table Structure
	pageManager.Structure.BeginStructureElement(StructTable)
	defer pageManager.Structure.EndStructureElement()

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

	// Reuse buffers for row processing to reduce allocations
	cellWidthsForRow := make([]float64, table.MaxColumns)
	wrappedTextLines := make([][][]byte, table.MaxColumns)
	rowCellProps := make([]models.Props, table.MaxColumns)
	rowResolvedFonts := make([]string, table.MaxColumns)
	// Scratch buffers reused across all cells to reduce allocations
	scratchBuf := make([]byte, 0, 128)
	borderBuf := make([]byte, 0, 64)
	xobjBuf := make([]byte, 0, 96)
	colorBuf := make([]byte, 0, 48)
	placeholderBuf := make([]byte, 0, 64)
	checkboxBuf := make([]byte, 0, 64)
	var wrapState WrapState

	// Caching maps to avoid repeated calls in hot loops (PERF-230)
	rowEst2 := len(table.Rows)
	propsCache := make(map[string]models.Props, rowEst2)
	fontRefCache := make(map[string]string, rowEst2)
	resolvedFontCache := make(map[string]string, rowEst2)
	colorCache := make(map[string]struct {
		r, g, b float64
		valid   bool
	}, rowEst2*2)

	// Pre-populate caches with all distinct cell values
	for _, row := range table.Rows {
		for _, cell := range row.Row {
			if _, ok := propsCache[cell.Props]; !ok {
				cp := parseProps(cell.Props)
				propsCache[cell.Props] = cp
				fontRefCache[cell.Props] = getFontReference(cp, pageManager.FontRegistry)
				resolvedFontCache[cell.Props] = resolveFontName(cp, pageManager.FontRegistry)
			}
			bgColor := cell.BgColor
			if bgColor == "" {
				bgColor = table.BgColor
			}
			if _, ok := colorCache[bgColor]; !ok {
				if r, g, b, _, valid := parseHexColor(bgColor); valid {
					colorCache[bgColor] = struct {
						r, g, b float64
						valid   bool
					}{r, g, b, true}
				}
			}
			textColor := cell.TextColor
			if textColor == "" {
				textColor = table.TextColor
			}
			if _, ok := colorCache[textColor]; !ok {
				if r, g, b, _, valid := parseHexColor(textColor); valid {
					colorCache[textColor] = struct {
						r, g, b float64
						valid   bool
					}{r, g, b, true}
				}
			}
		}
	}

	for rowIdx, row := range table.Rows {
		// PDF/UA: Start Row Structure
		pageManager.Structure.BeginStructureElement(StructTR)

		// Pre-calculate column widths and parsed props
		// We need to know each cell's width before calculating wrapped text height
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}
			// Reset wrapped text lines for this cell to avoid staleness
			wrappedTextLines[colIdx] = nil

			cellWidthsForRow[colIdx] = colWidths[colIdx]
			if cell.Width != nil && *cell.Width > 0 {
				cellWidthsForRow[colIdx] = *cell.Width
			}
		}

		// Pre-calculate wrapped text lines for cells that have wrap enabled (default: enabled)
		// This is used both for height calculation and later for rendering
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			// Parse props once per cell and cache it
			cellProps := propsCache[cell.Props]
			rowCellProps[colIdx] = cellProps

			// Resolve font name once per cell — used for text width, wrapping, and rendering
			rowResolvedFonts[colIdx] = resolvedFontCache[cell.Props]

			// Wrap is opt-in (only enabled when explicitly set to true)
			isWrapEnabled := cell.Wrap != nil && *cell.Wrap
			if isWrapEnabled && cell.Text != "" {
				// Account for cell padding (5pt on each side)
				maxTextWidth := cellWidthsForRow[colIdx] - 10
				if maxTextWidth < 10 {
					maxTextWidth = 10 // Minimum width to avoid issues
				}
				wrappedTextLines[colIdx] = WrapTextInto(&wrapState, cell.Text, rowResolvedFonts[colIdx], float64(cellProps.FontSize), maxTextWidth, pageManager.FontRegistry)
			}

			// Mark chars used for subsetting (once per cell)
			if cell.Text != "" {
				pageManager.FontRegistry.MarkCharsUsed(cellProps.FontName, cell.Text)
			}
		}

		// Determine this row's height - check if any cell in row has custom height
		rowHeight := baseRowHeight
		if rowIdx < len(table.RowHeights) && table.RowHeights[rowIdx] > 0 {
			rowHeight = baseRowHeight * table.RowHeights[rowIdx]
		}
		// Override with max cell height if any cell specifies it
		for colIdx, cell := range row.Row {
			if cell.Height != nil && *cell.Height > rowHeight {
				rowHeight = *cell.Height
			}
			// Calculate height needed for wrapped text (opt-in)
			isWrapEnabled := cell.Wrap != nil && *cell.Wrap
			if isWrapEnabled && len(wrappedTextLines[colIdx]) > 0 {
				cellProps := rowCellProps[colIdx] // Use cached props
				lineSpacing := 1.3                // 130% line height for readability
				padding := 12.0                   // Top + bottom padding
				wrappedHeight := CalculateWrappedTextHeight(len(wrappedTextLines[colIdx]), float64(cellProps.FontSize), lineSpacing) + padding
				if wrappedHeight > rowHeight {
					rowHeight = wrappedHeight
				}
			}
		}

		// Check if row fits on current page
		if pageManager.CheckPageBreak(rowHeight) {
			// Create new page and initialize it
			pageManager.AddNewPage()
			initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions, pageManager.Margins, pageManager.FontRegistry)
		}

		// Get current content stream for this page
		contentStream := pageManager.GetCurrentContentStream()

		// Draw row cells
		currentX := pageManager.Margins.Left
		for colIdx, cell := range row.Row {
			if colIdx >= table.MaxColumns {
				break
			}

			// PDF/UA: Start Cell Structure (TH for header if first row, else TD)
			// Assuming first row is header if Table has explicit header concept, but for now just use TD
			// Could be enhanced to detect header rows
			cellType := StructTD
			pageManager.Structure.BeginMarkedContentBuf(contentStream, pageManager.CurrentPageIndex, cellType, nil)

			cellProps := rowCellProps[colIdx] // Use cached props
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
			if cc, ok := colorCache[bgColor]; ok && cc.valid {
				contentStream.WriteString("q\n")
				bgColorBuf := appendFmtNum(scratchBuf[:0], cc.r)
				bgColorBuf = append(bgColorBuf, ' ')
				bgColorBuf = appendFmtNum(bgColorBuf, cc.g)
				bgColorBuf = append(bgColorBuf, ' ')
				bgColorBuf = appendFmtNum(bgColorBuf, cc.b)
				bgColorBuf = append(bgColorBuf, " rg\n"...)
				contentStream.Write(bgColorBuf)

				bgColorBuf = bgColorBuf[:0]
				bgColorBuf = appendFmtNum(bgColorBuf, cellX)
				bgColorBuf = append(bgColorBuf, ' ')
				bgColorBuf = appendFmtNum(bgColorBuf, pageManager.CurrentYPos-cellHeight)
				bgColorBuf = append(bgColorBuf, ' ')
				bgColorBuf = appendFmtNum(bgColorBuf, cellWidth)
				bgColorBuf = append(bgColorBuf, ' ')
				bgColorBuf = appendFmtNum(bgColorBuf, cellHeight)
				bgColorBuf = append(bgColorBuf, " re f\n"...)
				contentStream.Write(bgColorBuf)
				contentStream.WriteString("Q\n")
			}

			// Draw content (so borders are drawn on top of images)
			switch {
			case cell.Image != nil:
				// Check if we have an XObject for this cell image
				cellKey := buildCellKey2(imageKeyPrefix, rowIdx, colIdx)
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
					xobjBuf = xobjBuf[:0]
					xobjBuf = appendFmtNum(xobjBuf, imgX)
					xobjBuf = append(xobjBuf, ' ')
					xobjBuf = appendFmtNum(xobjBuf, imgY)
					xobjBuf = append(xobjBuf, ' ')
					xobjBuf = appendFmtNum(xobjBuf, imgWidth)
					xobjBuf = append(xobjBuf, ' ')
					xobjBuf = appendFmtNum(xobjBuf, imgHeight)
					xobjBuf = append(xobjBuf, " re W n\n"...)
					contentStream.Write(xobjBuf)

					xobjBuf = xobjBuf[:0]
					xobjBuf = appendFmtNum(xobjBuf, imgWidth)
					xobjBuf = append(xobjBuf, " 0 0 "...)
					xobjBuf = appendFmtNum(xobjBuf, imgHeight)
					xobjBuf = append(xobjBuf, ' ')
					xobjBuf = appendFmtNum(xobjBuf, imgX)
					xobjBuf = append(xobjBuf, ' ')
					xobjBuf = appendFmtNum(xobjBuf, imgY)
					xobjBuf = append(xobjBuf, " cm\n"...)
					contentStream.Write(xobjBuf)

					xobjBuf = xobjBuf[:0]
					xobjBuf = append(xobjBuf, "/C"...)
					xobjBuf = append(xobjBuf, shortKey...)
					xobjBuf = append(xobjBuf, " Do\n"...)
					contentStream.Write(xobjBuf)
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
					placeholderBuf = placeholderBuf[:0]
					placeholderBuf = appendFmtNum(placeholderBuf, imgX)
					placeholderBuf = append(placeholderBuf, ' ')
					placeholderBuf = appendFmtNum(placeholderBuf, imgY)
					placeholderBuf = append(placeholderBuf, ' ')
					placeholderBuf = appendFmtNum(placeholderBuf, imgWidth)
					placeholderBuf = append(placeholderBuf, ' ')
					placeholderBuf = appendFmtNum(placeholderBuf, imgHeight)
					placeholderBuf = append(placeholderBuf, " re S\n"...)
					contentStream.Write(placeholderBuf)
					contentStream.WriteString("Q\n")

					// Draw image name
					if cell.Image.ImageName != "" && len(cell.Image.ImageName) < 20 {
						contentStream.WriteString("BT\n")
						fontRef := getFontReference(models.Props{FontName: "Helvetica"}, pageManager.FontRegistry)
						var imgNameBuf []byte
						imgNameBuf = append(imgNameBuf, fontRef...)
						imgNameBuf = append(imgNameBuf, " 8 Tf\n"...)
						contentStream.Write(imgNameBuf)
						contentStream.WriteString("0.5 0.5 0.5 rg\n")
						textX := imgX + imgWidth/2 - float64(len(cell.Image.ImageName)*2)
						textY := imgY + imgHeight/2
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						imgNameBuf = imgNameBuf[:0]
						imgNameBuf = appendFmtNum(imgNameBuf, textX)
						imgNameBuf = append(imgNameBuf, ' ')
						imgNameBuf = appendFmtNum(imgNameBuf, textY)
						imgNameBuf = append(imgNameBuf, " Td\n"...)
						contentStream.Write(imgNameBuf)
						imgNameBuf = imgNameBuf[:0]
						imgNameBuf = append(imgNameBuf, '(')
						imgNameBuf = append(imgNameBuf, escapeText(cell.Image.ImageName)...)
						imgNameBuf = append(imgNameBuf, ") Tj\n"...)
						contentStream.Write(imgNameBuf)
						contentStream.WriteString("ET\n")
					}
				}
			case cell.FormField != nil:
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

			case cell.Checkbox != nil:
				// Draw checkbox using 're' operator
				checkboxSize := 10.0
				checkboxX := cellX + (cellWidth-checkboxSize)/2
				checkboxY := pageManager.CurrentYPos - (cellHeight+checkboxSize)/2

				contentStream.WriteString("q\n")
				contentStream.WriteString("1 w\n")
				checkboxBuf = checkboxBuf[:0]
				checkboxBuf = appendFmtNum(checkboxBuf, checkboxX)
				checkboxBuf = append(checkboxBuf, ' ')
				checkboxBuf = appendFmtNum(checkboxBuf, checkboxY)
				checkboxBuf = append(checkboxBuf, ' ')
				checkboxBuf = appendFmtNum(checkboxBuf, checkboxSize)
				checkboxBuf = append(checkboxBuf, ' ')
				checkboxBuf = appendFmtNum(checkboxBuf, checkboxSize)
				checkboxBuf = append(checkboxBuf, " re S\n"...)
				contentStream.Write(checkboxBuf)

				if *cell.Checkbox {
					checkboxBuf = checkboxBuf[:0]
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxX+2)
					checkboxBuf = append(checkboxBuf, ' ')
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxY+2)
					checkboxBuf = append(checkboxBuf, " m "...)
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxX+checkboxSize-2)
					checkboxBuf = append(checkboxBuf, ' ')
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxY+checkboxSize-2)
					checkboxBuf = append(checkboxBuf, " l "...)
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxX+checkboxSize-2)
					checkboxBuf = append(checkboxBuf, ' ')
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxY+2)
					checkboxBuf = append(checkboxBuf, " m "...)
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxX+2)
					checkboxBuf = append(checkboxBuf, ' ')
					checkboxBuf = appendFmtNum(checkboxBuf, checkboxY+checkboxSize-2)
					checkboxBuf = append(checkboxBuf, " l S\n"...)
					contentStream.Write(checkboxBuf)
				}
				contentStream.WriteString("Q\n")
			case cell.MathEnabled != nil && *cell.MathEnabled && cell.Text != "" && typstsyntax.IsMathExpression(cell.Text):
				// Draw math expression using Typst syntax parser + glyph-based rendering
				textColor := cell.TextColor
				if textColor == "" {
					textColor = table.TextColor
				}
				colorStr := "0 0 0"
				if cc, ok := colorCache[textColor]; ok && cc.valid {
					colorStr = fmtNum(cc.r) + " " + fmtNum(cc.g) + " " + fmtNum(cc.b)
				}

				// Set up render context with font callbacks
				mathCtx := &typstsyntax.RenderContext{
					FontSize:   float64(cellProps.FontSize),
					FontRef:    fontRefCache[cell.Props],
					CellWidth:  cellWidth,
					CellHeight: cellHeight,
					TextColor:  colorStr,
					EstimateWidth: func(text string, fontSize float64) float64 {
						return EstimateTextWidth(rowResolvedFonts[colIdx], text, fontSize, pageManager.FontRegistry)
					},
					FormatText: func(text string) string {
						// Math rendering emits many Unicode glyph fragments (integrals, set symbols,
						// superscripts/subscripts). Mark them so custom-font subsetting keeps them.
						pageManager.FontRegistry.MarkCharsUsed(rowResolvedFonts[colIdx], text)
						return formatTextForPDF(rowResolvedFonts[colIdx], text, pageManager.FontRegistry)
					},
				}

				// Calculate layout
				layout := typstsyntax.RenderMathToLayout(cell.Text, mathCtx)

				// Center the math expression within the cell
				var mathX float64
				switch cellProps.Alignment {
				case "center":
					mathX = cellX + (cellWidth-layout.Width)/2
				case "right":
					mathX = cellX + cellWidth - layout.Width - 5
				default:
					mathX = cellX + 5
				}
				mathY := pageManager.CurrentYPos - cellHeight/2 - float64(cellProps.FontSize)/2

				// Set position in context and render
				mathCtx.X = mathX
				mathCtx.Y = mathY
				typstsyntax.RenderToContentStream(contentStream, layout, mathCtx)

			case cell.Text != "":
				// Draw text with font styling
				contentStream.WriteString("BT\n")
				contentStream.WriteString(fontRefCache[cell.Props])
				contentStream.WriteString(" ")
				var tableCellFsBuf [12]byte
				contentStream.Write(strconv.AppendInt(tableCellFsBuf[:0], int64(cellProps.FontSize), 10))
				contentStream.WriteString(" Tf\n")

				// Set text color (cell-level takes precedence over table-level, default to black)
				// Always explicitly set the color to avoid state leakage from previous tables
				textColor := cell.TextColor
				if textColor == "" {
					textColor = table.TextColor
				}
				if cc, ok := colorCache[textColor]; ok && cc.valid {
					colorBuf = colorBuf[:0]
					colorBuf = appendFmtNum(colorBuf, cc.r)
					colorBuf = append(colorBuf, ' ')
					colorBuf = appendFmtNum(colorBuf, cc.g)
					colorBuf = append(colorBuf, ' ')
					colorBuf = appendFmtNum(colorBuf, cc.b)
					colorBuf = append(colorBuf, " rg\n"...)
					contentStream.Write(colorBuf)
				} else {
					// Default to black if no valid color specified
					contentStream.WriteString("0 0 0 rg\n")
				}

				// Check if this cell has wrapped text (opt-in)
				isWrapEnabled := cell.Wrap != nil && *cell.Wrap
				if isWrapEnabled && len(wrappedTextLines[colIdx]) > 0 {
					// Multi-line text rendering for wrapped cells
					lines := wrappedTextLines[colIdx]
					lineSpacing := 1.3
					fontSize := float64(cellProps.FontSize)
					lineHeight := fontSize * lineSpacing

					// Calculate starting Y position (top-aligned with padding)
					topPadding := 4.0
					startY := pageManager.CurrentYPos - topPadding - fontSize

					for lineIdx, line := range lines {
						if len(line) == 0 {
							continue
						}

						lineEstWidth := EstimateTextWidth(rowResolvedFonts[colIdx], byteString(line), fontSize, pageManager.FontRegistry)

						// Calculate X position based on alignment
						var textX float64
						switch cellProps.Alignment {
						case "center":
							textX = cellX + (cellWidth-lineEstWidth)/2
						case "right":
							textX = cellX + cellWidth - lineEstWidth - 5
						default:
							textX = cellX + 5
						}

						// Calculate Y position for this line
						textY := startY - float64(lineIdx)*lineHeight

						// Reset text matrix and position
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						textPosBuf := appendFmtNum(scratchBuf[:0], textX)
						textPosBuf = append(textPosBuf, ' ')
						textPosBuf = appendFmtNum(textPosBuf, textY)
						textPosBuf = append(textPosBuf, " Td\n"...)
						contentStream.Write(textPosBuf)

						// Render the line
						textPosBuf = appendTextForPDF(textPosBuf[:0], rowResolvedFonts[colIdx], byteString(line), pageManager.FontRegistry)
						textPosBuf = append(textPosBuf, " Tj\n"...)
						contentStream.Write(textPosBuf)
					}
					contentStream.WriteString("ET\n")
				} else {
					// Single-line text rendering (original behavior)
					// Calculate approximate text width
					resolvedName := rowResolvedFonts[colIdx]
					textWidth := EstimateTextWidth(resolvedName, cell.Text, float64(cellProps.FontSize), pageManager.FontRegistry)

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
					textPosBuf := appendFmtNum(scratchBuf[:0], textX)
					textPosBuf = append(textPosBuf, ' ')
					textPosBuf = appendFmtNum(textPosBuf, textY)
					textPosBuf = append(textPosBuf, " Td\n"...)
					contentStream.Write(textPosBuf)

					// Add underline support
					if cellProps.Underline {
						// End text object before drawing underline
						contentStream.WriteString("ET\n")
						contentStream.WriteString("q\n")
						contentStream.WriteString("0.5 w\n")
						underlineY := textY - 2
						textWidth := float64(len(cell.Text) * cellProps.FontSize / 2)
						underlineBuf := appendFmtNum(scratchBuf[:0], textX)
						underlineBuf = append(underlineBuf, ' ')
						underlineBuf = appendFmtNum(underlineBuf, underlineY)
						underlineBuf = append(underlineBuf, " m "...)
						underlineBuf = appendFmtNum(underlineBuf, textX+textWidth)
						underlineBuf = append(underlineBuf, ' ')
						underlineBuf = appendFmtNum(underlineBuf, underlineY)
						underlineBuf = append(underlineBuf, " l S\n"...)
						contentStream.Write(underlineBuf)
						contentStream.WriteString("Q\n")
						// Start text object again
						contentStream.WriteString("BT\n")
						contentStream.WriteString(fontRefCache[cell.Props])
						contentStream.WriteString(" ")
						contentStream.Write(strconv.AppendInt(tableCellFsBuf[:0], int64(cellProps.FontSize), 10))
						contentStream.WriteString(" Tf\n")
						contentStream.WriteString("1 0 0 1 0 0 Tm\n")
						textPosBuf = textPosBuf[:0]
						textPosBuf = appendFmtNum(textPosBuf, textX)
						textPosBuf = append(textPosBuf, ' ')
						textPosBuf = appendFmtNum(textPosBuf, textY)
						textPosBuf = append(textPosBuf, " Td\n"...)
						contentStream.Write(textPosBuf)
					}

					textPosBuf = appendTextForPDF(textPosBuf[:0], resolvedName, cell.Text, pageManager.FontRegistry)
					textPosBuf = append(textPosBuf, " Tj\n"...)
					contentStream.Write(textPosBuf)
					contentStream.WriteString("ET\n")
				}
			}

			// Draw cell borders AFTER content (so they appear on top of images)
			if cellProps.Borders[0] > 0 || cellProps.Borders[1] > 0 || cellProps.Borders[2] > 0 || cellProps.Borders[3] > 0 {
				contentStream.WriteString("q\n")
				if cellProps.Borders[0] > 0 { // left
					borderBuf = borderBuf[:0]
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[0]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				if cellProps.Borders[1] > 0 { // right
					borderBuf = borderBuf[:0]
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[1]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				if cellProps.Borders[2] > 0 { // top
					borderBuf = borderBuf[:0]
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[2]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				if cellProps.Borders[3] > 0 { // bottom
					borderBuf = borderBuf[:0]
					borderBuf = strconv.AppendInt(borderBuf, int64(cellProps.Borders[3]), 10)
					borderBuf = append(borderBuf, " w "...)
					borderBuf = appendFmtNum(borderBuf, cellX)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " m "...)
					borderBuf = appendFmtNum(borderBuf, cellX+cellWidth)
					borderBuf = append(borderBuf, ' ')
					borderBuf = appendFmtNum(borderBuf, pageManager.CurrentYPos-cellHeight)
					borderBuf = append(borderBuf, " l S\n"...)
					contentStream.Write(borderBuf)
				}
				contentStream.WriteString("Q\n")
			}

			// Create link annotation if cell has a link
			if cell.Link != "" {
				DrawCellLink(cell.Link, cellX, pageManager.CurrentYPos-cellHeight, cellWidth, cellHeight, pageManager)
			}
			// Register named destination anchor if provided
			if cell.Dest != "" {
				pageManager.NamedDests[cell.Dest] = NamedDest{
					PageIndex: pageManager.CurrentPageIndex,
					Y:         pageManager.CurrentYPos,
				}
			}

			// PDF/UA: End Cell Structure
			pageManager.Structure.EndMarkedContentBuf(contentStream)
		}

		// PDF/UA: End Row Structure
		pageManager.Structure.EndStructureElement()

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
	// PDF/UA: Start Artifact mark (Footer)
	contentStream.WriteString("/Artifact <</Attached [/Bottom] /Type /Pagination >> BDC\n")

	contentStream.WriteString("BT\n")
	contentStream.WriteString(getFontReference(footerProps, pageManager.FontRegistry))
	contentStream.WriteString(" ")
	var footerFsBuf [12]byte
	contentStream.Write(strconv.AppendInt(footerFsBuf[:0], int64(footerProps.FontSize), 10))
	contentStream.WriteString(" Tf\n")

	// Position footer outside the page border on the left side
	footerX := 20 // 20pt from left edge (outside margin)
	footerY := 20 // 20pt from bottom edge (outside margin)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	var footerBuf []byte
	footerBuf = strconv.AppendInt(footerBuf, int64(footerX), 10)
	footerBuf = append(footerBuf, ' ')
	footerBuf = strconv.AppendInt(footerBuf, int64(footerY), 10)
	footerBuf = append(footerBuf, " Td\n"...)
	contentStream.Write(footerBuf)

	// Mark chars used for subsetting
	pageManager.FontRegistry.MarkCharsUsed(footerProps.FontName, footer.Text)

	footerBuf = footerBuf[:0]
	// Resolve font name
	resolvedName := resolveFontName(footerProps, pageManager.FontRegistry)
	footerBuf = append(footerBuf, formatTextForPDF(resolvedName, footer.Text, pageManager.FontRegistry)...)
	footerBuf = append(footerBuf, " Tj\n"...)
	contentStream.Write(footerBuf)
	contentStream.WriteString("ET\n")

	// PDF/UA: End Artifact mark
	contentStream.WriteString("EMC\n")

	// Add Link Annotation if provided
	if footer.Link != "" {
		// Calculate approximate text width
		// Using standard estimation for width since we don't have exact calc here easily without refactoring
		// But footer is likely simple text.
		textWidth := EstimateTextWidth(footerProps.FontName, footer.Text, float64(footerProps.FontSize), pageManager.FontRegistry)

		rectX := float64(footerX)
		rectY := float64(footerY) - float64(footerProps.FontSize)*0.2
		rectW := textWidth
		rectH := float64(footerProps.FontSize) * 1.2
		pageManager.AddLinkAnnotation(rectX, rectY, rectW, rectH, footer.Link)
	}
}

// drawPageNumber renders page number in bottom right corner
func drawPageNumber(contentStream *bytes.Buffer, currentPage, totalPages int, pageDims PageDimensions, pageManager *PageManager) {
	var pageTmp [20]byte
	pageText := "Page " + string(strconv.AppendInt(pageTmp[:0], int64(currentPage), 10)) +
		" of " + string(strconv.AppendInt(pageTmp[:0], int64(totalPages), 10))

	// Track characters for font subsetting
	registry := pageManager.FontRegistry
	if registry.HasFont("Helvetica") {
		registry.MarkCharsUsed("Helvetica", pageText)
	}

	// Use props for proper font encoding
	pageProps := models.Props{FontName: "Helvetica", FontSize: 10}

	// PDF/UA: Start Artifact mark (PageNum)
	contentStream.WriteString("/Artifact <</Attached [/Bottom] /Type /Pagination >> BDC\n")

	contentStream.WriteString("BT\n")
	fontRef := getFontReference(pageProps, pageManager.FontRegistry)
	var pageNumBuf []byte
	pageNumBuf = append(pageNumBuf, fontRef...)
	pageNumBuf = append(pageNumBuf, " 10 Tf\n"...)
	contentStream.Write(pageNumBuf) // Use Helvetica, 10pt

	// Calculate text width for proper right alignment
	textWidth := float64(len(pageText)) * 6 // Approximate character width for 10pt font

	// Position outside the page border on the right side
	pageNumberX := pageDims.Width - textWidth - 20 // 20pt from right edge (outside margin)
	pageNumberY := 20                              // 20pt from bottom edge (outside margin)

	contentStream.WriteString("1 0 0 1 0 0 Tm\n") // Reset text matrix
	pageNumBuf = pageNumBuf[:0]
	pageNumBuf = appendFmtNum(pageNumBuf, pageNumberX)
	pageNumBuf = append(pageNumBuf, ' ')
	pageNumBuf = strconv.AppendInt(pageNumBuf, int64(pageNumberY), 10)
	pageNumBuf = append(pageNumBuf, " Td\n"...)
	contentStream.Write(pageNumBuf)

	pageNumBuf = pageNumBuf[:0]
	// resolvedName not available here, need to resolve using pageProps
	resolvedName := resolveFontName(pageProps, pageManager.FontRegistry)
	pageNumBuf = append(pageNumBuf, formatTextForPDF(resolvedName, pageText, pageManager.FontRegistry)...)
	pageNumBuf = append(pageNumBuf, " Tj\n"...)
	contentStream.Write(pageNumBuf)
	contentStream.WriteString("ET\n")

	// PDF/UA: End Artifact mark
	contentStream.WriteString("EMC\n")
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
		initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions, pageManager.Margins, pageManager.FontRegistry)
	}

	// Get current content stream for this page
	contentStream := pageManager.GetCurrentContentStream()

	// PDF/UA: Start Figure structure
	var sb strings.Builder
	props := map[string]string{}
	if image.ImageName != "" {
		props["Alt"] = image.ImageName
	} else {
		props["Alt"] = "Image"
	}
	pageManager.Structure.BeginMarkedContent(&sb, pageManager.CurrentPageIndex, StructFigure, props)
	contentStream.WriteString(sb.String())

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
	var imgBorderBuf []byte
	imgBorderBuf = appendFmtNum(imgBorderBuf, imageX)
	imgBorderBuf = append(imgBorderBuf, ' ')
	imgBorderBuf = appendFmtNum(imgBorderBuf, imageY)
	imgBorderBuf = append(imgBorderBuf, ' ')
	imgBorderBuf = appendFmtNum(imgBorderBuf, imageWidth)
	imgBorderBuf = append(imgBorderBuf, ' ')
	imgBorderBuf = appendFmtNum(imgBorderBuf, imageHeight)
	imgBorderBuf = append(imgBorderBuf, " re S\n"...)
	contentStream.Write(imgBorderBuf)
	contentStream.WriteString("Q\n")

	// Add image name text in the center
	if image.ImageName != "" {
		contentStream.WriteString("BT\n")
		fontRef := getFontReference(models.Props{FontName: "Helvetica"}, pageManager.FontRegistry)
		var imgTextBuf []byte
		imgTextBuf = append(imgTextBuf, fontRef...)
		imgTextBuf = append(imgTextBuf, " 10 Tf\n"...)
		contentStream.Write(imgTextBuf)
		contentStream.WriteString("0.6 0.6 0.6 rg\n") // Gray text

		// Center the text
		textX := imageX + imageWidth/2
		textY := imageY + imageHeight/2

		contentStream.WriteString("1 0 0 1 0 0 Tm\n")
		imgTextBuf = imgTextBuf[:0]
		imgTextBuf = appendFmtNum(imgTextBuf, textX)
		imgTextBuf = append(imgTextBuf, ' ')
		imgTextBuf = appendFmtNum(imgTextBuf, textY)
		imgTextBuf = append(imgTextBuf, " Td\n"...)
		contentStream.Write(imgTextBuf)

		imgTextBuf = imgTextBuf[:0]
		imgTextBuf = append(imgTextBuf, '(')
		imgTextBuf = append(imgTextBuf, escapeText(image.ImageName)...)
		imgTextBuf = append(imgTextBuf, ") Tj\n"...)
		contentStream.Write(imgTextBuf)
		contentStream.WriteString("ET\n")
	}

	// Add Link Annotation if provided
	if image.Link != "" {
		pageManager.AddLinkAnnotation(imageX, imageY, imageWidth, imageHeight, image.Link)
	}

	// PDF/UA: End Figure structure
	pageManager.Structure.EndMarkedContentBuf(contentStream)

	pageManager.CurrentYPos -= (imageHeight + spacing)
}

// drawImageWithXObjectInternal handles image drawing with XObject, including page breaks
func drawImageWithXObjectInternal(image models.Image, imageXObjectRef string, pageManager *PageManager, borderConfig, watermark string, originalImgWidth, originalImgHeight int) {
	// Calculate usable width to estimate height for page break check
	usableWidth := pageManager.ContentWidth()

	// Calculate height based on aspect ratio
	var imageHeight float64
	switch {
	case originalImgWidth > 0 && originalImgHeight > 0:
		aspectRatio := float64(originalImgHeight) / float64(originalImgWidth)
		imageHeight = usableWidth * aspectRatio
	case image.Height > 0 && image.Width > 0:
		aspectRatio := image.Height / image.Width
		imageHeight = usableWidth * aspectRatio
	default:
		imageHeight = 200 // Default height
	}

	// Check if image fits on current page (no extra spacing)
	if pageManager.CheckPageBreak(imageHeight) {
		// Create new page and initialize it
		pageManager.AddNewPage()
		initializePage(pageManager.GetCurrentContentStream(), borderConfig, watermark, pageManager.PageDimensions, pageManager.Margins, pageManager.FontRegistry)
	}

	// Get current content stream for this page
	contentStream := pageManager.GetCurrentContentStream()

	// PDF/UA: Start Figure structure
	var sb strings.Builder
	props := map[string]string{}
	if image.ImageName != "" {
		props["Alt"] = image.ImageName
	} else {
		props["Alt"] = "Image"
	}
	pageManager.Structure.BeginMarkedContent(&sb, pageManager.CurrentPageIndex, StructFigure, props)
	contentStream.WriteString(sb.String())

	// Draw the image using XObject
	drawImageWithXObject(contentStream, image, imageXObjectRef, pageManager, originalImgWidth, originalImgHeight)

	// PDF/UA: End Figure structure
	pageManager.Structure.EndMarkedContentBuf(contentStream)
}

// drawWidget creates a widget annotation for a form field
func drawWidget(cell models.Cell, x, y, w, h float64, pageManager *PageManager) {
	if cell.FormField == nil {
		return
	}

	field := cell.FormField
	// Calculate rect without fmt boxing (PERF-35)
	var rectB strings.Builder
	rectB.Grow(48)
	rectB.WriteByte('[')
	rectB.WriteString(fmtNum(x))
	rectB.WriteByte(' ')
	rectB.WriteString(fmtNum(y))
	rectB.WriteByte(' ')
	rectB.WriteString(fmtNum(x + w))
	rectB.WriteByte(' ')
	rectB.WriteString(fmtNum(y + h))
	rectB.WriteByte(']')
	rect := rectB.String()

	var widgetDict strings.Builder
	widgetDict.Grow(512)
	widgetDict.WriteString("<< /Type /Annot /Subtype /Widget")
	widgetDict.WriteString(" /Rect ")
	widgetDict.WriteString(rect)
	widgetDict.WriteString(" /T (")
	widgetDict.WriteString(escapeText(field.Name))
	widgetDict.WriteByte(')')
	widgetDict.WriteString(" /F 4") // Print flag
	var nbuf [20]byte

	switch field.Type {
	case "checkbox":
		widgetDict.WriteString(" /FT /Btn")

		onState := "/Yes"
		offState := "/Off"

		val := offState
		if field.Checked {
			val = onState
		}

		widgetDict.WriteString(" /V ")
		widgetDict.WriteString(val)
		widgetDict.WriteString(" /AS ")
		widgetDict.WriteString(val)

		// Checkbox Appearance Streams using 're' operator
		fw, fh := fmtNum(w), fmtNum(h)
		fw2, fh2 := fmtNum(w-2), fmtNum(h-2)
		var onAPB strings.Builder
		onAPB.Grow(96)
		onAPB.WriteString("q 1 w 0 0 0 RG 0 0 ")
		onAPB.WriteString(fw)
		onAPB.WriteByte(' ')
		onAPB.WriteString(fh)
		onAPB.WriteString(" re S 2 2 m ")
		onAPB.WriteString(fw2)
		onAPB.WriteByte(' ')
		onAPB.WriteString(fh2)
		onAPB.WriteString(" l 2 ")
		onAPB.WriteString(fh2)
		onAPB.WriteString(" m ")
		onAPB.WriteString(fw2)
		onAPB.WriteString(" 2 l S Q")
		onAP := onAPB.String()
		var onObjB strings.Builder
		onObjB.Grow(96 + len(onAP))
		onObjB.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
		onObjB.WriteString(fw)
		onObjB.WriteByte(' ')
		onObjB.WriteString(fh)
		onObjB.WriteString("] /Resources << /ProcSet [/PDF] >> /Length ")
		onObjB.Write(strconv.AppendInt(nbuf[:0], int64(len(onAP)), 10))
		onObjB.WriteString(" >> stream\n")
		onObjB.WriteString(onAP)
		onObjB.WriteString("\nendstream")
		onAPID := pageManager.AddExtraObject(onObjB.String())

		// Off Appearance (Empty Box)
		var offAPB strings.Builder
		offAPB.Grow(48)
		offAPB.WriteString("q 1 w 0 0 0 RG 0 0 ")
		offAPB.WriteString(fw)
		offAPB.WriteByte(' ')
		offAPB.WriteString(fh)
		offAPB.WriteString(" re S Q")
		offAP := offAPB.String()
		var offObjB strings.Builder
		offObjB.Grow(96 + len(offAP))
		offObjB.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
		offObjB.WriteString(fw)
		offObjB.WriteByte(' ')
		offObjB.WriteString(fh)
		offObjB.WriteString("] /Resources << /ProcSet [/PDF] >> /Length ")
		offObjB.Write(strconv.AppendInt(nbuf[:0], int64(len(offAP)), 10))
		offObjB.WriteString(" >> stream\n")
		offObjB.WriteString(offAP)
		offObjB.WriteString("\nendstream")
		offAPID := pageManager.AddExtraObject(offObjB.String())

		widgetDict.WriteString(" /AP << /N << /Yes ")
		widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(onAPID), 10))
		widgetDict.WriteString(" 0 R /Off ")
		widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(offAPID), 10))
		widgetDict.WriteString(" 0 R >> >>")

	case "radio":
		widgetDict.WriteString(" /FT /Btn /Ff 49152") // Radio button flag

		onState := "/" + field.Value
		offState := "/Off"

		val := offState
		if field.Checked {
			val = onState
		}

		widgetDict.WriteString(" /V ")
		widgetDict.WriteString(val)
		widgetDict.WriteString(" /AS ")
		widgetDict.WriteString(val)

		if field.Shape == "square" {
			// Radio Appearance Streams (Square with dot) using 're' operator
			fw, fh := fmtNum(w), fmtNum(h)
			fw6, fh6 := fmtNum(w-6), fmtNum(h-6)
			var onAPB strings.Builder
			onAPB.WriteString("q 1 w 0 0 0 RG 0 0 ")
			onAPB.WriteString(fw)
			onAPB.WriteByte(' ')
			onAPB.WriteString(fh)
			onAPB.WriteString(" re S 3 3 ")
			onAPB.WriteString(fw6)
			onAPB.WriteByte(' ')
			onAPB.WriteString(fh6)
			onAPB.WriteString(" re f Q")
			onAP := onAPB.String()
			onAPID := pageManager.AddExtraObject(formXObjectStream(fw, fh, onAP))

			var offAPB strings.Builder
			offAPB.WriteString("q 1 w 0 0 0 RG 0 0 ")
			offAPB.WriteString(fw)
			offAPB.WriteByte(' ')
			offAPB.WriteString(fh)
			offAPB.WriteString(" re S Q")
			offAP := offAPB.String()
			offAPID := pageManager.AddExtraObject(formXObjectStream(fw, fh, offAP))

			widgetDict.WriteString(" /AP << /N << /")
			widgetDict.WriteString(field.Value)
			widgetDict.WriteByte(' ')
			widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(onAPID), 10))
			widgetDict.WriteString(" 0 R /Off ")
			widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(offAPID), 10))
			widgetDict.WriteString(" 0 R >> >>")
		} else {
			// Default to Round (Circle)
			// Add /MK dictionary with appearance characteristics for circle radio button
			widgetDict.WriteString(" /MK << /BC [0 0 0] /BG [0.9 0.9 0.9] /CA (l) >>")

			// Center point and radius calculations
			cx := w / 2
			cy := h / 2
			outerR := cx - 0.5      // Outer circle radius
			innerR := outerR * 0.45 // Inner dot radius
			outerCirclePath := bezierCirclePath(outerR)
			innerCirclePath := bezierCirclePath(innerR)

			// ON appearance: Light background fill + dark stroke + dark inner dot
			fcx, fcy := fmtNum(cx), fmtNum(cy)
			fw, fh := fmtNum(w), fmtNum(h)
			var onAPB strings.Builder
			onAPB.Grow(64 + len(outerCirclePath) + len(innerCirclePath))
			onAPB.WriteString("q\n0.9 0.9 0.9 rg 0 0 0 RG 1 w\n1 0 0 1 ")
			onAPB.WriteString(fcx)
			onAPB.WriteByte(' ')
			onAPB.WriteString(fcy)
			onAPB.WriteString(" cm\n")
			onAPB.WriteString(outerCirclePath)
			onAPB.WriteString("\nB\nQ\nq\n0 0 0 rg\n1 0 0 1 ")
			onAPB.WriteString(fcx)
			onAPB.WriteByte(' ')
			onAPB.WriteString(fcy)
			onAPB.WriteString(" cm\n")
			onAPB.WriteString(innerCirclePath)
			onAPB.WriteString("\nf\nQ")
			onAP := onAPB.String()
			onAPID := pageManager.AddExtraObject(formXObjectStream(fw, fh, onAP))

			// OFF appearance: Light background fill + dark stroke (no inner dot)
			var offAPB strings.Builder
			offAPB.Grow(48 + len(outerCirclePath))
			offAPB.WriteString("q\n0.9 0.9 0.9 rg 0 0 0 RG 1 w\n1 0 0 1 ")
			offAPB.WriteString(fcx)
			offAPB.WriteByte(' ')
			offAPB.WriteString(fcy)
			offAPB.WriteString(" cm\n")
			offAPB.WriteString(outerCirclePath)
			offAPB.WriteString("\nB\nQ")
			offAP := offAPB.String()
			offAPID := pageManager.AddExtraObject(formXObjectStream(fw, fh, offAP))

			widgetDict.WriteString(" /AP << /N << /")
			widgetDict.WriteString(field.Value)
			widgetDict.WriteByte(' ')
			widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(onAPID), 10))
			widgetDict.WriteString(" 0 R /Off ")
			widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(offAPID), 10))
			widgetDict.WriteString(" 0 R >> >>")
		}
	case "text":
		widgetDict.WriteString(" /FT /Tx") // Text field
		widgetDict.WriteString(" /V (")
		widgetDict.WriteString(escapeText(field.Value))
		widgetDict.WriteByte(')')

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
			registry := pageManager.FontRegistry
			if registry.HasFont("Helvetica") {
				// PDF/A mode: Liberation font registered as Helvetica
				registry.MarkCharsUsed("Helvetica", field.Value)
			}
		}

		// Get the appropriate font reference for widgets (handles PDF/A mode)
		widgetFontRef := getWidgetFontReference(pageManager.FontRegistry)

		// Default Appearance string - used by viewer to render text
		// Use proper font reference instead of hardcoded /Helv
		ffs := fmtNum(fontSize)
		fw, fh := fmtNum(w), fmtNum(h)
		widgetDict.WriteString(" /DA (")
		widgetDict.WriteString(widgetFontRef)
		widgetDict.WriteByte(' ')
		widgetDict.WriteString(ffs)
		widgetDict.WriteString(" Tf 0 g)")

		// Build appearance stream: border + text properly structured
		// Use /Tx BMC ... EMC to mark text content area (viewer replaces this when editing)
		var apStream strings.Builder
		apStream.Grow(256)
		// Draw border first
		apStream.WriteString("q 1 w 0 0 0 RG 0 0 ")
		apStream.WriteString(fw)
		apStream.WriteByte(' ')
		apStream.WriteString(fh)
		apStream.WriteString(" re S Q ")
		// Text content marked with /Tx BMC ... EMC (marked content)
		// This tells PDF viewer this is the text area it should manage
		apStream.WriteString("/Tx BMC ")
		if field.Value != "" {
			textY := (h - fontSize) / 2
			textX := 2.0
			// Use proper encoding for field value
			fieldProps := models.Props{FontName: "Helvetica", FontSize: int(fontSize)}
			encodedValue := formatTextForPDF(resolveFontName(fieldProps, pageManager.FontRegistry), field.Value, pageManager.FontRegistry)
			// Use proper font reference in appearance stream
			apStream.WriteString("q BT ")
			apStream.WriteString(widgetFontRef)
			apStream.WriteByte(' ')
			apStream.WriteString(ffs)
			apStream.WriteString(" Tf 0 g ")
			apStream.WriteString(fmtNum(textX))
			apStream.WriteByte(' ')
			apStream.WriteString(fmtNum(textY))
			apStream.WriteString(" Td ")
			apStream.WriteString(encodedValue)
			apStream.WriteString(" Tj ET Q ")
		}
		apStream.WriteString("EMC")
		apContent := apStream.String()

		// Create appearance XObject
		// IMPORTANT: Form XObjects must declare all resources they use in their own Resources dictionary
		// This is required for PDF/A-4 compliance - resources cannot be inherited from page level
		var apObjContent string
		if getWidgetFontName(pageManager.FontRegistry) == "" {
			// PDF/A mode: Get the font object ID from the font registry
			// The widgetFontRef (e.g., /CF2000) references a custom font that must be in Resources
			fontObjID := getWidgetFontObjectID(pageManager.FontRegistry)
			if fontObjID > 0 {
				// Include the font reference in the XObject's Resources dictionary
				var apB strings.Builder
				apB.Grow(128 + len(apContent) + len(widgetFontRef))
				apB.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
				apB.WriteString(fw)
				apB.WriteByte(' ')
				apB.WriteString(fh)
				apB.WriteString("] /Resources << /Font << ")
				apB.WriteString(widgetFontRef)
				apB.WriteByte(' ')
				apB.Write(strconv.AppendInt(nbuf[:0], int64(fontObjID), 10))
				apB.WriteString(" 0 R >> >> /Length ")
				apB.Write(strconv.AppendInt(nbuf[:0], int64(len(apContent)), 10))
				apB.WriteString(" >> stream\n")
				apB.WriteString(apContent)
				apB.WriteString("\nendstream")
				apObjContent = apB.String()
			} else {
				// Fallback: empty resources (should not happen in PDF/A mode)
				var apB strings.Builder
				apB.Grow(96 + len(apContent))
				apB.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
				apB.WriteString(fw)
				apB.WriteByte(' ')
				apB.WriteString(fh)
				apB.WriteString("] /Resources << >> /Length ")
				apB.Write(strconv.AppendInt(nbuf[:0], int64(len(apContent)), 10))
				apB.WriteString(" >> stream\n")
				apB.WriteString(apContent)
				apB.WriteString("\nendstream")
				apObjContent = apB.String()
			}
		} else {
			// Standard mode: Embed Helvetica definition in XObject resources
			var helveticaFont string
			if pageManager.ArlingtonCompatible {
				helveticaFont = GetHelveticaFontResourceString()
			} else {
				helveticaFont = GetSimpleHelveticaFontResourceString()
			}
			var apB strings.Builder
			apB.Grow(128 + len(helveticaFont) + len(apContent))
			apB.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
			apB.WriteString(fw)
			apB.WriteByte(' ')
			apB.WriteString(fh)
			apB.WriteString("] /Resources << /Font << /F1 ")
			apB.WriteString(helveticaFont)
			apB.WriteString(" >> >> /Length ")
			apB.Write(strconv.AppendInt(nbuf[:0], int64(len(apContent)), 10))
			apB.WriteString(" >> stream\n")
			apB.WriteString(apContent)
			apB.WriteString("\nendstream")
			apObjContent = apB.String()
		}
		apID := pageManager.AddExtraObject(apObjContent)

		widgetDict.WriteString(" /AP << /N ")
		widgetDict.Write(strconv.AppendInt(nbuf[:0], int64(apID), 10))
		widgetDict.WriteString(" 0 R >>")
	}

	widgetDict.WriteString(" >>")

	objID := pageManager.AddExtraObject(widgetDict.String())
	pageManager.AddAnnotation(objID)
}
