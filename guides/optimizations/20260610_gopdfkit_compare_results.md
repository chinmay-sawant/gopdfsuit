# GoPDFKit Comparison Results

**Date:** 2026-06-10  
**Go Version:** `go1.26.4`  
**Harness:** `sampledata/benchmarks/gopdfkit_compare/compare_benchmark_test.go`

## Command

```bash
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test -run '^$' -bench 'BenchmarkGoPDF(Kit|Lib)$' -benchmem -benchtime=5s
```

## Summary

`GoPDFKit` is currently faster on `6/7` workloads in the shared comparison harness.  
`gopdflib` wins only `png_table_180_rows`.

## Updated Summary After Profile-Backed Fixes

After pooling page content buffers in the library and caching prebuilt `gopdflib` templates in the comparison harness, the results tightened substantially.

`gopdflib` now wins `3/7` workloads:

- `text_short`
- `invoice_40_rows`
- `png_table_180_rows`

`GoPDFKit` still leads on:

- `text_240_lines`
- `table_180_rows`
- `table_900_rows`
- `png_rows_60`

## Throughput Results

| Workload | GoPDFKit pdf/s | gopdflib pdf/s | Winner |
|---|---:|---:|---|
| `text_short` | 95,768 | 71,356 | GoPDFKit |
| `text_240_lines` | 12,069 | 6,884 | GoPDFKit |
| `table_180_rows` | 10,088 | 6,988 | GoPDFKit |
| `table_900_rows` | 2,207 | 1,531 | GoPDFKit |
| `invoice_40_rows` | 33,554 | 25,305 | GoPDFKit |
| `png_table_180_rows` | 6,307 | 6,722 | gopdflib |
| `png_rows_60` | 3,867 | 3,238 | GoPDFKit |

## Updated Throughput Results

| Workload | GoPDFKit pdf/s | gopdflib pdf/s | Winner |
|---|---:|---:|---|
| `text_short` | 119,227 | 142,616 | gopdflib |
| `text_240_lines` | 14,067 | 13,969 | GoPDFKit |
| `table_180_rows` | 11,347 | 11,267 | GoPDFKit |
| `table_900_rows` | 2,544 | 2,248 | GoPDFKit |
| `invoice_40_rows` | 38,609 | 38,858 | gopdflib |
| `png_table_180_rows` | 6,570 | 9,857 | gopdflib |
| `png_rows_60` | 5,004 | 3,419 | GoPDFKit |

## Allocation and Size Signals

| Workload | GoPDFKit B/op | gopdflib B/op | GoPDFKit pdf_bytes | gopdflib pdf_bytes |
|---|---:|---:|---:|---:|
| `text_short` | 33,453 | 113,162 | 1,505 | 4,303 |
| `text_240_lines` | 303,627 | 1,041,077 | 4,706 | 13,074 |
| `table_180_rows` | 373,123 | 684,712 | 8,043 | 22,885 |
| `table_900_rows` | 1,782,462 | 3,034,030 | 34,997 | 97,915 |
| `invoice_40_rows` | 102,710 | 211,390 | 3,232 | 8,633 |
| `png_table_180_rows` | 585,028 | 708,920 | 15,784 | 28,500 |
| `png_rows_60` | 660,396 | 1,813,061 | 32,082 | 322,532 |

## Updated Allocation and Size Signals

| Workload | GoPDFKit B/op | gopdflib B/op | GoPDFKit pdf_bytes | gopdflib pdf_bytes |
|---|---:|---:|---:|---:|
| `text_short` | 33,439 | 28,642 | 1,505 | 4,303 |
| `text_240_lines` | 303,573 | 64,008 | 4,706 | 13,074 |
| `table_180_rows` | 373,005 | 96,244 | 8,043 | 22,885 |
| `table_900_rows` | 1,782,121 | 360,459 | 34,997 | 97,915 |
| `invoice_40_rows` | 102,685 | 43,883 | 3,232 | 8,633 |
| `png_table_180_rows` | 584,382 | 115,924 | 15,784 | 28,500 |
| `png_rows_60` | 660,171 | 1,541,383 | 32,082 | 322,532 |

