package merge

import (
	"bytes"
	"compress/zlib"
	"io"
	"regexp"
	"strconv"
)

// PDF parsing functions for the merge package

// DetectPDFVersion extracts the PDF version from the header (e.g., "1.4", "1.7", "2.0")
func DetectPDFVersion(data []byte) string {
	versionRe := regexp.MustCompile(`%PDF-(\d+\.\d+)`)
	if m := versionRe.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return "1.7" // default fallback
}

// CompareVersions compares two PDF version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func CompareVersions(v1, v2 string) int {
	parse := func(v string) (int, int) {
		parts := bytes.Split([]byte(v), []byte("."))
		major, _ := strconv.Atoi(string(parts[0]))
		minor := 0
		if len(parts) > 1 {
			minor, _ = strconv.Atoi(string(parts[1]))
		}
		return major, minor
	}

	maj1, min1 := parse(v1)
	maj2, min2 := parse(v2)

	if maj1 != maj2 {
		if maj1 > maj2 {
			return 1
		}
		return -1
	}
	if min1 > min2 {
		return 1
	} else if min1 < min2 {
		return -1
	}
	return 0
}

// FindObjectBoundaries finds all PDF objects in the data
func FindObjectBoundaries(data []byte) []ObjectBoundary {
	var results []ObjectBoundary
	objStartRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)

	pos := 0
	for pos < len(data) {
		// Find next "obj" starting from pos
		loc := objStartRe.FindSubmatchIndex(data[pos:])
		if loc == nil {
			break
		}

		// Adjust indices to be absolute
		start := pos + loc[0]
		bodyStart := pos + loc[1]

		objNum, _ := strconv.Atoi(string(data[pos+loc[2] : pos+loc[3]]))
		genNum, _ := strconv.Atoi(string(data[pos+loc[4] : pos+loc[5]]))

		endPos := FindEndObj(data, bodyStart)
		if endPos == -1 {
			// Failed to find end. Skip this match and try next char?
			pos = start + 1
			continue
		}

		results = append(results, ObjectBoundary{
			ObjNum:    objNum,
			GenNum:    genNum,
			Start:     start,
			BodyStart: bodyStart,
			End:       endPos,
		})

		// Continue searching after this object
		pos = endPos
	}

	return results
}

// FindEndObj finds the position right after "endobj" starting from pos
func FindEndObj(data []byte, pos int) int {
	i := pos
	n := len(data)

	for i < n {
		// Skip whitespace
		for i < n && isWhitespace(data[i]) {
			i++
		}
		if i >= n {
			break
		}

		// Check for "endobj"
		if i+6 <= n && string(data[i:i+6]) == "endobj" {
			return i + 6
		}

		// Check for stream
		if i+6 <= n && string(data[i:i+6]) == "stream" {
			// Verify stream is followed by EOL
			if i+6 < n {
				b := data[i+6]
				if b == '\r' || b == '\n' {
					// It is a stream start.
					// Find endstream
					searchPos := i + 6
					for {
						idx := bytes.Index(data[searchPos:], []byte("endstream"))
						if idx == -1 {
							return -1 // Stream not closed
						}

						matchPos := searchPos + idx

						// Check if preceded by EOL
						validEnd := false
						if matchPos > 0 {
							if data[matchPos-1] == '\n' {
								validEnd = true
							} else if data[matchPos-1] == '\r' {
								validEnd = true
							}
						}

						if validEnd {
							i = matchPos + 9 // len("endstream")
							break
						}

						// Not a valid endstream (not preceded by EOL), continue searching
						searchPos = matchPos + 1
					}
					continue
				}
			}
			// If not followed by EOL, treat as normal text
		}

		switch {
		case data[i] == '(':
			i = SkipStringLiteral(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] != '<':
			i = SkipHexString(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] == '<':
			i = SkipDictionary(data, i)
		case data[i] == '[':
			i = SkipArray(data, i)
		default:
			i++
		}
	}

	return -1
}

