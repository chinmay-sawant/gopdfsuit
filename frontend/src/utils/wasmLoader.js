// Re-export shim: the WASM loader now lives split by op under ./wasm/.
// This file stays for one release so existing imports keep working; new
// code should import from ./wasm/<op>.js directly.
export {
  asUint8Array,
  loadWasmExec,
  waitForGlobal,
  ensureWasmModule,
  ensureCompressWasm,
  ensureGopdfsuitWasm,
  callWasm,
  callWasmObject,
  missingEngineError,
  cachedFetch,
  WASM_CACHE_NAME,
  COMPRESS_WASM_URL,
  GOPDFSUIT_WASM_URL,
} from './wasm/core.js'
export { shouldUseServerWasmTransport, WASM_TRANSPORT, smartLocal, opSmart } from './wasm/transports.js'
export {
  mergePDFViaWasm,
  splitPDFViaWasm,
  fillPDFViaWasm,
  mergeViaServer,
  splitViaServer,
  fillViaServer,
  mergePDFSmart,
  splitPDFSmart,
  fillPDFSmart,
} from './wasm/document.js'
export { redactSearchViaWasm, redactApplyViaWasm, redactAdvancedViaWasm } from './wasm/redact.js'
export { generatePDFViaWasm } from './wasm/generate.js'
export { generateViaServer, generatePDFSmart } from './wasm/generate.js'
export { BUNDLED_TEMPLATES, loadBundledTemplate } from './wasm/templates.js'
export { htmlToPDFViaWasm, htmlToImageViaWasm } from './wasm/html.js'
export { PDFA_FONT_MANIFEST, ensurePDFAFonts } from './wasm/fonts.js'
export { generateCompliantPDF } from './wasm/compliance.js'