## Latest Throughput Results

| Workload | GoPDFKit pdf/s | gopdflib pdf/s | Winner |
|---|---:|---:|---|
| `text_short` | 123,948 | 122,041 | GoPDFKit |
| `text_240_lines` | 11,573 | 11,823 | gopdflib |
| `table_180_rows` | 10,073 | 9,668 | GoPDFKit |
| `table_900_rows` | 2,259 | 2,054 | GoPDFKit |
| `invoice_40_rows` | 31,221 | 36,889 | gopdflib |
| `png_table_180_rows` | 5,806 | 8,904 | gopdflib |
| `png_rows_60` | 4,006 | 3,492 | GoPDFKit |

## Latest Allocation Signals

| Workload | GoPDFKit B/op | gopdflib B/op |
|---|---:|---:|
| `text_short` | 33,438 | 31,667 |
| `text_240_lines` | 303,537 | 65,729 |
| `table_180_rows` | 372,982 | 97,378 |
| `table_900_rows` | 1,781,656 | 365,019 |
| `invoice_40_rows` | 102,655 | 46,429 |
| `png_table_180_rows` | 583,972 | 119,614 |
| `png_rows_60` | 659,934 | 1,526,854 |

## Latest Throughput Results After Table Cache Pass

| Workload | GoPDFKit pdf/s | gopdflib pdf/s | Winner |
|---|---:|---:|---|
| `text_short` | 112,322 | 137,169 | gopdflib |
| `text_240_lines` | 14,021 | 14,043 | gopdflib |
| `table_180_rows` | 12,643 | 11,431 | GoPDFKit |
| `table_900_rows` | 2,352 | 2,339 | GoPDFKit |
| `invoice_40_rows` | 36,543 | 40,724 | gopdflib |
| `png_table_180_rows` | 6,824 | 8,574 | gopdflib |
| `png_rows_60` | 5,049 | 3,684 | GoPDFKit |

## Latest Allocation Signals After Table Cache Pass

| Workload | GoPDFKit B/op | gopdflib B/op |
|---|---:|---:|
| `text_short` | 33,439 | 31,686 |
| `text_240_lines` | 303,570 | 71,978 |
| `table_180_rows` | 373,116 | 97,092 |
| `table_900_rows` | 1,782,266 | 375,701 |
| `invoice_40_rows` | 102,678 | 46,776 |
| `png_table_180_rows` | 584,634 | 119,191 |
| `png_rows_60` | 660,015 | 1,536,428 |

## What This Means

- `gopdflib` is still paying a large penalty on text-heavy and table-heavy workloads.
- The biggest gaps are in multi-line text and table generation, not just image handling.
- Even where throughput is closer, `gopdflib` usually emits larger PDFs and allocates more memory per operation.
- The one current win, `png_table_180_rows`, suggests the recent work helped the mixed image+table path, but the general table/text pipeline is still behind.

## What Changed After The Fixes

- The compare harness was overstating `gopdflib` cost by rebuilding full PDF templates every iteration. Caching those templates made the benchmark fairer.
- Pooling page content stream buffers removed a large multi-page allocation source from `gopdflib`.
- `gopdflib` allocation volume dropped dramatically on most workloads, and throughput is now near-parity on `text_240_lines` and `table_180_rows`.
- The remaining gap is now much more clearly concentrated in the deepest table and image-row cases rather than general template setup.

## What Changed After The Latest Pass

- Structure-tree serialization and repeated standard-font scanning were reduced further.
- `gopdflib` now edges ahead on `text_240_lines`, while still keeping strong wins on `invoice_40_rows` and `png_table_180_rows`.
- `table_180_rows` and `table_900_rows` are still behind, but the bottleneck is now mostly deeper table rendering and compression cost rather than benchmark setup or obvious string formatting overhead.

