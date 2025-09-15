import React, { useState, useRef, useMemo } from 'react'
import { Edit, Table, Type, Square, Minus, CheckSquare, FileText, Upload, Play, Copy, Sun, Moon, Trash2, Plus, GripVertical, Settings, Eye, Download } from 'lucide-react'
import { useTheme } from '../theme'
import PdfPreview from '../components/PdfPreview'

// Enhanced PDF template editor with improved drag-drop, template loading, and resize functionality
// Features: JSON template parsing, visual component ordering, resize handles, props editor, PDF preview

const PAGE_SIZES = {
  A4: { width: 595, height: 842, name: 'A4' },
  LETTER: { width: 612, height: 792, name: 'Letter' },
  LEGAL: { width: 612, height: 1008, name: 'Legal' },
  A3: { width: 842, height: 1191, name: 'A3' },
  A5: { width: 420, height: 595, name: 'A5' }
}

// Component types matching the JSON template structure
const COMPONENT_TYPES = {
  title: { icon: Type, label: 'Title', defaultText: 'Document Title' },
  table: { icon: Table, label: 'Table', rows: 3, cols: 3 },
  footer: { icon: FileText, label: 'Footer', defaultText: 'Page footer text' },
  spacer: { icon: Minus, label: 'Spacer', height: 20 },
  checkbox: { icon: CheckSquare, label: 'Checkbox' }
}

// Font style helpers
const parseProps = (propsString) => {
  if (!propsString) return { font: 'font1', size: 12, style: '000', align: 'left', borders: [0, 0, 0, 0] }
  const parts = propsString.split(':')
  return {
    font: parts[0] || 'font1',
    size: parseInt(parts[1]) || 12,
    style: parts[2] || '000',
    align: parts[3] || 'left',
    borders: [
      parseInt(parts[4]) || 0,
      parseInt(parts[5]) || 0,
      parseInt(parts[6]) || 0,
      parseInt(parts[7]) || 0
    ]
  }
}

const formatProps = (props) => {
  return `${props.font}:${props.size}:${props.style}:${props.align}:${props.borders.join(':')}`
}

function Toolbar({ theme, setTheme, onLoadTemplate, onPreviewPDF, onCopyJSON, onDownloadPDF }) {
  const [templateInput, setTemplateInput] = useState('')

  return (
    <div className="card" style={{ marginBottom: '1rem', padding: '0.75rem 1rem' }}>
      <div className="flex items-center justify-between" style={{ flexWrap: 'wrap', gap: '1rem' }}>
        <div className="flex items-center gap-3">
          <Edit size={20} />
          <strong>PDF Template Editor</strong>
          <span className="text-muted">Enhanced Designer</span>
        </div>

        <div className="flex items-center gap-2" style={{ flexWrap: 'wrap' }}>
          <input
            type="text"
            value={templateInput}
            onChange={(e) => setTemplateInput(e.target.value)}
            placeholder="Enter filename (e.g., temp_multiplepage.json)"
            style={{ padding: '0.4rem 0.6rem', fontSize: '0.9rem', minWidth: '250px' }}
          />
          <button
            onClick={() => onLoadTemplate(templateInput)}
            className="btn"
            style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem' }}
          >
            <Upload size={14} /> Load
          </button>
          <button
            onClick={onPreviewPDF}
            className="btn btn-secondary"
            style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem' }}
          >
            <Eye size={14} /> Preview
          </button>
          <button
            onClick={onDownloadPDF}
            className="btn"
            style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem' }}
          >
            <Download size={14} /> Generate PDF
          </button>
          <button
            onClick={onCopyJSON}
            className="btn"
            style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem' }}
          >
            <Copy size={14} /> Copy JSON
          </button>

          <div className="flex items-center gap-1" style={{ marginLeft: '1rem' }}>
            <span className="text-muted" style={{ fontSize: '0.85rem' }}>Theme</span>
            <div style={{ display: 'inline-flex', border: '1px solid hsl(var(--border))', borderRadius: '6px' }}>
              <button
                onClick={() => setTheme('light')}
                className={theme === 'light' ? 'btn-secondary' : ''}
                style={{
                  padding: '0.3rem 0.5rem', background: 'transparent', color: 'hsl(var(--foreground))',
                  border: 'none', borderRight: '1px solid hsl(var(--border))', cursor: 'pointer', fontSize: '0.9rem'
                }}
                title="Light"
              >
                <Sun size={14} />
              </button>
              <button
                onClick={() => setTheme('dark')}
                className={theme === 'dark' ? 'btn-secondary' : ''}
                style={{ padding: '0.3rem 0.5rem', background: 'transparent', color: 'hsl(var(--foreground))', border: 'none', cursor: 'pointer', fontSize: '0.9rem' }}
                title="Dark"
              >
                <Moon size={14} />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function DraggableComponent({ type, componentData, isDragging, onDragStart, onDragEnd }) {
  const IconComponent = componentData.icon

  return (
    <div
      draggable
      onDragStart={(e) => {
        e.dataTransfer.setData('text/plain', type)
        onDragStart(type)
      }}
      onDragEnd={() => onDragEnd()}
      className={`draggable-item ${isDragging === type ? 'dragging' : ''}`}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '0.5rem',
        padding: '0.75rem',
        background: 'hsl(var(--accent))',
        border: '1px solid hsl(var(--border))',
        borderRadius: '6px',
        cursor: 'grab',
        userSelect: 'none',
        transition: 'all 0.2s ease',
        opacity: isDragging === type ? 0.5 : 1,
      }}
    >
      <IconComponent size={16} />
      <span style={{ fontSize: '0.9rem' }}>{componentData.label}</span>
    </div>
  )
}

