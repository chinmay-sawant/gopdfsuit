// Loader primitives for Go WASM bundles (browser main thread).
// No op logic here: see sibling modules (compress, generate, document,
// redact, fonts, compliance) plus transports.js for the consent matrix.
// Envelope ownership lives in ./envelope.js (callWasm); caching lives in
// cachedFetch below, shared by fonts.js, templates.js, and (by copy) the
// classic-worker compressWorker.js, which cannot import ESM.
//
// Artifact split (permanent, 2026-09-04): compress.wasm (~8M, only the
// compress binding) stays separate from gopdfsuit.wasm (~31M, full engine).
// Merging would force the full engine into the compress Worker and into
// every CSP worker-src allowlist that today scopes to the small bundle, for
// no behavior gain now that the envelope and loader are unified here.

import { callWasm, callWasmObject, isWasmDebugEnabled, missingEngineError } from './envelope.js'

export { callWasm, callWasmObject, missingEngineError }

const WASM_EXEC_URL = `${import.meta.env.BASE_URL}wasm_exec.js`
export const COMPRESS_WASM_URL = `${import.meta.env.BASE_URL}compress.wasm`
export const GOPDFSUIT_WASM_URL = `${import.meta.env.BASE_URL}gopdfsuit.wasm`

const modulePromises = new Map()

// Network-first fetch for every WASM-adjacent download (binaries, fonts,
// templates manifests). A pure cache-first loader pins the first binary it
// ever saw: Cache Storage survives hard refreshes, so rebuilt engines would
// never reach the page. Network-first lets rebuilt files replace stale
// entries while the stored entry keeps pages working fully offline.
// Falls back to plain fetch where Cache API is unavailable.
export const WASM_CACHE_NAME = 'gopdfsuit-wasm-v1'

export async function cachedFetch(url, { cacheName = WASM_CACHE_NAME, as = 'response' } = {}) {
  const readAs = async (response) => {
    if (!response.ok) throw new Error(`fetch failed: ${url} (${response.status})`)
    if (as === 'json') return response.json()
    if (as === 'bytes') return new Uint8Array(await response.arrayBuffer())
    return response
  }
  let cache = null
  try {
    if (typeof caches !== 'undefined') cache = await caches.open(cacheName)
  } catch {
    cache = null
  }
  if (cache) {
    try {
      // Network first: rebuilt binaries replace stale entries (plain HTTP
      // caching turns unchanged files into cheap 304s). The stored entry
      // keeps pages working fully offline afterwards.
      const response = await fetch(url)
      if (response.ok) {
        if (isWasmDebugEnabled()) console.log('[wasm] engine', url, 'network')
        cache.put(url, response.clone()).catch(() => {})
        return readAs(response)
      }
    } catch {
      // Offline or unreachable: fall through to the cache below.
    }
    const hit = await cache.match(url).catch(() => null)
    if (hit) {
      if (isWasmDebugEnabled()) console.log('[wasm] engine', url, 'cached')
      return readAs(hit)
    }
  }
  const response = await fetch(url)
  return readAs(response)
}

async function fetchCached(url) {
  return cachedFetch(url)
}

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
  const response = await fetchCached(wasmUrl)
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
