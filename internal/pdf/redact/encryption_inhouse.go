// Package redact provides functionality for PDF text and image redaction.
package redact

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var pdfPasswordPadding = []byte{
	0x28, 0xbf, 0x4e, 0x5e, 0x4e, 0x75, 0x8a, 0x41,
	0x64, 0x00, 0x4e, 0x56, 0xff, 0xfa, 0x01, 0x08,
	0x2e, 0x2e, 0x00, 0xb6, 0xd0, 0x68, 0x3e, 0x80,
	0x2f, 0x0c, 0xa9, 0xfe, 0x64, 0x53, 0x69, 0x7a,
}

// padPassword pads or truncates password to 32 bytes
func padPassword(password string) []byte {
	result := make([]byte, 32)
	n := copy(result, password)
	if n >= 32 {
		return result[:32]
	}
	copy(result[n:], pdfPasswordPadding[:32-n])
	return result
}

type standardEncryptDict struct {
	R               int
	P               int
	LengthBits      int
	O               []byte
	U               []byte
	EncryptMetadata bool
	UsesAES         bool
}

func decryptEncryptedPDFBytes(pdfBytes []byte, password string) ([]byte, error) {
	if !trailerHasEncrypt(pdfBytes) {
		return pdfBytes, nil
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("encrypted PDF detected; password is required")
	}

	objMap, err := buildObjectMap(pdfBytes)
	if err != nil {
		return nil, err
	}

	encRef, id0, err := parseEncryptRefAndID(pdfBytes)
	if err != nil {
		return nil, err
	}
	encBody, ok := objMap[encRef]
	if !ok {
		return nil, errors.New("encrypt object reference not found")
	}
	d, err := parseStandardEncryptDict(encBody)
	if err != nil {
		return nil, err
	}
	if d.UsesAES {
		return nil, errors.New("AES encrypted PDFs are not supported by in-house decryptor yet")
	}

	fileKey, ok := resolveFileKeyFromPassword(password, d, id0)
	if !ok {
		return nil, errors.New("invalid PDF password")
	}

	for key, body := range objMap {
		objNum, genNum, ok := parseObjectKey(key)
		if !ok {
			continue
		}
		updated, changed := decryptObjectStreams(body, fileKey, objNum, genNum)
		if changed {
			objMap[key] = updated
		}
	}

	// Rebuild output as decrypted PDF (no /Encrypt entry in trailer).
	return rebuildPDF(objMap, pdfBytes)
}

func parseEncryptRefAndID(pdfBytes []byte) (string, []byte, error) {
	trailers := reTrailerBlock.FindAllSubmatch(pdfBytes, -1)
	if len(trailers) == 0 {
		return "", nil, errors.New("missing trailer")
	}
	tr := trailers[len(trailers)-1][1]
	m := reEncryptRef.FindSubmatch(tr)
	if m == nil {
		return "", nil, errors.New("trailer has no /Encrypt reference")
	}
	encRef := string(m[1]) + " " + string(m[2])
	id := parseFirstID(tr)
	if len(id) == 0 {
		id = parseFirstID(pdfBytes)
	}
	if len(id) == 0 {
		return "", nil, errors.New("missing trailer /ID for encrypted PDF")
	}
	return encRef, id, nil
}

