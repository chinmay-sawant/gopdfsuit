// Package gopdflib owns the public Go surface of gopdfsuit. Every struct
// below is an owned type: same field names and JSON tags as the internal
// models, but a distinct Go type. The translation adapter in adapter.go is
// the only place that converts between public and internal representations,
// so internal refactors cannot silently change the public API and callers
// recompile untouched as long as field names are stable.
package gopdflib

// PDFTemplate is the main input structure for PDF generation.
type PDFTemplate struct {
	Config    Config     `json:"config"`
	Title     Title      `json:"title"`
	Table     []Table    `json:"table"`
	Spacer    []Spacer   `json:"spacer,omitempty"`
	Image     []Image    `json:"image,omitempty"`
	Elements  []Element  `json:"elements,omitempty"`
	Footer    Footer     `json:"footer"`
	Bookmarks []Bookmark `json:"bookmarks,omitempty"`
}

// Bookmark represents a PDF outline entry for document navigation.
type Bookmark struct {
	Title    string     `json:"title"`
	Dest     string     `json:"dest,omitempty"`
	Page     int        `json:"page,omitempty"`
	Y        float64    `json:"y,omitempty"`
	Children []Bookmark `json:"children,omitempty"`
	Open     bool       `json:"open,omitempty"`
}

// Spacer represents a vertical gap between elements.
type Spacer struct {
	Height float64 `json:"height"`
}

// Element represents a generic document component that can be a table, spacer, or image.
type Element struct {
	Type   string  `json:"type"`
	Index  int     `json:"index,omitempty"`
	Table  *Table  `json:"table,omitempty"`
	Spacer *Spacer `json:"spacer,omitempty"`
	Image  *Image  `json:"image,omitempty"`
}

// Config holds document-wide settings such as page size, margins, and security.
type Config struct {
	PageBorder          string             `json:"pageBorder"`
	PageMargin          string             `json:"pageMargin,omitempty"`
	Page                string             `json:"page"`
	PageAlignment       int                `json:"pageAlignment"`
	Watermark           string             `json:"watermark,omitempty"`
	PdfTitle            string             `json:"pdfTitle,omitempty"`
	ArlingtonCompatible bool               `json:"arlingtonCompatible,omitempty"`
	Bookmarks           []Bookmark         `json:"bookmarks,omitempty"`
	Security            *SecurityConfig    `json:"security,omitempty"`
	PDFA                *PDFAConfig        `json:"pdfa,omitempty"`
	Signature           *SignatureConfig   `json:"signature,omitempty"`
	EmbedFonts          *bool              `json:"embedFonts,omitempty"`
	CustomFonts         []CustomFontConfig `json:"customFonts,omitempty"`
	PDFACompliant       bool               `json:"pdfaCompliant,omitempty"`
	TaggedPDF           bool               `json:"taggedPDF,omitempty"`
}

// SecurityConfig holds PDF encryption and permission settings.
type SecurityConfig struct {
	Enabled               bool   `json:"enabled,omitempty"`
	UserPassword          string `json:"userPassword,omitempty"`
	OwnerPassword         string `json:"ownerPassword"`
	AllowPrinting         bool   `json:"allowPrinting,omitempty"`
	AllowModifying        bool   `json:"allowModifying,omitempty"`
	AllowCopying          bool   `json:"allowCopying,omitempty"`
	AllowAnnotations      bool   `json:"allowAnnotations,omitempty"`
	AllowFormFilling      bool   `json:"allowFormFilling,omitempty"`
	AllowAccessibility    bool   `json:"allowAccessibility,omitempty"`
	AllowAssembly         bool   `json:"allowAssembly,omitempty"`
	AllowHighQualityPrint bool   `json:"allowHighQualityPrint,omitempty"`
}

// PDFAConfig holds PDF/A compliance settings.
type PDFAConfig struct {
	Enabled     bool   `json:"enabled"`
	Conformance string `json:"conformance,omitempty"`
	Title       string `json:"title,omitempty"`
	Author      string `json:"author,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Creator     string `json:"creator,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
}

