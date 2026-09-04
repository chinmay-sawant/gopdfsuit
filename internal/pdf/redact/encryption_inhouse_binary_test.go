package redact

import (
	"bytes"
	"crypto/md5"
	"strings"
	"testing"
)

// ownerEncryptUser replicates PDF spec Algorithm 3 to build a valid /O value
// for a fixture owner/user password pair (R>=3 path).
func ownerEncryptUser(ownerPassword, userPassword string, keyLen int) []byte {
	h := md5.Sum(padPassword(ownerPassword))
	k := h[:]
	for range 50 {
		x := md5.Sum(k[:keyLen])
		k = x[:]
	}
	out := append([]byte{}, padPassword(userPassword)...)
	for i := 19; i >= 0; i-- {
		ki := xorKey(k[:keyLen], byte(i))
		out = rc4Crypt(ki, out)
	}
	return out
}

// fixtureDict builds an Encrypt dict whose /O corresponds to the given
// passwords and whose /U matches the user-password file key.
func fixtureDict(t *testing.T, ownerPassword, userPassword string, id0 []byte) standardEncryptDict {
	t.Helper()
	const keyLen = 16
	d := standardEncryptDict{
		R:               4,
		P:               -4,
		LengthBits:      128,
		O:               ownerEncryptUser(ownerPassword, userPassword, keyLen),
		EncryptMetadata: true,
	}
	// Compute the matching /U via the user-password file key (Algorithm 5).
	fileKey := deriveFileKey(userPassword, d, id0)
	h := md5.Sum(append(append([]byte{}, pdfPasswordPadding...), id0...))
	tmp := h[:]
	tmp = rc4Crypt(fileKey, tmp)
	for i := 1; i <= 19; i++ {
		tmp = rc4Crypt(xorKey(fileKey, byte(i)), tmp)
	}
	u := make([]byte, 32)
	copy(u, tmp)
	d.U = u
	if !validateUserPassword(fileKey, d, id0) {
		t.Fatal("fixture setup failed: user file key does not validate")
	}
	return d
}

// TestOwnerDerivationBinaryUserPassword proves that deriving the user password
// from the owner password keeps trailing 0x00 bytes, so a binary (NUL
// containing/trailing) user password still decrypts via the owner password.
func TestOwnerDerivationBinaryUserPassword(t *testing.T) {
	id0 := []byte("0123456789abcdef")
	// 32-byte user password ending in 0x00: old code stripped it with
	// TrimRight and re-padded, producing the wrong file key.
	binaryUser := strings.Repeat("a", 31) + "\x00"
	owner := "owner-secret"

	d := fixtureDict(t, owner, binaryUser, id0)

	derived := deriveUserPasswordFromOwner(owner, d)
	if len(derived) != 32 {
		t.Fatalf("derived password length = %d, want 32", len(derived))
	}
	if !bytes.Equal([]byte(derived), []byte(binaryUser)) {
		t.Fatal("derived user password does not preserve trailing 0x00 byte")
	}

	want := deriveFileKey(binaryUser, d, id0)
	got, ok := resolveFileKeyFromPassword(owner, d, id0)
	if !ok {
		t.Fatal("owner password failed to resolve file key for binary user password")
	}
	if !bytes.Equal(got, want) {
		t.Fatal("file key via owner derivation differs from direct user-password key")
	}
}

// TestOwnerDerivationEmbeddedNUL is a regression check for passwords with NUL
// bytes in the middle, which must round-trip exactly as well.
func TestOwnerDerivationEmbeddedNUL(t *testing.T) {
	id0 := []byte("fedcba9876543210")
	user := "ab\x00cd\x00ef"
	owner := "owner-secret"

	d := fixtureDict(t, owner, user, id0)
	derived := deriveUserPasswordFromOwner(owner, d)
	if !bytes.Equal([]byte(derived), padPassword(user)) {
		t.Fatal("derived user password does not match padded binary password")
	}
	if _, ok := resolveFileKeyFromPassword(owner, d, id0); !ok {
		t.Fatal("owner password failed to resolve file key")
	}
}
