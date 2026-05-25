//go:build compare

package pdf

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func BenchmarkTypst(b *testing.B) {
	absPath, _ := filepath.Abs("../../sampledata/benchmarks/typst-x86_64-unknown-linux-musl/typst")
	typFilePath, _ := filepath.Abs("../../sampledata/benchmarks/typst/benchmark.typ")

	// Output to typst folder
	outPath, _ := filepath.Abs("../../sampledata/benchmarks/typst/output.pdf")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Output to file
		cmd := exec.Command(absPath, "compile", typFilePath, outPath)
		// Set Cwd so Typst can find data.json
		cmd.Dir = filepath.Dir(typFilePath)

		if output, err := cmd.CombinedOutput(); err != nil {
			b.Fatalf("Typst failed: %v, Output: %s", err, string(output))
		}
	}
}
