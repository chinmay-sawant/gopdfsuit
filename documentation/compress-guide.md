# Compress Guide

Compress an existing PDF: image downsample plus JPEG recompression at the chosen tier, unused TTF glyph outlines dropped, document metadata stripped, streams Flate-compressed. No Ghostscript, no browser.

## Tiers

Empty or unknown level selects Medium. Level names are lowercase strings: `light`, `medium`, `heavy`.

| Tier | Value | JPEG quality | Max image edge (px) | Use when |
|------|-------|--------------|---------------------|----------|
| Light | `light` | 92 | 1920 | Print-quality output, mild shrink |
| Medium | `medium` | 75 | 1275 | Default, balanced size and quality |
| Heavy | `heavy` | 50 | 612 | Smallest output, screen-only reading |

`CompressOptions` also accepts overrides: `JPEGQuality` and `MaxImageDim` apply when greater than zero. Quality above 100 is clamped to 100. `MaxImageDim` is capped at 4096.

## Go

```go
package main

import (
    "log"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    pdfBytes, err := os.ReadFile("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    out, err := gopdflib.CompressPDF(pdfBytes, gopdflib.CompressOptions{
        Level: gopdflib.CompressMedium,
    })
    if err != nil {
        log.Fatal(err)
    }
    if err := os.WriteFile("document-compressed.pdf", out, 0644); err != nil {
        log.Fatal(err)
    }
}
```

## Python

Valid levels: `"light"`, `"medium"`, `"heavy"`. Empty string selects the engine default (Medium). Anything else raises `ValueError`.

```python
from pypdfsuit import compress_pdf

with open("document.pdf", "rb") as f:
    src = f.read()

out = compress_pdf(src, level="medium")

with open("document-compressed.pdf", "wb") as f:
    f.write(out)
```

## HTTP

Endpoint: `POST /api/v1/compress`. Fields: `pdf` (required file), `level` (optional, empty or unknown selects Medium), `quality` (optional JPEG override), `max_image_dim` (optional max edge override). Returns `application/pdf`.

```bash
curl -X POST http://localhost:8080/api/v1/compress \
  -F "pdf=@document.pdf;type=application/pdf" \
  -F "level=medium" \
  --output document-compressed.pdf
```

With overrides:

```bash
curl -X POST http://localhost:8080/api/v1/compress \
  -F "pdf=@document.pdf;type=application/pdf" \
  -F "level=heavy" \
  -F "quality=60" \
  -F "max_image_dim=800" \
  --output document-compressed.pdf
```

## WASM (in-browser)

- Artifact: `frontend/public/compress.wasm` (built with `GOOS=js GOARCH=wasm`, see `make wasm-compress`).
- Exposes global `goCompressPDF(bytes, level)`: `Uint8Array` of PDF bytes (required, non-empty), optional level string. Missing level defaults to Medium.
- Returns `Uint8Array` on success, or `{code, message, error}` envelope on failure.
- Size cap and level parsing are shared with Go, HTTP, and CGO through `pkg/gopdflib`, so browser results match server results for the same tier.

## Limits

- Max input: 32 MiB (`MaxCompressInputBytes`). Applies to library, HTTP, and WASM entry points. Larger input is rejected.
- Max objects: 50,000 object count and object number.
- Image guards: 16,000,000 pixels max, 8192 px max edge, 48 MiB single-stream Flate cap.
- Only 8-bit images and DCT / Flate / raw pixel streams are recompressed; SMask images, non-8-bit images, DecodeParms predictor streams, and unsupported color spaces are left untouched.
- Files with more than 2 object streams are returned unchanged (linearized and object-packed files stay opaque).

## No-shrink passthrough

Compression never grows your file. If the rewrite is not smaller than the input, the original bytes are returned unchanged. The same passthrough applies when the catalog object is missing after rewrite, the output drops all streams due to a parser miss, or the write fails. Callers can always write the result directly; identical bytes mean "already optimal".

## Encrypted PDFs

Encrypted files are rejected before any rewrite. Expect an error containing `cannot compress encrypted PDF` from Go, or the equivalent HTTP 4xx / Python exception / WASM `{code, message}` envelope. Decrypt first, then compress.
