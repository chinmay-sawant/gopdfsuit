## Summary

Ships wasm-first Viewer, Editor, and HTML flows plus offline templates and fonts so the UI runs Generate, Merge, Split, Compress, Fill, and text-path Redact fully in-browser via `gopdfsuit.wasm`. Server HTML switches to pure-Go gowkhtmltopdf with no browser dependency, and the engine gets pdfobj write seams, allocator ownership, and guard hardening.

---

## Motivation / context

- Plans: `plans/wasm/01-full-wasm-port.md`, `plans/wasm/02-gowkhtmltopdf-replace.md`, `plans/wasm/03-wasm-everywhere-noauth-editor.md`, `plans/wasm/04-frontend-wasm-split-fonts-compliance.md`, `plans/reviews/architecture-review-2026-09-04-round2.md`
- Issues: none linked per author choice for this chore PR
- Continues epic branch `chore/feature-improves-fixes-wasm` toward offline-first UI and lower server cost. Keeps OCR subprocess and HTML server-side only, moves the rest to WASM with transport consent.

---

## Changes

### Area 1 - Full WASM bundle and frontend split

- New `cmd/wasm/main.go` plus `cmd/wasm/fonts.go` and `cmd/wasm/html.go` exposing Generate, Merge, Split, Compress, Fill, text-path Redact to JS.
- New `frontend/src/utils/wasm/` modules: `core.js`, `document.js`, `generate.js`, `compress.js`, `levels.js`, `fonts.js`, `html.js`, `templates.js`, `transports.js`, `compliance.js`, `redact.js`, plus `compressWorker.js`.
- `frontend/public/gopdfsuit.wasm` (31 MB) plus `compress.wasm`, 12 Liberation fonts with NOTICE, and offline templates `resume1.json`, `resume2.json`, `financial_report.json`.
- Pages `Viewer.jsx`, `Editor.jsx`, `Compress.jsx`, `Merge.jsx`, `Split.jsx`, `Filler.jsx`, `HtmlToPdf.jsx`, `HtmlToImage.jsx`, `Redaction.jsx` rewritten wasm-first with `OperationShell.jsx`, `usePdfOperation.js`, `useToast.js`, `documentModel.js`.
- New `sampledata/wasm-js/` harness with `run.mjs`, `run_compliant.mjs`, fixtures, and README.

### Area 2 - Server HTML and handlers

- HTML conversion swaps to pure-Go `chinmay-sawant/gowkhtmltopdf`, no Chrome or Gotenberg required. Server HTML stays for OCR and complex pages.
- `internal/handlers/handlers.go` split into `compress.go`, `merge_split.go`, `fill.go`, `fonts.go`, `generate.go`, `html.go`, `redact.go`, `template_data.go`, `router.go`, `services.go`, `request.go`, `decode.go`, `json_decode.go`, plus mocks and security, envelope, router limit, template-data tests.
- New `internal/middleware/requestid.go` with request-ID logging, auth scope tightening, CORS fixes.
- `frontend/template.schema.json` plus `scripts/schema-selfcheck.mjs` and Go/Python golden schema gates.

### Area 3 - Engine hardening and docs

