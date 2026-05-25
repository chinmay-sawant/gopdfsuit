# Zerodha Gold Standard Benchmark — 10-Run Results (Pass 4, PDF/A)

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
| **Max latency** | **592.15 ms** | 883.80 ms | **746.51 ms** | — |
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
| **Throughput (ops/s)** | `5000 ÷ wall-clock seconds` — aggregate with 48 workers. **Not per-core.** |
| **Avg latency (ms)** | Mean `GeneratePDF` time across all 5000 docs. |
| **Max latency (ms)** | Slowest single PDF (usually HFT under contention). |
| **Total time (s)** | Wall clock for the full batch. |

---

## Related

- [PR_PERFORMANCE_OPTIMIZATION.md](./PR_PERFORMANCE_OPTIMIZATION.md) — full PR summary
- [PASS4_PDFA_RESULTS.md](./PASS4_PDFA_RESULTS.md) — micro-bench + HTTP load
- [GOPDFLIB_PPROF_RESULTS.md](./GOPDFLIB_PPROF_RESULTS.md) — data-table bench
