# Zerodha pprof → 5000 ops/sec plan (compliance-safe)

**Date:** 2026-07-09 (last scan: 2026-07-10)  
**Scan status:** Branch `codehound-2061k-191` @ `dffc13e` — 10-run mean throughput **2,551.77 ops/s** (best run **2,838.25 ops/s**)  
**Source:** `guides/cursor/baselines/zerodha_pprof_runs/` (2026-07-09)  
**Workload:** 80% Retail / 15% Active / 5% HFT · 5000 iters · 48 workers · PDF/A + embedded fonts + **RSA digital signature**  
**Hard constraint:** keep PDF/A, PDF/UA structure, Arlington-compatible output, and valid signatures — **no** “turn off compliance for speed”

---

## Scan summary (2026-07-10)

| Phase | Fixed | Partial | Not fixed | Incorrect/regressed | Total |
|-------|-------|---------|-----------|---------------------|-------|
| P0 — Copies & growth | 1 | 1 | 2 | 1 | 5 |
| P1 — Compression | 1 | 2 | 2 | 0 | 5 |
| P2 — Structure tree | 0 | 3 | 2 | 0 | 5 |
| P3 — drawTable | 2 | 3 | 1 | 0 | 6 |
| P4 — Signature | 3 | 1 | 1 | 0 | 5 |
| P5 — Font subset | 0 | 1 | 2 | 0 | 3 |
| P6 — Concurrency/GC | 0 | 2 | 2 | 0 | 4 |
| **Total** | **7** | **13** | **12** | **1** | **33** |

**Throughput gap:** 2,551.77 / 5,000 ops/s; the target is still **1.96× away**. The 2,838.25 figure is the best run, not the mean. Most impactful remaining items: P0.2 (compress copy), P0.3 (pre-grow), P1.5 (serial compress small docs), P2.1–2.2 (struct tree `strings.Builder`), P4.3 (signature copy), P5.1 (font subset cache).

The checklist contains **33 actual items**, not 35. The fully complete count is **7**, but only five of those are substantive performance changes; P4.4 is a certificate confirmation and P4.5 is a guardrail confirmation.

---

---

## Success gates (every PR)

- [ ] **Throughput:** Zerodha gold standard ≥ **5000 ops/s** mean over 10 runs (same machine class / `GOMAXPROCS`)
- [ ] **Compliance:** PDF/A validate (veraPDF) on retail/active/hft samples **pass**
- [x] **Structure:** tagged PDF / PDF/UA structure tree still present when `PDFACompliant` (current retail/active/HFT PDF-UA checks pass)
- [ ] **Signature:** signature validates (OpenSSL / pdfsig / existing tests)
- [x] **Byte-size sanity:** retail ~61 KB, active ~76 KB, HFT ~2.4 MB ± small compress delta (current outputs are within range)
- [ ] **No correctness regressions:** `go test ./internal/pdf/... ./pkg/gopdflib/...`
- [ ] **Re-profile after each phase** and attach top-20 flat/cum + heap tops

**Current gate failures:** PDF/A-4 validation fails for all three current Zerodha outputs because the sRGB ICC stream is empty/invalid; signature validity is unverified; and the focused Go test gate panics in redaction (`internal/pdf/redact/search.go:337-343`). The compliance shell wrapper must not be treated as proof of success until its parallel wait/child-status handling is corrected.

---

## Phase checklist

### P0 — Free / near-free copies & growth

Compliance impact: **none** (same PDF semantics).

- [x] **P0.1 — Kill double `slices.Clone` on signed path**  
  `generator.go` ~1478–1490 — **FIXED** (commit `3accc2c`).  
  Signed path now passes `pdfBuffer.Bytes()` directly to `UpdatePDFWithSignature` which does its own `bytes.Clone` internally. No clone before signing; only one `slices.Clone` on the unsigned error path.  
  **Evidence:** `generator.go:1480-1493` — single clone on non-signed path only.

- [ ] **P0.2 — Drop compressed-page extra `make`+`copy`**  
  `generator.go:826` — still does `cp := slices.Clone(compressedBuf.Bytes())`.  
  Functionally identical to old `make`+`copy`. Pool buffer ownership still requires a copy.  
  **To fix:** Take ownership of pool buffer or append into pre-sized slice without second copy.  
  **Validate:** CPU `memmove` / alloc down; PDF streams identical.

- [ ] **P0.3 — Pre-grow content streams from HFT/row estimates**  
  `pagemanager.go:48,81` — fixed 64KB `Grow()` per page stream. No template-stats-based sizing.  
  `estimateInitialContentStreamCap` was removed in refactoring — does not exist on this branch.  
  **To fix:** Add `estimateInitialContentStreamCap(rows, cols)` and pass it to `Grow()` at page creation.  
  **Validate:** fewer `bytes.growSlice` samples; lower peak MB.

