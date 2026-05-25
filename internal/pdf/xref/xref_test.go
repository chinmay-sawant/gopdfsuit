package xref

import (
	"bytes"
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
