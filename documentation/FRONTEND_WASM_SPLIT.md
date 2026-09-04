# Frontend wasm split

Date: 2026-09-04. Commits 5195ccb, 399e78e, b74d5e7.

## Bundle map

Two artifacts served from frontend/public and copied verbatim to docs by Vite with base /gopdfsuit/ in vite.config.js:13 and 16:

- frontend/public/compress.wasm from cmd/wasmcompress. Compress only, about 8M. Loaded by ensureCompressWasm at core.js:153 through worker compressWorker.js with main thread fallback in compress.js:22-34. Used by Compress.jsx only.
- frontend/public/gopdfsuit.wasm full engine for generate, merge, split, fill, redact, html. About 31M. Loaded by ensureGopdfsuitWasm at core.js:156. Used by every other op.
- frontend/public/wasm_exec.js Go runtime glue. Loaded by loadWasmExec at core.js:66 and importScripts in worker at compressWorker.js:17.

JS modules in frontend/src/utils/wasm, 14 files:

- core.js loader primitives only, no op logic.
- envelope.js shared call envelope, browser free so wasm-envelope.test.mjs can pin it.
- compress.js plus compressWorker.js plus levels.js for compress. levels.js is the single source for 1 vs 2 vs 3 and light vs medium vs heavy.
- document.js for merge, split, fill.
- generate.js plus compliance.js plus fonts.js for generate. Compliant generate forces ensurePDFAFonts first.
- templates.js for offline templates.
- transports.js for consent first smartLocal and opSmart.
- redact.js and html.js for remaining ops.

Shims for old imports: utils/compressPdf.js re-exports wasm/compress.js. utils/compressLevels.js re-exports wasm/levels.js. Compress.jsx still imports through shims.

Split stays permanent per core.js:8-12. Merging would push the full engine into the compress worker and into every CSP worker-src rule scoped to the small bundle.

## Loading and cache

cachedFetch at core.js:31-51 uses Cache API gopdfsuit-wasm-v1 with modes json, bytes, response and plain fetch fallback. Worker inlines a copy because a classic worker cannot import ESM at compressWorker.js:22-27.

Font cache is gopdfsuit-fonts-v1 at fonts.js:19. Template cache is gopdfsuit-templates-v1 at templates.js:18.

## Fonts

Vendored: frontend/public/fonts with 12 Liberation TTF plus NOTICE.

Flow:

- FONT_BASE_URL is BASE_URL plus fonts at fonts.js:17.
- fetchFontBytes at fonts.js:21 uses cachedFetch as bytes with direct fetch fallback.
- ensurePDFAFonts at fonts.js:41 calls goEnsurePDFAFonts, fetches only missing faces, registers each with goRegisterFont, re-checks, returns registered plus missing plus fetched.
- registerFontLocal at fonts.js:60 handles user TTF and OTF upload path.
- Server list owned by useFonts at hooks/useFonts.js:18-75. Module cache fontsCache and fontsFetchPromise at 8-9. Initial state uses cache or DEFAULT_FONTS at 19. Failure keeps defaults at 33 and 42. refreshFonts invalidates after upload at 57-72.

## Templates

Vendored: frontend/public/templates with financial_report.json, resume1.json, resume2.json.

- TEMPLATE_BASE_URL is BASE_URL plus templates at templates.js:16.
- loadBundledTemplate at templates.js:36 checks basename against BUNDLED_TEMPLATES, throws with fallbackAvailable true when unbundled at 38 or unreachable offline at 44.
- Hook useBundledTemplate at hooks/useBundledTemplate.js:12-37: github source fetches raw.githubusercontent sampledata at 16, else bundled first at 23, fallbackAvailable falls through to /api/v1/template-data at 27.

## Compliance in browser

generateCompliantPDF at compliance.js:28 calls ensurePDFAFonts, throws listing still missing faces, then calls goGeneratePDF with pdfaCompliant true forced by withPDFACompliant at 13. Header comment at 1 says veraPDF stays server side. This module produces bytes the server check can verify.

Auto path: generate.js:18 calls ensurePDFAFonts whenever config.pdfaCompliant is set.

## Compress page

Compress.jsx:15 reads shouldUseServerCompress once. Badge and copy switch at 104. runSmart at 71 goes server direct when forced else opSmart compressPDF vs compressViaServer. Server runs only from confirmConsentUpload through ConsentBanner at 119. Cap MAX_COMPRESS_BYTES is 32M at levels.js:6 and Compress.jsx:62.

## Snippets

Compress in browser, worker first:

```js
import { compressPDF } from './utils/wasm/compress.js'
const out = await compressPDF(pdfBytes, { level: 2 })
```

Compliant generate, fonts first:

```js
import { generateCompliantPDF } from './utils/wasm/compliance.js'
const pdf = await generateCompliantPDF(template)
```

Offline template with server fallback:

```js
import { useBundledTemplate } from './hooks/useBundledTemplate.js'
const { loadTemplateData } = useBundledTemplate({ runJson, getAuthHeaders, onError })
const data = await loadTemplateData('resume1.json')
```

Manifest gate in package.json plus regen:

```json
{ "scripts": { "prebuild": "node scripts/check-wasm-manifests.mjs" } }
```

```sh
node scripts/check-wasm-manifests.mjs --write
```

## Build gate

Generator scripts/check-wasm-manifests.mjs parses LiberationFontMapping plus LiberationFontFiles from internal/pdf/font/pdfa.go through regex at 17, lists frontend/public/templates json sorted at 36, compares against manifests.generated.js with 12 fonts plus 3 templates. No arg run exits 1 on drift. --write regenerates at 49. Wired as prebuild in frontend/package.json:7 so npm run build fails on drift. Never hand edit generated manifests.
