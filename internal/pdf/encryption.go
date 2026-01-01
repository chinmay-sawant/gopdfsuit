package pdf

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// PDFEncryption handles PDF document encryption using AES-256
type PDFEncryption struct {
	EncryptionKey     []byte // 32-byte AES-256 key
	UserPasswordHash  []byte // /U value (48 bytes for AES-256)
	OwnerPasswordHash []byte // /O value (48 bytes for AES-256)
	UserKeyHash       []byte // /UE value (32 bytes - encrypted file encryption key)
	OwnerKeyHash      []byte // /OE value (32 bytes - encrypted file encryption key)
	Perms             []byte // /Perms value (16 bytes - encrypted permissions)
	Permissions       int32  // /P value (permission flags)
	DocumentID        []byte // First element of document ID array
}

// PDF encryption padding string (32 bytes)
var paddingBytes = []byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41,
	0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80,
	0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

// NewPDFEncryption creates a new encryption handler with AES-256
func NewPDFEncryption(config *models.SecurityConfig, documentID []byte) (*PDFEncryption, error) {
	if config == nil || config.OwnerPassword == "" {
		return nil, fmt.Errorf("owner password is required for encryption")
	}

	enc := &PDFEncryption{
		DocumentID: documentID,
	}

	// Calculate permissions from config
	enc.Permissions = enc.calculatePermissions(config)

	// Generate random file encryption key (32 bytes for AES-256)
	enc.EncryptionKey = make([]byte, 32)
	if _, err := rand.Read(enc.EncryptionKey); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Compute password hashes using PDF 2.0 / ISO 32000-2 algorithm
	// This implements the encryption algorithm for /V 5 /R 6

	// Generate random salts
	userValidationSalt := make([]byte, 8)
	userKeySalt := make([]byte, 8)
	ownerValidationSalt := make([]byte, 8)
	ownerKeySalt := make([]byte, 8)

	rand.Read(userValidationSalt)
	rand.Read(userKeySalt)
	rand.Read(ownerValidationSalt)
	rand.Read(ownerKeySalt)

	// Compute User password hash (/U value)
	// U = SHA-256(password || user validation salt) || user validation salt || user key salt
	userPwd := truncateOrPadPassword(config.UserPassword)

	userHash := sha256.New()
	userHash.Write(userPwd)
	userHash.Write(userValidationSalt)
	userValidation := userHash.Sum(nil)

	enc.UserPasswordHash = make([]byte, 48)
	copy(enc.UserPasswordHash[0:32], userValidation)
	copy(enc.UserPasswordHash[32:40], userValidationSalt)
	copy(enc.UserPasswordHash[40:48], userKeySalt)

	// Compute UE (encrypted file encryption key with user key)
	// UE = AES-256-CBC(user key, file encryption key)
	// user key = SHA-256(password || user key salt)
	userKeyHash := sha256.New()
	userKeyHash.Write(userPwd)
	userKeyHash.Write(userKeySalt)
	userKey := userKeyHash.Sum(nil)

	enc.UserKeyHash = aesEncryptCBC(userKey, enc.EncryptionKey)

	// Compute Owner password hash (/O value)
	// O = SHA-256(password || owner validation salt || U) || owner validation salt || owner key salt
	ownerPwd := truncateOrPadPassword(config.OwnerPassword)

	ownerHash := sha256.New()
	ownerHash.Write(ownerPwd)
	ownerHash.Write(ownerValidationSalt)
	ownerHash.Write(enc.UserPasswordHash)
	ownerValidation := ownerHash.Sum(nil)

	enc.OwnerPasswordHash = make([]byte, 48)
	copy(enc.OwnerPasswordHash[0:32], ownerValidation)
	copy(enc.OwnerPasswordHash[32:40], ownerValidationSalt)
	copy(enc.OwnerPasswordHash[40:48], ownerKeySalt)

	// Compute OE (encrypted file encryption key with owner key)
	// OE = AES-256-CBC(owner key, file encryption key)
	// owner key = SHA-256(password || owner key salt || U)
	ownerKeyHash := sha256.New()
	ownerKeyHash.Write(ownerPwd)
	ownerKeyHash.Write(ownerKeySalt)
	ownerKeyHash.Write(enc.UserPasswordHash)
	ownerKey := ownerKeyHash.Sum(nil)

	enc.OwnerKeyHash = aesEncryptCBC(ownerKey, enc.EncryptionKey)

	// Compute Perms (encrypted permissions)
	// Perms = AES-256-ECB(file encryption key, permissions block)
	enc.Perms = enc.computePermsValue()

	return enc, nil
}

