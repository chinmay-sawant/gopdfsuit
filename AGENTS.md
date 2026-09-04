# AGENTS.md

> This file is the conventions ledger for every coding agent working in this
> repo (opencode, grok, gemini, codex, antigravity/agy, claude). Read it at
> session start. Replace placeholder sections below with project-specific
> conventions as the repo evolves.

## Project

[Placeholder - add a short description of what this repo owns, who it is for,
and what it is not.]

- Module: `[Placeholder - e.g. github.com/org/repo]`
- GitHub repo: `[Placeholder - https://github.com/org/repo]`
- Default branch: `[Placeholder - e.g. main / master]`
- Binaries / libraries / apps: `[Placeholder - what gets built and shipped]`
- Pipeline / architecture overview: `[Placeholder - stages, owners, data flow]`

## Todo protocol - response-only (mandatory)

[Placeholder - keep, adapt, or remove per project. Default below is generic.]

1. **Create todos in the response via API before any work.** On receiving any
   task (feature, fix, docs, question with multi-step work), immediately call
   the todo API (`todowrite` or equivalent) to publish the plan as todos. Do
   not start work until the todo list is visible in the API response.
2. **Show current todos in every response.** Each assistant turn must render
   the current todo list with status markers (`pending`, `in_progress`,
   `completed`, `cancelled`) and clearly highlight which item is
   `in_progress`.
3. **Do not store todos on disk.** Do not create `TODO.md`, `todos.json`, or
   any other todo-tracking file. Todos live only in the API response state.
4. **Keep response todos updated as you go.** Mark items `in_progress` when
   started and `completed` when finished. If scope changes, update the list
   immediately.
5. **Completion requires todos to show done.** A task is done only when all
   todos show `completed` and you have sent a final summary stating what
   shipped.

If the todo API is unavailable, state that in the response and list todos
inline as a fallback - still do not write a file.

## Golden rules

[Placeholder - replace with project-specific working agreements. Keep only
what applies.]

1. **No git commands without explicit permission.** Never run `git add`,
   `git commit`, `git push`, `git restore`, `git clean`, `git reset`, or
   `git stash` unless the user asks.
2. **No em dashes ("--") in any written output, docs, or commit messages.**
   Use plain hyphens or restructure.
3. [Placeholder - branch naming convention]
4. [Placeholder - PR / issue template and process]
5. [Placeholder - checklist / ledger update rule]
6. [Placeholder - writing / docs quality bar]

## Verification gates

[Placeholder - list the real gates that prove a change is safe. Replace the
example rows below.]

| Gate | Command | What it proves |
|------|---------|----------------|
| [Placeholder] | `[Placeholder - e.g. make test]` | [Placeholder] |
| [Placeholder] | `[Placeholder - e.g. make lint]` | [Placeholder] |

Run targeted checks during a session; run the full gate set once at session
end before claiming done.

## Things to AVOID

[Placeholder - add paid-for lessons specific to this repo. Remove generic
examples below if they do not apply.]

1. [Placeholder - e.g. guessing APIs without grepping]
2. [Placeholder - e.g. claiming completion without gate output]
3. [Placeholder - e.g. scope creep]

## Code structure

[Placeholder - describe layout, file size limits, module split strategy,
and verification after moves.]

## Skills (this repo)

Skills live under `skills/<name>/`. Current inventory:

- `skills/PR/` - templates for PRs, issues, review comments
- `skills/feynman/` - plain-words explanation loop with self-audit
- `skills/unslop/` - cut AI tells from writing, add human voice
- `skills/phase-wise-checklist/` - evidence-backed, phase-wise implementation checklists

[Placeholder - add new skills here as they are added. Remove entries that no
longer apply.]

## Dependency policy

[Placeholder - describe allowed dependencies, allowlist enforcement, and
approval process if any.]

---

## FAQ - does the root AGENTS.md get read automatically?

Yes. Creating `AGENTS.md` at the repo root is the standard, and it is read
automatically by every major tool: opencode, codex, gemini CLI,
antigravity/agy, grok, claude code, cursor. You do not need anything under
`.agents/`. Keep this one file at the root and keep it current.
