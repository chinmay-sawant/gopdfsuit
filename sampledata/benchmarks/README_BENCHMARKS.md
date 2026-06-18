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
- Gotenberg (HTML→PDF via Chromium, k6)

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

cd gopdflib && go run -mod=mod . data       # GoPDFLib data-table (5000 iterations, 48 workers, PDF/A)
cd gopdflib && bash run_pprof_bench.sh      # 5000× with 5 CPU + 1 heap pprof captures
cd gopdflib && go run .            # GoPDFLib Zerodha (48 iterations, 48 goroutine workers)
cd gopdfsuit && go run .           # GoPDFSuit Zerodha (48 iterations, 48 goroutine workers)
cd pypdfsuit && python3 bench.py   # PyPDFSuit Zerodha (500 iterations, 48 process workers)
cd fpdf && python3 bench.py        # FPDF2 data-table (48 iterations, 48 process workers)
cd jspdf && node bench.js          # jsPDF data-table (48 iterations, 48 worker_threads)
cd pdfkit && node bench.js         # PDFKit data-table (10 iterations, 10 worker_threads)
cd pdflib && node bench.js         # pdf-lib data-table (10 iterations, 10 worker_threads)
cd typst && bash bench.sh          # Typst data-table (10 parallel bash jobs)

# Gotenberg (Docker required) — weighted k6, mirrors test/generate_template-pdf
make bench-gotenberg               # 48 VU × 35s + artifact capture
make bench-gotenberg-smoke         # 1 VU quick check
```

## Concurrency Model

The suite no longer uses one universal iteration count. Each harness runs with its own configured iteration and worker count, and all iterations execute concurrently within that harness:

| Runtime | Concurrency Mechanism | Iterations | Worker Count |
|---|---|---:|---:|
| GoPDFLib data | Semaphore-guarded goroutine pool + `sync.WaitGroup` + `sync/atomic` | 5000 | 48 |
| GoPDFLib Zerodha | Semaphore-guarded goroutine pool + `sync.WaitGroup` + `sync/atomic` | 48 | 48 |
| GoPDFSuit Zerodha | Semaphore-guarded goroutine pool + `sync.WaitGroup` + `sync/atomic` | 48 | 48 |
| PyPDFSuit | `multiprocessing.Pool` (bypasses the GIL for CPU-bound PDF generation) | 500 | 48 |
| FPDF2 | `multiprocessing.Pool` | 48 | 48 |
| jsPDF | `worker_threads` (`Worker` per iteration, `Promise.all`) | 48 | 48 |
| PDFKit | `worker_threads` (`Worker` per iteration, `Promise.all`) | 10 | 10 |
| pdf-lib | `worker_threads` (`Worker` per iteration, `Promise.all`) | 10 | 10 |
| Typst/Bash | Background jobs (`&`) + `wait` | 10 | 10 |
| Gotenberg | k6 constant VUs → Chromium API | 35s window | 48 VUs / 6 Chromium workers |

These settings reflect the March 20, 2026 benchmark run captured in `benchmark_results_raw.txt`.

## k6 HTTP weighted workload (2026-06-13)

Back-to-back runs, 48 VUs × 35s, `tagged_ecdsa` (80/15/5 retail/active/HFT):

| Runtime | Harness | Peak | Latest avg | http med | http p99 | Notes |
|---|---|---:|---:|---:|---:|---|
| **gopdflib** | Zerodha `go run .` | **2,953 ops/s** | **2,646 ops/s** (30-run) | — | — | Library in-process; not HTTP |
| **gopdfsuit** | `make bench-k6` | **859 req/s** | **825 req/s** (5-run) | 16.0 ms | 347 ms | JSON template + PDF/A + ECDSA |
| **Gotenberg** | `make bench-gotenberg` | — | **10.3 req/s** | 4.26 s | 8.22 s | HTML Chromium; max 6 workers/container |

Comparison: [guides/cursor/baselines/gotenberg_runs/comparison_20260613.md](../../guides/cursor/baselines/gotenberg_runs/comparison_20260613.md)

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
| **GoPDFLib** (2026-06-14, 30-run) | 48 | peak **2953** / avg **2646** ops/sec | 17.67 ms | — | 726.15 ms | 4000 / 750 / 250 |
| GoPDFLib (2026-06-11) | 48 | 1913.13 ops/sec | 24.558 ms | 2.280 ms | 505.087 ms | 4004 / 766 / 230 |
| PyPDFSuit | 48 | 234.62 ops/sec | 178.338 ms | 1.602 ms | 3484.110 ms | 4015 / 767 / 218 |

This table used a weighted 80/15/5 mix of retail, active-trader, and HFT documents. Use it to reason about mixed-workload concurrency capacity; use the per-benchmark tables in `BENCHMARK_REPORT.md` for single-library comparisons.
