# Samples Catalog

Runnable and fixture samples under `sampledata/`. Fixture-only directories
have no `main.go`; drive them through the API (`bin/app` on `:8080`) or the
integration suite (`make test-integration`).

## Quick index

| Directory | Demonstrates | Run command |
|---|---|---|
| `sampledata/builder-snippets/` | Copy-clip JSON snippet: title table, bracketed cell, spacer, data table | POST `snippet.json` to `/api/v1/generate/template-pdf` (fixture, no `main.go`) |
| `sampledata/python/builder_fluent_sample.py` | Preferred pypdfsuit builder spelling (`Font(...).size(12).bold()...`) | `cd sampledata/python && python3 builder_fluent_sample.py` |
| `sampledata/financialreport/` | Full financial report template plus chart assets | POST `financial_report.json` to `/api/v1/generate/template-pdf` |
| `sampledata/merge/` | Merge fixtures (`em-16.pdf`, `em-19.pdf`, `em-51.pdf`) | POST to `/api/v1/merge` or `make test-integration` (fixture, no `main.go`) |
| `sampledata/split/` | Split fixtures (`em.pdf`, range and max-per-file outputs) | POST to `/api/v1/split` or `make test-integration` (fixture, no `main.go`) |
| `sampledata/compress/` | `gopdflib.CompressPDF` Light/Medium/Heavy tiers on `report.pdf` | `cd sampledata/compress && go run .` |
| `sampledata/filler/` | AcroForm/XFDF fill plus compressed-stream variant in `filler/compressed/` | POST to `/api/v1/fill` or `make test-integration` |
| `sampledata/htmltopdf/` | Pure-Go HTML to PDF fixtures | POST to `/api/v1/htmltopdf` or `make test-integration` (fixture, no `main.go`) |
| `sampledata/htmltoimg/` | HTML to PNG fixtures | POST to `/api/v1/htmltoimage` (fixture, no `main.go`) |
| `sampledata/gopdflib/zerodha/` | Compliant Zerodha contract-note benchmark (PDF/A-4, PDF/UA-2, signing) | `cd sampledata && go run ./gopdflib/zerodha` or `make bench-gopdflib-zerodha` |
| `sampledata/typstsyntax/` | Typst math and content fixtures | POST JSON to `/api/v1/generate/template-pdf` (fixture, no `main.go`) |
| `sampledata/gopdflib/builder-snippets/` | Go mirror of the builder snippet with fluent `Font` chains | `cd sampledata && go run ./gopdflib/builder-snippets` |
| `sampledata/gopdflib/financial_report/` | Go financial report with tables, styling, bookmarks, timing | `cd sampledata && go run ./gopdflib/financial_report` |
| `sampledata/gopdflib/load_from_json/` | Load `sampledata/editor/financial_digitalsignature.json` into `gopdflib.PDFTemplate` | `go run sampledata/gopdflib/load_from_json/main.go` |
| `sampledata/gopdflib/text_wrapping/` | Auto text wrapping and row-height growth | `go run sampledata/gopdflib/text_wrapping/main.go` |
| `sampledata/gopdflib/typst_math/` | Typst math cells (`MathEnabled=true`) | `go run sampledata/gopdflib/typst_math/main.go` |
| `sampledata/gopdflib/zerodha_bops/` | Bypass-cache ops/sec benchmark, fresh template per op | `make bench-gopdflib-bops-x10` |
| `sampledata/pypdflib/builder-snippets/` | Python mirror of builder snippet via `pypdfsuit.builder` | `python3 sampledata/pypdflib/builder-snippets/main.py` |
| `sampledata/python/` | pypdfsuit scripts: `financial_report_pypdfsuit.py`, `JsonFileExample.py`, `builder_fluent_sample.py`, `test_redact.py` | `cd sampledata/python && python3 financial_report_pypdfsuit.py` |
| `sampledata/python/amazonReceipt/` | Receipt PDF with money rounding and barcode helpers | `cd sampledata/python/amazonReceipt && python3 amazonReceipt.py` |
| `sampledata/caching/` | Content-cache cold vs warm timing gap, TTL, manual clear | `cd sampledata && go run ./caching` |
| `sampledata/acroform/` | Healthcare AcroForm templates | POST to `/api/v1/fill` (fixture, no `main.go`) |
| `sampledata/editor/` | Editor fixtures: signatures, encryption, bookmarks and links JSON | POST JSON to `/api/v1/generate/template-pdf` (fixture, no `main.go`) |
| `sampledata/svg/` | Math SVG assets plus `generate_math_svg.py` | `cd sampledata/svg && python3 generate_math_svg.py` |
| `sampledata/benchmarks/` | Cross-engine harnesses plus `run_all_benchmarks.sh` | `cd sampledata/benchmarks && ./run_all_benchmarks.sh` |

## Notes

- Go samples resolve `github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib`
  through `sampledata/go.mod`. Run them from inside `sampledata/`.
- Python samples need the `pypdfsuit` wheel plus built `libgopdfsuit.so`
  (see `bindings/python/`). Some scripts additionally need the server
  (`make build && ./bin/app`, Gin on `:8080`).
- Fixture-only directories have no `package main`. The runnable path is the
  API or `make test-integration` (`go test -count=1 -v ./test`).
