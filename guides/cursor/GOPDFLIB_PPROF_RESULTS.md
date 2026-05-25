# GoPDFLib — 5000× pprof Benchmark Results (PDF/A)

**Date:** 2026-05-25  
**Entry point:** `sampledata/benchmarks/gopdflib/` — `go run -mod=mod . data`  
**Workload:** 2000-row data table (`data.json`), PDF/A + tagged PDF, wrap on Email/Description  
**Concurrency:** 5000 iterations, 48 worker goroutines  

---

## How to run

```bash
cd sampledata/benchmarks/gopdflib

# Full suite: 1 timing + 5 CPU profiles + 1 heap (5000 iter each)
bash run_pprof_bench.sh

# Single run (custom iterations)
BENCH_ITERATIONS=5000 BENCH_WORKERS=48 BENCH_QUIET=1 GOWORK=off \
  go build -mod=mod -o gopdflib_bench . && ./gopdflib_bench data

# With CPU profile
./gopdflib_bench -cpuprofile=cpu.prof data
./gopdflib_bench -memprofile=heap.prof data
```

Environment variables: `BENCH_ITERATIONS` (default **5000**), `BENCH_WORKERS` (default **48**), `BENCH_QUIET=1` (progress every 500 runs).

---

## Timing results (7 executions)

Each run generates **5000 PDFs** with 48 concurrent workers.

| Run | Avg (ms) | Min (ms) | P95 (ms) | Max (ms) | Throughput |
|-----|----------|----------|----------|----------|------------|
| **Timing (no profile)** | **389.54** | **135.65** | **643.43** | 1240.96 | **121.69 ops/s** |
| CPU profile 1 | 420.79 | 152.45 | 677.43 | 1269.58 | 112.59 |
| CPU profile 2 | **583.86** | 180.24 | **938.63** | **1793.26** | **80.96** |
| CPU profile 3 | 571.11 | 204.18 | 905.65 | 1641.44 | 83.16 |
| CPU profile 4 | 527.25 | 186.87 | 814.18 | 1481.23 | 90.08 |
| CPU profile 5 | 514.73 | 191.26 | 806.84 | 1330.53 | 92.46 |
| Heap profile | 483.26 | 185.35 | 762.09 | 1500.60 | 98.43 |

### Aggregate (7 runs)

| Metric | Best | Worst | **Average** | σ |
|--------|------|-------|-------------|---|
| **Avg latency** | **389.54 ms** | **583.86 ms** | **498.65 ms** | 72.79 ms |
| **Throughput** | **121.69 ops/s** | **80.96 ops/s** | **~97 ops/s** | — |
| **P95 latency** | **643.43 ms** | **938.63 ms** | — | — |
| **Max memory** | 1713 MB | 1902 MB | ~1800 MB peak | — |

**Best run:** timing-only (`bench_gopdflib_5000_timing.txt`) — **389.54 ms avg**, **121.69 ops/s**  
**Worst run:** CPU profile run 2 — **583.86 ms avg**, **80.96 ops/s**, **1793 ms max**

---

## CPU pprof (5 runs during 5000× generation)

| Hotspot | Best | Worst | **Average** |
|---------|------|-------|-------------|
| **`memclrNoHeapPointers` (flat)** | **1.77%** | **2.03%** | **1.88%** |
| **`runtime.memmove` (flat)** | **5.34%** | **5.70%** | **5.49%** |
| **`GenerateTemplatePDF` (cum)** | **71.64%** | **72.23%** | **71.90%** |
| **`drawTable` (cum)** | **39.37%** | **40.34%** | **39.89%** |
| **`BeginMarkedContentBuf` (cum)** | **6.68%** | **6.92%** | **6.82%** |

Individual `memclr` runs (%): `2.03, 1.92, 1.90, 1.77, 1.79`

**Profiles:** [baselines/gopdflib_pprof_runs/](./baselines/gopdflib_pprof_runs/)

```bash
go tool pprof -http=:8083 guides/cursor/baselines/gopdflib_pprof_runs/cpu_gopdflib_data_run1.prof
```

---

## Heap profile (after 5000 iterations)

| Hotspot | In-use |
|---------|--------|
| **Total** | **751.71 MB** |
| **`bytes.growSlice`** | **443.40 MB** (59%) |
| **`GenerateTemplatePDF` (cum)** | 642.64 MB |
| **`compress/flate.NewWriter`** | 88.34 MB cum |
| **`BeginMarkedContentBuf`** | 63.90 MB cum (26.89 MB flat) |

Profile: [baselines/gopdflib_pprof_runs/heap_gopdflib_data.prof](./baselines/gopdflib_pprof_runs/heap_gopdflib_data.prof)

---

## Comparison with `internal/pdf` pprof (Pass 4 PDF/A)

Different workloads — compare trends, not absolute numbers.

| Context | Workload | Avg time | `memclr` CPU | `drawTable` cum | Heap pressure |
|---------|----------|----------|--------------|-----------------|---------------|
| **`internal/pdf` micro-bench** | 2000 rows, no wrap, 1 goroutine | **~36 ms** (5-run avg) | **~4.5%** flat | **~37%** | N/A (single iter) |
| **`internal/pdf` HTTP load** | Mixed k6 payloads, 48 VUs | **~11 ms med** / **119 ms avg** | **27.0%** flat | lower in top flat | **~55 MB** in-use |
| **GoPDFLib 5000×48** | 2000 rows + **wrap**, PDF/A | **499 ms** mean / **390 ms** best | **~1.9%** flat | **~40%** | **~752 MB** peak after 5000 |

### Interpretation

1. **GoPDFLib data bench is heavier** than `BenchmarkGenerateTemplatePDF/Rows2000` because it enables **text wrapping** on Email/Description columns and runs **48 concurrent** generators (queueing + GC pressure → **~390–584 ms** avg vs **~36 ms** single-thread).
2. **`memclr` is low (~1.9%)** in GoPDFLib pprof despite high concurrency — Pass 4 buffer pre-grow and pooling help; **`drawTable` (~40%)** dominates CPU (wrap + PDF/UA cells).
3. **Peak heap ~1.8 GB** during 5000×48 reflects **sustained concurrent allocation**; HTTP load test showed **~55 MB in-use** at a snapshot after requests complete (different measurement).
4. vs **May 25 pre-Pass-4 load** (`memclr` **49.7%**, heap **442 MB**): Pass 4 optimizations reduced load-test hotspots substantially; GoPDFLib sustained bench confirms **`drawTable` + buffer growth** remain the next targets for wrap-heavy PDF/A tables.

---

## Artifacts

| File | Description |
|------|-------------|
| [baselines/gopdflib_pprof_stats_20260525.txt](./baselines/gopdflib_pprof_stats_20260525.txt) | Machine-readable stats |
| [baselines/gopdflib_pprof_runs/bench_gopdflib_5000_timing.txt](./baselines/gopdflib_pprof_runs/bench_gopdflib_5000_timing.txt) | Best timing run |
| [baselines/gopdflib_pprof_runs/cpu_gopdflib_data_run{1..5}.prof](./baselines/gopdflib_pprof_runs/) | CPU profiles |
| [sampledata/benchmarks/gopdflib/run_pprof_bench.sh](../sampledata/benchmarks/gopdflib/run_pprof_bench.sh) | Reproduce script |

---

## Related

- [PASS4_PDFA_RESULTS.md](./PASS4_PDFA_RESULTS.md) — internal/pdf + HTTP load comparison
- [PASS4_OPTIMIZATION_PLAN.md](./PASS4_OPTIMIZATION_PLAN.md) — optimization backlog
