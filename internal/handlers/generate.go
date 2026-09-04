package handlers

import (
	"log"
	"net/http"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
	"github.com/gin-gonic/gin"
)

func handleGenerateTemplatePDF(c *gin.Context) {
	template := acquireTemplate()
	defer releaseTemplate(template)

	tier := c.GetHeader("X-Payload-Tier")
	if cl := c.Request.ContentLength; cl > 0 || tier != "" {
		template.PreallocForDecode(int(cl), tier)
	}

	// Bound the JSON body before streaming decode.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTemplateJSONBody)
	if err := decodeTemplate(c.Request.Body, int(c.Request.ContentLength), tier, template); err != nil {
		if isBodyTooLargeErr(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "template too large"})
			return
		}
		log.Printf("handleGenerateTemplatePDF: invalid template: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template data"})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=generated.pdf")

	if _, ok := pdfService.(defaultPDFService); ok {
		doc, err := pdf.GenerateTemplatePDFBorrowed(*template)
		if err != nil {
			log.Printf("handleGenerateTemplatePDF: generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
			return
		}
		defer doc.Release()
		c.Status(http.StatusOK)
		if _, err := c.Writer.Write(doc.Bytes()); err != nil {
			return
		}
		return
	}

	pdfBytes, err := pdfService.GenerateTemplatePDF(*template)
	if err != nil {
		log.Printf("handleGenerateTemplatePDF: generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
		return
	}
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}
