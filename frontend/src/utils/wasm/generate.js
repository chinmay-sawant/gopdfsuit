// Template generate via gopdfsuit.wasm. Accepts a template object or JSON
// string; the Go shim stringifies objects itself, so callers pass through.
// Large binary assets stay base64 imagedata/fontData strings, matching the
// server template contract. For the compliant variant (fonts ensured first)
// see ./compliance.js.

import { ensureGopdfsuitWasm } from './core.js'

function callWasmPdf(fnName, args) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') {
    const err = new Error(`${fnName} is not in the shipped WASM bundle yet`)
    err.fallbackAvailable = true
    err.missingEngine = true
    throw err
  }
  const result = fn(...args)
  if (result instanceof Uint8Array) return result
  const message = result && typeof result === 'object' ? result.error || result.message : undefined
  throw new Error(message || `${fnName} failed`)
}

export async function generatePDFViaWasm(template) {
  await ensureGopdfsuitWasm()
  return callWasmPdf('goGeneratePDF', [template])
}
