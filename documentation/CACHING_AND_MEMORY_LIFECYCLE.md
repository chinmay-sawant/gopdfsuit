# Caching & Memory Lifecycle

**Date:** 2026-06-11 · **Go:** 1.26.4

This guide explains what gopdfsuit/gopdflib **actually caches**, what **expires**, and how that behaves when you process millions of unique records.

---

## TTL policy: 3 minutes by default, configurable

All process-global content caches share one time-to-live policy owned by
`internal/cachettl`:

| Property | Value |
|----------|-------|
| **Default TTL** | 3 minutes (`cachettl.DefaultTTL`) |
| **Env override** | `GOPDFSUIT_CACHE_TTL` (Go duration, e.g. `"2m"`, `"90s"`), read once at startup |
| **Code override** | `gopdflib.SetCacheTTL(d)` / `gopdflib.CacheTTL()` |
| **Disable expiry** | Set TTL `<= 0` to restore size-only eviction |
| **Mechanism** | Lazy expiry on lookup: stale entries are dropped and recomputed on demand. No background goroutine. |
| **Not affected** | `sync.Pool` buffers (recycle memory, never reuse document content) |

Covered caches: font subsets, page compress output, compressed font bytes,
TrueType object maps, prop parses, decoded images, signers/PEM materials,
template-data files. BOPS benchmarks bypass them entirely with
`gopdflib.ClearBOPSCaches()`; see `sampledata/gopdflib/zerodha_bops/` and
`make bench-gopdflib-bops-x10`.

---

## Important distinction: “template” means three different things

| Term | What it is | Cached across requests? |
|------|------------|-------------------------|
| **Request JSON** (`PDFTemplate` from POST body) | Your per-document data (trades, text, etc.) | **No** - decoded per request, then cleared |
| **Shared row template row** (`sharedRowTemplateRow`) | One table row whose `props` define column layout for the HFT fast path | **No** - metadata inside one request’s JSON, not a server cache |
| **Zerodha benchmark templates** (`main.go` / `pypdfsuit_bench.py`) | Pre-built `PDFTemplate` structs for load tests | **No** - built once per benchmark **process**, not shared by the Gin server |

Generating millions of unique PDFs does **not** store millions of templates in RAM. Each HTTP request gets a fresh decode into a reused struct shell, generates the PDF, then resets that shell.

---

## Per-request lifecycle (Gin `/api/v1/generate/template-pdf`)

```
1. templatePDFPool.Get()     → empty *PDFTemplate shell (sync.Pool)
2. resetTemplate()           → zero previous request fields
3. decodeTemplateJSON()      → your JSON written into the struct
4. GenerateTemplatePDFBorrowed()
5. doc.Release()             → return PDF buffer to pool
6. resetTemplate() + Put()   → drop large slice backing arrays
```

**Expiration:** none (time-based). Memory is **request-scoped** - after the handler returns, payload slices are eligible for GC. The pool only keeps a small number of empty struct shells for allocation reuse.

**Code:** `internal/handlers/handlers.go` (`templatePDFPool`, `resetTemplate`), `internal/handlers/json_decode.go`

---

## Process-global caches (survive across requests)

These are the caches that matter for long-running servers and high cardinality workloads.

### 1. Page compression cache - **bounded, count-based eviction**

| Property | Value |
|----------|-------|
| **Purpose** | Reuse zlib output when page content streams are identical (common in repeated HFT pages) |
| **Key** | FNV-1a fingerprint of raw page bytes |
| **Max entries** | `2048` total, sharded across CPUs (~32–64 entries per shard) |
| **Eviction** | When a shard exceeds its cap → **`sync.Map.Clear()` on that entire shard** (not LRU, not TTL) |
| **TTL** | **None** |
| **Manual clear** | `font.ClearPageCompressCache()` |

**Millions of unique pages:** cache hit rate drops; memory stays bounded by shard clears.

**Code:** `internal/pdf/font/compress_cache.go`

---

### 2. Font subset cache - **unbounded, no expiration**

| Property | Value |
|----------|-------|
| **Purpose** | Reuse TTF subset bytes when the same font + glyph set appears again |
| **Key** | FNV fingerprint of PostScript name + sorted glyph IDs |
| **Eviction** | **None** - entries live until process exit |
| **TTL** | **None** |
| **Manual clear** | **No public API today** |

**Millions of unique glyph sets:** memory can grow. Typical brokerage workloads reuse a few fonts/glyph sets → cache stays small.

**Code:** `internal/pdf/font/subset_cache.go`

---

### 3. Image decode cache - **unbounded, no expiration**

| Property | Value |
|----------|-------|
| **Purpose** | Skip repeated PNG/JPEG decode for identical base64 image data |
| **Key** | FNV-1a hash of base64 string |
| **Eviction** | **None** |
| **TTL** | **None** |
| **Manual clear** | `pdf.ResetImageCache()` |

