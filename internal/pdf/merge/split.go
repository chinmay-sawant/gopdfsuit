package merge

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Package-level regexes for page-spec parsing (PERF-1).
var (
	commaPartsRe = regexp.MustCompile(`\s*,\s*`)
	rangeSpecRe  = regexp.MustCompile(`^(\d+)-(\d+)$`)
	numSpecRe    = regexp.MustCompile(`^\d+$`)
)

// SplitSpec defines split criteria
type SplitSpec struct {
	Pages      []int // explicit pages (1-based)
	Ranges     [][2]int
	MaxPerFile int
}

// ParsePageSpec parses a simple spec string like "1-3,5,7-9" into a sorted slice of pages (1-based).
func ParsePageSpec(spec string, totalPages int) ([]int, error) {
	if spec == "" {
		return nil, nil
	}
	parts := commaPartsRe.Split(spec, -1)

	// Use slice when totalPages is known, map otherwise
	var set map[int]bool
	var setSlice []bool
	useSlice := totalPages > 0
	if useSlice {
		setSlice = make([]bool, totalPages+1)
	} else {
		set = make(map[int]bool, len(parts))
	}

	for _, p := range parts {
		if p == "" {
			continue
		}
		if idx := strings.IndexByte(p, '-'); idx > 0 {
			a, err1 := strconv.Atoi(p[:idx])
			b, err2 := strconv.Atoi(p[idx+1:])
			if err1 != nil || err2 != nil || a < 1 || b < a {
				return nil, errors.New("invalid range: " + p)
			}
			if totalPages > 0 && a > totalPages {
				return nil, errors.New("invalid range: " + p)
			}
			if totalPages > 0 && b > totalPages {
				b = totalPages
			}
			for i := a; i <= b; i++ {
				if useSlice {
					setSlice[i] = true
				} else {
					set[i] = true
				}
			}
		} else if isAllDigits(p) {
			n, _ := strconv.Atoi(p)
			if n < 1 || (totalPages > 0 && n > totalPages) {
				return nil, errors.New("invalid page: " + p)
			}
			if useSlice {
				setSlice[n] = true
			} else {
				set[n] = true
			}
		} else {
			return nil, errors.New("invalid token: " + p)
		}
	}

	var pages []int
	if useSlice {
		pages = make([]int, 0, totalPages)
		for i := 1; i <= totalPages; i++ {
			if setSlice[i] {
				pages = append(pages, i)
			}
		}
	} else {
		pages = make([]int, 0, len(set))
		for k := range set {
			pages = append(pages, k)
		}
	}
	sort.Ints(pages)
	return pages, nil
}

// isAllDigits checks if a string consists entirely of decimal digits.
func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

// SplitPDF splits a single PDF file into one or more PDFs according to spec.
func SplitPDF(file []byte, spec SplitSpec) ([][]byte, error) {
	/*
		TODO: handle image XObjects (or other image data) in object streams that
		are being left over after split. For now, we just exclude object streams.
		I noticed this issue when tried splitting my lecture notes.
	*/

	if len(file) == 0 {
		return nil, errors.New("empty file")
	}
	if hasEncrypt(file) {
		return nil, errors.New("cannot split encrypted PDF")
	}

	fc := parseFile(file)
	if fc == nil {
		return nil, errors.New("invalid PDF")
	}

	totalPages := len(fc.Pages)
	if totalPages == 0 {
		return nil, errors.New("no pages found")
	}

	// Build requested page list (map 1-based indexes to page object numbers)
	// PERF-45: capacity hints for append-heavy page selection
	reqCap := len(spec.Pages)
	for _, r := range spec.Ranges {
		if r[1] >= r[0] {
			reqCap += r[1] - r[0] + 1
		}
	}
	if reqCap == 0 {
		reqCap = totalPages
	}
	requestedObjNums := make([]int, 0, reqCap)

	// explicit Pages
	for _, p := range spec.Pages {
		if p < 1 || p > totalPages {
			var pBuf [20]byte
			return nil, errors.New("page out of range: " + string(strconv.AppendInt(pBuf[:0], int64(p), 10)))
		}
		requestedObjNums = append(requestedObjNums, fc.Pages[p-1])
	}

	// ranges
	for _, r := range spec.Ranges {
		if r[0] < 1 || r[1] < r[0] || r[1] > totalPages {
			return nil, errors.New("invalid range")
		}
		for i := r[0]; i <= r[1]; i++ {
			requestedObjNums = append(requestedObjNums, fc.Pages[i-1])
		}
	}

	// if nothing requested, assume all pages
	if len(requestedObjNums) == 0 {
		requestedObjNums = append(requestedObjNums, fc.Pages...)
	}

	// dedupe while preserving document order
	seen := make(map[int]bool, len(requestedObjNums)) // PERF-192
	orderedPages := make([]int, 0, len(requestedObjNums))
	for _, obj := range requestedObjNums {
		if !seen[obj] {
			orderedPages = append(orderedPages, obj)
			seen[obj] = true
		}
	}

	// chunk according to MaxPerFile
	groupCap := 1
	if spec.MaxPerFile > 0 {
		groupCap = (len(orderedPages) + spec.MaxPerFile - 1) / spec.MaxPerFile
		if groupCap < 1 {
			groupCap = 1
		}
	}
	var groups [][]int
	if spec.MaxPerFile > 0 {
		n := (len(orderedPages) + spec.MaxPerFile - 1) / spec.MaxPerFile
		if n < 0 {
			n = 0
		}
		groups = make([][]int, 0, n)
		for i := 0; i < len(orderedPages); i += spec.MaxPerFile {
			end := i + spec.MaxPerFile
			if end > len(orderedPages) {
				end = len(orderedPages)
			}
			groups = append(groups, orderedPages[i:end])
		}
	} else {
		groups = [][]int{orderedPages}
	}

	outputs := make([][]byte, 0, len(groups))
	for _, grp := range groups {
		out, err := buildPDFFromPageObjs(fc, grp, file)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, out)
	}

	return outputs, nil
}

