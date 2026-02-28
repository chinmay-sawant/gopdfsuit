import { useState, useRef, useEffect } from 'react'
import ReactDOM from 'react-dom'
import './ContextMenu.css'

// ─── Sub-components ─────────────────────────────────────────────

function MenuSeparator() {
  return <div className="context-menu-separator" />
}

function MenuItem({ icon, label, shortcut, disabled, active, onClick, onHide }) {
  const handleClick = (e) => {
    e.stopPropagation()
    if (disabled) return
    onClick?.()
    onHide?.()
  }
  return (
    <div
      className={`context-menu-item${disabled ? ' disabled' : ''}${active ? ' active' : ''}`}
      onClick={handleClick}
    >
      <span className="context-menu-icon">
        {active ? <span className="context-menu-check">✓</span> : (icon || '')}
      </span>
      <span className="context-menu-label">{label}</span>
      {shortcut && <span className="context-menu-shortcut">{shortcut}</span>}
    </div>
  )
}

function SubMenu({ icon, label, disabled, children }) {
  const [open, setOpen] = useState(false)
  const timerRef = useRef(null)
  const containerRef = useRef(null)
  const submenuRef = useRef(null)
  const [flip, setFlip] = useState({ left: false, up: false })

  const handleEnter = () => {
    clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => setOpen(true), 120)
  }

  const handleLeave = () => {
    clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => setOpen(false), 200)
  }

  const handleClick = (e) => {
    e.stopPropagation()
    if (!open) {
      setOpen(true)
    }
  }

  // Measure submenu after open and flip if overflowing
  useEffect(() => {
    if (open && submenuRef.current) {
      requestAnimationFrame(() => {
        if (!submenuRef.current) return
        const rect = submenuRef.current.getBoundingClientRect()
        setFlip({
          left: rect.right > window.innerWidth - 8,
          up: rect.bottom > window.innerHeight - 8
        })
      })
    } else {
      setFlip({ left: false, up: false })
    }
  }, [open])

  useEffect(() => {
    return () => clearTimeout(timerRef.current)
  }, [])

  if (disabled) {
    return (
      <div className="context-menu-item disabled">
        <span className="context-menu-icon">{icon || ''}</span>
        <span className="context-menu-label">{label}</span>
        <span className="context-menu-arrow">›</span>
      </div>
    )
  }

  const submenuClass = [
    'context-menu-submenu-items',
    flip.left ? 'flip-left' : '',
    flip.up ? 'flip-up' : ''
  ].filter(Boolean).join(' ')

  return (
    <div
      className="context-menu-submenu"
      ref={containerRef}
      onMouseEnter={handleEnter}
      onMouseLeave={handleLeave}
    >
      <div className="context-menu-item" onClick={handleClick}>
        <span className="context-menu-icon">{icon || ''}</span>
        <span className="context-menu-label">{label}</span>
        <span className="context-menu-arrow">›</span>
      </div>
      {open && (
        <div ref={submenuRef} className={submenuClass}>
          {children}
        </div>
      )}
    </div>
  )
}

// ─── Menu Builder ───────────────────────────────────────────────

const isMac = typeof navigator !== 'undefined' && /Mac/.test(navigator.platform)
const mod = isMac ? '⌘' : 'Ctrl+'

