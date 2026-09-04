# Plans/WASM - Full gopdfsuit Port to Browser WASM

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; ref `cmd/wasmcompress/main.go`, `pkg/gopdflib/*`, `internal/pdf/*`
> **Status:** planning - only compress ships in WASM today
> **Estimated effort:** L - new `cmd/wasm` surface plus sonic/js build-tag work

---

## Overview

Port every pure-Go gopdfsuit op into `GOOS=js GOARCH=wasm` so the browser can run Generate, Merge, Split, Compress, Fill, and text-path Redact without `POST /api/v1/*`. Chrome-backed HTML ops stay server-side.

## Executive Summary

Today only `goCompressPDF(Uint8Array, levelString)` is exposed (`cmd/wasmcompress/main.go:18-70`, build `makefile:78-83`). `pkg/gopdflib` already owns validation with `internal/pdf/{merge,compress,redact,form,generator,font,pdfobj,svg}` pure-Go, while `internal/pdf/pdf.go` (gochromedp), `bindings/python/cgo` (cgo), `redact/ocr_adapter.go` (os/exec), and `font/provision.go` (net/http + fs) cannot enter the WASM closure. The plan builds seams first, then exposes one JS binding per op with envelope errors.

## Phase 1: WASM build seams and correctness

### 1.1 Isolate non-portable imports

- [ ] `internal/pdf/pdf.go` - gate `chinmay-sawant/gochromedp` import behind `//go:build !js` with `js` stub returning `ErrUpstream` - proof: `GOOS=js GOARCH=wasm go list ./internal/pdf`
- [ ] `pkg/gopdflib/html.go` - same `js` stub for `ConvertHTMLToPDF/ConvertHTMLToImage` - proof: wasm build log has no `gochromedp`
- [ ] `pkg/gopdflib/adapter.go` - audit `bytedance/sonic` under `js/wasm`; if build fails fall back to `encoding/json` behind `//go:build js` - proof: `GOOS=js GOARCH=wasm go build ./cmd/wasmcompress` succeeds and sample `sampledata/compress-js/run.mjs` still runs
- [ ] `internal/pdf/font/provision.go` + `font/pdfa.go` - gate `net/http FetchToTemp` and Liberation download for `js`; document `RegisterFontFromData/RegisterFontFromBase64` as WASM path - proof: `grep -rn os.CreateTemp internal/pdf/font/provision.go` has `!js` guard
- [ ] `internal/pdf/redact/ocr_adapter.go` + `ocr_exec_*.go` - `OCR.Enabled=true` returns `CodeInvalidInput unsupported in WASM` instead of `os/exec` - proof: unit test with OCR flag asserts envelope

### 1.2 Filesystem entry points

- [ ] `internal/pdf/font/registry.go` - verify `RegisterFontFromData` and `RegisterFontFromBase64` need no `os.ReadFile/ReadDir` - proof: code read plus wasm smoke loads TTF bytes
- [ ] `internal/pdf/font/ttf.go` - forbid `LoadTTFFromFile/LoadFontsFromDirectory` in WASM docs - proof: godoc comment updated

## Phase 2: JS API surface

### 2.1 New WASM entrypoint

- [ ] `cmd/wasm/main.go` (or extend `cmd/wasmcompress/main.go`) - expose `goGeneratePDF(templateObj)`, `goMergePDFs(pdfArray)`, `goSplitPDF(bytes, specObj)`, `goFillPDF(pdfBytes, xfdfBytes)`, `goRedact*` alongside existing `goCompressPDF` - proof: `js.Global().Get` lists all five in Node smoke
- [ ] `cmd/wasm/main.go` - use `js.CopyBytesToGo` in and `js.CopyBytesToJS` out plus `recover()` to `{code,message}` via `gopdflib.EnvelopeOf` - proof: mirrors `cmd/wasmcompress/main.go:31-70` pattern
- [ ] `makefile:78-83` - add `wasm:` target building `frontend/public/gopdfsuit.wasm` plus `wasm_exec.js` copy to `frontend/public` and `sampledata/wasm-js/` - proof: `make wasm` plus `file frontend/public/gopdfsuit.wasm`

