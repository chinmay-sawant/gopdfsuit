// Package typstsyntax provides Typst math syntax parsing and rendering for PDF generation.
//
// When math mode is enabled on a cell (MathEnabled: true), cell text containing
// Typst math expressions like `$ A = pi r^2 $` is parsed and rendered as
// proper mathematical notation in the PDF output.
//
// Usage:
//
//	if IsMathExpression(cellText) {
//	    parsed := ParseMath(cellText)
//	    text := FlattenToText(parsed) // Unicode text representation
//	}
//
// For PDF content stream rendering:
//
//	ast := ParseMath(cellText)
//	ctx := &RenderContext{...}
//	engine := NewLayoutEngine(ctx)
//	layout := engine.Layout(ast)
//	RenderToContentStream(buf, layout, ctx)
package typstsyntax

import (
	"strings"
)

// IsMathExpression checks if a string contains Typst math syntax (wrapped in $ ... $).
func IsMathExpression(text string) bool {
	text = strings.TrimSpace(text)
	if len(text) < 3 {
		return false
	}
	if text[0] != '$' || text[len(text)-1] != '$' {
		return false
	}
	// Check that there's actual content between the delimiters
	inner := strings.TrimSpace(text[1 : len(text)-1])
	return len(inner) > 0
}

// ExtractMathContent strips the $ delimiters and returns the inner math expression.
func ExtractMathContent(text string) string {
	text = strings.TrimSpace(text)
	if !IsMathExpression(text) {
		return text
	}
	// Remove $ from both ends
	inner := text[1 : len(text)-1]
	return strings.TrimSpace(inner)
}

// ParseMath parses a Typst math expression string into an AST.
// The input should be the raw cell text (with or without $ delimiters).
func ParseMath(text string) *Node {
	content := ExtractMathContent(text)
	if content == "" {
		return &Node{Type: NodeLiteral, Value: ""}
	}

	lexer := NewLexer(content)
	tokens := lexer.Tokenize()

	parser := NewParser(tokens)
	return parser.Parse()
}

// RenderMathToText converts a Typst math expression to Unicode text.
// This is the simplest rendering mode — useful for standard font output.
//
// Example:
//
//	RenderMathToText("$ A = pi r^2 $") => "A = πr²"
func RenderMathToText(text string) string {
	ast := ParseMath(text)
	return FlattenToText(ast)
}

// RenderMathToLayout computes the full layout with glyph positions for a math expression.
// Use this for high-quality glyph-based rendering in PDF content streams.
func RenderMathToLayout(text string, ctx *RenderContext) *MathLayout {
	ast := ParseMath(text)
	engine := NewLayoutEngine(ctx)
	return engine.Layout(ast)
}
