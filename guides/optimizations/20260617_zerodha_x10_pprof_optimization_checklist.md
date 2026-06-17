# Zerodha gopdflib x10 pprof Optimization Checklist

**Date:** 2026-06-17  
**Workload:** `sampledata/gopdflib/zerodha`, 80% retail / 15% active / 5% HFT  
**Primary command:** `make bench-gopdflib-zerodha-x10`  
**Profile command:** `make bench-gopdflib-zerodha-x5`  
**Go version observed:** `go1.26.4`  
**Concurrency:** 48 workers, `GOMAXPROCS=24`

## Artifacts

- `guides/cursor/baselines/zerodha_bench_x10_wsl/zerodha_run{1..10}.txt`
- `guides/cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt`
- `guides/cursor/baselines/zerodha_pprof_runs/zerodha_run{1..5}.txt`
- `guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run{1..5}.prof`
- `guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof`
- `guides/cursor/baselines/zerodha_pprof_runs/zerodha_bench`

## Measurement Summary

### Baseline x10 Timing

| Metric | Value |
|--------|------:|
| Best throughput | 3028.74 ops/sec |
| Worst throughput | 2470.10 ops/sec |
| Mean throughput | 2690.73 ops/sec |
| Median throughput | 2649.69 ops/sec |
| Stddev throughput | 194.91 ops/sec |
| Mean peak allocated | 1269.97 MB |

### Current x10 Timing

| Metric | Value |
|--------|------:|
| Best throughput | 6594.76 ops/sec |
| Worst throughput | 4846.96 ops/sec |
| Mean throughput | 5768.08 ops/sec |
| Median throughput | 5768.67 ops/sec |
| Stddev throughput | 516.54 ops/sec |
| Mean avg latency | 8.206 ms |
| Mean peak allocated | 606.02 MB |

### Delta

| Metric | Delta |
|--------|------:|
| Mean throughput | +3077.35 ops/sec / +114.37% |
| Median throughput | +3118.98 ops/sec / +117.71% |
| Mean peak allocated | -663.95 MB / -52.28% |

**Output-size check:** retail and active outputs stayed stable at `61293` and `76065` bytes. HFT intentionally changed from `2289155` to `748163` bytes after compacting shared-row table structure from per-cell `StructElem` objects to pre-attached MCID leaves.

## Current CPU Profile

Representative profile: `guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof`  
Total samples: 16.23 s

| Hotspot | Current | Baseline | Status |
|---------|--------:|---------:|--------|
| `drawTable` cumulative CPU | 25.94% / 4.21 s | 21.94% | Faster end-to-end, still dominant by share |
| `drawSharedLayoutRow` cumulative CPU | 16.02% / 2.60 s | n/a | Still dominant inside HFT path |
| `drawSharedDeferRow` cumulative CPU | 12.20% / 1.98 s | n/a | Still dominant inside HFT path |
| `formatStructElemObjectTo` flat CPU | 2.22% / 0.36 s | 7.57% | Done |
| `formatStructElemObjectTo` cumulative CPU | 6.41% / 1.04 s | 12.58% | Done |
| `BeginMarkedContentBufWithMCID` cumulative CPU | 1.05% / 0.17 s | 8.50% | Done |
| `WriteCellMarkedContentBDC` cumulative CPU | 1.42% / 0.23 s | n/a | New batched path |
| `ReleaseStructElemsToPool` cumulative CPU | 1.85% / 0.30 s | 7.04% | Done |
| `font.GenerateTrueTypeFontObjects` cumulative CPU | 10.72% / 1.74 s | 12.18% | Improved, still above target |
| `font.pageContentFingerprint` cumulative CPU | 3.51% / 0.57 s | 3.22% | Still open |
| `compress/flate` / `compress/zlib` close cumulative CPU | 8.75% / 1.42 s | 13.77% | Done |
| `signature.UpdatePDFWithSignatureBuffer` cumulative CPU | 6.53% / 1.06 s | 3.88% | Allocation fixed, CPU share higher after other wins |

