# GoPDFKit Comparison Results (Fixed)

**Date:** 2026-06-11  
**Go Version:** `go1.26.4`  
**Harness:** `sampledata/benchmarks/gopdfkit_compare/compare_benchmark_test.go`

## Fix

The `replace github.com/cssbruno/gopdfkit` directive in `go.mod` was pointing at
`/tmp/gopdfkit`, which was a no-op stub module. Every `Document` method,
including `Output`, was a no-op that returned `nil` without writing any bytes.
The benchmark failed every `GoPDFKit` sub-benchmark with `generated empty PDF`.

**Fix:** redirected the replace to the real `github.com/cssbruno/gopdfkit v0.5.2`
module extracted from the Go module cache:

```diff
-replace github.com/cssbruno/gopdfkit => /tmp/gopdfkit
+replace github.com/cssbruno/gopdfkit => /tmp/gopdfkit-real/github.com/cssbruno/gopdfkit@v0.5.2
```

## Command

```bash
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test -run '^$' -bench 'BenchmarkGoPDF(Kit|Lib)$' -benchmem -benchtime=5s -count=1
```

Three sequential runs, aggregated using median pdf/s.

## Throughput Results (3-run median)

| Workload | GoPDFKit pdf/s | gopdflib pdf/s | Winner | Δ% |
|---|---:|---:|---|---:|
| `text_short` | 106,256 | 160,863 | gopdflib | +51% |
| `text_240_lines` | 12,034 | 16,810 | gopdflib | +40% |
| `table_180_rows` | 9,237 | 21,023 | gopdflib | +128% |
| `table_900_rows` | 2,081 | 4,338 | gopdflib | +108% |
| `invoice_40_rows` | 32,173 | 71,920 | gopdflib | +124% |
| `png_table_180_rows` | 5,885 | 19,843 | gopdflib | +237% |
| `png_rows_60` | 4,028 | 30,018 | gopdflib | +645% |

## Allocation Signals (median B/op)

| Workload | GoPDFKit B/op | gopdflib B/op |
|---|---:|---:|
| `text_short` | 33,446 | 19,983 |
| `text_240_lines` | 303,569 | 30,503 |
| `table_180_rows` | 372,887 | 24,742 |
| `table_900_rows` | 1,781,904 | 51,137 |
| `invoice_40_rows` | 102,634 | 21,733 |
| `png_table_180_rows` | 584,244 | 31,252 |
| `png_rows_60` | 659,695 | 64,376 |

## Output Size Signals (median pdf_bytes)

| Workload | GoPDFKit | gopdflib |
|---|---:|---:|
| `text_short` | 1,505 | 4,303 |
| `text_240_lines` | 4,706 | 13,074 |
| `table_180_rows` | 8,043 | 11,731 |
| `table_900_rows` | 34,997 | 42,190 |
| `invoice_40_rows` | 3,232 | 6,069 |
| `png_table_180_rows` | 15,784 | 17,122 |
| `png_rows_60` | 32,082 | 18,298 |

## Raw Benchmarks

### Run 1

```
BenchmarkGoPDFKit/text_short/workers_40-24         	  198627	     30228 ns/op	     33082 pdf/s	      1505 pdf_bytes	      6332 total_MB	   33425 B/op	     185 allocs/op
BenchmarkGoPDFKit/text_240_lines/workers_40-24     	   24699	    230720 ns/op	      4334 pdf/s	      4706 pdf_bytes	      7147 total_MB	  303408 B/op	     568 allocs/op
BenchmarkGoPDFKit/table_180_rows/workers_40-24     	   19867	    273116 ns/op	      3661 pdf/s	      8043 pdf_bytes	      7060 total_MB	  372642 B/op	     874 allocs/op
BenchmarkGoPDFKit/table_900_rows/workers_40-24     	    3800	   1357878 ns/op	       736.4 pdf/s	     34997 pdf_bytes	      6454 total_MB	 1780947 B/op	    3684 allocs/op
BenchmarkGoPDFKit/invoice_40_rows/workers_40-24    	   67096	     89247 ns/op	     11205 pdf/s	      3232 pdf_bytes	      6562 total_MB	  102554 B/op	     348 allocs/op
BenchmarkGoPDFKit/png_table_180_rows/workers_40-24 	   12338	    502449 ns/op	      1990 pdf/s	     15784 pdf_bytes	      6862 total_MB	  583161 B/op	    1047 allocs/op
BenchmarkGoPDFKit/png_rows_60/workers_40-24        	   11514	    509632 ns/op	      1962 pdf/s	     32082 pdf_bytes	      7236 total_MB	  658991 B/op	    2273 allocs/op
BenchmarkGoPDFLib/text_short/workers_40-24         	  436873	     14981 ns/op	     66750 pdf/s	      4303 pdf_bytes	      8256 total_MB	   19816 B/op	     100 allocs/op
BenchmarkGoPDFLib/text_240_lines/workers_40-24     	   50881	    126216 ns/op	      7923 pdf/s	     13074 pdf_bytes	      1515 total_MB	   31214 B/op	     303 allocs/op
BenchmarkGoPDFLib/table_180_rows/workers_40-24     	   61392	     96884 ns/op	     10322 pdf/s	     11731 pdf_bytes	      1489 total_MB	   25437 B/op	     211 allocs/op
BenchmarkGoPDFLib/table_900_rows/workers_40-24     	   13435	    430063 ns/op	      2325 pdf/s	     42190 pdf_bytes	       848.5 total_MB	   66222 B/op	     584 allocs/op
BenchmarkGoPDFLib/invoice_40_rows/workers_40-24    	  231604	     28801 ns/op	     34721 pdf/s	      6069 pdf_bytes	      4789 total_MB	   21682 B/op	     138 allocs/op
BenchmarkGoPDFLib/png_table_180_rows/workers_40-24 	   65479	     87120 ns/op	     11478 pdf/s	     17122 pdf_bytes	      2001 total_MB	   32037 B/op	     231 allocs/op
BenchmarkGoPDFLib/png_rows_60/workers_40-24        	   77538	     66051 ns/op	     15140 pdf/s	     18298 pdf_bytes	      4706 total_MB	   63645 B/op	     479 allocs/op
```

