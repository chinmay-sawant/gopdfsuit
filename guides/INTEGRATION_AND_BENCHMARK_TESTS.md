# Integration & Benchmark Tests

**Date:** 2026-06-11 (updated 22:08 IST)

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

## End-to-end stack benchmarks (2-run mean, 2026-06-11 22:01)

| Engine | Harness | Throughput (peak) | 2-run avg | Prior (`19:45`) |
|--------|---------|------------------:|----------:|----------------:|
| **gopdflib** | Zerodha `go run .` | 2,463 ops/s | **2,459 ops/s** | 3,453 ops/s |
| **gopdfsuit** | k6 `tagged_ecdsa` | 693 req/s | **652 req/s** | 893 req/s |
| **pypdfsuit** | `pypdfsuit_bench.py` | 228 ops/s | **219 ops/s** | 236 ops/s |

Full comparison: [benchmark_suite_20260611_2128/comparison.md](./cursor/baselines/benchmark_suite_20260611_2128/comparison.md)

## Gin HTTP load benchmarks (`make load-pprof*`)

| Target | Command | Latest run | Throughput | Median | p99 |
|--------|---------|------------|----------:|-------:|----:|
| Weighted steady | `make load-pprof` | `20260611_220146` | **652 req/s** (2-run avg) | 20.1 ms | 467 ms |
| Weighted steady (best) | `make load-pprof` | `20260611_190806` | **1,054 req/s** | 15.5 ms | 143 ms |
| Weighted ≥1k gate | `make load-pprof-1k` | `20260611_190935` | 953 req/s | — | 150 ms |
| Weighted ≥1.5k gate | `make load-pprof-1500` | `20260611_191020` | 871 req/s | 18.6 ms | 176 ms |
| Retail-only ≥1.5k gate | `make load-pprof-gate` | `20260611_190850` | **3,965 req/s** | 7.8 ms | 29 ms |

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