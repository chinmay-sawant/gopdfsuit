import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.jsx'
import './index.css'
import { ThemeProvider } from './theme.jsx'
import { AuthProviderWithGoogle } from './contexts/AuthContext'
import { apiConfig } from './utils/apiConfig'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ThemeProvider>
      <AuthProviderWithGoogle clientId={apiConfig.googleClientId}>
        <App />
      </AuthProviderWithGoogle>
    </ThemeProvider>
  </React.StrictMode>,
)