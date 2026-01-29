// Package main demonstrates using gopdflib to generate a financial report PDF.
// This example creates a complete financial report with tables, styling, bookmarks,
// and measures generation performance.
//
// Run with: go run examples/financial_report/main.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
)

func main() {
	fmt.Println("=== gopdflib Financial Report Example ===")
	fmt.Println()

	// Number of iterations for benchmarking
	iterations := 10

	// Build the template
	template := buildFinancialReportTemplate()

	// Warm-up run (not counted in metrics)
	fmt.Println("Warm-up run...")
	_, err := gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("Error during warm-up: %v\n", err)
		os.Exit(1)
	}

	// Run multiple iterations and collect timing
	fmt.Printf("\nRunning %d iterations...\n\n", iterations)

	var totalDuration time.Duration
	var durations []time.Duration
	var lastPDF []byte

	for i := 1; i <= iterations; i++ {
		start := time.Now()
		pdfBytes, err := gopdflib.GeneratePDF(template)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("Iteration %d: ERROR - %v\n", i, err)
			os.Exit(1)
		}

		durations = append(durations, elapsed)
		totalDuration += elapsed
		lastPDF = pdfBytes

		fmt.Printf("  Iteration %2d: %8.3f ms  (%d bytes)\n", i, float64(elapsed.Microseconds())/1000.0, len(pdfBytes))
	}

	// Calculate statistics
	avgDuration := totalDuration / time.Duration(iterations)
	var minDuration, maxDuration time.Duration = durations[0], durations[0]
	for _, d := range durations {
		if d < minDuration {
			minDuration = d
		}
		if d > maxDuration {
			maxDuration = d
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("=== Performance Summary ===")
	fmt.Printf("  Iterations:    %d\n", iterations)
	fmt.Printf("  Average time:  %.3f ms\n", float64(avgDuration.Microseconds())/1000.0)
	fmt.Printf("  Min time:      %.3f ms\n", float64(minDuration.Microseconds())/1000.0)
	fmt.Printf("  Max time:      %.3f ms\n", float64(maxDuration.Microseconds())/1000.0)
	fmt.Printf("  Total time:    %.3f ms\n", float64(totalDuration.Microseconds())/1000.0)
	fmt.Printf("  PDF size:      %d bytes (%.2f KB)\n", len(lastPDF), float64(len(lastPDF))/1024.0)

	// Save the PDF
	outputPath := "financial_report_output.pdf"
	err = os.WriteFile(outputPath, lastPDF, 0644)
	if err != nil {
		fmt.Printf("\nError saving PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("PDF saved to: %s\n", outputPath)
	fmt.Println("=== Done ===")
}

// buildFinancialReportTemplate creates a PDFTemplate matching financial_digitalsignature.json
func buildFinancialReportTemplate() gopdflib.PDFTemplate {
	// Helper to create float64 pointer
	floatPtr := func(f float64) *float64 { return &f }

	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:          "0:0:0:0",
			Page:                "A4",
			PageAlignment:       1, // Portrait
			Watermark:           "",
			PdfTitle:            "Financial Report Q4 2025",
			PDFACompliant:       true,
			ArlingtonCompatible: true,
			EmbedFonts:          boolPtr(true),
			// Note: Signature config omitted (requires valid certificates)
		},
		Title: gopdflib.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "FINANCIAL REPORT",
			Table: &gopdflib.TitleTable{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{
							Props:     "Helvetica:24:100:center:0:0:0:0",
							Text:      "FINANCIAL REPORT",
							BgColor:   "#154360",
							TextColor: "#FFFFFF",
							Height:    floatPtr(50),
							Link:      "https://example.com/report",
						},
					}},
				},
			},
		},
		Elements: []gopdflib.Element{
			// Section A: Company Information Header
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{
								Props:     "Helvetica:12:100:left:1:1:1:1",
								Text:      "SECTION A: COMPANY INFORMATION",
								BgColor:   "#21618C",
								TextColor: "#FFFFFF",
							},
						}},
					},
				},
			},
			// Company Information Details
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   4,
					ColumnWidths: []float64{1.2, 2, 1.2, 2},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:left:1:0:0:1", Text: "Company Name:", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:000:left:0:0:0:1", Text: "TechCorp Industries Inc.", Link: "https://techcorp.example.com", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:100:left:0:0:0:1", Text: "Report Period:", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:000:left:0:1:0:1", Text: "Q4 2025", BgColor: "#F4F6F7"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:left:1:0:0:0", Text: "Address:", Height: floatPtr(26)},
							{Props: "Helvetica:10:000:left:0:0:0:0", Text: "123 Business Ave, ", Height: floatPtr(24)},
							{Props: "Helvetica:10:100:left:0:0:0:0", Text: "Fiscal Year:"},
							{Props: "Helvetica:10:000:left:0:1:0:0", Text: "2025"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Suite 456, City, State 12345"},
							{Props: "Helvetica:12:000:left:0:0:0:1", Text: ""},
							{Props: "Helvetica:12:000:left:0:0:0:1", Text: ""},
							{Props: "Helvetica:12:000:left:0:1:0:1", Text: ""},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:12:000:left:1:0:0:1", Text: ""},
							{Props: "Helvetica:10:000:left:0:0:0:1", Text: "Go to Financial Summary", TextColor: "#0000FF", Link: "#financial-summary"},
							{Props: "Helvetica:10:000:left:0:0:0:1", Text: "Go to Charts", TextColor: "#0000FF", Link: "#charts-section"},
							{Props: "Helvetica:12:000:left:0:1:0:1", Text: ""},
						}},
					},
				},
			},
			// Section B: Financial Summary Header
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							// Note: "dest" field would be added for bookmark destination
							{Props: "Helvetica:12:100:left:1:1:1:1", Text: "SECTION B: FINANCIAL SUMMARY", BgColor: "#21618C", TextColor: "#FFFFFF"},
						}},
					},
				},
			},
			// Financial Summary Table
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{2, 1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Total Revenue", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$2,450,000", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Cost of Goods Sold", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$1,225,000", BgColor: "#F4F6F7"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:left:1:0:0:1", Text: "Gross Profit", BgColor: "#D4E6F1"},
							{Props: "Helvetica:10:100:right:0:1:0:1", Text: "$1,225,000", BgColor: "#D4E6F1"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Operating Expenses", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$750,000", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Research & Development", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$200,000", BgColor: "#F4F6F7"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Marketing & Sales", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$150,000", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Administrative Expenses", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$100,000", BgColor: "#F4F6F7"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Depreciation & Amortization", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$50,000", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Interest Expense", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$25,000", BgColor: "#F4F6F7"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Taxes", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$75,000", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:11:100:left:1:0:1:1", Text: "Net Income", BgColor: "#A9CCE3"},
							{Props: "Helvetica:11:100:right:0:1:1:1", Text: "$125,000", BgColor: "#A9CCE3"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Earnings Per Share", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$2.50", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:left:1:0:0:1", Text: "Total Assets", BgColor: "#D4E6F1"},
							{Props: "Helvetica:10:100:right:0:1:0:1", Text: "$5,000,000", BgColor: "#D4E6F1"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:000:left:1:0:0:1", Text: "Total Liabilities", BgColor: "#FFFFFF"},
							{Props: "Helvetica:10:000:right:0:1:0:1", Text: "$2,500,000", BgColor: "#FFFFFF"},
						}},
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:left:1:0:0:1", Text: "Shareholders' Equity", BgColor: "#D4E6F1"},
							{Props: "Helvetica:10:100:right:0:1:0:1", Text: "$2,500,000", BgColor: "#D4E6F1"},
						}},
					},
				},
			},
			// Spacer before Charts section
			{
				Type:   "spacer",
				Spacer: &gopdflib.Spacer{Height: 160},
			},
			// Section C: Charts Header
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   1,
					ColumnWidths: []float64{1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:12:100:left:1:1:1:1", Text: "SECTION C: CHARTS", BgColor: "#21618C", TextColor: "#FFFFFF"},
						}},
					},
				},
			},
			// Chart Headers
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1, 1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:10:100:center:1:1:1:0", Text: "REVENUE BREAKDOWN", BgColor: "#F4F6F7"},
							{Props: "Helvetica:10:100:center:1:1:1:0", Text: "EXPENSE DISTRIBUTION", BgColor: "#F4F6F7"},
						}},
					},
				},
			},
			// Chart Placeholders (without actual images for this example)
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1, 1},
					RowHeights:   []float64{200},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:9:000:center:1:1:0:1", Text: "[Bar Chart Placeholder]", Height: floatPtr(200)},
							{Props: "Helvetica:9:000:center:1:1:0:1", Text: "[Pie Chart Placeholder]", Height: floatPtr(200)},
						}},
					},
				},
			},
			// Chart Captions
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1, 1},
					Rows: []gopdflib.Row{
						{Row: []gopdflib.Cell{
							{Props: "Helvetica:8:010:center:1:1:0:1", Text: "Figure 1: Quarterly revenue comparison by region", TextColor: "#566573"},
							{Props: "Helvetica:8:010:center:1:1:0:1", Text: "Figure 2: Breakdown of operating expenses", TextColor: "#566573"},
						}},
					},
				},
			},
			// Final spacer
			{
				Type:   "spacer",
				Spacer: &gopdflib.Spacer{Height: 50},
			},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:8:000:center",
			Text: "TECHCORP INDUSTRIES INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
			Link: "https://example.com/legal",
		},
		Bookmarks: []gopdflib.Bookmark{
			{
				Title: "Financial Report",
				Page:  1,
				Children: []gopdflib.Bookmark{
					{Title: "Company Information", Page: 1},
					{Title: "Financial Summary", Page: 1, Dest: "financial-summary"},
				},
			},
			{
				Title: "Charts",
				Page:  2,
				Dest:  "charts-section",
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
