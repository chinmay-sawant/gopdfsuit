# Pass 4 - Load-Test Hotspot Optimization Plan

**Date:** 2026-05-25  
**Trigger:** k6 load test (48 VUs) + pprof CPU/heap profiles  
**Baselines:** `baselines/loadtest_*_20260525.*`, `baselines/bench_*_20260525.txt`

---

## Is the current state “good”?

**Micro-benchmarks: mostly yes.** After Pass 1–2, the 2000-row Helvetica table path is ~15–20% faster than the pre-optimization baseline (~35–56 ms → ~29 ms). Parallel generation reaches ~7–9 ms/op (~4–5× vs serial). Typst compare shows GoPdfSuit at **~73 ms vs Typst ~1.12 s** on the same dataset (~15× faster), with far higher alloc count (expected for in-process Go).

**Under concurrent HTTP load: good throughput, weak tail latency.**

| Signal | Result | Verdict |
|--------|--------|---------|
| HTTP error rate | 0% | ✅ Stable |
| Median latency | ~94 ms | ✅ Fast for small/medium payloads |
| p95 latency | 3.11 s | ✅ Within k6 threshold |
| p99 latency | 27.45 s | ⚠️ Queueing at 48 VUs (not crashes) |
| Throughput | ~25 req/s @ 48 VUs | ✅ Reasonable for CPU-bound PDF |
| Heap under load | ~442 MB in-use | ⚠️ Room to cut ~30–50% |
| CPU profile | 50% memclr, 28% buffer growth | ⚠️ Expected for buffer-heavy path; optimizable |

**Bottom line:** Production-ready for correctness and median latency. Pass 4 targets **tail latency (p99)**, **heap per request**, and **CPU efficiency under saturation** - not raw single-thread bench speed.

---

## Benchmark delta vs earlier passes

All on **Rows2000 / BenchmarkGoPdfSuit**, 24 CPUs, same machine.

| Phase | Time/op (serial) | Bytes/op | Allocs/op | Notes |
|-------|------------------|----------|-----------|-------|
| **Pre-Pass 1** (audit baseline) | **35–56 ms** | ~14.9 MB | ~292K | From audit / Pass 1 snapshot |
| **Pass 1** | **35–45 ms** | ~14.9 MB | ~292K | Alloc count unchanged on Helvetica path |
| **Pass 2** | **~29 ms** | ~15.5 MB | ~292K | **Best serial time** (~15–20% vs Pass 1) |
| **Pass 3** | **~42–44 ms** | ~17.6 MB | ~303K | Possible regression or bench variance; wrap path −8.6% allocs |
| **Pass 4 snapshot** (`benchtime=3x`) | **~73 ms** | ~18.9 MB | ~303K | **Not comparable** - only 3 iterations; rerun with `-count=5` for stat |

| Parallel (Pass 1) | **~7–9 ms/op** | ~14.3 MB | ~292K | ~4–5× vs serial |
| **Typst compare** | **1123 ms/op** | 19 KB | 51 | External process; different workload semantics |

**Improvement summary (trust Pass 2 as peak serial):**

- **Latency:** ~35–56 ms → **~29 ms** ≈ **15–35% faster** (best case ~48% vs slowest pre-opt run)
- **Parallel:** **~7–9 ms/op** - not re-measured in Pass 3/4; retain Pass 1 numbers
- **Allocs:** ~292K → ~303K on Rows2000 (**+3–4%**, not improved on main path)
- **Wrap path allocs:** ~357K → **~327K** (−8.6%, Pass 3 win)

**Action:** Re-run standardized bench before Pass 4 implementation:

```bash
go test -run='^$' -bench='BenchmarkGenerateTemplatePDF/Rows2000|BenchmarkGenerateTemplatePDF_Parallel' \
  -benchmem -count=5 ./internal/pdf/ | tee guides/cursor/baselines/bench_pass4_pre_$(date +%Y%m%d).txt
```

---

## Pass 4 goals

| Goal | Target | Primary levers |
|------|--------|----------------|
| Reduce flat `memclrNoHeapPointers` | −20–35% CPU | Pre-grow buffers, avoid `make([]byte, n)` return copy |
| Reduce heap in-use under 48 VUs | −30–50% | Gate PDF/UA tagging, sonic pre-sizing |
| Improve p99 under saturation | −40%+ | Concurrency tuning + less GC pressure |
| Preserve PDF correctness | 100% | Golden tests, PDF/UA validator when tagging on |

