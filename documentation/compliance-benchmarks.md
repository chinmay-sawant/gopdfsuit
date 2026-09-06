# Compliance and Benchmarks

Router summary for compliance gates, signatures, encryption, and benchmark
harnesses. Detail lives in the source docs linked at the bottom. This file
does not duplicate them.

## PDF/A-4 opt-in plus PDF/UA-2

- Base output is PDF 2.0. PDF/A-4 plus PDF/UA-2 is opt-in per template
  (`pdfaCompliant: true`, tagging flags such as `TaggedPDF`).
- The Zerodha fixtures under `sampledata/gopdflib/zerodha/` are the canonical
  compliance-gated gold standard. The `sampledata/gpdf/zerodha/` harness is
  bench-only, not compliance-gated.
- veraPDF is the hard gate for PDF/A-4 and PDF/UA-2.
  Needs Java 11+. Never trust veraPDF alone: ParentTree MCID checks need
  `structure_tree_check.py` (MCID must point to TD or TH, not TR).
- avalpdf output is warnings-only unless `VERIFY_AVALPDF_STRICT=1`.

## RSA plus ECDSA P-256 signatures

- Engine: `internal/pdf/signature/`. Output is a PKCS#7 detached signature.
- RSA is the existing default and keeps working unchanged. ECDSA P-256 is
  opt-in by swapping the private-key PEM plus certificate PEM.
- Key and certificate must match algorithms. Mixed RSA key plus EC cert (or
  reverse) errors at parse or sign time.
- Tests: `go test ./internal/pdf/signature/... -v`.

## Encryption

> Doc and code mismatch note: top-level docs say AES-256, but the
> implementation in `internal/pdf/encryption/encrypt.go` is AES-128.
> Treat AES-128 as current truth until the engine or the top-level docs
> are updated.

- Owner password required. Permission flags map to `/P` bits (printing,
  modifying, copying, annotations, form filling, accessibility, assembly,
  high-quality print).
- Failures fail closed, never plaintext fallback.

## Validation commands

```bash
make test-verify-pdfs          # post-test gate: test/verify_pdfs.sh
make test-verify               # alias for the above
make test-zerodha-compliance   # verify_pdfs.sh --zerodha-only
make test-scan-pdfs            # verify_pdfs.sh --scan-all (every sampledata PDF)
make test-scan-pdfs-compliance # scan plus compliance table
go test -run TestZerodhaPDFCompliance ./test
python3 test/structure_tree_check.py sampledata/gopdflib/zerodha/*.pdf
make install-verapdf           # fetch project-local veraPDF CLI
make install-pdf-validators    # fetch avalpdf venv
```

Useful env: `VERIFY_PDFS_JOBS`, `VERAPDF_BIN`, `VERIFY_STRUCTURE_TREE=1`,
`VERIFY_AVALPDF=1`, `VERIFY_AVALPDF_STRICT=0`, `NO_COLOR=1`. Manifest
tolerances, baselines, and flavours live in `test/compliance_manifest.json`.

## Benchmark commands

```bash
make bench-smoke                # quick slice of every headless harness (<5 min)
make bench-gopdflib-zerodha     # Go in-process, compliant Zerodha mix
make bench-gopdflib-zerodha-nocomply   # same workload, compliance off
make bench-gopdflib-zerodha-x10 # x10 sequential gold standard
make bench-gopdfsuit-zerodha    # retail-only signed fast path
make bench-pypdfsuit-zerodha    # same mix via CGO/Python (rebuild bindings first)
make bench-handler              # Gin plus decode plus engine cost, no network
make bench-k6                   # live HTTP weighted load
make bench-gotenberg            # Chromium HTML-to-PDF baseline (needs Docker)
```

Notes:

- Rebuild CGO first for Python benches: `cd bindings/python && ./build.sh`.
- Raw logs and parsed stats live under `guides/cursor/baselines/`.

## Sources

- [COMPLIANCE_PIPELINE_TODAY.md](COMPLIANCE_PIPELINE_TODAY.md) - gates,
  manifest shape, veraPDF plus structure-tree plus avalpdf layering.
- [DIGITAL_SIGNATURE_RSA_ECDSA.md](DIGITAL_SIGNATURE_RSA_ECDSA.md) -
  key setup, PEM swap, OpenSSL steps.
- [BENCHMARKS.md](BENCHMARKS.md) - best-of tables, harness map,
  canonical payload pin, reproduce steps.
- [compliance_manifest.json](../test/compliance_manifest.json) - single source
  for tolerances and flavours.
- Runner: [verify_pdfs.sh](../test/verify_pdfs.sh).
