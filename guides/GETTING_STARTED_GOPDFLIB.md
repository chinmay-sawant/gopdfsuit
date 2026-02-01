# Getting Started with gopdflib

This guide provides a comprehensive overview of how to install and start using the `gopdflib` package from the [gopdfsuit](https://github.com/chinmay-sawant/gopdfsuit) repository.

## Table of Contents

1.  [Downloading and Installing](#downloading-and-installing)
2.  [Loading PDF Templates from JSON](#loading-pdf-templates-from-json)

---

## Downloading and Installing

This section explains how to download and install the `gopdflib` package using the latest version tag.

### Prerequisites

- Go (version 1.21 or later is recommended)

### Steps to Download

1.  **Identify the Latest Tag**
    The latest release tag for the repository is **v4.2.0**.

2.  **Get the Package**
    Run the following command in your terminal to download the package at the specific version:

    ```bash
    go get github.com/chinmay-sawant/gopdfsuit@v4.2.0
    ```

    This command will download the source code and add the dependency to your `go.mod` file.

### Usage in Your Project

To use the library in your Go code, import the `gopdflib` package:

```go
import (
    "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
)
```

#### Basic Configuration Example

Here is a simple example of how to reference the library in your code:

```go
package main

import (
    "fmt"
    "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
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

### Updating the Library

To update to the latest version in the future, simply run:

```bash
go get -u github.com/chinmay-sawant/gopdfsuit
```

---

## Loading PDF Templates from JSON

This section explains how to generate PDFs by loading template data from a JSON file. This approach is useful for separating data/content from your Go code, or when receiving template data from an external API.

### Overview

The `gopdflib.PDFTemplate` struct tags match standard JSON naming conventions (camelCase), allowing you to directly unmarshal JSON data into the struct.

### Prerequisites

- `gopdflib` installed (as described above)
- A valid JSON template file (e.g., `sampledata/editor/financial_digitalsignature.json`)

### Example Code

Create a file named `main.go` (or similar) with the following content:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
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

### Running the Sample

We have provided a ready-to-run example in the repository.

1.  Navigate to the project root.
2.  Run the example:

    ```bash
    go run sampledata/gopdflib/load_from_json/main.go
    ```

    This will read `sampledata/editor/financial_digitalsignature.json`, generate the PDF, and save it as `financial_from_json.pdf` in your current directory.

### JSON Structure

The JSON structure directly mirrors the `gopdflib.PDFTemplate` struct. Common top-level fields include:

- `config`: Page settings (size, margin, etc.)
- `title`: Document title section
- `elements`: Array of content elements (tables, spacers, images)
- `footer`: Footer configuration
- `bookmarks`: Navigation outline

Refer to the `sampledata/editor/financial_digitalsignature.json` file for a comprehensive example of the JSON schema.
