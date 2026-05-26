import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.jsx'
import './index.css'
import { ThemeProvider } from './theme.jsx'
import { AppAuthProvider } from './contexts/AuthContext'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ThemeProvider>
      <AppAuthProvider>
        <App />
      </AppAuthProvider>
    </ThemeProvider>
  </React.StrictMode>,
)