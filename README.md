# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![gochromedp](https://img.shields.io/badge/gochromedp-1.0+-00ADD8?style=flat)](https://github.com/chinmay-sawant/gochromedp)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🚀 A powerful Go web service for template-based PDF generation with multi-page support, PDF merging, form filling, and HTML to PDF/Image conversion.

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

---

## 📖 Overview

GoPdfSuit is a Go + Gin web service that generates professional PDF documents from JSON templates. 

**Key Capabilities:**
- 📄 Template-based PDF generation with auto page breaks
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
    "config": {"page": "A4", "pageAlignment": 1, "pageBorder": "1:1:1:1"},
    "title": {"props": "font1:24:100:center:0:0:1:0", "text": "Document Title"},
    "table": [{"maxcolumns": 2, "rows": [{"row": [
      {"props": "font1:12:100:left:1:1:1:1", "text": "Field:"},
      {"props": "font1:12:000:left:1:1:1:1", "text": "Value"}
    ]}]}],
    "footer": {"font": "font1:10:000:center", "text": "Footer"}
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

### Props Syntax
`"fontname:fontsize:style:alignment:left:right:top:bottom"`

| Component | Values |
|-----------|--------|
| **fontname** | font1, font2, etc. |
| **fontsize** | Size in points |
| **style** | 3-digit code: `[bold][italic][underline]` (0 or 1 each) |
| **alignment** | left, center, right |
| **borders** | Border widths (left:right:top:bottom) |

**Style Examples:** `000`=normal, `100`=bold, `010`=italic, `001`=underline, `111`=all

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
    "page": "A4",           // Page size
    "pageAlignment": 1,     // 1=Portrait, 2=Landscape
    "pageBorder": "1:1:1:1", // Border widths
    "watermark": "DRAFT"    // Optional diagonal watermark
  }
}
```

---

## ✨ Features

| Category | Features |
|----------|----------|
| **PDF Generation** | JSON templates, auto page breaks, multi-page, page numbering |
| **Styling** | Bold/italic/underline, borders, custom fonts, alignments |
| **Elements** | Tables, checkboxes, radio buttons, fillable form fields |
| **Page Options** | A3/A4/A5/Letter/Legal, portrait/landscape, watermarks |
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
<summary><b>XFDF form filling limitations?</b></summary>

Uses byte-oriented approach with `/NeedAppearances true`. Works for most AcroForms, but PDFs with compressed object streams may need a library like pdfcpu for full compatibility.
</details>

<details>
<summary><b>Performance benchmarks?</b></summary>

Sub-millisecond to 1.7ms response times for 2-page documents. In-memory processing with zero external dependencies.
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