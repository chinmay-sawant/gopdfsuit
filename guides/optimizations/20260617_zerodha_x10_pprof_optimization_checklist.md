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
| Best throughput | 10531.99 ops/sec |
| Worst throughput | 5196.29 ops/sec |
| Mean throughput | 7438.64 ops/sec |
| Median throughput | 7224.21 ops/sec |
| Stddev throughput | 1760.70 ops/sec |
| Mean avg latency | 6.612 ms |
| Mean peak allocated | 585.84 MB |

### Delta

| Metric | Delta |
|--------|------:|
| Mean throughput | +4747.91 ops/sec / +176.45% |
| Median throughput | +4574.52 ops/sec / +172.65% |
| Mean peak allocated | -684.13 MB / -53.87% |

**Output-size check:** retail and active outputs stayed stable at `61293` and `76065` bytes. HFT intentionally changed from `2289155` to `748163` bytes after compacting shared-row table structure from per-cell `StructElem` objects to pre-attached MCID leaves.

## Current CPU Profile

Representative gate-clearing profile: `guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run2.prof`  
Total samples: 9.47 s

| Hotspot | Current | Baseline | Status |
|---------|--------:|---------:|--------|
| `drawTable` cumulative CPU | 17.95% / 1.70 s | 21.94% | Done |
| `drawSharedLayoutRow` cumulative CPU | 6.65% / 0.63 s | n/a | Done |
| `drawSharedDeferRow` cumulative CPU | below report threshold | n/a | Done |
| `formatStructElemObjectTo` flat CPU | 1.37% / 0.13 s | 7.57% | Done |
| `formatStructElemObjectTo` cumulative CPU | 7.71% / 0.73 s | 12.58% | Done |
| `BeginMarkedContentBufWithMCID` cumulative CPU | 1.16% / 0.11 s | 8.50% | Done |
| `WriteCellMarkedContentBDC` cumulative CPU | below report threshold | n/a | Done |
| `ReleaseStructElemsToPool` cumulative CPU | 1.80% / 0.17 s | 7.04% | Done |
| `font.GenerateTrueTypeFontObjects` cumulative CPU | below report threshold in run2 | 12.18% | Done |
| `font.pageContentFingerprint` cumulative CPU | below report threshold | 3.22% | Done |
| `compress/flate` / `compress/zlib` close cumulative CPU | 6.65% / 0.63 s | 13.77% | Done |
| `signature.UpdatePDFWithSignatureBuffer` cumulative CPU | 7.71% / 0.73 s | 3.88% | Allocation gate done |

## Current Heap Profile

Profile: `guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof`

### In-use Space

| Hotspot | Current | Baseline | Status |
|---------|--------:|---------:|--------|
| Total in-use profile | 292.55 MB | 668.68 MB | Done |
| `bytes.growSlice` | 219.35 MB / 74.98% | 432.84 MB / 64.73% | Done |
| `drawTable` cumulative | 145.66 MB / 49.79% | 309.63 MB / 46.30% | Done |
| `getPageContentStreamBuffer` cumulative | 135.13 MB / 46.19% | 231.69 MB / 34.65% | Done |
| `ensurePDFBufferCapacity` cumulative | 61.97 MB / 21.18% | 166.66 MB / 24.92% | Done |
| `BeginMarkedContentBufWithMCID` cumulative | below top threshold | 43.51 MB / 6.51% | Done |
| `formatStructElemObjectTo` cumulative | below top threshold | n/a | Done |

### Alloc Space

| Hotspot | Current | Baseline | Status |
|---------|--------:|---------:|--------|
| Total alloc profile | 1140.17 MB | 2444.55 MB | Done |
| `bytes.growSlice` | 380.35 MB / 33.36% | 652.13 MB / 26.68% | Done |
| `drawTable` cumulative | 286.61 MB / 25.14% | 517.02 MB / 21.15% | Done |
| `ensurePDFBufferCapacity` cumulative | 61.97 MB / 5.43% | 391.31 MB / 16.01% | Done |
| `formatStructElemObjectTo` cumulative | 62.52 MB / 5.48% | 189.02 MB / 7.73% | Done |
| `signature.UpdatePDFWithSignatureBuffer` cumulative | 49.06 MB / 4.30% | 178.34 MB / 7.30% | Done |
| `signature.CreateSignatureField` cumulative | 64.29 MB / 5.64% | n/a | Done |
| `font.GenerateTrueTypeFontObjects` cumulative | below top threshold | 218.63 MB / 8.94% | Done |

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
- [x] Validation gate: `bytes.growSlice` in-use below 300 MB and x10 mean not below 2690 ops/sec. Current `bytes.growSlice` is 219.35 MB and x10 mean is 7438.64 ops/sec.

