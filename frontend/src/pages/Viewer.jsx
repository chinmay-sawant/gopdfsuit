import React, { useState, useRef } from 'react'
import { FileText, Download, Upload, Play, RefreshCw, Sparkles } from 'lucide-react'
import { makeAuthenticatedRequest } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'
import BackgroundAnimation from '../components/BackgroundAnimation'

const Viewer = () => {
  const [templateData, setTemplateData] = useState('')
  const [fileName, setFileName] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [pdfUrl, setPdfUrl] = useState('')
  const fileInputRef = useRef(null)
  const { getAuthHeaders, triggerLogin } = useAuth()

  // Check if running on GitHub Pages
  const isGitHubPages = window.location.hostname.includes('chinmay-sawant.github.io')

  const showError = (message) => {
    alert(message)
  }

  const loadTemplate = async () => {
    if (!fileName.trim()) return

    setIsLoading(true)
    try {
      const response = await makeAuthenticatedRequest(`/api/v1/template-data?file=${encodeURIComponent(fileName)}`, {}, getAuthHeaders)
      const data = await response.json()
      setTemplateData(JSON.stringify(data, null, 2))

      // Directly call the generate PDF API
      const pdfResponse = await makeAuthenticatedRequest('/api/v1/generate/template-pdf', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      }, getAuthHeaders)

      const blob = await pdfResponse.blob()
      const url = URL.createObjectURL(blob)
      setPdfUrl(url)
    } catch (error) {
      if (error.message.includes("Authentication failed") || error.message.includes("401") || error.message.includes("403") || error.message.includes("Not authenticated")) {
        triggerLogin()
      } else {
        showError('Error loading template: ' + error.message)
      }
    } finally {
      setIsLoading(false)
    }
  }

  const generatePDF = async () => {
    if (!templateData.trim()) return

    setIsLoading(true)
    try {
      const data = JSON.parse(templateData)
      const response = await makeAuthenticatedRequest('/api/v1/generate/template-pdf', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      }, getAuthHeaders)

      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      setPdfUrl(url)
    } catch (error) {
      if (error.message.includes("Authentication failed") || error.message.includes("401") || error.message.includes("403") || error.message.includes("Not authenticated")) {
        triggerLogin()
      } else {
        showError('Error generating PDF: ' + error.message)
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleFileUpload = (event) => {
    const file = event.target.files[0]
    if (file && file.type === 'application/json') {
      const reader = new FileReader()
      reader.onload = (e) => {
        setTemplateData(e.target.result)
        setFileName(file.name)
      }
      reader.readAsText(file)
    }
  }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />

      {/* Hero Header */}
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          {/* Badge */}
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
            Template-based PDF Generation
          </div>

          <h1
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '1rem',
              marginBottom: '1rem',
              fontSize: 'clamp(2rem, 5vw, 3rem)',
              fontWeight: '800',
              color: 'hsl(var(--foreground))',
            }}
          >
            <div className="feature-icon-box teal" style={{ width: '56px', height: '56px', marginBottom: 0 }}>
              <FileText size={28} />
            </div>
            PDF Viewer
          </h1>
          <p style={{
            color: 'hsl(var(--muted-foreground))',
            fontSize: '1.1rem',
            maxWidth: '600px',
            margin: '0 auto',
          }}>
            Load JSON templates and generate PDFs with live preview
          </p>
        </div>
      </section>

      {/* Main Content */}
      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container-full">
          <div className="grid grid-2" style={{ gap: '2rem' }}>
            {/* Template Input Section */}
            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{
                color: 'hsl(var(--foreground))',
                marginBottom: '1.5rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                fontSize: '1.2rem',
                fontWeight: '700',
              }}>
                <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}>
                  <Upload size={18} />
                </div>
                Template Input
              </h3>

              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{
                  display: 'block',
                  marginBottom: '0.5rem',
                  color: 'hsl(var(--foreground))',
                  fontWeight: '600',
                  fontSize: '0.9rem',
                }}>
                  Load from file:
                </label>
                <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
                  <input
                    type="text"
                    value={fileName}
                    onChange={(e) => setFileName(e.target.value)}
                    placeholder="Enter filename (e.g., temp.json)"
                    style={{
                      flex: 1,
                      padding: '0.75rem 1rem',
                      borderRadius: '8px',
                      border: '1px solid rgba(255, 255, 255, 0.15)',
                      background: 'rgba(255, 255, 255, 0.05)',
                      color: 'hsl(var(--foreground))',
                      fontSize: '0.95rem',
                      transition: 'all 0.2s ease',
                    }}
                  />
                  <button
                    onClick={loadTemplate}
                    disabled={isLoading || !fileName.trim()}
                    className="btn-glow"
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      padding: '0.75rem 1.25rem',
                    }}
                  >
                    {isLoading ? <RefreshCw size={16} className="spin" /> : <Download size={16} />}
                    Load
                  </button>
                </div>

                <div style={{
                  textAlign: 'center',
                  margin: '1.25rem 0',
                  color: 'hsl(var(--muted-foreground))',
                  fontSize: '0.9rem',
                  position: 'relative',
                }}>
                  <span style={{
                    background: 'hsl(var(--background))',
                    padding: '0 1rem',
                    position: 'relative',
                    zIndex: 1,
                  }}>or</span>
                  <div style={{
                    position: 'absolute',
                    top: '50%',
                    left: 0,
                    right: 0,
                    height: '1px',
                    background: 'rgba(255, 255, 255, 0.1)',
                  }} />
                </div>

                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".json"
                  onChange={handleFileUpload}
                  style={{ display: 'none' }}
                />
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="btn-outline-glow"
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '0.5rem',
                  }}
                >
                  <Upload size={16} />
                  Upload JSON File
                </button>
              </div>

              <label style={{
                display: 'block',
                marginBottom: '0.5rem',
                color: 'hsl(var(--foreground))',
                fontWeight: '600',
                fontSize: '0.9rem',
              }}>
                JSON Template:
              </label>
              <textarea
                value={templateData}
                onChange={(e) => setTemplateData(e.target.value)}
                placeholder="Enter or paste your JSON template here..."
                style={{
                  width: '100%',
                  height: '350px',
                  padding: '1rem',
                  borderRadius: '8px',
                  border: '1px solid rgba(255, 255, 255, 0.15)',
                  background: 'rgba(255, 255, 255, 0.05)',
                  color: 'hsl(var(--foreground))',
                  fontSize: '0.9rem',
                  fontFamily: "'SF Mono', 'Monaco', 'Cascadia Code', 'Consolas', monospace",
                  resize: 'vertical',
                  transition: 'all 0.2s ease',
                }}
              />

              <button
                onClick={generatePDF}
                disabled={isLoading || !templateData.trim()}
                className="btn-glow"
                style={{
                  width: '100%',
                  marginTop: '1.5rem',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '0.5rem',
                  padding: '1rem 2rem',
                }}
              >
                {isLoading ? <RefreshCw size={18} className="spin" /> : <Play size={18} />}
                Generate PDF
              </button>
            </div>

            {/* PDF Preview Section */}
            <div className="glass-card" style={{ padding: '2rem' }}>
              <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '1.5rem',
                flexWrap: 'wrap',
                gap: '0.5rem'
              }}>
                <h3 style={{
                  color: 'hsl(var(--foreground))',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  margin: 0,
                  fontSize: '1.2rem',
                  fontWeight: '700',
                }}>
                  <div className="feature-icon-box purple" style={{ width: '40px', height: '40px', marginBottom: 0 }}>
                    <FileText size={18} />
                  </div>
                  PDF Preview
                </h3>
                {pdfUrl && (
                  <button
                    onClick={() => {
                      const link = document.createElement('a')
                      link.href = pdfUrl
                      link.download = `template-pdf-${Date.now()}.pdf`
                      document.body.appendChild(link)
                      link.click()
                      document.body.removeChild(link)
                    }}
                    className="btn-glow"
                    style={{
                      padding: '0.5rem 1rem',
                      fontSize: '0.9rem',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                    }}
                  >
                    <Download size={14} />
                    Download
                  </button>
                )}
              </div>

              {pdfUrl ? (
                <div>
                  <div style={{ position: 'relative', marginBottom: '1rem' }}>
                    <iframe
                      src={pdfUrl}
                      style={{
                        width: '100%',
                        height: '550px',
                        border: '1px solid rgba(255, 255, 255, 0.15)',
                        borderRadius: '8px',
                        background: 'white',
                      }}
                      title="PDF Preview"
                    />
                    {isLoading && (
                      <div style={{
                        position: 'absolute',
                        top: '0',
                        left: '0',
                        right: '0',
                        bottom: '0',
                        background: 'rgba(0,0,0,0.4)',
                        backdropFilter: 'blur(4px)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        borderRadius: '8px',
                      }}>
                        <div style={{
                          background: 'rgba(30, 30, 30, 0.9)',
                          padding: '1rem 2rem',
                          borderRadius: '8px',
                          border: '1px solid rgba(255, 255, 255, 0.15)',
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.75rem',
                          color: '#4ecdc4',
                        }}>
                          <RefreshCw size={18} className="spin" />
                          Generating PDF...
                        </div>
                      </div>
                    )}
                  </div>

                  <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '1rem',
                    background: 'rgba(78, 205, 196, 0.08)',
                    borderRadius: '8px',
                    border: '1px solid rgba(78, 205, 196, 0.2)',
                    fontSize: '0.9rem',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                      <span style={{ color: 'hsl(var(--muted-foreground))' }}>
                        PDF generated successfully
                      </span>
                      <span style={{
                        background: 'rgba(78, 205, 196, 0.2)',
                        color: '#4ecdc4',
                        padding: '0.25rem 0.75rem',
                        borderRadius: '20px',
                        fontSize: '0.8rem',
                        fontWeight: '600'
                      }}>
                        Preview Ready
                      </span>
                    </div>
                    <button
                      onClick={() => {
                        const link = document.createElement('a')
                        link.href = pdfUrl
                        link.download = `template-pdf-${Date.now()}.pdf`
                        document.body.appendChild(link)
                        link.click()
                        document.body.removeChild(link)
                      }}
                      className="btn-glow"
                      style={{
                        padding: '0.5rem 1rem',
                        fontSize: '0.9rem',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                      }}
                    >
                      <Download size={14} />
                      Download PDF
                    </button>
                  </div>
                </div>
              ) : (
                <div style={{
                  height: '550px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'rgba(255, 255, 255, 0.02)',
                  borderRadius: '8px',
                  border: '2px dashed rgba(255, 255, 255, 0.1)',
                  color: 'hsl(var(--muted-foreground))',
                  textAlign: 'center',
                }}>
                  <div>
                    <div className="feature-icon-box teal" style={{
                      width: '64px',
                      height: '64px',
                      margin: '0 auto 1rem',
                      opacity: 0.5,
                    }}>
                      <FileText size={32} />
                    </div>
                    <p style={{ marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>
                      Load a JSON template to start
                    </p>
                    <p style={{ fontSize: '0.9rem', opacity: 0.7, marginBottom: 0 }}>
                      Enter template data above and click "Generate PDF" to see the preview
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Sample Templates */}
          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{
              color: 'hsl(var(--foreground))',
              marginBottom: '1.25rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.75rem',
              fontSize: '1.1rem',
              fontWeight: '700',
            }}>
              <div className="feature-icon-box yellow" style={{ width: '40px', height: '40px', marginBottom: 0 }}>
                <span style={{ fontSize: '1.2rem' }}>📋</span>
              </div>
              Sample Templates
            </h3>
            <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
              {['temp_multiplepage.json', 'temp.json', 'temp_og.json'].map((sample) => (
                <button
                  key={sample}
                  onClick={() => {
                    setFileName(sample)
                    loadTemplate()
                  }}
                  className="btn-outline-glow"
                  style={{
                    fontSize: '0.9rem',
                    padding: '0.75rem 1.25rem',
                  }}
                >
                  {sample}
                </button>
              ))}
            </div>
          </div>
        </div>
      </section>

      <style jsx>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .spin {
          animation: spin 1s linear infinite;
        }
      `}</style>
    </div>
  )
}

export default Viewer