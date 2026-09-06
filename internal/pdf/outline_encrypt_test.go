package pdf

import (
	"encoding/hex"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

// identityEncryptor passes plaintext through so tests can inspect the exact
// bytes the outline builder hands to the encryptor.
type identityEncryptor struct{}

func (identityEncryptor) EncryptString(data []byte, _, _ int) []byte {
	return append([]byte(nil), data...)
}

func (identityEncryptor) EncryptStream(data []byte, _, _ int) []byte {
	return append([]byte(nil), data...)
}

func (identityEncryptor) GetEncryptDictionary(_ int) string { return "" }

// decodeOutlineTitleHex decodes the /Title <hex> emitted for an encrypted
// outline item back to a Go string (UTF-16BE with BOM or plain ASCII).
func decodeOutlineTitleHex(t *testing.T, obj string) string {
	t.Helper()
	start := strings.Index(obj, "/Title <")
	if start < 0 {
		t.Fatalf("no /Title hex in %q", obj)
	}
	hexStr := obj[start+len("/Title <"):]
	end := strings.Index(hexStr, ">")
	raw, err := hex.DecodeString(hexStr[:end])
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) >= 2 && raw[0] == 0xFE && raw[1] == 0xFF {
		u16 := make([]uint16, 0, len(raw)/2)
		for i := 2; i+1 < len(raw); i += 2 {
			u16 = append(u16, uint16(raw[i])<<8|uint16(raw[i+1]))
		}
		return string(utf16.Decode(u16))
	}
	return string(raw)
}

// TestEncryptedOutlineMultiUnicodeTitles ensures consecutive non-ASCII titles
// each encode to exactly their own title (no accumulation across iterations).
func TestEncryptedOutlineMultiUnicodeTitles(t *testing.T) {
	pm := NewPageManager(PageDimensions{Width: 595, Height: 842},
		PageMargins{Top: 36, Bottom: 36, Left: 36, Right: 36}, false, nil, false, 0)
	ob := NewOutlineBuilder(pm, identityEncryptor{})

	titles := []string{"Zürich-Nord", "日本語タイトル", "Τίτλος-τρία", "Plain ASCII"}
	var bookmarks []models.Bookmark
	for _, title := range titles {
		bookmarks = append(bookmarks, models.Bookmark{Title: title, Page: 1, Open: true})
	}
	if root := ob.BuildOutlines(bookmarks); root == 0 {
		t.Fatal("BuildOutlines returned 0")
	}

	if len(ob.outlineItems) != len(titles) {
		t.Fatalf("got %d outline items, want %d", len(ob.outlineItems), len(titles))
	}
	for i, item := range ob.outlineItems {
		obj, ok := pm.ExtraObjects[item.ObjectID]
		if !ok {
			t.Fatalf("missing object %d", item.ObjectID)
		}
		if got := decodeOutlineTitleHex(t, string(obj)); got != titles[i] {
			t.Fatalf("item %d title = %q, want %q", i, got, titles[i])
		}
	}
}
