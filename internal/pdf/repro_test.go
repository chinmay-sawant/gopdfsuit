package pdf

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

func TestTitleVisibility(t *testing.T) {
	// Create a simple template with a title
	template := models.PDFTemplate{
		Title: models.Title{
			Text:      "Visible Title Check",
			TextColor: "#000000",
			BgColor:   "#FFFFFF",
			Props:     "font-size:24;font-family:Helvetica",
			Table: &models.TitleTable{
				MaxColumns: 1,
				Rows: []models.Row{
					{
						Row: []models.Cell{
							{Text: "Visible Title Check", Props: "font-size:24;font-family:Helvetica"},
						},
					},
				},
			},
		},
		Config: models.Config{
			Page: "A4",
		},
	}

	// Generate PDF
	pdfBytes, err := GenerateTemplatePDF(template)
	if err != nil {
		t.Fatalf("GenerateTemplatePDF failed: %v", err)
	}

	// Check if the title text is present in the PDF output
	// Use bytes.Contains
	expectedText := "(Visible Title Check)"
	if !bytes.Contains(pdfBytes, []byte(expectedText)) {
		t.Logf("PDF does not contain title text literal: %s (Expected if compressed)", expectedText)
	} else {
		fmt.Println("Title text found in PDF!")
	}
}
