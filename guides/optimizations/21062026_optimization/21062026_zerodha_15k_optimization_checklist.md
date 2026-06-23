# Zerodha gopdflib x10 pprof Optimization Checklist - 9,000 → 15,000 ops/sec

**Date:** 2026-06-22  
**Workload:** `sampledata/gopdflib/zerodha`, 80% retail / 15% active / 5% HFT  
**Primary command:** `make bench-gopdflib-zerodha-x10`  
**Profile command:** `make bench-gopdflib-zerodha-x5`  
**Combined target:** `make bench-gopdflib-zerodha-x10-pprof`  
**Go version:** `go1.26.4`  
**Concurrency:** 48 workers, `GOMAXPROCS=24`  
**Branch:** `feat/optimization-5.5-medium`  
**Prior checklist:** `guides/optimizations/20260620_zerodha_x10_pprof_optimization_checklist.md` (P0–P25, 8K target met)
**Commit profile sweep:** `87009f13418b23e4ae88d4ab044204b04f9765a4` (46 `.prof` files across `gin`, `zerodha`, and `pypdfsuit` baselines)

---

## Goal

Reach **x10 mean ≥ 15,000 ops/sec** on `make bench-gopdflib-zerodha-x10` while keeping
full compliance on all three templates:

| Template | Mix | Output size (bytes) | Requirements |
|----------|----:|--------------------:|--------------|
| Retail | 80% | 61,301 | PDF/A-4, PDF/UA-2, ECDSA P-256 signing |
| Active | 15% | 76,050 | PDF/A-4, PDF/UA-2, no signing |
| HFT | 5% | 2,291,942 | PDF/A-4, PDF/UA-2, compliant **TR → TD** hierarchy |

**Gap from current baseline:** +5,991 ops/sec (+66%) from mean **9,009** (idle best session).

---

## Four-Subagent Refresh (2026-06-21)

Four focused subagents reviewed `cpu_zerodha_run2.prof`, `heap_zerodha.prof`, the
cross-validation report, and this checklist. The validation report remains the source of
compliance guardrails and broad phase shape; `run2` is the source for refreshed ordering,
gates, and wording.

### Agent roles and refreshed findings

| Agent | Focus | `run2` takeaway | Checklist impact |
|-------|-------|-----------------|------------------|
| **A1 - HFT struct path** | `drawTable`, `drawSharedLayoutRow`, TR→TD setup, struct emit | HFT is still the mandatory tail: `drawTable` 29.6% cum, `drawSharedLayoutRow` 13.0% cum, `BeginTableRowWithTDMCIDs`/`beginTableRowArena` 6.7% cum, `formatStructElemObjectTo` 5.5% cum | Keep Phase C, but broaden `P36/P38/P39` to cover the whole hot path |
| **A2 - Memory / buffers / GC** | `bytes.growSlice`, `memmove`, `memclr`, arena slabs, page streams | The strongest direct costs are still memory pressure: `runtime.memmove` 12.0% flat, `runtime.memclrNoHeapPointers` 11.4% flat, `bytes.growSlice` 263 MB, arena slabs 246.5 MB, page-stream buffers 156.9 MB cum | Strengthen Phase B, fold `P27a` into `P31`, move `P34` ahead of `P33`, add explicit copy/variance gates |
| **A3 - Cross-format / signing / metadata** | sRGB ICC, signing, font walks, active table | `P26` remains the cleanest first win; signature work is real but the benchmark already signs in place; `P28` only removes the startup font walk; active `SharedRowLayout` is valid but lower leverage than the current wording implies | Keep `P26` first, narrow `P28/P35`, demote `P29` |
| **A4 - Synthesis / ranking** | Phase ordering and confidence | The report’s A→B→C structure still holds, but `run2` supports banded ranking more than precise gain ladders | Treat HFT weighting as directional, not exact; promote `P40` into Phase B; rank `P38` ahead of `P37` |

### Resolved priorities

