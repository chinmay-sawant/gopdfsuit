## Summary

Ships fluent font/cell builders for `pkg/gopdflib` plus `pypdfsuit` with Go/Python snippet parity, and implements the round3 builder-snippets architecture review (canonical props grammar, handler seams, frontend op-shell/WASM envelope, engine splits, compliance manifest).

---

## Motivation / context

- Plans: `plans/builder-snippets/plan.md`, `plans/gopdflib/fluent-builder.md`, `plans/builder-snippets/reviews/architecture-review-2026-09-04-builder-snippets.md`
- Issues: see **Related issues**
- Branch `feat/builder-snippets` continues the builder-snippets track plus round3 review phases 1-6. All ledger rows validated per the review handoff.

---

## Changes

### Area 1 - Fluent builder plus snippets (Go + Python)

- New `pkg/gopdflib/fontbuilder.go` with `Font(name)` and `Text(s)` chains plus `Props()`/`Cell()`/`Build()` terminals, byte-identical to `MakeProps` output.
- New `pkg/gopdflib/builder.go` document/table/cell helpers (`NewDocument`, `AddTitle`, `AddTable`, `AddSpacer`, `Build`, `Generate`, `MakeProps`, `NewCell`, `HeaderCell`, `SetCellFont`, `SetCellColor`, `AddBracketText`).
- New `pkg/gopdflib/props.go` typed props core (`Align`, `Color`, `Borders`, `FontOpts`, `ParseFontOpts`).
- Python parity in `bindings/python/pypdfsuit/builder.py` (`Font`/`Text` chains, `TemplateBuilder`, `make_props`, `new_cell`, `header_cell`, `set_cell_*`, `add_bracket_text`), plus `types.py` helpers and compress wiring via CGO.
- Samples `sampledata/gopdflib/builder-snippets/main.go` and `sampledata/pypdflib/builder-snippets/main.py` plus `sampledata/builder-snippets/snippet.json`, Go/Python PDFs byte-identical except dates/IDs.
- Docs: `documentation/BUILDER_FLUENT_GO.md`, `documentation/PY_BUILDER_PARITY.md`, `TEMPLATE_REFERENCE.md` fluent section, `pkg/gopdflib/doc.go` examples.

### Area 2 - Round3 engine, handlers, bindings refactor

- Canonical props grammar fallback policy unified across `internal/pdf/utils.go`, `pkg/gopdflib/props.go`, `fontbuilder.go`.
- Engine: shared `internal/pdf/html_convert.go`, `TableTagger` removal with `StructureManager.EmitRowCells`, `PageManager` inline plus bound-only `Allocator`, `generator.go` phase split into `generation` struct, new `internal/pdf/vector/` shared by Typst and SVG.
- Handlers: generic `decodeJSONBody[T]`, `readSingleUpload` across 5 handlers, `FastGenerateService` seam plus mocks, pure `parseRedactApply` parser, single error taxonomy via `gopdflib.CodeOf` plus `ClassifyMessage`.
- Bindings: fixed `Cell.checkbox` wire spelling mapping, `_to_dict` table plus parity tests, `_bindings.py` `require_bytes`/`json_payload`/`pdf_args`/`merge_args` helpers, CGO ownership notes, `TemplateBuilder` snapshot fix.

### Area 3 - Frontend, WASM, compliance, ops

