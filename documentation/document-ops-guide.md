# Document Ops Guide: Merge, Split, Fill, Redact

Operations on existing PDFs: combine files, extract pages, fill AcroForm
data from XFDF, and redact by coordinates or text search.

Sources of truth:

- Go library: `pkg/gopdflib/merge.go` (`MergePDFs`), `pkg/gopdflib/split.go`
  (`SplitPDF`, `ParsePageSpec`, `ParseSplitSpecJSON`,
  `SplitPDFWithSpecJSON`), `pkg/gopdflib/fill.go` (`FillPDFWithXFDF`),
  `pkg/gopdflib/redact.go` (`GetPageInfo`, `ExtractTextPositions`,
  `FindTextOccurrences`, `ApplyRedactions`, `ApplyRedactionsAdvanced`,
  `ApplyRedactionsAdvancedWithReport`, `AnalyzePageCapabilities`)
- Shared types: `pkg/gopdflib/types.go` (`SplitSpec`, `RedactionRect`,
  `ApplyRedactionOptions`, `RedactionTextQuery`)
- Python mirrors: `bindings/python/pypdfsuit/merge.py` (`merge_pdfs`),
  `bindings/python/pypdfsuit/split.py` (`split_pdf`, `parse_page_spec`),
  `bindings/python/pypdfsuit/fill.py` (`fill_pdf_with_xfdf`),
  `bindings/python/pypdfsuit/redact.py` (`get_page_info`,
  `extract_text_positions`, `apply_redactions`, `find_text_occurrences`,
  `apply_redactions_advanced`)
- Fixtures: [sampledata/merge](../sampledata/merge),
  [sampledata/split](../sampledata/split),
  [sampledata/filler](../sampledata/filler)

All REST examples assume the server on `http://localhost:8080`.

## 1. Merge

Combines two or more PDFs in caller-supplied order into one PDF.

Rules: at least one non-empty file is required. File count, per-file, and
combined-input caps apply. REST endpoint `POST /api/v1/merge` takes repeated
multipart field `pdf` and returns `application/pdf` (`merged.pdf`).

Fixtures: [sampledata/merge](../sampledata/merge) (`em-16.pdf`,
`em-19.pdf`, `em-51.pdf`).

### Go

```go
import (
    "os"
    "log"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

pdf1, _ := os.ReadFile("sampledata/merge/em-16.pdf")
pdf2, _ := os.ReadFile("sampledata/merge/em-19.pdf")

merged, err := gopdflib.MergePDFs([][]byte{pdf1, pdf2})
if err != nil {
    log.Fatal(err)
}
os.WriteFile("merged.pdf", merged, 0644)
```

### Python

```python
from pypdfsuit import merge_pdfs

with open("sampledata/merge/em-16.pdf", "rb") as f1, \
     open("sampledata/merge/em-19.pdf", "rb") as f2:
    merged = merge_pdfs([f1.read(), f2.read()])

with open("merged.pdf", "wb") as f:
    f.write(merged)
```

### curl

```bash
curl -X POST http://localhost:8080/api/v1/merge \
  -F "pdf=@sampledata/merge/em-16.pdf;type=application/pdf" \
  -F "pdf=@sampledata/merge/em-19.pdf;type=application/pdf" \
  --output merged.pdf
```

Order of `-F "pdf=..."` flags is the page order in the output.

## 2. Split

Splits one PDF into one or more PDFs. Three selector styles share one
`SplitSpec`:

```go
type SplitSpec struct {
    Pages      []int    `json:"pages,omitempty"`
    Ranges     [][2]int `json:"ranges,omitempty"`
    MaxPerFile int      `json:"maxPerFile,omitempty"`
}
```

### Page-spec syntax

`ParsePageSpec` parses strings like `"1-3,5,7-9"` into sorted
1-based page numbers:

- `"1-3,5,7-9"` with 10 total pages gives `[1, 2, 3, 5, 7, 8, 9]`.
- Empty or whitespace-only spec selects no pages (returns empty slice,
  no error). The WASM/JSON path treats nil or empty input as all pages in one
  file.
- `totalPages > 0` validates against the page count; `0` skips
  validation.
- Invalid tokens and reversed ranges (`[5,2]`) are rejected.

REST form fields on `POST /api/v1/split`: `pdf` file, optional `pages`
string (for example `"1-3,5"`), optional `max_per_file` positive int.
One output part returns `application/pdf`; multiple parts return
`application/zip`.

Fixtures: [sampledata/split](../sampledata/split) (`em.pdf`,
`split.pdf`, `split_range.pdf`).

