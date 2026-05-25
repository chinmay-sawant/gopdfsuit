# Cursor Performance Guides

Performance architecture audit and implementation tracking for GoPdfSuit.

## Documents

| File | Purpose |
|------|---------|
| [PERFORMANCE_AUDIT.md](./PERFORMANCE_AUDIT.md) | Full 6-agent audit report |
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | Phased roadmap with task status |
| [PASS1_BLUEPRINTS.md](./PASS1_BLUEPRINTS.md) | Before/after code for Pass 1 |
| [baselines/bench_pass1_20260525.txt](./baselines/bench_pass1_20260525.txt) | Post-Pass-1 benchmark |
| [baselines/bench_pass2_20260525.txt](./baselines/bench_pass2_20260525.txt) | Post-Pass-2 benchmark |
| [PASS4_OPTIMIZATION_PLAN.md](./PASS4_OPTIMIZATION_PLAN.md) | Pass 4 — load-test hotspot plan |
| [PASS4_PDFA_RESULTS.md](./PASS4_PDFA_RESULTS.md) | **Pass 4 PDF/A results vs May 25 baseline** |
| [baselines/loadtest_pprof_summary_20260525.txt](./baselines/loadtest_pprof_summary_20260525.txt) | k6 load test pprof summary (pre-P4) |
| [baselines/bench_pass4_pdfa_x5_20260525.txt](./baselines/bench_pass4_pdfa_x5_20260525.txt) | PDF/A benchmark ×5 raw |
| [baselines/bench_pass4_pdfa_x5_stats_20260525.txt](./baselines/bench_pass4_pdfa_x5_stats_20260525.txt) | PDF/A benchmark ×5 best/worst/avg |
| [baselines/pprof_runs/](./baselines/pprof_runs/) | CPU pprof ×5 during Rows2000 bench |
| [GOPDFLIB_PPROF_RESULTS.md](./GOPDFLIB_PPROF_RESULTS.md) | GoPDFLib 5000× pprof + PDF/A bench |
| [baselines/gopdflib_pprof_stats_20260525.txt](./baselines/gopdflib_pprof_stats_20260525.txt) | GoPDFLib stats summary |
| [ZERODHA_BENCHMARK_RESULTS.md](./ZERODHA_BENCHMARK_RESULTS.md) | Zerodha 5000×5-run + pprof |
| [baselines/zerodha_bench_x5_stats_20260525.txt](./baselines/zerodha_bench_x5_stats_20260525.txt) | Zerodha 5-run stats |
| [baselines/zerodha_pprof_runs/](./baselines/zerodha_pprof_runs/) | Zerodha raw output + profiles |

## Quick Commands

```bash
# Standard benchmarks
go test -run='^$' -bench=BenchmarkGenerateTemplatePDF -benchmem ./internal/pdf/

# Macro scale (2K / 10K / 25K rows)
go test -run='^$' -bench=BenchmarkGenerateTemplatePDF/Rows -benchmem -benchtime=5s ./internal/pdf/

# Wrap-enabled path
go test -run='^$' -bench=BenchmarkGenerateTemplatePDF_WrapEnabled -benchmem ./internal/pdf/

# CPU profile
go test -run='^$' -bench=BenchmarkGenerateTemplatePDF/Rows2000 -benchtime=30s \
  -cpuprofile=/tmp/cpu.prof ./internal/pdf/
go tool pprof -http=:8081 /tmp/cpu.prof

# All tests
go test ./internal/pdf/... ./internal/handlers/...
```

## Status

**Pass 1–3:** Complete (2026-05-25)  
**Pass 4 (A–D):** Complete (2026-05-25)  
**PDF/A validation:** [PASS4_PDFA_RESULTS.md](./PASS4_PDFA_RESULTS.md)

See [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md), [PASS4_OPTIMIZATION_PLAN.md](./PASS4_OPTIMIZATION_PLAN.md), and [PASS3_BLUEPRINTS.md](./PASS3_BLUEPRINTS.md).
