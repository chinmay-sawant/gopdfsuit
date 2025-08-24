// d:\Chinmay_Personal_Projects\GoPdfSuit\pdf\linearized_reader.go
package pdf

import (
	"bufio"
	"bytes"
	"compress/flate"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Add missing sentinel error used in buildXRef / parseXRefStreamAt
var errNotXRefStream = errors.New("not xref stream object")

type XRefEntryType int

const (
	XRefTypeFree         XRefEntryType = 0
	XRefTypeUncompressed               = 1
	XRefTypeObjStream                  = 2
)

type XRefEntry struct {
	Type         XRefEntryType
	Offset       int64 // for type 1
	Generation   int
	ObjStream    int // for type 2: object number of object stream
	ObjStreamIdx int // index inside that object stream
}

type PDF struct {
	FilePath   string
	Linearized bool
	LinearDict map[string]any
	XRef       map[int]XRefEntry
	RootObj    int
	Size       int
	ID         [2]string
	Raw        []byte
	XRefError  error
}

func Open(path string) (*PDF, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)

	// Header
	header, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(header, "%PDF-") {
		return nil, fmt.Errorf("not a PDF")
	}

	// Keep whole file in memory for simplicity (opt: map pages later)
	data, err := io.ReadAll(io.MultiReader(bytes.NewBufferString(header), br))
	if err != nil {
		return nil, err
	}

	p := &PDF{FilePath: path, XRef: map[int]XRefEntry{}, Raw: data}

	// Detect Linearization dictionary (first indirect object with /Linearized)
	if err := p.parseLinearization(data); err != nil {
		return nil, err
	}

	// Find last xref / trailer or xref stream (tolerant)
	if err := p.buildXRef(data); err != nil {
		p.XRefError = err // keep for diagnostics, but do not fail opening
	}
	return p, nil
}

func (p *PDF) parseLinearization(data []byte) error {
	// Per spec linearization dict is first object, pattern: n 0 obj << /Linearized ...
	// We scan first ~2KB
	window := data
	if len(window) > 4096 {
		window = window[:4096]
	}
	i := bytes.Index(window, []byte(" obj"))
	if i == -1 {
		return nil
	}
	startDict := bytes.Index(window[i:], []byte("<<"))
	if startDict == -1 {
		return nil
	}
	endDict := findMatchingDictEnd(window[i+startDict:])
	if endDict == -1 {
		return nil
	}
	raw := window[i+startDict : i+startDict+endDict]
	if bytes.Contains(raw, []byte("/Linearized")) {
		p.Linearized = true
		dict, _ := parseSimpleDict(raw)
		p.LinearDict = dict
	}
	return nil
}

func (p *PDF) buildXRef(data []byte) error {
	// Search from end for startxref
	const tailScan = 4096
	start := len(data) - tailScan
	if start < 0 {
		start = 0
	}
	tail := data[start:]
	pos := bytes.LastIndex(tail, []byte("startxref"))
	if pos == -1 {
		// Try whole file (linearized PDFs may have early startxref)
		if posAll := bytes.Index(data, []byte("startxref")); posAll != -1 {
			pos = posAll
			tail = data // reuse logic below with full slice
			start = 0
		}
	}
	if pos == -1 {
		// Fallback scans
		if err := p.scanForXRefStreams(data); err == nil {
			return nil
		}
		if err := p.scanClassicXRefTables(data); err == nil {
			return nil
		}
		return fmt.Errorf("no xref found")
	}

	offLineStart := pos + len("startxref")
	nl := bytes.IndexByte(tail[offLineStart:], '\n')
	if nl == -1 {
		return fmt.Errorf("cannot read startxref offset")
	}
	offStr := strings.TrimSpace(string(tail[offLineStart : offLineStart+nl]))
	off, _ := strconv.Atoi(offStr)

	// At offset may be "xref" or an object (xref stream). Validate before parsing.
	if off >= 0 && off < len(data) {
		if bytes.HasPrefix(data[off:], []byte("xref")) {
			if err := p.parseClassicXRef(data, off); err == nil {
				return nil
			}
		} else if isLikelyIndirectObject(data[off:]) {
			if err := p.parseXRefStreamAt(data, off); err == nil {
				return nil
			} else if !errors.Is(err, errNotXRefStream) {
				if scanErr := p.scanForXRefStreams(data); scanErr == nil {
					return nil
				}
				if scan2 := p.scanClassicXRefTables(data); scan2 == nil {
					return nil
				}
				return err
			}
		}
	}
	// Fallback scans
	if err := p.scanForXRefStreams(data); err == nil {
		return nil
	}
	if err := p.scanClassicXRefTables(data); err == nil {
		return nil
	}
	return fmt.Errorf("no xref found")
}

