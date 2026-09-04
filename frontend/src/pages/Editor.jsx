
import React, { useState, useRef, useMemo, useEffect } from 'react'
import { useTheme } from '../theme'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import { useToast } from '../hooks/useToast'
import PdfPreview from '../components/PdfPreview'
import { ToastContainer } from '../components/Toast'

// Imported Components
import ComponentList from '../components/editor/ComponentList'
import DocumentSettings from '../components/editor/DocumentSettings'
import PropertiesPanel from '../components/editor/PropertiesPanel'
import JsonTemplate from '../components/editor/JsonTemplate'
import ComponentItem from '../components/editor/ComponentItem'
import { PAGE_SIZES, DEFAULT_FONTS, COMPONENT_TYPES } from '../components/editor/constants'
import { generatePDFSmart, generateViaServer } from '../utils/wasm/generate.js'
import { loadBundledTemplate } from '../utils/wasm/templates.js'
import { shouldUseServerWasmTransport } from '../utils/wasm/transports.js'
import { registerFontLocal } from '../utils/wasm/fonts.js'
import { getFontFamily } from '../components/editor/utils'
import { parseProps, formatProps } from '../components/editor/utils'
import {
  DEFAULT_CONFIG,
  buildTemplate,
  buildTemplateJson,
  createComponent,
  createFooter,
  createTitle,
  parsePageMargins,
  parseTemplateData,
  parseTemplateJson,
  validateTemplate,
} from '../components/editor/documentModel'

import Toolbar from '../components/editor/Toolbar'
import ContextMenu from '../components/shortcut/ContextMenu'
import useContextMenu from '../components/shortcut/useContextMenu'

// Module-level font cache - cleared on any page refresh (hard or soft)
let _fontsCache = null
let _fontsFetchPromise = null

