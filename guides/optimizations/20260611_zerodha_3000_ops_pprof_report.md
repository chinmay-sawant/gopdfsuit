# Zerodha gopdflib — Path to 3000 ops/sec (pprof Analysis)

**Date:** 2026-06-11  
**Go Version:** `go1.26.4` (`/home/chinmay/go/bin/go1.26.4`)  
**Harness:** `sampledata/gopdflib/zerodha/main.go`  
**Profiler runner:** `sampledata/gopdflib/zerodha/run_bench_3000_pprof.sh`  
**Profiles:** `guides/cursor/baselines/zerodha_3000_pprof_runs/`

---

## Executive summary

| Milestone | Peak throughput | 5-run avg | Avg latency | Wall (5k docs) |
|-----------|----------------:|----------:|------------:|---------------:|
| **Pre-opt (RSA)** | 2,236 ops/s | ~1,776 ops/s | ~21 ms | ~2.24 s |
| **Phase F (ECDSA, full regen)** | 2,431 ops/s | 2,334 ops/s | ~19 ms | ~2.14 s |
| **Current (G2 only, no template cache)** | **2,751 ops/s** | **2,476 ops/s** | **~18.9 ms** | **~1.91 s** |

**Target (3,000 ops/s): not yet reached** — ~8% below peak at 48 workers / `GOMAXPROCS=24`. Every PDF is generated from scratch (no template cache).

**Verdict:** 3000 ops/sec is not reachable by micro-optimizing `drawTable` or `fmt.Sprintf` alone. Profiles show ECDSA signing, zlib, and buffer growth as dominant costs. Phases A–F + G2 (page compress cache) deliver incremental gains. **Template PDF cache (G4) was removed** — each request produces a separate PDF. **G3 (parallel struct tree) was reverted** — it spawned one goroutine per structure element and regressed HFT throughput by ~60%.

---

## How to reproduce

```bash
bash sampledata/gopdflib/zerodha/run_bench_3000_pprof.sh

# Inspect profiles
BIN=guides/cursor/baselines/zerodha_3000_pprof_runs/zerodha_bench
go tool pprof -http=:8083 $BIN guides/cursor/baselines/zerodha_3000_pprof_runs/cpu_zerodha_run3.prof
go tool pprof -http=:8084 $BIN guides/cursor/baselines/zerodha_3000_pprof_runs/heap_zerodha.prof
```

---

## Timing results (5 runs, go1.26.4)

| Run | Throughput (ops/s) | Avg latency (ms) | Wall time (s) | Peak mem (MB) |
|-----|-------------------|------------------|---------------|---------------|
| 1 | 1460.60 | 32.29 | 3.42 | 1160 |
| 2 | 1592.29 | 29.63 | 3.14 | 1205 |
| 3 | 1773.82 | 26.36 | 2.82 | 1169 |
| 4 | 1845.07 | 25.44 | 2.71 | 1091 |
| 5 | **2210.08** | **21.24** | **2.26** | 989 |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2210 ops/s** |
| **Worst** | 1461 ops/s |
| **Average** | 1776 ops/s |
| **σ (throughput)** | ~280 ops/s |

Run-to-run variance is high (~1460–2210) because the 5% HFT draw is random (211–293 HFT docs per run, each 50+ pages / 2.4 MB). Peak runs had fewer HFT docs.

---

## Theoretical ceiling (why 3000 is hard at 48 workers)

```
throughput ≈ workers ÷ avg_latency_seconds
```

| Avg latency | Theoretical max (48 workers) |
|------------:|-----------------------------:|
| 21 ms (peak run) | **2,286 ops/s** |
| 24 ms (this session avg) | 2,000 ops/s |
| 16 ms (**required**) | **3,000 ops/s** |
| 14 ms | 3,429 ops/s |

The observed **2,236 peak** is already within ~3% of the 21 ms theoretical ceiling. To hit 3000 at 48 workers, **average per-document latency must drop from ~21 ms to ~16 ms** — a 24% reduction across the weighted workload.

---

## CPU profile (5 runs, aggregated findings)

