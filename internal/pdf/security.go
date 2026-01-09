package pdf

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// SecurityHandler handles PDF encryption (PDF 2.0, AES-256, R=6)
type SecurityHandler struct {
	Enabled           bool
	UserPassword      string
	OwnerPassword     string
	FileEncryptionKey []byte // 32 bytes
	Permissions       int32

	// Encryption Dictionary Entries
	O, U, OE, UE []byte
	Perms        []byte
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(enabled bool, userPwd, ownerPwd string) *SecurityHandler {
	if !enabled {
		return &SecurityHandler{Enabled: false}
	}
	// Default permissions: Allow everything (-1) or restrict?
	// Usually for "password protection", we allow printing/copying if they have the password.
	// We'll default to -4 (0xFFFFFFFC) which filters top 2 reserved bits, essentially allowing all.
	return &SecurityHandler{
		Enabled:       true,
		UserPassword:  userPwd,
		OwnerPassword: ownerPwd,
		Permissions:   -4,
	}
}

// Prepare generates the encryption keys and dictionary entries (O, U, OE, UE, Perms)
func (sh *SecurityHandler) Prepare() error {
	if !sh.Enabled {
		return nil
	}

	// 1. Generate File Encryption Key (32 bytes)
	sh.FileEncryptionKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, sh.FileEncryptionKey); err != nil {
		return err
	}

	// 2. Compute U (User Validation Entry)
	var err error
	sh.U, sh.UE, err = sh.computeUserEntry()
	if err != nil {
		return err
	}

	// 3. Compute O (Owner Validation Entry)
	sh.O, sh.OE, err = sh.computeOwnerEntry()
	if err != nil {
		return err
	}

	// 4. Compute Perms
	sh.Perms, err = sh.computePerms()
	if err != nil {
		return err
	}

	return nil
}

// EncryptBytes encrypts data using AES-256-CBC with the File Encryption Key
// Output format: [IV (16 bytes)] [Encrypted Data]
func (sh *SecurityHandler) EncryptBytes(data []byte) ([]byte, error) {
	if !sh.Enabled {
		return data, nil
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Pad data (PKCS7)
	padding := aes.BlockSize - (len(data) % aes.BlockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	paddedData := append(data, padText...)

	// Encrypt
	block, err := aes.NewCipher(sh.FileEncryptionKey)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, paddedData)

	// Result = IV + CipherText
	result := append(iv, cipherText...)
	return result, nil
}

// EncryptString encrypts a string and returns it as a HEX PDF string <...>
func (sh *SecurityHandler) EncryptString(s string) (string, error) {
	if !sh.Enabled {
		return "(" + escapePDFString(s) + ")", nil
	}
	encrypted, err := sh.EncryptBytes([]byte(s))
	if err != nil {
		return "", err
	}
	return "<" + hex.EncodeToString(encrypted) + ">", nil
}

// Internal helpers for R=6 (ISO 32000-2)

func (sh *SecurityHandler) computeUserEntry() (u []byte, ue []byte, err error) {
	// U Entry: Validation Salt (8) + Key Salt (8) + Hash (32)
	validationSalt := make([]byte, 8)
	keySalt := make([]byte, 8)
	rand.Read(validationSalt)
	rand.Read(keySalt)

	// Hash for U check = SHA256(UTF8Trunc(Pwd) + ValidationSalt)
	pwdBytes := truncateUTF8(sh.UserPassword, 127)

	hashInput := append(pwdBytes, validationSalt...)
	hash := sha256.Sum256(hashInput)

	u = append(validationSalt, keySalt...)
	u = append(u, hash[:]...) // Total 48 bytes

	// UE Entry: AES-256-CBC(Key=SHA256(Pwd+KeySalt), IV=0, Data=FEK)
	keyInput := append(pwdBytes, keySalt...)
	encryptionKey := sha256.Sum256(keyInput)

	ue, err = aesEncryptZeroIV(encryptionKey[:], sh.FileEncryptionKey)
	return u, ue, err
}

func (sh *SecurityHandler) computeOwnerEntry() (o []byte, oe []byte, err error) {
	// O Entry: Validation Salt (8) + Key Salt (8) + Hash (32)
	validationSalt := make([]byte, 8)
	keySalt := make([]byte, 8)
	rand.Read(validationSalt)
	rand.Read(keySalt)

	// If no owner password, use user password (standard behavior)
	pwd := sh.OwnerPassword
	if pwd == "" {
		pwd = sh.UserPassword
	}
	pwdBytes := truncateUTF8(pwd, 127)

	hashInput := append(pwdBytes, validationSalt...)
	hash := sha256.Sum256(hashInput)

	o = append(validationSalt, keySalt...)
	o = append(o, hash[:]...)

	// OE Entry: AES-256-CBC(Key=SHA256(Pwd+KeySalt), IV=0, Data=FEK)
	keyInput := append(pwdBytes, keySalt...)
	encryptionKey := sha256.Sum256(keyInput)

	oe, err = aesEncryptZeroIV(encryptionKey[:], sh.FileEncryptionKey)
	return o, oe, err
}

func (sh *SecurityHandler) computePerms() ([]byte, error) {
	// Perms Block (16 bytes)
	// Byte 0-3: Permissions (int32 LE)
	// Byte 4-7: 0xFF 0xFF 0xFF 0xFF
	// Byte 8: 'T' (EncryptMetadata true)
	// Byte 9-11: 0, 0, 0 (or random/unused)
	// Byte 12-15: 0, 0, 0, 0

	permsBlock := make([]byte, 16)
	binary.LittleEndian.PutUint32(permsBlock[0:4], uint32(sh.Permissions))
	permsBlock[4] = 0xFF
	permsBlock[5] = 0xFF
	permsBlock[6] = 0xFF
	permsBlock[7] = 0xFF
	permsBlock[8] = 'T' // EncryptMetadata = true

	// Encrypt with FEK, IV=0
	return aesEncryptZeroIV(sh.FileEncryptionKey, permsBlock)
}

func aesEncryptZeroIV(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// AES-CBC with Zero IV
	iv := make([]byte, aes.BlockSize) // Zero IV

	// Data must be multiple of block size (all inputs here are 16 or 32 bytes)
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data length %d not multiple of block size", len(data))
	}

	cipherText := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, data)
	return cipherText, nil
}

func truncateUTF8(s string, limit int) []byte {
	b := []byte(s)
	if len(b) > limit {
		return b[:limit]
	}
	return b
}
