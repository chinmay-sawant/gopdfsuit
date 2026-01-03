package merge

import (
	"bytes"
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
	matches := objStartRe.FindAllSubmatchIndex(data, -1)

	for _, m := range matches {
		objNum, _ := strconv.Atoi(string(data[m[2]:m[3]]))
		genNum, _ := strconv.Atoi(string(data[m[4]:m[5]]))
		start := m[0]
		bodyStart := m[1]

		endPos := FindEndObj(data, bodyStart)
		if endPos == -1 {
			continue
		}

		results = append(results, ObjectBoundary{
			ObjNum:    objNum,
			GenNum:    genNum,
			Start:     start,
			BodyStart: bodyStart,
			End:       endPos,
		})
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
			endstreamPos := bytes.Index(data[i:], []byte("endstream"))
			if endstreamPos == -1 {
				return -1
			}
			i = i + endstreamPos + 9
			continue
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
		streamIdx := bytes.Index(data[i:], []byte("stream"))
		if streamIdx == -1 {
			tail := data[i:]
			replaced := refRe.ReplaceAllFunc(tail, replaceFunc)
			out.Write(replaced)
			break
		}

		streamPos := i + streamIdx
		pre := data[i:streamPos]
		replaced := refRe.ReplaceAllFunc(pre, replaceFunc)
		out.Write(replaced)

		endstreamIdx := bytes.Index(data[streamPos:], []byte("endstream"))
		if endstreamIdx == -1 {
			out.Write(data[streamPos:])
			break
		}

		endPos := streamPos + endstreamIdx + 9
		out.Write(data[streamPos:endPos])
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

// isWhitespace checks if byte is PDF whitespace
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