- [-] **P0.4 — Pre-size main `pdfBuffer` + xref map**  
  **Partial fix:** `xrefOffsets` pre-sized with estimate at `generator.go:126` (`make(map[int]int, 64+...)`).  
  Many builders pre-sized (fieldsRef, acroB, xobjBuilder, csBuilder, stdFontRefs, compressedBuf, ptBuilder).  
  **Not fixed:** `pdfBuffer` has no estimated `Grow()` — only the pool default 64KB.  
  **To fix:** Add `pdfBuffer.Grow(estimateFinalPDFSize(template))`.  
  **Validate:** lower `growslice` + map assign in func6 peek.

- [-] **P0.5 — Cache ICC / OutputIntent bytes process-wide**  
  **Incorrect claim / regression:** cache storage exists (`pdfa.go:347-355`), but `GetSRGBICCCompressed()` only returns `srgbICCCompressed` (`pdfa.go:377-380`) and no production caller initializes it through `GetSRGBICCProfile()`. Output-intent paths consume the empty slice directly (`metadata.go:312-315`, `pdfa.go:384-388`), causing PDF/A validation failures. Restore safe initialization before claiming this fixed.

**Exit:** mean ≥ **~2800–3200 ops/s** (stretch if clone+compress copy are big).

---

### P1 — Compression path (~15–25% of remaining CPU)

Compliance: keep **Flate** streams; do **not** switch to unsupported filters for PDF/A.

- [x] **P1.1 — Use `flate.BestSpeed` (or level 1) for page content**  
  **FIXED** — `font/compression.go:16` — `ZlibWriterPool` creates writers with `zlib.BestSpeed`.  
  All compression paths (pages, images, fonts, ICC, redaction, forms) use `BestSpeed` through pooled writers.  
  **Validate:** size increase budgeted (e.g. retail &lt; +15%); throughput up.

- [-] **P1.2 — Tighten zlib writer pool under 48 workers**  
  **Partial:** `font/compression.go:13-41` — pool exists with Get/Put + `Reset` before reuse. No `NewWriter` per page.  
  **Not fixed:** Parallel compression in `generator.go:800` uses bare `errgroup.Group` with **no concurrency limit** (`SetLimit`). All content streams compress concurrently without throttling.  
  **To fix:** Add `compGroup.SetLimit(48)` or use a bounded semaphore.  
  **Validate:** heap `NewWriter` flat ↓; no pool races.

- [ ] **P1.3 — Optional: `github.com/klauspost/compress/flate` drop-in**  
  **NOT FIXED** — `go.mod` does not list `klauspost/compress` as a dependency. Only `klauspost/cpuid/v2` is present (indirect). All flate/zlib uses standard library.  
  **Validate:** veraPDF + golden smoke; measure size/CPU.

- [-] **P1.4 — Font stream compress once per unique subset**  
  **Partial:** Within a single PDF generation, each font+subset is compressed exactly once (fonts deduped by registry).  
  **Not fixed:** No hash-based cross-generation cache (`fontCompressCache`, `glyphSetHash`). Each `GenerateTrueTypeFontObjects` call recompresses.  
  **To fix:** Add `sync.Map` keyed by (font name, glyph-set fingerprint) → compressed stream.  
  **Validate:** identical glyphs → identical font objects; HFT/retail both faster.

- [ ] **P1.5 — Skip parallel compress overhead for 1-page docs**  
  **NOT FIXED** — `generator.go:797-834` always uses `errgroup.Group` regardless of page count. No serial fallback for `nStreams <= 2`.  
  **To fix:** Check `len(pageManager.ContentStreams)` before spawning goroutines; serial path for ≤2 pages.  
  **Validate:** retail latency down without HFT regression.

**Exit:** mean ≥ **~3500–4000 ops/s** combined with P0.

---

### P2 — Structure tree write path (keep PDF/UA, cut ~8%+ CPU + ~11% alloc)

Compliance: **same** StructElem tree, MCIDs, roles — only serialization changes.

- [ ] **P2.1 — Rewrite `writeStructElems` with `[]byte` / `bytes.Buffer`, not `strings.Builder`**  
  **NOT FIXED** — `generator.go:1269` — `writeStructElems` still uses `var sb strings.Builder` per call.  
  Also `structure.go:339` — `GenerateStructTreeRoot` uses `strings.Builder`.  
  **To fix:** Rewrite both to use `[]byte` scratch buffer; flush to `pdfBuffer` once.  
  **Validate:** structure objects bit-identical (or canonical-equal); CPU func6 ↓.