// calculatePermissions calculates the /P value from security config
func (enc *PDFEncryption) calculatePermissions(config *models.SecurityConfig) int32 {
	// Start with all bits set that are required to be 1
	// Bits 1-2 are reserved (must be 0), bits 7-8 are reserved (must be 1)
	// Bits 13-32 are reserved (must be 1 for PDF 2.0)
	var p int32 = -1 // Start with all 1s (0xFFFFFFFF)

	// Clear permission bits based on config (set to 0 means NOT allowed)
	// Bit 3 (value 4): Print
	if !config.AllowPrinting {
		p &= ^int32(4)
	}
	// Bit 4 (value 8): Modify contents
	if !config.AllowModifying {
		p &= ^int32(8)
	}
	// Bit 5 (value 16): Copy/extract text and graphics
	if !config.AllowCopying {
		p &= ^int32(16)
	}
	// Bit 6 (value 32): Add/modify annotations
	if !config.AllowAnnotations {
		p &= ^int32(32)
	}
	// Bit 9 (value 256): Fill form fields (when bit 6 is clear)
	if !config.AllowFormFilling {
		p &= ^int32(256)
	}
	// Bit 10 (value 512): Accessibility (extract for disabilities)
	if !config.AllowAccessibility {
		p &= ^int32(512)
	}
	// Bit 11 (value 1024): Assemble document
	if !config.AllowAssembly {
		p &= ^int32(1024)
	}
	// Bit 12 (value 2048): Print high quality
	if !config.AllowHighQualityPrint {
		p &= ^int32(2048)
	}

	// Ensure required bits are set correctly
	// Bits 1 and 2 must be 0
	p &= ^int32(3)
	// Bits 7 and 8 must be 1
	p |= int32(192)

	return p
}

// computePermsValue creates the encrypted /Perms value
func (enc *PDFEncryption) computePermsValue() []byte {
	// Build 16-byte permissions block
	permsBlock := make([]byte, 16)

	// Bytes 0-3: P value (little-endian)
	permsBlock[0] = byte(enc.Permissions)
	permsBlock[1] = byte(enc.Permissions >> 8)
	permsBlock[2] = byte(enc.Permissions >> 16)
	permsBlock[3] = byte(enc.Permissions >> 24)

	// Bytes 4-7: 0xFFFFFFFF
	permsBlock[4] = 0xFF
	permsBlock[5] = 0xFF
	permsBlock[6] = 0xFF
	permsBlock[7] = 0xFF

	// Byte 8: 'T' if EncryptMetadata is true, 'F' otherwise
	permsBlock[8] = 'T' // We encrypt metadata

	// Byte 9: 'a'
	permsBlock[9] = 'a'

	// Byte 10: 'd'
	permsBlock[10] = 'd'

	// Byte 11: 'b'
	permsBlock[11] = 'b'

	// Bytes 12-15: random data
	rand.Read(permsBlock[12:16])

	// Encrypt with AES-256-ECB (single block, so ECB = CBC with zero IV effectively)
	return aesEncryptECB(enc.EncryptionKey, permsBlock)
}

// truncateOrPadPassword ensures password is properly formatted
// For PDF 2.0 AES-256, we use UTF-8 encoded passwords (max 127 bytes)
func truncateOrPadPassword(password string) []byte {
	pwd := []byte(password)
	// SASLprep normalization would be applied here in a full implementation
	// For simplicity, we just truncate to 127 bytes if needed
	if len(pwd) > 127 {
		pwd = pwd[:127]
	}
	return pwd
}

