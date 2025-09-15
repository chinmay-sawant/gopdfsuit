import React from 'react'
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
    { icon: <Zap />, title: 'Ultra Fast', desc: 'Sub-millisecond to low-millisecond response times' },
    { icon: <Shield />, title: 'Secure', desc: 'Path traversal protection and input validation' },
    { icon: <Download />, title: 'Self-contained', desc: 'Single binary deployment with zero dependencies' },
  ]

  return (
    <div style={{ minHeight: '100vh' }}>
      {/* Hero Section */}
      <section style={{ 
        padding: '4rem 0',
        textAlign: 'center',
      }}>
        <div className="container">
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '2rem',
          }}>
            <div style={{ fontSize: '4rem' }}>📄</div>
            <h1 style={{ 
              fontSize: '3.5rem',
              fontWeight: '800',
              color: 'hsl(var(--foreground))',
              marginBottom: 0,
            }}>
              GoPdfSuit
            </h1>
          </div>
          
          <p style={{ 
            fontSize: '1.5rem',
            marginBottom: '2rem',
            color: 'hsl(var(--muted-foreground))',
            maxWidth: '800px',
            margin: '0 auto 2rem',
          }}>
            A powerful Go web service that generates template-based PDF documents on-the-fly with 
            <strong> multi-page support</strong>, <strong>PDF merge capabilities</strong>, and 
            <strong> HTML to PDF/Image conversion</strong>.
          </p>

          <div style={{
            display: 'flex',
            gap: '1rem',
            justifyContent: 'center',
            flexWrap: 'wrap',
            marginBottom: '3rem',
          }}>
            <Link to="/viewer" className="btn btn-primary" style={{ fontSize: '1.1rem', padding: '1rem 2rem' }}>
              <FileText size={20} />
              Try PDF Generator
            </Link>
            <a 
              href="https://github.com/chinmay-sawant/gopdfsuit" 
              target="_blank" 
              rel="noopener noreferrer"
              className="btn btn-secondary"
              style={{ fontSize: '1.1rem', padding: '1rem 2rem' }}
            >
              <Github size={20} />
              View on GitHub
            </a>
          </div>

          {/* Quick Stats */}
          <div className="grid grid-3" style={{ marginTop: '3rem' }}>
            {highlights.map((highlight, index) => (
              <div key={index} className="card" style={{ textAlign: 'center', padding: '1.5rem' }}>
                <div style={{ color: '#4ecdc4', marginBottom: '1rem', display: 'flex', justifyContent: 'center' }}>
                  {React.cloneElement(highlight.icon, { size: 24 })}
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
      <section style={{ padding: '4rem 0' }}>
        <div className="container">
          <h2 className="text-center mb-8" style={{ color: 'hsl(var(--foreground))' }}>✨ Features</h2>
          
          <div className="grid grid-2">
            {features.map((feature, index) => (
              <Link 
                key={index}
                to={feature.link}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                <div className="card" style={{ 
                  height: '100%',
                  transition: 'all 0.3s ease',
                  cursor: 'pointer',
                }}>
                  <div style={{ 
                    color: '#4ecdc4', 
                    marginBottom: '1rem',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '1rem',
                  }}>
                    {feature.icon}
                    <h3 style={{ marginBottom: 0, color: 'hsl(var(--foreground))' }}>{feature.title}</h3>
                  </div>
                  <p style={{ 
                    color: 'hsl(var(--muted-foreground))',
                    marginBottom: 0,
                    lineHeight: 1.6,
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
      <section style={{ padding: '4rem 0' }}>
        <div className="container">
          <h2 className="text-center mb-4" style={{ color: 'hsl(var(--foreground))' }}>⚡ Quick Start</h2>
          
          <div className="card" style={{ maxWidth: '800px', margin: '0 auto' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🚀 Get Started in 3 Steps</h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '1rem' }}>
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
                }}>1</div>
                <div>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>Clone & Run</h4>
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
                </div>
              </div>

              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '1rem' }}>
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
                }}>2</div>
                <div>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>Server Ready</h4>
                  <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0 }}>
                    🌐 Server listening on: <code style={{ color: '#4ecdc4' }}>http://localhost:8080</code>
                  </p>
                </div>
              </div>

              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '1rem' }}>
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
                }}>3</div>
                <div>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>Start Creating</h4>
                  <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0 }}>
                    Navigate to any tool above to start generating PDFs, merging documents, or converting HTML!
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* API Overview */}
      <section style={{ padding: '4rem 0' }}>
        <div className="container">
          <h2 className="text-center mb-4" style={{ color: 'hsl(var(--foreground))' }}>📡 API Endpoints</h2>
          
          <div className="grid grid-2">
            <div className="card">
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🔧 REST API</h3>
              <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem' }}>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>POST /api/v1/generate/template-pdf</code> - Generate PDF from JSON template
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>POST /api/v1/merge</code> - Merge multiple PDF files
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>POST /api/v1/fill</code> - Fill PDF forms with XFDF data
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>POST /api/v1/htmltopdf</code> - Convert HTML to PDF
                </div>
                <div>
                  <code style={{ color: '#4ecdc4' }}>POST /api/v1/htmltoimage</code> - Convert HTML to Image
                </div>
              </div>
            </div>

            <div className="card">
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🌐 Web Interfaces</h3>
              <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem' }}>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>GET /</code> - PDF Viewer & Template Processor
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>GET /editor</code> - Drag-and-drop Template Editor
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>GET /merge</code> - PDF Merge Interface
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <code style={{ color: '#4ecdc4' }}>GET /filler</code> - PDF Form Filler
                </div>
                <div>
                  <code style={{ color: '#4ecdc4' }}>GET /htmltopdf</code> - HTML to PDF Converter
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Performance Section */}
      <section style={{ padding: '4rem 0' }}>
        <div className="container">
          <div className="card" style={{ textAlign: 'center', maxWidth: '600px', margin: '0 auto' }}>
            <h2 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🏃‍♂️ Performance</h2>
            <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '1rem' }}>
              Ultra-fast PDF generation with in-memory processing
            </p>
            <div style={{
              background: 'rgba(0, 0, 0, 0.3)',
              padding: '1rem',
              borderRadius: '8px',
              fontFamily: 'monospace',
              color: '#4ecdc4',
              fontSize: '0.9rem',
              textAlign: 'left',
            }}>
              [GIN] 200 | 697.8µs | POST "/api/v1/generate/template-pdf"<br/>
              [GIN] 200 | 179.6µs | POST "/api/v1/generate/template-pdf"<br/>
              [GIN] 200 | 573.7µs | POST "/api/v1/generate/template-pdf"
            </div>
            <p style={{ 
              color: 'hsl(var(--muted-foreground))', 
              marginTop: '1rem', 
              fontSize: '0.9rem',
              marginBottom: 0,
            }}>
              Performance benchmarks for multi-page PDF generation
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer style={{ 
        padding: '2rem 0',
        borderTop: '1px solid rgba(255, 255, 255, 0.2)',
        marginTop: '2rem',
      }}>
        <div className="container">
          <div style={{ 
            textAlign: 'center',
            color: 'hsl(var(--muted-foreground))',
          }}>
            <p style={{ marginBottom: '1rem' }}>
              Made with ❤️ and ☕ by{' '}
              <a 
                href="https://github.com/chinmay-sawant" 
                target="_blank" 
                rel="noopener noreferrer"
                style={{ color: '#4ecdc4', textDecoration: 'none' }}
              >
                Chinmay Sawant
              </a>
            </p>
            <p style={{ marginBottom: 0, fontSize: '0.9rem' }}>
              ⭐ Star this repo if you find it helpful!
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default Home