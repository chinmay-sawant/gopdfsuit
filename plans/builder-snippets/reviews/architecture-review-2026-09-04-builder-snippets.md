# Builder-Snippets - Architecture Review (Round 3)

> **Parent:** `plans/builder-snippets/plan.md` - builder-snippets track, branch `feat/builder-snippets`
> **Status:** review 2026-09-04 - 5 subagent walks complete, 0 rows implemented, no code changed
> **Estimated effort:** review 1 session; deepenings sized per row below (S per row, M for Phase 4)

---

## Overview

Round 3 architecture review over the post-builder-snippets tree. Five parallel subagent walks covered engine core (`internal/pdf/` + `pkg/gopdflib/`), API/contracts (`internal/handlers/` + `internal/models/` + middleware), frontend/WASM (`frontend/src/` + `cmd/wasmcompress/`), bindings/HTML/Typst (`bindings/python/` + `typstsyntax/` + gowkhtmltopdf path), and cross-cutting compliance/ops (`test/verify_pdfs.sh` + `plans/` + `documentation/` + `guides/` + makefile + CI).

Hotspots weighted per commit history: `449200c` fluent font/cell builder, `4427b94` copy-clip checklist, `399e78e` wasm-first viewer/editor, `091d9ee` wasm seams + gowkhtmltopdf swap, `38d85a9` round-1 deepenings, `9c08d8a` compress-no-ghostscript.

Prior ledgers read first; nothing below re-litigates them:
- `plans/reviews/architecture-review-2026-09-04.md` - Round 1, closed (all implemented).
- `plans/reviews/architecture-review-2026-09-04-round2.md` (+`.html`) - Round 2 residual friction, open, no implementation started. Rows overlapping it are marked `[~]` below with pointer, no duplicate active rows.
- `plans/reviews/go-review-2026-09-04.md` - Go sweep, closed.
- `plans/adr-2026-09-04-c3-single-budget-rejected.md` - ADR: keep 3 separate budgets. Not revisited.
- Companion visual report: `architecture-review-2026-09-04-builder-snippets.html` in this folder.

## Executive Summary

- **Highest leverage:** Props-string grammar implemented 3 times (`internal/pdf/utils.go:98-185` vs `pkg/gopdflib/props.go` vs `fontbuilder.go` overlay) - one fallback-policy fix diverges silently today. Single canonical grammar package kills the drift class.
- **Cheapest wins:** `internal/pdf/pdf.go` vs `pdf_js.go` near-full duplication (HTML path), handler upload-read copied in 5 handlers, WASM `callWasm*` prologue pasted in 5 JS modules, Python per-op 4-line ctypes prologue in 7 files. Each collapses to one helper.
- **Lies in the API:** HTML `DPI`/`LowQuality`/`Options` accepted in Go + Python + JSON then explicitly ignored before gowkhtmltopdf; Python `Cell.checkbox` serializes as `"chequebox"` (`types.py:337`). Delete or map, plus parity test.
- **Frontend:** 7 op pages paste the same hero/dropzone/consent-banner/how-to shell (~600-800 lines); `Editor.jsx` 1389 lines holds mutations + shortcuts + font cache; `react-pdf` + HeadlessUI paid for but barely used.
- **Engine depth:** `PageManager` facade + dual-mode `Allocator`, `TableTagger` nil-shell, `GenerateTemplatePDFBorrowed` 1200-line wiring with 4 image-decode loops + 3 emit loops, fluent-builder 3 spellings of one cell, owned-type mirror + per-call JSON round-trip adapter, typst vs svg duplicate vector emitters.
- **Ops:** compliance split across 4 severities in one 822-line bash, suite writes `temp_*.pdf` into `sampledata/` (111 PDFs, 25 `temp_*`), `plans/` duplicate `02-` prefix + overlapping PR ledgers, 4 doc homes, makefile gate aliasing + ~60 bench targets, committed `.wasm` duplicates, CI monolith `backend-test`.
- **Counts:** 30 candidates: 11 Strong, 14 Worth exploring / Medium, 5 Speculative. 3 rows deferred to Round 2 (no duplicate active work).

## Phase 1: Correctness, compliance, security first

### 1.1 Canonical props grammar (Strong)

