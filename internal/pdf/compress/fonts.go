package compress

import (
	"encoding/hex"
	"regexp"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

var (
	hexStringRe = regexp.MustCompile(`<([0-9A-Fa-f \t\r\n]+)>`)
	litStringRe = regexp.MustCompile(`\((?:\\.|[^\\)])*\)`)
)

func isFontFileStream(dict []byte) bool {
	return merge.HasSubstring(dict, []byte("/Length1")) ||
		merge.HasSubstring(dict, []byte("/Subtype /Type1C")) ||
		merge.HasSubstring(dict, []byte("/Subtype/Type1C")) ||
		merge.HasSubstring(dict, []byte("/Subtype /CIDFontType0C")) ||
		merge.HasSubstring(dict, []byte("/Subtype/CIDFontType0C"))
}

func compactFontStream(dict, data []byte, usedGlyphs []uint16) ([]byte, bool) {
	raw := data
	if streamFilter(dict) == filterFlate {
		decoded, err := decompressFlate(data)
		if err != nil {
			return nil, false
		}
		raw = decoded
	} else if streamFilter(dict) != "" {
		return nil, false
	}

	compacted, err := font.CompactUnusedGlyphs(raw, usedGlyphs)
	if err != nil || len(compacted) == 0 {
		compacted = raw
	}
	compressed, err := compressFlate(compacted)
	if err != nil {
		return nil, false
	}
	if streamFilter(dict) == filterFlate && len(compressed) >= len(data) && len(compacted) >= len(raw) {
		return nil, false
	}
	outDict := setFilter(dict, filterFlate)
	if merge.HasSubstring(outDict, []byte("/Length1")) {
		outDict = setNameInt(outDict, "Length1", len(compacted))
	}
	return buildStream(outDict, compressed), true
}

func collectUsedGlyphs(objects map[int]pdfObject) []uint16 {
	seen := map[uint16]struct{}{0: {}}
	for _, obj := range objects {
		dict, stream, ok := splitStream(obj.body)
		if !ok || isImageXObject(dict) || isFontFileStream(dict) {
			continue
		}
		raw := stream
		switch streamFilter(dict) {
		case filterFlate:
			decoded, err := decompressFlate(stream)
			if err != nil {
				continue
			}
			raw = decoded
		case filterDCT:
			continue
		}
		scanContentGlyphs(raw, seen)
	}
	out := make([]uint16, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	return out
}

func scanContentGlyphs(content []byte, seen map[uint16]struct{}) {
	for _, m := range hexStringRe.FindAllSubmatch(content, -1) {
		addHexGlyphs(m[1], seen)
	}
	for _, m := range litStringRe.FindAll(content, -1) {
		inner := m
		if len(inner) >= 2 && inner[0] == '(' {
			inner = inner[1 : len(inner)-1]
		}
		decoded := unescapePDFString(inner)
		for _, b := range decoded {
			seen[uint16(b)] = struct{}{}
		}
		for _, r := range string(decoded) {
			if r > 255 {
				seen[uint16(r)] = struct{}{}
			}
		}
	}
}

func addHexGlyphs(hexDigits []byte, seen map[uint16]struct{}) {
	cleaned := make([]byte, 0, len(hexDigits))
	for _, b := range hexDigits {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			cleaned = append(cleaned, b)
		}
	}
	if len(cleaned)%2 != 0 {
		return
	}
	raw := make([]byte, len(cleaned)/2)
	n, err := hex.Decode(raw, cleaned)
	if err != nil || n == 0 {
		return
	}
	raw = raw[:n]
	for i := 0; i+1 < len(raw); i += 2 {
		gid := uint16(raw[i])<<8 | uint16(raw[i+1])
		seen[gid] = struct{}{}
	}
	for _, b := range raw {
		seen[uint16(b)] = struct{}{}
	}
}

func unescapePDFString(s []byte) []byte {
	var out []byte
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			out = append(out, s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			out = append(out, '\n')
		case 'r':
			out = append(out, '\r')
		case 't':
			out = append(out, '\t')
		case 'b':
			out = append(out, '\b')
		case 'f':
			out = append(out, '\f')
		case '(', ')', '\\':
			out = append(out, s[i])
		default:
			out = append(out, s[i])
		}
	}
	return out
}
