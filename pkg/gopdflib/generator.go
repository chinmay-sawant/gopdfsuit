// Package gopdflib provides PDF generation from templates.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/internal/pdf"
)

// GeneratePDF creates a PDF document from a template and returns the PDF bytes.
//
// Example usage:
//
//	template := gopdflib.PDFTemplate{
//	    Config: gopdflib.Config{
//	        Page:          "A4",
//	        PageAlignment: 1, // Portrait
//	    },
//	    Title: gopdflib.Title{
//	        Props: "Helvetica:18:100:center:0:0:0:0",
//	        Text:  "My Document",
//	    },
//	    Elements: []gopdflib.Element{
//	        {
//	            Type: "table",
//	            Table: &gopdflib.Table{
//	                MaxColumns:   2,
//	                ColumnWidths: []float64{1, 1},
//	                Rows: []gopdflib.Row{
//	                    {Row: []gopdflib.Cell{
//	                        {Props: "Helvetica:12:100:left:1:1:1:1", Text: "Column 1"},
//	                        {Props: "Helvetica:12:100:left:1:1:1:1", Text: "Column 2"},
//	                    }},
//	                },
//	            },
//	        },
//	    },
//	}
//
//	pdfBytes, err := gopdflib.GeneratePDF(template)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.pdf", pdfBytes, 0644)
func GeneratePDF(template PDFTemplate) ([]byte, error) {
	return pdf.GenerateTemplatePDF(template)
}

// GetAvailableFonts returns a list of available fonts for PDF generation.
// This includes standard PDF fonts and any custom fonts that have been registered.
func GetAvailableFonts() []FontInfo {
	return pdf.GetAvailableFonts()
}

// GetFontRegistry returns the global font registry for registering custom fonts.
// Use this to register custom TTF/OTF fonts before generating PDFs.
//
// Example:
//
//	registry := gopdflib.GetFontRegistry()
//	err := registry.RegisterFontFromFile("MyFont", "/path/to/font.ttf")
func GetFontRegistry() *pdf.CustomFontRegistry {
	return pdf.GetFontRegistry()
}
