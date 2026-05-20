import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { GoogleOAuthProvider, useGoogleLogin } from '@react-oauth/google'
import { isGoogleOAuthEnabled } from '../utils/apiConfig'

const AuthContext = createContext()

const noop = () => {}

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

function useAuthState() {
  const [idToken, setIdToken] = useState(null)
  const [user, setUser] = useState(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    const storedToken = localStorage.getItem('google_id_token')
    const storedUser = localStorage.getItem('google_user')

    if (storedToken && storedUser) {
      setIdToken(storedToken)
      setUser(JSON.parse(storedUser))
      setIsAuthenticated(true)
    }
  }, [])

  const login = useCallback((credentialResponse) => {
    const token = credentialResponse.credential

    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      const userData = {
        email: payload.email,
        name: payload.name,
        picture: payload.picture,
      }

      setIdToken(token)
      setUser(userData)
      setIsAuthenticated(true)
      localStorage.setItem('google_id_token', token)
      localStorage.setItem('google_user', JSON.stringify(userData))
    } catch (err) {
      console.error('Failed to decode token:', err)
    }
  }, [])

  const logout = useCallback(() => {
    setIdToken(null)
    setUser(null)
    setIsAuthenticated(false)
    localStorage.removeItem('google_id_token')
    localStorage.removeItem('google_user')
  }, [])

  const getAuthHeaders = useCallback(() => {
    if (!idToken) {
      throw new Error('Not authenticated')
    }
    return {
      Authorization: `Bearer ${idToken}`,
    }
  }, [idToken])

  return {
    idToken,
    setIdToken,
    user,
    setUser,
    isAuthenticated,
    setIsAuthenticated,
    login,
    logout,
    getAuthHeaders,
  }
}

function AuthProviderLocal({ children }) {
  const auth = useAuthState()

  const value = {
    ...auth,
    triggerLogin: noop,
    googleEnabled: false,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

function AuthProviderGoogleInner({ children }) {
  const auth = useAuthState()
  const { setIdToken, setUser, setIsAuthenticated } = auth

  const triggerLogin = useGoogleLogin({
    onSuccess: (codeResponse) => {
      if (codeResponse.access_token) {
        setIdToken(codeResponse.access_token)
        setIsAuthenticated(true)
        localStorage.setItem('google_id_token', codeResponse.access_token)
        fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
          headers: { Authorization: `Bearer ${codeResponse.access_token}` },
        })
          .then((res) => res.json())
          .then((userData) => {
            setUser(userData)
            localStorage.setItem('google_user', JSON.stringify(userData))
          })
      }
    },
    onError: (error) => console.log('Login Failed:', error),
  })

  const value = {
    ...auth,
    triggerLogin,
    googleEnabled: true,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

function AuthProviderGoogle({ children, clientId }) {
  return (
    <GoogleOAuthProvider clientId={clientId}>
      <AuthProviderGoogleInner>{children}</AuthProviderGoogleInner>
    </GoogleOAuthProvider>
  )
}

/** Mounts Google OAuth only when auth is required and VITE_GOOGLE_CLIENT_ID is set. */
export function AppAuthProvider({ children, clientId = '' }) {
  const id = (clientId || '').trim()
  if (isGoogleOAuthEnabled(id)) {
    return <AuthProviderGoogle clientId={id}>{children}</AuthProviderGoogle>
  }
  return <AuthProviderLocal>{children}</AuthProviderLocal>
}

// Backward-compatible export
export const AuthProviderWithGoogle = AppAuthProvider