function getMenuItems(targetType, targetData, handlers, clipboard, hasTitle, onHide) {
  const items = []
  const elementId = targetData?.elementId
  const element = targetData?.element
  const index = targetData?.index
  const rowIdx = targetData?.rowIdx
  const colIdx = targetData?.colIdx

  switch (targetType) {
    case 'canvas':
      items.push(
        <MenuItem key="paste" label="Paste" shortcut={`${mod}V`} disabled={!clipboard} onClick={() => handlers.paste()} onHide={onHide} />,
        <MenuSeparator key="s1" />,
        <SubMenu key="insert" label="Insert" icon="➕">
          <MenuItem label="Table" onClick={() => handlers.addElement('table')} onHide={onHide} />
          <MenuItem label="Spacer" onClick={() => handlers.addElement('spacer')} onHide={onHide} />
          <MenuItem label="Image" onClick={() => handlers.addElement('image')} onHide={onHide} />
          <MenuSeparator />
          <MenuItem label="Title" disabled={hasTitle} onClick={() => handlers.addElement('title')} onHide={onHide} />
          <MenuItem label="Footer" onClick={() => handlers.addElement('footer')} onHide={onHide} />
        </SubMenu>
      )
      break

    case 'title':
    case 'footer':
      items.push(
        <MenuItem key="cut" icon="✂" label="Cut" shortcut={`${mod}X`} onClick={() => handlers.cut(elementId)} onHide={onHide} />,
        <MenuItem key="copy" icon="📋" label="Copy" shortcut={`${mod}C`} onClick={() => handlers.copy(elementId)} onHide={onHide} />,
        <MenuItem key="paste" label="Paste" shortcut={`${mod}V`} disabled={!clipboard} onClick={() => handlers.paste(elementId)} onHide={onHide} />,
        <MenuItem key="dup" label="Duplicate" shortcut={`${mod}D`} onClick={() => handlers.duplicate(elementId)} onHide={onHide} />,
        <MenuSeparator key="s1" />,
        <MenuItem key="del" icon="🗑" label="Delete" shortcut="Del" onClick={() => handlers.delete(elementId)} onHide={onHide} />,
        <MenuSeparator key="s2" />,
        <MenuItem key="bold" label="Bold" shortcut={`${mod}B`} onClick={() => handlers.toggleStyle(elementId, 0)} onHide={onHide} />,
        <MenuItem key="italic" label="Italic" shortcut={`${mod}I`} onClick={() => handlers.toggleStyle(elementId, 1)} onHide={onHide} />,
        <MenuItem key="underline" label="Underline" shortcut={`${mod}U`} onClick={() => handlers.toggleStyle(elementId, 2)} onHide={onHide} />,
        <MenuSeparator key="s3" />,
        <SubMenu key="align" label="Alignment" icon="☰">
          <MenuItem label="Left" onClick={() => handlers.setAlignment(elementId, 'left')} onHide={onHide} />
          <MenuItem label="Center" onClick={() => handlers.setAlignment(elementId, 'center')} onHide={onHide} />
          <MenuItem label="Right" onClick={() => handlers.setAlignment(elementId, 'right')} onHide={onHide} />
        </SubMenu>,
        <SubMenu key="borders" label="Borders" icon="▢">
          <MenuItem label="None" onClick={() => handlers.setBorderPreset(elementId, 'none')} onHide={onHide} />
          <MenuItem label="All" onClick={() => handlers.setBorderPreset(elementId, 'all')} onHide={onHide} />
          <MenuItem label="Box" onClick={() => handlers.setBorderPreset(elementId, 'box')} onHide={onHide} />
          <MenuItem label="Bottom Only" onClick={() => handlers.setBorderPreset(elementId, 'bottom')} onHide={onHide} />
        </SubMenu>
      )
      break

    case 'table':
      items.push(
        <MenuItem key="cut" icon="✂" label="Cut" shortcut={`${mod}X`} onClick={() => handlers.cut(elementId)} onHide={onHide} />,
        <MenuItem key="copy" icon="📋" label="Copy" shortcut={`${mod}C`} onClick={() => handlers.copy(elementId)} onHide={onHide} />,
        <MenuItem key="paste" label="Paste" shortcut={`${mod}V`} disabled={!clipboard} onClick={() => handlers.paste(elementId)} onHide={onHide} />,
        <MenuItem key="dup" label="Duplicate" shortcut={`${mod}D`} onClick={() => handlers.duplicate(elementId)} onHide={onHide} />,
        <MenuSeparator key="s1" />,
        <MenuItem key="up" label="Move Up" shortcut="Alt+↑" disabled={index <= 0} onClick={() => handlers.moveUp(index)} onHide={onHide} />,
        <MenuItem key="down" label="Move Down" shortcut="Alt+↓" onClick={() => handlers.moveDown(index)} onHide={onHide} />,
        <MenuSeparator key="s2" />,
        <MenuItem key="addrow" label="Add Row" onClick={() => handlers.addRow(elementId)} onHide={onHide} />,
        <MenuItem key="addcol" label="Add Column" onClick={() => handlers.addColumn(elementId)} onHide={onHide} />,
        <MenuItem key="rmrow" label="Remove Last Row" disabled={element?.rows?.length <= 1} onClick={() => handlers.removeRow(elementId)} onHide={onHide} />,
        <MenuItem key="rmcol" label="Remove Last Column" disabled={(element?.rows?.[0]?.row?.length || 0) <= 1} onClick={() => handlers.removeColumn(elementId)} onHide={onHide} />,
        <MenuSeparator key="s3" />,
        <MenuItem key="del" icon="🗑" label="Delete Table" shortcut="Del" onClick={() => handlers.delete(elementId)} onHide={onHide} />
      )
      break

    case 'cell':
      items.push(
        <MenuItem key="cut" icon="✂" label="Cut" shortcut={`${mod}X`} onClick={() => handlers.cut(elementId)} onHide={onHide} />,
        <MenuItem key="copy" icon="📋" label="Copy" shortcut={`${mod}C`} onClick={() => handlers.copy(elementId)} onHide={onHide} />,
        <MenuItem key="paste" label="Paste" shortcut={`${mod}V`} disabled={!clipboard} onClick={() => handlers.paste(elementId)} onHide={onHide} />,
        <MenuSeparator key="s1" />,
        <MenuItem key="bold" label="Bold" shortcut={`${mod}B`} onClick={() => handlers.toggleCellStyle(elementId, rowIdx, colIdx, 0)} onHide={onHide} />,
        <MenuItem key="italic" label="Italic" shortcut={`${mod}I`} onClick={() => handlers.toggleCellStyle(elementId, rowIdx, colIdx, 1)} onHide={onHide} />,
        <MenuItem key="underline" label="Underline" shortcut={`${mod}U`} onClick={() => handlers.toggleCellStyle(elementId, rowIdx, colIdx, 2)} onHide={onHide} />,
        <MenuSeparator key="s2" />,
        <SubMenu key="align" label="Alignment" icon="☰">
          <MenuItem label="Left" onClick={() => handlers.setCellAlignment(elementId, rowIdx, colIdx, 'left')} onHide={onHide} />
          <MenuItem label="Center" onClick={() => handlers.setCellAlignment(elementId, rowIdx, colIdx, 'center')} onHide={onHide} />
          <MenuItem label="Right" onClick={() => handlers.setCellAlignment(elementId, rowIdx, colIdx, 'right')} onHide={onHide} />
        </SubMenu>,
        <MenuSeparator key="s3" />,
        <SubMenu key="insertcell" label="Insert Into Cell" icon="📥">
          <MenuItem label="Checkbox" onClick={() => handlers.insertField(elementId, rowIdx, colIdx, 'checkbox')} onHide={onHide} />
          <MenuItem label="Radio Button" onClick={() => handlers.insertField(elementId, rowIdx, colIdx, 'radio')} onHide={onHide} />
          <MenuItem label="Text Input" onClick={() => handlers.insertField(elementId, rowIdx, colIdx, 'text_input')} onHide={onHide} />
          <MenuItem label="Image" onClick={() => handlers.insertField(elementId, rowIdx, colIdx, 'image')} onHide={onHide} />
          <MenuItem label="Hyperlink" onClick={() => handlers.insertField(elementId, rowIdx, colIdx, 'hyperlink')} onHide={onHide} />
        </SubMenu>,
        <MenuItem key="wrap" label="Toggle Text Wrap" onClick={() => handlers.toggleWrap(elementId, rowIdx, colIdx)} onHide={onHide} />,
        <MenuSeparator key="s4" />,
        <SubMenu key="borders" label="Cell Borders" icon="▢">
          <MenuItem label="None" onClick={() => handlers.setCellBorderPreset(elementId, rowIdx, colIdx, 'none')} onHide={onHide} />
          <MenuItem label="All" onClick={() => handlers.setCellBorderPreset(elementId, rowIdx, colIdx, 'all')} onHide={onHide} />
          <MenuItem label="Box" onClick={() => handlers.setCellBorderPreset(elementId, rowIdx, colIdx, 'box')} onHide={onHide} />
          <MenuItem label="Bottom Only" onClick={() => handlers.setCellBorderPreset(elementId, rowIdx, colIdx, 'bottom')} onHide={onHide} />
        </SubMenu>,
        <MenuSeparator key="s5" />,
        <MenuItem key="addrow" label="Add Row" onClick={() => handlers.addRow(elementId)} onHide={onHide} />,
        <MenuItem key="addcol" label="Add Column" onClick={() => handlers.addColumn(elementId)} onHide={onHide} />,
        <MenuSeparator key="s6" />,
        <SubMenu key="delete" label="Delete" icon="🗑">
          <MenuItem label="Clear Cell" onClick={() => handlers.clearCell(elementId, rowIdx, colIdx)} onHide={onHide} />
          <MenuSeparator />
          <MenuItem label={`Delete Row ${rowIdx + 1}`} disabled={element?.rows?.length <= 1} onClick={() => handlers.deleteRow(elementId, rowIdx)} onHide={onHide} />
          <MenuItem label={`Delete Column ${colIdx + 1}`} disabled={(element?.rows?.[0]?.row?.length || 0) <= 1} onClick={() => handlers.deleteColumn(elementId, colIdx)} onHide={onHide} />
          <MenuSeparator />
          <MenuItem label="Delete Table" shortcut="Del" onClick={() => handlers.delete(elementId)} onHide={onHide} />
        </SubMenu>
      )
      break

    case 'spacer':
    case 'image':
      items.push(
        <MenuItem key="cut" icon="✂" label="Cut" shortcut={`${mod}X`} onClick={() => handlers.cut(elementId)} onHide={onHide} />,
        <MenuItem key="copy" icon="📋" label="Copy" shortcut={`${mod}C`} onClick={() => handlers.copy(elementId)} onHide={onHide} />,
        <MenuItem key="dup" label="Duplicate" shortcut={`${mod}D`} onClick={() => handlers.duplicate(elementId)} onHide={onHide} />,
        <MenuSeparator key="s1" />,
        <MenuItem key="up" label="Move Up" shortcut="Alt+↑" disabled={index <= 0} onClick={() => handlers.moveUp(index)} onHide={onHide} />,
        <MenuItem key="down" label="Move Down" shortcut="Alt+↓" onClick={() => handlers.moveDown(index)} onHide={onHide} />,
        <MenuSeparator key="s2" />,
        <MenuItem key="del" icon="🗑" label="Delete" shortcut="Del" onClick={() => handlers.delete(elementId)} onHide={onHide} />
      )
      break

    default:
      break
  }

  return items
}

