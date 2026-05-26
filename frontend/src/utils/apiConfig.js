// Environment configuration for API endpoints
// Checks if the backend is deployed on Cloud Run

/**
 * Check if the application is running on Cloud Run
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
 * Base URL of the auth-ms microservice (login/register/verify).
 */
export const getAuthBaseUrl = () => {
  return import.meta.env.VITE_AUTH_URL || 'http://localhost:9090'
}

/**
 * Check if authentication is required.
 * Enabled on Cloud Run, or anywhere via VITE_AUTH_ENABLED=true (e.g. `make run`).
 */
export const isAuthRequired = () => {
  return isCloudRunEnvironment() || import.meta.env.VITE_AUTH_ENABLED === 'true'
}

/**
 * Make an authenticated API request
 */
export const makeAuthenticatedRequest = async (url, options = {}, getAuthHeaders) => {
  const { throwOnError = true, ...fetchOptions } = options

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
  const response = await fetch(fullUrl, fetchOptions)
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

/**
 * Configuration object for easy access
 */
export const apiConfig = {
  isCloudRun: isCloudRunEnvironment(),
  baseUrl: getApiBaseUrl(),
  authUrl: getAuthBaseUrl(),
  authRequired: isAuthRequired(),
}
