import React, { useState, useRef } from 'react'
import { Merge, Upload, Download, RefreshCw, FileText, X, Sparkles } from 'lucide-react'
import { makeAuthenticatedRequest } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'
import BackgroundAnimation from '../components/BackgroundAnimation'

const MergePage = () => {
  const [files, setFiles] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [mergedPdfUrl, setMergedPdfUrl] = useState('')
  const fileInputRef = useRef(null)
  const { getAuthHeaders, triggerLogin } = useAuth()

  const handleFileUpload = (event) => {
    const newFiles = Array.from(event.target.files).filter(file => file.type === 'application/pdf')
    setFiles(prev => [...prev, ...newFiles])
    event.target.value = ''
  }

  const removeFile = (index) => setFiles(prev => prev.filter((_, i) => i !== index))

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
      files.forEach(file => formData.append('pdf', file))
      const response = await makeAuthenticatedRequest('/api/v1/merge', { method: 'POST', body: formData }, getAuthHeaders)
      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      setMergedPdfUrl(url)
      const link = document.createElement('a')
      link.href = url; link.download = `merged-pdf-${Date.now()}.pdf`; link.click()
    } catch (error) {
      if (error.message.includes("401") || error.message.includes("403")) triggerLogin()
      else alert('Error merging PDFs: ' + error.message)
    } finally { setIsLoading(false) }
  }

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024, sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: 'rgba(240,147,251,0.1)', border: '1px solid rgba(240,147,251,0.3)', borderRadius: '50px', marginBottom: '1.5rem', color: '#f093fb', fontSize: '0.9rem', fontWeight: '500' }}>
            <Sparkles size={16} />Combine Multiple PDFs
          </div>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
            <div className="feature-icon-box purple" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Merge size={28} /></div>
            PDF Merge Tool
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>Combine multiple PDF files with drag-and-drop reordering</p>
        </div>
      </section>

      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          <div className="grid grid-2" style={{ gap: '2rem' }}>
            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
                <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Upload size={18} /></div>Upload PDF Files
              </h3>
              <input ref={fileInputRef} type="file" accept=".pdf" multiple onChange={handleFileUpload} style={{ display: 'none' }} />
              <div onClick={() => fileInputRef.current?.click()} style={{ border: '2px dashed rgba(255,255,255,0.15)', borderRadius: '8px', padding: '3rem 2rem', textAlign: 'center', cursor: 'pointer', transition: 'all 0.3s ease', marginBottom: '2rem', background: 'rgba(255,255,255,0.02)' }}
                onDragOver={(e) => { e.preventDefault(); e.currentTarget.style.borderColor = '#4ecdc4'; e.currentTarget.style.background = 'rgba(78,205,196,0.1)' }}
                onDragLeave={(e) => { e.currentTarget.style.borderColor = 'rgba(255,255,255,0.15)'; e.currentTarget.style.background = 'rgba(255,255,255,0.02)' }}
                onDrop={(e) => { e.preventDefault(); const droppedFiles = Array.from(e.dataTransfer.files).filter(f => f.type === 'application/pdf'); setFiles(prev => [...prev, ...droppedFiles]); e.currentTarget.style.borderColor = 'rgba(255,255,255,0.15)'; e.currentTarget.style.background = 'rgba(255,255,255,0.02)' }}>
                <div className="feature-icon-box teal" style={{ width: '56px', height: '56px', margin: '0 auto 1rem', opacity: 0.6 }}><Upload size={28} /></div>
                <p style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>Click to upload or drag & drop</p>
                <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>Select multiple PDF files to merge</p>
              </div>

              {files.length > 0 && (
                <div>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem', fontSize: '0.95rem', fontWeight: '600' }}>Selected Files ({files.length})</h4>
                  <div style={{ maxHeight: '280px', overflowY: 'auto' }}>
                    {files.map((file, index) => (
                      <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '1rem', padding: '0.75rem 1rem', background: 'rgba(78,205,196,0.08)', border: '1px solid rgba(78,205,196,0.2)', borderRadius: '8px', marginBottom: '0.5rem' }}>
                        <FileText size={18} style={{ color: '#4ecdc4' }} />
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontWeight: '500' }}>{file.name}</div>
                          <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>{formatFileSize(file.size)}</div>
                        </div>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                          <button onClick={() => moveFile(index, 'up')} disabled={index === 0} style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.15)', color: 'hsl(var(--foreground))', borderRadius: '6px', padding: '0.4rem 0.6rem', cursor: index === 0 ? 'not-allowed' : 'pointer', opacity: index === 0 ? 0.4 : 1 }}>↑</button>
                          <button onClick={() => moveFile(index, 'down')} disabled={index === files.length - 1} style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.15)', color: 'hsl(var(--foreground))', borderRadius: '6px', padding: '0.4rem 0.6rem', cursor: index === files.length - 1 ? 'not-allowed' : 'pointer', opacity: index === files.length - 1 ? 0.4 : 1 }}>↓</button>
                          <button onClick={() => removeFile(index)} style={{ background: 'rgba(255,100,100,0.1)', border: '1px solid rgba(255,100,100,0.3)', color: '#ff6b6b', borderRadius: '6px', padding: '0.4rem 0.6rem', cursor: 'pointer' }}><X size={14} /></button>
                        </div>
                      </div>
                    ))}
                  </div>
                  <button onClick={mergePDFs} disabled={isLoading || files.length < 2} className="btn-glow" style={{ width: '100%', marginTop: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '1rem 2rem' }}>
                    {isLoading ? <RefreshCw size={18} className="spin" /> : <Merge size={18} />}Merge PDFs
                  </button>
                </div>
              )}
            </div>

            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
                <div className="feature-icon-box teal" style={{ width: '40px', height: '40px', marginBottom: 0 }}><FileText size={18} /></div>Merged PDF Preview
              </h3>
              {mergedPdfUrl ? (
                <div>
                  <iframe src={mergedPdfUrl} style={{ width: '100%', height: '480px', border: '1px solid rgba(255,255,255,0.15)', borderRadius: '8px' }} title="Merged PDF" />
                  <button onClick={() => { const link = document.createElement('a'); link.href = mergedPdfUrl; link.download = `merged-pdf-${Date.now()}.pdf`; link.click() }} className="btn-glow" style={{ width: '100%', marginTop: '1rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '0.75rem 1.5rem' }}>
                    <Download size={16} />Download Merged PDF
                  </button>
                </div>
              ) : (
                <div style={{ height: '480px', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(255,255,255,0.02)', borderRadius: '8px', border: '2px dashed rgba(255,255,255,0.1)', color: 'hsl(var(--muted-foreground))', textAlign: 'center' }}>
                  <div>
                    <div className="feature-icon-box purple" style={{ width: '64px', height: '64px', margin: '0 auto 1rem', opacity: 0.5 }}><Merge size={32} /></div>
                    <p style={{ marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>Merged PDF preview will appear here</p>
                    <p style={{ fontSize: '0.9rem', opacity: 0.7, marginBottom: 0 }}>Upload at least 2 PDF files to get started</p>
                  </div>
                </div>
              )}
            </div>
          </div>

          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.1rem', fontWeight: '700' }}>
              <div className="feature-icon-box yellow" style={{ width: '40px', height: '40px', marginBottom: 0 }}><span style={{ fontSize: '1.2rem' }}>📋</span></div>How to Use
            </h3>
            <div className="grid grid-3" style={{ gap: '1.5rem' }}>
              {[{ num: '1️⃣', title: 'Upload PDFs', desc: 'Click or drag & drop multiple PDF files' },
              { num: '2️⃣', title: 'Reorder', desc: 'Use ↑↓ buttons to change the order' },
              { num: '3️⃣', title: 'Merge', desc: 'Click "Merge PDFs" to combine and download' }].map((step, i) => (
                <div key={i} style={{ textAlign: 'center', padding: '1rem', background: 'rgba(255,255,255,0.02)', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.05)' }}>
                  <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>{step.num}</div>
                  <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem', fontSize: '1rem' }}>{step.title}</h4>
                  <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.85rem', marginBottom: 0 }}>{step.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </div>
  )
}

export default MergePage