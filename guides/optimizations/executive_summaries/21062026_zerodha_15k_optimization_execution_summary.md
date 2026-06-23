# Optimization Execution Summary — 2026-06-22

## Date & Scope

Planning and partial implementation for **Zerodha x10 mean ≥ 15,000 ops/s** on `make bench-gopdflib-zerodha-x10` (`feat/optimization-5.5-medium`, commit `87009f1`). Builds on Phase 4 close-out (**9,009 idle** / **9,594** prior peak session) with four-subagent `run2` profile refresh and 46-profile commit sweep (`gin`, `zerodha`, `pypdfsuit`).

## Key Outcomes

**Target gap (idle baseline)**

| Metric | Current | Target | Gap |
|--------|--------:|-------:|----:|
| x10 mean | 9,009 ops/s | ≥ 15,000 | **−5,991 (−40%)** |
| x10 stddev | 1,333 | ≤ 400 | −933 |
| Mean peak allocated | 1,077 MB | ≤ 650 MB | −427 MB |

**Phased projection**

| Phase | Items | Projected mean | Δ vs 9,009 |
|-------|-------|---------------:|-----------:|
| A — Quick wins | P26, P28, P30, P29 | 10,400–10,900 | +15–21% |
| B — Memory wall | P31, P34, P40, P32, P33, P35 | 12,500–13,500 | +39–50% |
| C — HFT tail | P36, P38, P37, P39, P41, P42 | 14,500–15,500 | +61–72% |

**Phase A implementation (2026-06-22, code landed; gates pending)**

| Item | Status | Est. gain |
|------|--------|----------:|
| P26 sRGB ICC cache fix | **Landed** (+ unit tests) | +350–550 |
| P28 font precompute at template build | **Landed** | +150–250 |
| P30 static XMP metadata shell | **Landed** | +80–120 |
| P29 active `SharedRowLayout` | **Landed** (visual gate open) | +120–220 |

## Work Completed

- Refreshed checklist from four-subagent `cpu_zerodha_run2.prof` review: strengthened Phase B (memory-first), softened exact HFT latency math, demoted active-trader priority.
- Commit-wide `.prof` sweep: Zerodha x10 owns ordering; gin profiles inform k6 guardrails (signing, ICC, compression); pypdfsuit confirms memory-first story.
- **P26:** Cached uncompressed sRGB bytes at `init`; `GenerateOutputIntent` reuses compressed payload (`pdfa.go`, `metadata.go`).
- **P28:** Precomputed startup font hint on Zerodha templates (Helvetica-only).
- **P30:** Handler-local XMP fragment cache; emit-time patch for dates/ID only.
- **P29:** Enabled `SharedRowLayout` on active 41-row trade table in `bench.go`.

## Findings / Bottlenecks

**Fresh CPU (`cpu_zerodha_run2.prof`)**

| Hotspot | Cum % | Gate |
|---------|------:|------|
| `drawTable` | 29.6% | HFT-heavy |
| `drawSharedLayoutRow` | 13.0% | Shared-row replay |
| `runtime.memmove` / `memclr` | 12.0% / 11.4% flat | **Open** |
| TR→TD setup (`beginTableRowArena`) | 6.7% | Phase C |
| `GetSRGBICCProfile` | 4.2% | **P26 target** |

**Fresh heap (`heap_zerodha.prof`, 563 MB)**

| Hotspot | In-use | % |
|---------|-------:|--:|
| `bytes.growSlice` | 263 MB | 46.7% |
| Arena slabs | 247 MB | 43.8% |
| Page content streams (cum) | 157 MB | 27.9% cum |

**Hard guardrails:** No HFT TR→TD shortcuts; no key-based cross-request caches; bounded `sharedRowRenderCache` unchanged; veraPDF 6/6 after every phase; k6 must not regress.

## Open Items / Next Steps

**Phase A acceptance (pending measurement)**

- [ ] x10 mean ≥ **10,400 ops/s** (idle machine)
- [ ] `buildSRGBICCProfile` cum CPU ≤ 0.5%
- [ ] P29 visual + veraPDF active PASS
- [ ] Throughput gate after P26–P30 landing

**Phase B (next):** P31 pdfBuffer zero-grow, P34 page-stream caps, P40 row-stream direct append, P32 arena tiering, P35 signature cleanup.

**Phase C (HFT tail):** P36 arena TD template, P38 batch struct emit, P37 stripe-batch arena, P39 glyph dedupe, P41 retail row-batch PDF/UA, P42 hand-built PKCS#7 DER.

## Source Documents

| File | Role |
|------|------|
| `21062026_optimization/21062026_zerodha_15k_optimization_checklist.md` | Primary checklist (P26–P42, 15K target) |
| `21062026_optimization/21062026_subagent_cross_validation_report.md` | Six-agent validation + ranking |
| `20260620_zerodha_x10_pprof_optimization_checklist.md` | P0–P25 foundation (8K met) |
| `guides/cursor/baselines/zerodha_pprof_runs/` | `cpu_zerodha_run{1..5}.prof`, `heap_zerodha.prof` |