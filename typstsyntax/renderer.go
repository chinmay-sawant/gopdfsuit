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

// offsetElement shifts all coordinate fields (X/Y for glyphs, LineX/LineY for lines) by dx, dy.
// For ElemGroup elements, only the group's X/Y is shifted since children are positioned
// relative to the group origin and rendered via baseX+group.X in renderElements.
func offsetElement(el *MathElement, dx, dy float64) {
	el.X += dx
	el.Y += dy
	if el.Type == ElemLine {
		el.LineX1 += dx
		el.LineY1 += dy
		el.LineX2 += dx
		el.LineY2 += dy
	}
	// Only recurse into children for non-group elements.
	// ElemGroup children are rendered relative to group.X/group.Y,
	// so offsetting both would double the displacement.
	if el.Type != ElemGroup {
		for i := range el.Children {
			offsetElement(&el.Children[i], dx, dy)
		}
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
	if len(node.Children) >= 2 {
		if lay, ok := le.layoutBigOperatorLimits(node.Children[0], node.Children[1], nil, fontSize); ok {
			return lay
		}
	}

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
		offsetElement(&el, baseLay.Width, supRise)
		elements = append(elements, el)
	}

	return &MathLayout{Width: totalW, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutSubscript(node *Node, fontSize float64) *MathLayout {
	if len(node.Children) >= 2 {
		if lay, ok := le.layoutBigOperatorLimits(node.Children[0], nil, node.Children[1], fontSize); ok {
			return lay
		}
	}

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
		offsetElement(&el, baseLay.Width, -subDrop)
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
		offsetElement(&el, numX, numY)
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
		offsetElement(&el, denX, denY)
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
		offsetElement(&el, radWidth, 0)
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
		offsetElement(&el, fontSize*0.05, totalH-indexLay.Height*0.5)
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
		offsetElement(&el, radWidth, 0)
		elements = append(elements, el)
	}

	return &MathLayout{Width: totalW, Height: totalH, Elements: elements}
}

func (le *LayoutEngine) layoutGroup(node *Node, fontSize float64) *MathLayout {
	innerLay := le.layoutSequence(node.Children, fontSize)

	delims := getGroupDelimiters(node.Value)
	delimWidth := fontSize * 0.3

	// Scale delimiters conservatively to avoid oversized brackets
	delimFontSize := fontSize
	innerSpan := innerLay.Height + innerLay.Depth
	if innerSpan > fontSize*1.5 {
		ratio := innerSpan / fontSize
		scaleFactor := 1.0 + (ratio-1.0)*0.3
		if scaleFactor > 1.8 {
			scaleFactor = 1.8
		}
		delimFontSize = fontSize * scaleFactor
	}

	totalW := delimWidth*2 + innerLay.Width
	elements := make([]MathElement, 0)

	// Left delimiter
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[0], FontSize: delimFontSize,
		X: 0, Width: delimWidth,
	})

	// Inner content, shifted right
	for _, el := range innerLay.Elements {
		offsetElement(&el, delimWidth, 0)
		elements = append(elements, el)
	}

	// Right delimiter
	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[1], FontSize: delimFontSize,
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

