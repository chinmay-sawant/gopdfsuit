# pypdfsuit API Reference

Python bindings for gopdfsuit. Generation, merge, split, fill, compress, HTML conversion, and redaction over the Go shared library.

Start here:

- [bindings/python/README.md](../bindings/python/README.md) - install, quick start, Props format
- [PY_BUILDER_PARITY.md](PY_BUILDER_PARITY.md) - Python builder vs Go builder parity, gaps
- [builder_fluent_sample.py](../sampledata/python/builder_fluent_sample.py) - runnable fluent-builder sample

Builder-first rule: prefer `TemplateBuilder` plus `Font` chains over hand-written Props strings or raw `PDFTemplate` dicts.

## Install

Build the shared library first, then install the wheel source:

```bash
cd bindings/python
./build.sh
pip install .
```

Requires Python 3.8 plus Go 1.26.4 for the build. No browser or Ghostscript needed.

## Generate

`generate_pdf(template)`, `generate_pdf_from_dict(d)`, `serialize_template(t)`, `get_available_fonts()` from `pypdfsuit.generator`.

Builder-first generation with `TemplateBuilder` and `Font` chains:

```python
from pypdfsuit.builder import TemplateBuilder, Font

b = TemplateBuilder("A4", True)
b.add_title("My Document", font="Helvetica", size=24, bold=True)
tb = b.add_table(2, 1.0, 1.0)
tb.add_row(Font("Helvetica").size(12).bold().cell("Name"),
            Font("Helvetica").size(12).cell("John Doe"))
pdf_bytes = b.generate()
```

Raw template path (low level, builder preferred). Uses real types `PDFTemplate`, `Config`, `Title`:

```python
from pypdfsuit import generate_pdf, PDFTemplate, Config, Title
from pypdfsuit.builder import make_props

template = PDFTemplate(
    config=Config(page="A4", page_alignment=1),
    title=Title(props=make_props("Helvetica", 24, bold=True,
        align="center", borders=(0, 0, 0, 0)), text="My Document"),
    elements=[],
)
pdf_bytes = generate_pdf(template)
```

Also available: `generate_pdf_from_dict(d)` for raw JSON-template parity, `serialize_template(t)` for UTF-8 JSON bytes, `get_available_fonts()` returning `list[FontInfo]`.

## Merge

`merge_pdfs(pdf_files)` from `pypdfsuit.merge`.

```python
from pypdfsuit import merge_pdfs

with open("doc1.pdf", "rb") as f1, open("doc2.pdf", "rb") as f2:
    pdfs = [f1.read(), f2.read()]
merged = merge_pdfs(pdfs)
with open("merged.pdf", "wb") as f:
    f.write(merged)
```

Raises `ValueError` when the list is empty or any entry is empty.

## Split

`split_pdf(pdf_data, spec)`, `parse_page_spec(spec, total_pages)` from `pypdfsuit.split`. Spec type is real request type `SplitSpec`.

```python
from pypdfsuit import split_pdf, SplitSpec

with open("document.pdf", "rb") as f:
    pdf_data = f.read()
spec = SplitSpec(pages=[1, 3, 5])
parts = split_pdf(pdf_data, spec)
for i, part in enumerate(parts):
    open(f"part_{i + 1}.pdf", "wb").write(part)
```

Page helper with real function `parse_page_spec`:

```python
from pypdfsuit import parse_page_spec

pages = parse_page_spec("1-3,5,7-9", 10)
print(pages)
```

`SplitSpec` fields: `pages: list[int] | None`, `ranges: list[tuple[int, int]] | None`, `max_per_file: int | None`.

## Fill

`fill_pdf_with_xfdf(pdf_data, xfdf_data)` from `pypdfsuit.fill`.

```python
from pypdfsuit import fill_pdf_with_xfdf

with open("form.pdf", "rb") as f:
    pdf_data = f.read()
with open("data.xfdf", "rb") as f:
    xfdf_data = f.read()
filled = fill_pdf_with_xfdf(pdf_data, xfdf_data)
open("filled.pdf", "wb").write(filled)
```

Raises `ValueError` when either input is empty.

## Compress

`compress_pdf(src, level)` from `pypdfsuit.compress`. Tiers: `"light"`, `"medium"`, `"heavy"`. Empty string selects the engine default.

```python
from pypdfsuit import compress_pdf

with open("input.pdf", "rb") as f:
    src = f.read()
compressed = compress_pdf(src, level="medium")
with open("compressed.pdf", "wb") as f:
    f.write(compressed)
```

Invalid level raises `ValueError`. Pure Go, no Ghostscript.

## HTML to PDF and Image

`convert_html_to_pdf(request)`, `convert_html_to_image(request)` from `pypdfsuit.html`. Request types are real: `HtmlToPDFRequest`, `HtmlToImageRequest`.

```python
from pypdfsuit import convert_html_to_pdf, HtmlToPDFRequest

request = HtmlToPDFRequest(
    html="<html><body><h1>Hello World</h1></body></html>",
    page_size="A4",
    orientation="Portrait",
)
pdf_bytes = convert_html_to_pdf(request)
open("out.pdf", "wb").write(pdf_bytes)
```

