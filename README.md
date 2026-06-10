# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org/)
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

| Metric               | Industry Standard (Typst/LaTeX) | gopdfsuit (Go 1.26)             |
| :------------------- | :------------------------------ | :------------------------------ |
| **Infrastructure**   | ~40 Node Cluster                | **2 Nodes** (95% Less)          |
| **Cost (1.5M PDFs)** | ~$10.20 / day                   | **~$0.77 / day**                |
| **Throughput**       | ~1k PDFs/sec (Cluster)          | **~2,750 PDFs/sec peak** (single node, 48 workers) |

> **Result**: Generates 1.5 million financial PDFs in **~11 minutes** at peak throughput on a single machine (Intel i7-13700HX, 24 cores, Go 1.26.4).

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
- **Redaction**: Securely redact sensitive information using specific coordinates or text search.
- **Merge & Split**: Combine multiple PDFs or split them.
- **HTML Conversion**: High-fidelity HTML to PDF/Image conversion using headless Chrome.
- **Native Bindings**:
  - **Python**: Direct CGO bindings for high-performance integration.
  - **Go**: Usable as a standalone Go library (`gopdflib`).
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

Benchmarked on **Intel i7-13700HX (24 cores), WSL2, Go 1.26.4** — 10 runs each, peak values reported.

### gopdflib — Zerodha Gold Standard (48 workers, PDF/A + tagged + signed)

Real-world brokerage workload: 80% retail (1-page) · 15% active trader (2–3 page) · 5% HFT (50+ page). Every PDF generated from scratch. Retail signed with **ECDSA P-256** (`BENCH_SIGN_RSA=1` for RSA).

| Metric | Peak (best of 5) | 5-run average |
|--------|-----------------:|--------------:|
| **Throughput** | **2,751 ops/s** | 2,476 ops/s |
| **Avg latency** | **16.9 ms** | 18.9 ms |
| **Wall time (5,000 docs)** | **1.82 s** | 2.03 s |

Reproduce: `BENCH_SKIP_WRITE=1 BENCH_SEED=42 GOMAXPROCS=24 go1.26.4 run .` from `sampledata/gopdflib/zerodha`

### gopdfsuit — HTTP load test (k6, 48 VUs, PDF/A tagged payloads)

| Metric | Peak (best of 10) | 10-run average |
|--------|------------------:|---------------:|
| **Throughput** | **520 req/s** | 496 req/s |
| **Median latency** | **16 ms** | ~17 ms |
| **p99 latency** | **287 ms** | ~320 ms |

Reproduce: start server (`go run ./cmd/gopdfsuit`), then `k6 run test/generate_template-pdf/load_test.js`

### gopdfsuit — Micro-benchmarks (single-thread, PDF/A, 2,000-row table)

| Benchmark | Peak (best of 10) | 10-run average |
|-----------|------------------:|---------------:|
| `Rows2000` serial | **42.5 ms/op** | 54.8 ms/op |
| `WrapEnabled/Rows2000` | **32.2 ms/op** | 33.8 ms/op |
| `GoPdfSuit` (data.json) | **16.4 ms/op** | 17.1 ms/op |

Reproduce: `go test -run='^$' -bench='BenchmarkGenerateTemplatePDF/Rows2000$|BenchmarkGoPdfSuit$' -benchmem -count=10 ./internal/pdf/`

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
