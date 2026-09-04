# Editor snippets and typst

Date: 2026-09-04. Branch: feat/builder-snippets. Commits 4427b94 copy clip plan, 449200c fluent builder, plus typstsyntax/renderer.go state.

## Snippet JSON shape

Canonical shape is config plus title plus inline elements. From TEMPLATE_REFERENCE.md:37, documentModel.js:79-112, snippet.json:1-86, snippet.js:160-173.

```json
{
  "config": { "page": "A4" },
  "title": { "props": "...", "textprops": "...", "text": "...", "table": { "maxcolumns": 3, "columnwidths": [1,2,1], "rows": [{"row": [{"props":"...","text":"..."}]}] } },
  "elements": [
    { "type": "spacer", "spacer": { "height": 20 } },
    { "type": "image", "image": { "imagename": "", "imagedata": "<base64>", "width": 100, "height": 80 } },
    { "type": "table", "table": { "maxcolumns": 2, "columnwidths": [2,1], "rows": [{"row": [{"props":"...","text":"..."}]}] } }
  ],
  "footer": { "font": "...", "props": "...", "text": "..." }
}
```

Rules:

- wrapComponent at documentModel.js:79 wraps into type table or spacer or image objects.
- buildTemplate at 100 maps components through wrapComponent into elements, deletes title or footer when null, deletes bookmarks when empty.
- buildTemplateJson at 114 is JSON.stringify with indent 2.
- normalizeElements at snippet.js:168 prefers template.elements, falls back to template.components.
- No Versa conversion, no importer per TEMPLATE_REFERENCE.md:39.
- Props grammar Font:Size:Style:Align:L:R:T:B, color in bgcolor and textcolor hex.
- Aliases: title.textprops beats title.props, footer.props beats footer.font, config.embedStandardFonts beats config.embedFonts.
- parseProps at editor/utils.js:31 defaults to Helvetica:12:000:left:0:0:0:0.

## Copy clip flows

Two separate clipboards.

Text snippet copy for Go, Python, JSON:

- Editor.jsx:558 builds goSnippet and pythonSnippet with useMemo from buildTemplate output.
- Editor.jsx:548 handleCopyJson, Go, Python call navigator.clipboard.writeText then setCopiedId to go, python, or json with 2s timeout reset.
- Toolbar.jsx:4 and 190 takes onCopyJSON, onCopyGo, onCopyPython, copiedId. Buttons Copy, Copy Go, Copy Py show Copied check when matched.
- Editor.jsx:670 wires Toolbar handlers.
- JsonTemplate.jsx:4 and 18 holds same copyText plus timeout pattern with three buttons. Helper at 103 says Copy Go and Copy Py emit builder snippets for the current template.
- Generators at snippet.js:193 templateToGoSnippet and 223 templateToPythonSnippet.

Structural copy and paste for canvas elements, not text:

- Editor.jsx:69 copiedId is button feedback, distinct from clipboard with type plus data.
- handleCopy at 212 finds element by id and stores structuredClone in clipboard.
- handleCut at 220 is copy plus delete.
- handlePaste at 225 clones clipboard data, guards title and footer singletons, else pasteComponentAt.
- handleDuplicate at 242 rejects title and footer, else pasteComponentAt after id.
- Wired to ContextMenu and useEditorShortcuts at 437, 621, 1147.

## Template model fields

From internal/models/models.go:

- PDFTemplate:187-198 with Config 188, Title 189, Table legacy 190, Spacer 191, Image 192, Elements 193, Footer 194, Bookmarks 195
- Element:216-222 with Type 217, Index 218, Table 219, Spacer 220, Image 221
- Config:225-244 with PageBorder 226, PageMargin 227, Page 228, PageAlignment 229, Watermark 230, PdfTitle 231, Security 234, PDFA 235, Signature 236, EmbedFonts 237, PDFACompliant 242, TaggedPDF 243
- Title:303-319 with Props 304, Text 305, TextProps 309, Table 312, BgColor 314, TextColor 316, Link 318
- TitleTable:322-326, Table:329-353 with MaxColumns 330, Rows 331, ColumnWidths 335, RowHeights 338, SharedRowLayout 349
- Row:356-358, Cell:361-397 with Props 362, Text 363, Checkbox 364, Image 365, Width 370, Height 371, FormField 372, BgColor 376, TextColor 380, Link 384, Wrap 388, Dest 392, MathEnabled 396
- Spacer:211-213, Image:410-417, Footer:420-428, Bookmark:201-208, Props:431-440

