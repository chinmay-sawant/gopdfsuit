# Zerodha pprof → 5000 ops/sec plan (compliance-safe)

**Date:** 2026-07-09  
**Source:** `guides/cursor/baselines/zerodha_pprof_runs/` (2026-07-09)  
**Workload:** 80% Retail / 15% Active / 5% HFT · 5000 iters · 48 workers · PDF/A + embedded fonts + **RSA digital signature**  
**Hard constraint:** keep PDF/A, PDF/UA structure, Arlington-compatible output, and valid signatures — **no** “turn off compliance for speed”

---

## Baseline vs target

| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Mean throughput | **~2212 ops/s** (x10) | **5000 ops/s** | **~2.26×** |
| Best throughput | **~2716 ops/s** | 5000 | **~1.84×** |
| Mean avg latency | **~21.5 ms** | **~9.6 ms** @ 48 workers | **~55% latency cut** |
| Peak alloc | **~1.0–1.2 GB** | lower preferred | GC headroom |

**Rule of thumb at 48 workers:** `ops/s ≈ 48000 / avg_latency_ms` → need **~9.6 ms** avg.

Related artifacts:

- Timing (x10): `guides/cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt`
- CPU profiles: `guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run{1..5}.prof`
- Heap profile: `guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof`
- Run logs: `guides/cursor/baselines/zerodha_pprof_runs/zerodha_*.txt`

---

## Where time/memory goes (pprof)

### CPU (cum %, run1-style)

| Bucket | ~cum | What it is | Compliance note |
|--------|------|------------|-----------------|
| **RSA / PKCS#7 sign** | **~20–21%** | `SignPDF` → `addMulVVW1024` ~10–11% flat | **Must keep** valid RSA signature |
| **zlib/flate** | **~23%** | Page streams + font + ICC | Keep `/FlateDecode`; can change level/impl |
| **`drawTable`** | **~21%** | cells + MCID + widths + text | Keep structure; fix cost |
| **`func4` compress** | **~16%** | parallel page zlib close/write | Same streams, less waste |
| **`func6` struct tree** | **~8–9%** | `strings.Builder` + `AppendInt` + map | Keep same StructElem tree |
| **StructureManager** | **~6–8%** | `BeginMarkedContentBuf` / pool | Keep PDF/UA tags |
| **Font subset + compress** | **~8%** | `GenerateTrueTypeFontObjects` | Keep embedded/subset fonts |
| **memmove / memclr / grow** | **~12–15%** flat-ish | buffer growth + copies | Neutral |

### Heap (alloc_space)

| Hotspot | Share | Action class |
|---------|-------|--------------|
| `bytes.growSlice` | **~26%** | pre-grow / pool |
| `slices.Clone` (final PDF) | **~13%** | own buffer, drop double-copy |
| `strings.Builder` | **~11%** | write struct tree to `[]byte` / `bytes.Buffer` |
| `subsetGlyfAndLoca` / font build | **~12%+** | subset cache / reuse |
| `BeginMarkedContentBuf` | **~5%** | pool / fewer string ops |
| `flate.NewWriter` | **~6%** | better pool / reuse |
| Signature path | **~4–6%** | less ASN.1/PDF re-copy |

---

## Success gates (every PR)

- [ ] **Throughput:** Zerodha gold standard ≥ **5000 ops/s** mean over 10 runs (same machine class / `GOMAXPROCS`)
- [ ] **Compliance:** PDF/A validate (veraPDF) on retail/active/hft samples **pass**
- [ ] **Structure:** tagged PDF / PDF/UA structure tree still present when `PDFACompliant`
- [ ] **Signature:** signature validates (OpenSSL / pdfsig / existing tests)
- [ ] **Byte-size sanity:** retail ~61 KB, active ~76 KB, HFT ~2.4 MB ± small compress delta
- [ ] **No correctness regressions:** `go test ./internal/pdf/... ./pkg/gopdflib/...`
- [ ] **Re-profile after each phase** and attach top-20 flat/cum + heap tops

---

## Phase checklist

### P0 — Free / near-free copies & growth (~10–20% toward target)

