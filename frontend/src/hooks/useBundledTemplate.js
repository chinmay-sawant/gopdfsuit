import { useCallback } from 'react'
import { loadBundledTemplate } from '../utils/wasm/templates.js'

const GITHUB_RAW = 'https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/master/sampledata'

/**
 * useBundledTemplate shares the offline-first template load used by Viewer,
 * Editor, and the editor Toolbar: bundled /templates/ copies first (Cache
 * API, no server), unknown names fall through to /api/v1/template-data via
 * the shared hook (auth retry + error mapping included).
 */
export function useBundledTemplate({ runJson, getAuthHeaders, onError } = {}) {
  const loadTemplateData = useCallback(async (filename, { source = 'local' } = {}) => {
    const name = String(filename || '').trim()
    if (!name) throw new Error('Please enter a template filename')
    if (source === 'github') {
      const response = await fetch(`${GITHUB_RAW}/${name}`)
      if (!response.ok) {
        throw new Error(`Failed to load from GitHub: ${response.status} ${response.statusText}`)
      }
      return response.json()
    }
    try {
      return await loadBundledTemplate(name)
    } catch (bundledError) {
      if (!bundledError || !bundledError.fallbackAvailable) throw bundledError
      const data = await runJson({
        endpoint: `/api/v1/template-data?file=${encodeURIComponent(name)}`,
        method: 'GET',
        headers: { Accept: 'application/json' },
        getAuthHeaders,
        onError,
      })
      if (!data) throw new Error('Template not found on server')
      return data
    }
  }, [runJson, getAuthHeaders, onError])

  return { loadTemplateData }
}

export default useBundledTemplate
