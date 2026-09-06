# Features

gopdfsuit is a template based PDF platform. You send JSON plus base64 assets, the Go engine renders PDF bytes. Same engine ships as a REST API on :8080, a Go library in pkg/gopdflib, Python bindings in pypdfsuit, and WASM modules that run in the browser.

## Template PDF generation

The core op. POST a PDFTemplate to /api/v1/generate/template-pdf, or call it from Go or Python. Layout covers titles, tables with shared row layout, spacers, images, footers, bookmarks, watermarks, page borders, links, form fields, Typst math cells.

```bash
curl -X POST localhost:8080/api/v1/generate/template-pdf \
  -H 'Content-Type: application/json' -d @template.json -o out.pdf
```

```python
from pypdfsuit.builder import Font, TemplateBuilder
b = TemplateBuilder("A4", True)
b.add_title("My Document", font="Helvetica", size=18, bold=True)
tb = b.add_table(2, 1.0, 1.0)
tb.add_row(Font("Helvetica").size(12).bold().cell("Name"), Font("Helvetica").size(12).cell("John Doe"))
pdf_bytes = b.generate()
```

Props grammar for every cell and title: Name:Size:Style:Align:L:R:T:B, for example Helvetica:12:100:left:1:1:1:1. Colors live in bgcolor and textcolor hex fields, never in props. Full shape lives in documentation/TEMPLATE_REFERENCE.md. Handy aliases: title.textprops beats title.props, footer.props beats footer.font.

## Fluent builders

Go overlay in pkg/gopdflib/builder.go plus fontbuilder.go plus props.go. Python mirror in bindings/python/pypdfsuit/builder.py. Both sink through the same generate call, output is byte identical to hand written props strings.

```go
b := gopdflib.NewDocument("A4", true)
b.AddTitle("Document Title", gopdflib.WithTitleFontOpts(gopdflib.TitleFontOptions{Name: "Helvetica", Size: 18, Bold: true}))
row := b.AddTable(2, 2, 1).AddRow(
  gopdflib.Font("Helvetica").Size(10).Cell("Total Revenue"),
  gopdflib.Font("Helvetica").Size(10).Right().Cell("$2,450,000"))
gopdflib.SetCellTextColor(&row[1], "#B00020")
pdfBytes, err := b.Generate()
```

```python
from pypdfsuit.builder import Font, TemplateBuilder, set_cell_text_color
b = TemplateBuilder(page="A4", portrait=True)
t = b.add_table(2, 2, 1)
row = t.add_row(Font("Helvetica").size(10).cell("Total"), Font("Helvetica").size(10).cell("$5"))
set_cell_text_color(row[1], "#B00020")
pdf_bytes = b.generate()
```

Details and parity gaps in documentation/BUILDER_FLUENT_GO.md and documentation/PY_BUILDER_PARITY.md.

## Merge, split, compress

Merge takes N PDFs in order. Split takes page specs like 1-3,5 or chunks with max_per_file. Compress recompresses images per tier and strips metadata. No Ghostscript anywhere.

```go
merged, _ := gopdflib.MergePDFs([][]byte{pdf1, pdf2})
parts, _ := gopdflib.SplitPDF(pdf, gopdflib.SplitSpec{Pages: []int{1, 3, 5}})
out, _ := gopdflib.CompressPDF(pdf, gopdflib.CompressOptions{Level: gopdflib.CompressHeavy})
```

```bash
curl -X POST localhost:8080/api/v1/merge -F pdf=@a.pdf -F pdf=@b.pdf -o merged.pdf
curl -X POST localhost:8080/api/v1/split -F pdf=@doc.pdf -F pages="1-3,5" -o splits.zip
curl -X POST localhost:8080/api/v1/compress -F pdf=@doc.pdf -F level=heavy -o compressed.pdf
```

Tiers: Light keeps JPEG 92 up to 1920px, Medium JPEG 75 up to 1275px and is the default, Heavy JPEG 50 down to 612px. Cap per input is 32 MiB and 50k objects. Merge and split refuse encrypted PDFs. Compress returns the input unchanged when output would not shrink.

## Fill with XFDF

Fill AcroForm fields from XFDF data. Text fields write /V, buttons and radios write /AS plus /V, appearances regenerate through /NeedAppearances.

```go
filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
```

```bash
curl -X POST localhost:8080/api/v1/fill -F pdf=@form.pdf -F xfdf=@data.xfdf -o filled.pdf
```

One caveat from pkg/gopdflib/fill.go: when the AcroForm dictionary sits inside a compressed object stream, the appearance flag patch cannot reach it. Values still update, only the flag is affected.

## Redact

Search text, draw boxes, apply. Two modes: visual_allowed overlays black boxes and is the default, secure_required scrubs content streams first and reports failure when nothing could be removed. The apply endpoint returns the PDF plus an X-Redaction-Report header.

```bash
curl -X POST localhost:8080/api/v1/redact/search -F pdf=@doc.pdf -F texts='["secret"]'
curl -X POST localhost:8080/api/v1/redact/apply -F pdf=@doc.pdf -F mode=secure_required \
  -F textSearch='[{"text":"secret"}]' -o redacted.pdf -D -
```

```python
from pypdfsuit import redact as R
rects = R.find_text_occurrences(pdf_data, "secret")
out = R.apply_redactions_advanced(pdf_data, {"mode": "secure_required", "textSearch": [{"text": "secret"}]})
```

