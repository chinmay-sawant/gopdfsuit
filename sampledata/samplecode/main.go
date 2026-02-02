package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	pdf "github.com/chinmay-sawant/gopdfsuit/v4-client"
)

// PatientData model
type PatientData struct {
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	MiddleName            string `json:"middle_name"`
	DOB                   string `json:"dob"`
	SSN                   string `json:"ssn"`
	PatientID             string `json:"patient_id"`
	StreetAddress         string `json:"street_address"`
	City                  string `json:"city"`
	State                 string `json:"state"`
	ZipCode               string `json:"zip_code"`
	HomePhone             string `json:"home_phone"`
	CellPhone             string `json:"cell_phone"`
	WorkPhone             string `json:"work_phone"`
	Email                 string `json:"email"`
	EmergencyName         string `json:"emergency_name"`
	EmergencyRelationship string `json:"emergency_relationship"`
	EmergencyPhone        string `json:"emergency_phone"`
	EmergencyAltPhone     string `json:"emergency_alt_phone"`
	InsuranceCompany      string `json:"insurance_company"`
	PolicyNumber          string `json:"policy_number"`
	GroupNumber           string `json:"group_number"`
	SubscriberID          string `json:"subscriber_id"`
	SubscriberName        string `json:"subscriber_name"`
	SubscriberDOB         string `json:"subscriber_dob"`
	PrimaryPhysician      string `json:"primary_physician"`
	PhysicianPhone        string `json:"physician_phone"`
	PreferredPharmacy     string `json:"preferred_pharmacy"`
	PharmacyPhone         string `json:"pharmacy_phone"`
	CurrentMedications    string `json:"current_medications"`
	KnownAllergies        string `json:"known_allergies"`
	ReasonForVisit        string `json:"reason_for_visit"`
	SymptomsDescription   string `json:"symptoms_description"`
	SymptomDuration       string `json:"symptom_duration"`
	PatientSignature      string `json:"patient_signature"`
	SignatureDate         string `json:"signature_date"`
	GuardianName          string `json:"guardian_name"`
	GuardianRelationship  string `json:"guardian_relationship"`
	ReceivedBy            string `json:"received_by"`
	ReceivedDatetime      string `json:"received_datetime"`
	VerifiedBy            string `json:"verified_by"`
	MRN                   string `json:"mrn"`
}

func main() {
	startTime := time.Now()

	// 1. Define the data model
	patient := PatientData{
		FirstName:             "Michael",
		LastName:              "Thompson",
		MiddleName:            "James",
		DOB:                   "03/15/1985",
		SSN:                   "***-**-4589",
		PatientID:             "PT-2024-78542",
		StreetAddress:         "4521 Oak Ridge Boulevard, Apt 12B",
		City:                  "Austin",
		State:                 "Texas",
		ZipCode:               "78745",
		HomePhone:             "(512) 555-1234",
		CellPhone:             "(512) 555-9876",
		WorkPhone:             "(512) 555-4567",
		Email:                 "m.thompson@email.com",
		EmergencyName:         "Sarah Thompson",
		EmergencyRelationship: "Spouse",
		EmergencyPhone:        "(512) 555-2468",
		EmergencyAltPhone:     "(512) 555-1357",
		InsuranceCompany:      "Blue Cross Blue Shield",
		PolicyNumber:          "BCB-78542136",
		GroupNumber:           "GRP-45892",
		SubscriberID:          "SUB-123456789",
		SubscriberName:        "Michael J. Thompson",
		SubscriberDOB:         "03/15/1985",
		PrimaryPhysician:      "Dr. Robert Williams",
		PhysicianPhone:        "(512) 555-8900",
		PreferredPharmacy:     "CVS Pharmacy #4521",
		PharmacyPhone:         "(512) 555-7890",
		CurrentMedications:    "Metformin 500mg, Lisinopril 10mg, Aspirin 81mg",
		KnownAllergies:        "Penicillin, Shellfish",
		ReasonForVisit:        "Annual Physical Examination",
		SymptomsDescription:   "Occasional fatigue, routine checkup for diabetes management",
		SymptomDuration:       "2-3 weeks",
		PatientSignature:      "Michael J. Thompson",
		SignatureDate:         "11/26/2025",
		GuardianName:          "N/A",
		GuardianRelationship:  "N/A",
		ReceivedBy:            "Jane Smith, RN",
		ReceivedDatetime:      "11/26/2025 09:30 AM",
		VerifiedBy:            "Mary Johnson",
		MRN:                   "MRN-2024-785421",
	}

	// Convert model to map for placeholder replacement
	data := make(map[string]string)
	jsonData, _ := json.Marshal(patient)
	var tempMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &tempMap); err != nil {
		fmt.Printf("Error unmarshaling patient data: %v\n", err)
		return
	}
	for k, v := range tempMap {
		data["{"+k+"}"] = fmt.Sprint(v)
	}

	// 2. Read the template file
	templatePath := "./us_patient_healthcare_form_template.json"
	templateBytes, err := os.ReadFile(templatePath)
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
	outputPath := "./us_patient_healthcare_form_filled.pdf"
	if err := os.WriteFile(outputPath, outputBytes, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	// 7. Send to PDF Client
	baseURL := "http://localhost:8080" // Replace with actual base URL
	client := pdf.NewClient(
		baseURL,
		pdf.WithTimeout(60*time.Second),
		pdf.WithMaxRetries(3),
		pdf.WithHeader("Authorization", "Bearer your-token-here"),
	)

	// Create a reader from the bytes
	reader := pdf.NewJSONBytesReader(outputBytes)
	ctx := context.Background()
	doc, err := reader.Read(ctx)
	if err != nil {
		fmt.Printf("Error reading document from bytes: %v\n", err)
		return
	}

	err = client.SendAndSave(ctx, doc, outputPath)
	if err != nil {
		fmt.Printf("Error sending to PDF client: %v\n", err)
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
// Uses regex to find placeholders and look them up in the map
func ReplacePlaceholders(s string, data map[string]string) string {
	re := regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		if val, ok := data[match]; ok {
			return val
		}
		return match
	})
}