// SkipStringLiteral skips a PDF string literal (...) handling escapes and nested parens
func SkipStringLiteral(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != '(' {
		return pos + 1
	}
	i := pos + 1
	depth := 1
	for i < len(data) && depth > 0 {
		if data[i] == '\\' && i+1 < len(data) {
			i += 2 // skip escaped character
			continue
		}
		switch data[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		i++
	}
	return i
}

// SkipHexString skips a PDF hex string <...>
func SkipHexString(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != '<' {
		return pos + 1
	}
	i := pos + 1
	for i < len(data) && data[i] != '>' {
		i++
	}
	return i + 1
}

// SkipDictionary skips a PDF dictionary <<...>>
func SkipDictionary(data []byte, pos int) int {
	if pos+1 >= len(data) || data[pos] != '<' || data[pos+1] != '<' {
		return pos + 1
	}
	i := pos + 2
	depth := 1
	for i < len(data) && depth > 0 {
		switch {
		case data[i] == '(':
			i = SkipStringLiteral(data, i)
		case data[i] == '<':
			if i+1 < len(data) && data[i+1] == '<' {
				depth++
				i += 2
			} else {
				i = SkipHexString(data, i)
			}
		case data[i] == '>' && i+1 < len(data) && data[i+1] == '>':
			depth--
			i += 2
		default:
			i++
		}
	}
	return i
}

// SkipArray skips a PDF array [...]
func SkipArray(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != '[' {
		return pos + 1
	}
	i := pos + 1
	depth := 1
	for i < len(data) && depth > 0 {
		if data[i] == '(' {
			i = SkipStringLiteral(data, i)
			continue
		}
		if data[i] == '<' {
			if i+1 < len(data) && data[i+1] == '<' {
				i = SkipDictionary(data, i)
			} else {
				i = SkipHexString(data, i)
			}
			continue
		}
		switch data[i] {
		case '[':
			depth++
		case ']':
			depth--
		}
		i++
	}
	return i
}

// FindStreamStart finds the start of a stream in data, skipping strings/dicts
// Returns index of "stream" keyword, or -1
func FindStreamStart(data []byte) int {
	i := 0
	n := len(data)
	for i < n {
		// Check for stream
		if i+6 <= n && string(data[i:i+6]) == "stream" {
			// Check EOL
			if i+6 < n {
				b := data[i+6]
				if b == '\r' || b == '\n' {
					return i
				}
			}
		}

		switch {
		case data[i] == '(':
			i = SkipStringLiteral(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] != '<':
			i = SkipHexString(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] == '<':
			i = SkipDictionary(data, i)
		case data[i] == '[':
			i = SkipArray(data, i)
		default:
			i++
		}
	}
	return -1
}

// ReplaceRefsOutsideStreams rewrites indirect references only outside stream blocks
func ReplaceRefsOutsideStreams(data []byte, offset int) []byte {
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	var out bytes.Buffer
	i := 0
	n := len(data)

	replaceFunc := func(b []byte) []byte {
		sm := refRe.FindSubmatch(b)
		if len(sm) < 3 {
			return b
		}
		on, err := strconv.Atoi(string(sm[1]))
		if err != nil {
			return b
		}
		gen := string(sm[2])
		return []byte(strconv.Itoa(offset+on) + " " + gen + " R")
	}

	for i < n {
		// Find next "stream" keyword properly
		relStart := FindStreamStart(data[i:])
		if relStart == -1 {
			// No more streams, process rest
			tail := data[i:]
			replaced := refRe.ReplaceAllFunc(tail, replaceFunc)
			out.Write(replaced)
			break
		}

		streamStart := i + relStart

		// Process pre-stream
		pre := data[i:streamStart]
		replaced := refRe.ReplaceAllFunc(pre, replaceFunc)
		out.Write(replaced)

		// Find endstream
		// Skip "stream" and EOL
		ptr := streamStart + 6
		if ptr < n && data[ptr] == '\r' {
			ptr++
		}
		if ptr < n && data[ptr] == '\n' {
			ptr++
		}

		// Now we are in the stream content. Find "endstream"
		endstreamIdx := -1
		searchPos := ptr
		for {
			idx := bytes.Index(data[searchPos:], []byte("endstream"))
			if idx == -1 {
				break // Should not happen if object is valid
			}

			pos := searchPos + idx

			// Check if preceded by EOL
			valid := false
			if pos > 0 {
				b := data[pos-1]
				if b == '\r' || b == '\n' {
					valid = true
				}
			}

			if valid {
				endstreamIdx = pos
				break
			}

			searchPos = pos + 9 // len("endstream")
		}

		if endstreamIdx == -1 {
			// Could not find endstream, just write the rest raw (safety fallback)
			out.Write(data[streamStart:])
			break
		}

		endPos := endstreamIdx + 9 // len("endstream")
		out.Write(data[streamStart:endPos])
		i = endPos
	}

	return out.Bytes()
}

// HasSubstring checks if data contains substring
func HasSubstring(data, sub []byte) bool {
	return bytes.Contains(data, sub)
}

// IsPageObject checks if the object body is a Page object
func IsPageObject(body []byte) bool {
	return HasSubstring(body, []byte("/Type /Page")) ||
		HasSubstring(body, []byte("/Type/Page")) ||
		(HasSubstring(body, []byte("/MediaBox")) && !IsPagesTreeObject(body))
}

// IsPagesTreeObject checks if the object is a Pages tree node
func IsPagesTreeObject(body []byte) bool {
	return HasSubstring(body, []byte("/Type /Pages")) ||
		HasSubstring(body, []byte("/Type/Pages"))
}

// IsWidgetAnnotation checks if the object is a Widget annotation
func IsWidgetAnnotation(body []byte) bool {
	return HasSubstring(body, []byte("/Subtype /Widget")) ||
		HasSubstring(body, []byte("/Subtype/Widget"))
}

// IsFormField checks if the object has form field type
func IsFormField(body []byte) bool {
	return HasSubstring(body, []byte("/FT /")) ||
		HasSubstring(body, []byte("/FT/"))
}

// IsXObjectForm checks if object is a Form XObject (appearance stream)
func IsXObjectForm(body []byte) bool {
	return HasSubstring(body, []byte("/Type /XObject")) &&
		HasSubstring(body, []byte("/Subtype /Form"))
}

// IsObjectStream checks if the object is an Object Stream (ObjStm)
func IsObjectStream(body []byte) bool {
	return HasSubstring(body, []byte("/Type /ObjStm")) ||
		HasSubstring(body, []byte("/Type/ObjStm"))
}

// isWhitespace checks if byte is PDF whitespace
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// ParseObjectStream extracts objects from a compressed object stream
// Object streams contain multiple objects in a compressed format
// Format: /N <count> /First <offset> followed by stream with:
//   - Header: pairs of "objnum offset" separated by whitespace
//   - Body: object bodies starting at /First offset
func ParseObjectStream(body []byte) map[int][]byte {
	result := make(map[int][]byte)

	// Extract /N (number of objects)
	nRe := regexp.MustCompile(`/N\s+(\d+)`)
	nMatch := nRe.FindSubmatch(body)
	if nMatch == nil {
		return result
	}
	numObjects, _ := strconv.Atoi(string(nMatch[1]))
	if numObjects == 0 {
		return result
	}

	// Extract /First (offset to first object body)
	firstRe := regexp.MustCompile(`/First\s+(\d+)`)
	firstMatch := firstRe.FindSubmatch(body)
	if firstMatch == nil {
		return result
	}
	firstOffset, _ := strconv.Atoi(string(firstMatch[1]))

	// Find and decompress stream
	streamData := extractAndDecompressStream(body)
	if streamData == nil || len(streamData) < firstOffset {
		return result
	}

	// Parse header: pairs of "objnum offset"
	header := streamData[:firstOffset]
	objectData := streamData[firstOffset:]

	// Parse object number and offset pairs
	type objEntry struct {
		objNum int
		offset int
	}
	var entries []objEntry

	headerStr := string(bytes.TrimSpace(header))
	parts := regexp.MustCompile(`\s+`).Split(headerStr, -1)

	for i := 0; i+1 < len(parts); i += 2 {
		objNum, err1 := strconv.Atoi(parts[i])
		offset, err2 := strconv.Atoi(parts[i+1])
		if err1 == nil && err2 == nil {
			entries = append(entries, objEntry{objNum: objNum, offset: offset})
		}
	}

	// Extract each object body
	for i, entry := range entries {
		start := entry.offset
		var end int
		if i+1 < len(entries) {
			end = entries[i+1].offset
		} else {
			end = len(objectData)
		}

		if start >= 0 && end <= len(objectData) && start < end {
			objBody := bytes.TrimSpace(objectData[start:end])
			result[entry.objNum] = objBody
		}
	}

	return result
}

// extractAndDecompressStream extracts and decompresses a stream from an object body
func extractAndDecompressStream(body []byte) []byte {
	// Find stream start
	streamStart := FindStreamStart(body)
	if streamStart == -1 {
		return nil
	}

	// Skip "stream" and EOL
	ptr := streamStart + 6
	if ptr < len(body) && body[ptr] == '\r' {
		ptr++
	}
	if ptr < len(body) && body[ptr] == '\n' {
		ptr++
	}

	// Find endstream
	endstreamIdx := bytes.Index(body[ptr:], []byte("endstream"))
	if endstreamIdx == -1 {
		return nil
	}

	streamContent := body[ptr : ptr+endstreamIdx]

	// Trim trailing EOL before endstream
	for len(streamContent) > 0 {
		last := streamContent[len(streamContent)-1]
		if last == '\r' || last == '\n' {
			streamContent = streamContent[:len(streamContent)-1]
		} else {
			break
		}
	}

	// Check if FlateDecode is used
	if HasSubstring(body[:streamStart], []byte("/Filter /FlateDecode")) ||
		HasSubstring(body[:streamStart], []byte("/Filter/FlateDecode")) {
		// Decompress
		return decompressFlate(streamContent)
	}

	// No compression (or unsupported), return as-is
	return streamContent
}

// decompressFlate decompresses zlib/deflate data
func decompressFlate(data []byte) []byte {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer reader.Close()

	var out bytes.Buffer
	_, err = io.Copy(&out, reader)
	if err != nil {
		return nil
	}

	return out.Bytes()
}
