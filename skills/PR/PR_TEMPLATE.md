# goslop - Pull Request Template

Use this document as the base when authoring GitHub pull requests for [chinmay-sawant/goslop](https://github.com/chinmay-sawant/goslop). Copy the sections below into the PR description and fill in each section. Delete guidance comments before submitting.

---

## How to use this template

1. **Pick a title** using the convention in [PR title](#pr-title).
2. **Write a 1-3 sentence summary** - what changed and why (not a file list).
3. **Fill in each section** - keep `Summary`, `Changes`, `Test plan`, and `Related issues`.
4. **Link related tickets** in the body **and** in `gh pr create` metadata.
5. **Choose labels** and **self-assign**.
6. Save a filled copy under `plans/PR/pr-<short-slug>.md` **before** opening the PR when process-gated.
7. Open the PR with the CLI checklist so assignee, labels, and body stay in sync.

---

## Open the PR (`gh`) - required metadata

```sh
# From the feature branch (already pushed):
gh pr create \
  --base main \
  --head "$(git branch --show-current)" \
  --title "<type>(<scope>): <short imperative description>" \
  --body-file plans/PR/pr-<short-slug>.md \
  --assignee "@me" \
  --label documentation \
  --label enhancement
```

| Flag | Rule |
|------|------|
| `--assignee "@me"` | **Required.** Self-assign the opening author. |
| `--label …` | **Required.** At least one label. |
| `--body-file …` | **Required.** Full template body with **Related issues** filled. |
| `--title` | Must match [PR title](#pr-title) convention. |
| `--base main` | Default integration branch is **`main`** (not `master`). |

If the PR already exists without metadata:

```sh
gh pr edit <NUMBER> --add-assignee "@me"
gh pr edit <NUMBER> --add-label documentation --add-label enhancement
```

---

## Multi-workstream / epic integration (parallel agents)

When **multiple issue-sized branches** are developed in parallel, also ship a **single integration branch** targeting `main`.

```sh
git fetch origin main
git checkout -b chore/epic-N-integration origin/main
# merge child heads, validate, push, open PR to main
```

Prefer **merging only the integration PR** into `main` when an epic stack exists.

---

## Self-assign

- Every PR the author opens **must** list them as assignee (`--assignee "@me"`).

---

## Ticket linking

| Keyword | When to use |
|---------|-------------|
| `Closes #N` / `Fixes #N` / `Resolves #N` | This PR fully completes the issue |
| `Relates to #N` | Partial progress, dependency, or prior related work |
| `Refs #N` | Soft reference |

---

## Labels

| Title type | Suggested labels |
|------------|------------------|
| `feat` | `enhancement` |
| `fix` | `bug` or `enhancement` |
| `docs` | `documentation` |
| `perf` / `refactor` / `chore` | `enhancement` (+ `documentation` if plans) |

---

## PR title

```
<type>(<scope>): <short imperative description>
```

Types: `feat`, `fix`, `perf`, `refactor`, `test`, `docs`, `chore`, `ci`.

---

## PR description structure

Copy everything below this line into the GitHub PR body (and into `plans/PR/pr-<short-slug>.md`).

---

## Summary

<!-- 1-3 sentences: WHAT changed and WHY. -->

-

---

## Motivation / context

- Plans: `plans/...`
- Issues: see **Related issues**

---

## Changes

### Area 1

-

### Area 2

-

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | |
| **Memory** | |
| **Behavior / correctness** | |
| **API / CLI** | |
| **Dependencies** | |
| **Binary size / build time** | |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [ ] `make test`
- [ ] `make lint` / `go vet`
- [ ] `CGO_ENABLED=0 go build -o bin/goslop ./cmd/goslop` (when pure-Go is required)
- [ ] `make run` wall time vs baseline (hard &lt; 400ms; soft ±50ms of reference)
- [ ] `make reference-metrics` / gopdfsuit hard metrics if detector surface changed

### Commands

```sh
make test
make run
```

---

## Screenshots / sample output

```
(paste make run summary)
```

---

## Related issues

- Closes #NNN
- Relates to #NNN

---

## PR metadata checklist (author)

- [ ] Self-assigned (`--assignee @me`)
- [ ] Labels applied
- [ ] Related issues filled with real ticket IDs
- [ ] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

-

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed

---

## Example titles (goslop)

```
feat(engine): replace tree-sitter with go/ast pure-Go parse
fix: respect --skip when GoScan bundle runs
perf(engine): parallel file scan without CGO
docs: align README with §12.4 parity baseline
chore: drop tree-sitter CGO dependencies
```

## Example `gh pr create` (full)

```sh
gh pr create \
  --base main \
  --title "feat(engine): replace tree-sitter/CGO with go/ast" \
  --body-file plans/PR/pr-go-ast-no-cgo.md \
  --assignee "@me" \
  --label documentation \
  --label enhancement
```
