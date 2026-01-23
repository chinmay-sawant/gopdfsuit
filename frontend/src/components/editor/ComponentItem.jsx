
import React, { useState } from 'react'
import { Table, FileText, Minus, Image as ImageIcon, ChevronUp, ChevronDown, X, GripVertical } from 'lucide-react'
import { getStyleFromProps, getUsableWidth } from './utils'

export default function ComponentItem({ element, index, isSelected, onSelect, onUpdate, onMove, onDelete, canMoveUp, canMoveDown, selectedCell, onCellSelect, onDragStart, onDragEnd, onDrop, isDragging, draggedType, handleCellDrop, currentPageSize }) {
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
                                                        if (draggedType === 'checkbox' || draggedType === 'image' || draggedType === 'radio' || draggedType === 'text_input' || draggedType === 'hyperlink') {
                                                            e.preventDefault()
                                                            e.stopPropagation()
                                                        }
                                                    }}
                                                    onDrop={(e) => {
                                                        e.preventDefault()
                                                        e.stopPropagation()
                                                        const draggedData = e.dataTransfer.getData('text/plain')
                                                        if (draggedData === 'checkbox' || draggedData === 'image' || draggedData === 'radio' || draggedData === 'text_input' || draggedData === 'hyperlink') {
                                                            handleCellDrop(element, onUpdate, rowIdx, colIdx, draggedData)
                                                        }
                                                    }}
                                                    className={(draggedType === 'checkbox' || draggedType === 'image' || draggedType === 'radio' || draggedType === 'text_input' || draggedType === 'hyperlink') ? 'drop-target' : ''}
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
                                                            onMouseDown={(e) => handleCellWidthResizeStart(e, rowIdx, colIdx)}
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
                                                        onMouseDown={(e) => handleCellHeightResizeStart(e, rowIdx, colIdx)}
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
