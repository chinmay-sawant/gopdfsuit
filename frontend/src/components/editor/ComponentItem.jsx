
import { useState } from 'react'
import { Image as ImageIcon, ChevronUp, ChevronDown, X, GripVertical } from 'lucide-react'
import { getStyleFromProps, getUsableWidth, getImageSrc } from './utils'

export default function ComponentItem({ element, index, isSelected, onSelect, onUpdate, onMove, onDelete, canMoveUp, canMoveDown, selectedCell, onCellSelect, onDragStart, onDragEnd, onDrop, isDragging, draggedType, handleCellDrop, currentPageSize }) {
    const [, setIsResizing] = useState(false)

    const handleClick = (e) => {
        e.stopPropagation()
        onSelect(element.id)
        onCellSelect(null) // Clear cell selection when table is selected
    }

    const handleCellClick = (rowIdx, colIdx, e) => {
        if (e) e.stopPropagation()
        onSelect(element.id)
        onCellSelect({ elementId: element.id, rowIdx, colIdx })
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
            // Pass both draggedId and targetId for reordering logic
            onDrop(draggedId, element.id)
        }
    }

    const renderContent = () => {
        switch (element.type) {
            case 'title': {
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

                // Normalize column weights for title table
                const rawWeightsTitle = titleTable.columnwidths && titleTable.columnwidths.length === titleTable.maxcolumns
                    ? titleTable.columnwidths
                    : Array(titleTable.maxcolumns).fill(1)
                const totalWeightTitle = rawWeightsTitle.reduce((sum, w) => sum + w, 0)
                const colWeightsTitle = rawWeightsTitle.map(w => w / totalWeightTitle)

                // Column width resize handler for title table - per-cell width adjustment
                const handleTitleCellWidthResizeStart = (e, rowIdx, colIdx) => {
                    e.preventDefault()
                    e.stopPropagation()
                    const startX = e.clientX
                    const numCols = titleTable.maxcolumns || 3
                    const row = titleTable.rows[rowIdx]

                    // Get current widths for this specific row (use cell.width if set, otherwise column default)
                    const currentRowWidths = row.row.map((cell, idx) =>
                        cell.width !== undefined ? cell.width : (usableWidthForTitle * colWeightsTitle[idx])
                    )
                    const startWidth = currentRowWidths[colIdx]

                    // Determine which adjacent column will compensate
                    const adjacentColIdx = colIdx < numCols - 1 ? colIdx + 1 : colIdx - 1
                    const adjacentStartWidth = currentRowWidths[adjacentColIdx]
                    const minCellWidth = 30

                    const onMouseMove = (me) => {
                        const dx = me.clientX - startX

                        let newWidth = startWidth + dx
                        let adjacentNewWidth = adjacentStartWidth - dx

                        // Enforce minimum widths
                        if (newWidth < minCellWidth) {
                            newWidth = minCellWidth
                            adjacentNewWidth = startWidth + adjacentStartWidth - minCellWidth
                        }
                        if (adjacentNewWidth < minCellWidth) {
                            adjacentNewWidth = minCellWidth
                            newWidth = startWidth + adjacentStartWidth - minCellWidth
                        }

                        // Update only this specific row's cell widths
                        const newRows = [...titleTable.rows]
                        newRows[rowIdx] = {
                            ...newRows[rowIdx],
                            row: newRows[rowIdx].row.map((c, idx) => {
                                if (idx === colIdx) {
                                    return { ...c, width: newWidth }
                                } else if (idx === adjacentColIdx) {
                                    return { ...c, width: adjacentNewWidth }
                                }
                                // Preserve existing width or set default for other cells
                                return c.width !== undefined ? c : { ...c, width: currentRowWidths[idx] }
                            })
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
                        overflow: 'hidden'
                    }}>
                        {/* Use div-based layout for per-cell width control */
                            <div style={{ width: `${usableWidthForTitle}px` }}>
                                {titleTable.rows?.map((row, rowIdx) => (
                                    <div key={rowIdx} style={{ display: 'flex', position: 'relative' }}>
                                        {row.row?.map((cell, colIdx) => {
                                            const cellStyle = getStyleFromProps(cell.props)
                                            // Fix: Check elementId to prevent highlighting title when other tables are selected
                                            const isCellSelected = selectedCell && selectedCell.elementId === element.id && selectedCell.rowIdx === rowIdx && selectedCell.colIdx === colIdx

                                            // Use individual cell width if set, otherwise use column-based width
                                            const cellWidth = cell.width !== undefined ? cell.width : (usableWidthForTitle * colWeightsTitle[colIdx])
                                            const cellHeight = cell.height || 50

                                            const hasBorder = cellStyle.borderLeftWidth !== '0px' || cellStyle.borderRightWidth !== '0px' ||
                                                cellStyle.borderTopWidth !== '0px' || cellStyle.borderBottomWidth !== '0px'

                                            // Determine background color for title cells
                                            const titleCellBgColor = cell.bgcolor || element.bgcolor || '#fff'

                                            return (
                                                <div
                                                    key={colIdx}
                                                    style={{
                                                        borderLeft: hasBorder ? `${cellStyle.borderLeftWidth} solid #333` : 'none',
                                                        borderRight: hasBorder ? `${cellStyle.borderRightWidth} solid #333` : 'none',
                                                        borderTop: hasBorder ? `${cellStyle.borderTopWidth} solid #333` : 'none',
                                                        borderBottom: hasBorder ? `${cellStyle.borderBottomWidth} solid #333` : 'none',
                                                        padding: '4px 8px',
                                                        width: `${cellWidth}px`,
                                                        height: `${cellHeight}px`,
                                                        minHeight: '30px',
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                        overflow: 'hidden',
                                                        backgroundColor: titleCellBgColor,
                                                        cursor: 'pointer',
                                                        position: 'relative',
                                                        boxSizing: 'border-box',
                                                        outline: isCellSelected ? '2px solid #3b82f6' : 'none',
                                                        outlineOffset: '-2px',
                                                        flexShrink: 0
                                                    }}
                                                    onClick={(e) => {
                                                        e.stopPropagation()
                                                        onSelect(element.id)
                                                        onCellSelect({ elementId: element.id, rowIdx, colIdx })
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
                                                                onCellSelect({ elementId: element.id, rowIdx, colIdx })
                                                            }}
                                                        >
                                                            <img
                                                                src={getImageSrc(cell.image.imagedata, cell.image.imagename)}
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
                                                                    onCellSelect({ elementId: element.id, rowIdx, colIdx })
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
                                                                    onCellSelect({ elementId: element.id, rowIdx, colIdx })
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
                                                                        onCellSelect({ elementId: element.id, rowIdx, colIdx })
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

                                                    {/* Width resize handle - show on all cells */}
                                                    <div
                                                        onMouseDown={(e) => handleTitleCellWidthResizeStart(e, rowIdx, colIdx)}
                                                        style={{
                                                            position: 'absolute',
                                                            right: '-2px',
                                                            top: 0,
                                                            bottom: 0,
                                                            width: '6px',
                                                            cursor: 'col-resize',
                                                            background: isCellSelected ? 'rgba(25, 118, 210, 0.5)' : 'transparent',
                                                            zIndex: 5
                                                        }}
                                                        onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(59, 130, 246, 0.5)'}
                                                        onMouseLeave={(e) => e.currentTarget.style.background = isCellSelected ? 'rgba(25, 118, 210, 0.5)' : 'transparent'}
                                                        title="Drag to resize this cell width"
                                                    />

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
                                                </div>
                                            )
                                        })}
                                    </div>
                                ))}
                            </div>
                        }
                    </div>
                )
            }
            case 'table': {
                // Get page dimensions for width calculations
                // Use passed currentPageSize prop
                const usableWidthForTable = getUsableWidth(currentPageSize.width)

                // Normalize columnwidths so they represent fractions that sum to 1
                const rawColWidths = element.columnwidths && element.columnwidths.length === element.maxcolumns
                    ? element.columnwidths
                    : Array(element.maxcolumns).fill(1)
                const totalWeight = rawColWidths.reduce((sum, w) => sum + w, 0)
                const colWeights = rawColWidths.map(w => w / totalWeight)

                // Per-cell width resize handler - affects only individual cell in that specific row
                // Adjusts the cell width and the adjacent cell compensates to maintain row total
                const handleCellWidthResizeStart = (e, rowIdx, colIdx) => {
                    e.preventDefault()
                    e.stopPropagation()
                    setIsResizing(true)
                    const startX = e.clientX
                    const numCols = element.maxcolumns || 3
                    const row = element.rows[rowIdx]

                    // Get current widths for this specific row (use cell.width if set, otherwise column default)
                    const currentRowWidths = row.row.map((cell, idx) =>
                        cell.width !== undefined ? cell.width : (usableWidthForTable * colWeights[idx])
                    )
                    const startWidth = currentRowWidths[colIdx]

                    // Determine which adjacent column will compensate
                    const adjacentColIdx = colIdx < numCols - 1 ? colIdx + 1 : colIdx - 1
                    const adjacentStartWidth = currentRowWidths[adjacentColIdx]
                    const minCellWidth = 30 // Minimum cell width in pixels

                    const onMouseMove = (me) => {
                        const dx = me.clientX - startX

                        // Calculate new widths ensuring minimums are respected
                        let newWidth = startWidth + dx
                        let adjacentNewWidth = adjacentStartWidth - dx

                        // Enforce minimum widths
                        if (newWidth < minCellWidth) {
                            newWidth = minCellWidth
                            adjacentNewWidth = startWidth + adjacentStartWidth - minCellWidth
                        }
                        if (adjacentNewWidth < minCellWidth) {
                            adjacentNewWidth = minCellWidth
                            newWidth = startWidth + adjacentStartWidth - minCellWidth
                        }

                        // Update only this specific row's cell widths
                        const newRows = [...element.rows]
                        newRows[rowIdx] = {
                            ...newRows[rowIdx],
                            row: newRows[rowIdx].row.map((c, idx) => {
                                if (idx === colIdx) {
                                    return { ...c, width: newWidth }
                                } else if (idx === adjacentColIdx) {
                                    return { ...c, width: adjacentNewWidth }
                                }
                                // Preserve existing width or set default for other cells in this row
                                return c.width !== undefined ? c : { ...c, width: currentRowWidths[idx] }
                            })
                        }
                        onUpdate({ rows: newRows })
                    }
                    const onMouseUp = () => {
                        setIsResizing(false)
                        window.removeEventListener('mousemove', onMouseMove)
                        window.removeEventListener('mouseup', onMouseUp)
                    }
                    window.addEventListener('mousemove', onMouseMove)
                    window.addEventListener('mouseup', onMouseUp)
                }

                // Per-cell height resize handler - updates all cells in the row
                const handleCellHeightResizeStart = (e, rowIdx, colIdx) => {
                    e.preventDefault()
                    e.stopPropagation()
                    const startY = e.clientY
                    const cell = element.rows[rowIdx].row[colIdx]
                    const startHeight = cell.height || 25

                    const onMouseMove = (me) => {
                        const dy = me.clientY - startY
                        const newHeight = Math.max(20, startHeight + dy)

                        // Update all cells in this row to have the same height
                        const newRows = [...element.rows]
                        newRows[rowIdx] = {
                            ...newRows[rowIdx],
                            row: newRows[rowIdx].row.map(c => ({ ...c, height: newHeight }))
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

                // Calculate total table width (sum of all column default widths)
                const totalTableWidth = usableWidthForTable

                return (
                    <div style={{ borderRadius: '4px', padding: '10px', overflow: 'hidden', background: 'white' }}>
                        {/* Use div-based layout for per-cell width control */
                            <div style={{ width: `${totalTableWidth}px` }}>
                                {element.rows?.map((row, rowIdx) => {
                                    // Check if any cell in this row has wrap enabled (default: true)
                                    const hasWrappedCell = row.row?.some(cell => cell.wrap !== false)

                                    return (
                                        <div key={rowIdx} style={{ display: 'flex', position: 'relative', alignItems: hasWrappedCell ? 'stretch' : 'stretch' }}>
                                            {row.row?.map((cell, colIdx) => {
                                                const cellStyle = getStyleFromProps(cell.props)
                                                const isCellSelected = selectedCell && selectedCell.elementId === element.id && selectedCell.rowIdx === rowIdx && selectedCell.colIdx === colIdx

                                                // Use individual cell width if set, otherwise use column-based width
                                                const cellWidth = cell.width !== undefined ? cell.width : (usableWidthForTable * colWeights[colIdx])
                                                const baseHeight = cell.height || 25
                                                // Wrap is enabled by default (cell.wrap !== false means wrap is on)
                                                const isWrapEnabled = cell.wrap !== false

                                                // Determine background color: use cell's or table's bg color, or default white
                                                const cellBgColor = cell.bgcolor || element.bgcolor || '#fff'

                                                // Determine text color: cell textcolor > table textcolor > default black
                                                const cellTextColor = cell.textcolor || element.textcolor || '#000'

                                                // Ensure borders are visible - use explicit border if cell has border props
                                                const hasBorder = cellStyle.borderLeftWidth !== '0px' || cellStyle.borderRightWidth !== '0px' ||
                                                    cellStyle.borderTopWidth !== '0px' || cellStyle.borderBottomWidth !== '0px'

                                                const cellContainerStyle = {
                                                    borderLeft: hasBorder ? `${cellStyle.borderLeftWidth} solid #333` : 'none',
                                                    borderRight: hasBorder ? `${cellStyle.borderRightWidth} solid #333` : 'none',
                                                    borderTop: hasBorder ? `${cellStyle.borderTopWidth} solid #333` : 'none',
                                                    borderBottom: hasBorder ? `${cellStyle.borderBottomWidth} solid #333` : 'none',
                                                    padding: '4px 8px',
                                                    width: `${cellWidth}px`,
                                                    minHeight: `${baseHeight}px`,
                                                    display: 'flex',
                                                    alignItems: isWrapEnabled ? 'flex-start' : 'center',
                                                    overflow: isWrapEnabled ? 'visible' : 'hidden',
                                                    backgroundColor: cellBgColor,
                                                    cursor: 'pointer',
                                                    position: 'relative',
                                                    boxSizing: 'border-box',
                                                    outline: isCellSelected ? '2px solid #3b82f6' : 'none',
                                                    outlineOffset: '-2px',
                                                    flexShrink: 0
                                                }
                                                const inputStyle = {
                                                    fontSize: cellStyle.fontSize,
                                                    textAlign: cellStyle.textAlign,
                                                    fontWeight: cellStyle.fontWeight,
                                                    fontStyle: cellStyle.fontStyle,
                                                    textDecoration: cellStyle.textDecoration,
                                                    width: '100%',
                                                    height: isWrapEnabled ? 'auto' : '100%',
                                                    minHeight: isWrapEnabled ? `${baseHeight - 8}px` : 'auto',
                                                    border: 'none',
                                                    background: 'transparent',
                                                    padding: '2px',
                                                    color: cellTextColor,
                                                    outline: 'none',
                                                    resize: 'none',
                                                    // Wrap-related styles
                                                    whiteSpace: isWrapEnabled ? 'pre-wrap' : 'nowrap',
                                                    wordWrap: isWrapEnabled ? 'break-word' : 'normal',
                                                    overflowWrap: isWrapEnabled ? 'break-word' : 'normal'
                                                }
                                                return (
                                                    <div
                                                        key={colIdx}
                                                        style={cellContainerStyle}
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
                                                                handleCellDrop(element, element.id, onUpdate, rowIdx, colIdx, draggedData)
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
                                                                        src={getImageSrc(cell.image.imagedata, cell.image.imagename)}
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
                                                        ) : isWrapEnabled ? (
                                                            <textarea
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
                                                                rows={Math.max(1, Math.ceil((cell.text || '').length / 20))}
                                                            />
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
                                                        {/* Cell width resize handle - show on all cells */}
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
                                                                background: isCellSelected ? 'hsl(199 89% 48% / 0.3)' : 'transparent'
                                                            }}
                                                            onMouseEnter={(e) => e.currentTarget.style.background = 'hsl(199 89% 48% / 0.5)'}
                                                            onMouseLeave={(e) => e.currentTarget.style.background = isCellSelected ? 'hsl(199 89% 48% / 0.3)' : 'transparent'}
                                                            title="Drag to resize this cell width"
                                                        />
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
                                                            onMouseEnter={(e) => e.currentTarget.style.background = 'hsl(142 71% 45% / 0.5)'}
                                                            onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                                                            title="Drag to resize cell height"
                                                        />
                                                    </div>
                                                )
                                            })}
                                        </div>
                                    )
                                })}
                            </div>
                        }
                    </div>
                )
            }
            case 'footer': {
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
            }
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
                                    src={getImageSrc(element.imagedata, element.imagename)}
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
            style={{
                position: 'relative',
                margin: '4px 0',
                padding: isSelected && element.type !== 'table' ? '8px' : '0',
                border: isSelected && element.type !== 'table' ? '2px dashed hsl(var(--ring))' : '2px solid transparent',
                borderRadius: element.type === 'table' ? '0' : '6px',
                cursor: 'pointer',
                background: isSelected && element.type !== 'table' ? 'hsl(var(--accent) / 0.15)' : 'transparent',
                boxShadow: isSelected && element.type === 'table' ? '0 0 0 2px dashed hsl(var(--ring))' : 'none',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                opacity: isDragging ? 0.4 : 1,
                transform: isDragging ? 'scale(0.98) rotate(1deg)' : 'scale(1) rotate(0deg)',
                filter: isDragging ? 'blur(1px)' : 'none'
            }}
        >
            {/* Drag Handle - Only this should be draggable */}
            {isSelected && (
                <div style={{
                    position: 'absolute',
                    left: '-50px',
                    top: '50%',
                    transform: isDragging ? 'translateY(-50%) scale(1.1)' : 'translateY(-50%) scale(1)',
                    cursor: isDragging ? 'grabbing' : 'grab',
                    padding: '6px',
                    background: isDragging ? 'hsl(var(--accent))' : 'hsl(var(--muted))',
                    borderRadius: '6px',
                    border: isDragging ? '2px solid hsl(var(--accent))' : '1px solid hsl(var(--border))',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    zIndex: 5,
                    transition: 'all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1)',
                    boxShadow: isDragging ? '0 4px 12px rgba(0, 0, 0, 0.15)' : '0 2px 4px rgba(0, 0, 0, 0.05)'
                }}
                    draggable
                    onDragStart={handleDragStart}
                    onDragEnd={handleDragEnd}
                    onDragOver={handleDragOver}
                    onDrop={handleDrop}
                    onMouseEnter={(e) => {
                        if (!isDragging) {
                            e.currentTarget.style.background = 'hsl(var(--accent) / 0.3)'
                            e.currentTarget.style.transform = 'translateY(-50%) scale(1.05)'
                        }
                    }}
                    onMouseLeave={(e) => {
                        if (!isDragging) {
                            e.currentTarget.style.background = 'hsl(var(--muted))'
                            e.currentTarget.style.transform = 'translateY(-50%) scale(1)'
                        }
                    }}
                    title="Drag to reorder"
                >
                    <GripVertical size={16} style={{ color: isDragging ? 'hsl(var(--accent-foreground))' : 'hsl(var(--foreground))', transition: 'color 0.2s ease' }} />
                </div>
            )}
            {isSelected && (
                <div style={{
                    position: 'absolute',
                    top: '-40px',
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
            {renderContent()}
        </div>
    )
}
