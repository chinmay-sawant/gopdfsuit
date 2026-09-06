package handlers

import (
	"context"
	"io"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/form"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/merge"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/redact"
)

//go:generate go run go.uber.org/mock/mockgen@latest -destination=mocks/mock_services.go -package=mocks github.com/chinmay-sawant/gopdfsuit/v7/internal/handlers PDFService

// Upload-kind selectors for the body-limit policy. Untyped string constants
// (rather than a named type) so mocks/mock_services.go can implement
// ReadUpload/UploadLimit without importing this package (import cycle).
const (
	// UploadKindPDF caps single PDF uploads (mirrors compress.MaxInputBytes).
	UploadKindPDF = "pdf"
	// UploadKindXFDF caps XFDF form-data uploads.
	UploadKindXFDF = "xfdf"
	// UploadKindFont caps custom font uploads.
	UploadKindFont = "font"
)

// PDFService abstracts PDF operations used by HTTP handlers (mockable in unit tests).
// Beyond the engine ops it owns request policy: body-limit caps
// (UploadLimit/ReadUpload) and backend-error classification (ClassifyError),
// so handlers share one policy source instead of reimplementing 413/422
// semantics per endpoint.
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
	// UploadLimit returns the max accepted byte count for a kind of upload.
	UploadLimit(kind string) int64
	// ReadUpload reads r capped at the kind limit. ok=false means the input
	// exceeded the limit (caller should reject, e.g. with 413).
	ReadUpload(r io.Reader, kind string) (data []byte, ok bool, err error)
	// ClassifyError maps a backend failure to an HTTP status code.
	ClassifyError(err error) int
	// RedactPageInfo returns page dimensions for a redaction session.
	RedactPageInfo(pdfBytes []byte) (models.PageInfo, error)
	// RedactCapabilities returns per-page redaction capability info.
	RedactCapabilities(pdfBytes []byte) ([]models.PageCapability, error)
	// RedactTextPositions extracts positioned text for one page.
	RedactTextPositions(pdfBytes []byte, page int) ([]models.TextPosition, error)
	// RedactSearch finds candidate redaction rectangles for search terms.
	RedactSearch(pdfBytes []byte, terms []string) ([]models.RedactionRect, error)
	// RedactApply applies redaction options and returns the PDF plus report.
	RedactApply(pdfBytes []byte, opts models.ApplyRedactionOptions) ([]byte, models.RedactionApplyReport, error)
}

// FastGenerateService is the optional borrowed-render seam for the generate
// path. When the active PDFService implements it, handleGenerateTemplatePDF
// renders through the pooled buffer and streams without an extra copy;
// otherwise it falls back to GenerateTemplatePDF. Mocks implement it in
// mocks/mock_fast_generate.go so the hot path is test-covered.
type FastGenerateService interface {
	GenerateTemplatePDFBorrowed(template models.PDFTemplate) (*pdf.BorrowedPDF, error)
}

// ContextHTMLService is the cancellation-aware HTML seam. The legacy
// PDFService methods remain available for existing mocks and callers.
type ContextHTMLService interface {
	HTMLToPDFContext(context.Context, models.HTMLToPDFRequest) ([]byte, error)
	HTMLToImageContext(context.Context, models.HTMLToImageRequest) ([]byte, error)
}

type defaultPDFService struct{}

func (defaultPDFService) GenerateTemplatePDF(template models.PDFTemplate) ([]byte, error) {
	return pdf.GenerateTemplatePDF(template)
}

func (defaultPDFService) GenerateTemplatePDFBorrowed(template models.PDFTemplate) (*pdf.BorrowedPDF, error) {
	return pdf.GenerateTemplatePDFBorrowed(template)
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

func (defaultPDFService) HTMLToPDFContext(ctx context.Context, req models.HTMLToPDFRequest) ([]byte, error) {
	return pdf.ConvertHTMLToPDFContext(ctx, req)
}

func (defaultPDFService) HTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	return pdf.ConvertHTMLToImage(req)
}

func (defaultPDFService) HTMLToImageContext(ctx context.Context, req models.HTMLToImageRequest) ([]byte, error) {
	return pdf.ConvertHTMLToImageContext(ctx, req)
}

func (defaultPDFService) UploadLimit(kind string) int64 {
	return uploadLimitFor(kind)
}

func (defaultPDFService) ReadUpload(r io.Reader, kind string) ([]byte, bool, error) {
	return readBounded(r, uploadLimitFor(kind))
}

func (defaultPDFService) ClassifyError(err error) int {
	return pdfErrorStatus(err)
}

func (defaultPDFService) RedactPageInfo(pdfBytes []byte) (models.PageInfo, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return models.PageInfo{}, err
	}
	return r.GetPageInfo()
}

func (defaultPDFService) RedactCapabilities(pdfBytes []byte) ([]models.PageCapability, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.AnalyzePageCapabilities()
}

func (defaultPDFService) RedactTextPositions(pdfBytes []byte, page int) ([]models.TextPosition, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ExtractTextPositions(page)
}

func (defaultPDFService) RedactSearch(pdfBytes []byte, terms []string) ([]models.RedactionRect, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.FindTextOccurrencesMulti(terms)
}

func (defaultPDFService) RedactApply(pdfBytes []byte, opts models.ApplyRedactionOptions) ([]byte, models.RedactionApplyReport, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, models.RedactionApplyReport{}, err
	}
	return r.ApplyRedactionsAdvancedWithReport(opts)
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
