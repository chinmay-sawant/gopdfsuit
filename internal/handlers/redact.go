package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/gin-gonic/gin"
)

func fastTrimSpace(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] > ' ' && s[len(s)-1] > ' ' {
		return s
	}
	return strings.TrimSpace(s)
}

func parseCommaSeparatedTerms(raw string) []string {
	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(parts))
	terms := make([]string, 0, len(parts))
	for _, p := range parts {
		term := fastTrimSpace(p)
		if term == "" {
			continue
		}
		key := strings.ToLower(term)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}

func normalizeTextSearchQueries(queries []models.RedactionTextQuery) []models.RedactionTextQuery {
	if len(queries) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(queries))
	normalized := make([]models.RedactionTextQuery, 0, len(queries))
	for _, q := range queries {
		for _, term := range parseCommaSeparatedTerms(q.Text) {
			key := strings.ToLower(term)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			normalized = append(normalized, models.RedactionTextQuery{Text: term})
		}
	}
	return normalized
}

// readBoundedUpload opens the "pdf" form file and reads it through the
// service-owned body-limit policy, rejecting oversized uploads with 413. It
// returns nil when the handler must abort (response already written).
func readBoundedUpload(c *gin.Context) []byte {
	return readSingleUpload(c, "pdf", UploadKindPDF)
}

// handleRedactPageInfo handles requests to get PDF page dimensions
func handleRedactPageInfo(c *gin.Context) {
	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	info, err := pdfService.RedactPageInfo(pdfBytes)
	if err != nil {
		abortPDFError(c, err, "redaction failed")
		return
	}

	c.JSON(http.StatusOK, info)
}

// handleRedactCapabilities returns per-page capability information for redaction.
func handleRedactCapabilities(c *gin.Context) {
	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	caps, err := pdfService.RedactCapabilities(pdfBytes)
	if err != nil {
		abortPDFError(c, err, "redaction failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"capabilities": caps})
}

// handleRedactTextPositions handles requests to extract text positions from a page
func handleRedactTextPositions(c *gin.Context) {
	pageNumStr := c.PostForm("page")
	if pageNumStr == "" {
		abortError(c, http.StatusBadRequest, "page number is required")
		return
	}
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		abortError(c, http.StatusBadRequest, "invalid page number")
		return
	}

	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	positions, err := pdfService.RedactTextPositions(pdfBytes, pageNum)
	if err != nil {
		abortPDFError(c, err, "redaction failed")
		return
	}

	c.JSON(http.StatusOK, positions)
}

// handleRedactApply handles requests to apply redactions to a PDF
func handleRedactApply(c *gin.Context) {
	if _, err := c.FormFile("pdf"); err != nil {
		abortError(c, http.StatusBadRequest, "pdf file is required")
		return
	}

	options, err := parseRedactApply(redactApplyForm{
		mode:       c.PostForm("mode"),
		password:   c.PostForm("password"),
		blocks:     c.PostForm("blocks"),
		textSearch: c.PostForm("textSearch"),
		ocr:        c.PostForm("ocr"),
		redactions: c.PostForm("redactions"),
		text:       c.PostForm("text"),
	})
	if err != nil {
		abortError(c, http.StatusBadRequest, err.Error())
		return
	}

	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	redactedPDF, report, err := pdfService.RedactApply(pdfBytes, options)
	if err != nil {
		abortPDFError(c, err, "redaction failed")
		return
	}
	if b, err := sonic.Marshal(report); err == nil {
		c.Header("X-Redaction-Report", string(b))
	}

	c.Header("Content-Disposition", "attachment; filename=redacted.pdf")
	c.Data(http.StatusOK, mimeTypePDF, redactedPDF)
}

// handleRedactSearch searches for text and returns potential redaction rectangles
func handleRedactSearch(c *gin.Context) {
	if _, err := c.FormFile("pdf"); err != nil {
		abortError(c, http.StatusBadRequest, "pdf file is required")
		return
	}

	var terms []string
	textsJSON := fastTrimSpace(c.PostForm("texts"))
	if textsJSON != "" {
		if err := sonic.Unmarshal([]byte(textsJSON), &terms); err != nil {
			abortError(c, http.StatusBadRequest, "invalid texts json")
			return
		}
	}
	if len(terms) == 0 {
		searchText := fastTrimSpace(c.PostForm("text"))
		if searchText != "" {
			terms = parseCommaSeparatedTerms(searchText)
			if len(terms) == 0 {
				terms = []string{searchText}
			}
		}
	}
	if len(terms) == 0 {
		abortError(c, http.StatusBadRequest, "search text is required")
		return
	}

	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	rects, err := pdfService.RedactSearch(pdfBytes, terms)
	if err != nil {
		abortPDFError(c, err, "redaction failed")
		return
	}

	c.JSON(http.StatusOK, rects)
}