- [ ] `internal/pdf/utils.go:98-185` + `pkg/gopdflib/props.go:38-135` + `pkg/gopdflib/fontbuilder.go:1-232` - single canonical props grammar package (fallback: empty name to Helvetica, size <= 0 to 12, style len-3, unknown align to left), builders become thin constructors, no re-parse round-trip in `SetCell*` - proof: `go test ./internal/pdf ./pkg/gopdflib -run 'Props|Font|Cell|Builder' -v`
- [ ] `pkg/gopdflib/builder.go:169-324` - `MakeProps`/`SetCellFont`/`SetBracketFont` route through canonical grammar, delete parallel fallback comments - proof: cross-boundary case `FontBuilder.Size(0).Props()` equals engine `parseProps` output in new test

### 1.2 HTML convert duplication + phantom options (Strong)

- [ ] `internal/pdf/pdf.go:49-194` vs `internal/pdf/pdf_js.go:25-138` - extract shared `html_convert.go` (`buildPDFDoc`, `buildImageDoc`, `normalizeFormat`, `parseMarginMM`), keep two 5-line source-policy files (`htmlSourceContent` vs `ErrUpstream`) - proof: `go test ./internal/pdf -run TestHTML -v && GOOS=js GOARCH=wasm go build ./internal/pdf`
- [ ] `internal/models/models.go:442-491` + `pkg/gopdflib/types.go:249-280` + `pkg/gopdflib/html.go:32-79` + `bindings/python/pypdfsuit/types.py:541-581` - decide per field (`DPI`, `LowQuality`, `Options`): delete from all mirrors + reject unknown keys, or map to real gowkhtmltopdf knobs with one mapping table; warn on non-empty `Options` until mapped - proof: `go test ./pkg/gopdflib ./internal/models -v` + `pytest bindings/python/tests -k html -v`

### 1.3 Table tagging seam (Strong)

- [ ] `internal/pdf/structure_tag.go:1-80` - delete `TableTagger`, expose one `StructureManager.EmitRowCells(buf, page, count) (base int, end func())`, update ~4 `internal/pdf/draw.go:414-603` call sites, drop per-row `NewTableTagger` alloc on hot path - proof: `go test ./internal/pdf -run 'Table|Tag|Structure' -v` + `make test-verify-pdfs`

### 1.4 Request policy + auth invariant (Worth exploring, security-adjacent)

- [ ] `internal/handlers/handlers.go:92-106` + `internal/middleware/auth.go:16-55` + `internal/handlers/router.go:33-61` - centralize `RoutePolicy{RequireAuth, EnableCORS, BodyLimits, MaxConcurrent}` in `ResolveServerConfig`, share `resolveAudience`/`validateToken`, remove init-cached `isCloudRunCached`, test asserts v1 group always carries auth middleware regardless of `GIN_FAST_API` - proof: `go test ./internal/handlers ./internal/middleware -v`

### 1.5 Python type-mirror typo + parity (Strong)

- [ ] `bindings/python/pypdfsuit/types.py:337` - fix `Cell.checkbox` serializing as `"chequebox"`, collapse bespoke `to_dict` overrides (`Config`, `Cell`, `SignatureConfig`) into `_to_dict` + mapping table, add parity test of `to_dict()` keys vs Go JSON tags (`Config`/`Cell`/HTML requests) - proof: `pytest bindings/python/tests -v`

### 1.6 Error taxonomy (deferred, no duplicate active row)

- [~] Deferred to `plans/reviews/architecture-review-2026-09-04-round2.md` Phase 1 (B1+B2 error taxonomy + envelope) - `internal/handlers/request.go:62-144` `abortRedact*` vs `abortPDFError` copy, `pkg/gopdflib/errors.go:128-167` substring drift (`"limit"`), `pprofForbiddenResp` envelope drift. This ledger adds no active row; fix there first.

## Phase 2: API and data contracts, language boundaries

### 2.1 JSON decode + upload helpers (Strong)

- [ ] `internal/handlers/html.go:14-147` vs `generate.go:11-30` vs `template_data.go:132-138` - one generic `decodeJSONBody[T](c, limit)`, defaults move to `models.HTMLToPDFRequest.SetDefaults/Validate`, keep pooled generate path as override - proof: `go test ./internal/handlers -v`
- [ ] `internal/handlers/redact.go:66-99` + `compress.go:15-36` + `merge_split.go:33-52,69-86` + `fill.go:13-50` + `fonts.go:38-57` - promote `readSingleUpload(c, field, kind)` in `request.go:38-60`, rewrite 5 handlers through it (one 413/400 policy) - proof: `go test ./internal/handlers -run 'Upload|Compress|Merge|Fill|Redact' -v`

