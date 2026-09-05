# Go application review - structure and performance

> **Parent:** `AGENTS.md` and the user's Go-only review request, using `skills/phase-wise-checklist/SKILL.md`.
> **Status:** Review complete. Recommendations are unimplemented and unvalidated. All improvement gates remain open.
> **Estimated effort:** About 2-4 engineering weeks for the findings below, depending on parser fixtures and compatibility decisions. This is a planning estimate, not a commitment.

---

## Overview

Reviewed the current checkout on 2026-09-05. Scope included Go command entrypoints, HTTP handlers and middleware, public Go APIs, internal models and PDF engine packages, font utilities, Typst syntax, Go WASM adapters, the Go side of the Python CGO bridge, and relevant Go tests and benchmark drivers. Frontend code, Python implementation, vendored dependencies, and generated artifacts were outside the review.

Three parallel source reviews covered package/API structure, adapter control flow, and PDF operations. The coordinating review examined rendering, caching, allocation, benchmark evidence, and the reported call chains. This was a broad static review, not a claim that every line or every PDF-format rule was audited.

No Go code changed. No Git commands, benchmarks, test suites, formatters, or linters ran. The requested checklist skill explicitly excludes lint and test runs for documentation-only changes. Existing test names below identify coverage or future proof, not tests passed during this review.

Existing reports in `plans/reviews/` were treated as historical claims and checked against current source where relevant. This file owns the new residual findings identified by R01-R18. It does not copy or close the earlier ledgers' implementation rows. The existing decision to keep HTTP admission, page compression, and signing budgets separate remains applicable: `plans/adr-2026-09-04-c3-single-budget-rejected.md`.

## Executive Summary

The application's top-level structure is broadly idiomatic Go. Keep the command, private implementation, public library, and operation-specific package boundaries. A wholesale package rewrite is not justified by this review.

The implementation inside those boundaries needs improvement. The most consequential problems are suppressed generation errors, incomplete cache identities, unsafe traversal of malformed documents, late resource limits, and configuration fields that do not control behavior. Several optimizations also interfere with correctness or add avoidable work.

The saved benchmarks establish that the project has invested in allocation reduction and compliant PDF throughput. They do not establish the performance of the September checkout. The best supported optimization candidates are removing JSON translation between typed Go values, eliminating repeated redaction extraction, and tightening memory ownership. No percentage speedup is claimed.

Priority meanings:

- P1: Fix before relying on the affected behavior for untrusted input, document correctness, or explicit security requirements.
- P2: Address in the next maintenance/performance work, with focused regression proof.
- P3: Smaller improvement after higher-priority work.

Findings describe source-confirmed control flow. Trigger examples and expected failures were not executed during this review.

## Go structure assessment

| Area | Assessment | Recommendation |
|------|------------|----------------|
| Commands, `internal/`, `pkg/gopdflib/` | Sensible separation between deployment, implementation, and supported library use. | Retain these boundaries. |
| Handler files and shared helpers | Operation-focused files and early returns make request paths readable. | Keep shared upload/error helpers, but enforce policy before parsing. |
| Public template types | Owned public types and schema parity tests provide a useful compatibility boundary. | Retain ownership; replace runtime JSON conversion with typed conversion, R10. |
| Generation lifetime | Borrowed bytes have explicit `Release` and `CopyBytes` contracts. Per-generation font clones isolate usage tracking. | Preserve these contracts in any optimization. |
| Rendering orchestration | `GenerateTemplatePDFBorrowed` spans roughly 1,000 lines and owns setup, layout, security, emission, and cleanup. `drawTable` also manages many parallel scratch slices. | Extend the existing `generation` object with cohesive private phases and one request-owned table scratch structure as affected code changes. Avoid moving the same complexity into forwarding wrappers. |
| HTTP dependency ownership | `PDFService` mixes rendering, upload policy, and HTTP status classification, then lives in a mutable package global. | Prefer instance-owned handler dependencies. Keep transport policy in handlers and rendering operations in the engine. No production race is asserted for the test setter. |
| Font API | `GetFontRegistry` returns `*pdf.CustomFontRegistry`, an internal concrete type that external users cannot name through supported imports. | Plan a publicly nameable registration handle with a limited method set. An alias fixes nameability but preserves the broad internal API. |
| Typst integration | Rendering callbacks supply font metrics and escaping without importing the whole PDF generator. | Preserve this dependency direction. No sufficiently supported Typst defect was established. |

