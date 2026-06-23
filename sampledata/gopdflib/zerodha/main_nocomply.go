//go:build nocomply

// Non-compliant Zerodha benchmark: same templates/workload with PDF/A, tagging,
// Arlington, font embedding, and retail signing disabled (throughput ceiling reference).
//
// Run: make bench-gopdflib-zerodha-nocomply
//  or: go run -tags nocomply .
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
)

func init() {
	benchCompliant = false
}

func main() {
	flag.Parse()

	if *flagCPUProfile != "" {
		f, err := os.Create(*flagCPUProfile)
		if err != nil {
			fmt.Printf("cpu profile: %v\n", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()
			fmt.Printf("cpu profile: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			pprof.StopCPUProfile()
			_ = f.Close()
		}()
	}

	if err := runBenchmark(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *flagMemProfile != "" {
		f, err := os.Create(*flagMemProfile)
		if err != nil {
			fmt.Printf("heap profile: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Printf("heap profile: %v\n", err)
			os.Exit(1)
		}
	}
}