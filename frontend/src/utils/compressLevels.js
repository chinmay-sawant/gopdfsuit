// Single source of truth for Compress levels and caps.
// WASM (cmd/wasmcompress) takes 1|2|3; the server (/api/v1/compress) and the
// Go library (compress.Level) take light|medium|heavy. JPEG quality and max
// image edge come from internal/pdf/compress (LevelLight/Medium/Heavy).

export const MAX_COMPRESS_BYTES = 32 * 1024 * 1024

export const COMPRESS_LEVELS = [
  { value: 1, name: 'Light', server: 'light', wasm: 1, jpeg: 92, maxEdge: 1920 },
  { value: 2, name: 'Medium', server: 'medium', wasm: 2, jpeg: 75, maxEdge: 1275 },
  { value: 3, name: 'Heavy', server: 'heavy', wasm: 3, jpeg: 50, maxEdge: 612 },
]

export const DEFAULT_COMPRESS_LEVEL = 2

const levelBy = (predicate) => COMPRESS_LEVELS.find(predicate) || COMPRESS_LEVELS[1]

export const levelByValue = (value) => levelBy((entry) => entry.value === Number(value))

export const toServerLevel = (level) => {
  if (level === undefined || level === null || level === '') return levelByValue(DEFAULT_COMPRESS_LEVEL).server
  if (typeof level === 'number') return levelByValue(level).server
  const key = String(level).trim().toLowerCase()
  const byName = COMPRESS_LEVELS.find((entry) => entry.server === key || entry.name.toLowerCase() === key)
  if (byName) return byName.server
  const asNumber = Number(key)
  if (Number.isFinite(asNumber)) return levelByValue(asNumber).server
  throw new Error(`invalid compression level: ${level} (use 1|2|3 or light|medium|heavy)`)
}

export const toWasmLevel = (level) => {
  if (level === undefined || level === null || level === '') return DEFAULT_COMPRESS_LEVEL
  if (typeof level === 'number') return levelByValue(level).wasm
  const key = String(level).trim().toLowerCase()
  const byName = COMPRESS_LEVELS.find((entry) => entry.server === key || entry.name.toLowerCase() === key)
  if (byName) return byName.wasm
  const asNumber = Number(key)
  if (Number.isFinite(asNumber)) return levelByValue(asNumber).wasm
  throw new Error(`invalid compression level: ${level} (use 1|2|3 or light|medium|heavy)`)
}

export const assertCompressSize = (byteLength) => {
  if (byteLength > MAX_COMPRESS_BYTES) {
    throw new Error(`PDF exceeds maximum size (${MAX_COMPRESS_BYTES} bytes)`)
  }
}

/**
 * Transport selection for compression. Default is in-browser WASM (the file
 * never leaves the device); VITE_COMPRESS_TRANSPORT=server forces the server
 * endpoint, and callers may fall back to the server when WASM is unavailable.
 *
 * Env matrix (see also utils/wasmLoader.js WASM_TRANSPORT):
 * - VITE_COMPRESS_TRANSPORT unset or "wasm" -> compressPDFSmart runs WASM first.
 * - VITE_COMPRESS_TRANSPORT=server -> server endpoint directly.
 * - VITE_WASM_TRANSPORT=server -> merge/split/fill pages go server-first;
 *   compress still honors VITE_COMPRESS_TRANSPORT above.
 * - Either *_TRANSPORT=wasm (default) + allowServerFallback consent -> server
 *   only after an explicit user click, never silently.
 */
export const COMPRESS_TRANSPORT = import.meta.env.VITE_COMPRESS_TRANSPORT || 'wasm'

export const shouldUseServerCompress = () => COMPRESS_TRANSPORT === 'server'

export default COMPRESS_LEVELS
