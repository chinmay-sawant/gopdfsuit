import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { FileText, Edit, Merge, FileCheck, Globe, Image, Menu, X, Sun, Moon, Camera } from 'lucide-react'
import { useTheme } from '../theme'

const Navbar = () => {
  const [isOpen, setIsOpen] = useState(false)
  const { theme, toggle } = useTheme()
  const location = useLocation()

  const navItems = [
    { path: '/', label: 'Home', icon: FileText },
    { path: '/viewer', label: 'Viewer', icon: FileText },
    { path: '/editor', label: 'Editor', icon: Edit },
    { path: '/merge', label: 'Merge', icon: Merge },
    { path: '/filler', label: 'Filler', icon: FileCheck },
    { path: '/htmltopdf', label: 'HTML→PDF', icon: Globe },
    { path: '/htmltoimage', label: 'HTML→Image', icon: Image },
    { path: '/comparison', label: 'Comparison', icon: FileCheck },
    { path: '/screenshots', label: 'Screenshots', icon: Camera },
  ]

  return (
    <nav className="navbar" style={{ padding: '0.75rem 0' }}>
      <div className="container">
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <Link 
            to="/" 
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              color: 'hsl(var(--foreground))',
              textDecoration: 'none',
              fontSize: '1.5rem',
              fontWeight: '700',
            }}
          >
            📄 GoPdfSuit
          </Link>

          {/* Desktop Menu */}
          <div style={{
            display: 'flex',
            gap: '1rem',
            alignItems: 'center',
            '@media (max-width: 768px)': {
              display: 'none',
            },
          }} className="desktop-menu">
            {navItems.slice(1).map(({ path, label, icon: Icon }) => (
              <Link
                key={path}
                to={path}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  color: location.pathname === path ? 'var(--secondary-color)' : 'hsl(var(--muted-foreground))',
                  textDecoration: 'none',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  transition: 'all 0.3s ease',
                  background: location.pathname === path ? 'color-mix(in hsl, var(--secondary-color) 15%, transparent)' : 'transparent',
                }}
                onMouseEnter={(e) => {
                  if (location.pathname !== path) {
                    e.currentTarget.style.background = 'hsl(var(--accent))'
                    e.currentTarget.style.color = 'hsl(var(--accent-foreground))'
                  }
                }}
                onMouseLeave={(e) => {
                  if (location.pathname !== path) {
                    e.currentTarget.style.background = 'transparent'
                    e.currentTarget.style.color = 'hsl(var(--muted-foreground))'
                  }
                }}
              >
                <Icon size={16} />
                {label}
              </Link>
            ))}

            {/* Theme toggle - desktop */}
            <button
              onClick={toggle}
              title={theme === 'dark' ? 'Switch to light' : 'Switch to dark'}
              style={{
                background: 'transparent',
                border: '1px solid hsl(var(--border))',
                padding: '0.4rem 0.6rem',
                borderRadius: '8px',
                cursor: 'pointer',
                color: 'hsl(var(--foreground))',
                transition: 'all 0.3s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'hsl(var(--accent))'
                e.currentTarget.style.color = 'hsl(var(--accent-foreground))'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent'
                e.currentTarget.style.color = 'hsl(var(--foreground))'
              }}
            >
              {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
            </button>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsOpen(!isOpen)}
            style={{
              display: 'none',
              background: 'none',
              border: 'none',
              color: 'hsl(var(--foreground))',
              cursor: 'pointer',
              padding: '0.5rem',
            }}
            className="mobile-menu-button"
          >
            {isOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>

        {/* Mobile Menu */}
        {isOpen && (
          <div style={{
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
            marginTop: '1rem',
            padding: '1rem',
            background: 'hsl(var(--card))',
            borderRadius: '8px',
          }} className="mobile-menu">
            {navItems.slice(1).map(({ path, label, icon: Icon }) => (
              <Link
                key={path}
                to={path}
                onClick={() => setIsOpen(false)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  color: location.pathname === path ? 'var(--secondary-color)' : 'hsl(var(--muted-foreground))',
                  textDecoration: 'none',
                  padding: '0.75rem',
                  borderRadius: '6px',
                  background: location.pathname === path ? 'color-mix(in hsl, var(--secondary-color) 15%, transparent)' : 'transparent',
                }}
              >
                <Icon size={16} />
                {label}
              </Link>
            ))}

            {/* Theme toggle - mobile */}
            <button
              onClick={toggle}
              title={theme === 'dark' ? 'Switch to light' : 'Switch to dark'}
              style={{
                background: 'transparent',
                border: '1px solid hsl(var(--border))',
                padding: '0.6rem',
                borderRadius: '8px',
                cursor: 'pointer',
                color: 'hsl(var(--foreground))',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.5rem',
                transition: 'all 0.3s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'hsl(var(--accent))'
                e.currentTarget.style.color = 'hsl(var(--accent-foreground))'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent'
                e.currentTarget.style.color = 'hsl(var(--foreground))'
              }}
            >
              {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
              <span className="text-muted">{theme === 'dark' ? 'Light mode' : 'Dark mode'}</span>
            </button>
          </div>
        )}
      </div>

      <style jsx>{`
        @media (max-width: 768px) {
          .desktop-menu {
            display: none !important;
          }
          .mobile-menu-button {
            display: block !important;
          }
        }
        @media (min-width: 769px) {
          .mobile-menu {
            display: none !important;
          }
        }
      `}</style>
    </nav>
  )
}

export default Navbar