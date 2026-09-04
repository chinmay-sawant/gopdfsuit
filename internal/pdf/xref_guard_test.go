package pdf

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

func TestValidXRefWidths(t *testing.T) {
	if total, ok := merge.ValidXRefWidths(1, 2, 2); !ok || total != 5 {
		t.Fatalf("ValidXRefWidths(1,2,2) = %d,%v, want 5,true", total, ok)
	}
	for _, w := range [][3]int{{0, 0, 0}, {-1, 2, 2}, {1, -1, 2}, {20, 20, 20}} {
		if _, ok := merge.ValidXRefWidths(w[0], w[1], w[2]); ok {
			t.Fatalf("ValidXRefWidths(%v) = true, want false", w)
		}
	}
}

// TestParseXRefStreamsRejectsBadW is a hang/panic regression test: W=[0 0 0]
// used to make the entry loop never advance, negative widths panicked.
func TestParseXRefStreamsRejectsBadW(t *testing.T) {
	cases := []string{
		"/W[0 0 0]",
		"/W[1 -2 2]",
		"/W[99 99 99]",
	}
	for _, w := range cases {
		data := []byte("1 0 obj\n<< " + w + " /Index[0 1] /Length 4 >>\nstream\nabcd\nendstream\nendobj\n")
		done := make(chan struct{})
		go func() {
			defer close(done)
			parseXRefStreams(data, map[string][]byte{})
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("parseXRefStreams hung on %s", w)
		}
	}
}
