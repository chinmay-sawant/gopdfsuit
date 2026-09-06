# gopdflib API

Go library for template-based PDF generation, merge, split, compress, form fill, redact, and HTML conversion.

## Install

```sh
go get github.com/chinmay-sawant/gopdfsuit/v7@v7.0.0
```

## Import

```go
import "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
```

## Generate

`GeneratePDF(template PDFTemplate) ([]byte, error)` renders a template to PDF bytes. `GeneratePDFBorrowed` avoids the final copy; call `Release` exactly once.

```go
tmpl := gopdflib.PDFTemplate{
    Config: gopdflib.Config{Page: "A4", PageAlignment: 1},
    Title:  gopdflib.Title{Props: "Helvetica:18:100:center:0:0:0:0", Text: "Hello"},
}
pdfBytes, err := gopdflib.GeneratePDF(tmpl)
if err != nil { log.Fatal(err) }
os.WriteFile("out.pdf", pdfBytes, 0644)
```

Borrowed-buffer variant for hot paths:

```go
doc, err := gopdflib.GeneratePDFBorrowed(tmpl)
if err != nil { log.Fatal(err) }
defer doc.Release()
use(doc.Bytes())
```

JSON entry points: `DecodeTemplateJSON(data []byte) (PDFTemplate, error)`, `GeneratePDFFromJSON(data []byte) ([]byte, error)`, `GeneratePDFBorrowedFromJSON(data []byte) (*BorrowedPDF, error)`.

```go
raw, _ := os.ReadFile("template.json")
tmpl, err := gopdflib.DecodeTemplateJSON(raw)
if err != nil { log.Fatal(err) }
pdfBytes, err := gopdflib.GeneratePDF(tmpl)
_ = pdfBytes
```

Fonts: `GetAvailableFonts() []FontInfo`, `GetFontRegistry() *pdf.CustomFontRegistry`, `WarmRuntimePools()`.

```go
gopdflib.WarmRuntimePools()
for _, f := range gopdflib.GetAvailableFonts() {
    fmt.Println(f.Name, f.Reference)
}
```

## Merge

`MergePDFs(files [][]byte) ([]byte, error)` concatenates PDFs in order. Caps: `MaxMergeInputBytes`, `MaxMergeTotalInputBytes`, `MaxMergeFileCount`.

```go
a, _ := os.ReadFile("a.pdf")
b, _ := os.ReadFile("b.pdf")
merged, err := gopdflib.MergePDFs([][]byte{a, b})
if err != nil { log.Fatal(err) }
os.WriteFile("merged.pdf", merged, 0644)
```

## Split

`SplitPDF(file []byte, spec SplitSpec) ([][]byte, error)` splits by pages, ranges, or chunk size.

```go
src, _ := os.ReadFile("doc.pdf")
parts, err := gopdflib.SplitPDF(src, gopdflib.SplitSpec{Pages: []int{1, 3}})
if err != nil { log.Fatal(err) }
os.WriteFile("part1.pdf", parts[0], 0644)
```

Chunk every N pages:

```go
src, _ := os.ReadFile("doc.pdf")
parts, err := gopdflib.SplitPDF(src, gopdflib.SplitSpec{MaxPerFile: 5})
if err != nil { log.Fatal(err) }
fmt.Println(len(parts))
```

Helpers: `ParsePageSpec(spec string, totalPages int) ([]int, error)`, `ParseSplitSpecJSON(data []byte) (SplitSpec, error)`, `SplitPDFWithSpecJSON(file, specJSON []byte) ([][]byte, error)`.

```go
pages, err := gopdflib.ParsePageSpec("1-3,5", 10)
if err != nil { log.Fatal(err) }
parts, err := gopdflib.SplitPDF(src, gopdflib.SplitSpec{Pages: pages})
_ = parts
```

## Compress

`CompressPDF(data []byte, opts CompressOptions) ([]byte, error)`. Tiers: `CompressLight`, `CompressMedium`, `CompressHeavy`. Cap: `MaxCompressInputBytes`.

```go
src, _ := os.ReadFile("doc.pdf")
out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{Level: gopdflib.CompressMedium})
if err != nil { log.Fatal(err) }
os.WriteFile("small.pdf", out, 0644)
```

Level helpers: `ParseCompressLevel(s string) CompressLevel`, `ToServerLevel(level any) (CompressLevel, error)`, `ToWasmLevel(level any) (int, error)`.

```go
lvl := gopdflib.ParseCompressLevel("heavy")
out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{Level: lvl})
if err != nil { log.Fatal(err) }
_ = out
```

