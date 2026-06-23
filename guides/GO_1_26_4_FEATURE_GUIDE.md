# Go 1.26.4 Feature Guide for gopdfsuit

**Date:** 2026-06-17  
**Toolchain:** Go 1.26.4 (`go.mod` in the main module, `sampledata/`, `certs/`, and most submodules)  
**Audience:** Contributors optimizing PDF generation throughput, memory, and operational safety

This guide maps Go 1.26.x language, runtime, compiler, and standard-library changes to concrete areas of this repository. Go 1.26.4 is a patch release on the 1.26 line; the features below come from Go 1.26.0 and remain available in 1.26.4.

---

## Executive summary

| Category | Impact for gopdfsuit | Action |
|----------|----------------------|--------|
| Green Tea GC (default) | High - PDF work is allocation-heavy | Already active when built with Go 1.26.4; validate with existing k6 + pprof harness |
| Stack-backed slice allocation | Medium–High - many hot `make`/`append` paths | Already active; watch for rare miscompiles via bisect |
| Faster `io.ReadAll` | Medium - upload/redact handlers | Already active on stdlib call sites |
| Faster `image/jpeg` | Medium - base64 JPEG rows in templates | Already active; re-run image-heavy integration tests |
| Faster cgo | Low–Medium - indirect via chromedp stack | Automatic if cgo is linked |
| `go fix` modernizers | Medium - codebase hygiene | Run once, review diff |
| `B.Loop` benchmarks | Medium - more trustworthy bench numbers | Migrate remaining `b.N` loops |
| `runtime/secret` (experimental) | Medium - signing key handling | Evaluate for `internal/pdf/signature` |
| `goroutineleak` profile (experimental) | Medium - worker pools under load | Enable in CI/staging load tests |
| `new(expr)` | Low - readability in models/tests | Optional cleanup |
| `simd/archsimd` (experimental) | Unknown - SIMD for scan/compress paths | Research only; not production-ready |
| `crypto/hpke`, ML-KEM TLS | Low today | Relevant if adding HPKE or hardening outbound TLS |

**Bottom line:** The largest wins require no code changes - you are already on Go 1.26.4 and this workload (pooled buffers, caches, short-lived slices, JSON/PDF bytes) is exactly what Green Tea and the improved `io.ReadAll` target. Focus next on validation (`make bench-k6`), benchmark modernization, and selective experiments (`goroutineleak`, `runtime/secret`).

---

## 1. Already benefiting (no code changes required)

Build and run with Go 1.26.4 and these improvements apply automatically.

### 1.1 Green Tea garbage collector (default since Go 1.26)

**What it does:** Re-architects parallel GC marking/sc scanning for better cache locality on small objects. Typical programs see **10–40% lower GC CPU**; on newer amd64 (Intel Ice Lake+, AMD Zen 4+) an additional ~10% from vectorized small-object scanning.

**Why it matters here:** gopdfsuit is GC-heavy by design:

- `sync.Pool` for PDF buffers, scratch buffers, JSON decode bodies, SHA-256 hashers, zlib writers, RGB conversion, and pooled `PDFTemplate` instances (`internal/pdf/generator.go`, `internal/handlers/json_decode.go`, `internal/pdf/signature/signature.go`, `internal/pdf/image.go`, `internal/pdf/font/compression.go`).
- Per-request short-lived slices during PDF assembly, structure trees, and table rendering.
- Cross-request caches (`subsetCache`, `imgCache`, `propsCache`, `pdfSignerCache`) that retain heap objects between requests.

These patterns create many small, short-lived allocations - the profile Green Tea improves most.

**How to validate:**

```bash
make bench-k6
# Compare pprof GC time vs pre-1.26 baselines in guides/cursor/baselines/gin_pprof_runs/
```

Look for reduced time in `runtime.gcBgMarkWorker`, `mallocgc`, and `scanobject` in CPU profiles. Heap profiles may show similar in-use bytes but lower allocation churn (`alloc_space`).

**Opt-out (debug only):** `GOEXPERIMENT=nogreenteagc` at build time. Only use to A/B test; the setting is expected to be removed in Go 1.27.

**Note:** `frontend/go.mod` still declares `go 1.24.0`. Align it when that module is built with 1.26.4 so tooling and CI stay consistent.

---

### 1.2 Stack-backed slice backing stores (compiler)

**What it does:** The compiler can place slice backing arrays on the stack in more cases, avoiding heap allocations for slices that do not escape.

**Relevant code paths:**

