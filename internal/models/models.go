package models

type PDFTemplate struct {
	Config Config  `json:"config"`
	Title  Title   `json:"title"`
	Table  []Table `json:"table"`
	Image  []Image `json:"image,omitempty"`
	Footer Footer  `json:"footer"`
}

type Config struct {
	PageBorder    string `json:"pageBorder"`
	Page          string `json:"page"`                // Page size: "A4", "Letter", "Legal", etc.
	PageAlignment int    `json:"pageAlignment"`       // 1 = Portrait (vertical), 2 = Landscape (horizontal)
	Watermark     string `json:"watermark,omitempty"` // Optional diagonal watermark text
}

type Title struct {
	Props string `json:"props"`
	Text  string `json:"text"`
}

type Table struct {
	MaxColumns int   `json:"maxcolumns"`
	Rows       []Row `json:"rows"`
}

type Row struct {
	Row []Cell `json:"row"`
}

type Cell struct {
	Props    string `json:"props"`
	Text     string `json:"text,omitempty"`
	Checkbox *bool  `json:"chequebox,omitempty"`
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
