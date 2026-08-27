// PDF compress via Go WASM. Never fetch('/api/v1/compress').
//
//   import { compressPDF } from './compress.js'
//   const out = await compressPDF(uint8, { level: 2 })
//   // level 1 = light, 2 = medium, 3 = heavy

const LEVELS = {
  1: 'light',
  2: 'medium',
  3: 'heavy',
  light: 'light',
  medium: 'medium',
  heavy: 'heavy',
}

const isNode = typeof process !== 'undefined' && Boolean(process.versions?.node)
const MAX_INPUT_BYTES = 32 * 1024 * 1024 // keep in sync with compress.MaxInputBytes

let initPromise

/**
 * Compress PDF bytes in-process through compress.wasm.
 * @param {Uint8Array|ArrayBuffer} bytes
 * @param {{ level?: 1|2|3|'light'|'medium'|'heavy' }} [options]
 * @returns {Promise<Uint8Array>}
 */
export async function compressPDF(bytes, options = {}) {
  await ensureWasm()
  const level = LEVELS[options.level ?? 2]
  if (!level) {
    throw new Error(`unknown compress level: ${options.level}`)
  }
  const input = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  if (input.byteLength > MAX_INPUT_BYTES) {
    throw new Error(`PDF exceeds maximum size (${MAX_INPUT_BYTES} bytes)`)
  }
  const out = globalThis.goCompressPDF(input, level)
  if (out && typeof out === 'object' && out.error) {
    throw new Error(String(out.error))
  }
  if (!(out instanceof Uint8Array)) {
    throw new Error('goCompressPDF did not return a Uint8Array')
  }
  return out
}

function missingWasm(detail) {
  return new Error(
    `${detail}\nRun \`make wasm-compress\` first to build compress.wasm and wasm_exec.js.`,
  )
}

async function ensureWasm() {
  if (typeof globalThis.goCompressPDF === 'function') {
    return
  }
  if (!initPromise) {
    initPromise = loadWasm().catch((err) => {
      initPromise = undefined
      throw err
    })
  }
  await initPromise
}

async function polyfillGoRuntime() {
  if (!globalThis.crypto?.getRandomValues) {
    if (!isNode) {
      throw new Error('globalThis.crypto.getRandomValues is required')
    }
    const { webcrypto } = await import('node:crypto')
    globalThis.crypto = webcrypto
  }
  if (typeof globalThis.performance?.now !== 'function') {
    globalThis.performance = { now: () => Date.now() }
  }
}

async function readWasmAssets() {
  if (isNode) {
    const { existsSync, readFileSync } = await import('node:fs')
    const { dirname, join } = await import('node:path')
    const { fileURLToPath } = await import('node:url')
    const here = dirname(fileURLToPath(import.meta.url))
    const pairs = [
      [join(here, 'compress.wasm'), join(here, 'wasm_exec.js')],
      [join(here, '../../frontend/public/compress.wasm'), join(here, '../../frontend/public/wasm_exec.js')],
    ]
    for (const [wasmPath, execPath] of pairs) {
      if (existsSync(wasmPath) && existsSync(execPath)) {
        return {
          wasmBytes: readFileSync(wasmPath),
          execSource: readFileSync(execPath, 'utf8'),
          execLabel: execPath,
        }
      }
    }
    throw missingWasm('compress.wasm / wasm_exec.js not found next to this file or in frontend/public/')
  }

  const bases = [new URL('./', import.meta.url), new URL('../../frontend/public/', import.meta.url)]
  for (const base of bases) {
    const wasmUrl = new URL('compress.wasm', base)
    const execUrl = new URL('wasm_exec.js', base)
    try {
      const [wasmRes, execRes] = await Promise.all([fetch(wasmUrl), fetch(execUrl)])
      if (!wasmRes.ok || !execRes.ok) {
        continue
      }
      return {
        wasmBytes: new Uint8Array(await wasmRes.arrayBuffer()),
        execSource: await execRes.text(),
        execLabel: execUrl.href,
      }
    } catch {
      continue
    }
  }
  throw missingWasm('failed to fetch compress.wasm / wasm_exec.js')
}

async function evalWasmExec(source, filename) {
  if (typeof globalThis.Go === 'function') {
    return
  }
  if (isNode) {
    const vm = await import('node:vm')
    vm.runInThisContext(source, { filename })
    return
  }
  const run = new Function(source)
  run()
}

async function loadWasm() {
  await polyfillGoRuntime()
  const { wasmBytes, execSource, execLabel } = await readWasmAssets()
  await evalWasmExec(execSource, execLabel)
  if (typeof globalThis.Go !== 'function') {
    throw missingWasm('globalThis.Go missing after loading wasm_exec.js')
  }

  const go = new globalThis.Go()
  const result = await WebAssembly.instantiate(wasmBytes, go.importObject)
  // Go main blocks on select{}; awaiting go.run hangs forever.
  go.run(result.instance)
  if (typeof globalThis.goCompressPDF !== 'function') {
    throw missingWasm('goCompressPDF was not registered by compress.wasm')
  }
}
