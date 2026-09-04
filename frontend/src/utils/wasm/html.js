// Inline-HTML render via gopdfsuit.wasm (pure-Go, no upload). URL sources
// stay server-side (fetch plus SSRF guard live in /api/v1/*), so the
// HtmlToPdf/HtmlToImage pages keep their server transport; these helpers
// cover callers that already hold an HTML string.

import { ensureGopdfsuitWasm, callWasm, missingEngineError } from './core.js'

function callWasmBytes(fnName, args) {
  try {
    return callWasm(fnName, args)
  } catch (err) {
    if (err && err.missingEngine) {
      throw missingEngineError(fnName, 'the shipped WASM bundle (needs cmd/wasm html bindings)')
    }
    throw err
  }
}

export async function htmlToPDFViaWasm(html, options = {}) {
  if (typeof html !== 'string' || html === '') throw new Error('expected HTML content as a string')
  await ensureGopdfsuitWasm()
  return callWasmBytes('goHtmlToPDF', [html, options])
}

export async function htmlToImageViaWasm(html, options = {}) {
  if (typeof html !== 'string' || html === '') throw new Error('expected HTML content as a string')
  await ensureGopdfsuitWasm()
  return callWasmBytes('goHtmlToImage', [html, options])
}
