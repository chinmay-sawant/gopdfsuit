# Getting started with gopdflib

I use gopdflib when I need PDFs straight from Go code without shelling out to another tool. It generates, redacts, and compresses in one import, and templates stay as JSON you can version.

## Table of contents

1.  [Downloading and installing](#downloading-and-installing)
2.  [Loading PDF templates from JSON](#loading-pdf-templates-from-json)
3.  [Redacting a PDF](#redacting-a-pdf)
4.  [Compressing a PDF](#compressing-a-pdf)

---

## Downloading and installing

Install the `gopdflib` package with the `v7.0.0` release tag.

### Prerequisites

- You need Go 1.21 or later.

### Steps to download

1.  Run the following command in your terminal to download the package:

    ```bash
    go get github.com/chinmay-sawant/gopdfsuit/v7@v7.0.0
    ```

    This command downloads the source code and adds the dependency to your `go.mod` file.

### Usage in your project

To use the library in your Go code, import the `gopdflib` package:

```go
import (
    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)
```

#### Basic configuration example

Reference the library in your code like this:

```go
package main

import (
    "fmt"
    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    // Example: Create a simple configuration
    config := gopdflib.Config{
        Page:          "A4",
        PageAlignment: 1, // Portrait
    }

    fmt.Printf("gopdflib Config initialized: %+v\n", config)
}
```

### Updating the library

To update to a specific v7 release in the future, run:

```bash
go get github.com/chinmay-sawant/gopdfsuit/v7@v7.0.0
```

---

## Loading PDF templates from JSON

Load template data from a JSON file to keep content out of Go code, or to accept template data from an API.

### Overview

The `gopdflib.PDFTemplate` struct tags match standard JSON naming conventions (camelCase), so you can unmarshal JSON data directly into the struct.

### Prerequisites

- `gopdflib` installed (as described above)
- A valid JSON template file (e.g., `sampledata/editor/financial_digitalsignature.json`)

### Example code

Create a file named `main.go` (or similar) with the following content:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    // 1. Define the path to your JSON file
    jsonPath := "sampledata/editor/financial_digitalsignature.json"

    // 2. Read the file
    jsonData, err := os.ReadFile(jsonPath)
    if err != nil {
        panic(fmt.Errorf("failed to read file: %w", err))
    }

    // 3. Unmarshal directly into PDFTemplate
    var template gopdflib.PDFTemplate
    err = json.Unmarshal(jsonData, &template)
    if err != nil {
        panic(fmt.Errorf("failed to parse JSON: %w", err))
    }

    // 4. Generate the PDF
    pdfBytes, err := gopdflib.GeneratePDF(template)
    if err != nil {
        panic(fmt.Errorf("failed to generate PDF: %w", err))
    }

    // 5. Save or use the PDF bytes
    outputPath := "output.pdf"
    err = os.WriteFile(outputPath, pdfBytes, 0644)
    if err != nil {
        panic(fmt.Errorf("failed to save PDF: %w", err))
    }

    fmt.Printf("PDF successfully generated and saved to %s!\n", outputPath)
}
```

### Running the sample

Run the ready-made example in the repo.

1.  Navigate to the project root.
2.  Run the example:

    ```bash
    go run sampledata/gopdflib/load_from_json/main.go
    ```

    This reads `sampledata/editor/financial_digitalsignature.json`, generates the PDF, and saves it as `financial_from_json.pdf` in your current directory.

### JSON structure

The JSON structure mirrors the `gopdflib.PDFTemplate` struct. Common top-level fields include:

- `config`: Page settings (size, margin, etc.)
- `title`: Document title section
- `elements`: Array of content elements (tables, spacers, images)
- `footer`: Footer configuration
- `bookmarks`: Navigation outline

Refer to the `sampledata/editor/financial_digitalsignature.json` file for a full example of the JSON schema.

---

## Redacting a PDF

Scrub sensitive info from a PDF with `gopdflib`. Redact by coordinates, or search and redact text.

### Overview

Call `gopdflib.ApplyRedactionsAdvanced` with explicit coordinates and text queries to redact a PDF visually and structurally.

### Example code

Create a file named `main.go` with the following content:

```go
package main

import (
    "fmt"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    pdfPath := "sample.pdf"

    // 1. Read the input PDF
    pdfBytes, err := os.ReadFile(pdfPath)
    if err != nil {
        panic(fmt.Errorf("failed to read file: %w", err))
    }

    // 2. Define redaction options
    options := gopdflib.ApplyRedactionOptions{
        TextSearch: []gopdflib.RedactionTextQuery{
            {Text: "Confidential"},
            {Text: "Secret"},
        },
        // You can also redact specific regions:
        // Blocks: []gopdflib.RedactionRect{
        //     {PageNum: 1, X: 100, Y: 100, Width: 200, Height: 20},
        // },
        Mode: "visual_allowed", // optional
    }

    // 3. Apply the redactions
    redactedBytes, err := gopdflib.ApplyRedactionsAdvanced(pdfBytes, options)
    if err != nil {
        panic(fmt.Errorf("failed to redact PDF: %w", err))
    }

    // 4. Save the redacted PDF
    outputPath := "redacted.pdf"
    err = os.WriteFile(outputPath, redactedBytes, 0644)
    if err != nil {
        panic(fmt.Errorf("failed to save PDF: %w", err))
    }

    fmt.Printf("PDF successfully redacted and saved to %s!\n", outputPath)
}
```

---

## Compressing a PDF

`gopdflib.CompressPDF` rewrites an existing PDF: bicubic image downsample and JPEG recompression at a chosen tier, unused TTF glyph outlines dropped, document metadata stripped, and streams Flate-compressed. Encrypted files are rejected. Input larger than 32 MiB is rejected. If the rewrite is not smaller, the original bytes are returned.

There is no CLI. The same engine also runs in the browser as WebAssembly (`make wasm-compress`); see `sampledata/compress-js`.

### Levels

| Level | Constant | JPEG quality | Max image edge |
|-------|----------|--------------|----------------|
| Light | `gopdflib.CompressLight` | 92 | 1920 |
| Medium (default) | `gopdflib.CompressMedium` | 75 | 1275 |
| Heavy | `gopdflib.CompressHeavy` | 50 | 612 |

`JPEGQuality` and `MaxImageDim` on `CompressOptions` override the preset when greater than zero.

### Example code

```go
package main

import (
    "fmt"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
    src, err := os.ReadFile("document.pdf")
    if err != nil {
        panic(err)
    }

    out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{
        Level: gopdflib.CompressHeavy,
    })
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("document-compressed.pdf", out, 0644); err != nil {
        panic(err)
    }
    fmt.Printf("original=%d compressed=%d\n", len(src), len(out))
}
```

### Running the sample

```bash
cd sampledata/compress && go run .
```

This reads `report.pdf` and writes `report_level_1.pdf` (Light), `report_level_2.pdf` (Medium), and `report_level_3.pdf` (Heavy).

The JavaScript/WASM sample is separate:

```bash
make wasm-compress
node sampledata/compress-js/run.mjs
```

---

## Browser-local (WASM) vs server-only ops

Pure-Go ops also run in the browser as WebAssembly (`make wasm`, demo in
`sampledata/wasm-js/`). The file never leaves the tab: no `POST /api/v1/*`.

| Op | Browser-local (WASM) | Server-only |
|----|----------------------|-------------|
| Generate | yes (`goGeneratePDF`) | also via API |
| Generate, PDF/A-compliant | yes, after `ensurePDFAFonts()` fetches `/fonts/*.ttf` and `goRegisterFont` registers all 12 faces (`wasm/compliance.js`, demo `run_compliant.mjs`) | also via API |
| Merge | yes (`goMergePDF`) | also via API |
| Split | yes (`goSplitPDF`, returns JS array, zip in JS) | also via API |
| Compress | yes (`goCompressPDF`, Light/Medium/Heavy, Worker path) | also via API |
| Fill (XFDF) | yes (`goFillPDF`, two `Uint8Array`) | also via API |
| Redact | yes, **text path only** (`goRedactSearch`/`goRedactApply`, no OCR) | full path incl. OCR via API |
| HTML to PDF/Image | no | yes, server-side only |
| veraPDF validation | no (runs server-side; browser output is verified by uploading bytes or regenerating server-side) | yes |

Redact in WASM leaves the `OCR` field unset: there is no
pdftoppm/tesseract subprocess in a browser tab, so image-only pages are
reported, not OCRed. Fill notes one limit: `/NeedAppearances true` is a
byte-level patch, so files whose AcroForm lives inside a compressed
object stream may not auto-regenerate appearances on open.
