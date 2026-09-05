package redact

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// ApplyRedactions applies visual redaction rectangles to the PDF.
// Rectangles arrive in display space (see pageboxes.go) and are mapped
// into MediaBox user space before painting.
func (r *Redactor) ApplyRedactions(redactions []models.RedactionRect) ([]byte, error) {
	return r.applyRedactionsMedia(r.mapBlocksToMedia(redactions))
}

// applyRedactionsMedia paints rectangles already in MediaBox user space.
func (r *Redactor) applyRedactionsMedia(redactions []models.RedactionRect) ([]byte, error) {
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

	var sbuf [64]byte
	for pageNum, rects := range redactionsByPage {
		pageObjNum, err := findPageObject(objMap, r.pdfBytes, pageNum)
		if err != nil {
			return nil, fmt.Errorf("failed to find page %s: %w", string(strconv.AppendInt(sbuf[:0], int64(pageNum), 10)), err)
		}
		pageBody := objMap[pageObjNum]

		var sb strings.Builder
		// The appended stream inherits whatever CTM earlier streams left
		// behind (often a flipped/scaled matrix on printer-generated
		// files). Prefix its inverse so the boxes below paint in default
		// user space exactly as addressed.
		if inv, ok := netCTMInverse(objMap, pageBody); ok {
			sb.WriteString(inv)
			sb.WriteString(" cm ")
		}
		sb.WriteString("q 0 0 0 rg ")
		for _, rect := range rects {
			b := strconv.AppendFloat(sbuf[:0], rect.X, 'f', 2, 64)
			sb.Write(b)
			sb.WriteByte(' ')
			b = strconv.AppendFloat(sbuf[:0], rect.Y, 'f', 2, 64)
			sb.Write(b)
			sb.WriteByte(' ')
			b = strconv.AppendFloat(sbuf[:0], rect.Width, 'f', 2, 64)
			sb.Write(b)
			sb.WriteByte(' ')
			b = strconv.AppendFloat(sbuf[:0], rect.Height, 'f', 2, 64)
			sb.Write(b)
			sb.WriteString(" re f ")
		}
		sb.WriteString("Q ")
		streamContent := sb.String()

		streamGen := 0
		var streamBuf bytes.Buffer
		streamBuf.WriteString("<< /Length ")
		streamBuf.WriteString(string(strconv.AppendInt(sbuf[:0], int64(len(streamContent)), 10)))
		streamBuf.WriteString(" >>\nstream\n")
		streamBuf.WriteString(streamContent)
		streamBuf.WriteString("\nendstream")
		objMap[nextObj] = streamBuf.Bytes()
		objGen[nextObj] = streamGen

		newPageBody := appendStreamToPage(pageBody, nextObj, streamGen)
		objMap[pageObjNum] = newPageBody

		nextObj++
	}

	return rebuildPDF(objMap, objGen, r.pdfBytes)
}
