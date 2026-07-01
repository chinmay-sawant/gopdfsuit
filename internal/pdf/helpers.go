package pdf

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	byteEncrypt     = []byte("/Encrypt")
	byteWBracket    = []byte("/W[")
	byteIndex       = []byte("/Index")
	byteObjStreamRe = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	byteStreamRe    = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	byteRootRe      = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	byteTrailerRe   = regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	byteObjAtRe     = regexp.MustCompile(`(?s)^(\s*)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
)

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return bytes.Index(b, sub)
}

// trailerHasEncrypt checks if trailer or any trailer 'Encrypt' appears
func trailerHasEncrypt(data []byte) bool {
	for _, m := range byteTrailerRe.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], byteEncrypt) >= 0 {
			return true
		}
	}
	return bytesIndex(data, byteEncrypt) >= 0
}

// tryZlibDecompress attempts to decompress zlib data
func tryZlibDecompress(b []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("warning: zlib reader close failed: %v", err)
		}
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
		if err := r.Close(); err != nil {
			log.Printf("warning: flate reader close failed: %v", err)
		}
	}()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// findRootRef looks for /Root n m R in the PDF bytes
func findRootRef(data []byte) (string, bool) {
	if m := byteRootRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
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
			v, err := strconv.Atoi(p)
			if err == nil {
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

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	for _, m := range byteObjStreamRe.FindAllSubmatch(data, -1) {
		body := m[3]
		if bytesIndex(body, byteWBracket) < 0 || bytesIndex(body, byteIndex) < 0 {
			continue
		}
		sm := byteStreamRe.FindSubmatch(body)
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
		Index := parseArrayInts(body, `/Index`)
		if Index == nil {
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
					tail := data[off:]
					if ro := byteObjAtRe.FindSubmatch(tail); ro != nil {
						onum := string(ro[2])
						ogen := string(ro[3])
						key := onum + " " + ogen
						objMap[key] = ro[4]
					}
				}
			}
			if f1 == 2 {
				objstm := f2
				index := f3
				var keyBuf [16]byte
				key := string(appendObjMapKey(keyBuf[:0], objstm))
				if stm, ok := objMap[key]; ok {
					_ = index
					_ = stm
				}
			}
		}
	}
}