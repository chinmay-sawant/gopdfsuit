package redact

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// Redactor provides an object-oriented interface for PDF redaction,
// caching the parsed PDF structure to avoid redundant processing.
type Redactor struct {
	pdfBytes []byte
	objMap   map[string][]byte
	info     *models.PageInfo
}

// NewRedactor initializes a new Redactor with the given PDF bytes.
// It parses the PDF into an object map and retrieves page information.
func NewRedactor(pdfBytes []byte) (*Redactor, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}

	if trailerHasEncrypt(pdfBytes) {
		return nil, errors.New("encrypted PDFs are not supported directly; use ApplyRedactionsAdvancedWithReport for password handling")
	}

	// Create an empty objMap first, some operations only need pdfBytes
	r := &Redactor{
		pdfBytes: pdfBytes,
	}

	// Try to build the map, but don't fail immediately if it's just meant for GetPageInfo error handling
	objMap, err := buildObjectMap(pdfBytes)
	if err == nil {
		r.objMap = objMap

		info, err := r.GetPageInfo()
		if err == nil {
			r.info = &info
		}
	}

	// We return the redactor even if object map building failed,
	// as functions like GetPageInfo handle their own errors and some use cases
	// (like encrypted PDFs with password provided later) might need it.
	return r, nil
}

// GetPageInfo extracts page count and dimensions from a PDF
func (r *Redactor) GetPageInfo() (models.PageInfo, error) {
	if len(r.pdfBytes) == 0 {
		return models.PageInfo{}, errors.New("empty pdf bytes")
	}

	if trailerHasEncrypt(r.pdfBytes) {
		return models.PageInfo{}, errors.New("encrypted PDFs are not supported")
	}

	rootRef, ok := findRootRef(r.pdfBytes)
	if !ok {
		return models.PageInfo{}, errors.New("could not find PDF root object")
	}

	// Use cached objMap if available, else build
	objMap := r.objMap
	var err error
	if objMap == nil {
		objMap, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return models.PageInfo{}, err
		}
	}

	rootBody, ok := objMap[rootRef]
	if !ok {
		return models.PageInfo{}, errors.New("root object not found in map")
	}

	pagesRefRe := regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
	pm := pagesRefRe.FindSubmatch(rootBody)
	if pm == nil {
		return models.PageInfo{}, errors.New("no /Pages reference in Root")
	}
	pagesKey := string(pm[1]) + " " + string(pm[2])

	var pageDims []models.PageDetail
	if err := traversePages(pagesKey, objMap, &pageDims); err != nil {
		return models.PageInfo{}, fmt.Errorf("error traversing page tree: %w", err)
	}

	return models.PageInfo{
		TotalPages: len(pageDims),
		Pages:      pageDims,
	}, nil
}

// AnalyzePageCapabilities classifies each page for text/image redaction capability.
func (r *Redactor) AnalyzePageCapabilities() ([]models.PageCapability, error) {
	objMap := r.objMap
	var err error
	if objMap == nil {
		objMap, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, err
		}
	}

	info := r.info
	if info == nil {
		inf, err := r.GetPageInfo()
		if err != nil {
			return nil, err
		}
		info = &inf
	}

	caps := make([]models.PageCapability, 0, info.TotalPages)
	for i := 1; i <= info.TotalPages; i++ {
		pageRef, err := findPageObject(objMap, r.pdfBytes, i)
		if err != nil {
			caps = append(caps, models.PageCapability{PageNum: i, Type: "unknown", Note: err.Error()})
			continue
		}
		body := objMap[pageRef]
		keys := extractContentKeys(body)
		hasText := false
		hasImage := false
		for _, key := range keys {
			objBody, ok := objMap[key]
			if !ok {
				continue
			}
			rawStream, decStream, _ := inspectStream(objBody)
			combined := append(rawStream, decStream...) //nolint:gocritic
			s := string(combined)
			if strings.Contains(s, "BT") && (strings.Contains(s, "Tj") || strings.Contains(s, "TJ")) {
				hasText = true
			}
			if strings.Contains(s, " Do") || bytesIndex(objBody, []byte("/Image")) >= 0 {
				hasImage = true
			}
		}
		if len(keys) == 0 {
			content, _ := extractPageContent(body, objMap)
			s := string(content)
			hasText = strings.Contains(s, "BT") && (strings.Contains(s, "Tj") || strings.Contains(s, "TJ"))
			hasImage = strings.Contains(s, " Do") || bytesIndex(body, []byte("/Image")) >= 0
		}
		typeName := "unknown"
		switch {
		case hasText && hasImage:
			typeName = "mixed"
		case hasText:
			typeName = "text"
		case hasImage:
			typeName = "image_only"
		}
		capability := models.PageCapability{PageNum: i, Type: typeName, HasText: hasText, HasImage: hasImage}
		if typeName == "image_only" {
			capability.Note = "text search requires OCR for image-only content"
		}
		caps = append(caps, capability)
	}
	return caps, nil
}

