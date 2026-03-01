import { Link } from 'react-router-dom'
import {
  Star,
  Github,
  FileText,
  Edit,
  Merge,
  Scissors,
  FileCheck,
  Globe,
  Image,
  Camera,
  Book,
} from 'lucide-react'

const toolLinks = [
  { path: '/editor', label: 'Visual Editor', icon: <Edit size={14} /> },
  { path: '/viewer', label: 'PDF Viewer', icon: <FileText size={14} /> },
  { path: '/merge', label: 'Merge PDFs', icon: <Merge size={14} /> },
  { path: '/split', label: 'Split PDFs', icon: <Scissors size={14} /> },
  { path: '/filler', label: 'Form Filler', icon: <FileCheck size={14} /> },
]

const resourceLinks = [
  { path: '/documentation', label: 'Documentation', icon: <Book size={14} /> },
  { path: '/comparison', label: 'Comparison', icon: <FileCheck size={14} /> },
  { path: '/screenshots', label: 'Screenshots', icon: <Camera size={14} /> },
  { path: '/htmltopdf', label: 'HTML → PDF', icon: <Globe size={14} /> },
  { path: '/htmltoimage', label: 'HTML → Image', icon: <Image size={14} /> },
]

const FooterSection = ({ isVisible, starCount }) => {
  const visible = isVisible['section-footer']

  return (
    <footer
      id="section-footer"
      style={{
        padding: '4rem 0 2rem',
        marginTop: '2rem',
        background: 'linear-gradient(0deg, rgba(78,205,196,0.03) 0%, transparent 100%)',
      }}
    >
      <div className="container">
        <div className="section-divider" style={{ margin: '0 0 3rem' }} />

        <div className={`animate-fadeInUp stagger-animation ${visible ? 'visible' : ''}`}>
          <div className="footer-grid">
            {/* Brand column */}
            <div>
              <h4 style={{ fontSize: '1.2rem', textTransform: 'none', letterSpacing: 'normal' }}>
                📄 GoPdfSuit
              </h4>
              <p className="footer-brand-desc">
                An open-source, high-performance PDF generation engine built in Go. MIT licensed, enterprise-ready.
              </p>
              <a
                href="https://github.com/chinmay-sawant/gopdfsuit"
                target="_blank"
                rel="noopener noreferrer"
                className="btn-outline-glow"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  textDecoration: 'none',
                  padding: '0.5rem 1rem',
                  fontSize: '0.85rem',
                }}
              >
                <Github size={16} />
                GitHub
                {starCount !== null && (
                  <span className="star-badge" style={{ fontSize: '0.8rem' }}>
                    <Star size={12} fill="currentColor" />
                    {starCount}
                  </span>
                )}
              </a>
            </div>

            {/* Tools column */}
            <div>
              <h4>Tools</h4>
              <div className="footer-links">
                {toolLinks.map(({ path, label, icon }) => (
                  <Link key={path} to={path} className="footer-link">
                    {icon} {label}
                  </Link>
                ))}
              </div>
            </div>

            {/* Resources column */}
            <div>
              <h4>Resources</h4>
              <div className="footer-links">
                {resourceLinks.map(({ path, label, icon }) => (
                  <Link key={path} to={path} className="footer-link">
                    {icon} {label}
                  </Link>
                ))}
              </div>
            </div>

            {/* Community column */}
            <div>
              <h4>Community</h4>
              <div className="footer-links">
                <a
                  href="https://github.com/chinmay-sawant/gopdfsuit"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="footer-link"
                >
                  <Github size={14} /> GitHub Repository
                </a>
                <a
                  href="https://github.com/chinmay-sawant"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="footer-link"
                >
                  <Star size={14} /> Follow the Author
                </a>
              </div>
              <p style={{
                color: 'hsl(var(--muted-foreground))',
                fontSize: '0.9rem',
                marginTop: '1.5rem',
                lineHeight: '1.6',
              }}>
                Made with ❤️ and ☕ by{' '}
                <a
                  href="https://github.com/chinmay-sawant"
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{ color: '#4ecdc4', textDecoration: 'none', fontWeight: '600' }}
                >
                  Chinmay Sawant
                </a>
              </p>
            </div>
          </div>

          {/* Bottom bar */}
          <div className="footer-bottom">
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '0.85rem',
              marginBottom: 0,
              opacity: 0.7,
            }}>
              <Star size={14} style={{ display: 'inline', marginRight: '0.5rem', color: '#ffc107' }} />
              Star this repo if you find it helpful!
            </p>
          </div>
        </div>
      </div>
    </footer>
  )
}

export default FooterSection
