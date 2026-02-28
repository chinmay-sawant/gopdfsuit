package typstsyntax

import (
	"strings"
	"unicode"
)

// TokenType represents the type of a Typst math token.
type TokenType int

// TokenType constants.
const (
	TokenText        TokenType = iota // Plain text/identifier
	TokenNumber                       // Numeric literal
	TokenSymbol                       // Known symbol name (pi, alpha, etc.)
	TokenOperator                     // Math operator (+, -, *, =)
	TokenSuperscript                  // ^ (superscript)
	TokenSubscript                    // _ (subscript)
	TokenSlash                        // / (fraction)
	TokenComma                        // , (argument separator)
	TokenSemicolon                    // ; (row separator in matrices)
	TokenLParen                       // (
	TokenRParen                       // )
	TokenLBrace                       // {
	TokenRBrace                       // }
	TokenLBracket                     // [
	TokenRBracket                     // ]
	TokenQuotedText                   // "..." quoted text
	TokenPrime                        // ' (prime)
	TokenAmpersand                    // & (alignment)
	TokenBackslash                    // \\ (line break)
	TokenHash                         // # (code escape)
	TokenDot                          // . (namespace separator)
	TokenWhitespace                   // Space(s)
	TokenEOF                          // End of input
)

// Token represents a single lexical token from Typst math input.
type Token struct {
	Type  TokenType
	Value string
	Pos   int // Position in original string
}

// Lexer tokenizes Typst math input.
type Lexer struct {
	input  []rune
	pos    int
	tokens []Token
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
	}
}

// Tokenize processes the entire input and returns all tokens.
func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		l.skipWhitespaceAndEmit()
		if l.pos >= len(l.input) {
			break
		}

		ch := l.input[l.pos]

		switch {
		case ch == '^':
			l.emit(TokenSuperscript, "^")
			l.pos++
		case ch == '_':
			l.emit(TokenSubscript, "_")
			l.pos++
		case ch == '/':
			l.emit(TokenSlash, "/")
			l.pos++
		case ch == ',':
			l.emit(TokenComma, ",")
			l.pos++
		case ch == ';':
			l.emit(TokenSemicolon, ";")
			l.pos++
		case ch == '(':
			l.emit(TokenLParen, "(")
			l.pos++
		case ch == ')':
			l.emit(TokenRParen, ")")
			l.pos++
		case ch == '{':
			l.emit(TokenLBrace, "{")
			l.pos++
		case ch == '}':
			l.emit(TokenRBrace, "}")
			l.pos++
		case ch == '[':
			l.emit(TokenLBracket, "[")
			l.pos++
		case ch == ']':
			l.emit(TokenRBracket, "]")
			l.pos++
		case ch == '\'':
			l.lexPrimes()
		case ch == '&':
			l.emit(TokenAmpersand, "&")
			l.pos++
		case ch == '#':
			l.emit(TokenHash, "#")
			l.pos++
		case ch == '"':
			l.lexQuotedText()
		case ch == '\\':
			l.lexBackslash()
		case ch == '+' || ch == '-' || ch == '*' || ch == '=':
			l.lexOperatorSequence()
		case ch == '<' || ch == '>' || ch == '!':
			l.lexMultiCharOperator()
		case ch == '|':
			l.lexPipe()
		case unicode.IsDigit(ch):
			l.lexNumber()
		case unicode.IsLetter(ch):
			l.lexIdentifier()
		case ch == '.':
			l.emit(TokenDot, ".")
			l.pos++
		default:
			// Emit unknown characters as text
			l.emit(TokenText, string(ch))
			l.pos++
		}
	}

	l.emit(TokenEOF, "")
	return l.tokens
}

func (l *Lexer) emit(typ TokenType, value string) {
	l.tokens = append(l.tokens, Token{Type: typ, Value: value, Pos: l.pos})
}

func (l *Lexer) peek() rune {
	if l.pos+1 < len(l.input) {
		return l.input[l.pos+1]
	}
	return 0
}

func (l *Lexer) skipWhitespaceAndEmit() {
	start := l.pos
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t') {
		l.pos++
	}
	if l.pos > start {
		l.emit(TokenWhitespace, string(l.input[start:l.pos]))
	}
}

func (l *Lexer) lexNumber() {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		l.pos++
	}
	l.emit(TokenNumber, string(l.input[start:l.pos]))
}

