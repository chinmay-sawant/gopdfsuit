package typstsyntax

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

// RenderContext holds state needed during PDF rendering of math expressions.
type RenderContext struct {
	X          float64 // Current X position
	Y          float64 // Current Y position (baseline)
	FontSize   float64 // Current font size in points
	FontRef    string  // PDF font reference (e.g., "/F1")
	CellWidth  float64 // Available cell width
	CellHeight float64 // Available cell height
	TextColor  string  // RGB color string (e.g., "0 0 0")

	// Callbacks for font operations (provided by the PDF generator)
	EstimateWidth func(text string, fontSize float64) float64
	FormatText    func(text string) string // Formats text for PDF stream (escaped or hex-encoded)
}

// MathLayout represents the calculated layout of a math expression.
type MathLayout struct {
	Width    float64       // Total width of the expression
	Height   float64       // Total height above baseline
	Depth    float64       // Total depth below baseline
	Elements []MathElement // Ordered elements to render
}

// MathElement represents a single renderable element in a math layout.
type MathElement struct {
	Type     ElementType
	X        float64       // Relative X from start
	Y        float64       // Relative Y from baseline
	Text     string        // Text content (for glyph elements)
	FontSize float64       // Font size for this element
	FontRef  string        // Font reference override (optional)
	Width    float64       // Width of element
	Height   float64       // Height of element
	Children []MathElement // Sub-elements (for groups, fractions, etc.)
	// Line drawing (for fraction bars, roots, etc.)
	LineX1    float64
	LineY1    float64
	LineX2    float64
	LineY2    float64
	LineWidth float64
}

// ElementType classifies renderable elements.
type ElementType int

// ElementType constants.
const (
	ElemGlyph  ElementType = iota // Text/glyph rendering
	ElemLine                      // Line drawing (fraction bar, root line, etc.)
	ElemGroup                     // Group of sub-elements
	ElemOffset                    // Position offset
)

// LayoutEngine calculates positions and sizes for math AST nodes.
type LayoutEngine struct {
	ctx *RenderContext
}

// NewLayoutEngine creates a layout engine with the given context.
func NewLayoutEngine(ctx *RenderContext) *LayoutEngine {
	return &LayoutEngine{ctx: ctx}
}

// Layout calculates the layout of an AST node.
func (le *LayoutEngine) Layout(node *Node) *MathLayout {
	if node == nil {
		return &MathLayout{}
	}
	return le.layoutNode(node, le.ctx.FontSize)
}

func (le *LayoutEngine) layoutNode(node *Node, fontSize float64) *MathLayout {
	switch node.Type {
	case NodeLiteral, NodeSymbol, NodeQuotedText:
		return le.layoutText(node.Value, fontSize)
	case NodeOperator:
		return le.layoutOperator(node.Value, fontSize)
	case NodeSuperscript:
		return le.layoutSuperscript(node, fontSize)
	case NodeSubscript:
		return le.layoutSubscript(node, fontSize)
	case NodeFraction:
		return le.layoutFraction(node, fontSize)
	case NodeSqrt:
		return le.layoutSqrt(node, fontSize)
	case NodeRoot:
		return le.layoutRoot(node, fontSize)
	case NodeGroup:
		return le.layoutGroup(node, fontSize)
	case NodeAccent:
		return le.layoutAccent(node, fontSize)
	case NodeMatrix:
		return le.layoutMatrix(node, fontSize)
	case NodeVector:
		return le.layoutVector(node, fontSize)
	case NodeBinom:
		return le.layoutBinom(node, fontSize)
	case NodeCases:
		return le.layoutCases(node, fontSize)
	case NodeCancel:
		return le.layoutCancel(node, fontSize)
	case NodeLR:
		return le.layoutLR(node, fontSize)
	case NodePrime:
		return le.layoutPrime(node, fontSize)
	case NodeUnderOver:
		return le.layoutUnderOver(node, fontSize)
	case NodeStyle, NodeVariant, NodeSize, NodeClass:
		return le.layoutPassthrough(node, fontSize)
	case NodeOp:
		return le.layoutOp(node, fontSize)
	case NodeStretch:
		return le.layoutStretch(node, fontSize)
	case NodeSequence:
		return le.layoutSequence(node.Children, fontSize)
	case NodeFunc:
		return le.layoutGenericFunc(node, fontSize)
	case NodeLineBreak:
		return &MathLayout{Width: 0, Height: fontSize}
	case NodeAlign:
		return &MathLayout{Width: 0, Height: 0}
	default:
		return &MathLayout{}
	}
}

