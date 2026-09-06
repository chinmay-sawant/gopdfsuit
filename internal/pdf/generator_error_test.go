package pdf

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

func TestGenerateTemplatePDFRejectsIncompleteRequiredSettings(t *testing.T) {
	tests := map[string]struct {
		config models.Config
		want   string
	}{
		"encryption without owner password": {
			config: models.Config{
				Security: &models.SecurityConfig{Enabled: true},
			},
			want: "owner password",
		},
		"invalid signer": {
			config: models.Config{
				Signature: &models.SignatureConfig{
					Enabled:        true,
					CertificatePEM: "not a certificate",
					PrivateKeyPEM:  "not a private key",
				},
			},
			want: "signer",
		},
		"unavailable custom font": {
			config: models.Config{
				CustomFonts: []models.CustomFontConfig{{
					Name:     "MissingFont",
					FilePath: "/path/that/does/not/exist.ttf",
				}},
			},
			want: "custom font MissingFont",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := GenerateTemplatePDF(models.PDFTemplate{Config: test.config})
			if err == nil {
				t.Fatal("expected generation to fail")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(test.want)) {
				t.Fatalf("error = %q, want substring %q", err, test.want)
			}
		})
	}
}
