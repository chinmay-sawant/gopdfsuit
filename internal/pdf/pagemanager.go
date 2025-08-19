package pdf

import "bytes"

// PageManager handles multi-page document generation
type PageManager struct {
	Pages            []int   // List of page object IDs
	CurrentPageIndex int     // Current page index (0-based)
	CurrentYPos      float64 // Current Y position on page
	PageDimensions   PageDimensions
	ContentStreams   []bytes.Buffer // Content for each page
}

// NewPageManager creates a new page manager with initial page
func NewPageManager(pageDims PageDimensions) *PageManager {
	pm := &PageManager{
		Pages:            []int{3}, // First page starts at object 3
		CurrentPageIndex: 0,        // Start with first page
		CurrentYPos:      pageDims.Height - margin,
		PageDimensions:   pageDims,
		ContentStreams:   make([]bytes.Buffer, 1),
	}
	return pm
}

// AddNewPage creates a new page when current page is full
func (pm *PageManager) AddNewPage() {
	// Calculate next page object ID
	nextPageID := 3 + len(pm.Pages) // Sequential page IDs starting from 3
	pm.Pages = append(pm.Pages, nextPageID)
	pm.CurrentPageIndex = len(pm.Pages) - 1 // Move to new page
	pm.CurrentYPos = pm.PageDimensions.Height - margin
	pm.ContentStreams = append(pm.ContentStreams, bytes.Buffer{})
}

// CheckPageBreak determines if a new page is needed based on required height
func (pm *PageManager) CheckPageBreak(requiredHeight float64) bool {
	return pm.CurrentYPos-requiredHeight < margin
}

// GetCurrentContentStream returns the current page's content stream
func (pm *PageManager) GetCurrentContentStream() *bytes.Buffer {
	return &pm.ContentStreams[pm.CurrentPageIndex]
}

// GetCurrentPageID returns the current page object ID
func (pm *PageManager) GetCurrentPageID() int {
	return pm.Pages[pm.CurrentPageIndex]
}
