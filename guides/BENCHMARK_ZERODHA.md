# Performance Benchmark: gopdflib vs Zerodha

This guide compares the performance of `gopdflib` against the industry benchmark set by Zerodha (a massive Indian brokerage) for high-scale PDF generation.

## 🏆 The "Zerodha" Comparison

Zerodha famously blogged about their technical achievement of generating, digitally signing, and emailing **1.5 million PDF contract notes in approximately 25 minutes**. This required a distributed cluster of machines orchestrated by Nomad.

With `gopdflib`, we achieved similar throughput on a **single machine**.

### Comparison Table

| Metric                          | Zerodha (Distributed Cluster) | gopdflib (Single Machine)       |
| :------------------------------ | :---------------------------- | :------------------------------ |
| **Throughput**                  | **~1,000 PDFs/sec**           | **1,113 PDFs/sec**              |
| **Total Volume (extrapolated)** | 1.5 Million PDFs / 25 min     | **1.67 Million PDFs / 25 min**  |
| **Infrastructure**              | **Cluster of 40 machines**    | **Intel Core i7-13700HX**       |
| **Technology**                  | LaTeX / Typst / Distributed   | Pure Go / Native PDF Generation |
| **CPU Details**                 | Distributed EC2 Fleet         | 13th Gen i7 (24 Threads)        |

## 📊 Verification Calculation

- **Zerodha Benchmark:**
  - 1,500,000 PDFs / (25 minutes \* 60 seconds) = **1,000 PDFs/sec**
- **gopdflib Benchmark (Actual Results):**
  - 100,000 PDFs / 89.77 seconds = **1,113.90 PDFs/sec**

## ⚡ Why is gopdflib so fast?

The reason `gopdflib` can match a distributed cluster's performance on a single machine is due to its architectural design:

1.  **Native Binary Generation:** `gopdflib` generates the PDF binary structure directly in memory. It does not rely on heavy dependencies like Puppeteer (headless browsers) or complex typesetting systems like LaTeX.
2.  **Concurrency Model:** Leveraging Go's lightweight goroutines and a worker pool pattern allows it to saturate all available CPU cores with minimal context-switching overhead.
3.  **Thread-Safe Isolated Registry:** Each generation session uses a cloned font registry, preventing mutex contention during high-concurrency workloads.
4.  **Zero-Allocation Focus:** Reusing buffers and optimizing font subsetting reduces Garbage Collection (GC) pauses, which is critical for sustained high throughput.

## 🚀 How to Run the Benchmark

You can reproduce these results using the example code in `sampledata/gopdflib/financial_report/main.go`.

```bash
cd sampledata/gopdflib/financial_report
go run main.go
```

### Benchmark Environment

- **CPU**: 13th Gen Intel(R) Core(TM) i7-13700HX (24 Logical Cores)
- **Memory**: 8 GB RAM
- **OS**: Ubuntu (WSL2) on Windows 11

### Typical Output (Measured on i7-13700HX)

```text
=== gopdflib Financial Report Example ===

Running 100000 iterations using 48 workers...

=== Performance Summary ===
  Iterations:    100000
  Throughput:    1113.90 ops/sec
  Total time:    89.774 s
  Avg Latency:   43.082 ms
  PDF size:      131.07 KB
```
