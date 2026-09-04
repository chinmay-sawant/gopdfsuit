// Liberation font delivery for in-browser compliant generation.
// The engine cannot download fonts in WASM (pdfa.go EnsureFontsAvailable
// rejects); instead JS fetches TTF bytes from /fonts/ (vendored, OFL) and
// registers each under its STANDARD PDF name, mirroring the server flow in
// RegisterLiberationFontsForPDFA (pdfa.go:421-450).

import { ensureGopdfsuitWasm, asUint8Array, cachedFetch, callWasmObject } from './core.js'
import { PDFA_FONT_MANIFEST } from './manifests.generated.js'

// Standard name -> file. Generated at build time by
// scripts/check-wasm-manifests.mjs from internal/pdf/font/pdfa.go
// (LiberationFontMapping + LiberationFontFiles); re-exported here for
// backwards compatibility. Do not hand-edit: the prebuild check fails the
// build on drift.
export { PDFA_FONT_MANIFEST }

const FONT_BASE_URL = `${import.meta.env.BASE_URL}fonts`

const CACHE_NAME = 'gopdfsuit-fonts-v1'

async function fetchFontBytes(file) {
  const url = `${FONT_BASE_URL}/${file}`
  try {
    return await cachedFetch(url, { cacheName: CACHE_NAME, as: 'bytes' })
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

/**
 * Register a single user-supplied TTF/OTF under a display name (font upload
 * path). Works offline; the face joins the same registry generate reads.
 */
export async function registerFontLocal(name, bytes) {
  await ensureGopdfsuitWasm()
  return callWasmObject('goRegisterFont', [name, asUint8Array(bytes)])
}
