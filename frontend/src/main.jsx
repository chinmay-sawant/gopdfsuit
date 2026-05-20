import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.jsx'
import './index.css'
import { ThemeProvider } from './theme.jsx'
import { AppAuthProvider } from './contexts/AuthContext'
import { apiConfig } from './utils/apiConfig'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ThemeProvider>
      <AppAuthProvider clientId={apiConfig.googleClientId}>
        <App />
      </AppAuthProvider>
    </ThemeProvider>
  </React.StrictMode>,
)