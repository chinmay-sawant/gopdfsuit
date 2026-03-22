package main

import (
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/benchmarktemplates"
)

func main() {
	benchmarktemplates.Fail(benchmarktemplates.RunDocBenchmark("GoPDFSuit"))
}
