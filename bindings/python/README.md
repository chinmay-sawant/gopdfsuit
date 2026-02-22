# pypdfsuit

Python bindings for [gopdfsuit](https://github.com/chinmay-sawant/gopdfsuit) - a comprehensive PDF library for generation, merging, splitting, form filling, and HTML to PDF/Image conversion.

## Features

- **PDF Generation**: Create PDFs from structured templates with tables, images, and styled text
- **PDF Merging**: Combine multiple PDFs into a single document
- **PDF Splitting**: Split PDFs by pages, ranges, or maximum pages per file
- **Form Filling**: Fill PDF forms using XFDF data
- **HTML to PDF**: Convert HTML content or URLs to PDF documents
- **HTML to Image**: Convert HTML content or URLs to images (PNG, JPG, SVG)
- **PDF Redaction**: Securely redact sensitive information using coordinates or text search

## Installation

### From Source

1. Build the shared library:

```bash
cd bindings/python
chmod +x build.sh
./build.sh
```

2. Install the Python package:

```bash
pip install .
```

### Requirements

- Python 3.8+
- Go 1.22+ (for building the shared library)
- Chrome/Chromium (for HTML to PDF/Image conversion)

## Quick Start

### Generate a PDF

```python
from pypdfsuit import generate_pdf, PDFTemplate, Config, Title, Element, Table, Row, Cell

template = PDFTemplate(
    config=Config(page="A4", page_alignment=1),
    title=Title(
        props="Helvetica:24:100:center:0:0:0:0",
        text="My Document"
    ),
    elements=[
        Element(
            type="table",
            table=Table(
                max_columns=2,
                column_widths=[1.0, 1.0],
                rows=[
                    Row(row=[
                        Cell(props="Helvetica:12:100:left:1:1:1:1", text="Name"),
                        Cell(props="Helvetica:12:000:left:1:1:1:1", text="John Doe"),
                    ])
                ]
            )
        )
    ]
)

pdf_bytes = generate_pdf(template)
with open("output.pdf", "wb") as f:
    f.write(pdf_bytes)
```

### Merge PDFs

```python
from pypdfsuit import merge_pdfs

with open("doc1.pdf", "rb") as f1, open("doc2.pdf", "rb") as f2:
    merged = merge_pdfs([f1.read(), f2.read()])

with open("merged.pdf", "wb") as f:
    f.write(merged)
```

### Split a PDF

```python
from pypdfsuit import split_pdf, SplitSpec

with open("document.pdf", "rb") as f:
    pdf_data = f.read()

# Split specific pages
spec = SplitSpec(pages=[1, 3, 5])
parts = split_pdf(pdf_data, spec)

# Or split every 5 pages
spec = SplitSpec(max_per_file=5)
parts = split_pdf(pdf_data, spec)

for i, part in enumerate(parts):
    with open(f"part_{i+1}.pdf", "wb") as f:
        f.write(part)
```

### Convert HTML to PDF

```python
from pypdfsuit import convert_html_to_pdf, HtmlToPDFRequest

# Convert HTML string
request = HtmlToPDFRequest(
    html="<html><body><h1>Hello World</h1></body></html>",
    page_size="A4",
    orientation="Portrait",
)
pdf_bytes = convert_html_to_pdf(request)

# Or convert a URL
request = HtmlToPDFRequest(
    url="https://example.com",
    page_size="Letter",
)
pdf_bytes = convert_html_to_pdf(request)
```

### Fill a PDF Form

```python
from pypdfsuit import fill_pdf_with_xfdf

with open("form.pdf", "rb") as f:
    pdf_data = f.read()
with open("data.xfdf", "rb") as f:
    xfdf_data = f.read()

filled = fill_pdf_with_xfdf(pdf_data, xfdf_data)
with open("filled.pdf", "wb") as f:
    f.write(filled)
```

### Redact a PDF

```python
from pypdfsuit import apply_redactions_advanced

with open("document.pdf", "rb") as f:
    pdf_data = f.read()

redacted = apply_redactions_advanced(pdf_data, {
    "blocks": [
        {"pageNum": 1, "x": 120, "y": 620, "width": 180, "height": 24}
    ],
    "textSearch": [
        {"text": "Confidential"}
    ],
    "mode": "visual_allowed"
})

with open("redacted.pdf", "wb") as f:
    f.write(redacted)
```

## API Reference

### Types

- `PDFTemplate` - Main template structure for PDF generation
- `Config` - Page configuration (size, orientation, security, etc.)
- `Title` - Document title section
- `Table`, `Row`, `Cell` - Table structure
- `Element` - Generic element (table, spacer, image)
- `Image`, `Spacer` - Additional elements
- `SecurityConfig` - Encryption settings
- `PDFAConfig` - PDF/A compliance settings
- `SignatureConfig` - Digital signature settings
- `HtmlToPDFRequest` - HTML to PDF conversion options
- `HtmlToImageRequest` - HTML to image conversion options
- `SplitSpec` - PDF split specification
- `FontInfo` - Font information

### Functions

- `generate_pdf(template: PDFTemplate) -> bytes`
- `get_available_fonts() -> List[FontInfo]`
- `merge_pdfs(pdf_files: List[bytes]) -> bytes`
- `split_pdf(pdf_data: bytes, spec: SplitSpec) -> List[bytes]`
- `parse_page_spec(spec: str, total_pages: int = 0) -> List[int]`
- `fill_pdf_with_xfdf(pdf_data: bytes, xfdf_data: bytes) -> bytes`
- `convert_html_to_pdf(request: HtmlToPDFRequest) -> bytes`
- `convert_html_to_image(request: HtmlToImageRequest) -> bytes`
- `get_page_info(pdf_data: bytes) -> dict`
- `extract_text_positions(pdf_data: bytes, page_num: int) -> list[dict]`
- `find_text_occurrences(pdf_data: bytes, text: str) -> list[dict]`
- `apply_redactions(pdf_data: bytes, redactions: list[dict]) -> bytes`
- `apply_redactions_advanced(pdf_data: bytes, options: dict) -> bytes`

## Props String Format

The props string format for cells and titles is:

```
FontName:FontSize:StyleCode:Alignment:BorderLeft:BorderRight:BorderTop:BorderBottom
```

- **FontName**: Helvetica, Courier, Times-Roman, etc.
- **FontSize**: Integer size in points
- **StyleCode**: 3 digits for bold(1/0), italic(1/0), underline(1/0). e.g., "100" = bold only
- **Alignment**: left, center, right
- **Borders**: 1 = border, 0 = no border

Example: `"Helvetica:12:100:center:1:1:1:1"` = Helvetica 12pt, bold, centered, all borders

## License

MIT License - see [LICENSE](../../LICENSE) for details.
