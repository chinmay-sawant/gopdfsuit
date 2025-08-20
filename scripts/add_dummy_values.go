package scripts

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

// This tool scans a PDF for widget annotation dictionaries (those containing
// "/Subtype/Widget" and "/Type/Annot") and, if the widget dictionary does
// not already contain a /V entry, inserts ` /V (dummy)` before the closing >>.

// AddDummyValues is callable from other code; it scans the default input flag paths
// and writes the output. It was originally a CLI main.
func AddDummyValues() {
	inPath := flag.String("in", "sampledata/patient2/patient2.pdf", "input PDF path")
	outPath := flag.String("out", "output.pdf", "output PDF path")
	flag.Parse()

	b, err := os.ReadFile(*inPath)
	if err != nil {
		log.Fatalf("failed to read input: %v", err)
	}

	inds := findWidgetDictRanges(b)
	if len(inds) == 0 {
		fmt.Println("No widget dictionaries found")
		// still write a copy
		if err := os.WriteFile(*outPath, b, 0644); err != nil {
			log.Fatalf("failed to write output: %v", err)
		}
		fmt.Printf("wrote copy to %s\n", *outPath)
		return
	}

	// Process from last to first so edits don't invalidate earlier offsets
	sort.Slice(inds, func(i, j int) bool { return inds[i].start > inds[j].start })
	out := b
	for _, r := range inds {
		seg := out[r.start:r.end]
		// skip if /V already present in this dict in any common form
		if bytes.Contains(seg, []byte("/V ")) || bytes.Contains(seg, []byte("/V(")) || bytes.Contains(seg, []byte("/V/")) {
			continue
		}
		// insert before the final '>>' that closes the dictionary
		// r.end points to the index of the '>>' end (i.e., end is exclusive of the '>>')
		// but our find returns the end index pointing to after '>>', adjust
		insertPos := r.end
		insertion := []byte(" /V (dummy)")
		out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
	}

	if err := os.WriteFile(*outPath, out, 0644); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
	fmt.Printf("wrote modified PDF to %s (modified %d widgets)\n", *outPath, len(inds))
}

type rng struct{ start, end int }

func findWidgetDictRanges(pdf []byte) []rng {
	needle := []byte("/Subtype/Widget")
	var ranges []rng
	// Manual scan: find all indices of the needle
	searchPos := 0
	for {
		idx := bytes.Index(pdf[searchPos:], needle)
		if idx < 0 {
			break
		}
		absIdx := searchPos + idx
		// build a context window around this occurrence; we'll scan backwards for '<<'
		ctxStart := absIdx - 4096
		if ctxStart < 0 {
			ctxStart = 0
		}
		// we only need ctxStart (backward scan limit)
		// find the nearest '<<' before absIdx using LastIndex on the slice
		rel := bytes.LastIndex(pdf[ctxStart:absIdx+1], []byte("<<"))
		if rel < 0 {
			searchPos = absIdx + 1
			continue
		}
		dictStart := ctxStart + rel

		// find matching >> for the dictionary from dictStart
		depth := 0
		j := dictStart
		dictEnd := -1
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
					dictEnd = j // position after the second '>'
					break
				}
				continue
			}
			j++
		}
		if dictEnd < 0 {
			searchPos = absIdx + 1
			continue
		}

		// quick sanity - ensure this dict contains '/Type/Annot' too
		dictBytes := pdf[dictStart:dictEnd]
		if !bytes.Contains(dictBytes, []byte("/Type/Annot")) && !bytes.Contains(dictBytes, []byte("/Type /Annot")) {
			searchPos = absIdx + 1
			continue
		}

		ranges = append(ranges, rng{start: dictStart, end: dictEnd})
		searchPos = absIdx + 1
	}

	return ranges
}
