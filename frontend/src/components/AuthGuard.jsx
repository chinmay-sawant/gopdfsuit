import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'

export default function AuthGuard({ children }) {
  const { isAuthenticated } = useAuth()

  if (!isAuthenticated) {
    return <LoginView />
  }
  return children
}

function LoginView() {
  const { login, register } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const run = async (action) => {
    setError('')
    setBusy(true)
    try {
      await action(email, password)
    } catch (err) {
      setError(err.message || 'Authentication failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="container" style={{ maxWidth: '440px', marginTop: '3rem' }}>
      <div className="glass-card" style={{ padding: '2rem' }}>
        <h1 style={{ marginTop: 0, marginBottom: '0.5rem' }}>PDF Template Editor</h1>
        <p style={{ color: 'hsl(var(--muted-foreground))', marginTop: 0, marginBottom: '1.5rem' }}>
          Please sign in to access the editor.
        </p>

        <form
          data-testid="auth-form"
          onSubmit={(e) => { e.preventDefault(); run(login) }}
          style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}
        >
          <input
            data-testid="auth-email"
            type="text"
            autoComplete="username"
            placeholder="Usuario o email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            style={{ width: '100%' }}
          />
          <input
            data-testid="auth-password"
            type="password"
            autoComplete="current-password"
            placeholder="Contraseña (mín. 8 caracteres)"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            style={{ width: '100%' }}
          />

          {error && (
            <p data-testid="auth-error" role="alert" style={{ color: 'crimson', margin: 0, fontSize: '0.875rem' }}>
              {error}
            </p>
          )}

          <div style={{ display: 'flex', gap: '0.75rem', marginTop: '0.25rem' }}>
            <button data-testid="auth-login" type="submit" className="btn btn-primary" disabled={busy} style={{ flex: 1 }}>
              {busy ? 'Espere…' : 'Ingresar'}
            </button>
            <button data-testid="auth-register" type="button" className="btn" disabled={busy} onClick={() => run(register)} style={{ flex: 1 }}>
              Crear cuenta
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
