package pdf

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/gin-gonic/gin"
)

// GenerateTemplatePDF generates a PDF document with multi-page support and embedded images
func GenerateTemplatePDF(c *gin.Context, template models.PDFTemplate) {
	var pdfBuffer bytes.Buffer
	xrefOffsets := make(map[int]int)

	// Get page dimensions from config
	pageConfig := template.Config
	pageDims := getPageDimensions(pageConfig.Page, pageConfig.PageAlignment)

	// Initialize page manager with Arlington compatibility flag
	pageManager := NewPageManager(pageDims, template.Config.ArlingtonCompatible)

	// Process images and create XObjects
	imageObjects := make(map[int]*ImageObject) // map imageIndex to ImageObject
	imageObjectIDs := make(map[int]int)        // map imageIndex to PDF object ID

	// Process cell images - map tableIdx:rowIdx:colIdx to XObject ID
	// Also process title table images with prefix "title:"
	cellImageObjects := make(map[string]*ImageObject)
	cellImageObjectIDs := make(map[string]int)

	nextImageObjectID := 1000 // Start image objects at ID 1000

	// Process standalone images
	for i, img := range template.Image {
		if img.ImageData != "" {
			imgObj, err := DecodeImageData(img.ImageData)
			if err == nil {
				imgObj.ObjectID = nextImageObjectID
				imageObjects[i] = imgObj
				imageObjectIDs[i] = nextImageObjectID
				nextImageObjectID++
			}
		}
	}

	// Process title table images
	if template.Title.Table != nil {
		for rowIdx, row := range template.Title.Table.Rows {
			for colIdx, cell := range row.Row {
				if cell.Image != nil && cell.Image.ImageData != "" {
					imgObj, err := DecodeImageData(cell.Image.ImageData)
					if err == nil {
						imgObj.ObjectID = nextImageObjectID
						cellKey := fmt.Sprintf("title:%d:%d", rowIdx, colIdx)
						cellImageObjects[cellKey] = imgObj
						cellImageObjectIDs[cellKey] = nextImageObjectID
						nextImageObjectID++
					}
				}
			}
		}
	}

	// Process cell images in tables
	for tableIdx, table := range template.Table {
		for rowIdx, row := range table.Rows {
			for colIdx, cell := range row.Row {
				if cell.Image != nil && cell.Image.ImageData != "" {
					imgObj, err := DecodeImageData(cell.Image.ImageData)
					if err == nil {
						imgObj.ObjectID = nextImageObjectID
						cellKey := fmt.Sprintf("%d:%d:%d", tableIdx, rowIdx, colIdx)
						cellImageObjects[cellKey] = imgObj
						cellImageObjectIDs[cellKey] = nextImageObjectID
						nextImageObjectID++
					}
				}
			}
		}
	}

	// PDF Header (PDF 2.0 for modern standards compliance)
	pdfBuffer.WriteString("%PDF-2.0\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	// Generate all content first to know how many pages we need
	// Pass imageObjects, imageObjectIDs and cellImageObjectIDs so content generation can reference them
	generateAllContentWithImages(template, pageManager, imageObjects, imageObjectIDs, cellImageObjectIDs)

	// Collect all widget IDs for AcroForm
	var allWidgetIDs []int
	for _, annots := range pageManager.PageAnnots {
		allWidgetIDs = append(allWidgetIDs, annots...)
	}

	// Object 1: Catalog with accessibility and compliance improvements
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R")
	// Add language tag for accessibility (PDF/UA requirement)
	pdfBuffer.WriteString(" /Lang (en-US)")
	// Add MarkInfo to indicate this is a tagged PDF (even if minimal)
	pdfBuffer.WriteString(" /MarkInfo << /Marked false >>")
	if len(allWidgetIDs) > 0 {
		// Create AcroForm object
		acroFormID := pageManager.NextObjectID
		pageManager.NextObjectID++

		var fieldsRef strings.Builder
		fieldsRef.WriteString("[")
		for _, id := range allWidgetIDs {
			fieldsRef.WriteString(fmt.Sprintf(" %d 0 R", id))
		}
		fieldsRef.WriteString("]")

		// Note: /NeedAppearances removed (deprecated in PDF 2.0) - widget appearances are generated programmatically
		acroFormContent := fmt.Sprintf("<< /Fields %s /DA (/Helv 0 Tf 0 g) >>", fieldsRef.String())
		pageManager.ExtraObjects[acroFormID] = acroFormContent

		pdfBuffer.WriteString(fmt.Sprintf(" /AcroForm %d 0 R", acroFormID))
	}
	pdfBuffer.WriteString(" >>\nendobj\n")

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

	// Font object layout depends on Arlington compatibility mode:
	// All 14 standard PDF fonts: F1-F4 (Helvetica), F5-F8 (Times), F9-F12 (Courier), F13 (Symbol), F14 (ZapfDingbats)
	// - Arlington mode: 14 font dicts + 14 font descriptors + widths arrays = more objects
	// - Simple mode: 14 font dicts only = 14 objects
	numFonts := 14
	var fontDescriptorStart, widthsArrayStart int
	if template.Config.ArlingtonCompatible {
		fontDescriptorStart = fontObjectStart + numFonts  // FontDescriptors start after font dicts
		widthsArrayStart = fontDescriptorStart + numFonts // Widths arrays start after descriptors
	}

	// Build XObject references for page resources (standalone images + cell images)
	// Using short names: /I0, /I1 for images, /C0_1_2 for cell images, /X0 for appearance streams
	xobjectRefs := ""
	if len(imageObjects) > 0 || len(cellImageObjects) > 0 {
		xobjectRefs = " /XObject <<"
		// Add standalone images with short names
		for i, objID := range imageObjectIDs {
			xobjectRefs += fmt.Sprintf(" /I%d %d 0 R", i, objID)
		}
		// Add cell images with short names
		for cellKey, objID := range cellImageObjectIDs {
			// Use short cellKey identifier (e.g., C0_1_2 instead of CellImg_0:1:2)
			shortKey := strings.ReplaceAll(cellKey, ":", "_")
			xobjectRefs += fmt.Sprintf(" /C%s %d 0 R", shortKey, objID)
		}
		// Add extra objects (appearance streams) that are XObjects with short names
		for id, content := range pageManager.ExtraObjects {
			if strings.Contains(content, "/Type /XObject") {
				xobjectRefs += fmt.Sprintf(" /X%d %d 0 R", id, id)
			}
		}
		xobjectRefs += " >>"
	} else {
		// Even if no images, we might have XObjects from form fields (appearance streams)
		hasXObjects := false
		for _, content := range pageManager.ExtraObjects {
			if strings.Contains(content, "/Type /XObject") {
				hasXObjects = true
				break
			}
		}
		if hasXObjects {
			xobjectRefs = " /XObject <<"
			for id, content := range pageManager.ExtraObjects {
				if strings.Contains(content, "/Type /XObject") {
					xobjectRefs += fmt.Sprintf(" /X%d %d 0 R", id, id)
				}
			}
			xobjectRefs += " >>"
		}
	}

	// Generate page objects
	for i, pageID := range pageManager.Pages {
		xrefOffsets[pageID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", pageID))

		// Add Annots if present
		annotsStr := ""
		if i < len(pageManager.PageAnnots) && len(pageManager.PageAnnots[i]) > 0 {
			annotsStr = " /Annots ["
			for _, annotID := range pageManager.PageAnnots[i] {
				annotsStr += fmt.Sprintf(" %d 0 R", annotID)
			}
			annotsStr += "]"
		}

		pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] ",
			pageDims.Width, pageDims.Height))
		pdfBuffer.WriteString(fmt.Sprintf("/Contents %d 0 R ", contentObjectStart+i))
		// Include all 14 standard PDF fonts: Helvetica (F1-F4), Times (F5-F8), Courier (F9-F12), Symbol (F13), ZapfDingbats (F14)
		pdfBuffer.WriteString(fmt.Sprintf("/Resources << /Font << /F1 %d 0 R /F2 %d 0 R /F3 %d 0 R /F4 %d 0 R /F5 %d 0 R /F6 %d 0 R /F7 %d 0 R /F8 %d 0 R /F9 %d 0 R /F10 %d 0 R /F11 %d 0 R /F12 %d 0 R /F13 %d 0 R /F14 %d 0 R >>%s >>%s >>\n",
			fontObjectStart, fontObjectStart+1, fontObjectStart+2, fontObjectStart+3,
			fontObjectStart+4, fontObjectStart+5, fontObjectStart+6, fontObjectStart+7,
			fontObjectStart+8, fontObjectStart+9, fontObjectStart+10, fontObjectStart+11,
			fontObjectStart+12, fontObjectStart+13, xobjectRefs, annotsStr))
		pdfBuffer.WriteString("endobj\n")
	}

	// Generate content stream objects with FlateDecode compression
	for i, contentStream := range pageManager.ContentStreams {
		objectID := contentObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))

		// Compress content stream with zlib (PDF FlateDecode expects zlib format, not raw deflate)
		var compressedBuf bytes.Buffer
		zlibWriter := zlib.NewWriter(&compressedBuf)
		zlibWriter.Write(contentStream.Bytes())
		zlibWriter.Close()
		compressedData := compressedBuf.Bytes()

		// Write stream - Length is exact byte count of compressed data
		pdfBuffer.WriteString(fmt.Sprintf("<< /Filter /FlateDecode /Length %d >>\nstream\n", len(compressedData)))
		pdfBuffer.Write(compressedData)
		pdfBuffer.WriteString("\nendstream\nendobj\n")
	}

	// Generate font objects - conditional based on Arlington compatibility
	// All 14 standard PDF Type 1 fonts
	fontNames := []string{
		"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique", // F1-F4
		"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic", // F5-F8
		"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique", // F9-F12
		"Symbol", "ZapfDingbats", // F13-F14
	}
	fontRefs := []string{"/F1", "/F2", "/F3", "/F4", "/F5", "/F6", "/F7", "/F8", "/F9", "/F10", "/F11", "/F12", "/F13", "/F14"}

	if template.Config.ArlingtonCompatible {
		// Arlington mode: Generate PDF 2.0 compliant font objects with full metrics
		// Generate widths arrays for each unique set of widths
		widthsObjIDs := make(map[string]int)
		currentWidthsID := widthsArrayStart

		// Pre-generate widths arrays (some fonts share the same widths)
		widthGroups := map[string]string{
			"Helvetica":             "helvetica-regular",
			"Helvetica-Oblique":     "helvetica-regular",
			"Helvetica-Bold":        "helvetica-bold",
			"Helvetica-BoldOblique": "helvetica-bold",
			"Times-Roman":           "times-roman",
			"Times-Bold":            "times-bold",
			"Times-Italic":          "times-italic",
			"Times-BoldItalic":      "times-bolditalic",
			"Courier":               "courier",
			"Courier-Bold":          "courier",
			"Courier-Oblique":       "courier",
			"Courier-BoldOblique":   "courier",
			"Symbol":                "symbol",
			"ZapfDingbats":          "zapfdingbats",
		}

		// Create unique widths arrays
		widthsGenerated := make(map[string]bool)
		for _, fontName := range fontNames {
			group := widthGroups[fontName]
			if !widthsGenerated[group] {
				widthsObjIDs[group] = currentWidthsID
				xrefOffsets[currentWidthsID] = pdfBuffer.Len()
				pdfBuffer.WriteString(GenerateWidthsArrayObject(fontName, currentWidthsID))
				currentWidthsID++
				widthsGenerated[group] = true
			}
		}

		// Generate font objects and descriptors
		for i, fontName := range fontNames {
			fontObjID := fontObjectStart + i
			fdObjID := fontDescriptorStart + i
			widthsObjID := widthsObjIDs[widthGroups[fontName]]

			// Generate Font dictionary
			xrefOffsets[fontObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateFontObject(fontName, fontObjID, fdObjID, widthsObjID))

			// Generate FontDescriptor
			xrefOffsets[fdObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateFontDescriptorObject(fontName, fdObjID))
		}
	} else {
		// Simple mode: Generate basic font objects without full metrics (smaller file size)
		for i, fontName := range fontNames {
			fontObjID := fontObjectStart + i
			xrefOffsets[fontObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateSimpleFontObject(fontName, fontRefs[i], fontObjID))
		}
	}

	// Generate image XObjects (standalone images)
	for _, imgObj := range imageObjects {
		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(CreateImageXObject(imgObj, imgObj.ObjectID))
	}

	// Generate image XObjects (cell images)
	for _, imgObj := range cellImageObjects {
		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(CreateImageXObject(imgObj, imgObj.ObjectID))
	}

	// Generate Extra Objects (Widgets, Appearance Streams, AcroForm)
	for id, content := range pageManager.ExtraObjects {
		xrefOffsets[id] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", id, content))
	}

	// Generate Info dictionary - keeping minimal for PDF 2.0
	// Note: Producer, Creator, Title are deprecated in PDF 2.0 but still widely used
	// For full compliance, these should be in XMP metadata stream instead
	infoObjectID := pageManager.NextObjectID
	pageManager.NextObjectID++
	xrefOffsets[infoObjectID] = pdfBuffer.Len()
	// Format date according to PDF spec: D:YYYYMMDDHHmmSSOHH'mm'
	creationDate := time.Now().Format("D:20060102150405-07'00'")
	pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", infoObjectID))
	// Minimal Info dict with just dates (dates are not deprecated)
	pdfBuffer.WriteString(fmt.Sprintf("<< /CreationDate (%s) /ModDate (%s) >>\n", creationDate, creationDate))
	pdfBuffer.WriteString("endobj\n")

	// Generate Document ID (two MD5 hashes - one based on content, one random)
	contentHash := md5.Sum(pdfBuffer.Bytes())
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	randomHash := md5.Sum(randomBytes)
	documentID := fmt.Sprintf("[<%s> <%s>]", hex.EncodeToString(contentHash[:]), hex.EncodeToString(randomHash[:]))

	// Build compact XRef table - collect all used object IDs and sort them
	usedObjects := make([]int, 0, len(xrefOffsets)+1)
	usedObjects = append(usedObjects, 0) // Object 0 is always the free list head
	for objID := range xrefOffsets {
		usedObjects = append(usedObjects, objID)
	}

	// Sort the used objects
	for i := 0; i < len(usedObjects)-1; i++ {
		for j := i + 1; j < len(usedObjects); j++ {
			if usedObjects[i] > usedObjects[j] {
				usedObjects[i], usedObjects[j] = usedObjects[j], usedObjects[i]
			}
		}
	}

	// Find max object ID for Size field
	maxObjID := 0
	for objID := range xrefOffsets {
		if objID > maxObjID {
			maxObjID = objID
		}
	}
	if infoObjectID > maxObjID {
		maxObjID = infoObjectID
	}
	totalObjects := maxObjID + 1

	// Write compact XRef table using subsections
	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString("xref\n")

	// Group consecutive objects into subsections
	var subsections []struct{ start, count int }
	i := 0
	for i < len(usedObjects) {
		start := usedObjects[i]
		count := 1
		for i+count < len(usedObjects) && usedObjects[i+count] == start+count {
			count++
		}
		subsections = append(subsections, struct{ start, count int }{start, count})
		i += count
	}

	for _, sub := range subsections {
		pdfBuffer.WriteString(fmt.Sprintf("%d %d\n", sub.start, sub.count))
		for j := 0; j < sub.count; j++ {
			objID := sub.start + j
			if objID == 0 {
				pdfBuffer.WriteString("0000000000 65535 f \n")
			} else if offset, exists := xrefOffsets[objID]; exists {
				pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
			}
		}
	}

	// Trailer with Info and ID
	pdfBuffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R /Info %d 0 R /ID %s >>\n", totalObjects, infoObjectID, documentID))
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	// HTTP Response
	filename := fmt.Sprintf("template-pdf-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}

// generateAllContentWithImages processes the template and generates content with image support
func generateAllContentWithImages(template models.PDFTemplate, pageManager *PageManager, imageObjects map[int]*ImageObject, imageObjectIDs map[int]int, cellImageObjectIDs map[string]int) {
	// Initialize first page
	initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)

	// Title - Process if title text is provided OR if title has a table
	if template.Title.Text != "" || template.Title.Table != nil {
		titleProps := parseProps(template.Title.Props)

		// Calculate title height based on content
		var titleHeight float64
		if template.Title.Table != nil && len(template.Title.Table.Rows) > 0 {
			// Estimate height from table rows
			for _, row := range template.Title.Table.Rows {
				rowH := 25.0
				for _, cell := range row.Row {
					if cell.Height != nil && *cell.Height > rowH {
						rowH = *cell.Height
					}
				}
				titleHeight += rowH
			}
		} else {
			titleHeight = float64(titleProps.FontSize) // Title only, no extra spacing
		}

		if pageManager.CheckPageBreak(titleHeight) {
			pageManager.AddNewPage()
			initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)
		}

		drawTitle(pageManager.GetCurrentContentStream(), template.Title, titleProps, pageManager, cellImageObjectIDs)
	}

	// Check if we have ordered elements array
	if len(template.Elements) > 0 {
		// Process elements in order
		tableIdx := 0
		for _, elem := range template.Elements {
			switch elem.Type {
			case "table":
				var table models.Table
				if elem.Table != nil {
					table = *elem.Table
				} else if elem.Index < len(template.Table) {
					table = template.Table[elem.Index]
				} else {
					continue
				}
				drawTable(table, tableIdx, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
				tableIdx++
			case "spacer":
				var spacer models.Spacer
				if elem.Spacer != nil {
					spacer = *elem.Spacer
				} else if elem.Index < len(template.Spacer) {
					spacer = template.Spacer[elem.Index]
				} else {
					continue
				}
				drawSpacer(spacer, pageManager)
			case "image":
				var image models.Image
				var imgIdx int
				if elem.Image != nil {
					image = *elem.Image
					imgIdx = -1 // No index for inline image
				} else if elem.Index < len(template.Image) {
					image = template.Image[elem.Index]
					imgIdx = elem.Index
				} else {
					continue
				}
				if imgIdx >= 0 {
					if imgObj, exists := imageObjects[imgIdx]; exists {
						imageXObjectRef := fmt.Sprintf("/I%d", imgIdx)
						drawImageWithXObjectInternal(image, imageXObjectRef, pageManager, template.Config.PageBorder, template.Config.Watermark, imgObj.Width, imgObj.Height)
					} else {
						drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
					}
				} else {
					drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
				}
			}
		}
	} else {
		// Legacy mode: process tables, then spacers, then images (spacers at end)
		// Tables - Process each table with automatic page breaks
		for tableIdx, table := range template.Table {
			drawTable(table, tableIdx, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
		}

		// Spacers - Process each spacer (added after tables in legacy mode)
		for _, spacer := range template.Spacer {
			drawSpacer(spacer, pageManager)
		}

		// Images - Process each image with automatic page breaks
		for i, image := range template.Image {
			if imgObj, exists := imageObjects[i]; exists {
				// Image was successfully decoded, draw it with XObject reference
				imageXObjectRef := fmt.Sprintf("/I%d", i)
				drawImageWithXObjectInternal(image, imageXObjectRef, pageManager, template.Config.PageBorder, template.Config.Watermark, imgObj.Width, imgObj.Height)
			} else {
				// Fall back to placeholder if image couldn't be decoded
				drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
			}
		}
	}

	// Draw footer and page numbers on every page (footer first to avoid overlap)
	totalPages := len(pageManager.Pages)
	for i := 0; i < totalPages; i++ {
		// Draw footer on this page if footer text provided
		if template.Footer.Text != "" {
			drawFooter(&pageManager.ContentStreams[i], template.Footer)
		}
		// Draw page number on this page
		drawPageNumber(&pageManager.ContentStreams[i], i+1, totalPages, pageManager.PageDimensions)
	}
}
