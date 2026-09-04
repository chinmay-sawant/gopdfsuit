import { useCallback, useEffect, useRef, useState } from 'react'
import { formatApiError, makeAuthenticatedRequest } from '../utils/apiConfig'

export const isAuthErrorMessage = (message) => {
  const text = String(message || '')
  return (
    text.includes('Authentication failed') ||
    text.includes('Not authenticated') ||
    text.includes('login again') ||
    text.includes('401') ||
    text.includes('403')
  )
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
    if (isAuthErrorMessage(formatted.message)) {
      if (onAuthRequired) onAuthRequired()
      return null
    }
    setError(formatted.message)
    const notify = onErrorOverride || onError
    if (notify) notify(formatted.message)
    else alert(formatted.message)
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
    setIsLoading(true)
    setError(null)
    try {
      const output = await task()
      const blob = output instanceof Blob ? output : new Blob([output], { type: options.mimeType || 'application/pdf' })
      return storeBlob(blob, options)
    } catch (err) {
      handleFailure(err, options.onError)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [storeBlob, handleFailure])

  const download = useCallback((filename) => {
    downloadBlobUrl(resultUrl, filename)
  }, [resultUrl])

  return { isLoading, resultUrl, error, setError, setResultUrl, run, runJson, runLocal, request, reset, revokeResult, download }
}

export default usePdfOperation
