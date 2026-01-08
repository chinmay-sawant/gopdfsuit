import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { FileText, Edit, Merge, FileCheck, Globe, Image, Menu, X, Sun, Moon, Camera, LogOut, Scissors } from 'lucide-react'
import { useTheme } from '../theme'
import { useAuth } from '../contexts/AuthContext'
import { isAuthRequired } from '../utils/apiConfig'

const Navbar = () => {
  const [isOpen, setIsOpen] = useState(false)
  const { theme, toggle } = useTheme()
  const location = useLocation()
  const authRequired = isAuthRequired()
  const { isAuthenticated, user, logout } = authRequired ? useAuth() : { isAuthenticated: false, user: null, logout: null }

  // Check if running on GitHub Pages
  const isGitHubPages = window.location.hostname.includes('chinmay-sawant.github.io')

  const navItems = [
    { path: '/', label: 'Home', icon: FileText },
    { path: '/viewer', label: 'Viewer', icon: FileText },
    { path: '/editor', label: 'Editor', icon: Edit },
    { path: '/merge', label: 'Merge', icon: Merge },
    { path: '/split', label: 'Split', icon: Scissors },
    { path: '/filler', label: 'Filler', icon: FileCheck },
    { path: '/htmltopdf', label: 'HTML→PDF', icon: Globe },
    { path: '/htmltoimage', label: 'HTML→Image', icon: Image },
    { path: '/comparison', label: 'Comparison', icon: FileCheck },
    { path: '/screenshots', label: 'Screenshots', icon: Camera },
  ]

  return (
    <>
      {/* Preview Ribbon for GitHub Pages - Fixed position on left */}
      {isGitHubPages && (
        <div 
          className="preview-ribbon"
          title="Run the app locally to generate the PDF"
        >
          Preview
        </div>
      )}
      
      <nav className="navbar" style={{ padding: '0.75rem 0' }}>
        <div className="container-full">
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
                fontSize: '2rem',
                fontWeight: '700',
                lineHeight: '1',
                marginRight: '2rem',
              }}
            >
              <span style={{ verticalAlign: 'middle' }}>📄</span> GoPdfSuit
            </Link>
            
          {/* Desktop Menu */}
          <div style={{
            display: 'flex',
            gap: '0.5rem',
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
                  padding: '0.5rem 0.75rem',
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

            {/* User Profile and Sign Out - only show when authenticated and auth is required */}
            {authRequired && isAuthenticated && user && (
              <>
                <div 
                  title={user.email}
                  style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '0.5rem',
                    padding: '0.4rem 0.6rem',
                    borderRadius: '8px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--card))',
                    maxWidth: '200px'
                  }}>
                  {user.picture && (
                    <img 
                      src={user.picture} 
                      alt={user.name}
                      style={{
                        width: '28px',
                        height: '28px',
                        borderRadius: '50%',
                        border: '1px solid hsl(var(--border))'
                      }}
                    />
                  )}
                  <div style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
                    <span style={{ 
                      fontSize: '0.8rem', 
                      fontWeight: '600',
                      color: 'hsl(var(--foreground))',
                      lineHeight: '1.2',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}>
                      {user.name}
                    </span>
                    <span style={{ 
                      fontSize: '0.7rem', 
                      color: 'hsl(var(--muted-foreground))',
                      lineHeight: '1.2',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}>
                      {user.email}
                    </span>
                  </div>
                </div>

                <button
                  onClick={logout}
                  title="Sign Out"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    background: 'hsl(var(--destructive))',
                    color: 'hsl(var(--destructive-foreground))',
                    border: '1px solid hsl(var(--border))',
                    padding: '0.4rem 0.8rem',
                    borderRadius: '8px',
                    cursor: 'pointer',
                    fontSize: '0.875rem',
                    fontWeight: '500',
                    transition: 'all 0.3s ease',
                    whiteSpace: 'nowrap'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.opacity = '0.9'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.opacity = '1'
                  }}
                >
                  <LogOut size={16} />
                  Sign Out
                </button>
              </>
            )}
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

            {/* User Profile and Sign Out - mobile */}
            {authRequired && isAuthenticated && user && (
              <>
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: '0.75rem',
                  padding: '0.75rem',
                  borderRadius: '8px',
                  border: '1px solid hsl(var(--border))',
                  background: 'hsl(var(--card))',
                  marginTop: '0.5rem'
                }}>
                  {user.picture && (
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
                  <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
                    <span style={{ 
                      fontSize: '0.875rem', 
                      fontWeight: '600',
                      color: 'hsl(var(--foreground))'
                    }}>
                      {user.name}
                    </span>
                    <span style={{ 
                      fontSize: '0.75rem', 
                      color: 'hsl(var(--muted-foreground))' 
                    }}>
                      {user.email}
                    </span>
                  </div>
                </div>

                <button
                  onClick={() => {
                    logout()
                    setIsOpen(false)
                  }}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '0.5rem',
                    background: 'hsl(var(--destructive))',
                    color: 'hsl(var(--destructive-foreground))',
                    border: '1px solid hsl(var(--border))',
                    padding: '0.75rem',
                    borderRadius: '8px',
                    cursor: 'pointer',
                    fontSize: '0.875rem',
                    fontWeight: '500',
                    transition: 'all 0.3s ease',
                  }}
                >
                  <LogOut size={16} />
                  Sign Out
                </button>
              </>
            )}
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
        .preview-ribbon {
          position: fixed;
          left: -40px;
          top: 80px;
          transform: rotate(-45deg);
          background: hsl(var(--foreground));
          color: hsl(var(--background));
          padding: 0.5rem 3rem;
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          z-index: 1000;
          box-shadow: 0 2px 8px rgba(0,0,0,0.3);
          cursor: help;
          transition: all 0.2s ease;
        }
        .preview-ribbon:hover {
          filter: brightness(1.2);
        }
        @media (max-width: 768px) {
          .preview-ribbon {
            left: -50px;
            top: 60px;
            padding: 0.4rem 2.5rem;
            font-size: 0.65rem;
          }
        }
        @media (max-width: 480px) {
          .preview-ribbon {
            left: -55px;
            top: 50px;
            padding: 0.35rem 2rem;
            font-size: 0.6rem;
          }
        }
      `}</style>
    </nav>
    </>
  )
}

export default Navbar