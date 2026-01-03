import { GoogleLogin } from '@react-oauth/google'
import { useAuth } from '../contexts/AuthContext'
import { LogOut, User } from 'lucide-react'

export default function AuthGuard({ children }) {
  const { isAuthenticated, login, logout, user } = useAuth()

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

  return (
    <>
      <div style={{
        position: 'sticky',
        top: 0,
        zIndex: 50,
        background: 'hsl(var(--card))',
        borderBottom: '1px solid hsl(var(--border))',
        padding: '0.75rem 1.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          {user?.picture && (
            <img 
              src={user.picture} 
              alt={user.name}
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '50%',
                border: '2px solid hsl(var(--border))'
              }}
            />
          )}
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ 
              fontSize: '0.875rem', 
              fontWeight: '600',
              color: 'hsl(var(--foreground))' 
            }}>
              {user?.name || 'User'}
            </span>
            <span style={{ 
              fontSize: '0.75rem', 
              color: 'hsl(var(--muted-foreground))' 
            }}>
              {user?.email || ''}
            </span>
          </div>
        </div>

        <button
          onClick={logout}
          className="btn"
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.5rem 1rem',
            fontSize: '0.875rem',
            background: 'hsl(var(--secondary))',
            color: 'hsl(var(--secondary-foreground))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '6px',
            cursor: 'pointer',
            transition: 'all 0.2s ease'
          }}
        >
          <LogOut size={16} />
          Sign Out
        </button>
      </div>
      {children}
    </>
  )
}
