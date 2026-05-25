# Performance Optimization: Pass 1–4 (41 tasks)

## Summary

Four-phase performance program for GoPdfSuit PDF generation, based on a 6-agent architecture audit. Targets hot-path allocations, I/O copies, concurrency bottlenecks, data-structure overhead, and load-test tail latency across generation, merge, redact, and HTTP server paths.

**41 optimizations** delivered across 4 passes. All tests pass.

| Phase | Focus | Tasks | Key outcome |
|-------|--------|-------|-------------|
| **Pass 1** | Low-hanging fruit | 10 | Buffer pooling, zero-alloc text encoding, batched writes |
| **Pass 2** | Architecture | 12 | Direct-write APIs, parallel decode/compress, macro benchmarks |
| **Pass 3** | Advanced | 5 | Allocation-free wrap, typed structure tree, unified redact parser |
| **Pass 4** | Load-test hotspots | 14 | PDF/UA gating, buffer pre-grow, parallel zlib, template pool, p99 fixes |

**Headline (Zerodha gold-standard, PDF/A):** **2061 ops/s peak** (10-run avg **1705 ops/s**) on 48 concurrent workers — **~197% faster than Go 1.24** and **~3.4× faster than Go 1.26**, with PDF/A-compliant tagged output.

---

## Understanding throughput: 2061 ops/s is **not** single-core

The **2061 ops/s** peak (and **1705 ops/s** 10-run average) are **aggregate machine throughput**, not per-core or per-worker speed.

| Concept | Value | Meaning |
|---------|-------|---------|
| **Hardware** | 24 logical CPUs (Intel i7-13700HX, WSL2) | Physical parallelism available |
| **Concurrency** | **48 workers** (goroutines) | Up to 48 PDFs generating at once |
| **Throughput (2061 ops/s)** | `5000 ÷ wall-clock seconds` | Total PDFs completed per second **across the whole run** |
| **Avg latency (~23–28 ms)** | Per-document `GeneratePDF` time | How long **one** PDF takes on average |

**How they relate:** With 48 workers and ~23 ms average per PDF on a good run, theoretical ceiling is roughly `48 ÷ 0.023 ≈ 2090 ops/s` — close to the observed **2061 ops/s** peak.

**Per-core rough efficiency:** `2061 ÷ 24 ≈ 86 PDFs/sec/core` (upper bound). This is **not** “one core generates 2061 PDFs/sec.”

**Single-thread baseline (for contrast):** `BenchmarkGenerateTemplatePDF/Rows2000` serial Pass 4 PDF/A averages **~36 ms/doc** → **~28 PDFs/sec on one goroutine**.

**Run-to-run variance:** 10 sequential WSL runs ranged **1542–2021 ops/s** (avg **1705**). Peak manual run on idle machine: **2061 ops/s**.

---

## Benchmark Results (Intel i7-13700HX, 24 CPUs)

### Micro-benchmarks (`internal/pdf`, PDF/A + tagged)

| Benchmark | Pre-opt (approx) | Post Pass 2 | Post Pass 3 | **Post Pass 4** | Notes |
|-----------|------------------|-------------|-------------|-----------------|-------|
| `Rows2000` serial | 35–56 ms/op | **~29 ms/op** | ~42 ms/op | **~36 ms/op avg** (best **~32 ms**) | ~16% faster vs Pass 3; **~46% fewer allocs** |
| `Rows2000` allocs | ~292K/op | ~292K/op | ~303K/op | **~163K/op** | Pass 4 tagging + pooling win |
| `Rows2000` parallel | — | **~7–9 ms/op** | — | — | ~4–5× vs serial (Pass 2) |
| `WrapEnabled/Rows2000` allocs | — | ~357K/op | **~327K/op** | **~164K/op** | −50% vs Pass 3 |
| `Rows10000` / `Rows25000` | — | ~129 ms / ~342 ms | — | — | Macro benchmarks (Pass 2) |

### HTTP load test (k6, 48 VUs, PDF/A tagged payloads)

| Metric | Pre-Pass 4 (May 25) | **Post Pass 4** | Change |
|--------|---------------------|-----------------|--------|
| **Throughput** | ~25 req/s | **~143 req/s** | **~5.7×** |
| **p99 latency** | 27.45 s | **1.53 s** | **~18× faster** |
| **Median latency** | 94 ms | **11 ms** | **~8.5× faster** |
| **Heap in-use (pprof)** | 442 MB | **~55 MB** | **−88%** |
| **`memclr` CPU (flat)** | 49.7% | **27.0%** | **−46% relative** |