- `internal/pdf/generator.go` - extensive `append`/`make` for AcroForm content, structure output, scratch buffers.
- `internal/pdf/draw.go`, `internal/pdf/font/metrics.go` - formatting and layout temporaries.
- `typstsyntax/renderer.go` - grid layout slices and point arrays for math rendering.

**Expected effect:** Small per-operation savings that compound under 1000+ req/s load. Hard to see in macro benchmarks but should nudge `alloc_space` down in pprof.

**If something breaks:** Use bisect with `-compile=variablemake` or disable via `-gcflags=all=-d=variablemakehash=n` and file a Go issue.

---

### 1.3 Faster `io.ReadAll` (stdlib)

**What it does:** ~2× faster on large inputs, ~half the intermediate allocations, minimally sized result slice.

**Call sites in this repo:**

| Area | Files |
|------|-------|
| HTTP handlers | `internal/handlers/handlers.go`, `internal/handlers/redact.go` |
| Integration tests | `test/integration_*.go` |

Redact endpoints read full uploaded PDFs via `io.ReadAll`. Under load, this reduces allocator pressure before PDF parsing even starts.

**Note:** JSON decode for `/generate/template-pdf` uses pooled `io.ReadFull` in `internal/handlers/json_decode.go` - that path is already optimized and bypasses `ReadAll` for known `Content-Length`. The stdlib win applies to redact/multipart and test helpers.

---

### 1.4 Faster `image/jpeg` encoder/decoder (stdlib)

**What it does:** New JPEG implementation - faster and more accurate. Bit-exact output with the old encoder is not guaranteed.

**Relevant code:** `internal/pdf/image.go` registers `image/jpeg` and decodes base64 JPEG rows from templates.

**Action:** Re-run image-related integration tests and any golden PDF comparisons that embed JPEG assets. Expect decode CPU savings on templates with many image rows (retail Zerodha-style payloads).

---

### 1.5 ~30% lower cgo call overhead (runtime)

**What it does:** Reduces baseline cost of Go ↔ C transitions.

**Relevance:** gopdfsuit does not use `import "C"` directly, but dependencies such as **chromedp** (screenshots / headless flows in `screenshots/`) may link cgo. Any chromedp-heavy path benefits automatically.

---

### 1.6 Crypto speed and safety defaults (stdlib)

Relevant to signing and encryption in this repo:

| Package / change | gopdfsuit usage | Benefit |
|------------------|-----------------|---------|
| `crypto/ecdsa`, `crypto/rsa` signing | `internal/pdf/signature/signature.go` - `rsa.SignPKCS1v15`, `ecdsa.SignASN1` | Cleaner RNG handling; `rand` param ignored (always secure source) |
| `crypto/mlkem` ~18% faster | Not used directly | Future PQ work |
| `crypto/tls` ML-KEM hybrids default | Only if you terminate TLS in-process | Stronger default handshakes |
| `testing/cryptotest.SetGlobalRandom` | `internal/pdf/signature/signature_test.go` | Deterministic signature tests without passing custom `rand` |

Signing is on the ECDSA Zerodha benchmark path (~80% of weighted k6 traffic). Faster, safer crypto stacks marginally reduce tail latency on signed PDFs.

---

## 2. Recommended adoptions (low risk, clear payoff)

### 2.1 Run `go fix` modernizers

Go 1.26 rewrote `go fix` as a modernizer suite (same analysis framework as `go vet`). It applies dozens of safe idiomatic updates.

```bash
# Preview
go fix -fix=all ./... 

# Or run specific analyzers; review diff before commit
go fix ./internal/... ./pkg/... ./cmd/...
```

**Suggested workflow:**

1. Run on a clean branch.
2. Review the diff - modernizers should not change behavior.
3. Run `make test` (or your usual `go test ./...`) and a quick `make bench-k6` spot check.

Use `//go:fix inline` on any deprecated internal helpers you want callers migrated away from automatically.

---

### 2.2 Migrate benchmarks from `b.N` to `b.Loop`

Go 1.26 fixed `B.Loop` so loop bodies can inline correctly (older `B.Loop` prevented inlining and skewed results).

**Files still using `b.N`:**

- `internal/pdf/benchmark_test.go`
- `internal/pdf/benchmark_macro_test.go`
- `internal/pdf/benchmark_compare_test.go`
- `sampledata/benchmarks/gopdfkit_compare/compare_benchmark_test.go`
- `test/benchmark_handlers_test.go` (partial - also uses `pb.Next()`)

**Before:**

```go
for i := 0; i < b.N; i++ {
    generatePDF(template)
}
```

**After:**

