# gopdfsuit k6 weighted benchmark - 5-run report (2026-06-14)

**Environment:** WSL2, i7-13700HX, 24 logical CPUs, `GOMAXPROCS=24`, Go 1.26.4  
**Harness:** `make bench-k6` → 48 VUs × 35s steady, `PAYLOAD_SCENARIO=tagged_ecdsa` (80% retail / 15% active / 5% HFT)  
**Config:** `MAX_CONCURRENT=48`, `BENCH_MODE=1`, `GIN_FAST_API=1`, PDF/A-4 + PDF/UA-2, retail ECDSA P-256  
**Runs:** 5 sequential back-to-back (`00:40`–`00:43` IST)

## Headline

| Metric | Value |
|--------|------:|
| **Peak** | **859 req/s** (run 2, `20260614_004108`) |
| **5-run avg** | **825 req/s** |
| **Worst** | 788 req/s (run 1) |
| **σ throughput** | 32 req/s |

**vs prior session (`20260613_215040`):** avg **716 → 825 req/s** (+15%).  
**vs historical peak (`20260611_190806`):** peak **1,054 req/s** - current peak is **81%** of best.

## Per-run detail

| Run | Run ID | Throughput | http med | http p99 | Iterations | Errors |
|-----|--------|----------:|---------:|---------:|-----------:|-------:|
| 1 | `20260614_004024` | 788 req/s | 16.3 ms | 370 ms | 27,848 | 0% |
| 2 | `20260614_004108` | **859 req/s** | 15.6 ms | 338 ms | 30,385 | 0% |
| 3 | `20260614_004149` | 797 req/s | 16.2 ms | 353 ms | 28,193 | 0% |
| 4 | `20260614_004230` | 851 req/s | 15.9 ms | 338 ms | 30,078 | 0% |
| 5 | `20260614_004311` | 830 req/s | 16.1 ms | 334 ms | 29,341 | 0% |
| **Mean** | - | **825 req/s** | **16.0 ms** | **347 ms** | **29,169** | 0% |

## Per-tier latency (custom k6 metrics, 5-run avg)

| Tier | avg / med / p99 |
|------|-----------------|
| **Retail (80%)** | 20 / 16 / 84 ms |
| **Active (15%)** | 28 / 23 / 107 ms |
| **HFT (5%)** | 456 / 449 / 752 ms |

## Artifacts

| Run | k6 log | pprof summary |
|-----|--------|---------------|
| 1 | `k6_gin_20260614_004024.txt` | `pprof_summary_20260614_004024.txt` |
| 2 | `k6_gin_20260614_004108.txt` | `pprof_summary_20260614_004108.txt` |
| 3 | `k6_gin_20260614_004149.txt` | `pprof_summary_20260614_004149.txt` |
| 4 | `k6_gin_20260614_004230.txt` | `pprof_summary_20260614_004230.txt` |
| 5 | `k6_gin_20260614_004311.txt` | `pprof_summary_20260614_004311.txt` |

Stats: [k6_gin_x5_20260614_stats.txt](./k6_gin_x5_20260614_stats.txt)

## Reproduce

```bash
for i in 1 2 3 4 5; do make bench-k6; done
```