func (l *Lexer) lexIdentifier() {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos])) {
		l.pos++
	}
	word := string(l.input[start:l.pos])

	// Check for dot-qualified symbol names (e.g., "dots.c", "arrow.r")
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		// Peek ahead to see if it's a dot-qualified identifier
		savePos := l.pos
		l.pos++ // skip '.'
		if l.pos < len(l.input) && unicode.IsLetter(l.input[l.pos]) {
			dotStart := l.pos
			for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos])) {
				l.pos++
			}
			qualified := word + "." + string(l.input[dotStart:l.pos])
			// Check if the qualified name is a known symbol
			if _, ok := TypstSymbols[qualified]; ok {
				l.emit(TokenSymbol, qualified)
				return
			}
			// Check for further qualification (e.g., "arrow.r.double")
			if l.pos < len(l.input) && l.input[l.pos] == '.' {
				savePos2 := l.pos
				l.pos++
				if l.pos < len(l.input) && unicode.IsLetter(l.input[l.pos]) {
					dotStart2 := l.pos
					for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos])) {
						l.pos++
					}
					tripleQualified := qualified + "." + string(l.input[dotStart2:l.pos])
					if _, ok := TypstSymbols[tripleQualified]; ok {
						l.emit(TokenSymbol, tripleQualified)
						return
					}
					l.pos = savePos2
				} else {
					l.pos = savePos2
				}
			}
			// If qualified name is not a known symbol, emit parts separately
			if _, ok := TypstSymbols[qualified]; ok {
				l.emit(TokenSymbol, qualified)
				return
			}
			// Not a known qualified symbol, reset
			l.pos = savePos
		} else {
			l.pos = savePos
		}
	}

	// Check if it's a known symbol
	if _, ok := TypstSymbols[word]; ok {
		l.emit(TokenSymbol, word)
	} else {
		l.emit(TokenText, word)
	}
}

func (l *Lexer) lexQuotedText() {
	l.pos++ // skip opening "
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		l.pos++
	}
	text := string(l.input[start:l.pos])
	if l.pos < len(l.input) {
		l.pos++ // skip closing "
	}
	l.emit(TokenQuotedText, text)
}

func (l *Lexer) lexPrimes() {
	count := 0
	for l.pos < len(l.input) && l.input[l.pos] == '\'' {
		count++
		l.pos++
	}
	l.emit(TokenPrime, strings.Repeat("′", count))
}

func (l *Lexer) lexBackslash() {
	switch l.peek() {
	case '\\':
		l.emit(TokenBackslash, "\\\\")
		l.pos += 2
	case '{':
		l.emit(TokenText, "{")
		l.pos += 2
	case '}':
		l.emit(TokenText, "}")
		l.pos += 2
	default:
		l.emit(TokenText, "\\")
		l.pos++
	}
}

func (l *Lexer) lexOperatorSequence() {
	ch := l.input[l.pos]
	next := l.peek()
	// Check for multi-char operators
	if ch == '-' && next == '>' {
		l.emit(TokenSymbol, "->")
		l.pos += 2
		return
	}
	if ch == '=' && next == '>' {
		l.emit(TokenSymbol, "=>")
		l.pos += 2
		return
	}
	if ch == ':' && next == '=' {
		l.emit(TokenOperator, ":=")
		l.pos += 2
		return
	}
	l.emit(TokenOperator, string(ch))
	l.pos++
}

func (l *Lexer) lexMultiCharOperator() {
	ch := l.input[l.pos]
	next := l.peek()

	switch ch {
	case '!':
		switch next {
		case '=':
			l.emit(TokenSymbol, "!=")
			l.pos += 2
		default:
			l.emit(TokenOperator, "!")
			l.pos++
		}
	case '<':
		switch next {
		case '=':
			if l.pos+2 < len(l.input) && l.input[l.pos+2] == '>' {
				l.emit(TokenSymbol, "<=>")
				l.pos += 3
			} else {
				l.emit(TokenSymbol, "<=")
				l.pos += 2
			}
		case '-':
			l.emit(TokenSymbol, "<-")
			l.pos += 2
		default:
			l.emit(TokenOperator, "<")
			l.pos++
		}
	case '>':
		switch next {
		case '=':
			l.emit(TokenSymbol, ">=")
			l.pos += 2
		default:
			l.emit(TokenOperator, ">")
			l.pos++
		}
	default:
		l.emit(TokenOperator, string(ch))
		l.pos++
	}
}

func (l *Lexer) lexPipe() {
	if l.peek() == '-' && l.pos+2 < len(l.input) && l.input[l.pos+2] == '>' {
		l.emit(TokenSymbol, "|->")
		l.pos += 3
	} else {
		l.emit(TokenOperator, "|")
		l.pos++
	}
}
