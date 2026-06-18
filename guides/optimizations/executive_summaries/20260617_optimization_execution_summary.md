# Optimization Execution Summary — 2026-06-17

## Date & Scope

On branch `feat/optimization-5.5-medium`, delivered Zerodha x10 pprof optimizations (commit `fff4f16`), then diagnosed and fixed a **k6 HTTP regression** caused by an unbounded shared-row render cache introduced for in-process throughput gains.

## Key Outcomes

**Zerodha in-process (x10, 48 workers, GOMAXPROCS=24)**

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Mean throughput | 2,690.73 ops/s | **7,438.64 ops/s** | **+176%** |
| Median throughput | 2,649.69 ops/s | **7,224.21 ops/s** | **+173%** |
| Best throughput | 3,028.74 ops/s | **10,531.99 ops/s** | **+248%** |
| Mean peak allocated | 1,269.97 MB | **585.84 MB** | **−54%** |
| Post-recovery single run | — | **12,675.43 ops/s** | Target ~13,000 |

**k6 HTTP harness (48 VU × 35s, `tagged_ecdsa`)**

| State | Throughput | Heap | `drawSharedLayoutRow` |
|-------|------------|------|----------------------|
| Before fix (feature branch) | hung ~16k iter | **1,603 MB** | **497 MB** |
| Master baseline | 674 req/s | 587 MB | — |
| After bounded-cache fix (full k6) | **1,223.3 req/s** | **521.6 MB** | **7.0 MB** |
| After fix (`bench-k6-light`) | **1,276.7 req/s** | **130.8 MB** | **3.5 MB** |

**Output sizes:** Retail **61,293 B**, Active **76,065 B** stable; HFT **2,289,155 → 748,163 B** (intentional MCID compaction).

## Work Completed

- **Zerodha x10 pprof optimization (P0–P7):** HFT shared-table fast path, structure serialization/pooling, compression fingerprinting, page buffer pooling, font/PDF-A reuse, signature allocation cleanup
- **Benchmark harness:** fixed Zerodha package build; added `bench-gopdflib-zerodha-x10-pprof`; refreshed x10 stats and pprof artifacts
- **k6 regression analysis:** root-caused hang/OOM to unbounded global `sharedRowRenderCache` (`sync.Map`, near-zero hits, unbounded growth under 48 concurrent HTTP requests)
- **Recovery fix:** replaced with bounded entry/byte-capped cache; slice copy on store; eviction; uncached fallback preserved

## Findings / Bottlenecks

**Pre-fix pprof wins (Zerodha in-process)**

| Hotspot | Baseline | Optimized |
|---------|----------|-----------|
| `drawTable` cumulative CPU | 21.94% | **17.95%** |
| `formatStructElemObjectTo` flat CPU | 7.57% | **1.37%** |
| `compress/flate` close cumulative CPU | 13.77% | **6.65%** |
| `bytes.growSlice` in-use | 432.84 MB | **219.35 MB** |

**k6 regression root cause:** Global `sharedRowRenderCache` — keys vary per draw → stores on almost every row, few hits. Under 48 concurrent HFT docs (~2,001 shared rows each): heap **~2.7× master**; server slots blocked.

## Open Items / Next Steps

- **Closed on 2026-06-17:** bounded shared-row cache recovery; full k6 stability restored
- **Monitor:** `drawTable` CPU variance on short pprof samples (17.95%–22.67%)
- **Operational:** use `bench-k6-light` as interim gate; run k6 in isolation on WSL
- **Throughput context:** post-fix k6 **~1,223 req/s** vs master **674 req/s** — further HTTP-path tuning may be warranted separately from Zerodha in-process gains

## Source Documents

| File | Role |
|------|------|
| `20260617_zerodha_x10_pprof_optimization_checklist.md` | P0–P7 checklist, before/after metrics |
| `PR/20260617_zerodha_x10_pprof_pr_description.md` | PR summary, implementation details |
| `20260617_k6_bench_regression_analysis.md` | Master vs feature-branch k6 comparison |
| `20260617_shared_row_cache_recovery_checklist.md` | Bounded-cache fix, acceptance criteria |