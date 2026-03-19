package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
			PageBorder:    "0:0:0:0",
			Page:          "A4",
			PageAlignment: 1,
			PdfTitle:      "User Report",
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
	iterations := 10
	fmt.Printf("=== GoPDFLib Data Benchmark ===\n")
	fmt.Printf("Iterations: %d\n", iterations)

	timings := make([]float64, 0, iterations)
	totalStart := time.Now()
	outputPath := filepath.Join(filepath.Dir(mustCurrentFile()), "output_databench_gopdflib.pdf")

	for i := 1; i <= iterations; i++ {
		start := time.Now()
		pdfBytes, err := gopdflib.GeneratePDF(template)
		if err != nil {
			return fmt.Errorf("failed to generate pdf: %w", err)
		}
		elapsed := time.Since(start)
		timings = append(timings, float64(elapsed.Nanoseconds())/1_000_000)

		if err := os.WriteFile(outputPath, pdfBytes, 0o644); err != nil {
			return fmt.Errorf("failed to write output pdf: %w", err)
		}

		fmt.Printf("Run %d: %.2f ms\n", i, timings[len(timings)-1])
	}

	totalSeconds := time.Since(totalStart).Seconds()
	min := timings[0]
	max := timings[0]
	total := 0.0
	for _, t := range timings {
		total += t
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}
	avg := total / float64(len(timings))

	fmt.Println()
	fmt.Printf("Min: %.2f ms\n", min)
	fmt.Printf("Avg: %.2f ms\n", avg)
	fmt.Printf("Max: %.2f ms\n", max)
	fmt.Printf("Throughput: %.2f ops/sec\n", float64(len(timings))/totalSeconds)
	fmt.Printf("Output: %s\n", filepath.Base(outputPath))

	return nil
}

func mustCurrentFile() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to resolve current file path")
	}
	return currentFile
}
