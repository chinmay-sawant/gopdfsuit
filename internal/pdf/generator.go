package pdf

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
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

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Generate all content first to know how many pages we need
	// Pass imageObjectIDs and cellImageObjectIDs so content generation can reference them
	generateAllContentWithImages(template, pageManager, imageObjectIDs, cellImageObjectIDs)

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
		xobjectRefs += " >>"
	}

	// Generate page objects
	for i, pageID := range pageManager.Pages {
		xrefOffsets[pageID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", pageID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] ",
			pageDims.Width, pageDims.Height))
		pdfBuffer.WriteString(fmt.Sprintf("/Contents %d 0 R ", contentObjectStart+i))
		pdfBuffer.WriteString(fmt.Sprintf("/Resources << /Font << /F1 %d 0 R /F2 %d 0 R /F3 %d 0 R /F4 %d 0 R >>%s >> >>\n",
			fontObjectStart, fontObjectStart+1, fontObjectStart+2, fontObjectStart+3, xobjectRefs))
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
func generateAllContentWithImages(template models.PDFTemplate, pageManager *PageManager, imageObjectIDs map[int]int, cellImageObjectIDs map[string]int) {
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
	for tableIdx, table := range template.Table {
		drawTable(table, tableIdx, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
	}

	// Images - Process each image with automatic page breaks
	for i, image := range template.Image {
		if _, exists := imageObjectIDs[i]; exists {
			// Image was successfully decoded, draw it with XObject reference
			imageXObjectRef := fmt.Sprintf("/Im%d", i)
			drawImageWithXObjectInternal(image, imageXObjectRef, pageManager, template.Config.PageBorder, template.Config.Watermark)
		} else {
			// Fall back to placeholder if image couldn't be decoded
			drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
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
