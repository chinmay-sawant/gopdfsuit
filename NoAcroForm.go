// Save this file as temp2.go

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
)

// extractStringValue is a helper function to extract a value from a PDF dictionary string.
func extractStringValue(data []byte, re *regexp.Regexp) string {
	match := re.FindSubmatch(data)
	if len(match) > 1 {
		return string(match[1])
	}
	return "N/A"
}

func main() {
	// Define the input and output file names
	inputFile := "./sampledata/patient2/patient2.pdf" // Updated to match your path
	outputFile := "temp2.pdf"

	content, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file %s: %v", inputFile, err)
	}

	fmt.Printf("Successfully read %s (%d bytes).\n\n", inputFile, len(content))

	// Pre-compile regular expressions for efficiency
	reObject := regexp.MustCompile(`(?s)(\d+\s+\d+\s+obj)(.*?)(\bendobj\b)`)
	reIsAnnot := regexp.MustCompile(`/Type\s*/Annot`)
	reIsWidget := regexp.MustCompile(`/Subtype\s*/Widget`)
	reHasValue := regexp.MustCompile(`/V\s+`)
	reIsBtn := regexp.MustCompile(`/FT\s*/Btn`)
	reIsTx := regexp.MustCompile(`/FT\s*/Tx`)
	reFieldNameT := regexp.MustCompile(`/T\s*\((.*?)\)`)
	reFieldNameTU := regexp.MustCompile(`/TU\s*\((.*?)\)`)

	// *** THE KEY NEW REGEX ***
	// This regex finds and removes the entire /AP dictionary.
	// It looks for /AP, optional whitespace, <<, then anything (non-greedy) until >>.
	reRemoveAP := regexp.MustCompile(`/AP\s*<<.*?>>`)

	matches := reObject.FindAllSubmatchIndex(content, -1)
	if matches == nil {
		log.Fatal("No PDF objects found. The file may be invalid or empty.")
	}

	fmt.Println("--- Scanning for Form Fields ---")

	var newContent bytes.Buffer
	var lastIndex int = 0
	modifiedCount := 0

	for _, match := range matches {
		newContent.Write(content[lastIndex:match[0]])

		objHeader := content[match[2]:match[3]]
		objBody := content[match[4]:match[5]]
		objFooter := content[match[6]:match[7]]

		if reIsAnnot.Match(objBody) && reIsWidget.Match(objBody) {
			fieldName := extractStringValue(objBody, reFieldNameT)
			fieldTU := extractStringValue(objBody, reFieldNameTU)
			fmt.Printf("Found Field: Name=%-25s TU=%-25s\n", fieldName, fieldTU)

			if !reHasValue.Match(objBody) {
				var valueToAdd []byte
				modifiedObjBody := objBody // Start with the original body

				if fieldName == "Today" && reIsTx.Match(objBody) {
					// For text fields, remove the stale /AP dict and add the new /V value
					modifiedObjBody = reRemoveAP.ReplaceAll(modifiedObjBody, []byte(""))
					valueToAdd = []byte(" /V (todaysdate)")
				} else if reIsBtn.Match(objBody) {
					// For buttons, KEEP the /AP dict but set the /AS and /V states
					valueToAdd = []byte(" /AS /On /V /On")
				} else if reIsTx.Match(objBody) {
					// For other text fields, do the same as the "Today" field
					modifiedObjBody = reRemoveAP.ReplaceAll(modifiedObjBody, []byte(""))
					valueToAdd = []byte(" /V (Dummy)")
				}

				if valueToAdd != nil {
					// Find the insertion point in the (potentially already modified) body
					loc := reIsAnnot.FindIndex(modifiedObjBody)
					if loc != nil {
						modifiedCount++
						fmt.Printf("-> MODIFIED object %s\n", bytes.TrimSpace(objHeader))

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
	fmt.Printf("Modified %d form field objects.\n", modifiedCount)
	fmt.Printf("Output saved to: %s\n", outputFile)
}