// layoutMatrixGrid lays out the matrix grid content without any surrounding
// delimiters. Returns the grid layout plus bracket top/bottom coordinates
// for external delimiter placement.
func (le *LayoutEngine) layoutMatrixGrid(node *Node, fontSize float64) (grid *MathLayout, bracketTop, bracketBottom float64) {
	if len(node.Args) == 0 {
		lay := le.layoutText("[]", fontSize)
		return lay, fontSize, 0
	}

	elemFontSize := fontSize * 0.82
	cols := inferMatrixColumns(len(node.Args))
	rowCount := int(math.Ceil(float64(len(node.Args)) / float64(cols)))

	gridCells := make([][]*MathLayout, rowCount)
	colWidths := make([]float64, cols)
	rowHeights := make([]float64, rowCount)

	idx := 0
	for r := 0; r < rowCount; r++ {
		gridCells[r] = make([]*MathLayout, cols)
		for c := 0; c < cols; c++ {
			if idx >= len(node.Args) {
				break
			}
			cellLay := le.layoutNode(node.Args[idx], elemFontSize)
			gridCells[r][c] = cellLay
			if cellLay.Width > colWidths[c] {
				colWidths[c] = cellLay.Width
			}
			cellSpan := cellLay.Height + cellLay.Depth
			if cellSpan > rowHeights[r] {
				rowHeights[r] = cellSpan
			}
			idx++
		}
	}

	colGap := fontSize * 0.45
	rowGap := fontSize * 0.30

	// Calculate total grid height
	totalGridHeight := 0.0
	for r := 0; r < rowCount; r++ {
		totalGridHeight += rowHeights[r]
		if r > 0 {
			totalGridHeight += rowGap
		}
	}

	// Center grid around math axis (~0.25*fontSize above baseline)
	mathAxis := fontSize * 0.25
	yShift := mathAxis + (totalGridHeight-rowHeights[0])/2

	// Bracket top/bottom relative to baseline
	bTop := yShift + rowHeights[0]*0.7
	bBottom := yShift - totalGridHeight + rowHeights[0]*0.3

	innerW := 0.0
	for c := 0; c < cols; c++ {
		innerW += colWidths[c]
		if c > 0 {
			innerW += colGap
		}
	}

	var elements []MathElement

	// Column X positions (no delimiter offset — grid starts at x=0)
	padding := fontSize * 0.12
	colX := make([]float64, cols)
	x := padding
	for c := 0; c < cols; c++ {
		colX[c] = x
		x += colWidths[c] + colGap
	}

	y := yShift
	for r := 0; r < rowCount; r++ {
		if r > 0 {
			y -= rowHeights[r-1] + rowGap
		}
		for c := 0; c < cols; c++ {
			cellLay := gridCells[r][c]
			if cellLay == nil {
				continue
			}
			offsetX := colX[c] + (colWidths[c]-cellLay.Width)/2
			for _, el := range cellLay.Elements {
				offsetElement(&el, offsetX, y)
				elements = append(elements, el)
			}
		}
	}

	totalW := padding + innerW + padding
	halfSpan := totalGridHeight/2 + fontSize*0.35

	lay := &MathLayout{Width: totalW, Height: halfSpan + mathAxis, Depth: halfSpan - mathAxis, Elements: elements}
	return lay, bTop, bBottom
}

func (le *LayoutEngine) layoutMatrix(node *Node, fontSize float64) *MathLayout {
	gridLay, bracketTop, bracketBottom := le.layoutMatrixGrid(node, fontSize)
	if len(node.Args) == 0 {
		return gridLay
	}

	delimW := fontSize * 0.35

	var elements []MathElement

	// Left paren drawn as thin lines
	elements = append(elements, makeParenLeft(delimW*0.8, bracketTop, bracketBottom))

	// Inner grid content, shifted right past the left delimiter
	for _, el := range gridLay.Elements {
		offsetElement(&el, delimW, 0)
		elements = append(elements, el)
	}

	rightX := delimW + gridLay.Width

	// Right paren drawn as thin lines
	elements = append(elements, makeParenRight(rightX+delimW*0.2, bracketTop, bracketBottom))

	totalW := rightX + delimW

	return &MathLayout{Width: totalW, Height: gridLay.Height, Depth: gridLay.Depth, Elements: elements}
}

