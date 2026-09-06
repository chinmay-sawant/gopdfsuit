// Package svg provides support for converting simple vector graphics (SVG) to PDF commands.
package svg

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/vector"
)

// MaxSVGBytes caps accepted SVG input (4 MiB); larger payloads are rejected.
const MaxSVGBytes = 4 << 20

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
	if len(data) > MaxSVGBytes {
		return nil, 0, 0, errors.New("svg input exceeds 4 MiB limit")
	}
	if bytes.Contains(data, []byte("<!ENTITY")) || bytes.Contains(data, []byte("<!DOCTYPE")) {
		return nil, 0, 0, errors.New("svg input with DOCTYPE/ENTITY declarations is rejected")
	}
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

	vector.WriteFloat(&b, 6, 1.0/width)
	b.WriteString(" 0 0 ")
	vector.WriteFloat(&b, 6, -1.0/height)
	b.WriteString(" 0 1 cm\n")

	// State tracking
	inDefs := 0
	definitions := make(map[string]xml.StartElement)

	// Iterate children
	decoder := xml.NewDecoder(bytes.NewReader(data))

	attrs := make(map[string]string, 4)
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

				clear(attrs)
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
						vector.SetFill(&b, 3, r, g, bVal)
					}
				}
				if stroke, ok := attrs["stroke"]; ok {
					r, g, bVal, ok := parseColor(stroke)
					if ok {
						vector.SetStroke(&b, 3, r, g, bVal)
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
							vector.WriteFloats(&b, 6, x, -y)
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
	attrs := make(map[string]string, 4)
	for _, a := range se.Attr {
		attrs[a.Name.Local] = a.Value
	}

	// Parse style attribute if present
	if style, ok := attrs["style"]; ok {
		styleParts := strings.SplitSeq(style, ";")
		for part := range styleParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			k, v, ok := strings.Cut(part, ":")
			if ok {
				k = strings.TrimSpace(k)
				v = strings.TrimSpace(v)
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
				vector.WriteFloats(b, 2, tx, ty)
				b.WriteString(" cm\n")
			}
		case strings.HasPrefix(parts[i], "scale("):
			args := extractArgs(parts[i:])
			if len(args) >= 2 {
				sx, _ := strconv.ParseFloat(args[0], 64)
				sy, _ := strconv.ParseFloat(args[1], 64)
				vector.WriteFloats(b, 4, sx)
				b.WriteString(" 0 0 ")
				vector.WriteFloats(b, 4, sy)
				b.WriteString(" 0 0 cm\n")
			} else if len(args) == 1 {
				s, _ := strconv.ParseFloat(args[0], 64)
				vector.WriteFloats(b, 4, s)
				b.WriteString(" 0 0 ")
				vector.WriteFloats(b, 4, s)
				b.WriteString(" 0 0 cm\n")
			}
		case strings.HasPrefix(parts[i], "matrix("):
			args := extractArgs(parts[i:])
			if len(args) >= 6 {
				b.WriteString(args[0])
				b.WriteByte(' ')
				b.WriteString(args[1])
				b.WriteByte(' ')
				b.WriteString(args[2])
				b.WriteByte(' ')
				b.WriteString(args[3])
				b.WriteByte(' ')
				b.WriteString(args[4])
				b.WriteByte(' ')
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
	if before, ok := strings.CutSuffix(s, "%"); ok {
		val, _ := strconv.ParseFloat(before, 64)
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
		vector.SetStroke(b, 2, r, g, blue)
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
		vector.SetFill(b, 2, r, g, blue)
	} else {
		// Unknown fill value - default to black
		fill = "black"
		b.WriteString("0.00 0.00 0.00 rg\n")
	}

	vector.LineWidth(b, 2, sw)

	switch name {
	case "rect":
		x := parseDimension(attrs["x"])
		y := parseDimension(attrs["y"])
		w := parseDimension(attrs["width"])
		h := parseDimension(attrs["height"])
		vector.WriteFloats(b, 2, x, y, w, h)
		b.WriteString(" re\n")
		drawOp(b, fill, stroke)

	case "line":
		x1 := parseDimension(attrs["x1"])
		y1 := parseDimension(attrs["y1"])
		x2 := parseDimension(attrs["x2"])
		y2 := parseDimension(attrs["y2"])
		vector.StrokeLine(b, 2, x1, y1, x2, y2)

	case "circle":
		cx := parseDimension(attrs["cx"])
		cy := parseDimension(attrs["cy"])
		r := parseDimension(attrs["r"])
		magic := 0.551784
		d := r * magic
		vector.WriteFloats(b, 2, cx, cy-r)
		b.WriteString(" m\n")
		vector.WriteFloats(b, 2, cx+d, cy-r, cx+r, cy-d, cx+r, cy)
		b.WriteString(" c\n")
		vector.WriteFloats(b, 2, cx+r, cy+d, cx+d, cy+r, cx, cy+r)
		b.WriteString(" c\n")
		vector.WriteFloats(b, 2, cx-d, cy+r, cx-r, cy+d, cx-r, cy)
		b.WriteString(" c\n")
		vector.WriteFloats(b, 2, cx-r, cy-d, cx-d, cy-r, cx, cy-r)
		b.WriteString(" c\n")
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

	cur := &pathCursor{b: b, tokens: strings.Fields(d)}
	// PDF 'm' operator starts a new subpath. It does set the current point.
	for cur.i < len(cur.tokens) {
		cmd := cur.tokens[cur.i]
		cur.i++
		if !cur.command(cmd) {
			return
		}
	}
	b.WriteString("\n")
}

// pathCursor tracks SVG path tokenizing state across commands.
type pathCursor struct {
	b      *bytes.Buffer
	tokens []string
	i      int
	cx, cy float64
}

// take reads n floats with bounds and syntax checks. Malformed path
// data (e.g. d="M 10") aborts the path instead of panicking.
func (c *pathCursor) take(n int) ([]float64, bool) {
	if c.i+n > len(c.tokens) {
		return nil, false
	}
	vals := make([]float64, n)
	for k := range vals {
		v, err := strconv.ParseFloat(c.tokens[c.i+k], 64)
		if err != nil {
			return nil, false
		}
		vals[k] = v
	}
	c.i += n
	return vals, true
}

func (c *pathCursor) lineto(x, y float64) {
	vector.WriteFloats(c.b, 2, x, y)
	c.b.WriteString(" l ")
	c.cx, c.cy = x, y
}

func (c *pathCursor) moveto(x, y float64) {
	vector.WriteFloats(c.b, 2, x, y)
	c.b.WriteString(" m ")
	c.cx, c.cy = x, y
}

func (c *pathCursor) curveto(x1, y1, x2, y2, x, y float64) {
	vector.WriteFloats(c.b, 2, x1, y1, x2, y2, x, y)
	c.b.WriteString(" c ")
	c.cx, c.cy = x, y
}

// quadTo converts a quadratic Bezier to cubic and emits it.
func (c *pathCursor) quadTo(x1, y1, x, y float64) {
	const k = 2.0 / 3.0
	c.curveto(
		c.cx+k*(x1-c.cx), c.cy+k*(y1-c.cy),
		x+k*(x1-x), y+k*(y1-y),
		x, y,
	)
}

// command executes one path command; false aborts the path.
func (c *pathCursor) command(cmd string) bool {
	switch cmd {
	case "M":
		v, ok := c.take(2)
		if !ok {
			return false
		}
		c.moveto(v[0], v[1])
	case "m":
		v, ok := c.take(2)
		if !ok {
			return false
		}
		c.moveto(c.cx+v[0], c.cy+v[1])

	case "L":
		v, ok := c.take(2)
		if !ok {
			return false
		}
		c.lineto(v[0], v[1])
	case "l":
		v, ok := c.take(2)
		if !ok {
			return false
		}
		c.lineto(c.cx+v[0], c.cy+v[1])

	case "H":
		v, ok := c.take(1)
		if !ok {
			return false
		}
		c.lineto(v[0], c.cy)
	case "h":
		v, ok := c.take(1)
		if !ok {
			return false
		}
		c.lineto(c.cx+v[0], c.cy)

	case "V":
		v, ok := c.take(1)
		if !ok {
			return false
		}
		c.lineto(c.cx, v[0])
	case "v":
		v, ok := c.take(1)
		if !ok {
			return false
		}
		c.lineto(c.cx, c.cy+v[0])

	case "C":
		v, ok := c.take(6)
		if !ok {
			return false
		}
		c.curveto(v[0], v[1], v[2], v[3], v[4], v[5])

	case "c":
		v, ok := c.take(6)
		if !ok {
			return false
		}
		c.curveto(c.cx+v[0], c.cy+v[1], c.cx+v[2], c.cy+v[3], c.cx+v[4], c.cy+v[5])

	case "Q":
		// Quadratic Bezier: x1 y1 x y
		v, ok := c.take(4)
		if !ok {
			return false
		}
		c.quadTo(v[0], v[1], v[2], v[3])

	case "q":
		v, ok := c.take(4)
		if !ok {
			return false
		}
		c.quadTo(c.cx+v[0], c.cy+v[1], c.cx+v[2], c.cy+v[3])

	case "Z", "z":
		c.b.WriteString("h ")
		// Typically Z closes subpath to start point.
		// We should technically track subpath start, but for most simple shapes, it works.

	default:
		// Handle implicit repetitions? Or skip
	}
	return true
}
