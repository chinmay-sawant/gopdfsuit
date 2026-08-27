## Summary

Adds a Ghostscript-free PDF compressor on the existing engine: `gopdflib.CompressPDF` and `POST /api/v1/compress`. Input PDF bytes come back smaller via bicubic image downsample, JPEG recompress, unused TTF glyph drop, metadata strip, and Flate streams. Encrypted files are rejected.

---

## Motivation / context

- Plans: `plans/pdf-compress-nogs.md`
- Issues: see **Related issues**
- Compression used to live as leftover `internal/pdf/compress` with v5 imports and no public API. This wires it through gopdflib and the HTTP service, with Light / Medium / Heavy tiers matching the Ghostscript product.

---

## Changes

### Engine (`internal/pdf/compress`)

- `CompressPDF([]byte, Options)` is the only engine entry.
- Tiers: Light (JPEG 92, max edge 1920), Medium (75 / 1275, default), Heavy (50 / 612). `JPEGQuality` and `MaxImageDim` override the preset when > 0.
- Bicubic downsample (Keys, a = -0.5), Flate recompress at BestCompression, trailer `/Info` + `/ID` stripped, unused TTF glyph outlines dropped (GID-preserving).
- Non-PDF and encrypted input return an error. If the rewrite is not smaller, the original bytes are returned.

### Library (`pkg/gopdflib`)

- Public `CompressPDF` + `CompressOptions` alias and `CompressLight` / `CompressMedium` / `CompressHeavy`.
- Docs and example updated.

### HTTP API

- `PDFService.CompressPDF` + `POST /api/v1/compress`.
- Form field `pdf` required; optional `level`, `quality`, `max_image_dim`.
- Success: `application/pdf`, `filename=compressed.pdf`.

### Tests and sample

- Library tests: non-PDF error, image PDF shrinks, metadata strip, level override.
- Handler gomock: success with parsed options, missing-file 400.
- `sampledata/compress` demo on `report.pdf` at all three levels.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Extra CPU on compress (bicubic + JPEG + Flate BestCompression). Generation-path zlib BestSpeed is unchanged. |
| **Memory** | Whole-PDF object map in memory; large image PDFs allocate decode + downsample buffers. |
| **Behavior / correctness** | Text and vector operators stay intact. Encrypted / JPEG2000 / JBIG2 / filter-array PDFs are rejected or left as-is. If rewrite is not smaller, original bytes are returned. |
| **API / CLI** | New `gopdflib.CompressPDF`. New `POST /api/v1/compress`. No CLI. |
| **Dependencies** | None. |
| **Binary size / build time** | Small: compress + font compact packages only. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Additive API. Existing generate / merge / split paths unchanged. |

---

## Test plan

- [x] `make test`
- [x] `make lint` / `go vet`
- [x] `go test ./pkg/gopdflib/ ./internal/handlers/ ./internal/pdf/compress/`
- [x] `cd sampledata/compress && go run .` (sizes below)

### Commands

```sh
make lint
make test
go test ./pkg/gopdflib/ ./internal/handlers/
cd sampledata/compress && go run .
```

---

## Screenshots / sample output

`sampledata/compress` on `report.pdf` (596341 bytes):

```
source  report.pdf  596341 bytes
level 1 Light   report_level_1.pdf  530762 bytes  (11.0% smaller)
level 2 Medium  report_level_2.pdf  140897 bytes  (76.4% smaller)
level 3 Heavy   report_level_3.pdf   74890 bytes  (87.4% smaller)
```

---

## Related issues

- Closes #77

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [ ] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- In-browser WASM compress (plan Phase 7): no CLI, no `POST /api/v1/compress` from the UI.
- Python CGO export.
- CFF / CID remapping; JPEG2000 / JBIG2 / encrypted files.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in the diff
- [ ] Public API / CLI changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
