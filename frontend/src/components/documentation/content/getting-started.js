export const gettingStartedSection = {
    title: 'Getting Started',
    items: [
        {
            id: 'introduction',
            title: 'Introduction',
            description: 'GoPdfSuit is a powerful Go web service for template-based PDF generation with multi-page support, PDF merging, form filling, and HTML to PDF/Image conversion.',
            content: `GoPdfSuit provides a complete solution for generating professional PDF documents from JSON templates. Key capabilities include:

• Template-based PDF generation with auto page breaks
• Digital signatures (PKCS#7) with X.509 certificate chains
• PDF encryption with password protection & permissions
• Bookmarks, internal links, and named destinations
• PDF/A-4 and PDF/UA-2 compliance for archival standards
• PDF merging with drag-and-drop UI
• AcroForm/XFDF form filling
• HTML to PDF/Image conversion

**Python Support**: 
• **Native Python Bindings**: Direct CGO integration via [pypdfsuit](https://github.com/chinmay-sawant/gopdfsuit/tree/master/bindings/python).
• **Python Web Client**: Lightweight REST API client available [here](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/python/gopdf).

Requirements: Go 1.24+, Google Chrome (for HTML conversion)`,
            features: [
                { title: 'Template Generation', description: 'JSON-based PDF generation with auto page breaks', icon: 'FileText' },
                { title: 'Digital Signatures', description: 'PKCS#7 signing with X.509 certificate chains', icon: 'PenTool' },
                { title: 'PDF Encryption', description: 'AES-256 encryption with granular permissions', icon: 'Lock' },
                { title: 'Smart Navigation', description: 'Bookmarks, internal links, and named destinations', icon: 'BookOpen' },
                { title: 'Archival Standards', description: 'PDF/A-4 & PDF/UA-2 compliance for long-term preservation', icon: 'Archive' },
                { title: 'PDF Merging', description: 'Combine multiple documents with drag-and-drop', icon: 'Layers' },
                { title: 'Form Filling', description: 'Populate AcroForms and XFDF data automatically', icon: 'Edit3' },
                { title: 'HTML Conversion', description: 'High-fidelity HTML to PDF and Image rendering', icon: 'Globe' }
            ]
        },
        {
            id: 'quick-start',
            title: 'Quick Start',
            description: 'Get GoPdfSuit running locally in minutes.',
            content: `Clone the repository, install dependencies, build the frontend, and start the server.

Access the application at http://localhost:8080

Web interfaces available:
• / - PDF Viewer & Generator
• /editor - Drag-and-drop Template Builder
• /merge - PDF Merger
• /split - PDF Splitter
• /filler - Form Filler
• /htmltopdf - HTML to PDF Converter
• /htmltoimage - HTML to Image Converter`,
            code: {
                bash: `# Clone & install
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit
go mod download

# Run server
make run`
            }
        },
        {
            id: 'gopdflib-install',
            title: 'Install gopdflib Package',
            description: 'Use gopdflib as a standalone Go library in your own projects.',
            content: `The [gopdflib](https://github.com/chinmay-sawant/gopdfsuit/tree/master/pkg/gopdflib) package allows you to generate PDFs programmatically without running the web server.
View detailed sample data and examples [here](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/gopdflib).

Import the package in your Go code to access all PDF generation features.`,
            code: {
                bash: `go get github.com/chinmay-sawant/gopdfsuit/v4@latest`,
                go: `package main

import (
    "fmt"
    "github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
    config := gopdflib.Config{
        Page:          "A4",
        PageAlignment: 1, // Portrait
    }
    
    fmt.Printf("Config: %+v\\n", config)
}`
            }
        },
        {
            id: 'load-json-template',
            title: 'Load JSON Templates',
            description: 'Generate PDFs by loading template data from JSON files.',
            content: `The gopdflib.PDFTemplate struct tags match standard JSON naming conventions (camelCase), allowing you to directly unmarshal JSON data into the struct.

This approach is useful for separating data/content from your Go code, or when receiving template data from an external API.`,
            code: {
                go: `package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
    // Read JSON file
    jsonData, err := os.ReadFile("template.json")
    if err != nil {
        panic(err)
    }

    // Unmarshal into PDFTemplate
    var template gopdflib.PDFTemplate
    err = json.Unmarshal(jsonData, &template)
    if err != nil {
        panic(err)
    }

    // Generate PDF
    pdfBytes, err := gopdflib.GeneratePDF(template)
    if err != nil {
        panic(err)
    }

    // Save to file
    os.WriteFile("output.pdf", pdfBytes, 0644)
    fmt.Println("PDF generated!")
}`
            }
        },
        {
            id: 'gopdflib-financial-report',
            title: 'gopdflib Financial Report',
            description: 'Complete example generating a financial report with benchmarking.',
            content: `This example demonstrates building a complete financial report programmatically using gopdflib. Features include:

• Structured template with Config, Title, Elements, Footer
• Digital signature configuration with X.509 certificate chain
• PDF/A-4 compliance settings
• Performance benchmarking with multiple iterations
• Helper functions for pointer values`,
            code: {
                go: `package main

import (
    "fmt"
    "os"
    "time"
    "github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
    fmt.Println("=== gopdflib Financial Report Example ===")

    // Build the template
    template := buildFinancialReportTemplate()

    // Warm-up run
    fmt.Println("Warm-up run...")
    _, err := gopdflib.GeneratePDF(template)
    if err != nil {
        fmt.Printf("Error: %v\\n", err)
        os.Exit(1)
    }

    // Benchmark multiple iterations
    iterations := 10
    var totalDuration time.Duration

    for i := 1; i <= iterations; i++ {
        start := time.Now()
        pdfBytes, err := gopdflib.GeneratePDF(template)
        elapsed := time.Since(start)

        if err != nil {
            fmt.Printf("Iteration %d: ERROR - %v\\n", i, err)
            continue
        }

        totalDuration += elapsed
        fmt.Printf("Iteration %2d: %8.3f ms (%d bytes)\\n", 
            i, float64(elapsed.Microseconds())/1000.0, len(pdfBytes))

        if i == iterations {
            os.WriteFile("financial_report.pdf", pdfBytes, 0644)
        }
    }

    avgDuration := totalDuration / time.Duration(iterations)
    fmt.Printf("\\n=== Average: %.3f ms ===\\n", 
        float64(avgDuration.Microseconds())/1000.0)
}

func buildFinancialReportTemplate() gopdflib.PDFTemplate {
    floatPtr := func(f float64) *float64 { return &f }
    boolPtr := func(b bool) *bool { return &b }

    return gopdflib.PDFTemplate{
        Config: gopdflib.Config{
            Page:                "A4",
            PageAlignment:       1,
            PdfTitle:            "Financial Report Q4 2025",
            PDFACompliant:       true,
            ArlingtonCompatible: true,
            EmbedFonts:          boolPtr(true),
        },
        Title: gopdflib.Title{
            Props:     "Helvetica:24:100:center:0:0:0:0",
            Text:      "FINANCIAL REPORT",
            BgColor:   "#154360",
            TextColor: "#FFFFFF",
        },
        Elements: []gopdflib.Element{
            {
                Type: "table",
                Table: &gopdflib.Table{
                    MaxColumns:   2,
                    ColumnWidths: []float64{2, 1},
                    Rows: []gopdflib.Row{
                        {Row: []gopdflib.Cell{
                            {Props: "Helvetica:10:000:left:1:0:0:1", Text: "Total Revenue"},
                            {Props: "Helvetica:10:000:right:0:1:0:1", Text: "$2,450,000"},
                        }},
                        {Row: []gopdflib.Cell{
                            {Props: "Helvetica:11:100:left:1:0:1:1", Text: "Net Income", BgColor: "#A9CCE3"},
                            {Props: "Helvetica:11:100:right:0:1:1:1", Text: "$125,000", BgColor: "#A9CCE3"},
                        }},
                    },
                },
            },
        },
        Footer: gopdflib.Footer{
            Font: "Helvetica:8:000:center",
            Text: "FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
        },
    }
}`
            }
        },
        {
            id: 'text-wrapping',
            title: 'Text Wrapping',
            description: 'Auto text wrapping with dynamic row heights for long content.',
            content: `GoPdfSuit automatically wraps text within cells and adjusts row heights dynamically. Features include:

• Auto text wrapping enabled by default
• Dynamic row height based on wrapped content
• Ability to disable wrap for specific cells using wrap: false
• All cells in a row stretch to match the tallest cell
• Works with all text alignments (left, center, right)`,
            code: {
                go: `package main

import (
    "os"
    "github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
    boolPtr := func(b bool) *bool { return &b }

    template := gopdflib.PDFTemplate{
        Config: gopdflib.Config{
            Page:          "A4",
            PageAlignment: 1,
            PdfTitle:      "Text Wrapping Demo",
        },
        Title: gopdflib.Title{
            Props: "Helvetica:18:100:center:0:0:0:0",
            Text:  "Text Wrapping Feature Demo",
        },
        Elements: []gopdflib.Element{
            {
                Type: "table",
                Table: &gopdflib.Table{
                    MaxColumns:   3,
                    ColumnWidths: []float64{1, 2, 1},
                    Rows: []gopdflib.Row{
                        // Header row
                        {Row: []gopdflib.Cell{
                            {Props: "Helvetica:10:100:center:1:1:1:1", Text: "Column A", BgColor: "#D5DBDB"},
                            {Props: "Helvetica:10:100:center:1:1:1:1", Text: "Long Text Column", BgColor: "#D5DBDB"},
                            {Props: "Helvetica:10:100:center:1:1:1:1", Text: "Column C", BgColor: "#D5DBDB"},
                        }},
                        // Data row with auto-wrap (default)
                        {Row: []gopdflib.Cell{
                            {Props: "Helvetica:10:000:left:1:1:0:1", Text: "Short"},
                            {Props: "Helvetica:10:000:left:1:1:0:1", Text: "This is a very long piece of text that demonstrates the automatic text wrapping feature. The cell height will automatically adjust to fit all the content."},
                            {Props: "Helvetica:10:000:left:1:1:0:1", Text: "Also short"},
                        }},
                        // Row with wrap disabled for middle cell
                        {Row: []gopdflib.Cell{
                            {Props: "Helvetica:10:000:left:1:1:0:1", Text: "Wrap ON"},
                            {Props: "Helvetica:10:000:left:1:1:0:1", Text: "This text has wrap disabled - will be clipped.", Wrap: boolPtr(false)},
                            {Props: "Helvetica:10:000:left:1:1:0:1", Text: "Wrap ON"},
                        }},
                    },
                },
            },
        },
        Footer: gopdflib.Footer{
            Font: "Helvetica:8:000:center",
            Text: "Text Wrapping Demo",
        },
    }

    pdfBytes, _ := gopdflib.GeneratePDF(template)
    os.WriteFile("text_wrapping.pdf", pdfBytes, 0644)
}`,
                json: `{
  "type": "table",
  "table": {
    "maxcolumns": 2,
    "columnwidths": [1, 2],
    "rows": [
      {
        "row": [
          {"props": "Helvetica:10:000:left:1:1:1:1", "text": "Short text"},
          {"props": "Helvetica:10:000:left:1:1:1:1", "text": "This is a very long piece of text that will automatically wrap to multiple lines. The row height adjusts dynamically to fit all content.", "wrap": true}
        ]
      },
      {
        "row": [
          {"props": "Helvetica:10:000:left:1:1:1:1", "text": "Normal"},
          {"props": "Helvetica:10:000:left:1:1:1:1", "text": "Text with wrap disabled will be clipped if too long for the cell.", "wrap": false}
        ]
      }
    ]
  }
}`
            }
        },
        {
            id: 'svg-vector-support',
            title: 'SVG Vector Support',
            description: 'Embed SVG vector graphics in PDF documents with native rendering.',
            content: `GoPdfSuit supports embedding SVG vector graphics directly into PDF documents. The SVG content is converted to native PDF drawing commands for crisp rendering at any scale.

Supported SVG elements:
• rect - Rectangles with optional rounded corners
• circle - Circles
• ellipse - Ellipses
• line - Lines
• polyline - Connected lines
• polygon - Closed polygons
• path - Complex paths with M, L, H, V, C, S, Q, T, A, Z commands
• text - Text elements

SVG is ideal for charts, diagrams, logos, and mathematical formulas.`,
            code: {
                json: `{
  "props": "Helvetica:12:000:center:0:0:0:0",
  "height": 200,
  "image": {
    "imagename": "chart",
    "imagedata": "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyMDAiIGhlaWdodD0iMjAwIj4KICA8cmVjdCB4PSIxMCIgeT0iMTAiIHdpZHRoPSI4MCIgaGVpZ2h0PSIxODAiIGZpbGw9IiMzNDk4ZGIiLz4KICA8cmVjdCB4PSIxMTAiIHk9IjUwIiB3aWR0aD0iODAiIGhlaWdodD0iMTQwIiBmaWxsPSIjZTc0YzNjIi8+Cjwvc3ZnPg==",
    "width": 200,
    "height": 200
  }
}`
            }
        },
        {
            id: 'python-support',
            title: 'Python Client',
            description: 'Generate PDFs from Python using the GoPdfSuit API client.',
            content: `The Python client provides a simple interface to generate PDFs from JSON templates. Features include:

• Template placeholder filling with {key} syntax
• Pydantic models for request validation
• Support for loading templates from files
• Template structure transformation`,
            code: {
                python: `# Install requirements
# pip install requests pydantic

from gopdf import PdfClient
import json
import re

def fill_template(template, data):
    """Replace {key} placeholders with data values."""
    if isinstance(template, dict):
        return {k: fill_template(v, data) for k, v in template.items()}
    elif isinstance(template, list):
        return [fill_template(i, data) for i in template]
    elif isinstance(template, str):
        def replace_match(match):
            key = match.group(1)
            return str(data.get(key, match.group(0)))
        return re.sub(r'\\{(\\w+)\\}', replace_match, template)
    return template

def main():
    # Define template data
    user_input = {
        "company_name": "TechCorp Industries Inc.",
        "report_period": "Q4 2025",
        "total_revenue": "$2,450,000",
        "net_income": "$125,000",
    }

    # Load and fill template
    with open("financial_template.json", "r") as f:
        template_data = json.load(f)
    
    filled_data = fill_template(template_data, user_input)

    # Generate PDF
    client = PdfClient()
    pdf_content = client.generate_pdf(filled_data)
    
    if pdf_content:
        with open("financial_report.pdf", "wb") as f:
            f.write(pdf_content)
        print("PDF generated successfully!")

if __name__ == "__main__":
    main()`,
                bash: `# Python client setup
pip install requests pydantic

# Run the example
python main.py`
            }
        },
        {
            id: 'gopdflib-redaction',
            title: 'gopdflib Redaction',
            description: 'Redact sensitive information from PDFs programmatically.',
            content: `Use the gopdflib \`ApplyRedactionsAdvanced\` function to scrub text from an existing PDF. You can specify exact coordinates or search for text like "Confidential" to automatically locate and redact it. 
            
The \`secure_required\` mode can be used to ensure text is permanently removed from the document stream, while \`visual_allowed\` will attempt to visually redact text by painting over it.`,
            code: {
                go: `package main

import (
    "fmt"
    "os"

    "github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
    pdfBytes, err := os.ReadFile("document.pdf")
    if err != nil {
        panic(err)
    }

    options := gopdflib.ApplyRedactionOptions{
        TextSearch: []gopdflib.RedactionTextQuery{
            {Text: "Jeffrey epstein"},
            {Text: "Confidential"},
        },
        // We can also supply explicit regions:
        // Blocks: []gopdflib.RedactionRect{ ... },
        Mode: "secure_required",
    }
    
    redactedBytes, report, err := gopdflib.ApplyRedactionsAdvancedWithReport(pdfBytes, options)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Generated %v redactions\\n", report.GeneratedRects)
    os.WriteFile("redacted.pdf", redactedBytes, 0644)
}`
            }
        }
    ]
};

