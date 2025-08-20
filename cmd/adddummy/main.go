package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// CLI tool: insert /V (dummy) into widget annotation dictionaries that lack /V
func main() {
	inPath := flag.String("in", "sampledata/patient2/patient2.pdf", "input PDF path")
	outPath := flag.String("out", "sampledata/patient2/output.pdf", "output PDF path")
	flag.Parse()

	b, err := os.ReadFile(*inPath)
	if err != nil {
		log.Fatalf("failed to read input: %v", err)
	}

	// attempt to load an XFDF alongside the PDF (same base name)
	xfdfPath := strings.TrimSuffix(*inPath, ".pdf") + ".xfdf"
	xfdfMap, err := parseXfdf(xfdfPath)
	if err != nil {
		fmt.Printf("debug: couldn't parse xfdf %s: %v\n", xfdfPath, err)
		xfdfMap = map[string]string{}
	}

	inds := findWidgetDictRanges(b)
	fmt.Printf("debug: found %d candidate widget dict ranges\n", len(inds))
	if len(inds) == 0 {
		fmt.Println("No widget dictionaries found")
		if err := os.WriteFile(*outPath, b, 0644); err != nil {
			log.Fatalf("failed to write output: %v", err)
		}
		fmt.Printf("wrote copy to %s\n", *outPath)
		return
	}

	sort.Slice(inds, func(i, j int) bool { return inds[i].start > inds[j].start })
	out := b
	modified := 0
	for _, r := range inds {
		fmt.Printf("debug: candidate range start=%d end=%d\n", r.start, r.end)
		seg := out[r.start:r.end]
		if bytes.Contains(seg, []byte("/V ")) || bytes.Contains(seg, []byte("/V(")) || bytes.Contains(seg, []byte("/V/")) {
			continue
		}
		// determine field name (T) inside this widget dict and look up xfdf map
		dictBytes := out[r.start:r.end]
		fieldName := extractFieldName(dictBytes)
		val := "dummy"
		if fieldName != "" {
			if v, ok := xfdfMap[fieldName]; ok {
				val = v
			}
		}
		// escape value for PDF literal string
		esc := escapePDFString(val)

		// choose literal string or name token depending on value
		var insertion []byte
		if isPDFName(val) {
			// name object (no parentheses)
			insertion = []byte(" /V /" + val)
		} else {
			// escape value for PDF literal string
			insertion = []byte(" /V (" + esc + ")")
		}

		// insert before the closing >> (r.end points just after the >>)
		insertPos := r.end - 2
		if insertPos < r.start {
			// sanity: fallback to r.end if computed insert position invalid
			insertPos = r.end
		}
		out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
		modified++
	}

	if err := os.WriteFile(*outPath, out, 0644); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
	fmt.Printf("wrote modified PDF to %s (modified %d widgets)\n", *outPath, modified)
}

// isPDFName returns true when s is safe to write as a PDF name object (no spaces or delimiter chars)
func isPDFName(s string) bool {
	if s == "" {
		return false
	}
	// PDF name must not contain whitespace or delimiters: ()<>[]/% or space
	if strings.IndexAny(s, "()<>[]/%\\ \t\n\r\f") >= 0 {
		return false
	}
	// additionally avoid names that start with a digit-only value like "123"? it's fine, allow it
	return true
}

type rng struct{ start, end int }

