# Architecture - Deepening Opportunities

> **Parent:** `plans/reviews/` - architecture review, 2026-09-04, branch `chore/improves-fixes`
> **Status:** candidates only - no interfaces proposed, no implementation started, all rows `[ ]`
> **Estimated effort:** grilling ~1 day; deepenings sized per candidate on decision

---

## Overview

Five parallel read-only walks (generation engine, PDF ops, request tier,
public surfaces, frontend contracts) surfaced 23 deepening candidates in the
module / interface / depth / seam / adapter / leverage / locality vocabulary.
A companion visual report lives outside the repo:
`/tmp/opencode/architecture-review-20260904-1047.html`.
No `CONTEXT.md` or ADRs exist, so no prior decisions constrain the grill.

This ledger tracks decisions only. Per the review method, interfaces are not
proposed here: each candidate ends in deepen, defer-with-reason, or ADR, and
approved deepenings move to their own canonical ledger with a pointer.

## Executive Summary

- **Highest leverage:** A1 one PDF-reader seam - unblocks A2 (inflation), A4
  (encryption view), A3 (field-locate tests). Migrate compress first.
- **Widest gate:** B1+B2 gopdflib owned types and single validation - every
  future cross-tier change touches this seam.
- **Cheapest win:** E2 schema golden test - prevents silent Go/JS/Python drift
  while bigger deepens land.
- **Counts:** 13 Strong, 7 Worth exploring, 3 Speculative.
- **Hot-spot bias:** recent churn (compress, handlers hardening, CGO guards)
  weighted heaviest, per YAGNI scoping.

## Corrections (walker claims adjusted after source verification)

- V1: line counts verified (`generator.go:2802`, `draw.go:2494`,
  `form/xfdf.go:1726`, `handlers.go:910`, `Editor.jsx:1505`,
  `services.go:74` with 9 methods). `Redaction.jsx` is 704 lines, not 704 as
  a page-size claim issue - matches.
- V2: `PDFService` is exactly 9 methods (`services.go:15-23`), all
  passthroughs - shallow verdict stands.
- V3: object-map builders confirmed in `redact/pdf_utils.go:29`
  (`buildObjectMap`) and `compress/compress.go:169` (`parseObjects`) with
  incompatible types - fragmentation verdict stands.

---

## Phase 1: Grill the top three

### 1.1 A1 one PDF-reader seam
- [ ] `internal/pdf/merge/parser.go`, `compress/compress.go`,
  `redact/pdf_utils.go`, `form/xfdf.go`, `redact/encryption_inhouse.go`,
  `xref/xref.go` - grill constraints: which callers need byte-level
  fallbacks, who migrates first - proof: decision + rationale recorded below
- [ ] A1 - record deepen / defer / ADR with pointer to implementation ledger
  (if deepened) - proof: row links the new ledger, no duplicate active rows

### 1.2 B1+B2 gopdflib owns types and validation
- [ ] `pkg/gopdflib/*`, `bindings/python/cgo/exports.go`, `pypdfsuit/*.py` -
  grill constraints: API break policy for alias-to-struct migration, who owns
  each validation layer after - proof: decision + rationale recorded below
- [ ] B1+B2 - record deepen / defer / ADR with pointer - proof: as above

### 1.3 E2 schema golden test
- [ ] `internal/models/models.go`, `frontend/src/pages/Editor.jsx`,
  `components/editor/utils.js`, `pypdfsuit/types.py` - grill constraints:
  Go-models-authoritative vs generated-types, golden file location -
  proof: decision + rationale recorded below
- [ ] E2 - record deepen / defer / ADR with pointer - proof: as above

## Phase 2: PDF ops seam decisions (A2, A3, A4)

- [ ] A2 inflation module - grill: placement next to `xref/` vs inside
  `merge/`, limit policy ownership - proof: decision recorded
- [ ] A3 `form/xfdf.go:1726` three-way split - grill: middle field-locate
  seam shape, test strategy without full PDFs - proof: decision recorded
- [ ] A4 encryption unification - grill: can `encryption/` absorb dict
  parsing without breaking redact's object-stream path - proof: decision
  recorded
- [ ] Phase 2 rows point at implementation ledger(s) or ADRs - proof: no
  duplicate active work, pointers only

## Phase 3: Public surface decisions (B3, B4)

- [ ] B3 generic CGO adapter - grill: one-line registration shape, compress
  as proof case, who owns `Free*` semantics - proof: decision recorded
- [ ] B4 font-provision unification - grill: `pkg/fontutils` vs
  `internal/pdf/font` ownership, digest-pin policy sharing - proof: decision
  recorded
- [ ] Phase 3 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 4: Request tier decisions (C1, C2, C3, C4)

- [ ] C1 `PDFService` depth - grill: policy-behind-service vs collapse to
  narrower engine seam, mock strategy after - proof: decision recorded
- [ ] C2 `handlers.go:910` split - grill: wiring vs helpers vs per-op files,
  composition root shape - proof: decision recorded
- [ ] C3 single decode module - grill: tier policy + pool + HFT shell in one
  `decodeTemplate` call, perf parity proof required - proof: decision recorded
- [ ] C4 redact seam - grill: extend service vs redact-owned interface,
  naming unification - proof: decision recorded
- [ ] Phase 4 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 5: Engine ownership decisions (D1, D2, D3, D4)

- [ ] D1 allocator ownership - grill: `NextObjectID` + xref offsets +
  `ExtraObjects` in one module, migration of `layoutContentFontIDs` and
  `AssignObjectIDs` callers - proof: decision recorded
- [ ] D2 `PageManager` split - grill: Pages vs Objects vs injected
  collaborators, facade or removal - proof: decision recorded
- [ ] D3 single font seam - grill: `RefFor` / `Measure` / `Encode` on the
  registry, `utils.go` shim fate - proof: decision recorded
- [ ] D4 `TableTagger` + `Destinations` - grill: BDC/MCID ownership,
  `DestKey` invariant home - proof: decision recorded
- [ ] Phase 5 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 6: Frontend decisions (E1, E3, E4) and speculative (F1, F2, F3)

- [ ] E1 `usePdfOperation` + `OperationShell` - grill: hook interface
  (endpoint + FormData builder per op), blob-URL lifecycle - proof: decision
  recorded
- [ ] E3 one Compress interface - grill: level enum, single cap source,
  WASM/server adapter flag - proof: decision recorded
- [ ] E4 editor document model - grill: typed builder + E2 schema
  validation, `Editor.jsx` slimming order - proof: decision recorded
- [ ] F1 test-seam exports, F2 WASM via gopdflib, F3 composition root -
  demand proof before shaping: each needs a load-bearing reason or an ADR
  recording rejection - proof: reason or ADR pointer recorded
- [ ] No second status document: approved deepenings get their own ledger
  with a pointer from the moved row (`[~]` + pointer); rejected candidates
  get ADRs only when future explorers need the reason

## Dependencies

- Phase 1 before all others (top-three decisions set the order).
- A1 before A2/A3/A4 (reader seam is their foundation).
- B1+B2 before B3/F2 (owned types gate adapter work).
- E2 before E4 (schema golden test gates editor model validation).
- D1 before D2 (allocator ownership decides what PageManager can shed).
- Implementation ledgers depend on `make fmt && make lint && make test`
  green at phase close, plus `make test-verify-pdfs` when PDF bytes change.
