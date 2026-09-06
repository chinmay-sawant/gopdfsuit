# Engine Overview (`internal/pdf`)

In-memory template-to-PDF engine in Go. No Ghostscript, no browser.
`internal/pdf` renders, `pkg/gopdflib` exposes it to Go callers,
`cmd/gopdfsuit` serves it over HTTP, `bindings/python` wraps it for Python.

Related docs: [ARCHITECTURE.md](ARCHITECTURE.md),
[CACHING_AND_MEMORY_LIFECYCLE.md](CACHING_AND_MEMORY_LIFECYCLE.md),
[COMPLIANCE_PIPELINE_TODAY.md](COMPLIANCE_PIPELINE_TODAY.md).

## Template shape (`internal/models`)

Request JSON decodes into `models.PDFTemplate`:

- `Config`: page size, margins, watermark, bookmarks, plus opt-in
  `Security`, `PDFA`, `Signature` blocks and `CustomFonts`.
- `Title`, `Table` (`maxcolumns`, `rows`, `columnwidths`,
  `sharedRowLayout`), `Spacer`, `Image`, ordered `Elements`.
- `Row` holds `Cell` items; each cell carries `props`, `text`,
  colors, link, wrap, optional image or form field.

`Config` flags that change the pipeline: `security.enabled`,
`pdfaCompliant`, `taggedPDF`, `signature.enabled`, `embedFonts`.

## Pipeline overview

One borrowed-buffer generate run moves through these phases:

1. **Decode with Sonic.** Handlers acquire a pooled `PDFTemplate`,
   preallocate slices from `Content-Length` plus tier, then decode
   with pooled Sonic (`bytedance/sonic`). Tiers: HFT under 8 MiB,
   retail under 512 KiB, streaming above. See
   [ARCHITECTURE.md](ARCHITECTURE.md) for the handler path.
2. **Render.** The `generation` struct plus `PageManager` lay out pages,
   assign object IDs via `Allocator`, emit content streams in
   `draw.go` (integer-math number formatting, cell layout), images
   and SVG through the image and `svg/` paths.
3. **Fonts.** Standard 14 fonts inline; custom TTF/OTF parsed in
   `font/`, subset to used glyphs, embedded as `FontFile2` with
   `ToUnicode` plus `CIDToGIDMap` (see Fonts below).
4. **Tagging.** `structure.go` emits the structure tree and marked
   content (`BDC`/`EMC`) for PDF/UA-2; `TaggedPDF` or
   `pdfaCompliant` enables it.
5. **XMP.** `metadata.go` plus `pdfa.go` emit the XMP packet and
   output intent with pre-compressed ICC profiles.
6. **Sign.** `signature/` appends a PKCS#7 signature with a
   `ByteRange` placeholder, then patches the range in place.
7. **Encrypt.** `encryption/` applies password protection when
   `security.enabled` is set (see note below on AES-128 vs AES-256).

Zero-copy contract: `BorrowedPDF` hands a pooled buffer;
`Bytes()` borrows, `CopyBytes()` clones,
`Release()` returns exactly once.

```go
doc, err := pdf.GenerateTemplatePDFBorrowed(tmpl)
defer doc.Release()
```

## Generator

- Entry: `GenerateTemplatePDF` clones, `GenerateTemplatePDFBorrowed`
  returns the pooled buffer.
- The `generation` struct holds template, page manager, allocator, and image
  dedup maps so image XObjects are emitted once.
- `PageManager` sizes per-page content-stream buffers from measured
  tiers (retail 32 KiB, active 48 KiB, HFT stripe estimate).
- `draw.go` formats floats with integer math and
  clamps cell text origins; outline, destinations, and
  links modules add bookmarks and navigation.

## Merge / Split (`merge/`)

- `merge/merger.go`: parses objects, remaps numbers, builds one new
  `/Pages` tree preserving input order. Guards: 50000 max objects,
  32 MiB per input, 128 MiB total; encrypted inputs rejected.
