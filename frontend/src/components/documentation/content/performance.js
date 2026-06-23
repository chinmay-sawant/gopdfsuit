export const performanceSection = {
    title: 'Performance',
    items: [
        {
            id: 'performance-overview',
            title: 'Benchmark Results',
            description: 'Measured benchmark results for GoPDFLib, GoPDFSuit, and pypdfsuit on the weighted 48-worker Zerodha workload, data-table, HTTP, and GoPDFKit compare harnesses.',
            codePlacement: 'below',
            content: `These measurements were captured on **June 2026** on WSL2 (Intel i7-13700HX, 24 logical CPUs) on branch \`feat/optimization-5.5-medium\`. Zerodha headline numbers for **GoPDFLib** and **pypdfsuit** use **x10 sequential runs** (\`make bench-gopdflib-zerodha-x10\`, \`make bench-pypdfsuit-zerodha-x10\`, and their \`-nocomply\` variants); other harnesses use **best-of-5** from \`guides/cursor/baselines/benchmark_run_20260618_v2/\`.

All compliant GoPDFLib headline numbers run with **PDF/A-4**, **PDF/UA-2**, Arlington-compatible tagging, XML metadata generation, **ECDSA P-256** digital signatures, embedded fonts, bookmarks, and internal links enabled.

**Machine Profile**

- **Kernel:** Linux 6.6.87.2-microsoft-standard-WSL2
- **CPU:** 13th Gen Intel(R) Core(TM) i7-13700HX
- **Topology:** 12 cores, 24 logical CPUs, 2 threads per core
- **Go:** 1.26.4 · **Python 3** · **k6** v1.4.2

## Zerodha Gold Standard (5000×48, PDF/A-4 + PDF/UA-2)

Primary end-to-end benchmark in **sampledata/gopdflib/zerodha** - **5000 iterations**, **48 workers**, 80% Retail / 15% Active / 5% HFT, with ECDSA P-256 retail signing. **Throughput values are aggregate system throughput** across 48 concurrent workers - not per-core serial throughput.

| Metric | x10 peak | x10 mean |
| --- | ---: | ---: |
| **Throughput** | **6,611 ops/sec** | **6,203 ops/sec** |
| **Avg latency** | **6.962 ms** | **7.544 ms** |
| **Min latency (best run)** | **0.305 ms** | - |
| **Max latency (best run)** | **205.857 ms** | - |
| **Peak allocated (mean)** | - | **798 MB** |

**vs June 2026 baseline (\`feat/performance-improvements\`):** **+150% x10 peak throughput** (2,646 → 6,611 ops/sec) under full PDF/A-4 + PDF/UA-2 compliance.

### x10 detail (compliant timing runs)

| Run | Throughput | Avg latency |
| ---: | ---: | ---: |
| 1 | 5,545 ops/sec | 8.417 ms |
| 2 | 6,069 ops/sec | 7.612 ms |
| 3 | 5,965 ops/sec | 7.687 ms |
| 4 | 6,492 ops/sec | 7.205 ms |
| 5 | 6,536 ops/sec | 7.115 ms |
| 6 | 5,578 ops/sec | 8.485 ms |
| 7 | 6,508 ops/sec | 7.162 ms |
| 8 | 6,408 ops/sec | 7.345 ms |
| 9 | 6,315 ops/sec | 7.447 ms |
| 10 | **6,611 ops/sec** | **6.962 ms** |

## Zerodha Gold Standard - Non-Compliant (5000×48)

Same 80/15/5 workload outputs **PDF 2.0** with PDF/A, tagging, signing, and font embedding disabled (\`make bench-gopdflib-zerodha-nocomply-x10\`). HFT output shrinks to **221 KB** (vs **2.3 MB** compliant).

| Metric | x10 peak | x10 mean |
| --- | ---: | ---: |
| **Throughput** | **37,853 ops/sec** | **34,035 ops/sec** |
| **Avg latency** | **1.227 ms** | **1.376 ms** |
| **Peak allocated (mean)** | - | **310 MB** |

## PyPDFSuit Zerodha Gold Standard (5000×48, PDF/A-4 + PDF/UA-2)

Same 80/15/5 workload via **pypdfsuit** CGO bindings (\`make bench-pypdfsuit-zerodha-x10\`). Rebuild the shared library first: \`cd bindings/python && ./build.sh\`. Honest full path: \`to_dict\` + \`json.dumps\` on every call (no JSON cache).

| Metric | x10 peak | x10 mean |
| --- | ---: | ---: |
| **Throughput** | **937 ops/sec** | **916 ops/sec** |
| **Avg latency** | **45.80 ms** | **46.93 ms** |
| **P50 latency (mean)** | **29.54 ms** | **28.79 ms** |
| **P95 latency (mean)** | **143.92 ms** | **150.45 ms** |
| **Min latency (best run)** | **0.86 ms** | - |
| **Max latency (worst run)** | **743.45 ms** | - |

**vs June 2026 best-of-5 baseline (235 ops/sec):** **+290% x10 mean throughput** after Python serializer optimizations.

### x10 detail (pypdfsuit compliant timing runs)

| Run | Throughput | Avg latency |
| ---: | ---: | ---: |
| 1 | 937 ops/sec | 45.94 ms |
| 2 | 893 ops/sec | 48.41 ms |
| 3 | **937 ops/sec** | 45.80 ms |
| 4 | 921 ops/sec | 45.74 ms |
| 5 | 928 ops/sec | 46.14 ms |
| 6 | 929 ops/sec | 46.19 ms |
| 7 | 915 ops/sec | 46.99 ms |
| 8 | 858 ops/sec | 49.56 ms |
| 9 | 928 ops/sec | 47.13 ms |
| 10 | 908 ops/sec | 47.40 ms |

## PyPDFSuit Zerodha Gold Standard - Non-Compliant (5000×48)

Same 80/15/5 workload via **pypdfsuit** with PDF/A, tagging, signing, and font embedding disabled (\`make bench-pypdfsuit-zerodha-nocomply-x10\`, \`BENCH_NOCOMPLY=1\`). HFT output shrinks to **336 KB** (vs **2.4 MB** compliant).

| Metric | x10 peak | x10 mean |
| --- | ---: | ---: |
| **Throughput** | **1,284 ops/sec** | **1,242 ops/sec** |
| **Avg latency** | **32.82 ms** | **33.78 ms** |
| **P50 latency (mean)** | **21.07 ms** | **22.61 ms** |
| **P95 latency (mean)** | **98.80 ms** | **103.46 ms** |
| **Min latency (best run)** | **0.23 ms** | - |
| **Max latency (worst run)** | **355.60 ms** | - |

**vs compliant pypdfsuit x10 peak (937 ops/sec):** non-compliant peak **1,284** - **~1.4×** faster (Python CGO overhead limits the nocomply ceiling vs native Go's 5.7×).

### x10 detail (pypdfsuit non-compliant timing runs)

| Run | Throughput | Avg latency |
| ---: | ---: | ---: |
| 1 | 1,217 ops/sec | 34.69 ms |
| 2 | **1,284 ops/sec** | **32.82 ms** |
| 3 | 1,243 ops/sec | 33.72 ms |
| 4 | 1,235 ops/sec | 34.11 ms |
| 5 | 1,251 ops/sec | 33.28 ms |
| 6 | 1,261 ops/sec | 33.08 ms |
| 7 | 1,250 ops/sec | 33.46 ms |
| 8 | 1,199 ops/sec | 34.58 ms |
| 9 | 1,239 ops/sec | 34.10 ms |
| 10 | 1,238 ops/sec | 33.94 ms |

### Weighted Workload - runtime comparison

| Runtime | Harness | Workers | Best Throughput | Avg Latency | PDF/A | PDF/UA |
| --- | --- | ---: | ---: | ---: | --- | --- |
| **GoPDFLib** | Weighted 80/15/5 (compliant) | 48 | **6,611 ops/sec** | **6.962 ms** | PDF/A-4 | PDF/UA-2 |
| **GoPDFLib** | Weighted 80/15/5 (nocomply) | 48 | **37,853 ops/sec** | **1.227 ms** | PDF 2.0 (no PDF/A) | None |
| **GoPDFSuit** | Retail only | 48 | **6,146 ops/sec** | **6.29 ms** | PDF/A-4 | PDF/UA-2 |
| **pypdfsuit** | Weighted 80/15/5 (compliant, x10 peak) | 48 | **937 ops/sec** | **45.80 ms** | PDF/A-4 | PDF/UA-2 |
| **pypdfsuit** | Weighted 80/15/5 (compliant, x10 mean) | 48 | **916 ops/sec** | **46.93 ms** | PDF/A-4 | PDF/UA-2 |
| **pypdfsuit** | Weighted 80/15/5 (nocomply, x10 peak) | 48 | **1,284 ops/sec** | **32.82 ms** | PDF 2.0 (no PDF/A) | None |
| **pypdfsuit** | Weighted 80/15/5 (nocomply, x10 mean) | 48 | **1,242 ops/sec** | **33.78 ms** | PDF 2.0 (no PDF/A) | None |
| **gpdf** | Weighted 80/15/5 (compliant) | 48 | **178 ops/sec** | **267.37 ms** | PDF/A-2b | None |
| **gpdf** | Weighted 80/15/5 (nocomply) | 48 | **464 ops/sec** | **100.64 ms** | PDF 2.0 (no PDF/A) | None |

## Data Table Benchmark (2000 rows)

Dataset: **sampledata/benchmarks/data.json** - 2,000 user records. Best-of-5 peak throughput.

| Rank | Engine | Workers | Best ops/s | Avg latency | PDF/A | PDF/UA |
| ---: | --- | ---: | ---: | ---: | --- | --- |
| 1 | **GoPDFLib** | 48 | **288** | ~156 ms | PDF/A-4 | PDF/UA-2 |
| 2 | PDFKit | 10 | 10.1 | ~721 ms | PDF 1.7 | None |
| 3 | jsPDF | 10 | 9.4 | ~946 ms | PDF 1.3 | None |
| 4 | pdf-lib | 10 | 6.0 | ~1,484 ms | PDF 1.7 | None |
| 5 | FPDF2 | 10 | 2.2 | ~4,492 ms | PDF 1.7 | None |
| 6 | Typst | 10 | 2.2 | ~549 ms | PDF 1.7 | None |

## HTTP Load Tests (k6)

End-to-end HTTP benchmarks via Makefile targets. Same compliance stack unless noted.

| Harness | VUs × duration | Best req/s | PDF/A | PDF/UA |
| --- | --- | ---: | --- | --- |
| \`bench-k6\` weighted (ECDSA) | 48 × 35s | **1,333** | PDF/A-4 | PDF/UA-2 |
| \`bench-k6-retail\` | 48 × 35s | **7,515** | PDF/A-4 | PDF/UA-2 |
| \`bench-k6-light\` | 24 × 15s | **1,177** | PDF/A-4 | PDF/UA-2 |
| \`bench-gotenberg\` (same harness) | 48 × 35s | 16.1 | None | None |

**gopdfsuit vs Gotenberg:** 1,333 / 16.1 ≈ **83× faster** on the weighted k6 harness.

## GoPDFKit vs GoPDFLib (apples-to-apples)

Harness: \`make bench-gopdfkit-compare\` in \`sampledata/benchmarks/gopdfkit_compare\`. Compares **GoPDFLib** (gopdfsuit engine) against external **GoPDFKit** v0.5.2 on equivalent **PDF 1.7** templates (no PDF/A flags). **40 workers**, \`benchtime=5s\`, best-of-5.

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

## Go test benchmarks (best-of-5)

| Benchmark | PDF/A | PDF/UA | Best ns/op | Best ops/s |
| --- | --- | --- | ---: | ---: |
| \`BenchmarkGenerateTemplatePDF_FinancialReport\` | PDF 1.7 | Arlington | **55,528** | 18,009 |
| \`BenchmarkGenerateTemplatePDF_FinancialReport_Parallel\` | PDF 1.7 | Arlington | **54,814** | 18,244 |
| \`bench-pdf-micro\` Rows2000 | PDF/A-4 | PDF/UA-2 | **11,934,163** | 84 |

**How to read this page:**
• Use the **Zerodha 5000×48 section** for realistic broker-mix concurrent throughput under full PDF/A-4 + PDF/UA-2 compliance.
• Use the **data-table section** for large-table PDF generation across libraries.
• Use **HTTP k6** for end-to-end API throughput including signing.
• Use **GoPDFKit compare** for library-level apples-to-apples speed (PDF 1.7 only - production gopdfsuit also supports PDF/A-4 + PDF/UA-2).
• **Do not** compare data-table serial ops/sec directly with 5000×48 aggregate ops/sec - they measure different concurrency models.`,
            code: {
                bash: `# Full benchmark suite (best-of-5, June 2026 harness)
cd /path/to/gopdfsuit
bash guides/cursor/baselines/benchmark_run_20260618_v2/continue_one_by_one.sh
python3 guides/cursor/baselines/benchmark_run_20260618_v2/parse_results.py

# Zerodha gold standard (5000 iterations, 48 workers)
make bench-gopdflib-zerodha-x10
make bench-gopdflib-zerodha-nocomply-x10

# Data table (2000 rows)
make bench-gopdflib-data

# GoPDFSuit retail-only Zerodha path
make bench-gopdfsuit-zerodha

# Python weighted Zerodha (rebuild bindings first)
cd bindings/python && ./build.sh && cd ../..
make bench-pypdfsuit-zerodha-x10
make bench-pypdfsuit-zerodha-nocomply-x10

# HTTP load tests
make bench-k6
make bench-k6-retail

# GoPDFKit vs GoPDFLib apples-to-apples compare
make bench-gopdfkit-setup
make bench-gopdfkit-compare`
            }
        }
    ]
};