OCR is a server only extension hook through tesseract. WASM rejects it fast. Capabilities endpoint labels each page text, mixed, image_only, or unknown.

## HTML to PDF and image

Pure Go through gowkhtmltopdf, no browser and no Chrome. POST html or a URL. Image output is png or jpg, svg is rejected with 400.

```bash
curl -X POST localhost:8080/api/v1/htmltopdf \
  -H 'Content-Type: application/json' \
  -d '{"html":"<h1>Hello</h1>","page_size":"A4","orientation":"Portrait"}' -o out.pdf
```

```python
from pypdfsuit import convert_html_to_pdf, HtmlToPDFRequest
pdf = convert_html_to_pdf(HtmlToPDFRequest(html="<h1>Hi</h1>", page_size="A4"))
```

Scripts do not run, so JS heavy pages render static only. Complex flex and grid may reflow, background CSS is ignored, WOFF2 fonts are skipped. DPI and zoom knobs are accepted but ignored. Details in documentation/HTML_PUREGO_SWAP.md and documentation/HTML_CONVERSION.md.

## Signatures, encryption, compliance

Sign with RSA or ECDSA P-256. The engine emits PKCS#7 detached under Adobe.PPKLite with a ByteRange placeholder, then fills the signature. Keys parse from PKCS#8, PKCS#1, or SEC1 PEM.

```json
"signature": {"enabled": true, "visible": true, "name": "Chinmay Sawant",
  "reason": "Approved", "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n...",
  "certificatePem": "-----BEGIN CERTIFICATE-----\n..."}
```

Encryption uses AES-128 with V=4 R=4, owner password required, per permission flags for printing, copying, modifying, annotations, form filling, assembly.

```json
"security": {"enabled": true, "ownerPassword": "secret", "userPassword": "view",
  "allowPrinting": true, "allowCopying": false}
```

Set pdfaCompliant true for PDF/A-4 output. The engine registers 12 Liberation fonts, embeds subsets, writes XMP with pdfaid and pdfuaid parts, and turns on tagging. TaggedPDF true alone gives the structure tree without the archival constraints. Validate with make test-verify-pdfs for veraPDF plus structure tree checks. Full story in documentation/DIGITAL_SIGNATURE_RSA_ECDSA.md and documentation/COMPLIANCE_PIPELINE_TODAY.md.

## Fonts and templates

GET /api/v1/fonts lists available faces. POST /api/v1/fonts uploads .ttf or .otf for server side use. GET /api/v1/template-data?file=name.json serves bundled sample templates. The frontend also keeps 12 Liberation TTFs and 3 sample templates offline in the browser after first load, and registers user fonts locally before uploading.

```bash
curl -X POST localhost:8080/api/v1/fonts -F "font=@MyFont.ttf"
curl "localhost:8080/api/v1/template-data?file=financial_report.json"
```

## Python bindings

Wheel pypdfsuit, version 7.0.1. Pure Python types plus ctypes into libgopdfsuit.so. Import surface: generate_pdf, generate_pdf_from_dict, get_available_fonts, merge_pdfs, split_pdf, parse_page_spec, fill_pdf_with_xfdf, compress_pdf, convert_html_to_pdf, convert_html_to_image, redact ops, fluent Font and Text plus cell helpers plus TemplateBuilder, typed errors. CGO exports live in bindings/python/cgo/exports.go.

## Math, SVG, navigation, decor

Typst math: set mathEnabled true on a cell and wrap the text in dollars. Example: {"props": "Helvetica:12:000:left:0:0:0:0", "text": "$ A = pi r^2 $", "mathEnabled": true}. Fractions, roots, matrices, and stretchy brackets render to content streams.

SVG: images containing an svg tag convert to PDF Form XObjects automatically, capped at 4 MiB.

Bookmarks build the outline tree with nested children, page numbers, named dests, and open flags. Cells link out with link for URLs or #dest for internal jumps, and anchor with dest. Config watermark draws diagonal text on every page. Footer auto adds Page X of Y at bottom right. pageBorder sets per side widths as left:right:top:bottom.

## REST API

All routes live under /api/v1 and share one error envelope {"code", "message", "error"}. Generate, fill, merge, split, compress, htmltopdf, htmltoimage, fonts, template-data, redact page-info plus text-positions plus capabilities plus apply plus search. Google auth enforces when REQUIRE_AUTH=1 or on Cloud Run, else open locally. Caps: template JSON 8 MiB, HTML JSON 2 MiB, PDF uploads 32 MiB, fonts 10 MiB, XFDF 8 MiB. Notes in documentation/AUTH_REQUEST_LIMITS_TODAY.md and documentation/HANDLERS_FASTPATH_ENVELOPE.md.

## Frontend app

React plus Vite in frontend/, served as static docs/. Pages: Home landing, Editor visual template builder with canvas plus Go and Python snippet copy, Viewer generate and preview, Merge with reorder, Split by pages, Compress with tier buttons, Filler for XFDF, Redaction with box drawing, HtmlToPdf and HtmlToImage converters, Documentation viewer, Comparison, Screenshots.

## In browser WASM

Two modules: compress.wasm for the Compress page, gopdfsuit.wasm for everything else. Default path runs local first and uploads to the server only after explicit consent in ConsentBanner. Offline templates, fonts, and cached modules keep working after first load. OCR, URL fetching, and Chrome paths stay server only. Details in documentation/WASM_VIEWER_EDITOR.md and documentation/FRONTEND_WASM_SPLIT.md.
