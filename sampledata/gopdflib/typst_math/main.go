// Package main demonstrates using gopdflib to generate a PDF with Typst math syntax.
// Cells with MathEnabled=true render their text as mathematical expressions using
// Typst math syntax (e.g., $ frac(a, b) $, $ x^2 + y^2 $, $ sqrt(x) $).
//
// Run with: go run sampledata/gopdflib/typst_math/main.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
	fmt.Println("=== gopdflib Typst Math Example ===")
	fmt.Println()

	template := buildTypstMathTemplate()

	fmt.Println("Generating Typst math showcase PDF...")
	start := time.Now()

	pdfBytes, err := gopdflib.GeneratePDF(template)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error generating PDF: %v\n", err)
		os.Exit(1)
	}

	outputPath := "typst_math_showcase.pdf"
	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
		fmt.Printf("Error saving PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PDF saved to: %s (%d bytes)\n", outputPath, len(pdfBytes))
	fmt.Printf("Generated in %.3f ms\n", float64(elapsed.Microseconds())/1000.0)
}

func buildTypstMathTemplate() gopdflib.PDFTemplate {
	mathEnabled := true

	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:    "0:0:0:0",
			Page:          "A4",
			PageAlignment: 1,
		},
		Title: gopdflib.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "TYPST MATH SYNTAX - GOPDFLIB",
			Table: &gopdflib.TitleTable{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{
							Props:     "Helvetica:20:100:center:0:0:0:0",
							Text:      "TYPST MATH SYNTAX - GOPDFLIB",
							BgColor:   "#1B2631",
							TextColor: "#FFFFFF",
							Height:    floatPtr(50),
						},
					}},
				},
			},
		},
		Elements: []gopdflib.Element{
			// Section 1: Basic Arithmetic
			sectionHeader("SECTION 1: BASIC ARITHMETIC & SYMBOLS"),
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1.5, 2},
					Rows: []gopdflib.Row{
						tableHeader(),
						mathRow("Simple addition", "$ a + b = c $", &mathEnabled, ""),
						mathRow("Greek symbols", "$ alpha + beta = pi $", &mathEnabled, "#F2F4F4"),
						mathRow("Comparison operators", "$ x <= y $", &mathEnabled, ""),
						mathRow("Number sets (Real)", "$ x in RR $", &mathEnabled, "#F2F4F4"),
						mathRow("Arrows", "$ a -> b $", &mathEnabled, ""),
					},
				},
			},

			// Section 2: Superscripts & Subscripts
			sectionHeader("SECTION 2: SUPERSCRIPTS & SUBSCRIPTS"),
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1.5, 2},
					Rows: []gopdflib.Row{
						tableHeader(),
						mathRow("Superscript (x²)", "$ x^2 $", &mathEnabled, ""),
						mathRow("Subscript (xᵢ)", "$ x_i $", &mathEnabled, "#F2F4F4"),
						mathRow("Combined super/sub", "$ a_1^2 + b_2^3 = c $", &mathEnabled, ""),
						mathRowWithHeight("Area formula", "$ A = pi r^2 $", &mathEnabled, "#F2F4F4", 30),
					},
				},
			},

			// Section 3: Fractions & Roots
			sectionHeader("SECTION 3: FRACTIONS & ROOTS"),
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1.5, 2},
					Rows: []gopdflib.Row{
						tableHeader(),
						mathRowWithHeight("Simple fraction", "$ frac(a, b) $", &mathEnabled, "", 35),
						mathRowWithHeight("Square root", "$ sqrt(x) $", &mathEnabled, "#F2F4F4", 30),
						mathRowWithHeight("Pythagorean theorem", "$ c = sqrt(a^2 + b^2) $", &mathEnabled, "", 35),
						mathRowWithHeight("Quadratic formula", "$ x = frac(-b, 2 a) $", &mathEnabled, "#F2F4F4", 40),
					},
				},
			},

			// Section 4: Functions & Delimiters
			sectionHeader("SECTION 4: FUNCTIONS & DELIMITERS"),
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1.5, 2},
					Rows: []gopdflib.Row{
						tableHeader(),
						mathRow("Absolute value", "$ abs(x) $", &mathEnabled, ""),
						mathRow("Floor function", "$ floor(x) $", &mathEnabled, "#F2F4F4"),
						mathRow("Ceiling function", "$ ceil(x) $", &mathEnabled, ""),
						mathRowWithHeight("Binomial coefficient", "$ binom(n, k) $", &mathEnabled, "#F2F4F4", 40),
						mathRow("Vector", "$ vec(a, b, c) $", &mathEnabled, ""),
					},
				},
			},

			// Section 5: Combined Expressions
			sectionHeader("SECTION 5: COMBINED EXPRESSIONS"),
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   2,
					ColumnWidths: []float64{1.5, 2},
					Rows: []gopdflib.Row{
						tableHeader(),
						mathRowWithHeight("Euler's identity", "$ e^(i pi) + 1 = 0 $", &mathEnabled, "", 30),
						mathRowWithHeight("Energy-mass equiv.", "$ E = m c^2 $", &mathEnabled, "#F2F4F4", 30),
						mathRow("Norm of a vector", "$ norm(x) $", &mathEnabled, ""),
						mathRowWithHeight("Sum notation", "$ sum_(i=1)^n i $", &mathEnabled, "#F2F4F4", 35),
					},
				},
			},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:8:000:center",
			Text: "TYPST MATH SYNTAX SHOWCASE | GOPDFLIB",
		},
	}
}

func sectionHeader(title string) gopdflib.Element {
	return gopdflib.Element{
		Type: "table",
		Table: &gopdflib.Table{
			MaxColumns:   1,
			ColumnWidths: []float64{1},
			Rows: []gopdflib.Row{
				{Row: []gopdflib.Cell{
					{
						Props:     "Helvetica:12:100:left:1:1:1:1",
						Text:      title,
						BgColor:   "#2E4053",
						TextColor: "#FFFFFF",
					},
				}},
			},
		},
	}
}

func tableHeader() gopdflib.Row {
	return gopdflib.Row{
		Row: []gopdflib.Cell{
			{Props: "Helvetica:10:100:left:1:0:0:1", Text: "Description", BgColor: "#85929E", TextColor: "#FFFFFF"},
			{Props: "Helvetica:10:100:center:0:1:0:1", Text: "Math Expression", BgColor: "#85929E", TextColor: "#FFFFFF"},
		},
	}
}

func mathRow(description, expr string, mathEnabled *bool, bgColor string) gopdflib.Row {
	return gopdflib.Row{
		Row: []gopdflib.Cell{
			{Props: "Helvetica:10:000:left:1:0:0:1", Text: description, BgColor: bgColor},
			{Props: "Helvetica:12:000:center:0:1:0:1", Text: expr, MathEnabled: mathEnabled, BgColor: bgColor},
		},
	}
}

func mathRowWithHeight(description, expr string, mathEnabled *bool, bgColor string, height float64) gopdflib.Row {
	return gopdflib.Row{
		Row: []gopdflib.Cell{
			{Props: "Helvetica:10:000:left:1:0:0:1", Text: description, BgColor: bgColor},
			{Props: "Helvetica:14:000:center:0:1:0:1", Text: expr, MathEnabled: mathEnabled, BgColor: bgColor, Height: floatPtr(height)},
		},
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