// Scan entire file for any classic xref tables (heuristic)
func (p *PDF) scanClassicXRefTables(data []byte) error {
	idx := 0
	found := false
	for {
		pos := bytes.Index(data[idx:], []byte("xref"))
		if pos == -1 {
			break
		}
		global := idx + pos
		if err := p.parseClassicXRef(data, global); err == nil {
			found = true
			break
		}
		idx = global + 4
	}
	if !found {
		return fmt.Errorf("no classic xref table found")
	}
	return nil
}

// helper: quick check "n n obj"
func isLikelyIndirectObject(b []byte) bool {
	// trim leading whitespace
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\r' || b[i] == '\n' || b[i] == '\t') {
		i++
	}
	// read number
	start := i
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		i++
	}
	if i == start {
		return false
	}
	for i < len(b) && b[i] == ' ' {
		i++
	}
	start2 := i
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		i++
	}
	if i == start2 {
		return false
	}
	for i < len(b) && b[i] == ' ' {
		i++
	}
	if i+3 <= len(b) && string(b[i:i+3]) == "obj" {
		return true
	}
	return false
}

func (p *PDF) scanForXRefStreams(data []byte) error {
	// Broader scan: handle "/Type /XRef" or "/Type/XRef" and fallback heuristic.
	patterns := [][]byte{[]byte("/Type /XRef"), []byte("/Type/XRef")}
	for _, pat := range patterns {
		i := 0
		for {
			j := bytes.Index(data[i:], pat)
			if j == -1 {
				break
			}
			abs := i + j
			// Backtrack to start of object header "<num> <gen> obj"
			headerStart := bytes.LastIndex(data[:abs], []byte(" obj"))
			if headerStart != -1 {
				lineStart := bytes.LastIndex(data[:headerStart], []byte("\n"))
				if lineStart == -1 {
					lineStart = 0
				} else {
					lineStart++
				}
				if err := p.parseXRefStreamAt(data, lineStart); err == nil {
					return nil
				}
			}
			i = abs + len(pat)
		}
	}
	// Heuristic fallback: scan all objects, looking for dict with /W[ and /Index and /Size and 'stream'
	objRx := regexp.MustCompile(`(?s)(\d+\s+\d+\s+obj)(.*?)(stream)`)
	matches := objRx.FindAllSubmatchIndex(data, -1)
	for _, m := range matches {
		body := data[m[2]:m[3]]
		if bytes.Contains(body, []byte("/W[")) && bytes.Contains(body, []byte("/Size")) && bytes.Contains(body, []byte("/Index")) {
			if err := p.parseXRefStreamAt(data, m[0]); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no xref found")
}

func (p *PDF) parseClassicXRef(data []byte, off int) error {
	// Minimal: parse lines after "xref" until "trailer"
	section := data[off:]
	nl := bytes.IndexByte(section, '\n')
	if nl == -1 {
		return fmt.Errorf("bad xref")
	}
	rest := section[nl+1:]
	for {
		if bytes.HasPrefix(rest, []byte("trailer")) {
			// parse trailer dictionary
			trailerStart := bytes.Index(rest, []byte("<<"))
			if trailerStart == -1 {
				break
			}
			end := findMatchingDictEnd(rest[trailerStart:])
			if end == -1 {
				break
			}
			dict, _ := parseSimpleDict(rest[trailerStart : trailerStart+end])
			p.extractTrailer(dict)
			break
		}
		// subsection header: start count
		lineEnd := bytes.IndexByte(rest, '\n')
		if lineEnd == -1 {
			break
		}
		header := strings.TrimSpace(string(rest[:lineEnd]))
		rest = rest[lineEnd+1:]
		if header == "" {
			continue
		}
		if header == "trailer" {
			continue
		}
		parts := strings.Split(header, " ")
		if len(parts) != 2 {
			break
		}
		first, _ := strconv.Atoi(parts[0])
		count, _ := strconv.Atoi(parts[1])
		for i := 0; i < count; i++ {
			lineEnd = bytes.IndexByte(rest, '\n')
			if lineEnd == -1 {
				break
			}
			entry := string(rest[:lineEnd])
			rest = rest[lineEnd+1:]
			if len(entry) < 18 {
				continue
			}
			offStr := strings.TrimSpace(entry[0:10])
			genStr := strings.TrimSpace(entry[11:16])
			t := entry[17]
			offset, _ := strconv.Atoi(offStr)
			gen, _ := strconv.Atoi(genStr)
			if t == 'n' {
				p.XRef[first+i] = XRefEntry{Type: XRefTypeUncompressed, Offset: int64(offset), Generation: gen}
			}
		}
	}
	return nil
}

func (p *PDF) parseXRefStreamAt(data []byte, off int) error {
	// Expect: n n obj << ... /Type /XRef ... /W[...] /Index[...] /Length N >> stream ... endstream
	// Find "<<"
	startDict := bytes.Index(data[off:], []byte("<<"))
	if startDict == -1 {
		return fmt.Errorf("xref stream dict not found")
	}
	dictEnd := findMatchingDictEnd(data[off+startDict:])
	if dictEnd == -1 {
		return fmt.Errorf("unterminated dict")
	}
	dictBytes := data[off+startDict : off+startDict+dictEnd]
	dict, _ := parseSimpleDict(dictBytes)

	// Validate it's an xref stream
	if t, ok := dict["/Type"]; !ok || t != "/XRef" {
		return errNotXRefStream
	}

	lengthVal := getInt(dict["/Length"])
	if lengthVal <= 0 {
		// Try indirect length: pattern "n 0 R"
		if lv, ok := dict["/Length"].(string); ok && strings.HasSuffix(lv, " R") {
			if resolved := resolveIndirectLength(data, lv); resolved > 0 {
				lengthVal = resolved
			}
		}
	}
	if lengthVal <= 0 {
		return fmt.Errorf("missing /Length in xref stream")
	}

	streamPos := off + startDict + dictEnd
	streamKeyword := []byte("stream")
	kIdx := bytes.Index(data[streamPos:], streamKeyword)
	if kIdx == -1 {
		return fmt.Errorf("stream keyword missing")
	}
	afterStream := streamPos + kIdx + len(streamKeyword)
	// stream data starts after newline
	if data[afterStream] == '\r' && data[afterStream+1] == '\n' {
		afterStream += 2
	} else if data[afterStream] == '\n' {
		afterStream++
	}

	rawStream := data[afterStream : afterStream+lengthVal]
	// Decode (support only Flate)
	if filter, ok := dict["/Filter"]; ok {
		switch filter {
		case "/FlateDecode":
			dec := flate.NewReader(bytes.NewReader(rawStream))

			buf, err := io.ReadAll(dec)
			dec.Close()
			if err != nil {
				return err
			}
			rawStream = buf
		default:
			return fmt.Errorf("unsupported xref stream filter %v", filter)
		}
	}

	wArr := getIntArray(dict["/W"])
	if len(wArr) != 3 {
		return fmt.Errorf("invalid /W")
	}
	indexArray := getIntArray(dict["/Index"])
	if len(indexArray) == 0 {
		size := getInt(dict["/Size"])
		indexArray = []int{0, size}
	}

	cursor := 0
	for i := 0; i < len(indexArray); i += 2 {
		startObj := indexArray[i]
		count := indexArray[i+1]
		for j := 0; j < count; j++ {
			if cursor+wArr[0]+wArr[1]+wArr[2] > len(rawStream) {
				return fmt.Errorf("xref stream truncated")
			}
			f1 := readUInt(rawStream[cursor : cursor+wArr[0]])
			cursor += wArr[0]
			f2 := readUInt(rawStream[cursor : cursor+wArr[1]])
			cursor += wArr[1]
			f3 := readUInt(rawStream[cursor : cursor+wArr[2]])
			cursor += wArr[2]

			objNum := startObj + j
			switch f1 {
			case 0:
				p.XRef[objNum] = XRefEntry{Type: XRefTypeFree}
			case 1:
				p.XRef[objNum] = XRefEntry{Type: XRefTypeUncompressed, Offset: int64(f2), Generation: int(f3)}
			case 2:
				p.XRef[objNum] = XRefEntry{Type: XRefTypeObjStream, ObjStream: int(f2), ObjStreamIdx: int(f3)}
			}
		}
	}

	p.extractTrailer(dict)
	return nil
}

// resolveIndirectLength finds an indirect numeric object used as /Length
func resolveIndirectLength(data []byte, ref string) int {
	parts := strings.Fields(ref)
	if len(parts) < 3 {
		return 0
	}
	objNum := parts[0]
	needle := []byte(objNum + " 0 obj")
	idx := bytes.Index(data, needle)
	if idx == -1 {
		return 0
	}
	segment := data[idx : idx+64]
	// extract first integer after header
	reNum := regexp.MustCompile(objNum + ` 0 obj\s+(\d+)`)
	m := reNum.FindSubmatch(segment)
	if len(m) == 2 {
		v, _ := strconv.Atoi(string(m[1]))
		return v
	}
	return 0
}

func (p *PDF) extractTrailer(dict map[string]any) {
	if root, ok := dict["/Root"]; ok {
		if v, ok2 := root.(string); ok2 {
			p.RootObj = parseObjRef(v)
		}
	}
	if size, ok := dict["/Size"]; ok {
		p.Size = getInt(size)
	}
	if id, ok := dict["/ID"]; ok {
		if arr, ok2 := id.([]any); ok2 && len(arr) == 2 {
			p.ID[0] = stripHex(arr[0])
			p.ID[1] = stripHex(arr[1])
		}
	}
}

func parseObjRef(s string) int {
	parts := strings.Split(strings.TrimSpace(s), " ")
	if len(parts) >= 2 {
		n, _ := strconv.Atoi(parts[0])
		return n
	}
	return 0
}

func stripHex(v any) string {
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
			s = s[1 : len(s)-1]
			if _, err := hex.DecodeString(s); err == nil {
				return s
			}
		}
	}
	return ""
}

