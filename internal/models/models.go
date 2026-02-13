package models

type PDFTemplate struct {
	Config    Config     `json:"config"`
	Title     Title      `json:"title"`
	Table     []Table    `json:"table"`
	Spacer    []Spacer   `json:"spacer,omitempty"`
	Image     []Image    `json:"image,omitempty"`
	Elements  []Element  `json:"elements,omitempty"` // Ordered elements (tables, spacers, images)
	Footer    Footer     `json:"footer"`
	Bookmarks []Bookmark `json:"bookmarks,omitempty"` // Hierarchical bookmarks/outlines
}

// Bookmark represents a PDF outline entry for document navigation
type Bookmark struct {
	Title    string     `json:"title"`              // Display text in bookmark panel
	Dest     string     `json:"dest,omitempty"`     // Named destination (matches cell link #dest)
	Page     int        `json:"page,omitempty"`     // Target page number (1-based), used if Dest is empty
	Y        float64    `json:"y,omitempty"`        // Y position on target page (from top)
	Children []Bookmark `json:"children,omitempty"` // Nested bookmarks for hierarchical structure
	Open     bool       `json:"open,omitempty"`     // Whether children are expanded by default
}

type Spacer struct {
	Height float64 `json:"height"`
}

type Element struct {
	Type   string  `json:"type"`             // "table", "spacer", "image"
	Index  int     `json:"index,omitempty"`  // Index into the respective array (Table, Spacer, Image)
	Table  *Table  `json:"table,omitempty"`  // Inline table data (alternative to index)
	Spacer *Spacer `json:"spacer,omitempty"` // Inline spacer data (alternative to index)
	Image  *Image  `json:"image,omitempty"`  // Inline image data (alternative to index)
}

type Config struct {
	PageBorder          string             `json:"pageBorder"`
	PageMargin          string             `json:"pageMargin,omitempty"`          // Page margins in points: "left:right:top:bottom" (default: "72:72:72:72")
	Page                string             `json:"page"`                          // Page size: "A4", "Letter", "Legal", etc.
	PageAlignment       int                `json:"pageAlignment"`                 // 1 = Portrait (vertical), 2 = Landscape (horizontal)
	Watermark           string             `json:"watermark,omitempty"`           // Optional diagonal watermark text
	PdfTitle            string             `json:"pdfTitle,omitempty"`            // Document title for PDF metadata
	ArlingtonCompatible bool               `json:"arlingtonCompatible,omitempty"` // Enable PDF 2.0 Arlington Model compliance (full font metrics)
	Bookmarks           []Bookmark         `json:"bookmarks,omitempty"`           // Document outline/bookmarks for navigation
	Security            *SecurityConfig    `json:"security,omitempty"`            // Password protection and encryption settings
	PDFA                *PDFAConfig        `json:"pdfa,omitempty"`                // PDF/A compliance settings
	Signature           *SignatureConfig   `json:"signature,omitempty"`           // Digital signature settings
	EmbedFonts          *bool              `json:"embedFonts,omitempty"`          // Control standard font embedding optimization (default: true)
	CustomFonts         []CustomFontConfig `json:"customFonts,omitempty"`         // Custom TTF/OTF fonts to embed
	PDFACompliant       bool               `json:"pdfaCompliant,omitempty"`       // Enable PDF/A-4 compliance mode (PDF 2.0, requires all fonts to be embedded via Liberation fonts)
}

// SecurityConfig holds PDF encryption and permission settings
type SecurityConfig struct {
	Enabled       bool   `json:"enabled,omitempty"`      // Enable/disable encryption
	UserPassword  string `json:"userPassword,omitempty"` // Password to open document (empty = no password to open)
	OwnerPassword string `json:"ownerPassword"`          // Password for full access (required for encryption)
	// Permissions control what users can do with the document
	AllowPrinting         bool `json:"allowPrinting,omitempty"`         // Allow printing (default: true)
	AllowModifying        bool `json:"allowModifying,omitempty"`        // Allow modifying content
	AllowCopying          bool `json:"allowCopying,omitempty"`          // Allow copying text/images
	AllowAnnotations      bool `json:"allowAnnotations,omitempty"`      // Allow adding annotations
	AllowFormFilling      bool `json:"allowFormFilling,omitempty"`      // Allow filling form fields
	AllowAccessibility    bool `json:"allowAccessibility,omitempty"`    // Allow accessibility features
	AllowAssembly         bool `json:"allowAssembly,omitempty"`         // Allow document assembly
	AllowHighQualityPrint bool `json:"allowHighQualityPrint,omitempty"` // Allow high quality printing
}