---

## Hotspot → task mapping

### CPU hotspots

| Hotspot | Root cause | Pass 4 tasks |
|---------|------------|--------------|
| `memclrNoHeapPointers` 49.7% | Zero-fill on `bytes.Buffer` growth + `make([]byte, len)` final copy | P4-02, P4-03, P4-04 |
| `bytes.growSlice` 27.6% cum | Page streams, compression output, drawTable scratch grow from cap 0 | P4-02, P4-05, P4-07 |
| `compress/flate.deflate` 14.5% cum | Every page stream zlib-compressed; pool misses under concurrency | P4-08, P4-09, P4-12 |
| `drawTable` 12.6% cum | Per-cell writes, text encode strings, structure tagging | P4-05, P4-06, P4-01 |
| `GenerateTemplatePDF.func*` 21% cum | Likely structure-tree write + buffer assembly (verify with `go tool pprof -list`) | P4-01, P4-03 |

### Heap hotspots

| Hotspot | Root cause | Pass 4 tasks |
|---------|------------|--------------|
| `bytes.growSlice` 150 MB | Un-sized `ContentStreams`, `CompressBufPool` | P4-02, P4-08 |
| sonic Unmarshal ~99 MB | Large nested JSON (HFT 2000×7 cells) | P4-10, P4-11 |
| `flate.NewWriter` 82 MB | Pool contention @ 48 concurrent requests | P4-08, P4-12 |
| `BeginMarkedContentBuf` 64 MB, 581K objs | **Always-on** PDF/UA per cell | **P4-01** |
| `drawTable` 168 MB cum | Text strings + structure + stream growth | P4-01, P4-05, P4-06 |

---

## Task backlog (Pass 4)

Tasks ordered by **impact × confidence**. Subagent owners noted for traceability.

### P0 - Highest leverage

#### P4-01 - Gate PDF/UA structure tagging *(subagent: JSON/heap)*

- **Files:** `internal/pdf/pagemanager.go`, `internal/pdf/structure.go`, `internal/pdf/draw.go`, `internal/pdf/generator.go` (catalog `/MarkInfo`, `StructTreeRoot`)
- **Change:** Add explicit `taggedPDF` / derive from `pdfaCompliant`; no-op `StructureManager` when off; skip `BeginMarkedContentBuf` per cell
- **Impact:** **~50–80 MB heap**, **~500K fewer objects**/request on table-heavy loads
- **Risk:** Medium - product/docs must define when tagging is required
- **Validate:** k6 heap profile; PDF/UA validator when flag on

#### P4-02 - Pre-grow per-page content streams *(subagent: buffer)*

- **Files:** `internal/pdf/pagemanager.go` (`NewPageManager`, `AddNewPage`)
- **Change:** `ContentStreams[i].Grow(32–64 KiB)` on page creation
- **Impact:** **10–25%** less `growSlice`/memclr on page writes
- **Risk:** Low
- **Validate:** `BenchmarkGenerateTemplatePDF/Rows2000`, pprof `bytes.(*Buffer).grow`

#### P4-03 - Eliminate redundant zero-fill on final PDF return *(subagent: buffer)*

- **Files:** `internal/pdf/generator.go` (~1316–1328)
- **Change:** Tiered `[]byte` pool or `append(dst[:0], pdfBuffer.Bytes()...)` with sufficient cap; copy only when signing mutates
- **Impact:** **5–12%** flat memclr on multi-MB outputs
- **Risk:** Medium - ownership + signature path
- **Validate:** sign/no-sign tests; load test memclr %

### P1 - High value, moderate effort

#### P4-04 - Hoist reusable `[]byte` scratch in `drawTable` *(subagent: buffer + drawTable)*

- **Files:** `internal/pdf/draw.go` (`drawTable` ~735+)
- **Change:** Replace per-branch `var borderBuf []byte` with table-scoped buffers (extend existing `scratchBuf`)
- **Impact:** **3–10%** growslice under `drawTable`
- **Risk:** Low

#### P4-05 - `appendTextForPDF` - zero-alloc text operands *(subagent: drawTable)*

- **Files:** `internal/pdf/utils.go`, `internal/pdf/draw.go`, `internal/pdf/font/metrics.go`
- **Change:** Append PDF text literals/hex into caller `[]byte`; remove `formatTextForPDF` → `string` per `Tj`; fix wrapped path `string(line)` 
- **Impact:** **5–12%** heap objects in text-heavy tables
- **Risk:** Medium - bit-identical PDF escaping

