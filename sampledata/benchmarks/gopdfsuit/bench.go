package main

import (
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/benchmarktemplates"
)

func main() {
	benchmarktemplates.Fail(benchmarktemplates.RunSingleDocumentBenchmark("GoPDFSuit"))
}
