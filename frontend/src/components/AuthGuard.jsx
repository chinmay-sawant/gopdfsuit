import { GoogleLogin } from '@react-oauth/google'
import { useAuth } from '../contexts/AuthContext'

export default function AuthGuard({ children }) {
  const { isAuthenticated, login } = useAuth()

  if (!isAuthenticated) {
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
        color: 'hsl(var(--foreground))'
      }}>
        <div style={{
          textAlign: 'center',
          maxWidth: '500px'
        }}>
          <h1 style={{ fontSize: '2rem', marginBottom: '1rem' }}>
            PDF Template Editor
          </h1>
          <p style={{ 
            fontSize: '1rem', 
            color: 'hsl(var(--muted-foreground))',
            marginBottom: '2rem' 
          }}>
            Please sign in with your Google account to access the editor.
          </p>
          
          <div style={{
            display: 'flex',
            justifyContent: 'center',
            padding: '2rem',
            background: 'hsl(var(--card))',
            borderRadius: '12px',
            border: '1px solid hsl(var(--border))',
            boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
          }}>
            <GoogleLogin
              onSuccess={login}
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

          <p style={{ 
            fontSize: '0.875rem', 
            color: 'hsl(var(--muted-foreground))',
            marginTop: '1.5rem' 
          }}>
            By signing in, you agree to our terms of service and privacy policy.
          </p>
        </div>
      </div>
    )
  }
  return children
}