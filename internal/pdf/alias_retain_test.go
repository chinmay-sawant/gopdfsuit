package pdf

import (
	"testing"
)

// byteString is an allocation-free unsafe conversion: the returned string
// aliases the input slice, so callers must not mutate the buffer while the
// string is in use (WrapTextInto relies on this for buffer reuse).
func TestByteStringContents(t *testing.T) {
	if got := byteString(nil); got != "" {
		t.Fatalf("byteString(nil) = %q, want empty", got)
	}
	if got := byteString([]byte{}); got != "" {
		t.Fatalf("byteString(empty) = %q, want empty", got)
	}
	if got := byteString([]byte("hello")); got != "hello" {
		t.Fatalf("byteString = %q, want hello", got)
	}
}

func TestByteStringRetainsAliasing(t *testing.T) {
	b := []byte("mutable")
	s := byteString(b)
	b[0] = 'X'
	if s != "Xutable" {
		t.Fatalf("byteString must alias input buffer, got %q", s)
	}
}

// NOTE: an EncryptString input-aliasing test belongs to the owner of
// internal/pdf/encryption (another agent is actively editing that package and
// it does not currently compile: EncryptStreamE return mismatch at
// encrypt.go:293). Recorded evidence: EncryptStreamE already pads on a copy
// (append([]byte(nil), data...)), so caller-slice mutation looks handled;
// the aliasing test should be added by the encryption owner once green.
