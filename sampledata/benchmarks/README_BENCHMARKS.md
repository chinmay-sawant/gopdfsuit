# Benchmark Scripts

This directory contains benchmark scripts for:
- jsPDF (Node.js)
- pdf-lib (Node.js)
- PDFKit (Node.js)
- FPDF2 (Python)
- GoPDFSuit (Go)
- GoPDFLib (Go)
- PyPDFSuit (Python bindings)
- Typst

## Setup

1. Install Node dependencies:
   ```bash
   npm install
   ```

2. Install Python dependencies:
   ```bash
   python3 -m pip install fpdf2
   ```

3. Generate data (if not exists):
   ```bash
   go run gen_data.go
   ```

## Running Benchmarks

```bash
bash run_all_benchmarks.sh

cd gopdflib && go run . data       # GoPDFLib data-table (1000 iterations, 48 goroutine workers)
cd gopdflib && go run .            # GoPDFLib Zerodha (48 iterations, 48 goroutine workers)
cd gopdfsuit && go run .           # GoPDFSuit Zerodha (48 iterations, 48 goroutine workers)
cd pypdfsuit && python3 bench.py   # PyPDFSuit Zerodha (500 iterations, 48 process workers)
cd fpdf && python3 bench.py        # FPDF2 data-table (48 iterations, 48 process workers)
cd jspdf && node bench.js          # jsPDF data-table (48 iterations, 48 worker_threads)
cd pdfkit && node bench.js         # PDFKit data-table (10 iterations, 10 worker_threads)
cd pdflib && node bench.js         # pdf-lib data-table (10 iterations, 10 worker_threads)
cd typst && bash bench.sh          # Typst data-table (10 parallel bash jobs)
```

## Concurrency Model

The suite no longer uses one universal iteration count. Each harness runs with its own configured iteration and worker count, and all iterations execute concurrently within that harness:

| Runtime | Concurrency Mechanism | Iterations | Worker Count |
|---|---|---:|---:|
| GoPDFLib data | Semaphore-guarded goroutine pool + `sync.WaitGroup` + `sync/atomic` | 1000 | 48 |
| GoPDFLib Zerodha | Semaphore-guarded goroutine pool + `sync.WaitGroup` + `sync/atomic` | 48 | 48 |
| GoPDFSuit Zerodha | Semaphore-guarded goroutine pool + `sync.WaitGroup` + `sync/atomic` | 48 | 48 |
| PyPDFSuit | `multiprocessing.Pool` (bypasses the GIL for CPU-bound PDF generation) | 500 | 48 |
| FPDF2 | `multiprocessing.Pool` | 48 | 48 |
| jsPDF | `worker_threads` (`Worker` per iteration, `Promise.all`) | 48 | 48 |
| PDFKit | `worker_threads` (`Worker` per iteration, `Promise.all`) | 10 | 10 |
| pdf-lib | `worker_threads` (`Worker` per iteration, `Promise.all`) | 10 | 10 |
| Typst/Bash | Background jobs (`&`) + `wait` | 10 | 10 |

These settings reflect the March 20, 2026 benchmark run captured in `benchmark_results_raw.txt`.

## PDF Versions

| Library | PDF Standard | Notes |
|---|---|---|
| GoPDFLib / GoPDFSuit / PyPDFSuit | **PDF/A-4** (PDF 2.0) | `PDFACompliant: true, ArlingtonCompatible: true, EmbedFonts: true` |
| FPDF2 | **PDF 1.7** | `pdf.set_pdf_version("1.7")` — highest FPDF2 supports |
| jsPDF | PDF 1.3 | Library hardcodes the PDF 1.3 header — no version option available |
| PDFKit | **PDF 1.7** | `pdfVersion: '1.7'`; PDF/A-3b requires TTF fonts, incompatible with built-in AFM fonts |
| pdf-lib | PDF 1.7 | `save()` produces PDF 1.7; PDF/A not natively supported |
| Typst | PDF/A-2b | `--pdf-standard a-2b`; bundled binary is typst 0.12.0 (a-3b requires 0.13+) |

## Output Metrics

Each benchmark prints:
- Per-run elapsed time in milliseconds
- `Min` / `Avg` / `P95` / `Max` latency for all current harnesses except Typst, which prints `Min` / `Avg` / `Max`
- `Throughput` in ops/sec (wall-clock total ÷ completed iterations)
- `Max Memory Allocated` (Go benchmarks, via `runtime.MemStats`)

## Historical Parallel Weighted Workload Table

| Runtime | Workers | Throughput | Avg | Min | Max | Retail/Active/HFT |
|---|---:|---:|---:|---:|---:|---:|
| GoPDFLib | 48 | 1913.13 ops/sec | 24.558 ms | 2.280 ms | 505.087 ms | 4004 / 766 / 230 |
| PyPDFSuit | 48 | 233.76 ops/sec | 185.517 ms | 2.657 ms | 3516.474 ms | 4015 / 767 / 218 |

This table used a weighted 80/15/5 mix of retail, active-trader, and HFT documents. Use it to reason about mixed-workload concurrency capacity; use the per-benchmark tables in `BENCHMARK_REPORT.md` for single-library comparisons.
