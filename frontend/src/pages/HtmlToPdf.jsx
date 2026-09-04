import { useState } from 'react'
import { Globe, FileText, RefreshCw, Eye, Settings, Sparkles } from 'lucide-react'
import { isGitHubPagesHost, OFFLINE_DEMO_MESSAGE } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import OperationShell from '../components/OperationShell'
import BackgroundAnimation from '../components/BackgroundAnimation'

const HtmlToPdf = () => {
  const [htmlContent, setHtmlContent] = useState('')
  const [url, setUrl] = useState('')
  const [inputType, setInputType] = useState('html')
  const [showPreview, setShowPreview] = useState(false)
  const { getAuthHeaders, triggerLogin } = useAuth()
  const { isLoading, resultUrl: pdfUrl, run, download } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error: ${message}`),
  })

  const [config, setConfig] = useState({
    page_size: 'A4', orientation: 'Portrait',
    margin_top: '10mm', margin_right: '10mm', margin_bottom: '10mm', margin_left: '10mm',
    grayscale: false,
  })

  const convertToPdf = async () => {
    // [~] server-only per plans/wasm/03 Phase 3: Chromium today, gowkhtmltopdf
    // after plans/wasm/02-gowkhtmltopdf-replace.md; never WASM.
    if (isGitHubPagesHost()) {
      alert(OFFLINE_DEMO_MESSAGE)
      return
    }
    if ((!htmlContent.trim() && inputType === 'html') || (!url.trim() && inputType === 'url')) return
    const requestBody = { ...config, ...(inputType === 'html' ? { html: htmlContent } : { url }) }
    await run({
      endpoint: '/api/v1/htmltopdf',
      body: JSON.stringify(requestBody),
      headers: { 'Content-Type': 'application/json' },
      getAuthHeaders,
      filename: `html-to-pdf-${Date.now()}.pdf`,
    })
  }

  const sampleHtml = `<!DOCTYPE html><html><head><style>body{font-family:Arial;padding:20px}h1{color:#2c3e50;border-bottom:2px solid #3498db}.highlight{background:#4ecdc4;color:white;padding:15px;border-radius:8px;margin:20px 0}</style></head><body><h1>Sample PDF</h1><p>Sample document for conversion.</p><div class="highlight"><h3>Highlighted</h3><p>This section has styling.</p></div></body></html>`

  const inputStyles = { width: '100%', padding: '0.75rem 1rem', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.15)', background: 'rgba(255,255,255,0.05)', color: 'hsl(var(--foreground))', fontSize: '0.95rem' }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: 'rgba(0,122,204,0.1)', border: '1px solid rgba(0,122,204,0.3)', borderRadius: '50px', marginBottom: '1.5rem', color: '#007acc', fontSize: '0.9rem', fontWeight: '500' }}>
            <Sparkles size={16} />Pure-Go conversion
          </div>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
            <div className="feature-icon-box green" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Globe size={28} /></div>
            HTML to PDF
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>Convert HTML content or web pages to PDF</p>
        </div>
      </section>

      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          <div className="grid grid-2" style={{ gap: '2rem' }}>
            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
                <div className="feature-icon-box blue" style={{ width: '40px', height: '40px', marginBottom: 0 }}><FileText size={18} /></div>HTML Input
              </h3>
              <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', background: 'rgba(255,255,255,0.05)', padding: '0.25rem', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.1)' }}>
                {['html', 'url'].map(t => (
                  <button key={t} onClick={() => setInputType(t)} style={{ flex: 1, padding: '0.75rem', borderRadius: '6px', border: 'none', background: inputType === t ? 'rgba(78,205,196,0.2)' : 'transparent', color: inputType === t ? '#4ecdc4' : 'hsl(var(--muted-foreground))', fontWeight: '600', cursor: 'pointer' }}>
                    {t === 'html' ? 'HTML Content' : 'Website URL'}
                  </button>
                ))}
              </div>
              {inputType === 'html' ? (
                <div>
                  <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>HTML Content:</label>
                  <textarea value={htmlContent} onChange={(e) => setHtmlContent(e.target.value)} placeholder="Enter HTML..." style={{ ...inputStyles, height: '280px', fontFamily: 'monospace', fontSize: '0.9rem', resize: 'vertical' }} />
                  <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem' }}>
                    <button onClick={() => setHtmlContent(sampleHtml)} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem' }}>Load Sample</button>
                    <button onClick={() => setShowPreview(!showPreview)} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}><Eye size={14} />{showPreview ? 'Hide' : 'Preview'}</button>
                  </div>
                </div>
              ) : (
                <div>
                  <label style={{ display: 'block', marginBottom: '0.5rem', color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>Website URL:</label>
                  <input type="url" value={url} onChange={(e) => setUrl(e.target.value)} placeholder="https://example.com" style={{ ...inputStyles, marginBottom: '1rem' }} />
                  <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <button onClick={() => setUrl('https://example.com')} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem' }}>Example.com</button>
                    <button onClick={() => setUrl('https://github.com/chinmay-sawant/gopdfsuit')} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem' }}>GitHub</button>
                  </div>
                </div>
              )}
              {showPreview && inputType === 'html' && htmlContent && (
                <div style={{ marginTop: '1.5rem' }}>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.75rem', fontSize: '0.95rem', fontWeight: '600' }}>Preview:</h4>
                  <iframe srcDoc={htmlContent} style={{ width: '100%', height: '200px', border: '1px solid rgba(255,255,255,0.15)', borderRadius: '8px', background: 'white' }} title="Preview" />
                </div>
              )}
            </div>

            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
                <div className="feature-icon-box purple" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Settings size={18} /></div>PDF Configuration
              </h3>
              <div style={{ marginBottom: '2rem' }}>
                <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Page Size:</label>
                    <select value={config.page_size} onChange={(e) => setConfig(p => ({ ...p, page_size: e.target.value }))} style={inputStyles}><option value="A4">A4</option><option value="Letter">Letter</option><option value="Legal">Legal</option><option value="A3">A3</option></select>
                  </div>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Orientation:</label>
                    <select value={config.orientation} onChange={(e) => setConfig(p => ({ ...p, orientation: e.target.value }))} style={inputStyles}><option value="Portrait">Portrait</option><option value="Landscape">Landscape</option></select>
                  </div>
                </div>
                <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Margins (Top/Bottom):</label>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <input type="text" value={config.margin_top} onChange={(e) => setConfig(p => ({ ...p, margin_top: e.target.value }))} style={{ ...inputStyles, flex: 1 }} />
                      <input type="text" value={config.margin_bottom} onChange={(e) => setConfig(p => ({ ...p, margin_bottom: e.target.value }))} style={{ ...inputStyles, flex: 1 }} />
                    </div>
                  </div>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Margins (Left/Right):</label>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <input type="text" value={config.margin_left} onChange={(e) => setConfig(p => ({ ...p, margin_left: e.target.value }))} style={{ ...inputStyles, flex: 1 }} />
                      <input type="text" value={config.margin_right} onChange={(e) => setConfig(p => ({ ...p, margin_right: e.target.value }))} style={{ ...inputStyles, flex: 1 }} />
                    </div>
                  </div>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', justifyContent: 'center' }}>
                  <label style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', color: 'hsl(var(--foreground))', fontSize: '0.9rem', cursor: 'pointer' }}>
                    <input type="checkbox" checked={config.grayscale} onChange={(e) => setConfig(p => ({ ...p, grayscale: e.target.checked }))} style={{ width: '18px', height: '18px', accentColor: '#4ecdc4' }} />Grayscale
                  </label>
                  <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem', margin: 0 }}>Pure-Go engine: dpi / low_quality are no-ops and hidden. Scripts are stripped, flex/grid support is partial, backgrounds ignored, WOFF2 fonts skipped.</p>
                </div>
              </div>
              <button onClick={convertToPdf} disabled={isLoading || (inputType === 'html' && !htmlContent.trim()) || (inputType === 'url' && !url.trim())} className="btn-glow" style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', marginBottom: '1.5rem', padding: '1rem 2rem' }}>
                {isLoading ? <RefreshCw size={18} className="spin" /> : <FileText size={18} />}Convert to PDF
              </button>
              {pdfUrl && (
                <div style={{ marginTop: '1.5rem' }}>
                  <OperationShell
                    resultUrl={pdfUrl}
                    title="PDF Preview"
                    icon={<FileText size={18} />}
                    emptyTitle=""
                    emptySubtitle=""
                    onDownload={() => download(`html-to-pdf-${Date.now()}.pdf`)}
                    downloadLabel="Download PDF"
                    height={280}
                  />
                </div>
              )}
            </div>
          </div>
        </div>
      </section>
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </div>
  )
}

export default HtmlToPdf