package encryption

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
		ki := XORKey(k[:keyLen], byte(i))
		out = RC4Crypt(ki, out)
	}
	return out
}

func fixtureDictBytes(t *testing.T, ownerPassword, userPassword string, id0 []byte) []byte {
	t.Helper()
	const keyLen = 16
	o := ownerEncryptUser(ownerPassword, userPassword, keyLen)
	d := StandardDict{R: 4, P: -4, LengthBits: 128, O: o, EncryptMetadata: true}
	fileKey := DeriveFileKey(userPassword, d, id0)
	h := md5.Sum(append(append([]byte{}, PasswordPadding...), id0...))
	tmp := h[:]
	tmp = RC4Crypt(fileKey, tmp)
	for i := 1; i <= 19; i++ {
		tmp = RC4Crypt(XORKey(fileKey, byte(i)), tmp)
	}
	u := make([]byte, 32)
	copy(u, tmp)
	d.U = u
	if !ValidateUserPassword(fileKey, d, id0) {
		t.Fatal("fixture setup failed: user file key does not validate")
	}
	return []byte(fmt.Sprintf(
		"<< /Filter /Standard /V 4 /R 4 /Length 128 /P -4 /O <%s> /U <%s> >>",
		hex.EncodeToString(o), hex.EncodeToString(u)))
}

// TestUnifiedDictParseAndDecrypt is the A4 round-trip: parse a synthetic
// /Encrypt dict through the unified path, resolve the file key from the
// user password, the owner password, and reject a wrong one, then decrypt
// a stream encrypted with the derived object key.
func TestUnifiedDictParseAndDecrypt(t *testing.T) {
	id0 := []byte("0123456789abcdef")
	user, owner := "user-secret", "owner-secret"

	d, err := ParseStandardDict(fixtureDictBytes(t, owner, user, id0))
	if err != nil {
		t.Fatalf("ParseStandardDict failed: %v", err)
	}
	if d.R != 4 || d.LengthBits != 128 || d.P != -4 || d.UsesAES {
		t.Fatalf("parsed dict mismatch: R=%d LengthBits=%d P=%d UsesAES=%v", d.R, d.LengthBits, d.P, d.UsesAES)
	}

	fileKey, ok := ResolveFileKey(user, d, id0)
	if !ok {
		t.Fatal("user password failed to resolve file key")
	}
	if _, ok := ResolveFileKey("wrong-password", d, id0); ok {
		t.Fatal("wrong password resolved a file key")
	}
	ownerKey, ok := ResolveFileKey(owner, d, id0)
	if !ok {
		t.Fatal("owner password failed to resolve file key")
	}
	if !bytes.Equal(ownerKey, fileKey) {
		t.Fatal("owner-derived file key differs from user-derived key")
	}

	plaintext := []byte("sensitive stream payload")
	objKey := DeriveObjectKey(fileKey, 7, 0)
	enc := RC4Crypt(objKey, plaintext)
	body := append([]byte("<< /Length 0 >>\nstream\n"), enc...)
	body = append(body, []byte("\nendstream")...)

	dec, changed := DecryptObjectStream(body, fileKey, 7, 0)
	if !changed {
		t.Fatal("DecryptObjectStream reported no change")
	}
	start, end := bytes.Index(dec, []byte("stream\n"))+len("stream\n"), bytes.Index(dec, []byte("\nendstream"))
	if start < 0 || end < 0 || !bytes.Equal(dec[start:end], plaintext) {
		t.Fatalf("decrypted stream mismatch: %q", dec)
	}
	if !bytes.Contains(dec, []byte(fmt.Sprintf("/Length %d", len(plaintext)))) {
		t.Fatalf("Length not fixed after decrypt: %q", dec)
	}

	// Wrong file key must not reproduce the plaintext.
	badKey := append([]byte(nil), fileKey...)
	badKey[0] ^= 0xFF
	decBad, _ := DecryptObjectStream(body, badKey, 7, 0)
	if bytes.Contains(decBad, plaintext) {
		t.Fatal("decryption with wrong key leaked plaintext")
	}
}

type stubView struct {
	nums []int
	gens map[int]int
	bods map[int][]byte
}

func (s stubView) ObjectBody(num int) ([]byte, bool) { b, ok := s.bods[num]; return b, ok }
func (s stubView) ObjectGen(num int) int             { return s.gens[num] }
func (s stubView) ForEachObject(fn func(num, gen int, body []byte)) {
	for _, n := range s.nums {
		fn(n, s.gens[n], s.bods[n])
	}
}

// TestDecryptObjectsViaView proves the ObjectView seam decrypts every
// stream in the view with the file key (pdfobj.ParsedPDF can satisfy this
// interface later without new imports).
func TestDecryptObjectsViaView(t *testing.T) {
	id0 := []byte("0123456789abcdef")
	user, owner := "u", "o"
	d, err := ParseStandardDict(fixtureDictBytes(t, owner, user, id0))
	if err != nil {
		t.Fatal(err)
	}
	fileKey, ok := ResolveFileKey(user, d, id0)
	if !ok {
		t.Fatal("no file key")
	}
	mkStream := func(num int, plain string) []byte {
		enc := RC4Crypt(DeriveObjectKey(fileKey, num, 0), []byte(plain))
		b := append([]byte("<< /Length 0 >>\nstream\n"), enc...)
		return append(b, []byte("\nendstream")...)
	}
	view := stubView{
		nums: []int{3, 5},
		gens: map[int]int{3: 0, 5: 0},
		bods: map[int][]byte{3: mkStream(3, "page-three"), 5: mkStream(5, "page-five")},
	}
	updated := DecryptObjects(view, fileKey)
	if len(updated) != 2 {
		t.Fatalf("expected 2 decrypted objects, got %d", len(updated))
	}
	if !bytes.Contains(updated[3], []byte("page-three")) || !bytes.Contains(updated[5], []byte("page-five")) {
		t.Fatalf("view decrypt mismatch: %q %q", updated[3], updated[5])
	}
}