Compliance impact: **none** (same PDF semantics).

- [ ] **P0.1 — Kill double `slices.Clone` on signed path**  
  `generator.go` ~1478–1490 clones unsigned then signed (~**13% alloc**).  
  Own the buffer once; only copy when signature mutates in place.  
  **Validate:** sign tests + heap `slices.Clone` drops sharply.

- [ ] **P0.2 — Drop compressed-page extra `make`+`copy`**  
  `func4` does `cp := make([]byte, n); copy(cp, …)` after zlib (~line 822).  
  Take ownership of pool buffer or append into pre-sized slice without second copy.  
  **Validate:** CPU `memmove` / alloc down; PDF streams identical.

- [ ] **P0.3 — Pre-grow content streams from HFT/row estimates**  
  Heap: `growSlice` via `AddNewPage` / `drawTable` (~**14% alloc** under grow).  
  Grow page buffers from template stats (rows × cols × ~bytes/cell), not 0→power-of-two thrash.  
  **Validate:** fewer `bytes.growSlice` samples; lower peak MB.

- [ ] **P0.4 — Pre-size main `pdfBuffer` + xref map**  
  `func6` burns time on `mapassign` + buffer growth.  
  Reserve `xrefOffsets` capacity = estimated object count; `pdfBuffer.Grow(estimatedSize)`.  
  **Validate:** lower `growslice` + map assign in func6 peek.

- [ ] **P0.5 — Cache ICC / OutputIntent bytes process-wide**  
  `GenerateOutputIntent` / gray ICC still shows up (~**2–3%** + flate).  
  Compress once, reuse immutable `[]byte` for all PDFs in process.  
  **Validate:** OutputIntent object identical; CPU ICC path near zero after warm-up.

**Exit:** mean ≥ **~2800–3200 ops/s** (stretch if clone+compress copy are big).

---

### P1 — Compression path (~15–25% of remaining CPU)

Compliance: keep **Flate** streams; do **not** switch to unsupported filters for PDF/A.

- [ ] **P1.1 — Use `flate.BestSpeed` (or level 1) for page content**  
  ~**23%** CPU in flate; Close dominates. PDF/A allows FlateDecode at any level.  
  **Validate:** size increase budgeted (e.g. retail &lt; +15%); throughput up.

- [ ] **P1.2 — Tighten zlib writer pool under 48 workers**  
  Heap still has `flate.NewWriter` / `newDeflateFast`.  
  Ensure `GetZlibWriter` always hits pool; reset properly; avoid NewWriter per page under load.  
  **Validate:** heap `NewWriter` flat ↓; no pool races.

- [ ] **P1.3 — Optional: `github.com/klauspost/compress/flate` drop-in**  
  Same API, often 1.5–2× encode. Keep output filter `/FlateDecode`.  
  **Validate:** veraPDF + golden smoke; measure size/CPU.

- [ ] **P1.4 — Font stream compress once per unique subset**  
  Font path shares flate cost with pages. Cache compressed subset stream by font+glyph-set hash.  
  **Validate:** identical glyphs → identical font objects; HFT/retail both faster.

- [ ] **P1.5 — Skip parallel compress overhead for 1-page docs**  
  Retail/active are small; errgroup spawn cost may not pay. Serial compress for ≤2 pages.  
  **Validate:** retail latency down without HFT regression.

**Exit:** mean ≥ **~3500–4000 ops/s** combined with P0.

---

### P2 — Structure tree write path (keep PDF/UA, cut ~8%+ CPU + ~11% alloc)

Compliance: **same** StructElem tree, MCIDs, roles — only serialization changes.

- [ ] **P2.1 — Rewrite `writeStructElems` with `[]byte` / `bytes.Buffer`, not `strings.Builder`**  
  `func6` list: `WriteString` + `AppendInt` dominate.  
  Append object lines into shared scratch; write once to `pdfBuffer`.  
  **Validate:** structure objects bit-identical (or canonical-equal); CPU func6 ↓.

- [ ] **P2.2 — Avoid per-element `strings.Builder` alloc**  
  Reuse one builder/buffer for whole tree walk.  
  **Validate:** alloc `strings.Builder` ↓.

