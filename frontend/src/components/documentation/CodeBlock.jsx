import { useState } from 'react'
import { Copy, Check } from 'lucide-react'

export const CodeBlock = ({ code }) => {
    const [copied, setCopied] = useState(false)

    const languages = code && typeof code === 'object' ? Object.keys(code) : ['shell']
    const [language, setLanguage] = useState(languages[0])

    // If no code is provided, don't render anything
    if (!code) return null;

    const codeContent = typeof code === 'object' ? code[language] : code

    const handleCopy = () => {
        navigator.clipboard.writeText(codeContent)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    return (
        <div style={{
            background: '#1e293b',
            borderRadius: '8px',
            overflow: 'hidden',
            border: '1px solid #334155',
            marginTop: '1rem',
            boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)'
        }}>
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '0.5rem 1rem',
                borderBottom: '1px solid #334155',
                background: '#0f172a'
            }}>
                <div style={{ display: 'flex', gap: '1rem' }}>
                    {languages.map(lang => (
                        <button
                            key={lang}
                            onClick={() => setLanguage(lang)}
                            style={{
                                background: 'transparent',
                                border: 'none',
                                color: language === lang ? '#38bdf8' : '#94a3b8',
                                cursor: 'pointer',
                                fontSize: '0.75rem',
                                fontWeight: '600',
                                textTransform: 'uppercase',
                                padding: '0.25rem 0.5rem',
                                borderRadius: '4px',
                                transition: 'all 0.2s',
                                backgroundColor: language === lang ? 'rgba(56, 189, 248, 0.1)' : 'transparent'
                            }}
                        >
                            {lang}
                        </button>
                    ))}
                </div>
                <button
                    onClick={handleCopy}
                    style={{
                        background: 'transparent',
                        border: 'none',
                        cursor: 'pointer',
                        color: '#94a3b8',
                        padding: '4px',
                        borderRadius: '4px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center'
                    }}
                    title="Copy to clipboard"
                >
                    {copied ? <Check size={16} color="#4ade80" /> : <Copy size={16} />}
                </button>
            </div>
            <div style={{ position: 'relative' }}>
                <pre style={{
                    padding: '1rem',
                    margin: 0,
                    overflowX: 'auto',
                    color: '#e2e8f0',
                    fontSize: '0.85rem',
                    fontFamily: "'Fira Code', 'Roboto Mono', monospace",
                    lineHeight: '1.6',
                    whiteSpace: 'pre'
                }}>
                    {codeContent}
                </pre>
            </div>
        </div>
    )
}
