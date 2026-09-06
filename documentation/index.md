# Documentation index

documentation/ is truth for current behavior per plans/adr-2026-09-04-doc-homes.md. guides/ is a frozen archive. docs/ is Vite build output, never hand edit. plans/ holds decisions, mapped by plans/INDEX.md.

## Start here

- GETTING_STARTED_GOPDFLIB.md: Go library quickstart.
- PDF_SUITE_GUIDE.md: detailed Go plus Python plus REST suite guide with copy-paste samples.
- FEATURES.md: every feature with snippets, all ops in one page.
- ARCHITECTURE.md: gopdfsuit server, gopdflib plus engine, pypdfsuit, frontend WASM.
- TEMPLATE_REFERENCE.md: template JSON shape, props grammar, aliases.
- TODAY_2026-09-04_INDEX.md: what changed on 2026-09-04 on feat/builder-snippets, file map for the batch below.

## Operate

- DEPLOYMENT_CHECKLIST.md: ship checklist.
- AUTHENTICATION.md: Google auth, REQUIRE_AUTH, scopes.
- TROUBLESHOOTING_AUTH.md: auth failures and fixes.
- AUTH_REQUEST_LIMITS_TODAY.md (2026-09-04): env matrix, GIN_FAST_API scope, byte caps, concurrency, SSRF guards, cookbook.
- CACHING_AND_MEMORY_LIFECYCLE.md: pool and cache lifecycle.
- CROSS_REQUEST_CACHING_FAQ.md: cross request cache questions.
- BENCHMARKS.md: benchmark numbers and method.
- INTEGRATION_AND_BENCHMARK_TESTS.md: integration layers and older bench numbers.
- GATES_BENCHMARKS_TODAY.md (2026-09-04): make gates, test tiers, pool inventory with bounds, reproduce commands.

## API and guides (routers with copy-paste snippets)

- gopdflib-api.md: every `pkg/gopdflib` op with Go snippets.
- pypdfsuit-api.md: every `pypdfsuit` function with Python snippets.
- rest-api.md: every `/api/v1/*` endpoint with curl, auth, envelope, caps.
- engine-overview.md: in-memory pipeline (decode, render, fonts, tagging, XMP, sign, encrypt).
- compress-guide.md: Light/Medium/Heavy tiers, limits, passthrough rules.
- html-guide.md: pure-Go conversion fields, defaults, limits.
- document-ops-guide.md: merge, split page specs, XFDF fill, redact modes.
- wasm-frontend.md: local-first browser app, artifacts, consent, offline.
- compliance-benchmarks.md: PDF/A-4, PDF/UA-2, signatures, validation and bench commands.
- samples-catalog.md: every `sampledata/` directory with run commands.

## Build PDFs

- BUILDER_FLUENT_GO.md (2026-09-04): DocumentBuilder, Font chain, Cell chain, props grammar, helpers.
- PY_BUILDER_PARITY.md (2026-09-04): pypdfsuit.builder exports, usage, gaps vs Go.
- EDITOR_SNIPPETS_TYPST.md (2026-09-04): editor snippet JSON, text copy vs structural paste, model fields, Typst math cells.
- DIGITAL_SIGNATURE_RSA_ECDSA.md: RSA and ECDSA signing.
- HTML_CONVERSION.md: HTML conversion behavior and limits.
- HTML_PUREGO_SWAP.md (2026-09-04): gochromedp to gowkhtmltopdf swap, field mapping, fidelity gaps.

## Engine and handlers

- ENGINE_ROUND3_REFACTOR.md (2026-09-04): PageManager plus Allocator, generation split, vector policy, EmitRowCells seam, parity pin.
- HANDLERS_FASTPATH_ENVELOPE.md (2026-09-04): route table, pooled decode plus borrowed render, HFT tier, error envelope.

## Frontend and wasm

- WASM_VIEWER_EDITOR.md (2026-09-04): compress.wasm vs gopdfsuit.wasm, consent-first transports, Viewer and Editor flows.
- FRONTEND_WASM_SPLIT.md (2026-09-04): bundle map, offline fonts and templates, in-browser compliance, manifest gate.

## Compliance

- PDF_VALIDATORS.md: veraPDF, structure tree, avalpdf setup and use.
- COMPLIANCE_PIPELINE_TODAY.md (2026-09-04): manifest layout, hard gates vs warn-only, commands and env.
