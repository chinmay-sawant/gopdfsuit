# Frontend - minimal product site and tool workspaces

> **Parent:** `plans/INDEX.md` - canonical frontend execution ledger; read alongside `plans/wasm/03-wasm-everywhere-noauth-editor.md`, `plans/wasm/04-frontend-wasm-split-fonts-compliance.md`, and `plans/adr-2026-09-04-doc-homes.md`
> **Status:** planned. Discovery is complete. No application behavior has changed.
> **Estimated effort:** L - six phases, with the editor protected in the first delivery.

---

## Overview

Rebuild `frontend/` as a calm, minimal GoPdfSuit product site with focused tool
workspaces. The present gradient, glass-card, particle-animation look is not a
constraint. The existing tool behavior is.

The public site should explain the product, make its tools easy to find, and
give credible proof through a curated comparison and fresh screenshots. Each
tool then opens a workbench with a clear input, configuration, result, and
download path. `/editor` stays functionally and visually protected in this
first delivery. It is a real authoring workspace, not a marketing card.

In-app documentation is out of scope for the rebuilt application. Repository
Markdown under `documentation/` remains the documentation home. The frontend
must not retain a `/documentation` route, documentation tab, documentation
footer link, or the React documentation component tree after its unique content
has been accounted for.

## Executive summary

### Product shape

| Area | Intended place | Rule |
|---|---|---|
| Product home | `/` | Explain the template PDF product, show a compact tool catalogue, lead to Editor and the most useful tools, and make only sourced claims. |
| Tool workspaces | Existing `/viewer`, `/editor`, `/merge`, `/split`, `/compress`, `/filler`, `/htmltopdf`, `/htmltoimage`, `/redact` routes | Keep routes and runtime contracts. Rebuild the non-editor pages around one input, configure, result, and download pattern. |
| Product proof | `/comparison` and `/screenshots` | Keep only proof with a review date and source. Screenshots become fresh, local, versioned captures. |
| Documentation | `documentation/` | Move or retire the React documentation content after an explicit content matrix. Do not repackage it as another in-app tab. |

### Current capability map to preserve

| Route | User job | Required behavior to retain |
|---|---|---|
| `/viewer` | Render a JSON PDF template | Bundled, named, or uploaded template input; editable JSON; generate, preview, and download. |
| `/editor` | Build a JSON template visually | Palette, native component drag and drop, table-cell drops, reorder, properties, context menu, shortcuts, custom fonts, preview, JSON, and Go/Python snippets. |
| `/merge` | Combine PDFs | Multiple-file input, file removal, ordering, local-first processing, consent-only server fallback, preview, and download. |
| `/split` | Extract page ranges or chunks | One-file input, page range or maximum-pages controls, multi-file result handling, and download. |
| `/compress` | Reduce PDF size | One-file input, Light/Medium/Heavy choice, size limit, before/after sizes, browser processing when configured, and explicit upload consent when needed. |
| `/filler` | Populate an AcroForm PDF | PDF plus XFDF/XML input, local-first processing, consent-only server fallback, preview, and download. |
| `/htmltopdf` | Turn HTML or a URL into PDF | HTML and URL modes, page controls, result preview, and the correct local versus server explanation. |
| `/htmltoimage` | Turn HTML or a URL into an image | HTML and URL modes, PNG/JPG options, dimensions and quality, image preview, and the correct local versus server explanation. |
| `/redact` | Permanently remove selected content | File input, page navigation, visual regions, text search, mode/password controls, apply, and download. |

The user terms "pillar" and "retart" do not appear in the code. This ledger
uses **Filler** and **Redact** as likely matches until the product naming is
confirmed in Phase 1.

### Design direction

Use a document-workbench system rather than another animated SaaS landing page:

- Paper or near-white surfaces, graphite text, one restrained blue action
  color, hairline borders, real empty space, and one sans-serif family.
- No particle background, neon gradients, glass panels, emoji navigation
  icons, or decorative motion. Motion is limited to brief feedback and honors
  `prefers-reduced-motion`.
- A short header groups tools rather than listing every route. Comparison and
  screenshots live under product proof, not among daily tool actions.
- Non-editor tools share a responsive workbench. On wide screens it has input
  and options beside output. On narrow screens it becomes a readable sequence.
  The Editor is deliberately not forced into that layout.
- Semantic tokens own all color, spacing, typography, borders, focus rings,
  z-index, and motion values. Components do not add one-off colors or inline
  layout systems.

## Phase 1: establish the product contract and protect current behavior

### 1.1 Complete the discovery record

