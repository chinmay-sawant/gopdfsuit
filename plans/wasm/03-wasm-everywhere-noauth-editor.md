# Plans/WASM - WASM Everywhere plus Public Editor Default

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; ref `frontend/src/utils/compressPdf.js`, `frontend/src/pages/Compress.jsx`, `internal/middleware/auth.go`, `frontend/src/App.jsx`
> **Status:** planning - only Compress is WASM-first; auth already public-by-default via env
> **Estimated effort:** M - one shared loader plus per-page wiring plus auth default hardening

---

## Overview

Use in-browser WASM on every page that can run locally (Merge, Split, Filler, Redact where feasible) following the Compress reference, and make Google auth opt-in so the Editor screen defaults to public (`auth required = 0`).

## Executive Summary

Compress is the only WASM-first page (`Compress.jsx:57-92` via `compressPDFSmart`, loader `compressPdf.js:1-181`, levels `compressLevels.js:1-57`, `COMPRESS_TRANSPORT` default `wasm`). All other pages call `makeAuthenticatedRequest` (`utils/apiConfig.js:78-138`) through `usePdfOperation` (`hooks/usePdfOperation.js:83-148`). Auth is already open unless Cloud Run: backend `authEnforced()` only when `REQUIRE_AUTH=1` or `K_SERVICE/K_REVISION` set (`middleware/auth.go:31-49`); frontend `isAuthRequired()` only when `VITE_IS_CLOUD_RUN=true` or `VITE_ENVIRONMENT=cloudrun` (`apiConfig.js:47-73`), with `/editor` as the sole `AuthGuard` route (`App.jsx:21-28`). This ledger hardens that default to explicit public and wires WASM page by page.

## Phase 1: Auth default to public (correctness first)

### 1.1 Frontend default

- [ ] `frontend/src/utils/apiConfig.js:47-73` - make `isAuthRequired()` return false unless explicitly enabled; keep `VITE_IS_CLOUD_RUN/VITE_ENVIRONMENT` as opt-in - proof: `VITE_IS_CLOUD_RUN` unset build shows Editor without login
- [ ] `frontend/src/App.jsx:22` - default `EditorRoute` to unwrapped `<Editor/>`; only wrap in `AuthGuard` when `isAuthRequired()` true - proof: `/editor` renders with no token
- [ ] `frontend/src/components/Navbar.jsx:12-14,153-226` - hide avatar and Sign Out when auth not required - proof: visual check on public build
- [ ] `frontend/.env.example:1-13` - document public default: `VITE_IS_CLOUD_RUN=false` commented, `VITE_GOOGLE_CLIENT_ID` optional - proof: doc diff
- [ ] `frontend/src/components/AuthGuard.jsx:4-81` plus `contexts/AuthContext.jsx:22-149` - keep as opt-in Cloud Run path only, no behavior change when disabled - proof: `VITE_IS_CLOUD_RUN=true` build still gates

### 1.2 Backend default

- [ ] `internal/middleware/auth.go:31-49` - keep `GoogleAuthMiddleware` but verify `REQUIRE_AUTH` unset or `0` plus `K_SERVICE/K_REVISION` unset means public `c.Next()` - proof: `curl /api/v1/fonts` 200 with no token locally
- [ ] `internal/handlers/handlers.go:92-102` - no removal of `GoogleAuthMiddleware`; document env approach over code removal - proof: comment or docs link
- [ ] `documentation/AUTHENTICATION.md` - add public-default section plus `REQUIRE_AUTH=1` local enforcement test - proof: doc diff

## Phase 2: Shared WASM loader

- [ ] `frontend/src/utils/wasmLoader.js` (new) - generalize `compressPdf.js:31-110` (`loadWasmExec`, `instantiateStreaming` plus `arrayBuffer` fallback, `go.run`, 15s wait) for `compress.wasm` today and `gopdfsuit.wasm` from `plans/wasm/01-full-wasm-port.md` - proof: Compress works unchanged on new loader
- [ ] `frontend/src/hooks/usePdfOperation.js:148` - extend `runLocal` pattern (used by Compress plus Redaction apply) to Merge, Split, Filler - proof: no `makeAuthenticatedRequest` call on happy path
- [ ] Web Worker - move WASM call off main thread (current compress blocks UI) - proof: Lighthouse or manual jank note in ledger; `[~]` if deferred with reason

