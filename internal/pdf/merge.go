package pdf

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

// MergePDFs merges multiple PDF byte slices into a single PDF by parsing objects,
// remapping object numbers and building a new /Pages tree that references all
// page objects from the inputs. This avoids an external dependency.
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
	formFieldSet := make(map[int]bool)

	// currentMax tracks the highest object number assigned so far
	currentMax := 2

	// regex to find objects: num gen obj ... endobj
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)

	// Collect all remapped objects in-order and append them after catalog/pages
	var appended []struct {
		num  int
		body []byte
	}

	// Process files in the exact order they arrive
	for _, f := range files {
		// Reject encrypted PDFs for now
		if trailerHasEncrypt(f) {
			return nil, fmt.Errorf("cannot merge encrypted PDF")
		}

		// Build object map using same approach as DetectFormFieldsAdvanced
		objMatches := objRe.FindAllSubmatch(f, -1)
		if len(objMatches) == 0 {
			continue
		}

		objMap := make(map[int][]byte)
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
		tempObjMap := make(map[string][]byte)
		parseXRefStreams(f, tempObjMap)
		// merge tempObjMap into objMap (keys are like "<num> <gen>")
		for k, v := range tempObjMap {
			var onum int
			if _, err := fmt.Sscanf(k, "%d", &onum); err == nil {
				if _, exists := objMap[onum]; !exists {
					objMap[onum] = v
					if onum > maxObj {
						maxObj = onum
					}
				}
			}
		}

		offset := currentMax
		// Attempt to detect pages via the Pages tree (preferred) to avoid duplicates
		pagesFromTree := []int{}
		if rootRef, ok := findRootRef(f); ok {
			var rootNum, rootGen int
			if _, err := fmt.Sscanf(rootRef, "%d %d", &rootNum, &rootGen); err == nil {
				if rootBody, ok2 := objMap[rootNum]; ok2 {
					pagesRe := regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
					if pm := pagesRe.FindSubmatch(rootBody); pm != nil {
						if pnum, err := strconv.Atoi(string(pm[1])); err == nil {
							if pagesBody, ok3 := objMap[pnum]; ok3 {
								kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
								if km := kidsRe.FindSubmatch(pagesBody); km != nil {
									refReLocal := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
									for _, r := range refReLocal.FindAllSubmatch(km[1], -1) {
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
		var fileObjects []int
		for origNum := 1; origNum <= maxObj; origNum++ {
			if _, ok := objMap[origNum]; ok {
				fileObjects = append(fileObjects, origNum)
			}
		}

		// remap objects for this file
		for _, origNum := range fileObjects {
			body := objMap[origNum]

			// replace indirect references only outside stream blocks to avoid corrupting streams
			newBody := replaceRefsOutsideStreams(body, refRe, offset)

			newNum := offset + origNum
			appended = append(appended, struct {
				num  int
				body []byte
			}{num: newNum, body: newBody})

			// if this object is a Page object, record for Pages kids (maintain order)
			if len(pagesFromTree) == 0 {
				if bytesIndex(newBody, []byte("/Type /Page")) >= 0 || bytesIndex(newBody, []byte("/MediaBox")) >= 0 {
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
	catalogDict := "<< /Type /Catalog /Pages 2 0 R"
	if len(mergedFormFields) > 0 {
		catalogDict += " /AcroForm << /Fields ["
		for i, fieldNum := range mergedFormFields {
			if i > 0 {
				catalogDict += " "
			}
			catalogDict += fmt.Sprintf("%d 0 R", fieldNum)
		}
		catalogDict += "] >>"
	}
	catalogDict += " >>"
	out.WriteString(fmt.Sprintf("1 0 obj\n%s\nendobj\n", catalogDict))

	// Write Pages object (2) with all kids
	offsets[2] = out.Len()
	var kids []string
	for _, p := range mergedPages {
		kids = append(kids, fmt.Sprintf("%d 0 R", p))
	}
	// join kids into a single string
	var kidsStr string
	if len(kids) > 0 {
		kidsJoined := bytes.Join(byteSlice(kids), []byte(" "))
		kidsStr = string(kidsJoined)
	} else {
		kidsStr = ""
	}
	out.WriteString(fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", kidsStr, len(mergedPages)))

	// Append all remapped objects in the order they were processed
	for _, a := range appended {
		offsets[a.num] = out.Len()
		var b []byte
		b = strconv.AppendInt(b, int64(a.num), 10)
		b = append(b, " 0 obj\n"...)
		out.Write(b)

		// If this is a page object, ensure it has a parent reference
		body := a.body
		if bytesIndex(body, []byte("/Type /Page")) >= 0 {
			// Check if /Parent is already present
			if bytesIndex(body, []byte("/Parent")) < 0 {
				// Add parent reference to Pages object
				body = addParentRef(body, 2)
			} else {
				// Update existing parent reference to point to our Pages object
				parentRe := regexp.MustCompile(`/Parent\s+\d+\s+\d+\s+R`)
				body = parentRe.ReplaceAll(body, []byte("/Parent 2 0 R"))
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
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", maxObj+1))
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			out.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
		} else {
			out.WriteString("0000000000 65535 f \n")
		}
	}

	// trailer
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", maxObj+1, xrefStart))

	return out.Bytes(), nil
}

// helper to convert []string to [][]byte for bytes.Join
func byteSlice(in []string) [][]byte {
	out := make([][]byte, len(in))
	for i, s := range in {
		out[i] = []byte(s)
	}
	return out
}

// replaceRefsOutsideStreams rewrites indirect references (n m R) in data only in regions
// that are not within stream...endstream blocks, to avoid mangling compressed stream contents.
func replaceRefsOutsideStreams(data []byte, refRe *regexp.Regexp, offset int) []byte {
	var out bytes.Buffer
	streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	last := 0
	for _, sm := range streamRe.FindAllIndex(data, -1) {
		// process region before stream
		pre := data[last:sm[0]]
		replaced := refRe.ReplaceAllFunc(pre, func(b []byte) []byte {
			sm2 := refRe.FindSubmatch(b)
			if len(sm2) < 2 {
				return b
			}
			on, _ := strconv.Atoi(string(sm2[1]))
			gen := string(sm2[2])
			return []byte(fmt.Sprintf("%d %s R", offset+on, gen))
		})
		out.Write(replaced)
		// write stream block unchanged
		out.Write(data[sm[0]:sm[1]])
		last = sm[1]
	}
	// remaining tail
	if last < len(data) {
		tail := data[last:]
		replaced := refRe.ReplaceAllFunc(tail, func(b []byte) []byte {
			sm2 := refRe.FindSubmatch(b)
			if len(sm2) < 2 {
				return b
			}
			on, _ := strconv.Atoi(string(sm2[1]))
			gen := string(sm2[2])
			return []byte(fmt.Sprintf("%d %s R", offset+on, gen))
		})
		out.Write(replaced)
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
	result.Write(pageBody[:dictStart+2])
	result.WriteString(fmt.Sprintf(" /Parent %d 0 R", parentObjNum))
	result.Write(pageBody[dictStart+2:])

	return result.Bytes()
}

// extractFormFieldsFromFile finds form field objects in a specific PDF file
func extractFormFieldsFromFile(pdfData []byte, objMap map[int][]byte) []int {
	var fields []int
	fieldSet := make(map[int]bool) // Avoid duplicates within this file

	// First try to find AcroForm in the catalog
	if rootRef, ok := findRootRef(pdfData); ok {
		var rootNum int
		if _, err := fmt.Sscanf(rootRef, "%d", &rootNum); err == nil {
			if rootBody, exists := objMap[rootNum]; exists {
				acroFormRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s+\d+\s+R`)
				if match := acroFormRe.FindSubmatch(rootBody); match != nil {
					if acroFormNum, err := strconv.Atoi(string(match[1])); err == nil {
						if acroFormBody, exists := objMap[acroFormNum]; exists {
							fieldsRe := regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
							if fieldsMatch := fieldsRe.FindSubmatch(acroFormBody); fieldsMatch != nil {
								refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)
								for _, ref := range refRe.FindAllSubmatch(fieldsMatch[1], -1) {
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
		if bytesIndex(body, []byte("/Type /Page")) >= 0 {
			annotsRe := regexp.MustCompile(`/Annots\s*\[(.*?)\]`)
			if annotsMatch := annotsRe.FindSubmatch(body); annotsMatch != nil {
				refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)
				for _, ref := range refRe.FindAllSubmatch(annotsMatch[1], -1) {
					if annotNum, err := strconv.Atoi(string(ref[1])); err == nil {
						if annotBody, exists := objMap[annotNum]; exists {
							// Check if this annotation is a widget (form field)
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
	if bytesIndex(body, []byte("/T (")) >= 0 || bytesIndex(body, []byte("/T<")) >= 0 {
		return true
	}

	return false
}
