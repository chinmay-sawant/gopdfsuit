// Bundled sample templates for offline-first template-data.
// Server endpoint (/api/v1/template-data) serves the full sampledata tree;
// these static copies cover the samples the UI offers, fetched
// BASE_URL-relative (self-host safe) with Cache API, so the Viewer/Editor
// work offline once downloaded. Unknown names throw with fallbackAvailable
// so pages can offer the server endpoint instead.

const TEMPLATE_BASE_URL = `${import.meta.env.BASE_URL}templates`

export const BUNDLED_TEMPLATES = [
  'resume1.json',
  'resume2.json',
  'financial_report.json',
]

const TEMPLATE_CACHE = 'gopdfsuit-templates-v1'

async function fetchStatic(name) {
  const url = `${TEMPLATE_BASE_URL}/${name}`
  try {
    const cache = await caches.open(TEMPLATE_CACHE)
    const hit = await cache.match(url)
    if (hit) return hit.json()
    const response = await fetch(url)
    if (!response.ok) return null
    const data = await response.json()
    cache.put(url, new Response(JSON.stringify(data), { headers: { 'Content-Type': 'application/json' } })).catch(() => {})
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
