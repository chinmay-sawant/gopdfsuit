
export const docSections = [
    {
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
• PDF/A-4 compliance for archival standards
• PDF merging with drag-and-drop UI
• AcroForm/XFDF form filling
• HTML to PDF/Image conversion

Requirements: Go 1.20+, Google Chrome (for HTML conversion)`
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
• /filler - Form Filler
• /htmltopdf - HTML to PDF Converter
• /htmltoimage - HTML to Image Converter`,
                code: {
                    bash: `# Clone & install
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit
go mod download

# Build frontend
cd frontend && npm install && npm run build && cd ..

# Run server
go run ./cmd/gopdfsuit`
                }
            },
            {
                id: 'gopdflib-install',
                title: 'Install gopdflib Package',
                description: 'Use gopdflib as a standalone Go library in your own projects.',
                content: `The gopdflib package allows you to generate PDFs programmatically without running the web server.

Import the package in your Go code to access all PDF generation features.`,
                code: {
                    bash: `go get github.com/chinmay-sawant/gopdfsuit@v4.0.0`,
                    go: `package main

import (
    "fmt"
    "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
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

    "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
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
            }
        ]
    },
    {
        title: 'API Reference',
        items: [
            {
                id: 'generate-pdf',
                title: 'Generate PDF',
                method: 'POST',
                endpoint: '/api/v1/generate/template-pdf',
                description: 'Generate a PDF document from a JSON template. Returns the PDF as binary data.',
                params: [
                    { name: 'config', type: 'object', required: true, description: 'Page configuration (size, orientation, security, signature)' },
                    { name: 'title', type: 'object', required: true, description: 'Document title section' },
                    { name: 'elements', type: 'array', required: false, description: 'Ordered array of tables, spacers, images' },
                    { name: 'table', type: 'array', required: false, description: 'Legacy: array of tables' },
                    { name: 'footer', type: 'object', required: true, description: 'Page footer configuration' },
                    { name: 'bookmarks', type: 'array', required: false, description: 'Document outline/navigation bookmarks' }
                ],
                code: {
                    curl: `curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \\
  -H "Content-Type: application/json" \\
  -d '{
    "config": {
      "page": "A4",
      "pageAlignment": 1,
      "pdfTitle": "Report"
    },
    "title": {
      "props": "Helvetica:24:100:center:0:0:0:0",
      "text": "FINANCIAL REPORT",
      "bgcolor": "#154360",
      "textcolor": "#FFFFFF"
    },
    "elements": [
      {
        "type": "table",
        "table": {
          "maxcolumns": 2,
          "columnwidths": [1, 2],
          "rows": [{"row": [
            {"props": "Helvetica:10:100:left:1:1:1:1", "text": "Company:"},
            {"props": "Helvetica:10:000:left:1:1:1:1", "text": "TechCorp Inc."}
          ]}]
        }
      }
    ],
    "footer": {"font": "Helvetica:8:000:center", "text": "Confidential"}
  }' --output document.pdf`,
                    node: `const response = await fetch('http://localhost:8080/api/v1/generate/template-pdf', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    config: { page: 'A4', pageAlignment: 1, pdfTitle: 'Report' },
    title: {
      props: 'Helvetica:24:100:center:0:0:0:0',
      text: 'FINANCIAL REPORT',
      bgcolor: '#154360',
      textcolor: '#FFFFFF'
    },
    elements: [{
      type: 'table',
      table: {
        maxcolumns: 2,
        columnwidths: [1, 2],
        rows: [{row: [
          {props: 'Helvetica:10:100:left:1:1:1:1', text: 'Company:'},
          {props: 'Helvetica:10:000:left:1:1:1:1', text: 'TechCorp Inc.'}
        ]}]
      }
    }],
    footer: {font: 'Helvetica:8:000:center', text: 'Confidential'}
  })
});

const pdfBlob = await response.blob();`
                }
            },
            {
                id: 'merge-pdfs',
                title: 'Merge PDFs',
                method: 'POST',
                endpoint: '/api/v1/merge',
                description: 'Combine multiple PDF files into a single document. Send PDFs as multipart form data.',
                params: [
                    { name: 'pdf', type: 'file', required: true, description: 'PDF files to merge (multiple). Files are merged in order received.' }
                ],
                code: {
                    curl: `curl -X POST "http://localhost:8080/api/v1/merge" \\
  -F "pdf=@file1.pdf" \\
  -F "pdf=@file2.pdf" \\
  -F "pdf=@file3.pdf" \\
  --output merged.pdf`,
                    node: `const formData = new FormData();
formData.append('pdf', file1);
formData.append('pdf', file2);
formData.append('pdf', file3);

const response = await fetch('http://localhost:8080/api/v1/merge', {
  method: 'POST',
  body: formData
});

const mergedPdf = await response.blob();`
                }
            },
            {
                id: 'fill-form',
                title: 'Fill PDF Form',
                method: 'POST',
                endpoint: '/api/v1/fill',
                description: 'Fill AcroForm fields in a PDF using XFDF data. Returns the filled PDF.',
                params: [
                    { name: 'pdf', type: 'file', required: true, description: 'The PDF file with form fields to fill' },
                    { name: 'xfdf', type: 'file', required: true, description: 'XFDF file containing field values' }
                ],
                code: {
                    curl: `curl -X POST "http://localhost:8080/api/v1/fill" \\
  -F "pdf=@form.pdf" \\
  -F "xfdf=@data.xfdf" \\
  --output filled.pdf`,
                    node: `const formData = new FormData();
formData.append('pdf', pdfFile);
formData.append('xfdf', xfdfFile);

const response = await fetch('http://localhost:8080/api/v1/fill', {
  method: 'POST',
  body: formData
});

const filledPdf = await response.blob();`
                }
            },
            {
                id: 'html-to-pdf',
                title: 'HTML to PDF',
                method: 'POST',
                endpoint: '/api/v1/htmltopdf',
                description: 'Convert HTML content or a URL to PDF using headless Chrome.',
                params: [
                    { name: 'html', type: 'string', required: false, description: 'Raw HTML content to convert' },
                    { name: 'url', type: 'string', required: false, description: 'URL of the page to convert (if html not provided)' },
                    { name: 'page_size', type: 'string', default: 'A4', description: 'Page size: A4, Letter, Legal, A3, A5' },
                    { name: 'orientation', type: 'string', default: 'Portrait', description: 'Portrait or Landscape' }
                ],
                code: {
                    curl: `curl -X POST "http://localhost:8080/api/v1/htmltopdf" \\
  -H "Content-Type: application/json" \\
  -d '{
    "html": "<h1>Hello World</h1><p>This is a PDF.</p>",
    "page_size": "A4",
    "orientation": "Portrait"
  }' --output output.pdf`,
                    node: `const response = await fetch('http://localhost:8080/api/v1/htmltopdf', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    html: '<h1>Hello World</h1><p>This is a PDF.</p>',
    page_size: 'A4',
    orientation: 'Portrait'
  })
});

const pdfBlob = await response.blob();`
                }
            },
            {
                id: 'html-to-image',
                title: 'HTML to Image',
                method: 'POST',
                endpoint: '/api/v1/htmltoimage',
                description: 'Convert HTML content or a URL to PNG, JPG, or SVG using headless Chrome.',
                params: [
                    { name: 'html', type: 'string', required: false, description: 'Raw HTML content to convert' },
                    { name: 'url', type: 'string', required: false, description: 'URL of the page to convert' },
                    { name: 'format', type: 'string', default: 'png', description: 'Output format: png, jpg, svg' },
                    { name: 'width', type: 'int', default: '1920', description: 'Viewport width in pixels' },
                    { name: 'height', type: 'int', default: '1080', description: 'Viewport height in pixels' }
                ],
                code: {
                    curl: `curl -X POST "http://localhost:8080/api/v1/htmltoimage" \\
  -H "Content-Type: application/json" \\
  -d '{
    "url": "https://example.com",
    "format": "png",
    "width": 1920,
    "height": 1080
  }' --output screenshot.png`,
                    node: `const response = await fetch('http://localhost:8080/api/v1/htmltoimage', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    url: 'https://example.com',
    format: 'png',
    width: 1920,
    height: 1080
  })
});

const imageBlob = await response.blob();`
                }
            },
            {
                id: 'get-template',
                title: 'Get Template Data',
                method: 'GET',
                endpoint: '/api/v1/template-data',
                description: 'Retrieve a saved template JSON file from the server.',
                params: [
                    { name: 'file', type: 'string', required: true, description: 'Filename of the template (e.g., temp_multiplepage.json)' }
                ],
                code: {
                    curl: `curl "http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json"`,
                    node: `const response = await fetch('http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json');
const template = await response.json();`
                }
            }
        ]
    },
    {
        title: 'Template Format',
        items: [
            {
                id: 'template-structure',
                title: 'Template Structure',
                description: 'Overview of the JSON template structure used for PDF generation.',
                content: `A PDF template consists of several top-level objects that define the document layout and content.

The recommended approach uses the "elements" array for ordered content. The legacy approach uses separate "table", "spacer", and "image" arrays.`,
                code: {
                    json: `{
  "config": { },      // Page configuration (required)
  "title": { },       // Document title (required)
  "elements": [ ],    // Ordered elements - tables, spacers, images (recommended)
  "table": [ ],       // Legacy: array of tables
  "spacer": [ ],      // Legacy: array of spacers
  "image": [ ],       // Legacy: array of images
  "footer": { },      // Page footer (required)
  "bookmarks": [ ]    // Document outline/navigation (optional)
}`
                }
            },
            {
                id: 'config-object',
                title: 'Config Object',
                description: 'Page layout, appearance, and security settings.',
                params: [
                    { name: 'page', type: 'string', default: 'A4', description: 'Page size: A4, A3, A5, LETTER, LEGAL' },
                    { name: 'pageAlignment', type: 'int', default: '1', description: '1 = Portrait, 2 = Landscape' },
                    { name: 'pageBorder', type: 'string', default: '0:0:0:0', description: 'Border widths: left:right:top:bottom' },
                    { name: 'watermark', type: 'string', required: false, description: 'Diagonal watermark text across all pages' },
                    { name: 'pdfTitle', type: 'string', required: false, description: 'Document title for PDF metadata' },
                    { name: 'pdfaCompliant', type: 'bool', default: 'false', description: 'Enable PDF/A-4 compliance' },
                    { name: 'arlingtonCompatible', type: 'bool', default: 'false', description: 'Enable PDF 2.0 Arlington Model' },
                    { name: 'embedFonts', type: 'bool', default: 'true', description: 'Embed fonts for portability' },
                    { name: 'signature', type: 'object', required: false, description: 'Digital signature settings' },
                    { name: 'security', type: 'object', required: false, description: 'Encryption/password settings' }
                ],
                code: {
                    json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pageBorder": "0:0:0:0",
    "pdfTitle": "Financial Report Q4 2025",
    "pdfaCompliant": true,
    "arlingtonCompatible": true,
    "embedFonts": true,
    "watermark": "CONFIDENTIAL"
  }
}`
                }
            },
            {
                id: 'page-sizes',
                title: 'Page Sizes',
                description: 'Supported page sizes and their dimensions.',
                content: `The following page sizes are available:

| Size   | Dimensions (inches) | Dimensions (points) |
|--------|---------------------|---------------------|
| A3     | 11.69 × 16.54       | 842 × 1191          |
| A4     | 8.27 × 11.69        | 595 × 842           |
| A5     | 5.83 × 8.27         | 420 × 595           |
| LETTER | 8.5 × 11            | 612 × 792           |
| LEGAL  | 8.5 × 14            | 612 × 1008          |

Use pageAlignment: 1 for Portrait or pageAlignment: 2 for Landscape orientation.`
            },
            {
                id: 'props-syntax',
                title: 'Props Syntax',
                description: 'The props string defines text styling and cell borders.',
                content: `Format: "fontname:fontsize:style:alignment:left:right:top:bottom"

Components:
• fontname - Font identifier (Helvetica, Times-Roman, Courier)
• fontsize - Font size in points
• style - 3-digit code for Bold/Italic/Underline
• alignment - Text alignment: left, center, right
• left/right/top/bottom - Border flags (0=none, 1=draw)

Style Codes:
• 000 = Normal
• 100 = Bold
• 010 = Italic
• 001 = Underline
• 110 = Bold + Italic
• 101 = Bold + Underline
• 011 = Italic + Underline
• 111 = Bold + Italic + Underline`,
                code: {
                    json: `// Normal, all borders
"Helvetica:12:000:left:1:1:1:1"

// Bold, centered, top border only
"Helvetica:14:100:center:0:0:1:0"

// Bold+Italic+Underline, right-aligned, no borders
"Helvetica:10:111:right:0:0:0:0"`
                }
            },
            {
                id: 'title-object',
                title: 'Title Object',
                description: 'Defines the document header/title section.',
                params: [
                    { name: 'props', type: 'string', required: true, description: 'Styling properties (see Props Syntax)' },
                    { name: 'text', type: 'string', required: false, description: 'Title text (ignored if table is provided)' },
                    { name: 'table', type: 'object', required: false, description: 'Embedded table for complex layouts' },
                    { name: 'bgcolor', type: 'string', required: false, description: 'Background color (hex: #RRGGBB)' },
                    { name: 'textcolor', type: 'string', required: false, description: 'Text color (hex: #RRGGBB)' },
                    { name: 'link', type: 'string', required: false, description: 'External URL hyperlink' }
                ],
                code: {
                    json: `{
  "title": {
    "props": "Helvetica:24:100:center:0:0:0:0",
    "text": "FINANCIAL REPORT",
    "bgcolor": "#154360",
    "textcolor": "#FFFFFF",
    "link": "https://example.com/report"
  }
}`
                }
            },
            {
                id: 'table-object',
                title: 'Table Object',
                description: 'Tables are the primary content containers in PDF templates.',
                params: [
                    { name: 'maxcolumns', type: 'int', required: true, description: 'Number of columns' },
                    { name: 'rows', type: 'array', required: true, description: 'Array of row objects' },
                    { name: 'columnwidths', type: 'array', required: false, description: 'Relative width weights (e.g., [2,1,1] = 50%, 25%, 25%)' },
                    { name: 'rowheights', type: 'array', required: false, description: 'Row heights in points (default: 25)' },
                    { name: 'bgcolor', type: 'string', required: false, description: 'Default background color for all cells' },
                    { name: 'textcolor', type: 'string', required: false, description: 'Default text color for all cells' }
                ],
                code: {
                    json: `{
  "type": "table",
  "table": {
    "maxcolumns": 4,
    "columnwidths": [1.2, 2, 1.2, 2],
    "rows": [
      {
        "row": [
          {"props": "Helvetica:10:100:left:1:0:1:1", "text": "Company:", "bgcolor": "#F4F6F7"},
          {"props": "Helvetica:10:000:left:0:0:1:1", "text": "TechCorp Inc."},
          {"props": "Helvetica:10:100:left:0:0:1:1", "text": "Period:"},
          {"props": "Helvetica:10:000:left:0:1:1:1", "text": "Q4 2025"}
        ]
      }
    ]
  }
}`
                }
            },
            {
                id: 'cell-object',
                title: 'Cell Object',
                description: 'Individual cells within table rows support text, checkboxes, images, and form fields.',
                params: [
                    { name: 'props', type: 'string', required: true, description: 'Styling properties' },
                    { name: 'text', type: 'string', required: false, description: 'Cell text content' },
                    { name: 'chequebox', type: 'bool', required: false, description: 'Render checkbox (true=checked)' },
                    { name: 'image', type: 'object', required: false, description: 'Embedded image' },
                    { name: 'form_field', type: 'object', required: false, description: 'Interactive form field' },
                    { name: 'bgcolor', type: 'string', required: false, description: 'Cell background color' },
                    { name: 'textcolor', type: 'string', required: false, description: 'Cell text color' },
                    { name: 'link', type: 'string', required: false, description: 'URL or internal #destination' },
                    { name: 'dest', type: 'string', required: false, description: 'Named destination anchor' },
                    { name: 'height', type: 'float', required: false, description: 'Cell height in points' }
                ],
                code: {
                    json: `// Text cell with borders
{"props": "Helvetica:12:000:left:1:1:1:1", "text": "Hello World"}

// Checkbox cell
{"props": "Helvetica:12:000:center:1:1:1:1", "chequebox": true}

// Colored cell
{"props": "Helvetica:12:100:center:1:1:1:1", "text": "Alert!", "bgcolor": "#FF0000", "textcolor": "#FFFFFF"}

// External link
{"props": "Helvetica:10:000:left:1:1:1:1", "text": "Visit Website", "link": "https://example.com", "textcolor": "#0000FF"}

// Internal link (jumps to destination)
{"props": "Helvetica:10:000:left:1:1:1:1", "text": "Go to Summary", "link": "#financial-summary", "textcolor": "#0000FF"}

// Destination anchor
{"props": "Helvetica:12:100:left:1:1:1:1", "text": "FINANCIAL SUMMARY", "dest": "financial-summary", "bgcolor": "#21618C", "textcolor": "#FFFFFF"}`
                }
            },
            {
                id: 'footer-object',
                title: 'Footer Object',
                description: 'Footer appears at the bottom of every page. Page numbers are added automatically.',
                params: [
                    { name: 'font', type: 'string', required: true, description: 'Font props: "fontname:fontsize:style:alignment"' },
                    { name: 'text', type: 'string', required: true, description: 'Footer text' },
                    { name: 'link', type: 'string', required: false, description: 'External URL hyperlink' }
                ],
                code: {
                    json: `{
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "TECHCORP INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
    "link": "https://example.com/legal"
  }
}`
                }
            }
        ]
    },
    {
        title: 'Advanced Features',
        items: [
            {
                id: 'digital-signatures',
                title: 'Digital Signatures',
                description: 'Add legally-binding digital signatures with X.509 certificates.',
                params: [
                    { name: 'enabled', type: 'bool', required: true, description: 'Enable digital signing' },
                    { name: 'visible', type: 'bool', required: false, description: 'Show visible signature stamp' },
                    { name: 'name', type: 'string', required: false, description: 'Signer name (overrides certificate CN)' },
                    { name: 'reason', type: 'string', required: false, description: 'Reason for signing' },
                    { name: 'location', type: 'string', required: false, description: 'Geographic location' },
                    { name: 'contactInfo', type: 'string', required: false, description: 'Contact information' },
                    { name: 'privateKeyPem', type: 'string', required: true, description: 'PEM-encoded private key (RSA or ECDSA)' },
                    { name: 'certificatePem', type: 'string', required: true, description: 'PEM-encoded X.509 certificate' },
                    { name: 'certificateChain', type: 'array', required: false, description: 'Array of intermediate certificates' }
                ],
                code: {
                    json: `{
  "config": {
    "signature": {
      "enabled": true,
      "visible": true,
      "name": "John Doe",
      "reason": "Document Approval",
      "location": "New York, US",
      "contactInfo": "john@example.com",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----",
      "certificatePem": "-----BEGIN CERTIFICATE-----\\n...\\n-----END CERTIFICATE-----",
      "certificateChain": [
        "-----BEGIN CERTIFICATE-----\\n...intermediate...\\n-----END CERTIFICATE-----",
        "-----BEGIN CERTIFICATE-----\\n...root...\\n-----END CERTIFICATE-----"
      ]
    }
  }
}`
                }
            },
            {
                id: 'pdf-encryption',
                title: 'PDF Encryption',
                description: 'Password-protect documents with granular permission controls.',
                params: [
                    { name: 'enabled', type: 'bool', required: true, description: 'Enable encryption' },
                    { name: 'ownerPassword', type: 'string', required: true, description: 'Password for full document access' },
                    { name: 'userPassword', type: 'string', required: false, description: 'Password to open document (empty = no password to open)' },
                    { name: 'allowPrinting', type: 'bool', required: false, description: 'Allow document printing' },
                    { name: 'allowCopying', type: 'bool', required: false, description: 'Allow copying text/images' },
                    { name: 'allowModifying', type: 'bool', required: false, description: 'Allow content modification' },
                    { name: 'allowAnnotations', type: 'bool', required: false, description: 'Allow adding annotations' },
                    { name: 'allowFormFilling', type: 'bool', required: false, description: 'Allow filling form fields' }
                ],
                code: {
                    json: `{
  "config": {
    "security": {
      "enabled": true,
      "ownerPassword": "admin123",
      "userPassword": "view123",
      "allowPrinting": true,
      "allowCopying": false,
      "allowModifying": false,
      "allowAnnotations": false,
      "allowFormFilling": true
    }
  }
}`
                }
            },
            {
                id: 'bookmarks',
                title: 'Bookmarks & Navigation',
                description: 'Create document outlines with internal navigation links.',
                params: [
                    { name: 'title', type: 'string', required: true, description: 'Display text in bookmark panel' },
                    { name: 'page', type: 'int', required: false, description: 'Target page number (1-based)' },
                    { name: 'dest', type: 'string', required: false, description: 'Named destination (matches cell dest field)' },
                    { name: 'y', type: 'float', required: false, description: 'Y position on target page' },
                    { name: 'children', type: 'array', required: false, description: 'Nested bookmarks for hierarchy' },
                    { name: 'open', type: 'bool', required: false, description: 'Whether children are expanded' }
                ],
                content: `To create internal navigation:

1. Add a destination anchor to a cell using the "dest" field
2. Create a link to that destination using "link": "#destination-name"
3. Optionally add a bookmark for sidebar navigation`,
                code: {
                    json: `{
  "bookmarks": [
    {
      "title": "Financial Report",
      "page": 1,
      "children": [
        {"title": "Company Information", "page": 1},
        {"title": "Financial Summary", "dest": "financial-summary"}
      ]
    },
    {
      "title": "Charts & Visualizations",
      "page": 2,
      "dest": "charts-section"
    }
  ]
}`
                }
            },
            {
                id: 'form-fields',
                title: 'Interactive Form Fields',
                description: 'Add fillable form fields to create interactive PDFs.',
                params: [
                    { name: 'type', type: 'string', required: true, description: 'checkbox, radio, or text' },
                    { name: 'name', type: 'string', required: true, description: 'Field name for data extraction' },
                    { name: 'value', type: 'string', required: false, description: 'Export value or default text' },
                    { name: 'checked', type: 'bool', required: false, description: 'Initial checked state' },
                    { name: 'group_name', type: 'string', required: false, description: 'Radio button group name' },
                    { name: 'shape', type: 'string', required: false, description: 'round or square (for radio buttons)' }
                ],
                code: {
                    json: `// Text input field
{
  "props": "Helvetica:9:000:left:1:1:1:1",
  "text": "John Doe",
  "form_field": {
    "type": "text",
    "name": "customer_name",
    "value": "John Doe"
  }
}

// Checkbox
{
  "props": "Helvetica:8:000:center:1:1:1:1",
  "form_field": {
    "type": "checkbox",
    "name": "agree_terms",
    "value": "Yes",
    "checked": true
  }
}

// Radio button group
{
  "props": "Helvetica:9:000:center:1:1:1:1",
  "form_field": {
    "type": "radio",
    "name": "payment_method",
    "value": "credit_card",
    "checked": true,
    "group_name": "payment_type",
    "shape": "round"
  }
}`
                }
            },
            {
                id: 'images',
                title: 'Embedded Images',
                description: 'Embed images in cells using base64-encoded data.',
                params: [
                    { name: 'imagename', type: 'string', required: false, description: 'Image identifier' },
                    { name: 'imagedata', type: 'string', required: true, description: 'Base64-encoded image data' },
                    { name: 'width', type: 'float', required: true, description: 'Image width in points' },
                    { name: 'height', type: 'float', required: true, description: 'Image height in points' }
                ],
                code: {
                    json: `{
  "props": "Helvetica:12:000:center:0:0:0:0",
  "height": 150,
  "link": "https://example.com/charts",
  "image": {
    "imagename": "chart",
    "imagedata": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "width": 150,
    "height": 150
  }
}`
                }
            },
            {
                id: 'pdfa-compliance',
                title: 'PDF/A-4 Compliance',
                description: 'Generate archival-standard PDFs for long-term preservation.',
                content: `PDF/A-4 is the archival standard based on PDF 2.0. Enable it with "pdfaCompliant": true.

This automatically:
• Embeds all fonts (via Liberation fonts)
• Adds required XMP metadata
• Follows strict structure requirements
• Ensures long-term document accessibility

Combine with "arlingtonCompatible": true for full PDF 2.0 Arlington Model compliance.`,
                code: {
                    json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pdfTitle": "Archival Document",
    "pdfaCompliant": true,
    "arlingtonCompatible": true,
    "embedFonts": true
  }
}`
                }
            }
        ]
    },
    {
        title: 'Examples',
        items: [
            {
                id: 'example-financial-report',
                title: 'Financial Report',
                description: 'Complete financial report with sections, tables, and styling.',
                content: `This example demonstrates a professional financial report with:
• Colored section headers
• Multi-column data tables
• Internal navigation links
• Bookmarks for quick navigation
• Footer with company information`,
                code: {
                    json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pdfTitle": "Financial Report Q4 2025",
    "pdfaCompliant": true,
    "embedFonts": true
  },
  "title": {
    "props": "Helvetica:24:100:center:0:0:0:0",
    "text": "FINANCIAL REPORT",
    "bgcolor": "#154360",
    "textcolor": "#FFFFFF"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:12:100:left:1:1:1:1", "text": "COMPANY INFORMATION", "bgcolor": "#21618C", "textcolor": "#FFFFFF"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 4,
        "columnwidths": [1.2, 2, 1.2, 2],
        "rows": [
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:1", "text": "Company:", "bgcolor": "#F4F6F7"},
            {"props": "Helvetica:10:000:left:0:0:0:1", "text": "TechCorp Inc.", "link": "https://techcorp.example.com"},
            {"props": "Helvetica:10:100:left:0:0:0:1", "text": "Period:"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "Q4 2025"}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:12:100:left:1:1:1:1", "text": "FINANCIAL SUMMARY", "bgcolor": "#21618C", "textcolor": "#FFFFFF", "dest": "financial-summary"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "columnwidths": [2, 1],
        "rows": [
          {"row": [
            {"props": "Helvetica:10:000:left:1:0:0:1", "text": "Total Revenue"},
            {"props": "Helvetica:10:000:right:0:1:0:1", "text": "$2,450,000"}
          ]},
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:1", "text": "Gross Profit", "bgcolor": "#D4E6F1"},
            {"props": "Helvetica:10:100:right:0:1:0:1", "text": "$1,225,000", "bgcolor": "#D4E6F1"}
          ]},
          {"row": [
            {"props": "Helvetica:11:100:left:1:0:1:1", "text": "Net Income", "bgcolor": "#A9CCE3"},
            {"props": "Helvetica:11:100:right:0:1:1:1", "text": "$125,000", "bgcolor": "#A9CCE3"}
          ]}
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "TECHCORP INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL"
  },
  "bookmarks": [
    {"title": "Financial Summary", "dest": "financial-summary"}
  ]
}`
                }
            },
            {
                id: 'example-legal-contract',
                title: 'Legal Contract',
                description: 'Professional services agreement with watermark and signature blocks.',
                content: `This example demonstrates a legal contract with:
• CONFIDENTIAL watermark
• Dual-column party information
• Numbered sections with styling
• Signature blocks`,
                code: {
                    json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "watermark": "CONFIDENTIAL",
    "pdfaCompliant": true,
    "embedFonts": true
  },
  "title": {
    "props": "Helvetica:18:100:center:0:0:0:0",
    "text": "PROFESSIONAL SERVICES AGREEMENT",
    "bgcolor": "#2C3E50",
    "textcolor": "#FFFFFF"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "bgcolor": "#ECF0F1",
        "rows": [
          {"row": [
            {"props": "Helvetica:12:100:center:1:1:1:1", "text": "CLIENT", "bgcolor": "#34495E", "textcolor": "#FFFFFF"},
            {"props": "Helvetica:12:100:center:1:1:1:1", "text": "PROVIDER", "bgcolor": "#34495E", "textcolor": "#FFFFFF"}
          ]},
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:0", "text": "Global Tech Solutions Ltd."},
            {"props": "Helvetica:10:100:left:0:1:0:0", "text": "Apex Legal Consultants"}
          ]},
          {"row": [
            {"props": "Helvetica:10:000:left:1:0:0:1", "text": "Silicon Valley, CA 94025"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "New York, NY 10018"}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:14:100:left:0:0:0:1", "text": "1. SERVICES AND TERM", "textcolor": "#2980B9"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "columnwidths": [1, 3],
        "rows": [
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:1:1", "text": "Project:", "bgcolor": "#EBF5FB"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "Enterprise Software Licensing Review"}
          ]},
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:1", "text": "Start Date:", "bgcolor": "#EBF5FB"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "November 1, 2023"}
          ]}
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "Service Agreement - Confidential"
  }
}`
                }
            },
            {
                id: 'example-form',
                title: 'Interactive Form',
                description: 'Library form with text inputs, checkboxes, and radio buttons.',
                content: `This example demonstrates an interactive form with:
• Text input fields with default values
• Radio button groups for selections
• Checkboxes for multi-select options
• Color-coded sections`,
                code: {
                    json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pageBorder": "1:1:1:1",
    "pdfaCompliant": true,
    "embedFonts": true
  },
  "title": {
    "props": "Helvetica:18:100:center:0:0:0:1",
    "text": "LIBRARY BOOK RECEIVING FORM"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:11:100:left:1:1:1:1", "text": "LIBRARY INFORMATION", "bgcolor": "#E8F4FD", "textcolor": "#1565C0"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 4,
        "columnwidths": [1.2, 2, 1.2, 2],
        "rows": [
          {"row": [
            {"props": "Helvetica:9:100:left:1:1:1:1", "text": "Library Name:"},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Central City Public Library",
              "form_field": {"type": "text", "name": "library_name", "value": "Central City Public Library"}},
            {"props": "Helvetica:9:100:left:1:1:1:1", "text": "Branch Code:"},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "CCPL-MAIN-001",
              "form_field": {"type": "text", "name": "branch_code", "value": "CCPL-MAIN-001"}}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:11:100:left:1:1:1:1", "text": "ACQUISITION TYPE", "bgcolor": "#E8F4FD", "textcolor": "#1565C0"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 6,
        "columnwidths": [1.2, 0.4, 1.2, 0.4, 1.2, 0.4],
        "rows": [
          {"row": [
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Purchase"},
            {"props": "Helvetica:9:000:center:1:1:1:1",
              "form_field": {"type": "radio", "name": "acq_purchase", "value": "purchase", "checked": true, "group_name": "acquisition_type", "shape": "round"}},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Donation"},
            {"props": "Helvetica:9:000:center:1:1:1:1",
              "form_field": {"type": "radio", "name": "acq_donation", "value": "donation", "checked": false, "group_name": "acquisition_type", "shape": "round"}},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Exchange"},
            {"props": "Helvetica:9:000:center:1:1:1:1",
              "form_field": {"type": "radio", "name": "acq_exchange", "value": "exchange", "checked": false, "group_name": "acquisition_type", "shape": "round"}}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:11:100:left:1:1:1:1", "text": "INSPECTION CHECKLIST", "bgcolor": "#E8F4FD", "textcolor": "#1565C0"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 6,
        "columnwidths": [0.3, 1.5, 0.3, 1.5, 0.3, 1.5],
        "rows": [
          {"row": [
            {"props": "Helvetica:8:000:center:1:1:1:1",
              "form_field": {"type": "checkbox", "name": "check_cover", "value": "Yes", "checked": true}},
            {"props": "Helvetica:8:000:left:1:1:1:1", "text": "Cover Intact"},
            {"props": "Helvetica:8:000:center:1:1:1:1",
              "form_field": {"type": "checkbox", "name": "check_spine", "value": "Yes", "checked": true}},
            {"props": "Helvetica:8:000:left:1:1:1:1", "text": "Spine OK"},
            {"props": "Helvetica:8:000:center:1:1:1:1",
              "form_field": {"type": "checkbox", "name": "check_pages", "value": "Yes", "checked": true}},
            {"props": "Helvetica:8:000:left:1:1:1:1", "text": "All Pages"}
          ]}
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:7:000:center",
    "text": "CENTRAL CITY PUBLIC LIBRARY - BOOK RECEIVING DEPARTMENT"
  }
}`
                }
            }
        ]
    }
]
