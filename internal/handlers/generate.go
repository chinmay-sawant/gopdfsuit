package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleGenerateTemplatePDF(c *gin.Context) {
	if !applyBodyLimit(c, requestBodyLimit(c, maxTemplateJSONBody)) {
		return
	}
	template := acquireTemplate()
	defer releaseTemplate(template)

	tier := c.GetHeader("X-Payload-Tier")
	if cl := c.Request.ContentLength; cl > 0 || tier != "" {
		template.PreallocForDecode(int(cl), tier)
	}

	if err := decodeTemplate(c.Request.Body, int(c.Request.ContentLength), tier, template); err != nil {
		if isBodyTooLargeErr(err) {
			abortError(c, http.StatusRequestEntityTooLarge, "template too large")
			return
		}
		log.Printf("handleGenerateTemplatePDF: invalid template: %v", err)
		abortError(c, http.StatusBadRequest, "invalid template data")
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=generated.pdf")

	// Borrowed-render seam: services that expose the pooled fast path
	// stream it directly; mocks implement the seam so this branch is
	// test-covered, and any service without it falls back to the copy.
	if fast, ok := pdfService.(FastGenerateService); ok {
		doc, err := fast.GenerateTemplatePDFBorrowed(*template)
		if err != nil {
			log.Printf("handleGenerateTemplatePDF: generation failed: %v", err)
			abortError(c, http.StatusInternalServerError, "PDF generation failed")
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
		abortError(c, http.StatusInternalServerError, "PDF generation failed")
		return
	}
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}
