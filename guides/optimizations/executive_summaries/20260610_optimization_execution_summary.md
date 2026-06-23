# Optimization Execution Summary - 2026-06-10

## Date & Scope

Profile-driven performance work on **`master` at `484b991`** using **Go 1.26.4**, targeting sustainable throughput on internal PDF benchmarks and the GoPDFKit comparison harness without changing PDF output semantics. Work spanned Phase 1/2 engine optimizations, post-Phase-2 pprof analysis, SlopGuard remediation (311 findings), and iterative gopdflib vs GoPDFKit parity passes.

## Key Outcomes

**Internal micro-benchmarks (`./internal/pdf`, benchtime=5s)**

| Benchmark | Baseline | Post Phase 1 | Post Phase 2 (net vs baseline) |
|-----------|----------|--------------|--------------------------------|
| `BenchmarkGoPdfSuit` | 26,047,411 ns/op, 25,626,521 B/op | 23,531,821 ns/op (−9.7%), 22,539,669 B/op (−12.0%) | 23,025,278 ns/op (**−11.6%**), 21,517,345 B/op (**−16.0%**), 157,268 allocs (−3.7%) |
| `BenchmarkGenerateTemplatePDF/Rows2000` | 28,903,973 ns/op, 24,686,033 B/op | 24,053,195 ns/op (−16.8%), 22,563,899 B/op (−8.6%) | 23,838,900 ns/op (**−17.5%**), 22,465,256 B/op (−9.0%), 157,287 allocs (−3.7%) |

**SlopGuard fixes (10-run avg vs prior buffer-clone pass)**

| Workload | Δ throughput |
|----------|-------------|
| `table_180_rows` | **+13.0%** (13,051 vs 11,548 pdf/s) |
| `text_240_lines` | **+9.0%** |
| `table_900_rows` | **+4.6%** |
| `invoice_40_rows` | −1.0% (noise) |
| `text_short` | −6.6% (handler overhead on fastest path) |

**GoPDFKit comparison (end of day)**

- Evolved from gopdflib **5/7** to near-parity on tables after image dedup, borrowed-buffer API, and border fast path
- `table_180_rows`: **22,297** pdf/s; `table_900_rows`: **4,667** pdf/s; `png_rows_60`: **31,569** pdf/s
- Allocation volume dropped **20–40%** on most workloads; `slices.Clone` fell from 28–43% of table alloc volume to below top-15

## Work Completed

**Phase 1 - Buffer & compression**
- Increased pooled final PDF buffer capacity
- Removed extra scratch-slice hop at final assembly
- Streamed page content directly into zlib (no `contentStream.Bytes()`)
- Replaced `append([]byte(nil), …)` with explicit `slices.Clone`
- Pre-sized initial page content streams from template complexity

**Phase 2 - Structure tree & table pressure**
- `StructureManager` page-index maps → slices
- `BeginStructureElementCap` for known table/row child counts
- Pre-grew XMP metadata builders; stored annotation object IDs on Link elements

**SlopGuard (311 findings, 13 files)**
- PERF-31: 13 `defer` removals in font registry
- PERF-1: 20+ regex compilations hoisted to package level
- PERF-6/15: 30+ `fmt.Sprintf`/`strconv.Itoa` → `strconv.AppendInt`
- PERF-4: map pre-sizing; PERF-42: `errors.New`; PERF-41/43: `gin.CustomRecovery`

**GoPDFKit parity passes**
- Pooled page content stream buffers; per-cell font-ref cache; uniform-border fast path
- `DecodeImageData` MRU fast path; `crypto/md5` → `crypto/sha256` for document-ID hash
- Image XObject dedup; `GeneratePDFBorrowed` / `Release()` borrowed-buffer API

## Findings / Bottlenecks

**Pre-Phase-2 profile**
- CPU alloc leaders: `bytes.growSlice` **20.91%**, `compress/flate.NewWriter` **20.49%**, `slices.Clone` **7.14%**
- CPU time: `GenerateTemplatePDF` dominates; `drawTable` largest content hotspot

**Post-Phase-2 profile**
- Bottleneck shifted from final PDF assembly → table rendering, compression, string-heavy tagged-PDF serialization
- `drawTable` **30.68%** cumulative CPU; `fmt.Sprintf` **10.41%**; `BeginMarkedContentBuf` reduced to **2.91%**

## Open Items / Next Steps

- Remove hot-path `fmt.Sprintf` in structure-tree serialization and footer/page-number paths
- Pre-grow `strings.Builder` in structure-tree and font-resource serialization
- Pool/reuse page content buffers in `AddNewPage`
- Re-profile `compress/flate.NewWriter` (Phase 3); tune pooled writer/buffer reuse if still material
- GoPDFKit still led on table workloads in last full side-by-side compare (margins narrowing)

## Source Documents

| File | Role |
|------|------|
| `20260610_master_performance_execution_plan.md` | Master plan, Phase 1/2 checklist, benchmark results |
| `20260610_phase2_pprof_report.md` | Post-Phase-2 pprof findings and remaining checklist |
| `20260610_slopguard_fixes_report.md` | SlopGuard remediation validation |
| `20260610_gopdfkit_compare_results.md` | Iterative gopdflib vs GoPDFKit comparison |
| `PR/PR_DESCRIPTION.md` | Cross-phase PR summary (branch scope extends beyond this date) |