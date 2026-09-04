# Python builder parity

Date: 2026-09-04. Branch: feat/builder-snippets. Commit: 449200c.

pypdfsuit.builder is a pure Python overlay. It sinks through GeneratePDF over CGO. No engine change.

## What exists

From bindings/python/pypdfsuit/builder.py:

- make_props(font, size, bold, italic, underline, align, borders)
- Font class: Font(name).size(n).bold().italic().underline().left().center().right().borders(l,r,t,b).bordered().borderless(). Terminals props() and cell(text).
- Text class: Text(text).font(f).props(s).bg(color).fg(color).math(). Terminal build().
- new_cell, header_cell, math_cell
- set_cell_font, set_cell_alignment, set_cell_borders, set_cell_bg_color, set_cell_text_color, set_cell_color, set_row_color, set_table_colors, add_bracket_text, set_bracket_font
- image_from_path, font_from_path (file helpers, Python only)
- TableBuilder with add_row and build
- TemplateBuilder with add_title, add_table, add_spacer, build, generate

Re-exported in bindings/python/pypdfsuit/__init__.py:65-86.

## Usage

Font chain plus cell terminal:

```python
from pypdfsuit import Font
c = Font("Helvetica").size(18).bold().center().cell("hi")
assert c.props == "Helvetica:18:100:center:1:1:1:1"
```

Text with colors and math:

```python
from pypdfsuit import Font, Text
f = Font("Helvetica").size(12).bold()
c = Text("x^2").font(f).bg("#FF0000").fg("#00FF00").math().build()
```

Explicit props wins over font:

```python
from pypdfsuit import Font, Text, make_props, new_cell
f = Font("Helvetica").size(18).bold()
explicit = make_props("Courier", 10)
assert Text("t").font(f).props(explicit).build().props == explicit
assert new_cell("hi", explicit, font=f).props == explicit
```

Table flow from sampledata/pypdflib/builder-snippets/main.py:

```python
builder = TemplateBuilder(page="A4", portrait=True)
builder.add_title("Document Title", font="Helvetica", size=18, bold=True)
table = builder.add_table(2, 2, 1)
table.add_row(header_cell("Item"), header_cell("Price"))
amounts = table.add_row(
    new_cell("Total Revenue", make_props("Helvetica", 10, align="left", borders=(1, 1, 1, 1))),
    new_cell("$2,450,000", make_props("Helvetica", 10, align="right", borders=(1, 1, 1, 1))),
)
set_cell_text_color(amounts[1], "#B00020")
template = builder.build()
```

## Gaps vs Go

I checked Go builder.go and fontbuilder.go against builder.py. Gaps that matter:

- No AddTitleTable. Python sample assigns builder.title.table directly and touches private builder._elements.
- No TemplateBuilder.add_image or NewImageCell equivalent. Python has image_from_path helper but sample builds Element(type="image") by hand.
- No TitleFontOptions. Python add_title takes flat font/size/bold kwargs.
- No Cell() alias. Python only has build().
- No FontOpts or ParseFontOpts canonical grammar.
- Defaults differ. Python Font defaults to size 12 bordered. Go Font starts size 0 borderless and renders 12 through FontOpts.String. This changes default props strings.
- set_table_colors takes a built Table in Python (builder.py:313) and a TableBuilder pointer in Go (builder.go:323).
- SetCell helpers mutate and return Cell in Python for chaining. Go takes Cell pointer and returns nothing.
- set_cell_font in Python rewrites unconditionally. Go preserves current name on empty string and size on <= 0 through ParseFontOpts round trip.
- TemplateBuilder.generate returns bytes and raises. Go Generate returns ([]byte, error).

## Types and transport

Types live in bindings/python/pypdfsuit/types.py: Cell at 275, Row at 297, Table at 307, PDFTemplate at 430. Transport in _bindings.py has no builder code. CGO in bindings/python/cgo/exports.go exposes GeneratePDF, MergePDFs, SplitPDF, FillPDFWithXFDF, CompressPDF, ConvertHTMLToPDF, ConvertHTMLToImage, font and redact ops. Builder never adds a CGO entry.

## Tests

- bindings/python/tests/test_builder.py, includes generate test gated on libgopdfsuit.so
- bindings/python/tests/test_builder_fluent.py, parity pins for Font, Text, precedence
- Go mirrors: pkg/gopdflib/builder_test.go, fontbuilder_test.go