- [x] `frontend/src/App.jsx`, `Navbar.jsx`, `pages/*`, `components/home/*`, and `index.css` - record the current route map, shared seams, documentation references, and design debt - proof: source audit on 2026-09-05.
- [x] `pages/Editor.jsx`, `components/editor/*`, `hooks/useEditorShortcuts.js`, and `components/editor/documentModel.js` - record the editor's drag/drop, keyboard, context-menu, template, and PDF paths before any shell work - proof: source audit on 2026-09-05.
- [x] `pages/{Viewer,Merge,Split,Compress,Filler,HtmlConvertPage,Redaction}.jsx` plus `hooks/usePdfOperation.js` - record the input, output, local processing, consent, preview, and download contracts for every retained tool - proof: source audit on 2026-09-05.

### 1.2 Resolve names, claims, and content ownership

- [ ] Product vocabulary - confirm whether the requested "pillar" and "retart" labels mean **Filler** and **Redact**, then use one approved noun in navigation, headings, CTAs, and tests - proof: accepted wording recorded in this ledger before UI copy changes.
- [ ] `frontend/src/components/home/*`, `pages/Comparison.jsx`, `components/PerformanceSection.jsx`, `documentation/FEATURES.md`, `documentation/BENCHMARKS.md`, `pkg/gopdflib/html.go`, and `internal/pdf/encryption/encrypt.go` - create a claim matrix with source path or external URL, review date, and allowed copy - proof: every retained homepage, comparison, and security claim has a source; unsourced claims are removed.
- [ ] `frontend/src/pages/Documentation.jsx`, `frontend/src/components/documentation/**`, and `documentation/**` - compare the React documentation content to the repository Markdown; move only unique, still-correct material into the existing documentation home or explicitly retire it - proof: migration matrix names the destination or retirement decision for every React documentation section.
- [ ] `frontend/src/pages/Editor.jsx` and `components/editor/**` - capture baseline desktop and narrow-screen screenshots plus keyboard and native drag/drop results before changing shared styles - proof: dated local QA evidence covers palette-to-canvas drop, insertion, reorder, table-cell drop, Alt+Arrow movement, preview, and JSON download.

### 1.3 Set non-negotiable boundaries

- [ ] `plans/wasm/03-wasm-everywhere-noauth-editor.md`, `plans/wasm/04-frontend-wasm-split-fonts-compliance.md`, `frontend/src/utils/wasm/**`, `frontend/src/hooks/usePdfOperation.js`, and `frontend/src/utils/apiConfig.js` - write a concise UI contract for each tool's local processing, server fallback, explicit consent, asset loading, and Cloud Run auth behavior - proof: the contract is reviewed against source and linked from this ledger; this redesign does not change it.
- [ ] `frontend/vite.config.js`, `cmd/gopdfsuit/main.go`, and `frontend/scripts/check-wasm-manifests.mjs` - preserve the `/gopdfsuit/` base path, generated `docs/` output, server asset paths, and prebuild manifest check unless a separate deployment decision changes them - proof: existing hosted route and manifest checks pass after the redesign.
- [ ] `frontend/src/pages/Editor.jsx`, `frontend/src/components/editor/**`, `frontend/src/hooks/useEditorShortcuts.js`, and `frontend/src/components/editor/documentModel.js` - declare the Editor a first-wave protected boundary; no feature or visual redesign inside it - proof: baseline checks in 1.2 match the final first-wave editor evidence.

## Phase 2: replace the site shell and remove in-app documentation

### 2.1 Build the minimal design system

- [ ] `frontend/src/index.css` and new `frontend/src/styles/{tokens,base,site,workspace}.css` - replace gradients, glass cards, particle rules, and repeated inline style conventions with semantic tokens and a small CSS layer structure - proof: no retained public or workspace component contains raw theme colors or a duplicated page-layout system.
- [ ] `frontend/src/components/site/{SiteHeader,ToolMenu,SiteFooter}.jsx` and `frontend/src/App.jsx` - create a compact header with Home, grouped Tools, Editor, Proof, theme control, and an external repository link; retain each existing tool URL - proof: keyboard and mouse navigation reaches every retained route without an overloaded all-tools bar.
- [ ] `frontend/src/App.jsx` and `frontend/src/components/{Navbar,BackgroundAnimation}.jsx` - remove the old global navigation and particle background only after the replacement shell is live - proof: no retained public or tool page imports the old chrome.
- [ ] `frontend/src/App.jsx`, `pages/Documentation.jsx`, `components/documentation/**`, `components/home/FooterSection.jsx`, and `components/Navbar.jsx` - remove the `/documentation` route, imports, Documentation and Performance navigation entries, footer links, and React documentation tree after the 1.2 migration matrix closes - proof: source search finds no in-app documentation route or component import; repository Markdown remains intact.
- [ ] `frontend/src/App.jsx` - add a useful not-found route that returns users to the product home or tool catalogue - proof: an unknown hash route renders a clear recovery action instead of only the header.