| Topic | `run2` read | Checklist change |
|-------|-------------|------------------|
| First quick win | `GetSRGBICCProfile` / `buildSRGBICCProfile` still costs ~4.2% cum | Keep `P26` first |
| Memory-first bias | `memmove` + `memclr` dominate flat CPU; heap is still buffer + arena heavy | Strengthen Phase B before deep HFT rewrites |
| HFT claims | HFT remains the remaining gate, but `run2` does not justify exact weighted-latency math by itself | Keep the qualitative conclusion; soften exact `75–85%` style wording |
| Active priority | Active table is still eligible for `SharedRowLayout`, but no active-specific hotspot breaks into `run2` | Keep `P29`, but treat it as late Phase A / opportunistic |

### Commit-wide `.prof` sweep (87009f1)

The commit contains three profile families, and they do not all have the same decision weight:

| Family | Profiles read | What it validates | Checklist use |
|--------|---------------|-------------------|---------------|
| `zerodha_pprof_runs` | 5 CPU + 1 heap | Stable x10 mixed-workload ordering: `memmove`/`memclr`, TR→TD setup, struct emit, `bytes.growSlice`, arena slabs | **Primary source** for the 15K backlog |
| `gin_pprof_runs` | 19 CPU + 19 heap | k6-facing guardrails evolved from early shared-row/prealloc blowups into later signing, ICC, compression, and buffer-write costs | Guardrail context for `make bench-k6-light` / `make bench-k6` |
| `pypdfsuit_pprof_runs` | 1 CPU + 1 heap | Same broad pressure points reappear under a separate render harness: `memclr`, `memmove`, `bytes.growSlice`, struct emit, signing | Secondary confirmation only |

What changed from reading all 46 `.prof` files:

- The Zerodha x10 family still owns the checklist ordering.
- The gin family is the reason the k6 guardrail should watch signing and ICC/profile work, not only `drawSharedLayoutRow` heap.
- The pypdfsuit family backs the same memory-first story rather than introducing a different next target.

---

## Hard Guardrails (Non-Negotiable)

- [ ] **No compliance shortcuts on HFT.** Keep `TR → TD` with one `TD` per column, each
      carrying its own MCID. Do not collapse TDs into bare MCID leaves on `TR`.
- [ ] **No key-based cross-request caches.** No `sync.Map` keyed by row/content/output.
      The bounded `sharedRowRenderCache` may stay as-is but must not expand.
- [ ] **No disabling compliance flags.** `PDFACompliant`, `TaggedPDF`, `ArlingtonCompatible`,
      `EmbedFonts`, retail `Signature` (ECDSA P-256) stay on.
- [ ] **HFT output size stable.** Target **2,291,942 bytes ± 5%** (current compliant size).
- [ ] **veraPDF gate after every phase.** `make test-verify-pdfs` - 6/6 PASS
      (retail/active/HFT × PDF/A-4 + PDF/UA-2).
- [ ] **k6 must not regress.** `make bench-k6-light` / `make bench-k6` should track
      `drawSharedLayoutRow`, signing (`embedSignatureInPlace` / `CreateSignatureField`),
      and ICC/profile work in addition to heap.

---

## Fresh Measurement (2026-06-21 `make bench-gopdflib-zerodha-x10-pprof`)

Build cache reset: `GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache`

### x10 timing (this session - machine under load)

| Metric | Value | Notes |
|--------|------:|-------|
| Best throughput | 9,182.09 ops/sec | Run 9 |
| Worst throughput | 5,572.83 ops/sec | Run 1 (cold pool) |
| Mean throughput | **7,852.41 ops/sec** | Load-depressed vs idle 9,009 |
| Median throughput | 8,162.30 ops/sec | |
| Stddev throughput | 1,112.25 ops/sec | Gate target ≤ 400 |
| Mean avg latency | 6.066 ms | |
| Mean peak allocated | 1,119.42 MB | Gate target ≤ 650 MB |

**Artifacts:**
- `guides/cursor/baselines/zerodha_bench_x10_wsl/zerodha_run{1..10}.txt`
- `guides/cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt`
- `guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run{1..5}.prof`
- `guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof`

### Reference baseline (idle machine, prior session)

