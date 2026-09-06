# WASM Frontend: Local-First PDF in the Browser

The React frontend (`frontend/`, app `gopdfsuit-frontend`) runs the same Go
engine as the server inside the browser via two WebAssembly artifacts. Most
operations complete without any upload. The server is used only as an explicit,
per-click fallback.

Related deep dives:

- [WASM viewer and editor](WASM_VIEWER_EDITOR.md) - generate and preview loop,
  bundled templates, font delivery for compliant output.
- [Frontend WASM split](FRONTEND_WASM_SPLIT.md) - split UX, page-spec parsing,
  multi-file download and zip packaging in JS.

## Artifacts

Both files are build output committed under `frontend/public/`.

| Artifact | Size on disk | Source | JS exports |
|----------|--------------|--------|------------|
| `compress.wasm` | ~8.3M | `cmd/wasmcompress/main.go` | `goCompressPDF` only |
| `gopdfsuit.wasm` | ~31M | `cmd/wasm/` (`main.go`, `fonts.go`, `html.go`) | full engine (see below) |
| `wasm_exec.js` | ~17K | Go toolchain (`lib/wasm/wasm_exec.js`) | `globalThis.Go` runtime bridge |

Why two bundles (permanent split): `compress.wasm` keeps the compress Web
Worker small and keeps CSP `worker-src` allowlists scoped to the small bundle.
Merging would force the full engine into every compress path for no behavior gain.

### `gopdfsuit.wasm` export surface (`cmd/wasm/main.go`)

| Global | Op | Notes |
|--------|----|-------|
| `goGeneratePDF` | Generate from template JSON | object or JSON string; base64 assets match server contract |
| `goMergePDFs` (`goMergePDF` alias) | Merge | JS array of `Uint8Array` in, one `Uint8Array` out |
| `goSplitPDF` | Split | returns JS Array of `Uint8Array`; zipping stays in JS |
| `goFillPDF` | AcroForm/XFDF fill | same semantics as server; does not rewrite compressed object streams |
| `goCompressPDF` | Compress | accepts `light\|medium\|heavy`, `1\|2\|3`, numeric strings |
| `goRedactGetPageInfo` | Redact page info | returns `{totalPages, pages}` |
| `goRedactExtractText` | Redact text positions | `(bytes, pageNum)` |
| `goRedactFindText` (`goRedactSearch` alias) | Redact text search | returns rects |
| `goRedactApply` | Apply redactions | rejects `OCR.Enabled` (server-only) |
| `goRedactAdvanced` | Apply with report | returns `{pdf, report}`; rejects `OCR.Enabled` |
| `goRegisterFont` | Font upload path | `(name, bytes)`, TTF/OTF |
| `goEnsurePDFAFonts` | Compliant-generate status | `{registered, missing}` over the standard faces |
| `goHtmlToPDF` | HTML/URL to PDF | sync for inline HTML; Promise in URL mode |
| `goHtmlToImage` | HTML/URL to image | sync for inline HTML; Promise in URL mode |

Failure shape for every export: `{code, message, error}`, where `error` is a
legacy alias of `message`. Normalized in one place:
`frontend/src/utils/wasm/envelope.js` (`callWasm`, `callWasmAsync`,
`callWasmObject`). All op modules route through it.

## Page route table

Routes from `frontend/src/App.jsx` (HashRouter; pages under `frontend/src/pages/`):

