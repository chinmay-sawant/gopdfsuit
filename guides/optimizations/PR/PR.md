# Performance Optimization — gopdfsuit v6 (Phases 1–25)

Profile-guided performance work across the PDF engine, HTTP handlers, Python bindings, signing, benchmarks, and validation. Driven by SlopGuard static analysis, pprof-guided hot-path optimization, and Zerodha gold-standard compliance gates.

**Branch:** `feat/optimization-5.5-medium`  
**Workload:** Zerodha mix — 80% retail / 15% active / 5% HFT  
**Go:** 1.26.4 · **Module:** `gopdfsuit/v6`

---

## Summary

This PR delivers a sustained throughput improvement across three surfaces:

1. **Native Go (Zerodha in-process)** — **9,594 ops/s** mean on compliant TR→TD output (+243% vs pre-opt baseline), exceeding the **8,000 ops/s** gate with veraPDF **6/6 PASS**.
2. **Gin HTTP (k6 weighted)** — **1,223 req/s** post-regression recovery (+81% vs master 674 req/s); peak historical **1,232 req/s** (≥1,000 gate met).
3. **PyPDFSuit Python binding** — **836 ops/s** honest full-execution path (+278% vs 221 ops/s pre-pass); bottleneck shifted from Python serialization to FFI/render.

Cross-cutting: **218/218** SlopGuard PERF findings remediated, GoPDFKit parity **7/7 workloads**, veraPDF PDF/A-4 + PDF/UA-2 wired into `make test`.

---

## Impact at a Glance

| Surface | Baseline | Best Observed | Target | Status |
|---------|----------|---------------|--------|--------|
| Zerodha in-process (compliant TR→TD) | 573 ops/s (Go 1.24) | **9,594 ops/s** | 8,000 ops/s | **Exceeded** (+19.9%) |
| Zerodha in-process (3,000 ops/s push) | 573 ops/s | 2,751 ops/s peak | 3,000 | Not met (superseded by 8K/15K work) |
| Gin HTTP weighted (80/15/5) | ~674–825 req/s | **1,232 req/s** peak | 1,500 req/s | ~82% of target |
| Gin HTTP retail-only | — | **3,965 req/s** | 1,500 req/s | **Exceeded** |
| PyPDFSuit honest full path | 221 ops/s | **836 ops/s** | — | ~7% of native Go (11,721) |
| GoPDFKit comparison | 5/7 wins | **7/7 wins** | Parity | **Met** |
| `BenchmarkGoPdfSuit` (internal) | baseline | **−11.6%** ns/op, **−16%** B/op | — | **Met** |
| Zerodha 15,000 ops/s (next) | 9,009 idle | 9,594 peak session | 15,000 | ~64% — Phase A landed, gates pending |

---

## What Changed

### PDF Engine — Core (`internal/pdf/`)

**Buffer & compression (Phases 1–2)**
- Increased pooled final PDF buffer capacity; removed extra scratch-slice hop at final assembly
- Streamed page content directly into zlib (no `contentStream.Bytes()` intermediate)
- Replaced `append([]byte(nil), …)` with `slices.Clone`; pre-sized page content streams from template complexity
- Pooled `compress/flate.Writer` per worker; sharded page compress cache

**Structure tree & tagged PDF (Phases 2–4, P0–P25)**
- `StructureManager` page-index maps → slices; `BeginStructureElementCap` for known table/row child counts
- Stack-backed `[1024]byte` / `[128]byte` + `appendDecimal` direct writes; eliminated `appendObjRefToWriter`
- Iterative `assignStructIDs` and `writeStructElems` loops (replaced recursive walks)
- `xrefOffsets` map → pre-sized `[]int` slice with sentinel slots
- Lazy per-P struct-element arena (`arenaActivationThreshold=512`); batch arena TD allocation
- `tdLeafFast` flag; `beginTableRowArena` with inlineKids (eliminates `BeginStructureElementCap` + `acquireStructKids` on HFT rows)
- `appendDecimal` 5-digit fast path for MCIDs ≥ 10,000
- Bounded caches: `subsetCache` (1,024), `imgCache` (256), `propsCache` (8,192)

**HFT shared-table fast path (P0–P8, P17–P25)**
- Compliant **TR → TD** hierarchy restored (one `TD` per column, distinct MCID) — no compliance shortcuts
- `drawSharedLayoutRow` precomputed row fragments; per-stripe `PreallocatePageMCIDSlots`
- `charsPreScanned` flag eliminates `MarkCharsUsed` double-scan
- HFT-aware buffer capacity estimation for compliant ~2.29 MB output (not legacy 748 KB compacted shape)

