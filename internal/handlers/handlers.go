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
	v1.POST("/generate/pdf", handleGeneratePDF)
}

func handleGeneratePDF(c *gin.Context) {
	var requestData models.InvoiceData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	pdf.GeneratePDF(c, requestData)
}
