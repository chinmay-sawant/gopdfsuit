package merge

import (
	"fmt"
	"regexp"
	"strconv"
)

// Annotation and form field handling for PDF merge

// ExtractAnnotationsFromPage extracts annotation object references from a page object
// Returns a list of annotation object numbers
func ExtractAnnotationsFromPage(pageBody []byte, objMap map[int][]byte) []int {
	var annots []int
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)

	// Try inline array format: /Annots [...]
	annotsArrayRe := regexp.MustCompile(`/Annots\s*\[(.*?)\]`)
	if match := annotsArrayRe.FindSubmatch(pageBody); match != nil {
		for _, ref := range refRe.FindAllSubmatch(match[1], -1) {
			if num, err := strconv.Atoi(string(ref[1])); err == nil {
				annots = append(annots, num)
			}
		}
		return annots
	}

	// Try indirect reference format: /Annots N 0 R
	annotsRefRe := regexp.MustCompile(`/Annots\s+(\d+)\s+\d+\s+R`)
	if match := annotsRefRe.FindSubmatch(pageBody); match != nil {
		if annotsObjNum, err := strconv.Atoi(string(match[1])); err == nil {
			if annotsBody, exists := objMap[annotsObjNum]; exists {
				// The annotations object should be an array
				for _, ref := range refRe.FindAllSubmatch(annotsBody, -1) {
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
func ExtractAPDependencies(widgetBody []byte, objMap map[int][]byte) []int {
	var deps []int
	seen := make(map[int]bool)
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)

	// Find /AP dictionary - handles both simple and complex cases
	// Simple: /AP << /N 123 0 R >>
	// Complex: /AP << /N << /Yes 123 0 R /Off 124 0 R >> >>
	apRe := regexp.MustCompile(`(?s)/AP\s*<<(.+?)>>`)

	if match := apRe.FindSubmatch(widgetBody); match != nil {
		apContent := match[1]

		// Extract all references from the AP dictionary
		for _, ref := range refRe.FindAllSubmatch(apContent, -1) {
			if num, err := strconv.Atoi(string(ref[1])); err == nil && !seen[num] {
				deps = append(deps, num)
				seen[num] = true

				// Check if this object itself has nested references (XObject resources)
				if objBody, exists := objMap[num]; exists {
					for _, nestedRef := range refRe.FindAllSubmatch(objBody, -1) {
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
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)
	fieldSet := make(map[int]bool)

	// Method 1: Find widgets via AcroForm in Catalog
	rootRef := findRootRef(fc.Data)
	if rootRef != "" {
		var rootNum int
		if err := parseObjRef(rootRef, &rootNum); err == nil {
			if rootBody, exists := fc.Objects[rootNum]; exists {
				extractFromAcroForm(rootBody, fc.Objects, &fc.FormFields, fieldSet, refRe)
			}
		}
	}

	// Method 2: Scan for Widget annotations directly
	for objNum, body := range fc.Objects {
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
		if IsPageObject(body) && !IsPagesTreeObject(body) {
			pageAnnots := ExtractAnnotationsFromPage(body, fc.Objects)
			if len(pageAnnots) > 0 {
				fc.Annots[objNum] = pageAnnots
			}

			for _, annotNum := range pageAnnots {
				if annotBody, exists := fc.Objects[annotNum]; exists {
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
func extractFromAcroForm(catalogBody []byte, objMap map[int][]byte, fields *[]int, fieldSet map[int]bool, refRe *regexp.Regexp) {
	// Try indirect AcroForm: /AcroForm N 0 R
	acroFormRefRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s+\d+\s+R`)
	if match := acroFormRefRe.FindSubmatch(catalogBody); match != nil {
		if acroFormNum, err := strconv.Atoi(string(match[1])); err == nil {
			if acroFormBody, exists := objMap[acroFormNum]; exists {
				extractFieldsArray(acroFormBody, objMap, fields, fieldSet, refRe)
			}
		}
	}

	// Try inline AcroForm: /AcroForm << ... >>
	acroFormInlineRe := regexp.MustCompile(`(?s)/AcroForm\s*<<(.+?)>>`)
	if match := acroFormInlineRe.FindSubmatch(catalogBody); match != nil {
		extractFieldsArray(match[1], objMap, fields, fieldSet, refRe)
	}
}

// extractFieldsArray extracts fields from /Fields array
func extractFieldsArray(acroFormBody []byte, objMap map[int][]byte, fields *[]int, fieldSet map[int]bool, refRe *regexp.Regexp) {
	// Inline array: /Fields [...]
	fieldsArrayRe := regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	if match := fieldsArrayRe.FindSubmatch(acroFormBody); match != nil {
		for _, ref := range refRe.FindAllSubmatch(match[1], -1) {
			if fieldNum, err := strconv.Atoi(string(ref[1])); err == nil {
				addFieldRecursive(fieldNum, objMap, fields, fieldSet, refRe)
			}
		}
	}

	// Indirect array: /Fields N 0 R
	fieldsRefRe := regexp.MustCompile(`/Fields\s+(\d+)\s+\d+\s+R`)
	if match := fieldsRefRe.FindSubmatch(acroFormBody); match != nil {
		if fieldsObjNum, err := strconv.Atoi(string(match[1])); err == nil {
			if fieldsBody, exists := objMap[fieldsObjNum]; exists {
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
func addFieldRecursive(fieldNum int, objMap map[int][]byte, fields *[]int, fieldSet map[int]bool, refRe *regexp.Regexp) {
	if fieldSet[fieldNum] {
		return
	}
	*fields = append(*fields, fieldNum)
	fieldSet[fieldNum] = true

	// Check for /Kids in the field (hierarchical fields)
	if fieldBody, exists := objMap[fieldNum]; exists {
		kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
		if match := kidsRe.FindSubmatch(fieldBody); match != nil {
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
	rootRe := regexp.MustCompile(`/Root\s+(\d+\s+\d+)\s+R`)
	if m := rootRe.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return ""
}

// parseObjRef parses "N G" or "N" format into object number
func parseObjRef(ref string, num *int) error {
	parts := regexp.MustCompile(`\s+`).Split(ref, -1)
	if len(parts) >= 1 {
		n, err := strconv.Atoi(parts[0])
		if err == nil {
			*num = n
			return nil
		}
		return err
	}
	return fmt.Errorf("invalid object reference: %s", ref)
}

// UpdatePageAnnotations updates page annotation references with remapped object numbers
func UpdatePageAnnotations(pageBody []byte, offset int) []byte {
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)

	// Find and update /Annots array
	annotsArrayRe := regexp.MustCompile(`(/Annots\s*\[)([^\]]*?)(\])`)
	pageBody = annotsArrayRe.ReplaceAllFunc(pageBody, func(match []byte) []byte {
		parts := annotsArrayRe.FindSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		prefix := parts[1]
		content := parts[2]
		suffix := parts[3]

		// Replace references in content
		newContent := refRe.ReplaceAllFunc(content, func(ref []byte) []byte {
			sm := refRe.FindSubmatch(ref)
			if len(sm) < 3 {
				return ref
			}
			on, _ := strconv.Atoi(string(sm[1]))
			gen := string(sm[2])
			return []byte(strconv.Itoa(offset+on) + " " + gen + " R")
		})

		result := make([]byte, 0, len(prefix)+len(newContent)+len(suffix))
		result = append(result, prefix...)
		result = append(result, newContent...)
		result = append(result, suffix...)
		return result
	})

	return pageBody
}

// CollectAllDependencies collects all objects that a widget depends on
// This includes appearance streams and any nested references
func CollectAllDependencies(widgetNum int, objMap map[int][]byte) []int {
	var deps []int
	seen := make(map[int]bool)
	seen[widgetNum] = true // Don't include the widget itself

	if widgetBody, exists := objMap[widgetNum]; exists {
		collectDepsRecursive(widgetBody, objMap, &deps, seen)
	}

	return deps
}

// collectDepsRecursive recursively collects dependencies
func collectDepsRecursive(body []byte, objMap map[int][]byte, deps *[]int, seen map[int]bool) {
	refRe := regexp.MustCompile(`(\d+)\s+\d+\s+R`)

	// Only look in /AP dictionary to avoid false positives
	apRe := regexp.MustCompile(`(?s)/AP\s*<<(.+?)>>`)
	apMatch := apRe.FindSubmatch(body)
	if apMatch == nil {
		return
	}

	for _, ref := range refRe.FindAllSubmatch(apMatch[1], -1) {
		num, err := strconv.Atoi(string(ref[1]))
		if err != nil || seen[num] {
			continue
		}
		seen[num] = true
		*deps = append(*deps, num)

		// Recursively check this object (for nested resources)
		if objBody, exists := objMap[num]; exists {
			// For XObjects, also look at /Resources
			resourcesRe := regexp.MustCompile(`(?s)/Resources\s*<<(.+?)>>`)
			if resMatch := resourcesRe.FindSubmatch(objBody); resMatch != nil {
				for _, nestedRef := range refRe.FindAllSubmatch(resMatch[1], -1) {
					nestedNum, err := strconv.Atoi(string(nestedRef[1]))
					if err != nil || seen[nestedNum] {
						continue
					}
					seen[nestedNum] = true
					*deps = append(*deps, nestedNum)
				}
			}
		}
	}
}
