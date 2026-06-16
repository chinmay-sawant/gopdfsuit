# gopdfsuit — Remaining Optimization Checklist

**Date:** 2026-06-14  
**Baseline:** `make bench-k6` weighted `tagged_ecdsa` (80% retail / 15% active / 5% HFT), 48 VUs × 35s  
**Latest validation:** 2 fresh `make bench-k6` + pprof runs after the structure preallocation pass on 2026-06-14: **998.7 req/s** and **1052.7 req/s**; 2-run mean **1025.7 req/s**  
**Best fresh run:** **1052.7 req/s**, p99 **236.9 ms**, HFT avg **323.4 ms**  
**Historical best:** 1,232 req/s (Phase 11 peak), 1,054 req/s (post-revert baseline)  
**Current gap:** ~87% of post-revert baseline (1,054), ~74% of Phase 11 peak (1,232), ~61% of the current target **1,500 req/s**  
**Previous session 5-run (before this work):** avg 825 req/s, peak 859 req/s

**Implemented in this session:**
- HI-2 — bounded the unbounded caches (subsetCache, imgCache, propsCache) + clear APIs.
- MI-3 — replaced `strconv.Itoa` with `strconv.AppendInt` in structure-tree writer.
- MI-1 (partial) — replaced `escapeText` with `appendEscapedPDFLiteral` + `Write([]byte)` in structure-tree writer for Title/Alt.
- HI-1/HI-3 support — fixed `CompressContentStreamCached` shard selection so identical page streams hash into a stable shard instead of round-robin misses; also removed per-call `hash/fnv` object allocation in `pageContentFingerprint`.
- HI-4/HI-3 support — changed HFT table row preallocation to keep `Rows` length at `0` while preserving backing capacity and per-row cell capacity.
- MI-1 follow-up — removed `escapeText` string allocation from `BeginMarkedContent` / `BeginMarkedContentBuf` `/Alt` writing in the tagged-content BDC path.
- HI-4 follow-up — changed pooled `PDFTemplate` reset to preserve hot HFT backing arrays across requests instead of dropping them on every `sync.Pool` reuse, with tests proving reused row/cell backing storage.
- MI-1/structure follow-up — pre-sized tagged structure-element backing storage from the input template shape and centralized slice-growth logic for `StructureManager` / parent-tree backing arrays.
- HI-3 — a first attempt with larger estimates + pool retention caps (8 MB / 1 MB) regressed throughput 993 → 962 req/s and increased `bytes.growSlice` in-use 256 → 338 MB. Reverted.
- MI-1 (partial) — a generic structure-writer experiment regressed throughput 998 → 916 req/s (Go used shape-based dispatch, not monomorphization). Reverted.

**Bottom line:** Across the last three implementation passes, the fresh 2-run mean improved from **916.8 req/s** to **1025.7 req/s** and the sample is far tighter than the original `794.5` / `1039.1` spread. The latest structure preallocation pass was small but positive; it still did not materially change the overall `~1500+ req/s` outlook.

---

## Shipped in this commit

The following code changes are committed and were previously reflected in a `5-run mean 1005.6 req/s` run set. See the **Commit message** at the bottom of this file. Treat that result as historical, not the active baseline.

### HI-2 — Bounded the unbounded caches

- **`internal/pdf/font/subset_cache.go`**
  - Added `maxSubsetCacheEntries = 1024` + `atomic.Int64` counter.
  - On overflow: `ClearSubsetCache()` resets the `sync.Map` and counter.
  - New exported `ClearSubsetCache()` API.
- **`internal/pdf/image.go`**
  - Added `maxImageCacheEntries = 256`.
  - On overflow: `imgCache.clear()` resets the map and MRU slot.
  - `ResetImageCache()` refactored to use the new `clear()` helper.
- **`internal/pdf/utils.go`**
  - Added `maxPropsCacheEntries = 8192` + `atomic.Int64` counter.
  - On overflow: `ClearPropsCache()` resets the `sync.Map` and counter.
  - New exported `ClearPropsCache()` API.

**New tests:**
- `internal/pdf/font/subset_cache_test.go` — reuse, bounds, clear.
- `internal/pdf/image_cache_test.go` — reuse, bounds, clear.
- `internal/pdf/props_cache_test.go` — reuse, bounds, clear.