## Go and Python snippets

Cell from snippet.js:37-52 and 78-101. Input text [clause] with right aligned props and red textcolor:

```go
c := gopdflib.NewCell("[clause]", gopdflib.MakeProps("Helvetica", 12, false, false, false, "right", [4]int{1, 1, 1, 1}))
gopdflib.SetCellTextColor(&c, "#B00020")
gopdflib.AddBracketText(&c, "[", "]")
```

```python
c = new_cell("[clause]", make_props("Helvetica", 12, False, False, False, "right", (1, 1, 1, 1)))
set_cell_text_color(c, "#B00020")
add_bracket_text(c, "[", "]")
```

Table from snippet.js:103-158:

```go
tb := b.AddTable(2, 2, 1)
row1 := tb.AddRow(
	gopdflib.NewCell("Total Revenue", gopdflib.MakeProps("Helvetica", 10, false, false, false, "left", [4]int{1, 1, 1, 1})),
	gopdflib.NewCell("$2,450,000", gopdflib.MakeProps("Helvetica", 10, false, false, false, "right", [4]int{1, 1, 1, 1}))
)
gopdflib.SetCellTextColor(&row1[1], "#B00020")
```

```python
tb = b.add_table(2, 2, 1)
row1 = tb.add_row(
    new_cell("Total Revenue", make_props("Helvetica", 10, False, False, False, "left", (1, 1, 1, 1))),
    new_cell("$2,450,000", make_props("Helvetica", 10, False, False, False, "right", (1, 1, 1, 1)))
)
set_cell_text_color(row1[1], "#B00020")
```

Template Go from snippet.js:193-221:

```go
b := gopdflib.NewDocument("A4", true)
b.AddTitle("Document Title", gopdflib.WithTitleFont("Helvetica", 18, true))
b.AddSpacer(20)
pdfBytes, err := b.Generate()
```

Notes: clampSpacerHeight at 21 limits 1 to 200 default 20. clampImageDim at 27 clamps image dims. imagedata omitted with char count at 74, 213, 243. Footer emitted as comment.

Template Python from snippet.js:223-251:

```python
b = TemplateBuilder("A4", True)
b.add_title("Document Title", "Helvetica", 18, True)
b.add_spacer(20)
pdf_bytes = b.generate()
```

## Typst math cells

Cell.MathEnabled bool at models.go:396 enables Typst math parse. Renderer at typstsyntax/renderer.go:

- RenderContext 12, MathLayout 27, MathElement 35, ElementType 56 with glyph, line, group, offset, path kinds.
- LayoutEngine 68, layoutNode at 78 dispatches literal, symbol, superscript, subscript, fraction, sqrt, root, group, accent, matrix, vector, binom, cases, cancel, LR, prime, underover, op, stretch, sequence, func.
- RenderToContentStream at 1140, renderElements at 1147 emits BT, Tf, Td, Tj, ET for glyphs, StrokeLine for lines, StrokePath for paths, recursion for groups.
- Brackets: makeSquareBracketLeft and Right at 1262, makeParenLeft and Right at 1295 with 20 paren segments, 0.4 bracket line width, 0.15 paren line width.

End to end: snippet.json to Go to Python to POST /api/v1/generate/template-pdf or GET /api/v1/template-data?file=. Validation through validateTemplate errors and warnings for props, margin patterns, maxcolumns mismatch, image clamps.
