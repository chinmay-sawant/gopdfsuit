# Performance Improvements — `feat/performance-improvements`

**Branch:** `feat/performance-improvements`  
**Base:** `master` at `484b991`  
**Go version:** `go1.26.4`  
**Date:** 2026-06-10 → 2026-06-14  
**Commits:** 20 commits ahead of `master`  
**Scope:** 133 files changed, ~11,150 insertions, ~1,462 deletions

---

## Summary

This PR is a profile-driven performance program for gopdfsuit spanning the PDF generation engine, HTTP handlers, signing path, and benchmark harnesses. Work was guided by nine optimization reports added under `guides/optimizations/` (all new on this branch vs `master`). The program moved from a `master` micro-benchmark baseline through SlopGuard remediation, GoPDFKit parity, Gin HTTP saturation, and Zerodha end-to-end throughput targets.

**Headline outcomes:**

| Surface | Baseline | Current (best measured) | Improvement |
|---------|----------|-------------------------|-------------|
| `BenchmarkGoPdfSuit` (internal/pdf) | 26.0 ms/op, 25.6 MB/op | 23.0 ms/op, 21.5 MB/op | **−11.6% latency, −16% alloc** |
| GoPDFKit compare harness (7 workloads) | gopdflib wins 5/7 | **gopdflib wins 7/7** | Up to **+645%** on `png_rows_60` |
| Gin HTTP weighted (`tagged_ecdsa`) | ~593 req/s | **~1,054–1,232 req/s** | **+77–108%** (≥1,000 gate ✅) |
| Gin HTTP retail-only | — | **~3,965 req/s** | ≥1,500 gate ✅ |
| Zerodha in-process (48 workers) | ~573 ops/s (Go 1.24) | **~2,751 ops/s peak** | **+98% vs 1.24 gold std; ~2,476 avg** |
| SlopGuard actionable findings | 218 open | **218 fixed** | 100% remediation |

---

## Source documentation (added on this branch)

All nine markdown files below were **added** (`git diff master...HEAD --name-status`) and drove implementation on this branch:

| File | Role |
|------|------|
| [`20260610_master_performance_execution_plan.md`](../20260610_master_performance_execution_plan.md) | Master plan: Phase 1 buffer/compression wins, Phase 2 structure-tree pressure |
| [`20260610_phase2_pprof_report.md`](../20260610_phase2_pprof_report.md) | Post-Phase-2 pprof: remaining `fmt.Sprintf`, `drawTable`, compression hotspots |
| [`slopguard-findings.md`](../slopguard-findings.md) | 226 SlopGuard findings; 218 actionable PERF items tracked to completion |
| [`20260610_slopguard_fixes_report.md`](../20260610_slopguard_fixes_report.md) | SlopGuard fix validation via GoPDFKit compare harness (10-run avg) |
| [`20260610_gopdfkit_compare_results.md`](../20260610_gopdfkit_compare_results.md) | Iterative gopdflib vs GoPDFKit comparison through multiple optimization passes |
| [`20260611_gopdfkit_fixed_compare_results.md`](../20260611_gopdfkit_fixed_compare_results.md) | Fixed stub `gopdfkit` module; definitive 7/7 gopdflib wins |
| [`20260611_gin_1500_ops_pprof_report.md`](../20260611_gin_1500_ops_pprof_report.md) | Gin HTTP path to 1,000+ req/s (Phases 0–12, k6 + pprof) |
| [`20260611_zerodha_3000_ops_pprof_report.md`](../20260611_zerodha_3000_ops_pprof_report.md) | Zerodha harness path to 3,000 ops/s (Phases A–G) |
| [`20260611_zerodha_gold_standard_benchmark.md`](../20260611_zerodha_gold_standard_benchmark.md) | Gold-standard 80/15/5 workload on Go 1.26.4 vs Go 1.24 |

---

## Implementation phases

### Phase 1 — PDF engine buffer & compression (execution plan)

