package gopdflib

import (
	"fmt"
	"strings"
)

// DocumentBuilder is a pure overlay over PDFTemplate: it emits the current
// Props grammar and bgcolor/textcolor strings and sinks through GeneratePDF.
// It changes no engine draw behavior.
type DocumentBuilder struct {
	config   Config
	title    Title
	elements []Element
}

// NewDocument starts a builder for the given page size ("A4", "Letter", ...).
// portrait true selects PageAlignment 1, false selects 2 (landscape).
func NewDocument(page string, portrait bool) *DocumentBuilder {
	alignment := 2
	if portrait {
		alignment = 1
	}
	return &DocumentBuilder{
		config: Config{Page: page, PageAlignment: alignment},
		title: Title{
			Props: MakeProps("Helvetica", 18, true, false, false, "center", [4]int{0, 0, 0, 0}),
		},
	}
}

// TitleOption customizes the title set by AddTitle.
type TitleOption func(*Title)

// TitleFontOptions customizes the title font set by AddTitle without
// positional bools.
type TitleFontOptions struct {
	Name      string
	Size      int
	Bold      bool
	Italic    bool
	Underline bool
}

// WithTitleFontOpts overrides the title font from options. The title stays
// centered with no borders.
func WithTitleFontOpts(o TitleFontOptions) TitleOption {
	return func(t *Title) {
		if t == nil {
			return
		}
		t.Props = MakeProps(o.Name, o.Size, o.Bold, o.Italic, o.Underline, "center", [4]int{0, 0, 0, 0})
	}
}

// WithTitleFont overrides the title font. The optional extra bools are
// italic and underline in that order; the title stays centered with no
// borders.
//
// Deprecated: use WithTitleFontOpts with TitleFontOptions instead. Kept
// because builder-snippets samples and Editor snippets emit this name.
func WithTitleFont(name string, size int, bold bool, rest ...bool) TitleOption {
	italic, underline := false, false
	if len(rest) > 0 {
		italic = rest[0]
	}
	if len(rest) > 1 {
		underline = rest[1]
	}
	return WithTitleFontOpts(TitleFontOptions{Name: name, Size: size, Bold: bold, Italic: italic, Underline: underline})
}

// AddTitle sets the document title text and applies any options.
func (b *DocumentBuilder) AddTitle(text string, opts ...TitleOption) {
	if b == nil {
		return
	}
	b.title.Text = text
	for _, opt := range opts {
		if opt != nil {
			opt(&b.title)
		}
	}
}

// TableBuilder appends rows to one table element of a DocumentBuilder.
// The table is stored by pointer, so rows added later are visible to Build
// and Generate without re-attaching anything.
type TableBuilder struct {
	table *Table
}

// AddTable appends a table element with the given max columns and optional
// relative column widths, and returns a builder for its rows.
func (b *DocumentBuilder) AddTable(maxCols int, colWidths ...float64) *TableBuilder {
	tbl := &Table{MaxColumns: maxCols}
	if len(colWidths) > 0 {
		tbl.ColumnWidths = append([]float64(nil), colWidths...)
	}
	if b != nil {
		b.elements = append(b.elements, Element{Type: "table", Table: tbl})
	}
	return &TableBuilder{table: tbl}
}

// appendRowCells is the single row-append path behind TableBuilder.AddRow
// and TitleTableBuilder.AddRow. It appends the cells as one row and returns
// the stored row, which aliases the stored cells so SetCell* calls on its
// elements apply to the table.
func appendRowCells(rows *[]Row, cells []Cell) []Cell {
	row := make([]Cell, len(cells))
	copy(row, cells)
	*rows = append(*rows, Row{Row: row})
	return (*rows)[len(*rows)-1].Row
}

// AddRow appends the cells as one row and returns the stored row. The
// returned slice aliases the stored row, so SetCell* calls on its elements
// apply to the table.
func (t *TableBuilder) AddRow(cells ...Cell) []Cell {
	if t == nil || t.table == nil {
		return nil
	}
	return appendRowCells(&t.table.Rows, cells)
}

// TitleTableBuilder appends rows to the document title's embedded table.
type TitleTableBuilder struct {
	table *TitleTable
}

// AddTitleTable sets the title's embedded table and returns a builder for
// its rows.
func (b *DocumentBuilder) AddTitleTable(maxCols int, colWidths ...float64) *TitleTableBuilder {
	tbl := &TitleTable{MaxColumns: maxCols}
	if len(colWidths) > 0 {
		tbl.ColumnWidths = append([]float64(nil), colWidths...)
	}
	if b != nil {
		b.title.Table = tbl
	}
	return &TitleTableBuilder{table: tbl}
}

// AddRow appends the cells as one title-table row and returns the stored
// row, aliasing it the same way TableBuilder.AddRow does.
func (t *TitleTableBuilder) AddRow(cells ...Cell) []Cell {
	if t == nil || t.table == nil {
		return nil
	}
	return appendRowCells(&t.table.Rows, cells)
}

// AddSpacer appends a vertical gap element of height h.
func (b *DocumentBuilder) AddSpacer(h float64) {
	if b == nil {
		return
	}
	b.elements = append(b.elements, Element{Type: "spacer", Spacer: &Spacer{Height: h}})
}

// AddImage appends an image element. data is base64 image bytes; width and
// height are the requested display size in points.
func (b *DocumentBuilder) AddImage(name, data string, width, height float64) {
	if b == nil {
		return
	}
	b.elements = append(b.elements, Element{
		Type:  "image",
		Image: &Image{ImageName: name, ImageData: data, Width: width, Height: height},
	})
}

