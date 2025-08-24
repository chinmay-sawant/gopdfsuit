package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chinmay-sawant/gopdfsuit/internal/pdf"
)

// FillAndSave opens PDF, applies XFDF, writes output.
func FillAndSave(pdfPath, xfdfPath, outPath string) error {
	doc, err := pdf.Open(pdfPath)
	if err != nil {
		return err
	}
	fieldMap, err := pdf.ParseXFDFUk(xfdfPath)
	if err != nil {
		return err
	}
	modBytes, count, err := doc.FillFields(fieldMap)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, modBytes, 0644); err != nil {
		return err
	}
	fmt.Println("Fields modified:", count)
	fmt.Println("Output:", outPath)
	return nil
}

func DemoOpen(path string) {
	doc, err := pdf.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Linearized:", doc.Linearized)
	fmt.Println("Root Obj:", doc.RootObj)
	fmt.Println("Size:", doc.Size)
	fmt.Println("XRef entries:", len(doc.XRef))
}

func main() {
	// Adjust paths as needed.
	inPDF := "./sampledata/patientreg/patientreg.pdf"
	xfdf := "./sampledata/patientreg/patientreg.xfdf"
	outPDF := "temp2.pdf"

	if err := FillAndSave(inPDF, xfdf, outPDF); err != nil {
		log.Println("Fill error:", err)
		return
	}

	// Do not re-open modified PDF with our simplistic parser (xref offsets unchanged after edits).
	// DemoOpen(outPDF)
}
