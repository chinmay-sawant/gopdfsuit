import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  CheckCircle,
  X,
  TrendingUp,
  Zap,
  DollarSign,
  Box,
  Code,
  Globe,
  Star,
  Shield,
  FileText,
  Sparkles,
  ArrowRight
} from 'lucide-react'
import PerformanceSection from '../components/PerformanceSection'
import BackgroundAnimation from '../components/BackgroundAnimation'

const Comparison = () => {
  const [isVisible, setIsVisible] = useState({})
  const [activeCategory, setActiveCategory] = useState('all')

  // Scroll to top on mount
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [])

  // Intersection Observer for scroll animations
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsVisible(prev => ({ ...prev, [entry.target.id]: true }))
          }
        })
      },
      { threshold: 0.1 }
    )

    // Observe all sections
    const sections = document.querySelectorAll('[id^="section-"]')
    sections.forEach((section) => observer.observe(section))

    return () => observer.disconnect()
  }, [])

  const competitors = [
    {
      name: 'GoPdfSuit',
      isOurs: true,
      pricing: 'Free (MIT License)',
      performance: 'Ultra Fast (Sub-ms - ~7ms)',
      deployment: 'Microservice/Sidecar/Docker',
      memory: 'In-Memory Processing',
      integration: 'REST API + Native Python + gopdflib',
      template: 'JSON-based Templates',
      webInterface: 'Built-in Viewer/Editor',
      formFilling: 'XFDF Advanced Detection',
      pdfMerge: 'Drag & Drop + Form Preservation',
      htmlConversion: 'gochromedp (Chromium)',
      multipage: 'Auto Page Breaks',
      styling: 'Font Styles + Borders + Images',
      pdfaCompliance: 'PDF/A-4 with ICC Profiles',
      pdfuaCompliance: 'PDF/UA-2 Accessibility',
      encryption: 'AES-128 with Permissions',
      digitalSignatures: 'PKCS#7 + Visual Appearance',
      fontEmbedding: 'TrueType Subsetting',
      bookmarks: 'Outlines + Hyperlinks',
      dockerSupport: 'Multi-stage Alpine Image',
      pythonSupport: 'Native CGO + API Client',
      maintenance: 'Single Binary'
    },
    {
      name: 'UniPDF',
      pricing: '$3,000+/year',
      performance: 'High (Go)',
      deployment: 'Library Integration',
      memory: 'Efficient',
      integration: 'Go',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: 'Full Support',
      pdfMerge: 'Supported',
      htmlConversion: 'Limited',
      multipage: 'Manual Control',
      styling: 'Code-based',
      pdfaCompliance: 'PDF/A',
      pdfuaCompliance: 'PDF/UA',
      encryption: 'Supported',
      digitalSignatures: 'Supported',
      fontEmbedding: 'Supported',
      bookmarks: 'Supported',
      dockerSupport: 'N/A (Library)',
      pythonSupport: 'Not Supported',
      maintenance: 'Commercial Support'
    },
    {
      name: 'Aspose.PDF',
      pricing: '$1,199+/year',
      performance: 'High (C++)',
      deployment: 'Library Integration',
      memory: 'High',
      integration: '.NET/Java/C++/Go',
      template: 'XML/Code',
      webInterface: 'Cloud Only',
      formFilling: 'Full Support',
      pdfMerge: 'Supported',
      htmlConversion: 'Strong Support',
      multipage: 'Supported',
      styling: 'Comprehensive',
      pdfaCompliance: 'PDF/A-1 to A-3',
      pdfuaCompliance: 'PDF/UA',
      encryption: 'AES-256',
      digitalSignatures: 'Supported',
      fontEmbedding: 'Supported',
      bookmarks: 'Supported',
      dockerSupport: 'N/A (Library)',
      pythonSupport: 'Via .NET Wrapper',
      maintenance: 'Commercial Support'
    },
    {
      name: 'iText 7',
      pricing: '$3,500/dev/year',
      performance: 'Moderate',
      deployment: 'Library Integration',
      memory: 'Mixed',
      integration: 'Java/.NET',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: 'Full Support',
      pdfMerge: 'Programmatic',
      htmlConversion: 'pdfHTML add-on ($)',
      multipage: 'Manual Control',
      styling: 'Advanced',
      pdfaCompliance: 'PDF/A-1 to PDF/A-3',
      pdfuaCompliance: 'PDF/UA-1',
      encryption: 'AES-256',
      digitalSignatures: 'Full PKI Support',
      fontEmbedding: 'Full Embedding',
      bookmarks: 'Full Support',
      dockerSupport: 'N/A (Library)',
      pythonSupport: 'Via Wrapper',
      maintenance: 'Library Updates'
    },
    {
      name: 'wkhtmltopdf',
      pricing: 'Free (LGPL)',
      performance: 'Slow (Process spawn)',
      deployment: 'Binary + WebKit',
      memory: 'High (WebKit)',
      integration: 'Command Line',
      template: 'HTML/CSS',
      webInterface: 'None',
      formFilling: 'Not Supported',
      pdfMerge: 'Not Supported',
      htmlConversion: 'Native (Outdated WebKit)',
      multipage: 'CSS Page Breaks',
      styling: 'CSS-based',
      pdfaCompliance: 'Not Supported',
      pdfuaCompliance: 'Not Supported',
      encryption: 'Not Supported',
      digitalSignatures: 'Not Supported',
      fontEmbedding: 'Automatic',
      bookmarks: 'Limited (TOC)',
      dockerSupport: 'Manual Setup',
      pythonSupport: 'Wrapper (pdfkit)',
      maintenance: 'Deprecated'
    }
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
    { key: 'maintenance', label: 'Maintenance', icon: <CheckCircle size={18} />, category: 'integration' }
  ]

  const filteredFeatures = activeCategory === 'all'
    ? features
    : features.filter(f => f.category === activeCategory)

  const getFeatureStatus = (value) => {
    const lowerValue = value.toLowerCase()
    if (lowerValue.includes('not supported') || lowerValue === 'none' || lowerValue.includes('deprecated') || lowerValue.includes('n/a')) {
      return 'negative'
    }
    if (lowerValue.includes('limited') || lowerValue.includes('manual') || lowerValue.includes('add-on')) {
      return 'partial'
    }
    return 'positive'
  }

  // Interactive dots canvas background (Antigravity-style) - Imported from components

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />

      {/* Header */}
      <section
        id="section-header"
        style={{ padding: '3rem 0 2rem', textAlign: 'center' }}
      >
        <div className="container">

          {/* Sparkle badge */}
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
            Compare with Industry Leaders
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
            See how GoPdfSuit compares against industry-leading PDF libraries and commercial solutions
          </p>
        </div>
      </section>

      {/* Category Tabs */}
      <section style={{ padding: '1rem 0' }}>
        <div className="container">
          <div
            style={{
              display: 'flex',
              gap: '0.75rem',
              flexWrap: 'wrap',
              justifyContent: 'center',
              marginBottom: '2rem',
            }}
          >
            {categories.map((cat) => (
              <button
                key={cat.id}
                onClick={() => setActiveCategory(cat.id)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  padding: '0.75rem 1.25rem',
                  borderRadius: '50px',
                  border: activeCategory === cat.id
                    ? '2px solid rgba(78, 205, 196, 0.6)'
                    : '1px solid rgba(255, 255, 255, 0.15)',
                  background: activeCategory === cat.id
                    ? 'rgba(78, 205, 196, 0.15)'
                    : 'rgba(255, 255, 255, 0.05)',
                  color: activeCategory === cat.id ? '#4ecdc4' : 'hsl(var(--muted-foreground))',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease',
                  fontWeight: activeCategory === cat.id ? '600' : '500',
                  fontSize: '0.9rem',
                }}
              >
                {cat.icon}
                {cat.label}
              </button>
            ))}
          </div>
        </div>
      </section>

      {/* Comparison Cards */}
      <section
        id="section-comparison"
        style={{ padding: '1rem 0 3rem' }}
      >
        <div className="container">
          <div className="comparison-cards-grid">
            {competitors.map((competitor, compIndex) => (
              <div
                key={competitor.name}
                className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison'] ? 'visible' : ''}`}
                style={{
                  padding: '1.5rem',
                  animationDelay: `${0.1 + compIndex * 0.1}s`,
                  border: competitor.isOurs
                    ? '2px solid rgba(78, 205, 196, 0.5)'
                    : '1px solid var(--glass-border)',
                  position: 'relative',
                  overflow: 'hidden',
                }}
              >
                {/* Header */}
                <div style={{ marginBottom: '1.5rem', textAlign: 'center' }}>
                  {competitor.isOurs && (
                    <div
                      style={{
                        position: 'absolute',
                        top: '0',
                        left: '0',
                        right: '0',
                        padding: '0.35rem',
                        background: 'linear-gradient(135deg, rgba(78, 205, 196, 0.3), rgba(0, 122, 204, 0.3))',
                        fontSize: '0.7rem',
                        fontWeight: '700',
                        color: '#4ecdc4',
                        textTransform: 'uppercase',
                        letterSpacing: '1px',
                      }}
                    >
                      ⭐ Our Solution
                    </div>
                  )}
                  <h3
                    style={{
                      marginTop: competitor.isOurs ? '1.5rem' : '0',
                      marginBottom: '0.5rem',
                      fontSize: '1.4rem',
                      fontWeight: '700',
                      color: competitor.isOurs ? '#4ecdc4' : 'hsl(var(--foreground))',
                    }}
                  >
                    {competitor.name}
                  </h3>
                  <div
                    style={{
                      fontSize: '0.85rem',
                      padding: '0.35rem 0.75rem',
                      borderRadius: '20px',
                      display: 'inline-block',
                      background: competitor.isOurs
                        ? 'rgba(78, 205, 196, 0.15)'
                        : 'rgba(255, 193, 7, 0.1)',
                      color: competitor.isOurs ? '#4ecdc4' : '#ffc107',
                      fontWeight: '600',
                    }}
                  >
                    {competitor.pricing}
                  </div>
                </div>

                {/* Features List */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  {filteredFeatures.filter(f => f.key !== 'pricing').map((feature) => {
                    const value = competitor[feature.key]
                    const status = getFeatureStatus(value)

                    return (
                      <div
                        key={feature.key}
                        style={{
                          display: 'flex',
                          alignItems: 'flex-start',
                          gap: '0.75rem',
                          padding: '0.5rem',
                          borderRadius: '8px',
                          background: status === 'positive'
                            ? 'rgba(34, 197, 94, 0.08)'
                            : status === 'partial'
                              ? 'rgba(255, 193, 7, 0.08)'
                              : 'rgba(239, 68, 68, 0.08)',
                        }}
                      >
                        <div
                          style={{
                            color: status === 'positive'
                              ? '#22c55e'
                              : status === 'partial'
                                ? '#ffc107'
                                : '#ef4444',
                            flexShrink: 0,
                            marginTop: '2px',
                          }}
                        >
                          {status === 'positive' ? <CheckCircle size={16} /> :
                            status === 'partial' ? <TrendingUp size={16} /> :
                              <X size={16} />}
                        </div>
                        <div style={{ flex: 1 }}>
                          <div
                            style={{
                              fontSize: '0.75rem',
                              color: 'hsl(var(--muted-foreground))',
                              marginBottom: '0.15rem',
                              textTransform: 'uppercase',
                              letterSpacing: '0.5px',
                            }}
                          >
                            {feature.label}
                          </div>
                          <div
                            className="feature-value"
                            style={{
                              fontSize: '0.8rem',
                              color: 'hsl(var(--foreground))',
                              fontWeight: '500',
                              lineHeight: '1.4',
                              wordWrap: 'break-word',
                              overflowWrap: 'break-word',
                              wordBreak: 'break-word',
                            }}
                          >
                            {value}
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Performance Benchmarks */}
      <section
        id="section-benchmarks"
        style={{ padding: '2rem 0' }}
      >
        <div className="container">
          <div
            className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-benchmarks'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
              maxWidth: '800px',
              margin: '0 auto',
              padding: '2rem',
              animationDelay: '0.2s',
            }}
          >
            <PerformanceSection isVisible={isVisible['section-benchmarks']} />
          </div>
        </div>
      </section>

      {/* Key Advantages */}
      <section
        id="section-advantages"
        style={{ padding: '3rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-advantages'] ? 'visible' : ''}`}
            style={{ marginBottom: '2.5rem' }}
          >
            <h2
              className="gradient-text"
              style={{
                fontSize: '2.5rem',
                marginBottom: '1rem',
              }}
            >
              Why Choose GoPdfSuit?
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
              maxWidth: '600px',
              margin: '0 auto',
            }}>
              Key advantages that set us apart from the competition
            </p>
          </div>

          <div className="bento-grid">
            {[
              {
                icon: <Zap size={28} />,
                title: 'Ultra Fast Performance',
                description: 'Sub-millisecond to ~7ms response times vs moderate performance of commercial libraries',
                color: 'teal',
                size: 'large'
              },
              {
                icon: <DollarSign size={28} />,
                title: 'Zero Licensing Cost',
                description: 'MIT license vs $2,750-$3,500/dev/year for commercial solutions',
                color: 'green',
                size: 'normal'
              },
              {
                icon: <Shield size={28} />,
                title: 'PDF/A-4 & PDF/UA-2',
                description: 'Full archival and accessibility compliance with sRGB ICC profiles',
                color: 'blue',
                size: 'normal'
              },
              {
                icon: <Shield size={28} />,
                title: 'Enterprise Security',
                description: 'AES-128 encryption with permissions + PKCS#7 digital signatures',
                color: 'purple',
                size: 'normal'
              },
              {
                icon: <Globe size={28} />,
                title: 'Language Agnostic',
                description: 'REST API works with any programming language',
                color: 'teal',
                size: 'normal'
              },
              {
                icon: <Code size={28} />,
                title: 'Native Python Support',
                description: 'Direct CGO bindings for high-performance Python integration + Web Client',
                color: 'purple',
                size: 'normal'
              },
              {
                icon: <Box size={28} />,
                title: 'Single Binary Deploy',
                description: 'Zero dependencies with Docker-ready Alpine image',
                color: 'blue',
                size: 'wide'
              },
            ].map((advantage, index) => {
              const sizeClass = advantage.size === 'large' ? 'bento-item-large' :
                advantage.size === 'wide' ? 'bento-item-wide' : '';
              return (
                <div
                  key={index}
                  className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-advantages'] ? 'visible' : ''} ${sizeClass}`}
                  style={{
                    padding: advantage.size === 'large' ? '2.5rem' : '2rem',
                    animationDelay: `${0.2 + index * 0.1}s`,
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <div
                    className={`feature-icon-box ${advantage.color}`}
                    style={{ marginBottom: '1.5rem' }}
                  >
                    {advantage.icon}
                  </div>
                  <h3 style={{
                    marginBottom: '0.75rem',
                    color: 'hsl(var(--foreground))',
                    fontSize: advantage.size === 'large' ? '1.5rem' : '1.25rem',
                    fontWeight: '700',
                  }}>
                    {advantage.title}
                  </h3>
                  <p style={{
                    color: 'hsl(var(--muted-foreground))',
                    marginBottom: 0,
                    lineHeight: 1.7,
                    flex: 1,
                    fontSize: advantage.size === 'large' ? '1.05rem' : '0.95rem',
                  }}>
                    {advantage.description}
                  </p>
                </div>
              )
            })}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section
        id="section-cta"
        style={{ padding: '3rem 0 5rem' }}
      >
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
            <h2 style={{
              color: 'hsl(var(--foreground))',
              marginBottom: '1rem',
              fontSize: '2rem',
            }}>
              Ready to Try GoPdfSuit?
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              marginBottom: '2rem',
              fontSize: '1.1rem',
              lineHeight: '1.7',
            }}>
              Experience the power of fast, free, and flexible PDF generation today
            </p>
            <div style={{
              display: 'flex',
              gap: '1rem',
              justifyContent: 'center',
              flexWrap: 'wrap',
            }}>
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

      {/* Animation styles */}
      <style>
        {`
          @keyframes fadeInUp {
            from {
              opacity: 0;
              transform: translate3d(0, 40px, 0);
            }
            to {
              opacity: 1;
              transform: translate3d(0, 0, 0);
            }
          }
          
          @keyframes fadeInScale {
            from {
              opacity: 0;
              transform: scale(0.9);
            }
            to {
              opacity: 1;
              transform: scale(1);
            }
          }
          
          .animate-fadeInUp {
            animation: fadeInUp 0.8s ease-out forwards;
          }
          
          .animate-fadeInScale {
            animation: fadeInScale 0.6s ease-out forwards;
          }
          
          .stagger-animation {
            opacity: 0;
          }
          
          .stagger-animation.visible {
            opacity: 1;
          }
          
          .comparison-cards-grid {
            display: grid;
            grid-template-columns: 1.5fr 1fr 1fr 1fr 1fr;
            gap: 1.5rem;
          }
          
          .comparison-cards-grid .glass-card {
            word-wrap: break-word;
            overflow-wrap: break-word;
            word-break: break-word;
            hyphens: auto;
          }
          
          .comparison-cards-grid .feature-value {
            white-space: normal;
            word-wrap: break-word;
            overflow-wrap: break-word;
            word-break: break-word;
          }
          
          @media (max-width: 1600px) {
            .comparison-cards-grid {
              grid-template-columns: repeat(5, 1fr);
              gap: 1rem;
            }
          }
          
          @media (max-width: 1400px) {
            .comparison-cards-grid {
              grid-template-columns: repeat(3, 1fr);
            }
          }
          
          @media (max-width: 900px) {
            .comparison-cards-grid {
              grid-template-columns: repeat(2, 1fr);
            }
          }
          
          @media (max-width: 600px) {
            .comparison-cards-grid {
              grid-template-columns: 1fr;
            }
          }
        `}
      </style>
    </div>
  )
}

export default Comparison