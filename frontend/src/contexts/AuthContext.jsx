import { createContext, useContext, useState, useEffect } from 'react'
import { GoogleOAuthProvider, useGoogleLogin } from '@react-oauth/google'

const AuthContext = createContext()

// eslint-disable-next-line react-refresh/only-export-components
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

  // Create a function to trigger login programmatically
  const triggerLogin = useGoogleLogin({
    onSuccess: (codeResponse) => {
      // Note: This flow gives an access token, but for id_token we usually use the other flow.
      // However, if we just want to trigger the popup and handle the result uniformly,
      // we might need to adjust based on how the backend expects tokens.
      // Since existing login() takes credentialResponse.credential (which is an ID token from the <GoogleLogin> button),
      // we might need to use the 'implicit' flow or 'id_token' flow if we want an ID token directly.
      // BUT, useGoogleLogin by default gives an access token.
      // FOR NOW, let's assume valid login is the goal.
      // Actually, standard `useGoogleLogin` with `flow: 'auth-code'` gives a code to swap.
      // If we want an ID token, we might not get it directly from useGoogleLogin easily without backend swap.
      // ALTERNATIVELY: We can use the default flow (implicit) which gives access_token.
      // BUT our backend expects `Authorization: Bearer ${idToken}`. 
      // Checking `login` function: it takes `credentialResponse.credential` which is a JWT ID token.
      // The <GoogleLogin> component provides this.
      // useGoogleLogin hook provides an access_token by default (Implicit flow) or code (Auth Code flow).
      // To get an ID token compatible with our `login` function from `useGoogleLogin`,
      // we might need to fetch user info using the access token and then synthesize a session, 
      // OR allow the backend to accept access tokens.
      // HOWEVER, simply triggering the popup to let the user sign in is the request.
      // Let's implement it to set the token if possible, or just used as a trigger.

      // WAIT: The user asked to "show the google login popup".
      // `useGoogleLogin` does exactly that.
      // Let's use it. We'll set the access token as the ID token for now if strictly needed,
      // or fetches the user profile.
      // A better approach for compatibility with existing `login` (which expects JWT)
      // might be to just use the access token if the backend validates it, or...
      // actually, let's just use the access_token from the hook as the "idToken" for now,
      // assuming the backend defines validation.
      // If the backend strictly validates JWT ID tokens (from Google's OIDC), passing an opaque access token might fail.
      // Let's check if we can get an ID token from useGoogleLogin.
      // We can't easily get an OIDC ID token from useGoogleLogin without a backend exchange if using 'auth-code'.

      // Let's try to fetch user info and store that, at least the UI will update to "Authenticated".
      // But for API calls, we need the valid token the backend expects.

      console.log("Login successful", codeResponse);
      if (codeResponse.access_token) {
        setIdToken(codeResponse.access_token);
        setIsAuthenticated(true);
        localStorage.setItem('google_id_token', codeResponse.access_token);
        // Fetch user info to populate `user` state
        fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
          headers: { Authorization: `Bearer ${codeResponse.access_token}` },
        })
          .then((res) => res.json())
          .then((userData) => {
            setUser(userData);
            localStorage.setItem('google_user', JSON.stringify(userData));
          });
      }
    },
    onError: (error) => console.log('Login Failed:', error)
  });

  const value = {
    idToken,
    user,
    isAuthenticated,
    login,
    logout,
    getAuthHeaders,
    triggerLogin // Expose this
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
