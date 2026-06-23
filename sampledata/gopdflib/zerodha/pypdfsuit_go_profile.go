//go:build ignore

// Quick Go-side phase profile for the Python CGO JSON path.
// Run: go run pypdfsuit_go_profile.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/benchmarktemplates"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

const warmup = 5
const iters = 100

func mean(samples []time.Duration) float64 {
	var total time.Duration
	for _, s := range samples {
		total += s
	}
	return float64(total.Microseconds()) / float64(len(samples)) / 1000.0
}

func profileNative(template gopdflib.PDFTemplate) {
	for range warmup {
		if _, err := gopdflib.GeneratePDF(template); err != nil {
			panic(err)
		}
	}

	var gen []time.Duration
	for range iters {
		start := time.Now()
		if _, err := gopdflib.GeneratePDF(template); err != nil {
			panic(err)
		}
		gen = append(gen, time.Since(start))
	}
	fmt.Printf("  native GeneratePDF mean: %.3f ms (%.1f ops/s)\n", mean(gen), 1000.0/mean(gen))
}

func profileJSON(template gopdflib.PDFTemplate) {
	payload, err := json.Marshal(template)
	if err != nil {
		panic(err)
	}
	fmt.Printf("  JSON payload size: %d bytes\n", len(payload))

	for range warmup {
		var t gopdflib.PDFTemplate
		if err := json.Unmarshal(payload, &t); err != nil {
			panic(err)
		}
		if _, err := gopdflib.GeneratePDF(t); err != nil {
			panic(err)
		}
	}

	var unmarshal, gen, total []time.Duration
	for range iters {
		start := time.Now()
		var t gopdflib.PDFTemplate
		uStart := time.Now()
		if err := json.Unmarshal(payload, &t); err != nil {
			panic(err)
		}
		unmarshal = append(unmarshal, time.Since(uStart))
		gStart := time.Now()
		if _, err := gopdflib.GeneratePDF(t); err != nil {
			panic(err)
		}
		gen = append(gen, time.Since(gStart))
		total = append(total, time.Since(start))
	}

	fmt.Printf("  json.Unmarshal mean:     %.3f ms\n", mean(unmarshal))
	fmt.Printf("  GeneratePDF mean:        %.3f ms\n", mean(gen))
	fmt.Printf("  combined mean:             %.3f ms (%.1f ops/s)\n", mean(total), 1000.0/mean(total))
}

func main() {
	template, err := benchmarktemplates.BuildZerodhaRetailTemplate()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("=== Go Retail Phase Profile (100 iters) ===")
	fmt.Println("Native struct path:")
	profileNative(template)
	fmt.Println("JSON round-trip path (simulates CGO export):")
	profileJSON(template)
}