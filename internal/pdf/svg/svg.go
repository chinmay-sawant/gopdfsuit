// Package svg provides support for converting simple vector graphics (SVG) to PDF commands.
package svg

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

func fmtNum(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func fmtNum6(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

func fmtNum4(f float64) string {
	return strconv.FormatFloat(f, 'f', 4, 64)
}

func fmtNum3(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}

// SVG support for converting simple vector graphics to PDF commands

// SVG represents the root of an SVG document.
type SVG struct {
	XMLName  xml.Name `xml:"svg"`
	Width    string   `xml:"width,attr"`
	Height   string   `xml:"height,attr"`
	ViewBox  string   `xml:"viewBox,attr"`
	Children []Token  `xml:",any"`
}

// Token represents a generic SVG element.
type Token struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
}

func parseDimension(val string) float64 {
	val = strings.TrimSuffix(val, "px")
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

// ConvertSVGToPDFCommands parses SVG data and returns PDF content stream commands
//
//nolint:gocyclo
func ConvertSVGToPDFCommands(data []byte) ([]byte, int, int, error) {
	var svg SVG
	// Handle XML namespace issues by just ignoring them for simple parsing?
	// Go's XML parser is strict. We might need to handle xmlns.
	if err := xml.Unmarshal(data, &svg); err != nil {
		return nil, 0, 0, err
	}

	width := parseDimension(svg.Width)
	height := parseDimension(svg.Height)

	// If width/height missing, try ViewBox
	if width == 0 || height == 0 {
		parts := strings.Fields(strings.ReplaceAll(svg.ViewBox, ",", " "))
		if len(parts) == 4 {
			width, _ = strconv.ParseFloat(parts[2], 64)
			height, _ = strconv.ParseFloat(parts[3], 64)
		}
	}

	// Default fallbacks
	if width == 0 {
		width = 100
	}
	if height == 0 {
		height = 100
	}

	var b bytes.Buffer

	// PDF coordinate system is bottom-up (0,0 at bottom-left).
	// SVG is top-down (0,0 at top-left).
	// We want to map SVG content (0..width, 0..height) into PDF Unit Square (0..1, 0..1).
	// Specifically, SVG (0,0) -> PDF (0,1) [top-left]
	// SVG (width, height) -> PDF (1,0) [bottom-right]
	// Matrix: x' = x/width, y' = 1 - y/height
	// [ 1/w  0   0 ]
	// [  0 -1/h  0 ]
	// [  0   1   1 ]
	// M = [1/w 0 0 -1/h 0 1]

	b.WriteString(fmtNum6(1.0/width))
	b.WriteString(" 0 0 ")
	b.WriteString(fmtNum6(-1.0/height))
	b.WriteString(" 0 1 cm\n")

	// State tracking
	inDefs := 0
	definitions := make(map[string]xml.StartElement)

	// Iterate children
	decoder := xml.NewDecoder(bytes.NewReader(data))
	// Attribute map for group inheritance — moved outside to avoid allocation in loop
	attrs := make(map[string]string)
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, 0, err
		}
		switch se := t.(type) {
		case xml.StartElement:
			// Store elements with IDs for later reference (e.g. via <use>)
			for _, attr := range se.Attr {
				if attr.Name.Local == "id" {
					definitions[attr.Value] = se.Copy()
					break
				}
			}

			if se.Name.Local == "defs" {
				inDefs++
				continue
			}
			if se.Name.Local == "g" {
				b.WriteString("q\n")

				// clear attributes
				for k := range attrs {
					delete(attrs, k)
				}
				for _, a := range se.Attr {
					attrs[a.Name.Local] = a.Value
				}

				// Apply transforms
				if t, ok := attrs["transform"]; ok {
					applyTransform(&b, t)
				}

				// Apply group styles (fill/stroke) which inherit
				if fill, ok := attrs["fill"]; ok {
					r, g, bVal, ok := parseColor(fill)
					if ok {
						b.WriteString(fmtNum3(r))
						b.WriteString(" ")
						b.WriteString(fmtNum3(g))
						b.WriteString(" ")
						b.WriteString(fmtNum3(bVal))
						b.WriteString(" rg\n")
					}
				}
				if stroke, ok := attrs["stroke"]; ok {
					r, g, bVal, ok := parseColor(stroke)
					if ok {
						b.WriteString(fmtNum3(r))
						b.WriteString(" ")
						b.WriteString(fmtNum3(g))
						b.WriteString(" ")
						b.WriteString(fmtNum3(bVal))
						b.WriteString(" RG\n")
					}
				}
			}

			if se.Name.Local == "use" {
				if inDefs > 0 {
					continue
				}
				// Handle <use>
				var href string
				var transform string
				var x, y float64

				for _, attr := range se.Attr {
					if attr.Name.Local == "href" || (attr.Name.Space == "http://www.w3.org/1999/xlink" && attr.Name.Local == "href") {
						href = attr.Value
					}
					if attr.Name.Local == "transform" {
						transform = attr.Value
					}
					if attr.Name.Local == "x" {
						x = parseDimension(attr.Value)
					}
					if attr.Name.Local == "y" {
						y = parseDimension(attr.Value)
					}
				}

				if href != "" {
					id := strings.TrimPrefix(href, "#")
					if refEl, ok := definitions[id]; ok {
						// Save state
						b.WriteString("q\n")

						// Apply use-specific transform/translation
						if x != 0 || y != 0 {
							b.WriteString("1 0 0 1 ")
							b.WriteString(fmtNum6(x))
							b.WriteString(" ")
							b.WriteString(fmtNum6(-y))
							b.WriteString(" cm\n")
						}
						if transform != "" {
							applyTransform(&b, transform) // Note: height might be irrelevant for purely relative transforms but needed for coordinate flip
						}

						// Draw referenced element
						processElement(&b, refEl)

						// Restore state
						b.WriteString("Q\n")
					}
				}
				continue
			}

			// Don't draw regular elements if we are in a defs block
			if inDefs > 0 {
				continue
			}

			processElement(&b, se)

		case xml.EndElement:
			if se.Name.Local == "defs" {
				inDefs--
			}
			// basic group closure support
			if se.Name.Local == "g" {
				b.WriteString("Q\n")
			}
		}
	}

	return b.Bytes(), int(width), int(height), nil
}

