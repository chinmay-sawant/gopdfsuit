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
        
        // Also trigger download
        const link = document.createElement('a')
        link.href = url
        link.download = `template-pdf-${Date.now()}.pdf`
        link.click()
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
      <div className="container">
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <h1 style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem',
            color: 'white',
          }}>
            <FileText size={40} />
            PDF Viewer & Template Processor
          </h1>
          <p style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: '1.1rem' }}>
            Load JSON templates and generate PDFs with live preview
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* Template Input Section */}
          <div className="card">
            <h3 style={{ color: 'white', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Upload size={20} />
              Template Input
            </h3>
            
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ 
                display: 'block', 
                marginBottom: '0.5rem', 
                color: 'rgba(255, 255, 255, 0.9)',
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
                    border: '1px solid rgba(255, 255, 255, 0.3)',
                    background: 'rgba(255, 255, 255, 0.1)',
                    color: 'white',
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
              
              <div style={{ textAlign: 'center', margin: '1rem 0', color: 'rgba(255, 255, 255, 0.6)' }}>
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
              color: 'rgba(255, 255, 255, 0.9)',
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
                border: '1px solid rgba(255, 255, 255, 0.3)',
                background: 'rgba(0, 0, 0, 0.3)',
                color: 'white',
                fontSize: '0.9rem',
                fontFamily: 'monospace',
                resize: 'vertical',
              }}
            />
            
            <button 
              onClick={generatePDF}
              disabled={isLoading || !templateData.trim()}
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
              {isLoading ? <RefreshCw size={16} className="spin" /> : <Play size={16} />}
              Generate PDF
            </button>
          </div>

          {/* PDF Preview Section */}
          <div className="card">
            <h3 style={{ color: 'white', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={20} />
              PDF Preview
            </h3>
            
            {pdfUrl ? (
              <div>
                <iframe
                  src={pdfUrl}
                  style={{
                    width: '100%',
                    height: '500px',
                    border: '1px solid rgba(255, 255, 255, 0.3)',
                    borderRadius: '6px',
                  }}
                  title="PDF Preview"
                />
                <button 
                  onClick={() => {
                    const link = document.createElement('a')
                    link.href = pdfUrl
                    link.download = `template-pdf-${Date.now()}.pdf`
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
            ) : (
              <div style={{
                height: '500px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(0, 0, 0, 0.2)',
                borderRadius: '6px',
                border: '2px dashed rgba(255, 255, 255, 0.3)',
                color: 'rgba(255, 255, 255, 0.6)',
                textAlign: 'center',
              }}>
                <div>
                  <FileText size={48} style={{ marginBottom: '1rem', opacity: 0.5 }} />
                  <p>PDF preview will appear here after generation</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Sample Templates */}
        <div className="card" style={{ marginTop: '2rem' }}>
          <h3 style={{ color: 'white', marginBottom: '1rem' }}>📋 Sample Templates</h3>
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