- `merge/split.go`: `SplitSpec` plus `ParsePageSpec("1-3,5,7-9")`
  selects 1-based pages into new files.

## Compress tiers (`compress/`)

Ghostscript-style presets, pure Go, no external binary:

| Tier | JPEG quality | Max image edge |
|------|--------------|----------------|
| light | 92 | 1920 |
| medium (default) | 75 | 1275 |
| heavy | 50 | 612 |

`Options` allows `JPEGQuality` / `MaxImageDim` overrides; empty level
selects medium. Streams are rewritten and images downscaled within
`limits.go` bounds.

## Redact (`redact/`)

`Redactor` caches the parsed object map so
search plus paint need one parse:

- `search.go`: text query to rectangles.
- `visual.go`: paints opaque boxes (`ApplyRedactions`).
- `secure.go`: removes underlying content, not just paint-over.
- `pageboxes.go`: display space to MediaBox mapping.
- Encrypted PDFs are rejected unless the password path
  (`ApplyRedactionsAdvancedWithReport`) is used.

## Signatures (`signature/`)

PKCS#7 / X.509 signing:

- RSA and ECDSA P-256; other key types rejected.
- DER is hand-built, digest covers the `ByteRange` gaps,
  placeholder is patched in place so offsets stay stable.
- Config: `SignatureConfig` with PEM cert plus key,
  optional chain, visible stamp geometry, reason/location/contact.

## Encryption (`encryption/`) - strength note

Code comment and implementation use **AES-128**
(16-byte key via `aes.NewCipher`), with `/U`, `/O`, `/P` derived from
`SecurityConfig`. Top-level docs say AES-256; the source says AES-128,
so this file records AES-128. Do not relabel it AES-256 without a code change.

## Fonts and subsetting (`font/`)

- `ttf.go` parses TTF/OTF; `subset.go` keeps `.notdef` plus used
  glyphs plus composite dependencies, then rebuilds `cmap`, `glyf`,
  `loca`, `maxp` and friends.
- `registry.go`: global `CustomFontRegistry`; WASM builds register
  from base64 (`fontData`), server builds also accept `filePath`.
- Emit: Type0 plus CIDFontType2 with `Identity-H`,
  `FontDescriptor` with `FontFile2`, `Widths`, `CIDToGIDMap`,
  `ToUnicode` CMap for extraction. Standard 14 fonts stay
  non-embedded unless PDF/A mode requires metrics.

## HTML via gowkhtmltopdf (`html_convert.go`)

Pure-Go path through `chinmay-sawant/gowkhtmltopdf`, no browser.
`buildPDFDocument` maps `PageSize`, `Orientation`, margins,
`Grayscale` onto a restricted-network `Document`; image ops map
`Format`, `Width`, `Height`, `Quality`, `Zoom`, `Crop*`. `DPI`,
`LowQuality`, and free-form `Options` are accepted and ignored with
a warning log.

## Cache and memory (TTL note)

Shared policy in `internal/cachettl`, full inventory in
[CACHING_AND_MEMORY_LIFECYCLE.md](CACHING_AND_MEMORY_LIFECYCLE.md):

- Default TTL 3 minutes, override with `GOPDFSUIT_CACHE_TTL`
  (Go duration) or `gopdflib.SetCacheTTL(d)`; TTL `<= 0` restores
  size-only eviction. Lazy expiry on lookup, no background goroutine.
- Covered: font subsets, compressed font bytes, decoded images,
  page compress output, prop parses, signer materials,
  template-data files.
- Not cached across requests: the request `PDFTemplate` itself.
  Handlers borrow a struct shell from `sync.Pool`, decode, render,
  release the PDF buffer, reset and return the shell. `sync.Pool`
  buffers recycle memory only, never document content.
- Compliance angle: veraPDF is the hard gate, `structure_tree_check.py`
  is the hard structure gate (ParentTree MCID must point at TD/TH,
  not TR); avalpdf warns unless strict. Details in
  [COMPLIANCE_PIPELINE_TODAY.md](COMPLIANCE_PIPELINE_TODAY.md).
