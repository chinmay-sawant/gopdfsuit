import { CodeBlock } from './CodeBlock'
import { useTheme } from '../../theme'

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
                    <h1 style={{ fontSize: '2.5rem', fontWeight: '800', marginBottom: '1.5rem', lineHeight: 1.2, letterSpacing: '-0.02em' }}>{item.title}</h1>

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

                    <div style={{ fontSize: '1.05rem', lineHeight: '1.7', color: 'hsl(var(--foreground))', whiteSpace: 'pre-wrap' }}>
                        {item.description}
                    </div>

                    {item.content && (
                        <div style={{ marginTop: '1.5rem', fontSize: '1rem', lineHeight: '1.7', color: 'hsl(var(--muted-foreground))', whiteSpace: 'pre-wrap' }}>
                            {item.content}
                        </div>
                    )}
                </div>

                {item.params && item.params.length > 0 && (
                    <div style={{ marginTop: '3rem' }}>
                        <h3 style={{ fontSize: '0.85rem', fontWeight: '700', textTransform: 'uppercase', color: 'hsl(var(--muted-foreground))', marginBottom: '1rem', letterSpacing: '0.05em' }}>Body Parameters</h3>
                        <div style={{ border: '1px solid hsl(var(--border))', borderRadius: '8px', overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
                            {item.params.map((param, index) => (
                                <div key={index} style={{
                                    padding: '1.25rem',
                                    borderBottom: index < item.params.length - 1 ? '1px solid hsl(var(--border))' : 'none',
                                    background: 'hsl(var(--card))'
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem' }}>
                                        <span style={{ fontWeight: '700', fontFamily: "'Fira Code', monospace", fontSize: '0.95rem' }}>{param.name}</span>
                                        <span style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))', background: 'hsl(var(--muted))', padding: '2px 6px', borderRadius: '4px' }}>{param.type}</span>
                                        {param.required && <span style={{ fontSize: '0.7rem', color: '#ef4444', fontWeight: 'bold', border: '1px solid #ef4444', padding: '1px 4px', borderRadius: '3px' }}>REQUIRED</span>}
                                    </div>
                                    <div style={{ fontSize: '0.95rem', color: 'hsl(var(--foreground))', lineHeight: '1.5' }}>{param.description}</div>
                                    {param.default && <div style={{ fontSize: '0.85rem', color: 'hsl(var(--muted-foreground))', marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        Default: <code style={{ background: 'hsl(var(--muted))', padding: '2px 6px', borderRadius: '4px', fontFamily: 'monospace' }}>{param.default}</code>
                                    </div>}
                                </div>
                            ))}
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
                    <h3 style={{ color: isLight ? '#64748b' : '#94a3b8', fontSize: '0.8rem', fontWeight: '600', textTransform: 'uppercase', marginBottom: '1rem', letterSpacing: '0.05em' }}>Example Code</h3>
                </div>

                {item.code ? (
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
