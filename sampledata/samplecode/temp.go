package main

import (
"encoding/json"
"fmt"
"io/ioutil"
"strings"
"time"
)

func main() {
	startTime := time.Now()

	// 1. Define the data map
	data := map[string]string{
		"{first_name}":             "Michael",
		"{last_name}":              "Thompson",
		"{middle_name}":            "James",
		"{dob}":                    "03/15/1985",
		"{ssn}":                    "***-**-4589",
		"{patient_id}":             "PT-2024-78542",
		"{street_address}":         "4521 Oak Ridge Boulevard, Apt 12B",
		"{city}":                   "Austin",
		"{state}":                  "Texas",
		"{zip_code}":               "78745",
		"{home_phone}":             "(512) 555-1234",
		"{cell_phone}":             "(512) 555-9876",
		"{work_phone}":             "(512) 555-4567",
		"{email}":                  "m.thompson@email.com",
		"{emergency_name}":         "Sarah Thompson",
		"{emergency_relationship}": "Spouse",
		"{emergency_phone}":        "(512) 555-2468",
		"{emergency_alt_phone}":    "(512) 555-1357",
		"{insurance_company}":      "Blue Cross Blue Shield",
		"{policy_number}":          "BCB-78542136",
		"{group_number}":           "GRP-45892",
		"{subscriber_id}":          "SUB-123456789",
		"{subscriber_name}":        "Michael J. Thompson",
		"{subscriber_dob}":         "03/15/1985",
		"{primary_physician}":      "Dr. Robert Williams",
		"{physician_phone}":        "(512) 555-8900",
		"{preferred_pharmacy}":     "CVS Pharmacy #4521",
		"{pharmacy_phone}":         "(512) 555-7890",
		"{current_medications}":    "Metformin 500mg, Lisinopril 10mg, Aspirin 81mg",
		"{known_allergies}":        "Penicillin, Shellfish",
		"{reason_for_visit}":       "Annual Physical Examination",
		"{symptoms_description}":   "Occasional fatigue, routine checkup for diabetes management",
		"{symptom_duration}":       "2-3 weeks",
		"{patient_signature}":      "Michael J. Thompson",
		"{signature_date}":         "11/26/2025",
		"{guardian_name}":          "N/A",
		"{guardian_relationship}":  "N/A",
		"{received_by}":            "Jane Smith, RN",
		"{received_datetime}":      "11/26/2025 09:30 AM",
		"{verified_by}":            "Mary Johnson",
		"{mrn}":                    "MRN-2024-785421",
	}

	// 2. Read the template file
	templatePath := "sampledata/acroform/us_patient_healthcare_form_template.json"
	templateBytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Error reading template file: %v\n", err)
		return
	}

	// 3. Unmarshal into a generic map
	var templateMap map[string]interface{}
	if err := json.Unmarshal(templateBytes, &templateMap); err != nil {
		fmt.Printf("Error unmarshaling template: %v\n", err)
		return
	}

	// 4. Process the map recursively
	ProcessMap(templateMap, data)

	// 5. Marshal back to JSON
	outputBytes, err := json.MarshalIndent(templateMap, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling output: %v\n", err)
		return
	}

	// 6. Write to output file
	outputPath := "sampledata/acroform/us_patient_healthcare_form_filled.json"
	if err := ioutil.WriteFile(outputPath, outputBytes, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Successfully generated %s\n", outputPath)
	fmt.Printf("Total processed in seconds: %.6f\n", elapsed.Seconds())
	fmt.Printf("Total processed in ms: %d\n", elapsed.Milliseconds())
}

// ProcessMap recursively traverses the map and replaces placeholders in string values
func ProcessMap(m map[string]interface{}, data map[string]string) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = ReplacePlaceholders(val, data)
		case map[string]interface{}:
			ProcessMap(val, data)
		case []interface{}:
			ProcessSlice(val, data)
		}
	}
}

// ProcessSlice recursively traverses the slice
func ProcessSlice(s []interface{}, data map[string]string) {
	for i, v := range s {
		switch val := v.(type) {
		case string:
			s[i] = ReplacePlaceholders(val, data)
		case map[string]interface{}:
			ProcessMap(val, data)
		case []interface{}:
			ProcessSlice(val, data)
		}
	}
}

// ReplacePlaceholders replaces all occurrences of keys in data with their values
func ReplacePlaceholders(s string, data map[string]string) string {
	for k, v := range data {
		if strings.Contains(s, k) {
			s = strings.ReplaceAll(s, k, v)
		}
	}
	return s
}
