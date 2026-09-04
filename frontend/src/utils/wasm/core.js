// Loader primitives for Go WASM bundles (browser main thread).
// No op logic here: see sibling modules (compress, generate, document,
// redact, fonts, compliance) plus transports.js for the consent matrix.

const WASM_EXEC_URL = `${import.meta.env.BASE_URL}wasm_exec.js`
export const COMPRESS_WASM_URL = `${import.meta.env.BASE_URL}compress.wasm`
export const GOPDFSUIT_WASM_URL = `${import.meta.env.BASE_URL}gopdfsuit.wasm`

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
