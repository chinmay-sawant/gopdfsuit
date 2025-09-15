import React, { useState } from 'react'
import { Globe, FileText, Download, RefreshCw, Eye, Settings } from 'lucide-react'

const HtmlToPdf = () => {
  const [htmlContent, setHtmlContent] = useState('')
  const [url, setUrl] = useState('')
  const [inputType, setInputType] = useState('html') // 'html' or 'url'
  const [isLoading, setIsLoading] = useState(false)
  const [pdfUrl, setPdfUrl] = useState('')
  const [showPreview, setShowPreview] = useState(false)
  
  // PDF Configuration
  const [config, setConfig] = useState({
    page_size: 'A4',
    orientation: 'Portrait',
    margin_top: '10mm',
    margin_right: '10mm',
    margin_bottom: '10mm',
    margin_left: '10mm',
    dpi: 300,
    grayscale: false,
    low_quality: false,
  })

  const convertToPdf = async () => {
    if ((!htmlContent.trim() && inputType === 'html') || (!url.trim() && inputType === 'url')) return
    
    setIsLoading(true)
    try {
      const requestBody = {
        ...config,
        ...(inputType === 'html' ? { html: htmlContent } : { url: url })
      }

      const response = await fetch('/api/v1/htmltopdf', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        setPdfUrl(url)
        
        // Also trigger download
        const link = document.createElement('a')
        link.href = url
        link.download = `html-to-pdf-${Date.now()}.pdf`
        link.click()
      } else {
        alert('Failed to convert to PDF')
      }
    } catch (error) {
      alert('Error converting to PDF: ' + error.message)
    } finally {
      setIsLoading(false)
    }
  }

  const sampleHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sample Document</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        .highlight {
            background: linear-gradient(45deg, #f39c12, #e74c3c);
            color: white;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <h1>Sample PDF Document</h1>
    <p>This is a sample HTML document that will be converted to PDF.</p>
    
    <div class="highlight">
        <h3>Highlighted Section</h3>
        <p>This section has a gradient background and will appear in the PDF.</p>
    </div>
    
    <h2>Features</h2>
    <ul>
        <li>HTML to PDF conversion</li>
        <li>CSS styling support</li>
        <li>Custom page settings</li>
        <li>High-quality output</li>
    </ul>
</body>
</html>`

  return (
    <div style={{ padding: '2rem 0', minHeight: '100vh' }}>
      <div className="container">
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <h1 style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem',
            color: 'hsl(var(--foreground))',
          }}>
            <Globe size={40} />
            HTML to PDF Converter
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem' }}>
            Convert HTML content or web pages to PDF using Chromium engine
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* Input Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={20} />
              HTML Input
            </h3>
            
            {/* Input Type Toggle */}
            <div style={{ marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
                <button
                  onClick={() => setInputType('html')}
                  className={`btn ${inputType === 'html' ? 'btn' : 'btn'}`}
                  style={{ opacity: inputType === 'html' ? 1 : 0.6 }}
                >
                  HTML Content
                </button>
                <button
                  onClick={() => setInputType('url')}
                  className={`btn ${inputType === 'url' ? 'btn' : 'btn'}`}
                  style={{ opacity: inputType === 'url' ? 1 : 0.6 }}
                >
                  Website URL
                </button>
              </div>
            </div>

            {inputType === 'html' ? (
              <div>
                <label style={{ 
                  display: 'block', 
                  marginBottom: '0.5rem', 
                  color: 'hsl(var(--foreground))',
                  fontWeight: '500',
                }}>
                  HTML Content:
                </label>
                <textarea
                  value={htmlContent}
                  onChange={(e) => setHtmlContent(e.target.value)}
                  placeholder="Enter your HTML content here..."
                  style={{
                    width: '100%',
                    height: '300px',
                    padding: '1rem',
                    borderRadius: '6px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--background))',
                    color: 'hsl(var(--foreground))',
                    fontSize: '0.9rem',
                    fontFamily: 'monospace',
                    resize: 'vertical',
                  }}
                />
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                  <button
                    onClick={() => setHtmlContent(sampleHtml)}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem' }}
                  >
                    Load Sample HTML
                  </button>
                  <button
                    onClick={() => setShowPreview(!showPreview)}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}
                  >
                    <Eye size={14} />
                    {showPreview ? 'Hide' : 'Preview'} HTML
                  </button>
                </div>
              </div>
            ) : (
              <div>
                <label style={{ 
                  display: 'block', 
                  marginBottom: '0.5rem', 
                  color: 'hsl(var(--foreground))',
                  fontWeight: '500',
                }}>
                  Website URL:
                </label>
                <input
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com"
                  style={{
                    width: '100%',
                    padding: '0.75rem',
                    borderRadius: '6px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--background))',
                    color: 'hsl(var(--foreground))',
                    fontSize: '1rem',
                    marginBottom: '1rem',
                  }}
                />
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button
                    onClick={() => setUrl('https://example.com')}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem' }}
                  >
                    Example.com
                  </button>
                  <button
                    onClick={() => setUrl('https://github.com/chinmay-sawant/gopdfsuit')}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem' }}
                  >
                    GoPdfSuit GitHub
                  </button>
                </div>
              </div>
            )}

            {/* HTML Preview */}
            {showPreview && inputType === 'html' && htmlContent && (
              <div style={{ marginTop: '1rem' }}>
                <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>HTML Preview:</h4>
                <iframe
                  srcDoc={htmlContent}
                  style={{
                    width: '100%',
                    height: '200px',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                    background: 'white',
                  }}
                  title="HTML Preview"
                />
              </div>
            )}
          </div>

          {/* Configuration & Preview Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Settings size={20} />
              PDF Configuration
            </h3>
            
            <div style={{ marginBottom: '2rem' }}>
              <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Page Size:
                  </label>
                  <select
                    value={config.page_size}
                    onChange={(e) => setConfig(prev => ({ ...prev, page_size: e.target.value }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  >
                    <option value="A4">A4</option>
                    <option value="Letter">Letter</option>
                    <option value="Legal">Legal</option>
                    <option value="A3">A3</option>
                  </select>
                </div>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Orientation:
                  </label>
                  <select
                    value={config.orientation}
                    onChange={(e) => setConfig(prev => ({ ...prev, orientation: e.target.value }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  >
                    <option value="Portrait">Portrait</option>
                    <option value="Landscape">Landscape</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Margins (Top/Bottom):
                  </label>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <input
                      type="text"
                      value={config.margin_top}
                      onChange={(e) => setConfig(prev => ({ ...prev, margin_top: e.target.value }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                    <input
                      type="text"
                      value={config.margin_bottom}
                      onChange={(e) => setConfig(prev => ({ ...prev, margin_bottom: e.target.value }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                  </div>
                </div>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Margins (Left/Right):
                  </label>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <input
                      type="text"
                      value={config.margin_left}
                      onChange={(e) => setConfig(prev => ({ ...prev, margin_left: e.target.value }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                    <input
                      type="text"
                      value={config.margin_right}
                      onChange={(e) => setConfig(prev => ({ ...prev, margin_right: e.target.value }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                  </div>
                </div>
              </div>

              <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    DPI Quality:
                  </label>
                  <select
                    value={config.dpi}
                    onChange={(e) => setConfig(prev => ({ ...prev, dpi: parseInt(e.target.value) }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  >
                    <option value={150}>150 DPI</option>
                    <option value={300}>300 DPI</option>
                    <option value={600}>600 DPI</option>
                  </select>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'hsl(var(--foreground))', fontSize: '0.9rem' }}>
                    <input
                      type="checkbox"
                      checked={config.grayscale}
                      onChange={(e) => setConfig(prev => ({ ...prev, grayscale: e.target.checked }))}
                    />
                    Grayscale
                  </label>
                  <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'hsl(var(--foreground))', fontSize: '0.9rem' }}>
                    <input
                      type="checkbox"
                      checked={config.low_quality}
                      onChange={(e) => setConfig(prev => ({ ...prev, low_quality: e.target.checked }))}
                    />
                    Low Quality (Smaller File)
                  </label>
                </div>
              </div>
            </div>

            <button 
              onClick={convertToPdf}
              disabled={isLoading || (inputType === 'html' && !htmlContent.trim()) || (inputType === 'url' && !url.trim())}
              className="btn btn-primary"
              style={{ 
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.5rem',
                marginBottom: '1rem',
              }}
            >
              {isLoading ? <RefreshCw size={16} className="spin" /> : <FileText size={16} />}
              Convert to PDF
            </button>

            {/* PDF Preview */}
            {pdfUrl && (
              <div>
                <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>PDF Preview:</h4>
                <iframe
                  src={pdfUrl}
                  style={{
                    width: '100%',
                    height: '300px',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                  }}
                  title="PDF Preview"
                />
                <button 
                  onClick={() => {
                    const link = document.createElement('a')
                    link.href = pdfUrl
                    link.download = `html-to-pdf-${Date.now()}.pdf`
                    link.click()
                  }}
                  className="btn btn-primary"
                  style={{ 
                    width: '100%', 
                    marginTop: '1rem',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '0.5rem',
                  }}
                >
                  <Download size={16} />
                  Download PDF
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

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

export default HtmlToPdf