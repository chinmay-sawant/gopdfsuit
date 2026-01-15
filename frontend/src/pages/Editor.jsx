import React, { useState, useRef, useMemo, useEffect } from 'react'
import { Edit, Table, Type, Square, Minus, CheckSquare, FileText, Upload, Play, Copy, Sun, Moon, Trash2, Plus, GripVertical, Settings, Eye, Download, ChevronUp, ChevronDown, X, Image as ImageIcon, Circle, Check } from 'lucide-react'
import { useTheme } from '../theme'
import PdfPreview from '../components/PdfPreview'

// Default fonts - Standard PDF Type 1 fonts (guaranteed to work in all PDF readers)
const DEFAULT_FONTS = [
  // Helvetica family (sans-serif)
  { id: 'Helvetica', name: 'Helvetica', displayName: 'Helvetica' },
  { id: 'Helvetica-Bold', name: 'Helvetica-Bold', displayName: 'Helvetica Bold' },
  { id: 'Helvetica-Oblique', name: 'Helvetica-Oblique', displayName: 'Helvetica Italic' },
  { id: 'Helvetica-BoldOblique', name: 'Helvetica-BoldOblique', displayName: 'Helvetica Bold Italic' },
  // Times family (serif)
  { id: 'Times-Roman', name: 'Times-Roman', displayName: 'Times Roman' },
  { id: 'Times-Bold', name: 'Times-Bold', displayName: 'Times Bold' },
  { id: 'Times-Italic', name: 'Times-Italic', displayName: 'Times Italic' },
  { id: 'Times-BoldItalic', name: 'Times-BoldItalic', displayName: 'Times Bold Italic' },
  // Courier family (monospace)
  { id: 'Courier', name: 'Courier', displayName: 'Courier' },
  { id: 'Courier-Bold', name: 'Courier-Bold', displayName: 'Courier Bold' },
  { id: 'Courier-Oblique', name: 'Courier-Oblique', displayName: 'Courier Italic' },
  { id: 'Courier-BoldOblique', name: 'Courier-BoldOblique', displayName: 'Courier Bold Italic' },
  // Symbol and Decorative
  { id: 'Symbol', name: 'Symbol', displayName: 'Symbol' },
  { id: 'ZapfDingbats', name: 'ZapfDingbats', displayName: 'Zapf Dingbats' }
]

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
  radio: { icon: Circle, label: 'Radio Button' },
  text_input: { icon: Type, label: 'Text Input' },
  image: { icon: ImageIcon, label: 'Image' }
}

