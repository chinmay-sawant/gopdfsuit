package pdf

import (
	"reflect"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

func TestCollectAllStandardFontsInTemplateUsesPrecomputedHint(t *testing.T) {
	template := models.PDFTemplate{
		Title: models.Title{
			Text:  "Ignored",
			Props: "Courier:12:100:left:0:0:0:0",
		},
	}
	template.SetPrecomputedStandardFonts("Helvetica", "Times-Roman")

	got := collectAllStandardFontsInTemplate(template)
	want := map[string]bool{
		"Helvetica":   true,
		"Times-Roman": true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("font hint result = %#v, want %#v", got, want)
	}
}
