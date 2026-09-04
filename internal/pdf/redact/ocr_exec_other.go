//go:build !unix && !js

package redact

import (
	"context"
	"os/exec"
)

// ocrCommand is the portable fallback for platforms without process groups
// (Windows): a plain context-bound command without group kill.
// WASM (GOOS=js) uses ocr_exec_js.go instead: subprocesses cannot exist in
// the browser, so OCR is rejected with errOCRUnsupportedWASM before any
// command is built.
func ocrCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
