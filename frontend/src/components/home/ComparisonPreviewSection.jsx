import { Link } from 'react-router-dom'
import {
  CheckCircle,
  Zap,
  Download,
  ArrowRight,
} from 'lucide-react'

const stats = [
  { value: 'Free', label: 'vs $2K-4K/dev/year', color: '#4ecdc4', icon: <CheckCircle size={28} /> },
  { value: '< 100ms', label: 'Response time', color: '#007acc', icon: <Zap size={28} /> },
  { value: '0 deps', label: 'Pure Go binary', color: '#f093fb', icon: <Download size={28} /> },
]

const featureCategories = [
  {
    label: 'Language Support',
    features: [
      { name: 'Go Support (gopdflib)', desc: 'Direct Struct Access & HTTP API', color: '#ffc107' },
      { name: 'Native Python Bindings', desc: 'CGO + ctypes wrapper via pypdfsuit', color: '#4ecdc4' },
      { name: 'Python Web Client', desc: 'Lightweight REST API client', color: '#007acc' },
    ],
  },
  {
    label: 'Compliance & Security',
    features: [
      { name: 'PDF/A-4 Compliance', desc: 'Archival standard with sRGB ICC profiles', color: '#4ecdc4' },
      { name: 'PDF/UA-2 Accessibility', desc: 'Universal accessibility compliance', color: '#007acc' },
      { name: 'AES-128 Encryption', desc: 'Password protection with permissions', color: '#f093fb' },
      { name: 'Digital Signatures', desc: 'PKCS#7 certificates with visual appearance', color: '#ffc107' },
    ],
  },
  {
    label: 'Document Features',
    features: [
      { name: 'Font Subsetting', desc: 'TrueType embedding with glyph optimization', color: '#4ecdc4' },
      { name: 'PDF Merge', desc: 'Combine multiple PDFs, preserve forms', color: '#007acc' },
      { name: 'XFDF Form Filling', desc: 'Advanced field detection and population', color: '#f093fb' },
      { name: 'Bookmarks & Links', desc: 'Outlines with internal/external hyperlinks', color: '#ffc107' },
      { name: 'Language Agnostic', desc: 'REST API works with any programming language', color: '#f093fb' },
    ],
  },
]

const ComparisonPreviewSection = ({ isVisible }) => {
  const visible = isVisible['section-comparison-preview']

  return (
    <section id="section-comparison-preview" style={{ padding: '5rem 0' }}>
      <div className="container">
        <div
          className={`text-center animate-fadeInUp stagger-animation ${visible ? 'visible' : ''}`}
          style={{ marginBottom: '3rem' }}
        >
          <h2 className="gradient-text section-heading">
            Why Choose GoPdfSuit?
          </h2>
          <p className="section-subheading" style={{ maxWidth: '700px' }}>
            Enterprise features at zero cost — compare with iTextPDF, PDFLib, and commercial solutions
          </p>
        </div>

        {/* Quick Stats */}
        <div
          className="grid grid-3"
          style={{ marginBottom: '2.5rem', maxWidth: '900px', margin: '0 auto 2.5rem' }}
        >
          {stats.map((stat, index) => (
            <div
              key={index}
              className={`glass-card animate-fadeInScale stagger-animation ${visible ? 'visible' : ''}`}
              style={{
                textAlign: 'center',
                padding: '1.5rem',
                animationDelay: `${0.2 + index * 0.1}s`,
              }}
            >
              <div style={{ color: stat.color, marginBottom: '0.75rem', display: 'flex', justifyContent: 'center' }}>
                {stat.icon}
              </div>
              <div className="stat-value" style={{ color: stat.color, marginBottom: '0.25rem', fontSize: '1.8rem' }}>
                {stat.value}
              </div>
              <div style={{ fontSize: '0.85rem', color: 'hsl(var(--muted-foreground))' }}>
                {stat.label}
              </div>
            </div>
          ))}
        </div>

        {/* Feature Comparison - Categorized */}
        <div
          className={`glass-card animate-fadeInScale stagger-animation ${visible ? 'visible' : ''}`}
          style={{ width: '100%', padding: '2.5rem' }}
        >
          <h3 style={{
            color: 'hsl(var(--foreground))',
            marginBottom: '2rem',
            fontSize: '1.3rem',
            textAlign: 'center',
          }}>
            Built-in Enterprise Features
          </h3>

          <div className="comparison-categories">
            {featureCategories.map((category, catIdx) => (
              <div key={catIdx} className="comparison-category">
                <div className="comparison-category-label">{category.label}</div>
                <div className="comparison-category-grid">
                  {category.features.map((feature, index) => (
                    <div key={index} className="comparison-feature-item">
                      <CheckCircle size={18} style={{ color: feature.color, flexShrink: 0, marginTop: '2px' }} />
                      <div>
                        <div className="comparison-feature-name">{feature.name}</div>
                        <div className="comparison-feature-desc">{feature.desc}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>

          <div style={{ textAlign: 'center' }}>
            <Link
              to="/comparison"
              className="btn-glow"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.75rem',
                textDecoration: 'none',
              }}
            >
              View Full Comparison
              <ArrowRight size={18} />
            </Link>
          </div>
        </div>
      </div>
    </section>
  )
}

export default ComparisonPreviewSection
