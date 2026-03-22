//nolint:revive // package comment
package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

const (
	fileEncryptionKeyLength = 32
	passwordSaltLength      = 8
	passwordEntryLength     = 48
	permissionsEntryLength  = 16
	maxPasswordBytes        = 127
)

var randomReader io.Reader = rand.Reader

// PDFEncryption handles PDF document encryption using AES-256 (V=5, R=5).
type PDFEncryption struct {
	EncryptionKey        []byte
	UserPasswordHash     []byte
	OwnerPasswordHash    []byte
	UserEncryptedKey     []byte
	OwnerEncryptedKey    []byte
	EncryptedPermissions []byte
	Permissions          int32
	DocumentID           []byte
}

// NewPDFEncryption creates a new AES-256 encryption handler.
func NewPDFEncryption(config *models.SecurityConfig, documentID []byte) (*PDFEncryption, error) {
	if config == nil || config.OwnerPassword == "" {
		return nil, fmt.Errorf("owner password is required for encryption")
	}

	enc := &PDFEncryption{
		DocumentID: append([]byte(nil), documentID...),
	}
	enc.Permissions = enc.calculatePermissions(config)

	fileKey, err := readRandomBytes(fileEncryptionKeyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file encryption key: %w", err)
	}
	enc.EncryptionKey = fileKey

	userValidationSalt, err := readRandomBytes(passwordSaltLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user validation salt: %w", err)
	}
	userKeySalt, err := readRandomBytes(passwordSaltLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user key salt: %w", err)
	}
	enc.UserPasswordHash = enc.userHash(config.UserPassword, userValidationSalt, userKeySalt)
	enc.UserEncryptedKey, err = enc.userEncryptedKey(config.UserPassword, userKeySalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive user encrypted key: %w", err)
	}

	ownerValidationSalt, err := readRandomBytes(passwordSaltLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate owner validation salt: %w", err)
	}
	ownerKeySalt, err := readRandomBytes(passwordSaltLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate owner key salt: %w", err)
	}
	enc.OwnerPasswordHash = enc.ownerHash(config.OwnerPassword, ownerValidationSalt, ownerKeySalt, enc.UserPasswordHash)
	enc.OwnerEncryptedKey, err = enc.ownerEncryptedKey(config.OwnerPassword, ownerKeySalt, enc.UserPasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to derive owner encrypted key: %w", err)
	}

	enc.EncryptedPermissions, err = enc.permissionsHash()
	if err != nil {
		return nil, fmt.Errorf("failed to derive encrypted permissions: %w", err)
	}

	return enc, nil
}

