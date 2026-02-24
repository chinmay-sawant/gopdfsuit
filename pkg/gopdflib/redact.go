package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/redact"
)

type PageInfo = models.PageInfo
type PageDetail = models.PageDetail
type TextPosition = models.TextPosition
type RedactionRect = models.RedactionRect
type RedactionTextQuery = models.RedactionTextQuery
type ApplyRedactionOptions = models.ApplyRedactionOptions
type OCRSettings = models.OCRSettings
type PageCapability = models.PageCapability
type RedactionApplyReport = models.RedactionApplyReport

func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return PageInfo{}, err
	}
	return r.GetPageInfo()
}

func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ExtractTextPositions(pageNum)
}

func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.FindTextOccurrences(searchText)
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ApplyRedactions(redactions)
}

func ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ApplyRedactionsAdvanced(options)
}

func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, RedactionApplyReport{}, err
	}
	return r.ApplyRedactionsAdvancedWithReport(options)
}

func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.AnalyzePageCapabilities()
}
