package models

type PDFTemplate struct {
	Config   Config    `json:"config"`
	Title    Title     `json:"title"`
	Table    []Table   `json:"table"`
	Spacer   []Spacer  `json:"spacer,omitempty"`
	Image    []Image   `json:"image,omitempty"`
	Elements []Element `json:"elements,omitempty"` // Ordered elements (tables, spacers, images)
	Footer   Footer    `json:"footer"`
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
	PageBorder          string `json:"pageBorder"`
	Page                string `json:"page"`                          // Page size: "A4", "Letter", "Legal", etc.
	PageAlignment       int    `json:"pageAlignment"`                 // 1 = Portrait (vertical), 2 = Landscape (horizontal)
	Watermark           string `json:"watermark,omitempty"`           // Optional diagonal watermark text
	ArlingtonCompatible bool   `json:"arlingtonCompatible,omitempty"` // Enable PDF 2.0 Arlington Model compliance (full font metrics)
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
}

type Footer struct {
	Font string `json:"font"`
	Text string `json:"text"`
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