func parseFirstID(b []byte) []byte {
	// /ID [<hex1><hex2>] or /ID [<hex1> <hex2>]
	m := reTrailerIDHex.FindSubmatch(b)
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

func parseStandardEncryptDict(body []byte) (standardEncryptDict, error) {
	if bytesIndex(body, encryptFilterStd) < 0 && bytesIndex(body, encryptFilterStdAlt) < 0 {
		return standardEncryptDict{}, errors.New("only Standard security handler is supported")
	}

	r := parseIntField(body, `/R\s+(-?\d+)`, 0)
	p := parseIntField(body, `/P\s+(-?\d+)`, 0)
	lengthBits := parseIntField(body, `/Length\s+(\d+)`, 40)
	if r <= 0 {
		return standardEncryptDict{}, errors.New("invalid /R value in Encrypt dictionary")
	}
	o := parseHexOrLiteralField(body, "O")
	u := parseHexOrLiteralField(body, "U")
	if len(o) == 0 || len(u) == 0 {
		return standardEncryptDict{}, errors.New("missing O/U entries in Encrypt dictionary")
	}

	encryptMetadata := bytesIndex(body, encryptMetaFalse) < 0 && bytesIndex(body, encryptMetaFalseAlt) < 0
	usesAES := bytesIndex(body, aesV2Bytes) >= 0 || bytesIndex(body, aesV3Bytes) >= 0

	return standardEncryptDict{
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
	hexRe := regexp.MustCompile("/" + regexp.QuoteMeta(field) + `\s*<([0-9A-Fa-f\s]+)>`)
	if m := hexRe.FindSubmatch(b); m != nil {
		h := strings.ReplaceAll(string(m[1]), " ", "")
		h = strings.ReplaceAll(h, "\n", "")
		h = strings.ReplaceAll(h, "\r", "")
		v, err := hex.DecodeString(h)
		if err == nil {
			return v
		}
	}
	litRe := regexp.MustCompile("/" + regexp.QuoteMeta(field) + `\s*\(([^)]*)\)`)
	if m := litRe.FindSubmatch(b); m != nil {
		return []byte(m[1])
	}
	return nil
}

func resolveFileKeyFromPassword(password string, d standardEncryptDict, id0 []byte) ([]byte, bool) {
	if k, ok := deriveAndValidateUserKey(password, d, id0); ok {
		return k, true
	}
	if ownerDerived := deriveUserPasswordFromOwner(password, d); ownerDerived != "" {
		if k, ok := deriveAndValidateUserKey(ownerDerived, d, id0); ok {
			return k, true
		}
	}
	return nil, false
}

func deriveAndValidateUserKey(password string, d standardEncryptDict, id0 []byte) ([]byte, bool) {
	fileKey := deriveFileKey(password, d, id0)
	if len(fileKey) == 0 {
		return nil, false
	}
	if validateUserPassword(fileKey, d, id0) {
		return fileKey, true
	}
	return nil, false
}

func deriveFileKey(password string, d standardEncryptDict, id0 []byte) []byte {
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

	pad := padPassword(password)
	//nolint:gosec // PDF Standard Security Handler (ISO 32000-1 §7.6.3.3) requires MD5 for key derivation.
	h := md5.New()
	h.Write(pad)
	h.Write(d.O)
	h.Write(int32LEBytes(int32(d.P)))
	h.Write(id0)
	if d.R >= 4 && !d.EncryptMetadata {
		h.Write([]byte{0xff, 0xff, 0xff, 0xff})
	}
	sum := h.Sum(nil)
	if d.R >= 3 {
		for i := 0; i < 50; i++ {
			//nolint:gosec // ISO 32000-1 mandates iterated MD5 for file encryption key derivation.
			x := md5.Sum(sum[:keyLen])
			sum = x[:]
		}
	}
	return append([]byte{}, sum[:keyLen]...)
}

func validateUserPassword(fileKey []byte, d standardEncryptDict, id0 []byte) bool {
	if d.R == 2 {
		exp := rc4Crypt(fileKey, pdfPasswordPadding)
		if len(d.U) < 32 || len(exp) != 32 {
			return false
		}
		return bytes.Equal(exp, d.U[:32])
	}
	//nolint:gosec // PDF Standard Security Handler (ISO 32000-1 §7.6.3.3) requires MD5 for /U validation.
	uPad := make([]byte, len(pdfPasswordPadding)+len(id0))
	copy(uPad, pdfPasswordPadding)
	copy(uPad[len(pdfPasswordPadding):], id0)
	h := md5.Sum(uPad)
	tmp := h[:]
	tmp = rc4Crypt(fileKey, tmp)
	for i := 1; i <= 19; i++ {
		k := xorKey(fileKey, byte(i))
		tmp = rc4Crypt(k, tmp)
	}
	if len(d.U) < 16 {
		return false
	}
	if len(d.U) < 16 {
		return false
	}
	return bytes.Equal(tmp[:16], d.U[:16])
}

func deriveUserPasswordFromOwner(ownerPassword string, d standardEncryptDict) string {
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

	//nolint:gosec // PDF Standard Security Handler (ISO 32000-1 §7.6.3.3) requires MD5 for owner key derivation.
	h := md5.Sum(padPassword(ownerPassword))
	k := h[:]
	if d.R >= 3 {
		for i := 0; i < 50; i++ {
			//nolint:gosec // ISO 32000-1 mandates iterated MD5 for owner-password key derivation.
			x := md5.Sum(k[:keyLen])
			k = x[:]
		}
	}

	out := append([]byte{}, d.O...)
	if d.R == 2 {
		out = rc4Crypt(k[:keyLen], out)
	} else {
		for i := 19; i >= 0; i-- {
			ki := xorKey(k[:keyLen], byte(i))
			out = rc4Crypt(ki, out)
		}
	}
	out = bytes.TrimRight(out, string([]byte{0}))
	if len(out) > 32 {
		out = out[:32]
	}
	return string(out)
}

func decryptObjectStreams(objBody []byte, fileKey []byte, objNum, genNum int) ([]byte, bool) {
	loc := reDecryptStream.FindSubmatchIndex(objBody)
	if loc == nil {
		return objBody, false
	}
	raw := objBody[loc[2]:loc[3]]
	objKey := deriveObjectKey(fileKey, objNum, genNum)
	dec := rc4Crypt(objKey, raw)

	out := make([]byte, 0, len(objBody))
	out = append(out, objBody[:loc[2]]...)
	out = append(out, dec...)
	out = append(out, objBody[loc[3]:]...)
	var lenBuf []byte
	lenBuf = append(lenBuf, "/Length "...)
	lenBuf = strconv.AppendInt(lenBuf, int64(len(dec)), 10)
	out = reLengthReplace.ReplaceAll(out, lenBuf)
	return out, true
}

func deriveObjectKey(fileKey []byte, objNum, genNum int) []byte {
	b := make([]byte, 0, len(fileKey)+5)
	b = append(b, fileKey...)
	b = append(b, byte(objNum), byte(objNum>>8), byte(objNum>>16), byte(genNum), byte(genNum>>8))
	//nolint:gosec // PDF Standard Security Handler (ISO 32000-1 §7.6.3.3) requires MD5 for per-object keys.
	h := md5.Sum(b)
	kLen := len(fileKey) + 5
	if kLen > 16 {
		kLen = 16
	}
	return h[:kLen]
}



func int32LEBytes(v int32) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}

func rc4Crypt(key, in []byte) []byte {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return append([]byte{}, in...)
	}
	out := make([]byte, len(in))
	c.XORKeyStream(out, in)
	return out
}

func xorKey(key []byte, v byte) []byte {
	out := make([]byte, len(key))
	for i := range key {
		out[i] = key[i] ^ v
	}
	return out
}
