import { useState, useEffect } from 'react'
import { Check, Copy } from 'lucide-react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus, vs } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { useTheme } from '../../theme'
import './documentation.css'

export const CodeBlock = ({ code }) => {
    const { theme } = useTheme()
    const [copied, setCopied] = useState(false)

    // Determine available languages
    const languages = code && typeof code === 'object' ? Object.keys(code) : ['shell']

    // State for selected language
    const [language, setLanguage] = useState(languages[0])

    // Effect to update selected language if the current one is not available in the new item
    useEffect(() => {
        if (code && typeof code === 'object') {
            const availableLanguages = Object.keys(code)
            if (!availableLanguages.includes(language)) {
                setLanguage(availableLanguages[0])
            }
        }
    }, [code, language])

    const codeContent = typeof code === 'object' ? code[language] : code

    const handleCopy = () => {
        if (!codeContent) return
        navigator.clipboard.writeText(codeContent)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    // Map display names to Prism supported languages
    const getSyntaxLanguage = (langKey) => {
        const map = {
            'node': 'javascript',
            'curl': 'bash',
            'shell': 'bash',
            'html': 'markup',
            'vue': 'javascript',
            'react': 'jsx',
            'python': 'python'
        }
        return map[langKey] || langKey
    }

    if (!codeContent) return null

    const isLight = theme === 'light';

    return (
        <div className="code-block-wrapper" style={{
            background: isLight ? '#f8fafc' : '#1e293b',
            border: isLight ? '1px solid #cbd5e1' : '1px solid #334155'
        }}>
            <div className="code-block-header" style={{
                background: isLight ? '#f1f5f9' : '#0f172a',
                borderBottom: isLight ? '1px solid #e2e8f0' : '1px solid #334155'
            }}>
                <div className="code-language-tabs">
                    {languages.map(lang => (
                        <button
                            key={lang}
                            onClick={() => setLanguage(lang)}
                            className={`code-tab ${language === lang ? 'active' : ''} ${isLight ? 'light' : ''}`}
                        >
                            {lang === 'curl' ? 'cURL' : lang.charAt(0).toUpperCase() + lang.slice(1)}
                        </button>
                    ))}
                </div>
                <button
                    onClick={handleCopy}
                    className={`copy-button ${isLight ? 'light' : ''}`}
                    title="Copy to clipboard"
                >
                    {copied ? <Check size={16} color="#4ade80" /> : <Copy size={16} />}
                </button>
            </div>

            <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
                <SyntaxHighlighter
                    language={getSyntaxLanguage(language)}
                    style={isLight ? vs : vscDarkPlus}
                    customStyle={{
                        margin: 0,
                        padding: '1.5rem',
                        background: isLight ? '#ffffff' : '#1e293b',
                        height: '100%',
                        fontSize: '0.9rem',
                        lineHeight: '1.5'
                    }}
                    className="syntax-highlighter-scrollable"
                    showLineNumbers={false}
                    wrapLines={false}
                >
                    {codeContent}
                </SyntaxHighlighter>
            </div>
        </div>
    )
}