// SignatureConfig holds digital signature settings.
type SignatureConfig struct {
	Enabled          bool     `json:"enabled"`
	CertificatePEM   string   `json:"certificatePem"`
	PrivateKeyPEM    string   `json:"privateKeyPem"`
	CertificateChain []string `json:"certificateChain,omitempty"`
	Visible          bool     `json:"visible,omitempty"`
	Page             int      `json:"page,omitempty"`
	X                float64  `json:"x,omitempty"`
	Y                float64  `json:"y,omitempty"`
	Width            float64  `json:"width,omitempty"`
	Height           float64  `json:"height,omitempty"`
	Reason           string   `json:"reason,omitempty"`
	Location         string   `json:"location,omitempty"`
	ContactInfo      string   `json:"contactInfo,omitempty"`
	Name             string   `json:"name,omitempty"`
}

// CustomFontConfig specifies a custom font to embed in the PDF.
type CustomFontConfig struct {
	Name     string `json:"name"`
	FilePath string `json:"filePath,omitempty"`
	FontData string `json:"fontData,omitempty"`
}

// Title represents the header section of the document.
type Title struct {
	Props     string      `json:"props"`
	Text      string      `json:"text"`
	Table     *TitleTable `json:"table,omitempty"`
	BgColor   string      `json:"bgcolor,omitempty"`
	TextColor string      `json:"textcolor,omitempty"`
	Link      string      `json:"link,omitempty"`
}

// TitleTable represents an embedded table within the title section.
type TitleTable struct {
	MaxColumns   int       `json:"maxcolumns"`
	ColumnWidths []float64 `json:"columnwidths,omitempty"`
	Rows         []Row     `json:"rows"`
}

// Table represents a grid of data rows and cells.
type Table struct {
	MaxColumns           int       `json:"maxcolumns"`
	Rows                 []Row     `json:"rows"`
	ColumnWidths         []float64 `json:"columnwidths,omitempty"`
	RowHeights           []float64 `json:"rowheights,omitempty"`
	BgColor              string    `json:"bgcolor,omitempty"`
	TextColor            string    `json:"textcolor,omitempty"`
	SharedRowLayout      bool      `json:"sharedRowLayout,omitempty"`
	SharedRowTemplateRow int       `json:"sharedRowTemplateRow,omitempty"`
}

// Row represents a single horizontal line of cells in a table.
type Row struct {
	Row []Cell `json:"row"`
}

// Cell represents a single unit of data within a table row.
type Cell struct {
	Props       string     `json:"props"`
	Text        string     `json:"text,omitempty"`
	Checkbox    *bool      `json:"chequebox,omitempty"`
	Image       *Image     `json:"image,omitempty"`
	Width       *float64   `json:"width,omitempty"`
	Height      *float64   `json:"height,omitempty"`
	FormField   *FormField `json:"form_field,omitempty"`
	BgColor     string     `json:"bgcolor,omitempty"`
	TextColor   string     `json:"textcolor,omitempty"`
	Link        string     `json:"link,omitempty"`
	Wrap        *bool      `json:"wrap,omitempty"`
	Dest        string     `json:"dest,omitempty"`
	MathEnabled *bool      `json:"mathEnabled,omitempty"`
}

// FormField represents a fillable component in a PDF form.
type FormField struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Checked   bool   `json:"checked"`
	GroupName string `json:"group_name,omitempty"`
	Shape     string `json:"shape,omitempty"`
}

// Image represents a visual asset to be embedded in the document.
type Image struct {
	ImageName string  `json:"imagename"`
	ImageData string  `json:"imagedata"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Link      string  `json:"link,omitempty"`
}

// Footer represents the bottom section of each page.
type Footer struct {
	Font string `json:"font"`
	Text string `json:"text"`
	Link string `json:"link,omitempty"`
}

// Props defines the stylistic properties for document elements.
type Props struct {
	FontName  string
	FontSize  int
	StyleCode string
	Bold      bool
	Italic    bool
	Underline bool
	Alignment string
	Borders   [4]int
}

// FontInfo represents a font's information for the API response.
type FontInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Reference   string `json:"reference"`
}

// HTMLToPDFRequest represents the input for HTML to PDF conversion.
type HTMLToPDFRequest struct {
	HTML         string            `json:"html,omitempty"`
	URL          string            `json:"url,omitempty"`
	OutputPath   string            `json:"output_path,omitempty"`
	PageSize     string            `json:"page_size"`
	Orientation  string            `json:"orientation"`
	MarginTop    string            `json:"margin_top"`
	MarginRight  string            `json:"margin_right"`
	MarginBottom string            `json:"margin_bottom"`
	MarginLeft   string            `json:"margin_left"`
	DPI          int               `json:"dpi,omitempty"`
	Grayscale    bool              `json:"grayscale"`
	LowQuality   bool              `json:"low_quality"`
	Options      map[string]string `json:"options,omitempty"`
}

// HTMLToImageRequest represents the input for HTML to image conversion.
type HTMLToImageRequest struct {
	HTML       string            `json:"html,omitempty"`
	URL        string            `json:"url,omitempty"`
	OutputPath string            `json:"output_path,omitempty"`
	Format     string            `json:"format"`
	Width      int               `json:"width,omitempty"`
	Height     int               `json:"height,omitempty"`
	Quality    int               `json:"quality,omitempty"`
	Zoom       float64           `json:"zoom,omitempty"`
	CropWidth  int               `json:"crop_width,omitempty"`
	CropHeight int               `json:"crop_height,omitempty"`
	CropX      int               `json:"crop_x,omitempty"`
	CropY      int               `json:"crop_y,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
}

