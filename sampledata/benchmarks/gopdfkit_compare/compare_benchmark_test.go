// SPDX-License-Identifier: MIT
// Copyright (c) 2026 cssBruno

package gopdfsuitbench

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
	"github.com/cssbruno/gopdfkit/document"
	"github.com/cssbruno/gopdfkit/inspect"
)

const (
	workerCount40 = 40
	pageMarginPt  = 36.0
	rowHeightPt   = 18.0
	footerText    = "Shared PDF benchmark"
)

const (
	workloadText      = "text"
	workloadTable     = "table"
	workloadInvoice   = "invoice"
	workloadFullImage = "full_image"
	workloadImageRows = "image_rows"
)

type workload struct {
	name        string
	kind        string
	title       string
	rows        []rowData
	textLines   []string
	pngData     []byte
	pngBase64   string
	pngWidthPt  float64
	pngHeightPt float64
	template    gopdflib.PDFTemplate
}

type rowData struct {
	code        string
	description string
	value       string
}

type benchmarkEngine struct {
	name   string
	render func(workload, *pdfTarget) error
}

type pdfTarget struct {
	buf     bytes.Buffer
	release func()
}

func BenchmarkGoPDFKit(b *testing.B) {
	benchmarkEngineSuite(b, goPDFKitEngine())
}

func BenchmarkGoPDFLib(b *testing.B) {
	benchmarkEngineSuite(b, goPDFLibEngine())
}

func TestComparableOutputsArePDF(t *testing.T) {
	for _, wl := range workloads(t) {
		for _, engine := range engines() {
			var target pdfTarget
			if err := engine.render(wl, &target); err != nil {
				t.Fatalf("%s %s render error = %v", engine.name, wl.name, err)
			}
			if err := assertPDF(target.buf.Bytes()); err != nil {
				t.Fatalf("%s %s generated invalid PDF: %v", engine.name, wl.name, err)
			}
			if err := assertImportablePDF(target.buf.Bytes()); err != nil {
				t.Fatalf("%s %s generated non-importable PDF: %v", engine.name, wl.name, err)
			}
		}
	}
}

func benchmarkEngineSuite(b *testing.B, engine benchmarkEngine) {
	for _, wl := range workloads(b) {
		wl := wl
		for _, mode := range benchmarkModes() {
			mode := mode
			b.Run(fmt.Sprintf("%s/%s", wl.name, mode.name), func(b *testing.B) {
				benchmarkEngineWorkload(b, engine, wl, mode.workers)
			})
		}
	}
}

func benchmarkEngineWorkload(b *testing.B, engine benchmarkEngine, wl workload, workers int) {
	b.Helper()
	if workers < 1 {
		b.Fatalf("workers = %d, want >= 1", workers)
	}
	b.ReportAllocs()
	b.ReportMetric(float64(workers), "workers")

	var totalBytes atomic.Int64
	var firstErr error
	var errMu sync.Mutex
	setErr := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}
	runOne := func() {
		var target pdfTarget
		defer func() {
			if target.release != nil {
				target.release()
			}
		}()
		if err := engine.render(wl, &target); err != nil {
			setErr(err)
			return
		}
		if err := assertPDF(target.buf.Bytes()); err != nil {
			setErr(err)
			return
		}
		totalBytes.Add(int64(target.buf.Len()))
	}

	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	b.ResetTimer()
	start := time.Now()
	if workers == 1 {
		for i := 0; i < b.N; i++ {
			runOne()
		}
	} else {
		jobs := make(chan struct{}, workers)
		var wg sync.WaitGroup
		wg.Add(workers)
		for worker := 0; worker < workers; worker++ {
			go func() {
				defer wg.Done()
				for range jobs {
					runOne()
				}
			}()
		}
		for i := 0; i < b.N; i++ {
			jobs <- struct{}{}
		}
		close(jobs)
		wg.Wait()
	}
	elapsed := time.Since(start)
	b.StopTimer()

	if firstErr != nil {
		b.Fatal(firstErr)
	}
	if b.N > 0 {
		b.ReportMetric(float64(totalBytes.Load())/float64(b.N), "pdf_bytes")
		b.ReportMetric(float64(b.N)/elapsed.Seconds(), "pdf/s")
	}
	reportBenchmarkTotalAllocMB(b, before.TotalAlloc)
}

