// BOPS Zerodha benchmark: bypass-cache ops/sec with fresh template per op.
//
// Each operation builds a new retail template with a unique contract-note
// ID, clears all content caches via gopdflib.ClearBOPSCaches, then
// generates. No prepared-template reuse, no shared slices across ops.
//
// Run: make bench-gopdflib-bops-x10
//
//	or: BENCH_ITERATIONS=200 BENCH_WORKERS=1 go run . (quick latency truth)
//	or: BENCH_ITERATIONS=5000 BENCH_WORKERS=48 go run . (throughput truth)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func floatPtr(f float64) *float64 { return &f }

func boolPtr(b bool) *bool { return &b }

// bopsCerts holds PEM inputs loaded once at startup. They are document
// inputs, not caches: every op still pays full PEM parse and signing cost
// because ClearBOPSCaches drops signer entries before each generation.
type bopsCerts struct {
	key   string
	cert  string
	chain []string
}

func loadBOPSCerts() bopsCerts {
	// Bench dir is sampledata/gopdflib/zerodha_bops; certs live at repo root.
	root := filepath.Join("..", "..", "..", "certs")
	key, err := os.ReadFile(filepath.Join(root, "leaf.key"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "bops: cannot read leaf.key: %v (running unsigned)\n", err)
		return bopsCerts{}
	}
	cert, err := os.ReadFile(filepath.Join(root, "leaf.pem"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "bops: cannot read leaf.pem: %v (running unsigned)\n", err)
		return bopsCerts{}
	}
	chainRaw, err := os.ReadFile(filepath.Join(root, "chain.pem"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "bops: cannot read chain.pem: %v (running unsigned)\n", err)
		return bopsCerts{}
	}
	var chain []string
	for _, part := range strings.Split(strings.TrimSpace(string(chainRaw)), "-----END CERTIFICATE-----") {
		if stripped := strings.TrimSpace(part); stripped != "" {
			chain = append(chain, stripped+"\n-----END CERTIFICATE-----")
		}
	}
	return bopsCerts{key: string(key), cert: string(cert), chain: chain}
}

func envInt(key string, fallback int) int {
	if raw := os.Getenv(key); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return fallback
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

func buildBOPSRetailTemplate(idx int, certs bopsCerts) gopdflib.PDFTemplate {
	uniq := "CN-BOPS-" + strconv.Itoa(idx) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	cfg := gopdflib.Config{
		PageBorder:          "0:0:0:0",
		Page:                "A4",
		PageAlignment:       1,
		PdfTitle:            "Contract Note - " + uniq,
		PDFACompliant:       true,
		TaggedPDF:           true,
		ArlingtonCompatible: true,
		EmbedFonts:          boolPtr(true),
	}
	if certs.key != "" {
		cfg.Signature = &gopdflib.SignatureConfig{
			Enabled:          true,
			Visible:          true,
			Name:             "Zerodha Compliance",
			Reason:           "I am the author of this document",
			Location:         "Mumbai, India",
			ContactInfo:      "compliance@brokerage.com",
			PrivateKeyPEM:    certs.key,
			CertificatePEM:   certs.cert,
			CertificateChain: certs.chain,
		}
	}
	template := gopdflib.PDFTemplate{
		Config: cfg,
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
							Text:      uniq + " | 2024-02-12",
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
				MaxColumns: 1, ColumnWidths: []float64{1},
				Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
					{Props: "Helvetica:11:100:left:1:1:1:1", Text: "SECTION A: CLIENT INFORMATION", BgColor: "#21618C", TextColor: "#FFFFFF", Dest: "client-info"},
				}}},
			}},
			{Type: "table", Table: &gopdflib.Table{
				MaxColumns: 4, ColumnWidths: []float64{1.2, 2, 1.2, 2},
				Rows: []gopdflib.Row{
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:100:left:1:0:0:1", Text: "Client Name:", BgColor: "#EBF5FB"},
						{Props: "Helvetica:9:000:left:0:0:0:1", Text: "BOPS Client " + strconv.Itoa(idx), BgColor: "#EBF5FB"},
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
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:center:1:0:0:1", Text: "TATASTEEL"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "INE081A01012"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "BUY", TextColor: "#27AE60"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "10"},
						{Props: "Helvetica:9:000:right:0:0:0:1", Text: "145.50"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "1,455.00"},
					}},
					{Row: []gopdflib.Cell{
						{Props: "Helvetica:9:000:center:1:0:0:1", Text: "INFY", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "INE009A01021", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "SELL", TextColor: "#E74C3C", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:center:0:0:0:1", Text: "5", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:right:0:0:0:1", Text: "1,650.00", BgColor: "#F8F9F9"},
						{Props: "Helvetica:9:000:right:0:1:0:1", Text: "8,250.00", BgColor: "#F8F9F9"},
					}},
				},
			}},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:7:000:center",
			Text: "ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL",
		},
	}
	template.SetPrecomputedStandardFonts("Helvetica")
	return template
}

