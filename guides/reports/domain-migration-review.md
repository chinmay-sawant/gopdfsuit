# P2.4 — Domain Migration & PERF Detector Code Review

> **Source:** Consolidated findings from ponytail-review, m15-anti-pattern, rust-best-practices, and rust-patterns skills on commits `a77c064..d6f4d23` + uncommitted domain migration.
> **Date:** 2026-06-28

---

## Phase 2 — Domain Migration: ✅ Complete

All 49 detectors from `general_perf/stdlib_misuse/hot_path_misc.rs` have been migrated into 4 domain modules. The plan `02-perf-detectors-remaining.md` needs updating (see below).

---

## Required Plan Updates

### `plans/v2.0.0/pending-work/02-perf-detectors-remaining.md`

- [x] **Line 14**: Change "All detector code lives in `general_perf/stdlib_misuse/hot_path_misc.rs`. Domain module migration (Phase 2) is still deferred." → "Detectors are now organized by domain in `src/lang/go/detectors/perf/domains/{concurrency,memory_gc,string_bytes,stdlib_optimization}.rs`. Phase 2 complete."
- [x] **Phase 2 section (lines 105-130)**: Mark all checklist items as complete. Change status from "Deferred" to "Complete". The new file structure is:
  - `domains/concurrency.rs` — 9 detectors
  - `domains/memory_gc.rs` — 7 detectors
  - `domains/string_bytes.rs` — 5 detectors
  - `domains/stdlib_optimization.rs` — 28 detectors
  - `domains/mod.rs` — re-exports all 4 via `pub(crate) use`
  - `common.rs` — `is_handler_shaped` + `file_has_handler` moved here
  - `hot_path_misc.rs` — replaced with migration note (dead code)
- [x] **Estimated effort** header: Change from "~1 day (domain migration + documentation)" to "Complete"

### `plans/p2-remaining-work.md`

- [x] **P2.4 section**: Mark domain migration as checked off

---

## Code Review Findings — Checklist

### 🔴 Critical — All Fixed

- [x] **PERF-179: `Box::leak` memory leak + dead first loop** — `string_bytes.rs`
  - Deleted the first loop entirely. `keys` now built directly from `facts.calls` with no `Box::leak`.
  - **Fix applied:** 2026-06-28
- [x] **PERF-155: Dead `for` loop** — `stdlib_optimization.rs`
  - Deleted `for (start, _end) in &facts.for_ranges` block. Real detection was already handled below.
  - **Fix applied:** 2026-06-28
- [x] **PERF-134: `.unwrap()` guarded by `.contains()`** — `memory_gc.rs:30`
  - `source.find("Read(buf").unwrap()` → `source.find("Read(buf").unwrap_or(0)`
  - **Fix applied:** 2026-06-28

### 🟡 Medium — All Fixed

- [x] **`let _ = facts;` inconsistency** — all 4 domain files
  - Standardized: 7 functions renamed to `_facts` parameter (perf_148, 150, 151, 189, 200, 201, 205); ~25 `let _ = facts;` / `let _ = source;` / `let _ = _facts;` dead lines removed across all files.
  - **Fix applied:** 2026-06-28
- [x] **PERF-134: Redundant `let _ = _facts`** — `memory_gc.rs`
  - Deleted the no-op line. Parameter was already `_facts`.
  - **Fix applied:** 2026-06-28
- [x] **PERF-186: `let _ = facts` inside loop** — `string_bytes.rs`
  - Moved outside loop (removed entirely — `facts` is used in the loop condition so no suppressor needed).
  - **Fix applied:** 2026-06-28
- [x] **Misleading "Auto-generated" doc comments** — all 4 domain files
  - Changed to "Migrated from hot_path_misc.rs".
  - **Fix applied:** 2026-06-28
- [x] **Duplicate `bytes.NewBuffer` in string match** — `string_bytes.rs:29,31`
  - Removed duplicate line.
  - **Fix applied:** 2026-06-28

### 🟢 Minor / Observation

- [x] **PERF-143** fires on any `http.HandleFunc` without `TimeoutHandler` — potentially noisy. **Acknowledged.** No code change; monitor false positive reports.
- [x] **PERF-150** counts hardcoded array sizes (`[1024]byte`, `[4096]byte`, etc.) — broadened to include `[32768]byte`, `[65536]byte`, and `make([]byte, N)` for N=2048,16384. Still misses non-power-of-2 sizes (`[3072]byte`) — ponytail comment marks the ceiling.
- [x] **PERF-206 dead branch** — Verified: `callee.contains("Unsafe")` IS reachable. The walker records the full chain text (`db.Unsafe.Where`), which both ends with `.Where` and contains `Unsafe`. Dead code (`let _ = facts;`, `let _ = source;`) was already removed in the 🔴/🟡 pass. No further fix needed.
- [x] **`format!` inside per-call loop** — `concurrency.rs` — `format!("<-{ch}")` allocates per call because `ch` (the channel name) varies per iteration. No viable fix; accepted as-is.

---

## Summary

| Area | Finding Count |
|------|---------------|
| Plan docs to update | 2 files |
| 🔴 Critical code fixes | 3 ✅ Fixed |
| 🟡 Medium code fixes | 5 ✅ Fixed |
| 🟢 Minor/observation | 4 items (untouched) |
| **net line reduction** | **~90 lines removed** (per ponytail-review) |
