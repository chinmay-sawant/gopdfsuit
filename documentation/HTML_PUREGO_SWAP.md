# HTML pure Go swap

Date: 2026-09-04. Plan: plans/wasm/02-gowkhtmltopdf-replace.md. Commit family around 091d9ee.

## Old vs new

Old: chinmay-sawant/gochromedp fork with headless shell through CHROME_PATH, about 300M browser runtime.

New: github.com/chinmay-sawant/gowkhtmltopdf v0.2.5 at go.mod:8. Pure Go, no CGO, no Qt, no WebKit, no browser. Same HTTP plus handler plus service seam, terminal is internal/pdf/pdf.go ConvertHTMLToPDF at 51 and ConvertHTMLToImage at 82, shared core in internal/pdf/html_convert.go buildPDFDocument at 41 and buildImageDocument at 91.

Other deltas:

- PDF version claim: default unclaimed PDF 1.4. PDFVersion and PDFProfile stay unset pending veraPDF re-baseline at pdf.go:46-50.
- Docker: debian bookworm slim with dumb-init and ca-certificates. No CHROME_PATH or headless shell. Grep shows zero gochromedp hits in Go code.
- WASM split: pdf.go builds for server with URL fetch allowed through htmlSourceContent at 115. pdf_js.go builds for js with inline HTML only, URL returns ErrUpstream at 54.
- Python: chrome gates removed in bindings/python/pypdfsuit/html.py. No browser needed.
- Frontend: Chromium badge removed. dpi, low_quality, zoom, crop knobs hidden as no ops. svg option removed.

## Field mapping

Single table at html_convert.go:22-30:

- PDF mapped: PageSize, Orientation, margins, Grayscale.
- Image mapped: Format, Width, Height, Quality, Zoom, Crop fields. Note: HTML_CONVERSION.md says zoom and crop are no ops hidden in UI, while models.go:470-471 says they map onto ImageDocument. UI hides them either way.
- Source mapped: HTML or URL into Content per build policy.
- Ignored: DPI, LowQuality with no v0.2.5 field, Options free form map accepted with warnUnmappedHTMLOptions log.

## Code

Server convert at pdf.go:51-65:

```go
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	content, err := htmlSourceContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}
	warnUnmappedHTMLOptions("ConvertHTMLToPDF", req.Options)
	doc := buildPDFDocument(req, content)
	pdfData, err := runPDFDocument(doc)
	...
}
```

Shared builder at html_convert.go:41-56:

```go
func buildPDFDocument(req models.HTMLToPDFRequest, content gowkhtmltopdf.Content) *gowkhtmltopdf.Document {
	policy := gowkhtmltopdf.RestrictedNetworkPolicy()
	return &gowkhtmltopdf.Document{
		Pages: []gowkhtmltopdf.Page{{Source: content}},
		PageSize: req.PageSize, Orientation: req.Orientation,
		Margin: gowkhtmltopdf.Margin{Top: parseMarginMM(req.MarginTop), Right: parseMarginMM(req.MarginRight), Bottom: parseMarginMM(req.MarginBottom), Left: parseMarginMM(req.MarginLeft)},
		Grayscale: req.Grayscale, Network: &policy,
	}
}
```

Public guard at pkg/gopdflib/html.go:62-69:

```go
func ConvertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	if req.Format == "svg" {
		return nil, invalidInputError(op, "format svg is not supported: use png or jpg")
	}
```

Python at bindings/python/pypdfsuit/html.py:32-49:

```python
request = HtmlToPDFRequest(
    html="<html><body><h1>Hello World</h1></body></html>",
    page_size="A4",
    orientation="Portrait",
)
pdf_bytes = convert_html_to_pdf(request)
```

API with 2 MiB JSON cap and SSRF guard:

```bash
curl -X POST localhost:8080/api/v1/htmltopdf \
  -H 'Content-Type: application/json' \
  -d '{"html":"<h1>Hello</h1>","page_size":"A4","orientation":"Portrait","margin_top":"10mm","grayscale":false}' \
  --output converted.pdf
curl -X POST localhost:8080/api/v1/htmltoimage \
  -H 'Content-Type: application/json' \
  -d '{"html":"<h1>Hello</h1>","format":"png","width":800,"height":600,"quality":94}' \
  --output converted.png
```

## Limits

From documentation/HTML_CONVERSION.md plus code:

- Scripts stripped. JS heavy SPAs render static HTML only. Fixture spa-after-purego.pdf misses script built rows.
- Partial flex and grid. Complex layouts may reflow or stack.
- Background CSS does not paint.
- Fonts: WOFF2 and data URI font face skipped. Prefer TTF or WOFF or system fonts.
- svg output unsupported. htmltoimage accepts png and jpg only, svg returns 400 at handlers/html.go:82-85.
- No op but accepted: htmltopdf dpi and low_quality, htmltoimage zoom and crop fields, free form options map with warning log.
- Exactly one of html or url. Empty fails fast. WASM URL returns ErrUpstream, server only.
- Network: handler validateFetchURL SSRF guard plus RestrictedNetworkPolicy in engine. Private and link local and cross host redirects blocked.
- Output is plain PDF 1.4, not PDF/A-4 or PDF/UA-2.

## Entry map

- handlers.go:130-131 routes for htmltopdf and htmltoimage
- handlers/html.go:32 handleHTMLToPDF, 62 handleHTMLToImage, 15 rejectHTMLSource
- request_html.go:15 newHTMLToPDFRequest with A4 Portrait 10mm DPI 300 defaults, 40 newHTMLToImageRequest with png quality 94 zoom 1.0
- request.go:28 maxHTMLBodyBytes 2 MiB, 129 validateFetchURL
- models.go:451 HTMLToPDFRequest, 475 HTMLToImageRequest
- gopdflib/html.go:32 ConvertHTMLToPDF, 62 ConvertHTMLToImage
- Fixtures: sampledata/htmltopdf/purego_invoice.pdf, sampledata/htmltoimg/purego_invoice.png
