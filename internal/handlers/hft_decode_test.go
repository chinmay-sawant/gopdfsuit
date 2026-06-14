package handlers

import (
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

func TestDecodeHFTPayload_FillRowsInPlace(t *testing.T) {
	const payload = `{
		"config": {"page": "A4", "pageAlignment": 1, "taggedPDF": true},
		"title": {"props": "Helvetica:24:100:center:0:0:0:0", "text": "HFT"},
		"elements": [
			{
				"type": "table",
				"table": {
					"maxcolumns": 2,
					"rows": [{"row": [{"props": "a", "text": "x"}, {"props": "b", "text": "y"}]}]
				}
			},
			{
				"type": "table",
				"table": {
					"maxcolumns": 2,
					"sharedRowLayout": true,
					"sharedRowTemplateRow": 1,
					"rows": [
						{"row": [{"props": "h1", "text": "H"}]},
						{"row": [{"props": "d1", "text": "1", "bgcolor": "#fff"}, {"props": "d2", "text": "2", "textcolor": "#000"}]}
					]
				}
			}
		],
		"footer": {"font": "Helvetica:7:000:center", "text": "footer"}
	}`

	var tmpl models.PDFTemplate
	tmpl.PreallocForDecode(0, "hft")
	if err := decodeHFTPayload([]byte(payload), &tmpl); err != nil {
		t.Fatalf("decodeHFTPayload: %v", err)
	}
	if len(tmpl.Elements) != 2 {
		t.Fatalf("elements: got %d want 2", len(tmpl.Elements))
	}
	big := tmpl.Elements[1].Table
	if big == nil || len(big.Rows) != 2 {
		t.Fatalf("big table rows: got %v", big)
	}
	if big.Rows[1].Row[0].Text != "1" || big.Rows[1].Row[0].BgColor != "#fff" {
		t.Fatalf("row1 col0: %+v", big.Rows[1].Row[0])
	}
	if big.Rows[1].Row[1].TextColor != "#000" {
		t.Fatalf("row1 col1 textcolor: %q", big.Rows[1].Row[1].TextColor)
	}
	if cap(big.Rows) < 2001 {
		t.Fatalf("expected prealloc cap >= 2001, got %d", cap(big.Rows))
	}
}
