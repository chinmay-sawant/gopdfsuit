import { Link } from 'react-router-dom'
import {
  FileText,
  CheckCircle,
  Star,
  Github,
  ChevronDown,
  ArrowRight,
  Sparkles
} from 'lucide-react'

const pillRows = [
  {
    label: 'Compliance',
    items: ['PDF/A-4 & PDF/UA-2 Compliant', 'AES-128 Encryption']
  },
  {
    label: 'Core',
    items: ['Multi-page Support', 'Split PDFs', 'PDF Redaction']
  },
  {
    label: 'Conversion',
    items: ['HTML To PDF', 'HTML To Image', 'Typst Math Rendering']
  },
  {
    label: 'Platform',
    items: ['Native Python Support', 'Send Data via API', 'Docker Support', 'Private & In-Memory']
  }
]

const HeroSection = ({ starCount }) => {
  return (
    <section
      id="section-hero"
      className="hero-section"
      style={{ padding: '6rem 0 4rem', textAlign: 'center' }}
    >
      <div className="container">
        {/* Sparkle badge */}
        <div className="hero-badge animate-fadeInUp">
          <Sparkles size={16} />
          Open Source PDF Generation Engine
        </div>

        {/* Main Title */}
        <h1
          className="hero-title gradient-text animate-fadeInUp"
          style={{ animationDelay: '0.1s' }}
        >
          GoPdfSuit
        </h1>

        {/* Subtitle */}
        <div
          className="hero-subtitle animate-fadeInUp"
          style={{
            marginBottom: '3rem',
            animationDelay: '0.2s',
            maxWidth: '800px',
            marginLeft: 'auto',
            marginRight: 'auto',
          }}
        >
          <p className="hero-description">
            An high-performance, <span className="highlight-foreground">MIT-licensed</span> Go engine that <span className="highlight-teal">saves enterprise costs</span> and solves critical <span className="highlight-foreground">compliance challenges</span> for Fintechs & Enterprises by generating secure, <span className="highlight-blue">PDF/UA-2 & PDF/A-4</span> compliant documents in <span className="highlight-yellow">under 10ms*</span>.
          </p>

          <div className="hero-pills-container">
            {pillRows.map((row, rowIdx) => (
              <div key={rowIdx} className="hero-pills-row">
                <span className="row-label">{row.label}</span>
                {row.items.map((feature, i) => (
                  <span key={i} className="hero-pill">
                    <CheckCircle size={14} />
                    {feature}
                  </span>
                ))}
              </div>
            ))}
          </div>
        </div>

        {/* CTA Buttons */}
        <div className="hero-cta-group animate-fadeInUp" style={{ animationDelay: '0.3s' }}>
          <Link
            to="/editor"
            className="btn-glow glow-on-hover"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.75rem',
              textDecoration: 'none',
              fontSize: '1.15rem',
              padding: '1.1rem 2.8rem',
            }}
          >
            <FileText size={22} />
            Try PDF Generator
            <ArrowRight size={20} />
          </Link>
          <a
            href="https://github.com/chinmay-sawant/gopdfsuit"
            target="_blank"
            rel="noopener noreferrer"
            className="btn-outline-glow"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.75rem',
              textDecoration: 'none',
            }}
          >
            <Github size={20} />
            View on GitHub
            <div className="star-badge">
              <Star size={14} fill={starCount ? "currentColor" : "none"} />
              <span style={{ fontSize: '0.9rem' }}>{starCount !== null ? starCount : 'Star'}</span>
            </div>
          </a>
        </div>
      </div>

      {/* Scroll indicator */}
      <div
        className="scroll-indicator"
        onClick={() => document.getElementById('section-features')?.scrollIntoView({ behavior: 'smooth' })}
      >
        <ChevronDown size={32} color="#4ecdc4" />
      </div>
    </section>
  )
}

export default HeroSection
