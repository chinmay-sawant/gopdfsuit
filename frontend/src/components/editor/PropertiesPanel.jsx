
import React from 'react'
import { Edit, Settings, Trash2 } from 'lucide-react'
import { formatProps, parseProps } from './utils'
import { DEFAULT_FONTS } from './constants'

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
                <button className="btn-border" onClick={() => updateBorder(index, parsed.borders[index] - 1)} disabled={parsed.borders[index] <= 0}>−</button>
                <span style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem', minWidth: '2rem', textAlign: 'center', background: 'hsl(var(--muted))', borderRadius: '4px' }}>{parsed.borders[index]}px</span>
                <button className="btn-border" onClick={() => updateBorder(index, parsed.borders[index] + 1)} disabled={parsed.borders[index] >= 10}>+</button>
            </div>
        </div>
    )

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                <div>
                    <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Family</label>
                    <select value={parsed.font} onChange={(e) => onChange(formatProps({ ...parsed, font: e.target.value }))} style={{ width: '100%', padding: '0.5rem', border: '1px solid hsl(var(--border))', borderRadius: '6px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', fontSize: '0.9rem' }}>
                        {fonts.map(font => <option key={font.id} value={font.id}>{font.displayName}</option>)}
                    </select>
                </div>
                <div>
                    <label style={{ display: 'block', fontSize: '0.8rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Size</label>
                    <select value={parsed.size} onChange={(e) => onChange(formatProps({ ...parsed, size: parseInt(e.target.value) }))} style={{ width: '100%', padding: '0.5rem', border: '1px solid hsl(var(--border))', borderRadius: '6px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', fontSize: '0.9rem' }}>
                        {[8, 9, 10, 11, 12, 14, 16, 18, 20, 24, 28, 32, 36, 48, 72].map(size => <option key={size} value={size}>{size}px</option>)}
                    </select>
                </div>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
                {[{ key: 0, label: 'B' }, { key: 1, label: 'I' }, { key: 2, label: 'U' }].map(({ key, label }) => (
                    <button key={key} onClick={() => { const s = parsed.style.split(''); s[key] = s[key] === '1' ? '0' : '1'; onChange(formatProps({ ...parsed, style: s.join('') })) }} style={{ padding: '0.5rem 0.75rem', border: '1px solid hsl(var(--border))', borderRadius: '6px', background: parsed.style[key] === '1' ? 'hsl(var(--accent))' : 'hsl(var(--background))', color: parsed.style[key] === '1' ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))', fontSize: '0.9rem', fontWeight: parsed.style[key] === '1' ? '600' : '400', cursor: 'pointer' }}>{label}</button>
                ))}
            </div>
            <div style={{ display: 'flex', gap: '0.25rem' }}>
                {['left', 'center', 'right'].map((align) => (
                    <button key={align} onClick={() => onChange(formatProps({ ...parsed, align }))} style={{ flex: 1, padding: '0.5rem', border: '1px solid hsl(var(--border))', borderRadius: '6px', background: parsed.align === align ? 'hsl(var(--accent))' : 'hsl(var(--background))', color: parsed.align === align ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))', fontSize: '0.9rem', cursor: 'pointer' }}>{align}</button>
                ))}
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
                <BorderControls label="Left" index={0} />
                <BorderControls label="Right" index={1} />
                <BorderControls label="Top" index={2} />
                <BorderControls label="Bottom" index={3} />
            </div>
        </div>
    )
}

