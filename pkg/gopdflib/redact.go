// WASM text path: the browser bindings goRedactSearch/goRedactApply expose
// GetPageInfo, ExtractTextPositions, FindTextOccurrences, ApplyRedactions,
// and ApplyRedactionsAdvancedWithReport over searchable text only. OCR stays
// disabled in WASM (no pdftoppm/tesseract subprocess in the browser); the
// WASM entrypoint rejects options with OCR.Enabled=true and callers must
// leave the OCR field unset. Image-only pages report through
// AnalyzePageCapabilities instead of being OCRed.
package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/redact"
)

// GetPageInfo extracts page information from a PDF for redaction planning.
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	const op = "gopdflib: GetPageInfo"
	if len(pdfBytes) == 0 {
		return PageInfo{}, invalidInputError(op, "needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return PageInfo{}, wrapEngineError(op, err)
	}
	info, err := r.GetPageInfo()
	if err != nil {
		return PageInfo{}, wrapEngineError(op, err)
	}
	pub, err := fromInternal[models.PageInfo, PageInfo](info)
	if err != nil {
		return PageInfo{}, fmt.Errorf("%w: %s response translation: %w", ErrInternal, op, err)
	}
	return pub, nil
}

// ExtractTextPositions retrieves all text chunks and their coordinates from a specific page.
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	const op = "gopdflib: ExtractTextPositions"
	if len(pdfBytes) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	positions, err := r.ExtractTextPositions(pageNum)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	out := make([]TextPosition, 0, len(positions))
	for _, p := range positions {
		pub, err := fromInternal[models.TextPosition, TextPosition](p)
		if err != nil {
			return nil, fmt.Errorf("%w: %s response translation: %w", ErrInternal, op, err)
		}
		out = append(out, pub)
	}
	return out, nil
}

// FindTextOccurrences searches for text and returns match rectangles for redaction.
func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	const op = "gopdflib: FindTextOccurrences"
	if len(pdfBytes) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	if searchText == "" {
		return nil, invalidInputError(op, "needs non-empty searchText")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	rects, err := r.FindTextOccurrences(searchText)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	out := make([]RedactionRect, 0, len(rects))
	for _, rc := range rects {
		pub, err := fromInternal[models.RedactionRect, RedactionRect](rc)
		if err != nil {
			return nil, fmt.Errorf("%w: %s response translation: %w", ErrInternal, op, err)
		}
		out = append(out, pub)
	}
	return out, nil
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	const op = "gopdflib: ApplyRedactions"
	if len(pdfBytes) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	internal := make([]models.RedactionRect, 0, len(redactions))
	for _, rc := range redactions {
		in, err := toInternal[RedactionRect, models.RedactionRect](rc)
		if err != nil {
			return nil, fmt.Errorf("%w: %s request translation: %w", ErrInvalidInput, op, err)
		}
		internal = append(internal, in)
	}
	out, err := r.ApplyRedactions(internal)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}

// ApplyRedactionsAdvanced applies redaction using advanced options (search, OCR fallback, etc).
// WASM text path only: leave OCR unset. OCR.Enabled=true requires the
// pdftoppm/tesseract subprocess pipeline, which cannot run in the browser;
// the WASM entrypoint rejects it with an invalid_input envelope.
func ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error) {
	const op = "gopdflib: ApplyRedactionsAdvanced"
	if len(pdfBytes) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	in, err := toInternal[ApplyRedactionOptions, models.ApplyRedactionOptions](options)
	if err != nil {
		return nil, fmt.Errorf("%w: %s request translation: %w", ErrInvalidInput, op, err)
	}
	out, err := r.ApplyRedactionsAdvanced(in)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}

// ApplyRedactionsAdvancedWithReport applies redaction and returns a detailed execution report.
// WASM text path only: leave OCR unset (see ApplyRedactionsAdvanced).
func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	const op = "gopdflib: ApplyRedactionsAdvancedWithReport"
	if len(pdfBytes) == 0 {
		return nil, RedactionApplyReport{}, invalidInputError(op, "needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, RedactionApplyReport{}, wrapEngineError(op, err)
	}
	in, err := toInternal[ApplyRedactionOptions, models.ApplyRedactionOptions](options)
	if err != nil {
		return nil, RedactionApplyReport{}, fmt.Errorf("%w: %s request translation: %w", ErrInvalidInput, op, err)
	}
	out, report, err := r.ApplyRedactionsAdvancedWithReport(in)
	if err != nil {
		return nil, RedactionApplyReport{}, wrapEngineError(op, err)
	}
	pub, err := fromInternal[models.RedactionApplyReport, RedactionApplyReport](report)
	if err != nil {
		return nil, RedactionApplyReport{}, fmt.Errorf("%w: %s response translation: %w", ErrInternal, op, err)
	}
	return out, pub, nil
}

// AnalyzePageCapabilities determines which pages have searchable text or require OCR.
func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	const op = "gopdflib: AnalyzePageCapabilities"
	if len(pdfBytes) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	r, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	caps, err := r.AnalyzePageCapabilities()
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	out := make([]PageCapability, 0, len(caps))
	for _, c := range caps {
		pub, err := fromInternal[models.PageCapability, PageCapability](c)
		if err != nil {
			return nil, fmt.Errorf("%w: %s response translation: %w", ErrInternal, op, err)
		}
		out = append(out, pub)
	}
	return out, nil
}
