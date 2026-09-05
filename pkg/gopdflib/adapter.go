package gopdflib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

// This file is the only place that converts between owned public types and
// internal engine types. Template inputs use explicit deep conversion so
// caller-owned memory stays isolated without a serialization pass. Opaque
// engine handles (BorrowedPDF, CustomFontRegistry) are not translated: they
// are borrowed, never serialized.

func toInternal[Pub any, In any](pub Pub) (In, error) {
	var zero In
	raw, err := sonic.Marshal(pub)
	if err != nil {
		return zero, err
	}
	if err := sonic.Unmarshal(raw, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func fromInternal[In any, Pub any](in In) (Pub, error) {
	var zero Pub
	raw, err := sonic.Marshal(in)
	if err != nil {
		return zero, err
	}
	if err := sonic.Unmarshal(raw, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func toInternalTemplate(t PDFTemplate) (models.PDFTemplate, error) {
	in := models.PDFTemplate{
		Config:    toInternalConfig(t.Config),
		Title:     toInternalTitle(t.Title),
		Table:     toInternalTables(t.Table),
		Spacer:    toInternalSpacers(t.Spacer),
		Image:     toInternalImages(t.Image),
		Elements:  toInternalElements(t.Elements),
		Footer:    toInternalFooter(t.Footer),
		Bookmarks: toInternalBookmarks(t.Bookmarks),
	}
	in.SetPrecomputedStandardFonts(t.precomputedStandardFonts...)
	return in, nil
}

func toInternalConfig(c Config) models.Config {
	return models.Config{
		PageBorder:          c.PageBorder,
		PageMargin:          c.PageMargin,
		Page:                c.Page,
		PageAlignment:       c.PageAlignment,
		Watermark:           c.Watermark,
		PdfTitle:            c.PdfTitle,
		ArlingtonCompatible: c.ArlingtonCompatible,
		Bookmarks:           toInternalBookmarks(c.Bookmarks),
		Security:            toInternalSecurity(c.Security),
		PDFA:                toInternalPDFA(c.PDFA),
		Signature:           toInternalSignature(c.Signature),
		EmbedFonts:          cloneBool(c.EmbedFonts),
		EmbedStandardFonts:  cloneBool(c.EmbedStandardFonts),
		CustomFonts:         toInternalCustomFonts(c.CustomFonts),
		PDFACompliant:       c.PDFACompliant,
		TaggedPDF:           c.TaggedPDF,
	}
}

func toInternalSecurity(c *SecurityConfig) *models.SecurityConfig {
	if c == nil {
		return nil
	}
	return &models.SecurityConfig{
		Enabled:               c.Enabled,
		UserPassword:          c.UserPassword,
		OwnerPassword:         c.OwnerPassword,
		AllowPrinting:         c.AllowPrinting,
		AllowModifying:        c.AllowModifying,
		AllowCopying:          c.AllowCopying,
		AllowAnnotations:      c.AllowAnnotations,
		AllowFormFilling:      c.AllowFormFilling,
		AllowAccessibility:    c.AllowAccessibility,
		AllowAssembly:         c.AllowAssembly,
		AllowHighQualityPrint: c.AllowHighQualityPrint,
	}
}

func toInternalPDFA(c *PDFAConfig) *models.PDFAConfig {
	if c == nil {
		return nil
	}
	return &models.PDFAConfig{
		Enabled:     c.Enabled,
		Conformance: c.Conformance,
		Title:       c.Title,
		Author:      c.Author,
		Subject:     c.Subject,
		Creator:     c.Creator,
		Keywords:    c.Keywords,
	}
}

func toInternalSignature(c *SignatureConfig) *models.SignatureConfig {
	if c == nil {
		return nil
	}
	return &models.SignatureConfig{
		Enabled:          c.Enabled,
		CertificatePEM:   c.CertificatePEM,
		PrivateKeyPEM:    c.PrivateKeyPEM,
		CertificateChain: append([]string(nil), c.CertificateChain...),
		Visible:          c.Visible,
		Page:             c.Page,
		X:                c.X,
		Y:                c.Y,
		Width:            c.Width,
		Height:           c.Height,
		Reason:           c.Reason,
		Location:         c.Location,
		ContactInfo:      c.ContactInfo,
		Name:             c.Name,
	}
}

func toInternalCustomFonts(fonts []CustomFontConfig) []models.CustomFontConfig {
	if len(fonts) == 0 {
		return nil
	}
	out := make([]models.CustomFontConfig, len(fonts))
	for i, font := range fonts {
		out[i] = models.CustomFontConfig{
			Name:     font.Name,
			FilePath: font.FilePath,
			FontData: font.FontData,
		}
	}
	return out
}

func toInternalTitle(title Title) models.Title {
	return models.Title{
		Props:     title.Props,
		Text:      title.Text,
		TextProps: title.TextProps,
		Table:     toInternalTitleTable(title.Table),
		BgColor:   title.BgColor,
		TextColor: title.TextColor,
		Link:      title.Link,
	}
}

func toInternalTitleTable(table *TitleTable) *models.TitleTable {
	if table == nil {
		return nil
	}
	return &models.TitleTable{
		MaxColumns:   table.MaxColumns,
		ColumnWidths: append([]float64(nil), table.ColumnWidths...),
		Rows:         toInternalRows(table.Rows),
	}
}

func toInternalTables(tables []Table) []models.Table {
	if len(tables) == 0 {
		return nil
	}
	out := make([]models.Table, len(tables))
	for i, table := range tables {
		out[i] = toInternalTable(table)
	}
	return out
}

func toInternalTable(table Table) models.Table {
	return models.Table{
		MaxColumns:           table.MaxColumns,
		Rows:                 toInternalRows(table.Rows),
		ColumnWidths:         append([]float64(nil), table.ColumnWidths...),
		RowHeights:           append([]float64(nil), table.RowHeights...),
		BgColor:              table.BgColor,
		TextColor:            table.TextColor,
		SharedRowLayout:      table.SharedRowLayout,
		SharedRowTemplateRow: table.SharedRowTemplateRow,
	}
}

func toInternalRows(rows []Row) []models.Row {
	if len(rows) == 0 {
		return nil
	}
	out := make([]models.Row, len(rows))
	for i, row := range rows {
		out[i].Row = toInternalCells(row.Row)
	}
	return out
}

func toInternalCells(cells []Cell) []models.Cell {
	if len(cells) == 0 {
		return nil
	}
	out := make([]models.Cell, len(cells))
	for i, cell := range cells {
		out[i] = models.Cell{
			Props:       cell.Props,
			Text:        cell.Text,
			Checkbox:    cloneBool(cell.Checkbox),
			Image:       toInternalImage(cell.Image),
			Width:       cloneFloat(cell.Width),
			Height:      cloneFloat(cell.Height),
			FormField:   toInternalFormField(cell.FormField),
			BgColor:     cell.BgColor,
			TextColor:   cell.TextColor,
			Link:        cell.Link,
			Wrap:        cloneBool(cell.Wrap),
			Dest:        cell.Dest,
			MathEnabled: cloneBool(cell.MathEnabled),
		}
	}
	return out
}

func toInternalFormField(field *FormField) *models.FormField {
	if field == nil {
		return nil
	}
	return &models.FormField{
		Type:      field.Type,
		Name:      field.Name,
		Value:     field.Value,
		Checked:   field.Checked,
		GroupName: field.GroupName,
		Shape:     field.Shape,
	}
}

func toInternalSpacers(spacers []Spacer) []models.Spacer {
	if len(spacers) == 0 {
		return nil
	}
	out := make([]models.Spacer, len(spacers))
	for i, spacer := range spacers {
		out[i].Height = spacer.Height
	}
	return out
}

func toInternalImages(images []Image) []models.Image {
	if len(images) == 0 {
		return nil
	}
	out := make([]models.Image, len(images))
	for i, image := range images {
		out[i] = *toInternalImage(&image)
	}
	return out
}

func toInternalImage(image *Image) *models.Image {
	if image == nil {
		return nil
	}
	return &models.Image{
		ImageName: image.ImageName,
		ImageData: image.ImageData,
		Width:     image.Width,
		Height:    image.Height,
		Link:      image.Link,
	}
}

func toInternalElements(elements []Element) []models.Element {
	if len(elements) == 0 {
		return nil
	}
	out := make([]models.Element, len(elements))
	for i, element := range elements {
		out[i] = models.Element{
			Type:   element.Type,
			Index:  element.Index,
			Table:  toInternalTablePtr(element.Table),
			Spacer: toInternalSpacerPtr(element.Spacer),
			Image:  toInternalImage(element.Image),
		}
	}
	return out
}

func toInternalTablePtr(table *Table) *models.Table {
	if table == nil {
		return nil
	}
	in := toInternalTable(*table)
	return &in
}

func toInternalSpacerPtr(spacer *Spacer) *models.Spacer {
	if spacer == nil {
		return nil
	}
	return &models.Spacer{Height: spacer.Height}
}

func toInternalFooter(footer Footer) models.Footer {
	return models.Footer{
		Font:  footer.Font,
		Text:  footer.Text,
		Props: footer.Props,
		Link:  footer.Link,
	}
}

func toInternalBookmarks(bookmarks []Bookmark) []models.Bookmark {
	if len(bookmarks) == 0 {
		return nil
	}
	out := make([]models.Bookmark, len(bookmarks))
	for i, bookmark := range bookmarks {
		out[i] = models.Bookmark{
			Title:    bookmark.Title,
			Dest:     bookmark.Dest,
			Page:     bookmark.Page,
			Y:        bookmark.Y,
			Children: toInternalBookmarks(bookmark.Children),
			Open:     bookmark.Open,
		}
	}
	return out
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

// ParseCompressLevel normalizes a level string the same way the engine does:
// case-insensitive, empty or unknown selects Medium.
func ParseCompressLevel(s string) CompressLevel {
	switch CompressLevel(strings.ToLower(strings.TrimSpace(s))) {
	case CompressLight:
		return CompressLight
	case CompressHeavy:
		return CompressHeavy
	default:
		return CompressMedium
	}
}

// compressLevelByNumber maps the 1|2|3 tier numbers shared with the frontend
// (compressLevels.js COMPRESS_LEVELS) and the WASM entry point. Numbers
// outside 1-3 fall back to Medium, matching the frontend levelByValue policy.
func compressLevelByNumber(n int) CompressLevel {
	switch n {
	case 1:
		return CompressLight
	case 3:
		return CompressHeavy
	default:
		return CompressMedium
	}
}

// ToServerLevel normalizes a flexible compress level to its canonical server
// string (light|medium|heavy), mirroring frontend toServerLevel: nil and ""
// select Medium, ints and numeric strings map 1|2|3 (other numbers fall back
// to Medium), and unknown non-numeric strings return an error. Use
// ParseCompressLevel instead when the engine-compatible silent default
// (unknown selects Medium) is wanted.
func ToServerLevel(level any) (CompressLevel, error) {
	switch v := level.(type) {
	case nil:
		return CompressMedium, nil
	case CompressLevel:
		return ToServerLevel(string(v))
	case string:
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			return CompressMedium, nil
		}
		switch CompressLevel(key) {
		case CompressLight, CompressMedium, CompressHeavy:
			return CompressLevel(key), nil
		}
		if n, err := strconv.Atoi(key); err == nil {
			return compressLevelByNumber(n), nil
		}
		return "", fmt.Errorf("%w: invalid compression level: %q (use 1|2|3 or light|medium|heavy)", ErrInvalidInput, v)
	case int:
		return compressLevelByNumber(v), nil
	default:
		return "", fmt.Errorf("%w: invalid compression level type %T (use 1|2|3 or light|medium|heavy)", ErrInvalidInput, level)
	}
}

// ToWasmLevel normalizes a flexible compress level to its WASM tier number
// (1|2|3), mirroring frontend toWasmLevel with the same defaults and the
// same invalid-input error contract as ToServerLevel.
func ToWasmLevel(level any) (int, error) {
	normalized, err := ToServerLevel(level)
	if err != nil {
		return 0, err
	}
	switch normalized {
	case CompressLight:
		return 1, nil
	case CompressHeavy:
		return 3, nil
	default:
		return 2, nil
	}
}

// normalizeCompressOptions applies engine-compatible defaults and caps so
// every entry point (Go, CGO, WASM) shares one policy. The engine re-applies
// its own defaults idempotently, so this never changes output bytes.
func normalizeCompressOptions(o CompressOptions) CompressOptions {
	o.Level = ParseCompressLevel(string(o.Level))
	if o.JPEGQuality <= 0 {
		switch o.Level {
		case CompressLight:
			o.JPEGQuality = 92
		case CompressHeavy:
			o.JPEGQuality = 50
		default:
			o.JPEGQuality = 75
		}
	}
	if o.JPEGQuality > 100 {
		o.JPEGQuality = 100
	}
	if o.MaxImageDim <= 0 {
		switch o.Level {
		case CompressLight:
			o.MaxImageDim = 1920
		case CompressHeavy:
			o.MaxImageDim = 612
		default:
			o.MaxImageDim = 1275
		}
	}
	if o.MaxImageDim > 4096 {
		o.MaxImageDim = 4096
	}
	return o
}

func toInternalCompressOptions(o CompressOptions) compress.Options {
	o = normalizeCompressOptions(o)
	return compress.Options{
		Level:       compress.Level(o.Level),
		JPEGQuality: o.JPEGQuality,
		MaxImageDim: o.MaxImageDim,
	}
}

func toInternalSplitSpec(s SplitSpec) merge.SplitSpec {
	return merge.SplitSpec{
		Pages:      s.Pages,
		Ranges:     s.Ranges,
		MaxPerFile: s.MaxPerFile,
	}
}
