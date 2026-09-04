# gopdfsuit - Pull Request Template

Use this document as the base when authoring GitHub pull requests for [chinmay-sawant/gopdfsuit](https://github.com/chinmay-sawant/gopdfsuit). Copy the sections below into the PR description and fill in each section. Delete guidance comments before submitting.

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
  --base master \
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
| `--base master` | Default integration branch is **`master`** (not `main`). |

If the PR already exists without metadata:

```sh
gh pr edit <NUMBER> --add-assignee "@me"
gh pr edit <NUMBER> --add-label documentation --add-label enhancement
```

---

## Multi-workstream / epic integration (parallel agents)

When **multiple issue-sized branches** are developed in parallel, also ship a **single integration branch** targeting `master`.

```sh
git fetch origin master
git checkout -b chore/epic-N-integration origin/master
# merge child heads, validate, push, open PR to master
```

Prefer **merging only the integration PR** into `master` when an epic stack exists.

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

Scopes for gopdfsuit: `generate`, `merge`, `split`, `compress`, `fill`, `redact`, `html`, `fonts`, `handlers`, `engine`, `frontend`, `bindings`, `verapdf`, `docs`.

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
| **API (`/api/v1/*`) / UI** | |
| **Dependencies** | |
| **Binary size / build time** | |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [ ] `make test` (`go test ./...` plus Python bindings plus `test/verify_pdfs.sh`)
- [ ] `make test-integration` (`go test -count=1 -v ./test`) when handlers or engine changed
- [ ] `make lint` plus `go vet` (zero ESLint warnings in `frontend/`)
- [ ] `make build` (`go build -o bin/app ./cmd/gopdfsuit`) when shippable change
- [ ] `make test-verify-pdfs` or `make test-scan-pdfs-compliance` when PDF output changed
- [ ] `cd frontend && npm run build` when UI changed (never hand-edit `docs/`)
- [ ] `make wasm-compress` when `cmd/wasmcompress/` changed

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
(paste handler response, verify_pdfs.sh summary, or UI screenshot path)
```

For PDF output changes, attach a fixture from `sampledata/` and the `verify_pdfs.sh` result.

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
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed

---

## Example titles (gopdfsuit)

```
feat(compress): add Heavy tier preset for scanned PDFs
fix(fill): handle XFDF fill on compressed object streams
perf(engine): reuse sync.Pool buffers in template-pdf hot path
feat(html): support header footer in htmltopdf via gochromedp
docs: update TEMPLATE_REFERENCE for redact capabilities
chore: refresh zerodha fixtures in sampledata
```

## Example `gh pr create` (full)

```sh
gh pr create \
  --base master \
  --title "feat(compress): add Heavy tier preset for scanned PDFs" \
  --body-file plans/PR/pr-compress-heavy-tier.md \
  --assignee "@me" \
  --label documentation \
  --label enhancement
```
