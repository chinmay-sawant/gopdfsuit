package redact

import (
	"errors"
	"fmt"
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
	objGen := r.objGen
	if objMap == nil {
		var err error
		objMap, objGen, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, err
		}
	}

	redactionsByPage := make(map[int][]models.RedactionRect)
	for _, rect := range redactions {
		redactionsByPage[rect.PageNum] = append(redactionsByPage[rect.PageNum], rect)
	}

	maxObj := 0
	for n := range objMap {
		if n > maxObj {
			maxObj = n
		}
	}
	nextObj := maxObj + 1

	for pageNum, rects := range redactionsByPage {
		pageObjNum, err := findPageObject(objMap, r.pdfBytes, pageNum)
		if err != nil {
			return nil, fmt.Errorf("failed to find page %d: %w", pageNum, err)
		}
		pageBody := objMap[pageObjNum]

		var sb strings.Builder
		sb.WriteString("q 0 0 0 rg ")
		for _, rect := range rects {
			sb.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re f ", rect.X, rect.Y, rect.Width, rect.Height))
		}
		sb.WriteString("Q ")
		streamContent := sb.String()

		streamGen := 0
		streamObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(streamContent), streamContent)
		objMap[nextObj] = []byte(streamObj)
		objGen[nextObj] = streamGen

		newPageBody := appendStreamToPage(pageBody, nextObj, streamGen)
		objMap[pageObjNum] = newPageBody

		nextObj++
	}

	return rebuildPDF(objMap, objGen, r.pdfBytes)
}
