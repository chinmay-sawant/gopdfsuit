// Shared Go WASM loader for browser-local PDF ops.
// Generalizes frontend/src/utils/compressPdf.js loadWasmExec +
// instantiateStreaming/arrayBuffer fallback + go.run + global wait, so both
// compress.wasm (shipped today) and gopdfsuit.wasm (plans/wasm/01-full-wasm-port.md)
// load through one path. Compress behavior is unchanged: compressPdf.js
// delegates to ensureCompressWasm() here.
//
// Web Worker: [~] deferred. WASM still runs on the main thread (as does
// current compress); moving go.run + copy buffers into a Worker is future
// work. See plans/wasm/03-wasm-everywhere-noauth-editor.md Phase 2.

import { makeAuthenticatedRequest } from './apiConfig'

const WASM_EXEC_URL = `${import.meta.env.BASE_URL}wasm_exec.js`
export const COMPRESS_WASM_URL = `${import.meta.env.BASE_URL}compress.wasm`
// Future full-port bundle from plans/wasm/01-full-wasm-port.md Phase 2.
export const GOPDFSUIT_WASM_URL = `${import.meta.env.BASE_URL}gopdfsuit.wasm`

// Transport matrix (template: compressLevels.js COMPRESS_TRANSPORT):
// VITE_WASM_TRANSPORT=wasm (default) -> browser-local first, server only on
// explicit consent. VITE_WASM_TRANSPORT=server -> server endpoint directly.
// Per-op VITE_COMPRESS_TRANSPORT still wins for compress.
export const WASM_TRANSPORT = import.meta.env.VITE_WASM_TRANSPORT || 'wasm'
export const shouldUseServerWasmTransport = () => WASM_TRANSPORT === 'server'

const modulePromises = new Map()

export function asUint8Array(input, label = 'expected a Uint8Array or ArrayBuffer view of PDF bytes') {
  if (input instanceof Uint8Array) return input
  if (ArrayBuffer.isView(input)) {
    return new Uint8Array(input.buffer, input.byteOffset, input.byteLength)
  }
  if (input instanceof ArrayBuffer) return new Uint8Array(input)
  throw new Error(label)
}

export function loadWasmExec() {
  if (typeof globalThis.Go === 'function') return Promise.resolve()
  if (typeof document === 'undefined') {
    throw new Error('PDF WASM requires a browser document')
  }
  return new Promise((resolve, reject) => {
    const fail = () => reject(new Error(`failed to load ${WASM_EXEC_URL}`))
    const existing = document.querySelector(`script[src="${WASM_EXEC_URL}"]`)
    if (existing) {
      if (typeof globalThis.Go === 'function') {
        resolve()
        return
      }
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', fail, { once: true })
      return
    }
    const script = document.createElement('script')
    script.src = WASM_EXEC_URL
    script.onload = () => resolve()
    script.onerror = fail
    document.head.appendChild(script)
  })
}

async function instantiateWasm(wasmUrl, importObject) {
  const response = await fetch(wasmUrl)
  if (!response.ok) {
    throw new Error(`failed to load ${wasmUrl} (${response.status})`)
  }
  if (typeof WebAssembly.instantiateStreaming === 'function') {
    try {
      return await WebAssembly.instantiateStreaming(response.clone(), importObject)
    } catch {
      // Wrong MIME (not application/wasm): fall back to ArrayBuffer instantiate.
    }
  }
  const bytes = await response.arrayBuffer()
  return WebAssembly.instantiate(bytes, importObject)
}

export function waitForGlobal(fnName, timeoutMs = 15000) {
  if (typeof globalThis[fnName] === 'function') return Promise.resolve()
  return new Promise((resolve, reject) => {
    const start = Date.now()
    const tick = () => {
      if (typeof globalThis[fnName] === 'function') {
        resolve()
        return
      }
      if (Date.now() - start >= timeoutMs) {
        reject(new Error(`Go WASM did not register globalThis.${fnName}`))
        return
      }
      setTimeout(tick, 10)
    }
    tick()
  })
}

async function startModule({ wasmUrl, globalFn, timeoutMs = 15000 }) {
  if (typeof globalThis[globalFn] === 'function') return
  await loadWasmExec()
  if (typeof globalThis[globalFn] === 'function') return
  if (typeof globalThis.Go !== 'function') {
    throw new Error('wasm_exec.js did not define globalThis.Go')
  }
  const go = new globalThis.Go()
  const { instance } = await instantiateWasm(wasmUrl, go.importObject)
  // go.run never resolves: main blocks on select {}. Do not await it.
  go.run(instance)
  await waitForGlobal(globalFn, timeoutMs)
}

export function ensureWasmModule({ key, wasmUrl, globalFn, timeoutMs = 15000 }) {
  if (!modulePromises.has(key)) {
    modulePromises.set(
      key,
      startModule({ wasmUrl, globalFn, timeoutMs }).catch((err) => {
        modulePromises.delete(key)
        throw err
      }),
    )
  }
  return modulePromises.get(key)
}

export const ensureCompressWasm = () =>
  ensureWasmModule({ key: 'compress', wasmUrl: COMPRESS_WASM_URL, globalFn: 'goCompressPDF' })

