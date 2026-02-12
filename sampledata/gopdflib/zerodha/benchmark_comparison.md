# GoPdfLib (Go 1.24) vs Zerodha (Typst) Benchmark Comparison

## Executive Summary

This report compares the performance of **gopdflib (Go 1.24)** against **Zerodha's Typst-based architecture** for generating financial contract notes.
The benchmark uses the same workload mix as Zerodha's production system (80% Retail, 15% Active Trader, 5% HFT/Algo).

**Key Finding:** `gopdflib` running on a **single 24-core machine** achieves **over 57% of the total throughput** of Zerodha's entire ~40-node production cluster.

## 1. Test Environment

### System Specifications

- **Hardware**: Linux (WSL2), AMD64, 24 Logical Cores
- **Go Version**: 1.24.0 (Stable)
- **Concurrency**: 48 Workers

### Workload Mix

- **80% Retail**: 1-page PDF (2 trades)
- **15% Active**: 2-3 page PDF (40 trades)
- **5% HFT**: 50+ page PDF (2,000 trades)

## 2. Performance Results

| Metric                 | GoPdfLib (Single Node) | Zerodha Cluster (Production) | Difference                              |
| :--------------------- | :--------------------- | :--------------------------- | :-------------------------------------- |
| **Throughput (Mean)**  | **573.64 PDFs/sec**    | ~1,000 PDFs/sec              | Single node = **~57%** of total cluster |
| **Throughput (Max)**   | **637.87 PDFs/sec**    | N/A                          | Peak performance                        |
| **Time for 1.5M PDFs** | **~44 minutes**        | ~25 minutes                  | One server vs Forty servers             |
| **Avg Latency**        | **80.67 ms**           | N/A                          | End-to-end generation time              |
| **Min Latency**        | **71.88 ms**           | N/A                          | Best case latency                       |

## 3. Efficiency Analysis (Per Core)

To make a fair comparison, we normalize the throughput based on estimated compute resources.

|                      | ZeroDha (est.)            | GoPdfLib (Actual)        |
| :------------------- | :------------------------ | :----------------------- |
| **Total Nodes**      | ~40 Instances             | 1 Instance               |
| **Total vCPUs**      | ~640 vCPUs (40 \* 16 avg) | 24 vCPUs                 |
| **Total Throughput** | 1,000 PDFs/sec            | 573 PDFs/sec             |
| **Efficiency**       | **~1.56 PDFs/sec/core**   | **~23.90 PDFs/sec/core** |

> **Result:** `gopdflib` is approximately **15x more efficient per CPU core** than the Typst-based solution.

## 4. Why GoPdfLib Wins?

1. **In-Memory Generation**: Pure Go architecture avoids the overhead of spawning external processes (CLI) like `typst` or `pdflatex` for every single document.
2. **Zero IO Overhead**: No temporary files are written to disk during generation; everything happens in RAM streams.
3. **Goroutine Concurrency**: Go's lightweight thread scheduler handles thousands of concurrent generations more efficiently than OS-process based workers.
4. **Optimized for Scale**: Dedicated PDF generation libraries can reuse font objects and image assets across documents, whereas CLI tools must reload them for every invocation.

---

_Data based on local benchmarks run on 2026-02-12 and Zerodha engineering blog "1.5+ million PDFs in 25 minutes"._
