## Summary

- Restrict the `build-and-commit-frontend` job to `push` events on `master`/`main` (plus manual `workflow_dispatch`), so feature PRs no longer open `chore/auto-build-frontend` PRs against `master`.

---

## Motivation / context

- Plans: `plans/...` (n/a, CI-only fix)
- Issues: see **Related issues**
- PR #207 (`chore: auto-build frontend`, bot author, only `docs/compress.wasm`) is the second iteration of a self-triggering loop: every `pull_request` run rebuilt `docs/` and proposed it to `master`, and every merge of that docs PR (a `push` to `master`) rebuilt again because `compress.wasm` is non-deterministic.

---

## Changes

### Area 1 - `build-and-commit-frontend` trigger

- Replaced the `pull_request || push (!tag) || dispatch` guard with push-to-master-only plus bot-actor guard:
  `github.actor != 'github-actions[bot]' && ((push && (ref == master || ref == main)) || (dispatch && full-ci))`.
- Simplified checkout `ref` to `github.ref_name` and create-pull-request `base` to `master` (no more `head_ref`/`base_ref` PR logic).

### Area 2 - docs path loop guard

- Added `paths-ignore: docs/**` to both `push` and `pull_request` triggers, so merging the bot's own `docs/` PR does not trigger another full rebuild.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Fewer CI runs: no frontend rebuild per feature PR; one rebuild per real `master` merge |
| **Memory** | None |
| **Behavior / correctness** | Feature PRs run lint plus test only; `docs/` rebuild proposed once after `master` merges with frontend/WASM changes |
| **API (`/api/v1/*`) / UI** | None; build steps (`make wasm-compress`, `npm run build` Cloud Run env) unchanged |
| **Dependencies** | None |
| **Binary size / build time** | None |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | None |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [x] Workflow YAML parses (`yaml.safe_load`) and `build-and-commit-frontend.if` verified
- [ ] CI on this PR: Backend Lint, Backend Test, Frontend Lint pass
- [ ] After merge to `master`: exactly one `chore: auto-build frontend` PR only when the build produces a real `docs/` diff; merging it opens nothing further

### Commands

```sh
make fmt && make lint
make test
# plus when relevant:
make test-integration
make test-verify-pdfs
```

---

## Screenshots / sample output

```
YAML OK
job if: ${{ github.actor != 'github-actions[bot]' && ((github.event_name == 'push' && (github.ref == 'refs/heads/master' || github.ref == 'refs/heads/main')) || (github.event_name == 'workflow_dispatch' && inputs.run_mode == 'full-ci')) }}
```

For PDF output changes, attach a fixture from `sampledata/` and the `verify_pdfs.sh` result.

---

## Related issues

- Relates to #207
- Relates to #206
- Relates to #205

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- Consider deterministic/normalized `compress.wasm` output so docs-only churn drops to zero.
- PR #207 stays open independently; close it manually after this lands if its `compress.wasm` diff is stale.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
