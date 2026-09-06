package main

import (
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/benchmarktemplates"
)

func main() {
	benchmarktemplates.Fail(benchmarktemplates.RunSingleDocumentBenchmark("GoPDFSuit"))
}
