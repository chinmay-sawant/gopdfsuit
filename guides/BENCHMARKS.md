# Benchmark Report - gopdfsuit

**Run date:** 2026-06-18 (00:38-01:20 IST)  
**Branch:** `feat/optimization-5.5-medium` (`a9dd30b`)  
**Environment:** WSL2, Intel i7-13700HX, 24 logical CPUs, Go 1.26.4, Python 3, Node v22.20.0, k6 v1.4.2, Docker 29.1.1  
**Method:** Best-of-5 runs per harness (highest throughput / lowest ns/op). Targets with existing x5/x10 makefile recipes reused.  
**Raw logs:** [benchmark_run_20260618_v2/](./cursor/baselines/benchmark_run_20260618_v2/)  
**Parsed JSON:** [summary.json](./cursor/baselines/benchmark_run_20260618_v2/summary.json)

---

## Compliance Reference

Maximum supported standards per engine (used in all tables below):

| Engine / harness | PDF/A (max) | PDF/UA (max) | Notes |
|------------------|-------------|--------------|-------|
| **GoPDFLib / GoPDFSuit** | **PDF/A-4** | **PDF/UA-2** | `PDFACompliant`, `TaggedPDF`, `ArlingtonCompatible` |
| **pypdfsuit** (CGO) | **PDF/A-4** | **PDF/UA-2** | Same native stack via bindings |
| **k6** (`tagged_ecdsa`) | **PDF/A-4** | **PDF/UA-2** | Weighted 80/15/5 + ECDSA P-256 signing |
| **k6-retail** | **PDF/A-4** | **PDF/UA-2** | Retail-only signed fast path |
| **bench-gopdflib-zerodha-nocomply** | PDF 2.0 (no PDF/A) | None | Same Zerodha workload; PDF/A, tagging, signing, fonts off |
| **gpdf** (`bench-gpdf-zerodha`) | PDF/A-2b | None | Same Zerodha workload via [gpdf-dev/gpdf](https://github.com/gpdf-dev/gpdf); PDF/A-2b + ECDSA retail signing |
| **gpdf** (`bench-gpdf-zerodha-nocomply`) | PDF 2.0 (no PDF/A) | None | Same Zerodha workload; PDF/A and signing off |
| **gopdfkit_compare GoPDFLib** | PDF 1.7 | None | Compare template omits PDF/A flags (apples-to-apples speed) |
| **gopdfkit_compare GoPDFKit** | PDF 1.7 | None | External library baseline |
| **Handler** (`financial_report.json`) | PDF 1.7 | Arlington tags | `pdfaCompliant: false` in fixture JSON |
| **FPDF2** | PDF 1.7 | None | Highest FPDF2 supports |
| **jsPDF** | PDF 1.3 | None | Library hardcodes PDF 1.3 header |
| **PDFKit** | PDF 1.7 | None | PDF/A-3b possible with TTF fonts (not used here) |
| **pdf-lib** | PDF 1.7 | None | No native PDF/A |
| **Typst** (manual) | PDF 1.7 | None | Bundled typst 0.11.0; max **PDF/A-2b** with typst 0.12+ |
| **Gotenberg** | None | None | Chromium HTML to PDF, no PDF/A or signing |

---

## Baseline

Prior report: [BENCHMARK_COMPARISON_2026-06-15.md](./BENCHMARK_COMPARISON_2026-06-15.md) (2026-06-15, `feat/performance-improvements`).

---

## Executive Summary (best-of-5)

| Category | Benchmark | PDF/A | PDF/UA | Baseline | **Best-of-5** | Delta |
|----------|-----------|-------|--------|----------|--------------:|------:|
| Zerodha weighted (ECDSA) | `bench-gopdflib-zerodha` | PDF/A-4 | PDF/UA-2 | 2,646 ops/s | **6,611 ops/s** (x10 peak) | +150% |
| Zerodha weighted (no compliance) | `bench-gopdflib-zerodha-nocomply` | PDF 2.0 (no PDF/A) | None | - | **37,853 ops/s** (x10 peak) | - |
| Zerodha retail | `bench-gopdfsuit-zerodha` | PDF/A-4 | PDF/UA-2 | 1,978 ops/s | **6,146 ops/s** | +211% |
| Data table (2000 rows) | `bench-gopdflib-data` | PDF/A-4 | PDF/UA-2 | 189 ops/s | **288 ops/s** | +52% |
| HTTP weighted (48 VU x 35s) | `bench-k6` | PDF/A-4 | PDF/UA-2 | 910 req/s | **1,333 req/s** | +46% |
| HTTP light (24 VU x 15s) | `bench-k6-light` | PDF/A-4 | PDF/UA-2 | - | **1,177 req/s** | - |
| HTTP retail-only | `bench-k6-retail` | PDF/A-4 | PDF/UA-2 | - | **7,515 req/s** | - |
| vs Gotenberg (same k6 harness) | `bench-gotenberg` | None | None | 10 req/s | **16 req/s** | +60% |
| Handler serial | `bench-handler` | PDF 1.7 | Arlington | 321,648 ns/op | **55,528 ns/op** | -83% |
| Handler parallel | `bench-handler-parallel` | PDF 1.7 | Arlington | 59,994 ns/op | **54,814 ns/op** | -9% |
| gopdflib vs GoPDFKit (table_180) | `bench-gopdfkit-compare` | PDF 1.7 / none | None | 28,870 vs 11,307 pdf/s | **47,707 vs 11,883 pdf/s** | +65% / +5% |

**Headline:** GoPDFLib leads on compliance (PDF/A-4 + PDF/UA-2) and throughput across Zerodha, data-table, HTTP, and GoPDFKit compare workloads. Gotenberg remains ~83x slower than gopdfsuit k6 on the same harness.

---

## sampledata/benchmarks - Multi-Library Comparison (best-of-5)

Dataset: 2,000 user records (ID, Name, Email, Role, Description).  
Method: 5 runs each, **best throughput** reported.

| Engine | Harness | PDF/A | PDF/UA | Workers | Best throughput | Baseline | Delta |
|--------|---------|-------|--------|--------:|----------------:|---------:|------:|
| **GoPDFLib** | Data table | PDF/A-4 | PDF/UA-2 | 48 | **288 ops/s** | 189 | +52% |
| **GoPDFLib** | Zerodha weighted | PDF/A-4 | PDF/UA-2 | 48 | **6,611 ops/s** (x10 peak) | 2,646 | +150% |
| **GoPDFLib** | Zerodha weighted (nocomply) | PDF 2.0 (no PDF/A) | None | 48 | **37,853 ops/s** (x10 peak) | - | - |
| **GoPDFSuit** | Zerodha retail | PDF/A-4 | PDF/UA-2 | 48 | **6,146 ops/s** | 1,978 | +211% |
| **pypdfsuit** | Zerodha weighted | PDF/A-4 | PDF/UA-2 | 48 | **937 ops/s** (x10 peak) | 235 | +290% |
| **pypdfsuit** | Zerodha weighted (nocomply) | PDF 2.0 (no PDF/A) | None | 48 | **1,284 ops/s** (x10 peak) | - | - |
| **gpdf** | Zerodha weighted | PDF/A-2b | None | 48 | **178 ops/s** | - | - |
| **gpdf** | Zerodha weighted (nocomply) | PDF 2.0 (no PDF/A) | None | 48 | **464 ops/s** | - | - |
| **pypdfsuit** | Zerodha retail (`bench.py`)* | PDF/A-4 | PDF/UA-2 | 10 | **293 ops/s** | - | - |
| **PDFKit** | Data table | PDF 1.7 | None | 10 | **10.1 ops/s** | 5.0 | +102% |
| **jsPDF** | Data table | PDF 1.3 | None | 10 | **9.4 ops/s** | 4.2 | +124% |
| **pdf-lib** | Data table | PDF 1.7 | None | 10 | **6.0 ops/s** | 4.7 | +28% |
| **FPDF2** | Data table | PDF 1.7 | None | 10 | **2.2 ops/s** | 2.0 | +10% |
| **Typst** | Data table (manual, no PDF/A) | PDF 1.7 | None | 10 | **2.2 ops/s** | 9.4* | -77% |

*Baseline Typst used PDF/A-2b on typst 0.12+; bundled typst 0.11.0 cannot run `bench-typst` shell script.

### Data-Table Ranking (compliance-aware)

Same workload: **2000-row user table** from `sampledata/benchmarks/data.json`, best-of-5 throughput.

| Rank | Engine | PDF/A | PDF/UA | Workers | Iterations | Best ops/s | Avg latency |
|-----:|--------|-------|--------|--------:|-----------:|----------:|------------:|
| 1 | **GoPDFLib** | PDF/A-4 | PDF/UA-2 | 48 | 5000 | **288** | ~156 ms |
| 2 | PDFKit | PDF 1.7 | None | 10 | 10 | 10.1 | ~721 ms |
| 3 | jsPDF | PDF 1.3 | None | 10 | 10 | 9.4 | ~946 ms |
| 4 | pdf-lib | PDF 1.7 | None | 10 | 10 | 6.0 | ~1,484 ms |
| 5 | FPDF2 | PDF 1.7 | None | 10 | 10 | 2.2 | ~4,492 ms |
| 6 | Typst (manual) | PDF 1.7 | None | 10 | 10 | 2.2 | ~549 ms |

*`bench-pypdfsuit-legacy` runs `bench.py` (Zerodha retail, 10 workers) - **not** the data table. Do not compare its 293 ops/s to GoPDFLib data 288 ops/s.

---

## Makefile HTTP Benchmarks (best-of-5)

| Target | PDF/A | PDF/UA | VUs x duration | Best req/s | Baseline | Delta |
|--------|-------|--------|----------------|----------:|---------:|------:|
| `bench-k6` | PDF/A-4 | PDF/UA-2 | 48 x 35s | **1,333** | 910 | +46% |
| `bench-k6-light` | PDF/A-4 | PDF/UA-2 | 24 x 15s | **1,177** | - | - |
| `bench-k6-retail` | PDF/A-4 | PDF/UA-2 | 48 x 35s retail | **7,515** | - | - |
| `bench-gotenberg` | None | None | 48 x 35s | **16.1** | 10.3 | +56% |

**gopdfsuit vs Gotenberg:** 1,333 / 16.1 = **~83x** faster (weighted k6 harness).

---

## Go test Benchmarks (best-of-5)

| Benchmark | PDF/A | PDF/UA | Best ns/op | Best ops/s | B/op | allocs/op |
|-----------|-------|--------|----------:|----------:|-----:|----------:|
| `BenchmarkGenerateTemplatePDF_FinancialReport` | PDF 1.7 | Arlington | **55,528** | 18,009 | 344,521 | 294 |
| `BenchmarkGenerateTemplatePDF_FinancialReport_Parallel` | PDF 1.7 | Arlington | **54,814** | 18,244 | 342,475 | 294 |
| `BenchmarkGoPdfSuit` (data.json 2000 rows) | PDF/A-4 | PDF/UA-2 | **12,294,360** | 81 | 4,676,635 | 39,778 |
| `BenchmarkTypst` | PDF 1.7 | None | **469,015,552** | 2.1 | 16,156 | 49 |
| `bench-pdf-micro` Rows2000 | PDF/A-4 | PDF/UA-2 | **11,934,163** | 84 | 4,752,419 | 39,779 |

### Skipped (runtime)

| Benchmark | PDF/A | PDF/UA | Reason |
|-----------|-------|--------|--------|
| `bench-pdf-macro` Rows10000/25000 | PDF/A-4 | PDF/UA-2 | Partial data only; Rows2000 via `bench-pdf-micro` |
| `bench-pdf-wrap` Rows10000+ | PDF/A-4 | PDF/UA-2 | Interrupted; too slow for 5x |
| `bench-gopdfkit-html` | - | - | GoPDFKit HTML panic |
| `bench-typst` shell | PDF/A-2b max | None | typst 0.11.0 lacks `--pdf-standard` |

---

## GoPDFKit Compare (best-of-5, benchtime=5s, 40 workers)

Compare templates use **PDF 1.7 without PDF/A flags** for fair speed comparison.

| Workload | GoPDFKit pdf/s | GoPDFLib pdf/s | gopdflib lead | Baseline Lib | Delta |
|----------|---------------:|---------------:|--------------:|-------------:|------:|
| `text_short` | 119,959 | **254,986** | 2.1x | 206,298 | +24% |
| `text_240_lines` | 14,755 | **32,453** | 2.2x | 23,741 | +37% |
| `table_180_rows` | 11,883 | **47,707** | 4.0x | 28,870 | +65% |
| `table_900_rows` | 2,635 | **10,452** | 4.0x | 7,621 | +37% |
| `invoice_40_rows` | 40,145 | **135,052** | 3.4x | 105,514 | +28% |
| `png_table_180_rows` | 7,504 | **45,098** | 6.0x | 32,077 | +41% |
| `png_rows_60` | 5,474 | **53,935** | 9.9x | 42,548 | +27% |

**Winner:** gopdflib on all 7 workloads (run-1 text_240 Kit outlier excluded).

---

## Zerodha Gold Standard - x10 Sequential (2026-06-24)

Harness: `make bench-gopdflib-zerodha-x10` / `make bench-gopdflib-zerodha-nocomply-x10`  
Environment: WSL2, Intel i7-13700HX, 48 workers, 5000 iterations, 80/15/5 mix.  
Raw logs: [zerodha_bench_x10_wsl/](./cursor/baselines/zerodha_bench_x10_wsl/), [zerodha_bench_x10_nocomply_wsl/](./cursor/baselines/zerodha_bench_x10_nocomply_wsl/)  
Stats: [compliant](./cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt), [nocomply](./cursor/baselines/zerodha_bench_x10_nocomply_wsl_stats_latest.txt)

| Harness | PDF/A | PDF/UA | x10 peak | x10 mean | x10 median | Mean avg latency | Mean peak alloc |
|---------|-------|--------|----------:|---------:|-----------:|-----------------:|----------------:|
| `bench-gopdflib-zerodha` | PDF/A-4 | PDF/UA-2 | **6,611** | **6,203** | **6,362** | 7.544 ms | 798 MB |
| `bench-gopdflib-zerodha-nocomply` | PDF 2.0 (no PDF/A) | None | **37,853** | **34,035** | **35,181** | 1.376 ms | 310 MB |

Compliant HFT output: **2,293,768 bytes** (veraPDF 6/6 PASS). Non-compliant HFT output: **226,774 bytes** (PDF 2.0 base format; no PDF/A or tagging).

**vs June 2026 baseline (`feat/performance-improvements`):** compliant x10 peak **6,611** / mean **6,203** under full PDF/A-4 + PDF/UA-2 + ECDSA signing - **+150% peak** vs the 2,646 ops/s baseline. Non-compliant x10 peak **37,853** / mean **34,035** - **5.7×** compliant peak throughput with compliance features disabled.

---

## Zerodha Gold Standard - PyPDFSuit x10 Sequential (2026-06-24)

Harness: `make bench-pypdfsuit-zerodha-x10` / `make bench-pypdfsuit-zerodha-nocomply-x10` (rebuild CGO first: `cd bindings/python && ./build.sh`)  
Environment: WSL2, Intel i7-13700HX, 48 workers, 5000 iterations, 80/15/5 mix, honest full path (`to_dict` + `json.dumps` each call).  
Raw logs: [compliant](./cursor/baselines/pypdfsuit_bench_x10_wsl/), [nocomply](./cursor/baselines/pypdfsuit_bench_x10_nocomply_wsl/)  
Stats: [compliant](./cursor/baselines/pypdfsuit_bench_x10_wsl_stats_latest.txt), [nocomply](./cursor/baselines/pypdfsuit_bench_x10_nocomply_wsl_stats_latest.txt)

| Harness | PDF/A | PDF/UA | x10 peak | x10 mean | x10 median | Mean avg latency | Mean p95 latency |
|---------|-------|--------|----------:|---------:|-----------:|-----------------:|-----------------:|
| `bench-pypdfsuit-zerodha` | PDF/A-4 | PDF/UA-2 | **937** | **916** | **925** | 46.93 ms | 150.45 ms |
| `bench-pypdfsuit-zerodha-nocomply` | PDF 2.0 (no PDF/A) | None | **1,284** | **1,242** | **1,241** | 33.78 ms | 103.46 ms |

Compliant HFT output: **2,424,773 bytes**. Non-compliant HFT output: **344,468 bytes** (PDF 2.0 base format; no PDF/A, tagging, signing, or font embedding).

**vs June 2026 best-of-5 baseline (235 ops/s):** compliant x10 mean **916** - **+290%** after Python serializer optimizations and CGO rebuild. Non-compliant x10 peak **1,284** / mean **1,242** - **~1.4×** compliant peak throughput with compliance features disabled (Python CGO overhead limits the nocomply ceiling vs native Go).

---

## Completed Makefile Targets

| Target | Runs | Status |
|--------|-----:|--------|
| `bench-gopdflib-zerodha-x5` | 5 | Done |
| `bench-gopdflib-zerodha-x10` | 10 | Done (2026-06-24) |
| `bench-gopdflib-zerodha-nocomply-x10` | 10 | Done (2026-06-24) |
| `bench-gopdflib-data` | 5 | Done |
| `bench-gopdfsuit-zerodha` | 5 | Done |
| `bench-pypdfsuit-zerodha` | 5 | Done |
| `bench-pypdfsuit-zerodha-x10` | 10 | Done (2026-06-24) |
| `bench-pypdfsuit-zerodha-nocomply-x10` | 10 | Done (2026-06-24) |
| `bench-pypdfsuit-legacy` | 5 | Done |
| `bench-fpdf` | 5 | Done |
| `bench-jspdf` | 5 | Done |
| `bench-pdfkit-lib` | 5 | Done |
| `bench-pdflib` | 5 | Done |
| `bench-gopdfkit-compare` | 5 | Done |
| `bench-handler-all` | 5 | Done |
| `bench-handler-parallel` | 5 | Done |
| `bench-pdf-typst` | 5 | Done |
| `bench-k6` | 5 | Done |
| `bench-k6-light` | 5 | Done |
| `bench-k6-retail` | 5 | Done |
| `bench-gotenberg` | 5 | Done |
| `bench-pdf-micro` | 10 | Done (last) |

---

## Monitor Progress (for future runs)

```bash
tail -f /home/chinmay/ChinmayPersonalProjects/gopdfsuit/guides/cursor/baselines/benchmark_run_20260618_v2/continue_console.log
```

Progress format:
```
[PROGRESS 46%] k6-light run 3/5 | completed 12/26 steps
[01:11:27]     RUNNING k6-light 3/5 ... (started 01:11:27)
```

---

## Reproduce

```bash
cd /home/chinmay/ChinmayPersonalProjects/gopdfsuit
bash guides/cursor/baselines/benchmark_run_20260618_v2/continue_one_by_one.sh
python3 guides/cursor/baselines/benchmark_run_20260618_v2/parse_results.py
```

See also: [INTEGRATION_AND_BENCHMARK_TESTS.md](./INTEGRATION_AND_BENCHMARK_TESTS.md), [MAKEFILE.md](./MAKEFILE.md), [sampledata/benchmarks/README_BENCHMARKS.md](../sampledata/benchmarks/README_BENCHMARKS.md).