// buildPDFFromPageObjs builds a single PDF containing only the provided original page object numbers.
func buildPDFFromPageObjs(fc *FileContext, pageObjs []int, originalFile []byte) ([]byte, error) {
	// collect included objects via DFS starting from page objects
	n := fc.MaxObj + 1
	if n < 8 {
		n = 8
	}
	included := make([]bool, n)
	var stack []int
	for _, p := range pageObjs {
		if p < len(included) {
			included[p] = true
		}
		stack = append(stack, p)
	}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n >= len(fc.Objects) || fc.Objects[n] == nil {
			continue
		}
		body := fc.Objects[n]
		// include AP deps if any
		if deps, ok := fc.APDeps[n]; ok {
			for _, d := range deps {
				if d < len(included) && !included[d] && !isExcludedForSplit(fc, d) {
					included[d] = true
					stack = append(stack, d)
				}
			}
		}
		// find numeric refs in body (outside streams)
		for _, m := range objRefRe.FindAllSubmatch(body, -1) {
			refNum, _ := strconv.Atoi(string(m[1]))
			if refNum == 0 {
				continue
			}
			if isExcludedForSplit(fc, refNum) {
				continue
			}
			if refNum < len(included) && !included[refNum] {
				included[refNum] = true
				stack = append(stack, refNum)
			}
		}
	}

	// build ordered list of object numbers to write
	var objs []int
	for i := 1; i <= fc.MaxObj; i++ {
		if i < len(fc.Objects) && fc.Objects[i] != nil {
			if i < len(included) && included[i] && !isExcludedForSplit(fc, i) {
				objs = append(objs, i)
			}
		}
	}

	// prepare merge context and header (use original file version)
	ctx := NewMergeContext()
	ctx.HighestVersion = DetectPDFVersion(originalFile)
	ctx.Output.WriteString("%PDF-")
	ctx.Output.WriteString(ctx.HighestVersion)
	ctx.Output.WriteString("\n%\xe2\xe3\xcf\xd3\n")

	// remap offset: reserve 1 for Catalog and 2 for Pages
	offset := 2

	type appendedObj struct {
		num  int
		body []byte
	}
	var appended []appendedObj
	var mergedPages []int
	fieldSet := make([]bool, n)
	var mergedFields []int

	// collect remapped object bodies
	for _, origNum := range objs {
		body := fc.Objects[origNum]
		if body == nil {
			continue
		}
		newNum := offset + origNum
		newBody := ReplaceRefsOutsideStreams(body, offset)

		// If page leaf, record remapped page number
		if IsPageObject(newBody) && !IsPagesTreeObject(newBody) {
			mergedPages = append(mergedPages, newNum)
		}
		appended = append(appended, appendedObj{num: newNum, body: newBody})
	}

	// track form fields that are included
	for _, fn := range fc.FormFields {
		if fn < len(included) && included[fn] && !isExcludedForSplit(fc, fn) {
			remapped := offset + fn
			if remapped < len(fieldSet) && !fieldSet[remapped] {
				mergedFields = append(mergedFields, remapped)
				fieldSet[remapped] = true
			}
		}
	}

	// write Catalog and Pages
	ensureLen(&ctx.Offsets, 3)
	ctx.Offsets[1] = ctx.Output.Len()
	writeCatalog(&ctx.Output, mergedFields)
	ctx.Offsets[2] = ctx.Output.Len()
	writePages(&ctx.Output, mergedPages)

	// write appended objects in numeric order
	for _, obj := range appended {
		ensureLen(&ctx.Offsets, obj.num+1)
		ctx.Offsets[obj.num] = ctx.Output.Len()
		body := obj.body
		// ensure page objects have Parent -> 2 0 R
		if IsPageObject(body) && !IsPagesTreeObject(body) {
			body = updateParentRef(body)
		}
		writeObject(&ctx.Output, obj.num, body)
	}

	// write xref & trailer
	writeXRefAndTrailer(&ctx.Output, ctx.Offsets)

	return ctx.Output.Bytes(), nil
}

// isExcludedForSplit returns true for objects we must not copy into the new file
func isExcludedForSplit(fc *FileContext, objNum int) bool {
	if fc.OriginalCatalog > 0 && objNum == fc.OriginalCatalog {
		return true
	}
	if fc.OriginalPagesTree > 0 && objNum == fc.OriginalPagesTree {
		return true
	}
	for _, n := range fc.ObjectStreamNums {
		if objNum == n {
			return true
		}
	}
	// exclude Pages tree nodes (intermediate)
	if objNum < len(fc.Objects) && fc.Objects[objNum] != nil {
		if IsPagesTreeObject(fc.Objects[objNum]) {
			return true
		}
	}
	return false
}
