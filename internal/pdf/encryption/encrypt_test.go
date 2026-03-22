package encryption

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

func TestNewPDFEncryptionUsesRevision4Derivation(t *testing.T) {
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

	enc, err := NewPDFEncryption(config, documentID)
	if err != nil {
		t.Fatalf("NewPDFEncryption returned error: %v", err)
	}

	wantOwnerHash := referenceOwnerHash(config.UserPassword, config.OwnerPassword)
	wantFileKey := referenceFileKey(config.UserPassword, wantOwnerHash, enc.Permissions, documentID)
	wantUserHash := referenceUserHash(wantFileKey, documentID)
	wantObjectKey := referenceObjectKey(wantFileKey, 27, 0)

	if !bytes.Equal(enc.OwnerPasswordHash, wantOwnerHash) {
		t.Fatalf("owner hash mismatch\nwant: %x\n got: %x", wantOwnerHash, enc.OwnerPasswordHash)
	}
	if !bytes.Equal(enc.EncryptionKey, wantFileKey) {
		t.Fatalf("file key mismatch\nwant: %x\n got: %x", wantFileKey, enc.EncryptionKey)
	}
	if !bytes.Equal(enc.UserPasswordHash, wantUserHash) {
		t.Fatalf("user hash mismatch\nwant: %x\n got: %x", wantUserHash, enc.UserPasswordHash)
	}
	if !bytes.Equal(enc.objKey(27, 0), wantObjectKey) {
		t.Fatalf("object key mismatch\nwant: %x\n got: %x", wantObjectKey, enc.objKey(27, 0))
	}
}

func referenceOwnerHash(userPassword, ownerPassword string) []byte {
	hash := md5.Sum(referencePadPassword(ownerPassword))
	key := hash[:]
	for i := 0; i < 50; i++ {
		next := md5.Sum(key[:16])
		key = next[:]
	}

	result := append([]byte(nil), referencePadPassword(userPassword)...)
	for i := 0; i <= 19; i++ {
		result = referenceRC4Encrypt(referenceXORKey(key[:16], byte(i)), result)
	}

	return result
}

func referenceFileKey(userPassword string, ownerHash []byte, permissions int32, documentID []byte) []byte {
	hasher := md5.New()
	hasher.Write(referencePadPassword(userPassword))
	hasher.Write(ownerHash)
	hasher.Write([]byte{byte(permissions), byte(permissions >> 8), byte(permissions >> 16), byte(permissions >> 24)})
	hasher.Write(documentID)
	hash := hasher.Sum(nil)

	for i := 0; i < 50; i++ {
		next := md5.Sum(hash[:16])
		hash = next[:]
	}

	return append([]byte(nil), hash[:16]...)
}

func referenceUserHash(fileKey, documentID []byte) []byte {
	hasher := md5.New()
	hasher.Write(paddingBytes)
	hasher.Write(documentID)
	hash := hasher.Sum(nil)

	result := referenceRC4Encrypt(fileKey, hash)
	for i := 1; i <= 19; i++ {
		result = referenceRC4Encrypt(referenceXORKey(fileKey, byte(i)), result)
	}

	finalResult := make([]byte, 32)
	copy(finalResult, result)
	return finalResult
}

func referenceObjectKey(fileKey []byte, objNum, genNum int) []byte {
	hasher := md5.New()
	hasher.Write(fileKey)
	hasher.Write([]byte{byte(objNum), byte(objNum >> 8), byte(objNum >> 16), byte(genNum), byte(genNum >> 8)})
	hasher.Write([]byte("sAlT"))
	hash := hasher.Sum(nil)

	keyLength := len(fileKey) + 5
	if keyLength > 16 {
		keyLength = 16
	}

	return append([]byte(nil), hash[:keyLength]...)
}

func referencePadPassword(password string) []byte {
	pwd := []byte(password)
	if len(pwd) >= 32 {
		return append([]byte(nil), pwd[:32]...)
	}

	result := make([]byte, 32)
	copy(result, pwd)
	copy(result[len(pwd):], paddingBytes[:32-len(pwd)])
	return result
}

func referenceXORKey(key []byte, value byte) []byte {
	result := make([]byte, len(key))
	for i := range key {
		result[i] = key[i] ^ value
	}
	return result
}

func referenceRC4Encrypt(key, data []byte) []byte {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		panic(err)
	}

	result := make([]byte, len(data))
	cipher.XORKeyStream(result, data)
	return result
}
