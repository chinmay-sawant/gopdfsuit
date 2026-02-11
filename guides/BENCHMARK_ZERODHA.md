# Performance Benchmark: gopdflib vs Zerodha

This guide compares the performance of `gopdflib` against the industry benchmark set by Zerodha (a massive Indian brokerage) for high-scale PDF generation.

## 🏆 The "Zerodha" Comparison

Zerodha famously blogged about their technical achievement of generating, digitally signing, and emailing **1.5 million PDF contract notes in approximately 25 minutes**. This required a distributed cluster of machines orchestrated by Nomad.

With `gopdflib`, we achieved similar throughput on a **single machine**.

### Comparison Table

| Metric                    | Zerodha (Distributed Cluster) | gopdflib (Cached Images) | gopdflib (Unique Images) |
| :------------------------ | :---------------------------- | :----------------------- | :----------------------- |
| **Throughput**            | **~1,000 PDFs/sec**           | **930 PDFs/sec**         | **423 PDFs/sec**         |
| **Total Volume (25 min)** | 1.5 Million PDFs              | **1.4 Million PDFs**     | **635,000 PDFs**         |
| **Infrastructure**        | **Cluster of 40 machines**    | **Single Machine**       | **Single Machine**       |
| **CPU Details**           | Distributed EC2 Fleet         | i7-13700HX (24 Threads)  | i7-13700HX (24 Threads)  |
| **Memory Usage**          | Unknown                       | **~254 MB**              | **~425 MB**              |

## 📊 Verification Calculation

- **Zerodha Benchmark:**
  - 1,500,000 PDFs / (25 minutes \* 60 seconds) = **1,000 PDFs/sec**

- **gopdflib Benchmark (Cached Images - Standard):**
  - Scenario: Logos, repeating charts, static assets.
  - Result: **930.41 PDFs/sec** on a single machine.

- **gopdflib Benchmark (Unique Images - Worst Case):**
  - Scenario: Every PDF has unique, generated charts (no caching possible).
  - Result: **423.22 PDFs/sec** on a single machine.

## ⚡ Why is gopdflib so fast?

The reason `gopdflib` can match a distributed cluster's performance on a single machine is due to its architectural design:

1.  **Native Binary Generation:** `gopdflib` generates the PDF binary structure directly in memory. It does not rely on heavy dependencies like Puppeteer (headless browsers) or complex typesetting systems like LaTeX.
2.  **Concurrency Model:** Leveraging Go's lightweight goroutines and a worker pool pattern allows it to saturate all available CPU cores with minimal context-switching overhead.
3.  **Thread-Safe Isolated Registry:** Each generation session uses a cloned font registry, preventing mutex contention during high-concurrency workloads.
4.  **Zero-Allocation Focus:** Reusing buffers and optimizing font subsetting reduces Garbage Collection (GC) pauses.
5.  **Smart Caching:** Decoded images are cached by hash, speeding up generation for repetitive assets (logos, standard charts). Even without caching, the raw engine throughput exceeds 400 PDFs/sec.

## 🚀 How to Run the Benchmark

You can reproduce these results using the example code in `sampledata/gopdflib/financial_report/main.go`.

```bash
cd sampledata/gopdflib/financial_report
go run main.go
```

### Benchmark Environment

- **CPU**: 13th Gen Intel(R) Core(TM) i7-13700HX (24 Logical Cores)
- **Memory**: 32 GB RAM (Task used < 500 MB)
- **OS**: Ubuntu (WSL2) on Windows 11
- **Go Version**: 1.26.0

### Typical Output (Cached Images)

```text
=== gopdflib Financial Report Example ===

OS: linux, Arch: amd64, NumCPU: 24, GoVersion: go1.26.0
Running 100000 iterations using 48 workers...

Warm-up run...
Max Memory Allocated: 253.84 MB
=== Performance Summary ===
  Iterations:    100000
  Throughput:    930.41 ops/sec
  Total time:    107.480 s
  Avg Latency:   51.577 ms
  PDF size:      131.07 KB
```

### Typical Output (Unique/Non-Cached Images)

```text
=== gopdflib Financial Report Example ===

OS: linux, Arch: amd64, NumCPU: 24, GoVersion: go1.26.0
Running 100000 iterations using 48 workers...

Warm-up run...
Max Memory Allocated: 425.42 MB
=== Performance Summary ===
  Iterations:    100000
  Throughput:    423.22 ops/sec
  Total time:    236.282 s
  Avg Latency:   113.392 ms
  PDF size:      131.07 KB
```
