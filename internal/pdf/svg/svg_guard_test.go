package svg

import (
	"strings"
	"testing"
)

// TestConvertSVGRejectsEntities proves billion-laughs-style payloads are
// rejected fast instead of being expanded by the XML parser.
func TestConvertSVGRejectsEntities(t *testing.T) {
	entity := `<?xml version="1.0"?>
<!DOCTYPE svg [<!ENTITY lol "lollollollollollollollol"><!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;">]>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect width="10" height="10"/></svg>`
	if _, _, _, err := ConvertSVGToPDFCommands([]byte(entity)); err == nil {
		t.Fatalf("expected rejection of ENTITY payload")
	}
	doctype := `<!DOCTYPE svg><svg width="10" height="10"></svg>`
	if _, _, _, err := ConvertSVGToPDFCommands([]byte(doctype)); err == nil {
		t.Fatalf("expected rejection of DOCTYPE payload")
	}
}

// TestConvertSVGRejectsOversized proves inputs over the cap are rejected.
func TestConvertSVGRejectsOversized(t *testing.T) {
	big := []byte("<svg>" + strings.Repeat(" ", MaxSVGBytes) + "</svg>")
	if _, _, _, err := ConvertSVGToPDFCommands(big); err == nil {
		t.Fatalf("expected rejection of oversized SVG")
	}
}

// TestConvertSVGValidStillWorks guards the cap against false positives.
func TestConvertSVGValidStillWorks(t *testing.T) {
	small := `<svg width="100" height="50"><rect x="1" y="2" width="10" height="5"/></svg>`
	cmds, w, h, err := ConvertSVGToPDFCommands([]byte(small))
	if err != nil {
		t.Fatalf("valid SVG rejected: %v", err)
	}
	if w != 100 || h != 50 || len(cmds) == 0 {
		t.Fatalf("unexpected output w=%d h=%d len=%d", w, h, len(cmds))
	}
}
