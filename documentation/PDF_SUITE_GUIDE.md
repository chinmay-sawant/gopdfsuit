# PDF Suite Guide (Go plus Python plus REST)

One detailed entry point for the whole suite. Existing docs are fragmented:
`README.md` overviews, `documentation/FEATURES.md` lists ops, `documentation/GETTING_STARTED_GOPDFLIB.md`
covers Go install only, `bindings/python/README.md` covers Python only, and PyPI renders
only the Python README with zero links to `documentation/`. This file ties them together
with copy-paste samples in all three surfaces.

Scope: Go module `github.com/chinmay-sawant/gopdfsuit/v7`, Python `pypdfsuit==7.0.1`,
Go 1.26.4, pure-Go HTML via `gowkhtmltopdf v0.2.5`, no browser, no Ghostscript.

## 1. Pick your surface

| Stack | Use | Docs |
|---|---|---|
| REST service | `POST /api/v1/*` on :8080, Docker `chinmaysawant/gopdfsuit:7.0.0` | `documentation/FEATURES.md`, `documentation/TEMPLATE_REFERENCE.md` |
| Go library | `pkg/gopdflib` | `documentation/GETTING_STARTED_GOPDFLIB.md`, `documentation/BUILDER_FLUENT_GO.md` |
| Python bindings | `pypdfsuit` via CGO `libgopdfsuit.so` | `bindings/python/README.md`, `documentation/PY_BUILDER_PARITY.md` |
| Browser local | `gopdfsuit.wasm` full engine plus `compress.wasm` worker | `documentation/WASM_VIEWER_EDITOR.md`, `documentation/FRONTEND_WASM_SPLIT.md` |

Full map: `documentation/index.md`. Template schema: `documentation/TEMPLATE_REFERENCE.md`.

## 2. Prerequisites

- Go 1.26.4 or later for the Go library and server.
- Python 3.8+ for `pypdfsuit`. Build the native lib first (see section 4).
- Node 18+ for the frontend. Java 11+ only if you run veraPDF compliance checks.
- No Chrome, no Ghostscript anywhere in v7.

## 3. Install the Go library

```bash
go get github.com/chinmay-sawant/gopdfsuit/v7@v7.0.0
```

```go
import "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
```

## 4. Install the Python bindings

```bash
cd bindings/python
./build.sh
pip install .
```

This builds `pypdfsuit/lib/libgopdfsuit.so` (or `.dylib` / `.dll`) and installs `pypdfsuit`.
Requires Go 1.26.4+ at build time. No browser needed.

PyPI page renders only `bindings/python/README.md`. For template schema, compliance,
and HTML limits, follow the Further reading links in section 12.

## 5. Generate your first PDF

Preferred spelling is fluent builders. Raw `Props` strings
(`Name:Size:Style:Align:L:R:T:B`) stay supported as the low-level form.

Go (`sampledata/gopdflib/builder-snippets/main.go`):

```go
package main

import (
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    b := gopdflib.NewDocument("A4", true)
    b.AddTitle("My Document", gopdflib.WithTitleFontOpts(gopdflib.TitleFontOptions{Name: "Helvetica", Size: 24, Bold: true}))
    tb := b.AddTable(2, 1.0, 1.0)
    tb.AddRow(
        gopdflib.Font("Helvetica").Size(12).Bold().Cell("Name"),
        gopdflib.Font("Helvetica").Size(12).Cell("John Doe"),
    )
    pdfBytes, err := b.Generate()
    if err != nil {
        panic(err)
    }
    _ = os.WriteFile("out.pdf", pdfBytes, 0644)
}
```

```bash
cd sampledata && go run ./gopdflib/builder-snippets
```

Python (`sampledata/python/builder_fluent_sample.py`):

```python
from pypdfsuit.builder import Font, TemplateBuilder

b = TemplateBuilder("A4", True)
b.add_title("My Document", font="Helvetica", size=24, bold=True)
tb = b.add_table(2, 1.0, 1.0)
tb.add_row(
    Font("Helvetica").size(12).bold().cell("Name"),
    Font("Helvetica").size(12).cell("John Doe"),
)
pdf_bytes = b.generate()
with open("out.pdf", "wb") as f:
    f.write(pdf_bytes)
```

```bash
PYTHONPATH=bindings/python python3 sampledata/python/builder_fluent_sample.py
```

REST:

```bash
curl -X POST localhost:8080/api/v1/generate/template-pdf \
  -H 'Content-Type: application/json' --data-binary @template.json -o out.pdf
```

Equivalence proof: `Font("Helvetica").size(12).bold().center().bordered().cell("Name")`
emits `Cell(props="Helvetica:12:100:center:1:1:1:1")`. Same bytes either way.

## 6. Props grammar cheat sheet

One string per cell or title: `FontName:FontSize:StyleCode:Alignment:L:R:T:B`.

- StyleCode is 3 bits for bold, italic, underline. `100` is bold only, `000` is plain.
- Alignment is `left`, `center`, or `right`. Unknown falls back to `left`.
- Borders are `L:R:T:B` flags. Color never lives in props. Use `bgcolor` and
  `textcolor` hex fields on the cell.
- Helpers: Go `MakeProps` plus `ParseFontOpts`, Python `make_props`.
  Prefer `Font` chains; hand-write strings only when copying JSON fixtures verbatim.

## 7. Compress (no Ghostscript)

