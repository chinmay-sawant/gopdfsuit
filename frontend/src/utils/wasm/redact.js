// Redact search/apply via the WASM text path (OCR stays server-side).
// Page dims prefer client-side pdfjs (react-pdf already bundles it); the
// functions below cover the text-search plus apply engine calls.

import { ensureGopdfsuitWasm, asUint8Array, callWasm, missingEngineError } from './core.js'

function callRedactWasm(fnName, args) {
  try {
    return callWasm(fnName, args)
  } catch (err) {
    if (err && err.missingEngine) {
      throw missingEngineError(fnName, 'the shipped WASM bundle (needs plans/wasm/01-full-wasm-port.md redact bindings)')
    }
    throw err
  }
}

export async function redactSearchViaWasm(bytes, terms) {
  await ensureGopdfsuitWasm()
  return callRedactWasm('goRedactSearch', [asUint8Array(bytes), terms])
}

export async function redactApplyViaWasm(bytes, blocks, textQueries, mode) {
  await ensureGopdfsuitWasm()
  return callRedactWasm('goRedactApply', [asUint8Array(bytes), blocks, textQueries, mode])
}
