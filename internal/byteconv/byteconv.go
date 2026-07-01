package byteconv

import "unsafe"

// StringToBytes returns a byte slice that references the string's backing
// array without copying. The returned slice must not be modified.
func StringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}