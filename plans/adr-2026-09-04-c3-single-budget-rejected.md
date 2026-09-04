# ADR: Reject single unified concurrency budget (C3, round 2)

Date: 2026-09-04. Branch: chore/improves-fixes.

Decision: keep three separate admission budgets instead of one global budget.

- HTTP admission (`concurrencyLimiter` in `internal/handlers/router.go`):
  non-blocking try-acquire, 429 + `Retry-After: 1` on full, atomic depth
  gauge. This is the only layer that sheds user-visible load.
- Page-compress parallelism (`errgroup.SetLimit` in
  `internal/pdf/generator.go`): inner CPU budget for flate work per request.
- Signature slots (`signWorkerSlots` in
  `internal/pdf/signature/signature.go`): inner latency budget for slow
  signing ops.

Why not one budget: the three guard different resources (in-flight requests
vs CPU-bound compress workers vs blocking sign operations) with different
hold times. A single semaphore would couple slow signing to HTTP admission
and either starve fast paths or admit too much flate work. Tuning stays
per-layer; load tests decide values. B4 metrics ride the per-layer gauges
(`ConcurrencyDepth`, plus existing counters) rather than a merged number.
