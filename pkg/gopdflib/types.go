// Package gopdflib provides public type aliases for PDF template generation.
// These types are re-exported from the internal models package for use by external consumers.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// PDFTemplate is the main input structure for PDF generation.
// It contains configuration, content elements, and optional features like bookmarks.
type PDFTemplate = models.PDFTemplate

// Config holds page configuration and optional features.
type Config = models.Config

// SecurityConfig holds PDF encryption and permission settings.
type SecurityConfig = models.SecurityConfig

// PDFAConfig holds PDF/A compliance settings.
type PDFAConfig = models.PDFAConfig

// SignatureConfig holds digital signature settings.
type SignatureConfig = models.SignatureConfig

// CustomFontConfig specifies a custom font to embed in the PDF.
type CustomFontConfig = models.CustomFontConfig

// Title represents the document title section.
type Title = models.Title

// TitleTable represents an embedded table within the title section.
type TitleTable = models.TitleTable

// Table represents a table element in the PDF.
type Table = models.Table

// Row represents a row in a table.
type Row = models.Row

// Cell represents a cell in a table row.
type Cell = models.Cell

// FormField represents a fillable form field.
type FormField = models.FormField

// Image represents an image element.
type Image = models.Image

// Footer represents the document footer.
type Footer = models.Footer

// Spacer represents vertical space between elements.
type Spacer = models.Spacer

// Element represents an ordered element in the PDF (table, spacer, or image).
type Element = models.Element

// Bookmark represents a PDF outline entry for document navigation.
type Bookmark = models.Bookmark

// Props holds font and style properties parsed from a props string.
type Props = models.Props

// FontInfo represents a font's information.
type FontInfo = models.FontInfo

// HtmlToPDFRequest represents the input for HTML to PDF conversion.
type HtmlToPDFRequest = models.HtmlToPDFRequest

// HtmlToImageRequest represents the input for HTML to image conversion.
type HtmlToImageRequest = models.HtmlToImageRequest
