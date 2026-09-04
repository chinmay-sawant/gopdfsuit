//go:build js

package redact

// isWASM reports whether this build targets the browser (GOOS=js, wasm).
// OCR needs pdftoppm/tesseract subprocesses plus a temp directory, none of
// which exist in the browser; requesting OCR under WASM must fail fast with
// errOCRUnsupportedWASM (see ocr_adapter.go) instead of reaching os/exec.
// Defined here behind //go:build js with its mirror in wasm_other.go.
func isWASM() bool { return true }