| Metric | Value |
|--------|------:|
| Mean throughput | **9,009.02 ops/sec** |
| Best throughput | 10,658.69 ops/sec |
| Stddev | 1,332.69 ops/sec |
| Mean peak allocated | 1,077.42 MB |

**Use idle 9,009 as the regression baseline; do not compare load-depressed 7,852 runs.**

### Fresh CPU profile (`cpu_zerodha_run2.prof`, 11.13s samples, 1871.69% CPU)

| Hotspot | Cum % | Flat % | Format scope | Status |
|---------|------:|-------:|--------------|--------|
| `GenerateTemplatePDFBorrowed` | 91.6% | 4.0% | All | Top gate |
| `generateAllContentWithImages` | 32.0% | - | All | Content emit |
| `drawTable` | 29.6% | 1.3% | HFT-heavy | Table render |
| `drawSharedLayoutRow` | 13.0% | 0.1% | HFT-heavy | Shared-row replay path |
| `runtime.memmove` | - | **12.0%** | All | **Open gate** |
| `runtime.memclrNoHeapPointers` | - | **11.4%** | All (GC-driven) | **Open gate** |
| `BeginTableRowWithTDMCIDs` / `beginTableRowArena` | 6.7% | 4.3% | HFT-heavy | TR→TD setup |
| `signature.UpdatePDFWithSignatureBuffer` | 6.7% | - | Retail 80% | Signing still material |
| `formatStructElemObjectTo` | 5.5% | 4.3% | Tagged (HFT-heavy) | Struct emit |
| `GetSRGBICCProfile` / `buildSRGBICCProfile` | 4.2% | 0.2% | All 3 | **P26 still first** |
| `MarkCharsUsed` + `markSharedTableCharsUsed` | 4.9% | 0.8% | HFT-heavy | Font subset / prescan |
| `appendStructElemTDLeaf` | 3.7% | 1.7% | HFT-heavy | TD leaf emit |
| `estimateFinalPDFSize` | - | 1.1% | All | Better folded into buffer sizing |

### Fresh heap profile (`heap_zerodha.prof`, 563 MB in-use)

| Hotspot | In-use | % | Status |
|---------|-------:|--:|--------|
| `bytes.growSlice` | 263 MB | 46.7% | **Top gate** |
| Arena slabs (`acquireArenaSlabForCapacity`) | 247 MB | 43.8% | **Top gate** |
| `getPageContentStreamBuffer` (cum) | 157 MB | 27.9% cum | Page streams |
| `drawTable` (cum) | 166 MB | 29.6% cum | Content + struct |

### Delta vs 15,000 target

| Metric | Idle baseline | Target | Gap |
|--------|-------------:|-------:|----:|
| x10 mean throughput | 9,009 ops/sec | ≥ 15,000 | **−5,991 (−40%)** |
| x10 stddev | 1,333 ops/sec | ≤ 400 | −933 |
| Mean peak allocated | 1,077 MB | ≤ 650 MB | −427 MB |
| HFT output size | 2,291,942 bytes | stable ±5% | - |

---

## Phased Execution Plan

```
Phase A (Quick wins)     → ~10,400–10,900 ops/sec
Phase B (Memory wall)    → ~13,000 ops/sec  (+44%)
Phase C (HFT tail)       → ~15,000 ops/sec  (+66%)
```

---

## Phase A - Cross-Format Quick Wins → ~10,400–10,900 ops/sec

**Duration:** 1–2 days  
**Projected gain:** +800–1,300 ops/sec from idle 9,009 baseline  
**Risk:** Low

### P0 - Harness and measurement hygiene

- [ ] Always run with `GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache`.
- [ ] Run x10 on an **idle machine** (no parallel benchmarks, Docker, browser).
- [ ] Use x10 **mean** as regression gate; capture `zerodha_bench_x10_wsl_stats_latest.txt` after every accepted change.
- [ ] Re-profile after Phase A: `make bench-gopdflib-zerodha-x10-pprof`.
- [ ] **Gate:** track warmed-run variance separately from the cold first run; do not use
      run-1 outliers as the accept/reject signal for Phase A.

### P26 - Fix sRGB ICC cache leak (all 3 templates) ★ Highest ROI