// PDFAConfig holds PDF/A compliance settings
type PDFAConfig struct {
	Enabled     bool   `json:"enabled"`               // Enable PDF/A compliance
	Conformance string `json:"conformance,omitempty"` // PDF/A conformance level: "1b", "2b", "3b", "4", "4f", "4e" (default: "4")
	Title       string `json:"title,omitempty"`       // Document title for XMP metadata
	Author      string `json:"author,omitempty"`      // Document author for XMP metadata
	Subject     string `json:"subject,omitempty"`     // Document subject for XMP metadata
	Creator     string `json:"creator,omitempty"`     // Creating application name
	Keywords    string `json:"keywords,omitempty"`    // Document keywords (comma-separated)
}

// SignatureConfig holds digital signature settings
type SignatureConfig struct {
	Enabled          bool     `json:"enabled"`                    // Enable digital signing
	CertificatePEM   string   `json:"certificatePem"`             // PEM-encoded X.509 certificate
	PrivateKeyPEM    string   `json:"privateKeyPem"`              // PEM-encoded private key (RSA or ECDSA)
	CertificateChain []string `json:"certificateChain,omitempty"` // Optional intermediate certificates

	// Signature appearance
	Visible bool    `json:"visible,omitempty"` // Show visible signature stamp
	Page    int     `json:"page,omitempty"`    // Page number for visible signature (1-based, default: 1)
	X       float64 `json:"x,omitempty"`       // X position for visible signature
	Y       float64 `json:"y,omitempty"`       // Y position for visible signature
	Width   float64 `json:"width,omitempty"`   // Width of visible signature (default: 200)
	Height  float64 `json:"height,omitempty"`  // Height of visible signature (default: 50)

	// Signature info
	Reason      string `json:"reason,omitempty"`      // Reason for signing
	Location    string `json:"location,omitempty"`    // Location of signing
	ContactInfo string `json:"contactInfo,omitempty"` // Contact information
	Name        string `json:"name,omitempty"`        // Signer name (overrides certificate CN)
}

// CustomFontConfig specifies a custom font to embed in the PDF
type CustomFontConfig struct {
	Name     string `json:"name"`               // Reference name used in props (e.g., "MyFont")
	FilePath string `json:"filePath,omitempty"` // Path to TTF/OTF file (server-side)
	FontData string `json:"fontData,omitempty"` // Base64-encoded font data (alternative to FilePath)
}

type Title struct {
	Props string `json:"props"`
	Text  string `json:"text"`
	// Table allows embedding a table inside the title for complex layouts (e.g., logo + text)
	// When Table is provided, Text is ignored and the table is rendered instead
	Table *TitleTable `json:"table,omitempty"`
	// BgColor is the background color for the title section.
	BgColor string `json:"bgcolor,omitempty"`
	// TextColor is the text color for the title text.
	TextColor string `json:"textcolor,omitempty"`
	// Link URL for the title text
	Link string `json:"link,omitempty"`
}

// TitleTable represents an embedded table within the title section
type TitleTable struct {
	MaxColumns   int       `json:"maxcolumns"`
	ColumnWidths []float64 `json:"columnwidths,omitempty"`
	Rows         []Row     `json:"rows"`
}

type Table struct {
	MaxColumns int   `json:"maxcolumns"`
	Rows       []Row `json:"rows"`
	// ColumnWidths represent relative width weights per column. If empty or length
	// mismatches MaxColumns, widths are evenly distributed. Example: [2,1,1] will
	// allocate 50%,25%,25% of available table width respectively.
	ColumnWidths []float64 `json:"columnwidths,omitempty"`
	// RowHeights override the default row height (25). Values are in PDF points.
	// If a row index is out of bounds or value <=0 the default height is used.
	RowHeights []float64 `json:"rowheights,omitempty"`
	// BgColor is the default background color for all cells in the table.
	// Format: hexadecimal color code (e.g., "#FF0000" for red, "#00000000" for transparent).
	// Individual cell BgColor takes precedence over table BgColor.
	BgColor string `json:"bgcolor,omitempty"`
	// TextColor is the default text/font color for all cells in the table.
	// Format: hexadecimal color code (e.g., "#FF0000" for red). Default is black.
	// Individual cell TextColor takes precedence over table TextColor.
	TextColor string `json:"textcolor,omitempty"`
}

type Row struct {
	Row []Cell `json:"row"`
}

