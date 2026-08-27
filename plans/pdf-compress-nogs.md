# gopdfsuit - PDF compress without Ghostscript

> **Parent:** conversation request — PDF-in / PDF-out compressor on the custom engine; gopdflib + HTTP API only (no frontend, no Ghostscript)
> **Status:** engine + gopdflib done; next ledger is Phase 7 — client-side JS (no CLI, no API for the product path)
> **Estimated effort:** one implementation pass

---

## Overview

Add a standalone `CompressPDF` that takes existing PDF bytes and returns compressed PDF bytes. Wire it through `pkg/gopdflib` and `POST /api/v1/compress`. Keep generation-path zlib (`BestSpeed`) unchanged.

## Executive Summary

Leftover `internal/pdf/compress` on master imported `gopdfsuit/v5` and was not exported. Retargeted to v6, exposed `gopdflib.CompressPDF`, added `POST /api/v1/compress` via `PDFService`. Proven with library + handler tests. `make lint` and `make test` passed.

## Phase 1: Engine (`internal/pdf/compress`)

### 1.1 Module path
- [x] Replace `gopdfsuit/v5` imports with `gopdfsuit/v6` in `internal/pdf/compress/*.go` so the package builds on this module. Proof: `go test ./internal/pdf/compress/` compiles (no test files, package builds).

### 1.2 Standalone function
- [x] Keep `compress.CompressPDF([]byte, Options) ([]byte, error)` as the only engine entry. Encrypted / non-PDF input returns an error. If the rewrite is not smaller, return the original bytes. Proof: `TestCompressPDF_NotPDF` and `TestCompressPDF_ImagePDF` in `pkg/gopdflib/compress_test.go`.

## Phase 2: Library contract (`pkg/gopdflib`)

### 2.1 Public wrapper
- [x] Add `pkg/gopdflib/compress.go`: alias `CompressOptions`, one-line `CompressPDF` wrapping `internal/pdf/compress`.

### 2.2 Docs
- [x] List `[CompressPDF]` in `pkg/gopdflib/doc.go` features.

## Phase 3: HTTP API

### 3.1 Service
- [x] Add `CompressPDF(data []byte, opts compress.Options) ([]byte, error)` to `PDFService` / `defaultPDFService` in `internal/handlers/services.go`.

### 3.2 Route
- [x] Register `POST /api/v1/compress` next to merge/split. Form field `pdf` required; optional `quality`, `max_image_dim`. Success: `application/pdf`, `filename=compressed.pdf`.

### 3.3 Mock
- [x] Update `internal/handlers/mocks/mock_services.go` so gomock tests compile.

## Phase 4: Tests

### 4.1 Library
- [x] `pkg/gopdflib/compress_test.go`: stdlib test — valid PDF in, `%PDF-` out, `len(out) <= len(in)`; non-PDF returns error. Proof: `go test ./pkg/gopdflib/` ok.

### 4.2 Handler
- [x] Gomock success + missing-file 400 for `/api/v1/compress` in `internal/handlers/handlers_gomock_test.go`. Proof: `go test ./internal/handlers/` ok.

## Phase 5: Closure

### 5.1 Lint
- [x] `make lint` — exit 0. `golangci-lint run -E revive,gocritic,gocyclo,goconst ./...` and `frontend npm run lint` both passed.

### 5.2 Test
- [x] `make test` — exit 0. `go test ./...` ok; python pytest 43 passed, 4 skipped; `test/verify_pdfs.sh` PASS.

### 5.3 Branch
- [x] Create `feature/compression-nogs`.

## Dependencies

- Phase 2 depends on 1.1 compiling.
- Phase 3 depends on Phase 1 (handler calls internal, not gopdflib).
- Phase 4 depends on Phases 2 and 3.
- Phase 5 depends on 4 passing locally.
- Phase 7 depends on Phase 6 (same `CompressPDF` + levels). WASM must not import the HTTP stack.

## Phase 6: Ghostscript-parity constraints

Business rules from the existing Ghostscript product. Text and vector operators stay intact; only images, fonts, metadata, and stream wrappers change.

### 6.1 JPEG tiers
- [x] Light = quality 92, Medium = 75, Heavy = 50. Default Medium. Optional `quality` / `max_image_dim` still override. API form field `level=light|medium|heavy`. Proof: `Options.withDefaults` in `internal/pdf/compress/compress.go`; handler `level` in `handleCompressPDF`.

### 6.2 Bicubic downsample
- [x] Replace nearest-neighbor with Keys bicubic (a = -0.5). Pixel cap: Light 1920, Medium 1275, Heavy 612. Proof: `downscaleBicubic` / `sampleBicubic` in `internal/pdf/compress/image.go`.

### 6.3 Stream compression
- [x] Recompress existing Flate at BestCompression; wrap uncompressed non-image streams in FlateDecode when smaller. Do not Flate JPEG (`DCTDecode`) payloads. Proof: `recompressFlate` + `applyFlate` in `stream.go`.

