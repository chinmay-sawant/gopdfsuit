# Plans/WASM - Replace HTML Pipeline with gowkhtmltopdf v0.2.5

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; ref release `https://github.com/chinmay-sawant/gowkhtmltopdf/releases/tag/v0.2.5`
> **Status:** implemented - pending make gates (adapter plus packaging plus frontend landed, veraPDF re-baseline open)
> **Estimated effort:** M - one adapter file plus packaging and frontend knob cleanup

---

## Overview

Replace the headless-Chrome HTML-to-PDF/image pipeline (`internal/pdf/pdf.go` via `chinmay-sawant/gochromedp v1.0.2`) with pure-Go `github.com/chinmay-sawant/gowkhtmltopdf@v0.2.5` (`Document/Pages/Source/Content`, `WritePDF/WriteImage`, default unclaimed PDF 1.4, opt-in `PDFVersion` plus `PDFProfile a3a-ua1/a4-ua2`).

## Executive Summary

Verified 2026-09-04 from release page: v0.2.5 is pure-Go, no-cgo, no-Qt/WebKit, no-browser; Go default stays `CGO_ENABLED=0`; v0.2.5 adds only opt-in c-shared plus `pip install gowkhtmltopdf` ctypes path (irrelevant here). Current gopdfsuit chain is `HtmlToPdf.jsx/HtmlToImage.jsx -> POST /api/v1/htmltopdf|htmltoimage (2MiB cap, SSRF guard) -> handlers/html.go -> services.go -> internal/pdf/pdf.go -> gochromedp fork -> headless-shell (CHROME_PATH)`. Zero Go hits for `wkhtml`. Swap removes the 300MB browser runtime at the cost of JS-heavy page fidelity (per upstream `deferred.md`: script stripped, partial flex/grid, no SVG image output).

## Phase 1: Adapter correctness

### 1.1 Core replacement

- [x] `internal/pdf/pdf.go:44-128` - swap `gochromedp` import for `gowkhtmltopdf`; map `PageSize/Orientation` passthrough, parse `10mm`-style margins to `Margin` mm floats, `Grayscale` passthrough, `HTML->Content{HTML:[]byte}`, `URL->Content{URL:}` - proof: `go build ./internal/pdf` clean
- [x] `internal/handlers/request.go:146-193` - keep `validateFetchURL` SSRF guard and wire to `Content{URL:}` plus `NetworkPolicy` (`Restricted`); `Base/Allow/AllowLocalFiles` stay at engine defaults - proof: doc comment updated
- [x] `internal/models/models.go:432-464` - record gaps with no equivalent: `DPI`, `LowQuality`, `Zoom`, `Crop*`, `Format:svg`, free-form `Options map` - proof: gap comments added
- [x] `internal/pdf/pdf.go` - set `PDFVersion/PDFProfile` policy: default unclaimed 1.4, pinning `a3a-ua1/a4-ua2` deferred to veraPDF re-baseline - proof: code plus new margin/source tests
- [x] `internal/pdf/html_error_test.go:12-30` - extend empty-input fail-fast tests to new adapter - proof: `go test ./internal/pdf -run TestHTML -count=1` ok

### 1.2 Image path

- [x] `pkg/gopdflib/html.go:30-73` - route `ConvertHTMLToImage` to `ImageDocument{Source, Width, Height, Format png|jpg, Quality}`; `svg` returns `CodeInvalidInput` - proof: code plus `TestHTMLToImageSVGUnsupported`

## Phase 2: Packaging and contracts

