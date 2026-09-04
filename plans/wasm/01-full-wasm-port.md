# Plans/WASM - Full gopdfsuit Port to Browser WASM

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; ref `cmd/wasmcompress/main.go`, `pkg/gopdflib/*`, `internal/pdf/*`
> **Status:** implemented - pending make gates (10 workers landed, PDF 1 regen ok 73000 bytes)
> **Estimated effort:** L - new `cmd/wasm` surface plus sonic/js build-tag work

---

## Overview

Port every pure-Go gopdfsuit op into `GOOS=js GOARCH=wasm` so the browser can run Generate, Merge, Split, Compress, Fill, and text-path Redact without `POST /api/v1/*`. Chrome-backed HTML ops stay server-side.

## Executive Summary

Today only `goCompressPDF(Uint8Array, levelString)` is exposed (`cmd/wasmcompress/main.go:18-70`, build `makefile:78-83`). `pkg/gopdflib` already owns validation with `internal/pdf/{merge,compress,redact,form,generator,font,pdfobj,svg}` pure-Go, while `internal/pdf/pdf.go` (gochromedp), `bindings/python/cgo` (cgo), `redact/ocr_adapter.go` (os/exec), and `font/provision.go` (net/http + fs) cannot enter the WASM closure. The plan builds seams first, then exposes one JS binding per op with envelope errors.

## Phase 1: WASM build seams and correctness

### 1.1 Isolate non-portable imports

- [x] `internal/pdf/pdf.go` - gate `chinmay-sawant/gochromedp` import behind `//go:build !js` with `js` stub returning `ErrUpstream` - proof: `GOOS=js GOARCH=wasm go list ./internal/pdf` ok, no gochromedp in `go list -deps ./cmd/wasmcompress`
- [x] `pkg/gopdflib/html.go` - same `js` stub for `ConvertHTMLToPDF/ConvertHTMLToImage` - proof: wasm build log has no `gochromedp`
- [x] `pkg/gopdflib/adapter.go` - audit `bytedance/sonic` under `js/wasm`; compat fallback covers js/wasm so no `encoding/json` swap needed - proof: `GOOS=js GOARCH=wasm go build ./cmd/wasmcompress` succeeds (6.2M)
- [x] `internal/pdf/font/provision.go` + `font/pdfa.go` - gate `net/http FetchToTemp` and Liberation download for `js`; document `RegisterFontFromData/RegisterFontFromBase64` as WASM path - proof: `!js` guard plus WASM reject with unsupported error
- [x] `internal/pdf/redact/ocr_adapter.go` + `ocr_exec_*.go` - `OCR.Enabled=true` returns `CodeInvalidInput unsupported in WASM` instead of `os/exec` - proof: `go test ./internal/pdf/redact -run 'TestOCR|TestOcr|TestRedact'` ok

### 1.2 Filesystem entry points

- [x] `internal/pdf/font/registry.go` - verify `RegisterFontFromData` and `RegisterFontFromBase64` need no `os.ReadFile/ReadDir` - proof: `go test ./internal/pdf/font` ok plus docs marked WASM-safe path
- [x] `internal/pdf/font/ttf.go` - forbid `LoadTTFFromFile/LoadFontsFromDirectory` in WASM docs - proof: godoc comment updated

## Phase 2: JS API surface

### 2.1 New WASM entrypoint

- [x] `cmd/wasm/main.go` (new, `cmd/wasmcompress` untouched) - expose `goGeneratePDF`, `goMergePDFs`, `goSplitPDF`, `goFillPDF`, `goRedact*` alongside `goCompressPDF` - proof: `GOOS=js GOARCH=wasm go build -o /tmp/opencode/gopdfsuit.wasm ./cmd/wasm` exit 0 (11M)
- [x] `cmd/wasm/main.go` - use `js.CopyBytesToGo` in and `js.CopyBytesToJS` out plus `recover()` to `{code,message}` via `gopdflib.EnvelopeOf` - proof: mirrors `cmd/wasmcompress/main.go:31-70` pattern, `go vet ./cmd/wasm` clean
- [x] `makefile:78-83` - add `wasm:` target building `frontend/public/gopdfsuit.wasm` plus `wasm_exec.js` copy to `frontend/public` and `sampledata/wasm-js/` - proof: target added plus `.PHONY` (build execution deferred, no make run per task constraints)

### 2.2 Per-op bindings

