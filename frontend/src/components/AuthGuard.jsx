import { GoogleLogin } from '@react-oauth/google'
import { useAuth } from '../contexts/AuthContext'
import { apiConfig, isGoogleOAuthEnabled } from '../utils/apiConfig'

export default function AuthGuard({ children }) {
  const { isAuthenticated, login } = useAuth()

  if (!isAuthenticated) {
    const googleReady = isGoogleOAuthEnabled(apiConfig.googleClientId)

    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        gap: '2rem',
        padding: '2rem',
        background: 'hsl(var(--background))',
        color: 'hsl(var(--foreground))',
      }}>
        <div style={{
          textAlign: 'center',
          maxWidth: '500px',
        }}>
          <h1 style={{ fontSize: '2rem', marginBottom: '1rem' }}>
            PDF Template Editor
          </h1>
          <p style={{
            fontSize: '1rem',
            color: 'hsl(var(--muted-foreground))',
            marginBottom: '2rem',
          }}>
            Please sign in with your Google account to access the editor.
          </p>

          {googleReady ? (
            <GoogleLoginPanel onLogin={login} />
          ) : (
            <div style={{
              padding: '1.5rem 2rem',
              background: 'hsl(var(--card))',
              borderRadius: '12px',
              border: '1px solid hsl(var(--border))',
              color: 'hsl(var(--muted-foreground))',
              fontSize: '0.9rem',
              lineHeight: 1.5,
            }}>
              <p style={{ margin: 0 }}>
                Google sign-in is not configured. Set <code>VITE_GOOGLE_CLIENT_ID</code> in
                the project <code>.env</code> and rebuild the frontend.
              </p>
            </div>
          )}

          <p style={{
            fontSize: '0.875rem',
            color: 'hsl(var(--muted-foreground))',
            marginTop: '1.5rem',
          }}>
            By signing in, you agree to our terms of service and privacy policy.
          </p>
        </div>
      </div>
    )
  }
  return children
}

function GoogleLoginPanel({ onLogin }) {
  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      padding: '2rem',
      background: 'hsl(var(--card))',
      borderRadius: '12px',
      border: '1px solid hsl(var(--border))',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
    }}>
      <GoogleLogin
        onSuccess={onLogin}
        onError={() => {
          console.error('Login Failed')
          alert('Login failed. Please try again.')
        }}
        useOneTap
        theme="outline"
        size="large"
        text="signin_with"
        shape="rectangular"
      />
    </div>
  )
}
