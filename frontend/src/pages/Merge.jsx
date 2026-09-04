import { useState } from 'react'
import { Merge, Upload, RefreshCw, FileText, X, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import OpPageShell from '../components/OpPageShell'
import FileDropzone from '../components/FileDropzone'
import ConsentBanner from '../components/ConsentBanner'
import { formatFileSize } from '../utils/format'
import { mergePDFSmart, mergeViaServer, shouldUseServerWasmTransport } from '../utils/wasmLoader.js'

const serverTransport = shouldUseServerWasmTransport()

const MergePage = () => {
  const [files, setFiles] = useState([])
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const { isLoading, resultUrl: mergedPdfUrl, run, runLocal, download } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error merging PDFs: ${message}`),
  })

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
    // Browser-local first via gopdfsuit.wasm goMergePDF (see
    // plans/wasm/01-full-wasm-port.md); server only on explicit consent,
    // Compress.jsx:83-92 pattern. Until the engine lands the WASM call throws
    // missingEngine and the consent banner below offers the upload.
    if (serverTransport) {
      const formData = new FormData()
      files.forEach(file => formData.append('pdf', file))
      await run({
        endpoint: '/api/v1/merge',
        body: formData,
        getAuthHeaders,
        filename: `merged-pdf-${Date.now()}.pdf`,
      })
      return
    }
    setFallbackOffer(null)
    let wasmMessage = ''
    const snapshot = [...files]
    const url = await runLocal(() => mergePDFSmart(snapshot, {}, { getAuthHeaders }), {
      autoDownload: false,
      onError: (message) => { wasmMessage = message },
    })
    if (url) return
    if (!serverTransport && getAuthHeaders) {
      setFallbackOffer({ message: wasmMessage })
    }
    if (url) return
  }

  const mergeViaServerConsent = async () => {
    if (isLoading) return
    setFallbackOffer(null)
    const snapshot = [...files]
    // Explicit consent click: upload to the server via runLocal so blob
    // handling stays identical to the WASM path.
    await runLocal(() => mergeViaServer(snapshot, getAuthHeaders), {
      filename: `merged-pdf-${Date.now()}.pdf`,
    })
  }

  return (
    <OpPageShell
      badge={<><Sparkles size={16} />Combine Multiple PDFs</>}
      badgeTone="rgba(240,147,251,0.1)"
      badgeBorder="rgba(240,147,251,0.3)"
      badgeColor="#f093fb"
      title="PDF Merge Tool"
      icon={<div className="feature-icon-box purple" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Merge size={28} /></div>}
      description={serverTransport ? 'Server transport active (VITE_WASM_TRANSPORT=server): files are uploaded to /api/v1/merge.' : 'Combine multiple PDF files with drag-and-drop reordering - runs in your browser when the WASM engine lands, server upload only on consent.'}
      steps={[
        { num: '1️⃣', title: 'Upload PDFs', desc: 'Click or drag & drop multiple PDF files' },
        { num: '2️⃣', title: 'Reorder', desc: 'Use ↑↓ buttons to change the order' },
        { num: '3️⃣', title: 'Merge', desc: 'Click "Merge PDFs" to combine and download' },
      ]}
    >
      <ConsentBanner offer={fallbackOffer} onConsent={mergeViaServerConsent} onDismiss={() => setFallbackOffer(null)} isLoading={isLoading} actionLabel="Upload to server and merge" />
      <div className="grid grid-2" style={{ gap: '2rem' }}>
        <div className="glass-card" style={{ padding: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
            <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Upload size={18} /></div>Upload PDF Files
          </h3>
          <FileDropzone
            multiple
            onFiles={(dropped) => setFiles(prev => [...prev, ...dropped.filter(f => f.type === 'application/pdf')])}
            subtitle="Select multiple PDF files to merge"
          />

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

            <OperationShell
              resultUrl={mergedPdfUrl}
              title="Merged PDF Preview"
              icon={<FileText size={18} />}
              emptyTitle="Merged PDF preview will appear here"
              emptySubtitle="Upload at least 2 PDF files to get started"
              onDownload={() => download(`merged-pdf-${Date.now()}.pdf`)}
              downloadLabel="Download Merged PDF"
              height={480}
            />
          </div>
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </OpPageShell>
  )
}

export default MergePage