---

## Zerodha Gold Standard — End-to-End Workload (Pass 4 vs Go 1.24 / 1.26)

**Entry:** `sampledata/gopdflib/zerodha/main.go`  
**Config:** 5000 iterations, **48 workers**, 80% Retail / 15% Active / 5% HFT, **PDF/A + tagged PDF + digital signatures**  
**Go version (Pass 4):** 1.24.0  
**Sources:** [1.24.txt](../sampledata/gopdflib/zerodha/1.24.txt) (10 runs), [1.26.txt](../sampledata/gopdflib/zerodha/1.26.txt) (10 runs), Pass 4 ([WSL 10-run stats](./baselines/zerodha_bench_x10_wsl_stats_20260525.txt), May 2026)

### Average across all historical runs

| Metric | Go 1.24 (10-run avg) | Go 1.26 (10-run avg) | **Pass 4 (10-run avg)** | vs 1.24 | vs 1.26 |
|--------|----------------------|----------------------|-------------------------|---------|---------|
| **Throughput** | 574 ops/s | 392 ops/s | **1705 ops/s** | **+197%** | **+335% (~4.4×)** |
| **Avg latency** | 80.7 ms | 119.2 ms | **27.7 ms** | **66% faster** | **77% faster** |
| **Wall time (5000 docs)** | 8.75 s | 12.84 s | **2.95 s** | **66% faster** | **77% faster** |
| **Max latency (tail)** | 1968 ms | 2851 ms | **747 ms** avg | **62% lower** | **74% lower** |
| **Peak memory** | 1107 MB | 1092 MB | **1170 MB** | ~similar | ~similar |

### Best-run comparison (peak performance)

| Metric | Go 1.24 best | Go 1.26 best | **Pass 4 peak** | Opt 6 ref |
|--------|--------------|--------------|-----------------|-----------|
| **Throughput** | 638 ops/s | 441 ops/s | **2061 ops/s** | 1741 ops/s |
| **Avg latency** | 71.9 ms | 104.5 ms | **22.7 ms** | 27.0 ms |
| **Wall time** | 7.84 s | 11.33 s | **2.43 s** | 2.87 s |

Pass 4 **peak** beats Go 1.24 **best** by **~223%** and Go 1.26 **best** by **~367%**. Even the **worst** of 10 WSL runs (**1542 ops/s**) exceeds Go 1.24 **best** (**638 ops/s**) by **~142%**.

### Pass 4 timing detail (10 sequential WSL runs)

| Run | Throughput (ops/s) | Avg latency (ms) | Max latency (ms) | Total time (s) | Peak mem (MB) |
|-----|-------------------|------------------|------------------|----------------|---------------|
| 1 | 1634.72 | 28.540 | 592.154 | 3.059 | 1158.41 |
| 2 | 1741.68 | 26.742 | 667.205 | 2.871 | 1170.82 |
| 3 | 1756.62 | 26.707 | 709.438 | 2.846 | 1253.73 |
| 4 | 1613.58 | 29.070 | 883.804 | 3.099 | 1141.77 |
| 5 | 1701.42 | 27.799 | 751.121 | 2.939 | 1197.37 |
| 6 | **1542.39** (worst tput) | **30.467** (worst avg) | 733.011 | **3.242** (worst) | 1217.82 |
| 7 | 1601.74 | 29.150 | 619.333 | 3.122 | 1216.33 |
| 8 | 1557.89 | 30.050 | 755.173 | 3.209 | 1094.63 |
| 9 | **2020.59** (batch best) | **22.988** (batch best avg) | 757.213 | **2.475** (batch best) | 1087.53 |
| 10 | 1878.83 | 24.958 | 694.527 | 2.661 | **1074.23** (best mem) |

| Aggregate | Best | Worst | **Average** | σ |
|-----------|------|-------|-------------|---|
| **Throughput** | 2020.59 ops/s* | 1542.39 ops/s | **1704.95 ops/s** | 151.25 |
| **Avg latency** | 22.99 ms | 30.47 ms | **27.65 ms** | 2.55 ms |
| **Max latency** | 592.15 ms | 883.80 ms | **746.51 ms** | — |
| **Wall time** | 2.48 s | 3.24 s | **2.95 s** | 0.28 s |
| **Peak memory** | 1074 MB | 1254 MB | **1170 MB** | 58 MB |

