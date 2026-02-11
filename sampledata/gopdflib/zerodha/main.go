// Package main demonstrates the Zerodha "Gold Standard Benchmark" for gopdflib.
// It runs a weighted 80/15/5 mix of Retail (1-page), Active Trader (2-3 page),
// and HFT (50+ page) contract note generation to simulate a real-world
// brokerage workload.
//
// Run with: go run sampledata/gopdflib/zerodha/main.go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func boolPtr(b bool) *bool        { return &b }
func floatPtr(f float64) *float64 { return &f }

func getSystemInfo() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("OS: %s, Arch: %s, NumCPU: %d, GoVersion: %s",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.Version())
}

func monitorMemory(done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	var maxAlloc uint64
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			fmt.Printf("  Max Memory Allocated: %.2f MB\n", float64(maxAlloc)/1024/1024)
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.Alloc > maxAlloc {
				maxAlloc = m.Alloc
			}
		}
	}
}

// ──────────────────────────────────────────────
// Trade symbols pool
// ──────────────────────────────────────────────

var symbols = []string{
	"RELIANCE", "TCS", "INFY", "HDFCBANK", "TATASTEEL",
	"ICICIBANK", "SBIN", "WIPRO", "BHARTIARTL", "LT",
	"NIFTY24FEB22000CE", "NIFTY24FEB22000PE",
	"BANKNIFTY24FEB46000CE", "BANKNIFTY24FEB46000PE",
	"AXISBANK", "KOTAKBANK", "MARUTI", "TITAN", "ADANIENT", "BAJFINANCE",
}

var actions = []string{"BUY", "SELL"}

type trade struct {
	ID     int
	Time   string
	Symbol string
	ISIN   string
	Action string
	Qty    int
	Price  float64
	Total  float64
}

func generateTrades(n int, rng *rand.Rand) []trade {
	trades := make([]trade, n)
	hour, min, sec := 9, 15, 0
	for i := 0; i < n; i++ {
		sym := symbols[rng.Intn(len(symbols))]
		action := actions[rng.Intn(2)]
		qty := (rng.Intn(50) + 1) * 10 // 10..500
		price := 100.0 + rng.Float64()*3400.0
		price = float64(int(price*100)) / 100 // round to 2 decimals
		total := float64(qty) * price

		timeStr := fmt.Sprintf("%02d:%02d:%02d", hour, min, sec)
		sec++
		if sec >= 60 {
			sec = 0
			min++
		}
		if min >= 60 {
			min = 0
			hour++
		}

		trades[i] = trade{
			ID:     i + 1,
			Time:   timeStr,
			Symbol: sym,
			Action: action,
			Qty:    qty,
			Price:  price,
			Total:  total,
		}
	}
	return trades
}

// ──────────────────────────────────────────────
// Template 1: Retail Investor (1 page, 2 trades)
// ──────────────────────────────────────────────

