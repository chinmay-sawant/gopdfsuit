package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock data structure for an invoice.
// In a real application, this would be more complex.
type InvoiceData struct {
	CustomerName string  `json:"customer_name"`
	Items        []Item  `json:"items"`
	Total        float64 `json:"total"`
}

type Item struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func main() {
	// 1. Initialize Gin router
	router := gin.Default()

	// 2. Define API endpoints
	v1 := router.Group("/api/v1")
	{
		v1.POST("/generate/pdf", handleGeneratePDF)
	}

	// 3. Start the server
	router.Run()
}

// handleGeneratePDF is the handler function for our PDF generation endpoint.
// It now contains a raw PDF generation engine.
func handleGeneratePDF(c *gin.Context) {
	// --- Input Validation ---
	var requestData InvoiceData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// --- Raw PDF Generation Engine ---
	// We use a bytes.Buffer to build the PDF in memory. This is highly efficient.
	var pdfBuffer bytes.Buffer

	// Keep track of the byte offset of each object for the cross-reference table.
	xrefOffsets := make(map[int]int)

	// 1. PDF Header
	// Specifies the PDF version.
	pdfBuffer.WriteString("%PDF-1.7\n")
	// Add a comment with binary characters to ensure the file is treated as binary.
	pdfBuffer.WriteString("%âãÏÓ\n")

	// --- PDF Body: Objects ---
	// A PDF is a collection of numbered objects.

	// Object 1: The Catalog (The root of the document's object hierarchy)
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	pdfBuffer.WriteString("endobj\n")

	// Object 2: The Pages Collection
	// This object is a container for all the page objects.
	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	pdfBuffer.WriteString("endobj\n")

	// Object 3: The Page Object
	// Defines a single page, its dimensions, and its resources.
	// MediaBox [0 0 595 842] defines an A4 page size in points (1/72 inch).
	xrefOffsets[3] = pdfBuffer.Len()
	pdfBuffer.WriteString("3 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	pdfBuffer.WriteString("endobj\n")

	// Object 4: The Content Stream
	// This object contains the actual commands to draw content on the page.
	// We will build the stream content first, then write the object.
	var contentStream bytes.Buffer
	contentStream.WriteString("BT\n")                                                       // Begin Text
	contentStream.WriteString("/F1 24 Tf\n")                                                // Set Font to F1, size 24
	contentStream.WriteString("72 750 Td\n")                                                // Position the text cursor (x, y)
	contentStream.WriteString(fmt.Sprintf("(Invoice for: %s)\n", requestData.CustomerName)) // Show the text
	contentStream.WriteString("Tj\n")                                                       // Tj operator shows the string
	contentStream.WriteString("ET\n")                                                       // End Text

	xrefOffsets[4] = pdfBuffer.Len()
	pdfBuffer.WriteString("4 0 obj\n")
	pdfBuffer.WriteString(fmt.Sprintf("<< /Length %d >>\n", contentStream.Len()))
	pdfBuffer.WriteString("stream\n")
	pdfBuffer.Write(contentStream.Bytes())
	pdfBuffer.WriteString("\nendstream\n")
	pdfBuffer.WriteString("endobj\n")

	// Object 5: The Font Object
	// We use one of the 14 standard PDF fonts to avoid embedding a font file.
	xrefOffsets[5] = pdfBuffer.Len()
	pdfBuffer.WriteString("5 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Font /Subtype /Type1 /Name /F1 /BaseFont /Helvetica >>\n")
	pdfBuffer.WriteString("endobj\n")

	// --- PDF Cross-Reference Table (xref) ---
	// This is an index of all objects, allowing for fast random access.
	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString("xref\n")
	// The table starts with object 0, and we have 5 objects in total (plus the mandatory 0 object).
	pdfBuffer.WriteString("0 6\n")
	// Object 0 is a special case, always present and defined this way.
	pdfBuffer.WriteString("0000000000 65535 f \n")
	// Write the offsets for our objects. The format is: 10-digit offset, space, 5-digit generation, space, 'n' (for in-use).
	for i := 1; i <= 5; i++ {
		pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", xrefOffsets[i]))
	}

	// --- PDF Trailer ---
	// The trailer provides the location of the xref table and the root object (Catalog).
	pdfBuffer.WriteString("trailer\n")
	pdfBuffer.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")

	// --- End of File Marker ---
	pdfBuffer.WriteString("%%EOF\n")

	// --- HTTP Response ---
	// Set the headers to tell the browser this is a PDF file.
	filename := fmt.Sprintf("invoice-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)

	// Write the generated PDF from our buffer to the HTTP response.
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}