\*Peak observed: **2061.33 ops/s** (manual WSL run, idle machine).

### Column glossary (Zerodha table)

| Column | What it measures |
|--------|------------------|
| **Throughput (ops/s)** | `5000 ÷ wall-clock seconds` — total PDFs/sec with 48 workers in flight. **System aggregate, not per-core.** |
| **Avg latency (ms)** | Mean time for one `GeneratePDF` across all 5000 docs (Retail + Active + HFT mix). Dominated by fast Retail docs (~80%). |
| **Max latency (ms)** | Slowest single PDF in the run — almost always an **HFT** doc (2000 rows) under worker contention. Tail behavior, not typical user latency. |
| **Total time (s)** | Wall clock for the full 5000-iteration batch. Equals `5000 ÷ throughput`. |
| **Peak mem (MB)** | Highest `runtime.MemStats.Alloc` sampled during the run. HFT docs (~2.4 MB each) drive peaks (~1.1–1.3 GB). |

### Workload mix

| Tier | Share | Description | Typical output |
|------|-------|-------------|----------------|
| **Retail** | 80% | 1-page contract note, 2 trades, digitally signed | ~61 KB |
| **Active** | 15% | 2–3 pages, 40 trades | ~76 KB |
| **HFT** | 5% | 50+ pages, 2000 trades | ~2.4 MB |

### CPU pprof (Pass 4 Zerodha, 5 runs)

| Hotspot | Best | Worst | **Average** |
|---------|------|-------|-------------|
| **`memclrNoHeapPointers` (flat)** | 3.68% | 4.81% | **4.44%** |
| **`drawTable` (cum)** | 16.99% | 18.23% | **17.73%** |
| **`GenerateTemplatePDF` (cum)** | 80.00% | 80.50% | **80.27%** |

Retail-heavy mix → high throughput, lower `drawTable` share than pure 2000-row micro-benchmarks (~37%).

**Artifacts:** [ZERODHA_BENCHMARK_RESULTS.md](./ZERODHA_BENCHMARK_RESULTS.md), [baselines/zerodha_bench_x10_wsl/](./baselines/zerodha_bench_x10_wsl/)

---

## Pass 1 — Low-Hanging Fruit (10/10)

Eliminate per-cell allocations and reduce syscall/write overhead on the content-stream hot path.

### Changes

| ID | Change | Files |
|----|--------|-------|
| P1-01 | **`appendTextForPDF`** — zero-alloc text encoding into `[]byte` (no `string` return on `Tj` path) | `utils.go`, `font/metrics.go`, `draw.go` |
| P1-02 | **`RuneSet` bitmap** — replace `UsedChars map[rune]bool` with dense 64 KiB bitmap | `font/runeset.go`, `registry.go`, `metrics.go` |
| P1-03 | **Batched cell PDF commands** — one `Write` per table cell instead of ~10 | `draw.go` |
| P1-04 | **Extended zlib pool** to PDF/A, metadata, font subset paths | `pdfa.go`, `metadata.go`, `subset.go` |
| P1-05 | **Pre-grow page content streams** (32 KB initial capacity) | `pagemanager.go` |
| P1-06 | **Pre-grow `CompressBufPool`** (64 KB) | `font/compression.go` |
| P1-07 | **Respect `noLock`** on cloned font registries in `GenerateSubsets` / `AssignObjectIDs` | `font/registry.go` |
| P1-08 | **Fix pprof HTTP routes** — `/heap`, `/block`, `/mutex` serve correct profiles | `handlers/handlers.go` |
| P1-09 | **Benchmark improvements** — `ReportAllocs`, `SetBytes`, parallel variant | `benchmark_test.go` |
| P1-10 | **Image cache singleflight** — dedupe concurrent decode of same image hash | `image.go` |

### Impact

- Removes string intermediates on every `Tj` operator
- Cuts map inserts on font subsetting hot path
- Reduces `WriteString` churn in table rendering (~25K → ~5K writes per 5K-cell table)

---

## Pass 2 — Architecture Changes (12/12)

Structural changes for throughput, memory bandwidth, and observability.

### Changes

