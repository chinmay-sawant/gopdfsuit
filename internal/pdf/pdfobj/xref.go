package pdfobj

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

// Object is a single parsed PDF object: its number, generation, and body
// bytes (the slice between "obj" and "endobj", whitespace-trimmed).
type Object struct {
	Num  int
	Gen  int
	Body []byte
}

// ParsedPDF is the typed result of the read seam: the object map plus
// trailer and xref metadata. Byte-level fallbacks still operate on Data.
type ParsedPDF struct {
	Data      []byte
	Objects   map[int]Object
	Gens      map[int]int
	MaxObj    int
	Version   string
	RootNum   int
	RootGen   int
	HasRoot   bool
	TrailerID string
	StartXRef int
	Encrypted bool
	XRefed    bool
}

// Body returns the body of object num.
func (p *ParsedPDF) Body(num int) ([]byte, bool) {
	o, ok := p.Objects[num]
	if !ok {
		return nil, false
	}
	return o.Body, true
}

// Gen returns the generation number of object num (0 when unknown).
func (p *ParsedPDF) Gen(num int) int {
	if p.Gens == nil {
		return 0
	}
	return p.Gens[num]
}

var (
	encryptEntryRe = regexp.MustCompile(`/Encrypt\s*(\d+\s+\d+\s+R|<<)`)
	rootRefRe      = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	trailerRe      = regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`)
	idRe           = regexp.MustCompile(`(?s)/ID\s*(\[(?:.|\n|\r)*?\])`)
	startXRefRe    = regexp.MustCompile(`(?s)startxref\s*(\d+)\s*%%EOF\s*$`)
	startXRefAnyRe = regexp.MustCompile(`startxref\s*(\d+)`)
	objStartRe     = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)
	streamRe       = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
)

// HasEncryptEntry reports whether data declares document encryption: a real
// /Encrypt entry (indirect reference or inline dict) outside stream data.
func HasEncryptEntry(data []byte) bool {
	return encryptEntryRe.Match(BytesWithoutStreams(data))
}

// FindRootRef looks for /Root n m R in the PDF bytes.
func FindRootRef(data []byte) (num, gen int, ok bool) {
	m := rootRefRe.FindSubmatch(data)
	if m == nil {
		return 0, 0, false
	}
	num, _ = strconv.Atoi(string(m[1]))
	gen, _ = strconv.Atoi(string(m[2]))
	return num, gen, true
}

// ExtractLastStartXRef returns the last startxref offset, or 0.
func ExtractLastStartXRef(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	if m := startXRefRe.FindSubmatch(data); m != nil {
		if n, err := strconv.Atoi(string(m[1])); err == nil {
			return n
		}
	}
	all := startXRefAnyRe.FindAllSubmatch(data, -1)
	if len(all) == 0 {
		return 0
	}
	last := all[len(all)-1]
	if len(last) < 2 {
		return 0
	}
	n, err := strconv.Atoi(string(last[1]))
	if err != nil {
		return 0
	}
	return n
}

// ExtractTrailerID returns the first /ID array found in a trailer dict,
// falling back to a file-wide scan.
func ExtractTrailerID(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if tm := trailerRe.FindSubmatch(data); tm != nil {
		if idm := idRe.FindSubmatch(tm[1]); idm != nil {
			return strings.TrimSpace(string(idm[1]))
		}
	}
	if idm := idRe.FindSubmatch(data); idm != nil {
		return strings.TrimSpace(string(idm[1]))
	}
	return ""
}

// trimBodyEnd drops the trailing "endobj" keyword and whitespace.
func trimBodyEnd(data []byte, b ObjectBoundary) []byte {
	bodyEnd := b.End - len("endobj")
	for bodyEnd > b.BodyStart && IsWhitespace(data[bodyEnd-1]) {
		bodyEnd--
	}
	if bodyEnd < b.BodyStart {
		return nil
	}
	return data[b.BodyStart:bodyEnd]
}

// Parse runs the full read seam over data: object boundaries, object-stream
// expansion, xref-stream augmentation, and trailer metadata.
func Parse(data []byte) *ParsedPDF {
	p := &ParsedPDF{
		Data:    data,
		Objects: make(map[int]Object),
		Gens:    make(map[int]int),
		Version: DetectVersion(data),
	}
	for _, b := range FindObjectBoundaries(data) {
		body := trimBodyEnd(data, b)
		if body == nil {
			continue
		}
		p.Objects[b.ObjNum] = Object{Num: b.ObjNum, Gen: b.GenNum, Body: body}
		p.Gens[b.ObjNum] = b.GenNum
		if b.ObjNum > p.MaxObj {
			p.MaxObj = b.ObjNum
		}
		if IsObjectStream(body) {
			for onum, frag := range ParseObjectStream(body) {
				if _, exists := p.Objects[onum]; !exists {
					p.Objects[onum] = Object{Num: onum, Gen: 0, Body: frag}
					p.Gens[onum] = 0
					if onum > p.MaxObj {
						p.MaxObj = onum
					}
				}
			}
		}
	}
	AugmentIntMap(data, p.Objects, p.Gens)
	if n, g, ok := FindRootRef(data); ok {
		p.RootNum, p.RootGen, p.HasRoot = n, g, true
	}
	p.TrailerID = ExtractTrailerID(data)
	p.StartXRef = ExtractLastStartXRef(data)
	p.Encrypted = HasEncryptEntry(data)
	p.XRefed = p.StartXRef > 0
	return p
}

// AugmentIntMap expands xref-stream entries into objMap/objGen, adding
// objects reachable only through type-1 (offset) entries.
func AugmentIntMap(data []byte, objMap map[int]Object, objGen map[int]int) {
	if objGen == nil {
		return
	}
	view := make(map[int][]byte, len(objMap))
	for num, o := range objMap {
		view[num] = o.Body
	}
	AugmentRawIntMap(data, view, objGen)
	for num, body := range view {
		if _, ok := objMap[num]; !ok {
			objMap[num] = Object{Num: num, Gen: objGen[num], Body: body}
		}
	}
}

// AugmentRawIntMap is the raw-body variant of AugmentIntMap for callers
// whose object map is keyed by object number to body bytes (e.g. redact).
// New entries are added to objMap/objGen in place.
func AugmentRawIntMap(data []byte, objMap map[int][]byte, objGen map[int]int) {
	if objGen == nil {
		return
	}
	for _, b := range FindObjectBoundaries(data) {
		body := trimBodyEnd(data, b)
		if body == nil {
			continue
		}
		if !bytes.Contains(body, []byte("/W[")) || !bytes.Contains(body, []byte("/Index")) {
			continue
		}
		sm := streamRe.FindSubmatch(body)
		if sm == nil {
			continue
		}
		dec, err := DecompressAny(sm[1])
		if err != nil {
			dec = sm[1]
		}
		W := ParseArrayInts(body, `/W`)
		if len(W) < 3 {
			continue
		}
		if ParseArrayInts(body, `/Index`) == nil {
			continue
		}
		w0, w1, w2 := W[0], W[1], W[2]
		total, ok := ValidXRefWidths(w0, w1, w2)
		if !ok {
			continue
		}
		for pos := 0; pos+total <= len(dec); pos += total {
			f1 := int(ReadUint(dec[pos : pos+w0]))
			f3 := int(ReadUint(dec[pos+w0+w1 : pos+total]))
			if f1 != 1 {
				continue
			}
			off := f3
			if off <= 0 || off >= len(data) {
				continue
			}
			endPos := FindEndObj(data, off)
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
			for objBodyEnd > objBodyStart && IsWhitespace(data[objBodyEnd-1]) {
				objBodyEnd--
			}
			if objBodyEnd <= objBodyStart {
				continue
			}
			if _, exists := objMap[onum]; !exists {
				objMap[onum] = data[objBodyStart:objBodyEnd]
				objGen[onum] = ogen
			}
		}
	}
}

// AugmentStrMap is the string-keyed ("num gen") variant of AugmentIntMap
// for the form package's object map.
func AugmentStrMap(data []byte, objMap map[string][]byte) {
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	for _, m := range objRe.FindAllSubmatch(data, -1) {
		body := m[3]
		if !bytes.Contains(body, []byte("/W[")) || !bytes.Contains(body, []byte("/Index")) {
			continue
		}
		sm := streamRe.FindSubmatch(body)
		if sm == nil {
			continue
		}
		dec, err := DecompressAny(sm[1])
		if err != nil {
			dec = sm[1]
		}
		W := ParseArrayInts(body, `/W`)
		if len(W) < 3 {
			continue
		}
		if ParseArrayInts(body, `/Index`) == nil {
			continue
		}
		w0, w1, w2 := W[0], W[1], W[2]
		total, ok := ValidXRefWidths(w0, w1, w2)
		if !ok {
			continue
		}
		for pos := 0; pos+total <= len(dec); pos += total {
			f1 := int(ReadUint(dec[pos : pos+w0]))
			f3 := int(ReadUint(dec[pos+w0+w1 : pos+total]))
			if f1 != 1 {
				continue
			}
			off := f3
			if off <= 0 || off >= len(data) {
				continue
			}
			tail := data[off:]
			if ro := objRe.FindSubmatch(tail); ro != nil {
				key := string(ro[1]) + " " + string(ro[2])
				if _, exists := objMap[key]; !exists {
					objMap[key] = ro[3]
				}
			}
		}
	}
}
