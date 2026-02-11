# Zerodha Gold Standard Benchmark Results

This benchmark simulates a real-world brokerage workload with a weighted mix of contract notes:

- **Retail Investor (80%)**: Single-page, digital signature, PDF/A-4.
- **Active Trader (15%)**: 2-3 pages, multi-page tables, bookmarks.
- **HFT / Algo (5%)**: 50+ pages, 2000 trades, stress test.

## Execution Summary

- **Date**: 2026-02-12
- **Hardware**: Linux (amd64), 24 CPUs
- **Go Version**: go1.24.11
- **Iterations**: 5,000 (Sample run)
- **Concurrency**: 48 workers

## Key Metrics

| Metric          | Result             | Target       | Status      |
| --------------- | ------------------ | ------------ | ----------- |
| **Throughput**  | **496.06 ops/sec** | >300 ops/sec | ✅ **PASS** |
| **Avg Latency** | 93.49 ms           | <100 ms      | ✅ **PASS** |
| **Max Memory**  | 996.57 MB          | <2 GB        | ✅ **PASS** |

> **Note**: The benchmark maintained ~500 operations per second even with 5% of the workload being massive 50-page HFT documents. This confirms the engine's capability to handle mixed, high-stress workloads efficiently.

## Detailed Output

```text
=== Zerodha Gold Standard Benchmark ===
Workload Mix: 80% Retail | 15% Active | 5% HFT

OS: linux, Arch: amd64, NumCPU: 24, GoVersion: go1.24.11
Running 5000 iterations using 48 workers...

Building templates...
Templates built.
Warm-up runs...
  Retail PDF size:  55296 bytes (54.00 KB)
  Active PDF size:  75117 bytes (73.36 KB)
  HFT PDF size:     2379919 bytes (2324.14 KB)

  Max Memory Allocated: 996.57 MB
=== Performance Summary ===
  Iterations:      5000
  Concurrency:     48 workers
  Total time:      10.079 s
  Throughput:      496.06 ops/sec

  Avg Latency:     93.494 ms
  Min Latency:     4.038 ms
  Max Latency:     2153.693 ms

=== Workload Distribution ===
  Retail  (80%):   4017 iterations
  Active  (15%):   732 iterations
  HFT      (5%):   251 iterations

Saved: zerodha_retail_output.pdf (55296 bytes)
Saved: zerodha_active_output.pdf (75117 bytes)
Saved: zerodha_hft_output.pdf (2379919 bytes)

=== Done ===
```

## Generated Artifacts

- `zerodha_retail_output.pdf` (54 KB) - Standard signed contract note
- `zerodha_active_output.pdf` (73 KB) - Multi-page report with bookmarks
- `zerodha_hft_output.pdf` (2.3 MB) - Large volume trade report
