import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { apiConfig } from '../utils/apiConfig'

const AuthContext = createContext()

const TOKEN_KEY = 'auth_token'
const USER_KEY = 'auth_user'

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

function useAuthState() {
  const [token, setToken] = useState(null)
  const [user, setUser] = useState(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    const storedToken = localStorage.getItem(TOKEN_KEY)
    const storedUser = localStorage.getItem(USER_KEY)
    if (storedToken && storedUser) {
      setToken(storedToken)
      try {
        setUser(JSON.parse(storedUser))
      } catch {
        setUser(null)
      }
      setIsAuthenticated(true)
    }
  }, [])

  const persist = useCallback((newToken, newUser) => {
    setToken(newToken)
    setUser(newUser)
    setIsAuthenticated(true)
    localStorage.setItem(TOKEN_KEY, newToken)
    localStorage.setItem(USER_KEY, JSON.stringify(newUser))
  }, [])

  const logout = useCallback(() => {
    setToken(null)
    setUser(null)
    setIsAuthenticated(false)
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }, [])

  // submit posts credentials to auth-ms and stores the returned JWT.
  const submit = useCallback(async (path, email, password) => {
    const res = await fetch(`${apiConfig.authUrl}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    let data = {}
    try {
      data = await res.json()
    } catch {
      // non-JSON response
    }
    if (!res.ok) {
      throw new Error(data.error || 'Authentication failed')
    }
    persist(data.token, data.user)
    return data
  }, [persist])

  const login = useCallback((email, password) => submit('/auth/login', email, password), [submit])
  const register = useCallback((email, password) => submit('/auth/register', email, password), [submit])

  const getAuthHeaders = useCallback(() => {
    if (!token) {
      throw new Error('Not authenticated')
    }
    return { Authorization: `Bearer ${token}` }
  }, [token])

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    logout,
    // Form-based auth: on a 401 the caller clears the stale token so the login
    // view re-appears (the editor route is wrapped by AuthGuard).
    triggerLogin: logout,
    getAuthHeaders,
  }
}

export function AppAuthProvider({ children }) {
  const auth = useAuthState()
  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>
}
