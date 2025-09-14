import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { FileText, Edit, Merge, FileCheck, Globe, Image, Menu, X } from 'lucide-react'

const Navbar = () => {
  const [isOpen, setIsOpen] = useState(false)
  const location = useLocation()

  const navItems = [
    { path: '/', label: 'Home', icon: FileText },
    { path: '/viewer', label: 'Viewer', icon: FileText },
    { path: '/editor', label: 'Editor', icon: Edit },
    { path: '/merge', label: 'Merge', icon: Merge },
    { path: '/filler', label: 'Filler', icon: FileCheck },
    { path: '/htmltopdf', label: 'HTML→PDF', icon: Globe },
    { path: '/htmltoimage', label: 'HTML→Image', icon: Image },
  ]

  return (
    <nav style={{
      background: 'rgba(255, 255, 255, 0.1)',
      backdropFilter: 'blur(10px)',
      borderBottom: '1px solid rgba(255, 255, 255, 0.2)',
      padding: '1rem 0',
      position: 'sticky',
      top: 0,
      zIndex: 1000,
    }}>
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
              color: 'white',
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
                  color: location.pathname === path ? '#4ecdc4' : 'white',
                  textDecoration: 'none',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  transition: 'all 0.3s ease',
                  background: location.pathname === path ? 'rgba(78, 205, 196, 0.2)' : 'transparent',
                }}
                onMouseEnter={(e) => {
                  if (location.pathname !== path) {
                    e.target.style.background = 'rgba(255, 255, 255, 0.1)'
                  }
                }}
                onMouseLeave={(e) => {
                  if (location.pathname !== path) {
                    e.target.style.background = 'transparent'
                  }
                }}
              >
                <Icon size={16} />
                {label}
              </Link>
            ))}
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsOpen(!isOpen)}
            style={{
              display: 'none',
              background: 'none',
              border: 'none',
              color: 'white',
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
            background: 'rgba(255, 255, 255, 0.1)',
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
                  color: location.pathname === path ? '#4ecdc4' : 'white',
                  textDecoration: 'none',
                  padding: '0.75rem',
                  borderRadius: '6px',
                  background: location.pathname === path ? 'rgba(78, 205, 196, 0.2)' : 'transparent',
                }}
              >
                <Icon size={16} />
                {label}
              </Link>
            ))}
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