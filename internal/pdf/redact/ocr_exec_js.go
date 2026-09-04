//go:build js

package redact

import (
	"context"
	"os/exec"
)

// ocrCommand is the WASM stub for the os/exec-backed constructors in
// ocr_exec_unix.go (unix && !js) and ocr_exec_other.go (!unix && !js).
// Subprocesses do not exist under GOOS=js, and it is never reached:
// runOCRSearch and tesseractProvider.ExtractWords reject OCR up front with
// errOCRUnsupportedWASM. It panics instead of returning a *exec.Cmd so any
// future caller fails loudly rather than hanging on a fake command.
func ocrCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	_ = ctx
	_ = name
	_ = args
	panic("redact: OCR subprocesses are unsupported in WASM (GOOS=js)")
}