**Millions of unique images:** one `ImageObject` per distinct hash until cleared or process restart.

**Code:** `internal/pdf/image.go`

---

### 4. Template JSON file cache (`GET /template-data`) - **unbounded, no expiration**

| Property | Value |
|----------|-------|
| **Purpose** | Cache raw JSON bytes for sample template files on disk |
| **Key** | Absolute file path |
| **Eviction** | **None** |
| **TTL** | **None** |
| **Scope** | Dev/demo endpoint only - not used by `POST /generate/template-pdf` |

**Code:** `internal/handlers/handlers.go` (`templateDataCache sync.Map`)

---

## `sync.Pool` buffers (not long-lived caches)

These reuse allocations but **do not retain your data** across requests in a predictable way. The runtime may drop pool entries at any GC cycle.

| Pool | Max retained cap | Notes |
|------|------------------|-------|
| `bodyBufPool` | 128 KiB returned cap; larger arrays discarded | Retail/active JSON bodies ≤ 512 KiB |
| `hftBodyBufPool` | 8 MiB max; larger discarded | HFT JSON bodies |
| `pdfBufferPool` | Per-PDF output; released via `BorrowedPDF.Release()` | |
| `structElemPool` | Per-PDF structure nodes; released after generation | |
| `ZlibWriterPool` / `CompressBufPool` | Writer/buffer reuse | |

**Code:** `internal/handlers/json_decode.go`, `internal/pdf/generator.go`, `internal/pdf/font/compression.go`

---

## Per-PDF (not cross-request) reuse

| Mechanism | Lifetime |
|-----------|----------|
| `cachedPageInit` (border/watermark bytes) | One `PageManager` / one PDF |
| `sharedRowLayout` fast path | One table inside one PDF |
| `GeneratePDFBorrowed` | Caller must call `Release()` when done |

---

## Summary table: does it expire?

| Cache | TTL? | Bounded? | Safe for millions of **unique** payloads? |
|-------|------|----------|------------------------------------------|
| Request `PDFTemplate` data | N/A (not cached) | Per-request | **Yes** |
| `templatePDFPool` shell | GC | De facto bounded | **Yes** |
| Page compress cache | **Yes (3m default)** | **Yes (~2048)** | **Yes** |
| Font subset cache | **Yes (3m default)** | **Yes (1024)** | **Yes** - stale glyph sets age out |
| Compressed font / TrueType objects | **Yes (3m default)** | No (TTL is the bound) | **Yes** - entries expire even under cardinality |
| Props parse cache | **Yes (3m default)** | **Yes (8192)** | **Yes** |
| Image cache | **Yes (3m default)** | **Yes (256)** | **Yes** - unique images age out |
| Signer / PEM caches | **Yes (3m default)** | **Yes (FIFO 128 each)** | **Yes** - rotated certs re-parsed after TTL |
| `templateDataCache` | **Yes (3m default)** | **Yes (64 entries / 16 MiB)** | N/A for POST generate |
| HFT shared-row layout | N/A (not a cache) | Per-table | **Yes** |

---

## Recommendations at high volume

1. **Unique text-only PDFs (e.g. millions of contract notes):** default architecture is fine - payload JSON is not retained.
2. **Unique images per PDF:** call `pdf.ResetImageCache()` periodically, or restart workers on a schedule, or add a max-size eviction policy (not implemented today).
3. **Many custom fonts with unique glyph sets per doc:** monitor heap; subset cache is the main growth risk.
4. **Long-lived Gin pods:** compression cache self-caps; image/subset caches are the ones to watch in `debug/pprof/heap`.
5. **Do not confuse** Zerodha benchmark “Building templates…” with server caching - that only pre-builds structs inside the benchmark binary.

---

## Manual cache clearing (operations)

```go
import "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"

gopdflib.ClearBOPSCaches() // drops every content cache at once (BOPS path)

gopdflib.SetCacheTTL(2 * time.Minute) // shrink TTL; <= 0 disables expiry
```

Process restart clears all global caches. Individual clear functions also
remain available (`font.ClearSubsetCache`, `font.ClearPageCompressCache`,
`font.ClearFontObjectCaches`, `pdf.ClearPropsCache`, `pdf.ResetImageCache`,
`signature.ClearSignerCaches`).

---

## Related docs

- [CROSS_REQUEST_CACHING_FAQ.md](./CROSS_REQUEST_CACHING_FAQ.md) - does layout bleed across PDFs? (FAQ)
- [PERFORMANCE_OPTIMIZATIONS.md](../guides/additionalnotes/PERFORMANCE_OPTIMIZATIONS.md) - zlib pooling, image cache overview
- [PDF_GENERATION_INTERNALS.md](../guides/additionalnotes/PDF_GENERATION_INTERNALS.md) - assets and streams
- [TEMPLATE_REFERENCE.md](./TEMPLATE_REFERENCE.md) - `sharedRowLayout` / `sharedRowTemplateRow` fields