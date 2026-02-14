# GoPdfSuit Performance Latency Results (Go v1.26)

This report summarizes the performance results of the PDF generation API using the dynamic Zerodha workload (80% Retail, 15% Active Trader, 5% HFT).

## Test Environment

- **Go Version:** 1.26
- **Architecture:** linux/amd64
- **Date:** 2026-02-14

## Benchmark Results

| Test Scenario        | Total Req | Throughput (req/s) | Avg Latency | p95 Latency | p99 Latency | Status  |
| :------------------- | :-------- | :----------------- | :---------- | :---------- | :---------- | :------ |
| **Smoke & Load**     | 81        | 2.68               | 16.96ms     | 11.36ms     | 369.22ms    | ✅ PASS |
| **Spike (100 VUs)**  | 6,977     | 69.54              | 98.65ms     | 727.96ms    | 1,260.00ms  | ✅ PASS |
| **Soak (Sustained)** | 597       | 4.90               | 25.63ms     | 25.15ms     | 417.19ms    | ✅ PASS |

## Observations

- **p99 Characteristics:** The p99 latency spikes significantly during high concurrent load (Spike test) as the larger HFT payloads (2000 trades, 50+ pages) compete for processing resources.
- **Resource Usage:** The system handled up to 70 requests/second during the spike test without generating errors.