- New `internal/pdf/pdfobj/` write seam with `pdfobj.go`, `write.go`, `xref.go` plus tests. Removes `internal/pdf/xref/` and `internal/pdf/bookmarks.go` in favor of allocator-owned output.
- New `internal/pdf/allocator.go`, `destinations.go`, `outline.go`, `structure_tag.go` completion plus `pagemanager.go`, `draw.go`, `helpers.go`, `utils.go`, `metadata.go`, `pdfa.go` updates.
- Guards and fuzz: `merge/xref_guard.go`, `compress/stream_guard_test.go`, `svg/svg_fuzz_test.go` plus corpus, `font/subset_hardening_test.go`, `redact/ocr_timeout_test.go`, `encryption/decrypt.go`, signature security tests.
- XFDF split into `xfdf_fill.go`, `xfdf_locate.go`, `xfdf_parse.go` with locate and guard tests.
- Public API `pkg/gopdflib/` adapter, error taxonomy with envelope, boundary, concurrency, and wasm binding tests. Python bindings updated.
- Docs moved `guides/` to `documentation/` with BENCHMARKS, HTML_CONVERSION, TEMPLATE_REFERENCE updates. CI workflow builds frontend and commits `docs/`.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | In-browser WASM removes server round trips for core ops. Server HTML avoids Chromium spawn. Generation zlib BestSpeed path unchanged. |
| **Memory** | WASM holds full PDFs in browser memory. Server keeps pooled buffers, tier prealloc, 32 MiB input and 48 MiB inflate caps. |
| **Behavior / correctness** | Viewer, Editor, Compress, Merge, Split, Fill, Redact default to local WASM with explicit transport consent. OCR and server HTML remain remote. Template schema triple-gated. |
| **API (`/api/v1/*`) / UI** | No API breaking change. UI adds offline templates, bundled fonts, in-browser compliance check, worker-based compress. |
| **Dependencies** | Adds `chinmay-sawant/gowkhtmltopdf` pure-Go. Drops browser/Gotenberg runtime deps. `go.mod` to Go 1.26.4. |
| **Binary size / build time** | Adds 31 MB `gopdfsuit.wasm` plus 3.7 MB `compress.wasm` and fonts to `frontend/public` and `docs/`. Requires `make wasm` plus `make wasm-compress` before frontend build. |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | Tagging, MCID, ParentTree, XMP paths preserved in WASM and server. Fixtures refreshed and verified via `verify_pdfs.sh` plus structure checks. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Additive WASM entrypoints and UI transports. Existing `pkg/gopdflib` and `/api/v1/*` unchanged. Run `make wasm && make wasm-compress` before `npm run build`. |

---

## Test plan

- [ ] `make test` (`go test ./...` plus Python bindings plus `test/verify_pdfs.sh`)
- [ ] `make test-integration` (`go test -count=1 -v ./test`) when handlers or engine changed
- [ ] `make lint` plus `go vet` (zero ESLint warnings in `frontend/`)
- [ ] `make build` (`go build -o bin/app ./cmd/gopdfsuit`) when shippable change
- [ ] `make test-verify-pdfs` or `make test-scan-pdfs-compliance` when PDF output changed
- [ ] `cd frontend && npm run build` when UI changed (never hand-edit `docs/`)
- [ ] `make wasm-compress` when `cmd/wasmcompress/` changed

### Commands

```sh
make fmt && make lint
make test
# plus when relevant:
make test-integration
make test-verify-pdfs
```

---

## Screenshots / sample output

```
Branch chore/feature-improves-fixes-wasm clean vs origin, 342 files changed vs origin/master.
Latest: 399e78e chore: wasm-first viewer editor html plus offline templates fonts
WASM artifacts: frontend/public/gopdfsuit.wasm (31699617 bytes), frontend/public/compress.wasm, 12 Liberation fonts, 3 offline templates.
Harness: sampledata/wasm-js/run.mjs plus run_compliant.mjs produce generated.pdf, merged.pdf, filled.pdf, redacted.pdf, html.pdf, compliant.pdf.
```

---

## Related issues

- None

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`, `documentation`)
- [x] Related issues filled with real ticket IDs (none per author choice)
- [x] Filled body committed under `plans/PR/pr-wasm-first-viewer-editor.md`

---

## Follow-ups (out of scope)

- Wire WASM compress fully off main thread with Worker promise paths.
- HTML parity bench for gowkhtmltopdf vs Chromium on complex CSS.
- Python CGO exports for new WASM-covered ops where still server-only.
- VeraPDF plus structure-tree re-run on refreshed Zerodha and financial fixtures.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
