# Wasm viewer and editor

Date: 2026-09-04. Branch: feat/builder-snippets plus master commits 399e78e, 091d9ee, 5195ccb.

Default path is browser local. Server upload needs explicit consent.

## Artifacts

Two Go artifacts, one JS seam:

- frontend/public/compress.wasm from cmd/wasmcompress/main.go:18. Only goCompressPDF.
- frontend/public/gopdfsuit.wasm from cmd/wasm/main.go:34. Full engine: goGeneratePDF, goMergePDF and goMergePDFs, goSplitPDF, goFillPDF, goCompressPDF, redact page info plus extract plus find plus search plus apply plus advanced, goRegisterFont plus goEnsurePDFAFonts, HTML ops.
- Split is permanent per frontend/src/utils/wasm/core.js:8-12. Merging would force the 31M full engine into the compress worker and into every CSP worker-src rule scoped to the small bundle.

Layers in frontend/src/utils/wasm:

- core.js loads and caches. cachedFetch uses Cache API gopdfsuit-wasm-v1 so modules work offline after first download.
- envelope.js is the sole call gate. callWasm normalizes Uint8Array or Uint8Array array else throws Error with message or error fields. callWasmObject handles status objects.
- Op modules are thin wrappers: generate.js, document.js for merge plus split plus fill, compress.js worker first, redact.js text path only, html.js inline string only, fonts.js, compliance.js, levels.js, templates.js.
- transports.js holds the consent matrix. smartLocal tries localFn first unless VITE_WASM_TRANSPORT is server. Server fallback runs only with allowServerFallback plus auth headers, or after explicit user click.

Editor model is decoupled. documentModel.js holds pure state plus reducers with no WASM or fetch. snippet.js builds Go and Python text from the live template and omits base64.

## Offline vs server

Generate local: generatePDFViaWasm at generate.js:16 calls ensureGopdfsuitWasm plus ensurePDFAFonts when config.pdfaCompliant is set, then goGeneratePDF. Server: generateViaServer at generate.js:35 posts to /api/v1/generate/template-pdf.

Fonts local: ensurePDFAFonts at fonts.js:41 fetches 12 Liberation TTF from /fonts/ through cache gopdfsuit-fonts-v1 and registers each with goRegisterFont. Custom upload calls registerFontLocal at fonts.js:60, wired in Editor.jsx:691 before server upload to POST /api/v1/fonts.

Templates: useBundledTemplate loadTemplateData tries bundled /templates/ first, then falls back to server fetch with auth headers. Shared by Viewer.jsx:27 and Editor.jsx:87.

Failure UX: runLocal takes onError callback. When WASM fails and auth headers exist, Viewer sets fallbackOffer and Editor sets serverRetry, then ConsentBanner shows Upload to server and generate. No silent upload. Code: Viewer.jsx:70-72 and 120-122, Editor.jsx:518-522 and 660-668.

Compress nuance: compressPDFSmart at compress.js:139 prefers worker, falls back to main thread, uses server only with allowServerFallback plus auth. Level mapping lives in levels.js.

Not in WASM: OCR redact, URL HTML fetch with SSRF guard, Chrome HTML path. redactApply rejects OCR.enabled at cmd/wasm/main.go:523 and 559. html.js accepts inline strings only.

## Code

Consent wrapper at transports.js:15-24:

```js
export async function smartLocal(localFn, serverFn, { allowServerFallback = false, getAuthHeaders } = {}) {
  if (shouldUseServerWasmTransport()) return serverFn()
  try {
    return await localFn()
  } catch (wasmError) {
    if (allowServerFallback && getAuthHeaders) return serverFn()
    if (wasmError instanceof Error) wasmError.fallbackAvailable = Boolean(getAuthHeaders)
    throw wasmError
  }
}
```

Call gate at envelope.js:38-42:

```js
export function callWasm(fnName, args, { allowArray = false } = {}) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  return normalizeWasmResult(fnName, fn(...args), { allowArray })
}
```

Viewer preview at Viewer.jsx:52-73:

```jsx
const renderPreview = async (data) => {
  if (serverTransport) {
    await run({ endpoint: '/api/v1/generate/template-pdf', ... })
    return
  }
  let wasmMessage = ''
  const url = await runLocal(() => generatePDFSmart(data, { getAuthHeaders }), {
    autoDownload: false,
    onError: (message) => { wasmMessage = message },
  })
  if (url) return
  if (getAuthHeaders) {
    setFallbackOffer({ message: wasmMessage, data })
  }
}
```

Editor generate at Editor.jsx:492-523:

```js
const handleGeneratePdf = async (isPreview = false) => {
  // Browser-local first via gopdfsuit.wasm ... Server only on explicit retry
  const template = buildTemplate({ config, title, components, footer, bookmarks })
  if (shouldUseServerWasmTransport()) {
    await runServerGenerate(template, isPreview)
    return
  }
  let wasmMessage = ''
  const url = await runLocal(() => generatePDFSmart(template, { getAuthHeaders }), {
    autoDownload: !isPreview, filename: 'generated_document.pdf',
    onBlob: isPreview ? (blob, blobUrl) => { setPdfUrl(blobUrl); setShowPreviewModal(true) } : undefined,
    onError: (message) => { wasmMessage = message },
  })
  if (url) return
  if (getAuthHeaders) { setServerRetry({ message: wasmMessage, template, isPreview }) }
}
```

## Key refs

- core.js:19 COMPRESS_WASM_URL, 20 GOPDFSUIT_WASM_URL, 31 cachedFetch, 140 ensureWasmModule, 153 ensureCompressWasm, 156 ensureGopdfsuitWasm
- envelope.js:9 missingEngineError, 26 normalizeWasmResult, 38 callWasm, 48 callWasmObject
- generate.js:16 generatePDFViaWasm, 35 generateViaServer, 48 generatePDFSmart
- document.js:21 mergePDFViaWasm, 29 splitPDFViaWasm, 34 fillPDFViaWasm, 69 mergePDFSmart, 75 splitPDFSmart, 82 fillPDFSmart
- compress.js:17 compressViaWasmMainThread, 81 compressViaWorker, 106 compressViaServer, 139 compressPDFSmart
- redact.js:18 redactSearchViaWasm, 23 redactApplyViaWasm
- html.js:19 htmlToPDFViaWasm, 25 htmlToImageViaWasm
- fonts.js:41 ensurePDFAFonts, 60 registerFontLocal
- documentModel.js:100 buildTemplate, 215 validateTemplate, 291 insertComponent
- snippet.js:78 cellToGoSnippet, 88 cellToPythonSnippet, 193 templateToGoSnippet, 223 templateToPythonSnippet

## Verify

make wasm-compress, then cd frontend plus npm run build, then make test-integration. Smoke offline by going offline in devtools and checking Cache Storage for gopdfsuit-wasm-v1.
