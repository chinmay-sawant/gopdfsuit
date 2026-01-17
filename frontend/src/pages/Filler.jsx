import React, { useState, useRef } from 'react'
import { FileCheck, Upload, Download, RefreshCw, FileText } from 'lucide-react'

const Filler = () => {
  const [pdfFile, setPdfFile] = useState(null)
  const [xfdfFile, setXfdfFile] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [filledPdfUrl, setFilledPdfUrl] = useState('')
  const pdfInputRef = useRef(null)
  const xfdfInputRef = useRef(null)

  const handlePdfUpload = (event) => {
    const file = event.target.files[0]
    if (file && file.type === 'application/pdf') {
      setPdfFile(file)
    }
  }

  const handleXfdfUpload = (event) => {
    const file = event.target.files[0]
    if (file) {
      setXfdfFile(file)
    }
  }

  const fillPDF = async () => {
    if (!pdfFile || !xfdfFile) return
    
    setIsLoading(true)
    try {
      const formData = new FormData()
      formData.append('pdf', pdfFile)
      formData.append('xfdf', xfdfFile)

      const response = await fetch('/api/v1/fill', {
        method: 'POST',
        body: formData,
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        setFilledPdfUrl(url)
        
        // Also trigger download
        const link = document.createElement('a')
        link.href = url
        link.download = `filled-${pdfFile.name}`
        link.click()
      } else {
        alert('Failed to fill PDF')
      }
    } catch (error) {
      alert('Error filling PDF: ' + error.message)
    } finally {
      setIsLoading(false)
    }
  }

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
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
            color: 'hsl(var(--foreground))',
          }}>
            <FileCheck size={40} />
            PDF Form Filler
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem' }}>
            Fill PDF forms using XFDF data with AcroForm support
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* Upload Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Upload size={20} />
              Upload Files
            </h3>
            
            {/* PDF Upload */}
            <div style={{ marginBottom: '2rem' }}>
              <label style={{ 
                display: 'block', 
                marginBottom: '0.5rem', 
                color: 'hsl(var(--foreground))',
                fontWeight: '500',
              }}>
                PDF File (AcroForm):
              </label>
              <input
                ref={pdfInputRef}
                type="file"
                accept=".pdf"
                onChange={handlePdfUpload}
                style={{ display: 'none' }}
              />
              <div
                onClick={() => pdfInputRef.current?.click()}
                style={{
                  border: '2px dashed hsl(var(--border))',
                  borderRadius: '8px',
                  padding: '2rem',
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease',
                  background: 'hsl(var(--muted))',
                }}
              >
                <FileText size={32} style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }} />
                <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                  {pdfFile ? pdfFile.name : 'Click to upload PDF file'}
                </p>
                {pdfFile && (
                  <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem', marginBottom: 0 }}>
                    {formatFileSize(pdfFile.size)}
                  </p>
                )}
              </div>
            </div>

            {/* XFDF Upload */}
            <div>
              <label style={{ 
                display: 'block', 
                marginBottom: '0.5rem', 
                color: 'hsl(var(--foreground))',
                fontWeight: '500',
              }}>
                XFDF File (Form Data):
              </label>
              <input
                ref={xfdfInputRef}
                type="file"
                accept=".xfdf,.xml"
                onChange={handleXfdfUpload}
                style={{ display: 'none' }}
              />
              <div
                onClick={() => xfdfInputRef.current?.click()}
                style={{
                  border: '2px dashed hsl(var(--border))',
                  borderRadius: '8px',
                  padding: '2rem',
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease',
                  background: 'hsl(var(--muted))',
                }}
              >
                <FileText size={32} style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }} />
                <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                  {xfdfFile ? xfdfFile.name : 'Click to upload XFDF file'}
                </p>
                {xfdfFile && (
                  <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem', marginBottom: 0 }}>
                    {formatFileSize(xfdfFile.size)}
                  </p>
                )}
              </div>
            </div>

            <button 
              onClick={fillPDF}
              disabled={isLoading || !pdfFile || !xfdfFile}
              className="btn"
              style={{ 
                width: '100%', 
                marginTop: '1.5rem',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.5rem',
              }}
            >
              {isLoading ? <RefreshCw size={16} className="spin" /> : <FileCheck size={16} />}
              Fill PDF Form
            </button>
          </div>

          {/* Preview Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={20} />
              Filled PDF Preview
            </h3>
            
            {filledPdfUrl ? (
              <div>
                <iframe
                  src={filledPdfUrl}
                  style={{
                    width: '100%',
                    height: '500px',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                  }}
                  title="Filled PDF Preview"
                />
                <button 
                  onClick={() => {
                    const link = document.createElement('a')
                    link.href = filledPdfUrl
                    link.download = `filled-${pdfFile?.name || 'form.pdf'}`
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
                  Download Filled PDF
                </button>
              </div>
            ) : (
              <div style={{
                height: '500px',
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
                  <FileCheck size={48} style={{ marginBottom: '1rem', opacity: 0.5 }} />
                  <p>Filled PDF preview will appear here</p>
                  <p style={{ fontSize: '0.9rem', opacity: 0.7 }}>
                    Upload both PDF and XFDF files to get started
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Instructions */}
        <div className="card" style={{ marginTop: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>📋 How to Use PDF Form Filler</h3>
          
          <div className="grid grid-2" style={{ gap: '2rem', marginBottom: '2rem' }}>
            <div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '1rem' }}>Steps:</h4>
              <ol style={{ color: 'hsl(var(--muted-foreground))', lineHeight: 1.8, paddingLeft: '1.5rem' }}>
                <li>Upload a PDF file with AcroForm fields</li>
                <li>Upload an XFDF file containing the form data</li>
                <li>Click "Fill PDF Form" to process</li>
                <li>Preview and download the filled PDF</li>
              </ol>
            </div>
            
            <div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '1rem' }}>File Requirements:</h4>
              <ul style={{ color: 'hsl(var(--muted-foreground))', lineHeight: 1.8, paddingLeft: '1.5rem' }}>
                <li><strong>PDF:</strong> Must contain AcroForm fields</li>
                <li><strong>XFDF:</strong> XML file with field data mappings</li>
                <li>Field names in XFDF must match PDF form fields</li>
                <li>Supports text fields, checkboxes, and radio buttons</li>
              </ul>
            </div>
          </div>

          <div style={{ 
            padding: '1rem',
            background: 'color-mix(in hsl, var(--primary-color) 10%, transparent)',
            borderRadius: '6px',
            border: '1px solid color-mix(in hsl, var(--primary-color) 30%, transparent)',
          }}>
            <h4 style={{ color: 'var(--primary-color)', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={16} />
              Sample Files Available
            </h4>
            <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0, fontSize: '0.9rem' }}>
              Check the <code style={{ color: '#4ecdc4' }}>sampledata/</code> directory for example PDF and XFDF files to test with.
            </p>
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

export default Filler