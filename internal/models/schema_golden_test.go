package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/sonic"
)

// Triple-gate anchor (D1): the frontend schema and this golden fixture must
// stay in lockstep with models.PDFTemplate. The same golden file is checked
// by frontend npm run test:schema and bindings/python
// tests/test_golden_template.py, and test:schema is wired into CI via the
// frontend-lint job plus `make test-schema`.
//
// Paths resolve from the package directory (go test runs with the package
// dir as cwd), so the test embeds the live frontend schema and golden
// fixture instead of a frozen copy: drift fails loudly here.

var (
	templateSchemaPath = filepath.Join("..", "..", "frontend", "template.schema.json")
	goldenTemplatePath = filepath.Join("..", "..", "sampledata", "editor", "financial_report.json")
)

func loadTripleGateInputs(t *testing.T) (schemaRaw, goldenRaw []byte) {
	t.Helper()
	var err error
	if schemaRaw, err = os.ReadFile(templateSchemaPath); err != nil {
		t.Fatalf("read %s: %v", templateSchemaPath, err)
	}
	if goldenRaw, err = os.ReadFile(goldenTemplatePath); err != nil {
		t.Fatalf("read %s: %v", goldenTemplatePath, err)
	}
	return schemaRaw, goldenRaw
}

func TestTemplateSchemaGolden(t *testing.T) {
	schemaRaw, goldenRaw := loadTripleGateInputs(t)
	var schema struct {
		Title      string         `json:"title"`
		Properties map[string]any `json:"properties"`
	}
	if err := sonic.Unmarshal(schemaRaw, &schema); err != nil {
		t.Fatalf("unmarshal frontend/template.schema.json: %v", err)
	}
	if schema.Title != "GoPdfSuit PDFTemplate" {
		t.Fatalf("schema title = %q, want %q", schema.Title, "GoPdfSuit PDFTemplate")
	}
	if len(schema.Properties) == 0 {
		t.Fatal("schema has no top-level properties")
	}

	var golden map[string]any
	if err := sonic.Unmarshal(goldenRaw, &golden); err != nil {
		t.Fatalf("unmarshal golden financial_report.json: %v", err)
	}
	if len(golden) == 0 {
		t.Fatal("golden template has no top-level keys")
	}
	for key := range golden {
		if _, ok := schema.Properties[key]; !ok {
			t.Fatalf("golden top-level key %q missing from schema properties", key)
		}
	}

	var tmpl PDFTemplate
	if err := sonic.Unmarshal(goldenRaw, &tmpl); err != nil {
		t.Fatalf("parse golden via models.PDFTemplate: %v", err)
	}
	if tmpl.Config.Page != "A4" {
		t.Fatalf("golden config.page = %q, want A4", tmpl.Config.Page)
	}
	if tmpl.Title.Text == "" {
		t.Fatal("golden title text is empty")
	}
	if len(tmpl.Table) == 0 && len(tmpl.Elements) == 0 {
		t.Fatal("golden template has neither tables nor elements")
	}
}
