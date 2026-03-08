package benchmarktemplates

import (
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

func RunSingleDocumentBenchmark(name string) error {
	template, err := BuildZerodhaRetailTemplate()
	if err != nil {
		return err
	}

	const iterations = 5
	durations := make([]float64, 0, iterations)

	fmt.Println(BenchmarkHeader(name))
	fmt.Printf("Iterations: %d\n", iterations)

	totalStart := time.Now()
	for runIndex := 1; runIndex <= iterations; runIndex++ {
		start := time.Now()
		if _, err := gopdflib.GeneratePDF(template); err != nil {
			return err
		}
		elapsedMs := float64(time.Since(start).Nanoseconds()) / 1_000_000
		durations = append(durations, elapsedMs)
		fmt.Printf("Run %d: %.2f ms\n", runIndex, elapsedMs)
	}
	totalSeconds := time.Since(totalStart).Seconds()

	minDuration := durations[0]
	maxDuration := durations[0]
	totalDuration := 0.0
	for _, duration := range durations {
		totalDuration += duration
		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	fmt.Println()
	fmt.Printf("Min: %.2f ms\n", minDuration)
	fmt.Printf("Avg: %.2f ms\n", totalDuration/iterations)
	fmt.Printf("Max: %.2f ms\n", maxDuration)
	fmt.Printf("Throughput: %.2f ops/sec\n", float64(iterations)/totalSeconds)
	return nil
}

func Fail(err error) {
	if err == nil {
		return
	}
	fmt.Println("Benchmark error:", err)
	os.Exit(1)
}
