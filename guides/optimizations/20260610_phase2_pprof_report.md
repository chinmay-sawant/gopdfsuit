# gopdfsuit Phase 2 pprof Report

**Date:** 2026-06-10  
**Go Version:** `go1.26.4`  
**Scope:** Post-Phase-2 profile on top of `master` plus the Phase 1 and Phase 2 optimizations already applied

## Commands Used

```bash
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test ./internal/pdf -run '^$' -bench 'BenchmarkGoPdfSuit$' -benchmem -benchtime=5s
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test ./internal/pdf -run '^$' -bench 'BenchmarkGenerateTemplatePDF/Rows2000$' -benchmem -benchtime=5s
GOCACHE=/tmp/go1264-bench /home/chinmay/go/bin/go1.26.4 test ./internal/pdf -run '^$' -bench 'BenchmarkGoPdfSuit$' -benchtime=3s -cpuprofile /tmp/gopdfsuit_phase2_post.cpu -memprofile /tmp/gopdfsuit_phase2_post.mem
```

## Benchmark Progress

### Original `master` baseline

- `BenchmarkGoPdfSuit`: `26047411 ns/op`, `25626521 B/op`, `163386 allocs/op`
- `BenchmarkGenerateTemplatePDF/Rows2000`: `28903973 ns/op`, `24686033 B/op`, `163312 allocs/op`

### After Phase 1

- `BenchmarkGoPdfSuit`: `23531821 ns/op`, `22539669 B/op`, `163307 allocs/op`
- `BenchmarkGenerateTemplatePDF/Rows2000`: `24053195 ns/op`, `22563899 B/op`, `163347 allocs/op`

### After Phase 2

- `BenchmarkGoPdfSuit`: `23025278 ns/op`, `21517345 B/op`, `157268 allocs/op`
- `BenchmarkGenerateTemplatePDF/Rows2000`: `23838900 ns/op`, `22465256 B/op`, `157287 allocs/op`

### Net improvement versus original `master`

- `BenchmarkGoPdfSuit`: about `-11.6% ns/op`, `-16.0% B/op`, `-3.7% allocs/op`
- `BenchmarkGenerateTemplatePDF/Rows2000`: about `-17.5% ns/op`, `-9.0% B/op`, `-3.7% allocs/op`

## Phase 2 Changes Implemented

- Replaced `StructureManager` page-index maps with slices for `NextMCID` and `ParentTree`.
- Added bounded page-slot growth to structure tracking instead of per-cell map churn.
- Added `BeginStructureElementCap` and used it for table and row structure nodes where child counts are known.
- Preallocated leaf structure nodes with a single `Kids` slot for MCID storage.
- Stored annotation object ids directly on Link structure elements to avoid reverse scans during serialization.
- Pre-grew the XMP metadata builder and a few small metadata builders to reduce repeated `strings.Builder` growth.

## Post-Phase-2 pprof Findings

### CPU hotspots

- `drawTable`: `30.68%` cumulative
- `compress/flate` close/write path: still dominant after content generation
- `fmt.Sprintf`: `10.41%` cumulative
- `font.(*CustomFontRegistry).GetTextWidth`: `6.05%` cumulative
- `BeginMarkedContentBuf`: still visible but reduced to `2.91%` cumulative

### Allocation hotspots

- `bytes.growSlice`: `25.25%`
- `compress/flate.NewWriter`: `20.17%`
- `strings.Builder.WriteString`: `12.18%`
- `slices.Clone`: `9.77%`
- `fmt.Sprintf`: `3.41%`
- `BeginMarkedContentBuf`: `3.34%`
- `BeginStructureElementCap`: `0.80%`

## What Improved

- Structure bookkeeping alloc pressure dropped materially.
  Previous pre-Phase-2 profile: `BeginMarkedContentBuf` was `272.97 MB` alloc space.
  Current profile: `153.16 MB`.
- End-to-end benchmark allocations dropped by roughly `6k` allocs/op on the tracked workloads.
- The current bottleneck has shifted away from final PDF assembly and toward table rendering, compression, and string-heavy tagged-PDF/PDF-A serialization.

## Remaining Actionable Checklist

### High Priority

- [ ] Remove hot-path `fmt.Sprintf` calls from structure-tree serialization in `internal/pdf/generator.go`.
- [ ] Remove hot-path `fmt.Sprintf` calls from footer/page-number/link serialization in `internal/pdf/draw.go` and `internal/pdf/pagemanager.go`.
- [ ] Pre-grow `strings.Builder` instances in structure-tree and font-resource serialization based on known object counts.
- [ ] Reduce page-stream growth in multi-page documents by reusing or pooling page content buffers instead of always allocating a fresh `bytes.Buffer` in `AddNewPage`.

### Medium Priority

- [ ] Re-profile `drawTable` and target repeated `GetTextWidth` / `MarkCharsUsed` overhead for repeated fonts and repeated short strings.
- [ ] Audit `collectUsedStandardFonts` and related font scans so template-wide font discovery does not repeatedly parse identical props unnecessarily.
- [ ] Replace remaining `fmt.Sprintf` calls in PDF/A metadata/output-intent generation where the same object layouts are emitted every document.

### Optional but Potentially High Impact

- [ ] Evaluate whether all current PDF/A + PDF/UA tagging is required for every benchmark scenario. If not, benchmark separate compliance modes explicitly.
- [ ] Evaluate whether compression level or compression strategy can be tuned without violating output-size or compatibility requirements.

## Recommendation

The next implementation pass should stay focused on string-heavy serialization, not broad architectural changes. The profile now says the cleanest remaining wins are:

1. `fmt.Sprintf` removal in structure-tree and page-number/footer paths
2. builder pre-sizing in structure and font serialization
3. page content-buffer reuse for multi-page documents
