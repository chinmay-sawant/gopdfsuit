package typstsyntax

import (
	"strings"
)

// NodeType represents the type of an AST node.
type NodeType int

// NodeType constants.
const (
	NodeLiteral     NodeType = iota // Plain text or number
	NodeSymbol                      // Resolved symbol (e.g., pi → π)
	NodeSuperscript                 // Superscript: base^exponent
	NodeSubscript                   // Subscript: base_subscript
	NodeFraction                    // Fraction: frac(num, denom) or a/b
	NodeSqrt                        // Square root: sqrt(x)
	NodeRoot                        // Nth root: root(n, x)
	NodeGroup                       // Parenthesized group: (...)
	NodeFunc                        // Function call: func(args...)
	NodeAccent                      // Accent: hat(x), tilde(x), etc.
	NodeMatrix                      // Matrix: mat(1,2; 3,4)
	NodeVector                      // Vector: vec(a, b, c)
	NodeBinom                       // Binomial: binom(n, k)
	NodeCases                       // Cases: cases(...)
	NodeOperator                    // Infix operator: +, -, =, etc.
	NodeQuotedText                  // "text" verbatim
	NodePrime                       // Prime marks (′, ″, etc.)
	NodeUnderOver                   // underline, overline, underbrace, etc.
	NodeCancel                      // cancel(x)
	NodeLR                          // lr(), abs(), norm(), floor(), ceil()
	NodeSequence                    // Sequence of nodes
	NodeStyle                       // bold(), italic(), upright()
	NodeVariant                     // sans(), frak(), mono(), bb(), cal()
	NodeSize                        // display(), inline(), script(), sscript()
	NodeOp                          // op("custom")
	NodeStretch                     // stretch(=, size: 2em)
	NodeClass                       // class("relation", x)
	NodeLineBreak                   // \\
	NodeAlign                       // &
)

// Node represents a node in the Typst math AST.
type Node struct {
	Type     NodeType
	Value    string  // For literals, symbols, operators
	Children []*Node // Sub-expressions
	// Function-specific fields
	FuncName string            // Name of the function (e.g., "frac", "sqrt")
	Args     []*Node           // Function arguments
	Options  map[string]string // Named options (e.g., delim, gap)
}

// Parser converts a token stream into an AST.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser from tokens.
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Parse parses the entire token stream into an AST.
func (p *Parser) Parse() *Node {
	nodes := p.parseSequence()
	if len(nodes) == 1 {
		return nodes[0]
	}
	return &Node{Type: NodeSequence, Children: nodes}
}

func (p *Parser) current() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) advance() Token {
	tok := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) expect(typ TokenType) Token {
	tok := p.current()
	if tok.Type == typ {
		p.advance()
	}
	return tok
}