func buildRetailTemplate() gopdflib.PDFTemplate {
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
				Enabled:     true,
				Visible:     true,
				Name:        "Zerodha Compliance",
				Reason:      "I am the author of this document",
				Location:    "Mumbai, India",
				ContactInfo: "compliance@brokerage.com",
				// TODO: Replace with real PEM certs
				PrivateKeyPEM:  "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDFR55rSZyF0oGt\nJhn7kHXowBy+FhZLl7zJhMp7tCJy5rl6yh3xaf0BwNp/j0WDToTayLimpfCWtGrZ\nV5VEjzGMdtD3RvmHWZKMk5SHot80k+FtWVof3M8H5LpLf8Ye7CgfTMk6lsH7uLHI\nresZXF2Vle3KYDCcj/ZtOlamv+5SGOVGOyIXSaamerArUpHkHkirokr1sq8bSWFv\nYyxyzrLJIZ1jqqwNBBMdtQP7MZnDMekveQU5XEWFJz2n7/PmhJu/c+aw2uVAGZ4l\nsDIq99CejE2vmdcTiAZsKY0INfik/2cpwOiKL4AQuNVD0cbwH2paZjTdk5CTimqI\nviCEsQzVAgMBAAECggEAAsYKzFfAzQGnpxQl216T+c2LQE6CyiJ8M5noit9+eH6v\nIhkDXY7vt32YCAd71hvD5gH0CD74m49pZsLcRPywzD72mY0BTYDZ4zYT9lA45fFX\nHJ8PLR5N0guW/u0kXCNWPd82sqctKDY+WAolIW2MantgJRWmUun6cF1/AfspOGg9\npzEroFMILwVaN+yib5MPZxOWG2qxf9jZAJEgn0W5isgOWyL347tgBQHLbqxTm5FF\nh6bz8nqRUwBLYbmcOswSpJZEm3kQGiTyznGiC1NDZMzLHWwZj1Dc/NCR9wVLXROh\nHx4muAq/Zry8mBdED08OIkqoFIKaFCBQRiGLYX+mYQKBgQDNgnailOR8vDXU9qPA\nXhe/Azn2NkLnJqV7wk5WI4aJaf1Ff8ebeZHHTJIvcA93sOAyCxi7pRz3gEUYjKVi\nzoBZHu+3LtHMje8dSjoKEUI7dLZuXCZ9hhoFQGr8nLuxZRCVB5/NPDrjWlJEnqpJ\npQccmxGCoEKyLMEr79DmtRGheQKBgQD1v4mz0+WM7AJi76c3fNSIgi4Zyik7xmw0\naU72wuCUXYfwF4fCh/tIp5FJqUMsYYld+i1jPhXn8zUq74/RDNGuvrbdZTtkYIXs\n/nVMdAhKL2VQ5t4La4bs+ml+6yfSApaUd8tbCt3ttNyqT0kfHP9XisQ6cepnuqPK\nufv5yQBrPQKBgC7jqn/T6wIOy1WI5Lnafh6F9O6ZWNB2v+Ep50e+GU83EKOP0RJH\nPZy0etI6Bj1v7OdeIsmFlcNez+UXChEuPpiW92jbVOEQLVOIgQ+U+oCoU4uAmQOg\n2kUCeqaieCy0e4EVWT+xk1oWXJjtfrsI3UOImgks2arfjT+iGw7Yl2o5AoGAS5pL\ngNlVq48IBOv5o6ZxtDVofWKmYM9ghpdHRb8aXEqSAZkbmQtAkU+L8P9zvPmcyx6m\nS/vTvXIjDzx4IDYzY/EkTORR60WOriRybbzcuAXww3zjHtxLvCglwHgT3hYRwUdB\ndpbXQ8P6hyKxOjMvkv0L9XcKSDMxJLMnA+eEi3kCgYAJB6nr6W6KQRyotXM1gWP+\n9Ff2Zd5figCt/zD61gw5SAhMLMaR+dj/mfSrDu20jXKr4f/WYuQSRZXxSDHk/pDs\nIXdKNOFnoX+EyvIniTXzQsUWdJVdmZdXseclVfKepUCcQZReYLaesQxdDovlFWjC\nEdvw5H4P31EKT6I5hL6jrA==\n-----END PRIVATE KEY-----",
				CertificatePEM: "-----BEGIN CERTIFICATE-----\nMIIDaTCCAlECFDkGiITntD1ddujydCgb/KNltr4wMA0GCSqGSIb3DQEBCwUAMHkx\nCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVTdGF0ZTENMAsGA1UEBwwEQ2l0eTESMBAG\nA1UECgwJR29QREZTdWl0MRUwEwYDVQQLDAxJbnRlcm1lZGlhdGUxIDAeBgNVBAMM\nF0dvUERGU3VpdEludGVybWVkaWF0ZUNBMB4XDTI2MDExODA4MDIwOVoXDTI3MDEx\nODA4MDIwOVowaTELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQH\nDARDaXR5MRIwEAYDVQQKDAlHb1BERlN1aXQxDTALBgNVBAsMBExlYWYxGDAWBgNV\nBAMMD0dvUERGU3VpdFNpZ25lcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC\nggEBAMVHnmtJnIXSga0mGfuQdejAHL4WFkuXvMmEynu0InLmuXrKHfFp/QHA2n+P\nRYNOhNrIuKal8Ja0atlXlUSPMYx20PdG+YdZkoyTlIei3zST4W1ZWh/czwfkukt/\nxh7sKB9MyTqWwfu4scit6xlcXZWV7cpgMJyP9m06Vqa/7lIY5UY7IhdJpqZ6sCtS\nkeQeSKuiSvWyrxtJYW9jLHLOsskhnWOqrA0EEx21A/sxmcMx6S95BTlcRYUnPafv\n8+aEm79z5rDa5UAZniWwMir30J6MTa+Z1xOIBmwpjQg1+KT/ZynA6IovgBC41UPR\nxvAfalpmNN2TkJOKaoi+IISxDNUCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAh4d+\n3PV4bnOsGQIxSupZMIq+qXf1wcB8dLQWX9ILZz9uXho0E1nhDPHZXRvy/mWG3tZD\nedO8vzMhBY5sD2O8O7K7M+khajfG4gfwhCi3H1dTdze4Wq85K2/kNPqQ/d6qmnS4\nDbxIHWrm8p/wU1p4SYfWFijad9UVutaJixCI9FtCPfRYq5+s0c4cRSyKhjfZp6ic\nhQB01AsgOk1iDgQnSvvjwsz0n1BY/+Apnto3k42PYQx+FNIDIeRvtckVoHxWfmMl\ncWsY6Seqg6V41Yuts78fTKlfjhzI7gKdujl7JMtuyLrL3JVP1rZoMXnjf8SK4QAk\nPkJ5eGE0Ht4i9WkakA==\n-----END CERTIFICATE-----",
				CertificateChain: []string{
					"-----BEGIN CERTIFICATE-----\nMIID0DCCArigAwIBAgIUAO/6vdyeXlJjNCQMAaYSpzVnufkwDQYJKoZIhvcNAQEL\nBQAwaTELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5\nMRIwEAYDVQQKDAlHb1BERlN1aXQxDTALBgNVBAsMBFJvb3QxGDAWBgNVBAMMD0dv\nUERGU3VpdFJvb3RDQTAeFw0yNjAxMTgwODAyMDhaFw0yNzA2MDIwODAyMDhaMHkx\nCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVTdGF0ZTENMAsGA1UEBwwEQ2l0eTESMBAG\nA1UECgwJR29QREZTdWl0MRUwEwYDVQQLDAxJbnRlcm1lZGlhdGUxIDAeBgNVBAMM\nF0dvUERGU3VpdEludGVybWVkaWF0ZUNBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A\nMIIBCgKCAQEAo1lU1M6OuA2KMYe2sXMBrv8LTxEFG7T18qgCnUj7cf0snyTHF/Ws\n07TOYzYchdeKNCrbG0C0ZBhg3f+kGMexYf3CA3Dh1S+k6xbGeEf/LsVBSv8pU4XS\ntXps5/29bdiYWqhMc8Wp7NGbpOAT/PUgLG/xXQiWakYhowLose20d73Kp6PZsPHv\nD9E2IQzDiA/AWtCO+G5NWqagiz0R2fNfuJDPCZQFfURbtrdnEdEs+sAIqw0fzjPg\noSTGed+PcWSil/p/8ZKecKg0WWE7xzcdWdclsUOCh9hJplSyTG7aBssT3ebLYZjX\n4dPx84Lxmhzu+OW68791l1Jt1hCvO0BT2QIDAQABo2AwXjAMBgNVHRMEBTADAQH/\nMA4GA1UdDwEB/wQEAwIBhjAdBgNVHQ4EFgQUp81rlG9kaj7TgQJuKEBJzBbifbIw\nHwYDVR0jBBgwFoAUB/CjFA0QwEkFtNTYsUgutKPWOIEwDQYJKoZIhvcNAQELBQAD\nggEBAMkZEEwxvO7p/TOO97eA/VbXCUwXsU+2990HuWg9MgpBTXiaD61m7w/sD2Gu\nwV2Fvv8rW9GTBhCJ60zV6/HjBBvfGO0ci9xPAHmECVk8Z000vg2tqAHL8HXVJdFF\nYqaLXewHVHUBol2GlSk8sSSss3UNC/wUMBq+Xq0aP/Q2ueidsZhf8DT9aLYCLkdT\nOFp4HI4qJYfpU6h9kLiOcGhc6q/7ToLX8fXXwRrzj9BIQHGPkYQAFg99DYiauBY7\n+bzKhtzxy2ykOvGe7WEzXhKdWEQGkTJSZ8lvsSCafYwuSJ1RW4Hi6TQeP6Rr9qvn\nPYQ/v/v0IZ8qDtBCGTWHz7pNDPQ=\n-----END CERTIFICATE-----",
					"-----BEGIN CERTIFICATE-----\nMIIDszCCApugAwIBAgIUGPNNf0kWGKV8Tg0Kvlqt67kM2OkwDQYJKoZIhvcNAQEL\nBQAwaTELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5\nMRIwEAYDVQQKDAlHb1BERlN1aXQxDTALBgNVBAsMBFJvb3QxGDAWBgNVBAMMD0dv\nUERGU3VpdFJvb3RDQTAeFw0yNjAxMTgwODAyMDhaFw0yODExMDcwODAyMDhaMGkx\nCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVTdGF0ZTENMAsGA1UEBwwEQ2l0eTESMBAG\nA1UECgwJR29QREZTdWl0MQ0wCwYDVQQLDARSb290MRgwFgYDVQQDDA9Hb1BERlN1\naXRSb290Q0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDt+fuF/xXq\n1eZtUjL5PbMGgFatVpE2FAB5upEwehmGRhWo+AMhAXQtCBUsSHMcuCkB+5IQpDPT\nAdZZqni0nnKeKbSL76ryn0EjQHrWVsGa6nddPz1480ZRUXjNbSpmikT5uVc5j1ec\nR3tPw1jtP9B3xjvebEokSLX7Y0nrTPwCQeLIDzpKh80bshvRJ28vmnT38ha4UMOs\nyGV0A70J9ZzUGN9lHM68zDbsbt1ckP9EZRGWRFqjN06vXJpZkLqk/T4LcU+agwK4\n41/fhpMAy3QpYpgC9BNUWAdRzLx/Xl5F8IjGR6vV1dP7O3yKznNEth0ZMSDOsC+n\nX+67D0NLqifLAgMBAAGjUzBRMB0GA1UdDgQWBBQH8KMUDRDASQW01NixSC60o9Y4\ngTAfBgNVHSMEGDAWgBQH8KMUDRDASQW01NixSC60o9Y4gTAPBgNVHRMBAf8EBTAD\nAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQA/ntzzVNBa8bgWO8VigxTsNntGIwn/HR45\n4Og600Ynx+cLQuqIcVwT/stgjg+RO1jBSRSTCtqzbM4/LTgGTbRj4yvgluO6RDdE\n0EsLIioob97jkbLGcMRNGbI4svSBSUytDjhuvmwxz2wBJYGpxZIm6pkgtMeBHrXp\n4750iSj0ORy9TDUUkUdEXfeDBqbjeQ4M1+OaJ5LP3ze09mb1UDGnNKP2nM9m76Pt\ndT/rN+KQKFN48hLnIHMZykEVIoONEzMh3KkfJKhOdTsZrgvwyoLf56qVDCeuADfN\nztHHMRGR4xXSwWkDU/+F00oYhLi63RsFeL4IdGnXb1Tx8VbaPJVm\n-----END CERTIFICATE-----",
				},
			},
		},
		Title: gopdflib.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "CONTRACT NOTE",
			Table: &gopdflib.TitleTable{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{
							Props:     "Helvetica:20:100:center:0:0:0:0",
							Text:      "CONTRACT NOTE - CN2024001",
							BgColor:   "#1B4F72",
							TextColor: "#FFFFFF",
							Height:    floatPtr(45),
						},
					}},
				},
			},
		},
		Elements: []gopdflib.Element{
			// Client Information Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "CLIENT INFORMATION", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			// Client Details
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 4, ColumnWidths: []float64{1.2, 2, 1.2, 2},
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
			// Trades Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "TRADE DETAILS", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			// Trade Table Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 6, ColumnWidths: []float64{2, 1.5, 1, 1, 1.5, 1.5},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:center:1:0:1:1", Text: "Symbol", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:center:0:0:1:1", Text: "ISIN", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:center:0:0:1:1", Text: "Action", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:center:0:0:1:1", Text: "Qty", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:right:0:0:1:1", Text: "Price", BgColor: "#D4E6F1"},
						{Props: "Helvetica:9:100:right:0:1:1:1", Text: "Total", BgColor: "#D4E6F1"},
					}},
					// Trade 1
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:center:1:0:0:1", Text: "TATASTEEL"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "INE081A01012"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "BUY", TextColor: "#27AE60"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "10"},
						{Props: "Helvetica:9:000:right:0:0:0:1", Text: "₹145.50"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "₹1,455.00"},
					}},
					// Trade 2
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
			// Financials
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "FINANCIAL SUMMARY", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 2, ColumnWidths: []float64{2, 1},
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
						{Props: "Helvetica:10:100:left:1:0:1:1", Text: "Total Payable", BgColor: "#D4E6F1"},
						{Props: "Helvetica:10:100:right:0:1:1:1", Text: "₹6,807.50", BgColor: "#D4E6F1"},
					}},
				},
			}},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:7:000:center",
			Text: "ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL",
		},
	}
}

