const headlineStats = [
  { value: '2061 ops/sec', label: 'Peak Zerodha Throughput', color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)', border: 'rgba(16, 185, 129, 0.3)' },
  { value: '1705 ops/sec', label: '10-Run Avg Throughput', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
  { value: '22.7 ms', label: 'Best Avg Latency', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
  { value: '1.73 ms', label: 'Best Min Latency', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' },
]

const dataBenchmarks = [
  { name: 'GoPDFLib', avg: '119.48 ms', min: '112.51 ms', max: '127.17 ms', throughput: '77.81 ops/sec' },
  { name: 'PDFKit', avg: '905.61 ms', min: '820.49 ms', max: '1002.08 ms', throughput: '8.58 ops/sec' },
  { name: 'jsPDF', avg: '1120.94 ms', min: '1058.14 ms', max: '1187.31 ms', throughput: '7.74 ops/sec' },
  { name: 'Typst', avg: '1323.77 ms', min: '1306.09 ms', max: '1378.97 ms', throughput: '7.22 ops/sec' },
  { name: 'pdf-lib', avg: '2041.23 ms', min: '1904.82 ms', max: '2157.59 ms', throughput: '4.13 ops/sec' },
  { name: 'FPDF2', avg: '4829.08 ms', min: '4734.69 ms', max: '4927.40 ms', throughput: '2.02 ops/sec' },
]

const zerodhaBenchmarks = [
  { name: 'GoPDFLib', throughput: '783.34 ops/sec', avg: '10.88 ms', min: '9.47 ms', max: '12.53 ms' },
  { name: 'GoPDFSuit', throughput: '720.33 ops/sec', avg: '11.70 ms', min: '10.52 ms', max: '12.77 ms' },
  { name: 'PyPDFSuit', throughput: '157.26 ops/sec', avg: '39.53 ms', min: '38.33 ms', max: '40.71 ms' },
]

const parallelWeightedBenchmarks = [
  { name: 'GoPDFLib', workers: '48', throughput: '2061.33 ops/sec', avg: '22.680 ms', min: '1.725 ms', max: '659.165 ms', mix: '4002 / 732 / 266' },
  { name: 'GoPDFLib (10-run avg)', workers: '48', throughput: '1704.95 ops/sec', avg: '27.647 ms', min: '1.967 ms', max: '883.804 ms', mix: '~4000 / ~750 / ~250' },
  { name: 'PyPDFSuit', workers: '48', throughput: '233.76 ops/sec', avg: '185.517 ms', min: '2.657 ms', max: '3516.474 ms', mix: '4015 / 767 / 218' },
]

const machineProfile = [
  'Kernel: Linux 6.6.87.2-microsoft-standard-WSL2',
  'CPU: 13th Gen Intel(R) Core(TM) i7-13700HX',
  'Topology: 12 cores, 24 logical CPUs, 2 threads per core',
  'Memory: 7.6 GiB RAM',
]

const BenchmarkPanel = ({ title, description, columns, rows, wide = false }) => (
  <article className={`glass-card performance-panel ${wide ? 'performance-panel-wide' : ''}`}>
    <div className="performance-panel-header">
      <h3>{title}</h3>
      {description ? <p>{description}</p> : null}
    </div>

    <div className="performance-table-wrap custom-scrollbar">
      <table className="performance-table">
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.key}>{column.label}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.name}>
              {columns.map((column) => (
                <td key={column.key}>{row[column.key]}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  </article>
)

const PerformanceSection = ({ isVisible }) => {
  return (
    <div className={`performance-wrapper animate-fadeInScale stagger-animation ${isVisible ? 'visible' : ''}`}>
      <div className="performance-shell glass-card">
        <div className="performance-header-block">
          <div className="comparison-eyebrow">Benchmarks</div>
          <h2 className="gradient-text section-heading" style={{ animationDelay: '0.4s' }}>
            Measured Performance
          </h2>
          <p className="section-subheading performance-intro">
            Captured on WSL2 (May 2026) from the Zerodha Gold Standard benchmark: 5000 iterations, 48 workers, 80% Retail / 15% Active / 5% HFT, PDF/A + tagged PDF + digital signatures.
            Headline numbers are aggregate concurrent throughput (not per-core). Serial 48-iteration retail comparisons and cross-library data-table benchmarks are shown separately below.
          </p>
        </div>

        <div className="performance-stats-grid">
          {headlineStats.map((stat) => (
            <div
              key={stat.label}
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

        <div className="performance-panels-grid">
          <BenchmarkPanel
            title="Zerodha Contract Note (48× serial)"
            description="Single retail contract note rendered 48 times in-process (March 2026 harness). Serial throughput only — not comparable to the 5000×48 gold standard above."
            columns={[
              { key: 'name', label: 'Runtime' },
              { key: 'avg', label: 'Avg' },
              { key: 'min', label: 'Min' },
              { key: 'max', label: 'Max' },
              { key: 'throughput', label: 'Throughput' },
            ]}
            rows={zerodhaBenchmarks}
          />

          <BenchmarkPanel
            wide
            title="Data Table Benchmark"
            description="Single-document serial benchmark covering PDF generation with embedded fonts, internal links, bookmarks, and signing support where the runner enables them."
            columns={[
              { key: 'name', label: 'Library' },
              { key: 'avg', label: 'Best Avg' },
              { key: 'min', label: 'Best Min' },
              { key: 'max', label: 'Best Max' },
              { key: 'throughput', label: 'Peak Serial Throughput' },
            ]}
            rows={dataBenchmarks}
          />

          <BenchmarkPanel
            title="Parallel Weighted Workload (5000×48)"
            description="Mixed retail, active-trader, and HFT traffic with PDF/A compliance. Peak row is the best observed WSL run; avg row is the mean of 10 sequential runs. Throughput is aggregate system throughput across 48 workers."
            columns={[
              { key: 'name', label: 'Runtime' },
              { key: 'workers', label: 'Workers' },
              { key: 'throughput', label: 'Throughput' },
              { key: 'avg', label: 'Avg' },
              { key: 'min', label: 'Min' },
              { key: 'max', label: 'Max' },
              { key: 'mix', label: 'Retail / Active / HFT' },
            ]}
            rows={parallelWeightedBenchmarks}
          />

          <article className="glass-card performance-panel performance-machine-panel">
            <div className="performance-panel-header">
              <h3>Machine Profile</h3>
              <p>Reference environment for the measured numbers above.</p>
            </div>

            <div className="performance-machine-list">
              {machineProfile.map((line) => (
                <div key={line} className="performance-machine-item">
                  {line}
                </div>
              ))}
            </div>

            <div className="performance-note-box">
              Serial tables measure one benchmark process at a time. The parallel weighted table measures aggregate throughput across 48 workers, so it should be read as concurrent system throughput rather than single-document latency.
            </div>
          </article>
        </div>

        <p className="performance-disclaimer">
          Benchmarks cover PDF generation with PDF/A settings, embedded fonts, bookmarks, internal links, and digital signatures where the runner enables them.
        </p>
      </div>
    </div>
  )
}

export default PerformanceSection