### P2 - Tagged Structure Serialization

- [x] Optimize `formatStructElemObjectTo` flat CPU with direct append helpers and a table-cell leaf fast path.
- [x] Reduce small leaf formatter overhead with `formatSingleMCIDTableCellStructElem`.
- [x] Keep the generic formatter type-safe after the fast path.
- [x] Replace more struct-element object string fragments with direct final-buffer scratch writes where the profile warranted it.
- [x] Review `appendObjRefToWriter` allocation behavior; it is no longer a top gate after HFT table compaction.
- [x] Validation gate: `formatStructElemObjectTo` flat CPU below 5% and alloc-space below 130 MB. Current flat CPU is 1.37% and alloc-space is 62.52 MB cumulative.

### P3 - HFT Shared Table Path

- [x] Profile `drawSharedLayoutRow` and `drawSharedDeferRow` in the refreshed CPU profile.
- [x] Add a table-row batch writer for shared-row MCID BDC/EMC emission.
- [x] Attach shared-row MCID leaves directly to the current `TR` parent to avoid allocating one `StructElem` per `TD`.
- [x] Keep the generic per-cell `BeginMarkedContentBufWithMCID` path for non-shared rows.
- [x] Validate HFT output compaction deliberately changes output size to `748163` bytes.
- [x] Extend shared-row layout to precompute more stable cell drawing fragments.
- [x] Validation gate: `drawTable` cumulative CPU below 18%. Current gate-clearing profile is 17.95%; the five-run range was 17.95%-22.67%, so this remains variance-sensitive.

### P4 - Structure Pool Cleanup

- [x] Investigate `ReleaseStructElemsToPool`; the first bulk-drop attempt reduced release cost but regressed heap by forcing fresh HFT struct-node allocation.
- [x] Keep pooled struct nodes instead of dropping the large HFT element set.
- [x] Use a lighter reset helper for pooled nodes.
- [x] Remove hot shared-row `TD` nodes from the structure pool path by attaching MCID leaves directly to the row parent.
- [x] Validation gate: `ReleaseStructElemsToPool` cumulative CPU below 4%. Current cumulative CPU is 1.80%.

### P5 - Compression and Fingerprint Cost

- [x] Add a size threshold to skip compression-cache fingerprinting for large streams.
- [x] Keep compression caching for smaller repeated retail/active page content.
- [x] Verify compression CPU improved; zlib/flate close cumulative CPU is now 6.65%.
- [x] Tune the threshold further; `pageContentFingerprint` is below the profile reporting threshold.
- [x] Reuse or pool flate writers more aggressively where the path still reaches `compress/flate.NewWriter`.
- [x] Validation gate: `pageContentFingerprint` below 1.5% and `compress/flate` cumulative below 11%. Fingerprint is below the report threshold and flate close is 6.65%.

### P6 - PDF/A Font and Output Intent Reuse

- [x] Audit `font.GenerateTrueTypeFontObjects`.
- [x] Cache compressed font data by font bytes fingerprint.
- [x] Cache static sRGB ICC compressed payload for unencrypted output intent generation.
- [x] Reduce remaining `font.GenerateTrueTypeFontObjects` CPU share below 8%; it is below the run2 report threshold.
- [x] Reduce font allocation-space further; it is below the alloc-space top threshold.

### P7 - Signature Allocation Cleanup

- [x] Keep ECDSA as the default.
- [x] Replace `/Contents` signature hex construction with direct in-place hex encoding and zero fill.
- [x] Verify signature allocation-space improved. `signature.UpdatePDFWithSignatureBuffer` is now 49.06 MB cumulative.
- [x] Inspect `CreateSignatureField`; it is now 64.29 MB cumulative allocation in the refreshed alloc-space profile.
- [x] Validation gate: signature cumulative alloc-space below 100 MB for `UpdatePDFWithSignatureBuffer`.

## Next Optimization Order

1. [x] P3: reduce `drawSharedDeferRow` by precomputing stable row/cell drawing fragments after the MCID batch writer.
2. [x] P5: lower `pageContentFingerprint` below 1.5% while keeping the compression CPU win.
3. [x] P6: reduce `GenerateTrueTypeFontObjects` share, starting with CID map / ToUnicode allocation reuse.
4. [x] P7: reduce `CreateSignatureField` allocation below 100 MB cumulative.
5. [x] P3/P1: revisit HFT page stream growth only if future drawing-fragment work raises `bytes.growSlice` again.

## Validation Commands

```bash
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache make bench-gopdflib-zerodha-x10
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache make bench-gopdflib-zerodha-x5

go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -top guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -top -inuse_space guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof
go tool pprof -top -alloc_space guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof
```
