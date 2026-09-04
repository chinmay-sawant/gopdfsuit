import { useState, useRef } from 'react'
import { FileText, Download, Upload, Play, RefreshCw, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import { useBundledTemplate } from '../hooks/useBundledTemplate'
import OperationShell from '../components/OperationShell'
import OpPageShell from '../components/OpPageShell'
import ConsentBanner from '../components/ConsentBanner'
import { generatePDFSmart } from '../utils/wasm/generate.js'
import { BUNDLED_TEMPLATES } from '../utils/wasm/templates.js'
import { shouldUseServerWasmTransport } from '../utils/wasm/transports.js'

const serverTransport = shouldUseServerWasmTransport()

const Viewer = () => {
  const [templateData, setTemplateData] = useState('')
  const [fileName, setFileName] = useState('')
  const [error, setError] = useState(null)
  const fileInputRef = useRef(null)
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const { isLoading, resultUrl: pdfUrl, run, runJson, runLocal, download } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => setError(message),
  })
  // Offline-first template load shared with Editor (bundled -> server).
  const { loadTemplateData } = useBundledTemplate({
    runJson,
    getAuthHeaders,
    onError: (message) => setError(`Error loading template: ${message}`),
  })

  const loadTemplate = async (name) => {
    const target = (name ?? fileName).trim()
    if (!target) return
    setError(null)
    setFallbackOffer(null)

    let data
    try {
      data = await loadTemplateData(target)
    } catch (err) {
      setError(`Error loading template: ${err?.message || err}`)
      return
    }
    setTemplateData(JSON.stringify(data, null, 2))

    // Preview render, same WASM-first path as Generate below.
    await renderPreview(data)
  }

  const renderPreview = async (data) => {
    if (serverTransport) {
      await run({
        endpoint: '/api/v1/generate/template-pdf',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
        getAuthHeaders,
        autoDownload: false,
        onError: (message) => setError(`Error loading template: ${message}`),
      })
      return
    }
    let wasmMessage = ''
    const url = await runLocal(() => generatePDFSmart(data, { getAuthHeaders }), {
      autoDownload: false,
      onError: (message) => { wasmMessage = message },
    })
    if (url) return
    if (getAuthHeaders) {
      setFallbackOffer({ message: wasmMessage, data })
    }
  }

  const renderViaServerConsent = async () => {
    if (isLoading || !fallbackOffer) return
    const { data } = fallbackOffer
    setFallbackOffer(null)
    await run({
      endpoint: '/api/v1/generate/template-pdf',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      getAuthHeaders,
      autoDownload: false,
      onError: (message) => setError(`Error generating PDF: ${message}`),
    })
  }

  const generatePDF = async () => {
    if (!templateData.trim()) return
    setError(null)
    setFallbackOffer(null)

    let data
    try {
      data = JSON.parse(templateData)
    } catch {
      setError('Error generating PDF: invalid JSON template')
      return
    }
    if (serverTransport) {
      await run({
        endpoint: '/api/v1/generate/template-pdf',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
        getAuthHeaders,
        autoDownload: false,
        onError: (message) => setError(`Error generating PDF: ${message}`),
      })
      return
    }
    // Browser-local generate via gopdfsuit.wasm (offline once downloaded;
    // engine runs no JS, so output matches the server byte path).
    let wasmMessage = ''
    const url = await runLocal(() => generatePDFSmart(data, { getAuthHeaders }), {
      autoDownload: false,
      onError: (message) => { wasmMessage = message },
    })
    if (url) return
    if (getAuthHeaders) {
      setFallbackOffer({ message: wasmMessage, data })
    }
  }

  const handleFileUpload = (event) => {
    const file = event.target.files[0]
    if (file && file.type === 'application/json') {
      const reader = new FileReader()
      reader.onload = (e) => {
        setTemplateData(e.target.result)
        setFileName(file.name)
      }
      reader.readAsText(file)
    }
  }

  return (
    <OpPageShell
      badge={<><Sparkles size={16} />Template-based PDF Generation</>}
      title="PDF Viewer"
      icon={<div className="feature-icon-box teal" style={{ width: '56px', height: '56px', marginBottom: 0 }}><FileText size={28} /></div>}
      description="Load JSON templates and generate PDFs with live preview"
    >
      <div className="container-full">
        {error && (
          <div style={{
            padding: '1rem',
            background: 'rgba(255, 0, 0, 0.1)',
            border: '1px solid red',
            borderRadius: '8px',
            marginBottom: '1rem',
            color: 'hsl(var(--foreground))',
          }}>
            {error}
          </div>
        )}
        <ConsentBanner offer={fallbackOffer} onConsent={renderViaServerConsent} onDismiss={() => setFallbackOffer(null)} isLoading={isLoading} actionLabel="Upload to server and generate" />
        <div className="grid grid-2" style={{ gap: '2rem' }}>
            {/* Template Input Section */}
            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{
                color: 'hsl(var(--foreground))',
                marginBottom: '1.5rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                fontSize: '1.2rem',
                fontWeight: '700',
              }}>
                <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}>
                  <Upload size={18} />
                </div>
                Template Input
              </h3>

              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{
                  display: 'block',
                  marginBottom: '0.5rem',
                  color: 'hsl(var(--foreground))',
                  fontWeight: '600',
                  fontSize: '0.9rem',
                }}>
                  Load from file:
                </label>
                <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
                  <input
                    type="text"
                    value={fileName}
                    onChange={(e) => setFileName(e.target.value)}
                    placeholder="Enter filename (e.g., temp.json)"
                    style={{
                      flex: 1,
                      padding: '0.75rem 1rem',
                      borderRadius: '8px',
                      border: '1px solid rgba(255, 255, 255, 0.15)',
                      background: 'rgba(255, 255, 255, 0.05)',
                      color: 'hsl(var(--foreground))',
                      fontSize: '0.95rem',
                      transition: 'all 0.2s ease',
                    }}
                  />
                  <button
                    onClick={() => loadTemplate()}
                    disabled={isLoading || !fileName.trim()}
                    className="btn-glow"
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      padding: '0.75rem 1.25rem',
                    }}
                  >
                    {isLoading ? <RefreshCw size={16} className="spin" /> : <Download size={16} />}
                    Load
                  </button>
                </div>

                <div style={{
                  textAlign: 'center',
                  margin: '1.25rem 0',
                  color: 'hsl(var(--muted-foreground))',
                  fontSize: '0.9rem',
                  position: 'relative',
                }}>
                  <span style={{
                    background: 'hsl(var(--background))',
                    padding: '0 1rem',
                    position: 'relative',
                    zIndex: 1,
                  }}>or</span>
                  <div style={{
                    position: 'absolute',
                    top: '50%',
                    left: 0,
                    right: 0,
                    height: '1px',
                    background: 'rgba(255, 255, 255, 0.1)',
                  }} />
                </div>

                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".json"
                  onChange={handleFileUpload}
                  style={{ display: 'none' }}
                />
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="btn-outline-glow"
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '0.5rem',
                  }}
                >
                  <Upload size={16} />
                  Upload JSON File
                </button>
              </div>

              <label style={{
                display: 'block',
                marginBottom: '0.5rem',
                color: 'hsl(var(--foreground))',
                fontWeight: '600',
                fontSize: '0.9rem',
              }}>
                JSON Template:
              </label>
              <textarea
                value={templateData}
                onChange={(e) => setTemplateData(e.target.value)}
                placeholder="Enter or paste your JSON template here..."
                style={{
                  width: '100%',
                  height: '350px',
                  padding: '1rem',
                  borderRadius: '8px',
                  border: '1px solid rgba(255, 255, 255, 0.15)',
                  background: 'rgba(255, 255, 255, 0.05)',
                  color: 'hsl(var(--foreground))',
                  fontSize: '0.9rem',
                  fontFamily: "'SF Mono', 'Monaco', 'Cascadia Code', 'Consolas', monospace",
                  resize: 'vertical',
                  transition: 'all 0.2s ease',
                }}
              />

              <button
                onClick={generatePDF}
                disabled={isLoading || !templateData.trim()}
                className="btn-glow"
                style={{
                  width: '100%',
                  marginTop: '1.5rem',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '0.5rem',
                  padding: '1rem 2rem',
                }}
              >
                {isLoading ? <RefreshCw size={18} className="spin" /> : <Play size={18} />}
                Generate PDF
              </button>
            </div>

            {/* PDF Preview Section: native iframe via OperationShell (3.5:
                one preview stack; the bespoke PdfPreview iframe wrapper is
                deleted and react-pdf stays only for Redact dims). */}
            <OperationShell
              resultUrl={pdfUrl}
              title="PDF Preview"
              icon={<FileText size={18} />}
              emptyTitle="Load a JSON template to start"
              emptySubtitle="Enter template data and click &quot;Generate PDF&quot; to see the preview"
              onDownload={() => download(`template-pdf-${Date.now()}.pdf`)}
              downloadLabel="Download PDF"
              height={550}
              isLoading={isLoading}
              loadingLabel="Generating PDF..."
            />
          </div>

          {/* Sample Templates */}
          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{
              color: 'hsl(var(--foreground))',
              marginBottom: '1.25rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.75rem',
              fontSize: '1.1rem',
              fontWeight: '700',
            }}>
              <div className="feature-icon-box yellow" style={{ width: '40px', height: '40px', marginBottom: 0 }}>
                <span style={{ fontSize: '1.2rem' }}>📋</span>
              </div>
              Sample Templates
            </h3>
            <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
              {BUNDLED_TEMPLATES.map((sample) => (
                <button
                  key={sample}
                  onClick={() => {
                    setFileName(sample)
                    loadTemplate(sample)
                  }}
                  className="btn-outline-glow"
                  style={{
                    fontSize: '0.9rem',
                    padding: '0.75rem 1.25rem',
                  }}
                >
                  {sample}
                </button>
              ))}
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
    </OpPageShell>
  )
}

export default Viewer