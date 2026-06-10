package handlers

import (
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/form"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/merge"
)

//go:generate go run go.uber.org/mock/mockgen@latest -destination=mocks/mock_services.go -package=mocks github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers PDFService

// PDFService abstracts PDF operations used by HTTP handlers (mockable in unit tests).
type PDFService interface {
	GenerateTemplatePDF(template models.PDFTemplate) ([]byte, error)
	FillPDFWithXFDF(pdfBytes, xfdfBytes []byte) ([]byte, error)
	MergePDFs(pdfBytesList [][]byte) ([]byte, error)
	SplitPDF(pdfBytes []byte, spec merge.SplitSpec) ([][]byte, error)
}

type defaultPDFService struct{}

func (defaultPDFService) GenerateTemplatePDF(template models.PDFTemplate) ([]byte, error) {
	return pdf.GenerateTemplatePDF(template)
}

func (defaultPDFService) FillPDFWithXFDF(pdfBytes, xfdfBytes []byte) ([]byte, error) {
	return form.FillPDFWithXFDF(pdfBytes, xfdfBytes)
}

func (defaultPDFService) MergePDFs(pdfBytesList [][]byte) ([]byte, error) {
	return merge.MergePDFs(pdfBytesList)
}

func (defaultPDFService) SplitPDF(pdfBytes []byte, spec merge.SplitSpec) ([][]byte, error) {
	return merge.SplitPDF(pdfBytes, spec)
}

// pdfService is the active PDF backend (swap in tests via SetPDFService).
var pdfService PDFService = defaultPDFService{}

// SetPDFService replaces the PDF backend (for gomock unit tests). Pass nil to restore defaults.
func SetPDFService(s PDFService) {
	if s == nil {
		pdfService = defaultPDFService{}
		return
	}
	pdfService = s
}