// ApplyRedactionsAdvanced applies a unified redaction request.
func (r *Redactor) ApplyRedactionsAdvanced(opts models.ApplyRedactionOptions) ([]byte, error) {
	out, _, err := r.ApplyRedactionsAdvancedWithReport(opts)
	return out, err
}

// ApplyRedactionsAdvancedWithReport applies redactions and returns an execution report.
func (r *Redactor) ApplyRedactionsAdvancedWithReport(opts models.ApplyRedactionOptions) ([]byte, models.RedactionApplyReport, error) {
	if len(r.pdfBytes) == 0 {
		return nil, models.RedactionApplyReport{}, errors.New("empty pdf bytes")
	}

	report := models.RedactionApplyReport{
		Mode:            "visual_allowed",
		SecurityOutcome: "visual_only",
	}

	mode := strings.TrimSpace(strings.ToLower(opts.Mode))
	if mode == "" {
		mode = "visual_allowed" //nolint:goconst
	}
	report.Mode = mode
	if mode != "visual_allowed" && mode != "secure_required" {
		return nil, report, errors.New("invalid mode: expected visual_allowed or secure_required")
	}

	// For encrypted PDFs, we might process them recursively depending on if decrypt succeeds
	workingPDF := r.pdfBytes
	if trailerHasEncrypt(workingPDF) {
		dec, err := decryptEncryptedPDFBytes(workingPDF, opts.Password)
		if err != nil {
			return nil, report, err
		}
		// Try to build object map for decrypted bytes if successful
		objMap, objErr := buildObjectMap(dec)
		if objErr != nil {
			return nil, report, objErr
		}
		// Temporarily replace redactor state with decrypted version for this operation
		r.pdfBytes = dec
		r.objMap = objMap
		r.info = nil // Reset cached info to force recalculation on decrypted bytes
		report.Warnings = append(report.Warnings, "input PDF was decrypted using in-house pipeline and output is emitted decrypted")
	}

	caps, capErr := r.AnalyzePageCapabilities()
	if capErr == nil {
		report.Capabilities = caps
	}

	if opts.OCR != nil && opts.OCR.Enabled {
		report.Warnings = append(report.Warnings, "OCR requested but no OCR provider is configured; this is an extension hook")
	}

	all := make([]models.RedactionRect, 0, len(opts.Blocks)+8)
	all = append(all, opts.Blocks...)
	activeTextQueries := opts.TextSearch

	for _, q := range activeTextQueries {
		query := strings.TrimSpace(q.Text)
		if query == "" {
			continue
		}
		rects, err := r.FindTextOccurrences(query)
		if err != nil {
			return nil, report, err
		}
		all = append(all, rects...)
		report.MatchedTextCount += len(rects)
	}

	if opts.OCR != nil && opts.OCR.Enabled {
		ocrRects, err := r.runOCRSearch(activeTextQueries, *opts.OCR)
		if err != nil {
			report.Warnings = append(report.Warnings, "OCR fallback error: "+err.Error())
		} else {
			all = append(all, ocrRects...)
			report.MatchedTextCount += len(ocrRects)
		}
	}
	report.GeneratedRects = len(all)

	// Keep visual mode strictly visual to avoid mutating encoded text streams.
	// Best-effort secure rewriting can produce glyph corruption for complex font encodings.
	if mode == "visual_allowed" {
		report.Warnings = append(report.Warnings, "visual_allowed mode skipped secure text rewrite to preserve original glyph encoding")
	}

	if mode == "secure_required" {
		secureOut, secureChanged, secureWarns, err := r.applySecureContentRedactions(all, activeTextQueries)
		report.Warnings = append(report.Warnings, secureWarns...)
		if err != nil {
			report.SecurityOutcome = "failed" //nolint:goconst
			return nil, report, err
		}
		if !secureChanged {
			report.SecurityOutcome = "failed"
			return nil, report, errors.New("secure_required requested but no secure text content could be removed")
		}
		// Note: ApplyRedactions visual overlays still use the original ApplyRedactions design
		// We can make applyRedactionsToPage a method later.
		rOut, _ := NewRedactor(secureOut)
		visualOut, err := rOut.ApplyRedactions(all)
		if err != nil {
			report.SecurityOutcome = "failed"
			return nil, report, err
		}
		report.AppliedSecure = true
		report.AppliedVisual = true
		report.SecurityOutcome = "secure"
		report.AppliedRectangles = len(all)
		return visualOut, report, nil
	}

	out, err := r.ApplyRedactions(all)
	if err != nil {
		report.SecurityOutcome = "failed"
		return nil, report, err
	}
	report.AppliedVisual = true
	report.AppliedRectangles = len(all)
	return out, report, nil
}