**Basis:** validated by `cpu_zerodha_run2.prof`, the full `cpu_zerodha_run{1..5}.prof` sweep,
and the commit-wide profile review.

**Root cause:** `GetSRGBICCProfile()` calls `buildSRGBICCProfile()` (1024× `math.Pow`)
on every PDF. `GenerateOutputIntent` line 306 uses `getSRGBICCProfile()` for `Grow()`
sizing even though `compressedSRGBICCProfileCache()` is used at emit time.

**Fix:**
- [x] Cache uncompressed sRGB bytes at `init` alongside `srgbICCProfileCompressed`.
- [x] Change `GetSRGBICCProfile()` to return cached bytes.
- [x] Rework `GenerateOutputIntent` to reuse the cached compressed payload instead of
      re-reading / recompressing raw sRGB bytes per document.

**Files:** `internal/pdf/pdfa.go`, `internal/pdf/metadata.go`  
**Estimated gain:** +350–550 ops/sec (all 5,000 iterations)  
**Compliance risk:** Low - byte-identical ICC profile  
**Gate:** `buildSRGBICCProfile` not in top-40 CPU; veraPDF 6/6 PASS

**2026-06-22 implementation note:** landed in code; added focused cache-regression tests in
`internal/pdf/pdfa_test.go` and `internal/pdf/metadata_test.go`. Throughput / veraPDF gate
still pending.

### P28 - Precompute standard fonts at template build

**Agents:** A3 - `collectAllStandardFontsInTemplate` 2.88% cum at generator startup.

- [x] Build the PDF/A startup font set once when `buildRetailTemplate` /
      `buildActiveTraderTemplate` / `buildHFTTemplate` run.
- [x] Store on template metadata; skip the pre-generation standard-font walk in
      `GenerateTemplatePDFBorrowed`.

**Files:** `internal/models/models.go`, `internal/pdf/generator.go`,
`sampledata/gopdflib/zerodha/bench.go`  
**Estimated gain:** +150–250 ops/sec  
**Compliance risk:** Low

**2026-06-22 implementation note:** current Zerodha templates are Helvetica-only, so the
phase-1 implementation stores an explicit precomputed startup hint rather than rescanning
the template tree at generation start.

### P30 - Static XMP metadata shell (all 3 templates)

**Agents:** A2 - `GenerateXMPMetadata` 1.0% cum; dates/ID are only per-PDF variance.

- [x] Pre-build XMP packet templates with placeholder slots for date/ID.
- [x] Patch only variable fields at emit time.

**Files:** `internal/pdf/metadata.go`  
**Estimated gain:** +80–120 ops/sec  
**Compliance risk:** Low - PDF/A-4 + PDF/UA-2 metadata fields unchanged

**2026-06-22 implementation note:** handler-local XMP fragments now cache the static packet
shell; emit-time work is limited to timestamps, document ID, and encryption wrapping.

### P29 - Active trader `SharedRowLayout` enablement

**Agents:** A3 primary; validated as compliance-safe, but lower leverage than the prior
writeup implied.

**Root cause:** 41-row trade table has uniform `Props` per column; only `Text`/`TextColor`
vary. Passes `tableSupportsSharedRowLayout` but the flag is still not set.

- [x] Add `SharedRowLayout: true, SharedRowTemplateRow: 1` to the active trade table in
      `buildActiveTraderTemplate()`.
- [ ] Verify alternating `BgColor` and per-row `TextColor` still render correctly.
- [ ] Confirm veraPDF active output PASS.

**Files:** `sampledata/gopdflib/zerodha/bench.go`, `internal/pdf/draw.go`  
**Estimated gain:** +120–220 ops/sec (safe stacking win, not a near-term throughput lever)  
**Compliance risk:** Low - TR→TD hierarchy preserved via shared layout path  
**Gate:** `zerodha_active_output.pdf` 76,050 ± 5% bytes; veraPDF 2/2 PASS

**2026-06-22 implementation note:** code path is enabled; visual/compliance validation remains
open.

### Phase A acceptance gate