## Fill

`FillPDFWithXFDF(pdfBytes, xfdfBytes []byte) ([]byte, error)` fills AcroForm fields from XFDF.

```go
pdfBytes, _ := os.ReadFile("form.pdf")
xfdfBytes, _ := os.ReadFile("data.xfdf")
filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
if err != nil { log.Fatal(err) }
os.WriteFile("filled.pdf", filled, 0644)
```

## Redact

Search text, inspect positions, apply rectangles. Types: `RedactionRect`, `ApplyRedactionOptions`, `RedactionApplyReport`, `PageCapability`.

```go
src, _ := os.ReadFile("doc.pdf")
rects, err := gopdflib.FindTextOccurrences(src, "secret")
if err != nil { log.Fatal(err) }
out, err := gopdflib.ApplyRedactions(src, rects)
_ = out
```

Advanced path with report:

```go
src, _ := os.ReadFile("doc.pdf")
opts := gopdflib.ApplyRedactionOptions{Mode: "secure", Password: "pw"}
out, report, err := gopdflib.ApplyRedactionsAdvancedWithReport(src, opts)
if err != nil { log.Fatal(err) }
fmt.Println(report.AppliedSecure, report.AppliedRectangles)
```

Planning helpers: `GetPageInfo(pdfBytes []byte) (PageInfo, error)`, `ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error)`, `AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error)`, `ApplyRedactionsAdvanced(pdfBytes []byte, options ApplyRedactionOptions) ([]byte, error)`.

```go
src, _ := os.ReadFile("doc.pdf")
caps, err := gopdflib.AnalyzePageCapabilities(src)
if err != nil { log.Fatal(err) }
fmt.Println(caps[0].HasText, caps[0].HasImage)
```

## HTML

`ConvertHTMLToPDF(req HTMLToPDFRequest) ([]byte, error)` and `ConvertHTMLToImage(req HTMLToImageRequest) ([]byte, error)`. Pure-Go, no browser. Image formats: `png`, `jpg` (`svg` is rejected).

```go
pdfBytes, err := gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{
    HTML: "<h1>Hello</h1>", PageSize: "A4", Orientation: "Portrait",
})
if err != nil { log.Fatal(err) }
os.WriteFile("page.pdf", pdfBytes, 0644)
```

```go
img, err := gopdflib.ConvertHTMLToImage(gopdflib.HTMLToImageRequest{
    HTML: "<h1>Hello</h1>", Format: "png", Width: 800, Height: 600,
})
if err != nil { log.Fatal(err) }
os.WriteFile("page.png", img, 0644)
```

## Builder

Fluent overlay over `PDFTemplate`. Entry: `NewDocument(page string, portrait bool) *DocumentBuilder`. Sink: `Build() PDFTemplate` or `Generate() ([]byte, error)`.

```go
b := gopdflib.NewDocument("A4", true)
b.AddTitle("Report")
tb := b.AddTable(2, 1, 1)
tb.AddRow(gopdflib.HeaderCell("A"), gopdflib.HeaderCell("B"))
pdfBytes, err := b.Generate()
if err != nil { log.Fatal(err) }
_ = pdfBytes
```

Title options: `WithTitleFontOpts(o TitleFontOptions) TitleOption`, `WithTitleFont(name string, size int, bold bool, rest ...bool) TitleOption`.

```go
b := gopdflib.NewDocument("A4", true)
b.AddTitle("Title", gopdflib.WithTitleFontOpts(gopdflib.TitleFontOptions{
    Name: "Helvetica", Size: 18, Bold: true,
}))
pdfBytes, err := b.Generate()
_ = pdfBytes
```

Elements: `(b *DocumentBuilder) AddTable(maxCols int, colWidths ...float64) *TableBuilder`, `AddTitleTable`, `AddSpacer(h float64)`, `AddImage(name, data string, width, height float64)`, `(t *TableBuilder) AddRow(cells ...Cell) []Cell`.

```go
b := gopdflib.NewDocument("A4", true)
b.AddSpacer(12)
b.AddImage("logo", base64PNG, 100, 40)
tb := b.AddTable(2, 2, 1)
row := tb.AddRow(gopdflib.NewCell("a", "Helvetica:12:000:left:1:1:1:1"))
_ = row
```

Cells: `NewCell(text, props string) Cell`, `HeaderCell(text string) Cell`, `MathCell(text string) Cell`, `NewImageCell(name, data string, width, height float64, props string) Cell`, `Font(name string) *FontBuilder`, `Text(s string) *CellBuilder`.

