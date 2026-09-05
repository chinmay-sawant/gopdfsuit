package redact

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// Regression tests for viewer-space (display) coordinate handling.
//
// Blocks arrive from clients in display space: the CropBox region with
// /Rotate applied, top-left origin, in points. The paint operators work in
// MediaBox user space. These tests pin the mapping with hand-computed
// numbers on crafted PDFs (canonical, cropped, rotated, inherited boxes,
// nonzero origin).

// buildSpacePDF assembles a minimal single-page PDF with byte-exact xref so
// the object map always resolves. pagesDict holds the /Pages node dict,
// pageDict the page dict, content the raw content stream.
func buildSpacePDF(t *testing.T, pagesDict, pageDict, content string) []byte {
	t.Helper()
	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n")
	pos := make([]int, 0, 5)
	emit := func(num int, body string) {
		pos = append(pos, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", num, body)
	}
	emit(1, "<< /Type /Catalog /Pages 2 0 R >>")
	emit(2, pagesDict)
	emit(3, pageDict)
	contentBytes := "stream\n" + content + "\nendstream"
	pos = append(pos, out.Len())
	fmt.Fprintf(&out, "4 0 obj\n<< /Length %d >>\n%s\nendobj\n", len(contentBytes), contentBytes)
	xrefPos := out.Len()
	fmt.Fprintf(&out, "xref\n0 5\n0000000000 65535 f \n")
	for _, p := range pos {
		fmt.Fprintf(&out, "%010d 00000 n \n", p)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefPos)
	return out.Bytes()
}

var reOpRe = regexp.MustCompile(`(-?[\d.]+) (-?[\d.]+) (-?[\d.]+) (-?[\d.]+) re`)

func reOps(t *testing.T, pdfBytes []byte) string {
	t.Helper()
	var ops []string
	for _, match := range reOpRe.FindAllStringSubmatch(string(pdfBytes), -1) {
		ops = append(ops, strings.Join(match[1:], " "))
	}
	return strings.Join(ops, " | ")
}

func TestDisplayMappingCanonicalIsIdentity(t *testing.T) {
	pdf := buildSpacePDF(t,
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << >> /Contents 4 0 R >>",
		"BT /F1 12 Tf 150 712 Td (Marker) Tj ET",
	)
	r, err := NewRedactor(pdf)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.ApplyRedactions([]models.RedactionRect{{PageNum: 1, X: 140, Y: 70, Width: 120, Height: 24}})
	if err != nil {
		t.Fatal(err)
	}
	if got := reOps(t, out); !strings.Contains(got, "140.00 70.00 120.00 24.00") {
		t.Fatalf("canonical apply moved the box: %s", got)
	}
	rects, err := r.FindTextOccurrences("Marker")
	if err != nil {
		t.Fatal(err)
	}
	if len(rects) != 1 {
		t.Fatalf("expected 1 rect, got %d", len(rects))
	}
	// Canonical: display == media, text at x=150 with descender fudge.
	if rects[0].X < 149 || rects[0].X > 151 || rects[0].Y < 707 || rects[0].Y > 711 {
		t.Fatalf("search coords wrong: %+v", rects[0])
	}
}

func TestDisplayMappingCropBoxOffset(t *testing.T) {
	pdf := buildSpacePDF(t,
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /CropBox [100 400 612 792] /Resources << >> /Contents 4 0 R >>",
		"BT /F1 12 Tf 150 712 Td (Marker) Tj ET",
	)
	r, err := NewRedactor(pdf)
	if err != nil {
		t.Fatal(err)
	}
	// The text is visible at display top ~71 with height ~12. A frontend
	// box drawn over it (sent form against display height 392) must land
	// back at media (140, 697, 120x24).
	out, err := r.ApplyRedactions([]models.RedactionRect{{PageNum: 1, X: 40, Y: 297, Width: 120, Height: 24}})
	if err != nil {
		t.Fatal(err)
	}
	if got := reOps(t, out); !strings.Contains(got, "140.00 697.00 120.00 24.00") {
		t.Fatalf("cropped apply landed wrong: %s", got)
	}
	// Search must report sent-form coords: x=150-100=50, y=392-71-12=309.
	rects, err := r.FindTextOccurrences("Marker")
	if err != nil {
		t.Fatal(err)
	}
	if len(rects) != 1 {
		t.Fatalf("expected 1 rect, got %d", len(rects))
	}
	if rects[0].X < 49 || rects[0].X > 51 || rects[0].Y < 307 || rects[0].Y > 311 {
		t.Fatalf("search display coords wrong: %+v", rects[0])
	}
	// Round-trip: feed the display rect back into apply, expect media paint.
	out2, err := r.ApplyRedactions(rects)
	if err != nil {
		t.Fatal(err)
	}
	if got := reOps(t, out2); !strings.Contains(got, "150.00 709.00") {
		t.Fatalf("search->apply round-trip landed wrong: %s", got)
	}
}

func TestDisplayMappingRotate90(t *testing.T) {
	pdf := buildSpacePDF(t,
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Rotate 90 /Resources << >> /Contents 4 0 R >>",
		"BT /F1 12 Tf 150 712 Td (Marker) Tj ET",
	)
	r, err := NewRedactor(pdf)
	if err != nil {
		t.Fatal(err)
	}
	// Media rect (150,712,150x20) shows at display top-left (712,150,20x150)
	// under a 90-degree clockwise rotation; sent form against the rotated
	// display height 612 is (712,312,20x150).
	out, err := r.ApplyRedactions([]models.RedactionRect{{PageNum: 1, X: 712, Y: 312, Width: 20, Height: 150}})
	if err != nil {
		t.Fatal(err)
	}
	if got := reOps(t, out); !strings.Contains(got, "150.00 712.00 150.00 20.00") {
		t.Fatalf("rotated apply landed wrong: %s", got)
	}
}

func TestDisplayMappingInheritedMediaBox(t *testing.T) {
	pdf := buildSpacePDF(t,
		"<< /Type /Pages /Kids [3 0 R] /Count 1 /MediaBox [0 0 612 792] >>",
		"<< /Type /Page /Parent 2 0 R /CropBox [0 592 612 792] /Resources << >> /Contents 4 0 R >>",
		"BT /F1 12 Tf 100 750 Td (TopLine) Tj ET",
	)
	r, err := NewRedactor(pdf)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.ApplyRedactions([]models.RedactionRect{{PageNum: 1, X: 95, Y: 146, Width: 200, Height: 24}})
	if err != nil {
		t.Fatal(err)
	}
	if got := reOps(t, out); !strings.Contains(got, "95.00 738.00 200.00 24.00") {
		t.Fatalf("inherited-mediabox apply landed wrong: %s", got)
	}
}

func TestDisplayMappingNonzeroOrigin(t *testing.T) {
	pdf := buildSpacePDF(t,
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [100 100 712 892] /Resources << >> /Contents 4 0 R >>",
		"BT /F1 12 Tf 250 812 Td (Shifted) Tj ET",
	)
	r, err := NewRedactor(pdf)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.ApplyRedactions([]models.RedactionRect{{PageNum: 1, X: 140, Y: 702, Width: 100, Height: 20}})
	if err != nil {
		t.Fatal(err)
	}
	if got := reOps(t, out); !strings.Contains(got, "240.00 802.00 100.00 20.00") {
		t.Fatalf("offset-origin apply landed wrong: %s", got)
	}
}

func TestDisplayMappingRoundTripAllRotations(t *testing.T) {
	// Crop [100 400 612 792] on MediaBox [0 0 612 792]. Corner pairs are
	// display top-left space <-> media space for rect (150, 688, 120x24).
	cases := []struct {
		rotate int
		dx     [2]float64
		dy     [2]float64
		mx     [2]float64
		my     [2]float64
	}{
		{0, [2]float64{50, 170}, [2]float64{80, 104}, [2]float64{150, 270}, [2]float64{712, 688}},
		{90, [2]float64{288, 312}, [2]float64{50, 170}, [2]float64{150, 270}, [2]float64{688, 712}},
		{180, [2]float64{342, 462}, [2]float64{80, 104}, [2]float64{270, 150}, [2]float64{712, 688}},
		{270, [2]float64{80, 104}, [2]float64{342, 462}, [2]float64{270, 150}, [2]float64{712, 688}},
	}
	boxes := pageBoxes{crop: [4]float64{100, 400, 612, 792}}
	closeEnough := func(a, b float64) bool { return a-b < 0.001 && b-a < 0.001 }
	for _, tc := range cases {
		boxes.rotate = tc.rotate
		for k := 0; k < 2; k++ {
			ux, uy := displayToMediaPoint(boxes, tc.dx[k], tc.dy[k])
			if !closeEnough(ux, tc.mx[k]) || !closeEnough(uy, tc.my[k]) {
				t.Fatalf("rotate %d display->media: got (%.2f,%.2f) want (%.2f,%.2f)", tc.rotate, ux, uy, tc.mx[k], tc.my[k])
			}
			dx, dy := mediaToDisplayPoint(boxes, tc.mx[k], tc.my[k])
			if !closeEnough(dx, tc.dx[k]) || !closeEnough(dy, tc.dy[k]) {
				t.Fatalf("rotate %d media->display: got (%.2f,%.2f) want (%.2f,%.2f)", tc.rotate, dx, dy, tc.dx[k], tc.dy[k])
			}
		}
	}
}

func TestNormalizeRotate(t *testing.T) {
	for in, want := range map[int]int{0: 0, 90: 90, 180: 180, 270: 270, 360: 0, -90: 270, 45: 90, 135: 180, 200: 180} {
		if got := normalizeRotate(in); got != want {
			t.Fatalf("normalizeRotate(%d) = %d, want %d", in, got, want)
		}
	}
}
