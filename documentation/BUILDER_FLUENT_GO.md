# Fluent Go builder

Date: 2026-09-04. Branch: feat/builder-snippets. Commits: 449200c, d3233e8.

This covers the overlay in pkg/gopdflib that lets you build a template in code instead of hand writing JSON. It does not change the engine. The sink is still GeneratePDF.

## Props grammar

One string sets font, size, style, align, borders.

Format: Name:Size:Style:Align:L:R:T:B

Example: Helvetica:18:100:center:1:1:1:1 means Helvetica 18, bold, centered, full borders. Style uses three bits for bold, italic, underline, so 100 is bold only and 000 is plain.

Rules from pkg/gopdflib/props.go:

- Empty name falls back to Helvetica.
- Size <= 0 falls back to 12.
- Unknown align falls back to left.
- Color never lives in props. Use bgcolor and textcolor hex fields on the cell.

Parsed with ParseFontOpts, rendered with FontOpts.String.

## Old vs fluent

Old way: build PDFTemplate literal by hand with raw colon strings.

```go
tpl := gopdflib.PDFTemplate{
  Config: gopdflib.Config{Page: "A4", PageAlignment: 1},
  Title: gopdflib.Title{Props: "Helvetica:18:100:center:0:0:0:0", Text: "Document Title"},
}
pdfBytes, err := gopdflib.GeneratePDF(tpl)
```

Middle layer from 449200c: DocumentBuilder overlay.

```go
b := gopdflib.NewDocument("A4", true)
b.AddTitle("Document Title", gopdflib.WithTitleFont("Helvetica", 18, true))
pdfBytes, err := b.Generate()
```

New fluent spelling from d3233e8, byte identical output:

```go
row := titleTable.AddRow(
  gopdflib.Font("Helvetica").Size(12).Bordered().Cell(""),
  gopdflib.Font("Helvetica").Size(18).Bold().Center().Bordered().Cell("Document Title"),
  gopdflib.Font("Helvetica").Size(12).Right().Bordered().Cell("clause"),
)
gopdflib.AddBracketText(&row[2], "[", "]")
gopdflib.SetCellTextColor(&row[2], "#B00020")
```

Round trip proof from fontbuilder_test.go:

```go
chain := gopdflib.Font("Helvetica").Size(12).Bold().Center().Bordered().Props()
want := gopdflib.MakeProps("Helvetica", 12, true, false, false, "center", [4]int{1,1,1,1})
// chain == want == "Helvetica:12:100:center:1:1:1:1"
```

## DocumentBuilder lifecycle

From pkg/gopdflib/builder.go:

- NewDocument(page string, portrait bool) at builder.go:19. True means PageAlignment 1, false means 2.
- AddTitle(text string, opts ...TitleOption) at builder.go:74.
- AddTable(maxCols int, colWidths ...float64) returns TableBuilder at builder.go:95.
- AddTitleTable(maxCols int, colWidths ...float64) returns TitleTableBuilder at builder.go:134.
- AddSpacer(h float64) at builder.go:155.
- AddImage(name, data string, width, height float64) at builder.go:164. Data is base64.
- Build() PDFTemplate at builder.go:176.
- Generate() ([]byte, error) at builder.go:186.

AddRow returns an alias to the stored row, so SetCell helpers apply in place.

Title options: WithTitleFontOpts(TitleFontOptions{Name, Size, Bold, Italic, Underline}) is current. WithTitleFont(name, size, bold, rest ...bool) stays for compat but reads as deprecated. rest[0] is italic, rest[1] is underline.

## Font chain

From pkg/gopdflib/fontbuilder.go: Font(name).Size(n).Bold().Italic().Underline().Left().Center().Right().Borders(l,r,t,b).Bordered().Borderless(). Terminals are Props() string and Cell(text string) Cell. Nil receiver is safe and Props returns "".

## Cell chain

From fontbuilder.go:146: Text(s).WithFont(fb).Props(s).Bg(color).Fg(color).Math(). Terminals are Build() Cell and Cell() Cell alias. Explicit Props wins over WithFont. Bg sets bgcolor hex, Fg sets textcolor hex.

```go
got := gopdflib.Text("x^2").Bg("#F5F5F5").Fg("#B00020").Math().Build()
```

## Cell helpers

From builder.go:198-355: MakeProps, NewCell, HeaderCell (bold centered full borders), MathCell (MathEnabled true), NewImageCell, SetCellFont, SetCellAlignment, SetCellBorders, SetCellColor, SetCellTextColor, SetCellBgColor, SetRowColor, SetTableColors, AddBracketText (v1 wraps open+text+close, no segments), SetBracketFont.

## Canonical JSON

sampledata/builder-snippets/snippet.json shows what the builder emits:

```json
{"props":"Helvetica:18:100:center:0:0:0:0","text":"Document Title","table":{"maxcolumns":3,"columnwidths":[1,2,1],"rows":[{"row":[{"props":"Helvetica:12:000:right:1:1:1:1","text":"[clause]","textcolor":"#B00020"}]}]}}
```

Full example lives in sampledata/gopdflib/builder-snippets/main.go.

## Tests

- pkg/gopdflib/builder_test.go
- pkg/gopdflib/fontbuilder_test.go
- pkg/gopdflib/props_test.go

Gates: make test, make lint, make test-integration, make test-verify-pdfs.
