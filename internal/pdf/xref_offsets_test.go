package pdf

import (
	"testing"
)

func TestXrefOffsetsLazyGrowAndPool(t *testing.T) {
	h := newXrefOffsetsHandle(8)
	defer h.release()
	offsets := h.ptr()

	setXrefOffset(offsets, 1, 100)
	setXrefOffset(offsets, 4, 400)

	if len(*offsets) != 5 {
		t.Fatalf("len=%d want 5", len(*offsets))
	}
	if (*offsets)[2] != xrefOffsetUnused {
		t.Fatalf("unused slot not initialized: got %d", (*offsets)[2])
	}

	used := collectUsedXrefObjectIDs(*offsets)
	if len(used) != 3 || used[0] != 0 || used[1] != 1 || used[2] != 4 {
		t.Fatalf("unexpected used set: %v", used)
	}
}

func TestXrefOffsetsHandleReuse(t *testing.T) {
	h1 := newXrefOffsetsHandle(16)
	setXrefOffset(h1.ptr(), 2, 42)
	cap1 := cap(*h1.ptr())
	h1.release()

	h2 := newXrefOffsetsHandle(16)
	if cap(*h2.ptr()) < cap1 {
		t.Fatalf("pooled cap=%d smaller than released cap=%d", cap(*h2.ptr()), cap1)
	}
	if len(*h2.ptr()) != 0 {
		t.Fatalf("expected empty reused slice, len=%d", len(*h2.ptr()))
	}
	h2.release()
}