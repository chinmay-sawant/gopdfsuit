package gopdflib

import (
	"errors"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/redact"
)

// GetPageInfo extracts page information from a PDF for redaction planning.
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	if len(pdfBytes) == 0 {
		return PageInfo{}, errors.New("gopdflib: GetPageInfo needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return PageInfo{}, err
	}
	info, err := r.GetPageInfo()
	if err != nil {
		return PageInfo{}, err
	}
	return mustFromInternal[models.PageInfo, PageInfo](info), nil
}

// ExtractTextPositions retrieves all text chunks and their coordinates from a specific page.
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: ExtractTextPositions needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	positions, err := r.ExtractTextPositions(pageNum)
	if err != nil {
		return nil, err
	}
	out := make([]TextPosition, 0, len(positions))
	for _, p := range positions {
		out = append(out, mustFromInternal[models.TextPosition, TextPosition](p))
	}
	return out, nil
}

// FindTextOccurrences searches for text and returns match rectangles for redaction.
func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: FindTextOccurrences needs a non-empty PDF")
	}
	if searchText == "" {
		return nil, errors.New("gopdflib: FindTextOccurrences needs non-empty searchText")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	rects, err := r.FindTextOccurrences(searchText)
	if err != nil {
		return nil, err
	}
	out := make([]RedactionRect, 0, len(rects))
	for _, rc := range rects {
		out = append(out, mustFromInternal[models.RedactionRect, RedactionRect](rc))
	}
	return out, nil
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: ApplyRedactions needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	internal := make([]models.RedactionRect, 0, len(redactions))
	for _, rc := range redactions {
		internal = append(internal, mustToInternal[RedactionRect, models.RedactionRect](rc))
	}
	return r.ApplyRedactions(internal)
}

// ApplyRedactionsAdvanced applies redaction using advanced options (search, OCR fallback, etc).
func ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: ApplyRedactionsAdvanced needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	return r.ApplyRedactionsAdvanced(mustToInternal[ApplyRedactionOptions, models.ApplyRedactionOptions](options))
}

// ApplyRedactionsAdvancedWithReport applies redaction and returns a detailed execution report.
func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	if len(pdfBytes) == 0 {
		return nil, RedactionApplyReport{}, errors.New("gopdflib: ApplyRedactionsAdvancedWithReport needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, RedactionApplyReport{}, err
	}
	out, report, err := r.ApplyRedactionsAdvancedWithReport(mustToInternal[ApplyRedactionOptions, models.ApplyRedactionOptions](options))
	if err != nil {
		return nil, RedactionApplyReport{}, err
	}
	return out, mustFromInternal[models.RedactionApplyReport, RedactionApplyReport](report), nil
}

// AnalyzePageCapabilities determines which pages have searchable text or require OCR.
func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: AnalyzePageCapabilities needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	caps, err := r.AnalyzePageCapabilities()
	if err != nil {
		return nil, err
	}
	out := make([]PageCapability, 0, len(caps))
	for _, c := range caps {
		out = append(out, mustFromInternal[models.PageCapability, PageCapability](c))
	}
	return out, nil
}