## Phase 3: Per-page wiring

- [ ] `frontend/src/pages/Merge.jsx:40` (`POST /api/v1/merge FormData pdf[]`) - add `goMergePDF` via `runLocal`, keep server consent fallback like `Compress.jsx:83-92` - proof: merge two PDFs offline in build
- [ ] `frontend/src/pages/Split.jsx:35` (`POST /api/v1/split` plus `pages/max_per_file`) - add `goSplitPDF` via `runLocal`, multi-file download from JS array - proof: split fixture offline
- [ ] `frontend/src/pages/Filler.jsx:28` (`POST /api/v1/fill` pdf plus xfdf) - add `goFillPDF` via `runLocal` after `01` Fill binding lands - proof: fill `sampledata/filler/*` offline
- [ ] `frontend/src/pages/Redaction.jsx:103,257,343` - page-info via client `pdfjs` (already `react-pdf`), search plus apply via WASM text path where engine allows; current `runLocal(request(...))` still uploads so flag no privacy win until engine lands - proof: page-info with no network in devtools
- [ ] `frontend/src/pages/Viewer.jsx:24,34,55` plus `Editor.jsx:575,627` (`POST /api/v1/generate/template-pdf`, `GET template-data`, `GET/POST /api/v1/fonts`) - `[~]` deferred: generator is portable but bundle plus font-asset cost high; keep server - proof: pointer to `plans/wasm/01-full-wasm-port.md` Phase 2.2
- [ ] `frontend/src/pages/HtmlToPdf.jsx:34` plus `HtmlToImage.jsx:33` - `[~]` server-only: Chromium today, `gowkhtmltopdf` after `plans/wasm/02-gowkhtmltopdf-replace.md`, never WASM - proof: pointer to `02` ledger

## Phase 4: Cleanup and docs

- [ ] `frontend/src/utils/apiConfig.js:78-138` - document `OFFLINE_DEMO_MESSAGE` plus GitHub Pages block vs new offline WASM path - proof: doc comment
- [ ] `frontend/src/components/documentation/content/api-reference.js:8-286` plus `components/home/APIOverviewSection.jsx:5-15` - label each endpoint browser-local vs server-only - proof: doc diff
- [ ] `frontend/src/utils/compressLevels.js:53-55` - keep `VITE_COMPRESS_TRANSPORT` semantics (`wasm` default, `server` opt-in) as template for `VITE_WASM_TRANSPORT` - proof: env matrix in ledger

## Phase 5: Closure gates (docs-only now, code later)

- [ ] No lint/test gates for this planning change per checklist Required Checks (docs-only `plans/*.md`) - proof: none required
- [ ] On implementation: `cd frontend && npm run build` (never hand-edit `docs/`) plus `make test-integration` when handlers change - proof: pasted output then

## Dependencies

- Reference: `frontend/src/pages/Compress.jsx:1-256`, `utils/compressPdf.js:1-181`, `utils/compressLevels.js:1-57`, `cmd/wasmcompress/main.go`, `sampledata/compress-js/*`
- Auth: `internal/middleware/auth.go:31-49`, `cors.go:10-25`, `handlers.go:92-102`, `cmd/gopdfsuit/main.go:17-30`, `frontend/src/{App.jsx,main.jsx,utils/apiConfig.js,components/AuthGuard.jsx,contexts/AuthContext.jsx,components/Navbar.jsx,pages/Editor.jsx}`, `documentation/AUTHENTICATION.md`
- Engine unblock: `plans/wasm/01-full-wasm-port.md` (Merge/Split/Fill/Redact bindings); `plans/wasm/02-gowkhtmltopdf-replace.md` (HTML stays server)
- Explicit non-goals: removing backend `GoogleAuthMiddleware`, WASM HTML rendering, server font-install in browser