## Current Heap Profile

Profile: `guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof`

### In-use Space

| Hotspot | Current | Baseline | Status |
|---------|--------:|---------:|--------|
| Total in-use profile | 333.83 MB | 668.68 MB | Done |
| `bytes.growSlice` | 233.70 MB / 70.01% | 432.84 MB / 64.73% | Done |
| `drawTable` cumulative | 164.74 MB / 49.35% | 309.63 MB / 46.30% | Done |
| `getPageContentStreamBuffer` cumulative | 151.38 MB / 45.35% | 231.69 MB / 34.65% | Done |
| `ensurePDFBufferCapacity` cumulative | 63.37 MB / 18.98% | 166.66 MB / 24.92% | Done |
| `BeginMarkedContentBufWithMCID` cumulative | 2.00 MB / 0.60% | 43.51 MB / 6.51% | Done |
| `formatStructElemObjectTo` cumulative | 4.50 MB / 1.35% | n/a | Done |

### Alloc Space

| Hotspot | Current | Baseline | Status |
|---------|--------:|---------:|--------|
| Total alloc profile | 1302.91 MB | 2444.55 MB | Done |
| `bytes.growSlice` | 370.01 MB / 28.40% | 652.13 MB / 26.68% | Done |
| `drawTable` cumulative | 342.02 MB / 26.25% | 517.02 MB / 21.15% | Done |
| `ensurePDFBufferCapacity` cumulative | 63.37 MB / 4.86% | 391.31 MB / 16.01% | Done |
| `formatStructElemObjectTo` cumulative | 51.52 MB / 3.95% | 189.02 MB / 7.73% | Done |
| `signature.UpdatePDFWithSignatureBuffer` cumulative | 37.54 MB / 2.88% | 178.34 MB / 7.30% | Done |
| `signature.CreateSignatureField` cumulative | 115.68 MB / 8.88% | n/a | Improved but still open |
| `font.GenerateTrueTypeFontObjects` cumulative | 192.49 MB / 14.77% | 218.63 MB / 8.94% | Improved MB, still open by share |

## Priority Checklist

### P0 - Harness and Measurement Hygiene

- [x] Fix `sampledata/gopdflib/zerodha/run_bench_x5.sh` to build the whole Zerodha package so `ecdsa_certs.go` is included.
- [x] Add `bench-gopdflib-zerodha-x10-pprof` so timing and profiles can be run through one make target.
- [x] Write `guides/cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt` after x10 runs with best/worst/mean/median/stddev.
- [x] Keep `GOCACHE` and `GOMODCACHE` override examples in the benchmark command output.
- [x] Use x10 mean/median as the regression gate; use profile runs only for hotspot direction.

### P1 - Page and Final Buffer Growth

- [x] Revisit `estimateInitialContentStreamCap(template)` for HFT pages.
- [x] Split content stream pool buckets by capacity class.
- [x] Discard oversized page content buffers instead of always returning them to the pool.
- [x] Re-check `estimateFinalPDFSize` against actual retail, active, and compacted HFT output sizes.
- [x] Add debug-only `BENCH_DEBUG_CAPS` instrumentation for estimated vs actual PDF/content stream capacity.
- [x] Validation gate: `bytes.growSlice` in-use below 300 MB and x10 mean not below 2690 ops/sec. Current `bytes.growSlice` is 233.70 MB and x10 mean is 5768.08 ops/sec.

### P2 - Tagged Structure Serialization

- [x] Optimize `formatStructElemObjectTo` flat CPU with direct append helpers and a table-cell leaf fast path.
- [x] Reduce small leaf formatter overhead with `formatSingleMCIDTableCellStructElem`.
- [x] Keep the generic formatter type-safe after the fast path.
- [x] Replace more struct-element object string fragments with direct final-buffer scratch writes where the profile warranted it.
- [x] Review `appendObjRefToWriter` allocation behavior; it is no longer a top gate after HFT table compaction.
- [x] Validation gate: `formatStructElemObjectTo` flat CPU below 5% and alloc-space below 130 MB. Current flat CPU is 2.22% and alloc-space is 51.52 MB cumulative.