function PropsEditor({ props, onChange }) {
  const parsed = parseProps(props)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
      <div className="flex gap-2">
        <label style={{ width: '50px', fontSize: '0.9rem' }}>Font:</label>
        <select
          value={parsed.font}
          onChange={(e) => onChange(formatProps({ ...parsed, font: e.target.value }))}
          style={{ flex: 1, padding: '0.25rem' }}
        >
          <option value="font1">Font 1</option>
          <option value="font2">Font 2</option>
        </select>
      </div>
      <div className="flex gap-2">
        <label style={{ width: '50px', fontSize: '0.9rem' }}>Size:</label>
        <input
          type="number"
          min="8"
          max="72"
          value={parsed.size}
          onChange={(e) => onChange(formatProps({ ...parsed, size: parseInt(e.target.value) }))}
          style={{ flex: 1, padding: '0.25rem' }}
        />
      </div>
      <div className="flex gap-2">
        <label style={{ width: '50px', fontSize: '0.9rem' }}>Style:</label>
        <div style={{ display: 'flex', gap: '0.5rem', flex: 1 }}>
          <label style={{ fontSize: '0.8rem' }}>
            <input
              type="checkbox"
              checked={parsed.style[0] === '1'}
              onChange={(e) => {
                const newStyle = parsed.style.split('')
                newStyle[0] = e.target.checked ? '1' : '0'
                onChange(formatProps({ ...parsed, style: newStyle.join('') }))
              }}
            /> B
          </label>
          <label style={{ fontSize: '0.8rem' }}>
            <input
              type="checkbox"
              checked={parsed.style[1] === '1'}
              onChange={(e) => {
                const newStyle = parsed.style.split('')
                newStyle[1] = e.target.checked ? '1' : '0'
                onChange(formatProps({ ...parsed, style: newStyle.join('') }))
              }}
            /> I
          </label>
          <label style={{ fontSize: '0.8rem' }}>
            <input
              type="checkbox"
              checked={parsed.style[2] === '1'}
              onChange={(e) => {
                const newStyle = parsed.style.split('')
                newStyle[2] = e.target.checked ? '1' : '0'
                onChange(formatProps({ ...parsed, style: newStyle.join('') }))
              }}
            /> U
          </label>
        </div>
      </div>
      <div className="flex gap-2">
        <label style={{ width: '50px', fontSize: '0.9rem' }}>Align:</label>
        <select
          value={parsed.align}
          onChange={(e) => onChange(formatProps({ ...parsed, align: e.target.value }))}
          style={{ flex: 1, padding: '0.25rem' }}
        >
          <option value="left">Left</option>
          <option value="center">Center</option>
          <option value="right">Right</option>
        </select>
      </div>
    </div>
  )
}

