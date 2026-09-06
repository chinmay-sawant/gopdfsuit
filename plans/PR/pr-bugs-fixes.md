## Summary

- Fixes browser merge silently dropping a file's pages when a decoy `/Root` mention precedes the real trailer, and unblocks PDF/A-compliant generation in WASM when fonts are already JS-registered.
- Also removes the stray `jsx` attribute on `<style>` nodes across five frontend pages and rebuilds both WASM bundles so browsers pick up the fixes.

---

## Motivation / context

- Plans: none (single-commit bugfix branch, no design doc)
- Issues: see **Related issues**

---

## Changes

### Merge catalog resolution

- `internal/pdf/merge/annotations.go`: `findRootRef` now takes the object map, collects all `/Root N G R` matches, walks newest-first, and returns the first candidate that resolves to a real catalog object (body contains `/Pages` or `/Type/Catalog`). Dangling references from literal content-stream text or stale incremental trailers are skipped instead of silently zeroing a file's pages.
- `internal/pdf/merge/merger.go`: `findCatalogAndPages` and `extractPagesFromTree` pass `objMap` into the hardened lookup.
- New `internal/pdf/merge/root_ref_test.go`: crafts a 1-page PDF whose content stream draws the literal text `(/Root 99 0 R)` before the real trailer, asserts `findRootRef` returns `1 0`, and asserts merging the doc with itself yields 2 pages.

### PDF/A font provisioning in WASM

- `internal/pdf/font/pdfa.go`: `RegisterLiberationFontsForPDFA` now checks `neededFacesRegistered` first and skips `EnsureFontsAvailable` when every needed face is already in the registry. Under `GOOS=js` the ensure step always rejects (no net/http, no writable fonts dir), so requiring it first made every compliant browser generation fail even with a fully primed registry.
- New `internal/pdf/font/pdfa_provision_test.go`: mirrors the browser flow with vendored `frontend/public/fonts` TTF bytes via `RegisterFontFromData`, asserts provision succeeds without attempting ensure, plus the negative case asserting ensure still runs when faces are missing.

### Frontend style tag cleanup

- `frontend/src/components/Toast.jsx`, `frontend/src/pages/Editor.jsx`, `frontend/src/pages/Merge.jsx`, `frontend/src/pages/Split.jsx`, `frontend/src/pages/Viewer.jsx`: `<style jsx>` changed to `<style>`. The repo does not use styled-jsx, so the `jsx` prop leaked onto DOM nodes and tripped React unknown-prop warnings.

### Artifacts and hygiene

- Rebuilt `frontend/public/compress.wasm` and `frontend/public/gopdfsuit.wasm` so browsers get the engine fixes.
- Removed 3 stray root outputs (`zerodha_active_output.pdf`, `zerodha_hft_output.pdf`, `zerodha_retail_output.pdf`).

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Negligible. `findRootRef` now uses `FindAllSubmatch` plus a bounded newest-first walk with map lookups; font path skips a failing ensure call in the primed case. |
| **Memory** | Negligible. One extra match slice per merge parse, freed after lookup. |
| **Behavior / correctness** | Merge keeps pages when decoy `/Root` text or stale trailers precede the real trailer. PDF/A browser generation succeeds when JS pre-registers Liberation faces; server behavior unchanged when faces are missing (ensure still runs). |
| **API (`/api/v1/*`) / UI** | No API shape change. UI drops the `jsx` unknown-prop warning on five pages. |
| **Dependencies** | None added. |
| **Binary size / build time** | WASM bundles slightly larger (compress ~8.63MB to ~8.65MB, gopdfsuit ~31.86MB to ~31.89MB). No Go build-time impact. |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | No emitter change. Browser PDF/A path unblocked; previously it errored before emitting. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [x] `go test -count=1 ./internal/pdf/merge/... ./internal/pdf/font/...` - PASS (`ok merge 0.210s`, `ok font 0.066s`, includes new `TestMergeKeepsPagesWhenDecoyRootPrecedesTrailer`, `TestRegisterLiberationSkipsEnsureWhenFacesRegistered`, `TestRegisterLiberationStillEnsuresWhenFacesMissing`)
- [x] `go vet ./internal/pdf/merge/... ./internal/pdf/font/...` plus `gofmt -l` on both dirs - clean, exit 0
- [x] `cd frontend && npm run lint` - PASS, exit 0, zero ESLint warnings
- [x] `make test` (full: `go test ./...` plus Python bindings plus `test/verify_pdfs.sh`) - PASS, exit 0 (unit plus integration plus python plus veraPDF 10 passed, 0 failed)
- [x] `make test-integration` (`go test -count=1 -v ./test`) - PASS, exit 0 (`TestIntegrationSuite` plus `TestZerodhaPDFCompliance`, 17.5s)
- [x] `make test-verify-pdfs` - PASS, exit 0 (veraPDF 10 passed, 0 failed; PDF/A-4 plus PDF/UA-2)
- [x] `cd frontend && npm run build` - PASS, exit 0 (Vite built in 3.54s, 1465 modules)
- [x] `make wasm-compress` - PASS, exit 0 (rebuilt `frontend/public/compress.wasm`, wasm binary verified)

### Commands

```sh
make fmt && make lint
make test
# plus when relevant:
make test-integration
make test-verify-pdfs
cd frontend && npm run build
make wasm-compress
```

---

## Screenshots / sample output

```
ok  	github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge	0.210s
ok  	github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font	0.066s
VET_FMT_EXIT:0
> gopdfsuit-frontend@1.0.0 lint
> eslint . --ext js,jsx --report-unused-disable-directives --max-warnings 0
(exit 0)
make test-integration: PASS (TestIntegrationSuite plus TestZerodhaPDFCompliance, 17.544s)
make test-verify-pdfs: PASS (veraPDF 10 passed, 0 failed, PDF/A-4 plus PDF/UA-2)
make test: PASS exit 0 (go plus python plus verify_pdfs.sh)
frontend build: vite v5.4.20, 1465 modules transformed, built in 3.54s
wasm-compress: WebAssembly (wasm) binary module version 0x1 (MVP)
```

---

## Related issues

- Relates to #87 (regex-based detection in merge code; this fix hardens the `/Root` regex lookup against decoy matches without implementing the lexer that issue asks for)
- No `Closes` keyword: no pre-existing tracking issue covers the WASM PDF/A provisioning failure or the `style jsx` cleanup, so no keyword applied rather than a wrong link.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- Issue #87 long-term work (lexer instead of regex for merge field detection) remains open; this PR only hardens the `/Root` lookup.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
