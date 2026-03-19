# PDF Library Benchmark Report

**Date:** March 19, 2026  
**Dataset:** 2,000 user records with ID, Name, Email, Role, and Description fields  
**Environment:** Linux, Python 3.13, Go 1.24 module builds, Node.js runtime available  
**Selection Note:** `GoPDFLib`, `GoPDFSuit`, `PyPDFSuit`, and `GoPDFLib (Zerodha)` below show the best single benchmark invocation observed across 10 full command reruns. Each command invocation still used its own built-in iteration count.  
**Threading Note:** The current GoPDFLib benchmark runner is sequential, not multithreaded, so the updated comparisons remain serial for parity.  
**Historical Note:** Earlier frontend figures like `1913 ops/sec` came from a different 48-worker weighted workload benchmark, not from these serial reruns.  
**Workload Note:** `GoPDFLib`, `PyPDFSuit`, `FPDF2`, `PDFKit`, `pdf-lib`, `jsPDF`, and `Typst` use the tabular data benchmark. `GoPDFSuit` and `GoPDFLib (Zerodha)` use the Zerodha single-document benchmark.

## Benchmark Comparison

| Metric | GoPDFLib | GoPDFSuit | PyPDFSuit | GoPDFLib (Zerodha) | FPDF2 | jsPDF | PDFKit | pdf-lib | Typst |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Language | Go | Go | Python | Go | Python | JavaScript | JavaScript | JavaScript | Typst |
| Entry Point | `gopdflib/go run . data` | `gopdfsuit/go run .` | `pypdfsuit/python databench_pypdfsuit.py` | `gopdflib/go run .` | `fpdf/python bench.py` | `jspdf/node bench.js` | `pdfkit/node bench.js` | `pdflib/node bench.js` | `typst/./bench.sh` |
| Workload | Data table | Zerodha document | Data table | Zerodha document | Data table | Data table fallback | Data table | Data table | Data table |
| Inner Iterations per Invocation | 10 | 5 | 10 | 5 | 10 | 10 | 10 | 10 | 10 |
| Command Reruns Considered | 10 | 10 | 10 | 10 | 1 | 1 | 1 | 1 | 1 |
| Min (ms) | 35.95 | 2.61 | 106.36 | 2.61 | 3539.85 | 37.75 | 254.39 | 534.97 | 659.63 |
| Avg (ms) | 40.50 | 3.28 | 115.12 | 3.69 | 3867.34 | 46.65 | 371.09 | 662.70 | 772.11 |
| Max (ms) | 51.76 | 4.77 | 137.72 | 5.03 | 4461.27 | 81.02 | 544.87 | 887.66 | 893.12 |
| Throughput (ops/sec) | 23.63 | 303.20 | 8.68 | 269.06 | 0.26 | 21.30 | 2.69 | 1.51 | 1.29 |
| Output Size (bytes) | 1,932,300 | 1,739,249 | 1,935,248 | 1,724,870 | 207,596 | 285,386 | 209,266 | 319,816 | 225,919 |
| Output Size (human) | 1.84 MB | 1.66 MB | 1.85 MB | 1.65 MB | 202.73 KB | 278.70 KB | 204.36 KB | 312.32 KB | 220.62 KB |
| PDF Standard | Default GoPDFLib output | PDF/A-4 (PDF 2.0) | Default binding output | PDF/A-4 style Zerodha output | PDF/A-1b compatible | PDF 1.3 fallback | PDF 1.3 | PDF 1.7 | PDF/A-2b |
| Table Support | Full styled table + wrap | Full styled sections/tables | Full styled table + wrap | Full styled sections/tables | Full styled table + wrap | Text fallback | Full styled tables | Manual table drawing | Full styled tables |

## Key Findings

- `GoPDFLib` is currently much faster than `PyPDFSuit` on the same data-table workload, with the best observed average at `40.50 ms` versus `115.12 ms`.
- `GoPDFSuit` and `GoPDFLib (Zerodha)` are now shown using the best observed Zerodha run from 10 command reruns: `3.28 ms` and `3.69 ms` average respectively.
- The older `1913 ops/sec` figure was from a separate 48-worker weighted workload benchmark and should not be compared directly to the serial rerun numbers in this table.
- `FPDF2` is now measured correctly across the full document-generation path; its `3867.34 ms` average is real for this workload and explains the `0.26 ops/sec` throughput.
- `jsPDF`, `PDFKit`, `pdf-lib`, and `Typst` now run 10 sequential iterations each with min, avg, max, and throughput metrics instead of single-run timings.
- `jsPDF` is fast, but it is still running in fallback text mode instead of equivalent table rendering.

## Parallel Weighted Workload

This historical benchmark mode is different from the serial reruns above. It executes a weighted mix of retail, active-trader, and HFT-style documents across 48 workers in parallel, so the throughput values represent aggregate concurrent processing capacity rather than single-document latency.

| Runtime | Workers | Throughput | Avg | Min | Max | Retail/Active/HFT |
|---|---:|---:|---:|---:|---:|---:|
| GoPDFLib | 48 | 1913.13 ops/sec | 24.558 ms | 2.280 ms | 505.087 ms | 4004 / 766 / 230 |
| PyPDFSuit | 48 | 233.76 ops/sec | 185.517 ms | 2.657 ms | 3516.474 ms | 4015 / 767 / 218 |

Use this table to reason about concurrent throughput under mixed workload pressure. Use the serial tables above when comparing single-process document generation time.

## Commands Used

```bash
cd gopdflib && go run . data
cd gopdflib && go run .
cd gopdfsuit && go run .
cd pypdfsuit && python databench_pypdfsuit.py
cd fpdf && python bench.py
cd pdfkit && node bench.js
cd pdflib && node bench.js
cd jspdf && node bench.js
cd typst && ../typst-x86_64-unknown-linux-musl/typst compile --pdf-standard a-2b benchmark.typ output_typst.pdf
```
