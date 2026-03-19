const headlineStats = [
  { value: '10.88 ms', label: 'Best Zerodha Avg', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
  { value: '783.34 ops/sec', label: 'Peak Zerodha Throughput', color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)', border: 'rgba(16, 185, 129, 0.3)' },
  { value: '9.47 ms', label: 'Best Zerodha Min', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
  { value: '12.53 ms', label: 'Best Zerodha Max', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' },
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
  { name: 'GoPDFLib', workers: '48', throughput: '1913.13 ops/sec', avg: '24.558 ms', min: '2.280 ms', max: '505.087 ms', mix: '4004 / 766 / 230' },
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
            Captured locally on March 20, 2026 from the latest checked-in benchmark suite run.
            The headline numbers below refer to the Zerodha Contract Note benchmark, a real-world template workload focused on serial generation latency and throughput. Historical parallel weighted workload numbers are shown separately below.
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
            title="Zerodha Contract Note"
            description="Real-world template benchmark focused on serial generation latency and throughput for contract note workloads."
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
            title="Parallel Weighted Workload"
            description="Mixed retail, active-trader, and HFT traffic executed across 48 workers. Higher throughput here reflects aggregate concurrent processing rather than single-document latency."
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
