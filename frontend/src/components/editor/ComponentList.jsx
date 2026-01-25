
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
                alignItems: 'center',
                justifyContent: 'flex-start',
                gap: '0.5rem',
                padding: '0.65rem 0.75rem',
                background: 'hsl(var(--secondary))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
                cursor: 'grab',
                userSelect: 'none',
                transition: 'all 0.15s ease',
                opacity: isDragging === type ? 0.5 : 1,
                minHeight: '42px',
                color: 'hsl(var(--foreground))',
                fontSize: '0.825rem',
                fontWeight: '500'
            }}
            onMouseEnter={(e) => {
                if (isDragging !== type) {
                    e.currentTarget.style.background = 'hsl(var(--accent))'
                    e.currentTarget.style.transform = 'translateY(-1px)'
                    e.currentTarget.style.boxShadow = '0 2px 4px rgba(0,0,0,0.1)'
                }
            }}
            onMouseLeave={(e) => {
                e.currentTarget.style.background = 'hsl(var(--secondary))'
                e.currentTarget.style.transform = 'translateY(0)'
                e.currentTarget.style.boxShadow = 'none'
            }}
        >
            <IconComponent size={16} style={{ opacity: 0.9 }} />
            <span style={{ fontWeight: '500' }}>{componentData.label}</span>
        </div>
    )
}

export default function ComponentList({ draggedType, setDraggedType }) {
    return (
        <div style={{ 
            flexShrink: 0, 
            background: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '8px',
            padding: '1rem'
        }}>
            <h3 style={{
                margin: '0 0 0.75rem 0',
                fontSize: '0.9rem',
                fontWeight: '600',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                color: 'hsl(var(--foreground))'
            }}>
                <div style={{ 
                    border: '2px solid hsl(var(--foreground))', 
                    width: '14px', 
                    height: '14px', 
                    borderRadius: '2px'
                }}></div>
                Components
            </h3>
            <div style={{
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: '0.5rem'
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
