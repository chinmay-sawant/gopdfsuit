package main

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// fielddetect.go
// Simple, self-contained heuristic detector for PDF AcroForm fields.
// It does NOT fully parse PDF objects or resolve indirect references.
// Instead it scans the raw PDF bytes for /T (field name) tokens and
// looks nearby for /V or /AS tokens to heuristically extract values.

func decodeHexString(s string) string {
	// remove whitespace
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	if len(s)%2 == 1 {
		// odd length, pad with 0
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return "<invalid hex>"
	}
	// PDF text is usually ASCII or Latin-1; we'll convert bytes to string directly
	return string(b)
}

func extractTokenGroups(content []byte, pos int) (string, string) {
	// Look ahead up to a limit for /V or /AS tokens
	limit := pos + 800
	if limit > len(content) {
		limit = len(content)
	}
	window := content[pos:limit]

	// regex matching literal string (parens), hex <...>, or name /Name
	valueRe := regexp.MustCompile(`/V\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	asRe := regexp.MustCompile(`/AS\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)

	if m := valueRe.FindSubmatch(window); m != nil {
		if len(m[2]) > 0 {
			return "V", string(m[2])
		}
		if len(m[3]) > 0 {
			return "V", decodeHexString(string(m[3]))
		}
		if len(m[4]) > 0 {
			return "V", string(m[4])
		}
	}
	if m := asRe.FindSubmatch(window); m != nil {
		if len(m[2]) > 0 {
			return "AS", string(m[2])
		}
		if len(m[3]) > 0 {
			return "AS", decodeHexString(string(m[3]))
		}
		if len(m[4]) > 0 {
			return "AS", string(m[4])
		}
	}
	return "", ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fielddetect <file.pdf>")
		fmt.Println("Default: sampledata/patientreg/patientreg.pdf")
		os.Exit(1)
	}

	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Opened PDF %s (%d bytes)\n", path, len(data))

	// Build map of indirect objects: "<obj> <gen>" -> body
	// Use (?s) to allow dot to match newlines
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	objMatches := objRe.FindAllSubmatch(data, -1)
	if len(objMatches) == 0 {
		// Fall back to naive scan
		tRe := regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
		matches := tRe.FindAllSubmatchIndex(data, -1)
		if len(matches) == 0 {
			fmt.Println("No /T field tokens found in PDF.")
			os.Exit(0)
		}
		fmt.Printf("Detected %d candidate fields (flat scan):\n", len(matches))
		seen := make(map[string]bool)
		for _, mi := range matches {
			var name string
			if mi[4] != -1 && mi[5] != -1 {
				name = string(data[mi[4]:mi[5]])
			} else if mi[6] != -1 && mi[7] != -1 {
				name = decodeHexString(string(data[mi[6]:mi[7]]))
			} else if mi[8] != -1 && mi[9] != -1 {
				name = string(data[mi[8]:mi[9]])
			} else {
				continue
			}
			name = strings.TrimSpace(name)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			endPos := mi[1]
			tType, val := extractTokenGroups(data, endPos)
			if tType == "" {
				fmt.Printf("- Key: %s  Value: <not found nearby>\n", name)
			} else {
				fmt.Printf("- Key: %s  %s: %s\n", name, tType, val)
			}
		}
		os.Exit(0)
	}

	objMap := make(map[string][]byte)
	for _, m := range objMatches {
		// m[1]=objnum, m[2]=gen, m[3]=body
		key := string(m[1]) + " " + string(m[2])
		body := m[3]
		// attempt to locate any stream...endstream sections and decompress when Flate
		streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
		newBody := body
		changed := false
		for _, sm := range streamRe.FindAllSubmatchIndex(body, -1) {
			// sm gives start/end indices for full match and capture group
			sStart := sm[2]
			sEnd := sm[3]
			if sStart < 0 || sEnd < 0 || sEnd <= sStart {
				continue
			}
			streamBytes := body[sStart:sEnd]
			// try zlib then raw flate
			var dec []byte
			if d, err := tryZlibDecompress(streamBytes); err == nil {
				dec = d
			} else if d, err := tryFlateDecompress(streamBytes); err == nil {
				dec = d
			}
			if dec != nil {
				// replace the stream bytes with the decompressed content (as plain bytes)
				// build newBody using bytes.Buffer
				var buf bytes.Buffer
				buf.Write(newBody[:sm[0]])
				buf.Write(dec)
				buf.Write(newBody[sm[1]:])
				newBody = buf.Bytes()
				changed = true
				// continue but avoid overlapping indexes (we restart search on newBody)
				break
			}
		}
		if changed {
			// if we changed, re-run to decompress any further streams
			for {
				found := false
				for _, sm := range streamRe.FindAllSubmatchIndex(newBody, -1) {
					sStart := sm[2]
					sEnd := sm[3]
					if sStart < 0 || sEnd < 0 || sEnd <= sStart {
						continue
					}
					streamBytes := newBody[sStart:sEnd]
					var dec []byte
					if d, err := tryZlibDecompress(streamBytes); err == nil {
						dec = d
					} else if d, err := tryFlateDecompress(streamBytes); err == nil {
						dec = d
					}
					if dec != nil {
						var buf bytes.Buffer
						buf.Write(newBody[:sm[0]])
						buf.Write(dec)
						buf.Write(newBody[sm[1]:])
						newBody = buf.Bytes()
						found = true
						break
					}
				}
				if !found {
					break
				}
			}
		}
		objMap[key] = newBody
	}

	fmt.Printf("Parsed %d indirect objects from PDF.\n", len(objMap))

	// Helper to find refs in a body
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)

	found := make(map[string]bool)

	// We'll inspect any object that either contains /T or is referenced by
	// /Annots, /Fields or /Kids arrays. First gather candidate objects.
	candidateKeys := make(map[string]bool)
	for k, body := range objMap {
		if bytesIndex(body, []byte(`/T`)) >= 0 || bytesIndex(body, []byte(`/Annots`)) >= 0 || bytesIndex(body, []byte(`/Fields`)) >= 0 || bytesIndex(body, []byte(`/Kids`)) >= 0 {
			candidateKeys[k] = true
			// also collect direct refs inside this body
			for _, r := range refRe.FindAllSubmatch(body, -1) {
				refKey := string(r[1]) + " " + string(r[2])
				candidateKeys[refKey] = true
			}
		}
	}

	// tRe for detecting /T inside an object body
	tRe := regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)

	fmt.Printf("Detected %d candidate objects to scan for fields.\n", len(candidateKeys))
	// Also scan all objects as a fallback (no raw dumps)

	fmt.Println("\nFIELDS:")
	for k, body := range objMap {
		submatches := tRe.FindAllSubmatchIndex(body, -1)
		for _, mi := range submatches {
			var name string
			if mi[4] != -1 && mi[5] != -1 {
				name = string(body[mi[4]:mi[5]])
			} else if mi[6] != -1 && mi[7] != -1 {
				name = decodeHexString(string(body[mi[6]:mi[7]]))
			} else if mi[8] != -1 && mi[9] != -1 {
				name = string(body[mi[8]:mi[9]])
			} else {
				continue
			}
			name = strings.TrimSpace(name)
			if name == "" || found[name] {
				continue
			}
			found[name] = true

			// look for /V or /AS inside this same object first
			endPos := mi[1]
			tType, val := extractTokenGroups(body, endPos)
			if tType == "" {
				// if not found, search referenced objects (Kids/Annots)
				refs := refRe.FindAllSubmatch(body, -1)
				for _, r := range refs {
					refKey := string(r[1]) + " " + string(r[2])
					if rb, ok := objMap[refKey]; ok {
						rt, rv := extractTokenGroups(rb, 0)
						if rt != "" {
							tType, val = rt, rv
							break
						}
					}
				}
			}

			if tType == "" {
				fmt.Printf("- Obj %s Key: %s  Value: <not found>\n", k, name)
			} else {
				fmt.Printf("- Obj %s Key: %s  %s: %s\n", k, name, tType, val)
			}
		}
	}
}

// bytesIndex is a tiny helper to find a subsequence in a []byte without
// importing bytes package repeatedly; keeps code simple.
func bytesIndex(b, sub []byte) int {
	return strings.Index(string(b), string(sub))
}

func tryZlibDecompress(b []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func tryFlateDecompress(b []byte) ([]byte, error) {
	// flate.NewReader expects raw DEFLATE stream
	r := flate.NewReader(bytes.NewReader(b))
	defer r.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
