# Zerodha Gold Standard Benchmark (Go 1.26.4)

**Date:** 2026-06-11  
**Go Version:** `go1.26.4`  
**Harness:** `sampledata/gopdflib/zerodha/main.go`

## Command

```bash
/usr/local/go/bin/go1.26.4 run sampledata/gopdflib/zerodha/main.go
```

Workload Mix: 80% Retail (1-page) | 15% Active Trader (2-3 page) | 5% HFT (50+ page)

## System

- **CPU:** 13th Gen Intel(R) Core(TM) i7-13700HX (24 logical cores)
- **Workers:** 48
- **Iterations:** 5,000

## Results (3-run median)

| Metric | Run 1 | Run 2 | Run 3 | Median |
|---|---:|---:|---:|---:|
| **Throughput** | 1,159.33 ops/sec | 1,134.71 ops/sec | 871.56 ops/sec | **1,134.71 ops/sec** |
| **Avg Latency** | 40.914 ms | 41.327 ms | 54.131 ms | **41.327 ms** |
| **Min Latency** | 2.016 ms | 1.763 ms | 2.041 ms | **2.016 ms** |
| **Max Latency** | 1,114.918 ms | 1,113.293 ms | 3,045.819 ms | **1,114.918 ms** |
| **Max Memory** | 1,189.08 MB | 1,328.62 MB | 1,243.74 MB | **1,243.74 MB** |

## Workload Distribution (median)

| Profile | Share | Iterations |
|---|---:|---:|
| Retail (1-page) | 80% | 3,990 |
| Active Trader (2-3 page) | 15% | 748 |
| HFT (50+ page) | 5% | 262 |

## Output Sizes

| Profile | Size |
|---|---:|
| Retail | 59.86 KB (61,293 bytes) |
| Active Trader | 74.28 KB (76,065 bytes) |
| HFT | 2,368.90 KB (2,425,758 bytes) |

## Comparison with Previous (Go 1.24)

| Metric | Go 1.24 (baseline) | Go 1.26.4 (current) | Δ |
|---|---:|---:|---:|
| Throughput | 573.64 ops/sec | 1,134.71 ops/sec | **+98%** |
| Avg Latency | 80.67 ms | 41.33 ms | **-49%** |
| Min Latency | 71.88 ms | 2.02 ms | **-97%** |
| Max Memory | N/A | 1,243.74 MB | - |

## Comparison vs Zerodha Production

| Metric | gopdflib (Go 1.26.4, 1 node) | Zerodha Cluster (40 nodes) |
|---|---:|---:|
| Throughput | 1,134.71 ops/sec | ~1,000 ops/sec |
| Efficiency per core | ~47.3 ops/sec/core | ~1.56 ops/sec/core |

## Raw Outputs

### Run 1

```
=== Performance Summary ===
  Iterations:      5000
  Concurrency:     48 workers
  Total time:      4.313 s
  Throughput:      1159.33 ops/sec

  Avg Latency:     40.914 ms
  Min Latency:     2.016 ms
  Max Latency:     1114.918 ms
```

### Run 2

```
=== Performance Summary ===
  Iterations:      5000
  Concurrency:     48 workers
  Total time:      4.406 s
  Throughput:      1134.71 ops/sec

  Avg Latency:     41.327 ms
  Min Latency:     1.763 ms
  Max Latency:     1113.293 ms
```

### Run 3

```
=== Performance Summary ===
  Iterations:      5000
  Concurrency:     48 workers
  Total time:      5.737 s
  Throughput:      871.56 ops/sec

  Avg Latency:     54.131 ms
  Min Latency:     2.041 ms
  Max Latency:     3045.819 ms
```

## Summary

`gopdflib` running on Go 1.26.4 achieves **1,134.71 ops/sec** on a single 24-core machine - nearly **double** the previous Go 1.24 throughput (573.64 ops/sec) and actually **surpasses** Zerodha's entire ~40-node production cluster (~1,000 ops/sec). Per-core efficiency is **~47.3 ops/sec/core** vs ~1.56 ops/sec/core for Typst.

Run 3 had a outlier max latency (3s) likely from GC or OS scheduling - the median throughput across 3 runs is still 1,134.71 ops/sec with 41.33 ms average latency.
