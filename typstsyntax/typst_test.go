package typstsyntax

import (
	"bytes"
	"testing"
)

func TestIsMathExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"$ A = pi r^2 $", true},
		{"$ x + y $", true},
		{"$ $", false},   // too short
		{"hello", false}, // no delimiters
		{"$x$", true},    // minimal valid
		{"", false},
		{"  $ x $  ", true}, // with surrounding whitespace
	}

	for _, tt := range tests {
		got := IsMathExpression(tt.input)
		if got != tt.expected {
			t.Errorf("IsMathExpression(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestExtractMathContent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$ A = pi r^2 $", "A = pi r^2"},
		{"$x$", "x"},
		{"  $ hello $  ", "hello"},
		{"no dollars", "no dollars"}, // no delimiters, returned as-is
	}

	for _, tt := range tests {
		got := ExtractMathContent(tt.input)
		if got != tt.expected {
			t.Errorf("ExtractMathContent(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestLexerBasic(t *testing.T) {
	lexer := NewLexer("A = pi r^2")
	tokens := lexer.Tokenize()

	// Should produce: Text("A"), WS, Operator("="), WS, Symbol("pi"), WS, Text("r"), Superscript, Number("2"), EOF
	if len(tokens) < 5 {
		t.Fatalf("Expected at least 5 tokens, got %d", len(tokens))
	}

	// First token should be text "A"
	if tokens[0].Type != TokenText || tokens[0].Value != "A" {
		t.Errorf("Token 0: got %v %q, want Text/A", tokens[0].Type, tokens[0].Value)
	}
}

func TestLexerSymbols(t *testing.T) {
	tests := []struct {
		input  string
		expect string // expected symbol value
	}{
		{"pi", "pi"},
		{"alpha", "alpha"},
		{"NN", "NN"},
		{"dots.c", "dots.c"},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens := lexer.Tokenize()
		found := false
		for _, tok := range tokens {
			if tok.Type == TokenSymbol && tok.Value == tt.expect {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Lexer(%q): expected to find symbol %q", tt.input, tt.expect)
		}
	}
}

func TestLexerMultiCharOperators(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"!=", "!="},
		{"<=", "<="},
		{">=", ">="},
		{"->", "->"},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens := lexer.Tokenize()
		found := false
		for _, tok := range tokens {
			if tok.Type == TokenSymbol && tok.Value == tt.expect {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Lexer(%q): expected to find symbol %q, tokens: %+v", tt.input, tt.expect, tokens)
		}
	}
}

func TestParserBasic(t *testing.T) {
	ast := ParseMath("$ A = pi r^2 $")
	if ast == nil {
		t.Fatal("ParseMath returned nil")
	}
}

func TestParserFraction(t *testing.T) {
	ast := ParseMath("$ frac(a, b) $")
	if ast == nil {
		t.Fatal("ParseMath returned nil")
	}

	// The top-level should contain a fraction node somewhere
	found := findNodeType(ast, NodeFraction)
	if !found {
		t.Error("Expected to find NodeFraction in AST")
	}
}

func TestParserSqrt(t *testing.T) {
	ast := ParseMath("$ sqrt(x) $")
	if ast == nil {
		t.Fatal("ParseMath returned nil")
	}
	found := findNodeType(ast, NodeSqrt)
	if !found {
		t.Error("Expected to find NodeSqrt in AST")
	}
}

func TestParserSuperscript(t *testing.T) {
	ast := ParseMath("$ x^2 $")
	if ast == nil {
		t.Fatal("ParseMath returned nil")
	}
	found := findNodeType(ast, NodeSuperscript)
	if !found {
		t.Error("Expected to find NodeSuperscript in AST")
	}
}

func TestParserSubscript(t *testing.T) {
	ast := ParseMath("$ x_i $")
	if ast == nil {
		t.Fatal("ParseMath returned nil")
	}
	found := findNodeType(ast, NodeSubscript)
	if !found {
		t.Error("Expected to find NodeSubscript in AST")
	}
}

func TestRenderMathToText(t *testing.T) {
	tests := []struct {
		input    string
		contains string // substring that should be in the output
	}{
		{"$ pi $", "π"},
		{"$ alpha $", "α"},
		{"$ NN $", "ℕ"},
		{"$ x + y $", "+"},
		{"$ A = pi r^2 $", "π"},
	}

	for _, tt := range tests {
		got := RenderMathToText(tt.input)
		if got == "" {
			t.Errorf("RenderMathToText(%q) returned empty string", tt.input)
			continue
		}
		if !containsStr(got, tt.contains) {
			t.Errorf("RenderMathToText(%q) = %q, expected to contain %q", tt.input, got, tt.contains)
		}
	}
}

func TestRenderMathToLayout(t *testing.T) {
	ctx := &RenderContext{
		X:        100,
		Y:        500,
		FontSize: 12,
		FontRef:  "/F1",
	}

	layout := RenderMathToLayout("$ A = pi r^2 $", ctx)
	if layout == nil {
		t.Fatal("RenderMathToLayout returned nil")
		return //nolint:govet // unreachable but needed for staticcheck SA5011
	}
	if layout.Width <= 0 {
		t.Error("Expected positive layout width")
	}
	if len(layout.Elements) == 0 {
		t.Error("Expected layout to contain elements")
	}
}

func TestRenderToContentStream(t *testing.T) {
	ctx := &RenderContext{
		X:         100,
		Y:         500,
		FontSize:  12,
		FontRef:   "/F1",
		TextColor: "0 0 0",
	}

	layout := RenderMathToLayout("$ A = pi r^2 $", ctx)

	var buf bytes.Buffer
	RenderToContentStream(&buf, layout, ctx)

	output := buf.String()
	if output == "" {
		t.Error("Expected non-empty content stream output")
	}
	// Should contain PDF text operations
	if !containsStr(output, "BT") || !containsStr(output, "ET") {
		t.Error("Expected BT and ET in content stream output")
	}
	if !containsStr(output, "Tf") {
		t.Error("Expected Tf (font) in content stream output")
	}
}

func TestRenderFractionToContentStream(t *testing.T) {
	ctx := &RenderContext{
		X:         100,
		Y:         500,
		FontSize:  12,
		FontRef:   "/F1",
		TextColor: "0 0 0",
	}

	layout := RenderMathToLayout("$ frac(a, b) $", ctx)
	var buf bytes.Buffer
	RenderToContentStream(&buf, layout, ctx)

	output := buf.String()
	// Fraction should produce a line (fraction bar)
	// Look for line drawing commands (m ... l S)
	if !containsStr(output, " m ") || !containsStr(output, " l S") {
		t.Error("Expected fraction bar line in content stream")
	}
}

func TestComplexExpressions(t *testing.T) {
	expressions := []string{
		"$ sum_(i=1)^n i $",
		"$ frac(a + b, c - d) $",
		"$ sqrt(x^2 + y^2) $",
		"$ vec(a, b, c) $",
		"$ binom(n, k) $",
		"$ hat(x) $",
		"$ abs(x) $",
	}

	for _, expr := range expressions {
		ast := ParseMath(expr)
		if ast == nil {
			t.Errorf("ParseMath(%q) returned nil", expr)
			continue
		}
		text := FlattenToText(ast)
		if text == "" {
			t.Errorf("FlattenToText for %q returned empty", expr)
		}
	}
}

// Helper: recursively search for a node type in the AST.
func findNodeType(node *Node, typ NodeType) bool {
	if node == nil {
		return false
	}
	if node.Type == typ {
		return true
	}
	for _, child := range node.Children {
		if findNodeType(child, typ) {
			return true
		}
	}
	for _, arg := range node.Args {
		if findNodeType(arg, typ) {
			return true
		}
	}
	return false
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && bytes.Contains([]byte(s), []byte(substr))
}
