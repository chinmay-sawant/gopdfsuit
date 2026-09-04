# Architecture - Deepening Opportunities

> **Parent:** `plans/reviews/` - architecture review, 2026-09-04, branch `chore/improves-fixes`
> **Status:** all phases implemented 2026-09-04 - zero `[ ]`, zero `[~]` - gates green (see Fix Session)
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

> Done 2026-09-04: decisions were deepen in all three cases; implementation
> landed in Phases 2/3/6 (A1 reader seam, B1+B2 owned types, E2 schema +
> golden check).

### 1.1 A1 one PDF-reader seam
- [x] `internal/pdf/merge/parser.go`, `compress/compress.go`,
  `redact/pdf_utils.go`, `form/xfdf.go`, `redact/encryption_inhouse.go`,
  `xref/xref.go` - grill constraints: which callers need byte-level
  fallbacks, who migrates first - proof: decision + rationale recorded below
- [x] A1 - record deepen / defer / ADR with pointer to implementation ledger
  (if deepened) - proof: row links the new ledger, no duplicate active rows

### 1.2 B1+B2 gopdflib owns types and validation
- [x] `pkg/gopdflib/*`, `bindings/python/cgo/exports.go`, `pypdfsuit/*.py` -
  grill constraints: API break policy for alias-to-struct migration, who owns
  each validation layer after - proof: decision + rationale recorded below
- [x] B1+B2 - record deepen / defer / ADR with pointer - proof: as above

### 1.3 E2 schema golden test
- [x] `internal/models/models.go`, `frontend/src/pages/Editor.jsx`,
  `components/editor/utils.js`, `pypdfsuit/types.py` - grill constraints:
  Go-models-authoritative vs generated-types, golden file location -
  proof: decision + rationale recorded below
- [x] E2 - record deepen / defer / ADR with pointer - proof: as above

## Phase 2: PDF ops seam decisions (A2, A3, A4)

> Done 2026-09-04: NEW `internal/pdf/pdfobj/` owns boundaries, trailer,
> xref-table/stream, object-stream expansion, capped inflation, widths
> validation. Migrated compress, redact, pdf helpers, merge (thin aliases
> kept). `form/xfdf.go` split into parse/locate/fill with locate-level
> test. `encryption/` owns dict parse + key derivation + ObjectView;
> `encryption_inhouse.go` demoted to adapter. Dead wrappers deleted;
> `make lint` green.

- [x] A2 inflation module - grill: placement next to `xref/` vs inside
  `merge/`, limit policy ownership - proof: decision recorded
- [x] A3 `form/xfdf.go:1726` three-way split - grill: middle field-locate
  seam shape, test strategy without full PDFs - proof: decision recorded
- [x] A4 encryption unification - grill: can `encryption/` absorb dict
  parsing without breaking redact's object-stream path - proof: decision
  recorded
- [x] Phase 2 rows point at implementation ledger(s) or ADRs - proof: no
  duplicate active work, pointers only

## Phase 3: Public surface decisions (B3, B4)

> Done 2026-09-04: gopdflib owned structs + `adapter.go` translation,
> sole-validator doctrine in `doc.go`; generic CGO adapter
> (`bytesOp`/`jsonOp`/helpers) + `CompressPDF` export (header kept in sync);
> NEW `font/provision.go` shared fetch/verify/cache wired into fontutils and
> pdfa download; `wasmcompress` via gopdflib (`GOOS=js GOARCH=wasm` build
> clean; ocr unix/other split fixed wasm+windows builds).

- [x] B3 generic CGO adapter - grill: one-line registration shape, compress
  as proof case, who owns `Free*` semantics - proof: decision recorded
- [x] B4 font-provision unification - grill: `pkg/fontutils` vs
  `internal/pdf/font` ownership, digest-pin policy sharing - proof: decision
  recorded
- [x] Phase 3 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 4: Request tier decisions (C1, C2, C3, C4)

> Done 2026-09-04: `handlers.go` thinned to wiring; new `request.go`
> (limits/errors/fetch/pprof helpers), per-op files, `decode.go` single
> `decodeTemplate`, `router.go` composition root (`main.go` ~35 lines).
> Service at 17 methods owns limits + error classification; redact via
> service with private `handle*` names. Benchmark layering verified clean
> (no `internal/pdf` imports in sampledata). All handler tests green with
> unchanged routes/codes.

- [x] C1 `PDFService` depth - grill: policy-behind-service vs collapse to
  narrower engine seam, mock strategy after - proof: decision recorded
