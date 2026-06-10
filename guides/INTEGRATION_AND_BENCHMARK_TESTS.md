# Integration & Benchmark Tests

**Date:** 2026-06-11

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

Output PDF from integration test: `sampledata/financialreport/financial_report.pdf`

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