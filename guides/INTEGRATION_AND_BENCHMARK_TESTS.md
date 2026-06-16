# Integration & Benchmark Tests

**Date:** 2026-06-11 (updated 2026-06-14 01:10 IST)

## Run everything in one shot

```bash
# Full unit + integration (Go + Python)
make test-integration

# Go integration only (HTTP suite against Gin handlers)
go test -count=1 -v ./test

# Handler benchmarks (financial_report.json)
go test -bench=BenchmarkGenerateTemplatePDF_FinancialReport -benchmem ./test

# Handler unit tests with gomock
go test -v ./internal/handlers/... -run Mock

# Python integration
python3 -m pytest bindings/python/tests -v
```

## Gin handler benchmarks

| Benchmark | Input | Endpoint |
|-----------|-------|----------|
| `BenchmarkGenerateTemplatePDF_FinancialReport` | `sampledata/financialreport/financial_report.json` | `POST /api/v1/generate/template-pdf` |
| `BenchmarkGenerateTemplatePDF_FinancialReport_Parallel` | same | concurrent handler requests |

**Latest results** (2026-06-11 22:08, `GOMAXPROCS=24`, i7-13700HX, Go 1.26.4, 2-run avg):

| Benchmark | ns/op | MB/s | B/op | allocs/op | Prior (morning) |
|-----------|------:|-----:|-----:|----------:|----------------:|
| `BenchmarkGenerateTemplatePDF_FinancialReport` | **1,045,205** | 106.8 | 599,561 | 990 | 6,067,362 |
| `BenchmarkGenerateTemplatePDF_FinancialReport_Parallel` | **133,124** | — | 561,881 | 990 | 816,872 |

## End-to-end stack benchmarks

### gopdflib Zerodha (latest — 30-run, 2026-06-14)

| Engine | Harness | Peak | **30-run mean** | Prior (13-run `20260613`) |
|--------|---------|-----:|----------------:|--------------------------:|
| **gopdflib** | Zerodha `go run .` | **2,953 ops/s** | **2,646 ops/s** | 3,604 / 2,787 ops/s |

Full table: [ZERODHA_BENCHMARK_RESULTS.md](./cursor/ZERODHA_BENCHMARK_RESULTS.md)

### HTTP / Python (2-run mean, 2026-06-11 22:01)

| Engine | Harness | Peak | Latest avg | Prior (`19:45`) |
|--------|---------|-----:|-----------:|----------------:|
| **gopdfsuit** | k6 `tagged_ecdsa` | **859 req/s** | **825 req/s** (5-run, 2026-06-14) | 893 req/s |
| **pypdfsuit** | `pypdfsuit_bench.py` | 228 ops/s | **219 ops/s** | 236 ops/s |

Full comparison: [benchmark_suite_20260611_2128/comparison.md](./cursor/baselines/benchmark_suite_20260611_2128/comparison.md)

## Gin HTTP load benchmarks (`make load-pprof*`)

| Target | Command | Latest run | Peak | Latest avg | Median | p99 |
|--------|---------|------------|-----:|-----------:|-------:|----:|
| Weighted steady | `make bench-k6` | `20260614_004108` (5-run) | **859 req/s** | **825 req/s** | 16.0 ms | 347 ms |
| Weighted steady (best) | `make load-pprof` | `20260611_190806` | **1,054 req/s** | — | 15.5 ms | 143 ms |
| Weighted steady (2-run avg) | `make load-pprof` | `20260611_220146` | 693 req/s | **652 req/s** | 20.1 ms | 467 ms |
| Weighted ≥1k gate | `make load-pprof-1k` | `20260611_190935` | 953 req/s | — | — | 150 ms |
| Weighted ≥1.5k gate | `make load-pprof-1500` | `20260611_191020` | 871 req/s | — | 18.6 ms | 176 ms |
| Retail-only ≥1.5k gate | `make load-pprof-gate` | `20260611_190850` | **3,965 req/s** | — | 7.8 ms | 29 ms |

## Gotenberg HTML→PDF k6 (`make bench-gotenberg`)

Same weighted 80/15/5 mix as gopdfsuit, rendered via Gotenberg Chromium (`skipNetworkIdleEvent=true`). No PDF/A or signing.

| Target | Command | Latest run | Peak | Latest avg | Median | p99 |
|--------|---------|------------|-----:|-----------:|-------:|----:|
| Weighted steady | `make bench-gotenberg` | `20260613_215127` | — | **10.3 req/s** | 4.26 s | 8.22 s |

**vs gopdfsuit** (peak **859** / avg **825** req/s, 5-run `20260614`): ~**80×** higher avg throughput on the same k6 harness.  
Full comparison: [gotenberg_runs/comparison_20260613.md](./cursor/baselines/gotenberg_runs/comparison_20260613.md)

Output PDF from integration test: `sampledata/financialreport/financial_report.pdf`

## Caching & memory (long-running servers)

Request JSON is **not** retained after `POST /generate/template-pdf`. Process-global caches (compression, images, font subsets) have **no TTL** — see [CACHING_AND_MEMORY_LIFECYCLE.md](./CACHING_AND_MEMORY_LIFECYCLE.md).

## gomock handler tests

`internal/handlers/services.go` defines `PDFService`. Mocks live in `internal/handlers/mocks/`.

Regenerate mocks:

```bash
cd internal/handlers && go generate ./...
```

Unit tests (`handlers_gomock_test.go`) verify handler wiring without real PDF generation.

## Go HTTP integration coverage

| Feature | Test(s) | Sample data |
|---------|---------|-------------|
| Generate template PDF | `TestGenerateTemplatePDF`, `TestGenerateFinancialReportPDF`, Typst tests | `financial_report.json`, editor JSON, typst JSON |
| Fill XFDF | `TestFillPDF`, `TestFillPDFCompressed`, `TestFillPDFHospitalEncounter` | hospital + compressed medical forms |
| Merge | `TestMergePDFs` | `sampledata/merge/` |
| Split | `TestSplitPDF*`` | `sampledata/split/em.pdf` |
| HTML→PDF/Image | `TestHtmlToPDF`, `TestHtmlToImage` | Wikipedia URL (requires Chrome) |
| Redact | `TestRedactPageInfo`, `TestRedactCapabilities`, `TestRedactTextPositions`, `TestRedactSearch`, `TestRedactApply` | `financial_report.pdf` |
| Fonts | `TestGetFonts`, `TestUploadFontInvalidExtension` | — |
| Template data | `TestGetTemplateData` | `financial_report.json` |

## Python integration parity

| Feature | File |
|---------|------|
| Core suite | `bindings/python/tests/test_integration.py` |
| Financial report JSON | `bindings/python/tests/test_financial_report.py` |
| Compressed XFDF | `bindings/python/tests/test_xfdf_compressed.py` |
| Redaction | `bindings/python/tests/test_redact_financial_report.py` |