// Package benchmarktemplates provides reusable benchmark document builders and runners.
package benchmarktemplates

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

const (
	defaultWorkers = 48
)

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func monitorMemory(stop *atomic.Bool, wg *sync.WaitGroup) {
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

// measureGenerateMs times a single PDF generation (PERF-40).
func measureGenerateMs(template gopdflib.PDFTemplate) (float64, bool) {
	start := time.Now()
	_, err := gopdflib.GeneratePDF(template)
	if err != nil {
		return 0, false
	}
	return float64(time.Since(start).Nanoseconds()) / 1_000_000, true
}

// runBenchmarkIteration is a named worker so defer is not in a loop body (PERF-7).
func runBenchmarkIteration(
	wg *sync.WaitGroup,
	sem chan struct{},
	template gopdflib.PDFTemplate,
	idx int,
	mu *sync.Mutex,
	durations *[]float64,
	ops *atomic.Int64,
) {
	defer wg.Done()
	defer func() { <-sem }()

	if elapsedMs, ok := measureGenerateMs(template); ok {
		mu.Lock()
		*durations = append(*durations, elapsedMs)
		mu.Unlock()
		ops.Add(1)
		fmt.Printf("Run %d: %.2f ms\n", idx, elapsedMs)
	}
}

// RunSingleDocumentBenchmark renders the benchmark template concurrently using a
// 48-worker pool and prints timing statistics including P95 and throughput.
func RunSingleDocumentBenchmark(name string) error {
	template, err := BuildZerodhaRetailTemplate()
	if err != nil {
		return err
	}

	iterations := envInt("BENCH_ITERATIONS", defaultWorkers)
	workers := envInt("BENCH_WORKERS", defaultWorkers)
	if workers > iterations {
		workers = iterations
	}

	fmt.Println(BenchmarkHeader(name))
	fmt.Printf("Iterations: %d | Workers: %d\n", iterations, workers)

	var (
		mu        sync.Mutex
		durations []float64
		ops       atomic.Int64
		wg        sync.WaitGroup
	)

	var memStop atomic.Bool
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemory(&memStop, &memWg)

	sem := make(chan struct{}, workers)
	totalStart := time.Now()
	for runIndex := 1; runIndex <= iterations; runIndex++ {
		sem <- struct{}{}
		wg.Add(1)
		// PERF-7/36: named helper keeps defer out of loop body; pass idx explicitly
		go runBenchmarkIteration(&wg, sem, template, runIndex, &mu, &durations, &ops)
	}
	wg.Wait()
	totalSeconds := time.Since(totalStart).Seconds()

	memStop.Store(true)
	memWg.Wait()

	if len(durations) == 0 {
		return errors.New("no successful runs")
	}

	sort.Float64s(durations)
	p95idx := int(math.Ceil(float64(len(durations))*0.95)) - 1
	if p95idx < 0 {
		p95idx = 0
	}
	completed := len(durations)
	totalDur := 0.0
	for _, d := range durations {
		totalDur += d
	}

	fmt.Println()
	fmt.Printf("Min:        %.2f ms\n", durations[0])
	fmt.Printf("Avg:        %.2f ms\n", totalDur/float64(completed))
	fmt.Printf("P95:        %.2f ms\n", durations[p95idx])
	fmt.Printf("Max:        %.2f ms\n", durations[completed-1])
	fmt.Printf("Throughput: %.2f ops/sec\n", float64(completed)/totalSeconds)
	return nil
}

// Fail prints a benchmark error and exits the process with a non-zero status.
func Fail(err error) {
	if err == nil {
		return
	}
	fmt.Println("Benchmark error:", err)
	os.Exit(1)
}