type Cell struct {
	Props    string `json:"props"`
	Text     string `json:"text,omitempty"`
	Checkbox *bool  `json:"chequebox,omitempty"`
	Image    *Image `json:"image,omitempty"` // Support for images in cells
	// Optional explicit width/height for the cell. Width is treated as a weight
	// when ColumnWidths is not supplied at Table level. Height can influence
	// RowHeights if those are not explicitly set (frontend may promote edits
	// to RowHeights / ColumnWidths).
	Width     *float64   `json:"width,omitempty"`
	Height    *float64   `json:"height,omitempty"`
	FormField *FormField `json:"form_field,omitempty"` // Support for fillable form fields
	// BgColor is the background color for this specific cell.
	// Format: hexadecimal color code (e.g., "#FF0000" for red, "#00000000" for transparent).
	// Takes precedence over table-level BgColor.
	BgColor string `json:"bgcolor,omitempty"`
	// TextColor is the text/font color for this specific cell.
	// Format: hexadecimal color code (e.g., "#FF0000" for red). Default is black.
	// Takes precedence over table-level TextColor.
	TextColor string `json:"textcolor,omitempty"`
	// Link is an optional hyperlink for this cell.
	// External links: Use full URL (e.g., "https://example.com")
	// Internal links: Use bookmark name prefixed with # (e.g., "#section1")
	Link string `json:"link,omitempty"`
	// Wrap enables automatic text wrapping within the cell.
	// When true, text will wrap to multiple lines and row height will auto-adjust.
	// Default is false (text will be clipped if it exceeds cell width).
	Wrap *bool `json:"wrap,omitempty"`
	// Dest is an optional named destination anchor for this cell.
	// Other cells can link to it using Link: "#dest-name".
	// Bookmarks can also reference it via the Dest field.
	Dest string `json:"dest,omitempty"`
}

type FormField struct {
	Type      string `json:"type"` // "checkbox", "radio", "text"
	Name      string `json:"name"`
	Value     string `json:"value"` // Export value for radio/checkbox, default value for text
	Checked   bool   `json:"checked"`
	GroupName string `json:"group_name,omitempty"` // For radio buttons
	Shape     string `json:"shape,omitempty"`      // "round" or "square" (for radio)
}

type Image struct {
	ImageName string  `json:"imagename"`
	ImageData string  `json:"imagedata"` // Base64 encoded image data
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	// Link URL for the image
	Link string `json:"link,omitempty"`
}

type Footer struct {
	Font string `json:"font"`
	Text string `json:"text"`
	// Link URL for the footer text
	Link string `json:"link,omitempty"`
}

type Props struct {
	FontName  string
	FontSize  int
	StyleCode string // 3-digit style code: bold(1/0) + italic(1/0) + underline(1/0)
	Bold      bool   // Parsed from first digit of StyleCode
	Italic    bool   // Parsed from second digit of StyleCode
	Underline bool   // Parsed from third digit of StyleCode
	Alignment string
	Borders   [4]int // left, right, top, bottom
}

// htmlToPDFRequest represents the input for htmltopdf conversion
type HtmlToPDFRequest struct {
	HTML         string            `json:"html,omitempty"`        // Raw HTML content
	URL          string            `json:"url,omitempty"`         // URL to convert
	OutputPath   string            `json:"output_path,omitempty"` // Optional output path
	PageSize     string            `json:"page_size"`             // A4, Letter, etc. (default: A4)
	Orientation  string            `json:"orientation"`           // Portrait, Landscape (default: Portrait)
	MarginTop    string            `json:"margin_top"`            // Top margin (default: 10mm)
	MarginRight  string            `json:"margin_right"`          // Right margin (default: 10mm)
	MarginBottom string            `json:"margin_bottom"`         // Bottom margin (default: 10mm)
	MarginLeft   string            `json:"margin_left"`           // Left margin (default: 10mm)
	DPI          int               `json:"dpi,omitempty"`         // DPI for better quality (default: 300)
	Grayscale    bool              `json:"grayscale"`             // Convert to grayscale
	LowQuality   bool              `json:"low_quality"`           // Lower quality for smaller file size
	Options      map[string]string `json:"options,omitempty"`     // Additional htmltopdf options
}

// htmlToImageRequest represents the input for htmltoimage conversion
type HtmlToImageRequest struct {
	HTML       string            `json:"html,omitempty"`        // Raw HTML content
	URL        string            `json:"url,omitempty"`         // URL to convert
	OutputPath string            `json:"output_path,omitempty"` // Optional output path
	Format     string            `json:"format"`                // png, jpg, svg (default: png)
	Width      int               `json:"width,omitempty"`       // Image width in pixels
	Height     int               `json:"height,omitempty"`      // Image height in pixels
	Quality    int               `json:"quality,omitempty"`     // Image quality 1-100 (default: 94)
	Zoom       float64           `json:"zoom,omitempty"`        // Zoom factor (default: 1.0)
	CropWidth  int               `json:"crop_width,omitempty"`  // Crop width
	CropHeight int               `json:"crop_height,omitempty"` // Crop height
	CropX      int               `json:"crop_x,omitempty"`      // Crop X offset
	CropY      int               `json:"crop_y,omitempty"`      // Crop Y offset
	Options    map[string]string `json:"options,omitempty"`     // Additional htmltoimage options
}

// FontInfo represents a font's information for the API response
type FontInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Reference   string `json:"reference"`
}
