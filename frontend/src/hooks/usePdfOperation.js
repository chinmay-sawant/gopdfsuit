import { useCallback, useEffect, useRef, useState } from 'react'
import { formatApiError, makeAuthenticatedRequest } from '../utils/apiConfig'

/**
 * Fallback text per HTTP status when the server sent no usable message.
 * Auth (401/403) never reaches here: it routes to onAuthRequired.
 */
const fallbackMessageForStatus = (status) => {
  switch (status) {
    case 413:
      return 'Upload exceeds the server size limit.'
    case 422:
    case 400:
      return 'Request rejected as invalid.'
    case 500:
    case 502:
    case 503:
      return 'Server failed to process the request.'
    default:
      return null
  }
}

export const downloadBlobUrl = (url, filename) => {
  if (!url || !filename) return
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const toBlobUrl = (blob, mimeType) => {
  const typed = blob instanceof Blob ? blob : new Blob([blob], { type: mimeType })
  return URL.createObjectURL(typed)
}

/**
 * usePdfOperation owns the shared PDF-operation flow: POST (FormData or JSON),
 * blob-URL lifecycle (revoke on reset/unmount), error mapping via
 * formatApiError, and auth retry via onAuthRequired.
 *
 * Each page keeps its op-specific form (endpoint + body builder) and calls
 * run()/runJson() with the request it built.
 */
export const usePdfOperation = ({ onAuthRequired, onError } = {}) => {
  const [isLoading, setIsLoading] = useState(false)
  const [resultUrl, setResultUrl] = useState('')
  const [error, setError] = useState(null)
  const urlRef = useRef('')

  const revokeResult = useCallback(() => {
    if (urlRef.current) {
      URL.revokeObjectURL(urlRef.current)
      urlRef.current = ''
    }
  }, [])

  useEffect(() => () => revokeResult(), [revokeResult])

  const reset = useCallback(() => {
    revokeResult()
    setResultUrl('')
    setError(null)
  }, [revokeResult])

  const handleFailure = useCallback((err, onErrorOverride) => {
    const formatted = formatApiError(err)
    const status = err?.status
    if (status === 401 || status === 403) {
      if (onAuthRequired) onAuthRequired()
      return null
    }
    const message = formatted.message || fallbackMessageForStatus(status) || 'Request failed.'
    setError(message)
    const notify = onErrorOverride || onError
    if (notify) notify(message)
    else alert(message)
    return null
  }, [onAuthRequired, onError])

  const request = useCallback(async ({ endpoint, method = 'POST', body, headers, getAuthHeaders, throwOnError = true }) => {
    const response = await makeAuthenticatedRequest(
      endpoint,
      { method, body, headers, throwOnError },
      getAuthHeaders,
    )
    return response
  }, [])

  const runJson = useCallback(async (requestArgs) => {
    const { onError: onErrorOverride, ...args } = requestArgs || {}
    setIsLoading(true)
    setError(null)
    try {
      const response = await request(args)
      if (!response.ok) {
        const text = await response.text()
        throw new Error(text || `API request failed: ${response.statusText}`)
      }
      return await response.json()
    } catch (err) {
      handleFailure(err, onErrorOverride)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [request, handleFailure])

  const storeBlob = useCallback((blob, { filename, autoDownload = true, mimeType = 'application/pdf', onBlob } = {}) => {
    if (!blob || blob.size === 0) throw new Error('Received empty document')
    revokeResult()
    const url = toBlobUrl(blob, mimeType)
    urlRef.current = url
    setResultUrl(url)
    if (onBlob) onBlob(blob, url)
    if (autoDownload && filename) downloadBlobUrl(url, filename)
    return url
  }, [revokeResult])

  const run = useCallback(async ({
    endpoint,
    method = 'POST',
    body,
    headers,
    getAuthHeaders,
    filename,
    autoDownload = true,
    mimeType = 'application/pdf',
    onBlob,
    onError: onErrorOverride,
  }) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await request({ endpoint, method, body, headers, getAuthHeaders })
      const blob = await response.blob()
      return storeBlob(blob, { filename, autoDownload, mimeType, onBlob })
    } catch (err) {
      handleFailure(err, onErrorOverride)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [request, storeBlob, handleFailure])

  const runLocal = useCallback(async (task, options = {}) => {
    // Browser-local path shared by every op: task() must not upload
    // anything. Single output (Uint8Array/Blob) stores one preview URL;
    // array output (Split multi-file) stores the first part for preview and
    // downloads the rest, so one local call yields N files with no upload.
    // Server fallback (if any) lives behind an explicit consent click, which
    // runSmart owns via consentOffer below.
    const { filenames, filename, autoDownload = true, mimeType = 'application/pdf', onBlob, onError: onErrorOverride } = options
    const names = Array.isArray(filenames) ? filenames : (filename ? [filename] : [])
    setIsLoading(true)
    setError(null)
    try {
      const output = await task()
      const parts = Array.isArray(output) ? output : [output]
      if (parts.length === 0) throw new Error('Received empty document')
      const urls = parts.map((part, index) => {
        const blob = part instanceof Blob ? part : new Blob([part], { type: mimeType })
        if (blob.size === 0) throw new Error('Received empty document')
        const url = toBlobUrl(blob, mimeType)
        const name = names[index]
        if (autoDownload && name) downloadBlobUrl(url, name)
        return { blob, url }
      })
      revokeResult()
      urlRef.current = urls[0].url
      setResultUrl(urls[0].url)
      if (onBlob) {
        if (parts.length === 1) onBlob(urls[0].blob, urls[0].url)
        else onBlob(urls.map((entry) => entry.blob), urls.map((entry) => entry.url))
      }
      // Revoke non-preview URLs after download; keep urls[0] alive for preview.
      urls.slice(1).forEach((entry) => URL.revokeObjectURL(entry.url))
      return parts.length === 1 ? urls[0].url : urls.map((entry) => entry.url)
    } catch (err) {
      handleFailure(err, onErrorOverride)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [revokeResult, handleFailure])

  // runLocalMulti was the Split-only twin of runLocal; it now aliases the
  // unified runner above (arrays take the multi path). New code calls
  // runLocal for both shapes.
  const runLocalMulti = runLocal

  // Hook-owned consent state (3.3): pages run browser-local work through
  // runSmart(opSmart(...), { serverTask }) and render <ConsentBanner
  // offer={consentOffer} onConsent={confirmConsentUpload}
  // onDismiss={dismissConsent} />. The WASM error never alerts; the banner
  // offers the upload as an explicit click instead.
  const [consentOffer, setConsentOffer] = useState(null)

  const dismissConsent = useCallback(() => setConsentOffer(null), [])

  const runSmart = useCallback(async (smartTask, options = {}) => {
    // smartTask: (transport) => Promise<bytes|bytes[]> with transport shaped
    // as { getAuthHeaders, allowServerFallback: false }; the hook supplies
    // consent by running serverTask only from confirmConsentUpload.
    // serverTask: () => Promise<bytes|bytes[]> (explicit-consent upload).
    const { serverTask, serverLabel, getAuthHeaders, onError: onErrorOverride, ...storeOptions } = options
    setConsentOffer(null)
    setIsLoading(true)
    setError(null)
    let wasmMessage = ''
    try {
      const output = await smartTask({ getAuthHeaders, allowServerFallback: false })
      const parts = Array.isArray(output) ? output : [output]
      if (parts.length === 0) throw new Error('Received empty document')
      const mimeType = storeOptions.mimeType || 'application/pdf'
      const names = Array.isArray(storeOptions.filenames)
        ? storeOptions.filenames
        : (storeOptions.filename ? [storeOptions.filename] : [])
      const urls = parts.map((part, index) => {
        const blob = part instanceof Blob ? part : new Blob([part], { type: mimeType })
        if (blob.size === 0) throw new Error('Received empty document')
        const url = toBlobUrl(blob, mimeType)
        const name = names[index]
        if (storeOptions.autoDownload && name) downloadBlobUrl(url, name)
        return { blob, url }
      })
      revokeResult()
      urlRef.current = urls[0].url
      setResultUrl(urls[0].url)
      if (storeOptions.onBlob) {
        if (parts.length === 1) storeOptions.onBlob(urls[0].blob, urls[0].url)
        else storeOptions.onBlob(urls.map((entry) => entry.blob), urls.map((entry) => entry.url))
      }
      urls.slice(1).forEach((entry) => URL.revokeObjectURL(entry.url))
      return parts.length === 1 ? urls[0].url : urls.map((entry) => entry.url)
    } catch (err) {
      wasmMessage = err?.message || 'Request failed.'
      if (err && err.fallbackAvailable && getAuthHeaders && serverTask) {
        setConsentOffer({ message: wasmMessage, label: serverLabel, serverTask, storeOptions })
        return null
      }
      const formatted = formatApiError(err)
      const status = err?.status
      if (status === 401 || status === 403) {
        if (onAuthRequired) onAuthRequired()
        return null
      }
      const message = formatted.message || fallbackMessageForStatus(status) || wasmMessage
      setError(message)
      if (onErrorOverride) onErrorOverride(message)
      else if (onError) onError(message)
      else alert(message)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [revokeResult, onAuthRequired, onError])

  const confirmConsentUpload = useCallback(async () => {
    if (!consentOffer) return null
    const { serverTask, storeOptions } = consentOffer
    setConsentOffer(null)
    return runLocal(serverTask, storeOptions)
  }, [consentOffer, runLocal])

  const download = useCallback((filename) => {
    downloadBlobUrl(resultUrl, filename)
  }, [resultUrl])

  return { isLoading, resultUrl, error, setError, setResultUrl, run, runJson, runLocal, runLocalMulti, runSmart, consentOffer, dismissConsent, confirmConsentUpload, request, reset, revokeResult, download }
}

export default usePdfOperation