// ──────────────────────────────────────────────
// Template 2: Active Trader (2-3 pages, 40 trades)
// ──────────────────────────────────────────────

func buildActiveTraderTemplate() gopdflib.PDFTemplate {
	rng := rand.New(rand.NewSource(42))
	trades := generateTrades(40, rng)

	// Build trade rows
	tradeRows := make([]gopdflib.Row, 0, 41)
	// Header row
	tradeRows = append(tradeRows, gopdflib.Row{Row: []gopdflib.Cell{
		{Props: "Helvetica:8:100:center:1:0:1:1", Text: "Symbol", BgColor: "#D4E6F1"},
		{Props: "Helvetica:8:100:center:0:0:1:1", Text: "Action", BgColor: "#D4E6F1"},
		{Props: "Helvetica:8:100:center:0:0:1:1", Text: "Qty", BgColor: "#D4E6F1"},
		{Props: "Helvetica:8:100:right:0:0:1:1", Text: "Price", BgColor: "#D4E6F1"},
		{Props: "Helvetica:8:100:right:0:1:1:1", Text: "Total", BgColor: "#D4E6F1"},
	}})
	for i, t := range trades {
		bg := ""
		if i%2 == 1 {
			bg = "#F8F9F9"
		}
		actionColor := "#27AE60"
		if t.Action == "SELL" {
			actionColor = "#E74C3C"
		}
		tradeRows = append(tradeRows, gopdflib.Row{Row: []gopdflib.Cell{
			{Props: "Helvetica:8:000:center:1:0:0:1", Text: t.Symbol, BgColor: bg},
			{Props: "Helvetica:8:000:center:0:0:0:1", Text: t.Action, TextColor: actionColor, BgColor: bg},
			{Props: "Helvetica:8:000:center:0:0:0:1", Text: fmt.Sprintf("%d", t.Qty), BgColor: bg},
			{Props: "Helvetica:8:000:right:0:0:0:1", Text: fmt.Sprintf("₹%.2f", t.Price), BgColor: bg},
			{Props: "Helvetica:8:000:right:0:1:0:1", Text: fmt.Sprintf("₹%.2f", t.Total), BgColor: bg},
		}})
	}

	var totalTurnover float64
	for _, t := range trades {
		totalTurnover += t.Total
	}

	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:          "0:0:0:0",
			Page:                "A4",
			PageAlignment:       1,
			Watermark:           "CONFIDENTIAL",
			PdfTitle:            "Contract Note - Active Trader",
			PDFACompliant:       true,
			ArlingtonCompatible: true,
			EmbedFonts:          boolPtr(true),
		},
		Title: gopdflib.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "ACTIVE TRADER CONTRACT NOTE",
			Table: &gopdflib.TitleTable{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{
							Props:     "Helvetica:18:100:center:0:0:0:0",
							Text:      "CONTRACT NOTE - ACTIVE TRADER",
							BgColor:   "#1B4F72",
							TextColor: "#FFFFFF",
							Height:    floatPtr(45),
						},
					}},
				},
			},
		},
		Elements: []gopdflib.Element{
			// Client Info Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "CLIENT INFORMATION", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			// Client Details
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 4, ColumnWidths: []float64{1.2, 2, 1.2, 2},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Client Name:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "Priya Venkatesh", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:100:left:0:0:0:1", Text: "Client Code:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "PV5544", BgColor: "#EBF5FB"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "PAN:"},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "FGHIJ5678K"},
						{Props: "Helvetica:9:100:left:0:0:0:1", Text: "Trade Date:"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: "2024-02-12"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:left:1:0:0:1", Text: ""},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "Go to Trades", TextColor: "#2E86C1", Link: "#trades-section"},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "Go to Summary", TextColor: "#2E86C1", Link: "#summary-section"},
						{Props: "Helvetica:9:000:left:0:1:0:1", Text: ""},
					}},
				},
			}},
			// Trade Table Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "TRADE DETAILS (40 TRADES)", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			// Trade Table (40 rows + header)
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   5,
				ColumnWidths: []float64{2.5, 1, 1, 1.5, 1.5},
				Rows:         tradeRows,
			}},
			// Summary Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "SUMMARY", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			// Summary
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 2, ColumnWidths: []float64{2, 1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:left:1:0:0:1", Text: "Total Turnover"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: fmt.Sprintf("₹%.2f", totalTurnover)},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:left:1:0:0:1", Text: "Brokerage", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "₹20.00", BgColor: "#F8F9F9"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:left:1:0:0:1", Text: "Regulatory Charges"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "₹150.00"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:10:100:left:1:0:1:1", Text: "Net Payable", BgColor: "#D4E6F1"},
						{Props: "Helvetica:10:100:right:0:1:1:1", Text: fmt.Sprintf("₹%.2f", totalTurnover+20+150), BgColor: "#D4E6F1"},
					}},
				},
			}},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:7:000:center",
			Text: "ZERODHA BROKING LTD | ACTIVE TRADER CONTRACT NOTE | CONFIDENTIAL",
		},
		Bookmarks: []gopdflib.Bookmark{
			{
				Title: "Active Trader Contract Note",
				Page:  1,
				Children: []gopdflib.Bookmark{
					{Title: "Client Information", Page: 1},
					{Title: "Trade Details", Page: 1, Dest: "trades-section"},
					{Title: "Summary", Page: 2, Dest: "summary-section"},
				},
			},
		},
	}
}

