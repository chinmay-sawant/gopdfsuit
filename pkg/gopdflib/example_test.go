package gopdflib_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
)

// TestGeneratePDFFromJSON demonstrates loading a JSON template and generating a PDF.
func TestGeneratePDFFromJSON(t *testing.T) {
	// Load a sample JSON template
	data, err := os.ReadFile("../../sampledata/financial_digitalsignature.json")
	if err != nil {
		t.Skipf("Skipping test: sample file not found: %v", err)
	}

	var template gopdflib.PDFTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		t.Fatalf("Failed to parse template JSON: %v", err)
	}

	// Remove signature config for this test (requires valid certificates)
	template.Config.Signature = nil

	// Generate PDF
	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	// Verify PDF header
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Error("Generated file does not start with PDF header")
	}

	// Verify PDF footer
	if !bytes.HasSuffix(pdfBytes, []byte("%%EOF\n")) {
		t.Error("Generated file does not end with EOF marker")
	}

	t.Logf("Generated PDF: %d bytes", len(pdfBytes))
}

// TestGeneratePDFProgrammatically demonstrates creating a PDF from code.
func TestGeneratePDFProgrammatically(t *testing.T) {
	// Create a simple template programmatically
	template := gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			Page:          "A4",
			PageAlignment: 1, // Portrait
			PageBorder:    "20:20:20:20",
		},
		Title: gopdflib.Title{
			Props: "Helvetica:24:100:center:0:0:0:0",
			Text:  "Sample Document",
		},
		Elements: []gopdflib.Element{
			{
				Type: "spacer",
				Spacer: &gopdflib.Spacer{
					Height: 20,
				},
			},
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   3,
					ColumnWidths: []float64{2, 1, 1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:12:100:left:1:1:1:1", Text: "Item", BgColor: "#E8E8E8"},
							{Props: "Helvetica:12:100:center:1:1:1:1", Text: "Quantity", BgColor: "#E8E8E8"},
							{Props: "Helvetica:12:100:right:1:1:1:1", Text: "Price", BgColor: "#E8E8E8"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Widget A"},
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "10"},
							{Props: "Helvetica:10:000:right:1:1:0:1", Text: "$99.00"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Widget B"},
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "5"},
							{Props: "Helvetica:10:000:right:1:1:0:1", Text: "$149.00"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:left:1:1:1:1", Text: "Total", BgColor: "#D0D0D0"},
							{Props: "Helvetica:10:000:center:1:1:1:1", Text: "", BgColor: "#D0D0D0"},
							{Props: "Helvetica:10:100:right:1:1:1:1", Text: "$1,735.00", BgColor: "#D0D0D0"},
						}},
					},
				},
			},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:8:000:center",
			Text: "Generated with gopdflib",
		},
	}

	// Generate PDF
	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	// Verify PDF header
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Error("Generated file does not start with PDF header")
	}

	t.Logf("Generated programmatic PDF: %d bytes", len(pdfBytes))
}

// ExampleGeneratePDF demonstrates basic PDF generation.
func ExampleGeneratePDF() {
	template := gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			Page:          "A4",
			PageAlignment: 1,
		},
		Title: gopdflib.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "Hello, PDF!",
		},
	}

	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		panic(err)
	}

	// Use pdfBytes...
	_ = pdfBytes
}

// ExampleMergePDFs demonstrates merging multiple PDFs.
func ExampleMergePDFs() {
	// In a real scenario, you would read actual PDF files
	pdf1 := []byte("%PDF-1.4...")
	pdf2 := []byte("%PDF-1.4...")

	merged, err := gopdflib.MergePDFs([][]byte{pdf1, pdf2})
	if err != nil {
		panic(err)
	}

	// Use merged PDF...
	_ = merged
}