func (le *LayoutEngine) estimateWidth(text string, fontSize float64) float64 {
	if le.ctx.EstimateWidth != nil {
		return le.ctx.EstimateWidth(text, fontSize)
	}
	// Fallback approximation
	return float64(len([]rune(text))) * fontSize * 0.5
}

func (le *LayoutEngine) layoutText(text string, fontSize float64) *MathLayout {
	if text == " " {
		return &MathLayout{
			Width:  fontSize * 0.25,
			Height: fontSize,
			Elements: []MathElement{{
				Type: ElemGlyph, Text: " ", FontSize: fontSize, Width: fontSize * 0.25,
			}},
		}
	}
	w := le.estimateWidth(text, fontSize)
	return &MathLayout{
		Width:  w,
		Height: fontSize,
		Elements: []MathElement{{
			Type: ElemGlyph, Text: text, FontSize: fontSize, Width: w,
		}},
	}
}

func (le *LayoutEngine) layoutOperator(op string, fontSize float64) *MathLayout {
	// Operators have extra spacing
	spacing := fontSize * 0.2
	w := le.estimateWidth(op, fontSize)
	totalW := w + 2*spacing
	return &MathLayout{
		Width:  totalW,
		Height: fontSize,
		Elements: []MathElement{{
			Type: ElemGlyph, Text: op, FontSize: fontSize,
			X: spacing, Width: w,
		}},
	}
}

