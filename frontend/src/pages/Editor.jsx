
import React, { useState, useRef, useMemo, useEffect } from 'react'
import { Upload, Play, Sun, Moon, Eye, Download, Info, Copy, Check, Edit } from 'lucide-react'
import { useTheme } from '../theme'
import { useAuth } from '../contexts/AuthContext'
import { makeAuthenticatedRequest, isAuthRequired } from '../utils/apiConfig'
import PdfPreview from '../components/PdfPreview'

// Imported Components
import ComponentList from '../components/editor/ComponentList'
import DocumentSettings from '../components/editor/DocumentSettings'
import PropertiesPanel from '../components/editor/PropertiesPanel'
import JsonTemplate from '../components/editor/JsonTemplate'
import ComponentItem from '../components/editor/ComponentItem'
import { PAGE_SIZES, DEFAULT_FONTS, COMPONENT_TYPES } from '../components/editor/constants'
import { getUsableWidth, getFontFamily, MARGIN } from '../components/editor/utils'

import Toolbar from '../components/editor/Toolbar'

export default function Editor() {
  const { theme, setTheme } = useTheme()
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [config, setConfig] = useState({ pageBorder: '1:1:1:1', page: 'A4', pageAlignment: 1, watermark: '', pdfaCompliant: true, signature: { enabled: false } })
  const [title, setTitle] = useState(null)
  const [components, setComponents] = useState([]) // Combined ordered array for tables and spacers
  const [footer, setFooter] = useState(null)
  const [bookmarks, setBookmarks] = useState(null) // PDF outlines/bookmarks
  const [selectedId, setSelectedId] = useState(null)
  const [selectedCell, setSelectedCell] = useState(null)
  const [draggedType, setDraggedType] = useState(null)
  const [isDragOver, setIsDragOver] = useState(false)
  const [draggedComponentId, setDraggedComponentId] = useState(null)
  const [pdfUrl, setPdfUrl] = useState(null)
  const [showPreviewModal, setShowPreviewModal] = useState(false)
  const [fonts, setFonts] = useState(DEFAULT_FONTS)
  const [fontsLoading, setFontsLoading] = useState(true)
  const [copiedId, setCopiedId] = useState(null)
  const [templateInput, setTemplateInput] = useState('')
  const canvasRef = useRef(null)

  // Fetch fonts from API on component mount
  useEffect(() => {
    const fetchFonts = async () => {
      try {
        setFontsLoading(true)
        const response = await makeAuthenticatedRequest(
          '/api/v1/fonts',
          {},
          isAuthRequired() ? getAuthHeaders : null
        )
        if (response.ok) {
          const data = await response.json()
          if (data.fonts && Array.isArray(data.fonts)) {
            setFonts(data.fonts)
          }
        } else {
          console.warn('Failed to fetch fonts, using defaults')
        }
      } catch (error) {
        console.error('Error fetching fonts:', error)
      } finally {
        setFontsLoading(false)
      }
    }
    fetchFonts()
  }, [])

  // Get all elements in order for display
  const allElements = useMemo(() => {
    const elements = []
    if (title) elements.push({ ...title, id: 'title', type: 'title' })
    components.forEach((component, idx) => {
      if (component.type === 'table') {
        elements.push({ ...component, id: `table-${idx}`, type: 'table' })
      } else if (component.type === 'spacer') {
        elements.push({ ...component, id: `spacer-${idx}`, type: 'spacer' })
      } else if (component.type === 'image') {
        elements.push({ ...component, id: `image-${idx}`, type: 'image' })
      }
    })
    if (footer) elements.push({ ...footer, id: 'footer', type: 'footer' })
    return elements
  }, [title, components, footer])

  const selectedElement = allElements.find(el => el.id === selectedId) || null
  const selectedCellElement = selectedElement && selectedCell && selectedElement.type === 'table'
    ? selectedElement.rows[selectedCell.rowIdx].row[selectedCell.colIdx]
    : null

  const currentPageSize = PAGE_SIZES[config.page] || PAGE_SIZES.A4

  // --- Handlers ---
  const handleDropElement = (type, targetId = null) => {
    if (type === 'title') {
      if (!title) setTitle({
        props: 'Helvetica:12:000:left:1:1:1:1',
        text: 'Document Title',
        textprops: 'Helvetica:18:100:center:1:1:1:1',
        table: {
          maxcolumns: 3,
          columnwidths: [1, 2, 1],
          rows: [{
            row: [
              { props: 'Helvetica:12:000:left:1:1:1:1', text: '', image: null },
              { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' },
              { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }
            ]
          }]
        }
      })
    } else if (type === 'footer') {
      if (!footer) setFooter({ props: 'Helvetica:10:000:center:1:0:0:0', text: 'Page footer text' })
    } else {
      const newComponent = type === 'table'
        ? {
          type: 'table',
          maxcolumns: 3,
          rows: [
            { row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }] },
            { row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }] },
            { row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }] }
          ]
        }
        : type === 'image'
          ? { type: 'image', width: 200, height: 150, imagedata: null, imagename: '' }
          : { type: 'spacer', height: 20 }

      if (targetId) {
        // Insert before target
        const targetIndex = components.findIndex((c, i) =>
          targetId.startsWith('table-') ? `table-${i}` === targetId :
            targetId.startsWith('spacer-') ? `spacer-${i}` === targetId :
              `image-${i}` === targetId
        )
        if (targetIndex !== -1) {
          const newComponents = [...components]
          newComponents.splice(targetIndex, 0, newComponent)
          setComponents(newComponents)
        } else {
          setComponents([...components, newComponent])
        }
      } else {
        setComponents([...components, newComponent])
      }
    }
  }

  const handleDelete = (id) => {
    if (id === 'title') setTitle(null)
    else if (id === 'footer') setFooter(null)
    else {
      const idx = parseInt(id.split('-')[1])
      setComponents(components.filter((_, i) => i !== idx))
      if (selectedId === id) setSelectedId(null)
    }
  }

  const handleUpdate = (id, updates) => {
    if (id === 'title') setTitle({ ...title, ...updates })
    else if (id === 'footer') setFooter({ ...footer, ...updates })
    else {
      const idx = parseInt(id.split('-')[1])
      const newComponents = [...components]
      newComponents[idx] = { ...newComponents[idx], ...updates }
      setComponents(newComponents)
    }
  }

  const handleCellDrop = (element, onUpdate, rowIdx, colIdx, type) => {
    const defaultProps = 'Helvetica:12:000:left:0:0:0:0'
    const newRows = [...element.rows]
    const currentCell = newRows[rowIdx].row[colIdx]

    let newCellData = { ...currentCell }

    if (type === 'checkbox') {
      newCellData = { props: defaultProps, chequebox: true, text: undefined, image: undefined, form_field: undefined }
    } else if (type === 'text_input') {
      newCellData = { props: defaultProps, form_field: { name: `field_${Date.now()}`, value: '', type: 'text' }, text: undefined, image: undefined, chequebox: undefined }
    } else if (type === 'radio') {
      newCellData = { props: defaultProps, form_field: { name: `radio_${Date.now()}`, checked: false, type: 'radio' }, text: undefined, image: undefined, chequebox: undefined }
    } else if (type === 'image') {
      newCellData = { props: defaultProps, image: { imagename: '', imagedata: null, width: 100, height: 80 }, text: undefined, chequebox: undefined, form_field: undefined }
    } else if (type === 'hyperlink') {
      newCellData = { props: defaultProps, text: 'Link Text', link: 'https://example.com', image: undefined, chequebox: undefined, form_field: undefined }
    }

    newRows[rowIdx].row[colIdx] = newCellData
    onUpdate({ rows: newRows })
  }

  const handleMove = (index, direction) => {
    const newComponents = [...components]
    if (direction === 'up' && index > 0) {
      [newComponents[index], newComponents[index - 1]] = [newComponents[index - 1], newComponents[index]]
      const currentId = components[index].type === 'table' ? `table-${index}` : components[index].type === 'image' ? `image-${index}` : `spacer-${index}`
      if (selectedId === currentId) {
        const nextId = newComponents[index - 1].type === 'table' ? `table-${index - 1}` : newComponents[index - 1].type === 'image' ? `image-${index - 1}` : `spacer-${index - 1}`
        setSelectedId(nextId)
      }
    } else if (direction === 'down' && index < components.length - 1) {
      [newComponents[index], newComponents[index + 1]] = [newComponents[index + 1], newComponents[index]]
      const currentId = components[index].type === 'table' ? `table-${index}` : components[index].type === 'image' ? `image-${index}` : `spacer-${index}`
      if (selectedId === currentId) {
        const nextId = newComponents[index + 1].type === 'table' ? `table-${index + 1}` : newComponents[index + 1].type === 'image' ? `image-${index + 1}` : `spacer-${index + 1}`
        setSelectedId(nextId)
      }
    }
    setComponents(newComponents)
  }

  // --- JSON Handling ---
  const [jsonText, setJsonText] = useState('')
  const [isJsonEditing, setIsJsonEditing] = useState(false)

  useEffect(() => {
    if (isJsonEditing) return
    const template = {
      config: config,
      title: title,
      elements: components.map(c => {
        if (c.type === 'table') return { type: 'table', table: c }
        if (c.type === 'spacer') return { type: 'spacer', spacer: c }
        if (c.type === 'image') return { type: 'image', image: c }
        return c
      }),
      footer: footer,
      bookmarks: bookmarks
    }
    if (!title) delete template.title
    if (!footer) delete template.footer
    if (!bookmarks || bookmarks.length === 0) delete template.bookmarks
    setJsonText(JSON.stringify(template, null, 2))
  }, [config, title, components, footer, bookmarks, isJsonEditing])

  const handleJsonChange = (e) => setJsonText(e.target.value)

  const handleJsonBlur = () => {
    setIsJsonEditing(false)
    try {
      const parsed = JSON.parse(jsonText)
      const { config: newConfig, title: newTitle, elements, table, spacer, content, footer: newFooter, bookmarks: newBookmarks } = parsed

      setConfig(prev => ({ ...prev, ...(newConfig || {}) }))
      setTitle(newTitle || null)

      // Handle various input formats (legacy content, table, or new elements)
      let rawComponents = elements || content || []

      // If there's a separate table array (raw tables format), process it
      if (table && Array.isArray(table)) {
        rawComponents = table.map(t => ({ ...t, type: 'table' }))
      }

      // If there's a separate spacer array, add those too
      if (spacer && Array.isArray(spacer)) {
        const spacerComponents = spacer.map(s => ({ ...s, type: 'spacer' }))
        rawComponents = [...rawComponents, ...spacerComponents]
      }

      // If we have an "elements" array that references indices, process that
      if (parsed.elements && Array.isArray(parsed.elements) && parsed.elements[0]?.index !== undefined) {
        // This is the reference format: elements: [{type: 'table', index: 0}, ...]
        const orderedComponents = []
        for (const ref of parsed.elements) {
          if (ref.type === 'table' && table && table[ref.index]) {
            orderedComponents.push({ ...table[ref.index], type: 'table' })
          } else if (ref.type === 'spacer' && spacer && spacer[ref.index]) {
            orderedComponents.push({ ...spacer[ref.index], type: 'spacer' })
          }
        }
        if (orderedComponents.length > 0) {
          rawComponents = orderedComponents
        }
      }

      const processedComponents = rawComponents.map(c => {
        // If it's the wrapped format (element.table), unwrap it
        if (c.table) return { ...c.table, type: 'table' }
        if (c.spacer) return { ...c.spacer, type: 'spacer' }
        if (c.image) return { ...c.image, type: 'image' }

        // Auto-detect component type if not specified
        if (!c.type) {
          if (c.maxcolumns && c.rows) return { ...c, type: 'table' }
          if (c.height && !c.width) return { ...c, type: 'spacer' }
          if (c.imagedata || c.imagename) return { ...c, type: 'image' }
        }

        return c
      })

      setComponents(Array.isArray(processedComponents) ? processedComponents : [])
      setFooter(newFooter || null)
      setBookmarks(newBookmarks || null)
    } catch (e) {
      console.error('Invalid JSON', e)
    }
  }

  // --- PDF Generation ---
  const handleGeneratePdf = async (isPreview = false) => {
    try {
      setIsJsonEditing(false)
      const template = {
        config: config,
        title: title,
        elements: components.map(c => {
          if (c.type === 'table') return { type: 'table', table: c }
          if (c.type === 'spacer') return { type: 'spacer', spacer: c }
          if (c.type === 'image') return { type: 'image', image: c }
          return c
        }),
        footer: footer,
        bookmarks: bookmarks
      }
      if (!title) delete template.title
      if (!footer) delete template.footer
      if (!bookmarks || bookmarks.length === 0) delete template.bookmarks

      const response = await makeAuthenticatedRequest(
        '/api/v1/generate/template-pdf',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/pdf'
          },
          body: JSON.stringify(template)
        },
        isAuthRequired() ? getAuthHeaders : null
      )

      console.log('PDF Generation Response Status:', response.status)

      if (!response.ok) {
        if (response.status === 401) { triggerLogin(); return }
        const errorText = await response.text()
        throw new Error(`Failed to generate PDF: ${response.status} - ${errorText}`)
      }

      const blob = await response.blob()
      console.log('PDF Blob:', { size: blob.size, type: blob.type })

      if (blob.size === 0) {
        throw new Error('Received empty PDF document')
      }

      const url = URL.createObjectURL(blob)

      if (isPreview) {
        setPdfUrl(url)
        setShowPreviewModal(true)
      } else {
        const link = document.createElement('a')
        link.href = url
        link.download = 'generated_document.pdf'
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
      }
    } catch (err) {
      console.error(err)
      alert(err.message)
    }
  }

  const handlePreviewPdf = () => handleGeneratePdf(true)

  const handleCopyJson = async () => {
    try {
      await navigator.clipboard.writeText(jsonText)
      setCopiedId('json')
      setTimeout(() => setCopiedId(null), 2000)
    } catch (error) {
      console.error('Copy failed:', error)
    }
  }

  // --- File Upload ---
  const onLoadTemplate = async (filename) => {
    if (!filename || !filename.trim()) {
      alert('Please enter a template filename')
      return
    }

    try {
      // Make GET request to fetch the template
      const response = await makeAuthenticatedRequest(
        `/api/v1/template-data?file=${encodeURIComponent(filename)}`,
        {
          method: 'GET',
          headers: {
            'Accept': 'application/json'
          }
        },
        isAuthRequired() ? getAuthHeaders : null
      )

      if (!response.ok) {
        if (response.status === 401) {
          triggerLogin()
          return
        }
        if (response.status === 404) {
          throw new Error(`Template "${filename}" not found`)
        }
        throw new Error(`Failed to load template: ${response.status}`)
      }

      const templateData = await response.json()

      // Parse and load the template data
      const { config: newConfig, title: newTitle, elements, table, spacer, content, footer: newFooter, bookmarks: newBookmarks } = templateData

      setConfig(prev => ({ ...prev, ...(newConfig || {}) }))
      setTitle(newTitle || null)

      // Handle various input formats (legacy content, table, or new elements)
      let rawComponents = elements || content || []

      // If there's a separate table array (raw tables format), process it
      if (table && Array.isArray(table)) {
        rawComponents = table.map(t => ({ ...t, type: 'table' }))
      }

      // If there's a separate spacer array, add those too
      if (spacer && Array.isArray(spacer)) {
        const spacerComponents = spacer.map(s => ({ ...s, type: 'spacer' }))
        rawComponents = [...rawComponents, ...spacerComponents]
      }

      // If we have an "elements" array that references indices, process that
      if (templateData.elements && Array.isArray(templateData.elements) && templateData.elements[0]?.index !== undefined) {
        // This is the reference format: elements: [{type: 'table', index: 0}, ...]
        const orderedComponents = []
        for (const ref of templateData.elements) {
          if (ref.type === 'table' && table && table[ref.index]) {
            orderedComponents.push({ ...table[ref.index], type: 'table' })
          } else if (ref.type === 'spacer' && spacer && spacer[ref.index]) {
            orderedComponents.push({ ...spacer[ref.index], type: 'spacer' })
          }
        }
        if (orderedComponents.length > 0) {
          rawComponents = orderedComponents
        }
      }

      const processedComponents = rawComponents.map(c => {
        // If it's the wrapped format (element.table), unwrap it
        if (c.table) return { ...c.table, type: 'table' }
        if (c.spacer) return { ...c.spacer, type: 'spacer' }
        if (c.image) return { ...c.image, type: 'image' }

        // Auto-detect component type if not specified
        if (!c.type) {
          if (c.maxcolumns && c.rows) return { ...c, type: 'table' }
          if (c.height && !c.width) return { ...c, type: 'spacer' }
          if (c.imagedata || c.imagename) return { ...c, type: 'image' }
        }

        return c
      })

      setComponents(Array.isArray(processedComponents) ? processedComponents : [])
      setFooter(newFooter || null)
      setBookmarks(newBookmarks || null)

      // Update JSON display
      setIsJsonEditing(false)

      // Clear selection
      setSelectedId(null)
      setSelectedCell(null)

    } catch (error) {
      console.error('Error loading template:', error)
      alert(error.message || 'Failed to load template')
    }
  }

  // --- Keyboard Shortcuts ---
  useEffect(() => {
    const handleKeyDown = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        const element = document.createElement('a')
        const file = new Blob([jsonText], { type: 'application/json' })
        element.href = URL.createObjectURL(file)
        element.download = 'template.json'
        document.body.appendChild(element)
        element.click()
        document.body.removeChild(element)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [jsonText])

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      height: '100vh',
      background: 'hsl(var(--background))',
      color: 'hsl(var(--foreground))',
      fontFamily: getFontFamily('Helvetica'),
      overflow: 'hidden'
    }}>
      {/* Header / Toolbar */}
      <div style={{ padding: '0 1.5rem', paddingTop: '1rem' }}>
        <Toolbar
          theme={theme}
          setTheme={setTheme}
          onLoadTemplate={onLoadTemplate}
          onPreviewPDF={handlePreviewPdf}
          onCopyJSON={handleCopyJson}
          onDownloadPDF={handleGeneratePdf}
          templateInput={templateInput}
          setTemplateInput={setTemplateInput}
          copiedId={copiedId}
        />
      </div>

      {/* Main Content using CSS Grid */}
      <div style={{
        flex: 1,
        display: 'grid',
        gridTemplateColumns: '280px minmax(600px, 1fr) 320px',
        gap: '1.5rem',
        padding: '0 1.5rem 1.5rem 1.5rem',
        overflow: 'hidden',
        height: 'calc(100vh - 84px)' // Approximate header height
      }}>

        {/* Left Column: Settings and Components */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', overflowY: 'auto' }}>
          {/* We merge Settings and Components into the left column to match typical 3-col layout */}
          <ComponentList draggedType={draggedType} setDraggedType={setDraggedType} />
          <DocumentSettings config={config} setConfig={setConfig} currentPageSize={currentPageSize} />
        </div>

        {/* Center Column: Canvas */}
        <div style={{
          background: 'hsl(var(--muted))',
          borderRadius: '8px',
          padding: '1.5rem',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          position: 'relative',
          boxShadow: 'inset 0 0 10px rgba(0,0,0,0.05)'
        }}>
          {/* Size Display Chip */}
          <div style={{
            background: 'hsl(var(--card))',
            padding: '0.25rem 0.75rem',
            borderRadius: '12px',
            fontSize: '0.8rem',
            marginBottom: '1rem',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
            color: 'hsl(var(--foreground))',
            border: '1px solid hsl(var(--border))',
            zIndex: 10
          }}>
            {currentPageSize.name} - {currentPageSize.width} × {currentPageSize.height} pts
          </div>

          <div
            style={{
              flex: 1,
              width: '100%',
              overflow: 'auto',
              display: 'flex',
              justifyContent: 'center',
              padding: '2rem 0.5rem',
              background: 'hsl(var(--muted) / 0.3)'
            }}
          >
            <div
              ref={canvasRef}
              style={{
                width: `${currentPageSize.width + 40}px`,
                minHeight: `${currentPageSize.height}px`,
                background: isDragOver ? 'repeating-linear-gradient(45deg, hsl(var(--accent)) 0px, hsl(var(--accent)) 2px, transparent 2px, transparent 20px)' : 'white',
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                padding: `${MARGIN}px`,
                position: 'relative',
                display: 'flex',
                flexDirection: 'column',
                gap: '0px',
                border: isDragOver ? '2px dashed var(--secondary-color)' : '1px solid #e5e5e5',
                transition: 'all 0.2s ease',
                color: '#000'
              }}
              onDragOver={(e) => { e.preventDefault(); setIsDragOver(true) }}
              onDragLeave={() => setIsDragOver(false)}
              onDrop={(e) => {
                e.preventDefault(); setIsDragOver(false)
                const type = e.dataTransfer.getData('text/plain')
                if (COMPONENT_TYPES[type]) handleDropElement(type)
              }}
              onClick={() => { setSelectedId(null); setSelectedCell(null) }}
            >
              {/* Background Grid - only at top and left edge */}
              <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '20px', background: 'repeating-linear-gradient(90deg, transparent, transparent 49px, #f0f0f0 50px)', pointerEvents: 'none', opacity: 0.5 }} />
              <div style={{ position: 'absolute', top: 0, left: 0, height: '100%', width: '20px', background: 'repeating-linear-gradient(0deg, transparent, transparent 49px, #f0f0f0 50px)', pointerEvents: 'none', opacity: 0.5 }} />

              {/* Page Break Indicators (every page height) */}
              {Array.from({ length: Math.floor((canvasRef.current?.scrollHeight || currentPageSize.height) / currentPageSize.height) }).map((_, i) => i > 0 && (
                <div
                  key={`page-break-${i}`}
                  style={{
                    position: 'absolute',
                    top: `${i * currentPageSize.height}px`,
                    left: 0,
                    right: 0,
                    height: '2px',
                    background: 'repeating-linear-gradient(90deg, #e74c3c 0px, #e74c3c 10px, transparent 10px, transparent 20px)',
                    pointerEvents: 'none',
                    zIndex: 5,
                    opacity: 0.5
                  }}
                >
                  <span style={{
                    position: 'absolute',
                    right: '10px',
                    top: '-10px',
                    fontSize: '10px',
                    background: '#e74c3c',
                    color: 'white',
                    padding: '2px 6px',
                    borderRadius: '3px'
                  }}>
                    Page {i + 1}
                  </span>
                </div>
              ))}

              {/* Page Border (only for first page to avoid complexity) */}
              {config.pageBorder && config.pageBorder !== '0:0:0:0' && (
                <div style={{
                  position: 'absolute',
                  top: MARGIN,
                  left: MARGIN,
                  width: `calc(100% - ${2 * MARGIN}px)`,
                  height: `${currentPageSize.height - 2 * MARGIN}px`,
                  pointerEvents: 'none',
                  borderLeft: `${config.pageBorder.split(':')[0]}px solid #000`,
                  borderRight: `${config.pageBorder.split(':')[1]}px solid #000`,
                  borderTop: `${config.pageBorder.split(':')[2]}px solid #000`,
                  borderBottom: `${config.pageBorder.split(':')[3]}px solid #000`,
                  zIndex: 0
                }} />
              )}
              {/* Watermark */}
              {config.watermark && (
                <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%) rotate(-45deg)', fontSize: '64px', opacity: 0.1, color: '#000', pointerEvents: 'none', whiteSpace: 'nowrap', zIndex: 0 }}>
                  {config.watermark}
                </div>
              )}

              {/* Render Elements */}
              {allElements.map((element, index) => {
                // Calculate the actual component index for move operations
                let componentIndex = -1
                if (element.type !== 'title' && element.type !== 'footer') {
                  componentIndex = parseInt(element.id.split('-')[1])
                }
                const canMoveUp = componentIndex > 0
                const canMoveDown = componentIndex >= 0 && componentIndex < components.length - 1

                return (
                  <ComponentItem
                    key={element.id}
                    element={element}
                    index={componentIndex >= 0 ? componentIndex : index}
                    isSelected={selectedId === element.id}
                    onSelect={setSelectedId}
                    onUpdate={(updates) => handleUpdate(element.id, updates)}
                    onMove={handleMove}
                    onDelete={handleDelete}
                    canMoveUp={canMoveUp}
                    canMoveDown={canMoveDown}
                    selectedCell={selectedCell}
                    onCellSelect={setSelectedCell}
                    onDragStart={setDraggedComponentId}
                    onDragEnd={() => setDraggedComponentId(null)}
                    onDrop={handleDropElement}
                    isDragging={draggedComponentId === element.id}
                    draggedType={draggedType}
                    handleCellDrop={handleCellDrop}
                    currentPageSize={currentPageSize}
                  />
                )
              })}

              {allElements.length === 0 && (
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', color: '#999', border: '2px dashed #eee', borderRadius: '8px', margin: '2rem' }}>
                  <p>Drop components here</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Column: Properties and JSON */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', overflowY: 'auto' }}>
          <PropertiesPanel
            selectedElement={selectedElement}
            selectedCell={selectedCell}
            selectedCellElement={selectedCellElement}
            updateElement={handleUpdate}
            deleteElement={handleDelete}
            setSelectedCell={setSelectedCell}
            currentPageSize={currentPageSize}
            fonts={fonts}
          />
          <JsonTemplate
            jsonText={jsonText}
            handleJsonChange={handleJsonChange}
            setIsJsonEditing={setIsJsonEditing}
            handleJsonBlur={handleJsonBlur}
            copiedId={copiedId}
            setCopiedId={setCopiedId}
          />
        </div>
      </div>

      {/* Preview Modal */}
      {showPreviewModal && (
        <div style={{
          position: 'fixed',
          top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.8)',
          zIndex: 100,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '2rem'
        }} onClick={() => setShowPreviewModal(false)}>
          <div style={{
            width: '80%',
            height: '90%',
            background: 'hsl(var(--card))',
            borderRadius: '12px',
            padding: '1.5rem',
            display: 'flex',
            flexDirection: 'column',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
            border: '1px solid hsl(var(--border))'
          }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
              <h3 style={{ margin: 0, color: 'hsl(var(--foreground))' }}>PDF Preview</h3>
              <button
                onClick={() => setShowPreviewModal(false)}
                style={{
                  padding: '0.5rem 1rem',
                  background: 'hsl(var(--muted))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  color: 'hsl(var(--foreground))',
                  cursor: 'pointer',
                  fontWeight: '500',
                  transition: 'all 0.2s'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'hsl(var(--accent))'
                  e.currentTarget.style.borderColor = 'hsl(var(--accent-foreground))'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'hsl(var(--muted))'
                  e.currentTarget.style.borderColor = 'hsl(var(--border))'
                }}
              >
                Close
              </button>
            </div>
            <div style={{ flex: 1, background: '#525659', overflow: 'hidden', borderRadius: '8px' }}>
              <PdfPreview pdfUrl={pdfUrl} />
            </div>
          </div>
        </div>
      )}

      <style jsx>{`
        .dragging {
          transform: rotate(3deg) scale(0.95);
        }
        
        /* Custom Scrollbar Styles */
        ::-webkit-scrollbar {
          width: 6px;
          height: 6px;
        }
        ::-webkit-scrollbar-track {
          background: transparent; 
        }
        ::-webkit-scrollbar-thumb {
          background: hsl(var(--border)); 
          borderRadius: 3px;
        }
        ::-webkit-scrollbar-thumb:hover {
          background: hsl(var(--muted-foreground)); 
        }

        @media (max-width: 1200px) {
           /* Responsive adjustments if needed */
        }
      `}</style>
    </div>
  )
}