func (le *LayoutEngine) layoutVector(node *Node, fontSize float64) *MathLayout {
	if len(node.Args) == 0 {
		return le.layoutText("[]", fontSize)
	}

	elemFontSize := fontSize * 0.85
	innerW := 0.0
	rowHeights := make([]float64, len(node.Args))
	argLayouts := make([]*MathLayout, len(node.Args))

	for i, arg := range node.Args {
		lay := le.layoutNode(arg, elemFontSize)
		argLayouts[i] = lay
		if lay.Width > innerW {
			innerW = lay.Width
		}
		rowHeights[i] = lay.Height + lay.Depth
	}

	rowGap := fontSize * 0.30
	delimW := fontSize * 0.35
	contentX := delimW + fontSize*0.15

	// Calculate total grid height for dynamic bracket sizing and centering
	totalGridHeight := 0.0
	for i := range rowHeights {
		totalGridHeight += rowHeights[i]
		if i > 0 {
			totalGridHeight += rowGap
		}
	}

	// Center grid around math axis (~0.25*fontSize above baseline)
	mathAxis := fontSize * 0.25
	yShift := mathAxis + (totalGridHeight-rowHeights[0])/2

	// Bracket top/bottom relative to baseline
	bracketTop := yShift + rowHeights[0]*0.7
	bracketBottom := yShift - totalGridHeight + rowHeights[0]*0.3
	serifLen := fontSize * 0.15

	var elements []MathElement

	// Left square bracket drawn as thin lines
	elements = append(elements, makeSquareBracketLeft(delimW*0.3, bracketTop, bracketBottom, serifLen))

	y := yShift
	for i, lay := range argLayouts {
		if i > 0 {
			y -= rowHeights[i-1] + rowGap
		}
		x := contentX + (innerW-lay.Width)/2
		for _, el := range lay.Elements {
			offsetElement(&el, x, y)
			elements = append(elements, el)
		}
	}

	rightX := contentX + innerW + fontSize*0.10

	// Right square bracket drawn as thin lines
	elements = append(elements, makeSquareBracketRight(rightX+delimW*0.7, bracketTop, bracketBottom, serifLen))

	totalW := rightX + delimW
	halfSpan := totalGridHeight/2 + fontSize*0.35

	return &MathLayout{Width: totalW, Height: halfSpan + mathAxis, Depth: halfSpan - mathAxis, Elements: elements}
}

// findSingleMatrix looks through the LR node's args for a single matrix node,
// skipping whitespace literals and sequence wrappers.
func findSingleMatrix(args []*Node) *Node {
	if len(args) == 1 {
		arg := args[0]
		if arg.Type == NodeMatrix {
			return arg
		}
		// Unwrap sequence: look for a single non-space child that is a matrix
		if arg.Type == NodeSequence {
			var matNode *Node
			for _, c := range arg.Children {
				if c.Type == NodeLiteral && strings.TrimSpace(c.Value) == "" {
					continue
				}
				if c.Type == NodeMatrix && matNode == nil {
					matNode = c
				} else {
					return nil // multiple non-space children
				}
			}
			return matNode
		}
	}
	return nil
}

func inferMatrixColumns(argCount int) int {
	if argCount <= 1 {
		return 1
	}
	root := int(math.Round(math.Sqrt(float64(argCount))))
	if root > 1 && root*root == argCount {
		return root
	}
	if argCount%3 == 0 && argCount >= 6 {
		return 3
	}
	if argCount%2 == 0 {
		return 2
	}
	return 1
}

func (le *LayoutEngine) layoutBigOperatorLimits(baseNode, supNode, subNode *Node, fontSize float64) (*MathLayout, bool) {
	// Common parse shape for sum/product with both limits is:
	// NodeSuperscript(NodeSubscript(op, lower), upper)
	if subNode == nil && baseNode != nil && baseNode.Type == NodeSubscript && len(baseNode.Children) >= 2 {
		candidateBase := baseNode.Children[0]
		if isBigLimitOperator(candidateBase) {
			subNode = baseNode.Children[1]
			baseNode = candidateBase
		}
	}

	if !isBigLimitOperator(baseNode) {
		return nil, false
	}

	opLay := le.layoutNode(baseNode, fontSize*1.15)
	scriptSize := fontSize * 0.62

	var supLay, subLay *MathLayout
	if supNode != nil {
		supLay = le.layoutNode(supNode, scriptSize)
	}
	if subNode != nil {
		subLay = le.layoutNode(subNode, scriptSize)
	}

	width := opLay.Width
	if supLay != nil {
		width = math.Max(width, supLay.Width)
	}
	if subLay != nil {
		width = math.Max(width, subLay.Width)
	}

	gap := fontSize * 0.14
	var elements []MathElement

	opX := (width - opLay.Width) / 2
	for _, el := range opLay.Elements {
		offsetElement(&el, opX, 0)
		elements = append(elements, el)
	}

	height := opLay.Height
	depth := opLay.Depth

	if supLay != nil {
		supX := (width - supLay.Width) / 2
		supY := opLay.Height + gap
		for _, el := range supLay.Elements {
			offsetElement(&el, supX, supY)
			elements = append(elements, el)
		}
		height = math.Max(height, supY+supLay.Height)
	}

	if subLay != nil {
		subX := (width - subLay.Width) / 2
		subY := -(gap + subLay.Height)
		for _, el := range subLay.Elements {
			offsetElement(&el, subX, subY)
			elements = append(elements, el)
		}
		depth = math.Max(depth, gap+subLay.Height+subLay.Depth)
	}

	return &MathLayout{Width: width, Height: height, Depth: depth, Elements: elements}, true
}

