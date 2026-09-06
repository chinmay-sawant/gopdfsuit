# HTML Conversion Guide

Convert HTML markup or a web page URL to PDF or an image (PNG/JPG).
The engine is pure-Go via `github.com/chinmay-sawant/gowkhtmltopdf v0.2.5`.
There is no browser, no Chromium, no headless shell, no CGO, no Qt, no WebKit. The same request shapes work
over HTTP (`POST /api/v1/htmltopdf`, `POST /api/v1/htmltoimage`), the Go
library (`pkg/gopdflib/html.go`), and the Python bindings
(`bindings/python/pypdfsuit/html.py`).

JSON keys are snake_case (`page_size`, `margin_top`, `low_quality`, `crop_width`). Go struct
fields are CamelCase (`PageSize`, `MarginTop`, `LowQuality`, `CropWidth`). Python dataclass attributes match the JSON keys.

## HTML to PDF

### Go

```go
package main

import (
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    // Inline HTML.
    pdfBytes, err := gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{
        HTML:        "<html><body><h1>Hello World</h1></body></html>",
        PageSize:    "A4",
        Orientation: "Portrait",
    })
    if err != nil {
        panic(err)
    }
    if err := os.WriteFile("out.pdf", pdfBytes, 0o644); err != nil {
        panic(err)
    }

    // URL (server builds only; see Limits below).
    pdfBytes, err = gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{
        URL:      "https://example.com",
        PageSize: "Letter",
    })
    _ = pdfBytes
    _ = err
}
```

Either `HTML` or `URL` is required. Both empty returns invalid input.

### Python

```python
from pypdfsuit import HtmlToPDFRequest, convert_html_to_pdf

# Inline HTML.
req = HtmlToPDFRequest(
    html="<html><body><h1>Hello World</h1></body></html>",
    page_size="A4",
    orientation="Portrait",
)
pdf_bytes = convert_html_to_pdf(req)
with open("out.pdf", "wb") as f:
    f.write(pdf_bytes)

# URL (server builds only; see Limits below).
req = HtmlToPDFRequest(url="https://example.com", page_size="Letter")
pdf_bytes = convert_html_to_pdf(req)
```

Raises `ValueError` when neither `html` nor `url` is set.

### curl

```bash
curl -s -X POST http://localhost:8080/api/v1/htmltopdf \
  -H 'Content-Type: application/json' \
  -d '{
    "html": "<html><body><h1>Hello World</h1></body></html>",
    "page_size": "A4",
    "orientation": "Portrait",
    "margin_top": "10mm",
    "margin_right": "10mm",
    "margin_bottom": "10mm",
    "margin_left": "10mm",
    "grayscale": false
  }' --output out.pdf

# URL form (server only):
curl -s -X POST http://localhost:8080/api/v1/htmltopdf \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://example.com", "page_size": "Letter"}' \
  --output url.pdf
```

Server defaults: `page_size` A4, `orientation` Portrait, each `margin_*` `10mm`, `dpi` 300 (accepted, ignored; see Limits).

## HTML to image

### Go

```go
imgBytes, err := gopdflib.ConvertHTMLToImage(gopdflib.HTMLToImageRequest{
    HTML:   "<html><body><h1>Hello World</h1></body></html>",
    Format: "png",
    Width:  800,
    Height: 600,
})
if err != nil {
    panic(err)
}
_ = os.WriteFile("out.png", imgBytes, 0o644)
```

### Python

```python
from pypdfsuit import HtmlToImageRequest, convert_html_to_image

req = HtmlToImageRequest(
    html="<html><body><h1>Hello</h1></body></html>",
    format="png",
    width=800,
    height=600,
)
img_bytes = convert_html_to_image(req)
with open("output.png", "wb") as f:
    f.write(img_bytes)
```

### curl

```bash
curl -s -X POST http://localhost:8080/api/v1/htmltoimage \
  -H 'Content-Type: application/json' \
  -d '{
    "html": "<html><body><h1>Hello</h1></body></html>",
    "format": "png",
    "width": 800,
    "height": 600,
    "quality": 94
  }' --output out.png
```

Server defaults: `format` `png` (empty defaults to `png`, `jpeg` canonicalizes to `jpg`), `quality` 94, `zoom` 1.0.

## Field reference

### `HTMLToPDFRequest`

