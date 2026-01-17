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

	// Reset font registry usage tracking for this PDF
	fontRegistry := GetFontRegistry()
	fontRegistry.ResetUsage()

	// PDF/A mode: Register Liberation fonts for all used standard fonts
	if template.Config.PDFACompliant {
		usedStandardFonts := collectAllStandardFontsInTemplate(template)
		usedFontsList := make([]string, 0, len(usedStandardFonts))
		for fontName := range usedStandardFonts {
			usedFontsList = append(usedFontsList, fontName)
		}

		pdfaManager := GetPDFAFontManager()
		if err := pdfaManager.RegisterLiberationFontsForPDFA(fontRegistry, usedFontsList); err != nil {
			fmt.Printf("Warning: Failed to load Liberation fonts for PDF/A: %v\n", err)
			// Continue without PDF/A font compliance
		}
	}

	// Load custom fonts from config
	for _, fontConfig := range template.Config.CustomFonts {
		if fontConfig.Name == "" {
			continue
		}
		if fontConfig.FontData != "" {
			// Load from base64 data
			if err := fontRegistry.RegisterFontFromBase64(fontConfig.Name, fontConfig.FontData); err != nil {
				// Log error but continue
				fmt.Printf("Warning: failed to load custom font %s from data: %v\n", fontConfig.Name, err)
			}
		} else if fontConfig.FilePath != "" {
			// Load from file path
			if err := fontRegistry.RegisterFontFromFile(fontConfig.Name, fontConfig.FilePath); err != nil {
				// Log error but continue
				fmt.Printf("Warning: failed to load custom font %s from file: %v\n", fontConfig.Name, err)
			}
		}
	}

	// Pre-scan template to mark all font usage for subsetting
	scanTemplateForFontUsage(template, fontRegistry)

	// Generate font subsets
	if err := fontRegistry.GenerateSubsets(); err != nil {
		fmt.Printf("Warning: failed to generate font subsets: %v\n", err)
	}

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

	// Process inline images in elements array
	// These are images specified directly in the elements array with type="image" and image:{...}
	elemImageObjects := make(map[int]*ImageObject) // map element index to ImageObject
	elemImageObjectIDs := make(map[int]int)        // map element index to PDF object ID
	for elemIdx, elem := range template.Elements {
		if elem.Type == "image" && elem.Image != nil && elem.Image.ImageData != "" {
			imgObj, err := DecodeImageData(elem.Image.ImageData)
			if err == nil {
				imgObj.ObjectID = nextImageObjectID
				elemImageObjects[elemIdx] = imgObj
				elemImageObjectIDs[elemIdx] = nextImageObjectID
				nextImageObjectID++
			}
		}
	}

	pdfBuffer.WriteString("%PDF-2.0\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	// CRITICAL: Assign object IDs to custom fonts BEFORE generating content
	// This ensures that when content streams are generated, custom font references
	// (e.g., /CF2000) have valid object IDs. Custom fonts start AFTER image objects
	// to avoid conflicts with image XObjects.
	customFontObjectStart := nextImageObjectID
	fontRegistry.AssignObjectIDs(customFontObjectStart)

	// Generate all content first to know how many pages we need
	// Pass imageObjects, imageObjectIDs, cellImageObjectIDs and elemImageObjectIDs so content generation can reference them
	generateAllContentWithImages(template, pageManager, imageObjects, imageObjectIDs, cellImageObjectIDs, elemImageObjects, elemImageObjectIDs)

	// Build document outlines (bookmarks) if provided
	// Check both top-level Bookmarks and Config.Bookmarks (top-level takes precedence)
	outlineBuilder := NewOutlineBuilder(pageManager)
	bookmarksToUse := template.Bookmarks
	if len(bookmarksToUse) == 0 {
		bookmarksToUse = template.Config.Bookmarks
	}
	outlineObjID := outlineBuilder.BuildOutlines(bookmarksToUse)

	// Get named destinations for internal links
	namesObjID, hasNames := outlineBuilder.GetNamedDestinations()

	// Setup PDF/A handler if enabled
	var pdfaHandler *PDFAHandler
	if template.Config.PDFA != nil && template.Config.PDFA.Enabled {
		pdfaHandler = NewPDFAHandler(template.Config.PDFA, pageManager)
	}

	// Setup digital signature if enabled
	var pdfSigner *PDFSigner
	var sigIDs *SignatureIDs
	if template.Config.Signature != nil && template.Config.Signature.Enabled {
		var err error
		pdfSigner, err = NewPDFSigner(template.Config.Signature)
		if err == nil && pdfSigner != nil {
			sigIDs = pdfSigner.CreateSignatureField(pageManager, pageDims)
		}
	}

	// Collect all widget IDs for AcroForm (filter out link annotations)
	var allWidgetIDs []int
	for _, annots := range pageManager.PageAnnots {
		for _, annotID := range annots {
			// Check if this is a widget annotation (not a link)
			if content, exists := pageManager.ExtraObjects[annotID]; exists {
				if strings.Contains(content, "/Subtype /Widget") {
					allWidgetIDs = append(allWidgetIDs, annotID)
				}
			}
		}
	}

	// Reserve object IDs for PDF/A compliance objects (will be written at the end)
	// These need to be referenced in the Catalog
	metadataObjectID := pageManager.NextObjectID
	pageManager.NextObjectID++

	// Only reserve ICC profile and OutputIntent IDs for PDF/A mode
	var iccProfileObjectID, outputIntentObjectID int
	if template.Config.PDFACompliant {
		iccProfileObjectID = pageManager.NextObjectID
		pageManager.NextObjectID++
		outputIntentObjectID = pageManager.NextObjectID
		pageManager.NextObjectID++
	}

	// Calculate total pages for bookmarks
	totalPages := len(pageManager.Pages)

	// Bookmarks are generated using outlineBuilder earlier (lines 168-171)
	// outlineRootID := pageManager.GenerateBookmarks(template.Bookmarks, xrefOffsets, &pdfBuffer)

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R")
	// Add language tag for accessibility (PDF/UA requirement)
	pdfBuffer.WriteString(" /Lang (en-US)")
	// Add PDF/A specific entries or default MarkInfo
	if pdfaHandler != nil {
		// PDF/A entries will be added after we generate metadata/outputintent objects
		// For now, skip MarkInfo as it will be added via pdfaHandler
	} else {
		// Add MarkInfo to indicate this is a tagged PDF (even if minimal)
		pdfBuffer.WriteString(" /MarkInfo << /Marked false >>")
	}

	// Add Metadata reference (always include for document info)
	pdfBuffer.WriteString(fmt.Sprintf(" /Metadata %d 0 R", metadataObjectID))

	// Add OutputIntents ONLY for PDF/A compliance (required for color space validation)
	if template.Config.PDFACompliant {
		pdfBuffer.WriteString(fmt.Sprintf(" /OutputIntents [%d 0 R]", outputIntentObjectID))
	}

	// Add outlines (bookmarks) if present
	if outlineObjID > 0 {
		pdfBuffer.WriteString(fmt.Sprintf(" /Outlines %d 0 R", outlineObjID))
		pdfBuffer.WriteString(" /PageMode /UseOutlines") // Show bookmark panel by default
	}
	// Add named destinations if present
	if hasNames {
		pdfBuffer.WriteString(fmt.Sprintf(" /Names %d 0 R", namesObjID))
	}
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

		// Get appropriate font reference for AcroForm DA (handles PDF/A mode)
		widgetFontRef := getWidgetFontReference()

		// Build AcroForm content - include SigFlags if signatures are present
		var acroFormContent string
		if sigIDs != nil {
			// SigFlags 3 = SignaturesExist (1) + AppendOnly (2)
			acroFormContent = fmt.Sprintf("<< /Fields %s /DA (%s 0 Tf 0 g) /SigFlags %d >>", fieldsRef.String(), widgetFontRef, GetAcroFormSigFlags())
		} else {
			// Note: /NeedAppearances removed (deprecated in PDF 2.0) - widget appearances are generated programmatically
			acroFormContent = fmt.Sprintf("<< /Fields %s /DA (%s 0 Tf 0 g) >>", fieldsRef.String(), widgetFontRef)
		}
		pageManager.ExtraObjects[acroFormID] = acroFormContent

		pdfBuffer.WriteString(fmt.Sprintf(" /AcroForm %d 0 R", acroFormID))
	}

	// Store position where we'll need to inject PDF/A references
	// For now we close the catalog and will rebuild it if needed
	catalogEndPlaceholder := ""
	if pdfaHandler != nil {
		// Reserve space for metadata and outputintent references
		// These will be set after we generate those objects
		catalogEndPlaceholder = " /Metadata %METADATA% /OutputIntents [%OUTPUTINTENT%]"
		pdfBuffer.WriteString(catalogEndPlaceholder)
	}
	pdfBuffer.WriteString(" >>\nendobj\n")

	// Object 2: Pages (will be updated after we know total page count)
	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n")
	pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>\n",
		formatPageKids(pageManager.Pages), len(pageManager.Pages)))
	pdfBuffer.WriteString("endobj\n")

	// Calculate object IDs
	// totalPages is already calculated above
	contentObjectStart := totalPages + 3               // Content objects start after pages
	fontObjectStart := contentObjectStart + totalPages // Fonts start after content

	// Standard fonts definition
	fontNames := []string{
		"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique", // F1-F4
		"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic", // F5-F8
		"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique", // F9-F12
		"Symbol", "ZapfDingbats", // F13-F14
	}
	fontRefs := []string{"/F1", "/F2", "/F3", "/F4", "/F5", "/F6", "/F7", "/F8", "/F9", "/F10", "/F11", "/F12", "/F13", "/F14"}

	// Identify used standard fonts
	usedStandardFonts := collectUsedStandardFonts(template)
	shouldEmbed := template.Config.EmbedFonts == nil || *template.Config.EmbedFonts

	// Calculate Object IDs for standard fonts dynamically
	// Only assign IDs for fonts that are used
	fontObjectIDs := make(map[string]int)     // Font Name -> Font Dictionary Object ID
	fontDescriptorIDs := make(map[string]int) // Font Name -> Font Descriptor Object ID
	fontWidthsIDs := make(map[string]int)     // Font Name (group) -> Widths Array Object ID

	currentObjectID := fontObjectStart

	// Phase 1: Assign IDs for Font Dictionaries
	for _, name := range fontNames {
		if usedStandardFonts[name] {
			fontObjectIDs[name] = currentObjectID
			currentObjectID++
		}
	}

	// Phase 2: Assign IDs for Descriptors and Widths (Arlington mode only)
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

	if template.Config.ArlingtonCompatible && shouldEmbed {
		// Assign Descriptor IDs
		for _, name := range fontNames {
			if usedStandardFonts[name] {
				fontDescriptorIDs[name] = currentObjectID
				currentObjectID++
			}
		}

		// Assign Widths IDs (deduplicated by group)
		assignedGroups := make(map[string]bool)
		for _, name := range fontNames {
			if usedStandardFonts[name] {
				group := widthGroups[name]
				if !assignedGroups[group] {
					fontWidthsIDs[group] = currentObjectID
					currentObjectID++
					assignedGroups[group] = true
				}
			}
		}
	}

	// Assign object IDs to custom fonts (object IDs already assigned before content generation)
	// customFontObjectStart is already calculated, no need to assign again
	// Just build the custom font resource references
	customFontRefs := fontRegistry.GeneratePDFFontResources()

	// Build XObject references for page resources (standalone images + cell images + element images)
	// Using short names: /I0, /I1 for images, /C0_1_2 for cell images, /E0 for element images, /X0 for appearance streams
	xobjectRefs := ""
	if len(imageObjects) > 0 || len(cellImageObjects) > 0 || len(elemImageObjects) > 0 {
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
		// Add element images with /E prefix
		for elemIdx, objID := range elemImageObjectIDs {
			xobjectRefs += fmt.Sprintf(" /E%d %d 0 R", elemIdx, objID)
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
	// Build ColorSpace resources for PDF/A mode
	// Using DefaultRGB tells Adobe Acrobat that DeviceRGB colors are already in sRGB
	// This prevents the double color conversion that makes colors appear pale
	colorSpaceRefs := ""
	if template.Config.PDFACompliant && iccProfileObjectID > 0 {
		colorSpaceRefs = fmt.Sprintf(" /ColorSpace << /DefaultRGB [/ICCBased %d 0 R] >>", iccProfileObjectID)
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

		// Build standard font resources string dynamically
		var stdFontRefs strings.Builder
		for i, name := range fontNames {
			if id, ok := fontObjectIDs[name]; ok {
				stdFontRefs.WriteString(fmt.Sprintf(" %s %d 0 R", fontRefs[i], id))
			}
		}

		// Include ColorSpace resource for PDF/A mode
		pdfBuffer.WriteString(fmt.Sprintf("/Resources <<%s /Font <<%s%s >>%s >>%s >>\n",
			colorSpaceRefs, stdFontRefs.String(), customFontRefs, xobjectRefs, annotsStr))
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

	// Generate font objects (Standard Fonts)
	if template.Config.ArlingtonCompatible && shouldEmbed {
		// Arlington mode: Generate PDF 2.0 compliant font objects with full metrics

		// 1. Generate Widths Arrays (shared)
		generatedGroupWidths := make(map[string]bool)
		for _, name := range fontNames {
			if !usedStandardFonts[name] {
				continue
			}

			group := widthGroups[name]
			if !generatedGroupWidths[group] {
				widthsID := fontWidthsIDs[group]
				xrefOffsets[widthsID] = pdfBuffer.Len()
				pdfBuffer.WriteString(GenerateWidthsArrayObject(name, widthsID))
				generatedGroupWidths[group] = true
			}
		}

		// 2. Generate Font Dictionaries and Descriptors
		for _, name := range fontNames {
			if !usedStandardFonts[name] {
				continue
			}

			fontObjID := fontObjectIDs[name]
			fdObjID := fontDescriptorIDs[name]
			widthsObjID := fontWidthsIDs[widthGroups[name]]

			// Generate Font Dictionary
			xrefOffsets[fontObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateFontObject(name, fontObjID, fdObjID, widthsObjID))

			// Generate Font Descriptor
			xrefOffsets[fdObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateFontDescriptorObject(name, fdObjID))
		}
	} else {
		// Simple mode: Generate basic font objects without full metrics
		for i, name := range fontNames {
			if !usedStandardFonts[name] {
				continue
			}

			fontObjID := fontObjectIDs[name]
			xrefOffsets[fontObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateSimpleFontObject(name, fontRefs[i], fontObjID))
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

	// Generate image XObjects (element images)
	for _, imgObj := range elemImageObjects {
		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(CreateImageXObject(imgObj, imgObj.ObjectID))
	}

	// Generate custom font objects (TrueType/OpenType embedded fonts)
	usedFonts := fontRegistry.GetUsedFonts()
	for _, font := range usedFonts {
		fontObjects := GenerateTrueTypeFontObjects(font)
		for objID, content := range fontObjects {
			xrefOffsets[objID] = pdfBuffer.Len()
			pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objID, content))
		}
	}

	// Generate Extra Objects (Widgets, Appearance Streams, AcroForm)
	for id, content := range pageManager.ExtraObjects {
		xrefOffsets[id] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", id, content))
	}

	// Generate PDF/A metadata objects if enabled
	if pdfaHandler != nil {
		// Generate XMP metadata
		docIDForXMP := fmt.Sprintf("%x", time.Now().UnixNano())
		metadataObjID, metadataContent := pdfaHandler.GenerateXMPMetadata(docIDForXMP)
		xrefOffsets[metadataObjID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", metadataObjID, metadataContent))

		// Generate OutputIntent with ICC profile
		outputIntentObjID, outputIntentObjs := pdfaHandler.GenerateOutputIntent()
		iccObjID := pdfaHandler.GetICCProfileObjID()

		// Write ICC profile object (with stream)
		if len(outputIntentObjs) > 0 {
			xrefOffsets[iccObjID] = pdfBuffer.Len()
			// ICC profile needs raw data appended
			iccData := getSRGBICCProfile()
			pdfBuffer.WriteString(outputIntentObjs[0])
			pdfBuffer.Write(iccData)
			pdfBuffer.WriteString("\nendstream\nendobj\n")
		}

		// Write OutputIntent object
		if len(outputIntentObjs) > 1 {
			xrefOffsets[outputIntentObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(outputIntentObjs[1])
			pdfBuffer.WriteString("\n")
		}

		// Update catalog with PDF/A references
		// We need to rebuild the catalog object
		catalogContent := pdfBuffer.String()

		// Find and replace the placeholder in catalog
		if strings.Contains(catalogContent, "%METADATA%") {
			catalogContent = strings.Replace(catalogContent, "%METADATA%", fmt.Sprintf("%d 0 R", metadataObjID), 1)
			catalogContent = strings.Replace(catalogContent, "%OUTPUTINTENT%", fmt.Sprintf("%d 0 R", outputIntentObjID), 1)
			pdfBuffer.Reset()
			pdfBuffer.WriteString(catalogContent)
		}
	}

	// Generate Info dictionary - keeping minimal for PDF 2.0
	// Note: Producer, Creator, Title are deprecated in PDF 2.0 but still widely used
	// For full compliance, these should be in XMP metadata stream instead
	infoObjectID := pageManager.NextObjectID
	pageManager.NextObjectID++
	// Format date according to PDF spec: D:YYYYMMDDHHmmSSOHH'mm'
	// Go's time format doesn't support the PDF timezone format directly, so we build it manually
	now := time.Now()
	_, tzOffset := now.Zone()
	tzSign := "+"
	if tzOffset < 0 {
		tzSign = "-"
		tzOffset = -tzOffset
	}
	tzHours := tzOffset / 3600
	tzMinutes := (tzOffset % 3600) / 60
	creationDate := fmt.Sprintf("D:%s%s%02d'%02d'", now.Format("20060102150405"), tzSign, tzHours, tzMinutes)

	// For PDF/A-4: Skip Info object entirely (Clause 6.1.3, Test 4)
	// Info key shall not be present in trailer unless PieceInfo exists in catalog
	// All metadata should be in XMP stream instead
	if !template.Config.PDFACompliant {
		xrefOffsets[infoObjectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", infoObjectID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /CreationDate (%s) /ModDate (%s) >>\n", creationDate, creationDate))
		pdfBuffer.WriteString("endobj\n")
	}

	// Setup encryption if security config is provided
	var encryption *PDFEncryption
	var encryptObjID int
	if template.Config.Security != nil && template.Config.Security.OwnerPassword != "" {
		// Generate document ID first (needed for encryption)
		docID := GenerateDocumentID(pdfBuffer.Bytes())

		var err error
		encryption, err = NewPDFEncryption(template.Config.Security, docID)
		if err == nil {
			// Create encryption dictionary object
			encryptObjID = pageManager.NextObjectID
			pageManager.NextObjectID++
			xrefOffsets[encryptObjID] = pdfBuffer.Len()
			pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", encryptObjID, encryption.GetEncryptDictionary(encryptObjID)))
		}
	}

	// Generate Document ID (two MD5 hashes - one based on content, one random)
	// Generate Document ID (two MD5 hashes - one based on content, one random)
	contentHash := md5.Sum(pdfBuffer.Bytes())
	var documentID string
	if encryption != nil {
		documentID = FormatDocumentID(encryption.DocumentID)
	} else {
		randomBytes := make([]byte, 16)
		rand.Read(randomBytes)
		randomHash := md5.Sum(randomBytes)
		documentID = fmt.Sprintf("[<%s> <%s>]", hex.EncodeToString(contentHash[:]), hex.EncodeToString(randomHash[:]))
	}

	// Generate PDF/A-1b compliance objects
	// Always generate metadata (for document info)
	xrefOffsets[metadataObjectID] = pdfBuffer.Len()
	pdfBuffer.WriteString(GenerateXMPMetadataObject(metadataObjectID, hex.EncodeToString(contentHash[:]), creationDate))

	// Only generate ICC profile and OutputIntent for PDF/A mode
	// This is the key fix: without these, Adobe Acrobat won't apply color management
	// and colors will appear as intended (same as in Chrome/browser)
	if template.Config.PDFACompliant {
		xrefOffsets[iccProfileObjectID] = pdfBuffer.Len()
		pdfBuffer.Write(GenerateICCProfileObject(iccProfileObjectID))

		xrefOffsets[outputIntentObjectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(GenerateOutputIntentObject(outputIntentObjectID, iccProfileObjectID))
	}

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
	// For PDF/A-4, The Info key shall not be present in the trailer dictionary unless there exists a PieceInfo entry
	trailerExtra := ""
	if encryptObjID > 0 {
		trailerExtra = fmt.Sprintf(" /Encrypt %d 0 R", encryptObjID)
	}

	if template.Config.PDFACompliant {
		pdfBuffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R /ID %s%s >>\n", totalObjects, documentID, trailerExtra))
	} else {
		pdfBuffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R /Info %d 0 R /ID %s%s >>\n", totalObjects, infoObjectID, documentID, trailerExtra))
	}
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	// Apply digital signature if configured
	finalPDF := pdfBuffer.Bytes()
	if pdfSigner != nil && sigIDs != nil {
		signedPDF, err := UpdatePDFWithSignature(finalPDF, pdfSigner)
		if err == nil {
			finalPDF = signedPDF
		}
		// If signing fails, we still return the unsigned PDF
	}

	// HTTP Response
	filename := fmt.Sprintf("template-pdf-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", finalPDF)
}

// generateAllContentWithImages processes the template and generates content with image support
func generateAllContentWithImages(template models.PDFTemplate, pageManager *PageManager, imageObjects map[int]*ImageObject, imageObjectIDs map[int]int, cellImageObjectIDs map[string]int, elemImageObjects map[int]*ImageObject, elemImageObjectIDs map[int]int) {
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
		for elemIdx, elem := range template.Elements {
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
				if elem.Image != nil {
					// Inline image in elements array - use element index for XObject lookup
					image = *elem.Image
					if imgObj, exists := elemImageObjects[elemIdx]; exists {
						// Use element image XObject with /E prefix to distinguish from /I prefix
						imageXObjectRef := fmt.Sprintf("/E%d", elemIdx)
						drawImageWithXObjectInternal(image, imageXObjectRef, pageManager, template.Config.PageBorder, template.Config.Watermark, imgObj.Width, imgObj.Height)
					} else {
						// Fall back to placeholder if no XObject
						drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
					}
				} else if elem.Index < len(template.Image) {
					// Reference to template.Image array
					image = template.Image[elem.Index]
					if imgObj, exists := imageObjects[elem.Index]; exists {
						imageXObjectRef := fmt.Sprintf("/I%d", elem.Index)
						drawImageWithXObjectInternal(image, imageXObjectRef, pageManager, template.Config.PageBorder, template.Config.Watermark, imgObj.Width, imgObj.Height)
					} else {
						drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
					}
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
		// Set CurrentPageIndex so that any annotations added (e.g. by drawFooter) go to the correct page
		pageManager.CurrentPageIndex = i

		// Draw footer on this page if footer text provided
		if template.Footer.Text != "" {
			drawFooter(&pageManager.ContentStreams[i], template.Footer, pageManager)
		}
		// Draw page number on this page
		drawPageNumber(&pageManager.ContentStreams[i], i+1, totalPages, pageManager.PageDimensions)
	}
}

// scanTemplateForFontUsage scans all text in template and marks font usage for subsetting
func scanTemplateForFontUsage(template models.PDFTemplate, registry *CustomFontRegistry) {
	// Scan title
	if template.Title.Text != "" {
		props := parseProps(template.Title.Props)
		markFontUsage(props, template.Title.Text)
	}

	// Scan title table if present
	if template.Title.Table != nil {
		for _, row := range template.Title.Table.Rows {
			for _, cell := range row.Row {
				if cell.Text != "" {
					props := parseProps(cell.Props)
					markFontUsage(props, cell.Text)
				}
			}
		}
	}

	// Scan tables
	for _, table := range template.Table {
		for _, row := range table.Rows {
			for _, cell := range row.Row {
				if cell.Text != "" {
					props := parseProps(cell.Props)
					markFontUsage(props, cell.Text)
				}
			}
		}
	}

	// Scan elements (ordered)
	for _, elem := range template.Elements {
		if elem.Type == "table" {
			if elem.Table != nil {
				for _, row := range elem.Table.Rows {
					for _, cell := range row.Row {
						if cell.Text != "" {
							props := parseProps(cell.Props)
							markFontUsage(props, cell.Text)
						}
					}
				}
			} else if elem.Index >= 0 && elem.Index < len(template.Table) {
				for _, row := range template.Table[elem.Index].Rows {
					for _, cell := range row.Row {
						if cell.Text != "" {
							props := parseProps(cell.Props)
							markFontUsage(props, cell.Text)
						}
					}
				}
			}
		}
	}

	// Scan footer
	if template.Footer.Text != "" {
		props := parseProps(template.Footer.Font)
		markFontUsage(props, template.Footer.Text)
	}

	// Scan Watermark (uses Helvetica)
	if template.Config.Watermark != "" {
		markFontUsage(models.Props{FontName: "Helvetica"}, template.Config.Watermark)
	}

	// Scan Page Numbers (uses Helvetica)
	markFontUsage(models.Props{FontName: "Helvetica"}, "Page of 0123456789")

	// Scan Image Names (uses Helvetica)
	// Standalone images
	for _, img := range template.Image {
		if img.ImageName != "" {
			markFontUsage(models.Props{FontName: "Helvetica"}, img.ImageName)
		}
	}
}

// collectUsedStandardFonts returns a set of standard font names used in the template
// Always includes Helvetica as it's the default font and used for form fields
// Excludes fonts that are registered as custom fonts (e.g., Liberation fonts in PDF/A mode)
func collectUsedStandardFonts(template models.PDFTemplate) map[string]bool {
	used := make(map[string]bool)
	registry := GetFontRegistry()

	// Helper to mark font only if it's a true standard font (not overridden by custom)
	markFont := func(propsStr string) {
		props := parseProps(propsStr)
		// Only mark as standard font if it's not registered as a custom font
		if !IsCustomFont(props.FontName) && !registry.HasFont(props.FontName) {
			used[props.FontName] = true
		}
	}

	// Helvetica is default font - only add if not overridden by custom font
	if !registry.HasFont("Helvetica") {
		used["Helvetica"] = true // Default font, always required for AcroForm default appearance
	}

	// Scan title
	if template.Title.Text != "" {
		markFont(template.Title.Props)
	}

	// Scan title table
	if template.Title.Table != nil {
		for _, row := range template.Title.Table.Rows {
			for _, cell := range row.Row {
				if cell.Text != "" {
					markFont(cell.Props)
				}
			}
		}
	}

	// Scan tables
	for _, table := range template.Table {
		for _, row := range table.Rows {
			for _, cell := range row.Row {
				if cell.Text != "" {
					markFont(cell.Props)
				}
			}
		}
	}

	// Scan elements
	for _, elem := range template.Elements {
		if elem.Type == "table" {
			if elem.Table != nil {
				for _, row := range elem.Table.Rows {
					for _, cell := range row.Row {
						if cell.Text != "" {
							markFont(cell.Props)
						}
					}
				}
			} else if elem.Index >= 0 && elem.Index < len(template.Table) {
				for _, row := range template.Table[elem.Index].Rows {
					for _, cell := range row.Row {
						if cell.Text != "" {
							markFont(cell.Props)
						}
					}
				}
			}
		}
	}

	// Scan footer
	if template.Footer.Text != "" {
		markFont(template.Footer.Font)
	}

	// Scan watermark (always uses Helvetica if present)
	if template.Config.Watermark != "" {
		// Watermark uses Helvetica
		if !IsCustomFont("Helvetica") && !registry.HasFont("Helvetica") {
			used["Helvetica"] = true
		}
	}

	// Page numbers always use Helvetica
	if !IsCustomFont("Helvetica") && !registry.HasFont("Helvetica") {
		used["Helvetica"] = true
	}

	// Image placeholder names use Helvetica
	hasImages := len(template.Image) > 0
	for _, table := range template.Table {
		for _, row := range table.Rows {
			for _, cell := range row.Row {
				if cell.Image != nil {
					hasImages = true
					break
				}
			}
			if hasImages {
				break
			}
		}
		if hasImages {
			break
		}
	}
	// Check title table for images
	if !hasImages && template.Title.Table != nil {
		for _, row := range template.Title.Table.Rows {
			for _, cell := range row.Row {
				if cell.Image != nil {
					hasImages = true
					break
				}
			}
			if hasImages {
				break
			}
		}
	}
	if hasImages {
		// Image placeholders use Helvetica for displaying image names
		if !IsCustomFont("Helvetica") && !registry.HasFont("Helvetica") {
			used["Helvetica"] = true
		}
	}

	return used
}

// collectAllStandardFontsInTemplate returns all standard font names used in the template
// This does NOT check the font registry - used for determining which Liberation fonts to load
func collectAllStandardFontsInTemplate(template models.PDFTemplate) map[string]bool {
	used := make(map[string]bool)

	// Helper to mark font
	markFont := func(propsStr string) {
		props := parseProps(propsStr)
		if !IsCustomFont(props.FontName) {
			used[props.FontName] = true
		}
	}

	used["Helvetica"] = true // Default font, always required

	// Scan title
	if template.Title.Text != "" {
		markFont(template.Title.Props)
	}

	// Scan title table
	if template.Title.Table != nil {
		for _, row := range template.Title.Table.Rows {
			for _, cell := range row.Row {
				if cell.Text != "" {
					markFont(cell.Props)
				}
			}
		}
	}

	// Scan tables
	for _, table := range template.Table {
		for _, row := range table.Rows {
			for _, cell := range row.Row {
				if cell.Text != "" {
					markFont(cell.Props)
				}
			}
		}
	}

	// Scan elements
	for _, elem := range template.Elements {
		if elem.Type == "table" {
			if elem.Table != nil {
				for _, row := range elem.Table.Rows {
					for _, cell := range row.Row {
						if cell.Text != "" {
							markFont(cell.Props)
						}
					}
				}
			} else if elem.Index >= 0 && elem.Index < len(template.Table) {
				for _, row := range template.Table[elem.Index].Rows {
					for _, cell := range row.Row {
						if cell.Text != "" {
							markFont(cell.Props)
						}
					}
				}
			}
		}
	}

	// Scan footer
	if template.Footer.Text != "" {
		markFont(template.Footer.Font)
	}

	return used
}
