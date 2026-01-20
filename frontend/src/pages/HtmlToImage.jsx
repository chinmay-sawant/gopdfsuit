import React, { useState } from 'react'
import { Image, Globe, Download, RefreshCw, Eye, Settings } from 'lucide-react'
import { makeAuthenticatedRequest } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'
import BackgroundAnimation from '../components/BackgroundAnimation'

const HtmlToImage = () => {
  const [htmlContent, setHtmlContent] = useState('')
  const [url, setUrl] = useState('')
  const [inputType, setInputType] = useState('html') // 'html' or 'url'
  const [isLoading, setIsLoading] = useState(false)
  const [imageUrl, setImageUrl] = useState('')
  const [showPreview, setShowPreview] = useState(false)
  const { getAuthHeaders, triggerLogin } = useAuth()

  // Image Configuration
  const [config, setConfig] = useState({
    format: 'png',
    width: 800,
    height: 600,
    quality: 94,
    zoom: 1.0,
    crop_width: 0,
    crop_height: 0,
    crop_x: 0,
    crop_y: 0,
  })

  const convertToImage = async () => {
    if (window.location.href.includes('chinmay-sawant.github.io')) {
      alert("due to restriction of gcp we can't generate ethis \n run the app locally using the dockerfile")
      return
    }

    if ((!htmlContent.trim() && inputType === 'html') || (!url.trim() && inputType === 'url')) return

    setIsLoading(true)
    try {
      const requestBody = {
        ...config,
        ...(inputType === 'html' ? { html: htmlContent } : { url: url })
      }

      const response = await makeAuthenticatedRequest('/api/v1/htmltoimage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      }, getAuthHeaders)

      const blob = await response.blob()
      const imageBlobUrl = URL.createObjectURL(blob)
      setImageUrl(imageBlobUrl)

      // Also trigger download
      const link = document.createElement('a')
      link.href = imageBlobUrl
      link.download = `html-to-image-${Date.now()}.${config.format}`
      link.click()
    } catch (error) {
      if (error.message.includes("Authentication failed") || error.message.includes("401") || error.message.includes("403") || error.message.includes("Not authenticated")) {
        triggerLogin()
      } else {
        alert('Error converting to image: ' + error.message)
      }
    } finally {
      setIsLoading(false)
    }
  }

  const sampleHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sample Image</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-align: center;
            padding: 50px;
            margin: 0;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
        }
        h1 {
            font-size: 3rem;
            margin-bottom: 1rem;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        p {
            font-size: 1.2rem;
            margin-bottom: 2rem;
            opacity: 0.9;
        }
        .card {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 2rem;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }
    </style>
</head>
<body>
    <div class="card">
        <h1>🎨 Beautiful Image</h1>
        <p>Generated from HTML content using GoPdfSuit</p>
        <p>✨ Gradient backgrounds, shadows, and modern CSS styling!</p>
    </div>
</body>
</html>`

  return (
    <div style={{ padding: '2rem 0', minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <div className="container">
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <h1 style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem',
            color: 'hsl(var(--foreground))',
          }}>
            <Image size={40} />
            HTML to Image Converter
          </h1>
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem' }}>
            Convert HTML content or web pages to PNG, JPG, or SVG images
          </p>
        </div>

        <div className="grid grid-2" style={{ gap: '2rem' }}>
          {/* Input Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Globe size={20} />
              HTML Input
            </h3>

            {/* Input Type Toggle */}
            <div style={{ marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
                <button
                  onClick={() => setInputType('html')}
                  className={`btn ${inputType === 'html' ? 'btn' : 'btn'}`}
                  style={{ opacity: inputType === 'html' ? 1 : 0.6 }}
                >
                  HTML Content
                </button>
                <button
                  onClick={() => setInputType('url')}
                  className={`btn ${inputType === 'url' ? 'btn' : 'btn'}`}
                  style={{ opacity: inputType === 'url' ? 1 : 0.6 }}
                >
                  Website URL
                </button>
              </div>
            </div>

            {inputType === 'html' ? (
              <div>
                <label style={{
                  display: 'block',
                  marginBottom: '0.5rem',
                  color: 'hsl(var(--foreground))',
                  fontWeight: '500',
                }}>
                  HTML Content:
                </label>
                <textarea
                  value={htmlContent}
                  onChange={(e) => setHtmlContent(e.target.value)}
                  placeholder="Enter your HTML content here..."
                  style={{
                    width: '100%',
                    height: '300px',
                    padding: '1rem',
                    borderRadius: '6px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--background))',
                    color: 'hsl(var(--foreground))',
                    fontSize: '0.9rem',
                    fontFamily: 'monospace',
                    resize: 'vertical',
                  }}
                />
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                  <button
                    onClick={() => setHtmlContent(sampleHtml)}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem' }}
                  >
                    Load Sample HTML
                  </button>
                  <button
                    onClick={() => setShowPreview(!showPreview)}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}
                  >
                    <Eye size={14} />
                    {showPreview ? 'Hide' : 'Preview'} HTML
                  </button>
                </div>
              </div>
            ) : (
              <div>
                <label style={{
                  display: 'block',
                  marginBottom: '0.5rem',
                  color: 'hsl(var(--foreground))',
                  fontWeight: '500',
                }}>
                  Website URL:
                </label>
                <input
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com"
                  style={{
                    width: '100%',
                    padding: '0.75rem',
                    borderRadius: '6px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--background))',
                    color: 'hsl(var(--foreground))',
                    fontSize: '1rem',
                    marginBottom: '1rem',
                  }}
                />
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button
                    onClick={() => setUrl('https://example.com')}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem' }}
                  >
                    Example.com
                  </button>
                  <button
                    onClick={() => setUrl('https://github.com/chinmay-sawant/gopdfsuit')}
                    className="btn btn-secondary"
                    style={{ fontSize: '0.9rem' }}
                  >
                    GoPdfSuit GitHub
                  </button>
                </div>
              </div>
            )}

            {/* HTML Preview */}
            {showPreview && inputType === 'html' && htmlContent && (
              <div style={{ marginTop: '1rem' }}>
                <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>HTML Preview:</h4>
                <iframe
                  srcDoc={htmlContent}
                  style={{
                    width: '100%',
                    height: '200px',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                    background: 'white',
                  }}
                  title="HTML Preview"
                />
              </div>
            )}
          </div>

          {/* Configuration & Preview Section */}
          <div className="card">
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Settings size={20} />
              Image Configuration
            </h3>

            <div style={{ marginBottom: '2rem' }}>
              <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Format:
                  </label>
                  <select
                    value={config.format}
                    onChange={(e) => setConfig(prev => ({ ...prev, format: e.target.value }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  >
                    <option value="png">PNG</option>
                    <option value="jpg">JPG</option>
                    <option value="svg">SVG</option>
                  </select>
                </div>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Quality (1-100):
                  </label>
                  <input
                    type="number"
                    min="1"
                    max="100"
                    value={config.quality}
                    onChange={(e) => setConfig(prev => ({ ...prev, quality: parseInt(e.target.value) || 94 }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  />
                </div>
              </div>

              <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Width (px):
                  </label>
                  <input
                    type="number"
                    value={config.width}
                    onChange={(e) => setConfig(prev => ({ ...prev, width: parseInt(e.target.value) || 800 }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  />
                </div>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Height (px):
                  </label>
                  <input
                    type="number"
                    value={config.height}
                    onChange={(e) => setConfig(prev => ({ ...prev, height: parseInt(e.target.value) || 600 }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  />
                </div>
              </div>

              <div className="grid grid-2" style={{ gap: '1rem', marginBottom: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Zoom Factor:
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    min="0.1"
                    max="5"
                    value={config.zoom}
                    onChange={(e) => setConfig(prev => ({ ...prev, zoom: parseFloat(e.target.value) || 1.0 }))}
                    style={{
                      width: '100%',
                      padding: '0.5rem',
                      borderRadius: '4px',
                      border: '1px solid hsl(var(--border))',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))',
                    }}
                  />
                </div>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Crop (W×H):
                  </label>
                  <div style={{ display: 'flex', gap: '0.25rem' }}>
                    <input
                      type="number"
                      placeholder="Width"
                      value={config.crop_width || ''}
                      onChange={(e) => setConfig(prev => ({ ...prev, crop_width: parseInt(e.target.value) || 0 }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                    <input
                      type="number"
                      placeholder="Height"
                      value={config.crop_height || ''}
                      onChange={(e) => setConfig(prev => ({ ...prev, crop_height: parseInt(e.target.value) || 0 }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                  </div>
                </div>
              </div>

              <div className="grid grid-2" style={{ gap: '1rem' }}>
                <div>
                  <label style={{ color: 'hsl(var(--foreground))', fontSize: '0.9rem', display: 'block', marginBottom: '0.25rem' }}>
                    Crop Offset (X,Y):
                  </label>
                  <div style={{ display: 'flex', gap: '0.25rem' }}>
                    <input
                      type="number"
                      placeholder="X"
                      value={config.crop_x || ''}
                      onChange={(e) => setConfig(prev => ({ ...prev, crop_x: parseInt(e.target.value) || 0 }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                    <input
                      type="number"
                      placeholder="Y"
                      value={config.crop_y || ''}
                      onChange={(e) => setConfig(prev => ({ ...prev, crop_y: parseInt(e.target.value) || 0 }))}
                      style={{ flex: 1, padding: '0.5rem', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'end' }}>
                  <button
                    onClick={convertToImage}
                    disabled={isLoading || (inputType === 'html' && !htmlContent.trim()) || (inputType === 'url' && !url.trim())}
                    className="btn"
                    style={{
                      width: '100%',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: '0.5rem',
                    }}
                  >
                    {isLoading ? <RefreshCw size={16} className="spin" /> : <Image size={16} />}
                    Convert to Image
                  </button>
                </div>
              </div>
            </div>

            {/* Image Preview */}
            {imageUrl && (
              <div>
                <h4 style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem' }}>Image Preview:</h4>
                <div style={{
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  padding: '1rem',
                  textAlign: 'center',
                  background: 'hsl(var(--muted))',
                  marginBottom: '1rem',
                }}>
                  <img
                    src={imageUrl}
                    alt="Generated"
                    style={{
                      maxWidth: '100%',
                      maxHeight: '300px',
                      borderRadius: '4px',
                      boxShadow: '0 4px 8px rgba(0, 0, 0, 0.3)',
                    }}
                  />
                </div>
                <button
                  onClick={() => {
                    const link = document.createElement('a')
                    link.href = imageUrl
                    link.download = `html-to-image-${Date.now()}.${config.format}`
                    link.click()
                  }}
                  className="btn btn-primary"
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '0.5rem',
                  }}
                >
                  <Download size={16} />
                  Download Image
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Quick Presets */}
        <div className="card" style={{ marginTop: '2rem' }}>
          <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1rem' }}>🎨 Quick Presets</h3>
          <div className="grid grid-3" style={{ gap: '1rem' }}>
            <button
              onClick={() => setConfig(prev => ({ ...prev, width: 1920, height: 1080, zoom: 1.0 }))}
              className="btn btn-secondary"
              style={{ fontSize: '0.9rem' }}
            >
              📺 HD (1920×1080)
            </button>
            <button
              onClick={() => setConfig(prev => ({ ...prev, width: 800, height: 600, zoom: 1.0 }))}
              className="btn btn-secondary"
              style={{ fontSize: '0.9rem' }}
            >
              🖥️ Standard (800×600)
            </button>
            <button
              onClick={() => setConfig(prev => ({ ...prev, width: 400, height: 400, zoom: 1.0 }))}
              className="btn btn-secondary"
              style={{ fontSize: '0.9rem' }}
            >
              🔲 Square (400×400)
            </button>
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
    </div>
  )
}

export default HtmlToImage