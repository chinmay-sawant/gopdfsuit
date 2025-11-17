import React, { useState, useRef, useMemo } from 'react'
import { Edit, Table, Type, Square, Minus, CheckSquare, FileText, Upload, Play, Copy, Sun, Moon, Trash2, Plus, GripVertical, Settings, Eye, Download, ChevronUp, ChevronDown, X, Image as ImageIcon } from 'lucide-react'
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

// Standard margin in points (1 inch = 72 points, 2 inches total for left+right)
const MARGIN = 72
const getUsableWidth = (pageWidth) => pageWidth - (2 * MARGIN)

// Component types matching the JSON template structure
const COMPONENT_TYPES = {
  title: { icon: Type, label: 'Title', defaultText: 'Document Title' },
  table: { icon: Table, label: 'Table', rows: 3, cols: 3 },
  footer: { icon: FileText, label: 'Footer', defaultText: 'Page footer text' },
  spacer: { icon: Minus, label: 'Spacer', height: 20 },
  checkbox: { icon: CheckSquare, label: 'Checkbox' },
  image: { icon: ImageIcon, label: 'Image' }
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

// Page border helpers
const parsePageBorder = (borderString) => {
  if (!borderString) return [0, 0, 0, 0]
  const parts = borderString.split(':')
  return [
    parseInt(parts[0]) || 0,
    parseInt(parts[1]) || 0,
    parseInt(parts[2]) || 0,
    parseInt(parts[3]) || 0
  ]
}

const formatPageBorder = (borders) => {
  return borders.join(':')
}

// Helper function to convert props to CSS style object
const getStyleFromProps = (propsString) => {
  const parsed = parseProps(propsString)
  const style = {
    fontSize: `${parsed.size}px`,
    textAlign: parsed.align,
    borderLeftWidth: `${parsed.borders[0]}px`,
    borderRightWidth: `${parsed.borders[1]}px`,
    borderTopWidth: `${parsed.borders[2]}px`,
    borderBottomWidth: `${parsed.borders[3]}px`,
    borderStyle: 'solid',
    borderColor: 'hsl(var(--border))'
  }
  
  // Apply font weight
  if (parsed.style[0] === '1') {
    style.fontWeight = 'bold'
  }
  
  // Apply italic
  if (parsed.style[1] === '1') {
    style.fontStyle = 'italic'
  }
  
  // Apply underline
  if (parsed.style[2] === '1') {
    style.textDecoration = 'underline'
  }
  
  return style
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

function DropZone({ index, onDrop, onAddComponent, isVisible, isToolboxDragging }) {
  const [isHovered, setIsHovered] = useState(false)

  const handleDragOver = (e) => {
    e.preventDefault()
    e.stopPropagation() // Prevent main canvas from handling drag over
    e.dataTransfer.dropEffect = isToolboxDragging ? 'copy' : 'move'
    setIsHovered(true)
  }

  const handleDragLeave = (e) => {
    e.stopPropagation() // Prevent main canvas from handling drag leave
    if (!e.currentTarget.contains(e.relatedTarget)) {
      setIsHovered(false)
    }
  }

  const handleDrop = (e) => {
    e.preventDefault()
    e.stopPropagation() // Prevent the main canvas from also handling this drop
    setIsHovered(false)
    const draggedData = e.dataTransfer.getData('text/plain')

    if (isToolboxDragging && COMPONENT_TYPES[draggedData]) {
      // Dropping from toolbox - add new component
      onAddComponent(draggedData, index)
    } else {
      // Dropping existing component for reordering
      onDrop(draggedData)
    }
  }

  // Don't render anything if not visible (not dragging)
  if (!isVisible && !isHovered) {
    return null
  }

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      style={{
        height: isHovered ? '40px' : '20px',
        width: '100%',
        background: isHovered
          ? 'hsl(var(--accent))'
          : 'hsl(var(--muted))',
        border: isHovered ? '2px dashed var(--secondary-color)' : '1px dashed hsl(var(--border))',
        borderRadius: '4px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        transition: 'all 0.2s ease',
        opacity: 1,
        margin: '5px 0',
        cursor: isHovered ? 'pointer' : 'default'
      }}
    >
      {isHovered && (
        <span style={{
          fontSize: '12px',
          color: 'hsl(var(--muted-foreground))',
          fontWeight: '500'
        }}>
          {isToolboxDragging ? 'Drop to add component' : 'Drop here'}
        </span>
      )}
    </div>
  )
}

function PageBorderControls({ borders, onChange }) {
  const updateBorder = (index, value) => {
    const newBorders = [...borders]
    newBorders[index] = Math.max(0, Math.min(10, value))
    onChange(newBorders)
  }

  const BorderControl = ({ label, index }) => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
      <label style={{ fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>{label}</label>
      <div style={{ display: 'flex', gap: '0.25rem' }}>
        <button
          onClick={() => updateBorder(index, borders[index] - 1)}
          disabled={borders[index] <= 0}
          style={{
            padding: '0.25rem 0.5rem',
            fontSize: '0.8rem',
            border: '1px solid hsl(var(--border))',
            background: 'hsl(var(--background))',
            color: 'hsl(var(--foreground))',
            borderRadius: '4px',
            cursor: borders[index] > 0 ? 'pointer' : 'not-allowed',
            opacity: borders[index] > 0 ? 1 : 0.5
          }}
        >
          −
        </button>
        <span style={{
          padding: '0.25rem 0.5rem',
          fontSize: '0.8rem',
          minWidth: '2rem',
          textAlign: 'center',
          background: 'hsl(var(--muted))',
          borderRadius: '4px'
        }}>
          {borders[index]}px
        </span>
        <button
          onClick={() => updateBorder(index, borders[index] + 1)}
          disabled={borders[index] >= 10}
          style={{
            padding: '0.25rem 0.5rem',
            fontSize: '0.8rem',
            border: '1px solid hsl(var(--border))',
            background: 'hsl(var(--background))',
            color: 'hsl(var(--foreground))',
            borderRadius: '4px',
            cursor: borders[index] < 10 ? 'pointer' : 'not-allowed',
            opacity: borders[index] < 10 ? 1 : 0.5
          }}
        >
          +
        </button>
      </div>
    </div>
  )

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
      <h5 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', color: 'hsl(var(--foreground))' }}>Page Borders</h5>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
        <BorderControl label="Left" index={0} />
        <BorderControl label="Right" index={1} />
        <BorderControl label="Top" index={2} />
        <BorderControl label="Bottom" index={3} />
      </div>

      {/* Quick Border Presets */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
        <label style={{ fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>Quick Set</label>
        <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
          {[
            { label: 'None', borders: [0, 0, 0, 0] },
            { label: 'All', borders: [1, 1, 1, 1] },
            { label: 'Box', borders: [1, 1, 1, 1] },
            { label: 'Bottom', borders: [0, 0, 1, 0] }
          ].map(({ label, borders: presetBorders }) => (
            <button
              key={label}
              onClick={() => onChange(presetBorders)}
              style={{
                padding: '0.25rem 0.5rem',
                border: '1px solid hsl(var(--border))',
                borderRadius: '4px',
                background: 'hsl(var(--muted))',
                color: 'hsl(var(--muted-foreground))',
                fontSize: '0.8rem',
                cursor: 'pointer',
                transition: 'all 0.2s ease'
              }}
            >
              {label}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

function PropsEditor({ props, onChange }) {
  const parsed = parseProps(props)

  const updateBorder = (index, value) => {
    const newBorders = [...parsed.borders]
    newBorders[index] = Math.max(0, Math.min(10, value))
    onChange(formatProps({ ...parsed, borders: newBorders }))
  }

  const BorderControls = ({ label, index }) => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
      <label style={{ fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>{label}</label>
      <div style={{ display: 'flex', gap: '0.25rem' }}>
        <button
          onClick={() => updateBorder(index, parsed.borders[index] - 1)}
          disabled={parsed.borders[index] <= 0}
          style={{
            padding: '0.25rem 0.5rem',
            fontSize: '0.8rem',
            border: '1px solid hsl(var(--border))',
            background: 'hsl(var(--background))',
            color: 'hsl(var(--foreground))',
            borderRadius: '4px',
            cursor: parsed.borders[index] > 0 ? 'pointer' : 'not-allowed',
            opacity: parsed.borders[index] > 0 ? 1 : 0.5
          }}
        >
          −
        </button>
        <span style={{
          padding: '0.25rem 0.5rem',
          fontSize: '0.8rem',
          minWidth: '2rem',
          textAlign: 'center',
          background: 'hsl(var(--muted))',
          borderRadius: '4px'
        }}>
          {parsed.borders[index]}px
        </span>
        <button
          onClick={() => updateBorder(index, parsed.borders[index] + 1)}
          disabled={parsed.borders[index] >= 10}
          style={{
            padding: '0.25rem 0.5rem',
            fontSize: '0.8rem',
            border: '1px solid hsl(var(--border))',
            background: 'hsl(var(--background))',
            color: 'hsl(var(--foreground))',
            borderRadius: '4px',
            cursor: parsed.borders[index] < 10 ? 'pointer' : 'not-allowed',
            opacity: parsed.borders[index] < 10 ? 1 : 0.5
          }}
        >
          +
        </button>
      </div>
    </div>
  )

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
      {/* Font Section */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', color: 'hsl(var(--foreground))' }}>Font</h4>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Family</label>
            <select
              value={parsed.font}
              onChange={(e) => onChange(formatProps({ ...parsed, font: e.target.value }))}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
                background: 'hsl(var(--background))',
                color: 'hsl(var(--foreground))',
                fontSize: '0.9rem'
              }}
            >
              <option value="font1">Font 1</option>
              <option value="font2">Font 2</option>
            </select>
          </div>
          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Size</label>
            <select
              value={parsed.size}
              onChange={(e) => onChange(formatProps({ ...parsed, size: parseInt(e.target.value) }))}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
                background: 'hsl(var(--background))',
                color: 'hsl(var(--foreground))',
                fontSize: '0.9rem'
              }}
            >
              {[8, 9, 10, 11, 12, 14, 16, 18, 20, 24, 28, 32, 36, 48, 72].map(size => (
                <option key={size} value={size}>{size}px</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Style Section */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', color: 'hsl(var(--foreground))' }}>Style</h4>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          {[
            { key: 0, label: 'B', title: 'Bold' },
            { key: 1, label: 'I', title: 'Italic' },
            { key: 2, label: 'U', title: 'Underline' }
          ].map(({ key, label, title }) => (
            <button
              key={key}
              onClick={() => {
                const newStyle = parsed.style.split('')
                newStyle[key] = newStyle[key] === '1' ? '0' : '1'
                onChange(formatProps({ ...parsed, style: newStyle.join('') }))
              }}
              style={{
                padding: '0.5rem 0.75rem',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
                background: parsed.style[key] === '1' ? 'hsl(var(--accent))' : 'hsl(var(--background))',
                color: parsed.style[key] === '1' ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))',
                fontSize: '0.9rem',
                fontWeight: parsed.style[key] === '1' ? '600' : '400',
                cursor: 'pointer',
                transition: 'all 0.2s ease'
              }}
              title={title}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Alignment Section */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', color: 'hsl(var(--foreground))' }}>Alignment</h4>
        <div style={{ display: 'flex', gap: '0.25rem' }}>
          {[
            { value: 'left', label: 'Left', icon: '⬅' },
            { value: 'center', label: 'Center', icon: '⬌' },
            { value: 'right', label: 'Right', icon: '➡' }
          ].map(({ value, label, icon }) => (
            <button
              key={value}
              onClick={() => onChange(formatProps({ ...parsed, align: value }))}
              style={{
                flex: 1,
                padding: '0.5rem',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
                background: parsed.align === value ? 'hsl(var(--accent))' : 'hsl(var(--background))',
                color: parsed.align === value ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))',
                fontSize: '0.9rem',
                cursor: 'pointer',
                transition: 'all 0.2s ease'
              }}
            >
              {icon} {label}
            </button>
          ))}
        </div>
      </div>

      {/* Borders Section */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', color: 'hsl(var(--foreground))' }}>Borders</h4>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
          <BorderControls label="Left" index={0} />
          <BorderControls label="Right" index={1} />
          <BorderControls label="Top" index={2} />
          <BorderControls label="Bottom" index={3} />
        </div>

        {/* Quick Border Presets */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label style={{ fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>Quick Set</label>
          <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
            {[
              { label: 'None', borders: [0, 0, 0, 0] },
              { label: 'All', borders: [1, 1, 1, 1] },
              { label: 'Box', borders: [1, 1, 1, 1] },
              { label: 'Bottom', borders: [0, 0, 1, 0] }
            ].map(({ label, borders }) => (
              <button
                key={label}
                onClick={() => onChange(formatProps({ ...parsed, borders }))}
                style={{
                  padding: '0.25rem 0.5rem',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '4px',
                  background: 'hsl(var(--muted))',
                  color: 'hsl(var(--muted-foreground))',
                  fontSize: '0.8rem',
                  cursor: 'pointer',
                  transition: 'all 0.2s ease'
                }}
              >
                {label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function ComponentItem({ element, index, isSelected, onSelect, onUpdate, onMove, onDelete, canMoveUp, canMoveDown, selectedCell, onCellSelect, onDragStart, onDragEnd, onDrop, isDragging, draggedType, handleCellDrop }) {
  const [isResizing, setIsResizing] = useState(false)

  const handleClick = (e) => {
    e.stopPropagation()
    onSelect(element.id)
    onCellSelect(null) // Clear cell selection when table is selected
  }

  const handleCellClick = (rowIdx, colIdx, e) => {
    if (e) e.stopPropagation()
    onSelect(element.id)
    onCellSelect({ rowIdx, colIdx })
  }

  const handleDragStart = (e) => {
    e.dataTransfer.setData('text/plain', element.id)
    e.dataTransfer.effectAllowed = 'move'
    onDragStart(element.id)
  }

  const handleDragEnd = () => {
    onDragEnd()
  }

  const handleDragOver = (e) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  const handleDrop = (e) => {
    e.preventDefault()
    const draggedId = e.dataTransfer.getData('text/plain')
    if (draggedId !== element.id) {
      onDrop(draggedId, element.id)
    }
  }

  const renderContent = () => {
    switch (element.type) {
      case 'title':
        const titleStyle = getStyleFromProps(element.props)
        return (
          <div style={{ 
            padding: '10px',
            borderRadius: '4px',
            minHeight: '40px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderLeft: `${titleStyle.borderLeftWidth} ${titleStyle.borderStyle} ${titleStyle.borderColor}`,
            borderRight: `${titleStyle.borderRightWidth} ${titleStyle.borderStyle} ${titleStyle.borderColor}`,
            borderTop: `${titleStyle.borderTopWidth} ${titleStyle.borderStyle} ${titleStyle.borderColor}`,
            borderBottom: `${titleStyle.borderBottomWidth} ${titleStyle.borderStyle} ${titleStyle.borderColor}`
          }}>
            <input
              type="text"
              value={element.text || 'Document Title'}
              onChange={(e) => onUpdate({ text: e.target.value })}
              style={{
                width: '100%',
                border: 'none',
                background: 'transparent',
                color: 'hsl(var(--foreground))',
                outline: 'none',
                fontSize: titleStyle.fontSize,
                textAlign: titleStyle.textAlign,
                fontWeight: titleStyle.fontWeight,
                fontStyle: titleStyle.fontStyle,
                textDecoration: titleStyle.textDecoration
              }}
              placeholder="Document Title"
            />
          </div>
        )
      case 'table':
        // Resize state helpers (persisted in element.columnwidths / element.rowheights)
        const handleColumnResizeStart = (e, colIdx) => {
          e.preventDefault()
          e.stopPropagation()
          const startX = e.clientX
          const tableEl = e.currentTarget.closest('table')
          const tablePixelWidth = tableEl ? tableEl.getBoundingClientRect().width : 600
          const startWidths = element.columnwidths && element.columnwidths.length === element.maxcolumns
            ? [...element.columnwidths]
            : Array(element.maxcolumns).fill(1 / element.maxcolumns)
          
          const onMouseMove = (me) => {
            const dx = me.clientX - startX
            // Convert pixel delta to weight delta
            const weightDelta = dx / tablePixelWidth
            const newWidths = [...startWidths]
            newWidths[colIdx] = Math.max(0.05, startWidths[colIdx] + weightDelta)
            // Adjust next column opposite to keep total width stable
            if (colIdx < newWidths.length - 1) {
              newWidths[colIdx+1] = Math.max(0.05, startWidths[colIdx+1] - weightDelta)
            }
            // Normalize to sum to 1
            const sum = newWidths.reduce((a,b)=>a+b,0)
            const normalized = newWidths.map(w => w / sum)
            onUpdate({ columnwidths: normalized })
          }
          const onMouseUp = () => {
            window.removeEventListener('mousemove', onMouseMove)
            window.removeEventListener('mouseup', onMouseUp)
          }
          window.addEventListener('mousemove', onMouseMove)
          window.addEventListener('mouseup', onMouseUp)
        }
        const handleRowResizeStart = (e, rowIdx) => {
          e.preventDefault()
          e.stopPropagation()
          const startY = e.clientY
          const startHeights = element.rowheights && element.rowheights.length === element.rows.length
            ? [...element.rowheights]
            : Array(element.rows.length).fill(1)
          const baseHeight = 25 // Base row height in pixels
          
          const onMouseMove = (me) => {
            const dy = me.clientY - startY
            // Scale factor: how much to change the multiplier based on pixel movement
            const scaleDelta = dy / baseHeight
            const newHeights = [...startHeights]
            newHeights[rowIdx] = Math.max(0.5, Math.min(5, startHeights[rowIdx] + scaleDelta))
            onUpdate({ rowheights: newHeights })
          }
          const onMouseUp = () => {
            window.removeEventListener('mousemove', onMouseMove)
            window.removeEventListener('mouseup', onMouseUp)
          }
          window.addEventListener('mousemove', onMouseMove)
          window.addEventListener('mouseup', onMouseUp)
        }
        const colWeights = element.columnwidths && element.columnwidths.length === element.maxcolumns
          ? element.columnwidths
          : Array(element.maxcolumns).fill(1 / element.maxcolumns)
        const rowScales = element.rowheights && element.rowheights.length === element.rows.length
          ? element.rowheights
          : Array(element.rows.length).fill(1)
        return (
          <div style={{ borderRadius: '4px', padding: '10px' }}>
            <table style={{ borderCollapse: 'separate', width: '100%', borderSpacing: '0', tableLayout: 'fixed' }}>
              <colgroup>
                {colWeights.map((w,i)=>(<col key={i} style={{ width: `${(w*100).toFixed(3)}%` }} />))}
              </colgroup>
              <tbody>
                {element.rows?.map((row, rowIdx) => (
                  <tr key={rowIdx} style={{ height: `${rowScales[rowIdx]*25}px`, position: 'relative' }}>
                    {row.row?.map((cell, colIdx) => {
                      const cellStyle = getStyleFromProps(cell.props)
                      const isCellSelected = selectedCell && selectedCell.rowIdx === rowIdx && selectedCell.colIdx === colIdx
                      const computedRowHeight = `${rowScales[rowIdx]*25}px`
                      const tdStyle = {
                        borderLeft: `${cellStyle.borderLeftWidth} ${cellStyle.borderStyle} ${cellStyle.borderColor}`,
                        borderRight: `${cellStyle.borderRightWidth} ${cellStyle.borderStyle} ${cellStyle.borderColor}`,
                        borderTop: `${cellStyle.borderTopWidth} ${cellStyle.borderStyle} ${cellStyle.borderColor}`,
                        borderBottom: `${cellStyle.borderBottomWidth} ${cellStyle.borderStyle} ${cellStyle.borderColor}`,
                        padding: '4px 8px',
                        minWidth: '80px',
                        minHeight: '24px',
                        height: computedRowHeight,
                        verticalAlign: 'middle',
                        overflow: 'hidden',
                        backgroundColor: isCellSelected ? 'hsl(var(--accent))' : 'transparent',
                        cursor: 'pointer',
                        position: 'relative'
                      }
                      const inputStyle = {
                        fontSize: cellStyle.fontSize,
                        textAlign: cellStyle.textAlign,
                        fontWeight: cellStyle.fontWeight,
                        fontStyle: cellStyle.fontStyle,
                        textDecoration: cellStyle.textDecoration,
                        width: '100%',
                        height: '100%',
                        border: 'none',
                        background: 'transparent',
                        padding: '2px',
                        color: 'hsl(var(--foreground))',
                        outline: 'none'
                      }
                      return (
                        <td
                          key={colIdx}
                          style={tdStyle}
                          onClick={(e) => handleCellClick(rowIdx, colIdx, e)}
                          onDragOver={(e) => {
                            if (draggedType === 'checkbox' || draggedType === 'image') {
                              e.preventDefault()
                              e.stopPropagation()
                            }
                          }}
                          onDrop={(e) => {
                            e.preventDefault()
                            e.stopPropagation()
                            const draggedData = e.dataTransfer.getData('text/plain')
                            if (draggedData === 'checkbox' || draggedData === 'image') {
                            handleCellDrop(element, onUpdate, rowIdx, colIdx, draggedData)
                            }
                          }}
                          className={(draggedType === 'checkbox' || draggedType === 'image') ? 'drop-target' : ''}
                        >
                          {cell.chequebox !== undefined ? (
                            <input 
                              type="checkbox" 
                              checked={cell.chequebox} 
                              onChange={(e) => {
                                e.stopPropagation()
                                const newRows = [...element.rows]
                                newRows[rowIdx].row[colIdx] = { 
                                  ...newRows[rowIdx].row[colIdx], 
                                  chequebox: e.target.checked 
                                }
                                onUpdate({ rows: newRows })
                              }}
                              onFocus={() => handleCellClick(rowIdx, colIdx)}
                              onClick={(e) => {
                                e.stopPropagation()
                                handleCellClick(rowIdx, colIdx)
                              }}
                              style={inputStyle}
                            />
                          ) : cell.image !== undefined ? (
                            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px', padding: '4px' }}>
                              {cell.image.imagedata ? (
                                <img 
                                  src={cell.image.imagedata.startsWith('data:') ? cell.image.imagedata : `data:image/png;base64,${cell.image.imagedata}`}
                                  alt={cell.image.imagename || 'Cell Image'}
                                  style={{ 
                                    maxWidth: '100%', 
                                    maxHeight: cell.image.height || 80,
                                    objectFit: 'contain'
                                  }}
                                />
                              ) : (
                                <div style={{ 
                                  display: 'flex', 
                                  flexDirection: 'column', 
                                  alignItems: 'center',
                                  padding: '8px',
                                  fontSize: '10px',
                                  color: 'hsl(var(--muted-foreground))'
                                }}>
                                  <ImageIcon size={16} />
                                  <span>No image</span>
                                </div>
                              )}
                            </div>
                          ) : (
                            <input
                              type="text"
                              value={cell.text || ''}
                              onChange={(e) => {
                                e.stopPropagation()
                                const newRows = [...element.rows]
                                newRows[rowIdx].row[colIdx] = { 
                                  ...newRows[rowIdx].row[colIdx], 
                                  text: e.target.value 
                                }
                                onUpdate({ rows: newRows })
                              }}
                              onFocus={() => handleCellClick(rowIdx, colIdx)}
                              onClick={(e) => {
                                e.stopPropagation()
                                handleCellClick(rowIdx, colIdx)
                              }}
                              style={inputStyle}
                            />
                          )}
                          {/* Column resize handle (except last cell) */}
                          {colIdx < (element.maxcolumns - 1) && (
                            <div
                              onMouseDown={(e)=>handleColumnResizeStart(e,colIdx)}
                              style={{
                                position: 'absolute',
                                top: 0,
                                right: '-3px',
                                width: '6px',
                                height: '100%',
                                cursor: 'col-resize',
                                zIndex: 5,
                                userSelect: 'none',
                                background: 'transparent'
                              }}
                              onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(59, 130, 246, 0.3)'}
                              onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                            />
                          )}
                          {/* Row resize handle (only one per row at bottom spanning first cell) */}
                          {colIdx === 0 && rowIdx < element.rows.length - 1 && (
                            <div
                              onMouseDown={(e)=>handleRowResizeStart(e,rowIdx)}
                              style={{
                                position: 'absolute',
                                bottom: '-3px',
                                left: 0,
                                width: '100%',
                                height: '6px',
                                cursor: 'row-resize',
                                zIndex: 4,
                                userSelect: 'none',
                                background: 'transparent'
                              }}
                              onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(59, 130, 246, 0.3)'}
                              onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                            />
                          )}
                        </td>
                      )
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      case 'footer':
        const footerStyle = getStyleFromProps(element.props)
        return (
          <div style={{ 
            padding: '10px',
            borderRadius: '4px',
            minHeight: '30px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderLeft: `${footerStyle.borderLeftWidth} ${footerStyle.borderStyle} ${footerStyle.borderColor}`,
            borderRight: `${footerStyle.borderRightWidth} ${footerStyle.borderStyle} ${footerStyle.borderColor}`,
            borderTop: `${footerStyle.borderTopWidth} ${footerStyle.borderStyle} ${footerStyle.borderColor}`,
            borderBottom: `${footerStyle.borderBottomWidth} ${footerStyle.borderStyle} ${footerStyle.borderColor}`
          }}>
            <input
              type="text"
              value={element.text || 'Page footer text'}
              onChange={(e) => onUpdate({ text: e.target.value })}
              style={{
                width: '100%',
                border: 'none',
                background: 'transparent',
                color: 'hsl(var(--foreground))',
                outline: 'none',
                fontSize: footerStyle.fontSize,
                textAlign: footerStyle.textAlign,
                fontWeight: footerStyle.fontWeight,
                fontStyle: footerStyle.fontStyle,
                textDecoration: footerStyle.textDecoration
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
      case 'image':
        return (
          <div style={{ 
            padding: '10px',
            borderRadius: '4px',
            minHeight: '100px',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            border: '2px dashed hsl(var(--border))',
            background: 'hsl(var(--muted))'
          }}>
            {element.imagedata ? (
              <div style={{ width: '100%', textAlign: 'center' }}>
                <img 
                  src={`data:image/png;base64,${element.imagedata}`}
                  alt={element.imagename || 'Image'}
                  style={{ 
                    maxWidth: '100%', 
                    maxHeight: element.height || 200,
                    objectFit: 'contain',
                    borderRadius: '4px'
                  }}
                />
                <div style={{ marginTop: '8px', fontSize: '0.85rem', color: 'hsl(var(--muted-foreground))' }}>
                  {element.imagename || 'Uploaded Image'}
                </div>
              </div>
            ) : (
              <div style={{ textAlign: 'center' }}>
                <ImageIcon size={32} style={{ color: 'hsl(var(--muted-foreground))', marginBottom: '8px' }} />
                <div style={{ fontSize: '0.9rem', color: 'hsl(var(--muted-foreground))' }}>
                  No image selected
                </div>
                <div style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))', marginTop: '4px' }}>
                  Select an image from properties
                </div>
              </div>
            )}
          </div>
        )
      default:
        return null
    }
  }

  return (
    <div 
      onClick={handleClick}
      draggable
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      style={{
        position: 'relative',
        margin: '10px 0',
        padding: isSelected && element.type !== 'table' ? '8px' : '0',
        border: isSelected && element.type !== 'table' ? '2px solid var(--secondary-color)' : '2px solid transparent',
        borderRadius: element.type === 'table' ? '0' : '6px',
        cursor: isDragging ? 'grabbing' : 'grab',
        background: isSelected && element.type !== 'table' ? 'hsl(var(--accent))' : 'transparent',
        boxShadow: isSelected && element.type === 'table' ? '0 0 0 2px var(--secondary-color)' : 'none',
        transition: 'all 0.2s ease',
        opacity: isDragging ? 0.5 : 1
      }}
    >
      {isSelected && (
        <div style={{
          position: 'absolute',
          top: '-35px',
          right: '0',
          display: 'flex',
          gap: '4px',
          background: 'hsl(var(--card))',
          border: '1px solid hsl(var(--border))',
          borderRadius: '8px',
          padding: '4px',
          zIndex: 10,
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)'
        }}>
          <button
            onClick={(e) => { e.stopPropagation(); onMove(index, 'up') }}
            disabled={!canMoveUp}
            style={{ 
              padding: '6px',
              border: 'none',
              borderRadius: '6px',
              background: canMoveUp ? 'hsl(var(--muted))' : 'hsl(var(--muted))',
              color: canMoveUp ? 'hsl(var(--foreground))' : 'hsl(var(--muted-foreground))',
              cursor: canMoveUp ? 'pointer' : 'not-allowed',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s ease',
              opacity: canMoveUp ? 1 : 0.5
            }}
            title="Move Up"
          >
            <ChevronUp size={14} />
          </button>
          <button
            onClick={(e) => { e.stopPropagation(); onMove(index, 'down') }}
            disabled={!canMoveDown}
            style={{ 
              padding: '6px',
              border: 'none',
              borderRadius: '6px',
              background: canMoveDown ? 'hsl(var(--muted))' : 'hsl(var(--muted))',
              color: canMoveDown ? 'hsl(var(--foreground))' : 'hsl(var(--muted-foreground))',
              cursor: canMoveDown ? 'pointer' : 'not-allowed',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s ease',
              opacity: canMoveDown ? 1 : 0.5
            }}
            title="Move Down"
          >
            <ChevronDown size={14} />
          </button>
          <div style={{ width: '1px', background: 'hsl(var(--border))', margin: '4px 0' }}></div>
          <button
            onClick={(e) => { e.stopPropagation(); onDelete(element.id) }}
            style={{ 
              padding: '6px',
              border: 'none',
              borderRadius: '6px',
              background: 'hsl(var(--destructive))',
              color: 'white',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s ease'
            }}
            title="Delete Component"
          >
            <X size={14} />
          </button>
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
  const [components, setComponents] = useState([]) // Combined ordered array for tables and spacers
  const [footer, setFooter] = useState(null)
  const [selectedId, setSelectedId] = useState(null)
  const [selectedCell, setSelectedCell] = useState(null)
  const [draggedType, setDraggedType] = useState(null)
  const [isDragOver, setIsDragOver] = useState(false)
  const [draggedComponentId, setDraggedComponentId] = useState(null)
  const [pdfUrl, setPdfUrl] = useState(null)
  const canvasRef = useRef(null)

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
    ? selectedElement.rows[selectedCell.rowIdx]?.row[selectedCell.colIdx] 
    : null
  const currentPageSize = PAGE_SIZES[config.page] || PAGE_SIZES.A4

  const generateId = () => Math.random().toString(36).substr(2, 9)

  const addElementAtPosition = (type, position) => {
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
        const usableWidth = getUsableWidth(currentPageSize.width)
        const newTable = {
          type: 'table',
          maxcolumns: 3,
          columnwidths: [1/3, 1/3, 1/3],
          rows: [
            {
              row: [
                { props: 'font1:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'font1:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'font1:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) }
              ]
            },
            {
              row: [
                { props: 'font1:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'font1:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'font1:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) }
              ]
            }
          ]
        }
        setComponents(prev => {
          const newComponents = [...prev]
          const insertIndex = title ? position - 1 : position
          newComponents.splice(insertIndex, 0, newTable)
          return newComponents
        })
        setSelectedId(`table-${title ? position - 1 : position}`)
        break
      case 'spacer':
        const newSpacer = {
          type: 'spacer',
          height: 20
        }
        setComponents(prev => {
          const newComponents = [...prev]
          const insertIndex = title ? position - 1 : position
          newComponents.splice(insertIndex, 0, newSpacer)
          return newComponents
        })
        setSelectedId(`spacer-${title ? position - 1 : position}`)
        break
      case 'image':
        const newImage = {
          type: 'image',
          imagename: '',
          imagedata: '',
          height: 200,
          width: 300
        }
        setComponents(prev => {
          const newComponents = [...prev]
          const insertIndex = title ? position - 1 : position
          newComponents.splice(insertIndex, 0, newImage)
          return newComponents
        })
        setSelectedId(`image-${title ? position - 1 : position}`)
        break
      case 'footer':
        if (!footer) {
          setFooter({
            props: 'font1:10:001:center:0:0:0:0',
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
    } else if (id.startsWith('table-') || id.startsWith('spacer-') || id.startsWith('image-')) {
      const idx = parseInt(id.split('-')[1])
      setComponents(prev => prev.map((component, i) => i === idx ? { ...component, ...updates } : component))
    }
  }

  const deleteElement = (id) => {
    if (id === 'title') {
      setTitle(null)
    } else if (id === 'footer') {
      setFooter(null)
    } else if (id.startsWith('table-') || id.startsWith('spacer-') || id.startsWith('image-')) {
      const idx = parseInt(id.split('-')[1])
      setComponents(prev => prev.filter((_, i) => i !== idx))
    }
    setSelectedId(null)
    setSelectedCell(null)
  }

  const moveElement = (index, direction) => {
    const element = allElements[index]
    if (!element) return

    let targetIndex
    if (direction === 'up' && index > (title ? 1 : 0)) {
      targetIndex = index - 1
    } else if (direction === 'down' && index < allElements.length - (footer ? 2 : 1)) {
      targetIndex = index + 1
    } else {
      return // Cannot move in that direction
    }

    // Calculate the actual index in the components array (excluding title and footer)
    const componentIndex = title ? index - 1 : index
    const targetComponentIndex = title ? targetIndex - 1 : targetIndex

    // Swap elements in the components array
    setComponents(prev => {
      const newComponents = [...prev]
      const temp = newComponents[componentIndex]
      newComponents[componentIndex] = newComponents[targetComponentIndex]
      newComponents[targetComponentIndex] = temp
      return newComponents
    })
  }

  const handleComponentDrop = (draggedId, targetId) => {
    const draggedIndex = allElements.findIndex(el => el.id === draggedId)
    const targetIndex = allElements.findIndex(el => el.id === targetId)

    if (draggedIndex === -1 || targetIndex === -1 || draggedIndex === targetIndex) return

    // Calculate the actual indices in the components array (excluding title and footer)
    const draggedComponentIndex = title ? draggedIndex - 1 : draggedIndex
    const targetComponentIndex = title ? targetIndex - 1 : targetIndex

    // Move the dragged element to the target position
    setComponents(prev => {
      const newComponents = [...prev]
      const [draggedElement] = newComponents.splice(draggedComponentIndex, 1)
      newComponents.splice(targetComponentIndex, 0, draggedElement)
      return newComponents
    })

    setDraggedComponentId(null)
  }

  const handleCellDrop = (element, onUpdate, rowIdx, colIdx, draggedType) => {
    if (draggedType !== 'checkbox' && draggedType !== 'image') return

    const newRows = [...element.rows]
    
    if (draggedType === 'checkbox') {
      // Update the table cell to contain a checkbox
      newRows[rowIdx].row[colIdx] = {
        ...newRows[rowIdx].row[colIdx],
        chequebox: false, // Start unchecked
        text: undefined, // Remove any existing text
        image: undefined // Remove any existing image
      }
    } else if (draggedType === 'image') {
      // Update the table cell to contain an image placeholder
      newRows[rowIdx].row[colIdx] = {
        ...newRows[rowIdx].row[colIdx],
        image: {
          imagename: '',
          imagedata: '',
          width: 100,
          height: 100
        },
        text: undefined, // Remove any existing text
        chequebox: undefined // Remove any existing checkbox
      }
    }
    
    onUpdate({ rows: newRows })
  }

  const handleCanvasClick = (e) => {
    if (e.target === canvasRef.current) {
      setSelectedId(null)
      setSelectedCell(null)
    }
  }

  const handleDrop = (e) => {
    e.preventDefault()
    setIsDragOver(false)
    setDraggedType(null)

    const type = e.dataTransfer.getData('text/plain')
    if (!type || !COMPONENT_TYPES[type]) return

    // If there are no components, add at the beginning
    if (allElements.length === 0) {
      addElementAtPosition(type, title ? 1 : 0)
    }
    // If there are components but no specific drop zone was targeted,
    // add at the end (this handles drops on empty canvas areas)
    else if (allElements.length > 0) {
      addElementAtPosition(type, allElements.length)
    }
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

        // Combine tables, spacers, and images into components array, preserving order
        const combinedComponents = []
        if (data.table) {
          data.table.forEach(table => combinedComponents.push({ ...table, type: 'table' }))
        }
        if (data.spacer) {
          data.spacer.forEach(spacer => combinedComponents.push({ ...spacer, type: 'spacer' }))
        }
        if (data.image) {
          data.image.forEach(image => combinedComponents.push({ ...image, type: 'image' }))
        }
        setComponents(combinedComponents)

        setFooter(data.footer || null)
        setSelectedId(null)
        setSelectedCell(null)
        
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

    // Separate components back into tables, spacers, and images for JSON output
    const tables = components.filter(comp => comp.type === 'table').map(({ type, ...rest }) => rest)
    const spacers = components.filter(comp => comp.type === 'spacer').map(({ type, ...rest }) => rest)
    const images = components.filter(comp => comp.type === 'image').map(({ type, ...rest }) => rest)

    if (tables.length > 0) output.table = tables
    if (spacers.length > 0) output.spacer = spacers
    if (images.length > 0) output.image = images
    if (footer) output.footer = footer
    return output
  }

  const jsonOutput = useMemo(() => getJsonOutput(), [config, title, components, footer])

  return (
    <div style={{ padding: '1rem 0', minHeight: '100vh' }}>
      <style>
        {`
          .drop-target {
            background-color: hsl(var(--accent)) !important;
            border: 2px dashed hsl(var(--primary)) !important;
            opacity: 0.8;
          }
        `}
      </style>
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

                {/* Page Border Controls */}
                <div style={{ marginTop: '1rem' }}>
                  <PageBorderControls
                    borders={parsePageBorder(config.pageBorder)}
                    onChange={(borders) => setConfig(prev => ({ ...prev, pageBorder: formatPageBorder(borders) }))}
                  />
                </div>

                {/* Page Alignment Controls */}
                <div style={{ marginTop: '1rem' }}>
                  <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Page Alignment</label>
                  <div style={{ display: 'flex', gap: '0.25rem' }}>
                    {[
                      { value: 1, label: 'Portrait' },
                      { value: 2, label: 'Landscape' }
                    ].map(({ value, label }) => (
                      <button
                        key={value}
                        onClick={() => setConfig(prev => ({ ...prev, pageAlignment: value }))}
                        style={{
                          flex: 1,
                          padding: '0.5rem',
                          border: '1px solid hsl(var(--border))',
                          borderRadius: '6px',
                          background: config.pageAlignment === value ? 'hsl(var(--accent))' : 'hsl(var(--background))',
                          color: config.pageAlignment === value ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))',
                          fontSize: '0.9rem',
                          cursor: 'pointer',
                          transition: 'all 0.2s ease'
                        }}
                      >
                        {label}
                      </button>
                    ))}
                  </div>
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
                    {/* Drop zone at the beginning */}
                    {allElements.length > 0 && (draggedComponentId !== null || draggedType !== null) && (
                      <DropZone
                        index={title ? 1 : 0}
                        onDrop={(draggedId) => handleComponentDrop(draggedId, title ? 1 : 0)}
                        onAddComponent={(type, position) => addElementAtPosition(type, title ? 1 : 0)}
                        isVisible={true}
                        isToolboxDragging={draggedType !== null}
                      />
                    )}

                    {allElements.map((element, index) => (
                      <React.Fragment key={element.id}>
                        <ComponentItem
                          element={element}
                          index={index}
                          isSelected={element.id === selectedId}
                          onSelect={setSelectedId}
                          onUpdate={(updates) => updateElement(element.id, updates)}
                          onMove={moveElement}
                          onDelete={deleteElement}
                          canMoveUp={index > (title ? 1 : 0)}
                          canMoveDown={index < allElements.length - (footer ? 2 : 1)}
                          selectedCell={selectedCell}
                          onCellSelect={setSelectedCell}
                          onDragStart={setDraggedComponentId}
                          onDragEnd={() => setDraggedComponentId(null)}
                          onDrop={handleComponentDrop}
                          isDragging={draggedComponentId === element.id}
                          draggedType={draggedType}
                          handleCellDrop={handleCellDrop}
                        />
                        {index < allElements.length - 1 && (
                          <DropZone
                            index={index + 1}
                            onDrop={(draggedId) => handleComponentDrop(draggedId, index + 1)}
                            onAddComponent={(type, position) => addElementAtPosition(type, position)}
                            isVisible={draggedComponentId !== null || draggedType !== null}
                            isToolboxDragging={draggedType !== null}
                          />
                        )}
                      </React.Fragment>
                    ))}

                    {/* Drop zone at the end */}
                    {allElements.length > 0 && (draggedComponentId !== null || draggedType !== null) && (
                      <DropZone
                        index={allElements.length}
                        onDrop={(draggedId) => handleComponentDrop(draggedId, allElements.length)}
                        onAddComponent={(type, position) => addElementAtPosition(type, position)}
                        isVisible={true}
                        isToolboxDragging={draggedType !== null}
                      />
                    )}
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
                            if (isNaN(newCols) || newCols < 1 || newCols > 10) return

                            // Adjust column widths proportionally
                            let newWidths = []
                            const currentWidths = selectedElement.columnwidths || []
                            if (newCols === currentWidths.length) {
                              newWidths = currentWidths
                            } else if (newCols > currentWidths.length) {
                              const remain = newCols - currentWidths.length
                              const existingSum = currentWidths.reduce((a, b) => a + b, 0)
                              const newPartWidth = (1 - existingSum) / remain
                              const extra = Array(remain).fill(Math.max(0.01, newPartWidth))
                              newWidths = [...currentWidths, ...extra]
                            } else {
                              newWidths = currentWidths.slice(0, newCols)
                            }
                            
                            // Normalize to sum 1
                            const sum = newWidths.reduce((a,b)=>a+b,0) || 1
                            const normalizedWidths = newWidths.map(w => w / sum)

                            const updatedRows = selectedElement.rows?.map(row => {
                              const newRow = [...(row.row || [])]
                              while (newRow.length < newCols) {
                                newRow.push({ props: 'font1:12:000:left:1:1:1:1', text: '' })
                              }
                              if (newRow.length > newCols) {
                                newRow.splice(newCols)
                              }
                              
                              // Update widths for all cells in the row
                              const usableWidth = getUsableWidth(currentPageSize.width)
                              const updatedCells = newRow.map((cell, colIdx) => ({
                                ...cell,
                                width: usableWidth * normalizedWidths[colIdx]
                              }))

                              return { row: updatedCells }
                            })
                            
                            updateElement(selectedElement.id, { maxcolumns: newCols, rows: updatedRows, columnwidths: normalizedWidths })
                          }}
                          style={{ flex: 1, padding: '0.25rem' }}
                        />
                      </div>
                      
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          onClick={() => {
                            const usableWidth = getUsableWidth(currentPageSize.width)
                            const newRow = { 
                              row: Array(selectedElement.maxcolumns || 3).fill().map((_, i) => ({
                                props: 'font1:12:000:left:1:1:1:1',
                                text: '',
                                width: usableWidth * (selectedElement.columnwidths ? selectedElement.columnwidths[i] : 1 / selectedElement.maxcolumns)
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

                      {/* Column Widths Controls */}
                      <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Column Widths (weights)</div>
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(60px, 1fr))', gap: '0.5rem' }}>
                          {Array.from({ length: selectedElement.maxcolumns }).map((_, colIdx) => {
                            const colWeights = selectedElement.columnwidths && selectedElement.columnwidths.length === selectedElement.maxcolumns
                              ? selectedElement.columnwidths
                              : Array(selectedElement.maxcolumns).fill(1 / selectedElement.maxcolumns)
                            return (
                              <div key={colIdx} style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                                <label style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>Col {colIdx + 1}</label>
                                <input
                                  type="number"
                                  step="0.1"
                                  min="0.1"
                                  value={(colWeights[colIdx] * 100).toFixed(1)}
                                  onChange={(e) => {
                                    const newWeights = [...colWeights]
                                    newWeights[colIdx] = Math.max(0.1, parseFloat(e.target.value) || 0) / 100
                                    const sum = newWeights.reduce((a,b)=>a+b,0)
                                    const normalized = newWeights.map(w => w / sum)

                                    const usableWidth = getUsableWidth(currentPageSize.width)
                                    const updatedRows = selectedElement.rows.map(row => ({
                                      ...row,
                                      row: row.row.map((cell, cIdx) => ({
                                        ...cell,
                                        width: usableWidth * normalized[cIdx]
                                      }))
                                    }))

                                    updateElement(selectedElement.id, { columnwidths: normalized, rows: updatedRows })
                                  }}
                                  style={{ width: '100%', padding: '0.25rem', fontSize: '0.8rem' }}
                                />
                              </div>
                            )
                          })}
                        </div>
                        <button
                          onClick={() => {
                            const evenWeights = Array(selectedElement.maxcolumns).fill(1 / selectedElement.maxcolumns)
                            const usableWidth = getUsableWidth(currentPageSize.width)
                            const updatedRows = selectedElement.rows.map(row => ({
                              ...row,
                              row: row.row.map((cell, cIdx) => ({
                                ...cell,
                                width: usableWidth * evenWeights[cIdx]
                              }))
                            }))
                            updateElement(selectedElement.id, { columnwidths: evenWeights, rows: updatedRows })
                          }}
                          style={{
                            marginTop: '0.5rem',
                            padding: '0.25rem 0.5rem',
                            border: '1px solid hsl(var(--border))',
                            borderRadius: '4px',
                            background: 'hsl(var(--muted))',
                            color: 'hsl(var(--muted-foreground))',
                            fontSize: '0.75rem',
                            cursor: 'pointer'
                          }}
                        >
                          Reset to Equal
                        </button>
                      </div>

                      {/* Row Heights Controls */}
                      <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Row Heights (multipliers)</div>
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(60px, 1fr))', gap: '0.5rem' }}>
                          {selectedElement.rows?.map((row, rowIdx) => {
                            const rowScales = selectedElement.rowheights && selectedElement.rowheights.length === selectedElement.rows.length
                              ? selectedElement.rowheights
                              : Array(selectedElement.rows.length).fill(1)
                            return (
                              <div key={rowIdx} style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                                <label style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>Row {rowIdx + 1}</label>
                                <input
                                  type="number"
                                  step="0.1"
                                  min="0.5"
                                  max="5"
                                  value={rowScales[rowIdx].toFixed(1)}
                                  onChange={(e) => {
                                    const newHeights = [...rowScales]
                                    newHeights[rowIdx] = Math.max(0.5, Math.min(5, parseFloat(e.target.value) || 1))
                                    updateElement(selectedElement.id, { rowheights: newHeights })
                                  }}
                                  style={{ width: '100%', padding: '0.25rem', fontSize: '0.8rem' }}
                                />
                              </div>
                            )
                          })}
                        </div>
                        <button
                          onClick={() => {
                            const evenHeights = Array(selectedElement.rows.length).fill(1)
                            updateElement(selectedElement.id, { rowheights: evenHeights })
                          }}
                          style={{
                            marginTop: '0.5rem',
                            padding: '0.25rem 0.5rem',
                            border: '1px solid hsl(var(--border))',
                            borderRadius: '4px',
                            background: 'hsl(var(--muted))',
                            color: 'hsl(var(--muted-foreground))',
                            fontSize: '0.75rem',
                            cursor: 'pointer'
                          }}
                        >
                          Reset to Default
                        </button>
                      </div>

                      {selectedCell && selectedCellElement && (
                        <>
                          <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                            <div style={{ fontSize: '0.9rem', fontWeight: '500', marginBottom: '0.5rem' }}>
                              Cell Properties (Row {selectedCell.rowIdx + 1}, Column {selectedCell.colIdx + 1})
                            </div>
                            
                            {selectedCellElement.image !== undefined ? (
                              <>
                                <div style={{ marginBottom: '0.5rem' }}>
                                  <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Image:</label>
                                  <input
                                    type="file"
                                    accept="image/*"
                                    onChange={(e) => {
                                      const file = e.target.files[0]
                                      if (file) {
                                        const reader = new FileReader()
                                        reader.onload = (event) => {
                                          const newRows = [...selectedElement.rows]
                                          newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                            ...selectedCellElement,
                                            image: {
                                              ...selectedCellElement.image,
                                              imagename: file.name,
                                              imagedata: event.target.result
                                            }
                                          }
                                          updateElement(selectedElement.id, { rows: newRows })
                                        }
                                        reader.readAsDataURL(file)
                                      }
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                  />
                                </div>
                                <div style={{ marginBottom: '0.5rem' }}>
                                  <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Image Name:</label>
                                  <input
                                    type="text"
                                    value={selectedCellElement.image?.imagename || ''}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        image: {
                                          ...selectedCellElement.image,
                                          imagename: e.target.value
                                        }
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                  />
                                </div>
                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem', marginBottom: '0.5rem' }}>
                                  <div>
                                    <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Width:</label>
                                    <input
                                      type="number"
                                      value={selectedCellElement.image?.width || 100}
                                      onChange={(e) => {
                                        const newRows = [...selectedElement.rows]
                                        newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                          ...selectedCellElement,
                                          image: {
                                            ...selectedCellElement.image,
                                            width: parseFloat(e.target.value) || 100
                                          }
                                        }
                                        updateElement(selectedElement.id, { rows: newRows })
                                      }}
                                      style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                    />
                                  </div>
                                  <div>
                                    <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Height:</label>
                                    <input
                                      type="number"
                                      value={selectedCellElement.image?.height || 100}
                                      onChange={(e) => {
                                        const newRows = [...selectedElement.rows]
                                        newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                          ...selectedCellElement,
                                          image: {
                                            ...selectedCellElement.image,
                                            height: parseFloat(e.target.value) || 100
                                          }
                                        }
                                        updateElement(selectedElement.id, { rows: newRows })
                                      }}
                                      style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                    />
                                  </div>
                                </div>
                                {selectedCellElement.image?.imagedata && (
                                  <div style={{
                                    padding: '0.5rem',
                                    borderRadius: '4px',
                                    background: 'hsl(var(--muted))',
                                    fontSize: '0.85rem',
                                    color: 'hsl(var(--muted-foreground))'
                                  }}>
                                    Image loaded: {selectedCellElement.image.imagename || 'Unnamed'}
                                  </div>
                                )}
                              </>
                            ) : selectedCellElement.chequebox !== undefined ? (
                              <div style={{ marginBottom: '0.5rem' }}>
                                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.9rem' }}>
                                  <input
                                    type="checkbox"
                                    checked={selectedCellElement.chequebox}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                        ...selectedCellElement, 
                                        chequebox: e.target.checked 
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                  />
                                  Checked
                                </label>
                              </div>
                            ) : (
                              <>
                                <div style={{ marginBottom: '0.5rem' }}>
                                  <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Text:</label>
                                  <input
                                    type="text"
                                    value={selectedCellElement.text || ''}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                        ...selectedCellElement, 
                                        text: e.target.value 
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                  />
                                </div>
                                <div>
                                  <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.5rem' }}>Font Properties:</label>
                                  <PropsEditor 
                                    props={selectedCellElement.props} 
                                    onChange={(props) => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                        ...selectedCellElement, 
                                        props 
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                  />
                                </div>
                              </>
                            )}

                            {/* Cell Size Controls - applies to all cell types */}
                            <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                              <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Cell Size Override</div>
                              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                <div>
                                  <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Width (pts)</label>
                                  <input
                                    type="number"
                                    step="1"
                                    min="0"
                                    placeholder="Auto"
                                    value={selectedCellElement.width || ''}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      const value = parseFloat(e.target.value)
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        width: value > 0 ? value : undefined
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem' }}
                                  />
                                </div>
                                <div>
                                  <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Height (pts)</label>
                                  <input
                                    type="number"
                                    step="1"
                                    min="0"
                                    placeholder="Auto"
                                    value={selectedCellElement.height || ''}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      const value = parseFloat(e.target.value)
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        height: value > 0 ? value : undefined
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem' }}
                                  />
                                </div>
                              </div>
                              <button
                                onClick={() => {
                                  const newRows = [...selectedElement.rows]
                                  const { width, height, ...rest } = selectedCellElement
                                  newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = rest
                                  updateElement(selectedElement.id, { rows: newRows })
                                }}
                                style={{
                                  marginTop: '0.5rem',
                                  padding: '0.25rem 0.5rem',
                                  border: '1px solid hsl(var(--border))',
                                  borderRadius: '4px',
                                  background: 'hsl(var(--muted))',
                                  color: 'hsl(var(--muted-foreground))',
                                  fontSize: '0.75rem',
                                  cursor: 'pointer',
                                  width: '100%'
                                }}
                              >
                                Reset to Column/Row Defaults
                              </button>
                              <p style={{ 
                                fontSize: '0.75rem', 
                                color: 'hsl(var(--muted-foreground))', 
                                marginTop: '0.5rem',
                                marginBottom: 0,
                                lineHeight: 1.4
                              }}>
                                Leave empty to use table's column width and row height. Set values to override for this specific cell.
                              </p>
                            </div>
                          </div>
                        </>
                      )}
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
                          props={selectedElement.props} 
                          onChange={(props) => updateElement(selectedElement.id, { props })}
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

                  {selectedElement.type === 'image' && (
                    <>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Select Image:</label>
                        <input
                          type="file"
                          accept="image/*"
                          onChange={(e) => {
                            const file = e.target.files?.[0]
                            if (file) {
                              const reader = new FileReader()
                              reader.onloadend = () => {
                                const base64String = reader.result.split(',')[1] // Remove data:image/...;base64, prefix
                                updateElement(selectedElement.id, { 
                                  imagename: file.name,
                                  imagedata: base64String
                                })
                              }
                              reader.readAsDataURL(file)
                            }
                          }}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                        />
                      </div>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Image Name:</label>
                        <input
                          type="text"
                          value={selectedElement.imagename || ''}
                          onChange={(e) => updateElement(selectedElement.id, { imagename: e.target.value })}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                          placeholder="Image name"
                        />
                      </div>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Width (px):</label>
                        <input
                          type="number"
                          value={selectedElement.width || 300}
                          onChange={(e) => updateElement(selectedElement.id, { width: parseInt(e.target.value) || 300 })}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                          min="50"
                          max="800"
                        />
                      </div>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Height (px):</label>
                        <input
                          type="number"
                          value={selectedElement.height || 200}
                          onChange={(e) => updateElement(selectedElement.id, { height: parseInt(e.target.value) || 200 })}
                          style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                          min="50"
                          max="800"
                        />
                      </div>
                      {selectedElement.imagedata && (
                        <div style={{ 
                          padding: '0.5rem',
                          background: 'hsl(var(--muted))',
                          borderRadius: '4px',
                          fontSize: '0.8rem',
                          color: 'hsl(var(--muted-foreground))'
                        }}>
                          Image loaded: {selectedElement.imagename || 'Unnamed'}
                        </div>
                      )}
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