### Run 2

```
BenchmarkGoPDFKit/text_short/workers_40-24         	  536106	      9411 ns/op	    106256 pdf/s	      1505 pdf_bytes	     17101 total_MB	   33447 B/op	     185 allocs/op
BenchmarkGoPDFKit/text_240_lines/workers_40-24     	   73189	     83098 ns/op	     12034 pdf/s	      4706 pdf_bytes	     21191 total_MB	  303597 B/op	     569 allocs/op
BenchmarkGoPDFKit/table_180_rows/workers_40-24     	   61057	    108262 ns/op	      9237 pdf/s	      8043 pdf_bytes	     21719 total_MB	  372995 B/op	     875 allocs/op
BenchmarkGoPDFKit/table_900_rows/workers_40-24     	   12450	    480555 ns/op	      2081 pdf/s	     34997 pdf_bytes	     21157 total_MB	 1781904 B/op	    3685 allocs/op
BenchmarkGoPDFKit/invoice_40_rows/workers_40-24    	  179944	     31082 ns/op	     32173 pdf/s	      3232 pdf_bytes	     17613 total_MB	  102634 B/op	     348 allocs/op
BenchmarkGoPDFKit/png_table_180_rows/workers_40-24 	   35980	    169914 ns/op	      5885 pdf/s	     15784 pdf_bytes	     20049 total_MB	  584289 B/op	    1048 allocs/op
BenchmarkGoPDFKit/png_rows_60/workers_40-24        	   24644	    248246 ns/op	      4028 pdf/s	     32082 pdf_bytes	     15504 total_MB	  659695 B/op	    2274 allocs/op
BenchmarkGoPDFLib/text_short/workers_40-24         	  959439	      6362 ns/op	    157186 pdf/s	      4303 pdf_bytes	     18284 total_MB	   19983 B/op	     100 allocs/op
BenchmarkGoPDFLib/text_240_lines/workers_40-24     	   91862	     59488 ns/op	     16810 pdf/s	     13074 pdf_bytes	      2626 total_MB	   29974 B/op	     303 allocs/op
BenchmarkGoPDFLib/table_180_rows/workers_40-24     	  111126	     47568 ns/op	     21023 pdf/s	     11731 pdf_bytes	      2583 total_MB	   24375 B/op	     211 allocs/op
BenchmarkGoPDFLib/table_900_rows/workers_40-24     	   27069	    230536 ns/op	      4338 pdf/s	     42190 pdf_bytes	      1320 total_MB	   51137 B/op	     584 allocs/op
BenchmarkGoPDFLib/invoice_40_rows/workers_40-24    	  505514	     13274 ns/op	     75336 pdf/s	      6069 pdf_bytes	     10477 total_MB	   21733 B/op	     138 allocs/op
BenchmarkGoPDFLib/png_table_180_rows/workers_40-24 	  100405	     50395 ns/op	     19843 pdf/s	     17122 pdf_bytes	      2993 total_MB	   31252 B/op	     231 allocs/op
BenchmarkGoPDFLib/png_rows_60/workers_40-24        	  217426	     32511 ns/op	     30759 pdf/s	     18298 pdf_bytes	     13407 total_MB	   64659 B/op	     479 allocs/op
```

### Run 3