Image variant with real type `HtmlToImageRequest`:

```python
from pypdfsuit import convert_html_to_image, HtmlToImageRequest

request = HtmlToImageRequest(
    html="<html><body><h1>Hello</h1></body></html>",
    format="png",
    width=800,
    height=600,
)
img_bytes = convert_html_to_image(request)
open("out.png", "wb").write(img_bytes)
```

Either `html` or `url` must be set, else `ValueError`. Pure Go via gowkhtmltopdf, no browser.

## Redact

`get_page_info(pdf)`, `extract_text_positions(pdf, page)`, `find_text_occurrences(pdf, text)`, `apply_redactions(pdf, blocks)`, `apply_redactions_advanced(pdf, options)` from `pypdfsuit.redact`.

```python
from pypdfsuit import apply_redactions_advanced

with open("document.pdf", "rb") as f:
    pdf_data = f.read()
options = {"blocks": [{"pageNum": 1, "x": 120, "y": 620,
                        "width": 180, "height": 24}],
           "textSearch": [{"text": "Confidential"}],
           "mode": "visual_allowed"}
redacted = apply_redactions_advanced(pdf_data, options)
open("redacted.pdf", "wb").write(redacted)
```

Discovery flow with real functions `find_text_occurrences` and `apply_redactions`:

```python
from pypdfsuit import find_text_occurrences, apply_redactions

with open("document.pdf", "rb") as f:
    pdf_data = f.read()
boxes = find_text_occurrences(pdf_data, "secret")
redacted = apply_redactions(pdf_data, boxes)
open("redacted.pdf", "wb").write(redacted)
```

Helpers: `get_page_info(pdf)` returns `totalPages` plus page dims, `extract_text_positions(pdf, 1)` returns text plus coordinates for page 1.

## Builder

`Font`, `Text`, `make_props`, `new_cell`, `header_cell`, `math_cell`, setters, `TableBuilder`, `TemplateBuilder` from `pypdfsuit.builder`. See [PY_BUILDER_PARITY.md](PY_BUILDER_PARITY.md) and [builder_fluent_sample.py](../sampledata/python/builder_fluent_sample.py).

Font chain plus `Text` colors and math:

```python
from pypdfsuit.builder import Font, Text, TemplateBuilder

b = TemplateBuilder("A4", True)
b.add_title("Financial Summary", font="Helvetica", size=24, bold=True)
tb = b.add_table(2, 2.0, 1.0)
tb.add_row(Font("Helvetica").size(12).bold().center().cell("Metric"),
            Text("x^2").font(Font("Helvetica").size(12)).math().build())
pdf_bytes = b.generate()
```

Cell helpers with real names `header_cell`, `set_cell_text_color`, `make_props`:

```python
from pypdfsuit.builder import header_cell, make_props, set_cell_text_color

h = header_cell("Price")
props = make_props("Helvetica", 12, bold=True, align="center",
                   borders=(1, 1, 1, 1))
row = [h]
set_cell_text_color(row[0], "#B00020")
print(props, row[0].props)
```

File helpers: `image_from_path(path, width, height)` returns `Image` with base64 data, `font_from_path(name, path)` returns `CustomFontConfig`. Table flow: `tb = b.add_table(2, 2.0, 1.0)`, `cells = tb.add_row(c1, c2)`, `b.add_spacer(20)`, `b.build()` or `b.generate()`.

## Types

Real dataclasses from `pypdfsuit.types`: `PDFTemplate`, `Config`, `SecurityConfig`, `PDFAConfig`, `SignatureConfig`, `CustomFontConfig`, `Title`, `TitleTable`, `Table`, `Row`, `Cell`, `FormField`, `Image`, `Footer`, `Spacer`, `Element`, `Bookmark`, `FontInfo`, `HtmlToPDFRequest`, `HtmlToImageRequest`, `SplitSpec`.

```python
from pypdfsuit import PDFTemplate, Config, Title
from pypdfsuit.builder import make_props

template = PDFTemplate(
    config=Config(page="A4", page_alignment=1),
    title=Title(props=make_props("Helvetica", 18, bold=True,
        align="center", borders=(0, 0, 0, 0)), text="Hi"),
    elements=[],
)
print(template.to_dict()["config"]["page"])
print(template.to_dict()["title"]["text"])
```

Every type has `.to_dict()` with Go wire keys (`maxcolumns`, `columnwidths`, `bgcolor`, `textcolor`, `imagename`, `imagedata`).

## Errors

Real errors from `pypdfsuit._bindings`, re-exported at top level: `GoPDFSuitError`, `InvalidInputError`, `LimitExceededError`, `UpstreamError`, `InternalError`.

```python
from pypdfsuit import merge_pdfs, GoPDFSuitError

try:
    merged = merge_pdfs([open("a.pdf", "rb").read()])
except GoPDFSuitError as e:
    print(e.code, e.message)
    raise
print(len(merged))
```
