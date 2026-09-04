# AGENTS.md

> This file is the conventions ledger for every coding agent working in this
> repo (opencode, grok, gemini, codex, antigravity/agy, claude). Read it at
> session start. Keep it current as the repo evolves.

## Project

Template-based PDF platform in Go with React UI and Python bindings.

- Module: `github.com/chinmay-sawant/gopdfsuit/v6` (Go 1.26.4)
- GitHub repo: `https://github.com/chinmay-sawant/gopdfsuit`
- Default branch: `master` (`origin/HEAD` points to `origin/master`)
- Binaries / libraries / apps:
  - `bin/app` from `cmd/gopdfsuit/main.go` (Gin REST API on :8080)
  - `frontend/public/compress.wasm` from `cmd/wasmcompress/main.go` (in-browser compressor)
  - `pkg/gopdflib/` public Go library (Generate, Merge, Split, Compress, Fill, Redact, HTML)
  - `bindings/python/` wheel `pypdfsuit` via CGO `exports.go` plus `libgopdfsuit.so`
  - `frontend/` React 18 + Vite app `gopdfsuit-frontend`, builds to `docs/` static site
  - `dockerfolder/Dockerfile` and `Dockerfile_cloudrun` images
- Pipeline / architecture overview:
  - Upload JSON template plus base64 assets to `/api/v1/*`, handlers in `internal/handlers/` decode with pooled Sonic JSON, in-memory engine in `internal/pdf/` renders (fonts embed, tagging, XMP, ECDSA sign, AES-256), PDF bytes in response.
  - Ops: `generate/template-pdf`, `fill` (AcroForm/XFDF), `merge`/`split`, `compress` (Light/Med/Heavy tiers, server or WASM, no Ghostscript), `htmltopdf/htmltoimage` (pure-Go via gowkhtmltopdf, no browser), `redact/*`, `fonts` GET/POST, `template-data`.
  - Compliance: PDF 2.0 base, opt-in PDF/A-4 (`pdfaCompliant:true`) plus PDF/UA-2 (tagging, MCID, ParentTree), PKCS#7/X.509 RSA plus ECDSA P-256 signatures. Validation via vendored `verapdf/` (needs Java 11+) plus `structure_tree_check.py` and `test/verify_pdfs.sh`.

## Todo protocol - response-only (mandatory)

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

1. **No git commands without explicit permission.** Never run `git add`,
   `git commit`, `git push`, `git restore`, `git clean`, `git reset`, or
   `git stash` unless the user asks.
2. **No em dashes in any written output, docs, or commit messages.** Use plain hyphens or restructure.
3. Branch naming: `feat/<slug>`, `fix/<slug>`, `chore/<slug>`, `exp/<slug>` (see `feature/compression-nogs`, `chore/improves-fixes`, `exp/cursor-optimization`). Create from issue number context, push with `-u`.
4. PR / issue process: use `skills/PR/` templates verbatim. Self-assign with `--assignee "@me"`, apply at least one label, link `Closes #N` or `Relates to #N`, save process-gated body under `plans/PR/pr-<slug>.md` before `gh pr create`.
5. Checklist rule: for phased work use `skills/phase-wise-checklist/` with one canonical ledger under `plans/`. Update a row to `[x]` only after its `make` gate passes. Never keep duplicate active rows in two files.
6. Writing quality: run `skills/unslop/` on all prose, explain with `skills/feynman/` loop when touching PDF engine semantics (fonts, tagging, xref, compress tiers, XFDF).

## Verification gates

Run targeted checks during a session. Run the full set once at session end before claiming done.

| Gate | Command | What it proves |
|------|---------|----------------|
| Full Go plus Python plus PDF compliance | `make test` | `go test ./...` plus `bindings/python pytest` plus `test/verify_pdfs.sh` pass |
| Integration suite | `make test-integration` | `go test -count=1 -v ./test` handler suite standalone (full gate is `make test`: go plus python plus verify, sequential) |
| Lint backend plus frontend | `make lint` | `golangci-lint run -E revive,gocritic,gocyclo,goconst ./...` plus `frontend npm run lint` clean |
| Build binary | `make build` | Integration tests pass plus `go build -o bin/app ./cmd/gopdfsuit` succeeds |
| Frontend bundle | `cd frontend && npm run build` | Vite build plus WASM compressor bundle succeed |
| Format and vet | `make fmt && make vet` then `go fmt ./...`, `go vet ./...` | Formatting and vet clean before lint |
| PDF compliance spot check | `make test-verify-pdfs` or `make test-scan-pdfs-compliance` or `make test-zerodha-compliance` | veraPDF plus structure tree checks pass on fixtures |
| WASM artifact | `make wasm-compress` | `GOOS=js GOARCH=wasm` build to `frontend/public/compress.wasm` succeeds |