```go
for b.Loop() {
    generatePDF(template)
}
```

**Why:** More accurate microbenchmarks when evaluating optimizations documented in `guides/optimizations/`. Misleading bench numbers have already led to reverted changes (e.g. generic structure writer, aggressive pool caps in the optimization checklist).

---

### 2.3 Use pprof flame graphs by default

Go 1.26’s `go tool pprof -http` now opens the **flame graph** view first.

Your harness already captures profiles:

```bash
test/generate_template-pdf/run_gin_pprof_load.sh
```

When analyzing `guides/cursor/baselines/gin_pprof_runs/pprof_summary_*.txt`, prefer the web UI for hot-path visualization:

```bash
go tool pprof -http=:8081 ./bin/gopdfsuit guides/cursor/baselines/gin_pprof_runs/cpu_*.prof
```

Graph view remains under **View → Graph** if you need the old layout.

---

### 2.4 `errors.AsType` for typed error handling

Go 1.26 adds `errors.AsType[T](err) (T, bool)` - type-safe and faster than `errors.As`.

**Today:** The codebase rarely uses `errors.As`. As error wrapping grows (merge, redact, signature, encryption), prefer:

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) { ... }

// Go 1.26+
if pathErr, ok := errors.AsType[*os.PathError](err); ok { ... }
```

Most useful in handler error mapping and library boundaries (`pkg/gopdflib`).

---

### 2.5 `fmt.Errorf` allocation reduction

Static format strings like `fmt.Errorf("missing PDF header")` now allocate similarly to `errors.New`. Many validation errors in benchmarks and handlers use static strings - no migration required; they already benefit.

---

## 3. Language features (readability, limited performance impact)

### 3.1 `new(expression)` - initialize pointer fields in one step

Go 1.26 allows `new(expr)` instead of helper variables.

**Relevant models** (`internal/models/models.go`) use optional pointer fields:

- `EmbedFonts *bool`
- `Checkbox *bool`, `Wrap *bool`, `MathEnabled *bool`
- `Width *float64`, `Height *float64`

**Example - building a test fixture or default config:**

```go
// Before
v := true
cfg := models.Config{EmbedFonts: &v}

// Go 1.26
cfg := models.Config{EmbedFonts: new(true)}
```

```go
// Before
w, h := 200.0, 50.0
sig := models.SignatureConfig{Width: &w, Height: &h}