function ComponentItem({ element, index, isSelected, onSelect, onUpdate, onMove, onDelete, canMoveUp, canMoveDown }) {
  const [isResizing, setIsResizing] = useState(false)

  const handleClick = (e) => {
    e.stopPropagation()
    onSelect(element.id)
  }

  const renderContent = () => {
    switch (element.type) {
      case 'title':
        return (
          <div style={{ 
            fontSize: '18px', 
            fontWeight: 'bold', 
            textAlign: 'center',
            padding: '10px',
            border: '2px dashed hsl(var(--border))',
            borderRadius: '4px',
            minHeight: '40px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <input
              type="text"
              value={element.text || 'Document Title'}
              onChange={(e) => onUpdate({ text: e.target.value })}
              style={{
                width: '100%',
                border: 'none',
                background: 'transparent',
                fontSize: '18px',
                fontWeight: 'bold',
                textAlign: 'center',
                color: 'hsl(var(--foreground))',
                outline: 'none',
              }}
              placeholder="Document Title"
            />
          </div>
        )
      case 'table':
        return (
          <div style={{ border: '2px dashed hsl(var(--border))', borderRadius: '4px', padding: '10px' }}>
            <table style={{ borderCollapse: 'collapse', width: '100%' }}>
              <tbody>
                {element.rows?.map((row, rowIdx) => (
                  <tr key={rowIdx}>
                    {row.row?.map((cell, colIdx) => (
                      <td
                        key={colIdx}
                        style={{
                          border: '1px solid hsl(var(--border))',
                          padding: '4px 8px',
                          minWidth: '80px',
                          minHeight: '24px',
                          fontSize: '12px'
                        }}
                      >
                        {cell.chequebox !== undefined ? (
                          <input 
                            type="checkbox" 
                            checked={cell.chequebox} 
                            onChange={(e) => {
                              const newRows = [...element.rows]
                              newRows[rowIdx].row[colIdx] = { 
                                ...newRows[rowIdx].row[colIdx], 
                                chequebox: e.target.checked 
                              }
                              onUpdate({ rows: newRows })
                            }}
                          />
                        ) : (
                          <input
                            type="text"
                            value={cell.text || `Cell ${rowIdx + 1},${colIdx + 1}`}
                            onChange={(e) => {
                              const newRows = [...element.rows]
                              newRows[rowIdx].row[colIdx] = { 
                                ...newRows[rowIdx].row[colIdx], 
                                text: e.target.value 
                              }
                              onUpdate({ rows: newRows })
                            }}
                            style={{
                              width: '100%',
                              border: 'none',
                              background: 'transparent',
                              fontSize: '12px',
                              padding: '2px',
                              color: 'hsl(var(--foreground))',
                            }}
                          />
                        )}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      case 'footer':
        return (
          <div style={{ 
            fontSize: '12px', 
            fontStyle: 'italic',
            textAlign: 'center',
            padding: '10px',
            border: '2px dashed hsl(var(--border))',
            borderRadius: '4px',
            minHeight: '30px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <input
              type="text"
              value={element.text || 'Page footer text'}
              onChange={(e) => onUpdate({ text: e.target.value })}
              style={{
                width: '100%',
                border: 'none',
                background: 'transparent',
                fontSize: '12px',
                fontStyle: 'italic',
                textAlign: 'center',
                color: 'hsl(var(--foreground))',
                outline: 'none',
              }}
              placeholder="Page footer text"
            />
          </div>
        )
      case 'spacer':
        return (
          <div style={{ 
            height: element.height || 20, 
            width: '100%', 
            background: 'repeating-linear-gradient(90deg, hsl(var(--muted)) 0px, hsl(var(--muted)) 2px, transparent 2px, transparent 10px)',
            border: '2px dashed hsl(var(--border))',
            borderRadius: '4px',
            opacity: 0.7,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '12px',
            color: 'hsl(var(--muted-foreground))'
          }}>
            Spacer ({element.height || 20}px)
          </div>
        )
      default:
        return null
    }
  }

  return (
    <div 
      onClick={handleClick}
      style={{
        position: 'relative',
        margin: '10px 0',
        padding: isSelected ? '8px' : '0',
        border: isSelected ? '2px solid var(--secondary-color)' : '2px solid transparent',
        borderRadius: '6px',
        cursor: 'pointer',
        background: isSelected ? 'hsl(var(--accent))' : 'transparent',
        transition: 'all 0.2s ease'
      }}
    >
      {isSelected && (
        <div style={{
          position: 'absolute',
          top: '-30px',
          right: '0',
          display: 'flex',
          gap: '2px',
          background: 'hsl(var(--card))',
          border: '1px solid hsl(var(--border))',
          borderRadius: '4px',
          padding: '2px',
          zIndex: 10
        }}>
          <button
            onClick={(e) => { e.stopPropagation(); onMove(index, 'up') }}
            disabled={!canMoveUp}
            style={{ 
              padding: '2px 4px', 
              fontSize: '10px', 
              opacity: canMoveUp ? 1 : 0.5,
              cursor: canMoveUp ? 'pointer' : 'not-allowed'
            }}
          >↑</button>
          <button
            onClick={(e) => { e.stopPropagation(); onMove(index, 'down') }}
            disabled={!canMoveDown}
            style={{ 
              padding: '2px 4px', 
              fontSize: '10px',
              opacity: canMoveDown ? 1 : 0.5,
              cursor: canMoveDown ? 'pointer' : 'not-allowed'
            }}
          >↓</button>
          <button
            onClick={(e) => { e.stopPropagation(); onDelete(element.id) }}
            style={{ 
              padding: '2px 4px', 
              fontSize: '10px',
              color: 'hsl(var(--destructive))'
            }}
          >×</button>
        </div>
      )}
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
        <GripVertical size={14} style={{ color: 'hsl(var(--muted-foreground))' }} />
        <span style={{ fontSize: '12px', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>
          {element.type.charAt(0).toUpperCase() + element.type.slice(1)}
        </span>
      </div>
      {renderContent()}
    </div>
  )
}

export default function Editor() {
  const { theme, setTheme } = useTheme()
  const [config, setConfig] = useState({ pageBorder: '1:1:1:1', page: 'A4', pageAlignment: 1, watermark: '' })
  const [title, setTitle] = useState(null)
  const [tables, setTables] = useState([])
  const [footer, setFooter] = useState(null)
  const [spacers, setSpacers] = useState([])
  const [selectedId, setSelectedId] = useState(null)
  const [draggedType, setDraggedType] = useState(null)
  const [isDragOver, setIsDragOver] = useState(false)
  const [pdfUrl, setPdfUrl] = useState(null)
  const canvasRef = useRef(null)

  // Get all elements in order for display
  const allElements = useMemo(() => {
    const elements = []
    if (title) elements.push({ ...title, id: 'title', type: 'title' })
    tables.forEach((table, idx) => elements.push({ ...table, id: `table-${idx}`, type: 'table' }))
    spacers.forEach((spacer, idx) => elements.push({ ...spacer, id: `spacer-${idx}`, type: 'spacer' }))
    if (footer) elements.push({ ...footer, id: 'footer', type: 'footer' })
    return elements
  }, [title, tables, spacers, footer])

  const selectedElement = allElements.find(el => el.id === selectedId) || null
  const currentPageSize = PAGE_SIZES[config.page] || PAGE_SIZES.A4

  const generateId = () => Math.random().toString(36).substr(2, 9)

  const addElement = (type) => {
    switch (type) {
      case 'title':
        if (!title) {
          setTitle({
            props: 'font1:18:100:center:0:0:1:0',
            text: 'Document Title'
          })
          setSelectedId('title')
        }
        break
      case 'table':
        const newTable = {
          maxcolumns: 3,
          rows: [
            {
              row: [
                { props: 'font1:12:000:left:1:1:1:1', text: 'Cell 1,1' },
                { props: 'font1:12:000:left:1:1:1:1', text: 'Cell 1,2' },
                { props: 'font1:12:000:left:1:1:1:1', text: 'Cell 1,3' }
              ]
            },
            {
              row: [
                { props: 'font1:12:000:left:1:1:1:1', text: 'Cell 2,1' },
                { props: 'font1:12:000:left:1:1:1:1', text: 'Cell 2,2' },
                { props: 'font1:12:000:left:1:1:1:1', text: 'Cell 2,3' }
              ]
            }
          ]
        }
        setTables(prev => [...prev, newTable])
        setSelectedId(`table-${tables.length}`)
        break
      case 'spacer':
        const newSpacer = {
          height: 20
        }
        setSpacers(prev => [...prev, newSpacer])
        setSelectedId(`spacer-${spacers.length}`)
        break
      case 'footer':
        if (!footer) {
          setFooter({
            font: 'font1:10:001:center',
            text: 'Page footer text'
          })
          setSelectedId('footer')
        }
        break
    }
  }

  const updateElement = (id, updates) => {
    if (id === 'title') {
      setTitle(prev => ({ ...prev, ...updates }))
    } else if (id === 'footer') {
      setFooter(prev => ({ ...prev, ...updates }))
    } else if (id.startsWith('table-')) {
      const idx = parseInt(id.split('-')[1])
      setTables(prev => prev.map((table, i) => i === idx ? { ...table, ...updates } : table))
    } else if (id.startsWith('spacer-')) {
      const idx = parseInt(id.split('-')[1])
      setSpacers(prev => prev.map((spacer, i) => i === idx ? { ...spacer, ...updates } : spacer))
    }
  }

  const deleteElement = (id) => {
    if (id === 'title') {
      setTitle(null)
    } else if (id === 'footer') {
      setFooter(null)
    } else if (id.startsWith('table-')) {
      const idx = parseInt(id.split('-')[1])
      setTables(prev => prev.filter((_, i) => i !== idx))
    } else if (id.startsWith('spacer-')) {
      const idx = parseInt(id.split('-')[1])
      setSpacers(prev => prev.filter((_, i) => i !== idx))
    }
    setSelectedId(null)
  }

  const moveElement = (index, direction) => {
    // Only tables can be moved for now
    const tableElements = allElements.filter(el => el.type === 'table')
    const elementIndex = allElements.findIndex((_, i) => i === index)
    const element = allElements[elementIndex]
    
    if (element?.type !== 'table') return
    
    const tableIndex = parseInt(element.id.split('-')[1])
    
    if (direction === 'up' && tableIndex > 0) {
      setTables(prev => {
        const newTables = [...prev]
        const temp = newTables[tableIndex]
        newTables[tableIndex] = newTables[tableIndex - 1]
        newTables[tableIndex - 1] = temp
        return newTables
      })
    } else if (direction === 'down' && tableIndex < tables.length - 1) {
      setTables(prev => {
        const newTables = [...prev]
        const temp = newTables[tableIndex]
        newTables[tableIndex] = newTables[tableIndex + 1]
        newTables[tableIndex + 1] = temp
        return newTables
      })
    }
  }

  const handleCanvasClick = (e) => {
    if (e.target === canvasRef.current) {
      setSelectedId(null)
    }
  }

  const handleDrop = (e) => {
    e.preventDefault()
    setIsDragOver(false)
    setDraggedType(null)
    
    const type = e.dataTransfer.getData('text/plain')
    if (!type || !COMPONENT_TYPES[type]) return
    
    addElement(type)
  }

  const handleDragOver = (e) => {
    e.preventDefault()
    setIsDragOver(true)
  }

  const handleDragLeave = (e) => {
    if (!canvasRef.current.contains(e.relatedTarget)) {
      setIsDragOver(false)
    }
  }

  const loadTemplate = async (filename) => {
    if (!filename.trim()) return
    
    try {
      const response = await fetch(`/api/v1/template-data?file=${encodeURIComponent(filename)}`)
      if (response.ok) {
        const data = await response.json()
        
        // Parse the JSON structure from the template
        setConfig(data.config || { pageBorder: '1:1:1:1', page: 'A4', pageAlignment: 1, watermark: '' })
        setTitle(data.title || null)
        setTables(data.table || [])
        setSpacers(data.spacer || [])
        setFooter(data.footer || null)
        setSelectedId(null)
        
      } else {
        alert('Failed to load template')
      }
    } catch (error) {
      alert('Error loading template: ' + error.message)
    }
  }

  const previewPDF = async () => {
    const templateData = getJsonOutput()

    try {
      const response = await fetch('/api/v1/generate/template-pdf', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(templateData)
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        setPdfUrl(url)
      } else {
        alert('Failed to generate PDF')
      }
    } catch (error) {
      alert('Error generating PDF: ' + error.message)
    }
  }

  const downloadPDF = async () => {
    const templateData = getJsonOutput()

    try {
      const response = await fetch('/api/v1/generate/template-pdf', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(templateData)
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `template-${Date.now()}.pdf`
        a.click()
        URL.revokeObjectURL(url)
      } else {
        alert('Failed to generate PDF')
      }
    } catch (error) {
      alert('Error generating PDF: ' + error.message)
    }
  }

  const copyJSON = async () => {
    const templateData = getJsonOutput()
    
    try {
      await navigator.clipboard.writeText(JSON.stringify(templateData, null, 2))
      alert('JSON copied to clipboard!')
    } catch (error) {
      alert('Failed to copy JSON')
    }
  }

  const getJsonOutput = () => {
    const output = { config }
    if (title) output.title = title
    if (tables.length > 0) output.table = tables
    if (spacers.length > 0) output.spacer = spacers
    if (footer) output.footer = footer
    return output
  }

  const jsonOutput = useMemo(() => getJsonOutput(), [config, title, tables, spacers, footer])

  return (
    <div style={{ padding: '1rem 0', minHeight: '100vh' }}>
      <div className="container-full">
        {/* Header */}
        <div style={{ textAlign: 'center', marginBottom: '1rem' }}>
          <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--foreground))' }}>
            <Edit size={32} /> PDF Template Editor
          </h1>
          <p className="text-muted">Create professional PDF templates with enhanced controls</p>
        </div>

        <Toolbar 
          theme={theme} 
          setTheme={setTheme}
          onLoadTemplate={loadTemplate}
          onPreviewPDF={previewPDF}
          onDownloadPDF={downloadPDF}
          onCopyJSON={copyJSON}
        />

        {/* Main Layout: Toolbox | Canvas | Properties */}
        <div className="grid" style={{ gridTemplateColumns: '280px 1fr 350px', gap: '1rem' }}>
          
          {/* Left: Toolbox */}
          <div className="card" style={{ padding: '1rem' }}>
            <h3 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Square size={18} /> Components
            </h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {Object.entries(COMPONENT_TYPES).map(([type, data]) => (
                <DraggableComponent
                  key={type}
                  type={type}
                  componentData={data}
                  isDragging={draggedType}
                  onDragStart={setDraggedType}
                  onDragEnd={() => setDraggedType(null)}
                />
              ))}
            </div>

            {/* Document Settings */}
            <div style={{ marginTop: '1.5rem' }}>
              <h4 style={{ marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Settings size={16} /> Document Config
              </h4>
              
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.9rem' }}>
                <div>
                  <label style={{ display: 'block', marginBottom: '0.25rem' }}>Page Size:</label>
                  <select
                    value={config.page}
                    onChange={(e) => setConfig(prev => ({ ...prev, page: e.target.value }))}
                    style={{ width: '100%', padding: '0.25rem' }}
                  >
                    {Object.entries(PAGE_SIZES).map(([key, size]) => (
                      <option key={key} value={key}>{size.name}</option>
                    ))}
                  </select>
                </div>
                
                <div>
                  <label style={{ display: 'block', marginBottom: '0.25rem' }}>Orientation:</label>
                  <select
                    value={config.pageAlignment}
                    onChange={(e) => setConfig(prev => ({ ...prev, pageAlignment: parseInt(e.target.value) }))}
                    style={{ width: '100%', padding: '0.25rem' }}
                  >
                    <option value={1}>Portrait</option>
                    <option value={2}>Landscape</option>
                  </select>
                </div>

                <div>
                  <label style={{ display: 'block', marginBottom: '0.25rem' }}>Watermark:</label>
                  <input
                    type="text"
                    value={config.watermark || ''}
                    onChange={(e) => setConfig(prev => ({ ...prev, watermark: e.target.value }))}
                    placeholder="Optional watermark text"
                    style={{ width: '100%', padding: '0.25rem' }}
                  />
                </div>

                <div style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))', marginTop: '0.5rem' }}>
                  <div>Size: {currentPageSize.width} × {currentPageSize.height}</div>
                  <div>Elements: {allElements.length}</div>
                </div>
              </div>
            </div>
          </div>

          {/* Center: Canvas */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {/* Canvas */}
            <div className="card" style={{ padding: '1rem', flex: 1 }}>
              <div style={{ display: 'flex', justifyContent: 'center' }}>
                <div
                  ref={canvasRef}
                  onClick={handleCanvasClick}
                  onDrop={handleDrop}
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  style={{
                    width: '100%',
                    height: '100%',
                    background: isDragOver
                      ? 'repeating-linear-gradient(45deg, hsl(var(--secondary-color)) 0px, hsl(var(--secondary-color)) 2px, transparent 2px, transparent 20px)'
                      : 'hsl(var(--card))',
                    border: isDragOver ? '3px dashed var(--secondary-color)' : '2px solid hsl(var(--border))',
                    borderRadius: '8px',
                    cursor: 'default',
                    color: 'hsl(var(--foreground))',
                    padding: '20px',
                    overflow: 'auto'
                  }}
                >
                  {allElements.length === 0 && !isDragOver && (
                    <div style={{
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      height: '100%',
                      textAlign: 'center',
                      color: 'hsl(var(--muted-foreground))',
                      pointerEvents: 'none',
                    }}>
                      <Square size={48} style={{ opacity: 0.3, marginBottom: '1rem' }} />
                      <p>Drag components here to start building your template</p>
                      <p style={{ fontSize: '0.9rem', opacity: 0.7 }}>Load a template file to see existing content</p>
                    </div>
                  )}
                  {isDragOver && (
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      height: '100%',
                      textAlign: 'center',
                      color: 'var(--secondary-color)',
                      pointerEvents: 'none',
                      fontWeight: '600',
                      fontSize: '1.1rem'
                    }}>
                      Drop here to add component
                    </div>
                  )}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0' }}>
                    {allElements.map((element, index) => (
                      <ComponentItem
                        key={element.id}
                        element={element}
                        index={index}
                        isSelected={element.id === selectedId}
                        onSelect={setSelectedId}
                        onUpdate={(updates) => updateElement(element.id, updates)}
                        onMove={moveElement}
                        onDelete={deleteElement}
                        canMoveUp={element.type === 'table' && index > (title ? 1 : 0)}
                        canMoveDown={element.type === 'table' && index < allElements.length - (footer ? 2 : 1)}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* PDF Preview */}
            {pdfUrl && (
              <PdfPreview 
                pdfUrl={pdfUrl} 
                title="Template Preview"
                onClose={() => setPdfUrl(null)}
              />
            )}
          </div>

          {/* Right: Properties & JSON */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {/* Properties Panel */}
            <div className="card" style={{ padding: '1rem' }}>
              <h3 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Edit size={18} /> Properties
              </h3>
              
              {!selectedElement && (
                <div className="text-muted" style={{ textAlign: 'center', padding: '1rem' }}>
                  Select a component to edit properties
                </div>
              )}
              
              {selectedElement && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <div style={{ padding: '0.5rem', background: 'hsl(var(--accent))', borderRadius: '4px' }}>
                    <strong>{selectedElement.type.charAt(0).toUpperCase() + selectedElement.type.slice(1)}</strong>
                  </div>

                  {selectedElement.type === 'title' && (
                    <>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Text:</label>
                        <input
                          type="text"
                          value={selectedElement.text || ''}
                          onChange={(e) => updateElement(selectedElement.id, { text: e.target.value })}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                        />
                      </div>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.5rem' }}>Font Properties:</label>
                        <PropsEditor 
                          props={selectedElement.props} 
                          onChange={(props) => updateElement(selectedElement.id, { props })}
                        />
                      </div>
                    </>
                  )}

                  {selectedElement.type === 'table' && (
                    <>
                      <div className="flex gap-2">
                        <label style={{ width: '80px', fontSize: '0.9rem' }}>Max Columns:</label>
                        <input
                          type="number"
                          min="1"
                          max="10"
                          value={selectedElement.maxcolumns || 3}
                          onChange={(e) => {
                            const newCols = parseInt(e.target.value)
                            const updatedRows = selectedElement.rows?.map(row => {
                              const newRow = [...(row.row || [])]
                              while (newRow.length < newCols) {
                                newRow.push({ props: 'font1:12:000:left:1:1:1:1', text: `Cell ${row.row?.length || 0 + 1}` })
                              }
                              if (newRow.length > newCols) {
                                newRow.splice(newCols)
                              }
                              return { row: newRow }
                            })
                            updateElement(selectedElement.id, { maxcolumns: newCols, rows: updatedRows })
                          }}
                          style={{ flex: 1, padding: '0.25rem' }}
                        />
                      </div>
                      
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          onClick={() => {
                            const newRow = { 
                              row: Array(selectedElement.maxcolumns || 3).fill().map((_, i) => ({
                                props: 'font1:12:000:left:1:1:1:1',
                                text: `Cell ${(selectedElement.rows?.length || 0) + 1},${i + 1}`
                              }))
                            }
                            updateElement(selectedElement.id, { 
                              rows: [...(selectedElement.rows || []), newRow] 
                            })
                          }}
                          className="btn"
                          style={{ padding: '0.3rem 0.6rem', fontSize: '0.8rem', flex: 1 }}
                        >
                          <Plus size={12} /> Add Row
                        </button>
                        <button
                          onClick={() => {
                            if (selectedElement.rows?.length > 1) {
                              updateElement(selectedElement.id, { 
                                rows: selectedElement.rows.slice(0, -1)
                              })
                            }
                          }}
                          className="btn"
                          style={{ padding: '0.3rem 0.6rem', fontSize: '0.8rem', flex: 1 }}
                          disabled={!selectedElement.rows || selectedElement.rows.length <= 1}
                        >
                          Remove Row
                        </button>
                      </div>

                      <div style={{ fontSize: '0.85rem', color: 'hsl(var(--muted-foreground))' }}>
                        Rows: {selectedElement.rows?.length || 0}, Columns: {selectedElement.maxcolumns || 3}
                      </div>
                    </>
                  )}

                  {selectedElement.type === 'footer' && (
                    <>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Text:</label>
                        <input
                          type="text"
                          value={selectedElement.text || ''}
                          onChange={(e) => updateElement(selectedElement.id, { text: e.target.value })}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                        />
                      </div>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.5rem' }}>Font Properties:</label>
                        <PropsEditor 
                          props={selectedElement.font} 
                          onChange={(font) => updateElement(selectedElement.id, { font })}
                        />
                      </div>
                    </>
                  )}

                  {selectedElement.type === 'spacer' && (
                    <>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Height (px):</label>
                        <input
                          type="number"
                          value={selectedElement.height || 20}
                          onChange={(e) => updateElement(selectedElement.id, { height: parseInt(e.target.value) || 20 })}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                          min="1"
                          max="200"
                        />
                      </div>
                    </>
                  )}

                  <button
                    onClick={() => deleteElement(selectedElement.id)}
                    className="btn"
                    style={{ 
                      background: 'hsl(var(--destructive))', 
                      color: 'hsl(var(--destructive-foreground))',
                      padding: '0.5rem',
                      marginTop: '1rem',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      justifyContent: 'center'
                    }}
                  >
                    <Trash2 size={16} /> Delete Component
                  </button>
                </div>
              )}
            </div>

            {/* JSON Output */}
            <div className="card" style={{ padding: '1rem', flex: 1 }}>
              <h3 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <FileText size={18} /> JSON Template
              </h3>
              <textarea
                readOnly
                value={JSON.stringify(jsonOutput, null, 2)}
                style={{
                  width: '100%',
                  height: '300px',
                  fontFamily: 'ui-monospace, "SF Mono", "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace',
                  fontSize: '0.75rem',
                  padding: '0.75rem',
                  resize: 'vertical',
                  background: 'hsl(var(--muted))',
                  color: 'hsl(var(--foreground))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '4px',
                  lineHeight: '1.4'
                }}
              />
            </div>
          </div>
        </div>
      </div>
      
      <style jsx>{`
        .dragging {
          transform: rotate(3deg) scale(0.95);
        }
        
        @media (max-width: 1200px) {
          .grid {
            grid-template-columns: 1fr !important;
            grid-template-rows: auto auto auto;
          }
        }
      `}</style>
    </div>
  )
}
