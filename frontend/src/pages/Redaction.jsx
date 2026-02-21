
import { useState, useRef } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
import { Upload, Download, Eraser, Trash2, ChevronLeft, ChevronRight, AlertCircle, Check, Search } from 'lucide-react'
import { makeAuthenticatedRequest } from '../utils/apiConfig'
import { useAuth } from '../contexts/AuthContext'
import BackgroundAnimation from '../components/BackgroundAnimation'
import 'react-pdf/dist/Page/AnnotationLayer.css'
import 'react-pdf/dist/Page/TextLayer.css'

// Valid for React-PDF v7/v8/v9. Configure worker.
pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`

const Redaction = () => {
  const [file, setFile] = useState(null)
  const [numPages, setNumPages] = useState(null)
  const [pageNumber, setPageNumber] = useState(1)
  const [pdfPageDims, setPdfPageDims] = useState({ width: 0, height: 0 }) // Authoritative dims from backend
    // const [renderedDims, setRenderedDims] = useState({ width: 0, height: 0 }) // Screen pixels
    const [pageViewport, setPageViewport] = useState({ left: 0, top: 0, width: 0, height: 0 })
  
  // Redactions: Map<pageNum, Array<{x, y, w, h}>> (PDF coordinates)
  const [redactions, setRedactions] = useState({}) 
  
  // Drawing state
  const [isDrawing, setIsDrawing] = useState(false)
  const [startPos, setStartPos] = useState({ x: 0, y: 0 })
  const [currentRect, setCurrentRect] = useState(null) // {x, y, w, h} in pixels
  const [searchText, setSearchText] = useState('')
    const [searchQueries, setSearchQueries] = useState([])
  const [isSearching, setIsSearching] = useState(false)
    const [password, setPassword] = useState('')
    const [mode, setMode] = useState('auto')

  const canvasRef = useRef(null)
  const containerRef = useRef(null)
  const { getAuthHeaders } = useAuth()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState(null)
  const [successMsg, setSuccessMsg] = useState(null)

    const readErrorResponse = async (response, fallbackMessage) => {
        try {
            const text = await response.text()
            if (!text) return fallbackMessage
            try {
                const parsed = JSON.parse(text)
                return parsed?.error || parsed?.message || fallbackMessage
            } catch {
                return text
            }
        } catch {
            return fallbackMessage
        }
    }

    const parseSearchTerms = (raw) => {
        if (!raw) return []
        const seen = new Set()
        return raw
            .split(',')
            .map((term) => term.trim())
            .filter((term) => {
                if (!term) return false
                const key = term.toLowerCase()
                if (seen.has(key)) return false
                seen.add(key)
                return true
            })
    }

  const onDocumentLoadSuccess = ({ numPages }) => {
    setNumPages(numPages)
    setPageNumber(1)
    setError(null)
  }

    const refreshPageViewport = () => {
        if (!containerRef.current) return
        const pageCanvas = containerRef.current.querySelector('canvas')
        if (!pageCanvas) return

        const containerBounds = containerRef.current.getBoundingClientRect()
        const canvasBounds = pageCanvas.getBoundingClientRect()

        if (canvasBounds.width <= 0 || canvasBounds.height <= 0) return

        setPageViewport({
            left: canvasBounds.left - containerBounds.left,
            top: canvasBounds.top - containerBounds.top,
            width: canvasBounds.width,
            height: canvasBounds.height,
        })
            // Removed setRenderedDims call to fix ESLint no-undef error
    }

  const handleFileUpload = async (event) => {
    const selectedFile = event.target.files[0]
        if (selectedFile && selectedFile.type === 'application/pdf') {
            if (selectedFile.size === 0) {
                setError('Selected PDF is empty. Please choose a valid non-empty PDF file.')
                return
            }
      setFile(selectedFile)
      setRedactions({})
      setPageNumber(1)
      setSuccessMsg(null)
      setError(null)
      setPdfPageDims({ width: 0, height: 0 }) // Reset
      
      // Fetch authoritative page info from backend
      try {
        const formData = new FormData()
        formData.append('pdf', selectedFile)
        const response = await makeAuthenticatedRequest('/api/v1/redact/page-info', {
            method: 'POST',
            body: formData
        }, getAuthHeaders)
        
        if (response.ok) {
            const info = await response.json()
            if (info.pages && info.pages.length > 0) {
                // Store all pages? For now just assume they are similar or store map?
                // The current component assumes single page dim for conversion (simplification).
                // Ideally, we should look up dim by pageNumber.
                // Let's store the whole info.
                setPdfPageDims({ 
                    width: info.pages[0].width, 
                    height: info.pages[0].height,
                    allPages: info.pages 
                })
            }
        }
      } catch (e) {
        console.error("Failed to fetch page info", e)
      }

    } else {
      setError('Please select a valid PDF file')
    }
  }

  // Convert screen coordinates (pixels) to PDF coordinates (points)
  const toPDFCoords = (pixelRect) => {
    // Use current page dims
    let currentPageDim = pdfPageDims
    if (pdfPageDims.allPages && pdfPageDims.allPages[pageNumber-1]) {
        currentPageDim = pdfPageDims.allPages[pageNumber-1]
    }

        if (!pageViewport.width || !pageViewport.height || !currentPageDim.width || !currentPageDim.height) {
            return null
        }
    const sx = currentPageDim.width / pageViewport.width
    const sy = currentPageDim.height / pageViewport.height
    
    return {
      x: pixelRect.x * sx,
      y: currentPageDim.height - ((pixelRect.y + pixelRect.height) * sy), // Flip Y
      width: pixelRect.width * sx,
      height: pixelRect.height * sy
    }
  }
  
  // Convert PDF coordinates to screen pixels for rendering existing redactions
  const toScreenCoords = (pdfRect) => {
    let currentPageDim = pdfPageDims
    if (pdfPageDims.allPages && pdfPageDims.allPages[pageNumber-1]) {
        currentPageDim = pdfPageDims.allPages[pageNumber-1]
    }

        if (!pageViewport.width || !pageViewport.height || !currentPageDim.width || !currentPageDim.height) {
            return pdfRect
        }
    const sx = pageViewport.width / currentPageDim.width
    const sy = pageViewport.height / currentPageDim.height
    
    return {
      x: pdfRect.x * sx,
      y: (currentPageDim.height - (pdfRect.y + pdfRect.height)) * sy,
      width: pdfRect.width * sx,
      height: pdfRect.height * sy
    }
  }

  const handleMouseDown = (e) => {
    if (!file) return
        if (!containerRef.current) return
        const rect = containerRef.current.getBoundingClientRect()
        const x = e.clientX - rect.left - pageViewport.left
        const y = e.clientY - rect.top - pageViewport.top

        if (x < 0 || y < 0 || x > pageViewport.width || y > pageViewport.height) {
            return
        }

    setIsDrawing(true)
    setStartPos({ x, y })
    setCurrentRect({ x, y, width: 0, height: 0 })
  }

  const handleMouseMove = (e) => {
    if (!isDrawing) return
        if (!containerRef.current) return
        const rect = containerRef.current.getBoundingClientRect()
        const rawX = e.clientX - rect.left - pageViewport.left
        const rawY = e.clientY - rect.top - pageViewport.top
        const currentX = Math.min(Math.max(rawX, 0), pageViewport.width)
        const currentY = Math.min(Math.max(rawY, 0), pageViewport.height)
    
    const width = currentX - startPos.x
    const height = currentY - startPos.y
    
    setCurrentRect({
      x: width > 0 ? startPos.x : currentX,
      y: height > 0 ? startPos.y : currentY,
      width: Math.abs(width),
      height: Math.abs(height)
    })
  }

  const handleMouseUp = () => {
    if (!isDrawing || !currentRect) return
    setIsDrawing(false)
    
    if (currentRect.width < 5 || currentRect.height < 5) {
      setCurrentRect(null)
      return // Ignore tiny clicks
    }

        const pdfRect = toPDFCoords(currentRect)
        if (!pdfRect) {
            setError('Unable to map coordinates. Wait for the page to fully render and try again.')
            setCurrentRect(null)
            return
        }
    const newRedaction = { ...pdfRect, pageNum: pageNumber }
    
    setRedactions(prev => ({
      ...prev,
      [pageNumber]: [...(prev[pageNumber] || []), newRedaction]
    }))
    setCurrentRect(null)
  }

  const removeRedaction = (pageNum, index) => {
    setRedactions(prev => {
      const newList = [...(prev[pageNum] || [])]
      newList.splice(index, 1)
      return { ...prev, [pageNum]: newList }
    })
  }

  const handleSearch = async () => {
        const terms = parseSearchTerms(searchText)
        if (!file || terms.length === 0) return
        if (file.size === 0) {
            setError('Selected PDF is empty. Please re-upload a valid PDF.')
            return
        }
    setIsSearching(true)
    setError(null)
        setSuccessMsg(null)
    try {
        const formData = new FormData()
        formData.append('pdf', file)
                formData.append('text', terms.join(','))
                formData.append('texts', JSON.stringify(terms))

        const response = await makeAuthenticatedRequest('/api/v1/redact/search', {
            method: 'POST',
            body: formData,
        }, getAuthHeaders)

                if (!response.ok) {
                    const message = await readErrorResponse(response, 'Search failed')
                    throw new Error(message)
                }

                const payload = await response.json()
                const results = Array.isArray(payload)
                    ? payload
                    : (Array.isArray(payload?.rects) ? payload.rects : [])

        if (results && results.length > 0) {
            // Merge new redactions
            setRedactions(prev => {
                const next = { ...prev }
                results.forEach(r => {
                    const p = r.pageNum
                    if (!next[p]) next[p] = []
                    next[p].push(r)
                })
                return next
            })
            setSuccessMsg(`Found and marked ${results.length} occurrence(s) for ${terms.join(', ')}`)
        } else {
            setSuccessMsg(`No occurrences found for ${terms.join(', ')}`)
        }
        setSearchQueries(prev => {
          const existing = new Set(prev.map((x) => x.toLowerCase()))
          const additions = terms.filter((term) => !existing.has(term.toLowerCase()))
          if (additions.length === 0) return prev
          return [...prev, ...additions]
        })
    } catch (err) {
        setError(err.message)
    } finally {
        setIsSearching(false)
    }
  }

  const applyRedactions = async () => {
    // Robust check for empty redactions
        const hasAny = Object.values(redactions).some(arr => arr && arr.length > 0)
        const hasTextCriteria = searchQueries.length > 0 || parseSearchTerms(searchText).length > 0
        if (!file || (!hasAny && !hasTextCriteria)) return
        if (file.size === 0) {
            setError('Selected PDF is empty. Please re-upload a valid PDF.')
            return
        }
    
    setIsLoading(true)
    try {
      // Flatten redactions map to array
      const allRedactions = Object.values(redactions).flat()

            const buildPayload = (modeToUse) => {
                const formData = new FormData()
                formData.append('pdf', file)
                formData.append('blocks', JSON.stringify(allRedactions))
                formData.append('mode', modeToUse)
                if (password.trim()) {
                    formData.append('password', password)
                }

                if (!hasBlockRedactions) {
                    const allSearches = [...searchQueries]
                    const inlineTerms = parseSearchTerms(searchText)
                    allSearches.push(...inlineTerms)
                    const seen = new Set()
                    const uniqueSearches = allSearches.filter((term) => {
                        const key = term.trim().toLowerCase()
                        if (!key || seen.has(key)) return false
                        seen.add(key)
                        return true
                    })

                    if (uniqueSearches.length > 0) {
                        formData.append('textSearch', JSON.stringify(uniqueSearches.map((text) => ({ text }))))
                    }
                }
                return formData
            }

            const modesToTry = mode === 'auto' ? ['secure_required', 'visual_allowed'] : [mode]
            let appliedMode = ''
            let fallbackUsed = false
            let response = null
            let lastError = 'Redaction failed'

            for (let i = 0; i < modesToTry.length; i += 1) {
                const candidateMode = modesToTry[i]
                const tryResponse = await makeAuthenticatedRequest('/api/v1/redact/apply', {
                    method: 'POST',
                    body: buildPayload(candidateMode),
                    throwOnError: false,
                }, getAuthHeaders)

                if (tryResponse.ok) {
                    response = tryResponse
                    appliedMode = candidateMode
                    fallbackUsed = mode === 'auto' && i > 0
                    break
                }

                const message = await readErrorResponse(tryResponse, 'Redaction failed')
                lastError = message

                const canFallback = mode === 'auto' && candidateMode === 'secure_required'
                const secureUnavailable = message.toLowerCase().includes('secure_required requested but no secure text content could be removed')
                if (canFallback && secureUnavailable) {
                    continue
                }

                throw new Error(message)
            }

            if (!response) {
                throw new Error(lastError)
            }

      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `redacted_${file.name}`
      document.body.appendChild(a)
      a.click()
      a.remove()
      if (fallbackUsed) {
        setSuccessMsg('Secure redaction was unavailable for this PDF; visual redaction fallback was applied and downloaded successfully.')
      } else if (appliedMode === 'secure_required') {
        setSuccessMsg('Secure redaction applied and PDF downloaded successfully!')
      } else {
        setSuccessMsg('Visual redaction applied and PDF downloaded successfully!')
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setIsLoading(false)
    }
  }

  // Handle page load to get dimensions
  const onPageLoadSuccess = (page) => {
   // We now rely on backend for "Logical PDF Dimensions" (pdfPageDims).
   // renderedDims is set via onRenderSuccess or calculated here if React-PDF gives rendered size
   // But we stick to setRenderedDims in onRenderSuccess for safety.
   // Fallback if backend fetch failed:
   if (pdfPageDims.width === 0) {
      setPdfPageDims({ width: page.originalWidth, height: page.originalHeight })
   }
  }

    const hasBlockRedactions = Object.values(redactions).some(arr => arr && arr.length > 0)
    const hasTextCriteria = searchQueries.length > 0 || !!searchText.trim()
    const canApply = hasBlockRedactions || hasTextCriteria

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
        <BackgroundAnimation />
        
        <div className="container" style={{ padding: '2rem 1rem', position: 'relative', zIndex: 1 }}>
            <h1 style={{ fontSize: '2rem', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Eraser size={32} /> PDF Redaction
            </h1>

            {!file ? (
                <div className="glass-card" style={{ padding: '3rem', textAlign: 'center' }}>
                    <div style={{ marginBottom: '1.5rem' }}>
                        <Upload size={48} className="text-muted" style={{ opacity: 0.5 }} />
                    </div>
                    <h3 style={{ marginBottom: '1rem' }}>Upload a PDF to redact</h3>
                    <input 
                        type="file" 
                        accept="application/pdf" 
                        onChange={handleFileUpload} 
                        style={{ display: 'none' }} 
                        id="pdf-upload"
                    />
                    <label 
                        htmlFor="pdf-upload" 
                        className="btn-glow"
                        style={{ cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: '0.5rem' }}
                    >
                        <Upload size={18} /> Choose File
                    </label>
                </div>
            ) : (
                <div className="grid" style={{ gridTemplateColumns: '1fr 300px', gap: '2rem', alignItems: 'start' }}>
                    
                    {/* Main Viewer Area */}
                    <div className="glass-card" style={{ padding: '1rem', position: 'relative', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                         <div style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '1rem' }}>
                            <button 
                                onClick={() => setPageNumber(p => Math.max(1, p - 1))} 
                                disabled={pageNumber <= 1}
                                className="btn-outline-glow"
                                style={{ padding: '0.5rem' }}
                            >
                                <ChevronLeft size={20} />
                            </button>
                            <span>Page {pageNumber} of {numPages || '--'}</span>
                            <button 
                                onClick={() => setPageNumber(p => Math.min(numPages || 1, p + 1))} 
                                disabled={pageNumber >= numPages}
                                className="btn-outline-glow"
                                style={{ padding: '0.5rem' }}
                            >
                                <ChevronRight size={20} />
                            </button>
                         </div>

                         <div 
                            ref={containerRef}
                            style={{ position: 'relative', border: '1px solid #ccc', cursor: 'crosshair' }}
                            onMouseDown={handleMouseDown}
                            onMouseMove={handleMouseMove}
                            onMouseUp={handleMouseUp}
                            onMouseLeave={handleMouseUp}
                         >
                            <Document
                                file={file}
                                onLoadSuccess={onDocumentLoadSuccess}
                                loading={<div style={{padding: '2rem'}}>Loading PDF...</div>}
                            >
                                <Page 
                                    pageNumber={pageNumber} 
                                    onLoadSuccess={onPageLoadSuccess}
                                    renderTextLayer={false}
                                    renderAnnotationLayer={false}
                                    width={Math.min(800, window.innerWidth - 100)} // Responsive width
                                                                        onRenderSuccess={() => {
                                                                             // Measure against actual rendered PDF canvas bounds for precise coordinate transforms.
                                                                             refreshPageViewport()
                                    }}
                                />
                            </Document>
                            
                            {/* Overlay for existing redactions */}
                            {redactions[pageNumber]?.map((rect, idx) => {
                                const screenRect = toScreenCoords(rect)
                                return (
                                    <div 
                                        key={idx}
                                        style={{
                                            position: 'absolute',
                                            left: pageViewport.left + screenRect.x,
                                            top: pageViewport.top + screenRect.y,
                                            width: screenRect.width,
                                            height: screenRect.height,
                                            backgroundColor: 'rgba(0, 0, 0, 0.7)',
                                            border: '1px solid red',
                                            pointerEvents: 'none' // Click-through
                                        }}
                                    />
                                )
                            })}

                            {/* Current drawing rect */}
                            {currentRect && (
                                <div style={{
                                    position: 'absolute',
                                    left: pageViewport.left + currentRect.x,
                                    top: pageViewport.top + currentRect.y,
                                    width: currentRect.width,
                                    height: currentRect.height,
                                    backgroundColor: 'rgba(255, 0, 0, 0.3)',
                                    border: '1px solid red',
                                }} />
                            )}
                            
                            {/* Invisible canvas capture layer - simplified: we draw using divs above */}
                                                        <div
                                                            ref={canvasRef}
                                                            style={{
                                                                position: 'absolute',
                                                                top: pageViewport.top,
                                                                left: pageViewport.left,
                                                                width: pageViewport.width,
                                                                height: pageViewport.height,
                                                                zIndex: 10,
                                                            }}
                                                        />
                         </div>
                    </div>

                    {/* Sidebar */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                        <div className="glass-card" style={{ padding: '1.5rem' }}>
                            <h3 style={{ marginBottom: '1rem' }}>Redact by Text</h3>
                             <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
                                <input 
                                    type="text" 
                                    placeholder="Search words (comma-separated)..." 
                                    value={searchText}
                                    onChange={(e) => setSearchText(e.target.value)}
                                    style={{ 
                                        flex: 1, 
                                        padding: '0.5rem', 
                                        borderRadius: '4px', 
                                        border: '1px solid #ddd',
                                        background: 'rgba(255,255,255,0.1)',
                                        color: 'inherit'
                                    }}
                                />
                                <button 
                                    onClick={handleSearch}
                                    disabled={isSearching || !searchText.trim()}
                                    className="btn-outline-glow"
                                    style={{ padding: '0.5rem' }}
                                >
                                    <Search size={20} />
                                </button>
                            </div>
                            <hr style={{ opacity: 0.1, margin: '1rem 0' }} />

                            <h3 style={{ marginBottom: '1rem' }}>Actions</h3>
                                                        <div style={{ marginBottom: '0.75rem' }}>
                                                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.35rem' }}>Mode</label>
                                                            <select
                                                                value={mode}
                                                                onChange={(e) => setMode(e.target.value)}
                                                                style={{ width: '100%', padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd', background: 'rgba(255,255,255,0.1)', color: 'inherit' }}
                                                            >
                                                                <option value="auto">Default (Try Secure, then Visual Fallback)</option>
                                                                <option value="secure_required">Secure Required</option>
                                                                <option value="visual_allowed">Visual Allowed (current engine)</option>
                                                            </select>
                                                        </div>
                                                        <div style={{ marginBottom: '0.75rem' }}>
                                                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.35rem' }}>PDF Password (optional)</label>
                                                            <input
                                                                type="password"
                                                                placeholder="Enter password for encrypted PDFs"
                                                                value={password}
                                                                onChange={(e) => setPassword(e.target.value)}
                                                                style={{ width: '100%', padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd', background: 'rgba(255,255,255,0.1)', color: 'inherit' }}
                                                            />
                                                        </div>
                            <button 
                                onClick={applyRedactions}
                                disabled={isLoading || !canApply}
                                className="btn-glow"
                                style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', marginBottom: '1rem' }}
                            >
                                <Download size={16} /> 
                                {isLoading ? 'Processing...' : 'Apply & Download'}
                            </button>
                            
                            <button 
                                                                onClick={() => {
                                                                    setFile(null)
                                                                    setRedactions({})
                                                                    setSearchQueries([])
                                                                    setSearchText('')
                                                                    setPassword('')
                                                                    setMode('auto')
                                                                }}
                                className="btn-outline-glow"
                                style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}
                            >
                                <Trash2 size={16} /> Reset
                            </button>
                        </div>

                        <div className="glass-card" style={{ padding: '1rem', flex: 1 }}>
                             <h4 style={{ marginBottom: '0.5rem' }}>Redactions on Page {pageNumber}</h4>
                             {(!redactions[pageNumber] || redactions[pageNumber].length === 0) ? (
                                <p className="text-muted" style={{ fontSize: '0.9rem' }}>No redactions on this page.</p>
                             ) : (
                                <ul style={{ listStyle: 'none', padding: 0 }}>
                                    {redactions[pageNumber].map((r, idx) => (
                                        <li key={idx} style={{ 
                                            display: 'flex', 
                                            justifyContent: 'space-between', 
                                            alignItems: 'center',
                                            padding: '0.5rem',
                                            marginBottom: '0.5rem',
                                            background: 'rgba(255,255,255,0.05)',
                                            borderRadius: '4px'
                                        }}>
                                            <span style={{ fontSize: '0.8rem' }}>
                                                Box {idx + 1}: {Math.round(r.width)}x{Math.round(r.height)}
                                            </span>
                                            <button 
                                                onClick={() => removeRedaction(pageNumber, idx)}
                                                style={{ background: 'none', border: 'none', color: '#ff6b6b', cursor: 'pointer' }}
                                            >
                                                <Trash2 size={14} />
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                             )}
                        </div>
                    </div>
                </div>
            )}
            
            {error && (
                <div style={{ marginTop: '1rem', padding: '1rem', background: 'rgba(255, 0, 0, 0.1)', border: '1px solid red', borderRadius: '8px', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <AlertCircle size={20} className="text-danger" />
                    <span>{error}</span>
                </div>
            )}
             {successMsg && (
                <div style={{ marginTop: '1rem', padding: '1rem', background: 'rgba(0, 255, 0, 0.1)', border: '1px solid green', borderRadius: '8px', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <Check size={20} className="text-success" />
                    <span>{successMsg}</span>
                </div>
            )}
        </div>
    </div>
  )
}

export default Redaction