### 2.2 Per-op bindings

- [ ] `pkg/gopdflib/generator.go` - `GeneratePDF/GeneratePDFBorrowed` callable from WASM with JS template object; keep `WarmRuntimePools` no-op safe single-threaded - proof: generate sampledata fixture in browser harness
- [ ] `pkg/gopdflib/merge.go` - `MergePDFs([][]byte)` takes JS array of `Uint8Array` - proof: merge two `sampledata/merge/*` fixtures in Node
- [ ] `pkg/gopdflib/split.go` - `SplitPDF` plus `ParsePageSpec` accept `{pages, maxPerFile}` object; multi-file return as JS array (zip done in JS) - proof: split fixture returns N `Uint8Array`
- [ ] `pkg/gopdflib/fill.go` - `FillPDFWithXFDF` takes two `Uint8Array`; note `/NeedAppearances` object-stream limit in plan comment - proof: fill `sampledata/filler/*` fixture
- [ ] `pkg/gopdflib/redact.go` - expose `GetPageInfo/ExtractTextPositions/FindTextOccurrences/ApplyRedactions/ApplyRedactionsAdvancedWithReport` text path only - proof: redact search plus apply on text fixture

## Phase 3: Data contracts and limits

- [ ] WASM JS shim - enforce `MaxCompressInputBytes 32MiB` (`pkg/gopdflib/compress.go`, `internal/pdf/compress/limits.go:3-16`) in JS before copy - proof: oversize input rejects without Go call
- [ ] WASM JS shim - fix numeric level bug: map `1|2|3` to `light|medium|heavy` before `goCompressPDF` like `sampledata/compress-js/compress.js:29` (frontend `compressPdf.js:125` passes int today) - proof: light/medium/heavy produce different sizes
- [ ] Template JSON path - pass template as JS object or string; reuse `models.PDFTemplate` validation and `PreallocForDecode`; base64 `imagedata/fontData` stays string, prefer raw `Uint8Array` plus id map for large assets - proof: `sampledata/financialreport/*` generates byte-identical to server
- [ ] Split multi-file - return JS array of `Uint8Array`, never Go-side zip - proof: no `archive/zip` in WASM closure

## Phase 4: Validation and performance

- [ ] `test/verify_pdfs.sh` - WASM-generated Generate/Merge/Split/Fill outputs pass plus `structure_tree_check.py` MCID check - proof: paste script output in ledger
- [ ] Bundle size - record `ls -lh frontend/public/gopdfsuit.wasm` cold load plus `sampledata/benchmarks/` generate p50 before/after - proof: numbers in ledger, release vs dev-loop labeled
- [ ] `sampledata/wasm-js/` - add `index.html` plus `run.mjs` demo mirroring `sampledata/compress-js/` - proof: `node run.mjs` exercises all five bindings

## Phase 5: Closure

- [ ] `plans/wasm/03-wasm-everywhere-noauth-editor.md` - mark Generate/Merge/Split/Fill/Redact rows `[x]` with pointer back to this ledger once validated - proof: no duplicate active rows
- [ ] `documentation/*` plus `frontend/src/components/documentation/content/*` - document which ops are browser-local vs server-only (HTML stays server) - proof: doc diff linked

## Dependencies

- `cmd/wasmcompress/main.go:18-70` existing pattern; `makefile:70-83` build; `frontend/src/utils/compressPdf.js:1-181` loader to generalize
- `pkg/gopdflib/{generator,merge,split,compress,fill,redact,adapter,errors}.go`; `internal/pdf/{generator,merge,compress,redact,form,font,pdfobj,svg}`
- Blockers: `internal/pdf/pdf.go` gochromedp split, sonic `js/wasm` verdict, `gowkhtmltopdf` HTML decision in `plans/wasm/02-gowkhtmltopdf-replace.md`
- Explicit non-goals: `ConvertHTMLToPDF/ConvertHTMLToImage` in WASM, `bindings/python/cgo`, OCR subprocess path