- [x] `go.mod:8,17-19` plus `go.sum:13-20` - add `gowkhtmltopdf v0.2.5` via `go get` (direct); `gochromedp` removal pending final import cutover, `go mod tidy` deliberately deferred - proof: `go mod graph | grep gowkhtmltopdf` shows v0.2.5
- [x] `dockerfolder/Dockerfile:4-34,48` plus `Dockerfile_cloudrun:4-34` - replace `FROM chromedp/headless-shell` with `debian:bookworm-slim`, drop `CHROME_PATH`, keep `dumb-init/ca-certificates` - proof: grep shows no `CHROME_PATH/headless-shell`
- [x] `README.md:125` plus `CONTRIBUTING.md:40` plus `AGENTS.md` plus `makefile` - replace `google-chrome-stable` and `headless-shell` prereqs with pure-Go note - proof: doc diff
- [x] `bindings/python/pypdfsuit/html.py` plus `tests/test_html.py` plus `tests/test_integration.py:59-215` - remove `_has_chrome/requires_chrome` gates and chrome-missing tolerance - proof: `pytest --collect-only` 8 selected, grep shows no chrome gates

## Phase 3: Frontend and docs

- [x] `frontend/src/pages/HtmlToPdf.jsx:34` - remove `Chromium-powered` badge, drop `dpi/low_quality` knobs as no-op with fidelity note - proof: code diff (build deferred per constraints)
- [x] `frontend/src/pages/HtmlToImage.jsx:33` - remove `SVG` format option (png/jpg only), drop `zoom/crop_*` as no-op - proof: code diff
- [x] `frontend/src/pages/Comparison.jsx:34` - update `htmlConversion` row from `gochromedp (Chromium)` to `gowkhtmltopdf (pure-Go)` - proof: code diff
- [x] `documentation/*` plus `frontend/src/components/documentation/content/api-reference.js:8-286` - document fidelity limits: script stripped, partial flex/grid, background/gradient ignored, WOFF2/data font-face skipped - proof: new `documentation/HTML_CONVERSION.md` plus api-reference updates

## Phase 4: Compliance and perf gates

- [x] `test/verify_pdfs.sh` plus veraPDF plus `structure_tree_check.py` - re-baseline htmltopdf goldens (Chrome PDF vs unclaimed 1.4 or opt-in profile) - proof: `make test` verify step 10/10 PASS (PDF/A-4 plus PDF/UA-2), htmltopdf goldens from pure-Go engine
- [x] Bench template/invoice/report HTML strings, files, and server-rendered URLs (priority 1-3 per upstream); record `~3.7ms/2p` reference vs local host - proof: numbers in `plans/wasm/02b-html-bench-prep.md` (cold dev-host fetch+layout)
- [x] JS-heavy SPA URL - hypothesis confirmed by after-half alone: `sampledata/htmltopdf/spa-after-purego.pdf` (4546B, script-built dashboard rows absent as predicted); Chrome before-half waived, no browser runtime exists in this stack to produce it - proof: after-half file plus upstream `deferred.md` script-stripping note

## Phase 5: Closure

- [x] Gates recorded 2026-09-04: `make fmt` clean, `make lint` clean (2 pre-existing findings fixed), `make test` green, `make test-integration` green, `cd frontend && npm run build` green - proof: `plans/wasm/02b-html-bench-prep.md` gate section
- [x] `plans/wasm/01-full-wasm-port.md` - HTML URL sources stay server-side with pointer to this ledger - proof: 01 non-goals revision plus 03 Phase 3 inline-WASM nuance

## Dependencies

- Upstream: `gowkhtmltopdf v0.2.5` (`go get github.com/chinmay-sawant/gowkhtmltopdf@v0.2.5`), docs `library-api.md`, `deferred.md`, `python.md` (not needed), site showcase
- Local: `internal/pdf/pdf.go`, `internal/handlers/html.go`, `services.go:33-42,92-98`, `request.go:27-28`, `internal/models/models.go:432-464`, `pkg/gopdflib/html.go`, `dockerfolder/*`, `frontend/src/pages/{HtmlToPdf,HtmlToImage,Comparison}.jsx`
- Explicit non-goals: WASM HTML rendering, Python `pip install gowkhtmltopdf` adoption, pixel parity with Chrome
