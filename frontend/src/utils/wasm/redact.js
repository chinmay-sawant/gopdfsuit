// Redact search/apply via the WASM text path (OCR stays server-side).
// Page dims prefer client-side pdfjs (react-pdf already bundles it); the
// functions below cover the text-search plus apply engine calls.

import { ensureGopdfsuitWasm, asUint8Array } from './core.js'

function missingEngineError(fnName) {
  const err = new Error(
    `${fnName} is not in the shipped WASM bundle yet (needs plans/wasm/01-full-wasm-port.md redact bindings)`,
  )
  err.fallbackAvailable = true
  err.missingEngine = true
  return err
}

export async function redactSearchViaWasm(bytes, terms) {
  await ensureGopdfsuitWasm()
  const fn = globalThis.goRedactSearch
  if (typeof fn !== 'function') throw missingEngineError('goRedactSearch')
  return fn(asUint8Array(bytes), terms)
}

export async function redactApplyViaWasm(bytes, blocks, textQueries, mode) {
  await ensureGopdfsuitWasm()
  const fn = globalThis.goRedactApply
  if (typeof fn !== 'function') throw missingEngineError('goRedactApply')
  const out = fn(asUint8Array(bytes), blocks, textQueries, mode)
  if (out instanceof Uint8Array) return out
  throw new Error((out && out.error) || 'Redaction failed')
}
