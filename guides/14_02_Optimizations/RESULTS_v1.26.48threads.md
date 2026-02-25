# GoPdfSuit Performance Latency Results (Go v1.26 + 48 Workers)

This report summarizes the performance results of the PDF generation API using the dynamic Zerodha workload (80% Retail, 15% Active Trader, 5% HFT) with a **concurrency limit of 48 workers**.

## Test Environment

- **Go Version:** 1.26
- **Architecture:** linux/amd64
- **Concurrency Limit:** 48 workers (goroutines)
- **Date:** 2026-02-14

## Benchmark Results (48 Workers)

| Test Scenario        | Total Req | Throughput (req/s) | Avg Latency | p95 Latency | p99 Latency | Status  |
| :------------------- | :-------- | :----------------- | :---------- | :---------- | :---------- | :------ |
| **Smoke & Load**     | 79        | 2.55               | 47.24ms     | 390.54ms    | 426.87ms    | ✅ PASS |
| **Spike (100 VUs)**  | 6,991     | 69.57              | 98.82ms     | 586.76ms    | 1,340.00ms  | ✅ PASS |
| **Soak (Sustained)** | 594       | 4.87               | 32.41ms     | 364.53ms    | 442.63ms    | ✅ PASS |

## Comparison with No Limit (v1.26)

| Scenario         | p99 (No Limit) | p99 (48 Workers) | Diff     |
| :--------------- | :------------- | :--------------- | :------- |
| **Smoke & Load** | 369.22ms       | 426.87ms         | +57.65ms |
| **Spike**        | 1,260.00ms     | 1,340.00ms       | +80.00ms |
| **Soak**         | 417.19ms       | 442.63ms         | +25.44ms |

## Observations

- **Stability:** The system remains stable under heavy load (Spike test) while maintaining a strict limit on concurrent goroutines.
- **Queuing Effect:** The slightly higher p99 latencies compared to the "No Limit" run are expected, as requests are queued at the middleware level once the 48-worker limit is reached.
- **Throughput:** Throughput remains nearly identical, indicating that the system was already processing most requests efficiently within the worker capacity, but the limit adds a safety buffer against resource exhaustion.
