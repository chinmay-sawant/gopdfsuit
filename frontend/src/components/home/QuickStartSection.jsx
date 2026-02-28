import { Link } from 'react-router-dom'
import { Edit } from 'lucide-react'

const QuickStartSection = ({ isVisible }) => {
  const visible = isVisible['section-quickstart']

  return (
    <section id="section-quickstart" style={{ padding: '5rem 0' }}>
      <div className="container">
        <div className="split-layout">
          {/* Left side - Text content */}
          <div className={`animate-slideInLeft stagger-animation ${visible ? 'visible' : ''}`}>
            <h2 className="gradient-text section-heading" style={{ marginBottom: '1.5rem' }}>
              Get Started in Seconds
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
              marginBottom: '2rem',
              lineHeight: '1.7',
            }}>
              Clone the repository and start generating PDFs immediately.
              No complex setup required.
            </p>

            <ul className="check-list">
              <li>Zero external dependencies</li>
              <li>Single binary deployment</li>
              <li>Docker ready out of the box</li>
              <li>Cross-platform support</li>
            </ul>

            <Link
              to="/editor"
              className="btn-glow"
              style={{
                marginTop: '2rem',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.75rem',
                textDecoration: 'none',
              }}
            >
              <Edit size={20} />
              Open Editor
            </Link>
          </div>

          {/* Right side - Terminal mockup */}
          <div
            className={`terminal-window animate-slideInRight stagger-animation ${visible ? 'visible' : ''}`}
            style={{ animationDelay: '0.2s' }}
          >
            <div className="terminal-header">
              <span className="terminal-dot red"></span>
              <span className="terminal-dot yellow"></span>
              <span className="terminal-dot green"></span>
              <span style={{ color: '#888', marginLeft: '1rem', fontSize: '0.85rem' }}>Terminal</span>
            </div>
            <div className="terminal-body">
              <div style={{ marginBottom: '0.5rem' }}>
                <span className="terminal-prompt">$ </span>
                <span className="terminal-command">git clone https://github.com/chinmay-sawant/gopdfsuit.git</span>
              </div>
              <div style={{ marginBottom: '0.5rem' }}>
                <span className="terminal-prompt">$ </span>
                <span className="terminal-command">cd gopdfsuit</span>
              </div>
              <div style={{ marginBottom: '0.5rem' }}>
                <span className="terminal-prompt">$ </span>
                <span className="terminal-command">make run</span>
              </div>
              <div style={{ marginTop: '1rem' }}>
                <span className="terminal-success">✓ Server listening on http://localhost:8080</span>
              </div>
              <div style={{ marginTop: '0.5rem' }}>
                <span className="terminal-success">✓ Ready for PDF generation!</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

export default QuickStartSection