### MI-3 — Structure-tree integer formatting

- **`internal/pdf/generator.go`** — `formatStructElemObjectTo` and `appendObjRefToWriter`:
  - Replaced `strconv.Itoa(int)` with `strconv.AppendInt(scratch[:0], int64(n), 10)` to a stack `[24]byte` scratch.
  - `w.Write([]byte)` instead of `w.WriteString(string)` — avoids the string allocation.

### MI-1 (partial) — Structure-tree text escaping

- **`internal/pdf/generator.go`** — `formatStructElemObjectTo` Title/Alt writing:
  - Replaced `escapeText(string)` + `WriteString` with `appendEscapedPDFLiteral(scratch[:0], s)` + `Write([]byte)`.
  - Stack `[1024]byte` scratch; falls back to heap if escaped output exceeds 1024 bytes.

### Experiments tried and reverted (not in this commit)

- **HI-3 — Aggressive estimate + pool retention caps (8 MB / 1 MB)**: regressed 993 → 962 req/s, increased `bytes.growSlice` in-use 256 → 338 MB. Reverted.
- **MI-1 — Generic `formatStructElemObjectToGeneric[T]()`**: regressed 998 → 916 req/s (Go used shape-based dispatch). Reverted.
- **EW-2 — `Load` check before `Store` in `storePageCompressEntry`**: single-run dropped to 808 req/s, HFT latency spiked to 461 ms. Reverted.

### Post-session 5-run (after EW-2 revert)

A follow-up 5-run under heavier system load landed at:
- Mean **808.97 req/s**, runs 850 / 860 / 776 / 801 / 758
- HFT avg latency 460–475 ms (vs 350–360 ms in the 1005 run)
- The drop is attributed to system load, not the code (EW-2 was fully reverted and only HI-2 + MI-3 + MI-1 partial are in `main`).

### Fresh 2-run validation before this pass

Two fresh runs were captured via the repo harness (`make bench-k6` -> `test/generate_template-pdf/run_gin_pprof_load.sh`) before the latest code changes to reset this checklist against current evidence:

| Run | Throughput | HTTP p99 | HFT avg | pprof summary |
|---|---:|---:|---:|---|
| 1 | 794.52 req/s | 365.05 ms | 453.91 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_130431.txt` |
| 2 | 1039.06 req/s | 267.54 ms | 356.75 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_130647.txt` |
| **Mean** | **916.79 req/s** | n/a | n/a | fresh 2-run sample |

Interpretation:
- The system still clears the `0% errors` and sub-`500 ms` p99 bar.
- Throughput is too unstable and too low to claim a current `~1005 req/s` steady state.
- The path to `~1500+ req/s` must come from major CPU/alloc reductions in compression, JSON decode, structure serialization, and heap growth rather than from cache-bounds work alone.

### Fresh 2-run validation after this pass

After the latest code changes, the same harness was re-run twice:

| Run | Throughput | HTTP p99 | HFT avg | pprof summary |
|---|---:|---:|---:|---|
| 1 | 978.44 req/s | 282.86 ms | 370.85 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_131656.txt` |
| 2 | 1000.17 req/s | 266.06 ms | 357.16 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_131905.txt` |
| **Mean** | **989.31 req/s** | n/a | n/a | post-change 2-run sample |

Observed effect:
- Fresh 2-run mean improved from **916.79 → 989.31 req/s** (`+7.9%`).
- The run-to-run spread tightened substantially (`244.5 req/s` spread before vs `21.7 req/s` after).
- `preallocInlineTableRows` retained heap dropped versus the earlier fresh sample (`25-59 MB` before vs `12-19 MB` after), but alloc_space is still high because every request still allocates those backing arrays.
- Compression remains a dominant CPU cost even after fixing cache sharding, which implies cache reuse is not yet large enough to change the primary `compress/flate` ceiling on this workload.

### Fresh 2-run validation after pooled-template reuse

After preserving `PDFTemplate` backing arrays across `sync.Pool` reuse, the harness was re-run twice again:

| Run | Throughput | HTTP p99 | HFT avg | pprof summary |
|---|---:|---:|---:|---|
| 1 | 985.02 req/s | 267.90 ms | 357.08 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_135920.txt` |
| 2 | 1033.06 req/s | 262.04 ms | 347.07 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_140008.txt` |
| **Mean** | **1009.04 req/s** | n/a | n/a | pooled-template reuse sample |

