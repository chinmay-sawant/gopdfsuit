package pdf

import (
	"bytes"
	"strings"

	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
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
	// Pre-allocate capacity to prevent mid-flight resizing
	// 64 bytes is plenty for these small PDF lines
	xrefOffsets[outlinesID] = pdfBuffer.Len()
	// PERF-119: single Builder for outlines dictionary
	var sb strings.Builder
	sb.Grow(96)
	var numBuf [20]byte
	sb.Write(strconv.AppendInt(numBuf[:0], int64(outlinesID), 10))
	sb.WriteString(" 0 obj\n<< /Type /Outlines")
	if firstID > 0 {
		sb.WriteString(" /First ")
		sb.Write(strconv.AppendInt(numBuf[:0], int64(firstID), 10))
		sb.WriteString(" 0 R")
	}
	if lastID > 0 {
		sb.WriteString(" /Last ")
		sb.Write(strconv.AppendInt(numBuf[:0], int64(lastID), 10))
		sb.WriteString(" 0 R")
	}
	sb.WriteString(" /Count ")
	sb.Write(strconv.AppendInt(numBuf[:0], int64(count), 10))
	sb.WriteString(" >>\nendobj\n")
	pdfBuffer.WriteString(sb.String())

	return outlinesID
}

// generateBookmarkItems processes a list of bookmarks and returns (firstID, lastID, totalOpenDescendants)
func (pm *PageManager) generateBookmarkItems(items []models.Bookmark, parentID int, xrefOffsets map[int]int, pdfBuffer *bytes.Buffer) (int, int, int) {
	if len(items) == 0 {
		return 0, 0, 0
	}

	var itemIDs []int
	var totalCount int
	// Pre-allocate buffer with capacity for typical bookmark entry
	b := make([]byte, 0, 128)

	// First pass: Allocate IDs for all items at this level
	startID := pm.NextObjectID
	pm.NextObjectID += len(items)
	for i := 0; i < len(items); i++ {
		itemIDs = append(itemIDs, startID+i)
	}

	firstID := itemIDs[0]
	lastID := itemIDs[len(itemIDs)-1]

	// Second pass: Generate each item
	chunk := make([]byte, 1<<20) // PERF-3: single buffer for all items
	pChunk := make([]byte, 256)
	nChunk := make([]byte, 256)
	for i, item := range items {
		currentID := itemIDs[i]

		// Recurse for children
		// Pass currentID as parent for children
		childFirst, childLast, childCount := pm.generateBookmarkItems(item.Children, currentID, xrefOffsets, pdfBuffer)

		xrefOffsets[currentID] = pdfBuffer.Len()

		// Build complete bookmark entry in buffer before writing
		b = b[:0] // Reuse buffer
		escapedTitle := escapePDFString(item.Title)
		b = strconv.AppendInt(b, int64(currentID), 10)
		// Merge static fragments where possible (PERF-119/128)
		var idBuf [20]byte
		parentNum := strconv.AppendInt(idBuf[:0], int64(parentID), 10)
		head := " 0 obj\n<< /Title ("
		mid := ") /Parent "
		need := len(head) + len(escapedTitle) + len(mid) + len(parentNum) + 4
		chunk = chunk[:need]
		o := copy(chunk, head)
		o += copy(chunk[o:], escapedTitle)
		o += copy(chunk[o:], mid)
		o += copy(chunk[o:], parentNum)
		copy(chunk[o:], " 0 R")
		b = append(b, chunk...)

		if i > 0 {
			var pBuf [20]byte
			pNum := strconv.AppendInt(pBuf[:0], int64(itemIDs[i-1]), 10)
			pNeed := len(" /Prev ") + len(pNum) + 4
			pChunk = pChunk[:pNeed]
			o := copy(pChunk, " /Prev ")
			o += copy(pChunk[o:], pNum)
			copy(pChunk[o:], " 0 R")
			b = append(b, pChunk...)
		}

		if i < len(items)-1 {
			var nBuf [20]byte
			nNum := strconv.AppendInt(nBuf[:0], int64(itemIDs[i+1]), 10)
			nNeed := len(" /Next ") + len(nNum) + 4
			nChunk = nChunk[:nNeed]
			o := copy(nChunk, " /Next ")
			o += copy(nChunk[o:], nNum)
			copy(nChunk[o:], " 0 R")
			b = append(b, nChunk...)
		}

		if childFirst > 0 {
			var n1, n2, n3 [20]byte
			a := strconv.AppendInt(n1[:0], int64(childFirst), 10)
			c2 := strconv.AppendInt(n2[:0], int64(childLast), 10)
			c3 := strconv.AppendInt(n3[:0], int64(childCount), 10)
			pref1, mid1, mid2 := " /First ", " 0 R /Last ", " 0 R /Count "
			need := len(pref1) + len(a) + len(mid1) + len(c2) + len(mid2) + len(c3)
			chunk = chunk[:need]
			o := copy(chunk, pref1)
			o += copy(chunk[o:], a)
			o += copy(chunk[o:], mid1)
			o += copy(chunk[o:], c2)
			o += copy(chunk[o:], mid2)
			copy(chunk[o:], c3)
			b = append(b, chunk...)
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
			chunk = chunk[:0]
			var destBuf [20]byte
			destID := strconv.AppendInt(destBuf[:0], int64(pageID), 10)
			pref := " /Dest ["
			suff := " 0 R /Fit]"
			need := len(pref) + len(destID) + len(suff)
			chunk = chunk[:need]
			o := copy(chunk, pref)
			o += copy(chunk[o:], destID)
			copy(chunk[o:], suff)
			b = append(b, chunk...)
		}

		b = append(b, " >>\nendobj\n"...)

		// Single write per bookmark entry
		pdfBuffer.Write(b)

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