// ──────────────────────────────────────────────
// Template 3: HFT / Algo (50+ pages, 2000 trades)
// ──────────────────────────────────────────────

func buildHFTTemplate() gopdflib.PDFTemplate {
	rng := rand.New(rand.NewSource(99))
	trades := generateTrades(2000, rng)

	tradeRows := make([]gopdflib.Row, 0, 2001)
	tradeRows = append(tradeRows, gopdflib.Row{Row: []gopdflib.Cell{
		{Props: "Helvetica:7:100:center:1:0:1:1", Text: "ID", BgColor: "#D4E6F1"},
		{Props: "Helvetica:7:100:center:0:0:1:1", Text: "Time", BgColor: "#D4E6F1"},
		{Props: "Helvetica:7:100:center:0:0:1:1", Text: "Symbol", BgColor: "#D4E6F1"},
		{Props: "Helvetica:7:100:center:0:0:1:1", Text: "Action", BgColor: "#D4E6F1"},
		{Props: "Helvetica:7:100:center:0:0:1:1", Text: "Qty", BgColor: "#D4E6F1"},
		{Props: "Helvetica:7:100:right:0:0:1:1", Text: "Price", BgColor: "#D4E6F1"},
		{Props: "Helvetica:7:100:right:0:1:1:1", Text: "Total", BgColor: "#D4E6F1"},
	}})
	for i, t := range trades {
		bg := ""
		if i%2 == 1 {
			bg = "#F8F9F9"
		}
		actionColor := "#27AE60"
		if t.Action == "SELL" {
			actionColor = "#E74C3C"
		}
		tradeRows = append(tradeRows, gopdflib.Row{Row: []gopdflib.Cell{
			{Props: "Helvetica:7:000:center:1:0:0:1", Text: fmt.Sprintf("%d", t.ID), BgColor: bg},
			{Props: "Helvetica:7:000:center:0:0:0:1", Text: t.Time, BgColor: bg},
			{Props: "Helvetica:7:000:center:0:0:0:1", Text: t.Symbol, BgColor: bg},
			{Props: "Helvetica:7:000:center:0:0:0:1", Text: t.Action, TextColor: actionColor, BgColor: bg},
			{Props: "Helvetica:7:000:center:0:0:0:1", Text: fmt.Sprintf("%d", t.Qty), BgColor: bg},
			{Props: "Helvetica:7:000:right:0:0:0:1", Text: fmt.Sprintf("₹%.2f", t.Price), BgColor: bg},
			{Props: "Helvetica:7:000:right:0:1:0:1", Text: fmt.Sprintf("₹%.2f", t.Total), BgColor: bg},
		}})
	}

	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:          "0:0:0:0",
			Page:                "A4",
			PageAlignment:       1,
			PdfTitle:            "Contract Note - HFT Algo Capital LLP",
			PDFACompliant:       true,
			ArlingtonCompatible: true,
			EmbedFonts:          boolPtr(true),
		},
		Title: gopdflib.Title{
			Props: "Helvetica:18:100:center:0:0:0:0",
			Text:  "HFT CONTRACT NOTE",
			Table: &gopdflib.TitleTable{
				MaxColumns:   1,
				ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{
							Props:     "Helvetica:16:100:center:0:0:0:0",
							Text:      "CONTRACT NOTE - ALGO CAPITAL LLP (HFT)",
							BgColor:   "#1B4F72",
							TextColor: "#FFFFFF",
							Height:    floatPtr(40),
						},
					}},
				},
			},
		},
		Elements: []gopdflib.Element{
			// Client Info
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:10:100:left:1:1:1:1", Text: "CLIENT INFORMATION", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 4, ColumnWidths: []float64{1.2, 2, 1.2, 2},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:8:100:left:1:0:0:1", Text: "Client Name:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:8:000:left:0:0:0:1", Text: "Algo Capital LLP", BgColor: "#EBF5FB"},
						{Props: "Helvetica:8:100:left:0:0:0:1", Text: "Client Code:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:8:000:left:0:1:0:1", Text: "HFT001", BgColor: "#EBF5FB"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:8:100:left:1:0:0:1", Text: "PAN:"},
						{Props: "Helvetica:8:000:left:0:0:0:1", Text: "ZZZZZ9999Z"},
						{Props: "Helvetica:8:100:left:0:0:0:1", Text: "Mode:"},
						{Props: "Helvetica:8:000:left:0:1:0:1", Text: "BATCH PROCESSING"},
					}},
				},
			}},
			// Trade Table Header
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:10:100:left:1:1:1:1", Text: "TRADE DETAILS (2,000 TRADES)", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			// 2000-row trade table
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns:   7,
				ColumnWidths: []float64{0.6, 1, 2, 0.8, 0.6, 1.5, 1.5},
				Rows:         tradeRows,
			}},
			// Compliance audit
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:10:100:left:1:1:1:1", Text: "COMPLIANCE AUDIT", BgColor: "#2E86C1", TextColor: "#FFFFFF"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 2, ColumnWidths: []float64{2, 1},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:8:100:left:1:0:0:1", Text: "Audit Timestamp:"},
						{Props: "Helvetica:8:000:left:0:1:0:1", Text: "2024-02-12T17:00:00Z"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:8:100:left:1:0:0:1", Text: "Auditor Signature:", BgColor: "#F8F9F9"},
						{Props: "Helvetica:8:010:left:0:1:0:1", Text: "[Placeholder]", BgColor: "#F8F9F9"},
					}},
				},
			}},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:7:000:center",
			Text: "ALGO CAPITAL LLP | HFT CONTRACT NOTE | STRICTLY CONFIDENTIAL",
		},
	}
}

