# Zerodha Gold Standard Benchmark - gopdflib

## Latest - 30-run results (2026-06-14, Go 1.26.4)

**Date:** 2026-06-14  
**Entry:** `sampledata/gopdflib/zerodha/main.go`  
**Workload:** 5000 iterations, 48 workers, 80% Retail / 15% Active / 5% HFT  
**Config:** PDF/A + PDF/UA-2, retail **ECDSA P-256** signing  
**Environment:** WSL2 Linux amd64, i7-13700HX, 24 CPUs, `GOMAXPROCS=24`, Go **1.26.4**  
**Note:** Sequential back-to-back runs after other benchmarks same evening; thermal variance visible in runs 3–4 and 25–27.

### Aggregate (30 runs)

| Metric | Best | Worst | **Average** | σ |
|--------|------|-------|-------------|---|
| **Throughput** | **2952.88 ops/s** | 2183.43 ops/s | **2645.74 ops/s** | 187.28 |
| **Avg latency** | **15.86 ms** | 21.04 ms | **17.67 ms** | - |
| **Max latency** | 305.38 ms | **726.15 ms** | - | - |
| **Wall time (5000)** | **1.69 s** | 2.29 s | **1.90 s** | - |
| **Peak memory** | 1212.77 MB | **1491.17 MB** | **1330 MB** | - |

**Headline:** **2953 ops/s peak**, **2646 ops/s** 30-run mean.

**vs prior 13-run (`2026-06-13`, idle machine):** peak 3604 → **2953** (−18%), mean 2787 → **2646** (−5%).

**Artifacts:** [baselines/zerodha_bench_x30_20260614_stats.txt](./baselines/zerodha_bench_x30_20260614_stats.txt) · raw logs in `baselines/zerodha_bench_x30_20260614_010104/`

---

## Prior session - 13-run results (2026-06-13, Go 1.26.4, historical best)

**Date:** 2026-06-13  
**Entry:** `sampledata/gopdflib/zerodha/main.go`  
**Workload:** 5000 iterations, 48 workers, 80% Retail / 15% Active / 5% HFT  
**Config:** PDF/A + PDF/UA-2, retail **ECDSA P-256** signing  
**Environment:** WSL2 Linux amd64, i7-13700HX, 24 CPUs, `GOMAXPROCS=24`, Go **1.26.4**

| Run | Throughput (ops/s) | Avg latency (ms) | Max latency (ms) | Total time (s) | Peak mem (MB) |
|-----|-------------------:|-----------------:|-----------------:|---------------:|--------------:|
| 1 | 2981.52 | 15.724 | 365.196 | 1.677 | 1345.17 |
| 2 | 2505.63 | 18.612 | 445.660 | 1.996 | 1274.23 |
| 3 | 2593.90 | 18.100 | 461.272 | 1.928 | 1280.03 |
| 4 | 2420.95 | 19.327 | 535.979 | 2.065 | 1304.32 |
| 5 | 2472.77 | 18.873 | 526.137 | 2.022 | 1291.94 |
| 6 | 2674.73 | 17.499 | 411.122 | 1.869 | 1270.41 |
| 7 | 2547.56 | 18.490 | 523.635 | 1.963 | 1251.65 |
| 8 | 2722.93 | 17.081 | 620.455 | 1.836 | 1308.22 |
| 9 | 3279.75 | 14.287 | 420.948 | 1.525 | 1383.41 |
| 10 | 2960.37 | 15.646 | 407.111 | 1.689 | 1360.12 |
| 11 | 2367.88 | 19.798 | 687.504 | 2.112 | 1408.00 |
| 12 | **3603.53** (peak) | **12.986** (best avg) | 356.804 | **1.388** (best) | 1264.70 |
| 13 | 3096.90 | 15.179 | 457.171 | 1.615 | 1338.06 |

### Aggregate (13 runs)

| Metric | Best | Worst | **Average** | σ |
|--------|------|-------|-------------|---|
| **Throughput** | **3603.53 ops/s** | 2367.88 ops/s | **2786.80 ops/s** | 373.31 |
| **Avg latency** | **12.99 ms** | 19.80 ms | **17.05 ms** | - |
| **Max latency** | 356.80 ms | **687.50 ms** | - | - |
| **Wall time (5000)** | **1.39 s** | 2.11 s | **1.80 s** | - |
| **Peak memory** | 1251.65 MB | **1408.00 MB** | **1314 MB** | - |

**Headline:** **3604 ops/s peak**, **2787 ops/s** 13-run mean.

**vs prior 2-run avg (2026-06-11):** 2459 → **2787 ops/s** (+13%).  
**vs Pass 4 10-run avg (2026-05-25):** 1705 → **2787 ops/s** (+63%).

**Artifacts:** [baselines/zerodha_bench_x13_20260613_stats.txt](./baselines/zerodha_bench_x13_20260613_stats.txt)

---

## Historical - 10-Run Results (Pass 4, PDF/A)