func (le *LayoutEngine) layoutSuperscript(node *Node, fontSize float64) *MathLayout {
	baseLay := le.layoutNode(node.Children[0], fontSize)
	supFontSize := fontSize * 0.65
	var supLay *MathLayout
	if len(node.Children) >= 2 {
		supLay = le.layoutNode(node.Children[1], supFontSize)
	} else {
		supLay = &MathLayout{}
	}

	// Superscript is raised by ~40% of the base font size
	supRise := fontSize * 0.4

	totalW := baseLay.Width + supLay.Width
	totalH := math.Max(baseLay.Height, supRise+supLay.Height)

	elements := make([]MathElement, 0, len(baseLay.Elements)+len(supLay.Elements))
	elements = append(elements, baseLay.Elements...)

	// Offset superscript elements
	for _, el := range supLay.Elements {
		el.X += baseLay.Width
		el.Y += supRise
		elements = append(elements, el)
	}

	return &MathLayout{Width: totalW, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutSubscript(node *Node, fontSize float64) *MathLayout {
	baseLay := le.layoutNode(node.Children[0], fontSize)
	subFontSize := fontSize * 0.65
	var subLay *MathLayout
	if len(node.Children) >= 2 {
		subLay = le.layoutNode(node.Children[1], subFontSize)
	} else {
		subLay = &MathLayout{}
	}

	// Subscript is lowered by ~25% of the base font size
	subDrop := fontSize * 0.25

	totalW := baseLay.Width + subLay.Width
	totalH := baseLay.Height

	elements := make([]MathElement, 0, len(baseLay.Elements)+len(subLay.Elements))
	elements = append(elements, baseLay.Elements...)

	for _, el := range subLay.Elements {
		el.X += baseLay.Width
		el.Y -= subDrop
		elements = append(elements, el)
	}

	return &MathLayout{Width: totalW, Height: totalH, Depth: subDrop + subLay.Height, Elements: elements}
}

func (le *LayoutEngine) layoutFraction(node *Node, fontSize float64) *MathLayout {
	numLay := le.layoutNode(node.Children[0], fontSize*0.85)
	denLay := le.layoutNode(node.Children[1], fontSize*0.85)

	fracWidth := math.Max(numLay.Width, denLay.Width) + fontSize*0.4
	barY := fontSize * 0.35
	barPadding := fontSize * 0.15

	// Center numerator above bar
	numX := (fracWidth - numLay.Width) / 2
	numY := barY + barPadding

	// Center denominator below bar
	denX := (fracWidth - denLay.Width) / 2
	denY := barY - barPadding - denLay.Height

	elements := make([]MathElement, 0)

	// Numerator
	for _, el := range numLay.Elements {
		el.X += numX
		el.Y += numY
		elements = append(elements, el)
	}

	// Fraction bar
	elements = append(elements, MathElement{
		Type:      ElemLine,
		LineX1:    0,
		LineY1:    barY,
		LineX2:    fracWidth,
		LineY2:    barY,
		LineWidth: 0.5,
	})

	// Denominator
	for _, el := range denLay.Elements {
		el.X += denX
		el.Y += denY
		elements = append(elements, el)
	}

	totalH := numY + numLay.Height
	depth := math.Abs(denY)

	return &MathLayout{Width: fracWidth, Height: totalH, Depth: depth, Elements: elements}
}

func (le *LayoutEngine) layoutSqrt(node *Node, fontSize float64) *MathLayout {
	inner := le.layoutNode(node.Children[0], fontSize)

	// Radical sign dimensions
	radWidth := fontSize * 0.6
	overlineGap := fontSize * 0.1
	totalH := inner.Height + overlineGap + fontSize*0.15
	totalW := radWidth + inner.Width + fontSize*0.1

	elements := make([]MathElement, 0)

	// Radical sign (√) glyph
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "√", FontSize: fontSize * 1.1,
		X: 0, Y: 0, Width: radWidth,
	})

	// Overline (horizontal bar over the radicand)
	elements = append(elements, MathElement{
		Type:      ElemLine,
		LineX1:    radWidth - 1,
		LineY1:    totalH,
		LineX2:    totalW,
		LineY2:    totalH,
		LineWidth: 0.5,
	})

	// Inner content
	for _, el := range inner.Elements {
		el.X += radWidth
		elements = append(elements, el)
	}

	return &MathLayout{Width: totalW, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutRoot(node *Node, fontSize float64) *MathLayout {
	indexLay := le.layoutNode(node.Children[0], fontSize*0.55)
	innerLay := le.layoutNode(node.Children[1], fontSize)

	radWidth := fontSize * 0.6
	totalW := radWidth + innerLay.Width + fontSize*0.1
	totalH := innerLay.Height + fontSize*0.25

	elements := make([]MathElement, 0)

	// Index (small, top-left of radical)
	for _, el := range indexLay.Elements {
		el.X += fontSize * 0.05
		el.Y += totalH - indexLay.Height*0.5
		elements = append(elements, el)
	}

	// Radical sign
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "√", FontSize: fontSize * 1.1,
		X: indexLay.Width * 0.6, Y: 0, Width: radWidth,
	})

	// Overline
	elements = append(elements, MathElement{
		Type:      ElemLine,
		LineX1:    radWidth - 1,
		LineY1:    totalH,
		LineX2:    totalW,
		LineY2:    totalH,
		LineWidth: 0.5,
	})

	// Inner content
	for _, el := range innerLay.Elements {
		el.X += radWidth
		elements = append(elements, el)
	}

	return &MathLayout{Width: totalW, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutGroup(node *Node, fontSize float64) *MathLayout {
	innerLay := le.layoutSequence(node.Children, fontSize)

	delims := getGroupDelimiters(node.Value)
	delimWidth := fontSize * 0.3

	totalW := delimWidth*2 + innerLay.Width
	elements := make([]MathElement, 0)

	// Left delimiter
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[0], FontSize: fontSize,
		X: 0, Width: delimWidth,
	})

	// Inner content, shifted right
	for _, el := range innerLay.Elements {
		el.X += delimWidth
		elements = append(elements, el)
	}

	// Right delimiter
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[1], FontSize: fontSize,
		X: delimWidth + innerLay.Width, Width: delimWidth,
	})

	return &MathLayout{Width: totalW, Height: innerLay.Height, Depth: innerLay.Depth, Elements: elements}
}

func (le *LayoutEngine) layoutAccent(node *Node, fontSize float64) *MathLayout {
	baseLay := le.layoutNode(node.Children[0], fontSize)

	accentChar := ""
	if a, ok := AccentMap[node.FuncName]; ok {
		accentChar = a
	}

	elements := make([]MathElement, 0, len(baseLay.Elements)+1)
	elements = append(elements, baseLay.Elements...)

	// Place accent above the base content
	if accentChar != "" {
		elements = append(elements, MathElement{
			Type: ElemGlyph, Text: accentChar, FontSize: fontSize,
			X: baseLay.Width / 4, Y: fontSize * 0.8, Width: baseLay.Width / 2,
		})
	}

	return &MathLayout{Width: baseLay.Width, Height: baseLay.Height + fontSize*0.3, Elements: elements}
}

