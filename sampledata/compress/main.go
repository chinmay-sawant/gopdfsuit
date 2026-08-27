// Package main compresses report.pdf with gopdflib at Light, Medium, and Heavy.
//
//	cd sampledata/compress && go run .
//
// Browser / Node WASM is sampledata/compress-js (make wasm-compress, then
// node run.mjs). This go run does not build or call WASM.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
		os.Exit(1)
	}

	srcPath := filepath.Join(dir, "report.pdf")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", srcPath, err)
		os.Exit(1)
	}

	fmt.Printf("source  report.pdf  %d bytes\n", len(src))

	levels := []struct {
		name  string
		file  string
		level gopdflib.CompressOptions
	}{
		{"1 Light", "report_level_1.pdf", gopdflib.CompressOptions{Level: gopdflib.CompressLight}},
		{"2 Medium", "report_level_2.pdf", gopdflib.CompressOptions{Level: gopdflib.CompressMedium}},
		{"3 Heavy", "report_level_3.pdf", gopdflib.CompressOptions{Level: gopdflib.CompressHeavy}},
	}

	for _, lv := range levels {
		out, err := gopdflib.CompressPDF(src, lv.level)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compress %s: %v\n", lv.name, err)
			os.Exit(1)
		}
		dest := filepath.Join(dir, lv.file)
		if err := os.WriteFile(dest, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", dest, err)
			os.Exit(1)
		}
		pct := 100 - float64(len(out))*100/float64(len(src))
		fmt.Printf("level %s  %s  %d bytes  (%.1f%% smaller)\n", lv.name, lv.file, len(out), pct)
	}
}
