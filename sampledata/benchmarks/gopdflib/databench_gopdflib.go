package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

type benchmarkRecord struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Desc  string `json:"desc"`
}

func boolPtr(value bool) *bool {
	return &value
}

func repoRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
}

func readBenchmarkData() ([]benchmarkRecord, error) {
	dataPath := filepath.Join(repoRoot(), "sampledata", "benchmarks", "data.json")
	content, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var records []benchmarkRecord
	if err := json.Unmarshal(content, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func buildRows(records []benchmarkRecord) []gopdflib.Row {
	rows := []gopdflib.Row{{Row: []gopdflib.Cell{
		{Props: "Helvetica:10:100:left:1:1:1:1", Text: "ID", BgColor: "#D9EAF7"},
		{Props: "Helvetica:10:100:left:1:1:1:1", Text: "Name", BgColor: "#D9EAF7"},
		{Props: "Helvetica:10:100:left:1:1:1:1", Text: "Email", BgColor: "#D9EAF7"},
		{Props: "Helvetica:10:100:left:1:1:1:1", Text: "Role", BgColor: "#D9EAF7"},
		{Props: "Helvetica:10:100:left:1:1:1:1", Text: "Description", BgColor: "#D9EAF7"},
	}}}

	for index, record := range records {
		bgColor := ""
		if index%2 == 0 {
			bgColor = "#F0F0F0"
		}

		rows = append(rows, gopdflib.Row{Row: []gopdflib.Cell{
			{Props: "Helvetica:10:000:left:1:1:1:1", Text: fmt.Sprintf("%d", record.ID), BgColor: bgColor},
			{Props: "Helvetica:10:000:left:1:1:1:1", Text: record.Name, BgColor: bgColor},
			{Props: "Helvetica:10:000:left:1:1:1:1", Text: record.Email, BgColor: bgColor, Wrap: boolPtr(true)},
			{Props: "Helvetica:10:000:left:1:1:1:1", Text: record.Role, BgColor: bgColor},
			{Props: "Helvetica:10:000:left:1:1:1:1", Text: record.Desc, BgColor: bgColor, Wrap: boolPtr(true)},
		}})
	}

	return rows
}

func buildTemplate(records []benchmarkRecord) gopdflib.PDFTemplate {
	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			PageBorder:          "0:0:0:0",
			Page:                "A4",
			PageAlignment:       1,
			PdfTitle:            "User Report",
			PDFACompliant:       true,
			ArlingtonCompatible: true,
			EmbedFonts:          boolPtr(true),
		},
		Title: gopdflib.Title{
			Props: "Helvetica:16:100:center:0:0:0:0",
			Text:  "User Report",
		},
		Elements: []gopdflib.Element{
			{
				Type: "table",
				Table: &gopdflib.Table{
					MaxColumns:   5,
					ColumnWidths: []float64{15, 30, 50, 35, 60},
					Rows:         buildRows(records),
				},
			},
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:8:000:center",
			Text: "Generated with GoPDFLib data benchmark",
		},
	}
}

func runDataBenchGoPDFLib() error {
	records, err := readBenchmarkData()
	if err != nil {
		return fmt.Errorf("failed to read benchmark data: %w", err)
	}

	template := buildTemplate(records)
	const numDataWorkers = 48
	const dataIterations = 10

	fmt.Printf("=== GoPDFLib Data Benchmark ===\n")
	fmt.Printf("Iterations: %d | Workers: %d\n", dataIterations, numDataWorkers)

	var (
		mu      sync.Mutex
		timings []float64
		lastPDF []byte
		ops     atomic.Int64
		wg      sync.WaitGroup
	)

	memDone := make(chan bool)
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemoryData(memDone, &memWg)

	sem := make(chan struct{}, numDataWorkers)
	totalStart := time.Now()

	for i := 1; i <= dataIterations; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			pdfBytes, genErr := gopdflib.GeneratePDF(template)
			if genErr != nil {
				fmt.Printf("Run %d error: %v\n", idx, genErr)
				return
			}
			elapsed := float64(time.Since(start).Nanoseconds()) / 1_000_000
			ops.Add(1)
			mu.Lock()
			timings = append(timings, elapsed)
			lastPDF = pdfBytes
			mu.Unlock()
			fmt.Printf("Run %d: %.2f ms\n", idx, elapsed)
		}(i)
	}
	wg.Wait()
	totalSeconds := time.Since(totalStart).Seconds()

	memDone <- true
	memWg.Wait()

	if len(timings) == 0 {
		return fmt.Errorf("no successful runs")
	}

	outputPath := filepath.Join(filepath.Dir(mustCurrentFile()), "output_databench_gopdflib.pdf")
	if err := os.WriteFile(outputPath, lastPDF, 0o644); err != nil {
		return fmt.Errorf("failed to write output pdf: %w", err)
	}

	sort.Float64s(timings)
	p95idx := int(math.Ceil(float64(len(timings))*0.95)) - 1
	if p95idx < 0 {
		p95idx = 0
	}
	completed := len(timings)
	total := 0.0
	for _, t := range timings {
		total += t
	}

	fmt.Println()
	fmt.Printf("Min:        %.2f ms\n", timings[0])
	fmt.Printf("Avg:        %.2f ms\n", total/float64(completed))
	fmt.Printf("P95:        %.2f ms\n", timings[p95idx])
	fmt.Printf("Max:        %.2f ms\n", timings[completed-1])
	fmt.Printf("Throughput: %.2f ops/sec\n", float64(completed)/totalSeconds)
	fmt.Printf("Output: %s\n", filepath.Base(outputPath))

	return nil
}

func monitorMemoryData(done chan bool, wg *sync.WaitGroup) {
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

func mustCurrentFile() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to resolve current file path")
	}
	return currentFile
}
