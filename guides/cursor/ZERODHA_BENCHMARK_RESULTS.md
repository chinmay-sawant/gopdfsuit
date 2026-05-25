# Zerodha Gold Standard Benchmark — 5-Run Results (Pass 4, PDF/A)

**Date:** 2026-05-25  
**Entry:** `sampledata/gopdflib/zerodha/main.go`  
**Workload:** 5000 iterations, 48 workers, 80% Retail / 15% Active / 5% HFT  
**Config:** `pdfaCompliant: true`, `taggedPDF: true`, retail includes digital signature  

---

## How to run

```bash
# Single run (5000 iterations)
go run ./sampledata/gopdflib/zerodha/main.go

# 5 timing runs + 5 CPU profiles + 1 heap profile
bash sampledata/gopdflib/zerodha/run_bench_x5.sh

# Custom iteration count
BENCH_ITERATIONS=5000 BENCH_WORKERS=48 go run ./sampledata/gopdflib/zerodha/main.go

# With pprof
go build -o /tmp/zerodha_bench ./sampledata/gopdflib/zerodha/main.go
/tmp/zerodha_bench -cpuprofile=cpu.prof
/tmp/zerodha_bench -memprofile=heap.prof
```

---

## Timing results (5 runs, no profile overhead)

| Run | Throughput (ops/s) | Avg latency (ms) | Max latency (ms) | Total time (s) | Peak mem (MB) |
|-----|-------------------|------------------|------------------|----------------|---------------|
| 1 | 856.48 | 54.48 | 1241.49 | 5.84 | 1197.87 |
| 2 | 853.59 | 55.14 | 1340.78 | 5.86 | 1303.52 |
| 3 | **829.60** (worst tput) | **56.45** (worst avg) | **1489.98** (worst max) | **6.03** (worst) | 1177.13 |
| 4 | **907.83** (best tput) | **51.48** (best avg) | **992.74** (best max) | **5.51** (best) | 1157.20 |
| 5 | 848.96 | 54.77 | 1142.12 | 5.89 | **1153.43** (best mem) |

### Aggregate (5 runs)

| Metric | Best | Worst | **Average** | σ |
|--------|------|-------|-------------|---|
| **Throughput** | **907.83 ops/s** | 829.60 ops/s | **859.29 ops/s** | 29.09 |
| **Avg latency** | **51.48 ms** | 56.45 ms | **54.46 ms** | 1.83 ms |
| **Max latency** | **992.74 ms** | 1489.98 ms | **1241.42 ms** | 189.31 ms |
| **Wall time** | **5.51 s** | 6.03 s | **5.82 s** | 0.19 s |
| **Peak memory** | **1153 MB** | 1304 MB | **1198 MB** | 62 MB |

**Artifacts:** [baselines/zerodha_pprof_runs/zerodha_run{1..5}.txt](./baselines/zerodha_pprof_runs/)

---

## vs historical Go 1.24 baseline (single run, pre-Pass 4)

Source: [sampledata/gopdflib/zerodha/1.24.txt](../sampledata/gopdflib/zerodha/1.24.txt)

| Metric | Go 1.24 (Feb 2026) | Pass 4 avg (5 runs) | Change |
|--------|-------------------|---------------------|--------|
| **Throughput** | 637.87 ops/s | **859.29 ops/s** | **+35%** |
| **Avg latency** | 71.88 ms | **54.46 ms** | **−24%** |
| **Max latency** | 1755.76 ms | **1241.42 ms** avg | **−29%** |
| **Wall time (5000)** | 7.84 s | **5.82 s** | **−26%** |
| **Peak memory** | 1172 MB | **1198 MB** | ~similar |

Pass 4 optimizations (buffer pre-grow, struct pooling, parallel zlib, template pool, etc.) improve the Zerodha mixed workload substantially while keeping PDF/A compliance.

---

## CPU pprof (5 runs, `-cpuprofile` during 5000 iter)

| Hotspot | Best | Worst | **Average** |
|---------|------|-------|-------------|
| **`memclrNoHeapPointers` (flat)** | 3.68% | 4.81% | **4.44%** |
| **`runtime.memmove` (flat)** | 6.84% | 7.30% | **7.00%** |
| **`GenerateTemplatePDF` (cum)** | 80.00% | 80.50% | **80.27%** |
| **`drawTable` (cum)** | 16.99% | 18.23% | **17.73%** |

Profiles: [baselines/zerodha_pprof_runs/cpu_zerodha_run{1..5}.prof](./baselines/zerodha_pprof_runs/)

```bash
go tool pprof -http=:8084 guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run4.prof
```

---

## pprof comparison across benchmarks (PDF/A)

| Benchmark | Workload | `memclr` flat | `drawTable` cum | Throughput |
|-----------|----------|---------------|-----------------|------------|
| **internal/pdf Rows2000** (micro, 5-run avg) | 2000-row table, 1 thread | ~4.5% | ~37% | ~36 ms/doc |
| **GoPDFLib data 5000×48** (7-run avg) | 2000-row + wrap | ~1.9% | ~40% | ~97 ops/s |
| **Zerodha 5000×48** (5-run avg) | 80/15/5 mix + signatures | **~4.4%** | **~17.7%** | **859 ops/s** |
| **HTTP load test** (pre-P4 baseline) | k6 mixed | **49.7%** | — | ~25 req/s |
| **HTTP load test** (Pass 4 PDF/A) | k6 tagged | **27.0%** | — | ~143 req/s |

Zerodha mix is **retail-heavy** (small 1-page PDFs with signing) → high throughput, lower `drawTable` share than pure data-table benches. HFT tail (5%) drives max latency spikes.

---

## Workload distribution (typical run)

~80% Retail (~4000), ~15% Active (~750), ~5% HFT (~240) per 5000 iterations.

Sample PDF outputs saved to `sampledata/gopdflib/zerodha/`:
- `zerodha_retail_output.pdf` (~61 KB)
- `zerodha_active_output.pdf` (~76 KB)
- `zerodha_hft_output.pdf` (~2.4 MB)

---

## Related

- [GOPDFLIB_PPROF_RESULTS.md](./GOPDFLIB_PPROF_RESULTS.md) — data-table 5000× bench
- [PASS4_PDFA_RESULTS.md](./PASS4_PDFA_RESULTS.md) — internal/pdf + HTTP load
- [sampledata/gopdflib/zerodha/benchmark_comparison.md](../sampledata/gopdflib/zerodha/benchmark_comparison.md) — vs Typst/Zerodha cluster