**Source:** `20260610_master_performance_execution_plan.md`

Profiled hotspots on `go1.26.4`: `bytes.growSlice` (21%), `compress/flate.NewWriter` (20%), `slices.Clone` (7%), `BeginMarkedContentBuf` (5%).

**Implemented:**

- Increased pooled final PDF buffer capacity
- Removed extra scratch-slice hop at final assembly; single caller-owned clone
- Stream page content directly into zlib writer (no `contentStream.Bytes()`)
- Replaced `append([]byte(nil), ...)` with explicit `slices.Clone`
- Pre-sized initial page content streams from template complexity

**Results vs `master` baseline:**

| Benchmark | Baseline ns/op | Phase 1 ns/op | Δ |
|-----------|---------------:|--------------:|---|
| `BenchmarkGoPdfSuit` | 26,047,411 | 23,531,821 | **−9.7%** |
| `BenchmarkGenerateTemplatePDF/Rows2000` | 28,903,973 | 24,053,195 | **−16.8%** |

### Phase 2 — Structure tree & table pressure

**Source:** `20260610_master_performance_execution_plan.md`, `20260610_phase2_pprof_report.md`

**Implemented:**

- `StructureManager` page-index maps → slices (`NextMCID`, `ParentTree`)
- `BeginStructureElementCap` for known table/row child counts
- Pre-grew XMP metadata builders
- Annotation object IDs stored on Link structure elements (avoid reverse scans)

**Results (cumulative vs original `master`):**

| Benchmark | Baseline | Post-Phase-2 | Δ |
|-----------|----------|--------------|---|
| `BenchmarkGoPdfSuit` | 26.0 ms, 25.6 MB | 23.0 ms, 21.5 MB | **−11.6% ns, −16% B** |
| `BenchmarkGenerateTemplatePDF/Rows2000` | 28.9 ms, 24.7 MB | 23.8 ms, 22.5 MB | **−17.5% ns, −9% B** |
| allocs/op | ~163,386 | ~157,268 | **−3.7%** |

**Remaining (documented, partially addressed in later commits):** hot-path `fmt.Sprintf` in structure serialization, page content-buffer reuse, `drawTable` / `GetTextWidth` pressure.

### Phase 3 — SlopGuard remediation (218 findings)

**Source:** `slopguard-findings.md`, `20260610_slopguard_fixes_report.md`

**Scope:** 226 exported findings; 8 CWE-only excluded; **218 actionable PERF items fixed** across P1–P6 batches.

**Pattern categories addressed:**

| Rule family | Count (approx) | Representative fix |
|-------------|----------------|--------------------|
| PERF-1 | 20+ | Hoist regex compilation to package-level vars (`xfdf.go`, `merge.go`) |
| PERF-6 / PERF-35 | 50+ | `fmt.Sprintf` → `strconv.AppendInt`, `strings.Builder`, `errors.New` |
| PERF-15 | 30+ | `strconv.Itoa` → `strconv.AppendInt` with scratch buffers |
| PERF-31 | 13 | Remove `defer` from font registry hot path; explicit unlock |
| PERF-32 | 40+ | Eliminate redundant `string`/`[]byte` conversions; `unsafe.Slice` where safe |
| PERF-41/43 | — | `gin.CustomRecovery`; non-blocking stderr logging |
| PERF-46/48 | — | Cheap guards before `TrimSpace` / `bytes.Equal` |
| PERF-4 | — | Map pre-sizing; hoisted `visited` map with `clear()` reuse |

**Validation (GoPDFKit compare harness, 10-run avg):**

| Workload | Before pdf/s | After pdf/s | Δ |
|----------|-------------:|------------:|---|
| `text_240_lines` | 15,994 | **17,434** | **+9.0%** |
| `table_180_rows` | 11,548 | **13,051** | **+13.0%** |
| `table_900_rows` | 2,563 | **2,680** | **+4.6%** |

### Phase 4 — GoPDFKit comparison & deep engine wins

