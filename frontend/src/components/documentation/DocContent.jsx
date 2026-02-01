import { CodeBlock } from './CodeBlock'
import { useTheme } from '../../theme'
import * as Icons from 'lucide-react'

// Helper to simple render markdown-like tables and text
const parseInlineStyles = (text) => {
    const parts = [];
    const boldSplit = text.split(/(\*\*[^*]+\*\*)/g);

    boldSplit.forEach((segment, i) => {
        if (segment.startsWith('**') && segment.endsWith('**')) {
            parts.push(<strong key={i} className="doc-strong">{segment.slice(2, -2)}</strong>);
        } else {
            const linkSplit = segment.split(/(\[[^\]]+\]\([^)]+\))/g);
            linkSplit.forEach((subSegment, j) => {
                const linkMatch = subSegment.match(/\[([^\]]+)\]\(([^)]+)\)/);
                if (linkMatch) {
                    parts.push(
                        <a
                            key={`${i}-${j}`}
                            href={linkMatch[2]}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="doc-link"
                        >
                            {linkMatch[1]}
                        </a>
                    );
                } else if (subSegment) {
                    parts.push(subSegment);
                }
            });
        }
    });
    return parts;
};

// Helper to simple render markdown-like tables and text
const renderMarkdownContent = (content) => {
    if (!content) return null;

    const lines = content.split('\n');
    const elements = [];
    let tableBuffer = [];
    let textBuffer = [];
    let listBuffer = [];
    let inTable = false;
    let inList = false;

    const flushTable = (key) => {
        if (tableBuffer.length > 0) {
            const headerRow = tableBuffer[0].replace(/^\||\|$/g, '').split('|').map(c => c.trim());
            const bodyRows = tableBuffer.slice(2).map(row =>
                row.replace(/^\||\|$/g, '').split('|').map(c => c.trim())
            );

            elements.push(
                <div key={key} style={{ overflowX: 'auto', marginBottom: '1.5rem' }}>
                    <table className="doc-table compact">
                        <thead>
                            <tr>
                                {headerRow.map((h, i) => <th key={i}>{parseInlineStyles(h)}</th>)}
                            </tr>
                        </thead>
                        <tbody>
                            {bodyRows.map((row, r_i) => (
                                <tr key={r_i}>
                                    {row.map((cell, c_i) => <td key={c_i}>{parseInlineStyles(cell)}</td>)}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            );
            tableBuffer = [];
        }
    }

    const flushText = (key) => {
        if (textBuffer.length > 0) {
            const text = textBuffer.join('\n');
            elements.push(
                <div key={key} className="doc-content-text">
                    {parseInlineStyles(text)}
                </div>
            );
            textBuffer = [];
        }
    }

    const flushList = (key) => {
        if (listBuffer.length > 0) {
            elements.push(
                <ul key={key} className="doc-list">
                    {listBuffer.map((item, i) => (
                        <li key={i}>{parseInlineStyles(item)}</li>
                    ))}
                </ul>
            );
            listBuffer = [];
        }
    }

    lines.forEach((line, index) => {
        const trimmed = line.trim();

        // Handle Tables
        if (trimmed.startsWith('|')) {
            if (!inTable) {
                flushText(`text-before-table-${index}`);
                flushList(`list-before-table-${index}`);
                inTable = true;
            }
            tableBuffer.push(trimmed);
            return; // Skip other checks
        } else if (inTable) {
            flushTable(`table-${index}`);
            inTable = false;
        }

        // Handle Lists
        if (trimmed.match(/^[•-]\s/)) {
            if (!inList) {
                flushText(`text-before-list-${index}`);
                // No need to flush table as it's handled above
                inList = true;
            }
            // Remove bullet point
            listBuffer.push(trimmed.replace(/^[•-]\s/, ''));
        } else {
            if (inList) {
                flushList(`list-${index}`);
                inList = false;
            }
            if (trimmed) {
                textBuffer.push(line);
            } else if (textBuffer.length > 0) {
                // If empty line, flush text to create paragraph separation if desired, 
                // or just append empty line to buffer to preserve spacing.
                // For better styling, better to flush paragraphs.
                flushText(`text-para-${index}`);
            }
        }
    });

    if (inTable) flushTable('table-end');
    if (inList) flushList('list-end');
    flushText('text-end'); // Flush remaining text

    return <div className="doc-content">{elements}</div>;
};

export const DocContent = ({ item }) => {
    const { theme } = useTheme()
    const isLight = theme === 'light'

    if (!item) return (
        <div style={{
            flex: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'hsl(var(--muted-foreground))',
            flexDirection: 'column',
            gap: '1rem'
        }}>
            <div style={{ fontSize: '3rem' }}>📚</div>
            <div>Select an item from the sidebar to view documentation</div>
        </div>
    )

    return (
        <div style={{ display: 'flex', width: '100%', height: '100%', overflow: 'hidden' }}>
            {/* Content Column */}
            <div style={{
                flex: 1,
                padding: '3rem 4rem',
                overflowY: 'auto',
                maxWidth: '900px'
            }}>
                <div style={{ marginBottom: '2rem' }}>
                    <h1 style={{ fontSize: '2.5rem', fontWeight: '800', marginBottom: '1rem', lineHeight: 1.2, letterSpacing: '-0.02em' }}>{item.title}</h1>

                    {item.method && (
                        <div style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.75rem',
                            marginBottom: '2rem',
                            background: 'hsl(var(--muted) / 0.5)',
                            padding: '0.75rem 1rem',
                            borderRadius: '8px',
                            fontFamily: 'monospace',
                            fontSize: '0.9rem',
                            border: '1px solid hsl(var(--border))'
                        }}>
                            <span style={{
                                fontWeight: 'bold',
                                color: 'white',
                                background: getMethodColor(item.method),
                                padding: '4px 10px',
                                borderRadius: '4px',
                                fontSize: '0.75rem',
                                boxShadow: '0 1px 2px rgba(0,0,0,0.1)'
                            }}>{item.method}</span>
                            <span style={{ color: 'hsl(var(--foreground))', wordBreak: 'break-all' }}>{item.endpoint}</span>
                        </div>
                    )}

                    <div style={{ fontSize: '1.1rem', lineHeight: '1.7', color: 'hsl(var(--foreground))', marginBottom: '1.5rem' }}>
                        {item.description}
                    </div>

                    {renderMarkdownContent(item.content)}
                </div>

                {item.params && item.params.length > 0 && (
                    <div style={{ marginTop: '3rem' }}>
                        <h3 style={{ fontSize: '0.85rem', fontWeight: '700', textTransform: 'uppercase', color: 'hsl(var(--muted-foreground))', marginBottom: '1rem', letterSpacing: '0.05em' }}>Body Parameters</h3>
                        <div style={{ overflowX: 'auto' }}>
                            <table className="doc-table compact">
                                <thead>
                                    <tr>
                                        <th style={{ width: '30%' }}>Parameter</th>
                                        <th style={{ width: '15%' }}>Type</th>
                                        <th>Description</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {item.params.map((param, index) => (
                                        <tr key={index}>
                                            <td style={{ verticalAlign: 'top' }}>
                                                <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
                                                    <span className="param-name">{param.name}</span>
                                                    {param.required && <span className="param-required">REQUIRED</span>}
                                                </div>
                                            </td>
                                            <td style={{ verticalAlign: 'top' }}>
                                                <span className="param-type">{param.type}</span>
                                            </td>
                                            <td style={{ verticalAlign: 'top' }}>
                                                <div style={{ color: 'hsl(var(--foreground))', lineHeight: '1.5' }}>
                                                    {param.description}
                                                </div>
                                                {param.default && (
                                                    <div style={{ marginTop: '0.5rem', fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))' }}>
                                                        Default: <code>{param.default}</code>
                                                    </div>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </div>

            {/* Code Column */}
            <div style={{
                width: '45%',
                minWidth: '400px',
                background: isLight ? '#f8fafc' : '#0f172a',
                borderLeft: '1px solid hsl(var(--border))',
                padding: '2rem',
                display: 'flex',
                flexDirection: 'column',
                height: '100%',
                overflowY: 'auto'
            }}>
                <div style={{ position: 'sticky', top: 0, zIndex: 10 }}>
                    <h3 style={{ color: isLight ? '#64748b' : '#94a3b8', fontSize: '0.8rem', fontWeight: '600', textTransform: 'uppercase', marginBottom: '1rem', letterSpacing: '0.05em' }}>
                        {item.features ? 'Key Capabilities' : 'Example Code'}
                    </h3>
                </div>

                {item.features ? (
                    <div className="features-grid">
                        {item.features.map((feature, i) => {
                            const Icon = Icons[feature.icon] || Icons.Circle;
                            return (
                                <div key={i} className={`feature-card ${isLight ? 'light' : ''}`}>
                                    <div className="feature-icon-wrapper">
                                        <Icon size={18} />
                                    </div>
                                    <div className="feature-content">
                                        <div className="feature-title">{feature.title}</div>
                                        <div className="feature-description">{feature.description}</div>
                                    </div>
                                </div>
                            )
                        })}
                    </div>
                ) : item.code ? (
                    <CodeBlock code={item.code} />
                ) : (
                    <div style={{ padding: '2rem', textAlign: 'center', color: isLight ? '#94a3b8' : '#64748b', border: isLight ? '2px dashed #cbd5e1' : '2px dashed #334155', borderRadius: '8px' }}>
                        No usage example available for this item.
                    </div>
                )}
            </div>
        </div>
    )
}

function getMethodColor(method) {
    switch (method) {
        case 'GET': return '#3b82f6';
        case 'POST': return '#22c55e';
        case 'PUT': return '#f59e0b';
        case 'DELETE': return '#ef4444';
        default: return '#94a3b8';
    }
}
