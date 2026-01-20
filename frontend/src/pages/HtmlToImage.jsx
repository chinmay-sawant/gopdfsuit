import React, { useState } from 'react'
import { Image, Globe, Download, RefreshCw, Eye, Settings, Sparkles } from 'lucide-react'
import { makeAuthenticatedRequest } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'
import BackgroundAnimation from '../components/BackgroundAnimation'

const HtmlToImage = () => {
  const [htmlContent, setHtmlContent] = useState('')
  const [url, setUrl] = useState('')
  const [inputType, setInputType] = useState('html')
  const [isLoading, setIsLoading] = useState(false)
  const [imageUrl, setImageUrl] = useState('')
  const [showPreview, setShowPreview] = useState(false)
  const { getAuthHeaders, triggerLogin } = useAuth()

  const [config, setConfig] = useState({
    format: 'png', width: 800, height: 600, quality: 94, zoom: 1.0,
    crop_width: 0, crop_height: 0, crop_x: 0, crop_y: 0,
  })

  const convertToImage = async () => {
    if (window.location.href.includes('chinmay-sawant.github.io')) {
      alert("Run the app locally using the dockerfile"); return
    }
    if ((!htmlContent.trim() && inputType === 'html') || (!url.trim() && inputType === 'url')) return
    setIsLoading(true)
    try {
      const requestBody = { ...config, ...(inputType === 'html' ? { html: htmlContent } : { url }) }
      const response = await makeAuthenticatedRequest('/api/v1/htmltoimage', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(requestBody),
      }, getAuthHeaders)
      const blob = await response.blob()
      const imageBlobUrl = URL.createObjectURL(blob)
      setImageUrl(imageBlobUrl)
      const link = document.createElement('a')
      link.href = imageBlobUrl; link.download = `html-to-image-${Date.now()}.${config.format}`; link.click()
    } catch (error) {
      if (error.message.includes("401") || error.message.includes("403")) triggerLogin()
      else alert('Error: ' + error.message)
    } finally { setIsLoading(false) }
  }

  const sampleHtml = `<!DOCTYPE html><html><head><style>body{font-family:Arial;background:#4ecdc4;color:white;text-align:center;padding:50px;margin:0;min-height:100vh;display:flex;flex-direction:column;justify-content:center}h1{font-size:3rem;margin-bottom:1rem}.card{background:rgba(255,255,255,0.1);backdrop-filter:blur(10px);border-radius:15px;padding:2rem}</style></head><body><div class="card"><h1>🎨 Beautiful Image</h1><p>Generated from HTML using GoPdfSuit</p></div></body></html>`

  const inputStyles = { width: '100%', padding: '0.75rem 1rem', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.15)', background: 'rgba(255,255,255,0.05)', color: 'hsl(var(--foreground))', fontSize: '0.95rem' }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
        <div className="container">
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: 'rgba(240,147,251,0.1)', border: '1px solid rgba(240,147,251,0.3)', borderRadius: '50px', marginBottom: '1.5rem', color: '#f093fb', fontSize: '0.9rem', fontWeight: '500' }}>
            <Sparkles size={16} />PNG, JPG, SVG Support
          </div>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
            <div className="feature-icon-box blue" style={{ width: '56px', height: '56px', marginBottom: 0 }}><Image size={28} /></div>
            HTML to Image
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>Convert HTML content or web pages to PNG, JPG, or SVG images</p>
        </div>
      </section>

      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          <div className="grid grid-2" style={{ gap: '2rem' }}>
            <div className="glass-card" style={{ padding: '2rem' }}>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
                <div className="feature-icon-box green" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Globe size={18} /></div>HTML Input
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
                <div className="feature-icon-box purple" style={{ width: '40px', height: '40px', marginBottom: 0 }}><Settings size={18} /></div>Image Configuration
              </h3>
              <div style={{ marginBottom: '2rem' }}>
                <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Format:</label>
                    <select value={config.format} onChange={(e) => setConfig(p => ({ ...p, format: e.target.value }))} style={inputStyles}><option value="png">PNG</option><option value="jpg">JPG</option><option value="svg">SVG</option></select>
                  </div>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Quality (1-100):</label>
                    <input type="number" min="1" max="100" value={config.quality} onChange={(e) => setConfig(p => ({ ...p, quality: parseInt(e.target.value) || 94 }))} style={inputStyles} />
                  </div>
                </div>
                <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Width (px):</label>
                    <input type="number" value={config.width} onChange={(e) => setConfig(p => ({ ...p, width: parseInt(e.target.value) || 800 }))} style={inputStyles} />
                  </div>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Height (px):</label>
                    <input type="number" value={config.height} onChange={(e) => setConfig(p => ({ ...p, height: parseInt(e.target.value) || 600 }))} style={inputStyles} />
                  </div>
                </div>
                <div className="grid grid-2" style={{ gap: '1rem' }}>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Zoom Factor:</label>
                    <input type="number" step="0.1" min="0.1" max="5" value={config.zoom} onChange={(e) => setConfig(p => ({ ...p, zoom: parseFloat(e.target.value) || 1.0 }))} style={inputStyles} />
                  </div>
                  <div>
                    <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>Crop (W×H):</label>
                    <div style={{ display: 'flex', gap: '0.25rem' }}>
                      <input type="number" placeholder="W" value={config.crop_width || ''} onChange={(e) => setConfig(p => ({ ...p, crop_width: parseInt(e.target.value) || 0 }))} style={{ ...inputStyles, flex: 1 }} />
                      <input type="number" placeholder="H" value={config.crop_height || ''} onChange={(e) => setConfig(p => ({ ...p, crop_height: parseInt(e.target.value) || 0 }))} style={{ ...inputStyles, flex: 1 }} />
                    </div>
                  </div>
                </div>
              </div>
              <button onClick={convertToImage} disabled={isLoading || (inputType === 'html' && !htmlContent.trim()) || (inputType === 'url' && !url.trim())} className="btn-glow" style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', marginBottom: '1.5rem', padding: '1rem 2rem' }}>
                {isLoading ? <RefreshCw size={18} className="spin" /> : <Image size={18} />}Convert to Image
              </button>
              {imageUrl && (
                <div>
                  <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.75rem', fontSize: '0.95rem', fontWeight: '600' }}>Image Preview:</h4>
                  <div style={{ border: '1px solid rgba(255,255,255,0.15)', borderRadius: '8px', padding: '1rem', textAlign: 'center', background: 'rgba(255,255,255,0.02)', marginBottom: '1rem' }}>
                    <img src={imageUrl} alt="Generated" style={{ maxWidth: '100%', maxHeight: '280px', borderRadius: '6px', boxShadow: '0 4px 8px rgba(0,0,0,0.3)' }} />
                  </div>
                  <button onClick={() => { const link = document.createElement('a'); link.href = imageUrl; link.download = `html-to-image-${Date.now()}.${config.format}`; link.click() }} className="btn-glow" style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '0.75rem 1.5rem' }}>
                    <Download size={16} />Download Image
                  </button>
                </div>
              )}
            </div>
          </div>

          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.1rem', fontWeight: '700' }}>
              <div className="feature-icon-box yellow" style={{ width: '40px', height: '40px', marginBottom: 0 }}><span style={{ fontSize: '1.2rem' }}>🎨</span></div>Quick Presets
            </h3>
            <div className="grid grid-3" style={{ gap: '1rem' }}>
              <button onClick={() => setConfig(p => ({ ...p, width: 1920, height: 1080, zoom: 1.0 }))} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem' }}>📺 HD (1920×1080)</button>
              <button onClick={() => setConfig(p => ({ ...p, width: 800, height: 600, zoom: 1.0 }))} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem' }}>🖥️ Standard (800×600)</button>
              <button onClick={() => setConfig(p => ({ ...p, width: 400, height: 400, zoom: 1.0 }))} className="btn-outline-glow" style={{ fontSize: '0.9rem', padding: '0.75rem 1rem' }}>🔲 Square (400×400)</button>
            </div>
          </div>
        </div>
      </section>
      <style jsx>{`.spin{animation:spin 1s linear infinite}@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}`}</style>
    </div>
  )
}

export default HtmlToImage