import { createContext, useContext, useState, useEffect } from 'react'
import { GoogleOAuthProvider } from '@react-oauth/google'

const AuthContext = createContext()

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

export const AuthProvider = ({ children }) => {
  const [idToken, setIdToken] = useState(null)
  const [user, setUser] = useState(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    // Check for existing token in localStorage
    const storedToken = localStorage.getItem('google_id_token')
    const storedUser = localStorage.getItem('google_user')
    
    if (storedToken && storedUser) {
      setIdToken(storedToken)
      setUser(JSON.parse(storedUser))
      setIsAuthenticated(true)
    }
  }, [])

  const login = (credentialResponse) => {
    const token = credentialResponse.credential
    
    // Decode the JWT token to get user info (optional, for display purposes)
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
      
      // Store in localStorage for persistence
      localStorage.setItem('google_id_token', token)
      localStorage.setItem('google_user', JSON.stringify(userData))
    } catch (err) {
      console.error('Failed to decode token:', err)
    }
  }

  const logout = () => {
    setIdToken(null)
    setUser(null)
    setIsAuthenticated(false)
    localStorage.removeItem('google_id_token')
    localStorage.removeItem('google_user')
  }

  const getAuthHeaders = () => {
    if (!idToken) {
      throw new Error('Not authenticated')
    }
    return {
      Authorization: `Bearer ${idToken}`
    }
  }

  const value = {
    idToken,
    user,
    isAuthenticated,
    login,
    logout,
    getAuthHeaders
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// Wrapper component that provides Google OAuth
export const AuthProviderWithGoogle = ({ children, clientId }) => {
  return (
    <GoogleOAuthProvider clientId={clientId}>
      <AuthProvider>{children}</AuthProvider>
    </GoogleOAuthProvider>
  )
}