- [ ] **P2.3 — Batch xref offset recording without hot map churn**  
  Pre-sized slice/map; avoid repeated `mapassign_fast64` in tight loop.  
  **Validate:** func6 map time ↓.

- [ ] **P2.4 — StructureManager: reduce pool miss cost**  
  `BeginMarkedContentBuf` ~**6%** cum; pool getSlow / popTail shows up.  
  Larger local freelist per `PageManager`; pre-allocate N struct elems from row×col estimate.  
  **Validate:** tagging still correct; pool CPU ↓; **do not** disable tagging for PDF/A path.

- [ ] **P2.5 — MCID / BDC string encoding without intermediate strings**  
  Append ` /BDC` operands into page buffer with scratch `[]byte`.  
  **Validate:** content stream operators unchanged.

**Exit:** structure still validates; mean +**300–600 ops/s**.

---

### P3 — `drawTable` hot path (~21% cum; keep cells + tags)

- [ ] **P3.1 — Cache `parseProps` / font resolve per distinct props string**  
  Props re-parsed per cell (`parseProps` in drawTable children).  
  **Validate:** same fonts/styles; CPU parse ↓.

- [ ] **P3.2 — Speed `appendFmtNum` / coordinate writes**  
  ~**2.4%** cum; fixed-point or dtoA into scratch without extra strconv layers.  
  **Validate:** visual positions same (or within 0.01 pt).

- [ ] **P3.3 — Width measurement cache**  
  `EstimateTextWidth` / `GetTextWidth` ~**2–3%**.  
  Cache (font, size, text) or measure once per unique cell text.  
  **Validate:** wrap layout golden tests.

- [ ] **P3.4 — `appendTextForPDF` zero-extra-alloc**  
  Already partially optimized; ensure wrap path never `string(line)` + re-escape.  
  **Validate:** escaping tests; heap under drawTable ↓.

- [ ] **P3.5 — Row-level structure, not over-tagging**  
  Keep Table/TR/TH/TD as required; avoid redundant nested structure if any exists.  
  **Validate:** PDF/UA checker; **no** stripping required tags.

- [ ] **P3.6 — Reuse row scratch slices (already started)**  
  Confirm `wrappedTextLines` / props slices never re-allocate every row beyond capacity.  
  **Validate:** allocs/op on table microbench ↓.

**Exit:** `drawTable` cum **&lt; 15%**; retail/active latency down materially.

---

### P4 — Signature path efficiency (**keep RSA PKCS#7 validity**)

~**20%** CPU is crypto — cannot delete for this workload (retail has `Signature.Enabled: true`).

- [ ] **P4.1 — Parse PEM keys once globally; reuse `*rsa.PrivateKey` + certs**  
  Ensure every PDF does not re-parse PEM.  
  **Validate:** signature still verifies; CPU outside `addMulVVW` drops.

- [ ] **P4.2 — Reuse `PDFSigner` / precomputed digests scaffolding**  
  Avoid re-building cert chain ASN.1 every request; cache DER of certs/chain.  
  **Validate:** PKCS#7 still correct; heap signature path ↓.

- [ ] **P4.3 — Minimize PDF byte copies in `UpdatePDFWithSignature`**  
  ~**330 MB alloc** in signature update. Sign in place with reserved ByteRange holes.  
  **Validate:** ByteRange / Contents length correct; pdfsig pass.

- [ ] **P4.4 — Confirm RSA key size is 2048 (not 4096)** for bench certs  
  4096 ≈ 4–8× slower. Use 2048 if policy allows (typical PDF signing).  
  **Validate:** still “compliance-grade” for product; document choice.

- [ ] **P4.5 — Do not move to “unsigned mode” for the 5000 target**  
  Target is **with** signing. Separate optional unsigned bench is fine for isolation only.

**Exit:** sign path still ~RSA-bound but **overhead around RSA &lt; 3–5%**; total sign cum closer to pure mul cost.

---

### P5 — Font subset caching (big on repeated charset)

