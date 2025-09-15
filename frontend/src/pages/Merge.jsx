import React, { useState, useRef } from 'react'
import { Merge, Upload, Download, RefreshCw, FileText, X } from 'lucide-react'

const MergePage = () => {
  const [files, setFiles] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [mergedPdfUrl, setMergedPdfUrl] = useState('')
  const fileInputRef = useRef(null)

  const handleFileUpload = (event) => {
    const newFiles = Array.from(event.target.files).filter(file => file.type === 'application/pdf')
    setFiles(prev => [...prev, ...newFiles])
    event.target.value = '' // Reset input
  }

  const removeFile = (index) => {
    setFiles(prev => prev.filter((_, i) => i !== index))
  }

  const moveFile = (index, direction) => {
    const newFiles = [...files]
    const targetIndex = direction === 'up' ? index - 1 : index + 1
    
    if (targetIndex >= 0 && targetIndex < files.length) {
      [newFiles[index], newFiles[targetIndex]] = [newFiles[targetIndex], newFiles[index]]
      setFiles(newFiles)
    }
  }

  const mergePDFs = async () => {
    if (files.length < 2) return
    
    setIsLoading(true)
    try {
      const formData = new FormData()
      files.forEach(file => {
        formData.append('pdf', file)
      })

      const response = await fetch('/api/v1/merge', {
        method: 'POST',
        body: formData,
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        setMergedPdfUrl(url)
        
        // Also trigger download
        const link = document.createElement('a')
        link.href = url
        link.download = `merged-pdf-${Date.now()}.pdf`
        link.click()
      } else {
        alert('Failed to merge PDFs')
      }
    } catch (error) {
      alert('Error merging PDFs: ' + error.message)
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
            <Merge size={40} />
            PDF Merge Tool
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem' }}>
            Combine multiple PDF files with drag-and-drop reordering
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* File Upload Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Upload size={20} />
              Upload PDF Files
            </h3>
            
            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf"
              multiple
              onChange={handleFileUpload}
              style={{ display: 'none' }}
            />
            
            <div
              onClick={() => fileInputRef.current?.click()}
              style={{
                border: '2px dashed hsl(var(--border))',
                borderRadius: '8px',
                padding: '3rem 2rem',
                textAlign: 'center',
                cursor: 'pointer',
                transition: 'all 0.3s ease',
                marginBottom: '2rem',
                background: 'hsl(var(--muted))',
              }}
              onDragOver={(e) => {
                e.preventDefault()
                e.currentTarget.style.borderColor = 'var(--secondary-color)'
                e.currentTarget.style.background = 'color-mix(in hsl, var(--secondary-color) 10%, transparent)'
              }}
              onDragLeave={(e) => {
                e.currentTarget.style.borderColor = 'hsl(var(--border))'
                e.currentTarget.style.background = 'hsl(var(--muted))'
              }}
              onDrop={(e) => {
                e.preventDefault()
                const droppedFiles = Array.from(e.dataTransfer.files).filter(file => file.type === 'application/pdf')
                setFiles(prev => [...prev, ...droppedFiles])
                e.currentTarget.style.borderColor = 'hsl(var(--border))'
                e.currentTarget.style.background = 'hsl(var(--muted))'
              }}
            >
              <Upload size={48} style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '1rem' }} />
              <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem', fontSize: '1.1rem' }}>
                Click to upload or drag & drop PDF files
              </p>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Select multiple files to merge them together
              </p>
            </div>

            {files.length > 0 && (
              <div>
                <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>
                  Selected Files ({files.length})
                </h4>
                <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                  {files.map((file, index) => (
                    <div
                      key={index}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '1rem',
                        padding: '0.75rem',
                        background: 'hsl(var(--accent))',
                        borderRadius: '6px',
                        marginBottom: '0.5rem',
                      }}
                    >
                      <FileText size={16} style={{ color: '#4ecdc4' }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ 
                          color: 'hsl(var(--foreground))', 
                          fontSize: '0.9rem',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}>
                          {file.name}
                        </div>
                        <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>
                          {formatFileSize(file.size)}
                        </div>
                      </div>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          onClick={() => moveFile(index, 'up')}
                          disabled={index === 0}
                          style={{
                            background: 'none',
                            border: '1px solid hsl(var(--border))',
                            color: 'hsl(var(--foreground))',
                            borderRadius: '4px',
                            padding: '0.25rem 0.5rem',
                            cursor: index === 0 ? 'not-allowed' : 'pointer',
                            opacity: index === 0 ? 0.5 : 1,
                          }}
                        >
                          ↑
                        </button>
                        <button
                          onClick={() => moveFile(index, 'down')}
                          disabled={index === files.length - 1}
                          style={{
                            background: 'none',
                            border: '1px solid hsl(var(--border))',
                            color: 'hsl(var(--foreground))',
                            borderRadius: '4px',
                            padding: '0.25rem 0.5rem',
                            cursor: index === files.length - 1 ? 'not-allowed' : 'pointer',
                            opacity: index === files.length - 1 ? 0.5 : 1,
                          }}
                        >
                          ↓
                        </button>
                        <button
                          onClick={() => removeFile(index)}
                          style={{
                            background: 'none',
                            border: '1px solid rgba(255, 0, 0, 0.5)',
                            color: '#ff6b6b',
                            borderRadius: '4px',
                            padding: '0.25rem 0.5rem',
                            cursor: 'pointer',
                          }}
                        >
                          <X size={12} />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
                
                <button 
                  onClick={mergePDFs}
                  disabled={isLoading || files.length < 2}
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
                  {isLoading ? <RefreshCw size={16} className="spin" /> : <Merge size={16} />}
                  Merge PDFs
                </button>
              </div>
            )}
          </div>

          {/* Preview Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={20} />
              Merged PDF Preview
            </h3>
            
            {mergedPdfUrl ? (
              <div>
                <iframe
                  src={mergedPdfUrl}
                  style={{
                    width: '100%',
                    height: '500px',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                  }}
                  title="Merged PDF Preview"
                />
                <button 
                  onClick={() => {
                    const link = document.createElement('a')
                    link.href = mergedPdfUrl
                    link.download = `merged-pdf-${Date.now()}.pdf`
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
                  Download Merged PDF
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
                  <Merge size={48} style={{ marginBottom: '1rem', opacity: 0.5 }} />
                  <p>Merged PDF preview will appear here</p>
                  <p style={{ fontSize: '0.9rem', opacity: 0.7 }}>
                    Upload at least 2 PDF files to get started
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Instructions */}
        <div className="card" style={{ marginTop: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>📋 How to Use</h3>
          <div className="grid grid-3" style={{ gap: '1rem' }}>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>1️⃣</div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem' }}>Upload PDFs</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Click or drag & drop multiple PDF files
              </p>
            </div>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>2️⃣</div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem' }}>Reorder</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Use ↑↓ buttons to change the order
              </p>
            </div>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>3️⃣</div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem' }}>Merge</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Click "Merge PDFs" to combine and download
              </p>
            </div>
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

export default MergePage