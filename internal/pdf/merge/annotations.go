package merge

import (
	"errors"
	"regexp"
	"strconv"
)

// Package-level regexes (PERF-1): avoid MustCompile inside recursive/hot paths.
var (
	objRefRe         = regexp.MustCompile(`(\d+)\s+\d+\s+R`)
	objGenRefRe      = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	annotsArrayRe    = regexp.MustCompile(`/Annots\s*\[(.*?)\]`)
	annotsRefRe      = regexp.MustCompile(`/Annots\s+(\d+)\s+\d+\s+R`)
	annotsArrayUpRe  = regexp.MustCompile(`(/Annots\s*\[)([^\]]*?)(\])`)
	apDictRe         = regexp.MustCompile(`(?s)/AP\s*<<(.+?)>>`)
	resourcesRe      = regexp.MustCompile(`(?s)/Resources\s*<<(.+?)>>`)
	acroFormRefRe    = regexp.MustCompile(`/AcroForm\s+(\d+)\s+\d+\s+R`)
	acroFormInlineRe = regexp.MustCompile(`(?s)/AcroForm\s*<<(.+?)>>`)
	fieldsArrayRe    = regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	fieldsRefRe      = regexp.MustCompile(`/Fields\s+(\d+)\s+\d+\s+R`)
	kidsArrayRe      = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	rootRefRe        = regexp.MustCompile(`/Root\s+(\d+\s+\d+)\s+R`)
	wsSplitRe        = regexp.MustCompile(`\s+`)
)

// Annotation and form field handling for PDF merge

// ExtractAnnotationsFromPage extracts annotation object references from a page object
// Returns a list of annotation object numbers
func ExtractAnnotationsFromPage(pageBody []byte, objMap [][]byte) []int {
	var annots []int

	// Try inline array format: /Annots [...]
	if match := annotsArrayRe.FindSubmatch(pageBody); match != nil {
		for _, ref := range objRefRe.FindAllSubmatch(match[1], -1) {
			if num, err := strconv.Atoi(string(ref[1])); err == nil {
				annots = append(annots, num)
			}
		}
		return annots
	}

	// Try indirect reference format: /Annots N 0 R
	if match := annotsRefRe.FindSubmatch(pageBody); match != nil {
		if annotsObjNum, err := strconv.Atoi(string(match[1])); err == nil {
			if annotsObjNum < len(objMap) && objMap[annotsObjNum] != nil {
				annotsBody := objMap[annotsObjNum]
				// The annotations object should be an array
				for _, ref := range objRefRe.FindAllSubmatch(annotsBody, -1) {
					if num, err := strconv.Atoi(string(ref[1])); err == nil {
						annots = append(annots, num)
					}
				}
			}
		}
	}

	return annots
}

// ExtractAPDependencies extracts appearance stream dependencies from a widget
// These are XObject references in /AP << /N ... /D ... /R ... >>
func ExtractAPDependencies(widgetBody []byte, objMap [][]byte) []int {
	var deps []int
	seen := make(map[int]bool, 8)

	// Find /AP dictionary - handles both simple and complex cases
	// Simple: /AP << /N 123 0 R >>
	// Complex: /AP << /N << /Yes 123 0 R /Off 124 0 R >> >>
	if match := apDictRe.FindSubmatch(widgetBody); match != nil {
		apContent := match[1]

		// Extract all references from the AP dictionary
		for _, ref := range objRefRe.FindAllSubmatch(apContent, -1) {
			if num, err := strconv.Atoi(string(ref[1])); err == nil && !seen[num] {
				deps = append(deps, num)
				seen[num] = true

				// Check if this object itself has nested references (XObject resources)
				if num < len(objMap) && objMap[num] != nil {
					objBody := objMap[num]
					for _, nestedRef := range objRefRe.FindAllSubmatch(objBody, -1) {
						if nestedNum, err := strconv.Atoi(string(nestedRef[1])); err == nil && !seen[nestedNum] {
							deps = append(deps, nestedNum)
							seen[nestedNum] = true
						}
					}
				}
			}
		}
	}

	return deps
}

