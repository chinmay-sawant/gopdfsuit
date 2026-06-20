const complianceModes = [
  {
    mode: 'Compliant',
    harness: 'bench-gopdflib-zerodha',
    pdfa: 'PDF/A-4',
    pdfua: 'PDF/UA-2',
    peak: '10,005',
    mean: '9,594',
    median: '9,681',
    latency: '4.88 ms',
    alloc: '1,107 MB',
    hftSize: '2.29 MB',
    hftNote: 'veraPDF 6/6 PASS',
    enabled: ['ECDSA P-256 signing', 'Tagged PDF / PDF/UA-2', 'Font embedding', 'Arlington-compatible structure'],
    color: '#10b981',
    bg: 'rgba(16, 185, 129, 0.08)',
    border: 'rgba(16, 185, 129, 0.35)',
  },
  {
    mode: 'Non-compliant',
    harness: 'bench-gopdflib-zerodha-nocomply',
    pdfa: 'PDF 2.0',
    pdfua: 'No PDF/A',
    peak: '26,111',
    mean: '21,564',
    median: '21,621',
    latency: '2.19 ms',
    alloc: '643 MB',
    hftSize: '227 KB',
    hftNote: 'No PDF/A or tagging',
    enabled: ['PDF 2.0 base format', 'PDF/A off', 'Tagging off', 'Signing off', 'Font embedding off'],
    color: '#007acc',
    bg: 'rgba(0, 122, 204, 0.08)',
    border: 'rgba(0, 122, 204, 0.35)',
  },
]

