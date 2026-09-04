// Bundled sample templates for offline-first template-data.
// Server endpoint (/api/v1/template-data) serves the full sampledata tree;
// these static copies cover the samples the UI offers, fetched
// BASE_URL-relative (self-host safe) with Cache API, so the Viewer/Editor
// work offline once downloaded. Unknown names throw with fallbackAvailable
// so pages can offer the server endpoint instead.
//
// The list is generated at build time by scripts/check-wasm-manifests.mjs
// from public/templates/; re-exported here for backwards compatibility.
// Do not hand-edit: the prebuild check fails the build on drift.
import { cachedFetch } from './core.js'
import { BUNDLED_TEMPLATES } from './manifests.generated.js'

export { BUNDLED_TEMPLATES }

const TEMPLATE_BASE_URL = `${import.meta.env.BASE_URL}templates`

const TEMPLATE_CACHE = 'gopdfsuit-templates-v1'

async function fetchStatic(name) {
  const url = `${TEMPLATE_BASE_URL}/${name}`
  try {
    const data = await cachedFetch(url, { cacheName: TEMPLATE_CACHE, as: 'json' })
    if (!data) return null
    return data
  } catch {
    return null
  }
}

/**
 * Load a bundled sample template by basename. Returns parsed JSON, or throws
 * with fallbackAvailable=true when the name is not bundled (caller offers
 * the server endpoint).
 */
export async function loadBundledTemplate(name) {
  const base = String(name || '').split('/').pop()
  if (!BUNDLED_TEMPLATES.includes(base)) {
    const err = new Error(`template not bundled: ${base} (bundled: ${BUNDLED_TEMPLATES.join(', ')})`)
    err.fallbackAvailable = true
    throw err
  }
  const data = await fetchStatic(base)
  if (!data) {
    const err = new Error(`bundled template unreachable offline: ${base}`)
    err.fallbackAvailable = true
    throw err
  }
  return data
}
