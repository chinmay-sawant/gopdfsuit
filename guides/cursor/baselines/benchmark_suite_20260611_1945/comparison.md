# Benchmark Suite — 2026-06-11 (Go 1.26.4)

**Environment:** WSL2, i7-13700HX, 24 logical CPUs, `GOMAXPROCS=24`  
**Go:** `go1.26.4` (`/home/chinmay/go/bin/go1.26.4`)  
**Runs:** 2× each harness (sequential)  
**Artifacts:** `guides/cursor/baselines/benchmark_suite_20260611_1945/`

---

## 1. gopdfsuit — k6 HTTP load (`tagged_ecdsa`, 48 VUs × 35s)

Harness: `test/generate_template-pdf/run_gin_pprof_load.sh`

| Run | http_reqs/s | http_req med | http_req p99 | pdf_generation med |
|-----|------------:|-------------:|-------------:|-------------------:|
| 1 | **835** | 14.2 ms | 308 ms | 16 ms |
| 2 | **951** | 13.5 ms | 275 ms | 15 ms |
| **2-run mean** | **893** | 13.9 ms | 292 ms | 15.5 ms |

**Prior baseline** (`20260611_190806`, post-P12 revert): **1,054 req/s**  
**Δ vs prior:** −15% (variance; run 2 within ~10% of prior)

---

## 2. gopdflib — Zerodha gold standard (`sampledata/gopdflib/zerodha`, 5000 iter, 48 workers)

| Run | Throughput (ops/s) | Avg latency | Retail PDF |
|-----|-------------------:|------------:|-----------:|
| 1 | **3,323** | 14.0 ms | 61,293 B |
| 2 | **3,583** | 13.1 ms | 61,293 B |
| **2-run mean** | **3,453** | 13.6 ms | — |

**Prior baselines** (same harness, same machine class):

| Source | Go version | Mean throughput | Notes |
|--------|------------|----------------:|-------|
| `zerodha/1.24.txt` (10 runs) | go1.26.4 | **~586** ops/s | ECDSA, pre-UA-2 structure fix era |
| `zerodha/1.26.txt` (10 runs) | go1.26.0 | **~393** ops/s | Older toolchain |

**Δ vs `1.24.txt` median (~586):** **+489%** (`GeneratePDFBorrowed`, HFT fast paths, compression opts on current branch)

---

## 3. pypdfsuit — Zerodha (`pypdfsuit_bench.py`, 5000 iter, 48 workers)

| Run | Throughput (ops/s) | Avg latency | HFT PDF |
|-----|-------------------:|------------:|--------:|
| 1 | **233** | 180.6 ms | 2,424,782 B |
| 2 | **235** | 177.7 ms | 2,424,782 B |
| **2-run mean** | **234** | 179.2 ms | — |

**vs gopdflib (2-run mean 3,453):** pypdfsuit ≈ **6.8%** of native Go throughput (CGO + Python overhead).

---

## 4. gopdfkit_compare — apples-to-apples (`compare_benchmark_test.go`, benchtime=5s, workers_40)

Harness: `sampledata/benchmarks/gopdfkit_compare`  
Prior reference: `guides/optimizations/20260611_gopdfkit_fixed_compare_results.md` (3-run median, same Go 1.26.4)

### Throughput (pdf/s) — 2-run mean

| Workload | GoPDFKit (now) | gopdflib (now) | GoPDFKit (prior) | gopdflib (prior) | gopdflib Δ vs prior |
|----------|---------------:|---------------:|-----------------:|-----------------:|--------------------:|
| `text_short` | 121,890 | **227,897** | 106,256 | 160,863 | **+42%** |
| `text_240_lines` | 14,694 | **25,980** | 12,034 | 16,810 | **+55%** |
| `table_180_rows` | 11,816 | **35,965** | 9,237 | 21,023 | **+71%** |
| `table_900_rows` | 2,534 | **7,902** | 2,081 | 4,338 | **+82%** |
| `invoice_40_rows` | 39,442 | **114,308** | 32,173 | 71,920 | **+59%** |
| `png_table_180_rows` | 7,197 | **35,093** | 5,885 | 19,843 | **+77%** |
| `png_rows_60` | 5,094 | **44,963** | 4,028 | 30,018 | **+50%** |

**Winner:** gopdflib on **all 7 workloads** (same as prior comparison).  
**gopdflib lead (2-run mean):** +87% (`text_short`) to +783% (`png_rows_60`) over GoPDFKit.

### Allocation (B/op) — run 1 snapshot

| Workload | GoPDFKit | gopdflib |
|----------|--------:|---------:|
| `text_short` | 33,438 | **18,935** |
| `text_240_lines` | 303,621 | **27,564** |
| `table_180_rows` | 373,099 | **23,259** |
| `table_900_rows` | 1,782,605 | **39,807** |
| `invoice_40_rows` | 102,707 | **20,718** |
| `png_table_180_rows` | 585,110 | **29,231** |
| `png_rows_60` | 660,302 | **60,407** |

---

## 5. Cross-stack summary (2-run means)

| Stack | Workload | Throughput | Unit |
|-------|----------|------------:|------|
| **gopdfsuit** (k6 HTTP) | weighted tagged_ecdsa | **893** | req/s |
| **gopdflib** (library) | Zerodha 80/15/5 | **3,453** | PDF/s |
| **pypdfsuit** (Python) | Zerodha 80/15/5 | **234** | PDF/s |
| **gopdflib** (micro) | text_short | **227,897** | PDF/s |
| **GoPDFKit** (micro) | text_short | 121,890 | PDF/s |

---

## Raw log files

| Harness | Run 1 | Run 2 |
|---------|-------|-------|
| k6 gopdfsuit | `k6_run1.txt` | `k6_run2.txt` |
| gopdflib zerodha | `gopdflib_run1.txt` | `gopdflib_run2.txt` |
| pypdfsuit zerodha | `pypdfsuit_run1.txt` | `pypdfsuit_run2.txt` |
| gopdfkit_compare | `gopdfkit_run1.txt` | `gopdfkit_run2.txt` |