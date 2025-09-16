import React, { useState, useRef } from 'react'
import { FileText, Download, Upload, Play, RefreshCw } from 'lucide-react'

const Viewer = () => {
  const [templateData, setTemplateData] = useState('')
  const [fileName, setFileName] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [pdfUrl, setPdfUrl] = useState('')
  const fileInputRef = useRef(null)

  const loadTemplate = async () => {
    if (!fileName.trim()) return

    setIsLoading(true)
    try {
      const response = await fetch(`/api/v1/template-data?file=${encodeURIComponent(fileName)}`)
      if (response.ok) {
        const data = await response.json()
        setTemplateData(JSON.stringify(data, null, 2))

        // Directly call the generate PDF API
        const pdfResponse = await fetch('/api/v1/generate/template-pdf', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data),
        })

        if (pdfResponse.ok) {
          const blob = await pdfResponse.blob()
          const url = URL.createObjectURL(blob)
          setPdfUrl(url)
        } else {
          alert('Failed to generate PDF')
        }
      } else {
        alert('Failed to load template')
      }
    } catch (error) {
      alert('Error loading template: ' + error.message)
    } finally {
      setIsLoading(false)
    }
  }

  const generatePDF = async () => {
    if (!templateData.trim()) return
    
    setIsLoading(true)
    try {
      const data = JSON.parse(templateData)
      const response = await fetch('/api/v1/generate/template-pdf', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        setPdfUrl(url)
      } else {
        alert('Failed to generate PDF')
      }
    } catch (error) {
      alert('Error generating PDF: ' + error.message)
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
    <div style={{ padding: '2rem 0', minHeight: '100vh' }}>
      <div className="container-full">
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <h1 style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem',
            color: 'hsl(var(--foreground))',
          }}>
            <FileText size={40} />
            PDF Viewer & Template Processor
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem' }}>
            Load JSON templates and generate PDFs with live preview
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* Template Input Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Upload size={20} />
              Template Input
            </h3>
            
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ 
                display: 'block', 
                marginBottom: '0.5rem', 
                color: 'hsl(var(--foreground))',
                fontWeight: '500',
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
                    padding: '0.75rem',
                    borderRadius: '6px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--background))',
                    color: 'hsl(var(--foreground))',
                    fontSize: '1rem',
                  }}
                />
                <button 
                  onClick={loadTemplate}
                  disabled={isLoading || !fileName.trim()}
                  className="btn btn-secondary"
                  style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
                >
                  {isLoading ? <RefreshCw size={16} className="spin" /> : <Download size={16} />}
                  Load
                </button>
              </div>
              
              <div style={{ textAlign: 'center', margin: '1rem 0', color: 'hsl(var(--muted-foreground))' }}>
                or
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
                className="btn btn-secondary"
                style={{ width: '100%' }}
              >
                <Upload size={16} />
                Upload JSON File
              </button>
            </div>

            <label style={{ 
              display: 'block', 
              marginBottom: '0.5rem', 
              color: 'hsl(var(--foreground))',
              fontWeight: '500',
            }}>
              JSON Template:
            </label>
            <textarea
              value={templateData}
              onChange={(e) => setTemplateData(e.target.value)}
              placeholder="Enter or paste your JSON template here..."
              style={{
                width: '100%',
                height: '400px',
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
            
            <button 
              onClick={generatePDF}
              disabled={isLoading || !templateData.trim()}
              className="btn"
              style={{ 
                width: '100%', 
                marginTop: '1rem',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.5rem',
              }}
            >
              {isLoading ? <RefreshCw size={16} className="spin" /> : <Play size={16} />}
              Generate PDF
            </button>
          </div>

          {/* PDF Preview Section */}
          <div className="card">
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'space-between',
              marginBottom: '1.5rem',
              flexWrap: 'wrap',
              gap: '0.5rem'
            }}>
              <h3 style={{ color: 'hsl(var(--foreground))', display: 'flex', alignItems: 'center', gap: '0.5rem', margin: 0 }}>
                <FileText size={20} />
                PDF Preview
              </h3>
              {pdfUrl && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <button 
                    onClick={() => {
                      const link = document.createElement('a')
                      link.href = pdfUrl
                      link.download = `template-pdf-${Date.now()}.pdf`
                      document.body.appendChild(link)
                      link.click()
                      document.body.removeChild(link)
                    }}
                    className="btn"
                    style={{ 
                      padding: '0.5rem 0.75rem',
                      fontSize: '0.9rem',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.25rem',
                    }}
                  >
                    <Download size={14} />
                    Download Again
                  </button>
                </div>
              )}
            </div>
            
            {pdfUrl ? (
              <div>
                <div style={{ position: 'relative', marginBottom: '1rem' }}>
                  <iframe
                    src={pdfUrl}
                    style={{
                      width: '100%',
                      height: '600px',
                      border: '1px solid hsl(var(--border))',
                      borderRadius: '6px',
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
                      background: 'rgba(0,0,0,0.1)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      borderRadius: '6px',
                    }}>
                      <div style={{
                        background: 'hsl(var(--card))',
                        padding: '1rem 2rem',
                        borderRadius: '8px',
                        border: '1px solid hsl(var(--border))',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                      }}>
                        <RefreshCw size={16} className="spin" />
                        Generating PDF...
                      </div>
                    </div>
                  )}
                </div>
                
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '0.75rem',
                  background: 'hsl(var(--muted))',
                  borderRadius: '6px',
                  fontSize: '0.9rem',
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                    <span style={{ color: 'hsl(var(--muted-foreground))' }}>
                      PDF generated successfully
                    </span>
                    <span style={{ 
                      background: 'hsl(var(--accent))', 
                      padding: '0.25rem 0.5rem', 
                      borderRadius: '4px',
                      fontSize: '0.8rem',
                      fontWeight: '500'
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
                    className="btn"
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
                height: '600px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'hsl(var(--muted))',
                borderRadius: '6px',
                border: '2px dashed hsl(var(--border))',
                color: 'hsl(var(--muted-foreground))',
                textAlign: 'center',
              }}>
                <div>
                  <FileText size={48} style={{ marginBottom: '1rem', opacity: 0.3 }} />
                  <p style={{ marginBottom: '0.5rem', fontSize: '1.1rem' }}>
                    Load a JSON template to start generating PDFs
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
        <div className="card" style={{ marginTop: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>📋 Sample Templates</h3>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
            {['temp_multiplepage.json', 'temp.json', 'temp_og.json'].map((sample) => (
              <button
                key={sample}
                onClick={() => {
                  setFileName(sample)
                  loadTemplate()
                }}
                className="btn btn-secondary"
                style={{ fontSize: '0.9rem' }}
              >
                {sample}
              </button>
            ))}
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

export default Viewer