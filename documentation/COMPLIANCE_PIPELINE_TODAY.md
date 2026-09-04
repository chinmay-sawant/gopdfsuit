# Compliance pipeline today

Date: 2026-09-04. Branch: feat/builder-snippets.

## Gates

From makefile:

- test-verify-pdfs at 118 runs VERIFY_PDFS_JOBS bash test/verify_pdfs.sh. test-verify at 121 is the alias.
- test-zerodha-compliance at 166 runs verify_pdfs.sh --zerodha-only.
- test-scan-pdfs at 160 runs --scan-all. test-scan-pdfs-compliance at 163 runs --scan-all-compliance.
- install-verapdf at 154 and install-pdf-validators at 157 fetch project local validators.
- test-integration-zerodha at 110 runs go test with TestZerodhaPDFCompliance in ./test.

From PDF_VALIDATORS.md:7-11:

```bash
make install-pdf-validators
make test-verify-pdfs
make test-zerodha-compliance
```

## Manifest

test/compliance_manifest.json is the single source read by verify_pdfs.sh and zerodha_compliance_test.go.

Header at 1-12: version 1, defaults tolerance 0 plus flavours empty plus avalStrict false plus media pdf, perDirectoryAvalStrict for gopdflib zerodha false. Note at 3: veraPDF is hard gate, structure tree is hard gate on compliance entries, avalpdf warns by default. Per entry avalStrict allows opt in hard avalpdf. Zerodha was evaluated strict on 2026-09-04 and stays warn only because fixtures carry heuristic findings like no H1 and tables without TH that need engine tagging changes.

Entries at 13-299 carry path, baseline, tolerance or skipSize, flavours, avalStrict, media, suite:

- suite post-test: editor, filler with tolerance 700 and 500, htmltopdf and htmltoimg with skipSize true for live HTML, merge and split with tolerance 0, financialreport and typst with empty baseline meaning no size check.
- suite zerodha at 263-298: 3 entries zerodha_hft plus retail plus active output pdf, flavours 4 and ua2, avalStrict false, media pdf.
- Media types: pdf, png, zip.

Zerodha entry shape:

```json
{
  "path": "gopdflib/zerodha/zerodha_hft_output.pdf",
  "baseline": "",
  "tolerance": 0,
  "flavours": ["4","ua2"],
  "avalStrict": false,
  "media": "pdf",
  "suite": "zerodha"
}
```

## veraPDF

check_verapdf_compliance at verify_pdfs.sh:259 loops flavours and calls python3 test/verapdf_report.py check with verapdf, pdf, flavour, json out, sampledata args. Emits compliant or PASS 4 plus PASS ua2. Flavours 4 means PDF/A-4 and ua2 means PDF/UA-2.

TestZerodhaPDFCompliance at zerodha_compliance_test.go:74 reads suite zerodha entries through zerodhaManifestEntries at 31, maps labels at 53, runs parallel subtests invoking verapdf_report.py check, fails fast on failure. Skips when veraPDF is missing at 58.

check_valid_pdf at verify_pdfs.sh:112 uses verapdf lowLevelInfo parse check as Acrobat openable proof. PNG check at 138 looks for PNG magic. Zip check at 153 looks for PK header. Needs Java 11 plus, default path repo/verapdf/verapdf at verify_pdfs.sh:32.

## Structure tree

test/structure_tree_check.py is the strict check veraPDF misses. It caught the Zerodha ParentTree regression in 2026-06.

Rules in analyze_pdf:

- ParentTree MCID must map to TD or TH not TR at 88-104. TR refs fail.
- TR Pg must match child TD Pg at 106-129.
- Empty Sect without K is info only unless --fail-on-empty-sect at 166.

Runner check_structure_tree at verify_pdfs.sh:185 runs python3 structure_tree_check.py. Called only when flavours_csv is non empty in verify_manifest_entry at 558. Fail increments failures as hard gate, skip prints SKIP. Skipped when VERIFY_STRUCTURE_TREE is not 1.

Rule code at structure_tree_check.py:96-104:

```python
if report.parent_tree_tr_refs > 0:
    report.ok = False
    report.findings.append(Finding(
        code="parent_tree_tr_ref",
        message="ParentTree maps MCID slots to TR instead of TD/TH",
        count=report.parent_tree_tr_refs))
```

CLI: python3 test/structure_tree_check.py sampledata/gopdflib/zerodha/*.pdf per PDF_VALIDATORS.md:87.

Related unit pin touched today: structure_row_test.go TestEmitRowCellsMCIDAccounting checks consecutive MCID reservation, TR plus TD emit, BDC bytes, ParentTree growth. props_canonical_test.go pins parseProps fallbacks.

## avalpdf

check_avalpdf at verify_pdfs.sh:206 runs AVALPDF with report flag, parses validation_report json counts. Strict 1 plus issues above 0 means fail, else issues or warnings mean warn non blocking, clean means ok.

Strictness: VERIFY_AVALPDF_STRICT default 0 at 39, overridden per entry by manifest avalStrict through manifest_lookup at 338 and manifest_aval_strict at 379. Handling in verify_manifest_entry at 575: fail fails, warn prints INFO only.

Per PDF_VALIDATORS.md:33: not PDF/UA-2 conformance, WCAG style heuristics like missing H1 and tables without TH. Zerodha uses table layout not semantic headings, so warn only.

## Commands and env

- make test-verify-pdfs and make test-verify at makefile:118-121
- make test-zerodha-compliance at 166, modes --zerodha-only at 880
- make test-scan-pdfs and make test-scan-pdfs-compliance at 160-164
- go test -run TestZerodhaPDFCompliance ./test at zerodha_compliance_test.go:74
- verapdf -f 4 and -f ua2 per PDF_VALIDATORS.md:85-86
- python3 test/verapdf_report.py check per verify_pdfs.sh:276 and zerodha_compliance_test.go:95
- python3 test/structure_tree_check.py per verify_pdfs.sh:198
- .pdf-validators/venv/bin/avalpdf per PDF_VALIDATORS.md:88

Env at verify_pdfs.sh:12-46 and PDF_VALIDATORS.md:61-71: VERAPDF_BIN, AVALPDF_BIN, VERIFY_PDFS_JOBS, VERIFY_STRUCTURE_TREE=1, VERIFY_AVALPDF=1, VERIFY_AVALPDF_STRICT=0, NO_COLOR.

Note: sampledata/gopdflib/zerodha is the canonical compliance gated gold standard. sampledata/gpdf/zerodha harness is bench only, not compliance gated, per sampledata/gpdf/zerodha/README.md:2-6.