Analyzed `cpu_zerodha_run{1-5}.prof`. Representative cumulative tree from run 3 (2008 ops/s):

### Top hotspots by cumulative CPU

| Hotspot | Cumulative % | Flat % | Notes |
|---------|-------------|--------|-------|
| `gopdflib.GeneratePDF` → `GenerateTemplatePDFBorrowed` | **~79%** | — | E2E path |
| `compress/zlib.(*Writer).Close` → `flate` | **~22%** | ~5% encode | Per-page stream compression |
| `drawTable` | **~19%** | ~1.4% | Dominated by HFT (2000 rows) |
| `signature.UpdatePDFWithSignature` → `rsa.SignPKCS1v15` | **~17%** | **~10%** `bigmod.addMulVVW1024` | **Retail only (80%)** |
| `generateAllContentWithImages` | ~19% | — | Content layout |
| `GenerateTemplatePDFBorrowed.func7` | ~11% | ~4% | Final xref/stream assembly |
| `font.GenerateTrueTypeFontObjects` | ~8% | — | Liberation subset embedding |
| `runtime.memmove` | — | **~7%** | Buffer copies |
| `runtime.memclrNoHeapPointers` | — | **~3%** | Zeroing |
| `fmt.Sprintf` | ~0.2% | — | **Already negligible** |

### Key insight: signatures are the #1 flat CPU consumer

```
crypto/internal/fips140/bigmod.addMulVVW1024   10.35% flat (all 5 runs)
signature.(*PDFSigner).createPKCS7SignedData     16.50% cum
crypto/rsa.SignPKCS1v15                        15.76% cum
```

Retail template (`buildRetailTemplate`) enables `Signature.Enabled: true` with a **2048-bit RSA key** and 3-cert chain. Active (15%) and HFT (5%) templates do **not** sign. Because 80% of iterations sign, RSA math alone accounts for roughly **12–16% of total machine CPU** during the benchmark.

At ~21 ms avg latency for retail docs, PKCS#7 signing likely costs **~2–4 ms per retail PDF**. Removing or halving that cost gets you most of the way to 16 ms average.

### Compression is #2

Parallel zlib (`errgroup` per page in `generator.go:798`) is already in place, but `flate` still consumes ~22% cumulative:

- HFT docs: 50+ pages, megabytes of content streams
- `GenerateTemplatePDFBorrowed.func5` (parallel compress goroutine): ~15% cum
- Retail docs: compression is a smaller absolute cost but still allocates a new deflate context per page

### drawTable is #3 (HFT-weighted)

`drawTable` cum ~19%, but flat only ~1.4%. The cost is mostly in callees (structure tagging, text width, marked content). When a run draws 250+ HFT docs (2,000-row tables), wall time spikes. This explains the 1460–2210 ops/s variance.

---

## Heap profile (alloc_space, 5.22 GB total over 5000 docs)

| Hotspot | alloc_space | Notes |
|---------|------------|-------|
| `bytes.growSlice` | **25.0%** (1.30 GB) | PDF buffer + stream growth |
| `slices.Clone` | **14.8%** (0.77 GB) | Byte slice copies in assembly |
| `internal/bytealg.MakeNoZero` | **14.3%** (0.74 GB) | New buffer backing arrays |
| `signature.UpdatePDFWithSignature` | **5.5%** (0.29 GB) | PKCS#7 + signed byte ranges |
| `GenerateTemplatePDFBorrowed.func7` | **4.9%** (0.25 GB) | Final assembly |
| `BeginMarkedContentBuf` | **5.7%** (0.30 GB) | PDF/UA tagging |
| `compress/flate.NewWriter` | **6.2%** (0.32 GB) | Per-page deflate contexts |
| `strings.Builder.WriteString` | **3.7%** | Metadata / structure tree |

Peak in-use at end of run: **599 MB** (mostly `bytes.growSlice` retained capacity in pools).

---

## Workload breakdown (why averages mislead)

