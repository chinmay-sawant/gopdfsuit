package pdf

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/encryption"
)

// pdfBufferPool reuses bytes.Buffer across PDF generations to reduce GC pressure.
var pdfBufferPool = sync.Pool{
	New: func() interface{} {
		buf := new(bytes.Buffer)
		buf.Grow(64 * 1024) // 64KB initial capacity
		return buf
	},
}

// scratchBufPool reuses the small scratch buffer for strconv.Append* operations.
var scratchBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 128)
		return &buf
	},
}

// GenerateTemplatePDF generates a PDF document with multi-page support and embedded images.
// Returns the PDF bytes and any error encountered during generation.
func GenerateTemplatePDF(template models.PDFTemplate) ([]byte, error) {
	pdfBufferPtr := pdfBufferPool.Get().(*bytes.Buffer)
	pdfBufferPtr.Reset()
	defer pdfBufferPool.Put(pdfBufferPtr)
	pdfBuffer := pdfBufferPtr // use directly as *bytes.Buffer

	scratchPtr := scratchBufPool.Get().(*[]byte)
	b := (*scratchPtr)[:0]
	defer func() { *scratchPtr = b[:0]; scratchBufPool.Put(scratchPtr) }()

	xrefOffsets := make(map[int]int)

	// Get page dimensions from config
	pageConfig := template.Config
	pageDims := getPageDimensions(pageConfig.Page, pageConfig.PageAlignment)
	pageMargins := ParsePageMargins(pageConfig.PageMargin)

	// Create a local clone of the font registry for this PDF generation session
	// This ensures thread safety by isolating usage tracking (UsedChars) per generation
	globalRegistry := GetFontRegistry()
	fontRegistry := globalRegistry.CloneForGeneration()

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

	// Initialize page manager with Arlington compatibility flag and per-generation font registry
	pageManager := NewPageManager(pageDims, pageMargins, template.Config.ArlingtonCompatible, fontRegistry)

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
						cellKey := buildCellKey2("title", rowIdx, colIdx)
						cellImageObjects[cellKey] = imgObj
						cellImageObjectIDs[cellKey] = nextImageObjectID
						nextImageObjectID++
					}
				}
			}
		}
	}

	// Process cell images in tables (indexed tables from top-level Table array)
	for tableIdx, table := range template.Table {
		for rowIdx, row := range table.Rows {
			for colIdx, cell := range row.Row {
				if cell.Image != nil && cell.Image.ImageData != "" {
					imgObj, err := DecodeImageData(cell.Image.ImageData)
					if err == nil {
						imgObj.ObjectID = nextImageObjectID
						// Key for indexed tables: "0:0:0" (tableIdx:rowIdx:colIdx)
						cellKey := buildCellKey3(tableIdx, rowIdx, colIdx)
						cellImageObjects[cellKey] = imgObj
						cellImageObjectIDs[cellKey] = nextImageObjectID
						nextImageObjectID++
					}
				}
			}
		}
	}

	// Process inline table images in Elements array
	for elemIdx, elem := range template.Elements {
		if elem.Type == "table" && elem.Table != nil {
			for rowIdx, row := range elem.Table.Rows {
				for colIdx, cell := range row.Row {
					if cell.Image != nil && cell.Image.ImageData != "" {
						imgObj, err := DecodeImageData(cell.Image.ImageData)
						if err == nil {
							imgObj.ObjectID = nextImageObjectID
							// Key for inline tables: "elem_inline:5:0:0" (elem_inline:elemIdx:rowIdx:colIdx)
							cellKey := buildCellKeyElemInline(elemIdx, rowIdx, colIdx)
							cellImageObjects[cellKey] = imgObj
							cellImageObjectIDs[cellKey] = nextImageObjectID
							nextImageObjectID++
						}
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
	generateAllContentWithImages(template, pageManager, imageObjects, cellImageObjectIDs, elemImageObjects)

	// Generate font subsets MOVED to after signature generation to ensure signature chars are included

	// Setup encryption EARLY if security config is provided (before writing content)
	// This is needed because content streams need to be encrypted
	var enc *encryption.PDFEncryption
	if template.Config.Security != nil && template.Config.Security.Enabled && template.Config.Security.OwnerPassword != "" {
		// Generate a preliminary document ID for encryption setup
		preliminaryID := encryption.GenerateDocumentID([]byte(template.Title.Text + fmt.Sprintf("%d", len(pageManager.Pages))))
		var err error
		enc, err = encryption.NewPDFEncryption(template.Config.Security, preliminaryID)
		if err != nil {
			enc = nil // Fall back to no encryption on error
		}
	}

	var encryptor ObjectEncryptor
	if enc != nil {
		encryptor = enc
	}

	// Build document outlines (bookmarks) if provided
	// Check both top-level Bookmarks and Config.Bookmarks (top-level takes precedence)
	outlineBuilder := NewOutlineBuilder(pageManager, encryptor)
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
		// If main PdfTitle is set but PDFA Title isn't, copy it over
		if template.Config.PdfTitle != "" && template.Config.PDFA.Title == "" {
			template.Config.PDFA.Title = template.Config.PdfTitle
		}
		pdfaHandler = NewPDFAHandler(template.Config.PDFA, pageManager, encryptor)
	} else if template.Config.PDFACompliant {
		// If using valid PDF/A mode but no explicit PDFA config, create one to ensure metadata
		pdfaConfig := &models.PDFAConfig{
			Enabled:     true,
			Conformance: "4", // Default for PDF/A-4
			Title:       template.Config.PdfTitle,
		}
		pdfaHandler = NewPDFAHandler(pdfaConfig, pageManager, encryptor)
	}

	// Calculate object IDs for fonts early (needed for signature font embedding)
	// Calculate total pages first
	totalPages := len(pageManager.Pages)
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
	usedStandardFonts := collectUsedStandardFonts(template, fontRegistry)

	// If signature is enabled and visible, force usage of Helvetica
	signatureEnabled := template.Config.Signature != nil && template.Config.Signature.Enabled
	if signatureEnabled && template.Config.Signature.Visible {
		usedStandardFonts["Helvetica"] = true
	}

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

	// Setup digital signature if enabled
	var pdfSigner *PDFSigner
	var sigIDs *SignatureIDs
	if signatureEnabled {
		var err error
		pdfSigner, err = NewPDFSigner(template.Config.Signature)
		if err == nil && pdfSigner != nil {
			// Get the font ID for signature appearance
			// In PDF/A mode, this returns the Liberation font ID that replaces Helvetica
			// In standard mode, this returns the standard Helvetica font ID
			signatureFontID := getWidgetFontObjectID(pageManager.FontRegistry)
			if signatureFontID == 0 {
				// Fallback to standard font object ID if no custom font
				signatureFontID = fontObjectIDs["Helvetica"]
			}
			sigIDs = pdfSigner.CreateSignatureField(pageManager, pageDims, signatureFontID)
		}
	}

	// Generate font subsets after content generation AND signature creation
	// This ensures characters used in signature appearance are included in the subset
	if err := fontRegistry.GenerateSubsets(); err != nil {
		fmt.Printf("Warning: failed to generate font subsets: %v\n", err)
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
	// These need to be referenced in the Catalog
	metadataObjectID := pageManager.NextObjectID
	pageManager.NextObjectID++

	// Reserve StructTreeRoot ID for PDF/UA
	structTreeRootID := pageManager.NextObjectID
	pageManager.NextObjectID++

	// Only reserve ICC profile and OutputIntent IDs for PDF/A mode
	var iccProfileObjectID, outputIntentObjectID, grayICCProfileObjID int
	if template.Config.PDFACompliant {
		iccProfileObjectID = pageManager.NextObjectID
		pageManager.NextObjectID++
		outputIntentObjectID = pageManager.NextObjectID
		pageManager.NextObjectID++
		// Reserve Gray ICC profile object ID for DeviceGray color space
		grayICCProfileObjID = pageManager.NextObjectID
		pageManager.NextObjectID++
	}

	// Calculate total pages for bookmarks
	// totalPages is already calculated above

	// Bookmarks are generated using outlineBuilder earlier (lines 168-171)
	// outlineRootID := pageManager.GenerateBookmarks(template.Bookmarks, xrefOffsets, &pdfBuffer)

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R")
	// Add language tag for accessibility (PDF/UA requirement)
	pdfBuffer.WriteString(" /Lang (en-US)")
	// Add PDF/A specific entries or default MarkInfo
	// Add MarkInfo to indicate this is a tagged PDF (Required for PDF/UA)
	pdfBuffer.WriteString(" /MarkInfo << /Marked true >>")

	// Add ViewerPreferences (Required for PDF/UA)
	pdfBuffer.WriteString(" /ViewerPreferences << /DisplayDocTitle true >>")

	// Note: Metadata and OutputIntents are only added when pdfaHandler != nil (PDF/A mode)
	// because that's when we actually create those objects (see lines ~730-745)

	// Add outlines (bookmarks) if present
	if outlineObjID > 0 {
		pdfBuffer.WriteString(" /Outlines ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(outlineObjID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R")
		pdfBuffer.WriteString(" /PageMode /UseOutlines") // Show bookmark panel by default
	}
	// Add named destinations if present
	if hasNames {
		pdfBuffer.WriteString(" /Names ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(namesObjID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R")
	}
	if len(allWidgetIDs) > 0 {
		// Create AcroForm object
		acroFormID := pageManager.NextObjectID
		pageManager.NextObjectID++

		var fieldsRef strings.Builder
		fieldsRef.WriteString("[")
		for _, id := range allWidgetIDs {
			fieldsRef.WriteByte(' ')
			var widBuf [12]byte
			fieldsRef.Write(strconv.AppendInt(widBuf[:0], int64(id), 10))
			fieldsRef.WriteString(" 0 R")
		}
		fieldsRef.WriteString("]")

		// Get appropriate font reference for AcroForm DA (handles PDF/A mode)
		widgetFontRef := getWidgetFontReference(fontRegistry)

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

		pdfBuffer.WriteString(" /AcroForm ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(acroFormID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R")
	}

	// Store position where we'll need to inject PDF/A references
	// For now we close the catalog and will rebuild it if needed
	// Add StructTreeRoot reference (required for PDF/UA)
	pdfBuffer.WriteString(" /StructTreeRoot ")
	b = b[:0]
	b = strconv.AppendInt(b, int64(structTreeRootID), 10)
	pdfBuffer.Write(b)
	pdfBuffer.WriteString(" 0 R")

	// For PDF/A, add Metadata and OutputIntent references using pre-reserved object IDs
	// Note: We use the pre-reserved IDs directly here to avoid placeholder replacement
	// which would invalidate all xref offsets after the Catalog
	// PDF/UA-2 requires XMP metadata for ALL PDFs (ISO 14289-2:2024, Clause 8.11.1)
	pdfBuffer.WriteString(" /Metadata ")
	b = b[:0]
	b = strconv.AppendInt(b, int64(metadataObjectID), 10)
	pdfBuffer.Write(b)
	pdfBuffer.WriteString(" 0 R")
	if pdfaHandler != nil {
		pdfBuffer.WriteString(" /OutputIntents [")
		b = b[:0]
		b = strconv.AppendInt(b, int64(outputIntentObjectID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R]")
	}
	pdfBuffer.WriteString(" >>\nendobj\n")

	// Object 2: Pages (will be updated after we know total page count)
	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Pages /Kids [")
	pdfBuffer.WriteString(formatPageKids(pageManager.Pages))
	pdfBuffer.WriteString("] /Count ")
	b = b[:0]
	b = strconv.AppendInt(b, int64(len(pageManager.Pages)), 10)
	pdfBuffer.Write(b)
	pdfBuffer.WriteString(" >>\n")
	pdfBuffer.WriteString("endobj\n")

	// NOTE: Font ID calculation has been moved up to before signature generation
	// variables fontObjectIDs, fontDescriptorIDs, fontWidthsIDs, usedStandardFonts are already populated

	// Assign object IDs to custom fonts (object IDs already assigned before content generation)
	// customFontObjectStart is already calculated, no need to assign again
	// Just build the custom font resource references
	customFontRefs := fontRegistry.GeneratePDFFontResources()

	// Build XObject references for page resources (standalone images + cell images + element images)
	// Using short names: /I0, /I1 for images, /C0_1_2 for cell images, /E0 for element images, /X0 for appearance streams
	xobjectRefs := ""
	{
		var xobjBuf [20]byte // scratch for strconv.AppendInt
		var xobjBuilder strings.Builder

		writeXObjRef := func(prefix string, key string, objID int) {
			xobjBuilder.WriteString(" /")
			xobjBuilder.WriteString(prefix)
			xobjBuilder.WriteString(key)
			xobjBuilder.WriteByte(' ')
			xobjBuilder.Write(strconv.AppendInt(xobjBuf[:0], int64(objID), 10))
			xobjBuilder.WriteString(" 0 R")
		}
		writeXObjRefInt := func(prefix string, idx, objID int) {
			xobjBuilder.WriteString(" /")
			xobjBuilder.WriteString(prefix)
			xobjBuilder.Write(strconv.AppendInt(xobjBuf[:0], int64(idx), 10))
			xobjBuilder.WriteByte(' ')
			xobjBuilder.Write(strconv.AppendInt(xobjBuf[:0], int64(objID), 10))
			xobjBuilder.WriteString(" 0 R")
		}

		if len(imageObjects) > 0 || len(cellImageObjects) > 0 || len(elemImageObjects) > 0 {
			xobjBuilder.WriteString(" /XObject <<")
			for i, objID := range imageObjectIDs {
				writeXObjRefInt("I", i, objID)
			}
			for cellKey, objID := range cellImageObjectIDs {
				shortKey := strings.ReplaceAll(cellKey, ":", "_")
				writeXObjRef("C", shortKey, objID)
			}
			for elemIdx, objID := range elemImageObjectIDs {
				writeXObjRefInt("E", elemIdx, objID)
			}
			for id, content := range pageManager.ExtraObjects {
				if strings.Contains(content, "/Type /XObject") {
					writeXObjRefInt("X", id, id)
				}
			}
			xobjBuilder.WriteString(" >>")
		} else {
			hasXObjects := false
			for _, content := range pageManager.ExtraObjects {
				if strings.Contains(content, "/Type /XObject") {
					hasXObjects = true
					break
				}
			}
			if hasXObjects {
				xobjBuilder.WriteString(" /XObject <<")
				for id, content := range pageManager.ExtraObjects {
					if strings.Contains(content, "/Type /XObject") {
						writeXObjRefInt("X", id, id)
					}
				}
				xobjBuilder.WriteString(" >>")
			}
		}
		xobjectRefs = xobjBuilder.String()
	}
	// Build ColorSpace resources for PDF/A mode
	// Using DefaultRGB tells Adobe Acrobat that DeviceRGB colors are already in sRGB
	// This prevents the double color conversion that makes colors appear pale
	// IMPORTANT: Use pdfaHandler's ICC profile ID when available, as it creates its own objects
	// PDF/A-4 requires both DefaultRGB and DefaultGray ICC-based color spaces
	colorSpaceRefs := ""
	if template.Config.PDFACompliant {
		actualICCObjID := iccProfileObjectID
		if pdfaHandler != nil {
			actualICCObjID = pdfaHandler.GetICCProfileObjID()
		}
		if actualICCObjID > 0 {
			// Include both DefaultRGB and DefaultGray for full PDF/A-4 compliance
			var csBuf [20]byte
			var csBuilder strings.Builder
			csBuilder.WriteString(" /ColorSpace << /DefaultRGB [/ICCBased ")
			csBuilder.Write(strconv.AppendInt(csBuf[:0], int64(actualICCObjID), 10))
			csBuilder.WriteString(" 0 R] /DefaultGray [/ICCBased ")
			csBuilder.Write(strconv.AppendInt(csBuf[:0], int64(grayICCProfileObjID), 10))
			csBuilder.WriteString(" 0 R] >>")
			colorSpaceRefs = csBuilder.String()
		}
	}

	// Generate page objects
	for i, pageID := range pageManager.Pages {
		xrefOffsets[pageID] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(pageID), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)

		// Add Annots if present
		annotsStr := ""
		if i < len(pageManager.PageAnnots) && len(pageManager.PageAnnots[i]) > 0 {
			var annotBuf []byte
			annotBuf = append(annotBuf, " /Annots ["...)
			for _, annotID := range pageManager.PageAnnots[i] {
				annotBuf = append(annotBuf, ' ')
				annotBuf = strconv.AppendInt(annotBuf, int64(annotID), 10)
				annotBuf = append(annotBuf, " 0 R"...)
			}
			annotBuf = append(annotBuf, ']')
			annotsStr = string(annotBuf)
		}

		pdfBuffer.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ")
		pdfBuffer.WriteString(fmtNum(pageDims.Width))
		pdfBuffer.WriteByte(' ')
		pdfBuffer.WriteString(fmtNum(pageDims.Height))
		pdfBuffer.WriteString("] ")

		pdfBuffer.WriteString("/Contents ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(contentObjectStart+i), 10)
		b = append(b, " 0 R "...)
		pdfBuffer.Write(b)

		// Build standard font resources string dynamically
		var stdFontRefs strings.Builder
		for i, name := range fontNames {
			if id, ok := fontObjectIDs[name]; ok {
				stdFontRefs.WriteByte(' ')
				stdFontRefs.WriteString(fontRefs[i])
				stdFontRefs.WriteByte(' ')
				var fontBuf [12]byte
				stdFontRefs.Write(strconv.AppendInt(fontBuf[:0], int64(id), 10))
				stdFontRefs.WriteString(" 0 R")
			}
		}

		// Include ColorSpace resource for PDF/A mode
		// PDF/UA: Add StructParents entry ONLY if page has marked content
		structParentsEntry := ""
		if pageManager.Structure.NextMCID[i] > 0 {
			var spBuf [20]byte
			sp := append(spBuf[:0], " /StructParents "...)
			sp = strconv.AppendInt(sp, int64(i), 10)
			structParentsEntry = string(sp)
		}

		// PDF/UA-2: Add /Tabs S for pages with annotations (required by ISO 14289-2 8.9.3.3)
		tabsEntry := ""
		if len(pageManager.PageAnnots[i]) > 0 {
			tabsEntry = " /Tabs /S" // S = Structure order
		}

		pdfBuffer.WriteString("/Resources << ")
		pdfBuffer.WriteString(colorSpaceRefs)
		pdfBuffer.WriteString(" /Font <<")
		pdfBuffer.WriteString(stdFontRefs.String())
		pdfBuffer.WriteString(customFontRefs)
		pdfBuffer.WriteString(" >>")
		pdfBuffer.WriteString(xobjectRefs)
		pdfBuffer.WriteString(" >>")
		pdfBuffer.WriteString(annotsStr)
		pdfBuffer.WriteString(structParentsEntry)
		pdfBuffer.WriteString(tabsEntry)
		pdfBuffer.WriteString(" >>\n")
		pdfBuffer.WriteString("endobj\n")
	}

	// Generate content stream objects with FlateDecode compression
	for i, contentStream := range pageManager.ContentStreams {
		objectID := contentObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(objectID), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)

		// Compress content stream with pooled zlib writer (avoids allocation overhead)
		compressedBuf := getCompressBuffer()
		zlibWriter := getZlibWriter(compressedBuf)
		if _, err := zlibWriter.Write(contentStream.Bytes()); err != nil {
			_ = zlibWriter.Close()
			putZlibWriter(zlibWriter)
			compressBufPool.Put(compressedBuf)
			continue // Skip encryption if compression fails
		}
		_ = zlibWriter.Close()
		putZlibWriter(zlibWriter)
		compressedData := compressedBuf.Bytes()

		// Encrypt content stream if encryption is enabled
		if enc != nil {
			encryptedData := enc.EncryptStream(compressedData, objectID, 0)
			pdfBuffer.WriteString("<< /Filter /FlateDecode /Length ")
			b = b[:0]
			b = strconv.AppendInt(b, int64(len(encryptedData)), 10)
			pdfBuffer.Write(b)
			pdfBuffer.WriteString(" >>\nstream\n")
			pdfBuffer.Write(encryptedData)
		} else {
			// Write stream without encryption
			pdfBuffer.WriteString("<< /Filter /FlateDecode /Length ")
			b = b[:0]
			b = strconv.AppendInt(b, int64(len(compressedData)), 10)
			pdfBuffer.Write(b)
			pdfBuffer.WriteString(" >>\nstream\n")
			pdfBuffer.Write(compressedData)
		}
		compressBufPool.Put(compressedBuf)
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

	// Get the correct ICC profile object ID for images (if PDF/A compliance is enabled)
	// pdfaHandler creates its own ICC profile object IDs, so we must use those when available
	actualICCProfileObjID := iccProfileObjectID
	if pdfaHandler != nil {
		actualICCProfileObjID = pdfaHandler.GetICCProfileObjID()
	}

	// Generate image XObjects (standalone images)
	for _, imgObj := range imageObjects {
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if template.Config.PDFACompliant && actualICCProfileObjID > 0 {
			imgObj.ColorSpace = fmt.Sprintf("[/ICCBased %d 0 R]", actualICCProfileObjID)
		}

		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		if enc != nil {
			pdfBuffer.WriteString(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.WriteString(CreateImageXObject(imgObj, imgObj.ObjectID))
		}
	}

	// Generate image XObjects (cell images)
	for _, imgObj := range cellImageObjects {
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if template.Config.PDFACompliant && actualICCProfileObjID > 0 {
			imgObj.ColorSpace = fmt.Sprintf("[/ICCBased %d 0 R]", actualICCProfileObjID)
		}

		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		if enc != nil {
			pdfBuffer.WriteString(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.WriteString(CreateImageXObject(imgObj, imgObj.ObjectID))
		}
	}

	// Generate image XObjects (element images)
	for _, imgObj := range elemImageObjects {
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if template.Config.PDFACompliant && actualICCProfileObjID > 0 {
			imgObj.ColorSpace = fmt.Sprintf("[/ICCBased %d 0 R]", actualICCProfileObjID)
		}

		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		if enc != nil {
			pdfBuffer.WriteString(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.WriteString(CreateImageXObject(imgObj, imgObj.ObjectID))
		}
	}

	// Generate custom font objects (TrueType/OpenType embedded fonts)
	usedFonts := fontRegistry.GetUsedFonts()
	for _, font := range usedFonts {
		fontObjects := GenerateTrueTypeFontObjects(font, encryptor)
		for objID, content := range fontObjects {
			xrefOffsets[objID] = pdfBuffer.Len()
			b = b[:0]
			b = strconv.AppendInt(b, int64(objID), 10)
			b = append(b, " 0 obj\n"...)
			pdfBuffer.Write(b)
			pdfBuffer.WriteString(content)
			pdfBuffer.WriteString("\nendobj\n")
		}
	}

	// Generate Extra Objects (Widgets, Appearance Streams, AcroForm)
	for id, content := range pageManager.ExtraObjects {
		xrefOffsets[id] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(id), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(content)
		pdfBuffer.WriteString("\nendobj\n")
	}

	// Generate PDF/A metadata objects if enabled
	if pdfaHandler != nil {
		// Generate XMP metadata content (but use our pre-reserved metadataObjectID for consistency with Catalog)
		docIDForXMP := fmt.Sprintf("%x", time.Now().UnixNano())
		_, metadataContent := pdfaHandler.GenerateXMPMetadata(docIDForXMP)
		// Write metadata object using the pre-reserved ID that's already in the Catalog
		xrefOffsets[metadataObjectID] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(metadataObjectID), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(metadataContent)
		pdfBuffer.WriteString("\nendobj\n")

		// Generate OutputIntent with ICC profile
		// Use pre-reserved IDs to ensure consistency with Catalog/References
		_, outputIntentObjs, compressedICCData := pdfaHandler.GenerateOutputIntent(iccProfileObjectID, outputIntentObjectID)

		// Write ICC profile object (with stream)
		if len(outputIntentObjs) > 0 {
			xrefOffsets[iccProfileObjectID] = pdfBuffer.Len()
			// Write ICC profile dictionary header and compressed data
			pdfBuffer.WriteString(outputIntentObjs[0])
			pdfBuffer.Write(compressedICCData)
			pdfBuffer.WriteString("\nendstream\nendobj\n")
		}

		// Write Gray ICC profile object for DeviceGray color space compliance
		if grayICCProfileObjID > 0 {
			xrefOffsets[grayICCProfileObjID] = pdfBuffer.Len()
			pdfBuffer.Write(GenerateGrayICCProfileObject(grayICCProfileObjID, encryptor))
		}

		// Write OutputIntent object
		if len(outputIntentObjs) > 1 {
			xrefOffsets[outputIntentObjectID] = pdfBuffer.Len()
			pdfBuffer.WriteString(outputIntentObjs[1])
			pdfBuffer.WriteString("\n")
		}
	}
	// Note: For non-PDF/A, metadata is generated later (see pdfaHandler == nil block around line 804)

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
		b = b[:0]
		b = strconv.AppendInt(b, int64(infoObjectID), 10)
		b = append(b, " 0 obj\n"...)
		// Include Title in Info dictionary if provided
		titleEntry := ""
		if template.Config.PdfTitle != "" {
			titleEntry = fmt.Sprintf(" /Title (%s)", escapeText(template.Config.PdfTitle))
		}
		b = append(b, fmt.Sprintf("<< /CreationDate (%s) /ModDate (%s)%s >>\nendobj\n", creationDate, creationDate, titleEntry)...)
		pdfBuffer.Write(b)
	}

	// Write encryption dictionary object if encryption was set up
	var encryptObjID int
	if enc != nil {
		encryptObjID = pageManager.NextObjectID
		pageManager.NextObjectID++
		xrefOffsets[encryptObjID] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(encryptObjID), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(enc.GetEncryptDictionary(encryptObjID))
		pdfBuffer.WriteString("\nendobj\n")
	}

	// Generate Document ID (two MD5 hashes - one based on content, one random)
	// Use incremental hash to avoid copying the entire buffer
	contentHasher := md5.New()
	contentHasher.Write(pdfBuffer.Bytes())
	var contentHashArr [md5.Size]byte
	copy(contentHashArr[:], contentHasher.Sum(nil))
	var documentID string
	if enc != nil {
		documentID = encryption.FormatDocumentID(enc.DocumentID)
	} else {
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			// Fallback to time-based randomness if rand fails
			binary.BigEndian.PutUint64(randomBytes, uint64(time.Now().UnixNano()))
		}
		randomHash := md5.Sum(randomBytes)
		documentID = fmt.Sprintf("[<%s> <%s>]", hex.EncodeToString(contentHashArr[:]), hex.EncodeToString(randomHash[:]))
	}

	// Generate PDF/A-1b compliance objects
	// Always generate metadata (for document info)
	// Only generate if not already generated by pdfaHandler
	if pdfaHandler == nil {
		xrefOffsets[metadataObjectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(GenerateXMPMetadataObject(metadataObjectID, hex.EncodeToString(contentHashArr[:]), creationDate, encryptor))

		// Only generate ICC profile and OutputIntent for PDF/A mode
		// This is the key fix: without these, Adobe Acrobat won't apply color management
		// and colors will appear as intended (same as in Chrome/browser)
		if template.Config.PDFACompliant {
			xrefOffsets[iccProfileObjectID] = pdfBuffer.Len()
			pdfBuffer.Write(GenerateICCProfileObject(iccProfileObjectID, encryptor))

			// Write Gray ICC profile object for DeviceGray color space compliance
			if grayICCProfileObjID > 0 {
				xrefOffsets[grayICCProfileObjID] = pdfBuffer.Len()
				pdfBuffer.Write(GenerateGrayICCProfileObject(grayICCProfileObjID, encryptor))
			}

			xrefOffsets[outputIntentObjectID] = pdfBuffer.Len()
			pdfBuffer.WriteString(GenerateOutputIntentObject(outputIntentObjectID, iccProfileObjectID, encryptor))
		}
	}

	// Generate Structure Tree Objects (PDF/UA)
	// 1. Assign Object IDs to all elements
	// structTreeRootID is already reserved

	// Recursively assign IDs to all children

	// Recursively assign IDs to all children
	var assignStructIDs func(elem *StructElem)
	assignStructIDs = func(elem *StructElem) {
		// Only assign ID if not already assigned (e.g. by bookmarks or links logic)
		if elem.ObjectID == 0 {
			elem.ObjectID = pageManager.NextObjectID
			pageManager.NextObjectID++
		}
		for _, kid := range elem.Kids {
			if structElem, ok := kid.(*StructElem); ok {
				assignStructIDs(structElem)
			}
		}
	}
	// Start from root's children (Root itself has structTreeRootID, its children are elements)
	for _, kid := range pageManager.Structure.Root.Kids {
		if structElem, ok := kid.(*StructElem); ok {
			assignStructIDs(structElem)
		}
	}

	// 2. Generate ParentTree
	parentTreeID := pageManager.NextObjectID
	pageManager.NextObjectID++

	// 3. Generate PDF 2.0 Namespace for PDF/UA-2
	namespaceID := pageManager.NextObjectID
	pageManager.NextObjectID++

	xrefOffsets[namespaceID] = pdfBuffer.Len()
	b = b[:0]
	b = strconv.AppendInt(b, int64(namespaceID), 10)
	b = append(b, " 0 obj\n<< /Type /Namespace /NS (http://iso.org/pdf2/ssn) >>\nendobj\n"...)
	pdfBuffer.Write(b)

	xrefOffsets[structTreeRootID] = pdfBuffer.Len()
	b = b[:0]
	b = strconv.AppendInt(b, int64(structTreeRootID), 10)
	b = append(b, " 0 obj\n"...)
	pdfBuffer.Write(b)
	pdfBuffer.WriteString(pageManager.Structure.GenerateStructTreeRoot(structTreeRootID, parentTreeID, namespaceID))
	pdfBuffer.WriteString("\nendobj\n")

	// Write ParentTree
	xrefOffsets[parentTreeID] = pdfBuffer.Len()

	// Build ParentTree Nums map
	// Maps StructParents key (page index) to Array of IndirectRefs to StructElems
	var ptBuilder strings.Builder
	ptBuilder.WriteString(strconv.Itoa(parentTreeID))
	ptBuilder.WriteString(" 0 obj\n<< /Nums [")

	// Iterate through all pages that have marked content
	// We iterate by page index to keep Nums sorted
	maxPageIndex := len(pageManager.Pages)
	for i := 0; i < maxPageIndex; i++ {
		if elems, exists := pageManager.Structure.ParentTree[i]; exists && len(elems) > 0 {
			ptBuilder.WriteString(fmt.Sprintf(" %d [", i)) // Key is page index
			for _, elem := range elems {
				ptBuilder.WriteString(fmt.Sprintf(" %d 0 R", elem.ObjectID))
			}
			ptBuilder.WriteString(" ]")
		}
	}

	// PDF/UA-2: Add ParentTree entries for annotation StructParents
	// Each annotation's StructParent value maps to its Link structure element
	for _, annotInfo := range pageManager.AnnotStructElems {
		if linkElem, exists := pageManager.Structure.LinkElements[annotInfo.AnnotObjID]; exists {
			ptBuilder.WriteString(fmt.Sprintf(" %d %d 0 R", annotInfo.StructParentIdx, linkElem.ObjectID))
		}
	}

	ptBuilder.WriteString(" ] >>\nendobj\n")
	pdfBuffer.WriteString(ptBuilder.String())

	// Write all Structure Elements
	var writeStructElems func(elem *StructElem)
	writeStructElems = func(elem *StructElem) {
		xrefOffsets[elem.ObjectID] = pdfBuffer.Len()

		var sb strings.Builder
		sb.WriteString(strconv.Itoa(elem.ObjectID))
		sb.WriteString(" 0 obj\n<< /Type /StructElem /S /")
		sb.WriteString(string(elem.Type))

		// PDF/UA-2: Document element must be in PDF 2.0 namespace
		if elem.Type == StructDocument {
			sb.WriteString(fmt.Sprintf(" /NS %d 0 R", namespaceID))
		}

		if elem.Parent == pageManager.Structure.Root {
			sb.WriteString(fmt.Sprintf(" /P %d 0 R", structTreeRootID))
		} else if elem.Parent != nil {
			sb.WriteString(fmt.Sprintf(" /P %d 0 R", elem.Parent.ObjectID))
		}

		if elem.Title != "" {
			sb.WriteString(fmt.Sprintf(" /T (%s)", escapeText(elem.Title)))
		}
		if elem.Alt != "" {
			sb.WriteString(fmt.Sprintf(" /Alt (%s)", escapeText(elem.Alt)))
		}

		// Kids
		if len(elem.Kids) > 0 || elem.Type == StructLink {
			sb.WriteString(" /K [")

			// PDF/UA-2: For Link elements, output OBJR pointing to annotation
			if elem.Type == StructLink {
				// Find the annotation object ID for this Link element
				for annotObjID, linkElem := range pageManager.Structure.LinkElements {
					if linkElem == elem {
						// OBJR = Object Reference to the annotation
						// Format: << /Type /OBJR /Obj annotRef /Pg pageRef >>
						pageObjID := 3 // Default to first page
						if elem.PageID >= 0 && elem.PageID < len(pageManager.Pages) {
							pageObjID = pageManager.Pages[elem.PageID]
						}
						sb.WriteString(fmt.Sprintf(" << /Type /OBJR /Obj %d 0 R /Pg %d 0 R >>", annotObjID, pageObjID))
						break
					}
				}
			}

			for _, k := range elem.Kids {
				if kidElem, ok := k.(*StructElem); ok {
					sb.WriteString(fmt.Sprintf(" %d 0 R", kidElem.ObjectID))
				} else if mcid, ok := k.(int); ok {
					sb.WriteString(fmt.Sprintf(" %d", mcid))
				}
			}
			sb.WriteString(" ]")
		}

		// Pg entry (Page containing this element - required if not inherited)
		// We use elem.PageID + initial page offset (3) logic?
		// No, PageID in StructElem is index. We need absolute Object ID.
		// pm.Pages[elem.PageID] gives the object ID.
		if elem.PageID >= 0 && elem.PageID < len(pageManager.Pages) {
			pageObjID := pageManager.Pages[elem.PageID]
			sb.WriteString(fmt.Sprintf(" /Pg %d 0 R", pageObjID))
		}

		sb.WriteString(" >>\nendobj\n")
		pdfBuffer.WriteString(sb.String())

		// Recurse
		for _, k := range elem.Kids {
			if kidElem, ok := k.(*StructElem); ok {
				writeStructElems(kidElem)
			}
		}
	}

	for _, kid := range pageManager.Structure.Root.Kids {
		if structElem, ok := kid.(*StructElem); ok {
			writeStructElems(structElem)
		}
	}

	// Buffer replacement logic removed as we use reserved ID now

	// Build compact XRef table - collect all used object IDs and sort them
	usedObjects := make([]int, 0, len(xrefOffsets)+1)
	usedObjects = append(usedObjects, 0) // Object 0 is always the free list head
	for objID := range xrefOffsets {
		usedObjects = append(usedObjects, objID)
	}

	// Sort the used objects
	slices.Sort(usedObjects)

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
		b = b[:0]
		b = strconv.AppendInt(b, int64(sub.start), 10)
		b = append(b, ' ')
		b = strconv.AppendInt(b, int64(sub.count), 10)
		b = append(b, '\n')
		pdfBuffer.Write(b)
		for j := 0; j < sub.count; j++ {
			objID := sub.start + j
			if objID == 0 {
				pdfBuffer.WriteString("0000000000 65535 f \n")
			} else if offset, exists := xrefOffsets[objID]; exists {
				// Manual zero-padded 10-digit integer (replaces fmt.Sprintf("%010d 00000 n \n", offset))
				b = b[:0]
				b = strconv.AppendInt(b, int64(offset), 10)
				// Pad to 10 digits
				padding := 10 - len(b)
				if padding > 0 {
					// Shift existing digits right
					b = b[:10]
					copy(b[padding:], b[:10-padding])
					for k := 0; k < padding; k++ {
						b[k] = '0'
					}
				}
				b = append(b, " 00000 n \n"...)
				pdfBuffer.Write(b)
			}
		}
	}

	// Trailer with Info and ID
	// For PDF/A-4, The Info key shall not be present in the trailer dictionary unless there exists a PieceInfo entry
	trailerExtra := ""
	if encryptObjID > 0 {
		trailerExtra = fmt.Sprintf(" /Encrypt %d 0 R", encryptObjID)
	}

	pdfBuffer.WriteString("trailer\n<< /Size ")
	b = b[:0]
	b = strconv.AppendInt(b, int64(totalObjects), 10)
	pdfBuffer.Write(b)
	pdfBuffer.WriteString(" /Root 1 0 R ")
	if !template.Config.PDFACompliant {
		pdfBuffer.WriteString("/Info ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(infoObjectID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R ")
	}
	pdfBuffer.WriteString("/ID ")
	pdfBuffer.WriteString(documentID)
	if trailerExtra != "" {
		pdfBuffer.WriteString(trailerExtra)
	}
	pdfBuffer.WriteString(" >>\n")
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	// Apply digital signature if configured
	// Copy bytes before returning since pdfBuffer goes back to pool
	finalPDF := make([]byte, pdfBuffer.Len())
	copy(finalPDF, pdfBuffer.Bytes())
	if pdfSigner != nil && sigIDs != nil {
		signedPDF, err := UpdatePDFWithSignature(finalPDF, pdfSigner)
		if err == nil {
			finalPDF = signedPDF
		}
		// If signing fails, we still return the unsigned PDF
	}

	return finalPDF, nil
}

// generateAllContentWithImages processes the template and generates content with image support
func generateAllContentWithImages(template models.PDFTemplate, pageManager *PageManager, imageObjects map[int]*ImageObject, cellImageObjectIDs map[string]int, elemImageObjects map[int]*ImageObject) {
	// Initialize first page
	initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions, pageManager.Margins, pageManager.FontRegistry)

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
			initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions, pageManager.Margins, pageManager.FontRegistry)
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
				var imageKeyPrefix string
				if elem.Table != nil {
					table = *elem.Table
					imageKeyPrefix = fmt.Sprintf("elem_inline:%d", elemIdx) // Use elem_inline prefix for inline tables
				} else if elem.Index < len(template.Table) {
					table = template.Table[elem.Index]
					imageKeyPrefix = fmt.Sprintf("%d", elem.Index) // Use index as key for indexed tables
				} else {
					continue
				}
				drawTable(table, imageKeyPrefix, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
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
			// For legacy table array, use simple index as key
			imageKeyPrefix := fmt.Sprintf("%d", tableIdx)
			drawTable(table, imageKeyPrefix, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
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
		drawPageNumber(&pageManager.ContentStreams[i], i+1, totalPages, pageManager.PageDimensions, pageManager)
	}
}

// collectUsedStandardFonts returns a set of standard font names used in the template
// Always includes Helvetica as it's the default font and used for form fields
// Excludes fonts that are registered as custom fonts (e.g., Liberation fonts in PDF/A mode)
func collectUsedStandardFonts(template models.PDFTemplate, registry *CustomFontRegistry) map[string]bool {
	used := make(map[string]bool)

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

// buildCellKey2 builds a cell key with 2 integer components: "prefix:a:b"
// Replaces fmt.Sprintf("%s:%d:%d", prefix, a, b)
func buildCellKey2(prefix string, a, b int) string {
	var buf [64]byte
	n := copy(buf[:], prefix)
	buf[n] = ':'
	n++
	bA := strconv.AppendInt(buf[n:n], int64(a), 10)
	n += len(bA)
	buf[n] = ':'
	n++
	bB := strconv.AppendInt(buf[n:n], int64(b), 10)
	n += len(bB)
	return string(buf[:n])
}

// buildCellKey3 builds a cell key with 3 integer components: "a:b:c"
// Replaces fmt.Sprintf("%d:%d:%d", a, b, c) used for indexed tables
func buildCellKey3(a, b, c int) string {
	var buf [32]byte
	bA := strconv.AppendInt(buf[:0], int64(a), 10)
	n := len(bA)
	buf[n] = ':'
	n++
	bB := strconv.AppendInt(buf[n:n], int64(b), 10)
	n += len(bB)
	buf[n] = ':'
	n++
	bC := strconv.AppendInt(buf[n:n], int64(c), 10)
	n += len(bC)
	return string(buf[:n])
}

// buildCellKeyElemInline builds a cell key for inline tables: "elem_inline:elemIdx:rowIdx:colIdx"
// Replaces fmt.Sprintf("elem_inline:%d:%d:%d", elemIdx, rowIdx, colIdx)
func buildCellKeyElemInline(elemIdx, rowIdx, colIdx int) string {
	var buf [64]byte
	n := copy(buf[:], "elem_inline")
	buf[n] = ':'
	n++
	bElem := strconv.AppendInt(buf[n:n], int64(elemIdx), 10)
	n += len(bElem)
	buf[n] = ':'
	n++
	bRow := strconv.AppendInt(buf[n:n], int64(rowIdx), 10)
	n += len(bRow)
	buf[n] = ':'
	n++
	bCol := strconv.AppendInt(buf[n:n], int64(colIdx), 10)
	n += len(bCol)
	return string(buf[:n])
}
