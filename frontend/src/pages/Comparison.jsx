import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  ArrowRight,
  Box,
  CheckCircle,
  Code,
  DollarSign,
  FileText,
  Globe,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  X,
  Zap,
} from 'lucide-react'
import PerformanceSection from '../components/PerformanceSection'
import BackgroundAnimation from '../components/BackgroundAnimation'

const competitors = [
  {
    name: 'GoPdfSuit',
    isOurs: true,
    pricing: 'Free (MIT License)',
    performance: 'Ultra Fast (~11ms avg Zerodha)',
    deployment: 'Microservice / Sidecar / Docker',
    memory: 'In-memory processing',
    integration: 'REST API + Native Python + gopdflib',
    template: 'JSON-based templates',
    webInterface: 'Built-in viewer / editor',
    formFilling: 'XFDF advanced detection',
    pdfMerge: 'Drag and drop + form preservation',
    htmlConversion: 'gochromedp (Chromium)',
    multipage: 'Auto page breaks',
    styling: 'Font styles + borders + images',
    pdfaCompliance: 'PDF/A-4 with ICC profiles',
    pdfuaCompliance: 'PDF/UA-2 accessibility',
    encryption: 'AES-128 with permissions',
    digitalSignatures: 'PKCS#7 + visual appearance',
    fontEmbedding: 'TrueType subsetting',
    bookmarks: 'Outlines + hyperlinks',
    dockerSupport: 'Multi-stage Alpine image',
    pythonSupport: 'Native CGO + API client',
    maintenance: 'Single binary',
  },
  {
    name: 'UniPDF',
    pricing: '$3,000+/year',
    performance: 'High (Go)',
    deployment: 'Library integration',
    memory: 'Efficient',
    integration: 'Go',
    template: 'Code-based',
    webInterface: 'None',
    formFilling: 'Full support',
    pdfMerge: 'Supported',
    htmlConversion: 'Limited',
    multipage: 'Manual control',
    styling: 'Code-based',
    pdfaCompliance: 'PDF/A',
    pdfuaCompliance: 'PDF/UA',
    encryption: 'Supported',
    digitalSignatures: 'Supported',
    fontEmbedding: 'Supported',
    bookmarks: 'Supported',
    dockerSupport: 'N/A (Library)',
    pythonSupport: 'Not supported',
    maintenance: 'Commercial support',
  },
  {
    name: 'Aspose.PDF',
    pricing: '$1,199+/year',
    performance: 'High (C++)',
    deployment: 'Library integration',
    memory: 'High',
    integration: '.NET / Java / C++ / Go',
    template: 'XML / code',
    webInterface: 'Cloud only',
    formFilling: 'Full support',
    pdfMerge: 'Supported',
    htmlConversion: 'Strong support',
    multipage: 'Supported',
    styling: 'Comprehensive',
    pdfaCompliance: 'PDF/A-1 to A-3',
    pdfuaCompliance: 'PDF/UA',
    encryption: 'AES-256',
    digitalSignatures: 'Supported',
    fontEmbedding: 'Supported',
    bookmarks: 'Supported',
    dockerSupport: 'N/A (Library)',
    pythonSupport: 'Via .NET wrapper',
    maintenance: 'Commercial support',
  },
  {
    name: 'iText 7',
    pricing: '$3,500/dev/year',
    performance: 'Moderate',
    deployment: 'Library integration',
    memory: 'Mixed',
    integration: 'Java / .NET',
    template: 'Code-based',
    webInterface: 'None',
    formFilling: 'Full support',
    pdfMerge: 'Programmatic',
    htmlConversion: 'pdfHTML add-on ($)',
    multipage: 'Manual control',
    styling: 'Advanced',
    pdfaCompliance: 'PDF/A-1 to PDF/A-3',
    pdfuaCompliance: 'PDF/UA-1',
    encryption: 'AES-256',
    digitalSignatures: 'Full PKI support',
    fontEmbedding: 'Full embedding',
    bookmarks: 'Full support',
    dockerSupport: 'N/A (Library)',
    pythonSupport: 'Via wrapper',
    maintenance: 'Library updates',
  },
  {
    name: 'wkhtmltopdf',
    pricing: 'Free (LGPL)',
    performance: 'Slow (process spawn)',
    deployment: 'Binary + WebKit',
    memory: 'High (WebKit)',
    integration: 'Command line',
    template: 'HTML / CSS',
    webInterface: 'None',
    formFilling: 'Not supported',
    pdfMerge: 'Not supported',
    htmlConversion: 'Native (outdated WebKit)',
    multipage: 'CSS page breaks',
    styling: 'CSS-based',
    pdfaCompliance: 'Not supported',
    pdfuaCompliance: 'Not supported',
    encryption: 'Not supported',
    digitalSignatures: 'Not supported',
    fontEmbedding: 'Automatic',
    bookmarks: 'Limited (TOC)',
    dockerSupport: 'Manual setup',
    pythonSupport: 'Wrapper (pdfkit)',
    maintenance: 'Deprecated',
  },
]