**Date:** 2026-05-25 (revised, WSL native)
**Entry:** `sampledata/gopdflib/zerodha/main.go`  
**Workload:** 5000 iterations, 48 workers, 80% Retail / 15% Active / 5% HFT  
**Config:** `pdfaCompliant: true`, `taggedPDF: true`, retail includes digital signature  
**Environment:** WSL2 Linux amd64, i7-13700HX, 24 CPUs, Go 1.24.0  

---

## How to run

```bash
# Single run (5000 iterations)
cd sampledata/gopdflib/zerodha && go run .

# 10 sequential timing runs
bash sampledata/gopdflib/zerodha/run_bench_x10.sh

# 5 timing runs + 5 CPU profiles + 1 heap profile
bash sampledata/gopdflib/zerodha/run_bench_x5.sh
```

---

## Timing results (10 sequential WSL runs)

| Run | Throughput (ops/s) | Avg latency (ms) | Max latency (ms) | Total time (s) | Peak mem (MB) |
|-----|-------------------|------------------|------------------|----------------|---------------|
| 1 | 1634.72 | 28.540 | 592.154 | 3.059 | 1158.41 |
| 2 | 1741.68 | 26.742 | 667.205 | 2.871 | 1170.82 |
| 3 | 1756.62 | 26.707 | 709.438 | 2.846 | 1253.73 |
| 4 | 1613.58 | 29.070 | 883.804 | 3.099 | 1141.77 |
| 5 | 1701.42 | 27.799 | 751.121 | 2.939 | 1197.37 |
| 6 | **1542.39** (worst tput) | **30.467** (worst avg) | 733.011 | **3.242** (worst) | 1217.82 |
| 7 | 1601.74 | 29.150 | 619.333 | 3.122 | 1216.33 |
| 8 | 1557.89 | 30.050 | 755.173 | 3.209 | 1094.63 |
| 9 | **2020.59** (batch best) | **22.988** (batch best avg) | 757.213 | **2.475** (batch best) | 1087.53 |
| 10 | 1878.83 | 24.958 | 694.527 | 2.661 | **1074.23** (best mem) |

### Aggregate (10 runs)

| Metric | Best | Worst | **Average** | σ |
|--------|------|-------|-------------|---|
| **Throughput** | **2020.59 ops/s** | 1542.39 ops/s | **1704.95 ops/s** | 151.25 |
| **Avg latency** | **22.99 ms** | 30.47 ms | **27.65 ms** | 2.55 ms |
| **Max latency** | **592.15 ms** | 883.80 ms | **746.51 ms** | - |
| **Wall time** | **2.48 s** | 3.24 s | **2.95 s** | 0.28 s |
| **Peak memory** | **1074 MB** | 1254 MB | **1170 MB** | 58 MB |

**Peak observed (manual WSL run, idle machine):** **2061.33 ops/s**, **22.68 ms** avg, **2.43 s** wall time.

**Headline for reports:** **2061 ops/s peak**, **1705 ops/s** 10-run average.

**Run-to-run variance:** Throughput spans ~1542–2061 ops/s depending on system load and random HFT draw (209–275 HFT docs per run).

**Artifacts:** [baselines/zerodha_bench_x10_wsl/](./baselines/zerodha_bench_x10_wsl/), [baselines/zerodha_bench_x10_wsl_stats_20260525.txt](./baselines/zerodha_bench_x10_wsl_stats_20260525.txt)

---

## vs historical Go 1.24 / Go 1.26 (10-run averages)

Sources: [1.24.txt](../sampledata/gopdflib/zerodha/1.24.txt), [1.26.txt](../sampledata/gopdflib/zerodha/1.26.txt)

| Metric | Go 1.24 (10-run avg) | Go 1.26 (10-run avg) | **Pass 4 (10-run avg)** | **Pass 4 peak** | vs 1.24 avg |
|--------|----------------------|----------------------|-------------------------|-----------------|-------------|
| **Throughput** | 574 ops/s | 392 ops/s | **1705 ops/s** | **2061 ops/s** | **+197%** |
| **Avg latency** | 80.7 ms | 119.2 ms | **27.7 ms** | **22.7 ms** | **66% faster** |
| **Wall time (5000)** | 8.75 s | 12.84 s | **2.95 s** | **2.43 s** | **66% faster** |

Pass 4 **peak** beats Go 1.24 **best** (638 ops/s) by **~223%** and matches/exceeds historical Opt 6 best (1741 ops/s).

---

## Column glossary

| Column | What it measures |
|--------|------------------|
| **Throughput (ops/s)** | `5000 ÷ wall-clock seconds` - aggregate with 48 workers. **Not per-core.** |
| **Avg latency (ms)** | Mean `GeneratePDF` time across all 5000 docs. |
| **Max latency (ms)** | Slowest single PDF (usually HFT under contention). |
| **Total time (s)** | Wall clock for the full batch. |

---

## Related

- [PR_PERFORMANCE_OPTIMIZATION.md](./PR_PERFORMANCE_OPTIMIZATION.md) - full PR summary
- [PASS4_PDFA_RESULTS.md](./PASS4_PDFA_RESULTS.md) - micro-bench + HTTP load
- [GOPDFLIB_PPROF_RESULTS.md](./GOPDFLIB_PPROF_RESULTS.md) - data-table bench
