package merge

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
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
	partsRe := regexp.MustCompile(`\s*,\s*`)
	parts := partsRe.Split(spec, -1)
	set := make(map[int]bool)
	rngRe := regexp.MustCompile(`^(\d+)-(\d+)$`)
	numRe := regexp.MustCompile(`^\d+$`)

	for _, p := range parts {
		if p == "" {
			continue
		}
		if rngRe.MatchString(p) {
			m := rngRe.FindStringSubmatch(p)
			a, _ := strconv.Atoi(m[1])
			b, _ := strconv.Atoi(m[2])
			if a < 1 || b < a {
				return nil, fmt.Errorf("invalid range: %s", p)
			}
			if totalPages > 0 && a > totalPages {
				return nil, fmt.Errorf("invalid range: %s", p)
			}
			if totalPages > 0 && b > totalPages {
				b = totalPages
			}
			for i := a; i <= b; i++ {
				set[i] = true
			}
		} else if numRe.MatchString(p) {
			n, _ := strconv.Atoi(p)
			if n < 1 || (totalPages > 0 && n > totalPages) {
				return nil, fmt.Errorf("invalid page: %s", p)
			}
			set[n] = true
		} else {
			return nil, fmt.Errorf("invalid token: %s", p)
		}
	}

	var pages []int
	for k := range set {
		pages = append(pages, k)
	}
	sort.Ints(pages)
	return pages, nil
}

// SplitPDF splits a single PDF file into one or more PDFs according to spec.
func SplitPDF(file []byte, spec SplitSpec) ([][]byte, error) {
	if len(file) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	if hasEncrypt(file) {
		return nil, fmt.Errorf("cannot split encrypted PDF")
	}

	fc := parseFile(file)
	if fc == nil {
		return nil, fmt.Errorf("invalid PDF")
	}

	totalPages := len(fc.Pages)
	if totalPages == 0 {
		return nil, fmt.Errorf("no pages found")
	}

	// Build requested page list (map 1-based indexes to page object numbers)
	var requestedObjNums []int

	// explicit Pages
	for _, p := range spec.Pages {
		if p < 1 || p > totalPages {
			return nil, fmt.Errorf("page out of range: %d", p)
		}
		requestedObjNums = append(requestedObjNums, fc.Pages[p-1])
	}

	// ranges
	for _, r := range spec.Ranges {
		if r[0] < 1 || r[1] < r[0] || r[1] > totalPages {
			return nil, fmt.Errorf("invalid range: %v", r)
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
	seen := make(map[int]bool)
	var orderedPages []int
	for _, obj := range requestedObjNums {
		if !seen[obj] {
			orderedPages = append(orderedPages, obj)
			seen[obj] = true
		}
	}

	// chunk according to MaxPerFile
	var groups [][]int
	if spec.MaxPerFile > 0 {
		for i := 0; i < len(orderedPages); i += spec.MaxPerFile {
			end := i + spec.MaxPerFile
			if end > len(orderedPages) {
				end = len(orderedPages)
			}
			groups = append(groups, orderedPages[i:end])
		}
	} else {
		groups = append(groups, orderedPages)
	}

	var outputs [][]byte
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
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)
	included := make(map[int]bool)
	var stack []int
	for _, p := range pageObjs {
		included[p] = true
		stack = append(stack, p)
	}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		body, ok := fc.Objects[n]
		if !ok {
			continue
		}
		// include AP deps if any
		if deps, ok := fc.APDeps[n]; ok {
			for _, d := range deps {
				if !included[d] && !isExcludedForSplit(fc, d) {
					included[d] = true
					stack = append(stack, d)
				}
			}
		}
		// find numeric refs in body (outside streams)
		for _, m := range refRe.FindAllSubmatch(body, -1) {
			refNum, _ := strconv.Atoi(string(m[1]))
			if refNum == 0 {
				continue
			}
			if isExcludedForSplit(fc, refNum) {
				continue
			}
			if !included[refNum] {
				included[refNum] = true
				stack = append(stack, refNum)
			}
		}
	}

	// build ordered list of object numbers to write
	var objs []int
	for i := 1; i <= fc.MaxObj; i++ {
		if _, ok := fc.Objects[i]; ok {
			if included[i] && !isExcludedForSplit(fc, i) {
				objs = append(objs, i)
			}
		}
	}

	// prepare merge context and header (use original file version)
	ctx := NewMergeContext()
	ctx.HighestVersion = DetectPDFVersion(originalFile)
	ctx.Output.WriteString(fmt.Sprintf("%%PDF-%s\n%%\xe2\xe3\xcf\xd3\n", ctx.HighestVersion))

	// remap offset: reserve 1 for Catalog and 2 for Pages
	offset := 2

	type appendedObj struct {
		num  int
		body []byte
	}
	var appended []appendedObj
	var mergedPages []int
	fieldSet := make(map[int]bool)
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
		if included[fn] && !isExcludedForSplit(fc, fn) {
			remapped := offset + fn
			if !fieldSet[remapped] {
				mergedFields = append(mergedFields, remapped)
				fieldSet[remapped] = true
			}
		}
	}

	// write Catalog and Pages
	ctx.Offsets[1] = ctx.Output.Len()
	writeCatalog(&ctx.Output, mergedFields)
	ctx.Offsets[2] = ctx.Output.Len()
	writePages(&ctx.Output, mergedPages)

	// write appended objects in numeric order
	for _, obj := range appended {
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
	if body, ok := fc.Objects[objNum]; ok {
		if IsPagesTreeObject(body) {
			return true
		}
	}
	return false
}
