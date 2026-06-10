package signature

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/asn1"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

func oidEqual(a, b asn1.ObjectIdentifier) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNewPDFSigner_ECDSA_P256(t *testing.T) {
	keyPEM, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "ec_leaf.key"))
	if err != nil {
		t.Fatal(err)
	}
	certPEM, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "ec_leaf.pem"))
	if err != nil {
		t.Fatal(err)
	}

	signer, err := NewPDFSigner(&models.SignatureConfig{
		Enabled:        true,
		PrivateKeyPEM:  string(keyPEM),
		CertificatePEM: string(certPEM),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := signer.privateKey.(*ecdsa.PrivateKey); !ok {
		t.Fatalf("expected *ecdsa.PrivateKey, got %T", signer.privateKey)
	}
	if !oidEqual(signer.digestSigAlgorithm, oidECDSAWithSHA256) {
		t.Fatalf("expected ECDSA OID, got %v", signer.digestSigAlgorithm)
	}
}

func TestNewPDFSigner_RSA(t *testing.T) {
	keyPEM, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "leaf.key"))
	if err != nil {
		t.Fatal(err)
	}
	certPEM, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "leaf.pem"))
	if err != nil {
		t.Fatal(err)
	}

	signer, err := NewPDFSigner(&models.SignatureConfig{
		Enabled:        true,
		PrivateKeyPEM:  string(keyPEM),
		CertificatePEM: string(certPEM),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := signer.privateKey.(*rsa.PrivateKey); !ok {
		t.Fatalf("expected *rsa.PrivateKey, got %T", signer.privateKey)
	}
	if !oidEqual(signer.digestSigAlgorithm, oidRSAEncryption) {
		t.Fatalf("expected RSA OID, got %v", signer.digestSigAlgorithm)
	}
}

func TestSignPDF_ECDSA_PKCS7(t *testing.T) {
	keyPEM, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "ec_leaf.key"))
	if err != nil {
		t.Fatal(err)
	}
	certPEM, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "ec_leaf.pem"))
	if err != nil {
		t.Fatal(err)
	}

	signer, err := NewPDFSigner(&models.SignatureConfig{
		Enabled:        true,
		PrivateKeyPEM:  string(keyPEM),
		CertificatePEM: string(certPEM),
	})
	if err != nil {
		t.Fatal(err)
	}

	pdf := bytes.Repeat([]byte("A"), 200)
	byteRange := [4]int{0, 100, 120, len(pdf) - 120}
	sig, err := signer.SignPDF(pdf, byteRange)
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) == 0 {
		t.Fatal("empty PKCS#7 signature")
	}
}