Evidence: `pkg/gopdflib/generator.go:159,200`, `internal/handlers/services.go:33,165`, `internal/pdf/generator.go:373`, `internal/pdf/generation.go:25`, `internal/pdf/draw.go:527,1253`, and `typstsyntax/renderer.go`.

Go conventions should be distinguished from local preferences. Nil slices are valid Go, and a fixed argument count or line-length limit is not a language rule. Blank image-decoder imports have a functional purpose. Do not manufacture defects from those patterns or add a dependency merely because a general style skill mentions it. The review prioritizes clear ownership, meaningful errors, and small useful interfaces. See [Effective Go](https://go.dev/doc/effective_go).

## Findings

### R01 - P1: Generation can report success after required operations fail

Evidence: `internal/pdf/generator.go:411-415,481-491,550-551,571-573,1375-1380`.

An enabled signature with invalid PEM reaches `NewPDFSigner`, but generation proceeds when signer construction fails. A failure while writing the signature also returns PDF bytes with a nil error. Missing PDF/A fonts and failed font subsetting only print warnings. Callers cannot distinguish successful output from output missing requested properties.

Encryption has a separate concrete boundary gap: `Security.Enabled=true` with an empty owner password skips encryption altogether. The constructor-error fallback is also unsafe policy, but its only current constructor error is precluded by the caller's nonempty-password check. It is not counted as an independently reachable constructor failure.

Return wrapped errors for failed required operations, validate enabled security settings before rendering, and make any best-effort behavior explicit. Verification must exercise invalid signing material, enabled encryption without its required password, and unavailable required fonts through the public and HTTP entrypoints. Successful PDF/A output still needs veraPDF and structure-tree validation.

### R02 - P1: Row render cache keys mutable request memory

Evidence: `internal/pdf/draw.go:182-187,548-582,1369-1373`; `internal/handlers/decode.go:26-37`; `internal/models/models.go:148-171`; `internal/handlers/hft_decode.go:74-114`.

The global row cache keys rendered bytes by row address, page, marked-content index, and Y coordinate. It does not include text, column widths, font registration, or other rendering inputs. Pooled HTTP templates reuse row backing arrays. A later same-shaped shared-layout request can overwrite those rows while preserving the cache key and receive a previous request's rendered text. A direct internal caller that mutates a reused template has the same problem.

The key also keeps row backing storage reachable; the cache's byte accounting counts rendered values only. Public Go calls currently clone templates through JSON, so do not assume identical cache reuse on that path.

Prefer removing cross-request row rendering cache state or limiting reuse to an explicitly immutable prepared document. Existing `shared_row_cache_test.go` checks size eviction, not content correctness. Proof must render two different payloads through reused rows and compare extracted text and layout, including changed widths/fonts.

### R03 - P1: Font subset cache does not identify font contents

Evidence: `internal/pdf/font/subset_cache.go:36-58`; `internal/pdf/font/registry.go:159-193`; `internal/pdf/font/ttf.go:153-160`.

The subset key hashes the PostScript name and sorted glyph numbers. Different font versions can share both while containing different glyph outlines and metrics. Fonts without a usable name also fall back to `UnknownFont`. Once one font populates the cache, the other can receive its subset bytes and glyph mapping.

Give immutable parsed font data a content identity and include it in subset keys. Do not rehash an entire font for every rendered row. Existing subset-cache tests exercise one synthetic font, eviction, and clear. Add two different fonts with the same name and glyph IDs, then check embedded data and rendered glyphs.

### R04 - P1: Cyclic PDF trees have no traversal guard

Evidence: `internal/pdf/merge/merger.go:260-279`; `internal/pdf/redact/pdf_utils.go:55-65,133-146`; `internal/pdf/form/xfdf_locate.go:201-210`.

A `/Pages` object whose `/Kids` includes itself causes recursion without progress. Field trees have the same problem; named cycles can also allocate progressively longer field names. Small input can exhaust stack or memory. Merge checks object counts after `parseFile` returns, so those checks cannot protect this traversal.

Reject cycles and excessive depth with an error. Track traversal state explicitly, with a distinction between the current path and already validated nodes where needed. `merge/annotations.go` already contains a visited-object pattern. Add self-cycle, two-object-cycle, and excessive-depth cases, isolating dangerous regression runs in subprocesses. A shared validated page index would also support R06 and R16.

### R05 - P1: Unterminated object headers cause quadratic scanning

Evidence: `internal/pdf/pdfobj/pdfobj.go:156-177,195-258`.

`FindObjectBoundaries` scans to EOF looking for `endobj`, then retries from one byte after the candidate header. Repeated `1 0 obj` headers without a terminator cause almost the same suffix to be scanned repeatedly. For N evenly spaced headers, work grows as N + N-1 + ... + 1. Limits on successfully parsed objects cannot stop work that produces no completed objects.

Keep scanning progress monotonic or reject an unrecoverable unterminated object. Preserve legitimate stream handling. Add an adversarial regression that proves bounded work, alongside existing valid-object and embedded-token cases. This affects consumers of the shared scanner across several PDF operations.

### R06 - P1: Merge and split can change page order

Evidence: `internal/pdf/merge/merger.go:85-103,255-282,314-321`; `internal/pdf/merge/split.go:205-243`.

The parser obtains page-tree order, but output page lists are built while writing objects in numeric order. A document with `/Kids [6 0 R 3 0 R]` therefore becomes page 3 followed by page 6. Object numbers identify pages; they do not define reading order.

Build output page lists from `fc.Pages` for merge and the selected `pageObjs` for split. Keep byte serialization order independent. The inspected two-page fixture uses ascending object numbers and cannot expose this. Add nonascending and nested page-tree fixtures. Also verify inherited page resources and dimensions when removing ancestor Pages nodes; that inheritance concern remains a separate validation question, not a proven additional defect here.

### R07 - P1: Multipart limits run after multipart parsing

Evidence: `internal/handlers/request.go:202-240`; `internal/handlers/merge_split.go:17-43`; `internal/handlers/generate.go:19-20`.

Per-file bounded reads happen after `FormFile` or `MultipartForm` has parsed the request. Multipart routes lack the whole-body `MaxBytesReader` protection present on template JSON. Oversized uploads can consume incoming transfer and temporary storage before rejection. Merge additionally accumulates all accepted parts before calling the engine, without a total input-byte budget.

Apply endpoint-specific request limits before multipart parsing and retain per-file limits. Give merge an aggregate byte/count policy. Existing oversized-upload tests check the final 413 response, not early rejection. Add a counting request reader and a many-part request to prove the server stops reading at the intended boundary. The standard library distinguishes multipart memory thresholds from total body limits: [net/http request parsing](https://pkg.go.dev/net/http#Request.ParseMultipartForm).

### R08 - P1: Explicit route authentication policy is ignored

Evidence: `internal/handlers/router.go:22-32,61-66,160-167`; `internal/handlers/handlers.go:109-113`; `internal/middleware/auth.go:68`.

`RoutePolicy.RequireAuth` promises to enforce authentication, but route wiring constructs middleware that reads environment variables instead. With deployment auth variables unset, an explicitly constructed policy requiring auth does not enforce it. Environment-based deployment defaults still work; the defect is the ineffective explicit configuration contract.

Related fields have the same ownership problem: `MaxHTMLBodyBytes` is ignored by handlers, `Policy.MaxConcurrent` duplicates the top-level field, and profiling rereads the environment after resolution.

Resolve environment defaults at startup, then honor the supplied configuration. Existing auth tests set environment and configuration together. Add tests with deliberately different environment and explicit policy values.

### R09 - P2: Decode preallocation changes which content is rendered

Evidence: `internal/models/models.go:31-37,65-75`; `internal/handlers/generate.go:14-16`; `internal/handlers/json_decode.go:69,83`; `internal/pdf/generator.go:2051-2066,2112`.

The active tier preallocates an inline table by extending `Elements` to length two. If valid JSON uses the top-level `table` field and omits `elements`, decoding leaves these placeholder elements in place. Rendering sees nonempty `Elements`, skips its legacy-table branch, and emits neither real table. HFT streaming with unknown Content-Length has the same problem; the known-length HFT shell explicitly clears missing elements.

Capacity planning must not create logical document elements or defaults. Keep reusable storage separate from decoded length. Add semantic parity cases for the same legacy template under no tier, active, and HFT, with known and unknown body lengths. Existing allocation-shape tests do not prove this equivalence.

### R10 - P2: Typed Go templates are serialized to translate types

Evidence: `pkg/gopdflib/adapter.go:20-47`; `pkg/gopdflib/generator.go:109-132,146-178`.

`GeneratePDF` marshals a complete public template and unmarshals it into internal types. The JSON-to-generation path first decodes an internal template, serializes it into public types, then repeats conversion back to internal types. That path performs three unmarshals and two marshals before rendering. Rows and embedded assets pass through these allocations repeatedly.

Keep owned public types, but use explicit or generated typed conversion with an intentional copy/ownership policy. Let JSON-to-generation share validated internal decoding without the intermediate public round trip. Preserve schema parity and non-JSON font hints. This is a confirmed extra-work path, not a measured September regression. Do not trade it for unsafe aliases to mutable caller memory or revive R02 while removing copies.

### R11 - P2: HTTP cancellation does not reach HTML work

Evidence: `internal/handlers/services.go:41-42`; `internal/handlers/html.go:49,87`; `internal/pdf/html_convert.go:60,107`; `internal/handlers/request.go:162`.

HTML conversion calls the underlying renderer with `context.Background()`. Its preliminary DNS lookup also lacks the caller's context. Cancellation or disconnect does not stop this work through these call paths, so it retains admission capacity until completion. A socket write timeout does not supply the missing operation context.

Pass `c.Request.Context()` through the HTML service and rendering calls, with a deliberate operation deadline where required. Use context-aware DNS resolution. Preserve existing public entrypoints while defining any new context-taking API consistently. Verify cancellation propagation and admission-slot release with controlled local dependencies. The [context package contract](https://pkg.go.dev/context) describes propagating context through a call chain.

### R12 - P2: WASM and CGO can copy input that the engine will reject

Evidence: `cmd/wasm/main.go:63-80,158-170,291-295`; `bindings/python/cgo/exports.go:45-47,236-252`; `internal/pdf/merge/merger.go:19-20,41-42`.

The shared WASM copier allocates from JavaScript `byteLength` before downstream limits. Merge copies every input before validation. A large rejected upload therefore still increases module memory. The compression binding already checks size before copying and provides a local pattern to follow.

The CGO merge bridge permits a 512 MiB part and copies it, although the engine's per-file cap is 32 MiB. Its comment claiming matching limits is stale.

Share operation-specific byte/count limits and check them before crossing the memory boundary. Define aggregate merge limits too. Go wrapper tests do not exercise `syscall/js` copying; future proof needs the actual WASM adapter and CGO boundary, including rejection without an input-sized copy.

### R13 - P2: Pooled template reset retains image payload references

Evidence: `internal/models/models.go:98-120`; `internal/handlers/decode.go:34-37`.

`ResetForReuse` shortens `t.Image` without clearing its elements. Base64 image strings remain reachable through its backing array after the request finishes. Smaller requests can leave old slots referenced until overwritten or the pool entry is discarded. This is retained memory, not proof of a permanent leak or measured RSS growth.

Clear live image elements before shortening the slice, retaining useful capacity. Extend reset tests to inspect former slots after reslicing. Review other pointer-bearing pooled fields with the same ownership question rather than clearing buffers that intentionally contain immutable reusable data.

### R14 - P2: One template cache entry can exceed its byte budget

Evidence: `internal/handlers/template_data.go:18-20,61-78`.

When an incoming entry exceeds the 16 MiB cache budget, the cache clears existing entries and then inserts the oversized entry anyway. A sufficiently large local JSON template therefore exceeds the declared budget and evicts useful data.

Skip caching entries larger than the entire budget. Preserve the intended serving behavior separately. Existing tests cover entry-count eviction; add individual and aggregate byte-budget cases. This is a residual defect in the implemented bounded cache, not a request to duplicate the older bounded-cache design work.

### R15 - P1: Template image decoding has no decoded-pixel budget

Evidence: `internal/pdf/image.go:197-225,251-252`; `internal/handlers/generate.go:19-20`.

The template renderer calls `image.Decode` before checking image dimensions, then allocates width × height × 3 bytes for PNG RGB conversion. A small compressed image can describe a much larger raster. The JSON byte limit and the RGB pool's retention cap do not bound this active allocation. The image cache has an entry-count cap but no total byte cap.

Read image configuration first, check dimensions and pixel multiplication safely, and impose a decoded-image budget before raster allocation. Distinguish per-image and per-document limits. Add a small encoded image with excessive dimensions and normal-image regression fixtures. JPEG handling also fully decodes pixels even though it embeds the original JPEG bytes; consider header-only inspection after validating the required format properties. Timing benefits remain unmeasured.

### R16 - P2: Multi-term redaction repeats page discovery and extraction

Evidence: `internal/pdf/redact/search.go:27-40,64-74,106-120`; `internal/pdf/redact/pdf_utils.go:112-163`.

Each unique search term reruns single-term search. Each page extraction traverses from the page-tree root, decodes content, and parses text positions again. With Q terms and P pages, page lookup alone can approach O(Q × P²) visits on a flat page tree. Decoding and text parsing repeat for each term. Caching the object map does not remove these costs.

Build a validated ordered page index once per Redactor, then extract each page once within a multi-term operation and match all terms against it. Preserve coordinate conversion, phrase matching, and deduplication. Existing generation benchmarks say nothing about this path's latency. The algorithmic repetition is established; the throughput impact is not measured.

### R17 - P3: Stream encryption copies input twice before encryption

Evidence: `internal/pdf/encryption/encrypt.go:280,322-327`.

The caller clones plaintext before passing it to `Pkcs7Pad`, which already allocates and copies into a new slice. For nonempty streams, the first clone is one redundant input-sized allocation and copy.

Pass the original read-only input to padding and preserve its no-mutation contract. Existing `TestPkcs7PadNoMutate` and `TestEncryptStream_RoundTrip` are relevant future checks. Further combining ciphertext and IV allocation is a separate optimization to justify, not required for this cleanup.

### R18 - P2: Boundary normalization and error contracts diverge

Evidence: `pkg/gopdflib/adapter.go:110-124`; `pkg/gopdflib/errors.go:64`; `internal/pdf/html_convert.go:68-79`; `internal/handlers/html.go:77-104`.

Invalid compression-level strings return plain errors, so public `CodeOf` classifies them as internal failures rather than invalid input. The WASM caller currently compensates by wrapping the error, which avoids claiming an incorrect present WASM response.

HTML image conversion normalizes format case internally, but the handler chooses response metadata from the original string. `JPG` can therefore produce JPEG bytes labeled `image/png`; `SVG` misses the handler's lowercase check and reaches its 500 path.

Use the shared invalid-input sentinel at normalization boundaries. Normalize image format once and use that value for both conversion and response metadata. Extend tests to assert error identity/code and MIME type, not merely a nonnil error or nonempty output.

## Existing performance evidence

### Saved measurements and their limits

| Evidence | Saved result | What it supports |
|----------|--------------|------------------|
| [Zerodha compliant x10 summary](../../guides/cursor/baselines/zerodha_bench_x10_wsl_stats_latest.txt), dated 2026-06-24 by [the benchmark report](../../documentation/BENCHMARKS.md) | Mean 6,202.72 ops/s; median 6,361.56; peak 6,611.44; mean average latency 7.544 ms. | Historical repeated-template throughput, 48 workers, 5,000 iterations, 80/15/5 mix, compliant/tagged generation with ECDSA retail signing. |
| [Saved compliant run 10](../../guides/cursor/baselines/zerodha_bench_x10_wsl/zerodha_run10.txt) | Go 1.26.4, 24 CPUs, GOMAXPROCS 24; 4,000 retail, 750 active, 250 HFT; warm-up explicitly ran. | Confirms the saved workload/environment and warm state. |
| [Zerodha noncompliant x10 summary](../../guides/cursor/baselines/zerodha_bench_x10_nocomply_wsl_stats_latest.txt) | Mean 34,035.02 ops/s; mean average latency 1.376 ms. | A different feature configuration. It is not a drop-in performance target for compliant output. |
| [Financial-report handler results](../../documentation/BENCHMARKS.md) | Best-of-five serial 55,528 ns/op, 344,521 B/op, 294 allocs/op; parallel 54,814 ns/op. | Historical in-process handler/renderer measurements for that fixture. |
| [June 20 optimization/profile record](../../guides/optimizations/20260620_zerodha_x10_pprof_optimization_checklist.md) | Recorded cumulative costs include content generation, table rendering, and structure emission. | Explains existing arenas, capacity tiers, batching, and compression reuse. It is not a current hotspot ranking. |

The native benchmark command is `make bench-gopdflib-zerodha-x10`, with dataset construction in `sampledata/gopdflib/zerodha/bench.go` and runner `run_bench_x10.sh`. Its default warm-up generates all three templates before timed iterations. Timed iterations use borrowed output; they exclude caller copying, network transmission, and writing every generated PDF to disk. The x10 script starts separate processes; cold builds and warmed renderer caches are different concepts.

The saved compliant summary's mean peak allocation is 797.91 MB as labeled by the runner. The current `monitorMemory` samples `runtime.MemStats.Alloc` every 100 ms and divides by 1024², so this is sampled live Go heap in MiB, not process RSS, B/op, or total allocation traffic. Sampling can miss short peaks. See `sampledata/gopdflib/zerodha/bench.go:72-91`.

The handler benchmark constructs `gin.New()` and calls `RegisterRoutes`, bypassing `NewRouter` middleware such as admission control. It uses `httptest.ResponseRecorder`. Its numbers do not prove whole-server backpressure, socket performance, or overload behavior. See `test/benchmark_handlers_test.go:14-19`.

Older optimization documents contain intermediate higher throughput figures and refer to mutable `latest` artifact paths. Keep each document's date and conditions attached. No source revision was verified with Git in this review, and no saved result is presented as a September measurement.

### Performance priorities

Correct R02 and R03 before judging cache hit-rate improvements. A fast cache returning the wrong document content is unusable.

R10 is the strongest application-wide avoidable-work candidate on the public Go and JSON entrypoints. R16 is the strongest algorithmic opportunity outside generation. R12-R15 address memory consumed or retained at boundaries. R17 is a small, mechanically supported copy reduction.

Additional hypotheses deserve targeted evidence before tuning: the generator starts a compression goroutine even for a single stream; worker sizing uses CPU count; PDF/A availability checks perform filesystem work while holding a global manager lock. These paths are visible in `internal/pdf/generator.go:286-291,900-919` and `internal/pdf/font/pdfa.go:201-218,392-419`. They may matter for small documents or constrained runtimes, but the saved results cannot quantify their present impact. Keep the separate concurrency budgets already chosen by the project.

Preserve bounded PDF/page/compression buffers, the lazy structure arena, borrowed output ownership, and full table tagging. Existing source contains useful protections, including image-cache locking/copies, capped compressor inflation and dimensions, deferred signature-slot release, and SVG/XFDF input guards. Those protections do not close the residual findings above.

## Phase 1: Correctness and bounded input

### 1.1 Required generation behavior and cache identity

- [ ] R01: `internal/pdf/generator.go` returns errors when required signing/font work fails and rejects incomplete enabled encryption settings. Proof: public/HTTP failure regressions plus gates G1-G3.
- [ ] R02: `internal/pdf/draw.go` no longer reuses rendered output through mutable cross-request row identities. Proof: changed-text/geometry/font cases on reused pooled rows, race checks, and G1-G3.
- [ ] R03: `internal/pdf/font/subset_cache.go` includes immutable font-content identity. Proof: same-name/different-font fixtures, concurrent generation, and G1-G3.

### 1.2 Parser progress and page semantics

- [ ] R04: Merge/redact page walks and XFDF field walks reject cycles and excessive depth. Proof: bounded subprocess regressions and G1-G2.
- [ ] R05: `internal/pdf/pdfobj` stops rescanning malformed suffixes without progress. Proof: repeated unterminated headers plus valid stream fixtures and G1-G2.
- [ ] R06: Merge and split preserve page-tree order independently of object serialization order. Proof: nonascending/nested page fixtures, inherited-property checks, and G1-G3 for applicable fixtures.

### 1.3 Resource and configuration boundaries

- [ ] R07: HTTP upload routes enforce total body limits before multipart parsing and an aggregate merge budget. Proof: counted reads, oversized/many-part requests, 413 behavior, and G1-G2.
- [ ] R08: Explicit router policy governs auth and body limits; duplicated configuration ownership is removed. Proof: environment-independent router-policy tests and G1-G2.
- [ ] R15: `internal/pdf/image.go` checks decoded dimensions/pixel budgets before raster allocation. Proof: excessive-dimension rejection, ordinary PNG/JPEG/SVG fixtures, and G1-G3.

## Phase 2: API contracts and lifecycle

### 2.1 Decoding and cancellation

- [ ] R09: Tier preallocation preserves logical document contents for omitted fields. Proof: no-tier/active/HFT render parity with known/unknown Content-Length and G1-G3.
- [ ] R11: HTML rendering and DNS receive the request context. Proof: cancellation stops downstream work and releases admission capacity, plus G1-G2.

### 2.2 Adapter contracts

- [ ] R12: Go WASM and CGO reject oversized operation inputs before copying, with shared per-operation and aggregate limits. Proof: actual adapter rejection/ownership checks and G1/G4.
- [ ] R18a: Compression normalization returns errors identifiable as invalid input. Proof: `errors.Is` and `CodeOf` cases plus G1.
- [ ] R18b: HTML image format uses one canonical value for conversion and response metadata. Proof: mixed-case/whitespace JPG and SVG cases plus G1-G2.

## Phase 3: Allocation and repeated work

### 3.1 Remove avoidable conversion and retained references

- [ ] R10: Replace full-template JSON type translation with typed, ownership-safe conversion and a direct validated JSON render path. Proof: nested-field/schema parity, font hints, caller immutability, and G1-G3. Performance remains unproven until G5 is authorized.
- [ ] R13: Clear top-level image references before pooled template reuse. Proof: reset backing-array inspection and G1-G2.
- [ ] R14: Skip oversized individual template-cache entries. Proof: individual/aggregate byte budgets and G1-G2.
- [ ] R17: Remove the clone immediately before copying PKCS#7 padding. Proof: existing no-mutation/round-trip tests and G1-G3 for encrypted fixtures.

### 3.2 Reuse per-document redaction work

- [ ] R16: A Redactor reuses validated page order and extracts each page once per multi-term search. Proof: extraction-count tests, existing phrase/coordinate cases, and G1-G2. Performance remains unproven until G5 is authorized.

The structure recommendations above are design guidance for these phases, not a second implementation ledger. In particular, adjust handler ownership while implementing R08/R11 and rendering phases while implementing R01/R02. Public font-handle redesign can be planned separately if callers need it; it is not required to fix cache identity.

## Phase 4: Closure gates

All gates below describe future implementation validation. None ran during this documentation-only review.

- [ ] G1: For any non-documentation implementation, run `make fmt && make lint && make test`; record command output and exit status. `make test` includes Go, Python bindings, and PDF verification.
- [ ] G2: For `internal/handlers/` or `internal/pdf/` implementation changes, run `make test-integration` and the finding-specific cases. Record final-source results.
- [ ] G3: When PDF bytes change, run `make test-verify-pdfs` and record veraPDF plus structure-tree results for applicable compliant fixtures. Check output content/order as well as validity; use encryption/signature-specific checks where compliance is not the relevant contract.
- [ ] G4: For Go adapter changes, build the actual WASM targets and Python CGO library and exercise their input/error/ownership paths. Record the exact build and runtime commands used; Go wrapper tests alone do not close this gate.
- [ ] G5: Deferred by the user's no-benchmark-rerun instruction. If later authorized, compare matched workloads and cache states using the existing harnesses. Use `make bench-gopdflib-zerodha-x10` for native generation, `make bench-handler-all` for handler generation, and `make bench-k6-light` for HTTP load; add an operation-specific redaction case before claiming R16 speedups. Record source version, fixture, feature flags, workers, cache state, mean/median/tails, allocation metric, and applicable PDF proof. Do not close performance claims using unmatched historical results.

## Dependencies

R01-R09 and R15 take precedence over performance tuning where they affect the same paths. Resolve R02 before changing conversion ownership in R10. R04's validated page traversal should support R06 and R16 without creating another independent tree walker. Set aggregate input policy in R07 before sharing it with adapters in R12. Preserve the separate concurrency-budget ADR.

Changes to public types, normalization, or generated conversion must preserve `internal/models/schema_parity_test.go`, schema golden tests, and `pkg/gopdflib` boundary contracts. Performance work must preserve PDF/A-4 and PDF/UA-2 behavior where requested, including table-cell structure and font embedding. Disabling features is a different workload.

No improvement row is marked complete. The delivered artifact is this review and implementation checklist. Source fixes and validation remain future work.

## Review method and handoff

Applied the named phase-wise checklist skill to keep findings, dependencies, and evidence gates in one report. The Go style skill guided the parallel source reviews and the distinction between clarity problems and formatting preferences. Feynman checks tightened the explanations of cache identity, font subsets, page order, and allocation; the coordinating explanation audit was clean after two passes. Unslop guidance removed generic claims and kept recommendations tied to current source.

Report-only validation passed on 2026-09-05: 18 finding IDs, 24 open checklist rows, six local links, and 68 source references checked. Required headings are present; no closed implementation rows or em dashes were found. These checks validate references and report structure, not application behavior.

The next implementation work is Phase 1. No code fixes, benchmark reruns, commits, or publication are included in this review.
