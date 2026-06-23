# Pass 4 PDF/A - Benchmark & Load Test Comparison

**Date:** 2026-05-25  
**Config:** `pdfaCompliant: true`, `taggedPDF: true`, `arlingtonCompatible: true`  
**Server workers:** Post-Pass-4 uses `runtime.NumCPU()` (=24); May 25 baseline used 48.

---

## Micro-benchmarks (Rows2000, PDF/A enabled)

| Metric | Pass 3 (pre-P4, always tagged) | Pass 4 PDF/A | Change |
|--------|-------------------------------|--------------|--------|
| **Time/op** | ~42–44 ms | **~30–32 ms** | **~28% faster** |
| **Allocs/op** | ~303K | **~163K** | **~46% fewer** |
| **Bytes/op** | ~17.6 MB | **~16.5–17.9 MB** | ~similar |
| Wrap time/op | ~46–47 ms | **~51–59 ms** | ~similar/slightly slower |
| Wrap allocs | ~327K | **~164K** | **~50% fewer** |

**Untagged Pass 4 (reference only):** ~11 ms, ~1.3K allocs - not used for compliance workloads.

Artifacts: `bench_pass4_pdfa_20260525.txt`, `bench_pass3_20260525.txt`

---

## k6 load test (48 VUs, PDF/A + signed payloads)

| Metric | May 25 baseline (pre-P4) | Pass 4 PDF/A | Change |
|--------|--------------------------|--------------|--------|
| **Total requests** | 1,277 | **4,317** | **+238%** |
| **Throughput** | ~25 req/s | **~143 req/s** | **~5.7×** |
| **HTTP failures** | 0% | **0%** | - |
| **Median latency** | 94 ms | **11 ms** | **~8.5× faster** |
| **p95 latency** | 3.11 s | **314 ms** | **~10× faster** |
| **p99 latency** | 27.45 s ✗ | **1.53 s ✓** | **~18× faster** (threshold met) |
| **Avg PDF gen time** | 968 ms | **119 ms** | **~8× faster** |

Artifacts:
- Baseline: `loadtest_k6_20260525.txt`, `loadtest_cpu_20260525.prof`, `loadtest_heap_20260525.prof`
- Post-P4: `loadtest_k6_pass4_pdfa_20260525.txt`, `loadtest_cpu_pass4_pdfa_20260525.prof`, `loadtest_heap_pass4_pdfa_20260525.prof`

---

## pprof comparison

### CPU - flat `memclrNoHeapPointers`

| Run | memclr flat | memmove flat | flate.deflate cum |
|-----|-------------|--------------|-------------------|
| May 25 baseline | **49.7%** | 9.8% | 14.5% |
| Pass 4 PDF/A | **27.0%** | 22.2% | (lower in top flat) |

**−46% relative** memclr CPU despite PDF/A tagging enabled.

### Heap - in-use during load

| Run | Total in-use | `BeginMarkedContentBuf` | `bytes.growSlice` |
|-----|--------------|-------------------------|-------------------|
| May 25 baseline | **442 MB** | **64 MB** (581K objs) | 150 MB |
| Pass 4 PDF/A | **~55 MB** | **~0.6 MB** | ~14 MB |

Struct pooling + buffer pre-grow + template pool dramatically cut heap under PDF/A load.

Full pprof text: `loadtest_comparison_pass4_pdfa_20260525.txt`

---

## Compliance defaults (product)

- **Benchmarks** (`benchmark_test.go`, `benchmark_macro_test.go`): use `pdfAConfig()` with PDF/A + tagged PDF.
- **Default k6 load test** (`load_test.js`): uses `getPayloadOptions('tagged')`.
- **API:** Set `"pdfaCompliant": true` (implies tagging via `taggedPDF || pdfaCompliant` in generator). Optional explicit `"taggedPDF": true`.

Unsigned scenarios remain available via `load_test_unsigned.js` for perf regression only.

---

## Commands to reproduce

```bash
# Benchmarks (PDF/A)
go test -run='^$' -bench='BenchmarkGenerateTemplatePDF/Rows2000|BenchmarkGoPdfSuit' \
  -benchmem -count=3 ./internal/pdf/

# Load + pprof
go run ./cmd/gopdfsuit &
curl -o guides/cursor/baselines/loadtest_cpu.prof \
  "http://127.0.0.1:8080/debug/pprof/profile?seconds=35" &
cd test/generate_template-pdf && k6 run load_test.js
curl -o guides/cursor/baselines/loadtest_heap.prof \
  "http://127.0.0.1:8080/debug/pprof/heap"
```