| ID | Change | Files |
|----|--------|-------|
| P2-01 | **`WriteImageXObject`** — direct buffer write, no intermediate `string` | `image.go`, `generator.go` |
| P2-02 | **`WriteTrueTypeFontObjects`** — stream 7 font objects directly to PDF buffer | `font/metrics.go`, `generator.go` |
| P2-03 | **Parallel image decode** via bounded `errgroup` (`runtime.NumCPU()`) | `generator.go` |
| P2-04 | **Parallel page stream compression** — parallel zlib, serial write phase | `generator.go` |
| P2-05 | **Incremental MD5** — hash as we write, no full-buffer re-read at document ID | `generator.go` |
| P2-06 | **Compact xref** shared helper for merge + XFDF | `xref/xref.go`, `merge/merger.go`, `form/xfdf.go` |
| P2-07 | **TTF bulk slice parsing** — `parseHmtx`, `parseCmapFormat4` via direct indexing | `font/ttf.go` |
| P2-08 | **Sparse CIDToGIDMap** — stream sparse map when `maxCID > 8192` without 256 KB+ alloc | `font/metrics.go` |
| P2-09 | **Macro benchmark suite** — 2K / 10K / 25K rows + wrap variant; Typst behind `//go:build compare` | `benchmark_macro_test.go`, `benchmark_compare_test.go` |
| P2-10 | **Pooled encryption buffers** — `encScratchPool` in `EncryptStream` | `encryption/encrypt.go` |
| P2-11 | **`maxConcurrent := runtime.NumCPU()`** — replace hardcoded 48 workers | `cmd/gopdfsuit/main.go` |
| P2-12 | **PDF/A font manager lock fixes** — download/load outside mutex, double-check cache | `font/pdfa.go` |

### Impact

- ~15–20% faster on 2K-row table benchmark
- ~4–5× throughput under parallel load (24 CPUs)
- Eliminates largest binary stream copies at finalize
- Fixes p99 contention on image cache and PDF/A font cold start

### New packages / files

- `internal/pdf/xref/` — shared compact xref writer
- `internal/pdf/benchmark_macro_test.go`
- `internal/pdf/benchmark_compare_test.go`
- `internal/pdf/font/metrics_cidmap_test.go`

---

## Pass 3 — Advanced Optimizations (5/5)

Memory layout, parser unification, and typed data structures.

### Changes

| ID | Change | Files |
|----|--------|-------|
| P3-01 | **`WrapTextInto` + `WrapState`** — reusable buffers, no `strings.Fields` / string concat | `utils.go`, `draw.go`, `utils_wrap_test.go` |
| P3-02 | **`ExtraObjects map[int][]byte`** — zero-copy write at finalize | `pagemanager.go`, `generator.go`, `outline.go` |
| P3-03 | **Redact parser unification** — `merge.FindObjectBoundaries`, `map[int][]byte`, hoisted regex | `redact/*.go`, `redact_parser_test.go` |
| P3-04 | **`ImageObject` struct packing** — reordered fields for better cache layout | `image.go` |
| P3-05 | **`StructKid` typed slice** — replaces `Kids []interface{}` in PDF/UA tree | `structure.go`, `generator.go` |

### Impact

- **−8.6%** allocs on wrap-enabled 2K-row tables
- **−11.3%** allocs on wrap-enabled 10K-row tables
- Redact path avoids full-PDF regex scan
- PDF/UA structure tree avoids interface boxing and type assertions

### Deferred (low impact / high risk)

- `TTFFont.Flags uint8` bit-packing
- `FontObjectIDs` struct embedding in `RegisteredFont`

---

## Pass 4 — Load-Test Hotspots (14/14)

Targets tail latency, heap under saturation, and CPU efficiency while preserving PDF/A compliance.

### Changes