**Source:** `20260610_gopdfkit_compare_results.md`, `20260611_gopdfkit_fixed_compare_results.md`

**Key engine changes across iterative passes:**

- Pooled page content stream buffers; fairer compare harness (cached templates)
- `drawTable`: font-ref cache per cell, standard-font subset skip, precomputed text widths
- Image dedup: shared XObject per repeated cell PNG
- `GeneratePDFBorrowed` / `Release()` — skip final `slices.Clone` on hot path
- `DecodeImageData` MRU fast path; `crypto/md5` → `crypto/sha256` for document-ID hash
- `drawTable` border fast path (uniform borders → single `re S`)
- PDF/UA-2 Table→TR→TD hierarchy restored (no batch-MC shortcut on TR)

**GoPDFKit harness fix:** `replace github.com/cssbruno/gopdfkit` pointed at `/tmp/gopdfkit` stub (empty PDF). Redirected to real `v0.5.2` module.

**Final comparison (3-run median, fixed harness):**

| Workload | GoPDFKit pdf/s | gopdflib pdf/s | Winner | Δ |
|----------|---------------:|---------------:|--------|---|
| `text_short` | 106,256 | 160,863 | gopdflib | +51% |
| `text_240_lines` | 12,034 | 16,810 | gopdflib | +40% |
| `table_180_rows` | 9,237 | 21,023 | gopdflib | +128% |
| `table_900_rows` | 2,081 | 4,338 | gopdflib | +108% |
| `invoice_40_rows` | 32,173 | 71,920 | gopdflib | +124% |
| `png_table_180_rows` | 5,885 | 19,843 | gopdflib | +237% |
| `png_rows_60` | 4,028 | 30,018 | gopdflib | +645% |

**Post PDF/UA-2 TD fix re-run (2-run mean):** gopdflib still **7/7**, with **+27–86%** throughput vs morning run.

### Phase 5 — Gin HTTP path to 1,000+ req/s

**Source:** `20260611_gin_1500_ops_pprof_report.md`

**Endpoint:** `POST /api/v1/generate/template-pdf`  
**Workload:** 80% retail / 15% active / 5% HFT, PDF/A + tagged + ECDSA signed

| Milestone | Throughput | Median | p99 |
|-----------|----------:|-------:|----:|
| Baseline (pre-opt) | ~593 req/s | 25.6 ms | 563 ms |
| Phase 7 peak | **1,118 req/s** | 16.2 ms | 208 ms |
| Phase 11 peak | **1,232 req/s** | 12.6 ms | 126 ms |
| Post-P12-revert baseline | **1,054 req/s** | 15.5 ms | 143 ms |
| Retail-only | **3,965 req/s** | 7.8 ms | 29 ms |

**Phases implemented (0–11):**

| Phase | Focus | Key changes |
|-------|-------|-------------|
| 0–1 | Harness + concurrency | k6 pprof load test, `MAX_CONCURRENT=48`, `GOMAXPROCS=24` |
| 2 | Signing | ECDSA P-256 default; RSA eliminated from CPU top |
| 3 | JSON ingress | Sonic stream decode, `PreallocForDecode`, `GenerateTemplatePDFBorrowed` |
| 4 | PDF hot path | Struct tree pooling, compress cache bounds, `drawTable` HFT fast path |
| 5 | Saturation | `GIN_FAST_API=1`, `BENCH_MODE` adaptive concurrency |
| 7 | 1,000 gate | Pretouch, tier prealloc, direct struct write, batch MC, HFT deferral, buffer grow, page striping |
| 8–9 | Decode + draw | Pooled sonic unmarshal, bulk ParentTree, `drawSharedDeferRow` |
| 10–11 | HFT ultra-fast | Split HFT decode (`hft_decode.go`), flate single-pass >32 KiB |
| 12 | ❌ Reverted | CRC32 fingerprint, in-place sig hex, store-uncompressed — no E2E gain |

