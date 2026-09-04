# Plans > Copy-Clip - Keep Snippet JSON + Builder Snippets

> **Parent:** `plans/copy-clip/plan.md` - canonical ledger (this file, converted to phase-wise-checklist shape)
> **Status:** planning, 0/5 phases complete, branch `feat/copy-clip-builder-snippets`
> **Estimated effort:** S-M, 2-3 sessions (Go overlay, Python parity, frontend copy buttons)

---

## Overview

Keep the user snippet JSON as canonical (`config + title + inline elements`), drop any Versa conversion path, and add thin builder/snippet helpers for fonts, brackets, and per-cell color. No engine draw changes in v1.

Source snippet shape: `config{pageBorder,pageMargin,page,pageAlignment,watermark,pdfTitle,pdfaCompliant,signature}`, `title{props,text,textprops,table{maxcolumns,columnwidths,rows}}`, `elements[{type:table/spacer}]`.

Prior investigation: 5 subagents (schema, gopdflib API, python parity, frontend builder, props codec). No `Versa` code exists outside `universal-translator` in `go.mod:56`. No `Builder/AddTable/SetFont/SetColor` in `pkg/gopdflib/*`.

## Executive Summary

- Canonical contract stays `internal/models/models.go:186-198` `PDFTemplate`, inline `elements` per `internal/pdf/generator.go:2273-2328`.
- Props grammar `Font:Size:Style:Align:L:R:T:B` parsed by `internal/pdf/utils.go:98-185 parseProps`, colors via `bgcolor/textcolor` hex, weights via `columnwidths`.
- Gap: `title.textprops` + `footer.props` + `embedStandardFonts` are frontend-only, dropped by Go. Fix with optional aliases.
- Builders are pure overlays emitting current `Props`/hex strings, sink stays `GeneratePDF(PDFTemplate)`.
- Frontend general builder is `/editor` (`frontend/src/pages/Editor.jsx`); reuse `documentModel.js:1-158` + `utils.js:5-65` for copy-snippet output.

## Phase 1: Contract correctness - aliases, no migration

### 1.1 Schema aliases (Go)

- [ ] `internal/models/models.go:299-312` - add optional `Title.TextProps string json:"textprops,omitempty"`, engine falls back `textprops || props`, preserves `Title.Table` precedence in `internal/pdf/draw.go:717-831` - proof: `go test ./internal/models -run TestSchema -v`
- [ ] `internal/models/models.go:412-418` - add optional `Footer.Props string json:"props,omitempty"`, engine `internal/pdf/draw.go:2043` prefers `Props || Font` - proof: `go test ./internal/pdf -run TestFooter -v`
- [ ] `internal/models/models.go:225-241` - add optional `Config.EmbedStandardFonts *bool json:"embedStandardFonts,omitempty"`, normalize with `EmbedFonts` in `frontend/src/components/editor/documentModel.js:116-127` parity - proof: `go test ./internal/models -v`
- [ ] `pkg/gopdflib/types.go:73,141,177` - mirror `TextProps`, `Footer.Props`, `EmbedStandardFonts` aliases, keep JSON wire identical - proof: `go test ./pkg/gopdflib -v`

### 1.2 Golden fixtures

- [ ] `sampledata/editor/financial_report.json` - add case with `title.textprops` set, `title.props` distinct, confirm Go output uses textprops fallback - proof: `make test-integration`
- [ ] `sampledata/python/amazonReceipt/amazonReceipt.py:372` - document `textprops` now honored, not dropped - proof: `bindings/python pytest tests/test_golden_template.py -v`
- [ ] `internal/models/schema_golden_test.go:1-80` - extend to reject unknown-field drift, allow new aliases only - proof: `go test ./internal/models -run TestGolden -v`

## Phase 2: Go builder overlay - fonts, brackets, right-cell color

### 2.1 Typed props core

- [ ] `pkg/gopdflib/props.go` - add `Align`, `Color`, `Borders`, `FontOpts{Name,Size,Bold,Italic,Underline,Align,Borders}` with `String() -> Helvetica:12:100:left:1:1:1:1` and `ParseFontOpts()`, no `internal/*` import - proof: `go test ./pkg/gopdflib -run TestProps -v`
- [ ] `pkg/gopdflib/doc.go:47` - document grammar `Font:Size:Style:Align:L:R:T:B`, note `000` regular vs `100` bold is style not color, color lives in `bgcolor/textcolor` - proof: `go vet ./pkg/gopdflib`

### 2.2 Builder functions

- [ ] `pkg/gopdflib/builder.go` - add `NewDocument(page,portrait)`, `AddTitle(text,opts...)`, `AddTable(maxCols,colWidths...)`, `AddSpacer(h)`, `Build() PDFTemplate`, `Generate() ([]byte,error)` - proof: `go test ./pkg/gopdflib -run TestBuilder -v`
- [ ] `pkg/gopdflib/builder.go` - add `MakeProps`, `NewCell`, `HeaderCell`, `MathCell`, `SetCellFont`, `SetCellAlignment`, `SetCellBorders` - proof: `go test ./pkg/gopdflib -run TestCellFont -v`
- [ ] `pkg/gopdflib/builder.go` - add `SetCellColor`, `SetCellTextColor`, `SetCellBgColor`, `SetRowColor`, `SetTableColors` for different-color-on-right - proof: snippet below renders red right cell, `go test ./pkg/gopdflib -v`
- [ ] `pkg/gopdflib/builder.go` - add `AddBracketText(c,open,close)`, `SetBracketFont(c,font,size)` v1 without `Cell.Segments` rich-text - proof: `go test ./pkg/gopdflib -run TestBracket -v`
- [ ] `pkg/gopdflib/example_test.go` - add `TestGeneratePDFProgrammatically` extension covering 3-col title table, empty spacer rows, centered image cell `width:100 height:80` from snippet - proof: `make test-verify-pdfs`