const categories = [
  { id: 'all', label: 'All Features', icon: <Star size={16} /> },
  { id: 'pricing', label: 'Pricing', icon: <DollarSign size={16} /> },
  { id: 'performance', label: 'Performance', icon: <Zap size={16} /> },
  { id: 'integration', label: 'Integration', icon: <Code size={16} /> },
  { id: 'features', label: 'Features', icon: <FileText size={16} /> },
  { id: 'compliance', label: 'Compliance', icon: <Shield size={16} /> },
]

const features = [
  { key: 'pricing', label: 'Pricing', icon: <DollarSign size={18} />, category: 'pricing' },
  { key: 'performance', label: 'Performance', icon: <Zap size={18} />, category: 'performance' },
  { key: 'deployment', label: 'Deployment', icon: <Box size={18} />, category: 'integration' },
  { key: 'memory', label: 'Memory Usage', icon: <TrendingUp size={18} />, category: 'performance' },
  { key: 'integration', label: 'Integration', icon: <Code size={18} />, category: 'integration' },
  { key: 'template', label: 'Template Engine', icon: <FileText size={18} />, category: 'features' },
  { key: 'webInterface', label: 'Web Interface', icon: <Globe size={18} />, category: 'features' },
  { key: 'formFilling', label: 'Form Filling', icon: <CheckCircle size={18} />, category: 'features' },
  { key: 'pdfMerge', label: 'PDF Merge', icon: <CheckCircle size={18} />, category: 'features' },
  { key: 'htmlConversion', label: 'HTML to PDF', icon: <Globe size={18} />, category: 'features' },
  { key: 'multipage', label: 'Multi-page', icon: <CheckCircle size={18} />, category: 'features' },
  { key: 'styling', label: 'Styling & Images', icon: <Star size={18} />, category: 'features' },
  { key: 'pdfaCompliance', label: 'PDF/A Compliance', icon: <Shield size={18} />, category: 'compliance' },
  { key: 'pdfuaCompliance', label: 'PDF/UA Accessibility', icon: <Shield size={18} />, category: 'compliance' },
  { key: 'encryption', label: 'Encryption', icon: <Shield size={18} />, category: 'compliance' },
  { key: 'digitalSignatures', label: 'Digital Signatures', icon: <Shield size={18} />, category: 'compliance' },
  { key: 'fontEmbedding', label: 'Font Embedding', icon: <CheckCircle size={18} />, category: 'features' },
  { key: 'bookmarks', label: 'Bookmarks & Links', icon: <CheckCircle size={18} />, category: 'features' },
  { key: 'dockerSupport', label: 'Docker Support', icon: <Box size={18} />, category: 'integration' },
  { key: 'pythonSupport', label: 'Python Support', icon: <Code size={18} />, category: 'integration' },
  { key: 'maintenance', label: 'Maintenance', icon: <CheckCircle size={18} />, category: 'integration' },
]

const advantages = [
  {
    icon: <Zap size={28} />,
    title: 'Ultra Fast Performance',
    description: 'Sub-millisecond to ~7ms response times versus the slower request profiles common in commercial stacks.',
    color: 'teal',
    size: 'large',
  },
  {
    icon: <DollarSign size={28} />,
    title: 'Zero Licensing Cost',
    description: 'MIT license instead of multi-thousand-dollar annual developer licensing.',
    color: 'green',
    size: 'normal',
  },
  {
    icon: <Shield size={28} />,
    title: 'PDF/A-4 & PDF/UA-2',
    description: 'Archival and accessibility features are built into the same document pipeline.',
    color: 'blue',
    size: 'normal',
  },
  {
    icon: <Shield size={28} />,
    title: 'Enterprise Security',
    description: 'AES-128 permissions and PKCS#7 signatures without dragging in a commercial runtime.',
    color: 'purple',
    size: 'normal',
  },
  {
    icon: <Globe size={28} />,
    title: 'Language Agnostic',
    description: 'Use the REST API from any stack, not just Go or a single vendor SDK.',
    color: 'teal',
    size: 'normal',
  },
  {
    icon: <Code size={28} />,
    title: 'Native Python Support',
    description: 'Direct CGO bindings plus a web client when Python needs either throughput or portability.',
    color: 'purple',
    size: 'normal',
  },
  {
    icon: <Box size={28} />,
    title: 'Single Binary Deploy',
    description: 'Small operational footprint with Docker-ready packaging and no library SDK rollout across services.',
    color: 'blue',
    size: 'wide',
  },
]