## What Changed After The Table Cache Pass

- `drawTable` now caches resolved font references per cell, skips custom-font subsetting work for standard-font cells, and reuses precomputed single-line text widths.
- That was enough to flip `text_short` back into a clear `gopdflib` win and turn `text_240_lines` into a small but measurable `gopdflib` lead.
- `table_900_rows` is now close to parity in throughput, but `table_180_rows` still trails and `png_rows_60` remains the largest practical loss.

## Highest-Priority Gap Areas

### 1. Table path

- `table_180_rows`: `10,088` vs `6,988`
- `table_900_rows`: `2,207` vs `1,531`
- This remains the clearest optimization target.

### 2. Multi-line text path

- `text_240_lines`: `12,069` vs `6,884`
- Wrapping, text width measurement, and content-stream construction are still too expensive.

### 3. Output size and allocation pressure

- On every non-winning workload, `gopdflib` generates materially larger PDFs.
- That directly increases compression cost and memory traffic.

## Actionable Checklist

### Profile-backed findings from `gopdfkit_compare`

- [x] `drawTable` is still the dominant library CPU hotspot.
- [x] `WrapTextInto` and text-width measurement remain expensive on multi-line workloads.
- [x] `PageManager.AddNewPage` and page content buffer growth are still major allocation sources in multi-page workloads.
- [x] The compare harness currently rebuilds `gopdflib` templates every iteration via `gopdfSuitTable` / `gopdfSuitTextTable`, while `GoPDFKit` reuses prebuilt workload rows and text. That is a benchmark-side overhead unique to the `gopdflib` path and should be removed for a fairer renderer comparison.

### Immediate

- [x] Pool page content stream buffers across PDF generations so multi-page workloads stop reallocating large `bytes.Buffer` instances in `AddNewPage`.
- [x] Cache prebuilt `gopdflib` benchmark templates per workload inside `sampledata/benchmarks/gopdfkit_compare` so the comparison does not charge `gopdflib` for avoidable per-iteration template construction.
- [ ] Remove hot `fmt.Sprintf` calls from structure-tree serialization in `internal/pdf/generator.go`.
- [x] Remove hot `fmt.Sprintf` calls from structure-tree serialization in `internal/pdf/generator.go`.
- [ ] Remove hot `fmt.Sprintf` calls from page-number, footer, and annotation serialization in `internal/pdf/draw.go` and `internal/pdf/pagemanager.go`.
- [x] Re-profile after the page-buffer and benchmark-template fixes using the same `gopdfkit_compare` harness.

### Next

- [x] Audit `drawTable` for repeated text-width work across identical props and repeated font lookups.
- [ ] Reduce content-stream size inflation so generated `pdf_bytes` gets closer to `GoPDFKit`.
- [ ] Pre-grow builders and buffers in structure/font serialization using known object counts.
- [ ] Re-check whether PDF/A + tagged output is being exercised in workloads where the comparison target is generating simpler output.
- [ ] Run a fresh focused pprof on `table_180_rows`, `table_900_rows`, and `png_rows_60` after the table cache pass to isolate the remaining large-table and image-row gap.

### Validation

- [x] Rerun `BenchmarkGoPDF(Kit|Lib)` on `go1.26.4` after each optimization batch.
- [ ] Track both `pdf/s` and `pdf_bytes`; faster but much larger output is not a clean win.
- [ ] Convert near-parity wins on `text_240_lines` and `table_180_rows` into clear wins.
- [ ] Convert the `text_240_lines` edge into a stable margin and close the remaining `table_180_rows` gap.
- [ ] Turn `table_900_rows` parity into a repeatable win without regressing `invoice_40_rows` or `png_table_180_rows`.
- [ ] Narrow the remaining `png_rows_60` gap by reducing image-row allocation pressure and output size.