export const ensureGopdfsuitWasm = () =>
  ensureWasmModule({ key: 'gopdfsuit', wasmUrl: GOPDFSUIT_WASM_URL, globalFn: 'goMergePDF' })

function missingEngineError(fnName) {
  const err = new Error(
    `${fnName} is not in the shipped WASM bundle yet (needs plans/wasm/01-full-wasm-port.md Fill/Merge/Split bindings)`,
  )
  err.fallbackAvailable = true
  err.missingEngine = true
  return err
}

function callWasmArray(fnName, args) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  const result = fn(...args)
  if (result instanceof Uint8Array) return result
  if (Array.isArray(result) && result.every((entry) => entry instanceof Uint8Array)) return result
  const message = result && typeof result === 'object' ? result.error || result.message : undefined
  throw new Error(message || `${fnName} failed`)
}

// Merge / Split / Fill WASM entry points. These resolve once
// plans/wasm/01-full-wasm-port.md lands goMergePDF/goSplitPDF/goFillPDF in
// gopdfsuit.wasm; until then they throw missingEngineError so callers fall
// back to the server with explicit user consent (no silent upload).
export async function mergePDFViaWasm(files) {
  await ensureGopdfsuitWasm()
  const parts = await Promise.all(
    files.map(async (file) => asUint8Array(await file.arrayBuffer())),
  )
  return callWasmArray('goMergePDF', [parts])
}

export async function splitPDFViaWasm(bytes, { pages = '', maxPerFile = '' } = {}) {
  await ensureGopdfsuitWasm()
  return callWasmArray('goSplitPDF', [asUint8Array(bytes), String(pages || ''), String(maxPerFile || '')])
}

export async function fillPDFViaWasm(pdfBytes, xfdfBytes) {
  await ensureGopdfsuitWasm()
  return callWasmArray('goFillPDF', [asUint8Array(pdfBytes), asUint8Array(xfdfBytes)])
}

// Redact search/apply via the WASM text path once the engine lands. Current
// Redaction.jsx runLocal(request(...)) still uploads, so there is no privacy
// win yet; these stubs exist so the page can switch without re-plumbing.
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

// Server transports (explicit-consent fallback only, same as compress).
export async function mergeViaServer(files, getAuthHeaders) {
  const formData = new FormData()
  files.forEach((file) => formData.append('pdf', file))
  const response = await makeAuthenticatedRequest('/api/v1/merge', { method: 'POST', body: formData }, getAuthHeaders)
  const out = new Uint8Array(await (await response.blob()).arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return out
}

export async function splitViaServer(bytes, { pages = '', maxPerFile = '' } = {}, getAuthHeaders) {
  const formData = new FormData()
  formData.append('pdf', new Blob([asUint8Array(bytes)], { type: 'application/pdf' }), 'document.pdf')
  if (pages) formData.append('pages', pages)
  if (maxPerFile) formData.append('max_per_file', maxPerFile)
  const response = await makeAuthenticatedRequest('/api/v1/split', { method: 'POST', body: formData }, getAuthHeaders)
  const out = new Uint8Array(await (await response.blob()).arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return [out]
}

export async function fillViaServer(pdfBytes, xfdfBytes, pdfName, getAuthHeaders) {
  const formData = new FormData()
  formData.append('pdf', new Blob([asUint8Array(pdfBytes)], { type: 'application/pdf' }), pdfName || 'form.pdf')
  formData.append('xfdf', new Blob([asUint8Array(xfdfBytes)], { type: 'application/xml' }), 'data.xfdf')
  const response = await makeAuthenticatedRequest('/api/v1/fill', { method: 'POST', body: formData }, getAuthHeaders)
  const out = new Uint8Array(await (await response.blob()).arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return out
}

// Transport-aware smart wrappers: WASM first with server fallback only on
// explicit allowServerFallback consent (Compress.jsx:83-92 pattern).
async function smartLocal(localFn, serverFn, { allowServerFallback = false, getAuthHeaders } = {}) {
  if (shouldUseServerWasmTransport()) return serverFn()
  try {
    return await localFn()
  } catch (wasmError) {
    if (allowServerFallback && getAuthHeaders) return serverFn()
    if (wasmError instanceof Error) wasmError.fallbackAvailable = Boolean(getAuthHeaders)
    throw wasmError
  }
}

export const mergePDFSmart = (files, opts, transport = {}) =>
  smartLocal(() => mergePDFViaWasm(files), () => mergeViaServer(files, transport.getAuthHeaders), {
    ...transport,
    getAuthHeaders: transport.getAuthHeaders,
  })

export const splitPDFSmart = (bytes, splitOpts, transport = {}) =>
  smartLocal(
    () => splitPDFViaWasm(bytes, splitOpts),
    () => splitViaServer(bytes, splitOpts, transport.getAuthHeaders),
    { ...transport, getAuthHeaders: transport.getAuthHeaders },
  )

export const fillPDFSmart = (pdfBytes, xfdfBytes, pdfName, transport = {}) =>
  smartLocal(
    () => fillPDFViaWasm(pdfBytes, xfdfBytes),
    () => fillViaServer(pdfBytes, xfdfBytes, pdfName, transport.getAuthHeaders),
    { ...transport, getAuthHeaders: transport.getAuthHeaders },
  )