func readRandomBytes(length int) ([]byte, error) {
	buffer := make([]byte, length)
	if _, err := io.ReadFull(randomReader, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

func normalizePassword(password string) []byte {
	buffer := []byte(password)
	if len(buffer) > maxPasswordBytes {
		buffer = buffer[:maxPasswordBytes]
	}
	return append([]byte(nil), buffer...)
}

func sha256Digest(parts ...[]byte) []byte {
	hasher := sha256.New()
	for _, part := range parts {
		_, _ = hasher.Write(part)
	}
	return hasher.Sum(nil)
}

func zeroIV() []byte {
	return make([]byte, aes.BlockSize)
}

func aesCBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("iv length must be %d bytes", aes.BlockSize)
	}
	if len(plaintext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("plaintext length must be a multiple of %d", aes.BlockSize)
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

func (enc *PDFEncryption) userHash(userPassword string, validationSalt, keySalt []byte) []byte {
	password := normalizePassword(userPassword)
	digest := sha256Digest(password, validationSalt)

	result := make([]byte, 0, passwordEntryLength)
	result = append(result, digest...)
	result = append(result, validationSalt...)
	result = append(result, keySalt...)
	return result
}

func (enc *PDFEncryption) userEncryptedKey(userPassword string, keySalt []byte) ([]byte, error) {
	key := sha256Digest(normalizePassword(userPassword), keySalt)
	return aesCBCEncrypt(key, zeroIV(), enc.EncryptionKey)
}

func (enc *PDFEncryption) ownerHash(ownerPassword string, validationSalt, keySalt, userHash []byte) []byte {
	password := normalizePassword(ownerPassword)
	digest := sha256Digest(password, validationSalt, userHash)

	result := make([]byte, 0, passwordEntryLength)
	result = append(result, digest...)
	result = append(result, validationSalt...)
	result = append(result, keySalt...)
	return result
}

func (enc *PDFEncryption) ownerEncryptedKey(ownerPassword string, keySalt, userHash []byte) ([]byte, error) {
	key := sha256Digest(normalizePassword(ownerPassword), keySalt, userHash)
	return aesCBCEncrypt(key, zeroIV(), enc.EncryptionKey)
}

func (enc *PDFEncryption) permissionsHash() ([]byte, error) {
	block := make([]byte, permissionsEntryLength)
	writePermsLE(block[:4], enc.Permissions)
	for index := 4; index < 8; index++ {
		block[index] = 0xff
	}
	block[8] = 'T'
	copy(block[9:12], []byte("adb"))

	randomTail, err := readRandomBytes(4)
	if err != nil {
		return nil, err
	}
	copy(block[12:], randomTail)

	cipherBlock, err := aes.NewCipher(enc.EncryptionKey)
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, permissionsEntryLength)
	cipherBlock.Encrypt(encrypted, block)
	return encrypted, nil
}

func writePermsLE(dst []byte, permissions int32) {
	dst[0] = byte(permissions)
	dst[1] = byte(permissions >> 8)
	dst[2] = byte(permissions >> 16)
	dst[3] = byte(permissions >> 24)
}

func (enc *PDFEncryption) calculatePermissions(config *models.SecurityConfig) int32 {
	var permissions int32 = -4

	if !config.AllowPrinting {
		permissions &= ^int32(4)
	}
	if !config.AllowModifying {
		permissions &= ^int32(8)
	}
	if !config.AllowCopying {
		permissions &= ^int32(16)
	}
	if !config.AllowAnnotations {
		permissions &= ^int32(32)
	}
	if !config.AllowFormFilling {
		permissions &= ^int32(256)
	}
	if !config.AllowAccessibility {
		permissions &= ^int32(512)
	}
	if !config.AllowAssembly {
		permissions &= ^int32(1024)
	}
	if !config.AllowHighQualityPrint {
		permissions &= ^int32(2048)
	}

	return permissions
}

// EncryptStream encrypts a PDF stream using AES-256-CBC.
func (enc *PDFEncryption) EncryptStream(data []byte, objNum, genNum int) []byte {
	key := enc.objKey(objNum, genNum)

	iv, err := readRandomBytes(aes.BlockSize)
	if err != nil {
		return data
	}

	padded := Pkcs7Pad(data, aes.BlockSize)
	ciphertext, err := aesCBCEncrypt(key, iv, padded)
	if err != nil {
		return data
	}

	result := make([]byte, 0, len(iv)+len(ciphertext))
	result = append(result, iv...)
	result = append(result, ciphertext...)
	return result
}

// EncryptString encrypts a PDF string.
func (enc *PDFEncryption) EncryptString(data []byte, objNum, genNum int) []byte {
	return enc.EncryptStream(data, objNum, genNum)
}

func (enc *PDFEncryption) objKey(_, _ int) []byte {
	return append([]byte(nil), enc.EncryptionKey...)
}

// Pkcs7Pad pads data to block size using PKCS#7.
func Pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// GetEncryptDictionary returns the /Encrypt dictionary content.
func (enc *PDFEncryption) GetEncryptDictionary(_ int) string {
	var dict strings.Builder

	dict.WriteString("<< /Type /Encrypt")
	dict.WriteString(" /Filter /Standard")
	dict.WriteString(" /V 5")
	dict.WriteString(" /R 5")
	dict.WriteString(" /Length 256")
	dict.WriteString(" /StmF /StdCF")
	dict.WriteString(" /StrF /StdCF")
	dict.WriteString(" /CF << /StdCF << /Type /CryptFilter /CFM /AESV3 /Length 32 /AuthEvent /DocOpen >> >>")
	dict.WriteString(fmt.Sprintf(" /P %d", enc.Permissions))
	dict.WriteString(fmt.Sprintf(" /U <%s>", hex.EncodeToString(enc.UserPasswordHash)))
	dict.WriteString(fmt.Sprintf(" /O <%s>", hex.EncodeToString(enc.OwnerPasswordHash)))
	dict.WriteString(fmt.Sprintf(" /UE <%s>", hex.EncodeToString(enc.UserEncryptedKey)))
	dict.WriteString(fmt.Sprintf(" /OE <%s>", hex.EncodeToString(enc.OwnerEncryptedKey)))
	dict.WriteString(fmt.Sprintf(" /Perms <%s>", hex.EncodeToString(enc.EncryptedPermissions)))
	dict.WriteString(" /EncryptMetadata true")
	dict.WriteString(" >>")

	return dict.String()
}

func GenerateDocumentID(data []byte) []byte {
	hasher := sha256.New()
	_, _ = hasher.Write(data)

	randomBytes := make([]byte, 16)
	if _, err := io.ReadFull(randomReader, randomBytes); err == nil {
		_, _ = hasher.Write(randomBytes)
	}

	sum := hasher.Sum(nil)
	return append([]byte(nil), sum[:16]...)
}

// FormatDocumentID formats the document ID for PDF trailer.
func FormatDocumentID(id []byte) string {
	hexID := hex.EncodeToString(id)
	return fmt.Sprintf("[<%s> <%s>]", hexID, hexID)
}
