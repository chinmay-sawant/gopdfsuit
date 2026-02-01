import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  FileText,
  Edit,
  Merge,
  FileCheck,
  Globe,
  Image,
  Zap,
  Download,
  CheckCircle,
  Star,
  Github,
  ChevronDown,
  ArrowRight,
  Sparkles
} from 'lucide-react'
import PerformanceSection from '../components/PerformanceSection'
import BackgroundAnimation from '../components/BackgroundAnimation'

const Home = () => {
  const [isVisible, setIsVisible] = useState({})
  const [starCount, setStarCount] = useState(null)



  // Fetch GitHub stars
  useEffect(() => {
    fetch('https://api.github.com/repos/chinmay-sawant/gopdfsuit')
      .then(res => res.json())
      .then(data => {
        if (data.stargazers_count !== undefined) {
          setStarCount(data.stargazers_count)
        }
      })
      .catch(err => console.error('Error fetching stars:', err))
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
  const features = [
    {
      icon: <FileText size={24} />,
      title: 'Go Support',
      description: 'Use as a standalone library (gopdflib) or via HTTP API.',
      link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/feature/python-binary/pkg/gopdflib',
      color: 'blue',
      external: true
    },
    {
      icon: <Globe size={24} />,
      title: 'Python Web Client',
      description: 'Lightweight API client for interacting with the GoPdfSuit server.',
      link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/feature/python-binary/sampledata/python/gopdf',
      color: 'teal',
      external: true
    },
    {
      icon: <Zap size={24} />,
      title: 'Native Python Support',
      description: 'High-performance CGO bindings for direct PDF generation from Python.',
      link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/feature/python-binary/bindings/python',
      color: 'yellow',
      external: true
    },
    {
      icon: <Globe size={24} />,
      title: 'Language Agnostic',
      description: 'REST API works with any programming language.',
      link: '#section-api',
      color: 'purple',
      external: false
    },
    {
      icon: <FileText size={24} />,
      title: 'Template-based PDF',
      description: 'JSON-driven PDF creation with multi-page support and automatic page breaks.',
      link: '/viewer',
      color: 'teal'
    },
    {
      icon: <Edit size={24} />,
      title: 'Visual PDF Editor',
      description: 'Drag-and-drop interface for building PDF templates with live preview.',
      link: '/editor',
      color: 'blue'
    },
    {
      icon: <Merge size={24} />,
      title: 'PDF Merge',
      description: 'Combine multiple PDFs with drag-and-drop reordering and live preview.',
      link: '/merge',
      color: 'purple'
    },
    {
      icon: <FileCheck size={24} />,
      title: 'Form Filling',
      description: 'AcroForm and XFDF support for filling PDF forms programmatically.',
      link: '/filler',
      color: 'yellow'
    },
    {
      icon: <Globe size={24} />,
      title: 'HTML to PDF',
      description: 'Convert HTML content or web pages to PDF using Chromium.',
      link: '/htmltopdf',
      color: 'green'
    },
    {
      icon: <Image size={24} />,
      title: 'HTML to Image',
      description: 'Convert HTML to PNG, JPG, or SVG with custom dimensions.',
      link: '/htmltoimage',
      color: 'blue'
    }
  ]


  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />

      {/* Hero Section */}
      <section
        id="section-hero"
        className="hero-section"
        style={{
          padding: '6rem 0 4rem',
          textAlign: 'center',
        }}
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
              marginBottom: '2rem',
              color: '#4ecdc4',
              fontSize: '0.9rem',
              fontWeight: '500',
            }}
          >
            <Sparkles size={16} />
            Open Source PDF Generation Engine
          </div>

          {/* Main Title */}
          <h1
            className="hero-title gradient-text animate-fadeInUp"
            style={{
              animationDelay: '0.1s',
            }}
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
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.25rem',
              lineHeight: '1.8',
              marginBottom: '2rem',
              fontWeight: '400',
              maxWidth: '850px',
              marginLeft: 'auto',
              marginRight: 'auto'
            }}>
              An exceptional, <span style={{ color: 'hsl(var(--foreground))', fontWeight: '600' }}>MIT-licensed</span> Go engine that <span style={{ color: '#4ecdc4', fontWeight: '600' }}>saves enterprise costs</span> and solves critical <span style={{ color: 'hsl(var(--foreground))', fontWeight: '600' }}>compliance challenges</span> for Fintechs & Enterprises by generating secure, <span style={{ color: '#007acc', fontWeight: '600' }}>PDF/UA-2 & PDF/A-4</span> compliant documents in <span style={{ color: '#ffc107', fontWeight: '600' }}>under 10ms*</span>.
            </p>

            <div style={{
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'center',
              gap: '0.8rem',
            }}>
              {[
                "PDF/A-4 & PDF/UA-2 Compliant",
                "AES-128 Encryption",
                "Multi-page Support", "Split PDFs",
                "HTML To Image", "HTML To PDF",
                "Private", "In-Memory", "Native Python Support",
                "Send Data via API", "Docker Support"
              ].map((feature, i) => (
                <span key={i} style={{
                  background: 'rgba(78, 205, 196, 0.08)',
                  border: '1px solid rgba(78, 205, 196, 0.2)',
                  color: '#4ecdc4',
                  padding: '0.5rem 1rem',
                  borderRadius: '20px',
                  fontSize: '0.95rem',
                  fontWeight: '500',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  boxShadow: '0 2px 10px rgba(78, 205, 196, 0.05)'
                }}>
                  <CheckCircle size={15} />
                  {feature}
                </span>
              ))}
            </div>
          </div>

          {/* CTA Buttons */}
          <div
            className="animate-fadeInUp"
            style={{
              display: 'flex',
              gap: '1.5rem',
              justifyContent: 'center',
              flexWrap: 'wrap',
              marginBottom: '4rem',
              animationDelay: '0.3s',
            }}
          >
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
              <Github size={20} />
              View on GitHub
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.3rem',
                background: 'rgba(255,255,255,0.1)',
                padding: '0.2rem 0.6rem',
                borderRadius: '20px',
                marginLeft: '0.25rem'
              }}>
                <Star size={14} fill={starCount ? "currentColor" : "none"} />
                <span style={{ fontSize: '0.9rem' }}>{starCount !== null ? starCount : 'Star'}</span>
              </div>
            </a>
          </div>

          {/* Quick Stats - Glass Cards */}
          <div
            className="grid grid-3"
            style={{ marginTop: '2rem' }}
          >

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

      {/* Features Section */}
      <section
        id="section-features"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
            style={{ marginBottom: '3rem' }}
          >
            <h2
              className="gradient-text"
              style={{
                fontSize: '2.5rem',
                marginBottom: '1rem',
              }}
            >
              Powerful Features
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
              maxWidth: '600px',
              margin: '0 auto',
            }}>
              Everything you need for professional PDF workflows
            </p>
          </div>

          <div className="grid grid-3">
            {features.map((feature, index) => {
              const CardContent = () => (
                <div
                  className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
                  style={{
                    height: '100%',
                    padding: '1.5rem',
                    cursor: 'pointer',
                    animationDelay: `${0.1 + index * 0.08}s`,
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
                    <div
                      className={`feature-icon-box ${feature.color}`}
                      style={{ width: '48px', height: '48px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                    >
                      {feature.icon}
                    </div>
                    <h3 style={{
                      marginBottom: 0,
                      color: 'hsl(var(--foreground))',
                      fontSize: '1.1rem',
                      fontWeight: '700',
                    }}>
                      {feature.title}
                    </h3>
                  </div>
                  <p style={{
                    color: 'hsl(var(--muted-foreground))',
                    marginBottom: '0.75rem',
                    lineHeight: 1.6,
                    flex: 1,
                    fontSize: '0.9rem',
                  }}>
                    {feature.description}
                  </p>
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    color: '#4ecdc4',
                    fontSize: '0.85rem',
                    fontWeight: '600',
                  }}>
                    {feature.external ? 'View on GitHub' : 'Try it now'}
                    <ArrowRight size={14} />
                  </div>
                </div>
              )

              if (feature.link.startsWith('#')) {
                return (
                  <div
                    key={index}
                    onClick={() => {
                      const element = document.getElementById(feature.link.substring(1));
                      if (element) element.scrollIntoView({ behavior: 'smooth' });
                    }}
                    style={{ textDecoration: 'none', color: 'inherit' }}
                  >
                    <CardContent />
                  </div>
                )
              }

              return feature.external ? (
                <a
                  key={index}
                  href={feature.link}
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{ textDecoration: 'none', color: 'inherit' }}
                >
                  <CardContent />
                </a>
              ) : (
                <Link
                  key={index}
                  to={feature.link}
                  style={{ textDecoration: 'none', color: 'inherit' }}
                >
                  <CardContent />
                </Link>
              )
            })}
          </div>
        </div>
      </section>

      <div className="section-divider container" />

      {/* Quick Start Section */}
      <section
        id="section-quickstart"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div className="split-layout">
            {/* Left side - Text content */}
            <div
              className={`animate-slideInLeft stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
            >
              <h2
                className="gradient-text"
                style={{
                  fontSize: '2.5rem',
                  marginBottom: '1.5rem',
                }}
              >
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
              className={`terminal-window animate-slideInRight stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
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

      <div className="section-divider container" />

      {/* API Overview */}
      <section
        id="section-api"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
            style={{ marginBottom: '3rem' }}
          >
            <h2
              className="gradient-text"
              style={{ fontSize: '2.5rem', marginBottom: '1rem' }}
            >
              API Endpoints
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
            }}>
              RESTful API for seamless integration
            </p>
          </div>

          <div className="grid grid-2">
            <div
              className={`glass-card animate-slideInLeft stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
              style={{ padding: '2rem' }}
            >
              <div className="feature-icon-box blue" style={{ marginBottom: '1.5rem' }}>
                <Zap size={28} />
              </div>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', fontSize: '1.4rem' }}>REST API</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {[
                  { method: 'POST', path: '/api/v1/generate/template-pdf', desc: 'Generate PDF' },
                  { method: 'POST', path: '/api/v1/merge', desc: 'Merge PDFs' },
                  { method: 'POST', path: '/api/v1/split', desc: 'Split PDFs' },
                  { method: 'POST', path: '/api/v1/fill', desc: 'Fill forms' },
                  { method: 'GET', path: '/api/v1/template-data', desc: 'Get template data' },
                  { method: 'GET', path: '/api/v1/fonts', desc: 'List fonts' },
                  { method: 'POST', path: '/api/v1/htmltopdf', desc: 'HTML to PDF' },
                  { method: 'POST', path: '/api/v1/htmltoimage', desc: 'HTML to Image' }
                ].map((api, index) => (
                  <div
                    key={index}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.75rem',
                      padding: '0.5rem 0',
                      borderBottom: index < 7 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                    }}
                  >
                    <span style={{
                      background: 'rgba(0, 122, 204, 0.2)',
                      color: '#007acc',
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.7rem',
                      fontWeight: '700',
                    }}>
                      {api.method}
                    </span>
                    <code style={{ color: '#4ecdc4', fontSize: '0.85rem', flex: 1 }}>{api.path}</code>
                  </div>
                ))}
              </div>
            </div>

            <div
              className={`glass-card animate-slideInRight stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
              style={{ padding: '2rem', animationDelay: '0.2s' }}
            >
              <div className="feature-icon-box purple" style={{ marginBottom: '1.5rem' }}>
                <Globe size={28} />
              </div>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', fontSize: '1.4rem' }}>Web Interfaces</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {[
                  { path: '/viewer', desc: 'PDF Viewer & Template Processor' },
                  { path: '/editor', desc: 'Drag-and-drop Template Editor' },
                  { path: '/merge', desc: 'PDF Merge Interface' },
                  { path: '/split', desc: 'PDF Split Interface' },
                  { path: '/filler', desc: 'PDF Form Filler' },
                  { path: '/htmltopdf', desc: 'HTML to PDF Converter' },
                  { path: '/htmltoimage', desc: 'HTML to Image Converter' },
                  { path: '/screenshots', desc: 'Screenshots Page' },
                  { path: '/comparison', desc: 'Feature Comparison' }
                ].map((route, index) => (
                  <div
                    key={index}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.75rem',
                      padding: '0.5rem 0',
                      borderBottom: index < 8 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                    }}
                  >
                    <span style={{
                      background: 'rgba(240, 147, 251, 0.2)',
                      color: '#f093fb',
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.7rem',
                      fontWeight: '700',
                    }}>
                      GET
                    </span>
                    <code style={{ color: '#4ecdc4', fontSize: '0.85rem' }}>{route.path}</code>
                    <span style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.85rem' }}>- {route.desc}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      <div className="section-divider container" />

      {/* Performance Section */}
      <section
        id="section-performance"
        style={{ padding: '4rem 0' }}
      >
        <div className="container">
          <div
            className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
              maxWidth: '800px',
              margin: '0 auto',
              animationDelay: '0.2s',
            }}
          >
            <PerformanceSection isVisible={isVisible['section-performance']} />

          </div>
        </div>
      </section>

      {/* Comparison Preview Section */}
      <section
        id="section-comparison-preview"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{ marginBottom: '3rem' }}
          >
            <h2
              className="gradient-text"
              style={{ fontSize: '2.5rem', marginBottom: '1rem' }}
            >
              Why Choose GoPdfSuit?
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
              maxWidth: '700px',
              margin: '0 auto',
            }}>
              Enterprise features at zero cost — compare with iTextPDF, PDFLib, and commercial solutions
            </p>
          </div>

          {/* Quick Stats */}
          <div
            className="grid grid-3"
            style={{ marginBottom: '2.5rem', maxWidth: '900px', margin: '0 auto 2.5rem' }}
          >
            {[
              { value: 'Free', label: 'vs $2K-4K/dev/year', color: '#4ecdc4', icon: <CheckCircle size={28} /> },
              { value: '< 100ms', label: 'Response time', color: '#007acc', icon: <Zap size={28} /> },
              { value: '0 deps', label: 'Pure Go binary', color: '#f093fb', icon: <Download size={28} /> }
            ].map((stat, index) => (
              <div
                key={index}
                className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
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

          {/* Feature Comparison */}
          <div
            className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{ width: '100%', padding: '2.5rem' }}
          >
            <h3 style={{
              color: 'hsl(var(--foreground))',
              marginBottom: '1.5rem',
              fontSize: '1.3rem',
              textAlign: 'center',
            }}>
              Built-in Enterprise Features
            </h3>

            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
              gap: '1rem',
              marginBottom: '2rem',
            }}>
              {[
                { name: 'Go Support (gopdflib)', desc: 'Direct Struct Access & HTTP API', color: '#ffc107' },
                { name: 'Native Python Bindings', desc: 'CGO + ctypes wrapper via pypdfsuit', color: '#4ecdc4' },
                { name: 'Python Web Client', desc: 'Lightweight REST API client', color: '#007acc' },
                { name: 'PDF/A-4 Compliance', desc: 'Archival standard with sRGB ICC profiles', color: '#4ecdc4' },
                { name: 'PDF/UA-2 Accessibility', desc: 'Universal accessibility compliance', color: '#007acc' },
                { name: 'AES-128 Encryption', desc: 'Password protection with permissions', color: '#f093fb' },
                { name: 'Digital Signatures', desc: 'PKCS#7 certificates with visual appearance', color: '#ffc107' },
                { name: 'Font Subsetting', desc: 'TrueType embedding with glyph optimization', color: '#4ecdc4' },
                { name: 'PDF Merge', desc: 'Combine multiple PDFs, preserve forms', color: '#007acc' },
                { name: 'XFDF Form Filling', desc: 'Advanced field detection and population', color: '#f093fb' },
                { name: 'Bookmarks & Links', desc: 'Outlines with internal/external hyperlinks', color: '#ffc107' },
                { name: 'Language Agnostic', desc: 'REST API works with any programming language', color: '#f093fb' },
              ].map((feature, index) => (
                <div
                  key={index}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: '0.75rem',
                    padding: '0.75rem',
                    background: 'rgba(255,255,255,0.02)',
                    borderRadius: '8px',
                    border: '1px solid rgba(255,255,255,0.05)',
                  }}
                >
                  <CheckCircle size={18} style={{ color: feature.color, flexShrink: 0, marginTop: '2px' }} />
                  <div>
                    <div style={{ color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>
                      {feature.name}
                    </div>
                    <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>
                      {feature.desc}
                    </div>
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

      {/* Footer */}
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

          <div
            className={`animate-fadeInUp stagger-animation ${isVisible['section-footer'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
            }}
          >
            {/* Quick Links */}
            <div style={{
              display: 'flex',
              justifyContent: 'center',
              gap: '2rem',
              marginBottom: '2rem',
              flexWrap: 'wrap',
            }}>
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
                  padding: '0.75rem 1.5rem',
                }}
              >
                <Github size={18} />
                GitHub
              </a>
            </div>

            {/* Credits */}
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1rem',
              marginBottom: '0.5rem',
            }}>
              Made with ❤️ and ☕ by{' '}
              <a
                href="https://github.com/chinmay-sawant"
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  color: '#4ecdc4',
                  textDecoration: 'none',
                  fontWeight: '600',
                }}
              >
                Chinmay Sawant
              </a>
            </p>

            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '0.9rem',
              marginBottom: 0,
              opacity: 0.7,
            }}>
              <Star size={14} style={{ display: 'inline', marginRight: '0.5rem', color: '#ffc107' }} />
              Star this repo if you find it helpful!
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default Home