### Go

```go
import (
    "os"
    "log"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

pdfBytes, _ := os.ReadFile("sampledata/split/em.pdf")

// Option A: explicit pages via page-spec string.
pages, err := gopdflib.ParsePageSpec("1-3,5", 0)
if err != nil {
    log.Fatal(err)
}
parts, err := gopdflib.SplitPDF(pdfBytes, gopdflib.SplitSpec{Pages: pages})
if err != nil {
    log.Fatal(err)
}

// Option B: every N pages.
parts, err = gopdflib.SplitPDF(pdfBytes, gopdflib.SplitSpec{MaxPerFile: 5})
if err != nil {
    log.Fatal(err)
}

// Option C: inclusive ranges.
parts, err = gopdflib.SplitPDF(pdfBytes, gopdflib.SplitSpec{
    Ranges: [][2]int{{1, 3}, {5, 5}},
})
if err != nil {
    log.Fatal(err)
}

for i, part := range parts {
    os.WriteFile("part.pdf", part, 0644)
    _ = i
}
```

### Python

```python
from pypdfsuit import split_pdf, parse_page_spec, SplitSpec

with open("sampledata/split/em.pdf", "rb") as f:
    pdf_data = f.read()

# Same "1-3,5" syntax, 0 skips total-page validation.
pages = parse_page_spec("1-3,5", 0)
parts = split_pdf(pdf_data, SplitSpec(pages=[1, 3, 5]))

# Or split every N pages.
parts = split_pdf(pdf_data, SplitSpec(max_per_file=5))

for i, part in enumerate(parts, start=1):
    with open(f"part-{i}.pdf", "wb") as f:
        f.write(part)
```

### curl

```bash
# Specific pages (single PDF back when only one part results).
curl -X POST http://localhost:8080/api/v1/split \
  -F "pdf=@sampledata/split/em.pdf;type=application/pdf" \
  -F "pages=1-3,5" \
  --output split.pdf

# Chunk every 2 pages (ZIP back when multiple parts result).
curl -X POST http://localhost:8080/api/v1/split \
  -F "pdf=@sampledata/split/em.pdf;type=application/pdf" \
  -F "max_per_file=2" \
  --output splits.zip
```

## 3. Fill (XFDF)

Fills an AcroForm PDF with XFDF (XML Forms Data Format) data.

REST endpoint `POST /api/v1/fill` takes multipart fields `pdf` and `xfdf`
and returns `application/pdf` (`filled.pdf`).

Fixtures: [sampledata/filler](../sampledata/filler)
(`us_hospital_encounter_acroform.pdf`,
`us_hospital_encounter_data.xfdf`) plus compressed-stream fixtures in
[sampledata/filler/compressed](../sampledata/filler/compressed).

### Compressed object stream caveat

- The `/NeedAppearances true` flag is applied as a byte-level patch to
  the AcroForm dictionary.
- When the AcroForm dictionary lives inside a compressed object stream
  (`/ObjStm`), the byte patch cannot reach it, so viewers may not
  regenerate field appearances on open for such files.
- Field values written into object streams are still updated by the
  object-stream-aware fill path. Only the appearance flag is affected.

If a filled file shows correct values in one viewer but blank fields in
another, check whether the form uses object streams (see the
`sampledata/filler/compressed` fixtures). The pdfcpu-based path
is the fallback for appearance regeneration on such files.

### Go

```go
import (
    "os"
    "log"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

pdfBytes, _ := os.ReadFile("sampledata/filler/us_hospital_encounter_acroform.pdf")
xfdfBytes, _ := os.ReadFile("sampledata/filler/us_hospital_encounter_data.xfdf")

filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
if err != nil {
    log.Fatal(err)
}
os.WriteFile("filled.pdf", filled, 0644)
```

### Python

```python
from pypdfsuit import fill_pdf_with_xfdf

with open("sampledata/filler/us_hospital_encounter_acroform.pdf", "rb") as f:
    pdf_data = f.read()
with open("sampledata/filler/us_hospital_encounter_data.xfdf", "rb") as f:
    xfdf_data = f.read()

filled = fill_pdf_with_xfdf(pdf_data, xfdf_data)

with open("filled.pdf", "wb") as f:
    f.write(filled)
```

### curl

```bash
curl -X POST http://localhost:8080/api/v1/fill \
  -F "pdf=@sampledata/filler/us_hospital_encounter_acroform.pdf;type=application/pdf" \
  -F "xfdf=@sampledata/filler/us_hospital_encounter_data.xfdf;type=application/xml" \
  --output filled.pdf
```