// ExtractFormFields extracts all form field objects from a PDF
// This includes widgets, their dependencies, and AcroForm fields
func ExtractFormFields(fc *FileContext) {
	n := fc.MaxObj + 1
	if n < 16 {
		n = 16
	}
	fieldSet := make([]bool, n)

	// Method 1: Find widgets via AcroForm in Catalog
	rootRef := findRootRef(fc.Data)
	if rootRef != "" {
		var rootNum int
		if err := parseObjRef(rootRef, &rootNum); err == nil {
			if rootNum < len(fc.Objects) && fc.Objects[rootNum] != nil {
				extractFromAcroForm(fc.Objects[rootNum], fc.Objects, &fc.FormFields, fieldSet, objRefRe)
			}
		}
	}

	// Method 2: Scan for Widget annotations directly
	for objNum, body := range fc.Objects {
		if body == nil {
			continue
		}
		if IsWidgetAnnotation(body) {
			if !fieldSet[objNum] {
				fc.FormFields = append(fc.FormFields, objNum)
				fieldSet[objNum] = true
			}

			// Extract appearance stream dependencies
			deps := ExtractAPDependencies(body, fc.Objects)
			if len(deps) > 0 {
				fc.APDeps[objNum] = deps
			}
		}

		// Also check for /FT (field type) marker
		if IsFormField(body) && !fieldSet[objNum] {
			fc.FormFields = append(fc.FormFields, objNum)
			fieldSet[objNum] = true
		}
	}

	// Method 3: Extract annotations from pages
	for objNum, body := range fc.Objects {
		if body == nil {
			continue
		}
		if IsPageObject(body) && !IsPagesTreeObject(body) {
			pageAnnots := ExtractAnnotationsFromPage(body, fc.Objects)
			if len(pageAnnots) > 0 {
				fc.Annots[objNum] = pageAnnots
			}

			for _, annotNum := range pageAnnots {
				if annotNum < len(fc.Objects) && fc.Objects[annotNum] != nil {
					annotBody := fc.Objects[annotNum]
					if IsWidgetAnnotation(annotBody) && !fieldSet[annotNum] {
						fc.FormFields = append(fc.FormFields, annotNum)
						fieldSet[annotNum] = true

						// Extract AP dependencies for this widget
						deps := ExtractAPDependencies(annotBody, fc.Objects)
						if len(deps) > 0 {
							fc.APDeps[annotNum] = deps
						}
					}
				}
			}
		}
	}
}

// extractFromAcroForm extracts field references from AcroForm
func extractFromAcroForm(catalogBody []byte, objMap [][]byte, fields *[]int, fieldSet []bool, refRe *regexp.Regexp) {
	// Try indirect AcroForm: /AcroForm N 0 R
	if match := acroFormRefRe.FindSubmatch(catalogBody); match != nil {
		if acroFormNum, err := strconv.Atoi(string(match[1])); err == nil {
			if acroFormNum < len(objMap) && objMap[acroFormNum] != nil {
				extractFieldsArray(objMap[acroFormNum], objMap, fields, fieldSet, refRe)
			}
		}
	}

	// Try inline AcroForm: /AcroForm << ... >>
	if match := acroFormInlineRe.FindSubmatch(catalogBody); match != nil {
		extractFieldsArray(match[1], objMap, fields, fieldSet, refRe)
	}
}

// extractFieldsArray extracts fields from /Fields array
func extractFieldsArray(acroFormBody []byte, objMap [][]byte, fields *[]int, fieldSet []bool, refRe *regexp.Regexp) {
	// Inline array: /Fields [...]
	if match := fieldsArrayRe.FindSubmatch(acroFormBody); match != nil {
		for _, ref := range refRe.FindAllSubmatch(match[1], -1) {
			if fieldNum, err := strconv.Atoi(string(ref[1])); err == nil {
				addFieldRecursive(fieldNum, objMap, fields, fieldSet, refRe)
			}
		}
	}

	// Indirect array: /Fields N 0 R
	if match := fieldsRefRe.FindSubmatch(acroFormBody); match != nil {
		if fieldsObjNum, err := strconv.Atoi(string(match[1])); err == nil {
			if fieldsObjNum < len(objMap) && objMap[fieldsObjNum] != nil {
				fieldsBody := objMap[fieldsObjNum]
				for _, ref := range refRe.FindAllSubmatch(fieldsBody, -1) {
					if fieldNum, err := strconv.Atoi(string(ref[1])); err == nil {
						addFieldRecursive(fieldNum, objMap, fields, fieldSet, refRe)
					}
				}
			}
		}
	}
}