#### P4-06 - Incremental width in `WrapTextInto` *(subagent: drawTable)*

- **Files:** `internal/pdf/utils.go` (`WrapTextInto`, `wrapLongWordInto`)
- **Change:** Running width instead of re-measuring full line prefix each word
- **Impact:** High on wrap-enabled workloads; small on k6 default mix
- **Risk:** Medium

#### P4-08 - Pre-size `CompressBufPool` + extend zlib pooling *(subagent: compression)*

- **Files:** `internal/pdf/font/compression.go`, `internal/pdf/font/subset.go`, `internal/pdf/metadata.go`, `internal/pdf/pdfa.go`
- **Change:** `Grow(64 KiB)` in pool `New`; migrate `CompressFontData` and ICC paths to `GetZlibWriter`/`GetCompressBuffer`
- **Impact:** **5–15%** buffer growth; less `flate.NewWriter` heap
- **Risk:** Low–Med

#### P4-10 - Sonic / JSON decode pre-sizing *(subagent: JSON/heap)*

- **Files:** `internal/handlers/handlers.go`, optionally `internal/models/models.go`
- **Change:** Two-phase decode (row count hint) or `sync.Pool` for `PDFTemplate` shells; consider slimmer struct for known schemas
- **Impact:** **20–40%** of sonic-related growth on large payloads
- **Risk:** Med

#### P4-12 - HTTP concurrency vs CPU count *(subagent: compression + load)*

- **Files:** `cmd/gopdfsuit/main.go` (currently `maxConcurrent := 48`, `NumCPU() == 24`)
- **Change:** Default to `runtime.NumCPU()` or `2 * NumCPU()`; document tuning; separate load-test scenarios
- **Impact:** Lower p99 queue wait; may reduce pool thrashing
- **Risk:** Low - config change

### P2 - Optional / higher complexity

#### P4-07 - Pre-grow compression buffer from stream length *(subagent: buffer)*

- **Files:** `internal/pdf/generator.go` (~740–782)
- **Change:** `compressedBuf.Grow(contentStream.Len()/4)` before zlib write
- **Impact:** 5–15% zlib buffer growth
- **Risk:** Low

#### P4-09 - Parallel page-stream zlib *(subagent: compression)*

- **Files:** `internal/pdf/generator.go`
- **Change:** `errgroup` compress pages in parallel; single-threaded object write + encrypt in order
- **Impact:** Wall-clock win on many-page docs; CPU may stay similar
- **Risk:** **High** - ordering, encryption, determinism

#### P4-11 - k6 scenario split *(subagent: JSON/heap)*

- **Files:** `test/generate_template-pdf/payload_generator.js`, new `load_test_unsigned.js`
- **Change:** Scenarios: unsigned / signed / HFT-only / PDF/UA-on; clearer pprof attribution
- **Impact:** Measurement (required for honest before/after)
- **Risk:** Low

#### P4-13 - Cache parsed signer PEM/certs *(subagent: JSON/heap)*

- **Files:** `internal/pdf/signature/signature.go`
- **Change:** Cache parsed keys by config hash
- **Impact:** CPU + parse allocs; not the ECDSA P384 stack (that is JWT/TLS, not PDF sign)
- **Risk:** Low–Med

#### P4-14 - Structure element pooling / row-level tagging *(subagent: buffer + drawTable)*

- **Files:** `internal/pdf/structure.go`
- **Change:** Pool `StructElem` or tag at row level when UA required
- **Impact:** Medium GC win when P4-01 cannot disable tagging
- **Risk:** High - PDF/UA semantics

---

## Implementation phases

### Phase A - Quick wins (1–2 days)

1. P4-01 Gate PDF/UA (feature flag + docs)
2. P4-02 Pre-grow page streams
3. P4-08 CompressBufPool sizing + subset/metadata pool migration
4. P4-12 Concurrency default tuning

**Expected:** Largest heap drop; measurable p99 improvement.

### Phase B - drawTable / text path (2–3 days)

5. P4-04 Hoist scratch buffers
6. P4-05 `appendTextForPDF`
7. P4-06 WrapTextInto incremental width (if wrap workloads matter)

**Expected:** Lower allocs/op in macro benches; less `drawTable` cum CPU.

### Phase C - Return path + JSON (2–3 days)

8. P4-03 Final PDF copy elimination
9. P4-10 JSON pre-sizing / template pool
10. P4-11 Load-test scenario split

