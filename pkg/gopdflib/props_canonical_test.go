package gopdflib_test

import (
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// Cross-boundary props-grammar pin: the fluent builder must render exactly
// what the engine's parseProps produces for the same logical font. The
// engine (internal/pdf) cannot be imported here without an import cycle, so
// equality is pinned through the shared literal asserted on both sides
// (see internal/pdf/props_canonical_test.go).
func TestFontBuilderSizeZeroMatchesEngineParse(t *testing.T) {
	const want = "Helvetica:12:000:left:0:0:0:0"
	if got := gopdflib.Font("Helvetica").Size(0).Props(); got != want {
		t.Fatalf("Font(Helvetica).Size(0).Props() = %q, want engine parseProps output %q", got, want)
	}
	if got := gopdflib.Font("").Size(-3).Props(); got != want {
		t.Fatalf("Font(empty).Size(-3).Props() = %q, want %q", got, want)
	}
	if got := gopdflib.MakeProps("Helvetica", 0, false, false, false, "justify", [4]int{0, 0, 0, 0}); got != want {
		t.Fatalf("MakeProps zero/justify = %q, want %q", got, want)
	}
}
