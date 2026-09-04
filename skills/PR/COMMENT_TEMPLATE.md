# gopdfsuit - Issue / PR Comment Template

Use this for **progress updates** on GitHub issues and pull requests. Keep comments factual and complete. Avoid dangling next-step teases aimed at the human ("when you want", "let me know", "remaining for you").

---

## Tone rules

**Do**

- Report what landed (commit SHA, branch, PR number when known).
- Use tables for multi-item status.
- List **follow-up work as optional backlog**, not as incomplete homework for the reader.
- Link plans and prior PRs with full paths or `#N` references.

**Do not**

- End with casual prompts: "PR when you want", "let me know once done", "Remaining checklist: Phase 7 ship...".
- Address the maintainer as if the comment is a chat thread.
- Leave success criteria half-checked without stating whether the issue/PR is ready to close.

If work is complete for this issue: say **ready for review** / **issue can close on merge** and stop.
If work continues: list **next batch candidates** under a clear "Out of scope for this update" or "Future work (new issue)" heading - never "still need you to...".

---

## Progress update (issue)

```markdown
## Progress update

**Branch:** `<branch>`
**Commit(s):** `<sha>` - `<subject>`

### Delivered

| Area | Result |
|------|--------|
| ... | ... |

### Validation

- `make lint` - pass
- `make test` - N passed
- `make test-verify-pdfs` - pass (when PDF output changed)

### Plans / evidence

- `plans/...`
- `sampledata/<area>/` fixture updated (when relevant)

### Issue status

- **This issue:** ready to close on PR merge | still open for [concrete incomplete criterion]
- **Future work (optional, new issue):** short bullets only if useful - no "waiting on you" language
```

---

## Progress update (PR)

```markdown
## Update

**Commit:** `<sha>` - `<subject>`

### What changed since last push

-

### Validation

- `make lint` / `make test` - pass
- `make test-integration` - pass (when handlers or engine changed)

### Notes for reviewers

-
```

---

## Completion / ship note (issue after PR opened)

```markdown
## Ship

PR: https://github.com/chinmay-sawant/gopdfsuit/pull/N

Implements the checklist for this issue. Merge closes the process gate via `Closes #N` on the PR body.

### Validation

- `make lint` - pass
- `make test` - N passed
- `make test-verify-pdfs` - pass (when PDF output changed)

### Future work (optional)

Further PDF op audits should open a **new** issue; do not reopen this one for unrelated ops.
```

---

## Author checklist before posting

- [ ] No "when you want" / "let me know" / chatty remaining-phase lines
- [ ] Commit/PR/branch identified when work landed
- [ ] Tables used if more than two outcomes
- [ ] Incomplete work phrased as optional future backlog or explicit open success criteria