- [x] `pkg/gopdflib/generator.go` - `GeneratePDF/GeneratePDFBorrowed` callable from WASM (`DecodeTemplateJSON`, `GeneratePDFFromJSON`, `GeneratePDFBorrowedFromJSON`) - proof: `TestGeneratePDFFromJSONFinancialReport` 100170 bytes JSON to 73000 bytes PDF ok
- [x] `pkg/gopdflib/merge.go` - `MergePDFs([][]byte)` takes JS array of `Uint8Array` - proof: `go test ./internal/pdf/merge` ok
- [x] `pkg/gopdflib/split.go` - `SplitPDF` plus `ParsePageSpec` accept `{pages, maxPerFile}` object (`ParseSplitSpecJSON`, `SplitPDFWithSpecJSON`) - proof: `go test ./pkg/gopdflib -run 'TestMerge|TestSplit|TestGenerate'` 9 tests pass
- [x] `pkg/gopdflib/fill.go` - `FillPDFWithXFDF` takes two `Uint8Array`; note `/NeedAppearances` object-stream limit in plan comment - proof: `go test ./internal/pdf/form` ok
- [x] `pkg/gopdflib/redact.go` - expose `GetPageInfo/ExtractTextPositions/FindTextOccurrences/ApplyRedactions/ApplyRedactionsAdvancedWithReport` text path only - proof: `go test ./internal/pdf/redact` ok

## Phase 3: Data contracts and limits

- [x] WASM JS shim - enforce `MaxCompressInputBytes 32MiB` in JS before copy (`sampledata/wasm-js/gopdfsuit.js`) - proof: cap asserted in shim
- [x] WASM JS shim - fix numeric level bug: map `1|2|3` to `light|medium|heavy` (`sampledata/wasm-js/gopdfsuit.js`, `frontend/src/utils/compressPdf.js` now passes string via `toServerLevel`) - proof: code diff
- [x] Template JSON path - `DecodeTemplateJSON` reuses `models.PDFTemplate` validation and `PreallocForDecode`; base64 `imagedata/fontData` stays string - proof: PDF 1 regen 73000 bytes via `TestGeneratePDFFromJSONFinancialReport`; compliant browser variant proven in `plans/wasm/04-frontend-wasm-split-fonts-compliance.md` Phase 3
- [x] Split multi-file - return JS array of `Uint8Array`, never Go-side zip - proof: no `archive/zip` in WASM closure

## Phase 4: Validation and performance

- [x] `test/verify_pdfs.sh` - WASM-generated Generate/Merge/Split/Fill outputs pass plus `structure_tree_check.py` MCID check - proof: `node sampledata/wasm-js/run.mjs` 5/5 green (generate 4157B, merge 341381B, split 1 part, fill 82089B, redact search 2 hits plus 5175B), structure tree 5/5 clean (TD 0/0, TR 0), veraPDF PDF/A-4 FAIL as expected for unclaimed ad-hoc outputs (claimed `pdfaCompliant` pipeline stays 10/10 in `make test`)
- [x] Bundle size - record `ls -lh frontend/public/gopdfsuit.wasm` cold load plus `sampledata/benchmarks/` generate p50 before/after - proof: `gopdfsuit.wasm` 11M vs `compress.wasm` 3.6M; financial-report generate x21 dev-loop 73000B min 318us p50 446us max 138ms (first-run warmup)
- [x] `sampledata/wasm-js/` - add `index.html` plus `run.mjs` plus `gopdfsuit.js` plus `README.md` demo mirroring `sampledata/compress-js/` - proof: files present, `node --check` ok
- [x] `documentation/*` plus `frontend/src/components/documentation/content/*` - document which ops are browser-local vs server-only (HTML stays server) - proof: `GETTING_STARTED_GOPDFLIB.md` table plus getting-started and sample-data content updates

## Phase 5: Closure

- [x] `plans/wasm/03-wasm-everywhere-noauth-editor.md` - mark Generate/Merge/Split/Fill/Redact rows `[x]` with pointer back to this ledger once validated - proof: 03 Phase 3 rows now `[x]` pointing here; Node smoke 5/5 green
- [x] `documentation/*` plus `frontend/src/components/documentation/content/*` - document which ops are browser-local vs server-only (HTML stays server) - proof: `GETTING_STARTED_GOPDFLIB.md` WASM-vs-server table plus getting-started, sample-data, api-reference transport labels (shipped in `091d9ee`)

## Dependencies

- `cmd/wasmcompress/main.go:18-70` existing pattern; `makefile:70-83` build; `frontend/src/utils/compressPdf.js:1-181` loader to generalize
- `pkg/gopdflib/{generator,merge,split,compress,fill,redact,adapter,errors}.go`; `internal/pdf/{generator,merge,compress,redact,form,font,pdfobj,svg}`
- Blockers: `internal/pdf/pdf.go` gochromedp split, sonic `js/wasm` verdict, `gowkhtmltopdf` HTML decision in `plans/wasm/02-gowkhtmltopdf-replace.md`
- Explicit non-goals: `ConvertHTMLToPDF/ConvertHTMLToImage` in WASM, `bindings/python/cgo`, OCR subprocess path
