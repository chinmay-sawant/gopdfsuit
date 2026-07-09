// Package svg provides support for converting simple vector graphics (SVG) to PDF commands.
package svg

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
	"sync"
)

// SVG support for converting simple vector graphics to PDF commands

// writeFixed writes f formatted with prec digits after the decimal (same as fmt %.Nf).
func writeFixed(b *bytes.Buffer, f float64, prec int) {
	var tmp [32]byte
	b.Write(strconv.AppendFloat(tmp[:0], f, 'f', prec, 64))
}

// writeFixedN writes space-separated floats at the given precision, then suffix.
func writeFixedN(b *bytes.Buffer, prec int, suffix string, vals ...float64) {
	var buf [128]byte
	p := buf[:0]
	for i, v := range vals {
		if i > 0 {
			p = append(p, ' ')
		}
		p = strconv.AppendFloat(p, v, 'f', prec, 64)
	}
	p = append(p, suffix...)
	b.Write(p)
}

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
	b.Grow(2048)

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

	// Matrix: [1/w 0 0 -1/h 0 1]
	writeFixed(&b, 1.0/width, 6)
	b.WriteString(" 0 0 ")
	writeFixed(&b, -1.0/height, 6)
	b.WriteString(" 0 1 cm\n")

	// State tracking
	inDefs := 0
	definitions := make(map[string]xml.StartElement, 8) // PERF-192

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

				attrs := make(map[string]string, len(se.Attr))
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
						writeFixedN(&b, 3, " rg\n", r, g, bVal)
					}
				}
				if stroke, ok := attrs["stroke"]; ok {
					r, g, bVal, ok := parseColor(stroke)
					if ok {
						writeFixedN(&b, 3, " RG\n", r, g, bVal)
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
							writeFixedN(&b, 6, " cm\n", x, -y)
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
	attrs := make(map[string]string, len(se.Attr))
	for _, a := range se.Attr {
		attrs[a.Name.Local] = a.Value
	}

	// Parse style attribute if present (PERF-47: Index/Cut loop, no Split)
	if style, ok := attrs["style"]; ok {
		rest := style
		for rest != "" {
			var part string
			if i := strings.IndexByte(rest, ';'); i >= 0 {
				part = rest[:i]
				rest = rest[i+1:]
			} else {
				part = rest
				rest = ""
			}
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			k, v, found := strings.Cut(part, ":")
			if !found {
				continue
			}
			attrs[strings.TrimSpace(k)] = strings.TrimSpace(v)
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
		p := parts[i]
		// PERF-122: prefer CutPrefix over HasPrefix + manual slice
		switch {
		case hasSVGTransform(p, "translate("):
			args := extractArgs(parts[i:])
			if len(args) >= 2 {
				tx, _ := strconv.ParseFloat(args[0], 64)
				ty, _ := strconv.ParseFloat(args[1], 64)
				b.WriteString("1 0 0 1 ")
				writeFixedN(b, 2, " cm\n", tx, ty)
			}
		case hasSVGTransform(p, "scale("):
			args := extractArgs(parts[i:])
			if len(args) >= 2 {
				sx, _ := strconv.ParseFloat(args[0], 64)
				sy, _ := strconv.ParseFloat(args[1], 64)
				writeFixed(b, sx, 4)
				b.WriteString(" 0 0 ")
				writeFixed(b, sy, 4)
				b.WriteString(" 0 0 cm\n")
			} else if len(args) == 1 {
				s, _ := strconv.ParseFloat(args[0], 64)
				writeFixed(b, s, 4)
				b.WriteString(" 0 0 ")
				writeFixed(b, s, 4)
				b.WriteString(" 0 0 cm\n")
			}
		case hasSVGTransform(p, "matrix("):
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

func hasSVGTransform(s, prefix string) bool {
	_, ok := strings.CutPrefix(s, prefix)
	return ok
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

var namedColors = map[string][3]float64{
	"black":   {0, 0, 0},
	"white":   {1, 1, 1},
	"red":     {1, 0, 0},
	"green":   {0, 1, 0},
	"lime":    {0, 1, 0},
	"blue":    {0, 0, 1},
	"yellow":  {1, 1, 0},
	"cyan":    {0, 1, 1},
	"aqua":    {0, 1, 1},
	"magenta": {1, 0, 1},
	"fuchsia": {1, 0, 1},
	"gray":    {0.5, 0.5, 0.5},
	"grey":    {0.5, 0.5, 0.5},
	"silver":  {0.75, 0.75, 0.75},
	"maroon":  {0.5, 0, 0},
	"olive":   {0.5, 0.5, 0},
	"navy":    {0, 0, 0.5},
	"purple":  {0.5, 0, 0.5},
	"teal":    {0, 0.5, 0.5},
	"orange":  {1, 0.647, 0},
}

// parsedColor caches parseColor results (PERF-230: avoid repeated parseColor in loops).
type parsedColor [4]float64

var colorCache sync.Map

func parseColor(c string) (float64, float64, float64, bool) {
	// PERF-230: cache parsed color results for repeated references
	if v, ok := colorCache.Load(c); ok {
		pc := v.(parsedColor)
		return pc[0], pc[1], pc[2], pc[3] != 0
	}

	r, g, b, ok := parseColorUncached(c)
	var cached parsedColor
	cached[0] = r
	cached[1] = g
	cached[2] = b
	if ok {
		cached[3] = 1
	}
	colorCache.Store(c, cached)
	return r, g, b, ok
}

func parseColorUncached(c string) (float64, float64, float64, bool) {
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
	// Handle rgb(r, g, b) and rgba(r, g, b, a) formats (PERF-122)
	if afterRGB, ok := strings.CutPrefix(c, "rgb"); ok {
		// afterRGB is "a(...)" for rgba or "(...)" for rgb
		start := strings.Index(afterRGB, "(")
		end := strings.LastIndex(afterRGB, ")")
		if start != -1 && end != -1 && end > start {
			inner := afterRGB[start+1 : end]
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
	if rgb, ok := namedColors[strings.ToLower(c)]; ok {
		return rgb[0], rgb[1], rgb[2], true
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
		writeFixedN(b, 2, " RG\n", r, g, blue)
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
		writeFixedN(b, 2, " rg\n", r, g, blue)
	} else {
		// Unknown fill value - default to black
		fill = "black"
		b.WriteString("0.00 0.00 0.00 rg\n")
	}

	writeFixedN(b, 2, " w\n", sw)

	switch name {
	case "rect":
		x := parseDimension(attrs["x"])
		y := parseDimension(attrs["y"])
		w := parseDimension(attrs["width"])
		h := parseDimension(attrs["height"])
		writeFixedN(b, 2, " re\n", x, y, w, h)
		drawOp(b, fill, stroke)

	case "line":
		x1 := parseDimension(attrs["x1"])
		y1 := parseDimension(attrs["y1"])
		x2 := parseDimension(attrs["x2"])
		y2 := parseDimension(attrs["y2"])
		writeFixedN(b, 2, " m ", x1, y1)
		writeFixedN(b, 2, " l\n", x2, y2)
		b.WriteString("S\n")

	case "circle":
		cx := parseDimension(attrs["cx"])
		cy := parseDimension(attrs["cy"])
		r := parseDimension(attrs["r"])
		magic := 0.551784
		d := r * magic
		writeFixedN(b, 2, " m\n", cx, cy-r)
		writeFixedN(b, 2, " c\n", cx+d, cy-r, cx+r, cy-d, cx+r, cy)
		writeFixedN(b, 2, " c\n", cx+r, cy+d, cx+d, cy+r, cx, cy+r)
		writeFixedN(b, 2, " c\n", cx-d, cy+r, cx-r, cy+d, cx-r, cy)
		writeFixedN(b, 2, " c\n", cx-r, cy-d, cx-d, cy-r, cx, cy-r)
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
			writeFixedN(b, 2, " m ", x, y)
			cx, cy = x, y
		case "m":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			cy += dy
			writeFixedN(b, 2, " m ", cx, cy)

		case "L":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			writeFixedN(b, 2, " l ", x, y)
			cx, cy = x, y
		case "l":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			cy += dy
			writeFixedN(b, 2, " l ", cx, cy)

		case "H":
			x, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx = x
			writeFixedN(b, 2, " l ", cx, cy)
		case "h":
			dx, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cx += dx
			writeFixedN(b, 2, " l ", cx, cy) // Treat z inside h case? No, separate case.

		case "V":
			y, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cy = y
			writeFixedN(b, 2, " l ", cx, cy)
		case "v":
			dy, _ := strconv.ParseFloat(tokens[i], 64)
			i++
			cy += dy
			writeFixedN(b, 2, " l ", cx, cy)

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
			writeFixedN(b, 2, " c ", x1, y1, x2, y2, x, y)
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
			writeFixedN(b, 2, " c ", cx+dx1, cy+dy1, cx+dx2, cy+dy2, cx+dx, cy+dy)
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

			writeFixedN(b, 2, " c ", cp1x, cp1y, cp2x, cp2y, x, y)
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

			writeFixedN(b, 2, " c ", cp1x, cp1y, cp2x, cp2y, absX, absY)
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
