# Optimization Execution Summary — Cross-Cutting

## Scope

Cross-cutting performance work on `feat/performance-improvements` (undated in source docs; PR spans **2026-06-10 → 2026-06-16**). Covers the PDF engine, HTTP handlers, signing, benchmarks, validation, and release/CI — driven by SlopGuard static analysis and profile-guided optimization across the stack.

## Key Outcomes

- **SlopGuard:** 218 actionable PERF findings remediated (**100%**); 8 CWE-only items excluded from the perf pass
- **PDF micro-benchmarks:** `BenchmarkGoPdfSuit` **−11.6% latency, −16% allocations** vs `master` baseline
- **GoPDFKit parity:** gopdflib **7/7 workloads** (up to **+788%** on `png_rows_60`)
- **Gin HTTP (weighted tagged+ECDSA):** **~910–1,232 req/s** (≥1,000 gate met); retail-only **~3,965 req/s**
- **Zerodha E2E:** **~2,646 avg / 2,898 peak ops/s** (+362% vs Go 1.24 gold standard)
- **Long-lived pod safety:** Bounded caches for subset (1024), image (256), and props (8192)
- **Quality gates:** veraPDF PDF/A-4 + PDF/UA-2 validation wired into `make test`; v6.0.0 release prep complete

## Work Completed

**SlopGuard remediation (P1–P6 batches)**
- 226 exported findings tracked; fixes across `cmd/`, `internal/handlers/`, `internal/pdf/`, `typstsyntax/`, and `sampledata/`
- Recurring patterns: hoist regex compilation (PERF-1), replace `fmt.Sprintf` with `AppendInt`/`strings.Builder` (PERF-6/15/35), eliminate redundant conversions (PERF-32), map reuse with `clear()` (PERF-4), non-blocking logging (PERF-41)

**PDF engine (Phases 1–2, 7)**
- Buffer pooling, direct zlib streaming, `slices.Clone` instead of `append([]byte(nil), …)`
- Structure-tree optimizations: slice-based page maps, `BeginStructureElementCap`, bounded caches
- Tagged-PDF allocation reduction: compress-cache sharding, HFT row prealloc retention, pooled template `ResetForReuse()`

**HTTP & harness paths (Phases 5–6)**
- Gin: Sonic JSON decode, `GenerateTemplatePDFBorrowed`, ECDSA P-256, `GIN_FAST_API=1`, HFT decode fast path
- Zerodha: PEM/signer cache, 48-worker concurrency, flate pool, page compress cache

**Validation & release (Phases 8–10)**
- veraPDF post-test gate; XFDF `startxref` repair
- Cross-stack benchmark validation; CI `backend-test` job; Go **1.26.4** + module bump to **gopdfsuit/v6**

## Findings / Issues

- **Throughput variance:** Gin k6 results vary (825–1,232 req/s) with system load and run order
- **Intentional skips:** ~62 SlopGuard items in P6 marked unfixable/false-positive
- **Reverted experiments:** Phase 12 Gin CRC32/sig-hex, store-uncompressed pages, G3 parallel struct-tree, G4 template cache, aggressive buffer caps, generic structure writer
- **Dominant remaining costs:** `bytes.growSlice` (~50% heap), `compress/flate` (~20% CPU), `drawTable` (~15%), sonic JSON decode (~16% alloc_space)

## Open Items / Next Steps

| Target | Status |
|--------|--------|
| Gin weighted **1,500 req/s** | ~61% of target (best 2-run mean **1,025.7** req/s) |
| Zerodha **3,000 ops/s** | Exceeded in later x10 work (**12,675** ops/s) |
| Hot-path `fmt.Sprintf` in structure serialization | Partially addressed |
| Page content-buffer reuse, `drawTable`/`GetTextWidth` pressure | On remaining-optimizations checklist |
| HFT `drawTable` + zlib (~40% CPU from 5% of jobs) | Primary Zerodha bottleneck |

## Source Documents

| Document | Path |
|----------|------|
| SlopGuard findings tracker (P1–P6) | `slopguard-findings.md` |
| Performance PR description (Phases 1–10) | `PR/PR_DESCRIPTION.md` |