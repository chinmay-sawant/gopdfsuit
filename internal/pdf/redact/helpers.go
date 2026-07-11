package redact

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/merge"
)

// Package-level PDF markers (avoid per-call []byte allocations).
var (
	markerEncrypt  = []byte(`/Encrypt`)
	markerWBracket = []byte(`/W[`)
	markerIndex    = []byte(`/Index`)
)

// Package-level compiled regexes (hoisted out of loops / hot paths).
var (
	trailerDictRe   = regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	rootRefRe       = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	streamContentRe = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	objStartRe      = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)
	arrayWRe        = regexp.MustCompile(`/W\s*\[(.*?)\]`)
	arrayIndexRe    = regexp.MustCompile(`/Index\s*\[(.*?)\]`)
)

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return bytes.Index(b, sub)
}

// trailerHasEncrypt checks if trailer or any trailer 'Encrypt' appears
func trailerHasEncrypt(data []byte) bool {
	for _, m := range trailerDictRe.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], markerEncrypt) >= 0 {
			return true
		}
	}
	return bytesIndex(data, markerEncrypt) >= 0
}

// tryZlibDecompress attempts to decompress zlib data
func tryZlibDecompress(b []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// tryFlateDecompress attempts to decompress raw flate data
func tryFlateDecompress(b []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(b))
	defer func() {
		_ = r.Close()
	}()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// findRootRef looks for /Root n m R in the PDF bytes.
func findRootRef(data []byte) (objNum int, genNum int, ok bool) {
	if m := rootRefRe.FindSubmatch(data); m != nil {
		objNum, _ = strconv.Atoi(string(m[1]))
		genNum, _ = strconv.Atoi(string(m[2]))
		return objNum, genNum, true
	}
	return 0, 0, false
}

func objGenNum(objGen map[int]int, objNum int) int {
	if objGen == nil {
		return 0
	}
	if g, ok := objGen[objNum]; ok {
		return g
	}
	return 0
}

func isPDFWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// isASCIISpace reports whether c is ASCII whitespace (same set as strings.TrimSpace fast path).
func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

// trimSpace trims leading/trailing ASCII space without calling strings.TrimSpace.
// The returned substring shares the input backing array (no alloc when trimmed).
// Falls back to strings.TrimSpace if non-ASCII bytes remain at the edges.
func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && isASCIISpace(s[start]) {
		start++
	}
	for end > start && isASCIISpace(s[end-1]) {
		end--
	}
	if start < end && (s[start] >= utf8.RuneSelf || s[end-1] >= utf8.RuneSelf) {
		return strings.TrimSpace(s)
	}
	return s[start:end]
}

// parseArrayInts parses array values from PDF dictionary
func parseArrayInts(dict []byte, key string) []int {
	var re *regexp.Regexp
	switch key {
	case `/W`:
		re = arrayWRe
	case `/Index`:
		re = arrayIndexRe
	default:
		re = regexp.MustCompile(regexp.QuoteMeta(key) + `\s*\[(.*?)\]`)
	}
	if m := re.FindSubmatch(dict); m != nil {
		inner := trimSpace(string(m[1]))
		if inner == "" {
			return nil
		}
		parts := strings.Fields(inner)
		res := make([]int, 0, len(parts))
		for _, p := range parts {
			if v, err := strconv.Atoi(p); err == nil {
				res = append(res, v)
			}
		}
		return res
	}
	return nil
}

// readUint reads bytes as unsigned integer
func readUint(b []byte) uint64 {
	var v uint64
	for _, c := range b {
		v = (v << 8) | uint64(byte(c))
	}
	return v
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap / objGen.
func parseXRefStreams(data []byte, objMap map[int][]byte, objGen map[int]int) {
	if objGen == nil {
		return
	}
	for _, b := range merge.FindObjectBoundaries(data) {
		bodyEnd := b.End - len("endobj")
		for bodyEnd > b.BodyStart && isPDFWhitespace(data[bodyEnd-1]) {
			bodyEnd--
		}
		body := data[b.BodyStart:bodyEnd]
		if bytesIndex(body, markerWBracket) < 0 || bytesIndex(body, markerIndex) < 0 {
			continue
		}
		sm := streamContentRe.FindSubmatch(body)
		if sm == nil {
			continue
		}
		streamBytes := sm[1]
		var dec []byte
		if d, err := tryZlibDecompress(streamBytes); err == nil {
			dec = d
		} else if d, err := tryFlateDecompress(streamBytes); err == nil {
			dec = d
		} else {
			dec = streamBytes
		}
		W := parseArrayInts(body, `/W`)
		if len(W) < 3 {
			continue
		}
		if idx := parseArrayInts(body, `/Index`); idx == nil {
			continue
		}
		w0, w1, w2 := W[0], W[1], W[2]
		total := w0 + w1 + w2
		for pos := 0; pos+total <= len(dec); pos += total {
			f1 := int(readUint(dec[pos : pos+w0]))
			f2 := int(readUint(dec[pos+w0 : pos+w0+w1]))
			f3 := int(readUint(dec[pos+w0+w1 : pos+total]))
			if f1 == 1 {
				off := f3
				if off > 0 && off < len(data) {
					endPos := merge.FindEndObj(data, off)
					if endPos == -1 {
						continue
					}
					loc := objStartRe.FindSubmatchIndex(data[off:endPos])
					if loc == nil {
						continue
					}
					onum, _ := strconv.Atoi(string(data[off+loc[2] : off+loc[3]]))
					ogen, _ := strconv.Atoi(string(data[off+loc[4] : off+loc[5]]))
					objBodyStart := off + loc[1]
					objBodyEnd := endPos - len("endobj")
					for objBodyEnd > objBodyStart && isPDFWhitespace(data[objBodyEnd-1]) {
						objBodyEnd--
					}
					objMap[onum] = data[objBodyStart:objBodyEnd]
					objGen[onum] = ogen
				}
			}
			if f1 == 2 {
				objstm := f2
				index := f3
				if stm, ok := objMap[objstm]; ok {
					_ = index
					_ = stm
				}
			}
		}
	}
}
