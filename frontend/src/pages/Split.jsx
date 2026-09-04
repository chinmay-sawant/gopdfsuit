import { useState, useRef } from 'react'
import { Scissors, Upload, RefreshCw, FileText, X, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import { formatFileSize } from '../utils/format'
import BackgroundAnimation from '../components/BackgroundAnimation'
import { splitPDFSmart, splitViaServer, shouldUseServerWasmTransport } from '../utils/wasmLoader.js'

const serverTransport = shouldUseServerWasmTransport()

const SplitPage = () => {
  const [file, setFile] = useState(null)
  const [pages, setPages] = useState('')
  const [maxPerFile, setMaxPerFile] = useState('')
  const fileInputRef = useRef(null)
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const { isLoading, resultUrl: splitPdfUrl, run, runLocalMulti, download, reset } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error splitting PDF: ${message}`),
  })

  const handleFileUpload = (event) => {
    const selectedFile = Array.from(event.target.files).find(f => f.type === 'application/pdf')
    if (selectedFile) setFile(selectedFile)
    event.target.value = ''
  }

  const removeFile = () => { setFile(null); reset() }

  const splitPDF = async () => {
    if (!file) return
    // Browser-local first via gopdfsuit.wasm goSplitPDF (see
    // plans/wasm/01-full-wasm-port.md); multi-file download from the JS
    // array via runLocalMulti. Server only on explicit consent,
    // Compress.jsx:83-92 pattern.
    if (serverTransport) {
      const formData = new FormData()
      formData.append('pdf', file)
      if (pages) formData.append('pages', pages)
      if (maxPerFile) formData.append('max_per_file', maxPerFile)
      await run({
        endpoint: '/api/v1/split',
        body: formData,
        getAuthHeaders,
        autoDownload: false,
      })
      return
    }
    setFallbackOffer(null)
    const buf = await file.arrayBuffer()
    const bytes = new Uint8Array(buf)
    const splitOpts = { pages, maxPerFile }
    let wasmMessage = ''
    const base = file.name.replace(/\.pdf$/i, '')
    const urls = await runLocalMulti(() => splitPDFSmart(bytes, splitOpts, { getAuthHeaders }), {
      filenames: [0, 1, 2, 3, 4, 5, 6, 7].map((i) => `split-${base}-part${i + 1}.pdf`),
      autoDownload: false,
      onError: (message) => { wasmMessage = message },
    })
    if (urls) return
    if (getAuthHeaders) {
      setFallbackOffer({ bytes, splitOpts, message: wasmMessage })
    }
  }

  const splitViaServerConsent = async () => {
    if (!fallbackOffer || isLoading) return
    const { bytes, splitOpts } = fallbackOffer
    setFallbackOffer(null)
    const base = (file?.name || 'document.pdf').replace(/\.pdf$/i, '')
    await runLocalMulti(() => splitViaServer(bytes, splitOpts, getAuthHeaders), {
      filenames: [0, 1, 2, 3, 4, 5, 6, 7].map((i) => `split-${base}-part${i + 1}.pdf`),
      autoDownload: false,
    })
  }

  const inputStyles = { width: '100%', padding: '0.75rem 1rem', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.15)', background: 'rgba(255,255,255,0.05)', color: 'hsl(var(--foreground))', fontSize: '0.95rem' }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: 'rgba(255,193,7,0.1)', border: '1px solid rgba(255,193,7,0.3)', borderRadius: '50px', marginBottom: '1.5rem', color: '#ffc107', fontSize: '0.9rem', fontWeight: '500' }}>
            <Sparkles size={16} />Extract & Split Pages
          </div>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
            <div className="feature-icon-box yellow" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Scissors size={28} /></div>
            PDF Split Tool
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>{serverTransport ? 'Server transport active (VITE_WASM_TRANSPORT=server): the file is uploaded to /api/v1/split.' : 'Extract specific pages or split PDF into multiple files - runs in your browser when the WASM engine lands, server upload only on consent.'}</p>
        </div>
      </section>

      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          {fallbackOffer && (
            <div style={{ padding: '1rem', background: 'rgba(255, 193, 7, 0.1)', border: '1px solid #ffc107', borderRadius: '8px', marginBottom: '1rem', color: 'hsl(var(--foreground))' }}>
              <div style={{ marginBottom: '0.75rem' }}>
                Browser split is not available in this build{fallbackOffer.message ? `: ${fallbackOffer.message}` : '.'} The file was not uploaded.
                Upload it to the server to split instead?
              </div>
              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <button onClick={splitViaServerConsent} disabled={isLoading} className="btn-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
                  Upload to server and split
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
              <input ref={fileInputRef} type="file" accept=".pdf" onChange={handleFileUpload} style={{ display: 'none' }} />
              <div onClick={() => fileInputRef.current?.click()} style={{ border: '2px dashed rgba(255,255,255,0.15)', borderRadius: '8px', padding: '3rem 2rem', textAlign: 'center', cursor: 'pointer', transition: 'all 0.3s ease', marginBottom: '2rem', background: 'rgba(255,255,255,0.02)' }}
                onDragOver={(e) => { e.preventDefault(); e.currentTarget.style.borderColor = '#4ecdc4'; e.currentTarget.style.background = 'rgba(78,205,196,0.1)' }}
                onDragLeave={(e) => { e.currentTarget.style.borderColor = 'rgba(255,255,255,0.15)'; e.currentTarget.style.background = 'rgba(255,255,255,0.02)' }}
                onDrop={(e) => { e.preventDefault(); const f = Array.from(e.dataTransfer.files).find(f => f.type === 'application/pdf'); if (f) setFile(f); e.currentTarget.style.borderColor = 'rgba(255,255,255,0.15)'; e.currentTarget.style.background = 'rgba(255,255,255,0.02)' }}>
                <div className="feature-icon-box teal" style={{ width: '56px', height: '56px', margin: '0 auto 1rem', opacity: 0.6 }}><Upload size={28} /></div>
                <p style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>Click to upload or drag & drop</p>
                <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>Select a PDF file to split</p>
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
                    <button onClick={removeFile} style={{ background: 'rgba(255,100,100,0.1)', border: '1px solid rgba(255,100,100,0.3)', color: '#ff6b6b', borderRadius: '6px', padding: '0.4rem 0.6rem', cursor: 'pointer' }}><X size={14} /></button>
                  </div>
                  <div style={{ marginBottom: '1rem' }}>
                    <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>Pages (optional)</label>
                    <input type="text" value={pages} onChange={(e) => setPages(e.target.value)} placeholder="e.g. 1-3,5" style={inputStyles} disabled={isLoading} />
                    <small style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>Specify page ranges or leave empty</small>
                  </div>
                  <div style={{ marginBottom: '1rem' }}>
                    <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>Max per file (optional)</label>
                    <input type="number" min="1" value={maxPerFile} onChange={(e) => setMaxPerFile(e.target.value)} placeholder="e.g. 1" style={inputStyles} disabled={isLoading} />
                    <small style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>Split into files with this many pages each</small>
                  </div>
                  <button onClick={splitPDF} disabled={isLoading} className="btn-glow" style={{ width: '100%', marginTop: '0.5rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '1rem 2rem' }}>
                    {isLoading ? <RefreshCw size={18} className="spin" /> : <Scissors size={18} />}Split PDF
                  </button>
                </div>
              )}
            </div>

            <OperationShell
              resultUrl={splitPdfUrl}
              title="Split PDF Preview"
              icon={<FileText size={18} />}
              emptyTitle="Split PDF preview will appear here"
              emptySubtitle="Upload a PDF file to get started"
              onDownload={() => download(`split-pdf-${Date.now()}.pdf`)}
              downloadLabel="Download Split PDF"
              height={550}
            />
          </div>

          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.1rem', fontWeight: '700' }}>
              <div className="feature-icon-box green" style={{ width: '40px', height: '40px', marginBottom: 0 }}><span style={{ fontSize: '1.2rem' }}>📋</span></div>How to Use
            </h3>
            <div className="grid grid-3" style={{ gap: '1.5rem' }}>
              {[{ num: '1️⃣', title: 'Upload PDF', desc: 'Click or drag & drop a PDF file' },
              { num: '2️⃣', title: 'Configure', desc: 'Set page ranges or max pages per file' },
              { num: '3️⃣', title: 'Split', desc: 'Click "Split PDF" to extract and download' }].map((step, i) => (
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

export default SplitPage
