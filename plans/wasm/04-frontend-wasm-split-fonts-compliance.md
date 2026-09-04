# Plans/WASM - Frontend JS Split plus Fonts plus In-Browser Compliance

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; refs `plans/wasm/01-full-wasm-port.md`, `plans/wasm/03-wasm-everywhere-noauth-editor.md`
> **Status:** implemented - split landed, fonts vendored, compliant generate proven (PDF/A-4 plus PDF/UA-2 PASS on browser bytes)
> **Estimated effort:** M - JS re-split plus 12-file font manifest plus one Go binding plus compliance smoke

---

## Overview

Split the frontend WASM JS by PDF op (one module per feature), ship the 12 Liberation TTFs the engine needs, and prove `pdfaCompliant` generate works in the browser with subsets embedded from JS-registered fonts. Validation (veraPDF) stays server-side.

## Executive Summary

Today `frontend/src/utils/wasmLoader.js` (253 lines) mixes loader primitives, transport matrix, merge/split/fill/redact bindings, and server fallbacks, while `compressPdf.js` mixes compress plus the Worker path. Nothing ships fonts: `internal/pdf/font/pdfa.go:204-207` rejects `EnsureFontsAvailable` in WASM, so any `pdfaCompliant:true` template fails in-browser. The good news, verified 2026-09-04: generation consults the registry first (`internal/pdf/generator.go:2377-2389`, `draw.go:94,614`), so JS-registered fonts short-circuit the download path entirely. The 12 needed faces total ~1.6MB (Sans ~574KB, Serif ~596KB, Mono ~456KB, per the `gowkhtmltopdf/internal/pdf/assets/` reference set). Only used glyphs embed (`SubsetTTFForText`), so output stays small. `DEFAULT_FONTS` (`editor/constants.js:5`) is display names only, no bytes.

## Phase 1: Frontend JS split (correctness first, no behavior change)

### 1.1 New module layout under `frontend/src/utils/wasm/`

- [x] `wasm/core.js` - loader primitives only - proof: `npx eslint` clean plus `npm run build` green
- [x] `wasm/compress.js` - `compressPdf.js` plus `compressLevels.js` plus `compressWorker.js` wiring - proof: Compress page unchanged (same imports via shims), Worker path intact
- [x] `wasm/generate.js` - template-JSON generate - proof: build green; compliant preset lives in `wasm/compliance.js`
- [x] `wasm/document.js` - merge/split/fill plus `*Smart` transports - proof: Merge/Split/Filler page imports unchanged via shim
- [x] `wasm/redact.js` - search/apply text path - proof: Redaction page unchanged
- [x] `wasm/transports.js` - `smartLocal` plus server fallbacks plus `WASM_TRANSPORT` matrix - proof: pages import via `wasmLoader.js` shim, build green
- [x] `wasmLoader.js` plus `compressPdf.js` plus `compressLevels.js` - re-export shims - proof: page imports untouched, full `npm run lint` zero warnings
- [x] `wasm/html.js` (added beyond plan) - inline-HTML `goHtmlToPDF`/`goHtmlToImage` via `wasmLoader.js` shim; URL sources stay server-side - proof: eslint plus build green, Node smoke `html.pdf` 7350B plus `html.png` 11799B

### 1.2 Demo shim parity

- [x] `sampledata/wasm-js/gopdfsuit.js` (309 lines) - stays single-file by decision: it is a throwaway Node harness, not shipped UI code; extended with `registerFont`/`ensurePDFAFonts` plus `EXPECTED_BINDS` - proof: `node run.mjs` 5/5 plus `node run_compliant.mjs` green

## Phase 2: Font delivery

### 2.1 Manifest and sourcing

- [x] `frontend/public/fonts/` - 12 Liberation TTFs (2.1.5 set) plus OFL `NOTICE`, 1.6M total - proof: `ls` 13 files present
- [x] Rejected alternatives recorded: go:embed (11M to ~13M always-paid), CDN fetch (offline break plus SRI), pdfjs `standard_fonts` reuse (version drift) - proof: this row
- [x] Engine SHA pin (`pdfa.go:76`) gates only the server auto-download tarball, not registry content - proof: compliant smoke embeds subsets from JS-registered 2.1.5 faces below

### 2.2 Runtime registration

- [x] `cmd/wasm/main.go` - `goRegisterFont(name, bytes)` over `RegisterFontFromData` plus `goEnsurePDFAFonts()` reporting registered/missing (new `cmd/wasm/fonts.go`, also splits the 566-line entrypoint) - proof: `GOOS=js GOARCH=wasm go build ./cmd/wasm` plus vet clean
- [x] `wasm/fonts.js` (new) - fetch-on-demand from `/fonts/*.ttf`, Cache API first with network fallback, `ensurePDFAFonts()` - proof: smoke registers 12/12 below
- [x] `wasm/compliance.js` (new) - `generateCompliantPDF(template)`: `ensurePDFAFonts()` then generate with `pdfaCompliant:true`, missing faces throw before generation - proof: compliant smoke below

## Phase 3: In-browser compliance proof

- [x] Compliance smoke - `node sampledata/wasm-js/run_compliant.mjs`: 12/12 faces registered, financial_report with `pdfaCompliant:true` plus `pdfTitle` generates 111728B with `%PDF-`, `%%EOF`, embedded FontFile subset - proof: output above
- [x] Server-side veraPDF on the browser-generated file - PDF/A-4 PASS (109 rules, 2816 checks), PDF/UA-2 PASS (1727 rules, 8069 checks), `structure_tree_check.py` ParentTree TD=52/52 TR 0 - proof: pasted verdicts (first UA-2 run failed only on missing `dc:title`, fixed via `pdfTitle`)
- [x] Document what stays server-side - proof: `GETTING_STARTED_GOPDFLIB.md` table gained compliant-generate plus veraPDF rows

## Phase 4: Closure gates

- [x] `npx eslint` zero warnings plus `npm run build` green (never hand-edit `docs/`) - proof: lint clean, build green 7.29s
- [x] `go test ./pkg/gopdflib -run 'TestGenerate|TestMerge|TestSplit' -count=1` plus `GOOS=js GOARCH=wasm go build ./cmd/wasm` - proof: ok 0.223s plus WASM_OK
- [x] Mark `01` Phase 3 template-JSON row and `03` Worker row with pointers to this ledger - proof: pointers added (see 01/03 edits)

## Dependencies

- Engine: `pkg/gopdflib/generator.go` (`GeneratePDFFromJSON`), `internal/pdf/font/{registry.go:76-89,pdfa.go:197-207}`, `internal/pdf/generator.go:2377-2389`, `internal/pdf/draw.go:94,614`
- Reference font set: `gowkhtmltopdf/internal/pdf/assets/` (12 Liberation TTFs ~1.6MB plus NOTICE) and tarball `pdfa.go:71` (SHA `pdfa.go:76`)
- Frontend: `utils/wasmLoader.js:1-253`, `utils/compressPdf.js`, `utils/compressWorker.js`, `utils/compressLevels.js`, `editor/constants.js:5`
- Explicit non-goals: veraPDF in browser, server download path changes, DejaVu/Unicode fallback shipping