### 2.2 Make the shell accessible and cheap to load

- [ ] `frontend/src/App.jsx` and route modules - lazy-load tool, comparison, and screenshot routes while keeping the initial home bundle free of the 31.7 MB `gopdfsuit.wasm` and 8.6 MB `compress.wasm` downloads - proof: browser network evidence shows neither WASM file downloads until a tool needs it.
- [ ] `frontend/src/styles/*`, `components/site/*`, and every changed interactive component - supply visible keyboard focus, semantic landmarks, named icon buttons, 44px minimum targets, 4.5:1 normal-text contrast, and no focus obscured by sticky UI - proof: keyboard and automated accessibility checks pass at 375px, 768px, 1024px, and 1440px.
- [ ] `frontend/src/theme.jsx` and the new token layers - preserve a deliberate light and dark theme with the same action, focus, disabled, and error distinctions - proof: visual QA records both themes and reduced-motion mode.

## Phase 3: rebuild the public product and proof pages

### 3.1 Make the home page earn its place

- [ ] `frontend/src/pages/Home.jsx` and replacement `frontend/src/components/home/*` - replace the current hero, bento grid, API table, benchmark block, and footer with concise product framing, grouped tool entry points, a protected Editor story, local-processing transparency, and sourced proof links - proof: content review matches the Phase 1 claim matrix and every primary CTA lands on a real route.
- [ ] `frontend/src/pages/Home.jsx` - remove the runtime GitHub star request and render no invented availability or popularity state when that third-party request fails - proof: home works offline after its own static assets load.
- [ ] `frontend/src/components/home/*` and `pages/Comparison.jsx` - remove duplicate comparison and performance content so the home has one short proof summary and `/comparison` owns the full reviewed evidence - proof: each approved claim appears once per intended audience and links to its source.

### 3.2 Turn screenshots and comparison into trustworthy product proof

- [ ] `frontend/src/pages/Screenshots.jsx`, new `frontend/public/showcase/**`, and new `frontend/src/content/showcase.js` - replace mutable `raw.githubusercontent.com/.../master` image loading and the stale autoplay carousel with current, locally versioned captures that have tool name, viewport, capture date, alt text, and a fixed ordering - proof: browser network records no remote GitHub screenshot request; all captures match the new product shell and current pure-Go HTML conversion copy.
- [ ] `frontend/src/pages/Comparison.jsx` and new `frontend/src/content/claims.js` - build comparison from one reviewed data source that distinguishes product facts from external competitor statements and records URLs and review dates - proof: no price, benchmark, compliance, dependency, security, or competitor capability appears without a source and review date.
- [ ] `frontend/src/pages/Comparison.jsx` - remove or defer every comparison row that cannot be independently sourced, including time-sensitive pricing and competitor capability rows - proof: claim-matrix review has no unsourced published row.

## Phase 4: rebuild the non-editor tool workspaces

### 4.1 Make one workbench without erasing tool differences

- [ ] `frontend/src/components/{FileDropzone,OpPageShell,OperationShell,ConsentBanner}.jsx` and new `frontend/src/components/workspace/*` - separate reusable behavior from presentation and build the common input, configure, result, error, consent, and download states - proof: every non-editor tool has the same state vocabulary without changing its request or blob lifecycle.
- [ ] `frontend/src/pages/{Viewer,Merge,Split,Compress,Filler,HtmlConvertPage,Redaction}.jsx` - remove page-specific gradients, glass styling, emoji step markers, and ad-hoc inline layout after the shared workbench supports each needed state - proof: `npm run lint` is clean and a source audit finds no obsolete visual primitives in retained tool pages.
- [ ] `frontend/src/pages/Editor.jsx` plus editor styles isolated from the new site CSS - retain the editor's current three-column authoring experience rather than applying the common non-editor workbench - proof: Phase 1 editor baseline passes unchanged.

### 4.2 Preserve and make each workflow honest

