# Zerodha gopdflib x10 pprof Optimization Checklist — 9,000 → 15,000 ops/sec

**Date:** 2026-06-21  
**Workload:** `sampledata/gopdflib/zerodha`, 80% retail / 15% active / 5% HFT  
**Primary command:** `make bench-gopdflib-zerodha-x10`  
**Profile command:** `make bench-gopdflib-zerodha-x5`  
**Combined target:** `make bench-gopdflib-zerodha-x10-pprof`  
**Go version:** `go1.26.4`  
**Concurrency:** 48 workers, `GOMAXPROCS=24`  
**Branch:** `feat/optimization-5.5-medium`  
**Prior checklist:** `guides/optimizations/20260620_zerodha_x10_pprof_optimization_checklist.md` (P0–P25, 8K target met)

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

## Six-Agent Analysis Summary (2026-06-21)

Six specialized subagents analyzed `cpu_zerodha_run3.prof`, `heap_zerodha.prof`, and
the 20260620 checklist. Cross-validation findings below.

### Agent roles and consensus

| Agent | Focus | Key finding | Validated by |
|-------|-------|-------------|--------------|
| **A1 — HFT TR→TD** | Structure tree, arena, TD leaves | HFT = **75–85% of mean latency** despite 5% of docs; `beginTableRowArena` 7.5% CPU | A4, A5, A6 |
| **A2 — Retail signing** | ECDSA P-256, PDF/A metadata | Signature chain **11.4% CPU** (retail-only); retail alone cannot close 15K mixed gap | A4, A6 |
| **A3 — Active trader** | 41-row table, 2-page layout | Active = **7.5% of CPU** despite 15% of docs; `SharedRowLayout` eligible but not enabled | A5 |
| **A4 — Memory/heap** | Pools, GC, variance | **90% live heap** = `bytes.growSlice` (47%) + arena slabs (44%); GC tax **35–40% CPU** | A1, A5, A6 |
| **A5 — Table rendering** | `drawTable`, shared layout | `drawTable` 29% cum is HFT-dominated; text wrap is **0%** — not a gate | A1, A3 |
| **A6 — Cross-format** | Shared bottlenecks, phased path | **P9 sRGB ICC leak** is highest-ROI quick win across all 3 formats (+400–700 ops/sec) | A2, A4 |

### Cross-agent disagreements resolved

| Topic | Agent positions | Resolution |
|-------|----------------|------------|
| Can retail-only opts reach 15K mixed? | A2: No (ceiling ~11.5K); A3: Active adds little | **Agreed:** mixed 15K requires HFT path work |
| Arena slab size | A1: right-size to 16K; A4: tier 4K/16K/32K | **Phase B:** tier slabs; HFT needs ~16K not 32K |
| Active priority | A3: P26 SharedRowLayout +250–350; A1: defer | **Phase A item** — low risk, stacks with HFT wins |
| Variance root cause | A4: pool cold-start + GC; A6: WSL2 load | **Both:** idle-machine gate mandatory (P0) |

---

## Hard Guardrails (Non-Negotiable)

- [ ] **No compliance shortcuts on HFT.** Keep `TR → TD` with one `TD` per column, each
      carrying its own MCID. Do not collapse TDs into bare MCID leaves on `TR`.
- [ ] **No key-based cross-request caches.** No `sync.Map` keyed by row/content/output.
      The bounded `sharedRowRenderCache` may stay as-is but must not expand.
- [ ] **No disabling compliance flags.** `PDFACompliant`, `TaggedPDF`, `ArlingtonCompatible`,
      `EmbedFonts`, retail `Signature` (ECDSA P-256) stay on.
- [ ] **HFT output size stable.** Target **2,291,942 bytes ± 5%** (current compliant size).
- [ ] **veraPDF gate after every phase.** `make test-verify-pdfs` — 6/6 PASS
      (retail/active/HFT × PDF/A-4 + PDF/UA-2).
- [ ] **k6 must not regress.** `make bench-k6-light` with `drawSharedLayoutRow` flat heap ≤ 10 MB.

---

## Fresh Measurement (2026-06-21 `make bench-gopdflib-zerodha-x10-pprof`)

Build cache reset: `GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache`

### x10 timing (this session — machine under load)

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