// ─── Main Component ─────────────────────────────────────────────

export default function ContextMenu({ menuState, onHide, handlers, clipboard, hasTitle }) {
  const menuRef = useRef(null)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [ready, setReady] = useState(false)

  // WhatsApp-style: render offscreen, measure, then place in view
  useEffect(() => {
    if (!menuState.visible) {
      setReady(false)
      return
    }

    // Initially place offscreen so we can measure
    setPosition({ x: -9999, y: -9999 })
    setReady(false)

    // After first paint, measure and compute final position
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        if (!menuRef.current) return
        const rect = menuRef.current.getBoundingClientRect()
        const vw = window.innerWidth
        const vh = window.innerHeight
        const pad = 8
        let x = menuState.x
        let y = menuState.y

        // If menu overflows right, flip to left of cursor
        if (x + rect.width > vw - pad) {
          x = Math.max(pad, x - rect.width)
        }
        // If menu overflows bottom, place above cursor
        if (y + rect.height > vh - pad) {
          y = Math.max(pad, y - rect.height)
        }
        // Final clamp so it never goes offscreen
        x = Math.max(pad, Math.min(x, vw - rect.width - pad))
        y = Math.max(pad, Math.min(y, vh - rect.height - pad))

        setPosition({ x, y })
        setReady(true)
      })
    })
  }, [menuState])

  if (!menuState.visible) return null

  const items = getMenuItems(
    menuState.targetType,
    menuState.targetData,
    handlers,
    clipboard,
    hasTitle,
    onHide
  )

  return ReactDOM.createPortal(
    <>
      <div className="context-menu-overlay" onClick={onHide} onContextMenu={(e) => { e.preventDefault(); onHide() }} />
      <div
        ref={menuRef}
        className="context-menu"
        style={{
          left: position.x,
          top: position.y,
          opacity: ready ? 1 : 0,
          pointerEvents: ready ? 'auto' : 'none'
        }}
        onClick={(e) => e.stopPropagation()}
        onContextMenu={(e) => e.preventDefault()}
      >
        {items}
      </div>
    </>,
    document.body
  )
}