**PDF/A, metadata, fonts (P9, Phase A 15K)**
- Gray + sRGB ICC profiles cached at `init` (`grayICCProfileCompressed`, `srgbICCProfileCompressed`)
- **P26 (landed):** `GetSRGBICCProfile()` returns cached bytes; `GenerateOutputIntent` reuses compressed payload
- **P28 (landed):** Precomputed startup font hint on Zerodha templates
- **P30 (landed):** Static XMP metadata shell with emit-time date/ID patch only

**Signing (`internal/pdf/signature/`)**
- ECDSA P-256 default (replacing RSA hot path)
- PKCS#7 marshal buffer pooling (`pkcs7MarshalBuffersPool`); `appendByteRangeMarker` + `encodeHexUpper`
- PEM/signer cache for benchmark harness

### HTTP Handlers (`internal/handlers/`, Gin)

- Sonic JSON decode with pretouch; tier prealloc; HFT split decode fast path
- `GenerateTemplatePDFBorrowed` borrowed-buffer API; `GIN_FAST_API=1`
- Gin Phases 0–11: concurrency alignment (48 workers), struct direct-write, batch MC, HFT deferral, page striping
- k6 regression fix: unbounded `sharedRowRenderCache` → bounded entry/byte-capped cache (max 4,096 entries, 64 MB)

### Python Bindings (`bindings/python/pypdfsuit/`)

- Root-caused Python latency: `to_dict()` tree walks + JSON/CGO overhead (not the renderer)
- Precomputed dataclass field-to-JSON-key mappings; specialized serializers for `PDFTemplate`, `Table`, `Row`, `Cell`, `Config`
- Compact UTF-8 JSON (`ensure_ascii=False`); HFT payload **~1.055 → 1.043 MB**
- **Removed** automatic JSON payload caching and cache benchmark targets — honest full-execution benchmark restored
- Added `PAYLOAD_SCENARIO`, p50/p95/p99 latency reporting; `test_serializer_schema.py` for Go-facing key parity

### SlopGuard Remediation (218/218 PERF findings)

Across `cmd/`, `internal/handlers/`, `internal/pdf/`, `typstsyntax/`, `sampledata/`:
- PERF-1: 20+ regex compilations hoisted to package level
- PERF-6/15/35: 30+ `fmt.Sprintf`/`strconv.Itoa` → `strconv.AppendInt` / `strings.Builder`
- PERF-4: map pre-sizing + `clear()` reuse; PERF-31: 13 `defer` removals in font registry
- PERF-41/43: non-blocking logging; `gin.CustomRecovery`

### Benchmarks, Validation & Release

- Zerodha: `bench-gopdflib-zerodha-x10-pprof` as single timing+profile target; x10 mean as regression gate
- PyPDFSuit: `make bench-pypdfsuit-zerodha` / `-x5` / `-x10` on full execution path only
- veraPDF PDF/A-4 + PDF/UA-2 post-test gate (`make test-verify-pdfs` — **36/36 PASS**)
- XFDF `startxref` repair; CI `backend-test` job; Go **1.26.4** + module bump to **gopdfsuit/v6**

### Zerodha Active Trader (Phase A 15K — landed, gates pending)

- **P29:** `SharedRowLayout: true` on active 41-row trade table in `buildActiveTraderTemplate()`

---

## Benchmark Results (Detailed)

### Zerodha In-Process (`make bench-gopdflib-zerodha-x10`, 48 workers, GOMAXPROCS=24)

| Milestone | Mean ops/s | Best ops/s | Notes |
|-----------|----------:|----------:|-------|
| Go 1.24 gold standard | 573 | — | Historical baseline |
| 2026-06-11 gold standard | 1,135 | — | +98% vs Go 1.24 |
| 2026-06-17 x10 (non-compliant HFT) | 7,439 | 10,532 | HFT output 748 KB — TR→TD collapsed |
| 2026-06-20 pre-opt (compliant TR→TD) | 2,799 | 3,272 | Compliance rebuild cost |
| Phase 1 (P0–P8) | 5,268 | 5,644 | +88% vs pre-opt |
| Phase 2 (P9–P16) | 5,543 | 5,703 | Lazy arena fix after P12 regression |
| Phase 3 (P17–P20) | 7,432 | 8,327 | Arena race fix (P20) |
| **Phase 4 idle (P21–P25)** | **9,594** | **10,005** | **8K target met** |

**Compliance gate (held after every phase):**

