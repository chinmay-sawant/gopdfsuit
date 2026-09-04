package pdf

import "strings"

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
// Returns the annotation object ID. Emission lives in DestinationStore so
// link annots and their structure elements share one home with named dests.
func CreateLinkAnnotation(annot LinkAnnotation, pageManager *PageManager) int {
	return NewDestinationStore(pageManager, nil).CreateLinkAnnotation(annot)
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
	if after, ok := strings.CutPrefix(link, "#"); ok {
		return false, "", after
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
