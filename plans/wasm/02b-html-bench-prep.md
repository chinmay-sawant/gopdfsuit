# Phase 4 Prep - HTML Bench Notes (gowkhtmltopdf)

Parent ledger: `plans/wasm/02-gowkhtmltopdf-replace.md` Phase 4.

## Bench inputs (priority order per upstream)

1. Template HTML string: reuse `HtmlToPdf.jsx` sample doc
   (`sampleHtml`) - small styled page with highlight block.
2. Invoice/report HTML files: `sampledata/htmltopdf/` goldens
   (`temp_htmltopdf.pdf`, `temp_htmltopdf_python.pdf`) plus
   `sampledata/htmltoimg/` (`temp_htmltoimage.png`).
3. Server-rendered URL: `https://example.com` (static, no JS) via
   `POST /api/v1/htmltopdf` with `{"url": ...}` (2MiB cap, SSRF guard).

## Reference point

- Upstream reference: ~3.7ms / 2 pages (cold/warm state to be recorded
  locally against this host in the ledger before claiming parity).

## Known regression hypothesis (not a defect)

- JS-heavy SPA URL (e.g. a client-rendered dashboard): script stripping
  means the PDF/image will miss JS-built content. Record as a labeled
  hypothesis with a before/after pair:
  `sampledata/htmltopdf/spa-before-chrome.pdf` (archived Chrome output)
  vs `sampledata/htmltopdf/spa-after-purego.pdf` (new engine output).
  Pair location: `sampledata/htmltopdf/` and `sampledata/htmltoimg/`.

## Measured 2026-09-04 (dev host, cold first-run, includes fetch+layout)

- small-string pdf (Invoice #42 + table, A4): 5924 bytes in 14.4ms
- small-string png (1024x768): 9421 bytes in 38.4ms
- example.com pdf (A4): 7468 bytes in 81.7ms
- example.com png (default size): 15118 bytes in 5.08s
- Upstream reference ~3.7ms/2p is same-host Go bench on a dirty
  report template, not comparable to cold fetch+layout above.

## Gate outcomes 2026-09-04 (branch chore/feature-improves-fixes-wasm)

- `make fmt`: clean, no diff
- `make lint`: clean after fixing pre-existing `revive` unused-param
  (`cache_race_test.go`) and `unused` dead helper (`pdf_utils.go`);
  frontend eslint zero warnings
- `make test`: green - go plus pytest plus `test/verify_pdfs.sh`
  (PDF/A-4 and PDF/UA-2, 10/10 checks). One fix on the way:
  `test_url_to_png` used a Wikipedia URL that lays out 43345px,
  over the engine 16384px raster budget; switched to example.com
  (Go plus Python integration tests) with a budget comment.
- `make test-integration`: green including Zerodha compliance
- `cd frontend && npm run build`: green in 6.01s
- veraPDF re-baseline: template-pipeline goldens still pass; htmltopdf
  goldens now come from the pure-Go engine (unclaimed 1.4 default).
