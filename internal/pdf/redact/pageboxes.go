package redact

import (
	"regexp"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

// Viewer-space coordinate handling for redaction.
//
// Clients (the pdf.js-based frontend) address blocks in SENT form: x is the
// left edge in points, and y is the lower-left corner after flipping
// against the DISPLAY height (rotated crop height) -- i.e. the frontend
// already performs `y = displayHeight - (top + h)` with its own pixel
// ratio. The paint operators (`re`) and the text-position extractor instead
// work in MediaBox user space (bottom-left origin, unrotated).
//
// The engine therefore adapts each block with the page's own display
// height: top = displayHeight - y - h recovers the top-left corner, which
// is mapped through crop offset and inverse rotation into MediaBox space.
// For canonical PDFs (no CropBox, no Rotate, zero-based MediaBox) this is
// the identity on values, so previously-working files are unaffected.
//
// Search results travel the exact inverse path, so overlay, re-apply, and
// server/WASM transports all agree. OCR rects are already MediaBox-relative
// (pdftoppm render scaled against MediaBox dims) and pass through untouched.

// displayHeight returns the visible page height in points: the crop height,
// swapped when rotation turns the page sideways.
func displayHeight(boxes pageBoxes) float64 {
	if boxes.rotate == 90 || boxes.rotate == 270 {
		return boxes.crop[2] - boxes.crop[0]
	}
	return boxes.crop[3] - boxes.crop[1]
}

// pageBoxes describes one page in both spaces.
type pageBoxes struct {
	media  [4]float64 // MediaBox [x1 y1 x2 y2], inherited
	crop   [4]float64 // CropBox, inherited, defaults to media
	rotate int        // normalized to 0/90/180/270
}

var (
	boxArrayRe  = regexp.MustCompile(`\[\s*([\d.+-]+)\s+([\d.+-]+)\s+([\d.+-]+)\s+([\d.+-]+)\s*\]`)
	boxRefRe    = regexp.MustCompile(`/(MediaBox|CropBox)\s+(\d+)\s+(\d+)\s+R`)
	rotateRe    = regexp.MustCompile(`/Rotate\s+(-?\d+)`)
	rotateRefRe = regexp.MustCompile(`/Rotate\s+(\d+)\s+(\d+)\s+R`)
	parentRefRe = regexp.MustCompile(`/Parent\s+(\d+)\s+(\d+)\s+R`)
)

func parseBoxArray(body []byte) ([4]float64, bool) {
	match := boxArrayRe.FindSubmatch(body)
	if match == nil {
		return [4]float64{}, false
	}
	var box [4]float64
	for i := 0; i < 4; i++ {
		value, err := strconv.ParseFloat(string(match[i+1]), 64)
		if err != nil {
			return [4]float64{}, false
		}
		box[i] = value
	}
	return box, true
}

// lookupBoxedKey resolves an inheritable page key (/MediaBox, /CropBox,
// /Rotate) by walking the page object up through its /Parent chain, first
// hit wins per the PDF inheritance rules.
func lookupBoxedKey(objMap map[int][]byte, pageObjNum int, key string) []byte {
	current, ok := objMap[pageObjNum]
	if !ok {
		return nil
	}
	for range 16 {
		pattern := regexp.MustCompile(`/` + key + `\s*(\[|[+-]?\d)`)
		if hit := pattern.Find(current); hit != nil {
			return current
		}
		parent := parentRefRe.FindSubmatch(current)
		if parent == nil {
			return current
		}
		parentNum, err := strconv.Atoi(string(parent[1]))
		if err != nil || parentNum <= 0 {
			return current
		}
		next, ok := objMap[parentNum]
		if !ok {
			return current
		}
		current = next
	}
	return current
}

func resolveBox(body []byte, objMap map[int][]byte, key string, fallback [4]float64) [4]float64 {
	if body == nil {
		return fallback
	}
	if ref := boxRefRe.FindSubmatch(body); ref != nil && string(ref[1]) == key {
		refNum, err := strconv.Atoi(string(ref[2]))
		if err == nil {
			if refBody, ok := objMap[refNum]; ok {
				if box, ok := parseBoxArray(refBody); ok {
					return box
				}
			}
		}
		return fallback
	}
	// Direct array: parse starting exactly at this key so a neighboring
	// key's brackets (Annots, the other box) can never match instead.
	keyRe := regexp.MustCompile(`/` + key + `\s*\[`)
	if loc := keyRe.FindIndex(body); loc != nil {
		if box, ok := parseBoxArray(body[loc[0]:]); ok {
			return box
		}
	}
	return fallback
}

func resolveRotate(body []byte, objMap map[int][]byte) int {
	if body == nil {
		return 0
	}
	if ref := rotateRefRe.FindSubmatch(body); ref != nil {
		if refNum, err := strconv.Atoi(string(ref[1])); err == nil {
			if refBody, ok := objMap[refNum]; ok {
				if num := rotateRe.FindSubmatch(refBody); num != nil {
					if value, err := strconv.Atoi(string(num[1])); err == nil {
						return normalizeRotate(value)
					}
				}
			}
		}
		return 0
	}
	if num := rotateRe.FindSubmatch(body); num != nil {
		if value, err := strconv.Atoi(string(num[1])); err == nil {
			return normalizeRotate(value)
		}
	}
	return 0
}

func normalizeRotate(degrees int) int {
	normalized := ((degrees % 360) + 360) % 360
	return ((normalized + 45) / 90 * 90) % 360
}

// resolvePageBoxes returns the MediaBox/CropBox/Rotate triple for a page,
// following inheritance. ok=false means the caller should treat the page
// as canonical (identity transform).
func resolvePageBoxes(objMap map[int][]byte, pdfBytes []byte, pageNum int) (pageBoxes, bool) {
	pageObjNum, err := findPageObject(objMap, pdfBytes, pageNum)
	if err != nil {
		return pageBoxes{}, false
	}
	if _, ok := objMap[pageObjNum]; !ok {
		return pageBoxes{}, false
	}
	mediaBody := lookupBoxedKey(objMap, pageObjNum, "MediaBox")
	media := resolveBox(mediaBody, objMap, "MediaBox", [4]float64{0, 0, 595.28, 841.89})
	cropBody := lookupBoxedKey(objMap, pageObjNum, "CropBox")
	crop := resolveBox(cropBody, objMap, "CropBox", media)
	rotateBody := lookupBoxedKey(objMap, pageObjNum, "Rotate")
	return pageBoxes{media: media, crop: crop, rotate: resolveRotate(rotateBody, objMap)}, true
}

// displayToMediaPoint maps one display-space point (top-left origin, in
// points) into MediaBox user space for the given boxes.
func displayToMediaPoint(boxes pageBoxes, dx, dy float64) (float64, float64) {
	cx1, cy1, cx2, cy2 := boxes.crop[0], boxes.crop[1], boxes.crop[2], boxes.crop[3]
	switch boxes.rotate {
	case 90:
		return cx1 + dy, cy1 + dx
	case 180:
		return cx2 - dx, cy2 - dy
	case 270:
		return cx2 - dy, cy2 - dx
	default:
		return cx1 + dx, cy2 - dy
	}
}

// mediaToDisplayPoint is the exact inverse of displayToMediaPoint.
func mediaToDisplayPoint(boxes pageBoxes, ux, uy float64) (float64, float64) {
	cx1, cy1, cx2, cy2 := boxes.crop[0], boxes.crop[1], boxes.crop[2], boxes.crop[3]
	switch boxes.rotate {
	case 90:
		return uy - cy1, ux - cx1
	case 180:
		return cx2 - ux, cy2 - uy
	case 270:
		return cy2 - uy, cx2 - ux
	default:
		return ux - cx1, cy2 - uy
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

// ensureObjMap returns a usable object map, building one when the redactor
// was constructed without it.
func (r *Redactor) ensureObjMap() map[int][]byte {
	if r.objMap != nil {
		return r.objMap
	}
	objMap, _, err := buildObjectMap(r.pdfBytes)
	if err != nil {
		return nil
	}
	return objMap
}

// mapBlocksToMedia converts client sent-form blocks into MediaBox user
// space, per page. Pages that cannot be resolved pass through unchanged so
// canonical files never regress.
func (r *Redactor) mapBlocksToMedia(blocks []models.RedactionRect) []models.RedactionRect {
	if len(blocks) == 0 {
		return blocks
	}
	objMap := r.ensureObjMap()
	if objMap == nil {
		return blocks
	}
	cache := make(map[int]pageBoxes)
	out := make([]models.RedactionRect, len(blocks))
	for i, rect := range blocks {
		boxes, ok := cache[rect.PageNum]
		if !ok {
			var resolved bool
			boxes, resolved = resolvePageBoxes(objMap, r.pdfBytes, rect.PageNum)
			if !resolved {
				return blocks
			}
			cache[rect.PageNum] = boxes
		}
		// Recover the top-left corner in display points, then map it.
		top := displayHeight(boxes) - rect.Y - rect.Height
		x1, y1 := displayToMediaPoint(boxes, rect.X, top)
		x2, y2 := displayToMediaPoint(boxes, rect.X+rect.Width, top+rect.Height)
		out[i] = models.RedactionRect{
			PageNum: rect.PageNum,
			X:       minFloat(x1, x2),
			Y:       minFloat(y1, y2),
			Width:   absFloat(x2 - x1),
			Height:  absFloat(y2 - y1),
		}
	}
	return out
}

// mapRectsToDisplay converts engine MediaBox-space rects into the sent form
// clients use, per page. Unresolvable pages pass through unchanged.
func (r *Redactor) mapRectsToDisplay(rects []models.RedactionRect) []models.RedactionRect {
	if len(rects) == 0 {
		return rects
	}
	objMap := r.ensureObjMap()
	if objMap == nil {
		return rects
	}
	cache := make(map[int]pageBoxes)
	out := make([]models.RedactionRect, len(rects))
	for i, rect := range rects {
		boxes, ok := cache[rect.PageNum]
		if !ok {
			var resolved bool
			boxes, resolved = resolvePageBoxes(objMap, r.pdfBytes, rect.PageNum)
			if !resolved {
				return rects
			}
			cache[rect.PageNum] = boxes
		}
		dx1, dy1 := mediaToDisplayPoint(boxes, rect.X, rect.Y)
		dx2, dy2 := mediaToDisplayPoint(boxes, rect.X+rect.Width, rect.Y+rect.Height)
		left := minFloat(dx1, dx2)
		top := minFloat(dy1, dy2)
		out[i] = models.RedactionRect{
			PageNum: rect.PageNum,
			X:       left,
			Y:       displayHeight(boxes) - top - absFloat(dy2-dy1),
			Width:   absFloat(dx2 - dx1),
			Height:  absFloat(dy2 - dy1),
		}
	}
	return out
}
