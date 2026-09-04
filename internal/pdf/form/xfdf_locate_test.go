package form

import (
	"bytes"
	"compress/zlib"
	"testing"
)

// locateFixture builds a synthetic field object map without any full PDF:
// catalog 1 0 -> AcroForm 2 0 -> Fields [3 0 R 4 0 R], where 3 is a text
// field with a direct value and 4 is a parent with a kid carrying the value.
func locateFixture() (map[string][]byte, []byte) {
	objMap := map[string][]byte{
		"1 0": []byte(`<< /Type /Catalog /Pages 9 0 R /AcroForm 2 0 R >>`),
		"2 0": []byte(`<< /Fields [3 0 R 4 0 R] /DA (/Helv 0 Tf 0 g) >>`),
		"3 0": []byte(`<< /FT /Tx /T (FullName) /V (Alice) /Rect [0 0 100 20] >>`),
		"4 0": []byte(`<< /T (Parent) /Kids [5 0 R] >>`),
		"5 0": []byte(`<< /FT /Tx /T (Child) /V (nested-value) /Rect [0 0 100 20] >>`),
	}
	pdfBytes := []byte(`/Root 1 0 R`)
	return objMap, pdfBytes
}

// TestLocateFieldsSyntheticMap proves the locate seam resolves direct and
// kid-inherited field values from a bare object map, no PDF parsing needed.
func TestLocateFieldsSyntheticMap(t *testing.T) {
	objMap, pdfBytes := locateFixture()

	got := LocateFields(objMap, pdfBytes)

	if got["FullName"] != "Alice" {
		t.Fatalf("FullName = %q, want %q", got["FullName"], "Alice")
	}
	if got["Parent.Child"] != "nested-value" {
		t.Fatalf("Parent.Child = %q, want %q", got["Parent.Child"], "nested-value")
	}
}

// TestLocateFieldsMissingAcroForm proves an object map without a catalog or
// AcroForm yields no fields instead of hanging or panicking.
func TestLocateFieldsMissingAcroForm(t *testing.T) {
	objMap := map[string][]byte{
		"1 0": []byte(`<< /Type /Catalog /Pages 9 0 R >>`),
	}
	if got := LocateFields(objMap, []byte(`/Root 1 0 R`)); len(got) != 0 {
		t.Fatalf("expected no fields, got %v", got)
	}
	if got := LocateFields(map[string][]byte{}, []byte(`no root here`)); len(got) != 0 {
		t.Fatalf("expected no fields for empty map, got %v", got)
	}
}

// TestBuildFieldObjectMapExpandsObjStm proves the map builder expands an
// object stream carried inside the raw bytes.
func TestBuildFieldObjectMapExpandsObjStm(t *testing.T) {
	// Minimal ObjStm: N=1, First=len("10 0 ")=5, one member with a field body.
	// Real object streams are Flate-compressed, so compress the payload.
	inner := `<< /FT /Tx /T (FromStm) /V (stm-value) >>`
	var comp bytes.Buffer
	zw := zlib.NewWriter(&comp)
	if _, err := zw.Write([]byte("10 0 " + inner)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	stmObj := "6 0 obj\n<< /Type /ObjStm /N 1 /First 5 /Length 0 >>\nstream\n" + comp.String() + "\nendstream\nendobj\n"
	pdf := []byte("%PDF-1.4\n" + stmObj + "trailer\n<< /Size 11 /Root 1 0 R >>\nstartxref\n0\n%%EOF\n")

	objMap := buildFieldObjectMap(pdf)
	body, ok := objMap["10 0"]
	if !ok {
		t.Fatalf("object-stream member 10 0 missing from map: %v", objMap)
	}
	if string(body) != inner {
		t.Fatalf("member body = %q, want %q", body, inner)
	}
}
