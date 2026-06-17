# k6 benchmark regression analysis — feature branch vs master

**Date:** 2026-06-17  
**Branch (failing):** `feat/optimization-5.5-medium`  
**Branch (working):** `master` at `/home/chinmay/ChinmayPersonalProjects/t2/gopdfsuit`  
**Harness:** `make bench-k6` → 48 VU × 35s, `PAYLOAD_SCENARIO=tagged_ecdsa`  
**Conclusion:** Regression is caused by commit `fff4f16` (Zerodha x10 pprof optimizations), not by k6 or WSL alone.

---

## Summary

Master completes full `make bench-k6` on the same machine. The feature branch hangs or crashes mid-run. Side-by-side artifacts show **~2.7× higher live heap** on the feature branch and throughput freezing when all 48 server slots block.

The primary culprit is an **unbounded global `sharedRowRenderCache`** added in `internal/pdf/draw.go` — optimized for in-process Zerodha benchmarks, not validated under concurrent HTTP k6 load.

---

## Side-by-side: master vs feature branch

| | **master** (`t2/gopdfsuit`) | **feature branch** (`feat/optimization-5.5-medium`) |
|---|---|---|
| Run artifact | `20260617_233247` — **completed** | `20260617_232525` — **hung ~22s** |
| Harness | 48 VU × 35s | 48 VU × 35s |
| Iterations | **23,969** | **16,064** then frozen |
| Throughput | **674 req/s** | ~730 req/s then **0** |
| Errors | 0% | 0% (server hung, did not die) |
| Heap in-use | **587 MB** | **1,603 MB** (even on light 24-VU run) |
| `preallocInlineTableRows` | **46 MB** | **592 MB** |
| `drawSharedLayoutRow` | **~31 MB** | **497 MB** |

Master finishes cleanly. The feature branch uses **~2.7× more live heap** and stalls when all 48 VUs block waiting on the server.

### Light harness on feature branch (for reference)

`make bench-k6-light` (24 VU × 15s) **completed** on the feature branch (`20260617_232133`):

| Metric | Value |
|--------|------:|
| Throughput | 574 req/s |
| HTTP median | 8.0 ms |
| HTTP p99 | 517 ms |
| Errors | 0% |
| Heap in-use | 1,603 MB |

Lower concurrency avoids the hang but heap is still far above master.

---

## Failure modes observed on the feature branch

### 1. Server hard kill (~68–77%, earlier run)

- k6 errors: `EOF` → `connection reset by peer` → `connection refused`
- Server log: startup only, no panic, no graceful shutdown
- Likely OOM / hard kill when running alongside other heavy benchmarks

### 2. Server hang (~50–65%, run `20260617_232525`)

- k6 timer reaches 100% but iteration count freezes (~16,064)
- No connection-refused spam — server process alive but not completing requests
- All 48 VUs blocked; graceful shutdown took 80+ seconds
- No `pprof_summary` or heap/cpu profiles saved for this run

---

## What changed on the feature branch (not on master)

Only **3 commits** ahead of master merge base `abf6b4f`:

```
0283cb8  Refactored the path for the PR file
fff4f16  perf(pdf): optimize Zerodha x10 pprof hot paths   ← primary culprit
b70dbc4  generated CPU/heap profiles with make bench-gopdflib-zerodha-x5, ...
```

The large code change is **`fff4f16`**, touching:

- `internal/pdf/draw.go`
- `internal/pdf/generator.go`
- `internal/pdf/structure.go`
- `internal/pdf/font/compress_cache.go`
- `internal/pdf/font/metrics.go`
- `internal/pdf/metadata.go`
- `internal/pdf/pagemanager.go`
- `internal/pdf/signature/`

These were optimized for **in-process Zerodha** (`make bench-gopdflib-zerodha-x10`), not re-validated under **HTTP k6 load** (`make bench-k6`).

---

## Root cause: unbounded global row cache

