package pdf

import (
	"bytes"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/pdfobj"
)

func TestSharedLayoutRendersMutatedRowContent(t *testing.T) {
	const props = "Helvetica:12:000:left"
	template := models.PDFTemplate{
		Elements: []models.Element{{
			Type: "table",
			Table: &models.Table{
				MaxColumns:           1,
				SharedRowLayout:      true,
				SharedRowTemplateRow: 1,
				Rows: []models.Row{
					{Row: []models.Cell{{Props: props, Text: "first request"}}},
					{Row: []models.Cell{{Props: props, Text: "template row"}}},
				},
			},
		}},
	}

	if _, err := GenerateTemplatePDF(template); err != nil {
		t.Fatalf("first generation: %v", err)
	}
	template.Elements[0].Table.Rows[0].Row[0].Text = "second request"
	pdfBytes, err := GenerateTemplatePDF(template)
	if err != nil {
		t.Fatalf("second generation: %v", err)
	}

	content := inflatePDFStreams(t, pdfBytes)
	if !bytes.Contains(content, []byte("second request")) {
		t.Fatalf("second generation content does not contain mutated row text: %q", content)
	}
	if bytes.Contains(content, []byte("first request")) {
		t.Fatalf("second generation content reused stale row text: %q", content)
	}
}

func inflatePDFStreams(t *testing.T, pdfBytes []byte) []byte {
	t.Helper()
	var content []byte
	streams := bytes.Split(pdfBytes, []byte("\nstream\n"))[1:]
	for _, stream := range streams {
		end := bytes.Index(stream, []byte("\nendstream"))
		if end < 0 {
			continue
		}
		decoded, err := pdfobj.DecompressAny(stream[:end])
		if err != nil {
			content = append(content, stream[:end]...)
			continue
		}
		content = append(content, decoded...)
	}
	return content
}
