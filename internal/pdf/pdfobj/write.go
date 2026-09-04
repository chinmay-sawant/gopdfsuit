package pdfobj

import (
	"bytes"
	"slices"
	"strconv"
)

// Style controls xref entry line endings. Merge and XFDF use CRLF with slightly
// different spacing; the main generator uses LF.
type Style struct {
	FreeLine    string // e.g. "0000000000 65535 f\r\n"
	InUseSuffix string // appended after the 10-digit offset, e.g. " 00000 n\r\n"
}

// MergeStyle matches the merge/compress dense xref line endings.
var MergeStyle = Style{
	FreeLine:    "0000000000 65535 f\r\n",
	InUseSuffix: " 00000 n\r\n",
}

// XFDFStyle matches the form XFDF xref line endings.
var XFDFStyle = Style{
	FreeLine:    "0000000000 65535 f \r\n",
	InUseSuffix: " 00000 n \r\n",
}

// GeneratorStyle matches the generator compact xref line endings.
var GeneratorStyle = Style{
	FreeLine:    "0000000000 65535 f \n",
	InUseSuffix: " 00000 n \n",
}

// WriteCompactXRef writes a compact PDF xref table using consecutive object
// subsections. scratch is reused for numeric formatting; pass nil to allocate.
// Returns the byte offset where "xref" starts (for startxref).
func WriteCompactXRef(out *bytes.Buffer, offsets map[int]int, scratch []byte, style Style) int {
	usedObjects := make([]int, 0, len(offsets)+1)
	usedObjects = append(usedObjects, 0)
	for objID := range offsets {
		usedObjects = append(usedObjects, objID)
	}
	slices.Sort(usedObjects)
	return WriteCompactXRefSorted(out, usedObjects, func(id int) (int, bool) {
		off, ok := offsets[id]
		return off, ok
	}, scratch, style)
}

// WriteCompactXRefSorted is the core behind WriteCompactXRef for callers that
// already hold sorted object IDs (e.g. the generator's pooled xref slice) and
// look offsets up through offsetOf. Output is identical to WriteCompactXRef.
func WriteCompactXRefSorted(out *bytes.Buffer, usedObjects []int, offsetOf func(int) (int, bool), scratch []byte, style Style) int {
	var subsections []struct{ start, count int }
	for i := 0; i < len(usedObjects); {
		start := usedObjects[i]
		count := 1
		for i+count < len(usedObjects) && usedObjects[i+count] == start+count {
			count++
		}
		subsections = append(subsections, struct{ start, count int }{start, count})
		i += count
	}

	xrefStart := out.Len()
	out.WriteString("xref\n")

	b := scratch
	if b == nil {
		b = make([]byte, 0, 32)
	}

	for _, sub := range subsections {
		b = b[:0]
		b = strconv.AppendInt(b, int64(sub.start), 10)
		b = append(b, ' ')
		b = strconv.AppendInt(b, int64(sub.count), 10)
		b = append(b, '\n')
		out.Write(b)

		for j := 0; j < sub.count; j++ {
			objID := sub.start + j
			if objID == 0 {
				out.WriteString(style.FreeLine)
				continue
			}
			if offset, exists := offsetOf(objID); exists {
				out.Write(AppendXRefOffset(b[:0], offset, style.InUseSuffix))
			} else {
				// Defense-in-depth: usedObjects derives from offset keys
				// so this lookup always hits today. Emit a free entry so
				// the entry count still matches the subsection header if
				// that invariant ever changes.
				out.WriteString(style.FreeLine)
			}
		}
	}

	return xrefStart
}

// WriteDenseXRef writes a dense `0 N` xref section covering objects 0..maxObj,
// emitting a free entry for gaps. This is the shared shape of the merge,
// compress, and XFDF full-rewrite writers; style selects their line endings.
func WriteDenseXRef(out *bytes.Buffer, maxObj int, offsetOf func(int) (int, bool), scratch []byte, style Style) {
	b := scratch
	if b == nil {
		b = make([]byte, 0, 32)
	}
	b = b[:0]
	b = append(b, "xref\n0 "...)
	b = strconv.AppendInt(b, int64(maxObj+1), 10)
	b = append(b, '\n')
	out.Write(b)

	for i := 0; i <= maxObj; i++ {
		if i == 0 {
			out.WriteString(style.FreeLine)
			continue
		}
		if off, ok := offsetOf(i); ok {
			out.Write(AppendXRefOffset(b[:0], off, style.InUseSuffix))
			continue
		}
		out.WriteString(style.FreeLine)
	}
}

// AppendXRefOffset appends a zero-padded 10-digit offset plus suffix.
func AppendXRefOffset(dst []byte, offset int, suffix string) []byte {
	b := strconv.AppendInt(dst, int64(offset), 10)
	padding := 10 - (len(b) - len(dst))
	if padding > 0 {
		b = b[:len(dst)+10]
		copy(b[len(dst)+padding:], b[len(dst):len(dst)+10-padding])
		for k := 0; k < padding; k++ {
			b[len(dst)+k] = '0'
		}
	}
	b = append(b, suffix...)
	return b
}

// IncrementalEntry is one object appended by an incremental update.
type IncrementalEntry struct {
	ID     int
	Offset int
	Gen    int
}

// WriteIncrementalXRef writes the xref section of an incremental update:
// subsection-grouped entries for exactly the appended objects in sortedIDs
// order, with real generation numbers. Trailer emission stays with the caller
// (it carries /Prev and /ID specifics).
func WriteIncrementalXRef(out *bytes.Buffer, sortedIDs []int, entryOf func(int) IncrementalEntry) {
	start := -1
	var block []int
	flushBlock := func() {
		if len(block) == 0 {
			return
		}
		var hdr [32]byte
		b := strconv.AppendInt(hdr[:0], int64(start), 10)
		b = append(b, ' ')
		b = strconv.AppendInt(b, int64(len(block)), 10)
		b = append(b, '\n')
		out.Write(b)
		for _, id := range block {
			e := entryOf(id)
			out.Write(AppendXRefEntry(nil, e.Offset, e.Gen))
		}
	}

	for _, id := range sortedIDs {
		if start < 0 {
			start = id
			block = []int{id}
			continue
		}
		if id == block[len(block)-1]+1 {
			block = append(block, id)
			continue
		}
		flushBlock()
		start = id
		block = []int{id}
	}
	flushBlock()
}

// AppendXRefEntry appends one "offset gen n \n" entry with a 10-digit offset
// and 5-digit generation, matching the incremental-update writer.
func AppendXRefEntry(dst []byte, offset, gen int) []byte {
	var buf [20]byte
	pos := 9
	off := offset
	for range 10 {
		buf[pos] = byte('0' + off%10)
		off /= 10
		pos--
	}
	buf[10] = ' '
	pos = 15
	g := gen
	for range 5 {
		buf[pos] = byte('0' + g%10)
		g /= 10
		pos--
	}
	buf[16] = ' '
	buf[17] = 'n'
	buf[18] = ' '
	buf[19] = '\n'
	return append(dst, buf[:]...)
}
