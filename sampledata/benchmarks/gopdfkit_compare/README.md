# GoPDFKit vs gopdfsuit Benchmarks

This directory is a separate Go module for apples-to-apples comparison against
`github.com/chinmay-sawant/gopdfsuit/v5`.

The module pins:

- GoPDFKit: local checkout through `replace github.com/cssbruno/gopdfkit => ../..`
- gopdfsuit: `e61b05028120937d62408ca700d10a41f48e3899`

Run from this directory:

```shell
go test -run '^TestComparableOutputsArePDF$' -bench 'Benchmark(GoPDFKit|GoPDFLib)' -benchmem -count=3
```

The comparable workloads intentionally use features both libraries expose
through public API:

- `table_180_rows`
- `table_900_rows`
- `text_short`
- `text_240_lines`
- `invoice_40_rows`
- `png_table_180_rows`
- `png_rows_60`

Each workload is run in `workers_40` mode. The harness reports
`workers`, `pdf_bytes`, `pdf/s`, and `total_MB` custom metrics in addition to
the standard Go benchmark timing and allocation metrics.

HTML conversion is included as an opt-in benchmark because the implementations
are not equivalent: GoPDFKit renders its supported HTML subset in-process, while
gopdfsuit uses Chrome/Chromium. Enable it only on machines with Chrome:

```shell
GOPDFKIT_COMPARE_HTML=1 go test -run '^$' -bench 'HTML' -benchmem -count=3
```

Compliance workloads are not included in the apples-to-apples table unless both
libraries can be configured with equivalent PDF/A, PDF/UA, Arlington, metadata,
font, signing, and external validation behavior. Document unsupported or
non-equivalent features instead of mixing them into throughput rows.

## Local Snapshot

Run on `12th Gen Intel(R) Core(TM) i7-12700` with 20 logical CPUs. Rows are
from `make bench-gopdfsuit`.

| Workload | Library | Mode | ns/PDF | PDF/sec | Memory/PDF | Allocs/PDF | Output size | Total allocated |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| text_short | GoPDFKit | workers_40 | 4,355 | 229,606 | 33,430 B | 185 | 1,505 B | 8,574 MB |
| text_short | gopdflib | workers_40 | 7,427 | 134,641 | 104,269 B | 126 | 4,303 B | 16,905 MB |
| text_240_lines | GoPDFKit | workers_40 | 36,304 | 27,546 | 303,507 B | 568 | 4,706 B | 9,381 MB |
| text_240_lines | gopdflib | workers_40 | 90,125 | 11,096 | 1,028,017 B | 811 | 13,074 B | 12,824 MB |
| table_180_rows | GoPDFKit | workers_40 | 46,308 | 21,595 | 372,881 B | 874 | 8,043 B | 9,402 MB |
| table_180_rows | gopdflib | workers_40 | 100,742 | 9,926 | 699,241 B | 940 | 22,885 B | 6,668 MB |
| table_900_rows | GoPDFKit | workers_40 | 213,970 | 4,674 | 1,781,727 B | 3,685 | 34,997 B | 9,546 MB |
| table_900_rows | gopdflib | workers_40 | 488,571 | 2,047 | 3,140,994 B | 4,247 | 97,915 B | 7,848 MB |
| invoice_40_rows | GoPDFKit | workers_40 | 16,499 | 60,611 | 102,583 B | 348 | 3,232 B | 8,667 MB |
| invoice_40_rows | gopdflib | workers_40 | 29,455 | 33,950 | 208,499 B | 303 | 8,633 B | 8,684 MB |
| png_table_180_rows | GoPDFKit | workers_40 | 116,066 | 8,616 | 584,406 B | 1,048 | 15,784 B | 5,573 MB |
| png_table_180_rows | gopdflib | workers_40 | 105,201 | 9,506 | 719,850 B | 963 | 28,500 B | 7,737 MB |
| png_rows_60 | GoPDFKit | workers_40 | 176,534 | 5,665 | 660,280 B | 2,274 | 32,082 B | 4,758 MB |
| png_rows_60 | gopdflib | workers_40 | 174,381 | 5,735 | 1,868,885 B | 1,025 | 322,532 B | 10,594 MB |
