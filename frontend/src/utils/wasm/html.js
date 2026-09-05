// HTML and URL render via gopdfsuit.wasm (pure-Go, offline-capable).
// URL sources go through goHtmlToPDF/goHtmlToImage with options.url set;
// the Go binding pre-fetches the page via browser fetch and resolves async,
// so the URL helpers below await a Promise. The fetch is CORS-gated, so only
// CORS-permissive sites convert.

import { ensureGopdfsuitWasm, callWasm, callWasmAsync, missingEngineError } from './core.js'

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

function callWasmURLBytes(fnName, args) {
  try {
    return callWasmAsync(fnName, args)
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

export async function htmlURLToPDFViaWasm(url, options = {}) {
  if (typeof url !== 'string' || url.trim() === '') throw new Error('expected a website URL as a non-empty string')
  await ensureGopdfsuitWasm()
  return callWasmURLBytes('goHtmlToPDF', [url, { ...options, url }])
}

export async function htmlURLToImageViaWasm(url, options = {}) {
  if (typeof url !== 'string' || url.trim() === '') throw new Error('expected a website URL as a non-empty string')
  await ensureGopdfsuitWasm()
  return callWasmURLBytes('goHtmlToImage', [url, { ...options, url }])
}