### 2.2 Handler testability seam (Strong interface part, Speculative rollout)

- [ ] `internal/handlers/generate.go:35-48` - promote borrowed-render to seam (`FastGenerateService` optional interface or `GenerateTemplatePDFBorrowed` on `PDFService`), mock implements it so hot path is test-covered, delete concrete-type assertion - proof: `go test ./internal/handlers -run TestGenerate -v`
- [ ] `internal/handlers/redact.go:199-263` - extract `parseRedactApply(form)` pure parser (blocks/textSearch/ocr/redactions compat) returning values/errors, no Gin import; handlers become thin adapters - proof: `go test ./internal/handlers -run TestRedact -v`

### 2.3 Models vs gopdflib parity (Worth exploring)

- [ ] `internal/models/models.go:186-500` vs `pkg/gopdflib/types.go:10-388` - keep owned-type split (rationale sound), add `models`-to-`gopdflib` field-parity test beside `schema_golden_test.go` so drift fails CI; route handler inputs through `gopdflib` constructors where they exist (SplitSpec, CompressOptions, HTML) - proof: `go test ./internal/models ./pkg/gopdflib -v`

### 2.4 CGO ownership table (Medium)

- [ ] `bindings/python/cgo/exports.go:288-322` + `pypdfsuit/{fill,compress,html,split,redact}.py` - record ownership (Python: type/shape; CGO: nil/len/cap only; gopdflib: semantics), move `MergePDFs` count cap into `gopdflib`, delete `apply_redactions` empty-list passthrough so Python matches HTTP on edge inputs - proof: `go test ./pkg/gopdflib -v` + `pytest bindings/python/tests -v`

### 2.5 Python ctypes prologue (Strong, small)

- [ ] `bindings/python/pypdfsuit/{generator,merge,split,fill,compress,html,redact}.py` + `_bindings.py:241-310` - add `require_bytes` + `json_payload` + `pdf_args` + `merge_args` helpers in `_bindings.py`, each op becomes 3 lines with uniform empty-input errors - proof: `pytest bindings/python/tests -v`

### 2.6 TemplateBuilder snapshot (Medium)

- [ ] `bindings/python/pypdfsuit/builder.py:410-434` - append placeholder holding `TableBuilder` ref, materialize once in `build()`, drop `ti` index replay so mixed tables/spacers cannot skew - proof: `pytest bindings/python/tests/test_builder.py -v`

## Phase 3: Frontend and WASM seams

### 3.1 Op-page shell (Strong)

- [ ] `frontend/src/pages/{Merge,Split,Compress,Viewer,Filler,HtmlToPdf,HtmlToImage}.jsx` - extract `OpPageShell` + `FileDropzone` + `ConsentBanner`, collapse `HtmlToPdf`/`HtmlToImage` to `HtmlConvertPage({mode})`, delete ~600-800 pasted lines - proof: `cd frontend && npm run lint && npm run build`

### 3.2 WASM call envelope + single artifact (Strong)

- [ ] `frontend/src/utils/wasm/{document,generate,compliance,html,redact,core}.js` + `compressWorker.js:15-69` + `cmd/wasmcompress/main.go` - one `callWasm(fnName, args)` in `core.js` owning Uint8Array vs `{code,message,error}` envelope; merge `compress.wasm` (8.3M) into `gopdfsuit.wasm` (31M) or document why split is permanent; share `fetchCachedWasm` with worker; node test pins envelope shape - proof: `make wasm-compress && cd frontend && npm run build`

### 3.3 Consent transport (Medium-Strong)

- [ ] `frontend/src/utils/wasm/transports.js:14-23` + `levels.js:48-63` + `hooks/usePdfOperation.js:148-198` - pages call `runSmart(opSmart(..., {allowServerFallback: consent}))`, hook owns banner state; unify `runLocal`/`runLocalMulti`; merge `VITE_COMPRESS_TRANSPORT` into `VITE_WASM_TRANSPORT` - proof: `cd frontend && npm run lint && npm run build`

### 3.4 Offline cache (Medium)