- [ ] x10 mean ≥ **10,400 ops/sec** (idle machine)
- [ ] veraPDF 6/6 PASS
- [ ] HFT output 2,291,942 ± 5% bytes
- [ ] `buildSRGBICCProfile` cum CPU ≤ 0.5%

---

## Phase B - Memory Wall → ~13,000 ops/sec

**Duration:** 1 week  
**Projected gain:** +2,200–2,800 ops/sec cumulative  
**Risk:** Low–Medium  
**Basis:** `cpu_zerodha_run{1..5}.prof`, `heap_zerodha.prof`, plus commit-wide gin /
pypdfsuit confirmation that copy pressure and signing remain hot after the early HFT fixes

### P31 - pdfBuffer zero-grow + final-size instrumentation (all 3 templates, includes former `P27a`)

**Root cause:** `bytes.growSlice` = 263 MB in-use (47%); `runtime.memmove` stays in the
8–12% flat range across `cpu_zerodha_run1..5`. The current 2.5 MiB HFT cap still grows,
and `estimateFinalPDFSize` is better treated as a sizing input than a standalone Phase A CPU win.

- [ ] Instrument final PDF buffer high-water by template class in the benchmark harness.
- [ ] Rework `estimateFinalPDFSize` so it overestimates compliant HFT output, including
      ParentTree/object refs, page streams, and signature slack.
- [ ] Set retail/active/HFT caps from measured max + safety margin; do not shrink the
      current HFT cap until warmed runs show zero growth.
- [ ] Keep `pdfBufferPoolMedium` optional; only add a middle tier if active PDFs still
      cross pool boundaries after cap fixes.
- [ ] Add a `BENCH_DEBUG_CAPS=1` style assert: zero `Grow` calls after the initial cap on warmed runs.

**Files:** `internal/pdf/generator.go`  
**Estimated gain:** +900–1,200 ops/sec + −200 MB heap  
**Compliance risk:** Low - sizing only  
**Gate:** `bytes.growSlice` in-use ≤ 50 MB; `runtime.memmove` flat ≤ 6%

### P34 - Page content stream exact caps (HFT + active)

**Root cause:** `getPageContentStreamBuffer` is still 157 MB cumulative live pressure in
`heap_zerodha.prof`, which is materially larger than the xref-offset footprint. Shared-row
replay still pays for page-stream growth when the per-page cap is too loose.

- [ ] Profile max page stream length per template class and per HFT stripe.
- [ ] Pre-cap page streams to measured max + small margin; grow once per stripe, not per row.
- [ ] Keep the pool budget bounded by the heap target; do not pre-retain a 48 × 50 × 128 KiB steady-state reserve.

**Files:** `internal/pdf/pagemanager.go`, `internal/pdf/draw.go`  
**Estimated gain:** +350–550 ops/sec + −60–90 MB heap  
**Compliance risk:** Medium

### P40 - Row stream direct append (HFT draw path, promoted from prior Phase C)

**Root cause:** `drawSharedLayoutRow` cache replay still hits `contentStream.Write(cached)`,
which lines up with the 8–12% `runtime.memmove` band in the Zerodha runs and the
buffer-write / signing-heavy gin sweep.

- [ ] Replace `contentStream.Write(cached)` with append into a pre-sized page `[]byte`.
- [ ] Grow the page stream once per stripe, not per row replay.
- [ ] Right-size `rowBuf.Grow` from the profiled shared-row template size.

**Files:** `internal/pdf/draw.go`, `internal/pdf/pagemanager.go`  
**Estimated gain:** +300–500 ops/sec  
**Compliance risk:** Medium - preserve BDC/EMC order for PDF/UA-2

### P32 - Arena slab right-sizing (HFT)

**Root cause:** 32K-entry slabs (~8 MB each) still dominate 247 MB of live heap, but the
CPU cost in `beginTableRowArena` belongs mostly to the Phase C TR→TD setup path, not to slab sizing alone.

