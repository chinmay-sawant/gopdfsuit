// Package redact provides functionality for PDF text and image redaction.
package redact

import (
	"errors"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/encryption"
)

// pdfPasswordPadding is the PDF spec password padding; owned by encryption.
var pdfPasswordPadding = encryption.PasswordPadding

// standardEncryptDict is the parsed /Encrypt dictionary; owned by encryption.
type standardEncryptDict = encryption.StandardDict

// padPassword pads or truncates password to 32 bytes
func padPassword(password string) []byte {
	pwd := []byte(password)
	if len(pwd) >= 32 {
		return pwd[:32]
	}
	// Pad with standard padding bytes
	result := make([]byte, 32)
	copy(result, pwd)
	copy(result[len(pwd):], pdfPasswordPadding[:32-len(pwd)])
	return result
}

func decryptEncryptedPDFBytes(pdfBytes []byte, password string) ([]byte, error) {
	if !trailerHasEncrypt(pdfBytes) {
		return pdfBytes, nil
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("encrypted PDF detected; password is required")
	}

	objMap, objGen, err := buildObjectMap(pdfBytes)
	if err != nil {
		return nil, err
	}

	encNum, _, id0, err := parseEncryptRefAndID(pdfBytes)
	if err != nil {
		return nil, err
	}
	encBody, ok := objMap[encNum]
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

	for objNum, body := range objMap {
		genNum := objGenNum(objGen, objNum)
		updated, changed := decryptObjectStreams(body, fileKey, objNum, genNum)
		if changed {
			objMap[objNum] = updated
		}
	}

	// Rebuild output as decrypted PDF (no /Encrypt entry in trailer).
	return rebuildPDF(objMap, objGen, pdfBytes)
}

func parseEncryptRefAndID(pdfBytes []byte) (encNum int, encGen int, id []byte, err error) {
	return encryption.ParseEncryptRefAndID(pdfBytes)
}

func parseStandardEncryptDict(body []byte) (standardEncryptDict, error) {
	return encryption.ParseStandardDict(body)
}

func resolveFileKeyFromPassword(password string, d standardEncryptDict, id0 []byte) ([]byte, bool) {
	return encryption.ResolveFileKey(password, d, id0)
}

func deriveFileKey(password string, d standardEncryptDict, id0 []byte) []byte {
	return encryption.DeriveFileKey(password, d, id0)
}

func validateUserPassword(fileKey []byte, d standardEncryptDict, id0 []byte) bool {
	return encryption.ValidateUserPassword(fileKey, d, id0)
}

func deriveUserPasswordFromOwner(ownerPassword string, d standardEncryptDict) string {
	return encryption.DeriveUserPasswordFromOwner(ownerPassword, d)
}

func decryptObjectStreams(objBody []byte, fileKey []byte, objNum, genNum int) ([]byte, bool) {
	return encryption.DecryptObjectStream(objBody, fileKey, objNum, genNum)
}

func rc4Crypt(key, in []byte) []byte {
	return encryption.RC4Crypt(key, in)
}

func xorKey(key []byte, v byte) []byte {
	return encryption.XORKey(key, v)
}
