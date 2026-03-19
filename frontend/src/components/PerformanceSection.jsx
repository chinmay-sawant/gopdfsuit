
const headlineStats = [
    { value: '40.50 ms', label: 'Best Data Avg', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
    { value: '23.63 ops/sec', label: 'Peak Data Throughput', color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)', border: 'rgba(16, 185, 129, 0.3)' },
    { value: '3.28 ms', label: 'Best Zerodha Avg', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
    { value: '303.20 ops/sec', label: 'Peak Zerodha Throughput', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' },
]

const dataBenchmarks = [
    { name: 'GoPDFLib', avg: '40.50 ms', min: '35.95 ms', max: '51.76 ms', throughput: '23.63 ops/sec' },
    { name: 'PyPDFSuit', avg: '115.12 ms', min: '106.36 ms', max: '137.72 ms', throughput: '8.68 ops/sec' },
]

const zerodhaBenchmarks = [
    { name: 'GoPDFSuit', throughput: '303.20 ops/sec', avg: '3.28 ms', min: '2.61 ms', max: '4.77 ms' },
    { name: 'GoPDFLib', throughput: '269.06 ops/sec', avg: '3.69 ms', min: '2.61 ms', max: '5.03 ms' },
    { name: 'PyPDFSuit', throughput: '179.97 ops/sec', avg: '5.54 ms', min: '4.05 ms', max: '10.54 ms' },
]

const parallelWeightedBenchmarks = [
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
                Captured locally on March 19, 2026 from the checked-in benchmark runners, showing the best observed serial result across 10 full command reruns.
                Historical parallel weighted workload numbers are shown separately below.
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
                <div style={{ marginBottom: '0.75rem', fontWeight: 'bold' }}>Data Table Benchmark</div>
                <table style={tableStyle}>
                    <thead>
                        <tr>
                            <th style={cellStyle}>Library</th>
                            <th style={cellStyle}>Best Avg</th>
                            <th style={cellStyle}>Best Min</th>
                            <th style={cellStyle}>Best Max</th>
                            <th style={cellStyle}>Peak Serial Throughput</th>
                        </tr>
                    </thead>
                    <tbody>
                        {dataBenchmarks.map((benchmark) => (
                            <tr key={benchmark.name}>
                                <td style={cellStyle}>{benchmark.name}</td>
                                <td style={cellStyle}>{benchmark.avg}</td>
                                <td style={cellStyle}>{benchmark.min}</td>
                                <td style={cellStyle}>{benchmark.max}</td>
                                <td style={cellStyle}>{benchmark.throughput}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="performance-logs custom-scrollbar" style={{ marginBottom: '1.25rem' }}>
                <div style={{ marginBottom: '0.75rem', fontWeight: 'bold' }}>Zerodha Contract Note Benchmark</div>
                <table style={tableStyle}>
                    <thead>
                        <tr>
                            <th style={cellStyle}>Runtime</th>
                            <th style={cellStyle}>Avg</th>
                            <th style={cellStyle}>Min</th>
                            <th style={cellStyle}>Max</th>
                            <th style={cellStyle}>Throughput</th>
                        </tr>
                    </thead>
                    <tbody>
                        {zerodhaBenchmarks.map((benchmark) => (
                            <tr key={benchmark.name}>
                                <td style={cellStyle}>{benchmark.name}</td>
                                <td style={cellStyle}>{benchmark.avg}</td>
                                <td style={cellStyle}>{benchmark.min}</td>
                                <td style={cellStyle}>{benchmark.max}</td>
                                <td style={cellStyle}>{benchmark.throughput}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="performance-logs custom-scrollbar" style={{ marginBottom: '1.25rem' }}>
                <div style={{ marginBottom: '0.75rem', fontWeight: 'bold' }}>Parallel Weighted Workload</div>
                <div style={{ marginBottom: '0.75rem', opacity: 0.82, fontSize: '0.85rem' }}>
                    This is a different benchmark mode: a mixed retail, active-trader, and HFT workload executed across 48 workers in parallel. Higher throughput here reflects concurrent aggregate processing, not single-document serial latency.
                </div>
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
                        {parallelWeightedBenchmarks.map((benchmark) => (
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
                    * Serial tables measure one benchmark process at a time. The parallel weighted table measures aggregate throughput across 48 workers, so it should be read as concurrent system throughput rather than single-document latency.
                </span>
            </p>
        </div>
    )
}

export default PerformanceSection