func findWidgetDictRanges(pdf []byte) []rng {
	needle := []byte("/Subtype/Widget")
	var ranges []rng
	searchPos := 0
	for {
		idx := bytes.Index(pdf[searchPos:], needle)
		if idx < 0 {
			break
		}
		absIdx := searchPos + idx
		// debug: show where needle found and a small surrounding context
		startCtx := absIdx - 40
		if startCtx < 0 {
			startCtx = 0
		}
		endCtx := absIdx + 40
		if endCtx > len(pdf) {
			endCtx = len(pdf)
		}
		fmt.Printf("debug: needle at %d context=%q\n", absIdx, pdf[startCtx:endCtx])
		// context window bounds (we'll scan backwards for the '<<' within this window)
		ctxStart := absIdx - 4096
		if ctxStart < 0 {
			ctxStart = 0
		}
		// only ctxStart is needed for backward scan limit

		// collect all '<<' positions in the context window and try each candidate
		var candidates []int
		slice := pdf[ctxStart : absIdx+1]
		off := 0
		for {
			r := bytes.Index(slice[off:], []byte("<<"))
			if r < 0 {
				break
			}
			candidates = append(candidates, ctxStart+off+r)
			off += r + 2
			if off >= len(slice) {
				break
			}
		}
		fmt.Printf("debug: candidates count=%d (nearest shown first)\n", len(candidates))
		if len(candidates) > 0 {
			// show up to 6 nearest candidate positions for context
			for i := len(candidates) - 1; i >= 0 && i >= len(candidates)-6; i-- {
				pos := candidates[i]
				ctxS := pos - 40
				if ctxS < 0 {
					ctxS = 0
				}
				ctxE := pos + 40
				if ctxE > len(pdf) {
					ctxE = len(pdf)
				}
				fmt.Printf("debug: candidate << at %d context=%q\n", pos, pdf[ctxS:ctxE])
			}
		}
		if len(candidates) == 0 {
			searchPos = absIdx + 1
			continue
		}

		// try candidates from nearest to farthest (reverse order)
		chosen := -1
		chosenEnd := -1
		for ci := len(candidates) - 1; ci >= 0; ci-- {
			cand := candidates[ci]
			// find matching >> for the dictionary from cand
			depth := 0
			j := cand
			var dictEnd int
			dictEnd = -1
			for j < len(pdf)-1 {
				if j+1 < len(pdf) && pdf[j] == '<' && pdf[j+1] == '<' {
					depth++
					j += 2
					continue
				}
				if j+1 < len(pdf) && pdf[j] == '>' && pdf[j+1] == '>' {
					depth--
					j += 2
					if depth == 0 {
						dictEnd = j
						break
					}
					continue
				}
				j++
			}
			// log candidate outcome
			if dictEnd < 0 {
				fmt.Printf("debug: cand=%d -> no matching >> found\n", cand)
				continue
			}
			if dictEnd <= absIdx {
				fmt.Printf("debug: cand=%d -> dictEnd=%d is before needle %d (skip)\n", cand, dictEnd, absIdx)
				continue
			}
			dictBytes := pdf[cand:dictEnd]
			// prepare snippet for logging
			snippet := dictBytes
			if len(snippet) > 200 {
				snippet = snippet[:200]
			}
			hasTypeAnnot := bytes.Contains(dictBytes, []byte("/Type/Annot")) || bytes.Contains(dictBytes, []byte("/Type /Annot"))
			fmt.Printf("debug: cand=%d dictEnd=%d hasTypeAnnot=%v dictSnippet=%q\n", cand, dictEnd, hasTypeAnnot, snippet)
			if !hasTypeAnnot {
				continue
			}
			chosen = cand
			chosenEnd = dictEnd
		}
		if chosen < 0 {
			searchPos = absIdx + 1
			continue
		}
		dictStart := chosen
		dictEnd := chosenEnd
		ranges = append(ranges, rng{start: dictStart, end: dictEnd})
		searchPos = absIdx + 1
	}

	return ranges
}

// (helper removed)

// parseXfdf reads a .xfdf file and returns a map[fieldName]value
func parseXfdf(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(bytes.NewReader(b))
	m := map[string]string{}
	type XField struct {
		XMLName xml.Name `xml:"field"`
		Name    string   `xml:"name,attr"`
		Value   string   `xml:"value"`
	}
	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF { // io isn't imported now; replace with nil check
				break
			}
			return nil, err
		}
		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "field" {
			var xf XField
			if err := dec.DecodeElement(&xf, &se); err != nil {
				// ignore individual decode errors
				continue
			}
			if xf.Name != "" {
				m[xf.Name] = xf.Value
			}
		}
	}
	return m, nil
}

// extractFieldName finds /T(fieldname) inside a dictionary blob and returns the fieldname (unescaped)
func extractFieldName(dict []byte) string {
	idx := bytes.Index(dict, []byte("/T("))
	if idx < 0 {
		return ""
	}
	start := idx + 3
	var sb strings.Builder
	for i := start; i < len(dict); i++ {
		b := dict[i]
		if b == ')' {
			return sb.String()
		}
		if b == '\\' && i+1 < len(dict) {
			sb.WriteByte(dict[i+1])
			i++
			continue
		}
		sb.WriteByte(b)
	}
	return sb.String()
}

// escapePDFString escapes parentheses and backslashes for a PDF literal string
func escapePDFString(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return r.Replace(s)
}
