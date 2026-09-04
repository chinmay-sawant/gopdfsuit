// Package form provides functionality for parsing XFDF and filling PDF forms.
package form

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"strings"
)

// XFDF structures for minimal parsing
type xfdfField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

type xfdfRoot struct {
	XMLName xml.Name    `xml:"xfdf"`
	Fields  []xfdfField `xml:"fields>field"`
}

// MaxXFDFBytes caps accepted XFDF input (4 MiB); larger payloads are rejected.
const MaxXFDFBytes = 4 << 20

// ParseXFDF parses XFDF bytes and returns a map of field name -> value.
// Inputs over MaxXFDFBytes or containing <!ENTITY/<!DOCTYPE declarations
// are rejected: Go's encoding/xml expands internal entities (billion-laughs).
func ParseXFDF(xfdfBytes []byte) (map[string]string, error) {
	if len(xfdfBytes) > MaxXFDFBytes {
		return nil, errors.New("xfdf input exceeds 4 MiB limit")
	}
	if bytes.Contains(xfdfBytes, []byte("<!ENTITY")) || bytes.Contains(xfdfBytes, []byte("<!DOCTYPE")) {
		return nil, errors.New("xfdf input with DOCTYPE/ENTITY declarations is rejected")
	}
	var root xfdfRoot
	if err := xml.Unmarshal(xfdfBytes, &root); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, f := range root.Fields {
		name := strings.TrimSpace(f.Name)
		val := strings.TrimSpace(f.Value)
		m[name] = val
	}
	return m, nil
}

// Field represents a detected or targetable PDF form field.
type Field struct {
	Name  string
	Value string
	Type  string // V, AS, or detected type
}

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return bytes.Index(b, sub)
}

// decodeHexString converts hex string to regular string
func decodeHexString(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	if len(s)%2 == 1 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return "<invalid hex>"
	}
	return string(b)
}

// Helper to build minimal XFDF XML from field map
func buildXFDF(fields map[string]string) []byte {
	type xfdfField struct {
		XMLName xml.Name `xml:"field"`
		Name    string   `xml:"name,attr"`
		Value   string   `xml:"value"`
	}
	type xfdfRoot struct {
		XMLName xml.Name    `xml:"xfdf"`
		XMLNS   string      `xml:"xmlns,attr,omitempty"`
		Fields  []xfdfField `xml:"fields>field"`
	}
	root := xfdfRoot{XMLNS: "http://ns.adobe.com/xfdf/", Fields: make([]xfdfField, 0, len(fields))}
	for k, v := range fields {
		root.Fields = append(root.Fields, xfdfField{Name: k, Value: v})
	}
	out, _ := xml.Marshal(root)
	return out
}

// unescapePDFString reverses escapePDFString for literal PDF strings.
func unescapePDFString(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			sb.WriteByte(s[i])
			continue
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// escapePDFString escapes characters as required for PDF literal strings.
func escapePDFString(s string) string {
	// Fast path: most text has no special characters
	if !strings.ContainsAny(s, `()\`) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s) + 4)
	for _, r := range s {
		switch r {
		case '(', ')', '\\':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
