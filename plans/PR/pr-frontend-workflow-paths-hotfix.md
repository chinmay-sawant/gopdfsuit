## Summary

- Fix invalid workflow syntax from #208: `paths` and `paths-ignore` cannot be combined in one trigger. Keep only `paths-ignore: docs/**` (all paths except `docs/`).

---

## Motivation / context

- Plans: `plans/...` (n/a, CI-only hotfix)
- Issues: see **Related issues**
- Both runs for #208 (`33892428981` on the branch, `33892458044` on `master`) failed immediately with "This run likely failed because of a workflow file issue" - GitHub rejects a trigger that sets `paths` and `paths-ignore` together.

---

## Changes

### Area 1 - trigger syntax

- `.github/workflows/frontend-build-commit.yml`: dropped `paths: ["**"]` from `push` and `pull_request`, kept `paths-ignore: docs/**`. Semantics unchanged from intent: run on everything except `docs/`-only changes.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | None beyond #208 intent |
| **Memory** | None |
| **Behavior / correctness** | Workflow parses again; push-to-master-only `build-and-commit-frontend` guard from #208 unchanged |
| **API (`/api/v1/*`) / UI** | None |
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

- [x] Workflow YAML parses (`yaml.safe_load`), triggers verified
- [ ] CI on this PR goes green (proves the workflow file is valid)

### Commands

```sh
make fmt && make lint
make test
```

---

## Screenshots / sample output

```
push: {'branches': ['main', 'master'], 'tags': ['v*'], 'paths-ignore': ['docs/**']}
pr: {'branches': ['*'], 'paths-ignore': ['docs/**']}
YAML OK
```

---

## Related issues

- Relates to #208
- Relates to #207

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- After merge, confirm the `master` push run starts normally (no "workflow file issue") and updates PR #207 in place.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
