import { useState, useRef } from 'react'
import { FileCheck, Upload, RefreshCw, FileText, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import { formatFileSize } from '../utils/format'
import BackgroundAnimation from '../components/BackgroundAnimation'
import { fillPDFSmart, fillViaServer, shouldUseServerWasmTransport } from '../utils/wasmLoader.js'

const serverTransport = shouldUseServerWasmTransport()

const Filler = () => {
  const [pdfFile, setPdfFile] = useState(null)
  const [xfdfFile, setXfdfFile] = useState(null)
  const pdfInputRef = useRef(null)
  const xfdfInputRef = useRef(null)
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const { isLoading, resultUrl: filledPdfUrl, run, runLocal, download } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error filling PDF: ${message}`),
  })

  const handlePdfUpload = (event) => { const file = event.target.files[0]; if (file?.type === 'application/pdf') setPdfFile(file) }
  const handleXfdfUpload = (event) => { const file = event.target.files[0]; if (file) setXfdfFile(file) }

  const fillPDF = async () => {
    if (!pdfFile || !xfdfFile) return
    // Browser-local first via gopdfsuit.wasm goFillPDF after the 01 Fill
    // binding lands (see plans/wasm/01-full-wasm-port.md); server only on
    // explicit consent, Compress.jsx:83-92 pattern.
    if (serverTransport) {
      const formData = new FormData()
      formData.append('pdf', pdfFile); formData.append('xfdf', xfdfFile)
      await run({
        endpoint: '/api/v1/fill',
        body: formData,
        getAuthHeaders,
        filename: `filled-${pdfFile.name}`,
      })
      return
    }
    setFallbackOffer(null)
    const pdfBytes = new Uint8Array(await pdfFile.arrayBuffer())
    const xfdfBytes = new Uint8Array(await xfdfFile.arrayBuffer())
    const pdfName = pdfFile.name
    let wasmMessage = ''
    const url = await runLocal(() => fillPDFSmart(pdfBytes, xfdfBytes, pdfName, { getAuthHeaders }), {
      autoDownload: false,
      onError: (message) => { wasmMessage = message },
    })
    if (url) return
    if (getAuthHeaders) {
      setFallbackOffer({ pdfBytes, xfdfBytes, pdfName, message: wasmMessage })
    }
  }

  const fillViaServerConsent = async () => {
    if (!fallbackOffer || isLoading) return
    const { pdfBytes, xfdfBytes, pdfName } = fallbackOffer
    setFallbackOffer(null)
    await runLocal(() => fillViaServer(pdfBytes, xfdfBytes, pdfName, getAuthHeaders), {
      filename: `filled-${pdfName}`,
    })
  }

  const FileUploadBox = ({ file, label, onClick }) => (
    <div>
      <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>{label}</label>
      <div onClick={onClick} style={{ border: '2px dashed rgba(255,255,255,0.15)', borderRadius: '8px', padding: '2rem', textAlign: 'center', cursor: 'pointer', transition: 'all 0.3s ease', background: 'rgba(255,255,255,0.02)' }}>
        <FileText size={32} style={{ color: file ? '#4ecdc4' : 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }} />
        <p style={{ color: file ? 'hsl(var(--foreground))' : 'hsl(var(--muted-foreground))', marginBottom: '0.25rem', fontWeight: file ? '500' : '400' }}>{file ? file.name : `Click to upload ${label.split(' ')[0]}`}</p>
        {file && <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem', marginBottom: 0 }}>{formatFileSize(file.size)}</p>}
      </div>
    </div>
  )

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: 'rgba(16,185,129,0.1)', border: '1px solid rgba(16,185,129,0.3)', borderRadius: '50px', marginBottom: '1.5rem', color: '#10b981', fontSize: '0.9rem', fontWeight: '500' }}>
            <Sparkles size={16} />AcroForm Support
          </div>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
            <div className="feature-icon-box green" style={{ width: '56px', height: '56px', marginBottom: 0 }}><FileCheck size={28} /></div>
            PDF Form Filler
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>{serverTransport ? 'Server transport active (VITE_WASM_TRANSPORT=server): files are uploaded to /api/v1/fill.' : 'Fill PDF forms using XFDF data with AcroForm support - runs in your browser when the WASM Fill binding lands, server upload only on consent.'}</p>
        </div>
      </section>

      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          {fallbackOffer && (
            <div style={{ padding: '1rem', background: 'rgba(255, 193, 7, 0.1)', border: '1px solid #ffc107', borderRadius: '8px', marginBottom: '1rem', color: 'hsl(var(--foreground))' }}>
              <div style={{ marginBottom: '0.75rem' }}>
                Browser fill is not available in this build{fallbackOffer.message ? `: ${fallbackOffer.message}` : '.'} The files were not uploaded.
                Upload them to the server to fill instead? Try sampledata/filler/* fixtures.
              </div>
              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <button onClick={fillViaServerConsent} disabled={isLoading} className="btn-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
                  Upload to server and fill
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
                <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Upload size={18} /></div>Upload Files
              </h3>
              <input ref={pdfInputRef} type="file" accept=".pdf" onChange={handlePdfUpload} style={{ display: 'none' }} />
              <input ref={xfdfInputRef} type="file" accept=".xfdf,.xml" onChange={handleXfdfUpload} style={{ display: 'none' }} />

              <div style={{ marginBottom: '1.5rem' }}>
                <FileUploadBox file={pdfFile} label="PDF File (AcroForm):" onClick={() => pdfInputRef.current?.click()} />
              </div>
              <div style={{ marginBottom: '1.5rem' }}>
                <FileUploadBox file={xfdfFile} label="XFDF File (Form Data):" onClick={() => xfdfInputRef.current?.click()} />
              </div>

              <button onClick={fillPDF} disabled={isLoading || !pdfFile || !xfdfFile} className="btn-glow" style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '1rem 2rem' }}>
                {isLoading ? <RefreshCw size={18} className="spin" /> : <FileCheck size={18} />}Fill PDF Form
              </button>
            </div>

            <OperationShell
              resultUrl={filledPdfUrl}
              title="Filled PDF Preview"
              icon={<FileText size={18} />}
              emptyTitle="Filled PDF preview will appear here"
              emptySubtitle="Upload both PDF and XFDF files"
              onDownload={() => download(`filled-${pdfFile?.name || 'form.pdf'}`)}
              downloadLabel="Download Filled PDF"
              height={550}
            />
          </div>

          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.1rem', fontWeight: '700' }}>
              <div className="feature-icon-box yellow" style={{ width: '40px', height: '40px', marginBottom: 0 }}><span style={{ fontSize: '1.2rem' }}>📋</span></div>How to Use
            </h3>
            <div className="grid grid-2" style={{ gap: '2rem', marginBottom: '1.5rem' }}>
              <div>
                <h4 style={{ color: '#4ecdc4', marginBottom: '0.75rem', fontSize: '1rem' }}>Steps:</h4>
                <ol style={{ color: 'hsl(var(--muted-foreground))', lineHeight: 2, paddingLeft: '1.5rem', marginBottom: 0 }}>
                  <li>Upload a PDF file with AcroForm fields</li>
                  <li>Upload an XFDF file containing form data</li>
                  <li>Click &quot;Fill PDF Form&quot; to process</li>
                  <li>Preview and download the filled PDF</li>
                </ol>
              </div>
              <div>
                <h4 style={{ color: '#4ecdc4', marginBottom: '0.75rem', fontSize: '1rem' }}>File Requirements:</h4>
                <ul style={{ color: 'hsl(var(--muted-foreground))', lineHeight: 2, paddingLeft: '1.5rem', marginBottom: 0 }}>
                  <li><strong>PDF:</strong> Must contain AcroForm fields</li>
                  <li><strong>XFDF:</strong> XML file with field data mappings</li>
                  <li>Field names in XFDF must match PDF</li>
                  <li>Supports text, checkboxes, and radios</li>
                </ul>
              </div>
            </div>
            <div style={{ padding: '1rem', background: 'rgba(78,205,196,0.08)', borderRadius: '8px', border: '1px solid rgba(78,205,196,0.2)' }}>
              <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.95rem' }}><FileText size={16} />Sample Files Available</h4>
              <p style={{ color: 'hsl(var(--muted-foreground))', marginBottom: 0, fontSize: '0.9rem' }}>Check the <code style={{ color: '#4ecdc4', background: 'rgba(78,205,196,0.15)', padding: '0.15rem 0.4rem', borderRadius: '4px' }}>sampledata/</code> directory for example files.</p>
            </div>
          </div>
        </div>
      </section>
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </div>
  )
}

export default Filler