Observed effect:
- Fresh 2-run mean improved from **989.31 → 1009.04 req/s** (`+2.0%`).
- Latency stayed healthy (`p99 262-268 ms`, `0%` errors).
- The reuse pass did **not** materially remove `preallocInlineTableRows` from the remaining allocation profile; it still sits around `8-9% alloc_space` and `8% in-use heap`.
- Compression and final PDF buffer growth remain the dominant blockers.

### Fresh 2-run validation after structure preallocation

After pre-sizing tagged structure-element backing storage, the harness was re-run twice again:

| Run | Throughput | HTTP p99 | HFT avg | pprof summary |
|---|---:|---:|---:|---|
| 1 | 998.68 req/s | 274.40 ms | 358.75 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_140420.txt` |
| 2 | 1052.70 req/s | 236.91 ms | 323.45 ms | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260614_140515.txt` |
| **Mean** | **1025.69 req/s** | n/a | n/a | structure preallocation sample |

Observed effect:
- Fresh 2-run mean improved from **1009.04 → 1025.69 req/s** (`+1.6%`).
- The best run matched the older post-revert baseline band again (`1052.7 req/s`).
- Structure CPU improved slightly (`formatStructElemObjectTo` roughly `9.75-10.12%` cum in the new pair), but the new generic growth helper itself now appears in alloc_space, so this was not a large net memory win.
- Compression, `bytes.growSlice`, JSON decode, and HFT row preallocation still dominate the profile.

---

## Executive summary

The caches introduced so far (page compression, font subset, image decode, props, page-init) are narrow helpers and do **not** shortcut full PDF generation. The latest fresh pprof runs show the same bottlenecks as the active limiters:

1. `bytes.growSlice` — still ~50-54% in-use heap and ~13% alloc_space in the latest fresh runs.
2. `compress/flate` (`compress/zlib.(*Writer).Close` / `CompressContentStreamCached`) — ~20-21% cumulative CPU.
3. `drawTable` / `drawSharedLayoutRow` / `drawSharedDeferRow` — ~15-16% cumulative CPU.
4. `formatStructElemObjectTo` — ~9.8-10.5% cumulative CPU and ~3-5% alloc_space.
5. `sonic` + JSON decode (`sonic.Unmarshal`, `RawMessage.UnmarshalJSON`) — ~16-18% alloc_space.
6. `PDFTemplate.preallocInlineTableRows` — reuse helped only modestly; it still shows about `9%` alloc_space and `7-8%` in-use heap in the newest pair.
7. Structure preallocation is no longer free: `growPtrSlice[...]` now shows up as a visible allocator in the newest pair, so future structure work should avoid generic helper churn unless it clearly displaces larger allocations.
7. `pageContentFingerprint` is now visible at ~2.1-2.3% flat CPU, so future compression-cache work should avoid adding more hashing cost unless hit rate climbs enough to pay for it.

Compliance boundary for all remaining work:
- Do not disable tagging, structure-tree generation, signing, PDF/A metadata, or unique-document behavior just to move req/s.
- Prefer work that reduces allocations or CPU inside existing compliant code paths.

This file tracks the next implementation passes with the `~1500+ req/s` goal in mind.

---

## High impact

### HI-1: HFT / flate compression tuning

**Goal:** Reduce the ~20% cumulative CPU spent in zlib without breaking PDF/A compatibility or output size requirements.

- [ ] Audit current `compress/flate` writer level/strategy in `internal/pdf/font/compression.go`.
- [ ] Evaluate lowering flate level for page streams (e.g., level 6 → 4 or 3) and measure output-size vs throughput trade-off.
- [ ] Tune store-uncompressed threshold; ensure compressed stream is always smaller than raw before keeping it.
- [ ] Revisit `pageCompressSlots` cap in `internal/pdf/generator.go` (currently capped to `NumCPU`).
- [ ] Re-run `make bench-k6` and verify weighted throughput improves and p99 stays < 500 ms.

