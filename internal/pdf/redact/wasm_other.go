//go:build !js

package redact

// isWASM is the non-browser mirror of wasm_js.go: always false, so the
// OCR subprocess pipeline is unaffected on server builds.
func isWASM() bool { return false }
