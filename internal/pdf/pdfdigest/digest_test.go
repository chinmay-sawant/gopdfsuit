package pdfdigest_test

import (
	"crypto/md5"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/pdfdigest"
)

func TestDigest16MatchesStdlib(t *testing.T) {
	vectors := [][]byte{
		{},
		[]byte("a"),
		[]byte("abc"),
		[]byte("message digest"),
		[]byte("abcdefghijklmnopqrstuvwxyz"),
		[]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"),
		[]byte("12345678901234567890123456789012345678901234567890123456789012345678901234567890"),
		make([]byte, 100),
		make([]byte, 1000),
	}
	for i, v := range vectors {
		want := md5.Sum(v)
		got := pdfdigest.Digest16(v)
		if got != want {
			t.Fatalf("vector %d: got %x want %x", i, got, want)
		}
	}
}

func TestNewIncrementalMatchesDigest16(t *testing.T) {
	parts := [][]byte{
		[]byte("hello"),
		[]byte(" "),
		[]byte("world"),
		[]byte("!"),
	}
	var joined []byte
	h := pdfdigest.New()
	for _, p := range parts {
		joined = append(joined, p...)
		if _, err := h.Write(p); err != nil {
			t.Fatal(err)
		}
	}
	var got [16]byte
	copy(got[:], h.Sum(nil))
	want := pdfdigest.Digest16(joined)
	if got != want {
		t.Fatalf("incremental got %x want %x", got, want)
	}
	if got != md5.Sum(joined) {
		t.Fatalf("incremental != stdlib: got %x", got)
	}
}
