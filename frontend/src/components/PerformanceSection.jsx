

const PerformanceSection = ({ isVisible }) => {
    return (
        <>
            <h2
                className={`animate-fadeInUp stagger-animation ${isVisible ? 'visible' : ''}`}
                style={{
                    color: 'hsl(var(--foreground))',
                    marginBottom: '1rem',
                    animationDelay: '0.4s',
                }}
            >
                🏃‍♂️ Performance
            </h2>
            <p
                className={`animate-fadeInUp stagger-animation ${isVisible ? 'visible' : ''}`}
                style={{
                    color: 'hsl(var(--muted-foreground))',
                    marginBottom: '2rem',
                    animationDelay: '0.6s',
                }}
            >
                Ultra-fast PDF generation with in-memory processing
            </p>

            {/* Performance Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
                {[
                    { value: '42.3 ms', label: 'Average Response', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
                    { value: '39.0 ms', label: 'Min Response', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
                    { value: '66.8 ms', label: 'Max Response', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' }
                ].map((stat, index) => (
                    <div
                        key={index}
                        className={`animate-fadeInScale stagger-animation ${isVisible ? 'visible' : ''}`}
                        style={{
                            background: stat.bg,
                            padding: '1rem',
                            borderRadius: '8px',
                            border: `1px solid ${stat.border}`,
                            transition: 'all 0.3s ease',
                            animationDelay: `${0.8 + index * 0.2}s`,
                        }}
                    >
                        <div
                            className="animate-pulse"
                            style={{
                                fontSize: '1.5rem',
                                fontWeight: 'bold',
                                color: stat.color,
                                animationDelay: `${2 + index * 0.5}s`,
                            }}
                        >
                            {stat.value}
                        </div>
                        <div style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))' }}>
                            {stat.label}
                        </div>
                    </div>
                ))}
            </div>

            {/* Sample Logs */}
            <div style={{
                background: 'hsl(var(--card))',
                border: '1px solid hsl(var(--border))',
                padding: '1rem',
                borderRadius: '8px',
                fontFamily: 'monospace',
                color: '#4ecdc4',
                fontSize: '0.8rem',
                textAlign: 'left',
                maxHeight: '200px',
                overflowY: 'auto',
                scrollbarWidth: 'thin',
                scrollbarColor: 'rgba(78, 205, 196, 0.5) hsl(var(--card))',
            }}
                className="custom-scrollbar"
            >
                <div style={{ marginBottom: '0.5rem', fontWeight: 'bold' }}>Recent Performance Logs:</div>
                [GIN] 2026/01/19 - 22:45:10 | 200 |      41.25ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:11 | 200 |      43.82ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:12 | 200 |      45.12ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:13 | 200 |      66.79ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:13 | 200 |      42.05ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:14 | 200 |      39.56ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:14 | 200 |      40.11ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:15 | 200 |      44.30ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:15 | 200 |      42.98ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:16 | 200 |      41.77ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:16 | 200 |      48.55ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:17 | 200 |      52.12ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/19 - 22:45:17 | 200 |      40.88ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;
            </div>
            <p style={{
                color: 'hsl(var(--muted-foreground))',
                marginTop: '1rem',
                fontSize: '0.85rem',
                marginBottom: 0,
                lineHeight: 1.6,
            }}>
                Benchmarks for PDF generation with PDF/A compliance, font embedding, digital signatures, bookmarks, and internal links.
                <br />
                <span style={{ fontSize: '0.75rem', fontStyle: 'italic', opacity: 0.8 }}>
                    * Results may vary based on selected options, hardware configuration, data complexity, and network conditions.
                </span>
            </p>
        </>
    )
}

export default PerformanceSection