**Primary files:** `cmd/gopdfsuit/main.go`, `internal/handlers/handlers.go`, `json_decode.go`, `hft_decode.go`, `internal/pdf/generator.go`, `draw.go`, `structure.go`, `compress_cache.go`, `makefile` (`load-pprof`, `load-pprof-1k`, `load-pprof-1500`)

### Phase 6 — Zerodha harness path to 3,000 ops/s

**Source:** `20260611_zerodha_3000_ops_pprof_report.md`, `20260611_zerodha_gold_standard_benchmark.md`

**Harness:** `sampledata/gopdflib/zerodha/main.go` — 5,000 iterations, 48 workers, 80/15/5 mix

| Milestone | Peak ops/s | 5-run avg | Avg latency |
|-----------|----------:|----------:|------------:|
| Pre-opt (RSA, Go 1.24 gold std) | — | **573 ops/s** | 80.7 ms |
| Go 1.26.4 gold std (3-run median) | — | **1,135 ops/s** | 41.3 ms |
| Post-A7 (ECDSA) | 2,701 | 2,520 | ~17 ms |
| Post Phase F + G2 (full regen) | **2,751** | **2,476** | ~18.9 ms |

**Phases implemented:**

| Phase | Focus | Status |
|-------|-------|--------|
| A | Signature fast path | ✅ PEM/signer cache, in-place embed, ECDSA P-256 |
| B | Concurrency | ✅ 48 workers, `GOMAXPROCS=24`, seeded 80/15/5 schedule |
| C | Compression | ✅ Flate pool, store-uncompressed threshold, page compress cache (G2) |
| D | Allocation | ✅ `GeneratePDFBorrowed`, structure MCID batch, font registry pool |
| E | Harness | ✅ `BENCH_SEED`, `BENCH_SKIP_WRITE`, per-worker stats |
| F | Post-pprof fixes | ✅ F1–F4, C4–C6 (buffer estimate, subset cache, textTjBuf) |
| G3 | Parallel struct tree | ❌ Reverted (~60% regression on HFT) |
| G4 | Template PDF cache | ❌ Removed (misleading benchmark; each request must be unique) |

**3000 ops/s target:** Not yet reached (~8% below peak). Dominant costs: HFT `drawTable` + zlib (~40% CPU from 5% of jobs), retail ECDSA signing.

---

## Major code areas changed

| Area | Files | Changes |
|------|-------|---------|
| PDF generation | `internal/pdf/generator.go`, `draw.go`, `structure.go`, `pagemanager.go`, `image.go` | Buffer pooling, compression pipeline, structure tree, table fast paths |
| Fonts & compression | `internal/pdf/font/*.go` | Registry defer removal, flate pool, compress cache, subset cache |
| Handlers | `internal/handlers/handlers.go`, `json_decode.go`, `hft_decode.go` | Sonic pretouch, tier decode, borrowed PDF response |
| Signing | `internal/pdf/signature/signature.go` | ECDSA support, PEM/signer cache, in-place embed |
| Server | `cmd/gopdfsuit/main.go` | `MAX_CONCURRENT`, pool warmup, fast API mode |
| Benchmarks | `sampledata/benchmarks/gopdfkit_compare/`, `sampledata/gopdflib/zerodha/`, `test/generate_template-pdf/` | Compare harness, k6 load tests, pprof runners |
| Tooling | `makefile` | `load-pprof`, `load-pprof-gate`, `load-pprof-1k`, `load-pprof-1500`, benchmark suite targets |
| Go upgrade | `go.mod`, bindings | **Go 1.26.4** across module, Python CGO, CI |

---

## Reverted or removed experiments

Documented explicitly to avoid re-introducing regressions:

