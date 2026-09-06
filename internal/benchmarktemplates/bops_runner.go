package benchmarktemplates

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

// RunBOPSBenchmark renders fresh Zerodha retail templates with all content
// caches cleared before every operation. BOPS is bypass-cache ops/sec: the
// cold-path cost of generating a unique document, as opposed to the warm
// WOPS path that reuses one prepared template and hot caches.
//
// Each operation builds a new template, mutates a unique contract-note ID
// so no two documents share identical bytes, clears caches, then generates.
// Output lines match run_bench_x10.sh regexes: Throughput, Avg Latency,
// Max Memory Allocated.
func RunBOPSBenchmark(name string) error {
	iterations := envInt("BENCH_ITERATIONS", defaultWorkers)
	workers := min(envInt("BENCH_WORKERS", defaultWorkers), iterations)

	fmt.Println(BenchmarkHeader(name + " BOPS (no-cache)"))
	fmt.Printf("Iterations: %d | Workers: %d\n", iterations, workers)
	fmt.Println("Mode: bypass-cache, fresh template per op, unique doc ID per op")

	var (
		mu        sync.Mutex
		durations []float64
		ops       atomic.Int64
		failures  atomic.Int64
		wg        sync.WaitGroup
	)

	memDone := make(chan bool)
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemory(memDone, &memWg)

	sem := make(chan struct{}, workers)
	if workers == 1 {
		// Serial path keeps cache clearing race-free and gives clean latency truth.
		totalStart := time.Now()
		for idx := 1; idx <= iterations; idx++ {
			start := time.Now()
			if err := bopsSingleOp(idx); err != nil {
				failures.Add(1)
				fmt.Fprintf(os.Stderr, "op %d failed: %v\n", idx, err)
				continue
			}
			elapsedMs := float64(time.Since(start).Nanoseconds()) / 1_000_000
			durations = append(durations, elapsedMs)
			ops.Add(1)
			fmt.Printf("Run %d: %.2f ms\n", idx, elapsedMs)
		}
		totalSeconds := time.Since(totalStart).Seconds()
		memDone <- true
		memWg.Wait()
		return printBOPSStats(durations, totalSeconds, failures.Load())
	}

	totalStart := time.Now()
	for runIndex := 1; runIndex <= iterations; runIndex++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			start := time.Now()
			if err := bopsSingleOp(idx); err != nil {
				failures.Add(1)
				fmt.Fprintf(os.Stderr, "op %d failed: %v\n", idx, err)
				return
			}
			elapsedMs := float64(time.Since(start).Nanoseconds()) / 1_000_000
			mu.Lock()
			durations = append(durations, elapsedMs)
			mu.Unlock()
			ops.Add(1)
			fmt.Printf("Run %d: %.2f ms\n", idx, elapsedMs)
		}(runIndex)
	}
	wg.Wait()
	totalSeconds := time.Since(totalStart).Seconds()

	memDone <- true
	memWg.Wait()

	mu.Lock()
	defer mu.Unlock()
	return printBOPSStats(durations, totalSeconds, failures.Load())
}

func bopsSingleOp(idx int) error {
	template, err := BuildZerodhaRetailTemplate()
	if err != nil {
		return err
	}
	mutateBOPSTemplate(&template, idx)
	// Clear before generate so this op cannot reuse a prior op output.
	// Serial mode is exact. Parallel mode is best-effort by design:
	// shards are per-content so unique doc IDs still miss, and the
	// residual hit rate is reported rather than hidden.
	gopdflib.ClearBOPSCaches()
	doc, err := gopdflib.GeneratePDFBorrowed(template)
	if err != nil {
		return err
	}
	doc.Release()
	return nil
}

func mutateBOPSTemplate(template *gopdflib.PDFTemplate, idx int) {
	uniq := "CN-BOPS-" + strconv.Itoa(idx) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	template.Config.PdfTitle = "Contract Note - " + uniq
	if template.Title.Table != nil && len(template.Title.Table.Rows) > 0 && len(template.Title.Table.Rows[0].Row) > 1 {
		template.Title.Table.Rows[0].Row[1].Text = uniq + " | 2024-02-12"
	}
	// Touch one body cell so page content bytes differ per op and the
	// page compress cache cannot hit across iterations.
	for i := range template.Elements {
		tbl := template.Elements[i].Table
		if tbl == nil || len(tbl.Rows) == 0 {
			continue
		}
		for r := range tbl.Rows {
			for c := range tbl.Rows[r].Row {
				if tbl.Rows[r].Row[c].Text == "Rahul Sharma" {
					tbl.Rows[r].Row[c].Text = "BOPS Client " + strconv.Itoa(idx)
					return
				}
			}
		}
	}
}

func printBOPSStats(durations []float64, totalSeconds float64, failures int64) error {
	if len(durations) == 0 {
		return fmt.Errorf("bops: no successful runs (%d failures)", failures)
	}
	sort.Float64s(durations)
	completed := len(durations)
	totalDur := 0.0
	for _, d := range durations {
		totalDur += d
	}
	avgMs := totalDur / float64(completed)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println()
	fmt.Println("=== Performance Summary ===")
	fmt.Printf("  Iterations:      %d\n", completed)
	fmt.Printf("  Failures:        %d\n", failures)
	fmt.Printf("  Total time:      %.3f s\n", totalSeconds)
	fmt.Printf("  Throughput:      %.2f ops/sec\n", float64(completed)/totalSeconds)
	fmt.Println()
	fmt.Printf("  Avg Latency:     %.3f ms\n", avgMs)
	fmt.Printf("  Min Latency:     %.3f ms\n", durations[0])
	fmt.Printf("  Max Latency:     %.3f ms\n", durations[completed-1])
	fmt.Printf("  Heap Alloc:      %.2f MB\n", float64(m.Alloc)/1024/1024)
	// Extra lines for the x10 stats parser compatibility.
	fmt.Printf("Min:        %.2f ms\n", durations[0])
	fmt.Printf("Avg:        %.2f ms\n", avgMs)
	fmt.Printf("Max:        %.2f ms\n", durations[completed-1])
	return nil
}
