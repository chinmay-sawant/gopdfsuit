# Optimization Execution Summaries

Date-by-date executive summaries for all documents in `guides/optimizations/`. Generated **2026-06-18** from six parallel analysis passes.

## Timeline at a Glance

| Date | Focus | Headline Result | Documents |
|------|-------|-----------------|-----------|
| [2026-06-10](20260610_optimization_execution_summary.md) | Master plan, Phase 1–2 engine, SlopGuard, GoPDFKit parity | `BenchmarkGoPdfSuit` **−11.6%** ns/op, **−16%** B/op | 4 reports + PR context |
| [2026-06-11](20260611_optimization_execution_summary.md) | GoPDFKit fix, Zerodha gold standard, Gin HTTP, 3k push | gopdflib **7/7** workloads; Zerodha **1,135 ops/s**; Gin **1,232 req/s** peak | 4 reports |
| [2026-06-14](20260614_optimization_execution_summary.md) | k6 weighted bench, cache bounds, structure prealloc | **1,026 req/s** 2-run mean (+24% vs 825 baseline) | 1 checklist |
| [2026-06-17](20260617_optimization_execution_summary.md) | Zerodha x10 pprof, k6 regression recovery | **7,439 ops/s** mean (+176%); k6 restored to **1,223 req/s** | 4 reports |
| [2026-06-18](20260618_optimization_execution_summary.md) | PyPDFSuit Python binding profile | **253 → 1,505 ops/s** (~5.9×) via JSON cache | 1 report |
| [Cross-cutting](cross_cutting_optimization_execution_summary.md) | SlopGuard remediation, PR phases 1–10 | **218/218** PERF findings fixed; v6 release prep | 2 docs |

## Cumulative Progress Toward Targets

| Target | Baseline | Best Observed | Status |
|--------|----------|---------------|--------|
| Gin weighted 1,500 req/s (80/15/5) | ~825 req/s | **1,232 req/s** (2026-06-11) | ~82% of target |
| Zerodha in-process 3,000 ops/s | ~573 ops/s (Go 1.24) | **12,675 ops/s** (2026-06-17 recovery) | Exceeded |
| PyPDFSuit weighted throughput | 253 ops/s | **1,505 ops/s** (2026-06-18) | ~13% of native Go (11,721) |
| GoPDFKit parity (7 workloads) | 5/7 wins | **7/7 wins** (2026-06-11) | Met |

## Source Folder

All underlying reports live in [`../`](../) (parent `guides/optimizations/`).