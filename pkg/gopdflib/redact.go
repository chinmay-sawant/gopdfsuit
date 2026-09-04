package gopdflib

import (
	"errors"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/redact"
)

// PageInfo holds information about a PDF's pages for redaction.
type PageInfo = models.PageInfo

// PageDetail holds dimensions and metadata for a single page.
type PageDetail = models.PageDetail

// TextPosition represents the coordinates and content of a text string on a page.
type TextPosition = models.TextPosition

// RedactionRect represents a rectangle to be redacted on a page.
type RedactionRect = models.RedactionRect

// RedactionTextQuery holds search parameters for text-based redaction.
type RedactionTextQuery = models.RedactionTextQuery

// ApplyRedactionOptions configures advanced redaction behavior.
type ApplyRedactionOptions = models.ApplyRedactionOptions

// OCRSettings configures OCR fallback for image-only pages.
type OCRSettings = models.OCRSettings

// PageCapability describes if a page can be redacted via text search or requires OCR.
type PageCapability = models.PageCapability

// RedactionApplyReport provides results and warnings from an advanced redaction operation.
type RedactionApplyReport = models.RedactionApplyReport

// GetPageInfo extracts page information from a PDF for redaction planning.
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	if len(pdfBytes) == 0 {
		return PageInfo{}, errors.New("gopdflib: GetPageInfo needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return PageInfo{}, err
	}
	return r.GetPageInfo()
}

// ExtractTextPositions retrieves all text chunks and their coordinates from a specific page.
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: ExtractTextPositions needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ExtractTextPositions(pageNum)
}

// FindTextOccurrences searches for text and returns match rectangles for redaction.
func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: FindTextOccurrences needs a non-empty PDF")
	}
	if searchText == "" {
		return nil, errors.New("gopdflib: FindTextOccurrences needs non-empty searchText")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.FindTextOccurrences(searchText)
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: ApplyRedactions needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ApplyRedactions(redactions)
}

// ApplyRedactionsAdvanced applies redaction using advanced options (search, OCR fallback, etc).
func ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: ApplyRedactionsAdvanced needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ApplyRedactionsAdvanced(options)
}

// ApplyRedactionsAdvancedWithReport applies redaction and returns a detailed execution report.
func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	if len(pdfBytes) == 0 {
		return nil, RedactionApplyReport{}, errors.New("gopdflib: ApplyRedactionsAdvancedWithReport needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, RedactionApplyReport{}, err
	}
	return r.ApplyRedactionsAdvancedWithReport(options)
}

// AnalyzePageCapabilities determines which pages have searchable text or require OCR.
func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: AnalyzePageCapabilities needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.AnalyzePageCapabilities()
}
