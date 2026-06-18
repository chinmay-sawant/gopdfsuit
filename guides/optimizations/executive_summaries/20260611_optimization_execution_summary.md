# Optimization Execution Summary — 2026-06-11

## Date & Scope

Full-day performance validation on **Go 1.26.4** (i7-13700HX, 24 logical cores): library-vs-library micro-benchmarks, Zerodha gold-standard in-process workload, Gin HTTP load tests with pprof, and the Zerodha **3,000 ops/s** push. All work targets PDF/A + tagged + signed generation under an **80% retail / 15% active / 5% HFT** mix.

## Key Outcomes

- **GoPDFKit comparison (corrected):** gopdflib wins **7/7** workloads. Median throughput lead **+40% to +645%** (e.g. `text_short`: 160,863 vs 106,256 pdf/s; `png_rows_60`: 30,018 vs 4,028 pdf/s). Evening re-run after PDF/UA-2 TD fix: **+27–86%** further gain (`text_short` **204,214** pdf/s).
- **Zerodha gold standard (in-process):** **1,134.71 ops/s** median (41.3 ms avg) — **+98%** vs Go 1.24 baseline (573.64 ops/s); exceeds Zerodha production cluster (~1,000 ops/s).
- **Zerodha 3,000 ops/s target:** **Not met.** Peak **2,751 ops/s**, 5-run avg **2,476 ops/s** (~8% below 3,000 peak).
- **Gin HTTP weighted (80/15/5):** **≥1,000 req/s gate passed** — peak **1,232 req/s** (Phase 11), post-revert baseline **1,054 req/s** (0% errors).
- **Gin HTTP 1,500 req/s weighted target:** **Not met** (871–1,232 req/s; ±20% machine variance).
- **Gin retail-only:** **3,965 req/s** (median 7.8 ms, p99 29 ms) — **≥1,500 gate passed**.

## Work Completed

- **GoPDFKit harness fix:** Redirected broken `go.mod` replace to real `gopdfkit v0.5.2`; re-ran 3-run median + evening 2-run mean post PDF/UA-2 Table→TR→TD hierarchy fix.
- **Zerodha gold-standard benchmark:** 3-run median on 48 workers × 5,000 iterations.
- **Gin Phases 0–11:** Measurement harness, concurrency alignment (48 workers), ECDSA signing, JSON ingress (sonic pretouch, tier prealloc, HFT split decode), PDF hot path (struct direct-write, batch MC, HFT deferral, page striping, sharded compress cache), flate tuning.
- **Gin Phase 12 experiments reverted:** CRC32 fingerprint, in-place sig hex, store-uncompressed large pages, 2× compress workers — no meaningful E2E gain.
- **Zerodha 3k Phases A–F + G1/G2:** Signature caching, ECDSA P-256, compression pooling/thresholds, buffer pre-sizing, harness stabilization (`BENCH_SEED=42`), structure + HFT shared layout, page compress cache (G2).
- **Reverted harmful opts:** G3 parallel structure-tree build (~60% regression); G4 template PDF cache removed.

## Findings / Bottlenecks

- **GoPDFKit vs gopdflib:** Largest advantages on PNG workloads (+237–645%) and large tables (+108–128%); GoPDFKit allocates **5–35×** more bytes on heavy workloads.
- **Zerodha in-process:** RSA signing was **~10% flat CPU** on 80% retail; zlib **~22% cum**; `drawTable` **~19% cum** (HFT variance drives ±20% run-to-run swing).
- **Gin HTTP pprof (post-Phase 11):** HFT (5% traffic) dominates weighted avg — **222–262 ms** avg per HFT request; **flate ~35% cum CPU**; sonic JSON heavy on alloc (~51% alloc_space).
- **Run variance:** Same binary/config can swing **±20%** on weighted Gin runs.

## Open Items / Next Steps

- **Zerodha 3,000 ops/s:** Need **~16 ms** avg latency — further flate tuning on HFT stripes, HFT draw/compress path.
- **Gin 1,500 req/s weighted:** HFT tail (5% × ~250 ms) is primary ceiling; flate on HFT stripes, HFT JSON/decode, optional sonic codegen unmarshaler.
- **GoPDFKit gap:** Smallest remaining leads on text workloads (+40–51%).
- **Measurement hygiene:** `BENCH_SIGN_RSA` unset for ECDSA defaults; `BENCH_SEED=42` for reproducible HFT counts.

## Source Documents

- `20260611_gopdfkit_fixed_compare_results.md`
- `20260611_zerodha_gold_standard_benchmark.md`
- `20260611_gin_1500_ops_pprof_report.md`
- `20260611_zerodha_3000_ops_pprof_report.md`