// ──────────────────────────────────────────────
// Benchmark Runner
// ──────────────────────────────────────────────

func main() {
	fmt.Println("=== Zerodha Gold Standard Benchmark ===")
	fmt.Println("Workload Mix: 80% Retail | 15% Active | 5% HFT")
	fmt.Println()

	iterations := 5000
	numWorkers := 48

	fmt.Println(getSystemInfo())
	fmt.Printf("Running %d iterations using %d workers...\n\n", iterations, numWorkers)

	// Pre-build all 3 templates
	fmt.Println("Building templates...")
	retailTemplate := buildRetailTemplate()
	activeTemplate := buildActiveTraderTemplate()
	hftTemplate := buildHFTTemplate()
	fmt.Println("Templates built.")

	// Warm-up runs
	fmt.Println("Warm-up runs...")
	retailPDF, err := gopdflib.GeneratePDF(retailTemplate)
	if err != nil {
		fmt.Printf("Error generating retail PDF: %v\n", err)
		os.Exit(1)
	}
	activePDF, err := gopdflib.GeneratePDF(activeTemplate)
	if err != nil {
		fmt.Printf("Error generating active PDF: %v\n", err)
		os.Exit(1)
	}
	hftPDF, err := gopdflib.GeneratePDF(hftTemplate)
	if err != nil {
		fmt.Printf("Error generating HFT PDF: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Retail PDF size:  %d bytes (%.2f KB)\n", len(retailPDF), float64(len(retailPDF))/1024.0)
	fmt.Printf("  Active PDF size:  %d bytes (%.2f KB)\n", len(activePDF), float64(len(activePDF))/1024.0)
	fmt.Printf("  HFT PDF size:     %d bytes (%.2f KB)\n", len(hftPDF), float64(len(hftPDF))/1024.0)
	fmt.Println()

	// Counters
	var retailCount, activeCount, hftCount int64

	// Channels
	jobs := make(chan int, iterations)
	results := make(chan time.Duration, iterations)
	errors := make(chan error, iterations)

	var wg sync.WaitGroup

	// Memory monitor
	memDone := make(chan bool)
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemory(memDone, &memWg)

	// Start workers
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localRng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(10000))))
			for range jobs {
				roll := localRng.Intn(100)
				var tmpl gopdflib.PDFTemplate
				switch {
				case roll < 80:
					tmpl = retailTemplate
					atomic.AddInt64(&retailCount, 1)
				case roll < 95:
					tmpl = activeTemplate
					atomic.AddInt64(&activeCount, 1)
				default:
					tmpl = hftTemplate
					atomic.AddInt64(&hftCount, 1)
				}

				start := time.Now()
				_, err := gopdflib.GeneratePDF(tmpl)
				elapsed := time.Since(start)

				if err != nil {
					errors <- err
					continue
				}
				results <- elapsed
			}
		}()
	}

	// Start timer and send jobs
	totalStart := time.Now()
	for i := 1; i <= iterations; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for workers
	wg.Wait()
	totalTime := time.Since(totalStart)

	// Stop memory monitor
	memDone <- true
	memWg.Wait()

	close(results)
	close(errors)

	// Check errors
	errCount := len(errors)
	if errCount > 0 {
		fmt.Printf("Encountered %d errors during execution.\n", errCount)
		// Print first error
		for e := range errors {
			fmt.Printf("  First error: %v\n", e)
			break
		}
		os.Exit(1)
	}

	// Collect timing data
	var durations []time.Duration
	var sumDuration time.Duration
	for d := range results {
		durations = append(durations, d)
		sumDuration += d
	}

	if len(durations) == 0 {
		fmt.Println("No results collected.")
		return
	}

	var minDuration, maxDuration time.Duration = durations[0], durations[0]
	for _, d := range durations {
		if d < minDuration {
			minDuration = d
		}
		if d > maxDuration {
			maxDuration = d
		}
	}
	avgDuration := sumDuration / time.Duration(len(durations))
	opsPerSec := float64(iterations) / totalTime.Seconds()

	// Print summary
	fmt.Println("=== Performance Summary ===")
	fmt.Printf("  Iterations:      %d\n", iterations)
	fmt.Printf("  Concurrency:     %d workers\n", numWorkers)
	fmt.Printf("  Total time:      %.3f s\n", totalTime.Seconds())
	fmt.Printf("  Throughput:      %.2f ops/sec\n", opsPerSec)
	fmt.Println()
	fmt.Printf("  Avg Latency:     %.3f ms\n", float64(avgDuration.Microseconds())/1000.0)
	fmt.Printf("  Min Latency:     %.3f ms\n", float64(minDuration.Microseconds())/1000.0)
	fmt.Printf("  Max Latency:     %.3f ms\n", float64(maxDuration.Microseconds())/1000.0)
	fmt.Println()
	fmt.Println("=== Workload Distribution ===")
	fmt.Printf("  Retail  (80%%):   %d iterations\n", atomic.LoadInt64(&retailCount))
	fmt.Printf("  Active  (15%%):   %d iterations\n", atomic.LoadInt64(&activeCount))
	fmt.Printf("  HFT      (5%%):   %d iterations\n", atomic.LoadInt64(&hftCount))
	fmt.Println()

	// Save sample PDFs into ./zerodha output directory (relative to working directory)
	outputDir := "sampledata/gopdflib/zerodha"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
	}
	for name, data := range map[string][]byte{
		"zerodha_retail_output.pdf": retailPDF,
		"zerodha_active_output.pdf": activePDF,
		"zerodha_hft_output.pdf":    hftPDF,
	} {
		outPath := filepath.Join(outputDir, name)
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			fmt.Printf("Error saving %s: %v\n", outPath, err)
		} else {
			fmt.Printf("Saved: %s (%d bytes)\n", outPath, len(data))
		}
	}

	fmt.Println()
	fmt.Println("=== Done ===")
}
