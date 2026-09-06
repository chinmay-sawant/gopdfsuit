# gopdfsuit - Pull Request: chore(release): prep v7.0.0 with /v7 module path

## Summary

- Major bump `/v6` to `/v7` for `v7.0.0` (small v), driven by `pkg/gopdflib` alias to owned-struct migration.
- Covers everything merged after `v6.0.0` (86 commits, PRs #203-#215): no-GS compress, WASM-first viewer/editor, pure-Go HTML, fluent Go/Python builders, cache TTL, merge catalog fix.
- Local-only prep so far: module rewrite (157 files), Python `7.0.0` strings, live docs to `/v7`. No tag, no merge, no PyPI upload.

---

## Motivation / context

- Plans: `plans/PR/pr-release-v7.0.0.md` (this file), `guides/release-prep.md`, `plans/reviews/`, `plans/review/go-structure-performance-review-2026-09-05.md`
- Issues: see **Related issues**
- Baseline: `v6.0.0` (2026-06-23) is latest GitHub release; HEAD `683b2d7` on `master` before this branch.

---

## Changes

### Release mechanics (/v6 to /v7)

- Root `go.mod`: `github.com/chinmay-sawant/gopdfsuit/v6` to `.../v7`; `go mod tidy` clean.
- Bulk import rewrite `gopdfsuit/v6` to `gopdfsuit/v7` across 149 `.go`/`go.mod` files (`pkg/`, `internal/`, `cmd/`, `test/`, `typstsyntax/`, `bindings/python/cgo/exports.go`).
- `sampledata/go.mod` and `sampledata/benchmarks/gopdfkit_compare/go.mod`: require `.../v7 v7.0.0` (replace targets kept).
- `makefile`: `VERSION ?= 7.0.0`.
- Python: `bindings/python/pyproject.toml` and `pypdfsuit/__init__.py` to `7.0.0`.
- Live docs to `/v7`: `documentation/GETTING_STARTED_GOPDFLIB.md`, `documentation/CACHING_AND_MEMORY_LIFECYCLE.md`, `frontend/src/components/editor/snippet.js`, `sampledata/benchmarks/gopdfkit_compare/README.md`, `AGENTS.md`.
- History (`plans/`, `guides/optimizations/`, `.vscode/`) untouched; `docs/` (Vite output) not hand-edited.

### Feature summary since v6.0.0 (bullet points)

- Compress without Ghostscript: `CompressPDF`, `/api/v1/compress`, WASM tiers Light/Medium/Heavy.
- WASM browser-local app: `gopdfsuit.wasm` full engine plus `compress.wasm` worker, offline cache, consent-gated server fallback; WASM-first viewer/editor.
- HTML engine swap: headless-Chrome `gochromedp` replaced with pure-Go `gowkhtmltopdf v0.2.5`; inline-HTML converts in browser, URL fetch server-only.
- Fluent builders: raw `"Helvetica:18:100:center:0:0:0:0"` also spells as `Font("Helvetica").Size(18).Bold().Center()`, `NewDocument`/`WithTitleFontOpts`, table/cell helpers; Python `builder.py` parity; samples in `sampledata/builder-snippets/`.
- Engine hardening: cache TTL (`GOPDFSUIT_CACHE_TTL`, default 3m), BOPS bench, merge catalog from newest valid trailer, PDFA fonts in WASM, redact CTM/MediaBox fixes.
- Frontend rebuild: site shell, shared `PdfPreview`, viewport-fit viewer, filler fit, stars pill.
- Reviews landed: round3 engine refactor, round2 residual-friction decisions, Go structure/performance R01-R18.

### PR links since v6.0.0 (all merged into master)

- #203 feat(compress): PDF compression without Ghostscript - https://github.com/chinmay-sawant/gopdfsuit/pull/203
- #204 chore(frontend + backend): wasm-first viewer editor - https://github.com/chinmay-sawant/gopdfsuit/pull/204
- #205 fix(frontend): open PR for docs auto-build instead of pushing to master - https://github.com/chinmay-sawant/gopdfsuit/pull/205
- #206 chore: auto-build frontend - https://github.com/chinmay-sawant/gopdfsuit/pull/206
- #207 chore: auto-build frontend - https://github.com/chinmay-sawant/gopdfsuit/pull/207
- #208 fix(frontend): restrict docs auto-build to master pushes only - https://github.com/chinmay-sawant/gopdfsuit/pull/208
- #209 fix(frontend): drop invalid paths alongside paths-ignore in workflow - https://github.com/chinmay-sawant/gopdfsuit/pull/209
- #210 feat(engine): round3 review plus fluent builders and snippets - https://github.com/chinmay-sawant/gopdfsuit/pull/210
- #211 chore: auto-build frontend - https://github.com/chinmay-sawant/gopdfsuit/pull/211
- #212 feat(frontend + backend): rebuild site shell and workspaces, engine cache TTL plus BOPS bench - https://github.com/chinmay-sawant/gopdfsuit/pull/212
- #213 chore: auto-build frontend - https://github.com/chinmay-sawant/gopdfsuit/pull/213
- #214 fix(merge): resolve catalog from newest valid trailer and unblock PDFA fonts in WASM - https://github.com/chinmay-sawant/gopdfsuit/pull/214
- #215 chore: auto-build frontend - https://github.com/chinmay-sawant/gopdfsuit/pull/215

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Cache TTL bounds cross-request caches; pools kept; BOPS bench added |
| **Memory** | Signer/font/subset/page-compress caches bounded; no API change |
| **Behavior / correctness** | Merge catalog fix; redact box fixes; HTML `svg` rejected, `DPI`/`LowQuality` warned |
| **API (`/api/v1/*`) / UI** | Routes unchanged; UI adds WASM-first flows with consent banner |
| **Dependencies** | `gochromedp` removed; `gowkhtmltopdf v0.2.5`; `x/sync v0.20.0` |
| **Binary size / build time** | WASM split permanent (`compress.wasm` ~8M vs `gopdfsuit.wasm` ~31M) |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | veraPDF hard gate plus structure-tree checks; PDFA fonts in WASM |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| Go import path `/v6` to `/v7` | `go get github.com/chinmay-sawant/gopdfsuit/v7@v7.0.0`, update imports, `go mod tidy` |
| `gopdflib.X` alias to owned struct | Same fields/JSON; fix type assertions and `reflect` checks |
| Template JSON | Additive `omitempty` only; no action |
| Auth default | Still open locally; `REQUIRE_AUTH=1` forces auth |

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

Done locally on this branch so far: `go mod tidy` clean, `go build ./...` clean, `go vet ./...` clean. Full `make test`, frontend build, and veraPDF spot check pending before tag.

---

## Screenshots / sample output

```
go mod tidy -> clean
go build ./... -> clean
go vet ./... -> clean
git status -> 157 files changed (import rewrite + versions + docs), no commits on master
```

---

## Release notes draft (v7.0.0, for GitHub release title `GoPdfSuit v7.0.0`)

Tag `v7.0.0` (small v). Full changelog `v6.0.0...v7.0.0`. Do not publish yet.

Overview: major bump from owned-types migration; 86 commits and PRs #203-#215 above.
Performance: cache TTL, BOPS bench, no-GS compress tiers, pure-Go HTML, WASM split.
Breaking: `/v6` to `/v7` imports; type-identity break; JSON wire compatible.
New: compress, WASM browser-local conversions (inline HTML local, URL server-only),
`gochromedp` to `gowkhtmltopdf v0.2.5`, fluent builders (colon strings to `.Font().Size()` chains),
Python `pypdfsuit==7.0.0`, compliance gates, docs truth home `documentation/`.
Upgrade: `go get .../v7@v7.0.0`; `pip install pypdfsuit==7.0.0`;
Docker `chinmaysawant/gopdfsuit:7.0.0` plus `latest`.
PyPI: CI pushes sdist plus wheels with rebuilt native lib on tag; idempotent guard; no TestPyPI.

---

## Related issues

- Relates to #203
- Relates to #204
- Relates to #210
- Relates to #212
- Relates to #214

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- Merge this PR into `master` first (no auto-merge here).
- Then rebuild `docs/` via frontend build and `compress.wasm` via `make wasm-compress`.
- Then run full `make test` plus veraPDF spot check.
- Then create tag `v7.0.0` and GitHub release (`GoPdfSuit v7.0.0`) plus PyPI publish via CI.
- No tag creation in this PR.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
