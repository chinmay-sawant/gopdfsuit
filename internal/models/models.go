package models

type PDFTemplate struct {
	Config Config  `json:"config"`
	Title  Title   `json:"title"`
	Table  []Table `json:"table"`
	Footer Footer  `json:"footer"`
}

type Config struct {
	PageBorder string `json:"pageBorder"`
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
	Alignment string
	Borders   [4]int // left, right, top, bottom
}
