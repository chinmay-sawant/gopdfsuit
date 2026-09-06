# GoPdfSuit - three PDF engines, one repo

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![Python](https://img.shields.io/badge/Python-Bindings-3776AB?style=flat&logo=python)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/chinmay-sawant/gopdfsuit)

> gopdfsuit (REST API), gopdflib (Go library), pypdfsuit (Python bindings). I built it to generate PDFs from JSON templates, with multi-page support, merge, form fill, and HTML to PDF or image.

I got tired of hosted PDF services that charge per page and hide the layout logic. This repo keeps the template in JSON where you can read it, and the engine runs in memory with no Ghostscript and no external calls. Compliance output is slower, and that tradeoff is real. The numbers below show both sides.

## Star history

[![Star History Chart](https://star-history.dera.page/svg?repos=chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)](https://star-history.dera.page/#chinmay-sawant/gopdfsuit&type=timeline&logscale&legend=top-left)

## Performance and efficiency

Zerodha reference workload (5000 iterations, 48 workers, 80 percent retail, 15 percent active, 5 percent HFT), measured with x10 sequential runs on WSL2 (Intel i7-13700HX, Go 1.26.4, June 2026). See [documentation/BENCHMARKS.md](documentation/BENCHMARKS.md) for the full suite.

| Harness | PDF/A | PDF/UA | x10 peak | x10 mean | x10 median | Avg latency | Peak alloc |
| :------ | :---- | :----- | -------: | -------: | ---------: | :---------- | :--------- |
| `bench-gopdflib-zerodha` (compliant) | PDF/A-4 | PDF/UA-2 | 6,611 ops/s | 6,203 ops/s | 6,362 ops/s | 7.54 ms | 798 MB |
| `bench-gopdflib-zerodha-nocomply` | PDF 2.0 (no PDF/A) | None | 37,853 ops/s | 34,035 ops/s | 35,181 ops/s | 1.38 ms | 310 MB |

Compliant runs turn on PDF/A-4, PDF/UA-2, Arlington-compatible tagging, ECDSA P-256 signing, and font embedding (HFT output 2.29 MB, veraPDF 6/6 PASS). Non-compliant runs still output PDF 2.0 but skip PDF/A, tagging, signing, and font embedding to show the throughput ceiling (HFT output 221 KB).

> Full compliance hits about 6,600 ops/s peak on one machine, up 150 percent from the June 2026 baseline at 2,646 ops/s. The same workload without compliance hits about 37,900 ops/s peak, 5.7 times faster.

---

## Table of contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [FAQ](#faq)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

### Documentation

Start with the suite guide, then pick your surface:

- [PDF suite guide](documentation/PDF_SUITE_GUIDE.md) - one detailed walkthrough: install, generate, compress, merge, split, HTML, redact, fill, limits, privacy
- [Go library API](documentation/gopdflib-api.md) - every `pkg/gopdflib` op with copy-paste Go snippets
- [Python API](documentation/pypdfsuit-api.md) - every `pypdfsuit` function with copy-paste Python snippets ([package README](bindings/python/README.md))
- [REST API](documentation/rest-api.md) - every `/api/v1/*` endpoint with curl, auth, error envelope, caps
- [Engine overview](documentation/engine-overview.md) - in-memory pipeline: decode, render, fonts, tagging, XMP, sign, encrypt
- [Compress guide](documentation/compress-guide.md) - Light/Medium/Heavy tiers, limits, passthrough rules
- [HTML guide](documentation/html-guide.md) - pure-Go conversion fields, defaults, limits
- [Document ops guide](documentation/document-ops-guide.md) - merge, split page specs, XFDF fill, redact modes
- [WASM frontend](documentation/wasm-frontend.md) - local-first browser app, artifacts, consent, offline
- [Compliance and benchmarks](documentation/compliance-benchmarks.md) - PDF/A-4, PDF/UA-2, signatures, validation and bench commands
- [Samples catalog](documentation/samples-catalog.md) - every `sampledata/` directory with run commands
- [Template reference](documentation/TEMPLATE_REFERENCE.md) - JSON template format with examples
- [Docs index](documentation/index.md) - full map of `documentation/` (truth) vs `guides/` (frozen archive)
- [Web documentation](https://chinmay-sawant.github.io/gopdfsuit/#/documentation) - interactive API docs and playground
- [Makefile reference](guides/MAKEFILE.md) - build, test, and deployment commands

### Quickstart

Go library:

```bash
go get github.com/chinmay-sawant/gopdfsuit/v7@v7.0.0
```

```go
import "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"

b := gopdflib.NewDocument("A4", true)
b.AddTitle("Hello", gopdflib.WithTitleFontOpts(gopdflib.TitleFontOptions{Name: "Helvetica", Size: 18, Bold: true}))
pdfBytes, err := b.Generate()
```

See [Go library API](documentation/gopdflib-api.md) and [suite guide](documentation/PDF_SUITE_GUIDE.md).

Python bindings:

```bash
cd bindings/python && ./build.sh && pip install .
```

```python
from pypdfsuit.builder import TemplateBuilder, Font

b = TemplateBuilder("A4", True)
b.add_title("Hello", font="Helvetica", size=18, bold=True)
pdf_bytes = b.generate()
```

See [Python API](documentation/pypdfsuit-api.md) and [package README](bindings/python/README.md).

REST API (server on `:8080`):

```bash
docker run --rm -p 8080:8080 chinmaysawant/gopdfsuit:7.0.0
curl -X POST http://localhost:8080/api/v1/generate/template-pdf \
  -H "Content-Type: application/json" \
  --data-binary @template.json --output out.pdf
```

See [REST API](documentation/rest-api.md). Samples live in [sampledata/](sampledata/) ([catalog](documentation/samples-catalog.md)).

---

## Overview

Three apps, one repo. Pick what fits your stack.

| Component | Type | Use case |
| :-------- | :--- | :------- |
| gopdfsuit | Language-agnostic REST API | Run as a service. Call it from Go, Python, JS, curl, and others |
| gopdflib | Go library | Import `github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib` in your Go project |
| pypdfsuit | Python bindings | Use `from pypdfsuit.builder import TemplateBuilder, Font`. CGO wrapper around gopdflib |

Key features.

- **Template-based generation.** You send JSON, the engine returns PDF. It handles page breaks and flow across pages.
- **Fluent builders.** Go `gopdflib.NewDocument("A4", true)` plus `Font(...).Size(...).Cell(...)` in `pkg/gopdflib`, mirrored by Python `TemplateBuilder` in `bindings/python/pypdfsuit/builder.py`. See `documentation/BUILDER_FLUENT_GO.md` and `documentation/PY_BUILDER_PARITY.md`.
- **Security and compliance.** PKCS#7 and X.509 signatures, AES-256 encryption, granular permissions, PDF/A-4 and PDF/UA-2 when you opt in.
- **Advanced elements.** Text styling, tables, barcodes, QR codes, SVG vector graphics, and interactive forms with checkboxes and radio buttons.
- **Navigation.** Bookmarks, internal links, and named destinations the engine wires up for you.
- **Form filling.** Fills AcroForms and XFDF data.
- **Redaction.** Redacts by coordinates or by text search.
- **Merge and split.** Combines PDFs or splits them by page ranges.
- **PDF compress.** Shrinks PDFs in gopdflib or in the browser with WASM. No Ghostscript, no server round trip. Light, Medium, and Heavy JPEG tiers. Default path is local-first: `compress.wasm` on the Compress page, `gopdfsuit.wasm` elsewhere, server upload only after explicit consent. See `documentation/WASM_VIEWER_EDITOR.md`.
- **HTML conversion.** Turns HTML into PDF or image with pure-Go gowkhtmltopdf, no browser needed.
- **Web interfaces.** React UI for viewer, editor, merger, filler, and converters.

Requirements. Go 1.26.4, Make for build and test targets. HTML conversion is pure-Go, no Chrome needed. See [Prerequisites](#prerequisites) for the full list.

---

## Prerequisites

| Requirement | Version and notes |
|-------------|-----------------|
| Go | 1.26.4, must match `go.mod` |
| Make | Needed for `make build`, `make test`, `make run` |
| HTML conversion | Pure-Go via gowkhtmltopdf, no browser needed |
| Node.js plus npm | Frontend build, Node 18 or later |
| Python 3.8 or later | Python bindings tests for pypdfsuit |
| Java 11 or later | Optional, only for veraPDF PDF/A-4 and PDF/UA-2 validation with `make install-pdf-validators` |

### Windows

On Windows, use WSL. The project needs Make and Unix shell scripts, which PowerShell and CMD do not provide. See [CONTRIBUTING.md](CONTRIBUTING.md) for setup.

---

## FAQ

<details>
<summary><b>Go version compatibility?</b></summary>

This project needs Go 1.26.4 to build and run. The `go.mod` directive says `go 1.26.4`, and CI uses Go 1.26.4.

Install the exact version.

```bash
# Using go install (if you use multiple Go versions)
go install golang.org/dl/go1.26.4@latest
go1.26.4 download

# Verify
go1.26.4 version
```

Go 1.26.4 brings faster GC, better goroutine scheduling, and hardware crypto. The code does not need new language features, but I only test the module and deps against 1.26.4.

For older toolchains, you can try editing the `go` directive in `go.mod` and running `go mod tidy`. That path is unsupported. Releases track Go 1.26.4.

</details>

<details>
<summary><b>HTML conversion needs Chrome?</b></summary>

No. HTML to PDF or image is pure-Go via gowkhtmltopdf. No browser install needed.

</details>

<details>
<summary><b>How do auto page breaks work?</b></summary>

The engine tracks Y position and starts a new page when content passes the bottom margin. It redraws page borders and keeps page numbers correct across pages.

</details>

<details>
<summary><b>How do I create a digitally signed PDF?</b></summary>

Pass your PEM certificate and private key in the signature config.

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

The engine accepts RSA and ECDSA keys, with optional chains.

</details>

<details>
<summary><b>What is PDF/A-4 compliance?</b></summary>

PDF/A-4 is the archival standard built on PDF 2.0. Set `"pdfaCompliant": true`. The engine embeds Liberation fonts, adds XMP metadata, and enforces the structure rules needed for long term storage.

</details>

<details>
<summary><b>How do internal links work?</b></summary>

1. Add a destination anchor to a cell: `"dest": "my-section"`
2. Link to it from another cell: `"link": "#my-section"`
3. Add a bookmark if you want sidebar navigation: `{"title": "My Section", "dest": "my-section"}`
</details>

<details>
<summary><b>XFDF form filling limitations?</b></summary>

The filler sets `/NeedAppearances true` at the byte level. It covers most AcroForms. PDFs with compressed object streams may need pdfcpu for full compatibility.

</details>

<details>
<summary><b>Performance benchmarks?</b></summary>

I measured on Intel i7-13700HX (24 cores), WSL2, Go 1.26.4. Zerodha mix is 80 percent retail, 15 percent active, 5 percent HFT. All PDFs below use PDF/A-4 plus PDF/UA-2 except where noted. Retail signs with ECDSA P-256.

| Engine | Harness | Peak | Latest avg | Notes |
|--------|---------|-----:|-----------:|-------|
| gopdflib | Zerodha x10 (compliant) | 6,611 ops/s | 6,203 ops/s (x10, 2026-06-24) | PDF/A-4 plus PDF/UA-2, library in process |
| gopdflib | Zerodha x10 (nocomply) | 37,853 ops/s | 34,035 ops/s (x10, 2026-06-24) | PDF 2.0, compliance off |
| gopdfsuit | k6 `tagged_ecdsa` | 1,333 req/s | best-of-5 (2026-06-18) | HTTP plus Gin |
| pypdfsuit | Zerodha x10 (compliant) | 937 ops/s | 916 ops/s (x10, 2026-06-24) | PDF/A-4 plus PDF/UA-2, Python CGO |
| pypdfsuit | Zerodha x10 (nocomply) | 1,284 ops/s | 1,242 ops/s (x10, 2026-06-24) | PDF 2.0, compliance off |
| Gotenberg | k6 HTML to PDF | 16.1 req/s | best-of-5 (2026-06-18) | Chromium, no PDF/A |

Reproduce.

```bash
# gopdflib Zerodha reference (5000x48)
make bench-gopdflib-zerodha-x10
make bench-gopdflib-zerodha-nocomply-x10

# gopdfsuit (k6 plus Gin)
make bench-k6

# pypdfsuit (rebuild bindings first: cd bindings/python && ./build.sh)
make bench-pypdfsuit-zerodha-x10
make bench-pypdfsuit-zerodha-nocomply-x10
```

All processing runs in memory with zero external runtime deps.

</details>

---

## Development

> Windows users. Use WSL. Make and shell scripts are required. See [Prerequisites](#prerequisites).

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

# Format and lint
make fmt && make lint
```

### Project structure

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
│   └── pdf/            # PDF generation and processing
├── pkg/
│   └── gopdflib/       # Standalone Go library
├── sampledata/         # Sample templates and data
└── test/               # Integration tests
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, workflow, testing, and pull request rules.

Quick start.

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Run `make fmt && make lint && make test`
4. Commit and open a pull request

---

## License

MIT License - see [LICENSE](LICENSE)

---

<div align="center">
  <p>Made by <a href="https://github.com/chinmay-sawant">Chinmay Sawant</a></p>
  <p>Star this repo if it helped you ship a PDF faster.</p>
  <p><em>Built from scratch with assistance from GitHub Copilot.</em></p>
</div>