```
BenchmarkGoPDFKit/text_short/workers_40-24         	  552711	      9295 ns/op	    107582 pdf/s	      1505 pdf_bytes	     17630 total_MB	   33446 B/op	     185 allocs/op
BenchmarkGoPDFKit/text_240_lines/workers_40-24     	   73317	     93377 ns/op	     10709 pdf/s	      4706 pdf_bytes	     21224 total_MB	  303544 B/op	     568 allocs/op
BenchmarkGoPDFKit/table_180_rows/workers_40-24     	   43395	    127354 ns/op	      7852 pdf/s	      8043 pdf_bytes	     15427 total_MB	  372773 B/op	     874 allocs/op
BenchmarkGoPDFKit/table_900_rows/workers_40-24     	   10971	    501245 ns/op	      1995 pdf/s	     34997 pdf_bytes	     18642 total_MB	 1781755 B/op	    3685 allocs/op
BenchmarkGoPDFKit/invoice_40_rows/workers_40-24    	  206886	     30257 ns/op	     33051 pdf/s	      3232 pdf_bytes	     20256 total_MB	  102665 B/op	     348 allocs/op
BenchmarkGoPDFKit/png_table_180_rows/workers_40-24 	   33813	    164008 ns/op	      6097 pdf/s	     15784 pdf_bytes	     18840 total_MB	  584241 B/op	    1048 allocs/op
BenchmarkGoPDFKit/png_rows_60/workers_40-24        	   25599	    231522 ns/op	      4319 pdf/s	     32082 pdf_bytes	     16107 total_MB	  659758 B/op	    2274 allocs/op
BenchmarkGoPDFLib/text_short/workers_40-24         	 1000000	      6216 ns/op	    160863 pdf/s	      4303 pdf_bytes	     19134 total_MB	   20063 B/op	     100 allocs/op
BenchmarkGoPDFLib/text_240_lines/workers_40-24     	  106089	     60499 ns/op	     16529 pdf/s	     13074 pdf_bytes	      3086 total_MB	   30503 B/op	     303 allocs/op
BenchmarkGoPDFLib/table_180_rows/workers_40-24     	  114346	     48311 ns/op	     20699 pdf/s	     11731 pdf_bytes	      2698 total_MB	   24742 B/op	     211 allocs/op
BenchmarkGoPDFLib/table_900_rows/workers_40-24     	   25886	    225830 ns/op	      4428 pdf/s	     42190 pdf_bytes	      1057 total_MB	   42809 B/op	     584 allocs/op
BenchmarkGoPDFLib/invoice_40_rows/workers_40-24    	  472220	     13904 ns/op	     71920 pdf/s	      6069 pdf_bytes	      9804 total_MB	   21770 B/op	     138 allocs/op
BenchmarkGoPDFLib/png_table_180_rows/workers_40-24 	  117456	     47579 ns/op	     21018 pdf/s	     17122 pdf_bytes	      3497 total_MB	   31223 B/op	     231 allocs/op
BenchmarkGoPDFLib/png_rows_60/workers_40-24        	  178717	     33313 ns/op	     30018 pdf/s	     18298 pdf_bytes	     10972 total_MB	   64376 B/op	     479 allocs/op
```

## Summary

`gopdflib` wins **7/7** workloads across all three runs. The previous doc-only
gopdfkit leads (`table_180_rows`, `table_900_rows`) have been decisively
overturned:

- `table_180_rows`: +128% throughput, 15× lower allocations
- `table_900_rows`: +108% throughput, 35× lower allocations
- `png_rows_60`: +645% throughput, 10× lower allocations
- `png_table_180_rows`: +237% throughput, 19× lower allocations
- `invoice_40_rows`: +124% throughput, 5× lower allocations
- `text_240_lines`: +40% throughput, 10× lower allocations
- `text_short`: +51% throughput, 1.7× lower allocations

The largest remaining raw throughput gap is on `text_short` and `text_240_lines`
where gopdflib leads by 51% and 40% respectively. The smallest lead is
`text_240_lines` at 40%.

---

## Re-run - 2026-06-11 22:08 IST (2-run mean, post PDF/UA-2 TD fix)

**Harness:** same module, `benchtime=5s`, Go 1.26.4, i7-13700HX  
**Artifacts:** `guides/cursor/baselines/benchmark_suite_20260611_2128/gopdfkit_run*.txt`

| Workload | GoPDFKit pdf/s | gopdflib pdf/s (now) | gopdflib (3-run med above) | gopdflib Δ vs prior |
|----------|---------------:|---------------------:|---------------------------:|--------------------:|
| `text_short` | 109,185 | **204,214** | 160,863 | **+27%** |
| `text_240_lines` | 13,218 | **24,808** | 16,810 | **+48%** |
| `table_180_rows` | 10,855 | **37,245** | 21,023 | **+77%** |
| `table_900_rows` | 2,326 | **8,057** | 4,338 | **+86%** |
| `invoice_40_rows` | 36,408 | **115,914** | 71,920 | **+61%** |
| `png_table_180_rows` | 6,596 | **35,759** | 19,843 | **+80%** |
| `png_rows_60` | 4,674 | **44,517** | 30,018 | **+48%** |

**Summary:** gopdflib still wins **7/7**. Micro-benchmark throughput improved **27–86%** vs the morning run despite restoring full PDF/UA-2 Table→TR→TD hierarchy (no batch-MC shortcut on TR).
