package form

import (
	"os"
	"testing"
)

func TestFindWidgetDictBoundsHospital(t *testing.T) {
	data, err := os.ReadFile("../../../sampledata/filler/us_hospital_encounter_acroform.pdf")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"patient_name", "address", "mrn", "hipaa_ack"} {
		start, end, ok := findWidgetDictBounds(data, name)
		if !ok {
			t.Errorf("%s: not found", name)
			continue
		}
		chunk := data[start:end]
		if !bytesContains(chunk, []byte("/T ("+name+")")) {
			t.Errorf("%s: bounds %d-%d missing /T", name, start, end)
		}
		if !bytesContains(chunk, []byte("/Subtype /Widget")) && !bytesContains(chunk, []byte("/Subtype/Widget")) {
			t.Errorf("%s: bounds %d-%d missing Widget subtype", name, start, end)
		}
		t.Logf("%s: %d-%d len=%d", name, start, end, end-start)
	}
}

func bytesContains(b, sub []byte) bool {
	return len(b) >= len(sub) && indexOf(b, sub) >= 0
}

func indexOf(b, sub []byte) int {
	for i := 0; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == string(sub) {
			return i
		}
	}
	return -1
}
