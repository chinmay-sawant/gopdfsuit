package merge

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

// MergePDFs merges multiple PDF files into one
// It properly handles form fields, widgets, appearance streams, and various PDF versions
//nolint:revive // exported
func MergePDFs(files [][]byte) ([]byte, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no PDF files provided")
	}

	ctx := NewMergeContext()

	// Parse all input files
	var fileContexts []*FileContext
	for _, f := range files {
		if hasEncrypt(f) {
			return nil, fmt.Errorf("cannot merge encrypted PDF")
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
		return nil, fmt.Errorf("no valid PDF files to merge")
	}

	// Write PDF header
	ctx.Output.WriteString(fmt.Sprintf("%%PDF-%s\n%%\xe2\xe3\xcf\xd3\n", ctx.HighestVersion))

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
			body, exists := fc.Objects[origNum]
			if !exists {
				continue
			}

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
	ctx.Offsets[1] = ctx.Output.Len()
	writeCatalog(&ctx.Output, ctx.MergedFields)

	// Write Pages (object 2)
	ctx.Offsets[2] = ctx.Output.Len()
	writePages(&ctx.Output, ctx.MergedPages)

	// Write all remapped objects
	for _, obj := range appendedObjects {
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
		if IsObjectStream(body) {
			extractedObjs := ParseObjectStream(body)
			for extractedNum, extractedBody := range extractedObjs {
				// Only add if not already present (top-level objects take precedence)
				if _, exists := fc.Objects[extractedNum]; !exists {
					fc.Objects[extractedNum] = extractedBody
					if extractedNum > fc.MaxObj {
						fc.MaxObj = extractedNum
					}
				}
			}
			// Mark original ObjStm for exclusion (we've expanded it)
			fc.ObjectStreamNums = append(fc.ObjectStreamNums, objNum)
		}
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
func findCatalogAndPages(data []byte, objMap map[int][]byte) (catalogNum int, pagesNum int) {
	rootRef := findRootRef(data)
	if rootRef == "" {
		return 0, 0
	}

	if _, err := fmt.Sscanf(rootRef, "%d", &catalogNum); err != nil {
		return 0, 0
	}

	if catalogNum > 0 {
		if body, exists := objMap[catalogNum]; exists {
			pagesRe := regexp.MustCompile(`/Pages\s+(\d+)\s+\d+\s+R`)
			match := pagesRe.FindSubmatch(body)
			if match != nil {
				pagesNum, _ = strconv.Atoi(string(match[1]))
			}
		}
	}

	return catalogNum, pagesNum
}

// extractPagesFromTree extracts page object numbers from the Pages tree
func extractPagesFromTree(data []byte, objMap map[int][]byte) []int {
	var pages []int
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)

	rootRef := findRootRef(data)
	if rootRef == "" {
		return pages
	}

	// Parse root object number
	var rootNum int
	if _, err := fmt.Sscanf(rootRef, "%d", &rootNum); err != nil {
		return pages
	}

	rootBody, exists := objMap[rootNum]
	if !exists {
		return pages
	}

	// Find /Pages reference
	pagesRe := regexp.MustCompile(`/Pages\s+(\d+)\s+\d+\s+R`)
	match := pagesRe.FindSubmatch(rootBody)
	if match == nil {
		return pages
	}

	pagesNum, _ := strconv.Atoi(string(match[1]))
	pagesBody, exists := objMap[pagesNum]
	if !exists {
		return pages
	}

	// Recursively extract kids
	return extractKidsRecursive(pagesBody, objMap, refRe)
}

// extractKidsRecursive extracts page numbers from /Kids array
func extractKidsRecursive(pagesBody []byte, objMap map[int][]byte, refRe *regexp.Regexp) []int {
	var pages []int
	kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)

	match := kidsRe.FindSubmatch(pagesBody)
	if match == nil {
		return pages
	}

	for _, r := range refRe.FindAllSubmatch(match[1], -1) {
		kidNum, _ := strconv.Atoi(string(r[1]))
		kidBody, ok := objMap[kidNum]
		if !ok {
			pages = append(pages, kidNum)
			continue
		}

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
	included := make(map[int]bool)
	excluded := make(map[int]bool)
	var result []int

	// Mark objects to exclude
	if fc.OriginalCatalog > 0 {
		excluded[fc.OriginalCatalog] = true
	}
	if fc.OriginalPagesTree > 0 {
		excluded[fc.OriginalPagesTree] = true
	}
	for _, objStmNum := range fc.ObjectStreamNums {
		excluded[objStmNum] = true
	}

	// Also exclude any intermediate Pages tree nodes
	for num, body := range fc.Objects {
		if IsPagesTreeObject(body) {
			excluded[num] = true
		}
	}

	// Add all objects in numeric order, excluding catalog/pages/objstm
	for i := 1; i <= fc.MaxObj; i++ {
		if _, exists := fc.Objects[i]; exists {
			if !included[i] && !excluded[i] {
				result = append(result, i)
				included[i] = true
			}
		}
	}

	// Ensure all AP dependencies are included
	for _, deps := range fc.APDeps {
		for _, dep := range deps {
			if !included[dep] && !excluded[dep] {
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
	parentRe := regexp.MustCompile(`/Parent\s+\d+\s+\d+\s+R`)

	if parentRe.Match(body) {
		// Update existing
		return parentRe.ReplaceAll(body, []byte("/Parent 2 0 R"))
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
func writeXRefAndTrailer(out *bytes.Buffer, offsets map[int]int) {
	// Find max object number
	maxObj := 0
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

	// Object 0 is always free
	out.WriteString("0000000000 65535 f\r\n")

	// Write entries for objects 1 to maxObj
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			xrefBuf = xrefBuf[:0]
			// Format as 10-digit zero-padded number
			offStr := strconv.FormatInt(int64(off), 10)
			for j := 0; j < 10-len(offStr); j++ {
				xrefBuf = append(xrefBuf, '0')
			}
			xrefBuf = append(xrefBuf, offStr...)
			xrefBuf = append(xrefBuf, " 00000 n\r\n"...)
			out.Write(xrefBuf)
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
	trailerRe := regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`)
	matches := trailerRe.FindAllSubmatch(data, -1)
	for _, m := range matches {
		if bytes.Contains(m[1], []byte("/Encrypt")) {
			return true
		}
	}
	return bytes.Contains(data, []byte("/Encrypt"))
}
