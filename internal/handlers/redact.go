package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/byteconv"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/redact"
	"github.com/gin-gonic/gin"
)

// isAllWhitespace reports whether s is empty or only Unicode spaces (no alloc).
func isAllWhitespace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

func trimStringSpace(s string) string {
	start, end := 0, len(s)
	for start < end && isASCIISpace(s[start]) {
		start++
	}
	for end > start && isASCIISpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func parseCommaSeparatedTerms(raw string) []string {
	if raw == "" {
		return nil
	}
	seen := make(map[string]struct{}, 4)
	terms := make([]string, 0, 4)
	rest := raw
	for {
		var part string
		if i := strings.IndexByte(rest, ','); i >= 0 {
			part = rest[:i]
			rest = rest[i+1:]
		} else {
			part = rest
			rest = ""
		}
		term := trimStringSpace(part)
		if term != "" {
			key := strings.ToLower(term)
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				terms = append(terms, term)
			}
		}
		if rest == "" {
			break
		}
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
		if strings.IndexByte(q.Text, ',') < 0 {
			key := strings.ToLower(q.Text)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			normalized = append(normalized, models.RedactionTextQuery{Text: q.Text})
			continue
		}
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
	if len(pdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is empty"})
		return
	}

	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load PDF for info: " + err.Error()})
		return
	}
	info, err := r.GetPageInfo()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// HandleRedactCapabilities returns per-page capability information for redaction.
func HandleRedactCapabilities(c *gin.Context) {
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
	if len(pdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is empty"})
		return
	}

	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load PDF: " + err.Error()})
		return
	}
	caps, err := r.AnalyzePageCapabilities()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"capabilities": caps})
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
	if len(pdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is empty"})
		return
	}

	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load PDF: " + err.Error()})
		return
	}
	positions, err := r.ExtractTextPositions(pageNum)
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

	var options models.ApplyRedactionOptions
	if mode := c.PostForm("mode"); !isAllWhitespace(mode) {
		options.Mode = trimStringSpace(mode)
	}
	options.Password = c.PostForm("password")

	blocksJSON := c.PostForm("blocks")
	if blocksJSON != "" {
		if err := json.Unmarshal(byteconv.StringToBytes(blocksJSON), &options.Blocks); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid blocks json"})
			return
		}
	}

	textSearchJSON := c.PostForm("textSearch")
	if textSearchJSON != "" {
		textSearchBytes := byteconv.StringToBytes(textSearchJSON)
		if err := json.Unmarshal(textSearchBytes, &options.TextSearch); err != nil {
			var plain []string
			if err2 := json.Unmarshal(textSearchBytes, &plain); err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid textSearch json"})
				return
			}
			for _, text := range plain {
				if isAllWhitespace(text) {
					continue
				}
				text = trimStringSpace(text)
				options.TextSearch = append(options.TextSearch, models.RedactionTextQuery{Text: text})
			}
		}
	}

	ocrJSON := c.PostForm("ocr")
	if !isAllWhitespace(ocrJSON) {
		var ocr models.OCRSettings
		if err := json.Unmarshal(byteconv.StringToBytes(ocrJSON), &ocr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ocr json"})
			return
		}
		options.OCR = &ocr
	}

	// Backward compatibility: old frontend sends "redactions".
	if len(options.Blocks) == 0 {
		redactionsJSON := c.PostForm("redactions")
		if redactionsJSON != "" {
			if err := json.Unmarshal(byteconv.StringToBytes(redactionsJSON), &options.Blocks); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redactions json"})
				return
			}
		}
	}

	// Backward compatibility: allow plain text search field for one-shot apply.
	if len(options.TextSearch) == 0 {
		rawText := c.PostForm("text")
		if !isAllWhitespace(rawText) {
			searchText := trimStringSpace(rawText)
			terms := parseCommaSeparatedTerms(searchText)
			if len(terms) == 0 {
				terms = []string{searchText}
			}
			options.TextSearch = make([]models.RedactionTextQuery, 0, len(terms))
			for _, t := range terms {
				options.TextSearch = append(options.TextSearch, models.RedactionTextQuery{Text: t})
			}
		}
	}
	options.TextSearch = normalizeTextSearchQueries(options.TextSearch)

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
	if len(pdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is empty"})
		return
	}

	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load PDF: " + err.Error()})
		return
	}
	redactedPDF, report, err := r.ApplyRedactionsAdvancedWithReport(options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if b, err := json.Marshal(report); err == nil {
		c.Header("X-Redaction-Report", string(b))
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

	var terms []string
	textsJSON := c.PostForm("texts")
	if !isAllWhitespace(textsJSON) {
		textsJSON = trimStringSpace(textsJSON)
		if err := json.Unmarshal(byteconv.StringToBytes(textsJSON), &terms); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid texts json"})
			return
		}
	}
	if len(terms) == 0 {
		rawText := c.PostForm("text")
		if !isAllWhitespace(rawText) {
			searchText := trimStringSpace(rawText)
			terms = parseCommaSeparatedTerms(searchText)
			if len(terms) == 0 {
				terms = []string{searchText}
			}
		}
	}
	if len(terms) == 0 {
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
	if len(pdfBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdf file is empty"})
		return
	}

	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load PDF: " + err.Error()})
		return
	}
	rects, err := r.FindTextOccurrencesMulti(terms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rects)
}
