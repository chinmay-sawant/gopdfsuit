# Cross-Request Caching FAQ

**Date:** 2026-06-11

This document answers a common question: *does gopdfsuit cache row layout or design from one PDF and reuse it on the next?* And when could caching cause problems?

See also: [CACHING_AND_MEMORY_LIFECYCLE.md](./CACHING_AND_MEMORY_LIFECYCLE.md) for cache inventory, bounds, and manual clear APIs.

---

## Does the server cache row/design from PDF A and apply it to PDF B?

**No.** That would be a serious bug. That is not how the system works.

When PDF #2 arrives, the server does **not**:

- Remember PDF #1’s table layout and reuse it
- Cache “row design” globally from earlier requests
- Auto-enable the HFT fast path for unrelated PDFs

Each `POST /api/v1/generate/template-pdf` request:

1. Gets an empty `PDFTemplate` shell from `templatePDFPool`
2. `resetTemplate()` clears anything left from the prior request
3. Your **new JSON** is decoded into that struct
4. PDF is generated from **that JSON only**
5. Struct is reset again and returned to the pool

PDF B is drawn from PDF B’s JSON, not PDF A’s design.

**Code:** `internal/handlers/handlers.go` (`handleGenerateTemplatePDF`, `resetTemplate`)

---

## What *is* shared across requests (and why it is usually safe)

These are **content-dedup caches**, not layout caches:

| Cache | What it stores | When it hits |
|-------|----------------|--------------|
| **Page zlib cache** | Compressed bytes for a page content stream | Only if **raw page bytes are identical** (same fingerprint + length) |
| **Image cache** | Decoded image for base64 data | Only if the **same image bytes** appear again |
| **Font subset cache** | Subset TTF for font + glyph set | Only if the **same glyphs** are used again |

**Example:** 1 million contract notes with the **same logo** → logo decode is cached once. Text and trade rows differ → no cross-PDF layout bleed.

These caches return **read-only reused bytes**. They do not change your template structure or row design.

---

## HFT fast path ≠ cross-request cache

`sharedRowLayout` and `sharedRowTemplateRow` are **fields in your JSON** for **one table in one request**.

| PDF | Behavior |
|-----|----------|
| PDF A with `sharedRowLayout: true` | Fast path for A’s table only |
| PDF B without it | Normal per-cell path |
| PDF C with different `props` | Uses C’s JSON; nothing inherited from A |

The “template row” is a row index (e.g. `1`) **inside that request’s table**, not a server-side remembered design.

**Code:** `internal/pdf/draw.go` (`sharedRowTemplateIndex`, `drawSharedLayoutRow`)

---

## When could there be an issue?

| Scenario | Risk |
|----------|------|
| Same layout, different text (millions of contract notes) | **Safe** — intended use |
| Different layouts per PDF | **Safe** — each JSON decoded fresh |
| Millions of **unique images** | **Memory growth** (image cache has no eviction) — not wrong PDF content |
| Millions of **unique font glyph sets** | **Memory growth** (subset cache has no eviction) — not wrong layout |
| `sharedRowLayout: true` but data rows **don’t** have uniform `props` | **Wrong rendering within that one PDF** — not cross-PDF contamination |
| Hypothetical hash collision on page content | **Theoretically possible, practically negligible** |

---

## Bottom line

| Statement | True? |
|-----------|-------|
| Caches row/design from arbitrary new PDFs for the next PDF | **No** |
| Caches identical binary content to skip repeat work | **Yes** |
| Clears per-request template data after each generate call | **Yes** |
| HFT fast path is opt-in per table in JSON | **Yes** |

For long-lived servers with huge variety: periodic `pdf.ResetImageCache()` and `font.ClearPageCompressCache()`, or worker restarts — controls **memory**, not layout correctness.

```go
import (
    "github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf"
    "github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/font"
)

font.ClearPageCompressCache()
pdf.ResetImageCache()
```

---

## Related docs

- [CACHING_AND_MEMORY_LIFECYCLE.md](./CACHING_AND_MEMORY_LIFECYCLE.md) — full cache list, eviction, millions-of-records guidance
- [TEMPLATE_REFERENCE.md](./TEMPLATE_REFERENCE.md) — `sharedRowLayout` / `sharedRowTemplateRow` JSON fields
- [INTEGRATION_AND_BENCHMARK_TESTS.md](./INTEGRATION_AND_BENCHMARK_TESTS.md) — benchmarks and ops notes