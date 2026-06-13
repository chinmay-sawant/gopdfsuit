# 📄 GoPdfSuit - Three PDF Engines, One Repo

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![Python](https://img.shields.io/badge/Python-Bindings-3776AB?style=flat&logo=python)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![gochromedp](https://img.shields.io/badge/gochromedp-1.0+-00ADD8?style=flat)](https://github.com/chinmay-sawant/gochromedp)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/chinmay-sawant/gopdfsuit)

> 🚀 **gopdfsuit** (language-agnostic REST API) · **gopdflib** (Go library) · **pypdfsuit** (Python bindings) - template-based PDF generation with multi-page support, merging, form filling, and HTML to PDF/Image conversion.

## Star History

## [![Star History Chart](https://api.star-history.com/svg?repos=chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)](https://www.star-history.com/#chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)

## ⚡ Performance & Efficiency

**92% Cost Reduction** vs traditional distributed architectures. All benchmarks below run with **PDF/A-4 (PDF 2.0)** compliance enabled - no shortcuts.

| Metric               | Industry Standard (Typst/LaTeX) | gopdfsuit (Go 1.26)             |
| :------------------- | :------------------------------ | :------------------------------ |
| **Infrastructure**   | ~40 Node Cluster                | **2 Nodes** (95% Less)          |
| **Cost (1.5M PDFs)** | ~$10.20 / day                   | **~$0.77 / day**                |
| **Throughput**       | ~1k PDFs/sec (Cluster)          | **peak 2,953 / avg 2,646 PDFs/sec** (gopdflib Zerodha, 48 workers, 30-run, 2026-06-14) |

> **Result**: Generates 1.5 million financial PDFs in **~9.4 minutes** at 30-run avg throughput (**~8.5 min** at peak) on a single machine (Intel i7-13700HX, 24 cores, Go 1.26.4). Historical best (idle machine, Jun 13): peak **3,604** / avg **2,787** ops/s.

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

**Three applications, one repository** - pick the flavor that fits your stack:

| Component | Type | Use Case |
| :-------- | :--- | :------- |
| **gopdfsuit** | Language-agnostic REST API | Spin up as a service - call it from **any language** (Go, Python, JS, cURL, etc.) |
| **gopdflib** | Go library | `import "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"` directly in your Go project |
| **pypdfsuit** | Python bindings | `from pypdfsuit import Generator` - CGO-powered extension of gopdflib for Python |

**Key Features:**

- **Template-Based Generation**: Create PDFs from JSON templates with auto page breaks and flow control.
- **Security & Compliance**: Digital signatures (PKCS#7, X.509), AES-256 encryption, granular permissions, and PDF/A-4 & PDF/UA-2 compliance.
- **Advanced Elements**: Rich text styling, tables, barcodes, QR codes, SVG vector graphics, and interactive forms (checkboxes, radio buttons).
- **Navigation**: Auto-generated bookmarks, internal links, and named destinations.
- **Form Filling**: Fill generic AcroForms and XFDF data.
- **Redaction**: Securely redact sensitive information using specific coordinates or text search.
- **Merge & Split**: Combine multiple PDFs or split them.
- **HTML Conversion**: High-fidelity HTML to PDF/Image via headless Chrome.
- **Web Interfaces**: Built-in React UI for viewer, editor, merger, filler, and converters.

**Requirements**: Go 1.26+, Google Chrome (for HTML conversion)

---

## ❓ FAQ

<details>
<summary><b>Go version compatibility?</b></summary>

This module requires **Go 1.26+** to benefit from runtime performance improvements (better GC, goroutine scheduling, hardware-accelerated crypto). The `go.mod` directive is set to `go 1.26.4`.

**For Go 1.24 users:** You can still use `gopdflib` by cloning the repository and changing the `go` directive back to `1.26.4` in `go.mod`. The code itself does not use Go 1.26 language features — only the `sonic` dependency was bumped to `v1.15.2` for compatibility. Run `go mod tidy` after editing.

```bash
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit
# Edit go.mod: change "go 1.26.4" to "go 1.26.4"
go mod tidy
go build ./...
```

**Note:** The official module releases will track the latest stable Go version. If you need long-term compatibility with an older Go toolchain, maintain a fork or pin to an earlier tagged release.

</details>

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

Benchmarked on **Intel i7-13700HX (24 cores), WSL2, Go 1.26.4**. Zerodha workload: 80% retail · 15% active · 5% HFT. All PDFs **PDF/A-4 + PDF/UA-2**; retail **ECDSA P-256**.

| Engine | Harness | Peak | Latest avg | Notes |
|--------|---------|-----:|-----------:|-------|
| **gopdflib** | Zerodha `go run .` | **2,953 ops/s** | **2,646 ops/s** (30-run, 2026-06-14) | Library in-process |
| **gopdfsuit** | k6 `tagged_ecdsa` | **859 req/s** | **825 req/s** (5-run, 2026-06-14) | HTTP + Gin |
| **pypdfsuit** | `pypdfsuit_bench.py` | 228 ops/s | 219 ops/s (2-run, 2026-06-11) | Python CGO |
| **Gotenberg** | k6 HTML→PDF | — | **10.3 req/s** (2026-06-13) | Chromium, no PDF/A |

Reproduce:

```bash
# gopdflib
cd sampledata/gopdflib/zerodha && GOMAXPROCS=24 go1.26.4 run .

# gopdfsuit (k6 + Gin)
make load-pprof

# pypdfsuit
cd sampledata/gopdflib/zerodha && python3 pypdfsuit_bench.py
```

All processing is in-memory with zero external runtime dependencies.

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
