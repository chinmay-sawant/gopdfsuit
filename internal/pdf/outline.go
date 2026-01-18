package pdf

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// OutlineBuilder builds the PDF outline (bookmarks) tree structure
type OutlineBuilder struct {
	pageManager  *PageManager
	outlineObjID int             // Root outline object ID
	outlineItems []OutlineItem   // Flat list of all outline items with their object IDs
	encryptor    ObjectEncryptor // Encryptor for strings
}

// OutlineItem represents a single outline entry with its object ID
type OutlineItem struct {
	ObjectID         int
	Title            string
	DestKey          string  // Named destination key for PDF/UA-2 compliance
	DestPageID       int     // Page object ID for destination
	DestY            float64 // Y position on destination page
	DestStructElemID int     // PDF/UA-2: Structure element ID for structure destination
	ParentID         int     // Parent outline item object ID
	PrevID           int     // Previous sibling object ID (0 if first)
	NextID           int     // Next sibling object ID (0 if last)
	FirstID          int     // First child object ID (0 if no children)
	LastID           int     // Last child object ID (0 if no children)
	Count            int     // Number of visible descendants (negative if closed)
	Open             bool    // Whether item is open (children visible)
}

// NewOutlineBuilder creates a new outline builder
func NewOutlineBuilder(pm *PageManager, encryptor ObjectEncryptor) *OutlineBuilder {
	return &OutlineBuilder{
		pageManager: pm,
		encryptor:   encryptor,
	}
}

// RegisterNamedDest registers a named destination for internal linking
func (ob *OutlineBuilder) RegisterNamedDest(name string, pageIndex int, y float64) {
	ob.pageManager.NamedDests[name] = NamedDest{
		PageIndex: pageIndex,
		Y:         y,
	}
}

// BuildOutlines creates the outline tree from bookmarks
// Returns the root outline object ID, or 0 if no bookmarks
func (ob *OutlineBuilder) BuildOutlines(bookmarks []models.Bookmark) int {
	if len(bookmarks) == 0 {
		return 0
	}

	// Allocate object IDs for all outline items first
	// We need to know all IDs upfront to set up the tree structure
	ob.allocateOutlineIDs(bookmarks)

	if len(ob.outlineItems) == 0 {
		return 0
	}

	// Create root outline object
	ob.outlineObjID = ob.pageManager.NextObjectID
	ob.pageManager.NextObjectID++

	// Set up parent/sibling relationships
	ob.buildTreeRelationships(bookmarks, ob.outlineObjID, 0)

	// Calculate counts for each item
	ob.calculateCounts()

	// Generate all outline objects
	ob.generateOutlineObjects()

	return ob.outlineObjID
}

