// Shared gpdf Zerodha gold-standard benchmark (compliant and non-compliant).
// Entry points: main.go (compliant) and main_nocomply.go (build tag nocomply).
package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdfa"
	"github.com/gpdf-dev/gpdf/signature"
	"github.com/gpdf-dev/gpdf/template"
)

var (
	flagCPUProfile = flag.String("cpuprofile", "", "write CPU profile to file")
	flagMemProfile = flag.String("memprofile", "", "write heap profile to file")
)

// benchCompliant is set by the build-tagged main entry (main.go vs main_nocomply.go).
var benchCompliant bool

var (
	colorPrimary  = pdf.RGBHex(0x154360)
	colorSection  = pdf.RGBHex(0x21618C)
	colorHeader   = pdf.RGBHex(0xD4E6F1)
	colorStripe   = pdf.RGBHex(0xF8F9F9)
	colorSubtitle = pdf.RGBHex(0xAED6F1)
)

type pdfGenerator func() ([]byte, error)

type trade struct {
	ID     int
	Time   string
	Symbol string
	Action string
	Qty    int
	Price  float64
	Total  float64
}

var symbols = []string{
	"RELIANCE", "TCS", "INFY", "HDFCBANK", "TATASTEEL",
	"ICICIBANK", "SBIN", "WIPRO", "BHARTIARTL", "LT",
	"NIFTY24FEB22000CE", "NIFTY24FEB22000PE",
	"BANKNIFTY24FEB46000CE", "BANKNIFTY24FEB46000PE",
	"AXISBANK", "KOTAKBANK", "MARUTI", "TITAN", "ADANIENT", "BAJFINANCE",
}

var actions = []string{"BUY", "SELL"}

func generateTrades(n int, rng *rand.Rand) []trade {
	trades := make([]trade, n)
	hour, min, sec := 9, 15, 0
	for i := range n {
		sym := symbols[rng.Intn(len(symbols))]
		action := actions[rng.Intn(2)]
		qty := (rng.Intn(50) + 1) * 10
		price := 100.0 + rng.Float64()*3400.0
		price = float64(int(price*100)) / 100
		total := float64(qty) * price

		var tb [8]byte
		tb[0] = byte('0' + hour/10)
		tb[1] = byte('0' + hour%10)
		tb[2] = ':'
		tb[3] = byte('0' + min/10)
		tb[4] = byte('0' + min%10)
		tb[5] = ':'
		tb[6] = byte('0' + sec/10)
		tb[7] = byte('0' + sec%10)
		timeStr := string(tb[:])
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

func getSystemInfo() string {
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

func baseDocOpts(title string) []template.Option {
	opts := []template.Option{
		gpdf.WithPageSize(gpdf.A4),
		gpdf.WithMargins(document.UniformEdges(document.Mm(12))),
		gpdf.WithMetadata(document.DocumentMetadata{
			Title:  title,
			Author: "gpdf benchmark",
		}),
	}
	if benchCompliant {
		opts = append(opts, gpdf.WithPDFA(
			pdfa.WithLevel(pdfa.LevelA2b),
			pdfa.WithMetadata(pdfa.MetadataInfo{
				Title:    title,
				Author:   "gpdf benchmark",
				Creator:  "gpdf",
				Producer: "gpdf",
			}),
		))
	}
	return opts
}

func contractFooter(text string) func(*template.PageBuilder) {
	return func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text(text, template.FontSize(7), template.AlignCenter())
			})
		})
	}
}

func sectionRow(page *template.PageBuilder, text string) {
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text(text,
				template.FontSize(11),
				template.Bold(),
				template.TextColor(pdf.White),
				template.BgColor(colorSection),
				template.TextPadding(document.UniformEdges(document.Mm(2))),
			)
		})
	})
}

func labelValueTable(page *template.PageBuilder, rows [][4]string, stripe bool) {
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			body := make([][]string, len(rows))
			for i, row := range rows {
				body[i] = []string{row[0], row[1], row[2], row[3]}
			}
			opts := []template.TableOption{
				template.ColumnWidths(15, 35, 15, 35),
				template.WithTableCellBorder(cellBorder()),
			}
			if stripe {
				opts = append(opts, template.TableStripe(colorStripe))
			}
			c.Table(nil, body, opts...)
		})
	})
}

