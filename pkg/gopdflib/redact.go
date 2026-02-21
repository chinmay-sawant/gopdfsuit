package gopdflib

import "github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf"

// PageInfo aliases for public API
type PageInfo = pdf.PageInfo
type PageDetail = pdf.PageDetail
type TextPosition = pdf.TextPosition
type RedactionRect = pdf.RedactionRect
type RedactionTextQuery = pdf.RedactionTextQuery
type ApplyRedactionOptions = pdf.ApplyRedactionOptions
type OCRSettings = pdf.OCRSettings
type PageCapability = pdf.PageCapability
type RedactionApplyReport = pdf.RedactionApplyReport

// GetPageInfo extracts metadata about PDF pages (count, dimensions)
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	return pdf.GetPageInfo(pdfBytes)
}

// ExtractTextPositions extracts text coordinates from a specific page
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	return pdf.ExtractTextPositions(pdfBytes, pageNum)
}

// FindTextOccurrences searches for text and returns candidate redaction rectangles.
func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	return pdf.FindTextOccurrences(pdfBytes, searchText)
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	return pdf.ApplyRedactions(pdfBytes, redactions)
}

// ApplyRedactionsAdvanced applies a unified redaction request.
func ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error) {
	return pdf.ApplyRedactionsAdvanced(pdfBytes, options)
}

// ApplyRedactionsAdvancedWithReport applies a unified redaction request and returns an execution report.
func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	return pdf.ApplyRedactionsAdvancedWithReport(pdfBytes, options)
}

// AnalyzePageCapabilities classifies each page for text/image redaction capability.
func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	return pdf.AnalyzePageCapabilities(pdfBytes)
}