### 6.4 Metadata stripping
- [x] Drop trailer `/Info` and `/ID`. Remove catalog `/Metadata`, page `/Thumb`, `/PieceInfo`, and `/Type /Metadata` objects. Proof: `stripDocumentMetadata`; `TestCompressPDF_StripsMetadata`.

### 6.5 Font subsetting
- [x] For `FontFile2` TTFs, drop unused glyph outlines while keeping original GIDs so content streams stay valid. CFF/`FontFile3` left as Flate-only. Proof: `font.CompactUnusedGlyphs` + `compactFontStream`.

### 6.6 Tests + gates
- [x] Library tests for levels, metadata strip, and image shrink. Handler accepts `level`. Proof: `go test ./pkg/gopdflib/ ./internal/handlers/` ok.
- [x] `make lint` exit 0 after this phase. Targeted `go test` for compress/font/gopdflib/handlers ok.

## Phase 7: Client-side JavaScript (no CLI, no HTTP)

Friend’s product constraint: compression runs **only in the browser**. Do not add a CLI. Do not call `POST /api/v1/compress` from this UI. The Go HTTP handler may stay in the server for other gopdfsuit users; the compress product path is WASM + JS.

**How we code it:** compile `compress.CompressPDF` to WebAssembly (`GOOS=js GOARCH=wasm`). JS loads the wasm, passes a `Uint8Array`, gets a `Uint8Array` back. Same Light / Medium / Heavy engine as `sampledata/compress`. No Ghostscript, no server round-trip, MIT engine stays private.

Public JS shape:

```js
import { compressPDF } from './compress.js'

const out = await compressPDF(uint8, { level: 2 })
// level 1 = light (JPEG 92), 2 = medium (75), 3 = heavy (50)
```

### 7.1 WASM stub
- [x] Add `cmd/wasmcompress/main.go` (`//go:build js && wasm`) that registers `goCompressPDF(bytes, level)` via `syscall/js`. Import only `internal/pdf/compress` — not gin, not gochromedp.
- [x] Expected: `js.CopyBytesToGo` / `js.CopyBytesToJS`; errors returned as `{ error: string }`.

### 7.2 Build
- [x] Makefile target `wasm-compress`: `GOOS=js GOARCH=wasm go build -o frontend/public/compress.wasm ./cmd/wasmcompress` and copy `$(go env GOROOT)/lib/wasm/wasm_exec.js` (or `misc/wasm/wasm_exec.js`) next to it.
- [x] Proof: `file frontend/public/compress.wasm` reports `WebAssembly (wasm) binary module version 0x1 (MVP)`.

### 7.3 JS wrapper
- [x] Add `frontend/src/utils/compressPdf.js`: init wasm once, map `level` `1|2|3` or `light|medium|heavy` to the Go Level, return `Uint8Array`. Never `fetch('/api/v1/compress')`.
- [x] Proof: unit-level smoke in Node is optional; browser path is the gate. Node sample `node sampledata/compress/run.mjs` matches Go sizes.

### 7.4 UI
- [x] Add `/compress` page (Split.jsx pattern): local file picker, three level buttons, original vs compressed size, download. All work happens in the tab.
- [x] Wire `App.jsx` + `Navbar.jsx`. No auth required for this page (file never leaves the machine).

### 7.5 Sample
- [x] Keep `sampledata/compress/` as the Go library example. Add a short note in that folder’s `main.go` comment that the browser path is WASM, not this `go run`.
- [x] JS/WASM sample is `sampledata/compress-js/` (`compress.js`, `run.mjs`, `index.html`); writes `report_js_level_{1,2,3}.pdf`.

### 7.6 Docs
- [x] README features: PDF compress (library + **in-browser WASM**).
- [x] `guides/GETTING_STARTED_GOPDFLIB.md`: `CompressPDF` + level table.
- [x] Do **not** document a CLI.

### 7.7 Closure
- [x] `make lint` — exit 0 (`golangci-lint` + `frontend npm run lint`).
- [x] `make test` — exit 0 (`go test ./...`; python pytest 43 passed, 4 skipped; `test/verify_pdfs.sh` PASS).
- [x] Node WASM sample on `report.pdf` matches Go: Light 11.0%, Medium 76.4%, Heavy 87.4%. Vite serves `compress.wasm` as `application/wasm`. In-app click-through of `/compress` still needs a browser.

## Out of scope (deferred)

- [~] **CLI** — friend does not want CLI/API for this product. Not started. Revisit only if a separate ops tool is requested.
- [~] Frontend compress **via HTTP** — superseded by Phase 7 WASM. Do not POST to `/api/v1/compress` from the new page.
- [~] Python CGO export — same boundary as before.
- [~] Arbitrary third-party PDFs with encryption, JPEG2000, JBIG2, or filter arrays — rejected or left as-is.
- [~] True CID remapping / CFF subset — GID-preserving TTF hollow only.
- [~] Rewriting the compressor in pure JS (`pdf-lib` etc.) — would fork the engine and drop Ghostscript-parity. WASM reuses `internal/pdf/compress`.