func cellBorder() template.BorderSpec {
	return template.Border(
		template.BorderWidth(document.Pt(0.5)),
		template.BorderColor(pdf.Gray(0.75)),
	)
}

func tradeTableOpts(fontSize float64, widths ...float64) []template.TableOption {
	return []template.TableOption{
		template.ColumnWidths(widths...),
		template.TableHeaderStyle(
			template.TextColor(pdf.Black),
			template.BgColor(colorHeader),
			template.FontSize(fontSize),
			template.Bold(),
		),
		template.TableStripe(colorStripe),
		template.WithTableCellBorder(cellBorder()),
	}
}

func formatINR(value float64) string {
	return "₹" + strconv.FormatFloat(value, 'f', 2, 64)
}

func loadRetailSigner() (signature.Signer, error) {
	certBlock, _ := pem.Decode([]byte(ecLeafCertPEM))
	if certBlock == nil {
		return signature.Signer{}, errors.New("decode certificate PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return signature.Signer{}, err
	}

	keyBlock, _ := pem.Decode([]byte(ecLeafKeyPEM))
	if keyBlock == nil {
		return signature.Signer{}, errors.New("decode private key PEM")
	}
	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return signature.Signer{}, err
	}

	return signature.Signer{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

func maybeSignRetail(data []byte, signer signature.Signer) ([]byte, error) {
	if !benchCompliant {
		return data, nil
	}
	return gpdf.SignDocument(data, signer,
		signature.WithReason("I am the author of this document"),
		signature.WithLocation("Mumbai, India"),
	)
}

func retailSignAlgorithmLabel() string {
	if os.Getenv("BENCH_SIGN_RSA") == "1" {
		return "RSA-2048"
	}
	return "ECDSA P-256"
}

func buildRetailGenerator(signer signature.Signer) pdfGenerator {
	return func() ([]byte, error) {
		doc := gpdf.NewDocument(baseDocOpts("Contract Note - CN2024001")...)
		doc.Footer(contractFooter("ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL"))

		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(7, func(c *template.ColBuilder) {
				c.Text("CONTRACT NOTE",
					template.FontSize(20), template.Bold(),
					template.TextColor(pdf.White), template.BgColor(colorPrimary),
					template.AlignCenter(),
					template.TextPadding(document.UniformEdges(document.Mm(4))),
				)
			})
			r.Col(5, func(c *template.ColBuilder) {
				c.Text("CN2024001 | 2024-02-12",
					template.FontSize(11),
					template.TextColor(colorSubtitle), template.BgColor(colorPrimary),
					template.AlignRight(),
					template.TextPadding(document.UniformEdges(document.Mm(4))),
				)
			})
		})

		sectionRow(page, "SECTION A: CLIENT INFORMATION")
		labelValueTable(page, [][4]string{
			{"Client Name:", "Rahul Sharma", "Client Code:", "RS9988"},
			{"PAN:", "ABCDE1234F", "Trade Date:", "2024-02-12"},
		}, true)

		sectionRow(page, "SECTION B: TRADE DETAILS")
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				header := []string{"Symbol", "ISIN", "Action", "Qty", "Price", "Total"}
				rows := [][]string{
					{"TATASTEEL", "INE081A01012", "BUY", "10", "₹145.50", "₹1,455.00"},
					{"INFY", "INE009A01021", "SELL", "5", "₹1,650.00", "₹8,250.00"},
				}
				c.Table(header, rows, append(tradeTableOpts(9, 20, 15, 10, 8, 15, 15),
					template.ColumnAlign(
						document.AlignCenter, document.AlignCenter, document.AlignCenter,
						document.AlignCenter, document.AlignRight, document.AlignRight,
					),
				)...)
			})
		})

		sectionRow(page, "SECTION C: FINANCIAL SUMMARY")
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Table(nil, [][]string{
					{"Net Obligation", "₹6,795.00"},
					{"STT Tax", "₹12.50"},
					{"Total Payable", "₹6,807.50"},
				},
					template.ColumnWidths(70, 30),
					template.TableStripe(colorStripe),
					template.WithTableCellBorder(cellBorder()),
				)
			})
		})

		data, err := doc.Generate()
		if err != nil {
			return nil, err
		}
		return maybeSignRetail(data, signer)
	}
}

