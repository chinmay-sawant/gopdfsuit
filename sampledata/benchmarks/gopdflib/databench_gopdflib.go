package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
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

		var idBuf [20]byte
		idStr := string(strconv.AppendInt(idBuf[:0], int64(record.ID), 10))
		rows = append(rows, gopdflib.Row{Row: []gopdflib.Cell{
			{Props: "Helvetica:10:000:left:1:1:1:1", Text: idStr, BgColor: bgColor},
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
			TaggedPDF:           true,
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
		return errors.Join(errors.New("failed to read benchmark data"), err)
	}

	template := buildTemplate(records)
	iterations := benchDataIterations()
	workers := effectiveWorkers(benchDataWorkers())
	quiet := benchQuiet()

	fmt.Printf("=== GoPDFLib Data Benchmark ===\n")
	fmt.Printf("Iterations: %d | Workers: %d | PDF/A: true\n", iterations, workers)

	var (
		mu      sync.Mutex
		timings []float64
		lastPDF []byte
		ops     atomic.Int64
		wg      sync.WaitGroup
	)

	var memStop atomic.Bool
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemoryData(&memStop, &memWg)

	sem := make(chan struct{}, workers)

	totalSeconds := func() float64 {
		start := time.Now()
		for i := 1; i <= iterations; i++ {
			sem <- struct{}{}
			wg.Add(1)
			go runDataBenchIter(&wg, sem, template, i, iterations, quiet, &mu, &timings, &lastPDF, &ops)
		}
		wg.Wait()
		close(sem)
		return time.Since(start).Seconds()
	}()

	memStop.Store(true)
	memWg.Wait()

	if len(timings) == 0 {
		return errors.New("no successful runs")
	}

	outputPath := filepath.Join(filepath.Dir(mustCurrentFile()), "output_databench_gopdflib.pdf")
	if err := os.WriteFile(outputPath, lastPDF, 0o644); err != nil {
		return errors.Join(errors.New("failed to write output pdf"), err)
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

func monitorMemoryData(stop *atomic.Bool, wg *sync.WaitGroup) {
	defer wg.Done()
	var maxAlloc uint64
	for !stop.Load() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.Alloc > maxAlloc {
			maxAlloc = m.Alloc
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("  Max Memory Allocated: %.2f MB\n", float64(maxAlloc)/1024/1024)
}

// runDataBenchIter is a named worker so defer is not in a loop body (PERF-7/40).
func runDataBenchIter(
	wg *sync.WaitGroup,
	sem chan struct{},
	template gopdflib.PDFTemplate,
	idx, iterations int,
	quiet bool,
	mu *sync.Mutex,
	timings *[]float64,
	lastPDF *[]byte,
	ops *atomic.Int64,
) {
	defer wg.Done()
	defer func() { <-sem }()

	start := time.Now()
	pdfBytes, genErr := gopdflib.GeneratePDF(template)
	if genErr != nil {
		var idBuf [20]byte
		idStr := string(strconv.AppendInt(idBuf[:0], int64(idx), 10))
		fmt.Printf("Run %s error: %v\n", idStr, genErr)
		return
	}
	elapsed := float64(time.Since(start).Nanoseconds()) / 1_000_000
	n := ops.Add(1)
	mu.Lock()
	*timings = append(*timings, elapsed)
	*lastPDF = pdfBytes
	mu.Unlock()
	if !quiet {
		var idBuf [20]byte
		idStr := string(strconv.AppendInt(idBuf[:0], int64(idx), 10))
		fmt.Printf("Run %s: %.2f ms\n", idStr, elapsed)
	} else if n%500 == 0 || int(n) == iterations {
		var nBuf, itBuf [20]byte
		nStr := string(strconv.AppendInt(nBuf[:0], n, 10))
		itStr := string(strconv.AppendInt(itBuf[:0], int64(iterations), 10))
		fmt.Printf("  progress: %s / %s (latest %.2f ms)\n", nStr, itStr, elapsed)
	}
}

func mustCurrentFile() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to resolve current file path")
	}
	return currentFile
}
