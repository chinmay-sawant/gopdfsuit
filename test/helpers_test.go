package tests

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/form"
)

func repoRoot(t testing.TB) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// test/ is one level below repo root
	root := filepath.Clean(filepath.Join(wd, ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root not found from %s: %v", wd, err)
	}
	return root
}

func samplePath(t testing.TB, parts ...string) string {
	t.Helper()
	all := append([]string{repoRoot(t), "sampledata"}, parts...)
	return filepath.Join(all...)
}

func writeMultipartPDFXfdf(t testing.TB, pdfBytes, xfdfBytes []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	pdfPart, err := writer.CreateFormFile("pdf", "form.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pdfPart.Write(pdfBytes); err != nil {
		t.Fatal(err)
	}
	xfdfPart, err := writer.CreateFormFile("xfdf", "data.xfdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := xfdfPart.Write(xfdfBytes); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func writeMultipartPDF(t testing.TB, fieldName, filename string, pdfBytes []byte, extraFields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pdfBytes); err != nil {
		t.Fatal(err)
	}
	for k, v := range extraFields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func assertFilledXFDFFields(t testing.TB, filledPDF, xfdf []byte) {
	t.Helper()
	expected, err := form.ParseXFDF(xfdf)
	if err != nil {
		t.Fatalf("parse xfdf: %v", err)
	}
	got, err := form.DetectFormFieldsAdvanced(filledPDF)
	if err != nil {
		t.Fatalf("detect filled fields: %v", err)
	}
	for name, want := range expected {
		if have := got[name]; have != want {
			t.Errorf("field %q: got %q want %q", name, have, want)
		}
	}
}