func processElement(b *bytes.Buffer, se xml.StartElement) {
	attrs := make(map[string]string)
	for _, a := range se.Attr {
		attrs[a.Name.Local] = a.Value
	}

	// Parse style attribute if present
	if style, ok := attrs["style"]; ok {
		styleParts := strings.Split(style, ";")
		for _, part := range styleParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				k := strings.TrimSpace(kv[0])
				v := strings.TrimSpace(kv[1])
				attrs[k] = v
			}
		}
	}

	// Handle transforms on the element itself
	if t, ok := attrs["transform"]; ok {
		b.WriteString("q\n")
		applyTransform(b, t)
	}

	processVisualElement(b, se.Name.Local, attrs)

	if _, ok := attrs["transform"]; ok {
		b.WriteString("Q\n")
	}
}

func applyTransform(b *bytes.Buffer, t string) {
	// Basic transform parser
	t = strings.ReplaceAll(t, ",", " ")
	parts := strings.Fields(t)

	for i := range parts {
		switch {
		case strings.HasPrefix(parts[i], "translate("):
			args := extractArgs(parts[i:])
			if len(args) >= 2 {
				tx, _ := strconv.ParseFloat(args[0], 64)
				ty, _ := strconv.ParseFloat(args[1], 64)
				b.WriteString("1 0 0 1 ")
				b.WriteString(fmtNum(tx))
				b.WriteString(" ")
				b.WriteString(fmtNum(ty))
				b.WriteString(" cm\n")
			}
		case strings.HasPrefix(parts[i], "scale("):
			args := extractArgs(parts[i:])
			if len(args) >= 2 {
				sx, _ := strconv.ParseFloat(args[0], 64)
				sy, _ := strconv.ParseFloat(args[1], 64)
				b.WriteString(fmtNum4(sx))
				b.WriteString(" 0 0 ")
				b.WriteString(fmtNum4(sy))
				b.WriteString(" 0 0 cm\n")
			} else if len(args) == 1 {
				s, _ := strconv.ParseFloat(args[0], 64)
				b.WriteString(fmtNum4(s))
				b.WriteString(" 0 0 ")
				b.WriteString(fmtNum4(s))
				b.WriteString(" 0 0 cm\n")
			}
		case strings.HasPrefix(parts[i], "matrix("):
			args := extractArgs(parts[i:])
			if len(args) >= 6 {
				b.WriteString(args[0])
				b.WriteString(" ")
				b.WriteString(args[1])
				b.WriteString(" ")
				b.WriteString(args[2])
				b.WriteString(" ")
				b.WriteString(args[3])
				b.WriteString(" ")
				b.WriteString(args[4])
				b.WriteString(" ")
				b.WriteString(args[5])
				b.WriteString(" cm\n")
			}
		}
	}
}

