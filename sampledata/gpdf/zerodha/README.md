# gpdf Zerodha harness (bench-only comparison)

Survivor decision (row 5.6): `sampledata/gopdflib/zerodha` is the canonical
compliance-gated gold standard. It is listed in `test/compliance_manifest.json`,
checked by `test/zerodha_compliance_test.go`, and run through
`test/verify_pdfs.sh` (PDF/A-4 + PDF/UA-2).

This directory keeps the `gpdf` engine comparison harness
(`make bench-gpdf-zerodha*`) only. Its outputs are not compliance-gated and
must not be mistaken for release baselines.