```go
c := gopdflib.Font("Helvetica").Size(12).Bold().Center().Bordered().Cell("Hi")
gopdflib.SetCellTextColor(&c, "#B00020")
tb := gopdflib.NewDocument("A4", true).AddTable(1)
tb.AddRow(c)
```

Mutators: `SetCellFont`, `SetCellAlignment`, `SetCellBorders`, `SetCellColor`, `SetCellBgColor`, `SetCellTextColor`, `SetRowColor`, `SetTableColors`, `AddBracketText`, `SetBracketFont`, `MakeProps`.

```go
c := gopdflib.NewCell("x", gopdflib.MakeProps("Helvetica", 12, false, false, false, "left", [4]int{1, 1, 1, 1}))
gopdflib.AddBracketText(&c, "[", "]")
gopdflib.SetBracketFont(&c, "Helvetica", 12)
fmt.Println(c.Text, c.Props)
```

## Prepared

`PrepareTemplate(template PDFTemplate) (*PreparedTemplate, error)` translates once for repeated renders. Safe for concurrent use after construction.

```go
prep, err := gopdflib.PrepareTemplate(tmpl)
if err != nil { log.Fatal(err) }
first, err := prep.GeneratePDF()
if err != nil { log.Fatal(err) }
_ = first
```

Borrowed repeat path: `(*PreparedTemplate) GeneratePDFBorrowed() (*BorrowedPDF, error)`.

```go
prep, _ := gopdflib.PrepareTemplate(tmpl)
doc, err := prep.GeneratePDFBorrowed()
if err != nil { log.Fatal(err) }
defer doc.Release()
out := doc.CopyBytes()
_ = out
```

## Props

Typed form of the props grammar `FontName:FontSize:StyleCode:Align:L:R:T:B`. Types: `FontOpts`, `Align`, `Borders`, `Color`. Consts: `AlignLeft`, `AlignCenter`, `AlignRight`.

```go
s := gopdflib.FontOpts{Name: "Helvetica", Size: 12, Bold: true, Align: gopdflib.AlignCenter}.String()
opts := gopdflib.ParseFontOpts(s)
fmt.Println(s, opts.Bold, opts.Align)
```

`MakeProps(name string, size int, bold, italic, underline bool, align string, borders [4]int) string` is the function form of `FontOpts.String()`.

```go
props := gopdflib.MakeProps("Helvetica", 12, true, false, false, "center", [4]int{1, 1, 1, 1})
cell := gopdflib.NewCell("Title", props)
fmt.Println(cell.Props)
```

## Caching and BOPS

`ClearBOPSCaches()` drops content caches (subset, page compress, font objects, props, images, signers) without dropping `sync.Pool` buffers. `SetCacheTTL(d time.Duration)`, `CacheTTL() time.Duration`, `DefaultCacheTTL`.

```go
gopdflib.ClearBOPSCaches()
gopdflib.SetCacheTTL(2 * time.Minute)
fmt.Println(gopdflib.CacheTTL())
```

## Errors

Sentinels: `ErrInvalidInput`, `ErrLimitExceeded`, `ErrUpstream`, `ErrInternal`. Codes: `CodeInvalidInput`, `CodeLimitExceeded`, `CodeUpstream`, `CodeInternal`. Envelope: `ErrorCode`, `ErrorEnvelope`.

```go
_, err := gopdflib.MergePDFs(nil)
if errors.Is(err, gopdflib.ErrInvalidInput) {
    log.Println(gopdflib.CodeOf(err))
}
env := gopdflib.EnvelopeOf(err)
fmt.Println(env.Code, env.Message)
```

Mapping helpers: `EnvelopeJSON(err error) string`, `CodeForStatus(status int) ErrorCode`, `StatusForCode(code ErrorCode) int`, `ClassifyMessage(err error) ErrorCode`.

```go
code := gopdflib.CodeForStatus(http.StatusTooManyRequests)
status := gopdflib.StatusForCode(code)
fmt.Println(code, status, gopdflib.EnvelopeJSON(err))
```

## Links

- [PDF_SUITE_GUIDE.md](PDF_SUITE_GUIDE.md) - full suite walkthrough across all ops
- [BUILDER_FLUENT_GO.md](BUILDER_FLUENT_GO.md) - fluent builder and font chains
- [TEMPLATE_REFERENCE.md](TEMPLATE_REFERENCE.md) - template JSON fields and props grammar
- [builder-snippets sample](../sampledata/gopdflib/builder-snippets/main.go) - runnable builder samples
