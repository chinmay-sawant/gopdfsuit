package pdf

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/encryption"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/signature"
	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/fontutils"
)

type BorrowedPDF struct {
	buf   *bytes.Buffer
	bytes []byte
}

func (d *BorrowedPDF) Bytes() []byte {
	if d == nil {
		return nil
	}
	if d.buf != nil {
		return d.buf.Bytes()
	}
	return d.bytes
}

func (d *BorrowedPDF) Len() int {
	return len(d.Bytes())
}

func (d *BorrowedPDF) Release() {
	if d == nil || d.buf == nil {
		return
	}
	pdfBufferPool.Put(d.buf)
	d.buf = nil
}

// pdfBufferPool reuses bytes.Buffer across PDF generations to reduce GC pressure.
var pdfBufferPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
		buf.Grow(256 * 1024) // 256KB initial capacity reduces final assembly growth for multi-page PDFs.
		return buf
	},
}

// pageCompressSlots limits concurrent per-page zlib compression (C4: reduces flate.NewWriter churn).
var pageCompressSlots = make(chan struct{}, maxPageCompressWorkers())

func maxPageCompressWorkers() int {
	n := runtime.NumCPU()
	if n < 4 {
		return 4
	}
	return n
}

// scratchBufPool reuses the small scratch buffer for strconv.Append* operations.
var scratchBufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 128)
		return &buf
	},
}

// signatureContextAdapter wraps PageManager to implement signature.SignaturePageContext
type signatureContextAdapter struct {
	pm *PageManager
}

func (a *signatureContextAdapter) AllocObjectID() int {
	id := a.pm.NextObjectID
	a.pm.NextObjectID++
	return id
}

func (a *signatureContextAdapter) SetExtraObject(id int, content string) {
	a.pm.ExtraObjects[id] = []byte(content)
}

func (a *signatureContextAdapter) AppendPageAnnot(pageIndex int, annotID int) {
	for len(a.pm.PageAnnots) <= pageIndex {
		a.pm.PageAnnots = append(a.pm.PageAnnots, []int{})
	}
	a.pm.PageAnnots[pageIndex] = append(a.pm.PageAnnots[pageIndex], annotID)
}

func (a *signatureContextAdapter) GetMargins() signature.PageMargins {
	return signature.PageMargins{Right: a.pm.Margins.Right, Bottom: a.pm.Margins.Bottom}
}

func (a *signatureContextAdapter) FontHas(name string) bool {
	return a.pm.FontRegistry.HasFont(name)
}

func (a *signatureContextAdapter) FontMarkChars(name, text string) {
	a.pm.FontRegistry.MarkCharsUsed(name, text)
}

func (a *signatureContextAdapter) EncodeTextForFont(fontName, text string) string {
	return EncodeTextForCustomFont(fontName, text, a.pm.FontRegistry)
}

// WarmRuntimePools pre-warms compression and buffer pools at process start.
func WarmRuntimePools() {
	WarmCompressionPools(maxPageCompressWorkers())
}

// GenerateTemplatePDF generates a PDF document with multi-page support and embedded images.
// Returns the PDF bytes and any error encountered during generation.
//
//nolint:gocyclo
func GenerateTemplatePDF(template models.PDFTemplate) ([]byte, error) {
	doc, err := GenerateTemplatePDFBorrowed(template)
	if err != nil {
		return nil, err
	}
	defer doc.Release()
	return slices.Clone(doc.Bytes()), nil
}

