# Engine round3 refactor

Date: 2026-09-04. Branch: feat/builder-snippets. Commits: d3233e8, 48c57d2.

Ledger: plans/builder-snippets/reviews/architecture-review-2026-09-04-builder-snippets.md. Status per ledger header is implemented plus validated, 5 subagent waves complete, all gates green.

## What moved

PageManager plus Allocator. Before: pageObjectStore and pageLayoutStore were separate, Allocator had standalone plus bound modes. Now: PageManager owns Pages, ContentStreams, PageAnnots, ExtraObjects, NextObjectID, alloc at internal/pdf/pagemanager.go:92-114. Allocator at internal/pdf/allocator.go:13 is bound only, a wrapper over live PageManager plus xref offsets. Tests bind a real PageManager. Standalone state survives only as test fake.

Generator phase split. Before: GenerateTemplatePDFBorrowed held about 1200 lines with 4 image decode loops plus 3 image emit loops plus inline font ID layout. Now: generation struct at internal/pdf/generation.go:22 owns decodeImages, layoutFontIDs, emitImageXObjects. Call order in generator.go:446-469 is newGeneration, decodeImages, SeekTo, generateAllContentWithImages, layoutFontIDs, emitImageXObjects. emitPages, emitFonts, emitTrailer stay inline. The ledger records that as an accepted deviation.

Shared vector emission. Before: typstsyntax/renderer.go and internal/pdf/svg/svg.go each had float and color logic. Now: internal/pdf/vector/vector.go owns AppendFloat, WriteFloat, WriteFloats, FormatFloat, EscapeText, SetFill, SetStroke, LineWidth, StrokeLine, StrokePath. Both callers route through it.

HTML convert dedup. Before: pdf.go and pdf_js.go duplicated option mapping. Now: internal/pdf/html_convert.go:31-122 owns warnUnmappedHTMLOptions, buildPDFDocument, runPDFDocument, normalizeImageFormat, buildImageCrop, buildImageDocument, runImageDocument, parseMarginMM. pdf.go keeps ConvertHTMLToPDF at 51, ConvertHTMLToImage at 82, htmlSourceContent at 115. pdf_js.go keeps js build with ErrUpstream variant.

Table tagging seam. structure_tag.go TableTagger with 80 lines is deleted. Replacement is StructureManager.EmitRowCells at internal/pdf/structure.go:763. Callers in draw.go use drawSharedLayoutRow at 527 and drawSharedDeferRow at 414.

Models parity pin. Owned type split between internal/models and pkg/gopdflib stays. New pin is internal/models/schema_parity_test.go:1-113 with TestHandlerInputJSONParity at 70 and TestEngineOptionFieldParity at 88.

## Code

generation home at generation.go:22-42:

```go
type generation struct {
    template          models.PDFTemplate
    pageManager       *PageManager
    alloc             *Allocator
    deduper           *imageObjectDeduper
    nextImageObjectID int
    extraRegionBase   int
    imageObjects   map[int]*ImageObject
    imageObjectIDs map[int]int
    cellImageObjects   map[string]*ImageObject
    cellImageObjectIDs map[string]int
    elemImageObjects   map[int]*ImageObject
    elemImageObjectIDs map[int]int
}
```

Allocator contract at allocator.go:5-24:

```go
// Allocator is the single seam for object-ID reservation, ExtraObjects
// commit, and xref-offset recording during generation (Phase 5 D1).
// It is bound-only: it wraps a live PageManager plus the generation xref
// offsets slice, so every reservation flows through one counter. There is no
// standalone mode; unit tests bind a real PageManager...
type Allocator struct {
    pm      *PageManager
    offsets *[]int
    own     []int
}
```

Row seam at structure.go:763-772:

```go
func (sm *StructureManager) EmitRowCells(buf *bytes.Buffer, pageIndex, count int) (int, func()) {
    _ = buf // reserved: batch BDC emission stays a caller concern for now
    noop := func() {}
    if sm == nil || !sm.Enabled || count <= 0 {
        return 0, noop
    }
    base := sm.ReserveMCIDsLite(pageIndex, count)
    sm.beginTableRowWithTDMCIDs(pageIndex, base, count)
    return base, sm.EndStructureElement
}
```

Vector policy at vector/vector.go:57-60:

```go
func SetFill(b *bytes.Buffer, prec int, r, g, bl float64) {
    WriteFloats(b, prec, r, g, bl)
    b.WriteString(" rg\n")
}
```

## Key refs

- generation.go:44 newGeneration, 73 decodeImages, 179 layoutFontIDs, 275 emitImageXObjects
- generator.go:373 GenerateTemplatePDFBorrowed, 302 signatureContextAdapter
- pagemanager.go:132 NewPageManager, 395 ObjectAllocator, 404 allocator, 372 AllocObjectID, 377 AllocObjectIDs, 382 CommitExtraObject
- allocator.go:20 BindPageManager, 27 Next, 32 Alloc, 37 AllocN, 49 SeekTo, 56 EnsureBeyond, 63 Commit, 79 SetOffset, 96 LayoutContentFontIDs
- structure.go:111 StructureManager, 448 ReserveElementCapacity, 467 ReserveMCIDsLite
- html_convert.go:31 warnUnmappedHTMLOptions, 41 buildPDFDocument, 91 buildImageDocument, 122 parseMarginMM

## Validation

Per row: Page and Alloc tests, full internal/pdf plus test-verify-pdfs, typst plus svg plus pdf tests, HTML plus js wasm build, Table and Tag plus Structure tests, models plus gopdflib tests.

Closure phase 6, all checked: make fmt plus make lint plus make test passes with pytest 90 passed and 10/10 veraPDF plus structure tree checks. make test-integration passes in 16.6s. make test-verify-pdfs shows 10 passed, 0 failed. make wasm-compress plus frontend build passes with vite 9.55s and npm test 16/16.

Deferred: full sampledata split, Redact redesign, financial report dual purpose reroute, zerodha strict aval stays warn only.
