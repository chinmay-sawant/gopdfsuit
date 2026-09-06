# gopdfsuit - Pull Request: docs(router): v7.0.0 README hub plus ten sub-guides

## Summary

- Root `README.md` becomes the single router hub: documentation links to all guides, Go/Python/REST quickstarts, samples entry.
- Ten new sub-docs under `documentation/` cover every surface (Go API, Python API, REST, engine, WASM frontend, compress, HTML, doc ops, compliance plus benchmarks, samples catalog).
- Fixes stale links and wrong symbols found by audit (`/v7` import, `Generator` import, dead checklist links, frozen frontend paths).

---

## Motivation / context

- Plans: `documentation/PDF_SUITE_GUIDE.md`, `documentation/index.md`
- Issues: see **Related issues**
- 12 parallel scan agents audited every surface; all 56 Go signatures grepped against `pkg/gopdflib`, all Python symbols verified, 72/72 relative links resolve, zero em dashes.

---

## Changes

### Root router README

- Documentation hub with links to all ten new guides plus existing deep dives.
- 5-minute quickstarts: `go get .../v7@v7.0.0`, `pip install .` via `build.sh`, Docker plus curl generate.
- Fixed: Go import gains `/v7`; `from pypdfsuit import Generator` (never exported) replaced with `TemplateBuilder, Font`; samples entry added.

### New sub-docs (`documentation/`)

- `gopdflib-api.md`: every `pkg/gopdflib` op with Go snippets (generate, merge, split, compress, fill, redact, HTML, builder, prepared, props, BOPS, errors).
- `pypdfsuit-api.md`: every binding with Python snippets, builder-first.
- `rest-api.md`: all 16 `/api/v1/*` routes with curl, auth, error envelope, caps.
- `engine-overview.md`: in-memory pipeline, tiers, redact, signing, fonts, HTML, TTL. Notes AES-128 code truth vs AES-256 claims.
- `wasm-frontend.md`: artifacts, page routes, consent model, offline caches, rebuilds.
- `compress-guide.md`: tiers, snippets, limits, no-shrink passthrough, encrypted rejection.
- `html-guide.md`: field tables, defaults, ignored knobs, WASM and fidelity limits.
- `document-ops-guide.md`: merge, split page-spec syntax, XFDF caveat, redact modes.
- `compliance-benchmarks.md`: gates, signatures, validation and bench commands (router, no duplication).
- `samples-catalog.md`: every `sampledata/` directory with verified run commands.

### Link hygiene

- `documentation/index.md`: hub entries for the suite guide plus ten API/guide routers.
- `documentation/DEPLOYMENT_CHECKLIST.md`: dropped 2 dead links (`AUTHENTICATION_SUMMARY.md`, `AUTH_FLOW_DIAGRAMS.md`), fixed `docs/` prefix text.
- `CONTRIBUTING.md`: replaced 2 nonexistent frontend paths with `documentation/rest-api.md` and `internal/handlers/`.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | None (docs only) |
| **Memory** | None |
| **Behavior / correctness** | None; no code touched |
| **API (`/api/v1/*`) / UI** | Docs describe existing routes; no route changes |
| **Dependencies** | None |
| **Binary size / build time** | None |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | Docs only; gates unchanged |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [x] Symbol check: 56 `pkg/gopdflib` exports grepped against code, all documented names exist
- [x] Link check: 72/72 relative links across new plus edited files resolve on disk
- [x] Style check: zero em dashes in new plus edited files
- [ ] `make test` (docs only, no code touched)
- [ ] `make lint` (docs only, no code touched)

### Commands

```sh
python3 - <<'EOF'
# relative-link check over new plus edited md files: 72 checked, 0 broken
EOF
```

---

## Screenshots / sample output

```
checked=72 broken=0
```

---

## Related issues

- Relates to #216
- Relates to #218

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- Decide AES-256 (README/AGENTS) vs AES-128 (engine code) and update one side.
- `makefile` HTML comment still claims HTML conversion is server-side; WASM `goHtmlToPDF` exists.
- Consider a `concurrency` group on the PyPI publish job (tag-push and release runs raced on v7.0.0).

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
