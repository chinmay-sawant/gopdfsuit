
import React, { useState } from 'react'
import { Edit, Settings, Trash2, ArrowLeft, ArrowRight, ArrowDown, ArrowUp } from 'lucide-react'
import { formatProps, parseProps } from './utils'
import { DEFAULT_FONTS } from './constants'

function PropsEditor({ props, onChange, fonts = DEFAULT_FONTS, showAlignment = true, showBorders = true }) {
    const parsed = parseProps(props)

    const updateBorder = (index, value) => {
        const newBorders = [...parsed.borders]
        newBorders[index] = Math.max(0, Math.min(10, value))
        onChange(formatProps({ ...parsed, borders: newBorders }))
    }

    const BorderControls = ({ label, index }) => (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <label style={{ fontSize: '0.75rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>{label}</label>
            <div style={{ display: 'flex', gap: '0.25rem', alignItems: 'center' }}>
                <button className="btn-border" onClick={() => updateBorder(index, parsed.borders[index] - 1)} disabled={parsed.borders[index] <= 0}>−</button>
                <span style={{ padding: '0.25rem 0.4rem', fontSize: '0.75rem', minWidth: '2.5rem', textAlign: 'center', background: 'hsl(var(--muted))', borderRadius: '4px' }}>{parsed.borders[index]}px</span>
                <button className="btn-border" onClick={() => updateBorder(index, parsed.borders[index] + 1)} disabled={parsed.borders[index] >= 10}>+</button>
            </div>
        </div>
    )

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {/* Font Section */}
            <div>
                <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Font</label>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                    <div>
                        <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Family</label>
                        <select value={parsed.font} onChange={(e) => onChange(formatProps({ ...parsed, font: e.target.value }))} style={{ width: '100%', padding: '0.4rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', fontSize: '0.85rem' }}>
                            {fonts.map(font => <option key={font.id} value={font.id}>{font.displayName}</option>)}
                        </select>
                    </div>
                    <div>
                        <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Size</label>
                        <select value={parsed.size} onChange={(e) => onChange(formatProps({ ...parsed, size: parseInt(e.target.value) }))} style={{ width: '100%', padding: '0.4rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', fontSize: '0.85rem' }}>
                            {[8, 9, 10, 11, 12, 14, 16, 18, 20, 24, 28, 32, 36, 48, 72].map(size => <option key={size} value={size}>{size}px</option>)}
                        </select>
                    </div>
                </div>
            </div>

            {/* Style Section */}
            <div>
                <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Style</label>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                    {[{ key: 0, label: 'B' }, { key: 1, label: 'I' }, { key: 2, label: 'U' }].map(({ key, label }) => (
                        <button key={key} onClick={() => { const s = parsed.style.split(''); s[key] = s[key] === '1' ? '0' : '1'; onChange(formatProps({ ...parsed, style: s.join('') })) }} style={{ padding: '0.4rem 0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: parsed.style[key] === '1' ? 'hsl(var(--accent))' : 'hsl(var(--background))', color: parsed.style[key] === '1' ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))', fontSize: '0.85rem', fontWeight: parsed.style[key] === '1' ? '600' : '400', cursor: 'pointer' }}>{label}</button>
                    ))}
                </div>
            </div>

            {/* Alignment Section */}
            {showAlignment && (
                <div>
                    <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Alignment</label>
                    <div style={{ display: 'flex', gap: '0.25rem' }}>
                        {[
                            { value: 'left', icon: <ArrowLeft size={14} />, label: 'Left' },
                            { value: 'center', icon: <ArrowDown size={14} style={{ transform: 'rotate(0deg)' }} />, label: 'Center' },
                            { value: 'right', icon: <ArrowRight size={14} />, label: 'Right' }
                        ].map(({ value, icon, label }) => (
                            <button key={value} onClick={() => onChange(formatProps({ ...parsed, align: value }))} style={{ flex: 1, padding: '0.4rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: parsed.align === value ? 'hsl(var(--accent))' : 'hsl(var(--background))', color: parsed.align === value ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))', fontSize: '0.75rem', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.25rem' }}>
                                {icon} {label}
                            </button>
                        ))}
                    </div>
                </div>
            )}

            {/* Borders Section */}
            {showBorders && (
                <div>
                    <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Borders</label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem', marginBottom: '0.5rem' }}>
                        <BorderControls label="Left" index={0} />
                        <BorderControls label="Right" index={1} />
                        <BorderControls label="Top" index={2} />
                        <BorderControls label="Bottom" index={3} />
                    </div>
                    <div>
                        <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Quick Set</label>
                        <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
                            {[
                                { label: 'None', borders: [0, 0, 0, 0] },
                                { label: 'All', borders: [1, 1, 1, 1] },
                                { label: 'Box', borders: [1, 1, 1, 1] },
                                { label: 'Bottom', borders: [0, 0, 0, 1] }
                            ].map(({ label, borders: presetBorders }) => (
                                <button
                                    key={label}
                                    onClick={() => onChange(formatProps({ ...parsed, borders: presetBorders }))}
                                    style={{
                                        padding: '0.25rem 0.5rem',
                                        border: '1px solid hsl(var(--border))',
                                        borderRadius: '4px',
                                        background: 'hsl(var(--muted))',
                                        color: 'hsl(var(--muted-foreground))',
                                        fontSize: '0.75rem',
                                        cursor: 'pointer'
                                    }}
                                >
                                    {label}
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

export default function PropertiesPanel({ selectedElement, selectedCell, selectedCellElement, updateElement, deleteElement, setSelectedCell, fonts }) {
    const [showColorPicker, setShowColorPicker] = useState(null)

    if (!selectedElement) {
        return (
            <div className="card" style={{ padding: '2rem 1rem', textAlign: 'center', color: 'hsl(var(--muted-foreground))', background: 'hsl(var(--muted))' }}>
                <Settings size={24} style={{ opacity: 0.3, marginBottom: '0.5rem' }} />
                <p style={{ fontSize: '0.85rem', margin: 0 }}>Select a component to edit</p>
            </div>
        )
    }

    const handleDelete = () => deleteElement(selectedElement.id)

    // Color preset swatches - using light pastel colors for table backgrounds
    const tableBackgroundPresets = [
        { label: 'White', color: '#FFFFFF' },
        { label: 'Light Gray', color: '#F0F0F0' },
        { label: 'Light Blue', color: '#E3F2FD' },
        { label: 'Light Green', color: '#E8F5E9' },
        { label: 'Light Yellow', color: '#FFFDE7' },
        { label: 'Light Red', color: '#FFEBEE' }
    ]
    const cellBackgroundPresets = [
        { label: 'White', color: '#FFFFFF' },
        { label: 'Light Gray', color: '#F0F0F0' },
        { label: 'Light Blue', color: '#E3F2FD' },
        { label: 'Light Green', color: '#E8F5E9' },
        { label: 'Light Yellow', color: '#FFFDE7' },
        { label: 'Light Red', color: '#FFEBEE' }
    ]
    const cellTextPresets = ['#1E1E1E', '#2D2D2D', '#424242', '#FFFFFF', '#FF0000', '#0000FF', '#00FF00']



    return (
        <div className="card" style={{ padding: '1rem', flexShrink: 0 }}>
            {/* Header */}
            <div style={{ marginBottom: '1rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                    <h3 style={{ margin: 0, fontSize: '0.85rem', fontWeight: '600', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'hsl(var(--foreground))' }}>
                        <Edit size={14} /> Properties
                    </h3>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingBottom: '0.75rem', borderBottom: '1px solid hsl(var(--border))' }}>
                    {/* Component Type Badge */}
                    <div style={{
                        display: 'inline-block',
                        padding: '0.3rem 0.65rem',
                        background: 'hsl(var(--accent))',
                        color: 'hsl(var(--accent-foreground))',
                        borderRadius: '4px',
                        fontSize: '0.8rem',
                        fontWeight: '600',
                        textTransform: 'capitalize'
                    }}>
                        {selectedElement.type}
                    </div>
                    <button
                        onClick={handleDelete}
                        style={{
                            padding: '0.35rem 0.65rem',
                            background: 'hsl(var(--destructive))',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontSize: '0.75rem',
                            fontWeight: '500',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.25rem'
                        }}
                    >
                        <Trash2 size={12} /> Delete
                    </button>
                </div>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>

                {/* TITLE Properties */}
                {selectedElement.type === 'title' && (
                    <>
                        {/* Title Background Color */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Title Background Color</label>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                <input
                                    type="color"
                                    value={selectedElement.bgcolor || '#ffffff'}
                                    onChange={(e) => updateElement(selectedElement.id, { bgcolor: e.target.value })}
                                    style={{ width: '48px', height: '32px', border: '1px solid hsl(var(--border))', borderRadius: '4px', cursor: 'pointer', padding: '2px', WebkitAppearance: 'none', background: 'transparent' }}
                                />
                                <input
                                    type="text"
                                    value={selectedElement.bgcolor || '#ffffff'}
                                    onChange={(e) => updateElement(selectedElement.id, { bgcolor: e.target.value })}
                                    placeholder="#RRGGBB or transparent"
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                                <button
                                    onClick={() => updateElement(selectedElement.id, { bgcolor: '' })}
                                    style={{ padding: '0.4rem 0.6rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                                >
                                    Clear
                                </button>
                            </div>
                            <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                {tableBackgroundPresets.map(({ label, color }) => (
                                    <button
                                        key={color}
                                        onClick={() => updateElement(selectedElement.id, { bgcolor: color })}
                                        style={{ width: '24px', height: '24px', border: '1px solid #999', borderRadius: '4px', background: color, cursor: 'pointer', boxShadow: 'inset 0 0 0 1px rgba(0,0,0,0.1)' }}
                                        title={label}
                                    />
                                ))}
                            </div>
                        </div>

                        {/* Title Text Color */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Title Text Color</label>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                <input
                                    type="color"
                                    value={selectedElement.textcolor || '#000000'}
                                    onChange={(e) => updateElement(selectedElement.id, { textcolor: e.target.value })}
                                    style={{ width: '48px', height: '32px', border: '1px solid hsl(var(--border))', borderRadius: '4px', cursor: 'pointer', padding: '2px', WebkitAppearance: 'none', background: 'transparent' }}
                                />
                                <input
                                    type="text"
                                    value={selectedElement.textcolor || '#000000'}
                                    onChange={(e) => updateElement(selectedElement.id, { textcolor: e.target.value })}
                                    placeholder="#RRGGBB (default: black)"
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                                <button
                                    onClick={() => updateElement(selectedElement.id, { textcolor: '' })}
                                    style={{ padding: '0.4rem 0.6rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                                >
                                    Clear
                                </button>
                            </div>
                            <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                {cellTextPresets.map(color => (
                                    <button
                                        key={color}
                                        onClick={() => updateElement(selectedElement.id, { textcolor: color })}
                                        style={{ width: '24px', height: '24px', border: '2px solid hsl(var(--border))', borderRadius: '4px', background: color, cursor: 'pointer' }}
                                        title={color}
                                    />
                                ))}
                            </div>
                        </div>

                        {/* Title Table Settings */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Title Table Settings</label>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', marginBottom: '0.5rem' }}>
                                <label style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>Columns:</label>
                                <input
                                    type="number"
                                    min="1"
                                    max="5"
                                    value={selectedElement.table?.maxcolumns || 3}
                                    onChange={(e) => {
                                        const newCols = parseInt(e.target.value) || 3
                                        const currentTable = selectedElement.table || { maxcolumns: 3, columnwidths: [1, 2, 1], rows: [{ row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' }, { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }] }] }
                                        const oldCols = currentTable.maxcolumns
                                        if (newCols !== oldCols) {
                                            const newRows = currentTable.rows.map(row => {
                                                if (newCols > oldCols) {
                                                    // Add columns
                                                    const fillerCells = Array(newCols - oldCols).fill(null).map(() => ({ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }))
                                                    return { row: [...row.row, ...fillerCells] }
                                                } else {
                                                    // Remove columns
                                                    return { row: row.row.slice(0, newCols) }
                                                }
                                            })
                                            const newWidths = Array(newCols).fill(1)
                                            updateElement(selectedElement.id, { table: { ...currentTable, maxcolumns: newCols, columnwidths: newWidths, rows: newRows } })
                                        }
                                    }}
                                    style={{ width: '60px', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                            </div>
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                <button
                                    onClick={() => {
                                        const currentTable = selectedElement.table || { maxcolumns: 3, columnwidths: [1, 2, 1], rows: [{ row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' }, { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }] }] }
                                        const newRow = { row: Array(currentTable.maxcolumns).fill(null).map(() => ({ props: 'Helvetica:12:000:left:1:1:1:1', text: '' })) }
                                        updateElement(selectedElement.id, { table: { ...currentTable, rows: [...currentTable.rows, newRow] } })
                                    }}
                                    style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    Add Row
                                </button>
                                <button
                                    onClick={() => {
                                        const currentTable = selectedElement.table || { maxcolumns: 3, columnwidths: [1, 2, 1], rows: [{ row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' }, { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }] }] }
                                        const newCols = currentTable.maxcolumns + 1
                                        if (newCols > 5) return
                                        const newRows = currentTable.rows.map(row => ({ row: [...row.row, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }] }))
                                        const newWidths = [...(currentTable.columnwidths || []), 1]
                                        updateElement(selectedElement.id, { table: { ...currentTable, maxcolumns: newCols, columnwidths: newWidths, rows: newRows } })
                                    }}
                                    style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    Add Column
                                </button>
                                {selectedCell && (
                                    <button
                                        onClick={() => {
                                            const currentTable = selectedElement.table || { maxcolumns: 3, columnwidths: [1, 2, 1], rows: [{ row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' }, { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }] }] }
                                            if (currentTable.maxcolumns <= 1) return
                                            const colToRemove = selectedCell.colIdx
                                            const newRows = currentTable.rows.map(row => ({ row: row.row.filter((_, idx) => idx !== colToRemove) }))
                                            const newWidths = (currentTable.columnwidths || []).filter((_, idx) => idx !== colToRemove)
                                            updateElement(selectedElement.id, { table: { ...currentTable, maxcolumns: currentTable.maxcolumns - 1, columnwidths: newWidths, rows: newRows } })
                                            setSelectedCell(null)
                                        }}
                                        style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--destructive))', borderRadius: '4px', background: 'hsl(var(--destructive))', color: 'white', cursor: 'pointer', gridColumn: 'span 2' }}
                                    >
                                        Remove Column (Col {selectedCell.colIdx + 1})
                                    </button>
                                )}
                            </div>
                        </div>

                        {/* Title Table Borders Toggle */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Title Table Borders</label>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <button
                                    onClick={() => {
                                        // Toggle all borders to 1:1:1:1
                                        const currentTable = selectedElement.table || { maxcolumns: 3, columnwidths: [1, 2, 1], rows: [{ row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' }, { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }] }] }
                                        const updatedRows = currentTable.rows.map(row => ({
                                            ...row,
                                            row: row.row.map(cell => {
                                                const parsed = parseProps(cell.props)
                                                return { ...cell, props: formatProps({ ...parsed, borders: [1, 1, 1, 1] }) }
                                            })
                                        }))
                                        updateElement(selectedElement.id, { table: { ...currentTable, rows: updatedRows } })
                                    }}
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    All Borders On
                                </button>
                                <button
                                    onClick={() => {
                                        // Toggle all borders to 0:0:0:0
                                        const currentTable = selectedElement.table || { maxcolumns: 3, columnwidths: [1, 2, 1], rows: [{ row: [{ props: 'Helvetica:12:000:left:1:1:1:1', text: '' }, { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' }, { props: 'Helvetica:12:000:right:1:1:1:1', text: '' }] }] }
                                        const updatedRows = currentTable.rows.map(row => ({
                                            ...row,
                                            row: row.row.map(cell => {
                                                const parsed = parseProps(cell.props)
                                                return { ...cell, props: formatProps({ ...parsed, borders: [0, 0, 0, 0] }) }
                                            })
                                        }))
                                        updateElement(selectedElement.id, { table: { ...currentTable, rows: updatedRows } })
                                    }}
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    All Borders Off
                                </button>
                            </div>
                        </div>

                        {/* Title Cell Editing */}
                        {selectedCell && selectedCellElement && (
                            <div style={{ padding: '0.75rem', background: 'hsl(var(--muted))', borderRadius: '6px', border: '1px solid hsl(var(--border))' }}>
                                <h4 style={{ margin: '0 0 0.75rem 0', fontSize: '0.85rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>
                                    Title Cell (Row {selectedCell.rowIdx + 1}, Col {selectedCell.colIdx + 1})
                                </h4>

                                {/* Cell Text */}
                                <div style={{ marginBottom: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Text:</label>
                                    <input
                                        type="text"
                                        value={selectedCellElement.text || ''}
                                        onChange={(e) => {
                                            const newRows = [...selectedElement.table.rows]
                                            newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], text: e.target.value }
                                            updateElement(selectedElement.id, { table: { ...selectedElement.table, rows: newRows } })
                                        }}
                                        placeholder="Cell text content"
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                </div>

                                {/* Add Image */}
                                <div style={{ marginBottom: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Add Image:</label>
                                    <input
                                        type="file"
                                        accept="image/*"
                                        onChange={(e) => {
                                            const file = e.target.files[0]
                                            if (file) {
                                                const reader = new FileReader()
                                                reader.onload = (event) => {
                                                    const newRows = [...selectedElement.table.rows]
                                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = {
                                                        ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx],
                                                        image: {
                                                            imagename: file.name,
                                                            imagedata: event.target.result,
                                                            width: 100,
                                                            height: 50
                                                        }
                                                    }
                                                    updateElement(selectedElement.id, { table: { ...selectedElement.table, rows: newRows } })
                                                }
                                                reader.readAsDataURL(file)
                                            }
                                        }}
                                        style={{ width: '100%', fontSize: '0.75rem', padding: '0.4rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                </div>

                                {/* Cell Styling */}
                                <PropsEditor
                                    props={selectedCellElement.props}
                                    onChange={(newProps) => {
                                        const newRows = [...selectedElement.table.rows]
                                        newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], props: newProps }
                                        updateElement(selectedElement.id, { table: { ...selectedElement.table, rows: newRows } })
                                    }}
                                    fonts={fonts}
                                />

                                {/* Link URL */}
                                <div style={{ marginTop: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Link URL:</label>
                                    <input
                                        type="text"
                                        value={selectedCellElement.link || ''}
                                        onChange={(e) => {
                                            const newRows = [...selectedElement.table.rows]
                                            newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], link: e.target.value }
                                            updateElement(selectedElement.id, { table: { ...selectedElement.table, rows: newRows } })
                                        }}
                                        placeholder="https://..."
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                </div>
                            </div>
                        )}
                    </>
                )}

                {/* TABLE Properties */}
                {selectedElement.type === 'table' && (
                    <>
                        {/* Column Count and Layout Controls */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Table</label>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', marginBottom: '0.5rem' }}>
                                <label style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', minWidth: '60px' }}>Columns:</label>
                                <input
                                    type="number"
                                    min="1"
                                    max="10"
                                    value={selectedElement.maxcolumns || 3}
                                    readOnly
                                    style={{ width: '60px', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', color: 'hsl(var(--foreground))' }}
                                />
                            </div>
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                <button
                                    onClick={() => {
                                        const newRow = { row: Array(selectedElement.maxcolumns).fill(null).map(() => ({ props: 'Helvetica:12:000:left:1:1:1:1', text: '' })) }
                                        updateElement(selectedElement.id, { rows: [...(selectedElement.rows || []), newRow] })
                                    }}
                                    style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    + Add Row
                                </button>
                                <button
                                    onClick={() => {
                                        if ((selectedElement.rows?.length || 0) <= 1) return
                                        const newRows = selectedElement.rows.slice(0, -1)
                                        updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    Remove Row (Last)
                                </button>
                                <button
                                    onClick={() => {
                                        const newCols = (selectedElement.maxcolumns || 3) + 1
                                        if (newCols > 10) return
                                        const newWidths = selectedElement.columnwidths ? [...selectedElement.columnwidths, 1] : Array(newCols).fill(1)
                                        const updatedRows = selectedElement.rows.map(r => ({ row: [...r.row, { props: 'Helvetica:12:000:left:1:1:1:1', text: '' }] }))
                                        updateElement(selectedElement.id, { maxcolumns: newCols, rows: updatedRows, columnwidths: newWidths })
                                    }}
                                    style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    + Add Column
                                </button>
                                <button
                                    onClick={() => {
                                        const newCols = (selectedElement.maxcolumns || 3) - 1
                                        if (newCols < 1) return
                                        const updatedRows = selectedElement.rows.map(r => ({ row: r.row.slice(0, -1) }))
                                        const newWidths = selectedElement.columnwidths ? selectedElement.columnwidths.slice(0, -1) : undefined
                                        updateElement(selectedElement.id, { maxcolumns: newCols, rows: updatedRows, columnwidths: newWidths })
                                    }}
                                    style={{ padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    Remove Column (Last)
                                </button>
                            </div>
                            <div style={{ marginTop: '0.5rem', fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>
                                Rows: {selectedElement.rows?.length || 0}, Columns: {selectedElement.maxcolumns || 3}
                            </div>
                        </div>

                        {/* Table Borders Toggle */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Table Borders</label>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <button
                                    onClick={() => {
                                        // Toggle all borders to 1:1:1:1
                                        const updatedRows = selectedElement.rows.map(row => ({
                                            ...row,
                                            row: row.row.map(cell => {
                                                const parsed = parseProps(cell.props)
                                                return { ...cell, props: formatProps({ ...parsed, borders: [1, 1, 1, 1] }) }
                                            })
                                        }))
                                        updateElement(selectedElement.id, { rows: updatedRows })
                                    }}
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    All Borders On
                                </button>
                                <button
                                    onClick={() => {
                                        // Toggle all borders to 0:0:0:0
                                        const updatedRows = selectedElement.rows.map(row => ({
                                            ...row,
                                            row: row.row.map(cell => {
                                                const parsed = parseProps(cell.props)
                                                return { ...cell, props: formatProps({ ...parsed, borders: [0, 0, 0, 0] }) }
                                            })
                                        }))
                                        updateElement(selectedElement.id, { rows: updatedRows })
                                    }}
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', cursor: 'pointer' }}
                                >
                                    All Borders Off
                                </button>
                            </div>
                        </div>

                        {/* Column Widths */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Column Widths (weights)</label>
                            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(60px, 1fr))', gap: '0.5rem', marginBottom: '0.5rem' }}>
                                {Array.from({ length: selectedElement.maxcolumns || 3 }).map((_, idx) => {
                                    const currentWidths = selectedElement.columnwidths || Array(selectedElement.maxcolumns).fill(1)
                                    return (
                                        <div key={idx}>
                                            <label style={{ display: 'block', fontSize: '0.7rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Col {idx + 1}</label>
                                            <input
                                                type="number"
                                                min="0.1"
                                                step="0.1"
                                                value={currentWidths[idx] || 1}
                                                onChange={(e) => {
                                                    const newWidths = [...currentWidths]
                                                    newWidths[idx] = parseFloat(e.target.value) || 1
                                                    updateElement(selectedElement.id, { columnwidths: newWidths })
                                                }}
                                                style={{ width: '100%', padding: '0.35rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                            />
                                        </div>
                                    )
                                })}
                            </div>
                            <button
                                onClick={() => {
                                    const equalWidths = Array(selectedElement.maxcolumns).fill(1)
                                    updateElement(selectedElement.id, { columnwidths: equalWidths })
                                }}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                            >
                                Reset to Equal
                            </button>
                        </div>

                        {/* Row Heights */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Row Heights (multipliers)</label>
                            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(60px, 1fr))', gap: '0.5rem', marginBottom: '0.5rem', maxHeight: '200px', overflowY: 'auto' }}>
                                {selectedElement.rows?.map((row, idx) => {
                                    const rowHeight = row.height || 1.0
                                    return (
                                        <div key={idx}>
                                            <label style={{ display: 'block', fontSize: '0.7rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Row {idx + 1}</label>
                                            <input
                                                type="number"
                                                min="0.1"
                                                step="0.1"
                                                value={rowHeight}
                                                onChange={(e) => {
                                                    const newRows = [...selectedElement.rows]
                                                    newRows[idx] = { ...newRows[idx], height: parseFloat(e.target.value) || 1.0 }
                                                    updateElement(selectedElement.id, { rows: newRows })
                                                }}
                                                style={{ width: '100%', padding: '0.35rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                            />
                                        </div>
                                    )
                                })}
                            </div>
                            <button
                                onClick={() => {
                                    const newRows = selectedElement.rows.map(row => ({ ...row, height: 1.0 }))
                                    updateElement(selectedElement.id, { rows: newRows })
                                }}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                            >
                                Reset to Default
                            </button>
                        </div>

                        {/* Table Background Color */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Table Background Color</label>
                            <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                Sets the default background color for all cells. Individual cells can override this.
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                <input
                                    type="color"
                                    value={selectedElement.bgcolor || '#ffffff'}
                                    onChange={(e) => updateElement(selectedElement.id, { bgcolor: e.target.value })}
                                    style={{ width: '48px', height: '32px', border: '1px solid hsl(var(--border))', borderRadius: '4px', cursor: 'pointer', padding: '2px', WebkitAppearance: 'none', background: 'transparent' }}
                                />
                                <input
                                    type="text"
                                    value={selectedElement.bgcolor || '#ffffff'}
                                    onChange={(e) => updateElement(selectedElement.id, { bgcolor: e.target.value })}
                                    placeholder="#RRGGBB or transparent"
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                                <button
                                    onClick={() => updateElement(selectedElement.id, { bgcolor: '' })}
                                    style={{ padding: '0.4rem 0.6rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                                >
                                    Clear
                                </button>
                            </div>
                            <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                {tableBackgroundPresets.map(({ label, color }) => (
                                    <button
                                        key={color}
                                        onClick={() => updateElement(selectedElement.id, { bgcolor: color })}
                                        style={{ width: '24px', height: '24px', border: '1px solid #999', borderRadius: '4px', background: color, cursor: 'pointer', boxShadow: 'inset 0 0 0 1px rgba(0,0,0,0.1)' }}
                                        title={label}
                                    />
                                ))}
                            </div>
                        </div>

                        {/* Table Text Color */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Table Text Color</label>
                            <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                Sets the default text color for all cells. Individual cells can override this.
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                <input
                                    type="color"
                                    value={selectedElement.textcolor || '#000000'}
                                    onChange={(e) => updateElement(selectedElement.id, { textcolor: e.target.value })}
                                    style={{ width: '48px', height: '32px', border: '1px solid hsl(var(--border))', borderRadius: '4px', cursor: 'pointer', padding: '2px', WebkitAppearance: 'none', background: 'transparent' }}
                                />
                                <input
                                    type="text"
                                    value={selectedElement.textcolor || '#000000'}
                                    onChange={(e) => updateElement(selectedElement.id, { textcolor: e.target.value })}
                                    placeholder="#RRGGBB (default: black)"
                                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                                <button
                                    onClick={() => updateElement(selectedElement.id, { textcolor: '' })}
                                    style={{ padding: '0.4rem 0.6rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                                >
                                    Clear
                                </button>
                            </div>
                            <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                {cellTextPresets.map(color => (
                                    <button
                                        key={color}
                                        onClick={() => updateElement(selectedElement.id, { textcolor: color })}
                                        style={{ width: '24px', height: '24px', border: '2px solid hsl(var(--border))', borderRadius: '4px', background: color, cursor: 'pointer' }}
                                        title={color}
                                    />
                                ))}
                            </div>
                        </div>

                        {/* Cell Editing */}
                        {selectedCell && selectedCellElement && (
                            <div style={{ padding: '0.75rem', background: 'hsl(var(--muted))', borderRadius: '6px', border: '1px solid hsl(var(--border))' }}>
                                <h4 style={{ margin: '0 0 0.75rem 0', fontSize: '0.85rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>
                                    Cell (Row {selectedCell.rowIdx + 1}, Col {selectedCell.colIdx + 1})
                                </h4>

                                {/* Cell Text */}
                                <div style={{ marginBottom: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Text:</label>
                                    <input
                                        type="text"
                                        value={selectedCellElement.text || ''}
                                        onChange={(e) => {
                                            const newRows = [...selectedElement.rows]
                                            newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], text: e.target.value }
                                            updateElement(selectedElement.id, { rows: newRows })
                                        }}
                                        placeholder="Cell text content"
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                </div>

                                {/* Cell Styling */}
                                <PropsEditor
                                    props={selectedCellElement.props}
                                    onChange={(newProps) => {
                                        const newRows = [...selectedElement.rows]
                                        newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], props: newProps }
                                        updateElement(selectedElement.id, { rows: newRows })
                                    }}
                                    fonts={fonts}
                                />

                                {/* Link URL */}
                                <div style={{ marginTop: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Link URL:</label>
                                    <input
                                        type="text"
                                        value={selectedCellElement.link || ''}
                                        onChange={(e) => {
                                            const newRows = [...selectedElement.rows]
                                            newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], link: e.target.value }
                                            updateElement(selectedElement.id, { rows: newRows })
                                        }}
                                        placeholder="https://... or #bookmark-id"
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                    <div style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))', marginTop: '0.25rem' }}>
                                        Use # prefix for internal links (e.g., #section-id)
                                    </div>
                                </div>

                                {/* Destination ID (dest) for bookmark target */}
                                <div style={{ marginTop: '0.5rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Destination ID (dest):</label>
                                    <input
                                        type="text"
                                        value={selectedCellElement.dest || ''}
                                        onChange={(e) => {
                                            const newRows = [...selectedElement.rows]
                                            const val = e.target.value || undefined
                                            newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], dest: val }
                                            updateElement(selectedElement.id, { rows: newRows })
                                        }}
                                        placeholder="e.g., financial-summary"
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                    <div style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))', marginTop: '0.25rem' }}>
                                        ID used as bookmark target for internal links
                                    </div>
                                </div>

                                {/* Cell Size Override */}
                                <div style={{ marginTop: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Cell Size Override</label>
                                    <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                        ⚠️ Only use Blue handle (right-adjust to resize width, resizing height adjusts cell height to value in its JSON).
                                    </div>
                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.7rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Width (px):</label>
                                            <input
                                                type="number"
                                                min="10"
                                                value={selectedCellElement.width || ''}
                                                onChange={(e) => {
                                                    const newRows = [...selectedElement.rows]
                                                    const val = e.target.value ? parseInt(e.target.value) : undefined
                                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], width: val }
                                                    updateElement(selectedElement.id, { rows: newRows })
                                                }}
                                                placeholder="Auto"
                                                style={{ width: '100%', padding: '0.35rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                            />
                                        </div>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.7rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Height (px):</label>
                                            <input
                                                type="number"
                                                min="10"
                                                value={selectedCellElement.height || ''}
                                                onChange={(e) => {
                                                    const newRows = [...selectedElement.rows]
                                                    const val = e.target.value ? parseInt(e.target.value) : undefined
                                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], height: val }
                                                    updateElement(selectedElement.id, { rows: newRows })
                                                }}
                                                placeholder="Auto"
                                                style={{ width: '100%', padding: '0.35rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                            />
                                        </div>
                                    </div>
                                </div>

                                {/* Cell Background Color */}
                                <div style={{ marginTop: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Cell Background Color</label>
                                    <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                        Overrides (from bgcolor #FFFFFF)
                                    </div>
                                    <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                        <input
                                            type="color"
                                            value={selectedCellElement.bgcolor || '#ffffff'}
                                            onChange={(e) => {
                                                const newRows = [...selectedElement.rows]
                                                newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], bgcolor: e.target.value }
                                                updateElement(selectedElement.id, { rows: newRows })
                                            }}
                                            style={{ width: '48px', height: '32px', border: '1px solid hsl(var(--border))', borderRadius: '4px', cursor: 'pointer', padding: '2px', WebkitAppearance: 'none', background: 'transparent' }}
                                        />
                                        <input
                                            type="text"
                                            value={selectedCellElement.bgcolor || ''}
                                            onChange={(e) => {
                                                const newRows = [...selectedElement.rows]
                                                newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], bgcolor: e.target.value }
                                                updateElement(selectedElement.id, { rows: newRows })
                                            }}
                                            placeholder="#RRGGBB"
                                            style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                        />
                                        <button
                                            onClick={() => {
                                                const newRows = [...selectedElement.rows]
                                                newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], bgcolor: undefined }
                                                updateElement(selectedElement.id, { rows: newRows })
                                            }}
                                            style={{ padding: '0.4rem 0.6rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                                        >
                                            Clear
                                        </button>
                                    </div>
                                    <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                        {cellBackgroundPresets.map(({ label, color }) => (
                                            <button
                                                key={color}
                                                onClick={() => {
                                                    const newRows = [...selectedElement.rows]
                                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], bgcolor: color }
                                                    updateElement(selectedElement.id, { rows: newRows })
                                                }}
                                                style={{ width: '24px', height: '24px', border: '1px solid #999', borderRadius: '4px', background: color, cursor: 'pointer', boxShadow: 'inset 0 0 0 1px rgba(0,0,0,0.1)' }}
                                                title={label}
                                            />
                                        ))}
                                    </div>
                                </div>

                                {/* Cell Text Color */}
                                <div style={{ marginTop: '0.75rem' }}>
                                    <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Cell Text Color</label>
                                    <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginBottom: '0.5rem' }}>
                                        Sets the text color (default: black)
                                    </div>
                                    <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                        <input
                                            type="color"
                                            value={selectedCellElement.textcolor || '#000000'}
                                            onChange={(e) => {
                                                const newRows = [...selectedElement.rows]
                                                newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], textcolor: e.target.value }
                                                updateElement(selectedElement.id, { rows: newRows })
                                            }}
                                            style={{ width: '48px', height: '32px', border: '1px solid hsl(var(--border))', borderRadius: '4px', cursor: 'pointer', padding: '2px', WebkitAppearance: 'none', background: 'transparent' }}
                                        />
                                        <input
                                            type="text"
                                            value={selectedCellElement.textcolor || ''}
                                            onChange={(e) => {
                                                const newRows = [...selectedElement.rows]
                                                newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], textcolor: e.target.value }
                                                updateElement(selectedElement.id, { rows: newRows })
                                            }}
                                            placeholder="#RRGGBB"
                                            style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                        />
                                        <button
                                            onClick={() => {
                                                const newRows = [...selectedElement.rows]
                                                newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], textcolor: undefined }
                                                updateElement(selectedElement.id, { rows: newRows })
                                            }}
                                            style={{ padding: '0.4rem 0.6rem', fontSize: '0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--muted))', cursor: 'pointer' }}
                                        >
                                            Clear
                                        </button>
                                    </div>
                                    <div style={{ display: 'flex', gap: '0.25rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                                        {cellTextPresets.map(color => (
                                            <button
                                                key={color}
                                                onClick={() => {
                                                    const newRows = [...selectedElement.rows]
                                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx] = { ...newRows[selectedCell.rowIdx].row[selectedCell.colIdx], textcolor: color }
                                                    updateElement(selectedElement.id, { rows: newRows })
                                                }}
                                                style={{ width: '24px', height: '24px', border: '2px solid hsl(var(--border))', borderRadius: '4px', background: color, cursor: 'pointer' }}
                                                title={color}
                                            />
                                        ))}
                                    </div>
                                </div>
                            </div>
                        )}
                    </>
                )}

                {/* FOOTER Properties */}
                {selectedElement.type === 'footer' && (
                    <>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Text:</label>
                            <input
                                type="text"
                                value={selectedElement.text || ''}
                                onChange={(e) => updateElement(selectedElement.id, { text: e.target.value })}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                        <PropsEditor
                            props={selectedElement.props}
                            onChange={(newProps) => updateElement(selectedElement.id, { props: newProps })}
                            fonts={fonts}
                        />
                    </>
                )}

                {/* IMAGE Properties */}
                {selectedElement.type === 'image' && (
                    <>
                        {/* Image Upload/Change */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Image Source</label>
                            <input
                                type="file"
                                accept="image/*"
                                onChange={(e) => {
                                    const file = e.target.files[0]
                                    if (file) {
                                        const reader = new FileReader()
                                        reader.onload = (event) => {
                                            updateElement(selectedElement.id, {
                                                imagename: file.name,
                                                imagedata: event.target.result
                                            })
                                        }
                                        reader.readAsDataURL(file)
                                    }
                                }}
                                style={{ width: '100%', fontSize: '0.8rem', padding: '0.5rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                            {selectedElement.imagename && (
                                <div style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))', marginTop: '0.5rem', wordBreak: 'break-all' }}>
                                    Current: {selectedElement.imagename}
                                </div>
                            )}
                        </div>

                        {/* Image Dimensions */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Dimensions</label>
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                <div>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Width (px)</label>
                                    <input
                                        type="number"
                                        min="10"
                                        max="1000"
                                        value={selectedElement.width || 200}
                                        onChange={(e) => updateElement(selectedElement.id, { width: parseInt(e.target.value) || 200 })}
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                </div>
                                <div>
                                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Height (px)</label>
                                    <input
                                        type="number"
                                        min="10"
                                        max="1000"
                                        value={selectedElement.height || 150}
                                        onChange={(e) => updateElement(selectedElement.id, { height: parseInt(e.target.value) || 150 })}
                                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                    />
                                </div>
                            </div>
                        </div>

                        {/* Link URL */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '0.5rem', fontWeight: '600', color: 'hsl(var(--foreground))' }}>Link URL</label>
                            <input
                                type="text"
                                value={selectedElement.link || ''}
                                onChange={(e) => updateElement(selectedElement.id, { link: e.target.value })}
                                placeholder="https://..."
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                    </>
                )}

                {/* SPACER Properties */}
                {selectedElement.type === 'spacer' && (
                    <div>
                        <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Height (px):</label>
                        <input
                            type="number"
                            min="1"
                            max="200"
                            value={selectedElement.height || 20}
                            onChange={(e) => updateElement(selectedElement.id, { height: parseInt(e.target.value) || 20 })}
                            style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                        />
                    </div>
                )}

                {/* IMAGE Properties */}
                {selectedElement.type === 'image' && (
                    <>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Select Image:</label>
                            <input
                                type="file"
                                accept="image/*"
                                onChange={(e) => {
                                    const file = e.target.files[0]
                                    if (file) {
                                        const reader = new FileReader()
                                        reader.onload = (event) => {
                                            updateElement(selectedElement.id, {
                                                imagedata: event.target.result.split(',')[1],
                                                imagename: file.name
                                            })
                                        }
                                        reader.readAsDataURL(file)
                                    }
                                }}
                                style={{ width: '100%', fontSize: '0.75rem', padding: '0.4rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Image Name:</label>
                            <input
                                type="text"
                                value={selectedElement.imagename || ''}
                                onChange={(e) => updateElement(selectedElement.id, { imagename: e.target.value })}
                                placeholder="Image name"
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                            <div>
                                <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Width (px):</label>
                                <input
                                    type="number"
                                    min="10"
                                    max="800"
                                    value={selectedElement.width || 200}
                                    onChange={(e) => updateElement(selectedElement.id, { width: parseInt(e.target.value) || 200 })}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                            </div>
                            <div>
                                <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Height (px):</label>
                                <input
                                    type="number"
                                    min="10"
                                    max="800"
                                    value={selectedElement.height || 150}
                                    onChange={(e) => updateElement(selectedElement.id, { height: parseInt(e.target.value) || 150 })}
                                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                                />
                            </div>
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: '600', marginBottom: '0.5rem', color: 'hsl(var(--foreground))' }}>Link URL:</label>
                            <input
                                type="text"
                                value={selectedElement.link || ''}
                                onChange={(e) => updateElement(selectedElement.id, { link: e.target.value })}
                                placeholder="https://..."
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.8rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                    </>
                )}
            </div>

            <style jsx>{`
                .btn-border { 
                    padding: 0.25rem 0.5rem; 
                    border: 1px solid hsl(var(--border)); 
                    background: hsl(var(--background)); 
                    border-radius: 4px; 
                    cursor: pointer; 
                    font-size: 0.75rem;
                }
                .btn-border:disabled { 
                    opacity: 0.5; 
                    cursor: not-allowed; 
                }
            `}</style>
        </div>
    )
}