func reportBenchmarkTotalAllocMB(b *testing.B, beforeTotalAlloc uint64) {
	b.Helper()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if after.TotalAlloc < beforeTotalAlloc {
		return
	}
	b.ReportMetric(float64(after.TotalAlloc-beforeTotalAlloc)/(1024*1024), "total_MB")
}

func benchmarkModes() []struct {
	name    string
	workers int
} {
	return []struct {
		name    string
		workers int
	}{
		{name: "workers_40", workers: workerCount40},
	}
}

func workloads(tb testing.TB) []workload {
	tb.Helper()
	image := loadBenchmarkPNG(tb)
	workloads := []workload{
		{name: "text_short", kind: workloadText, title: "Shared short text benchmark", textLines: sharedTextLines(12)},
		{name: "text_240_lines", kind: workloadText, title: "Shared multipage text benchmark", textLines: sharedTextLines(240)},
		{name: "table_180_rows", kind: workloadTable, title: "Shared table benchmark", rows: sharedRows(180)},
		{name: "table_900_rows", kind: workloadTable, title: "Shared multipage table benchmark", rows: sharedRows(900)},
		{name: "invoice_40_rows", kind: workloadInvoice, title: "Shared invoice benchmark", rows: invoiceRows(40)},
		{
			name:        "png_table_180_rows",
			kind:        workloadFullImage,
			title:       "Shared image and table benchmark",
			rows:        sharedRows(180),
			pngData:     image.data,
			pngBase64:   image.base64,
			pngWidthPt:  image.widthPt,
			pngHeightPt: image.heightPt,
		},
		{
			name:        "png_rows_60",
			kind:        workloadImageRows,
			title:       "Shared table image rows benchmark",
			rows:        sharedRows(60),
			pngData:     image.data,
			pngBase64:   image.base64,
			pngWidthPt:  28,
			pngHeightPt: image.heightPt * 28 / image.widthPt,
		},
	}
	for i := range workloads {
		workloads[i].template = gopdfSuitTemplate(workloads[i])
	}
	return workloads
}

func engines() []benchmarkEngine {
	return []benchmarkEngine{goPDFKitEngine(), goPDFLibEngine()}
}

func goPDFKitEngine() benchmarkEngine {
	return benchmarkEngine{name: "gopdfkit", render: renderGoPDFKit}
}

func goPDFLibEngine() benchmarkEngine {
	return benchmarkEngine{name: "gopdflib", render: renderGoPDFLib}
}

func renderGoPDFKit(wl workload, target *pdfTarget) error {
	pdf := document.New("P", "pt", "A4", "")
	pdf.SetCompression(true)
	pdf.SetMargins(pageMarginPt, pageMarginPt, pageMarginPt)
	pdf.SetAutoPageBreak(true, pageMarginPt)
	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-24)
		pdf.SetFont("Helvetica", "", 8)
		pdf.CellFormat(0, 10, footerText, "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 22, wl.title, "", 1, "C", false, 0, "")

	switch wl.kind {
	case workloadText:
		renderGoPDFKitText(pdf, wl)
	case workloadInvoice:
		renderGoPDFKitInvoice(pdf, wl)
	case workloadFullImage:
		renderGoPDFKitFullImage(pdf, wl)
		renderGoPDFKitTable(pdf, wl.rows)
	case workloadImageRows:
		renderGoPDFKitImageRows(pdf, wl)
	default:
		renderGoPDFKitTable(pdf, wl.rows)
	}

	if err := pdf.Output(&target.buf); err != nil {
		return fmt.Errorf("gopdfkit output: %w", err)
	}
	return nil
}

