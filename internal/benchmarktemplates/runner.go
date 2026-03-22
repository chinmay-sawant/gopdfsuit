// Package benchmarktemplates provides reusable benchmark document builders and runners.
package benchmarktemplates

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

const (
	numWorkers = 48
	iterations = 10
)

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

// RunDocBenchmark renders the benchmark template concurrently using a
// 48-worker pool and prints timing statistics including P95 and throughput.
func RunDocBenchmark(name string) error {
	template, err := BuildZerodhaTemplate()
	if err != nil {
		return err
	}

	fmt.Println(BenchmarkHeader(name))
	fmt.Printf("Iterations: %d | Workers: %d\n", iterations, numWorkers)

	var (
		durations = make([]float64, iterations)
		ops       atomic.Int64
		wg        sync.WaitGroup
	)

	memDone := make(chan bool)
	var memWg sync.WaitGroup
	memWg.Add(1)
	go monitorMemory(memDone, &memWg)

	sem := make(chan struct{}, numWorkers)
	totalStart := time.Now()
	for runIndex := 1; runIndex <= iterations; runIndex++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			if _, genErr := gopdflib.GeneratePDF(template); genErr == nil {
				elapsedMs := float64(time.Since(start).Nanoseconds()) / 1_000_000
				durations[idx-1] = elapsedMs
				ops.Add(1)
				fmt.Printf("Run %d: %.2f ms\n", idx, elapsedMs)
			}
		}(runIndex)
	}
	wg.Wait()
	totalSeconds := time.Since(totalStart).Seconds()

	memDone <- true
	memWg.Wait()

	if len(durations) == 0 {
		return fmt.Errorf("no successful runs")
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
