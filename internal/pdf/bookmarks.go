package pdf

import (
	"bytes"

	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// GenerateBookmarks generates the outline hierarchy for the PDF
// It returns the object ID of the Outlines dictionary (the root)
func (pm *PageManager) GenerateBookmarks(bookmarks []models.Bookmark, xrefOffsets map[int]int, pdfBuffer *bytes.Buffer) int {
	if len(bookmarks) == 0 {
		return 0
	}

	// Reserve object ID for Outlines dictionary
	outlinesID := pm.NextObjectID
	pm.NextObjectID++

	// Recurse to generate items
	firstID, lastID, count := pm.generateBookmarkItems(bookmarks, outlinesID, xrefOffsets, pdfBuffer)

	// Write Outlines dictionary
	// Write Outlines dictionary
	xrefOffsets[outlinesID] = pdfBuffer.Len()
	var b []byte
	b = strconv.AppendInt(b, int64(outlinesID), 10)
	b = append(b, " 0 obj\n"...)
	pdfBuffer.Write(b)
	pdfBuffer.WriteString("<< /Type /Outlines")
	if firstID > 0 {
		b = b[:0]
		b = append(b, " /First "...)
		b = strconv.AppendInt(b, int64(firstID), 10)
		b = append(b, " 0 R"...)
		pdfBuffer.Write(b)
	}
	if lastID > 0 {
		b = b[:0]
		b = append(b, " /Last "...)
		b = strconv.AppendInt(b, int64(lastID), 10)
		b = append(b, " 0 R"...)
		pdfBuffer.Write(b)
	}
	b = b[:0]
	b = append(b, " /Count "...)
	b = strconv.AppendInt(b, int64(count), 10)
	b = append(b, " >>\nendobj\n"...)
	pdfBuffer.Write(b) // Count includes all visible descendants

	return outlinesID
}

// generateBookmarkItems processes a list of bookmarks and returns (firstID, lastID, totalOpenDescendants)
func (pm *PageManager) generateBookmarkItems(items []models.Bookmark, parentID int, xrefOffsets map[int]int, pdfBuffer *bytes.Buffer) (int, int, int) {
	if len(items) == 0 {
		return 0, 0, 0
	}

	var itemIDs []int
	var totalCount int
	var b []byte

	// First pass: Allocate IDs for all items at this level
	startID := pm.NextObjectID
	pm.NextObjectID += len(items)
	for i := 0; i < len(items); i++ {
		itemIDs = append(itemIDs, startID+i)
	}

	firstID := itemIDs[0]
	lastID := itemIDs[len(itemIDs)-1]

	// Second pass: Generate each item
	for i, item := range items {
		currentID := itemIDs[i]

		// Recurse for children
		// Pass currentID as parent for children
		childFirst, childLast, childCount := pm.generateBookmarkItems(item.Children, currentID, xrefOffsets, pdfBuffer)

		xrefOffsets[currentID] = pdfBuffer.Len()
		b = b[:0]
		b = strconv.AppendInt(b, int64(currentID), 10)
		b = append(b, " 0 obj\n"...)
		pdfBuffer.Write(b)
		pdfBuffer.WriteString("<< /Title (")
		pdfBuffer.WriteString(escapePDFString(item.Title))
		pdfBuffer.WriteString(")")
		b = b[:0]
		b = append(b, " /Parent "...)
		b = strconv.AppendInt(b, int64(parentID), 10)
		b = append(b, " 0 R"...)
		pdfBuffer.Write(b)

		if i > 0 {
			b = b[:0]
			b = append(b, " /Prev "...)
			b = strconv.AppendInt(b, int64(itemIDs[i-1]), 10)
			b = append(b, " 0 R"...)
			pdfBuffer.Write(b)
		}
		if i < len(items)-1 {
			b = b[:0]
			b = append(b, " /Next "...)
			b = strconv.AppendInt(b, int64(itemIDs[i+1]), 10)
			b = append(b, " 0 R"...)
			pdfBuffer.Write(b)
		}
		if childFirst > 0 {
			b = b[:0]
			b = append(b, " /First "...)
			b = strconv.AppendInt(b, int64(childFirst), 10)
			b = append(b, " 0 R /Last "...)
			b = strconv.AppendInt(b, int64(childLast), 10)
			b = append(b, " 0 R /Count "...)
			b = strconv.AppendInt(b, int64(childCount), 10)
			pdfBuffer.Write(b)
		}

		// Link to page (Dest)
		// Destination array: [PageRef /Fit]
		// Determine page object ID. Page numbers in models are 1-based.
		pageIdx := item.Page - 1
		if pageIdx < 0 {
			pageIdx = 0
		}
		if pageIdx >= len(pm.Pages) {
			pageIdx = len(pm.Pages) - 1
		}

		// We need to resolve the Page Object ID.
		// Since we might not know the exact IDs yet if called early, we rely on the fact that
		// we know the structure of IDs from generator.go, OR we use the logic that Pages are stored in pm.Pages
		// In generator.go, page IDs are already assigned in pm.Pages by the time we generate content?
		// Wait, GenerateBookmarks will be called at the end.

		if pageIdx >= 0 && pageIdx < len(pm.Pages) {
			pageID := pm.Pages[pageIdx]
			b = b[:0]
			b = append(b, " /Dest ["...)
			b = strconv.AppendInt(b, int64(pageID), 10)
			b = append(b, " 0 R /Fit]"...)
			pdfBuffer.Write(b)
		}

		pdfBuffer.WriteString(" >>\nendobj\n")

		// Count: 1 (self) + visible children.
		// Actually, /Outlines Count is "Total number of visible open outline items at all levels".
		// For an item, Count is "number of open descendant items".
		// If Count is positive, children are open. If negative, closed.
		// We'll assume open by default, so we sum up children's counts + number of children.
		// Wait, the spec says: "If the item is open, Count is the sum of the number of its immediate children plus the absolute value of the Count entries of each of those children."
		// So totalCount for this level += 1 (for this item? No, this function returns count of descendants)
		// Correction: The Count in the Dictionary is total open items.

		// Let's assume all are open.
		// Local accumulation for the return value (siblings + descendants)
		totalCount += 1 + childCount
	}

	return firstID, lastID, totalCount
}