func bopsSingleOp(idx int, certs bopsCerts) error {
	template := buildBOPSRetailTemplate(idx, certs)
	gopdflib.ClearBOPSCaches()
	doc, err := gopdflib.GeneratePDFBorrowed(template)
	if err != nil {
		return err
	}
	doc.Release()
	return nil
}

func main() {
	flag.Parse()
	fmt.Println("=== Zerodha BOPS Benchmark (bypass-cache, no content reuse) ===")

	iterations := envInt("BENCH_ITERATIONS", 48)
	workers := min(envInt("BENCH_WORKERS", 48), iterations)
	fmt.Printf("Iterations: %d | Workers: %d\n", iterations, workers)
	certs := loadBOPSCerts()
	if certs.key != "" {
		fmt.Println("Retail signing: enabled (cold PEM parse per op)")
	} else {
		fmt.Println("Retail signing: disabled (certs unreadable)")
	}

	var (
		mu        sync.Mutex
		durations []float64
		failures  atomic.Int64
		wg        sync.WaitGroup
	)

	memDone := make(chan bool)
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemory(memDone, &memWg)

	totalStart := time.Now()
	if workers == 1 {
		for idx := 1; idx <= iterations; idx++ {
			start := time.Now()
			if err := bopsSingleOp(idx, certs); err != nil {
				failures.Add(1)
				fmt.Fprintf(os.Stderr, "op %d failed: %v\n", idx, err)
				continue
			}
			elapsedMs := float64(time.Since(start).Nanoseconds()) / 1_000_000
			durations = append(durations, elapsedMs)
			fmt.Printf("Run %d: %.2f ms\n", idx, elapsedMs)
		}
	} else {
		sem := make(chan struct{}, workers)
		for runIndex := 1; runIndex <= iterations; runIndex++ {
			sem <- struct{}{}
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				defer func() { <-sem }()
				start := time.Now()
				if err := bopsSingleOp(idx, certs); err != nil {
					failures.Add(1)
					fmt.Fprintf(os.Stderr, "op %d failed: %v\n", idx, err)
					return
				}
				elapsedMs := float64(time.Since(start).Nanoseconds()) / 1_000_000
				mu.Lock()
				durations = append(durations, elapsedMs)
				mu.Unlock()
				fmt.Printf("Run %d: %.2f ms\n", idx, elapsedMs)
			}(runIndex)
		}
		wg.Wait()
	}
	totalSeconds := time.Since(totalStart).Seconds()

	memDone <- true
	memWg.Wait()

	if len(durations) == 0 {
		fmt.Printf("bops: no successful runs (%d failures)\n", failures.Load())
		os.Exit(1)
	}
	sort.Float64s(durations)
	completed := len(durations)
	totalDur := 0.0
	for _, d := range durations {
		totalDur += d
	}
	avgMs := totalDur / float64(completed)
	opsPerSec := float64(completed) / totalSeconds

	fmt.Println("=== Performance Summary ===")
	fmt.Printf("  Iterations:      %d\n", completed)
	fmt.Printf("  Failures:        %d\n", failures.Load())
	fmt.Printf("  Total time:      %.3f s\n", totalSeconds)
	fmt.Printf("  Throughput:      %.2f ops/sec\n", opsPerSec)
	fmt.Println()
	fmt.Printf("  Avg Latency:     %.3f ms\n", avgMs)
	fmt.Printf("  Min Latency:     %.3f ms\n", durations[0])
	fmt.Printf("  Max Latency:     %.3f ms\n", durations[completed-1])
	fmt.Println()
	fmt.Printf("Min:        %.2f ms\n", durations[0])
	fmt.Printf("Avg:        %.2f ms\n", avgMs)
	fmt.Printf("Max:        %.2f ms\n", durations[completed-1])
	fmt.Println()
	fmt.Println("=== Done ===")
}