- [x] C2 `handlers.go:910` split - grill: wiring vs helpers vs per-op files,
  composition root shape - proof: decision recorded
- [x] C3 single decode module - grill: tier policy + pool + HFT shell in one
  `decodeTemplate` call, perf parity proof required - proof: decision recorded
- [x] C4 redact seam - grill: extend service vs redact-owned interface,
  naming unification - proof: decision recorded
- [x] Phase 4 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 5: Engine ownership decisions (D1, D2, D3, D4)

> Done 2026-09-04: NEW `allocator.go` (IDs + xref offsets + ExtraObjects;
> 28 generator `setXrefOffset` sites + all reservations on it);
> `PageManager` facade over `pageObjectStore`/`pageLayoutStore`;
> registry `RefFor`/`Measure`/`EncodeForPDF` seam with utils shims; NEW
> `structure_tag.go` TableTagger (draw shared-row rewired) + NEW
> `destinations.go` DestinationStore (outline/links delegate). veraPDF
> 10/10 still green.

- [x] D1 allocator ownership - grill: `NextObjectID` + xref offsets +
  `ExtraObjects` in one module, migration of `layoutContentFontIDs` and
  `AssignObjectIDs` callers - proof: decision recorded
- [x] D2 `PageManager` split - grill: Pages vs Objects vs injected
  collaborators, facade or removal - proof: decision recorded
- [x] D3 single font seam - grill: `RefFor` / `Measure` / `Encode` on the
  registry, `utils.go` shim fate - proof: decision recorded
- [x] D4 `TableTagger` + `Destinations` - grill: BDC/MCID ownership,
  `DestKey` invariant home - proof: decision recorded
- [x] Phase 5 rows point at implementation ledger(s) or ADRs - proof: as above

## Phase 6: Frontend decisions (E1, E3, E4) and speculative (F1, F2, F3)

> Done 2026-09-04: `usePdfOperation` + `OperationShell` adopted by all 8
> pages; `template.schema.json` + `npm run test:schema` (11 checks pass);
> `compressLevels.js` single vocabulary/cap with WASM-first fallback;
> editor `documentModel.js` with E2-schema validation, `Editor.jsx`
> slimmed, UI unchanged. F1 landed as pdfobj table tests; F2 as WASM via
> gopdflib; F3 as `router.go` composition root. No ADRs needed - nothing
> rejected.

- [x] E1 `usePdfOperation` + `OperationShell` - grill: hook interface
  (endpoint + FormData builder per op), blob-URL lifecycle - proof: decision
  recorded
- [x] E3 one Compress interface - grill: level enum, single cap source,
  WASM/server adapter flag - proof: decision recorded
- [x] E4 editor document model - grill: typed builder + E2 schema
  validation, `Editor.jsx` slimming order - proof: decision recorded
- [x] F1 test-seam exports, F2 WASM via gopdflib, F3 composition root -
  demand proof before shaping: each needs a load-bearing reason or an ADR
  recording rejection - proof: reason or ADR pointer recorded
- [x] No second status document: approved deepenings get their own ledger
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

## Fix Session 2026-09-04 (evidence)

Five parallel implementers on disjoint file sets, lead reconciliation after:
dead-wrapper deletion (`helpers.go`, `parser.go`, `stream.go`,
`redact/helpers.go`, `encryption_inhouse.go`, `xfdf_locate.go`),
`pagemanager.go` facade unexported, pdfobj/decrypt `bytes.Contains`,
`font/metrics.go` + `font/registry.go` goconst constants, `.gitignore`
revert (unrelated), `libgopdfsuit.h` re-synced with the new CompressPDF
export. No commits were made (per instruction); everything sits in the
working tree on `chore/improves-fixes`.

Gates (all 2026-09-04):
- `gofmt -l` - clean; `go vet ./...` - clean
- `golangci-lint run -E revive,gocritic,gocyclo,goconst ./...` - exit 0,
  zero findings (was red on pre-existing goconst debt too - fixed as well)
- `go build ./...` - ok; `GOOS=windows go build ./...` - ok;
  `GOOS=js GOARCH=wasm go build ./cmd/wasmcompress/` - ok
- `go test ./...` - 18 packages ok, 0 FAIL
- Python `pytest bindings/python/tests` - 43 passed, 5 skipped
- `bash test/verify_pdfs.sh` - 10 passed, 0 failed (PDF/A-4 + PDF/UA-2)
- Frontend `npm run lint` - zero warnings; `npm run build` - success per
  implementer; `npm run test:schema` - 11 checks pass
