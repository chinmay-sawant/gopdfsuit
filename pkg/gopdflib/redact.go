package gopdflib

import "github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf"

// PageInfo aliases for public API
type PageInfo = pdf.PageInfo
type PageDetail = pdf.PageDetail
type TextPosition = pdf.TextPosition
type RedactionRect = pdf.RedactionRect

// GetPageInfo extracts metadata about PDF pages (count, dimensions)
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	return pdf.GetPageInfo(pdfBytes)
}

// ExtractTextPositions extracts text coordinates from a specific page
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	return pdf.ExtractTextPositions(pdfBytes, pageNum)
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	return pdf.ApplyRedactions(pdfBytes, redactions)
}
