package signature

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

func testLeafKeyPEM(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "leaf.key"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testLeafCertPEM(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "certs", "leaf.pem"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// craftedSignedPDF builds a minimal PDF containing the exact ByteRange and
// sized-Contents placeholders produced by CreateSignatureField.
func craftedSignedPDF(t *testing.T) []byte {
	t.Helper()
	var sigValue bytes.Buffer
	sigValue.WriteString("<< /Type /Sig /Filter /Adobe.PPKLite /SubFilter /adbe.pkcs7.detached ")
	sigValue.WriteString(sigByteRangePlaceholder)
	sigValue.WriteString(" /Contents <")
	for range sigContentsHexLen {
		sigValue.WriteByte('0')
	}
	sigValue.WriteString(">>>")
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-2.0\n1 0 obj\n")
	pdf.Write(sigValue.Bytes())
	pdf.WriteString("\nendobj\ntrailer\n<< /Root 1 0 R >>\n")
	return pdf.Bytes()
}

// TestDigestByteRangesOutOfRange asserts crafted out-of-range ByteRanges fail.
func TestDigestByteRangesOutOfRange(t *testing.T) {
	pdf := bytes.Repeat([]byte("A"), 200)
	for _, br := range [][4]int{
		{0, 100, 120, 500},   // tail past EOF
		{0, -1, 10, 10},      // negative offset
		{0, 150, 100, 10},    // second range starts before first ends
		{150, 100, 120, 10},  // end before start
		{0, 100, 120, 81},    // 120+81 > 200
		{0, 100, 300, 10},    // start past EOF
		{0, 100, 120, 10000}, // huge length
	} {
		if _, err := digestByteRanges(pdf, br); err == nil {
			t.Fatalf("expected error for ByteRange %v", br)
		}
	}
	if _, err := digestByteRanges(pdf, [4]int{0, 100, 120, 80}); err != nil {
		t.Fatalf("valid ByteRange rejected: %v", err)
	}
}

// TestLocateSignaturePlacementMissingMarker asserts fail-closed behavior when
// the exact sized marker is absent, even if a generic marker exists.
func TestLocateSignaturePlacementMissingMarker(t *testing.T) {
	// Generic small marker must NOT be accepted as a fallback.
	pdf := []byte("%PDF-2.0\n1 0 obj\n<< /Type /Sig " + sigByteRangePlaceholder +
		" /Contents <0123456789ABCDEF>>> \nendobj\n")
	if _, err := locateSignaturePlacement(pdf); err == nil {
		t.Fatal("expected error for missing exact sized-marker match")
	} else if !strings.Contains(err.Error(), "contents placeholder not found") {
		t.Fatalf("unexpected error: %v", err)
	}

	// No ByteRange placeholder at all.
	if _, err := locateSignaturePlacement([]byte("%PDF-2.0\ntrailer\n")); err == nil {
		t.Fatal("expected error for missing ByteRange placeholder")
	}

	// Well-formed crafted fixture locates successfully.
	if _, err := locateSignaturePlacement(craftedSignedPDF(t)); err != nil {
		t.Fatalf("valid crafted fixture rejected: %v", err)
	}
}

// TestEmbedSignatureInPlaceBounds asserts placement offsets past EOF fail.
func TestEmbedSignatureInPlaceBounds(t *testing.T) {
	pdf := craftedSignedPDF(t)
	sp, err := locateSignaturePlacement(pdf)
	if err != nil {
		t.Fatal(err)
	}
	bad := sp
	bad.contentsEnd = len(pdf) + 100
	if err := embedSignatureInPlace(pdf, &PDFSigner{}, bad); err == nil {
		t.Fatal("expected error for contents range past EOF")
	}
	bad = sp
	bad.byteRangePos = len(pdf) + 1
	if err := embedSignatureInPlace(pdf, &PDFSigner{}, bad); err == nil {
		t.Fatal("expected error for ByteRange position past EOF")
	}
	if err := embedSignatureInPlace(pdf, nil, sp); err == nil {
		t.Fatal("expected error for nil signer")
	}
}

// TestSignPDFFailureReleasesSlot proves a failing sign cannot leak a worker
// slot into deadlock: many failures followed by a real sign must succeed.
func TestSignPDFFailureReleasesSlot(t *testing.T) {
	ClearSignerCaches()
	bad := &PDFSigner{} // nil private key -> createPKCS7SignedData errors
	pdf := bytes.Repeat([]byte("A"), 200)
	br := [4]int{0, 100, 120, 80}
	slots := cap(signWorkerSlots)
	for i := 0; i < 3*slots; i++ {
		if _, err := bad.SignPDF(pdf, br); err == nil {
			t.Fatal("expected failing sign to error")
		}
	}

	signer := testSignerForSlot(t)
	done := make(chan error, 1)
	go func() {
		_, err := signer.SignPDF(pdf, br)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("successful sign after failures errored: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("successful sign deadlocked: worker slot leaked by failing signs")
	}
}

func testSignerForSlot(t *testing.T) *PDFSigner {
	t.Helper()
	signer, err := NewPDFSigner(&models.SignatureConfig{
		Enabled:        true,
		PrivateKeyPEM:  testLeafKeyPEM(t),
		CertificatePEM: testLeafCertPEM(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

// TestSignerCacheBounded inserts over the cap and asserts eviction keeps the
// cache bounded for both the signer and PEM-material caches.
func TestSignerCacheBounded(t *testing.T) {
	ClearSignerCaches()
	for i := 0; i < 3*signerCacheMaxEntries; i++ {
		key := fmt.Sprintf("key-%d", i)
		pdfSignerCache.store(key, &PDFSigner{})
		signerPEMMaterialCache.store(key, &parsedSignerPEMEntry{})
	}
	if n := pdfSignerCache.size(); n > signerCacheMaxEntries {
		t.Fatalf("signer cache unbounded: %d entries", n)
	}
	if n := signerPEMMaterialCache.size(); n > signerCacheMaxEntries {
		t.Fatalf("PEM material cache unbounded: %d entries", n)
	}
	ClearSignerCaches()
	if n := pdfSignerCache.size(); n != 0 {
		t.Fatalf("ClearSignerCaches did not empty signer cache: %d", n)
	}
}
