# goslop - GitHub Issue Creation Template

Use this when opening a **process-gated** implementation issue (same discipline as `PR_TEMPLATE.md`). Fill the body sections, then create the issue with the CLI so **assignee**, **labels**, and **title** are set correctly.

Repo: [chinmay-sawant/goslop](https://github.com/chinmay-sawant/goslop) (Go port of goslop).

---

## How to use

1. **Pick a title** - short, imperative or noun phrase; include the work area (`engine parse…`, `CWE catalog…`).
2. **Fill the body** below (Context, Scope, Out of scope, Success criteria, Plan, References).
3. **Save a record** (optional but recommended) under `plans/PR/issue-<slug>-body.md` or `plans/v0.0.x/`.
4. **Create the issue** with the `gh` command in [Open the issue](#open-the-issue-gh--required-metadata).
5. **Open a local branch** after the issue exists (e.g. `feat/go-ast-no-cgo`).
6. When shipping: open the PR with `PR_TEMPLATE.md` and `Closes #N` / `Relates to #N`.
7. Progress comments on the issue must follow `COMMENT_TEMPLATE.md` (factual tables; no chatty remaining-step lines).

---

## Open the issue (`gh`) - required metadata

```sh
gh issue create \
  --title "<short title>" \
  --assignee "@me" \
  --label documentation \
  --label enhancement \
  --body-file plans/PR/issue-<short-slug>-body.md
```

| Flag | Rule |
|------|------|
| `--assignee "@me"` | **Required.** Self-assign the author. |
| `--label …` | **Required.** At least one. Prefer `documentation` + `enhancement` for plan/engine work; use `bug` for defect-driven issues. |
| `--body-file` or `--body` | **Required.** Full body with Context / Scope / Success criteria. |
| `--title` | Specific enough to find later; include domain (engine, CWE, BP). |

If the issue already exists without metadata:

```sh
gh issue edit <NUMBER> --add-assignee "@me"
gh issue edit <NUMBER> --add-label documentation --add-label enhancement
```

List labels:

```sh
gh label list
```

---

## Self-assign

- Every issue the author opens for **their** implementation batch **must** list them as assignee.
- Do not leave assignees empty for process-gated work.

---

## Labels

| Label | Use when |
|-------|----------|
| `enhancement` | Product/detector/engine work, new capability |
| `documentation` | Plans, audits, decision records, ledger |
| `bug` | Incorrect behavior / regression |
| `duplicate` / `wontfix` / `invalid` | Triage only |
| `good first issue` / `help wanted` | Community contribution issues |

Mixed plan + code batches: **`documentation` + `enhancement`**.

---

## Ticket references

| In the body | Purpose |
|-------------|---------|
| Link closed parents (`#8`, `#16`) | Continuity |
| Link merged PRs | Evidence |
| Link plan paths | Single source of checklist |
| Link architecture docs | Policy boundaries |

When the work ships, the PR must use:

- `Closes #N` if this issue is fully done
- `Relates to #N` for partial / parent / prior work

See `PR_TEMPLATE.md` → Ticket linking.

---

## Issue body structure

Copy from the line below into `--body` / body file. Delete HTML comments before submit.

---

## Context

<!-- Why this issue exists now. Link closed issues / merged PRs / plan status. -->

-

## Scope (in)

<!-- Concrete, domain-sized. Checklist-friendly bullets. -->

1.
2.

## Out of scope

<!-- Explicit non-goals. -->

-

## Success criteria

- [ ]
- [ ]

## Plan

- Checklist: `plans/port-phasewise-checklist.md`
- Parent: `plans/...`

## References

- Relates to #N / Continues from #N
- PRs: #N
- Docs: `plans/...`

---

## Author checklist before create

- [ ] Self-assigned (`--assignee @me`)
- [ ] Labels applied (at least one)
- [ ] Scope is one issue-sized batch
- [ ] Checklist plan path named in body
- [ ] Out of scope listed
- [ ] Local branch name chosen (create after issue number exists)

---

## Example

```sh
gh issue create \
  --title "engine: replace tree-sitter/CGO with go/ast pure-Go parse" \
  --assignee "@me" \
  --label documentation \
  --label enhancement \
  --body-file plans/PR/issue-go-ast-no-cgo-body.md
```
