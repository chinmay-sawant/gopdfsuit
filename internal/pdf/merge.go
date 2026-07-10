package pdf

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// Package-level compiled regexes used during merge (hoisted out of per-file loops).
var (
	mergeObjRe       = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	mergeRefRe       = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	mergeSimpleRefRe = regexp.MustCompile(`(\d+)\s+\d+\s+R`)
	pagesRefRe       = regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
	kidsArrayRe      = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	parentRefRe      = regexp.MustCompile(`/Parent\s+\d+\s+\d+\s+R`)
	streamBlockRe    = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	acroFormRefRe    = regexp.MustCompile(`/AcroForm\s+(\d+)\s+\d+\s+R`)
	fieldsArrayRe    = regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	annotsArrayRe    = regexp.MustCompile(`/Annots\s*\[(.*?)\]`)
)

// Package-level PDF markers for merge scans.
var (
	markerTypePage      = []byte("/Type /Page")
	markerMediaBox      = []byte("/MediaBox")
	markerParent        = []byte("/Parent")
	markerSubtypeWidget = []byte("/Subtype /Widget")
	markerTParen        = []byte("/T (")
	markerTAngle        = []byte("/T<")
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

	// Reserve object numbers 1 (Catalog) and 2 (Pages)
	offsets := map[int]int{}

	// Keep list of merged page object numbers (new numbers) in order
	mergedPages := []int{}

	// Keep list of merged form fields (new numbers) avoiding duplicates
	mergedFormFields := []int{}
	formFieldSet := make(map[int]bool, 16) // PERF-192

	// currentMax tracks the highest object number assigned so far
	currentMax := 2

	// Collect all remapped objects in-order and append them after catalog/pages
	var appended []struct {
		num  int
		body []byte
	}

	// Process files in the exact order they arrive
	for _, f := range files {
		// Reject encrypted PDFs for now
		if trailerHasEncrypt(f) {
			return nil, errors.New("cannot merge encrypted PDF")
		}

		// Build object map using same approach as DetectFormFieldsAdvanced
		objMatches := mergeObjRe.FindAllSubmatch(f, -1)
		if len(objMatches) == 0 {
			continue
		}

		objMap := make(map[int][]byte, len(objMatches))
		maxObj := 0
		for _, m := range objMatches {
			if n, err := strconv.Atoi(string(m[1])); err == nil {
				// Preserve original body (including stream markers) so we don't corrupt streams
				body := m[3]
				objMap[n] = body
				if n > maxObj {
					maxObj = n
				}
			}
		}

		// Allow parseXRefStreams to augment object map (it operates on raw bytes in this package)
		tempObjMap := make(map[string][]byte, 16) // PERF-192
		parseXRefStreams(f, tempObjMap)
		// merge tempObjMap into objMap (keys are like "<num> <gen>")
		for k, v := range tempObjMap {
			onum, ok := parseLeadingInt(k)
			if !ok {
				continue
			}
			if _, exists := objMap[onum]; !exists {
				objMap[onum] = v
				if onum > maxObj {
					maxObj = onum
				}
			}
		}

		offset := currentMax
		// Attempt to detect pages via the Pages tree (preferred) to avoid duplicates
		pagesFromTree := []int{}
		if rootRef, ok := findRootRef(f); ok {
			rootNum, _, okRoot := parseObjGenRef(rootRef)
			if okRoot {
				if rootBody, ok2 := objMap[rootNum]; ok2 {
					if pm := pagesRefRe.FindSubmatch(rootBody); pm != nil {
						if pnum, err := strconv.Atoi(string(pm[1])); err == nil {
							if pagesBody, ok3 := objMap[pnum]; ok3 {
								if km := kidsArrayRe.FindSubmatch(pagesBody); km != nil {
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

		// Extract form fields from this PDF ONLY
		formFields := extractFormFieldsFromFile(f, objMap)

		// Process objects in numeric order to maintain consistency
		fileObjects := make([]int, 0, len(objMap))
		for origNum := 1; origNum <= maxObj; origNum++ {
			if _, ok := objMap[origNum]; ok {
				fileObjects = append(fileObjects, origNum)
			}
		}

		// remap objects for this file
		for _, origNum := range fileObjects {
			body := objMap[origNum]

			// replace indirect references only outside stream blocks to avoid corrupting streams
			newBody := replaceRefsOutsideStreams(body, mergeRefRe, offset)

			newNum := offset + origNum
			appended = append(appended, struct {
				num  int
				body []byte
			}{num: newNum, body: newBody})

			// if this object is a Page object, record for Pages kids (maintain order)
			if len(pagesFromTree) == 0 {
				if bytesIndex(newBody, markerTypePage) >= 0 || bytesIndex(newBody, markerMediaBox) >= 0 {
					mergedPages = append(mergedPages, newNum)
				}
			}
		}

		// if we obtained pages from Pages tree, map them to remapped numbers and add to mergedPages (maintain order)
		for _, pn := range pagesFromTree {
			mergedPages = append(mergedPages, offset+pn)
		}

		// Map form fields to remapped numbers (avoid duplicates across files)
		for _, fn := range formFields {
			remappedFieldNum := offset + fn
			if !formFieldSet[remappedFieldNum] {
				mergedFormFields = append(mergedFormFields, remappedFieldNum)
				formFieldSet[remappedFieldNum] = true
			}
		}

		currentMax = offset + maxObj + 1
	}

	// Write Catalog object (1) - now includes AcroForm if we have form fields
	offsets[1] = out.Len()
	var catalogBuf strings.Builder
	var tmp [20]byte
	catalogBuf.WriteString("<< /Type /Catalog /Pages 2 0 R")
	if len(mergedFormFields) > 0 {
		catalogBuf.WriteString(" /AcroForm << /Fields [")
		for i, fieldNum := range mergedFormFields {
			if i > 0 {
				catalogBuf.WriteByte(' ')
			}
			catalogBuf.Write(strconv.AppendInt(tmp[:0], int64(fieldNum), 10))
			catalogBuf.WriteString(" 0 R")
		}
		catalogBuf.WriteString("] >>")
	}
	catalogBuf.WriteString(" >>")
	out.WriteString("1 0 obj\n")
	out.WriteString(catalogBuf.String())
	out.WriteString("\nendobj\n")

	// Write Pages object (2) with all kids
	offsets[2] = out.Len()
	var kidsBuf strings.Builder
	for i, p := range mergedPages {
		if i > 0 {
			kidsBuf.WriteByte(' ')
		}
		kidsBuf.Write(strconv.AppendInt(tmp[:0], int64(p), 10))
		kidsBuf.WriteString(" 0 R")
	}
	out.WriteString("2 0 obj\n<< /Type /Pages /Kids [")
	out.WriteString(kidsBuf.String())
	out.WriteString("] /Count ")
	out.Write(strconv.AppendInt(tmp[:0], int64(len(mergedPages)), 10))
	out.WriteString(" >>\nendobj\n")

	// Append all remapped objects in the order they were processed
	for _, a := range appended {
		offsets[a.num] = out.Len()
		var b []byte
		b = strconv.AppendInt(b, int64(a.num), 10)
		b = append(b, " 0 obj\n"...)
		out.Write(b)

		// If this is a page object, ensure it has a parent reference
		body := a.body
		if bytesIndex(body, markerTypePage) >= 0 {
			// Check if /Parent is already present
			if bytesIndex(body, markerParent) < 0 {
				// Add parent reference to Pages object
				body = addParentRef(body, 2)
			} else {
				// Update existing parent reference to point to our Pages object
				body = parentRefRe.ReplaceAll(body, []byte("/Parent 2 0 R"))
			}
		}

		out.Write(body)
		out.WriteString("\nendobj\n")
	}

	// Build xref
	maxObj := 2
	for k := range offsets {
		if k > maxObj {
			maxObj = k
		}
	}
	xrefStart := out.Len()
	var xrefTmp [20]byte
	out.WriteString("xref\n0 ")
	out.Write(strconv.AppendInt(xrefTmp[:0], int64(maxObj+1), 10))
	out.WriteString("\n0000000000 65535 f \n")
	// Fixed 20-byte entry: 10-digit offset + " 00000 n \n" (PERF-119/128)
	var xrefEntry [20]byte
	copy(xrefEntry[10:], " 00000 n \n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			offStr := strconv.AppendInt(xrefTmp[:0], int64(off), 10)
			pad := 10 - len(offStr)
			for j := 0; j < pad; j++ {
				xrefEntry[j] = '0'
			}
			copy(xrefEntry[pad:10], offStr)
			out.Write(xrefEntry[:])
		} else {
			out.WriteString("0000000000 65535 f \n")
		}
	}

	// trailer
	out.WriteString("trailer\n<< /Size ")
	out.Write(strconv.AppendInt(xrefTmp[:0], int64(maxObj+1), 10))
	out.WriteString(" /Root 1 0 R >>\nstartxref\n")
	out.Write(strconv.AppendInt(xrefTmp[:0], int64(xrefStart), 10))
	out.WriteString("\n%%%%EOF\n")

	return out.Bytes(), nil
}

// parseLeadingInt parses a leading integer from strings like "12 0" or "12".
func parseLeadingInt(s string) (int, bool) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(s[:end])
	return n, err == nil
}

// parseObjGenRef parses "num gen" object references from findRootRef.
func parseObjGenRef(ref string) (objNum, genNum int, ok bool) {
	space := strings.IndexByte(ref, ' ')
	if space <= 0 {
		return 0, 0, false
	}
	var err error
	objNum, err = strconv.Atoi(ref[:space])
	if err != nil {
		return 0, 0, false
	}
	genNum, err = strconv.Atoi(ref[space+1:])
	if err != nil {
		return 0, 0, false
	}
	return objNum, genNum, true
}

// replaceRefsOutsideStreams rewrites indirect references (n m R) in data only in regions
// that are not within stream...endstream blocks, to avoid mangling compressed stream contents.
func replaceRefsOutsideStreams(data []byte, refRe *regexp.Regexp, offset int) []byte {
	var out bytes.Buffer
	var refBuf [32]byte

	buildRef := func(b []byte) []byte {
		sm2 := refRe.FindSubmatch(b)
		if len(sm2) < 2 {
			return b
		}
		on, _ := strconv.Atoi(string(sm2[1]))
		gen := sm2[2]
		num := strconv.AppendInt(refBuf[:0], int64(offset+on), 10)
		ref := make([]byte, len(num)+1+len(gen)+2)
		copy(ref, num)
		ref[len(num)] = ' '
		copy(ref[len(num)+1:], gen)
		copy(ref[len(num)+1+len(gen):], " R")
		return ref
	}

	last := 0
	for _, sm := range streamBlockRe.FindAllIndex(data, -1) {
		pre := data[last:sm[0]]
		out.Write(refRe.ReplaceAllFunc(pre, buildRef))
		out.Write(data[sm[0]:sm[1]])
		last = sm[1]
	}

	if last < len(data) {
		out.Write(refRe.ReplaceAllFunc(data[last:], buildRef))
	}
	return out.Bytes()
}

// addParentRef adds a /Parent reference to a page object's dictionary
func addParentRef(pageBody []byte, parentObjNum int) []byte {
	// Find the end of the opening dictionary
	dictStart := bytes.Index(pageBody, []byte("<<"))
	if dictStart == -1 {
		return pageBody
	}

	// Insert /Parent reference after the opening <<
	var result bytes.Buffer
	var tmp [20]byte
	result.Write(pageBody[:dictStart+2])
	result.WriteString(" /Parent ")
	result.Write(strconv.AppendInt(tmp[:0], int64(parentObjNum), 10))
	result.WriteString(" 0 R")
	result.Write(pageBody[dictStart+2:])

	return result.Bytes()
}

// extractFormFieldsFromFile finds form field objects in a specific PDF file
func extractFormFieldsFromFile(pdfData []byte, objMap map[int][]byte) []int {
	var fields []int
	fieldSet := make(map[int]bool, 16) // PERF-192: avoid duplicates within this file

	// First try to find AcroForm in the catalog
	if rootRef, ok := findRootRef(pdfData); ok {
		rootNum, okRoot := parseLeadingInt(rootRef)
		if okRoot {
			if rootBody, exists := objMap[rootNum]; exists {
				if match := acroFormRefRe.FindSubmatch(rootBody); match != nil {
					if acroFormNum, err := strconv.Atoi(string(match[1])); err == nil {
						if acroFormBody, exists := objMap[acroFormNum]; exists {
							if fieldsMatch := fieldsArrayRe.FindSubmatch(acroFormBody); fieldsMatch != nil {
								for _, ref := range mergeSimpleRefRe.FindAllSubmatch(fieldsMatch[1], -1) {
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

	// Also scan for widget annotations in page objects from THIS file only
	for _, body := range objMap {
		if bytesIndex(body, markerTypePage) >= 0 {
			if annotsMatch := annotsArrayRe.FindSubmatch(body); annotsMatch != nil {
				for _, ref := range mergeSimpleRefRe.FindAllSubmatch(annotsMatch[1], -1) {
					if annotNum, err := strconv.Atoi(string(ref[1])); err == nil {
						if annotBody, exists := objMap[annotNum]; exists {
							// Check if this annotation is a widget (form field)
							if bytesIndex(annotBody, markerSubtypeWidget) >= 0 {
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

// isFormFieldObject checks if an object body represents a form field
//
//nolint:revive // exported
func IsFormFieldObject(body []byte) bool {
	// Check for common form field indicators
	formFieldTypes := [][]byte{
		[]byte("/FT /Tx"),          // Text field
		[]byte("/FT /Ch"),          // Choice field (combo/list)
		[]byte("/FT /Btn"),         // Button field (radio/checkbox)
		[]byte("/FT /Sig"),         // Signature field
		[]byte("/Subtype /Widget"), // Widget annotation
	}

	for _, fieldType := range formFieldTypes {
		if bytesIndex(body, fieldType) >= 0 {
			return true
		}
	}

	// Also check for /T (field name) which is required for form fields
	if bytesIndex(body, markerTParen) >= 0 || bytesIndex(body, markerTAngle) >= 0 {
		return true
	}

	return false
}
