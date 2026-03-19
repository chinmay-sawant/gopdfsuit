# PDF Library Benchmark Report

**Date:** March 20, 2026  
**Dataset:** 2,000 user records with ID, Name, Email, Role, and Description fields  
**Environment:** Linux, Python 3.13.10 (venv), Go 1.24, Node.js  
**Concurrency:** Mixed harnesses: 48-worker goroutine/process/thread pools for GoPDFLib, GoPDFSuit, PyPDFSuit, FPDF2, and jsPDF; 10-worker runs for PDFKit, pdf-lib, and Typst  
**Iterations:** Mixed by harness: 1000 (GoPDFLib data), 500 (PyPDFSuit), 48 (GoPDFLib Zerodha, GoPDFSuit, FPDF2, jsPDF), 10 (PDFKit, pdf-lib, Typst)  
**Workload Note:** `GoPDFLib (data)`, `FPDF2`, `jsPDF`, `PDFKit`, `pdf-lib`, and `Typst` use the tabular data benchmark. `GoPDFLib (Zerodha)`, `GoPDFSuit`, and `PyPDFSuit` use the Zerodha single-document benchmark.

## Benchmark Comparison

| Metric | GoPDFLib (data) | GoPDFSuit | PyPDFSuit | GoPDFLib (Zerodha) | FPDF2 | jsPDF | PDFKit | pdf-lib | Typst |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Language | Go | Go | Python | Go | Python | JavaScript | JavaScript | JavaScript | Typst |
| Entry Point | `gopdflib/go run . data` | `gopdfsuit/go run .` | `pypdfsuit/python bench.py` | `gopdflib/go run .` | `fpdf/python bench.py` | `jspdf/node bench.js` | `pdfkit/node bench.js` | `pdflib/node bench.js` | `typst/./bench.sh` |
| Workload | Data table | Zerodha document | Zerodha document | Zerodha document | Data table | Data table | Data table | Data table | Data table |
| Workers | 48 goroutines | 48 goroutines | 48 processes | 48 goroutines | 48 processes | 48 threads | 10 threads | 10 threads | 10 jobs |
| Iterations | 1000 | 48 | 500 | 48 | 48 | 48 | 10 | 10 | 10 |
| Min (ms) | 341.34 | 31.99 | 11.61 | 24.78 | 31551.22 | 15604.86 | 2524.66 | 5258.82 | 2323.02 |
| Avg (ms) | 1327.02 | 65.62 | 99.82 | 55.49 | 39095.73 | 25479.97 | 2917.45 | 5641.60 | 2549.48 |
| P95 (ms) | 3752.29 | 84.37 | 317.47 | 73.93 | 42578.07 | 28404.20 | 3175.15 | 5900.59 | — |
| Max (ms) | 6323.63 | 85.57 | 447.54 | 75.21 | 42847.66 | 30408.45 | 3175.15 | 5900.59 | 2627.21 |
| Throughput (ops/sec) | 35.44 | 546.19 | 328.88 | 627.07 | 1.10 | 1.46 | 2.37 | 1.52 | 3.78 |
| PDF Standard | PDF/A-4 (PDF 2.0) | PDF/A-4 (PDF 2.0) | PDF/A-4 (PDF 2.0) | PDF/A-4 (PDF 2.0) | PDF 1.7 | PDF 1.3 (library max) | PDF 1.7 | PDF 1.7 | PDF/A-2b (typst 0.12 max) |
| Table Support | Full styled table + wrap | Full styled sections/tables | Full styled sections/tables | Full styled sections/tables | Full styled table + wrap | Text fallback | Full styled tables | Manual table drawing | Full styled tables |

## Key Findings

- **GoPDFLib (Zerodha)** is still the fastest overall harness at **627.07 ops/sec**, with 24.78 ms minimum latency and 75.21 ms maximum latency across 48 concurrent runs.
- **GoPDFSuit** remains close behind on the same Zerodha workload at **546.19 ops/sec**, ranging from 31.99 ms to 85.57 ms.
- **PyPDFSuit** reaches **328.88 ops/sec** on the Zerodha workload with a 48-process pool, but its long tail is much wider than the Go implementations at **317.47 ms P95** and **447.54 ms max**.
- **GoPDFLib (data)** remains the fastest tabular-data benchmark at **35.44 ops/sec**, though the 1000-iteration run shows a pronounced tail under sustained concurrency with **3752.29 ms P95** and **6323.63 ms max**.
- **Typst** is the fastest non-Go data-table runtime in this run at **3.78 ops/sec**, compiling each document in roughly **2.32-2.63 s**.
- **PDFKit** follows at **2.37 ops/sec**, averaging **2917.45 ms** per document across 10 worker threads.
- **pdf-lib** manages **1.52 ops/sec**, slightly ahead of **jsPDF** at **1.46 ops/sec** and **FPDF2** at **1.10 ops/sec**.
- **jsPDF** remains constrained by its text-fallback implementation and now averages **25.48 s** per run in the 48-thread harness.
- **FPDF2** is the slowest benchmark in the current suite, averaging **39.10 s** per document over 48 concurrent processes.

## PDF Standard Upgrade Summary

| Library | Before | After | Notes |
|---|---|---|---|
| GoPDFLib / GoPDFSuit / PyPDFSuit | default | **PDF/A-4** (PDF 2.0) | `PDFACompliant: true, ArlingtonCompatible: true, EmbedFonts: true` |
| FPDF2 | 1.4 | **1.7** | `set_pdf_version("1.7")`; v2.8+ supports 1.7 max |
| jsPDF | 1.3 | 1.3 | Library hardcodes PDF 1.3 header; no upgrade option |
| PDFKit | 1.3 (default) | **1.7** | `pdfVersion: '1.7'`; PDF/A-3b skipped (requires TTF embedding) |
| pdf-lib | 1.7 | 1.7 | Already at max; PDF/A not natively supported |
| Typst | a-2b | a-2b | `a-3b` requires typst 0.13+; bundled binary is 0.12.0 |

## Parallel Weighted Workload (Historical)

This earlier benchmark mode drives a weighted 80/15/5 mix of retail, active-trader, and HFT documents across 48 workers in parallel. Numbers represent aggregate concurrent throughput, not single-document latency.

| Runtime | Workers | Throughput | Avg | Min | Max | Retail/Active/HFT |
|---|---:|---:|---:|---:|---:|---:|
| GoPDFLib | 48 | 1913.13 ops/sec | 24.558 ms | 2.280 ms | 505.087 ms | 4004 / 766 / 230 |
| PyPDFSuit | 48 | 233.76 ops/sec | 185.517 ms | 2.657 ms | 3516.474 ms | 4015 / 767 / 218 |

## Commands Used

```bash
bash run_all_benchmarks.sh

cd gopdflib && go run . data         # data-table benchmark (1000 iterations, 48 goroutines)
cd gopdflib && go run .              # Zerodha benchmark (48 iterations, 48 goroutines)
cd gopdfsuit && go run .             # Zerodha benchmark (48 iterations, 48 goroutines)
cd pypdfsuit && python3 bench.py     # Zerodha benchmark (500 iterations, 48 processes)
cd fpdf && python3 bench.py          # data-table benchmark (48 iterations, 48 processes)
cd jspdf && node bench.js            # data-table benchmark (48 iterations, 48 worker_threads)
cd pdfkit && node bench.js           # data-table benchmark (10 iterations, 10 worker_threads)
cd pdflib && node bench.js           # data-table benchmark (10 iterations, 10 worker_threads)
cd typst && bash bench.sh            # data-table benchmark (10 parallel jobs)
```