### P3 - HFT Shared Table Path

- [x] Profile `drawSharedLayoutRow` and `drawSharedDeferRow` in the refreshed CPU profile.
- [x] Add a table-row batch writer for shared-row MCID BDC/EMC emission.
- [x] Attach shared-row MCID leaves directly to the current `TR` parent to avoid allocating one `StructElem` per `TD`.
- [x] Keep the generic per-cell `BeginMarkedContentBufWithMCID` path for non-shared rows.
- [x] Validate HFT output compaction deliberately changes output size to `748163` bytes.
- [ ] Extend shared-row layout to precompute more stable cell drawing fragments.
- [ ] Validation gate: `drawTable` cumulative CPU below 18%. Current share is 25.94%, although absolute runtime and x10 throughput improved materially.

### P4 - Structure Pool Cleanup

- [x] Investigate `ReleaseStructElemsToPool`; the first bulk-drop attempt reduced release cost but regressed heap by forcing fresh HFT struct-node allocation.
- [x] Keep pooled struct nodes instead of dropping the large HFT element set.
- [x] Use a lighter reset helper for pooled nodes.
- [x] Remove hot shared-row `TD` nodes from the structure pool path by attaching MCID leaves directly to the row parent.
- [x] Validation gate: `ReleaseStructElemsToPool` cumulative CPU below 4%. Current cumulative CPU is 1.85%.

### P5 - Compression and Fingerprint Cost

- [x] Add a size threshold to skip compression-cache fingerprinting for large streams.
- [x] Keep compression caching for smaller repeated retail/active page content.
- [x] Verify compression CPU improved; zlib/flate close cumulative CPU is now 8.75%.
- [ ] Tune the threshold further; `pageContentFingerprint` cumulative CPU is still 3.51%.
- [ ] Reuse or pool flate writers more aggressively where the path still reaches `compress/flate.NewWriter`.
- [ ] Validation gate: `pageContentFingerprint` below 1.5% and `compress/flate` cumulative below 11%. Compression is below gate, fingerprint is not.

### P6 - PDF/A Font and Output Intent Reuse

- [x] Audit `font.GenerateTrueTypeFontObjects`.
- [x] Cache compressed font data by font bytes fingerprint.
- [x] Cache static sRGB ICC compressed payload for unencrypted output intent generation.
- [ ] Reduce remaining `font.GenerateTrueTypeFontObjects` CPU share below 8%; current cumulative CPU is 10.72%.
- [ ] Reduce font allocation-space further; current cumulative allocation is 192.49 MB.

### P7 - Signature Allocation Cleanup

- [x] Keep ECDSA as the default.
- [x] Replace `/Contents` signature hex construction with direct in-place hex encoding and zero fill.
- [x] Verify signature allocation-space improved. `signature.UpdatePDFWithSignatureBuffer` is now 37.54 MB cumulative.
- [ ] Inspect `CreateSignatureField`; it is still 115.68 MB cumulative allocation in the refreshed alloc-space profile.
- [x] Validation gate: signature cumulative alloc-space below 100 MB for `UpdatePDFWithSignatureBuffer`.

## Next Optimization Order

1. [ ] P3: reduce `drawSharedDeferRow` by precomputing stable row/cell drawing fragments after the MCID batch writer.
2. [ ] P5: lower `pageContentFingerprint` below 1.5% while keeping the compression CPU win.
3. [ ] P6: reduce `GenerateTrueTypeFontObjects` share, starting with CID map / ToUnicode allocation reuse.
4. [ ] P7: reduce `CreateSignatureField` allocation below 100 MB cumulative.
5. [ ] P3/P1: revisit HFT page stream growth only if future drawing-fragment work raises `bytes.growSlice` again.

## Validation Commands

```bash
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache make bench-gopdflib-zerodha-x10
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache make bench-gopdflib-zerodha-x5

go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -top guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -top -inuse_space guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof
go tool pprof -top -alloc_space guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof
```
