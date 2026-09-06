package mocks

import (
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf"
)

// FastMockPDFService embeds MockPDFService and additionally implements the
// handlers.FastGenerateService borrowed-render seam, so handler tests can
// cover the pooled hot path in generate.go. The embedded mock serves the
// rest of the PDFService surface (including the ReadUpload/UploadLimit/
// ClassifyError policy stubs); BorrowedFunc serves the seam call.
type FastMockPDFService struct {
	*MockPDFService
	BorrowedFunc func(models.PDFTemplate) (*pdf.BorrowedPDF, error)
}

// GenerateTemplatePDFBorrowed implements handlers.FastGenerateService.
func (m *FastMockPDFService) GenerateTemplatePDFBorrowed(template models.PDFTemplate) (*pdf.BorrowedPDF, error) {
	return m.BorrowedFunc(template)
}
