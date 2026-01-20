import React, { useState, useRef } from 'react'
import { Scissors, Upload, Download, RefreshCw, FileText, X } from 'lucide-react'
import { makeAuthenticatedRequest } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'

import BackgroundAnimation from '../components/BackgroundAnimation'

const SplitPage = () => {
  const [file, setFile] = useState(null)
  const [pages, setPages] = useState('')
  const [maxPerFile, setMaxPerFile] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [splitPdfUrl, setSplitPdfUrl] = useState('')
  const fileInputRef = useRef(null)
  const { getAuthHeaders, triggerLogin } = useAuth()

  const handleFileUpload = (event) => {
    const selectedFile = Array.from(event.target.files).find(file => file.type === 'application/pdf')
    if (selectedFile) {
      setFile(selectedFile)
    }
    event.target.value = '' // Reset input
  }

  const removeFile = () => {
    setFile(null)
    setSplitPdfUrl('')
  }

  const splitPDF = async () => {
    if (!file) return

    setIsLoading(true)
    try {
      const formData = new FormData()
      formData.append('pdf', file)
      if (pages) formData.append('pages', pages)
      if (maxPerFile) formData.append('max_per_file', maxPerFile)

      const response = await makeAuthenticatedRequest('/api/v1/split', {
        method: 'POST',
        body: formData,
      }, getAuthHeaders)

      //TODO: Output specific error messages from backend

      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      setSplitPdfUrl(url)

    } catch (error) {
      if (error.message.includes("Authentication failed") || error.message.includes("401") || error.message.includes("403") || error.message.includes("Not authenticated")) {
        triggerLogin()
      } else {
        alert('Error splitting PDF: ' + error.message)
      }
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
    <div style={{ padding: '2rem 0', minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
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
            <Scissors size={40} />
            PDF Split Tool
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem' }}>
            Extract specific pages or split PDF into multiple files
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* File Upload Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Upload size={20} />
              Upload PDF File
            </h3>

            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf"
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
                const droppedFile = Array.from(e.dataTransfer.files).find(file => file.type === 'application/pdf')
                if (droppedFile) {
                  setFile(droppedFile)
                }
                e.currentTarget.style.borderColor = 'hsl(var(--border))'
                e.currentTarget.style.background = 'hsl(var(--muted))'
              }}
            >
              <Upload size={48} style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '1rem' }} />
              <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem', fontSize: '1.1rem' }}>
                Click to upload or drag & drop PDF file
              </p>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Select a PDF file to split
              </p>
            </div>

            {file && (
              <div>
                <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>
                  Selected File
                </h4>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '1rem',
                  padding: '0.75rem',
                  background: 'hsl(var(--accent))',
                  borderRadius: '6px',
                  marginBottom: '1rem',
                }}>
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
                  <button
                    onClick={removeFile}
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

                <div style={{ marginBottom: '1rem' }}>
                  <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>
                    Pages (optional)
                  </label>
                  <input
                    type="text"
                    value={pages}
                    onChange={(e) => setPages(e.target.value)}
                    placeholder="e.g. 1-3,5"
                    style={{ width: '100%' }}
                    disabled={isLoading}
                  />
                  <small style={{ color: 'hsl(var(--muted-foreground))' }}>
                    Specify page ranges or leave empty
                  </small>
                </div>

                <div style={{ marginBottom: '1rem' }}>
                  <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>
                    Max per file (optional)
                  </label>
                  <input
                    type="number"
                    min="1"
                    value={maxPerFile}
                    onChange={(e) => setMaxPerFile(e.target.value)}
                    placeholder="e.g. 1"
                    style={{ width: '100%' }}
                    disabled={isLoading}
                  />
                  <small style={{ color: 'hsl(var(--muted-foreground))' }}>
                    Split into files with this many pages each
                  </small>
                </div>

                <button
                  onClick={splitPDF}
                  disabled={isLoading}
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
                  {isLoading ? <RefreshCw size={16} className="spin" /> : <Scissors size={16} />}
                  Split PDF
                </button>
              </div>
            )}
          </div>

          {/* Preview Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={20} />
              Split PDF Preview
            </h3>

            {splitPdfUrl ? (
              <div>
                <iframe
                  src={splitPdfUrl}
                  style={{
                    width: '100%',
                    height: '500px',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                  }}
                  title="Split PDF Preview"
                />
                <button
                  onClick={() => {
                    const link = document.createElement('a')
                    link.href = splitPdfUrl
                    link.download = `split-pdf-${Date.now()}.pdf`
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
                  Download Split PDF
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
                  <Scissors size={48} style={{ marginBottom: '1rem', opacity: 0.5 }} />
                  <p>Split PDF preview will appear here</p>
                  <p style={{ fontSize: '0.9rem', opacity: 0.7 }}>
                    Upload a PDF file to get started
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
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem' }}>Upload PDF</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Click or drag & drop a PDF file
              </p>
            </div>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>2️⃣</div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem' }}>Configure</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Set page ranges or max pages per file
              </p>
            </div>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>3️⃣</div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem' }}>Split</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>
                Click "Split PDF" to extract and download
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

export default SplitPage