func renderGoPDFKitText(pdf *document.Document, wl workload) {
	pdf.SetFont("Helvetica", "", 10)
	for _, line := range wl.textLines {
		pdf.MultiCell(0, 14, line, "", "L", false)
	}
}

func renderGoPDFKitFullImage(pdf *document.Document, wl workload) {
	if len(wl.pngData) == 0 {
		return
	}
	options := document.ImageOptions{ImageType: "png"}
	pdf.RegisterImageOptionsReader("shared-benchmark-png", options, bytes.NewReader(wl.pngData))
	pdf.ImageOptions("shared-benchmark-png", pageMarginPt, pdf.GetY()+8, wl.pngWidthPt, 0, false, options, 0, "")
	pdf.Ln(wl.pngHeightPt + 8)
}

func renderGoPDFKitTable(pdf *document.Document, rows []rowData) {
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(90, rowHeightPt, "Code", "1", 0, "L", false, 0, "")
	pdf.CellFormat(330, rowHeightPt, "Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(103, rowHeightPt, "Value", "1", 1, "R", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	for _, row := range rows {
		pdf.CellFormat(90, rowHeightPt, row.code, "1", 0, "L", false, 0, "")
		pdf.CellFormat(330, rowHeightPt, row.description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(103, rowHeightPt, row.value, "1", 1, "R", false, 0, "")
	}
}

func renderGoPDFKitInvoice(pdf *document.Document, wl workload) {
	renderGoPDFKitTable(pdf, wl.rows)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(420, rowHeightPt, "Subtotal", "1", 0, "R", false, 0, "")
	pdf.CellFormat(103, rowHeightPt, invoiceTotal(wl.rows), "1", 1, "R", false, 0, "")
	pdf.CellFormat(420, rowHeightPt, "Tax", "1", 0, "R", false, 0, "")
	pdf.CellFormat(103, rowHeightPt, "81.42", "1", 1, "R", false, 0, "")
	pdf.CellFormat(420, rowHeightPt, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(103, rowHeightPt, "977.04", "1", 1, "R", false, 0, "")
}

func renderGoPDFKitImageRows(pdf *document.Document, wl workload) {
	if len(wl.pngData) > 0 {
		options := document.ImageOptions{ImageType: "png"}
		pdf.RegisterImageOptionsReader("shared-benchmark-png", options, bytes.NewReader(wl.pngData))
	}
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(54, rowHeightPt, "Image", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, rowHeightPt, "Code", "1", 0, "L", false, 0, "")
	pdf.CellFormat(330, rowHeightPt, "Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(49, rowHeightPt, "Value", "1", 1, "R", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	for _, row := range wl.rows {
		x, y := pdf.GetX(), pdf.GetY()
		pdf.CellFormat(54, 32, "", "1", 0, "L", false, 0, "")
		if len(wl.pngData) > 0 {
			pdf.ImageOptions("shared-benchmark-png", x+4, y+4, 24, 0, false, document.ImageOptions{ImageType: "png"}, 0, "")
		}
		pdf.CellFormat(90, rowHeightPt, row.code, "1", 0, "L", false, 0, "")
		pdf.CellFormat(330, rowHeightPt, row.description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(49, rowHeightPt, row.value, "1", 1, "R", false, 0, "")
		pdf.SetY(y + 32)
	}
}

func renderGoPDFLib(wl workload, target *pdfTarget) error {
	doc, err := gopdflib.GeneratePDFBorrowed(wl.template)
	if err != nil {
		return fmt.Errorf("gopdflib output: %w", err)
	}
	target.buf = *bytes.NewBuffer(doc.Bytes())
	target.release = doc.Release
	return nil
}

func gopdfSuitTemplate(wl workload) gopdflib.PDFTemplate {
	embedFonts := false
	elements := []gopdflib.Element{}
	switch wl.kind {
	case workloadText:
		elements = append(elements, gopdflib.Element{Type: "table", Table: gopdfSuitTextTable(wl)})
	case workloadInvoice:
		elements = append(elements, gopdflib.Element{Type: "table", Table: gopdfSuitTable(wl.rows)})
		elements = append(elements, gopdflib.Element{Type: "table", Table: gopdfSuitTotalsTable(wl.rows)})
	case workloadFullImage:
		if wl.pngBase64 != "" {
			elements = append(elements, gopdflib.Element{Type: "image", Image: &gopdflib.Image{
				ImageName: "shared-benchmark-png",
				ImageData: wl.pngBase64,
				Width:     wl.pngWidthPt,
				Height:    wl.pngHeightPt,
			}})
		}
		elements = append(elements, gopdflib.Element{Type: "table", Table: gopdfSuitTable(wl.rows)})
	case workloadImageRows:
		elements = append(elements, gopdflib.Element{Type: "table", Table: gopdfSuitImageRowsTable(wl)})
	default:
		elements = append(elements, gopdflib.Element{Type: "table", Table: gopdfSuitTable(wl.rows)})
	}

	return gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			Page:          "A4",
			PageAlignment: 1,
			PageMargin:    "36:36:36:36",
			PdfTitle:      wl.title,
			EmbedFonts:    &embedFonts,
		},
		Title: gopdflib.Title{
			Props: "Helvetica:16:100:center:0:0:0:0",
			Text:  wl.title,
		},
		Footer: gopdflib.Footer{
			Font: "Helvetica:8:000:left",
			Text: footerText,
		},
		Elements: elements,
	}
}

func gopdfSuitTable(rows []rowData) *gopdflib.Table {
	table := gopdflib.Table{
		MaxColumns:   3,
		ColumnWidths: []float64{90, 330, 103},
		Rows:         make([]gopdflib.Row, 0, len(rows)+1),
	}
	table.Rows = append(table.Rows, gopdflib.Row{Row: []gopdflib.Cell{
		gopdfSuitCell("Helvetica:9:100:left:1:1:1:1", "Code"),
		gopdfSuitCell("Helvetica:9:100:left:1:1:1:1", "Description"),
		gopdfSuitCell("Helvetica:9:100:right:1:1:1:1", "Value"),
	}})
	for _, row := range rows {
		table.Rows = append(table.Rows, gopdflib.Row{Row: []gopdflib.Cell{
			gopdfSuitCell("Helvetica:9:000:left:1:1:1:1", row.code),
			gopdfSuitCell("Helvetica:9:000:left:1:1:1:1", row.description),
			gopdfSuitCell("Helvetica:9:000:right:1:1:1:1", row.value),
		}})
	}
	return &table
}

func gopdfSuitTextTable(wl workload) *gopdflib.Table {
	rows := make([]gopdflib.Row, 0, len(wl.textLines))
	wrap := true
	for _, line := range wl.textLines {
		rows = append(rows, gopdflib.Row{Row: []gopdflib.Cell{{
			Props:  "Helvetica:10:000:left:0:0:0:0",
			Text:   line,
			Wrap:   &wrap,
			Height: floatPtr(18),
		}}})
	}
	return &gopdflib.Table{MaxColumns: 1, ColumnWidths: []float64{1}, Rows: rows}
}

func gopdfSuitTotalsTable(rows []rowData) *gopdflib.Table {
	return &gopdflib.Table{
		MaxColumns:   2,
		ColumnWidths: []float64{420, 103},
		Rows: []gopdflib.Row{
			{Row: []gopdflib.Cell{
				gopdfSuitCell("Helvetica:10:100:right:1:1:1:1", "Subtotal"),
				gopdfSuitCell("Helvetica:10:100:right:1:1:1:1", invoiceTotal(rows)),
			}},
			{Row: []gopdflib.Cell{
				gopdfSuitCell("Helvetica:10:100:right:1:1:1:1", "Tax"),
				gopdfSuitCell("Helvetica:10:100:right:1:1:1:1", "81.42"),
			}},
			{Row: []gopdflib.Cell{
				gopdfSuitCell("Helvetica:10:100:right:1:1:1:1", "Total"),
				gopdfSuitCell("Helvetica:10:100:right:1:1:1:1", "977.04"),
			}},
		},
	}
}

func gopdfSuitImageRowsTable(wl workload) *gopdflib.Table {
	table := gopdflib.Table{
		MaxColumns:   4,
		ColumnWidths: []float64{54, 90, 330, 49},
		Rows:         make([]gopdflib.Row, 0, len(wl.rows)+1),
	}
	table.Rows = append(table.Rows, gopdflib.Row{Row: []gopdflib.Cell{
		gopdfSuitCell("Helvetica:9:100:left:1:1:1:1", "Image"),
		gopdfSuitCell("Helvetica:9:100:left:1:1:1:1", "Code"),
		gopdfSuitCell("Helvetica:9:100:left:1:1:1:1", "Description"),
		gopdfSuitCell("Helvetica:9:100:right:1:1:1:1", "Value"),
	}})
	for _, row := range wl.rows {
		image := &gopdflib.Image{ImageName: "shared-benchmark-png", ImageData: wl.pngBase64, Width: 24, Height: wl.pngHeightPt}
		table.Rows = append(table.Rows, gopdflib.Row{Row: []gopdflib.Cell{
			{Props: "Helvetica:9:000:left:1:1:1:1", Image: image, Height: floatPtr(32)},
			gopdfSuitCell("Helvetica:9:000:left:1:1:1:1", row.code),
			gopdfSuitCell("Helvetica:9:000:left:1:1:1:1", row.description),
			gopdfSuitCell("Helvetica:9:000:right:1:1:1:1", row.value),
		}})
	}
	return &table
}

func gopdfSuitCell(props, text string) gopdflib.Cell {
	height := rowHeightPt
	return gopdflib.Cell{Props: props, Text: text, Height: &height}
}

func floatPtr(v float64) *float64 {
	return &v
}

func sharedRows(count int) []rowData {
	rows := make([]rowData, 0, count)
	for row := 0; row < count; row++ {
		rows = append(rows, rowData{
			code:        fmt.Sprintf("ROW-%03d", row),
			description: "Operational PDF generation benchmark row",
			value:       fmt.Sprintf("%0.2f", float64(row)*1.25),
		})
	}
	return rows
}

func invoiceRows(count int) []rowData {
	rows := make([]rowData, 0, count)
	for row := 0; row < count; row++ {
		rows = append(rows, rowData{
			code:        fmt.Sprintf("INV-%03d", row+1),
			description: fmt.Sprintf("Professional services line %03d", row+1),
			value:       fmt.Sprintf("%0.2f", float64(row+1)*1.09),
		})
	}
	return rows
}

func invoiceTotal(rows []rowData) string {
	var total float64
	for idx := range rows {
		total += float64(idx+1) * 1.09
	}
	return fmt.Sprintf("%0.2f", total)
}

func sharedTextLines(count int) []string {
	lines := make([]string, 0, count)
	for line := 0; line < count; line++ {
		lines = append(lines, fmt.Sprintf(
			"Paragraph %03d: shared benchmark text with predictable length, punctuation, and line flow for both PDF libraries.",
			line+1,
		))
	}
	return lines
}

type benchmarkImage struct {
	data     []byte
	base64   string
	widthPt  float64
	heightPt float64
}

func loadBenchmarkPNG(tb testing.TB) benchmarkImage {
	tb.Helper()
	data, err := os.ReadFile("logo.png")
	if err != nil {
		tb.Fatalf("read benchmark PNG: %v", err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		tb.Fatalf("decode benchmark PNG config: %v", err)
	}
	width := 523.0
	return benchmarkImage{
		data:     data,
		base64:   base64.StdEncoding.EncodeToString(data),
		widthPt:  width,
		heightPt: width * float64(cfg.Height) / float64(cfg.Width),
	}
}

func assertPDF(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("generated empty PDF")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return fmt.Errorf("missing PDF header")
	}
	if !bytes.Contains(data, []byte("%%EOF")) {
		return fmt.Errorf("missing PDF EOF marker")
	}
	return nil
}

func assertImportablePDF(data []byte) error {
	if _, err := inspect.PageCount(data); err != nil {
		return fmt.Errorf("page count: %w", err)
	}
	return nil
}

func BenchmarkGoPDFKitHTMLSubset(b *testing.B) {
	if os.Getenv("GOPDFKIT_COMPARE_HTML") == "" {
		b.Skip("set GOPDFKIT_COMPARE_HTML=1 to run non-equivalent HTML comparison")
	}
	benchmarkOptionalHTML(b, "gopdfkit", renderGoPDFKitHTML)
}

func BenchmarkGoPDFLibHTMLChrome(b *testing.B) {
	if os.Getenv("GOPDFKIT_COMPARE_HTML") == "" {
		b.Skip("set GOPDFKIT_COMPARE_HTML=1 to run Chrome-backed HTML comparison")
	}
	benchmarkOptionalHTML(b, "gopdflib", renderGoPDFLibHTML)
}

func benchmarkOptionalHTML(b *testing.B, engine string, render func(*pdfTarget) error) {
	b.ReportAllocs()
	b.ReportMetric(1, "workers")
	var totalBytes atomic.Int64
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		var target pdfTarget
		if err := render(&target); err != nil {
			b.Fatalf("%s HTML render error = %v", engine, err)
		}
		if err := assertPDF(target.buf.Bytes()); err != nil {
			b.Fatalf("%s HTML generated invalid PDF: %v", engine, err)
		}
		totalBytes.Add(int64(target.buf.Len()))
	}
	elapsed := time.Since(start)
	b.StopTimer()
	if b.N > 0 {
		b.ReportMetric(float64(totalBytes.Load())/float64(b.N), "pdf_bytes")
		b.ReportMetric(float64(b.N)/elapsed.Seconds(), "pdf/s")
	}
	reportBenchmarkTotalAllocMB(b, before.TotalAlloc)
}