// allocateOutlineIDs allocates object IDs for all bookmark items recursively
func (ob *OutlineBuilder) allocateOutlineIDs(bookmarks []models.Bookmark) {
	for _, bm := range bookmarks {
		item := OutlineItem{
			ObjectID: ob.pageManager.NextObjectID,
			Title:    bm.Title,
			Open:     bm.Open,
		}
		ob.pageManager.NextObjectID++

		// Resolve destination
		if bm.Dest != "" {
			// Try to find named destination
			if dest, exists := ob.pageManager.NamedDests[bm.Dest]; exists {
				if dest.PageIndex < len(ob.pageManager.Pages) {
					item.DestPageID = ob.pageManager.Pages[dest.PageIndex]
				} else if len(ob.pageManager.Pages) > 0 {
					item.DestPageID = ob.pageManager.Pages[0]
				}
				item.DestY = dest.Y
			} else {
				// Destination doesn't exist yet, so we treat this bookmark as the DEFINITION
				// of the destination if it has a valid page.
				if bm.Page > 0 {
					pageIndex := bm.Page - 1
					var yPos float64
					if bm.Y > 0 {
						yPos = ob.pageManager.PageDimensions.Height - bm.Y
					} else {
						yPos = ob.pageManager.PageDimensions.Height - margin
					}

					// Register the destination
					ob.RegisterNamedDest(bm.Dest, pageIndex, yPos)

					// Now use it for this bookmark item too
					if pageIndex < len(ob.pageManager.Pages) {
						item.DestPageID = ob.pageManager.Pages[pageIndex]
					} else if len(ob.pageManager.Pages) > 0 {
						item.DestPageID = ob.pageManager.Pages[0]
					}
					item.DestY = yPos
				} else {
					// Fallback if no page specified
					if len(ob.pageManager.Pages) > 0 {
						item.DestPageID = ob.pageManager.Pages[0]
					}
					item.DestY = ob.pageManager.PageDimensions.Height - margin
				}
			}
		} else if bm.Page > 0 {
			// Use explicit page number
			pageIndex := bm.Page - 1 // Convert to 0-based
			if pageIndex >= len(ob.pageManager.Pages) {
				pageIndex = len(ob.pageManager.Pages) - 1
			}
			if pageIndex < 0 {
				pageIndex = 0
			}
			if len(ob.pageManager.Pages) > 0 {
				item.DestPageID = ob.pageManager.Pages[pageIndex]
			}
			if bm.Y > 0 {
				item.DestY = ob.pageManager.PageDimensions.Height - bm.Y
			} else {
				item.DestY = ob.pageManager.PageDimensions.Height - margin
			}
		} else {
			// Default to first page
			if len(ob.pageManager.Pages) > 0 {
				item.DestPageID = ob.pageManager.Pages[0]
			}
			item.DestY = ob.pageManager.PageDimensions.Height - margin
		}

		// PDF/UA-2: Create a Sect (Section) structure element for this bookmark target
		// This enables using /SD (structure destination) in the GoTo action
		sectElem := ob.pageManager.Structure.CreateBookmarkSect(bm.Title)

		// Assign Object ID immediately so we can reference it in the outline dictionary
		sectElem.ObjectID = ob.pageManager.NextObjectID
		ob.pageManager.NextObjectID++

		item.DestStructElemID = sectElem.ObjectID

		// PDF/UA-2: Generate unique destination key and register named destination
		// This allows using /Dest (name) instead of /A << /S /GoTo ... >>
		destKey := fmt.Sprintf("_bm_%d", len(ob.outlineItems))
		item.DestKey = destKey

		// Register the named destination with structure element for PDF/UA-2
		ob.pageManager.NamedDests[destKey] = NamedDest{
			PageIndex:    0, // Will be determined from DestPageID
			Y:            item.DestY,
			StructElemID: item.DestStructElemID,
		}
		// Find the page index from DestPageID
		for pageIdx, pageObjID := range ob.pageManager.Pages {
			if pageObjID == item.DestPageID {
				ob.pageManager.NamedDests[destKey] = NamedDest{
					PageIndex:    pageIdx,
					Y:            item.DestY,
					StructElemID: item.DestStructElemID,
				}
				break
			}
		}

		// PDF/UA-2: If this bookmark defines a user-specified destination (bm.Dest),
		// update that destination with the structure element ID so internal links work
		if bm.Dest != "" {
			if existingDest, exists := ob.pageManager.NamedDests[bm.Dest]; exists {
				existingDest.StructElemID = item.DestStructElemID
				ob.pageManager.NamedDests[bm.Dest] = existingDest
			}
		}

		ob.outlineItems = append(ob.outlineItems, item)

		// Recursively process children
		if len(bm.Children) > 0 {
			ob.allocateOutlineIDs(bm.Children)
		}
	}
}

// buildTreeRelationships sets up parent/sibling/child relationships
func (ob *OutlineBuilder) buildTreeRelationships(bookmarks []models.Bookmark, parentID int, startIdx int) int {
	idx := startIdx
	firstChildIdx := -1
	prevIdx := -1

	for i, bm := range bookmarks {
		if idx >= len(ob.outlineItems) {
			break
		}

		currentIdx := idx
		ob.outlineItems[currentIdx].ParentID = parentID

		// Track first child for parent reference
		if i == 0 {
			firstChildIdx = currentIdx
		}

		// Set up sibling links
		if prevIdx >= 0 {
			ob.outlineItems[prevIdx].NextID = ob.outlineItems[currentIdx].ObjectID
			ob.outlineItems[currentIdx].PrevID = ob.outlineItems[prevIdx].ObjectID
		}

		idx++
		prevIdx = currentIdx

		// Process children
		if len(bm.Children) > 0 {
			childStartIdx := idx
			idx = ob.buildTreeRelationships(bm.Children, ob.outlineItems[currentIdx].ObjectID, idx)

			// Set first/last child pointers
			if childStartIdx < len(ob.outlineItems) {
				ob.outlineItems[currentIdx].FirstID = ob.outlineItems[childStartIdx].ObjectID
				// Find last child
				for j := childStartIdx; j < idx; j++ {
					if ob.outlineItems[j].ParentID == ob.outlineItems[currentIdx].ObjectID && ob.outlineItems[j].NextID == 0 {
						ob.outlineItems[currentIdx].LastID = ob.outlineItems[j].ObjectID
						break
					}
				}
			}
		}
	}

	// Silence unused variable warning - firstChildIdx is used for documentation
	_ = firstChildIdx

	return idx
}