// aesEncryptCBC encrypts data using AES-256-CBC with zero IV
func aesEncryptCBC(key, plaintext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	// Pad plaintext to AES block size
	padded := pkcs7Pad(plaintext, aes.BlockSize)

	// Use zero IV for key encryption
	iv := make([]byte, aes.BlockSize)

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// Return first 32 bytes
	if len(ciphertext) > 32 {
		return ciphertext[:32]
	}
	return ciphertext
}

// aesEncryptECB encrypts a single block using AES-256-ECB
func aesEncryptECB(key, plaintext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	ciphertext := make([]byte, len(plaintext))
	block.Encrypt(ciphertext, plaintext)
	return ciphertext
}

// pkcs7Pad pads data to block size using PKCS#7
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// EncryptString encrypts a PDF string using AES-256-CBC
func (enc *PDFEncryption) EncryptString(data []byte, objNum, genNum int) []byte {
	// Compute object encryption key
	key := enc.computeObjectKey(objNum, genNum)

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	// Encrypt data
	block, err := aes.NewCipher(key)
	if err != nil {
		return data
	}

	padded := pkcs7Pad(data, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// Prepend IV to ciphertext
	return append(iv, ciphertext...)
}

// EncryptStream encrypts a PDF stream using AES-256-CBC
func (enc *PDFEncryption) EncryptStream(data []byte, objNum, genNum int) []byte {
	return enc.EncryptString(data, objNum, genNum)
}

// computeObjectKey computes the encryption key for a specific object
// For AES-256 (/V 5), the file encryption key is used directly
func (enc *PDFEncryption) computeObjectKey(objNum, genNum int) []byte {
	// For AES-256 (V=5, R=6), we use the file encryption key directly
	return enc.EncryptionKey
}

// GetEncryptDictionary returns the /Encrypt dictionary content
func (enc *PDFEncryption) GetEncryptDictionary(encryptObjID int) string {
	var dict strings.Builder

	dict.WriteString("<< /Type /Encrypt")
	dict.WriteString(" /Filter /Standard")
	dict.WriteString(" /V 5")        // AES-256
	dict.WriteString(" /R 6")        // PDF 2.0 revision
	dict.WriteString(" /Length 256") // Key length in bits

	// String/Stream filters
	dict.WriteString(" /StmF /StdCF")
	dict.WriteString(" /StrF /StdCF")

	// Crypt filters definition
	dict.WriteString(" /CF << /StdCF << /Type /CryptFilter /CFM /AESV3 /Length 32 >> >>")

	// Permission flags
	dict.WriteString(fmt.Sprintf(" /P %d", enc.Permissions))

	// Password hashes (hex encoded)
	dict.WriteString(fmt.Sprintf(" /U <%s>", hex.EncodeToString(enc.UserPasswordHash)))
	dict.WriteString(fmt.Sprintf(" /O <%s>", hex.EncodeToString(enc.OwnerPasswordHash)))
	dict.WriteString(fmt.Sprintf(" /UE <%s>", hex.EncodeToString(enc.UserKeyHash)))
	dict.WriteString(fmt.Sprintf(" /OE <%s>", hex.EncodeToString(enc.OwnerKeyHash)))
	dict.WriteString(fmt.Sprintf(" /Perms <%s>", hex.EncodeToString(enc.Perms)))

	// Encrypt metadata flag
	dict.WriteString(" /EncryptMetadata true")

	dict.WriteString(" >>")

	return dict.String()
}

// GenerateDocumentID generates a unique document ID
func GenerateDocumentID(data []byte) []byte {
	// Create MD5 hash of document content + timestamp
	hasher := md5.New()
	hasher.Write(data)

	// Add some randomness
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	hasher.Write(randomBytes)

	return hasher.Sum(nil)
}

// FormatDocumentID formats the document ID for PDF trailer
func FormatDocumentID(id []byte) string {
	hexID := hex.EncodeToString(id)
	return fmt.Sprintf("[<%s> <%s>]", hexID, hexID)
}
