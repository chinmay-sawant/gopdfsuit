# Benchmark Scripts

This directory contains benchmark scripts for:
- jsPDF (Node.js)
- pdf-lib (Node.js)
- PDFKit (Node.js)
- FPDF (Python)
- GoPDFSuit (Go)
- GoPDFLib (Go)
- PyPDFSuit (Python bindings)

## Setup

1. Install Node dependencies:
   ```bash
   npm install
   ```

2. Install Python dependencies:
   ```bash
   pip install fpdf2
   ```

3. Generate data (if not exists):
   ```bash
   go run gen_data.go
   ```

## Running Benchmarks

Run individually:
```bash
node bench_jspdf.js
node bench_pdflib.js
node bench_pdfkit.js
python3 bench_fpdf.py
go run gopdfsuit/bench.go
go run gopdflib/bench.go
python3 pypdfsuit/bench.py
```

## Benchmark Modes

There are two benchmark families in this directory:

1. Serial rerun benchmarks
These are the current default comparisons used in the report and frontend performance section. Each runner executes its own internal iteration count, and we compare the best serial invocation across repeated command reruns.

2. Parallel weighted workload benchmarks
These are historical concurrency benchmarks that drive a weighted retail, active-trader, and HFT workload across 48 workers. Numbers such as `1913.13 ops/sec` for GoPDFLib belong to this mode and are not directly comparable to the serial single-process rerun tables.

### Historical Parallel Weighted Workload Table

| Runtime | Workers | Throughput | Avg | Min | Max | Retail/Active/HFT |
|---|---:|---:|---:|---:|---:|---:|
| GoPDFLib | 48 | 1913.13 ops/sec | 24.558 ms | 2.280 ms | 505.087 ms | 4004 / 766 / 230 |
| PyPDFSuit | 48 | 233.76 ops/sec | 185.517 ms | 2.657 ms | 3516.474 ms | 4015 / 767 / 218 |

Read the parallel table as aggregate throughput under concurrency. Read the serial tables as single-process benchmark results.
