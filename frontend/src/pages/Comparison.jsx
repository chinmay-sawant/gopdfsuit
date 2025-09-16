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
      performance: '✅ Ultra Fast (179µs-1.7ms)',
      deployment: '✅ Microservice/Sidecar/Docker',
      memory: '✅ In-Memory Processing',
      integration: '✅ REST API (Language Agnostic)',
      template: '✅ JSON-based Templates',
      webInterface: '✅ Built-in Viewer/Editor',
      formFilling: '✅ XFDF Support',
      pdfMerge: '✅ Drag & Drop UI',
      htmlConversion: '✅ gochromedp (Chromium)',
      multipage: 'Auto Page Breaks',
      styling: 'Font Styles + Borders',
      interactive: 'Checkboxes',
      pageFormats: 'A3, A4, A5, Letter, Legal',
      security: 'Basic Validation',
      dockerSupport: '✅ Built-in (Multi-stage, Alpine-based)',
      maintenance: '✅ Single Binary'
    },
    {
      name: 'UniPDF',
      pricing: '$3,990/dev/year',
      performance: 'Moderate',
      deployment: 'Library Integration',
      memory: 'File-based',
      integration: 'Go Library Only',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: '✅ Full Support',
      pdfMerge: '✅ Programmatic',
      htmlConversion: 'Requires UniHTML product',
      multipage: '✅ Manual Control',
      styling: '✅ Advanced',
      interactive: '✅ Full Support',
      pageFormats: '✅ All Formats',
      security: '✅ Advanced',
      dockerSupport: '❌ Not Applicable (Library)',
      maintenance: 'Library Updates'
    },
    {
      name: 'Aspose.PDF',
      pricing: '$1,999/dev/year',
      performance: 'Moderate',
      deployment: 'Library Integration',
      memory: 'Mixed',
      integration: '.NET/Java/C++',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: '✅ Full Support',
      pdfMerge: '✅ Programmatic',
      htmlConversion: 'Requires Aspose.HTML product',
      multipage: '✅ Manual Control',
      styling: '✅ Advanced',
      interactive: '✅ Full Support',
      pageFormats: '✅ All Formats',
      security: '✅ Enterprise',
      dockerSupport: '❌ Not Applicable (Library)',
      maintenance: 'Library Updates'
    },
    {
      name: 'iText',
      pricing: '$3,800/dev/year',
      performance: 'Moderate',
      deployment: 'Library Integration',
      memory: 'Mixed',
      integration: 'Java/.NET/Python',
      template: 'Code-based',
      webInterface: 'None',
      formFilling: '✅ Full Support',
      pdfMerge: '✅ Programmatic',
      htmlConversion: 'Requires custom integration',
      multipage: '✅ Manual Control',
      styling: '✅ Advanced',
      interactive: '✅ Full Support',
      pageFormats: '✅ All Formats',
      security: '✅ Enterprise',
      dockerSupport: '❌ Not Applicable (Library)',
      maintenance: 'Library Updates'
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
    { key: 'formFilling', label: '📋 Form Filling', icon: <CheckCircle size={20} /> },
    { key: 'pdfMerge', label: '🔗 PDF Merge', icon: <CheckCircle size={20} /> },
    { key: 'htmlConversion', label: '🌐 HTML to PDF/Image', icon: <Globe size={20} /> },
    { key: 'multipage', label: '📱 Multi-page Support', icon: <CheckCircle size={20} /> },
    { key: 'styling', label: '🎨 Styling', icon: <Star size={20} /> },
    { key: 'interactive', label: '☑️ Interactive Elements', icon: <CheckCircle size={20} /> },
    { key: 'pageFormats', label: '📏 Page Formats', icon: <CheckCircle size={20} /> },
    { key: 'security', label: '🔒 Security', icon: <CheckCircle size={20} /> },
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
            <div style={{ minWidth: '800px' }}>
              {/* Header Row */}
              <div style={{ 
                display: 'grid',
                gridTemplateColumns: '2fr repeat(4, 1fr)',
                gap: '1rem',
                marginBottom: '1rem',
                paddingBottom: '1rem',
                borderBottom: '2px solid rgba(78, 205, 196, 0.3)',
              }}>
                <div style={{ fontWeight: 'bold', fontSize: '1.1rem', color: 'hsl(var(--foreground))' }}>
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
                      padding: '0.5rem',
                      borderRadius: '8px',
                      background: competitor.isOurs ? 'rgba(78, 205, 196, 0.1)' : 'transparent',
                    }}
                  >
                    {competitor.name}
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
                    gridTemplateColumns: '2fr repeat(4, 1fr)',
                    gap: '1rem',
                    marginBottom: '0.75rem',
                    padding: '0.5rem',
                    borderRadius: '8px',
                    background: featureIndex % 2 === 0 ? 'rgba(0, 0, 0, 0.05)' : 'transparent',
                    animationDelay: `${0.4 + featureIndex * 0.05}s`,
                  }}
                >
                  <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '0.75rem',
                    fontWeight: '500',
                    color: 'hsl(var(--foreground))',
                  }}>
                    <div style={{ color: '#4ecdc4' }}>
                      {feature.icon}
                    </div>
                    {feature.label}
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
          <h2 
            className={`text-center mb-4 animate-fadeInUp stagger-animation ${isVisible['section-benchmarks'] ? 'visible' : ''}`}
            style={{ 
              color: 'hsl(var(--foreground))',
              animationDelay: '0.2s',
            }}
          >
            🏃‍♂️ Performance Benchmarks
          </h2>
          
          <div 
            className={`card comparison-card animate-fadeInUp stagger-animation ${isVisible['section-benchmarks'] ? 'visible' : ''}`}
            style={{ 
              maxWidth: '800px',
              margin: '0 auto',
              textAlign: 'center',
              animationDelay: '0.4s',
            }}
          >
            <div style={{ marginBottom: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>
                GoPdfSuit Performance (temp_multiplepage.json - 2 pages)
              </h3>
              
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
                <div style={{ marginBottom: '0.5rem', fontWeight: 'bold' }}>Recent Benchmark Results:</div>
                [GIN] 2025/08/28 - 00:40:18 | 200 |       697.8µs | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:40:55 | 200 |      1.7542ms | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:40:57 | 200 |       179.6µs | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:40:58 | 200 |       573.7µs | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:41:02 | 200 |       445.2µs | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:41:05 | 200 |      1.2341ms | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:41:08 | 200 |       892.1µs | POST "/api/v1/generate/template-pdf"<br/>
                [GIN] 2025/08/28 - 00:41:12 | 200 |       634.7µs | POST "/api/v1/generate/template-pdf"
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
              <div style={{ 
                background: 'rgba(78, 205, 196, 0.1)',
                padding: '1.5rem', 
                borderRadius: '8px',
                border: '1px solid rgba(78, 205, 196, 0.3)',
              }}>
                <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#4ecdc4' }}>179µs - 1.7ms</div>
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>Response Time Range</div>
              </div>
              <div style={{ 
                background: 'rgba(0, 122, 204, 0.1)',
                padding: '1.5rem', 
                borderRadius: '8px',
                border: '1px solid rgba(0, 122, 204, 0.3)',
              }}>
                <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#007acc' }}>In-Memory</div>
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>Processing Method</div>
              </div>
              <div style={{ 
                background: 'rgba(255, 193, 7, 0.1)',
                padding: '1.5rem', 
                borderRadius: '8px',
                border: '1px solid rgba(255, 193, 7, 0.3)',
              }}>
                <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#ffc107' }}>Zero</div>
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>Dependencies</div>
              </div>
            </div>
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
                description: 'Sub-millisecond to low-millisecond response times vs moderate performance of competitors',
              },
              {
                icon: '💰',
                title: 'Cost Effective',
                description: 'MIT license vs $2k-$4k/developer/year licensing costs',
              },
              {
                icon: '🔧',
                title: 'Easy Deployment',
                description: 'Microservice architecture vs complex library integration requirements',
              },
              {
                icon: '🌐',
                title: 'Language Agnostic',
                description: 'REST API accessible from any programming language vs library-specific constraints',
              },
              {
                icon: '📦',
                title: 'Zero Dependencies',
                description: 'Single binary deployment vs managing multiple library dependencies',
              },
              {
                icon: '🎨',
                title: 'Built-in Web Interface',
                description: 'Ready-to-use viewer/editor vs no web interface in competitors',
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