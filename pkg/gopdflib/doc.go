// Package gopdflib provides a Go library for PDF generation, manipulation, and conversion.
//
// This package can be imported into your Go application to generate PDFs from templates,
// merge/split existing PDFs, fill PDF forms, and convert HTML to PDF/images.
//
// # Quick Start
//
// To generate a PDF from a template:
//
//	import "github.com/chinmay-sawant/gopdfsuit/pkg/gopdflib"
//
//	template := gopdflib.PDFTemplate{
//	    Config: gopdflib.Config{
//	        Page:          "A4",
//	        PageAlignment: 1, // Portrait
//	    },
//	    Title: gopdflib.Title{
//	        Props: "Helvetica:18:100:center:0:0:0:0",
//	        Text:  "My Document Title",
//	    },
//	    Elements: []gopdflib.Element{
//	        {Type: "table", Table: &gopdflib.Table{...}},
//	        {Type: "spacer", Spacer: &gopdflib.Spacer{Height: 20}},
//	    },
//	}
//
//	pdfBytes, err := gopdflib.GeneratePDF(template)
//
// # Template Structure
//
// The PDFTemplate structure supports:
//   - Multiple page sizes (A4, Letter, Legal, etc.)
//   - Portrait and landscape orientation
//   - Tables with customizable fonts, colors, and borders
//   - Images (embedded as base64)
//   - Spacers for vertical spacing
//   - Document title with optional embedded table
//   - Footer with custom text
//   - Bookmarks/outlines for navigation
//   - Digital signatures
//   - PDF/A compliance
//   - Password protection and encryption
//   - Custom font embedding (TTF/OTF)
//
// # Props String Format
//
// Cell and text properties are defined using a props string with format:
//
//	"FontName:FontSize:StyleCode:Alignment:LeftBorder:RightBorder:TopBorder:BottomBorder"
//
// Example: "Helvetica:12:100:left:1:1:1:1"
//   - FontName: Helvetica, Times-Roman, Courier, or custom font name
//   - FontSize: Size in points (e.g., 12)
//   - StyleCode: 3-digit code for bold(1/0), italic(1/0), underline(1/0)
//   - Alignment: left, center, right
//   - Borders: 1 = visible, 0 = hidden
//
// # Thread Safety
//
// The font registry uses a global singleton. When generating PDFs concurrently,
// the library calls ResetUsage() internally for each generation. For true concurrent
// PDF generation, consider using external synchronization or generate PDFs sequentially.
//
// # Features
//
//   - [GeneratePDF] - Generate PDF from template
//   - [MergePDFs] - Combine multiple PDFs
//   - [SplitPDF] - Split PDF into parts
//   - [FillPDFWithXFDF] - Fill PDF forms
//   - [ConvertHTMLToPDF] - HTML to PDF conversion
//   - [ConvertHTMLToImage] - HTML to image conversion
package gopdflib
