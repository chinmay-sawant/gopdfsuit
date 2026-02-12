# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![gochromedp](https://img.shields.io/badge/gochromedp-1.0+-00ADD8?style=flat)](https://github.com/chinmay-sawant/gochromedp)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/chinmay-sawant/gopdfsuit)

> 🚀 A powerful Go web service for template-based PDF generation with multi-page support, PDF merging, form filling, and HTML to PDF/Image conversion.

## Star History

## [![Star History Chart](https://api.star-history.com/svg?repos=chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)](https://www.star-history.com/#chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)

## ⚡ Performance & Efficiency

**92% Cost Reduction** vs traditional distributed architectures.

| Metric               | Industry Standard (Typst/LaTeX) | gopdfsuit (Go 1.24)             |
| :------------------- | :------------------------------ | :------------------------------ |
| **Infrastructure**   | ~40 Node Cluster                | **2 Nodes** (95% Less)          |
| **Cost (1.5M PDFs)** | ~$10.20 / day                   | **~$0.77 / day**                |
| **Throughput**       | ~1k PDFs/sec (Cluster)          | **~600 PDFs/sec (Single Node)** |

> **Result**: Generates 1.5 million financial PDFs in ~45 mins on a single machine.

---

## 📑 Table of Contents

- [Overview](#-overview)
- [FAQ](#-faq)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)

### 📚 Documentation

- [🌐 Web Documentation](https://chinmay-sawant.github.io/gopdfsuit/#/documentation) - Interactive API documentation and playground

- [📋 Template Reference](guides/TEMPLATE_REFERENCE.md) - Complete JSON template format guide with examples
- [🛠️ Makefile Reference](guides/MAKEFILE.md) - Build, test, and deployment commands

---

## 📖 Overview

GoPdfSuit is a powerful Go web service for template-based PDF generation.

**Key Features:**

- **Template-Based Generation**: Create PDFs from JSON templates with auto page breaks and flow control.
- **Security & Compliance**: Digital signatures (PKCS#7, X.509), AES-256 encryption, granular permissions, and PDF/A-4 & PDF/UA-2 compliance.
- **Advanced Elements**: Rich text styling, tables, barcodes, QR codes, SVG vector graphics, and interactive forms (checkboxes, radio buttons).
- **Navigation**: Auto-generated bookmarks, internal links, and named destinations for easy document navigation.
- **Form Filling**: Fill generic AcroForms and XFDF data.
- **Merge & Split**: Combine multiple PDFs or split them.
- **HTML Conversion**: High-fidelity HTML to PDF/Image conversion using headless Chrome.
- **Native Bindings**:
  - **Python**: Direct CGO bindings for high-performance integration.
  - **Go**: Usable as a standalone Go library (`gopdflib`).
- **Web Interfaces**: Built-in React UI for viewer, editor, merger, filler, and converters.

**Requirements**: Go 1.24+, Google Chrome (for HTML conversion)

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
├── bindings/           # Native language bindings (Python CGO)
├── cmd/gopdfsuit/      # Application entrypoint
├── docs/               # Built frontend assets
├── frontend/           # React frontend (Vite)
├── guides/             # Documentation guides
├── internal/
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Gin middleware
│   ├── models/         # Template models
│   └── pdf/            # PDF generation & processing
├── pkg/
│   └── gopdflib/       # Standalone Go library
├── sampledata/         # Sample templates & data
└── test/               # Integration tests
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
