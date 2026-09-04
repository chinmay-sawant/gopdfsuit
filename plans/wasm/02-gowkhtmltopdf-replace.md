# Plans/WASM - Replace HTML Pipeline with gowkhtmltopdf v0.2.5

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; ref release `https://github.com/chinmay-sawant/gowkhtmltopdf/releases/tag/v0.2.5`
> **Status:** planning - HTML path is gochromedp/Chromium today, zero `wkhtml` in Go code
> **Estimated effort:** M - one adapter file plus packaging and frontend knob cleanup

---

## Overview

Replace the headless-Chrome HTML-to-PDF/image pipeline (`internal/pdf/pdf.go` via `chinmay-sawant/gochromedp v1.0.2`) with pure-Go `github.com/chinmay-sawant/gowkhtmltopdf@v0.2.5` (`Document/Pages/Source/Content`, `WritePDF/WriteImage`, default unclaimed PDF 1.4, opt-in `PDFVersion` plus `PDFProfile a3a-ua1/a4-ua2`).

## Executive Summary

Verified 2026-09-04 from release page: v0.2.5 is pure-Go, no-cgo, no-Qt/WebKit, no-browser; Go default stays `CGO_ENABLED=0`; v0.2.5 adds only opt-in c-shared plus `pip install gowkhtmltopdf` ctypes path (irrelevant here). Current gopdfsuit chain is `HtmlToPdf.jsx/HtmlToImage.jsx -> POST /api/v1/htmltopdf|htmltoimage (2MiB cap, SSRF guard) -> handlers/html.go -> services.go -> internal/pdf/pdf.go -> gochromedp fork -> headless-shell (CHROME_PATH)`. Zero Go hits for `wkhtml`. Swap removes the 300MB browser runtime at the cost of JS-heavy page fidelity (per upstream `deferred.md`: script stripped, partial flex/grid, no SVG image output).

## Phase 1: Adapter correctness

### 1.1 Core replacement

- [ ] `internal/pdf/pdf.go:44-128` - swap `gochromedp` import for `gowkhtmltopdf`; map `PageSize/Orientation` passthrough, parse `10mm`-style margins to `Margin` mm floats, `Grayscale` passthrough, `HTML->Content{HTML:[]byte}`, `URL->Content{URL:}` - proof: `go build ./internal/pdf`
- [ ] `internal/handlers/request.go:146-193` - keep `validateFetchURL` SSRF guard and wire to `Content{URL:}` plus `NetworkPolicy` (`Compatible` vs `Restricted`); decide `Base/Allow/AllowLocalFiles` for subresource CSS - proof: blocked-host test still 4xx
- [ ] `internal/models/models.go:432-464` - record gaps with no equivalent: `DPI`, `LowQuality`, `Zoom`, `Crop*`, `Format:svg`, free-form `Options map` - proof: comment plus handler no-op or 400 matrix in ledger
- [ ] `internal/pdf/pdf.go` - set `PDFVersion/PDFProfile` policy: keep default unclaimed 1.4 or pin 1.7/2.0 plus `a3a-ua1/a4-ua2` to match gopdfsuit PDF/A-4 and PDF/UA-2 claims - proof: veraPDF output cited in ledger
- [ ] `internal/pdf/html_error_test.go:12-30` - extend empty-input fail-fast tests to new adapter - proof: `go test ./internal/pdf -run HTML -count=1`

### 1.2 Image path

- [ ] `pkg/gopdflib/html.go:30-73` - route `ConvertHTMLToImage` to `ImageDocument{Source, Width, Height, Format png|jpg, Quality}`; drop `svg` output or return `CodeInvalidInput` - proof: png plus jpg golden outputs

## Phase 2: Packaging and contracts

- [ ] `go.mod:8,17-19` plus `go.sum:13-20` - drop `chinmay-sawant/gochromedp` plus `chromedp/*` indirect, add `gowkhtmltopdf v0.2.5` via `go get` plus `go mod tidy` - proof: `go mod graph | grep -i -E 'chromedp|gowkhtmltopdf'`
- [ ] `dockerfolder/Dockerfile:4-34,48` plus `Dockerfile_cloudrun:4-34` - replace `FROM chromedp/headless-shell` with slim/distroless, drop `CHROME_PATH`, keep `dumb-init/ca-certificates` - proof: `docker build` smoke plus image size in ledger
- [ ] `README.md:125` plus `CONTRIBUTING.md:40` plus `AGENTS.md` plus `makefile` - replace `google-chrome-stable` and `headless-shell` prereqs with pure-Go note - proof: doc diff
- [ ] `bindings/python/pypdfsuit/html.py` plus `tests/test_html.py` plus `tests/test_integration.py:59-215` - remove `_has_chrome/requires_chrome` gates and chrome-missing tolerance - proof: `pytest bindings/python -k html`

## Phase 3: Frontend and docs

- [ ] `frontend/src/pages/HtmlToPdf.jsx:34` - remove `Chromium-powered` badge, drop `dpi/low_quality` knobs or mark no-op - proof: `cd frontend && npm run build`
- [ ] `frontend/src/pages/HtmlToImage.jsx:33` - remove `SVG` format option (png/jpg only), drop `zoom/crop_*` or mark no-op - proof: `cd frontend && npm run build`
- [ ] `frontend/src/pages/Comparison.jsx:34` - update `htmlConversion` row from `gochromedp (Chromium)` to `gowkhtmltopdf (pure-Go)` - proof: UI screenshot or build
- [ ] `documentation/*` plus `frontend/src/components/documentation/content/api-reference.js:8-286` - document fidelity limits: script stripped, partial flex/grid, background/gradient ignored, WOFF2/data font-face skipped - proof: doc diff

## Phase 4: Compliance and perf gates

- [ ] `test/verify_pdfs.sh` plus veraPDF plus `structure_tree_check.py` - re-baseline htmltopdf goldens (Chrome PDF vs unclaimed 1.4 or opt-in profile) - proof: paste verifier output
- [ ] Bench template/invoice/report HTML strings, files, and server-rendered URLs (priority 1-3 per upstream); record `~3.7ms/2p` reference vs local host - proof: numbers with cold/warm state in ledger
- [ ] JS-heavy SPA URL - record as known regression with hypothesis label, not defect - proof: before/after PDF pair in `sampledata/`

## Phase 5: Closure

- [ ] `make fmt && make lint && make test && make test-integration` - record outcomes in ledger; leave rows unchecked on failure - proof: pasted gate output
- [ ] `plans/wasm/01-full-wasm-port.md` - confirm HTML stays server-side `[~]` with pointer to this ledger - proof: no duplicate active HTML-in-WASM rows

## Dependencies

- Upstream: `gowkhtmltopdf v0.2.5` (`go get github.com/chinmay-sawant/gowkhtmltopdf@v0.2.5`), docs `library-api.md`, `deferred.md`, `python.md` (not needed), site showcase
- Local: `internal/pdf/pdf.go`, `internal/handlers/html.go`, `services.go:33-42,92-98`, `request.go:27-28`, `internal/models/models.go:432-464`, `pkg/gopdflib/html.go`, `dockerfolder/*`, `frontend/src/pages/{HtmlToPdf,HtmlToImage,Comparison}.jsx`
- Explicit non-goals: WASM HTML rendering, Python `pip install gowkhtmltopdf` adoption, pixel parity with Chrome