Reference snippet (must compile after Phase 2):

```go
b := gopdflib.NewDocument("A4", true)
b.AddTitle("Document Title", gopdflib.WithTitleFont("Helvetica", 18, true))
tb := b.AddTable(3, 1, 2, 1)
row := tb.AddRow(
  gopdflib.NewCell("", "Helvetica:12:000:left:1:1:1:1"),
  gopdflib.NewCell("Document Title", "Helvetica:18:100:center:1:1:1:1"),
  gopdflib.NewCell("", "Helvetica:12:000:right:1:1:1:1"),
)
gopdflib.SetCellTextColor(&row[2], "#B00020")
gopdflib.SetCellFont(&row[1], "Helvetica", 18, true, false, false)
c := gopdflib.NewCell("clause", gopdflib.MakeProps("Helvetica", 12, false, false, false, "left", [4]int{1,1,1,1}))
gopdflib.AddBracketText(&c, "[", "]")
pdfBytes, err := b.Generate()
```

## Phase 3: Python parity - builder plus compress wiring

### 3.1 Builder module

- [ ] `bindings/python/pypdfsuit/builder.py` - add `make_props`, `new_cell`, `header_cell`, `set_cell_font`, `set_cell_color`, `add_bracket_text` mirroring Go names - proof: `pytest bindings/python/tests/test_builder.py -v`
- [ ] `bindings/python/pypdfsuit/builder.py` - add `TemplateBuilder` with `add_title`, `add_table`, `add_spacer`, `build()->PDFTemplate`, `generate()->bytes` - proof: `pytest bindings/python/tests/test_builder.py -v`
- [ ] `bindings/python/pypdfsuit/types.py:311-355` - add `image_from_path(path)->Image` (read+base64) and `font_from_path` helper, no behavior change to `Cell.props` wire - proof: `pytest bindings/python/tests/test_builder.py -v`
- [ ] `bindings/python/pypdfsuit/generator.py:12-18` - add `generate_pdf_from_dict(d:dict)->bytes` for raw JSON-template parity used in `tests/test_integration.py:52-56` - proof: `pytest bindings/python/tests/test_integration.py -v`

### 3.2 Compress gap

- [ ] `bindings/python/pypdfsuit/compress.py` - new `compress_pdf(src,level)` via CGO `bindings/python/cgo/exports.go:310-322`, wire `bindings/python/pypdfsuit/_bindings.py:157-222` + `__init__.py:72-116` - proof: `pytest bindings/python/tests/test_compress.py -v` no longer skips

## Phase 4: Frontend copy-clip - reuse general builder

### 4.1 Snippet export

- [ ] `frontend/src/components/editor/utils.js:31-50` - reuse `parseProps/formatProps` to emit `MakeProps/SetCellColor/AddBracketText` strings, no new codec - proof: `cd frontend && npm run lint`
- [ ] `frontend/src/pages/Editor.jsx:632-640` + `frontend/src/components/editor/Toolbar.jsx:1-202` + `frontend/src/components/documentation/content/template-format.js:30-50` - add Copy Go snippet and Copy Python snippet beside Copy JSON (`JsonTemplate.jsx:18-26`) - proof: `cd frontend && npm run build`
- [ ] `frontend/src/components/editor/PropertiesPanel.jsx:869-1465` - snippet reflects selected cell `props/text/bgcolor/textcolor/link` and table `columnwidths`, spacer `height 1-200`, image `width/height/imagedata` - proof: manual editor check + `npm run build`

## Phase 5: Docs and closure gates

- [ ] `documentation/TEMPLATE_REFERENCE.md:60,86,537` - document keep-snippet decision, no Versa path, `template-data` + `GenerateTemplatePDFBorrowed` as load path - proof: docs-only, no lint/test per skill
- [ ] `pkg/gopdflib/doc.go:29-56` - document builder overlay with bracket and right-color examples - proof: docs-only
- [ ] Full gates once - `make fmt && make lint && make test` - proof: paste output in handoff
- [ ] Integration gate - `make test-integration` - proof: paste output
- [ ] PDF compliance spot - `make test-verify-pdfs` on snippet-shaped PDF (3-col title, spacer 20, image 100x80) - proof: `test/verify_pdfs.sh` + veraPDF + structure tree check

## Dependencies

- Phase 2 depends on Phase 1 aliases (builder emits `textprops`-safe JSON).
- Phase 3 depends on Phase 2 naming (Python mirrors Go `MakeProps/SetCellColor/AddBracketText`).
- Phase 4 depends on Phase 2 string format (frontend emits identical `Props` grammar).
- Phase 5 depends on Phases 1-4 gates.

## Non-goals

- No Versa importer.
- No `Cell.Segments` rich-text in v1, bracket v1 is `AddBracketText` plus font override.
- No engine draw changes, builder emits current `Props`/`bgcolor` strings.
- No hand-edit `docs/` Vite output, edit `frontend/src/` and rebuild.

## Completion Handoff

- [ ] Confirm rows above with commands beside each gate, synchronize this ledger, report next unchecked phase. Do not create a second status document.