// Go 1.26
sig := models.SignatureConfig{
    Width:  new(200.0),
    Height: new(50.0),
}
```

**Performance:** Negligible. **Value:** Less boilerplate in tests and sampledata (`sampledata/gopdflib/zerodha/main.go`, benchmarks).

---

### 3.2 Self-referential generic constraints

Generic types can now refer to themselves in constraints, e.g. `type Adder[A Adder[A]] interface { Add(A) A }`.

**Relevance to gopdfsuit:** Limited today - the codebase is not generic-heavy. A prior generic structure-writer experiment in `internal/pdf/generator.go` **regressed** throughput (shape-based dispatch, not monomorphization) and was reverted (`guides/optimizations/20260614_remaining_optimizations_checklist.md`).

**Guidance:** Do **not** reintroduce generics for PDF hot paths based on this feature alone. Self-referential constraints are useful if you later model tree-shaped layout IR or pluggable element types - not for micro-optimizations.

---

## 4. Experimental features worth evaluating

### 4.1 Goroutine leak profile

**Enable at build:**

```bash
GOEXPERIMENT=goroutineleakprofile go build -o bin/gopdfsuit ./cmd/gopdfsuit
```

**Endpoint:** `/debug/pprof/goroutineleak` (when pprof routes are enabled and `ENABLE_PROFILING=1` / localhost middleware allows access).

**Why it matters here:**

- Concurrency limiting via semaphores: `cmd/gopdfsuit/main.go`, `internal/pdf/signature/signature.go` (`signWorkerSlots`), `internal/pdf/generator.go` (`pageCompressSlots`).
- Worker pools in benchmarks and `golang.org/x/sync/errgroup` in `internal/pdf/generator.go`.
- Early-return paths in parallel work (classic leak pattern from Go release notes).

**Use case:** Run under k6 load, then capture `goroutineleak` profile after sustained traffic. Helps catch stuck workers that hold memory and slots, depressing throughput toward the ~1500 req/s target.

Expected to become default in Go 1.27 - validating now reduces surprise later.

---

### 4.2 `runtime/secret` - secure erasure of key material

**Enable at build:**

```bash
GOEXPERIMENT=runtimesecret go build ./...
```

**Relevant code:** `internal/pdf/signature/signature.go` parses PEM private keys into `crypto.PrivateKey` and signs PDFs; `internal/models/models.go` carries `PrivateKeyPEM` in JSON config.

**Potential use:** Wrap signing operations so key bytes and temporaries are cleared from stack/registers/heap after use. Important for forward secrecy and compliance narratives - not a throughput optimization.

**Status:** Experimental, amd64/arm64 Linux only. Evaluate in staging before production signing paths.

---

### 4.3 `simd/archsimd` - architecture-specific SIMD

**Enable:** `GOEXPERIMENT=simd`  
**Status:** Experimental, amd64 only, API unstable.

**Hypothetical fits (research only):**

- RGB conversion / image scan loops in `internal/pdf/image.go`
- Zlib compression in `internal/pdf/font/compression.go`
- FNV-1a hashing in `internal/pdf/image.go`

**Guidance:** Do not ship in production PDF paths until a portable API exists. Benchmark in isolation first; PDF generation is rarely SIMD-bound compared to JSON decode, font subsetting, and structure-tree I/O.

---

## 5. Standard-library additions with future relevance

| Feature | When to adopt |
|---------|----------------|
| `crypto/hpke` | If adding hybrid encryption for document or key exchange APIs |
| `bytes.Buffer.Peek` | PDF stream parsers that currently read-then-unread bytes |
| `log/slog.NewMultiHandler` | If structured logging is added alongside Gin |
| `net/http/httputil.ReverseProxy.Rewrite` | If adding a reverse proxy in front of the API |
| `reflect.Type.Fields` iterators | Reflection-heavy tooling, not hot PDF paths |
| `testing.{T,B,F}.ArtifactDir` | Store generated PDF artifacts from failing tests with `go test -artifacts` |
| `runtime/metrics` scheduler counters | Custom dashboards for goroutine/scheduling pressure under k6 |

---

## 6. Features with little or no benefit here

| Feature | Reason |
|---------|--------|
| `go mod init` default version change | Existing modules already pin `go 1.26.4` |
| Removal of `cmd/doc` | Use `go doc` |
| WebAssembly heap changes | Not a WASM deployment |
| Darwin / PowerPC port notes | Deployment targets differ |
| `net/url` strict colon parsing | Unless user-supplied URLs are parsed in new code |

---

## 7. Suggested validation plan

Align with existing benchmark culture in `guides/INTEGRATION_AND_BENCHMARK_TESTS.md` and `guides/optimizations/`.

### Phase A - Confirm runtime wins (no code diff)

1. Build with Go 1.26.4: `go build -o bin/gopdfsuit ./cmd/gopdfsuit`
2. Run `make bench-k6` twice; compare to baselines in `guides/cursor/baselines/gin_pprof_runs/`
3. Check CPU profile for GC/mark shrinkage; check `alloc_space` on heap profile

### Phase B - Tooling and benchmarks

1. `go fix ./...` - review and commit idiomatic updates
2. Convert `b.N` → `b.Loop` in `internal/pdf/benchmark_*.go`
3. Re-run `go test -bench=. ./internal/pdf/...` and note delta (expect small, directionally trustworthy)

### Phase C - Experiments (optional branch)

1. `GOEXPERIMENT=goroutineleakprofile` - load test + leak profile
2. `GOEXPERIMENT=runtimesecret` - signing integration tests + security review
3. A/B `GOEXPERIMENT=nogreenteagc` only to quantify Green Tea benefit; do not ship opt-out

---

## 8. Priority matrix (for gopdfsuit)

```
Impact ↑
  │
  │  Green Tea GC ●          Stack slice alloc ●
  │  io.ReadAll ●            image/jpeg ●
  │  go fix / B.Loop ●       goroutineleak exp ●
  │  cgo (chromedp) ●        runtime/secret exp ●
  │  new(expr) ○             errors.AsType ○
  │  self-ref generics ○     simd/archsimd △
  └──────────────────────────────────────────→ Effort
        (already done)   (low)    (medium)   (research)
```

● = recommended focus · ○ = optional polish · △ = experimental / high uncertainty

---

## References

- [Go 1.26 release notes](https://go.dev/doc/go1.26)
- [Go 1.26 release blog post](https://go.dev/blog/go1.26)
- Internal baselines: `guides/cursor/baselines/gin_pprof_runs/`
- Optimization history: `guides/optimizations/20260614_remaining_optimizations_checklist.md`
- Benchmark harness: `test/generate_template-pdf/run_gin_pprof_load.sh`
