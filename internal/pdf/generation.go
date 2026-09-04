package pdf

import (
	"bytes"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/encryption"
)

// Standard font inventory shared by the font-ID layout phase and the page
// resource and font emit phases.
var (
	standardFontNames = []string{
		"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique", // F1-F4
		"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic", // F5-F8
		"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique", // F9-F12
		"Symbol", "ZapfDingbats", // F13-F14
	}
	standardFontRefs = []string{"/F1", "/F2", "/F3", "/F4", "/F5", "/F6", "/F7", "/F8", "/F9", "/F10", "/F11", "/F12", "/F13", "/F14"}
)

// generation carries the shared state for one GenerateTemplatePDFBorrowed
// run so the decode, layout, and emit phases share one home instead of a
// long locals list. Callers build it with newGeneration right after the
// PageManager and Allocator exist, then run the phases in order.
type generation struct {
	template          models.PDFTemplate
	pageManager       *PageManager
	alloc             *Allocator
	deduper           *imageObjectDeduper
	nextImageObjectID int
	extraRegionBase   int
	// map imageIndex to ImageObject / PDF object ID.
	imageObjects   map[int]*ImageObject
	imageObjectIDs map[int]int
	// map tableIdx:rowIdx:colIdx (or "title:row:col") to XObject.
	cellImageObjects   map[string]*ImageObject
	cellImageObjectIDs map[string]int
	// map element index to ImageObject / PDF object ID.
	elemImageObjects   map[int]*ImageObject
	elemImageObjectIDs map[int]int
}

func newGeneration(template models.PDFTemplate, pageManager *PageManager, alloc *Allocator) *generation {
	// Image object IDs come from the page manager allocator so they can never
	// collide with page, content, font, or extra object IDs (a fixed base
	// collided once documents grew past ~997 pages).
	// extraRegionBase marks where allocator-backed IDs start; the dense
	// low-region layout (pages/contents/fonts) must stay clear of it.
	nextID := alloc.Next()
	return &generation{
		template:           template,
		pageManager:        pageManager,
		alloc:              alloc,
		deduper:            newImageObjectDeduper(),
		nextImageObjectID:  nextID,
		extraRegionBase:    nextID,
		imageObjects:       make(map[int]*ImageObject),
		imageObjectIDs:     make(map[int]int),
		cellImageObjects:   make(map[string]*ImageObject),
		cellImageObjectIDs: make(map[string]int),
		elemImageObjects:   make(map[int]*ImageObject),
		elemImageObjectIDs: make(map[int]int),
	}
}

// decodeImages decodes every image reference in the template over one
// iterator, interning duplicates through the shared deduper so repeated
// cell PNGs point at one shared XObject instead of serializing duplicates.
// Sources run in fixed order (standalone, title table, indexed tables,
// inline tables, element images) so first-seen ObjectID assignment matches
// the historical layout. Empty data and decode errors are skipped.
func (g *generation) decodeImages() {
	type ref struct {
		data  string
		store func(*ImageObject)
	}
	var refs []ref
	for i := range g.template.Image {
		img := &g.template.Image[i]
		if img.ImageData == "" {
			continue
		}
		i := i
		data := img.ImageData
		refs = append(refs, ref{data: data, store: func(o *ImageObject) {
			g.imageObjects[i] = o
			g.imageObjectIDs[i] = o.ObjectID
		}})
	}
	if tbl := g.template.Title.Table; tbl != nil {
		for rowIdx := range tbl.Rows {
			for colIdx := range tbl.Rows[rowIdx].Row {
				cell := &tbl.Rows[rowIdx].Row[colIdx]
				if cell.Image == nil || cell.Image.ImageData == "" {
					continue
				}
				key := buildCellKey2("title", rowIdx, colIdx)
				data := cell.Image.ImageData
				refs = append(refs, ref{data: data, store: func(o *ImageObject) {
					g.cellImageObjects[key] = o
					g.cellImageObjectIDs[key] = o.ObjectID
				}})
			}
		}
	}
	for tableIdx := range g.template.Table {
		for rowIdx := range g.template.Table[tableIdx].Rows {
			for colIdx := range g.template.Table[tableIdx].Rows[rowIdx].Row {
				cell := &g.template.Table[tableIdx].Rows[rowIdx].Row[colIdx]
				if cell.Image == nil || cell.Image.ImageData == "" {
					continue
				}
				// Key for indexed tables: "0:0:0" (tableIdx:rowIdx:colIdx)
				key := buildCellKey3(tableIdx, rowIdx, colIdx)
				data := cell.Image.ImageData
				refs = append(refs, ref{data: data, store: func(o *ImageObject) {
					g.cellImageObjects[key] = o
					g.cellImageObjectIDs[key] = o.ObjectID
				}})
			}
		}
	}
	for elemIdx := range g.template.Elements {
		elem := &g.template.Elements[elemIdx]
		if elem.Type == "table" && elem.Table != nil { //nolint:goconst
			for rowIdx := range elem.Table.Rows {
				for colIdx := range elem.Table.Rows[rowIdx].Row {
					cell := &elem.Table.Rows[rowIdx].Row[colIdx]
					if cell.Image == nil || cell.Image.ImageData == "" {
						continue
					}
					// Key for inline tables: "elem_inline:5:0:0" (elem_inline:elemIdx:rowIdx:colIdx)
					key := buildCellKeyElemInline(elemIdx, rowIdx, colIdx)
					data := cell.Image.ImageData
					refs = append(refs, ref{data: data, store: func(o *ImageObject) {
						g.cellImageObjects[key] = o
						g.cellImageObjectIDs[key] = o.ObjectID
					}})
				}
			}
		}
		if elem.Type == "image" && elem.Image != nil && elem.Image.ImageData != "" {
			elemIdx := elemIdx
			data := elem.Image.ImageData
			refs = append(refs, ref{data: data, store: func(o *ImageObject) {
				g.elemImageObjects[elemIdx] = o
				g.elemImageObjectIDs[elemIdx] = o.ObjectID
			}})
		}
	}
	for _, r := range refs {
		imgObj, err := DecodeImageData(r.data)
		if err != nil {
			continue
		}
		imgObj = g.deduper.intern(imgObj, &g.nextImageObjectID)
		r.store(imgObj)
	}
}

