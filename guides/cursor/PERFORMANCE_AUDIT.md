# GoPdfSuit Performance Architecture & Profiling Audit

**Date:** 2026-05-25  
**Scope:** Full codebase - `internal/pdf/`, font pipeline, merge, redact, HTTP server, benchmarks

---

## Executive Summary

GoPdfSuit already shows deliberate performance engineering: `sync.Pool` for PDF buffers, zlib writers, and RGB scratch space; `appendFmtNum`; compact xref subsections; image FNV-1a dedup cache; per-request font registry cloning.

Three **structural bottlenecks** dominate throughput and p99 latency:

| # | Bottleneck | Impact | Root Cause |
|---|-----------|--------|------------|
| **1** | String-materialization on content-stream hot path | Very high alloc rate on table-heavy workloads | Every `Tj` passes through `formatTextForPDF` → `string`; `WrapText` allocates per word/line |
| **2** | Binary stream routing through intermediate `string` values | High CPU + 2× memory bandwidth at finalize | `CreateImageXObject`, `GenerateTrueTypeFontObjects`, `ExtraObjects map[int]string` |
| **3** | Sequential pipeline + global lock contention | p99 spikes at 48 concurrent requests | Image decode/compression not parallelized; global `imgCache` RWMutex; PDF/A font manager locks during I/O |

**Estimated optimization potential:** 30–50% throughput (Pass 1 + Pass 2); 40–60% alloc reduction on custom-font workloads.

---

## Agent 1 - Memory (Allocation Eliminator)

### Existing Pools

| Pool | File | Purpose |
|------|------|---------|
| `pdfBufferPool` | `generator.go:24` | Final PDF assembly (64KB pre-grow) |
| `scratchBufPool` | `generator.go:33` | 128-cap scratch for strconv |
| `rgbDataPool` | `image.go:25` | 1MB RGB conversion |
| `ZlibWriterPool` | `font/compression.go:13` | Recycles ~256KB zlib tables |
| `CompressBufPool` | `font/compression.go:22` | Compression output buffers |

### Critical Findings

1. **`formatTextForPDF` → string** - `utils.go:368-373`, every `Tj` in `draw.go`
2. **`EncodeTextForCustomFont` → `string(buf)`** - `font/metrics.go:1035-1057`
3. **`WrapText` string churn** - `utils.go:378-474`
4. **PDF objects as `map[int]string`** - finalize spike
5. **`EncryptStream` 3–4 allocs per stream** - when encryption enabled
6. **Page buffers zero-capacity start** - `pagemanager.go:70`

### Top Optimizations

- `appendTextOperand(dst []byte, ...)` - zero-alloc text encoding
- `RuneSet` bitmap - replace `UsedChars map[rune]bool`
- Pooled encryption scratch
- Pre-grow page content streams

---

## Agent 2 - I/O & CPU (Hotpath Optimizer)

### Bottleneck Ranking

1. String intermediates for binary streams - `image.go:441-490`, `font/metrics.go:647-705`
2. Fragmented content-stream writes - `draw.go:1125-1197` (~25K Write calls / 5K cells)
3. Unpooled `zlib.NewWriter` - `pdfa.go`, `metadata.go`, `subset.go`, `form/xfdf.go`
4. Full-buffer MD5 - `generator.go:1004-1007`
5. TTF `parseHmtx` per-glyph `binary.Read` - `font/ttf.go:284-301`
6. Sparse `CIDToGIDMap` - `font/metrics.go:857`

### Existing Optimizations (Preserve)

- `appendFmtNum` in `draw.go:20-45`
- `parseProps` via `sync.Map` (`utils.go:80-81`)
- Compact xref in generator (`generator.go:1217-1285`)

---

## Agent 3 - Concurrency Auditor

### Model

- **Inter-request:** Semaphore (48 slots) + `CloneForGeneration()`
- **Intra-document:** Fully serial (layout dependencies are real)

### Contention Hotspots

| Hotspot | Severity |
|---------|----------|
| Global `imgCache` RWMutex | High |
| `PDFAFontManager` lock during HTTP download | High |
| `CloneForGeneration` global RLock | Medium |
| Hardcoded 48 vs `runtime.NumCPU()` | Medium |

### Safe Parallelism Wins

- Image decode (`generator.go:157-241`)
- Page stream zlib (`generator.go:741-782`)
- Font subsetting (per-font independent)

---

## Agent 4 - Data Structure Evaluator

### Map Replacements

| Current | Recommendation |
|---------|----------------|
| `UsedChars map[rune]bool` | `[65536/64]uint64` bitmap |
| `cellImageObjects map[string]*ImageObject` | `map[uint64]*ImageObject` |
| `FieldSet map[int]bool` | Bitset |
| `objMap map[string][]byte` (redact) | `map[int][]byte` |

### Complexity Hot Paths

```
GenerateTemplatePDF (table-heavy): O(R × C × (W + T))
Redact FindTextOccurrences: O(P × (|content| + N²))
```

---

## Agent 5 - Profiling Strategy

### Existing Benchmarks

| Function | File | Measures |
|----------|------|----------|
| `BenchmarkGoPdfSuit` | `benchmark_test.go:79` | 2,000-row Helvetica table E2E |
| `BenchmarkTypst` | `benchmark_test.go:98` | External Typst comparison |

### Gaps

- No `b.ReportAllocs()` / `b.SetBytes()`
- No 1000+ page macro benchmark
- No mixed-font / image-heavy bench
- No `b.RunParallel` concurrent bench
- Broken pprof heap/block HTTP routes

---

## Agent 6 - Synthesis (Impact vs Effort)

See [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) for the phased roadmap.

---

## Validation Commands

```bash
# Baseline
go test -run='^$' -bench=BenchmarkGoPdfSuit -benchmem -benchtime=10s -count=3 ./internal/pdf/

# CPU + heap profiles
go test -run='^$' -bench=BenchmarkGoPdfSuit -benchtime=30s \
  -cpuprofile=/tmp/cpu.prof -memprofile=/tmp/mem.prof ./internal/pdf/

go tool pprof -http=:8081 /tmp/cpu.prof
go tool pprof -http=:8081 -alloc_space /tmp/mem.prof

# Delta comparison
benchstat /tmp/bench_before.txt /tmp/bench_after.txt
```

---

## Related Docs

- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - phased execution plan
- [PASS1_BLUEPRINTS.md](./PASS1_BLUEPRINTS.md) - before/after code for Pass 1
- [../additionalnotes/PERFORMANCE_OPTIMIZATIONS.md](../additionalnotes/PERFORMANCE_OPTIMIZATIONS.md) - prior pprof analysis
