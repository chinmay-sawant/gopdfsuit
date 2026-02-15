package typstsyntax

// TypstSymbols maps Typst symbol names to their Unicode characters.
// This is used by the parser to resolve symbol references in math mode.
var TypstSymbols = map[string]string{
	// Greek lowercase
	"alpha":   "α",
	"beta":    "β",
	"gamma":   "γ",
	"delta":   "δ",
	"epsilon": "ε",
	"zeta":    "ζ",
	"eta":     "η",
	"theta":   "θ",
	"iota":    "ι",
	"kappa":   "κ",
	"lambda":  "λ",
	"mu":      "μ",
	"nu":      "ν",
	"xi":      "ξ",
	"omicron": "ο",
	"pi":      "π",
	"rho":     "ρ",
	"sigma":   "σ",
	"tau":     "τ",
	"upsilon": "υ",
	"phi":     "φ",
	"chi":     "χ",
	"psi":     "ψ",
	"omega":   "ω",

	// Greek uppercase
	"Alpha":   "Α",
	"Beta":    "Β",
	"Gamma":   "Γ",
	"Delta":   "Δ",
	"Epsilon": "Ε",
	"Zeta":    "Ζ",
	"Eta":     "Η",
	"Theta":   "Θ",
	"Iota":    "Ι",
	"Kappa":   "Κ",
	"Lambda":  "Λ",
	"Mu":      "Μ",
	"Nu":      "Ν",
	"Xi":      "Ξ",
	"Omicron": "Ο",
	"Pi":      "Π",
	"Rho":     "Ρ",
	"Sigma":   "Σ",
	"Tau":     "Τ",
	"Upsilon": "Υ",
	"Phi":     "Φ",
	"Chi":     "Χ",
	"Psi":     "Ψ",
	"Omega":   "Ω",

	// Number sets (blackboard bold)
	"NN": "ℕ",
	"ZZ": "ℤ",
	"QQ": "ℚ",
	"RR": "ℝ",
	"CC": "ℂ",

	// Operators and relations
	"times":      "×",
	"div":        "÷",
	"dot":        "·",
	"dot.c":      "⋯",
	"dots":       "…",
	"dots.c":     "⋯",
	"dots.v":     "⋮",
	"dots.down":  "⋱",
	"cdot":       "·",
	"star":       "⋆",
	"ast":        "∗",
	"plus":       "+",
	"minus":      "−",
	"plus.minus": "±",
	"minus.plus": "∓",

	// Relations
	"eq":     "=",
	"ne":     "≠",
	"neq":    "≠",
	"lt":     "<",
	"gt":     ">",
	"le":     "≤",
	"ge":     "≥",
	"leq":    "≤",
	"geq":    "≥",
	"prec":   "≺",
	"succ":   "≻",
	"approx": "≈",
	"sim":    "∼",
	"simeq":  "≃",
	"cong":   "≅",
	"equiv":  "≡",
	"prop":   "∝",
	"propto": "∝",

	// Set theory
	"in":        "∈",
	"in.not":    "∉",
	"notin":     "∉",
	"ni":        "∋",
	"subset":    "⊂",
	"supset":    "⊃",
	"subset.eq": "⊆",
	"supset.eq": "⊇",
	"union":     "∪",
	"sect":      "∩",
	"emptyset":  "∅",

	// Logic
	"and":    "∧",
	"or":     "∨",
	"not":    "¬",
	"forall": "∀",
	"exists": "∃",
	"top":    "⊤",
	"bot":    "⊥",
	"tack":   "⊢",
	"models": "⊨",

	// Arrows
	"arrow.r":         "→",
	"arrow.l":         "←",
	"arrow.t":         "↑",
	"arrow.b":         "↓",
	"arrow.lr":        "↔",
	"arrow.r.double":  "⇒",
	"arrow.l.double":  "⇐",
	"arrow.lr.double": "⇔",
	"implies":         "⇒",
	"iff":             "⇔",
	"mapsto":          "↦",

	// Arrow shorthand
	"->":  "→",
	"<-":  "←",
	"=>":  "⇒",
	"<=>": "⇔",
	"|->": "↦",

	// Comparison shorthand
	"!=": "≠",
	"<=": "≤",
	">=": "≥",
	"<<": "≪",
	">>": "≫",

	// Big operators
	"sum":             "∑",
	"product":         "∏",
	"integral":        "∫",
	"integral.double": "∬",
	"integral.triple": "∭",
	"integral.cont":   "∮",
	"coprod":          "∐",
	"bigoplus":        "⨁",
	"bigotimes":       "⨂",

	// Misc math
	"oo":       "∞",
	"infinity": "∞",
	"partial":  "∂",
	"nabla":    "∇",
	"gradient": "∇",
	"hbar":     "ℏ",
	"ell":      "ℓ",
	"wp":       "℘",
	"Re":       "ℜ",
	"Im":       "ℑ",
	"aleph":    "ℵ",
	"angle":    "∠",
	"perp":     "⊥",
	"parallel": "∥",
	"degree":   "°",

	// Delimiters
	"lbrace": "{",
	"rbrace": "}",
	"lbrack": "[",
	"rbrack": "]",
	"lparen": "(",
	"rparen": ")",
	"langle": "⟨",
	"rangle": "⟩",
	"lfloor": "⌊",
	"rfloor": "⌋",
	"lceil":  "⌈",
	"rceil":  "⌉",
	"vert":   "|",
	"Vert":   "‖",

	// Suit symbols
	"suit.heart":   "♥",
	"suit.diamond": "♦",
	"suit.club":    "♣",
	"suit.spade":   "♠",

	// Checkmarks and crosses
	"checkmark": "✓",
	"crossmark": "✗",
}