func extractArgs(tokens []string) []string {
	s := strings.Join(tokens, " ")
	start := strings.Index(s, "(")
	end := strings.Index(s, ")")
	if start == -1 || end == -1 {
		return nil
	}
	inner := s[start+1 : end]
	return strings.Fields(strings.ReplaceAll(inner, ",", " "))
}

func parseColor(c string) (float64, float64, float64, bool) {
	c = strings.TrimSpace(c)
	if c == "" || c == "none" || c == "transparent" { //nolint:goconst
		return 0, 0, 0, false
	}
	if after, ok := strings.CutPrefix(c, "#"); ok {
		// Parse hex
		hex := after
		if len(hex) == 3 {
			r, _ := strconv.ParseInt(string(hex[0])+string(hex[0]), 16, 64)
			g, _ := strconv.ParseInt(string(hex[1])+string(hex[1]), 16, 64)
			b, _ := strconv.ParseInt(string(hex[2])+string(hex[2]), 16, 64)
			return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0, true
		} else if len(hex) == 6 {
			r, _ := strconv.ParseInt(hex[0:2], 16, 64)
			g, _ := strconv.ParseInt(hex[2:4], 16, 64)
			b, _ := strconv.ParseInt(hex[4:6], 16, 64)
			return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0, true
		}
	}
	// Handle rgb(r, g, b) and rgba(r, g, b, a) formats
	if strings.HasPrefix(c, "rgb") {
		start := strings.Index(c, "(")
		end := strings.LastIndex(c, ")")
		if start != -1 && end != -1 && end > start {
			inner := c[start+1 : end]
			parts := strings.Split(inner, ",")
			if len(parts) >= 3 {
				r := parseColorComponent(strings.TrimSpace(parts[0]))
				g := parseColorComponent(strings.TrimSpace(parts[1]))
				b := parseColorComponent(strings.TrimSpace(parts[2]))
				return r, g, b, true
			}
		}
	}
	// Basic color names
	switch strings.ToLower(c) {
	case "black": //nolint:goconst
		return 0, 0, 0, true
	case "white":
		return 1, 1, 1, true
	case "red":
		return 1, 0, 0, true
	case "green", "lime":
		return 0, 1, 0, true
	case "blue":
		return 0, 0, 1, true
	case "yellow":
		return 1, 1, 0, true
	case "cyan", "aqua":
		return 0, 1, 1, true
	case "magenta", "fuchsia":
		return 1, 0, 1, true
	case "gray", "grey":
		return 0.5, 0.5, 0.5, true
	case "silver":
		return 0.75, 0.75, 0.75, true
	case "maroon":
		return 0.5, 0, 0, true
	case "olive":
		return 0.5, 0.5, 0, true
	case "navy":
		return 0, 0, 0.5, true
	case "purple":
		return 0.5, 0, 0.5, true
	case "teal":
		return 0, 0.5, 0.5, true
	case "orange":
		return 1, 0.647, 0, true
	}
	return 0, 0, 0, false
}

// parseColorComponent parses a single RGB component (0-255 or percentage)
func parseColorComponent(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		return val / 100.0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val / 255.0
}

