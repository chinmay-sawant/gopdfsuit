
import React, { useState, useEffect } from 'react'
import { Settings, HelpCircle, PenTool, CheckSquare, Square, Lock } from 'lucide-react'
import { PAGE_SIZES } from './constants'

function PageBorderControls({ borders, onChange }) {
    const updateBorder = (index, value) => {
        const newBorders = [...borders]
        newBorders[index] = Math.max(0, Math.min(10, value))
        onChange(newBorders)
    }

    const BorderControl = ({ label, index }) => (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>{label}</label>
            <div style={{ display: 'flex', gap: '0.25rem' }}>
                <button
                    onClick={() => updateBorder(index, borders[index] - 1)}
                    disabled={borders[index] <= 0}
                    className="btn-border"
                >
                    −
                </button>
                <span style={{
                    padding: '0.25rem 0.5rem',
                    fontSize: '0.8rem',
                    minWidth: '2rem',
                    textAlign: 'center',
                    background: 'hsl(var(--muted))',
                    borderRadius: '4px'
                }}>
                    {borders[index]}px
                </span>
                <button
                    onClick={() => updateBorder(index, borders[index] + 1)}
                    disabled={borders[index] >= 10}
                    className="btn-border"
                >
                    +
                </button>
            </div>
        </div>
    )

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <h5 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', color: 'hsl(var(--foreground))' }}>Page Borders</h5>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
                <BorderControl label="Left" index={0} />
                <BorderControl label="Right" index={1} />
                <BorderControl label="Top" index={2} />
                <BorderControl label="Bottom" index={3} />
            </div>

            {/* Quick Border Presets */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <label style={{ fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--muted-foreground))' }}>Quick Set</label>
                <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
                    {[
                        { label: 'None', borders: [0, 0, 0, 0] },
                        { label: 'All', borders: [1, 1, 1, 1] },
                        { label: 'Box', borders: [1, 1, 1, 1] },
                        { label: 'Bottom', borders: [0, 0, 1, 0] }
                    ].map(({ label, borders: presetBorders }) => (
                        <button
                            key={label}
                            onClick={() => onChange(presetBorders)}
                            style={{
                                padding: '0.25rem 0.5rem',
                                border: '1px solid hsl(var(--border))',
                                borderRadius: '4px',
                                background: 'hsl(var(--muted))',
                                color: 'hsl(var(--muted-foreground))',
                                fontSize: '0.8rem',
                                cursor: 'pointer',
                                transition: 'all 0.2s ease'
                            }}
                        >
                            {label}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    )
}

function SignatureSettings({ config, onChange }) {
    const handleChange = (key, value) => {
        onChange({ ...config, [key]: value })
    }

    // Handle number inputs specifically
    const handleNumberChange = (key, value) => {
        onChange({ ...config, [key]: parseInt(value) || 0 })
    }

    if (!config.enabled) {
        return (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <h5 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'hsl(var(--foreground))' }}>
                    <PenTool size={14} /> Digital Signature
                </h5>
                <button
                    onClick={() => onChange({ ...config, enabled: true, visible: true, page: 1, width: 200, height: 50, x: 0, y: 0 })}
                    className="btn"
                    style={{ width: '100%', fontSize: '0.85rem', padding: '0.5rem', background: 'hsl(var(--primary))', color: 'hsl(var(--primary-foreground))', borderRadius: '4px', border: 'none', cursor: 'pointer' }}
                >
                    Enable Signature
                </button>
            </div>
        )
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <h5 style={{ fontSize: '0.9rem', fontWeight: '600', margin: '0', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'hsl(var(--foreground))' }}>
                    <PenTool size={14} /> Digital Signature
                </h5>
                <button
                    onClick={() => onChange({ ...config, enabled: false })}
                    style={{
                        background: 'transparent',
                        border: 'none',
                        color: 'hsl(var(--destructive))',
                        cursor: 'pointer',
                        fontSize: '0.8rem'
                    }}
                >
                    Disable
                </button>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Name</label>
                    <input
                        type="text"
                        value={config.name || ''}
                        onChange={(e) => handleChange('name', e.target.value)}
                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                </div>
                <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Reason</label>
                    <input
                        type="text"
                        value={config.reason || ''}
                        onChange={(e) => handleChange('reason', e.target.value)}
                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Location</label>
                    <input
                        type="text"
                        value={config.location || ''}
                        onChange={(e) => handleChange('location', e.target.value)}
                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                </div>
                <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Contact Info</label>
                    <input
                        type="text"
                        value={config.contactInfo || ''}
                        onChange={(e) => handleChange('contactInfo', e.target.value)}
                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <button
                    onClick={() => handleChange('visible', !config.visible)}
                    style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, display: 'flex', alignItems: 'center' }}
                >
                    {config.visible ? <CheckSquare size={16} /> : <Square size={16} />}
                </button>
                <label style={{ fontSize: '0.85rem', fontWeight: '500' }}>Visible Signature</label>
            </div>

            {config.visible && (
                <>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Page</label>
                            <input
                                type="number"
                                value={config.page || 1}
                                onChange={(e) => handleNumberChange('page', e.target.value)}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Width</label>
                            <input
                                type="number"
                                value={config.width || 200}
                                onChange={(e) => handleNumberChange('width', e.target.value)}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                    </div>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.5rem' }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>X</label>
                            <input
                                type="number"
                                value={config.x || 0}
                                onChange={(e) => handleNumberChange('x', e.target.value)}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Y</label>
                            <input
                                type="number"
                                value={config.y || 0}
                                onChange={(e) => handleNumberChange('y', e.target.value)}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Height</label>
                            <input
                                type="number"
                                value={config.height || 50}
                                onChange={(e) => handleNumberChange('height', e.target.value)}
                                style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                            />
                        </div>
                    </div>
                </>
            )}

            <div>
                <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Certificate (PEM)</label>
                <textarea
                    value={config.certPEM || ''}
                    onChange={(e) => handleChange('certPEM', e.target.value)}
                    placeholder="-----BEGIN CERTIFICATE-----..."
                    rows={3}
                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.75rem', fontFamily: 'monospace', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', resize: 'vertical' }}
                />
            </div>

            <div>
                <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Private Key (PEM)</label>
                <textarea
                    value={config.keyPEM || ''}
                    onChange={(e) => handleChange('keyPEM', e.target.value)}
                    placeholder="-----BEGIN PRIVATE KEY-----..."
                    rows={3}
                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.75rem', fontFamily: 'monospace', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', resize: 'vertical' }}
                />
            </div>

            <div>
                <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Intermediate Certificates (Optional)</label>
                <textarea
                    value={config.chainPEM || ''}
                    onChange={(e) => handleChange('chainPEM', e.target.value)}
                    placeholder="Paste intermediate certificates here..."
                    rows={3}
                    style={{ width: '100%', padding: '0.4rem', fontSize: '0.75rem', fontFamily: 'monospace', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))', resize: 'vertical' }}
                />
            </div>

        </div>
    )
}

