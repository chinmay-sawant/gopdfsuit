---
name: phase-wise-checklist
description: Create or maintain evidence-backed, phase-wise implementation checklists in CodeHound plans. Use when a task asks to plan work by phase, update checklist status while implementing, reconcile shipped work with plan rows, or produce one canonical execution ledger.
---

# Phase-Wise Checklist

Create one canonical plan under `plans/` and treat it as a live execution ledger. For gopdfsuit this usually means one ledger per PDF op or frontend area, with fixtures in `sampledata/<area>/`.

## Workflow

1. Read the named parent plan/checklist and inspect the current implementation before changing any status. For engine work read `internal/pdf/` plus `pkg/gopdflib/` first. For API work read `internal/handlers/` plus `internal/models/`.
2. Use the repository graph tools for code discovery when available; otherwise use focused source search. Treat plan prose as a claim, not evidence.
3. Start with the required plan shape:

   ```markdown
   # <Hierarchy> - <Title>

   > **Parent:** `<path>` - reference
   > **Status:** current state
   > **Estimated effort:** estimate

   ---

   ## Overview
   ## Executive Summary
   ## Phase 1: <Area>
   ### 1.1 <Slice>
   - [ ] Verifiable action
   ## Dependencies
   ```

4. Make every checklist row atomic: one code change or one validation result. Include the affected path, rule/issue ID, expected behavior, and required proof where useful. Example: `internal/pdf/compress/* - Heavy tier keeps text sharp - proof: make test-verify-pdfs`.
5. Order phases by dependency and risk: correctness/security first, then API/data contracts, then performance/cleanup, then closure gates. For PDF work put correctness plus PDF/A-4 and PDF/UA-2 compliance before perf tiers.
6. Use statuses consistently:
   - `[ ]` not started or not proven;
   - `[x]` implemented and validated with current evidence;
   - `[~]` intentionally deferred/partial, with reason, owner boundary, and next gate.
7. Update a row only after the matching source/test/benchmark check succeeds. Record the command and outcome beside closure gates; keep release measurements distinct from dev-loop measurements.
8. If a row is moved to a newer canonical ledger, rewrite the source row as `[~]` with an explicit pointer. Do not leave duplicate active work in multiple files.
9. Preserve unrelated worktree changes. Do not close a security, performance, or migration item merely because a plan says it was done.

## Evidence Rules

- Prefer current code, tests, benchmarks, scanner output, and CI configuration over historical notes.
- State negative results precisely: e.g. "no production `unsafe` beyond X found in this audit", never "no bugs exist".
- Keep risk statements separate from confirmed defects. Label hypotheses and give their validation step.
- For performance items, include the exact release command, dataset/path, cold/warm cache state, and metric; successful execution is not benchmark proof. Use `sampledata/gopdflib/zerodha/` or `sampledata/benchmarks/` paths when relevant.
- For PDF output items, cite `test/verify_pdfs.sh` output plus veraPDF plus structure tree check, not just `go test`.

## Completion Handoff

Before declaring a phase complete, confirm its rows, run the smallest relevant validation, synchronize the checklist, and provide a concise result plus the next unchecked phase. Do not create a second status document unless it is explicitly a deferred ledger with a pointer from the canonical plan.

## Required Checks

- For documentation-only changes (`guides/`, `plans/`, `*.md`), do not run lint or test checks.
- For every non-documentation change, run `make fmt && make lint && make test` before marking the phase complete. Add `make test-integration` when `internal/handlers/` or `internal/pdf/` changed. Add `make test-verify-pdfs` when PDF bytes changed. Add `cd frontend && npm run build` when `frontend/src/` changed. Record all outcomes in the canonical checklist; leave the row unchecked if any command fails.