- [ ] `frontend/src/utils/wasm/{templates,fonts,core}.js` + `compressWorker.js:26-33` - one `cachedFetch(url, {cacheName, as})` in `core.js`; generate `PDFA_FONT_MANIFEST` + `BUNDLED_TEMPLATES` at build time from `internal/pdf/font/pdfa.go:31-67` + `public/templates/`, fail build on drift - proof: `cd frontend && npm run build`

### 3.5 Preview stack direction (Medium-Strong, needs product call)

- [ ] `frontend/src/components/PdfPreview.jsx` (407) vs `OperationShell.jsx` vs `Viewer.jsx:422-538` - pick one: native iframe + `OperationShell` everywhere (delete `PdfPreview`, drop `react-pdf` unless Redact dims need it) or commit to `react-pdf` Document/Page and delete iframes; drop HeadlessUI (2 usages) or justify - proof: `cd frontend && npm run build` + bundle-size check

### 3.6 Editor split (Medium)

- [ ] `frontend/src/pages/Editor.jsx:1-1389` - move mutations to `components/editor/documentModel.js` pure reducers, shortcuts to `useEditorShortcuts`, template load to shared `useBundledTemplate` (Viewer + Editor + Toolbar), `_fontsCache:41-43,86-120` to `useFonts` context cache - proof: `cd frontend && npm run lint` + pure-function tests for `documentModel`

## Phase 4: Engine depth and shared emission

### 4.1 PageManager + Allocator (Strong)

- [ ] `internal/pdf/pagemanager.go:159-549` + `allocator.go:1-157` - inline `pageObjectStore`/`pageLayoutStore` into `PageManager`, make `Allocator` bound-only (standalone state becomes test fake in `_test.go`) - proof: `go test ./internal/pdf -run 'Page|Alloc' -v`

### 4.2 Generator phase split (Worth exploring)

- [ ] `internal/pdf/generator.go:373-1600` - split `GenerateTemplatePDFBorrowed` into `generation` struct with phase methods (`decodeImages`, `layoutIDs`, `emitPages`, `emitFonts`, `emitTrailer`) sharing `genState`; unify 4 image-decode loops over `(key, *models.Image)` iterator and 3 image-XObject emit loops over one deduped map - proof: `go test ./internal/pdf -v` + `make test-verify-pdfs`

### 4.3 Fluent-builder collapse (Worth exploring, public API churn)

- [ ] `pkg/gopdflib/builder.go:1-324` + `fontbuilder.go:1-232` + `props.go` - keep one cell path (`FontOpts` struct + `Cell` literal + option functions), delete `TableBuilder`/`TitleTableBuilder` aliasing (`builder.go:90-98` vs `120-128`), replace `WithTitleFont` positional varargs with options struct, deprecation pass for public API - proof: `go test ./pkg/gopdflib -v`

### 4.4 Owned-type adapter cost (Worth exploring, breaking-change decision)

- [ ] `pkg/gopdflib/types.go:1-388` + `adapter.go:20-57` - type-alias public types to internal (`type PDFTemplate = models.PDFTemplate`) with stability docs, or `go generate` the mirror + field-coverage test; at minimum hoist `redact.go:31-188` per-item JSON translation out of loops - proof: `go test ./pkg/gopdflib -v` + benchmark before/after on generate path

### 4.5 Shared vector emission (Strong, two emitters)

- [ ] `typstsyntax/renderer.go:1138-1349` + `internal/pdf/svg/svg.go:17-24,269-330,416-501` - extract `internal/pdf/vector/` (`WriteFloat`, `EscapeText`, `SetFill/SetStroke`, `StrokeLine/StrokePath`) used by both, one float/color/escape policy - proof: `go test ./typstsyntax ./internal/pdf/svg ./internal/pdf -v`

## Phase 5: Compliance, tests, plans, docs, gates

### 5.1 Compliance manifest (Strong)

- [ ] `test/verify_pdfs.sh` (822) + `test/structure_tree_check.py` + `test/verapdf_report.py` + `makefile:111-157` - extract `test/compliance_manifest.json` (path, baseline, tolerance, flavours, avalStrict), thin bash to runner, per-directory aval strictness (zerodha strict, contract-notes warn), `zerodha_compliance_test.go` reads same manifest - proof: `make test-verify-pdfs`

### 5.2 Test pyramid + hermetic outputs (Strong)

