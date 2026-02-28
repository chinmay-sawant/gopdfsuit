
import React, { useState, useRef, useMemo, useEffect } from 'react'
import { useTheme } from '../theme'
import { useAuth } from '../contexts/AuthContext'
import { makeAuthenticatedRequest, isAuthRequired } from '../utils/apiConfig'
import PdfPreview from '../components/PdfPreview'
import Toast from '../components/Toast'

// Imported Components
import ComponentList from '../components/editor/ComponentList'
import DocumentSettings from '../components/editor/DocumentSettings'
import PropertiesPanel from '../components/editor/PropertiesPanel'
import JsonTemplate from '../components/editor/JsonTemplate'
import ComponentItem from '../components/editor/ComponentItem'
import { PAGE_SIZES, DEFAULT_FONTS, COMPONENT_TYPES } from '../components/editor/constants'
import { getFontFamily, parsePageMargins, parseProps, formatProps } from '../components/editor/utils'

import Toolbar from '../components/editor/Toolbar'
import ContextMenu from '../components/shortcut/ContextMenu'
import useContextMenu from '../components/shortcut/useContextMenu'

// Module-level font cache — cleared on any page refresh (hard or soft)
let _fontsCache = null
let _fontsFetchPromise = null

export default function Editor() {
  const { theme, setTheme } = useTheme()
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [config, setConfig] = useState({ pageBorder: '1:1:1:1', pageMargin: '72:72:72:72', page: 'A4', pageAlignment: 1, watermark: '', pdfTitle: '', pdfaCompliant: true, signature: { enabled: false } })
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
  const [templateInput, setTemplateInput] = useState('editor/financial_report.json')
  const canvasRef = useRef(null)
  const [toasts, setToasts] = useState([])
  const { menuState, showMenu, hideMenu } = useContextMenu()

  const showToast = (message, type = 'success', duration = 3000) => {
    const id = Date.now()
    setToasts(prev => [...prev, { id, message, type, duration }])
  }

  const removeToast = (id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id))
  }

  // Fetch fonts from API on component mount (module-level cache, single request)
  useEffect(() => {
    const loadFonts = async () => {
      if (_fontsCache) {
        setFonts(_fontsCache)
        return
      }
      if (!_fontsFetchPromise) {
        _fontsFetchPromise = makeAuthenticatedRequest(
          '/api/v1/fonts',
          {},
          isAuthRequired() ? getAuthHeaders : null
        ).then(async (response) => {
          if (response.ok) {
            const data = await response.json()
            if (data.fonts && Array.isArray(data.fonts)) {
              _fontsCache = data.fonts
              return data.fonts
            }
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
  }, [getAuthHeaders])

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
      showToast('Cannot duplicate title — only one allowed', 'error', 2000)
      return
    }
    if (id === 'footer') {
      showToast('Cannot duplicate footer — only one allowed', 'error', 2000)
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

      // Fix embedStandardFonts loading - check both key names since templates may use either
      const embedValue = newConfig?.embedStandardFonts !== undefined
        ? newConfig.embedStandardFonts
        : (newConfig?.embedFonts !== undefined ? newConfig.embedFonts : undefined)

      setConfig(prev => ({
        ...prev,
        ...(newConfig || {}),
        embedStandardFonts: embedValue !== undefined ? embedValue : prev.embedStandardFonts,
        arlingtonCompatible: newConfig?.arlingtonCompatible !== undefined ? newConfig.arlingtonCompatible : prev.arlingtonCompatible,
        pdfaCompliant: newConfig?.pdfaCompliant !== undefined ? newConfig.pdfaCompliant : prev.pdfaCompliant
      }))
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
  const onLoadTemplate = async (filename, source = 'local') => {
    if (!filename || !filename.trim()) {
      alert('Please enter a template filename')
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

        templateData = await response.json()
      }

      // Parse and load the template data
      const { config: newConfig, title: newTitle, elements, table, spacer, content, footer: newFooter, bookmarks: newBookmarks } = templateData

      // Fix embedStandardFonts loading - check both key names since templates may use either
      const embedValue = newConfig?.embedStandardFonts !== undefined
        ? newConfig.embedStandardFonts
        : (newConfig?.embedFonts !== undefined ? newConfig.embedFonts : undefined)

      setConfig(prev => ({
        ...prev,
        ...(newConfig || {}),
        embedStandardFonts: embedValue !== undefined ? embedValue : prev.embedStandardFonts,
        arlingtonCompatible: newConfig?.arlingtonCompatible !== undefined ? newConfig.arlingtonCompatible : prev.arlingtonCompatible,
        pdfaCompliant: newConfig?.pdfaCompliant !== undefined ? newConfig.pdfaCompliant : prev.pdfaCompliant
      }))
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
      // Skip shortcuts when typing in input fields
      const tag = e.target.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') {
        // Only handle Ctrl+S to save JSON even in inputs
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
        return
      }

      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        const element = document.createElement('a')
        const file = new Blob([jsonText], { type: 'application/json' })
        element.href = URL.createObjectURL(file)
        element.download = 'template.json'
        document.body.appendChild(element)
        element.click()
        document.body.removeChild(element)
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
            try {
              const formData = new FormData()
              formData.append('font', file)
              const response = await makeAuthenticatedRequest(
                '/api/v1/fonts',
                {
                  method: 'POST',
                  body: formData
                },
                isAuthRequired() ? getAuthHeaders : null
              )
              if (response.ok) {
                const data = await response.json()
                showToast(`Font "${data.name}" uploaded successfully!`, 'success')
                // Refresh fonts list (invalidate cache)
                _fontsCache = null
                _fontsFetchPromise = null
                const fontsResponse = await makeAuthenticatedRequest(
                  '/api/v1/fonts',
                  {},
                  isAuthRequired() ? getAuthHeaders : null
                )
                if (fontsResponse.ok) {
                  const fontsData = await fontsResponse.json()
                  if (fontsData.fonts && Array.isArray(fontsData.fonts)) {
                    _fontsCache = fontsData.fonts
                    setFonts(fontsData.fonts)
                  }
                }
              } else {
                const error = await response.json()
                showToast(`Failed to upload font: ${error.error || 'Unknown error'}`, 'error')
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
      {toasts.map((toast, index) => (
        <div key={toast.id} style={{ top: `${80 + index * 100}px` }}>
          <Toast
            message={toast.message}
            type={toast.type}
            duration={toast.duration}
            onClose={() => removeToast(toast.id)}
          />
        </div>
      ))}

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

      {/* Toast Notifications */}
      {toasts.map((toast, index) => (
        <div key={toast.id} style={{ top: `${80 + index * 100}px` }}>
          <Toast
            message={toast.message}
            type={toast.type}
            duration={toast.duration}
            onClose={() => removeToast(toast.id)}
          />
        </div>
      ))}

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
