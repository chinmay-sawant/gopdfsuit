// Save this file as NoAcroForm.go

package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// Structs to unmarshal the XFDF XML data into.
type XFDF struct {
	XMLName xml.Name `xml:"xfdf"`
	Fields  Fields   `xml:"fields"`
}

type Fields struct {
	XMLName xml.Name `xml:"fields"`
	Fields  []Field  `xml:"field"`
}

type Field struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

// escapePDFString escapes characters that are special inside a PDF string literal.
func escapePDFString(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`(`, `\(`,
		`)`, `\)`,
	)
	return r.Replace(s)
}

// parseXFDF reads an XFDF file and returns a map of field names to values.
func parseXFDF(filename string) (map[string]string, error) {
	xmlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read XFDF file %s: %w", filename, err)
	}

	var xfdfData XFDF
	if err := xml.Unmarshal(xmlFile, &xfdfData); err != nil {
		return nil, fmt.Errorf("failed to parse XFDF xml: %w", err)
	}

	fieldMap := make(map[string]string)
	for _, field := range xfdfData.Fields.Fields {
		fieldMap[field.Name] = field.Value
	}

	fmt.Printf("Successfully parsed %s, found %d field values.\n\n", filename, len(fieldMap))
	return fieldMap, nil
}

func main() {
	// Define the input file names
	inputFile := "./sampledata/patient2/patient2.pdf"
	xfdfFile := "./sampledata/patient2/patient2.xfdf" // Updated to match your path
	outputFile := "temp2.pdf"

	// 1. Parse the XFDF file to get the data map.
	xfdfData, err := parseXFDF(xfdfFile)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Read the entire PDF file.
	content, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input PDF file %s: %v", inputFile, err)
	}
	fmt.Printf("Successfully read PDF %s (%d bytes).\n\n", inputFile, len(content))

	// Pre-compile regular expressions
	reObject := regexp.MustCompile(`(?s)(\d+\s+\d+\s+obj)(.*?)(\bendobj\b)`)
	reIsAnnot := regexp.MustCompile(`/Type\s*/Annot`)
	reIsWidget := regexp.MustCompile(`/Subtype\s*/Widget`)
	reIsBtn := regexp.MustCompile(`/FT\s*/Btn`)
	reIsTx := regexp.MustCompile(`/FT\s*/Tx`)
	reFieldNameT := regexp.MustCompile(`/T\s*\((.*?)\)`)
	reRemoveAP := regexp.MustCompile(`/AP\s*<<.*?>>`)

	matches := reObject.FindAllSubmatchIndex(content, -1)
	if matches == nil {
		log.Fatal("No PDF objects found.")
	}

	fmt.Println("--- Scanning and Filling Form Fields ---")

	var newContent bytes.Buffer
	var lastIndex int = 0
	modifiedCount := 0

	for _, match := range matches {
		newContent.Write(content[lastIndex:match[0]])

		objHeader := content[match[2]:match[3]]
		objBody := content[match[4]:match[5]]
		objFooter := content[match[6]:match[7]]

		if reIsAnnot.Match(objBody) && reIsWidget.Match(objBody) {
			// *** THE CRITICAL FIX IS HERE ***
			// First, safely check if a field name exists before trying to access it.
			fieldNameMatch := reFieldNameT.FindSubmatch(objBody)
			if len(fieldNameMatch) < 2 {
				// This widget has no /T key. It cannot be matched to the XFDF.
				// Write the original object and skip to the next one.
				newContent.Write(content[match[0]:match[1]])
				lastIndex = match[1]
				continue
			}
			fieldName := string(fieldNameMatch[1])
			// *** END OF FIX ***

			// Check if we have a value for this field from our XFDF map
			if value, ok := xfdfData[fieldName]; ok {
				var valueToAdd []byte
				modifiedObjBody := objBody

				if reIsBtn.Match(objBody) {
					// Handle buttons: Check for "On" or "Yes" to turn on.
					if strings.EqualFold(value, "On") || strings.EqualFold(value, "Yes") {
						valueToAdd = []byte(" /AS /On /V /On")
					}
				} else if reIsTx.Match(objBody) {
					// Handle text fields: remove stale appearance and add new value.
					if value != "N/A" { // Don't fill fields marked as N/A
						modifiedObjBody = reRemoveAP.ReplaceAll(modifiedObjBody, []byte(""))
						safeValue := escapePDFString(value)
						valueToAdd = []byte(fmt.Sprintf(" /V (%s)", safeValue))
					}
				}

				if valueToAdd != nil {
					loc := reIsAnnot.FindIndex(modifiedObjBody)
					if loc != nil {
						modifiedCount++
						fmt.Printf("-> Filling field '%s'\n", fieldName)

						insertionIndex := loc[1]
						finalBody := new(bytes.Buffer)
						finalBody.Write(modifiedObjBody[:insertionIndex])
						finalBody.Write(valueToAdd)
						finalBody.Write(modifiedObjBody[insertionIndex:])

						newContent.Write(objHeader)
						newContent.Write(finalBody.Bytes())
						newContent.Write(objFooter)
						lastIndex = match[1]
						continue
					}
				}
			}
		}

		// If no modifications were made, write the original object content.
		newContent.Write(content[match[0]:match[1]])
		lastIndex = match[1]
	}

	newContent.Write(content[lastIndex:])

	err = os.WriteFile(outputFile, newContent.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Failed to write to output file %s: %v", outputFile, err)
	}

	fmt.Println("\n--- Processing Complete ---")
	fmt.Printf("Modified and filled %d form field objects.\n", modifiedCount)
	fmt.Printf("Output saved to: %s\n", outputFile)
}
