package redact

import (
	"testing"
)

// TestTrailerHasEncryptIgnoresStreamText proves "/Encrypt" text inside a
// content stream no longer triggers a false encrypted-PDF signal.
func TestTrailerHasEncryptIgnoresStreamText(t *testing.T) {
	plain := []byte("%PDF-1.4\n1 0 obj\n<< /Length 20 >>\nstream\nBT (/Encrypt) Tj ET\nendstream\nendobj\ntrailer\n<< /Size 2 /Root 1 0 R >>")
	if trailerHasEncrypt(plain) {
		t.Fatalf("stream /Encrypt text falsely detected as encryption")
	}
	encrypted := []byte("trailer\n<< /Size 2 /Root 1 0 R /Encrypt 9 0 R >>")
	if !trailerHasEncrypt(encrypted) {
		t.Fatalf("real /Encrypt trailer entry was not detected")
	}
	inline := []byte("trailer\n<< /Size 2 /Root 1 0 R /Encrypt << /Filter /Standard >> >>")
	if !trailerHasEncrypt(inline) {
		t.Fatalf("inline /Encrypt dict was not detected")
	}
}