func (le *LayoutEngine) layoutMatrix(node *Node, fontSize float64) *MathLayout {
	// For simplicity, layout as a sequence of comma-separated elements
	elemFontSize := fontSize * 0.85
	var elements []MathElement
	totalW := fontSize * 0.3 // left bracket
	maxH := fontSize

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "[", FontSize: fontSize,
		X: 0, Width: fontSize * 0.3,
	})

	for i, arg := range node.Args {
		if i > 0 {
			elements = append(elements, MathElement{
				Type: ElemGlyph, Text: " ", FontSize: elemFontSize,
				X: totalW, Width: elemFontSize * 0.15,
			})
			totalW += elemFontSize * 0.15
		}
		argLay := le.layoutNode(arg, elemFontSize)
		for _, el := range argLay.Elements {
			el.X += totalW
			elements = append(elements, el)
		}
		totalW += argLay.Width
		if argLay.Height > maxH {
			maxH = argLay.Height
		}
	}

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "]", FontSize: fontSize,
		X: totalW, Width: fontSize * 0.3,
	})
	totalW += fontSize * 0.3

	return &MathLayout{Width: totalW, Height: maxH, Elements: elements}
}

func (le *LayoutEngine) layoutVector(node *Node, fontSize float64) *MathLayout {
	return le.layoutMatrix(&Node{Type: NodeMatrix, Args: node.Args}, fontSize)
}

func (le *LayoutEngine) layoutBinom(node *Node, fontSize float64) *MathLayout {
	topLay := le.layoutNode(node.Children[0], fontSize*0.85)
	botLay := le.layoutNode(node.Children[1], fontSize*0.85)

	innerW := math.Max(topLay.Width, botLay.Width) + fontSize*0.3
	delimW := fontSize * 0.3
	totalW := delimW*2 + innerW
	gap := fontSize * 0.15

	topY := fontSize*0.4 + gap
	botY := fontSize*0.4 - gap - botLay.Height

	elements := make([]MathElement, 0)

	// Left paren
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "(", FontSize: fontSize * 1.3,
		X: 0, Width: delimW,
	})

	// Top
	topX := delimW + (innerW-topLay.Width)/2
	for _, el := range topLay.Elements {
		el.X += topX
		el.Y += topY
		elements = append(elements, el)
	}

	// Bottom
	botX := delimW + (innerW-botLay.Width)/2
	for _, el := range botLay.Elements {
		el.X += botX
		el.Y += botY
		elements = append(elements, el)
	}

	// Right paren
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: ")", FontSize: fontSize * 1.3,
		X: delimW + innerW, Width: delimW,
	})

	totalH := topY + topLay.Height
	depth := math.Abs(botY)

	return &MathLayout{Width: totalW, Height: totalH, Depth: depth, Elements: elements}
}

func (le *LayoutEngine) layoutCases(node *Node, fontSize float64) *MathLayout {
	caseFontSize := fontSize * 0.85
	delimW := fontSize * 0.4
	lineGap := fontSize * 0.3
	var elements []MathElement
	var maxW float64
	totalH := float64(0)

	// Left brace
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "{", FontSize: fontSize * 1.5,
		X: 0, Width: delimW,
	})

	// Layout each case
	yOffset := fontSize * 0.5
	for i, arg := range node.Args {
		if i > 0 {
			yOffset -= lineGap
		}
		argLay := le.layoutNode(arg, caseFontSize)
		for _, el := range argLay.Elements {
			el.X += delimW + fontSize*0.2
			el.Y += yOffset
			elements = append(elements, el)
		}
		if argLay.Width > maxW {
			maxW = argLay.Width
		}
		yOffset -= argLay.Height
	}

	totalW := delimW + fontSize*0.2 + maxW
	totalH = fontSize*0.5 - yOffset

	return &MathLayout{Width: totalW, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutCancel(node *Node, fontSize float64) *MathLayout {
	innerLay := le.layoutNode(node.Children[0], fontSize)

	elements := make([]MathElement, 0, len(innerLay.Elements)+1)
	elements = append(elements, innerLay.Elements...)

	// Add diagonal cancel line
	elements = append(elements, MathElement{
		Type:      ElemLine,
		LineX1:    0,
		LineY1:    -innerLay.Depth,
		LineX2:    innerLay.Width,
		LineY2:    innerLay.Height,
		LineWidth: 0.5,
	})

	return &MathLayout{Width: innerLay.Width, Height: innerLay.Height, Depth: innerLay.Depth, Elements: elements}
}

func (le *LayoutEngine) layoutLR(node *Node, fontSize float64) *MathLayout {
	delims := getLRDelimiters(node.FuncName)
	innerLay := le.layoutSequence(node.Args, fontSize)

	delimW := fontSize * 0.3
	totalW := delimW*2 + innerLay.Width

	elements := make([]MathElement, 0)

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[0], FontSize: fontSize,
		X: 0, Width: delimW,
	})

	for _, el := range innerLay.Elements {
		el.X += delimW
		elements = append(elements, el)
	}

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[1], FontSize: fontSize,
		X: delimW + innerLay.Width, Width: delimW,
	})

	return &MathLayout{Width: totalW, Height: innerLay.Height, Depth: innerLay.Depth, Elements: elements}
}