// Build assembles the accumulated config, title, and elements into a
// PDFTemplate ready for GeneratePDF.
func (b *DocumentBuilder) Build() PDFTemplate {
	if b == nil {
		return PDFTemplate{}
	}
	out := PDFTemplate{Config: b.config, Title: b.title}
	out.Elements = append([]Element(nil), b.elements...)
	return out
}

// Generate builds the template and renders it to PDF bytes.
func (b *DocumentBuilder) Generate() ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("%w: gopdflib: DocumentBuilder.Generate on nil builder", ErrInvalidInput)
	}
	return GeneratePDF(b.Build())
}

// MakeProps renders a props string from typed parts. It is the function
// form of FontOpts.String for snippet call sites: the single canonical
// props grammar (fallbacks: empty name to Helvetica, size <= 0 to 12,
// non-3-char style to regular, unknown align to left) shared with the
// engine's parseProps.
func MakeProps(name string, size int, bold, italic, underline bool, align string, borders [4]int) string {
	return FontOpts{
		Name:      name,
		Size:      size,
		Bold:      bold,
		Italic:    italic,
		Underline: underline,
		Align:     Align(align),
		Borders:   Borders(borders),
	}.String()
}

// NewCell returns a cell with the given text and props string. It is the
// single cell path: every spelling (FontOpts struct, Cell literal, option
// functions like SetCellFont, FontBuilder, CellBuilder) renders through
// FontOpts.String, so all of them produce the same props grammar.
func NewCell(text, props string) Cell {
	return Cell{Props: props, Text: text}
}

// HeaderCell returns a bold centered cell with full borders.
func HeaderCell(text string) Cell {
	return Cell{
		Props: MakeProps("Helvetica", 12, true, false, false, "center", [4]int{1, 1, 1, 1}),
		Text:  text,
	}
}

// MathCell returns a cell with math rendering enabled.
func MathCell(text string) Cell {
	c := NewCell(text, MakeProps("Helvetica", 12, false, false, false, "left", [4]int{1, 1, 1, 1}))
	c.MathEnabled = boolPtr(true)
	return c
}

func boolPtr(v bool) *bool {
	return &v
}

// NewImageCell returns a cell with an embedded image. data is base64 image
// bytes; width and height are the requested display size in points.
func NewImageCell(name, data string, width, height float64, props string) Cell {
	return Cell{
		Props: props,
		Image: &Image{ImageName: name, ImageData: data, Width: width, Height: height},
	}
}

// SetCellFont rewrites the cell's font name, size, and style flags while
// preserving its alignment and borders. Empty name and size <= 0 keep the
// current values. It round-trips through the canonical FontOpts grammar
// (ParseFontOpts/String), never a bespoke parser.
func SetCellFont(c *Cell, name string, size int, bold, italic, underline bool) {
	if c == nil {
		return
	}
	opts := ParseFontOpts(c.Props)
	if strings.TrimSpace(name) != "" {
		opts.Name = strings.TrimSpace(name)
	}
	if size > 0 {
		opts.Size = size
	}
	opts.Bold = bold
	opts.Italic = italic
	opts.Underline = underline
	c.Props = opts.String()
}

// SetCellAlignment rewrites the cell's alignment, keeping font and borders.
func SetCellAlignment(c *Cell, align string) {
	if c == nil {
		return
	}
	opts := ParseFontOpts(c.Props)
	opts.Align = normalizeAlign(align)
	c.Props = opts.String()
}

// SetCellBorders rewrites the cell's L:R:T:B borders, keeping font and
// alignment.
func SetCellBorders(c *Cell, borders [4]int) {
	if c == nil {
		return
	}
	opts := ParseFontOpts(c.Props)
	opts.Borders = Borders(borders)
	c.Props = opts.String()
}

// SetCellColor sets both the background and text color of a cell. Either
// may be "" to clear that color.
func SetCellColor(c *Cell, bgColor, textColor string) {
	if c == nil {
		return
	}
	c.BgColor = bgColor
	c.TextColor = textColor
}

// SetCellTextColor sets the text color of a cell (for example "#B00020").
func SetCellTextColor(c *Cell, color string) {
	if c == nil {
		return
	}
	c.TextColor = color
}

// SetCellBgColor sets the background color of a cell.
func SetCellBgColor(c *Cell, color string) {
	if c == nil {
		return
	}
	c.BgColor = color
}

// SetRowColor applies the background and text color to every cell in row.
func SetRowColor(row []Cell, bgColor, textColor string) {
	for i := range row {
		SetCellColor(&row[i], bgColor, textColor)
	}
}

// SetTableColors applies the background and text color to every cell of
// every row in the table built by t.
func SetTableColors(t *TableBuilder, bgColor, textColor string) {
	if t == nil || t.table == nil {
		return
	}
	for ri := range t.table.Rows {
		SetRowColor(t.table.Rows[ri].Row, bgColor, textColor)
	}
}

// AddBracketText wraps the cell text as open + text + closeDelim (v1 bracket
// support without rich-text segments).
func AddBracketText(c *Cell, open, closeDelim string) {
	if c == nil {
		return
	}
	c.Text = open + c.Text + closeDelim
}

// SetBracketFont overrides the cell's font name and size, preserving style,
// alignment, and borders. Empty font and size <= 0 keep current values.
// Like SetCellFont it round-trips through the canonical FontOpts grammar.
func SetBracketFont(c *Cell, font string, size int) {
	if c == nil {
		return
	}
	opts := ParseFontOpts(c.Props)
	if strings.TrimSpace(font) != "" {
		opts.Name = strings.TrimSpace(font)
	}
	if size > 0 {
		opts.Size = size
	}
	c.Props = opts.String()
}
