package benchmarktemplates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

func boolPtr(value bool) *bool {
	return &value
}

func floatPtr(value float64) *float64 {
	return &value
}

func repoRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}

func readText(relativePath string) (string, error) {
	content, err := os.ReadFile(filepath.Join(repoRoot(), relativePath))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func readChain() ([]string, error) {
	chainText, err := readText(filepath.Join("certs", "chain.pem"))
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(chainText), "-----END CERTIFICATE-----")
	chain := make([]string, 0, len(parts))
	for _, part := range parts {
		stripped := strings.TrimSpace(part)
		if stripped == "" {
			continue
		}
		var sb strings.Builder
		sb.WriteString(stripped)
		sb.WriteString("\n-----END CERTIFICATE-----")
		chain = append(chain, sb.String())
	}
	return chain, nil
}

// BuildZerodhaTemplate constructs the sample Zerodha retail contract note used for benchmarks.
func BuildZerodhaTemplate() (gopdflib.PDFTemplate, error) {
	privateKey, err := readText(filepath.Join("certs", "leaf.key"))
	if err != nil {
		return gopdflib.PDFTemplate{}, err
	}

	certificate, err := readText(filepath.Join("certs", "leaf.pem"))
	if err != nil {
		return gopdflib.PDFTemplate{}, err
	}

	chain, err := readChain()
	if err != nil {
		return gopdflib.PDFTemplate{}, err
	}

	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:          "0:0:0:0",
			Page:                "A4",
			PageAlignment:       1,
			PdfTitle:            "Contract Note - CN2024001",
			PDFACompliant:       true,
			ArlingtonCompatible: true,
			EmbedFonts:          boolPtr(true),
			Signature: &gopdflib.SignatureConfig{
				Enabled:          true,
				Visible:          true,
				Name:             "Zerodha Compliance",
				Reason:           "I am the author of this document",
				Location:         "Mumbai, India",
				ContactInfo:      "compliance@brokerage.com",
				PrivateKeyPEM:    privateKey,
				CertificatePEM:   certificate,
				CertificateChain: chain,
			},
		},
		Title: gopdflib.Title{
			Props: "Helvetica:24:100:center:0:0:0:0",
			Text:  "CONTRACT NOTE",
			Table: &gopdflib.TitleTable{
				MaxColumns:   2,
				ColumnWidths: []float64{1.5, 2.5},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{
							Props:     "Helvetica:20:100:center:0:0:0:0",
							Text:      "CONTRACT NOTE",
							BgColor:   "#154360",
							TextColor: "#FFFFFF",
							Height:    floatPtr(45),
						},
						{
							Props:     "Helvetica:11:000:right:0:0:0:0",
							Text:      "CN2024001 | 2024-02-12",
							BgColor:   "#154360",
							TextColor: "#AED6F1",
							Height:    floatPtr(45),
						},
					}},
				},
			},
		},
		Elements: []gopdflib.Element{
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   4,
				ColumnWidths: []float64{1, 1, 1, 1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:8:000:center:0:0:0:1", Text: "Go to Client Info", TextColor: "#2E86C1", Link: "#client-info"},
					{Props: "Helvetica:8:000:center:0:0:0:1", Text: "Go to Trades", TextColor: "#2E86C1", Link: "#trade-details"},
					{Props: "Helvetica:8:000:center:0:0:0:1", Text: "Go to Financials", TextColor: "#2E86C1", Link: "#financial-summary"},
					{Props: "Helvetica:8:000:center:0:0:0:1", Text: ""},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "SECTION A: CLIENT INFORMATION", BgColor: "#21618C", TextColor: "#FFFFFF", Dest: "client-info"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   4,
				ColumnWidths: []float64{1.2, 2, 1.2, 2},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Client Name:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "Rahul Sharma", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:100:left:0:0:0:1", Text: "Client Code:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "RS9988", BgColor: "#EBF5FB"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "PAN:"},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "ABCDE1234F"},
						{Props: "Helvetica:9:100:left:0:0:0:1", Text: "Trade Date:"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "2024-02-12"},
					}},
				},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "SECTION B: TRADE DETAILS", BgColor: "#21618C", TextColor: "#FFFFFF", Dest: "trade-details"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   6,
				ColumnWidths: []float64{2, 1.5, 1, 1, 1.5, 1.5},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:center:1:0:1:1", Text: "Symbol", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:center:0:0:1:1", Text: "ISIN", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:center:0:0:1:1", Text: "Action", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:center:0:0:1:1", Text: "Qty", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:right:0:0:1:1", Text: "Price", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:right:0:1:1:1", Text: "Total", BgColor: "#D4E6F1"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:center:1:0:0:1", Text: "TATASTEEL"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "INE081A01012"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "BUY", TextColor: "#27AE60"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "10"},
						{Props: "Helvetica:9:000:right:0:0:0:1", Text: "₹145.50"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "₹1,455.00"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:center:1:0:0:1", Text: "INFY", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "INE009A01021", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "SELL", TextColor: "#E74C3C", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "5", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:right:0:0:0:1", Text: "₹1,650.00", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "₹8,250.00", BgColor: "#F8F9F9"},
					}},
				},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "SECTION C: FINANCIAL SUMMARY", BgColor: "#21618C", TextColor: "#FFFFFF", Dest: "financial-summary"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   2,
				ColumnWidths: []float64{2, 1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:left:1:0:0:1", Text: "Net Obligation"},
						{Props: "Helvetica:9:100:right:0:1:0:1", Text: "₹6,795.00"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:left:1:0:0:1", Text: "STT Tax", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "₹12.50", BgColor: "#F8F9F9"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:10:100:left:1:0:1:1", Text: "Total Payable", BgColor: "#A9CCE3"},
						{Props: "Helvetica:10:100:right:0:1:1:1", Text: "₹6,807.50", BgColor: "#A9CCE3"},
					}},
				},
			}},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:7:000:center",
			Text: "ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL",
		},
		Bookmarks: []gopdflib.Bookmark{
			{
				Title: "Contract Note - CN2024001",
				Page:  1,
				Children: []gopdflib.Bookmark{
					{Title: "Client Information", Page: 1, Dest: "client-info"},
					{Title: "Trade Details", Page: 1, Dest: "trade-details"},
					{Title: "Financial Summary", Page: 1, Dest: "financial-summary"},
				},
			},
		},
	}, nil
}

// BenchmarkHeader formats a standard heading for benchmark output.
func BenchmarkHeader(name string) string {
	return fmt.Sprintf("=== %s Single Zerodha Benchmark ===", name)
}
