package main

import (
	"log"
	"os"

	"github.com/chinmay-sawant/gopdfsuit/internal/pdf"
)

func main() {
	pdfBytes, err := os.ReadFile("sampledata/pdf+xfdf/us_hospital_encounter_acroform.pdf")
	if err != nil {
		log.Fatalf("read pdf: %v", err)
	}
	xfdfBytes, err := os.ReadFile("sampledata/pdf+xfdf/us_hospital_encounter_data.xfdf")
	if err != nil {
		log.Fatalf("read xfdf: %v", err)
	}

	detected, err := pdf.DetectFormFieldsAdvanced(pdfBytes)
	if err != nil {
		log.Fatalf("detect: %v", err)
	}
	log.Printf("detected %d fields", len(detected))
	for k, v := range detected {
		log.Printf("%s => %s", k, v)
	}
	out, err := pdf.FillPDFWithXFDFAdvanced(pdfBytes, xfdfBytes)
	if err != nil {
		log.Fatalf("fill: %v", err)
	}
	_ = os.WriteFile("internal/pdf/debug_filled.pdf", out, 0644)
	log.Printf("wrote debug_filled.pdf size=%d", len(out))
}
