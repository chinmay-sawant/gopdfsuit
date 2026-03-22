package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"io"
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

func TestNewAES256(t *testing.T) {
	config := &models.SecurityConfig{
		Enabled:               true,
		OwnerPassword:         "owner-secret",
		UserPassword:          "user-secret",
		AllowPrinting:         true,
		AllowModifying:        true,
		AllowCopying:          true,
		AllowAnnotations:      true,
		AllowFormFilling:      true,
		AllowAccessibility:    true,
		AllowAssembly:         true,
		AllowHighQualityPrint: true,
	}
	documentID := []byte{0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}

	fileKey := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	}
	userValidationSalt := []byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
	userKeySalt := []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}
	ownerValidationSalt := []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
	ownerKeySalt := []byte{0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58}
	permissionsTail := []byte{0x61, 0x62, 0x63, 0x64}

	randomBytes := append([]byte{}, fileKey...)
	randomBytes = append(randomBytes, userValidationSalt...)
	randomBytes = append(randomBytes, userKeySalt...)
	randomBytes = append(randomBytes, ownerValidationSalt...)
	randomBytes = append(randomBytes, ownerKeySalt...)
	randomBytes = append(randomBytes, permissionsTail...)
	setRandomReader(t, bytes.NewReader(randomBytes))

	enc, err := NewPDFEncryption(config, documentID)
	if err != nil {
		t.Fatalf("NewPDFEncryption returned error: %v", err)
	}

	wantUserHash := referenceUserHash(config.UserPassword, userValidationSalt, userKeySalt)
	wantUserEncryptedKey, err := referenceEncryptedKey(config.UserPassword, userKeySalt, nil, fileKey)
	if err != nil {
		t.Fatalf("referenceEncryptedKey user returned error: %v", err)
	}
	wantOwnerHash := referenceOwnerHash(config.OwnerPassword, ownerValidationSalt, ownerKeySalt, wantUserHash)
	wantOwnerEncryptedKey, err := referenceEncryptedKey(config.OwnerPassword, ownerKeySalt, wantUserHash, fileKey)
	if err != nil {
		t.Fatalf("referenceEncryptedKey owner returned error: %v", err)
	}
	wantPerms, err := referencePermissionsHash(fileKey, enc.Permissions, permissionsTail)
	if err != nil {
		t.Fatalf("referencePermissionsHash returned error: %v", err)
	}

	if !bytes.Equal(enc.EncryptionKey, fileKey) {
		t.Fatalf("file key mismatch\nwant: %x\n got: %x", fileKey, enc.EncryptionKey)
	}
	if !bytes.Equal(enc.UserPasswordHash, wantUserHash) {
		t.Fatalf("user hash mismatch\nwant: %x\n got: %x", wantUserHash, enc.UserPasswordHash)
	}
	if !bytes.Equal(enc.UserEncryptedKey, wantUserEncryptedKey) {
		t.Fatalf("user encrypted key mismatch\nwant: %x\n got: %x", wantUserEncryptedKey, enc.UserEncryptedKey)
	}
	if !bytes.Equal(enc.OwnerPasswordHash, wantOwnerHash) {
		t.Fatalf("owner hash mismatch\nwant: %x\n got: %x", wantOwnerHash, enc.OwnerPasswordHash)
	}
	if !bytes.Equal(enc.OwnerEncryptedKey, wantOwnerEncryptedKey) {
		t.Fatalf("owner encrypted key mismatch\nwant: %x\n got: %x", wantOwnerEncryptedKey, enc.OwnerEncryptedKey)
	}
	if !bytes.Equal(enc.EncryptedPermissions, wantPerms) {
		t.Fatalf("permissions mismatch\nwant: %x\n got: %x", wantPerms, enc.EncryptedPermissions)
	}

	dict := enc.GetEncryptDictionary(99)
	for _, want := range []string{"/V 5", "/R 5", "/Length 256", "/CFM /AESV3", "/UE <", "/OE <", "/Perms <"} {
		if !strings.Contains(dict, want) {
			t.Fatalf("encrypt dictionary missing %q: %s", want, dict)
		}
	}
}

func TestEncryptStreamAES256(t *testing.T) {
	enc := &PDFEncryption{
		EncryptionKey: bytes.Repeat([]byte{0x7a}, fileEncryptionKeyLength),
	}
	iv := []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80, 0x90, 0xa0, 0xb0, 0xc0, 0xd0, 0xe0, 0xf0}
	setRandomReader(t, bytes.NewReader(iv))

	plaintext := []byte("classified payload")
	got := enc.EncryptStream(plaintext, 27, 0)
	wantCiphertext, err := referenceCBCEncrypt(enc.EncryptionKey, iv, Pkcs7Pad(plaintext, aes.BlockSize))
	if err != nil {
		t.Fatalf("referenceCBCEncrypt returned error: %v", err)
	}
	want := append(append([]byte{}, iv...), wantCiphertext...)

	if !bytes.Equal(got, want) {
		t.Fatalf("encrypted stream mismatch\nwant: %x\n got: %x", want, got)
	}
	if !bytes.Equal(enc.objKey(27, 0), enc.EncryptionKey) {
		t.Fatalf("object key should reuse the 256-bit file key")
	}
}

func setRandomReader(t *testing.T, reader io.Reader) {
	t.Helper()
	previous := randomReader
	randomReader = reader
	t.Cleanup(func() {
		randomReader = previous
	})
}

func referenceUserHash(userPassword string, validationSalt, keySalt []byte) []byte {
	password := referenceNormalizePassword(userPassword)
	digest := referenceSHA256(password, validationSalt)

	result := make([]byte, 0, passwordEntryLength)
	result = append(result, digest...)
	result = append(result, validationSalt...)
	result = append(result, keySalt...)
	return result
}

func referenceOwnerHash(ownerPassword string, validationSalt, keySalt, userHash []byte) []byte {
	password := referenceNormalizePassword(ownerPassword)
	digest := referenceSHA256(password, validationSalt, userHash)

	result := make([]byte, 0, passwordEntryLength)
	result = append(result, digest...)
	result = append(result, validationSalt...)
	result = append(result, keySalt...)
	return result
}

func referenceEncryptedKey(password string, keySalt, userHash, fileKey []byte) ([]byte, error) {
	parts := [][]byte{referenceNormalizePassword(password), keySalt}
	if len(userHash) > 0 {
		parts = append(parts, userHash)
	}
	key := referenceSHA256(parts...)
	return referenceCBCEncrypt(key, make([]byte, aes.BlockSize), fileKey)
}

func referencePermissionsHash(fileKey []byte, permissions int32, randomTail []byte) ([]byte, error) {
	block := make([]byte, permissionsEntryLength)
	writePermsLE(block[:4], permissions)
	for index := 4; index < 8; index++ {
		block[index] = 0xff
	}
	block[8] = 'T'
	copy(block[9:12], []byte("adb"))
	copy(block[12:], randomTail)

	cipherBlock, err := aes.NewCipher(fileKey)
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, len(block))
	cipherBlock.Encrypt(encrypted, block)
	return encrypted, nil
}

func referenceNormalizePassword(password string) []byte {
	buffer := []byte(password)
	if len(buffer) > maxPasswordBytes {
		buffer = buffer[:maxPasswordBytes]
	}
	return append([]byte(nil), buffer...)
}

func referenceSHA256(parts ...[]byte) []byte {
	hasher := sha256.New()
	for _, part := range parts {
		_, _ = hasher.Write(part)
	}
	return hasher.Sum(nil)
}

func referenceCBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}