| Profile | Share | Pages | Signs? | Relative cost |
|---------|------:|------:|:------:|----------------|
| **Retail** | 80% | 1 | **Yes** | Low layout + **high crypto** |
| **Active** | 15% | 2–3 | No | Medium layout + compress |
| **HFT** | 5% | 50+ | No | **Very high** layout + compress |

A single HFT PDF can cost as much as 50–100 retail PDFs in CPU. Random HFT count per run (211–293 vs peak run's 211) shifts throughput by ±20%.

---

## Implementation checklist (path to 3000 ops/sec)

**Validation gate** (after each phase): `bash sampledata/gopdflib/zerodha/run_bench_3000_pprof.sh`  
**Target:** 5-run best ≥ **3000 ops/s**, 5-run average ≥ **2500 ops/s**

### Phase A — Digital signature fast path (est. +12–18%)

- [x] **A1** Cache parsed PEM materials (`*rsa.PrivateKey`, cert chain) — `signerPEMMaterialCache` in `signature.go`
- [x] **A2** Cache `*PDFSigner` instances per PEM fingerprint (`pdfSignerCache`)
- [x] **A3** Precompute static PKCS#7 fields (`certBytes`, `issuerAndSerial`) on signer init
- [x] **A4** In-place signature embedding — `UpdatePDFWithSignatureBuffer` avoids full-PDF copy
- [x] **A5** Pooled SHA-256 hasher + `digestByteRanges` fast path in `SignPDF`
- [x] **A6** Signature concurrency slots (`signWorkerSlots`, 2× NumCPU) + pooled auth attr buffers
- [x] **A7** ECDSA P-256 support — retail default; `BENCH_SIGN_RSA=1` for RSA baseline

**Files:** `internal/pdf/signature/signature.go`, `internal/pdf/generator.go`

### Phase B — Concurrency tuning (est. +8–15%)

- [x] **B1** Default `BENCH_WORKERS=72` in 3000-target harness (oversubscribe past 24 cores)
- [x] **B2** Pin `GOMAXPROCS=24` in benchmark runner
- [x] **B3** Exact 80/15/5 workload schedule with seeded shuffle (reproducible mix)

**Files:** `sampledata/gopdflib/zerodha/main.go`, `run_bench_3000_pprof.sh`

### Phase C — Compression optimization (est. +5–8%)

- [x] **C0** `flate.Writer` pool with `Reset()` — already in `internal/pdf/font/compression.go`
- [x] **C1** Store-uncompressed threshold — skip FlateDecode when compressed ≥ raw
- [x] **C2** Pre-size `compressedBuf.Grow()` from page stream length (reduce `growSlice`)
- [x] **C3** HFT shared row layout (`SharedRowLayout` + `SharedRowTemplateRow`) + page-init stream cache

**Files:** `internal/pdf/generator.go`, `internal/pdf/font/compression.go`

### Phase D — Allocation / assembly (est. +3–5%)

- [x] **D1** Improve `estimateFinalPDFSize` using content-stream byte totals
- [x] **D2** Benchmark hot path uses `GeneratePDFBorrowed` + `Release` (no clone)
- [x] **D3** `sync.Pool` shell for `fontRegistry.CloneForGeneration()`
- [x] **D4** Structure tree: batch MCID allocation (`ReserveMCIDs` + `BeginMarkedContentBufWithMCID`)

**Files:** `internal/pdf/generator.go`, `internal/pdf/font/registry.go`

### Phase E — Benchmark harness (consistency)

- [x] **E1** `BENCH_SEED` for fixed workload distribution (reproducible HFT count)
- [x] **E2** `BENCH_SKIP_WRITE=1` to skip sample PDF disk I/O in timed runs
- [x] **E3** Per-worker latency stats (no 5000-slot results channel)

**Files:** `sampledata/gopdflib/zerodha/main.go`

### Implementation order (active)

```
1. A2 → A3 → A4   (signature cache + in-place embed)
2. C1 → C2        (compression threshold + pre-grow)
3. D1             (buffer pre-sizing)
4. B1 → E1 → E2   (harness tuning)
5. A5 → A6        (incremental hash + sign pool, if still below 3000)
6. A7             (ECDSA, optional)
```

---

## Combined scenario analysis

| Scenario | Estimated throughput | Meets 3000? |
|----------|---------------------:|:-----------:|
| Current peak | 2,236 ops/s | No |
| Phase A only (signing) | ~2,550 ops/s | No |
| Phase A + B (sign + workers) | ~2,850 ops/s | Close |
| **Phase A + B + C** | **~3,050 ops/s** | **Yes** |
| Retail-only, no signing (theoretical) | ~3,500+ ops/s | Yes (not production-representative) |
| 96 workers + Phase A | ~3,100 ops/s | Yes (higher memory) |

**Most realistic path to 3000:** Phase A (signature) + Phase B (72–96 workers) + Phase C (compression pooling).

---

## What will NOT get you to 3000

| Already optimized / low impact | Profile evidence |
|-------------------------------|------------------|
| `fmt.Sprintf` removal | 0.2% cum — done |
| Structure tree / `BeginMarkedContentBuf` | 5.7% alloc, ~3% CPU — helpful but not sufficient alone |
| `drawTable` micro-opts on retail | Retail has ~10 rows; HFT variance dominates |
| More PDF/A metadata tweaks | Flat cost < 1% |

---

---

## Artifacts

| File | Contents |
|------|----------|
| `run_bench_3000_pprof.sh` | 5 timing + 5 CPU + 1 heap profile runner |
| `zerodha_3000_pprof_runs/zerodha_run{1-5}.txt` | Timing run logs |
| `zerodha_3000_pprof_runs/cpu_zerodha_run{1-5}.prof` | CPU profiles |
| `zerodha_3000_pprof_runs/heap_zerodha.prof` | Heap profile |
| `zerodha_3000_pprof_runs/zerodha_bench` | Built benchmark binary |

---

## Bottom line

The engine is **already within ~3% of its 48-worker theoretical ceiling** at peak. The gap to 3000 ops/sec is not a mystery — it's:

1. **~2–4 ms of RSA signing per retail doc** (fixable)
2. **~22% CPU in zlib** (reducible via pooling + store-uncompressed)
3. **HFT variance** pulling run averages down (fixable in harness)

Implementing signature caching + incremental hashing + worker oversubscription + flate pooling should cross **3000 ops/sec** without sacrificing PDF/A compliance or the 80/15/5 workload mix.

---

## Post-implementation results (2026-06-11, all phases except A7/D4)

**48 workers**, `BENCH_SEED=42`, `BENCH_SKIP_WRITE=1`, exact 80/15/5 schedule:

| Run | Throughput | Avg latency | Wall time |
|-----|----------:|------------:|----------:|
| 1 | 2363 ops/s | 19.9 ms | 2.12 s |
| 2 | **2439 ops/s** | **19.3 ms** | **2.05 s** |
| 3 | 2348 ops/s | 20.0 ms | 2.13 s |
| 4 | 2322 ops/s | 20.3 ms | 2.15 s |
| 5 | 2407 ops/s | 19.5 ms | 2.08 s |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2439 ops/s** (+9% vs 2236 pre-opt peak) |
| **5-run avg** | **2376 ops/s** |
| **Avg latency** | **19.8 ms** |

**Still below 3000** (~24% gap). Next: **D4** (structure tree batch MCID), **C3** (HFT shared page templates).

---

## Post-A7 results (ECDSA P-256, 2026-06-11)

**48 workers**, `BENCH_SEED=42`, `BENCH_SKIP_WRITE=1`, retail signing **ECDSA P-256** (default):

| Run | Throughput | Avg latency | Wall time |
|-----|----------:|------------:|----------:|
| 1 | 2601 ops/s | 18.0 ms | 1.92 s |
| 2 | 2374 ops/s | 19.7 ms | 2.11 s |
| 3 | 2404 ops/s | 19.4 ms | 2.08 s |
| 4 | **2701 ops/s** | **17.4 ms** | **1.85 s** |
| 5 | 2523 ops/s | 18.6 ms | 1.98 s |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2701 ops/s** (+11% vs 2439 RSA peak) |
| **5-run avg** | **2520 ops/s** |
| **Avg latency (best run)** | **17.4 ms** |

RSA comparison (`BENCH_SIGN_RSA=1`, same harness): **2172 ops/s** (single run).

**Still below 3000** (~10% gap at peak). **D4** + **C3** implemented (see below); further gains likely need compression pooling tuning or harness stabilization.

---

## Post-D4/C3 results (2026-06-11)

**48 workers**, ECDSA P-256, `BENCH_SEED=42`, HFT table uses `SharedRowLayout`:

| Run | Throughput | Avg latency | Wall time |
|-----|----------:|------------:|----------:|
| 1 | 2318 ops/s | 20.4 ms | 2.16 s |
| 2 | 2250 ops/s | 20.9 ms | 2.22 s |
| 3 | 2290 ops/s | 20.5 ms | 2.18 s |
| 4 | 2255 ops/s | 20.8 ms | 2.22 s |
| 5 | 2293 ops/s | 20.3 ms | 2.18 s |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2318 ops/s** |
| **5-run avg** | **2281 ops/s** |

Run-to-run variance remains high (HFT count + system load). Peak **2701** (post-A7) and **2318** (post-D4/C3) are within the observed σ band — re-profile with `run_bench_3000_pprof.sh` to confirm D4/C3 flat CPU impact.

---

## Post-pprof re-run (2026-06-11, `run_bench_3000_pprof.sh`)

**Note:** First pprof batch ran with stale shell `BENCH_SIGN_RSA=1` (RSA-2048). Script now `unset BENCH_SIGN_RSA` for ECDSA default.

### Timing (5 runs, RSA accidental)

| Run | Throughput | Avg latency | Peak mem |
|-----|----------:|------------:|---------:|
| 1 | 2281 ops/s | 20.6 ms | 1116 MB |
| 2 | 2303 ops/s | 20.4 ms | 1122 MB |
| 3 | 2311 ops/s | 20.3 ms | 1140 MB |
| 4 | 2256 ops/s | 20.8 ms | 1186 MB |
| 5 | **2324 ops/s** | **20.1 ms** | 1114 MB |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2324 ops/s** |
| **5-run avg** | **2295 ops/s** |

Workload fixed: 4000 retail / 750 active / 250 HFT (`BENCH_SEED=42`).

### CPU profile (5 runs, aggregated)

| Hotspot | Flat % | Cum % | Notes |
|---------|-------:|------:|-------|
| `bigmod.addMulVVW1024` (RSA) | **~9.5%** | — | #1 flat — retail 80% signed |
| `compress/flate.(*deflateFast).encode` | **~5.9%** | **~23%** zlib close | Per-page parallel compress |
| `GenerateTemplatePDFBorrowed.func7` | **~5.3%** | **~12%** | Final xref/stream assembly |
| `runtime.memmove` | **~5.9%** | — | Buffer copies |
| `drawTable` | **~1.7%** | **~19%** | HFT 2000-row tables |
| `signature.SignPDF` → RSA | — | **~16%** | PKCS#7 + RSA math |
| `BeginMarkedContentBuf` / structure | — | **~3%** CPU | D4 batch MCID — CPU OK |

### Heap profile (alloc_space, RSA run — pre-F1)

| Hotspot | alloc_space | Notes |
|---------|------------:|-------|
| `bytes.growSlice` | **31%** (1.31 GB) | PDF buffer growth |
| `bytealg.MakeNoZero` | **16%** (0.68 GB) | New backing arrays |
| **`ReserveMCIDs` ParentTree pre-grow** | **7.8%** (328 MB) | **D4 regression — fixed in F1** |
| `beginMarkedContentBuf` | 5.5% (155 MB) | Structure leaf elems |
| `compress/flate.NewWriter` | 3.2% (135 MB) | Pool cold / high concurrency |
| `GenerateTemplatePDFBorrowed.func7` | 5.9% (247 MB) | Final assembly |

### Post-F1 + C4 validation (ECDSA, single heap run)

| Metric | Value |
|--------|------:|
| Throughput | **2617 ops/s** |
| `ReserveMCIDs` in heap top-15 | **gone** (was 328 MB) |
| `compress/flate.NewWriter` | 117 MB (down from 135 MB) |

ECDSA 5-run spot-check after F1+C4: **2056–2596 ops/s**, best **2596**, avg **2318**.

---

## Phase F — Post-pprof fixes & remaining path to 3000

**Validation gate:** `bash sampledata/gopdflib/zerodha/run_bench_3000_pprof.sh` (ECDSA default)  
**Target:** 5-run best ≥ **3000 ops/s**, 5-run average ≥ **2600 ops/s**

### Phase F — Regression fixes & harness (done)

- [x] **F1** Remove `ReserveMCIDs` ParentTree pre-grow (was copying slice every row → 328 MB alloc)
- [x] **C4** Cap parallel page compression to `NumCPU` (`pageCompressSlots` in `generator.go`)
- [x] **E4** Pprof runner: `unset BENCH_SIGN_RSA`, build from `zerodha/` dir, ECDSA default

**Files:** `internal/pdf/structure.go`, `internal/pdf/generator.go`, `run_bench_3000_pprof.sh`

### Phase F — Next implementation (ordered)

- [x] **F2** Final assembly: pessimistic `estimateFinalPDFSize` (store-uncompressed upper bound + structure tree slack)
- [x] **F3** Structure tree: `StructKid` slice pool + exponential `ParentTree` pre-cap in `ReserveMCIDs`
- [x] **C5** Compression: `CompressContentStream` early-abort when first 4 KB does not shrink
- [x] **C6** Font subsetting: global glyph-set subset cache (`subset_cache.go`) for repeated Liberation workloads
- [x] **F4** `drawTable` text path: dedicated reusable `textTjBuf` for `appendTextForPDF` / `Tj` emission
- [x] **E5** Benchmark harness: `GOMAXPROCS` + signing algo in summary; `BENCH_WARMUP=0` skips warm-up

**Files:** `internal/pdf/generator.go`, `internal/pdf/structure.go`, `internal/pdf/draw.go`, `internal/pdf/font/`, `sampledata/gopdflib/zerodha/main.go`

### Implementation order (active)

```
1. F1 → C4 → E4 → F2   (done — pprof-driven fixes + buffer estimate)
2. C5 → F3 → F4 → C6 → E5   (done — full Phase F implementation)
3. (optional) Re-profile; tune C4 slot count or worker count if still below 3000
```

### Revised gap analysis (post-pprof, ECDSA)

| Cost center | Cum CPU | Est. savings | Item |
|-------------|--------:|-------------:|------|
| RSA → ECDSA (retail 80%) | ~16% | **~12–14%** throughput | A7 (use ECDSA in all timed runs) |
| zlib / flate | ~23% | **~5–8%** | C4 (done), C5 |
| `growSlice` / func7 assembly | ~12% + 32% alloc | **~4–6%** | F2 |
| Structure tree alloc | ~8% alloc | **~2–3%** | F1 (done), F3 |
| drawTable layout | ~19% cum | **~3–5%** | C3 (done), F4 |

**Verdict:** With **ECDSA + F1 + C4**, peak should land **2600–2800 ops/s**. Crossing **3000** requires **F2 + C5** (assembly + compression) without dropping the 80/15/5 mix or PDF/A tagging.

### Artifacts (this pprof batch)

| File | Contents |
|------|----------|
| `guides/cursor/baselines/zerodha_3000_pprof_runs/zerodha_run{1-5}.txt` | Timing logs (RSA run) |
| `guides/cursor/baselines/zerodha_3000_pprof_runs/cpu_zerodha_run{1-5}.prof` | CPU profiles |
| `guides/cursor/baselines/zerodha_3000_pprof_runs/heap_zerodha.prof` | Heap profile (RSA, pre-F1) |
| `guides/cursor/baselines/zerodha_3000_pprof_runs/zerodha_bench` | Built binary |

---

## Post-Phase-F complete (2026-06-11)

All checklist items **F3, C5, C6, F4, E5** implemented after pprof re-run.

### Timing (5 runs, ECDSA P-256, `BENCH_SEED=42`, Phase F complete)

| Run | Throughput | Avg latency |
|-----|----------:|------------:|
| 1 | 2195 ops/s | 21.2 ms |
| 2 | 2407 ops/s | 19.5 ms |
| 3 | 2276 ops/s | 20.4 ms |
| 4 | 2359 ops/s | 19.8 ms |
| 5 | **2431 ops/s** | **19.3 ms** |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2431 ops/s** |
| **5-run avg** | **2334 ops/s** |

### Status vs 3000 target

Still **~19% below** peak target at 48 workers / GOMAXPROCS=24. Dominant CPU remains **zlib (~23% cum)** and **drawTable (~19% cum)** on HFT; signing is **ECDSA** (~3–5% cum vs ~16% RSA).

### Optional next steps (not in original checklist)

- [x] **G1** Worker sweep: 48 vs 64 vs 72 at `GOMAXPROCS=24` — **48 workers wins**
- [x] **G2** Page-stream fingerprint cache (`CompressContentStreamCached` in `font/compress_cache.go`)
- [x] ~~**G3** Parallel structure-tree payload build~~ **Reverted** — one goroutine per struct elem (~10k+ on HFT) caused ~60% regression
- [x] ~~**G4** Template PDF cache~~ **Removed** — each request must produce a separate PDF; cache gave misleading 10k ops/s benchmark

---

## Current state (2026-06-11, post G4 removal)

**Template PDF cache (G4) removed** from `models.Config`, `generator.go`, and Zerodha harness. Every iteration — Retail, Active, and HFT — runs the full generate path.

**G2 (page compress cache) retained** — reuses zlib output for identical page content streams within a process. Does not skip PDF generation.

**G3 reverted** to sequential structure-tree write after measuring ~970 ops/s (vs ~2,500) when parallel goroutines were spawned per structure element on 2,000-row HFT tables.

### Benchmark results — full regeneration (ECDSA P-256, 48 workers, `GOMAXPROCS=24`, `BENCH_SEED=42`, warmup on)

| Run | Throughput (ops/s) | Avg latency (ms) | Wall time (s) |
|-----|-------------------:|-----------------:|--------------:|
| 1 | **2,750.53** | 16.91 | 1.818 |
| 2 | 2,496.30 | 18.54 | 2.003 |
| 3 | 2,326.50 | 19.92 | 2.149 |
| 4 | 2,196.01 | 21.32 | 2.277 |
| 5 | 2,612.62 | 17.65 | 1.914 |

| Aggregate | Value |
|-----------|------:|
| **Best** | **2,751 ops/s** |
| **5-run avg** | **2,476 ops/s** |
| **σ (throughput)** | ~220 ops/s |

Reproduce:

```bash
unset BENCH_SIGN_RSA
export BENCH_ITERATIONS=5000 BENCH_WORKERS=48 BENCH_SKIP_WRITE=1 BENCH_SEED=42 GOMAXPROCS=24
cd sampledata/gopdflib/zerodha && go1.26.4 run .
```

### Benchmark comparison — all phases

| Phase | Signing | Full regen | Peak (ops/s) | 5-run avg | Avg latency |
|-------|---------|:----------:|-------------:|----------:|------------:|
| Pre-opt | RSA-2048 | Yes | 2,236 | ~1,776 | ~21 ms |
| Post-A7 | ECDSA P-256 | Yes | 2,701 | — | ~17 ms |
| Phase F | ECDSA P-256 | Yes | 2,431 | 2,334 | ~19 ms |
| **Current** | ECDSA P-256 | **Yes** | **2,751** | **2,476** | **~18.9 ms** |

### Status vs 3000 target

**~8% below peak target.** Dominant costs: **HFT drawTable + zlib** (~40% of CPU from 5% of jobs), **retail ECDSA signing** (80% of jobs).