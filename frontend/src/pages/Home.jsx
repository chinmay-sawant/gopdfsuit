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
  Github
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
      link: '/viewer'
    },
    {
      icon: <Edit size={32} />,
      title: 'Visual PDF Editor',
      description: 'Drag-and-drop interface for building PDF templates with live preview and real-time JSON generation.',
      link: '/editor'
    },
    {
      icon: <Merge size={32} />,
      title: 'PDF Merge',
      description: 'Combine multiple PDF files with intuitive drag-and-drop reordering and live preview.',
      link: '/merge'
    },
    {
      icon: <FileCheck size={32} />,
      title: 'Form Filling',
      description: 'AcroForm and XFDF support for filling PDF forms programmatically.',
      link: '/filler'
    },
    {
      icon: <Globe size={32} />,
      title: 'HTML to PDF',
      description: 'Convert HTML content or web pages to PDF using Chromium with full control over page settings.',
      link: '/htmltopdf'
    },
    {
      icon: <Image size={32} />,
      title: 'HTML to Image',
      description: 'Convert HTML content to PNG, JPG, or SVG images with custom dimensions and quality settings.',
      link: '/htmltoimage'
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
        style={{ 
          padding: '4rem 0',
          textAlign: 'center',
        }}
      >
        <div className="container">
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '2rem',
          }}>
            <div 
              className="animate-bounce"
              style={{ 
                fontSize: '4rem',
                animationDelay: '0.5s',
              }}
            >
              📄
            </div>
            <h1 
              className="animate-fadeInUp"
              style={{ 
                fontSize: '3.5rem',
                fontWeight: '800',
                color: 'hsl(var(--foreground))',
                marginBottom: 0,
                animationDelay: '0.2s',
              }}
            >
              GoPdfSuit
            </h1>
          </div>
          
          <div 
            className="animate-fadeInUp"
            style={{ 
              fontSize: '1.5rem',
              marginBottom: '2rem',
              color: 'hsl(var(--muted-foreground))',
              maxWidth: '800px',
              margin: '0 auto 2rem',
              animationDelay: '0.4s',
              minHeight: '3rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <span style={{ position: 'relative' }}>
              {typewriterText}
              <span 
                style={{
                  opacity: typewriterText.length < fullText.length ? 1 : 0,
                  animation: 'blink 1s infinite',
                  marginLeft: '2px',
                }}
              >
                |
              </span>
            </span>
          </div>

          <div 
            className="animate-fadeInUp"
            style={{
              display: 'flex',
              gap: '1rem',
              justifyContent: 'center',
              flexWrap: 'wrap',
              marginBottom: '3rem',
              animationDelay: '0.6s',
            }}
          >
            <Link 
              to="/viewer" 
              className="btn btn btn-hover animate-pulse" 
              style={{ 
                fontSize: '1.1rem', 
                padding: '1rem 2rem',
                animationDelay: '2s',
              }}
            >
              <FileText size={20} />
              Try PDF Generator
            </Link>
            <a 
              href="https://github.com/chinmay-sawant/gopdfsuit" 
              target="_blank" 
              rel="noopener noreferrer"
              className="btn btn-secondary btn-hover"
              style={{ fontSize: '1.1rem', padding: '1rem 2rem' }}
            >
              <Github size={20} />
              View on GitHub
            </a>
          </div>

          {/* Quick Stats */}
          <div className="grid grid-3" style={{ marginTop: '3rem' }}>
            {highlights.map((highlight, index) => (
              <div 
                key={index} 
                className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-hero'] ? 'visible' : ''}`}
                style={{ 
                  textAlign: 'center', 
                  padding: '1.5rem',
                  animationDelay: `${0.8 + index * 0.2}s`,
                }}
              >
                <div style={{ 
                  color: '#4ecdc4', 
                  marginBottom: '1rem', 
                  display: 'flex', 
                  justifyContent: 'center',
                  transition: 'transform 0.3s ease',
                }}>
                  {React.cloneElement(highlight.icon, { 
                    size: 24,
                    style: { transition: 'transform 0.3s ease' }
                  })}
                </div>
                <h3 style={{ marginBottom: '0.5rem', fontSize: '1.2rem' }}>{highlight.title}</h3>
                <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0, fontSize: '0.9rem' }}>
                  {highlight.desc}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section 
        id="section-features"
        style={{ padding: '4rem 0' }}
      >
        <div className="container">
          <h2 
            className={`text-center mb-8 animate-fadeInUp stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
            style={{ 
              color: 'hsl(var(--foreground))',
              animationDelay: '0.2s',
            }}
          >
            ✨ Features
          </h2>
          
          <div className="grid grid-2">
            {features.map((feature, index) => (
              <Link 
                key={index}
                to={feature.link}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                <div 
                  className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
                  style={{ 
                    height: '100%',
                    cursor: 'pointer',
                    animationDelay: `${0.4 + index * 0.1}s`,
                  }}
                >
                  <div style={{ 
                    color: '#4ecdc4', 
                    marginBottom: '1rem',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '1rem',
                    transition: 'all 0.3s ease',
                  }}>
                    <div style={{ 
                      transition: 'transform 0.3s ease',
                    }}>
                      {feature.icon}
                    </div>
                    <h3 style={{ 
                      marginBottom: 0, 
                      color: 'hsl(var(--foreground))',
                      transition: 'color 0.3s ease',
                    }}>
                      {feature.title}
                    </h3>
                  </div>
                  <p style={{ 
                    color: 'hsl(var(--muted-foreground))',
                    marginBottom: 0,
                    lineHeight: 1.6,
                    transition: 'color 0.3s ease',
                  }}>
                    {feature.description}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* Quick Start Section */}
      <section 
        id="section-quickstart"
        style={{ padding: '4rem 0' }}
      >
        <div className="container">
          <h2 
            className={`text-center mb-4 animate-fadeInUp stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
            style={{ 
              color: 'hsl(var(--foreground))',
              animationDelay: '0.2s',
            }}
          >
            ⚡ Quick Start
          </h2>
          
          <div 
            className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
            style={{ 
              maxWidth: '800px', 
              margin: '0 auto',
              animationDelay: '0.4s',
            }}
          >
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🚀 Get Started in 3 Steps</h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {[
                {
                  title: 'Clone & Run',
                  content: (
                    <code style={{ 
                      background: 'rgba(0, 0, 0, 0.3)',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      color: '#4ecdc4',
                      display: 'block',
                      fontSize: '0.9rem',
                    }}>
                      git clone https://github.com/chinmay-sawant/gopdfsuit.git<br/>
                      cd gopdfsuit<br/>
                      go run ./cmd/gopdfsuit
                    </code>
                  )
                },
                {
                  title: 'Server Ready',
                  content: (
                    <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0 }}>
                      🌐 Server listening on: <code style={{ color: '#4ecdc4' }}>http://localhost:8080</code>
                    </p>
                  )
                },
                {
                  title: 'Start Creating',
                  content: (
                    <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0 }}>
                      Navigate to any tool above to start generating PDFs, merging documents, or converting HTML!
                    </p>
                  )
                }
              ].map((step, index) => (
                <div 
                  key={index}
                  className={`animate-slideInLeft stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
                  style={{ 
                    display: 'flex', 
                    alignItems: 'flex-start', 
                    gap: '1rem',
                    animationDelay: `${0.6 + index * 0.2}s`,
                  }}
                >
                  <div style={{ 
                    background: '#007acc',
                    color: 'white',
                    borderRadius: '50%',
                    width: '32px',
                    height: '32px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 'bold',
                    flexShrink: 0,
                    transition: 'all 0.3s ease',
                  }}>
                    {index + 1}
                  </div>
                  <div>
                    <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>{step.title}</h4>
                    {step.content}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* API Overview */}
      <section 
        id="section-api"
        style={{ padding: '4rem 0' }}
      >
        <div className="container">
          <h2 
            className={`text-center mb-4 animate-fadeInUp stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
            style={{ 
              color: 'hsl(var(--foreground))',
              animationDelay: '0.2s',
            }}
          >
            📡 API Endpoints
          </h2>
          
          <div className="grid grid-2">
            <div 
              className={`card card-hover animate-slideInLeft stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
              style={{ animationDelay: '0.4s' }}
            >
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🔧 REST API</h3>
              <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem' }}>
                {[
                  'POST /api/v1/generate/template-pdf - Generate PDF from JSON template',
                  'POST /api/v1/merge - Merge multiple PDF files',
                  'POST /api/v1/fill - Fill PDF forms with XFDF data',
                  'POST /api/v1/htmltopdf - Convert HTML to PDF',
                  'POST /api/v1/htmltoimage - Convert HTML to Image'
                ].map((api, index) => (
                  <div 
                    key={index}
                    className={`animate-fadeInUp stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
                    style={{ 
                      marginBottom: index < 4 ? '0.5rem' : '0',
                      animationDelay: `${0.6 + index * 0.1}s`,
                    }}
                  >
                    <code style={{ color: '#4ecdc4' }}>{api.split(' - ')[0]}</code> - {api.split(' - ')[1]}
                  </div>
                ))}
              </div>
            </div>

            <div 
              className={`card card-hover animate-slideInRight stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
              style={{ animationDelay: '0.4s' }}
            >
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🌐 Web Interfaces</h3>
              <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem' }}>
                {[
                  'GET / - PDF Viewer & Template Processor',
                  'GET /editor - Drag-and-drop Template Editor',
                  'GET /merge - PDF Merge Interface',
                  'GET /filler - PDF Form Filler',
                  'GET /htmltopdf - HTML to PDF Converter'
                ].map((api, index) => (
                  <div 
                    key={index}
                    className={`animate-fadeInUp stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
                    style={{ 
                      marginBottom: index < 4 ? '0.5rem' : '0',
                      animationDelay: `${0.6 + index * 0.1}s`,
                    }}
                  >
                    <code style={{ color: '#4ecdc4' }}>{api.split(' - ')[0]}</code> - {api.split(' - ')[1]}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

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
              [GIN] 2025/09/16 - 01:25:53 | 200 |       417.4µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:56 | 200 |       505.1µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:57 | 200 |      1.1047ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:57 | 200 |       515.1µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:58 | 200 |      2.0475ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:58 | 200 |       850.4µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:59 | 200 |       503.6µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:59 | 200 |       503.8µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:59 | 200 |       681.8µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:25:59 | 200 |      1.0021ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:26:10 | 200 |       504.3µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:26:10 | 200 |       504.5µs |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
              [GIN] 2025/09/16 - 01:26:10 | 200 |      1.5052ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br/>
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
        style={{ padding: '4rem 0' }}
      >
        <div className="container">
          <h2 
            className={`text-center mb-4 animate-fadeInUp stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{ 
              color: 'hsl(var(--foreground))',
              animationDelay: '0.2s',
            }}
          >
            🆚 How We Compare
          </h2>
          
          <div 
            className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{ 
              maxWidth: '900px', 
              margin: '0 auto',
              padding: '2rem',
              textAlign: 'center',
              animationDelay: '0.4s',
            }}
          >
            <p style={{ 
              color: 'hsl(var(--muted-foreground))', 
              fontSize: '1.1rem',
              marginBottom: '2rem',
            }}>
              See how GoPdfSuit stacks up against industry-leading PDF libraries and commercial solutions
            </p>
            
            <div style={{ 
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
              gap: '1.5rem',
              marginBottom: '2rem',
            }}>
              <div style={{ 
                background: 'rgba(78, 205, 196, 0.1)',
                padding: '1.5rem', 
                borderRadius: '8px',
                border: '1px solid rgba(78, 205, 196, 0.3)',
              }}>
                <div style={{ fontSize: '1.8rem', fontWeight: 'bold', color: '#4ecdc4', marginBottom: '0.5rem' }}>
                  Free
                </div>
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>
                  vs $2K-4K/dev/year
                </div>
              </div>
              
              <div style={{ 
                background: 'rgba(0, 122, 204, 0.1)',
                padding: '1.5rem', 
                borderRadius: '8px',
                border: '1px solid rgba(0, 122, 204, 0.3)',
              }}>
                <div style={{ fontSize: '1.8rem', fontWeight: 'bold', color: '#007acc', marginBottom: '0.5rem' }}>
                  179µs-1.7ms
                </div>
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>
                  Ultra-fast response
                </div>
              </div>
              
              <div style={{ 
                background: 'rgba(255, 193, 7, 0.1)',
                padding: '1.5rem', 
                borderRadius: '8px',
                border: '1px solid rgba(255, 193, 7, 0.3)',
              }}>
                <div style={{ fontSize: '1.8rem', fontWeight: 'bold', color: '#ffc107', marginBottom: '0.5rem' }}>
                  REST API
                </div>
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>
                  Language agnostic
                </div>
              </div>
            </div>
            
            <Link 
              to="/comparison"
              className="btn btn-primary btn-hover"
              style={{ 
                fontSize: '1.1rem', 
                padding: '1rem 2rem',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}
            >
              📊 View Full Comparison
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer 
        id="section-footer"
        style={{ 
          padding: '2rem 0',
          borderTop: '1px solid rgba(255, 255, 255, 0.2)',
          marginTop: '2rem',
        }}
      >
        <div className="container">
          <div 
            className={`animate-fadeInUp stagger-animation ${isVisible['section-footer'] ? 'visible' : ''}`}
            style={{ 
              textAlign: 'center',
              color: 'hsl(var(--muted-foreground))',
              animationDelay: '0.2s',
            }}
          >
            <p 
              className={`animate-fadeInUp stagger-animation ${isVisible['section-footer'] ? 'visible' : ''}`}
              style={{ 
                marginBottom: '1rem',
                animationDelay: '0.4s',
              }}
            >
              Made with ❤️ and ☕ by{' '}
              <a 
                href="https://github.com/chinmay-sawant" 
                target="_blank" 
                rel="noopener noreferrer"
                style={{ 
                  color: '#4ecdc4', 
                  textDecoration: 'none',
                  transition: 'all 0.3s ease',
                }}
                onMouseEnter={(e) => {
                  e.target.style.transform = 'scale(1.1)'
                  e.target.style.textShadow = '0 0 10px rgba(78, 205, 196, 0.5)'
                }}
                onMouseLeave={(e) => {
                  e.target.style.transform = 'scale(1)'
                  e.target.style.textShadow = 'none'
                }}
              >
                Chinmay Sawant
              </a>
            </p>
            <p 
              className={`animate-fadeInUp stagger-animation ${isVisible['section-footer'] ? 'visible' : ''}`}
              style={{ 
                marginBottom: 0, 
                fontSize: '0.9rem',
                animationDelay: '0.6s',
              }}
            >
              <span 
                className="animate-pulse"
                style={{ animationDelay: '3s' }}
              >
                ⭐
              </span>{' '}
              Star this repo if you find it helpful!
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default Home