package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
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
	file, err := c.FormFile("pdf")
	if err != nil {
		abortError(c, http.StatusBadRequest, "pdf file is required")
		return nil
	}

	f, err := file.Open()
	if err != nil {
		log.Printf("redact: open pdf failed: %v", err)
		abortError(c, http.StatusInternalServerError, "failed to process upload")
		return nil
	}
	defer func() { _ = f.Close() }()

	pdfBytes, ok, err := pdfService.ReadUpload(f, UploadKindPDF)
	if err != nil {
		log.Printf("redact: read pdf failed: %v", err)
		abortError(c, http.StatusBadRequest, "invalid request")
		return nil
	}
	if !ok {
		abortError(c, http.StatusRequestEntityTooLarge, "pdf exceeds maximum size")
		return nil
	}
	if len(pdfBytes) == 0 {
		abortError(c, http.StatusBadRequest, "pdf file is empty")
		return nil
	}
	return pdfBytes
}

// abortRedactLoad handles redactor-construction failures (client-malformed
// PDFs -> 422) without echoing backend internals.
func abortRedactLoad(c *gin.Context, err error) {
	abortRedact(c, err, "load")
}

// abortRedactOp handles redaction operation failures the same way.
func abortRedactOp(c *gin.Context, err error) {
	abortRedact(c, err, "operation")
}

// abortRedact maps the classified status to a generic client message plus
// the envelope code, mirroring abortPDFError.
func abortRedact(c *gin.Context, err error, op string) {
	log.Printf("redact: %s failed: %v", op, err)
	status := pdfService.ClassifyError(err)
	var message string
	switch status {
	case http.StatusUnprocessableEntity:
		message = "invalid PDF input"
	case http.StatusRequestEntityTooLarge:
		message = "pdf exceeds maximum size"
	case http.StatusBadGateway:
		message = "upstream dependency failed"
	default:
		status = http.StatusInternalServerError
		message = "redaction failed"
	}
	c.JSON(status, errorBody(gopdflib.CodeForStatus(status), message))
}

// handleRedactPageInfo handles requests to get PDF page dimensions
func handleRedactPageInfo(c *gin.Context) {
	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	info, err := pdfService.RedactPageInfo(pdfBytes)
	if err != nil {
		abortRedactLoad(c, err)
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
		abortRedactLoad(c, err)
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
		abortRedactOp(c, err)
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

	var options models.ApplyRedactionOptions
	options.Mode = fastTrimSpace(c.PostForm("mode"))
	options.Password = c.PostForm("password")

	blocksJSON := c.PostForm("blocks")
	if blocksJSON != "" {
		if err := sonic.Unmarshal([]byte(blocksJSON), &options.Blocks); err != nil {
			abortError(c, http.StatusBadRequest, "invalid blocks json")
			return
		}
	}

	textSearchJSON := c.PostForm("textSearch")
	if textSearchJSON != "" {
		textSearchBytes := []byte(textSearchJSON)
		if err := sonic.Unmarshal(textSearchBytes, &options.TextSearch); err != nil {
			var plain []string
			if err2 := sonic.Unmarshal(textSearchBytes, &plain); err2 != nil {
				abortError(c, http.StatusBadRequest, "invalid textSearch json")
				return
			}
			for _, text := range plain {
				text = fastTrimSpace(text)
				if text == "" {
					continue
				}
				options.TextSearch = append(options.TextSearch, models.RedactionTextQuery{Text: text})
			}
		}
	}

	ocrJSON := c.PostForm("ocr")
	if fastTrimSpace(ocrJSON) != "" {
		var ocr models.OCRSettings
		if err := sonic.Unmarshal([]byte(ocrJSON), &ocr); err != nil {
			abortError(c, http.StatusBadRequest, "invalid ocr json")
			return
		}
		options.OCR = &ocr
	}

	// Backward compatibility: old frontend sends "redactions".
	if len(options.Blocks) == 0 {
		redactionsJSON := c.PostForm("redactions")
		if redactionsJSON != "" {
			if err := sonic.Unmarshal([]byte(redactionsJSON), &options.Blocks); err != nil {
				abortError(c, http.StatusBadRequest, "invalid redactions json")
				return
			}
		}
	}

	// Backward compatibility: allow plain text search field for one-shot apply.
	if len(options.TextSearch) == 0 {
		if searchText := fastTrimSpace(c.PostForm("text")); searchText != "" {
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

	pdfBytes := readBoundedUpload(c)
	if pdfBytes == nil {
		return
	}

	redactedPDF, report, err := pdfService.RedactApply(pdfBytes, options)
	if err != nil {
		abortRedactOp(c, err)
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
		abortRedactOp(c, err)
		return
	}

	c.JSON(http.StatusOK, rects)
}
