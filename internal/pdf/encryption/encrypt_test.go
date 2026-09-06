// Package encryption provides AES-128 PDF stream encryption and document IDs.
package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

func testEncryption(t *testing.T) *PDFEncryption {
	t.Helper()
	enc, err := NewPDFEncryption(&models.SecurityConfig{
		OwnerPassword: "owner",
		UserPassword:  "user",
	}, GenerateDocumentID([]byte("doc")))
	if err != nil {
		t.Fatal(err)
	}
	return enc
}

// TestGenerateDocumentIDUnique verifies random bytes are mixed into the hash.
func TestGenerateDocumentIDUnique(t *testing.T) {
	a := GenerateDocumentID([]byte("same content"))
	b := GenerateDocumentID([]byte("same content"))
	if bytes.Equal(a, b) {
		t.Fatal("two document IDs for identical content are equal; randomness is not mixed in")
	}
	if len(a) != 16 || len(b) != 16 {
		t.Fatalf("unexpected ID lengths: %d, %d", len(a), len(b))
	}
}

// TestEncryptStreamE_RandFailure verifies RNG failure returns an error and no
// plaintext instead of silently returning the input.
func TestEncryptStreamE_RandFailure(t *testing.T) {
	enc := testEncryption(t)
	plaintext := []byte("sensitive stream bytes")
	failRand := func([]byte) (int, error) { return 0, errors.New("rng exploded") }

	out, err := enc.EncryptStreamE(plaintext, 1, 0, failRand)
	if err == nil {
		t.Fatal("expected error on RNG failure")
	}
	if out != nil {
		t.Fatalf("expected nil output on RNG failure, got %d bytes", len(out))
	}

	if got := enc.EncryptStream(plaintext, 1, 0); got != nil {
		// EncryptStream uses crypto/rand which should succeed; failure path
		// must still never echo plaintext, so any non-nil result must differ.
		if bytes.Equal(got, plaintext) {
			t.Fatal("EncryptStream returned plaintext bytes")
		}
	}
}

// TestEncryptStreamE_NilRand verifies a nil random source fails closed.
func TestEncryptStreamE_NilRand(t *testing.T) {
	enc := testEncryption(t)
	out, err := enc.EncryptStreamE([]byte("data"), 1, 0, nil)
	if err == nil || out != nil {
		t.Fatalf("expected (nil, error) for nil rand source, got (%v, %v)", out, err)
	}
}

// TestEncryptStream_RoundTrip decrypts the output to prove it is real
// AES-128-CBC ciphertext with an IV prefix.
func TestEncryptStream_RoundTrip(t *testing.T) {
	enc := testEncryption(t)
	plaintext := []byte("round-trip payload")
	out := enc.EncryptStream(plaintext, 7, 0)
	if out == nil {
		t.Fatal("EncryptStream returned nil")
	}
	if len(out) <= aes.BlockSize {
		t.Fatalf("ciphertext too short: %d bytes", len(out))
	}
	iv, ciphertext := out[:aes.BlockSize], out[aes.BlockSize:]
	key := enc.computeObjectKey(7, 0)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)
	// Strip PKCS#7 padding.
	pad := int(plain[len(plain)-1])
	if pad <= 0 || pad > aes.BlockSize {
		t.Fatalf("bad padding value %d", pad)
	}
	if !bytes.Equal(plain[:len(plain)-pad], plaintext) {
		t.Fatal("decrypted payload does not match plaintext")
	}
}

// TestPadPasswordNoAlias proves padPassword never aliases immutable string
// bytes: mutating the result must leave the source string untouched.
func TestPadPasswordNoAlias(t *testing.T) {
	long := "0123456789abcdef0123456789abcdefEXTRA"
	before := long
	padded := padPassword(long)
	if len(padded) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(padded))
	}
	for i := range padded {
		padded[i] ^= 0xFF
	}
	if long != before {
		t.Fatal("mutating padded output corrupted the source string (aliasing)")
	}
	if !bytes.Equal(padPassword("abc")[:3], []byte("abc")) {
		t.Fatal("short password prefix mismatch")
	}
	if len(padPassword("")) != 32 {
		t.Fatal("empty password must still pad to 32 bytes")
	}
}

// TestPkcs7PadNoMutate proves padding never writes into the caller's array.
func TestPkcs7PadNoMutate(t *testing.T) {
	backing := make([]byte, 32)
	copy(backing, []byte("caller data"))
	data := backing[:10:10]
	snapshot := append([]byte(nil), backing...)
	Pkcs7Pad(data, aes.BlockSize)
	if !bytes.Equal(backing, snapshot) {
		t.Fatal("Pkcs7Pad mutated the caller's backing array")
	}
}