func findMatchingDictEnd(b []byte) int {
	depth := 0
	for i := 0; i < len(b)-1; i++ {
		if b[i] == '<' && b[i+1] == '<' {
			depth++
			i++
			continue
		}
		if b[i] == '>' && b[i+1] == '>' {
			depth--
			i++
			if depth == 0 {
				return i + 1
			}
			continue
		}
	}
	return -1
}

func parseSimpleDict(b []byte) (map[string]any, error) {
	res := map[string]any{}
	s := string(b)
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(s, ">>"), "<<"))
	toks := tokenize(s)
	for i := 0; i < len(toks); i++ {
		if strings.HasPrefix(toks[i], "/") {
			key := toks[i]
			if i+1 < len(toks) {
				val := toks[i+1]
				if val == "<<" { // nested dict (skip simplistic)
					// find matching
					sub := toks[i+1:]
					joined := strings.Join(sub, " ")
					end := findMatchingDictEnd([]byte(joined))
					if end > 0 {
						res[key] = string(joined[:end])
						// advance tokens naive
					}
				} else if strings.HasPrefix(val, "[") {
					res[key] = parseArray(val)
				} else {
					res[key] = val
				}
			}
		}
	}
	return res, nil
}

func tokenize(s string) []string {
	var out []string
	var cur strings.Builder
	inHex := false
	inArr := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '<' && (i+1 < len(s) && s[i+1] != '<'):
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			inHex = true
			cur.WriteByte(c)
		case c == '>' && inHex:
			cur.WriteByte(c)
			out = append(out, cur.String())
			cur.Reset()
			inHex = false
		case inHex:
			cur.WriteByte(c)
		case c == '[':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			inArr++
			cur.WriteByte(c)
		case c == ']':
			cur.WriteByte(c)
			inArr--
			if inArr == 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		case (c == ' ' || c == '\n' || c == '\r' || c == '\t') && inArr == 0:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func parseArray(s string) []any {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return nil
	}
	body := strings.TrimSpace(s[1 : len(s)-1])
	if body == "" {
		return nil
	}
	parts := tokenize(body)
	out := make([]any, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(p); err == nil {
			out = append(out, n)
		} else {
			out = append(out, p)
		}
	}
	return out
}

