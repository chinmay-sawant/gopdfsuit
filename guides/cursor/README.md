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

**Pass 1:** Complete (2026-05-25)  
**Pass 3:** Complete (2026-05-25)  
**All phases complete.**

See [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) and [PASS3_BLUEPRINTS.md](./PASS3_BLUEPRINTS.md).