| Output | Size (bytes) | PDF/A-4 | PDF/UA-2 |
|--------|-------------:|:-------:|:--------:|
| Retail | 61,293 | PASS | PASS |
| Active | 76,065 | PASS | PASS |
| HFT (compliant TR→TD) | 2,291,950 | PASS | PASS |

### Gin HTTP (`make bench-k6`, 48 VU × 35s, `tagged_ecdsa`)

| State | Throughput | Heap | `drawSharedLayoutRow` heap |
|-------|----------:|-----:|---------------------------:|
| Master baseline | 674 req/s | 587 MB | — |
| Phase 11 peak | **1,232 req/s** | — | — |
| Feature branch (pre-fix) | hung ~16k iter | **1,603 MB** | **497 MB** |
| Post bounded-cache fix | **1,223 req/s** | **522 MB** | **7.0 MB** |
| `bench-k6-light` post-fix | **1,277 req/s** | **131 MB** | **3.5 MB** |
| 2026-06-14 2-run mean | **1,026 req/s** | — | — |

### PyPDFSuit (`make bench-pypdfsuit-zerodha`, 48 workers, honest full path)

| Pass | Throughput | Notes |
|------|----------:|-------|
| Pre-pass baseline (x5) | 221 ops/s | `to_dict` dominated (50–67% HFT) |
| P0 serializers + cache removal | 565 ops/s | HFT `to_dict` 60 → 20 ms |
| P2 direct serializers | 737 ops/s | Python no longer dominates cProfile |
| P3 x10 mean | 751 ops/s | Best 927 ops/s |
| **P4 compact UTF-8 (x5)** | **836 ops/s** | **Publishable number**; `cgo_call` now 73–92% |

Native Go comparison: **11,721 ops/s** on same weighted mix.

### Internal Micro-Benchmarks (`./internal/pdf`, benchtime=5s)

| Benchmark | Δ vs baseline |
|-----------|--------------|
| `BenchmarkGoPdfSuit` | **−11.6%** ns/op, **−16.0%** B/op, −3.7% allocs |
| `BenchmarkGenerateTemplatePDF/Rows2000` | **−17.5%** ns/op, −9.0% B/op |

### GoPDFKit Comparison (7 workloads, 3-run median)

gopdflib wins **7/7** — median lead **+40% to +645%** (e.g. `png_rows_60`: 30,018 vs 4,028 pdf/s; `text_short`: 204,214 pdf/s post TD fix). GoPDFKit allocates **5–35×** more bytes on heavy workloads.

---

## Compliance & Quality Gates

All optimizations preserve PDF output semantics. No compliance shortcuts were taken in the shipped path.

| Gate | Result |
|------|--------|
| veraPDF PDF/A-4 (retail/active/HFT) | **6/6 PASS** |
| veraPDF PDF/UA-2 (retail/active/HFT) | **6/6 PASS** |
| HFT TR→TD hierarchy | Preserved — one `TD` per column, distinct MCID |
| HFT output size | **2,291,950 bytes** (±5% of compliant baseline) |
| Retail ECDSA P-256 signing | Enabled (80% workload) |
| `make test-verify-pdfs` | **36/36 PASS** |
| `go test ./internal/...` | PASS |

### Hard Guardrails (enforced throughout)

- **No TR→TD collapse** on HFT (748 KB output explicitly rejected)
- **No key-based cross-request caches** (`sharedRowRenderCache` bounded only — k6 regression documented 2026-06-17)
- **No disabling** of tagging, structure-tree, signing, PDF/A metadata, output intent, or font subsetting
- **veraPDF gate after every phase** — any HFT FAIL reverts immediately

---

## Regressions Found & Fixed

| Issue | Root Cause | Fix |
|-------|-----------|-----|
| k6 HTTP hang / OOM | Unbounded global `sharedRowRenderCache` (`sync.Map`, near-zero hits, unbounded growth under 48 concurrent HFT docs) | Bounded entry/byte-capped cache; slice copy on store; eviction |
| P12 arena regression (4,547 ops/s) | 32 KiB arena slab activated for every tagged PDF including 80% retail | Lazy arena activation (`ReserveElementCapacity ≥ 512` only) |
| P17 arena data race | Slice header copy from `sync.Pool` aliased backing arrays across 48 workers | Return pool object pointer directly; `WarmArenaSlabPool(6)` |
| PyPDFSuit inflated benchmark | JSON payload cache bypassed per-call serialization | Cache removed from benchmark surface; honest full path restored |
| G3 parallel structure-tree | ~60% throughput regression | Reverted |
| G4 template PDF cache | No meaningful E2E gain | Removed |
| Gin Phase 12 (CRC32, sig-hex, store-uncompressed, 2× compress workers) | No E2E gain | Reverted |

