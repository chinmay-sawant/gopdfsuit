import { ExternalLink, FileText } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="site-footer-inner">
        <div>
          <Link className="site-brand" to="/">
            <FileText aria-hidden="true" size={20} strokeWidth={1.7} />
            <span>GoPdfSuit</span>
          </Link>
          <p>Templates, PDF work, and HTML conversion in one Go toolkit.</p>
        </div>
        <div className="site-footer-links" aria-label="Footer navigation">
          <Link to="/editor">Editor</Link>
          <Link to="/viewer">Template preview</Link>
          <Link to="/comparison">Product proof</Link>
          <a href="https://github.com/chinmay-sawant/gopdfsuit" rel="noreferrer" target="_blank">
            Repository <ExternalLink aria-hidden="true" size={13} />
          </a>
        </div>
      </div>
    </footer>
  )
}