- [ ] Tier arena slabs: 4K / 16K / 32K; HFT should land in the 16K tier, not 32K.
- [ ] Make `arenaCapForNeed(need)` round to the next useful tier and cap HFT-scale requests at 16K.
- [ ] Keep warmup conservative until tiering lands; do not increase the current warm count before smaller slabs exist.
- [ ] Return oversize slabs aggressively and release arena ownership before final pdfBuffer write.

**Files:** `internal/pdf/structure.go`  
**Estimated gain:** +500–900 ops/sec + −120–180 MB heap  
**Compliance risk:** Medium - must not reintroduce P20 slice-header alias race  
**Gate:** arena in-use ≤ 100 MB

### P33 - xref offset slice pooling

**Root cause:** `newXrefOffsets` / `collectUsedXrefObjectIDs` remain visible, but they are
no longer the dominant live-heap driver relative to page streams and pdfBuffer growth.

- [ ] `sync.Pool` `[]int` by cap class once the larger copy-pressure items are down.
- [ ] Lazy grow-on-write without full sentinel pre-fill.
- [ ] Reuse slice across PDFs in the same worker where output stability remains byte-identical.

**Files:** `internal/pdf/generator.go`  
**Estimated gain:** +150–300 ops/sec + −70 MB alloc-space  
**Compliance risk:** Low–Medium - xref table byte-identical

### P35 - Retail signature path cleanup (partial)

**Basis:** Zerodha, gin, and pypdfsuit profiles all keep the signing path visible, but the
benchmark already uses in-place signing.

- [ ] Reduce placeholder scan and `/ByteRange` patch overhead around `embedSignatureInPlace`.
- [ ] Pool PKCS#7/auth-attrs marshal buffers further (extend P14 / prepare for `P42` if needed).
- [ ] Trim `CreateSignatureField` appearance and object-building work where the benchmark does not need extra slack.

**Files:** `internal/pdf/signature/signature.go`  
**Estimated gain:** +250–400 ops/sec  
**Compliance risk:** Medium - `/ByteRange` correctness critical  
**Gate:** openssl/cms verification PASS on signed retail output

### Phase B acceptance gate

- [ ] x10 mean ≥ **13,000 ops/sec** (idle machine)
- [ ] Mean peak allocated ≤ **750 MB**
- [ ] `bytes.growSlice` in-use ≤ 50 MB
- [ ] `runtime.memmove` flat ≤ 6%
- [ ] `runtime.memclrNoHeapPointers` flat ≤ 7%
- [ ] warmed-run stddev ≤ 900 ops/sec before re-introducing the cold first run to the acceptance stats
- [ ] veraPDF 6/6 PASS

---

## Phase C - HFT Tail + Deep Struct Path → ~15,000 ops/sec

**Duration:** 2–3 weeks  
**Projected gain:** +1,500–2,500 ops/sec cumulative  
**Risk:** Medium  
**Basis:** Zerodha x10 run sweep still leaves the HFT struct path as the remaining mixed-workload gate

> **Critical insight:** run1–run5 keep the HFT struct path as the last large compliant
> bottleneck. The exact weighted-latency split should be treated as directional, but
> retail-only and active-only optimizations still do not close the mixed 15K gap.

### P36 - Arena TD row template + bulk init (HFT TR→TD setup) ★ HFT #1 CPU

**Root cause:** the TR→TD setup path stays split across `BeginTableRowWithTDMCIDs`,
`beginTableRowArena`, `appendParentTreeRefs`, parent kid append, and TD field stores.

