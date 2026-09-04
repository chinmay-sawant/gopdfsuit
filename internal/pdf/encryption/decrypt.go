//nolint:revive // package comment lives on encrypt.go
package encryption

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PasswordPadding is the PDF spec Algorithm 2 padding string (32 bytes).
var PasswordPadding = []byte{
	0x28, 0xbf, 0x4e, 0x5e, 0x4e, 0x75, 0x8a, 0x41,
	0x64, 0x00, 0x4e, 0x56, 0xff, 0xfa, 0x01, 0x08,
	0x2e, 0x2e, 0x00, 0xb6, 0xd0, 0x68, 0x3e, 0x80,
	0x2f, 0x0c, 0xa9, 0xfe, 0x64, 0x53, 0x69, 0x7a,
}

// ObjectView is the minimal decrypted-object view over a parsed PDF.
// pdfobj.ParsedPDF can satisfy it later without either package importing
// the other: Body/Gen/ForEach line up with its object map.
type ObjectView interface {
	ObjectBody(num int) ([]byte, bool)
	ObjectGen(num int) int
	ForEachObject(fn func(num, gen int, body []byte))
}

// StandardDict is a parsed Standard-security-handler /Encrypt dictionary
// (V 2/4, R 2-4, RC4 path). AES dictionaries parse but do not decrypt here.
type StandardDict struct {
	R               int
	P               int
	LengthBits      int
	O               []byte
	U               []byte
	EncryptMetadata bool
	UsesAES         bool
}

