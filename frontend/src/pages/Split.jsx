import { useState } from 'react'
import { Scissors, Upload, RefreshCw, FileText, X, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import OpPageShell from '../components/OpPageShell'
import FileDropzone from '../components/FileDropzone'
import ConsentBanner from '../components/ConsentBanner'
import { formatFileSize } from '../utils/format'
import { splitPDFSmart, splitViaServer, shouldUseServerWasmTransport } from '../utils/wasmLoader.js'

const serverTransport = shouldUseServerWasmTransport()

const SplitPage = () => {
  const [file, setFile] = useState(null)
  const [pages, setPages] = useState('')
  const [maxPerFile, setMaxPerFile] = useState('')
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const { isLoading, resultUrl: splitPdfUrl, run, runLocal, download, reset } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error splitting PDF: ${message}`),
  })

  const removeFile = () => { setFile(null); reset() }

  const splitPDF = async () => {
    if (!file) return
    // Browser-local first via gopdfsuit.wasm goSplitPDF (see
    // plans/wasm/01-full-wasm-port.md); multi-file download from the JS
    // array via runLocal (arrays take the multi path). Server only on
    // explicit consent, offered by the banner below.
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
    const urls = await runLocal(() => splitPDFSmart(bytes, splitOpts, { getAuthHeaders }), {
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
    await runLocal(() => splitViaServer(bytes, splitOpts, getAuthHeaders), {
      filenames: [0, 1, 2, 3, 4, 5, 6, 7].map((i) => `split-${base}-part${i + 1}.pdf`),
      autoDownload: false,
    })
  }

  const inputStyles = { width: '100%', padding: '0.75rem 1rem', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.15)', background: 'rgba(255,255,255,0.05)', color: 'hsl(var(--foreground))', fontSize: '0.95rem' }

  return (
    <OpPageShell
      badge={<><Sparkles size={16} />Extract & Split Pages</>}
      badgeTone="rgba(255,193,7,0.1)"
      badgeBorder="rgba(255,193,7,0.3)"
      badgeColor="#ffc107"
      title="PDF Split Tool"
      icon={<div className="feature-icon-box yellow" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Scissors size={28} /></div>}
      description={serverTransport ? 'Server transport active (VITE_WASM_TRANSPORT=server): the file is uploaded to /api/v1/split.' : 'Extract specific pages or split PDF into multiple files - runs in your browser when the WASM engine lands, server upload only on consent.'}
      steps={[
        { num: '1️⃣', title: 'Upload PDF', desc: 'Click or drag & drop a PDF file' },
        { num: '2️⃣', title: 'Configure', desc: 'Set page ranges or max pages per file' },
        { num: '3️⃣', title: 'Split', desc: 'Click "Split PDF" to extract and download' },
      ]}
    >
      <ConsentBanner offer={fallbackOffer} onConsent={splitViaServerConsent} onDismiss={() => setFallbackOffer(null)} isLoading={isLoading} actionLabel="Upload to server and split" />
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1.5fr', gap: '2rem' }}>
        <div className="glass-card" style={{ padding: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
            <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Upload size={18} /></div>Upload PDF File
          </h3>
          <FileDropzone
            onFiles={(dropped) => { const f = dropped.find(f => f.type === 'application/pdf'); if (f) setFile(f) }}
            subtitle="Select a PDF file to split"
          />

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
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </OpPageShell>
  )
}

export default SplitPage
