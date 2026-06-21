package pdf

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

func TestPDFAHandlerGenerateXMPMetadataUsesCachedTemplateSections(t *testing.T) {
	handler := NewPDFAHandler(&models.PDFAConfig{
		Conformance: "4",
		Title:       "Quarterly Note",
		Author:      "Ops Team",
		Subject:     "Desk Summary",
		Keywords:    "alpha, beta, gamma",
		Creator:     "bench-runner",
	}, &PageManager{NextObjectID: 9}, nil)

	objectID, metadata := handler.GenerateXMPMetadata("abc123", time.Date(2026, time.June, 22, 10, 11, 12, 0, time.UTC))
	if objectID != 9 {
		t.Fatalf("metadata object id = %d, want 9", objectID)
	}
	if handler.pageManager.NextObjectID != 10 {
		t.Fatalf("next object id = %d, want 10", handler.pageManager.NextObjectID)
	}

	wantSnippets := []string{
		"<xmp:CreateDate>2026-06-22T10:11:12Z</xmp:CreateDate>",
		"<xmp:ModifyDate>2026-06-22T10:11:12Z</xmp:ModifyDate>",
		"<xmp:MetadataDate>2026-06-22T10:11:12Z</xmp:MetadataDate>",
		"<xmpMM:DocumentID>uuid:abc123</xmpMM:DocumentID>",
		"<xmpMM:InstanceID>uuid:abc123</xmpMM:InstanceID>",
		"<rdf:li xml:lang=\"x-default\">Quarterly Note</rdf:li>",
		"<rdf:li>Ops Team</rdf:li>",
		"<rdf:li>alpha</rdf:li>",
		"<rdf:li>beta</rdf:li>",
		"<rdf:li>gamma</rdf:li>",
	}
	for _, snippet := range wantSnippets {
		if !strings.Contains(metadata, snippet) {
			t.Fatalf("metadata missing snippet %q", snippet)
		}
	}
}
