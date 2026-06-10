package redact

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/merge"
)

var encryptBytes = []byte("/Encrypt")
var wBytes = []byte("/W[")
var indexBytes = []byte("/Index")
var streamRe = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
var objStartRe = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return strings.Index(string(b), string(sub))
}

// trailerHasEncrypt checks if trailer or any trailer 'Encrypt' appears
func trailerHasEncrypt(data []byte) bool {
	trRe := regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	for _, m := range trRe.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], encryptBytes) >= 0 {
			return true
		}
	}
	return bytesIndex(data, encryptBytes) >= 0
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
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	if m := rootRe.FindSubmatch(data); m != nil {
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

// parseArrayInts parses array values from PDF dictionary
func parseArrayInts(dict []byte, key string) []int {
	re := regexp.MustCompile(key + `\s*\[(.*?)\]`)
	if m := re.FindSubmatch(dict); m != nil {
		inner := strings.TrimSpace(string(m[1]))
		if inner == "" {
			return nil
		}
		parts := strings.Fields(inner)
		res := make([]int, 0, len(parts))
		for _, p := range parts {
			var v int
			if _, err := fmt.Sscanf(p, "%d", &v); err == nil {
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
		if bytesIndex(body, wBytes) < 0 || bytesIndex(body, indexBytes) < 0 {
			continue
		}
		sm := streamRe.FindSubmatch(body)
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
		if parseArrayInts(body, `/Index`) == nil {
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
