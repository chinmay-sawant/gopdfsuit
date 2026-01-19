import React, { useState, useEffect } from 'react'
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
  ArrowLeft
} from 'lucide-react'
import PerformanceSection from '../components/PerformanceSection'

const Comparison = () => {
  const [isVisible, setIsVisible] = useState({})

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
      pricing: '✅ Free (MIT License)',
      performance: '✅ Ultra Fast (1.9ms - 42ms)',
      deployment: '✅ Microservice/Sidecar/Docker',
      memory: '✅ In-Memory Processing',
      integration: '✅ REST API (Language Agnostic)',
      template: '✅ JSON-based Templates',
      webInterface: '✅ Built-in Viewer/Editor',
      formFilling: '✅ XFDF Advanced Detection',
      pdfMerge: '✅ Drag & Drop + Form Preservation',
      htmlConversion: '✅ gochromedp (Chromium)',
      multipage: '✅ Auto Page Breaks',
      styling: '✅ Font Styles + Borders + Images',
      pdfaCompliance: '✅ PDF/A-4 with ICC Profiles',
      pdfuaCompliance: '✅ PDF/UA-2 Accessibility',
      encryption: '✅ AES-128 with Permissions',
      digitalSignatures: '✅ PKCS#7 + Visual Appearance',
      fontEmbedding: '✅ TrueType Subsetting',
      bookmarks: '✅ Outlines + Hyperlinks',
      dockerSupport: '✅ Multi-stage Alpine Image',
      maintenance: '✅ Single Binary'
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
      formFilling: '✅ Full Support',
      pdfMerge: '✅ Supported',
      htmlConversion: 'Limited',
      multipage: '✅ Manual Control',
      styling: '✅ Code-based',
      pdfaCompliance: '✅ PDF/A',
      pdfuaCompliance: '✅ PDF/UA',
      encryption: '✅ Supported',
      digitalSignatures: '✅ Supported',
      fontEmbedding: '✅ Supported',
      bookmarks: '✅ Supported',
      dockerSupport: '❌ N/A (Library)',
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
      formFilling: '✅ Full Support',
      pdfMerge: '✅ Supported',
      htmlConversion: '✅ Strong Support',
      multipage: '✅ Supported',
      styling: '✅ Comprehensive',
      pdfaCompliance: '✅ PDF/A-1 to A-3',
      pdfuaCompliance: '✅ PDF/UA',
      encryption: '✅ AES-256',
      digitalSignatures: '✅ Supported',
      fontEmbedding: '✅ Supported',
      bookmarks: '✅ Supported',
      dockerSupport: '❌ N/A (Library)',
      maintenance: 'Commercial Support'
    },
    {
      name: 'iText 7',
      pricing: '$3,500/dev/year (AGPL free)',
      performance: 'Moderate',
      deployment: 'Library Integration',
      memory: 'Mixed',
      integration: 'Java/.NET',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: '✅ Full Support',
      pdfMerge: '✅ Programmatic',
      htmlConversion: 'pdfHTML add-on ($)',
      multipage: '✅ Manual Control',
      styling: '✅ Advanced',
      pdfaCompliance: '✅ PDF/A-1 to PDF/A-3',
      pdfuaCompliance: '✅ PDF/UA-1',
      encryption: '✅ AES-256',
      digitalSignatures: '✅ Full PKI Support',
      fontEmbedding: '✅ Full Embedding',
      bookmarks: '✅ Full Support',
      dockerSupport: '❌ N/A (Library)',
      maintenance: 'Library Updates'
    },
    {
      name: 'PDFLib',
      pricing: '$2,750/dev/year',
      performance: 'Fast (C-based)',
      deployment: 'Library Integration',
      memory: 'Streaming',
      integration: 'C/C++/Java/.NET/PHP',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: 'Limited',
      pdfMerge: '✅ pCOS Required',
      htmlConversion: '❌ Not Supported',
      multipage: '✅ Manual Control',
      styling: '✅ Advanced',
      pdfaCompliance: '✅ PDF/A-1 to PDF/A-3',
      pdfuaCompliance: '✅ PDF/UA-1',
      encryption: '✅ AES-256',
      digitalSignatures: '✅ Add-on Available',
      fontEmbedding: '✅ Full Embedding',
      bookmarks: '✅ Full Support',
      dockerSupport: '❌ N/A (Library)',
      maintenance: 'Library Updates'
    },
    {
      name: 'wkhtmltopdf',
      pricing: '✅ Free (LGPL)',
      performance: 'Slow (Process spawn)',
      deployment: 'Binary + WebKit',
      memory: 'High (WebKit)',
      integration: 'Command Line',
      template: 'HTML/CSS',
      webInterface: 'None',
      formFilling: '❌ Not Supported',
      pdfMerge: '❌ Not Supported',
      htmlConversion: '✅ Native (Outdated WebKit)',
      multipage: '✅ CSS Page Breaks',
      styling: '✅ CSS-based',
      pdfaCompliance: '❌ Not Supported',
      pdfuaCompliance: '❌ Not Supported',
      encryption: '❌ Not Supported',
      digitalSignatures: '❌ Not Supported',
      fontEmbedding: '✅ Automatic',
      bookmarks: 'Limited (TOC)',
      dockerSupport: 'Manual Setup',
      maintenance: '❌ Deprecated'
    }
  ]

  const features = [
    { key: 'pricing', label: '💰 Pricing', icon: <DollarSign size={20} /> },
    { key: 'performance', label: '🚀 Performance', icon: <Zap size={20} /> },
    { key: 'deployment', label: '📦 Deployment', icon: <Box size={20} /> },
    { key: 'memory', label: '🧠 Memory Usage', icon: <TrendingUp size={20} /> },
    { key: 'integration', label: '🔧 Integration', icon: <Code size={20} /> },
    { key: 'template', label: '📄 Template Engine', icon: <CheckCircle size={20} /> },
    { key: 'webInterface', label: '🌐 Web Interface', icon: <Globe size={20} /> },
    { key: 'formFilling', label: '📋 Form Filling (XFDF)', icon: <CheckCircle size={20} /> },
    { key: 'pdfMerge', label: '🔗 PDF Merge', icon: <CheckCircle size={20} /> },
    { key: 'htmlConversion', label: '🌐 HTML to PDF', icon: <Globe size={20} /> },
    { key: 'multipage', label: '📱 Multi-page Support', icon: <CheckCircle size={20} /> },
    { key: 'styling', label: '🎨 Styling & Images', icon: <Star size={20} /> },
    { key: 'pdfaCompliance', label: '📜 PDF/A Compliance', icon: <CheckCircle size={20} /> },
    { key: 'pdfuaCompliance', label: '♿ PDF/UA Accessibility', icon: <CheckCircle size={20} /> },
    { key: 'encryption', label: '🔒 Encryption', icon: <CheckCircle size={20} /> },
    { key: 'digitalSignatures', label: '✍️ Digital Signatures', icon: <CheckCircle size={20} /> },
    { key: 'fontEmbedding', label: '🔤 Font Embedding', icon: <CheckCircle size={20} /> },
    { key: 'bookmarks', label: '📑 Bookmarks & Links', icon: <CheckCircle size={20} /> },
    { key: 'dockerSupport', label: '🐳 Docker Support', icon: <Box size={20} /> },
    { key: 'maintenance', label: '🛠️ Maintenance', icon: <CheckCircle size={20} /> }
  ]

  const getValueStyle = (value, isOurs) => {
    const baseStyle = {
      padding: '0.75rem',
      borderRadius: '6px',
      fontSize: '0.9rem',
      fontWeight: isOurs ? '600' : '500',
      transition: 'all 0.3s ease',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      textAlign: 'center',
      minHeight: '60px',
      wordWrap: 'break-word',
      overflowWrap: 'break-word',
      wordBreak: 'break-word',
      whiteSpace: 'normal',
      lineHeight: '1.4',
    }

    if (isOurs) {
      return {
        ...baseStyle,
        background: 'linear-gradient(135deg, rgba(78, 205, 196, 0.15), rgba(0, 122, 204, 0.15))',
        border: '2px solid rgba(78, 205, 196, 0.4)',
        color: '#4ecdc4',
      }
    }

    // Style for competitor values
    if (value.includes('✅')) {
      return {
        ...baseStyle,
        background: 'rgba(34, 197, 94, 0.1)',
        border: '1px solid rgba(34, 197, 94, 0.3)',
        color: 'hsl(var(--foreground))',
      }
    } else if (value.includes('❌') || value === 'None') {
      return {
        ...baseStyle,
        background: 'rgba(239, 68, 68, 0.1)',
        border: '1px solid rgba(239, 68, 68, 0.3)',
        color: 'hsl(var(--muted-foreground))',
      }
    }

    return {
      ...baseStyle,
      background: 'rgba(100, 100, 100, 0.1)',
      border: '1px solid rgba(100, 100, 100, 0.3)',
      color: 'hsl(var(--foreground))',
    }
  }

  const backgroundAnimation = () => {
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
        {[...Array(8)].map((_, i) => (
          <div
            key={i}
            style={{
              position: 'absolute',
              width: Math.random() * 4 + 2 + 'px',
              height: Math.random() * 4 + 2 + 'px',
              backgroundColor: `rgba(78, 205, 196, ${Math.random() * 0.2 + 0.05})`,
              borderRadius: '50%',
              left: Math.random() * 100 + '%',
              animation: `float-${i % 2} ${Math.random() * 15 + 20}s infinite linear`,
              animationDelay: Math.random() * 10 + 's',
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
              50% { transform: translateY(50vh) translateX(30px); }
              90% { opacity: 1; }
              100% { transform: translateY(-100px) translateX(0px); opacity: 0; }
            }
            
            @keyframes fadeInUp {
              from {
                opacity: 0;
                transform: translate3d(0, 30px, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            @keyframes slideInLeft {
              from {
                opacity: 0;
                transform: translate3d(-50px, 0, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            .animate-fadeInUp {
              animation: fadeInUp 0.6s ease-out forwards;
            }
            
            .animate-slideInLeft {
              animation: slideInLeft 0.6s ease-out forwards;
            }
            
            .stagger-animation {
              opacity: 0;
            }
            
            .stagger-animation.visible {
              opacity: 1;
            }
            
            .comparison-card {
              transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            }
            
            .comparison-card:hover {
              transform: translateY(-4px);
              box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
            }
          `}
        </style>
      </div>
    )
  }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      {backgroundAnimation()}

      {/* Header */}
      <section
        id="section-header"
        style={{ padding: '2rem 0 1rem', textAlign: 'center' }}
      >
        <div className="container">
          <Link
            to="/"
            className="btn"
            style={{
              marginBottom: '2rem',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.5rem',
            }}
          >
            <ArrowLeft size={18} />
            Back to Home
          </Link>

          <h1
            className={`animate-fadeInUp stagger-animation ${isVisible['section-header'] ? 'visible' : ''}`}
            style={{
              fontSize: '3rem',
              fontWeight: '800',
              color: 'hsl(var(--foreground))',
              marginBottom: '1rem',
              animationDelay: '0.2s',
            }}
          >
            🆚 Feature Comparison
          </h1>

          <p
            className={`animate-fadeInUp stagger-animation ${isVisible['section-header'] ? 'visible' : ''}`}
            style={{
              fontSize: '1.2rem',
              color: 'hsl(var(--muted-foreground))',
              maxWidth: '800px',
              margin: '0 auto',
              animationDelay: '0.4s',
            }}
          >
            See how GoPdfSuit compares against industry-leading PDF libraries and solutions
          </p>
        </div>
      </section>

      {/* Comparison Table */}
      <section
        id="section-comparison"
        style={{ padding: '2rem 0' }}
      >
        <div className="container">
          <div
            className={`card comparison-card animate-slideInLeft stagger-animation ${isVisible['section-comparison'] ? 'visible' : ''}`}
            style={{
              padding: '2rem',
              overflow: 'auto',
              animationDelay: '0.2s',
              width: '100%',
              margin: '0 auto',
            }}
          >
            <div style={{ width: '100%', minWidth: '100%' }}>
              {/* Header Row */}
              <div style={{
                display: 'grid',
                gridTemplateColumns: '2fr repeat(6, 1fr)',
                gap: '1rem',
                marginBottom: '1rem',
                paddingBottom: '1rem',
                borderBottom: '2px solid rgba(78, 205, 196, 0.3)',
                alignItems: 'center',
              }}>
                <div style={{ fontWeight: 'bold', fontSize: '1.1rem', color: 'hsl(var(--foreground))', display: 'flex', alignItems: 'center', justifyContent: 'flex-start', minHeight: '60px' }}>
                  Feature
                </div>
                {competitors.map((competitor, index) => (
                  <div
                    key={competitor.name}
                    style={{
                      textAlign: 'center',
                      fontWeight: 'bold',
                      fontSize: '1.1rem',
                      color: competitor.isOurs ? '#4ecdc4' : 'hsl(var(--foreground))',
                      padding: '0.75rem',
                      borderRadius: '8px',
                      background: competitor.isOurs ? 'rgba(78, 205, 196, 0.1)' : 'transparent',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      minHeight: '60px',
                      border: competitor.isOurs ? '1px solid rgba(78, 205, 196, 0.3)' : 'none',
                    }}
                  >
                    <div>{competitor.name}</div>
                    {competitor.isOurs && (
                      <div style={{
                        fontSize: '0.8rem',
                        color: '#4ecdc4',
                        marginTop: '0.25rem',
                      }}>
                        (Our Solution)
                      </div>
                    )}
                  </div>
                ))}
              </div>

              {/* Feature Rows */}
              {features.map((feature, featureIndex) => (
                <div
                  key={feature.key}
                  className={`animate-fadeInUp stagger-animation ${isVisible['section-comparison'] ? 'visible' : ''}`}
                  style={{
                    display: 'grid',
                    gridTemplateColumns: '2fr repeat(6, 1fr)',
                    gap: '1rem',
                    marginBottom: '0.75rem',
                    padding: '1rem',
                    borderRadius: '8px',
                    background: featureIndex % 2 === 0 ? 'rgba(0, 0, 0, 0.05)' : 'transparent',
                    animationDelay: `${0.4 + featureIndex * 0.05}s`,
                    alignItems: 'center',
                    minHeight: '80px',
                  }}
                >
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'flex-start',
                    gap: '0.75rem',
                    fontWeight: '500',
                    color: 'hsl(var(--foreground))',
                    height: '100%',
                  }}>
                    <div style={{ color: '#4ecdc4', flexShrink: 0 }}>
                      {feature.icon}
                    </div>
                    <span>{feature.label}</span>
                  </div>

                  {competitors.map((competitor) => (
                    <div
                      key={`${competitor.name}-${feature.key}`}
                      style={getValueStyle(competitor[feature.key], competitor.isOurs)}
                    >
                      {competitor[feature.key]}
                    </div>
                  ))}
                </div>
              ))}
            </div>
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
            className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-benchmarks'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
              maxWidth: '800px',
              margin: '0 auto',
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
        style={{ padding: '2rem 0' }}
      >
        <div className="container">
          <h2
            className={`text-center mb-4 animate-fadeInUp stagger-animation ${isVisible['section-advantages'] ? 'visible' : ''}`}
            style={{
              color: 'hsl(var(--foreground))',
              animationDelay: '0.2s',
            }}
          >
            ⭐ Key Advantages
          </h2>

          <div className="grid grid-2">
            {[
              {
                icon: '🚀',
                title: 'Ultra Fast Performance',
                description: 'Sub-millisecond response times (179µs-1.7ms) vs moderate performance of commercial libraries',
              },
              {
                icon: '💰',
                title: 'Zero Licensing Cost',
                description: 'MIT license vs $2,750-$3,500/dev/year for commercial solutions',
              },
              {
                icon: '📜',
                title: 'PDF/A-4 & PDF/UA-2',
                description: 'Full archival and accessibility compliance with sRGB ICC profiles built-in',
              },
              {
                icon: '🔒',
                title: 'Enterprise Security',
                description: 'AES-128 encryption with permissions + PKCS#7 digital signatures',
              },
              {
                icon: '🌐',
                title: 'Language Agnostic',
                description: 'REST API works with any programming language vs library-specific constraints',
              },
              {
                icon: '📦',
                title: 'Single Binary Deploy',
                description: 'Zero dependencies with Docker-ready Alpine image vs complex library management',
              },
              {
                icon: '🔤',
                title: 'Font Subsetting',
                description: 'TrueType embedding with glyph optimization for smaller file sizes',
              },
              {
                icon: '🎨',
                title: 'Built-in Web Interface',
                description: 'Ready-to-use PDF viewer, template editor, and merge UI included',
              },
            ].map((advantage, index) => (
              <div
                key={index}
                className={`card comparison-card animate-fadeInUp stagger-animation ${isVisible['section-advantages'] ? 'visible' : ''}`}
                style={{
                  padding: '1.5rem',
                  animationDelay: `${0.4 + index * 0.1}s`,
                }}
              >
                <div style={{
                  fontSize: '3rem',
                  marginBottom: '1rem',
                  textAlign: 'center',
                }}>
                  {advantage.icon}
                </div>
                <h3 style={{
                  color: 'hsl(var(--foreground))',
                  marginBottom: '0.5rem',
                  textAlign: 'center',
                }}>
                  {advantage.title}
                </h3>
                <p style={{
                  color: 'hsl(var(--muted-foreground))',
                  marginBottom: 0,
                  textAlign: 'center',
                  lineHeight: 1.6,
                }}>
                  {advantage.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section
        id="section-cta"
        style={{ padding: '2rem 0 4rem' }}
      >
        <div className="container">
          <div
            className={`card comparison-card animate-fadeInUp stagger-animation ${isVisible['section-cta'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
              maxWidth: '600px',
              margin: '0 auto',
              padding: '2rem',
              animationDelay: '0.2s',
            }}
          >
            <h2 style={{
              color: 'hsl(var(--foreground))',
              marginBottom: '1rem',
            }}>
              Ready to Try GoPdfSuit?
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              marginBottom: '2rem',
              fontSize: '1.1rem',
            }}>
              Experience the power of fast, free, and flexible PDF generation
            </p>
            <div style={{
              display: 'flex',
              gap: '1rem',
              justifyContent: 'center',
              flexWrap: 'wrap',
            }}>
              <Link
                to="/"
                className="btn"
                style={{ padding: '1rem 2rem', fontSize: '1.1rem' }}
              >
                Try Demo
              </Link>
              <a
                href="https://github.com/chinmay-sawant/gopdfsuit"
                target="_blank"
                rel="noopener noreferrer"
                className="btn btn-secondary"
                style={{ padding: '1rem 2rem', fontSize: '1.1rem' }}
              >
                View Source
              </a>
            </div>
          </div>
        </div>
      </section>

      {/* Custom Scrollbar Styles */}
      <style>
        {`
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
        `}
      </style>
    </div>
  )
}

export default Comparison