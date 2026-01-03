package merge

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

// MergePDFs merges multiple PDF files into one
// It properly handles form fields, widgets, appearance streams, and various PDF versions
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

	// Extract pages from Pages tree
	fc.Pages = extractPagesFromTree(data, fc.Objects)

	// Extract form fields and annotation dependencies
	ExtractFormFields(fc)

	return fc
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
	fmt.Sscanf(rootRef, "%d", &rootNum)

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
// ensuring annotation dependencies are included
func collectObjectsWithDependencies(fc *FileContext) []int {
	included := make(map[int]bool)
	var result []int

	// Add all objects in numeric order
	for i := 1; i <= fc.MaxObj; i++ {
		if _, exists := fc.Objects[i]; exists {
			if !included[i] {
				result = append(result, i)
				included[i] = true
			}
		}
	}

	// Ensure all AP dependencies are included
	for _, deps := range fc.APDeps {
		for _, dep := range deps {
			if !included[dep] {
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
		for i, f := range fields {
			if i > 0 {
				out.WriteString(" ")
			}
			out.WriteString(fmt.Sprintf("%d 0 R", f))
		}
		out.WriteString("] /NeedAppearances true >>")
	}

	out.WriteString(" >>\nendobj\n")
}

// writePages writes the Pages object
func writePages(out *bytes.Buffer, pages []int) {
	out.WriteString("2 0 obj\n<< /Type /Pages /Kids [")
	for i, p := range pages {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(fmt.Sprintf("%d 0 R", p))
	}
	out.WriteString(fmt.Sprintf("] /Count %d >>\nendobj\n", len(pages)))
}

// writeObject writes a single PDF object
func writeObject(out *bytes.Buffer, num int, body []byte) {
	out.WriteString(fmt.Sprintf("%d 0 obj", num))

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
	out.WriteString(fmt.Sprintf("xref\n0 %d\n", maxObj+1))

	// Object 0 is always free
	out.WriteString("0000000000 65535 f\r\n")

	// Write entries for objects 1 to maxObj
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			out.WriteString(fmt.Sprintf("%010d 00000 n\r\n", off))
		} else {
			out.WriteString("0000000000 65535 f\r\n")
		}
	}

	// Trailer
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		maxObj+1, xrefStart))
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
