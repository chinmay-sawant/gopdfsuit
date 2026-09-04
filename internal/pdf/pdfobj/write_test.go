package pdfobj

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestWriteCompactXRefSubsections(t *testing.T) {
	offsets := map[int]int{
		1: 100,
		2: 200,
		5: 500,
		7: 700,
		8: 800,
	}

	var out bytes.Buffer
	start := WriteCompactXRef(&out, offsets, nil, MergeStyle)
	if start != 0 {
		t.Fatalf("expected xref at start of buffer, got %d", start)
	}

	got := out.String()
	wantParts := []string{
		"xref\n",
		"0 3\n",
		"0000000000 65535 f\r\n",
		"0000000100 00000 n\r\n",
		"0000000200 00000 n\r\n",
		"5 1\n",
		"0000000500 00000 n\r\n",
		"7 2\n",
		"0000000700 00000 n\r\n",
		"0000000800 00000 n\r\n",
	}
	for _, part := range wantParts {
		if !bytes.Contains([]byte(got), []byte(part)) {
			t.Fatalf("missing xref subsection fragment %q in:\n%s", part, got)
		}
	}
}

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

func TestWriteCompactXRefSortedMatchesMap(t *testing.T) {
	offsets := map[int]int{1: 100, 2: 200, 5: 500, 7: 700, 8: 800}
	var want bytes.Buffer
	WriteCompactXRef(&want, offsets, nil, GeneratorStyle)

	used := []int{0, 1, 2, 5, 7, 8}
	var got bytes.Buffer
	WriteCompactXRefSorted(&got, used, func(id int) (int, bool) {
		off, ok := offsets[id]
		return off, ok
	}, nil, GeneratorStyle)

	if got.String() != want.String() {
		t.Fatalf("sorted core differs from map wrapper:\n%q\n%q", got.String(), want.String())
	}
}

func TestWriteDenseXRefMergeShape(t *testing.T) {
	offsets := map[int]int{1: 100, 3: 300}
	var out bytes.Buffer
	WriteDenseXRef(&out, 3, func(id int) (int, bool) {
		off, ok := offsets[id]
		return off, ok
	}, nil, MergeStyle)

	want := "xref\n0 4\n" +
		"0000000000 65535 f\r\n" +
		"0000000100 00000 n\r\n" +
		"0000000000 65535 f\r\n" +
		"0000000300 00000 n\r\n"
	if out.String() != want {
		t.Fatalf("dense merge xref differs:\n%q\n%q", out.String(), want)
	}
}

func TestWriteDenseXRefXFDFShape(t *testing.T) {
	offsets := map[int]int{1: 10}
	var out bytes.Buffer
	WriteDenseXRef(&out, 2, func(id int) (int, bool) {
		off, ok := offsets[id]
		return off, ok
	}, nil, XFDFStyle)

	want := "xref\n0 3\n" +
		"0000000000 65535 f \r\n" +
		"0000000010 00000 n \r\n" +
		"0000000000 65535 f \r\n"
	if out.String() != want {
		t.Fatalf("dense xfdf xref differs:\n%q\n%q", out.String(), want)
	}
}

func TestWriteIncrementalXRefGrouping(t *testing.T) {
	entries := map[int]IncrementalEntry{
		3: {ID: 3, Offset: 1000, Gen: 0},
		4: {ID: 4, Offset: 1100, Gen: 0},
		7: {ID: 7, Offset: 1200, Gen: 1},
	}
	var out bytes.Buffer
	WriteIncrementalXRef(&out, []int{3, 4, 7}, func(id int) IncrementalEntry {
		return entries[id]
	})

	want := "3 2\n" +
		"0000001000 00000 n \n" +
		"0000001100 00000 n \n" +
		"7 1\n" +
		"0000001200 00001 n \n"
	if out.String() != want {
		t.Fatalf("incremental xref differs:\n%q\n%q", out.String(), want)
	}
}
