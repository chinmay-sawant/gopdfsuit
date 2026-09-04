package handlers

import (
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/form"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

//go:generate go run go.uber.org/mock/mockgen@latest -destination=mocks/mock_services.go -package=mocks github.com/chinmay-sawant/gopdfsuit/v6/internal/handlers PDFService

// PDFService abstracts PDF operations used by HTTP handlers (mockable in unit tests).
type PDFService interface {
	GenerateTemplatePDF(template models.PDFTemplate) ([]byte, error)
	FillPDFWithXFDF(pdfBytes, xfdfBytes []byte) ([]byte, error)
	MergePDFs(pdfBytesList [][]byte) ([]byte, error)
	SplitPDF(pdfBytes []byte, spec merge.SplitSpec) ([][]byte, error)
	CompressPDF(pdfBytes []byte, opts compress.Options) ([]byte, error)
	GetFonts() []models.FontInfo
	RegisterFont(name string, data []byte) error
	HTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error)
	HTMLToImage(req models.HTMLToImageRequest) ([]byte, error)
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

func (defaultPDFService) CompressPDF(pdfBytes []byte, opts compress.Options) ([]byte, error) {
	return compress.CompressPDF(pdfBytes, opts)
}

func (defaultPDFService) GetFonts() []models.FontInfo {
	return pdf.GetAvailableFonts()
}

func (defaultPDFService) RegisterFont(name string, data []byte) error {
	return pdf.GetFontRegistry().RegisterFontFromData(name, data)
}

func (defaultPDFService) HTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	return pdf.ConvertHTMLToPDF(req)
}

func (defaultPDFService) HTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	return pdf.ConvertHTMLToImage(req)
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