- [ ] `frontend/src/pages/Viewer.jsx` - make bundled, named, uploaded, and pasted JSON template input understandable; retain generate, preview, download, local processing, and consent fallback - proof: browser QA succeeds for one bundled template, one JSON upload, and one edited template.
- [ ] `frontend/src/pages/Merge.jsx` - preserve multiple-file drop/select, remove, explicit ordering, merge, preview, and download; either add true file reordering or stop claiming it exists - proof: a manual test confirms output page order for both mouse and keyboard controls.
- [ ] `frontend/src/pages/Split.jsx` and `hooks/usePdfOperation.js` - present each produced PDF as a named downloadable result instead of implying that a single iframe represents a multi-file operation - proof: a multi-part split exposes every output and every download opens the expected file.
- [ ] `frontend/src/pages/Compress.jsx` - retain tier selection, max-size error, before/after size reporting, browser-first processing, and consent-only fallback; copy must say what the chosen transport actually does - proof: one successful local run and one consent fallback run show correct size and upload messaging.
- [ ] `frontend/src/pages/Filler.jsx` - use the shared accessible file-input behavior for both PDF and XFDF/XML while retaining file requirements and result flow - proof: mouse, keyboard, and drag/drop tests cover valid paired files and an incomplete pair.
- [ ] `frontend/src/pages/HtmlConvertPage.jsx`, `pages/HtmlToPdf.jsx`, and `pages/HtmlToImage.jsx` - keep HTML and URL modes distinct, expose only supported output/options, and state the pure-Go renderer's limits in plain language - proof: HTML and URL paths for both formats match the Phase 1 transport contract and source-backed copy.
- [ ] `frontend/src/pages/Redaction.jsx` - retain manual regions, text search, mode, password, apply, and output report; replace mouse-only canvas handling only if pointer and keyboard alternatives are proven - proof: desktop pointer, keyboard alternative, and touch emulation checks complete without losing the redaction queue.

## Phase 5: add regression coverage and prove the real interface

### 5.1 Test the paths the current unit tests do not reach

- [ ] `frontend/tests/**` and the chosen browser-test harness - add route, upload, consent, result, error, download, responsive, and accessibility coverage without selecting a new dependency until its maintenance and CI fit are documented - proof: focused tests cover every retained route and the locked local versus server behavior.
- [ ] `frontend/tests/**` - keep the existing document model and WASM envelope tests, then add direct regressions for split multi-file output and any chosen merge reorder behavior - proof: `npm test` covers the fixed regression path.
- [ ] `frontend/src/components/editor/**`, `pages/Editor.jsx`, and `hooks/useEditorShortcuts.js` - write or record manual browser evidence for native drag/drop and keyboard shortcuts that generic DOM tests cannot honestly simulate - proof: Phase 1 interaction matrix is rerun after the final CSS and route changes.

### 5.2 Run final gates and record only current evidence

- [ ] `frontend/` - run `npm run lint`, `npm test`, and `npm run build` after the last frontend edit; let Vite regenerate `docs/` and never hand-edit it - proof: command output with zero ESLint warnings, passing Node tests, and a successful manifest-checked build.
- [ ] repository root - run `make fmt`, `make vet`, `make lint`, and `make test` after the last code edit; add `make test-integration` if handler contracts changed and `make wasm-compress` if WASM artifacts changed - proof: each required command exits zero and its result is recorded beside this row.
- [ ] running application - capture the home, every tool, comparison, screenshots, not-found route, and protected Editor at 375px, 768px, 1024px, and 1440px in light, dark, and reduced-motion modes - proof: visual evidence confirms no clipping, horizontal overflow, stale claim, remote screenshot dependency, or hidden focus state.
- [ ] `plans/frontend/phase-wise-checklist.md` and `plans/INDEX.md` - close only the rows backed by the final source and validation evidence, list any deferred claim or workflow openly, and keep this as the only active frontend-redesign ledger - proof: no duplicate active frontend redesign row exists under `plans/`.

## Dependencies

- Runtime and consent behavior stays owned by `plans/wasm/03-wasm-everywhere-noauth-editor.md` and `frontend/src/utils/wasm/**`, `frontend/src/hooks/usePdfOperation.js`, and `frontend/src/utils/apiConfig.js`.
- WASM splitting, font assets, and the manifest check stay owned by `plans/wasm/04-frontend-wasm-split-fonts-compliance.md`.
- The active builder-snippets work may change Editor snippets. Coordinate around `plans/builder-snippets/plan.md`; do not redesign its controls in this first wave.
- Current repository documentation lives under `documentation/` by `plans/adr-2026-09-04-doc-homes.md`. `docs/` is generated Vite output, not authored content.
- This ledger owns frontend information architecture, visual system, proof content, documentation removal from React, and tool-workspace presentation. It does not change PDF rendering, request contracts, backend authentication, WASM bindings, or generated artifacts without a separately approved ledger.

## Open product decisions

- Confirm the final product names for Filler and Redact.
- Decide whether screenshots remain a dedicated Proof route or become a curated section of the home page after fresh captures exist.
- Decide whether a sourced, dated competitor table has enough value to retain. The minimal site can show product evidence without competitor pricing.
- Decide whether a public link to repository Markdown documentation belongs in the footer. It must not restore an in-app documentation route or tab.
