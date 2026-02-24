package font

// ObjectEncryptor defines the interface for encrypting PDF objects.
// This is defined here to avoid a circular dependency with the parent pdf package.
type ObjectEncryptor interface {
	EncryptString(data []byte, objNum, genNum int) []byte
	EncryptStream(data []byte, objNum, genNum int) []byte
	GetEncryptDictionary(encryptObjID int) string
}
