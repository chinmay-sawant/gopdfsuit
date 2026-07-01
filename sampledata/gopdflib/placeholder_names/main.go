// Package main loads a JSON PDF template with {firstname} and {lastname} placeholders,
// fills them with random data, and generates a PDF via gopdflib.
//
// Run from this directory:
//
//	go run .
//
// Or from the repo root:
//
//	go run ./sampledata/gopdflib/placeholder_names
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

var (
	firstNames = []string{
		"Aarav", "Priya", "Rohan", "Ananya", "Vikram", "Neha", "Arjun", "Kavya",
		"James", "Emily", "Michael", "Sarah", "David", "Olivia", "Daniel", "Sophia",
	}
	lastNames = []string{
		"Sharma", "Patel", "Iyer", "Reddy", "Kapoor", "Nair", "Mehta", "Desai",
		"Johnson", "Williams", "Brown", "Davis", "Miller", "Wilson", "Anderson", "Thomas",
	}
)

func main() {
	fmt.Println("=== gopdflib Placeholder Names Example ===")

	templatePath := resolveTemplatePath()
	fmt.Printf("Loading template: %s\n", templatePath)

	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Error reading template: %v\n", err)
		os.Exit(1)
	}

	var templateMap map[string]interface{}
	if err := json.Unmarshal(templateBytes, &templateMap); err != nil {
		fmt.Printf("Error parsing template JSON: %v\n", err)
		os.Exit(1)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := map[string]string{
		"{firstname}": firstNames[rng.Intn(len(firstNames))],
		"{lastname}":  lastNames[rng.Intn(len(lastNames))],
	}
	processMap(templateMap, data)

	filledJSON, err := json.Marshal(templateMap)
	if err != nil {
		fmt.Printf("Error marshaling filled template: %v\n", err)
		os.Exit(1)
	}

	var template gopdflib.PDFTemplate
	if err := json.Unmarshal(filledJSON, &template); err != nil {
		fmt.Printf("Error unmarshaling into PDFTemplate: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generating PDF for: %s %s\n", data["{firstname}"], data["{lastname}"])
	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		fmt.Printf("Error generating PDF: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(filepath.Dir(templatePath), "name_card_output.pdf")
	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
		fmt.Printf("Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PDF saved to: %s (%d bytes)\n", outputPath, len(pdfBytes))
	fmt.Println("=== Done ===")
}

func resolveTemplatePath() string {
	candidates := []string{
		"name_card_template.json",
		filepath.Join("sampledata", "gopdflib", "placeholder_names", "name_card_template.json"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			abs, err := filepath.Abs(path)
			if err == nil {
				return abs
			}
			return path
		}
	}
	return candidates[0]
}

func processMap(m map[string]interface{}, data map[string]string) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = replacePlaceholders(val, data)
		case map[string]interface{}:
			processMap(val, data)
		case []interface{}:
			processSlice(val, data)
		}
	}
}

func processSlice(s []interface{}, data map[string]string) {
	for i, v := range s {
		switch val := v.(type) {
		case string:
			s[i] = replacePlaceholders(val, data)
		case map[string]interface{}:
			processMap(val, data)
		case []interface{}:
			processSlice(val, data)
		}
	}
}

func replacePlaceholders(s string, data map[string]string) string {
	re := regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		if val, ok := data[match]; ok {
			return val
		}
		return match
	})
}