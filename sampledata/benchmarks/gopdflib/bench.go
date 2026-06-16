package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/benchmarktemplates"
)

var (
	flagCPUProfile = flag.String("cpuprofile", "", "write CPU profile to file")
	flagMemProfile = flag.String("memprofile", "", "write heap profile to file")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if *flagCPUProfile != "" {
		f, err := os.Create(*flagCPUProfile)
		if err != nil {
			benchmarktemplates.Fail(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()
			benchmarktemplates.Fail(err)
		}
		defer func() {
			pprof.StopCPUProfile()
			_ = f.Close()
		}()
	}

	var err error
	switch {
	case len(args) > 0 && args[0] == "data":
		err = runDataBenchGoPDFLib()
	case len(args) == 0:
		err = benchmarktemplates.RunSingleDocumentBenchmark("GoPDFLib")
	default:
		err = fmt.Errorf("unknown mode %q (use data or omit for Zerodha)", args[0])
	}
	benchmarktemplates.Fail(err)

	if *flagMemProfile != "" {
		f, err := os.Create(*flagMemProfile)
		if err != nil {
			benchmarktemplates.Fail(err)
		}
		defer func() { _ = f.Close() }()
		if werr := pprof.WriteHeapProfile(f); werr != nil {
			benchmarktemplates.Fail(werr)
		}
	}
}