---

## Reverted / Not Shipped Experiments

These were tried and intentionally reverted or rejected:

- HFT TR→TD collapse (748 KB output — non-compliant)
- Key-based row render cache expansion
- G3 parallel structure-tree build, G4 template PDF cache
- Gin Phase 12: CRC32 fingerprint, in-place sig hex, store-uncompressed pages
- Generic structure writer, aggressive buffer caps (HI-3)
- PyPDFSuit JSON-cache benchmark targets (harness tuning, not execution throughput)
- Per-SM 4 MB struct-element arena (1 GB live-heap pressure across 48 workers — reverted; per-P shard fix retained)

---

## Remaining Bottlenecks

| Hotspot | Share | Surface |
|---------|-------|---------|
| `bytes.growSlice` + arena slabs | ~47% + 44% heap | Zerodha (all templates) |
| `runtime.memmove` / `memclr` | 12% / 11% flat CPU | Zerodha |
| `drawTable` / `drawSharedLayoutRow` | 22–30% cum CPU | HFT-heavy |
| `compress/flate` | ~20% cum CPU | Gin + Zerodha |
| Sonic JSON decode | ~16–18% alloc_space | Gin HTTP |
| `_bindings.call_bytes_result` (CGO) | 73–92% latency | PyPDFSuit |
| Mean peak allocated | ~1.1–1.2 GB | vs 600–650 MB aspirational gate |

---

## What's Next (15K Roadmap — In Progress)

Six-agent cross-validation and four-subagent profile refresh established the path from idle **9,009 → 15,000 ops/s** (−5,991 gap).

**Phase A (landed in code; throughput gates pending):**
- P26 sRGB ICC cache fix, P28 font precompute, P30 static XMP shell, P29 active SharedRowLayout

**Phase B (planned — memory wall):**
- P31 pdfBuffer zero-grow, P34 page-stream caps, P40 row-stream direct append, P32 arena tiering, P35 signature cleanup

**Phase C (planned — HFT tail, mandatory for 15K):**
- P36 arena TD template, P38 batch struct emit, P37 stripe-batch arena, P39 glyph dedupe, P41 retail row-batch PDF/UA, P42 hand-built PKCS#7 DER

```
9,009 ──Phase A──► ~10,500–11,000 ──Phase B──► ~13,000 ──Phase C──► ~15,000
```

**Gin 1,500 req/s weighted:** HFT tail (5% × ~250 ms) remains primary ceiling; flate tuning, sonic codegen unmarshaler, buffer pre-sizing on remaining checklist.

**PyPDFSuit:** Further gains require Go render-boundary work or a new API contract (handle/batch/service mode) — not achievable with Python-only changes alone.

---

## How to Verify

```bash
# Unit tests
go test ./internal/...

# Compliance gate (mandatory)
make test-verify-pdfs

# Zerodha in-process (the 8K gate)
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache \
  make bench-gopdflib-zerodha-x10

# Zerodha timing + pprof
GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache \
  make bench-gopdflib-zerodha-x10-pprof

# Gin HTTP regression check
make bench-k6-light

# PyPDFSuit honest full path
make bench-pypdfsuit-zerodha
make bench-pypdfsuit-zerodha-x5

# Python binding tests
cd bindings/python && python3 -m pytest tests
```

**Measurement hygiene:**
- Run x10 benchmarks on an **idle machine** (no parallel Docker/browser/load)
- Use `GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache` for reproducible builds
- Compare against idle **9,009 ops/s** baseline, not load-depressed runs (~7,852 ops/s under WSL2 load)

---

## Documentation

Detailed execution logs and checklists:

| Date | Summary |
|------|---------|
| 2026-06-10 | Master plan, Phase 1–2 engine, SlopGuard, GoPDFKit parity |
| 2026-06-11 | GoPDFKit 7/7, Zerodha gold standard, Gin HTTP, 3k push |
| 2026-06-14 | k6 weighted bench, cache bounds, structure prealloc |
| 2026-06-17 | Zerodha x10 pprof, k6 regression recovery |
| 2026-06-18 | PyPDFSuit profile (root cause: `to_dict` + CGO) |
| 2026-06-19 | PyPDFSuit honest benchmark, serializer cleanup |
| 2026-06-20 | Zerodha compliant TR→TD to 8K (P0–P25) |
| 2026-06-21 | Six-agent 15K cross-validation |
| 2026-06-22 | 15K phased checklist, Phase A partial landing |
| Cross-cutting | SlopGuard 218/218, PR phases 1–10, v6 release prep |

All executive summaries: `guides/optimizations/executive_summaries/`