//nolint:gocyclo // large template renderer with many element-type branches
func GenerateTemplatePDFBorrowed(template models.PDFTemplate) (doc *BorrowedPDF, err error) {
	pdfBufferPtr := pdfBufferPool.Get().(*bytes.Buffer)
	pdfBufferPtr.Reset()
	pdfBuffer := pdfBufferPtr // use directly as *bytes.Buffer
	ensurePDFBufferCapacity(pdfBuffer, estimateTemplatePDFBufferSize(template))
	defer func() {
		if err != nil {
			pdfBufferPool.Put(pdfBuffer)
		}
	}()

	scratchPtr := scratchBufPool.Get().(*[]byte)
	b := (*scratchPtr)[:0]
	defer func() { *scratchPtr = b[:0]; scratchBufPool.Put(scratchPtr) }()

	xrefOffsets := make(map[int]int)

	// Get page dimensions from config
	pageConfig := template.Config
	pageDims := getPageDimensions(pageConfig.Page, pageConfig.PageAlignment)
	pageMargins := ParsePageMargins(pageConfig.PageMargin)
	taggedPDF := template.Config.TaggedPDF || template.Config.PDFACompliant

	// Create a local clone of the font registry for this PDF generation session
	// This ensures thread safety by isolating usage tracking (UsedChars) per generation
	globalRegistry := GetFontRegistry()
	fontRegistry := globalRegistry.CloneForGeneration()
	defer globalRegistry.ReleaseGenerationClone(fontRegistry)

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

	// Auto-resolve math fonts: scan for math-enabled cells referencing unregistered fonts
	autoResolveMathFonts(template, fontRegistry)

	// Initialize page manager with Arlington compatibility flag and per-generation font registry
	pageManager := NewPageManager(pageDims, pageMargins, template.Config.ArlingtonCompatible, fontRegistry, taggedPDF, estimateInitialContentStreamCap(template))
	defer pageManager.ReleaseContentStreams()
	if taggedPDF {
		pageManager.Structure.ReserveElementCapacity(estimateStructureElementCount(template))
	}

	// Process images and create XObjects
	imageObjects := make(map[int]*ImageObject) // map imageIndex to ImageObject
	imageObjectIDs := make(map[int]int)        // map imageIndex to PDF object ID

	// Process cell images - map tableIdx:rowIdx:colIdx to XObject ID
	// Also process title table images with prefix "title:"
	cellImageObjects := make(map[string]*ImageObject)
	cellImageObjectIDs := make(map[string]int)

	nextImageObjectID := 1000 // Start image objects at ID 1000

	// Reuse identical decoded images across all document references so repeated
	// cell PNGs point at one shared XObject instead of serializing duplicates.
	imageDeduper := newImageObjectDeduper()

	// Process standalone images
	for i, img := range template.Image {
		if img.ImageData != "" {
			imgObj, err := DecodeImageData(img.ImageData)
			if err == nil {
				imgObj = imageDeduper.intern(imgObj, &nextImageObjectID)
				imageObjects[i] = imgObj
				imageObjectIDs[i] = imgObj.ObjectID
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
						imgObj = imageDeduper.intern(imgObj, &nextImageObjectID)
						cellKey := buildCellKey2("title", rowIdx, colIdx)
						cellImageObjects[cellKey] = imgObj
						cellImageObjectIDs[cellKey] = imgObj.ObjectID
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
						imgObj = imageDeduper.intern(imgObj, &nextImageObjectID)
						// Key for indexed tables: "0:0:0" (tableIdx:rowIdx:colIdx)
						cellKey := buildCellKey3(tableIdx, rowIdx, colIdx)
						cellImageObjects[cellKey] = imgObj
						cellImageObjectIDs[cellKey] = imgObj.ObjectID
					}
				}
			}
		}
	}

	// Process inline table images in Elements array
	for elemIdx, elem := range template.Elements {
		if elem.Type == "table" && elem.Table != nil { //nolint:goconst
			for rowIdx, row := range elem.Table.Rows {
				for colIdx, cell := range row.Row {
					if cell.Image != nil && cell.Image.ImageData != "" {
						imgObj, err := DecodeImageData(cell.Image.ImageData)
						if err == nil {
							imgObj = imageDeduper.intern(imgObj, &nextImageObjectID)
							// Key for inline tables: "elem_inline:5:0:0" (elem_inline:elemIdx:rowIdx:colIdx)
							cellKey := buildCellKeyElemInline(elemIdx, rowIdx, colIdx)
							cellImageObjects[cellKey] = imgObj
							cellImageObjectIDs[cellKey] = imgObj.ObjectID
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
				imgObj = imageDeduper.intern(imgObj, &nextImageObjectID)
				elemImageObjects[elemIdx] = imgObj
				elemImageObjectIDs[elemIdx] = imgObj.ObjectID
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
	ensurePDFBufferCapacity(pdfBuffer, estimateFinalPDFSize(pageManager, imageDeduper.uniqueObjectCount()))

	// Generate font subsets MOVED to after signature generation to ensure signature chars are included

	// Setup encryption EARLY if security config is provided (before writing content)
	// This is needed because content streams need to be encrypted
	var enc *encryption.PDFEncryption
	if template.Config.Security != nil && template.Config.Security.Enabled && template.Config.Security.OwnerPassword != "" {
		// Generate a preliminary document ID for encryption setup
		preliminaryID := encryption.GenerateDocumentID(append([]byte(template.Title.Text), strconv.AppendInt((*scratchPtr)[:0], int64(len(pageManager.Pages)), 10)...))
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
	var pdfSigner *signature.PDFSigner
	var sigIDs *signature.SignatureIDs
	if signatureEnabled {
		var err error
		pdfSigner, err = signature.NewPDFSigner(template.Config.Signature)
		if err == nil && pdfSigner != nil {
			// Get the font ID for signature appearance
			// In PDF/A mode, this returns the Liberation font ID that replaces Helvetica
			// In standard mode, this returns the standard Helvetica font ID
			signatureFontID := getWidgetFontObjectID(pageManager.FontRegistry)
			if signatureFontID == 0 {
				// Fallback to standard font object ID if no custom font
				signatureFontID = fontObjectIDs["Helvetica"]
			}

			sigPageDims := signature.PageDimensions{
				Width:  pageDims.Width,
				Height: pageDims.Height,
			}
			sigIDs = pdfSigner.CreateSignatureField(&signatureContextAdapter{pm: pageManager}, sigPageDims, signatureFontID)
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
				if bytes.Contains(content, []byte("/Subtype /Widget")) {
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

	var structTreeRootID int
	if taggedPDF {
		structTreeRootID = pageManager.NextObjectID
		pageManager.NextObjectID++
	}
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
	if taggedPDF {
		// Tagged PDF catalog entries (omit when generating untagged PDFs for performance/size)
		pdfBuffer.WriteString(" /MarkInfo << /Marked true >>")
	}

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

		var acroFormContent []byte
		if sigIDs != nil {
			acroFormContent = append(acroFormContent, "<< /Fields "...)
			acroFormContent = append(acroFormContent, fieldsRef.String()...)
			acroFormContent = append(acroFormContent, " /DA ("...)
			acroFormContent = append(acroFormContent, widgetFontRef...)
			acroFormContent = append(acroFormContent, " 0 Tf 0 g) /SigFlags "...)
			acroFormContent = strconv.AppendInt(acroFormContent, int64(signature.GetAcroFormSigFlags()), 10)
			acroFormContent = append(acroFormContent, " >>"...)
		} else {
			acroFormContent = append(acroFormContent, "<< /Fields "...)
			acroFormContent = append(acroFormContent, fieldsRef.String()...)
			acroFormContent = append(acroFormContent, " /DA ("...)
			acroFormContent = append(acroFormContent, widgetFontRef...)
			acroFormContent = append(acroFormContent, " 0 Tf 0 g) >>"...)
		}
		pageManager.ExtraObjects[acroFormID] = acroFormContent

		pdfBuffer.WriteString(" /AcroForm ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(acroFormID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R")
	}

	if taggedPDF {
		pdfBuffer.WriteString(" /StructTreeRoot ")
		b = b[:0]
		b = strconv.AppendInt(b, int64(structTreeRootID), 10)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString(" 0 R")
	}

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
				if bytes.Contains(content, []byte("/Type /XObject")) {
					writeXObjRefInt("X", id, id)
				}
			}
			xobjBuilder.WriteString(" >>")
		} else {
			hasXObjects := false
			for _, content := range pageManager.ExtraObjects {
				if bytes.Contains(content, []byte("/Type /XObject")) {
					hasXObjects = true
					break
				}
			}
			if hasXObjects {
				xobjBuilder.WriteString(" /XObject <<")
				for id, content := range pageManager.ExtraObjects {
					if bytes.Contains(content, []byte("/Type /XObject")) {
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
		if i < len(pageManager.Structure.NextMCID) && pageManager.Structure.NextMCID[i] > 0 {
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

	// Generate content stream objects with FlateDecode compression (zlib in parallel, encrypt + write serialized).
	// Each page is compressed into a pooled bytes.Buffer that we hand off to
	// the write loop. The buffer is only returned to the pool after the
	// serialized consumer is done with it, which lets us skip the
	// slices.Clone that was previously needed to escape pool ownership.
	nStreams := len(pageManager.ContentStreams)
	compressedPages := make([]*bytes.Buffer, nStreams)
	useFlate := make([]bool, nStreams)
	var compGroup errgroup.Group
	for si := range pageManager.ContentStreams {
		si := si
		compGroup.Go(func() error {
			pageCompressSlots <- struct{}{}
			defer func() { <-pageCompressSlots }()

			contentStream := pageManager.ContentStreams[si]
			compressedBuf, ok := compressContentStream(contentStream.Bytes())
			if !ok {
				useFlate[si] = false
				return nil
			}
			useFlate[si] = true
			compressedPages[si] = compressedBuf
			return nil
		})
	}
	if err := compGroup.Wait(); err != nil {
		return nil, err
	}
	for i := range pageManager.ContentStreams {
		objectID := contentObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(objectID), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)

		var streamData []byte
		if useFlate[i] {
			streamData = compressedPages[i].Bytes()
		} else {
			streamData = pageManager.ContentStreams[i].Bytes()
		}

		// Encrypt content stream per object order when encryption is enabled
		switch {
		case enc != nil:
			encryptedData := enc.EncryptStream(streamData, objectID, 0)
			if useFlate[i] {
				pdfBuffer.WriteString("<< /Filter /FlateDecode /Length ")
			} else {
				pdfBuffer.WriteString("<< /Length ")
			}
			b = b[:0]
			b = strconv.AppendInt(b, int64(len(encryptedData)), 10)
			pdfBuffer.Write(b)
			pdfBuffer.WriteString(" >>\nstream\n")
			pdfBuffer.Write(encryptedData)
		case useFlate[i]:
			pdfBuffer.WriteString("<< /Filter /FlateDecode /Length ")
			b = b[:0]
			b = strconv.AppendInt(b, int64(len(streamData)), 10)
			pdfBuffer.Write(b)
			pdfBuffer.WriteString(" >>\nstream\n")
			pdfBuffer.Write(streamData)
		default:
			pdfBuffer.WriteString("<< /Length ")
			b = b[:0]
			b = strconv.AppendInt(b, int64(len(streamData)), 10)
			pdfBuffer.Write(b)
			pdfBuffer.WriteString(" >>\nstream\n")
			pdfBuffer.Write(streamData)
		}
		pdfBuffer.WriteString("\nendstream\nendobj\n")
		if useFlate[i] {
			putCompressBuffer(compressedPages[i])
			compressedPages[i] = nil
		}
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

	// Pre-compute ICC-based color space string for PDF/A images
	iccColorSpace := "[/ICCBased " + strconv.Itoa(actualICCProfileObjID) + " 0 R]"

	// Generate image XObjects (standalone images)
	writtenImageObjects := make(map[int]struct{}, imageDeduper.uniqueObjectCount())
	for _, imgObj := range imageObjects {
		if _, exists := writtenImageObjects[imgObj.ObjectID]; exists {
			continue
		}
		writtenImageObjects[imgObj.ObjectID] = struct{}{}
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if template.Config.PDFACompliant && actualICCProfileObjID > 0 {
			imgObj.ColorSpace = iccColorSpace
		}

		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		if enc != nil {
			pdfBuffer.Write(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.Write(CreateImageXObject(imgObj, imgObj.ObjectID))
		}
	}

	// Generate image XObjects (cell images)
	for _, imgObj := range cellImageObjects {
		if _, exists := writtenImageObjects[imgObj.ObjectID]; exists {
			continue
		}
		writtenImageObjects[imgObj.ObjectID] = struct{}{}
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if template.Config.PDFACompliant && actualICCProfileObjID > 0 {
			imgObj.ColorSpace = iccColorSpace
		}

		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		if enc != nil {
			pdfBuffer.Write(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.Write(CreateImageXObject(imgObj, imgObj.ObjectID))
		}
	}

	// Generate image XObjects (element images)
	for _, imgObj := range elemImageObjects {
		if _, exists := writtenImageObjects[imgObj.ObjectID]; exists {
			continue
		}
		writtenImageObjects[imgObj.ObjectID] = struct{}{}
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if template.Config.PDFACompliant && actualICCProfileObjID > 0 {
			imgObj.ColorSpace = iccColorSpace
		}

		xrefOffsets[imgObj.ObjectID] = pdfBuffer.Len()
		if enc != nil {
			pdfBuffer.Write(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.Write(CreateImageXObject(imgObj, imgObj.ObjectID))
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
		pdfBuffer.Write(content)
		pdfBuffer.WriteString("\nendobj\n")
	}

	// Capture generation time once for reuse in metadata
	genTime := time.Now()

	// Generate PDF/A metadata objects if enabled
	if pdfaHandler != nil {
		// Generate XMP metadata content (but use our pre-reserved metadataObjectID for consistency with Catalog)
		docIDForXMP := strconv.FormatInt(genTime.UnixNano(), 16)
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
	now := genTime
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

	// Generate Document ID (SHA-256 content hash + random hash — PDF spec requires
	// a content-based permanent identifier and a per-generation random identifier).
	// SHA-256 is chosen over MD5 because modern amd64 CPUs provide AVX2 acceleration
	// for SHA-256 that makes it ~2x faster than MD5 for bulk hashing.
	sum := sha256.Sum256(pdfBuffer.Bytes())
	var contentHashArr [16]byte
	copy(contentHashArr[:], sum[:16])
	var documentID string
	if enc != nil {
		documentID = encryption.FormatDocumentID(enc.DocumentID)
	} else {
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			binary.BigEndian.PutUint64(randomBytes, uint64(time.Now().UnixNano()))
		}
		randomSum := sha256.Sum256(randomBytes)
		var randomHashArr [16]byte
		copy(randomHashArr[:], randomSum[:16])
		documentID = fmt.Sprintf("[<%s> <%s>]", hex.EncodeToString(contentHashArr[:]), hex.EncodeToString(randomHashArr[:]))
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

	if taggedPDF {
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
				if kid.Elem != nil {
					assignStructIDs(kid.Elem)
				}
			}
		}
		// Start from root's children (Root itself has structTreeRootID, its children are elements)
		for _, kid := range pageManager.Structure.Root.Kids {
			if kid.Elem != nil {
				assignStructIDs(kid.Elem)
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
		ptBuilder.Grow(128 + len(pageManager.Pages)*32 + len(pageManager.AnnotStructElems)*24)
		ptBuilder.WriteString(strconv.Itoa(parentTreeID))
		ptBuilder.WriteString(" 0 obj\n<< /Nums [")

		// Iterate through all pages that have marked content
		// We iterate by page index to keep Nums sorted
		maxPageIndex := len(pageManager.Pages)
		for i := 0; i < maxPageIndex; i++ {
			if i < len(pageManager.Structure.ParentTree) {
				elems := pageManager.Structure.ParentTree[i]
				if len(elems) > 0 {
					ptBuilder.WriteByte(' ')
					b = b[:0]
					b = strconv.AppendInt(b, int64(i), 10)
					ptBuilder.Write(b)
					ptBuilder.WriteString(" [") // Key is page index
					for _, elem := range elems {
						appendObjRefToBuilder(&ptBuilder, elem.ObjectID)
					}
					ptBuilder.WriteString(" ]")
				}
			}
		}

		// PDF/UA-2: Add ParentTree entries for annotation StructParents
		// Each annotation's StructParent value maps to its Link structure element
		for _, annotInfo := range pageManager.AnnotStructElems {
			if linkElem, exists := pageManager.Structure.LinkElements[annotInfo.AnnotObjID]; exists {
				ptBuilder.WriteByte(' ')
				b = b[:0]
				b = strconv.AppendInt(b, int64(annotInfo.StructParentIdx), 10)
				ptBuilder.Write(b)
				appendObjRefToBuilder(&ptBuilder, linkElem.ObjectID)
			}
		}

		ptBuilder.WriteString(" ] >>\nendobj\n")
		pdfBuffer.WriteString(ptBuilder.String())

		// Write all Structure Elements
		var writeStructElems func(elem *StructElem)
		writeStructElems = func(elem *StructElem) {
			xrefOffsets[elem.ObjectID] = pdfBuffer.Len()
			formatStructElemObjectTo(pdfBuffer, elem, structElemFormatCtx{
				namespaceID:      namespaceID,
				structTreeRootID: structTreeRootID,
				root:             pageManager.Structure.Root,
				pages:            pageManager.Pages,
			})
			for _, k := range elem.Kids {
				if k.Elem != nil {
					writeStructElems(k.Elem)
				}
			}
		}
		for _, kid := range pageManager.Structure.Root.Kids {
			if kid.Elem != nil {
				writeStructElems(kid.Elem)
			}
		}

		pageManager.Structure.ReleaseStructElemsToPool()
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

	if pdfSigner != nil && sigIDs != nil {
		if err := signature.UpdatePDFWithSignatureBuffer(pdfBuffer, pdfSigner); err != nil {
			return &BorrowedPDF{buf: pdfBuffer}, nil
		}
	}

	return &BorrowedPDF{buf: pdfBuffer}, nil
}

func appendObjRefToBuilder(sb *strings.Builder, objID int) {
	sb.WriteByte(' ')
	sb.WriteString(strconv.Itoa(objID))
	sb.WriteString(" 0 R")
}

type structElemFormatCtx struct {
	namespaceID      int
	structTreeRootID int
	root             *StructElem
	pages            []int
}

var structElemBuilderPool = sync.Pool{
	New: func() any {
		sb := new(strings.Builder)
		sb.Grow(256)
		return sb
	},
}

func formatStructElemObject(elem *StructElem, ctx structElemFormatCtx) string {
	sb := structElemBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	formatStructElemObjectInto(sb, elem, ctx)
	out := sb.String()
	structElemBuilderPool.Put(sb)
	return out
}

func formatStructElemObjectInto(sb *strings.Builder, elem *StructElem, ctx structElemFormatCtx) {
	formatStructElemObjectTo(sb, elem, ctx)
}

func estimateStructureElementCount(template models.PDFTemplate) int {
	count := 1 // Document

	if template.Title.Text != "" {
		count++
	}
	if template.Title.Table != nil {
		count += 1 + len(template.Title.Table.Rows)
		for _, row := range template.Title.Table.Rows {
			count += len(row.Row)
		}
	}

	for _, table := range template.Table {
		count += 1 + len(table.Rows)
		for _, row := range table.Rows {
			count += min(len(row.Row), table.MaxColumns)
		}
	}

	for _, elem := range template.Elements {
		if elem.Table == nil {
			continue
		}
		count += 1 + len(elem.Table.Rows)
		for _, row := range elem.Table.Rows {
			count += min(len(row.Row), elem.Table.MaxColumns)
		}
	}

	if template.Footer.Text != "" {
		count++
	}

	return count
}

type structElemObjectWriter interface {
	Grow(int)
	WriteString(string) (int, error)
	Write([]byte) (int, error)
	WriteByte(byte) error
}

func formatStructElemObjectTo(w structElemObjectWriter, elem *StructElem, ctx structElemFormatCtx) {
	w.Grow(128 + len(elem.Kids)*16 + len(elem.Title) + len(elem.Alt))
	var scratch [24]byte
	w.Write(strconv.AppendInt(scratch[:0], int64(elem.ObjectID), 10))
	w.WriteString(" 0 obj\n<< /Type /StructElem /S /")
	w.WriteString(string(elem.Type))

	if elem.Type == StructDocument {
		w.WriteString(" /NS ")
		w.Write(strconv.AppendInt(scratch[:0], int64(ctx.namespaceID), 10))
		w.WriteString(" 0 R")
	}

	if elem.Parent == ctx.root {
		w.WriteString(" /P ")
		w.Write(strconv.AppendInt(scratch[:0], int64(ctx.structTreeRootID), 10))
		w.WriteString(" 0 R")
	} else if elem.Parent != nil {
		w.WriteString(" /P ")
		w.Write(strconv.AppendInt(scratch[:0], int64(elem.Parent.ObjectID), 10))
		w.WriteString(" 0 R")
	}

	if elem.Title != "" {
		w.WriteString(" /T (")
		var scratch [1024]byte
		escaped := appendEscapedPDFLiteral(scratch[:0], elem.Title)
		w.Write(escaped)
		w.WriteByte(')')
	}
	if elem.Alt != "" {
		w.WriteString(" /Alt (")
		var scratch [1024]byte
		escaped := appendEscapedPDFLiteral(scratch[:0], elem.Alt)
		w.Write(escaped)
		w.WriteByte(')')
	}

	if len(elem.Kids) > 0 || elem.Type == StructLink {
		w.WriteString(" /K [")

		if elem.Type == StructLink {
			pageObjID := 3
			if elem.PageID >= 0 && elem.PageID < len(ctx.pages) {
				pageObjID = ctx.pages[elem.PageID]
			}
			w.WriteString(" << /Type /OBJR /Obj ")
			w.Write(strconv.AppendInt(scratch[:0], int64(elem.AnnotObjID), 10))
			w.WriteString(" 0 R /Pg ")
			w.Write(strconv.AppendInt(scratch[:0], int64(pageObjID), 10))
			w.WriteString(" 0 R >>")
		}

		var mcidScratch [16]byte
		for _, kid := range elem.Kids {
			if kid.Elem != nil {
				appendObjRefToWriter(w, kid.Elem.ObjectID)
			} else {
				w.WriteByte(' ')
				mcid := strconv.AppendInt(mcidScratch[:0], int64(kid.MCID), 10)
				w.Write(mcid)
			}
		}
		w.WriteString(" ]")
	}

	if elem.PageID >= 0 && elem.PageID < len(ctx.pages) {
		pageObjID := ctx.pages[elem.PageID]
		w.WriteString(" /Pg ")
		w.Write(strconv.AppendInt(scratch[:0], int64(pageObjID), 10))
		w.WriteString(" 0 R")
	}

	w.WriteString(" >>\nendobj\n")
}

func appendObjRefToWriter(w structElemObjectWriter, objID int) {
	w.WriteByte(' ')
	var scratch [24]byte
	w.Write(strconv.AppendInt(scratch[:0], int64(objID), 10))
	w.WriteString(" 0 R")
}

type imageObjectKey struct {
	hash         uint64
	sourceLen    int
	width        int
	height       int
	bitsPerComp  int
	imageDataLen int
	filter       string
	isForm       bool
}

type imageObjectDeduper struct {
	objects map[imageObjectKey][]*ImageObject
	unique  int
}

func newImageObjectDeduper() *imageObjectDeduper {
	return &imageObjectDeduper{
		objects: make(map[imageObjectKey][]*ImageObject),
	}
}

func (d *imageObjectDeduper) intern(imgObj *ImageObject, nextObjectID *int) *ImageObject {
	key := imageObjectKey{
		hash:         imgObj.CacheKey,
		sourceLen:    imgObj.SourceLen,
		width:        imgObj.Width,
		height:       imgObj.Height,
		bitsPerComp:  imgObj.BitsPerComp,
		imageDataLen: imgObj.ImageDataLen,
		filter:       imgObj.Filter,
		isForm:       imgObj.IsForm,
	}
	for _, existing := range d.objects[key] {
		if len(existing.ImageData) == len(imgObj.ImageData) && bytes.Equal(existing.ImageData, imgObj.ImageData) {
			return existing
		}
	}
	imgObj.ObjectID = *nextObjectID
	*nextObjectID++
	d.objects[key] = append(d.objects[key], imgObj)
	d.unique++
	return imgObj
}

func (d *imageObjectDeduper) uniqueObjectCount() int {
	if d == nil {
		return 0
	}
	return d.unique
}

func ensurePDFBufferCapacity(pdfBuffer *bytes.Buffer, want int) {
	if want <= 0 {
		return
	}
	if pdfBuffer.Cap() >= want {
		return
	}
	pdfBuffer.Grow(want - pdfBuffer.Cap())
}

func estimateFinalPDFSize(pageManager *PageManager, uniqueImageObjects int) int {
	estimate := 256 * 1024 // header, catalog, xref, trailer slack
	estimate += len(pageManager.Pages) * 2048
	estimate += len(pageManager.ExtraObjects) * 512
	estimate += uniqueImageObjects * 2048
	for _, stream := range pageManager.ContentStreams {
		if stream == nil {
			continue
		}
		rawLen := stream.Len()
		// F2: pessimistic upper bound — store-uncompressed pages are rawLen; compressed ~35%.
		compressedEst := rawLen * 2 / 5
		if rawLen > compressedEst {
			estimate += rawLen + 512
		} else {
			estimate += compressedEst + 512
		}
	}
	for _, extra := range pageManager.ExtraObjects {
		estimate += len(extra) + 128
	}
	if sm := pageManager.Structure; sm != nil && sm.Enabled {
		estimate += len(sm.Elements) * 160
	}
	return estimate
}

func estimateInitialContentStreamCap(template models.PDFTemplate) int {
	const (
		minCap       = 64 * 1024
		retailMaxCap = 128 * 1024
		pageCap      = 256 * 1024
	)

	maxRows := 0
	maxCols := 1
	score := 0
	for _, table := range template.Table {
		rows := len(table.Rows)
		cols := max(table.MaxColumns, 1)
		score += rows * cols * 8
		if rows > maxRows {
			maxRows = rows
			maxCols = cols
		}
	}
	if template.Title.Text != "" {
		score += 1024
	}
	if template.Footer.Text != "" {
		score += 1024
	}
	for _, elem := range template.Elements {
		score += 512
		if elem.Table != nil {
			rows := len(elem.Table.Rows)
			cols := max(elem.Table.MaxColumns, 1)
			score += rows * cols * 8
			if rows > maxRows {
				maxRows = rows
				maxCols = cols
			}
		}
	}
	if score < minCap {
		score = minCap
	}
	if maxRows > 40 {
		rowsPerPage := 40
		perPage := rowsPerPage * maxCols * 512
		if perPage > score {
			score = perPage
		}
		if score > pageCap {
			score = pageCap
		}
		return score
	}
	if score > retailMaxCap {
		return retailMaxCap
	}
	return score
}

func estimateTemplatePDFBufferSize(template models.PDFTemplate) int {
	estimate := 192 * 1024
	for _, table := range template.Table {
		estimate += len(table.Rows) * max(table.MaxColumns, 1) * 96
	}
	for _, elem := range template.Elements {
		estimate += 2048
		if elem.Table != nil {
			estimate += len(elem.Table.Rows) * max(elem.Table.MaxColumns, 1) * 96
		}
	}
	if template.Config.TaggedPDF || template.Config.PDFACompliant {
		estimate += 64 * 1024
	}
	if template.Config.Signature != nil && template.Config.Signature.Enabled {
		estimate += 32 * 1024
	}
	return estimate
}

// generateAllContentWithImages processes the template and generates content with image support
func generateAllContentWithImages(template models.PDFTemplate, pageManager *PageManager, imageObjects map[int]*ImageObject, cellImageObjectIDs map[string]int, elemImageObjects map[int]*ImageObject) {
	// Initialize first page
	initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions, pageManager.Margins, pageManager.FontRegistry)

	var intBuf []byte

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
			appendPageInitialization(pageManager.GetCurrentContentStream(), pageManager, template.Config.PageBorder, template.Config.Watermark)
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
				switch {
				case elem.Table != nil:
					table = *elem.Table
					imageKeyPrefix = "elem_inline:" + string(strconv.AppendInt(intBuf[:0], int64(elemIdx), 10)) // Use elem_inline prefix for inline tables
				case elem.Index < len(template.Table):
					table = template.Table[elem.Index]
					imageKeyPrefix = string(strconv.AppendInt(intBuf[:0], int64(elem.Index), 10)) // Use index as key for indexed tables
				default:
					continue
				}
				drawTable(table, imageKeyPrefix, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
				tableIdx++
			case "spacer":
				var spacer models.Spacer
				switch {
				case elem.Spacer != nil:
					spacer = *elem.Spacer
				case elem.Index < len(template.Spacer):
					spacer = template.Spacer[elem.Index]
				default:
					continue
				}
				drawSpacer(spacer, pageManager)
			case "image":
				var image models.Image
				if elem.Image != nil {
					// Inline image in elements array - use element index for XObject lookup
					image = *elem.Image
					if imgObj, exists := elemImageObjects[elemIdx]; exists {
						imageXObjectRef := "/E" + string(strconv.AppendInt(intBuf[:0], int64(elemIdx), 10))
						drawImageWithXObjectInternal(image, imageXObjectRef, pageManager, template.Config.PageBorder, template.Config.Watermark, imgObj.Width, imgObj.Height)
					} else {
						// Fall back to placeholder if no XObject
						drawImage(image, pageManager, template.Config.PageBorder, template.Config.Watermark)
					}
				} else if elem.Index < len(template.Image) {
					// Reference to template.Image array
					image = template.Image[elem.Index]
					if imgObj, exists := imageObjects[elem.Index]; exists {
						imageXObjectRef := "/I" + string(strconv.AppendInt(intBuf[:0], int64(elem.Index), 10))
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
			imageKeyPrefix := string(strconv.AppendInt(intBuf[:0], int64(tableIdx), 10))
			drawTable(table, imageKeyPrefix, pageManager, template.Config.PageBorder, template.Config.Watermark, cellImageObjectIDs)
		}

		// Spacers - Process each spacer (added after tables in legacy mode)
		for _, spacer := range template.Spacer {
			drawSpacer(spacer, pageManager)
		}

		// Images - Process each image with automatic page breaks
		for i, image := range template.Image {
			if imgObj, exists := imageObjects[i]; exists {
				imageXObjectRef := "/I" + string(strconv.AppendInt(intBuf[:0], int64(i), 10))
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
			drawFooter(pageManager.ContentStreams[i], template.Footer, pageManager)
		}
		// Draw page number on this page
		drawPageNumber(pageManager.ContentStreams[i], i+1, totalPages, pageManager.PageDimensions, pageManager)
	}
}

// collectUsedStandardFonts returns a set of standard font names used in the template
// Always includes Helvetica as it's the default font and used for form fields
// Excludes fonts that are registered as custom fonts (e.g., Liberation fonts in PDF/A mode)
//
//nolint:gocyclo
func collectUsedStandardFonts(template models.PDFTemplate, registry *CustomFontRegistry) map[string]bool {
	used := make(map[string]bool)
	propFontCache := make(map[string]string, 32)
	helveticaAvailable := registry.HasFont("Helvetica")

	// Helper to mark font only if it's a true standard font (not overridden by custom)
	markFont := func(propsStr string) {
		if cached, ok := propFontCache[propsStr]; ok {
			if cached != "" {
				used[cached] = true
			}
			return
		}

		fontName := parseProps(propsStr).FontName
		if !IsCustomFont(fontName) && !registry.HasFont(fontName) {
			propFontCache[propsStr] = fontName
			used[fontName] = true
			return
		}
		propFontCache[propsStr] = ""
	}

	// Helvetica is default font - only add if not overridden by custom font
	if !helveticaAvailable {
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
		if !helveticaAvailable {
			used["Helvetica"] = true
		}
	}

	// Page numbers always use Helvetica
	if !helveticaAvailable {
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
		if !helveticaAvailable {
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

// autoResolveMathFonts scans the template for cells with mathEnabled=true that
// reference a font not yet registered in the registry. When found, it attempts
// to auto-discover a suitable math-capable system font and register it under
// the requested name so that Unicode math symbols render correctly.
func autoResolveMathFonts(template models.PDFTemplate, registry *CustomFontRegistry) {
	needed := collectUnregisteredMathFontNames(template, registry)
	if len(needed) == 0 {
		return
	}

	// Find a suitable math font on the system (cross-platform paths from fontutils)
	var fontPath string
	for _, candidate := range fontutils.MathFontCandidates() {
		if _, err := os.Stat(candidate); err == nil {
			fontPath = candidate
			break
		}
	}
	if fontPath == "" {
		return // no suitable font found on this system
	}

	for name := range needed {
		if err := registry.RegisterFontFromFile(name, fontPath); err != nil {
			fmt.Printf("Warning: auto-resolve math font %s failed: %v\n", name, err)
		}
	}
}

// collectUnregisteredMathFontNames scans all cells in the template for those
// with mathEnabled=true and returns font names (from props) not already
// registered in the font registry.
func collectUnregisteredMathFontNames(template models.PDFTemplate, registry *CustomFontRegistry) map[string]struct{} {
	needed := make(map[string]struct{})

	checkCell := func(cell models.Cell) {
		if cell.MathEnabled == nil || !*cell.MathEnabled {
			return
		}
		if cell.Props == "" {
			return
		}
		fontName := strings.SplitN(cell.Props, ":", 2)[0]
		if fontName == "" {
			return
		}
		if registry.HasFont(fontName) || IsStandardFont(fontName) {
			return
		}
		needed[fontName] = struct{}{}
	}

	// Scan title table
	if template.Title.Table != nil {
		for _, row := range template.Title.Table.Rows {
			for _, cell := range row.Row {
				checkCell(cell)
			}
		}
	}

	// Scan indexed tables
	for _, tbl := range template.Table {
		for _, row := range tbl.Rows {
			for _, cell := range row.Row {
				checkCell(cell)
			}
		}
	}

	// Scan element inline tables
	for _, elem := range template.Elements {
		if elem.Table != nil {
			for _, row := range elem.Table.Rows {
				for _, cell := range row.Row {
					checkCell(cell)
				}
			}
		}
	}

	return needed
}