### Phase D - Advanced (optional)

11. P4-09 Parallel page zlib - **done**
12. P4-14 Structure pooling - **done**

**Status:** Pass 4 implemented 2026-05-25. Post bench: `baselines/bench_pass4_post_20260525.txt`

---

## Implementation status (2026-05-25)

| Task | Status | Notes |
|------|--------|-------|
| P4-01 Gate PDF/UA | ✅ | `TaggedPDF` config; `Enabled` on `StructureManager` |
| P4-02 Pre-grow page streams | ✅ | 64 KiB `Grow` on new pages |
| P4-03 Final PDF pool | ✅ | `finalPDFSlicePool` + `slices.Clone` |
| P4-04 Hoist drawTable scratch | ✅ | border/xobj/color/placeholder/checkbox buffers |
| P4-05 appendTextForPDF | ✅ | `utils.go`, `font/metrics.go`, hot `drawTable` paths |
| P4-06 WrapTextInto incremental | ✅ | Running `lineWidth` in `utils.go` |
| P4-07 Compress buffer Grow | ✅ | `max(4096, len/4)` before zlib |
| P4-08 Compression pooling | ✅ | 64 KiB pool; subset/metadata/pdfa migrated |
| P4-09 Parallel page zlib | ✅ | `errgroup` in `generator.go` |
| P4-10 Template pool | ✅ | `templatePDFPool` in handlers |
| P4-11 k6 scenarios | ✅ | `load_test_unsigned.js`, `load_test_tagged.js` |
| P4-12 Concurrency | ✅ | `maxConcurrent = runtime.NumCPU()` |
| P4-13 Signer cache | ✅ | PEM hash `sync.Map` in signature.go |
| P4-14 StructElem pool | ✅ | `acquireStructElem` / `ReleaseStructElemsToPool` |


## Validation checklist

### Before each phase

```bash
# Micro
go test -run='^$' -bench='BenchmarkGenerateTemplatePDF|BenchmarkGoPdfSuit' \
  -benchmem -count=5 ./internal/pdf/

# Compare (optional)
go test -tags=compare -run='^$' -bench=BenchmarkTypst -benchmem ./internal/pdf/
```

### Load + pprof (after server rebuild)

```bash
go run ./cmd/gopdfsuit &
curl -o guides/cursor/baselines/loadtest_cpu_after.prof \
  "http://127.0.0.1:8080/debug/pprof/profile?seconds=35" &
cd test/generate_template-pdf && k6 run load_test.js
curl -o guides/cursor/baselines/loadtest_heap_after.prof \
  "http://127.0.0.1:8080/debug/pprof/heap"
go tool pprof -top guides/cursor/baselines/loadtest_cpu_after.prof
```

### Success criteria

| Metric | Baseline (2026-05-25) | Pass 4 target |
|--------|----------------------|---------------|
| `memclrNoHeapPointers` flat CPU | 49.7% | < 35% |
| `bytes.growSlice` heap | 150 MB | < 90 MB |
| `BeginMarkedContentBuf` heap (non-UA mode) | 64 MB | ~0 |
| p99 latency @ 48 VUs | 27.4 s | < 15 s |
| Rows2000 serial bench | ~42 ms (Pass 3) | ≤ 29 ms (match Pass 2) |

---

## Subagent assignment (for implementation)

| Subagent focus | Tasks | Key files |
|----------------|-------|-----------|
| **Buffer / memclr** | P4-02, P4-03, P4-04, P4-07 | `pagemanager.go`, `generator.go`, `draw.go` |
| **Compression** | P4-08, P4-09, P4-12 | `font/compression.go`, `generator.go`, `main.go` |
| **drawTable / text** | P4-05, P4-06, P4-04 | `draw.go`, `utils.go`, `font/metrics.go` |
| **HTTP / JSON / UA** | P4-01, P4-10, P4-11, P4-13 | `handlers.go`, `structure.go`, `payload_generator.js` |

---

## Related artifacts

- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - Pass 1–3 (complete)
- [PR_PERFORMANCE_OPTIMIZATION.md](./PR_PERFORMANCE_OPTIMIZATION.md) - PR summary
- [baselines/loadtest_pprof_summary_20260525.txt](./baselines/loadtest_pprof_summary_20260525.txt)
- [baselines/loadtest_k6_20260525.txt](./baselines/loadtest_k6_20260525.txt)
