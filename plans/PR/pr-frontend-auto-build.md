## Summary

Fixes the failing `build-and-commit-frontend` job by proposing the `docs/` rebuild as a pull request instead of pushing directly to protected `master`. Also moves the job off EOL Node 20 and drops the Node 20 runtime action that triggered the deprecation warning.

---

## Motivation / context

- Plans: `plans/PR/pr-frontend-auto-build.md`
- Issues: none
- Run 33889364564 failed at the `Commit and push changes` step: `stefanzweifel/git-auto-commit-action@v5` pushed `docs/**` straight to `master`, and branch protection rejected it with `GH013: Changes must be made through a pull request`. The same step emitted the Node 20 deprecation warning because the v5 action runs on the Node 20 runtime.

---

## Changes

### Area 1 - PR-based docs auto-build

- Replaces `stefanzweifel/git-auto-commit-action@v5` with `peter-evans/create-pull-request@v8` (Node 24 runtime, no deprecation warning).
- The rebuilt `docs/**` is committed to branch `chore/auto-build-frontend` and opened as a `chore: auto-build frontend` PR with the `documentation` label. Branch is deleted on merge via `delete-branch: true`.
- Base is explicit: `master` for push events, the PR base ref for `pull_request` events.
- Adds `pull-requests: write` permission to the job, which the new step needs to open PRs.
- Behavior change: on `pull_request` runs the docs rebuild no longer amends the contributor branch in place. It opens a separate docs PR instead, which stays compliant when the base is protected.
- No loop risk: the auto commit carries `[skip ci]`, and the follow-up run finds no diff so it converges without opening another PR.

### Area 2 - Node version

- `build-and-commit-frontend` setup-node moves from Node 20 (EOL April 2026) to Node 22. Local toolchain here runs Node v22.20.0. `frontend-lint` stays on 18, out of scope.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No runtime impact. One extra short-lived PR per frontend change that alters build output. |
| **Memory** | None. |
| **Behavior / correctness** | `docs/` stays auto-generated and reviewable via PR. Direct-push failure on protected branches is gone. |
| **API (`/api/v1/*`) / UI** | None. |
| **Dependencies** | CI-only: swaps one third-party action for another pinned major (`v8`). No app deps. |
| **Binary size / build time** | None. |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | None. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Contributors no longer get silent `docs/` amendments pushed into their PR branches. Merge the auto-opened docs PR instead. |

---

## Test plan

- [ ] YAML parses (`python3 -c "yaml.safe_load(...)"`)
- [ ] `make lint` plus `go vet` (workflow file is CI-only, no app code touched)
- [ ] Post-merge: confirm the next push to `master` opens a `chore: auto-build frontend` PR instead of failing with GH013
- [ ] Post-merge: confirm no Node 20 deprecation warning on the `Open PR with rebuilt docs` step

### Commands

```sh
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/frontend-build-commit.yml'))"
make fmt && make lint
```

---

## Screenshots / sample output

```
YAML OK, steps: 7
last step uses: peter-evans/create-pull-request@v8
node: ['22']
```

Failing run before fix: `actions/runs/33889364564/job/101078728767`
`remote: error: GH013: Repository rule violations found for refs/heads/master.`
`- Changes must be made through a pull request.`

---

## Related issues

- None

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`bug`)
- [x] Related issues filled with real ticket IDs (none)
- [x] Filled body committed under `plans/PR/pr-frontend-auto-build.md`

---

## Follow-ups (out of scope)

- Consider assigning or auto-merging the bot docs PRs to keep `docs/` fresh without manual merges.
- Align `frontend-lint` off Node 18 in a separate chore.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
