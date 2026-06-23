# Optimization Execution Summaries

Date-by-date executive summaries for all documents in `guides/optimizations/`. Generated **2026-06-23** from ten parallel analysis passes.

## Timeline at a Glance

| Date | Focus | Headline Result | Documents |
|------|-------|-----------------|-----------|
| [2026-06-10](20260610_optimization_execution_summary.md) | Master plan, Phase 1–2 engine, SlopGuard, GoPDFKit parity | `BenchmarkGoPdfSuit` **−11.6%** ns/op, **−16%** B/op | 4 reports + PR context |
| [2026-06-11](20260611_optimization_execution_summary.md) | GoPDFKit fix, Zerodha gold standard, Gin HTTP, 3k push | gopdflib **7/7** workloads; Zerodha **1,135 ops/s**; Gin **1,232 req/s** peak | 4 reports |
| [2026-06-14](20260614_optimization_execution_summary.md) | k6 weighted bench, cache bounds, structure prealloc | **1,026 req/s** 2-run mean (+24% vs 825 baseline) | 1 checklist |
| [2026-06-17](20260617_optimization_execution_summary.md) | Zerodha x10 pprof, k6 regression recovery | **7,439 ops/s** mean (+176%); k6 restored to **1,223 req/s** | 4 reports |
| [2026-06-18](20260618_optimization_execution_summary.md) | PyPDFSuit Python binding profile | **235 ops/s** honest bench (`BENCH_USE_JSON_CACHE=0`); optional cache ~1,505 ops/s | 1 report |
| [2026-06-19](20260619_optimization_execution_summary.md) | PyPDFSuit honest benchmark, serializer cleanup | **836 ops/s** mean full path (P4 x5); cache tuning removed from benchmark surface | 1 checklist |
| [2026-06-20](20260620_optimization_execution_summary.md) | Zerodha compliant TR→TD path to 8K | **9,594 ops/s** mean (+243% vs 2,799 pre-opt); **8K target met**; veraPDF **6/6** | 1 checklist |
| [2026-06-21 (cross-val)](21062026_cross_validation_optimization_execution_summary.md) | Six-agent 15K plan cross-validation | **P26 ICC fix** first (6/6 consensus); HFT tail mandatory; idle **9,009** baseline gate | 1 report |
| [2026-06-21 (15K)](21062026_zerodha_15k_optimization_execution_summary.md) | Zerodha 9K → 15K phased checklist | **−5,991 ops/s** gap to 15K; Phase A→B→C model to ~15K | 1 checklist |
| [Cross-cutting](cross_cutting_optimization_execution_summary.md) | SlopGuard remediation, PR phases 1–10 | **218/218** PERF findings fixed; v6 release prep | 2 docs |

## Cumulative Progress Toward Targets

| Target | Baseline | Best Observed | Status |
|--------|----------|---------------|--------|
| Gin weighted 1,500 req/s (80/15/5) | ~825 req/s | **1,232 req/s** (2026-06-11) | ~82% of target |
| Zerodha in-process 3,000 ops/s | ~573 ops/s (Go 1.24) | **9,594 ops/s** (2026-06-20, compliant TR→TD) | Exceeded |
| Zerodha in-process 8,000 ops/s (compliant) | 2,799 ops/s (2026-06-20 pre-opt) | **9,594 ops/s** (2026-06-20 Phase 4 idle) | Exceeded (+19.9%) |
| Zerodha in-process 15,000 ops/s | 9,009 ops/s (idle) | **9,594 ops/s** (2026-06-20 peak session) | ~64% of target |
| PyPDFSuit weighted throughput (honest full path) | 221 ops/s (2026-06-19 pre-pass) | **836 ops/s** (2026-06-19 P4 x5) | ~7% of native Go (11,721) |
| GoPDFKit parity (7 workloads) | 5/7 wins | **7/7 wins** (2026-06-11) | Met |

## Source Folder

All underlying reports live in [`../`](../) (parent `guides/optimizations/`).