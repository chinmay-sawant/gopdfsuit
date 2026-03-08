export const performanceSection = {
    title: 'Performance',
    items: [
        {
            id: 'performance-overview',
            title: 'Benchmark Results',
            description: 'Measured benchmark results for GoPDFSuit, GoPDFLib, and pypdfsuit on a single Zerodha contract note and on the weighted 48-worker Zerodha workload.',
            codePlacement: 'below',
            content: `These measurements were captured on **March 9, 2026** by running the benchmark files checked into this repository and keeping the best observed result shown below.

Two benchmark modes were executed:

1. A **single Zerodha retail contract note** rendered repeatedly in-process by the runners under **sampledata/benchmarks**.
2. The **weighted Zerodha workload** in **sampledata/gopdflib/zerodha** that mixes retail, active trader, and HFT-style contract notes across 48 workers.

**Machine Profile**

- **Kernel:** Linux 6.6.87.2-microsoft-standard-WSL2
- **CPU:** 13th Gen Intel(R) Core(TM) i7-13700HX
- **Topology:** 12 cores, 24 logical CPUs, 2 threads per core
- **Memory:** 7.6 GiB RAM

## Single Zerodha Retail Contract Note

This section measures the same retail contract-note document rendered serially in-process. The **ops/sec** values here are **single-instance serial throughput**, not multi-worker throughput.

| Library | Runtime | Best Observed Time | Peak Serial Throughput |
| --- | --- | ---: | ---: |
| **GoPDFLib** | Go | **2.48 ms** | **306.05 ops/sec** |
| **GoPDFSuit** | Go | 2.87 ms | 243.00 ops/sec |
| **pypdfsuit** | Python bindings | 3.05 ms | 211.51 ops/sec |

| Ranking | Observation |
| --- | --- |
| 1 | **GoPDFLib** posts the fastest single-document retail render in the current local run. |
| 2 | **GoPDFSuit** remains close on the same contract-note template and stays in the same performance tier. |
| 3 | **pypdfsuit** is slower on the serial single-document pass, but remains within the same order of magnitude on the exact same retail template. |

### Zerodha Weighted Workload

This workload keeps the realistic retail-heavy mix and reports the strongest observed 48-worker pass for each runtime. These throughput values are aggregate system throughput and should not be compared directly with the single-document serial throughput above.

| Runtime | Iterations | Workers | Best Throughput | Avg Latency | Min Latency | Max Latency | Retail / Active / HFT |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| **GoPDFLib** | 5000 | 48 | **1913.13 ops/sec** | **24.558 ms** | 2.280 ms | 505.087 ms | 4004 / 766 / 230 |
| **pypdfsuit** | 5000 | 48 | 233.76 ops/sec | 185.517 ms | 2.657 ms | 3516.474 ms | 4015 / 767 / 218 |

**Benchmark Notes:**
• **GoPDFLib** and **GoPDFSuit** single-document benchmarks now execute a **single Zerodha retail contract note** instead of the old generic user-report dataset.
• **pypdfsuit** now uses the same retail contract-note template for its single-document benchmark under **sampledata/benchmarks/pypdfsuit/bench.py**.
• A missing **pypdfsuit** Zerodha benchmark runner was added at **sampledata/gopdflib/zerodha/pypdfsuit_bench.py** and executed against the weighted retail/active/HFT workload.
• The single-document benchmark values shown here come from repeated in-process renders of the same retail contract note and no longer show output-size data.
• Both Zerodha runners were executed with **48 workers** for the current comparison.

**How to read this page:**
• Use the **single retail contract-note section** to compare per-document render speed on the same Zerodha template.
• Use the **weighted Zerodha workload section** to compare end-to-end concurrent throughput under a realistic broker mix.
• Use **pypdfsuit** when Python ergonomics matter and the measured binding overhead is acceptable for the target workload.`,
            code: {
                bash: `# Single-document Zerodha retail benchmark runners
cd sampledata/benchmarks/gopdfsuit && go run bench.go
cd sampledata/benchmarks/gopdflib && go run bench.go

# Dedicated Python benchmark runners
cd /home/chinmay/ChinmayPersonalProjects/gopdfsuit
/home/chinmay/ChinmayPersonalProjects/gopdfsuit/.venv/bin/python sampledata/benchmarks/pypdfsuit/bench.py
/home/chinmay/ChinmayPersonalProjects/gopdfsuit/.venv/bin/python sampledata/gopdflib/zerodha/pypdfsuit_bench.py

# Existing Zerodha Go runner
go run sampledata/gopdflib/zerodha/main.go`
            }
        }
    ]
};