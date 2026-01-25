
import React from 'react'
import { FileText, Check, Copy } from 'lucide-react'

export default function JsonTemplate({ jsonText, handleJsonChange, setIsJsonEditing, handleJsonBlur, copiedId, setCopiedId }) {
    return (
        <div style={{ 
            padding: '1rem', 
            flex: 1,
            background: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '8px'
        }}>
            <div style={{ marginBottom: '0.75rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <h3 style={{ margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.9rem', color: 'hsl(var(--foreground))' }}>
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
                    style={{ 
                        padding: '0.25rem 0.5rem', 
                        fontSize: '0.75rem', 
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: '0.25rem',
                        background: 'hsl(var(--secondary))',
                        color: 'hsl(var(--foreground))',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        transition: 'all 0.2s ease'
                    }}
                    onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'hsl(var(--accent))'
                    }}
                    onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'hsl(var(--secondary))'
                    }}
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
                    background: 'hsl(var(--muted))',
                    color: 'hsl(var(--foreground))',
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
    )
}
