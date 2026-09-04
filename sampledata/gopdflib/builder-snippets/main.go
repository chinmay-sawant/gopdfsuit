// Package main demonstrates the builder-snippets overlay.
//
// It mirrors sampledata/builder-snippets/snippet.json: a 3-column title table with a
// bracketed right cell in a different color, a spacer, a placeholder image,
// and a 2-column data table with a red right cell.
//
// NOTE: cells below use the fluent Font chains (the preferred spelling);
// raw Props strings stay supported as the low-level form.
//
// Run from the sampledata module:
//
//	cd sampledata && go run ./gopdflib/builder-snippets
package main

import (
	"fmt"
	"os"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// placeholderPNG is a 1x1 PNG used so the sample needs no image files.
const placeholderPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func main() {
	b := gopdflib.NewDocument("A4", true)
	b.AddTitle("Document Title", gopdflib.WithTitleFont("Helvetica", 18, true))

	titleTable := b.AddTitleTable(3, 1, 2, 1)
	row := titleTable.AddRow(
		gopdflib.Font("Helvetica").Size(12).Bordered().Cell(""),
		gopdflib.Font("Helvetica").Size(18).Bold().Center().Bordered().Cell("Document Title"),
		gopdflib.Font("Helvetica").Size(12).Right().Bordered().Cell("clause"),
	)
	gopdflib.SetCellFont(&row[1], "Helvetica", 18, true, false, false)
	gopdflib.AddBracketText(&row[2], "[", "]")
	gopdflib.SetBracketFont(&row[2], "Helvetica", 12)
	gopdflib.SetCellTextColor(&row[2], "#B00020")

	b.AddSpacer(20)
	b.AddImage("placeholder", placeholderPNG, 100, 80)

	tb := b.AddTable(2, 2, 1)
	tb.AddRow(gopdflib.HeaderCell("Item"), gopdflib.HeaderCell("Price"))
	amounts := tb.AddRow(
		gopdflib.Font("Helvetica").Size(10).Bordered().Cell("Total Revenue"),
		gopdflib.Font("Helvetica").Size(10).Right().Bordered().Cell("$2,450,000"),
	)
	gopdflib.SetCellTextColor(&amounts[1], "#B00020")

	template := b.Build()
	template.Config.PageBorder = "1:1:1:1"
	template.Config.PageMargin = "72:72:72:72"
	template.Config.PdfTitle = "Copy-clip snippet"
	template.Footer = gopdflib.Footer{
		Font: "Helvetica:8:000:center:0:0:0:0",
		Text: "Copy-clip snippet footer",
	}

	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("generate: %v\n", err)
		os.Exit(1)
	}
	// Works from the repo root and from the sampledata module directory.
	outputs := []string{
		"sampledata/gopdflib/builder-snippets/builder_snippets.pdf",
		"gopdflib/builder-snippets/builder_snippets.pdf",
	}
	for _, out := range outputs {
		if err := os.WriteFile(out, pdfBytes, 0644); err == nil {
			fmt.Printf("wrote %d bytes to %s\n", len(pdfBytes), out)
			return
		}
	}
	fmt.Printf("write: could not save output from %s\n", mustCWD())
	os.Exit(1)
}

func mustCWD() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
