package main

import _ "embed"

// ECDSA P-256 (prime256v1) signer materials for the retail workload fast-sign path.

//go:embed certs/ec_leaf.pem
var ecLeafCertPEM string

//go:embed certs/ec_leaf.key
var ecLeafKeyPEM string