- [ ] **P2.2 — Avoid per-element `strings.Builder` alloc**  
  **NOT FIXED** — `generator.go:1269` — a new `strings.Builder` is allocated on every recursive `writeStructElems` call.  
  **To fix:** Pass a shared `*bytes.Buffer` through the tree walk; no per-element builder allocation.  
  **Validate:** alloc `strings.Builder` ↓.

- [-] **P2.3 — Batch xref offset recording without hot map churn**  
  **Partial:** `generator.go:126` — `xrefOffsets` map is pre-sized with estimate.  
  **Not fixed:** Still uses individual `xrefOffsets[objID] = pdfBuffer.Len()` assignments (~30+ scattered sites). No batching.  
  **To fix:** Collect offsets in a pre-sized slice, batch-flush at end, or use `[]struct{objID, offset}` + sort.  
  **Validate:** func6 map time ↓.

- [-] **P2.4 — StructureManager: reduce pool miss cost**  
  **Partial:** `structure.go:119-136` — `sync.Pool` with `acquireStructElem` / `ReleaseStructElemsToPool` exists.  
  **Not fixed:** No local freelist per `PageManager`. No pre-allocation from row×col estimates.  
  **To fix:** Add per-SM ring buffer or slice of pre-allocated `StructElem` slots.  
  **Validate:** tagging still correct; pool CPU ↓; **do not** disable tagging for PDF/A path.

- [-] **P2.5 — MCID / BDC string encoding without intermediate strings**  
  **Partial:** `structure.go:238-289` — `BeginMarkedContentBuf` writes BDC directly to `*bytes.Buffer` (correct). Used in hot table path (`draw.go:1071`).  
  **Not fixed:** Some paths still use `BeginMarkedContent` with intermediate `strings.Builder` (e.g. `draw.go:282` — title headings, figure elements).  
  **To fix:** Migrate remaining `BeginMarkedContent` callers to `BeginMarkedContentBuf`.  
  **Validate:** content stream operators unchanged.

**Exit:** structure still validates; mean +**300–600 ops/s**.

---

### P3 — `drawTable` hot path (~21% cum; keep cells + tags)

- [-] **P3.1 — Cache `parseProps` / font resolve per distinct props string**  
  **Partial:** global `parseProps` caching exists (`utils.go:82-96`), but hot `drawTable` still resolves fonts per cell (`draw.go:1003-1008`). The separate resolved-font cache is limited to `drawTitleTable` (`draw.go:452-455,755-760`).
  **Validate:** same fonts/styles; CPU parse ↓.

- [-] **P3.2 — Speed `appendFmtNum` / coordinate writes**  
  **Partial:** `draw.go:77-105` — uses integer math (`int64(f*100 + 0.5)`) + `strconv.AppendInt`, avoiding expensive `strconv.AppendFloat`. Used at ~100 call sites.  
  **Residual:** Still calls `strconv.AppendInt` — a fully custom fixed-point buffer with manual digit writing would eliminate that too.  
  **Validate:** visual positions same (or within 0.01 pt).

- [ ] **P3.3 — Width measurement cache**  
  **NOT FIXED** — `utils.go:369-385` (`EstimateTextWidth`) and `font/registry.go:388-417` (`GetTextWidth`) both recompute per call.  
  No (font, size, text) → width look-up table exists. Custom fonts loop over every character.  
  **To fix:** Add LRU cache keyed by (resolvedName, text, font) or measure once per unique cell text.  
  **Validate:** wrap layout golden tests.

- [-] **P3.4 — `appendTextForPDF` zero-extra-alloc**  
  **Partial:** standard-font paths avoid an intermediate string (`utils.go:411-424`), but custom-font encoding still allocates `enc := make([]byte, ...)` on every call (`font/metrics.go:1110-1135`).
  **Validate:** escaping tests; heap under drawTable ↓.

- [x] **P3.5 — Row-level structure, not over-tagging**  
  **FIXED** — `draw.go:932,979,1071,1540-1545` — table/row/cell structure is emitted without redundant wrappers. Current code emits TD/MCID cells; the prior `TD/TH` wording overstated the actual TH behavior.  
  Table opens once per table (`StructTable`), row per row (`StructTR`), cell as marked content (`BeginMarkedContentBuf`).  
  No redundant nesting, no extra wrapper elements.  
  **Validate:** PDF/UA checker; **no** stripping required tags.

- [x] **P3.6 — Reuse row scratch slices (already started)**  
  **FIXED** — `draw.go:963-975` — all scratch buffers allocated once before row loop:  
  `cellWidthsForRow`, `wrappedTextLines`, `rowCellProps`, `rowResolvedFonts`, `scratchBuf`, `borderBuf`, `xobjBuf`, `colorBuf`, `placeholderBuf`, `checkboxBuf` + `WrapState`.  
  `wrappedTextLines[colIdx] = nil` resets per row; `buf[:0]` pattern throughout. No re-allocation per row.  
  **Validate:** allocs/op on table microbench ↓.