## 4. Redact

Two modes share the redaction engine:

- Coordinate mode: caller supplies rectangles (`RedactionRect` with
  `pageNum`, `x`, `y`, `width`, `height`) via `ApplyRedactions`.
- Text-search mode: caller supplies search strings and the engine finds
  match rectangles (`FindTextOccurrences`) or redacts them in one call
  (`ApplyRedactionsAdvanced` with `textSearch`).

Planning helpers: `GetPageInfo` (page count and dimensions),
`ExtractTextPositions` (text chunks plus coordinates for one 1-based
page), `AnalyzePageCapabilities` (which pages have searchable text vs.
need OCR). `ApplyRedactionsAdvancedWithReport` returns the PDF plus a
report (`mode`, `securityOutcome`, `appliedSecure`,
`appliedVisual`, `generatedRects`, `matchedTextCount`, `warnings`).

REST endpoints:

- `POST /api/v1/redact/page-info` (`pdf` file, returns page info JSON)
- `POST /api/v1/redact/text-positions` (`pdf` file plus `page` form
  field, returns text-position JSON)
- `POST /api/v1/redact/capabilities` (`pdf` file)
- `POST /api/v1/redact/search` (`pdf` file plus `text` or `texts`)
- `POST /api/v1/redact/apply` (`pdf` file plus `blocks` JSON,
  `textSearch` JSON, `mode`, `password`, `ocr`; returns
  `application/pdf` with `X-Redaction-Report` header)

OCR note: OCR-enabled options need the
pdftoppm/tesseract subprocess pipeline and are rejected in the WASM
browser path. Leave OCR unset for searchable-text files.

### Go

```go
import (
    "log"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

// Coordinate mode: cover a known rectangle on page 1.
rects := []gopdflib.RedactionRect{
    {PageNum: 1, X: 100, Y: 200, Width: 150, Height: 20},
}
redacted, err := gopdflib.ApplyRedactions(pdfBytes, rects)
if err != nil {
    log.Fatal(err)
}

// Text-search mode: two-step (inspect, then apply).
matches, err := gopdflib.FindTextOccurrences(pdfBytes, "Confidential")
if err != nil {
    log.Fatal(err)
}
redacted, err = gopdflib.ApplyRedactions(pdfBytes, matches)
if err != nil {
    log.Fatal(err)
}

// Text-search mode: one-step with report.
out, report, err := gopdflib.ApplyRedactionsAdvancedWithReport(pdfBytes,
    gopdflib.ApplyRedactionOptions{
        Blocks: []gopdflib.RedactionRect{
            {PageNum: 1, X: 100, Y: 200, Width: 150, Height: 20},
        },
        TextSearch: []gopdflib.RedactionTextQuery{{Text: "Confidential"}},
    })
if err != nil {
    log.Fatal(err)
}
_ = out
_ = report
```

### Python

```python
from pypdfsuit import redact as redact_ops

with open("document.pdf", "rb") as f:
    pdf_data = f.read()

# Planning helpers.
info = redact_ops.get_page_info(pdf_data)
positions = redact_ops.extract_text_positions(pdf_data, 1)

# Coordinate mode.
redacted = redact_ops.apply_redactions(pdf_data, [
    {"pageNum": 1, "x": 100, "y": 200, "width": 150, "height": 20},
])

# Text-search mode: two-step.
rects = redact_ops.find_text_occurrences(pdf_data, "Confidential")
redacted = redact_ops.apply_redactions(pdf_data, rects)

# Text-search mode: one-step.
redacted = redact_ops.apply_redactions_advanced(pdf_data, {
    "blocks": [{"pageNum": 1, "x": 100, "y": 200, "width": 150, "height": 20}],
    "textSearch": [{"text": "Confidential"}],
})

with open("redacted.pdf", "wb") as f:
    f.write(redacted)
```

### curl

```bash
# Preview match rectangles before applying.
curl -X POST http://localhost:8080/api/v1/redact/search \
  -F "pdf=@document.pdf;type=application/pdf" \
  -F "text=Confidential"

# Coordinate mode.
curl -X POST http://localhost:8080/api/v1/redact/apply \
  -F "pdf=@document.pdf;type=application/pdf" \
  -F 'blocks=[{"pageNum":1,"x":10,"y":10,"width":100,"height":20}]' \
  --output redacted.pdf

# Text-search mode.
curl -X POST http://localhost:8080/api/v1/redact/apply \
  -F "pdf=@document.pdf;type=application/pdf" \
  -F 'textSearch=[{"text":"Confidential"}]' \
  --output redacted.pdf
```
