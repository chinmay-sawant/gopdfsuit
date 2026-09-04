import { useState, useRef } from 'react'
import { Minimize2, Upload, RefreshCw, FileText, X, Sparkles } from 'lucide-react'
import BackgroundAnimation from '../components/BackgroundAnimation'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import { formatFileSize, resetDropStyles } from '../utils/format'
import { compressPDFSmart, compressViaServer } from '../utils/compressPdf.js'
import { COMPRESS_LEVELS, COMPRESS_TRANSPORT, DEFAULT_COMPRESS_LEVEL, MAX_COMPRESS_BYTES, shouldUseServerCompress } from '../utils/compressLevels.js'

const MAX_COMPRESS_MIB = MAX_COMPRESS_BYTES / (1024 * 1024)
const serverTransport = shouldUseServerCompress()

const CompressPage = () => {
  const [file, setFile] = useState(null)
  const [level, setLevel] = useState(DEFAULT_COMPRESS_LEVEL)
  const [compressedSize, setCompressedSize] = useState(0)
  const [error, setError] = useState(null)
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const fileInputRef = useRef(null)
  const { getAuthHeaders } = useAuth()
  const {
    isLoading,
    resultUrl: compressedPdfUrl,
    runLocal,
    reset: resetOperation,
    download,
  } = usePdfOperation({
    onError: (message) => setError(`Error compressing PDF: ${message}`),
  })

  const clearCompressed = () => {
    resetOperation()
    setCompressedSize(0)
    setFallbackOffer(null)
  }

  const selectPdf = (selectedFile) => {
    if (!selectedFile || isLoading) return
    clearCompressed()
    setFile(selectedFile)
    setError(null)
  }

  const handleFileUpload = (event) => {
    const selectedFile = Array.from(event.target.files).find(f => f.type === 'application/pdf')
    if (selectedFile) selectPdf(selectedFile)
    event.target.value = ''
  }

  const removeFile = () => {
    if (isLoading) return
    setFile(null)
    clearCompressed()
  }

  const compressFile = async () => {
    if (!file || isLoading) return
    if (file.size > MAX_COMPRESS_BYTES) {
      setError(`Error compressing PDF: PDF exceeds maximum size (${MAX_COMPRESS_MIB} MiB)`)
      return
    }
    setError(null)
    setFallbackOffer(null)
    const buf = await file.arrayBuffer()
    const bytes = new Uint8Array(buf)
    let wasmMessage = ''
    const url = await runLocal(() => compressPDFSmart(bytes, { level }, { getAuthHeaders }), {
      autoDownload: false,
      onBlob: (blob) => setCompressedSize(blob.size),
      onError: (message) => { wasmMessage = message },
    })
    if (url) return
    // WASM failed without uploading anything. Offer the server as an
    // explicit consent click instead of silently uploading the file.
    if (!serverTransport && getAuthHeaders) {
      setFallbackOffer({ bytes, message: wasmMessage })
    } else if (wasmMessage) {
      setError(`Error compressing PDF: ${wasmMessage}`)
    }
  }

  const compressViaServerConsent = async () => {
    if (!fallbackOffer || isLoading) return
    setError(null)
    const { bytes } = fallbackOffer
    setFallbackOffer(null)
    await runLocal(() => compressViaServer(bytes, level, getAuthHeaders), {
      autoDownload: false,
      onBlob: (blob) => setCompressedSize(blob.size),
    })
  }

  const downloadCompressed = () => {
    if (!compressedPdfUrl || !file) return
    const originalBase = file.name.replace(/\.pdf$/i, '')
    download(`compressed-${originalBase}-${level}.pdf`)
  }

  const percentSmaller = file && compressedSize > 0 && file.size > 0
    ? Math.max(0, ((file.size - compressedSize) / file.size) * 100)
    : 0

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: 'rgba(78,205,196,0.1)', border: '1px solid rgba(78,205,196,0.3)', borderRadius: '50px', marginBottom: '1.5rem', color: '#4ecdc4', fontSize: '0.9rem', fontWeight: '500' }}>
            <Sparkles size={16} />{serverTransport ? 'Server compression (uploads file)' : 'Runs in your browser'}
          </div>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
            <div className="feature-icon-box teal" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Minimize2 size={28} /></div>
            PDF Compress Tool
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>{serverTransport ? `Server transport active (VITE_COMPRESS_TRANSPORT=${COMPRESS_TRANSPORT}): the file is uploaded to /api/v1/compress.` : 'Shrink PDFs locally with WASM — the file never leaves this device. No upload.'}</p>
        </div>
      </section>

      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          {error && (
            <div style={{ padding: '1rem', background: 'rgba(255, 0, 0, 0.1)', border: '1px solid red', borderRadius: '8px', marginBottom: '1rem', color: 'hsl(var(--foreground))' }}>
              {error}
            </div>
          )}
          {fallbackOffer && (
            <div style={{ padding: '1rem', background: 'rgba(255, 193, 7, 0.1)', border: '1px solid #ffc107', borderRadius: '8px', marginBottom: '1rem', color: 'hsl(var(--foreground))' }}>
              <div style={{ marginBottom: '0.75rem' }}>
                Browser compression failed{fallbackOffer.message ? `: ${fallbackOffer.message}` : '.'} The file was not uploaded.
                Upload it to the server to try server-side compression instead?
              </div>
              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <button onClick={compressViaServerConsent} disabled={isLoading} className="btn-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
                  Upload to server and compress
                </button>
                <button onClick={() => setFallbackOffer(null)} disabled={isLoading} className="btn-outline-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
                  Stay local
                </button>
              </div>
            </div>
          )}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1.5fr', gap: '2rem' }}>
            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
                <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Upload size={18} /></div>Upload PDF File
              </h3>
              <input ref={fileInputRef} type="file" accept=".pdf,application/pdf" onChange={handleFileUpload} disabled={isLoading} style={{ display: 'none' }} />
              <div onClick={() => { if (!isLoading) fileInputRef.current?.click() }} style={{ border: '2px dashed rgba(255,255,255,0.15)', borderRadius: '8px', padding: '3rem 2rem', textAlign: 'center', cursor: isLoading ? 'not-allowed' : 'pointer', transition: 'all 0.3s ease', marginBottom: '2rem', background: 'rgba(255,255,255,0.02)', opacity: isLoading ? 0.6 : 1 }}
                onDragOver={(e) => { e.preventDefault(); if (isLoading) return; e.currentTarget.style.borderColor = '#4ecdc4'; e.currentTarget.style.background = 'rgba(78,205,196,0.1)' }}
                onDragLeave={(e) => { resetDropStyles(e.currentTarget) }}
                onDrop={(e) => { e.preventDefault(); resetDropStyles(e.currentTarget); if (isLoading) return; const f = Array.from(e.dataTransfer.files).find(f => f.type === 'application/pdf'); if (f) selectPdf(f) }}>
                <div className="feature-icon-box teal" style={{ width: '56px', height: '56px', margin: '0 auto 1rem', opacity: 0.6 }}><Upload size={28} /></div>
                <p style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>Click to upload or drag & drop</p>
                <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>Select a PDF — compression stays on this machine</p>
              </div>

              {file && (
                <div>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem', fontSize: '0.95rem', fontWeight: '600' }}>Selected File</h4>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', padding: '0.75rem 1rem', background: 'rgba(78,205,196,0.08)', border: '1px solid rgba(78,205,196,0.2)', borderRadius: '8px', marginBottom: '1.5rem' }}>
                    <FileText size={18} style={{ color: '#4ecdc4' }} />
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontWeight: '500' }}>{file.name}</div>
                      <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>{formatFileSize(file.size)}</div>
                    </div>
                    <button onClick={removeFile} disabled={isLoading} style={{ background: 'rgba(255,100,100,0.1)', border: '1px solid rgba(255,100,100,0.3)', color: '#ff6b6b', borderRadius: '6px', padding: '0.4rem 0.6rem', cursor: isLoading ? 'not-allowed' : 'pointer', opacity: isLoading ? 0.5 : 1 }}><X size={14} /></button>
                  </div>
                  <div style={{ marginBottom: '1rem' }}>
                    <label style={{ display: 'block', marginBottom: '0.75rem', color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>Compression level</label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.75rem' }}>
                      {COMPRESS_LEVELS.map((opt) => {
                        const selected = level === opt.value
                        return (
                          <button key={opt.value} type="button" onClick={() => setLevel(opt.value)} disabled={isLoading}
                            style={{
                              padding: '0.85rem 0.5rem',
                              borderRadius: '8px',
                              border: selected ? '1px solid rgba(78,205,196,0.6)' : '1px solid rgba(255,255,255,0.15)',
                              background: selected ? 'rgba(78,205,196,0.15)' : 'rgba(255,255,255,0.05)',
                              color: 'hsl(var(--foreground))',
                              cursor: isLoading ? 'not-allowed' : 'pointer',
                              opacity: isLoading ? 0.6 : 1,
                              textAlign: 'center',
                            }}>
                            <div style={{ fontWeight: '700', fontSize: '0.95rem', marginBottom: '0.25rem', color: selected ? '#4ecdc4' : 'hsl(var(--foreground))' }}>{opt.name}</div>
                            <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>JPEG {opt.jpeg}</div>
                            <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>max {opt.maxEdge}px</div>
                          </button>
                        )
                      })}
                    </div>
                  </div>
                  <button onClick={compressFile} disabled={isLoading} className="btn-glow" style={{ width: '100%', marginTop: '0.5rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '1rem 2rem', opacity: isLoading ? 0.7 : 1, cursor: isLoading ? 'not-allowed' : 'pointer' }}>
                    {isLoading ? <RefreshCw size={18} className="spin" /> : <Minimize2 size={18} />}{isLoading ? 'Compressing…' : 'Compress PDF'}
                  </button>
                </div>
              )}
            </div>

            <OperationShell
              resultUrl={compressedPdfUrl}
              title="Compressed PDF Preview"
              icon={<FileText size={18} />}
              emptyTitle="Compressed PDF preview will appear here"
              emptySubtitle="Pick a local PDF and compress in the browser"
              onDownload={downloadCompressed}
              downloadLabel="Download Compressed PDF"
              height={550}
              isLoading={isLoading}
              stats={file && (
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.75rem', marginBottom: '1rem' }}>
                  <div style={{ padding: '0.75rem', background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.08)', borderRadius: '8px', textAlign: 'center' }}>
                    <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.25rem' }}>Original</div>
                    <div style={{ fontWeight: '700', color: 'hsl(var(--foreground))', fontSize: '0.95rem' }}>{formatFileSize(file.size)}</div>
                    <div style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))' }}>{file.size.toLocaleString()} bytes</div>
                  </div>
                  <div style={{ padding: '0.75rem', background: 'rgba(78,205,196,0.08)', border: '1px solid rgba(78,205,196,0.2)', borderRadius: '8px', textAlign: 'center' }}>
                    <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.25rem' }}>Compressed</div>
                    <div style={{ fontWeight: '700', color: '#4ecdc4', fontSize: '0.95rem' }}>{formatFileSize(compressedSize)}</div>
                    <div style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))' }}>{compressedSize.toLocaleString()} bytes</div>
                  </div>
                  <div style={{ padding: '0.75rem', background: 'rgba(16,185,129,0.08)', border: '1px solid rgba(16,185,129,0.2)', borderRadius: '8px', textAlign: 'center' }}>
                    <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.25rem' }}>Smaller by</div>
                    <div style={{ fontWeight: '700', color: '#10b981', fontSize: '0.95rem' }}>{percentSmaller.toFixed(1)}%</div>
                    <div style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))' }}>{Math.max(0, file.size - compressedSize).toLocaleString()} bytes</div>
                  </div>
                </div>
              )}
            />
          </div>

          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.1rem', fontWeight: '700' }}>
              <div className="feature-icon-box green" style={{ width: '40px', height: '40px', marginBottom: 0 }}><span style={{ fontSize: '1.2rem' }}>📋</span></div>How to Use
            </h3>
            <div className="grid grid-3" style={{ gap: '1.5rem' }}>
              {[{ num: '1️⃣', title: 'Choose PDF', desc: 'Click or drag & drop a local PDF — nothing is uploaded' },
              { num: '2️⃣', title: 'Pick a level', desc: 'Light, Medium, or Heavy JPEG quality and image size' },
              { num: '3️⃣', title: 'Compress', desc: 'Runs in the browser via WASM. Preview and download.' }].map((step, i) => (
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

export default CompressPage
