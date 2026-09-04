---
name: PR
description: Templates and checklists for PRs, issues, and progress comments so every PR has the same shape. Use when creating a PR, issue, or progress comment.
---

# PR - templates for shipping gopdfsuit

Use these templates when opening a GitHub PR, issue, or progress comment for [chinmay-sawant/gopdfsuit](https://github.com/chinmay-sawant/gopdfsuit) so every thread has the same shape. Default branch is `master`.

## Templates in this skill

- `PR_TEMPLATE.md` - PR title convention with gopdfsuit scopes (`generate`, `compress`, `fill`, `redact`, `html`, `frontend`, `bindings`), summary, changes, impact table including PDF compliance, test plan with `make test`, `make test-integration`, `make lint`, `make build`, `make test-verify-pdfs`, and `gh pr create` with `--assignee "@me"` and labels
- `ISSUE_TEMPLATE.md` - context, scope, out of scope, success criteria, fixtures under `sampledata/<area>/`, and `gh issue create`
- `COMMENT_TEMPLATE.md` - factual progress updates without chatty trailing prompts, with gopdfsuit validation lines

Load the relevant template file and follow it verbatim.
