

const PerformanceSection = ({ isVisible }) => {
    return (
        <div className={`performance-wrapper animate-fadeInScale stagger-animation ${isVisible ? 'visible' : ''}`}>
            <h2
                className="gradient-text section-heading"
                style={{ animationDelay: '0.4s' }}
            >
                🏃‍♂️ Performance
            </h2>
            <p className="section-subheading" style={{ marginBottom: '2rem' }}>
                Ultra-fast PDF generation with in-memory processing
            </p>

            {/* Performance Stats */}
            <div className="performance-stats-grid">
                {[
                    { value: '~6.64 ms', label: 'Average Response', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
                    { value: '5.79 ms', label: 'Min Response', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
                    { value: '7.33 ms', label: 'Max Response', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' }
                ].map((stat, index) => (
                    <div
                        key={index}
                        className="performance-stat-card"
                        style={{
                            background: stat.bg,
                            borderColor: stat.border,
                        }}
                    >
                        <div className="performance-stat-value" style={{ color: stat.color }}>
                            {stat.value}
                        </div>
                        <div className="performance-stat-label">{stat.label}</div>
                    </div>
                ))}
            </div>

            {/* Sample Logs */}
            <div className="performance-logs custom-scrollbar">
                <div style={{ marginBottom: '0.5rem', fontWeight: 'bold' }}>Recent Performance Logs:</div>
                [GIN] 2026/01/31 - 22:45:10 | 200 |       7.18ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:11 | 200 |       6.47ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:12 | 200 |       6.62ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:13 | 200 |       6.66ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:13 | 200 |       6.57ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:14 | 200 |       6.11ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:14 | 200 |       5.79ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:15 | 200 |       7.31ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:15 | 200 |       7.33ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:16 | 200 |       6.39ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:16 | 200 |       6.25ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:17 | 200 |       6.88ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;<br />
                [GIN] 2026/01/31 - 22:45:17 | 200 |       6.42ms |             ::1 | POST     &quot;/api/v1/generate/template-pdf&quot;
            </div>
            <p className="performance-disclaimer">
                Benchmarks for PDF generation with PDF/A compliance, font embedding, digital signatures, bookmarks, and internal links.
                <br />
                <span style={{ fontSize: '0.75rem', fontStyle: 'italic', opacity: 0.8 }}>
                    * Results may vary based on selected options, hardware configuration, data complexity, and network conditions.
                </span>
            </p>
        </div>
    )
}

export default PerformanceSection