- [ ] Pre-fill one TD template struct per row (`Type=TD, Parent=tr, PageID, HasMCID=true,
      `tdLeafFast=true`); vary only `MCID` per column.
- [ ] Bulk-init TD fields via copy/template instead of per-field stores.
- [ ] Make `appendParentTreeRefs` pure slice-extend when `PreallocatePageMCIDSlots`
      already reserved capacity (eliminate 140ms/row redundant work).

**Files:** `internal/pdf/structure.go`, `internal/pdf/draw.go`  
**Estimated gain:** +600–900 ops/sec  
**Compliance risk:** **Medium** - must preserve one TD per column with distinct MCID  
**Gate:** `TestBeginTableRowWithTDMCIDs_*` PASS; HFT veraPDF PASS; combined TR→TD setup path cum ≤ 5%

### P38 - Batch struct-object emit (TR `/K [...]` refs + TD leaves)

**Root cause:** `formatStructElemObjectTo` (TR emit / child refs) is now larger than the
TD-leaf helper in the commit-wide Zerodha sweep, so the emit fix needs to cover both.

- [ ] Add `appendStructElemTDLeafBatch(buf, elems []*StructElem, ctx, batchSize=64)`.
- [ ] Fast-path TR `/K [ ... ]` object-ref formatting to avoid per-kid writes in `formatStructElemObjectTo`.
- [ ] Accumulate N leaves or refs into `[8KiB]` stack buffers; single `pdfBuffer.Write` per batch.
- [ ] Pre-assign ObjectIDs in iterative loop for sequential batching.

**Files:** `internal/pdf/generator.go`  
**Estimated gain:** +800–1,200 ops/sec  
**Compliance risk:** Medium - golden-bytes test required  
**Gate:** `TestFormatStructElemTDLeaf_StableOutput` PASS; output byte-identical; `formatStructElemObjectTo` + `appendStructElemTDLeaf` cum ≤ 5%

### P37 - Stripe-batch arena allocation (HFT table)

**Basis:** once `P36` removes the per-row field-store waste, striping still reduces slab churn.

- [ ] Allocate TR+7TD slabs per page stripe (~40 rows) not per row.
- [ ] Keep per-row `BeginTableRowWithTDMCIDs` semantics; batch backing slab extend only.

**Files:** `internal/pdf/structure.go`, `internal/pdf/draw.go`  
**Estimated gain:** +450–700 ops/sec  
**Compliance risk:** Medium - TR→TD per row preserved  
**Gate:** 48-worker race test 0 races

### P39 - Dedup shared-table glyph usage before registry writes

**Root cause:** the hotspot now spans the caller (`markSharedTableCharsUsed`) and the
registry callee (`MarkCharsUsed` / `mapassign_fast32`), not the map write alone.

- [ ] Per-table font charset: dedupe by (font, text) before map writes.
- [ ] Mark unique column text patterns once (HFT columns have repeating values).
- [ ] Optional: bitmap `UsedChars` instead of `map[rune]bool` (P1-02 pattern).

**Files:** `internal/pdf/font/registry.go`, `internal/pdf/draw.go`  
**Estimated gain:** +250–350 ops/sec  
**Compliance risk:** Low - subset unchanged, veraPDF font gate  
**Gate:** `markSharedTableCharsUsed` + `MarkCharsUsed` + `mapassign_fast32` cum ≤ 3%

### P41 - Retail `drawTable` row-batch PDF/UA (retail + active small tables)

**Agents:** A2 - retail uses slow per-cell path for ~125 cells; all tables ≤ 20 rows.

- [ ] Route retail/active `drawTable` rows through `BeginTableRowWithTDMCIDs` for tables
      ≤ 20 rows (all retail/active tables qualify).
- [ ] Use `appendDecimal` in `beginMarkedContentBuf` BDC MCID emit (replace `strconv.AppendInt`).

**Files:** `internal/pdf/draw.go`, `internal/pdf/structure.go`  
**Estimated gain:** +200–350 ops/sec (retail) + +80–120 ops/sec (active, stacks with P29)  
**Compliance risk:** Low - TR→TD per row preserved

### P42 - Hand-built PKCS#7 DER (retail signing)

**Agents:** A2 - ASN.1 reflection 15% of sign chain; fixed cert chain + 3 auth attributes.

- [ ] Replace `encoding/asn1.Marshal` with hand-coded DER for SignedData + ContentInfo.
- [ ] Keep ECDSA P-256 signing unchanged.

**Files:** `internal/pdf/signature/signature.go`  
**Estimated gain:** +150–250 ops/sec  
**Compliance risk:** Medium - openssl/cms verification required

### Phase C acceptance gate

- [ ] x10 mean ≥ **15,000 ops/sec** (idle machine)
- [ ] x10 best ≥ 16,000 ops/sec
- [ ] stddev ≤ **400 ops/sec**
- [ ] Mean peak allocated ≤ **650 MB**
- [ ] combined TR→TD setup path cum ≤ 5%
- [ ] `drawTable` cum ≤ 20%
- [ ] `runtime.memclr` + `memmove` flat combined ≤ 8%
- [ ] veraPDF 6/6 PASS
- [ ] HFT output 2,291,942 ± 5% bytes
- [ ] `make bench-k6-light` no regression

---

## Projected Throughput by Phase

| Phase | Items | Projected mean | Δ vs 9,009 | Confidence |
|-------|-------|---------------:|-----------:|:----------:|
| **Baseline (idle)** | P0–P25 done | 9,009 | - | measured |
| **A - Quick wins** | P26, P28, P30, P29 | 10,400–10,900 | +15–21% | High |
| **B - Memory wall** | P31, P34, P40, P32, P33, P35 | 12,500–13,500 | +39–50% | High |
| **C - HFT tail** | P36, P38, P37, P39, P41, P42 | 14,500–15,500 | +61–72% | Medium |

```
9,009 ──Phase A──► ~10,400–10,900 ──Phase B──► ~13,000 ──Phase C──► ~15,000
         +15–21%                         +44%                  +66%
```

---

## Per-Format Optimization Map

### Retail (80% of iterations, ~25–28% of CPU)

| Priority | Item | Gain | Notes |
|----------|------|-----:|-------|
| P1 | P26 sRGB ICC | +320–560 | 80% weight of all-format win |
| P2 | P35 signature path cleanup | +200–320 | in-place signing already exists; remaining work is smaller |
| P3 | P41 row-batch PDF/UA | +160–280 | 125 cells slow path today |
| P4 | P42 PKCS#7 DER | +120–200 | ASN.1 reflection |
| P5 | P31 pdfBuffer | +640–960 | 80% of buffer pool benefit |

**Retail cannot reach 15K mixed alone** (A2 math: zero retail cost → ~12,200 mixed ceiling).

### Active (15% of iterations, ~7.5% of CPU)

| Priority | Item | Gain | Notes |
|----------|------|-----:|-------|
| P1 | P29 SharedRowLayout | +120–220 | 41-row table eligible, but lower leverage than HFT/memory work |
| P2 | P41 row-batch PDF/UA | +40–60 | Stacks with P29 |
| P3 | P26/P28/P30 | +45–80 | All-format shared wins |
| P4 | P34 page stream caps | +60–90 | 2-page layout |

### HFT (5% of iterations, dominant remaining compliant tail)

| Priority | Item | Gain | Notes |
|----------|------|-----:|-------|
| P1 | P36 arena TD template | +600–900 | **Do not shortcut TR→TD** |
| P2 | P38 batch struct-object emit | +800–1,200 | TR `/K` refs + 14K TD emits |
| P3 | P37 stripe-batch arena | +450–700 | 2000 → ~50 alloc calls |
| P4 | P31 pdfBuffer zero-grow | +400–600 | 2.29 MB output |
| P5 | P40 row stream append | +300–500 | lands in Phase B because it is a copy-pressure fix |
| P6 | P32 arena slab sizing | +250–450 | 8 MB → 4 MB slabs |
| P7 | P39 glyph dedupe | +250–350 | shared-table prescan + registry writes |

---

## Verification Commands

```bash
# Full timing + profile
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache \
  make bench-gopdflib-zerodha-x10-pprof

# Unit tests
go test ./internal/...

# Compliance gate (mandatory after every P-item)
make test-verify-pdfs

# Profile analysis
go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run2.prof
go tool pprof -top       guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run2.prof
go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof

# k6 regression (after Phase B)
make bench-k6-light
```

---

## Related Documents

- `guides/optimizations/20260620_zerodha_x10_pprof_optimization_checklist.md` - P0–P25 (8K target)
- `guides/optimizations/20260617_k6_bench_regression_analysis.md` - key-based cache regression
- `guides/cursor/ZERODHA_BENCHMARK_RESULTS.md` - historical benchmark results
- `guides/cursor/GOPDFLIB_PPROF_RESULTS.md` - data-table bench profiles
