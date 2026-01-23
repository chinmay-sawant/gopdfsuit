
import React, { useState } from 'react'
import { Plus, GripVertical } from 'lucide-react'
import { COMPONENT_TYPES } from './constants'

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
                alignItems: 'center', // Align horizontally
                justifyContent: 'flex-start',
                gap: '0.5rem',
                padding: '0.6rem 0.8rem', // Reduced padding
                background: 'hsl(var(--card))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px', // Smaller radius
                cursor: 'grab',
                userSelect: 'none',
                transition: 'all 0.1s ease',
                opacity: isDragging === type ? 0.5 : 1,
                height: '42px', // Fixed smaller height
                color: 'hsl(var(--foreground))',
                fontSize: '0.85rem' // Smaller font
            }}
        >
            <IconComponent size={16} style={{ opacity: 0.9 }} />
            <span style={{ fontWeight: '500' }}>{componentData.label}</span>
        </div>
    )
}

export default function ComponentList({ draggedType, setDraggedType }) {
    return (
        <div style={{ flexShrink: 0 }}>
            <h3 style={{
                margin: '0 0 0.5rem 0',
                fontSize: '0.85rem',
                fontWeight: '600',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                color: 'hsl(var(--foreground))'
            }}>
                <div style={{ border: '1px solid hsl(var(--foreground))', width: '12px', height: '12px', marginRight: '4px' }}></div>
                Components
            </h3>
            <div style={{
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: '0.5rem' // Reduced gap
            }}>
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
    )
}
