package main

import (
	"os"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/benchmarktemplates"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "data" {
		benchmarktemplates.Fail(runDataBenchGoPDFLib())
		return
	}

	benchmarktemplates.Fail(benchmarktemplates.RunSingleDocumentBenchmark("GoPDFLib"))
}
