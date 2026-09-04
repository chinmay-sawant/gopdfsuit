# Architecture

Four systems share one engine. internal/pdf renders. pkg/gopdflib exposes it to Go. cmd/gopdfsuit serves it over HTTP. bindings/python wraps it for Python. cmd/wasm plus cmd/wasmcompress compile it to WASM for the frontend.

```
                     +-------------------+
                     |   internal/pdf    |
                     | engine (private)  |
                     +--------+----------+
                              |
        +---------------------+---------------------+
        |                     |                     |
pkg/gopdflib           cmd/gopdfsuit          bindings/python/cgo
Go facade              REST on :8080          C shared lib
        |                     |                     |
        +----------+----------+----------+----------+
                   |                     |
             cmd/wasm +            pypdfsuit
             cmd/wasmcompress      Python package
             WASM modules
                   |
            frontend/ (React)
```

Rule: external Go code imports pkg/gopdflib only, never internal/*.

## gopdfsuit server

Entry is cmd/gopdfsuit/main.go:17. Order: ResolveServerConfig, optional profiling, background math font fetch, WarmRuntime, gin release mode, NewRouter, NewServer on :8080, serve, graceful shutdown with 15s timeout.

Router is internal/handlers/router.go:160 NewRouter. It stacks RequestID, recovery as 500, concurrencyLimiter as 429 plus Retry-After 1, then RegisterRoutesWithPolicy at handlers.go:89. That registers static docs assets, the /api/v1 group with optional CORS plus always-on GoogleAuthMiddleware, localhost-only /debug/pprof, root redirect, and SPA fallback to docs/index.html.

Handlers stay thin. Each decodes input, calls PDFService, writes bytes. Service boundary is internal/handlers/services.go:33 PDFService plus 67 FastGenerateService. defaultPDFService at 71 maps every op to the engine: Generate to pdf.GenerateTemplatePDF, Fill to form, Merge and Split to merge, Compress to compress, Fonts to the registry, HTML to pdf convert functions. SetPDFService at 169 swaps in mocks.

Generate path at generate.go:10 shows the pattern all ops follow. Acquire a pooled PDFTemplate, preallocate from ContentLength plus X-Payload-Tier, cap the body at 8 MiB, decode, then render through the borrowed path with zero extra copy: GenerateTemplatePDFBorrowed returns a pooled buffer, handler writes it, Release returns it.

Decode tiers live in json_decode.go:49. HFT tier with known length under 8 MiB reads into hftBodyBufPool and runs JIT row unmarshal. Retail under 512 KiB reads into bodyBufPool with sonic unmarshal. Larger or unknown lengths stream decode with no copy. WarmRuntime at router.go:102 pretouches pools, JSON schema, and HFT path at startup.

Auth self gates in internal/middleware/auth.go:65. Open locally, enforced when REQUIRE_AUTH=1 or on Cloud Run. GIN_FAST_API=1 drops CORS only, never auth. Details in documentation/AUTH_REQUEST_LIMITS_TODAY.md.

## gopdflib and the engine

pkg/gopdflib is the public facade, one file per op: generator, merge, split, compress, fill, redact, html, plus builder, fontbuilder, props, adapter, errors, types, doc. It owns semantic validation and defaults, for example ParseCompressLevel at adapter.go:61. doc.go:1 states the policy: the builder is a thin overlay emitting the same props grammar, the sink stays GeneratePDF, CGO and handlers and WASM do transport-only guards.

internal/models holds the JSON schema truth: PDFTemplate, Config, Title, Table, Row, Cell, plus Security, PDFA, Signature configs. It imports nothing internal. pkg/gopdflib/types.go mirrors these for outside callers.

internal/pdf renders. Root package owns orchestration: generator.go with BorrowedPDF plus GenerateTemplatePDF at 359 plus borrowed variant at 373, pagemanager.go for pagination and object IDs, allocator.go as the bound-only ID seam, generation.go as the shared home for image decode plus font layout plus image emit, draw.go for content streams, structure.go for tagging, metadata.go for XMP, outline.go for bookmarks, destinations.go and links.go for navigation, image.go for raster plus SVG, html_convert.go for HTML mapping, pdf.go plus pdf_js.go for server vs WASM builds.

Subpackages own one concern each: font for TTF parse plus subset plus metrics plus registry plus PDF/A provisioning, pdfobj for low level object and xref serialization, merge for merge plus split, compress for tiers, form for XFDF, redact for search plus secure plus visual, encryption for AES, signature for PKCS#7 with RSA and ECDSA P-256, svg for path parsing, vector as the shared float and color policy leaf. typstsyntax outside pdf owns math layout and renders through vector.

Dependency direction has no cycles: models, pdfobj, vector are leaves. svg reads vector. font, encryption, signature read models. merge, compress, form read pdfobj. redact reads pdfobj plus encryption plus models. pdf root reads models, pdfobj, encryption, signature, font, svg. gopdflib reads models, pdf, merge, compress, form, redact. Nothing internal imports gopdflib back.

BorrowedPDF at generator.go:27 is the zero copy contract. Borrowed hands a pooled buffer, Bytes borrows, CopyBytes copies, Release returns exactly once. GenerateTemplatePDF wraps borrowed plus clone. Pools sit at every hot spot: template structs and decode slices in handlers, PDF buffers and xref slices and scratch in generator, zlib and subset and image caches in font, struct nodes in structure. Bounds and clear APIs are inventoried in documentation/GATES_BENCHMARKS_TODAY.md.

Round3 refactor notes live in documentation/ENGINE_ROUND3_REFACTOR.md.

## pypdfsuit

Call chain per op: Python dataclass with to_dict in types.py, json_payload in _bindings.py:325 dumps compact JSON bytes, call_bytes_result at 241 passes it through ctypes to the CGO export and copies the result out with string_at before FreeBytesResult, raise_for_error maps the error string to typed InvalidInput, LimitExceeded, Upstream, or Internal errors.

CGO layer is bindings/python/cgo/exports.go. Each export follows one shape, shown by GeneratePDF at 206: read the C string, sonic unmarshal to gopdflib.PDFTemplate, call the engine, pack bytes or envelope JSON error. Sixteen exports cover generate, merge, split, page spec, fill, compress, HTML both ways, fonts, redact five ways, plus two free functions.

Build runs bindings/python/build.sh: CGO_ENABLED=1 go build -buildmode=c-shared ./bindings/python/cgo/ into pypdfsuit/lib/libgopdfsuit.so. Runtime discovery in _bindings.py:119 checks pypdfsuit/lib first, then dev fallbacks, else tells you to run build.sh. Tests auto rebuild the .so when Go sources are newer unless PYPDFSUIT_SKIP_AUTO_BUILD=1. Full import surface and gaps vs Go are in documentation/PY_BUILDER_PARITY.md.

## Frontend and WASM

Two WASM builds from the same engine: cmd/wasmcompress/main.go exposes goCompressPDF only and ships as compress.wasm near 8M, cmd/wasm/main.go exposes generate, merge, split, fill, compress, redact, fonts, HTML and ships as gopdfsuit.wasm near 31M. The split stays so the small compress worker never loads the full engine.

JS seam is frontend/src/utils/wasm: core.js loads and caches through Cache API gopdfsuit-wasm-v1, envelope.js is the single call gate normalizing Uint8Array results or code plus message errors, one thin module per op, transports.js decides local vs server. Default VITE_WASM_TRANSPORT=wasm runs browser local first and reaches the server only after explicit consent in ConsentBanner. No silent uploads.

Shared hook usePdfOperation.js gives every page the same flow: run for server blobs, runJson for server JSON, runLocal for WASM bytes, runSmart for consent-first local with server fallback. OpPageShell plus FileDropzone give shared chrome and file input.

Routes in App.jsx: / for Home, /viewer, /editor with optional AuthGuard on Cloud Run, /merge, /split, /compress, /filler, /htmltopdf, /htmltoimage through one shared converter, /redact, /documentation, /comparison, /screenshots.

Build is Vite with base /gopdfsuit/ and outDir ../docs, so the app deploys as static output. prebuild runs check-wasm-manifests.mjs which fails on font or template manifest drift. Details in documentation/WASM_VIEWER_EDITOR.md and documentation/FRONTEND_WASM_SPLIT.md.
