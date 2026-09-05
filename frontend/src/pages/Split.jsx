import { useEffect, useRef, useState } from 'react'
import { Download, Scissors, Upload, RefreshCw, FileText, X, Archive } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import OpPageShell from '../components/OpPageShell'
import FileDropzone from '../components/FileDropzone'
import ConsentBanner from '../components/ConsentBanner'
import { formatFileSize } from '../utils/format'
import { splitPDFSmart, splitViaServer, shouldUseServerWasmTransport } from '../utils/wasmLoader.js'
import { createZip } from '../utils/zip.js'
import { downloadBlobUrl } from '../hooks/usePdfOperation'

const serverTransport = shouldUseServerWasmTransport()

// Unbounded per-part naming: the old fixed 8-slot array left part 9+ as
// generic result-N.pdf. A 3-page PDF with max 1 per file yields exactly 3
// parts named <base>-part1..3.pdf.
const splitNameFor = (base) => (index) => `split-${base}-part${index + 1}.pdf`

const SplitPage = () => {
  const [file, setFile] = useState(null)
  const [pages, setPages] = useState('')
  const [maxPerFile, setMaxPerFile] = useState('')
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const [isZipping, setIsZipping] = useState(false)
  const resultsRef = useRef(null)
  const announcedCount = useRef(0)
  const { isLoading, resultUrl: splitPdfUrl, resultFiles, run, runLocal, download, downloadResultFile, reset } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error splitting PDF: ${message}`),
  })

  const removeFile = () => { setFile(null); reset() }

  // Bring fresh results on screen: the preview is tall, so without this the
  // count plus Download All panel lands below the fold and users miss it.
  useEffect(() => {
    if (resultFiles.length > 1 && announcedCount.current !== resultFiles.length) {
      announcedCount.current = resultFiles.length
      resultsRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    }
    if (resultFiles.length === 0) announcedCount.current = 0
  }, [resultFiles])

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
    reset()
    const buf = await file.arrayBuffer()
    const bytes = new Uint8Array(buf)
    const splitOpts = { pages, maxPerFile }
    let wasmMessage = ''
    const base = file.name.replace(/\.pdf$/i, '')
    const urls = await runLocal(() => splitPDFSmart(bytes, splitOpts, { getAuthHeaders }), {
      filenames: splitNameFor(base),
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
    reset()
    const base = (file?.name || 'document.pdf').replace(/\.pdf$/i, '')
    await runLocal(() => splitViaServer(bytes, splitOpts, getAuthHeaders), {
      filenames: splitNameFor(base),
      autoDownload: false,
    })
  }

  const downloadAllAsZip = async () => {
    if (resultFiles.length === 0 || isZipping) return
    setIsZipping(true)
    try {
      const entries = []
      for (const resultFile of resultFiles) {
        const data = resultFile.blob
          ? new Uint8Array(await resultFile.blob.arrayBuffer())
          : new Uint8Array(await (await fetch(resultFile.url)).arrayBuffer())
        entries.push({ name: resultFile.filename, data })
      }
      const zipBytes = createZip(entries)
      const blob = new Blob([zipBytes], { type: 'application/zip' })
      const url = URL.createObjectURL(blob)
      try {
        const base = (file?.name || 'document.pdf').replace(/\.pdf$/i, '')
        downloadBlobUrl(url, `split-${base}.zip`)
      } finally {
        setTimeout(() => URL.revokeObjectURL(url), 5000)
      }
    } finally {
      setIsZipping(false)
    }
  }

  const inputStyles = { width: '100%', padding: '0.75rem 1rem', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.15)', background: 'rgba(255,255,255,0.05)', color: 'hsl(var(--foreground))', fontSize: '0.95rem' }
  // Preview column stays hidden on file select alone. It expands only
  // while splitting or after split results exist.
  const hasPreview = Boolean(isLoading || resultFiles.length > 0 || splitPdfUrl)

  return (
    <OpPageShell
      badgeTone="rgba(255,193,7,0.1)"
      badgeBorder="rgba(255,193,7,0.3)"
      badgeColor="#ffc107"
      title="PDF Split Tool"
      className={`split-page tool-wide${hasPreview ? '' : ' tool-single-page'}`}
      icon={<div className="feature-icon-box yellow" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Scissors size={28} /></div>}
      description={serverTransport ? 'Server transport active (VITE_WASM_TRANSPORT=server): the file is uploaded to /api/v1/split.' : 'Extract specific pages or split PDF into multiple files - runs in your browser when the WASM engine lands, server upload only on consent.'}
      steps={[
        { num: '1️⃣', title: 'Upload PDF', desc: 'Click or drag & drop a PDF file' },
        { num: '2️⃣', title: 'Configure', desc: 'Set page ranges or max pages per file' },
        { num: '3️⃣', title: 'Split', desc: 'Click "Split PDF" to extract and download' },
      ]}
    >
      <ConsentBanner offer={fallbackOffer} onConsent={splitViaServerConsent} onDismiss={() => setFallbackOffer(null)} isLoading={isLoading} actionLabel="Upload to server and split" />
      <div className={`split-layout${hasPreview ? '' : ' tool-single'}`}>
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
                    {isLoading ? <RefreshCw size={18} className="spin" /> : <Scissors size={18} />}{isLoading ? 'Splitting…' : 'Split PDF'}
                  </button>
                </div>
              )}
            </div>

            {hasPreview && (
            <div className="split-preview-col">
            {resultFiles.length > 1 && (
              <section ref={resultsRef} className="split-results" aria-labelledby="split-results-title" aria-live="polite">
                <h3 id="split-results-title">Split files ({resultFiles.length})</h3>
                <p>Showing part 1 of {resultFiles.length} in the preview below. Download each part, or grab them all as one zip.</p>
                <div className="split-results-actions">
                  <button onClick={downloadAllAsZip} disabled={isZipping} type="button" className="split-results-zip">
                    {isZipping ? <RefreshCw aria-hidden="true" size={16} className="spin" /> : <Archive aria-hidden="true" size={16} />}
                    {isZipping ? 'Zipping…' : 'Download all (.zip)'}
                  </button>
                </div>
                <div className="split-results-list">
                  {resultFiles.map((resultFile, index) => (
                    <button key={`${resultFile.filename}-${index}`} onClick={() => downloadResultFile(resultFile)} type="button">
                      <Download aria-hidden="true" size={16} />
                      {resultFile.filename}
                    </button>
                  ))}
                </div>
              </section>
            )}
            <OperationShell
              resultUrl={splitPdfUrl}
              title="Split PDF Preview"
              icon={<FileText size={18} />}
              emptyTitle="Split PDF preview will appear here"
              emptySubtitle="Upload a PDF file to get started"
              onDownload={() => download(`split-pdf-${Date.now()}.pdf`)}
              downloadLabel="Download Split PDF"
            />
            </div>
            )}
          </div>
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </OpPageShell>
  )
}

export default SplitPage