**Exit:** `drawTable` cum **&lt; 15%**; retail/active latency down materially. Current status: 2 fixed, 3 partial, 1 not fixed.

---

### P4 — Signature path efficiency (**keep RSA PKCS#7 validity**)

~**20%** CPU is crypto — cannot delete for this workload (retail has `Signature.Enabled: true`).

- [x] **P4.1 — Parse PEM keys once globally; reuse `*rsa.PrivateKey` + certs**  
  **FIXED** — `signature.go:56-68` — `signerPEMMaterialCache` is a global `sync.Map` with SHA256-based cache key.  
  `signature.go:125-134` — cache check at top of `parseSignerPEMMaterials`. Eviction at 64 entries.  
  **Validate:** signature still verifies; CPU outside `addMulVVW` drops.

- [-] **P4.2 — Reuse `PDFSigner` / precomputed digests scaffolding**  
  **Partial:** `signature.go:645-649` — cert chain uses `Raw` bytes (no re-marshaling).  
  **Not fixed:** Full PKCS#7 `signedData` (lines 539-686) is re-marshaled every `SignPDF` call. `authenticatedAttrs` includes fresh `signingTime`, preventing full precomputation.  
  **To fix:** Pre-marshal static portion of PKCS#7 SignedData; only re-marshal the time-varying `authenticatedAttrs`.  
  **Validate:** PKCS#7 still correct; heap signature path ↓.

- [ ] **P4.3 — Minimize PDF byte copies in `UpdatePDFWithSignature`**  
  **NOT FIXED** — `signature.go:812` still does `result := bytes.Clone(pdfData)` — full PDF copy (~330 MB).  
  Then two `copy()` calls for ByteRange and Contents replacement.  
  Generator side (`generator.go:1481`) avoids a *second* clone by passing raw bytes directly, but the signer's own clone remains.  
  **To fix:** Sign in place with reserved ByteRange holes or pass a writable buffer.  
  **Validate:** ByteRange / Contents length correct; pdfsig pass.

- [x] **P4.4 — Confirm RSA key size is 2048 (not 4096)** for bench certs  
  **FIXED** — `certs/leaf.pem` confirmed: `Public-Key: (2048 bit)`.  
  Code uses `rsa.SignPKCS1v15` / SHA256 (`signature.go:593-595`), consistent with 2048-bit key.  
  **Validate:** still “compliance-grade” for product; document choice.

- [x] **P4.5 — Do not move to “unsigned mode” for the 5000 target**  
  **FIXED** — `generator.go:426-444` — signer created only when `SignatureConfig.Enabled`.  
  `generator.go:1480-1489` — signing unconditional when signer exists. No unsigned-skip optimization.  
  The only unsigned path is the error fallback (line 1486).  
  **Validate:** no unsigned mode for perf; separate optional bench is fine.

**Exit:** sign path still ~RSA-bound but **overhead around RSA &lt; 3–5%**; total sign cum closer to pure mul cost. Current status: 3 fixed, 1 partial, 1 not fixed; signature validity remains unverified.

---

### P5 — Font subset caching (big on repeated charset)

- [ ] **P5.1 — Process-level cache: font file + used rune set → subset + compressed stream**  
  **NOT FIXED** — `font/registry.go:149-178` — `GenerateSubsets()` calls `SubsetTTF()` fresh on every invocation.  
  No `sync.Map` or cache keyed by (font identity, rune set fingerprint). ~**12% alloc** overhead.  
  **To fix:** Add global `sync.Map` with key = hash(font-name + sorted-glyphs) → (subset-data, glyf/loca).  
  **Validate:** ToUnicode/CID maps correct; no cross-request glyph leaks.

- [ ] **P5.2 — `MarkCharsUsed` cheaper set**  
  **NOT FIXED** — `font/registry.go:27` — `UsedChars map[rune]bool` (standard map).  
  `registry.go:130-146` — `MarkCharsUsed` does `for _, char := range text { font.UsedChars[char] = true }`.  
  Only optimization: pre-sized `make(map[rune]bool, 256)`. No bitset / roaring bitmap.  
  **To fix:** For BMP runes, replace with `[8192]uint64` bitset (~65 KB per font).  
  **Validate:** subset completeness tests.