- [ ] `test/integration_test.go` + `integration_coverage/misc/xfdf/redact/financial_report_test.go` + `test/helpers_test.go` - document pyramid in `documentation/INTEGRATION_AND_BENCHMARK_TESTS.md`, route suite outputs to `t.TempDir()` or gitignored `test/output/`, keep `sampledata/` for baselines only, delete `TestCoverage*` shims duplicating package tests - proof: `make test-integration` leaves `git status --short` clean of `sampledata/` writes

### 5.3 Plans index + filename collisions (Strong, cheap)

- [ ] `plans/` - add `plans/INDEX.md` (closed vs active table), rename `plans/wasm/02-html-bench-prep.md` to `02b-` or fold as appendix, merge `plans/PR/pr-frontend-auto-build.md` + `pr-frontend-auto-build-push-only.md` keeping push-only survivor - proof: docs-only, grep gate shows zero duplicate active `[ ]` rows

### 5.4 Doc homes ADR (Strong, cheap)

- [ ] `documentation/` vs `guides/` vs `docs/` vs `plans/` - record doc-home ADR (`documentation/` truth, `guides/` frozen archive, `docs/` generated, `plans/` decisions), add header pointers, stop updating `guides/MAKEFILE.md` - proof: docs-only, link check passes

### 5.5 Makefile tiers (Worth exploring)

- [ ] `makefile:77-157,260-604` - collapse `test-verify`/`test-verify-pdfs` to one name, define `test-fast` / pre-push / full `test` / nightly tiers, move ~60 bench targets under `bench-help` generation - proof: `make -n <tier>` prints expected gate set, docs-only plus gate rename

### 5.6 Sampledata split (Strong)

- [ ] `sampledata/` (111 PDFs, 25 `temp_*`, 7 `generated.*`) - split `{fixtures,outputs,goldens}`, gitignore `temp_*`, stop committing wasm under `sampledata/compress-js` + `sampledata/wasm-js` (CI builds via `make wasm-compress`), decide `gopdflib/zerodha` vs `gpdf/zerodha` survivor - proof: fresh suite run shows all `temp_*` ignored via `git check-ignore`

### 5.7 CI split + toolchain pin (Worth exploring)

- [ ] `.github/workflows/frontend-build-commit.yml:247` - split `backend-test` into `go-test`/`python-test`/`verify-pdfs` with shared validator-setup action, drop `needs: backend-lint` for frontend lint, pin Go 1.26.4 everywhere (fixes `python-build` 1.22 skew), single staged Dockerfile replacing `Dockerfile_cloudrun` fork - proof: one red job maps to exactly one layer

## Phase 6: Closure gates

- [ ] `make fmt && make lint && make test` - full Go + Python + PDF compliance - proof: paste output in handoff
- [ ] `make test-integration` - handler suite standalone - proof: paste output
- [ ] `make test-verify-pdfs` - veraPDF + structure tree on fixtures - proof: `test/verify_pdfs.sh` output
- [ ] `make wasm-compress && cd frontend && npm run build` - WASM + frontend bundle - proof: paste output

## Dependencies

- Phase 1 before Phase 2 (canonical grammar + HTML parity pin the contracts handlers rely on).
- Phase 2 before Phase 3 (Python/CGO ownership decided before frontend snippet emission changes).
- Phase 4 engine splits are independent of Phases 2-3 but need Phase 6 gates per row.
- Phase 5 ops cleanup is independent; do `plans/INDEX.md` + doc-home ADR first (cheap, unblocks navigation).
- Phase 6 depends on all prior implemented rows. Do not mark rows `[x]` without the listed proof command passing.
- Round 2 pointer: `plans/reviews/architecture-review-2026-09-04-round2.md` owns error-taxonomy/envelope, allocator bypass, pdfobj/xref merge, cache doctrine. No duplicate active rows live here.

## Non-goals

- No `CONTEXT.md`/ADR relitigation beyond the doc-home ADR in 5.4.
- No unified budget (rejected in `plans/adr-2026-09-04-c3-single-budget-rejected.md`).
- No `Cell.Segments` rich-text, no Versa importer (per parent `plans/builder-snippets/plan.md`).
- No hand-edit of `docs/` Vite output.
- No `unsafe` audit claimed: precise negative results only (e.g. `font/` split and `merge/`/`compress/`/`redact/` per-op seams checked, not nominated).

## Completion Handoff

- [ ] Confirm rows above with commands beside each gate, synchronize this ledger, report next unchecked phase. Do not create a second status document.
