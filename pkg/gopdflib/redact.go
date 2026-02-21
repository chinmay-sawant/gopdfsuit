package gopdflib

import "github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/redact"

// PageInfo aliases for public API
type PageInfo = redact.PageInfo
type PageDetail = redact.PageDetail
type TextPosition = redact.TextPosition
type RedactionRect = redact.RedactionRect
type RedactionTextQuery = redact.RedactionTextQuery
type ApplyRedactionOptions = redact.ApplyRedactionOptions
type OCRSettings = redact.OCRSettings
type PageCapability = redact.PageCapability
type RedactionApplyReport = redact.RedactionApplyReport

// GetPageInfo extracts metadata about PDF pages (count, dimensions)
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	return redact.GetPageInfo(pdfBytes)
}

// ExtractTextPositions extracts text coordinates from a specific page
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	return redact.ExtractTextPositions(pdfBytes, pageNum)
}

// FindTextOccurrences searches for text and returns candidate redaction rectangles.
func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	return redact.FindTextOccurrences(pdfBytes, searchText)
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	return redact.ApplyRedactions(pdfBytes, redactions)
}

// ApplyRedactionsAdvanced applies a unified redaction request.
func ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error) {
	return redact.ApplyRedactionsAdvanced(pdfBytes, options)
}

// ApplyRedactionsAdvancedWithReport applies a unified redaction request and returns an execution report.
func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	return redact.ApplyRedactionsAdvancedWithReport(pdfBytes, options)
}

// AnalyzePageCapabilities classifies each page for text/image redaction capability.
func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	return redact.AnalyzePageCapabilities(pdfBytes)
}