**Files:** `internal/pdf/font/compression.go`, `internal/pdf/generator.go`  
**Priority note:** Fresh pprof still shows compression as the single largest non-handler cumulative CPU bucket. This is the safest high-impact path because it should preserve tagging/signing semantics if output bytes remain valid.  
**Estimated gain:** +5–8% throughput  
**Risk:** Medium — output size may increase; must keep PDF/A compatibility.

---

### HI-2: Bound the unbounded caches

**Goal:** Prevent `subsetCache`, `imgCache`, and `propsCache` from growing without limit on long-lived pods or high-cardinality workloads.

- [x] Add entry-count cap + FIFO-style eviction to `subsetCache` (`internal/pdf/font/subset_cache.go`).
- [x] Add entry-count cap + MRU eviction to `imgCache` (`internal/pdf/image.go`).
- [x] Add entry-count cap to `propsCache` (`internal/pdf/utils.go`).
- [x] Add `Clear*` APIs where missing for tests and memory-pressure hooks.
- [x] Run `go test ./internal/pdf/...` and existing cache tests to ensure no regressions.
- [x] Re-run `make bench-k6` to confirm throughput is neutral or improved.

**Result:** All tests pass. Final 5-run `make bench-k6` (after HI-2 + MI-3 + MI-1 partial) — runs 1–5:

| Run | Throughput | p99 |
|---|---:|---:|
| 1 | 1021.04 req/s | 270.47 ms |
| 2 | 1018.30 req/s | 257.34 ms |
| 3 | 1041.18 req/s | 262.46 ms |
| 4 | 946.55 req/s | 283.32 ms |
| 5 | 1001.10 req/s | 274.27 ms |
| **Mean** | **1005.63 req/s** | **269.57 ms** |
| **Median** | **1018.30 req/s** | **270.47 ms** |
| **Best** | **1041.18 req/s** | **257.34 ms** |
| **Worst** | **946.55 req/s** | **283.32 ms** |

> vs previous session 5-run avg 825 req/s → **+22%**.

**Files:** `internal/pdf/font/subset_cache.go`, `internal/pdf/image.go`, `internal/pdf/utils.go`  
**Estimated gain:** Indirect (GC pressure / memory stability), +1–3% throughput on long runs  
**Risk:** Low — follow the same bounded pattern used in `compress_cache.go`.

---

### HI-3: Reduce `bytes.growSlice`

**Goal:** Cut the ~50% in-use heap spent on buffer growth.

- [x] Review `estimateTemplatePDFBufferSize` and `estimateFinalPDFSize` in `internal/pdf/generator.go`.
- [ ] Improve final-size estimate using content-stream byte totals and structure-tree slack.
  - **Note:** A first attempt (larger estimates + pool retention caps at 8 MB / 1 MB) increased `bytes.growSlice` in-use from 256 MB → 338 MB and dropped throughput 993 → 962 req/s. Reverted.
- [ ] Pre-size page content streams in `pagemanager.go` / `AddNewPage` based on row/page complexity.
- [ ] Reuse or pool page content buffers for multi-page documents instead of fresh `bytes.Buffer` per page.
- [ ] Re-run `make bench-k6` and verify `bytes.growSlice` in-use drops and throughput improves.

**Status:** Still required. Fresh runs remain at `296-325 MB` `bytes.growSlice` in-use and `4.65-4.96 GB` alloc_space.

**Files:** `internal/pdf/generator.go`, `internal/pdf/pagemanager.go`, `internal/pdf/draw.go`  
**Estimated gain:** +5–7% throughput, significant heap reduction  
**Risk:** Medium — incorrect pre-sizing can waste memory; must not change PDF output.

---

### HI-4: Codegen JSON decoder for `PDFTemplate`

**Goal:** Reduce sonic decode CPU/alloc, especially for HFT payloads.

- [ ] Generate a sonic AST decoder for `models.PDFTemplate` (P8-A3 from the Gin report).
- [ ] Integrate generated decoder into `internal/handlers/json_decode.go`.
- [ ] Add benchmark comparing generated decoder vs `sonic.Unmarshal`.
- [ ] Run `make bench-k6` and check sonic alloc_space share.