Notes: CI only runs lint plus `docs/` build. Always run `make fmt && make lint && make test` locally. Frontend needs zero ESLint warnings. Never hand-edit `docs/` output, it is Vite build output auto-committed by CI.

## Things to AVOID

1. Do not hand-edit `docs/` (Vite output). Edit `frontend/src/` and rebuild.
2. Do not import `internal/*` from outside the module. External Go code uses `pkg/gopdflib`.
3. Do not use `encoding/json` on hot paths. Use `bytedance/sonic` plus pooled `StreamDecoder`. Preserve `sync.Pool` reuse (pdfBuffer, zlib, fontRegistry) and tier prealloc.
4. Do not trust veraPDF alone. ParentTree MCID checks need `structure_tree_check.py` (MCID must point to TD or TH, not TR). `avalpdf` output is warnings-only unless `VERIFY_AVALPDF_STRICT=1`.
5. Do not assume XFDF fill works on compressed object streams. The `/NeedAppearances true` byte approach fails there and needs pdfcpu path.
6. Do not commit secrets, certs, `.env`, `verapdf/` binaries, `.pdf-validators/`, `reports/`, or `temp_*.pdf`. Respect `.gitignore` (`*.exe`, `*.so`, `*.test`, `*.out`, coverage profiles).
7. Do not rely on WSL-only scripts or `make` targets inside PowerShell or CMD.
8. Do not claim completion without gate output. Paste `make test`, `make lint`, or `make test-integration` results.

## Code structure

- `cmd/gopdfsuit/main.go`: Gin server entrypoint. `cmd/wasmcompress/main.go`: WASM compressor entrypoint.
- `pkg/gopdflib/`: public API, one file per op (generator, merge, split, compress, fill, redact, html). `pkg/fontutils/`: font helpers.
- `internal/pdf/`: private engine (generator, pagemanager, draw, merge/, compress/, redact/, form/xfdf, signature/, encryption/, svg/, xref/). `internal/pdf/font/`: split into `subset.go`, `ttf.go`, `metrics.go`, `compression.go`, `compress_cache.go`, `pdfa.go`, `registry.go`, `types.go`.
- `internal/handlers/`, `internal/models/` (PDFTemplate JSON), `internal/middleware/` (cors, Google auth), `internal/benchmarktemplates/`.
- `frontend/src/pages/`: Merge, Split, Compress, Viewer, Editor, converters. Stack: Vite 5, React 18, react-router-dom 6, react-pdf 10, HeadlessUI 2.
- `bindings/python/pypdfsuit/`: generator, merge, split, fill, redact, html, compress via `cgo/exports.go`.
- `test/`: Go integration tests `integration_*_test.go` plus k6 scripts in `test/generate_template-pdf/`.
- `sampledata/`: fixtures per feature (`merge/`, `split/`, `compress/`, `filler/`, `financialreport/`, `typstsyntax/`, `gopdflib/zerodha/`, `benchmarks/`).
- `documentation/`: operational docs (`AUTHENTICATION.md`, `BENCHMARKS.md`, `DEPLOYMENT_CHECKLIST.md`, `TEMPLATE_REFERENCE.md`, plus caching, validators, signatures, getting started). `guides/` holds remaining build and optimization notes. `plans/`: design notes plus `plans/PR/` ledger. `typstsyntax/`: Typst subset lexer, parser, renderer.
- Tests colocated as `*_test.go`. Directory name equals package name (`pdf`, `font`, `xref`, `models`, `middleware`, `gopdflib`, `typstsyntax`).

## Skills (this repo)

Skills live under `skills/<name>/`. Current inventory:

- `skills/PR/` - templates for PRs, issues, review comments (gopdfsuit gates and examples)
- `skills/feynman/` - plain-words explanation loop with self-audit
- `skills/unslop/` - cut AI tells from writing, add human voice
- `skills/phase-wise-checklist/` - evidence-backed, phase-wise implementation checklists

## Dependency policy

Allowed runtime deps are pinned in `go.mod` (`v6`, `go 1.26.4` exact):

- `gin-gonic/gin`, `chinmay-sawant/gowkhtmltopdf` (pure-Go HTML), `bytedance/sonic`, `golang.org/x/sync`, `google.golang.org/api`. Indirect: `goccy/go-json`, `ugorji/go`.
- No Ghostscript. No browser. System prereqs only: Java 11+ (veraPDF), Node 18+ (frontend), Python 3.8+ (bindings).
- `policy.json` is GAR artifact retention only, not a PDF policy.
- Add deps only via `go mod tidy` plus `setup-auth.sh` flow. Frontend deps via `frontend/package.json` plus `npm ci`. Keep `go.sum` in sync.

---

## FAQ - does the root AGENTS.md get read automatically?

Yes. Creating `AGENTS.md` at the repo root is the standard, and it is read
automatically by every major tool: opencode, codex, gemini CLI,
antigravity/agy, grok, claude code, cursor. You do not need anything under
`.agents/`. Keep this one file at the root and keep it current.
