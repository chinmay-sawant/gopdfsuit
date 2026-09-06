package compress

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/pdfobj"
)

const (
	filterFlate = "FlateDecode"
	filterDCT   = "DCTDecode"
)

var (
	lengthIndirectRe = regexp.MustCompile(`/Length\s+\d+\s+\d+\s+R`)
	lengthDirectRe   = regexp.MustCompile(`/Length\s+\d+`)
	filterArrayRe    = regexp.MustCompile(`/Filter\s*\[\s*/([A-Za-z0-9]+)\s*\]`)
	filterNameRe     = regexp.MustCompile(`/Filter\s*/([A-Za-z0-9]+)`)
	widthRe          = regexp.MustCompile(`/Width\s+\d+`)
	heightRe         = regexp.MustCompile(`/Height\s+\d+`)
	bitsRe           = regexp.MustCompile(`/BitsPerComponent\s+(\d+)`)
	colorSpaceRe     = regexp.MustCompile(`/ColorSpace\s*/([A-Za-z0-9]+)`)
)

func splitStream(body []byte) (dict, data []byte, ok bool) {
	start := pdfobj.FindStreamStart(body)
	if start < 0 {
		return nil, nil, false
	}
	dict = bytes.TrimSpace(body[:start])
	ptr := start + len("stream")
	if ptr < len(body) && body[ptr] == '\r' {
		ptr++
	}
	if ptr < len(body) && body[ptr] == '\n' {
		ptr++
	}
	if ptr > len(body) {
		return nil, nil, false
	}

	if n := directLength(dict); n >= 0 && ptr+n <= len(body) {
		return dict, body[ptr : ptr+n], true
	}

	// Without a usable /Length, the stream ends at the first endstream
	// token: LastIndex would swallow trailing objects when binary stream
	// data itself contains that token.
	end := bytes.Index(body[ptr:], []byte("endstream"))
	if end < 0 {
		return nil, nil, false
	}
	data = body[ptr : ptr+end]
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r') {
		data = data[:len(data)-1]
	}
	return dict, data, true
}

func buildStream(dict, data []byte) []byte {
	dict = setDirectLength(dict, len(data))
	var b bytes.Buffer
	b.Grow(len(dict) + len(data) + 24)
	b.Write(dict)
	if !bytes.HasSuffix(dict, []byte("\n")) {
		b.WriteByte('\n')
	}
	b.WriteString("stream\n")
	b.Write(data)
	b.WriteString("\nendstream")
	return b.Bytes()
}

func decompressFlate(data []byte) ([]byte, error) {
	return pdfobj.DecompressAny(data)
}

func streamFilter(dict []byte) string {
	if m := filterArrayRe.FindSubmatch(dict); m != nil {
		return string(m[1])
	}
	if m := filterNameRe.FindSubmatch(dict); m != nil {
		return string(m[1])
	}
	return ""
}

func directLength(dict []byte) int {
	if lengthIndirectRe.Match(dict) {
		return -1
	}
	m := lengthDirectRe.FindSubmatch(dict)
	if m == nil {
		return -1
	}
	n, err := strconv.Atoi(string(bytes.TrimSpace(bytes.TrimPrefix(m[0], []byte("/Length")))))
	if err != nil {
		return -1
	}
	return n
}

func setDirectLength(dict []byte, n int) []byte {
	repl := []byte("/Length " + strconv.Itoa(n))
	if lengthIndirectRe.Match(dict) {
		return lengthIndirectRe.ReplaceAll(dict, repl)
	}
	if lengthDirectRe.Match(dict) {
		return lengthDirectRe.ReplaceAll(dict, repl)
	}
	return insertBeforeDictEnd(dict, repl)
}

func setFilter(dict []byte, name string) []byte {
	repl := []byte("/Filter /" + name)
	if filterArrayRe.Match(dict) {
		return filterArrayRe.ReplaceAll(dict, repl)
	}
	if filterNameRe.Match(dict) {
		return filterNameRe.ReplaceAll(dict, repl)
	}
	return insertBeforeDictEnd(dict, repl)
}

