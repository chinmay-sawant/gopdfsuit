// Shared WASM call envelope. Every Go WASM export returns either raw bytes
// (Uint8Array, or an array of Uint8Array for multi-file ops) or the failure
// envelope {code, message, error} where `error` is a legacy alias of
// `message` kept for callers written before cmd/wasmcompress/main.go pinned
// the shape. This module is browser-free on purpose so node tests
// (frontend/tests/wasm-envelope.test.mjs) can pin the shape without
// import.meta.env or a DOM.

export function missingEngineError(fnName, hint = 'the shipped WASM bundle') {
  const err = new Error(`${fnName} is not in ${hint} yet`)
  err.fallbackAvailable = true
  err.missingEngine = true
  return err
}

export function wasmErrorMessage(result, fallback) {
  if (result && typeof result === 'object') {
    return result.message || result.error || fallback
  }
  return fallback
}

// Normalize one raw Go WASM return into bytes. Accepts a single Uint8Array
// or (when allowArray) an array of Uint8Array; anything else is the failure
// envelope and throws Error(message).
export function normalizeWasmResult(fnName, result, { allowArray = false } = {}) {
  if (result instanceof Uint8Array) return result
  if (allowArray && Array.isArray(result) && result.every((entry) => entry instanceof Uint8Array)) {
    return result
  }
  throw new Error(wasmErrorMessage(result, `${fnName} failed`))
}

// Single entry point for calling a Go WASM global. Owns the envelope:
// missing-global detection, bytes passthrough, and {code,message,error}
// failure mapping. All op modules must route through here instead of
// pasting their own callWasm* prologue.
export function callWasm(fnName, args, { allowArray = false } = {}) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  return normalizeWasmResult(fnName, fn(...args), { allowArray })
}

// Non-bytes variant for calls returning status objects (e.g.
// goEnsurePDFAFonts -> {registered, missing}). Error envelopes always carry
// a string `message` (plus legacy `error` alias), so any object with one is
// a failure; anything else passes through untouched.
export function callWasmObject(fnName, args, { fallback } = {}) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  const out = fn(...args)
  if (out && typeof out === 'object' && !(out instanceof Uint8Array) && !Array.isArray(out)) {
    if (typeof out.message === 'string' || typeof out.error === 'string') {
      throw new Error(out.message || out.error || fallback || `${fnName} failed`)
    }
  }
  return out
}
