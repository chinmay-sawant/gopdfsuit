import { useRef, useState, useEffect } from 'react'
import { Upload, Moon, Sun, Eye, Download, Copy, Check, Edit, Github, HardDrive, FolderOpen } from 'lucide-react'

export default function Toolbar({ theme, setTheme, onLoadTemplate, onPreviewPDF, onCopyJSON, onDownloadPDF, templateInput, setTemplateInput, copiedId, elementCount = 0, pageSize = 'A4', onUploadFont }) {
    const fileInputRef = useRef(null)
    const [githubFiles, setGithubFiles] = useState([])
    const [loadMethod, setLoadMethod] = useState('github')

    useEffect(() => {
        fetch('https://api.github.com/repos/chinmay-sawant/gopdfsuit/git/trees/master?recursive=1')
            .then(res => res.json())
            .then(data => {
                if (data && data.tree) {
                    const skipFolders = ['benchmarks/', 'gopdflib/'];
                    const jsonFiles = data.tree
                        .filter(f => f.type === 'blob' && f.path.startsWith('sampledata/') && f.path.endsWith('.json'))
                        .map(f => f.path.substring('sampledata/'.length))
                        .filter(f => !skipFolders.some(skip => f.startsWith(skip)));
                    setGithubFiles(jsonFiles);
                }
            })
            .catch(err => console.error('Error fetching github files:', err))
    }, [])

    const handleFontUpload = (e) => {
        const file = e.target.files?.[0]
        if (file) {
            onUploadFont?.(file)
            e.target.value = ''
        }
    }

    const toggleBtnStyle = (active) => ({
        padding: '0.3rem 0.6rem',
        fontSize: '0.78rem',
        border: 'none',
        borderRadius: '4px',
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        gap: '0.3rem',
        fontWeight: active ? '600' : '400',
        background: active ? 'hsl(var(--primary))' : 'transparent',
        color: active ? 'hsl(var(--primary-foreground))' : 'hsl(var(--muted-foreground))',
        transition: 'all 0.15s ease'
    })

    return (
        <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '1rem'
        }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', color: 'hsl(var(--foreground))' }}>
                <Edit size={20} />
                <div>
                    <strong style={{ display: 'block', lineHeight: 1, color: 'hsl(var(--foreground))' }}>PDF Template Editor</strong>
                    <span style={{ fontSize: '0.75rem', color: 'hsl(var(--muted-foreground))' }}>{elementCount} elements • {pageSize} Portrait</span>
                </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                {/* Load Section - grouped with border */}
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.4rem',
                    padding: '0.3rem',
                    borderRadius: '8px',
                    border: '1px solid hsl(var(--border))',
                    background: 'hsl(var(--muted) / 0.5)'
                }}>
                    {/* Segmented Toggle */}
                    <div style={{
                        display: 'flex',
                        borderRadius: '6px',
                        background: 'hsl(var(--muted))',
                        padding: '2px',
                        gap: '2px'
                    }}>
                        <button
                            onClick={() => {
                                setLoadMethod('github')
                                if (loadMethod !== 'github') setTemplateInput('editor/financial_report.json')
                            }}
                            style={toggleBtnStyle(loadMethod === 'github')}
                            title="Load from GitHub repository"
                        >
                            <Github size={13} /> GitHub
                        </button>
                        <button
                            onClick={() => {
                                setLoadMethod('local')
                                if (loadMethod !== 'local') setTemplateInput('')
                            }}
                            style={toggleBtnStyle(loadMethod === 'local')}
                            title="Load from local server"
                        >
                            <HardDrive size={13} /> Local
                        </button>
                    </div>

                    {/* File Picker */}
                    {loadMethod === 'local' ? (
                        <input
                            type="text"
                            value={templateInput}
                            onChange={(e) => setTemplateInput(e.target.value)}
                            placeholder="Enter template path..."
                            style={{
                                padding: '0.35rem 0.6rem',
                                fontSize: '0.8rem',
                                minWidth: '200px',
                                borderRadius: '5px',
                                border: '1px solid hsl(var(--border))',
                                background: 'hsl(var(--background))',
                                color: 'hsl(var(--foreground))',
                                outline: 'none'
                            }}
                        />
                    ) : (
                        <select
                            value={templateInput}
                            onChange={(e) => {
                                setTemplateInput(e.target.value)
                                if (e.target.value) onLoadTemplate(e.target.value, 'github')
                            }}
                            style={{
                                padding: '0.35rem 0.6rem',
                                fontSize: '0.8rem',
                                minWidth: '220px',
                                borderRadius: '5px',
                                border: '1px solid hsl(var(--border))',
                                background: 'hsl(var(--background))',
                                color: 'hsl(var(--foreground))',
                                outline: 'none',
                                cursor: 'pointer'
                            }}
                        >
                            <option value="">Select a template...</option>
                            {githubFiles.map(f => (
                                <option key={f} value={f}>{f}</option>
                            ))}
                        </select>
                    )}

                    {/* Load Button */}
                    <button
                        onClick={() => onLoadTemplate(templateInput, loadMethod)}
                        className="btn"
                        style={{
                            padding: '0.35rem 0.7rem',
                            fontSize: '0.8rem',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.3rem',
                            borderRadius: '5px',
                            fontWeight: '500'
                        }}
                    >
                        <FolderOpen size={13} /> Load
                    </button>
                </div>

                <div style={{ width: '1px', height: '24px', background: 'hsl(var(--border))' }}></div>

                <button onClick={onPreviewPDF} className="btn primary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.3rem', background: 'var(--secondary-color)', color: 'white', borderRadius: '6px', fontWeight: '500' }}>
                    <Eye size={14} /> Preview
                </button>
                <button onClick={onDownloadPDF} className="btn" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.3rem', borderRadius: '6px' }}>
                    <Download size={14} /> Generate
                </button>
                <button
                    onClick={() => fileInputRef.current?.click()}
                    className="btn"
                    style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.3rem', borderRadius: '6px' }}
                    title="Upload custom font (.ttf or .otf)"
                >
                    <Upload size={14} /> Font
                </button>
                <input
                    ref={fileInputRef}
                    type="file"
                    accept=".ttf,.otf"
                    style={{ display: 'none' }}
                    onChange={handleFontUpload}
                />
                <button onClick={onCopyJSON} className="btn" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.3rem', borderRadius: '6px' }}>
                    {copiedId === 'json' ? <><Check size={14} /> Copied</> : <><Copy size={14} /> Copy</>}
                </button>

                <div style={{ width: '1px', height: '24px', background: 'hsl(var(--border))' }}></div>

                <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')} className="btn icon-only" title={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}>
                    {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
                </button>
            </div>
        </div>
    )
}
