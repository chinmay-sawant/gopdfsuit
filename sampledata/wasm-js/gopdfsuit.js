// Browser/Node shim for gopdfsuit.wasm (Generate, Merge, Split, Compress,
// Fill, text-path Redact). Never fetch('/api/v1/*').
//
//   import { compressPDF, mergePDFs } from './gopdfsuit.js'
//   const out = await compressPDF(uint8, { level: 2 })
//   // level 1 = light, 2 = medium, 3 = heavy
//
// Build the bundle with `make wasm`, which copies gopdfsuit.wasm and
// wasm_exec.js next to this file (and into frontend/public/).
//
// Contracts (plans/wasm/01-full-wasm-port.md Phase 3):
// - Inputs larger than MAX_INPUT_BYTES (32 MiB, mirroring
//   gopdflib.MaxCompressInputBytes) reject in JS before any bytes cross into
//   Go, so oversize files never pay the CopyBytesToGo copy.
// - Numeric levels 1|2|3 map to light|medium|heavy before goCompressPDF, like
//   sampledata/compress-js/compress.js. The Go binding also accepts tier
//   numbers defensively, but the canonical call passes strings.
// - Split returns a JS Array of Uint8Array, one entry per output part. There
//   is no archive/zip in Go; package parts in JS when a single download is
//   needed.

const LEVELS = {
  1: 'light',
  2: 'medium',
  3: 'heavy',
  light: 'light',
  medium: 'medium',
  heavy: 'heavy',
}

const isNode = typeof process !== 'undefined' && Boolean(process.versions?.node)
// Keep in sync with gopdflib.MaxCompressInputBytes (32 MiB).
const MAX_INPUT_BYTES = 32 * 1024 * 1024

export { MAX_INPUT_BYTES }

let initPromise

function asUint8Array(bytes, label = 'expected a Uint8Array or ArrayBuffer of PDF bytes') {
  const input = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  if (!(input instanceof Uint8Array)) throw new Error(label)
  return input
}

function assertSize(input, label = 'PDF') {
  if (input.byteLength > MAX_INPUT_BYTES) {
    throw new Error(`${label} exceeds maximum size (${MAX_INPUT_BYTES} bytes)`)
  }
}

function assertPdfOut(out, fnName) {
  if (out && typeof out === 'object' && out.error) {
    throw new Error(String(out.error))
  }
  if (!(out instanceof Uint8Array)) {
    throw new Error(`${fnName} did not return a Uint8Array`)
  }
  return out
}

function toLevelString(level) {
  const mapped = LEVELS[level ?? 2]
  if (!mapped) throw new Error(`unknown compress level: ${level}`)
  return mapped
}

/**
 * Compress PDF bytes in-process through gopdfsuit.wasm.
 * @param {Uint8Array|ArrayBuffer} bytes
 * @param {{ level?: 1|2|3|'light'|'medium'|'heavy' }} [options]
 * @returns {Promise<Uint8Array>}
 */
export async function compressPDF(bytes, options = {}) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goCompressPDF(input, toLevelString(options.level))
  return assertPdfOut(out, 'goCompressPDF')
}

/**
 * Generate a PDF from a template object (same shape as POST /api/v1/generate/template-pdf).
 * Template JSON may carry base64 imagedata/fontData strings as on the server.
 * @param {object|string} template
 * @returns {Promise<Uint8Array>}
 */
export async function generatePDF(template) {
  await ensureWasm()
  const out = globalThis.goGeneratePDF(template)
  return assertPdfOut(out, 'goGeneratePDF')
}

/**
 * Merge PDFs in order.
 * @param {Array<Uint8Array|ArrayBuffer>} files
 * @returns {Promise<Uint8Array>}
 */
export async function mergePDFs(files) {
  await ensureWasm()
  if (!Array.isArray(files) || files.length === 0) {
    throw new Error('mergePDFs needs at least 1 PDF file')
  }
  const inputs = files.map((f) => {
    const input = asUint8Array(f)
    assertSize(input)
    return input
  })
  const out = globalThis.goMergePDFs(inputs)
  return assertPdfOut(out, 'goMergePDFs')
}

/**
 * Split a PDF. Spec accepts {pages, ranges, maxPerFile} or a "1-3,5"
 * page-spec string. Returns one Uint8Array per part (never a zip).
 * @param {Uint8Array|ArrayBuffer} bytes
 * @param {object|string} [spec]
 * @returns {Promise<Uint8Array[]>}
 */
