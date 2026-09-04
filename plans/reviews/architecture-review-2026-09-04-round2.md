# Architecture - Residual Friction (Round 2)

> **Parent:** `plans/reviews/` - follow-up to `architecture-review-2026-09-04.md`, 2026-09-04, branch `chore/improves-fixes`
> **Status:** grilled 2026-09-04 - 10 Strong deepen, 13 Worth exploring triaged, decisions below - no implementation started
> **Estimated effort:** grilling ~1 day; deepenings sized per candidate on decision

---

## Overview

Five follow-up walks (new-seam consolidation, errors/observability,
concurrency/perf, tests/contracts, frontend residuals) over the post-deepening
tree. Prior ledgers (`go-review-2026-09-04.md`,
`architecture-review-2026-09-04.md`) were read first; nothing below repeats
them. Companion visual report: `architecture-review-2026-09-04-round2.html`
in the same folder.

Theme: round 1 created depth - this round gives the new seams sole ownership,
one error language, one cache doctrine, and tests at the new seams.

## Executive Summary

- **Highest leverage:** B1+B2 error taxonomy + envelope - one change reused
  across HTTP, CGO, WASM, Python; hardens 413/422/500 into observable codes.
- **Ownership completion:** A2 allocator bypass + A1 pdfobj/xref merge finish
  the D1/A1 stories; C2 kills the only true growth risk (unbounded
  templateDataCache).
- **Cheapest win:** D1 schema triple-gate - prevents silent drift.
- **Counts:** 10 Strong, 13 Worth exploring.
- Still no `CONTEXT.md`/ADRs; the review ledgers remain the de-facto log.

## Corrections (walker claims adjusted after source verification)

- V1: `xref.go` is 103 lines write-only, `usePdfOperation.js` 159 lines,
  zero `func Fuzz` repo-wide, no `request/decode/router` test files -
  all as claimed.
- V2: `metadata.go:317,355,382` direct `NextObjectID++` confirmed;
  `mustToInternal` panics (`adapter.go:42-56`) confirmed - unlogged crash
  path is real.
- V3: Editor double Toast blocks (`:1199`, `:1276`), Viewer stale
  `loadTemplate`, Redaction dead `canvasRef`, `mapLevel` leftover -
  confirmed.
- V4: `templateDataCache` unbounded `sync.Map` confirmed;
  `rgbDataPool.Put` unconditional confirmed.

---

## Phase 1: Grill error language (B1, B2)

### 1.1 Taxonomy before envelope
- [ ] `internal/handlers/request.go:62-77`, `internal/pdf/font/*`,
  `internal/pdf/font/provision.go:18-27` - grill: sentinel set
  (`ErrInvalidInput`, `ErrLimitExceeded`, `ErrUpstream`), `%w` adoption
  order, substring list as pinned fallback - proof: decision + rationale
  recorded below
- [ ] B1 - record deepen / defer / ADR with pointer - proof: row links the
  new ledger, no duplicate active rows

### 1.2 Adapter envelope
- [ ] `bindings/python/cgo/exports.go`, `pypdfsuit/_bindings.py`,
  `cmd/wasmcompress/main.go`, `apiConfig.js` - grill: `{code,message}`
  shape, Python subclasses, contract doc home (gopdflib per sole-validator
  doctrine) - proof: decision + rationale recorded below
- [ ] B2 - record deepen / defer / ADR with pointer - proof: as above

## Phase 2: Sole ownership for new seams (A1, A2, A3, A4)

- [ ] A1 `pdfobj/` vs `xref/` - grill: move `WriteCompactXRef` into pdfobj
  vs rename to `xrefwrite`, Style preset home - proof: decision recorded
- [ ] A2 allocator bypass - grill: store + destinations through bound
  `Allocator`, metadata `Alloc` migration, `TestShiftedLayoutLargeDoc` as
  proof gate - proof: decision recorded
- [ ] A3 facade commit-or-collapse - grill: held values vs plain methods,
  field unexporting order - proof: decision recorded
- [ ] A4 tagger/destinations completion - grill: unexport table primitives,
  `DestKey` move (count outside writers first), link emission fold -
  proof: decision recorded
- [ ] Phase 2 rows point at implementation ledger(s) or ADRs - proof: no
  duplicate active work, pointers only

## Phase 3: Caches, budgets, capabilities (C1, C2, C3, C4, C5)

- [ ] C1 pool policy - grill: `poolpolicy` module shape, gauge home,
  rgbDataPool cap first - proof: decision recorded
- [ ] C2 bounded caches - grill: single LRU interface, templateDataCache
  cap + key, eviction counters - proof: decision recorded
- [ ] C3 admission budget - grill: single goroutine budget, 429/Retry-After,
  depth metrics; load test decides tuning - proof: decision recorded
- [ ] C4 WASM serial path - grill: build-tag vs `GOMAXPROCS==1` selection,
  parity benchmark - proof: decision recorded
- [ ] C5 race handles - grill: typed-handle pattern extension,
  atomic-around-clear, `-race` test set - proof: decision recorded
