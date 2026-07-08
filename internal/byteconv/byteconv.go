// Package byteconv provides zero-copy string/byte slice conversion utilities
// for hot PDF serialization paths (PERF-32).
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

// BytesToString returns a string that references the same backing array as b
// without copying. Do not mutate b after calling this.
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
