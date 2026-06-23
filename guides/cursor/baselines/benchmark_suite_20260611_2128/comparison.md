# Benchmark Suite - 2026-06-11 22:01 IST (Go 1.26.4)

**Environment:** WSL2, i7-13700HX, 24 logical CPUs, `GOMAXPROCS=24`  
**Go:** `go1.26.4`  
**Runs:** 2× each harness (sequential, full suite back-to-back)  
**Artifacts:** `guides/cursor/baselines/benchmark_suite_20260611_2128/`

> **Note:** This suite ran immediately after k6 load tests; gopdflib/k6 numbers are lower than the earlier `19:45` run due to thermal/load variance. Compare trends, not single peaks.

---

## Master comparison (2-run mean vs prior baselines)

| Stack | Harness | **Now** (22:01) | **Prior** (`19:45` same day) | **Prior** (`19:08` k6 gate) | Δ vs 19:45 |
|-------|---------|----------------:|-----------------------------:|------------------------------:|-----------:|
| **gopdfsuit** | k6 `tagged_ecdsa` | **652 req/s** | 893 req/s | 1,054 req/s | −27% |
| **gopdflib** | Zerodha 5k/48w | **2,459 ops/s** | 3,453 ops/s | ~586 ops/s (Jun old) | −29% |
| **pypdfsuit** | Zerodha 5k/48w | **219 ops/s** | 236 ops/s | - | −7% |
| **gopdflib** | handler bench serial | **1.05 ms/op** | - | 6.07 ms/op (doc) | −83% |
| **gopdflib** | handler bench parallel | **0.133 ms/op** | - | 0.817 ms/op (doc) | −84% |

### k6 detail (48 VU × 35s, `tagged_ecdsa`)

| Run | req/s | med | p99 |
|-----|------:|----:|----:|
| 1 | 611 | 20.9 ms | 497 ms |
| 2 | 693 | 19.2 ms | 437 ms |
| **Mean** | **652** | **20.1 ms** | **467 ms** |

### gopdflib Zerodha (ECDSA P-256, PDF/UA-2 TD hierarchy)

| Run | ops/s | Avg latency |
|-----|------:|------------:|
| 1 | 2,455 | 19.1 ms |
| 2 | 2,463 | 18.9 ms |
| **Mean** | **2,459** | **19.0 ms** |

### pypdfsuit Zerodha

| Run | ops/s | Avg latency |
|-----|------:|------------:|
| 1 | 211 | 195 ms |
| 2 | 228 | 187 ms |
| **Mean** | **219** | **191 ms** |

---

## gopdfkit_compare (pdf/s, 2-run mean, workers_40)

| Workload | GoPDFKit now | gopdflib now | gopdflib prior (Jun 11 3-run med) | gopdflib Δ |
|----------|-------------:|-------------:|----------------------------------:|-----------:|
| `text_short` | 109,185 | **204,214** | 160,863 | +27% |
| `text_240_lines` | 13,218 | **24,808** | 16,810 | +48% |
| `table_180_rows` | 10,855 | **37,245** | 21,023 | +77% |
| `table_900_rows` | 2,326 | **8,057** | 4,338 | +86% |
| `invoice_40_rows` | 36,408 | **115,914** | 71,920 | +61% |
| `png_table_180_rows` | 6,596 | **35,759** | 19,843 | +80% |
| `png_rows_60` | 4,674 | **44,517** | 30,018 | +48% |

**Winner:** gopdflib on all 7 workloads (unchanged).

---

## Handler micro-benchmark (`financial_report.json`)

| Benchmark | ns/op (2-run avg) | B/op | allocs/op |
|-----------|------------------:|-----:|----------:|
| `FinancialReport` serial | **1,045,205** | 599,561 | 990 |
| `FinancialReport_Parallel` | **133,124** | 561,881 | 990 |

Prior doc values (2026-06-11 morning): serial **6,067,362** ns/op, parallel **816,872** ns/op.

---

## Raw logs

| Harness | Run 1 | Run 2 |
|---------|-------|-------|
| k6 | `k6_run1.txt` | `k6_run2.txt` |
| gopdflib | `gopdflib_run1.txt` | `gopdflib_run2.txt` |
| pypdfsuit | `pypdfsuit_run1.txt` | `pypdfsuit_run2.txt` |
| gopdfkit | `gopdfkit_run1.txt` | `gopdfkit_run2.txt` |
| handler | `handler_bench.txt` | - |