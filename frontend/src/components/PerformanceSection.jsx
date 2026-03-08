
const headlineStats = [
    { value: '2.48 ms', label: 'Fastest Retail Render', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
    { value: '306.05 ops/sec', label: 'Peak Serial Throughput', color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)', border: 'rgba(16, 185, 129, 0.3)' },
    { value: '1913.13 ops/sec', label: 'Go Zerodha Peak', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
    { value: '233.76 ops/sec', label: 'Py Zerodha Peak', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' },
]

const sharedBenchmarks = [
    { name: 'GoPDFLib', time: '2.48 ms', throughput: '306.05 ops/sec' },
    { name: 'GoPDFSuit', time: '2.87 ms', throughput: '243.00 ops/sec' },
    { name: 'PyPDFSuit', time: '3.05 ms', throughput: '211.51 ops/sec' },
]

const zerodhaBenchmarks = [
    { name: 'GoPDFLib', workers: '48', throughput: '1913.13 ops/sec', avg: '24.558 ms', min: '2.280 ms', max: '505.087 ms', mix: '4004 / 766 / 230' },
    { name: 'PyPDFSuit', workers: '48', throughput: '233.76 ops/sec', avg: '185.517 ms', min: '2.657 ms', max: '3516.474 ms', mix: '4015 / 767 / 218' },
]

const machineProfile = [
    'Kernel: Linux 6.6.87.2-microsoft-standard-WSL2',
    'CPU: 13th Gen Intel(R) Core(TM) i7-13700HX',
    'Topology: 12 cores, 24 logical CPUs, 2 threads per core',
    'Memory: 7.6 GiB RAM',
]

const tableStyle = {
    width: '100%',
    borderCollapse: 'collapse',
    fontSize: '0.85rem',
}

const cellStyle = {
    padding: '0.65rem 0.5rem',
    borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
    textAlign: 'left',
}

const PerformanceSection = ({ isVisible }) => {
    return (
        <div className={`performance-wrapper animate-fadeInScale stagger-animation ${isVisible ? 'visible' : ''}`}>
            <h2
                className="gradient-text section-heading"
                style={{ animationDelay: '0.4s' }}
            >
                Measured Performance
            </h2>
            <p className="section-subheading" style={{ marginBottom: '2rem' }}>
                Captured locally on March 9, 2026 from the checked-in benchmark runners, showing the best observed result for each benchmark mode.
            </p>

            <div className="performance-stats-grid">
                {headlineStats.map((stat, index) => (
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

            <div className="performance-logs custom-scrollbar" style={{ marginBottom: '1.25rem' }}>
                <div style={{ marginBottom: '0.75rem', fontWeight: 'bold' }}>Single Zerodha Retail Contract Note</div>
                <table style={tableStyle}>
                    <thead>
                        <tr>
                            <th style={cellStyle}>Library</th>
                            <th style={cellStyle}>Best Time</th>
                            <th style={cellStyle}>Peak Serial Throughput</th>
                        </tr>
                    </thead>
                    <tbody>
                        {sharedBenchmarks.map((benchmark) => (
                            <tr key={benchmark.name}>
                                <td style={cellStyle}>{benchmark.name}</td>
                                <td style={cellStyle}>{benchmark.time}</td>
                                <td style={cellStyle}>{benchmark.throughput}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="performance-logs custom-scrollbar" style={{ marginBottom: '1.25rem' }}>
                <div style={{ marginBottom: '0.75rem', fontWeight: 'bold' }}>Zerodha Weighted Workload</div>
                <table style={tableStyle}>
                    <thead>
                        <tr>
                            <th style={cellStyle}>Runtime</th>
                            <th style={cellStyle}>Workers</th>
                            <th style={cellStyle}>Throughput</th>
                            <th style={cellStyle}>Avg</th>
                            <th style={cellStyle}>Min</th>
                            <th style={cellStyle}>Max</th>
                            <th style={cellStyle}>Retail/Active/HFT</th>
                        </tr>
                    </thead>
                    <tbody>
                        {zerodhaBenchmarks.map((benchmark) => (
                            <tr key={benchmark.name}>
                                <td style={cellStyle}>{benchmark.name}</td>
                                <td style={cellStyle}>{benchmark.workers}</td>
                                <td style={cellStyle}>{benchmark.throughput}</td>
                                <td style={cellStyle}>{benchmark.avg}</td>
                                <td style={cellStyle}>{benchmark.min}</td>
                                <td style={cellStyle}>{benchmark.max}</td>
                                <td style={cellStyle}>{benchmark.mix}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="performance-logs custom-scrollbar">
                <div style={{ marginBottom: '0.5rem', fontWeight: 'bold' }}>Machine Profile</div>
                {machineProfile.map((line) => (
                    <div key={line}>{line}</div>
                ))}
            </div>
            <p className="performance-disclaimer">
                Benchmarks cover PDF generation with PDF/A settings, embedded fonts, bookmarks, internal links, and digital signatures where the runner enables them.
                <br />
                <span style={{ fontSize: '0.75rem', fontStyle: 'italic', opacity: 0.8 }}>
                    * The retail section measures serial single-document rendering, while the Zerodha workload section measures aggregate throughput across 48 workers.
                </span>
            </p>
        </div>
    )
}

export default PerformanceSection