- [-] **P5.3 — Avoid re-cloning glyf/loca buffers**  
  **Partial:** `font/subset.go:253` — `glyphScratch := make([]byte, len(glyfData))` — single reusable scratch per call.  
  `putU16BE`/`putU32BE` replace `binary.Write`. Map pre-sized.  
  **Not fixed:** No `sync.Pool` for scratch or loca slices. `newLoca` still fresh `make([]byte, ...)`. `newGlyf.Bytes()` returns fresh copy.  
  **To fix:** Add `sync.Pool` for glyf/loca buffers; reuse across calls.  
  **Validate:** font checksum / render smoke.

**Exit:** second+ PDF of same template family much cheaper; multi-worker steady-state ↑.

---

### P6 — Concurrency / GC (system-level, still compliance-neutral)

- [-] **P6.1 — Measure GOMAXPROCS sweet spot** (not always 48)  
  **Partial:** `makefile:105` — `GOMAXPROCS_BENCH ?= 24`. `makefile:120` — `K6_LIGHT_GOMAXPROCS ?= 12`.  
  `main.go:61` — `runtime.GOMAXPROCS(0)` used for semaphore size.  
  **Not done:** No systematic A/B sweep documented (12, 24, 48, 96). No harness that varies GOMAXPROCS programmatically.  
  **To do:** Run benchmark at GOMAXPROCS=12,24,48,96 and record throughput + p99 latency.  
  **Validate:** mean ops/s and p99 together.

- [-] **P6.2 — Reduce alloc rate to cut GC** (`mallocgc` ~**7.5%** cum)  
  **Indirect progress:** Multiple pools added across P0-P4 (pdfBufferPool, scratchBufPool, ZlibWriterPool, CompressBufPool, structElemPool, templatePDFPool). Pre-sized maps everywhere.  
  **Not done:** No targeted P6.2 profiling pass. GC CPU share not re-measured after optimizations.  
  **To do:** Re-profile with `-sample_index=alloc_space` after remaining phases; quantify GC CPU.  
  **Validate:** GC CPU % ↓ in pprof.

- [ ] **P6.3 — Optional `GOMEMLIMIT` experiment**  
  **NOT FIXED** — No `GOMEMLIMIT`, `GOGC`, or `debug.SetGCPercent()` anywhere in codebase or makefile.  
  **To do:** Add `GOMEMLIMIT=800MiB` or similar experiment in benchmark harness.  
  **Validate:** peak heap stabilises, throughput does not regress.

- [ ] **P6.4 — HFT 5% is a latency bomb** (~2.4 MB PDF)  
  **NOT FIXED** — HFT remains 5% of mix; no HFT-specific path optimizations.  
  Max latency (707ms) is always the HFT doc under contention.  
  **To do:** Profile HFT specifically; optimize its page pre-growing, compression, font handling.  
  **Validate:** HFT max latency ↓ without removing from mix.

---

## Suggested implementation order (updated 2026-07-10)

| Order | Item | Status | Est. gain | Risk to compliance |
|------:|------|--------|-----------|--------------------|
| 1 | P0.1 double Clone | **FIXED** | — | — |
| 2 | P0.2 compress copy | **NOT FIXED** | Medium | Low |
| 3 | P0.3–P0.4 pre-grow | P0.4 PARTIAL, P0.3 NOT | Medium | Low |
| 4 | P0.5 ICC cache | **REGRESSED / NOT SAFE** | — | **High until PDF/A passes** |
| 5 | P1.1 BestSpeed + P1.2 pool | P1.1 FIXED, P1.2 PARTIAL | **High** | Low |
| 6 | P2.1–P2.3 struct serialize | **NOT FIXED** (P2.3 PARTIAL) | **High** | Low if golden-tested |
| 7 | P5.1 font subset cache | **NOT FIXED** | **High** steady-state | Medium (cache keys) |
| 8 | P4.1–P4.3 signer reuse | P4.1 FIXED, P4.2 PARTIAL, P4.3 NOT | Medium (RSA remains) | Medium |
| 9 | P3 drawTable microopts | **2 FIXED / 3 PARTIAL / 1 NOT FIXED** | Medium | Low–medium |
| 10 | P1.3 klauspost | **NOT FIXED** | High optional | Low |
| 11 | P6 GOMAXPROCS / GC | Mostly NOT FIXED | Variable | None |

**Remaining high-impact items to reach 5000:**
1. **P4.3** — Sign in-place (kill `bytes.Clone` of ~330 MB PDF per sign)
2. **P0.2** — Eliminate compress-page `slices.Clone` copy
3. **P0.3** — Pre-grow content streams from template estimates
4. **P2.1–P2.2** — Rewrite struct tree serialization (drop `strings.Builder`)
5. **P1.5** — Serial compress for ≤2-page docs
6. **P5.1** — Font subset cache across PDF generations

