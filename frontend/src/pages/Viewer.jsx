import { useState, useRef } from 'react'
import { FileText, Download, Upload, Play, RefreshCw } from 'lucide-react'
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
      title="PDF Viewer"
      className="viewer-page tool-wide"
      icon={<div className="feature-icon-box teal" style={{ width: '56px', height: '56px', marginBottom: 0 }}><FileText size={28} /></div>}
      description="Load JSON templates and generate PDFs with live preview"
    >
      <div className="container-full viewer-fit">
        {error && (
          <div style={{
            padding: '0.6rem 0.9rem',
            background: 'rgba(255, 0, 0, 0.1)',
            border: '1px solid red',
            borderRadius: '8px',
            marginBottom: '0.6rem',
            color: 'hsl(var(--foreground))',
          }}>
            {error}
          </div>
        )}
        <ConsentBanner offer={fallbackOffer} onConsent={renderViaServerConsent} onDismiss={() => setFallbackOffer(null)} isLoading={isLoading} actionLabel="Upload to server and generate" />
        {/* Sample templates sit above the fold so no scroll is needed to reach them. */}
        <div className="glass-card viewer-samples">
          <span className="viewer-samples-label">Samples</span>
          <div className="viewer-samples-list">
            {BUNDLED_TEMPLATES.map((sample) => (
              <button
                key={sample}
                onClick={() => {
                  setFileName(sample)
                  loadTemplate(sample)
                }}
                className="btn-outline-glow viewer-sample-btn"
                type="button"
              >
                {sample}
              </button>
            ))}
          </div>
        </div>
        <div className="viewer-grid">
          {/* Template Input Section */}
          <div className="glass-card viewer-input-card">
            <h3 className="viewer-card-title">
              <div className="feature-icon-box blue" style={{ width: '32px', height: '32px', marginBottom: 0 }}>
                <Upload size={15} />
              </div>
              Template Input
            </h3>

            <div className="viewer-file-row">
              <input
                type="text"
                value={fileName}
                onChange={(e) => setFileName(e.target.value)}
                placeholder="Enter filename (e.g., temp.json)"
                aria-label="Template filename"
              />
              <button
                onClick={() => loadTemplate()}
                disabled={isLoading || !fileName.trim()}
                className="btn-glow viewer-load-btn"
                type="button"
              >
                {isLoading ? <RefreshCw size={15} className="spin" /> : <Download size={15} />}
                Load
              </button>
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
              className="btn-outline-glow viewer-upload-btn"
              type="button"
            >
              <Upload size={15} />
              Upload JSON File
            </button>

            <label className="viewer-json-label" htmlFor="viewer-json-template">
              JSON Template:
            </label>
            <textarea
              id="viewer-json-template"
              className="viewer-json"
              value={templateData}
              onChange={(e) => setTemplateData(e.target.value)}
              placeholder="Enter or paste your JSON template here..."
            />

            <button
              onClick={generatePDF}
              disabled={isLoading || !templateData.trim()}
              className="btn-glow viewer-generate-btn"
              type="button"
            >
              {isLoading ? <RefreshCw size={16} className="spin" /> : <Play size={16} />}
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
            height={480}
            isLoading={isLoading}
            loadingLabel="Generating PDF..."
          />
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
