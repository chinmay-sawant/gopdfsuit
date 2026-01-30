// Package main demonstrates the auto text wrapping feature in gopdflib.
// This example shows how cells automatically wrap text and adjust row heights.
//
// Run with: go run sampledata/gopdflib/text_wrapping/main.go
package main

import (
	"fmt"
	"os"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
)

func main() {
	fmt.Println("=== gopdflib Text Wrapping Example ===")
	fmt.Println()

	// Build the template with text wrapping examples
	template := buildTextWrappingTemplate()

	// Generate PDF
	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("Error generating PDF: %v\n", err)
		os.Exit(1)
	}

	// Save the PDF
	outputPath := "text_wrapping_output.pdf"
	err = os.WriteFile(outputPath, pdfBytes, 0644)
	if err != nil {
		fmt.Printf("Error saving PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PDF generated successfully: %s (%d bytes)\n", outputPath, len(pdfBytes))
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("  1. Auto text wrapping (enabled by default)")
	fmt.Println("  2. Dynamic row height based on wrapped content")
	fmt.Println("  3. Disabling wrap for specific cells")
	fmt.Println("  4. Mixed wrapped and non-wrapped cells in same row")
}

func buildTextWrappingTemplate() models.PDFTemplate {
	// Helper to create bool pointer
	boolPtr := func(b bool) *bool { return &b }

	return models.PDFTemplate{
		Config: models.Config{
			PageBorder:    "0:0:0:0",
			Page:          "A4",
			PageAlignment: 1,
			PdfTitle:      "Text Wrapping Demo",
		},
		Title: models.Title{
			Table: &models.TitleTable{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []models.Row{
					{Row: []models.Cell{{
						Props: "Helvetica:18:100:center:0:0:0:0",
						Text:  "Text Wrapping Feature Demo",
					}}},
				},
			},
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "Text Wrapping Feature Demo",
		},
		Elements: []models.Element{
			// Section header
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []models.Row{
						{Row: []models.Cell{{
							Props:     "Helvetica:14:100:left:0:0:1:1",
							Text:      "1. Auto Text Wrapping (Default Behavior)",
							BgColor:   "#E8F4F8",
							TextColor: "#1A5276",
						}}},
					},
				},
			},
			// Example 1: Default wrapping behavior
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   3,
					ColumnWidths: []float64{1, 2, 1},
					Rows: []models.Row{
						// Header row
						{Row: []models.Cell{
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Column A", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Column B (Long Text)", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Column C", BgColor: "#D5DBDB"},
						}},
						// Data row with long text - wrap is enabled by default
						{Row: []models.Cell{
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Short"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "This is a very long piece of text that demonstrates the automatic text wrapping feature. The cell height will automatically adjust to fit all the content, and all cells in the same row will stretch to match the tallest cell."},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Also short"},
						}},
						// Another row
						{Row: []models.Cell{
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Item 2"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Medium length text that wraps to two lines for demonstration purposes."},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Value"},
						}},
					},
				},
			},
			// Spacer
			{Type: "spacer", Spacer: &models.Spacer{Height: 20}},
			// Section header 2
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []models.Row{
						{Row: []models.Cell{{
							Props:     "Helvetica:14:100:left:0:0:1:1",
							Text:      "2. Disabling Wrap for Specific Cells",
							BgColor:   "#E8F4F8",
							TextColor: "#1A5276",
						}}},
					},
				},
			},
			// Example 2: Mixed wrap enabled/disabled
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   3,
					ColumnWidths: []float64{1, 1, 1},
					Rows: []models.Row{
						{Row: []models.Cell{
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Wrap ON (default)", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Wrap OFF", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Wrap ON (default)", BgColor: "#D5DBDB"},
						}},
						{Row: []models.Cell{
							// Default: wrap is on
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "This long text will wrap automatically because wrap is enabled by default. Notice how the row height adjusts."},
							// Explicitly disable wrap
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "This text has wrap disabled so it will be clipped if too long for the cell width.", Wrap: boolPtr(false)},
							// Default: wrap is on
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Another cell with wrapping enabled by default showing multiline content."},
						}},
					},
				},
			},
			// Spacer
			{Type: "spacer", Spacer: &models.Spacer{Height: 20}},
			// Section header 3
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []models.Row{
						{Row: []models.Cell{{
							Props:     "Helvetica:14:100:left:0:0:1:1",
							Text:      "3. Various Column Widths with Wrapping",
							BgColor:   "#E8F4F8",
							TextColor: "#1A5276",
						}}},
					},
				},
			},
			// Example 3: Different column widths
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   4,
					ColumnWidths: []float64{0.5, 1.5, 1, 1},
					Rows: []models.Row{
						{Row: []models.Cell{
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "ID", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Description", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Status", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Notes", BgColor: "#D5DBDB"},
						}},
						{Row: []models.Cell{
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "001"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Implement automatic text wrapping feature for PDF table cells with dynamic row height calculation based on content length."},
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "Complete", BgColor: "#D5F5E3"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Released in v2.0"},
						}},
						{Row: []models.Cell{
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "002"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Add frontend preview support for wrapped text in the visual editor."},
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "Complete", BgColor: "#D5F5E3"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Editor shows live preview"},
						}},
						{Row: []models.Cell{
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "003"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Performance optimization for large tables with many wrapped cells."},
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "In Progress", BgColor: "#FCF3CF"},
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "Benchmarking required"},
						}},
					},
				},
			},
			// Spacer
			{Type: "spacer", Spacer: &models.Spacer{Height: 20}},
			// Section header 4
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []models.Row{
						{Row: []models.Cell{{
							Props:     "Helvetica:14:100:left:0:0:1:1",
							Text:      "4. Text Alignment with Wrapping",
							BgColor:   "#E8F4F8",
							TextColor: "#1A5276",
						}}},
					},
				},
			},
			// Example 4: Different alignments
			{
				Type: "table",
				Table: &models.Table{
					MaxColumns:   3,
					ColumnWidths: []float64{1, 1, 1},
					Rows: []models.Row{
						{Row: []models.Cell{
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Left Aligned", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Center Aligned", BgColor: "#D5DBDB"},
							{Props: "Helvetica:10:100:center:1:1:1:1", Text: "Right Aligned", BgColor: "#D5DBDB"},
						}},
						{Row: []models.Cell{
							{Props: "Helvetica:10:000:left:1:1:0:1", Text: "This is left-aligned wrapped text. Each line starts from the left edge of the cell."},
							{Props: "Helvetica:10:000:center:1:1:0:1", Text: "This is center-aligned wrapped text. Each line is centered within the cell."},
							{Props: "Helvetica:10:000:right:1:1:0:1", Text: "This is right-aligned wrapped text. Each line ends at the right edge."},
						}},
					},
				},
			},
		},
		Footer: models.Footer{
			Font: "Helvetica:8:000:left:0:0:0:0",
			Text: "Generated with gopdflib - Text Wrapping Demo",
		},
	}
}