- Frontend op-shell extraction (`OpPageShell`, `FileDropzone`, `ConsentBanner`), `HtmlConvertPage({mode})`, WASM `callWasm` envelope in `core.js`, consent transport unification, build-time font/template manifests, `PdfPreview` removal, Editor split (`documentModel.js`, `useEditorShortcuts`, `useBundledTemplate`, `useFonts`).
- Compliance: new `test/compliance_manifest.json`, `verify_pdfs.sh` runner thinning, per-directory strictness, `INTEGRATION_AND_BENCHMARK_TESTS.md` pyramid, suite outputs to temp dirs.
- Ops: `plans/INDEX.md`, wasm doc rename to `02b-`, PR ledger merge, doc-home ADR (`documentation/` truth, `guides/` frozen, `docs/` generated, `plans/` decisions), makefile test tiers, sampledata `temp_*` gitignore plus un-tracked generated files, CI split with shared validator action plus Go 1.26.4 pin.
- Topic docs `documentation/TODAY_2026-09-04_INDEX.md` plus builder/engine/wasm/html/auth/compliance notes, `FEATURES.md`, `ARCHITECTURE.md`, `index.md` links.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Drops per-row `NewTableTagger` alloc on hot path, hoists redact JSON translation out of loops, unifies image-decode/emit loops. No benchmark deltas claimed. |
| **Memory** | Allocator bound-only, standalone state kept as test fake. Server pooled buffers and tier prealloc unchanged. |
| **Behavior / correctness** | Builder overlays emit current `Props`/hex strings, no engine draw change in v1. HTML phantom fields mapped with non-empty `Options` warning. `chequebox` wire spelling kept with Python `checkbox` mapping documented. WASM split documented permanent (worker bundle vs full engine). |
| **API (`/api/v1/*`) / UI** | No breaking API change. UI adds op-shell pages, snippet copy paths, consent transport, offline manifests. |
| **Dependencies** | No new runtime deps. `sampledata/go.mod`/`go.sum` resynced for snippet runs. |
| **Binary size / build time** | No binary size claim. Requires `make wasm-compress` before frontend build where applicable. |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | Tagging/MCID/ParentTree paths preserved. `make test-verify-pdfs` 10/10 per ledger. Zerodha-strict avalpdf stays warn-only (pre-existing heading/TH findings). |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Additive builders and internal seams. Existing `NewCell(text, props)`, `MakeProps`, `make_props`, `new_cell(text, props)` signatures and outputs byte-identical. |

---

## Test plan

- [x] `make test` (`go test ./...` plus Python bindings plus `test/verify_pdfs.sh`)
- [x] `make test-integration` (`go test -count=1 -v ./test`) when handlers or engine changed
- [x] `make lint` plus `go vet` (zero ESLint warnings in `frontend/`)
- [x] `make build` (`go build -o bin/app ./cmd/gopdfsuit`) when shippable change
- [x] `make test-verify-pdfs` or `make test-scan-pdfs-compliance` when PDF output changed
- [x] `cd frontend && npm run build` when UI changed (never hand-edit `docs/`)
- [x] `make wasm-compress` when `cmd/wasmcompress/` changed

### Commands

```sh
make fmt && make lint
make test
# plus when relevant:
make test-integration
make test-verify-pdfs
```

### Session results (2026-09-04, branch `feat/builder-snippets`, PR #210)

```
make lint                        exit 0 (golangci-lint plus frontend eslint, zero warnings)
go vet ./...                     exit 0
go build -o bin/app              exit 0 (bin/app 48 MB)
make wasm-compress               exit 0 (compress.wasm valid MVP module)
cd frontend && npm run build     exit 0 (vite 8.69s, manifests in sync: 12 fonts, 3 templates)
make test-integration            PASS ok test 17.446s
make test-verify-pdfs            Totals: 10 passed, 0 failed, 10 checks
make test                        exit 0 (test-go plus test-python plus test-verify,
                                 post-test veraPDF/structure-tree validation passed)
```

---

## Screenshots / sample output

```
Branch feat/builder-snippets vs origin/master: 208 files changed, +10321/-6040.
Commits: 4427b94 feat copy-clip checklist, 449200c fluent builder, 48c57d2 round3 ledger,
d3233e8 round3 phases 1-6, 76349c8 topic docs, 765fe6c docs index, 20465e0 FEATURES+ARCHITECTURE.
```

Verified this session on the PR head, all exit 0: see Session results under Test plan
(`make lint`, `go vet`, `make build`, `make wasm-compress`, frontend build,
`make test-integration` 17.446s, `make test-verify-pdfs` 10/10, `make test`).

---

## Related issues

- None

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`, `documentation`)
- [x] Related issues filled with real ticket IDs (none per author choice)
- [x] Filled body committed under `plans/PR/pr-builder-snippets-round3.md`

---

## Follow-ups (out of scope)

- Full sampledata `{fixtures,outputs,goldens}` split (minimal `temp_*` ignore variant shipped here).
- Redact redesign to drop react-pdf.
- `financial_report.pdf` dual-purpose reroute.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