func (le *LayoutEngine) layoutPrime(node *Node, fontSize float64) *MathLayout {
	baseLay := le.layoutNode(node.Children[0], fontSize)
	primeW := le.estimateWidth(node.Value, fontSize*0.6)

	elements := make([]MathElement, 0, len(baseLay.Elements)+1)
	elements = append(elements, baseLay.Elements...)

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: node.Value, FontSize: fontSize * 0.6,
		X: baseLay.Width, Y: fontSize * 0.4, Width: primeW,
	})

	return &MathLayout{Width: baseLay.Width + primeW, Height: baseLay.Height, Elements: elements}
}

func (le *LayoutEngine) layoutUnderOver(node *Node, fontSize float64) *MathLayout {
	if len(node.Args) == 0 {
		return &MathLayout{}
	}
	baseLay := le.layoutNode(node.Args[0], fontSize)

	elements := make([]MathElement, 0, len(baseLay.Elements)+2)
	elements = append(elements, baseLay.Elements...)

	fnName := node.FuncName
	totalH := baseLay.Height

	if strings.HasPrefix(fnName, "over") {
		// Add line/brace above
		elements = append(elements, MathElement{
			Type:      ElemLine,
			LineX1:    0,
			LineY1:    baseLay.Height + fontSize*0.1,
			LineX2:    baseLay.Width,
			LineY2:    baseLay.Height + fontSize*0.1,
			LineWidth: 0.5,
		})
		totalH += fontSize * 0.2
	} else if strings.HasPrefix(fnName, "under") {
		elements = append(elements, MathElement{
			Type:      ElemLine,
			LineX1:    0,
			LineY1:    -fontSize * 0.1,
			LineX2:    baseLay.Width,
			LineY2:    -fontSize * 0.1,
			LineWidth: 0.5,
		})
	}

	// Annotation text (if present)
	if len(node.Args) >= 2 {
		annLay := le.layoutNode(node.Args[1], fontSize*0.7)
		annX := (baseLay.Width - annLay.Width) / 2
		annY := -fontSize * 0.3
		if strings.HasPrefix(fnName, "over") {
			annY = baseLay.Height + fontSize*0.3
			totalH += annLay.Height + fontSize*0.2
		}
		for _, el := range annLay.Elements {
			el.X += annX
			el.Y += annY
			elements = append(elements, el)
		}
	}

	return &MathLayout{Width: baseLay.Width, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutPassthrough(node *Node, fontSize float64) *MathLayout {
	if len(node.Children) >= 1 {
		return le.layoutNode(node.Children[0], fontSize)
	}
	return &MathLayout{}
}

func (le *LayoutEngine) layoutOp(node *Node, fontSize float64) *MathLayout {
	text := node.FuncName
	if text == "" && node.Value != "" {
		text = node.Value
	}
	if text == "" && len(node.Children) >= 1 {
		text = FlattenToText(node.Children[0])
	}
	return le.layoutText(text, fontSize)
}

func (le *LayoutEngine) layoutStretch(node *Node, fontSize float64) *MathLayout {
	if len(node.Children) >= 1 {
		lay := le.layoutNode(node.Children[0], fontSize)
		lay.Width *= 2 // Simple stretch approximation
		return lay
	}
	return &MathLayout{}
}

func (le *LayoutEngine) layoutSequence(nodes []*Node, fontSize float64) *MathLayout {
	var elements []MathElement
	var totalW float64
	var maxH float64
	var maxD float64

	for _, child := range nodes {
		childLay := le.layoutNode(child, fontSize)
		for _, el := range childLay.Elements {
			el.X += totalW
			elements = append(elements, el)
		}
		totalW += childLay.Width
		if childLay.Height > maxH {
			maxH = childLay.Height
		}
		if childLay.Depth > maxD {
			maxD = childLay.Depth
		}
	}

	return &MathLayout{Width: totalW, Height: maxH, Depth: maxD, Elements: elements}
}

func (le *LayoutEngine) layoutGenericFunc(node *Node, fontSize float64) *MathLayout {
	// Render as: funcName(arg1, arg2, ...)
	nameLay := le.layoutText(node.FuncName, fontSize)
	delimW := fontSize * 0.3

	totalW := nameLay.Width + delimW // name + (
	elements := make([]MathElement, 0)
	elements = append(elements, nameLay.Elements...)

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: "(", FontSize: fontSize,
		X: nameLay.Width, Width: delimW,
	})

	maxH := nameLay.Height
	for i, arg := range node.Args {
		if i > 0 {
			elements = append(elements, MathElement{
				Type: ElemGlyph, Text: ", ", FontSize: fontSize,
				X: totalW, Width: fontSize * 0.3,
			})
			totalW += fontSize * 0.3
		}
		argLay := le.layoutNode(arg, fontSize)
		for _, el := range argLay.Elements {
			el.X += totalW
			elements = append(elements, el)
		}
		totalW += argLay.Width
		if argLay.Height > maxH {
			maxH = argLay.Height
		}
	}

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: ")", FontSize: fontSize,
		X: totalW, Width: delimW,
	})
	totalW += delimW

	return &MathLayout{Width: totalW, Height: maxH, Elements: elements}
}