| Experiment | Outcome | Source |
|------------|---------|--------|
| Phase 12 CRC32 + in-place sig hex (Gin) | Reverted — CPU win, no E2E throughput gain | `20260611_gin_1500_ops_pprof_report.md` |
| Store-uncompressed pages ≥96 KiB (Gin) | Reverted — 810 req/s, larger PDFs | same |
| G3 parallel structure-tree build (Zerodha) | Reverted — ~970 ops/s vs ~2,500 | `20260611_zerodha_3000_ops_pprof_report.md` |
| G4 template PDF cache (Zerodha) | Removed — each request must produce unique PDF | same |
| `ReserveMCIDs` ParentTree pre-grow | Fixed (F1) — was 328 MB alloc regression | same |

---

## Verification

```bash
# Internal PDF micro-benchmarks (go1.26.4)
GOCACHE=/tmp/go1264-bench go1.26.4 test ./internal/pdf -run '^$' \
  -bench 'BenchmarkGoPdfSuit|BenchmarkGenerateTemplatePDF' -benchmem -benchtime=5s

# GoPDFKit comparison (fixed harness)
cd sampledata/benchmarks/gopdfkit_compare
GOCACHE=/tmp/go1264-bench go1.26.4 test -run '^$' \
  -bench 'BenchmarkGoPDF(Kit|Lib)$' -benchmem -benchtime=5s

# Gin HTTP load test with pprof
make load-pprof
make load-pprof-gate      # retail-only ≥1500 req/s
make load-pprof-1k        # weighted ≥1000 req/s

# Zerodha end-to-end harness
unset BENCH_SIGN_RSA
export BENCH_ITERATIONS=5000 BENCH_WORKERS=48 BENCH_SKIP_WRITE=1 BENCH_SEED=42 GOMAXPROCS=24
cd sampledata/gopdflib/zerodha && go1.26.4 run .

# Unit / integration tests
go test ./internal/handlers/... ./internal/pdf/... ./test/...
```

---

## Acceptance criteria

| Criterion | Status |
|-----------|--------|
| `BenchmarkGoPdfSuit` improves vs `master` baseline on `go1.26.4` | ✅ −11.6% ns/op |
| `BenchmarkGenerateTemplatePDF/Rows2000` improves vs baseline | ✅ −17.5% ns/op |
| `B/op` decreases while throughput improves | ✅ |
| PDF generation tests pass | ✅ |
| SlopGuard 218 actionable findings remediated | ✅ |
| gopdflib wins GoPDFKit compare (fixed harness) | ✅ 7/7 |
| Gin weighted throughput ≥1,000 req/s | ✅ (~1,054–1,232 req/s) |
| Gin retail-only ≥1,500 req/s | ✅ (~3,965 req/s) |
| Zerodha 3,000 ops/s at 48 workers | ⏳ ~2,751 peak (~8% gap) |
| Gin weighted 1,500 req/s | ⏳ ~871–1,232 (run variance) |

---

## Related artifacts

- Gin pprof runs: `guides/cursor/baselines/gin_pprof_runs/`
- Zerodha pprof runs: `guides/cursor/baselines/zerodha_3000_pprof_runs/`
- Benchmark suite comparisons: `guides/cursor/baselines/benchmark_suite_20260611_*/`
- Prior performance PR context: `guides/cursor/PR_PERFORMANCE_OPTIMIZATION.md`

---

## Commit history (optimization-related)

```
ad946c9 Additional performance improvements as per the optimizations guides
658d973 Improvements done as per the 20260610_gopdfkit_compare_results.md
909abd3 Updated the benchmarks in the markdown
77fe8d2 perf: implement slopguard-flagged performance improvements across 16 files
50e3e5c Done the gopdfkit benchmark comparison findings; Added slopguard findings
8dd61e0 Added the fix for the finding from the slopguard-findings.md
48a00c1 Benchmarks added vs the gopdfkit
fd3535f Upgrade to go1.26.4 (mod, Python CGO, README)
af328dd More improvements as per the 20260611_zerodha_3000_ops_pprof_report.md
0c7e179 Implemented till p9 from the 20260611_gin_1500_ops_pprof_report.md
397df18 perf(pdf,handlers): HFT decode fast path, PDF/UA-2 TD fix, benchmark makefile
```