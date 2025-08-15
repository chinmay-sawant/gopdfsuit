package handlers

import (
	"net/http"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/internal/pdf"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up API routes onto the provided Gin router.
func RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	v1.POST("/generate/template-pdf", handleGenerateTemplatePDF)
}

func handleGenerateTemplatePDF(c *gin.Context) {
	var template models.PDFTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template data: " + err.Error()})
		return
	}
	pdf.GenerateTemplatePDF(c, template)
}
