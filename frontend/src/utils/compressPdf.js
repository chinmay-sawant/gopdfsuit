// In-browser PDF compress via Go WASM, with optional server transport.
// Server endpoint (/api/v1/compress) accepts light|medium|heavy; WASM takes
// the same light|medium|heavy strings (toServerLevel maps 1|2|3 for callers).

import { ensureCompressWasm, asUint8Array } from './wasmLoader'
import { makeAuthenticatedRequest } from './apiConfig'
import {
  MAX_COMPRESS_BYTES,
  assertCompressSize,
  shouldUseServerCompress,
  toServerLevel,
} from './compressLevels'

// Loader (loadWasmExec, instantiate fallback, go.run, global wait) lives in
// ./wasmLoader; re-exported here for any direct importers.

export { MAX_COMPRESS_BYTES }

// Shared loader lives in ./wasmLoader (generalized for compress.wasm today
// and gopdfsuit.wasm from plans/wasm/01-full-wasm-port.md). The helpers below
// are kept as thin aliases so this module's public API is unchanged.
const ensureWasm = () => ensureCompressWasm()

async function compressViaWasm(bytes, level) {
  await ensureWasm()

  // Go WASM takes light|medium|heavy strings; toServerLevel maps the 1|2|3
  // tier numbers callers pass today (the old int path silently fell back to
  // medium because the WASM shim only read string args).
  const result = globalThis.goCompressPDF(bytes, toServerLevel(level))
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