func buildActiveGenerator(trades []trade) pdfGenerator {
	tradeRows := make([][]string, len(trades))
	var totalTurnover float64
	for i, t := range trades {
		totalTurnover += t.Total
		tradeRows[i] = []string{
			t.Symbol,
			t.Action,
			strconv.Itoa(t.Qty),
			formatINR(t.Price),
			formatINR(t.Total),
		}
	}

	return func() ([]byte, error) {
		doc := gpdf.NewDocument(baseDocOpts("Contract Note - Active Trader")...)
		doc.Footer(contractFooter("ZERODHA BROKING LTD | ACTIVE TRADER CONTRACT NOTE | CONFIDENTIAL"))

		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(7, func(c *template.ColBuilder) {
				c.Text("ACTIVE TRADER CONTRACT NOTE",
					template.FontSize(18), template.Bold(),
					template.TextColor(pdf.White), template.BgColor(colorPrimary),
					template.AlignCenter(),
					template.TextPadding(document.UniformEdges(document.Mm(4))),
				)
			})
			r.Col(5, func(c *template.ColBuilder) {
				c.Text("40 Trades | 2024-02-12",
					template.FontSize(11),
					template.TextColor(colorSubtitle), template.BgColor(colorPrimary),
					template.AlignRight(),
					template.TextPadding(document.UniformEdges(document.Mm(4))),
				)
			})
		})

		sectionRow(page, "SECTION A: CLIENT INFORMATION")
		labelValueTable(page, [][4]string{
			{"Client Name:", "Priya Venkatesh", "Client Code:", "PV5544"},
			{"PAN:", "FGHIJ5678K", "Trade Date:", "2024-02-12"},
		}, false)

		sectionRow(page, "SECTION B: TRADE DETAILS (40 TRADES)")
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				header := []string{"Symbol", "Action", "Qty", "Price", "Total"}
				c.Table(header, tradeRows, append(tradeTableOpts(8, 25, 10, 10, 15, 15),
					template.ColumnAlign(
						document.AlignCenter, document.AlignCenter, document.AlignCenter,
						document.AlignRight, document.AlignRight,
					),
				)...)
			})
		})

		sectionRow(page, "SECTION C: SUMMARY")
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Table(nil, [][]string{
					{"Total Turnover", formatINR(totalTurnover)},
					{"Brokerage", "₹20.00"},
					{"Regulatory Charges", "₹150.00"},
					{"Net Payable", formatINR(totalTurnover + 20 + 150)},
				},
					template.ColumnWidths(70, 30),
					template.TableStripe(colorStripe),
					template.WithTableCellBorder(cellBorder()),
				)
			})
		})

		return doc.Generate()
	}
}

func buildHFTGenerator(trades []trade) pdfGenerator {
	tradeRows := make([][]string, len(trades))
	for i, t := range trades {
		tradeRows[i] = []string{
			strconv.Itoa(t.ID),
			t.Time,
			t.Symbol,
			t.Action,
			strconv.Itoa(t.Qty),
			formatINR(t.Price),
			formatINR(t.Total),
		}
	}

	return func() ([]byte, error) {
		doc := gpdf.NewDocument(baseDocOpts("Contract Note - HFT Algo Capital LLP")...)
		doc.Footer(contractFooter("ALGO CAPITAL LLP | HFT CONTRACT NOTE | STRICTLY CONFIDENTIAL"))

		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(7, func(c *template.ColBuilder) {
				c.Text("HFT CONTRACT NOTE",
					template.FontSize(16), template.Bold(),
					template.TextColor(pdf.White), template.BgColor(colorPrimary),
					template.AlignCenter(),
					template.TextPadding(document.UniformEdges(document.Mm(4))),
				)
			})
			r.Col(5, func(c *template.ColBuilder) {
				c.Text("ALGO CAPITAL LLP | 2,000 Trades",
					template.FontSize(10),
					template.TextColor(colorSubtitle), template.BgColor(colorPrimary),
					template.AlignRight(),
					template.TextPadding(document.UniformEdges(document.Mm(4))),
				)
			})
		})

		sectionRow(page, "SECTION A: CLIENT INFORMATION")
		labelValueTable(page, [][4]string{
			{"Client Name:", "Algo Capital LLP", "Client Code:", "HFT001"},
			{"PAN:", "ZZZZZ9999Z", "Mode:", "BATCH PROCESSING"},
		}, true)

		sectionRow(page, "SECTION B: TRADE DETAILS (2,000 TRADES)")
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				header := []string{"ID", "Time", "Symbol", "Action", "Qty", "Price", "Total"}
				c.Table(header, tradeRows, append(tradeTableOpts(7, 6, 10, 22, 8, 6, 14, 14),
					template.ColumnAlign(
						document.AlignCenter, document.AlignCenter, document.AlignCenter,
						document.AlignCenter, document.AlignCenter, document.AlignRight, document.AlignRight,
					),
				)...)
			})
		})

		sectionRow(page, "SECTION C: COMPLIANCE AUDIT")
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Table(nil, [][]string{
					{"Audit Timestamp:", "2024-02-12T17:00:00Z"},
					{"Auditor Signature:", "[Placeholder]"},
				},
					template.ColumnWidths(70, 30),
					template.TableStripe(colorStripe),
					template.WithTableCellBorder(cellBorder()),
				)
			})
		})

		return doc.Generate()
	}
}

