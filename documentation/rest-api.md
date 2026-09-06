# GoPDFSuit REST API

Base URL (local and Docker): `http://localhost:8080`

Run the pinned image:

```bash
docker run --rm -p 8080:8080 chinmaysawant/gopdfsuit:7.0.0
```

The server listens on `:8080`. The React SPA is served from `/gopdfsuit` (Vite build in `docs/`); `GET /` issues a `301` to `/gopdfsuit`.

All routes below are registered in `internal/handlers/handlers.go` inside `RegisterRoutesWithPolicy` on the `/api/v1` group.

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/generate/template-pdf` | `handleGenerateTemplatePDF` |
| POST | `/api/v1/fill` | `handleFillPDF` |
| POST | `/api/v1/merge` | `handleMergePDFs` |
| POST | `/api/v1/split` | `handlerSplitPDF` |
| POST | `/api/v1/compress` | `handleCompressPDF` |
| GET | `/api/v1/template-data` | `handleGetTemplateData` |
| GET | `/api/v1/fonts` | `handleGetFonts` |
| POST | `/api/v1/fonts` | `handleUploadFont` |
| POST | `/api/v1/htmltopdf` | `handleHTMLToPDF` |
| POST | `/api/v1/htmltoimage` | `handleHTMLToImage` |
| POST | `/api/v1/redact/page-info` | `handleRedactPageInfo` |
| POST | `/api/v1/redact/text-positions` | `handleRedactTextPositions` |
| POST | `/api/v1/redact/capabilities` | `handleRedactCapabilities` |
| POST | `/api/v1/redact/apply` | `handleRedactApply` |
| POST | `/api/v1/redact/search` | `handleRedactSearch` |
| OPTIONS | `/api/v1/*path` | CORS preflight |

## Authentication

Open by default locally. Enforced when `REQUIRE_AUTH=1`, or on Cloud Run when `K_SERVICE` or `K_REVISION` is set. When enforced, every `/api/v1/*` route requires `Authorization: Bearer <Google ID token>`. `OPTIONS` preflight bypasses auth.

```bash
curl -H "Authorization: Bearer $ID_TOKEN" http://localhost:8080/api/v1/fonts
```

Missing header returns `401 {"error": "Authorization header required"}`. Bad shape returns `401 {"error": "Invalid authorization header format. Expected: Bearer <token>"}`. Failed validation returns `401 {"error": "authentication failed"}`. These three auth rejections use the legacy `{"error"}` shape with no `code`/`message` fields.

## Error envelope

All handler rejections except the three auth cases above use one shape:

```json
{ "code": "invalid_input", "message": "human readable text", "error": "human readable text" }
```

`error` is a legacy alias always equal to `message`. `code` is one of `invalid_input`, `limit_exceeded`, `upstream`, `internal`.

| HTTP | `code` | Meaning |
|------|--------|---------|
| 400 | `invalid_input` | Bad JSON, bad multipart field, bad query, bad font extension, missing file field |
| 401 | (no code) | Auth only, legacy `{"error"}` shape |
| 422 | `invalid_input` | Backend classified input as invalid (corrupt PDF, bad page spec, bad XFDF) |
| 413 | `limit_exceeded` | Any body or upload cap hit |
| 429 | `limit_exceeded` | Concurrency limiter shed load (`Retry-After: 1`, message `"server busy"`) |
| 502 | `upstream` | Font-asset fetch or conversion dependency failed |
| 500 | `internal` | Unclassified engine failure |
| 403 | `invalid_input` | HTML `url` blocked by SSRF guard |

## Body-size caps

| Scope | Cap |
|-------|-----|
| `POST /api/v1/generate/template-pdf` JSON body | 8 MiB |
| `POST /api/v1/htmltopdf`, `/api/v1/htmltoimage` JSON body | 2 MiB |
| `POST /api/v1/merge` whole body and sum of PDF bytes | 128 MiB |
| `POST /api/v1/merge` file count (`pdf` fields) | 32 files |
| All other multipart bodies pre-parse | 40 MiB |
| Single PDF upload (`pdf` field) | 32 MiB |
| XFDF upload (`xfdf` field) | 8 MiB |
| Font upload (`font` field) | 10 MiB |

## POST /api/v1/generate/template-pdf

Renders a JSON template to PDF. Success: `200 application/pdf`, `Content-Disposition: attachment; filename=generated.pdf`. Cap: 8 MiB.

```bash
curl -X POST http://localhost:8080/api/v1/generate/template-pdf \
  -H "Content-Type: application/json" \
  --data-binary @template.json \
  --output generated.pdf
```

## POST /api/v1/fill

Fills AcroForm fields from XFDF. Success: `200 application/pdf`, `filename=filled.pdf`. Fields: `pdf` (file, required), `xfdf` (file, required).

```bash
curl -X POST http://localhost:8080/api/v1/fill \
  -F "pdf=@form.pdf;type=application/pdf" \
  -F "xfdf=@data.xfdf" \
  --output filled.pdf
```

## POST /api/v1/merge

Merges 2+ PDFs in form order. Success: `200 application/pdf`, `filename=merged.pdf`. Repeat the `pdf` field once per input, in merge order.

```bash
curl -X POST http://localhost:8080/api/v1/merge \
  -F "pdf=@a.pdf;type=application/pdf" \
  -F "pdf=@b.pdf;type=application/pdf" \
  --output merged.pdf
```

## POST /api/v1/split

Splits one PDF. Fields: `pdf` (file, required), `pages` (optional spec like `"1-3,5"`), `max_per_file` (optional positive int). One output part returns `application/pdf`; multiple parts return `application/zip`.

```bash
curl -X POST http://localhost:8080/api/v1/split \
  -F "pdf=@doc.pdf;type=application/pdf" \
  -F "pages=1-3,5" \
  -F "max_per_file=2" \
  --output splits.zip
```

## POST /api/v1/compress

Rewrites streams and images at a tier preset. Success: `200 application/pdf`, `filename=compressed.pdf`. Fields: `pdf` (file, required), `level` (optional `light|medium|heavy`, empty or unknown selects `medium`), `quality` (optional JPEG override), `max_image_dim` (optional max edge override).

```bash
curl -X POST http://localhost:8080/api/v1/compress \
  -F "pdf=@big.pdf;type=application/pdf" \
  -F "level=medium" \
  --output compressed.pdf
```

## GET /api/v1/template-data

Serves a repo-root `*.json` template file as JSON (for the editor). Query: `?file=<basename>.json` (only `.json`, no path traversal).

```bash
curl "http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json" --output template.json
```

## GET /api/v1/fonts

Lists fonts available to the generator. Success: `200 application/json {"fonts": [...]}`.

```bash
curl http://localhost:8080/api/v1/fonts
```

## POST /api/v1/fonts

Registers a custom font. Field: `font` (file, required, `.ttf` or `.otf` only, max 10 MiB).

```bash
curl -X POST http://localhost:8080/api/v1/fonts \
  -F "font=@MyFont.ttf"
```

## POST /api/v1/htmltopdf

Converts HTML (or a public URL) to PDF via the pure-Go pipeline. Success: `200 application/pdf`, `filename=converted.pdf`. Body: `html`, `url` (one required), `page_size` (default `A4`), `orientation` (default `Portrait`), `margin_top|margin_right|margin_bottom|margin_left` (default `10mm`), `grayscale`. Cap: 2 MiB.

```bash
curl -X POST http://localhost:8080/api/v1/htmltopdf \
  -H "Content-Type: application/json" \
  -d '{"html":"<h1>Hello</h1>","page_size":"A4","orientation":"Portrait"}' \
  --output converted.pdf
```

## POST /api/v1/htmltoimage

Converts HTML (or a public URL) to an image. Success: `200 image/png` (default) or `200 image/jpeg`. Fields: `format` (`png` default, `jpg`/`jpeg` alias; `svg` is rejected with `400`), `width`, `height`, `quality` (default `94`).

```bash
curl -X POST http://localhost:8080/api/v1/htmltoimage \
  -H "Content-Type: application/json" \
  -d '{"html":"<h1>Hello</h1>","format":"png","width":800,"quality":94}' \
  --output converted.png
```

## POST /api/v1/redact/page-info

Returns page dimensions. Fields: `pdf` (file, required). Success: `200 application/json {"totalPages": N, "pages": [...]}`.

```bash
curl -X POST http://localhost:8080/api/v1/redact/page-info \
  -F "pdf=@doc.pdf;type=application/pdf"
```

## POST /api/v1/redact/text-positions

Extracts positioned text for one page. Fields: `pdf` (file, required), `page` (required, 1-based int).

```bash
curl -X POST http://localhost:8080/api/v1/redact/text-positions \
  -F "pdf=@doc.pdf;type=application/pdf" \
  -F "page=1"
```

## POST /api/v1/redact/capabilities

Returns per-page capability info (`text | image_only | mixed | unknown`). Fields: `pdf` (file, required).

```bash
curl -X POST http://localhost:8080/api/v1/redact/capabilities \
  -F "pdf=@doc.pdf;type=application/pdf"
```

## POST /api/v1/redact/apply

Applies redactions. Success: `200 application/pdf` (`filename=redacted.pdf`) plus an `X-Redaction-Report` response header. Fields: `pdf` (file, required), `mode` (optional), `password` (optional), `blocks` (optional JSON array of `{pageNum, x, y, width, height}`), `textSearch` (optional JSON array), `text` (legacy comma-separated), `ocr` (optional JSON).

```bash
curl -X POST http://localhost:8080/api/v1/redact/apply \
  -F "pdf=@doc.pdf;type=application/pdf" \
  -F 'blocks=[{"pageNum":1,"x":10,"y":10,"width":100,"height":20}]' \
  -F 'textSearch=[{"text":"confidential"}]' \
  -F "mode=secure_required" \
  --output redacted.pdf -D headers.txt
```

## POST /api/v1/redact/search

Finds candidate redaction rectangles. Fields: `pdf` (file, required), `texts` (optional JSON string array) or `text` (comma-separated string).

```bash
curl -X POST http://localhost:8080/api/v1/redact/search \
  -F "pdf=@doc.pdf;type=application/pdf" \
  -F 'texts=["confidential","secret"]'
```