// calculateCounts calculates the Count value for each outline item
func (ob *OutlineBuilder) calculateCounts() {
	for i := range ob.outlineItems {
		count := ob.countDescendants(i)
		if !ob.outlineItems[i].Open && count > 0 {
			count = -count // Negative count means closed
		}
		ob.outlineItems[i].Count = count
	}
}

// countDescendants counts all descendants of an outline item
func (ob *OutlineBuilder) countDescendants(idx int) int {
	count := 0
	parentID := ob.outlineItems[idx].ObjectID

	for i := range ob.outlineItems {
		if ob.outlineItems[i].ParentID == parentID {
			count++ // Count direct child
			// If child is open, add its descendants too
			if ob.outlineItems[i].Open {
				count += ob.countDescendants(i)
			}
		}
	}

	return count
}

// generateOutlineObjects creates all the outline PDF objects
func (ob *OutlineBuilder) generateOutlineObjects() {
	// Find first and last top-level items
	var firstTopLevel, lastTopLevel int
	for i := range ob.outlineItems {
		if ob.outlineItems[i].ParentID == ob.outlineObjID {
			if firstTopLevel == 0 {
				firstTopLevel = ob.outlineItems[i].ObjectID
			}
			lastTopLevel = ob.outlineItems[i].ObjectID
		}
	}

	// Calculate total count for root
	totalCount := 0
	for i := range ob.outlineItems {
		if ob.outlineItems[i].ParentID == ob.outlineObjID {
			totalCount++
			if ob.outlineItems[i].Open {
				totalCount += ob.countDescendants(i)
			}
		}
	}

	// Generate root outline dictionary
	var rootDict strings.Builder
	rootDict.WriteString("<< /Type /Outlines")
	if firstTopLevel > 0 {
		rootDict.WriteString(fmt.Sprintf(" /First %d 0 R", firstTopLevel))
		rootDict.WriteString(fmt.Sprintf(" /Last %d 0 R", lastTopLevel))
	}
	rootDict.WriteString(fmt.Sprintf(" /Count %d", totalCount))
	rootDict.WriteString(" >>")
	ob.pageManager.ExtraObjects[ob.outlineObjID] = rootDict.String()

	// Generate each outline item
	for _, item := range ob.outlineItems {
		var itemDict strings.Builder
		itemDict.WriteString("<<")

		// Handle Title encryption
		if ob.encryptor != nil {
			// Encrypt title (handle UTF-16BE encoding if needed)
			var titleBytes []byte
			hasUnicode := false
			for _, r := range item.Title {
				if r > 127 {
					hasUnicode = true
					break
				}
			}

			if hasUnicode {
				// Create UTF-16BE bytes with BOM
				titleBytes = append(titleBytes, 0xFE, 0xFF)
				for _, r := range item.Title {
					if r <= 0xFFFF {
						titleBytes = append(titleBytes, byte(r>>8), byte(r))
					} else {
						r -= 0x10000
						high := 0xD800 + ((r >> 10) & 0x3FF)
						low := 0xDC00 + (r & 0x3FF)
						titleBytes = append(titleBytes, byte(high>>8), byte(high), byte(low>>8), byte(low))
					}
				}
			} else {
				// ASCII bytes
				titleBytes = []byte(item.Title)
			}

			encrypted := ob.encryptor.EncryptString(titleBytes, item.ObjectID, 0)
			itemDict.WriteString(fmt.Sprintf(" /Title <%s>", hex.EncodeToString(encrypted)))
		} else {
			itemDict.WriteString(fmt.Sprintf(" /Title (%s)", escapeTextUnicode(item.Title)))
		}

		itemDict.WriteString(fmt.Sprintf(" /Parent %d 0 R", item.ParentID))

		// PDF/UA-2 Compliance: Use /Dest (name) instead of /A << /S /GoTo ... >>
		// The named destination contains both /D and /SD entries
		if item.DestKey != "" {
			itemDict.WriteString(fmt.Sprintf(" /Dest (%s)", escapeText(item.DestKey)))
		} else if item.DestPageID > 0 {
			// Fallback for items without a destination key (shouldn't happen normally)
			itemDict.WriteString(fmt.Sprintf(" /Dest [%d 0 R /XYZ null %s null]",
				item.DestPageID, fmtNum(item.DestY)))
		}

		if item.PrevID > 0 {
			itemDict.WriteString(fmt.Sprintf(" /Prev %d 0 R", item.PrevID))
		}
		if item.NextID > 0 {
			itemDict.WriteString(fmt.Sprintf(" /Next %d 0 R", item.NextID))
		}
		if item.FirstID > 0 {
			itemDict.WriteString(fmt.Sprintf(" /First %d 0 R", item.FirstID))
			itemDict.WriteString(fmt.Sprintf(" /Last %d 0 R", item.LastID))
			itemDict.WriteString(fmt.Sprintf(" /Count %d", item.Count))
		}

		itemDict.WriteString(" >>")
		ob.pageManager.ExtraObjects[item.ObjectID] = itemDict.String()
	}
}

