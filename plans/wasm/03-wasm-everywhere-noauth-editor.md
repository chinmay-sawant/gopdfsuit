# Plans/WASM - WASM Everywhere plus Public Editor Default

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; ref `frontend/src/utils/compressPdf.js`, `frontend/src/pages/Compress.jsx`, `internal/middleware/auth.go`, `frontend/src/App.jsx`
> **Status:** implemented - pending frontend build plus handler integration gates
> **Estimated effort:** M - one shared loader plus per-page wiring plus auth default hardening

---

## Overview

Use in-browser WASM on every page that can run locally (Merge, Split, Filler, Redact where feasible) following the Compress reference, and make Google auth opt-in so the Editor screen defaults to public (`auth required = 0`).

## Executive Summary

Compress is the only WASM-first page (`Compress.jsx:57-92` via `compressPDFSmart`, loader `compressPdf.js:1-181`, levels `compressLevels.js:1-57`, `COMPRESS_TRANSPORT` default `wasm`). All other pages call `makeAuthenticatedRequest` (`utils/apiConfig.js:78-138`) through `usePdfOperation` (`hooks/usePdfOperation.js:83-148`). Auth is already open unless Cloud Run: backend `authEnforced()` only when `REQUIRE_AUTH=1` or `K_SERVICE/K_REVISION` set (`middleware/auth.go:31-49`); frontend `isAuthRequired()` only when `VITE_IS_CLOUD_RUN=true` or `VITE_ENVIRONMENT=cloudrun` (`apiConfig.js:47-73`), with `/editor` as the sole `AuthGuard` route (`App.jsx:21-28`). This ledger hardens that default to explicit public and wires WASM page by page.

## Phase 1: Auth default to public (correctness first)

### 1.1 Frontend default

- [x] `frontend/src/utils/apiConfig.js:47-73` - public-by-default doc comments; `isAuthRequired()` stays opt-in via `VITE_IS_CLOUD_RUN/VITE_ENVIRONMENT` - proof: code diff
- [x] `frontend/src/App.jsx:22` - `EditorRoute` defaults to unwrapped `<Editor/>`, `AuthGuard` only when required - proof: code comment
- [x] `frontend/src/components/Navbar.jsx:12-14,153-226` - avatar and Sign Out hidden unless auth explicitly required - proof: code comments
- [x] `frontend/.env.example:1-13` - public default documented with `VITE_IS_CLOUD_RUN=false` commented and `VITE_GOOGLE_CLIENT_ID` optional - proof: doc diff
- [x] `frontend/src/components/AuthGuard.jsx:4-81` plus `contexts/AuthContext.jsx:22-149` - kept as opt-in Cloud Run path only - proof: no behavior change when disabled

### 1.2 Backend default

- [x] `internal/middleware/auth.go:31-49` - verified `REQUIRE_AUTH` unset or `0` plus `K_SERVICE/K_REVISION` unset means public `c.Next()`; `GoogleAuthMiddleware` kept - proof: `go test ./internal/middleware` ok
- [x] `internal/handlers/handlers.go:92-102` - env approach documented over code removal - proof: comment added
- [x] `documentation/AUTHENTICATION.md` - public-default section plus `REQUIRE_AUTH=1` local enforcement test - proof: doc diff

## Phase 2: Shared WASM loader

- [x] `frontend/src/utils/wasmLoader.js` (new) - generalize `compressPdf.js:31-110` for `compress.wasm` and `gopdfsuit.wasm` - proof: `compressPdf.js` delegates to it, `node --check` ok
- [x] `frontend/src/hooks/usePdfOperation.js:148` - extend `runLocal` pattern plus new `runLocalMulti` for Split - proof: code diff
- [~] Web Worker - deferred with reason (current WASM blocks main thread) - proof: header note in `wasmLoader.js`

## Phase 3: Per-page wiring

- [x] `frontend/src/pages/Merge.jsx:40` - `mergePDFSmart` via `runLocal` with server consent fallback - proof: code diff
- [x] `frontend/src/pages/Split.jsx:35` - `splitPDFSmart` via `runLocalMulti` with consent fallback - proof: code diff
- [x] `frontend/src/pages/Filler.jsx:28` - `fillPDFSmart` via `runLocal` with consent fallback - proof: code diff
- [x] `frontend/src/pages/Redaction.jsx:103,257,343` - page-info via client `pdfjs`, search/apply flagged no-privacy-win until engine lands - proof: code comments plus sidebar note
- [ ] `frontend/src/pages/Viewer.jsx:24,34,55` plus `Editor.jsx:575,627` (`POST /api/v1/generate/template-pdf`, `GET template-data`, `GET/POST /api/v1/fonts`) - `[~]` deferred: generator is portable but bundle plus font-asset cost high; keep server - proof: pointer to `plans/wasm/01-full-wasm-port.md` Phase 2.2
- [ ] `frontend/src/pages/HtmlToPdf.jsx:34` plus `HtmlToImage.jsx:33` - `[~]` server-only: Chromium today, `gowkhtmltopdf` after `plans/wasm/02-gowkhtmltopdf-replace.md`, never WASM - proof: pointer to `02` ledger

## Phase 4: Cleanup and docs

- [x] `frontend/src/utils/apiConfig.js:78-138` - `OFFLINE_DEMO_MESSAGE` plus GitHub Pages block vs offline WASM path documented - proof: code comment
- [x] `frontend/src/components/documentation/content/api-reference.js:8-286` plus `components/home/APIOverviewSection.jsx:5-15` - browser-local vs server-only `transport` labels - proof: doc diff
- [x] `frontend/src/utils/compressLevels.js:53-55` - `VITE_WASM_TRANSPORT` env matrix documented - proof: code comment

## Phase 5: Closure gates (docs-only now, code later)

- [ ] No lint/test gates for this planning change per checklist Required Checks (docs-only `plans/*.md`) - proof: none required
- [ ] On implementation: `cd frontend && npm run build` (never hand-edit `docs/`) plus `make test-integration` when handlers change - proof: pasted output then

## Dependencies

- Reference: `frontend/src/pages/Compress.jsx:1-256`, `utils/compressPdf.js:1-181`, `utils/compressLevels.js:1-57`, `cmd/wasmcompress/main.go`, `sampledata/compress-js/*`
- Auth: `internal/middleware/auth.go:31-49`, `cors.go:10-25`, `handlers.go:92-102`, `cmd/gopdfsuit/main.go:17-30`, `frontend/src/{App.jsx,main.jsx,utils/apiConfig.js,components/AuthGuard.jsx,contexts/AuthContext.jsx,components/Navbar.jsx,pages/Editor.jsx}`, `documentation/AUTHENTICATION.md`
- Engine unblock: `plans/wasm/01-full-wasm-port.md` (Merge/Split/Fill/Redact bindings); `plans/wasm/02-gowkhtmltopdf-replace.md` (HTML stays server)
- Explicit non-goals: removing backend `GoogleAuthMiddleware`, WASM HTML rendering, server font-install in browser
