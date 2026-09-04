// Liberation font delivery for in-browser compliant generation.
// The engine cannot download fonts in WASM (pdfa.go EnsureFontsAvailable
// rejects); instead JS fetches TTF bytes from /fonts/ (vendored, OFL) and
// registers each under its STANDARD PDF name, mirroring the server flow in
// RegisterLiberationFontsForPDFA (pdfa.go:421-450).

import { ensureGopdfsuitWasm } from './core.js'

const FONT_BASE_URL = `${import.meta.env.BASE_URL}fonts`

// Standard name -> file, mirroring LiberationFontMapping keys plus
// LiberationFontFiles values (internal/pdf/font/pdfa.go:31-67).
export const PDFA_FONT_MANIFEST = [
  { name: 'Helvetica', file: 'LiberationSans-Regular.ttf' },
  { name: 'Helvetica-Bold', file: 'LiberationSans-Bold.ttf' },
  { name: 'Helvetica-Oblique', file: 'LiberationSans-Italic.ttf' },
  { name: 'Helvetica-BoldOblique', file: 'LiberationSans-BoldItalic.ttf' },
  { name: 'Times-Roman', file: 'LiberationSerif-Regular.ttf' },
  { name: 'Times-Bold', file: 'LiberationSerif-Bold.ttf' },
  { name: 'Times-Italic', file: 'LiberationSerif-Italic.ttf' },
  { name: 'Times-BoldItalic', file: 'LiberationSerif-BoldItalic.ttf' },
  { name: 'Courier', file: 'LiberationMono-Regular.ttf' },
  { name: 'Courier-Bold', file: 'LiberationMono-Bold.ttf' },
  { name: 'Courier-Oblique', file: 'LiberationMono-Italic.ttf' },
  { name: 'Courier-BoldOblique', file: 'LiberationMono-BoldItalic.ttf' },
]

const CACHE_NAME = 'gopdfsuit-fonts-v1'

async function fetchFontBytes(file) {
  const url = `${FONT_BASE_URL}/${file}`
  try {
    const cache = await caches.open(CACHE_NAME)
    const hit = await cache.match(url)
    if (hit) return new Uint8Array(await hit.arrayBuffer())
    const response = await fetch(url)
    if (!response.ok) throw new Error(`font fetch failed: ${file} (${response.status})`)
    const bytes = new Uint8Array(await response.arrayBuffer())
    cache.put(url, new Response(bytes, { headers: { 'Content-Type': 'font/ttf' } })).catch(() => {})
    return bytes
  } catch (err) {
    if (err instanceof TypeError) {
      // caches/fetch unavailable (or offline with empty cache): direct fetch.
      const response = await fetch(url)
      if (!response.ok) throw new Error(`font fetch failed: ${file} (${response.status})`)
      return new Uint8Array(await response.arrayBuffer())
    }
    throw err
  }
}

function callWasmObject(fnName, args) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') {
    const err = new Error(`${fnName} is not in the shipped WASM bundle yet`)
    err.fallbackAvailable = true
    err.missingEngine = true
    throw err
  }
  const out = fn(...args)
  if (out && typeof out === 'object' && !(out instanceof Uint8Array) && out.error) {
    throw new Error(out.error)
  }
  return out
}

/**
 * Ensure the 12 Liberation faces are registered for compliant generation.
 * Fetches each TTF once (Cache API) and registers it under its standard
 * name. Returns { registered, missing, fetched }.
 */
export async function ensurePDFAFonts() {
  await ensureGopdfsuitWasm()
  const status = callWasmObject('goEnsurePDFAFonts', [])
  const missing = new Set(status.missing || [])
  let fetched = 0
  for (const entry of PDFA_FONT_MANIFEST) {
    if (!missing.has(entry.name)) continue
    const bytes = await fetchFontBytes(entry.file)
    callWasmObject('goRegisterFont', [entry.name, bytes])
    fetched += 1
  }
  const after = callWasmObject('goEnsurePDFAFonts', [])
  return { registered: after.registered || [], missing: after.missing || [], fetched }
}