export default function Editor() {
  const { theme, setTheme } = useTheme()
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [config, setConfig] = useState({ ...DEFAULT_CONFIG, signature: { enabled: false } })
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

  const [copiedId, setCopiedId] = useState(null)
  const [clipboard, setClipboard] = useState(null)
  const [serverRetry, setServerRetry] = useState(null)
  const [templateInput, setTemplateInput] = useState('editor/financial_report.json')
  const canvasRef = useRef(null)
  const { toasts, showToast, removeToast } = useToast()
  const { menuState, showMenu, hideMenu } = useContextMenu()
  const { runJson, runLocal } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => showToast(message, 'error'),
  })

  const downloadJsonTemplate = () => {
    const element = document.createElement('a')
    const file = new Blob([jsonText], { type: 'application/json' })
    const url = URL.createObjectURL(file)
    element.href = url
    element.download = 'template.json'
    document.body.appendChild(element)
    element.click()
    document.body.removeChild(element)
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  // Fetch fonts from API on component mount (module-level cache, single request).
  // Offline-first: failures fall back to DEFAULT_FONTS with a warning, and
  // user uploads register locally via goRegisterFont (see onUploadFont).
  useEffect(() => {
    const loadFonts = async () => {
      if (_fontsCache) {
        setFonts(_fontsCache)
        return
      }
      if (!_fontsFetchPromise) {
        _fontsFetchPromise = runJson(
          {
            endpoint: '/api/v1/fonts',
            method: 'GET',
            getAuthHeaders,
            onError: (message) => console.warn('Failed to fetch fonts, using defaults:', message),
          }
        ).then((data) => {
          if (data && data.fonts && Array.isArray(data.fonts)) {
            _fontsCache = data.fonts
            return data.fonts
          }
          console.warn('Failed to fetch fonts, using defaults')
          return null
        }).catch((error) => {
          console.error('Error fetching fonts:', error)
          _fontsFetchPromise = null
          return null
        })
      }
      const fonts = await _fontsFetchPromise
      if (fonts) setFonts(fonts)
    }
    loadFonts()
  }, [getAuthHeaders, runJson])

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
  const selectedCellElement = selectedElement && selectedCell && selectedElement.type === 'table' && selectedCell.elementId === selectedId
    ? selectedElement.rows[selectedCell.rowIdx].row[selectedCell.colIdx]
    : null

  const currentPageSize = PAGE_SIZES[config.page] || PAGE_SIZES.A4
  const pageMargins = parsePageMargins(config.pageMargin)

  // --- Handlers ---
  const handleDropElement = (type, targetId = null) => {
    if (type === 'title') {
      if (!title) setTitle(createTitle())
    } else if (type === 'footer') {
      if (!footer) setFooter(createFooter())
    } else {
      const newComponent = createComponent(type)

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

  const handleCellDrop = (element, elementId, onUpdate, rowIdx, colIdx, type) => {
    const defaultProps = 'Helvetica:12:000:left:0:0:0:0'
    const newRows = [...element.rows]
    const currentCell = newRows[rowIdx].row[colIdx]

    let newCellData = { ...currentCell }

    if (type === 'checkbox') {
      newCellData = { props: defaultProps, form_field: { name: `checkbox_${Date.now()}`, checked: false, type: 'checkbox' }, text: undefined, image: undefined, chequebox: undefined }
    } else if (type === 'checkbox_simple') {
      newCellData = { props: defaultProps, chequebox: false, text: undefined, image: undefined, form_field: undefined }
    } else if (type === 'text_input') {
      newCellData = { props: defaultProps, form_field: { name: `field_${Date.now()}`, value: '', type: 'text' }, text: undefined, image: undefined, chequebox: undefined }
    } else if (type === 'radio') {
      newCellData = { props: defaultProps, form_field: { name: `radio_${Date.now()}`, checked: false, type: 'radio' }, text: undefined, image: undefined, chequebox: undefined }
    } else if (type === 'radio_simple') {
      newCellData = { props: defaultProps, radio: false, text: undefined, image: undefined, form_field: undefined, chequebox: undefined }
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

  // Handle drag and drop reordering of components
  const handleReorder = (draggedId, targetId) => {
    // Check if draggedId is an existing component (not a new component type)
    if (COMPONENT_TYPES[draggedId]) {
      // This is a new component being dropped, use existing handleDropElement
      handleDropElement(draggedId, targetId)
      return
    }

    // This is an existing component being reordered
    if (draggedId === 'title' || draggedId === 'footer' || targetId === 'title' || targetId === 'footer') {
      // Don't allow reordering title/footer for now
      return
    }

    // Get indices from IDs
    const draggedIndex = parseInt(draggedId.split('-')[1])
    const targetIndex = parseInt(targetId.split('-')[1])

    if (isNaN(draggedIndex) || isNaN(targetIndex) || draggedIndex === targetIndex) {
      return
    }

    // Reorder the components array
    const newComponents = [...components]
    const [draggedComponent] = newComponents.splice(draggedIndex, 1)
    newComponents.splice(targetIndex, 0, draggedComponent)
    setComponents(newComponents)

    // Update selection to follow the dragged component
    const newId = `${draggedComponent.type}-${targetIndex}`
    setSelectedId(newId)
  }

  // --- Context Menu Handlers ---

  // Find element by ID across title, components, footer
  const findElementById = (id) => {
    if (id === 'title') return title ? { ...title, type: 'title' } : null
    if (id === 'footer') return footer ? { ...footer, type: 'footer' } : null
    const idx = parseInt(id.split('-')[1])
    const comp = components[idx]
    return comp || null
  }

  const handleCopy = (id) => {
    const el = findElementById(id)
    if (!el) return
    const type = id === 'title' ? 'title' : id === 'footer' ? 'footer' : el.type
    setClipboard({ type, data: structuredClone(el) })
    showToast('Copied to clipboard', 'success', 1500)
  }

  const handleCut = (id) => {
    handleCopy(id)
    handleDelete(id)
  }

  const handlePaste = (afterId) => {
    if (!clipboard) return
    const { type, data } = clipboard
    const clone = structuredClone(data)

    if (type === 'title') {
      if (!title) setTitle(clone)
      else showToast('Title already exists', 'error', 2000)
    } else if (type === 'footer') {
      if (!footer) setFooter(clone)
      else showToast('Footer already exists', 'error', 2000)
    } else {
      // Insert after the target, or append at end
      if (afterId && afterId !== 'title' && afterId !== 'footer') {
        const idx = parseInt(afterId.split('-')[1])
        const newComponents = [...components]
        newComponents.splice(idx + 1, 0, clone)
        setComponents(newComponents)
      } else {
        setComponents([...components, clone])
      }
    }
  }

  const handleDuplicate = (id) => {
    const el = findElementById(id)
    if (!el) return
    const clone = structuredClone(el)

    if (id === 'title') {
      showToast('Cannot duplicate title - only one allowed', 'error', 2000)
      return
    }
    if (id === 'footer') {
      showToast('Cannot duplicate footer - only one allowed', 'error', 2000)
      return
    }

    const idx = parseInt(id.split('-')[1])
    const newComponents = [...components]
    newComponents.splice(idx + 1, 0, clone)
    setComponents(newComponents)
  }

  // Toggle style bit (0=bold, 1=italic, 2=underline) on an element's props
  const handleToggleStyle = (id, bitIndex) => {
    const el = findElementById(id)
    if (!el) return

    if (id === 'title' && el.table) {
      // For title, toggle on the textprops
      const parsed = parseProps(el.textprops || el.props)
      const s = parsed.style.split('')
      s[bitIndex] = s[bitIndex] === '1' ? '0' : '1'
      handleUpdate(id, { textprops: formatProps({ ...parsed, style: s.join('') }) })
    } else if (id === 'footer') {
      const parsed = parseProps(el.props)
      const s = parsed.style.split('')
      s[bitIndex] = s[bitIndex] === '1' ? '0' : '1'
      handleUpdate(id, { props: formatProps({ ...parsed, style: s.join('') }) })
    }
  }

  // Toggle style on a specific cell
  const handleToggleCellStyle = (elementId, rowIdx, colIdx, bitIndex) => {
    const el = findElementById(elementId)
    if (!el || !el.rows) return
    const newRows = structuredClone(el.rows)
    const cell = newRows[rowIdx]?.row?.[colIdx]
    if (!cell) return
    const parsed = parseProps(cell.props)
    const s = parsed.style.split('')
    s[bitIndex] = s[bitIndex] === '1' ? '0' : '1'
    cell.props = formatProps({ ...parsed, style: s.join('') })
    handleUpdate(elementId, { rows: newRows })
  }

  // Set alignment on element
  const handleSetAlignment = (id, align) => {
    const el = findElementById(id)
    if (!el) return

    if (id === 'title' && el.table) {
      const parsed = parseProps(el.textprops || el.props)
      handleUpdate(id, { textprops: formatProps({ ...parsed, align }) })
    } else if (id === 'footer') {
      const parsed = parseProps(el.props)
      handleUpdate(id, { props: formatProps({ ...parsed, align }) })
    }
  }

  // Set alignment on a cell
  const handleSetCellAlignment = (elementId, rowIdx, colIdx, align) => {
    const el = findElementById(elementId)
    if (!el || !el.rows) return
    const newRows = structuredClone(el.rows)
    const cell = newRows[rowIdx]?.row?.[colIdx]
    if (!cell) return
    const parsed = parseProps(cell.props)
    cell.props = formatProps({ ...parsed, align })
    handleUpdate(elementId, { rows: newRows })
  }

  // Border presets for element props
  const borderPresets = { none: [0, 0, 0, 0], all: [1, 1, 1, 1], box: [1, 1, 1, 1], bottom: [0, 0, 0, 1] }

  const handleSetBorderPreset = (id, preset) => {
    const el = findElementById(id)
    if (!el) return
    const borders = borderPresets[preset] || [0, 0, 0, 0]

    if (id === 'title' && el.table) {
      const parsed = parseProps(el.textprops || el.props)
      handleUpdate(id, { textprops: formatProps({ ...parsed, borders }), props: formatProps({ ...parseProps(el.props), borders }) })
    } else if (id === 'footer') {
      const parsed = parseProps(el.props)
      handleUpdate(id, { props: formatProps({ ...parsed, borders }) })
    }
  }

  const handleSetCellBorderPreset = (elementId, rowIdx, colIdx, preset) => {
    const el = findElementById(elementId)
    if (!el || !el.rows) return
    const borders = borderPresets[preset] || [0, 0, 0, 0]
    const newRows = structuredClone(el.rows)
    const cell = newRows[rowIdx]?.row?.[colIdx]
    if (!cell) return
    const parsed = parseProps(cell.props)
    cell.props = formatProps({ ...parsed, borders })
    handleUpdate(elementId, { rows: newRows })
  }

  // Add/remove rows and columns
  const handleAddRow = (id) => {
    const el = findElementById(id)
    if (!el || !el.rows) return
    const colCount = el.rows[0]?.row?.length || el.maxcolumns || 3
    const newRow = { row: Array.from({ length: colCount }, () => ({ props: 'Helvetica:12:000:left:1:1:1:1', text: '' })) }
    handleUpdate(id, { rows: [...el.rows, newRow] })
  }

  const handleAddColumn = (id) => {
    const el = findElementById(id)
    if (!el || !el.rows) return
    const newRows = el.rows.map(r => ({
      ...r,
      row: [...r.row, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }]
    }))
    handleUpdate(id, { rows: newRows, maxcolumns: (el.maxcolumns || el.rows[0].row.length) + 1 })
  }

  const handleRemoveRow = (id) => {
    const el = findElementById(id)
    if (!el || !el.rows || el.rows.length <= 1) return
    handleUpdate(id, { rows: el.rows.slice(0, -1) })
  }

  const handleRemoveColumn = (id) => {
    const el = findElementById(id)
    if (!el || !el.rows) return
    const colCount = el.rows[0]?.row?.length || 0
    if (colCount <= 1) return
    const newRows = el.rows.map(r => ({ ...r, row: r.row.slice(0, -1) }))
    handleUpdate(id, { rows: newRows, maxcolumns: Math.max(1, (el.maxcolumns || colCount) - 1) })
  }

  // Toggle text wrap on a cell
  const handleToggleWrap = (elementId, rowIdx, colIdx) => {
    const el = findElementById(elementId)
    if (!el || !el.rows) return
    const newRows = structuredClone(el.rows)
    const cell = newRows[rowIdx]?.row?.[colIdx]
    if (!cell) return
    cell.wrap = !cell.wrap
    handleUpdate(elementId, { rows: newRows })
  }

  // Insert form field into cell (used by context menu)
  const handleInsertField = (elementId, rowIdx, colIdx, type) => {
    const el = findElementById(elementId)
    if (!el) return
    handleCellDrop(el, elementId, (updates) => handleUpdate(elementId, updates), rowIdx, colIdx, type)
  }

  // Delete a specific row by index
  const handleDeleteRow = (id, rowIdx) => {
    const el = findElementById(id)
    if (!el || !el.rows || el.rows.length <= 1) return
    const newRows = el.rows.filter((_, i) => i !== rowIdx)
    handleUpdate(id, { rows: newRows })
  }

  // Delete a specific column by index
  const handleDeleteColumn = (id, colIdx) => {
    const el = findElementById(id)
    if (!el || !el.rows) return
    const colCount = el.rows[0]?.row?.length || 0
    if (colCount <= 1) return
    const newRows = el.rows.map(r => ({ ...r, row: r.row.filter((_, i) => i !== colIdx) }))
    handleUpdate(id, { rows: newRows, maxcolumns: Math.max(1, (el.maxcolumns || colCount) - 1) })
  }

  // Clear a specific cell (reset text and props)
  const handleClearCell = (id, rowIdx, colIdx) => {
    const el = findElementById(id)
    if (!el || !el.rows) return
    const newRows = structuredClone(el.rows)
    const cell = newRows[rowIdx]?.row?.[colIdx]
    if (!cell) return
    cell.text = ''
    cell.props = 'Helvetica:12:000:left:1:1:1:1'
    delete cell.image
    delete cell.checkbox
    delete cell.radio
    delete cell.hyperlink
    delete cell.text_input
    handleUpdate(id, { rows: newRows })
  }

  // Aggregate all context menu handlers
  const contextMenuHandlers = {
    cut: handleCut,
    copy: handleCopy,
    paste: handlePaste,
    duplicate: handleDuplicate,
    delete: handleDelete,
    toggleStyle: handleToggleStyle,
    toggleCellStyle: handleToggleCellStyle,
    setAlignment: handleSetAlignment,
    setCellAlignment: handleSetCellAlignment,
    setBorderPreset: handleSetBorderPreset,
    setCellBorderPreset: handleSetCellBorderPreset,
    addRow: handleAddRow,
    addColumn: handleAddColumn,
    removeRow: handleRemoveRow,
    removeColumn: handleRemoveColumn,
    deleteRow: handleDeleteRow,
    deleteColumn: handleDeleteColumn,
    clearCell: handleClearCell,
    toggleWrap: handleToggleWrap,
    insertField: handleInsertField,
    addElement: handleDropElement,
    moveUp: (index) => handleMove(index, 'up'),
    moveDown: (index) => handleMove(index, 'down')
  }

  // --- JSON Handling ---
  const [jsonText, setJsonText] = useState('')
  const [isJsonEditing, setIsJsonEditing] = useState(false)

  useEffect(() => {
    if (isJsonEditing) return
    setJsonText(buildTemplateJson({ config, title, components, footer, bookmarks }))
  }, [config, title, components, footer, bookmarks, isJsonEditing])

  const handleJsonChange = (e) => setJsonText(e.target.value)

  const handleJsonBlur = () => {
    setIsJsonEditing(false)
    try {
      const parsed = parseTemplateJson(jsonText, config)
      setConfig(parsed.config)
      setTitle(parsed.title)
      setComponents(parsed.components)
      setFooter(parsed.footer)
      setBookmarks(parsed.bookmarks)
    } catch (e) {
      console.error('Invalid JSON', e)
    }
  }

  // --- PDF Generation ---
  // Browser-local first via gopdfsuit.wasm (offline once downloaded; the
  // engine runs no JS, so output matches the server byte path). Server only
  // on explicit retry (serverRetry banner) or VITE_WASM_TRANSPORT=server.
  const handleGeneratePdf = async (isPreview = false) => {
    setIsJsonEditing(false)
    setServerRetry(null)
    const template = buildTemplate({ config, title, components, footer, bookmarks })
    const { errors, warnings } = validateTemplate(template)
    if (errors.length > 0 || warnings.length > 0) {
      console.warn('Template schema issues:', { errors, warnings })
    }

    if (shouldUseServerWasmTransport()) {
      await runServerGenerate(template, isPreview)
      return
    }
    let wasmMessage = ''
    const url = await runLocal(() => generatePDFSmart(template, { getAuthHeaders }), {
      autoDownload: !isPreview,
      filename: 'generated_document.pdf',
      onBlob: isPreview
        ? (blob, blobUrl) => {
          setPdfUrl(blobUrl)
          setShowPreviewModal(true)
        }
        : undefined,
      onError: (message) => { wasmMessage = message },
    })
    if (url) return
    if (getAuthHeaders) {
      setServerRetry({ message: wasmMessage, template, isPreview })
    } else {
      showToast(wasmMessage || 'Browser generate failed', 'error')
    }
  }

  const runServerGenerate = async (template, isPreview = false) => {
    await runLocal(() => generateViaServer(template, getAuthHeaders), {
      filename: 'generated_document.pdf',
      autoDownload: !isPreview,
      onBlob: isPreview
        ? (blob, url) => {
          setPdfUrl(url)
          setShowPreviewModal(true)
        }
        : undefined,
      onError: (message) => showToast(message, 'error'),
    })
  }

  const retryServerGenerate = async () => {
    if (!serverRetry) return
    const { template, isPreview } = serverRetry
    setServerRetry(null)
    await runServerGenerate(template, isPreview)
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
  const onLoadTemplate = async (filename, source = 'local') => {
    if (!filename || !filename.trim()) {
      showToast('Please enter a template filename', 'error')
      return
    }

    try {
      let templateData;

      if (source === 'github') {
        const response = await fetch(`https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/master/sampledata/${filename}`);
        if (!response.ok) {
          throw new Error(`Failed to load from GitHub: ${response.status} ${response.statusText}`);
        }
        templateData = await response.json();
      } else {
        // Offline-first: bundled samples from /templates/ (Cache API, no
        // server). Unknown names fall through to the server endpoint below.
        try {
          templateData = await loadBundledTemplate(filename)
        } catch (bundledError) {
          if (!bundledError || !bundledError.fallbackAvailable) throw bundledError
          // GET the template through the shared hook (auth retry + error mapping included)
          templateData = await runJson(
            {
              endpoint: `/api/v1/template-data?file=${encodeURIComponent(filename)}`,
              method: 'GET',
              headers: {
                'Accept': 'application/json'
              },
              getAuthHeaders,
              onError: (message) => showToast(message, 'error'),
            }
          )
          if (!templateData) return
        }
      }

      // Parse and load the template data
      const parsed = parseTemplateData(templateData, config)
      setConfig(parsed.config)
      setTitle(parsed.title)
      setComponents(parsed.components)
      setFooter(parsed.footer)
      setBookmarks(parsed.bookmarks)

      // Update JSON display
      setIsJsonEditing(false)

      // Clear selection
      setSelectedId(null)
      setSelectedCell(null)

    } catch (error) {
      console.error('Error loading template:', error)
      showToast(error.message || 'Failed to load template', 'error')
    }
  }

  // --- Keyboard Shortcuts ---
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Skip shortcuts when typing in input fields
      const tag = e.target.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') {
        // Only handle Ctrl+S to save JSON even in inputs
        if ((e.metaKey || e.ctrlKey) && e.key === 's') {
          e.preventDefault()
          downloadJsonTemplate()
        }
        return
      }

      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        downloadJsonTemplate()
        return
      }

      if (!selectedId) return

      const ctrl = e.metaKey || e.ctrlKey

      if (e.key === 'Delete' || e.key === 'Backspace') {
        e.preventDefault()
        handleDelete(selectedId)
      } else if (ctrl && e.key === 'c') {
        e.preventDefault()
        handleCopy(selectedId)
      } else if (ctrl && e.key === 'x') {
        e.preventDefault()
        handleCut(selectedId)
      } else if (ctrl && e.key === 'v') {
        e.preventDefault()
        handlePaste(selectedId)
      } else if (ctrl && e.key === 'd') {
        e.preventDefault()
        handleDuplicate(selectedId)
      } else if (ctrl && e.key === 'b') {
        e.preventDefault()
        if (selectedCell) {
          handleToggleCellStyle(selectedCell.elementId, selectedCell.rowIdx, selectedCell.colIdx, 0)
        } else {
          handleToggleStyle(selectedId, 0)
        }
      } else if (ctrl && e.key === 'i') {
        e.preventDefault()
        if (selectedCell) {
          handleToggleCellStyle(selectedCell.elementId, selectedCell.rowIdx, selectedCell.colIdx, 1)
        } else {
          handleToggleStyle(selectedId, 1)
        }
      } else if (ctrl && e.key === 'u') {
        e.preventDefault()
        if (selectedCell) {
          handleToggleCellStyle(selectedCell.elementId, selectedCell.rowIdx, selectedCell.colIdx, 2)
        } else {
          handleToggleStyle(selectedId, 2)
        }
      } else if (e.altKey && e.key === 'ArrowUp') {
        e.preventDefault()
        const el = allElements.find(el => el.id === selectedId)
        if (el && el.type !== 'title' && el.type !== 'footer') {
          const idx = parseInt(selectedId.split('-')[1])
          handleMove(idx, 'up')
        }
      } else if (e.altKey && e.key === 'ArrowDown') {
        e.preventDefault()
        const el = allElements.find(el => el.id === selectedId)
        if (el && el.type !== 'title' && el.type !== 'footer') {
          const idx = parseInt(selectedId.split('-')[1])
          handleMove(idx, 'down')
        }
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [jsonText, selectedId, selectedCell, clipboard, allElements, components, title, footer])

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      minHeight: '100vh',
      background: 'hsl(var(--background))',
      color: 'hsl(var(--foreground))',
      fontFamily: getFontFamily('Helvetica')
    }}>
      {/* Header / Toolbar - Sticky Position */}
      <div className="sticky-header" style={{
        position: 'sticky',
        top: 0,
        zIndex: 100,
        background: 'hsl(var(--card))',
        borderBottom: '1px solid hsl(var(--border))',
        padding: '0.75rem 1rem',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
      }}>
        {serverRetry && (
          <div style={{ padding: '0.75rem 1rem', background: 'rgba(255, 193, 7, 0.1)', border: '1px solid #ffc107', borderRadius: '8px', marginBottom: '0.75rem', color: 'hsl(var(--foreground))', fontSize: '0.9rem' }}>
            <div style={{ marginBottom: '0.5rem' }}>
              Browser generate is not available in this build{serverRetry.message ? `: ${serverRetry.message}` : '.'} The template was not uploaded.
              Upload it to the server to generate instead?
            </div>
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button onClick={retryServerGenerate} className="btn-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
                Upload to server and generate
              </button>
              <button onClick={() => setServerRetry(null)} className="btn-outline-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
                Stay local
              </button>
            </div>
          </div>
        )}
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
          elementCount={allElements.length}
          pageSize={config.page}
          onUploadFont={async (file) => {
            // Local-first: register into the WASM registry so offline
            // generation embeds it, then keep the server upload so shared
            // backends see it too.
            try {
              const bytes = new Uint8Array(await file.arrayBuffer())
              const name = String(file.name || 'custom').replace(/\.(ttf|otf)$/i, '')
              await registerFontLocal(name, bytes)
              setFonts((prev) => {
                if (prev.some((f) => f.id === name || f.name === name)) return prev
                return [...prev, { id: name, name, displayName: name }]
              })
              showToast(`Font "${name}" registered locally!`, 'success')
            } catch (error) {
              console.error('Local font registration failed:', error)
              showToast(error.message || 'Local font registration failed', 'error')
            }
            try {
              const formData = new FormData()
              formData.append('font', file)
              const data = await runJson(
                {
                  endpoint: '/api/v1/fonts',
                  method: 'POST',
                  body: formData,
                  getAuthHeaders,
                  onError: (message) => showToast(message, 'error'),
                }
              )
              if (data) {
                showToast(`Font "${data.name}" uploaded successfully!`, 'success')
                // Refresh fonts list (invalidate cache)
                _fontsCache = null
                _fontsFetchPromise = null
                const fontsData = await runJson(
                  {
                    endpoint: '/api/v1/fonts',
                    method: 'GET',
                    getAuthHeaders,
                    onError: (message) => showToast(message, 'error'),
                  }
                )
                if (fontsData && fontsData.fonts && Array.isArray(fontsData.fonts)) {
                  _fontsCache = fontsData.fonts
                  setFonts(fontsData.fonts)
                }
              }
            } catch (error) {
              console.error('Error uploading font:', error)
              showToast(`Error uploading font: ${error.message}`, 'error')
            }
          }}
        />
      </div>

      {/* Main Content using CSS Grid */}
      <div className="editor-main-grid" style={{
        flex: 1,
        display: 'grid',
        gridTemplateColumns: '280px minmax(600px, 1fr) 320px',
        gap: '1.5rem',
        padding: '1.5rem',
        minHeight: 0
      }}>

        {/* Left Column: Settings and Components */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', overflowY: 'auto' }}>
          {/* We merge Settings and Components into the left column to match typical 3-col layout */}
          <ComponentList draggedType={draggedType} setDraggedType={setDraggedType} />
          <DocumentSettings config={config} setConfig={setConfig} currentPageSize={currentPageSize} />
        </div>

        {/* Center Column: Canvas */}
        <div className="canvas-container" style={{
          background: 'hsl(var(--muted))',
          borderRadius: '8px',
          padding: '1.5rem',
          overflowY: 'auto',
          overflowX: 'visible',
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
              width: '100%',
              display: 'flex',
              justifyContent: 'center',
              padding: '2rem 0.5rem',
              background: 'hsl(var(--muted) / 0.3)'
            }}
          >
            <div
              ref={canvasRef}
              style={{
                width: `${currentPageSize.width}px`,
                minHeight: `${currentPageSize.height}px`,
                // Auto height allows content to push it down, min-height ensures at least one page
                height: 'auto',
                background: isDragOver ? 'repeating-linear-gradient(45deg, hsl(var(--accent)) 0px, hsl(var(--accent)) 2px, transparent 2px, transparent 20px)' : 'white',
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                padding: `${pageMargins.top}px ${pageMargins.right}px ${pageMargins.bottom + 50}px ${pageMargins.left}px`,
                position: 'relative',
                display: 'flex',
                flexDirection: 'column',
                gap: '0px',
                border: isDragOver ? '2px dashed hsl(var(--accent))' : '1px solid #e5e5e5',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                color: '#000'
              }}
              onDragOver={(e) => { e.preventDefault(); setIsDragOver(true) }}
              onDragLeave={() => setIsDragOver(false)}
              onDrop={(e) => {
                e.preventDefault(); setIsDragOver(false)
                const type = e.dataTransfer.getData('text/plain')
                // Basic drop on canvas background works, but we also handle drop on items for insertion
                if (COMPONENT_TYPES[type]) handleDropElement(type)
              }}
              onClick={() => { setSelectedId(null); setSelectedCell(null) }}
              onContextMenu={(e) => showMenu(e, 'canvas', {})}
            >
              {/* Background Grid - only at top and left edge */}
              <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '20px', background: 'repeating-linear-gradient(90deg, transparent, transparent 49px, #f0f0f0 50px)', pointerEvents: 'none', opacity: 0.5 }} />
              <div style={{ position: 'absolute', top: 0, left: 0, height: '100%', width: '20px', background: 'repeating-linear-gradient(0deg, transparent, transparent 49px, #f0f0f0 50px)', pointerEvents: 'none', opacity: 0.5 }} />



              {/* Page Border (only for first page to avoid complexity) */}
              {config.pageBorder && config.pageBorder !== '0:0:0:0' && (
                <div style={{
                  position: 'absolute',
                  top: pageMargins.top,
                  left: pageMargins.left,
                  width: `${currentPageSize.width - pageMargins.left - pageMargins.right}px`,
                  height: `${currentPageSize.height - pageMargins.top - pageMargins.bottom}px`,
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

              {/* Render Elements with Drop Zones */}
              {allElements.length === 0 ? (
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', color: '#999', border: '2px dashed #eee', borderRadius: '8px', margin: '2rem', padding: '3rem' }}>
                  <p style={{ margin: 0, fontSize: '14px' }}>Drop components here to start</p>
                </div>
              ) : (
                <>
                  {allElements.map((element, index) => {
                    // Calculate the actual component index for move operations
                    let componentIndex = -1
                    if (element.type !== 'title' && element.type !== 'footer') {
                      componentIndex = parseInt(element.id.split('-')[1])
                    }
                    const canMoveUp = componentIndex > 0
                    const canMoveDown = componentIndex >= 0 && componentIndex < components.length - 1

                    return (
                      <React.Fragment key={element.id}>
                        {/* Drop Zone Before Element */}
                        {(draggedType || draggedComponentId) && (
                          <div
                            style={{
                              height: '4px',
                              width: '100%',
                              margin: '2px 0',
                              background: 'transparent',
                              position: 'relative',
                              transition: 'all 0.35s cubic-bezier(0.4, 0, 0.2, 1)',
                              overflow: 'hidden'
                            }}
                            onDragOver={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              e.currentTarget.style.height = '48px'
                              e.currentTarget.style.background = 'linear-gradient(90deg, hsl(var(--accent) / 0.1) 0%, hsl(var(--accent) / 0.25) 50%, hsl(var(--accent) / 0.1) 100%)'
                              e.currentTarget.style.border = '2px dashed hsl(var(--accent))'
                              e.currentTarget.style.borderRadius = '8px'
                              e.currentTarget.style.boxShadow = '0 4px 12px hsl(var(--accent) / 0.2)'
                              const textEl = e.currentTarget.querySelector('div')
                              if (textEl) {
                                textEl.style.opacity = '1'
                                textEl.style.transform = 'translate(-50%, -50%) scale(1)'
                              }
                            }}
                            onDragLeave={(e) => {
                              e.currentTarget.style.height = '4px'
                              e.currentTarget.style.background = 'transparent'
                              e.currentTarget.style.border = 'none'
                              e.currentTarget.style.boxShadow = 'none'
                              const textEl = e.currentTarget.querySelector('div')
                              if (textEl) {
                                textEl.style.opacity = '0'
                                textEl.style.transform = 'translate(-50%, -50%) scale(0.9)'
                              }
                            }}
                            onDrop={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              e.currentTarget.style.height = '4px'
                              e.currentTarget.style.background = 'transparent'
                              e.currentTarget.style.border = 'none'
                              const textEl = e.currentTarget.querySelector('div')
                              if (textEl) textEl.style.opacity = '0'
                              const type = e.dataTransfer.getData('text/plain')
                              // Check if it's a new component or existing component being reordered
                              if (COMPONENT_TYPES[type]) {
                                handleDropElement(type, element.id)
                              } else {
                                // It's an existing component ID, handle reordering
                                handleReorder(type, element.id)
                              }
                            }}
                          >
                            <div style={{
                              position: 'absolute',
                              top: '50%',
                              left: '50%',
                              transform: 'translate(-50%, -50%) scale(0.9)',
                              fontSize: '12px',
                              fontWeight: '600',
                              color: 'hsl(var(--accent))',
                              opacity: 0,
                              pointerEvents: 'none',
                              whiteSpace: 'nowrap',
                              transition: 'all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1)',
                              textShadow: '0 1px 2px rgba(0,0,0,0.1)'
                            }}>
                              📍 Drop here to insert
                            </div>
                          </div>
                        )}

                        <ComponentItem
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
                          onDrop={(draggedId, targetId) => handleReorder(draggedId, targetId)}
                          isDragging={draggedComponentId === element.id}
                          draggedType={draggedType}
                          handleCellDrop={handleCellDrop}
                          currentPageSize={currentPageSize}
                          pageMargins={pageMargins}
                          onContextMenu={showMenu}
                        />

                        {/* Drop Zone After Last Element */}
                        {index === allElements.length - 1 && (draggedType || draggedComponentId) && (
                          <div
                            style={{
                              height: '4px',
                              width: '100%',
                              margin: '2px 0',
                              background: 'transparent',
                              position: 'relative',
                              transition: 'all 0.35s cubic-bezier(0.4, 0, 0.2, 1)',
                              overflow: 'hidden'
                            }}
                            onDragOver={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              e.currentTarget.style.height = '48px'
                              e.currentTarget.style.background = 'linear-gradient(90deg, hsl(var(--accent) / 0.1) 0%, hsl(var(--accent) / 0.25) 50%, hsl(var(--accent) / 0.1) 100%)'
                              e.currentTarget.style.border = '2px dashed hsl(var(--accent))'
                              e.currentTarget.style.borderRadius = '8px'
                              e.currentTarget.style.boxShadow = '0 4px 12px hsl(var(--accent) / 0.2)'
                              const textEl = e.currentTarget.querySelector('div')
                              if (textEl) {
                                textEl.style.opacity = '1'
                                textEl.style.transform = 'translate(-50%, -50%) scale(1)'
                              }
                            }}
                            onDragLeave={(e) => {
                              e.currentTarget.style.height = '4px'
                              e.currentTarget.style.background = 'transparent'
                              e.currentTarget.style.border = 'none'
                              e.currentTarget.style.boxShadow = 'none'
                              const textEl = e.currentTarget.querySelector('div')
                              if (textEl) {
                                textEl.style.opacity = '0'
                                textEl.style.transform = 'translate(-50%, -50%) scale(0.9)'
                              }
                            }}
                            onDrop={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              e.currentTarget.style.height = '4px'
                              e.currentTarget.style.background = 'transparent'
                              e.currentTarget.style.border = 'none'
                              const textEl = e.currentTarget.querySelector('div')
                              if (textEl) textEl.style.opacity = '0'
                              const type = e.dataTransfer.getData('text/plain')
                              // Check if it's a new component or existing component being reordered
                              if (COMPONENT_TYPES[type]) {
                                handleDropElement(type, null) // null means append at end
                              } else {
                                // It's an existing component ID, move it to the end
                                const draggedIndex = parseInt(type.split('-')[1])
                                if (!isNaN(draggedIndex)) {
                                  const newComponents = [...components]
                                  const [draggedComponent] = newComponents.splice(draggedIndex, 1)
                                  newComponents.push(draggedComponent)
                                  setComponents(newComponents)
                                  const newId = `${draggedComponent.type}-${newComponents.length - 1}`
                                  setSelectedId(newId)
                                }
                              }
                            }}
                          >
                            <div style={{
                              position: 'absolute',
                              top: '50%',
                              left: '50%',
                              transform: 'translate(-50%, -50%) scale(0.9)',
                              fontSize: '12px',
                              fontWeight: '600',
                              color: 'hsl(var(--accent))',
                              opacity: 0,
                              pointerEvents: 'none',
                              whiteSpace: 'nowrap',
                              transition: 'all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1)',
                              textShadow: '0 1px 2px rgba(0,0,0,0.1)'
                            }}>
                              📍 Drop here to add at end
                            </div>
                          </div>
                        )}
                      </React.Fragment>
                    )
                  })}
                </>
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
            bookmarks={bookmarks}
            setBookmarks={setBookmarks}
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

      {/* Toast Notifications */}
      <ToastContainer toasts={toasts} removeToast={removeToast} />

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

      {/* Context Menu */}
      <ContextMenu
        menuState={menuState}
        onHide={hideMenu}
        handlers={contextMenuHandlers}
        clipboard={clipboard}
        hasTitle={!!title}
      />

      <style jsx>{`
        .dragging {
          transform: rotate(3deg) scale(0.95);
        }
        
        .canvas-container {
          min-height: 500px;
          max-height: calc(100vh - 200px);
          overflow-y: auto !important;
          overflow-x: hidden;
        }
        
        .sticky-header {
          position: sticky;
          top: 0;
          z-index: 100;
          background: hsl(var(--background));
          border-bottom: 1px solid hsl(var(--border));
          padding: 0.75rem 1rem;
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

        @media (max-width: 1400px) {
          .editor-main-grid {
            grid-template-columns: 240px 1fr 300px !important;
          }
        }
        
        @media (max-width: 1100px) {
          .editor-main-grid {
            grid-template-columns: 1fr !important;
          }
          .editor-sidebar {
            height: auto;
            position: relative;
            top: 0;
          }
          .canvas-container {
            min-height: 400px;
            max-height: none;
          }
        }
      `}</style>
    </div>
  )
}
