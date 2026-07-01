package redact

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
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
	redactionsByPage := make(map[int][]models.RedactionRect, 8)
	for _, rect := range redactions {
		redactionsByPage[rect.PageNum] = append(redactionsByPage[rect.PageNum], rect)
	}

	// Find highest object number
	maxObj := 0
	for k := range objMap {
		if n, ok := parseObjectKeyPrefix(k); ok && n > maxObj {
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
		sb.Grow(len(rects) * 48)
		sb.WriteString("q 0 0 0 rg ") // Save state, set black color
		for _, rect := range rects {
			appendPDFRect(&sb, rect.X, rect.Y, rect.Width, rect.Height)
		}
		sb.WriteString("Q ") // Restore state
		streamContent := sb.String()

		// Create new stream object
		streamObjKey := strconv.Itoa(nextObj) + " 0"
		nextObj++

		// NOTE: objMap stores body content between "obj" and "endobj" markers.
		// rebuildPDF wraps each body with "N G obj\n...\nendobj\n".
		var streamObj bytes.Buffer
		streamObj.Grow(len(streamContent) + 48)
		streamObj.WriteString("<< /Length ")
		streamObj.WriteString(strconv.Itoa(len(streamContent)))
		streamObj.WriteString(" >>\nstream\n")
		streamObj.WriteString(streamContent)
		streamObj.WriteString("\nendstream")
		objMap[streamObjKey] = bytes.Clone(streamObj.Bytes())

		// Append this new object to the page's /Contents
		newPageBody := appendStreamToPage(pageBody, streamObjKey)
		objMap[pageRef] = newPageBody
	}

	return rebuildPDF(objMap, r.pdfBytes)
}

// parseObjectKeyPrefix returns the object number from a "num gen" key.
func parseObjectKeyPrefix(key string) (int, bool) {
	id, _, ok := parseObjectKey(key)
	return id, ok
}