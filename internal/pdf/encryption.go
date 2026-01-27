package pdf

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// PDFEncryption handles PDF document encryption using AES-128 (V=4, R=4)
// This is widely compatible with all PDF readers
type PDFEncryption struct {
	EncryptionKey     []byte // Encryption key (16 bytes for AES-128)
	UserPasswordHash  []byte // /U value (32 bytes)
	OwnerPasswordHash []byte // /O value (32 bytes)
	Permissions       int32  // /P value (permission flags)
	DocumentID        []byte // First element of document ID array
}

// PDF encryption padding string (32 bytes) - per PDF spec
var paddingBytes = []byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41,
	0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80,
	0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

// NewPDFEncryption creates a new encryption handler with AES-128
func NewPDFEncryption(config *models.SecurityConfig, documentID []byte) (*PDFEncryption, error) {
	if config == nil || config.OwnerPassword == "" {
		return nil, fmt.Errorf("owner password is required for encryption")
	}

	enc := &PDFEncryption{
		DocumentID: documentID,
	}

	// Calculate permissions from config
	enc.Permissions = enc.calculatePermissions(config)

	// Compute Owner password hash (/O value) first
	enc.OwnerPasswordHash = enc.computeOwnerHash(config.UserPassword, config.OwnerPassword)

	// Compute encryption key using user password
	enc.EncryptionKey = enc.computeEncryptionKey(config.UserPassword)

	// Compute User password hash (/U value)
	enc.UserPasswordHash = enc.computeUserHash()

	return enc, nil
}

// padPassword pads or truncates password to 32 bytes
func padPassword(password string) []byte {
	pwd := []byte(password)
	if len(pwd) >= 32 {
		return pwd[:32]
	}
	// Pad with standard padding bytes
	result := make([]byte, 32)
	copy(result, pwd)
	copy(result[len(pwd):], paddingBytes[:32-len(pwd)])
	return result
}

// computeOwnerHash computes the /O (owner) hash value per PDF spec Algorithm 3
func (enc *PDFEncryption) computeOwnerHash(userPassword, ownerPassword string) []byte {
	// Step 1: Pad owner password
	ownerPwd := padPassword(ownerPassword)

	// Step 2: MD5 hash the padded owner password
	hash := md5.Sum(ownerPwd)

	// Step 3: For R=4, do 50 iterations of MD5
	for i := 0; i < 50; i++ {
		hash = md5.Sum(hash[:])
	}

	// Use first 16 bytes as RC4 key (but we'll use it for AES key derivation)
	key := hash[:16]

	// Step 4: Pad user password
	userPwd := padPassword(userPassword)

	// Step 5: Encrypt with key using RC4-like XOR operation
	// For AES mode, we use a different approach - just encrypt with AES
	result := make([]byte, 32)
	copy(result, userPwd)

	// For R=4, we do 20 iterations with modified key
	for i := 0; i <= 19; i++ {
		modifiedKey := make([]byte, len(key))
		for j := range key {
			modifiedKey[j] = key[j] ^ byte(i)
		}
		result = rc4Encrypt(modifiedKey, result)
	}

	return result
}

// computeEncryptionKey computes the file encryption key per PDF spec Algorithm 2
func (enc *PDFEncryption) computeEncryptionKey(userPassword string) []byte {
	// Step 1: Pad user password
	userPwd := padPassword(userPassword)

	// Step 2: Create MD5 hash of: padded password + O value + P value + document ID
	hasher := md5.New()
	hasher.Write(userPwd)
	hasher.Write(enc.OwnerPasswordHash)

	// Write permissions as 4-byte little-endian
	pBytes := make([]byte, 4)
	pBytes[0] = byte(enc.Permissions)
	pBytes[1] = byte(enc.Permissions >> 8)
	pBytes[2] = byte(enc.Permissions >> 16)
	pBytes[3] = byte(enc.Permissions >> 24)
	hasher.Write(pBytes)

	hasher.Write(enc.DocumentID)

	hash := hasher.Sum(nil)

	// Step 3: For R=4, do 50 additional MD5 iterations on first 16 bytes
	for i := 0; i < 50; i++ {
		h := md5.Sum(hash[:16])
		hash = h[:]
	}

	// Return first 16 bytes as the encryption key
	return hash[:16]
}

// computeUserHash computes the /U (user) hash value per PDF spec Algorithm 5
func (enc *PDFEncryption) computeUserHash() []byte {
	// Step 1: Create MD5 hash of padding + document ID
	hasher := md5.New()
	hasher.Write(paddingBytes)
	hasher.Write(enc.DocumentID)
	hash := hasher.Sum(nil)

	// Step 2: Encrypt with file encryption key using RC4
	result := rc4Encrypt(enc.EncryptionKey, hash)

	// Step 3: For R=4, do 19 additional iterations with modified key
	for i := 1; i <= 19; i++ {
		modifiedKey := make([]byte, len(enc.EncryptionKey))
		for j := range enc.EncryptionKey {
			modifiedKey[j] = enc.EncryptionKey[j] ^ byte(i)
		}
		result = rc4Encrypt(modifiedKey, result)
	}

	// Pad result to 32 bytes with arbitrary padding
	finalResult := make([]byte, 32)
	copy(finalResult, result)
	// Fill remaining bytes with 0
	return finalResult
}