func outputPDFName(base string) string {
	if benchCompliant {
		return base
	}
	return strings.Replace(base, "_output.pdf", "_nocomply_output.pdf", 1)
}

func envInt(key string, fallback int) int {
	if raw := os.Getenv(key); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func runBenchmark() error {
	fmt.Println("=== gpdf Zerodha Gold Standard Benchmark ===")
	if benchCompliant {
		fmt.Println("Mode: compliant (PDF/A-2b, signed retail)")
	} else {
		fmt.Println("Mode: non-compliant (PDF/A and signing off)")
	}
	fmt.Println("Workload Mix: 80% Retail | 15% Active | 5% HFT")
	fmt.Println()

	iterations := envInt("BENCH_ITERATIONS", 5000)
	numWorkers := envInt("BENCH_WORKERS", 48)
	skipWrite := os.Getenv("BENCH_SKIP_WRITE") == "1"
	benchSeed := int64(42)
	if raw := os.Getenv("BENCH_SEED"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			benchSeed = n
		}
	}

	fmt.Println(getSystemInfo())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	if benchCompliant {
		fmt.Printf("Running %d iterations using %d workers (retail sign: %s)...\n\n", iterations, numWorkers, retailSignAlgorithmLabel())
	} else {
		fmt.Printf("Running %d iterations using %d workers (retail signing: disabled)...\n\n", iterations, numWorkers)
	}

	var signer signature.Signer
	if benchCompliant {
		var err error
		signer, err = loadRetailSigner()
		if err != nil {
			return fmt.Errorf("load retail signer: %w", err)
		}
	}

	fmt.Println("Building generators...")
	activeTrades := generateTrades(40, rand.New(rand.NewSource(42)))
	hftTrades := generateTrades(2000, rand.New(rand.NewSource(99)))
	retailGen := buildRetailGenerator(signer)
	activeGen := buildActiveGenerator(activeTrades)
	hftGen := buildHFTGenerator(hftTrades)
	fmt.Println("Generators ready.")

	var retailPDF, activePDF, hftPDF []byte
	if os.Getenv("BENCH_WARMUP") != "0" {
		fmt.Println("Warm-up runs...")
		var err error
		retailPDF, err = retailGen()
		if err != nil {
			return fmt.Errorf("error generating retail PDF: %w", err)
		}
		activePDF, err = activeGen()
		if err != nil {
			return fmt.Errorf("error generating active PDF: %w", err)
		}
		hftPDF, err = hftGen()
		if err != nil {
			return fmt.Errorf("error generating HFT PDF: %w", err)
		}
		fmt.Printf("  Retail PDF size:  %d bytes (%.2f KB)\n", len(retailPDF), float64(len(retailPDF))/1024.0)
		fmt.Printf("  Active PDF size:  %d bytes (%.2f KB)\n", len(activePDF), float64(len(activePDF))/1024.0)
		fmt.Printf("  HFT PDF size:     %d bytes (%.2f KB)\n", len(hftPDF), float64(len(hftPDF))/1024.0)
		fmt.Println()
	} else {
		fmt.Println("Warm-up skipped (BENCH_WARMUP=0).")
		fmt.Println()
	}

	var retailCount, activeCount, hftCount int64

	const (
		workloadRetail = iota
		workloadActive
		workloadHFT
	)
	schedule := make([]int, iterations)
	retailTarget := iterations * 80 / 100
	activeTarget := iterations * 15 / 100
	for i := range schedule {
		switch {
		case i < retailTarget:
			schedule[i] = workloadRetail
		case i < retailTarget+activeTarget:
			schedule[i] = workloadActive
		default:
			schedule[i] = workloadHFT
		}
	}
	scheduleRNG := rand.New(rand.NewSource(benchSeed))
	scheduleRNG.Shuffle(len(schedule), func(i, j int) {
		schedule[i], schedule[j] = schedule[j], schedule[i]
	})

	type latencyStats struct {
		count int64
		sumNs int64
		minNs int64
		maxNs int64
	}

	jobs := make(chan int, iterations)
	errCh := make(chan error, iterations)
	workerStats := make([]latencyStats, numWorkers)

	var wg sync.WaitGroup

	memDone := make(chan bool)
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemory(memDone, &memWg)

	for w := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			stats := &workerStats[workerID]
			for jobIdx := range jobs {
				var gen pdfGenerator
				switch schedule[jobIdx] {
				case workloadRetail:
					gen = retailGen
					atomic.AddInt64(&retailCount, 1)
				case workloadActive:
					gen = activeGen
					atomic.AddInt64(&activeCount, 1)
				default:
					gen = hftGen
					atomic.AddInt64(&hftCount, 1)
				}

				start := time.Now()
				_, err := gen()
				elapsed := time.Since(start)
				if err != nil {
					errCh <- err
					continue
				}
				ns := elapsed.Nanoseconds()
				stats.count++
				stats.sumNs += ns
				if stats.count == 1 || ns < stats.minNs {
					stats.minNs = ns
				}
				if ns > stats.maxNs {
					stats.maxNs = ns
				}
			}
		}(w)
	}

	totalStart := time.Now()
	for i := range iterations {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	totalTime := time.Since(totalStart)

	memDone <- true
	memWg.Wait()
	close(errCh)

	var errCount int
	for e := range errCh {
		if errCount == 0 {
			fmt.Printf("  First error: %v\n", e)
		}
		errCount++
	}
	if errCount > 0 {
		fmt.Printf("Encountered %d errors during execution.\n", errCount)
		return fmt.Errorf("benchmark failed with %d errors", errCount)
	}

	var totalCount, totalSumNs, minNs, maxNs int64
	for _, stats := range workerStats {
		if stats.count == 0 {
			continue
		}
		totalCount += stats.count
		totalSumNs += stats.sumNs
		if minNs == 0 || (stats.minNs > 0 && stats.minNs < minNs) {
			minNs = stats.minNs
		}
		if stats.maxNs > maxNs {
			maxNs = stats.maxNs
		}
	}
	if totalCount == 0 {
		fmt.Println("No results collected.")
		return errors.New("no results collected")
	}

	avgDuration := time.Duration(totalSumNs / totalCount)
	minDuration := time.Duration(minNs)
	maxDuration := time.Duration(maxNs)
	opsPerSec := float64(iterations) / totalTime.Seconds()

	fmt.Println("=== Performance Summary ===")
	fmt.Printf("  Iterations:      %d\n", iterations)
	fmt.Printf("  Concurrency:     %d workers\n", numWorkers)
	fmt.Printf("  GOMAXPROCS:      %d\n", runtime.GOMAXPROCS(0))
	if benchCompliant {
		fmt.Printf("  Retail signing:  %s\n", retailSignAlgorithmLabel())
	} else {
		fmt.Println("  Retail signing:  disabled")
	}
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

	if !skipWrite && len(retailPDF) > 0 {
		outputDir := "./"
		for name, data := range map[string][]byte{
			outputPDFName("gpdf_zerodha_retail_output.pdf"): retailPDF,
			outputPDFName("gpdf_zerodha_active_output.pdf"): activePDF,
			outputPDFName("gpdf_zerodha_hft_output.pdf"):    hftPDF,
		} {
			outPath := filepath.Join(outputDir, name)
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				fmt.Printf("Error saving %s: %v\n", outPath, err)
			} else {
				fmt.Printf("Saved: %s (%d bytes)\n", outPath, len(data))
			}
		}
	}

	fmt.Println()
	fmt.Println("=== Done ===")
	return nil
}