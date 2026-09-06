// Package main demonstrates gopdfsuit's content caches: TTL policy,
// manual clearing, and the warm vs cold generation timing gap.
//
// It builds one template, generates it cold (caches cleared, full rebuild
// cost) and warm (caches hot), prints both timings, and saves the result
// as a 2-page PDF.
//
// Run from inside sampledata (its own Go module):
//
//	cd sampledata && go run ./caching
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

// ledgerRows is tuned so the table flows onto exactly 2 A4 pages.
// Verified by counting /Type /Page objects in the output below.
const ledgerRows = 36

func floatPtr(f float64) *float64 { return &f }

func buildCachingTemplate() gopdflib.PDFTemplate {
	rows := make([]gopdflib.Row, 0, ledgerRows+1)
	rows = append(rows, gopdflib.Row{Row: []gopdflib.Cell{
		{Props: "Helvetica:10:100:center:1:0:1:1", Text: "Line", BgColor: "#D4E6F1"},
		{Props: "Helvetica:10:100:center:0:0:1:1", Text: "Cache key", BgColor: "#D4E6F1"},
		{Props: "Helvetica:10:100:center:0:0:1:1", Text: "Status", BgColor: "#D4E6F1"},
		{Props: "Helvetica:10:100:right:0:1:1:1", Text: "Size (bytes)", BgColor: "#D4E6F1"},
	}})
	for i := 1; i <= ledgerRows; i++ {
		bg := ""
		if i%2 == 1 {
			bg = "#F8F9F9"
		}
		rows = append(rows, gopdflib.Row{Row: []gopdflib.Cell{
			{Props: "Helvetica:9:000:center:1:0:0:1", Text: strconv.Itoa(i), BgColor: bg},
			{Props: "Helvetica:9:000:left:0:0:0:1", Text: "subset+glyphs dia:" + strconv.Itoa(1000+i), BgColor: bg},
			{Props: "Helvetica:9:000:center:0:0:0:1", Text: "cached", TextColor: "#27AE60", BgColor: bg},
			{Props: "Helvetica:9:000:right:0:1:0:1", Text: strconv.Itoa(512 + i*8), BgColor: bg},
		}})
	}

	template := gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:    "0:0:0:0",
			Page:          "A4",
			PageAlignment: 1,
			PdfTitle:      "Caching Example - 2 pages",
		},
		Title: gopdflib.Title{
			Props: "Helvetica:24:100:center:0:0:0:0",
			Text:  "CONTENT CACHE WALKTHROUGH",
		},
		Elements: []gopdflib.Element{
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "WHAT IS CACHED", BgColor: "#21618C", TextColor: "#FFFFFF"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   2,
				ColumnWidths: []float64{1.4, 2.6},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Font subsets", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "Subset TTF bytes per font plus glyph set", BgColor: "#EBF5FB"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Page streams"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "zlib output for identical page content"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Prop parses", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "Parsed Helvetica:9:000:... strings", BgColor: "#EBF5FB"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Default TTL"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "3 minutes, GOPDFSUIT_CACHE_TTL overrides"},
					}},
				},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:           4,
				ColumnWidths:         []float64{0.7, 2.3, 1.2, 1.3},
				Rows:                 rows,
				SharedRowLayout:      true,
				SharedRowTemplateRow: 1,
			}},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:7:000:center",
			Text: "CACHING EXAMPLE | PAGE",
		},
	}
	template.SetPrecomputedStandardFonts("Helvetica")
	return template
}

// countPages counts /Type /Page objects without matching /Type /Pages.
func countPages(pdf []byte) int {
	const needle = "/Type /Page"
	n := 0
	for i := 0; i+len(needle) <= len(pdf); {
		j := bytes.Index(pdf[i:], []byte(needle))
		if j < 0 {
			break
		}
		i += j + len(needle)
		// Skip the /Pages parent object: next byte is 's' there.
		if i < len(pdf) && pdf[i] == 's' {
			continue
		}
		n++
	}
	return n
}

func main() {
	fmt.Println("=== gopdfsuit Caching Example ===")
	fmt.Println()

	// 1. Show the default TTL, then configure a shorter one for the demo.
	fmt.Printf("Default cache TTL: %v\n", gopdflib.CacheTTL())
	gopdflib.SetCacheTTL(2 * time.Minute)
	fmt.Printf("Configured cache TTL: %v (restored to default on exit)\n", gopdflib.CacheTTL())
	defer gopdflib.SetCacheTTL(gopdflib.DefaultCacheTTL)
	fmt.Println()

	template := buildCachingTemplate()

	// 2. Cold generation: clear every content cache first, like BOPS does.
	gopdflib.ClearBOPSCaches()
	start := time.Now()
	coldPDF, err := gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("Cold generation failed: %v\n", err)
		os.Exit(1)
	}
	coldElapsed := time.Since(start)
	fmt.Printf("Cold:  %8.3f ms  (%d bytes, caches cleared first)\n",
		float64(coldElapsed.Microseconds())/1000.0, len(coldPDF))

	// 3. Warm generation: same template, caches hot from the cold run.
	start = time.Now()
	warmPDF, err := gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("Warm generation failed: %v\n", err)
		os.Exit(1)
	}
	warmElapsed := time.Since(start)
	fmt.Printf("Warm:  %8.3f ms  (%d bytes, subset plus compress plus props reuse)\n",
		float64(warmElapsed.Microseconds())/1000.0, len(warmPDF))
	fmt.Println()

	// 4. Save the 2-page file next to this example.
	outDir := "sampledata/caching"
	if _, err := os.Stat("sampledata/go.mod"); err != nil {
		// Running from inside sampledata/: output lands in ./caching/.
		outDir = "caching"
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Cannot create output dir: %v\n", err)
		os.Exit(1)
	}
	outPath := filepath.Join(outDir, "caching_example.pdf")
	if err := os.WriteFile(outPath, warmPDF, 0644); err != nil {
		fmt.Printf("Cannot write PDF: %v\n", err)
		os.Exit(1)
	}

	pages := countPages(warmPDF)
	fmt.Printf("Saved: %s (%d bytes, %d pages)\n", outPath, len(warmPDF), pages)
	if pages != 2 {
		fmt.Printf("NOTE: expected exactly 2 pages, adjust ledgerRows (now %d).\n", ledgerRows)
		os.Exit(1)
	}
}