// addFieldRecursive adds a field and its children (hierarchical form fields)
func addFieldRecursive(fieldNum int, objMap [][]byte, fields *[]int, fieldSet []bool, refRe *regexp.Regexp) {
	if fieldNum < len(fieldSet) && fieldSet[fieldNum] {
		return
	}
	*fields = append(*fields, fieldNum)
	if fieldNum < len(fieldSet) {
		fieldSet[fieldNum] = true
	}

	// Check for /Kids in the field (hierarchical fields)
	if fieldNum < len(objMap) && objMap[fieldNum] != nil {
		fieldBody := objMap[fieldNum]
		if match := kidsArrayRe.FindSubmatch(fieldBody); match != nil {
			for _, ref := range refRe.FindAllSubmatch(match[1], -1) {
				if kidNum, err := strconv.Atoi(string(ref[1])); err == nil {
					addFieldRecursive(kidNum, objMap, fields, fieldSet, refRe)
				}
			}
		}
	}
}

// findRootRef finds the /Root reference in PDF trailer
func findRootRef(data []byte) string {
	if m := rootRefRe.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return ""
}

// parseObjRef parses "N G" or "N" format into object number
func parseObjRef(ref string, num *int) error {
	parts := wsSplitRe.Split(ref, -1)
	if len(parts) >= 1 {
		n, err := strconv.Atoi(parts[0])
		if err == nil {
			*num = n
			return nil
		}
		return err
	}
	return errors.New("invalid object reference: " + ref)
}

// UpdatePageAnnotations updates page annotation references with remapped object numbers
func UpdatePageAnnotations(pageBody []byte, offset int) []byte {
	// Find and update /Annots array
	pageBody = annotsArrayUpRe.ReplaceAllFunc(pageBody, func(match []byte) []byte {
		parts := annotsArrayUpRe.FindSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		prefix := parts[1]
		content := parts[2]
		suffix := parts[3]

		// Replace references in content
		newContent := objGenRefRe.ReplaceAllFunc(content, func(match []byte) []byte {
			sm := objGenRefRe.FindSubmatch(match)
			if len(sm) < 3 {
				return match
			}
			on, _ := strconv.Atoi(string(sm[1]))
			gen := sm[2]
			var nbuf [20]byte
			num := strconv.AppendInt(nbuf[:0], int64(offset+on), 10)
			var refBuf [64]byte
			ref := refBuf[:0]
			ref = append(ref, num...)
			ref = append(ref, ' ')
			ref = append(ref, gen...)
			ref = append(ref, ' ', 'R')
			return ref
		})

		result := append(append(append([]byte(nil), prefix...), newContent...), suffix...)
		return result
	})

	return pageBody
}

// CollectAllDependencies collects all objects that a widget depends on
// This includes appearance streams and any nested references
func CollectAllDependencies(widgetNum int, objMap [][]byte) []int {
	var deps []int
	seen := make([]bool, len(objMap))
	if widgetNum < len(seen) {
		seen[widgetNum] = true // Don't include the widget itself
	}

	if widgetNum < len(objMap) && objMap[widgetNum] != nil {
		collectDepsRecursive(objMap[widgetNum], objMap, &deps, seen)
	}

	return deps
}

// collectDepsRecursive recursively collects dependencies
func collectDepsRecursive(body []byte, objMap [][]byte, deps *[]int, seen []bool) {
	// Only look in /AP dictionary to avoid false positives
	apMatch := apDictRe.FindSubmatch(body)
	if apMatch == nil {
		return
	}

	for _, ref := range objRefRe.FindAllSubmatch(apMatch[1], -1) {
		num, err := strconv.Atoi(string(ref[1]))
		if err != nil || (num < len(seen) && seen[num]) {
			continue
		}
		if num < len(seen) {
			seen[num] = true
		}
		*deps = append(*deps, num)

		// Recursively check this object (for nested resources)
		if num < len(objMap) && objMap[num] != nil {
			objBody := objMap[num]
			// For XObjects, also look at /Resources (resourcesRe hoisted package-level)
			if resMatch := resourcesRe.FindSubmatch(objBody); resMatch != nil {
				for _, nestedRef := range objRefRe.FindAllSubmatch(resMatch[1], -1) {
					nestedNum, err := strconv.Atoi(string(nestedRef[1]))
					if err != nil || (nestedNum < len(seen) && seen[nestedNum]) {
						continue
					}
					if nestedNum < len(seen) {
						seen[nestedNum] = true
					}
					*deps = append(*deps, nestedNum)
				}
			}
		}
	}
}