func getInt(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case string:
		i, _ := strconv.Atoi(t)
		return i
	}
	return 0
}

func getIntArray(v any) []int {
	var res []int
	switch a := v.(type) {
	case []any:
		for _, e := range a {
			res = append(res, getInt(e))
		}
	case string:
		if strings.HasPrefix(a, "[") {
			arr := parseArray(a)
			for _, e := range arr {
				res = append(res, getInt(e))
			}
		}
	}
	return res
}

func readUInt(b []byte) uint64 {
	var v uint64
	for _, c := range b {
		v = (v << 8) | uint64(c)
	}
	return v
}

// --- XFDF parsing ---

type xfdfRootUk struct {
	XMLName xml.Name   `xml:"xfdf"`
	Fields  xfdfFields `xml:"fields"`
}
type xfdfFields struct {
	XMLName xml.Name      `xml:"fields"`
	Fields  []xfdfFieldUk `xml:"field"`
}
type xfdfFieldUk struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

// ParseXFDF loads an XFDF file into a map[fieldName]value.
func ParseXFDFUk(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var root xfdfRootUk
	if err := xml.Unmarshal(b, &root); err != nil {
		return nil, err
	}
	m := make(map[string]string, len(root.Fields.Fields))
	for _, f := range root.Fields.Fields {
		m[f.Name] = f.Value
	}
	return m, nil
}

