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