**Realistic stacking:** P0+P1+P2+P5 should approach **~1.7–2.1×**, but the current PDF/A regression must be repaired first. P3+P4+P6 are still needed to clear the remaining gap to 5000 **with** RSA+PDF/A still on.

---

## What we will **not** do (compliance red lines)

- [ ] ❌ Disable `PDFACompliant` / OutputIntent / ICC to chase ops/s  
- [ ] ❌ Strip PDF/UA structure tree or per-cell MCIDs on the PDF/A path  
- [ ] ❌ Disable digital signatures on the Zerodha gold mix  
- [ ] ❌ Replace Flate with non-standard filters that break PDF/A  
- [ ] ❌ Drop HFT from the mix just to inflate throughput  
- [ ] ❌ Change visual layout/fonts without golden/PDF compare  

---

## Measurement recipe (repeat each phase)

```bash
# Timing baseline (10×) — existing zerodha x10 harness
# → guides/cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt

# CPU + heap (same as current artifacts)
go tool pprof -top -cum guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run1.prof
go tool pprof -top -sample_index=alloc_space guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof

# Focused peeks
go tool pprof -peek=drawTable guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run1.prof
go tool pprof -list='GenerateTemplatePDF.func4' guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run1.prof
go tool pprof -list='GenerateTemplatePDF.func6' guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run1.prof

# Compliance
# veraPDF on zerodha_*_output.pdf + signature verify + unit tests
go test ./internal/pdf/... ./pkg/gopdflib/...
```

**Promotion rule:** merge only if **ops/s ↑** and **compliance gates green**.

---

## Bottom line

- Today: **2,551.77 ops/s mean / 2,838.25 ops/s best**; need **5,000** (about **1.96×** the current mean).  
- Confirmed complete checklist items: **7/33**; substantive performance fixes: **5**.  
- PDF/A is currently failing because the cached sRGB ICC payload is not initialized; fix this before accepting throughput changes.  
- The focused correctness gate is also red due to the `strings.Builder` panic in redaction, and signature validity is not yet independently verified.  
- Hitting 5000 **without losing compliance** means **optimize around** PDF/A, tagging, and RSA — not disable them: cut copies, cache ICC/fonts, faster Flate, cheaper structure serialization, leaner drawTable, and tighter signer plumbing.

---

## Related docs

- [GOPDFLIB_PPROF_RESULTS.md](./GOPDFLIB_PPROF_RESULTS.md) — earlier 5000× GoPDFLib pprof
- [PASS4_OPTIMIZATION_PLAN.md](./PASS4_OPTIMIZATION_PLAN.md) — prior load-test hotspot plan
- [ZERODHA_BENCHMARK_RESULTS.md](./ZERODHA_BENCHMARK_RESULTS.md) — Zerodha benchmark notes
- [baselines/zerodha_bench_x10_wsl_stats_latest.txt](./baselines/zerodha_bench_x10_wsl_stats_latest.txt) — latest x10 stats

---

## Codehound detection coverage (2026-07-11)

All 724 codehound findings across 29 chunk files were cross-referenced against the 26 non-`[x]` plan items. **8 items have codehound coverage; 18 have zero detection.**

### Detected by codehound

- [-] **P0.5 — Cache ICC / OutputIntent bytes process-wide** (regressed)
  `internal/pdf/font/pdfa.go:380:16` — `m.GetLiberationFont(stdFont)` in loop
  ```go
  func (m *PDFAFontManager) RegisterLiberationFontsForPDFA(...) error {
      for _, stdFont := range usedStandardFonts {
          font, err := m.GetLiberationFont(stdFont) // PERF-230: pure function re-evaluated
  ```
  `internal/pdf/font/pdfa.go:385:13` — `registry.RegisterFont(stdFont, font)` in loop
  ```go
          if err := registry.RegisterFont(stdFont, font); err != nil { // PERF-230
  ```
  `internal/pdf/metadata.go:313:20` — `GetSRGBICCCompressed()` called per operation
  ```go
  func (h *PDFAHandler) GenerateOutputIntent(...) (int, []string, []byte) {
      compressedData := GetSRGBICCCompressed() // PERF-217: static computation rebuilt per op
  ```
  `internal/pdf/pdfa.go:381:2` — `GetSRGBICCProfile()` triggers `sync.Once` every call
  ```go
  func GetSRGBICCCompressed() []byte {
      GetSRGBICCProfile() // ensure sync.Once has run; PERF-217: rebuilt per op
      return srgbICCCompressed
  }
  ```

