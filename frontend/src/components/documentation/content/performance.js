export const performanceSection = {
    title: 'Performance',
    items: [
        {
            id: 'performance-overview',
            title: 'Benchmark Results',
            description: 'Measured benchmark results for GoPDFSuit, GoPDFLib, and pypdfsuit on a single Zerodha contract note and on the weighted 48-worker Zerodha workload.',
            codePlacement: 'below',
            content: `These measurements were captured on **May 25, 2026** on WSL2 (Intel i7-13700HX, 24 logical CPUs) by running the benchmark files checked into this repository.

Two benchmark modes were executed:

1. A **single Zerodha retail contract note** rendered repeatedly in-process by the runners under **sampledata/benchmarks** (48 iterations, serial comparison).
2. The **weighted Zerodha workload** in **sampledata/gopdflib/zerodha** — **5000 iterations**, **48 workers**, 80% Retail / 15% Active / 5% HFT, with **PDF/A + tagged PDF + digital signatures**.

**Machine Profile**

- **Kernel:** Linux 6.6.87.2-microsoft-standard-WSL2
- **CPU:** 13th Gen Intel(R) Core(TM) i7-13700HX
- **Topology:** 12 cores, 24 logical CPUs, 2 threads per core
- **Memory:** 7.6 GiB RAM
- **Go:** 1.24.0

## Zerodha Gold Standard (5000×48, PDF/A)

This is the primary end-to-end benchmark. **Throughput values are aggregate system throughput** across 48 concurrent workers — not per-core and not single-thread serial throughput.

| Metric | Peak observed | 10-run average |
| --- | ---: | ---: |
| **Throughput** | **2061.33 ops/sec** | **1704.95 ops/sec** |
| **Avg latency** | **22.680 ms** | **27.647 ms** |
| **Min latency** | **1.725 ms** | **1.967 ms** |
| **Max latency** | **659.165 ms** | **746.510 ms** |
| **Wall time (5000 docs)** | **2.426 s** | **2.952 s** |

**vs Go 1.24 historical (10-run avg):** **+197% throughput**, **66% lower avg latency**.

### 10-run detail (sequential WSL runs)

| Run | Throughput | Avg | Wall (s) |
| ---: | ---: | ---: | ---: |
| 1 | 1634.72 ops/sec | 28.540 ms | 3.059 |
| 2 | 1741.68 ops/sec | 26.742 ms | 2.871 |
| 3 | 1756.62 ops/sec | 26.707 ms | 2.846 |
| 4 | 1613.58 ops/sec | 29.070 ms | 3.099 |
| 5 | 1701.42 ops/sec | 27.799 ms | 2.939 |
| 6 | 1542.39 ops/sec | 30.467 ms | 3.242 |
| 7 | 1601.74 ops/sec | 29.150 ms | 3.122 |
| 8 | 1557.89 ops/sec | 30.050 ms | 3.209 |
| 9 | 2020.59 ops/sec | 22.988 ms | 2.475 |
| 10 | 1878.83 ops/sec | 24.958 ms | 2.661 |

Run-to-run variance is driven by system load and random HFT draw (~209–275 HFT docs per 5000).

## Single Zerodha Retail Contract Note (48 iterations)

This section measures the same retail contract-note document rendered serially in-process. The **ops/sec** values here are **single-instance serial throughput**, not multi-worker throughput.

| Library | Runtime | Best Observed Time | Peak Serial Throughput |
| --- | --- | ---: | ---: |
| **GoPDFLib** | Go | **2.48 ms** | **306.05 ops/sec** |
| **GoPDFSuit** | Go | 2.87 ms | 243.00 ops/sec |
| **pypdfsuit** | Python bindings | 3.05 ms | 211.51 ops/sec |

### Weighted Workload — PyPDFSuit comparison

| Runtime | Iterations | Workers | Best Throughput | Avg Latency |
| --- | ---: | ---: | ---: | ---: |
| **GoPDFLib** | 5000 | 48 | **2061.33 ops/sec** | **22.680 ms** |
| **pypdfsuit** | 5000 | 48 | 234.62 ops/sec | 178.338 ms |

## GoPDFKit vs GoPDFLib (apples-to-apples)

Harness: \`make bench-gopdfkit-compare\` in \`sampledata/benchmarks/gopdfkit_compare\`. Compares **GoPDFLib** (gopdfsuit engine) against external **GoPDFKit** v0.5.2 on equivalent PDF 1.7 templates (no PDF/A flags). **40 workers**, \`benchtime=5s\`, best-of-5 from June 2026 suite (i7-13700HX, 24 logical CPUs).

| Workload | GoPDFKit pdf/s | GoPDFLib pdf/s | gopdflib lead |
| --- | ---: | ---: | ---: |
| text_short | 119,959 | **254,986** | **2.1x** |
| text_240_lines | 14,755 | **32,453** | **2.2x** |
| table_180_rows | 11,883 | **47,707** | **4.0x** |
| table_900_rows | 2,635 | **10,452** | **4.0x** |
| invoice_40_rows | 40,145 | **135,052** | **3.4x** |
| png_table_180_rows | 7,504 | **45,098** | **6.0x** |
| png_rows_60 | 5,474 | **53,935** | **9.9x** |

**Result:** gopdflib wins **7/7** workloads. Lead ranges **2.1x** (text) to **9.9x** (PNG rows).

**How to read this page:**
• Use the **gold standard 5000×48 section** for realistic broker-mix concurrent throughput under PDF/A compliance.
• Use the **single retail contract-note section** to compare per-document render speed on the same template across runtimes.
• Use **GoPDFKit compare** for library-level apples-to-apples speed (PDF 1.7 only — production gopdfsuit also supports PDF/A-4 + PDF/UA-2).
• **Do not** compare serial ops/sec directly with 5000×48 aggregate ops/sec — they measure different concurrency models.`,
            code: {
                bash: `# Zerodha gold standard (5000 iterations, 48 workers)
cd sampledata/gopdflib/zerodha && go run .

# 10 sequential timing runs
bash sampledata/gopdflib/zerodha/run_bench_x10.sh

# GoPDFKit vs GoPDFLib apples-to-apples compare
make bench-gopdfkit-setup
make bench-gopdfkit-compare

# Single-document Zerodha retail benchmark runners
cd sampledata/benchmarks/gopdfsuit && go run bench.go
cd sampledata/benchmarks/gopdflib && go run bench.go

# Python benchmark runners
cd ./gopdfsuit
.venv/bin/python sampledata/benchmarks/pypdfsuit/bench.py
.venv/bin/python sampledata/gopdflib/zerodha/pypdfsuit_bench.py`
            }
        }
    ]
};
