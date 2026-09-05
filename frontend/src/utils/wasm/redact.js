// Redact search/apply via the WASM text path (OCR stays server-side).
// Page dims prefer client-side pdfjs (react-pdf already bundles it); the
// functions below cover the text-search plus apply engine calls.

import { ensureGopdfsuitWasm, asUint8Array, callWasm, callWasmObject, missingEngineError } from './core.js'

function missingRedactEngine(fnName) {
  return missingEngineError(fnName, 'the shipped WASM bundle (needs plans/wasm/01-full-wasm-port.md redact bindings)')
}

function callRedactWasm(fnName, args) {
  try {
    return callWasm(fnName, args)
  } catch (err) {
    if (err && err.missingEngine) {
      throw missingRedactEngine(fnName)
    }
    throw err
  }
}

// Object-returning calls (search rects, apply report) must NOT go through
// callWasm: its envelope only passes Uint8Array through and would reject
// plain arrays/objects as failures.
function callRedactData(fnName, args) {
  try {
    return callWasmObject(fnName, args)
  } catch (err) {
    if (err && err.missingEngine) {
      throw missingRedactEngine(fnName)
    }
    throw err
  }
}

export async function redactSearchViaWasm(bytes, terms) {
  await ensureGopdfsuitWasm()
  return callRedactData('goRedactSearch', [asUint8Array(bytes), terms])
}

export async function redactApplyViaWasm(bytes, blocks, textQueries, mode) {
  await ensureGopdfsuitWasm()
  return callRedactWasm('goRedactApply', [asUint8Array(bytes), blocks, textQueries, mode])
}

export async function redactAdvancedViaWasm(bytes, { blocks = [], textSearch = [], mode = '', password = '' } = {}) {
  await ensureGopdfsuitWasm()
  const queries = (Array.isArray(textSearch) ? textSearch : []).map((item) => (
    typeof item === 'string' ? { text: item } : item
  ))
  return callRedactData('goRedactAdvanced', [asUint8Array(bytes), {
    blocks,
    textSearch: queries,
    mode,
    password,
  }])
}