// SuperscriptDigits maps ASCII digits and common chars to their Unicode superscript counterparts.
var SuperscriptDigits = map[rune]rune{
	'0': '⁰',
	'1': '¹',
	'2': '²',
	'3': '³',
	'4': '⁴',
	'5': '⁵',
	'6': '⁶',
	'7': '⁷',
	'8': '⁸',
	'9': '⁹',
	'+': '⁺',
	'-': '⁻',
	'=': '⁼',
	'(': '⁽',
	')': '⁾',
	'n': 'ⁿ',
	'i': 'ⁱ',
}

// SubscriptDigits maps ASCII digits and common chars to their Unicode subscript counterparts.
var SubscriptDigits = map[rune]rune{
	'0': '₀',
	'1': '₁',
	'2': '₂',
	'3': '₃',
	'4': '₄',
	'5': '₅',
	'6': '₆',
	'7': '₇',
	'8': '₈',
	'9': '₉',
	'+': '₊',
	'-': '₋',
	'=': '₌',
	'(': '₍',
	')': '₎',
	'a': 'ₐ',
	'e': 'ₑ',
	'o': 'ₒ',
	'x': 'ₓ',
	'i': 'ᵢ',
	'j': 'ⱼ',
	'k': 'ₖ',
	'n': 'ₙ',
	'p': 'ₚ',
	'r': 'ᵣ',
	's': 'ₛ',
	't': 'ₜ',
	'u': 'ᵤ',
	'v': 'ᵥ',
}

// PredefinedOperators are text operators displayed in upright (roman) style.
var PredefinedOperators = map[string]bool{
	"arccos": true, "arcsin": true, "arctan": true, "arg": true,
	"cos": true, "cosh": true, "cot": true, "coth": true,
	"csc": true, "csch": true, "ctg": true, "deg": true,
	"det": true, "dim": true, "exp": true, "gcd": true, "lcm": true,
	"hom": true, "id": true, "im": true, "inf": true,
	"ker": true, "lg": true, "lim": true, "liminf": true,
	"limsup": true, "ln": true, "log": true, "max": true,
	"min": true, "mod": true, "Pr": true, "sec": true,
	"sech": true, "sin": true, "sinc": true, "sinh": true,
	"sup": true, "tan": true, "tanh": true, "tg": true, "tr": true,
}

// AccentMap maps Typst accent function names to combining Unicode characters.
var AccentMap = map[string]string{
	"grave":  "\u0300", // ̀
	"acute":  "\u0301", // ́
	"hat":    "\u0302", // ̂
	"tilde":  "\u0303", // ̃
	"macron": "\u0304", // ̄
	"breve":  "\u0306", // ̆
	"dot":    "\u0307", // ̇
	"ddot":   "\u0308", // ̈
	"dddot":  "\u20DB", // combining three dots above
	"ddddot": "\u20DC", // combining four dots above
	"circle": "\u030A", // ̊
	"arrow":  "\u20D7", // combining right arrow above
	"bar":    "\u0305", // combining overline
}
