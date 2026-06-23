# 📄 GoPdfSuit - Three PDF Engines, One Repo

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat&logo=go)](https://golang.org/)
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

Zerodha gold-standard workload (5000 iterations, 48 workers, 80% retail · 15% active · 5% HFT), measured with x10 sequential runs on WSL2 (Intel i7-13700HX, Go 1.26.4, June 2026). See [guides/BENCHMARKS.md](guides/BENCHMARKS.md) for full suite.

| Harness | PDF/A | PDF/UA | x10 peak | x10 mean | x10 median | Avg latency | Peak alloc |
| :------ | :---- | :----- | -------: | -------: | ---------: | :---------- | :--------- |
| **`bench-gopdflib-zerodha`** (compliant) | PDF/A-4 | PDF/UA-2 | **6,611 ops/s** | **6,203 ops/s** | **6,362 ops/s** | 7.54 ms | 798 MB |
| **`bench-gopdflib-zerodha-nocomply`** | PDF 2.0 (no PDF/A) | None | **37,853 ops/s** | **34,035 ops/s** | **35,181 ops/s** | 1.38 ms | 310 MB |

**Compliant** runs enable PDF/A-4, PDF/UA-2, Arlington-compatible tagging, ECDSA P-256 signing, and font embedding (HFT output **2.29 MB**, veraPDF 6/6 PASS). **Non-compliant** still outputs **PDF 2.0** but turns PDF/A, tagging, signing, and font embedding off for a throughput ceiling reference (HFT output **221 KB**).

> **Headline:** Full compliance delivers **~6,600 ops/s** peak on one machine - **+150%** vs the June 2026 baseline (2,646 ops/s) - while the same workload without compliance reaches **~37,900 ops/s** peak (**5.7×** faster).

---

## 📑 Table of Contents

- [Overview](#-overview)
- [Prerequisites](#-prerequisites)
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

**Requirements**: Go **1.26.4**, Google Chrome (for HTML conversion), Make (for build/test targets). See [Prerequisites](#-prerequisites) for the full list.

---

## 📋 Prerequisites

| Requirement | Version / notes |
|-------------|-----------------|
| **Go** | **1.26.4** (required - matches `go.mod`) |
| **Make** | Required for `make build`, `make test`, `make run`, and other targets |
| **Google Chrome** | Required for HTML→PDF/Image conversion |
| **Node.js + npm** | Frontend build (Node 18+ recommended) |
| **Python 3.8+** | Python bindings tests (`pypdfsuit`) |
| **Java 11+** | Optional - needed for veraPDF PDF/A-4 + PDF/UA-2 validation (`make install-pdf-validators`) |

### Windows

On Windows, use **WSL (Windows Subsystem for Linux)** for the best compatibility. The project relies on **Make** and Unix shell scripts that are not available in PowerShell or CMD. See [CONTRIBUTING.md](CONTRIBUTING.md) for setup details.

---

## ❓ FAQ

<details>
<summary><b>Go version compatibility?</b></summary>

This project requires **Go 1.26.4** to build and run. The `go.mod` directive is set to `go 1.26.4`, and CI uses Go 1.26.4.

Install the exact version:

```bash
# Using go install (if you use multiple Go versions)
go install golang.org/dl/go1.26.4@latest
go1.26.4 download

# Verify
go1.26.4 version
```

Go 1.26.4 is recommended for runtime performance improvements (better GC, goroutine scheduling, hardware-accelerated crypto). The code does not rely on unreleased language features, but the module and dependencies are tested against **1.26.4** only.

**For older Go toolchains:** You can try changing the `go` directive in `go.mod` and running `go mod tidy`, but this is unsupported - official releases track Go 1.26.4.

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
| **gopdflib** | Zerodha x10 (compliant) | **6,611 ops/s** | **6,203 ops/s** (x10, 2026-06-24) | PDF/A-4 + PDF/UA-2, library in-process |
| **gopdflib** | Zerodha x10 (nocomply) | **37,853 ops/s** | **34,035 ops/s** (x10, 2026-06-24) | PDF 2.0, compliance off |
| **gopdfsuit** | k6 `tagged_ecdsa` | **1,333 req/s** | best-of-5 (2026-06-18) | HTTP + Gin |
| **pypdfsuit** | Zerodha x10 (compliant) | **937 ops/s** | **916 ops/s** (x10, 2026-06-24) | PDF/A-4 + PDF/UA-2, Python CGO |
| **pypdfsuit** | Zerodha x10 (nocomply) | **1,284 ops/s** | **1,242 ops/s** (x10, 2026-06-24) | PDF 2.0, compliance off |
| **Gotenberg** | k6 HTML→PDF | **16.1 req/s** | best-of-5 (2026-06-18) | Chromium, no PDF/A |

Reproduce:

```bash
# gopdflib Zerodha gold standard (5000×48)
make bench-gopdflib-zerodha-x10
make bench-gopdflib-zerodha-nocomply-x10

# gopdfsuit (k6 + Gin)
make bench-k6

# pypdfsuit (rebuild bindings first: cd bindings/python && ./build.sh)
make bench-pypdfsuit-zerodha-x10
make bench-pypdfsuit-zerodha-nocomply-x10
```

All processing is in-memory with zero external runtime dependencies.

</details>

---

## 🛠️ Development

> **Windows users:** Use WSL - Make and shell scripts are required. See [Prerequisites](#-prerequisites).

```bash
# Build
make build
# or directly:
go build -o bin/gopdfsuit ./cmd/gopdfsuit

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/gopdfsuit-linux ./cmd/gopdfsuit

# Test
make test
# or:
go test -cover ./...

# Format & lint
make fmt && make lint
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

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for setup, development workflow, testing, and pull request guidelines.

Quick start:

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Run `make fmt && make lint && make test`
4. Commit changes and open a Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

<div align="center">
  <p>Made with ❤️ by <a href="https://github.com/chinmay-sawant">Chinmay Sawant</a></p>
  <p>⭐ Star this repo if you find it helpful!</p>
  <p><em>Developed from scratch with assistance from <strong>GitHub Copilot</strong>.</em></p>
</div>
