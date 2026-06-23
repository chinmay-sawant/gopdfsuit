# GoPDFKit vs gopdfsuit Benchmarks

This directory is a separate Go module for apples-to-apples comparison of
**GoPDFLib** (the PDF engine inside gopdfsuit) against external
**GoPDFKit** (`github.com/cssbruno/gopdfkit` v0.5.2).

The module pins:

- **gopdfsuit:** `github.com/chinmay-sawant/gopdfsuit/v6` (local `replace` to repo root)
- **GoPDFKit:** real checkout via `replace` → `/tmp/gopdfkit-real/.../gopdfkit@v0.5.2`

## Quick start (from repo root)

```shell
make bench-gopdfkit-setup        # download gopdfkit v0.5.2 + symlink (one-time)
make bench-gopdfkit-compare-test # verify both libraries emit valid PDFs
make bench-gopdfkit-compare      # timing run (default: benchtime=5s, count=1)
make bench-gopdfkit-compare-x2   # two sequential runs for stability
```

Override defaults:

```shell
make bench-gopdfkit-compare BENCH_TIME=10s BENCH_COUNT=3 GOMAXPROCS_BENCH=24
```

Run directly from this directory:

```shell
go test -run '^TestComparableOutputsArePDF$' -bench 'BenchmarkGoPDF(Kit|Lib)$' -benchmem -count=3
```

## Workloads

Comparable workloads use features both libraries expose through public API:

- `table_180_rows`
- `table_900_rows`
- `text_short`
- `text_240_lines`
- `invoice_40_rows`
- `png_table_180_rows`
- `png_rows_60`

Each workload runs in `workers_40` mode. The harness reports `workers`,
`pdf_bytes`, `pdf/s`, and `total_MB` custom metrics in addition to standard Go
benchmark timing and allocation metrics.

Templates intentionally use **PDF 1.7 without PDF/A flags** for fair speed
comparison. Production gopdfsuit also supports PDF/A-4 + PDF/UA-2, which
GoPDFKit does not match in this harness.

## HTML (opt-in)

HTML conversion is opt-in because implementations are not equivalent: GoPDFKit
renders its supported HTML subset in-process; gopdfsuit uses Chrome/Chromium.
Enable only on machines with Chrome:

```shell
make bench-gopdfkit-html
# or:
GOPDFKIT_COMPARE_HTML=1 go test -run '^$' -bench 'HTML' -benchmem -count=3
```

Compliance workloads (PDF/A, PDF/UA, Arlington, signing, etc.) are excluded
from the apples-to-apples table unless both libraries can be configured with
equivalent behavior.

## Results (best-of-5, benchtime=5s, 40 workers)

**Environment:** WSL2, 13th Gen Intel(R) Core(TM) i7-13700HX, 24 logical CPUs,
Go 1.26.4 - June 2026 suite (`make bench-gopdfkit-compare`, best pdf/s across 5 runs).

| Workload | GoPDFKit pdf/s | GoPDFLib pdf/s | gopdflib lead | Baseline Lib | Delta |
| --- | ---: | ---: | ---: | ---: | ---: |
| `text_short` | 119,959 | **254,986** | 2.1x | 206,298 | +24% |
| `text_240_lines` | 14,755 | **32,453** | 2.2x | 23,741 | +37% |
| `table_180_rows` | 11,883 | **47,707** | 4.0x | 28,870 | +65% |
| `table_900_rows` | 2,635 | **10,452** | 4.0x | 7,621 | +37% |
| `invoice_40_rows` | 40,145 | **135,052** | 3.4x | 105,514 | +28% |
| `png_table_180_rows` | 7,504 | **45,098** | 6.0x | 32,077 | +41% |
| `png_rows_60` | 5,474 | **53,935** | 9.9x | 42,548 | +27% |

**Winner:** GoPDFLib (gopdfsuit engine) on all 7 workloads. Lead ranges **2.1x**
(text) to **9.9x** (PNG rows).

Full benchmark report: [`guides/BENCHMARKS.md`](../../../guides/BENCHMARKS.md).