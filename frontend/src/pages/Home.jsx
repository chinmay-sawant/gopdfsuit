import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  FileText,
  Edit,
  Merge,
  FileCheck,
  Globe,
  Image,
  Zap,
  Shield,
  Download,
  CheckCircle,
  Star,
  Github,
  ChevronDown,
  ArrowRight,
  Sparkles
} from 'lucide-react'

const Home = () => {
  const [isVisible, setIsVisible] = useState({})
  const [typewriterText, setTypewriterText] = useState('')

  const fullText = "  A powerful Go web service that generates template-based PDF documents on-the-fly with multi-page support, PDF merge capabilities, and HTML to PDF/Image conversion."

  // Typewriter effect
  useEffect(() => {
    let i = 0
    const timer = setInterval(() => {
      if (i < fullText.length) {
        setTypewriterText(prev => prev + fullText.charAt(i))
        i++
      } else {
        clearInterval(timer)
      }
    }, 30)

    return () => clearInterval(timer)
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
      icon: <FileText size={32} />,
      title: 'Template-based PDF Generation',
      description: 'JSON-driven PDF creation with multi-page support, automatic page breaks, and custom layouts.',
      link: '/viewer',
      color: 'teal',
      size: 'large'
    },
    {
      icon: <Edit size={32} />,
      title: 'Visual PDF Editor',
      description: 'Drag-and-drop interface for building PDF templates with live preview and real-time JSON generation.',
      link: '/editor',
      color: 'blue',
      size: 'normal'
    },
    {
      icon: <Merge size={32} />,
      title: 'PDF Merge',
      description: 'Combine multiple PDF files with intuitive drag-and-drop reordering and live preview.',
      link: '/merge',
      color: 'purple',
      size: 'normal'
    },
    {
      icon: <FileCheck size={32} />,
      title: 'Form Filling',
      description: 'AcroForm and XFDF support for filling PDF forms programmatically.',
      link: '/filler',
      color: 'yellow',
      size: 'normal'
    },
    {
      icon: <Globe size={32} />,
      title: 'HTML to PDF',
      description: 'Convert HTML content or web pages to PDF using Chromium with full control over page settings.',
      link: '/htmltopdf',
      color: 'green',
      size: 'normal'
    },
    {
      icon: <Image size={32} />,
      title: 'HTML to Image',
      description: 'Convert HTML content to PNG, JPG, or SVG images with custom dimensions and quality settings.',
      link: '/htmltoimage',
      color: 'blue',
      size: 'wide'
    }
  ]

  const highlights = [
    { icon: <Zap />, title: 'Ultra Fast', desc: 'Average 0.8ms response time for PDF generation' },
    { icon: <Shield />, title: 'Secure', desc: 'Path traversal protection and input validation' },
    { icon: <Download />, title: 'Self-contained', desc: 'Single binary deployment with zero dependencies' },
  ]

  // Animated background particles
  const BackgroundAnimation = () => {
    return (
      <div style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        pointerEvents: 'none',
        zIndex: -1,
        overflow: 'hidden',
      }}>
        {[...Array(15)].map((_, i) => (
          <div
            key={i}
            style={{
              position: 'absolute',
              width: Math.random() * 6 + 4 + 'px',
              height: Math.random() * 6 + 4 + 'px',
              backgroundColor: `rgba(78, 205, 196, ${Math.random() * 0.3 + 0.1})`,
              borderRadius: '50%',
              left: Math.random() * 100 + '%',
              animation: `float-${i % 3} ${Math.random() * 10 + 15}s infinite linear`,
              animationDelay: Math.random() * 10 + 's',
            }}
          />
        ))}

        {/* Geometric shapes */}
        {[...Array(8)].map((_, i) => (
          <div
            key={`geo-${i}`}
            style={{
              position: 'absolute',
              width: Math.random() * 30 + 20 + 'px',
              height: Math.random() * 30 + 20 + 'px',
              border: `1px solid rgba(0, 122, 204, ${Math.random() * 0.2 + 0.1})`,
              left: Math.random() * 100 + '%',
              top: Math.random() * 100 + '%',
              animation: `rotate-float-${i % 2} ${Math.random() * 20 + 20}s infinite linear`,
              animationDelay: Math.random() * 5 + 's',
            }}
          />
        ))}

        <style>
          {`
            @keyframes float-0 {
              0% { transform: translateY(100vh) rotate(0deg); opacity: 0; }
              10% { opacity: 1; }
              90% { opacity: 1; }
              100% { transform: translateY(-100px) rotate(360deg); opacity: 0; }
            }
            
            @keyframes float-1 {
              0% { transform: translateY(100vh) translateX(0px); opacity: 0; }
              10% { opacity: 1; }
              50% { transform: translateY(50vh) translateX(50px); }
              90% { opacity: 1; }
              100% { transform: translateY(-100px) translateX(0px); opacity: 0; }
            }
            
            @keyframes float-2 {
              0% { transform: translateY(100vh) translateX(0px) scale(0.5); opacity: 0; }
              10% { opacity: 1; }
              50% { transform: translateY(50vh) translateX(-30px) scale(1); }
              90% { opacity: 1; }
              100% { transform: translateY(-100px) translateX(0px) scale(0.5); opacity: 0; }
            }
            
            @keyframes rotate-float-0 {
              0% { transform: rotate(0deg) translateY(0px); }
              50% { transform: rotate(180deg) translateY(-20px); }
              100% { transform: rotate(360deg) translateY(0px); }
            }
            
            @keyframes rotate-float-1 {
              0% { transform: rotate(0deg) scale(1) translateX(0px); }
              33% { transform: rotate(120deg) scale(1.1) translateX(10px); }
              66% { transform: rotate(240deg) scale(0.9) translateX(-10px); }
              100% { transform: rotate(360deg) scale(1) translateX(0px); }
            }
            
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
                transform: scale(0.8);
              }
              to {
                opacity: 1;
                transform: scale(1);
              }
            }
            
            @keyframes slideInLeft {
              from {
                opacity: 0;
                transform: translate3d(-100px, 0, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            @keyframes slideInRight {
              from {
                opacity: 0;
                transform: translate3d(100px, 0, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            @keyframes bounce {
              0%, 20%, 53%, 80%, 100% {
                animation-timing-function: cubic-bezier(0.215, 0.610, 0.355, 1.000);
                transform: translate3d(0, 0, 0);
              }
              40%, 43% {
                animation-timing-function: cubic-bezier(0.755, 0.050, 0.855, 0.060);
                transform: translate3d(0, -10px, 0);
              }
              70% {
                animation-timing-function: cubic-bezier(0.755, 0.050, 0.855, 0.060);
                transform: translate3d(0, -5px, 0);
              }
              90% {
                transform: translate3d(0, -2px, 0);
              }
            }
            
            @keyframes pulse {
              0% {
                transform: scale(1);
              }
              50% {
                transform: scale(1.05);
              }
              100% {
                transform: scale(1);
              }
            }
            
            @keyframes blink {
              0%, 50% {
                opacity: 1;
              }
              51%, 100% {
                opacity: 0;
              }
            }
            
            .animate-fadeInUp {
              animation: fadeInUp 0.8s ease-out forwards;
            }
            
            .animate-fadeInScale {
              animation: fadeInScale 0.6s ease-out forwards;
            }
            
            .animate-slideInLeft {
              animation: slideInLeft 0.8s ease-out forwards;
            }
            
            .animate-slideInRight {
              animation: slideInRight 0.8s ease-out forwards;
            }
            
            .animate-bounce {
              animation: bounce 2s infinite;
            }
            
            .animate-pulse {
              animation: pulse 2s infinite;
            }
            
            .card-hover {
              transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
              position: relative;
              overflow: hidden;
            }
            
            .card-hover::before {
              content: '';
              position: absolute;
              top: 0;
              left: -100%;
              width: 100%;
              height: 100%;
              background: linear-gradient(90deg, transparent, rgba(78, 205, 196, 0.1), transparent);
              transition: left 0.6s;
            }
            
            .card-hover:hover::before {
              left: 100%;
            }
            
            .card-hover:hover {
              transform: translateY(-12px) scale(1.03);
              box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
            }
            
            .card-hover:hover svg {
              transform: scale(1.2) rotate(5deg);
            }
            
            .btn-hover {
              transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
              position: relative;
              overflow: hidden;
            }
            
            .btn-hover::before {
              content: '';
              position: absolute;
              top: 50%;
              left: 50%;
              width: 0;
              height: 0;
              background: rgba(255, 255, 255, 0.1);
              border-radius: 50%;
              transform: translate(-50%, -50%);
              transition: width 0.6s, height 0.6s;
            }
            
            .btn-hover:hover::before {
              width: 300px;
              height: 300px;
            }
            
            .btn-hover:hover {
              transform: translateY(-3px) scale(1.05);
              box-shadow: 0 15px 30px rgba(0, 0, 0, 0.3);
            }
            
            .btn-hover:active {
              transform: translateY(-1px) scale(1.02);
            }
            
            .stagger-animation {
              opacity: 0;
            }
            
            .stagger-animation.visible {
              opacity: 1;
            }
            
            /* Custom Scrollbar Styles */
            .custom-scrollbar::-webkit-scrollbar {
              width: 8px;
            }
            
            .custom-scrollbar::-webkit-scrollbar-track {
              background: rgba(0, 0, 0, 0.3);
              border-radius: 4px;
            }
            
            .custom-scrollbar::-webkit-scrollbar-thumb {
              background: rgba(78, 205, 196, 0.5);
              border-radius: 4px;
              transition: background 0.3s ease;
            }
            
            .custom-scrollbar::-webkit-scrollbar-thumb:hover {
              background: rgba(78, 205, 196, 0.8);
            }
            
            .custom-scrollbar::-webkit-scrollbar-corner {
              background: rgba(0, 0, 0, 0.3);
            }
          `}
        </style>
      </div>
    )
  }

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

          {/* Typewriter subtitle */}
          <div
            className="hero-subtitle animate-fadeInUp"
            style={{
              marginBottom: '3rem',
              color: 'hsl(var(--muted-foreground))',
              animationDelay: '0.2s',
              minHeight: '4rem',
            }}
          >
            <span style={{ position: 'relative' }}>
              {typewriterText}
              <span
                style={{
                  opacity: typewriterText.length < fullText.length ? 1 : 0,
                  animation: 'blink 1s infinite',
                  marginLeft: '2px',
                  color: '#4ecdc4',
                }}
              >
                |
              </span>
            </span>
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
              to="/viewer"
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
              <Star size={16} />
            </a>
          </div>

          {/* Quick Stats - Glass Cards */}
          <div
            className="grid grid-3"
            style={{ marginTop: '2rem' }}
          >
            {highlights.map((highlight, index) => (
              <div
                key={index}
                className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-hero'] ? 'visible' : ''}`}
                style={{
                  textAlign: 'center',
                  padding: '2rem 1.5rem',
                  animationDelay: `${0.4 + index * 0.15}s`,
                }}
              >
                <div
                  className={`feature-icon-box ${index === 0 ? 'teal' : index === 1 ? 'blue' : 'purple'}`}
                  style={{
                    margin: '0 auto 1rem',
                  }}
                >
                  {React.cloneElement(highlight.icon, {
                    size: 28,
                  })}
                </div>
                <h3 style={{
                  marginBottom: '0.5rem',
                  fontSize: '1.3rem',
                  fontWeight: '700',
                }}>
                  {highlight.title}
                </h3>
                <p style={{
                  color: 'hsl(var(--muted-foreground))',
                  marginBottom: 0,
                  fontSize: '0.95rem',
                  lineHeight: '1.6',
                }}>
                  {highlight.desc}
                </p>
              </div>
            ))}
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

          <div className="bento-grid">
            {features.map((feature, index) => {
              const sizeClass = feature.size === 'large' ? 'bento-item-large' :
                feature.size === 'wide' ? 'bento-item-wide' : '';
              return (
                <Link
                  key={index}
                  to={feature.link}
                  style={{ textDecoration: 'none', color: 'inherit' }}
                  className={sizeClass}
                >
                  <div
                    className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
                    style={{
                      height: '100%',
                      padding: feature.size === 'large' ? '2.5rem' : '2rem',
                      cursor: 'pointer',
                      animationDelay: `${0.2 + index * 0.1}s`,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      className={`feature-icon-box ${feature.color}`}
                      style={{ marginBottom: '1.5rem' }}
                    >
                      {feature.icon}
                    </div>
                    <h3 style={{
                      marginBottom: '0.75rem',
                      color: 'hsl(var(--foreground))',
                      fontSize: feature.size === 'large' ? '1.5rem' : '1.25rem',
                      fontWeight: '700',
                    }}>
                      {feature.title}
                    </h3>
                    <p style={{
                      color: 'hsl(var(--muted-foreground))',
                      marginBottom: '1rem',
                      lineHeight: 1.7,
                      flex: 1,
                      fontSize: feature.size === 'large' ? '1.05rem' : '0.95rem',
                    }}>
                      {feature.description}
                    </p>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      color: '#4ecdc4',
                      fontSize: '0.9rem',
                      fontWeight: '600',
                    }}>
                      Try it now
                      <ArrowRight size={16} />
                    </div>
                  </div>
                </Link>
              );
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
                  <span className="terminal-command">go run ./cmd/gopdfsuit</span>
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
                  { method: 'POST', path: '/api/v1/fill', desc: 'Fill forms' },
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
                      borderBottom: index < 4 ? '1px solid rgba(255,255,255,0.05)' : 'none',
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
                  { path: '/', desc: 'PDF Viewer & Template Processor' },
                  { path: '/editor', desc: 'Drag-and-drop Template Editor' },
                  { path: '/merge', desc: 'PDF Merge Interface' },
                  { path: '/filler', desc: 'PDF Form Filler' },
                  { path: '/htmltopdf', desc: 'HTML to PDF Converter' }
                ].map((route, index) => (
                  <div
                    key={index}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.75rem',
                      padding: '0.5rem 0',
                      borderBottom: index < 4 ? '1px solid rgba(255,255,255,0.05)' : 'none',
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
            <h2
              className={`animate-fadeInUp stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
              style={{
                color: 'hsl(var(--foreground))',
                marginBottom: '1rem',
                animationDelay: '0.4s',
              }}
            >
              🏃‍♂️ Performance
            </h2>
            <p
              className={`animate-fadeInUp stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
              style={{
                color: 'hsl(var(--muted-foreground))',
                marginBottom: '2rem',
                animationDelay: '0.6s',
              }}
            >
              Ultra-fast PDF generation with in-memory processing
            </p>

            {/* Performance Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
              {[
                { value: '806 µs', label: 'Average Response', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
                { value: '417 µs', label: 'Min Response', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
                { value: '2.05 ms', label: 'Max Response', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' }
              ].map((stat, index) => (
                <div
                  key={index}
                  className={`animate-fadeInScale stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
                  style={{
                    background: stat.bg,
                    padding: '1rem',
                    borderRadius: '8px',
                    border: `1px solid ${stat.border}`,
                    transition: 'all 0.3s ease',
                    animationDelay: `${0.8 + index * 0.2}s`,
                  }}
                >
                  <div
                    className="animate-pulse"
                    style={{
                      fontSize: '1.5rem',
                      fontWeight: 'bold',
                      color: stat.color,
                      animationDelay: `${2 + index * 0.5}s`,
                    }}
                  >
                    {stat.value}
                  </div>
                  <div style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))' }}>
                    {stat.label}
                  </div>
                </div>
              ))}
            </div>

            {/* Sample Logs */}
            <div style={{
              background: 'rgba(0, 0, 0, 0.3)',
              padding: '1rem',
              borderRadius: '8px',
              fontFamily: 'monospace',
              color: '#4ecdc4',
              fontSize: '0.8rem',
              textAlign: 'left',
              maxHeight: '200px',
              overflowY: 'auto',
              scrollbarWidth: 'thin',
              scrollbarColor: 'rgba(78, 205, 196, 0.5) rgba(0, 0, 0, 0.3)',
            }}
              className="custom-scrollbar"
            >
              <div style={{ marginBottom: '0.5rem', fontWeight: 'bold' }}>Recent Performance Logs:</div>
              [GIN] 2025/09/16 - 01:25:53 | 200 |       417.4µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:56 | 200 |       505.1µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:57 | 200 |      1.1047ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:57 | 200 |       515.1µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:58 | 200 |      2.0475ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:58 | 200 |       850.4µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:59 | 200 |       503.6µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:59 | 200 |       503.8µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:59 | 200 |       681.8µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:25:59 | 200 |      1.0021ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:26:10 | 200 |       504.3µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:26:10 | 200 |       504.5µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:26:10 | 200 |      1.5052ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2025/09/16 - 01:26:10 | 200 |         652µs |             ::1 | POST     "/api/v1/generate/template-pdf"
            </div>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              marginTop: '1rem',
              fontSize: '0.9rem',
              marginBottom: 0,
            }}>
              Performance benchmarks for multi-page PDF generation (14 samples - temp_multiplepage.json)
            </p>
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
              maxWidth: '600px',
              margin: '0 auto',
            }}>
              See how we compare to other PDF solutions
            </p>
          </div>

          <div
            className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{
              maxWidth: '1000px',
              margin: '0 auto',
              padding: '3rem',
            }}
          >
            <div
              className="grid grid-3"
              style={{ marginBottom: '2.5rem' }}
            >
              {[
                { value: 'Free', label: 'vs $2K-4K/dev/year', color: '#4ecdc4', icon: <CheckCircle size={24} /> },
                { value: '< 1ms', label: 'Ultra-fast response', color: '#007acc', icon: <Zap size={24} /> },
                { value: '100% Go', label: 'Single binary deploy', color: '#f093fb', icon: <Download size={24} /> }
              ].map((stat, index) => (
                <div
                  key={index}
                  className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
                  style={{
                    textAlign: 'center',
                    padding: '2rem',
                    animationDelay: `${0.2 + index * 0.1}s`,
                  }}
                >
                  <div style={{
                    color: stat.color,
                    marginBottom: '1rem',
                    display: 'flex',
                    justifyContent: 'center',
                  }}>
                    {stat.icon}
                  </div>
                  <div className="stat-value" style={{ color: stat.color, marginBottom: '0.5rem' }}>
                    {stat.value}
                  </div>
                  <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>
                    {stat.label}
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
              <Link
                to="/viewer"
                className="btn-outline-glow"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  textDecoration: 'none',
                  padding: '0.75rem 1.5rem',
                }}
              >
                <FileText size={18} />
                Documentation
              </Link>
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