// RenderToContentStream renders a MathLayout to PDF content stream operations.
func RenderToContentStream(buf *bytes.Buffer, layout *MathLayout, ctx *RenderContext) {
	baseX := ctx.X
	baseY := ctx.Y

	renderElements(buf, layout.Elements, baseX, baseY, ctx)
}

func renderElements(buf *bytes.Buffer, elements []MathElement, baseX, baseY float64, ctx *RenderContext) {
	for _, el := range elements {
		switch el.Type {
		case ElemGlyph:
			if el.Text == "" || el.Text == " " {
				continue
			}
			fontSize := el.FontSize
			if fontSize <= 0 {
				fontSize = ctx.FontSize
			}
			fontRef := ctx.FontRef
			if el.FontRef != "" {
				fontRef = el.FontRef
			}

			x := baseX + el.X
			y := baseY + el.Y

			buf.WriteString("BT\n")
			buf.WriteString(fontRef)
			buf.WriteString(" ")
			buf.WriteString(fmtFloat(fontSize))
			buf.WriteString(" Tf\n")
			if ctx.TextColor != "" {
				buf.WriteString(ctx.TextColor)
				buf.WriteString(" rg\n")
			}
			buf.WriteString("1 0 0 1 0 0 Tm\n")
			buf.WriteString(fmtFloat(x))
			buf.WriteString(" ")
			buf.WriteString(fmtFloat(y))
			buf.WriteString(" Td\n")

			// Format text for PDF
			if ctx.FormatText != nil {
				buf.WriteString(ctx.FormatText(el.Text))
			} else {
				buf.WriteString("(")
				buf.WriteString(escapePDFText(el.Text))
				buf.WriteString(")")
			}
			buf.WriteString(" Tj\n")
			buf.WriteString("ET\n")

		case ElemLine:
			x1 := baseX + el.LineX1
			y1 := baseY + el.LineY1
			x2 := baseX + el.LineX2
			y2 := baseY + el.LineY2
			lineW := el.LineWidth
			if lineW <= 0 {
				lineW = 0.5
			}

			buf.WriteString("q\n")
			if ctx.TextColor != "" {
				buf.WriteString(ctx.TextColor)
				buf.WriteString(" RG\n")
			}
			buf.WriteString(fmtFloat(lineW))
			buf.WriteString(" w\n")
			buf.WriteString(fmtFloat(x1))
			buf.WriteString(" ")
			buf.WriteString(fmtFloat(y1))
			buf.WriteString(" m ")
			buf.WriteString(fmtFloat(x2))
			buf.WriteString(" ")
			buf.WriteString(fmtFloat(y2))
			buf.WriteString(" l S\n")
			buf.WriteString("Q\n")

		case ElemGroup:
			renderElements(buf, el.Children, baseX+el.X, baseY+el.Y, ctx)
		}
	}
}

// fmtFloat formats a float64 for PDF with 2 decimal places.
func fmtFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

// escapePDFText escapes special characters for PDF literal strings.
func escapePDFText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