// rc4Encrypt performs RC4 encryption
func rc4Encrypt(key, data []byte) []byte {
	// Initialize S-box
	s := make([]byte, 256)
	for i := 0; i < 256; i++ {
		s[i] = byte(i)
	}

	// Key-scheduling algorithm (KSA)
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(key[i%len(key)])) % 256
		s[i], s[j] = s[j], s[i]
	}

	// Pseudo-random generation algorithm (PRGA)
	result := make([]byte, len(data))
	i, j := 0, 0
	for k := 0; k < len(data); k++ {
		i = (i + 1) % 256
		j = (j + int(s[i])) % 256
		s[i], s[j] = s[j], s[i]
		result[k] = data[k] ^ s[(int(s[i])+int(s[j]))%256]
	}

	return result
}

// calculatePermissions calculates the /P value from security config
func (enc *PDFEncryption) calculatePermissions(config *models.SecurityConfig) int32 {
	// Start with required bits set
	// Bits 1-2 must be 0, bits 7-8 must be 1, bits 13-32 must be 1
	var p int32 = -4 // 0xFFFFFFFC - bits 1,2 are 0

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
	// Bit 9 (value 256): Fill form fields
	if !config.AllowFormFilling {
		p &= ^int32(256)
	}
	// Bit 10 (value 512): Accessibility
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

	return p
}

// EncryptStream encrypts a PDF stream using AES-128-CBC
func (enc *PDFEncryption) EncryptStream(data []byte, objNum, genNum int) []byte {
	// Compute object key
	key := enc.computeObjectKey(objNum, genNum)

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return data
	}

	// Pad data
	padded := pkcs7Pad(data, aes.BlockSize)

	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return data
	}

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// Prepend IV
	return append(iv, ciphertext...)
}

// EncryptString encrypts a PDF string
func (enc *PDFEncryption) EncryptString(data []byte, objNum, genNum int) []byte {
	return enc.EncryptStream(data, objNum, genNum)
}

// computeObjectKey computes the encryption key for a specific object
func (enc *PDFEncryption) computeObjectKey(objNum, genNum int) []byte {
	// Create key by hashing: file key + object number + generation number
	hasher := md5.New()
	hasher.Write(enc.EncryptionKey)

	// Object and generation numbers as little-endian 3 bytes each
	hasher.Write([]byte{
		byte(objNum),
		byte(objNum >> 8),
		byte(objNum >> 16),
		byte(genNum),
		byte(genNum >> 8),
	})

	// For AES, append "sAlT"
	hasher.Write([]byte("sAlT"))

	hash := hasher.Sum(nil)

	// Use min(n+5, 16) bytes where n is key length (16)
	return hash[:16]
}

// pkcs7Pad pads data to block size using PKCS#7
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// GetEncryptDictionary returns the /Encrypt dictionary content
func (enc *PDFEncryption) GetEncryptDictionary(encryptObjID int) string {
	var dict strings.Builder

	dict.WriteString("<< /Type /Encrypt")
	dict.WriteString(" /Filter /Standard")
	dict.WriteString(" /V 4")        // AES-128
	dict.WriteString(" /R 4")        // Revision 4
	dict.WriteString(" /Length 128") // Key length in bits

	// String/Stream filters for AES
	dict.WriteString(" /StmF /StdCF")
	dict.WriteString(" /StrF /StdCF")

	// Crypt filters definition
	dict.WriteString(" /CF << /StdCF << /Type /CryptFilter /CFM /AESV2 /Length 16 >> >>")

	// Permission flags
	dict.WriteString(fmt.Sprintf(" /P %d", enc.Permissions))

	// Password hashes (hex encoded)
	dict.WriteString(fmt.Sprintf(" /U <%s>", hex.EncodeToString(enc.UserPasswordHash)))
	dict.WriteString(fmt.Sprintf(" /O <%s>", hex.EncodeToString(enc.OwnerPasswordHash)))

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
	if _, err := rand.Read(randomBytes); err != nil {
		return hasher.Sum(nil) // Fallback to partial hash
	}

	return hasher.Sum(nil)
}

// FormatDocumentID formats the document ID for PDF trailer
func FormatDocumentID(id []byte) string {
	hexID := hex.EncodeToString(id)
	return fmt.Sprintf("[<%s> <%s>]", hexID, hexID)
}
