# PDF Validators

gopdfsuit targets **PDF/A-4** and **PDF/UA-2** for compliant Zerodha workloads. No single open-source tool covers every rule that commercial auditors apply, so we run a **layered validation stack**.

## Quick start

```bash
make install-pdf-validators   # veraPDF + avalpdf (project-local, gitignored)
make test-verify-pdfs         # full post-test manifest
make test-zerodha-compliance  # Zerodha retail/active/HFT only
```

## Validator stack

| Tool | Type | PDF/A-4 | PDF/UA-2 | Install | CI | Catches ParentTree bugs |
|------|------|---------|----------|---------|----|-------------------------|
| **veraPDF** | ISO profiles (primary gate) | Yes (`-f 4`) | Yes (`-f ua2`) | `make install-verapdf` | Yes | No |
| **structure_tree_check.py** | Custom strict structural checks | - | - | Built-in (Python 3) | Yes | **Yes** |
| **avalpdf** | Tag-tree heuristics (supplementary) | No | No (WCAG-style) | `make install-pdf-validators` | Yes | Partial |
| **PAC 2026** (axes4) | Matterhorn / PDF/UA reference checker | Indirect | Yes (PDF/UA + WCAG) | Windows GUI only | Manual | **Yes** |
| **callas pdfToolbox CLI** | Commercial preflight | Yes (profile) | PDF/UA-1 syntax | Docker + license | Optional | **Yes** |
| **Adobe Acrobat Preflight** | Commercial desktop | Yes | Yes | Desktop license | Manual | **Yes** |

### Why veraPDF alone is not enough

veraPDF validates ISO PDF/A and PDF/UA **machine-checkable profiles** and is the industry open-source baseline. It passed Zerodha PDFs while **PAC/Adobe** reported *"Structure Tree is structurally invalid"* because stricter tools also verify:

- `ParentTree[MCID]` must reference the **owning TD/TH**, not the TR parent
- TR `/Pg` must match child TD pages on multi-page tables

`test/structure_tree_check.py` enforces those rules in CI so we do not rely on veraPDF for structural consistency.

### avalpdf (supplementary)

[avalpdf](https://github.com/dennisangemi/avalpdf) checks tagged structure, headings, table/list semantics, and alt text heuristics. It is **not** a PDF/UA-2 conformance validator, but it surfaces accessibility issues veraPDF ignores (e.g. missing H1, tables without `TH` headers).

By default avalpdf findings are **warnings** in `verify_pdfs.sh` (Zerodha contract notes use table layout, not semantic headings). Set `VERIFY_AVALPDF_STRICT=1` to fail on avalpdf issues.

## Installation

### Automated (recommended)

```bash
make install-pdf-validators
```

Installs into gitignored paths:

| Path | Contents |
|------|----------|
| `verapdf/` | veraPDF CLI |
| `.pdf-validators/venv/` | avalpdf + pdfix-sdk |

### Manual veraPDF only

```bash
make install-verapdf
# Requires Java 11+: sudo apt install openjdk-11-jre-headless
```

### Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `VERAPDF_BIN` | `<repo>/verapdf/verapdf` | veraPDF CLI path |
| `AVALPDF_BIN` | `<repo>/.pdf-validators/venv/bin/avalpdf` | avalpdf CLI path |
| `VERIFY_PDFS_JOBS` | `nproc` or `4` | Parallel workers |
| `VERIFY_AVALPDF` | `1` | Run avalpdf on compliance PDFs (`0` to skip) |
| `VERIFY_AVALPDF_STRICT` | `0` | Fail on avalpdf issues when `1` |
| `VERIFY_STRUCTURE_TREE` | `1` | Run structure_tree_check.py (`0` to skip) |
| `NO_COLOR` | unset | Disable ANSI colours |

## Commands

```bash
# Primary gates (used by make test)
make test-verify-pdfs
make test-zerodha-compliance

# Scan all sampledata PDFs
make test-scan-pdfs
make test-scan-pdfs-compliance

# Individual tools
verapdf/verapdf -f 4 sampledata/gopdflib/zerodha/zerodha_retail_output.pdf
verapdf/verapdf -f ua2 sampledata/gopdflib/zerodha/zerodha_retail_output.pdf
python3 test/structure_tree_check.py sampledata/gopdflib/zerodha/*.pdf
.pdf-validators/venv/bin/avalpdf sampledata/gopdflib/zerodha/zerodha_retail_output.pdf
```

## Optional commercial validators

### callas pdfToolbox CLI (Docker)

Stricter PDF/A-4 and PDF/UA checks; requires a **callas license**.

```bash
docker pull callassoftware/pdftoolbox-cli:latest

# Example (requires activation - set CALLAS_LICENSE or use license server)
docker run --rm --entrypoint /bin/bash \
  -v "$PWD/sampledata:/pdfs" -v "$PWD/.pdftoolbox-reports:/out" \
  callassoftware/pdftoolbox-cli:latest -c \
  '/opt/callas/callas_pdfToolboxCLI_Linux_*/pdfToolbox -o=/out \
    "/opt/callas/callas_pdfToolboxCLI_Linux_*/var/Profiles/PDFA compliance/Verify compliance with PDFA-4.kfpx" \
    /pdfs/gopdflib/zerodha/zerodha_retail_output.pdf'
```

Profiles in the image include `Verify compliance with PDFA-4.kfpx` and `Verify compliance with PDFUA-1 (syntax checks only).kfpx` (PDF/UA-2 profiles may require a newer license tier).

### PAC 2026 (axes4)

Free Windows GUI - the checker that reported the Zerodha structure-tree failure. Download: [pac.pdf-accessibility.org](https://pac.pdf-accessibility.org/en/download). No Linux CLI; use on Windows for pre-release manual sign-off.

## CI integration

`.github/workflows/frontend-build-commit.yml` runs `make install-verapdf` and `make test`, which executes `test/verify_pdfs.sh` (veraPDF + structure tree + avalpdf when installed).

Zerodha compliance PDFs are always checked for PDF/A-4 and PDF/UA-2:

- `sampledata/gopdflib/zerodha/zerodha_retail_output.pdf`
- `sampledata/gopdflib/zerodha/zerodha_active_output.pdf`
- `sampledata/gopdflib/zerodha/zerodha_hft_output.pdf`

## Pre-release checklist

1. `make test` - unit + integration + PDF validation
2. `make test-zerodha-compliance` - fast Zerodha gate
3. Optional: PAC on Windows for Matterhorn Protocol manual checkpoints
4. Optional: pdfToolbox with license for client-mandated preflight reports