- [-] **P1.2 — Tighten zlib writer pool under 48 workers**
  `internal/pdf/font/compression.go:13:1` — `ZlibWriterPool` is a package-level global
  ```go
  var ZlibWriterPool = sync.Pool{ // BP-37: package-level mutable global
      New: func() any {
          w, _ := zlib.NewWriterLevel(io.Discard, zlib.BestSpeed) // BP-1: discarded error
  ```
  `internal/pdf/font/compression.go:23:1` — `CompressBufPool` is a package-level global
  ```go
  var CompressBufPool = sync.Pool{ // BP-37: package-level mutable global
  ```
  `internal/pdf/font/compression.go:59:2` — `PutCompressBuffer` has capacity guard but still flagged
  ```go
  func PutCompressBuffer(buf *bytes.Buffer) {
      if buf.Cap() > 131072 { return }
      CompressBufPool.Put(buf) // PERF-219: oversized object returned to pool
  }
  ```

- [ ] **P1.4 — Font stream compress once per unique subset**
  `internal/pdf/font/metrics.go:680:3` — `_ = zlibWriter.Close()` in `GenerateTrueTypeFontObjects`
  ```go
  func GenerateTrueTypeFontObjects(font *RegisteredFont, ...) map[int]string {
      if _, err := zlibWriter.Write(fontData); err != nil {
          _ = zlibWriter.Close() // BP-1: discarded error (compress path)
  ```
  `internal/pdf/font/metrics.go:685:2` — `_ = zlibWriter.Close()` same function
  ```go
      _ = zlibWriter.Close() // BP-1: discarded error
  ```
  `internal/pdf/font/metrics.go:975:3` — same pattern in `GenerateCIDToGIDMap`
  ```go
  func GenerateCIDToGIDMap(...) string {
      if _, err := zlibWriter.Write(mapData); err != nil {
          _ = zlibWriter.Close() // BP-1: discarded error
  ```
  `internal/pdf/font/metrics.go:980:2` — `_ = zlibWriter.Close()` same function
  ```go
      _ = zlibWriter.Close() // BP-1: discarded error
  ```
  `internal/pdf/font/metrics.go:1082:3` — same in `GenerateToUnicodeCMap`
  ```go
  func GenerateToUnicodeCMap(...) string {
      if _, err := zlibWriter.Write(byteconv.StringToBytes(cmapData)); err != nil {
          _ = zlibWriter.Close() // BP-1: discarded error
  ```
  `internal/pdf/font/metrics.go:1087:2` — `_ = zlibWriter.Close()` same function
  ```go
      _ = zlibWriter.Close() // BP-1: discarded error
  ```

- [-] **P2.4 — StructureManager: reduce pool miss cost**
  `internal/pdf/structure.go:119:1` — `structElemPool` is a package-level global
  ```go
  var structElemPool = sync.Pool{ // BP-37: package-level mutable global
      New: func() any {
  ```

- [-] **P3.1 — Cache `parseProps` / font resolve per distinct props string**
  `internal/pdf/draw.go:520:27` — `parseHexColor(bgColor)` in cell loop
  ```go
      for colIdx, cell := range row.Row {
          cc, ok := colorCache[bgColor]
          if !ok {
              r, g, b, _, valid := parseHexColor(bgColor) // PERF-230 in drawTitleTable
  ```
  `internal/pdf/draw.go:698:12` — `getFontReference(cellProps, ...)` in cell loop (content pass)
  ```go
          fr = getFontReference(cellProps, pageManager.FontRegistry) // PERF-230 in drawTitleTable
  ```
  `internal/pdf/draw.go:732:36` — `parseHexColor(textColor)` in text rendering loop
  ```go
          } else if r, g, b, _, valid := parseHexColor(textColor); valid { // PERF-230 in drawTitleTable
  ```
  `internal/pdf/draw.go:757:12` — `resolveFontName(cellProps, ...)` in text loop
  ```go
          rn = resolveFontName(cellProps, pageManager.FontRegistry) // PERF-230 in drawTitleTable
  ```
  `internal/pdf/draw.go:762:18` — `EstimateTextWidth(...)` per cell
  ```go
          textWidth := EstimateTextWidth(resolvedName, cell.Text, float64(cellProps.FontSize), pageManager.FontRegistry) // PERF-230
  ```

- [ ] **P3.3 — Width measurement cache**
  `internal/pdf/font/registry.go:399:26` — `getWidth(char)` per character in `GetTextWidth`
  ```go
  func (r *CustomFontRegistry) GetTextWidth(fontName string, text string) float64 {
      for _, char := range text {
          totalWidth += float64(getWidth(char)) * invUPE // PERF-230: pure function in loop
  ```
  `internal/pdf/font/registry.go:413:25` — same call in the locked code path
  ```go
      for _, char := range text {
          totalWidth += float64(getWidth(char)) * invUPE // PERF-230: pure function in loop
  ```

