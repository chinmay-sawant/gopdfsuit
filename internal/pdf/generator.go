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

// GenerateTemplatePDF generates a PDF document with multi-page support and embedded images
func GenerateTemplatePDF(c *gin.Context, template models.PDFTemplate) {
	var pdfBuffer bytes.Buffer
	xrefOffsets := make(map[int]int)

	// Get page dimensions from config
	pageConfig := template.Config
	pageDims := getPageDimensions(pageConfig.Page, pageConfig.PageAlignment)

	// Initialize page manager
	pageManager := NewPageManager(pageDims)

	// Process images and create XObjects
	imageObjects := make(map[int]*ImageObject) // map imageIndex to ImageObject
	imageObjectIDs := make(map[int]int)        // map imageIndex to PDF object ID

	// Process cell images - map tableIdx:rowIdx:colIdx to XObject ID
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

	// PDF Header
	pdfBuffer.WriteString("%PDF-1.7\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	// Generate all content first to know how many pages we need
	// Pass imageObjects, imageObjectIDs and cellImageObjectIDs so content generation can reference them
	generateAllContentWithImages(template, pageManager, imageObjects, imageObjectIDs, cellImageObjectIDs)

	// Collect all widget IDs for AcroForm
	var allWidgetIDs []int
	for _, annots := range pageManager.PageAnnots {
		allWidgetIDs = append(allWidgetIDs, annots...)
	}

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R")
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

		acroFormContent := fmt.Sprintf("<< /Fields %s /NeedAppearances true /DA (/Helv 0 Tf 0 g) >>", fieldsRef.String())
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

	// Build XObject references for page resources (standalone images + cell images)
	xobjectRefs := ""
	if len(imageObjects) > 0 || len(cellImageObjects) > 0 {
		xobjectRefs = " /XObject <<"
		// Add standalone images
		for i, objID := range imageObjectIDs {
			xobjectRefs += fmt.Sprintf(" /Im%d %d 0 R", i, objID)
		}
		// Add cell images
		for cellKey, objID := range cellImageObjectIDs {
			// Use cellKey as unique identifier (e.g., CellImg_0_1_2)
			xobjectRefs += fmt.Sprintf(" /CellImg_%s %d 0 R", cellKey, objID)
		}
		// Add extra objects (appearance streams) that are XObjects
		for id, content := range pageManager.ExtraObjects {
			if strings.Contains(content, "/Type /XObject") {
				xobjectRefs += fmt.Sprintf(" /XObj%d %d 0 R", id, id)
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
					xobjectRefs += fmt.Sprintf(" /XObj%d %d 0 R", id, id)
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
		pdfBuffer.WriteString(fmt.Sprintf("/Resources << /Font << /F1 %d 0 R /F2 %d 0 R /F3 %d 0 R /F4 %d 0 R >>%s >>%s >>\n",
			fontObjectStart, fontObjectStart+1, fontObjectStart+2, fontObjectStart+3, xobjectRefs, annotsStr))
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

	// Cross-reference table
	totalObjects := fontObjectStart + 4
	if len(imageObjects) > 0 || len(cellImageObjects) > 0 {
		// Find max image object ID (both standalone and cell images)
		maxImgID := 0
		for _, objID := range imageObjectIDs {
			if objID > maxImgID {
				maxImgID = objID
			}
		}
		for _, objID := range cellImageObjectIDs {
			if objID > maxImgID {
				maxImgID = objID
			}
		}
		if maxImgID >= totalObjects {
			totalObjects = maxImgID + 1
		}
	}

	// Check extra objects for max ID
	if pageManager.NextObjectID > totalObjects {
		totalObjects = pageManager.NextObjectID
	}

	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", totalObjects))
	for i := 1; i < totalObjects; i++ {
		if offset, exists := xrefOffsets[i]; exists {
			pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
		} else {
			pdfBuffer.WriteString("0000000000 65535 f \n")
		}
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

// generateAllContentWithImages processes the template and generates content with image support
func generateAllContentWithImages(template models.PDFTemplate, pageManager *PageManager, imageObjects map[int]*ImageObject, imageObjectIDs map[int]int, cellImageObjectIDs map[string]int) {
	// Initialize first page
	initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)

	// Title - Only process if title text is provided
	if template.Title.Text != "" {
		titleProps := parseProps(template.Title.Props)
		titleHeight := float64(titleProps.FontSize + 50) // Title + spacing

		if pageManager.CheckPageBreak(titleHeight) {
			pageManager.AddNewPage()
			initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)
		}

		drawTitle(pageManager.GetCurrentContentStream(), template.Title, titleProps, pageManager)
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
						imageXObjectRef := fmt.Sprintf("/Im%d", imgIdx)
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
				imageXObjectRef := fmt.Sprintf("/Im%d", i)
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