func processVisualElement(b *bytes.Buffer, name string, attrs map[string]string) {
	stroke := attrs["stroke"]
	fill := attrs["fill"]

	strokeWidth := attrs["stroke-width"]
	sw := 1.0
	if strokeWidth != "" {
		sw = parseDimension(strokeWidth)
	}

	b.WriteString("q\n") // Save state

	// Apply styles
	if r, g, blue, ok := parseColor(stroke); ok {
		b.WriteString(fmtNum(r)); b.WriteString(" "); b.WriteString(fmtNum(g)); b.WriteString(" "); b.WriteString(fmtNum(blue)); b.WriteString(" RG\n")
	}

	// SVG default: fill is black if not specified, NOT transparent
	// Treat "none" and "transparent" as no fill
	if fill == "" {
		// Default fill is black per SVG spec
		fill = "black"
		b.WriteString("0.00 0.00 0.00 rg\n") // Black fill
	} else if fill == "none" || fill == "transparent" {
		// Explicit no fill - keep as "none" for drawOp logic
		fill = "none"
	} else if r, g, blue, ok := parseColor(fill); ok {
		b.WriteString(fmtNum(r)); b.WriteString(" "); b.WriteString(fmtNum(g)); b.WriteString(" "); b.WriteString(fmtNum(blue)); b.WriteString(" rg\n")
	} else {
		// Unknown fill value - default to black
		fill = "black"
		b.WriteString("0.00 0.00 0.00 rg\n")
	}

	b.WriteString(fmtNum(sw)); b.WriteString(" w\n")

	switch name {
	case "rect":
		x := parseDimension(attrs["x"])
		y := parseDimension(attrs["y"])
		w := parseDimension(attrs["width"])
		h := parseDimension(attrs["height"])
		b.WriteString(fmtNum(x)); b.WriteString(" "); b.WriteString(fmtNum(y)); b.WriteString(" "); b.WriteString(fmtNum(w)); b.WriteString(" "); b.WriteString(fmtNum(h)); b.WriteString(" re\n")
		drawOp(b, fill, stroke)

	case "line":
		x1 := parseDimension(attrs["x1"])
		y1 := parseDimension(attrs["y1"])
		x2 := parseDimension(attrs["x2"])
		y2 := parseDimension(attrs["y2"])
		b.WriteString(fmtNum(x1)); b.WriteString(" "); b.WriteString(fmtNum(y1)); b.WriteString(" m "); b.WriteString(fmtNum(x2)); b.WriteString(" "); b.WriteString(fmtNum(y2)); b.WriteString(" l\n")
		b.WriteString("S\n")

	case "circle":
		cx := parseDimension(attrs["cx"])
		cy := parseDimension(attrs["cy"])
		r := parseDimension(attrs["r"])
		magic := 0.551784
		d := r * magic
		b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy-r)); b.WriteString(" m\n")
		b.WriteString(fmtNum(cx+d)); b.WriteString(" "); b.WriteString(fmtNum(cy-r)); b.WriteString(" "); b.WriteString(fmtNum(cx+r)); b.WriteString(" "); b.WriteString(fmtNum(cy-d)); b.WriteString(" "); b.WriteString(fmtNum(cx+r)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" c\n")
		b.WriteString(fmtNum(cx+r)); b.WriteString(" "); b.WriteString(fmtNum(cy+d)); b.WriteString(" "); b.WriteString(fmtNum(cx+d)); b.WriteString(" "); b.WriteString(fmtNum(cy+r)); b.WriteString(" "); b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy+r)); b.WriteString(" c\n")
		b.WriteString(fmtNum(cx-d)); b.WriteString(" "); b.WriteString(fmtNum(cy+r)); b.WriteString(" "); b.WriteString(fmtNum(cx-r)); b.WriteString(" "); b.WriteString(fmtNum(cy+d)); b.WriteString(" "); b.WriteString(fmtNum(cx-r)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" c\n")
		b.WriteString(fmtNum(cx-r)); b.WriteString(" "); b.WriteString(fmtNum(cy-d)); b.WriteString(" "); b.WriteString(fmtNum(cx-d)); b.WriteString(" "); b.WriteString(fmtNum(cy-r)); b.WriteString(" "); b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy-r)); b.WriteString(" c\n")
		drawOp(b, fill, stroke)

	case "path":
		d := attrs["d"]
		parseSVGPath(b, d)
		drawOp(b, fill, stroke)
	}

	b.WriteString("Q\n") // Restore state
}

func drawOp(b *bytes.Buffer, fill, stroke string) {
	//nolint:gocritic
	if fill != "none" && stroke != "none" && stroke != "" {
		b.WriteString("B\n") // Fill and Stroke
	} else if fill != "none" {
		b.WriteString("f\n") // Fill
	} else if stroke != "none" && stroke != "" {
		b.WriteString("S\n") // Stroke
	}
}

