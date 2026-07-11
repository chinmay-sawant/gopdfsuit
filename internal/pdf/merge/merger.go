package merge

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Package-level regexes used by merger helpers (PERF-1).
var (
	pagesRefRe  = regexp.MustCompile(`/Pages\s+(\d+)\s+\d+\s+R`)
	parentRefRe = regexp.MustCompile(`/Parent\s+\d+\s+\d+\s+R`)
	trailerRe   = regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`)
)

// ensureLen grows a slice to at least the required length.
func ensureLen[T any](s *[]T, need int) {
	if need > len(*s) {
		newCap := need * 2
		if newCap < 64 {
			newCap = 64
		}
		newS := make([]T, need, newCap)
		copy(newS, *s)
		*s = newS
	}
}

// MergePDFs merges multiple PDF files into one
// It properly handles form fields, widgets, appearance streams, and various PDF versions
//
//nolint:revive // exported
func MergePDFs(files [][]byte) ([]byte, error) {
	if len(files) == 0 {
		return nil, errors.New("no PDF files provided")
	}

	ctx := NewMergeContext()

	// Parse all input files
	var fileContexts []*FileContext
	for _, f := range files {
		if hasEncrypt(f) {
			return nil, errors.New("cannot merge encrypted PDF")
		}

		fc := parseFile(f)
		if fc == nil {
			continue
		}

		// Track highest PDF version
		v := DetectPDFVersion(f)
		if CompareVersions(v, ctx.HighestVersion) > 0 {
			ctx.HighestVersion = v
		}

		fileContexts = append(fileContexts, fc)
	}

	if len(fileContexts) == 0 {
		return nil, errors.New("no valid PDF files to merge")
	}

	// Write PDF header (PERF-35: no fmt boxing)
	ctx.Output.WriteString("%PDF-")
	ctx.Output.WriteString(ctx.HighestVersion)
	ctx.Output.WriteString("\n%\xe2\xe3\xcf\xd3\n")

	// Process each file
	var appendedObjects []struct {
		num  int
		body []byte
	}

	for _, fc := range fileContexts {
		offset := ctx.CurrentMax

		// Collect all objects to process (including annotation dependencies)
		objectsToProcess := collectObjectsWithDependencies(fc)

		// Process objects maintaining order
		for _, origNum := range objectsToProcess {
			if origNum >= len(fc.Objects) || fc.Objects[origNum] == nil {
				continue
			}
			body := fc.Objects[origNum]

			newNum := offset + origNum

			// Remap references in body
			newBody := ReplaceRefsOutsideStreams(body, offset)

			// Special handling for Page objects - update annotations
			if IsPageObject(newBody) && !IsPagesTreeObject(newBody) {
				ctx.MergedPages = append(ctx.MergedPages, newNum)
			}

			appendedObjects = append(appendedObjects, struct {
				num  int
				body []byte
			}{num: newNum, body: newBody})
		}

		// Track form fields with remapped numbers
		for _, fn := range fc.FormFields {
			remapped := offset + fn
			if !ctx.FieldSet[remapped] {
				ctx.MergedFields = append(ctx.MergedFields, remapped)
				ctx.FieldSet[remapped] = true
			}
		}

		ctx.CurrentMax = offset + fc.MaxObj + 1
	}

	// Write Catalog (object 1)
	ensureLen(&ctx.Offsets, 2)
	ctx.Offsets[1] = ctx.Output.Len()
	writeCatalog(&ctx.Output, ctx.MergedFields)

	// Write Pages (object 2)
	ctx.Offsets[2] = ctx.Output.Len()
	writePages(&ctx.Output, ctx.MergedPages)

	// Write all remapped objects
	for _, obj := range appendedObjects {
		ensureLen(&ctx.Offsets, obj.num+1)
		ctx.Offsets[obj.num] = ctx.Output.Len()
		writeObject(&ctx.Output, obj.num, obj.body)
	}

	// Write xref and trailer
	writeXRefAndTrailer(&ctx.Output, ctx.Offsets)

	return ctx.Output.Bytes(), nil
}

// parseFile parses a PDF file into a FileContext
func parseFile(data []byte) *FileContext {
	fc := NewFileContext(data)

	// Find all objects
	boundaries := FindObjectBoundaries(data)
	if len(boundaries) == 0 {
		return nil
	}

	for _, b := range boundaries {
		// Extract body (from after "obj" to before "endobj")
		bodyEnd := b.End - 6 // subtract len("endobj")
		// Trim trailing whitespace
		for bodyEnd > b.BodyStart && isWhitespace(data[bodyEnd-1]) {
			bodyEnd--
		}
		body := data[b.BodyStart:bodyEnd]
		fc.Objects[b.ObjNum] = body
		if b.ObjNum > fc.MaxObj {
			fc.MaxObj = b.ObjNum
		}
	}

	// Extract objects from Object Streams (PDF 1.5+)
	for objNum, body := range fc.Objects {
		if body == nil || !IsObjectStream(body) {
			continue
		}
		extractedObjs := ParseObjectStream(body)
		for extractedNum, extractedBody := range extractedObjs {
			if extractedBody == nil {
				continue
			}
			// Only add if not already present (top-level objects take precedence)
			if extractedNum >= len(fc.Objects) || fc.Objects[extractedNum] == nil {
				if extractedNum >= len(fc.Objects) {
					newObjs := make([][]byte, extractedNum+1)
					copy(newObjs, fc.Objects)
					fc.Objects = newObjs
				}
				fc.Objects[extractedNum] = extractedBody
				if extractedNum > fc.MaxObj {
					fc.MaxObj = extractedNum
				}
			}
		}
		// Mark original ObjStm for exclusion (we've expanded it)
		fc.ObjectStreamNums = append(fc.ObjectStreamNums, objNum)
	}

	// Extract pages from Pages tree
	fc.Pages = extractPagesFromTree(data, fc.Objects)

	// Find the original catalog and pages tree to exclude them
	fc.OriginalCatalog, fc.OriginalPagesTree = findCatalogAndPages(data, fc.Objects)

	// Extract form fields and annotation dependencies
	ExtractFormFields(fc)

	return fc
}

// findCatalogAndPages finds the original Catalog and Pages tree object numbers
func findCatalogAndPages(data []byte, objMap [][]byte) (catalogNum int, pagesNum int) {
	rootRef := findRootRef(data)
	if rootRef == "" {
		return 0, 0
	}

	if _, err := fmt.Sscanf(rootRef, "%d", &catalogNum); err != nil {
		return 0, 0
	}

	if catalogNum > 0 && catalogNum < len(objMap) && objMap[catalogNum] != nil {
		match := pagesRefRe.FindSubmatch(objMap[catalogNum])
		if match != nil {
			pagesNum, _ = strconv.Atoi(string(match[1]))
		}
	}

	return catalogNum, pagesNum
}

// extractPagesFromTree extracts page object numbers from the Pages tree
func extractPagesFromTree(data []byte, objMap [][]byte) []int {
	var pages []int

	rootRef := findRootRef(data)
	if rootRef == "" {
		return pages
	}

	// Parse root object number
	var rootNum int
	if _, err := fmt.Sscanf(rootRef, "%d", &rootNum); err != nil {
		return pages
	}

	if rootNum >= len(objMap) || objMap[rootNum] == nil {
		return pages
	}
	rootBody := objMap[rootNum]

	// Find /Pages reference
	match := pagesRefRe.FindSubmatch(rootBody)
	if match == nil {
		return pages
	}

	pagesNum, _ := strconv.Atoi(string(match[1]))
	if pagesNum >= len(objMap) || objMap[pagesNum] == nil {
		return pages
	}
	pagesBody := objMap[pagesNum]

	// Recursively extract kids
	return extractKidsRecursive(pagesBody, objMap, objRefRe)
}

// extractKidsRecursive extracts page numbers from /Kids array
func extractKidsRecursive(pagesBody []byte, objMap [][]byte, refRe *regexp.Regexp) []int {
	var pages []int

	match := kidsArrayRe.FindSubmatch(pagesBody)
	if match == nil {
		return pages
	}

	for _, r := range refRe.FindAllSubmatch(match[1], -1) {
		kidNum, _ := strconv.Atoi(string(r[1]))
		if kidNum >= len(objMap) || objMap[kidNum] == nil {
			pages = append(pages, kidNum)
			continue
		}
		kidBody := objMap[kidNum]

		if IsPagesTreeObject(kidBody) {
			// Recursive: nested Pages node
			pages = append(pages, extractKidsRecursive(kidBody, objMap, refRe)...)
		} else {
			// Leaf: Page object
			pages = append(pages, kidNum)
		}
	}

	return pages
}

// collectObjectsWithDependencies returns all object numbers to process
// ensuring annotation dependencies are included but excluding original catalog/pages/objstm
func collectObjectsWithDependencies(fc *FileContext) []int {
	n := fc.MaxObj + 1
	if n < 8 {
		n = 8
	}
	included := make([]bool, n)
	excluded := make([]bool, n)
	var result []int

	// Mark objects to exclude
	if fc.OriginalCatalog > 0 && fc.OriginalCatalog < n {
		excluded[fc.OriginalCatalog] = true
	}
	if fc.OriginalPagesTree > 0 && fc.OriginalPagesTree < n {
		excluded[fc.OriginalPagesTree] = true
	}
	for _, objStmNum := range fc.ObjectStreamNums {
		if objStmNum < n {
			excluded[objStmNum] = true
		}
	}

	// Also exclude any intermediate Pages tree nodes
	for num, body := range fc.Objects {
		if body == nil || num >= n {
			continue
		}
		if IsPagesTreeObject(body) {
			excluded[num] = true
		}
	}

	// Add all objects in numeric order, excluding catalog/pages/objstm
	for i := 1; i <= fc.MaxObj; i++ {
		if i < n && i < len(fc.Objects) && fc.Objects[i] != nil {
			if !included[i] && !excluded[i] {
				result = append(result, i)
				included[i] = true
			}
		}
	}

	// Ensure all AP dependencies are included
	for _, deps := range fc.APDeps {
		for _, dep := range deps {
			if dep < n && !included[dep] && !excluded[dep] {
				result = append(result, dep)
				included[dep] = true
			}
		}
	}

	return result
}

// writeCatalog writes the Catalog object
func writeCatalog(out *bytes.Buffer, fields []int) {
	out.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R")

	if len(fields) > 0 {
		out.WriteString(" /AcroForm << /Fields [")
		var fieldBuf []byte
		for i, f := range fields {
			if i > 0 {
				out.WriteString(" ")
			}
			fieldBuf = fieldBuf[:0]
			fieldBuf = strconv.AppendInt(fieldBuf, int64(f), 10)
			fieldBuf = append(fieldBuf, " 0 R"...)
			out.Write(fieldBuf)
		}
		out.WriteString("] /NeedAppearances true >>")
	}

	out.WriteString(" >>\nendobj\n")
}

// writePages writes the Pages object
func writePages(out *bytes.Buffer, pages []int) {
	out.WriteString("2 0 obj\n<< /Type /Pages /Kids [")
	var pageBuf []byte
	for i, p := range pages {
		if i > 0 {
			out.WriteString(" ")
		}
		pageBuf = pageBuf[:0]
		pageBuf = strconv.AppendInt(pageBuf, int64(p), 10)
		pageBuf = append(pageBuf, " 0 R"...)
		out.Write(pageBuf)
	}
	pageBuf = pageBuf[:0]
	pageBuf = append(pageBuf, "] /Count "...)
	pageBuf = strconv.AppendInt(pageBuf, int64(len(pages)), 10)
	pageBuf = append(pageBuf, " >>\nendobj\n"...)
	out.Write(pageBuf)
}

// writeObject writes a single PDF object
func writeObject(out *bytes.Buffer, num int, body []byte) {
	var objBuf []byte
	objBuf = strconv.AppendInt(objBuf, int64(num), 10)
	objBuf = append(objBuf, " 0 obj"...)
	out.Write(objBuf)

	// Handle Page objects - ensure /Parent points to our Pages object
	if IsPageObject(body) && !IsPagesTreeObject(body) {
		body = updateParentRef(body)
	}

	body = bytes.TrimSpace(body)
	out.Write(body)
	out.WriteString("\nendobj\n")
}

// updateParentRef updates or adds /Parent reference
func updateParentRef(body []byte) []byte {
	if parentRefRe.Match(body) {
		// Update existing
		return parentRefRe.ReplaceAll(body, []byte("/Parent 2 0 R"))
	}

	// Add new parent reference after <<
	dictStart := bytes.Index(body, []byte("<<"))
	if dictStart == -1 {
		return body
	}

	var result bytes.Buffer
	result.Write(body[:dictStart+2])
	result.WriteString(" /Parent 2 0 R")
	result.Write(body[dictStart+2:])
	return result.Bytes()
}

// writeXRefAndTrailer writes the xref table and trailer
func writeXRefAndTrailer(out *bytes.Buffer, offsets []int) {
	maxObj := len(offsets) - 1
	if maxObj < 0 {
		maxObj = 0
	}

	xrefStart := out.Len()
	var xrefBuf []byte
	xrefBuf = append(xrefBuf, "xref\n0 "...)
	xrefBuf = strconv.AppendInt(xrefBuf, int64(maxObj+1), 10)
	xrefBuf = append(xrefBuf, '\n')
	out.Write(xrefBuf)

	// Object 0 is always free
	out.WriteString("0000000000 65535 f\r\n")

	// Write entries for objects 1 to maxObj (fixed 19-byte line, PERF-119/128)
	var entry [19]byte
	copy(entry[10:], " 00000 n\r\n")
	var offTmp [20]byte
	for i := 1; i <= maxObj; i++ {
		if off := offsets[i]; off != 0 {
			offStr := strconv.AppendInt(offTmp[:0], int64(off), 10)
			pad := 10 - len(offStr)
			for j := 0; j < pad; j++ {
				entry[j] = '0'
			}
			copy(entry[pad:10], offStr)
			out.Write(entry[:])
		} else {
			out.WriteString("0000000000 65535 f\r\n")
		}
	}

	// Trailer
	xrefBuf = xrefBuf[:0]
	xrefBuf = append(xrefBuf, "trailer\n<< /Size "...)
	xrefBuf = strconv.AppendInt(xrefBuf, int64(maxObj+1), 10)
	xrefBuf = append(xrefBuf, " /Root 1 0 R >>\nstartxref\n"...)
	xrefBuf = strconv.AppendInt(xrefBuf, int64(xrefStart), 10)
	xrefBuf = append(xrefBuf, "\n%%%%EOF\n"...)
	out.Write(xrefBuf)
}

// hasEncrypt checks if PDF is encrypted
func hasEncrypt(data []byte) bool {
	matches := trailerRe.FindAllSubmatch(data, -1)
	for _, m := range matches {
		if bytes.Contains(m[1], []byte("/Encrypt")) {
			return true
		}
	}
	return bytes.Contains(data, []byte("/Encrypt"))
}