// Helper for page borders format: "L:R:T:B"
const parsePageBorder = (str) => {
    if (!str) return [0, 0, 0, 0]
    return str.split(':').map(Number)
}

export default function DocumentSettings({ config, setConfig, currentPageSize }) {
    const [showPdfTooltip, setShowPdfTooltip] = useState(false)

    return (
        <div className="card" style={{ padding: '1rem', flexShrink: 0 }}>
            <h3 style={{
                margin: '0 0 0.75rem 0',
                fontSize: '0.9rem',
                fontWeight: '600',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                color: 'hsl(var(--foreground))'
            }}>
                <Settings size={16} /> Document Settings
            </h3>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                {/* Page Size & Orientation Row */}
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                    <div>
                        <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Page Size</label>
                        <select
                            value={config.page}
                            onChange={(e) => setConfig(prev => ({ ...prev, page: e.target.value }))}
                            style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                        >
                            {Object.entries(PAGE_SIZES).map(([key, size]) => (
                                <option key={key} value={key}>{size.name}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Orientation</label>
                        <select
                            value={config.pageAlignment}
                            onChange={(e) => setConfig(prev => ({ ...prev, pageAlignment: parseInt(e.target.value) }))}
                            style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                        >
                            <option value={1}>Portrait</option>
                            <option value={2}>Landscape</option>
                        </select>
                    </div>
                </div>

                {/* Watermark */}
                <div>
                    <label style={{ display: 'block', fontSize: '0.75rem', marginBottom: '0.25rem', color: 'hsl(var(--muted-foreground))' }}>Watermark</label>
                    <input
                        type="text"
                        value={config.watermark || ''}
                        onChange={(e) => setConfig(prev => ({ ...prev, watermark: e.target.value }))}
                        placeholder="Optional watermark text"
                        style={{ width: '100%', padding: '0.4rem', fontSize: '0.85rem', border: '1px solid hsl(var(--border))', borderRadius: '4px', background: 'hsl(var(--background))', color: 'hsl(var(--foreground))' }}
                    />
                </div>

                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', background: 'hsl(var(--muted))', borderRadius: '8px', position: 'relative' }}>
                    {showPdfTooltip && (
                        <div style={{ position: 'absolute', top: '-65px', left: '50%', transform: 'translateX(-50%)', background: 'black', color: 'white', padding: '8px', borderRadius: '6px', fontSize: '0.75rem', width: '200px', textAlign: 'center', zIndex: 100, pointerEvents: 'none' }}>
                            If the file is encrypted, it violates PDF/A compliance.
                            <div style={{ position: 'absolute', bottom: '-4px', left: '50%', transform: 'translateX(-50%) rotate(45deg)', width: '8px', height: '8px', background: 'black' }} />
                        </div>
                    )}
                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                            <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--foreground))' }}>PDF/A Compliant</label>
                            <HelpCircle size={14} onMouseEnter={() => setShowPdfTooltip(true)} onMouseLeave={() => setShowPdfTooltip(false)} style={{ cursor: 'help', color: 'hsl(var(--muted-foreground))' }} />
                        </div>
                        <span style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))' }}>PDF/A-4 Standard</span>
                    </div>
                    <label style={{
                        position: 'relative',
                        display: 'inline-block',
                        width: '52px',
                        height: '28px',
                        cursor: 'pointer'
                    }}>
                        <input
                            type="checkbox"
                            checked={config.pdfaCompliant !== false}
                            onChange={(e) => {
                                const isEnabled = e.target.checked
                                if (isEnabled) {
                                    setConfig(prev => ({ ...prev, pdfaCompliant: true, security: { ...(prev.security || {}), enabled: false } }))
                                } else {
                                    setConfig(prev => ({ ...prev, pdfaCompliant: false }))
                                }
                            }}
                            style={{ opacity: 0, width: 0, height: 0, position: 'absolute' }}
                        />
                        <span style={{
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            background: config.pdfaCompliant !== false ? '#4ecdc4' : 'hsl(var(--border))',
                            borderRadius: '28px',
                            transition: '0.3s',
                            cursor: 'pointer'
                        }}>
                            <span style={{
                                position: 'absolute',
                                content: '',
                                height: '20px',
                                width: '20px',
                                left: config.pdfaCompliant !== false ? '28px' : '4px',
                                bottom: '4px',
                                background: 'white',
                                borderRadius: '50%',
                                transition: '0.3s',
                                boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                            }} />
                        </span>
                    </label>
                </div>

                {/* Arlington Compatible */}
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', background: 'hsl(var(--muted))', borderRadius: '8px' }}>
                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                        <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--foreground))' }}>Arlington Compatible</label>
                        <span style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))' }}>PDF 2.0 compliant fonts</span>
                    </div>
                    <label style={{
                        position: 'relative',
                        display: 'inline-block',
                        width: '52px',
                        height: '28px',
                        cursor: 'pointer'
                    }}>
                        <input
                            type="checkbox"
                            checked={config.arlingtonCompatible || false}
                            onChange={(e) => setConfig(prev => ({ ...prev, arlingtonCompatible: e.target.checked }))}
                            style={{ opacity: 0, width: 0, height: 0, position: 'absolute' }}
                        />
                        <span style={{
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            background: config.arlingtonCompatible ? '#4ecdc4' : 'hsl(var(--border))',
                            borderRadius: '28px',
                            transition: '0.3s',
                            cursor: 'pointer'
                        }}>
                            <span style={{
                                position: 'absolute',
                                content: '',
                                height: '20px',
                                width: '20px',
                                left: config.arlingtonCompatible ? '28px' : '4px',
                                bottom: '4px',
                                background: 'white',
                                borderRadius: '50%',
                                transition: '0.3s',
                                boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                            }} />
                        </span>
                    </label>
                </div>

                {/* Embed Standard Fonts */}
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', background: 'hsl(var(--muted))', borderRadius: '8px' }}>
                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                        <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: '500', color: 'hsl(var(--foreground))' }}>Embed Standard Fonts</label>
                        <span style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))' }}>Embed used standard fonts</span>
                    </div>
                    <label style={{
                        position: 'relative',
                        display: 'inline-block',
                        width: '52px',
                        height: '28px',
                        cursor: 'pointer'
                    }}>
                        <input
                            type="checkbox"
                            checked={config.embedStandardFonts || false}
                            onChange={(e) => setConfig(prev => ({ ...prev, embedStandardFonts: e.target.checked }))}
                            style={{ opacity: 0, width: 0, height: 0, position: 'absolute' }}
                        />
                        <span style={{
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            background: config.embedStandardFonts ? '#4ecdc4' : 'hsl(var(--border))',
                            borderRadius: '28px',
                            transition: '0.3s',
                            cursor: 'pointer'
                        }}>
                            <span style={{
                                position: 'absolute',
                                content: '',
                                height: '20px',
                                width: '20px',
                                left: config.embedStandardFonts ? '28px' : '4px',
                                bottom: '4px',
                                background: 'white',
                                borderRadius: '50%',
                                transition: '0.3s',
                                boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                            }} />
                        </span>
                    </label>
                </div>

                {/* PDF Security Card */}
                <div style={{
                    background: config.security?.enabled ? 'hsl(217.2 32.6% 17.5%)' : 'hsl(var(--muted))',
                    borderRadius: '8px',
                    padding: '0.75rem',
                    border: '1px solid hsl(var(--border))',
                    transition: 'all 0.3s ease'
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: config.security?.enabled ? '0.75rem' : '0' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <Lock size={16} style={{ color: config.security?.enabled ? '#4ecdc4' : 'hsl(var(--foreground))' }} />
                            <div>
                                <div style={{ fontSize: '0.85rem', fontWeight: '600', color: config.security?.enabled ? '#fff' : 'hsl(var(--foreground))' }}>
                                    PDF Security
                                </div>
                                {config.security?.enabled && (
                                    <div style={{ fontSize: '0.7rem', color: 'hsl(var(--muted-foreground))', marginTop: '1px' }}>
                                        Encryption & permissions
                                    </div>
                                )}
                            </div>
                        </div>
                        <label style={{
                            position: 'relative',
                            display: 'inline-block',
                            width: '52px',
                            height: '28px',
                            cursor: 'pointer'
                        }}>
                            <input
                                type="checkbox"
                                checked={config.security?.enabled || false}
                                onChange={(e) => {
                                    const isEnabled = e.target.checked
                                    if (isEnabled) {
                                        setConfig(prev => ({
                                            ...prev,
                                            security: {
                                                enabled: true,
                                                ownerPassword: '',
                                                userPassword: '',
                                                allowPrinting: true,
                                                allowCopying: true,
                                                allowModifying: true,
                                                allowAnnotations: true,
                                                allowFormFilling: true,
                                                allowAccessibility: true
                                            },
                                            pdfaCompliant: false
                                        }))
                                    } else {
                                        setConfig(prev => ({ ...prev, security: { ...prev.security, enabled: false } }))
                                    }
                                }}
                                style={{ opacity: 0, width: 0, height: 0, position: 'absolute' }}
                            />
                            <span style={{
                                position: 'absolute',
                                top: 0,
                                left: 0,
                                right: 0,
                                bottom: 0,
                                background: config.security?.enabled ? '#4ecdc4' : 'hsl(var(--border))',
                                borderRadius: '28px',
                                transition: '0.3s',
                                cursor: 'pointer'
                            }}>
                                <span style={{
                                    position: 'absolute',
                                    content: '',
                                    height: '20px',
                                    width: '20px',
                                    left: config.security?.enabled ? '28px' : '4px',
                                    bottom: '4px',
                                    background: 'white',
                                    borderRadius: '50%',
                                    transition: '0.3s',
                                    boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                                }} />
                            </span>
                        </label>
                    </div>

                    {config.security?.enabled && (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                            {/* Owner Password */}
                            <div>
                                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: '600', color: '#fff', marginBottom: '0.35rem' }}>
                                    Owner Password <span style={{ color: '#ff5f56' }}>*</span>
                                </label>
                                <input
                                    type="password"
                                    value={config.security?.ownerPassword || ''}
                                    onChange={(e) => setConfig(prev => ({ ...prev, security: { ...prev.security, ownerPassword: e.target.value } }))}
                                    placeholder="Full access password"
                                    style={{
                                        width: '100%',
                                        padding: '0.5rem',
                                        fontSize: '0.8rem',
                                        border: '1px solid rgba(255,255,255,0.1)',
                                        borderRadius: '4px',
                                        background: 'rgba(0,0,0,0.3)',
                                        color: '#fff',
                                        outline: 'none'
                                    }}
                                />
                            </div>

                            {/* User Password */}
                            <div>
                                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: '600', color: '#fff', marginBottom: '0.35rem' }}>
                                    User Password (Optional)
                                </label>
                                <input
                                    type="password"
                                    value={config.security?.userPassword || ''}
                                    onChange={(e) => setConfig(prev => ({ ...prev, security: { ...prev.security, userPassword: e.target.value } }))}
                                    placeholder="To open PDF (leave empty for none)"
                                    style={{
                                        width: '100%',
                                        padding: '0.5rem',
                                        fontSize: '0.8rem',
                                        border: '1px solid rgba(255,255,255,0.1)',
                                        borderRadius: '4px',
                                        background: 'rgba(0,0,0,0.3)',
                                        color: '#fff',
                                        outline: 'none'
                                    }}
                                />
                            </div>

                            {/* Permissions */}
                            <div>
                                <h5 style={{ fontSize: '0.85rem', fontWeight: '600', color: '#fff', marginBottom: '0.5rem' }}>Permissions</h5>
                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                    {[
                                        { key: 'allowPrinting', label: 'Printing' },
                                        { key: 'allowCopying', label: 'Copying' },
                                        { key: 'allowModifying', label: 'Modifying' },
                                        { key: 'allowAnnotations', label: 'Annotations' },
                                        { key: 'allowFormFilling', label: 'Form Filling' },
                                        { key: 'allowAccessibility', label: 'Accessibility' }
                                    ].map(({ key, label }) => (
                                        <label
                                            key={key}
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: '0.4rem',
                                                cursor: 'pointer',
                                                padding: '0.35rem',
                                                borderRadius: '4px',
                                                transition: 'background 0.2s',
                                                background: 'transparent'
                                            }}
                                            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(255,255,255,0.05)'}
                                            onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                                        >
                                            <input
                                                type="checkbox"
                                                checked={config.security?.[key] !== false}
                                                onChange={(e) => setConfig(prev => ({ ...prev, security: { ...prev.security, [key]: e.target.checked } }))}
                                                style={{
                                                    width: '16px',
                                                    height: '16px',
                                                    cursor: 'pointer',
                                                    accentColor: '#4ecdc4'
                                                }}
                                            />
                                            <span style={{ fontSize: '0.8rem', color: '#fff', fontWeight: '500' }}>{label}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                <hr style={{ border: 'none', borderTop: '1px solid hsl(var(--border))', margin: '0' }} />

                {/* Page Borders */}
                <PageBorderControls
                    borders={parsePageBorder(config.pageBorder)}
                    onChange={(newBorders) => setConfig(prev => ({ ...prev, pageBorder: newBorders.join(':') }))}
                />

                <hr style={{ border: 'none', borderTop: '1px solid hsl(var(--border))', margin: '0' }} />

                {/* Signature Settings */}
                <SignatureSettings
                    config={config.signature || { enabled: false }}
                    onChange={(newSig) => setConfig(prev => ({ ...prev, signature: newSig }))}
                />

            </div>

            <style jsx>{`
        .btn-border { padding: 0.25rem 0.5rem; border: 1px solid hsl(var(--border)); background: hsl(var(--background)); border-radius: 4px; cursor: pointer; }
        .btn-border:disabled { opacity: 0.5; cursor: not-allowed; }
      `}</style>
        </div>
    )
}