| Field (JSON / Go / Python) | Type | Default | Effect |
| --- | --- | --- | --- |
| `html` / `HTML` / `html` | string | - | Inline markup source. One of `html` or `url` required. |
| `url` / `URL` / `url` | string | - | Page URL source. Server-only fetch; WASM takes it only as base URL alongside `html` (see Limits). |
| `output_path` / `OutputPath` / `output_path` | string | - | Optional output path. |
| `page_size` / `PageSize` / `page_size` | string | `A4` | Page size, e.g. `A4`, `Letter`. |
| `orientation` / `Orientation` / `orientation` | string | `Portrait` | `Portrait` or `Landscape`. |
| `margin_top` / `MarginTop` / `margin_top` | string | `10mm` | CSS-length string, parsed to mm float. |
| `margin_right` / `MarginRight` / `margin_right` | string | `10mm` | Same as above. |
| `margin_bottom` / `MarginBottom` / `margin_bottom` | string | `10mm` | Same as above. |
| `margin_left` / `MarginLeft` / `margin_left` | string | `10mm` | Same as above. |
| `grayscale` / `Grayscale` / `grayscale` | bool | `false` | Grayscale output. |
| `dpi` / `DPI` / `dpi` | int | `300` | Accepted but ignored (no engine knob). See Limits. |
| `low_quality` / `LowQuality` / `low_quality` | bool | `false` | Accepted but ignored (no engine knob). See Limits. |
| `options` / `Options` / `options` | map[string]string | - | Accepted but ignored; non-empty logs a warning. |

### `HTMLToImageRequest`

| Field (JSON / Go / Python) | Type | Default | Effect |
| --- | --- | --- | --- |
| `html` / `HTML` / `html` | string | - | Inline markup source. One of `html` or `url` required. |
| `url` / `URL` / `url` | string | - | Page URL source. Server-only fetch; WASM takes it only as base URL alongside `html`. |
| `output_path` / `OutputPath` / `output_path` | string | - | Optional output path. |
| `format` / `Format` / `format` | string | `png` | `png` or `jpg` (`jpeg` accepted, canonicalized to `jpg`). `svg` is rejected as invalid input. |
| `width` / `Width` / `width` | int | - | Image width in pixels. |
| `height` / `Height` / `height` | int | - | Image height in pixels. |
| `quality` / `Quality` / `quality` | int | `94` | Image quality 1-100. |
| `zoom` / `Zoom` / `zoom` | float | `1.0` | Accepted; no-op in current engine. |
| `crop_width` / `CropWidth` / `crop_width` | int | - | Accepted; no-op in current engine. |
| `crop_height` / `CropHeight` / `crop_height` | int | - | Accepted; no-op in current engine. |
| `crop_x` / `CropX` / `crop_x` | int | - | Accepted; no-op in current engine. |
| `crop_y` / `CropY` / `crop_y` | int | - | Accepted; no-op in current engine. |
| `options` / `Options` / `options` | map[string]string | - | Accepted but ignored with a warning. |

## Limits

- `dpi` and `low_quality` are ignored. Accepted for compatibility but the engine has no corresponding knob; hidden in the UI.
- `zoom` and `crop_*` are ignored. Accepted but currently no-ops, hidden in the UI.
- `options` is ignored. A non-empty map logs a server-side warning and is otherwise dropped.
- `format: svg` is rejected. `htmltoimage` supports `png`/`jpg` only.
- WASM is inline-HTML-only. A URL-only request fails fast with guidance; the supported shape is inline `HTML` with `URL` optionally set as the document base URL. Sites without CORS headers cannot be fetched cross-origin. URL fetching is server-only.
- Fidelity limits (static renderer, scripts stripped): JS-heavy SPAs render as static HTML only; partial flex/grid support; `background` CSS does not paint; WOFF2 and data-URI `@font-face` are skipped (prefer TTF/WOFF or system fonts).
- SSRF guard: server URL fetches allow only `http/https` with no loopback, private, link-local, multicast, or `localhost` targets.

## See also

- [HTML_CONVERSION.md](HTML_CONVERSION.md) - engine note plus supported-knob summary.
- [HTML_PUREGO_SWAP.md](HTML_PUREGO_SWAP.md) - gochromedp to gowkhtmltopdf swap record.
- [FEATURES.md](FEATURES.md) - one-line feature entry (png/jpg only, svg 400).
