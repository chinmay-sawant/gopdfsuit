import { useCallback, useEffect, useState } from 'react'
import { DEFAULT_FONTS } from '../components/editor/constants'

// Module-level registry shared across Editor mounts (replaces the old
// Editor.jsx _fontsCache/_fontsFetchPromise pair). The entry holds the last
// good list so remounts and the Viewer/Editor pair never refetch; refreshFonts
// invalidates after an upload.
let fontsCache = null
let fontsFetchPromise = null

const isFontList = (data) => data && Array.isArray(data.fonts)

/**
 * useFonts owns the /api/v1/fonts list with a shared module cache.
 * Offline-first: failures keep DEFAULT_FONTS; user uploads register locally
 * via goRegisterFont at the call site, then call refreshFonts() to repull.
 */
export function useFonts({ runJson, getAuthHeaders, onError } = {}) {
  const [fonts, setFonts] = useState(() => fontsCache || DEFAULT_FONTS)

  useEffect(() => {
    let cancelled = false
    const loadFonts = async () => {
      if (fontsCache) {
        setFonts(fontsCache)
        return
      }
      if (!fontsFetchPromise) {
        fontsFetchPromise = runJson({
          endpoint: '/api/v1/fonts',
          method: 'GET',
          getAuthHeaders,
          onError: (message) => {
            if (onError) onError(message)
            else console.warn('Failed to fetch fonts, using defaults:', message)
          },
        }).then((data) => {
          if (isFontList(data)) {
            fontsCache = data.fonts
            return data.fonts
          }
          console.warn('Failed to fetch fonts, using defaults')
          return null
        }).catch((error) => {
          console.error('Error fetching fonts:', error)
          fontsFetchPromise = null
          return null
        })
      }
      const list = await fontsFetchPromise
      if (!cancelled && list) setFonts(list)
    }
    loadFonts()
    return () => { cancelled = true }
  }, [runJson, getAuthHeaders, onError])

  const refreshFonts = useCallback(async () => {
    fontsCache = null
    fontsFetchPromise = null
    const data = await runJson({
      endpoint: '/api/v1/fonts',
      method: 'GET',
      getAuthHeaders,
      onError,
    })
    if (isFontList(data)) {
      fontsCache = data.fonts
      setFonts(data.fonts)
      return data.fonts
    }
    return null
  }, [runJson, getAuthHeaders, onError])

  return { fonts, setFonts, refreshFonts }
}

export default useFonts
