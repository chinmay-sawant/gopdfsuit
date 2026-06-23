# Performance Benchmark: gopdflib vs Zerodha

This guide compares the performance of `gopdflib` against the industry benchmark set by Zerodha (a massive Indian brokerage) for high-scale PDF generation.

## 🏆 The "Zerodha" Comparison

Zerodha famously blogged about their technical achievement of generating, digitally signing, and emailing **1.5 million PDF contract notes in approximately 25 minutes**. This required a distributed cluster of machines orchestrated by Nomad.

With `gopdflib`, we achieved **58% of that total cluster throughput** on a **single machine**. By scaling to just 2 machines, `gopdflib` can match the entire 40-node cluster.

### Comparison Table

We ran a benchmark simulating Zerodha's exact weighted workload:

- **80% Retail**: 1-page PDF (2 trades)
- **15% Active Trader**: 2-3 page PDF (40 trades)
- **5% HFT/Algo**: 50+ page PDF (2,000 trades)

| Metric                    | Zerodha (Production Cluster)   | gopdflib (Single Node)   |
| :------------------------ | :----------------------------- | :----------------------- |
| **Throughput (peak)**     | **~1,000 PDFs/sec**            | **2,953 PDFs/sec** (30-run peak, Jun 14 2026) |
| **Throughput (avg)**      | N/A                            | **2,646 PDFs/sec** (30-run mean, same session) |
| **Infrastructure**        | **~40 Machines** (Distributed) | **1 Machine** (24 vCPUs) |
| **Efficiency (Per Core)** | ~1.6 PDFs/sec/core             | **~110 PDFs/sec/core** (at 2,646 avg) |
| **Latency (Avg)**         | N/A                            | **~18 ms** (ECDSA, PDF/UA-2) |
| **Time for 1.5M PDFs**    | 25 Minutes (40 nodes)          | **~9.4 min** (avg) / **~8.5 min** (peak) |

> **Historical:** Go 1.24-era runs reported **~586 PDFs/sec** mean / **637** peak. Current branch with `GeneratePDFBorrowed`, HFT fast path, and PDF/UA-2 TD hierarchy is **4×+** faster on the same hardware class.

## 💰 Cost Analysis

By moving from a heavy distributed architecture (CLI-based Typst/LaTeX) to a pure-Go in-memory architecture, the infrastructure requirements drop drastically.

**Scenario:** Generating 1.5 million PDFs (Daily Batch).

| Architecture           | Required Nodes  | Est. Hourly Cost (AWS) | Batch Cost (Daily) | Monthly Cost | Savings  |
| :--------------------- | :-------------- | :--------------------- | :----------------- | :----------- | :------- |
| **Zerodha (Typst)**    | ~40 Instances   | ~$24.50 / hr           | ~$10.20            | ~$306.00     | -        |
| **gopdflib (Go 1.24)** | **2 Instances** | **~$1.84 / hr**        | **~$0.77**         | **~$23.00**  | **~92%** |

> **📉 Result:** switching to `gopdflib` offers a **~92-95% reduction** in compute costs.

## 📊 Verification Calculation

- **Zerodha Benchmark:**
  - 1,500,000 PDFs / (25 minutes \* 60 seconds) = **1,000 PDFs/sec**

- **gopdflib Benchmark (Weighted Mix):**
  - Scenario: Real-world mix of small, medium, and massive (2000+ row) PDFs.
  - Result (2026-06-14): **2,646 PDFs/sec** (30-run avg), peak **2,953 PDFs/sec** on a single 24-core machine - **exceeds** the Zerodha cluster rate on one node.
  - Historical best (2026-06-13, idle): peak **3,604** / avg **2,787** ops/s.
  - Prior (2026-06-11): **2,459 PDFs/sec** (2-run avg).
  - Scaling: One node at **2,646 ops/s avg** (peak **2,953**) ≈ **2.6–3.0×** the 1,000 PDFs/sec cluster target.

## ⚡ Why is gopdflib so fast?

The reason `gopdflib` can match a distributed cluster's performance with 95% less hardware is due to its architectural design:

1.  **Native Binary Generation:** `gopdflib` generates the PDF binary structure directly in memory (RAM). It does not spawn external processes (like `typst`, `wkhtmltopdf`, or `pdflatex`) which incur heavy OS overhead for every single document.
2.  **Zero IO Overhead:** No temporary files are written to disk. Zerodha's architecture uploads/downloads files to S3 and writes temps to disk; `gopdflib` streams bytes directly.
3.  **Goroutine Concurrency:** Leveraging Go's lightweight goroutines allows us to saturate 24+ cores with thousands of concurrent generations without the memory overhead of OS threads.
4.  **Optimized for Scale:** Font subsets and image assets are processed once and reused across millions of documents, whereas CLI tools often reload assets for every invocation.

## 🚀 How to Run the Benchmark

You can reproduce these results using the Zerodha simulation in `sampledata/gopdflib/zerodha/main.go`.

```bash
cd sampledata/gopdflib/zerodha
go run main.go
```

### Benchmark Environment

- **CPU**: 13th Gen Intel(R) Core(TM) i7-13700HX (24 Logical Cores)
- **Go Version**: 1.24.0
- **Concurrency**: 48 Workers