func parseSVGPath(b *bytes.Buffer, d string) {
	// Normalize
	d = strings.ReplaceAll(d, ",", " ")
	replacer := strings.NewReplacer(
		"M", " M ", "L", " L ", "C", " C ", "Z", " Z ", "Q", " Q ", "H", " H ", "V", " V ",
		"m", " m ", "l", " l ", "c", " c ", "z", " z ", "q", " q ", "h", " h ", "v", " v ",
	)
	d = replacer.Replace(d)

	tokens := strings.Fields(d)
	i := 0

	cx, cy := 0.0, 0.0
	// PDF 'm' operator starts a new subpath. It does set the current point.

	for i < len(tokens) {
		cmd := tokens[i]
		i++
		switch cmd {
		case "M":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			b.WriteString(fmtNum(x)); b.WriteString(" "); b.WriteString(fmtNum(y)); b.WriteString(" m ")
			cx, cy = x, y
		case "m":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			cy += dy
			b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" m ")

		case "L":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			b.WriteString(fmtNum(x)); b.WriteString(" "); b.WriteString(fmtNum(y)); b.WriteString(" l ")
			cx, cy = x, y
		case "l":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			cy += dy
			b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" l ")

		case "H":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx = x
			b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" l ")
		case "h":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" l ") // Treat z inside h case? No, separate case.

		case "V":
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cy = y
			b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" l ")
		case "v":
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cy += dy
			b.WriteString(fmtNum(cx)); b.WriteString(" "); b.WriteString(fmtNum(cy)); b.WriteString(" l ")

		case "C":
			x1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			x2, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y2, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			b.WriteString(fmtNum(x1)); b.WriteString(" "); b.WriteString(fmtNum(y1)); b.WriteString(" "); b.WriteString(fmtNum(x2)); b.WriteString(" "); b.WriteString(fmtNum(y2)); b.WriteString(" "); b.WriteString(fmtNum(x)); b.WriteString(" "); b.WriteString(fmtNum(y)); b.WriteString(" c ")
			cx, cy = x, y

		case "c":
			dx1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dx2, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy2, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			b.WriteString(fmtNum(cx+dx1)); b.WriteString(" "); b.WriteString(fmtNum(cy+dy1)); b.WriteString(" "); b.WriteString(fmtNum(cx+dx2)); b.WriteString(" "); b.WriteString(fmtNum(cy+dy2)); b.WriteString(" "); b.WriteString(fmtNum(cx+dx)); b.WriteString(" "); b.WriteString(fmtNum(cy+dy)); b.WriteString(" c ")
			cx += dx
			cy += dy

		case "Q":
			// Quadratic Bezier: x1 y1 x y
			// Convert to Cubic:
			// CP1 = current + 2/3 * (Q1 - current)
			// CP2 = end + 2/3 * (Q1 - end)
			x1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++

			const k = 2.0 / 3.0
			cp1x := cx + k*(x1-cx)
			cp1y := cy + k*(y1-cy)
			cp2x := x + k*(x1-x)
			cp2y := y + k*(y1-y)

			b.WriteString(fmtNum(cp1x)); b.WriteString(" "); b.WriteString(fmtNum(cp1y)); b.WriteString(" "); b.WriteString(fmtNum(cp2x)); b.WriteString(" "); b.WriteString(fmtNum(cp2y)); b.WriteString(" "); b.WriteString(fmtNum(x)); b.WriteString(" "); b.WriteString(fmtNum(y)); b.WriteString(" c ")
			cx, cy = x, y

		case "q":
			dx1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy1, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++

			// Abs coords for calculation
			absX1 := cx + dx1
			absY1 := cy + dy1
			absX := cx + dx
			absY := cy + dy

			const k = 2.0 / 3.0
			cp1x := cx + k*(absX1-cx)
			cp1y := cy + k*(absY1-cy)
			cp2x := absX + k*(absX1-absX)
			cp2y := absY + k*(absY1-absY)

			b.WriteString(fmtNum(cp1x)); b.WriteString(" "); b.WriteString(fmtNum(cp1y)); b.WriteString(" "); b.WriteString(fmtNum(cp2x)); b.WriteString(" "); b.WriteString(fmtNum(cp2y)); b.WriteString(" "); b.WriteString(fmtNum(absX)); b.WriteString(" "); b.WriteString(fmtNum(absY)); b.WriteString(" c ")
			cx, cy = absX, absY

		case "Z", "z":
			b.WriteString("h ")
			// Typically Z closes subpath to start point.
			// We should technically track subpath start, but for most simple shapes, it works.

		default:
			// Handle implicit repetitions? Or skip
		}
	}
	b.WriteString("\n")
}