// parseSequence parses a sequence of atoms until EOF or a closing delimiter.
func (p *Parser) parseSequence() []*Node {
	var nodes []*Node
	for {
		tok := p.current()
		if tok.Type == TokenEOF || tok.Type == TokenRParen || tok.Type == TokenRBrace || tok.Type == TokenRBracket {
			break
		}
		node := p.parseExpr()
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// parseExpr parses a single expression (atom with optional postfix like ^, _, /)
func (p *Parser) parseExpr() *Node {
	left := p.parseAtom()
	if left == nil {
		return nil
	}

	// Handle postfix operators
	for {
		tok := p.current()
		switch tok.Type {
		case TokenSuperscript:
			p.advance()
			right := p.parseAtom()
			// Unwrap paren grouping — parens after ^ are for grouping, not display
			if right != nil && right.Type == NodeGroup && right.Value == "" {
				right = &Node{Type: NodeSequence, Children: right.Children}
			}
			left = &Node{Type: NodeSuperscript, Children: []*Node{left, right}}
		case TokenSubscript:
			p.advance()
			right := p.parseAtom()
			// Unwrap paren grouping — parens after _ are for grouping, not display
			if right != nil && right.Type == NodeGroup && right.Value == "" {
				right = &Node{Type: NodeSequence, Children: right.Children}
			}
			left = &Node{Type: NodeSubscript, Children: []*Node{left, right}}
		case TokenSlash:
			p.advance()
			right := p.parseAtom()
			left = &Node{Type: NodeFraction, FuncName: "frac", Children: []*Node{left, right}}
		case TokenPrime:
			tok = p.advance()
			left = &Node{Type: NodePrime, Value: tok.Value, Children: []*Node{left}}
		default:
			return left
		}
	}
}

// parseAtom parses a single atomic expression.
func (p *Parser) parseAtom() *Node {
	tok := p.current()

	switch tok.Type {
	case TokenNumber:
		p.advance()
		return &Node{Type: NodeLiteral, Value: tok.Value}

	case TokenSymbol:
		p.advance()
		// Resolve symbol to Unicode
		if unicode, ok := TypstSymbols[tok.Value]; ok {
			return &Node{Type: NodeSymbol, Value: unicode}
		}
		return &Node{Type: NodeLiteral, Value: tok.Value}

	case TokenText:
		p.advance()
		// Check if it's a function call
		if p.current().Type == TokenLParen {
			return p.parseFunctionCall(tok.Value)
		}
		// Check if it's a known symbol
		if unicode, ok := TypstSymbols[tok.Value]; ok {
			return &Node{Type: NodeSymbol, Value: unicode}
		}
		return &Node{Type: NodeLiteral, Value: tok.Value}

	case TokenQuotedText:
		p.advance()
		return &Node{Type: NodeQuotedText, Value: tok.Value}

	case TokenOperator:
		p.advance()
		return &Node{Type: NodeOperator, Value: tok.Value}

	case TokenLParen:
		return p.parseParenGroup()

	case TokenLBracket:
		return p.parseBracketGroup()

	case TokenLBrace:
		return p.parseBraceGroup()

	case TokenBackslash:
		p.advance()
		return &Node{Type: NodeLineBreak}

	case TokenAmpersand:
		p.advance()
		return &Node{Type: NodeAlign}

	case TokenWhitespace:
		p.advance()
		return &Node{Type: NodeLiteral, Value: " "}

	case TokenHash:
		p.advance()
		// Skip code-level references for now
		if p.current().Type == TokenText {
			p.advance()
		}
		return nil

	default:
		p.advance()
		return &Node{Type: NodeLiteral, Value: tok.Value}
	}
}

// parseFunctionCall parses a function invocation like frac(a, b).
func (p *Parser) parseFunctionCall(name string) *Node {
	p.expect(TokenLParen) // consume (

	var args []*Node
	for p.current().Type != TokenRParen && p.current().Type != TokenEOF {
		arg := p.parseFunctionArgSequence()
		if len(arg) == 1 {
			args = append(args, arg[0])
		} else if len(arg) > 1 {
			args = append(args, &Node{Type: NodeSequence, Children: arg})
		}
		if p.current().Type == TokenComma {
			p.advance()
		} else if p.current().Type == TokenSemicolon {
			// For matrix/vector: semicolons separate rows
			p.advance()
		}
	}
	p.expect(TokenRParen) // consume )

	return p.buildFuncNode(name, args)
}

// parseFunctionArgSequence parses a function argument until a top-level
// separator (comma/semicolon), closing parenthesis, or EOF.
func (p *Parser) parseFunctionArgSequence() []*Node {
	var nodes []*Node
	for {
		tok := p.current()
		if tok.Type == TokenEOF || tok.Type == TokenRParen || tok.Type == TokenComma || tok.Type == TokenSemicolon {
			break
		}
		node := p.parseExpr()
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// buildFuncNode creates the appropriate AST node for a function call.
func (p *Parser) buildFuncNode(name string, args []*Node) *Node {
	switch name {
	case "frac":
		if len(args) >= 2 {
			return &Node{Type: NodeFraction, FuncName: "frac", Children: args[:2]}
		}
	case "sqrt":
		if len(args) >= 1 {
			return &Node{Type: NodeSqrt, FuncName: "sqrt", Children: args[:1]}
		}
	case "root":
		if len(args) >= 2 {
			return &Node{Type: NodeRoot, FuncName: "root", Children: args[:2]}
		}
	case "mat":
		return &Node{Type: NodeMatrix, FuncName: "mat", Args: args}
	case "vec":
		return &Node{Type: NodeVector, FuncName: "vec", Args: args}
	case "binom":
		if len(args) >= 2 {
			return &Node{Type: NodeBinom, FuncName: "binom", Children: args[:2]}
		}
	case "cases":
		return &Node{Type: NodeCases, FuncName: "cases", Args: args}
	case "cancel":
		if len(args) >= 1 {
			return &Node{Type: NodeCancel, FuncName: "cancel", Children: args[:1]}
		}
	}

	if node := p.tryBuildLRNode(name, args); node != nil {
		return node
	}
	if node := p.tryBuildAccentNode(name, args); node != nil {
		return node
	}
	if node := p.tryBuildUnderOverNode(name, args); node != nil {
		return node
	}
	if node := p.tryBuildStyleOrVariantNode(name, args); node != nil {
		return node
	}

	switch name {
	case "scripts", "limits":
		if len(args) >= 1 {
			return &Node{Type: NodeFunc, FuncName: name, Args: args}
		}
	case "stretch":
		if len(args) >= 1 {
			return &Node{Type: NodeStretch, FuncName: "stretch", Children: args[:1]}
		}
	case "class":
		if len(args) >= 2 {
			return &Node{Type: NodeClass, FuncName: "class", Children: args[:2]}
		}
	}

	// If it's a known predefined operator used as a function
	if PredefinedOperators[name] {
		return &Node{Type: NodeOp, FuncName: name, Value: name}
	}

	// Generic function call
	return &Node{Type: NodeFunc, FuncName: name, Args: args}
}

func (p *Parser) tryBuildLRNode(name string, args []*Node) *Node {
	switch name {
	case "lr", "abs", "norm", "floor", "ceil", "round":
		return &Node{Type: NodeLR, FuncName: name, Args: args}
	}
	return nil
}

func (p *Parser) tryBuildAccentNode(name string, args []*Node) *Node {
	switch name {
	case "hat", "tilde", "grave", "acute", "macron", "breve",
		"dot", "ddot", "dddot", "ddddot", "circle", "arrow", "bar":
		if len(args) >= 1 {
			return &Node{Type: NodeAccent, FuncName: name, Children: args[:1]}
		}
	case "accent":
		if len(args) >= 2 {
			return &Node{Type: NodeAccent, FuncName: "accent", Children: args[:2]}
		}
	}
	return nil
}

func (p *Parser) tryBuildUnderOverNode(name string, args []*Node) *Node {
	switch name {
	case "underline", "overline", "underbrace", "overbrace",
		"underbracket", "overbracket", "underparen", "overparen",
		"undershell", "overshell":
		return &Node{Type: NodeUnderOver, FuncName: name, Args: args}
	}
	return nil
}

func (p *Parser) tryBuildStyleOrVariantNode(name string, args []*Node) *Node {
	switch name {
	case "bold", "italic", "upright":
		if len(args) >= 1 {
			return &Node{Type: NodeStyle, FuncName: name, Children: args[:1]}
		}
	case "serif", "sans", "frak", "mono", "bb", "cal", "scr":
		if len(args) >= 1 {
			return &Node{Type: NodeVariant, FuncName: name, Children: args[:1]}
		}
	case "display", "inline", "script", "sscript":
		if len(args) >= 1 {
			return &Node{Type: NodeSize, FuncName: name, Children: args[:1]}
		}
	case "op":
		if len(args) >= 1 {
			return &Node{Type: NodeOp, FuncName: "op", Children: args[:1]}
		}
	}
	return nil
}

func (p *Parser) parseParenGroup() *Node {
	p.advance() // consume (
	inner := p.parseSequence()
	p.expect(TokenRParen) // consume )

	return &Node{Type: NodeGroup, Children: inner}
}

func (p *Parser) parseBracketGroup() *Node {
	p.advance() // consume [
	inner := p.parseSequence()
	p.expect(TokenRBracket)
	return &Node{Type: NodeGroup, Value: "[]", Children: inner}
}

func (p *Parser) parseBraceGroup() *Node {
	p.advance() // consume {
	inner := p.parseSequence()
	p.expect(TokenRBrace)
	return &Node{Type: NodeGroup, Value: "{}", Children: inner}
}

// FlattenToText converts an AST node tree into plain Unicode text representation.
// This is a simplified renderer that produces Unicode text suitable for standard font rendering.
func FlattenToText(node *Node) string {
	if node == nil {
		return ""
	}

	var sb strings.Builder
	flattenNode(node, &sb)
	return sb.String()
}

func flattenNode(node *Node, sb *strings.Builder) {
	switch node.Type {
	case NodeLiteral, NodeSymbol, NodeOperator, NodeQuotedText:
		sb.WriteString(node.Value)

	case NodeSuperscript:
		flattenSuperscript(node, sb)
	case NodeSubscript:
		flattenSubscript(node, sb)
	case NodeFraction:
		flattenFraction(node, sb)
	case NodeSqrt:
		flattenSqrt(node, sb)
	case NodeRoot:
		flattenRoot(node, sb)
	case NodeGroup:
		flattenGroup(node, sb)
	case NodePrime:
		flattenPrime(node, sb)
	case NodeAccent:
		flattenAccent(node, sb)
	case NodeMatrix, NodeVector:
		flattenMatrixVector(node, sb)
	case NodeBinom:
		flattenBinom(node, sb)
	case NodeCases:
		flattenCases(node, sb)
	case NodeCancel:
		flattenCancel(node, sb)
	case NodeLR:
		flattenLR(node, sb)
	case NodeUnderOver:
		flattenUnderOver(node, sb)
	case NodeStyle, NodeVariant, NodeSize, NodeStretch:
		if len(node.Children) >= 1 {
			flattenNode(node.Children[0], sb)
		}
	case NodeOp:
		flattenOp(node, sb)
	case NodeClass:
		if len(node.Children) >= 2 {
			flattenNode(node.Children[1], sb)
		}
	case NodeFunc:
		flattenFunc(node, sb)
	case NodeSequence:
		for _, child := range node.Children {
			flattenNode(child, sb)
		}
	case NodeLineBreak:
		sb.WriteString("\n")
	}
}

func flattenSuperscript(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 1 {
		flattenNode(node.Children[0], sb)
	}
	if len(node.Children) >= 2 {
		expText := FlattenToText(node.Children[1])
		sb.WriteString(toSuperscript(expText))
	}
}

func flattenSubscript(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 1 {
		flattenNode(node.Children[0], sb)
	}
	if len(node.Children) >= 2 {
		subText := FlattenToText(node.Children[1])
		sb.WriteString(toSubscript(subText))
	}
}

func flattenFraction(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 2 {
		num := FlattenToText(node.Children[0])
		den := FlattenToText(node.Children[1])
		sb.WriteString(num)
		sb.WriteString("/")
		sb.WriteString(den)
	}
}

func flattenSqrt(node *Node, sb *strings.Builder) {
	sb.WriteString("√")
	if len(node.Children) >= 1 {
		sb.WriteString("(")
		flattenNode(node.Children[0], sb)
		sb.WriteString(")")
	}
}

func flattenRoot(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 2 {
		idxText := FlattenToText(node.Children[0])
		sb.WriteString(toSuperscript(idxText))
		sb.WriteString("√")
		sb.WriteString("(")
		flattenNode(node.Children[1], sb)
		sb.WriteString(")")
	}
}

func flattenGroup(node *Node, sb *strings.Builder) {
	delims := getGroupDelimiters(node.Value)
	sb.WriteString(delims[0])
	for i, child := range node.Children {
		if i > 0 {
			sb.WriteString(" ")
		}
		flattenNode(child, sb)
	}
	sb.WriteString(delims[1])
}

func flattenPrime(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 1 {
		flattenNode(node.Children[0], sb)
	}
	sb.WriteString(node.Value)
}

func flattenAccent(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 1 {
		flattenNode(node.Children[0], sb)
		if accent, ok := AccentMap[node.FuncName]; ok {
			sb.WriteString(accent)
		}
	}
}

func flattenMatrixVector(node *Node, sb *strings.Builder) {
	openDelim, sep, closeDelim := "⌈", "; ", "⌉"
	if node.Type == NodeVector {
		openDelim, sep, closeDelim = "(", ", ", ")"
	}
	sb.WriteString(openDelim)
	for i, arg := range node.Args {
		if i > 0 {
			sb.WriteString(sep)
		}
		flattenNode(arg, sb)
	}
	sb.WriteString(closeDelim)
}

func flattenBinom(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 2 {
		sb.WriteString("(")
		flattenNode(node.Children[0], sb)
		sb.WriteString(" choose ")
		flattenNode(node.Children[1], sb)
		sb.WriteString(")")
	}
}

func flattenCases(node *Node, sb *strings.Builder) {
	sb.WriteString("{")
	for i, arg := range node.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		flattenNode(arg, sb)
	}
}

func flattenCancel(node *Node, sb *strings.Builder) {
	if len(node.Children) >= 1 {
		flattenNode(node.Children[0], sb)
	}
}

func flattenLR(node *Node, sb *strings.Builder) {
	delims := getLRDelimiters(node.FuncName)
	sb.WriteString(delims[0])
	for i, arg := range node.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		flattenNode(arg, sb)
	}
	sb.WriteString(delims[1])
}

func flattenUnderOver(node *Node, sb *strings.Builder) {
	if len(node.Args) >= 1 {
		flattenNode(node.Args[0], sb)
	}
	if len(node.Args) >= 2 {
		sb.WriteString(" (")
		flattenNode(node.Args[1], sb)
		sb.WriteString(")")
	}
}

func flattenOp(node *Node, sb *strings.Builder) {
	if node.Value != "" {
		sb.WriteString(node.Value)
	} else if len(node.Children) >= 1 {
		flattenNode(node.Children[0], sb)
	}
}

func flattenFunc(node *Node, sb *strings.Builder) {
	sb.WriteString(node.FuncName)
	sb.WriteString("(")
	for i, arg := range node.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		flattenNode(arg, sb)
	}
	sb.WriteString(")")
}

// toSuperscript converts text to Unicode superscript characters.
func toSuperscript(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if sup, ok := SuperscriptDigits[r]; ok {
			sb.WriteRune(sup)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// toSubscript converts text to Unicode subscript characters.
func toSubscript(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if sub, ok := SubscriptDigits[r]; ok {
			sb.WriteRune(sub)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func getGroupDelimiters(groupType string) [2]string {
	switch groupType {
	case "[]":
		return [2]string{"[", "]"}
	case "{}":
		return [2]string{"{", "}"}
	default:
		return [2]string{"(", ")"}
	}
}

func getLRDelimiters(funcName string) [2]string {
	switch funcName {
	case "abs":
		return [2]string{"|", "|"}
	case "norm":
		return [2]string{"‖", "‖"}
	case "floor":
		return [2]string{"⌊", "⌋"}
	case "ceil":
		return [2]string{"⌈", "⌉"}
	case "round":
		return [2]string{"⌊", "⌉"}
	default:
		return [2]string{"(", ")"}
	}
}
