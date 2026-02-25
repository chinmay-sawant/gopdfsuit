// Package svg provides support for converting simple vector graphics (SVG) to PDF commands.
package svg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

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

	fmt.Fprintf(&b, "%.6f 0 0 %.6f 0 1 cm\n", 1.0/width, -1.0/height)

	// State tracking
	inDefs := 0
	definitions := make(map[string]xml.StartElement)

	// Iterate children
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		t, err := decoder.Token()
		if err != nil {
			break
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

				attrs := make(map[string]string)
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
						fmt.Fprintf(&b, "%.3f %.3f %.3f rg\n", r, g, bVal)
					}
				}
				if stroke, ok := attrs["stroke"]; ok {
					r, g, bVal, ok := parseColor(stroke)
					if ok {
						fmt.Fprintf(&b, "%.3f %.3f %.3f RG\n", r, g, bVal)
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
							b.WriteString(fmt.Sprintf("1 0 0 1 %.6f %.6f cm\n", x, -y))
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
				fmt.Fprintf(b, "1 0 0 1 %.2f %.2f cm\n", tx, ty)
			}
		case strings.HasPrefix(parts[i], "scale("):
			args := extractArgs(parts[i:])
			if len(args) >= 2 {
				sx, _ := strconv.ParseFloat(args[0], 64)
				sy, _ := strconv.ParseFloat(args[1], 64)
				fmt.Fprintf(b, "%.4f 0 0 %.4f 0 0 cm\n", sx, sy)
			} else if len(args) == 1 {
				s, _ := strconv.ParseFloat(args[0], 64)
				fmt.Fprintf(b, "%.4f 0 0 %.4f 0 0 cm\n", s, s)
			}
		case strings.HasPrefix(parts[i], "matrix("):
			args := extractArgs(parts[i:])
			if len(args) >= 6 {
				fmt.Fprintf(b, "%s %s %s %s %s %s cm\n", args[0], args[1], args[2], args[3], args[4], args[5])
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
		fmt.Fprintf(b, "%.2f %.2f %.2f RG\n", r, g, blue)
	}

	// SVG default: fill is black if not specified, NOT transparent
	// Treat "none" and "transparent" as no fill
	if fill == "" {
		// Default fill is black per SVG spec
		fill = "black"
		fmt.Fprintf(b, "0.00 0.00 0.00 rg\n") // Black fill
	} else if fill == "none" || fill == "transparent" {
		// Explicit no fill - keep as "none" for drawOp logic
		fill = "none"
	} else if r, g, blue, ok := parseColor(fill); ok {
		fmt.Fprintf(b, "%.2f %.2f %.2f rg\n", r, g, blue)
	} else {
		// Unknown fill value - default to black
		fill = "black"
		fmt.Fprintf(b, "0.00 0.00 0.00 rg\n")
	}

	fmt.Fprintf(b, "%.2f w\n", sw)

	switch name {
	case "rect":
		x := parseDimension(attrs["x"])
		y := parseDimension(attrs["y"])
		w := parseDimension(attrs["width"])
		h := parseDimension(attrs["height"])
		fmt.Fprintf(b, "%.2f %.2f %.2f %.2f re\n", x, y, w, h)
		drawOp(b, fill, stroke)

	case "line":
		x1 := parseDimension(attrs["x1"])
		y1 := parseDimension(attrs["y1"])
		x2 := parseDimension(attrs["x2"])
		y2 := parseDimension(attrs["y2"])
		fmt.Fprintf(b, "%.2f %.2f m %.2f %.2f l\n", x1, y1, x2, y2)
		b.WriteString("S\n")

	case "circle":
		cx := parseDimension(attrs["cx"])
		cy := parseDimension(attrs["cy"])
		r := parseDimension(attrs["r"])
		magic := 0.551784
		d := r * magic
		fmt.Fprintf(b, "%.2f %.2f m\n", cx, cy-r)
		fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+d, cy-r, cx+r, cy-d, cx+r, cy)
		fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+r, cy+d, cx+d, cy+r, cx, cy+r)
		fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-d, cy+r, cx-r, cy+d, cx-r, cy)
		fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-r, cy-d, cx-d, cy-r, cx, cy-r)
		drawOp(b, fill, stroke)

	case "path":
		d := attrs["d"]
		parsePathData(b, d)
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

func parsePathData(b *bytes.Buffer, d string) {
	// Normalize
	d = strings.ReplaceAll(d, ",", " ")
	// Add spaces around commands
	for _, cmd := range []string{"M", "L", "C", "Z", "Q", "H", "V", "m", "l", "c", "z", "q", "h", "v"} {
		d = strings.ReplaceAll(d, cmd, " "+cmd+" ")
	}

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
			fmt.Fprintf(b, "%.2f %.2f m ", x, y)
			cx, cy = x, y
		case "m":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			cy += dy
			fmt.Fprintf(b, "%.2f %.2f m ", cx, cy)

		case "L":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			fmt.Fprintf(b, "%.2f %.2f l ", x, y)
			cx, cy = x, y
		case "l":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			cy += dy
			fmt.Fprintf(b, "%.2f %.2f l ", cx, cy)

		case "H":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx = x
			fmt.Fprintf(b, "%.2f %.2f l ", cx, cy)
		case "h":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			fmt.Fprintf(b, "%.2f %.2f l ", cx, cy) // Treat z inside h case? No, separate case.

		case "V":
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cy = y
			fmt.Fprintf(b, "%.2f %.2f l ", cx, cy)
		case "v":
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cy += dy
			fmt.Fprintf(b, "%.2f %.2f l ", cx, cy)

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
			fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c ", x1, y1, x2, y2, x, y)
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
			fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c ", cx+dx1, cy+dy1, cx+dx2, cy+dy2, cx+dx, cy+dy)
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

			fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c ", cp1x, cp1y, cp2x, cp2y, x, y)
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

			fmt.Fprintf(b, "%.2f %.2f %.2f %.2f %.2f %.2f c ", cp1x, cp1y, cp2x, cp2y, absX, absY)
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
