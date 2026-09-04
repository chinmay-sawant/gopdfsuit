# Plans Index

Single map of every decision ledger under `plans/`. Closed rows stay closed;
only one ledger per topic is active at a time. Do not open a duplicate active
row that another ledger already owns (see Round 2 pointer below).

Doc-home rule (see `plans/adr-2026-09-04-doc-homes.md`): `plans/` holds
decisions, `documentation/` holds truth, `guides/` is a frozen archive,
`docs/` is generated Vite output.

## Closed (reference only, no new rows)

| Ledger | Status |
|--------|--------|
| `plans/reviews/architecture-review-2026-09-04.md` (Round 1) | Closed, all phases implemented 2026-09-04 |
| `plans/reviews/go-review-2026-09-04.md` (Go sweep) | Closed, all rows fixed 2026-09-04 |
| `plans/adr-2026-09-04-c3-single-budget-rejected.md` | Closed ADR, keep 3 separate budgets |
| `plans/wasm/01-full-wasm-port.md` | Closed except HTML-URL pointer to `02-gowkhtmltopdf-replace.md` |
| `plans/wasm/02-gowkhtmltopdf-replace.md` | Closed, gates recorded 2026-09-04 |
| `plans/wasm/02b-html-bench-prep.md` | Closed bench-prep appendix of `02-gowkhtmltopdf-replace.md` (renamed from `02-html-bench-prep.md` to end the `02-` prefix collision) |
| `plans/PR/pr-frontend-auto-build-push-only.md` | Closed, push-only auto-build survivor (supersedes deleted `pr-frontend-auto-build.md`) |

## Active (new work goes here)

| Ledger | Status |
|--------|--------|
| `plans/reviews/architecture-review-2026-09-04-round2.md` | Open, owns error-taxonomy/envelope, allocator bypass, pdfobj/xref merge, cache doctrine. No duplicate rows elsewhere. |
| `plans/builder-snippets/reviews/architecture-review-2026-09-04-builder-snippets.md` | Active, this track (`feat/builder-snippets`). Phases 1-6 per ledger. |
| `plans/builder-snippets/plan.md` | Active parent track plan |
| `plans/gopdflib/fluent-builder.md` | Active unless its parent track closed it |
| `plans/wasm/03-wasm-everywhere-noauth-editor.md` | Active unless its parent track closed it |
| `plans/wasm/04-frontend-wasm-split-fonts-compliance.md` | Active unless its parent track closed it |
| `plans/ADR-compress-fallback-default.md` | Active unless superseded |
| `plans/pdf-compress-nogs.md` | Active unless superseded |

## Gate

Grep for duplicate active rows before adding a new ledger:

```sh
grep -rn "\[ \]" plans/ --include="*.md" | grep -v builder-snippets/reviews | grep -v round2
```
