# Optimization Execution Summary — 2026-06-20

## Date & Scope

Compliant Zerodha gopdflib x10 pprof optimization on `feat/optimization-5.5-medium` (post `80541ca` HFT TR→TD hardening). Target: **≥ 8,000 ops/s** mean on `make bench-gopdflib-zerodha-x10` while preserving PDF/A-4, PDF/UA-2, retail ECDSA signing, and full HFT **TR → TD** hierarchy — without key-based cross-request caches.

## Key Outcomes

**Throughput (compliant TR→TD, x10 idle Phase 4)**

| Metric | Pre-opt (2026-06-20) | Phase 4 final | Change |
|--------|---------------------:|--------------:|-------:|
| Mean throughput | 2,799.13 ops/s | **9,594.17 ops/s** | **+243%** |
| Best throughput | 3,271.57 ops/s | **10,004.50 ops/s** | **+206%** |
| Median throughput | 2,966.46 ops/s | **9,680.69 ops/s** | **+226%** |
| Mean peak allocated | 1,337.15 MB | **1,198.81 MB** | −10% |

- **8,000 ops/s mean target: met** (+19.9% margin).
- **veraPDF:** **6/6 PASS** (retail/active/HFT × PDF/A-4 + PDF/UA-2) after every phase.
- **HFT output:** **2,291,950 bytes** (compliant TR→TD; within ±5% of baseline).
- **k6 guardrail:** `make bench-k6-light` must not regress vs 2026-06-17 recovery baseline.

**Phase progression:** 2,799 → 5,543 (P0–P8) → 7,432 (P17–P20) → **9,594** (P21–P25).

## Work Completed

- **Phase 1 (P0–P8):** HFT shared-table fast path, structure serialization pooling, compression fingerprinting, page buffer pooling, font/PDF-A reuse, signature allocation cleanup — **+88%** vs pre-opt.
- **Phase 2 (P9–P16):** Lazy arena activation, tiered `arenaCapForNeed()`, split pdfBuffer pools, `WarmRuntimePools()` — fixed P12 regression (4,547 ops/s → 5,543).
- **Phase 3 (P17–P20):** Batch arena TD, `tdLeafFast`, xref pre-size, arena pool direct-return race fix.
- **Phase 4 (P21–P25):** Pprof-driven close-out items; idle-machine confirmation of 8K gate.

## Findings / Bottlenecks

| Hotspot (post-opt CPU) | Notes |
|------------------------|-------|
| `drawTable` / `drawSharedLayoutRow` | HFT-heavy; still material after P0–P25 |
| `bytes.growSlice` + arena slabs | Heap pressure; mean peak still **~1.2 GB** vs 600 MB aspirational |
| `formatStructElemObjectTo` | Struct emit on tagged PDFs |
| `buildSRGBICCProfile` | ~4% cum — P9 incomplete; deferred to 15K checklist as **P26** |

- **2026-06-17 non-compliant peak (10,532 / 12,675 ops/s)** used HFT output **748 KB** (collapsed TR→TD) — explicitly out of scope.
- Compliant rebuild from 2,799 ops/s cost **~3.4×** vs the non-compliant fast path but restored veraPDF acceptance.

## Open Items / Next Steps

| Priority | Item | Target | Status |
|----------|------|--------|--------|
| **Next** | 15K checklist (`21062026_zerodha_15k_optimization_checklist.md`) | **15,000 ops/s** from idle **9,009** | Planning |
| **P26** | sRGB ICC cache fix | +550 ops/s (6/6 agent consensus) | Not started |
| **Phase B** | pdfBuffer zero-grow, arena slab sizing | ~13,000 ops/s projected | Planned |
| **Phase C** | HFT TR→TD deep path (P36–P40) | ~15,000 ops/s projected | Planned |

## Source Documents

| File | Role |
|------|------|
| `20260620_zerodha_x10_pprof_optimization_checklist.md` | Primary checklist (P0–P25) |
| `guides/cursor/baselines/zerodha_bench_x10_wsl/` | x10 run artifacts |
| `guides/cursor/baselines/zerodha_pprof_runs/` | CPU and heap profiles |
| `20260617_k6_bench_regression_analysis.md` | Key-based cache guardrail context |