export async function splitPDF(bytes, spec = {}) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goSplitPDF(input, spec)
  if (out && typeof out === 'object' && out.error) {
    throw new Error(String(out.error))
  }
  if (!Array.isArray(out) || !out.every((entry) => entry instanceof Uint8Array)) {
    throw new Error('goSplitPDF did not return an array of Uint8Array')
  }
  return out
}

/**
 * Fill a PDF form with XFDF bytes.
 * @param {Uint8Array|ArrayBuffer} pdfBytes
 * @param {Uint8Array|ArrayBuffer|string} xfdfBytes
 * @returns {Promise<Uint8Array>}
 */
export async function fillPDF(pdfBytes, xfdfBytes) {
  await ensureWasm()
  const pdf = asUint8Array(pdfBytes)
  assertSize(pdf)
  const xfdf = typeof xfdfBytes === 'string' ? new TextEncoder().encode(xfdfBytes) : asUint8Array(xfdfBytes)
  const out = globalThis.goFillPDF(pdf, xfdf)
  return assertPdfOut(out, 'goFillPDF')
}

/**
 * Text-path redact helpers (OCR stays server-side).
 */
export async function getPageInfo(bytes) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goRedactGetPageInfo(input)
  if (out && typeof out === 'object' && out.error) throw new Error(String(out.error))
  return out
}

export async function extractTextPositions(bytes, pageNum = 1) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goRedactExtractText(input, pageNum)
  if (out && typeof out === 'object' && !Array.isArray(out) && out.error) {
    throw new Error(String(out.error))
  }
  return out
}

export async function findTextOccurrences(bytes, searchText) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goRedactFindText(input, searchText)
  if (out && typeof out === 'object' && !Array.isArray(out) && out.error) {
    throw new Error(String(out.error))
  }
  return out
}

export async function applyRedactions(bytes, blocks) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goRedactApply(input, blocks)
  return assertPdfOut(out, 'goRedactApply')
}

export async function applyRedactionsAdvanced(bytes, options) {
  await ensureWasm()
  const input = asUint8Array(bytes)
  assertSize(input)
  const out = globalThis.goRedactAdvanced(input, options)
  if (out && typeof out === 'object' && out.error) throw new Error(String(out.error))
  if (!out || !(out.pdf instanceof Uint8Array)) {
    throw new Error('goRedactAdvanced did not return {pdf, report}')
  }
  return out
}

function missingWasm(detail) {
  return new Error(`${detail}\nRun \`make wasm\` first to build gopdfsuit.wasm and wasm_exec.js.`)
}

async function ensureWasm() {
  if (typeof globalThis.goGeneratePDF === 'function') return
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
    if (!isNode) throw new Error('globalThis.crypto.getRandomValues is required')
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
      [join(here, 'gopdfsuit.wasm'), join(here, 'wasm_exec.js')],
      [join(here, '../../frontend/public/gopdfsuit.wasm'), join(here, '../../frontend/public/wasm_exec.js')],
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
    throw missingWasm('gopdfsuit.wasm / wasm_exec.js not found next to this file or in frontend/public/')
  }

  const bases = [new URL('./', import.meta.url), new URL('../../frontend/public/', import.meta.url)]
  for (const base of bases) {
    const wasmUrl = new URL('gopdfsuit.wasm', base)
    const execUrl = new URL('wasm_exec.js', base)
    try {
      const [wasmRes, execRes] = await Promise.all([fetch(wasmUrl), fetch(execUrl)])
      if (!wasmRes.ok || !execRes.ok) continue
      return {
        wasmBytes: new Uint8Array(await wasmRes.arrayBuffer()),
        execSource: await execRes.text(),
        execLabel: execUrl.href,
      }
    } catch {
      continue
    }
  }
  throw missingWasm('failed to fetch gopdfsuit.wasm / wasm_exec.js')
}

async function evalWasmExec(source, filename) {
  if (typeof globalThis.Go === 'function') return
  if (isNode) {
    const vm = await import('node:vm')
    vm.runInThisContext(source, { filename })
    return
  }
  const run = new Function(source)
  run()
}

const EXPECTED_BINDS = [
  'goGeneratePDF',
  'goMergePDFs',
  'goSplitPDF',
  'goFillPDF',
  'goCompressPDF',
  'goRedactGetPageInfo',
  'goRedactExtractText',
  'goRedactFindText',
  'goRedactApply',
  'goRedactAdvanced',
]

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
  const missing = EXPECTED_BINDS.filter((name) => typeof globalThis[name] !== 'function')
  if (missing.length > 0) {
    throw missingWasm(`gopdfsuit.wasm did not register: ${missing.join(', ')}`)
  }
}
