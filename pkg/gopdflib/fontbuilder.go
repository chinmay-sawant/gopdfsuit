package gopdflib

// FontBuilder is a chainable overlay over FontOpts. It changes no engine
// behavior: the terminal Props renders the same colon-separated grammar
// FontOpts.String produces, so every chain spells the same string as the
// equivalent MakeProps call.
//
// Zero defaults: Font(name) starts at size 0 (rendered as 12 by
// FontOpts.String), regular style, left alignment, and no borders. An
// empty name renders as Helvetica. These match the engine's parseProps
// fallbacks.
//
// All methods are nil-receiver safe, mirroring the AddTitle nil pattern:
// chaining on a nil *FontBuilder no-ops and stays nil, Props on nil
// returns "", and Cell on nil returns a cell carrying only the text.
type FontBuilder struct {
	opts FontOpts
}

// Font starts a fluent font chain with the given font name.
func Font(name string) *FontBuilder {
	return &FontBuilder{opts: FontOpts{Name: name, Align: AlignLeft}}
}

// Size sets the font size. Values <= 0 render as 12 via FontOpts.String,
// matching the engine default.
func (b *FontBuilder) Size(n int) *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Size = n
	return b
}

// Bold enables bold.
func (b *FontBuilder) Bold() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Bold = true
	return b
}

// Italic enables italic.
func (b *FontBuilder) Italic() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Italic = true
	return b
}

// Underline enables underline.
func (b *FontBuilder) Underline() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Underline = true
	return b
}

// Left aligns text left.
func (b *FontBuilder) Left() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Align = AlignLeft
	return b
}

// Center aligns text center.
func (b *FontBuilder) Center() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Align = AlignCenter
	return b
}

// Right aligns text right.
func (b *FontBuilder) Right() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Align = AlignRight
	return b
}

// Borders sets the L:R:T:B border flags in props order.
func (b *FontBuilder) Borders(l, r, t, bot int) *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Borders = Borders{l, r, t, bot}
	return b
}

// Bordered enables all four borders.
func (b *FontBuilder) Bordered() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Borders = Borders{1, 1, 1, 1}
	return b
}

// Borderless clears all four borders.
func (b *FontBuilder) Borderless() *FontBuilder {
	if b == nil {
		return nil
	}
	b.opts.Borders = Borders{0, 0, 0, 0}
	return b
}

// Props renders the accumulated options as a props string. On a nil
// builder it returns "".
func (b *FontBuilder) Props() string {
	if b == nil {
		return ""
	}
	return b.opts.String()
}

// Cell returns a cell with the given text and the accumulated props. On a
// nil builder it returns a cell carrying only the text with empty props.
func (b *FontBuilder) Cell(text string) Cell {
	if b == nil {
		return Cell{Text: text}
	}
	return NewCell(text, b.opts.String())
}

// CellBuilder is a chainable overlay over Cell. It changes no engine
// behavior: Build emits a plain Cell whose Props matches the equivalent
// MakeProps output.
//
// Zero defaults: Text(s) starts with the FontBuilder zero default
// (Helvetica 12, regular, left, no borders), no background or text color,
// and math rendering disabled. WithFont replaces the font options;
// Props sets an explicit props string that wins over WithFont; Bg/Fg set
// the bgcolor/textcolor fields; Math enables math rendering.
//
// All methods are nil-receiver safe: chaining on a nil *CellBuilder
// no-ops and stays nil, Build and Cell on nil return the zero Cell.
type CellBuilder struct {
	text     string
	opts     FontOpts
	props    string
	hasProps bool
	bg       string
	fg       string
	math     bool
}

// Text starts a fluent cell chain with the given text.
func Text(s string) *CellBuilder {
	return &CellBuilder{text: s, opts: FontOpts{Align: AlignLeft}}
}

// WithFont replaces the builder's font options with a copy of fb's. A nil
// fb or a call on a nil builder no-ops.
func (c *CellBuilder) WithFont(fb *FontBuilder) *CellBuilder {
	if c == nil || fb == nil {
		return c
	}
	c.opts = fb.opts
	c.hasProps = false
	return c
}

// Props sets an explicit props string, overriding any WithFont options.
// Call WithFont after Props to switch back to font-driven props.
func (c *CellBuilder) Props(s string) *CellBuilder {
	if c == nil {
		return nil
	}
	c.props = s
	c.hasProps = true
	return c
}

// Bg sets the cell background color (for example "#F5F5F5").
func (c *CellBuilder) Bg(color string) *CellBuilder {
	if c == nil {
		return nil
	}
	c.bg = color
	return c
}

// Fg sets the cell text color (for example "#B00020").
func (c *CellBuilder) Fg(color string) *CellBuilder {
	if c == nil {
		return nil
	}
	c.fg = color
	return c
}

// Math enables math rendering for the cell.
func (c *CellBuilder) Math() *CellBuilder {
	if c == nil {
		return nil
	}
	c.math = true
	return c
}

// Build assembles the accumulated text, props, colors, and math flag into
// a Cell. On a nil builder it returns the zero Cell.
func (c *CellBuilder) Build() Cell {
	if c == nil {
		return Cell{}
	}
	props := c.props
	if !c.hasProps {
		props = c.opts.String()
	}
	out := NewCell(c.text, props)
	out.BgColor = c.bg
	out.TextColor = c.fg
	if c.math {
		out.MathEnabled = boolPtr(true)
	}
	return out
}

// Cell is an alias for Build. On a nil builder it returns the zero Cell.
func (c *CellBuilder) Cell() Cell {
	return c.Build()
}