// ParseEncryptRefAndID extracts the /Encrypt object reference and the first
// trailer /ID from the PDF bytes.
func ParseEncryptRefAndID(pdfBytes []byte) (encNum int, encGen int, id []byte, err error) {
	trailers := regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`).FindAllSubmatch(pdfBytes, -1)
	if len(trailers) == 0 {
		return 0, 0, nil, errors.New("missing trailer")
	}
	tr := trailers[len(trailers)-1][1]
	re := regexp.MustCompile(`/Encrypt\s+(\d+)\s+(\d+)\s+R`)
	m := re.FindSubmatch(tr)
	if m == nil {
		return 0, 0, nil, errors.New("trailer has no /Encrypt reference")
	}
	encNum, _ = strconv.Atoi(string(m[1]))
	encGen, _ = strconv.Atoi(string(m[2]))
	id = ParseFirstID(tr)
	if len(id) == 0 {
		id = ParseFirstID(pdfBytes)
	}
	if len(id) == 0 {
		return 0, 0, nil, errors.New("missing trailer /ID for encrypted PDF")
	}
	return encNum, encGen, id, nil
}

// ParseFirstID decodes the first hex string of an /ID array.
func ParseFirstID(b []byte) []byte {
	re := regexp.MustCompile(`/ID\s*\[\s*<([0-9A-Fa-f\s]+)>`)
	m := re.FindSubmatch(b)
	if m == nil {
		return nil
	}
	h := strings.ReplaceAll(string(m[1]), " ", "")
	h = strings.ReplaceAll(h, "\n", "")
	h = strings.ReplaceAll(h, "\r", "")
	id, err := hex.DecodeString(h)
	if err != nil {
		return nil
	}
	return id
}

// ParseStandardDict parses an /Encrypt object body into a StandardDict.
func ParseStandardDict(body []byte) (StandardDict, error) {
	if !bytes.Contains(body, []byte(`/Filter /Standard`)) && !bytes.Contains(body, []byte(`/Filter/Standard`)) {
		return StandardDict{}, errors.New("only Standard security handler is supported")
	}

	r := parseIntField(body, `/R\s+(-?\d+)`, 0)
	p := parseIntField(body, `/P\s+(-?\d+)`, 0)
	lengthBits := parseIntField(body, `/Length\s+(\d+)`, 40)
	if r <= 0 {
		return StandardDict{}, errors.New("invalid /R value in Encrypt dictionary")
	}
	o := parseHexOrLiteralField(body, "O")
	u := parseHexOrLiteralField(body, "U")
	if len(o) == 0 || len(u) == 0 {
		return StandardDict{}, errors.New("missing O/U entries in Encrypt dictionary")
	}

	encryptMetadata := !bytes.Contains(body, []byte(`/EncryptMetadata false`)) && !bytes.Contains(body, []byte(`/EncryptMetadatafalse`))
	usesAES := bytes.Contains(body, []byte("/AESV2")) || bytes.Contains(body, []byte("/AESV3"))

	return StandardDict{
		R:               r,
		P:               p,
		LengthBits:      lengthBits,
		O:               o,
		U:               u,
		EncryptMetadata: encryptMetadata,
		UsesAES:         usesAES,
	}, nil
}

func parseIntField(b []byte, pattern string, def int) int {
	re := regexp.MustCompile(pattern)
	m := re.FindSubmatch(b)
	if m == nil {
		return def
	}
	v, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return def
	}
	return v
}

func parseHexOrLiteralField(b []byte, field string) []byte {
	hexRe := regexp.MustCompile(fmt.Sprintf(`/%s\s*<([0-9A-Fa-f\s]+)>`, regexp.QuoteMeta(field)))
	if m := hexRe.FindSubmatch(b); m != nil {
		h := strings.ReplaceAll(string(m[1]), " ", "")
		h = strings.ReplaceAll(h, "\n", "")
		h = strings.ReplaceAll(h, "\r", "")
		v, err := hex.DecodeString(h)
		if err == nil {
			return v
		}
	}
	litRe := regexp.MustCompile(fmt.Sprintf(`/%s\s*\(([^)]*)\)`, regexp.QuoteMeta(field)))
	if m := litRe.FindSubmatch(b); m != nil {
		return []byte(m[1])
	}
	return nil
}

// ResolveFileKey derives the file encryption key from a user or owner
// password, validating it against the /U entry.
func ResolveFileKey(password string, d StandardDict, id0 []byte) ([]byte, bool) {
	if k, ok := deriveAndValidateUserKey(password, d, id0); ok {
		return k, true
	}
	if ownerDerived := DeriveUserPasswordFromOwner(password, d); ownerDerived != "" {
		if k, ok := deriveAndValidateUserKey(ownerDerived, d, id0); ok {
			return k, true
		}
	}
	return nil, false
}

func deriveAndValidateUserKey(password string, d StandardDict, id0 []byte) ([]byte, bool) {
	fileKey := DeriveFileKey(password, d, id0)
	if len(fileKey) == 0 {
		return nil, false
	}
	if ValidateUserPassword(fileKey, d, id0) {
		return fileKey, true
	}
	return nil, false
}

// DeriveFileKey implements PDF spec Algorithm 2 (MD5-based file key).
func DeriveFileKey(password string, d StandardDict, id0 []byte) []byte {
	keyLen := d.LengthBits / 8
	if d.R == 2 {
		keyLen = 5
	}
	if keyLen < 5 {
		keyLen = 5
	}
	if keyLen > 16 {
		keyLen = 16
	}

	padded := padPassword(password)
	h := md5.New()
	h.Write(padded)
	h.Write(d.O)
	h.Write(int32LE(int32(d.P)))
	h.Write(id0)
	if d.R >= 4 && !d.EncryptMetadata {
		h.Write([]byte{0xff, 0xff, 0xff, 0xff})
	}
	sum := h.Sum(nil)
	if d.R >= 3 {
		for range 50 {
			x := md5.Sum(sum[:keyLen])
			sum = x[:]
		}
	}
	return append([]byte{}, sum[:keyLen]...)
}

// ValidateUserPassword checks a file key against the /U entry (Algorithm 5).
func ValidateUserPassword(fileKey []byte, d StandardDict, id0 []byte) bool {
	if d.R == 2 {
		exp := RC4Crypt(fileKey, PasswordPadding)
		return len(d.U) >= 32 && bytes.Equal(exp, d.U[:32])
	}
	h := md5.Sum(append(append([]byte{}, PasswordPadding...), id0...))
	tmp := h[:]
	tmp = RC4Crypt(fileKey, tmp)
	for i := 1; i <= 19; i++ {
		k := XORKey(fileKey, byte(i))
		tmp = RC4Crypt(k, tmp)
	}
	if len(d.U) < 16 {
		return false
	}
	return bytes.Equal(tmp[:16], d.U[:16])
}

// DeriveUserPasswordFromOwner recovers the (possibly binary) user password
// from the owner password via the /O entry (Algorithm 3, reversed).
func DeriveUserPasswordFromOwner(ownerPassword string, d StandardDict) string {
	if d.R < 2 || len(d.O) == 0 {
		return ""
	}
	keyLen := d.LengthBits / 8
	if d.R == 2 {
		keyLen = 5
	}
	if keyLen < 5 {
		keyLen = 5
	}
	if keyLen > 16 {
		keyLen = 16
	}

	h := md5.Sum(padPassword(ownerPassword))
	k := h[:]
	if d.R >= 3 {
		for range 50 {
			x := md5.Sum(k[:keyLen])
			k = x[:]
		}
	}

	out := append([]byte{}, d.O...)
	if d.R == 2 {
		out = RC4Crypt(k[:keyLen], out)
	} else {
		for i := 19; i >= 0; i-- {
			ki := XORKey(k[:keyLen], byte(i))
			out = RC4Crypt(ki, out)
		}
	}
	// Per the PDF spec the decrypted /O value is the full 32-byte padded
	// user password. Never strip trailing 0x00 bytes: they may be part of
	// a binary user password.
	if len(out) > 32 {
		out = out[:32]
	}
	return string(out)
}

// DeriveObjectKey builds the per-object RC4 key from the file key.
func DeriveObjectKey(fileKey []byte, objNum, genNum int) []byte {
	b := make([]byte, 0, len(fileKey)+5)
	b = append(b, fileKey...)
	b = append(b, byte(objNum), byte(objNum>>8), byte(objNum>>16))
	b = append(b, byte(genNum), byte(genNum>>8))
	h := md5.Sum(b)
	kLen := min(len(fileKey)+5, 16)
	return h[:kLen]
}

// DecryptObjectStream RC4-decrypts the stream of one object body and fixes
// its /Length. It reports whether the body changed.
func DecryptObjectStream(objBody []byte, fileKey []byte, objNum, genNum int) ([]byte, bool) {
	streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	loc := streamRe.FindSubmatchIndex(objBody)
	if loc == nil {
		return objBody, false
	}
	raw := objBody[loc[2]:loc[3]]
	objKey := DeriveObjectKey(fileKey, objNum, genNum)
	dec := RC4Crypt(objKey, raw)

	out := make([]byte, 0, len(objBody))
	out = append(out, objBody[:loc[2]]...)
	out = append(out, dec...)
	out = append(out, objBody[loc[3]:]...)
	lenRe := regexp.MustCompile(`/Length\s+\d+`)
	out = lenRe.ReplaceAll(out, fmt.Appendf(nil, `/Length %d`, len(dec)))
	return out, true
}

// DecryptObjects decrypts every stream reachable through view with fileKey,
// returning the updated bodies keyed by object number.
func DecryptObjects(view ObjectView, fileKey []byte) map[int][]byte {
	updated := make(map[int][]byte)
	view.ForEachObject(func(num, gen int, body []byte) {
		if out, changed := DecryptObjectStream(body, fileKey, num, gen); changed {
			updated[num] = out
		}
	})
	return updated
}

func int32LE(v int32) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}

// RC4Crypt encrypts/decrypts with RC4.
func RC4Crypt(key, in []byte) []byte {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return append([]byte{}, in...)
	}
	out := make([]byte, len(in))
	c.XORKeyStream(out, in)
	return out
}

// XORKey returns key with every byte XORed by v (Algorithm 3 key stepping).
func XORKey(key []byte, v byte) []byte {
	out := make([]byte, len(key))
	for i := range key {
		out[i] = key[i] ^ v
	}
	return out
}
