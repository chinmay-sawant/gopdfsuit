package merge

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/pdfobj"
)

// MaxMergeObjects caps the highest object number accepted from a single
// input file and the total object count of a merge. It rejects sparse
// files like `999999 0 obj` fast instead of blowing time/memory in the
// numeric-order collection and xref loops.
const MaxMergeObjects = 50000

// MaxMergeInputBytes caps a single input file accepted by MergePDFs.
const MaxMergeInputBytes = 32 << 20 // 32 MiB

const maxPDFTreeDepth = 1000

// MaxMergeTotalInputBytes caps the combined size of all merge inputs.
const MaxMergeTotalInputBytes = 128 << 20 // 128 MiB

// encryptRefRe matches a real /Encrypt entry: an indirect reference or an
// inline encryption dictionary. Bare "/Encrypt" text inside content streams
// or strings must not count.
var encryptRefRe = regexp.MustCompile(`/Encrypt\s*(\d+\s+\d+\s+R|<<)`)

// MergePDFs merges multiple PDF files into one
// It properly handles form fields, widgets, appearance streams, and various PDF versions
//
//nolint:revive // exported
func MergePDFs(files [][]byte) ([]byte, error) {
	if len(files) == 0 {
		return nil, errors.New("no PDF files provided")
	}
	var totalBytes uint64

	ctx := NewMergeContext()

	// Parse all input files
	var fileContexts []*FileContext
	for _, f := range files {
		if len(f) > MaxMergeInputBytes {
			return nil, fmt.Errorf("input PDF exceeds %d bytes", MaxMergeInputBytes)
		}
		totalBytes += uint64(len(f))
		if totalBytes > MaxMergeTotalInputBytes {
			return nil, fmt.Errorf("combined PDF inputs exceed %d bytes", MaxMergeTotalInputBytes)
		}
		if hasEncrypt(f) {
			return nil, errors.New("cannot merge encrypted PDF")
		}

		fc, err := parseFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse PDF: %w", err)
		}
		if fc == nil {
			continue
		}
		if fc.MaxObj > MaxMergeObjects || len(fc.Objects) > MaxMergeObjects {
			return nil, fmt.Errorf("input PDF exceeds %d objects", MaxMergeObjects)
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

	// Write PDF header
	ctx.Output.WriteString(fmt.Sprintf("%%PDF-%s\n%%\xe2\xe3\xcf\xd3\n", ctx.HighestVersion))

	// Process each file
	var appendedObjects []struct {
		num  int
		body []byte
	}

	for _, fc := range fileContexts {
		offset := ctx.CurrentMax

		if offset+fc.MaxObj+1 > MaxMergeObjects {
			return nil, fmt.Errorf("merged PDF would exceed %d objects", MaxMergeObjects)
		}

		// Collect all objects to process (including annotation dependencies)
		objectsToProcess := collectObjectsWithDependencies(fc)
		for _, pageNum := range fc.Pages {
			if _, exists := fc.Objects[pageNum]; exists {
				ctx.MergedPages = append(ctx.MergedPages, offset+pageNum)
			}
		}

		// Process objects maintaining order
		for _, origNum := range objectsToProcess {
			body, exists := fc.Objects[origNum]
			if !exists {
				continue
			}

			newNum := offset + origNum

			// Materialize inherited page properties before removing the source
			// page tree, then remap all references in the copied body.
			newBody, err := materializeInheritedPageProperties(origNum, body, fc.Objects)
			if err != nil {
				return nil, err
			}
			newBody = ReplaceRefsOutsideStreams(newBody, offset)

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
func parseFile(data []byte) (*FileContext, error) {
	fc := NewFileContext(data)

	// Find all objects
	boundaries := FindObjectBoundaries(data)
	if len(boundaries) == 0 {
		return nil, nil
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
	var err error
	fc.Pages, err = extractPagesFromTree(data, fc.Objects)
	if err != nil {
		return nil, err
	}

	// Find the original catalog and pages tree to exclude them
	fc.OriginalCatalog, fc.OriginalPagesTree = findCatalogAndPages(data, fc.Objects)

	// Extract form fields and annotation dependencies
	if err := ExtractFormFields(fc); err != nil {
		return nil, err
	}

	return fc, nil
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
func extractPagesFromTree(data []byte, objMap map[int][]byte) ([]int, error) {
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)

	rootRef := findRootRef(data)
	if rootRef == "" {
		return nil, nil
	}

	// Parse root object number
	var rootNum int
	if _, err := fmt.Sscanf(rootRef, "%d", &rootNum); err != nil {
		return nil, nil
	}

	rootBody, exists := objMap[rootNum]
	if !exists {
		return nil, nil
	}

	// Find /Pages reference
	pagesRe := regexp.MustCompile(`/Pages\s+(\d+)\s+\d+\s+R`)
	match := pagesRe.FindSubmatch(rootBody)
	if match == nil {
		return nil, nil
	}

	pagesNum, _ := strconv.Atoi(string(match[1]))
	return extractKidsRecursive(pagesNum, objMap, refRe, make(map[int]bool), 0)
}

// extractKidsRecursive extracts page numbers from /Kids array
func extractKidsRecursive(pageObjNum int, objMap map[int][]byte, refRe *regexp.Regexp, active map[int]bool, depth int) ([]int, error) {
	if depth > maxPDFTreeDepth {
		return nil, fmt.Errorf("page tree exceeds maximum depth %d", maxPDFTreeDepth)
	}
	if active[pageObjNum] {
		return nil, fmt.Errorf("page tree cycle at object %d", pageObjNum)
	}

	pageBody, ok := objMap[pageObjNum]
	if !ok || !IsPagesTreeObject(pageBody) {
		return []int{pageObjNum}, nil
	}

	active[pageObjNum] = true
	defer delete(active, pageObjNum)

	var pages []int
	kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)

	match := kidsRe.FindSubmatch(pageBody)
	if match == nil {
		return pages, nil
	}

	for _, r := range refRe.FindAllSubmatch(match[1], -1) {
		kidNum, _ := strconv.Atoi(string(r[1]))
		kidPages, err := extractKidsRecursive(kidNum, objMap, refRe, active, depth+1)
		if err != nil {
			return nil, err
		}
		pages = append(pages, kidPages...)
	}

	return pages, nil
}

var inheritablePageKeys = []string{
	"Resources",
	"MediaBox",
	"CropBox",
	"BleedBox",
	"TrimBox",
	"ArtBox",
	"BoxColorInfo",
	"Rotate",
	"UserUnit",
}

var indirectObjectRefRe = regexp.MustCompile(`^\d+\s+\d+\s+R`)

func materializeInheritedPageProperties(pageNum int, body []byte, objMap map[int][]byte) ([]byte, error) {
	values := make(map[string][]byte, len(inheritablePageKeys))
	active := make(map[int]bool)
	currentNum := pageNum
	currentBody := body

	for depth := 0; ; depth++ {
		if depth > maxPDFTreeDepth {
			return nil, fmt.Errorf("page parent chain exceeds maximum depth %d", maxPDFTreeDepth)
		}
		if active[currentNum] {
			return nil, fmt.Errorf("page parent cycle at object %d", currentNum)
		}
		active[currentNum] = true

		for _, key := range inheritablePageKeys {
			if _, found := values[key]; !found {
				if value := pdfKeyValue(dictPart(currentBody), key); value != nil {
					values[key] = value
				}
			}
		}

		parentValue := pdfKeyValue(dictPart(currentBody), "Parent")
		parentNum, ok := parseIndirectObjectNumber(parentValue)
		if !ok {
			break
		}
		parentBody, ok := objMap[parentNum]
		if !ok {
			break
		}
		currentNum = parentNum
		currentBody = parentBody
	}

	var additions bytes.Buffer
	for _, key := range inheritablePageKeys {
		if pdfKeyValue(dictPart(body), key) != nil {
			continue
		}
		if value := values[key]; value != nil {
			additions.WriteByte(' ')
			additions.WriteByte('/')
			additions.WriteString(key)
			additions.WriteByte(' ')
			additions.Write(value)
		}
	}
	if additions.Len() == 0 {
		return body, nil
	}

	dictStart := bytes.Index(body, []byte("<<"))
	if dictStart < 0 {
		return body, nil
	}
	var result bytes.Buffer
	result.Write(body[:dictStart+2])
	result.Write(additions.Bytes())
	result.Write(body[dictStart+2:])
	return result.Bytes(), nil
}

func pdfKeyValue(body []byte, key string) []byte {
	needle := []byte("/" + key)
	for search := 0; search < len(body); {
		rel := bytes.Index(body[search:], needle)
		if rel < 0 {
			return nil
		}
		keyStart := search + rel
		valueStart := keyStart + len(needle)
		if keyStart > 0 && !isWhitespace(body[keyStart-1]) && body[keyStart-1] != '<' {
			search = valueStart
			continue
		}
		for valueStart < len(body) && isWhitespace(body[valueStart]) {
			valueStart++
		}
		if valueStart >= len(body) {
			return nil
		}
		valueEnd := valueStart
		switch {
		case body[valueStart] == '<' && valueStart+1 < len(body) && body[valueStart+1] == '<':
			valueEnd = pdfobj.SkipDictionary(body, valueStart)
		case body[valueStart] == '[':
			valueEnd = pdfobj.SkipArray(body, valueStart)
		case body[valueStart] == '(':
			valueEnd = pdfobj.SkipStringLiteral(body, valueStart)
		case body[valueStart] == '<':
			valueEnd = pdfobj.SkipHexString(body, valueStart)
		default:
			if ref := indirectObjectRefRe.Find(body[valueStart:]); ref != nil {
				valueEnd += len(ref)
			} else {
				for valueEnd < len(body) && !isWhitespace(body[valueEnd]) && body[valueEnd] != '>' {
					valueEnd++
				}
			}
		}
		if valueEnd > valueStart && valueEnd <= len(body) {
			return body[valueStart:valueEnd]
		}
		search = valueStart + 1
	}
	return nil
}

func parseIndirectObjectNumber(value []byte) (int, bool) {
	if len(value) == 0 {
		return 0, false
	}
	var num int
	var gen int
	var marker string
	if _, err := fmt.Sscanf(string(value), "%d %d %1s", &num, &gen, &marker); err != nil || marker != "R" || num <= 0 {
		return 0, false
	}
	return num, true
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
	pdfobj.WriteDenseXRef(out, maxObj, func(id int) (int, bool) {
		off, ok := offsets[id]
		return off, ok
	}, nil, pdfobj.MergeStyle)

	// Trailer
	var xrefBuf []byte
	xrefBuf = append(xrefBuf, "trailer\n<< /Size "...)
	xrefBuf = strconv.AppendInt(xrefBuf, int64(maxObj+1), 10)
	xrefBuf = append(xrefBuf, " /Root 1 0 R >>\nstartxref\n"...)
	xrefBuf = strconv.AppendInt(xrefBuf, int64(xrefStart), 10)
	xrefBuf = append(xrefBuf, "\n%%%%EOF\n"...)
	out.Write(xrefBuf)
}

// hasEncrypt checks if the PDF declares document encryption: a real
// /Encrypt entry (indirect reference or inline dict) outside stream data.
// Plain "/Encrypt" text inside a content stream no longer counts.
func hasEncrypt(data []byte) bool {
	return encryptRefRe.Match(BytesWithoutStreams(data))
}

// BytesWithoutStreams returns data with raw stream contents blanked so
// keyword scans only see dictionary/trailer context, never stream bytes.
func BytesWithoutStreams(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	i := 0
	for i < len(data) {
		rel := FindStreamStart(out[i:])
		if rel == -1 {
			break
		}
		start := i + rel
		ptr := start + 6
		if ptr < len(out) && out[ptr] == '\r' {
			ptr++
		}
		if ptr < len(out) && out[ptr] == '\n' {
			ptr++
		}
		idx := bytes.Index(out[ptr:], []byte("endstream"))
		if idx == -1 {
			for j := ptr; j < len(out); j++ {
				out[j] = ' '
			}
			break
		}
		for j := ptr; j < ptr+idx; j++ {
			out[j] = ' '
		}
		i = ptr + idx + 9
	}
	return out
}
