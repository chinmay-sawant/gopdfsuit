// In-browser PDF compress via Go WASM, with optional server transport.
// Server endpoint (/api/v1/compress) accepts light|medium|heavy; WASM takes
// the same light|medium|heavy strings (toServerLevel maps 1|2|3 for callers).

import { ensureCompressWasm, asUint8Array, callWasm, COMPRESS_WASM_URL, GOPDFSUIT_WASM_URL } from './core.js'
import { makeAuthenticatedRequest } from '../apiConfig.js'
import {
  MAX_COMPRESS_BYTES,
  assertCompressSize,
  shouldUseServerCompress,
  toServerLevel,
} from './levels.js'

export { MAX_COMPRESS_BYTES }
export { COMPRESS_WASM_URL, GOPDFSUIT_WASM_URL }

async function compressViaWasmMainThread(bytes, level) {
  await ensureCompressWasm()
  return callWasm('goCompressPDF', [bytes, toServerLevel(level)])
}

async function compressViaWasm(bytes, level) {
  // Prefer the Worker path (off main thread); any Worker failure falls
  // back to the main-thread call below so behavior never regresses.
  try {
    return await compressViaWorker(bytes, level)
  } catch {
    // Fall through to main-thread compress.
  }
  // Go WASM takes light|medium|heavy strings; toServerLevel maps the 1|2|3
  // tier numbers callers pass today (the old int path silently fell back to
  // medium because the WASM shim only read string args).
  return compressViaWasmMainThread(bytes, level)
}

// VITE_WASM_WORKER=off forces main-thread compress (e.g. CSP without
// worker-src). Default is the Worker path with main-thread fallback.
const workerEnabled = () => import.meta.env.VITE_WASM_WORKER !== 'off' && typeof Worker !== 'undefined'

let workerInit = null

function ensureWorker() {
  if (workerInit) return workerInit
  workerInit = (async () => {
    const worker = new Worker(new URL('./compressWorker.js', import.meta.url))
    const base = import.meta.env.BASE_URL || '/'
    // Single source of truth for the artifact URL lives in core.js
    // (COMPRESS_WASM_URL); the worker init below passes the resolved name
    // through so main thread and worker can never drift apart.
    const ready = new Promise((resolve, reject) => {
      const onMessage = (event) => {
        const msg = event.data || {}
        if (msg.type === 'ready') {
          worker.removeEventListener('message', onMessage)
          resolve(worker)
        } else if (msg.type === 'init-error') {
          worker.removeEventListener('message', onMessage)
          reject(new Error(msg.error || 'compress worker init failed'))
        }
      };
      worker.addEventListener('message', onMessage)
      worker.addEventListener('error', (event) => reject(event.error || new Error('compress worker error')), { once: true })
    })
    worker.postMessage({
      type: 'init',
      wasmExecUrl: `${base}wasm_exec.js`,
      // COMPRESS_WASM_URL already carries BASE_URL; the worker fetches it
      // as an absolute path so main thread and worker share one constant.
      wasmUrl: COMPRESS_WASM_URL,
    })
    return ready
  })()
  workerInit.catch(() => {
    workerInit = null
  })
  return workerInit
}

let workerSeq = 0

async function compressViaWorker(bytes, level) {
  if (!workerEnabled()) throw new Error('worker path disabled')
  const worker = await ensureWorker()
  const id = workerSeq + 1
  workerSeq = id
  const levelString = toServerLevel(level)
  return new Promise((resolve, reject) => {
    const onMessage = (event) => {
      const msg = event.data || {}
      if (msg.type !== 'result' || msg.id !== id) return
      worker.removeEventListener('message', onMessage)
      if (msg.ok && msg.bytes instanceof Uint8Array) {
        resolve(msg.bytes)
      } else {
        reject(new Error(msg.error || 'PDF compression failed'))
      }
    }
    worker.addEventListener('message', onMessage)
    // Copy before transfer: postMessage detaches the buffer, and the
    // caller's bytes must stay usable for main-thread fallback.
    const owned = bytes.slice()
    worker.postMessage({ type: 'compress', id, level: levelString, bytes: owned }, [owned.buffer])
  })
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
 * Server fallback after a WASM failure only happens with explicit
 * allowServerFallback consent (the Compress page asks the user first);
 * otherwise the WASM error is rethrown with fallbackAvailable set so the
 * UI can offer the upload as a consent click.
 */
export async function compressPDFSmart(uint8, opts = {}, { getAuthHeaders, allowServerFallback = false } = {}) {
  const bytes = asUint8Array(uint8)
  assertCompressSize(bytes.byteLength)
  const level = opts == null ? undefined : opts.level
  if (shouldUseServerCompress()) {
    return compressViaServer(bytes, level, getAuthHeaders)
  }
  try {
    return await compressViaWasm(bytes, level)
  } catch (wasmError) {
    if (allowServerFallback && getAuthHeaders) {
      return compressViaServer(bytes, level, getAuthHeaders)
    }
    if (wasmError instanceof Error) wasmError.fallbackAvailable = Boolean(getAuthHeaders)
    throw wasmError
  }
}
