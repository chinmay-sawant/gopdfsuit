import { useEffect, useRef } from 'react'
import { ExternalLink, FileText, Moon, Sun } from 'lucide-react'
import { NavLink, useLocation } from 'react-router-dom'
import { allTools } from '../../content/tools'
import { useTheme } from '../../theme'

const navClass = ({ isActive }) => `site-nav-link${isActive ? ' active' : ''}`

export default function SiteHeader() {
  const { theme, toggle } = useTheme()
  const { pathname } = useLocation()
  const toolMenuRef = useRef(null)

  const closeToolMenu = () => {
    if (toolMenuRef.current) toolMenuRef.current.open = false
  }

  useEffect(() => {
    closeToolMenu()
  }, [pathname])

  return (
    <>
      <a className="skip-link" href="#main-content">Skip to content</a>
      <header className="site-header">
        <div className="site-header-inner">
          <NavLink className="site-brand" onClick={closeToolMenu} to="/" end>
            <FileText aria-hidden="true" size={22} strokeWidth={1.7} />
            <span>GoPdfSuit</span>
          </NavLink>
          <nav className="site-nav" aria-label="Primary navigation">
            <NavLink className={navClass} onClick={closeToolMenu} to="/" end>Home</NavLink>
            <details className="tool-menu" ref={toolMenuRef}>
              <summary>Tools</summary>
              <div className="tool-menu-panel">
                {allTools.map(({ to, label, icon: Icon }) => (
                  <NavLink className="tool-menu-link" key={`${to}-${label}`} onClick={closeToolMenu} to={to}>
                    <Icon aria-hidden="true" size={16} strokeWidth={1.8} />
                    {label}
                  </NavLink>
                ))}
              </div>
            </details>
            <NavLink className={navClass} onClick={closeToolMenu} to="/editor">Editor</NavLink>
            <NavLink className={navClass} onClick={closeToolMenu} to="/comparison">Proof</NavLink>
            <NavLink className={navClass} onClick={closeToolMenu} to="/screenshots">Screens</NavLink>
          </nav>
          <div className="site-header-actions">
            <button aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'} className="icon-button" onClick={toggle} type="button">
              {theme === 'dark' ? <Sun aria-hidden="true" size={18} /> : <Moon aria-hidden="true" size={18} />}
            </button>
            <a className="site-repository-link" href="https://github.com/chinmay-sawant/gopdfsuit" rel="noreferrer" target="_blank">
              Repository <ExternalLink aria-hidden="true" size={14} />
            </a>
          </div>
        </div>
      </header>
    </>
  )
}
