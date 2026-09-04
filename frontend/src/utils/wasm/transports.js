// Transport matrix plus the consent-first smart wrapper shared by every
// browser-local op. Server fallback after a WASM failure only happens with
// explicit allowServerFallback consent (a page asks the user first);
// otherwise the WASM error is rethrown with fallbackAvailable set so the UI
// can offer the upload as a consent click. Nothing uploads silently.

// Transport matrix (template: levels.js COMPRESS_TRANSPORT):
// VITE_WASM_TRANSPORT=wasm (default) -> browser-local first, server only on
// explicit consent. VITE_WASM_TRANSPORT=server -> server endpoint directly.
// Per-op VITE_COMPRESS_TRANSPORT still wins for compress.
export const WASM_TRANSPORT = import.meta.env.VITE_WASM_TRANSPORT || 'wasm'
export const shouldUseServerWasmTransport = () => WASM_TRANSPORT === 'server'

export async function smartLocal(localFn, serverFn, { allowServerFallback = false, getAuthHeaders } = {}) {
  if (shouldUseServerWasmTransport()) return serverFn()
  try {
    return await localFn()
  } catch (wasmError) {
    if (allowServerFallback && getAuthHeaders) return serverFn()
    if (wasmError instanceof Error) wasmError.fallbackAvailable = Boolean(getAuthHeaders)
    throw wasmError
  }
}
