import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { GoogleOAuthProvider, useGoogleLogin } from '@react-oauth/google'
import { isGoogleOAuthConfigured } from '../utils/apiConfig'

const AuthContext = createContext()

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

const noopTriggerLogin = () => {
  console.warn(
    'Google OAuth is not configured. Copy frontend/.env.example to frontend/.env and set VITE_GOOGLE_CLIENT_ID.'
  )
}

const useAuthState = () => {
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
        picture: payload.picture
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

  const loginWithAccessToken = useCallback((accessToken) => {
    if (!accessToken) return

    setIdToken(accessToken)
    setIsAuthenticated(true)
    localStorage.setItem('google_id_token', accessToken)

    fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
      .then((res) => res.json())
      .then((userData) => {
        setUser(userData)
        localStorage.setItem('google_user', JSON.stringify(userData))
      })
      .catch((err) => console.error('Failed to fetch user info:', err))
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
      Authorization: `Bearer ${idToken}`
    }
  }, [idToken])

  return {
    idToken,
    user,
    isAuthenticated,
    login,
    logout,
    getAuthHeaders,
    loginWithAccessToken,
  }
}

const AuthProviderBase = ({ children, triggerLogin = noopTriggerLogin }) => {
  const authState = useAuthState()

  const value = {
    ...authState,
    triggerLogin,
    googleOAuthEnabled: false,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

const GoogleAuthProviderInner = ({ children }) => {
  const authState = useAuthState()

  const triggerLogin = useGoogleLogin({
    onSuccess: (codeResponse) => {
      console.log('Login successful', codeResponse)
      authState.loginWithAccessToken(codeResponse.access_token)
    },
    onError: (error) => console.log('Login Failed:', error),
  })

  const value = {
    ...authState,
    triggerLogin,
    googleOAuthEnabled: true,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// Wrapper component that provides Google OAuth when configured
export const AuthProviderWithGoogle = ({ children, clientId }) => {
  if (isGoogleOAuthConfigured()) {
    return (
      <GoogleOAuthProvider clientId={clientId}>
        <GoogleAuthProviderInner>{children}</GoogleAuthProviderInner>
      </GoogleOAuthProvider>
    )
  }

  return <AuthProviderBase>{children}</AuthProviderBase>
}