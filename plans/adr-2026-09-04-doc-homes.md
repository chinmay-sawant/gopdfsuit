# ADR 2026-09-04: Doc Homes (documentation vs guides vs docs vs plans)

## Status

Accepted 2026-09-04.

## Context

Four doc homes grew in parallel: `documentation/`, `guides/`, `docs/`, and
`plans/`. Writers guessed where new prose belongs, stale guides contradicted
current behavior, and reviews re-litigated placement every round.

## Decision

| Home | Role | Rule |
|------|------|------|
| `documentation/` | Truth | Current operational docs live here. Edit on behavior change. |
| `guides/` | Frozen archive | Historical notes. Do not update content; a header marks each file frozen. Pointers to `documentation/` allowed. |
| `docs/` | Generated | Vite build output from `frontend/`, auto-committed by CI. Never hand-edit. |
| `plans/` | Decisions | Ledgers, ADRs, checklists. `plans/INDEX.md` is the map; one active ledger per topic. |

## Consequences

- `guides/MAKEFILE.md` is frozen; the makefile tiers are documented in the
  makefile itself (`make bench-help`) and `documentation/`.
- New behavior docs go to `documentation/`; new decisions go to `plans/`
  with an `plans/INDEX.md` entry.
- Link checks treat `guides/` as archive: links may rot, truth must not.
