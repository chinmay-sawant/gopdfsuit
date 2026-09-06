package pdf

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/merge"
)

// Edge inputs to the page-spec parser via its public entry must never panic.
// Empty spec yields no pages and no error; everything else malformed or out
// of range must yield an error, never a panic or silent garbage.
func TestParsePageSpecEdgesNoPanic(t *testing.T) {
	seeds := []string{
		"",
		"0",
		"999-1",
		"abc",
		"1-3-5",
		",,,",
		"  ",
		"1,",
		",1",
		"-1",
		"1-",
		"0-0",
		"999999999999999999999",
		"1-999999999999999999999",
		"1,1,1,2,2",
		"3-1",
		"5-5",
		"1;2",
		"1..3",
		"<script>",
		strings.Repeat("1,", 5000) + "2",
		"2147483647",
		"2147483648",
		"00001",
		" 1 - 3 ",
	}
	for _, s := range seeds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ParsePageSpec(%q) panicked: %v", s, r)
				}
			}()
			pages, err := merge.ParsePageSpec(s, 10)
			switch {
			case s == "" || s == ",,,":
				if err != nil || len(pages) != 0 {
					t.Fatalf("empty/commas-only spec: got %v, %v; want empty, nil", pages, err)
				}
			case s == "1,1,1,2,2" || s == "5-5" || s == "00001" || s == "1," || s == ",1":
				if err != nil {
					t.Fatalf("ParsePageSpec(%q) unexpected error: %v", s, err)
				}
			case s == "1-999999999999999999999":
				// Huge range end clamps to totalPages (or errors); either is
				// safe as long as there is no panic and no silent garbage.
				if err == nil && (len(pages) == 0 || pages[len(pages)-1] > 10) {
					t.Fatalf("ParsePageSpec(%q) unclamped result: %v", s, pages)
				}
			case s == strings.Repeat("1,", 5000)+"2":
				if err != nil {
					t.Fatalf("large duplicate spec unexpected error: %v", s[:20]+"...")
				}
				if len(pages) != 2 || pages[0] != 1 || pages[1] != 2 {
					t.Fatalf("large duplicate spec: got %v", pages)
				}
			default:
				if err == nil {
					t.Fatalf("ParsePageSpec(%q) expected error, got %v", s, pages)
				}
			}
		}()
	}
}