func renderGoPDFKitHTML(target *pdfTarget) error {
	pdf := document.New("P", "pt", "A4", "")
	pdf.SetCompression(true)
	pdf.SetMargins(pageMarginPt, pageMarginPt, pageMarginPt)
	pdf.SetAutoPageBreak(true, pageMarginPt)
	pdf.AddPage()
	html := pdf.HTMLNew()
	html.Write(12, benchmarkHTML())
	if err := pdf.Output(&target.buf); err != nil {
		return fmt.Errorf("gopdfkit HTML output: %w", err)
	}
	return nil
}

func renderGoPDFLibHTML(target *pdfTarget) error {
	pdfBytes, err := gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{
		HTML:         "<!doctype html><html><body>" + benchmarkHTML() + "</body></html>",
		PageSize:     "A4",
		Orientation:  "Portrait",
		MarginTop:    "12.7mm",
		MarginRight:  "12.7mm",
		MarginBottom: "12.7mm",
		MarginLeft:   "12.7mm",
	})
	if err != nil {
		return fmt.Errorf("gopdflib HTML output: %w", err)
	}
	_, err = target.buf.Write(pdfBytes)
	return err
}

func benchmarkHTML() string {
	var out bytes.Buffer
	out.WriteString(`<h1>Comparable HTML</h1><p>This fragment uses headings, paragraphs, lists, and a simple table.</p><ul>`)
	for i := 0; i < 12; i++ {
		fmt.Fprintf(&out, `<li>Item %02d</li>`, i+1)
	}
	out.WriteString(`</ul><table border="1"><tr><th>Code</th><th>Status</th></tr>`)
	for i := 0; i < 16; i++ {
		fmt.Fprintf(&out, `<tr><td>HTML-%02d</td><td>Ready</td></tr>`, i+1)
	}
	out.WriteString(`</table>`)
	return out.String()
}