- [ ] **P5.1 — Process-level cache: font file + used rune set → subset + compressed stream**  
  Zerodha docs share Latin/digit charset; re-subsetting every PDF wastes **~12% alloc** + flate.  
  **Validate:** ToUnicode/CID maps correct; no cross-request glyph leaks.

- [ ] **P5.2 — `MarkCharsUsed` cheaper set**  
  Bitset / roaring for BMP runes vs map churn in tables.  
  **Validate:** subset completeness tests.

- [ ] **P5.3 — Avoid re-cloning glyf/loca buffers**  
  `subsetGlyfAndLoca` ~**7% alloc**. Write into pooled buffers.  
  **Validate:** font checksum / render smoke.

**Exit:** second+ PDF of same template family much cheaper; multi-worker steady-state ↑.

---

### P6 — Concurrency / GC (system-level, still compliance-neutral)

- [ ] **P6.1 — Measure GOMAXPROCS sweet spot** (not always 48)  
  Over-subscription can inflate max latency (400–800 ms tails).  
  **Validate:** mean ops/s and p99 together.

- [ ] **P6.2 — Reduce alloc rate to cut GC** (`mallocgc` ~**7.5%** cum)  
  P0–P5 should cut this; re-check GC CPU after each phase.

- [ ] **P6.3 — Optional `GOMEMLIMIT` experiment**  
  Cap heap thrash under 48 workers; only if it helps throughput without OOM.

- [ ] **P6.4 — HFT 5% is a latency bomb** (~2.4 MB PDF)  
  Keep mix fixed for the gold standard; do not remove HFT to “hit” 5000.  
  Optimize HFT path (pages, compress, fonts) so it does not dominate tails.

---

## Suggested implementation order

| Order | Item | Est. gain | Risk to compliance |
|------:|------|-----------|--------------------|
| 1 | P0.1 double Clone | High alloc / some CPU | Low |
| 2 | P0.2 compress copy | Medium | Low |
| 3 | P0.3–P0.4 pre-grow | Medium | Low |
| 4 | P0.5 ICC cache | Small–medium | Low |
| 5 | P1.1 BestSpeed + P1.2 pool | **High** | Low (size ↑ only) |
| 6 | P2.1–P2.3 struct serialize | **High** | Low if golden-tested |
| 7 | P5.1 font subset cache | **High** steady-state | Medium (cache keys) |
| 8 | P4.1–P4.3 signer reuse | Medium (RSA remains) | Medium |
| 9 | P3 drawTable microopts | Medium | Low–medium |
| 10 | P1.3 klauspost | High optional | Low |
| 11 | P6 GOMAXPROCS / GC | Variable | None |

**Realistic stacking:** P0+P1+P2+P5 should approach **~1.7–2.1×**. P3+P4+P6 needed to clear **2.26×** to 5000 **with** RSA+PDF/A still on.

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

- Today: **~2.2k mean / ~2.7k best** ops/s; need **~5k** (**~2.3×** / **~1.8×**).  
- Time is split three ways: **RSA sign (~20%)**, **flate (~23%)**, **table+structure (~25–30%)**, with **huge alloc tax** (`growSlice` + `slices.Clone` + struct `strings.Builder` + font subset).  
- Hitting 5000 **without losing compliance** means **optimize around** PDF/A, tagging, and RSA — not disable them: cut copies, cache ICC/fonts, faster Flate, cheaper structure serialization, leaner drawTable, and tighter signer plumbing.

---

## Related docs

- [GOPDFLIB_PPROF_RESULTS.md](./GOPDFLIB_PPROF_RESULTS.md) — earlier 5000× GoPDFLib pprof
- [PASS4_OPTIMIZATION_PLAN.md](./PASS4_OPTIMIZATION_PLAN.md) — prior load-test hotspot plan
- [ZERODHA_BENCHMARK_RESULTS.md](./ZERODHA_BENCHMARK_RESULTS.md) — Zerodha benchmark notes
- [baselines/zerodha_bench_x10_wsl_stats_latest.txt](./baselines/zerodha_bench_x10_wsl_stats_latest.txt) — latest x10 stats
