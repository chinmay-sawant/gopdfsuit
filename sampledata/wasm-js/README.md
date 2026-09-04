# Full-port WASM demo — generate, merge, split, fill, redact

Same pure-Go engine as `gopdflib`, compiled to WebAssembly
(`GOOS=js GOARCH=wasm`). This sample does **not** call any
`POST /api/v1/*` endpoint. The compress-only predecessor is
[`sampledata/compress-js`](../compress-js).

## Where it runs

**Only in the visitor's browser.** You host static files
(`gopdfsuit.wasm`, `wasm_exec.js`, `gopdfsuit.js`). PDFs are read in the
tab, processed in the tab, and downloaded from the tab.

Binding names match `cmd/wasm` (Phase 2.1) and the frontend loader
(`frontend/src/utils/wasmLoader.js`):

| JS wrapper | Go global | Notes |
|------------|-----------|-------|
| `generatePDF(template)` | `goGeneratePDF` | Template object, same shape as the generate API |
| `mergePDFs([a, b])` | `goMergePDF` | Array of `Uint8Array` |
| `splitPDF(bytes, {pages, maxPerFile})` | `goSplitPDF` | Returns JS array of `Uint8Array`; zip in JS, never Go-side |
| `fillPDF(pdf, xfdf)` | `goFillPDF` | Two `Uint8Array` values |
| `redactSearch(bytes, terms)` / `redactApply(...)` | `goRedactSearch` / `goRedactApply` | Text path only, no OCR |

## Run

```bash
make wasm                         # from repo root, once
node sampledata/wasm-js/run.mjs
```

Writes `generated.pdf`, `merged.pdf`, `split_part_N.pdf`, `filled.pdf`,
and `redacted.pdf` next to this file, using fixtures from
`sampledata/merge/` and `sampledata/filler/`.

Browser (files stay in the tab):

```bash
cd sampledata/wasm-js && python3 -m http.server
# open http://localhost:8000
```

## Limits

- Redact is **text path only**. There is no OCR in WASM (no
  pdftoppm/tesseract subprocess in a tab); image-only pages are reported,
  not OCRed. Leave the `OCR` field unset.
- Fill applies `/NeedAppearances true` as a byte-level patch to the
  AcroForm dictionary. If the AcroForm lives inside a compressed object
  stream, viewers may not regenerate appearances on open; field values
  are still written.
- HTML conversion stays **server-side** and has no binding here.
- `wasm_exec.js` is BSD-3-Clause, Copyright The Go Authors. Keep its
  header. Note the Go WASM runtime in LICENSE, NOTICE, or an About page.