// fontLayout is the output of the font-ID layout phase: dense low-region
// content/font blocks plus the per-font object inventories.
type fontLayout struct {
	contentStart  int
	fontStart     int
	objectIDs     map[string]int
	descriptorIDs map[string]int
	widthsIDs     map[string]int
	used          map[string]bool
	shouldEmbed   bool
	widthGroups   map[string]string
}

// layoutFontIDs assigns the content-stream and std-font object ID blocks
// for the laid-out page count and reserves every standard-font dictionary,
// descriptor, and widths ID. It ends with EnsureBeyond so later extras can
// never reuse the font block.
func (g *generation) layoutFontIDs(fontRegistry *CustomFontRegistry) (fontLayout, error) {
	var fl fontLayout
	totalPages := len(g.pageManager.Pages)
	contentStart, fontStart, err := g.alloc.LayoutContentFontIDs(totalPages, g.extraRegionBase)
	if err != nil {
		return fl, err
	}
	used := collectUsedStandardFonts(g.template, fontRegistry)

	// If signature is enabled and visible, force usage of Helvetica
	if g.template.Config.Signature != nil && g.template.Config.Signature.Enabled && g.template.Config.Signature.Visible {
		used["Helvetica"] = true
	}

	shouldEmbed := true
	if g.template.Config.EmbedStandardFonts != nil {
		shouldEmbed = *g.template.Config.EmbedStandardFonts
	} else if g.template.Config.EmbedFonts != nil {
		shouldEmbed = *g.template.Config.EmbedFonts
	}

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

	// Only assign IDs for fonts that are used
	objectIDs := make(map[string]int)
	descriptorIDs := make(map[string]int)
	widthsIDs := make(map[string]int)
	currentObjectID := fontStart

	// Phase 1: Assign IDs for Font Dictionaries
	for _, name := range standardFontNames {
		if used[name] {
			objectIDs[name] = currentObjectID
			currentObjectID++
		}
	}

	// Phase 2: Assign IDs for Descriptors and Widths (Arlington mode only)
	if g.template.Config.ArlingtonCompatible && shouldEmbed {
		for _, name := range standardFontNames {
			if used[name] {
				descriptorIDs[name] = currentObjectID
				currentObjectID++
			}
		}
		assignedGroups := make(map[string]bool)
		for _, name := range standardFontNames {
			if used[name] {
				group := widthGroups[name]
				if !assignedGroups[group] {
					widthsIDs[group] = currentObjectID
					currentObjectID++
					assignedGroups[group] = true
				}
			}
		}
	}

	// Keep the allocator past the font block so later extras (signature
	// field, metadata, struct tree, ICC, AcroForm, info, encrypt) can never
	// reuse those IDs. In the legacy layout this is a no-op (fonts sit far
	// below NextObjectID); in a shifted layout it advances past the fonts.
	// It must run before signature creation, which allocates via the manager.
	g.alloc.EnsureBeyond(currentObjectID)

	fl = fontLayout{
		contentStart:  contentStart,
		fontStart:     fontStart,
		objectIDs:     objectIDs,
		descriptorIDs: descriptorIDs,
		widthsIDs:     widthsIDs,
		used:          used,
		shouldEmbed:   shouldEmbed,
		widthGroups:   widthGroups,
	}
	return fl, nil
}

// emitImageXObjects writes every decoded image XObject once over one deduped
// map. When useICC is set, images use the ICC-based color space for PDF/A
// compliance. A nil enc writes plain XObjects.
func (g *generation) emitImageXObjects(pdfBuffer *bytes.Buffer, iccColorSpace string, useICC bool, enc *encryption.PDFEncryption) {
	all := make([]*ImageObject, 0, len(g.imageObjects)+len(g.cellImageObjects)+len(g.elemImageObjects))
	for _, imgObj := range g.imageObjects {
		all = append(all, imgObj)
	}
	for _, imgObj := range g.cellImageObjects {
		all = append(all, imgObj)
	}
	for _, imgObj := range g.elemImageObjects {
		all = append(all, imgObj)
	}
	written := make(map[int]struct{}, g.deduper.uniqueObjectCount())
	for _, imgObj := range all {
		if _, exists := written[imgObj.ObjectID]; exists {
			continue
		}
		written[imgObj.ObjectID] = struct{}{}
		// PDF/UA-2: Ensure images use the ICC profile for color space
		if useICC {
			imgObj.ColorSpace = iccColorSpace
		}
		g.alloc.SetOffset(imgObj.ObjectID, pdfBuffer.Len())
		if enc != nil {
			pdfBuffer.Write(CreateEncryptedImageXObject(imgObj, imgObj.ObjectID, enc))
		} else {
			pdfBuffer.Write(CreateImageXObject(imgObj, imgObj.ObjectID))
		}
	}
}
