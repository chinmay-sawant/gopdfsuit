# Digital Signatures — RSA vs ECDSA (gopdfsuit)

**Date:** 2026-06-11  
**Applies to:** `internal/pdf/signature/`, `SignatureConfig` in API/JSON templates  
**Sample certs:** `certs/` (RSA and ECDSA leaf keys)

---

## Summary

gopdfsuit supports **both RSA and ECDSA P-256** for PDF digital signatures. Adding ECDSA is **opt-in** — it does **not** break or replace existing RSA-2048 keys. You choose the algorithm by which PEM files you pass in `signature.privateKeyPem` and `signature.certificatePem`.

| Algorithm | Key strength | Sample files | Typical use |
|-----------|--------------|--------------|-------------|
| **RSA-2048** | 2048-bit RSA | `certs/leaf.key`, `certs/leaf.pem` | Existing production configs |
| **ECDSA P-256** | Elliptic curve `prime256v1` (~128-bit security level) | `certs/ec_leaf.key`, `certs/ec_leaf.pem` | Faster signing (benchmark default) |

---

## Will ECDSA break my existing RSA key?

**No.** Existing RSA-2048 setups continue to work unchanged.

| What you have today | Still works? |
|---------------------|:------------:|
| RSA-2048 private key (`certs/leaf.key`) | Yes |
| RSA leaf certificate (`certs/leaf.pem`) | Yes |
| JSON/API configs with RSA PEMs in `signature` | Yes |
| Intermediate + root CA chain (`chain.pem`) | Yes (signs both RSA and EC leaf certs) |

ECDSA was added as an **additional** signing path for performance. Sample payloads under `sampledata/python/` and `sampledata/editor/` still use **RSA-2048** unless you change the PEMs yourself.

### Terminology

- **RSA-2048** — 2048-bit RSA key. This is what `certs/generate_certs.sh` creates for `leaf.key`.
- **ECDSA P-256** (`prime256v1`) — elliptic-curve key with roughly **128-bit security strength**. It is **not** a “128-bit RSA key” (which would be insecure). It is a different algorithm family entirely.

---

## How the library picks RSA vs ECDSA

`NewPDFSigner` parses the private key PEM and branches on key type:

1. Try PKCS#8 (`BEGIN PRIVATE KEY`)
2. Fall back to PKCS#1 RSA (`BEGIN RSA PRIVATE KEY`)
3. Fall back to SEC1 EC (`BEGIN EC PRIVATE KEY`)

| Private key type | Signature algorithm in PKCS#7 |
|------------------|-------------------------------|
| `*rsa.PrivateKey` | RSA PKCS#1 v1.5 + SHA-256 |
| `*ecdsa.PrivateKey` (P-256) | ECDSA + SHA-256 (`ecdsa-with-SHA256`) |
| Anything else | Error: *unsupported private key type* |

Relevant code: `internal/pdf/signature/signature.go` (`parseSignerPEMMaterials`, `NewPDFSigner`, `SignPDF`).

### Critical rule: key and certificate must match

The private key and leaf certificate **must be the same algorithm**. The library does not convert or migrate keys.

| Config you pass | Result |
|-----------------|--------|
| `leaf.key` + `leaf.pem` (RSA) | RSA PKCS#7 signature |
| `ec_leaf.key` + `ec_leaf.pem` (ECDSA) | ECDSA PKCS#7 signature |
| RSA key + EC certificate | **Error** at parse/sign time |
| EC key + RSA certificate | **Error** at parse/sign time |

---

## Generate ECDSA keys with OpenSSL

Your intermediate CA can remain **RSA**. Only the **leaf signer** needs to be ECDSA. This matches the layout in `certs/ec_leaf.*`.

### Option A — ECDSA leaf with existing intermediate CA

Run from the repo root after you already have `certs/intermediate.pem` and `certs/intermediate.key` (from `generate_certs.sh` or your own PKI):

```bash
cd certs

# 1. Generate ECDSA P-256 private key
openssl ecparam -genkey -name prime256v1 -noout -out ec_leaf.key

# OpenSSL 3+ alternative:
# openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out ec_leaf.key

# 2. Create certificate signing request (CSR)
openssl req -new -key ec_leaf.key -out ec_leaf.csr \
  -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Leaf/CN=GoPDFSuitSigner"

# 3. Sign leaf certificate with intermediate CA
openssl x509 -req -in ec_leaf.csr \
  -CA intermediate.pem -CAkey intermediate.key -CAcreateserial \
  -out ec_leaf.pem -days 365 -sha256

# 4. Verify curve
openssl x509 -in ec_leaf.pem -noout -text | grep -A2 "Public Key Algorithm"
# Expected: id-ecPublicKey, ASN1 OID: prime256v1
```

