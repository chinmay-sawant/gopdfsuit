package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
)

// PreparedTemplate owns one translated copy of a PDFTemplate for repeated
// generation. It is safe to use concurrently after construction. The source
// PDFTemplate may be reused or changed independently after PrepareTemplate
// returns.
type PreparedTemplate struct {
	template models.PDFTemplate
}

// PrepareTemplate translates a public template once so repeated generation
// does not repeat the public-to-internal ownership conversion. The returned
// prepared template owns its translated slices and can be used concurrently.
func PrepareTemplate(template PDFTemplate) (*PreparedTemplate, error) {
	ensureRuntimePools()
	in, err := toInternalTemplate(template)
	if err != nil {
		return nil, fmt.Errorf("%w: gopdflib: prepare template translation: %w", ErrInvalidInput, err)
	}
	return &PreparedTemplate{template: in}, nil
}

// GeneratePDF creates a PDF from the prepared template and returns owned PDF
// bytes. The prepared template remains reusable after this call.
func (p *PreparedTemplate) GeneratePDF() ([]byte, error) {
	doc, err := p.GeneratePDFBorrowed()
	if err != nil {
		return nil, err
	}
	defer doc.Release()
	return doc.CopyBytes(), nil
}

// GeneratePDFBorrowed creates a PDF from the prepared template without
// cloning the final pooled assembly buffer. The caller owns the returned
// buffer until Release and may call this method concurrently.
func (p *PreparedTemplate) GeneratePDFBorrowed() (*BorrowedPDF, error) {
	if p == nil {
		return nil, fmt.Errorf("%w: gopdflib: prepared template is nil", ErrInvalidInput)
	}
	ensureRuntimePools()
	doc, err := pdf.GenerateTemplatePDFBorrowed(p.template)
	if err != nil {
		return nil, wrapEngineError("gopdflib: GeneratePDFBorrowed prepared template", err)
	}
	return doc, nil
}