func setNameInt(dict []byte, name string, n int) []byte {
	var re *regexp.Regexp
	switch name {
	case "Height":
		re = heightRe
	case "Width":
		re = widthRe
	default:
		re = regexp.MustCompile(`/` + name + `\s+\d+`)
	}
	repl := []byte("/" + name + " " + strconv.Itoa(n))
	if re.Match(dict) {
		return re.ReplaceAll(dict, repl)
	}
	return insertBeforeDictEnd(dict, repl)
}

func insertBeforeDictEnd(dict, token []byte) []byte {
	idx := bytes.LastIndex(dict, []byte(">>"))
	if idx < 0 {
		return dict
	}
	var out bytes.Buffer
	out.Grow(len(dict) + len(token) + 1)
	out.Write(dict[:idx])
	if idx > 0 && dict[idx-1] != ' ' && dict[idx-1] != '\n' {
		out.WriteByte(' ')
	}
	out.Write(token)
	out.WriteByte(' ')
	out.Write(dict[idx:])
	return out.Bytes()
}

func dictInt(dict []byte, re *regexp.Regexp) int {
	m := re.FindSubmatch(dict)
	if m == nil {
		return 0
	}
	if len(m) > 1 {
		n, _ := strconv.Atoi(string(m[1]))
		return n
	}
	fields := bytes.Fields(m[0])
	if len(fields) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(string(fields[1]))
	return n
}

func dictNameAfter(dict []byte, re *regexp.Regexp) string {
	m := re.FindSubmatch(dict)
	if len(m) < 2 {
		return ""
	}
	return string(m[1])
}

func isImageXObject(dict []byte) bool {
	return pdfobj.HasSubstring(dict, []byte("/Subtype /Image")) ||
		pdfobj.HasSubstring(dict, []byte("/Subtype/Image"))
}

func hasSMask(dict []byte) bool {
	return pdfobj.HasSubstring(dict, []byte("/SMask"))
}

func hasDecodeParms(dict []byte) bool {
	return pdfobj.HasSubstring(dict, []byte("/DecodeParms")) ||
		pdfobj.HasSubstring(dict, []byte("/DP "))
}

func removeNamedValue(dict []byte, name string) []byte {
	key := []byte("/" + name)
	idx := bytes.Index(dict, key)
	if idx < 0 {
		return dict
	}
	after := idx + len(key)
	if after < len(dict) && isNameContinue(dict[after]) {
		return dict
	}
	j := after
	for j < len(dict) && isPDFWhitespace(dict[j]) {
		j++
	}
	if j >= len(dict) {
		return dict
	}
	end := skipPDFValue(dict, j)
	out := make([]byte, 0, len(dict)-(end-idx))
	out = append(out, dict[:idx]...)
	out = append(out, dict[end:]...)
	return out
}

func skipPDFValue(data []byte, pos int) int {
	if pos >= len(data) {
		return pos
	}
	switch {
	case data[pos] == '<' && pos+1 < len(data) && data[pos+1] == '<':
		return pdfobj.SkipDictionary(data, pos)
	case data[pos] == '[':
		return pdfobj.SkipArray(data, pos)
	case data[pos] == '/':
		pos++
		for pos < len(data) && isNameContinue(data[pos]) {
			pos++
		}
		return pos
	case data[pos] == '(':
		return pdfobj.SkipStringLiteral(data, pos)
	default:
		return skipNumberOrRef(data, pos)
	}
}

func skipNumberOrRef(data []byte, pos int) int {
	pos = skipNumber(data, pos)
	j := pos
	for j < len(data) && isPDFWhitespace(data[j]) {
		j++
	}
	if j < len(data) && (data[j] == '+' || data[j] == '-' || (data[j] >= '0' && data[j] <= '9')) {
		k := skipNumber(data, j)
		for k < len(data) && isPDFWhitespace(data[k]) {
			k++
		}
		if k+1 <= len(data) && data[k] == 'R' {
			return k + 1
		}
	}
	return pos
}

func skipNumber(data []byte, pos int) int {
	if pos < len(data) && (data[pos] == '+' || data[pos] == '-') {
		pos++
	}
	for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' {
		pos++
	}
	if pos < len(data) && data[pos] == '.' {
		pos++
		for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' {
			pos++
		}
	}
	return pos
}

func isNameContinue(b byte) bool {
	if isPDFWhitespace(b) {
		return false
	}
	switch b {
	case '/', '<', '>', '[', ']', '(', ')', '{', '}':
		return false
	}
	return true
}
