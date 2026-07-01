package pdf

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	mergeObjRe    = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	mergeRefRe    = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	mergeStreamRe = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	mergePagesRe  = regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
	mergeKidsRe   = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	mergeParentRe = regexp.MustCompile(`/Parent\s+\d+\s+\d+\s+R`)
	mergeAcroRe   = regexp.MustCompile(`/AcroForm\s+(\d+)\s+\d+\s+R`)
	mergeFieldsRe = regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	mergeAnnotsRe = regexp.MustCompile(`/Annots\s*\[(.*?)\]`)
)

// MergePDFs merges multiple PDF byte slices into a single PDF by parsing objects,
// remapping object numbers and building a new /Pages tree that references all
// page objects from the inputs. This avoids an external dependency.
//
//nolint:gocyclo
func MergePDFs(files [][]byte) ([]byte, error) {
	header := []byte("%PDF-1.7\n%âãÏÓ\n")
	var out bytes.Buffer
	out.Write(header)

	offsets := map[int]int{}
	mergedPages := []int{}
	mergedFormFields := []int{}
	formFieldSet := make(map[int]bool, 16)
	currentMax := 2

	var appended []struct {
		num  int
		body []byte
	}

	tempObjMap := make(map[string][]byte, 32)

	for _, f := range files {
		if trailerHasEncrypt(f) {
			return nil, errors.New("cannot merge encrypted PDF")
		}

		objMatches := mergeObjRe.FindAllSubmatch(f, -1)
		if len(objMatches) == 0 {
			continue
		}

		objMap := make(map[int][]byte, len(objMatches))
		maxObj := 0
		for _, m := range objMatches {
			if n, err := strconv.Atoi(string(m[1])); err == nil {
				body := m[3]
				objMap[n] = body
				if n > maxObj {
					maxObj = n
				}
			}
		}

		for k := range tempObjMap {
			delete(tempObjMap, k)
		}
		parseXRefStreams(f, tempObjMap)
		for k, v := range tempObjMap {
			space := strings.IndexByte(k, ' ')
			if space <= 0 {
				continue
			}
			onum, err := strconv.Atoi(k[:space])
			if err == nil {
				if _, exists := objMap[onum]; !exists {
					objMap[onum] = v
					if onum > maxObj {
						maxObj = onum
					}
				}
			}
		}

		offset := currentMax
		pagesFromTree := []int{}
		if rootRef, ok := findRootRef(f); ok {
			parts := splitASCIIFields(rootRef)
			if len(parts) >= 2 {
				rootNum, err1 := strconv.Atoi(parts[0])
				if err1 == nil {
					if rootBody, ok2 := objMap[rootNum]; ok2 {
						if pm := mergePagesRe.FindSubmatch(rootBody); pm != nil {
							if pnum, err := strconv.Atoi(string(pm[1])); err == nil {
								if pagesBody, ok3 := objMap[pnum]; ok3 {
									if km := mergeKidsRe.FindSubmatch(pagesBody); km != nil {
										for _, r := range mergeRefRe.FindAllSubmatch(km[1], -1) {
											if pn, err := strconv.Atoi(string(r[1])); err == nil {
												pagesFromTree = append(pagesFromTree, pn)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		formFields := extractFormFieldsFromFile(f, objMap)

		var fileObjects []int
		for origNum := 1; origNum <= maxObj; origNum++ {
			if _, ok := objMap[origNum]; ok {
				fileObjects = append(fileObjects, origNum)
			}
		}

		for _, origNum := range fileObjects {
			body := objMap[origNum]
			newBody := replaceRefsOutsideStreams(body, mergeRefRe, offset)
			newNum := offset + origNum
			appended = append(appended, struct {
				num  int
				body []byte
			}{num: newNum, body: newBody})

			if len(pagesFromTree) == 0 {
				if bytesIndex(newBody, []byte("/Type /Page")) >= 0 || bytesIndex(newBody, []byte("/MediaBox")) >= 0 {
					mergedPages = append(mergedPages, newNum)
				}
			}
		}

		for _, pn := range pagesFromTree {
			mergedPages = append(mergedPages, offset+pn)
		}

		for _, fn := range formFields {
			remappedFieldNum := offset + fn
			if !formFieldSet[remappedFieldNum] {
				mergedFormFields = append(mergedFormFields, remappedFieldNum)
				formFieldSet[remappedFieldNum] = true
			}
		}

		currentMax = offset + maxObj + 1
	}

	offsets[1] = out.Len()
	var catalog strings.Builder
	catalog.WriteString("<< /Type /Catalog /Pages 2 0 R")
	if len(mergedFormFields) > 0 {
		catalog.WriteString(" /AcroForm << /Fields [")
		var fieldBuf []byte
		for i, fieldNum := range mergedFormFields {
			if i > 0 {
				catalog.WriteByte(' ')
			}
			fieldBuf = fieldBuf[:0]
			fieldBuf = strconv.AppendInt(fieldBuf, int64(fieldNum), 10)
			fieldBuf = append(fieldBuf, " 0 R"...)
			catalog.Write(fieldBuf)
		}
		catalog.WriteString("] >>")
	}
	catalog.WriteString(" >>")
	var objHdr []byte
	objHdr = strconv.AppendInt(objHdr, 1, 10)
	objHdr = append(objHdr, " 0 obj\n"...)
	out.Write(objHdr)
	out.WriteString(catalog.String())
	out.WriteString("\nendobj\n")

	offsets[2] = out.Len()
	var pagesHdr strings.Builder
	pagesHdr.WriteString("2 0 obj\n<< /Type /Pages /Kids [")
	var pageBuf []byte
	for i, p := range mergedPages {
		if i > 0 {
			pagesHdr.WriteByte(' ')
		}
		pageBuf = pageBuf[:0]
		pageBuf = strconv.AppendInt(pageBuf, int64(p), 10)
		pageBuf = append(pageBuf, " 0 R"...)
		pagesHdr.Write(pageBuf)
	}
	pageBuf = pageBuf[:0]
	pageBuf = append(pageBuf, "] /Count "...)
	pageBuf = strconv.AppendInt(pageBuf, int64(len(mergedPages)), 10)
	pageBuf = append(pageBuf, " >>\nendobj\n"...)
	pagesHdr.Write(pageBuf)
	out.WriteString(pagesHdr.String())

	for _, a := range appended {
		offsets[a.num] = out.Len()
		objHdr = objHdr[:0]
		objHdr = strconv.AppendInt(objHdr, int64(a.num), 10)
		objHdr = append(objHdr, " 0 obj\n"...)
		out.Write(objHdr)

		body := a.body
		if bytesIndex(body, []byte("/Type /Page")) >= 0 {
			if bytesIndex(body, []byte("/Parent")) < 0 {
				body = addParentRef(body, 2)
			} else {
				body = mergeParentRe.ReplaceAll(body, []byte("/Parent 2 0 R"))
			}
		}

		out.Write(body)
		out.WriteString("\nendobj\n")
	}

	maxObj := 2
	for k := range offsets {
		if k > maxObj {
			maxObj = k
		}
	}
	xrefStart := out.Len()
	var xrefBuf []byte
	xrefBuf = append(xrefBuf, "xref\n0 "...)
	xrefBuf = strconv.AppendInt(xrefBuf, int64(maxObj+1), 10)
	xrefBuf = append(xrefBuf, '\n')
	out.Write(xrefBuf)
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			var numBuf [16]byte
			numPart := strconv.AppendInt(numBuf[:0], int64(off), 10)
			xrefBuf = xrefBuf[:0]
			for j := 0; j < 10-len(numPart); j++ {
				xrefBuf = append(xrefBuf, '0')
			}
			xrefBuf = append(xrefBuf, numPart...)
			xrefBuf = append(xrefBuf, " 00000 n \n"...)
			out.Write(xrefBuf)
		} else {
			out.WriteString("0000000000 65535 f \n")
		}
	}

	var trailer []byte
	trailer = append(trailer, "trailer\n<< /Size "...)
	trailer = strconv.AppendInt(trailer, int64(maxObj+1), 10)
	trailer = append(trailer, " /Root 1 0 R >>\nstartxref\n"...)
	trailer = strconv.AppendInt(trailer, int64(xrefStart), 10)
	trailer = append(trailer, "\n%%EOF\n"...)
	out.Write(trailer)

	return out.Bytes(), nil
}

// replaceRefsOutsideStreams rewrites indirect references (n m R) in data only in regions
// that are not within stream...endstream blocks, to avoid mangling compressed stream contents.
func replaceRefsOutsideStreams(data []byte, refRe *regexp.Regexp, offset int) []byte {
	var out bytes.Buffer
	last := 0
	for _, sm := range mergeStreamRe.FindAllIndex(data, -1) {
		pre := data[last:sm[0]]
		replaced := refRe.ReplaceAllFunc(pre, func(b []byte) []byte {
			sm2 := refRe.FindSubmatch(b)
			if len(sm2) < 2 {
				return b
			}
			on, err := strconv.Atoi(string(sm2[1]))
			if err != nil {
				return b
			}
			gen := string(sm2[2])
			var buf [24]byte
			n := strconv.AppendInt(buf[:0], int64(offset+on), 10)
			n = append(n, ' ')
			n = append(n, gen...)
			n = append(n, ' ', 'R')
			return append([]byte(nil), n...)
		})
		out.Write(replaced)
		out.Write(data[sm[0]:sm[1]])
		last = sm[1]
	}
	if last < len(data) {
		tail := data[last:]
		replaced := refRe.ReplaceAllFunc(tail, func(b []byte) []byte {
			sm2 := refRe.FindSubmatch(b)
			if len(sm2) < 2 {
				return b
			}
			on, err := strconv.Atoi(string(sm2[1]))
			if err != nil {
				return b
			}
			gen := string(sm2[2])
			var buf [24]byte
			n := strconv.AppendInt(buf[:0], int64(offset+on), 10)
			n = append(n, ' ')
			n = append(n, gen...)
			n = append(n, ' ', 'R')
			return append([]byte(nil), n...)
		})
		out.Write(replaced)
	}
	return out.Bytes()
}

// addParentRef adds a /Parent reference to a page object's dictionary
func addParentRef(pageBody []byte, parentObjNum int) []byte {
	dictStart := bytes.Index(pageBody, []byte("<<"))
	if dictStart == -1 {
		return pageBody
	}

	var result bytes.Buffer
	var parentBuf []byte
	parentBuf = strconv.AppendInt(parentBuf, int64(parentObjNum), 10)
	parentBuf = append(parentBuf, " 0 R"...)
	result.Write(pageBody[:dictStart+2])
	result.WriteString(" /Parent ")
	result.Write(parentBuf)
	result.Write(pageBody[dictStart+2:])
	return result.Bytes()
}

// extractFormFieldsFromFile finds form field objects in a specific PDF file
func extractFormFieldsFromFile(pdfData []byte, objMap map[int][]byte) []int {
	var fields []int
	fieldSet := make(map[int]bool, 16)

	if rootRef, ok := findRootRef(pdfData); ok {
		parts := splitASCIIFields(rootRef)
		if len(parts) >= 1 {
			rootNum, err := strconv.Atoi(parts[0])
			if err == nil {
				if rootBody, exists := objMap[rootNum]; exists {
					if match := mergeAcroRe.FindSubmatch(rootBody); match != nil {
						if acroFormNum, err := strconv.Atoi(string(match[1])); err == nil {
							if acroFormBody, exists := objMap[acroFormNum]; exists {
								if fieldsMatch := mergeFieldsRe.FindSubmatch(acroFormBody); fieldsMatch != nil {
									for _, ref := range mergeRefRe.FindAllSubmatch(fieldsMatch[1], -1) {
										if fieldNum, err := strconv.Atoi(string(ref[1])); err == nil {
											if !fieldSet[fieldNum] {
												fields = append(fields, fieldNum)
												fieldSet[fieldNum] = true
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	for _, body := range objMap {
		if bytesIndex(body, []byte("/Type /Page")) >= 0 {
			if annotsMatch := mergeAnnotsRe.FindSubmatch(body); annotsMatch != nil {
				for _, ref := range mergeRefRe.FindAllSubmatch(annotsMatch[1], -1) {
					if annotNum, err := strconv.Atoi(string(ref[1])); err == nil {
						if annotBody, exists := objMap[annotNum]; exists {
							if bytesIndex(annotBody, []byte("/Subtype /Widget")) >= 0 {
								if !fieldSet[annotNum] {
									fields = append(fields, annotNum)
									fieldSet[annotNum] = true
								}
							}
						}
					}
				}
			}
		}
	}

	return fields
}

//nolint:revive // exported
// isFormFieldObject checks if an object body represents a form field
func IsFormFieldObject(body []byte) bool {
	formFieldTypes := [][]byte{
		[]byte("/FT /Tx"),
		[]byte("/FT /Ch"),
		[]byte("/FT /Btn"),
		[]byte("/FT /Sig"),
		[]byte("/Subtype /Widget"),
	}

	for _, fieldType := range formFieldTypes {
		if bytesIndex(body, fieldType) >= 0 {
			return true
		}
	}

	if bytesIndex(body, []byte("/T (")) >= 0 || bytesIndex(body, []byte("/T<")) >= 0 {
		return true
	}

	return false
}