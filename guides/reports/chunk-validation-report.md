# Slopguard Chunk Validation — Domain-Agnostic Report

> **Date:** 2026-07-01
> **Findings evaluated:** 721
> **Method:** Manual review per `CHUNK_VALIDATOR.md` — pattern match only, no project context

---

## Overall Results

| Outcome | Count | Rate |
|---------|------:|-----:|
| **True positives** (rule correctly fired) | **624** | **86.5%** |
| **False positives** (rule misfired) | **97** | **13.5%** |
| **Total** | **721** | 100% |

**Verdict:** The detector suite is **86.5% accurate** on a project-agnostic basis. Roughly 1 in 7 findings is a false positive, driven by a small set of rules with systematic misclassification.

---

## Results by Review Batch

| Batch | Range | Count | TP | FP | FP Rate |
|-------|-------|------:|---:|---:|--------:|
| Agent 1 | 1–150 | 150 | 112 | 38 | 25.3% |
| Agent 2 | 151–300 | 150 | 142 | 8 | 5.3% |
| Agent 3 | 301–450 | 150 | 132 | 18 | 12.0% |
| Agent 4 | 451–600 | 150 | 134 | 16 | 10.7% |
| Agent 5 | 601–721 | 121 | 104 | 17 | 14.0% |

---

## False Positive Taxonomy (Domain-Agnostic)

| Category | Est. FPs | Affected rules | Description |
|----------|--------:|----------------|-------------|
| Loop over-triggering | ~28 | PERF-109 | Any `for` loop flagged without map-key recomputation |
| Non-error discard | ~24 | BP-1 | `_` binding on non-`error` return values |
| Hot-path misclassification | ~22 | PERF-31, PERF-151, PERF-35, PERF-41 | Startup/library/benchmark code flagged as request hot path |
| Pattern shape mismatch | ~12 | PERF-122, PERF-114, PERF-121, PERF-44 | Rule expects different syntactic pattern than present |
| Concurrency false alarms | ~7 | PERF-148, PERF-36, PERF-30 | Buffered channels, no capture, sync context use |
| Framework / ordering | ~4 | PERF-88, PERF-200, CWE-22, CWE-497 | Wrong framework, inverted logic, missed sanitization |

---

## Rule Reliability

### Rules with highest false-positive rate

| Rule | TP | FP | FP Rate | Issue |
|------|---:|---:|--------:|-------|
| PERF-109 | ~8 | ~22 | ~73% | Fires on loop header without map-key evidence |
| PERF-31 | ~11 | ~11 | ~50% | Any `func (` receiver treated as HTTP handler |
| PERF-122 | 0 | 3 | 100% | Expects trim; misfires on `HasPrefix` + slice |
| PERF-148 | 0 | 3 | 100% | Ignores channel buffer capacity |
| PERF-151 | 0 | 3 | 100% | Flags `main`/constructors as non-inlinable handlers |
| PERF-36 | 0 | 2 | 100% | `for range N` has no variable capture |

### Rules with highest true-positive rate (≥10 hits, ≤5% FP)

| Rule | Est. TP | Role |
|------|--------:|------|
| PERF-6 | ~79 | `fmt.Sprintf`/`Fprintf` inside loops |
| PERF-192 | ~78 | `make(map)` without capacity hint |
| BP-1 | ~140 | Genuine discarded `error` returns |
| PERF-107 | ~40 | `binary.Read`/`Write` inside loops |
| PERF-1 | ~40 | `regexp.MustCompile` inside loops |
| PERF-32 | ~37 | String↔`[]byte` conversion |
| PERF-15 | ~17 | `strconv.Itoa` inside loops |
| PERF-188 | ~16 | `fmt.Sscanf` inside parsing loops |
| BP-2 | ~9 | Bare `return err` without wrapping |

---

## Detection Accuracy by Rule Category

| Category | Est. TP share | Est. FP share | Accuracy |
|----------|-------------:|-------------:|---------:|
| Loop allocations (PERF-3/4/6/15/45) | ~22% | ~3% | ~95% |
| Map/slice hints (PERF-192/123) | ~15% | ~1% | ~98% |
| Error handling (BP-1/2/5) | ~28% | ~18% | ~85% |
| Hot-path heuristics (PERF-31/35/151) | ~8% | ~25% | ~45% |
| Loop-body heuristics (PERF-109/119/128) | ~5% | ~20% | ~40% |
| Security (CWE-*) | ~1% | ~1% | ~90% |

---

## Detail Reports

| File | Range |
|------|-------|
| [`agent-1-findings-001-150.md`](chunk-validation/agent-1-findings-001-150.md) | 1–150 |
| [`agent-2-findings-151-300.md`](chunk-validation/agent-2-findings-151-300.md) | 151–300 |
| [`agent-3-findings-301-450.md`](chunk-validation/agent-3-findings-301-450.md) | 301–450 |
| [`agent-4-findings-451-600.md`](chunk-validation/agent-4-findings-451-600.md) | 451–600 |
| [`agent-5-findings-601-721.md`](chunk-validation/agent-5-findings-601-721.md) | 601–721 |