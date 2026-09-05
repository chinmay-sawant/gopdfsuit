import { useState } from 'react'
import { Minimize2, Upload, RefreshCw, FileText, X } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import OpPageShell from '../components/OpPageShell'
import FileDropzone from '../components/FileDropzone'
import ConsentBanner from '../components/ConsentBanner'
import { formatFileSize } from '../utils/format'
import { compressPDF, compressViaServer } from '../utils/compressPdf.js'
import { opSmart } from '../utils/wasm/transports.js'
import { COMPRESS_LEVELS, COMPRESS_TRANSPORT, DEFAULT_COMPRESS_LEVEL, MAX_COMPRESS_BYTES, shouldUseServerCompress } from '../utils/compressLevels.js'

const MAX_COMPRESS_MIB = MAX_COMPRESS_BYTES / (1024 * 1024)
const serverTransport = shouldUseServerCompress()

const CompressPage = () => {
  const [file, setFile] = useState(null)
  const [level, setLevel] = useState(DEFAULT_COMPRESS_LEVEL)
  const [compressedSize, setCompressedSize] = useState(0)
  const [error, setError] = useState(null)
  const { getAuthHeaders } = useAuth()
  const {
    isLoading,
    resultUrl: compressedPdfUrl,
    runSmart,
    consentOffer,
    dismissConsent,
    confirmConsentUpload,
    reset: resetOperation,
    download,
  } = usePdfOperation({
    onError: (message) => setError(`Error compressing PDF: ${message}`),
  })

  const clearCompressed = () => {
    resetOperation()
    setCompressedSize(0)
    dismissConsent()
  }

  const selectPdf = (selectedFile) => {
    if (!selectedFile || isLoading) return
    clearCompressed()
    setFile(selectedFile)
    setError(null)
  }

  const handleFileUpload = (selectedFiles) => {
    const selectedFile = selectedFiles.find(f => f.type === 'application/pdf')
    if (selectedFile) selectPdf(selectedFile)
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
    const bytes = new Uint8Array(await file.arrayBuffer())
    // Hook-owned consent: runSmart runs WASM first via opSmart and, on a
    // consentable failure, arms the ConsentBanner below instead of
    // uploading. The server runs only from confirmConsentUpload.
    await runSmart(
      // VITE_COMPRESS_TRANSPORT=server (or the merged VITE_WASM_TRANSPORT)
      // goes straight to the server; otherwise opSmart runs WASM first and
      // the hook arms the consent banner on a consentable failure.
      (transport) => (shouldUseServerCompress()
        ? compressViaServer(bytes, level, transport.getAuthHeaders)
        : opSmart(
          () => compressPDF(bytes, { level }),
          () => compressViaServer(bytes, level, transport.getAuthHeaders),
          transport,
        )),
      {
        getAuthHeaders,
        serverTask: () => compressViaServer(bytes, level, getAuthHeaders),
        serverLabel: 'Upload to server and compress',
        autoDownload: false,
        onBlob: (blob) => setCompressedSize(blob.size),
      },
    )
  }

  const downloadCompressed = () => {
    if (!compressedPdfUrl || !file) return
    const originalBase = file.name.replace(/\.pdf$/i, '')
    download(`compressed-${originalBase}-${level}.pdf`)
  }

  const percentSmaller = file && compressedSize > 0 && file.size > 0
    ? Math.max(0, ((file.size - compressedSize) / file.size) * 100)
    : 0
  // Keep the layout single-column on file select alone. The preview
  // column appears only while generating or after a result exists.
  const hasPreview = Boolean(isLoading || compressedPdfUrl)

  return (
    <OpPageShell
      title="PDF Compress Tool"
      className={`compress-page tool-wide${hasPreview ? '' : ' tool-single-page'}`}
      icon={<div className="feature-icon-box teal" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Minimize2 size={28} /></div>}
      description={serverTransport ? `Server transport active (VITE_COMPRESS_TRANSPORT=${COMPRESS_TRANSPORT}): the file is uploaded to /api/v1/compress.` : 'Shrink PDFs locally with WASM - the file never leaves this device. No upload.'}
      steps={[
        { num: '1️⃣', title: 'Choose PDF', desc: 'Click or drag & drop a local PDF - nothing is uploaded' },
        { num: '2️⃣', title: 'Pick a level', desc: 'Light, Medium, or Heavy JPEG quality and image size' },
        { num: '3️⃣', title: 'Compress', desc: 'Runs in the browser via WASM. Preview and download.' },
      ]}
    >
      {error && (
        <div style={{ padding: '1rem', background: 'rgba(255, 0, 0, 0.1)', border: '1px solid red', borderRadius: '8px', marginBottom: '1rem', color: 'hsl(var(--foreground))' }}>
          {error}
        </div>
      )}
      <ConsentBanner offer={consentOffer} onConsent={confirmConsentUpload} onDismiss={dismissConsent} isLoading={isLoading} actionLabel="Upload to server and compress" />
      <div className={`tool-layout${hasPreview ? '' : ' tool-single'}`}>
        <div className="glass-card" style={{ padding: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
            <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Upload size={18} /></div>Upload PDF File
          </h3>
          <FileDropzone
            onFiles={handleFileUpload}
            disabled={isLoading}
            subtitle="Select a PDF - compression stays on this machine"
          />

              {file && (
                <div style={{ marginTop: '1.5rem' }}>
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
                    {isLoading ? <RefreshCw size={18} className="animate-spin" /> : <Minimize2 size={18} />}{isLoading ? 'Compressing…' : 'Compress PDF'}
                  </button>
                </div>
              )}
            </div>

            {hasPreview && (
              <OperationShell
                resultUrl={compressedPdfUrl}
                title="Compressed PDF Preview"
                icon={<FileText size={18} />}
                emptyTitle="Compressed PDF preview will appear here"
                emptySubtitle="Pick a local PDF and compress in the browser"
                onDownload={downloadCompressed}
                downloadLabel="Download Compressed PDF"
                isLoading={isLoading}
                loadingLabel="Compressing…"
                stats={compressedPdfUrl && file && (
                  <div className="stat-grid">
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
            )}
          </div>
    </OpPageShell>
  )
}

export default CompressPage
