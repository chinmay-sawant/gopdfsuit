// In-browser PDF compress via Go WASM, with optional server transport.
// Server endpoint (/api/v1/compress) accepts light|medium|heavy; WASM takes 1|2/3.

import { makeAuthenticatedRequest } from './apiConfig'
import {
  MAX_COMPRESS_BYTES,
  assertCompressSize,
  shouldUseServerCompress,
  toServerLevel,
  toWasmLevel,
} from './compressLevels'

const WASM_EXEC_URL = `${import.meta.env.BASE_URL}wasm_exec.js`
const COMPRESS_WASM_URL = `${import.meta.env.BASE_URL}compress.wasm`

export { MAX_COMPRESS_BYTES }

let initPromise

export function mapLevel(level) {
  return toServerLevel(level)
}

function asUint8Array(input) {
  if (input instanceof Uint8Array) return input
  if (ArrayBuffer.isView(input)) {
    return new Uint8Array(input.buffer, input.byteOffset, input.byteLength)
  }
  if (input instanceof ArrayBuffer) {
    return new Uint8Array(input)
  }
  throw new Error('compressPDF expects a Uint8Array or ArrayBuffer view of PDF bytes')
}

function loadWasmExec() {
  if (typeof globalThis.Go === 'function') return Promise.resolve()
  if (typeof document === 'undefined') {
    throw new Error('PDF compression WASM requires a browser document')
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

async function instantiateWasm(importObject) {
  const response = await fetch(COMPRESS_WASM_URL)
  if (!response.ok) {
    throw new Error(`failed to load compress.wasm (${response.status})`)
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

function waitForGoCompressPDF(timeoutMs = 15000) {
  if (typeof globalThis.goCompressPDF === 'function') return Promise.resolve()

  return new Promise((resolve, reject) => {
    const start = Date.now()
    const tick = () => {
      if (typeof globalThis.goCompressPDF === 'function') {
        resolve()
        return
      }
      if (Date.now() - start >= timeoutMs) {
        reject(new Error('Go WASM did not register globalThis.goCompressPDF'))
        return
      }
      setTimeout(tick, 10)
    }
    tick()
  })
}

async function startWasm() {
  if (typeof globalThis.goCompressPDF === 'function') return

  await loadWasmExec()
  if (typeof globalThis.goCompressPDF === 'function') return
  if (typeof globalThis.Go !== 'function') {
    throw new Error('wasm_exec.js did not define globalThis.Go')
  }

  const go = new globalThis.Go()
  const { instance } = await instantiateWasm(go.importObject)
  // go.run never resolves: main blocks on select {}. Do not await it.
  go.run(instance)
  await waitForGoCompressPDF()
}

function ensureWasm() {
  if (!initPromise) {
    initPromise = startWasm().catch((err) => {
      initPromise = undefined
      throw err
    })
  }
  return initPromise
}

async function compressViaWasm(bytes, level) {
  await ensureWasm()

  const result = globalThis.goCompressPDF(bytes, toWasmLevel(level))
  if (result instanceof Uint8Array) return result

  const message = result && typeof result === 'object' ? result.error : undefined
  throw new Error(message || 'PDF compression failed')
}

export async function compressViaServer(bytes, level, getAuthHeaders) {
  assertCompressSize(bytes.byteLength)
  const formData = new FormData()
  formData.append('pdf', new Blob([bytes], { type: 'application/pdf' }), 'document.pdf')
  formData.append('level', toServerLevel(level))
  const response = await makeAuthenticatedRequest('/api/v1/compress', { method: 'POST', body: formData }, getAuthHeaders)
  const blob = await response.blob()
  const out = new Uint8Array(await blob.arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return out
}

/**
 * Compress PDF bytes in the browser (Go WASM).
 * @param {Uint8Array|ArrayBufferView|ArrayBuffer} uint8 PDF bytes
 * @param {{ level?: 1|2|3|'light'|'medium'|'heavy' }} [opts] default medium (2)
 * @returns {Promise<Uint8Array>}
 */
export async function compressPDF(uint8, opts = {}) {
  const bytes = asUint8Array(uint8)
  assertCompressSize(bytes.byteLength)
  const level = opts == null ? undefined : opts.level
  return compressViaWasm(bytes, level)
}

/**
 * Transport-aware compression: WASM first (local, no upload) with server
 * fallback, or server first when VITE_COMPRESS_TRANSPORT=server.
 */
export async function compressPDFSmart(uint8, opts = {}, { getAuthHeaders } = {}) {
  const bytes = asUint8Array(uint8)
  assertCompressSize(bytes.byteLength)
  const level = opts == null ? undefined : opts.level
  if (shouldUseServerCompress()) {
    return compressViaServer(bytes, level, getAuthHeaders)
  }
  try {
    return await compressViaWasm(bytes, level)
  } catch (wasmError) {
    if (!getAuthHeaders) throw wasmError
    return compressViaServer(bytes, level, getAuthHeaders)
  }
}