- [ ] Phase 3 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 4: Contracts with teeth (D1, D2, D3, D4, D5)

- [ ] D1 schema triple-gate - grill: Go golden test shape, Python check on
  same file, CI wiring (`make test` vs workflow) - proof: decision recorded
- [ ] D2 real fuzz - grill: `FuzzConvertSVG` entry + corpus path, extension
  order (page-spec, aliasing), seeds as regression layer - proof: decision
  recorded
- [ ] D3 parity tests - grill: JS mapping tests (needs first frontend unit
  harness), Go default-parity test, fallback mocking - proof: decision
  recorded
- [ ] D4 bench map - grill: harness-to-question mapping doc, canonical
  payload pin, `make bench-smoke` slice - proof: decision recorded
- [ ] D5 seam tests - grill: `decodeTemplate` + route-table adapter tests,
  mock router redact extension - proof: decision recorded
- [ ] Phase 4 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 5: Observability seam (B3, B4)

- [ ] B3 logging seam - grill: request-ID middleware, per-request line
  fields, engine-noise removal, frontend onError-state adoption -
  proof: decision recorded
- [ ] B4 metrics decorator - grill: `PDFService` decorator vs gin
  middleware, stdlib counters first, otel later - proof: decision recorded
- [ ] Phase 5 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 6: Frontend residuals (E1, E2, E3, E4, E5)

- [ ] E1 Editor adoption - grill: generate/load/upload through hook,
  Toast dedup, id-scheme move into documentModel - proof: decision recorded
- [ ] E2 Viewer collapse - grill: overlay slot in OperationShell, explicit
  filename param (fixes stale-state bug) - proof: decision recorded
- [ ] E3 Redaction completion - grill first: banner-vs-alert ownership,
  then adopt run/storeBlob, delete readErrorResponse + canvasRef -
  proof: decision recorded
- [ ] E4 hook split - grill: shared `useToast` seam, hook narrowed to
  request + lifecycle, status-code classification - proof: decision recorded
- [ ] E5 transport seam - grill: flag + override + fallback notice (cleanup
  implementable now); fallback-consent is product decision - proof:
  decision recorded, consent owner named
- [ ] No second status document: approved deepenings get their own ledger
  with a pointer from the moved row (`[~]` + pointer); rejected candidates
  get ADRs only when future explorers need the reason

## Grill Decisions 2026-09-04 (5 parallel walkers, source-verified)

- B1 deepen M: `pkg/gopdflib/errors.go` sentinels (ErrInvalidInput, ErrLimitExceeded, ErrUpstream) + `%w`, ClassifyError via errors.Is/As, substring list as pinned fallback.
- B2 deepen M (after B1): `{code,message}` envelope at gopdflib seam, CGO keeps ABI with JSON envelope, Python subclasses, same codes in WASM/HTTP.
- A1 deepen M reshaped: WriteCompactXRef into pdfobj + Style presets, delete `xref/`, migrate 5 inline writers (generator first), gate `make test-verify-pdfs`.
- A2 deepen M (before C2/C5): pageObjectStore holds bound Allocator, PDFAHandler Alloc migration, DestinationStore via Allocator, EnsureBeyond for generator fixups.
- A3 defer: adapters not walls, revisit after A2. A4 deepen M sequenced: encryptor thread, DestKey into store, unexport after tests migrate.
- C1 defer module, deepen rgb cap S. C2 deepen templateDataCache S (cap + mtime/size + eviction counter), defer unified LRU. C3 deepen admission 429 S + ADR rejecting single budget. C4 defer (benchmark-gated). C5 narrow deepen M (atomic clear-all x2 + 3 race tests + make wiring).
- D1 deepen S (Go golden + Python check + CI wire). D2 deepen S (FuzzConvertSVG + corpus). D3 deepen phased S now (Go parity + delete mapLevel), JS tests on E5 harness. D4 deepen docs-only S (BENCHMARKS map + bench-smoke). D5 deepen narrow S (decode + route-table + redact mock test).
- B3 deepen S (after B1): request-ID + one structured line, engine noise removal, must* to errors. B4 defer to C3 tuning.
- E1 deepen M (via hook, needs E4 toast first). E2 deepen S (explicit filename param, fixes stale bug). E3 deepen S (banners win, run/storeBlob, delete readErrorResponse + canvasRef). E4 deepen S narrowed (useToast split, status codes). E5 deepen cleanup S now (flag + notice + consent click), product default owner: chinmay-sawant.

## Dependencies

- Phase 1 before B3/B4 (taxonomy first, logging/metrics ride it).
- Phase 2 A2 before C2/C5 (allocator ownership decides cache-handle shape).
- D1 before E-phase schema validation work; D3 needs the first frontend
  unit harness (E5 test setup can host it).
- E3 banner-vs-alert decision before E4 `useToast` shape.
- Implementation ledgers depend on `make fmt && make lint && make test`
  green at phase close, plus `make test-verify-pdfs` when PDF bytes change
  and `npm run lint && npm run build` for frontend work.
