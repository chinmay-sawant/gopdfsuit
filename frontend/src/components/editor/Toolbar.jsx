
import React, { useRef } from 'react'
import { Upload, Moon, Sun, Eye, Download, Copy, Check, Edit } from 'lucide-react'

export default function Toolbar({ theme, setTheme, onLoadTemplate, onPreviewPDF, onCopyJSON, onDownloadPDF, templateInput, setTemplateInput, copiedId, elementCount = 0, pageSize = 'A4', onUploadFont }) {
    const fileInputRef = useRef(null)

    const handleFontUpload = (e) => {
        const file = e.target.files?.[0]
        if (file) {
            onUploadFont?.(file)
            // Reset input so same file can be uploaded again if needed
            e.target.value = ''
        }
    }

    return (
        <div className="card" style={{
            marginBottom: '1rem',
            padding: '0.75rem 1rem',
            position: 'sticky',
            top: '74px',
            zIndex: 40,
            borderRadius: '0',
            borderLeft: 'none',
            borderRight: 'none',
            marginLeft: '-1rem',
            marginRight: '-1rem',
            width: 'calc(100% + 2rem)',
            boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
            background: 'hsl(var(--card))'
        }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <Edit size={20} />
                    <div>
                        <strong style={{ display: 'block', lineHeight: 1 }}>PDF Template Editor</strong>
                        <span style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>{elementCount} elements • {pageSize} Portrait</span>
                    </div>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                    <input
                        type="text"
                        value={templateInput}
                        onChange={(e) => setTemplateInput(e.target.value)}
                        placeholder="Load template file..."
                        style={{ padding: '0.4rem 0.6rem', fontSize: '0.9rem', minWidth: '200px', borderRadius: '4px', border: '1px solid hsl(var(--border))', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                    <button onClick={() => onLoadTemplate(templateInput)} className="btn" style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                        <Upload size={14} /> Load
                    </button>
                    <button onClick={onPreviewPDF} className="btn primary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.25rem', background: 'var(--secondary-color)', color: 'white' }}>
                        <Eye size={14} /> Preview
                    </button>
                    <button onClick={onDownloadPDF} className="btn" style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                        <Download size={14} /> Generate
                    </button>
                    <button 
                        onClick={() => fileInputRef.current?.click()} 
                        className="btn" 
                        style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}
                        title="Upload custom font (.ttf or .otf)"
                    >
                        <Upload size={14} /> Upload Font
                    </button>
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept=".ttf,.otf"
                        style={{ display: 'none' }}
                        onChange={handleFontUpload}
                    />
                    <button onClick={onCopyJSON} className="btn" style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                        {copiedId === 'json' ? <><Check size={14} /> Copied</> : <><Copy size={14} /> Copy</>}
                    </button>

                    <div style={{ width: '1px', height: '20px', background: 'hsl(var(--border))', margin: '0 0.5rem' }}></div>

                    <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')} className="btn icon-only" title={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}>
                        {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
                    </button>
                </div>
            </div>
        </div>
    )
}