### Fresh CPU profile (`cpu_zerodha_run3.prof`, 11.27s samples, 1771% CPU)

| Hotspot | Cum % | Flat % | Format scope | Status |
|---------|------:|-------:|--------------|--------|
| `GenerateTemplatePDFBorrowed` | 91.8% | 4.3% | All | Top gate |
| `generateAllContentWithImages` | 31.0% | 0.2% | All | Content emit |
| `drawTable` | 29.0% | 1.4% | HFT-heavy | Table render |
| `drawSharedLayoutRow` | 12.6% | 0.4% | HFT only | Shared layout |
| `runtime.memclrNoHeapPointers` | 11.3% | **11.3%** | All (GC-driven) | **Open gate** |
| `runtime.memmove` | 10.1% | **10.1%** | All | Buffer growth |
| `beginTableRowArena` | 7.5% | **5.8%** | HFT only | TR→TD alloc |
| `signature.*` (SignPDF chain) | 6.3% | — | Retail 80% | Signing |
| `buildSRGBICCProfile` + `math.pow` | 4.5% | 2.7% | **All 3 equally** | **P9 leak** |
| `appendStructElemTDLeaf` | 4.2% | 1.6% | HFT-heavy | TD leaf emit |
| `MarkCharsUsed` | 4.1% | 0.9% | HFT-heavy | Font subset |
| `formatStructElemObjectTo` | 3.6% | 2.4% | Tagged (HFT-heavy) | Slow path |
| `estimateFinalPDFSize` | — | 1.5% | All | Size estimate |

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
| HFT output size | 2,291,942 bytes | stable ±5% | — |

---

## Phased Execution Plan

```
Phase A (Quick wins)     → ~11,000 ops/sec  (+22%)
Phase B (Memory wall)    → ~13,000 ops/sec  (+44%)
Phase C (HFT tail)       → ~15,000 ops/sec  (+66%)
```

---

## Phase A — Cross-Format Quick Wins → ~11,000 ops/sec

**Duration:** 1–2 days  
**Projected gain:** +1,000–1,500 ops/sec from idle 9,009 baseline  
**Risk:** Low

### P0 — Harness and measurement hygiene

- [ ] Always run with `GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache`.
- [ ] Run x10 on an **idle machine** (no parallel benchmarks, Docker, browser).
- [ ] Use x10 **mean** as regression gate; capture `zerodha_bench_x10_wsl_stats_latest.txt` after every accepted change.
- [ ] Re-profile after Phase A: `make bench-gopdflib-zerodha-x10-pprof`.
- [ ] **Gate:** stddev ≤ 600 ops/sec on idle machine.

### P26 — Fix sRGB ICC cache leak (all 3 templates) ★ Highest ROI

**Agents:** A2, A4, A6 unanimous — P9 marked done but sRGB still rebuilt per PDF.

**Root cause:** `GetSRGBICCProfile()` calls `buildSRGBICCProfile()` (1024× `math.Pow`)
on every PDF. `GenerateOutputIntent` line 306 uses `getSRGBICCProfile()` for `Grow()`
sizing even though `compressedSRGBICCProfileCache()` is used at emit time.

**Fix:**
- [ ] Cache uncompressed sRGB bytes at `init` alongside `srgbICCProfileCompressed`.
- [ ] Change `GetSRGBICCProfile()` to return cached bytes.
- [ ] Use `len(srgbICCProfileCompressed)` for `sb.Grow()` in `GenerateOutputIntent`.

**Files:** `internal/pdf/pdfa.go`, `internal/pdf/metadata.go`  
**Estimated gain:** +400–700 ops/sec (all 5,000 iterations)  
**Compliance risk:** Low — byte-identical ICC profile  
**Gate:** `buildSRGBICCProfile` not in top-40 CPU; veraPDF 6/6 PASS

### P27a — Cache `estimateFinalPDFSize` per template class

**Agents:** A6 identified 1.5% flat CPU on size estimation.

- [ ] Pre-compute retail/active/HFT final size estimates at template build time in `bench.go`.
- [ ] Pass through `GenerateTemplatePDFBorrowed` or store on template metadata.

**Files:** `internal/pdf/generator.go`, `sampledata/gopdflib/zerodha/bench.go`  
**Estimated gain:** +100–200 ops/sec  
**Compliance risk:** Low

