package main

import (
	"os"
	"runtime"
	"strconv"
)

const (
	defaultDataIterations = 5000
	defaultDataWorkers    = 48
	defaultZerodhaIters   = 48
)

func benchDataIterations() int {
	return envInt("BENCH_ITERATIONS", defaultDataIterations)
}

func benchDataWorkers() int {
	return envInt("BENCH_WORKERS", defaultDataWorkers)
}

func benchQuiet() bool {
	return os.Getenv("BENCH_QUIET") == "1"
}

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

func effectiveWorkers(requested int) int {
	if requested <= 0 {
		requested = 1
	}
	cpus := runtime.NumCPU()
	if requested > cpus*2 {
		return cpus * 2
	}
	return requested
}