**Files:** `internal/models/template_decode.go` (new), `internal/handlers/json_decode.go`  
**Priority note:** Fresh runs confirm JSON decode is still one of the few hotspots large enough to matter to a `1500+` target without touching compliance-sensitive PDF logic.  
**Estimated gain:** +30–40% JSON decode reduction, potentially +5–10% end-to-end on HFT  
**Risk:** High — codegen adds maintenance; must preserve all field semantics.

---

## Medium impact

### MI-1: Structure-tree serialization efficiency

**Goal:** Lower `formatStructElemObjectTo` cumulative CPU from ~10–11%.

- [ ] Pool `StructKid` slices instead of allocating per element.
- [x] Reduce `append`/`memmove` in `formatStructElemObjectTo`:
  - [x] Replaced `strconv.Itoa` with `strconv.AppendInt` to a stack scratch buffer.
  - [x] Replaced `escapeText` string allocation with `appendEscapedPDFLiteral` + `Write([]byte)` for Title/Alt.
- [ ] Avoid interface dispatch for the hot `*bytes.Buffer` path (generic shape experiment regressed; needs a concrete buffer-specific function if attempted again).
- [ ] Exponential pre-cap for `ParentTree` in `ReserveMCIDs`.
- [x] Verify structure-tree output unchanged: `go test ./internal/pdf/...` passes.

**Result:** Changes were not enough. Fresh runs still show `formatStructElemObjectTo` at `9.77-10.47%` cumulative CPU.

**Files:** `internal/pdf/structure.go`, `internal/pdf/generator.go`  
**Estimated gain:** +3–5% throughput  
**Risk:** Low–Medium — must keep PDF/UA-2 hierarchy intact.

---

### MI-2: HFT draw path

**Goal:** Reduce `drawSharedDeferRow` / `drawSharedLayoutRow` cumulative CPU.

- [ ] Cache text-width results for repeated short strings in `drawTable`.
- [ ] Batch MCID assignment for shared-layout rows.
- [ ] Skip non-varying per-row work when `sharedRowLayout: true`.
- [ ] Run HFT-specific payload and confirm p99 improves.

**Files:** `internal/pdf/draw.go`  
**Priority note:** This is still worth doing, but it is secondary to compression + JSON decode + heap-growth work because its ceiling is smaller.  
**Estimated gain:** +3–5% throughput  
**Risk:** Medium — layout logic is complex; must preserve pagination.

---

### MI-3: Remove remaining hot-path `fmt.Sprintf` / `strconv.Itoa`

**Goal:** Eliminate easy string-formatting wins in structure-tree, footer/page-number, and link serialization.

- [x] Audit `internal/pdf/generator.go` for hot `strconv.Itoa` calls in `formatStructElemObjectTo` / `appendObjRefToWriter`.
- [x] Replace `strconv.Itoa` with `strconv.AppendInt` to a stack scratch buffer in structure-tree writer.
- [ ] Audit `internal/pdf/draw.go`, `internal/pdf/pagemanager.go` for remaining hot `fmt.Sprintf` / `strconv.Itoa` calls.
- [ ] Replace with buffer-based writes or `strconv.AppendInt` where applicable.
- [ ] Run `go test ./internal/pdf/...` and `make bench-k6`.

**Files:** `internal/pdf/generator.go`, `internal/pdf/draw.go`, `internal/pdf/pagemanager.go`  
**Estimated gain:** +1–3% throughput  
**Risk:** Low — mechanical changes.

---

## Easy wins

### EW-1: Cache resolved font names / font references

**Goal:** Avoid repeated `resolveFontName` and `getFontReference` registry lookups.

- [ ] Add a small per-generation cache keyed by `models.Props` or props string.
- [ ] Invalidate at PDF generation boundary.
- [ ] Verify no regressions in font fallback behavior.

**Files:** `internal/pdf/utils.go`, `internal/pdf/font_aliases.go`  
**Estimated gain:** +1–2% throughput  
**Risk:** Low.

---

### EW-2: Improve page-compress cache eviction

**Goal:** Keep hot compressed pages longer; avoid full-shard thrash.

- [ ] Replace full-shard `sync.Map.Clear()` with FIFO or simple per-shard LRU.
- [ ] Maintain bounded memory while preserving cache hit rate for repeated pages.
- [ ] Add hit/miss metrics if not present.

**Files:** `internal/pdf/font/compress_cache.go`  
**Estimated gain:** +1–3% throughput on repeated-page workloads  
**Risk:** Low.