func isBigLimitOperator(node *Node) bool {
	if node == nil || node.Type != NodeSymbol {
		return false
	}
	return node.Value == "∑" || node.Value == "∏"
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
		offsetElement(&el, topX, topY)
		elements = append(elements, el)
	}

	// Bottom
	botX := delimW + (innerW-botLay.Width)/2
	for _, el := range botLay.Elements {
		offsetElement(&el, botX, botY)
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
			offsetElement(&el, delimW+fontSize*0.2, yOffset)
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

	// Special case: if the only child is a matrix (possibly wrapped in a sequence
	// with whitespace), use the grid-only layout to avoid double brackets.
	matNode := findSingleMatrix(node.Args)
	if matNode != nil {
		gridLay, bracketTop, bracketBottom := le.layoutMatrixGrid(matNode, fontSize)

		innerSpan := gridLay.Height + gridLay.Depth
		delimFontSize := lrDelimFontSize(innerSpan, fontSize)
		delimW := delimFontSize * 0.35

		var elements []MathElement

		// Use line-drawn parentheses for "(" ")" to keep them thin
		if delims[0] == "(" && delims[1] == ")" {
			elements = append(elements, makeParenLeft(delimW*0.8, bracketTop, bracketBottom))

			for _, el := range gridLay.Elements {
				offsetElement(&el, delimW, 0)
				elements = append(elements, el)
			}

			rightX := delimW + gridLay.Width
			elements = append(elements, makeParenRight(rightX+delimW*0.2, bracketTop, bracketBottom))

			totalW := rightX + delimW
			return &MathLayout{Width: totalW, Height: gridLay.Height, Depth: gridLay.Depth, Elements: elements}
		}

		// For other delimiters (|, ‖, ⌊, ⌋, etc.) use font glyphs
		delimY := (gridLay.Height-gridLay.Depth)/2 - delimFontSize*0.35

		elements = append(elements, MathElement{
			Type: ElemGlyph, Text: delims[0], FontSize: delimFontSize,
			X: 0, Y: delimY, Width: delimW,
		})

		for _, el := range gridLay.Elements {
			offsetElement(&el, delimW, 0)
			elements = append(elements, el)
		}

		rightX := delimW + gridLay.Width
		elements = append(elements, MathElement{
			Type: ElemGlyph, Text: delims[1], FontSize: delimFontSize,
			X: rightX, Y: delimY, Width: delimW,
		})

		totalW := rightX + delimW
		return &MathLayout{Width: totalW, Height: gridLay.Height, Depth: gridLay.Depth, Elements: elements}
	}

	innerLay := le.layoutSequence(node.Args, fontSize)

	innerSpan := innerLay.Height + innerLay.Depth
	delimFontSize := lrDelimFontSize(innerSpan, fontSize)
	delimW := delimFontSize * 0.35

	totalW := delimW*2 + innerLay.Width

	elements := make([]MathElement, 0)

	// Position delimiter baseline so the glyph is vertically centered on the content
	delimY := 0.0
	if delimFontSize > fontSize*1.2 {
		delimY = -innerLay.Depth - delimFontSize*0.25
	}

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[0], FontSize: delimFontSize,
		X: 0, Y: delimY, Width: delimW,
	})

	for _, el := range innerLay.Elements {
		offsetElement(&el, delimW, 0)
		elements = append(elements, el)
	}

	elements = append(elements, MathElement{
		Type: ElemGlyph, Text: delims[1], FontSize: delimFontSize,
		X: delimW + innerLay.Width, Y: delimY, Width: delimW,
	})

	return &MathLayout{Width: totalW, Height: innerLay.Height, Depth: innerLay.Depth, Elements: elements}
}

