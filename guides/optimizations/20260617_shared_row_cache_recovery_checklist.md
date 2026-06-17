# Shared row cache recovery checklist

**Date:** 2026-06-17
**Branch:** `feat/optimization-5.5-medium`
**Context:** `guides/optimizations/20260617_k6_bench_regression_analysis.md`

## Goal

Recover full `make bench-k6` stability without giving up the shared-layout row fast path that protects the high Zerodha in-process throughput target.

## Guardrails

- [x] Keep the shared-layout renderer (`drawSharedLayoutRow`, `drawSharedDeferRow`, shared column layouts, batched MCID writes, and row text-prefix reuse).
- [x] Remove the process-lifetime pointer-keyed cache behavior that grows with every decoded k6 request.
- [x] Keep the cache bounded by entries and retained bytes so memory plateaus under concurrent HTTP load.
- [x] Preserve the current in-process Zerodha throughput target of roughly 13,000 ops/sec before promoting the fix.
- [x] Use k6 completion and heap shape as release gates, not only micro-benchmark throughput.

## Implementation Checklist

- [x] Replace `sharedRowRenderCache sync.Map` with a bounded cache object.
- [x] Keep the original pointer/page/MCID/Y cache key so the in-process hot path keeps cheap lookup behavior.
- [x] Include output-affecting dynamic fields from the original key:
  - [x] page index
  - [x] MCID base
  - [x] current Y position
  - [x] row pointer
- [x] Copy cached byte slices on store so pooled row buffers cannot alias cached data.
- [x] Evict or clear when the cache exceeds a hard entry or byte cap.
- [x] Keep uncached shared rendering as the fallback path.

## Verification Checklist

- [x] Run focused Go tests for the PDF package.
- [x] Run the Zerodha in-process benchmark and compare against the ~13,000 ops/sec target.
- [x] Run `make bench-k6-light` first to verify the heap no longer balloons.
- [x] Run full `make bench-k6` only after the light run completes cleanly.
- [x] Confirm `drawSharedLayoutRow` no longer retains hundreds of MB in heap profiles.
- [x] Confirm k6 iteration count does not freeze and the server exits cleanly.

## Acceptance Criteria

- [x] Full k6 harness completes without WSL2 crash or server hang.
- [x] Heap in-use returns near master behavior instead of the 1.6 GB feature-branch profile.
- [x] The cache reaches a steady-state size during k6.
- [x] Zerodha in-process throughput remains within the accepted operating band for this branch.

## Validation Notes

- `GOCACHE=/tmp/gopdfsuit-go-test-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache go test ./internal/pdf` passed.
- `BENCH_ITERATIONS=5000 BENCH_WORKERS=48 make bench-gopdflib-zerodha` reached **12,675.43 ops/sec**.
- `make bench-k6-light` completed at **1,276.7 req/s**, **0% errors**, p99 **117.9 ms**, heap in-use **130.8 MB**.
- `drawSharedLayoutRow` retained **3.5 MB** flat in the light-run heap profile, down from the regression analysis' hundreds-of-MB failure mode.
- Full `make bench-k6` completed at **1,223.3 req/s**, **0% errors**, p99 **581.1 ms**, heap in-use **521.6 MB**.
- `drawSharedLayoutRow` retained **7.0 MB** flat in the full-run heap profile.
