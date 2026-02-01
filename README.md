# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![gochromedp](https://img.shields.io/badge/gochromedp-1.0+-00ADD8?style=flat)](https://github.com/chinmay-sawant/gochromedp)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/chinmay-sawant/gopdfsuit)

> 🚀 A powerful Go web service for template-based PDF generation with multi-page support, PDF merging, form filling, and HTML to PDF/Image conversion.

## Star History
[![Star History Chart](https://api.star-history.com/svg?repos=chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)](https://www.star-history.com/#chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)
---

## 📑 Table of Contents

- [Overview](#-overview)
- [Quick Start](#-quick-start)
- [Docker Deployment](#-docker-deployment)
- [Web Interfaces](#-web-interfaces)
- [API Reference](#-api-reference)
- [Template Format](#-template-format)
- [Features](#-features)
- [FAQ](#-faq)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)

### 📚 Documentation

- [📋 Template Reference](guides/TEMPLATE_REFERENCE.md) - Complete JSON template format guide with examples
- [🛠️ Makefile Reference](guides/MAKEFILE.md) - Build, test, and deployment commands

---

## 📖 Overview

GoPdfSuit is a Go + Gin web service that generates professional PDF documents from JSON templates. 

**Key Capabilities:**
- 📄 Template-based PDF generation with auto page breaks
- 🔐 Digital signatures (PKCS#7) with X.509 certificate chains
- 🔒 PDF encryption with password protection & permissions
- 📑 Bookmarks, internal links, and named destinations
- ✅ PDF/A-4 compliance for archival standards
- 🔗 PDF merging with drag-and-drop UI
- 🖊️ AcroForm/XFDF form filling
- 🌐 HTML to PDF/Image conversion (via gochromedp)
- 🎨 Font styling (bold, italic, underline), tables, checkboxes, radio buttons
- 📏 Multiple page sizes (A3, A4, A5, Letter, Legal) & orientations

**Requirements:** Go 1.20+, Google Chrome (for HTML conversion)

---

## ⚡ Quick Start

```bash
# 1. Clone & install
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit
go mod download

# 2. Build frontend
cd frontend && npm install && npm run build && cd ..

# 3. Run server
go run ./cmd/gopdfsuit
```

**Access:** `http://localhost:8080`

<details>
<summary><b>📦 Install Chrome (required for HTML conversion)</b></summary>

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y google-chrome-stable

# macOS
brew install --cask google-chrome

# Verify
google-chrome --version
```
</details>

<details>
<summary><b>🔍 Find your IP (recommended over localhost)</b></summary>

```bash
# Linux/WSL
hostname -I

# Windows
ipconfig | findstr IPv4
```
</details>

---

## 🐳 Docker Deployment

```bash
# Option 1: Build locally
make docker

# Option 2: Pull from Docker Hub
docker pull chinmaysawant/gopdfsuit:latest
docker run -d -p 8080:8080 chinmaysawant/gopdfsuit:latest
```

<details>
<summary><b>📋 Makefile Targets</b></summary>

> 📖 **Full documentation:** [Makefile Reference Guide](guides/MAKEFILE.md)

| Command | Description |
|---------|-------------|
| `make docker` | Build and run Docker container |
| `make dockertag` | Tag and push to Docker Hub |
| `make build` | Build Go application |
| `make test` | Run tests |
| `make run` | Run locally |
| `make pull` | Pull and run from Docker Hub |
</details>

---

## 🖥️ Web Interfaces

| Interface | URL | Description |
|-----------|-----|-------------|
| **PDF Viewer** | `/` | View and generate PDFs from templates |
| **Template Editor** | `/editor` | Drag-and-drop PDF template builder |
| **PDF Merger** | `/merge` | Combine multiple PDFs |
| **Form Filler** | `/filler` | Fill PDF forms with XFDF data |
| **HTML to PDF** | `/htmltopdf` | Convert HTML/URLs to PDF |
| **HTML to Image** | `/htmltoimage` | Convert HTML/URLs to PNG/JPG/SVG |

---

## 📡 API Reference

### PDF Generation

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/generate/template-pdf` | POST | Generate PDF from JSON template |
| `/api/v1/template-data?file=` | GET | Get template JSON data |
| `/api/v1/merge` | POST | Merge multiple PDFs (multipart) |
| `/api/v1/fill` | POST | Fill PDF forms (PDF + XFDF) |
| `/api/v1/htmltopdf` | POST | Convert HTML/URL to PDF |
| `/api/v1/htmltoimage` | POST | Convert HTML/URL to image |

<details>
<summary><b>📄 Generate PDF Example</b></summary>

```bash
curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "page": "A4",
      "pageAlignment": 1,
      "pageBorder": "0:0:0:0",
      "pdfTitle": "Financial Report",
      "pdfaCompliant": true
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
          "rows": [
            {"row": [
              {"props": "Helvetica:10:100:left:1:1:1:1", "text": "Company:"},
              {"props": "Helvetica:10:000:left:1:1:1:1", "text": "TechCorp Inc.", "link": "https://example.com"}
            ]},
            {"row": [
              {"props": "Helvetica:10:100:left:1:1:1:1", "text": "Summary:", "dest": "summary"},
              {"props": "Helvetica:10:000:left:1:1:1:1", "text": "Q4 2025"}
            ]}
          ]
        }
      }
    ],
    "footer": {"font": "Helvetica:8:000:center", "text": "Confidential"},
    "bookmarks": [{"title": "Summary", "dest": "summary"}]
  }' --output document.pdf
```
</details>

<details>
<summary><b>🔗 Merge PDFs Example</b></summary>

```bash
curl -X POST "http://localhost:8080/api/v1/merge" \
  -F "pdf=@file1.pdf" -F "pdf=@file2.pdf" --output merged.pdf
```
</details>

<details>
<summary><b>🌐 HTML to PDF Example</b></summary>

```bash
curl -X POST "http://localhost:8080/api/v1/htmltopdf" \
  -H "Content-Type: application/json" \
  -d '{"html": "<h1>Hello World</h1>", "page_size": "A4", "orientation": "Portrait"}' \
  --output output.pdf
```
</details>

<details>
<summary><b>🖊️ Form Fill Example</b></summary>

```bash
curl -X POST "http://localhost:8080/api/v1/fill" \
  -F "pdf=@form.pdf" -F "xfdf=@data.xfdf" --output filled.pdf
```
</details>

---

## 📋 Template Format

> 📖 **Full documentation:** [Template Reference Guide](guides/TEMPLATE_REFERENCE.md)

### Props Syntax
`"fontname:fontsize:style:alignment:left:right:top:bottom"`

| Component | Values |
|-----------|--------|
| **fontname** | Helvetica, Times-Roman, Courier, etc. |
| **fontsize** | Size in points |
| **style** | 3-digit code: `[bold][italic][underline]` (0 or 1 each) |
| **alignment** | left, center, right |
| **borders** | Border flags (0=none, 1=draw) |

**Style Examples:** `000`=normal, `100`=bold, `010`=italic, `001`=underline, `111`=all

**Props Example:** `"Helvetica:12:100:left:1:1:1:1"` = Helvetica 12pt bold, left-aligned, all borders

### Page Sizes

| Size | Dimensions | Points |
|------|------------|--------|
| A4 | 8.27 × 11.69 in | 595 × 842 |
| Letter | 8.5 × 11 in | 612 × 792 |
| Legal | 8.5 × 14 in | 612 × 1008 |
| A3 | 11.69 × 16.54 in | 842 × 1191 |
| A5 | 5.83 × 8.27 in | 420 × 595 |

### Config Options

```json
{
  "config": {
    "page": "A4",              // Page size
    "pageAlignment": 1,        // 1=Portrait, 2=Landscape
    "pageBorder": "0:0:0:0",   // Border widths (left:right:top:bottom)
    "watermark": "DRAFT",      // Optional diagonal watermark
    "pdfTitle": "Document",    // PDF metadata title
    "pdfaCompliant": true,     // Enable PDF/A-4 compliance
    "arlingtonCompatible": true, // Enable PDF 2.0 Arlington Model
    "embedFonts": true,        // Embed fonts for portability
    "signature": { },          // Digital signature settings
    "security": { }            // Encryption settings
  }
}
```

<details>
<summary><b>🔐 Digital Signatures</b></summary>

Add legally-binding digital signatures with X.509 certificates:

```json
{
  "config": {
    "signature": {
      "enabled": true,
      "visible": true,
      "name": "John Doe",
      "reason": "Document Approval",
      "location": "New York, US",
      "contactInfo": "john@example.com",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n...",
      "certificatePem": "-----BEGIN CERTIFICATE-----\n...",
      "certificateChain": ["-----BEGIN CERTIFICATE-----\n..."]
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `enabled` | Enable digital signing |
| `visible` | Show signature stamp on document |
| `name` | Signer name (overrides certificate CN) |
| `reason` | Signing reason |
| `location` | Signing location |
| `privateKeyPem` | PEM-encoded private key (RSA/ECDSA) |
| `certificatePem` | PEM-encoded X.509 certificate |
| `certificateChain` | Intermediate certificates (optional) |
</details>

<details>
<summary><b>🔒 PDF Encryption</b></summary>

Password-protect documents with permission controls:

```json
{
  "config": {
    "security": {
      "enabled": true,
      "ownerPassword": "admin123",
      "userPassword": "view123",
      "allowPrinting": true,
      "allowCopying": false,
      "allowModifying": false
    }
  }
}
```

| Permission | Description |
|------------|-------------|
| `allowPrinting` | Allow document printing |
| `allowCopying` | Allow copying text/images |
| `allowModifying` | Allow content modification |
| `allowAnnotations` | Allow adding annotations |
| `allowFormFilling` | Allow filling form fields |
</details>

<details>
<summary><b>📑 Bookmarks & Navigation</b></summary>

Create document outlines with internal links:

```json
{
  "bookmarks": [
    {
      "title": "Chapter 1",
      "page": 1,
      "children": [
        { "title": "Section 1.1", "dest": "section-1-1" },
        { "title": "Section 1.2", "page": 2 }
      ]
    }
  ]
}
```

**Internal Links in Cells:**
```json
{
  "props": "Helvetica:10:000:left:1:1:1:1",
  "text": "Go to Summary",
  "link": "#financial-summary",
  "textcolor": "#0000FF"
}
```

**Named Destinations:**
```json
{
  "props": "Helvetica:12:100:left:1:1:1:1",
  "text": "FINANCIAL SUMMARY",
  "dest": "financial-summary"
}
```
</details>

---

## ✨ Features

| Category | Features |
|----------|----------|
| **PDF Generation** | JSON templates, auto page breaks, multi-page, page numbering |
| **Styling** | Bold/italic/underline, borders, custom fonts, alignments |
| **Elements** | Tables, checkboxes, radio buttons, fillable form fields, images |
| **Page Options** | A3/A4/A5/Letter/Legal, portrait/landscape, watermarks |
| **Security** | Digital signatures (PKCS#7), PDF encryption, password protection |
| **Compliance** | PDF/A-4, PDF 2.0 Arlington Model, PDF/UA accessibility |
| **Navigation** | Bookmarks/outlines, internal links, external hyperlinks, named destinations |
| **Tools** | PDF merge, form filling (XFDF), HTML to PDF/Image |
| **Deployment** | Single binary, Docker, REST API, React web UI |

---

## ❓ FAQ

<details>
<summary><b>Chrome not found error?</b></summary>

Install Google Chrome - required for HTML to PDF/Image conversion:
```bash
sudo apt install -y google-chrome-stable
```
</details>

<details>
<summary><b>How do auto page breaks work?</b></summary>

The system tracks Y position and creates new pages when content exceeds boundaries. Page borders and numbering are preserved across pages.
</details>

<details>
<summary><b>How do I create a digitally signed PDF?</b></summary>

Include the signature config with your PEM-encoded certificate and private key:
```json
{
  "config": {
    "signature": {
      "enabled": true,
      "visible": true,
      "certificatePem": "-----BEGIN CERTIFICATE-----\n...",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n..."
    }
  }
}
```
Supports RSA and ECDSA keys with optional certificate chains.
</details>

<details>
<summary><b>What is PDF/A-4 compliance?</b></summary>

PDF/A-4 is the archival standard based on PDF 2.0. Enable it with `"pdfaCompliant": true`. This embeds all fonts (via Liberation fonts), adds XMP metadata, and follows strict structure requirements for long-term preservation.
</details>

<details>
<summary><b>How do internal links work?</b></summary>

1. Add a destination anchor to a cell: `"dest": "my-section"`
2. Link to it from another cell: `"link": "#my-section"`
3. Optionally add a bookmark: `{"title": "My Section", "dest": "my-section"}`
</details>

<details>
<summary><b>XFDF form filling limitations?</b></summary>

Uses byte-oriented approach with `/NeedAppearances true`. Works for most AcroForms, but PDFs with compressed object streams may need a library like pdfcpu for full compatibility.
</details>

<details>
<summary><b>Performance benchmarks?</b></summary>

Sub-millisecond to ~7ms response times for complex 2-page financial reports. In-memory processing with zero external dependencies.
</details>

---

## 🛠️ Development

```bash
# Build
go build -o bin/gopdfsuit ./cmd/gopdfsuit

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/gopdfsuit-linux ./cmd/gopdfsuit

# Test
go test -cover ./...
```

### Project Structure

```
gopdfsuit/
├── cmd/gopdfsuit/      # Application entrypoint
├── frontend/           # React frontend (Vite)
├── internal/
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Template models
│   └── pdf/            # PDF generation & processing
├── docs/               # Built frontend assets
└── sampledata/         # Sample templates & data
```

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push (`git push origin feature/amazing`)
5. Open a Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

<div align="center">
  <p>Made with ❤️ by <a href="https://github.com/chinmay-sawant">Chinmay Sawant</a></p>
  <p>⭐ Star this repo if you find it helpful!</p>
  <p><em>Developed from scratch with assistance from <strong>GitHub Copilot</strong>.</em></p>
</div>
