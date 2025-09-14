import React, { createContext, useContext, useEffect, useMemo, useState } from 'react'

const ThemeContext = createContext({
  theme: 'light',
  setTheme: () => {},
  toggle: () => {},
})

export function ThemeProvider({ children }) {
  const [theme, setThemeState] = useState('light')

  // Initialize from localStorage or system preference
  useEffect(() => {
    try {
      const saved = localStorage.getItem('theme')
      if (saved === 'light' || saved === 'dark') {
        setThemeState(saved)
        return
      }
    } catch {}
    const preferDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
    setThemeState(preferDark ? 'dark' : 'light')
  }, [])

  // Apply to document
  useEffect(() => {
    const root = document.documentElement
    if (theme === 'dark') root.classList.add('dark')
    else root.classList.remove('dark')
    try { localStorage.setItem('theme', theme) } catch {}
  }, [theme])

  const api = useMemo(() => ({
    theme,
    setTheme: (t) => setThemeState(t === 'dark' ? 'dark' : 'light'),
    toggle: () => setThemeState((t) => (t === 'dark' ? 'light' : 'dark')),
  }), [theme])

  return (
    <ThemeContext.Provider value={api}>{children}</ThemeContext.Provider>
  )
}

export function useTheme() {
  return useContext(ThemeContext)
}
