// Inline-HTML render via gopdfsuit.wasm (pure-Go, no upload). URL sources
// stay server-side (fetch plus SSRF guard live in /api/v1/*), so the
// HtmlToPdf/HtmlToImage pages keep their server transport; these helpers
// cover callers that already hold an HTML string.

import { ensureGopdfsuitWasm } from './core.js'

function missingEngineError(fnName) {
  const err = new Error(
    `${fnName} is not in the shipped WASM bundle yet (needs cmd/wasm html bindings)`,
  )
  err.fallbackAvailable = true
  err.missingEngine = true
  return err
}

function callWasmBytes(fnName, args) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  const out = fn(...args)
  if (out instanceof Uint8Array) return out
  const message = out && typeof out === 'object' ? out.error || out.message : undefined
  throw new Error(message || `${fnName} failed`)
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
