// Package main demonstrates loading a PDF template from a JSON file and generating the PDF.
// This example reads sampledata/editor/financial_digitalsignature.json, matches it to the
// gopdflib.PDFTemplate structure, and generates the output.
//
// Run with: go run sampledata/gopdflib/load_from_json/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

func main() {
	fmt.Println("=== gopdflib Load from JSON Example ===")
	fmt.Println()

	// Path to the JSON file
	// Assuming execution from project root
	jsonPath := "sampledata/editor/financial_digitalsignature.json"

	// Check if file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		fmt.Printf("Could not find JSON file at: %s\n", jsonPath)
		fmt.Println("Please run this example from the project root: go run sampledata/gopdflib/load_from_json/main.go")
		os.Exit(1)
	}

	fmt.Printf("Loading data from: %s\n", jsonPath)

	// Read JSON file
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal JSON into PDFTemplate
	var template gopdflib.PDFTemplate
	err = json.Unmarshal(jsonData, &template)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully loaded and parsed JSON template.")

	// Benchmark generation
	iterations := 10
	fmt.Printf("\nRunning %d iterations...\n", iterations)

	var totalDuration time.Duration
	var lastPDF []byte

	// Warm-up run
	_, err = gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("Error during warm-up: %v\n", err)
		os.Exit(1)
	}

	for i := 1; i <= iterations; i++ {
		start := time.Now()
		pdfBytes, err := gopdflib.GeneratePDF(template)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("Iteration %d: ERROR - %v\n", i, err)
			os.Exit(1)
		}

		totalDuration += elapsed
		lastPDF = pdfBytes
		fmt.Printf("  Iteration %2d: %8.3f ms  (%d bytes)\n", i, float64(elapsed.Microseconds())/1000.0, len(pdfBytes))
	}

	avgDuration := totalDuration / time.Duration(iterations)

	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("  Average time:  %.3f ms\n", float64(avgDuration.Microseconds())/1000.0)

	// Save the PDF
	outputPath := "financial_from_json.pdf"
	err = os.WriteFile(outputPath, lastPDF, 0644)
	if err != nil {
		fmt.Printf("Error saving PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PDF saved to: %s\n", outputPath)
}