// escapeTextUnicode escapes text for PDF string, handling Unicode
func escapeTextUnicode(s string) string {
	// Check if string contains non-ASCII characters
	hasUnicode := false
	for _, r := range s {
		if r > 127 {
			hasUnicode = true
			break
		}
	}

	if hasUnicode {
		// Use UTF-16BE encoding with BOM for Unicode strings
		var result strings.Builder
		result.WriteString("\\xFE\\xFF") // UTF-16BE BOM
		for _, r := range s {
			// Convert to UTF-16BE
			if r <= 0xFFFF {
				result.WriteString(fmt.Sprintf("\\x%02X\\x%02X", (r>>8)&0xFF, r&0xFF))
			} else {
				// Surrogate pair for characters > 0xFFFF
				r -= 0x10000
				high := 0xD800 + ((r >> 10) & 0x3FF)
				low := 0xDC00 + (r & 0x3FF)
				result.WriteString(fmt.Sprintf("\\x%02X\\x%02X\\x%02X\\x%02X",
					(high>>8)&0xFF, high&0xFF, (low>>8)&0xFF, low&0xFF))
			}
		}
		return result.String()
	}

	// ASCII string - use standard escaping
	return escapeText(s)
}

// GetNamedDestinations returns the names dictionary object content for catalog
// This enables internal links to work with named destinations
func (ob *OutlineBuilder) GetNamedDestinations() (int, bool) {
	if len(ob.pageManager.NamedDests) == 0 {
		return 0, false
	}

	// Build Names array for Dests name tree
	var namesArray strings.Builder
	namesArray.WriteString("[")

	// Sort names for binary search tree compliance
	names := make([]string, 0, len(ob.pageManager.NamedDests))
	for name := range ob.pageManager.NamedDests {
		names = append(names, name)
	}
	// Simple sort
	for i := 0; i < len(names)-1; i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}

	// Create Dests name tree object ID upfront for encryption key generation
	destsTreeID := ob.pageManager.NextObjectID
	ob.pageManager.NextObjectID++

	for i, name := range names {
		dest := ob.pageManager.NamedDests[name]
		pageObjID := 0
		if dest.PageIndex < len(ob.pageManager.Pages) {
			pageObjID = ob.pageManager.Pages[dest.PageIndex]
		} else if len(ob.pageManager.Pages) > 0 {
			pageObjID = ob.pageManager.Pages[0]
		}
		if i > 0 {
			namesArray.WriteString(" ")
		}

		// Handle Name encryption
		nameStr := ""
		if ob.encryptor != nil {
			// Names in name tree are strings and must be encrypted
			// Usually names are ASCII, but handle them as bytes
			encrypted := ob.encryptor.EncryptString([]byte(name), destsTreeID, 0)
			nameStr = fmt.Sprintf("<%s>", hex.EncodeToString(encrypted))
		} else {
			nameStr = fmt.Sprintf("(%s)", escapeText(name))
		}

		// PDF/UA-2: Output as dictionary with both /D and /SD keys
		// /D is the page-based destination (for compatibility)
		// /SD is the structure destination (required for PDF/UA-2)
		if dest.StructElemID > 0 {
			namesArray.WriteString(fmt.Sprintf("%s << /D [%d 0 R /XYZ null %s null] /SD [%d 0 R /XYZ null %s null] >>",
				nameStr, pageObjID, fmtNum(dest.Y), dest.StructElemID, fmtNum(dest.Y)))
		} else {
			// Fallback for destinations without structure element (not fully PDF/UA-2 compliant)
			namesArray.WriteString(fmt.Sprintf("%s [%d 0 R /XYZ null %s null]",
				nameStr, pageObjID, fmtNum(dest.Y)))
		}
	}
	namesArray.WriteString("]")

	destsTreeContent := fmt.Sprintf("<< /Names %s >>", namesArray.String())
	ob.pageManager.ExtraObjects[destsTreeID] = destsTreeContent

	// Create Names dictionary object
	namesID := ob.pageManager.NextObjectID
	ob.pageManager.NextObjectID++

	namesContent := fmt.Sprintf("<< /Dests %d 0 R >>", destsTreeID)
	ob.pageManager.ExtraObjects[namesID] = namesContent

	return namesID, true
}