Tiers: Light (JPEG 92, 1920px), Medium (75, 1275px, default), Heavy (50, 612px).
Input cap 32 MiB. Returns the input unchanged when output would not shrink,
when the file has more than 2 object streams, or when `/Root` is missing.
Encrypted PDFs are rejected.

Go:

```go
out, err := gopdflib.CompressPDF(pdfBytes, gopdflib.CompressOptions{Level: gopdflib.Medium})
```

Python:

```python
from pypdfsuit import compress_pdf
small = compress_pdf(pdf_bytes, level="medium")
```

REST:

```bash
curl -X POST localhost:8080/api/v1/compress \
  -H 'Content-Type: application/pdf' --data-binary @in.pdf -o small.pdf
```

Note: Go and HTTP accept `JPEGQuality` and `MaxImageDim` overrides (clamped at 100
and 4096). Python sends level only. Unknown levels map to Medium on Go, HTTP, and
WASM; Python raises `ValueError`.

## 8. Merge and split

Go:

```go
merged, err := gopdflib.MergePDFs([][]byte{a, b})
parts, err := gopdflib.SplitPDF(pdfBytes, gopdflib.SplitSpec{})
pages, err := gopdflib.ParsePageSpec("1-3,5", 10)
```

Python:

```python
from pypdfsuit import merge_pdfs, split_pdf, parse_page_spec
merged = merge_pdfs([a, b])
pages = parse_page_spec("1-3,5", total_pages=10)
```

REST: `POST /api/v1/merge`, `POST /api/v1/split`.

## 9. HTML to PDF and image (pure-Go)

No browser. Static content only. `DPI` and `LowQuality` are accepted and ignored
with a warning. `htmltoimage format=svg` is rejected; use `png` or `jpg`.

Go:

```go
pdf, err := gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{HTML: "<h1>Hi</h1>", PageSize: "A4"})
img, err := gopdflib.ConvertHTMLToImage(gopdflib.HTMLToImageRequest{HTML: "<h1>Hi</h1>", Format: "png"})
```

Python:

```python
from pypdfsuit import convert_html_to_pdf, convert_html_to_image, HtmlToPDFRequest, HtmlToImageRequest
pdf = convert_html_to_pdf(HtmlToPDFRequest(html="<h1>Hi</h1>", page_size="A4"))
png = convert_html_to_image(HtmlToImageRequest(html="<h1>Hi</h1>", format="png"))
```

WASM takes inline HTML strings only for full local mode; URL fetch stays server-side.

## 10. Redact and fill

Python redact (all five ops exist; start here):

```python
from pypdfsuit import find_text_occurrences, apply_redactions_advanced
hits = find_text_occurrences(pdf_bytes, "Confidential")
out = apply_redactions_advanced(pdf_bytes, {"blocks": [{"pageNum": 1, "x": 120, "y": 620, "width": 180, "height": 24}], "mode": "visual_allowed"})
```

Go redact: `GetPageInfo`, `ExtractTextPositions`, `FindTextOccurrences`,
`ApplyRedactions`, `ApplyRedactionsAdvanced`, `ApplyRedactionsAdvancedWithReport`,
`AnalyzePageCapabilities`. Full flow in `documentation/GETTING_STARTED_GOPDFLIB.md`.

Fill (XFDF):

```go
filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
```

```python
from pypdfsuit import fill_pdf_with_xfdf
filled = fill_pdf_with_xfdf(pdf_bytes, xfdf_bytes)
```

Note: XFDF fill on compressed object streams needs the pdfcpu path; the byte
approach does not apply there.

## 11. Privacy, offline, and limits

- Default path is browser local. `compress.wasm` (~8M) handles Compress;
  `gopdfsuit.wasm` (~31M) handles the rest. Server upload needs explicit consent
  in `ConsentBanner`. No silent upload.
- First visit downloads WASM plus fonts, then works offline. Bundled templates work
  offline; GitHub fetch and server fallback do not.
- Caps: template JSON 8 MiB, HTML JSON 2 MiB, PDF uploads 32 MiB, fonts 10 MiB,
  XFDF 8 MiB. Compress input 32 MiB. Shared cache TTL `GOPDFSUIT_CACHE_TTL`
  defaults to 3 minutes.
- Compliance: PDF/A-4 opt-in plus PDF/UA-2 tagging. Validate with
  `test/verify_pdfs.sh` (veraPDF hard gate plus structure-tree checks).

## 12. Further reading

- Template schema: `documentation/TEMPLATE_REFERENCE.md`
- Go install plus JSON load: `documentation/GETTING_STARTED_GOPDFLIB.md`
- Go builders: `documentation/BUILDER_FLUENT_GO.md`
- Python parity plus gaps: `documentation/PY_BUILDER_PARITY.md`
- Python package: `bindings/python/README.md`
- Samples: `sampledata/gopdflib/builder-snippets/main.go`,
  `sampledata/python/builder_fluent_sample.py`,
  `sampledata/python/financial_report_pypdfsuit.py`
- HTML limits: `documentation/HTML_CONVERSION.md`, `documentation/HTML_PUREGO_SWAP.md`
- WASM: `documentation/WASM_VIEWER_EDITOR.md`, `documentation/FRONTEND_WASM_SPLIT.md`
- Compliance: `documentation/COMPLIANCE_PIPELINE_TODAY.md`, `documentation/PDF_VALIDATORS.md`
- Signatures: `documentation/DIGITAL_SIGNATURE_RSA_ECDSA.md`
- Full map: `documentation/index.md`
