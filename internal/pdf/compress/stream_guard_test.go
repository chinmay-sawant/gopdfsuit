package compress

import (
	"bytes"
	"testing"
)

// TestSplitStreamHonorsLength proves stream binary containing the
// "endstream" token round-trips via /Length instead of LastIndex
// swallowing trailing objects.
func TestSplitStreamHonorsLength(t *testing.T) {
	data := []byte("HELLO endstream WORLD0123456789")
	if len(data) != 31 {
		t.Fatalf("fixture must be 31 bytes, got %d", len(data))
	}
	var body bytes.Buffer
	body.WriteString("<< /Length 31 /Filter /FlateDecode >>\nstream\n")
	body.Write(data)
	body.WriteString("\nendstream\nendobj\ntrailer endstream junk")
	dict, got, ok := splitStream(body.Bytes())
	if !ok {
		t.Fatalf("splitStream failed")
	}
	if !bytes.Contains(dict, []byte("/Length")) {
		t.Fatalf("dict lost: %q", dict)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("got %q, want exact %d-byte payload", got, len(data))
	}
}

// TestSplitStreamFirstEndstream proves the no-Length fallback stops at the
// first endstream token instead of swallowing trailing bytes.
func TestSplitStreamFirstEndstream(t *testing.T) {
	body := []byte("<< /Filter /FlateDecode >>\nstream\nabc\nendstream\ntrailer endstream junk")
	_, got, ok := splitStream(body)
	if !ok {
		t.Fatalf("splitStream failed")
	}
	if string(got) != "abc" {
		t.Fatalf("got %q, want %q", got, "abc")
	}
}
