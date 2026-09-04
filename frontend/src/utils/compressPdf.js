// In-browser PDF compress via Go WASM. Never POSTs to /api/v1/compress.

const WASM_EXEC_URL = `${import.meta.env.BASE_URL}wasm_exec.js`
const COMPRESS_WASM_URL = `${import.meta.env.BASE_URL}compress.wasm`
const MAX_INPUT_BYTES = 32 * 1024 * 1024 // keep in sync with compress.MaxInputBytes

const LEVEL_MAP = {
  1: 'light',
  2: 'medium',
  3: 'heavy',
  light: 'light',
  medium: 'medium',
  heavy: 'heavy',
}

let initPromise

function mapLevel(level) {
  if (level === undefined || level === null || level === '') return 'medium'
  const key = typeof level === 'string' ? level.trim().toLowerCase() : level
  const mapped = LEVEL_MAP[key]
  if (!mapped) {
    throw new Error(`invalid compression level: ${level} (use 1|2|3 or light|medium|heavy)`)
  }
  return mapped
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

/**
 * Compress PDF bytes in the browser (Go WASM).
 * @param {Uint8Array|ArrayBufferView|ArrayBuffer} uint8 PDF bytes
 * @param {{ level?: 1|2|3|'light'|'medium'|'heavy' }} [opts] default medium (2)
 * @returns {Promise<Uint8Array>}
 */
export async function compressPDF(uint8, opts = {}) {
  const bytes = asUint8Array(uint8)
  if (bytes.byteLength > MAX_INPUT_BYTES) {
    throw new Error(`PDF exceeds maximum size (${MAX_INPUT_BYTES} bytes)`)
  }
  const level = mapLevel(opts == null ? undefined : opts.level)
  await ensureWasm()

  const result = globalThis.goCompressPDF(bytes, level)
  if (result instanceof Uint8Array) return result

  const message = result && typeof result === 'object' ? result.error : undefined
  throw new Error(message || 'PDF compression failed')
}
