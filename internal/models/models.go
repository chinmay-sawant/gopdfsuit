package models

type PDFTemplate struct {
	Config Config  `json:"config"`
	Title  Title   `json:"title"`
	Table  []Table `json:"table"`
	Footer Footer  `json:"footer"`
}

type Config struct {
	PageBorder    string `json:"pageBorder"`
	Page          string `json:"page"`          // Page size: "A4", "Letter", "Legal", etc.
	PageAlignment int    `json:"pageAlignment"` // 1 = Portrait (vertical), 2 = Landscape (horizontal)
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