export default function PropertiesPanel({ selectedElement, selectedCell, updateElement, deleteElement, setSelectedCell, fonts }) {
    if (!selectedElement) {
        return (
            <div className="card" style={{ padding: '2rem 1rem', textAlign: 'center', color: 'hsl(var(--muted-foreground))', background: 'hsl(var(--muted))' }}>
                <Settings size={24} style={{ opacity: 0.3, marginBottom: '0.5rem' }} />
                <p style={{ fontSize: '0.85rem', margin: 0 }}>Select a component to edit</p>
            </div>
        )
    }

    const handleDelete = () => deleteElement(selectedElement.id)

    return (
        <div className="card" style={{ padding: '1rem' }}>
            <h3 style={{ margin: '0 0 0.75rem 0', fontSize: '0.9rem', fontWeight: '600', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'hsl(var(--foreground))' }}>
                <Edit size={16} /> Properties: {selectedElement.type}
            </h3>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                {selectedElement.type === 'title' && (
                    <>
                        <div><label>Background Color:</label><input type="color" value={selectedElement.bgcolor || '#ffffff'} onChange={(e) => updateElement(selectedElement.id, { bgcolor: e.target.value })} /></div>
                        <div><label>Text Color:</label><input type="color" value={selectedElement.textcolor || '#000000'} onChange={(e) => updateElement(selectedElement.id, { textcolor: e.target.value })} /></div>
                    </>
                )}

                {selectedElement.type === 'table' && (
                    <>
                        <div>
                            <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem' }}>Layout</label>
                            <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '0.5rem' }}>
                                <button onClick={() => {
                                    const table = selectedElement;
                                    const newRow = { row: Array(table.maxcolumns).fill(null).map(() => ({ props: 'Helvetica:12:000:left:0:0:0:0', text: '' })) };
                                    updateElement(selectedElement.id, { rows: [...(table.rows || []), newRow] })
                                }} className="btn" style={{ flex: 1, fontSize: '0.75rem', padding: '0.4rem' }}>+ Row</button>
                                <button onClick={() => {
                                    const table = selectedElement;
                                    if ((table.rows?.length || 0) <= 1) return;
                                    const newRows = table.rows.slice(0, -1);
                                    updateElement(selectedElement.id, { rows: newRows })
                                }} className="btn" style={{ flex: 1, fontSize: '0.75rem', padding: '0.4rem' }}>- Row</button>
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <button onClick={() => {
                                    const table = selectedElement;
                                    const newCols = (table.maxcolumns || 3) + 1;
                                    if (newCols > 10) return;
                                    // Logic to resize columns
                                    const newWidths = table.columnwidths ? [...table.columnwidths, 1] : Array(newCols).fill(1).map(() => 1);
                                    // Normalize (simplified logic for robustness)
                                    const updatedRows = table.rows.map(r => ({ row: [...r.row, { props: 'Helvetica:12:000:left:0:0:0:0', text: '' }] }));
                                    updateElement(selectedElement.id, { maxcolumns: newCols, rows: updatedRows, columnwidths: newWidths })
                                }} className="btn" style={{ flex: 1, fontSize: '0.75rem', padding: '0.4rem' }}>+ Column</button>
                                <button onClick={() => {
                                    const table = selectedElement;
                                    const newCols = (table.maxcolumns || 3) - 1;
                                    if (newCols < 1) return;
                                    const updatedRows = table.rows.map(r => ({ row: r.row.slice(0, -1) }));
                                    updateElement(selectedElement.id, { maxcolumns: newCols, rows: updatedRows })
                                }} className="btn" style={{ flex: 1, fontSize: '0.75rem', padding: '0.4rem' }}>- Column</button>
                            </div>
                        </div>
                    </>
                )}

                {(selectedElement.type === 'footer' || selectedElement.type === 'title') && (
                    !selectedCell && (
                        <div><label>Text:</label><input type="text" value={selectedElement.text || ''} onChange={(e) => updateElement(selectedElement.id, { text: e.target.value })} style={{ width: '100%', padding: '0.4rem' }} /></div>
                    )
                )}

                {(selectedElement.type === 'footer' || (selectedElement.type === 'table' && selectedCell)) && (
                    <div>
                        <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '600' }}>
                            {selectedElement.type === 'table' ? `Cell (${selectedCell.rowIdx + 1},${selectedCell.colIdx + 1}) Style` : 'Element Style'}
                        </label>
                        <PropsEditor
                            props={
                                selectedElement.type === 'table'
                                    ? selectedElement.rows[selectedCell.rowIdx].row[selectedCell.colIdx].props
                                    : selectedElement.props
                            }
                            onChange={(newProps) => {
                                if (selectedElement.type === 'table') {
                                    const newRows = [...selectedElement.rows];
                                    newRows[selectedCell.rowIdx].row[selectedCell.colIdx].props = newProps;
                                    updateElement(selectedElement.id, { rows: newRows });
                                } else {
                                    updateElement(selectedElement.id, { props: newProps });
                                }
                            }}
                            fonts={fonts}
                        />
                        {selectedElement.type === 'table' && (
                            <div style={{ marginTop: '0.5rem' }}>
                                <label>Cell Text:</label>
                                <input
                                    type="text"
                                    value={selectedElement.rows[selectedCell.rowIdx].row[selectedCell.colIdx].text || ''}
                                    onChange={(e) => {
                                        const newRows = [...selectedElement.rows];
                                        newRows[selectedCell.rowIdx].row[selectedCell.colIdx].text = e.target.value;
                                        updateElement(selectedElement.id, { rows: newRows });
                                    }}
                                    style={{ width: '100%', padding: '0.4rem' }}
                                />
                            </div>
                        )}
                    </div>
                )}

                <button onClick={handleDelete} className="btn-destructive" style={{ marginTop: '1rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '0.5rem', background: 'hsl(var(--destructive))', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
                    <Trash2 size={16} /> Delete Element
                </button>
            </div>

            <style jsx>{`
        .btn-border { padding: 0.25rem 0.5rem; border: 1px solid hsl(var(--border)); background: hsl(var(--background)); border-radius: 4px; cursor: pointer; }
        .btn-border:disabled { opacity: 0.5; cursor: not-allowed; }
      `}</style>
        </div>
    )
}
