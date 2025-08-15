package pdf

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/gin-gonic/gin"
)

// GeneratePDF writes a simple PDF to the gin context based on the invoice data.
func GeneratePDF(c *gin.Context, requestData models.InvoiceData) {
	var pdfBuffer bytes.Buffer
	xrefOffsets := make(map[int]int)

	pdfBuffer.WriteString("%PDF-1.7\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	pdfBuffer.WriteString("endobj\n")

	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	pdfBuffer.WriteString("endobj\n")

	xrefOffsets[3] = pdfBuffer.Len()
	pdfBuffer.WriteString("3 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	pdfBuffer.WriteString("endobj\n")

	var contentStream bytes.Buffer
	contentStream.WriteString("BT\n")
	contentStream.WriteString("/F1 24 Tf\n")
	contentStream.WriteString("72 750 Td\n")
	contentStream.WriteString(fmt.Sprintf("(Invoice for: %s)\n", requestData.CustomerName))
	contentStream.WriteString("Tj\n")
	contentStream.WriteString("ET\n")

	xrefOffsets[4] = pdfBuffer.Len()
	pdfBuffer.WriteString("4 0 obj\n")
	pdfBuffer.WriteString(fmt.Sprintf("<< /Length %d >>\n", contentStream.Len()))
	pdfBuffer.WriteString("stream\n")
	pdfBuffer.Write(contentStream.Bytes())
	pdfBuffer.WriteString("\nendstream\n")
	pdfBuffer.WriteString("endobj\n")

	xrefOffsets[5] = pdfBuffer.Len()
	pdfBuffer.WriteString("5 0 obj\n")
	pdfBuffer.WriteString("<< /Type /Font /Subtype /Type1 /Name /F1 /BaseFont /Helvetica >>\n")
	pdfBuffer.WriteString("endobj\n")

	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString("xref\n")
	pdfBuffer.WriteString("0 6\n")
	pdfBuffer.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", xrefOffsets[i]))
	}

	pdfBuffer.WriteString("trailer\n")
	pdfBuffer.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	filename := fmt.Sprintf("invoice-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}
