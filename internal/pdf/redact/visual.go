package redact

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/byteconv"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

// appendFloat2 writes f with 2 decimal places into sb (PERF-6/15: AppendFloat).
func appendFloat2(sb *strings.Builder, f float64) {
	var buf [32]byte
	sb.Write(strconv.AppendFloat(buf[:0], f, 'f', 2, 64))
}

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

	redactionsByPage := make(map[int][]models.RedactionRect, len(redactions)) // PERF-192
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
		sb.Grow(16 + len(rects)*48)
		sb.WriteString("q 0 0 0 rg ")
		for _, rect := range rects {
			// PERF-6: avoid fmt.Sprintf in loop
			appendFloat2(&sb, rect.X)
			sb.WriteByte(' ')
			appendFloat2(&sb, rect.Y)
			sb.WriteByte(' ')
			appendFloat2(&sb, rect.Width)
			sb.WriteByte(' ')
			appendFloat2(&sb, rect.Height)
			sb.WriteString(" re f ")
		}
		sb.WriteString("Q ")
		streamContent := sb.String()

		streamGen := 0
		// PERF-6: build stream object without fmt
		var objSB strings.Builder
		objSB.Grow(48 + len(streamContent))
		var lenTmp [20]byte
		objSB.WriteString("<< /Length ")
		objSB.Write(strconv.AppendInt(lenTmp[:0], int64(len(streamContent)), 10))
		objSB.WriteString(" >>\nstream\n")
		objSB.WriteString(streamContent)
		objSB.WriteString("\nendstream")
		objMap[nextObj] = byteconv.StringToBytes(objSB.String())
		objGen[nextObj] = streamGen

		newPageBody := appendStreamToPage(pageBody, nextObj, streamGen)
		objMap[pageObjNum] = newPageBody

		nextObj++
	}

	return rebuildPDF(objMap, objGen, r.pdfBytes)
}
