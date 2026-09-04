package xref

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

// TestWriteCompactXRefEntryCountsMatchHeaders asserts every subsection header
// count equals the number of entry lines that follow it, including a gap case
// (object 3 missing between 2 and 4 forces separate subsections).
func TestWriteCompactXRefEntryCountsMatchHeaders(t *testing.T) {
	offsets := map[int]int{1: 10, 2: 20, 4: 40, 6: 60}

	var out bytes.Buffer
	WriteCompactXRef(&out, offsets, nil, GeneratorStyle)

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if lines[0] != "xref" {
		t.Fatalf("first line = %q, want xref", lines[0])
	}
	i := 1
	totalEntries := 0
	for i < len(lines) {
		fields := strings.Fields(lines[i])
		if len(fields) != 2 {
			t.Fatalf("bad subsection header %q", lines[i])
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil || count <= 0 {
			t.Fatalf("bad subsection count in %q", lines[i])
		}
		i++
		for j := 0; j < count; j++ {
			if i >= len(lines) {
				t.Fatalf("header %q promises %d entries, output ends early", lines[i-j-1], count)
			}
			entry := lines[i]
			if !(strings.HasSuffix(entry, "f ") || strings.HasSuffix(entry, "n ")) {
				t.Fatalf("bad xref entry %q", entry)
			}
			i++
		}
		totalEntries += count
	}
	// Object 0 free entry + 4 used objects = 5 entries across subsections.
	if totalEntries != len(offsets)+1 {
		t.Fatalf("total entries = %d, want %d", totalEntries, len(offsets)+1)
	}
}