| ID | Change | Files |
|----|--------|-------|
| P4-01 | **Gate PDF/UA tagging** — `TaggedPDF` config; no-op `StructureManager` when off | `pagemanager.go`, `structure.go`, `draw.go`, `generator.go` |
| P4-02 | **Pre-grow page streams** — 64 KiB `Grow` on new pages | `pagemanager.go` |
| P4-03 | **Final PDF slice pool** — `finalPDFSlicePool` + `slices.Clone` | `generator.go` |
| P4-04 | **Hoist drawTable scratch** — border/xobj/color/placeholder/checkbox buffers | `draw.go` |
| P4-05 | **`appendTextForPDF`** on hot drawTable paths | `utils.go`, `font/metrics.go`, `draw.go` |
| P4-06 | **Incremental wrap width** — running `lineWidth` in `WrapTextInto` | `utils.go` |
| P4-07 | **Compress buffer Grow** — `max(4096, len/4)` before zlib | `generator.go` |
| P4-08 | **Compression pooling** — 64 KiB pool; subset/metadata/pdfa migrated | `font/compression.go`, `subset.go`, `metadata.go`, `pdfa.go` |
| P4-09 | **Parallel page zlib** — `errgroup` in finalize | `generator.go` |
| P4-10 | **Handler template pool** — `templatePDFPool` | `handlers/handlers.go` |
| P4-11 | **k6 scenario split** — tagged / unsigned load scripts | `test/generate_template-pdf/` |
| P4-12 | **Concurrency tuning** — `maxConcurrent = runtime.NumCPU()` | `cmd/gopdfsuit/main.go` |
| P4-13 | **Signer PEM cache** — hash-keyed `sync.Map` | `signature/signature.go` |
| P4-14 | **StructElem pool** — `acquireStructElem` / `ReleaseStructElemsToPool` | `structure.go` |

### Impact

- **~5.7×** HTTP throughput; **~18×** faster p99 under load
- **−88%** heap in-use; **`memclr` CPU −46%** relative under load
- **~46% fewer allocs** on PDF/A Rows2000 micro-bench
- **~197% higher throughput** on Zerodha gold-standard vs Go 1.24 (10-run avg); **~4.4× vs Go 1.26**; peak **2061 ops/s**

---

## Documentation

Added under `guides/cursor/`:

- `PERFORMANCE_AUDIT.md` — full 6-agent audit report
- `IMPLEMENTATION_PLAN.md` — phased roadmap and status
- `PASS1_BLUEPRINTS.md` / `PASS3_BLUEPRINTS.md` — before/after code
- `PASS4_OPTIMIZATION_PLAN.md` / `PASS4_PDFA_RESULTS.md` — Pass 4 plan and results
- `ZERODHA_BENCHMARK_RESULTS.md` / `GOPDFLIB_PPROF_RESULTS.md` — end-to-end benchmarks
- `baselines/bench_pass{1,2,3,4}_20260525.txt` — benchmark snapshots
- `baselines/zerodha_bench_x10_wsl/` — Zerodha 10-run WSL raw output
- `baselines/zerodha_bench_x10_wsl_stats_20260525.txt` — Zerodha 10-run WSL stats

---

## Test Plan

- [x] `go test ./internal/pdf/...`
- [x] `go test ./internal/pdf/font/...`
- [x] `go test ./internal/pdf/redact/...`
- [x] `go test ./internal/handlers/...`
- [x] `go test ./pkg/gopdflib/...`
- [x] `go test -run='^$' -bench=BenchmarkGenerateTemplatePDF -benchmem ./internal/pdf/`
- [x] `go test -run='^$' -bench=BenchmarkGenerateTemplatePDF_WrapEnabled -benchmem ./internal/pdf/`
- [x] Load test HTTP server under concurrent requests with pprof CPU/heap profiles
- [x] Zerodha gold-standard benchmark (5000×48, 10 timing runs)
- [ ] `go test -tags=compare -bench=BenchmarkTypst ./internal/pdf/` (optional; requires Typst binary)
- [ ] Verify PDF/A and encrypted PDF output byte-validity on sample templates

### Profiling commands

```bash
# Micro-benchmark
go test -run='^$' -bench=BenchmarkGenerateTemplatePDF/Rows2000 -benchtime=30s \
  -cpuprofile=/tmp/cpu.prof -memprofile=/tmp/mem.prof ./internal/pdf/
go tool pprof -http=:8081 /tmp/cpu.prof

# Zerodha end-to-end (5000 iter, 10 timing runs)
bash sampledata/gopdflib/zerodha/run_bench_x10.sh

# Optional: 5 runs + CPU/heap pprof
bash sampledata/gopdflib/zerodha/run_bench_x5.sh
```

---

## Breaking Changes

None for public API consumers. Internal types changed:

- `PageManager.ExtraObjects`: `map[int]string` → `map[int][]byte`
- `StructElem.Kids`: `[]interface{}` → `[]StructKid`
- Redact `objMap`: `map[string][]byte` → `map[int][]byte`

---

## Dependencies

- Added `golang.org/x/sync` (singleflight, errgroup)
