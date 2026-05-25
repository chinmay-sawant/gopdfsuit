package xref

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

// MergeStyle matches internal/pdf/merge writeXRefAndTrailer line endings.
var MergeStyle = Style{
	FreeLine:    "0000000000 65535 f\r\n",
	InUseSuffix: " 00000 n\r\n",
}

// XFDFStyle matches internal/pdf/form XFDF xref line endings.
var XFDFStyle = Style{
	FreeLine:    "0000000000 65535 f \r\n",
	InUseSuffix: " 00000 n \r\n",
}

// GeneratorStyle matches internal/pdf/generator compact xref line endings.
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

	xrefStart := out.Len()
	out.WriteString("xref\n")

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
			if offset, exists := offsets[objID]; exists {
				b = b[:0]
				b = strconv.AppendInt(b, int64(offset), 10)
				padding := 10 - len(b)
				if padding > 0 {
					b = b[:10]
					copy(b[padding:], b[:10-padding])
					for k := 0; k < padding; k++ {
						b[k] = '0'
					}
				}
				b = append(b, style.InUseSuffix...)
				out.Write(b)
			}
		}
	}

	return xrefStart
}