The feature branch adds in `internal/pdf/draw.go` ( **not present on master** ):

```go
type sharedRowRenderCacheKey struct {
	row      *models.Row
	page     int
	mcidBase int
	y        int64
}

var sharedRowRenderCache sync.Map
```

In `drawSharedLayoutRow`:

```go
cacheKey := sharedRowRenderCacheKey{
	row:      rowPtr,
	page:     pageManager.CurrentPageIndex,
	mcidBase: rowMCIDBase,
	y:        scaledCoordKey(pageManager.CurrentYPos),
}
if cached, ok := sharedRowRenderCache.Load(cacheKey); ok {
	// ...
}
// ... render row ...
rendered := append([]byte(nil), rowBuf.Bytes()...)
sharedRowRenderCache.Store(cacheKey, rendered)
```

### Why this breaks under k6

1. **Global `sync.Map` with no eviction** — every stored row lives forever for the process lifetime.
2. **Cache key includes `mcidBase` and `y`**, which change on almost every row draw → **stores on nearly every row, almost no hits**.
3. Under **48 concurrent HTTP requests**, each HFT doc has ~2001 shared rows → thousands of cached `[]byte` blobs pile up per request.
4. Heap profile confirms it: `drawSharedLayoutRow` **497 MB** and `preallocInlineTableRows` **592 MB** on the feature branch vs **46 MB** prealloc on master.

This matches the hang at ~50–65%: memory balloons, swap thrashes, all 48 server slots fill with stuck work, k6 VUs wait forever.

### Why Zerodha in-process bench looked fine

The Zerodha **in-process** benchmark showed lower heap per run and higher ops/s because it is sequential and single-process. The **HTTP k6** path exposes the global cache leak under concurrent load.

---

## Why master works

Master at `t2/gopdfsuit` has **no `sharedRowRenderCache`**. Same k6 harness, same machine:

- **674 req/s**
- **587 MB** heap in-use
- Full **35s** completion, 0% errors

Artifacts: `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260617_233247.txt`

---

## Recommended fixes

The optimization is sound for Zerodha micro-bench but broken for production/k6. Options:

1. **Remove `sharedRowRenderCache`** — safest; restores master k6 behavior.
2. **Fix the cache key** — key only on row content signature, not `mcidBase`/`y` (MCIDs must still be emitted per draw).
3. **Scope cache per-request** — attach cache to `PageManager` or `BorrowedPDF`, cleared when PDF is released.
4. **Add LRU + size cap** if keeping a global cache.

---

## Workarounds until fixed

```bash
# Completes on feature branch today
make bench-k6-light

# Full harness — use master or fix branch first
make bench-k6   # hangs/crashes on feat/optimization-5.5-medium
```

Run k6 in isolation (no parallel `run_all_benchmarks`, zerodha x10, Gotenberg). Reset WSL memory if swap is heavily used (`wsl --shutdown`).

---

## Artifacts referenced

| Run | Path | Notes |
|-----|------|-------|
| master k6 success | `t2/gopdfsuit/guides/cursor/baselines/gin_pprof_runs/k6_gin_20260617_233247.txt` | 674 req/s, complete |
| feature branch hang | `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260617_232525.txt` | frozen ~16k iter |
| feature branch light OK | `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260617_232133.txt` | 574 req/s, complete |
| feature branch heap | `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260617_232133.txt` | 1603 MB in-use |
| master heap | `t2/gopdfsuit/guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260617_233247.txt` | 587 MB in-use |

---

## Related docs

- `guides/optimizations/PR/20260617_zerodha_x10_pprof_pr_description.md` — optimization PR that introduced shared-row caching
- `guides/optimizations/20260617_zerodha_x10_pprof_optimization_checklist.md` — Zerodha validation checklist
- `guides/cursor/baselines/gin_pprof_runs/comparison_20260614.md` — historical k6 5-run baseline (825 req/s avg)