const headlineStats = [
  { value: '2.6×', label: 'Non-compliant peak vs compliant (26,111 / 10,005)', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
  { value: '+278%', label: 'Compliant peak vs June 2026 baseline (2,646 → 10,005)', color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)', border: 'rgba(16, 185, 129, 0.3)' },
  { value: '48 workers', label: 'Zerodha mix: 80% retail · 15% active · 5% HFT', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
  { value: '288 ops/sec', label: 'Peak data-table throughput (PDF/A-4 + PDF/UA-2)', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' },
]

const dataBenchmarks = [
  { name: 'GoPDFLib', avg: '~156 ms', min: '-', max: '-', throughput: '288 ops/sec' },
  { name: 'PDFKit', avg: '~721 ms', min: '-', max: '-', throughput: '10.1 ops/sec' },
  { name: 'jsPDF', avg: '~946 ms', min: '-', max: '-', throughput: '9.4 ops/sec' },
  { name: 'pdf-lib', avg: '~1,484 ms', min: '-', max: '-', throughput: '6.0 ops/sec' },
  { name: 'FPDF2', avg: '~4,492 ms', min: '-', max: '-', throughput: '2.2 ops/sec' },
  { name: 'Typst', avg: '~549 ms', min: '-', max: '-', throughput: '2.2 ops/sec' },
]

const parallelWeightedBenchmarks = [
  { name: 'GoPDFLib (PDF/A-4, x10 peak)', workers: '48', throughput: '10,005 ops/sec', avg: '4.62 ms', min: '0.30 ms', max: '212.14 ms', mix: '4000 / 750 / 250' },
  { name: 'GoPDFLib (PDF/A-4, x10 mean)', workers: '48', throughput: '9,594 ops/sec', avg: '4.88 ms', min: '-', max: '-', mix: '~4000 / ~750 / ~250' },
  { name: 'GoPDFLib (nocomply, x10 peak)', workers: '48', throughput: '26,111 ops/sec', avg: '1.77 ms', min: '-', max: '-', mix: '4000 / 750 / 250' },
  { name: 'GoPDFLib (nocomply, x10 mean)', workers: '48', throughput: '21,564 ops/sec', avg: '2.19 ms', min: '-', max: '-', mix: '~4000 / ~750 / ~250' },
  { name: 'GoPDFSuit (retail)', workers: '48', throughput: '6,146 ops/sec', avg: '6.29 ms', min: '1.36 ms', max: '95.13 ms', mix: '5000 retail' },
  { name: 'PyPDFSuit (weighted)', workers: '48', throughput: '235 ops/sec', avg: '169.07 ms', min: '-', max: '-', mix: '4000 / 750 / 250' },
]

const httpBenchmarks = [
  { name: 'k6 weighted (ECDSA)', vus: '48 x 35s', throughput: '1,333 req/s', compliance: 'PDF/A-4, PDF/UA-2' },
  { name: 'k6 retail-only', vus: '48 x 35s', throughput: '7,515 req/s', compliance: 'PDF/A-4, PDF/UA-2' },
  { name: 'k6 light', vus: '24 x 15s', throughput: '1,177 req/s', compliance: 'PDF/A-4, PDF/UA-2' },
  { name: 'Gotenberg (same harness)', vus: '48 x 35s', throughput: '16.1 req/s', compliance: 'None' },
]

const gopdfkitCompareBenchmarks = [
  { name: 'text_short', workload: 'text_short', gopdfkit: '119,959', gopdflib: '254,986', lead: '2.1x' },
  { name: 'text_240_lines', workload: 'text_240_lines', gopdfkit: '14,755', gopdflib: '32,453', lead: '2.2x' },
  { name: 'table_180_rows', workload: 'table_180_rows', gopdfkit: '11,883', gopdflib: '47,707', lead: '4.0x' },
  { name: 'table_900_rows', workload: 'table_900_rows', gopdfkit: '2,635', gopdflib: '10,452', lead: '4.0x' },
  { name: 'invoice_40_rows', workload: 'invoice_40_rows', gopdfkit: '40,145', gopdflib: '135,052', lead: '3.4x' },
  { name: 'png_table_180_rows', workload: 'png_table_180_rows', gopdfkit: '7,504', gopdflib: '45,098', lead: '6.0x' },
  { name: 'png_rows_60', workload: 'png_rows_60', gopdfkit: '5,474', gopdflib: '53,935', lead: '9.9x' },
]

const machineProfile = [
  'Kernel: Linux 6.6.87.2-microsoft-standard-WSL2',
  'CPU: 13th Gen Intel(R) Core(TM) i7-13700HX',
  'Topology: 12 cores, 24 logical CPUs, 2 threads per core',
  'Go 1.26.4, Python 3, Node v22.20.0, k6 v1.4.2',
  'Branch: feat/optimization-5.5-medium (June 2026)',
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
            Compliant vs Non-Compliant Throughput
          </h2>
          <p className="section-subheading performance-intro">
            Same Zerodha gold-standard workload (5000 iterations, 48 workers, 80/15/5 mix) measured with{' '}
            <code>make bench-gopdflib-zerodha-x10</code> (full compliance) and{' '}
            <code>make bench-gopdflib-zerodha-nocomply-x10</code> (compliance off). WSL2, Intel i7-13700HX, Go 1.26.4 — June 2026 x10 sequential runs.
          </p>
        </div>

        <div className="compliance-compare-grid">
          {complianceModes.map((item) => (
            <article
              key={item.mode}
              className="compliance-compare-card"
              style={{ background: item.bg, borderColor: item.border }}
            >
              <div className="compliance-compare-header">
                <span className="compliance-compare-mode" style={{ color: item.color }}>
                  {item.mode}
                </span>
                <code className="compliance-compare-harness">{item.harness}</code>
              </div>

              <div className="compliance-compare-badges">
                <span className="compliance-badge">{item.pdfa}</span>
                <span className="compliance-badge">{item.pdfua}</span>
              </div>

              <div className="compliance-compare-metrics">
                <div>
                  <div className="compliance-metric-value" style={{ color: item.color }}>
                    {item.peak}
                  </div>
                  <div className="compliance-metric-label">x10 peak ops/sec</div>
                </div>
                <div>
                  <div className="compliance-metric-value">{item.mean}</div>
                  <div className="compliance-metric-label">x10 mean ops/sec</div>
                </div>
                <div>
                  <div className="compliance-metric-value">{item.median}</div>
                  <div className="compliance-metric-label">x10 median</div>
                </div>
              </div>

              <div className="compliance-compare-details">
                <div><strong>Avg latency:</strong> {item.latency}</div>
                <div><strong>Peak alloc:</strong> {item.alloc}</div>
                <div><strong>HFT output:</strong> {item.hftSize} ({item.hftNote})</div>
              </div>

              <ul className="compliance-feature-list">
                {item.enabled.map((feature) => (
                  <li key={feature}>{feature}</li>
                ))}
              </ul>
            </article>
          ))}
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
            title="Parallel Weighted Workload (5000×48)"
            description="Mixed retail, active-trader, and HFT traffic. Compliant and nocomply rows from June 2026 x10 sequential runs (make bench-gopdflib-zerodha-x10 / -nocomply-x10)."
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

          <BenchmarkPanel
            wide
            title="Data Table Benchmark (2000 rows)"
            description="Single-document serial benchmark from sampledata/benchmarks/data.json. GoPDFLib runs with PDF/A-4 and PDF/UA-2; other libraries generate PDF 1.7 without accessibility tagging."
            columns={[
              { key: 'name', label: 'Library' },
              { key: 'avg', label: 'Avg Latency' },
              { key: 'min', label: 'Min' },
              { key: 'max', label: 'Max' },
              { key: 'throughput', label: 'Peak Throughput' },
            ]}
            rows={dataBenchmarks}
          />

          <BenchmarkPanel
            title="HTTP Load Tests (k6)"
            description="End-to-end HTTP benchmarks via make bench-k6 targets. gopdfsuit vs Gotenberg uses the same k6 harness (~83× faster on weighted workload)."
            columns={[
              { key: 'name', label: 'Harness' },
              { key: 'vus', label: 'VUs x Duration' },
              { key: 'throughput', label: 'Peak req/s' },
              { key: 'compliance', label: 'PDF/A / PDF/UA' },
            ]}
            rows={httpBenchmarks}
          />

          <BenchmarkPanel
            wide
            title="GoPDFKit vs GoPDFLib (apples-to-apples)"
            description="make bench-gopdfkit-compare — PDF 1.7 templates without PDF/A flags, 40 workers, benchtime=5s, best-of-5 (June 2026 suite). gopdflib (GoPDFSuit engine) wins all 7 workloads; peak lead 9.9x on PNG rows."
            columns={[
              { key: 'workload', label: 'Workload' },
              { key: 'gopdfkit', label: 'GoPDFKit pdf/s' },
              { key: 'gopdflib', label: 'GoPDFLib pdf/s' },
              { key: 'lead', label: 'gopdflib lead' },
            ]}
            rows={gopdfkitCompareBenchmarks}
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
              Serial data-table numbers measure one document at a time. Parallel Zerodha and HTTP tables measure aggregate throughput across 48 workers, so they should be read as concurrent system throughput rather than single-document latency.
            </div>
          </article>
        </div>

        <p className="performance-disclaimer">
          * All GoPDFLib headline benchmarks run with PDF/A-4, PDF/UA-2, Arlington-compatible tagging, XML metadata generation, ECDSA P-256 digital signatures, embedded fonts, bookmarks, and internal links enabled. GoPDFKit compare templates omit PDF/A flags for fair speed comparison. See guides/BENCHMARKS.md for raw logs and reproduction steps.
        </p>
      </div>
    </div>
  )
}

export default PerformanceSection