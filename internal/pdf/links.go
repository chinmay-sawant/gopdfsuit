package pdf

import (
	"fmt"
	"strings"
)

// LinkAnnotation represents a hyperlink annotation in a PDF
type LinkAnnotation struct {
	Rect      [4]float64 // [x1, y1, x2, y2] - annotation rectangle
	URI       string     // External URL (for /URI action)
	Dest      string     // Internal destination name (for /GoTo action)
	PageIndex int        // Target page index for internal links (0-based)
	DestY     float64    // Y coordinate on target page
}

// CreateLinkAnnotation creates a PDF link annotation object with PDF/UA-2 structure
// For external links (URLs), it creates a /URI action
// For internal links (bookmarks), it creates a /GoTo action with named destination
// Returns the annotation object ID
func CreateLinkAnnotation(annot LinkAnnotation, pageManager *PageManager) int {
	var annotDict strings.Builder

	// Get the StructParent index for this annotation
	// This links the annotation to the ParentTree for PDF/UA-2
	structParentIdx := pageManager.GetNextAnnotStructParent()

	annotDict.WriteString("<< /Type /Annot /Subtype /Link")
	annotDict.WriteString(fmt.Sprintf(" /Rect [%s %s %s %s]",
		fmtNum(annot.Rect[0]), fmtNum(annot.Rect[1]),
		fmtNum(annot.Rect[2]), fmtNum(annot.Rect[3])))

	// Border style - no visible border (0 0 0 means no border)
	annotDict.WriteString(" /Border [0 0 0]")

	// Highlight mode - invert when clicked
	annotDict.WriteString(" /H /I")

	// PDF/A-4 compliance: F key is required
	// Flag 4 = Print. This ensures the annotation is considered printable,
	// satisfying the requirement that all non-popup annotations must have an F key.
	annotDict.WriteString(" /F 4")

	// PDF/UA-2: StructParent links annotation to structure tree
	annotDict.WriteString(fmt.Sprintf(" /StructParent %d", structParentIdx))

	// Add action based on link type
	if annot.URI != "" {
		// External URL - use URI action
		annotDict.WriteString(fmt.Sprintf(" /A << /Type /Action /S /URI /URI (%s) >>",
			escapeText(annot.URI)))
	} else if annot.Dest != "" {
		// Internal link - use named destination
		// PDF/UA-2: Try to use structure destination if available
		if dest, ok := pageManager.NamedDests[annot.Dest]; ok && dest.StructElemID > 0 {
			// We need to construct the /D array for the fallback view
			// Find object ID for the destination page
			pageObjID := 3 // Default
			if dest.PageIndex < len(pageManager.Pages) {
				pageObjID = pageManager.Pages[dest.PageIndex]
			}
			// PDF 2.0 structure destination format: /SD [structElemRef destType params...]
			annotDict.WriteString(fmt.Sprintf(" /A << /S /GoTo /D [%d 0 R /XYZ null %s null] /SD [%d 0 R /XYZ null %s null] >>",
				pageObjID, fmtNum(dest.Y), dest.StructElemID, fmtNum(dest.Y)))
		} else {
			annotDict.WriteString(fmt.Sprintf(" /Dest (%s)", escapeText(annot.Dest)))
		}
	} else if annot.PageIndex >= 0 {
		// Internal link with explicit page destination
		// Format: [pageRef /XYZ left top zoom]
		// XYZ = position at (left, top) with zoom factor
		pageObjID := 3 + annot.PageIndex // Pages start at object 3
		annotDict.WriteString(fmt.Sprintf(" /Dest [%d 0 R /XYZ null %s null]",
			pageObjID, fmtNum(annot.DestY)))
	}

	annotDict.WriteString(" >>")

	objID := pageManager.AddExtraObject(annotDict.String())

	// PDF/UA-2: Create Link structure element that references this annotation
	// The Link element contains an OBJR (object reference) kid pointing to the annotation
	pageManager.AddLinkStructureElement(objID, structParentIdx)

	return objID
}

// ParseLink parses a link string and determines if it's external or internal
// External links start with http://, https://, mailto:, etc.
// Internal links start with # followed by a bookmark name
func ParseLink(link string) (isExternal bool, uri string, dest string) {
	link = strings.TrimSpace(link)
	if link == "" {
		return false, "", ""
	}

	// Check for internal bookmark link (starts with #)
	if strings.HasPrefix(link, "#") {
		return false, "", strings.TrimPrefix(link, "#")
	}

	// Check for common URL schemes
	lowerLink := strings.ToLower(link)
	if strings.HasPrefix(lowerLink, "http://") ||
		strings.HasPrefix(lowerLink, "https://") ||
		strings.HasPrefix(lowerLink, "mailto:") ||
		strings.HasPrefix(lowerLink, "tel:") ||
		strings.HasPrefix(lowerLink, "ftp://") {
		return true, link, ""
	}

	// Default: treat as external URL (add https:// if no scheme)
	if !strings.Contains(link, "://") && strings.Contains(link, ".") {
		return true, "https://" + link, ""
	}

	// Assume internal destination if doesn't look like URL
	return false, "", link
}

// DrawCellLink creates a link annotation for a cell if it has a link
// Returns the annotation object ID, or 0 if no link
func DrawCellLink(link string, cellX, cellY, cellWidth, cellHeight float64, pageManager *PageManager) int {
	if link == "" {
		return 0
	}

	isExternal, uri, dest := ParseLink(link)

	annot := LinkAnnotation{
		Rect: [4]float64{
			cellX,
			cellY,
			cellX + cellWidth,
			cellY + cellHeight,
		},
		PageIndex: -1, // Not using explicit page index
	}

	if isExternal {
		annot.URI = uri
	} else {
		annot.Dest = dest
	}

	objID := CreateLinkAnnotation(annot, pageManager)
	pageManager.AddAnnotation(objID)

	return objID
}