### P28 — Precompute standard fonts at template build

**Agents:** A6 — `collectAllStandardFontsInTemplate` 2.9% cum, hits every PDF.

- [ ] Build font set once when `buildRetailTemplate` / `buildActiveTraderTemplate` / `buildHFTTemplate` run.
- [ ] Store on template struct; skip per-PDF font walk.

**Files:** `internal/pdf/generator.go`, `sampledata/gopdflib/zerodha/bench.go`  
**Estimated gain:** +200–350 ops/sec  
**Compliance risk:** Low

### P29 — Active trader `SharedRowLayout` enablement

**Agents:** A3 primary; validated by A5 as compliance-safe.

**Root cause:** 41-row trade table has uniform `Props` per column; only `Text`/`TextColor`
vary. Passes `tableSupportsSharedRowLayout` but flag is not set.

- [ ] Add `SharedRowLayout: true, SharedRowTemplateRow: 1` to active trade table in
      `buildActiveTraderTemplate()`.
- [ ] Verify alternating `BgColor` and per-row `TextColor` still render correctly.
- [ ] Confirm veraPDF active output PASS.

**Files:** `sampledata/gopdflib/zerodha/bench.go`, `internal/pdf/draw.go`  
**Estimated gain:** +250–350 ops/sec (active is 15% of docs, 7.5% of CPU)  
**Compliance risk:** Low — TR→TD hierarchy preserved via shared layout path  
**Gate:** `zerodha_active_output.pdf` 76,050 ± 5% bytes; veraPDF 2/2 PASS

### P30 — Static XMP metadata shell (all 3 templates)

**Agents:** A2 — `GenerateXMPMetadata` 1.0% cum; dates/ID are only per-PDF variance.

- [ ] Pre-build XMP packet templates with placeholder slots for date/ID.
- [ ] Patch only variable fields at emit time.

**Files:** `internal/pdf/metadata.go`  
**Estimated gain:** +80–120 ops/sec  
**Compliance risk:** Low — PDF/A-4 + PDF/UA-2 metadata fields unchanged

### Phase A acceptance gate

- [ ] x10 mean ≥ **11,000 ops/sec** (idle machine)
- [ ] veraPDF 6/6 PASS
- [ ] HFT output 2,291,942 ± 5% bytes
- [ ] `buildSRGBICCProfile` cum CPU ≤ 0.5%

---

## Phase B — Memory Wall → ~13,000 ops/sec

**Duration:** 1 week  
**Projected gain:** +2,000–2,500 ops/sec cumulative  
**Risk:** Low–Medium  
**Agents:** A4 (memory), A1 (HFT buffers), A6 (cross-format) aligned on this phase

### P31 — pdfBuffer zero-grow (all 3 templates)

**Root cause:** `bytes.growSlice` = 263 MB in-use (47%); `memmove` = 10.1% flat CPU.
HFT cap 2.5 MiB but output 2.29 MiB + struct-tree overhead still triggers mid-emit grow.

- [ ] Set `hftPDFBufferCap = 2_292_096` (actual output + 512 byte margin).
- [ ] Fix `estimateFinalPDFSize` to include ParentTree `strings.Builder` (~40 KiB) + signature gap.
- [ ] Add `pdfBufferPoolMedium` (512 KiB–2 MiB) for active tier dead-zone.
- [ ] Warm pools to worker count: `initPDFBufferPools(48)` not `GOMAXPROCS(24)`.
- [ ] Add `BENCH_DEBUG_CAPS=1` assert: zero `Grow` calls after initial cap.

**Files:** `internal/pdf/generator.go`  
**Estimated gain:** +800–1,200 ops/sec + −200 MB heap  
**Compliance risk:** Low — sizing only  
**Gate:** `bytes.growSlice` in-use ≤ 50 MB; `memmove` flat ≤ 5%

### P32 — Arena slab right-sizing (HFT)

**Root cause:** 32K-entry slabs (~8 MB each) for ~16K TD need; 247 MB in-use (44% heap).
`WarmArenaSlabPool(6)` for 48 workers → cold alloc on first HFT per worker cohort.

