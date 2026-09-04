## Summary

Adds a Ghostscript-free PDF compressor on the existing engine: `gopdflib.CompressPDF`, `POST /api/v1/compress`, and an in-browser WASM path on `/compress`. Light / Medium / Heavy tiers downsample images, drop unused TTF glyphs, strip metadata, and Flate streams. Encrypted files are rejected. Input and inflate size are capped so a hostile PDF cannot grow without bound.

---

## Motivation / context

- Plans: `plans/pdf-compress-nogs.md`
- Issues: see **Related issues**
- Leftover `internal/pdf/compress` on master imported v5 and was not exported. This retargets it to v6, exposes the library + HTTP API, then compiles the same engine to WASM so the product UI never POSTs the file.
- No Ghostscript (AGPL). Engine is MIT; `wasm_exec.js` is Go Authors BSD-3-Clause. See `sampledata/compress-js/README.md`.

---

## Changes

### Engine (`internal/pdf/compress`)

- `CompressPDF([]byte, Options)` is the only engine entry.
- Tiers: Light (JPEG 92, max edge 1920), Medium (75 / 1275, default), Heavy (50 / 612). `JPEGQuality` and `MaxImageDim` override the preset when > 0, with `MaxImageDim` capped at 4096.
- Bicubic downsample (Keys, a = -0.5), Flate recompress at BestCompression, trailer `/Info` + `/ID` stripped, unused TTF glyph outlines dropped (GID-preserving).
- Non-PDF and encrypted input return an error.
- Shared limits in `limits.go` (library, HTTP, WASM): 32 MiB input, 48 MiB Flate inflate, 16 MP / 8192 image edge, 50_000 objects.

### Library (`pkg/gopdflib`)

- Public `CompressPDF` + `CompressOptions` alias and `CompressLight` / `CompressMedium` / `CompressHeavy`.
- Docs and example updated (`guides/GETTING_STARTED_GOPDFLIB.md`, package doc).

### HTTP API

- `PDFService.CompressPDF` + `POST /api/v1/compress`.
- Form field `pdf` required; optional `level`, `quality`, `max_image_dim`.
- Upload read through `LimitReader` (32 MiB). Success: `application/pdf`, `filename=compressed.pdf`.

### In-browser WASM (product path)

- `cmd/wasmcompress` (`//go:build js && wasm`) registers `goCompressPDF`. Uint8Array only; panics recovered.
- `make wasm-compress` writes `frontend/public/compress.wasm` + `wasm_exec.js`. CI frontend build runs that target.
- `/compress` page (Split-style): local file picker, three levels, size delta, download. No auth, no `fetch('/api/v1/compress')`.
- Wired in `App.jsx`, `Navbar`, home features.

### Samples

- `sampledata/compress` — Go library (`go run .`) → `report_level_{1,2,3}.pdf`.
- `sampledata/compress-js` — WASM (`node run.mjs` or `index.html`) → `report_js_level_{1,2,3}.pdf`. README lists handled vs not-handled cases and license.

### Tests

- Library: non-PDF error, image PDF shrinks, metadata strip, level override.
- Handler gomock: success with parsed options, missing-file 400.
- Engine: high object number rejected, inflate bomb capped, image pixel budget, `MaxImageDim` clamp.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Extra CPU on compress (bicubic + JPEG + Flate BestCompression). Generation-path zlib BestSpeed is unchanged. WASM compress is synchronous on the main thread. |
| **Memory** | Whole-PDF object map; decode buffers capped (32 MiB in, 48 MiB inflate, 16 MP images). |
| **Behavior / correctness** | Text and vector operators stay intact. Encrypted / JPEG2000 / JBIG2 / filter-array PDFs are rejected or left as-is. Hostile oversized PDFs error instead of allocating without bound. |
| **API / CLI** | New `gopdflib.CompressPDF`. New `POST /api/v1/compress`. New `/compress` WASM UI. No CLI. |
| **Dependencies** | None. `wasm_exec.js` is copied from the Go toolchain (BSD-3-Clause). |
| **Binary size / build time** | `compress.wasm` ~6 MiB in `frontend/public` and `sampledata/compress-js`. `make wasm-compress` before frontend build. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Additive API and a new page. Existing generate / merge / split paths unchanged. |

---

## Test plan

- [x] `make test`
- [x] `make lint` / `go vet`
- [x] `go test ./pkg/gopdflib/ ./internal/handlers/ ./internal/pdf/compress/`
- [x] `cd sampledata/compress && go run .`
- [x] `make wasm-compress && node sampledata/compress-js/run.mjs` (sizes match Go)

### Commands

```sh
make lint
make test
make wasm-compress
cd sampledata/compress && go run .
node sampledata/compress-js/run.mjs
```

---

## Screenshots / sample output

`sampledata/compress` and `sampledata/compress-js` on `report.pdf` (596341 bytes) produce the same sizes:

```
source  report.pdf  596341 bytes
level 1 Light   530762 bytes  (11.0% smaller)
level 2 Medium  140897 bytes  (76.4% smaller)
level 3 Heavy    74890 bytes  (87.4% smaller)
```

Handled vs not-handled (WASM): `sampledata/compress-js/README.md`.

---

## Related issues

- Closes #77

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-pdf-compress-nogs.md`

---

## Follow-ups (out of scope)

- Python CGO export.
- CFF / CID remapping; JPEG2000 / JBIG2 / encrypted files.
- Moving WASM compress off the main thread (Worker / Promise).
- Hard timeout / isolated process around `CompressPDF`.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in the diff
- [ ] Public API / CLI changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] `wasm_exec.js` keeps the Go Authors copyright header
