import { useState } from 'react'
import { ChevronDown, ChevronRight, Search } from 'lucide-react'

export const Sidebar = ({ sections, activeId, onItemClick }) => {
    const [expanded, setExpanded] = useState(
        sections.reduce((acc, section) => ({ ...acc, [section.title]: true }), {})
    )

    const [searchQuery, setSearchQuery] = useState('')

    const toggleSection = (title) => {
        setExpanded(prev => ({ ...prev, [title]: !prev[title] }))
    }

    // Filter sections based on search query
    const filteredSections = sections.map(section => {
        // If search is empty, return section as is
        if (!searchQuery.trim()) return section;

        // Filter items match title or description
        const filteredItems = section.items.filter(item =>
            item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
            (item.description && item.description.toLowerCase().includes(searchQuery.toLowerCase()))
        );

        // Return section only if it has matching items
        if (filteredItems.length > 0) {
            return { ...section, items: filteredItems };
        }
        return null;
    }).filter(Boolean);

    return (
        <div style={{
            width: '280px',
            height: '100%',
            overflowY: 'auto',
            borderRight: '1px solid hsl(var(--border))',
            background: 'hsl(var(--background))',
            padding: '1.5rem 1rem',
            flexShrink: 0,
            display: 'flex',
            flexDirection: 'column'
        }}>
            <div style={{ marginBottom: '1.5rem', position: 'relative' }}>
                <input
                    placeholder="Jump to..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    style={{
                        width: '100%',
                        padding: '0.5rem 0.5rem 0.5rem 2.2rem',
                        borderRadius: '6px',
                        border: '1px solid hsl(var(--border))',
                        background: 'hsl(var(--input))',
                        color: 'hsl(var(--foreground))',
                        fontSize: '0.9rem',
                        outline: 'none'
                    }}
                />
                <Search size={14} style={{ position: 'absolute', left: '10px', top: '50%', transform: 'translateY(-50%)', color: 'hsl(var(--muted-foreground))' }} />
            </div>

            <div style={{ flex: 1 }}>
                {filteredSections.map(section => (
                    <div key={section.title} style={{ marginBottom: '1.5rem' }}>
                        <div
                            onClick={() => toggleSection(section.title)}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'space-between',
                                cursor: 'pointer',
                                marginBottom: '0.5rem',
                                color: 'hsl(var(--foreground))',
                                fontWeight: '700',
                                fontSize: '0.8rem',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                                padding: '0 0.5rem'
                            }}
                        >
                            {section.title}
                            {expanded[section.title] ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                        </div>

                        {(expanded[section.title] || searchQuery) && (
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                                {section.items.map(item => (
                                    <div
                                        key={item.id}
                                        id={`sidebar-item-${item.id}`}
                                        onClick={() => onItemClick(item)}
                                        style={{
                                            padding: '0.4rem 0.5rem',
                                            cursor: 'pointer',
                                            fontSize: '0.9rem',
                                            color: activeId === item.id ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground))',
                                            background: activeId === item.id ? 'hsl(var(--primary) / 0.1)' : 'transparent',
                                            borderRadius: '4px',
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '8px',
                                            transition: 'all 0.2s ease',
                                            fontWeight: activeId === item.id ? '500' : 'normal'
                                        }}
                                        onMouseEnter={(e) => {
                                            if (activeId !== item.id) {
                                                e.currentTarget.style.color = 'hsl(var(--foreground))'
                                                e.currentTarget.style.background = 'hsl(var(--accent))'
                                            }
                                        }}
                                        onMouseLeave={(e) => {
                                            if (activeId !== item.id) {
                                                e.currentTarget.style.color = 'hsl(var(--muted-foreground))'
                                                e.currentTarget.style.background = 'transparent'
                                            }
                                        }}
                                    >
                                        <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: activeId === item.id ? 'hsl(var(--primary))' : 'transparent' }}></div>
                                        <span style={{ flex: 1, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{item.title}</span>
                                        {item.method && (
                                            <span style={{
                                                fontSize: '0.6rem',
                                                fontWeight: 'bold',
                                                padding: '2px 6px',
                                                borderRadius: '4px',
                                                color: 'white',
                                                backgroundColor: getMethodColor(item.method),
                                                minWidth: '40px',
                                                textAlign: 'center',
                                                lineHeight: '1.2'
                                            }}>
                                                {item.method}
                                            </span>
                                        )}
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                ))}
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