- [ ] Tier arena slabs: 4K / 16K / 32K; HFT uses 16K tier (~4 MB not 8 MB).
- [ ] `arenaCapForNeed(need)`: ceil to next power of 2 capped at 16K for HFT-scale.
- [ ] Warm `min(24, expectedConcurrentHFT)` slabs at init.
- [ ] Return undersized slabs aggressively; release arena before final pdfBuffer write.

**Files:** `internal/pdf/structure.go`  
**Estimated gain:** +500–900 ops/sec + −120–180 MB heap  
**Compliance risk:** Medium — must not reintroduce P20 slice-header alias race  
**Gate:** arena in-use ≤ 100 MB; `beginTableRowArena` cum ≤ 5%

### P33 — xref offset slice pooling

**Root cause:** `newXrefOffsets` + `collectUsedXrefObjectIDs` = 156 MB alloc per run;
`newXrefOffsets` zero-fills 16K slots (2.2% CPU).

- [ ] `sync.Pool` of `[]int` keyed by cap class (small/medium/large).
- [ ] Lazy grow-on-write without full sentinel pre-fill.
- [ ] Reuse slice across PDFs in same worker.

**Files:** `internal/pdf/generator.go`  
**Estimated gain:** +250–400 ops/sec + −70 MB alloc  
**Compliance risk:** Low–Medium — xref table byte-identical

### P34 — Page content stream exact caps (HFT + active)

**Root cause:** `getPageContentStreamBuffer` 157 MB cum; HFT pages ~90–120 KiB land in
128 KiB bucket but still grow on some pages.

- [ ] Profile max page stream length per template class.
- [ ] Pre-cap page streams to measured max + 256 byte margin.
- [ ] Pool 48 × 50 × 128 KiB bounded for HFT steady state.

**Files:** `internal/pdf/pagemanager.go`  
**Estimated gain:** +400–600 ops/sec + −60–90 MB heap  
**Compliance risk:** Medium

### P35 — Retail signature buffer optimization (partial)

**Agents:** A2 — signature chain 6.3% cum; retail is 80% of workload.

- [ ] Eliminate full-PDF memclr copy in `embedSignatureInPlace` — patch `/Contents` +
      `/ByteRange` in pooled buffer.
- [ ] Pool ASN.1 marshal buffers further (extend P14).
- [ ] Defer `CreateSignatureField` appearance alloc where possible.

**Files:** `internal/pdf/signature/signature.go`  
**Estimated gain:** +400–600 ops/sec  
**Compliance risk:** Medium — `/ByteRange` correctness critical  
**Gate:** openssl/cms verification PASS on signed retail output

### Phase B acceptance gate

- [ ] x10 mean ≥ **13,000 ops/sec** (idle machine)
- [ ] Mean peak allocated ≤ **750 MB**
- [ ] `bytes.growSlice` in-use ≤ 50 MB
- [ ] `runtime.memclrNoHeapPointers` flat ≤ 5%
- [ ] stddev ≤ 500 ops/sec
- [ ] veraPDF 6/6 PASS

---

## Phase C — HFT Tail + Deep Struct Path → ~15,000 ops/sec

**Duration:** 2–3 weeks  
**Projected gain:** +1,500–2,500 ops/sec cumulative  
**Risk:** Medium  
**Agents:** A1 (HFT), A5 (table), A2 (retail ceiling) — HFT tail is mandatory for 15K

> **Critical insight (A1 + A2 cross-validation):** HFT is 5% of documents but **75–85% of
> mean latency**. Every 10% HFT speedup ≈ +750–850 ops/sec on the mixed benchmark.
> Retail-only optimizations ceiling ≈ 11,500 ops/sec mixed — Phase C is non-optional.

### P36 — Arena TD row template + bulk init (HFT TR→TD) ★ HFT #1 CPU

**Root cause:** `beginTableRowArena` 7.5% cum; `td.Parent = tr` loop alone is 370ms flat
(7 field writes × 7 cols × 1999 rows × 250 HFT PDFs).