| Route | Page file | WASM path (local-first) | Server fallback endpoint |
|-------|-----------|-------------------------|--------------------------|
| `/` | `Home.jsx` | none (landing) | - |
| `/viewer` | `Viewer.jsx` | `generatePDFSmart` (`goGeneratePDF`) | `POST /api/v1/generate/template-pdf` |
| `/editor` | `Editor.jsx` | `generatePDFSmart` (`goGeneratePDF`) | `POST /api/v1/generate/template-pdf` |
| `/merge` | `Merge.jsx` | `mergePDFSmart` (`goMergePDF`) | `POST /api/v1/merge` |
| `/split` | `Split.jsx` | `splitPDFSmart` (`goSplitPDF`) | `POST /api/v1/split` |
| `/compress` | `Compress.jsx` | `compressPDFSmart` (`goCompressPDF`, Worker-first) | `POST /api/v1/compress` |
| `/filler` | `Filler.jsx` | `fillPDFSmart` (`goFillPDF`) | `POST /api/v1/fill` |
| `/htmltopdf` | `HtmlToPdf.jsx` via shared `HtmlConvertPage.jsx` (`mode='pdf'`) | `htmlToPDFViaWasm` (`goHtmlToPDF`) | `POST /api/v1/htmltopdf` |
| `/htmltoimage` | `HtmlToImage.jsx` via shared `HtmlConvertPage.jsx` (`mode='image'`) | `htmlToImageViaWasm` (`goHtmlToImage`) | `POST /api/v1/htmltoimage` |
| `/redact` | `Redaction.jsx` | `redactSearchViaWasm` / `redactAdvancedViaWasm` (text path only) | redact endpoints incl. OCR |
| `/screenshots` | `Screenshots.jsx` | none (gallery) | - |
| `/comparison` | `Comparison.jsx` | none (comparison view) | - |
| `*` | `NotFound.jsx` | none | - |

## Local-first plus consent model

Rule: nothing uploads silently.

- Default transport is browser-local. `VITE_WASM_TRANSPORT=server`
  forces the server endpoint for every op. `VITE_COMPRESS_TRANSPORT=server`
  overrides compress alone.
- Smart wrappers (`compressPDFSmart`, `mergePDFSmart`, `splitPDFSmart`,
  `fillPDFSmart`, `generatePDFSmart`) run the WASM path first with server
  fallback off. On failure the error is rethrown with `fallbackAvailable`
  so the UI can offer the upload as a consent click.
- Server tasks run only from explicit consent, rendered through the
  shared `ConsentBanner` component.
- Compress runs off the main thread: the classic Worker uses
  `compress.wasm` only, with main-thread fallback on any Worker failure.
- Compress tiers: `1 Light` (JPEG 92, edge 1920), `2 Medium`
  (JPEG 75, edge 1275, default), `3 Heavy` (JPEG 50, edge 612). Cap:
  32 MiB, enforced before WASM and server alike.

## Offline caches

No service worker. Offline works through the Cache Storage API with
network-first reads: rebuilt binaries replace stale entries over the network
while the stored entry keeps pages working fully offline afterwards.

| Cache name | Contents | Populated by |
|------------|----------|--------------|
| `gopdfsuit-wasm-v1` | `compress.wasm`, `gopdfsuit.wasm`, `wasm_exec.js`, manifest fetches | loader `cachedFetch` |
| `gopdfsuit-fonts-v1` | Liberation TTFs under `/fonts/` for compliant generate | font ensure step (fetched once, registered via `goRegisterFont`) |
| `gopdfsuit-templates-v1` | bundled sample templates under `/templates/` | template loader |

## What stays server-only

- OCR redaction: `goRedactApply`/`goRedactAdvanced` reject
  `OCR.Enabled` options outright. Scanned pages needing OCR go through the
  server redact chain, offered only via the consent banner.
- veraPDF compliance validation stays server-side.
- URL HTML conversion without CORS headers fails locally; the consent banner
  offers the server fetch instead.
- XFDF fill on compressed object streams: same limitation as the server path.

## Rebuild commands

```bash
# Small bundle: compress.wasm + wasm_exec.js into frontend/public/
make wasm-compress

# Full bundle: gopdfsuit.wasm + wasm_exec.js into frontend/public/
make wasm

# Frontend bundle (prebuild verifies WASM manifests, then Vite builds)
cd frontend && npm run build

# Frontend lint (must be zero warnings)
cd frontend && npm run lint
```