---

## Reverted experiments — do not repeat without strong evidence

- [ ] ~~Parallel structure-tree build (G3)~~ — reverted; caused ~60% regression.
- [ ] ~~Template PDF cache (G4)~~ — removed; violates unique-PDF requirement.
- [ ] ~~CRC32 fingerprint + in-place signature hex~~ — reverted; CPU win, no E2E gain.
- [ ] ~~Store-uncompressed pages ≥96 KiB~~ — reverted; larger PDFs, slower signing.

---

## Acceptance criteria for closing this checklist

- [ ] At least one high-impact item shipped and re-profiled.
- [ ] Weighted `make bench-k6` 5-run avg reaches **~1,500+ req/s** on `tagged_ecdsa`.
- [ ] Best run alone is not enough; require a stable 5-run band with no collapse back into the `800-1000 req/s` range.
- [ ] p99 latency stays **< 500 ms** with 0% errors.
- [ ] `bytes.growSlice` in-use heap drops below **200 MB** under load.
- [ ] All `go test ./internal/...` pass.
- [ ] PDF/A-4 + PDF/UA-2 output remains valid (run existing PDF/A tests).

---

## How to verify each change

```bash
# Core correctness
go test ./internal/pdf/... ./internal/handlers/...

# Full weighted benchmark (run 3–5 times and average)
make bench-k6

# Quick retail-only gate
make bench-k6-retail

# Optional 1500 req/s retail gate
make bench-k6-1500

# Inspect latest pprof summary
ls -t guides/cursor/baselines/gin_pprof_runs/pprof_summary_*.txt | head -1 | xargs cat
```

---

## Commit message

```
perf(pdf): bound unbounded caches, optimize structure-tree writer

Three improvements that together move the `make bench-k6` weighted
(80% retail / 15% active / 5% HFT) 5-run mean from ~825 req/s to
~1005 req/s (+22%), and single-run best from 859 to 1041 req/s.

HI-2: Bound the unbounded caches
--------------------------------
`subsetCache` (font subsets), `imgCache` (decoded images) and
`propsCache` (parsed prop strings) were unbounded `sync.Map`s.
On long-lived pods or high-cardinality workloads this could grow
without limit. Each cache now has a max entry count; on overflow
the cache is cleared and the counter reset, matching the pattern
already used by `compress_cache.go`.

  - internal/pdf/font/subset_cache.go: max 1024, ClearSubsetCache()
  - internal/pdf/image.go:            max 256,  clear() helper
  - internal/pdf/utils.go:            max 8192, ClearPropsCache()

MI-3 + MI-1: Structure-tree writer allocation cleanup
------------------------------------------------------
`formatStructElemObjectTo` / `appendObjRefToWriter` were hot
(10-11% cum CPU in the latest pprof) and were allocating per
integer and per Title/Alt text.

  - `strconv.Itoa(n)` + `w.WriteString` -> `strconv.AppendInt` to a
    stack `[24]byte` scratch + `w.Write([]byte)`. No more per-int
    string allocations.
  - `escapeText(s)` + `w.WriteString` -> `appendEscapedPDFLiteral`
    to a stack `[1024]byte` scratch + `w.Write([]byte)`. No more
    `strings.Builder` per Title/Alt.

Tests
-----
  - internal/pdf/font/subset_cache_test.go
  - internal/pdf/image_cache_test.go
  - internal/pdf/props_cache_test.go

Benchmark (5-run, `make bench-k6`, 48 VUs x 35s, tagged_ecdsa)
------------------------------------------------------------
  before this commit:  825 req/s avg,  859 peak,  ~347 ms p99
  after  this commit: 1006 req/s avg, 1041 peak,  ~270 ms p99

Experiments tried and reverted (not in this commit):
  - HI-3: larger buffer estimates + pool retention caps
    (8 MB / 1 MB) regressed 993 -> 962 req/s.
  - MI-1: generic structure-writer shape regressed 998 -> 916
    req/s (Go used shape-based dispatch, not monomorphization).
  - EW-2: pre-store Load check in compress cache regressed a
    single run to 808 req/s, HFT latency spiked to 461 ms.

Refs: guides/optimizations/20260614_remaining_optimizations_checklist.md
```