// lrDelimFontSize computes the delimiter font size for lr() based on content height.
func lrDelimFontSize(innerSpan, fontSize float64) float64 {
	if innerSpan <= fontSize*1.2 {
		return fontSize
	}
	// A parenthesis glyph typically spans ~1.15x the em size.
	// Scale so the rendered glyph roughly matches the content height.
	return innerSpan * 0.85
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
			offsetElement(&el, annX, annY)
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
			offsetElement(&el, totalW, 0)
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
			offsetElement(&el, totalW, 0)
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

// bracketLineWidth is a constant thin line width for all drawn brackets.
const bracketLineWidth = 0.4

// parenLineWidth is a thinner line width specifically for curved parentheses.
const parenLineWidth = 0.15

// makeSquareBracketLeft draws a "[" bracket as 3 lines: top serif, vertical, bottom serif.
// x is the right edge of the bracket area, top/bottom are relative Y coords.
func makeSquareBracketLeft(x, top, bottom, serifLen float64) MathElement {
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			// vertical line
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			// top serif
			{Type: ElemLine, LineX1: x, LineY1: top, LineX2: x + serifLen, LineY2: top, LineWidth: bracketLineWidth},
			// bottom serif
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x + serifLen, LineY2: bottom, LineWidth: bracketLineWidth},
		},
	}
}

// makeSquareBracketRight draws a "]" bracket as 3 lines.
func makeSquareBracketRight(x, top, bottom, serifLen float64) MathElement {
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			// vertical line
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			// top serif
			{Type: ElemLine, LineX1: x - serifLen, LineY1: top, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			// bottom serif
			{Type: ElemLine, LineX1: x - serifLen, LineY1: bottom, LineX2: x, LineY2: bottom, LineWidth: bracketLineWidth},
		},
	}
}

// makeParenLeft draws a "(" as a series of line segments approximating a curve.
func makeParenLeft(x, top, bottom float64) MathElement {
	mid := (top + bottom) / 2
	h := top - bottom
	// curve inward by a fraction of height
	curve := h * 0.12
	if curve < 1.5 {
		curve = 1.5
	}
	// approximate with 4 segments
	q1 := bottom + h*0.25
	q3 := bottom + h*0.75
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			{Type: ElemLine, LineX1: x, LineY1: top, LineX2: x - curve*0.5, LineY2: q3, LineWidth: parenLineWidth},
			{Type: ElemLine, LineX1: x - curve*0.5, LineY1: q3, LineX2: x - curve, LineY2: mid, LineWidth: parenLineWidth},
			{Type: ElemLine, LineX1: x - curve, LineY1: mid, LineX2: x - curve*0.5, LineY2: q1, LineWidth: parenLineWidth},
			{Type: ElemLine, LineX1: x - curve*0.5, LineY1: q1, LineX2: x, LineY2: bottom, LineWidth: parenLineWidth},
		},
	}
}

// makeParenRight draws a ")" as a series of line segments approximating a curve.
func makeParenRight(x, top, bottom float64) MathElement {
	mid := (top + bottom) / 2
	h := top - bottom
	curve := h * 0.12
	if curve < 1.5 {
		curve = 1.5
	}
	q1 := bottom + h*0.25
	q3 := bottom + h*0.75
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			{Type: ElemLine, LineX1: x, LineY1: top, LineX2: x + curve*0.5, LineY2: q3, LineWidth: parenLineWidth},
			{Type: ElemLine, LineX1: x + curve*0.5, LineY1: q3, LineX2: x + curve, LineY2: mid, LineWidth: parenLineWidth},
			{Type: ElemLine, LineX1: x + curve, LineY1: mid, LineX2: x + curve*0.5, LineY2: q1, LineWidth: parenLineWidth},
			{Type: ElemLine, LineX1: x + curve*0.5, LineY1: q1, LineX2: x, LineY2: bottom, LineWidth: parenLineWidth},
		},
	}
}