const Comparison = () => {
  const [isVisible, setIsVisible] = useState({})
  const [activeCategory, setActiveCategory] = useState('all')

  useEffect(() => {
    window.scrollTo(0, 0)
  }, [])

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsVisible((prev) => ({ ...prev, [entry.target.id]: true }))
          }
        })
      },
      { threshold: 0.1 }
    )

    const sections = document.querySelectorAll('[id^="section-"]')
    sections.forEach((section) => observer.observe(section))

    return () => observer.disconnect()
  }, [])

  const primaryCompetitor = competitors.find((competitor) => competitor.isOurs)
  const filteredFeatures = activeCategory === 'all'
    ? features
    : features.filter((feature) => feature.category === activeCategory)
  const spotlightKeys = ['performance', 'deployment', 'pythonSupport', 'pdfaCompliance']
  const highlightChips = [
    {
      icon: <DollarSign size={18} />,
      label: 'License',
      value: primaryCompetitor.pricing,
      accent: '#4ecdc4',
    },
    {
      icon: <Zap size={18} />,
      label: 'Best benchmark',
      value: primaryCompetitor.performance,
      accent: '#007acc',
    },
    {
      icon: <Code size={18} />,
      label: 'Integration model',
      value: primaryCompetitor.integration,
      accent: '#f093fb',
    },
    {
      icon: <Shield size={18} />,
      label: 'Compliance',
      value: `${primaryCompetitor.pdfaCompliance} + ${primaryCompetitor.pdfuaCompliance}`,
      accent: '#ffc107',
    },
  ]

  const getFeatureStatus = (value, featureKey) => {
    const lowerValue = String(value).toLowerCase()

    if (featureKey === 'pricing') {
      return lowerValue.includes('free') ? 'positive' : 'negative'
    }

    if (
      lowerValue.includes('not supported') ||
      lowerValue === 'none' ||
      lowerValue.includes('deprecated') ||
      lowerValue.includes('n/a')
    ) {
      return 'negative'
    }

    if (lowerValue.includes('limited') || lowerValue.includes('manual') || lowerValue.includes('add-on')) {
      return 'partial'
    }

    return 'positive'
  }

  const getStatusMeta = (status) => {
    if (status === 'negative') {
      return { label: 'Missing', icon: <X size={14} />, color: '#ef4444' }
    }

    if (status === 'partial') {
      return { label: 'Partial', icon: <TrendingUp size={14} />, color: '#ffc107' }
    }

    return { label: 'Supported', icon: <CheckCircle size={14} />, color: '#22c55e' }
  }

  const getFeatureCounts = (competitor) => features
    .filter((feature) => feature.key !== 'pricing')
    .reduce((counts, feature) => {
      const status = getFeatureStatus(competitor[feature.key], feature.key)
      counts[status] += 1
      return counts
    }, { positive: 0, partial: 0, negative: 0 })

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />

      <section id="section-header" style={{ padding: '3rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div
            className="animate-fadeInUp"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.5rem 1rem',
              background: 'rgba(78, 205, 196, 0.1)',
              border: '1px solid rgba(78, 205, 196, 0.3)',
              borderRadius: '50px',
              marginBottom: '1.5rem',
              color: '#4ecdc4',
              fontSize: '0.9rem',
              fontWeight: '500',
            }}
          >
            <Sparkles size={16} />
            Compare with industry leaders
          </div>

          <h1
            className={`gradient-text animate-fadeInUp stagger-animation ${isVisible['section-header'] ? 'visible' : ''}`}
            style={{
              fontSize: 'clamp(2rem, 5vw, 3.5rem)',
              fontWeight: '800',
              marginBottom: '1rem',
              animationDelay: '0.2s',
            }}
          >
            Feature Comparison
          </h1>

          <p
            className={`animate-fadeInUp stagger-animation ${isVisible['section-header'] ? 'visible' : ''}`}
            style={{
              fontSize: '1.2rem',
              color: 'hsl(var(--muted-foreground))',
              maxWidth: '700px',
              margin: '0 auto',
              animationDelay: '0.4s',
            }}
          >
            See how GoPdfSuit compares against established PDF libraries and commercial platforms without forcing every data point into the same cramped card layout.
          </p>

          <div className="comparison-highlight-grid">
            {highlightChips.map((chip, index) => (
              <div
                key={chip.label}
                className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-header'] ? 'visible' : ''}`}
                style={{
                  padding: '1.1rem 1.25rem',
                  textAlign: 'left',
                  animationDelay: `${0.55 + index * 0.1}s`,
                }}
              >
                <div
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    width: '2.5rem',
                    height: '2.5rem',
                    borderRadius: '14px',
                    marginBottom: '0.9rem',
                    background: `${chip.accent}1a`,
                    color: chip.accent,
                  }}
                >
                  {chip.icon}
                </div>
                <div className="comparison-chip-label">{chip.label}</div>
                <div className="comparison-chip-value">{chip.value}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section style={{ padding: '1rem 0' }}>
        <div className="container comparison-wide-container">
          <div className="glass-card comparison-filter-shell">
            <div className="comparison-filter-header">
              <div>
                <div className="comparison-eyebrow">Filter the matrix</div>
                <h2 style={{ marginBottom: '0.35rem', fontSize: '1.3rem' }}>Feature categories</h2>
                <p style={{ marginBottom: 0, maxWidth: '42rem' }}>
                  Start with the overview cards, then switch categories to compare detailed capabilities without turning every competitor into a long, unreadable vertical stack.
                </p>
              </div>
              <div className="comparison-filter-pills">
                {categories.map((category) => (
                  <button
                    key={category.id}
                    onClick={() => setActiveCategory(category.id)}
                    className={`comparison-filter-button ${activeCategory === category.id ? 'comparison-filter-button-active' : ''}`}
                  >
                    {category.icon}
                    {category.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="section-comparison" style={{ padding: '1rem 0 3rem' }}>
        <div className="container comparison-wide-container">
          <div className="comparison-overview-grid">
            {competitors.map((competitor, compIndex) => {
              const counts = getFeatureCounts(competitor)

              return (
                <div
                  key={competitor.name}
                  className={`glass-card animate-fadeInScale stagger-animation comparison-overview-card ${competitor.isOurs ? 'comparison-overview-card-primary' : ''} ${isVisible['section-comparison'] ? 'visible' : ''}`}
                  style={{ animationDelay: `${0.1 + compIndex * 0.08}s`, padding: '1.5rem' }}
                >
                  <div className="comparison-overview-header">
                    <div>
                      <div className="comparison-card-kicker">{competitor.isOurs ? 'GoPdfSuit' : 'Alternative'}</div>
                      <h3
                        style={{
                          marginBottom: '0.45rem',
                          fontSize: '1.5rem',
                          fontWeight: '700',
                          color: competitor.isOurs ? '#4ecdc4' : 'hsl(var(--foreground))',
                        }}
                      >
                        {competitor.name}
                      </h3>
                      <p style={{ marginBottom: 0, fontSize: '0.95rem' }}>{competitor.deployment}</p>
                    </div>
                    {competitor.isOurs ? <div className="comparison-card-badge">Recommended</div> : null}
                  </div>

                  <div className="comparison-price-pill">{competitor.pricing}</div>

                  <div className="comparison-score-grid">
                    <div className="comparison-score-card comparison-score-positive">
                      <span className="comparison-score-value">{counts.positive}</span>
                      <span className="comparison-score-label">Supported</span>
                    </div>
                    <div className="comparison-score-card comparison-score-partial">
                      <span className="comparison-score-value">{counts.partial}</span>
                      <span className="comparison-score-label">Partial</span>
                    </div>
                    <div className="comparison-score-card comparison-score-negative">
                      <span className="comparison-score-value">{counts.negative}</span>
                      <span className="comparison-score-label">Missing</span>
                    </div>
                  </div>

                  <div className="comparison-spotlight-list">
                    {spotlightKeys.map((key) => {
                      const feature = features.find((item) => item.key === key)
                      const value = competitor[key]
                      const status = getFeatureStatus(value, key)
                      const meta = getStatusMeta(status)

                      return (
                        <div key={key} className="comparison-spotlight-row">
                          <div className="comparison-spotlight-label-wrap">
                            <span className="comparison-spotlight-icon">{feature.icon}</span>
                            <span className="comparison-spotlight-label">{feature.label}</span>
                          </div>
                          <div className="comparison-spotlight-value-wrap">
                            <span style={{ color: meta.color, display: 'inline-flex', alignItems: 'center' }}>
                              {meta.icon}
                            </span>
                            <span className="comparison-spotlight-value">{value}</span>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )
            })}
          </div>

          <div
            className={`glass-card animate-fadeInScale stagger-animation comparison-matrix-shell ${isVisible['section-comparison'] ? 'visible' : ''}`}
            style={{ animationDelay: '0.45s' }}
          >
            <div className="comparison-matrix-header">
              <div>
                <div className="comparison-eyebrow">Detailed breakdown</div>
                <h2 style={{ fontSize: '1.55rem', marginBottom: '0.45rem' }}>Feature-by-feature comparison</h2>
                <p style={{ marginBottom: 0, maxWidth: '44rem' }}>
                  The overview cards surface the big trade-offs. This matrix keeps the full detail visible and readable, especially on desktop, while still degrading cleanly to horizontal scroll on small screens.
                </p>
              </div>
              <div className="comparison-matrix-summary-pill">
                Showing {filteredFeatures.length} {activeCategory === 'all' ? 'capabilities' : categories.find((category) => category.id === activeCategory)?.label.toLowerCase()}
              </div>
            </div>

            <div className="comparison-matrix-scroll custom-scrollbar">
              <table className="comparison-matrix-table">
                <thead>
                  <tr>
                    <th>Capability</th>
                    {competitors.map((competitor) => (
                      <th key={competitor.name} className={competitor.isOurs ? 'comparison-column-highlight' : ''}>
                        <div className="comparison-column-title">{competitor.name}</div>
                        <div className="comparison-column-subtitle">{competitor.pricing}</div>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {filteredFeatures.map((feature) => (
                    <tr key={feature.key}>
                      <th scope="row" className="comparison-feature-column">
                        <div className="comparison-feature-heading">
                          <span className="comparison-feature-icon">{feature.icon}</span>
                          <div>
                            <div className="comparison-feature-title">{feature.label}</div>
                            <div className="comparison-feature-category">{categories.find((category) => category.id === feature.category)?.label}</div>
                          </div>
                        </div>
                      </th>
                      {competitors.map((competitor) => {
                        const value = competitor[feature.key]
                        const status = getFeatureStatus(value, feature.key)
                        const meta = getStatusMeta(status)

                        return (
                          <td key={`${competitor.name}-${feature.key}`}>
                            <div className={`comparison-matrix-cell comparison-matrix-cell-${status}`}>
                              <div className="comparison-matrix-cell-status" style={{ color: meta.color }}>
                                {meta.icon}
                                <span>{meta.label}</span>
                              </div>
                              <div className="comparison-matrix-cell-value">{value}</div>
                            </div>
                          </td>
                        )
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <section id="section-benchmarks" style={{ padding: '2rem 0' }}>
        <div className="container comparison-wide-container comparison-benchmark-container">
          <div
            className={`animate-fadeInScale stagger-animation ${isVisible['section-benchmarks'] ? 'visible' : ''}`}
            style={{ animationDelay: '0.2s' }}
          >
            <PerformanceSection isVisible={isVisible['section-benchmarks']} />
          </div>
        </div>
      </section>

      <section id="section-advantages" style={{ padding: '3rem 0' }}>
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-advantages'] ? 'visible' : ''}`}
            style={{ marginBottom: '2.5rem' }}
          >
            <h2 className="gradient-text" style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>
              Why Choose GoPdfSuit?
            </h2>
            <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>
              Key advantages that set the platform apart once you step back from the raw feature checklist.
            </p>
          </div>

          <div className="bento-grid">
            {advantages.map((advantage, index) => {
              const sizeClass = advantage.size === 'large'
                ? 'bento-item-large'
                : advantage.size === 'wide'
                  ? 'bento-item-wide'
                  : ''

              return (
                <div
                  key={advantage.title}
                  className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-advantages'] ? 'visible' : ''} ${sizeClass}`}
                  style={{
                    padding: advantage.size === 'large' ? '2.5rem' : '2rem',
                    animationDelay: `${0.2 + index * 0.1}s`,
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <div className={`feature-icon-box ${advantage.color}`} style={{ marginBottom: '1.5rem' }}>
                    {advantage.icon}
                  </div>
                  <h3
                    style={{
                      marginBottom: '0.75rem',
                      color: 'hsl(var(--foreground))',
                      fontSize: advantage.size === 'large' ? '1.5rem' : '1.25rem',
                      fontWeight: '700',
                    }}
                  >
                    {advantage.title}
                  </h3>
                  <p
                    style={{
                      color: 'hsl(var(--muted-foreground))',
                      marginBottom: 0,
                      lineHeight: 1.7,
                      flex: 1,
                      fontSize: advantage.size === 'large' ? '1.05rem' : '0.95rem',
                    }}
                  >
                    {advantage.description}
                  </p>
                </div>
              )
            })}
          </div>
        </div>
      </section>

      <section id="section-cta" style={{ padding: '3rem 0 5rem' }}>
        <div className="container">
          <div
            className={`glass-card animate-fadeInUp stagger-animation ${isVisible['section-cta'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
              maxWidth: '700px',
              margin: '0 auto',
              padding: '3rem',
              animationDelay: '0.2s',
            }}
          >
            <div
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.5rem 1rem',
                background: 'rgba(78, 205, 196, 0.1)',
                border: '1px solid rgba(78, 205, 196, 0.3)',
                borderRadius: '50px',
                marginBottom: '1.5rem',
                color: '#4ecdc4',
                fontSize: '0.85rem',
              }}
            >
              <Sparkles size={14} />
              Open Source & Free Forever
            </div>
            <h2 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem', fontSize: '2rem' }}>
              Ready to Try GoPdfSuit?
            </h2>
            <p
              style={{
                color: 'hsl(var(--muted-foreground))',
                marginBottom: '2rem',
                fontSize: '1.1rem',
                lineHeight: '1.7',
              }}
            >
              Experience the power of fast, free, and flexible PDF generation today.
            </p>
            <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
              <Link
                to="/editor"
                className="btn-glow glow-on-hover"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  textDecoration: 'none',
                }}
              >
                <FileText size={20} />
                Try PDF Generator
                <ArrowRight size={18} />
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
                <Star size={18} />
                Star on GitHub
              </a>
            </div>
          </div>
        </div>
      </section>

      <style>
        {`
          .comparison-wide-container {
            max-width: min(1520px, calc(100vw - 2rem));
          }

          .comparison-benchmark-container {
            max-width: min(1580px, calc(100vw - 2rem));
          }

          .comparison-highlight-grid {
            display: grid;
            grid-template-columns: repeat(4, minmax(0, 1fr));
            gap: 1.5rem;
            margin-top: 2.5rem;
          }

          .comparison-chip-label {
            font-size: 0.75rem;
            color: hsl(var(--muted-foreground));
            text-transform: uppercase;
            letter-spacing: 0.08em;
            margin-bottom: 0.35rem;
          }

          .comparison-chip-value {
            color: hsl(var(--foreground));
            font-size: 0.98rem;
            line-height: 1.45;
            font-weight: 600;
          }

          .comparison-filter-shell {
            padding: 1.4rem;
            margin-bottom: 2rem;
          }

          .comparison-filter-header {
            display: flex;
            align-items: flex-start;
            justify-content: space-between;
            gap: 1.5rem;
          }

          .comparison-filter-pills {
            display: flex;
            flex-wrap: wrap;
            justify-content: flex-end;
            gap: 0.75rem;
          }

          .comparison-filter-button {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.75rem 1.15rem;
            border-radius: 999px;
            border: 1px solid rgba(255, 255, 255, 0.15);
            background: rgba(255, 255, 255, 0.04);
            color: hsl(var(--muted-foreground));
            cursor: pointer;
            transition: all 0.25s ease;
            font-weight: 600;
            font-size: 0.9rem;
          }

          .comparison-filter-button:hover {
            transform: translateY(-2px);
            border-color: rgba(78, 205, 196, 0.35);
            color: hsl(var(--foreground));
          }

          .comparison-filter-button-active {
            border-color: rgba(78, 205, 196, 0.55);
            background: rgba(78, 205, 196, 0.14);
            color: #4ecdc4;
            box-shadow: inset 0 0 0 1px rgba(78, 205, 196, 0.18);
          }

          .comparison-overview-grid {
            display: grid;
            grid-template-columns: repeat(12, minmax(0, 1fr));
            gap: 1.25rem;
            margin-bottom: 1.5rem;
          }

          .comparison-overview-card {
            grid-column: span 4;
            display: flex;
            flex-direction: column;
            gap: 1.25rem;
            min-height: 100%;
          }

          .comparison-overview-card-primary {
            grid-column: span 6;
            background: linear-gradient(180deg, rgba(78, 205, 196, 0.14), rgba(255, 255, 255, 0.04));
            border-color: rgba(78, 205, 196, 0.38);
          }

          .comparison-overview-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            gap: 1rem;
          }

          .comparison-card-kicker {
            font-size: 0.72rem;
            color: hsl(var(--muted-foreground));
            text-transform: uppercase;
            letter-spacing: 0.08em;
            margin-bottom: 0.55rem;
            font-weight: 700;
          }

          .comparison-card-badge {
            display: inline-flex;
            align-items: center;
            padding: 0.45rem 0.75rem;
            border-radius: 999px;
            background: rgba(78, 205, 196, 0.16);
            color: #4ecdc4;
            font-size: 0.78rem;
            font-weight: 700;
            border: 1px solid rgba(78, 205, 196, 0.28);
          }

          .comparison-price-pill {
            display: inline-flex;
            align-items: center;
            width: fit-content;
            padding: 0.45rem 0.75rem;
            border-radius: 999px;
            background: rgba(255, 255, 255, 0.06);
            border: 1px solid rgba(255, 255, 255, 0.1);
            color: hsl(var(--foreground));
            font-size: 0.85rem;
            font-weight: 600;
          }

          .comparison-score-grid {
            display: grid;
            grid-template-columns: repeat(3, minmax(0, 1fr));
            gap: 0.75rem;
          }

          .comparison-score-card {
            border-radius: 14px;
            padding: 0.9rem;
            display: flex;
            flex-direction: column;
            gap: 0.2rem;
            min-height: 5rem;
          }

          .comparison-score-positive {
            background: rgba(34, 197, 94, 0.1);
            border: 1px solid rgba(34, 197, 94, 0.18);
          }

          .comparison-score-partial {
            background: rgba(255, 193, 7, 0.1);
            border: 1px solid rgba(255, 193, 7, 0.18);
          }

          .comparison-score-negative {
            background: rgba(239, 68, 68, 0.1);
            border: 1px solid rgba(239, 68, 68, 0.18);
          }

          .comparison-score-value {
            font-size: 1.5rem;
            font-weight: 800;
            color: hsl(var(--foreground));
            line-height: 1;
          }

          .comparison-score-label {
            font-size: 0.78rem;
            color: hsl(var(--muted-foreground));
            text-transform: uppercase;
            letter-spacing: 0.06em;
          }

          .comparison-spotlight-list {
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
          }

          .comparison-spotlight-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 1rem;
            padding-top: 0.75rem;
            border-top: 1px solid rgba(255, 255, 255, 0.08);
          }

          .comparison-spotlight-label-wrap,
          .comparison-spotlight-value-wrap {
            display: flex;
            align-items: center;
            gap: 0.55rem;
          }

          .comparison-spotlight-icon {
            display: inline-flex;
            color: hsl(var(--muted-foreground));
          }

          .comparison-spotlight-label {
            color: hsl(var(--muted-foreground));
            font-size: 0.86rem;
          }

          .comparison-spotlight-value {
            color: hsl(var(--foreground));
            font-size: 0.88rem;
            font-weight: 600;
            text-align: right;
          }

          .comparison-matrix-shell {
            padding: 1.5rem;
          }

          .comparison-matrix-header {
            display: flex;
            align-items: flex-end;
            justify-content: space-between;
            gap: 1rem;
            margin-bottom: 1.25rem;
          }

          .comparison-matrix-summary-pill {
            display: inline-flex;
            align-items: center;
            padding: 0.6rem 0.9rem;
            border-radius: 999px;
            background: rgba(0, 122, 204, 0.14);
            color: #8acbff;
            font-size: 0.84rem;
            border: 1px solid rgba(0, 122, 204, 0.22);
            white-space: nowrap;
          }

          .comparison-matrix-scroll {
            overflow-x: visible;
            padding-bottom: 0.5rem;
          }

          .comparison-matrix-table {
            width: 100%;
            min-width: 0;
            border-collapse: separate;
            border-spacing: 0;
            table-layout: fixed;
          }

          .comparison-matrix-table thead th {
            padding: 0 0.75rem 1rem;
            text-align: left;
            font-size: 0.8rem;
            color: hsl(var(--muted-foreground));
            text-transform: uppercase;
            letter-spacing: 0.08em;
          }

          .comparison-column-title {
            color: hsl(var(--foreground));
            font-size: 0.95rem;
            font-weight: 700;
            text-transform: none;
            letter-spacing: normal;
            margin-bottom: 0.25rem;
            word-break: break-word;
          }

          .comparison-column-subtitle {
            font-size: 0.76rem;
            color: hsl(var(--muted-foreground));
            text-transform: none;
            letter-spacing: normal;
            word-break: break-word;
          }

          .comparison-column-highlight .comparison-column-title {
            color: #4ecdc4;
          }

          .comparison-matrix-table tbody th,
          .comparison-matrix-table tbody td {
            padding: 0.65rem 0.75rem 0;
            vertical-align: top;
          }

          .comparison-feature-column {
            width: 14rem;
            background: transparent;
            backdrop-filter: none;
          }

          .comparison-feature-heading {
            display: flex;
            align-items: flex-start;
            gap: 0.75rem;
            padding: 0.7rem 0;
            min-height: 100%;
            border: none;
            background: transparent;
          }

          .comparison-feature-icon {
            display: inline-flex;
            color: #4ecdc4;
            margin-top: 0.1rem;
          }

          .comparison-feature-title {
            color: hsl(var(--foreground));
            font-size: 0.94rem;
            font-weight: 700;
            margin-bottom: 0.18rem;
          }

          .comparison-feature-category {
            color: hsl(var(--muted-foreground));
            font-size: 0.76rem;
            font-weight: 500;
          }

          .comparison-matrix-cell {
            min-height: 6.5rem;
            border-radius: 16px;
            padding: 0.9rem;
            border: 1px solid rgba(255, 255, 255, 0.08);
            background: rgba(255, 255, 255, 0.03);
          }

          .comparison-matrix-cell-positive {
            background: linear-gradient(180deg, rgba(34, 197, 94, 0.12), rgba(255, 255, 255, 0.04));
            border-color: rgba(34, 197, 94, 0.18);
          }

          .comparison-matrix-cell-partial {
            background: linear-gradient(180deg, rgba(255, 193, 7, 0.12), rgba(255, 255, 255, 0.04));
            border-color: rgba(255, 193, 7, 0.18);
          }

          .comparison-matrix-cell-negative {
            background: linear-gradient(180deg, rgba(239, 68, 68, 0.12), rgba(255, 255, 255, 0.04));
            border-color: rgba(239, 68, 68, 0.18);
          }

          .comparison-matrix-cell-status {
            display: inline-flex;
            align-items: center;
            gap: 0.4rem;
            font-size: 0.78rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.06em;
            margin-bottom: 0.65rem;
          }

          .comparison-matrix-cell-value {
            color: hsl(var(--foreground));
            font-size: 0.86rem;
            line-height: 1.55;
            word-break: break-word;
          }

          @media (max-width: 1100px) {
            .comparison-highlight-grid {
              grid-template-columns: repeat(2, minmax(0, 1fr));
            }

            .comparison-wide-container,
            .comparison-benchmark-container {
              max-width: min(100%, calc(100vw - 2rem));
            }

            .comparison-filter-header,
            .comparison-matrix-header {
              flex-direction: column;
              align-items: stretch;
            }

            .comparison-filter-pills {
              justify-content: flex-start;
            }

            .comparison-overview-card,
            .comparison-overview-card-primary {
              grid-column: span 6;
            }
          }

          @media (max-width: 820px) {
            .comparison-highlight-grid,
            .comparison-score-grid {
              grid-template-columns: 1fr;
            }

            .comparison-overview-grid {
              grid-template-columns: 1fr;
            }

            .comparison-overview-card,
            .comparison-overview-card-primary {
              grid-column: auto;
            }

            .comparison-overview-header,
            .comparison-spotlight-row {
              flex-direction: column;
              align-items: flex-start;
            }

            .comparison-spotlight-value {
              text-align: left;
            }
          }

          @media (max-width: 600px) {
            .comparison-filter-shell,
            .comparison-matrix-shell {
              padding: 1rem;
            }

            .comparison-highlight-grid {
              gap: 1rem;
            }

            .comparison-feature-column {
              width: 11.5rem;
            }
          }
        `}
      </style>
    </div>
  )
}

export default Comparison
