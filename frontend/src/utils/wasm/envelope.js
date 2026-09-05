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

// Debug logging for every WASM call. WASM traffic never appears in the
// network tab, so each call logs its request (function name plus summarized
// args) and a result summary with timing. PDF bytes are NEVER logged, only
// their sizes. On by default in browsers; silence with localStorage
// gopdfsuit:wasm-debug = '0'. Always silent outside browsers (node tests).
const MAX_LOGGED_CHARS = 500

function wasmDebugEnabled() {
  try {
    if (typeof localStorage !== 'undefined') {
      const flag = localStorage.getItem('gopdfsuit:wasm-debug')
      if (flag === '0' || flag === 'off') return false
      if (flag === '1' || flag === 'on') return true
    }
  } catch {
    // Storage unavailable (private mode): fall through to the default.
  }
  return typeof window !== 'undefined'
}

export function isWasmDebugEnabled() {
  return wasmDebugEnabled()
}

function summarizeWasmValue(value, depth = 0) {
  if (value instanceof Uint8Array) return `Uint8Array[${value.byteLength} bytes]`
  if (typeof ArrayBuffer !== 'undefined' && value instanceof ArrayBuffer) {
    return `ArrayBuffer[${value.byteLength} bytes]`
  }
  if (typeof value === 'string') {
    return value.length <= 200 ? JSON.stringify(value) : `${JSON.stringify(value.slice(0, 200))}... (${value.length} chars)`
  }
  if (Array.isArray(value)) {
    if (value.length === 0) return '[]'
    if (depth > 0) return `[${value.length} items]`
    const head = value.slice(0, 3).map((entry) => summarizeWasmValue(entry, depth + 1))
    const tail = value.length > 3 ? `, ... (${value.length - 3} more)` : ''
    return `[${head.join(', ')}${tail}]`
  }
  if (typeof value === 'object' && value !== null) {
    try {
      const raw = JSON.stringify(value)
      return raw.length <= MAX_LOGGED_CHARS ? raw : `${raw.slice(0, MAX_LOGGED_CHARS)}... (${raw.length} chars)`
    } catch {
      return '[unserializable object]'
    }
  }
  return String(value)
}

function logWasmRequest(fnName, args) {
  if (!wasmDebugEnabled()) return
  console.log('[wasm] call', fnName, args.map((arg) => summarizeWasmValue(arg)))
}

function logWasmResult(fnName, startedMs, result) {
  if (!wasmDebugEnabled()) return
  console.log('[wasm] result', fnName, `${Date.now() - startedMs}ms`, summarizeWasmValue(result))
}

function logWasmError(fnName, startedMs, err) {
  if (!wasmDebugEnabled()) return
  console.log('[wasm] error', fnName, `${Date.now() - startedMs}ms`, (err && err.message) || String(err))
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
  const startedMs = Date.now()
  logWasmRequest(fnName, args)
  try {
    const result = normalizeWasmResult(fnName, fn(...args), { allowArray })
    logWasmResult(fnName, startedMs, result)
    return result
  } catch (err) {
    logWasmError(fnName, startedMs, err)
    throw err
  }
}

// Non-bytes variant for calls returning status objects (e.g.
// goEnsurePDFAFonts -> {registered, missing}). Error envelopes always carry
// a string `message` (plus legacy `error` alias), so any object with one is
// a failure; anything else passes through untouched.
export function callWasmObject(fnName, args, { fallback } = {}) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  const startedMs = Date.now()
  logWasmRequest(fnName, args)
  try {
    const out = fn(...args)
    if (out && typeof out === 'object' && !(out instanceof Uint8Array) && !Array.isArray(out)) {
      if (typeof out.message === 'string' || typeof out.error === 'string') {
        throw new Error(out.message || out.error || fallback || `${fnName} failed`)
      }
    }
    logWasmResult(fnName, startedMs, out)
    return out
  } catch (err) {
    logWasmError(fnName, startedMs, err)
    throw err
  }
}
