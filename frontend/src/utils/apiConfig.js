// Environment configuration for API endpoints
// Checks if the backend is deployed on Google Cloud Run

export const OFFLINE_DEMO_MESSAGE =
  'Online PDF generation is disabled on this demo site. Run the application locally to generate PDFs.'

/**
 * True when the UI is served from GitHub Pages (static demo; API calls are blocked by CORS).
 */
export const isGitHubPagesHost = () => {
  if (typeof window === 'undefined') return false
  return window.location.hostname.endsWith('github.io')
}

const isNetworkFetchError = (error) => {
  if (!error) return false
  const message = String(error.message || error).toLowerCase()
  return (
    (error.name === 'TypeError' && message.includes('failed to fetch')) ||
    message.includes('networkerror') ||
    message.includes('load failed') ||
    message.includes('net::err_failed') ||
    message.includes('cors')
  )
}

/**
 * Turn opaque fetch/CORS failures into a clear local-run instruction for demo hosts.
 */
export const formatApiError = (error) => {
  if (isGitHubPagesHost()) {
    return new Error(OFFLINE_DEMO_MESSAGE)
  }
  if (isNetworkFetchError(error)) {
    const baseUrl = getApiBaseUrl()
    if (baseUrl && !/localhost|127\.0\.0\.1/.test(baseUrl)) {
      return new Error(OFFLINE_DEMO_MESSAGE)
    }
  }
  return error instanceof Error ? error : new Error(String(error))
}

/**
 * Check if the application is running on Google Cloud Run
 * Cloud Run sets K_SERVICE environment variable automatically
 */
export const isCloudRunEnvironment = () => {
  // In a frontend app, we need to check via backend or config
  // Since we can't access process.env directly in browser, we'll use import.meta.env
  return import.meta.env.VITE_IS_CLOUD_RUN === 'true' || 
         import.meta.env.VITE_ENVIRONMENT === 'cloudrun'
}

/**
 * Get the API base URL based on environment
 */
export const getApiBaseUrl = () => {
  // If running on Cloud Run, use the Cloud Run URL
  if (isCloudRunEnvironment()) {
    return import.meta.env.VITE_CLOUD_RUN_URL || import.meta.env.VITE_API_URL || ''
  }
  
  // Otherwise use local/development URL
  return import.meta.env.VITE_API_URL || 'http://localhost:8080'
}

/**
 * Check if authentication is required
 * Only require auth when running on Cloud Run
 */
export const isAuthRequired = () => {
  return isCloudRunEnvironment()
}

/**
 * Make an authenticated API request
 */
export const makeAuthenticatedRequest = async (url, options = {}, getAuthHeaders) => {
  const { throwOnError = true, ...fetchOptions } = options

  if (isGitHubPagesHost()) {
    throw new Error(OFFLINE_DEMO_MESSAGE)
  }

  // If auth is required, ensure we have auth headers
  if (isAuthRequired()) {
    if (!getAuthHeaders) {
      throw new Error('Authentication required but no auth headers provided')
    }
    
    const authHeaders = getAuthHeaders()
    fetchOptions.headers = {
      ...fetchOptions.headers,
      ...authHeaders
    }
  }
  
  const baseUrl = getApiBaseUrl()
  const fullUrl = url.startsWith('http') ? url : `${baseUrl}${url}`

  let response
  try {
    response = await fetch(fullUrl, fetchOptions)
  } catch (error) {
    throw formatApiError(error)
  }

  if (!response.ok && throwOnError) {
    const errorText = await response.text()
    console.log('Error response body:', errorText)
    if (response.status === 401 || response.status === 403) {
      throw new Error('Authentication failed. Please login again.')
    }

    let serverMessage = ''
    if (errorText) {
      try {
        const parsed = JSON.parse(errorText)
        serverMessage = parsed?.error || parsed?.message || ''
      } catch {
        serverMessage = errorText
      }
    }

    if (serverMessage) {
      throw new Error(serverMessage)
    }
    throw new Error(`API request failed: ${response.statusText}`)
  }
  
  return response
}

const PLACEHOLDER_GOOGLE_CLIENT_ID = 'your-google-oauth-client-id.apps.googleusercontent.com'

/**
 * Check if Google OAuth is configured with a real client ID.
 * Local dev works without it; Cloud Run deployments should set VITE_GOOGLE_CLIENT_ID.
 */
export const isGoogleOAuthConfigured = () => {
  const clientId = (import.meta.env.VITE_GOOGLE_CLIENT_ID || '').trim()
  return Boolean(clientId && clientId !== PLACEHOLDER_GOOGLE_CLIENT_ID)
}

/**
 * Configuration object for easy access
 */
export const apiConfig = {
  isCloudRun: isCloudRunEnvironment(),
  baseUrl: getApiBaseUrl(),
  authRequired: isAuthRequired(),
  googleClientId: import.meta.env.VITE_GOOGLE_CLIENT_ID || '',
  googleOAuthEnabled: isGoogleOAuthConfigured(),
}
