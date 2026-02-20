package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf"
	"github.com/gin-gonic/gin"
)

// HandleRedactPageInfo handles requests to get PDF page dimensions
func HandleRedactPageInfo(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open pdf file"})
		return
	}
	defer func() { _ = f.Close() }()

	pdfBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read pdf file"})
		return
	}

	info, err := pdf.GetPageInfo(pdfBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// HandleRedactTextPositions handles requests to extract text positions from a page
func HandleRedactTextPositions(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is required"})
		return
	}

	pageNumStr := c.PostForm("page")
	if pageNumStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page number is required"})
		return
	}
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open pdf file"})
		return
	}
	defer func() { _ = f.Close() }()

	pdfBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read pdf file"})
		return
	}

	positions, err := pdf.ExtractTextPositions(pdfBytes, pageNum)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// HandleRedactApply handles requests to apply redactions to a PDF
func HandleRedactApply(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is required"})
		return
	}

	redactionsJSON := c.PostForm("redactions")
	var redactions []pdf.RedactionRect
	if redactionsJSON != "" {
		if err := json.Unmarshal([]byte(redactionsJSON), &redactions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redactions json"})
			return
		}
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open pdf file"})
		return
	}
	defer func() { _ = f.Close() }()

	pdfBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read pdf file"})
		return
	}

	redactedPDF, err := pdf.ApplyRedactions(pdfBytes, redactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=redacted.pdf")
	c.Data(http.StatusOK, "application/pdf", redactedPDF)
}

// HandleRedactSearch searches for text and returns potential redaction rectangles
func HandleRedactSearch(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is required"})
		return
	}

	searchText := c.PostForm("text")
	if searchText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search text is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open pdf file"})
		return
	}
	defer func() { _ = f.Close() }()

	pdfBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read pdf file"})
		return
	}

	rects, err := pdf.FindTextOccurrences(pdfBytes, searchText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rects)
}
