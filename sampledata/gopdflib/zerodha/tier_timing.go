//go:build ignore

package main

import (
	"fmt"
	"time"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func benchTier(name string, tmpl gopdflib.PDFTemplate, n int) {
	gopdflib.GeneratePDF(tmpl)
	start := time.Now()
	for i := 0; i < n; i++ {
		gopdflib.GeneratePDF(tmpl)
	}
	elapsed := time.Since(start)
	perDoc := float64(elapsed.Microseconds()) / float64(n) / 1000.0
	ops := float64(n) / elapsed.Seconds()
	fmt.Printf("%s: %.2f ms/doc, %.0f ops/sec (n=%d)\n", name, perDoc, ops, n)
}

func main() {
	benchCompliant = true
	benchTier("retail", buildRetailTemplate(), 300)
	benchTier("active", buildActiveTraderTemplate(), 100)
	benchTier("hft", buildHFTTemplate(), 5)
}