// escapePDFString escapes (), \ as required in literal strings.
func escapePDFStringUk(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`)
	return r.Replace(s)
}

// FillFields applies field values and returns modified bytes & count.
func (p *PDF) FillFields(fieldMap map[string]string) ([]byte, int, error) {
	if len(fieldMap) == 0 {
		return p.Raw, 0, nil
	}
	content := p.Raw

	reObject := regexp.MustCompile(`(?s)(\d+\s+\d+\s+obj)(.*?)(\bendobj\b)`)
	reIsAnnot := regexp.MustCompile(`/Type\s*/Annot`)
	reIsWidget := regexp.MustCompile(`/Subtype\s*/Widget`)
	reIsBtn := regexp.MustCompile(`/FT\s*/Btn`)
	reIsTx := regexp.MustCompile(`/FT\s*/Tx`)
	reFieldNameT := regexp.MustCompile(`/T\s*\((.*?)\)`)
	reRemoveAP := regexp.MustCompile(`/AP\s*<<.*?>>`)

	matches := reObject.FindAllSubmatchIndex(content, -1)
	if matches == nil {
		return p.Raw, 0, fmt.Errorf("no objects found")
	}

	var out bytes.Buffer
	last := 0
	modified := 0

	for _, m := range matches {
		out.Write(content[last:m[0]])

		objHeader := content[m[2]:m[3]]
		objBody := content[m[4]:m[5]]
		objFooter := content[m[6]:m[7]]

		if reIsAnnot.Match(objBody) && reIsWidget.Match(objBody) {
			nameMatch := reFieldNameT.FindSubmatch(objBody)
			if len(nameMatch) >= 2 {
				if val, ok := fieldMap[string(nameMatch[1])]; ok {
					origBody := objBody
					var inject []byte

					if reIsBtn.Match(objBody) {
						if strings.EqualFold(val, "on") || strings.EqualFold(val, "yes") {
							inject = []byte(" /AS /On /V /On")
						}
					} else if reIsTx.Match(objBody) {
						if !strings.EqualFold(val, "N/A") {
							origBody = reRemoveAP.ReplaceAll(origBody, []byte(""))
							inject = []byte(" /V (" + escapePDFStringUk(val) + ")")
						}
					}

					if inject != nil {
						loc := reIsAnnot.FindIndex(origBody)
						if loc != nil {
							modified++
							var merged bytes.Buffer
							merged.Write(origBody[:loc[1]])
							merged.Write(inject)
							merged.Write(origBody[loc[1]:])
							out.Write(objHeader)
							out.Write(merged.Bytes())
							out.Write(objFooter)
							last = m[1]
							continue
						}
					}
				}
			}
		}

		out.Write(content[m[0]:m[1]])
		last = m[1]
	}
	out.Write(content[last:])
	return out.Bytes(), modified, nil
}