- [ ] Pre-fill one TD template struct per row (`Type=TD, Parent=tr, PageID, HasMCID=true,
      `tdLeafFast=true`); vary only `MCID` per column.
- [ ] Bulk-init TD fields via copy/template instead of per-field stores.
- [ ] Make `appendParentTreeRefs` pure slice-extend when `PreallocatePageMCIDSlots`
      already reserved capacity (eliminate 140ms/row redundant work).

**Files:** `internal/pdf/structure.go`, `internal/pdf/draw.go`  
**Estimated gain:** +600–900 ops/sec  
**Compliance risk:** **Medium** — must preserve one TD per column with distinct MCID  
**Gate:** `TestBeginTableRowWithTDMCIDs_*` PASS; HFT veraPDF PASS; `beginTableRowArena` cum ≤ 3%

### P37 — Stripe-batch arena allocation (HFT table)

**Agents:** A5 primary — cut 2000 arena calls → ~50 per-page stripes.

- [ ] Allocate TR+7TD slabs per page stripe (~40 rows) not per row.
- [ ] Keep per-row `BeginTableRowWithTDMCIDs` semantics; batch backing slab extend only.

**Files:** `internal/pdf/structure.go`, `internal/pdf/draw.go`  
**Estimated gain:** +450–700 ops/sec  
**Compliance risk:** Medium — TR→TD per row preserved  
**Gate:** 48-worker race test 0 races

### P38 — Batch TD struct-object writes (HFT emit)

**Root cause:** 14K calls to `appendStructElemTDLeaf` each do `Grow` + `Write` + 3×
`appendDecimal` (4.2% cum).

- [ ] Add `appendStructElemTDLeafBatch(buf, elems []*StructElem, ctx, batchSize=64)`.
- [ ] Accumulate N leaves into `[8KiB]` stack buffer; single `pdfBuffer.Write` per batch.
- [ ] Pre-assign ObjectIDs in iterative loop for sequential batching.

**Files:** `internal/pdf/generator.go`  
**Estimated gain:** +800–1,200 ops/sec  
**Compliance risk:** Medium — golden-bytes test required  
**Gate:** `TestFormatStructElemTDLeaf_StableOutput` PASS; output byte-identical

### P39 — Batch `MarkCharsUsed` (HFT font subsetting)

**Root cause:** `MarkCharsUsed` 4.1% cum; `mapassign_fast32` 2.9% from per-rune map writes.
14K cells × Liberation TTF (PDF/A substitution).

- [ ] Per-table font charset: dedupe by (font, text) before map writes.
- [ ] Mark unique column text patterns once (HFT columns have repeating values).
- [ ] Optional: bitmap `UsedChars` instead of `map[rune]bool` (P1-02 pattern).

**Files:** `internal/pdf/font/registry.go`, `internal/pdf/draw.go`  
**Estimated gain:** +250–350 ops/sec  
**Compliance risk:** Low — subset unchanged, veraPDF font gate  
**Gate:** `MarkCharsUsed` cum ≤ 2%

### P40 — Row stream direct append (HFT draw path)

**Root cause:** `drawSharedLayoutRow` cache replay via `contentStream.Write(cached)` drives
`memmove` (460ms cum in draw path).

- [ ] Replace `contentStream.Write(cached)` with append into pre-sized page `[]byte`.
- [ ] Grow page stream once per stripe, not per row.
- [ ] Right-size `rowBuf.Grow` from profiled ~800–1200 B row template.

**Files:** `internal/pdf/draw.go`  
**Estimated gain:** +300–500 ops/sec  
**Compliance risk:** Medium — BDC/EMC nesting order for PDF/UA-2

### P41 — Retail `drawTable` row-batch PDF/UA (retail + active small tables)

**Agents:** A2 — retail uses slow per-cell path for ~125 cells; all tables ≤ 20 rows.

- [ ] Route retail/active `drawTable` rows through `BeginTableRowWithTDMCIDs` for tables
      ≤ 20 rows (all retail/active tables qualify).
- [ ] Use `appendDecimal` in `beginMarkedContentBuf` BDC MCID emit (replace `strconv.AppendInt`).

**Files:** `internal/pdf/draw.go`, `internal/pdf/structure.go`  
**Estimated gain:** +200–350 ops/sec (retail) + +80–120 ops/sec (active, stacks with P29)  
**Compliance risk:** Low — TR→TD per row preserved

### P42 — Hand-built PKCS#7 DER (retail signing)

**Agents:** A2 — ASN.1 reflection 15% of sign chain; fixed cert chain + 3 auth attributes.

- [ ] Replace `encoding/asn1.Marshal` with hand-coded DER for SignedData + ContentInfo.
- [ ] Keep ECDSA P-256 signing unchanged.

**Files:** `internal/pdf/signature/signature.go`  
**Estimated gain:** +150–250 ops/sec  
**Compliance risk:** Medium — openssl/cms verification required

### Phase C acceptance gate

- [ ] x10 mean ≥ **15,000 ops/sec** (idle machine)
- [ ] x10 best ≥ 16,000 ops/sec
- [ ] stddev ≤ **400 ops/sec**
- [ ] Mean peak allocated ≤ **650 MB**
- [ ] `beginTableRowArena` cum ≤ 3%
- [ ] `drawTable` cum ≤ 20%
- [ ] `runtime.memclr` + `memmove` flat combined ≤ 8%
- [ ] veraPDF 6/6 PASS
- [ ] HFT output 2,291,942 ± 5% bytes
- [ ] `make bench-k6-light` no regression

---

## Projected Throughput by Phase

| Phase | Items | Projected mean | Δ vs 9,009 | Confidence |
|-------|-------|---------------:|-----------:|:----------:|
| **Baseline (idle)** | P0–P25 done | 9,009 | — | measured |
| **A — Quick wins** | P26–P30 | 10,500–11,000 | +11–22% | High |
| **B — Memory wall** | P31–P35 | 12,500–13,500 | +39–50% | High |
| **C — HFT tail** | P36–P42 | 14,500–15,500 | +61–72% | Medium |

```
9,009 ──Phase A──► ~11,000 ──Phase B──► ~13,000 ──Phase C──► ~15,000
         +22%                  +44%                  +66%
```

---

## Per-Format Optimization Map

### Retail (80% of iterations, ~25–28% of CPU)

| Priority | Item | Gain | Notes |
|----------|------|-----:|-------|
| P1 | P26 sRGB ICC | +320–560 | 80% weight of all-format win |
| P2 | P35 signature embed | +320–480 | 11.4% CPU retail-only |
| P3 | P41 row-batch PDF/UA | +160–280 | 125 cells slow path today |
| P4 | P42 PKCS#7 DER | +120–200 | ASN.1 reflection |
| P5 | P31 pdfBuffer | +640–960 | 80% of buffer pool benefit |

**Retail cannot reach 15K mixed alone** (A2 math: zero retail cost → ~12,200 mixed ceiling).

### Active (15% of iterations, ~7.5% of CPU)

| Priority | Item | Gain | Notes |
|----------|------|-----:|-------|
| P1 | P29 SharedRowLayout | +250–350 | 41-row table eligible |
| P2 | P41 row-batch PDF/UA | +40–60 | Stacks with P29 |
| P3 | P26/P28/P30 | +45–80 | All-format shared wins |
| P4 | P33 page stream caps | +60–90 | 2-page layout |

### HFT (5% of iterations, ~50–60% of CPU, ~75–85% of mean latency)

| Priority | Item | Gain | Notes |
|----------|------|-----:|-------|
| P1 | P36 arena TD template | +600–900 | **Do not shortcut TR→TD** |
| P2 | P37 stripe-batch arena | +450–700 | 2000 → ~50 alloc calls |
| P3 | P38 batch TD leaf write | +800–1,200 | 14K struct emits |
| P4 | P31 pdfBuffer zero-grow | +400–600 | 2.29 MB output |
| P5 | P32 arena slab sizing | +250–450 | 8 MB → 4 MB slabs |
| P6 | P39 MarkCharsUsed batch | +250–350 | 14K cells |
| P7 | P40 row stream append | +300–500 | Cache replay memmove |

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
go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -top       guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof

# k6 regression (after Phase B)
make bench-k6-light
```

---

## Related Documents

- `guides/optimizations/20260620_zerodha_x10_pprof_optimization_checklist.md` — P0–P25 (8K target)
- `guides/optimizations/20260617_k6_bench_regression_analysis.md` — key-based cache regression
- `guides/cursor/ZERODHA_BENCHMARK_RESULTS.md` — historical benchmark results
- `guides/cursor/GOPDFLIB_PPROF_RESULTS.md` — data-table bench profiles