### Option B — Full chain from scratch (RSA CA + ECDSA leaf)

1. Generate the RSA CA chain and RSA leaf (existing script):

```bash
bash certs/generate_certs.sh
```

2. Add the ECDSA leaf steps from **Option A** (steps 1–4 above). The script today only creates RSA `leaf.key` / `leaf.pem`; ECDSA leaf generation is a separate step.

### Files produced

| File | Purpose |
|------|---------|
| `ec_leaf.key` | Signer private key → `signature.privateKeyPem` |
| `ec_leaf.pem` | Signer certificate → `signature.certificatePem` |
| `ec_leaf.csr` | CSR (intermediate step; not needed in runtime config) |
| `intermediate.pem` + `rootCA.pem` | Optional → `signature.certificateChain` |

For RSA-only setup, `generate_certs.sh` already documents:

- `leaf.pem` → `certificatePem`
- `leaf.key` → `privateKeyPem`
- `chain.pem` → intermediate + root for the chain

---

## Using keys in gopdfsuit

### ECDSA P-256 example

```json
{
  "config": {
    "signature": {
      "enabled": true,
      "privateKeyPem": "-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----",
      "certificatePem": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
      "certificateChain": [
        "-----BEGIN CERTIFICATE-----\n...intermediate...\n-----END CERTIFICATE-----"
      ],
      "reason": "Document approval",
      "location": "Mumbai, India"
    }
  }
}
```

### RSA-2048 example (existing / unchanged)

```json
{
  "config": {
    "signature": {
      "enabled": true,
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
      "certificatePem": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
      "certificateChain": ["..."]
    }
  }
}
```

PEM strings can use `\n` escapes in JSON (as in `sampledata/python/financial_digitalsignature.json`) or literal newlines depending on your loader.

### Go / gopdflib

```go
template.Config.Signature = &gopdflib.SignatureConfig{
    Enabled:          true,
    PrivateKeyPEM:    string(keyPEM),   // from ec_leaf.key or leaf.key
    CertificatePEM:   string(certPEM),  // matching leaf.pem or ec_leaf.pem
    CertificateChain: chainPEMs,
}
```

---

## RSA vs ECDSA comparison

| | RSA-2048 (existing) | ECDSA P-256 (optional) |
|--|---------------------|-------------------------|
| Key file | `certs/leaf.key` | `certs/ec_leaf.key` |
| Cert file | `certs/leaf.pem` | `certs/ec_leaf.pem` |
| Typical PEM header | `BEGIN PRIVATE KEY` / `BEGIN RSA PRIVATE KEY` | `BEGIN EC PRIVATE KEY` |
| PDF output | PKCS#7 detached signature | PKCS#7 detached signature |
| Adobe / PDF readers | Widely supported | Widely supported (P-256) |
| Signing CPU (benchmark) | Higher (~10–16% machine CPU in RSA profile) | Lower (~3–5% with ECDSA) |
| Breaks existing deploy? | — | **No** (opt-in PEM swap) |

---

## Benchmark note (Zerodha harness)

The Zerodha benchmark (`sampledata/gopdflib/zerodha/main.go`) defaults to **ECDSA P-256** for retail signing throughput. To profile or compare RSA:

```bash
BENCH_SIGN_RSA=1 go run .   # from sampledata/gopdflib/zerodha
```

Production JSON templates are unaffected unless you change the PEMs.

---

## Tests and verification

```bash
go test ./internal/pdf/signature/... -v -run 'RSA|ECDSA'
```

Tests load `certs/leaf.*` (RSA) and `certs/ec_leaf.*` (ECDSA) and verify signer creation and PKCS#7 output.

---

## Related files

| Path | Description |
|------|-------------|
| `internal/pdf/signature/signature.go` | Signer implementation |
| `internal/pdf/signature/signature_test.go` | RSA + ECDSA tests |
| `certs/generate_certs.sh` | RSA CA chain + RSA leaf |
| `certs/ec_leaf.key`, `certs/ec_leaf.pem` | Sample ECDSA signer |
| `certs/leaf.key`, `certs/leaf.pem` | Sample RSA signer |
| `internal/models/models.go` | `SignatureConfig` schema (`privateKeyPem` supports RSA or ECDSA) |