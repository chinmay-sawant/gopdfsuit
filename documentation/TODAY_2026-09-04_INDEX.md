# Today 2026-09-04 index

Branch: feat/builder-snippets. Base work plus master churn from the same day.

## Branch commits

- 4427b94 feat: add copy-clip phase-wise checklist plan
- 449200c chore: fluent font plus cell builder for gopdflib and pypdfsuit
- 48c57d2 chore: add round3 builder-snippets architecture review ledger plus visual report
- d3233e8 feat: implement round3 architecture review ledger phases 1-6, all rows validated

Master churn same day: wasm first viewer editor plus offline templates and fonts at 399e78e, wasm seams plus gowkhtmltopdf swap plus wasm everywhere wiring at 091d9ee, frontend wasm split plus fonts plus in-browser compliance at 5195ccb, gate fixes plus ledger sync, round2 residual friction ledger, architecture deepening review, REQUIRE_AUTH plus GIN_FAST_API plus request limits at eae7a7a, Go review fixes, operational docs move to documentation at ddf377a.

Diff master to HEAD is 193 files, 8844 insertions, 6040 deletions.

## Files added here

- BUILDER_FLUENT_GO.md: Go DocumentBuilder, Font chain, Cell chain, props grammar, helpers, canonical JSON, tests.
- PY_BUILDER_PARITY.md: pypdfsuit.builder exports, usage, gaps vs Go, types and CGO sink, tests.
- ENGINE_ROUND3_REFACTOR.md: PageManager plus Allocator, generation split, vector policy, html_convert dedup, EmitRowCells seam, parity pin, gates.
- WASM_VIEWER_EDITOR.md: compress.wasm vs gopdfsuit.wasm, core plus envelope plus transports seam, offline vs server matrix, Viewer and Editor flows, verify.
- HTML_PUREGO_SWAP.md: gochromedp to gowkhtmltopdf v0.2.5, field mapping, limits, security guards, fixtures.
- AUTH_REQUEST_LIMITS_TODAY.md: REQUIRE_AUTH, GIN_FAST_API scope, route table, byte caps, concurrency, SSRF, cookbook.
- FRONTEND_WASM_SPLIT.md: bundle map, loader and cache, fonts, templates, in-browser compliance, compress page, manifest gate.
- COMPLIANCE_PIPELINE_TODAY.md: manifest layout, veraPDF hard gate, structure tree hard gate, avalpdf warn only, commands and env.
- HANDLERS_FASTPATH_ENVELOPE.md: route table, pooled decode plus borrowed render, HFT tier, error envelope shape, per op notes.
- GATES_BENCHMARKS_TODAY.md: make gates, test layers, benchmark numbers with date pins, pool inventory with bounds, reproduce commands.
- EDITOR_SNIPPETS_TYPST.md: snippet JSON shape, text copy vs structural paste, model fields, Go and Python snippets, Typst renderer.

## How to verify

Run make fmt plus make lint plus make test locally. Run make test-integration for handler suite. Run make test-verify-pdfs for 10 file manifest. Run make wasm-compress plus frontend build for WASM artifacts.

Style: plain words, sentence case headings, no em dashes, code refs carry file and line.