// SplitSpec defines split criteria for splitting PDFs.
type SplitSpec struct {
	Pages      []int    `json:"pages,omitempty"`
	Ranges     [][2]int `json:"ranges,omitempty"`
	MaxPerFile int      `json:"maxPerFile,omitempty"`
}

// CompressLevel is a Ghostscript-style compression tier.
type CompressLevel string

// Compression tiers for CompressOptions.Level.
const (
	CompressLight  CompressLevel = "light"
	CompressMedium CompressLevel = "medium"
	CompressHeavy  CompressLevel = "heavy"
)

// CompressOptions controls how aggressively CompressPDF rewrites streams and images.
// Empty Level selects Medium (JPEG 75, max image edge 1275). JPEGQuality and
// MaxImageDim override the preset when greater than zero.
//
// Tiers:
//
//	Light  — JPEG 92, max edge 1920
//	Medium — JPEG 75, max edge 1275
//	Heavy  — JPEG 50, max edge 612
type CompressOptions struct {
	Level       CompressLevel `json:"level,omitempty"`
	JPEGQuality int           `json:"jpegQuality,omitempty"`
	MaxImageDim int           `json:"maxImageDim,omitempty"`
}

// PageDetail holds dimensions and metadata for a single page.
type PageDetail struct {
	PageNum int     `json:"pageNum"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// PageInfo holds information about a PDF's pages for redaction.
type PageInfo struct {
	TotalPages int          `json:"totalPages"`
	Pages      []PageDetail `json:"pages"`
}

// TextPosition represents the coordinates and content of a text string on a page.
type TextPosition struct {
	Text   string  `json:"text"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RedactionRect represents a rectangle to be redacted on a page.
type RedactionRect struct {
	PageNum int     `json:"pageNum"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// RedactionTextQuery holds search parameters for text-based redaction.
type RedactionTextQuery struct {
	Text string `json:"text"`
}

// ApplyRedactionOptions configures advanced redaction behavior.
type ApplyRedactionOptions struct {
	Blocks     []RedactionRect      `json:"blocks,omitempty"`
	TextSearch []RedactionTextQuery `json:"textSearch,omitempty"`
	Mode       string               `json:"mode,omitempty"`
	Password   string               `json:"password,omitempty"`
	OCR        *OCRSettings         `json:"ocr,omitempty"`
}

// OCRSettings configures OCR fallback for image-only pages.
type OCRSettings struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider,omitempty"`
	Language string `json:"language,omitempty"`
}

// PageCapability describes if a page can be redacted via text search or requires OCR.
type PageCapability struct {
	PageNum   int    `json:"pageNum"`
	Type      string `json:"type"`
	HasText   bool   `json:"hasText"`
	HasImage  bool   `json:"hasImage"`
	OCREnable bool   `json:"ocrEnabled"`
	Note      string `json:"note,omitempty"`
}

// RedactionApplyReport provides results and warnings from an advanced redaction operation.
type RedactionApplyReport struct {
	Mode              string           `json:"mode"`
	SecurityOutcome   string           `json:"securityOutcome"`
	AppliedSecure     bool             `json:"appliedSecure"`
	AppliedVisual     bool             `json:"appliedVisual"`
	GeneratedRects    int              `json:"generatedRects"`
	AppliedRectangles int              `json:"appliedRectangles"`
	MatchedTextCount  int              `json:"matchedTextCount"`
	Capabilities      []PageCapability `json:"capabilities,omitempty"`
	UnsupportedPages  []int            `json:"unsupportedPages,omitempty"`
	Warnings          []string         `json:"warnings,omitempty"`
}
