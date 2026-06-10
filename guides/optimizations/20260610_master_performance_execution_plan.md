# gopdfsuit Master Performance Execution Plan

**Date:** 2026-06-10  
**Target Go Version:** `go1.26.4`  
**Branch Baseline:** `master` at `484b991`

## Goal

Push `master` to the highest sustainable benchmark throughput on the tracked Go benchmarks without changing PDF output semantics.

## Repro Baseline

### Benchmark commands

```bash
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test ./internal/pdf -run '^$' -bench 'BenchmarkGoPdfSuit$' -benchmem -benchtime=5s
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test ./internal/pdf -run '^$' -bench 'BenchmarkGenerateTemplatePDF/Rows2000$' -benchmem -benchtime=5s
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test ./internal/pdf -run '^$' -bench 'BenchmarkGoPdfSuit$' -benchtime=3s -cpuprofile /tmp/gopdfsuit_go1264.cpu -memprofile /tmp/gopdfsuit_go1264.mem
```

### Baseline results

- [x] `BenchmarkGoPdfSuit`: `218`, `26047411 ns/op`, `25626521 B/op`, `163386 allocs/op`
- [x] `BenchmarkGenerateTemplatePDF/Rows2000`: `216`, `28903973 ns/op`, `24686033 B/op`, `163312 allocs/op`

## Profile Findings

### CPU

- [x] `GenerateTemplatePDF` dominates end-to-end time.
- [x] `drawTable` is the largest content-generation hotspot.
- [x] `compress/flate` remains a major cost after content generation.
- [x] `GetTextWidth`, `fmt.Sprintf`, and structure-marking writes still show up in the hot path.

### Allocation

- [x] `bytes.growSlice`: `1036.46 MB`, `20.91%`
- [x] `compress/flate.NewWriter`: `1015.41 MB`, `20.49%`
- [x] `slices.Clone`: `353.68 MB`, `7.14%`
- [x] `BeginMarkedContentBuf`: `249.04 MB`, `5.02%`
- [x] `fmt.Sprintf`: `145.00 MB`, `2.93%`

## Execution Checklist

### Phase 1: Immediate Wins

- [x] Increase pooled final PDF buffer capacity to reduce `bytes.growSlice` in document assembly.
- [x] Remove the extra scratch-slice hop at final PDF assembly and keep a single caller-owned clone.
- [x] Avoid `contentStream.Bytes()` in the page compression loop and stream directly into the zlib writer.
- [x] Replace `append([]byte(nil), compressedBuf.Bytes()...)` with `slices.Clone` for a single explicit ownership copy.
- [x] Pre-size only the initial page content stream from template complexity instead of forcing every page to start at the same larger size.

### Phase 2: Structure and Table Pressure

- [ ] Reduce allocations in tagged-PDF structure writes, especially repeated map/slice growth in `BeginMarkedContentBuf`.
- [ ] Re-check table row scratch reuse in `drawTable` and eliminate any per-row reallocation still remaining.
- [ ] Audit expensive `fmt.Sprintf` usage on the document-generation path and replace the hottest cases with buffer-based writes.

### Phase 3: Compression Pressure

- [ ] Re-profile `compress/flate.NewWriter` after Phase 1 to verify whether writer-pool churn is still material.
- [ ] If still hot, tune pooled compression buffers and writer reuse strategy around actual page stream sizes.

## Acceptance Criteria

- [x] `BenchmarkGoPdfSuit` improves versus the `26047411 ns/op` baseline on `go1.26.4`.
- [x] `BenchmarkGenerateTemplatePDF/Rows2000` improves versus the `28903973 ns/op` baseline on `go1.26.4`.
- [x] `B/op` decreases while throughput improves.
- [x] PDF generation tests still pass.

## Executed Results

### Post-change benchmark results

- [x] `BenchmarkGoPdfSuit`: `242`, `23531821 ns/op`, `22539669 B/op`, `163307 allocs/op`
- [x] `BenchmarkGenerateTemplatePDF/Rows2000`: `241`, `24053195 ns/op`, `22563899 B/op`, `163347 allocs/op`

### Improvement versus baseline

- [x] `BenchmarkGoPdfSuit`: `-9.7% ns/op`, `-12.0% B/op`, allocs essentially flat
- [x] `BenchmarkGenerateTemplatePDF/Rows2000`: `-16.8% ns/op`, `-8.6% B/op`, allocs essentially flat

## Phase 2 Follow-through

- [x] Converted `StructureManager` page-index bookkeeping from maps to slices.
- [x] Preallocated table and row structure children where table shape is known.
- [x] Pre-grew XMP metadata builders to reduce repeated `strings.Builder` growth.
- [x] Revalidated with `go1.26.4` and a fresh pprof capture.
- [x] Stored a detailed post-Phase-2 profile report in `guides/optimizations/20260610_phase2_pprof_report.md`.

### Latest benchmark results

- [x] `BenchmarkGoPdfSuit`: `256`, `23025278 ns/op`, `21517345 B/op`, `157268 allocs/op`
- [x] `BenchmarkGenerateTemplatePDF/Rows2000`: `228`, `23838900 ns/op`, `22465256 B/op`, `157287 allocs/op`

## Notes

- `GOCACHE` is only the writable build-cache location for this environment. It is not a benchmark tuning parameter.
- Benchmark commands must be run serially. Parallel benchmark processes distorted results during earlier measurements.
