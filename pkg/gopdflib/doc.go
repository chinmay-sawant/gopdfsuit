// Package gopdflib provides a Go library for PDF generation, manipulation, and conversion.
//
// This package can be imported into your Go application to generate PDFs from templates,
// merge/split/compress existing PDFs, fill PDF forms, and convert HTML to PDF/images.
//
// # Quick Start
//
// To generate a PDF from a template:
//
//	import "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
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
//   - StyleCode: 3-digit code for bold(1/0), italic(1/0), underline(1/0).
//     000 is regular, 100 is bold. This field is font style only, not color.
//     Color lives in bgcolor/textcolor hex fields on the cell, title, or table.
//   - Alignment: left, center, right
//   - Borders: 1 = visible, 0 = hidden
//
// # Builder overlay
//
// Builders are a thin overlay over PDFTemplate. They emit the same Props
// grammar and bgcolor/textcolor hex strings documented above, with no engine
// draw changes. The sink stays [GeneratePDF] / GeneratePDFBorrowed.
//
// Raw Props strings are the supported low-level form (handy when copying
// JSON fixtures verbatim). The fluent Font/Text chains below are the
// preferred spelling; both emit identical strings.
//
// Right-cell color example:
//
//	b := gopdflib.NewDocument("A4", true)
//	b.AddTitle("Document Title", gopdflib.WithTitleFont("Helvetica", 18, true))
//	tb := b.AddTable(3, 1, 2, 1)
//	row := tb.AddRow(
//	    gopdflib.Font("Helvetica").Size(12).Bordered().Cell(""),
//	    gopdflib.Font("Helvetica").Size(18).Bold().Center().Bordered().Cell("Document Title"),
//	    gopdflib.Font("Helvetica").Size(12).Right().Bordered().Cell(""),
//	)
//	gopdflib.SetCellTextColor(&row[2], "#B00020")
//	gopdflib.SetCellFont(&row[1], "Helvetica", 18, true, false, false)
//	pdfBytes, err := b.Generate()
//
// Bracket text example:
//
//	c := gopdflib.Font("Helvetica").Size(12).Bordered().Cell("clause")
//	gopdflib.AddBracketText(&c, "[", "]")
//
// # Thread Safety
//
// The font registry is a mutex-guarded global singleton. Each GeneratePDF
// call clones the registry for the generation, so concurrent GeneratePDF
// calls sharing fonts are safe. Registering or clearing fonts mutates the
// global registry: do that before spawning concurrent generations, or guard
// it with external synchronization.
//
// Borrowed buffers: GeneratePDFBorrowed hands out a pooled buffer. Call
// Release exactly once via defer; Bytes() borrows the buffer (invalid after
// Release), CopyBytes() copies it for longer retention:
//
//	doc, err := gopdflib.GeneratePDFBorrowed(template)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer doc.Release()
//	out := doc.CopyBytes()
//
// # Validation ownership
//
// gopdflib is the single validating interface for PDF operations. Every
// public function rejects obviously-invalid input at the boundary (empty
// PDFs, empty XFDF, empty search text, missing HTML/URL) before reaching the
// engine, and applies shared defaults (see ParseCompressLevel and
// normalizeCompressOptions). Entry points above this package - the CGO
// exports in bindings/python/cgo, the HTTP handlers, the WASM shim - enforce
// transport-only guards (nil pointers, insane lengths, malformed JSON) and
// delegate all semantic validation here. Do not duplicate content checks in
// those layers; fix them here so Go, Python, HTTP, and WASM share one policy.
//
// # Features
//
//   - [GeneratePDF] - Generate PDF from template
//   - [MergePDFs] - Combine multiple PDFs
//   - [SplitPDF] - Split PDF into parts
//   - [CompressPDF] - Compress an existing PDF
//   - [FillPDFWithXFDF] - Fill PDF forms
//   - [ConvertHTMLToPDF] - HTML to PDF conversion
//   - [ConvertHTMLToImage] - HTML to image conversion
package gopdflib
