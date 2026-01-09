package pdf

import (
	"testing"
)

func TestSecurityHandler_Prepare(t *testing.T) {
	sh := NewSecurityHandler(true, "user123", "owner456")
	if err := sh.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if len(sh.FileEncryptionKey) != 32 {
		t.Errorf("Expected FEK length 32, got %d", len(sh.FileEncryptionKey))
	}
	if len(sh.U) != 48 {
		t.Errorf("Expected U length 48, got %d", len(sh.U))
	}
	if len(sh.O) != 48 {
		t.Errorf("Expected O length 48, got %d", len(sh.O))
	}
	// UE and OE are 32 bytes (AES-256 block aligned for 32 byte key? 32 bytes is 2 blocks. 32 bytes plain + padding?
	// Wait, FEK is 32 bytes.
	// In AES-CBC with padding (PKCS7), 32 bytes needs 1 full block of padding if input is multiple of block size?
	// Ah, usually PKCS7 adds a full block of padding if size % blocksize == 0.
	// But in my implementation `aesEncryptZeroIV`:
	// `if len(data)%aes.BlockSize != 0` -> returns error.
	// It assumes input is already aligned.
	// In `computeUserEntry`, I encrypt `sh.FileEncryptionKey` (32 bytes). 32 % 16 == 0.
	// So `ue` size should be 32 bytes (no padding added by `aesEncryptZeroIV`, just encrypts blocks).
	// Checked implementation: `mode.CryptBlocks`. Yes, size remains same.
	if len(sh.UE) != 32 {
		t.Errorf("Expected UE length 32, got %d", len(sh.UE))
	}
}

func TestSecurityHandler_EncryptBytes(t *testing.T) {
	sh := NewSecurityHandler(true, "user", "owner")
	sh.Prepare()

	data := []byte("Hello World")
	enc, err := sh.EncryptBytes(data)
	if err != nil {
		t.Fatalf("EncryptBytes failed: %v", err)
	}

	// 11 bytes -> 16 bytes padded + 16 bytes IV = 32 bytes
	if len(enc) != 32 {
		t.Errorf("Expected encrypted length 32, got %d", len(enc))
	}

	// IV should be random
	enc2, _ := sh.EncryptBytes(data)
	if string(enc[:16]) == string(enc2[:16]) {
		t.Errorf("IV should be random")
	}
}

func TestSecurityHandler_EncryptString(t *testing.T) {
	sh := NewSecurityHandler(true, "user", "owner")
	sh.Prepare()

	s := "Test String"
	encStr, err := sh.EncryptString(s)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	if encStr[0] != '<' || encStr[len(encStr)-1] != '>' {
		t.Errorf("Encoded string should be in hex <...> format")
	}
}