// makeVerticalLine draws a simple vertical line (used for |, ‖, etc.)
func makeVerticalLine(x, top, bottom float64) MathElement {
	return MathElement{
		Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth,
	}
}

// makeFloorLeft draws "⌊" — vertical line + bottom serif.
func makeFloorLeft(x, top, bottom, serifLen float64) MathElement {
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x + serifLen, LineY2: bottom, LineWidth: bracketLineWidth},
		},
	}
}

// makeFloorRight draws "⌋" — vertical line + bottom serif.
func makeFloorRight(x, top, bottom, serifLen float64) MathElement {
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			{Type: ElemLine, LineX1: x - serifLen, LineY1: bottom, LineX2: x, LineY2: bottom, LineWidth: bracketLineWidth},
		},
	}
}

// makeCeilLeft draws "⌈" — vertical line + top serif.
func makeCeilLeft(x, top, bottom, serifLen float64) MathElement {
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			{Type: ElemLine, LineX1: x, LineY1: top, LineX2: x + serifLen, LineY2: top, LineWidth: bracketLineWidth},
		},
	}
}

// makeCeilRight draws "⌉" — vertical line + top serif.
func makeCeilRight(x, top, bottom, serifLen float64) MathElement {
	return MathElement{
		Type: ElemGroup, X: 0, Y: 0,
		Children: []MathElement{
			{Type: ElemLine, LineX1: x, LineY1: bottom, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
			{Type: ElemLine, LineX1: x - serifLen, LineY1: top, LineX2: x, LineY2: top, LineWidth: bracketLineWidth},
		},
	}
}

// makeDelimiterElement draws any delimiter as thin lines at the given position.
func makeDelimiterElement(delim string, xStart, delimW, top, bottom, fontSize float64) MathElement {
	serifLen := fontSize * 0.15
	switch delim {
	case "(":
		return makeParenLeft(xStart+delimW*0.8, top, bottom)
	case ")":
		return makeParenRight(xStart+delimW*0.2, top, bottom)
	case "[":
		return makeSquareBracketLeft(xStart+delimW*0.3, top, bottom, serifLen)
	case "]":
		return makeSquareBracketRight(xStart+delimW*0.7, top, bottom, serifLen)
	case "|":
		return makeVerticalLine(xStart+delimW*0.5, top, bottom)
	case "‖":
		gap := fontSize * 0.08
		return MathElement{
			Type: ElemGroup, X: 0, Y: 0,
			Children: []MathElement{
				{Type: ElemLine, LineX1: xStart + delimW*0.5 - gap, LineY1: bottom, LineX2: xStart + delimW*0.5 - gap, LineY2: top, LineWidth: bracketLineWidth},
				{Type: ElemLine, LineX1: xStart + delimW*0.5 + gap, LineY1: bottom, LineX2: xStart + delimW*0.5 + gap, LineY2: top, LineWidth: bracketLineWidth},
			},
		}
	case "⌊":
		return makeFloorLeft(xStart+delimW*0.3, top, bottom, serifLen)
	case "⌋":
		return makeFloorRight(xStart+delimW*0.7, top, bottom, serifLen)
	case "⌈":
		return makeCeilLeft(xStart+delimW*0.3, top, bottom, serifLen)
	case "⌉":
		return makeCeilRight(xStart+delimW*0.7, top, bottom, serifLen)
	default:
		// Fallback: draw as vertical line
		return makeVerticalLine(xStart+delimW*0.5, top, bottom)
	}
}

// escapePDFText escapes special characters for PDF literal strings.
func escapePDFText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
