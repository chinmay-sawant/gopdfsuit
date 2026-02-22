package redact

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// ApplyRedactions applies visual redaction rectangles to the PDF
func (r *Redactor) ApplyRedactions(redactions []models.RedactionRect) ([]byte, error) {
	if len(r.pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	if len(redactions) == 0 {
		return r.pdfBytes, nil
	}

	objMap := r.objMap
	if objMap == nil {
		var err error
		objMap, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, err
		}
	}

	// Group redactions by page
	redactionsByPage := make(map[int][]models.RedactionRect)
	for _, rect := range redactions {
		redactionsByPage[rect.PageNum] = append(redactionsByPage[rect.PageNum], rect)
	}

	// Find highest object number
	maxObj := 0
	for k := range objMap {
		var n int
		_, _ = fmt.Sscanf(k, "%d", &n)
		if n > maxObj {
			maxObj = n
		}
	}
	nextObj := maxObj + 1

	// For each page with redactions, append a new content stream
	for pageNum, rects := range redactionsByPage {
		pageRef, err := findPageObject(objMap, r.pdfBytes, pageNum)
		if err != nil {
			return nil, fmt.Errorf("failed to find page %d: %w", pageNum, err)
		}
		pageBody := objMap[pageRef]

		// Create redaction stream content
		var sb strings.Builder
		sb.WriteString("q 0 0 0 rg ") // Save state, set black color
		for _, rect := range rects {
			// Construct rectangle path: x y w h re f (fill)
			sb.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re f ", rect.X, rect.Y, rect.Width, rect.Height))
		}
		sb.WriteString("Q ") // Restore state
		streamContent := sb.String()

		// Create new stream object
		streamObjKey := fmt.Sprintf("%d 0", nextObj)
		nextObj++

		// NOTE: objMap stores body content between "obj" and "endobj" markers.
		// rebuildPDF wraps each body with "N G obj\n...\nendobj\n".
		streamObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(streamContent), streamContent)
		objMap[streamObjKey] = []byte(streamObj)

		// Append this new object to the page's /Contents
		newPageBody := appendStreamToPage(pageBody, streamObjKey)
		objMap[pageRef] = newPageBody
	}

	return rebuildPDF(objMap, r.pdfBytes)
}