- [-] **P3.4 — `appendTextForPDF` zero-extra-alloc**
  `internal/pdf/font/metrics.go:1142:9` — `len(text)*4+2` allocation in `EncodeTextForCustomFont`
  ```go
  func EncodeTextForCustomFont(fontName string, text string, registry *CustomFontRegistry) string {
      buf := make([]byte, 0, len(text)*4+2) // BP-52: unchecked integer multiplication
  ```

- [-] **P6.2 — Reduce alloc rate to cut GC**
  `internal/pdf/encryption/encrypt.go:65:12` — `make+copy` in `padPassword`
  ```go
  func padPassword(password string) []byte {
      result := make([]byte, 32) // PERF-226: post-producer buffer re-copy
      if len(pwd) >= 32 {
          copy(result, pwd)
  ```
  `internal/pdf/font/metrics.go:1098:2` — `strings.Builder` without pre-sizing in `GenerateToUnicodeCMap`
  ```go
      var sb strings.Builder
      var lenTmp [20]byte
      sb.WriteString("<< /Filter /FlateDecode /Length ") // PERF-215: no Grow() call
  ```

### Not detected by codehound (no matching findings)

- [ ] **P0.2** — Drop compressed-page extra `make`+`copy` (`generator.go:826`) — no rule for redundant `slices.Clone`
- [ ] **P0.3** — Pre-grow content streams from HFT/row estimates (`pagemanager.go:48,81`) — no rule for missing `Grow()` with estimates
- [-] **P0.4** — Pre-size main `pdfBuffer` + xref map (`generator.go:126`) — no rule for missing buffer pre-sizing
- [-] **P1.2** — `errgroup.Group` without `SetLimit` (`generator.go:800`) — no rule for unbounded goroutine concurrency
- [ ] **P1.3** — Optional `klauspost/compress` drop-in — dependency check, not structural
- [ ] **P1.5** — Skip parallel compress overhead for 1-page docs (`generator.go:797-834`) — no rule for conditional parallelism
- [ ] **P2.1** — Rewrite `writeStructElems` with `[]byte` / `bytes.Buffer` (`generator.go:1269`) — no rule targeting `strings.Builder` in struct serialization
- [ ] **P2.2** — Avoid per-element `strings.Builder` alloc (`generator.go:1269`) — same, no per-call allocation pattern rule
- [-] **P2.3** — Batch xref offset recording (`generator.go:126`) — no rule for map-assign vs slice batch
- [-] **P2.5** — MCID / BDC string encoding without intermediate strings (`structure.go:238-289`, `draw.go:282`) — no rule for `BeginMarkedContent` vs `BeginMarkedContentBuf`
- [-] **P3.2** — Speed `appendFmtNum` / coordinate writes (`draw.go:77-105`) — no rule for `strconv.AppendInt` vs custom fixed-point
- [-] **P4.2** — Reuse `PDFSigner` / precomputed digests scaffolding (`signature.go:539-686`) — no rule for PKCS#7 pre-marshaling
- [ ] **P4.3** — Minimize PDF byte copies in `UpdatePDFWithSignature` (`signature.go:812`, `bytes.Clone`) — no rule for redundant signature-path `bytes.Clone`
- [ ] **P5.1** — Process-level cache: font file + used rune set → subset + compressed stream (`font/registry.go:149-178`) — no cross-generation cache detection
- [ ] **P5.2** — `MarkCharsUsed` cheaper set (`font/registry.go:27,130-146` — `map[rune]bool`) — no rule suggesting bitset for BMP runes
- [-] **P5.3** — Avoid re-cloning glyf/loca buffers (`font/subset.go:253`) — no `sync.Pool` miss detection for scratch buffers
- [-] **P6.1** — Measure GOMAXPROCS sweet spot — no rule for missing A/B sweep harness
- [ ] **P6.3** — Optional `GOMEMLIMIT` experiment — no rule checks for `GOMEMLIMIT`/`GOGC` presence
- [ ] **P6.4** — HFT 5% is a latency bomb — no HFT-specific profiling detection

**Summary:** Codehound detected **8 / 26** non-fixed plan items (30.8%). The **PERF-230** rule (pure function re-evaluated in loop) was the most productive, catching 9 of the 23 matched findings across `draw.go`, `pdfa.go`, and `registry.go`. The 18 misses are architectural/pattern-level issues that codehound's current rule set doesn't target — missing `Grow()`, `strings.Builder` vs `[]byte`, conditional parallelism, map vs bitset, cross-generation caching, and container-level configuration. Top-impact blind spots: **P0.2** (compress copy), **P2.1/P2.2** (struct tree `strings.Builder`), **P4.3** (signature `bytes.Clone`), and **P5.1** (font subset cache).