// Font style helpers
const parseProps = (propsString) => {
  if (!propsString) return { font: 'Helvetica', size: 12, style: '000', align: 'left', borders: [0, 0, 0, 0] }
  const parts = propsString.split(':')
  return {
    font: parts[0] || 'Helvetica',
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

// Helper function to get CSS font family from font name
const getFontFamily = (fontName) => {
  // Map PDF font names to CSS font families
  if (!fontName) return 'Helvetica, Arial, sans-serif'
  
  // Standard PDF Type 1 fonts mapping to CSS
  const fontMap = {
    // Helvetica family
    'Helvetica': 'Helvetica, Arial, sans-serif',
    'Helvetica-Bold': 'Helvetica, Arial, sans-serif',
    'Helvetica-Oblique': 'Helvetica, Arial, sans-serif',
    'Helvetica-BoldOblique': 'Helvetica, Arial, sans-serif',
    // Times family
    'Times-Roman': 'Times New Roman, Times, serif',
    'Times-Bold': 'Times New Roman, Times, serif',
    'Times-Italic': 'Times New Roman, Times, serif',
    'Times-BoldItalic': 'Times New Roman, Times, serif',
    // Courier family
    'Courier': 'Courier New, Courier, monospace',
    'Courier-Bold': 'Courier New, Courier, monospace',
    'Courier-Oblique': 'Courier New, Courier, monospace',
    'Courier-BoldOblique': 'Courier New, Courier, monospace',
    // Symbol fonts
    'Symbol': 'Symbol, serif',
    'ZapfDingbats': 'ZapfDingbats, Wingdings, serif',
  }
  
  // Return mapped font or the font name itself
  return fontMap[fontName] || `"${fontName}", sans-serif`
}

// Helper function to determine if font name implies bold/italic
const getFontStyleFromName = (fontName) => {
  if (!fontName) return { isBold: false, isItalic: false }
  const lower = fontName.toLowerCase()
  return {
    isBold: lower.includes('bold'),
    isItalic: lower.includes('oblique') || lower.includes('italic')
  }
}

// Helper function to convert props to CSS style object
const getStyleFromProps = (propsString) => {
  const parsed = parseProps(propsString)
  const fontStyles = getFontStyleFromName(parsed.font)
  
  const style = {
    fontSize: `${parsed.size}px`,
    textAlign: parsed.align,
    fontFamily: getFontFamily(parsed.font),
    borderLeftWidth: `${parsed.borders[0]}px`,
    borderRightWidth: `${parsed.borders[1]}px`,
    borderTopWidth: `${parsed.borders[2]}px`,
    borderBottomWidth: `${parsed.borders[3]}px`,
    borderStyle: 'solid',
    borderColor: '#333'
  }
  
  // Apply font weight - from style code OR font name
  if (parsed.style[0] === '1' || fontStyles.isBold) {
    style.fontWeight = 'bold'
  }
  
  // Apply italic - from style code OR font name
  if (parsed.style[1] === '1' || fontStyles.isItalic) {
    style.fontStyle = 'italic'
  }
  
  // Apply underline
  if (parsed.style[2] === '1') {
    style.textDecoration = 'underline'
  }
  
  return style
}

function Toolbar({ theme, setTheme, onLoadTemplate, onPreviewPDF, onCopyJSON, onDownloadPDF, onUploadFont }) {
  const [templateInput, setTemplateInput] = useState('')
  const fileInputRef = useRef(null)

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      onUploadFont(e.target.files[0])
      e.target.value = null // Reset input
    }
  }

  return (
    <div className="card" style={{ 
      marginBottom: '1rem', 
      padding: '0.75rem 1rem',
      position: 'sticky',
      top: '74px', // Increased to ensure it clears the navbar
      zIndex: 40,
      transition: 'all 0.2s ease',
      borderRadius: '0',
      borderLeft: 'none',
      borderRight: 'none',
      marginLeft: '-1rem',
      marginRight: '-1rem',
      width: 'calc(100% + 2rem)',
      boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
      background: 'hsl(var(--card))' // Ensure background is opaque
    }}>
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
            onClick={() => fileInputRef.current?.click()}
            className="btn"
            style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem' }}
            title="Upload Custom Font (.ttf/.otf)"
          >
            <Type size={14} /> Upload Font
          </button>
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileChange}
            accept=".ttf,.otf"
            style={{ display: 'none' }}
          />
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

function PropsEditor({ props, onChange, fonts = DEFAULT_FONTS }) {
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
              {fonts.map(font => (
                <option key={font.id} value={font.id}>{font.displayName}</option>
              ))}
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

function ComponentItem({ element, index, isSelected, onSelect, onUpdate, onMove, onDelete, canMoveUp, canMoveDown, selectedCell, onCellSelect, onDragStart, onDragEnd, onDrop, isDragging, draggedType, handleCellDrop, currentPageSize }) {
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
        // Title now uses an embedded table structure for logo + text support
        const MARGIN_TITLE = 72
        const getUsableWidthTitle = (pageWidth) => pageWidth - (2 * MARGIN_TITLE)
        const usableWidthForTitle = getUsableWidthTitle(currentPageSize.width)
        
        // Get or create the title table structure
        const titleTable = element.table || {
          maxcolumns: 3,
          columnwidths: [1, 2, 1],
          rows: [{
            row: [
              { props: 'Helvetica:12:000:left:0:0:0:0', text: '', image: null },
              { props: 'Helvetica:18:100:center:0:0:0:0', text: element.text || 'Document Title' },
              { props: 'Helvetica:12:000:right:0:0:0:0', text: '' }
            ]
          }]
        }
        
        // Helper to get normalized column weight for title table
        const getNormalizedColWeightTitle = (colIdx) => {
          const rawWeights = titleTable.columnwidths && titleTable.columnwidths.length === titleTable.maxcolumns
            ? titleTable.columnwidths
            : Array(titleTable.maxcolumns).fill(1)
          const total = rawWeights.reduce((sum, w) => sum + w, 0)
          return rawWeights[colIdx] / total
        }
        
        // Per-cell width resize handler for title table
        const handleTitleCellWidthResizeStart = (e, rowIdx, colIdx) => {
          e.preventDefault()
          e.stopPropagation()
          const startX = e.clientX
          const cell = titleTable.rows[rowIdx].row[colIdx]
          const startWidth = cell.width || (usableWidthForTitle * getNormalizedColWeightTitle(colIdx))
          
          const onMouseMove = (me) => {
            const dx = me.clientX - startX
            let newWidth = Math.max(50, startWidth + dx)
            const widthChange = newWidth - startWidth
            
            const newRows = [...titleTable.rows]
            newRows[rowIdx] = {
              ...newRows[rowIdx],
              row: newRows[rowIdx].row.map((c, idx) => 
                idx === colIdx ? { ...c, width: newWidth } : c
              )
            }
            
            // Redistribute width to adjacent columns
            if (colIdx < titleTable.maxcolumns - 1) {
              const nextCell = newRows[rowIdx].row[colIdx + 1]
              const nextWidth = nextCell.width || (usableWidthForTitle * getNormalizedColWeightTitle(colIdx + 1))
              const newNextWidth = nextWidth - widthChange
              newRows[rowIdx].row[colIdx + 1] = { ...nextCell, width: Math.max(50, newNextWidth) }
            }
            
            onUpdate({ table: { ...titleTable, rows: newRows } })
          }
          const onMouseUp = () => {
            window.removeEventListener('mousemove', onMouseMove)
            window.removeEventListener('mouseup', onMouseUp)
          }
          window.addEventListener('mousemove', onMouseMove)
          window.addEventListener('mouseup', onMouseUp)
        }
        
        // Per-cell height resize handler for title table
        const handleTitleCellHeightResizeStart = (e, rowIdx, colIdx) => {
          e.preventDefault()
          e.stopPropagation()
          const startY = e.clientY
          const cell = titleTable.rows[rowIdx].row[colIdx]
          const startHeight = cell.height || 50
          
          const onMouseMove = (me) => {
            const dy = me.clientY - startY
            const newHeight = Math.max(30, startHeight + dy)
            
            // Update all cells in this row to same height
            const newRows = [...titleTable.rows]
            newRows[rowIdx] = {
              ...newRows[rowIdx],
              row: newRows[rowIdx].row.map(c => ({ ...c, height: newHeight }))
            }
            onUpdate({ table: { ...titleTable, rows: newRows } })
          }
          const onMouseUp = () => {
            window.removeEventListener('mousemove', onMouseMove)
            window.removeEventListener('mouseup', onMouseUp)
          }
          window.addEventListener('mousemove', onMouseMove)
          window.addEventListener('mouseup', onMouseUp)
        }
        
        // Handle image upload for title cells
        const handleTitleImageUpload = (rowIdx, colIdx, file) => {
          const reader = new FileReader()
          reader.onload = (e) => {
            const imageData = e.target.result
            const newRows = [...titleTable.rows]
            newRows[rowIdx] = {
              ...newRows[rowIdx],
              row: newRows[rowIdx].row.map((c, idx) => 
                idx === colIdx ? { 
                  ...c, 
                  image: { 
                    imagename: file.name, 
                    imagedata: imageData,
                    width: 100,
                    height: 50
                  },
                  text: '' // Clear text when image is added
                } : c
              )
            }
            onUpdate({ table: { ...titleTable, rows: newRows } })
          }
          reader.readAsDataURL(file)
        }
        
        // Helper to update a specific title table cell with proper immutable updates
        const updateTitleTableCell = (rowIdx, colIdx, cellUpdates) => {
          const newRows = titleTable.rows.map((row, rIdx) => 
            rIdx === rowIdx
              ? {
                  ...row,
                  row: row.row.map((c, cIdx) =>
                    cIdx === colIdx
                      ? { ...c, ...cellUpdates }
                      : c
                  )
                }
              : row
          )
          onUpdate({ table: { ...titleTable, rows: newRows } })
        }
        
        return (
          <div style={{ 
            borderRadius: '4px',
            background: 'white',
            overflowX: 'auto'
          }}>
            <table style={{ borderCollapse: 'collapse', borderSpacing: '0', tableLayout: 'fixed', width: '100%' }}>
              <tbody>
                {titleTable.rows?.map((row, rowIdx) => (
                  <tr key={rowIdx} style={{ position: 'relative' }}>
                    {row.row?.map((cell, colIdx) => {
                      const cellStyle = getStyleFromProps(cell.props)
                      const isCellSelected = selectedCell && selectedCell.rowIdx === rowIdx && selectedCell.colIdx === colIdx
                      
                      const cellWidth = cell.width || (usableWidthForTitle * getNormalizedColWeightTitle(colIdx))
                      const cellHeight = cell.height || 50
                      
                      const hasBorder = cellStyle.borderLeftWidth !== '0px' || cellStyle.borderRightWidth !== '0px' || 
                                        cellStyle.borderTopWidth !== '0px' || cellStyle.borderBottomWidth !== '0px'
                      
                      // Determine background color for title cells
                      const titleCellBgColor = isCellSelected ? '#e3f2fd' : (cell.bgcolor || element.bgcolor || '#fff')
                      
                      return (
                        <td
                          key={colIdx}
                          style={{
                            borderLeft: hasBorder ? `${cellStyle.borderLeftWidth} solid #333` : 'none',
                            borderRight: hasBorder ? `${cellStyle.borderRightWidth} solid #333` : 'none',
                            borderTop: hasBorder ? `${cellStyle.borderTopWidth} solid #333` : 'none',
                            borderBottom: hasBorder ? `${cellStyle.borderBottomWidth} solid #333` : 'none',
                            padding: '4px 8px',
                            width: `${cellWidth}px`,
                            height: `${cellHeight}px`,
                            minWidth: `${cellWidth}px`,
                            maxWidth: `${cellWidth}px`,
                            minHeight: '30px',
                            verticalAlign: 'middle',
                            overflow: 'hidden',
                            backgroundColor: titleCellBgColor,
                            cursor: 'pointer',
                            position: 'relative',
                            boxSizing: 'border-box'
                          }}
                          onClick={(e) => {
                            e.stopPropagation()
                            onSelect(element.id)
                            onCellSelect({ rowIdx, colIdx })
                          }}
                          onDragOver={(e) => {
                            if (draggedType === 'image') {
                              e.preventDefault()
                              e.stopPropagation()
                            }
                          }}
                          onDrop={(e) => {
                            e.preventDefault()
                            e.stopPropagation()
                            const files = e.dataTransfer.files
                            if (files.length > 0 && files[0].type.startsWith('image/')) {
                              handleTitleImageUpload(rowIdx, colIdx, files[0])
                            }
                          }}
                        >
                          {/* Cell content: image or text */}
                          {cell.image && cell.image.imagedata ? (
                            <div 
                              style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px', height: '100%', justifyContent: 'center' }}
                              onClick={(e) => {
                                e.stopPropagation()
                                onSelect(element.id)
                                onCellSelect({ rowIdx, colIdx })
                              }}
                            >
                              <img 
                                src={cell.image.imagedata.startsWith('data:') ? cell.image.imagedata : `data:image/png;base64,${cell.image.imagedata}`}
                                alt={cell.image.imagename || 'Logo'}
                                style={{ 
                                  maxWidth: '100%', 
                                  maxHeight: cellHeight - 10,
                                  objectFit: 'contain'
                                }}
                              />
                              <button
                                onClick={(e) => {
                                  e.stopPropagation()
                                  onSelect(element.id)
                                  onCellSelect({ rowIdx, colIdx })
                                  updateTitleTableCell(rowIdx, colIdx, { image: null })
                                }}
                                style={{
                                  position: 'absolute',
                                  top: '2px',
                                  right: '2px',
                                  background: 'rgba(255,0,0,0.7)',
                                  color: 'white',
                                  border: 'none',
                                  borderRadius: '50%',
                                  width: '16px',
                                  height: '16px',
                                  fontSize: '10px',
                                  cursor: 'pointer',
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'center'
                                }}
                                title="Remove image"
                              >
                                ×
                              </button>
                            </div>
                          ) : (
                            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px', height: '100%' }}>
                              <input
                                type="text"
                                value={cell.text || ''}
                                onChange={(e) => {
                                  e.stopPropagation()
                                  updateTitleTableCell(rowIdx, colIdx, { text: e.target.value })
                                }}
                                placeholder={colIdx === 0 ? 'Logo/Image' : colIdx === 1 ? 'Document Title' : 'Right Text'}
                                style={{
                                  width: '100%',
                                  flex: 1,
                                  border: 'none',
                                  background: 'transparent',
                                  color: cell.textcolor || element.textcolor || '#000',
                                  outline: 'none',
                                  fontSize: cellStyle.fontSize,
                                  textAlign: cellStyle.textAlign,
                                  fontWeight: cellStyle.fontWeight,
                                  fontStyle: cellStyle.fontStyle,
                                  textDecoration: cellStyle.textDecoration
                                }}
                                onClick={(e) => {
                                  e.stopPropagation()
                                  onSelect(element.id)
                                  onCellSelect({ rowIdx, colIdx })
                                }}
                              />
                              {colIdx === 0 && (
                                <label 
                                  style={{ 
                                    fontSize: '9px', 
                                    color: 'hsl(var(--muted-foreground))',
                                    cursor: 'pointer',
                                    padding: '2px 4px',
                                    background: 'hsl(var(--muted))',
                                    borderRadius: '4px'
                                  }}
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    onSelect(element.id)
                                    onCellSelect({ rowIdx, colIdx })
                                  }}
                                >
                                  <input
                                    type="file"
                                    accept="image/*"
                                    style={{ display: 'none' }}
                                    onChange={(e) => {
                                      if (e.target.files[0]) {
                                        handleTitleImageUpload(rowIdx, colIdx, e.target.files[0])
                                      }
                                    }}
                                  />
                                  + Add Logo
                                </label>
                              )}
                            </div>
                          )}
                          
                          {/* Width resize handle */}
                          {colIdx < titleTable.maxcolumns - 1 && (
                            <div
                              onMouseDown={(e) => handleTitleCellWidthResizeStart(e, rowIdx, colIdx)}
                              style={{
                                position: 'absolute',
                                right: 0,
                                top: 0,
                                bottom: 0,
                                width: '4px',
                                cursor: 'col-resize',
                                background: isCellSelected ? 'rgba(25, 118, 210, 0.3)' : 'transparent'
                              }}
                              title="Drag to resize width"
                            />
                          )}
                          
                          {/* Height resize handle */}
                          <div
                            onMouseDown={(e) => handleTitleCellHeightResizeStart(e, rowIdx, colIdx)}
                            style={{
                              position: 'absolute',
                              left: 0,
                              right: 0,
                              bottom: 0,
                              height: '4px',
                              cursor: 'row-resize',
                              background: isCellSelected ? 'rgba(25, 118, 210, 0.3)' : 'transparent'
                            }}
                            title="Drag to resize height"
                          />
                        </td>
                      )
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      case 'table':
        // Get page dimensions for width calculations
        const MARGIN = 72
        const getUsableWidth = (pageWidth) => pageWidth - (2 * MARGIN)
        
        // Use passed currentPageSize prop
        const usableWidthForTable = getUsableWidth(currentPageSize.width)
        
        // Helper to get normalized column weight
        const getNormalizedColWeight = (colIdx) => {
          const rawWeights = element.columnwidths && element.columnwidths.length === element.maxcolumns
            ? element.columnwidths
            : Array(element.maxcolumns).fill(1)
          const total = rawWeights.reduce((sum, w) => sum + w, 0)
          return rawWeights[colIdx] / total
        }
        
        // Per-cell width resize handler
        const handleCellWidthResizeStart = (e, rowIdx, colIdx) => {
          e.preventDefault()
          e.stopPropagation()
          const startX = e.clientX
          const cell = element.rows[rowIdx].row[colIdx]
          const startWidth = cell.width || (usableWidthForTable * getNormalizedColWeight(colIdx))
          
          const onMouseMove = (me) => {
            const dx = me.clientX - startX
            let newWidth = Math.max(50, startWidth + dx)
            const widthChange = newWidth - startWidth
            
            // Update only this specific cell's width
            const newRows = [...element.rows]
            newRows[rowIdx] = {
              ...newRows[rowIdx],
              row: newRows[rowIdx].row.map((c, idx) => 
                idx === colIdx ? { ...c, width: newWidth } : c
              )
            }
            
            // If this is column 0, redistribute the change across columns 1 and 2+
            if (colIdx === 0) {
              const numOtherCols = element.maxcolumns - 1
              const redistributePerCol = widthChange / numOtherCols
              
              newRows[rowIdx].row = newRows[rowIdx].row.map((c, idx) => {
                if (idx === 0) return c
                const currentWidth = c.width || (usableWidthForTable * getNormalizedColWeight(idx))
                const newColWidth = currentWidth - redistributePerCol
                return { ...c, width: Math.max(0, newColWidth) }
              })
            }
            // If this is a middle column (not first, not last), only subtract from next column
            // When expanding (positive widthChange), subtract from next column
            // When shrinking (negative widthChange), add space back to next column
            else if (colIdx > 0 && colIdx < element.maxcolumns - 1) {
              const nextCell = newRows[rowIdx].row[colIdx + 1]
              const nextWidth = nextCell.width || (usableWidthForTable * getNormalizedColWeight(colIdx + 1))
              const newNextWidth = nextWidth - widthChange
              // Always subtract the change from the next column (if expanding this column, next shrinks; if shrinking, next expands)
              newRows[rowIdx].row[colIdx + 1] = { ...nextCell, width: Math.max(0, newNextWidth) }
            }
            // Last column should not be resizable (handled in render)
            
            // Final safety check: ensure total doesn't exceed usable width
            const totalWidth = newRows[rowIdx].row.reduce((sum, c) => sum + (c.width || 0), 0)
            if (totalWidth > usableWidthForTable + 1) { // +1 for rounding tolerance
              // Proportionally scale down all cells to fit
              const scale = usableWidthForTable / totalWidth
              newRows[rowIdx].row = newRows[rowIdx].row.map(c => ({
                ...c,
                width: (c.width || 0) * scale
              }))
            }
            
            onUpdate({ rows: newRows })
          }
          const onMouseUp = () => {
            window.removeEventListener('mousemove', onMouseMove)
            window.removeEventListener('mouseup', onMouseUp)
          }
          window.addEventListener('mousemove', onMouseMove)
          window.addEventListener('mouseup', onMouseUp)
        }
        
        // Per-cell height resize handler
        const handleCellHeightResizeStart = (e, rowIdx, colIdx) => {
          e.preventDefault()
          e.stopPropagation()
          const startY = e.clientY
          const cell = element.rows[rowIdx].row[colIdx]
          const startHeight = cell.height || 25
          
          const onMouseMove = (me) => {
            const dy = me.clientY - startY
            const newHeight = Math.max(20, startHeight + dy)
            
            // Update only this specific cell's height
            const newRows = [...element.rows]
            newRows[rowIdx] = {
              ...newRows[rowIdx],
              row: newRows[rowIdx].row.map((c, idx) => 
                idx === colIdx ? { ...c, height: newHeight } : c
              )
            }
            onUpdate({ rows: newRows })
          }
          const onMouseUp = () => {
            window.removeEventListener('mousemove', onMouseMove)
            window.removeEventListener('mouseup', onMouseUp)
          }
          window.addEventListener('mousemove', onMouseMove)
          window.addEventListener('mouseup', onMouseUp)
        }
        // Normalize columnwidths so they represent fractions that sum to 1
        const rawColWidths = element.columnwidths && element.columnwidths.length === element.maxcolumns
          ? element.columnwidths
          : Array(element.maxcolumns).fill(1)
        const totalWeight = rawColWidths.reduce((sum, w) => sum + w, 0)
        const colWeights = rawColWidths.map(w => w / totalWeight)
        return (
          <div style={{ borderRadius: '4px', padding: '10px', overflowX: 'auto', background: 'white' }}>
            <table style={{ borderCollapse: 'collapse', borderSpacing: '0', tableLayout: 'fixed', width: '100%' }}>
              <tbody>
                {element.rows?.map((row, rowIdx) => (
                  <tr key={rowIdx} style={{ position: 'relative' }}>
                    {row.row?.map((cell, colIdx) => {
                      const cellStyle = getStyleFromProps(cell.props)
                      const isCellSelected = selectedCell && selectedCell.rowIdx === rowIdx && selectedCell.colIdx === colIdx
                      
                      // Use cell-specific width if available, otherwise fall back to column width
                      const cellWidth = cell.width || (usableWidthForTable * colWeights[colIdx])
                      const cellHeight = cell.height || 25
                      
                      // Determine background color: selection > cell bgcolor > table bgcolor > default white
                      const cellBgColor = isCellSelected 
                        ? '#e3f2fd' 
                        : (cell.bgcolor || element.bgcolor || '#fff')
                      
                      // Determine text color: cell textcolor > table textcolor > default black
                      const cellTextColor = cell.textcolor || element.textcolor || '#000'
                      
                      // Ensure borders are visible - use explicit border if cell has border props
                      const hasBorder = cellStyle.borderLeftWidth !== '0px' || cellStyle.borderRightWidth !== '0px' || 
                                        cellStyle.borderTopWidth !== '0px' || cellStyle.borderBottomWidth !== '0px'
                      const tdStyle = {
                        borderLeft: hasBorder ? `${cellStyle.borderLeftWidth} solid #333` : 'none',
                        borderRight: hasBorder ? `${cellStyle.borderRightWidth} solid #333` : 'none',
                        borderTop: hasBorder ? `${cellStyle.borderTopWidth} solid #333` : 'none',
                        borderBottom: hasBorder ? `${cellStyle.borderBottomWidth} solid #333` : 'none',
                        padding: '4px 8px',
                        width: `${cellWidth}px`,
                        height: `${cellHeight}px`,
                        minWidth: `${cellWidth}px`,
                        maxWidth: `${cellWidth}px`,
                        minHeight: '20px',
                        verticalAlign: 'middle',
                        overflow: 'hidden',
                        backgroundColor: cellBgColor,
                        cursor: 'pointer',
                        position: 'relative',
                        boxSizing: 'border-box',
                        flexShrink: 0
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
                        color: cellTextColor,
                        outline: 'none'
                      }
                      return (
                        <td
                          key={colIdx}
                          style={tdStyle}
                          onClick={(e) => handleCellClick(rowIdx, colIdx, e)}
                          onDragOver={(e) => {
                            if (draggedType === 'checkbox' || draggedType === 'image' || draggedType === 'radio' || draggedType === 'text_input') {
                              e.preventDefault()
                              e.stopPropagation()
                            }
                          }}
                          onDrop={(e) => {
                            e.preventDefault()
                            e.stopPropagation()
                            const draggedData = e.dataTransfer.getData('text/plain')
                            if (draggedData === 'checkbox' || draggedData === 'image' || draggedData === 'radio' || draggedData === 'text_input') {
                            handleCellDrop(element, onUpdate, rowIdx, colIdx, draggedData)
                            }
                          }}
                          className={(draggedType === 'checkbox' || draggedType === 'image' || draggedType === 'radio' || draggedType === 'text_input') ? 'drop-target' : ''}
                        >
                          {cell.form_field ? (
                            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: '2px', width: '100%' }}>
                              {cell.form_field.type === 'text' ? (
                                <input
                                  type="text"
                                  value={cell.form_field.value || ''}
                                  onChange={(e) => {
                                    e.stopPropagation()
                                    const newRows = [...element.rows]
                                    newRows[rowIdx].row[colIdx] = {
                                      ...newRows[rowIdx].row[colIdx],
                                      form_field: {
                                        ...cell.form_field,
                                        value: e.target.value
                                      }
                                    }
                                    onUpdate({ rows: newRows })
                                  }}
                                  placeholder={cell.form_field.name}
                                  style={{ 
                                    width: '100%', 
                                    height: '100%', 
                                    border: 'none', 
                                    borderRadius: '0',
                                    fontSize: '10px',
                                    padding: '4px',
                                    background: 'transparent',
                                    color: '#000'
                                  }}
                                  onFocus={() => handleCellClick(rowIdx, colIdx)}
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    handleCellClick(rowIdx, colIdx)
                                  }}
                                />
                              ) : (
                                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '2px' }}>
                                  <input 
                                    type={cell.form_field.type === 'radio' ? 'radio' : 'checkbox'}
                                    checked={cell.form_field.checked}
                                    onChange={(e) => {
                                      e.stopPropagation()
                                      const newRows = [...element.rows]
                                      newRows[rowIdx].row[colIdx] = { 
                                        ...newRows[rowIdx].row[colIdx], 
                                        form_field: {
                                          ...cell.form_field,
                                          checked: e.target.checked
                                        }
                                      }
                                      onUpdate({ rows: newRows })
                                    }}
                                    onFocus={() => handleCellClick(rowIdx, colIdx)}
                                    onClick={(e) => {
                                      e.stopPropagation()
                                      handleCellClick(rowIdx, colIdx)
                                    }}
                                    style={{ cursor: 'pointer' }}
                                  />
                                  <span style={{ fontSize: '9px', color: 'hsl(var(--muted-foreground))' }}>{cell.form_field.name}</span>
                                </div>
                              )}
                            </div>
                          ) : cell.chequebox !== undefined ? (
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
                          {/* Cell width resize handle (except last column) */}
                          {colIdx < (element.maxcolumns - 1) && (
                            <div
                              onMouseDown={(e)=>handleCellWidthResizeStart(e, rowIdx, colIdx)}
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
                              onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(59, 130, 246, 0.5)'}
                              onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                              title="Drag to resize cell width"
                            />
                          )}
                          {/* Cell height resize handle (all cells) */}
                          <div
                            onMouseDown={(e)=>handleCellHeightResizeStart(e, rowIdx, colIdx)}
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
                            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(34, 197, 94, 0.5)'}
                            onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                            title="Drag to resize cell height"
                          />
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
            background: 'white',
            borderLeft: `${footerStyle.borderLeftWidth} solid ${footerStyle.borderColor}`,
            borderRight: `${footerStyle.borderRightWidth} solid ${footerStyle.borderColor}`,
            borderTop: `${footerStyle.borderTopWidth} solid ${footerStyle.borderColor}`,
            borderBottom: `${footerStyle.borderBottomWidth} solid ${footerStyle.borderColor}`
          }}>
            <input
              type="text"
              value={element.text || 'Page footer text'}
              onChange={(e) => onUpdate({ text: e.target.value })}
              style={{
                width: '100%',
                border: 'none',
                background: 'transparent',
                color: '#000',
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
            background: 'white',
            border: '2px dashed #bbb',
            borderRadius: '4px',
            opacity: 0.9,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '12px',
            color: '#666'
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
            border: '2px dashed #bbb',
            background: '#f5f5f5'
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
                <div style={{ marginTop: '8px', fontSize: '0.85rem', color: '#666' }}>
                  {element.imagename || 'Uploaded Image'}
                </div>
              </div>
            ) : (
              <div style={{ textAlign: 'center' }}>
                <ImageIcon size={32} style={{ color: '#999', marginBottom: '8px' }} />
                <div style={{ fontSize: '0.9rem', color: '#666' }}>
                  No image selected
                </div>
                <div style={{ fontSize: '0.8rem', color: '#888', marginTop: '4px' }}>
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
        margin: '4px 0',
        padding: isSelected && element.type !== 'table' ? '8px' : '0',
        border: isSelected && element.type !== 'table' ? '2px solid var(--secondary-color)' : '2px solid transparent',
        borderRadius: element.type === 'table' ? '0' : '6px',
        cursor: isDragging ? 'grabbing' : 'grab',
        background: isSelected && element.type !== 'table' ? '#e3f2fd' : 'transparent',
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
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
        <GripVertical size={14} style={{ color: '#888' }} />
        <span style={{ fontSize: '11px', fontWeight: '500', color: '#888', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
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
  const [fonts, setFonts] = useState(DEFAULT_FONTS)
  const [fontsLoading, setFontsLoading] = useState(true)
  const [copiedId, setCopiedId] = useState(null)
  const canvasRef = useRef(null)

  // Fetch fonts from API on component mount
  useEffect(() => {
    const fetchFonts = async () => {
      try {
        setFontsLoading(true)
        const response = await fetch('/api/v1/fonts')
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
    ? selectedElement.rows[selectedCell.rowIdx]?.row[selectedCell.colIdx] 
    : null
  const currentPageSize = PAGE_SIZES[config.page] || PAGE_SIZES.A4

  const generateId = () => Math.random().toString(36).substr(2, 9)

  const addElementAtPosition = (type, position) => {
    switch (type) {
      case 'title':
        if (!title) {
          setTitle({
            props: 'Helvetica:18:100:center:0:0:1:0',
            text: 'Document Title',
            table: {
              maxcolumns: 3,
              columnwidths: [1/3, 1/3, 1/3],
              rows: [
                {
                  row: [
                    { props: 'Helvetica:12:000:left:0:0:0:0', text: '' },
                    { props: 'Helvetica:18:100:center:0:0:0:0', text: 'Document Title' },
                    { props: 'Helvetica:12:000:left:0:0:0:0', text: '' }
                  ]
                }
              ]
            }
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
                { props: 'Helvetica:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'Helvetica:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'Helvetica:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) }
              ]
            },
            {
              row: [
                { props: 'Helvetica:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'Helvetica:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) },
                { props: 'Helvetica:12:000:left:1:1:1:1', text: '', width: (usableWidth * (1/3)) }
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
            props: 'Helvetica:10:001:center:0:0:0:0',
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
    if (draggedType !== 'checkbox' && draggedType !== 'image' && draggedType !== 'radio' && draggedType !== 'text_input') return

    const newRows = [...element.rows]
    
    if (draggedType === 'checkbox') {
      // Update the table cell to contain a checkbox form field
      newRows[rowIdx].row[colIdx] = {
        ...newRows[rowIdx].row[colIdx],
        form_field: {
          type: 'checkbox',
          name: `chk_${rowIdx}_${colIdx}`,
          checked: false,
          value: 'Yes'
        },
        chequebox: undefined, // Remove legacy checkbox
        text: undefined, // Remove any existing text
        image: undefined // Remove any existing image
      }
    } else if (draggedType === 'radio') {
      // Update the table cell to contain a radio form field
      newRows[rowIdx].row[colIdx] = {
        ...newRows[rowIdx].row[colIdx],
        form_field: {
          type: 'radio',
          name: `radio_group_${rowIdx}`,
          checked: false,
          value: `opt_${colIdx}`,
          shape: 'round'
        },
        chequebox: undefined,
        text: undefined,
        image: undefined
      }
    } else if (draggedType === 'text_input') {
      // Update the table cell to contain a text input form field
      newRows[rowIdx].row[colIdx] = {
        ...newRows[rowIdx].row[colIdx],
        form_field: {
          type: 'text',
          name: `text_${rowIdx}_${colIdx}`,
          value: '',
          checked: false
        },
        chequebox: undefined,
        text: undefined,
        image: undefined
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
        chequebox: undefined, // Remove any existing checkbox
        form_field: undefined
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
        
        // Ensure title has table structure
        if (data.title) {
          const titleWithTable = {
            ...data.title,
            table: data.title.table || {
              maxcolumns: 3,
              columnwidths: [1/3, 1/3, 1/3],
              rows: [{
                row: [
                  { props: 'Helvetica:12:000:left:0:0:0:0', text: '' },
                  { props: 'Helvetica:18:100:center:0:0:0:0', text: data.title.text || 'Document Title' },
                  { props: 'Helvetica:12:000:left:0:0:0:0', text: '' }
                ]
              }]
            }
          }
          setTitle(titleWithTable)
        } else {
          setTitle(null)
        }

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
      setCopiedId('toolbar')
      setTimeout(() => setCopiedId(null), 2000)
    } catch (error) {
      console.error('Copy failed:', error)
    }
  }

  const handleUploadFont = async (e) => {
    const file = e.target.files[0]
    if (!file) return

    const formData = new FormData()
    formData.append('font', file)

    try {
      setFontsLoading(true)
      const response = await fetch('/api/v1/fonts/upload', {
        method: 'POST',
        body: formData,
      })
      
      if (response.ok) {
        const data = await response.json()
        if (data.font) {
          setFonts(prev => {
             // Avoid duplicates
             if (prev.some(f => f.id === data.font.id)) return prev;
             return [...prev, data.font]
          })
          alert(`Font ${data.font.displayName} uploaded successfully!`)
        } else {
             // Refresh list if no font object returned but success
             const refresh = await fetch('/api/v1/fonts')
             if (refresh.ok) {
                 const refreshData = await refresh.json()
                 if(refreshData.fonts) setFonts(refreshData.fonts)
             }
             alert('Font uploaded successfully!')
        }
      } else {
        const errorData = await response.json()
        alert(`Failed to upload font: ${errorData.error || 'Unknown error'}`)
      }
    } catch (error) {
      console.error('Error uploading font:', error)
      alert('Error uploading font: ' + error.message)
    } finally {
      setFontsLoading(false)
      // Reset input
      e.target.value = null
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

    // Also include ordered elements array to preserve the order
    if (components.length > 0) {
      let tableIdx = 0
      let spacerIdx = 0
      let imageIdx = 0
      output.elements = components.map(comp => {
        if (comp.type === 'table') {
          return { type: 'table', index: tableIdx++ }
        } else if (comp.type === 'spacer') {
          return { type: 'spacer', index: spacerIdx++ }
        } else if (comp.type === 'image') {
          return { type: 'image', index: imageIdx++ }
        }
        return null
      }).filter(Boolean)
    }

    if (footer) output.footer = footer
    return output
  }

  const jsonOutput = useMemo(() => getJsonOutput(), [config, title, components, footer])

  // Local state for JSON editing
  const [jsonText, setJsonText] = useState('')
  const [isJsonEditing, setIsJsonEditing] = useState(false)

  // Update jsonText when jsonOutput changes (but not while editing)
  React.useEffect(() => {
    if (!isJsonEditing) {
      setJsonText(JSON.stringify(jsonOutput, null, 2))
    }
  }, [jsonOutput, isJsonEditing])

  const handleJsonChange = (e) => {
    setJsonText(e.target.value)
  }

  const handleJsonBlur = () => {
    setIsJsonEditing(false)
    try {
      const data = JSON.parse(jsonText)
      
      // Parse the JSON structure from the pasted content
      setConfig(data.config || { pageBorder: '1:1:1:1', page: 'A4', pageAlignment: 1, watermark: '' })
      
      // Ensure title has table structure
      if (data.title) {
        const titleWithTable = {
          ...data.title,
          table: data.title.table || {
            maxcolumns: 3,
            columnwidths: [1/3, 1/3, 1/3],
            rows: [{
              row: [
                { props: 'Helvetica:12:000:left:0:0:0:0', text: '' },
                { props: 'Helvetica:18:100:center:0:0:0:0', text: data.title.text || 'Document Title' },
                { props: 'Helvetica:12:000:left:0:0:0:0', text: '' }
              ]
            }]
          }
        }
        setTitle(titleWithTable)
      } else {
        setTitle(null)
      }

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
    } catch (err) {
      // Invalid JSON - reset to current state
      setJsonText(JSON.stringify(jsonOutput, null, 2))
    }
  }

  return (
    <>
      <style>
        {`
          .drop-target {
            background-color: hsl(var(--accent)) !important;
            border: 2px dashed hsl(var(--primary)) !important;
            opacity: 0.8;
          }
          .editor-sidebar {
            position: sticky;
            top: 144px;
            height: calc(100vh - 164px);
            overflow-y: auto;
            align-self: flex-start;
          }
          .editor-sidebar::-webkit-scrollbar {
            width: 6px;
          }
          .editor-sidebar::-webkit-scrollbar-track {
            background: transparent;
          }
          .editor-sidebar::-webkit-scrollbar-thumb {
            background: hsl(var(--border));
            border-radius: 3px;
          }
          .editor-sidebar::-webkit-scrollbar-thumb:hover {
            background: hsl(var(--muted-foreground));
          }
          .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 0.5rem 0;
            cursor: pointer;
            user-select: none;
          }
          .section-header:hover {
            opacity: 0.8;
          }
          .canvas-container {
            min-height: 500px;
            max-height: calc(100vh - 200px);
            overflow-y: auto;
          }
          .sticky-header {
            position: sticky;
            top: 0;
            z-index: 100;
            background: hsl(var(--background));
            border-bottom: 1px solid hsl(var(--border));
            padding: 0.75rem 1rem;
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
        `}
      </style>
      
      {/* Fixed Header Wrapper */}
      <div style={{ position: 'fixed', top: '58px', left: 0, right: 0, zIndex: 1000 }}>
      {/* Sticky Header */}
      <div className="sticky-header">
        <div className="container-full">
          {/* Compact Header */}
          <div style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'space-between',
            gap: '1rem',
            flexWrap: 'wrap'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <Edit size={24} style={{ color: 'var(--secondary-color)' }} />
              <div>
                <h1 style={{ fontSize: '1.25rem', margin: 0, color: 'hsl(var(--foreground))' }}>PDF Template Editor</h1>
                <p style={{ fontSize: '0.75rem', margin: 0, color: 'hsl(var(--muted-foreground))' }}>
                  {allElements.length} elements • {config.page} {config.pageAlignment === 1 ? 'Portrait' : 'Landscape'}
                </p>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
              <input
                id="template-file-input"
                type="text"
                placeholder="Load template file..."
                style={{ 
                  padding: '0.4rem 0.75rem', 
                  fontSize: '0.85rem', 
                  width: '200px',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  background: 'hsl(var(--background))',
                  color: 'hsl(var(--foreground))'
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && e.target.value.trim()) {
                    loadTemplate(e.target.value.trim())
                  }
                }}
              />
              <button
                onClick={() => {
                  const input = document.getElementById('template-file-input')
                  if (input && input.value.trim()) {
                    loadTemplate(input.value.trim())
                  }
                }}
                className="btn"
                style={{ padding: '0.4rem 0.75rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}
              >
                <Upload size={14} /> Load
              </button>
              <button
                onClick={previewPDF}
                className="btn btn-secondary"
                style={{ padding: '0.4rem 0.75rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}
              >
                <Eye size={14} /> Preview
              </button>
              <button
                onClick={downloadPDF}
                className="btn"
                style={{ padding: '0.4rem 0.75rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}
              >
                <Download size={14} /> Generate
              </button>

              <input
                type="file"
                id="font-upload-input"
                accept=".ttf,.otf"
                style={{ display: 'none' }}
                onChange={handleUploadFont}
                disabled={fontsLoading}
              />
              <button
                onClick={() => document.getElementById('font-upload-input').click()}
                className="btn"
                disabled={fontsLoading}
                style={{ padding: '0.4rem 0.75rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}
                title="Upload Custom Font (.ttf, .otf)"
              >
                <Type size={14} /> Upload Font
              </button>

              <button
                onClick={copyJSON}
                className="btn"
                style={{ padding: '0.4rem 0.75rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}
              >
                {copiedId === 'toolbar' ? <><Check size={14} /> Copied</> : <><Copy size={14} /> Copy</>}
              </button>
              
              <div style={{ 
                display: 'flex', 
                border: '1px solid hsl(var(--border))', 
                borderRadius: '6px',
                marginLeft: '0.5rem'
              }}>
                <button
                  onClick={() => setTheme('light')}
                  style={{
                    padding: '0.4rem 0.5rem',
                    background: theme === 'light' ? 'hsl(var(--accent))' : 'transparent',
                    color: 'hsl(var(--foreground))',
                    border: 'none',
                    borderRight: '1px solid hsl(var(--border))',
                    cursor: 'pointer',
                    borderRadius: '5px 0 0 5px'
                  }}
                >
                  <Sun size={14} />
                </button>
                <button
                  onClick={() => setTheme('dark')}
                  style={{
                    padding: '0.4rem 0.5rem',
                    background: theme === 'dark' ? 'hsl(var(--accent))' : 'transparent',
                    color: 'hsl(var(--foreground))',
                    border: 'none',
                    cursor: 'pointer',
                    borderRadius: '0 5px 5px 0'
                  }}
                >
                  <Moon size={14} />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      </div>

      {/* Main Content Wrapper */}
      <div style={{ marginTop: '138px', minHeight: 'calc(100vh - 138px)', display: 'flex', flexDirection: 'column' }}>

      {/* Main Content */}
      <div className="container-full" style={{ flex: 1, padding: '1rem' }}>
        <div className="grid editor-main-grid" style={{ gridTemplateColumns: '280px 1fr 350px', gap: '1rem', alignItems: 'start' }}>
          
          {/* Left: Components & Settings */}
          <div className="editor-sidebar">
            {/* Components Section */}
            <div className="card" style={{ padding: '1rem', marginBottom: '1rem' }}>
              <h3 style={{ 
                margin: '0 0 0.75rem 0', 
                fontSize: '0.9rem',
                fontWeight: '600',
                display: 'flex', 
                alignItems: 'center', 
                gap: '0.5rem',
                color: 'hsl(var(--foreground))'
              }}>
                <Square size={16} /> Components
              </h3>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
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
            </div>

            {/* Document Settings */}
            <div className="card" style={{ padding: '1rem' }}>
              <h3 style={{ 
                margin: '0 0 0.75rem 0', 
                fontSize: '0.9rem',
                fontWeight: '600',
                display: 'flex', 
                alignItems: 'center', 
                gap: '0.5rem',
                color: 'hsl(var(--foreground))'
              }}>
                <Settings size={16} /> Document Settings
              </h3>
              
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {/* Page Size & Orientation Row */}
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Page Size</label>
                    <select
                      value={config.page}
                      onChange={(e) => setConfig(prev => ({ ...prev, page: e.target.value }))}
                      style={{ 
                        width: '100%', 
                        padding: '0.4rem',
                        fontSize: '0.85rem',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '4px',
                        background: 'hsl(var(--background))',
                        color: 'hsl(var(--foreground))'
                      }}
                    >
                      {Object.entries(PAGE_SIZES).map(([key, size]) => (
                        <option key={key} value={key}>{size.name}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Orientation</label>
                    <select
                      value={config.pageAlignment}
                      onChange={(e) => setConfig(prev => ({ ...prev, pageAlignment: parseInt(e.target.value) }))}
                      style={{ 
                        width: '100%', 
                        padding: '0.4rem',
                        fontSize: '0.85rem',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '4px',
                        background: 'hsl(var(--background))',
                        color: 'hsl(var(--foreground))'
                      }}
                    >
                      <option value={1}>Portrait</option>
                      <option value={2}>Landscape</option>
                    </select>
                  </div>
                </div>

                {/* Watermark */}
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Watermark</label>
                  <input
                    type="text"
                    value={config.watermark || ''}
                    onChange={(e) => setConfig(prev => ({ ...prev, watermark: e.target.value }))}
                    placeholder="Optional watermark text"
                    style={{ 
                      width: '100%', 
                      padding: '0.4rem',
                      fontSize: '0.85rem',
                      border: '1px solid hsl(var(--border))',
                      borderRadius: '4px',
                      background: 'hsl(var(--background))',
                      color: 'hsl(var(--foreground))'
                    }}
                  />
                </div>

                {/* Arlington Compatible Toggle */}
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  justifyContent: 'space-between',
                  padding: '0.5rem',
                  background: 'hsl(var(--muted))',
                  borderRadius: '4px'
                }}>
                  <div>
                    <label style={{ 
                      display: 'block', 
                      fontSize: '0.8rem', 
                      fontWeight: '500',
                      color: 'hsl(var(--foreground))'
                    }}>
                      Arlington Compatible
                    </label>
                    <span style={{ 
                      fontSize: '0.7rem', 
                      color: 'hsl(var(--muted-foreground))'
                    }}>
                      PDF 2.0 compliant fonts
                    </span>
                  </div>
                  <label style={{ 
                    position: 'relative', 
                    display: 'inline-block', 
                    width: '40px', 
                    height: '22px' 
                  }}>
                    <input
                      type="checkbox"
                      checked={config.arlingtonCompatible || false}
                      onChange={(e) => setConfig(prev => ({ ...prev, arlingtonCompatible: e.target.checked }))}
                      style={{ 
                        opacity: 0, 
                        width: 0, 
                        height: 0 
                      }}
                    />
                    <span style={{
                      position: 'absolute',
                      cursor: 'pointer',
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      backgroundColor: config.arlingtonCompatible ? 'var(--secondary-color)' : 'hsl(var(--border))',
                      transition: '0.3s',
                      borderRadius: '22px'
                    }}>
                      <span style={{
                        position: 'absolute',
                        content: '""',
                        height: '16px',
                        width: '16px',
                        left: config.arlingtonCompatible ? '21px' : '3px',
                        bottom: '3px',
                        backgroundColor: 'white',
                        transition: '0.3s',
                        borderRadius: '50%'
                      }} />
                    </span>
                  </label>
                </div>

                {/* Embed Fonts Toggle */}
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  justifyContent: 'space-between',
                  padding: '0.5rem',
                  marginTop: '0.5rem',
                  background: 'hsl(var(--muted))',
                  borderRadius: '4px'
                }}>
                  <div>
                    <label style={{ 
                      display: 'block', 
                      fontSize: '0.8rem', 
                      fontWeight: '500',
                      color: 'hsl(var(--foreground))'
                    }}>
                      Embed Standard Fonts
                    </label>
                    <span style={{ 
                      fontSize: '0.7rem', 
                      color: 'hsl(var(--muted-foreground))'
                    }}>
                      Embed used standard fonts
                    </span>
                  </div>
                  <label style={{ 
                    position: 'relative', 
                    display: 'inline-block', 
                    width: '40px', 
                    height: '22px' 
                  }}>
                    <input
                      type="checkbox"
                      checked={config.embedFonts !== false}
                      onChange={(e) => setConfig(prev => ({ ...prev, embedFonts: e.target.checked }))}
                      style={{ 
                        opacity: 0, 
                        width: 0, 
                        height: 0 
                      }}
                    />
                    <span style={{
                      position: 'absolute',
                      cursor: 'pointer',
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      backgroundColor: (config.embedFonts !== false) ? 'var(--secondary-color)' : 'hsl(var(--border))',
                      transition: '0.3s',
                      borderRadius: '22px'
                    }}>
                      <span style={{
                        position: 'absolute',
                        content: '""',
                        height: '16px',
                        width: '16px',
                        left: (config.embedFonts !== false) ? '21px' : '3px',
                        bottom: '3px',
                        backgroundColor: 'white',
                        transition: '0.3s',
                        borderRadius: '50%'
                      }} />
                    </span>
                  </label>
                </div>

                {/* Page Borders */}
                <div style={{ paddingTop: '0.5rem', borderTop: '1px solid hsl(var(--border))' }}>
                  <PageBorderControls
                    borders={parsePageBorder(config.pageBorder)}
                    onChange={(borders) => setConfig(prev => ({ ...prev, pageBorder: formatPageBorder(borders) }))}
                  />
                </div>

                {/* Page Info */}
                <div style={{ 
                  padding: '0.5rem', 
                  background: 'hsl(var(--muted))', 
                  borderRadius: '4px',
                  fontSize: '0.75rem', 
                  color: 'hsl(var(--muted-foreground))'
                }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>Dimensions:</span>
                    <span>{currentPageSize.width} × {currentPageSize.height} pts</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '0.25rem' }}>
                    <span>Usable Width:</span>
                    <span>{getUsableWidth(currentPageSize.width)} pts</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Center: Canvas */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', alignItems: 'center' }}>
            {/* Page size indicator */}
            <div style={{ 
              fontSize: '0.75rem', 
              color: '#666', 
              textAlign: 'center',
              padding: '0.25rem 0.5rem',
              background: '#e9ecef',
              borderRadius: '4px'
            }}>
              {currentPageSize.name} - {currentPageSize.width} × {currentPageSize.height} pts
            </div>
            <div className="card" style={{ 
              padding: '1rem',
              width: `${Math.min(currentPageSize.width, 650)}px`,
              maxWidth: '100%',
              transition: 'width 0.3s ease'
            }}>
              <div 
                className="canvas-container"
                ref={canvasRef}
                onClick={handleCanvasClick}
                onDrop={handleDrop}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                style={{
                  background: isDragOver
                    ? 'repeating-linear-gradient(45deg, hsl(var(--accent)) 0px, hsl(var(--accent)) 2px, transparent 2px, transparent 20px)'
                    : '#fff',
                  border: isDragOver ? '3px dashed var(--secondary-color)' : '1px solid #ddd',
                  borderRadius: '4px',
                  padding: '1rem',
                  transition: 'all 0.2s ease',
                  minHeight: `${Math.min(currentPageSize.height * 0.6, 600)}px`,
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'flex-start',
                  boxShadow: '0 2px 8px rgba(0,0,0,0.08)'
                }}
              >
                {allElements.length === 0 && (
                  <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    minHeight: '300px',
                    textAlign: 'center',
                    color: '#999',
                    pointerEvents: 'none',
                  }}>
                    <Square size={48} style={{ opacity: 0.3, marginBottom: '1rem' }} />
                    <p style={{ fontSize: '1rem', fontWeight: '500', marginBottom: '0.5rem' }}>
                      {isDragOver ? 'Drop here to add component' : 'Drop components here'}
                    </p>
                    {!isDragOver && (
                      <p style={{ fontSize: '0.85rem', opacity: 0.7 }}>Drag from the left panel or load a template</p>
                    )}
                  </div>
                )}
                {/* Elements container - align items at top */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0', alignItems: 'stretch' }}>
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
                        currentPageSize={currentPageSize}
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
          <div className="editor-sidebar" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {/* Properties Panel */}
            <div className="card" style={{ padding: '1rem' }}>
              <h3 style={{ 
                margin: '0 0 0.75rem 0', 
                fontSize: '0.9rem',
                fontWeight: '600',
                display: 'flex', 
                alignItems: 'center', 
                gap: '0.5rem',
                color: 'hsl(var(--foreground))'
              }}>
                <Edit size={16} /> Properties
              </h3>
              
              {!selectedElement && (
                <div style={{ 
                  textAlign: 'center', 
                  padding: '2rem 1rem',
                  color: 'hsl(var(--muted-foreground))',
                  background: 'hsl(var(--muted))',
                  borderRadius: '6px'
                }}>
                  <Settings size={24} style={{ opacity: 0.3, marginBottom: '0.5rem' }} />
                  <p style={{ fontSize: '0.85rem', margin: 0 }}>Select a component to edit</p>
                </div>
              )}
              
              {selectedElement && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <div style={{ 
                    padding: '0.5rem 0.75rem', 
                    background: 'hsl(var(--accent))', 
                    borderRadius: '6px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between'
                  }}>
                    <strong style={{ fontSize: '0.9rem' }}>
                      {selectedElement.type.charAt(0).toUpperCase() + selectedElement.type.slice(1)}
                    </strong>
                    <button
                      onClick={() => deleteElement(selectedElement.id)}
                      style={{ 
                        background: 'hsl(var(--destructive))', 
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        padding: '0.25rem 0.5rem',
                        fontSize: '0.75rem',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.25rem'
                      }}
                    >
                      <Trash2 size={12} /> Delete
                    </button>
                  </div>

                  {selectedElement.type === 'title' && (
                    <>
                      {/* Title Background Color */}
                      <div style={{ marginBottom: '1rem', paddingBottom: '1rem', borderBottom: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Title Background Color</div>
                        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <input
                            type="color"
                            value={selectedElement.bgcolor || '#ffffff'}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { bgcolor: e.target.value })
                            }}
                            style={{ 
                              width: '40px', 
                              height: '40px', 
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              cursor: 'pointer',
                              padding: '2px'
                            }}
                          />
                          <input
                            type="text"
                            value={selectedElement.bgcolor || ''}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { bgcolor: e.target.value })
                            }}
                            placeholder="#RRGGBB or transparent"
                            style={{ 
                              flex: 1, 
                              padding: '0.4rem', 
                              fontSize: '0.85rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px'
                            }}
                          />
                          <button
                            onClick={() => {
                              updateElement(selectedElement.id, { bgcolor: '' })
                            }}
                            style={{
                              padding: '0.4rem 0.75rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              background: 'hsl(var(--muted))',
                              color: 'hsl(var(--muted-foreground))',
                              fontSize: '0.75rem',
                              cursor: 'pointer'
                            }}
                          >
                            Clear
                          </button>
                        </div>
                        <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                          {[
                            { label: 'White', color: '#FFFFFF' },
                            { label: 'Light Gray', color: '#F0F0F0' },
                            { label: 'Light Blue', color: '#E3F2FD' },
                            { label: 'Light Green', color: '#E8F5E9' },
                            { label: 'Light Yellow', color: '#FFFDE7' },
                            { label: 'Light Red', color: '#FFEBEE' }
                          ].map(({ label, color }) => (
                            <button
                              key={color}
                              onClick={() => updateElement(selectedElement.id, { bgcolor: color })}
                              title={label}
                              style={{
                                width: '24px',
                                height: '24px',
                                background: color,
                                border: selectedElement.bgcolor === color ? '2px solid var(--secondary-color)' : '1px solid hsl(var(--border))',
                                borderRadius: '4px',
                                cursor: 'pointer'
                              }}
                            />
                          ))}
                        </div>
                      </div>

                      {/* Title Text Color */}
                      <div style={{ marginBottom: '1rem', paddingBottom: '1rem', borderBottom: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Title Text Color</div>
                        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <input
                            type="color"
                            value={selectedElement.textcolor || '#000000'}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { textcolor: e.target.value })
                            }}
                            style={{ 
                              width: '40px', 
                              height: '40px', 
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              cursor: 'pointer',
                              padding: '2px'
                            }}
                          />
                          <input
                            type="text"
                            value={selectedElement.textcolor || ''}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { textcolor: e.target.value })
                            }}
                            placeholder="#RRGGBB (default: black)"
                            style={{ 
                              flex: 1, 
                              padding: '0.4rem', 
                              fontSize: '0.85rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px'
                            }}
                          />
                          <button
                            onClick={() => {
                              updateElement(selectedElement.id, { textcolor: '' })
                            }}
                            style={{
                              padding: '0.4rem 0.75rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              background: 'hsl(var(--muted))',
                              color: 'hsl(var(--muted-foreground))',
                              fontSize: '0.75rem',
                              cursor: 'pointer'
                            }}
                          >
                            Clear
                          </button>
                        </div>
                        <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                          {[
                            { label: 'Black', color: '#000000' },
                            { label: 'Dark Gray', color: '#333333' },
                            { label: 'White', color: '#FFFFFF' },
                            { label: 'Red', color: '#F44336' },
                            { label: 'Blue', color: '#2196F3' },
                            { label: 'Green', color: '#4CAF50' }
                          ].map(({ label, color }) => (
                            <button
                              key={color}
                              onClick={() => updateElement(selectedElement.id, { textcolor: color })}
                              title={label}
                              style={{
                                width: '24px',
                                height: '24px',
                                background: color,
                                border: selectedElement.textcolor === color ? '2px solid var(--secondary-color)' : '1px solid hsl(var(--border))',
                                borderRadius: '4px',
                                cursor: 'pointer'
                              }}
                            />
                          ))}
                        </div>
                      </div>
                      <div style={{ marginBottom: '1rem', paddingBottom: '1rem', borderBottom: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>
                          Title Table Settings
                        </div>
                        
                        {selectedElement.table && (
                          <>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                              <label style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))' }}>Columns:</label>
                              <input
                                type="number"
                                min="1"
                                max="10"
                                value={selectedElement.table.maxcolumns || 3}
                                onChange={(e) => {
                                  const newCols = parseInt(e.target.value)
                                  if (isNaN(newCols) || newCols < 1 || newCols > 10) return

                                  const table = selectedElement.table
                                  let newWidths = []
                                  const rawWidths = table.columnwidths || []
                                  const rawSum = rawWidths.reduce((a, b) => a + b, 0) || 1
                                  const currentWidths = rawWidths.map(w => w / rawSum)
                                  
                                  if (newCols === currentWidths.length) {
                                    newWidths = currentWidths
                                  } else if (newCols > currentWidths.length) {
                                    const remain = newCols - currentWidths.length
                                    const existingSum = currentWidths.reduce((a, b) => a + b, 0)
                                    const newPartWidth = (1 - existingSum) / remain
                                    newWidths = [...currentWidths, ...Array(remain).fill(Math.max(0.01, newPartWidth))]
                                  } else {
                                    newWidths = currentWidths.slice(0, newCols)
                                  }
                                  
                                  const sum = newWidths.reduce((a, b) => a + b, 0) || 1
                                  const normalizedWidths = newWidths.map(w => w / sum)

                                  const updatedRows = table.rows?.map(row => {
                                    const newRow = [...(row.row || [])]
                                    while (newRow.length < newCols) {
                                      newRow.push({ props: 'Helvetica:12:000:left:0:0:0:0', text: '' })
                                    }
                                    if (newRow.length > newCols) {
                                      newRow.splice(newCols)
                                    }
                                    return { row: newRow }
                                  })
                                  
                                  updateElement(selectedElement.id, { 
                                    table: { 
                                      ...table, 
                                      maxcolumns: newCols, 
                                      rows: updatedRows, 
                                      columnwidths: normalizedWidths 
                                    } 
                                  })
                                }}
                                style={{ flex: 1, padding: '0.25rem' }}
                              />
                            </div>
                            
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                              <button
                                onClick={() => {
                                  const table = selectedElement.table
                                  const newRow = {
                                    row: Array(table.maxcolumns).fill(null).map(() => ({
                                      props: 'Helvetica:12:000:left:0:0:0:0',
                                      text: ''
                                    }))
                                  }
                                  updateElement(selectedElement.id, {
                                    table: { ...table, rows: [...(table.rows || []), newRow] }
                                  })
                                }}
                                style={{
                                  flex: 1,
                                  padding: '0.4rem',
                                  fontSize: '0.75rem',
                                  border: '1px solid hsl(var(--border))',
                                  borderRadius: '4px',
                                  background: 'hsl(var(--muted))',
                                  cursor: 'pointer'
                                }}
                              >
                                Add Row
                              </button>
                              {selectedElement.table.rows?.length > 1 && (
                                <button
                                  onClick={() => {
                                    const table = selectedElement.table
                                    // Remove selected row if a cell is selected, otherwise remove last row
                                    const rowIdxToRemove = selectedCell ? selectedCell.rowIdx : table.rows.length - 1
                                    const newRows = table.rows.filter((_, idx) => idx !== rowIdxToRemove)
                                    updateElement(selectedElement.id, {
                                      table: { ...table, rows: newRows }
                                    })
                                    // Clear cell selection if the selected row was removed
                                    if (selectedCell && selectedCell.rowIdx === rowIdxToRemove) {
                                      setSelectedCell(null)
                                    } else if (selectedCell && selectedCell.rowIdx > rowIdxToRemove) {
                                      // Adjust selected cell row index if it was after the removed row
                                      setSelectedCell({ ...selectedCell, rowIdx: selectedCell.rowIdx - 1 })
                                    }
                                  }}
                                  style={{
                                    flex: 1,
                                    padding: '0.4rem',
                                    fontSize: '0.75rem',
                                    border: '1px solid hsl(var(--destructive))',
                                    borderRadius: '4px',
                                    background: 'hsl(var(--destructive))',
                                    color: 'white',
                                    cursor: 'pointer'
                                  }}
                                >
                                  Remove Row {selectedCell ? `(Row ${selectedCell.rowIdx + 1})` : '(Last)'}
                                </button>
                              )}
                            </div>
                            
                            <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem' }}>
                              <button
                                onClick={() => {
                                  const table = selectedElement.table
                                  const currentCols = table.maxcolumns || 3
                                  const newCols = currentCols + 1
                                  
                                  if (newCols > 10) {
                                    alert('Maximum 10 columns allowed')
                                    return
                                  }

                                  // Calculate new column widths
                                  const rawWidths = table.columnwidths || Array(currentCols).fill(1)
                                  const rawSum = rawWidths.reduce((a, b) => a + b, 0)
                                  const currentWidths = rawWidths.map(w => w / rawSum)
                                  
                                  const newColumnWeight = 1 / newCols
                                  const scaleFactor = (1 - newColumnWeight) / currentWidths.reduce((a, b) => a + b, 0)
                                  const newWidths = [...currentWidths.map(w => w * scaleFactor), newColumnWeight]
                                  
                                  const sum = newWidths.reduce((a, b) => a + b, 0)
                                  const normalizedWidths = newWidths.map(w => w / sum)

                                  // Add new cell to all rows
                                  const updatedRows = table.rows?.map(row => {
                                    const newRow = [...(row.row || [])]
                                    newRow.push({ props: 'Helvetica:12:000:left:0:0:0:0', text: '' })
                                    return { row: newRow }
                                  })

                                  updateElement(selectedElement.id, {
                                    table: { ...table, maxcolumns: newCols, rows: updatedRows, columnwidths: normalizedWidths }
                                  })
                                }}
                                style={{
                                  flex: 1,
                                  padding: '0.4rem',
                                  fontSize: '0.75rem',
                                  border: '1px solid hsl(var(--border))',
                                  borderRadius: '4px',
                                  background: 'hsl(var(--muted))',
                                  cursor: 'pointer'
                                }}
                                disabled={selectedElement.table.maxcolumns >= 10}
                              >
                                Add Column
                              </button>
                              {selectedElement.table.maxcolumns > 1 && (
                                <button
                                  onClick={() => {
                                    const table = selectedElement.table
                                    const currentCols = table.maxcolumns || 3
                                    
                                    if (currentCols <= 1) {
                                      alert('Cannot remove the last column')
                                      return
                                    }

                                    // Remove selected column if a cell is selected, otherwise remove last column
                                    const colIdxToRemove = selectedCell ? selectedCell.colIdx : currentCols - 1
                                    const newCols = currentCols - 1

                                    // Calculate new column widths
                                    const rawWidths = table.columnwidths || Array(currentCols).fill(1)
                                    const rawSum = rawWidths.reduce((a, b) => a + b, 0)
                                    const currentWidths = rawWidths.map(w => w / rawSum)
                                    
                                    const removedWeight = currentWidths[colIdxToRemove]
                                    const newWidths = currentWidths.filter((_, idx) => idx !== colIdxToRemove)
                                    
                                    const currentSum = newWidths.reduce((a, b) => a + b, 0)
                                    const normalizedWidths = newWidths.map(w => w / currentSum)

                                    // Remove cell from all rows at the selected column
                                    const updatedRows = table.rows?.map(row => {
                                      const newRow = row.row.filter((_, idx) => idx !== colIdxToRemove)
                                      return { row: newRow }
                                    })

                                    updateElement(selectedElement.id, {
                                      table: { ...table, maxcolumns: newCols, rows: updatedRows, columnwidths: normalizedWidths }
                                    })
                                    
                                    // Clear or adjust cell selection
                                    if (selectedCell && selectedCell.colIdx === colIdxToRemove) {
                                      setSelectedCell(null)
                                    } else if (selectedCell && selectedCell.colIdx > colIdxToRemove) {
                                      setSelectedCell({ ...selectedCell, colIdx: selectedCell.colIdx - 1 })
                                    }
                                  }}
                                  style={{
                                    flex: 1,
                                    padding: '0.4rem',
                                    fontSize: '0.75rem',
                                    border: '1px solid hsl(var(--destructive))',
                                    borderRadius: '4px',
                                    background: 'hsl(var(--destructive))',
                                    color: 'white',
                                    cursor: 'pointer'
                                  }}
                                >
                                  Remove Column {selectedCell ? `(Col ${selectedCell.colIdx + 1})` : '(Last)'}
                                </button>
                              )}
                            </div>
                          </>
                        )}
                      </div>
                      
                      {selectedCell && selectedElement.table && selectedElement.table.rows?.[selectedCell.rowIdx]?.row?.[selectedCell.colIdx] && (
                        <div style={{ marginBottom: '1rem', paddingBottom: '1rem', borderBottom: '1px solid hsl(var(--border))' }}>
                          <div style={{ fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>
                            Title Cell (Row {selectedCell.rowIdx + 1}, Col {selectedCell.colIdx + 1})
                          </div>
                          
                          {(() => {
                            const cell = selectedElement.table.rows[selectedCell.rowIdx].row[selectedCell.colIdx]
                            
                            // Helper function to update title table cell with proper immutable updates
                            const updateTitleCell = (cellUpdates) => {
                              const table = selectedElement.table
                              const newRows = table.rows.map((row, rIdx) => 
                                rIdx === selectedCell.rowIdx
                                  ? {
                                      ...row,
                                      row: row.row.map((c, cIdx) =>
                                        cIdx === selectedCell.colIdx
                                          ? { ...c, ...cellUpdates }
                                          : c
                                      )
                                    }
                                  : row
                              )
                              updateElement(selectedElement.id, { table: { ...table, rows: newRows } })
                            }
                            
                            return (
                              <>
                                <div style={{ marginBottom: '0.5rem' }}>
                                  <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Text:</label>
                                  <input
                                    type="text"
                                    value={cell.text || ''}
                                    onChange={(e) => updateTitleCell({ text: e.target.value })}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem' }}
                                  />
                                </div>
                                
                                {cell.image ? (
                                  <div style={{ marginBottom: '0.5rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Image:</label>
                                    <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '0.25rem' }}>
                                      <input
                                        type="number"
                                        value={cell.image.width || 100}
                                        onChange={(e) => updateTitleCell({ 
                                          image: { ...cell.image, width: parseFloat(e.target.value) || 100 } 
                                        })}
                                        placeholder="Width"
                                        style={{ flex: 1, padding: '0.25rem', fontSize: '0.75rem' }}
                                      />
                                      <input
                                        type="number"
                                        value={cell.image.height || 50}
                                        onChange={(e) => updateTitleCell({ 
                                          image: { ...cell.image, height: parseFloat(e.target.value) || 50 } 
                                        })}
                                        placeholder="Height"
                                        style={{ flex: 1, padding: '0.25rem', fontSize: '0.75rem' }}
                                      />
                                    </div>
                                    <button
                                      onClick={() => updateTitleCell({ image: null })}
                                      style={{
                                        padding: '0.25rem 0.5rem',
                                        fontSize: '0.75rem',
                                        border: '1px solid hsl(var(--destructive))',
                                        borderRadius: '4px',
                                        background: 'hsl(var(--destructive))',
                                        color: 'white',
                                        cursor: 'pointer'
                                      }}
                                    >
                                      Remove Image
                                    </button>
                                  </div>
                                ) : (
                                  <div style={{ marginBottom: '0.5rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Add Image:</label>
                                    <input
                                      type="file"
                                      accept="image/*"
                                      onChange={(e) => {
                                        const file = e.target.files[0]
                                        if (file) {
                                          const reader = new FileReader()
                                          reader.onload = (event) => {
                                            updateTitleCell({
                                              image: {
                                                imagename: file.name,
                                                imagedata: event.target.result,
                                                width: 100,
                                                height: 50
                                              }
                                            })
                                          }
                                          reader.readAsDataURL(file)
                                        }
                                      }}
                                      style={{ width: '100%', fontSize: '0.75rem' }}
                                    />
                                  </div>
                                )}
                                
                                <div>
                                  <PropsEditor 
                                    props={cell.props} 
                                    onChange={(props) => updateTitleCell({ props })}
                                    fonts={fonts}
                                  />
                                </div>
                              </>
                            )
                          })()}
                        </div>
                      )}
                    </>
                  )}

                  {selectedElement.type === 'table' && (
                    <>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <label style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))' }}>Columns:</label>
                        <input
                          type="number"
                          min="1"
                          max="10"
                          value={selectedElement.maxcolumns || 3}
                          onChange={(e) => {
                            const newCols = parseInt(e.target.value)
                            if (isNaN(newCols) || newCols < 1 || newCols > 10) return

                            // Adjust column widths proportionally - normalize first
                            let newWidths = []
                            const rawWidths = selectedElement.columnwidths || []
                            const rawSum = rawWidths.reduce((a, b) => a + b, 0) || 1
                            const currentWidths = rawWidths.map(w => w / rawSum)
                            
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
                                newRow.push({ props: 'Helvetica:12:000:left:1:1:1:1', text: '' })
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
                            // Normalize columnwidths for proper width calculation
                            const rawWeights = selectedElement.columnwidths && selectedElement.columnwidths.length === selectedElement.maxcolumns
                              ? selectedElement.columnwidths
                              : Array(selectedElement.maxcolumns).fill(1)
                            const totalWeight = rawWeights.reduce((sum, w) => sum + w, 0)
                            const normalizedWeights = rawWeights.map(w => w / totalWeight)
                            
                            const newRow = { 
                              row: Array(selectedElement.maxcolumns || 3).fill().map((_, i) => ({
                                props: 'Helvetica:12:000:left:1:1:1:1',
                                text: '',
                                width: usableWidth * normalizedWeights[i]
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
                              // Remove selected row if a cell is selected, otherwise remove last row
                              const rowIdxToRemove = selectedCell ? selectedCell.rowIdx : selectedElement.rows.length - 1
                              const newRows = selectedElement.rows.filter((_, idx) => idx !== rowIdxToRemove)
                              
                              // Also update rowheights if they exist
                              let newRowHeights = selectedElement.rowheights
                              if (newRowHeights && newRowHeights.length > 0) {
                                newRowHeights = newRowHeights.filter((_, idx) => idx !== rowIdxToRemove)
                              }
                              
                              updateElement(selectedElement.id, { 
                                rows: newRows,
                                ...(newRowHeights ? { rowheights: newRowHeights } : {})
                              })
                              
                              // Clear or adjust cell selection
                              if (selectedCell && selectedCell.rowIdx === rowIdxToRemove) {
                                setSelectedCell(null)
                              } else if (selectedCell && selectedCell.rowIdx > rowIdxToRemove) {
                                setSelectedCell({ ...selectedCell, rowIdx: selectedCell.rowIdx - 1 })
                              }
                            }
                          }}
                          className="btn"
                          style={{ padding: '0.3rem 0.6rem', fontSize: '0.8rem', flex: 1 }}
                          disabled={!selectedElement.rows || selectedElement.rows.length <= 1}
                        >
                          Remove Row {selectedCell ? `(Row ${selectedCell.rowIdx + 1})` : '(Last)'}
                        </button>
                      </div>

                      <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem' }}>
                        <button
                          onClick={() => {
                            try {
                              const currentCols = selectedElement.maxcolumns || 3
                              const newCols = currentCols + 1
                              
                              if (newCols > 10) {
                                alert('Maximum 10 columns allowed')
                                return
                              }

                              // Calculate new column widths - normalize current weights first
                              const rawWidths = selectedElement.columnwidths || Array(currentCols).fill(1)
                              const rawSum = rawWidths.reduce((a, b) => a + b, 0)
                              const currentWidths = rawWidths.map(w => w / rawSum)
                              
                              const newColumnWeight = 1 / newCols
                              const scaleFactor = (1 - newColumnWeight) / currentWidths.reduce((a, b) => a + b, 0)
                              const newWidths = [...currentWidths.map(w => w * scaleFactor), newColumnWeight]
                              
                              // Normalize to ensure sum is exactly 1
                              const sum = newWidths.reduce((a, b) => a + b, 0)
                              const normalizedWidths = newWidths.map(w => w / sum)

                              // Get usable width for calculations
                              const usableWidth = getUsableWidth(currentPageSize.width)

                              // Add new cell to all rows with default borders (1:1:1:1)
                              const updatedRows = selectedElement.rows?.map(row => {
                                const newRow = [...(row.row || [])]
                                // Add new cell with all borders enabled
                                newRow.push({ 
                                  props: 'Helvetica:12:000:left:1:1:1:1',
                                  text: '',
                                  width: usableWidth * normalizedWidths[newCols - 1]
                                })
                                
                                // Update widths for all cells including existing ones
                                const updatedCells = newRow.map((cell, colIdx) => ({
                                  ...cell,
                                  width: usableWidth * normalizedWidths[colIdx]
                                }))

                                return { row: updatedCells }
                              })

                              // Validate that all rows have the correct number of columns
                              const allRowsValid = updatedRows.every(row => row.row.length === newCols)
                              
                              if (!allRowsValid) {
                                alert('Failed to update JSON: Row validation failed')
                                return
                              }

                              // Update the element
                              updateElement(selectedElement.id, { 
                                maxcolumns: newCols, 
                                rows: updatedRows, 
                                columnwidths: normalizedWidths 
                              })

                              // Success - component will re-render automatically
                              console.log('Column added successfully')
                            } catch (error) {
                              console.error('Failed to add column:', error)
                              alert('Failed to update JSON: ' + error.message)
                            }
                          }}
                          className="btn"
                          style={{ padding: '0.3rem 0.6rem', fontSize: '0.8rem', flex: 1 }}
                          disabled={selectedElement.maxcolumns >= 10}
                        >
                          <Plus size={12} /> Add Column
                        </button>
                        <button
                          onClick={() => {
                            try {
                              const currentCols = selectedElement.maxcolumns || 3
                              
                              if (currentCols <= 1) {
                                alert('Cannot remove the last column')
                                return
                              }

                              // Remove selected column if a cell is selected, otherwise remove last column
                              const colIdxToRemove = selectedCell ? selectedCell.colIdx : currentCols - 1
                              const newCols = currentCols - 1

                              // Calculate new column widths (redistribute the removed column's weight)
                              // First normalize current weights
                              const rawWidths = selectedElement.columnwidths || Array(currentCols).fill(1)
                              const rawSum = rawWidths.reduce((a, b) => a + b, 0)
                              const currentWidths = rawWidths.map(w => w / rawSum)
                              
                              const removedWeight = currentWidths[colIdxToRemove]
                              const newWidths = currentWidths.filter((_, idx) => idx !== colIdxToRemove)
                              
                              // Distribute removed weight proportionally to remaining columns
                              const currentSum = newWidths.reduce((a, b) => a + b, 0)
                              const normalizedWidths = newWidths.map(w => w / currentSum)

                              // Get usable width for calculations
                              const usableWidth = getUsableWidth(currentPageSize.width)

                              // Remove cell at selected column from all rows
                              const updatedRows = selectedElement.rows?.map(row => {
                                const newRow = row.row.filter((_, idx) => idx !== colIdxToRemove)
                                
                                // Update widths for remaining cells
                                const updatedCells = newRow.map((cell, colIdx) => ({
                                  ...cell,
                                  width: usableWidth * normalizedWidths[colIdx]
                                }))

                                return { row: updatedCells }
                              })

                              // Validate that all rows have the correct number of columns
                              const allRowsValid = updatedRows.every(row => row.row.length === newCols)
                              
                              if (!allRowsValid) {
                                alert('Failed to update JSON: Row validation failed')
                                return
                              }

                              // Update the element
                              updateElement(selectedElement.id, { 
                                maxcolumns: newCols, 
                                rows: updatedRows, 
                                columnwidths: normalizedWidths 
                              })
                              
                              // Clear or adjust cell selection
                              if (selectedCell && selectedCell.colIdx === colIdxToRemove) {
                                setSelectedCell(null)
                              } else if (selectedCell && selectedCell.colIdx > colIdxToRemove) {
                                setSelectedCell({ ...selectedCell, colIdx: selectedCell.colIdx - 1 })
                              }

                              // Success - component will re-render automatically
                              console.log('Column removed successfully')
                            } catch (error) {
                              console.error('Failed to remove column:', error)
                              alert('Failed to update JSON: ' + error.message)
                            }
                          }}
                          className="btn"
                          style={{ padding: '0.3rem 0.6rem', fontSize: '0.8rem', flex: 1 }}
                          disabled={!selectedElement.rows || selectedElement.maxcolumns <= 1}
                        >
                          Remove Column {selectedCell ? `(Col ${selectedCell.colIdx + 1})` : '(Last)'}
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
                            // Normalize columnwidths for display
                            const rawWeights = selectedElement.columnwidths && selectedElement.columnwidths.length === selectedElement.maxcolumns
                              ? selectedElement.columnwidths
                              : Array(selectedElement.maxcolumns).fill(1)
                            const totalWeight = rawWeights.reduce((sum, w) => sum + w, 0)
                            const colWeights = rawWeights.map(w => w / totalWeight)
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
                                    const sum = newWeights.reduce((a,b)=>a+b,0) || 1
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

                      {/* Table Background Color */}
                      <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Table Background Color</div>
                        <p style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                          Sets the default background color for all cells. Individual cells can override this.
                        </p>
                        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <input
                            type="color"
                            value={selectedElement.bgcolor || '#ffffff'}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { bgcolor: e.target.value })
                            }}
                            style={{ 
                              width: '40px', 
                              height: '40px', 
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              cursor: 'pointer',
                              padding: '2px'
                            }}
                          />
                          <input
                            type="text"
                            value={selectedElement.bgcolor || ''}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { bgcolor: e.target.value })
                            }}
                            placeholder="#RRGGBB or transparent"
                            style={{ 
                              flex: 1, 
                              padding: '0.4rem', 
                              fontSize: '0.85rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px'
                            }}
                          />
                          <button
                            onClick={() => {
                              updateElement(selectedElement.id, { bgcolor: '' })
                            }}
                            style={{
                              padding: '0.4rem 0.75rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              background: 'hsl(var(--muted))',
                              color: 'hsl(var(--muted-foreground))',
                              fontSize: '0.75rem',
                              cursor: 'pointer'
                            }}
                          >
                            Clear
                          </button>
                        </div>
                        <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                          {[
                            { label: 'White', color: '#FFFFFF' },
                            { label: 'Light Gray', color: '#F0F0F0' },
                            { label: 'Light Blue', color: '#E3F2FD' },
                            { label: 'Light Green', color: '#E8F5E9' },
                            { label: 'Light Yellow', color: '#FFFDE7' },
                            { label: 'Light Red', color: '#FFEBEE' }
                          ].map(({ label, color }) => (
                            <button
                              key={color}
                              onClick={() => updateElement(selectedElement.id, { bgcolor: color })}
                              title={label}
                              style={{
                                width: '24px',
                                height: '24px',
                                background: color,
                                border: selectedElement.bgcolor === color ? '2px solid var(--secondary-color)' : '1px solid hsl(var(--border))',
                                borderRadius: '4px',
                                cursor: 'pointer'
                              }}
                            />
                          ))}
                        </div>
                      </div>

                      {/* Table Text Color */}
                      <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                        <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Table Text Color</div>
                        <p style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                          Sets the default text color for all cells. Individual cells can override this.
                        </p>
                        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <input
                            type="color"
                            value={selectedElement.textcolor || '#000000'}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { textcolor: e.target.value })
                            }}
                            style={{ 
                              width: '40px', 
                              height: '40px', 
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              cursor: 'pointer',
                              padding: '2px'
                            }}
                          />
                          <input
                            type="text"
                            value={selectedElement.textcolor || ''}
                            onChange={(e) => {
                              updateElement(selectedElement.id, { textcolor: e.target.value })
                            }}
                            placeholder="#RRGGBB (default: black)"
                            style={{ 
                              flex: 1, 
                              padding: '0.4rem', 
                              fontSize: '0.85rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px'
                            }}
                          />
                          <button
                            onClick={() => {
                              updateElement(selectedElement.id, { textcolor: '' })
                            }}
                            style={{
                              padding: '0.4rem 0.75rem',
                              border: '1px solid hsl(var(--border))',
                              borderRadius: '4px',
                              background: 'hsl(var(--muted))',
                              color: 'hsl(var(--muted-foreground))',
                              fontSize: '0.75rem',
                              cursor: 'pointer'
                            }}
                          >
                            Clear
                          </button>
                        </div>
                        <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                          {[
                            { label: 'Black', color: '#000000' },
                            { label: 'Dark Gray', color: '#333333' },
                            { label: 'White', color: '#FFFFFF' },
                            { label: 'Red', color: '#F44336' },
                            { label: 'Blue', color: '#2196F3' },
                            { label: 'Green', color: '#4CAF50' }
                          ].map(({ label, color }) => (
                            <button
                              key={color}
                              onClick={() => updateElement(selectedElement.id, { textcolor: color })}
                              title={label}
                              style={{
                                width: '24px',
                                height: '24px',
                                background: color,
                                border: selectedElement.textcolor === color ? '2px solid var(--secondary-color)' : '1px solid hsl(var(--border))',
                                borderRadius: '4px',
                                cursor: 'pointer'
                              }}
                            />
                          ))}
                        </div>
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
                            ) : selectedCellElement.form_field ? (
                              <div style={{ marginBottom: '0.5rem' }}>
                                <div style={{ padding: '0.5rem', background: 'hsl(var(--muted))', borderRadius: '4px', marginBottom: '0.5rem' }}>
                                  <strong>{selectedCellElement.form_field.type === 'radio' ? 'Radio Button' : selectedCellElement.form_field.type === 'text' ? 'Text Input' : 'Checkbox'} Field</strong>
                                </div>
                                
                                <div style={{ marginBottom: '0.5rem' }}>
                                  <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Field Name:</label>
                                  <input
                                    type="text"
                                    value={selectedCellElement.form_field.name}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                        ...selectedCellElement, 
                                        form_field: {
                                          ...selectedCellElement.form_field,
                                          name: e.target.value
                                        }
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                  />
                                </div>

                                {selectedCellElement.form_field.type === 'radio' && (
                                  <div style={{ marginBottom: '0.5rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>Shape:</label>
                                    <select
                                      value={selectedCellElement.form_field.shape || 'round'}
                                      onChange={(e) => {
                                        const newRows = [...selectedElement.rows]
                                        newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                          ...selectedCellElement, 
                                          form_field: {
                                            ...selectedCellElement.form_field,
                                            shape: e.target.value
                                          }
                                        }
                                        updateElement(selectedElement.id, { rows: newRows })
                                      }}
                                      style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                    >
                                      <option value="round">Round</option>
                                      <option value="square">Square</option>
                                    </select>
                                  </div>
                                )}

                                <div style={{ marginBottom: '0.5rem' }}>
                                  <label style={{ display: 'block', fontSize: '0.9rem', marginBottom: '0.25rem' }}>{selectedCellElement.form_field.type === 'text' ? 'Default Value:' : 'Export Value:'}</label>
                                  <input
                                    type="text"
                                    value={selectedCellElement.form_field.value}
                                    onChange={(e) => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                        ...selectedCellElement, 
                                        form_field: {
                                          ...selectedCellElement.form_field,
                                          value: e.target.value
                                        }
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.9rem' }}
                                  />
                                </div>

                                {selectedCellElement.form_field.type !== 'text' && (
                                  <div style={{ marginBottom: '0.5rem' }}>
                                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.9rem' }}>
                                      <input
                                        type="checkbox"
                                        checked={selectedCellElement.form_field.checked}
                                        onChange={(e) => {
                                          const newRows = [...selectedElement.rows]
                                          newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { 
                                            ...selectedCellElement, 
                                            form_field: {
                                              ...selectedCellElement.form_field,
                                              checked: e.target.checked
                                            }
                                          }
                                          updateElement(selectedElement.id, { rows: newRows })
                                        }}
                                      />
                                      Default Checked
                                    </label>
                                  </div>
                                )}
                              </div>
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
                                    fonts={fonts}
                                  />
                                </div>
                              </>
                            )}

                            {/* Cell Size Controls - applies to all cell types */}
                            <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                              <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.25rem', color: 'hsl(var(--foreground))' }}>Cell Size Override</div>
                              <div style={{ fontSize: '0.75rem', marginBottom: '0.5rem', color: 'hsl(var(--muted-foreground))' }}>
                                💡 Drag the blue handle (right edge) to resize width, or green handle (bottom edge) to resize height
                              </div>
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
                                      const value = e.target.value === '' ? undefined : parseFloat(e.target.value)
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        width: value
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem' }}
                                  />
                                  {selectedCellElement.width && (
                                    <div style={{ fontSize: '0.7rem', marginTop: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>
                                      Custom: {selectedCellElement.width.toFixed(2)}pts
                                    </div>
                                  )}
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
                                      const value = e.target.value === '' ? undefined : parseFloat(e.target.value)
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        height: value
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem' }}
                                  />
                                  {selectedCellElement.height && (
                                    <div style={{ fontSize: '0.7rem', marginTop: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>
                                      Custom: {selectedCellElement.height.toFixed(2)}pts
                                    </div>
                                  )}
                                </div>
                              </div>
                              
                              {/* Width Redistribution Info */}
                              <div style={{ 
                                marginTop: '0.75rem', 
                                padding: '0.5rem', 
                                borderRadius: '4px', 
                                background: 'hsl(var(--muted))',
                                fontSize: '0.7rem',
                                color: 'hsl(var(--muted-foreground))',
                                lineHeight: '1.4'
                              }}>
                                <strong>Width Adjustment Rules:</strong><br/>
                                • Column {selectedCell.colIdx + 1} of {selectedElement.maxcolumns}<br/>
                                {selectedCell.colIdx === 0 ? (
                                  "• First column: width distributed to all other columns"
                                ) : selectedCell.colIdx === selectedElement.maxcolumns - 1 ? (
                                  "• Last column: resize disabled (cannot adjust)"
                                ) : (
                                  `• Middle column: width only affects next column`
                                )}
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

                            {/* Cell Background Color */}
                            <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                              <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.25rem', color: 'hsl(var(--foreground))' }}>Cell Background Color</div>
                              <p style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                {selectedElement.bgcolor ? `Inherits from table: ${selectedElement.bgcolor}` : 'No table background set.'}
                              </p>
                              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                <input
                                  type="color"
                                  value={selectedCellElement.bgcolor || selectedElement.bgcolor || '#ffffff'}
                                  onChange={(e) => {
                                    const newRows = [...selectedElement.rows]
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                      ...selectedCellElement,
                                      bgcolor: e.target.value
                                    }
                                    updateElement(selectedElement.id, { rows: newRows })
                                  }}
                                  style={{ 
                                    width: '40px', 
                                    height: '40px', 
                                    border: '1px solid hsl(var(--border))',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    padding: '2px'
                                  }}
                                />
                                <input
                                  type="text"
                                  value={selectedCellElement.bgcolor || ''}
                                  onChange={(e) => {
                                    const newRows = [...selectedElement.rows]
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                      ...selectedCellElement,
                                      bgcolor: e.target.value
                                    }
                                    updateElement(selectedElement.id, { rows: newRows })
                                  }}
                                  placeholder={selectedElement.bgcolor || '#RRGGBB'}
                                  style={{ 
                                    flex: 1, 
                                    padding: '0.4rem', 
                                    fontSize: '0.85rem',
                                    border: '1px solid hsl(var(--border))',
                                    borderRadius: '4px'
                                  }}
                                />
                                <button
                                  onClick={() => {
                                    const newRows = [...selectedElement.rows]
                                    const { bgcolor, ...rest } = selectedCellElement
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = rest
                                    updateElement(selectedElement.id, { rows: newRows })
                                  }}
                                  style={{
                                    padding: '0.4rem 0.75rem',
                                    border: '1px solid hsl(var(--border))',
                                    borderRadius: '4px',
                                    background: 'hsl(var(--muted))',
                                    color: 'hsl(var(--muted-foreground))',
                                    fontSize: '0.75rem',
                                    cursor: 'pointer'
                                  }}
                                >
                                  Clear
                                </button>
                              </div>
                              <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                {[
                                  { label: 'White', color: '#FFFFFF' },
                                  { label: 'Light Gray', color: '#F0F0F0' },
                                  { label: 'Light Blue', color: '#E3F2FD' },
                                  { label: 'Light Green', color: '#E8F5E9' },
                                  { label: 'Light Yellow', color: '#FFFDE7' },
                                  { label: 'Light Red', color: '#FFEBEE' },
                                  { label: 'Blue', color: '#2196F3' },
                                  { label: 'Green', color: '#4CAF50' },
                                  { label: 'Orange', color: '#FF9800' },
                                  { label: 'Red', color: '#F44336' }
                                ].map(({ label, color }) => (
                                  <button
                                    key={color}
                                    onClick={() => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        bgcolor: color
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    title={label}
                                    style={{
                                      width: '24px',
                                      height: '24px',
                                      background: color,
                                      border: selectedCellElement.bgcolor === color ? '2px solid var(--secondary-color)' : '1px solid hsl(var(--border))',
                                      borderRadius: '4px',
                                      cursor: 'pointer'
                                    }}
                                  />
                                ))}
                              </div>
                              {selectedCellElement.bgcolor && (
                                <div style={{ 
                                  marginTop: '0.5rem',
                                  padding: '0.25rem 0.5rem',
                                  background: selectedCellElement.bgcolor,
                                  border: '1px solid hsl(var(--border))',
                                  borderRadius: '4px',
                                  fontSize: '0.75rem',
                                  color: parseInt(selectedCellElement.bgcolor.replace('#', ''), 16) > 0xffffff/2 ? '#000' : '#fff'
                                }}>
                                  Preview: {selectedCellElement.bgcolor}
                                </div>
                              )}
                            </div>

                            {/* Cell Text Color */}
                            <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid hsl(var(--border))' }}>
                              <div style={{ fontSize: '0.9rem', fontWeight: '600', marginBottom: '0.25rem', color: 'hsl(var(--foreground))' }}>Cell Text Color</div>
                              <p style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                {selectedElement.textcolor ? `Inherits from table: ${selectedElement.textcolor}` : 'No table text color set (default: black).'}
                              </p>
                              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                <input
                                  type="color"
                                  value={selectedCellElement.textcolor || selectedElement.textcolor || '#000000'}
                                  onChange={(e) => {
                                    const newRows = [...selectedElement.rows]
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                      ...selectedCellElement,
                                      textcolor: e.target.value
                                    }
                                    updateElement(selectedElement.id, { rows: newRows })
                                  }}
                                  style={{ 
                                    width: '40px', 
                                    height: '40px', 
                                    border: '1px solid hsl(var(--border))',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    padding: '2px'
                                  }}
                                />
                                <input
                                  type="text"
                                  value={selectedCellElement.textcolor || ''}
                                  onChange={(e) => {
                                    const newRows = [...selectedElement.rows]
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                      ...selectedCellElement,
                                      textcolor: e.target.value
                                    }
                                    updateElement(selectedElement.id, { rows: newRows })
                                  }}
                                  placeholder={selectedElement.textcolor || '#RRGGBB'}
                                  style={{ 
                                    flex: 1, 
                                    padding: '0.4rem', 
                                    fontSize: '0.85rem',
                                    border: '1px solid hsl(var(--border))',
                                    borderRadius: '4px'
                                  }}
                                />
                                <button
                                  onClick={() => {
                                    const newRows = [...selectedElement.rows]
                                    const { textcolor, ...rest } = selectedCellElement
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = rest
                                    updateElement(selectedElement.id, { rows: newRows })
                                  }}
                                  style={{
                                    padding: '0.4rem 0.75rem',
                                    border: '1px solid hsl(var(--border))',
                                    borderRadius: '4px',
                                    background: 'hsl(var(--muted))',
                                    color: 'hsl(var(--muted-foreground))',
                                    fontSize: '0.75rem',
                                    cursor: 'pointer'
                                  }}
                                >
                                  Clear
                                </button>
                              </div>
                              <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                {[
                                  { label: 'Black', color: '#000000' },
                                  { label: 'Dark Gray', color: '#333333' },
                                  { label: 'Gray', color: '#666666' },
                                  { label: 'White', color: '#FFFFFF' },
                                  { label: 'Red', color: '#F44336' },
                                  { label: 'Dark Red', color: '#B71C1C' },
                                  { label: 'Blue', color: '#2196F3' },
                                  { label: 'Dark Blue', color: '#1565C0' },
                                  { label: 'Green', color: '#4CAF50' },
                                  { label: 'Dark Green', color: '#2E7D32' }
                                ].map(({ label, color }) => (
                                  <button
                                    key={color}
                                    onClick={() => {
                                      const newRows = [...selectedElement.rows]
                                      newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                        ...selectedCellElement,
                                        textcolor: color
                                      }
                                      updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    title={label}
                                    style={{
                                      width: '24px',
                                      height: '24px',
                                      background: color,
                                      border: selectedCellElement.textcolor === color ? '2px solid var(--secondary-color)' : '1px solid hsl(var(--border))',
                                      borderRadius: '4px',
                                      cursor: 'pointer'
                                    }}
                                  />
                                ))}
                              </div>
                              {selectedCellElement.textcolor && (
                                <div style={{ 
                                  marginTop: '0.5rem',
                                  padding: '0.25rem 0.5rem',
                                  background: 'hsl(var(--muted))',
                                  border: '1px solid hsl(var(--border))',
                                  borderRadius: '4px',
                                  fontSize: '0.75rem',
                                  color: selectedCellElement.textcolor
                                }}>
                                  Sample Text: {selectedCellElement.textcolor}
                                </div>
                              )}
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
                          fonts={fonts}
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
              <div style={{ marginBottom: '0.75rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <h3 style={{ margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.9rem' }}>
                  <FileText size={16} /> JSON Template
                </h3>
                <button
                  onClick={async () => {
                    try {
                      await navigator.clipboard.writeText(jsonText)
                      setCopiedId('json')
                      setTimeout(() => setCopiedId(null), 2000)
                    } catch (error) {
                      console.error('Copy failed:', error)
                    }
                  }}
                  className="btn"
                  style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}
                >
                  {copiedId === 'json' ? <><Check size={12} /> Copied</> : <><Copy size={12} /> Copy</>}
                </button>
              </div>
              <textarea
                value={jsonText}
                onChange={handleJsonChange}
                onFocus={() => setIsJsonEditing(true)}
                onBlur={handleJsonBlur}
                style={{
                  width: '100%',
                  height: '250px',
                  fontFamily: 'ui-monospace, "SF Mono", "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace',
                  fontSize: '0.7rem',
                  padding: '0.75rem',
                  resize: 'vertical',
                  background: '#1e1e1e',
                  color: '#d4d4d4',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '4px',
                  lineHeight: '1.4'
                }}
                spellCheck={false}
              />
              <p style={{ 
                marginTop: '0.5rem', 
                fontSize: '0.7rem', 
                color: 'hsl(var(--muted-foreground))'
              }}>
                Edit JSON directly or paste to load template. Changes apply on blur.
              </p>
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
    </>
  )
}
